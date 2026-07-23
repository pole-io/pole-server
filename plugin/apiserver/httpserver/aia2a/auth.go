package aia2a

import (
	"context"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func (h *HTTPServer) checkA2AAgentPermission(ctx context.Context, op authtypes.ResourceOperation,
	method authtypes.ServerFunctionName, ids []string) (*authtypes.AcquireContext, error) {

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithModule(authtypes.MaintainModule),
		authtypes.WithOperation(op),
		authtypes.WithMethod(method),
		authtypes.WithAccessResources(a2aAgentResourceEntries(ids)),
	)
	_, err := h.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	return authCtx, err
}

func (h *HTTPServer) canReadA2AAgent(authCtx *authtypes.AcquireContext, id string) bool {
	return h.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
		Type: apisecurity.ResourceType_A2AAgentResources,
		ID:   id,
	})
}

func a2aAgentResourceEntries(ids []string) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(ids) == 0 {
		return nil
	}
	entries := make([]authtypes.ResourceEntry, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		entries = append(entries, authtypes.ResourceEntry{
			Type: apisecurity.ResourceType_A2AAgentResources,
			ID:   id,
		})
	}
	if len(entries) == 0 {
		return nil
	}
	return map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_A2AAgentResources: entries,
	}
}

func a2aAgentIDs(agents []*aitypes.A2AAgent) []string {
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		if agent == nil || agent.Id == "" {
			continue
		}
		ids = append(ids, agent.Id)
	}
	return ids
}
