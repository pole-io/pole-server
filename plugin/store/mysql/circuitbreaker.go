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
	"strconv"
	"strings"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
)

const (
	labelCreateCircuitBreakerRule = "createCircuitBreakerRule"
	labelUpdateCircuitBreakerRule = "updateCircuitBreakerRule"
	labelDeleteCircuitBreakerRule = "deleteCircuitBreakerRule"
)

const (
	insertCircuitBreakerRuleSql = `insert into circuitbreaker_rule(
			id, name, namespace, enable, revision, description, level, src_service, src_namespace, 
			dst_service, dst_namespace, dst_method, config, ctime, mtime, etime)
			values(?,?,?,?,?,?,?,?,?,?,?,?,?, sysdate(),sysdate(), %s)`
	updateCircuitBreakerRuleSql = `update circuitbreaker_rule set name = ?, namespace=?, enable = ?, revision= ?,
			description = ?, level = ?, src_service = ?, src_namespace = ?,
            dst_service = ?, dst_namespace = ?, dst_method = ?,
			config = ?, mtime = sysdate(), etime=%s where id = ?`
	deleteCircuitBreakerRuleSql    = `update circuitbreaker_rule set flag = 1, mtime = sysdate() where id = ?`
	countCircuitBreakerRuleSql     = `select count(*) from circuitbreaker_rule where flag = 0`
	queryCircuitBreakerRuleFullSql = `select id, name, namespace, enable, revision, description, level, src_service, 
			src_namespace, dst_service, dst_namespace, dst_method, config, unix_timestamp(ctime), unix_timestamp(mtime), 
			unix_timestamp(etime) from circuitbreaker_rule where flag = 0`
	queryCircuitBreakerRuleBriefSql = `select id, name, namespace, enable, revision, level, src_service, src_namespace, 
			dst_service, dst_namespace, dst_method, unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime)
			from circuitbreaker_rule where flag = 0`
	queryCircuitBreakerRuleCacheSql = `select id, name, namespace, enable, revision, description, level, src_service, 
			src_namespace, dst_service, dst_namespace, dst_method, config, flag, unix_timestamp(ctime), 
			unix_timestamp(mtime), unix_timestamp(etime) from circuitbreaker_rule where mtime > FROM_UNIXTIME(?)`
)

var _ store.CircuitBreakerStore = (*circuitBreakerStore)(nil)

const (
	labelCreateCircuitBreakerRuleOld    = "createCircuitBreakerRuleOld"
	labelTagCircuitBreakerRuleOld       = "tagCircuitBreakerRuleOld"
	labelDeleteTagCircuitBreakerRuleOld = "deleteTagCircuitBreakerRuleOld"
	labelReleaseCircuitBreakerRuleOld   = "releaseCircuitBreakerRuleOld"
	labelUnbindCircuitBreakerRuleOld    = "unbindCircuitBreakerRuleOld"
	labelUpdateCircuitBreakerRuleOld    = "updateCircuitBreakerRuleOld"
	labelDeleteCircuitBreakerRuleOld    = "deleteCircuitBreakerRuleOld"
)

// circuitBreakerStore 的实现
type circuitBreakerStore struct {
	master *BaseDB
	slave  *BaseDB
}

func (c *circuitBreakerStore) CreateCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error {
	err := RetryTransaction(labelCreateCircuitBreakerRule, func() error {
		return c.createCircuitBreakerRule(cbRule)
	})

	return store.Error(err)
}

func (c *circuitBreakerStore) createCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error {
	return c.master.processWithTransaction(labelCreateCircuitBreakerRule, func(tx *BaseTx) error {
		etimeStr := buildEtimeStr(cbRule.Enable)
		str := fmt.Sprintf(insertCircuitBreakerRuleSql, etimeStr)
		if _, err := tx.Exec(str, cbRule.ID, cbRule.Name, cbRule.Namespace, cbRule.Enable, cbRule.Revision,
			cbRule.Description, cbRule.Level, cbRule.SrcService, cbRule.SrcNamespace, cbRule.DstService,
			cbRule.DstNamespace, cbRule.DstMethod, cbRule.Rule); err != nil {
			log.Errorf("[Store][database] fail to %s exec sql, err: %s", labelCreateCircuitBreakerRule, err.Error())
			return err
		}
		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] fail to %s commit tx, rule(%+v) commit tx err: %s",
				labelCreateCircuitBreakerRule, cbRule, err.Error())
			return err
		}
		return nil
	})
}

