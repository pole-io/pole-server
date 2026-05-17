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

package cache

import (
	"github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// MCPServerCache MCP Server 缓存接口
type MCPServerCache interface {
	Cache

	// GetMCPServerByID 根据 ID 获取 MCP Server
	GetMCPServerByID(id string) *ai.MCPServer

	// GetMCPServerByName 根据名称和命名空间获取 MCP Server
	GetMCPServerByName(name, namespace string) *ai.MCPServer

	// GetMCPServersByNamespace 获取命名空间下的所有 MCP Server
	GetMCPServersByNamespace(namespace string) []*ai.MCPServer

	// GetMCPServerTools 获取 MCP Server 的所有工具
	GetMCPServerTools(serverID string) []*ai.MCPServerTool

	// Query 按过滤条件分页查询，返回总数和当前页结果
	// filter 支持的 key:
	//   - name       前缀匹配（strings.HasPrefix）
	//   - namespace  精确匹配
	//   - business   精确匹配
	//   - department 精确匹配
	//   - protocol   精确匹配
	// 排序：MTime DESC
	Query(filter map[string]string, offset, limit uint32) (uint32, []*ai.MCPServer)
}
