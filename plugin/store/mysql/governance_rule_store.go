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
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

type governanceRuleType string

const (
	governanceRuleTypeRoute           governanceRuleType = "route"
	governanceRuleTypeRateLimit       governanceRuleType = "ratelimit"
	governanceRuleTypeCircuitBreaker  governanceRuleType = "circuitbreaker"
	governanceRuleTypeFaultDetect     governanceRuleType = "faultdetect"
	governanceRuleTypeLossless        governanceRuleType = "lossless"
	governanceRuleTypeLaneGroup       governanceRuleType = "lane-group"
	governanceRuleTypeTrafficSecurity governanceRuleType = "traffic-security"
	governanceRuleTypeTrafficMirror   governanceRuleType = "traffic-mirror"
	governanceRuleTypeTrafficMock     governanceRuleType = "traffic-mock"
)

const insertGovernanceRuleSQL = `INSERT INTO governance_rule (
	id, rule_type, namespace, name, service_id, service, method, priority,
	enable, disable, level, src_service, src_namespace, dst_service, dst_namespace,
	dst_method, labels, policy, config, rule, revision, description, metadata,
	flag, ctime, etime, mtime
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, sysdate(), sysdate(), sysdate())`

const updateGovernanceRuleSQL = `UPDATE governance_rule SET
	namespace = ?, name = ?, service_id = ?, service = ?, method = ?, priority = ?,
	enable = ?, disable = ?, level = ?, src_service = ?, src_namespace = ?,
	dst_service = ?, dst_namespace = ?, dst_method = ?, labels = ?, policy = ?,
	config = ?, rule = ?, revision = ?, description = ?, metadata = ?, mtime = sysdate()
WHERE rule_type = ? AND id = ?`

const deleteGovernanceRuleSQL = `UPDATE governance_rule SET flag = 1, mtime = sysdate() WHERE rule_type = ? AND id = ?`

const governanceRuleSelectColumns = `id, rule_type, namespace, name, service_id, service, method, priority,
	enable, disable, level, src_service, src_namespace, dst_service, dst_namespace,
	dst_method, labels, policy, config, rule, revision, description, metadata,
	flag, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(etime), UNIX_TIMESTAMP(mtime)`

const selectGovernanceRuleByIDSQL = `SELECT ` + governanceRuleSelectColumns + `
FROM governance_rule
WHERE rule_type = ? AND id = ? AND flag = 0`

const selectGovernanceRuleByNameSQL = `SELECT ` + governanceRuleSelectColumns + `
FROM governance_rule
WHERE rule_type = ? AND name = ? AND flag = 0`

const lockGovernanceRuleSQL = `SELECT ` + governanceRuleSelectColumns + `
FROM governance_rule
WHERE rule_type = ? AND (name = ? OR id = ?) AND flag = 0
FOR UPDATE`

const selectMoreGovernanceRulesSQL = `SELECT ` + governanceRuleSelectColumns + `
FROM governance_rule
WHERE mtime > FROM_UNIXTIME(?)`

const selectMoreGovernanceRulesByTypeSQL = `SELECT ` + governanceRuleSelectColumns + `
FROM governance_rule
WHERE rule_type = ? AND mtime > FROM_UNIXTIME(?)`

const inactiveGovernanceRuleReleaseSQL = `UPDATE governance_rule_release
SET active = 0, mtime = sysdate()
WHERE rule_type = ? AND rule_id = ? AND active = 1 AND release_type = ?`

const selectMaxGovernanceRuleReleaseVersionSQL = `SELECT IFNULL(MAX(version), 0)
FROM governance_rule_release
WHERE rule_type = ? AND rule_id = ?`

const activeGovernanceRuleReleaseSQL = `UPDATE governance_rule_release
SET active = 1, version = ?, mtime = sysdate()
WHERE rule_type = ? AND rule_id = ? AND name = ? AND release_type = ? AND active = 0`

const insertGovernanceRuleReleaseSQL = `INSERT INTO governance_rule_release (
	id, rule_type, name, rule_id, rule_name, namespace, service, rule, version,
	active, description, release_type, client_labels, metadata, flag, ctime, mtime
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, 0, sysdate(), sysdate())`

