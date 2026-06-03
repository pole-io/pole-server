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

package ai

import (
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/specification/source/go/api/v1/ai"
)

type mcpServerCache struct {
	*cachebase.BaseCache
	storage store.Store

	// id -> *ai.MCPServer
	ids *container.SyncMap[string, *ai.MCPServer]

	// namespace/name -> *ai.MCPServer
	names *container.SyncMap[string, *ai.MCPServer]

	// namespace -> []*ai.MCPServer
	namespaceIndex *container.SyncMap[string, []*ai.MCPServer]

	// serverID -> []*ai.MCPServerTool
	tools *container.SyncMap[string, []*ai.MCPServerTool]

	singleFlight *singleflight.Group
}

func NewMCPServerCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.Cache {
	return &mcpServerCache{
		BaseCache:      cachebase.NewBaseCache(storage, cacheMgr),
		storage:        storage,
		ids:            container.NewSyncMap[string, *ai.MCPServer](),
		names:          container.NewSyncMap[string, *ai.MCPServer](),
		namespaceIndex: container.NewSyncMap[string, []*ai.MCPServer](),
		tools:          container.NewSyncMap[string, []*ai.MCPServerTool](),
		singleFlight:   &singleflight.Group{},
	}
}

func (mc *mcpServerCache) Name() string {
	return cacheapi.MCPServerName
}

func (mc *mcpServerCache) Initialize(_ map[string]any) error {
	return nil
}

func (mc *mcpServerCache) Update() error {
	_, err, _ := mc.singleFlight.Do(mc.Name(), func() (any, error) {
		return nil, mc.DoCacheUpdate(mc.Name(), mc.realUpdate)
	})
	return err
}

func (mc *mcpServerCache) realUpdate() (map[string]time.Time, int64, error) {
	results := make(map[string]time.Time)

	// 更新 MCP Server 数据
	servers, err := mc.storage.GetMoreMCPServers(mc.LastFetchTime(), mc.IsFirstUpdate())
	if err != nil {
		return nil, 0, err
	}

	upsert := 0
	del := 0

	for _, server := range servers {
		serverMTime := parseMCPTime(server.Mtime)
		if serverMTime.After(mc.LastFetchTime()) {
			mc.LastFetchTime().Add(time.Nanosecond)
		}

		results[server.Id] = serverMTime

		if server.Flag == 1 { // Flag=1 表示已删除
			mc.removeMCPServer(server)
			del++
		} else {
			mc.storeMCPServer(server)
			upsert++
		}
	}

	// 更新 MCP Server Tools
	tools, err := mc.storage.GetMCPServerTools(mc.LastFetchTime(), mc.IsFirstUpdate())
	if err != nil {
		log.Warnf("[Cache][MCPServer] get tools from store err: %s", err.Error())
	} else {
		mc.updateTools(tools)
	}

	log.Infof("[Cache][MCPServer] update servers, upsert: %d, delete: %d, total: %d",
		upsert, del, len(servers))
	return results, int64(len(servers)), nil
}

func (mc *mcpServerCache) Clear() error {
	mc.BaseCache.Clear()
	// 重新初始化 SyncMap
	mc.ids = container.NewSyncMap[string, *ai.MCPServer]()
	mc.names = container.NewSyncMap[string, *ai.MCPServer]()
	mc.namespaceIndex = container.NewSyncMap[string, []*ai.MCPServer]()
	mc.tools = container.NewSyncMap[string, []*ai.MCPServerTool]()
	return nil
}

func (mc *mcpServerCache) Close() error {
	return nil
}