// UpdateCircuitBreakerRule 更新熔断规则
func (c *circuitBreakerStore) UpdateCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error {
	err := RetryTransaction(labelUpdateCircuitBreakerRule, func() error {
		return c.updateCircuitBreakerRule(cbRule)
	})

	return store.Error(err)
}

func (c *circuitBreakerStore) updateCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error {
	return c.master.processWithTransaction(labelUpdateCircuitBreakerRule, func(tx *BaseTx) error {
		etimeStr := buildEtimeStr(cbRule.Enable)
		str := fmt.Sprintf(updateCircuitBreakerRuleSql, etimeStr)
		if _, err := tx.Exec(str, cbRule.Name, cbRule.Namespace, cbRule.Enable,
			cbRule.Revision, cbRule.Description, cbRule.Level, cbRule.SrcService, cbRule.SrcNamespace,
			cbRule.DstService, cbRule.DstNamespace, cbRule.DstMethod, cbRule.Rule, cbRule.ID); err != nil {
			log.Errorf("[Store][database] fail to %s exec sql, err: %s", labelUpdateCircuitBreakerRule, err.Error())
			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] fail to %s commit tx, rule(%+v) commit tx err: %s",
				labelUpdateCircuitBreakerRule, cbRule, err.Error())
			return err
		}

		return nil
	})
}

// DeleteCircuitBreakerRule 删除熔断规则
func (c *circuitBreakerStore) DeleteCircuitBreakerRule(id string) error {
	err := RetryTransaction("deleteCircuitBreakerRule", func() error {
		return c.deleteCircuitBreakerRule(id)
	})

	return store.Error(err)
}

func (c *circuitBreakerStore) deleteCircuitBreakerRule(id string) error {
	return c.master.processWithTransaction(labelDeleteCircuitBreakerRule, func(tx *BaseTx) error {
		if _, err := tx.Exec(deleteCircuitBreakerRuleSql, id); err != nil {
			log.Errorf(
				"[Store][database] fail to %s exec sql, err: %s", labelDeleteCircuitBreakerRule, err.Error())
			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] fail to %s commit tx, rule(%s) commit tx err: %s",
				labelDeleteCircuitBreakerRule, id, err.Error())
			return err
		}
		return nil
	})
}

// HasCircuitBreakerRule check circuitbreaker rule exists
func (c *circuitBreakerStore) HasCircuitBreakerRule(id string) (bool, error) {
	queryParams := map[string]string{"id": id}
	count, err := c.getCircuitBreakerRulesCount(queryParams)
	if nil != err {
		return false, err
	}
	return count > 0, nil
}

// HasCircuitBreakerRuleByName check circuitbreaker rule exists by name
func (c *circuitBreakerStore) HasCircuitBreakerRuleByName(name string, namespace string) (bool, error) {
	queryParams := map[string]string{exactName: name, "namespace": namespace}
	count, err := c.getCircuitBreakerRulesCount(queryParams)
	if nil != err {
		return false, err
	}
	return count > 0, nil
}

// HasCircuitBreakerRuleByNameExcludeId check circuitbreaker rule exists by name exclude id
func (c *circuitBreakerStore) HasCircuitBreakerRuleByNameExcludeId(
	name string, namespace string, id string) (bool, error) {
	queryParams := map[string]string{exactName: name, "namespace": namespace, excludeId: id}
	count, err := c.getCircuitBreakerRulesCount(queryParams)
	if nil != err {
		return false, err
	}
	return count > 0, nil
}

