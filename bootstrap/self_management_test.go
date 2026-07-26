package bootstrap

import (
	"encoding/json"
	"testing"

	specai "github.com/pole-io/specification/source/go/api/v1/ai"
	"github.com/stretchr/testify/require"

	boot_config "github.com/pole-io/pole-server/bootstrap/config"
	consolebootstrap "github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/poleagent"
)

func TestSelfManagementDesiredOwnsPoleMCPAndA2AProjection(t *testing.T) {
	cfg := &boot_config.Config{Bootstrap: boot_config.Bootstrap{
		Console: consolebootstrap.Config{Agent: consolebootstrap.AgentConfig{
			Definition: consolebootstrap.AgentDefinitionConfig{ID: "pole-control-plane"},
			MCP: consolebootstrap.AgentMCPConfig{
				Endpoint: "http://pole-server:8090/ai/mcp/v1/sse",
			},
		}},
	}}
	tools := []*specai.MCPServerTool{{Name: "list_namespaces", InputSchema: `{"type":"object"}`}}
	desired := selfManagementDesired(cfg, tools, nil, "startup")
	require.Equal(t, "pole-system", desired.MCP.Server.GetNamespace())
	require.Equal(t, "pole-control-plane", desired.MCP.Server.GetName())
	require.Equal(t, "address", desired.MCP.Server.GetBackendType())
	require.Equal(t, "http://pole-server:8090/ai/mcp/v1/sse", desired.MCP.Server.GetBackendAddress())
	require.Equal(t, "list_namespaces", desired.MCP.Tools[0].GetName())

	card := &poleagent.A2AAgentCard{
		Name: "Pole Agent", Description: "Pole 管理", URL: "http://pole-console:8080/ai/agent/a2a/v1",
		Version: "v1", ProtocolVersion: "0.3.0",
		Skills: []poleagent.A2ASkill{{
			ID: "pole-control-plane-management", Name: "Pole 控制面管理",
			InputModes: []string{"text"}, OutputModes: []string{"text"},
		}},
	}
	raw, err := json.Marshal(card)
	require.NoError(t, err)
	agent := poleAgentRegistryProjection(cfg, card, string(raw))
	require.Equal(t, "pole-control-plane", agent.Name)
	require.Equal(t, "pole-system", agent.Namespace)
	require.Equal(t, "jsonrpc", agent.PreferredProtocolBinding)
	require.Equal(t, card.URL, agent.PreferredInterfaceUrl)
	require.Equal(t, "pole-self-manager", agent.Metadata["managed_by"])
	require.Len(t, agent.Skills, 1)
	require.Equal(t, "pole-control-plane-management", agent.Skills[0].SkillId)
}
