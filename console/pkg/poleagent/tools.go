package poleagent

import (
	"context"

	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
)

type ToolResult struct {
	Content   string `json:"content"`
	IsError   bool   `json:"isError"`
	RequestID string `json:"requestId,omitempty"`
}

// ToolPort creates an actor-bound tool session for one Agent turn. Binding
// credentials at session creation prevents one user's MCP connection from
// being reused by another user.
type ToolPort interface {
	Open(ctx context.Context, actor agentworkbench.Actor) (ToolSession, error)
}

type ToolSession interface {
	ListTools(ctx context.Context) ([]ToolDefinition, error)
	CallTool(ctx context.Context, name string, arguments map[string]any) (ToolResult, error)
	Close() error
}
