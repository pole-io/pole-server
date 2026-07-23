package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestTrafficGovernanceRuleSeparatesOwnerNamespaceFromServiceScope(t *testing.T) {
	cases := []struct {
		name   string
		rule   *TrafficGovernanceRule
		spec   func(*TrafficGovernanceRule) *apitraffic.DestinationService
		owner  func(*TrafficGovernanceRule) string
		caller func(*TrafficGovernanceRule) *apitraffic.SourceService
	}{
		{
			name: "security",
			rule: NewTrafficSecurityRule(&apisecurity.TrafficSecurityRule{
				Namespace:     "prod",
				TargetService: &apitraffic.DestinationService{Namespace: "default", Service: "checkout"},
			}),
			spec: func(rule *TrafficGovernanceRule) *apitraffic.DestinationService {
				return rule.ToTrafficSecuritySpec().GetTargetService()
			},
			owner: func(rule *TrafficGovernanceRule) string { return rule.ToTrafficSecuritySpec().GetNamespace() },
		},
		{
			name: "mirror",
			rule: NewTrafficMirrorRule(&apitraffic.TrafficMirror{
				Namespace: "prod",
				Callee:    &apitraffic.DestinationService{Namespace: "default", Service: "checkout"},
			}),
			spec: func(rule *TrafficGovernanceRule) *apitraffic.DestinationService {
				return rule.ToTrafficMirrorSpec().GetCallee()
			},
			owner: func(rule *TrafficGovernanceRule) string { return rule.ToTrafficMirrorSpec().GetNamespace() },
			caller: func(rule *TrafficGovernanceRule) *apitraffic.SourceService {
				return rule.ToTrafficMirrorSpec().GetCaller()
			},
		},
		{
			name: "mock",
			rule: NewTrafficMockRule(&apitraffic.TrafficMock{
				Namespace: "prod",
				Callee:    &apitraffic.DestinationService{Namespace: "default", Service: "checkout"},
			}),
			spec: func(rule *TrafficGovernanceRule) *apitraffic.DestinationService {
				return rule.ToTrafficMockSpec().GetCallee()
			},
			owner: func(rule *TrafficGovernanceRule) string { return rule.ToTrafficMockSpec().GetNamespace() },
			caller: func(rule *TrafficGovernanceRule) *apitraffic.SourceService {
				return rule.ToTrafficMockSpec().GetCaller()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, "prod", tc.rule.Namespace)
			require.Equal(t, "default", tc.rule.ServiceNamespace)
			require.Equal(t, "checkout", tc.rule.Service)

			tc.rule.Namespace = "staging"
			tc.rule.ServiceNamespace = "shared"
			tc.rule.Service = "payment"
			target := tc.spec(tc.rule)
			require.Equal(t, "staging", tc.owner(tc.rule))
			require.Equal(t, "shared", target.GetNamespace())
			require.Equal(t, "payment", target.GetService())
			if tc.caller != nil {
				caller := tc.caller(tc.rule)
				require.Equal(t, "*", caller.GetNamespace())
				require.Equal(t, "*", caller.GetService())
			}
		})
	}
}
