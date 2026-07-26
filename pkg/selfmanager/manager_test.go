package selfmanager_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	specai "github.com/pole-io/specification/source/go/api/v1/ai"
	"google.golang.org/protobuf/proto"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/pkg/selfmanager"
)

func TestManagerReconcileCreatesOwnMCPAggregate(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	audit := &memoryAuditSink{}
	manager := selfmanager.NewManager(store, audit)

	result, err := manager.Reconcile(context.Background(), selfmanager.DesiredState{
		Trigger:      "startup",
		ConfiguredBy: "admin",
		MCP: &selfmanager.DesiredMCP{
			Server: &specai.MCPServer{
				Name:           " pole-control-plane ",
				Namespace:      " default ",
				Description:    " Pole 自身 MCP ",
				Protocol:       " sse ",
				BackendType:    " address ",
				BackendAddress: " http://pole-server:8090/ai/mcp/v1/sse ",
			},
			Tools: []*specai.MCPServerTool{{
				Name:        " list_namespaces ",
				Description: " 查询命名空间 ",
				InputSchema: `{"properties":{},"type":"object"}`,
			}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Created)
	require.Equal(t, 0, result.Updated)
	require.Equal(t, 0, result.Deleted)
	require.Equal(t, 0, result.Unchanged)

	server, err := store.GetMCPServerByName("pole-control-plane", "default")
	require.NoError(t, err)
	require.Len(t, server.Id, 32)
	require.Len(t, server.Revision, 32)
	require.Equal(t, "http://pole-server:8090/ai/mcp/v1/sse", server.BackendAddress)

	tools, err := store.GetMCPServerToolsByServerID(server.Id)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	require.Len(t, tools[0].Id, 32)
	require.Equal(t, server.Id, tools[0].McpServerId)
	require.Equal(t, "list_namespaces", tools[0].Name)

	require.Equal(t, []selfmanager.AuditEvent{{
		Actor:         selfmanager.SystemActor,
		Trigger:       "startup",
		ConfiguredBy:  "admin",
		ResourceKind:  selfmanager.ResourceMCPServer,
		Namespace:     "default",
		Name:          "pole-control-plane",
		ResourceID:    server.Id,
		Action:        selfmanager.ActionCreate,
		AfterRevision: server.Revision,
	}}, audit.events)
}

func TestManagerReconcileKeepsMCPStableAndConvergesToolDrift(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	audit := &memoryAuditSink{}
	manager := selfmanager.NewManager(store, audit)
	initial := selfmanager.DesiredState{
		Trigger: "startup",
		MCP: &selfmanager.DesiredMCP{
			Server: &specai.MCPServer{
				Name:           "pole-control-plane",
				Namespace:      "default",
				Description:    "Pole MCP",
				Protocol:       "sse",
				BackendType:    "address",
				BackendAddress: "http://pole-server:8090/ai/mcp/v1/sse",
			},
			Tools: []*specai.MCPServerTool{
				{Name: "list_namespaces", InputSchema: `{"type":"object"}`},
				{Name: "stale_tool", InputSchema: `{"type":"object"}`},
			},
		},
	}
	_, err := manager.Reconcile(context.Background(), initial)
	require.NoError(t, err)
	serverBefore, err := store.GetMCPServerByName("pole-control-plane", "default")
	require.NoError(t, err)
	toolsBefore, err := store.GetMCPServerToolsByServerID(serverBefore.Id)
	require.NoError(t, err)
	toolIDs := map[string]string{}
	for _, tool := range toolsBefore {
		toolIDs[tool.Name] = tool.Id
	}

	audit.events = nil
	second, err := manager.Reconcile(context.Background(), initial)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Unchanged: 1}, second)
	require.Empty(t, audit.events)

	changed := initial
	changed.Trigger = "periodic"
	changed.MCP = &selfmanager.DesiredMCP{
		Server: cloneMCPServer(initial.MCP.Server),
		Tools: []*specai.MCPServerTool{
			{Name: "list_namespaces", Description: "新的描述", InputSchema: `{ "type": "object" }`},
			{Name: "list_mcp_servers", InputSchema: `{"type":"object"}`},
		},
	}
	changed.MCP.Server.Description = "Pole 自管理 MCP"
	third, err := manager.Reconcile(context.Background(), changed)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Updated: 1, Deleted: 1}, third)

	serverAfter, err := store.GetMCPServerByName("pole-control-plane", "default")
	require.NoError(t, err)
	require.Equal(t, serverBefore.Id, serverAfter.Id)
	require.NotEqual(t, serverBefore.Revision, serverAfter.Revision)
	toolsAfter, err := store.GetMCPServerToolsByServerID(serverAfter.Id)
	require.NoError(t, err)
	require.Len(t, toolsAfter, 2)
	afterByName := map[string]*specai.MCPServerTool{}
	for _, tool := range toolsAfter {
		afterByName[tool.Name] = tool
	}
	require.Equal(t, toolIDs["list_namespaces"], afterByName["list_namespaces"].Id)
	require.Equal(t, "新的描述", afterByName["list_namespaces"].Description)
	require.NotEmpty(t, afterByName["list_mcp_servers"].Id)
	require.NotContains(t, afterByName, "stale_tool")

	require.Len(t, audit.events, 1)
	require.Equal(t, selfmanager.ActionUpdate, audit.events[0].Action)
	require.Equal(t, serverBefore.Revision, audit.events[0].BeforeRevision)
	require.Equal(t, serverAfter.Revision, audit.events[0].AfterRevision)
}

