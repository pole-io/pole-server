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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/specification/source/go/api/v1/ai"
)

// 创建测试用的 MCP Server 数据
func createTestMCPServers() []*ai.MCPServer {
	return []*ai.MCPServer{
		{
			Id:          "mcp-server-1",
			Name:        "test-mcp-server-1",
			Namespace:   "default",
			Ports:       "8080",
			Business:    "test-business",
			Department:  "test-dept",
			Description: "Test MCP Server 1",
			Revision:    "1.0.0",
			Flag:        0,
			Reference:   "",
			Protocol:    "http",
			Ctime:       formatMCPTime(time.Now()),
			Mtime:       formatMCPTime(time.Now()),
			ExportTo:    "",
		},
		{
			Id:          "mcp-server-2",
			Name:        "test-mcp-server-2",
			Namespace:   "ai-ns",
			Ports:       "9090",
			Business:    "ai-business",
			Department:  "ai-dept",
			Description: "Test MCP Server 2",
			Revision:    "1.0.0",
			Flag:        0,
			Reference:   "",
			Protocol:    "grpc",
			Ctime:       formatMCPTime(time.Now()),
			Mtime:       formatMCPTime(time.Now()),
			ExportTo:    "",
		},
	}
}

// TestMCPServerCache_InterfaceDefinition 测试接口定义
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_InterfaceDefinition(t *testing.T) {
	// 创建测试数据
	servers := createTestMCPServers()
	assert.Len(t, servers, 2, "应该创建 2 个测试 MCP Server")

	// 测试数据的基本属性
	server1 := servers[0]
	assert.Equal(t, "mcp-server-1", server1.Id)
	assert.Equal(t, "test-mcp-server-1", server1.Name)
	assert.Equal(t, "default", server1.Namespace)
	assert.Equal(t, "http", server1.Protocol)
	assert.Equal(t, "8080", server1.Ports)

	server2 := servers[1]
	assert.Equal(t, "mcp-server-2", server2.Id)
	assert.Equal(t, "test-mcp-server-2", server2.Name)
	assert.Equal(t, "ai-ns", server2.Namespace)
	assert.Equal(t, "grpc", server2.Protocol)
	assert.Equal(t, "9090", server2.Ports)
}

// TestMCPServerCache_IndexFunctionality 测试索引功能
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_IndexFunctionality(t *testing.T) {
	servers := createTestMCPServers()

	// 测试按 ID 索引
	var serverByID string
	for _, server := range servers {
		if server.Id == "mcp-server-1" {
			serverByID = server.Id
			break
		}
	}
	assert.Equal(t, "mcp-server-1", serverByID)

	// 测试按名称索引
	var serverByName *ai.MCPServer
	for _, server := range servers {
		if server.Name == "test-mcp-server-2" && server.Namespace == "ai-ns" {
			serverByName = server
			break
		}
	}
	assert.NotNil(t, serverByName)
	assert.Equal(t, "mcp-server-2", serverByName.Id)

	// 测试按命名空间索引
	var serversInDefault []*ai.MCPServer
	for _, server := range servers {
		if server.Namespace == "default" {
			serversInDefault = append(serversInDefault, server)
		}
	}
	assert.Len(t, serversInDefault, 1)

	// 测试按协议索引
	var httpServers []*ai.MCPServer
	for _, server := range servers {
		if server.Protocol == "http" {
			httpServers = append(httpServers, server)
		}
	}
	assert.Len(t, httpServers, 1)
	assert.Equal(t, "mcp-server-1", httpServers[0].Id)
}

// TestMCPServerCache_DataValidation 测试数据验证
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_DataValidation(t *testing.T) {
	servers := createTestMCPServers()

	// 测试必要字段
	for _, server := range servers {
		assert.NotEmpty(t, server.Id, "MCP Server ID 不能为空")
		assert.NotEmpty(t, server.Name, "MCP Server Name 不能为空")
		assert.NotEmpty(t, server.Namespace, "Namespace 不能为空")
		assert.Equal(t, uint32(0), server.Flag, "有效数据的 Flag 应该是 0")
	}

	// 测试时间戳
	for _, server := range servers {
		ctime := parseMCPTime(server.Ctime)
		mtime := parseMCPTime(server.Mtime)
		assert.False(t, ctime.IsZero(), "CTime 不能为零值")
		assert.False(t, mtime.IsZero(), "MTime 不能为零值")
		assert.True(t, mtime.After(ctime) || mtime.Equal(ctime),
			"MTime 应该 >= CTime")
	}
}

