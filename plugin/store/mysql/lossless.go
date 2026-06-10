package sqldb

import (
	"context"
	"errors"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

var _ store.LosslessRuleStore = (*losslessStore)(nil)

type losslessStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
}

// CreateLossLessRule 创建无损规则
func (s *losslessStore) CreateLossLessRule(rule *rules.LosslessRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][lossless] create lossless rule missing id")
	}
	err := RetryTransaction("createLosslessRule", func() error {
		return s.master.processWithTransaction("createLosslessRule", func(tx *BaseTx) error {
			if err := s.repo().CreateRule(NewSqlDBTx(tx), losslessRuleToGovernanceRuleRecord(rule)); err != nil {
				log.Errorf("[store][mysql][lossless] create lossless rule(%+v) err: %v", rule, err)
				return err
			}
			if err := tx.Commit(); err != nil {
				log.Errorf("[store][mysql][lossless] create lossless rule(%+v) commit err: %v", rule, err)
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

func (s *losslessStore) repo() *governanceRuleRepository {
	if s.governanceRuleRepository == nil {
		s.governanceRuleRepository = newGovernanceRuleRepository(s.master, s.slave)
	}
	return s.governanceRuleRepository
}

// UpdateLossLessRule 更新无损规则
func (s *losslessStore) UpdateLossLessRule(rule *rules.LosslessRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][lossless] update lossless rule missing id")
	}
	err := RetryTransaction("updateLosslessRule", func() error {
		return s.master.processWithTransaction("updateLosslessRule", func(tx *BaseTx) error {
			if err := s.repo().UpdateRule(NewSqlDBTx(tx), losslessRuleToGovernanceRuleRecord(rule)); err != nil {
				log.Errorf("[store][mysql][lossless] update lossless rule(%+v) err: %v", rule, err)
				return err
			}
			if err := tx.Commit(); err != nil {
				log.Errorf("[store][mysql][lossless] update lossless rule(%+v) commit err: %v", rule, err)
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

// DeleteLossLessRule 删除无损规则
func (s *losslessStore) DeleteLossLessRule(rule *rules.LosslessRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][lossless] delete lossless rule missing id")
	}
	err := RetryTransaction("deleteLosslessRule", func() error {
		return s.master.processWithTransaction("deleteLosslessRule", func(tx *BaseTx) error {
			if err := s.repo().DeleteRule(NewSqlDBTx(tx), governanceRuleTypeLossless, rule.ID); err != nil {
				log.Errorf("[store][mysql][lossless] delete lossless rule(%+v) err: %v", rule, err)
				return err
			}
			if err := tx.Commit(); err != nil {
				log.Errorf("[store][mysql][lossless] delete lossless rule(%+v) commit err: %v", rule, err)
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

// GetLossLessRuleWithID 根据无损ID拉取无损规则
func (s *losslessStore) GetOneLosslessRule(id string) (*rules.LosslessRule, error) {
	if id == "" {
		return nil, errors.New("id empty")
	}
	record, err := s.repo().GetRuleByID(governanceRuleTypeLossless, id)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleRecordToLosslessRule(record)
}

// GetMoreLosslessRules 根据修改时间拉取增量无损规则及最新版本号, 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
func (s *losslessStore) GetMoreLosslessRules(mtime time.Time, firstUpdate bool) ([]*rules.LosslessRule, error) {
	records, err := s.repo().GetMoreRulesByType(governanceRuleTypeLossless, mtime, firstUpdate)
	if err != nil {
		return nil, err
	}
	out := make([]*rules.LosslessRule, 0, len(records))
	for i := range records {
		item, err := governanceRuleRecordToLosslessRule(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// LockLosslessRule 锁住一个无损规则
func (s *losslessStore) LockLosslessRule(tx store.Tx, name string) (*rules.LosslessRule, error) {
	// Schema 中没有 name 字段，这里将 name 视为 id 处理
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if name == "" {
		return nil, ErrorMissingParams
	}
	record, err := s.repo().LockRule(tx, governanceRuleTypeLossless, name)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleRecordToLosslessRule(record)
}

// 关于规则发布
func (s *losslessStore) GetLosslessRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return s.repo().QueryReleaseVersions(ctx, governanceRuleTypeLossless, apimodel.RuleRelease_LosslessRules, filter, offset, limit)
}

// GetActiveLosslessRule 获取处于使用状态的无损规则
func (s *losslessStore) GetActiveLosslessRule(tx store.Tx, release *rules.LosslessRuleRelease) (*rules.LosslessRuleRelease, error) {
	record, err := s.repo().GetActiveRelease(tx, losslessRuleReleaseToGovernanceReleaseRecord(release))
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToLosslessRuleRelease(record)
}

// GetReleaseLosslessRule 获取已发布的无损规则
func (s *losslessStore) GetReleaseLosslessRule(tx store.Tx, release *rules.RuleRelease) (*rules.LosslessRuleRelease, error) {
	record, err := s.repo().GetRelease(tx, governanceRuleTypeLossless, release)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToLosslessRuleRelease(record)
}

// ActiveLosslessRule 设置某个无损规则发布为使用状态
func (s *losslessStore) ActiveLosslessRule(tx store.Tx, release *rules.LosslessRuleRelease) error {
	if release.RuleId == "" && release.Rule != nil {
		release.RuleId = release.Rule.ID
	}
	return s.repo().ActiveRelease(tx, losslessRuleReleaseToGovernanceReleaseRecord(release))
}

// InactiveLosslessRule 设置某个无损规则的发布为不使用状态
func (s *losslessStore) InactiveLosslessRule(tx store.Tx, release *rules.LosslessRuleRelease) error {
	if release.RuleId == "" && release.Rule != nil {
		release.RuleId = release.Rule.ID
	}
	return s.repo().InactiveRelease(tx, losslessRuleReleaseToGovernanceReleaseRecord(release))
}

// PublishLosslessRule 发布无损规则
func (s *losslessStore) PublishLosslessRules(tx store.Tx, rule *rules.LosslessRuleRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	if rule.Rule == nil || rule.Rule.ID == "" || rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][lossless] publish lossless rule missing params")
	}
	return s.repo().PublishRelease(tx, losslessRuleReleaseToGovernanceReleaseRecord(rule))
}

// 获取已发布的无损规则
func (s *losslessStore) GetMoreLosslessReleases(mtime time.Time, firstUpdate bool) ([]*rules.LosslessRuleRelease, error) {
	records, err := s.repo().GetMoreReleasesByType(governanceRuleTypeLossless, firstUpdate, mtime)
	if err != nil {
		return nil, err
	}
	out := make([]*rules.LosslessRuleRelease, 0, len(records))
	for i := range records {
		release, err := governanceRuleReleaseRecordToLosslessRuleRelease(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, release)
	}
	return out, nil
}

func (l *losslessStore) DeleteLosslessReleases(tx store.Tx, rule *rules.LosslessRuleRelease) error {
	return l.repo().DeleteRelease(tx, governanceRuleTypeLossless, rule.Id)
}
