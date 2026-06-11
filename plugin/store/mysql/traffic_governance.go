package sqldb

import (
	"context"
	"errors"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

var (
	_ store.TrafficSecurityRuleStore = (*trafficSecurityStore)(nil)
	_ store.TrafficMirrorRuleStore   = (*trafficMirrorStore)(nil)
	_ store.TrafficMockRuleStore     = (*trafficMockStore)(nil)
)

type trafficGovernanceStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
}

type trafficSecurityStore struct{ trafficGovernanceStore }
type trafficMirrorStore struct{ trafficGovernanceStore }
type trafficMockStore struct{ trafficGovernanceStore }

func (s *trafficGovernanceStore) repo() *governanceRuleRepository {
	if s.governanceRuleRepository == nil {
		s.governanceRuleRepository = newGovernanceRuleRepository(s.master, s.slave)
	}
	return s.governanceRuleRepository
}

func (s *trafficGovernanceStore) createRule(kind string, rule *rules.TrafficGovernanceRule, toRecord func(*rules.TrafficGovernanceRule) *governanceRuleRecord) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][" + kind + "] create rule missing id")
	}
	err := RetryTransaction("create"+kind+"Rule", func() error {
		return s.master.processWithTransaction("create"+kind+"Rule", func(tx *BaseTx) error {
			if err := s.repo().CreateRule(NewSqlDBTx(tx), toRecord(rule)); err != nil {
				return err
			}
			return tx.Commit()
		})
	})
	return store.Error(err)
}

func (s *trafficGovernanceStore) updateRule(kind string, rule *rules.TrafficGovernanceRule, toRecord func(*rules.TrafficGovernanceRule) *governanceRuleRecord) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][" + kind + "] update rule missing id")
	}
	err := RetryTransaction("update"+kind+"Rule", func() error {
		return s.master.processWithTransaction("update"+kind+"Rule", func(tx *BaseTx) error {
			if err := s.repo().UpdateRule(NewSqlDBTx(tx), toRecord(rule)); err != nil {
				return err
			}
			return tx.Commit()
		})
	})
	return store.Error(err)
}

func (s *trafficGovernanceStore) deleteRule(kind string, ruleType governanceRuleType, rule *rules.TrafficGovernanceRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][" + kind + "] delete rule missing id")
	}
	err := RetryTransaction("delete"+kind+"Rule", func() error {
		return s.master.processWithTransaction("delete"+kind+"Rule", func(tx *BaseTx) error {
			if err := s.repo().DeleteRule(NewSqlDBTx(tx), ruleType, rule.ID); err != nil {
				return err
			}
			return tx.Commit()
		})
	})
	return store.Error(err)
}

func (s *trafficGovernanceStore) getOneRule(ruleType governanceRuleType, id string,
	fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error),
) (*rules.TrafficGovernanceRule, error) {
	if id == "" {
		return nil, errors.New("id empty")
	}
	record, err := s.repo().GetRuleByID(ruleType, id)
	if err != nil {
		return nil, store.Error(err)
	}
	return fromRecord(record)
}

func (s *trafficGovernanceStore) getMoreRules(ruleType governanceRuleType, mtime time.Time, firstUpdate bool,
	fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error),
) ([]*rules.TrafficGovernanceRule, error) {
	records, err := s.repo().GetMoreRulesByType(ruleType, mtime, firstUpdate)
	if err != nil {
		return nil, err
	}
	out := make([]*rules.TrafficGovernanceRule, 0, len(records))
	for i := range records {
		item, err := fromRecord(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *trafficGovernanceStore) lockRule(ruleType governanceRuleType, tx store.Tx, name string,
	fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error),
) (*rules.TrafficGovernanceRule, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if name == "" {
		return nil, ErrorMissingParams
	}
	record, err := s.repo().LockRule(tx, ruleType, name)
	if err != nil {
		return nil, store.Error(err)
	}
	return fromRecord(record)
}

func (s *trafficGovernanceStore) getActiveRelease(tx store.Tx, release *rules.TrafficGovernanceRuleRelease, ruleType governanceRuleType,
	resource apimodel.RuleRelease_RuleType, fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error),
) (*rules.TrafficGovernanceRuleRelease, error) {
	record, err := s.repo().GetActiveRelease(tx, trafficGovernanceRuleReleaseToGovernanceReleaseRecord(release, ruleType))
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToTrafficGovernanceRuleRelease(record, resource, fromRecord)
}