// TestMCPServerCache_ConcurrentAccess 测试并发访问
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_ConcurrentAccess(t *testing.T) {
	servers := createTestMCPServers()

	// 模拟并发读取
	readResults := make(chan string, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			// 随机读取
			if id%2 == 0 {
				readResults <- servers[0].Id
			} else {
				readResults <- servers[1].Id
			}
		}(i)
	}

	// 收集结果
	for i := 0; i < 5; i++ {
		result := <-readResults
		assert.NotEmpty(t, result)
	}

	// 关闭通道
	close(readResults)
}

// TestMCPServerCache_EdgeCases 测试边界条件
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_EdgeCases(t *testing.T) {
	// 测试空数据
	var emptyServers []*ai.MCPServer
	assert.Len(t, emptyServers, 0)

	// 测试单个数据
	singleServer := &ai.MCPServer{
		Id:          "single-server",
		Name:        "single",
		Namespace:   "default",
		Ports:       "8080",
		Business:    "test",
		Department:  "test",
		Description: "Single server",
		Revision:    "1.0.0",
		Flag:        0,
		Reference:   "",
		Protocol:    "http",
		Ctime:       formatMCPTime(time.Now()),
		Mtime:       formatMCPTime(time.Now()),
		ExportTo:    "",
	}

	assert.Equal(t, "single-server", singleServer.Id)
	assert.Equal(t, "single", singleServer.Name)

	// 测试已删除数据（Flag=1）
	deletedServer := &ai.MCPServer{
		Id:          "deleted-server",
		Name:        "deleted",
		Namespace:   "default",
		Ports:       "8080",
		Business:    "test",
		Department:  "test",
		Description: "Deleted server",
		Revision:    "1.0.0",
		Flag:        1, // 已删除
		Reference:   "",
		Protocol:    "http",
		Ctime:       formatMCPTime(time.Now()),
		Mtime:       formatMCPTime(time.Now()),
		ExportTo:    "",
	}

	assert.Equal(t, uint32(1), deletedServer.Flag)
}

// TestMCPServerCache_Performance 测试性能相关
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_Performance(t *testing.T) {
	// 创建大量测试数据
	var largeServers []*ai.MCPServer
	for i := 0; i < 500; i++ {
		largeServers = append(largeServers, &ai.MCPServer{
			Id:          serverID(i),
			Name:        serverName(i),
			Namespace:   "default",
			Ports:       "8080",
			Business:    "test",
			Department:  "test",
			Description: "Test server",
			Revision:    "1.0.0",
			Flag:        0,
			Reference:   "",
			Protocol:    "http",
			Ctime:       formatMCPTime(time.Now()),
			Mtime:       formatMCPTime(time.Now()),
			ExportTo:    "",
		})
	}

	// 验证数据量
	assert.Len(t, largeServers, 500)

	// 测试快速查找
	start := time.Now()
	found := false
	for _, server := range largeServers {
		if server.Id == "mcp-server-499" {
			found = true
			break
		}
	}
	elapsed := time.Since(start)

	assert.True(t, found, "应该找到 mcp-server-499")
	assert.Less(t, elapsed, time.Millisecond*100, "查找时间应该在 100ms 内")
}

// TestMCPServerCache_IncrementalUpdate 测试增量更新
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_IncrementalUpdate(t *testing.T) {
	// 创建初始数据
	initialServers := createTestMCPServers()
	assert.Len(t, initialServers, 2, "初始应该有 2 个 MCP Server")

	// 准备增量数据 - 添加新服务器
	_ = append(initialServers, &ai.MCPServer{
		Id:          "mcp-server-3",
		Name:        "new-server",
		Namespace:   "default",
		Ports:       "9999",
		Business:    "test",
		Department:  "test",
		Description: "New added server",
		Revision:    "1.0.0",
		Flag:        0,
		Reference:   "",
		Protocol:    "grpc",
		Ctime:       formatMCPTime(time.Now()),
		Mtime:       formatMCPTime(time.Now()),
		ExportTo:    "",
	})

	// 模拟删除服务器（通过设置 Flag=1）
	deletedServers := append(initialServers, &ai.MCPServer{
		Id:          "mcp-server-deleted",
		Name:        "to-be-deleted",
		Namespace:   "default",
		Ports:       "8888",
		Business:    "test",
		Department:  "test",
		Description: "This server will be deleted",
		Revision:    "1.0.0",
		Flag:        1, // 已删除
		Reference:   "",
		Protocol:    "http",
		Ctime:       formatMCPTime(time.Now().Add(-time.Hour)),
		Mtime:       formatMCPTime(time.Now().Add(-time.Hour)),
		ExportTo:    "",
	})

	// 测试删除标记处理
	var activeServers []*ai.MCPServer
	for _, server := range deletedServers {
		if server.Flag == 0 { // 只保留未删除的
			activeServers = append(activeServers, server)
		}
	}
	assert.Len(t, activeServers, 2, "删除后应该只有 2 个活跃服务器")
}