const selectGovernanceRuleReleaseSQL = `SELECT id, rule_type, name, rule_id, rule_name, namespace, service, rule,
	version, active, description, release_type, client_labels, metadata, flag,
	UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
FROM governance_rule_release
WHERE rule_type = ? AND (id = ? OR (rule_id = ? AND name = ? AND release_type = ?)) AND flag = 0
LIMIT 1`

const selectActiveGovernanceRuleReleaseSQL = `SELECT id, rule_type, name, rule_id, rule_name, namespace, service, rule,
	version, active, description, release_type, client_labels, metadata, flag,
	UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
FROM governance_rule_release
WHERE %s AND active = 1 AND flag = 0
ORDER BY version DESC
LIMIT 1`

const selectMoreGovernanceRuleReleasesByTypeSQL = `SELECT id, rule_type, name, rule_id, rule_name, namespace, service, rule,
	version, active, description, release_type, client_labels, metadata, flag,
	UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
FROM governance_rule_release
WHERE rule_type = ? AND mtime > FROM_UNIXTIME(?)`

const selectMoreGovernanceRuleReleasesSQL = `SELECT id, rule_type, name, rule_id, rule_name, namespace, service, rule,
	version, active, description, release_type, client_labels, metadata, flag,
	UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
FROM governance_rule_release
WHERE mtime > FROM_UNIXTIME(?)`

const inactiveGovernanceRuleReleaseByNameSQL = `UPDATE governance_rule_release
SET active = 0, mtime = sysdate()
WHERE rule_type = ? AND name = ? AND rule_name = ? AND active = 1 AND release_type = ?`

const deleteGovernanceRuleReleaseSQL = `UPDATE governance_rule_release SET flag = 1, mtime = sysdate() WHERE rule_type = ? AND id = ?`

type governanceRuleRecord struct {
	ID           string
	RuleType     governanceRuleType
	Namespace    string
	Name         string
	ServiceID    string
	Service      string
	Method       string
	Priority     int
	Enable       int
	Disable      int
	Level        string
	SrcService   string
	SrcNamespace string
	DstService   string
	DstNamespace string
	DstMethod    string
	Labels       string
	Policy       string
	Config       string
	Rule         string
	Revision     string
	Description  string
	Metadata     string
	Valid        bool
	CreateTime   time.Time
	EnableTime   time.Time
	ModifyTime   time.Time
}

type governanceRuleReleaseRecord struct {
	ID           string
	RuleType     governanceRuleType
	ReleaseName  string
	RuleID       string
	RuleName     string
	Namespace    string
	Service      string
	Rule         string
	Version      uint64
	Active       bool
	Description  string
	ReleaseType  string
	ClientLabels string
	Metadata     string
	Valid        bool
	CreateTime   time.Time
	ModifyTime   time.Time
}

type governanceRuleRepository struct {
	master *BaseDB
	slave  *BaseDB
}

func newGovernanceRuleRepository(master, slave *BaseDB) *governanceRuleRepository {
	return &governanceRuleRepository{
		master: master,
		slave:  slave,
	}
}

func (r *governanceRuleRepository) CreateRule(tx store.Tx, record *governanceRuleRecord) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	_, err = dbTx.Exec(insertGovernanceRuleSQL, record.insertArgs()...)
	return store.Error(err)
}

func (r *governanceRuleRepository) UpdateRule(tx store.Tx, record *governanceRuleRecord) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	_, err = dbTx.Exec(updateGovernanceRuleSQL, record.updateArgs()...)
	return store.Error(err)
}

func (r *governanceRuleRepository) DeleteRule(tx store.Tx, ruleType governanceRuleType, id string) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	_, err = dbTx.Exec(deleteGovernanceRuleSQL, string(ruleType), id)
	return store.Error(err)
}

func (r *governanceRuleRepository) GetRuleByID(ruleType governanceRuleType, id string) (*governanceRuleRecord, error) {
	return r.getRule(r.master.QueryRow(selectGovernanceRuleByIDSQL, string(ruleType), id))
}

func (r *governanceRuleRepository) GetRuleByName(ruleType governanceRuleType, name string) (*governanceRuleRecord, error) {
	return r.getRule(r.master.QueryRow(selectGovernanceRuleByNameSQL, string(ruleType), name))
}

