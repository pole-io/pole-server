package selfmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	specai "github.com/pole-io/specification/source/go/api/v1/ai"
	"google.golang.org/protobuf/proto"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

const SystemActor = "pole-self-manager"

type ResourceKind string

const (
	ResourceMCPServer ResourceKind = "MCPServer"
	ResourceA2AAgent  ResourceKind = "A2AAgent"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete-stale"
)

type Store interface {
	CreateMCPServer(server *specai.MCPServer) error
	UpdateMCPServer(server *specai.MCPServer) error
	GetMCPServer(id string) (*specai.MCPServer, error)
	GetMCPServerByName(name, namespace string) (*specai.MCPServer, error)
	CreateMCPServerTool(tool *specai.MCPServerTool) error
	UpdateMCPServerTool(tool *specai.MCPServerTool) error
	DeleteMCPServerTool(id string) error
	GetMCPServerToolsByServerID(serverID string) ([]*specai.MCPServerTool, error)

	CreateA2AAgent(agent *aitypes.A2AAgent) error
	UpdateA2AAgent(agent *aitypes.A2AAgent) error
	GetA2AAgent(id string) (*aitypes.A2AAgent, error)
	GetA2AAgentByName(name, namespace string) (*aitypes.A2AAgent, error)
}

type AuditSink interface {
	Record(ctx context.Context, event AuditEvent) error
}

type AuditEvent struct {
	Actor          string       `json:"actor"`
	Trigger        string       `json:"trigger"`
	ConfiguredBy   string       `json:"configured_by,omitempty"`
	ResourceKind   ResourceKind `json:"resource_kind"`
	Namespace      string       `json:"namespace"`
	Name           string       `json:"name"`
	ResourceID     string       `json:"resource_id"`
	Action         Action       `json:"action"`
	BeforeRevision string       `json:"before_revision,omitempty"`
	AfterRevision  string       `json:"after_revision,omitempty"`
}

type DesiredState struct {
	Trigger      string
	ConfiguredBy string
	MCP          *DesiredMCP
	A2A          *aitypes.A2AAgent
}

type DesiredMCP struct {
	Server *specai.MCPServer
	Tools  []*specai.MCPServerTool
}

type Result struct {
	Created   int
	Updated   int
	Deleted   int
	Unchanged int
}

type Manager struct {
	store Store
	audit AuditSink
}

func NewManager(store Store, audit AuditSink) *Manager {
	return &Manager{store: store, audit: audit}
}

