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
	"strings"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/specification/source/go/api/v1/model"
)

const (
	labelCreateCircuitBreakerRule = "createCircuitBreakerRule"
	labelUpdateCircuitBreakerRule = "updateCircuitBreakerRule"
	labelDeleteCircuitBreakerRule = "deleteCircuitBreakerRule"
)

var _ store.CircuitBreakerStore = (*circuitBreakerStore)(nil)

// circuitBreakerStore 的实现
type circuitBreakerStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
}

func (c *circuitBreakerStore) CreateCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error {
	err := RetryTransaction(labelCreateCircuitBreakerRule, func() error {
		return c.createCircuitBreakerRule(cbRule)
	})

	return store.Error(err)
}

func (c *circuitBreakerStore) createCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error {
	return c.master.processWithTransaction(labelCreateCircuitBreakerRule, func(tx *BaseTx) error {
		if err := c.repo().CreateRule(NewSqlDBTx(tx), circuitBreakerRuleToGovernanceRuleRecord(cbRule)); err != nil {
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

func (c *circuitBreakerStore) repo() *governanceRuleRepository {
	if c.governanceRuleRepository == nil {
		c.governanceRuleRepository = newGovernanceRuleRepository(c.master, c.slave)
	}
	return c.governanceRuleRepository
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
		if err := c.repo().UpdateRule(NewSqlDBTx(tx), circuitBreakerRuleToGovernanceRuleRecord(cbRule)); err != nil {
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
		if err := c.repo().DeleteRule(NewSqlDBTx(tx), governanceRuleTypeCircuitBreaker, id); err != nil {
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

func (c *circuitBreakerStore) GetCircuitBreakerRules(
	filter map[string]string, offset uint32, limit uint32) (uint32, []*rules.CircuitBreakerRule, error) {
	delete(filter, briefSearch)
	num, records, err := c.repo().QueryRules(context.Background(), governanceRuleTypeCircuitBreaker, filter, offset, limit)
	if err != nil {
		return 0, nil, err
	}
	out := make([]*rules.CircuitBreakerRule, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToCircuitBreakerRule(records[i])
		if err != nil {
			return 0, nil, err
		}
		out = append(out, item)
	}
	return num, out, nil
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

func (c *circuitBreakerStore) getCircuitBreakerRulesCount(filter map[string]string) (uint32, error) {
	total, _, err := c.repo().QueryRules(context.Background(), governanceRuleTypeCircuitBreaker, filter, 0, 0)
	return total, err
}

// GetMoreCircuitBreakers list circuitbreaker rules by query
func (c *circuitBreakerStore) GetMoreCircuitBreakers(mtime time.Time, firstUpdate bool) ([]*rules.CircuitBreakerRule, error) {
	records, err := c.repo().GetMoreRulesByType(governanceRuleTypeCircuitBreaker, mtime, firstUpdate)
	if err != nil {
		log.Errorf("[Store][database] query circuitbreaker rules with mtime err: %s", err.Error())
		return nil, err
	}
	cbRules := make([]*rules.CircuitBreakerRule, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToCircuitBreakerRule(records[i])
		if err != nil {
			return nil, err
		}
		cbRules = append(cbRules, item)
	}
	return cbRules, nil
}

func (c *circuitBreakerStore) GetCircuitBreakerRule(id string) (*rules.CircuitBreakerRule, error) {
	record, err := c.repo().GetRuleByID(governanceRuleTypeCircuitBreaker, id)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleRecordToCircuitBreakerRule(record)
}

// LockCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) LockCircuitBreakerRule(tx store.Tx, namespace, keyword string) (*rules.CircuitBreakerRule, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}
	record, err := c.repo().LockRule(tx, governanceRuleTypeCircuitBreaker, keyword, namespace, keyword)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleRecordToCircuitBreakerRule(record)
}

// ActiveCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) ActiveCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	return c.repo().ActiveRelease(tx, circuitBreakerReleaseToGovernanceReleaseRecord(release))
}

// GetCircuitBreakerRuleVersions .
func (c *circuitBreakerStore) GetCircuitBreakerRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return c.repo().QueryReleaseVersions(ctx, governanceRuleTypeCircuitBreaker, model.RuleRelease_CircuitBreakerRules, filter, offset, limit)
}

// GetReleaseCircuitBreakerRule 获取处于使用状态的熔断规则
func (c *circuitBreakerStore) GetReleaseCircuitBreakerRule(tx store.Tx, req *rules.RuleRelease) (*rules.CircuitBreakerRelease, error) {
	record, err := c.repo().GetRelease(tx, governanceRuleTypeCircuitBreaker, req)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToCircuitBreakerRelease(record)
}

// GetActiveCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) GetActiveCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) (*rules.CircuitBreakerRelease, error) {
	record, err := c.repo().GetActiveRelease(tx, circuitBreakerReleaseToGovernanceReleaseRecord(release))
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToCircuitBreakerRelease(record)
}

// InactiveCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) InactiveCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	return c.repo().InactiveRelease(tx, circuitBreakerReleaseToGovernanceReleaseRecord(release))
}

// PublishCircuitBreakerRule implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) PublishCircuitBreakerRule(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	if release.ReleaseName == "" || release.ReleaseType == "" {
		return errors.New(
			"[store][mysql][circuitbreaker] publish circuit breaker rule missing some params",
		)
	}
	return c.repo().PublishRelease(tx, circuitBreakerReleaseToGovernanceReleaseRecord(release))
}

func (c *circuitBreakerStore) GetMoreCircuitBreakerReleases(mtime time.Time, firstUpdate bool) ([]*rules.CircuitBreakerRelease, error) {
	records, err := c.repo().GetMoreReleasesByType(governanceRuleTypeCircuitBreaker, firstUpdate, mtime)
	if err != nil {
		log.Errorf("[Store][database] query circuitbreaker releases with mtime err: %s", err.Error())
		return nil, err
	}
	releases := make([]*rules.CircuitBreakerRelease, 0, len(records))
	for i := range records {
		release, err := governanceRuleReleaseRecordToCircuitBreakerRelease(records[i])
		if err != nil {
			return nil, err
		}
		releases = append(releases, release)
	}
	return releases, nil
}

// DeleteCircuitBreakerReleases implements store.CircuitBreakerStore.
func (c *circuitBreakerStore) DeleteCircuitBreakerReleases(tx store.Tx, release *rules.CircuitBreakerRelease) error {
	return c.repo().DeleteRelease(tx, governanceRuleTypeCircuitBreaker, release.Id)
}
