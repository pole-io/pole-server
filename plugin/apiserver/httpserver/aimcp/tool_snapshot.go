package aimcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/pole-io/specification/source/go/api/v1/ai"
)

// SnapshotTools asks the in-process MCP server for tools/list so self
// registration is derived from the same protocol contract clients observe.
func (h *HTTPServer) SnapshotTools(ctx context.Context) ([]*ai.MCPServerTool, error) {
	if h == nil || h.mcpSvr == nil {
		return nil, errors.New("Pole MCP server is not initialized")
	}
	response := h.mcpSvr.HandleMessage(ctx, json.RawMessage(
		`{"jsonrpc":"2.0","id":"pole-self-manager","method":"tools/list","params":{}}`,
	))
	encoded, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Result mcp.ListToolsResult `json:"result"`
		Error  *mcp.JSONRPCError   `json:"error,omitempty"`
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return nil, err
	}
	if envelope.Error != nil {
		return nil, errors.New(envelope.Error.Error.Message)
	}
	tools := make([]*ai.MCPServerTool, 0, len(envelope.Result.Tools))
	for _, tool := range envelope.Result.Tools {
		schema, err := json.Marshal(tool)
		if err != nil {
			return nil, err
		}
		var definition struct {
			InputSchema json.RawMessage `json:"inputSchema"`
		}
		if err := json.Unmarshal(schema, &definition); err != nil {
			return nil, err
		}
		tools = append(tools, &ai.MCPServerTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: string(definition.InputSchema),
		})
	}
	return tools, nil
}