func (r *governanceRuleRepository) LockRule(tx store.Tx, ruleType governanceRuleType, keyword string) (*governanceRuleRecord, error) {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return nil, err
	}
	return r.getRule(dbTx.QueryRow(lockGovernanceRuleSQL, string(ruleType), keyword, keyword))
}

func (r *governanceRuleRepository) QueryRules(
	ctx context.Context, ruleType governanceRuleType, filter map[string]string, offset, limit uint32,
) (uint32, []*governanceRuleRecord, error) {
	countSQL := "SELECT COUNT(*) FROM governance_rule WHERE rule_type = ? AND flag = 0"
	querySQL := "SELECT " + governanceRuleSelectColumns + " FROM governance_rule WHERE rule_type = ? AND flag = 0"
	args := []interface{}{string(ruleType)}
	for k, v := range filter {
		if v == "" {
			continue
		}
		switch k {
		case "name":
			if pv, ok := matchs.ParseWildName(v); ok {
				countSQL += " AND name = ?"
				querySQL += " AND name = ?"
				args = append(args, pv)
			} else {
				countSQL += " AND name LIKE ?"
				querySQL += " AND name LIKE ?"
				args = append(args, "%"+v+"%")
			}
		case "id":
			countSQL += " AND id = ?"
			querySQL += " AND id = ?"
			args = append(args, v)
		case exactName:
			countSQL += " AND name = ?"
			querySQL += " AND name = ?"
			args = append(args, v)
		case excludeId:
			countSQL += " AND id != ?"
			querySQL += " AND id != ?"
			args = append(args, v)
		case "namespace", "description", "srcService", "srcNamespace", "dstService", "dstNamespace", "dstMethod":
			field := governanceRuleFilterField(k)
			if field == "" {
				continue
			}
			if governanceRuleFilterUsesLike(k) {
				countSQL += fmt.Sprintf(" AND %s LIKE ?", field)
				querySQL += fmt.Sprintf(" AND %s LIKE ?", field)
				args = append(args, "%"+v+"%")
			} else {
				countSQL += fmt.Sprintf(" AND %s = ?", field)
				querySQL += fmt.Sprintf(" AND %s = ?", field)
				args = append(args, v)
			}
		case "enable":
			enable := 0
			if strings.EqualFold(v, "true") || v == "1" {
				enable = 1
			}
			countSQL += " AND enable = ?"
			querySQL += " AND enable = ?"
			args = append(args, enable)
		case "level":
			tokens := strings.Split(v, ",")
			countSQL += fmt.Sprintf(" AND level in (%s)", placeholders(len(tokens)))
			querySQL += fmt.Sprintf(" AND level in (%s)", placeholders(len(tokens)))
			for _, token := range tokens {
				args = append(args, token)
			}
		case svcSpecificQueryKeyService:
			countSQL += " AND (dst_service = ? OR dst_service = '*' OR src_service = ? OR src_service = '*')"
			querySQL += " AND (dst_service = ? OR dst_service = '*' OR src_service = ? OR src_service = '*')"
			args = append(args, v, v)
		case svcSpecificQueryKeyNamespace:
			countSQL += " AND (dst_namespace = ? OR dst_namespace = '*' OR src_namespace = ? OR src_namespace = '*')"
			querySQL += " AND (dst_namespace = ? OR dst_namespace = '*' OR src_namespace = ? OR src_namespace = '*')"
			args = append(args, v, v)
		}
	}
	var count uint32
	if err := r.slave.QueryRow(countSQL, args...).Scan(&count); err != nil {
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}
	orderField := sanitizeGovernanceRuleOrderField(filter["order_field"])
	orderType := sanitizeOrderType(filter["order_type"])
	querySQL += fmt.Sprintf(" ORDER BY %s %s LIMIT ?, ?", orderField, orderType)
	queryArgs := append(append([]interface{}{}, args...), offset, limit)
	rows, err := r.slave.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return 0, nil, store.Error(err)
	}
	records, err := scanGovernanceRuleRecords(rows)
	return count, records, err
}

