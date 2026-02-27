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

	"github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// DiscoverStore Service discovery storage interface
type AIStore interface {
	MCPServerStore
}

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
}
