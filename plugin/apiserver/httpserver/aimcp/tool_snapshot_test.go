package aimcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

func TestSnapshotToolsMatchesPoleMCPToolCatalog(t *testing.T) {
	httpServer := &HTTPServer{mcpSvr: server.NewMCPServer("pole-test", "v1")}
	httpServer.addMcpTools()

	tools, err := httpServer.SnapshotTools(context.Background())
	require.NoError(t, err)
	require.Len(t, tools, 11)
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.GetName()] = true
		require.True(t, json.Valid([]byte(tool.GetInputSchema())), tool.GetName())
	}
	for _, name := range []string{
		"list_namespaces", "create_namespaces", "update_namespaces", "delete_namespaces",
		"list_mcp_servers", "create_mcp_servers", "update_mcp_servers", "delete_mcp_servers",
		"list_mcp_server_tools", "get_config_file", "search_config_files",
	} {
		require.True(t, names[name], name)
	}
}
