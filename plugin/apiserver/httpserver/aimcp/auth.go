package aimcp

import (
	"context"

	apiai "github.com/pole-io/specification/source/go/api/v1/ai"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func (h *HTTPServer) checkMCPServerPermission(ctx context.Context, op authtypes.ResourceOperation,
	method authtypes.ServerFunctionName, ids []string) (*authtypes.AcquireContext, error) {

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithModule(authtypes.MaintainModule),
		authtypes.WithOperation(op),
		authtypes.WithMethod(method),
		authtypes.WithAccessResources(mcpServerResourceEntries(ids)),
	)
	_, err := h.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	return authCtx, err
}

func (h *HTTPServer) canReadMCPServer(authCtx *authtypes.AcquireContext, id string) bool {
	return h.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
		Type: apisecurity.ResourceType_MCPServerResources,
		ID:   id,
	})
}

func mcpServerResourceEntries(ids []string) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(ids) == 0 {
		return nil
	}
	entries := make([]authtypes.ResourceEntry, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		entries = append(entries, authtypes.ResourceEntry{
			Type: apisecurity.ResourceType_MCPServerResources,
			ID:   id,
		})
	}
	if len(entries) == 0 {
		return nil
	}
	return map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_MCPServerResources: entries,
	}
}

func mcpServerIDs(servers []*apiai.MCPServer) []string {
	ids := make([]string, 0, len(servers))
	for _, server := range servers {
		if server == nil || server.GetId() == "" {
			continue
		}
		ids = append(ids, server.GetId())
	}
	return ids
}
