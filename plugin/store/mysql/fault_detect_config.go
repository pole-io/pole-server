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

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/specification/source/go/api/v1/model"
)

var _ store.FaultDetectRuleStore = (*faultDetectRuleStore)(nil)

type faultDetectRuleStore struct {
	master *BaseDB
	slave  *BaseDB
}

const (
	labelCreateFaultDetectRule = "createFaultDetectRule"
	labelUpdateFaultDetectRule = "updateFaultDetectRule"
	labelDeleteFaultDetectRule = "deleteFaultDetectRule"
)

const (
	insertFaultDetectSql = `insert into fault_detect_rule(
			id, name, namespace, revision, description, dst_service, dst_namespace, dst_method, config, ctime, mtime)
			values(?,?,?,?,?,?,?,?,?, sysdate(),sysdate())`
	updateFaultDetectSql = `update fault_detect_rule set name = ?, namespace = ?, revision = ?, description = ?,
			dst_service = ?, dst_namespace = ?, dst_method = ?, config = ?, mtime = sysdate() where id = ?`
	deleteFaultDetectSql    = `update fault_detect_rule set flag = 1, mtime = sysdate() where id = ?`
	countFaultDetectSql     = `select count(*) from fault_detect_rule where flag = 0`
	queryFaultDetectFullSql = `select id, name, namespace, revision, description, dst_service, 
			dst_namespace, dst_method, config, unix_timestamp(ctime), unix_timestamp(mtime)
            from fault_detect_rule where flag = 0`
	queryFaultDetectBriefSql = `select id, name, namespace, revision, description, dst_service, 
			dst_namespace, dst_method, unix_timestamp(ctime), unix_timestamp(mtime)
            from fault_detect_rule where flag = 0`
	queryFaultDetectCacheSql = `select id, name, namespace, revision, description, dst_service, 
			dst_namespace, dst_method, config, flag, unix_timestamp(ctime), unix_timestamp(mtime)
			from fault_detect_rule where mtime > FROM_UNIXTIME(?)`
)

// CreateFaultDetectRule create fault detect rule
func (f *faultDetectRuleStore) CreateFaultDetectRule(fdRule *rules.FaultDetectRule) error {
	err := RetryTransaction(labelCreateFaultDetectRule, func() error {
		return f.createFaultDetectRule(fdRule)
	})
	return store.Error(err)
}

func (f *faultDetectRuleStore) createFaultDetectRule(fdRule *rules.FaultDetectRule) error {
	return f.master.processWithTransaction(labelCreateFaultDetectRule, func(tx *BaseTx) error {
		if _, err := tx.Exec(insertFaultDetectSql, fdRule.ID, fdRule.Name, fdRule.Namespace, fdRule.Revision,
			fdRule.Description, fdRule.DstService, fdRule.DstNamespace, fdRule.DstMethod, fdRule.Rule); err != nil {
			log.Errorf("[Store][database] fail to %s exec sql, rule(%+v), err: %s",
				labelCreateFaultDetectRule, fdRule, err.Error())
			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] fail to %s commit tx, rule(%+v), err: %s",
				labelCreateFaultDetectRule, fdRule, err.Error())
			return err
		}
		return nil
	})
}