func TestManagerReconcileRevivesSoftDeletedMCPAggregate(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	audit := &memoryAuditSink{}
	manager := selfmanager.NewManager(store, audit)
	desired := selfmanager.DesiredState{
		Trigger: "startup",
		MCP: &selfmanager.DesiredMCP{
			Server: &specai.MCPServer{
				Name:           "pole-control-plane",
				Namespace:      "default",
				Protocol:       "sse",
				BackendType:    "address",
				BackendAddress: "http://pole-server:8090/ai/mcp/v1/sse",
			},
			Tools: []*specai.MCPServerTool{{
				Name:        "list_namespaces",
				InputSchema: `{"type":"object"}`,
			}},
		},
	}
	_, err := manager.Reconcile(context.Background(), desired)
	require.NoError(t, err)
	deleted, err := store.GetMCPServerByName("pole-control-plane", "default")
	require.NoError(t, err)
	deleted.Flag = 1
	store.mcpServers[resourceKey(deleted.Namespace, deleted.Name)] = deleted
	audit.events = nil

	result, err := manager.Reconcile(context.Background(), desired)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Updated: 1}, result)

	revived, err := store.GetMCPServer(deleted.Id)
	require.NoError(t, err)
	require.Equal(t, deleted.Id, revived.Id)
	require.Zero(t, revived.Flag)
	require.Len(t, audit.events, 1)
	require.Equal(t, selfmanager.ActionUpdate, audit.events[0].Action)
}