func (s *trafficGovernanceStore) getRelease(tx store.Tx, ruleType governanceRuleType, release *rules.RuleRelease,
	resource apimodel.RuleRelease_RuleType, fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error),
) (*rules.TrafficGovernanceRuleRelease, error) {
	record, err := s.repo().GetRelease(tx, ruleType, release)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToTrafficGovernanceRuleRelease(record, resource, fromRecord)
}

func (s *trafficGovernanceStore) publishRelease(kind string, tx store.Tx, release *rules.TrafficGovernanceRuleRelease, ruleType governanceRuleType) error {
	if tx == nil {
		return ErrTxIsNil
	}
	if release.Rule == nil || release.Rule.ID == "" || release.ReleaseName == "" || release.ReleaseType == "" {
		return errors.New("[store][mysql][" + kind + "] publish rule missing params")
	}
	return s.repo().PublishRelease(tx, trafficGovernanceRuleReleaseToGovernanceReleaseRecord(release, ruleType))
}

func (s *trafficGovernanceStore) getMoreReleases(ruleType governanceRuleType, mtime time.Time, firstUpdate bool,
	resource apimodel.RuleRelease_RuleType, fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error),
) ([]*rules.TrafficGovernanceRuleRelease, error) {
	records, err := s.repo().GetMoreReleasesByType(ruleType, firstUpdate, mtime)
	if err != nil {
		return nil, err
	}
	out := make([]*rules.TrafficGovernanceRuleRelease, 0, len(records))
	for i := range records {
		item, err := governanceRuleReleaseRecordToTrafficGovernanceRuleRelease(records[i], resource, fromRecord)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *trafficGovernanceStore) activeRelease(tx store.Tx, release *rules.TrafficGovernanceRuleRelease, ruleType governanceRuleType) error {
	if release.RuleId == "" && release.Rule != nil {
		release.RuleId = release.Rule.ID
	}
	return s.repo().ActiveRelease(tx, trafficGovernanceRuleReleaseToGovernanceReleaseRecord(release, ruleType))
}

func (s *trafficGovernanceStore) inactiveRelease(tx store.Tx, release *rules.TrafficGovernanceRuleRelease, ruleType governanceRuleType) error {
	if release.RuleId == "" && release.Rule != nil {
		release.RuleId = release.Rule.ID
	}
	return s.repo().InactiveRelease(tx, trafficGovernanceRuleReleaseToGovernanceReleaseRecord(release, ruleType))
}

func (s *trafficSecurityStore) CreateTrafficSecurityRule(rule *rules.TrafficGovernanceRule) error {
	return s.createRule("traffic-security", rule, trafficSecurityRuleToGovernanceRuleRecord)
}
func (s *trafficSecurityStore) UpdateTrafficSecurityRule(rule *rules.TrafficGovernanceRule) error {
	return s.updateRule("traffic-security", rule, trafficSecurityRuleToGovernanceRuleRecord)
}
func (s *trafficSecurityStore) DeleteTrafficSecurityRule(rule *rules.TrafficGovernanceRule) error {
	return s.deleteRule("traffic-security", governanceRuleTypeTrafficSecurity, rule)
}
func (s *trafficSecurityStore) GetOneTrafficSecurityRule(id string) (*rules.TrafficGovernanceRule, error) {
	return s.getOneRule(governanceRuleTypeTrafficSecurity, id, governanceRuleRecordToTrafficSecurityRule)
}
func (s *trafficSecurityStore) GetMoreTrafficSecurityRules(mtime time.Time, firstUpdate bool) ([]*rules.TrafficGovernanceRule, error) {
	return s.getMoreRules(governanceRuleTypeTrafficSecurity, mtime, firstUpdate, governanceRuleRecordToTrafficSecurityRule)
}
func (s *trafficSecurityStore) LockTrafficSecurityRule(tx store.Tx, name string) (*rules.TrafficGovernanceRule, error) {
	return s.lockRule(governanceRuleTypeTrafficSecurity, tx, name, governanceRuleRecordToTrafficSecurityRule)
}
func (s *trafficSecurityStore) GetTrafficSecurityRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return s.repo().QueryReleaseVersions(ctx, governanceRuleTypeTrafficSecurity, apimodel.RuleRelease_TrafficSecurityRules, filter, offset, limit)
}
func (s *trafficSecurityStore) GetActiveTrafficSecurityRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) (*rules.TrafficGovernanceRuleRelease, error) {
	return s.getActiveRelease(tx, release, governanceRuleTypeTrafficSecurity, apimodel.RuleRelease_TrafficSecurityRules, governanceRuleRecordToTrafficSecurityRule)
}
func (s *trafficSecurityStore) GetReleaseTrafficSecurityRule(tx store.Tx, release *rules.RuleRelease) (*rules.TrafficGovernanceRuleRelease, error) {
	return s.getRelease(tx, governanceRuleTypeTrafficSecurity, release, apimodel.RuleRelease_TrafficSecurityRules, governanceRuleRecordToTrafficSecurityRule)
}
func (s *trafficSecurityStore) ActiveTrafficSecurityRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.activeRelease(tx, release, governanceRuleTypeTrafficSecurity)
}
func (s *trafficSecurityStore) InactiveTrafficSecurityRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.inactiveRelease(tx, release, governanceRuleTypeTrafficSecurity)
}
func (s *trafficSecurityStore) PublishTrafficSecurityRules(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.publishRelease("traffic-security", tx, release, governanceRuleTypeTrafficSecurity)
}
func (s *trafficSecurityStore) GetMoreTrafficSecurityReleases(mtime time.Time, firstUpdate bool) ([]*rules.TrafficGovernanceRuleRelease, error) {
	return s.getMoreReleases(governanceRuleTypeTrafficSecurity, mtime, firstUpdate, apimodel.RuleRelease_TrafficSecurityRules, governanceRuleRecordToTrafficSecurityRule)
}
func (s *trafficSecurityStore) DeleteTrafficSecurityReleases(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.repo().DeleteRelease(tx, governanceRuleTypeTrafficSecurity, release.Id)
}

