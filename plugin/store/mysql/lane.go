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
	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
)

var _ store.LaneStore = (*laneStore)(nil)

type laneStore struct {
	master *BaseDB
	slave  *BaseDB
}

// AddLaneGroup 添加泳道组
func (l *laneStore) AddLaneGroup(tx store.Tx, item *ruletypes.LaneGroup) error {
	if err := l.cleanSoftDeletedRules(); err != nil {
		return err
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	// 先清理无效的泳道组
	if _, err := dbTx.Exec("DELETE FROM lane_group WHERE name = ? AND flag = 1", item.Name); err != nil {
		log.Error("[Store][Lane] clean invalid lane group", zap.String("id", item.ID),
			zap.String("name", item.Name), zap.Error(err))
		return err
	}
	args := []interface{}{
		item.ID,
		item.Name,
		item.Rule,
		item.Revision,
		item.Description,
	}

	addSql := `
INSERT INTO lane_group (id, name, rule, revision, description, flag
	, ctime, mtime)
VALUES (?, ?, ?, ?, ?, 0, sysdate(), sysdate())
`
	if _, err := dbTx.Exec(addSql, args...); err != nil {
		log.Error("[Store][Lane] add lane group", zap.String("id", item.ID),
			zap.String("name", item.Name), zap.Error(err))
		return store.Error(err)
	}
	return nil
}

// UpdateLaneGroup 更新泳道组
func (l *laneStore) UpdateLaneGroup(tx store.Tx, item *ruletypes.LaneGroup) error {
	if err := l.cleanSoftDeletedRules(); err != nil {
		return err
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	args := []interface{}{
		item.Rule,
		item.Revision,
		item.Description,
		item.ID,
	}

	addSql := "UPDATE lane_group SET rule = ?, revision = ?, description = ?, mtime = sysdate() WHERE id = ?"
	if _, err := dbTx.Exec(addSql, args...); err != nil {
		log.Error("[Store][Lane] update lane group", zap.String("id", item.ID),
			zap.String("name", item.Name), zap.Error(err))
		return store.Error(err)
	}
	return nil
}

// GetLaneGroup 查询泳道组
func (l *laneStore) GetLaneGroup(name string) (*ruletypes.LaneGroup, error) {
	querySql := `
SELECT id, name, rule, description, revision, flag, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime) FROM lane_group WHERE flag = 0 AND name = ?
`
	result := make([]*ruletypes.LaneGroup, 0, 1)
	err := l.master.processWithTransaction("GetLaneGroup", func(tx *BaseTx) error {
		rows, err := tx.Query(querySql, name)
		if err != nil {
			log.Error("[Store][Lane] select one lane group", zap.String("querySql", querySql), zap.Error(err))
			return err
		}
		if err := transferLaneGroups(rows, func(group *ruletypes.LaneGroup) {
			result = append(result, group)
		}); err != nil {
			log.Error("[Store][Lane] transfer one lane group row", zap.Error(err))
			return err
		}
		return tx.Commit()
	})
	if err != nil {
		return nil, store.Error(err)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

// GetLaneGroupByID .
func (l *laneStore) GetLaneGroupByID(id string) (*ruletypes.LaneGroup, error) {
	querySql := `
SELECT id, name, rule, description, revision, flag, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime) FROM lane_group WHERE flag = 0 AND id = ?
`
	result := make([]*ruletypes.LaneGroup, 0, 1)
	err := l.master.processWithTransaction("GetLaneGroupByID", func(tx *BaseTx) error {
		rows, err := tx.Query(querySql, id)
		if err != nil {
			log.Error("[Store][Lane] select one lane group", zap.String("querySql", querySql), zap.Error(err))
			return err
		}
		if err := transferLaneGroups(rows, func(group *ruletypes.LaneGroup) {
			result = append(result, group)
		}); err != nil {
			log.Error("[Store][Lane] transfer one lane group row", zap.Error(err))
			return err
		}
		return tx.Commit()
	})
	if err != nil {
		return nil, store.Error(err)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

func (l *laneStore) LockLaneGroup(tx store.Tx, name string) (*ruletypes.LaneGroup, error) {
	querySql := `
SELECT id, name, rule, description, revision
	, flag, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
FROM lane_group
WHERE flag = 0
	AND name = ?
FOR UPDATE
`
	dbTx := tx.GetDelegateTx().(*BaseTx)
	result := make([]*ruletypes.LaneGroup, 0, 1)
	rows, err := dbTx.Query(querySql, name)
	if err != nil {
		log.Error("[Store][Lane] select one lane group", zap.String("querySql", querySql), zap.String("name", name),
			zap.Error(err))
		return nil, err
	}
	if err := transferLaneGroups(rows, func(group *ruletypes.LaneGroup) {
		result = append(result, group)
	}); err != nil {
		log.Error("[Store][Lane] transfer one lane group row", zap.String("name", name), zap.Error(err))
		return nil, store.Error(err)
	}
	if len(result) == 0 {
		return nil, nil
	}
	rules, err := l.getLaneRulesByGroup(dbTx, []string{name})
	if err != nil {
		log.Error("[Store][Lane] load lane_group all lane_rule", zap.String("name", name), zap.Error(err))
		return nil, store.Error(err)
	}
	if len(rules) != 0 {
		result[0].LaneRules = rules[name]
	}
	return result[0], nil
}

// GetLaneGroups 查询泳道组
func (l *laneStore) GetLaneGroups(filter map[string]string, offset, limit uint32) (uint32, []*ruletypes.LaneGroup, error) {
	countSql := `
SELECT COUNT(*) FROM lane_group WHERE flag = 0 
`
	querySql := `
SELECT id, name, rule, description, revision, flag, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime) FROM lane_group WHERE flag = 0
`
	conditions := []string{}
	args := []any{}
	for k, v := range filter {
		switch k {
		case "name":
			if pv, ok := matchs.ParseWildName(v); ok {
				conditions = append(conditions, "name = ?")
				args = append(args, pv)
			} else {
				conditions = append(conditions, "name LIKE ?")
				args = append(args, "%"+v+"%")
			}
		case "id":
			conditions = append(conditions, "id = ?")
			args = append(args, v)
		}
	}
	if len(conditions) > 0 {
		countSql += " AND " + strings.Join(conditions, " AND ")
		querySql += " AND " + strings.Join(conditions, " AND ")
	}

	querySql += fmt.Sprintf(" ORDER BY %s %s LIMIT ?, ? ", filter["order_field"], filter["order_type"])

	var count int64
	var result []*ruletypes.LaneGroup

	err := l.master.processWithTransaction("GetLaneGroups", func(tx *BaseTx) error {
		row := tx.QueryRow(countSql, args...)
		if err := row.Scan(&count); err != nil {
			log.Error("[Store][Lane] count lane group", zap.String("countSql", countSql), zap.Error(err))
			return err
		}

		// count 阶段不需要分页参数，因此留到这里在进行追加
		args = append(args, offset, limit)
		rows, err := tx.Query(querySql, args...)
		if err != nil {
			log.Error("[Store][Lane] select lane group", zap.String("querySql", querySql), zap.Error(err))
			return err
		}
		if err := transferLaneGroups(rows, func(group *ruletypes.LaneGroup) {
			result = append(result, group)
		}); err != nil {
			log.Error("[Store][Lane] transfer lane group row", zap.Error(err))
			return err
		}
		brief := filter[briefSearch] == "true"
		if !brief {
			names := make([]string, 0, len(result))
			for i := range result {
				names = append(names, result[i].Name)
			}
			rules, err := l.getLaneRulesByGroup(tx, names)
			if err != nil {
				return err
			}
			for i := range result {
				item := result[i]
				item.LaneRules = rules[item.Name]
			}
		}
		return tx.Commit()
	})
	if err != nil {
		return 0, nil, store.Error(err)
	}
	return uint32(count), result, nil
}

// DeleteLaneGroup 删除泳道组
func (l *laneStore) DeleteLaneGroup(id string) error {
	err := l.master.processWithTransaction("DeleteLaneGroup", func(tx *BaseTx) error {
		args := []any{
			id,
		}

		addSql := "UPDATE lane_rule SET flag = 1, mtime = sysdate() WHERE group_name IN (SELECT name FROM lane_group WHERE id = ?)"
		if _, err := tx.Exec(addSql, args...); err != nil {
			log.Error("[Store][Lane] delete lane group", zap.String("id", id), zap.Error(err))
			return err
		}

		addSql = "UPDATE lane_group SET flag = 1, mtime = sysdate() WHERE id = ?"
		if _, err := tx.Exec(addSql, args...); err != nil {
			log.Error("[Store][Lane] delete lane group", zap.String("id", id), zap.Error(err))
			return err
		}
		return tx.Commit()
	})
	return store.Error(err)
}

// getLaneRulesByGroup .
func (l *laneStore) getLaneRulesByGroup(tx *BaseTx, names []string) (map[string]map[string]*ruletypes.LaneRule, error) {
	if len(names) == 0 {
		return map[string]map[string]*ruletypes.LaneRule{}, nil
	}

	querySql := `
SELECT id, name, group_name, rule, revision, priority, description, enable, flag, UNIX_TIMESTAMP(ctime), 
UNIX_TIMESTAMP(etime), UNIX_TIMESTAMP(mtime) FROM lane_rule WHERE flag = 0 AND group_name IN (%s)
`
	querySql = fmt.Sprintf(querySql, placeholders(len(names)))

	rows, err := tx.Query(querySql, StringsToArgs(names)...)
	if err != nil {
		log.Error("[Store][Lane] fetch lane group all lane_rules", zap.String("sql", querySql), zap.Error(err))
		return nil, store.Error(err)
	}
	result := make(map[string]map[string]*ruletypes.LaneRule, len(names))
	if err := transferLaneRules(rows, func(rule *ruletypes.LaneRule) {
		if _, ok := result[rule.LaneGroup]; !ok {
			result[rule.LaneGroup] = make(map[string]*ruletypes.LaneRule, 32)
		}
		result[rule.LaneGroup][rule.ID] = rule
	}); err != nil {
		return nil, store.Error(err)
	}
	return result, nil
}

// GetMoreLaneGroups 获取泳道规则列表到缓存层
func (l *laneStore) GetMoreLaneGroups(mtime time.Time, firstUpdate bool) (map[string]*ruletypes.LaneGroup, error) {
	if firstUpdate {
		mtime = time.Unix(0, 1)
	}
	deltaGroupSql := `
SELECT id, name, rule, description
	, revision, flag, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
FROM lane_group
WHERE mtime >= FROM_UNIXTIME(?)
`

	deltaRuleSql := `
SELECT
  lr.id,
  lr.name,
  group_name,
  lr.rule,
  lr.revision,
  lr.priority,
  lr.description,
  enable,
  lr.flag,
  UNIX_TIMESTAMP(lr.ctime),
  UNIX_TIMESTAMP(lr.etime),
  UNIX_TIMESTAMP(lr.mtime)
FROM
  lane_rule lr
  LEFT JOIN lane_group lg ON lr.group_name = lg.name
WHERE
  lg.mtime >= FROM_UNIXTIME(?)
`
	deltaGroups := map[string]*ruletypes.LaneGroup{}
	var deltaRules []*ruletypes.LaneRule

	err := l.slave.processWithTransaction("GetMoreLaneGroups", func(tx *BaseTx) error {
		rows, err := tx.Query(deltaGroupSql, mtime)
		if err != nil {
			log.Error("[Store][Lane] delta lane group", zap.String("querySql", deltaGroupSql), zap.Error(err))
			return err
		}
		if err := transferLaneGroups(rows, func(group *ruletypes.LaneGroup) {
			group.LaneRules = make(map[string]*ruletypes.LaneRule)
			deltaGroups[group.Name] = group
		}); err != nil {
			log.Error("[Store][Lane] transfer lane group row", zap.Error(err))
			return err
		}
		if len(deltaGroups) > 0 {
			// 走 join 操作获取每个 group 下的 lane_rule 列表
			rows, err := tx.Query(deltaRuleSql, mtime)
			if err != nil {
				log.Error("[Store][Lane] delta lane rule", zap.String("querySql", deltaRuleSql), zap.Error(err))
				return err
			}
			if err := transferLaneRules(rows, func(rule *ruletypes.LaneRule) {
				deltaRules = append(deltaRules, rule)
			}); err != nil {
				log.Error("[Store][Lane] transfer lane rule row", zap.Error(err))
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, store.Error(err)
	}
	for i := range deltaRules {
		item := deltaRules[i]
		group, ok := deltaGroups[item.LaneGroup]
		if !ok {
			continue
		}
		group.LaneRules[item.ID] = item
	}
	return deltaGroups, nil
}

// ActiveLaneGroup implements store.LaneStore.
func (l *laneStore) ActiveLaneGroup(tx store.Tx, release *ruletypes.LaneGroupRelease) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := l.inactiveLaneGroupRelease(dbTx, release)
	if err != nil {
		return store.Error(err)
	}
	args := []any{maxVersion + 1, release.RuleName, release.ReleaseName}
	_, err = dbTx.Exec(`UPDATE lane_group_release
SET active = 1, version = ?, mtime = sysdate()
WHERE rule_name = ?
	AND name = ? AND active = 0`, args...)
	return store.Error(err)
}

// GetLaneGroupVersions .
func (l *laneStore) GetLaneGroupVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	countSql := `SELECT COUNT(*) FROM lane_group_release WHERE rule_name = ? AND flag = 0`
	row := l.slave.QueryRow(countSql, filter["rule_name"])
	var count uint64
	if err := row.Scan(&count); err != nil {
		log.Errorf("[store][mysql][lane] query lane_group rule versions count err: %s", err.Error())
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}

	querySql := `SELECT id, name, rule_id, rule_name, flag, active, version, description, release_type, ctime, mtime
	FROM lane_group_release
	WHERE rule_id = ?
		AND flag = 0 ORDER BY version DESC LIMIT ?, ?`
	rows, err := l.slave.Query(querySql, filter["rule_name"], offset, limit)
	if err != nil {
		log.Errorf("[store][mysql][lane] query lane_group rule versions err: %s", err.Error())
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
			log.Errorf("[store][mysql][lane] fetch lane_group rule versions scan err: %s", err.Error())
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

func (l *laneStore) GetReleaseLaneGroupRule(tx store.Tx, release *rules.RuleRelease) (*rules.LaneGroupRelease, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx, ok := tx.GetDelegateTx().(*BaseTx)
	if !ok || dbTx == nil {
		return nil, errors.New("invalid tx delegate")
	}
	querySql := `SELECT id, name, rule_name, rule, version
	, active, description, release_type
FROM lane_group_release
WHERE rule_name = ?
	AND name = ?
	AND release_type = ?
	AND flag = 0
LIMIT 1`
	row := dbTx.QueryRow(querySql, release.RuleName, release.ReleaseName, release.ReleaseType)
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
	ruleObj := &ruletypes.LaneGroup{}
	if err := json.Unmarshal([]byte(ruleStr), ruleObj); err != nil {
		return nil, store.Error(err)
	}
	proto, _ := ruleObj.ToProto()
	return &ruletypes.LaneGroupRelease{
		RuleRelease: ruletypes.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       true,
		},
		Rule: proto,
	}, nil
}

// GetActiveLaneGroup implements store.LaneStore.
func (l *laneStore) GetActiveLaneGroup(tx store.Tx, release *ruletypes.LaneGroupRelease) (*ruletypes.LaneGroupRelease, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, version
		, active, description, release_type
FROM lane_group_release
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
	ruleObj := &ruletypes.LaneGroup{}
	if err := json.Unmarshal([]byte(ruleStr), ruleObj); err != nil {
		return nil, store.Error(err)
	}
	proto, _ := ruleObj.ToProto()
	return &ruletypes.LaneGroupRelease{
		RuleRelease: ruletypes.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       true,
		},
		Rule: proto,
	}, nil
}

// InactiveLaneGroup implements store.LaneStore.
func (l *laneStore) InactiveLaneGroup(tx store.Tx, release *ruletypes.LaneGroupRelease) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	execSql := `UPDATE lane_group_release
SET active = 0, mtime = sysdate()
WHERE name = ?
	AND rule_name = ?
	AND release_type = ?
	AND active = 1
	`
	_, err := dbTx.Exec(execSql, release.ReleaseName, release.Rule.Name, release.ReleaseType)
	return store.Error(err)
}

func (l *laneStore) inactiveLaneGroupRelease(tx *BaseTx, release *rules.LaneGroupRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.RuleName, release.ReleaseType}
	//	先取消所有 active == true 的记录
	if _, err := tx.Exec("UPDATE lane_group_release SET active = 0, mtime = sysdate() "+
		" WHERE rule_name = ? AND active = 1 AND release_type = ?", args...); err != nil {
		return 0, err
	}
	return l.selectMaxVersion(tx, release)
}

func (l *laneStore) selectMaxVersion(tx *BaseTx, release *rules.LaneGroupRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.Rule.Name}
	var maxVersion uint64
	//	查询当前 release 的最大版本号
	if err := tx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM lane_group_release WHERE rule_name = ?",
		args...).Scan(&maxVersion); err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// PublishLaneGroup implements store.LaneStore.
func (l *laneStore) PublishLaneGroup(tx store.Tx, rule *ruletypes.LaneGroupRelease) error {
	if rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][lane] publish lane group missing some params")
	}
	if tx == nil {
		return errors.New("tx is nil")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := l.inactiveLaneGroupRelease(dbTx, rule)
	if err != nil {
		return store.Error(err)
	}

	ruleJson, err := json.Marshal(rule.Rule)
	if err != nil {
		return store.Error(err)
	}
	// 3. 插入新发布并激活
	insertSql := `INSERT INTO lane_group_release (id, name, rule_name, rule, version
	, active, description, release_type, ctime, mtime)
VALUES (?, ?, ?, ?, ?
	, 1, ?, ?, sysdate(), sysdate())`
	_, err = dbTx.Exec(insertSql,
		rule.Id,
		rule.ReleaseName,
		"", // rule_name 暂未使用
		string(ruleJson),
		maxVersion+1,
		rule.Description,
		rule.ReleaseType,
	)
	return store.Error(err)
}

// GetMoreLaneGroupReleases implements store.LaneStore.
func (l *laneStore) GetMoreLaneGroupReleases(firstUpdate bool, mtime time.Time) ([]*ruletypes.LaneGroupRelease, error) {
	str := `SELECT id, name, rule_name, rule, version
	, active, description, release_type, mtime
FROM lane_group_release
WHERE mtime >= FROM_UNIXTIME(?)`
	if firstUpdate {
		str += " AND active = 1"
	}
	rows, err := l.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	var out []*ruletypes.LaneGroupRelease
	for rows.Next() {
		var (
			id, name, ruleName, ruleStr, description, releaseType string
			version                                               uint64
			active                                                int
			mtime                                                 time.Time
		)
		err := rows.Scan(&id, &name, &ruleName, &ruleStr, &version, &active, &description, &releaseType, &mtime)
		if err != nil {
			return nil, store.Error(err)
		}
		ruleObj := &ruletypes.LaneGroup{}
		if err := json.Unmarshal([]byte(ruleStr), ruleObj); err != nil {
			return nil, store.Error(err)
		}
		proto, _ := ruleObj.ToProto()
		release := &ruletypes.LaneGroupRelease{
			RuleRelease: ruletypes.RuleRelease{
				Id:          id,
				ReleaseName: name,
				Description: description,
				ReleaseType: rules.ReleaseType(releaseType),
				Active:      active == 1,
				Version:     version,
				Valid:       true,
			},
			Rule: proto,
		}
		out = append(out, release)
	}
	if err := rows.Err(); err != nil {
		return nil, store.Error(err)
	}
	return out, nil
}

// GetLaneRule 查询泳道规则
func (l *laneStore) GetLaneRule(id string) (*ruletypes.LaneRule, error) {
	querySql := `SELECT id, name, group_name, rule, revision
	, priority, description, enable, flag
	, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(etime)
	, UNIX_TIMESTAMP(mtime)
FROM lane_rule
WHERE flag = 0
	AND id = ?`
	var item ruletypes.LaneRule
	err := l.master.processWithTransaction("GetLaneRule", func(tx *BaseTx) error {
		row := tx.QueryRow(querySql, id)
		var ctime, etime, mtime int64
		var flag int
		if err := row.Scan(&item.ID, &item.Name, &item.LaneGroup, &item.Rule, &item.Revision,
			&item.Priority, &item.Description, &item.Enable, &flag, &ctime, &etime, &mtime); err != nil {
			log.Error("[Store][Lane] select one lane rule", zap.String("querySql", querySql), zap.Error(err))
			return err
		}
		item.Valid = flag == 0
		item.CreateTime = time.Unix(ctime, 0)
		item.EnableTime = time.Unix(etime, 0)
		item.ModifyTime = time.Unix(mtime, 0)
		return tx.Commit()
	})
	if err != nil {
		return nil, store.Error(err)
	}
	if item.ID == "" {
		return nil, nil
	}
	return &item, nil
}

// AddLaneRules 添加泳道规则
func (l *laneStore) AddLaneRules(tx store.Tx, rules []*ruletypes.LaneRule) error {
	if len(rules) == 0 {
		return nil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)

	for i := range rules {
		item := rules[i]
		var args []interface{}

		var upsertSql string
		args = []interface{}{
			item.ID,
			item.Name,
			item.LaneGroup,
			item.Rule,
			item.Revision,
			item.Priority,
			item.Description,
			item.Enable,
		}
		addSql := `
INSERT INTO lane_rule (id, name, group_name, rule, revision, priority, description, enable, flag
	, ctime, etime, mtime)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0
	, sysdate(), %s, sysdate())
`
		etimeStr := "sysdate()"
		if !item.Enable {
			etimeStr = emptyEnableTime
		}
		upsertSql = fmt.Sprintf(addSql, etimeStr)
		if _, err := dbTx.Exec(upsertSql, args...); err != nil {
			log.Error("[Store][Lane] add lane rule", zap.String("id", item.ID), zap.String("sql", upsertSql),
				zap.String("group", item.LaneGroup), zap.String("name", item.Name), zap.Error(err))
			return store.Error(err)
		}
	}
	return nil
}

// UpdateLaneRules 更新泳道规则
func (l *laneStore) UpdateLaneRules(tx store.Tx, rules []*ruletypes.LaneRule) error {
	if len(rules) == 0 {
		return nil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)

	for i := range rules {
		item := rules[i]
		var args []any

		var upsertSql string
		args = []any{
			item.Rule,
			item.Revision,
			item.Priority,
			item.Description,
			item.Enable,
			item.ID,
		}
		if item.IsChangeEnable() {
			addSql := `
UPDATE lane_rule SET rule = ?, revision = ?, priority = ?, description = ?, enable = ?
	, etime = %s, mtime = sysdate() WHERE id = ?
`
			etimeStr := "sysdate()"
			if !item.Enable {
				etimeStr = emptyEnableTime
			}
			upsertSql = fmt.Sprintf(addSql, etimeStr)
		} else {
			upsertSql = `
UPDATE lane_rule SET rule = ?, revision = ?, priority = ?, description = ?, enable = ?
	, mtime = sysdate() WHERE id = ?
`
		}
		if _, err := dbTx.Exec(upsertSql, args...); err != nil {
			log.Error("[Store][Lane] add lane rule", zap.String("id", item.ID), zap.String("sql", upsertSql),
				zap.String("group", item.LaneGroup), zap.String("name", item.Name), zap.Error(err))
			return store.Error(err)
		}
	}
	return nil
}

// DeleteLaneRules 删除泳道规则
func (l *laneStore) DeleteLaneRules(tx store.Tx, group string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)

	// 如果 items.size > 0，只清理不再 items 里面的泳道规则
	args := make([]any, 0, len(ids))
	args = append(args, group)
	for i := range ids {
		args = append(args, ids[i])
	}

	cleanSql := fmt.Sprintf("UPDATE lane_rule SET flag = 1 WHERE group_name = ? AND name NOT IN (%s)", placeholders(len(ids)))
	if _, err := dbTx.Exec(cleanSql, args...); err != nil {
		log.Error("[Store][Lane] clean invalid lane rule", zap.String("sql", cleanSql), zap.Any("args", args), zap.Error(err))
		return store.Error(err)
	}
	return nil
}

// cleanSoftDeletedRules .
func (l *laneStore) cleanSoftDeletedRules() error {
	err := l.master.processWithTransaction("cleanSoftDeletedRules", func(tx *BaseTx) error {
		if _, err := tx.Exec("DELETE FROM lane_rule WHERE flag = 1"); err != nil {
			log.Error("[Store][Lane] clean soft delete lane_rule", zap.Error(err))
		}
		return tx.Commit()
	})
	return store.Error(err)
}

func transferLaneGroups(rows *sql.Rows, op func(group *ruletypes.LaneGroup)) error {
	if rows == nil {
		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		item := &ruletypes.LaneGroup{}
		var ctime, mtime int64
		var flag int

		if err := rows.Scan(&item.ID, &item.Name, &item.Rule, &item.Description, &item.Revision, &flag, &ctime, &mtime); err != nil {
			return err
		}
		item.Valid = flag == 0
		item.CreateTime = time.Unix(ctime, 0)
		item.ModifyTime = time.Unix(mtime, 0)
		op(item)
	}
	return nil
}

func transferLaneRules(rows *sql.Rows, op func(rule *ruletypes.LaneRule)) error {
	if rows == nil {
		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		item := &ruletypes.LaneRule{}
		var ctime, etime, mtime int64
		var flag, enable int

		if err := rows.Scan(&item.ID, &item.Name, &item.LaneGroup, &item.Rule, &item.Revision, &item.Priority, &item.Description,
			&enable, &flag, &ctime, &etime, &mtime); err != nil {
			return err
		}
		item.Valid = flag == 0
		item.Enable = enable == 1
		item.CreateTime = time.Unix(ctime, 0)
		item.EnableTime = time.Unix(etime, 0)
		item.ModifyTime = time.Unix(mtime, 0)
		op(item)
	}

	return nil
}
