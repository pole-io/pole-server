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
	"errors"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/specification/source/go/api/v1/model"
)

var _ store.RateLimitStore = (*rateLimitStore)(nil)

// rateLimitStore RateLimitStore的实现
type rateLimitStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
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

	if err := rls.repo().CreateRule(NewSqlDBTx(tx), rateLimitToGovernanceRuleRecord(limit)); err != nil {
		log.Errorf("[store][mysql][ratelimit] create rate limit(%+v), err: %s", limit, err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[store][mysql][ratelimit] create rate limit(%+v) commit tx err: %s", limit, err.Error())
		return err
	}

	return nil
}

func (rls *rateLimitStore) repo() *governanceRuleRepository {
	if rls.governanceRuleRepository == nil {
		rls.governanceRuleRepository = newGovernanceRuleRepository(rls.master, rls.slave)
	}
	return rls.governanceRuleRepository
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

	save, err := rls.GetRateLimitWithID(limit.ID)
	if err != nil {
		return err
	}
	if save == nil {
		return nil
	}
	save.Disable = limit.Disable
	save.Revision = limit.Revision
	if err := rls.repo().UpdateRule(NewSqlDBTx(tx), rateLimitToGovernanceRuleRecord(save)); err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v), err: %s", limit, err)
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

	if err := rls.repo().UpdateRule(NewSqlDBTx(tx), rateLimitToGovernanceRuleRecord(limit)); err != nil {
		log.Errorf("[store][mysql][ratelimit] update rate limit(%+v), err: %s", limit, err)
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

	if err := rls.repo().DeleteRule(NewSqlDBTx(tx), governanceRuleTypeRateLimit, limit.ID); err != nil {
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

	record, err := rls.repo().GetRuleByID(governanceRuleTypeRateLimit, id)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limit with id(%s) err: %s", id, err.Error())
		return nil, err
	}
	return governanceRuleRecordToRateLimit(record)
}

// GetMoreRateLimits 根据修改时间拉取增量限流规则及最新版本号
func (rls *rateLimitStore) GetMoreRateLimits(mtime time.Time,
	firstUpdate bool) ([]*rules.RateLimit, error) {
	records, err := rls.repo().GetMoreRulesByType(governanceRuleTypeRateLimit, mtime, firstUpdate)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limits with mtime err: %s", err.Error())
		return nil, err
	}
	out := make([]*rules.RateLimit, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToRateLimit(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// GetMoreRateLimitReleases 根据修改时间拉取增量限流规则及最新版本号
func (rls *rateLimitStore) GetMoreRateLimitReleases(mtime time.Time,
	firstUpdate bool) ([]*rules.RateLimitRelease, error) {
	records, err := rls.repo().GetMoreReleasesByType(governanceRuleTypeRateLimit, firstUpdate, mtime)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limits with mtime err: %s", err.Error())
		return nil, err
	}
	releases := make([]*rules.RateLimitRelease, 0, len(records))
	for i := range records {
		item, err := governanceRuleReleaseRecordToRateLimitRelease(records[i])
		if err != nil {
			return nil, err
		}
		releases = append(releases, item)
	}
	return releases, nil
}

const (
	briefSearch = "brief"
)

// LockRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) LockRateLimitRule(tx store.Tx, namespace, keyword string) (*rules.RateLimit, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}

	record, err := rls.repo().LockRule(tx, governanceRuleTypeRateLimit, keyword, namespace, keyword)
	if err != nil {
		log.Errorf("[store][mysql][ratelimit] query rate limit with keyword(%s) err: %s", keyword, err.Error())
		return nil, err
	}
	return governanceRuleRecordToRateLimit(record)
}

// GetRateLimitRuleVersions .
func (rls *rateLimitStore) GetRateLimitRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return rls.repo().QueryReleaseVersions(ctx, governanceRuleTypeRateLimit, model.RuleRelease_RateLimitRules, filter, offset, limit)
}

// GetReleaseRateLimitRule
func (rls *rateLimitStore) GetReleaseRateLimitRule(tx store.Tx, release *rules.RuleRelease) (*rules.RateLimitRelease, error) {
	record, err := rls.repo().GetRelease(tx, governanceRuleTypeRateLimit, release)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToRateLimitRelease(record)
}

// GetActiveRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) GetActiveRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) (*rules.RateLimitRelease, error) {
	record, err := rls.repo().GetActiveRelease(tx, rateLimitReleaseToGovernanceReleaseRecord(release))
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToRateLimitRelease(record)
}

// ActiveRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) ActiveRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) error {
	return rls.repo().ActiveRelease(tx, rateLimitReleaseToGovernanceReleaseRecord(release))
}

// InactiveRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) InactiveRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) error {
	return rls.repo().InactiveRelease(tx, rateLimitReleaseToGovernanceReleaseRecord(release))
}

// PublishRateLimitRule implements store.RateLimitStore.
func (rls *rateLimitStore) PublishRateLimitRule(tx store.Tx, release *rules.RateLimitRelease) error {
	if release.Rule.Name == "" || release.ReleaseName == "" || release.ReleaseType == "" {
		return errors.New("[store][mysql][ratelimit] publish rate limit rule missing some params")
	}
	return rls.repo().PublishRelease(tx, rateLimitReleaseToGovernanceReleaseRecord(release))
}

func (rls *rateLimitStore) DeleteRateLimitReleases(tx store.Tx, rule *rules.RateLimitRelease) error {
	return rls.repo().DeleteRelease(tx, governanceRuleTypeRateLimit, rule.Id)
}