func (m *Manager) Reconcile(ctx context.Context, desired DesiredState) (Result, error) {
	if m == nil || m.store == nil {
		return Result{}, errors.New("self manager store is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var result Result
	if desired.MCP != nil {
		if err := m.reconcileMCP(ctx, desired, &result); err != nil {
			return result, err
		}
	}
	if desired.A2A != nil {
		if err := m.reconcileA2A(ctx, desired, &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (m *Manager) reconcileMCP(ctx context.Context, state DesiredState, result *Result) error {
	if state.MCP.Server == nil {
		return errors.New("desired MCP server is required")
	}
	server, tools, err := normalizeMCP(state.MCP)
	if err != nil {
		return err
	}
	existing, err := m.store.GetMCPServerByName(server.Name, server.Namespace)
	if err != nil {
		return fmt.Errorf("get MCP server %s/%s: %w", server.Namespace, server.Name, err)
	}
	if existing == nil {
		existing, err = m.store.GetMCPServer(server.Id)
		if err != nil {
			return fmt.Errorf("get MCP server %s by deterministic ID: %w", server.Id, err)
		}
	}
	if existing == nil {
		if err := m.store.CreateMCPServer(server); err != nil {
			return fmt.Errorf("create MCP server %s/%s: %w", server.Namespace, server.Name, err)
		}
		for _, tool := range tools {
			if err := m.store.CreateMCPServerTool(tool); err != nil {
				return fmt.Errorf("create MCP tool %s: %w", tool.Name, err)
			}
		}
		result.Created++
		return m.recordAudit(ctx, AuditEvent{
			Actor:         SystemActor,
			Trigger:       strings.TrimSpace(state.Trigger),
			ConfiguredBy:  strings.TrimSpace(state.ConfiguredBy),
			ResourceKind:  ResourceMCPServer,
			Namespace:     server.Namespace,
			Name:          server.Name,
			ResourceID:    server.Id,
			Action:        ActionCreate,
			AfterRevision: server.Revision,
		})
	}

	currentTools, err := m.store.GetMCPServerToolsByServerID(existing.Id)
	if err != nil {
		return fmt.Errorf("get MCP tools for %s/%s: %w", server.Namespace, server.Name, err)
	}
	server.Id = existing.Id
	currentByName := make(map[string]*specai.MCPServerTool, len(currentTools))
	for _, tool := range currentTools {
		if tool != nil {
			currentByName[strings.TrimSpace(tool.Name)] = tool
		}
	}
	for _, tool := range tools {
		tool.McpServerId = server.Id
		if current := currentByName[tool.Name]; current != nil {
			tool.Id = current.Id
		} else {
			tool.Id = stableID("mcp-tool", server.Id, tool.Name)
		}
	}
	if err := setMCPRevision(server, tools); err != nil {
		return err
	}

	changed := existing.Flag == 1 || !sameMCPServer(existing, server)
	for _, tool := range tools {
		current := currentByName[tool.Name]
		switch {
		case current == nil:
			if err := m.store.CreateMCPServerTool(tool); err != nil {
				return fmt.Errorf("create MCP tool %s: %w", tool.Name, err)
			}
			changed = true
		case !sameMCPTool(current, tool):
			if err := m.store.UpdateMCPServerTool(tool); err != nil {
				return fmt.Errorf("update MCP tool %s: %w", tool.Name, err)
			}
			changed = true
		}
		delete(currentByName, tool.Name)
	}
	for _, stale := range currentByName {
		if err := m.store.DeleteMCPServerTool(stale.Id); err != nil {
			return fmt.Errorf("delete stale MCP tool %s: %w", stale.Name, err)
		}
		result.Deleted++
		changed = true
	}
	if !changed {
		result.Unchanged++
		return nil
	}
	if err := m.store.UpdateMCPServer(server); err != nil {
		return fmt.Errorf("update MCP server %s/%s: %w", server.Namespace, server.Name, err)
	}
	result.Updated++
	return m.recordAudit(ctx, AuditEvent{
		Actor:          SystemActor,
		Trigger:        strings.TrimSpace(state.Trigger),
		ConfiguredBy:   strings.TrimSpace(state.ConfiguredBy),
		ResourceKind:   ResourceMCPServer,
		Namespace:      server.Namespace,
		Name:           server.Name,
		ResourceID:     server.Id,
		Action:         ActionUpdate,
		BeforeRevision: existing.Revision,
		AfterRevision:  server.Revision,
	})
}

func (m *Manager) recordAudit(ctx context.Context, event AuditEvent) error {
	if m.audit == nil {
		return nil
	}
	if err := m.audit.Record(ctx, event); err != nil {
		return fmt.Errorf("record self-management audit: %w", err)
	}
	return nil
}

func normalizeMCP(desired *DesiredMCP) (*specai.MCPServer, []*specai.MCPServerTool, error) {
	server := cloneMCPServer(desired.Server)
	trimMCPServer(server)
	if server.Name == "" || server.Namespace == "" {
		return nil, nil, errors.New("desired MCP server name and namespace are required")
	}
	server.Id = stableID("mcp-server", server.Namespace, server.Name)
	server.Flag, server.Ctime, server.Mtime = 0, "", ""

	tools := make([]*specai.MCPServerTool, 0, len(desired.Tools))
	toolNames := make(map[string]struct{}, len(desired.Tools))
	for _, source := range desired.Tools {
		if source == nil {
			continue
		}
		tool := cloneMCPTool(source)
		tool.Name = strings.TrimSpace(tool.Name)
		tool.Description = strings.TrimSpace(tool.Description)
		if tool.Name == "" {
			return nil, nil, errors.New("desired MCP tool name is required")
		}
		if _, duplicate := toolNames[tool.Name]; duplicate {
			return nil, nil, fmt.Errorf("duplicate desired MCP tool %s", tool.Name)
		}
		toolNames[tool.Name] = struct{}{}
		tool.McpServerId = server.Id
		tool.Id = stableID("mcp-tool", server.Id, tool.Name)
		tool.Flag, tool.Ctime, tool.Mtime = 0, "", ""
		var err error
		if tool.InputSchema, err = canonicalJSON(tool.InputSchema); err != nil {
			return nil, nil, fmt.Errorf("normalize MCP tool %s input schema: %w", tool.Name, err)
		}
		if tool.OutputSchema, err = canonicalJSON(tool.OutputSchema); err != nil {
			return nil, nil, fmt.Errorf("normalize MCP tool %s output schema: %w", tool.Name, err)
		}
		if tool.Annotations, err = canonicalJSON(tool.Annotations); err != nil {
			return nil, nil, fmt.Errorf("normalize MCP tool %s annotations: %w", tool.Name, err)
		}
		tools = append(tools, tool)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
	if err := setMCPRevision(server, tools); err != nil {
		return nil, nil, err
	}
	return server, tools, nil
}

func setMCPRevision(server *specai.MCPServer, tools []*specai.MCPServerTool) error {
	server.Revision = ""
	revisionPayload, err := json.Marshal(struct {
		Server *specai.MCPServer
		Tools  []*specai.MCPServerTool
	}{Server: server, Tools: tools})
	if err != nil {
		return fmt.Errorf("build MCP revision: %w", err)
	}
	server.Revision = digest32(revisionPayload)
	return nil
}

func trimMCPServer(server *specai.MCPServer) {
	server.Name = strings.TrimSpace(server.Name)
	server.Namespace = strings.TrimSpace(server.Namespace)
	server.Ports = strings.TrimSpace(server.Ports)
	server.Business = strings.TrimSpace(server.Business)
	server.Department = strings.TrimSpace(server.Department)
	server.Description = strings.TrimSpace(server.Description)
	server.Reference = strings.TrimSpace(server.Reference)
	server.Protocol = strings.TrimSpace(server.Protocol)
	server.ExportTo = strings.TrimSpace(server.ExportTo)
	server.BackendType = strings.TrimSpace(server.BackendType)
	server.BackendServiceNamespace = strings.TrimSpace(server.BackendServiceNamespace)
	server.BackendServiceName = strings.TrimSpace(server.BackendServiceName)
	server.BackendAddress = strings.TrimSpace(server.BackendAddress)
}

func canonicalJSON(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func stableID(parts ...string) string {
	return digest32([]byte(strings.Join(parts, "\x00")))
}

func digest32(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:16])
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

func sameMCPServer(left, right *specai.MCPServer) bool {
	leftCopy := cloneMCPServer(left)
	rightCopy := cloneMCPServer(right)
	leftCopy.Ctime, leftCopy.Mtime, leftCopy.Flag = "", "", 0
	rightCopy.Ctime, rightCopy.Mtime, rightCopy.Flag = "", "", 0
	return proto.Equal(leftCopy, rightCopy)
}

func sameMCPTool(left, right *specai.MCPServerTool) bool {
	leftCopy := cloneMCPTool(left)
	rightCopy := cloneMCPTool(right)
	leftCopy.Ctime, leftCopy.Mtime, leftCopy.Flag = "", "", 0
	rightCopy.Ctime, rightCopy.Mtime, rightCopy.Flag = "", "", 0
	return proto.Equal(leftCopy, rightCopy)
}
