package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	specai "github.com/pole-io/specification/source/go/api/v1/ai"

	"github.com/pole-io/pole-server/apis/apiserver"
	"github.com/pole-io/pole-server/apis/observability/history"
	"github.com/pole-io/pole-server/apis/pkg/types"
	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	storeapi "github.com/pole-io/pole-server/apis/store"
	boot_config "github.com/pole-io/pole-server/bootstrap/config"
	"github.com/pole-io/pole-server/console/pkg/poleagent"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/selfmanager"
)

const selfManagementInterval = 30 * time.Second

type mcpToolSnapshotter interface {
	SnapshotMCPTools(context.Context) ([]*specai.MCPServerTool, error)
}

type selfManagementAudit struct{}

func (selfManagementAudit) Record(_ context.Context, event selfmanager.AuditEvent) error {
	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}
	resourceType := types.RMCPServer
	if event.ResourceKind == selfmanager.ResourceA2AAgent {
		resourceType = types.RA2AAgent
	}
	history.GetHistory().Record(&types.RecordEntry{
		ResourceType:  resourceType,
		ResourceName:  event.Name,
		Namespace:     event.Namespace,
		Operator:      selfmanager.SystemActor,
		OperationType: selfManagementOperation(event.Action),
		Detail:        string(encoded),
		Server:        "pole-control-plane",
		HappenTime:    time.Now().UTC(),
	})
	return nil
}

func selfManagementOperation(action selfmanager.Action) types.OperationType {
	switch action {
	case selfmanager.ActionCreate:
		return types.OCreate
	case selfmanager.ActionDelete:
		return types.ODelete
	default:
		return types.OUpdate
	}
}

// StartSelfManagement starts the deterministic controller that owns only
// Pole's own MCP and A2A registry projections.
func StartSelfManagement(ctx context.Context, cfg *boot_config.Config, storage storeapi.Store,
	servers []apiserver.Apiserver) error {
	snapshotter := findMCPSnapshotter(servers)
	if snapshotter == nil {
		return fmt.Errorf("pole-self-manager requires the enabled HTTP MCP server")
	}
	manager := selfmanager.NewManager(storage, selfManagementAudit{})
	reconcile := func(trigger string) error {
		tools, err := snapshotter.SnapshotMCPTools(ctx)
		if err != nil {
			return err
		}
		desired := selfManagementDesired(cfg, tools, nil, trigger)
		if cfg.Bootstrap.Mode == boot_config.StartModeAll {
			card, rawCard, cardErr := fetchPoleAgentCard(ctx, cfg)
			if cardErr != nil {
				commonlog.Warnf("[SelfManager] Pole Agent card is not ready: %v", cardErr)
			} else {
				desired.A2A = poleAgentRegistryProjection(cfg, card, rawCard)
			}
		}
		result, err := manager.Reconcile(ctx, desired)
		if err != nil {
			return err
		}
		if result.Created+result.Updated+result.Deleted > 0 {
			commonlog.Infof("[SelfManager] reconciled own capabilities: created=%d updated=%d deleted=%d",
				result.Created, result.Updated, result.Deleted)
		}
		return nil
	}
	if err := reconcile("startup"); err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(selfManagementInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := reconcile("periodic"); err != nil {
					commonlog.Warnf("[SelfManager] reconcile failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func findMCPSnapshotter(servers []apiserver.Apiserver) mcpToolSnapshotter {
	for _, server := range servers {
		if provider, ok := server.(mcpToolSnapshotter); ok {
			return provider
		}
	}
	return nil
}

func selfManagementDesired(cfg *boot_config.Config, tools []*specai.MCPServerTool,
	agent *aitypes.A2AAgent, trigger string) selfmanager.DesiredState {
	agentConfig := cfg.Bootstrap.Console.Agent.Normalize()
	return selfmanager.DesiredState{
		Trigger:      trigger,
		ConfiguredBy: "bootstrap",
		MCP: &selfmanager.DesiredMCP{
			Server: &specai.MCPServer{
				Name:           "pole-control-plane",
				Namespace:      "pole-system",
				Description:    "Pole Control Plane 自身管理 MCP",
				Business:       "pole-system",
				Department:     "self-management",
				Reference:      selfmanager.SystemActor,
				Protocol:       "sse",
				BackendType:    "address",
				BackendAddress: agentConfig.MCP.Endpoint,
			},
			Tools: tools,
		},
		A2A: agent,
	}
}

func fetchPoleAgentCard(ctx context.Context,
	cfg *boot_config.Config) (*poleagent.A2AAgentCard, string, error) {
	agentConfig := cfg.Bootstrap.Console.Agent.Normalize()
	endpoint := strings.TrimSpace(agentConfig.A2A.Endpoint)
	if endpoint == "" {
		listenPort := cfg.Bootstrap.Console.WebServer.ListenPort
		if listenPort <= 0 {
			listenPort = 8080
		}
		endpoint = fmt.Sprintf("http://127.0.0.1:%d/ai/agent/a2a/v1",
			listenPort)
	}
	cardURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", err
	}
	cardURL.Path, cardURL.RawQuery, cardURL.Fragment = "/.well-known/agent-card.json", "", ""
	var lastErr error
	client := &http.Client{Timeout: 2 * time.Second}
	for attempt := 0; attempt < 10; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, cardURL.String(), nil)
		if err != nil {
			return nil, "", err
		}
		response, err := client.Do(request)
		if err == nil && response.StatusCode == http.StatusOK {
			defer response.Body.Close()
			var card poleagent.A2AAgentCard
			if decodeErr := json.NewDecoder(response.Body).Decode(&card); decodeErr != nil {
				return nil, "", decodeErr
			}
			raw, encodeErr := json.Marshal(card)
			return &card, string(raw), encodeErr
		}
		if response != nil {
			_ = response.Body.Close()
			lastErr = fmt.Errorf("agent card returned HTTP %d", response.StatusCode)
		} else {
			lastErr = err
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return nil, "", ctx.Err()
		}
	}
	return nil, "", lastErr
}

