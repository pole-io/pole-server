package aimcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

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

// ProbeSelfCapabilities verifies only Pole's reserved MCP projection and the
// requested tool names. It intentionally returns no registry record or schema,
// so a split Console can perform autonomous health checks without a user token.
func (h *HTTPServer) ProbeSelfCapabilities(ctx context.Context, allowlist []string) error {
	if h == nil || h.storage == nil {
		return errors.New("Pole MCP storage is not initialized")
	}
	registered, err := h.storage.GetMCPServerByName("pole-control-plane", "pole-system")
	if err != nil {
		return fmt.Errorf("resolve Pole MCP registry projection: %w", err)
	}
	if registered == nil || registered.GetFlag() == 1 ||
		registered.GetReference() != "pole-self-manager" ||
		registered.GetBackendType() != "address" {
		return errors.New("Pole MCP registry projection is unavailable or not self-managed")
	}
	endpoint, err := url.ParseRequestURI(strings.TrimSpace(registered.GetBackendAddress()))
	if err != nil || endpoint.Host == "" ||
		(endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return errors.New("Pole MCP registry projection has an invalid address")
	}
	tools, err := h.SnapshotTools(ctx)
	if err != nil {
		return err
	}
	available := make(map[string]struct{}, len(tools))
	for _, tool := range tools {
		if tool != nil {
			available[strings.TrimSpace(tool.GetName())] = struct{}{}
		}
	}
	for _, name := range allowlist {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := available[name]; !ok {
			return fmt.Errorf("Pole MCP tool %q is not available", name)
		}
	}
	return nil
}