func TestManagerReconcileConvergesLegacyMCPRootOnFirstRevive(t *testing.T) {
	store := newMemoryStore()
	store.preserveMCPRootIDOnCreate = true
	store.mcpServers["pole-system/pole-control-plane"] = &specai.MCPServer{
		Id: "legacy-root-id", Name: "pole-control-plane", Namespace: "pole-system",
		Protocol: "legacy", BackendType: "address", BackendAddress: "http://legacy.invalid",
		Revision: "legacy-revision", Flag: 1,
	}
	manager := selfmanager.NewManager(store, nil)
	result, err := manager.Reconcile(context.Background(), selfmanager.DesiredState{
		MCP: &selfmanager.DesiredMCP{
			Server: &specai.MCPServer{
				Name: "pole-control-plane", Namespace: "pole-system",
				Protocol: "sse", BackendType: "address",
				BackendAddress: "http://pole-server:8090/ai/mcp/v1/sse",
			},
			Tools: []*specai.MCPServerTool{{Name: "list_namespaces", InputSchema: `{}`}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Created: 1}, result)
	actual, err := store.GetMCPServerByName("pole-control-plane", "pole-system")
	require.NoError(t, err)
	require.Equal(t, "legacy-root-id", actual.Id)
	require.Equal(t, "sse", actual.Protocol)
	require.Equal(t, "http://pole-server:8090/ai/mcp/v1/sse", actual.BackendAddress)
	require.NotEqual(t, "legacy-revision", actual.Revision)
	tools, err := store.GetMCPServerToolsByServerID("legacy-root-id")
	require.NoError(t, err)
	require.Len(t, tools, 1)
	require.Equal(t, "legacy-root-id", tools[0].McpServerId)
}

func TestManagerReconcileCreatesAndUpdatesA2AAgentIdempotently(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	audit := &memoryAuditSink{}
	manager := selfmanager.NewManager(store, audit)
	initial := selfmanager.DesiredState{
		Trigger:      "startup",
		ConfiguredBy: "admin",
		A2A: &aitypes.A2AAgent{
			Name:                     " pole-agent ",
			Namespace:                " default ",
			Description:              " Pole Agent ",
			Version:                  " v1 ",
			ProtocolVersion:          " 0.3.0 ",
			BackendType:              " address ",
			BackendAddress:           " http://pole-console:8080/ai/agent/v1 ",
			PreferredInterfaceUrl:    " http://pole-console:8080/ai/a2a ",
			PreferredProtocolBinding: " JSONRPC ",
			RawCardJson:              `{ "name": "pole-agent", "version": "v1" }`,
			Metadata:                 map[string]string{" owner ": " platform "},
			Interfaces: []*aitypes.A2AAgentInterface{
				{Url: " http://pole-console:8080/ai/a2a ", ProtocolBinding: " JSONRPC ", ProtocolVersion: " 0.3.0 "},
			},
			Skills: []*aitypes.A2AAgentSkill{
				{SkillId: " manage ", Name: " 管理 Pole ", Tags: []string{" admin ", "pole"}},
				{SkillId: " inspect ", Name: " 查询 Pole ", Tags: []string{"read"}},
			},
		},
	}

	first, err := manager.Reconcile(context.Background(), initial)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Created: 1}, first)
	agentBefore, err := store.GetA2AAgentByName("pole-agent", "default")
	require.NoError(t, err)
	require.Len(t, agentBefore.Id, 32)
	require.Equal(t, `{"name":"pole-agent","version":"v1"}`, agentBefore.RawCardJson)
	require.Equal(t, map[string]string{"owner": "platform"}, agentBefore.Metadata)
	require.Len(t, agentBefore.Interfaces[0].Id, 32)
	require.Len(t, agentBefore.Skills[0].Id, 32)

	audit.events = nil
	second, err := manager.Reconcile(context.Background(), initial)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Unchanged: 1}, second)
	require.Empty(t, audit.events)

	changed := initial
	changed.Trigger = "card-change"
	changed.A2A = cloneA2AAgent(initial.A2A)
	changed.A2A.Description = "Pole 自管理 Agent"
	changed.A2A.Interfaces = nil
	changed.A2A.Skills = []*aitypes.A2AAgentSkill{
		{SkillId: "manage", Name: "管理 Pole", Tags: []string{"pole", "admin"}},
		{SkillId: "diagnose", Name: "诊断 Pole", Tags: []string{"read"}},
	}
	third, err := manager.Reconcile(context.Background(), changed)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Updated: 1, Deleted: 2}, third)

	agentAfter, err := store.GetA2AAgentByName("pole-agent", "default")
	require.NoError(t, err)
	require.Equal(t, agentBefore.Id, agentAfter.Id)
	require.Empty(t, agentAfter.Interfaces)
	require.Len(t, agentAfter.Skills, 2)
	beforeSkillIDs := map[string]string{}
	for _, skill := range agentBefore.Skills {
		beforeSkillIDs[skill.SkillId] = skill.Id
	}
	afterSkills := map[string]*aitypes.A2AAgentSkill{}
	for _, skill := range agentAfter.Skills {
		afterSkills[skill.SkillId] = skill
	}
	require.Equal(t, beforeSkillIDs["manage"], afterSkills["manage"].Id)
	require.NotEmpty(t, afterSkills["diagnose"].Id)
	require.NotContains(t, afterSkills, "inspect")

	require.Len(t, audit.events, 1)
	require.Equal(t, selfmanager.ResourceA2AAgent, audit.events[0].ResourceKind)
	require.Equal(t, selfmanager.ActionUpdate, audit.events[0].Action)
	require.NotEmpty(t, audit.events[0].BeforeRevision)
	require.NotEmpty(t, audit.events[0].AfterRevision)
	require.NotEqual(t, audit.events[0].BeforeRevision, audit.events[0].AfterRevision)
}

