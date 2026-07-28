package poleagent

import (
	"context"
	"encoding/json"
)

type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type Message struct {
	Role       MessageRole `json:"role"`
	Content    string      `json:"content,omitempty"`
	ToolCalls  []ToolCall  `json:"toolCalls,omitempty"`
	ToolCallID string      `json:"toolCallId,omitempty"`
}

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ModelRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

type ModelResponse struct {
	Message      Message `json:"message"`
	Model        string  `json:"model,omitempty"`
	FinishReason string  `json:"finishReason,omitempty"`
	RequestID    string  `json:"requestId,omitempty"`
}

// ModelPort is the only model-facing seam used by PoleAgent. Implementations
// must never receive Pole credentials or write-capable resource adapters.
type ModelPort interface {
	Complete(ctx context.Context, req ModelRequest) (ModelResponse, error)
}