// UpdateFaultDetectRule update fault detect rule
func (f *faultDetectRuleStore) UpdateFaultDetectRule(fdRule *rules.FaultDetectRule) error {
	err := RetryTransaction(labelUpdateFaultDetectRule, func() error {
		return f.master.processWithTransaction(labelUpdateFaultDetectRule, func(tx *BaseTx) error {
			log.Infof("%#v", fdRule)
			if _, err := tx.Exec(updateFaultDetectSql, fdRule.Name, fdRule.Namespace, fdRule.Revision,
				fdRule.Description, fdRule.DstService, fdRule.DstNamespace, fdRule.DstMethod, fdRule.Rule, fdRule.ID); err != nil {
				log.Errorf("[Store][database] fail to %s exec sql, rule(%+v), err: %s",
					labelUpdateFaultDetectRule, fdRule, err.Error())
				return err
			}

			if err := tx.Commit(); err != nil {
				log.Errorf("[Store][database] fail to %s commit tx, rule(%+v), err: %s",
					labelUpdateFaultDetectRule, fdRule, err.Error())
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

// DeleteFaultDetectRule delete fault detect rule
func (f *faultDetectRuleStore) DeleteFaultDetectRule(id string) error {
	err := RetryTransaction(labelDeleteFaultDetectRule, func() error {
		return f.deleteFaultDetectRule(id)
	})
	return store.Error(err)
}

func (f *faultDetectRuleStore) deleteFaultDetectRule(id string) error {
	return f.master.processWithTransaction(labelDeleteFaultDetectRule, func(tx *BaseTx) error {
		if _, err := tx.Exec(deleteFaultDetectSql, id); err != nil {
			log.Errorf("[Store][database] fail to %s exec sql, rule(%s), err: %s",
				labelDeleteFaultDetectRule, id, err.Error())
			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] fail to %s commit tx, rule(%s), err: %s",
				labelDeleteFaultDetectRule, id, err.Error())
			return err
		}
		return nil
	})
}

// HasFaultDetectRule check fault detect rule exists
func (f *faultDetectRuleStore) HasFaultDetectRule(id string) (bool, error) {
	queryParams := map[string]string{"id": id}
	count, err := f.getFaultDetectRulesCount(queryParams)
	if nil != err {
		return false, err
	}
	return count > 0, nil
}

// HasFaultDetectRuleByName check fault detect rule exists by name
func (f *faultDetectRuleStore) HasFaultDetectRuleByName(name string, namespace string) (bool, error) {
	queryParams := map[string]string{exactName: name, "namespace": namespace}
	count, err := f.getFaultDetectRulesCount(queryParams)
	if nil != err {
		return false, err
	}
	return count > 0, nil
}

// HasFaultDetectRuleByNameExcludeId check fault detect rule exists by name not this id
func (f *faultDetectRuleStore) HasFaultDetectRuleByNameExcludeId(
	name string, namespace string, id string) (bool, error) {
	queryParams := map[string]string{exactName: name, "namespace": namespace, excludeId: id}
	count, err := f.getFaultDetectRulesCount(queryParams)
	if nil != err {
		return false, err
	}
	return count > 0, nil
}

// GetFaultDetectRules get all fault detect rules by query and limit
func (f *faultDetectRuleStore) GetFaultDetectRules(
	filter map[string]string, offset uint32, limit uint32) (uint32, []*rules.FaultDetectRule, error) {
	var out []*rules.FaultDetectRule
	var err error

	bValue, ok := filter[briefSearch]
	var isBrief = ok && strings.ToLower(bValue) == "true"
	delete(filter, briefSearch)

	if isBrief {
		out, err = f.getBriefFaultDetectRules(filter, offset, limit)
	} else {
		out, err = f.getFullFaultDetectRules(filter, offset, limit)
	}
	if err != nil {
		return 0, nil, err
	}
	num, err := f.getFaultDetectRulesCount(filter)
	if err != nil {
		return 0, nil, err
	}
	return num, out, nil
}

// GetMoreFaultDetects(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRule, error) get increment circuitbreaker rules
func (f *faultDetectRuleStore) GetMoreFaultDetects(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRule, error) {
	str := queryFaultDetectCacheSql
	if firstUpdate {
		str += " and flag != 1"
	}
	rows, err := f.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] query fault detect rules with mtime err: %s", err.Error())
		return nil, err
	}
	fdRules, err := fetchFaultDetectRulesRows(rows)
	if err != nil {
		return nil, err
	}
	return fdRules, nil
}

// GetFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) GetFaultDetectRule(id string) (*rules.FaultDetectRule, error) {
	querySql := `SELECT id, name, namespace, revision, description, dst_service, dst_namespace, dst_method,
	 config, unix_timestamp(ctime), unix_timestamp(mtime) FROM fault_detect_rule WHERE id = ? AND flag = 0 FOR UPDATE`
	row := f.master.QueryRow(querySql, id)
	var fdRule rules.FaultDetectRule
	var ctime, mtime int64
	err := row.Scan(&fdRule.ID, &fdRule.Name, &fdRule.Namespace, &fdRule.Revision, &fdRule.Description,
		&fdRule.DstService, &fdRule.DstNamespace, &fdRule.DstMethod, &fdRule.Rule, &ctime, &mtime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	fdRule.CreateTime = time.Unix(ctime, 0)
	fdRule.ModifyTime = time.Unix(mtime, 0)
	fdRule.Valid = true
	return &fdRule, nil
}

// LockFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) LockFaultDetectRule(tx store.Tx, keyword string) (*rules.FaultDetectRule, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, namespace, revision, description, dst_service, dst_namespace, dst_method,
	 config, unix_timestamp(ctime), unix_timestamp(mtime) FROM fault_detect_rule WHERE (id = ? OR name = ?) AND flag = 0 FOR UPDATE`
	row := dbTx.QueryRow(querySql, keyword, keyword)
	var fdRule rules.FaultDetectRule
	var ctime, mtime int64
	err := row.Scan(&fdRule.ID, &fdRule.Name, &fdRule.Namespace, &fdRule.Revision, &fdRule.Description,
		&fdRule.DstService, &fdRule.DstNamespace, &fdRule.DstMethod, &fdRule.Rule, &ctime, &mtime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	fdRule.CreateTime = time.Unix(ctime, 0)
	fdRule.ModifyTime = time.Unix(mtime, 0)
	fdRule.Valid = true
	return &fdRule, nil
}

// ActiveFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) ActiveFaultDetectRule(tx store.Tx, release *rules.FaultDetectRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := f.inactiveFaultDetectRelease(dbTx, release)
	if err != nil {
		return store.Error(err)
	}
	args := []any{maxVersion + 1, release.ReleaseName, release.RuleName}
	_, err = dbTx.Exec(`UPDATE fault_detect_rule_release
SET active = 1, version = ?, mtime = sysdate()
WHERE name = ?
	AND rule_name = ?`, args...)
	return store.Error(err)
}

// GetFaultDetectRuleVersions .
func (f *faultDetectRuleStore) GetFaultDetectRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	countSql := `SELECT COUNT(*) FROM fault_detect_rule_release WHERE rule_name = ? AND flag = 0`
	row := f.slave.QueryRow(countSql, filter["rule_name"])
	var count uint64
	if err := row.Scan(&count); err != nil {
		log.Errorf("[store][mysql][faultdetect] query fault_detect_rule versions count err: %s", err.Error())
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}

	querySql := `SELECT id, name, rule_id, rule_name, flag, active, version, description, release_type, unix_timestamp(ctime), unix_timestamp(mtime)
	FROM fault_detect_rule_release
	WHERE rule_id = ?
		AND flag = 0 ORDER BY version DESC LIMIT ?, ?`
	rows, err := f.slave.Query(querySql, filter["rule_name"], offset, limit)
	if err != nil {
		log.Errorf("[store][mysql][faultdetect] query fault_detect_rule rule versions err: %s", err.Error())
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
			log.Errorf("[store][mysql][faultdetect] fetch fault_detect_rule rule versions scan err: %s", err.Error())
			return 0, nil, store.Error(err)
		}
		item.Active = active == 1
		item.Valid = flag == 0
		item.Ctime = time.Unix(ctime, 0)
		item.Mtime = time.Unix(mtime, 0)
		item.Resource = model.RuleRelease_FaultDetectRules
		releases = append(releases, item)
	}

	return count, releases, nil
}

func (f *faultDetectRuleStore) GetReleaseFaultDetectRule(tx store.Tx, release *rules.RuleRelease) (*rules.FaultDetectRelease, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, version, active, description, release_type
	FROM fault_detect_rule_release
	WHERE name = ?
		AND rule_id = ?
		AND release_type = ?
		AND flag = 0
	LIMIT 1`
	row := dbTx.QueryRow(querySql, release.ReleaseName, release.RuleId, release.ReleaseType)
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
	var ruleObj rules.FaultDetectRule
	if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
		return nil, err
	}
	return &rules.FaultDetectRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       true,
		},
		Rule: &ruleObj,
	}, nil
}

// GetActiveFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) GetActiveFaultDetectRule(tx store.Tx, release *rules.FaultDetectRelease) (*rules.FaultDetectRelease, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, version
		, active, description, release_type
FROM fault_detect_rule_release
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
	var ruleObj rules.FaultDetectRule
	if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
		return nil, err
	}
	return &rules.FaultDetectRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       true,
		},
		Rule: &ruleObj,
	}, nil
}

// InactiveFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) InactiveFaultDetectRule(tx store.Tx, release *rules.FaultDetectRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	_, err := dbTx.Exec(`UPDATE fault_detect_rule_release
SET active = 0, mtime = sysdate()
WHERE name = ?
	AND rule_name = ?
	AND active = 1`, release.ReleaseName, release.Rule.Name)
	return store.Error(err)
}

func (f *faultDetectRuleStore) inactiveFaultDetectRelease(tx *BaseTx, release *rules.FaultDetectRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.RuleName, release.ReleaseType}
	//	先取消所有 active == true 的记录
	if _, err := tx.Exec("UPDATE fault_detect_rule_release SET active = 0, mtime = sysdate() "+
		" WHERE rule_name = ? AND active = 1 AND release_type = ?", args...); err != nil {
		return 0, err
	}
	return f.selectMaxVersion(tx, release)
}

func (f *faultDetectRuleStore) selectMaxVersion(tx *BaseTx, release *rules.FaultDetectRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.Rule.Name}
	var maxVersion uint64
	//	查询当前 release 的最大版本号
	if err := tx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM fault_detect_rule_release WHERE rule_name = ?",
		args...).Scan(&maxVersion); err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// PublishFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) PublishFaultDetectRule(tx store.Tx, rule *rules.FaultDetectRelease) error {
	if rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][faultdetect] publish fault detect rule missing some params")
	}
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := f.inactiveFaultDetectRelease(dbTx, rule)
	if err != nil {
		return store.Error(err)
	}

	ruleJson, err := json.Marshal(rule.Rule)
	if err != nil {
		return err
	}
	insertSql := `INSERT INTO fault_detect_rule_release (
		id, name, rule_name, rule, version, active, description, release_type, ctime, mtime
	) VALUES (?, ?, ?, ?, ?, 1, ?, ?, sysdate(), sysdate())`
	_, err = dbTx.Exec(insertSql,
		rule.Id,
		rule.ReleaseName,
		"", // rule_name 暂未使用
		string(ruleJson),
		maxVersion+1,
		rule.Description,
		rule.ReleaseType,
	)
	return err
}

// GetMoreFaultDetectReleases implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) GetMoreFaultDetectReleases(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRelease, error) {
	str := `SELECT id, name, rule_name, rule, version, active, description, release_type, mtime FROM
	 fault_detect_rule_release WHERE mtime > FROM_UNIXTIME(?)`
	if firstUpdate {
		str += " AND active = 1"
	}
	rows, err := f.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*rules.FaultDetectRelease
	for rows.Next() {
		var (
			id, name, ruleName, ruleStr, description, releaseType string
			version                                               uint64
			active                                                int
			mtime                                                 time.Time
		)
		err := rows.Scan(&id, &name, &ruleName, &ruleStr, &version, &active, &description, &releaseType, &mtime)
		if err != nil {
			return nil, err
		}
		var ruleObj rules.FaultDetectRule
		if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
			return nil, err
		}
		release := &rules.FaultDetectRelease{
			RuleRelease: rules.RuleRelease{
				Id:          id,
				ReleaseName: name,
				Description: description,
				ReleaseType: rules.ReleaseType(releaseType),
				Active:      active == 1,
				Version:     version,
				Valid:       true,
			},
			Rule: &ruleObj,
		}
		out = append(out, release)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteFaultDetectReleases implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) DeleteFaultDetectReleases(tx store.Tx, release *rules.FaultDetectRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	_, err := dbTx.Exec(`UPDATE fault_detect_rule_release SET flag = 1, mtime = sysdate() WHERE id = ?`, release.Id)
	return store.Error(err)
}

func fetchFaultDetectRulesRows(rows *sql.Rows) ([]*rules.FaultDetectRule, error) {
	defer rows.Close()
	var out []*rules.FaultDetectRule
	for rows.Next() {
		var fdRule rules.FaultDetectRule
		var flag int
		var ctime, mtime int64
		err := rows.Scan(&fdRule.ID, &fdRule.Name, &fdRule.Namespace, &fdRule.Revision,
			&fdRule.Description, &fdRule.DstService, &fdRule.DstNamespace,
			&fdRule.DstMethod, &fdRule.Rule, &flag, &ctime, &mtime)
		if err != nil {
			log.Errorf("[Store][database] fetch brief fault detect rule scan err: %s", err.Error())
			return nil, err
		}
		fdRule.CreateTime = time.Unix(ctime, 0)
		fdRule.ModifyTime = time.Unix(mtime, 0)
		fdRule.Valid = true
		if flag == 1 {
			fdRule.Valid = false
		}
		out = append(out, &fdRule)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch brief fault detect rule next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}

func genFaultDetectRuleSQL(query map[string]string) (string, []interface{}) {
	str := ""
	args := make([]interface{}, 0, len(query))
	var svcNamespaceQueryValue string
	var svcQueryValue string
	for key, value := range query {
		if len(value) == 0 {
			continue
		}
		if key == svcSpecificQueryKeyService {
			svcQueryValue = value
			continue
		}
		if key == svcSpecificQueryKeyNamespace {
			svcNamespaceQueryValue = value
			continue
		}
		storeKey := toUnderscoreName(key)
		if _, ok := blurQueryKeys[key]; ok {
			str += fmt.Sprintf(" and %s like ?", storeKey)
			args = append(args, "%"+value+"%")
		} else if key == exactName {
			str += " and name = ?"
			args = append(args, value)
		} else if key == excludeId {
			str += " and id != ?"
			args = append(args, value)
		} else {
			str += fmt.Sprintf(" and %s = ?", storeKey)
			args = append(args, value)
		}
	}
	if len(svcQueryValue) > 0 {
		str += " and (dst_service = ? or dst_service = '*')"
		args = append(args, svcQueryValue)
	}
	if len(svcNamespaceQueryValue) > 0 {
		str += " and (dst_namespace = ? or dst_namespace = '*')"
		args = append(args, svcNamespaceQueryValue)
	}
	return str, args
}

func (f *faultDetectRuleStore) getFaultDetectRulesCount(filter map[string]string) (uint32, error) {
	queryStr, args := genFaultDetectRuleSQL(filter)
	str := countFaultDetectSql + queryStr
	var total uint32
	err := f.master.QueryRow(str, args...).Scan(&total)
	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		log.Errorf("[Store][database] get fault detect rule count err: %s", err.Error())
		return 0, err
	default:
	}
	return total, nil
}

func (f *faultDetectRuleStore) getBriefFaultDetectRules(
	filter map[string]string, offset uint32, limit uint32) ([]*rules.FaultDetectRule, error) {
	queryStr, args := genFaultDetectRuleSQL(filter)
	args = append(args, offset, limit)
	str := queryFaultDetectBriefSql + queryStr + ` order by mtime desc limit ?, ?`

	rows, err := f.master.Query(str, args...)
	if err != nil {
		log.Errorf("[Store][database] query brief fault detect rule rules err: %s", err.Error())
		return nil, err
	}
	out, err := fetchBriefFaultDetectRules(rows)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func fetchBriefFaultDetectRules(rows *sql.Rows) ([]*rules.FaultDetectRule, error) {
	defer rows.Close()
	var out []*rules.FaultDetectRule
	for rows.Next() {
		var fdRule rules.FaultDetectRule
		var ctime, mtime int64
		err := rows.Scan(&fdRule.ID, &fdRule.Name, &fdRule.Namespace, &fdRule.Revision,
			&fdRule.Description, &fdRule.DstService, &fdRule.DstNamespace,
			&fdRule.DstMethod, &ctime, &mtime)
		if err != nil {
			log.Errorf("[Store][database] fetch brief fault detect rule scan err: %s", err.Error())
			return nil, err
		}
		fdRule.CreateTime = time.Unix(ctime, 0)
		fdRule.ModifyTime = time.Unix(mtime, 0)
		out = append(out, &fdRule)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch brief fault detect rule next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}

func (f *faultDetectRuleStore) getFullFaultDetectRules(
	filter map[string]string, offset uint32, limit uint32) ([]*rules.FaultDetectRule, error) {
	queryStr, args := genFaultDetectRuleSQL(filter)
	args = append(args, offset, limit)
	str := queryFaultDetectFullSql + queryStr + ` order by mtime desc limit ?, ?`

	rows, err := f.master.Query(str, args...)
	if err != nil {
		log.Errorf("[Store][database] query brief fault detect rules err: %s", err.Error())
		return nil, err
	}
	out, err := fetchFullFaultDetectRules(rows)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func fetchFullFaultDetectRules(rows *sql.Rows) ([]*rules.FaultDetectRule, error) {
	defer rows.Close()
	var out []*rules.FaultDetectRule
	for rows.Next() {
		var fdRule rules.FaultDetectRule
		var ctime, mtime int64
		err := rows.Scan(&fdRule.ID, &fdRule.Name, &fdRule.Namespace, &fdRule.Revision,
			&fdRule.Description, &fdRule.DstService, &fdRule.DstNamespace,
			&fdRule.DstMethod, &fdRule.Rule, &ctime, &mtime)
		if err != nil {
			log.Errorf("[Store][database] fetch brief fault detect rule scan err: %s", err.Error())
			return nil, err
		}
		fdRule.CreateTime = time.Unix(ctime, 0)
		fdRule.ModifyTime = time.Unix(mtime, 0)
		out = append(out, &fdRule)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch brief fault detect rule next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}