func (s *trafficMirrorStore) CreateTrafficMirrorRule(rule *rules.TrafficGovernanceRule) error {
	return s.createRule("traffic-mirror", rule, trafficMirrorRuleToGovernanceRuleRecord)
}
func (s *trafficMirrorStore) UpdateTrafficMirrorRule(rule *rules.TrafficGovernanceRule) error {
	return s.updateRule("traffic-mirror", rule, trafficMirrorRuleToGovernanceRuleRecord)
}
func (s *trafficMirrorStore) DeleteTrafficMirrorRule(rule *rules.TrafficGovernanceRule) error {
	return s.deleteRule("traffic-mirror", governanceRuleTypeTrafficMirror, rule)
}
func (s *trafficMirrorStore) GetOneTrafficMirrorRule(id string) (*rules.TrafficGovernanceRule, error) {
	return s.getOneRule(governanceRuleTypeTrafficMirror, id, governanceRuleRecordToTrafficMirrorRule)
}
func (s *trafficMirrorStore) GetMoreTrafficMirrorRules(mtime time.Time, firstUpdate bool) ([]*rules.TrafficGovernanceRule, error) {
	return s.getMoreRules(governanceRuleTypeTrafficMirror, mtime, firstUpdate, governanceRuleRecordToTrafficMirrorRule)
}
func (s *trafficMirrorStore) LockTrafficMirrorRule(tx store.Tx, name string) (*rules.TrafficGovernanceRule, error) {
	return s.lockRule(governanceRuleTypeTrafficMirror, tx, name, governanceRuleRecordToTrafficMirrorRule)
}
func (s *trafficMirrorStore) GetTrafficMirrorRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return s.repo().QueryReleaseVersions(ctx, governanceRuleTypeTrafficMirror, apimodel.RuleRelease_TrafficMirrorRules, filter, offset, limit)
}
func (s *trafficMirrorStore) GetActiveTrafficMirrorRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) (*rules.TrafficGovernanceRuleRelease, error) {
	return s.getActiveRelease(tx, release, governanceRuleTypeTrafficMirror, apimodel.RuleRelease_TrafficMirrorRules, governanceRuleRecordToTrafficMirrorRule)
}
func (s *trafficMirrorStore) GetReleaseTrafficMirrorRule(tx store.Tx, release *rules.RuleRelease) (*rules.TrafficGovernanceRuleRelease, error) {
	return s.getRelease(tx, governanceRuleTypeTrafficMirror, release, apimodel.RuleRelease_TrafficMirrorRules, governanceRuleRecordToTrafficMirrorRule)
}
func (s *trafficMirrorStore) ActiveTrafficMirrorRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.activeRelease(tx, release, governanceRuleTypeTrafficMirror)
}
func (s *trafficMirrorStore) InactiveTrafficMirrorRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.inactiveRelease(tx, release, governanceRuleTypeTrafficMirror)
}
func (s *trafficMirrorStore) PublishTrafficMirrorRules(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.publishRelease("traffic-mirror", tx, release, governanceRuleTypeTrafficMirror)
}
func (s *trafficMirrorStore) GetMoreTrafficMirrorReleases(mtime time.Time, firstUpdate bool) ([]*rules.TrafficGovernanceRuleRelease, error) {
	return s.getMoreReleases(governanceRuleTypeTrafficMirror, mtime, firstUpdate, apimodel.RuleRelease_TrafficMirrorRules, governanceRuleRecordToTrafficMirrorRule)
}
func (s *trafficMirrorStore) DeleteTrafficMirrorReleases(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.repo().DeleteRelease(tx, governanceRuleTypeTrafficMirror, release.Id)
}

