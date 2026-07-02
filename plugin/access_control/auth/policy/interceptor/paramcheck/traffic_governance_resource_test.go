package paramcheck

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

type fakeTrafficGovernanceCache struct {
	rules map[string]*rules.TrafficGovernanceRule
}

func (f fakeTrafficGovernanceCache) Initialize(map[string]interface{}) error { return nil }
func (f fakeTrafficGovernanceCache) Update() error                           { return nil }
func (f fakeTrafficGovernanceCache) Clear() error                            { return nil }
func (f fakeTrafficGovernanceCache) Name() string                            { return "fake-traffic-governance" }
func (f fakeTrafficGovernanceCache) Close() error                            { return nil }
func (f fakeTrafficGovernanceCache) Query(context.Context, *cachetypes.TrafficGovernanceArgs) (uint32, []*rules.TrafficGovernanceRule, error) {
	return 0, nil, nil
}
func (f fakeTrafficGovernanceCache) GetRule(id string) *rules.TrafficGovernanceRule {
	return f.rules[id]
}
func (f fakeTrafficGovernanceCache) GetRulesForService(string, string) ([]*rules.TrafficGovernanceRule, string) {
	return nil, ""
}

func TestCheckTrafficGovernanceResourceExist(t *testing.T) {
	cache := fakeTrafficGovernanceCache{
		rules: map[string]*rules.TrafficGovernanceRule{
			"rule-1": {ID: "rule-1", Name: "security-rule"},
		},
	}

	require.Nil(t, checkTrafficGovernanceResourceExist([]*apisecurity.StrategyResourceEntry{
		{Id: "*"},
		{Id: "rule-1"},
	}, cache))

	resp := checkTrafficGovernanceResourceExist([]*apisecurity.StrategyResourceEntry{
		{Id: "missing-rule"},
	}, cache)
	require.NotNil(t, resp)
	require.Equal(t, uint32(apimodel.Code_NotFoundResource), resp.GetCode())
}