func TestManagerReconcileDeletesStaleSelfManagedA2ARootAfterRename(t *testing.T) {
	store := newMemoryStore()
	require.NoError(t, store.CreateA2AAgent(&aitypes.A2AAgent{
		Id: "old-agent", Name: "old-agent", Namespace: "pole-system",
		RawCardJson: `{}`, Metadata: map[string]string{"managed_by": selfmanager.SystemActor},
	}))
	manager := selfmanager.NewManager(store, nil)
	result, err := manager.Reconcile(context.Background(), selfmanager.DesiredState{
		Trigger: "periodic",
		A2A: &aitypes.A2AAgent{
			Name: "new-agent", Namespace: "pole-system", RawCardJson: `{}`,
			Metadata: map[string]string{"managed_by": selfmanager.SystemActor},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Created)
	require.Equal(t, 1, result.Deleted)
	old, err := store.GetA2AAgent("old-agent")
	require.NoError(t, err)
	require.Equal(t, uint32(1), old.Flag)
}

func TestManagerReconcileFailsClosedUntilCanonicalRootsAreObservable(t *testing.T) {
	t.Run("MCP", func(t *testing.T) {
		store := newMemoryStore()
		store.hideMCPByName = true
		manager := selfmanager.NewManager(store, nil)
		result, err := manager.Reconcile(context.Background(), selfmanager.DesiredState{
			MCP: &selfmanager.DesiredMCP{
				Server: &specai.MCPServer{
					Name: "pole-control-plane", Namespace: "pole-system",
					Protocol: "sse", BackendType: "address",
					BackendAddress: "http://pole-server:8090/ai/mcp/v1/sse",
				},
				Tools: []*specai.MCPServerTool{{Name: "list_namespaces", InputSchema: `{}`}},
			},
		})
		require.ErrorContains(t, err, "is not observable after create")
		require.Equal(t, selfmanager.Result{}, result)
		require.Empty(t, store.mcpTools)
	})

	t.Run("A2A", func(t *testing.T) {
		store := newMemoryStore()
		require.NoError(t, store.CreateA2AAgent(&aitypes.A2AAgent{
			Id: "old-agent", Name: "old-agent", Namespace: "pole-system",
			Metadata: map[string]string{"managed_by": selfmanager.SystemActor},
		}))
		store.hideA2AByName = true
		manager := selfmanager.NewManager(store, nil)
		result, err := manager.Reconcile(context.Background(), selfmanager.DesiredState{
			A2A: &aitypes.A2AAgent{
				Name: "new-agent", Namespace: "pole-system", RawCardJson: `{}`,
				Metadata: map[string]string{"managed_by": selfmanager.SystemActor},
			},
		})
		require.ErrorContains(t, err, "is not observable after create")
		require.Equal(t, selfmanager.Result{}, result)
		old, getErr := store.GetA2AAgent("old-agent")
		require.NoError(t, getErr)
		require.Zero(t, old.Flag)
	})
}

func TestManagerReconcilePreservesLegacyA2AIDs(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	existing := &aitypes.A2AAgent{
		Id:        "legacy-agent-id",
		Name:      "pole-agent",
		Namespace: "default",
		Interfaces: []*aitypes.A2AAgentInterface{{
			Id:              "legacy-interface-id",
			AgentId:         "legacy-agent-id",
			Url:             "https://pole.example/a2a",
			ProtocolBinding: "JSONRPC",
		}},
		Skills: []*aitypes.A2AAgentSkill{{
			Id:      "legacy-skill-id",
			AgentId: "legacy-agent-id",
			SkillId: "manage",
			Name:    "管理 Pole",
		}},
	}
	require.NoError(t, store.CreateA2AAgent(existing))
	manager := selfmanager.NewManager(store, nil)

	result, err := manager.Reconcile(context.Background(), selfmanager.DesiredState{A2A: cloneA2AAgent(existing)})
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Unchanged: 1}, result)

	actual, err := store.GetA2AAgentByName("pole-agent", "default")
	require.NoError(t, err)
	require.Equal(t, "legacy-agent-id", actual.Id)
	require.Equal(t, "legacy-interface-id", actual.Interfaces[0].Id)
	require.Equal(t, "legacy-skill-id", actual.Skills[0].Id)
}