func (s *trafficMockStore) CreateTrafficMockRule(rule *rules.TrafficGovernanceRule) error {
	return s.createRule("traffic-mock", rule, trafficMockRuleToGovernanceRuleRecord)
}
func (s *trafficMockStore) UpdateTrafficMockRule(rule *rules.TrafficGovernanceRule) error {
	return s.updateRule("traffic-mock", rule, trafficMockRuleToGovernanceRuleRecord)
}
func (s *trafficMockStore) DeleteTrafficMockRule(rule *rules.TrafficGovernanceRule) error {
	return s.deleteRule("traffic-mock", governanceRuleTypeTrafficMock, rule)
}
func (s *trafficMockStore) GetOneTrafficMockRule(id string) (*rules.TrafficGovernanceRule, error) {
	return s.getOneRule(governanceRuleTypeTrafficMock, id, governanceRuleRecordToTrafficMockRule)
}
func (s *trafficMockStore) GetMoreTrafficMockRules(mtime time.Time, firstUpdate bool) ([]*rules.TrafficGovernanceRule, error) {
	return s.getMoreRules(governanceRuleTypeTrafficMock, mtime, firstUpdate, governanceRuleRecordToTrafficMockRule)
}
func (s *trafficMockStore) LockTrafficMockRule(tx store.Tx, name string) (*rules.TrafficGovernanceRule, error) {
	return s.lockRule(governanceRuleTypeTrafficMock, tx, name, governanceRuleRecordToTrafficMockRule)
}
func (s *trafficMockStore) GetTrafficMockRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return s.repo().QueryReleaseVersions(ctx, governanceRuleTypeTrafficMock, apimodel.RuleRelease_TrafficMockRules, filter, offset, limit)
}
func (s *trafficMockStore) GetActiveTrafficMockRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) (*rules.TrafficGovernanceRuleRelease, error) {
	return s.getActiveRelease(tx, release, governanceRuleTypeTrafficMock, apimodel.RuleRelease_TrafficMockRules, governanceRuleRecordToTrafficMockRule)
}
func (s *trafficMockStore) GetReleaseTrafficMockRule(tx store.Tx, release *rules.RuleRelease) (*rules.TrafficGovernanceRuleRelease, error) {
	return s.getRelease(tx, governanceRuleTypeTrafficMock, release, apimodel.RuleRelease_TrafficMockRules, governanceRuleRecordToTrafficMockRule)
}
func (s *trafficMockStore) ActiveTrafficMockRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.activeRelease(tx, release, governanceRuleTypeTrafficMock)
}
func (s *trafficMockStore) InactiveTrafficMockRule(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.inactiveRelease(tx, release, governanceRuleTypeTrafficMock)
}
func (s *trafficMockStore) PublishTrafficMockRules(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.publishRelease("traffic-mock", tx, release, governanceRuleTypeTrafficMock)
}
func (s *trafficMockStore) GetMoreTrafficMockReleases(mtime time.Time, firstUpdate bool) ([]*rules.TrafficGovernanceRuleRelease, error) {
	return s.getMoreReleases(governanceRuleTypeTrafficMock, mtime, firstUpdate, apimodel.RuleRelease_TrafficMockRules, governanceRuleRecordToTrafficMockRule)
}
func (s *trafficMockStore) DeleteTrafficMockReleases(tx store.Tx, release *rules.TrafficGovernanceRuleRelease) error {
	return s.repo().DeleteRelease(tx, governanceRuleTypeTrafficMock, release.Id)
}