func poleAgentRegistryProjection(cfg *boot_config.Config, card *poleagent.A2AAgentCard,
	rawCard string) *aitypes.A2AAgent {
	agentConfig := cfg.Bootstrap.Console.Agent.Normalize()
	agent := &aitypes.A2AAgent{
		Name:                     agentConfig.Definition.ID,
		Namespace:                "pole-system",
		Visibility:               "internal",
		Description:              card.Description,
		Version:                  card.Version,
		ProtocolVersion:          card.ProtocolVersion,
		ProviderOrganization:     "pole.io",
		BackendType:              "address",
		BackendAddress:           card.URL,
		PreferredInterfaceUrl:    card.URL,
		PreferredProtocolBinding: "jsonrpc",
		PreferredProtocolVersion: card.ProtocolVersion,
		RawCardJson:              rawCard,
		SourceType:               "well-known",
		SourceUrl:                wellKnownURL(card.URL),
		LastFetchStatus:          "success",
		Metadata: map[string]string{
			"managed_by": selfmanager.SystemActor,
			"mcp_server": "pole-system/pole-control-plane",
		},
		Interfaces: []*aitypes.A2AAgentInterface{{
			Url: card.URL, ProtocolBinding: "jsonrpc", ProtocolVersion: card.ProtocolVersion,
		}},
	}
	for _, skill := range card.Skills {
		agent.Skills = append(agent.Skills, &aitypes.A2AAgentSkill{
			SkillId: skill.ID, Name: skill.Name, Description: skill.Description,
			Tags: skill.Tags, Examples: skill.Examples,
			InputModes: skill.InputModes, OutputModes: skill.OutputModes,
			SecurityRequirementsJson: `{"bearerAuth":[]}`,
		})
	}
	return agent
}

func wellKnownURL(endpoint string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	parsed.Path, parsed.RawQuery, parsed.Fragment = "/.well-known/agent-card.json", "", ""
	return parsed.String()
}
