package rules

import (
	"testing"
	"time"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/stretchr/testify/require"
)

func TestLosslessSetRulesClientSkipsMissingInactiveOrDeletedRelease(t *testing.T) {
	cache := &LossLessCache{
		BaseCache: cachebase.NewBaseCache(nil, nil),
		rules:     container.NewSyncMap[string, *ruletypes.LosslessRuleRelease](),
	}
	releases := []*ruletypes.LosslessRuleRelease{
		losslessReleaseForCache("deleted", false, false),
		losslessReleaseForCache("inactive", true, false),
		{RuleRelease: ruletypes.RuleRelease{Valid: true, Active: true}},
	}

	_, add, update, del := cache.setRulesClient(releases)

	require.Equal(t, 0, add)
	require.Equal(t, 0, update)
	require.Equal(t, 0, del)
}

func TestLosslessSetRulesClientStoresByActiveKey(t *testing.T) {
	cache := &LossLessCache{
		BaseCache: cachebase.NewBaseCache(nil, nil),
		rules:     container.NewSyncMap[string, *ruletypes.LosslessRuleRelease](),
	}
	release := losslessReleaseForCache("active", true, true)

	_, add, update, del := cache.setRulesClient([]*ruletypes.LosslessRuleRelease{release})

	require.Equal(t, 1, add)
	require.Equal(t, 0, update)
	require.Equal(t, 0, del)
	cached, ok := cache.rules.Load(release.ActiveKey())
	require.True(t, ok)
	require.Same(t, release, cached)
}

func losslessReleaseForCache(id string, valid bool, active bool) *ruletypes.LosslessRuleRelease {
	return &ruletypes.LosslessRuleRelease{
		RuleRelease: ruletypes.RuleRelease{
			Id:          id,
			ReleaseType: ruletypes.ReleaseTypeNormal,
			Valid:       valid,
			Active:      active,
			Version:     1,
			Mtime:       time.Unix(1, 0),
		},
		Rule: &ruletypes.LosslessRule{
			Namespace: "default",
			Service:   "svc",
		},
	}
}
