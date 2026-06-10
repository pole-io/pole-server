package rules

import (
	"testing"
	"time"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/stretchr/testify/require"
)

func TestGovernanceRuleUpdateCacheFetchesOnceAndFansOut(t *testing.T) {
	store := &fakeGovernanceRuleUpdateStore{
		storeTime: 10,
		updates: &ruletypes.GovernanceRuleUpdates{
			RateLimitRules: []*ruletypes.RateLimit{{ID: "limit-1", Valid: true}},
		},
		releases: &ruletypes.GovernanceRuleReleaseUpdates{
			RateLimitRules: []*ruletypes.RateLimitRelease{{
				RuleRelease: ruletypes.RuleRelease{Id: "release-1", Valid: true, Active: true},
				Rule:        &ruletypes.RateLimit{ID: "limit-1", Valid: true},
			}},
		},
	}
	watcher := &fakeGovernanceRuleUpdateWatcher{}
	cache := &governanceRuleUpdateCache{
		storage:       store,
		firstUpdate:   true,
		lastFetchTime: 1,
		watchers:      []governanceRuleUpdateWatcher{watcher},
	}

	require.NoError(t, cache.Update())
	require.Equal(t, 1, store.ruleCalls)
	require.Equal(t, 1, store.releaseCalls)
	require.Equal(t, 1, watcher.calls)
	require.Len(t, watcher.updates.RateLimitRules, 1)
	require.Len(t, watcher.releases.RateLimitRules, 1)

	require.NoError(t, cache.Update())
	require.Equal(t, 1, store.ruleCalls)
	require.Equal(t, 1, store.releaseCalls)
	require.Equal(t, 1, watcher.calls)
}

type fakeGovernanceRuleUpdateStore struct {
	storeTime    int64
	ruleCalls    int
	releaseCalls int
	updates      *ruletypes.GovernanceRuleUpdates
	releases     *ruletypes.GovernanceRuleReleaseUpdates
}

func (f *fakeGovernanceRuleUpdateStore) GetUnixSecond(offset int64) (int64, error) {
	return f.storeTime + offset, nil
}

func (f *fakeGovernanceRuleUpdateStore) GetMoreGovernanceRuleUpdates(time.Time, bool) (*ruletypes.GovernanceRuleUpdates, error) {
	f.ruleCalls++
	return f.updates, nil
}

func (f *fakeGovernanceRuleUpdateStore) GetMoreGovernanceRuleReleaseUpdates(time.Time, bool) (*ruletypes.GovernanceRuleReleaseUpdates, error) {
	f.releaseCalls++
	return f.releases, nil
}

type fakeGovernanceRuleUpdateWatcher struct {
	calls    int
	updates  *ruletypes.GovernanceRuleUpdates
	releases *ruletypes.GovernanceRuleReleaseUpdates
}

func (f *fakeGovernanceRuleUpdateWatcher) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	f.calls++
	f.updates = updates
	f.releases = releases
	return 0, nil
}