func (r *governanceRuleRepository) GetMoreRules(mtime time.Time, firstUpdate bool) ([]*governanceRuleRecord, error) {
	querySQL := selectMoreGovernanceRulesSQL
	if firstUpdate {
		querySQL += " AND flag = 0"
	}
	rows, err := r.slave.Query(querySQL, timeToTimestamp(mtime))
	if err != nil {
		return nil, store.Error(err)
	}
	return scanGovernanceRuleRecords(rows)
}

func (r *governanceRuleRepository) GetMoreRulesByType(
	ruleType governanceRuleType, mtime time.Time, firstUpdate bool,
) ([]*governanceRuleRecord, error) {
	querySQL := selectMoreGovernanceRulesByTypeSQL
	if firstUpdate {
		querySQL += " AND flag = 0"
	}
	rows, err := r.slave.Query(querySQL, string(ruleType), timeToTimestamp(mtime))
	if err != nil {
		return nil, store.Error(err)
	}
	return scanGovernanceRuleRecords(rows)
}

func (r *governanceRuleRepository) ActiveRelease(tx store.Tx, release *governanceRuleReleaseRecord) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	if _, err := dbTx.Exec(inactiveGovernanceRuleReleaseSQL, string(release.RuleType), release.RuleID, release.ReleaseType); err != nil {
		return store.Error(err)
	}
	var maxVersion uint64
	if err := dbTx.QueryRow(selectMaxGovernanceRuleReleaseVersionSQL, string(release.RuleType), release.RuleID).Scan(&maxVersion); err != nil {
		return store.Error(err)
	}
	_, err = dbTx.Exec(activeGovernanceRuleReleaseSQL, maxVersion+1, string(release.RuleType), release.RuleID, release.ReleaseName, release.ReleaseType)
	return store.Error(err)
}

func (r *governanceRuleRepository) PublishRelease(tx store.Tx, release *governanceRuleReleaseRecord) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	if _, err := dbTx.Exec(inactiveGovernanceRuleReleaseSQL, string(release.RuleType), release.RuleID, release.ReleaseType); err != nil {
		return store.Error(err)
	}
	var maxVersion uint64
	if err := dbTx.QueryRow(selectMaxGovernanceRuleReleaseVersionSQL, string(release.RuleType), release.RuleID).Scan(&maxVersion); err != nil {
		return store.Error(err)
	}
	release.Version = maxVersion + 1
	_, err = dbTx.Exec(insertGovernanceRuleReleaseSQL, release.insertArgs()...)
	return store.Error(err)
}

func (r *governanceRuleRepository) InactiveRelease(tx store.Tx, release *governanceRuleReleaseRecord) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	_, err = dbTx.Exec(
		inactiveGovernanceRuleReleaseByNameSQL,
		string(release.RuleType),
		release.ReleaseName,
		release.RuleName,
		release.ReleaseType,
	)
	return store.Error(err)
}

func (r *governanceRuleRepository) GetRelease(
	tx store.Tx, ruleType governanceRuleType, release *rules.RuleRelease,
) (*governanceRuleReleaseRecord, error) {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return nil, err
	}
	return scanGovernanceRuleReleaseRecord(dbTx.QueryRow(
		selectGovernanceRuleReleaseSQL,
		string(ruleType),
		release.Id,
		release.RuleId,
		release.ReleaseName,
		release.ReleaseType,
	))
}

func (r *governanceRuleRepository) GetActiveRelease(
	tx store.Tx, release *governanceRuleReleaseRecord,
) (*governanceRuleReleaseRecord, error) {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return nil, err
	}
	where, args := buildGovernanceRuleReleaseActiveWhere(release)
	row := dbTx.QueryRow(fmt.Sprintf(selectActiveGovernanceRuleReleaseSQL, where), args...)
	return scanGovernanceRuleReleaseRecord(row)
}

func (r *governanceRuleRepository) GetMoreReleasesByType(
	ruleType governanceRuleType, firstUpdate bool, mtime time.Time,
) ([]*governanceRuleReleaseRecord, error) {
	querySQL := selectMoreGovernanceRuleReleasesByTypeSQL
	if firstUpdate {
		querySQL += " AND active = 1"
	}
	rows, err := r.slave.Query(querySQL, string(ruleType), timeToTimestamp(mtime))
	if err != nil {
		return nil, store.Error(err)
	}
	return scanGovernanceRuleReleaseRecords(rows)
}