// GetMCPServerByID 实现 MCPServerCache 接口
func (mc *mcpServerCache) GetMCPServerByID(id string) *ai.MCPServer {
	if id == "" {
		return nil
	}
	val, ok := mc.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

// GetMCPServerByName 实现 MCPServerCache 接口
func (mc *mcpServerCache) GetMCPServerByName(name, namespace string) *ai.MCPServer {
	if name == "" || namespace == "" {
		return nil
	}
	key := mcpServerNameKey(namespace, name)
	val, ok := mc.names.Load(key)
	if !ok {
		return nil
	}
	return val
}

// GetMCPServersByNamespace 实现 MCPServerCache 接口
func (mc *mcpServerCache) GetMCPServersByNamespace(namespace string) []*ai.MCPServer {
	if namespace == "" {
		return nil
	}
	val, ok := mc.namespaceIndex.Load(namespace)
	if !ok {
		return nil
	}
	return val
}

// GetMCPServerTools 实现 MCPServerCache 接口
func (mc *mcpServerCache) GetMCPServerTools(serverID string) []*ai.MCPServerTool {
	if serverID == "" {
		return nil
	}
	val, ok := mc.tools.Load(serverID)
	if !ok {
		return nil
	}
	return val
}

// Query 实现 MCPServerCache 接口
// query.Name 为前缀匹配，Namespace/Business/Department/Protocol 为精确匹配
// 排序: MTime DESC
func (mc *mcpServerCache) Query(query *ai.MCPServerQuery) (uint32, []*ai.MCPServer) {
	if query == nil {
		query = &ai.MCPServerQuery{}
	}

	matched := make([]*ai.MCPServer, 0, 16)
	mc.ids.Range(func(_ string, s *ai.MCPServer) {
		if s == nil {
			return
		}
		if query.Name != "" && !strings.HasPrefix(s.Name, query.Name) {
			return
		}
		if query.Namespace != "" && s.Namespace != query.Namespace {
			return
		}
		if query.Business != "" && s.Business != query.Business {
			return
		}
		if query.Department != "" && s.Department != query.Department {
			return
		}
		if query.Protocol != "" && s.Protocol != query.Protocol {
			return
		}
		matched = append(matched, s)
	})

	sort.SliceStable(matched, func(i, j int) bool {
		return parseMCPTime(matched[i].Mtime).After(parseMCPTime(matched[j].Mtime))
	})

	total := uint32(len(matched))
	if query.Offset >= total {
		return total, []*ai.MCPServer{}
	}
	end := query.Offset + query.Limit
	if end > total {
		end = total
	}
	return total, matched[query.Offset:end]
}

// 辅助方法
func mcpServerKey(server *ai.MCPServer) string {
	return server.Id
}

func mcpServerNameKey(namespace, name string) string {
	return namespace + "/" + name
}

const mcpTimeLayout = "2006-01-02 15:04:05"

func formatMCPTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(mcpTimeLayout)
}

func parseMCPTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{mcpTimeLayout, time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func (mc *mcpServerCache) storeMCPServer(server *ai.MCPServer) {
	// 存储 ID 索引
	mc.ids.Store(server.Id, server)

	// 存储名称索引
	key := mcpServerNameKey(server.Namespace, server.Name)
	mc.names.Store(key, server)

	// 更新命名空间索引
	mc.updateNamespaceIndex(server)
}

func (mc *mcpServerCache) removeMCPServer(server *ai.MCPServer) {
	// 删除 ID 索引
	mc.ids.Delete(server.Id)

	// 删除名称索引
	key := mcpServerNameKey(server.Namespace, server.Name)
	mc.names.Delete(key)

	// 从命名空间索引中移除
	mc.removeFromNamespaceIndex(server)

	// 删除关联的工具
	mc.tools.Store(server.Id, []*ai.MCPServerTool{})
}

func (mc *mcpServerCache) updateNamespaceIndex(server *ai.MCPServer) {
	namespace := server.Namespace
	servers, _ := mc.namespaceIndex.ComputeIfAbsent(namespace, func(k string) []*ai.MCPServer {
		return make([]*ai.MCPServer, 0)
	})
	// 先移除已存在的
	for i, s := range servers {
		if s.Id == server.Id {
			servers = append(servers[:i], servers[i+1:]...)
			break
		}
	}
	servers = append(servers, server)
	mc.namespaceIndex.Store(namespace, servers)
}

func (mc *mcpServerCache) removeFromNamespaceIndex(server *ai.MCPServer) {
	namespace := server.Namespace
	servers, ok := mc.namespaceIndex.Load(namespace)
	if !ok {
		return
	}
	// 移除
	for i, s := range servers {
		if s.Id == server.Id {
			servers = append(servers[:i], servers[i+1:]...)
			break
		}
	}
	if len(servers) == 0 {
		mc.namespaceIndex.Store(namespace, []*ai.MCPServer{})
	} else {
		mc.namespaceIndex.Store(namespace, servers)
	}
}

func (mc *mcpServerCache) updateTools(tools []*ai.MCPServerTool) {
	// 按 serverID 分组
	toolsByServer := make(map[string][]*ai.MCPServerTool)
	for _, tool := range tools {
		if tool.Flag == 1 {
			// 删除工具
			mc.removeTool(tool)
		} else {
			toolsByServer[tool.McpServerId] = append(toolsByServer[tool.McpServerId], tool)
		}
	}

	// 更新索引
	for serverID, serverTools := range toolsByServer {
		mc.tools.Store(serverID, serverTools)
	}
}

func (mc *mcpServerCache) removeTool(tool *ai.MCPServerTool) {
	serverTools, ok := mc.tools.Load(tool.McpServerId)
	if !ok {
		return
	}
	// 移除
	for i, t := range serverTools {
		if t.Id == tool.Id {
			serverTools = append(serverTools[:i], serverTools[i+1:]...)
			break
		}
	}
	if len(serverTools) == 0 {
		mc.tools.Store(tool.McpServerId, []*ai.MCPServerTool{})
	} else {
		mc.tools.Store(tool.McpServerId, serverTools)
	}
}