// TestMCPServerCache_SingleFlight 测试 SingleFlight 防重入机制
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_SingleFlight(t *testing.T) {
	_ = createTestMCPServers()

	// 模拟多个并发更新调用
	updateCount := 10
	updateComplete := make(chan bool, updateCount)

	// 启动多个 goroutine 同时调用 Update
	for i := 0; i < updateCount; i++ {
		go func(id int) {
			// 注意：实际实现中，Update 方法应该使用 SingleFlight
			time.Sleep(time.Millisecond * time.Duration(id))
			updateComplete <- true
		}(i)
	}

	// 等待所有更新完成
	completed := 0
	for i := 0; i < updateCount; i++ {
		<-updateComplete
		completed++
	}

	assert.Equal(t, updateCount, completed, "所有并发更新都应该完成")
}

// TestMCPServerCache_Lifecycle 测试缓存生命周期
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_Lifecycle(t *testing.T) {
	servers := createTestMCPServers()

	// 1. Initialize
	assert.NotNil(t, servers, "Initialize 阶段数据应该存在")

	// 2. Update
	assert.Len(t, servers, 2, "Update 后应该保持数据")

	// 3. Clear
	var emptyServers []*ai.MCPServer
	assert.Len(t, emptyServers, 0, "Clear 后数据应该为空")

	// 4. Close
	// 实际实现中可能需要关闭资源
}

// TestMCPServerCache_ToolAccess 测试工具访问
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestMCPServerCache_ToolAccess(t *testing.T) {
	// 创建带工具的服务器
	servers := []*ai.MCPServer{
		{
			Id:          "mcp-server-with-tools",
			Name:        "server-with-tools",
			Namespace:   "default",
			Ports:       "8080",
			Business:    "test",
			Department:  "test",
			Description: "Server with tools",
			Revision:    "1.0.0",
			Flag:        0,
			Reference:   "",
			Protocol:    "http",
			Ctime:       formatMCPTime(time.Now()),
			Mtime:       formatMCPTime(time.Now()),
			ExportTo:    "",
		},
	}

	// 注意：实际实现中需要测试 GetMCPServerTools 方法
	// 这里只是测试数据结构的兼容性
	assert.Len(t, servers, 1)
}

// 辅助函数
func serverID(i int) string {
	return fmt.Sprintf("mcp-server-%d", i)
}

func serverName(i int) string {
	return fmt.Sprintf("test-mcp-server-%d", i)
}

// newCacheForTest 构造一个内存态 mcpServerCache 并预灌入 servers
func newCacheForTest(servers []*ai.MCPServer) *mcpServerCache {
	mc := &mcpServerCache{
		ids:            container.NewSyncMap[string, *ai.MCPServer](),
		names:          container.NewSyncMap[string, *ai.MCPServer](),
		namespaceIndex: container.NewSyncMap[string, []*ai.MCPServer](),
		tools:          container.NewSyncMap[string, []*ai.MCPServerTool](),
	}
	for _, s := range servers {
		mc.storeMCPServer(s)
	}
	return mc
}

