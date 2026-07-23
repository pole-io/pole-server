package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func TestPrincipalResourceContainerMatchesFutureAIResourcesWithWildcard(t *testing.T) {
	resources := NewPrincipalResourceContainer()
	resources.SaveResource(apisecurity.AuthAction_ALLOW, authtypes.StrategyResource{
		StrategyID: "default-policy",
		ResType:    apisecurity.ResourceType_MCPServerResources,
		ResID:      "*",
	})
	resources.SaveResource(apisecurity.AuthAction_ALLOW, authtypes.StrategyResource{
		StrategyID: "default-policy",
		ResType:    apisecurity.ResourceType_A2AAgentResources,
		ResID:      "*",
	})

	for resourceType, resourceID := range map[apisecurity.ResourceType]string{
		apisecurity.ResourceType_MCPServerResources: "mcp-created-after-policy",
		apisecurity.ResourceType_A2AAgentResources:  "a2a-created-after-policy",
	} {
		action, matched := resources.Hint(resourceType, resourceID)
		require.True(t, matched)
		require.Equal(t, apisecurity.AuthAction_ALLOW, action)
	}
}
