/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package sqldb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
)

var _ store.RouterRuleConfigStore = (*routerRuleStore)(nil)

// routerRuleStore impl
type routerRuleStore struct {
	master *BaseDB
	slave  *BaseDB
}

// CreateRoutingConfig Add a new routing configuration
func (r *routerRuleStore) CreateRoutingConfig(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		log.Errorf("[Store][boltdb] create routing config  missing id or revision")
		return store.NewStatusError(store.EmptyParamsErr, "missing id or revision")
	}
	if conf.Policy == "" || conf.Config == "" {
		log.Errorf("[Store][boltdb] create routing config  missing params")
		return store.NewStatusError(store.EmptyParamsErr, "missing some params")
	}

	err := RetryTransaction("CreateRoutingConfig", func() error {
		tx, err := r.master.Begin()
		if err != nil {
			return err
		}

		defer func() {
			_ = tx.Rollback()
		}()
		if err := r.createRoutingConfigTx(tx, conf); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] create routing config (%+v) commit: %s", conf, err.Error())
			return store.Error(err)
		}

		return nil
	})

	return store.Error(err)
}

func (r *routerRuleStore) CreateRoutingConfigTx(tx store.Tx, conf *rules.RouterConfig) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.createRoutingConfigTx(dbTx, conf)
}

func (r *routerRuleStore) createRoutingConfigTx(tx *BaseTx, conf *rules.RouterConfig) error {
	// 删除无效的数据
	if _, err := tx.Exec("DELETE FROM router_rule WHERE id = ? AND flag = 1", conf.ID); err != nil {
		log.Errorf("[Store][database] create routing (%+v) err: %s", conf, err.Error())
		return store.Error(err)
	}

	insertSQL := "INSERT INTO router_rule(id, namespace, name, policy, config, enable, " +
		" priority, revision, description, ctime, mtime, etime) VALUES (?,?,?,?,?,?,?,?,?,sysdate(),sysdate(),%s)"

	var enable int
	if conf.Enable {
		enable = 1
		insertSQL = fmt.Sprintf(insertSQL, "sysdate()")
	} else {
		enable = 0
		insertSQL = fmt.Sprintf(insertSQL, emptyEnableTime)
	}

	log.Debug("[Store][database] create routing ", zap.String("sql", insertSQL))

	if _, err := tx.Exec(insertSQL, conf.ID, conf.Namespace, conf.Name, conf.Policy,
		conf.Config, enable, conf.Priority, conf.Revision, conf.Description); err != nil {
		log.Errorf("[Store][database] create routing (%+v) err: %s", conf, err.Error())
		return store.Error(err)
	}
	return nil
}

// UpdateRoutingConfig Update a routing configuration
func (r *routerRuleStore) UpdateRoutingConfig(conf *rules.RouterConfig) error {

	tx, err := r.master.Begin()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := r.updateRoutingConfigTx(tx, conf); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[Store][database] update routing config (%+v) commit: %s", conf, err.Error())
		return store.Error(err)
	}

	return nil
}

func (r *routerRuleStore) UpdateRoutingConfigTx(tx store.Tx, conf *rules.RouterConfig) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.updateRoutingConfigTx(dbTx, conf)
}

func (r *routerRuleStore) updateRoutingConfigTx(tx *BaseTx, conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		log.Errorf("[Store][database] update routing config  missing id or revision")
		return store.NewStatusError(store.EmptyParamsErr, "missing id or revision")
	}
	if conf.Policy == "" || conf.Config == "" {
		log.Errorf("[Store][boltdb] create routing config  missing params")
		return store.NewStatusError(store.EmptyParamsErr, "missing some params")
	}

	str := "update router_rule set name = ?, policy = ?, config = ?, revision = ?, priority = ?, " +
		" description = ?, mtime = sysdate() where id = ?"
	if _, err := tx.Exec(str, conf.Name, conf.Policy, conf.Config, conf.Revision, conf.Priority, conf.Description,
		conf.ID); err != nil {
		log.Errorf("[Store][database] update routing config (%+v) exec err: %s", conf, err.Error())
		return store.Error(err)
	}
	return nil
}