func fetchCircuitBreakerRuleRows(rows *sql.Rows) ([]*rules.CircuitBreakerRule, error) {
	defer rows.Close()
	var out []*rules.CircuitBreakerRule
	for rows.Next() {
		var cbRule rules.CircuitBreakerRule
		var flag int
		var ctime, mtime, etime int64
		err := rows.Scan(&cbRule.ID, &cbRule.Name, &cbRule.Namespace, &cbRule.Enable, &cbRule.Revision,
			&cbRule.Description, &cbRule.Level, &cbRule.SrcService, &cbRule.SrcNamespace, &cbRule.DstService,
			&cbRule.DstNamespace, &cbRule.DstMethod, &cbRule.Rule, &flag, &ctime, &mtime, &etime)
		if err != nil {
			log.Errorf("[Store][database] fetch circuitbreaker rule scan err: %s", err.Error())
			return nil, err
		}
		cbRule.CreateTime = time.Unix(ctime, 0)
		cbRule.ModifyTime = time.Unix(mtime, 0)
		cbRule.EnableTime = time.Unix(etime, 0)
		cbRule.Valid = true
		if flag == 1 {
			cbRule.Valid = false
		}
		out = append(out, &cbRule)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch circuitbreaker rule next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}

func (c *circuitBreakerStore) GetCircuitBreakerRules(
	filter map[string]string, offset uint32, limit uint32) (uint32, []*rules.CircuitBreakerRule, error) {
	var out []*rules.CircuitBreakerRule
	var err error

	bValue, ok := filter[briefSearch]
	var isBrief = ok && strings.ToLower(bValue) == "true"
	delete(filter, briefSearch)

	if isBrief {
		out, err = c.getBriefCircuitBreakerRules(filter, offset, limit)
	} else {
		out, err = c.getFullCircuitBreakerRules(filter, offset, limit)
	}
	if err != nil {
		return 0, nil, err
	}
	num, err := c.getCircuitBreakerRulesCount(filter)
	if err != nil {
		return 0, nil, err
	}
	return num, out, nil
}

func (c *circuitBreakerStore) getBriefCircuitBreakerRules(
	filter map[string]string, offset uint32, limit uint32) ([]*rules.CircuitBreakerRule, error) {
	queryStr, args := genCircuitBreakerRuleSQL(filter)
	args = append(args, offset, limit)
	str := queryCircuitBreakerRuleBriefSql + queryStr + ` order by mtime desc limit ?, ?`

	rows, err := c.master.Query(str, args...)
	if err != nil {
		log.Errorf("[Store][database] query brief circuitbreaker rules err: %s", err.Error())
		return nil, err
	}
	out, err := fetchBriefCircuitBreakerRules(rows)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var blurQueryKeys = map[string]bool{
	"name":         true,
	"description":  true,
	"srcService":   true,
	"srcNamespace": true,
	"dstService":   true,
	"dstNamespace": true,
	"dstMethod":    true,
}

const (
	svcSpecificQueryKeyService   = "service"
	svcSpecificQueryKeyNamespace = "serviceNamespace"
	exactName                    = "exactName"
	excludeId                    = "excludeId"
)

func placeholders(n int) string {
	var b strings.Builder
	for i := 0; i < n-1; i++ {
		b.WriteString("?,")
	}
	if n > 0 {
		b.WriteString("?")
	}
	return b.String()
}

func genCircuitBreakerRuleSQL(query map[string]string) (string, []interface{}) {
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
		} else if key == "enable" {
			str += fmt.Sprintf(" and %s = ?", storeKey)
			arg, _ := strconv.ParseBool(value)
			args = append(args, arg)
		} else if key == "level" {
			tokens := strings.Split(value, ",")
			str += fmt.Sprintf(" and %s in (%s)", storeKey, placeholders(len(tokens)))
			for _, token := range tokens {
				args = append(args, token)
			}
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
		str += " and (dst_service = ? or dst_service = '*' or src_service = ? or src_service = '*')"
		args = append(args, svcQueryValue, svcQueryValue)
	}
	if len(svcNamespaceQueryValue) > 0 {
		str += " and (dst_namespace = ? or dst_namespace = '*' or dst_namespace = ? or dst_namespace = '*')"
		args = append(args, svcNamespaceQueryValue, svcNamespaceQueryValue)
	}
	return str, args
}

// fetchBriefRateLimitRows fetch the brief ratelimit list
func fetchBriefCircuitBreakerRules(rows *sql.Rows) ([]*rules.CircuitBreakerRule, error) {
	defer rows.Close()
	var out []*rules.CircuitBreakerRule
	for rows.Next() {
		var cbRule rules.CircuitBreakerRule
		var ctime, mtime, etime int64
		err := rows.Scan(&cbRule.ID, &cbRule.Name, &cbRule.Namespace, &cbRule.Enable, &cbRule.Revision,
			&cbRule.Level, &cbRule.SrcService, &cbRule.SrcNamespace, &cbRule.DstService, &cbRule.DstNamespace,
			&cbRule.DstMethod, &ctime, &mtime, &etime)
		if err != nil {
			log.Errorf("[Store][database] fetch brief circuitbreaker rule scan err: %s", err.Error())
			return nil, err
		}
		cbRule.CreateTime = time.Unix(ctime, 0)
		cbRule.ModifyTime = time.Unix(mtime, 0)
		cbRule.EnableTime = time.Unix(etime, 0)
		out = append(out, &cbRule)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch brief circuitbreaker rule next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}

func (c *circuitBreakerStore) getFullCircuitBreakerRules(
	filter map[string]string, offset uint32, limit uint32) ([]*rules.CircuitBreakerRule, error) {
	queryStr, args := genCircuitBreakerRuleSQL(filter)
	args = append(args, offset, limit)
	str := queryCircuitBreakerRuleFullSql + queryStr + ` order by mtime desc limit ?, ?`

	rows, err := c.master.Query(str, args...)
	if err != nil {
		log.Errorf("[Store][database] query brief circuitbreaker rules err: %s", err.Error())
		return nil, err
	}
	out, err := fetchFullCircuitBreakerRules(rows)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func fetchFullCircuitBreakerRules(rows *sql.Rows) ([]*rules.CircuitBreakerRule, error) {
	defer rows.Close()
	var out []*rules.CircuitBreakerRule
	for rows.Next() {
		var cbRule rules.CircuitBreakerRule
		var ctime, mtime, etime int64
		err := rows.Scan(&cbRule.ID, &cbRule.Name, &cbRule.Namespace, &cbRule.Enable, &cbRule.Revision,
			&cbRule.Description, &cbRule.Level, &cbRule.SrcService, &cbRule.SrcNamespace, &cbRule.DstService,
			&cbRule.DstNamespace, &cbRule.DstMethod, &cbRule.Rule, &ctime, &mtime, &etime)
		if err != nil {
			log.Errorf("[Store][database] fetch full circuitbreaker rule scan err: %s", err.Error())
			return nil, err
		}
		cbRule.CreateTime = time.Unix(ctime, 0)
		cbRule.ModifyTime = time.Unix(mtime, 0)
		cbRule.EnableTime = time.Unix(etime, 0)
		out = append(out, &cbRule)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch full circuitbreaker rule next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}

func (c *circuitBreakerStore) getCircuitBreakerRulesCount(filter map[string]string) (uint32, error) {
	queryStr, args := genCircuitBreakerRuleSQL(filter)
	str := countCircuitBreakerRuleSql + queryStr
	var total uint32
	err := c.master.QueryRow(str, args...).Scan(&total)
	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		log.Errorf("[Store][database] get circuitbreaker rule count err: %s", err.Error())
		return 0, err
	default:
	}
	return total, nil
}

// GetMoreCircuitBreakers list circuitbreaker rules by query
func (c *circuitBreakerStore) GetMoreCircuitBreakers(mtime time.Time, firstUpdate bool) ([]*rules.CircuitBreakerRule, error) {
	str := queryCircuitBreakerRuleCacheSql
	if firstUpdate {
		str += " and flag != 1"
	}
	rows, err := c.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] query circuitbreaker rules with mtime err: %s", err.Error())
		return nil, err
	}
	cbRules, err := fetchCircuitBreakerRuleRows(rows)
	if err != nil {
		return nil, err
	}
	return cbRules, nil
}

func (c *circuitBreakerStore) GetCircuitBreakerRule(id string) (*rules.CircuitBreakerRule, error) {
	querySql := `SELECT id, name, namespace, enable, revision
	, description, level, src_service, src_namespace, dst_service
	, dst_namespace, dst_method, config, unix_timestamp(ctime)
	, unix_timestamp(mtime), unix_timestamp(etime)
FROM circuitbreaker_rule
WHERE id = ?
	AND flag = 0
FOR UPDATE`
	row := c.master.QueryRow(querySql, id)
	var cbRule rules.CircuitBreakerRule
	var ctime, mtime, etime int64
	err := row.Scan(&cbRule.ID, &cbRule.Name, &cbRule.Namespace, &cbRule.Enable, &cbRule.Revision,
		&cbRule.Description, &cbRule.Level, &cbRule.SrcService, &cbRule.SrcNamespace, &cbRule.DstService,
		&cbRule.DstNamespace, &cbRule.DstMethod, &cbRule.Rule, &ctime, &mtime, &etime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	cbRule.CreateTime = time.Unix(ctime, 0)
	cbRule.ModifyTime = time.Unix(mtime, 0)
	cbRule.EnableTime = time.Unix(etime, 0)
	cbRule.Valid = true
	return &cbRule, nil
}

// LockCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) LockCircuitBreakerRule(tx store.Tx, name string) (*rules.CircuitBreakerRule, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, namespace, enable, revision
	, description, level, src_service, src_namespace, dst_service
	, dst_namespace, dst_method, config, unix_timestamp(ctime)
	, unix_timestamp(mtime), unix_timestamp(etime)
FROM circuitbreaker_rule
WHERE name = ?
	AND flag = 0
FOR UPDATE`
	row := dbTx.QueryRow(querySql, name)
	var cbRule rules.CircuitBreakerRule
	var ctime, mtime, etime int64
	err := row.Scan(&cbRule.ID, &cbRule.Name, &cbRule.Namespace, &cbRule.Enable, &cbRule.Revision,
		&cbRule.Description, &cbRule.Level, &cbRule.SrcService, &cbRule.SrcNamespace, &cbRule.DstService,
		&cbRule.DstNamespace, &cbRule.DstMethod, &cbRule.Rule, &ctime, &mtime, &etime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	cbRule.CreateTime = time.Unix(ctime, 0)
	cbRule.ModifyTime = time.Unix(mtime, 0)
	cbRule.EnableTime = time.Unix(etime, 0)
	cbRule.Valid = true
	return &cbRule, nil
}

// ActiveCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) ActiveCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := c.inactiveCircuitBreakerRelease(dbTx, release)
	if err != nil {
		return store.Error(err)
	}
	args := []any{maxVersion + 1, release.ReleaseName, release.RuleName}
	_, err = dbTx.Exec(`UPDATE circuitbreaker_rule_release
SET active = 1, version = ?, mtime = sysdate()
WHERE name = ?
	AND rule_name = ?`, args...)
	return store.Error(err)
}

// GetCircuitBreakerRuleVersions .
func (c *circuitBreakerStore) GetCircuitBreakerRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	countSql := `SELECT COUNT(*) FROM circuitbreaker_rule_release WHERE rule_name = ? AND flag = 0`
	row := c.slave.QueryRow(countSql, filter["rule_name"])
	var count uint64
	if err := row.Scan(&count); err != nil {
		log.Errorf("[store][mysql][circuitbreaker] query circuitbreaker versions count err: %s", err.Error())
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}

	querySql := `SELECT id, name, rule_id, rule_name, flag, active, version, description, release_type, ctime, mtime
	FROM circuitbreaker_rule_release
	WHERE rule_id = ?
		AND flag = 0 ORDER BY version DESC LIMIT ?, ?`
	rows, err := c.slave.Query(querySql, filter["rule_name"], offset, limit)
	if err != nil {
		log.Errorf("[store][mysql][circuitbreaker] query circuitbreaker rule versions err: %s", err.Error())
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
			log.Errorf("[store][mysql][circuitbreaker] fetch circuitbreaker rule versions scan err: %s", err.Error())
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

// GetReleaseCircuitBreakerRule 获取处于使用状态的熔断规则
func (c *circuitBreakerStore) GetReleaseCircuitBreakerRule(tx store.Tx, req *rules.RuleRelease) (*rules.CircuitBreakerRelease, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, version
	, active, description, release_type
FROM circuitbreaker_rule_release
WHERE name = ?
	AND rule_name = ?
	AND release_type = ?
	AND flag = 0
LIMIT 1`
	row := dbTx.QueryRow(querySql, req.ReleaseName, req.RuleName, req.ReleaseType)
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
		return nil, store.Error(err)
	}
	var ruleObj rules.CircuitBreakerRule
	if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
		return nil, store.Error(err)
	}
	return &rules.CircuitBreakerRelease{
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

// GetActiveCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) GetActiveCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) (*rules.CircuitBreakerRelease, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, version
	, active, description, release_type
FROM circuitbreaker_rule_release
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
		return nil, store.Error(err)
	}
	var ruleObj rules.CircuitBreakerRule
	if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
		return nil, store.Error(err)
	}
	return &rules.CircuitBreakerRelease{
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

// InactiveCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) InactiveCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	args := []any{release.RuleName, release.ReleaseName, release.ReleaseType}
	_, err := dbTx.Exec(`UPDATE circuitbreaker_rule_release
SET active = 0, mtime = sysdate()
WHERE rule_name = ?
	AND name = ?
	AND release_type = ?
	AND active = 1`, args...)
	return store.Error(err)
}

