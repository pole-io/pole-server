package sqldb

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestTrafficGovernanceRuleRecordRoundTrip(t *testing.T) {
	cases := []struct {
		name       string
		rule       *rules.TrafficGovernanceRule
		ruleType   governanceRuleType
		toRecord   func(*rules.TrafficGovernanceRule) *governanceRuleRecord
		fromRecord func(*governanceRuleRecord) (*rules.TrafficGovernanceRule, error)
	}{
		{
			name: "security",
			rule: rules.NewTrafficSecurityRule(&apisecurity.TrafficSecurityRule{
				Id:        "security-1",
				Name:      "security-rule",
				Namespace: "default",
				Service:   "svc-a",
				Enable:    true,
				Priority:  10,
				Metadata:  map[string]string{"owner": "qa"},
			}),
			ruleType:   governanceRuleTypeTrafficSecurity,
			toRecord:   trafficSecurityRuleToGovernanceRuleRecord,
			fromRecord: governanceRuleRecordToTrafficSecurityRule,
		},
		{
			name: "mirror",
			rule: rules.NewTrafficMirrorRule(&apitraffic.TrafficMirror{
				Id:        "mirror-1",
				Name:      "mirror-rule",
				Namespace: "default",
				Service:   "svc-a",
				Enable:    true,
				Priority:  20,
				Metadata:  map[string]string{"owner": "qa"},
			}),
			ruleType:   governanceRuleTypeTrafficMirror,
			toRecord:   trafficMirrorRuleToGovernanceRuleRecord,
			fromRecord: governanceRuleRecordToTrafficMirrorRule,
		},
		{
			name: "mock",
			rule: rules.NewTrafficMockRule(&apitraffic.TrafficMock{
				Id:        "mock-1",
				Name:      "mock-rule",
				Namespace: "default",
				Service:   "svc-a",
				Enable:    true,
				Priority:  30,
				Metadata:  map[string]string{"owner": "qa"},
			}),
			ruleType:   governanceRuleTypeTrafficMock,
			toRecord:   trafficMockRuleToGovernanceRuleRecord,
			fromRecord: governanceRuleRecordToTrafficMockRule,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.rule.Revision = "rev-1"
			tc.rule.Valid = true
			record := tc.toRecord(tc.rule)
			require.Equal(t, tc.rule.ID, record.ID)
			require.Equal(t, tc.ruleType, record.RuleType)
			require.Equal(t, tc.rule.Namespace, record.Namespace)
			require.Equal(t, tc.rule.Name, record.Name)
			require.Equal(t, tc.rule.Service, record.Service)
			require.Equal(t, int(tc.rule.Priority), record.Priority)
			require.Equal(t, 1, record.Enable)
			require.Equal(t, tc.rule.Revision, record.Revision)
			require.NotEmpty(t, record.Rule)

			got, err := tc.fromRecord(record)
			require.NoError(t, err)
			require.Equal(t, tc.rule.ID, got.ID)
			require.Equal(t, tc.rule.Name, got.Name)
			require.Equal(t, tc.rule.Namespace, got.Namespace)
			require.Equal(t, tc.rule.Service, got.Service)
			require.Equal(t, tc.rule.Priority, got.Priority)
			require.Equal(t, tc.rule.Metadata, got.Metadata)
			require.NotNil(t, got.Proto)
		})
	}
}
