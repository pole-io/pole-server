package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	model "github.com/pole-io/specification/source/go/api/v1/model"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

func TestCircuitBreakerGrayMismatchFallsBackToNamespaceRelease(t *testing.T) {
	exact := ruletypes.NewServiceWithCircuitBreakerRules(svctypes.ServiceKey{Namespace: "prod", Name: "orders"})
	exact.AddCircuitBreakerRule(circuitRelease("exact-gray", "exact-rule", ruletypes.ReleaseTypeGray, exactLabel("env", "canary")))
	namespace := ruletypes.NewServiceWithCircuitBreakerRules(svctypes.ServiceKey{Namespace: "prod", Name: "*"})
	namespace.AddCircuitBreakerRule(circuitRelease("namespace-normal", "namespace-rule", ruletypes.ReleaseTypeNormal, nil))

	cache := &circuitBreakerCache{
		circuitBreakers: map[string]map[string]*ruletypes.ServiceWithCircuitBreakerRules{
			"prod": {"orders": exact},
		},
		nsWildcardRules: map[string]*ruletypes.ServiceWithCircuitBreakerRules{"prod": namespace},
		allWildcardRules: ruletypes.NewServiceWithCircuitBreakerRules(
			svctypes.ServiceKey{Namespace: "*", Name: "*"},
		),
	}

	got, revision := cache.GetCircuitBreakerConfigWithLabels("orders", "prod", map[string]string{"env": "stable"})

	require.NotNil(t, got)
	require.NotEmpty(t, revision)
	var names []string
	got.IterateCircuitBreakerRules(func(release *ruletypes.CircuitBreakerRelease) {
		names = append(names, release.ReleaseName)
	})
	require.Equal(t, []string{"namespace-normal"}, names)
}

func TestFaultDetectGrayMismatchFallsBackToGlobalRelease(t *testing.T) {
	exact := ruletypes.NewServiceWithFaultDetectRules(svctypes.ServiceKey{Namespace: "prod", Name: "orders"})
	exact.AddFaultDetectRule(faultRelease("exact-gray", "exact-rule", ruletypes.ReleaseTypeGray, exactLabel("env", "canary")))
	global := ruletypes.NewServiceWithFaultDetectRules(svctypes.ServiceKey{Namespace: "*", Name: "*"})
	global.AddFaultDetectRule(faultRelease("global-normal", "global-rule", ruletypes.ReleaseTypeNormal, nil))

	serviceRules := container.NewSyncMap[string, *container.SyncMap[string, *ruletypes.ServiceWithFaultDetectRules]]()
	namespaceRules := container.NewSyncMap[string, *ruletypes.ServiceWithFaultDetectRules]()
	prod := container.NewSyncMap[string, *ruletypes.ServiceWithFaultDetectRules]()
	prod.Store("orders", exact)
	serviceRules.Store("prod", prod)
	cache := &faultDetectCache{
		svcSpecificRules: serviceRules,
		nsWildcardRules:  namespaceRules,
		allWildcardRules: global,
	}

	got, revision := cache.GetFaultDetectConfigWithLabels("orders", "prod", map[string]string{"env": "stable"})

	require.NotNil(t, got)
	require.NotEmpty(t, revision)
	var names []string
	got.IterateFaultDetectRules(func(release *ruletypes.FaultDetectRelease) {
		names = append(names, release.ReleaseName)
	})
	require.Equal(t, []string{"global-normal"}, names)
}

func circuitRelease(
	name, ruleID string, releaseType ruletypes.ReleaseType, labels []*model.ClientLabel,
) *ruletypes.CircuitBreakerRelease {
	return &ruletypes.CircuitBreakerRelease{
		RuleRelease: hierarchyRelease(name, ruleID, releaseType, labels),
		Rule:        &ruletypes.CircuitBreakerRule{ID: ruleID},
	}
}

func faultRelease(
	name, ruleID string, releaseType ruletypes.ReleaseType, labels []*model.ClientLabel,
) *ruletypes.FaultDetectRelease {
	return &ruletypes.FaultDetectRelease{
		RuleRelease: hierarchyRelease(name, ruleID, releaseType, labels),
		Rule:        &ruletypes.FaultDetectRule{ID: ruleID},
	}
}

func hierarchyRelease(
	name, ruleID string, releaseType ruletypes.ReleaseType, labels []*model.ClientLabel,
) ruletypes.RuleRelease {
	return ruletypes.RuleRelease{
		Id: name + "-id", Namespace: "prod", RuleId: ruleID, RuleName: ruleID,
		ReleaseName: name, ReleaseType: releaseType, ClientLabels: labels,
		Active: true, Valid: true, Version: 1, Mtime: time.Unix(1, 0),
	}
}