func (c *circuitBreakerStore) inactiveCircuitBreakerRelease(tx *BaseTx, release *rules.CircuitBreakerRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.RuleName, release.ReleaseType}
	//	先取消所有 active == true 的记录
	if _, err := tx.Exec(`UPDATE circuitbreaker_rule_release
SET active = 0, mtime = sysdate()
WHERE rule_name = ?
	AND active = 1
	AND release_type = ?`, args...); err != nil {
		return 0, err
	}
	return c.selectMaxVersion(tx, release)
}

func (c *circuitBreakerStore) selectMaxVersion(tx *BaseTx, release *rules.CircuitBreakerRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.Rule.Name}
	var maxVersion uint64
	//	查询当前 release 的最大版本号
	if err := tx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM circuitbreaker_rule_release WHERE rule_name = ?",
		args...).Scan(&maxVersion); err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// PublishCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) PublishCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	if release.ReleaseName == "" || release.ReleaseType == "" {
		return errors.New(
			"[store][mysql][circuitbreaker] publish circuit breaker rule missing some params",
		)
	}
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := c.inactiveCircuitBreakerRelease(dbTx, release)
	if err != nil {
		return store.Error(err)
	}

	ruleJson, err := json.Marshal(release.Rule)
	if err != nil {
		return store.Error(err)
	}
	// 3. 插入新发布并激活
	insertSql := `INSERT INTO circuitbreaker_rule_release (
		id, name, rule_name, rule, version, active, description, release_type, ctime, mtime
	) VALUES (?, ?, ?, ?, ?, 1, ?, ?, sysdate(), sysdate())`
	_, err = dbTx.Exec(
		insertSql,
		release.Id,
		release.ReleaseName,
		"", // rule_name 暂未使用
		string(ruleJson),
		maxVersion+1,
		release.Description,
		release.ReleaseType,
	)
	return store.Error(err)
}