// EnableRateLimit Enable current limit rules
func (r *routerRuleStore) EnableRouting(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		return errors.New("[Store][database] enable routing config  missing some params")
	}

	err := RetryTransaction("EnableRouting", func() error {
		var (
			enable   int
			etimeStr string
		)
		if conf.Enable {
			enable = 1
			etimeStr = "sysdate()"
		} else {
			enable = 0
			etimeStr = emptyEnableTime
		}
		str := fmt.Sprintf(
			`update router_rule set enable = ?, revision = ?, mtime = sysdate(), etime=%s where id = ?`, etimeStr)
		if _, err := r.master.Exec(str, enable, conf.Revision, conf.ID); err != nil {
			log.Errorf("[Store][database] update outing config (%+v), sql %s, err: %s", conf, str, err)
			return err
		}

		return nil
	})

	return store.Error(err)
}

// DeleteRoutingConfig Delete a routing configuration
func (r *routerRuleStore) DeleteRoutingConfig(ruleID string) error {

	if ruleID == "" {
		log.Errorf("[Store][database] delete routing config  missing service id")
		return store.NewStatusError(store.EmptyParamsErr, "missing service id")
	}

	str := `update router_rule set flag = 1, mtime = sysdate() where id = ?`
	if _, err := r.master.Exec(str, ruleID); err != nil {
		log.Errorf("[Store][database] delete routing config (%s) err: %s", ruleID, err.Error())
		return store.Error(err)
	}

	return nil
}

// GetMoreRouterRule Pull the incremental routing configuration information through mtime
func (r *routerRuleStore) GetMoreRouterRule(mtime time.Time, firstUpdate bool) ([]*rules.RouterConfig, error) {
	str := `select id, name, policy, config, enable, revision, flag, priority, description,
	unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime)  
	from router_rule where mtime > FROM_UNIXTIME(?) `

	if firstUpdate {
		str += " and flag != 1"
	}
	rows, err := r.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] query routing configs  with mtime err: %s", err.Error())
		return nil, err
	}
	out, err := fetchRoutingConfigRows(rows)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// GetRoutingConfigWithID Pull the routing configuration according to the rules ID
func (r *routerRuleStore) GetRoutingConfigWithID(ruleID string) (*rules.RouterConfig, error) {

	tx, err := r.master.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()
	return r.getRoutingConfigWithIDTx(tx, ruleID)
}

// GetRoutingConfigWithIDTx Pull the routing configuration according to the rules ID
func (r *routerRuleStore) GetRoutingConfigWithIDTx(tx store.Tx, ruleID string) (*rules.RouterConfig, error) {

	if tx == nil {
		return nil, errors.New("transaction is nil")
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.getRoutingConfigWithIDTx(dbTx, ruleID)
}

func (r *routerRuleStore) getRoutingConfigWithIDTx(tx *BaseTx, ruleID string) (*rules.RouterConfig, error) {

	str := `select id, name, policy, config, enable, revision, flag, priority, description,
	unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime)
	from router_rule 
	where id = ? and flag = 0`
	rows, err := tx.Query(str, ruleID)
	if err != nil {
		log.Errorf("[Store][database] query routing  with id(%s) err: %s", ruleID, err.Error())
		return nil, err
	}

	out, err := fetchRoutingConfigRows(rows)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	return out[0], nil
}

// GetRateLimitRuleVersions .
func (r *routerRuleStore) GetRouterRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	countSql := `SELECT COUNT(*) FROM router_rule_release WHERE rule_id = ? AND flag = 0`
	row := r.slave.QueryRow(countSql, filter["rule_id"])
	var count uint64
	if err := row.Scan(&count); err != nil {
		log.Errorf("[store][mysql][router] query router rule versions count err: %s", err.Error())
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}

	querySql := `SELECT id, name, rule_id, rule_name, flag, active, version, description, release_type, unix_timestamp(ctime), unix_timestamp(mtime)
	FROM router_rule_release
	WHERE rule_id = ?
		AND flag = 0 ORDER BY version DESC LIMIT ?, ?`
	rows, err := r.slave.Query(querySql, filter["rule_id"], offset, limit)
	if err != nil {
		log.Errorf("[store][mysql][router] query router rule versions err: %s", err.Error())
		return 0, nil, store.Error(err)
	}

	defer rows.Close()
	var releases []*rules.RuleRelease
	for rows.Next() {
		var (
			item         = &rules.RuleRelease{}
			flag, active int
			ctime, mtime int64
		)
		err := rows.Scan(&item.Id, &item.ReleaseName, &item.RuleId, &item.RuleName, &flag, &active, &item.Version, &item.Description, &item.ReleaseType, &ctime, &mtime)
		if err != nil {
			log.Errorf("[store][mysql][router] fetch router rule versions scan err: %s", err.Error())
			return 0, nil, store.Error(err)
		}
		item.Active = active == 1
		item.Valid = flag == 0
		item.Ctime = time.Unix(ctime, 0)
		item.Mtime = time.Unix(mtime, 0)
		releases = append(releases, item)
	}

	return count, releases, nil
}

// LockRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) LockRouterRule(tx store.Tx, name string) (*rules.RouterConfig, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)

	str := `select id, name, policy, config, enable, revision, flag, priority, description,
	unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime)
	from router_rule 
	where name = ? or id = ? and flag = 0 for update`
	rows, err := dbTx.Query(str, name, name)
	if err != nil {
		log.Errorf("[Store][database] query routing  with id(%s) err: %s", name, err.Error())
		return nil, err
	}

	out, err := fetchRoutingConfigRows(rows)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out[0], nil
}

// ActiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) ActiveRouterRule(tx store.Tx, release *rules.RouterRuleRelease) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := r.inactiveRouterRuleRelease(dbTx, release)
	if err != nil {
		return store.Error(err)
	}

	// 3. 设置目标规则 active=1, version=maxVersion+1, mtime=sysdate()
	_, err = dbTx.Exec("UPDATE router_rule_release SET active=1, version=?, mtime=sysdate() WHERE id=?", maxVersion+1, release.Id)
	return store.Error(err)
}

// GetReleaseRouterRule 获取已发布的路由规则
func (r *routerRuleStore) GetReleaseRouterRule(tx store.Tx, release *rules.RuleRelease) (*rules.RouterRuleRelease, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_id, rule_name, rule, version, active, description, release_type
	FROM router_rule_release
	WHERE name = ?
		AND rule_name = ?
		AND release_type = ?
		AND flag = 0
	LIMIT 1`
	row := dbTx.QueryRow(querySql, release.ReleaseName, release.RuleName, release.ReleaseType)
	var (
		id, name, ruleId, ruleName, ruleStr, description, releaseType string
		version                                                       uint64
		active                                                        int
	)
	err := row.Scan(&id, &name, &ruleId, &ruleName, &ruleStr, &version, &active, &description, &releaseType)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ruleObj := &rules.RouterConfig{}
	if err := json.Unmarshal([]byte(ruleStr), ruleObj); err != nil {
		return nil, err
	}
	pdata, _ := ruleObj.ToExpendRoutingConfig()
	return &rules.RouterRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			RuleId:      ruleId,
			RuleName:    ruleName,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       true,
		},
		Rule: pdata,
	}, nil
}

// GetActiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) GetActiveRouterRule(tx store.Tx, release *rules.RouterRuleRelease) (*rules.RouterRuleRelease, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, version
		, active, description, release_type
FROM router_rule_release
WHERE %s
	AND active = 1
	AND flag = 0
ORDER by version DESC
LIMIT 1`
	whereHolder := []string{"1=1"}
	args := make([]any, 0, 4)
	if release.RuleName != "" {
		whereHolder = append(whereHolder, "rule_name = ?")
		args = append(args, release.RuleName)
	}
	if release.ReleaseName != "" {
		whereHolder = append(whereHolder, "name = ?")
		args = append(args, release.ReleaseName)
	}
	if release.ReleaseType != "" {
		whereHolder = append(whereHolder, "release_type = ?")
		args = append(args, release.ReleaseType)
	}
	querySql = fmt.Sprintf(querySql, strings.Join(whereHolder, " AND "))
	row := dbTx.QueryRow(querySql, args...)
	var (
		id, name, ruleName, ruleStr, description, releaseType string
		version                                               uint64
		active                                                int
	)
	err := row.Scan(&id, &name, &ruleName, &ruleStr, &version, &active, &description, &releaseType)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ruleObj := &rules.RouterConfig{}
	if err := json.Unmarshal([]byte(ruleStr), ruleObj); err != nil {
		return nil, err
	}
	pdata, _ := ruleObj.ToExpendRoutingConfig()
	return &rules.RouterRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       true,
		},
		Rule: pdata,
	}, nil
}

// InactiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) InactiveRouterRule(tx store.Tx, release *rules.RouterRuleRelease) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	_, err := dbTx.Exec("UPDATE router_rule_release SET active = 0, mtime = sysdate() WHERE name = ? AND rule_name = ? AND active = 1",
		release.ReleaseName, release.Rule.Name)
	return err
}

// PublishRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) PublishRouterRule(tx store.Tx, rule *rules.RouterRuleRelease) error {
	if rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][router] publish router rule missing some params")
	}
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := r.inactiveRouterRuleRelease(dbTx, rule)
	if err != nil {
		return store.Error(err)
	}
	ruleJson, err := json.Marshal(rule.Rule)
	if err != nil {
		return err
	}
	// 3. 插入新发布并激活
	insertSql := `INSERT INTO router_rule_release (
		id, name, rule_id, rule_name, description, release_type, rule, version, active, ctime, mtime
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, sysdate(), sysdate())`
	_, err = dbTx.Exec(insertSql,
		rule.Id,
		rule.ReleaseName,
		rule.RuleId,
		rule.RuleName,
		rule.Description,
		rule.ReleaseType,
		string(ruleJson),
		maxVersion+1,
	)
	return err
}

func (r *routerRuleStore) inactiveRouterRuleRelease(tx *BaseTx, release *rules.RouterRuleRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.RuleName, release.ReleaseType}
	//	先取消所有 active == true 的记录
	if _, err := tx.Exec("UPDATE router_rule_release SET active = 0, mtime = sysdate() "+
		" WHERE rule_name = ? AND active = 1 AND release_type = ?", args...); err != nil {
		return 0, err
	}
	return r.selectMaxVersion(tx, release)
}

func (r *routerRuleStore) selectMaxVersion(tx *BaseTx, release *rules.RouterRuleRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.Rule.Name}
	var maxVersion uint64
	//	查询当前 release 的最大版本号
	if err := tx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM router_rule_release WHERE rule_name = ?",
		args...).Scan(&maxVersion); err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// GetMoreRouterRuleReleases implements store.RouterRuleConfigStore.
func (r *routerRuleStore) GetMoreRouterRuleReleases(firstUpdate bool, mtime time.Time) ([]*rules.RouterRuleRelease, error) {
	str := `SELECT id, name, description, release_type, rule, version, active, unix_timestamp(mtime) FROM router_rule_release WHERE mtime > FROM_UNIXTIME(?)`
	if firstUpdate {
		str += " AND active = 1"
	}
	rows, err := r.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*rules.RouterRuleRelease
	for rows.Next() {
		var (
			id, name, description, releaseType, ruleStr string
			version                                     uint64
			active                                      int
			mtime                                       time.Time
		)
		err := rows.Scan(&id, &name, &description, &releaseType, &ruleStr, &version, &active, &mtime)
		if err != nil {
			return nil, err
		}
		ruleObj := &rules.RouterConfig{}
		if err := json.Unmarshal([]byte(ruleStr), ruleObj); err != nil {
			return nil, err
		}
		pdata, _ := ruleObj.ToExpendRoutingConfig()
		release := &rules.RouterRuleRelease{
			RuleRelease: rules.RuleRelease{
				Id:          id,
				ReleaseName: name,
				Description: description,
				ReleaseType: rules.ReleaseType(releaseType),
				Active:      active == 1,
				Version:     version,
				Valid:       true,
			},
			Rule: pdata,
		}
		out = append(out, release)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// fetchRoutingConfigRows Read the data of the database and release ROWS
func fetchRoutingConfigRows(rows *sql.Rows) ([]*rules.RouterConfig, error) {
	defer rows.Close()
	var out []*rules.RouterConfig
	for rows.Next() {
		var (
			entry               rules.RouterConfig
			flag, enable        int
			ctime, mtime, etime int64
		)

		err := rows.Scan(&entry.ID, &entry.Name, &entry.Policy, &entry.Config, &enable, &entry.Revision,
			&flag, &entry.Priority, &entry.Description, &ctime, &mtime, &etime)
		if err != nil {
			log.Errorf("[database][store] fetch routing config  scan err: %s", err.Error())
			return nil, err
		}

		entry.CreateTime = time.Unix(ctime, 0)
		entry.ModifyTime = time.Unix(mtime, 0)
		entry.EnableTime = time.Unix(etime, 0)
		entry.Valid = true
		if flag == 1 {
			entry.Valid = false
		}
		entry.Enable = enable == 1

		out = append(out, &entry)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[database][store] fetch routing config  next err: %s", err.Error())
		return nil, err
	}

	return out, nil
}