func TestManagerReconcileRevivesSoftDeletedA2AAgent(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	audit := &memoryAuditSink{}
	manager := selfmanager.NewManager(store, audit)
	desired := selfmanager.DesiredState{
		Trigger: "startup",
		A2A: &aitypes.A2AAgent{
			Name:      "pole-agent",
			Namespace: "default",
			Skills: []*aitypes.A2AAgentSkill{{
				SkillId: "manage",
				Name:    "管理 Pole",
			}},
		},
	}
	_, err := manager.Reconcile(context.Background(), desired)
	require.NoError(t, err)
	deleted, err := store.GetA2AAgentByName("pole-agent", "default")
	require.NoError(t, err)
	deleted.Flag = 1
	store.a2aAgents[resourceKey(deleted.Namespace, deleted.Name)] = deleted
	audit.events = nil

	result, err := manager.Reconcile(context.Background(), desired)
	require.NoError(t, err)
	require.Equal(t, selfmanager.Result{Updated: 1}, result)

	revived, err := store.GetA2AAgent(deleted.Id)
	require.NoError(t, err)
	require.Equal(t, deleted.Id, revived.Id)
	require.Zero(t, revived.Flag)
	require.Len(t, audit.events, 1)
	require.Equal(t, selfmanager.ActionUpdate, audit.events[0].Action)
}

type memoryStore struct {
	mcpServers                map[string]*specai.MCPServer
	mcpTools                  map[string]map[string]*specai.MCPServerTool
	a2aAgents                 map[string]*aitypes.A2AAgent
	hideMCPByName             bool
	hideA2AByName             bool
	preserveMCPRootIDOnCreate bool
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		mcpServers: map[string]*specai.MCPServer{},
		mcpTools:   map[string]map[string]*specai.MCPServerTool{},
		a2aAgents:  map[string]*aitypes.A2AAgent{},
	}
}

func (s *memoryStore) CreateMCPServer(server *specai.MCPServer) error {
	if existing := s.mcpServers[resourceKey(server.Namespace, server.Name)]; s.preserveMCPRootIDOnCreate && existing != nil {
		existing.Flag = 0
		return nil
	}
	s.mcpServers[resourceKey(server.Namespace, server.Name)] = cloneMCPServer(server)
	return nil
}

func (s *memoryStore) UpdateMCPServer(server *specai.MCPServer) error {
	s.mcpServers[resourceKey(server.Namespace, server.Name)] = cloneMCPServer(server)
	return nil
}

func (s *memoryStore) GetMCPServerByName(name, namespace string) (*specai.MCPServer, error) {
	if s.hideMCPByName {
		return nil, nil
	}
	server := s.mcpServers[resourceKey(namespace, name)]
	if server != nil && server.Flag == 1 {
		return nil, nil
	}
	return cloneMCPServer(server), nil
}

func (s *memoryStore) GetMCPServer(id string) (*specai.MCPServer, error) {
	for _, server := range s.mcpServers {
		if server.Id == id {
			return cloneMCPServer(server), nil
		}
	}
	return nil, nil
}

func (s *memoryStore) CreateMCPServerTool(tool *specai.MCPServerTool) error {
	if s.mcpTools[tool.McpServerId] == nil {
		s.mcpTools[tool.McpServerId] = map[string]*specai.MCPServerTool{}
	}
	s.mcpTools[tool.McpServerId][tool.Name] = cloneMCPTool(tool)
	return nil
}

