package poleagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
)

const (
	maxMCPToolPages   = 20
	maxMCPResultBytes = 256 << 10
)

type MCPToolConfig struct {
	Endpoint      string
	ToolAllowlist []string
	ClientName    string
	ClientVersion string
	ReadTimeout   time.Duration
}

type MCPToolPort struct {
	config  MCPToolConfig
	allowed map[string]struct{}
}

func NewMCPToolPort(cfg MCPToolConfig) (*MCPToolPort, error) {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	if cfg.Endpoint == "" {
		return nil, runtimeError(CategoryRuntimeUnavailable, 503121,
			"Pole MCP endpoint is not configured", false, nil)
	}
	if cfg.ClientName == "" {
		cfg.ClientName = "pole-console-agent"
	}
	if cfg.ClientVersion == "" {
		cfg.ClientVersion = "v1"
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 60 * time.Second
	}
	allowed := make(map[string]struct{}, len(cfg.ToolAllowlist))
	for _, name := range cfg.ToolAllowlist {
		if name = strings.TrimSpace(name); name != "" {
			allowed[name] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		return nil, runtimeError(CategoryRuntimeUnavailable, 503122,
			"Pole MCP tool allowlist is empty", false, nil)
	}
	return &MCPToolPort{config: cfg, allowed: allowed}, nil
}

func (p *MCPToolPort) Open(ctx context.Context, actor agentworkbench.Actor) (ToolSession, error) {
	if strings.TrimSpace(actor.UserID) == "" || strings.TrimSpace(actor.Token) == "" {
		return nil, runtimeError(CategoryInvalidRequest, 400121,
			"authenticated actor is required for Pole MCP", false, ErrInvalidRequest)
	}
	headers := map[string]string{
		"X-Pole-User":   actor.UserID,
		"Authorization": bearerToken(actor.Token),
	}
	if actor.RequestID != "" {
		headers["X-Request-Id"] = actor.RequestID
	}
	client, err := mcpclient.NewSSEMCPClient(
		p.config.Endpoint,
		mcpclient.WithHeaders(headers),
		mcpclient.WithSSEReadTimeout(p.config.ReadTimeout),
	)
	if err != nil {
		return nil, runtimeError(CategoryToolUnavailable, 503123,
			"create Pole MCP client", false, err)
	}
	if err := client.Start(ctx); err != nil {
		_ = client.Close()
		return nil, runtimeError(CategoryToolUnavailable, 503124,
			"connect Pole MCP", true, err)
	}
	initialize := mcp.InitializeRequest{}
	initialize.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initialize.Params.ClientInfo = mcp.Implementation{
		Name:    p.config.ClientName,
		Version: p.config.ClientVersion,
	}
	if _, err := client.Initialize(ctx, initialize); err != nil {
		_ = client.Close()
		return nil, runtimeError(CategoryToolUnavailable, 503125,
			"initialize Pole MCP session", true, err)
	}
	return &mcpToolSession{client: client, allowed: p.allowed}, nil
}

type mcpToolSession struct {
	client  mcpclient.MCPClient
	allowed map[string]struct{}
}

func (s *mcpToolSession) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	var (
		cursor mcp.Cursor
		tools  []ToolDefinition
		seen   = map[string]struct{}{}
	)
	for page := 0; page < maxMCPToolPages; page++ {
		request := mcp.ListToolsRequest{}
		request.Params.Cursor = cursor
		result, err := s.client.ListTools(ctx, request)
		if err != nil {
			return nil, runtimeError(CategoryToolUnavailable, 503126,
				"list Pole MCP tools", true, err)
		}
		for _, tool := range result.Tools {
			if _, ok := s.allowed[tool.Name]; !ok {
				continue
			}
			if _, duplicate := seen[tool.Name]; duplicate {
				return nil, runtimeError(CategoryToolUnavailable, 503127,
					"Pole MCP returned a duplicate tool name", false, nil)
			}
			schema := map[string]any{
				"type":       tool.InputSchema.Type,
				"properties": tool.InputSchema.Properties,
			}
			if schema["type"] == "" {
				schema["type"] = "object"
			}
			if tool.InputSchema.Required != nil {
				schema["required"] = tool.InputSchema.Required
			}
			tools = append(tools, ToolDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: schema,
			})
			seen[tool.Name] = struct{}{}
		}
		if result.NextCursor == "" {
			return tools, nil
		}
		cursor = result.NextCursor
	}
	return nil, runtimeError(CategoryToolUnavailable, 503128,
		"Pole MCP tool pagination exceeded its limit", false, nil)
}

func (s *mcpToolSession) CallTool(ctx context.Context, name string, arguments map[string]any) (ToolResult, error) {
	if _, ok := s.allowed[name]; !ok {
		return ToolResult{}, runtimeError(CategoryToolRejected, 403121,
			"tool is not allowed by the Pole MCP allowlist", false, nil)
	}
	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = arguments
	result, err := s.client.CallTool(ctx, request)
	if err != nil {
		return ToolResult{}, runtimeError(CategoryDownstreamFailed, 502121,
			"Pole MCP tool call failed", true, err)
	}
	content, err := mcpTextContent(result.Content)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{Content: content, IsError: result.IsError}, nil
}

func (s *mcpToolSession) Close() error {
	return s.client.Close()
}

func bearerToken(token string) string {
	token = strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token
	}
	return "Bearer " + token
}

func mcpTextContent(contents []mcp.Content) (string, error) {
	parts := make([]string, 0, len(contents))
	size := 0
	for _, content := range contents {
		var text string
		switch value := content.(type) {
		case mcp.TextContent:
			text = value.Text
		case *mcp.TextContent:
			text = value.Text
		default:
			encoded, err := json.Marshal(map[string]string{
				"type":   fmt.Sprintf("%T", content),
				"status": "non-text MCP content omitted",
			})
			if err != nil {
				return "", runtimeError(CategoryDownstreamFailed, 502122,
					"encode Pole MCP tool result", false, err)
			}
			text = string(encoded)
		}
		size += len(text)
		if size > maxMCPResultBytes {
			return "", runtimeError(CategoryDownstreamFailed, 502123,
				"Pole MCP tool result is too large", false, errors.New("tool result limit exceeded"))
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n"), nil
}