func (r *governanceRuleRepository) GetMoreReleases(
	mtime time.Time, firstUpdate bool,
) ([]*governanceRuleReleaseRecord, error) {
	querySQL := selectMoreGovernanceRuleReleasesSQL
	if firstUpdate {
		querySQL += " AND active = 1"
	}
	rows, err := r.slave.Query(querySQL, timeToTimestamp(mtime))
	if err != nil {
		return nil, store.Error(err)
	}
	return scanGovernanceRuleReleaseRecords(rows)
}

func (r *governanceRuleRepository) DeleteRelease(tx store.Tx, ruleType governanceRuleType, id string) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	_, err = dbTx.Exec(deleteGovernanceRuleReleaseSQL, string(ruleType), id)
	return store.Error(err)
}

func (r *governanceRuleRepository) QueryReleaseVersions(
	ctx context.Context, ruleType governanceRuleType, resource apimodel.RuleRelease_RuleType,
	filter map[string]string, offset, limit uint32,
) (uint64, []*rules.RuleRelease, error) {
	countWhere := "rule_type = ?"
	args := []interface{}{string(ruleType)}
	if ruleID := filter["rule_id"]; ruleID != "" {
		countWhere += " AND rule_id = ?"
		args = append(args, ruleID)
	} else if ruleName := filter["rule_name"]; ruleName != "" {
		countWhere += " AND rule_name = ?"
		args = append(args, ruleName)
	}
	countWhere += " AND flag = 0"
	queryWhere := countWhere
	return QueryRuleVersions(
		r.slave,
		"governance_rule_release",
		countWhere,
		queryWhere,
		args,
		append([]interface{}{}, args...),
		offset,
		limit,
		func(release *rules.RuleRelease) {
			release.Resource = resource
		},
	)
}

