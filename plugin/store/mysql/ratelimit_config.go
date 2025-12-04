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
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/specification/source/go/api/v1/model"
	"go.uber.org/zap"
)

var _ store.RateLimitStore = (*rateLimitStore)(nil)

// rateLimitStore RateLimitStore的实现
type rateLimitStore struct {
	master *BaseDB
	slave  *BaseDB
}

// CreateRateLimit 新建限流规则
func (rls *rateLimitStore) CreateRateLimit(limit *rules.RateLimit) error {
	if limit.ID == "" || limit.Revision == "" {
		return errors.New("[store][mysql][ratelimit] create rate limit missing some params")
	}
	err := RetryTransaction("createRateLimit", func() error {
		return rls.createRateLimit(limit)
	})

	return store.Error(err)
}

func limitToEtimeStr(limit *rules.RateLimit) string {
	etimeStr := "sysdate()"
	if limit.Disable {
		etimeStr = emptyEnableTime
	}
	return etimeStr
}

// createRateLimit
func (rls *rateLimitStore) createRateLimit(limit *rules.RateLimit) error {
	tx, err := rls.master.Begin()
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] create rate limit(%+v) begin tx err: %s", limit, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	etimeStr := limitToEtimeStr(limit)
	// 新建限流规则
	str := fmt.Sprintf(`insert into ratelimit_rule(
			id, name, disable, service_id, method, labels, priority, rule, revision, ctime, mtime, etime, metadata)
			values(?,?,?,?,?,?,?,?,?,sysdate(),sysdate(), %s, ?)`, etimeStr)
	if _, err := tx.Exec(str, limit.ID, limit.Name, limit.Disable, limit.ServiceID, limit.Method, limit.Labels,
		limit.Priority, limit.Rule, limit.Revision, utils.MustJson(limit.Metadata)); err != nil {
		log.Errorf("[store][mysql][ratelimit] create rate limit(%+v), sql %s err: %s", limit, str, err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[store][mysql][ratelimit] create rate limit(%+v) commit tx err: %s", limit, err.Error())
		return err
	}

	return nil
}

// UpdateRateLimit 更新限流规则
func (rls *rateLimitStore) UpdateRateLimit(limit *rules.RateLimit) error {
	if limit.ID == "" || limit.Revision == "" {
		return errors.New("[store][mysql][ratelimit] update rate limit missing some params")
	}

	err := RetryTransaction("updateRateLimit", func() error {
		return rls.updateRateLimit(limit)
	})

	return store.Error(err)
}

// EnableRateLimit 启用限流规则
func (rls *rateLimitStore) EnableRateLimit(limit *rules.RateLimit) error {
	if limit.ID == "" || limit.Revision == "" {
		return errors.New("[store][mysql][ratelimit] enable rate limit missing some params")
	}

	err := RetryTransaction("enableRateLimit", func() error {
		return rls.enableRateLimit(limit)
	})

	return store.Error(err)
}

// enableRateLimit
func (rls *rateLimitStore) enableRateLimit(limit *rules.RateLimit) error {
	tx, err := rls.master.Begin()
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v) begin tx err: %s", limit, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	etimeStr := limitToEtimeStr(limit)
	str := fmt.Sprintf(
		`update ratelimit_rule set disable = ?, revision = ?, mtime = sysdate(), etime=%s where id = ?`, etimeStr)
	if _, err := tx.Exec(str, limit.Disable, limit.Revision, limit.ID); err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v), sql %s, err: %s", limit, str, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v) commit tx err: %s", limit, err.Error())
		return err
	}
	return nil
}

