/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package store

import (
	"time"

	"github.com/pole-io/specification/source/go/api/v1/ai"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// AIStore AI module storage interface
type AIStore interface {
	MCPServerStore
	A2AAgentStore
	AIResourceDefinitionStore
}

// AIResourceDefinitionStore 保存控制面逻辑定义，以及它们与现有 MCP Server
// 或 A2A Agent 环境记录之间的显式关联。
type AIResourceDefinitionStore interface {
	CreateAIResourceDefinition(definition *aitypes.ResourceDefinition) error
	UpdateAIResourceDefinition(definition *aitypes.ResourceDefinition, previousRevision string) error
	DeleteAIResourceDefinition(kind aitypes.ResourceKind, id string) error
	GetAIResourceDefinition(kind aitypes.ResourceKind, id string) (*aitypes.ResourceDefinition, error)
	ListAIResourceDefinitions(kind aitypes.ResourceKind, name string, offset, limit uint32) (
		uint32, []*aitypes.ResourceDefinition, error)
	BindAIResourceEnvironment(binding *aitypes.EnvironmentBinding, revision string) error
	UnbindAIResourceEnvironment(
		kind aitypes.ResourceKind, definitionID, resourceID, revision string) error
	ListAIResourceEnvironmentBindings(
		kind aitypes.ResourceKind, definitionID string) ([]*aitypes.EnvironmentBinding, error)
	GetAIResourceEnvironmentBinding(
		kind aitypes.ResourceKind, resourceID string) (*aitypes.EnvironmentBinding, error)
	CountAIResourcesByNamespace(namespace string) (uint32, error)
	CountAIBackendReferences(serviceID string) (uint32, error)
}

// ===== MCP Server Store =====

type MCPServerStore interface {
	// CreateMCPServer 创建 MCP Server
	CreateMCPServer(server *ai.MCPServer) error
	// UpdateMCPServer 更新 MCP Server
	UpdateMCPServer(server *ai.MCPServer) error
	// DeleteMCPServer 删除 MCP Server
	DeleteMCPServer(id string) error
	// GetMCPServer 获取 MCP Server
	GetMCPServer(id string) (*ai.MCPServer, error)
	// GetMCPServerByName 获取 MCP Server by name and namespace
	GetMCPServerByName(name, namespace string) (*ai.MCPServer, error)
	// GetMoreMCPServers 增量获取 MCP Servers
	GetMoreMCPServers(mtime time.Time, firstUpdate bool) ([]*ai.MCPServer, error)
	// HasMCPServer 检查 MCP Server 是否存在
	HasMCPServer(id string) (bool, error)
	// HasMCPServerByName 检查 MCP Server 是否存在 by name
	HasMCPServerByName(name, namespace string) (bool, error)
	// HasMCPServerByNameExcludeId 检查 MCP Server 是否存在 by name exclude id
	HasMCPServerByNameExcludeId(name, namespace, id string) (bool, error)

	// MCP Server Tool 相关方法
	// CreateMCPServerTool 创建 MCP Server Tool
	CreateMCPServerTool(tool *ai.MCPServerTool) error
	// UpdateMCPServerTool 更新 MCP Server Tool
	UpdateMCPServerTool(tool *ai.MCPServerTool) error
	// DeleteMCPServerTool 删除 MCP Server Tool
	DeleteMCPServerTool(id string) error
	// GetMCPServerTool 获取 MCP Server Tool
	GetMCPServerTool(id string) (*ai.MCPServerTool, error)
	// GetMCPServerToolsByServerID 获取 MCP Server 的所有 Tools
	GetMCPServerToolsByServerID(serverID string) ([]*ai.MCPServerTool, error)
	// GetMCPServerTools 增量获取 MCP Server Tools
	GetMCPServerTools(mtime time.Time, firstUpdate bool) ([]*ai.MCPServerTool, error)

	// QueryMCPServers 查询 MCP Servers（支持过滤和分页）
	QueryMCPServers(query *ai.MCPServerQuery) (uint32, []*ai.MCPServer, error)
}

type A2AAgentStore interface {
	CreateA2AAgent(agent *aitypes.A2AAgent) error
	UpdateA2AAgent(agent *aitypes.A2AAgent) error
	DeleteA2AAgent(id string) error
	GetA2AAgent(id string) (*aitypes.A2AAgent, error)
	GetA2AAgentByName(name, namespace string) (*aitypes.A2AAgent, error)
	GetMoreA2AAgents(mtime time.Time, firstUpdate bool) ([]*aitypes.A2AAgent, error)
	HasA2AAgent(id string) (bool, error)
	HasA2AAgentByName(name, namespace string) (bool, error)
	HasA2AAgentByNameExcludeId(name, namespace, id string) (bool, error)
	GetA2AAgentSkills(agentID string) ([]*aitypes.A2AAgentSkill, error)
	QueryA2AAgents(query *aitypes.A2AAgentQuery) (uint32, []*aitypes.A2AAgent, error)
}