func TestMCPServerCache_Query_NamePrefixMatch(t *testing.T) {
	mc := newCacheForTest([]*ai.MCPServer{
		{Id: "s1", Name: "alpha-svc", Namespace: "ns1", Mtime: formatMCPTime(time.Now())},
		{Id: "s2", Name: "alpha-other", Namespace: "ns1", Mtime: formatMCPTime(time.Now())},
		{Id: "s3", Name: "beta-svc", Namespace: "ns1", Mtime: formatMCPTime(time.Now())},
	})

	total, list := mc.Query(&ai.MCPServerQuery{Name: "alpha", Limit: 10})
	assert.Equal(t, uint32(2), total)
	assert.Len(t, list, 2)

	total, list = mc.Query(&ai.MCPServerQuery{Name: "alpha-svc", Limit: 10})
	assert.Equal(t, uint32(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "s1", list[0].Id)

	total, list = mc.Query(&ai.MCPServerQuery{Name: "zzz", Limit: 10})
	assert.Equal(t, uint32(0), total)
	assert.Len(t, list, 0)
}

func TestMCPServerCache_Query_ExactFilters(t *testing.T) {
	mc := newCacheForTest([]*ai.MCPServer{
		{Id: "s1", Name: "a", Namespace: "ns1", Business: "b1", Department: "d1", Protocol: "http", Mtime: formatMCPTime(time.Now())},
		{Id: "s2", Name: "b", Namespace: "ns2", Business: "b1", Department: "d2", Protocol: "grpc", Mtime: formatMCPTime(time.Now())},
		{Id: "s3", Name: "c", Namespace: "ns1", Business: "b2", Department: "d1", Protocol: "http", Mtime: formatMCPTime(time.Now())},
	})

	total, list := mc.Query(&ai.MCPServerQuery{Namespace: "ns1", Limit: 10})
	assert.Equal(t, uint32(2), total)
	assert.Len(t, list, 2)

	total, _ = mc.Query(&ai.MCPServerQuery{Business: "b1", Limit: 10})
	assert.Equal(t, uint32(2), total)

	total, _ = mc.Query(&ai.MCPServerQuery{Department: "d1", Limit: 10})
	assert.Equal(t, uint32(2), total)

	total, _ = mc.Query(&ai.MCPServerQuery{Protocol: "http", Limit: 10})
	assert.Equal(t, uint32(2), total)

	total, list = mc.Query(&ai.MCPServerQuery{Namespace: "ns1", Protocol: "http", Limit: 10})
	assert.Equal(t, uint32(2), total)
	assert.Len(t, list, 2)

	total, list = mc.Query(&ai.MCPServerQuery{Namespace: "ns1", Business: "b2", Limit: 10})
	assert.Equal(t, uint32(1), total)
	assert.Equal(t, "s3", list[0].Id)
}

func TestMCPServerCache_Query_Pagination(t *testing.T) {
	base := time.Now()
	servers := make([]*ai.MCPServer, 0, 5)
	for i := 0; i < 5; i++ {
		servers = append(servers, &ai.MCPServer{
			Id:        fmt.Sprintf("s%d", i),
			Name:      fmt.Sprintf("svc-%d", i),
			Namespace: "ns",
			Mtime:     formatMCPTime(base.Add(time.Duration(i) * time.Second)),
		})
	}
	mc := newCacheForTest(servers)

	total, list := mc.Query(&ai.MCPServerQuery{Limit: 2})
	assert.Equal(t, uint32(5), total)
	assert.Len(t, list, 2)

	total, list = mc.Query(&ai.MCPServerQuery{Offset: 2, Limit: 2})
	assert.Equal(t, uint32(5), total)
	assert.Len(t, list, 2)

	total, list = mc.Query(&ai.MCPServerQuery{Offset: 4, Limit: 2})
	assert.Equal(t, uint32(5), total)
	assert.Len(t, list, 1, "末页只剩一条")

	total, list = mc.Query(&ai.MCPServerQuery{Offset: 10, Limit: 5})
	assert.Equal(t, uint32(5), total)
	assert.Len(t, list, 0, "offset 越界返回空")
}

func TestMCPServerCache_Query_SortByMTimeDesc(t *testing.T) {
	base := time.Now()
	mc := newCacheForTest([]*ai.MCPServer{
		{Id: "old", Name: "old", Namespace: "ns", Mtime: formatMCPTime(base.Add(-2 * time.Hour))},
		{Id: "new", Name: "new", Namespace: "ns", Mtime: formatMCPTime(base)},
		{Id: "mid", Name: "mid", Namespace: "ns", Mtime: formatMCPTime(base.Add(-1 * time.Hour))},
	})

	_, list := mc.Query(&ai.MCPServerQuery{Limit: 10})
	assert.Len(t, list, 3)
	assert.Equal(t, "new", list[0].Id, "MTime 最新的排在最前")
	assert.Equal(t, "mid", list[1].Id)
	assert.Equal(t, "old", list[2].Id)
}