// updateRateLimit
func (rls *rateLimitStore) updateRateLimit(limit *rules.RateLimit) error {
	tx, err := rls.master.Begin()
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v) begin tx err: %s", limit, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	etimeStr := limitToEtimeStr(limit)
	str := fmt.Sprintf(`update ratelimit_rule set name = ?, service_id=?, disable = ?, method= ?,
			labels = ?, priority = ?, rule = ?, revision = ?, mtime = sysdate(), etime=%s, metadata = ? where id = ?`, etimeStr)
	if _, err := tx.Exec(str, limit.Name, limit.ServiceID, limit.Disable,
		limit.Method, limit.Labels, limit.Priority, limit.Rule, limit.Revision, utils.MustJson(limit.Metadata), limit.ID); err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v), sql %s, err: %s", limit, str, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v) commit tx err: %s", limit, err.Error())
		return err
	}
	return nil
}

// DeleteRateLimit 删除限流规则
func (rls *rateLimitStore) DeleteRateLimit(limit *rules.RateLimit) error {
	if limit.ID == "" || limit.Revision == "" {
		return errors.New("[store][mysql][ratelimit] delete rate limit missing some params")
	}

	err := RetryTransaction("deleteRateLimit", func() error {
		return rls.deleteRateLimit(limit)
	})

	return store.Error(err)
}

// deleteRateLimit
func (rls *rateLimitStore) deleteRateLimit(limit *rules.RateLimit) error {
	tx, err := rls.master.Begin()
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] delete rate limit(%+v) begin tx err: %s", limit, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	str := `update ratelimit_rule set flag = 1, mtime = sysdate() where id = ?`
	if _, err := tx.Exec(str, limit.ID); err != nil {
		log.Errorf("[store][mysql][ratelimit] delete rate limit(%+v) err: %s", limit, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[store][mysql][ratelimit] delete rate limit(%+v) commit tx err: %s", limit, err.Error())
		return err
	}
	return nil
}

// GetRateLimitWithID 根据限流规则ID获取限流规则
func (rls *rateLimitStore) GetRateLimitWithID(id string) (*rules.RateLimit, error) {
	if id == "" {
		log.Errorf("[store][mysql][ratelimit] get rate limit missing some params")
		return nil, errors.New("get rate limit missing some params")
	}

	str := `select id, name, disable, service_id, method, labels, priority, rule, revision, flag,
			unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime), IFNULL(metadata, '{}')
			from ratelimit_rule where id = ? and flag = 0`
	rows, err := rls.master.Query(str, id)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limit with id(%s) err: %s", id, err.Error())
		return nil, err
	}
	out, err := fetchRateLimitRows(rows)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out[0], nil
}

// fetchRateLimitRows 读取限流数据
func fetchRateLimitRows(rows *sql.Rows) ([]*rules.RateLimit, error) {
	defer rows.Close()
	var out []*rules.RateLimit
	for rows.Next() {
		var rateLimit rules.RateLimit
		var flag int
		var ctime, mtime, etime int64
		var metadata string
		err := rows.Scan(&rateLimit.ID, &rateLimit.Name, &rateLimit.Disable, &rateLimit.ServiceID, &rateLimit.Method,
			&rateLimit.Labels, &rateLimit.Priority, &rateLimit.Rule, &rateLimit.Revision, &flag, &ctime, &mtime, &etime, &metadata)
		if err != nil {
			log.Errorf("[store][mysql][ratelimit] fetch rate limit scan err: %s", err.Error())
			return nil, err
		}
		rateLimit.CreateTime = time.Unix(ctime, 0)
		rateLimit.ModifyTime = time.Unix(mtime, 0)
		rateLimit.EnableTime = time.Unix(etime, 0)
		if metadata != "" {
			rateLimit.Metadata = make(map[string]string)
			_ = json.Unmarshal([]byte(metadata), &rateLimit.Metadata)
		}
		rateLimit.Valid = true
		if flag == 1 {
			rateLimit.Valid = false
		}
		out = append(out, &rateLimit)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[store][mysql][ratelimit] fetch rate limit next err: %s", err.Error())
		return nil, err
	}
	return out, nil
}

