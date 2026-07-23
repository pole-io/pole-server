package poleagent

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestMCPTextContentKeepsTextAndOmitsBinaryPayload(t *testing.T) {
	content, err := mcpTextContent([]mcp.Content{
		mcp.TextContent{Type: "text", Text: `{"namespaces":["default"]}`},
		mcp.ImageContent{Type: "image", Data: "must-not-reach-model", MIMEType: "image/png"},
	})
	require.NoError(t, err)
	require.Contains(t, content, "default")
	require.Contains(t, content, "non-text MCP content omitted")
	require.NotContains(t, content, "must-not-reach-model")
}

func TestNewMCPToolPortRequiresEndpointAndAllowlist(t *testing.T) {
	_, err := NewMCPToolPort(MCPToolConfig{ToolAllowlist: []string{"list_namespaces"}})
	require.Error(t, err)
	_, err = NewMCPToolPort(MCPToolConfig{Endpoint: "http://pole/ai/mcp/v1/sse"})
	require.Error(t, err)
	port, err := NewMCPToolPort(MCPToolConfig{
		Endpoint:      "http://pole/ai/mcp/v1/sse",
		ToolAllowlist: []string{" list_namespaces ", ""},
	})
	require.NoError(t, err)
	require.Contains(t, port.allowed, "list_namespaces")
}

func TestBearerTokenPreservesExistingScheme(t *testing.T) {
	require.Equal(t, "Bearer token", bearerToken("token"))
	require.Equal(t, "Bearer token", bearerToken("Bearer token"))
}
