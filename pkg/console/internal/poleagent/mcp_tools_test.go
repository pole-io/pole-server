package poleagent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/pkg/console/internal/agentworkbench"
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

func TestMCPToolPortResolvesOwnedServerFromRegistry(t *testing.T) {
	var gotAuthorization, gotUser string
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		gotUser = r.Header.Get("X-Pole-User")
		require.Equal(t, "/ai/mcp/v1/servers", r.URL.Path)
		require.Equal(t, "pole-system", r.URL.Query().Get("namespace"))
		require.Equal(t, "pole-control-plane", r.URL.Query().Get("name"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"amount":1,"size":1,"data":[{"@type":"type.googleapis.com/v1.MCPServer","name":"pole-control-plane","namespace":"pole-system","backend_type":"address","backend_address":"http://pole-mcp/ai/mcp/v1/sse"}]}`))
	}))
	defer registry.Close()

	port, err := NewMCPToolPort(MCPToolConfig{
		RegistryEndpoint: registry.URL + "/ai/mcp/v1/servers",
		ServerNamespace:  "pole-system",
		ServerName:       "pole-control-plane",
		ToolAllowlist:    []string{"list_namespaces"},
	})
	require.NoError(t, err)
	endpoint, err := port.resolveEndpoint(context.Background(), agentworkbench.Actor{
		UserID: "admin", Token: "admin-token",
	})
	require.NoError(t, err)
	require.Equal(t, "http://pole-mcp/ai/mcp/v1/sse", endpoint)
	require.Equal(t, "Bearer admin-token", gotAuthorization)
	require.Equal(t, "admin", gotUser)
}

func TestBearerTokenPreservesExistingScheme(t *testing.T) {
	require.Equal(t, "Bearer token", bearerToken("token"))
	require.Equal(t, "Bearer token", bearerToken("Bearer token"))
}