// GetMoreRateLimits 根据修改时间拉取增量限流规则及最新版本号
func (rls *rateLimitStore) GetMoreRateLimits(mtime time.Time,
	firstUpdate bool) ([]*rules.RateLimit, error) {
	str := `select id, name, disable, ratelimit_rule.service_id, method, labels, priority, rule, revision, flag,
			unix_timestamp(ratelimit_rule.ctime), unix_timestamp(ratelimit_rule.mtime),
			unix_timestamp(ratelimit_rule.etime), IFNULL(metadata, '{}') from ratelimit_rule
			where ratelimit_rule.mtime > FROM_UNIXTIME(?)`
	if firstUpdate {
		str += " and flag != 1"
	}
	rows, err := rls.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limits with mtime err: %s", err.Error())
		return nil, err
	}
	rateLimits, err := fetchRateLimitCacheRows(rows)
	if err != nil {
		return nil, err
	}
	return rateLimits, nil
}

// fetchRateLimitCacheRows 读取限流数据以及最新版本号
func fetchRateLimitCacheRows(rows *sql.Rows) ([]*rules.RateLimit, error) {
	defer rows.Close()

	var rateLimits []*rules.RateLimit

	for rows.Next() {
		var (
			rateLimit           rules.RateLimit
			ctime, mtime, etime int64
			serviceID           string
			flag                int
			metadata            string
		)
		err := rows.Scan(&rateLimit.ID, &rateLimit.Name, &rateLimit.Disable, &serviceID, &rateLimit.Method, &rateLimit.Labels,
			&rateLimit.Priority, &rateLimit.Rule, &rateLimit.Revision, &flag, &ctime, &mtime, &etime, &metadata)
		if err != nil {
			log.Errorf("[store][mysql][ratelimit] fetch rate limit cache scan err: %s", err.Error())
			return nil, err
		}
		rateLimit.CreateTime = time.Unix(ctime, 0)
		rateLimit.ModifyTime = time.Unix(mtime, 0)
		rateLimit.EnableTime = time.Unix(etime, 0)
		rateLimit.Valid = true
		if flag == 1 {
			rateLimit.Valid = false
		}
		if metadata != "" {
			rateLimit.Metadata = make(map[string]string)
			_ = json.Unmarshal([]byte(metadata), &rateLimit.Metadata)
		}
		rateLimit.ServiceID = serviceID

		rateLimits = append(rateLimits, &rateLimit)
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[store][mysql][ratelimit] fetch rate limit cache next err: %s", err.Error())
		return nil, err
	}
	return rateLimits, nil
}

// GetMoreRateLimitReleases 根据修改时间拉取增量限流规则及最新版本号
func (rls *rateLimitStore) GetMoreRateLimitReleases(mtime time.Time,
	firstUpdate bool) ([]*rules.RateLimitRelease, error) {
	str := `select id, name, rule_name, rule, flag, active, version, description, release_type,
			unix_timestamp(ctime), unix_timestamp(mtime) from ratelimit_rule_release
			where mtime >= FROM_UNIXTIME(?)`
	if firstUpdate {
		mtime = time.Time{}
		str += " and flag != 1"
	}
	rows, err := rls.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limits with mtime err: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	var releases []*rules.RateLimitRelease

	for rows.Next() {
		var (
			item = &rules.RateLimitRelease{
				Rule: &rules.RateLimit{},
			}
			ctime, mtime int64
			flag, active int
			ruleStr      string
		)
		err := rows.Scan(&item.Id, &item.ReleaseName, &item.RuleName, &ruleStr, &flag, &active, &item.Version, &item.Description,
			&item.ReleaseType, &ctime, &mtime)
		if err != nil {
			log.Errorf("[store][mysql][ratelimit] fetch rate limit cache scan err: %s", err.Error())
			return nil, err
		}
		_ = json.Unmarshal([]byte(ruleStr), item.Rule)

		item.Rule.CreateTime = time.Unix(ctime, 0)
		item.Rule.ModifyTime = time.Unix(mtime, 0)
		if active == 1 {
			item.Active = true
		}
		if flag == 0 {
			item.Valid = true
		}
		releases = append(releases, item)
	}

	if err := rows.Err(); err != nil {
		log.Error("[store][mysql][ratelimit] fetch rate_limit release cache next", zap.Error(err))
		return nil, err
	}
	return releases, nil
}

