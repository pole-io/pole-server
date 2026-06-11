package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func TestEnrichResourceDetailHandlesDefaultPolicyResources(t *testing.T) {
	svr := &Server{}
	resp := &apisecurity.AuthStrategy{}

	require.NotPanics(t, func() {
		svr.enrichResourceInfo(context.Background(), resp, &authtypes.StrategyDetail{
			ID: "default-policy",
			Resources: []authtypes.StrategyResource{
				{StrategyID: "default-policy", ResType: apisecurity.ResourceType_LosslessRules, ResID: "*"},
				{StrategyID: "default-policy", ResType: apisecurity.ResourceType_MirrorRules, ResID: "*"},
				{StrategyID: "default-policy", ResType: apisecurity.ResourceType_SecurityRules, ResID: "*"},
				{StrategyID: "default-policy", ResType: apisecurity.ResourceType_MockRules, ResID: "*"},
			},
		})
	})

	require.Equal(t, "*", resp.Resources.GetLosslessRules()[0].GetId())
	require.Equal(t, "*", resp.Resources.GetMirrorRules()[0].GetId())
	require.Equal(t, "*", resp.Resources.GetSecurityRules()[0].GetId())
	require.Equal(t, "*", resp.Resources.GetMockRules()[0].GetId())
}

func TestEnrichResourceDetailSkipsUnsupportedResourceType(t *testing.T) {
	svr := &Server{}

	require.NotPanics(t, func() {
		svr.enrichResourceDetial(context.Background(), authtypes.StrategyResource{
			StrategyID: "unknown-policy",
			ResType:    apisecurity.ResourceType(999),
			ResID:      "*",
		}, map[apisecurity.ResourceType]struct{}{}, &apisecurity.AuthStrategy{
			Resources: &apisecurity.StrategyResources{},
		})
	})
}