func (s *memoryStore) UpdateMCPServerTool(tool *specai.MCPServerTool) error {
	return s.CreateMCPServerTool(tool)
}

func (s *memoryStore) DeleteMCPServerTool(id string) error {
	for _, tools := range s.mcpTools {
		for name, tool := range tools {
			if tool.Id == id {
				delete(tools, name)
			}
		}
	}
	return nil
}

func (s *memoryStore) GetMCPServerToolsByServerID(serverID string) ([]*specai.MCPServerTool, error) {
	tools := make([]*specai.MCPServerTool, 0, len(s.mcpTools[serverID]))
	for _, tool := range s.mcpTools[serverID] {
		tools = append(tools, cloneMCPTool(tool))
	}
	return tools, nil
}

func (s *memoryStore) CreateA2AAgent(agent *aitypes.A2AAgent) error {
	s.a2aAgents[resourceKey(agent.Namespace, agent.Name)] = cloneA2AAgent(agent)
	return nil
}

func (s *memoryStore) UpdateA2AAgent(agent *aitypes.A2AAgent) error {
	return s.CreateA2AAgent(agent)
}

func (s *memoryStore) GetA2AAgentByName(name, namespace string) (*aitypes.A2AAgent, error) {
	if s.hideA2AByName {
		return nil, nil
	}
	agent := s.a2aAgents[resourceKey(namespace, name)]
	if agent != nil && agent.Flag == 1 {
		return nil, nil
	}
	return cloneA2AAgent(agent), nil
}

func (s *memoryStore) GetA2AAgent(id string) (*aitypes.A2AAgent, error) {
	for _, agent := range s.a2aAgents {
		if agent.Id == id {
			return cloneA2AAgent(agent), nil
		}
	}
	return nil, nil
}

func (s *memoryStore) DeleteA2AAgent(id string) error {
	for _, agent := range s.a2aAgents {
		if agent.Id == id {
			agent.Flag = 1
		}
	}
	return nil
}

func (s *memoryStore) QueryA2AAgents(query *aitypes.A2AAgentQuery) (uint32, []*aitypes.A2AAgent, error) {
	agents := make([]*aitypes.A2AAgent, 0, len(s.a2aAgents))
	for _, agent := range s.a2aAgents {
		if agent.Flag == 1 || query != nil && query.Namespace != "" && agent.Namespace != query.Namespace {
			continue
		}
		agents = append(agents, cloneA2AAgent(agent))
	}
	return uint32(len(agents)), agents, nil
}

type memoryAuditSink struct {
	events []selfmanager.AuditEvent
}

func (s *memoryAuditSink) Record(_ context.Context, event selfmanager.AuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

func resourceKey(namespace, name string) string {
	return namespace + "/" + name
}

func cloneMCPServer(value *specai.MCPServer) *specai.MCPServer {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*specai.MCPServer)
}

func cloneMCPTool(value *specai.MCPServerTool) *specai.MCPServerTool {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*specai.MCPServerTool)
}

func cloneA2AAgent(value *aitypes.A2AAgent) *aitypes.A2AAgent {
	if value == nil {
		return nil
	}
	copyValue := *value
	copyValue.Metadata = make(map[string]string, len(value.Metadata))
	for key, item := range value.Metadata {
		copyValue.Metadata[key] = item
	}
	copyValue.Interfaces = make([]*aitypes.A2AAgentInterface, 0, len(value.Interfaces))
	for _, item := range value.Interfaces {
		itemCopy := *item
		copyValue.Interfaces = append(copyValue.Interfaces, &itemCopy)
	}
	copyValue.Skills = make([]*aitypes.A2AAgentSkill, 0, len(value.Skills))
	for _, item := range value.Skills {
		itemCopy := *item
		itemCopy.Tags = append([]string(nil), item.Tags...)
		itemCopy.Examples = append([]string(nil), item.Examples...)
		itemCopy.InputModes = append([]string(nil), item.InputModes...)
		itemCopy.OutputModes = append([]string(nil), item.OutputModes...)
		copyValue.Skills = append(copyValue.Skills, &itemCopy)
	}
	return &copyValue
}
