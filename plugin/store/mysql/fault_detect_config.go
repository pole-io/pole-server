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

var _ store.FaultDetectRuleStore = (*faultDetectRuleStore)(nil)

type faultDetectRuleStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
}

const (
	labelCreateFaultDetectRule = "createFaultDetectRule"
	labelUpdateFaultDetectRule = "updateFaultDetectRule"
	labelDeleteFaultDetectRule = "deleteFaultDetectRule"
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
		if err := f.repo().CreateRule(NewSqlDBTx(tx), faultDetectRuleToGovernanceRuleRecord(fdRule)); err != nil {
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

func (f *faultDetectRuleStore) repo() *governanceRuleRepository {
	if f.governanceRuleRepository == nil {
		f.governanceRuleRepository = newGovernanceRuleRepository(f.master, f.slave)
	}
	return f.governanceRuleRepository
}

// UpdateFaultDetectRule update fault detect rule
func (f *faultDetectRuleStore) UpdateFaultDetectRule(fdRule *rules.FaultDetectRule) error {
	err := RetryTransaction(labelUpdateFaultDetectRule, func() error {
		return f.master.processWithTransaction(labelUpdateFaultDetectRule, func(tx *BaseTx) error {
			if err := f.repo().UpdateRule(NewSqlDBTx(tx), faultDetectRuleToGovernanceRuleRecord(fdRule)); err != nil {
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
		if err := f.repo().DeleteRule(NewSqlDBTx(tx), governanceRuleTypeFaultDetect, id); err != nil {
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
	delete(filter, briefSearch)
	num, records, err := f.repo().QueryRules(context.Background(), governanceRuleTypeFaultDetect, filter, offset, limit)
	if err != nil {
		return 0, nil, err
	}
	out := make([]*rules.FaultDetectRule, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToFaultDetectRule(records[i])
		if err != nil {
			return 0, nil, err
		}
		out = append(out, item)
	}
	return num, out, nil
}

// GetMoreFaultDetects(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRule, error) get increment circuitbreaker rules
func (f *faultDetectRuleStore) GetMoreFaultDetects(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRule, error) {
	records, err := f.repo().GetMoreRulesByType(governanceRuleTypeFaultDetect, mtime, firstUpdate)
	if err != nil {
		log.Errorf("[Store][database] query fault detect rules with mtime err: %s", err.Error())
		return nil, err
	}
	fdRules := make([]*rules.FaultDetectRule, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToFaultDetectRule(records[i])
		if err != nil {
			return nil, err
		}
		fdRules = append(fdRules, item)
	}
	return fdRules, nil
}

// GetFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) GetFaultDetectRule(id string) (*rules.FaultDetectRule, error) {
	record, err := f.repo().GetRuleByID(governanceRuleTypeFaultDetect, id)
	if err != nil {
		return nil, err
	}
	return governanceRuleRecordToFaultDetectRule(record)
}

// LockFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) LockFaultDetectRule(tx store.Tx, namespace, keyword string) (*rules.FaultDetectRule, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}
	record, err := f.repo().LockRule(tx, governanceRuleTypeFaultDetect, keyword, namespace, keyword)
	if err != nil {
		return nil, err
	}
	return governanceRuleRecordToFaultDetectRule(record)
}

// ActiveFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) ActiveFaultDetectRule(tx store.Tx, release *rules.FaultDetectRelease) error {
	return f.repo().ActiveRelease(tx, faultDetectReleaseToGovernanceReleaseRecord(release))
}

// GetFaultDetectRuleVersions .
func (f *faultDetectRuleStore) GetFaultDetectRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return f.repo().QueryReleaseVersions(ctx, governanceRuleTypeFaultDetect, model.RuleRelease_FaultDetectRules, filter, offset, limit)
}

func (f *faultDetectRuleStore) GetReleaseFaultDetectRule(tx store.Tx, release *rules.RuleRelease) (*rules.FaultDetectRelease, error) {
	record, err := f.repo().GetRelease(tx, governanceRuleTypeFaultDetect, release)
	if err != nil {
		return nil, err
	}
	return governanceRuleReleaseRecordToFaultDetectRelease(record)
}

// GetActiveFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) GetActiveFaultDetectRule(tx store.Tx, release *rules.FaultDetectRelease) (*rules.FaultDetectRelease, error) {
	record, err := f.repo().GetActiveRelease(tx, faultDetectReleaseToGovernanceReleaseRecord(release))
	if err != nil {
		return nil, err
	}
	return governanceRuleReleaseRecordToFaultDetectRelease(record)
}

// InactiveFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) InactiveFaultDetectRule(tx store.Tx, release *rules.FaultDetectRelease) error {
	return f.repo().InactiveRelease(tx, faultDetectReleaseToGovernanceReleaseRecord(release))
}

// PublishFaultDetectRule implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) PublishFaultDetectRule(tx store.Tx, rule *rules.FaultDetectRelease) error {
	if rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][faultdetect] publish fault detect rule missing some params")
	}
	return f.repo().PublishRelease(tx, faultDetectReleaseToGovernanceReleaseRecord(rule))
}

// GetMoreFaultDetectReleases implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) GetMoreFaultDetectReleases(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRelease, error) {
	records, err := f.repo().GetMoreReleasesByType(governanceRuleTypeFaultDetect, firstUpdate, mtime)
	if err != nil {
		return nil, err
	}
	out := make([]*rules.FaultDetectRelease, 0, len(records))
	for i := range records {
		release, err := governanceRuleReleaseRecordToFaultDetectRelease(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, release)
	}
	return out, nil
}

// DeleteFaultDetectReleases implements store.FaultDetectRuleStore.
func (f *faultDetectRuleStore) DeleteFaultDetectReleases(tx store.Tx, release *rules.FaultDetectRelease) error {
	return f.repo().DeleteRelease(tx, governanceRuleTypeFaultDetect, release.Id)
}

func (f *faultDetectRuleStore) getFaultDetectRulesCount(filter map[string]string) (uint32, error) {
	total, _, err := f.repo().QueryRules(context.Background(), governanceRuleTypeFaultDetect, filter, 0, 0)
	return total, err
}
