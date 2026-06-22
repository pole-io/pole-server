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
				Id:            "security-1",
				Name:          "security-rule",
				TargetService: &apitraffic.DestinationService{Namespace: "default", Service: "svc-a"},
				Enable:        true,
				Priority:      10,
				Metadata:      map[string]string{"owner": "qa"},
			}),
			ruleType:   governanceRuleTypeTrafficSecurity,
			toRecord:   trafficSecurityRuleToGovernanceRuleRecord,
			fromRecord: governanceRuleRecordToTrafficSecurityRule,
		},
		{
			name: "mirror",
			rule: rules.NewTrafficMirrorRule(&apitraffic.TrafficMirror{
				Id:       "mirror-1",
				Name:     "mirror-rule",
				Caller:   &apitraffic.SourceService{Namespace: "*", Service: "*"},
				Callee:   &apitraffic.DestinationService{Namespace: "default", Service: "svc-a"},
				Enable:   true,
				Priority: 20,
				Metadata: map[string]string{"owner": "qa"},
				Rules: []*apitraffic.MirrorRule{{
					TrafficMatchRule: &apitraffic.TrafficMatchRule{Arguments: []*apitraffic.SourceMatch{{
						Type: apitraffic.SourceMatch_CALLER_SERVICE,
						Key:  "default",
					}}},
				}},
			}),
			ruleType:   governanceRuleTypeTrafficMirror,
			toRecord:   trafficMirrorRuleToGovernanceRuleRecord,
			fromRecord: governanceRuleRecordToTrafficMirrorRule,
		},
		{
			name: "mock",
			rule: rules.NewTrafficMockRule(&apitraffic.TrafficMock{
				Id:       "mock-1",
				Name:     "mock-rule",
				Caller:   &apitraffic.SourceService{Namespace: "*", Service: "*"},
				Callee:   &apitraffic.DestinationService{Namespace: "default", Service: "svc-a"},
				Enable:   true,
				Priority: 30,
				Metadata: map[string]string{"owner": "qa"},
				Rules: []*apitraffic.MockRule{{
					TrafficMatchRule: &apitraffic.TrafficMatchRule{Arguments: []*apitraffic.SourceMatch{{
						Type: apitraffic.SourceMatch_CALLER_SERVICE,
						Key:  "default",
					}}},
				}},
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

func TestTrafficGovernanceReleaseRecordKeepsRuleSnapshotFields(t *testing.T) {
	rule := rules.NewTrafficSecurityRule(&apisecurity.TrafficSecurityRule{
		Id:            "security-1",
		Name:          "security-rule",
		Description:   "rule description",
		TargetService: &apitraffic.DestinationService{Namespace: "default", Service: "svc-a"},
		Enable:        true,
		Priority:      10,
		Revision:      "rev-1",
		Metadata:      map[string]string{"owner": "qa"},
	})
	rule.Valid = true
	release := &rules.TrafficGovernanceRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          "release-1",
			ReleaseName: "normal",
			RuleId:      rule.ID,
			RuleName:    rule.Name,
			ReleaseType: rules.ReleaseTypeNormal,
			Active:      true,
			Valid:       true,
		},
		Rule: rule,
	}

	record := trafficGovernanceRuleReleaseToGovernanceReleaseRecord(release, governanceRuleTypeTrafficSecurity)
	got, err := governanceRuleReleaseRecordToTrafficGovernanceRuleRelease(
		record,
		0,
		governanceRuleRecordToTrafficSecurityRule,
	)

	require.NoError(t, err)
	require.NotNil(t, got.Rule)
	require.Equal(t, rule.ID, got.Rule.ID)
	require.Equal(t, rule.Name, got.Rule.Name)
	require.Equal(t, rule.Namespace, got.Rule.Namespace)
	require.Equal(t, rule.Service, got.Rule.Service)
	require.Equal(t, rule.Description, got.Rule.Description)
	require.Equal(t, rule.Priority, got.Rule.Priority)
	require.True(t, got.Rule.Enable)
	require.Equal(t, rule.Revision, got.Rule.Revision)
	require.Equal(t, rule.Metadata, got.Rule.Metadata)
}

func TestTrafficMirrorRecordIgnoresLegacySourceField(t *testing.T) {
	record := &governanceRuleRecord{
		ID:        "mirror-legacy",
		RuleType:  governanceRuleTypeTrafficMirror,
		Namespace: "default",
		Name:      "mirror-rule",
		Service:   "checkout",
		Priority:  10,
		Enable:    1,
		Revision:  "rev-1",
		Valid:     true,
		Rule: `{
			"id": "mirror-legacy",
			"name": "mirror-rule",
			"target_service": {"namespace": "default", "service": "checkout"},
			"enable": true,
			"priority": 10,
			"rules": [{
				"source": {"namespace": "default", "service": "caller"},
				"mirror_percent": 50
			}]
		}`,
	}

	got, err := governanceRuleRecordToTrafficMirrorRule(record)

	require.NoError(t, err)
	require.Equal(t, "mirror-legacy", got.ID)
	require.Equal(t, "default", got.Namespace)
	require.Equal(t, "checkout", got.Service)
	require.NotNil(t, got.Proto)
}
