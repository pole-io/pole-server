package systemsettings

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pole-io/pole-server/pkg/systemconfig"
)

func TestNewSystemSettingsProviderNormalizesObservabilityAndRedactsSecrets(t *testing.T) {
	cfg := &Config{
		WebServer: WebServer{
			Mode: "release",
			JWT:  JWT{SecretKey: "console-private-secret", Expired: 1800},
		},
		Agent: AgentConfig{
			Model: AgentModelConfig{
				BaseURL: "https://llm-gateway.example.test",
				APIKey:  "agent-private-secret",
			},
		},
	}
	cfg.SystemConfigSources = systemconfig.SourceIndex{
		"webServer.mode":             {Kind: systemconfig.SourceStaticFile},
		"webServer.jwt.secretKey":    {Kind: systemconfig.SourceEnvironment, Reference: "env:CONSOLE_JWT_SECRET"},
		"observabilityQuery.timeout": {Kind: systemconfig.SourceStaticFile},
		"agent.model.baseURL":        {Kind: systemconfig.SourceEnvironment, Reference: "env:POLE_AGENT_LLM_BASE_URL"},
		"agent.model.apiKey":         {Kind: systemconfig.SourceEnvironment, Reference: "env:POLE_AGENT_LLM_API_KEY"},
	}

	provider, err := NewSystemSettingsProvider(cfg)
	if err != nil {
		t.Fatalf("NewSystemSettingsProvider() error = %v", err)
	}
	snapshot, err := provider.Effective(context.Background(), systemconfig.Scope{Component: systemconfig.ComponentConsole})
	if err != nil {
		t.Fatalf("Effective() error = %v", err)
	}
	settings := make(map[string]systemconfig.EffectiveSetting)
	for _, setting := range snapshot.Settings {
		settings[setting.Key] = setting
	}
	if got := settings["console.observability.timeout"].Value; got != "10s" {
		t.Fatalf("observability timeout = %#v, want 10s", got)
	}
	if got := settings["console.observability.timeout"].Source.Kind; got != systemconfig.SourceCompiledDefault {
		t.Fatalf("normalized observability source = %q", got)
	}
	if got := settings["console.agent.proposal_ttl"].Value; got != "30m0s" {
		t.Fatalf("agent proposal ttl = %#v, want 30m0s", got)
	}
	if got := settings["console.agent.runtime_mode"].Value; got != "llm" {
		t.Fatalf("agent runtime mode = %#v, want llm", got)
	}
	if got := settings["console.agent.model_base_url"].Value; got != "https://llm-gateway.example.test" {
		t.Fatalf("agent model base URL = %#v", got)
	}
	if got := settings["console.agent.mcp_endpoint"].Value; got != "http://127.0.0.1:8090/ai/mcp/v1/sse" {
		t.Fatalf("agent MCP endpoint = %#v", got)
	}
	secret := settings["console.web.jwt.secret_key"]
	if !secret.Redacted || secret.Value != nil || secret.Source.Reference != "env:CONSOLE_JWT_SECRET" {
		t.Fatalf("jwt secret setting = %#v", secret)
	}
	body, _ := json.Marshal(snapshot)
	if strings.Contains(string(body), "console-private-secret") {
		t.Fatalf("snapshot leaked JWT secret: %s", body)
	}
	agentSecret := settings["console.agent.model_api_key"]
	if !agentSecret.Redacted || agentSecret.Value != nil || agentSecret.Source.Reference != "env:POLE_AGENT_LLM_API_KEY" {
		t.Fatalf("agent API key setting = %#v", agentSecret)
	}
	if strings.Contains(string(body), "agent-private-secret") {
		t.Fatalf("snapshot leaked Agent API key: %s", body)
	}
}

func TestConsoleSystemSettingsHaveExplicitFieldEditPolicies(t *testing.T) {
	provider, err := NewSystemSettingsProvider(&Config{})
	if err != nil {
		t.Fatalf("NewSystemSettingsProvider() error = %v", err)
	}
	definitions, err := provider.Describe(context.Background(), systemconfig.Scope{Component: systemconfig.ComponentConsole})
	if err != nil {
		t.Fatalf("Describe() error = %v", err)
	}
	if len(definitions) != 33 {
		t.Fatalf("definition count = %d, want 33", len(definitions))
	}
	editable := 0
	for _, definition := range definitions {
		if definition.Description == "" || definition.EditReason == "" {
			t.Fatalf("setting %s has no explicit edit policy: %#v", definition.Key, definition)
		}
		if definition.Editable {
			editable++
		}
	}
	if editable != 22 {
		t.Fatalf("editable count = %d, want 22", editable)
	}
}