func (c *circuitBreakerStore) GetMoreCircuitBreakerReleases(mtime time.Time, firstUpdate bool) ([]*rules.CircuitBreakerRelease, error) {
	str := `SELECT id, name, rule_name, rule, version, active, description, release_type,
		unix_timestamp(ctime), unix_timestamp(mtime)
	FROM circuitbreaker_rule_release
	WHERE mtime > FROM_UNIXTIME(?)
		AND flag = 0`
	if firstUpdate {
		str += " AND active != 1"
	}
	rows, err := c.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] query circuitbreaker releases with mtime err: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	var releases []*rules.CircuitBreakerRelease
	for rows.Next() {
		var release rules.CircuitBreakerRelease
		var ruleStr string
		var ctime, mtime int64
		err := rows.Scan(&release.Id, &release.ReleaseName, &release.RuleName, &ruleStr, &release.Version,
			&release.Active, &release.Description, &release.ReleaseType, &ctime, &mtime)
		if err != nil {
			log.Errorf("[Store][database] fetch circuitbreaker release scan err: %s", err.Error())
			return nil, err
		}
		release.Ctime = time.Unix(ctime, 0)
		release.Mtime = time.Unix(mtime, 0)
		release.Valid = true

		if err := json.Unmarshal([]byte(ruleStr), &release.Rule); err != nil {
			log.Errorf("[Store][database] unmarshal circuitbreaker rule err: %s", err.Error())
			return nil, store.Error(err)
		}
		releases = append(releases, &release)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch circuitbreaker release next err: %s", err.Error())
		return nil, err
	}
	return releases, nil
}
