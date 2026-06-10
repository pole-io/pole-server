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

var _ store.RouterRuleConfigStore = (*routerRuleStore)(nil)

// routerRuleStore impl
type routerRuleStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
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
		return ErrTxIsNil
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.createRoutingConfigTx(dbTx, conf)
}

func (r *routerRuleStore) createRoutingConfigTx(tx *BaseTx, conf *rules.RouterConfig) error {
	if err := r.repo().CreateRule(NewSqlDBTx(tx), routerConfigToGovernanceRuleRecord(conf)); err != nil {
		log.Errorf("[Store][database] create routing (%+v) err: %s", conf, err.Error())
		return store.Error(err)
	}
	return nil
}

func (r *routerRuleStore) repo() *governanceRuleRepository {
	if r.governanceRuleRepository == nil {
		r.governanceRuleRepository = newGovernanceRuleRepository(r.master, r.slave)
	}
	return r.governanceRuleRepository
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
		return ErrTxIsNil
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

	if err := r.repo().UpdateRule(NewSqlDBTx(tx), routerConfigToGovernanceRuleRecord(conf)); err != nil {
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
		return r.master.processWithTransaction("EnableRouting", func(tx *BaseTx) error {
			save, err := r.repo().GetRuleByID(governanceRuleTypeRoute, conf.ID)
			if err != nil {
				return err
			}
			if save == nil {
				return nil
			}
			rule, err := governanceRuleRecordToRouterConfig(save)
			if err != nil {
				return err
			}
			rule.Enable = conf.Enable
			rule.Revision = conf.Revision
			if err := r.repo().UpdateRule(NewSqlDBTx(tx), routerConfigToGovernanceRuleRecord(rule)); err != nil {
				log.Errorf("[Store][database] update outing config (%+v), err: %s", conf, err)
				return err
			}
			return tx.Commit()
		})
	})

	return store.Error(err)
}

// DeleteRoutingConfig Delete a routing configuration
func (r *routerRuleStore) DeleteRoutingConfig(ruleID string) error {

	if ruleID == "" {
		log.Errorf("[Store][database] delete routing config  missing service id")
		return store.NewStatusError(store.EmptyParamsErr, "missing service id")
	}

	err := r.master.processWithTransaction("DeleteRoutingConfig", func(tx *BaseTx) error {
		if err := r.repo().DeleteRule(NewSqlDBTx(tx), governanceRuleTypeRoute, ruleID); err != nil {
			log.Errorf("[Store][database] delete routing config (%s) err: %s", ruleID, err.Error())
			return err
		}
		return tx.Commit()
	})

	return store.Error(err)
}

// GetMoreRouterRule Pull the incremental routing configuration information through mtime
func (r *routerRuleStore) GetMoreRouterRule(mtime time.Time, firstUpdate bool) ([]*rules.RouterConfig, error) {
	records, err := r.repo().GetMoreRulesByType(governanceRuleTypeRoute, mtime, firstUpdate)
	if err != nil {
		log.Errorf("[Store][database] query routing configs  with mtime err: %s", err.Error())
		return nil, err
	}
	out := make([]*rules.RouterConfig, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToRouterConfig(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// GetRoutingConfigWithID Pull the routing configuration according to the rules ID
func (r *routerRuleStore) GetRoutingConfigWithID(ruleID string) (*rules.RouterConfig, error) {
	record, err := r.repo().GetRuleByID(governanceRuleTypeRoute, ruleID)
	if err != nil {
		return nil, err
	}
	return governanceRuleRecordToRouterConfig(record)
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
	record, err := r.repo().GetRuleByID(governanceRuleTypeRoute, ruleID)
	if err != nil {
		log.Errorf("[Store][database] query routing  with id(%s) err: %s", ruleID, err.Error())
		return nil, err
	}
	return governanceRuleRecordToRouterConfig(record)
}

// GetRateLimitRuleVersions .
func (r *routerRuleStore) GetRouterRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return r.repo().QueryReleaseVersions(ctx, governanceRuleTypeRoute, model.RuleRelease_RouteRules, filter, offset, limit)
}

// LockRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) LockRouterRule(tx store.Tx, keyword string) (*rules.RouterConfig, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}
	record, err := r.repo().LockRule(tx, governanceRuleTypeRoute, keyword)
	if err != nil {
		log.Errorf("[Store][database] query routing  with keyword(%s) err: %s", keyword, err.Error())
		return nil, err
	}
	return governanceRuleRecordToRouterConfig(record)
}

// ActiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) ActiveRouterRule(tx store.Tx, release *rules.RouterRuleRelease) error {
	return r.repo().ActiveRelease(tx, routerRuleReleaseToGovernanceReleaseRecord(release))
}

// GetReleaseRouterRule 获取已发布的路由规则
func (r *routerRuleStore) GetReleaseRouterRule(tx store.Tx, release *rules.RuleRelease) (*rules.RouterRuleRelease, error) {
	record, err := r.repo().GetRelease(tx, governanceRuleTypeRoute, release)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToRouterRuleRelease(record)
}

// GetActiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) GetActiveRouterRule(tx store.Tx, release *rules.RouterRuleRelease) (*rules.RouterRuleRelease, error) {
	record, err := r.repo().GetActiveRelease(tx, routerRuleReleaseToGovernanceReleaseRecord(release))
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToRouterRuleRelease(record)
}

// InactiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) InactiveRouterRule(tx store.Tx, release *rules.RouterRuleRelease) error {
	return r.repo().InactiveRelease(tx, routerRuleReleaseToGovernanceReleaseRecord(release))
}

// PublishRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) PublishRouterRule(tx store.Tx, rule *rules.RouterRuleRelease) error {
	if rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][router] publish router rule missing some params")
	}
	return r.repo().PublishRelease(tx, routerRuleReleaseToGovernanceReleaseRecord(rule))
}

// GetMoreRouterRuleReleases implements store.RouterRuleConfigStore.
func (r *routerRuleStore) GetMoreRouterRuleReleases(firstUpdate bool, mtime time.Time) ([]*rules.RouterRuleRelease, error) {
	records, err := r.repo().GetMoreReleasesByType(governanceRuleTypeRoute, firstUpdate, mtime)
	if err != nil {
		return nil, err
	}
	out := make([]*rules.RouterRuleRelease, 0, len(records))
	for i := range records {
		release, err := governanceRuleReleaseRecordToRouterRuleRelease(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, release)
	}
	return out, nil
}

func (r *routerRuleStore) DeleteRouterRuleReleases(tx store.Tx, rule *rules.RouterRuleRelease) error {
	return r.repo().DeleteRelease(tx, governanceRuleTypeRoute, rule.Id)
}

// fetchRoutingConfigRows Read the data of the database and release ROWS