func (r *governanceRuleRepository) getRule(row governanceRuleScanner) (*governanceRuleRecord, error) {
	record, err := scanGovernanceRuleRecord(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	return record, nil
}

func (r *governanceRuleRecord) insertArgs() []interface{} {
	return []interface{}{
		r.ID,
		string(r.RuleType),
		r.Namespace,
		r.Name,
		r.ServiceID,
		r.Service,
		r.Method,
		r.Priority,
		r.Enable,
		r.Disable,
		r.Level,
		r.SrcService,
		r.SrcNamespace,
		r.DstService,
		r.DstNamespace,
		r.DstMethod,
		r.Labels,
		r.Policy,
		r.Config,
		r.Rule,
		r.Revision,
		r.Description,
		r.Metadata,
	}
}

func (r *governanceRuleRecord) updateArgs() []interface{} {
	return []interface{}{
		r.Namespace,
		r.Name,
		r.ServiceID,
		r.Service,
		r.Method,
		r.Priority,
		r.Enable,
		r.Disable,
		r.Level,
		r.SrcService,
		r.SrcNamespace,
		r.DstService,
		r.DstNamespace,
		r.DstMethod,
		r.Labels,
		r.Policy,
		r.Config,
		r.Rule,
		r.Revision,
		r.Description,
		r.Metadata,
		string(r.RuleType),
		r.ID,
	}
}

func (r *governanceRuleReleaseRecord) insertArgs() []interface{} {
	return []interface{}{
		r.ID,
		string(r.RuleType),
		r.ReleaseName,
		r.RuleID,
		r.RuleName,
		r.Namespace,
		r.Service,
		r.Rule,
		r.Version,
		r.Description,
		r.ReleaseType,
		r.ClientLabels,
		r.Metadata,
	}
}

func scanGovernanceRuleRecords(rows *sql.Rows) ([]*governanceRuleRecord, error) {
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()

	out := make([]*governanceRuleRecord, 0)
	for rows.Next() {
		item, err := scanGovernanceRuleRecord(rows)
		if err != nil {
			return nil, store.Error(err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, store.Error(err)
	}
	return out, nil
}

type governanceRuleScanner interface {
	Scan(dest ...interface{}) error
}

func scanGovernanceRuleRecord(row governanceRuleScanner) (*governanceRuleRecord, error) {
	if row == nil {
		return nil, errors.New("nil governance rule row")
	}
	item := &governanceRuleRecord{}
	var ruleType string
	var flag int
	var ctime, etime, mtime int64
	if err := row.Scan(
		&item.ID,
		&ruleType,
		&item.Namespace,
		&item.Name,
		&item.ServiceID,
		&item.Service,
		&item.Method,
		&item.Priority,
		&item.Enable,
		&item.Disable,
		&item.Level,
		&item.SrcService,
		&item.SrcNamespace,
		&item.DstService,
		&item.DstNamespace,
		&item.DstMethod,
		&item.Labels,
		&item.Policy,
		&item.Config,
		&item.Rule,
		&item.Revision,
		&item.Description,
		&item.Metadata,
		&flag,
		&ctime,
		&etime,
		&mtime,
	); err != nil {
		return nil, err
	}
	item.RuleType = governanceRuleType(ruleType)
	item.Valid = flag == 0
	item.CreateTime = time.Unix(ctime, 0)
	item.EnableTime = time.Unix(etime, 0)
	item.ModifyTime = time.Unix(mtime, 0)
	return item, nil
}

type governanceRuleReleaseScanner interface {
	Scan(dest ...interface{}) error
}

func scanGovernanceRuleReleaseRecords(rows *sql.Rows) ([]*governanceRuleReleaseRecord, error) {
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()

	out := make([]*governanceRuleReleaseRecord, 0)
	for rows.Next() {
		item, err := scanGovernanceRuleReleaseRecord(rows)
		if err != nil {
			return nil, store.Error(err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, store.Error(err)
	}
	return out, nil
}

func scanGovernanceRuleReleaseRecord(row governanceRuleReleaseScanner) (*governanceRuleReleaseRecord, error) {
	if row == nil {
		return nil, errors.New("nil governance rule release row")
	}
	item := &governanceRuleReleaseRecord{}
	var ruleType string
	var active, flag int
	var ctime, mtime int64
	if err := row.Scan(
		&item.ID,
		&ruleType,
		&item.ReleaseName,
		&item.RuleID,
		&item.RuleName,
		&item.Namespace,
		&item.Service,
		&item.Rule,
		&item.Version,
		&active,
		&item.Description,
		&item.ReleaseType,
		&item.ClientLabels,
		&item.Metadata,
		&flag,
		&ctime,
		&mtime,
	); err != nil {
		return nil, err
	}
	item.RuleType = governanceRuleType(ruleType)
	item.Active = active == 1
	item.Valid = flag == 0
	item.CreateTime = time.Unix(ctime, 0)
	item.ModifyTime = time.Unix(mtime, 0)
	return item, nil
}

func buildGovernanceRuleReleaseActiveWhere(release *governanceRuleReleaseRecord) (string, []interface{}) {
	whereHolder := []string{"rule_type = ?"}
	args := []interface{}{string(release.RuleType)}
	if release.RuleID != "" {
		whereHolder = append(whereHolder, "rule_id = ?")
		args = append(args, release.RuleID)
	}
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
	return strings.Join(whereHolder, " AND "), args
}

func sanitizeGovernanceRuleOrderField(field string) string {
	switch field {
	case "ctime", "mtime", "name", "priority":
		return field
	default:
		return "mtime"
	}
}

func sanitizeOrderType(orderType string) string {
	if strings.EqualFold(orderType, "asc") {
		return "ASC"
	}
	return "DESC"
}

func marshalMetadata(metadata map[string]string) string {
	if len(metadata) == 0 {
		return "{}"
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func governanceRuleFilterField(key string) string {
	switch key {
	case "namespace":
		return "namespace"
	case "description":
		return "description"
	case "srcService":
		return "src_service"
	case "srcNamespace":
		return "src_namespace"
	case "dstService":
		return "dst_service"
	case "dstNamespace":
		return "dst_namespace"
	case "dstMethod":
		return "dst_method"
	default:
		return ""
	}
}

func governanceRuleFilterUsesLike(key string) bool {
	switch key {
	case "description", "srcService", "srcNamespace", "dstService", "dstNamespace", "dstMethod":
		return true
	default:
		return false
	}
}