const (
	briefSearch = "brief"
)

// LockRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) LockRateLimitRule(tx store.Tx, keyword string) (*rules.RateLimit, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}

	str := `select id, name, disable, service_id, method, labels, priority, rule, revision, flag,
			unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime), IFNULL(metadata, '{}')
			from ratelimit_rule where (id = ? OR name = ?) and flag = 0 for update`
	rows, err := rls.master.Query(str, keyword, keyword)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limit with keyword(%s) err: %s", keyword, err.Error())
		return nil, err
	}
	out, err := fetchRateLimitRows(rows)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out[0], nil
}

// GetRateLimitRuleVersions .
func (rls *rateLimitStore) GetRateLimitRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	countSql := `SELECT COUNT(*) FROM ratelimit_rule_release WHERE rule_name = ? AND flag = 0`
	row := rls.slave.QueryRow(countSql, filter["rule_name"])
	var count uint64
	if err := row.Scan(&count); err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limit rule versions count err: %s", err.Error())
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}

	querySql := `SELECT id, name, rule_id, rule_name, flag, active, version, description, release_type, unix_timestamp(ctime), unix_timestamp(mtime)
	FROM ratelimit_rule_release
	WHERE rule_name = ?
		AND flag = 0 ORDER BY version DESC LIMIT ?, ?`
	rows, err := rls.slave.Query(querySql, filter["rule_name"], offset, limit)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limit rule versions err: %s", err.Error())
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
			log.Errorf("[store][mysql][ratelimit] fetch rate limit rule versions scan err: %s", err.Error())
			return 0, nil, store.Error(err)
		}
		item.Active = active == 1
		item.Valid = flag == 0
		item.Ctime = time.Unix(ctime, 0)
		item.Mtime = time.Unix(mtime, 0)
		item.Resource = model.RuleRelease_RateLimitRules
		releases = append(releases, item)
	}

	return count, releases, nil
}

