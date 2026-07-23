package policy

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestBackfillDefaultPolicyAIResourceWildcards(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := storemock.NewMockStore(ctrl)
	policies := []*authtypes.StrategyDetail{
		{
			ID:      "main-default",
			Default: true,
			Valid:   true,
			Comment: defaultMainUserPolicyComment,
			Resources: []authtypes.StrategyResource{
				{StrategyID: "main-default", ResType: apisecurity.ResourceType_Services, ResID: "*"},
				{StrategyID: "main-default", ResType: apisecurity.ResourceType_MCPServerResources, ResID: "mcp-existing"},
			},
		},
		{
			ID:      "global-read-write",
			Default: true,
			Valid:   true,
			Comment: defaultReadWritePolicyComment,
			Metadata: map[string]string{
				authtypes.MetadKeySystemDefaultPolicy: "true",
			},
			Resources: []authtypes.StrategyResource{
				{StrategyID: "global-read-write", ResType: apisecurity.ResourceType_MCPServerResources, ResID: "*"},
			},
		},
		{
			ID:      "custom-default",
			Default: true,
			Valid:   true,
			Comment: "custom default policy",
		},
	}

	storage.EXPECT().GetMoreStrategies(gomock.Any(), true).Return(policies, nil)
	storage.EXPECT().LooseAddStrategyResources(gomock.Any()).DoAndReturn(func(resources []authtypes.StrategyResource) error {
		require.ElementsMatch(t, []authtypes.StrategyResource{
			{StrategyID: "main-default", ResType: apisecurity.ResourceType_MCPServerResources, ResID: "*"},
			{StrategyID: "main-default", ResType: apisecurity.ResourceType_A2AAgentResources, ResID: "*"},
			{StrategyID: "global-read-write", ResType: apisecurity.ResourceType_A2AAgentResources, ResID: "*"},
		}, resources)
		return nil
	})

	count, err := backfillDefaultPolicyAIResourceWildcards(storage)
	require.NoError(t, err)
	require.Equal(t, 3, count)
}

func TestBackfillDefaultPolicyAIResourceWildcardsIsIdempotent(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := storemock.NewMockStore(ctrl)
	storage.EXPECT().GetMoreStrategies(gomock.Any(), true).Return([]*authtypes.StrategyDetail{
		{
			ID:      "main-default",
			Default: true,
			Valid:   true,
			Comment: defaultMainUserPolicyComment,
			Resources: []authtypes.StrategyResource{
				{StrategyID: "main-default", ResType: apisecurity.ResourceType_MCPServerResources, ResID: "*"},
				{StrategyID: "main-default", ResType: apisecurity.ResourceType_A2AAgentResources, ResID: "*"},
			},
		},
	}, nil)

	count, err := backfillDefaultPolicyAIResourceWildcards(storage)
	require.NoError(t, err)
	require.Zero(t, count)
}
