package bootstrap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	specai "github.com/pole-io/specification/source/go/api/v1/ai"
	"github.com/stretchr/testify/require"

	boot_config "github.com/pole-io/pole-server/bootstrap/config"
	consolebootstrap "github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/poleagent"
	"github.com/pole-io/pole-server/pkg/selfmanager"
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
		Version: "v1", ProtocolVersion: "0.3.0", PreferredTransport: "JSONRPC",
		Skills: []poleagent.A2ASkill{{
			ID: "pole-control-plane-management", Name: "Pole 控制面管理",
			InputModes: []string{"text/plain"}, OutputModes: []string{"text/plain"},
		}},
	}
	raw, err := json.Marshal(card)
	require.NoError(t, err)
	agent := poleAgentRegistryProjection(cfg, card, string(raw))
	require.Equal(t, "Pole Agent", agent.Name)
	require.Equal(t, "pole-system", agent.Namespace)
	require.Equal(t, "JSONRPC", agent.PreferredProtocolBinding)
	require.Equal(t, card.URL, agent.PreferredInterfaceUrl)
	require.Equal(t, "pole-self-manager", agent.Metadata["managed_by"])
	require.Len(t, agent.Skills, 1)
	require.Equal(t, "pole-control-plane-management", agent.Skills[0].SkillId)
}

func TestRemoteConsoleCapabilityProbeSupportsSplitDeployment(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	probeKey, err := selfmanager.DeriveCapabilityProbeKey(masterKey)
	require.NoError(t, err)
	var received struct {
		Allowlist []string `json:"allowlist"`
	}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/ai/mcp/v1/self-capabilities/probe", r.URL.Path)
		var raw json.RawMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		require.NoError(t, selfmanager.VerifyCapabilityProbe(probeKey, r.Method, r.URL.Path, raw,
			r.Header.Get(selfmanager.ProbeTimestampHeader),
			r.Header.Get(selfmanager.ProbeSignatureHeader), time.Now()))
		require.NoError(t, json.Unmarshal(raw, &received))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	cfg := &consolebootstrap.Config{
		PoleServer:    consolebootstrap.PoleServer{Address: backend.URL},
		SystemSecrets: consolebootstrap.SystemSecretsConfig{MasterKey: masterKey},
		Agent: consolebootstrap.AgentConfig{
			SelfManagementProbeKey: probeKey,
		},
	}
	configureRemoteConsoleAgentCapabilityProbe(cfg)
	require.NotNil(t, cfg.AgentCapabilityProbe)
	require.NoError(t, cfg.AgentCapabilityProbe(context.Background(),
		[]string{"list_namespaces", "list_mcp_servers"}))
	require.Equal(t, []string{"list_namespaces", "list_mcp_servers"}, received.Allowlist)
}

func TestEnsureA2AAdvertisedEndpointUsesDiscoveredReachableHost(t *testing.T) {
	cfg := &boot_config.Config{Bootstrap: boot_config.Bootstrap{
		Console: consolebootstrap.Config{
			WebServer: consolebootstrap.WebServer{ListenPort: 8080},
		},
	}}
	ensureA2AAdvertisedEndpoint(cfg, "10.0.0.8")
	require.Equal(t, "http://10.0.0.8:8080/ai/agent/a2a/v1",
		cfg.Bootstrap.Console.Agent.A2A.Endpoint)

	cfg.Bootstrap.Console.Agent.A2A.Endpoint = "https://pole.example/a2a"
	ensureA2AAdvertisedEndpoint(cfg, "10.0.0.9")
	require.Equal(t, "https://pole.example/a2a", cfg.Bootstrap.Console.Agent.A2A.Endpoint)
}