// GetReleaseRateLimitRule
func (rls *rateLimitStore) GetReleaseRateLimitRule(tx store.Tx, release *rules.RuleRelease) (*rules.RateLimitRelease, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, flag, active, version, description, release_type
	FROM ratelimit_rule_release
	WHERE name = ?
		AND rule_id = ?
		AND release_type = ?
		AND flag = 0
	LIMIT 1`
	row := dbTx.QueryRow(querySql, release.ReleaseName, release.RuleId, release.ReleaseType)
	var (
		id, name, ruleName, ruleStr, description, releaseType string
		flag, active                                          int
		version                                               uint64
	)
	err := row.Scan(&id, &name, &ruleName, &ruleStr, &flag, &active, &version, &description, &releaseType)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ruleObj rules.RateLimit
	if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
		return nil, err
	}
	return &rules.RateLimitRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       flag == 0,
		},
		Rule: &ruleObj,
	}, nil
}

// GetActiveRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) GetActiveRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) (*rules.RateLimitRelease, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_name, rule, flag
		, active, version, description, release_type
FROM ratelimit_rule_release
WHERE %s
	AND flag = 0
	AND active = 1
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
		flag, active                                          int
		version                                               uint64
	)
	err := row.Scan(&id, &name, &ruleName, &ruleStr, &flag, &active, &version, &description, &releaseType)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var ruleObj rules.RateLimit
	if err := json.Unmarshal([]byte(ruleStr), &ruleObj); err != nil {
		return nil, err
	}

	return &rules.RateLimitRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       flag == 0,
		},
		Rule: &ruleObj,
	}, nil
}

// ActiveRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) ActiveRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	maxVersion, err := rls.inactiveRatelimitRelease(dbTx, release)
	if err != nil {
		return err
	}
	args := []any{maxVersion + 1, release.ReleaseType, release.ReleaseName, release.RuleName}
	//	update 指定的 release 记录，设置其 active、version 以及 mtime
	updateSql := `UPDATE ratelimit_rule_release
SET active = 1, version = ?, mtime = sysdate(), release_type = ?
WHERE name = ? AND rule_name = ?`
	if _, err := dbTx.Exec(updateSql, args...); err != nil {
		return store.Error(err)
	}
	return nil
}

func (rls *rateLimitStore) inactiveRatelimitRelease(tx *BaseTx, release *rules.RateLimitRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.RuleName, release.ReleaseType}
	//	先取消所有 active == true 的记录
	if _, err := tx.Exec("UPDATE ratelimit_rule_release SET active = 0, mtime = sysdate() "+
		" WHERE rule_name = ? AND active = 1 AND release_type = ?", args...); err != nil {
		return 0, err
	}
	return rls.selectMaxVersion(tx, release)
}

func (rls *rateLimitStore) selectMaxVersion(tx *BaseTx, release *rules.RateLimitRelease) (uint64, error) {
	if tx == nil {
		return 0, ErrTxIsNil
	}

	args := []any{release.Rule.Name}
	var maxVersion uint64
	//	查询当前 release 的最大版本号
	if err := tx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM ratelimit_rule_release WHERE rule_name = ?",
		args...).Scan(&maxVersion); err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// InactiveRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) InactiveRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)

	args := []any{release.Rule.Name, release.ReleaseName, release.ReleaseType}
	//	先取消所有 active == true 的记录
	if _, err := dbTx.Exec("UPDATE ratelimit_rule_release SET active = 0, mtime = sysdate() "+
		" WHERE rule_name = ? AND name = ? AND active = 1 AND release_type = ?", args...); err != nil {
		log.Errorf("[store][mysql][ratelimit] inactive rate limit rule(%+v) err: %s", release, err.Error())
		return store.Error(err)
	}
	return nil
}

// PublishRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) PublishRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) error {
	if release.Rule.Name == "" || release.ReleaseName == "" || release.ReleaseType == "" {
		return errors.New("[store][mysql][ratelimit] publish rate limit rule missing some params")
	}
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)

	// 1. 先将同名同类型的所有发布置为 inactive
	if _, err := dbTx.Exec("UPDATE ratelimit_rule_release SET active = 0, mtime = sysdate() WHERE name = ? "+
		" AND rule_name = ? AND active = 1", release.ReleaseName, release.Rule.Name); err != nil {
		log.Errorf("[store][mysql][ratelimit] publish rate limit rule(%+v) err: %s", release, err.Error())
		return store.Error(err)
	}

	// 2. 获取当前最大版本号
	maxVersion, err := rls.selectMaxVersion(dbTx, release)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] select max version for rate limit rule(%+v) err: %s", release, err.Error())
		return store.Error(err)
	}

	rule := release.Rule
	ruleJson, err := json.Marshal(rule)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] marshal rule(%+v) err: %s", release.Rule, err.Error())
		return err
	}

	// 3. 插入新发布并激活
	str := `replace into ratelimit_rule_release(
			id, name, rule_id, rule_name, rule, flag, version, active, description, release_type, ctime, mtime)
			values(?,?,?,?,?,0,?,1,?,?,sysdate(),sysdate())`
	if _, err := dbTx.Exec(str, release.Id, release.ReleaseName, rule.ID, rule.Name, ruleJson, maxVersion+1,
		release.Description, release.ReleaseType); err != nil {
		log.Error("[store][mysql][ratelimit] create rate_limit release", zap.String("rule-id", rule.ID),
			zap.String("release-name", release.ReleaseName), zap.Error(err))
		return store.Error(err)
	}
	return nil
}

func (rls *rateLimitStore) DeleteRateLimitReleases(tx store.Tx, rule *rules.RateLimitRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	deleteSql := `UPDATE ratelimit_rule_release SET flag = 1, mtime = sysdate() WHERE id = ?`
	_, err := dbTx.Exec(deleteSql, rule.Id)
	return store.Error(err)
}
