package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestTrafficGovernanceRuleUsesServiceEndpointAsScope(t *testing.T) {
	cases := []struct {
		name   string
		rule   *TrafficGovernanceRule
		spec   func(*TrafficGovernanceRule) *apitraffic.DestinationService
		caller func(*TrafficGovernanceRule) *apitraffic.SourceService
	}{
		{
			name: "security",
			rule: NewTrafficSecurityRule(&apisecurity.TrafficSecurityRule{
				TargetService: &apitraffic.DestinationService{Namespace: "default", Service: "checkout"},
			}),
			spec: func(rule *TrafficGovernanceRule) *apitraffic.DestinationService {
				return rule.ToTrafficSecuritySpec().GetTargetService()
			},
		},
		{
			name: "mirror",
			rule: NewTrafficMirrorRule(&apitraffic.TrafficMirror{
				Callee: &apitraffic.DestinationService{Namespace: "default", Service: "checkout"},
			}),
			spec: func(rule *TrafficGovernanceRule) *apitraffic.DestinationService {
				return rule.ToTrafficMirrorSpec().GetCallee()
			},
			caller: func(rule *TrafficGovernanceRule) *apitraffic.SourceService {
				return rule.ToTrafficMirrorSpec().GetCaller()
			},
		},
		{
			name: "mock",
			rule: NewTrafficMockRule(&apitraffic.TrafficMock{
				Callee: &apitraffic.DestinationService{Namespace: "default", Service: "checkout"},
			}),
			spec: func(rule *TrafficGovernanceRule) *apitraffic.DestinationService {
				return rule.ToTrafficMockSpec().GetCallee()
			},
			caller: func(rule *TrafficGovernanceRule) *apitraffic.SourceService {
				return rule.ToTrafficMockSpec().GetCaller()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, "default", tc.rule.Namespace)
			require.Equal(t, "checkout", tc.rule.Service)

			tc.rule.Namespace = "prod"
			tc.rule.Service = "payment"
			target := tc.spec(tc.rule)
			require.Equal(t, "prod", target.GetNamespace())
			require.Equal(t, "payment", target.GetService())
			if tc.caller != nil {
				caller := tc.caller(tc.rule)
				require.Equal(t, "*", caller.GetNamespace())
				require.Equal(t, "*", caller.GetService())
			}
		})
	}
}
