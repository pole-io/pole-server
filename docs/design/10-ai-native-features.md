# AI Native 特性架构设计

## 1. 概述

Pole Control Plane 是一个 AI Native 的服务治理平台，支持 MCP (Model Context Protocol) 协议。该特性使 AI Agent 能够动态发现服务、获取配置、调用工具，实现 AI 与服务治理的深度融合。

## 2. MCP 协议支持

### 2.1 MCP Server 模型

```go
// apis/pkg/types/ai/mcp.go
type MCPServer struct {
    Id          string
    Name        string
    Namespace   string
    Description string
    Endpoint    string            // MCP Server 端点地址
    Transport   string            // stdio, sse, websocket
    Tools       []*MCPServerTool  // MCP Server 提供的工具
    Metadata    map[string]string
    CreateTime  time.Time
    ModifyTime  time.Time
    Revision    string
}

type MCPServerTool struct {
    Id          string
    ServerId    string
    Name        string
    Description string
    InputSchema map[string]interface{}  // JSON Schema
    OutputSchema map[string]interface{}
    CreateTime  time.Time
    ModifyTime  time.Time
}
```

### 2.2 MCP Server 存储接口

```go
// apis/store/ai.go
type MCPServerStore interface {
    // MCP Server 管理
    CreateMCPServer(server *ai.MCPServer) error
    UpdateMCPServer(server *ai.MCPServer) error
    DeleteMCPServer(id string) error
    GetMCPServer(id string) (*ai.MCPServer, error)
    GetMCPServerByName(name, namespace string) (*ai.MCPServer, error)
    GetMoreMCPServers(mtime time.Time, firstUpdate bool) ([]*ai.MCPServer, error)
    QueryMCPServers(filter map[string]string, offset, limit uint32) (uint32, []*ai.MCPServer, error)

    // MCP Server Tool 管理
    CreateMCPServerTool(tool *ai.MCPServerTool) error
    UpdateMCPServerTool(tool *ai.MCPServerTool) error
    DeleteMCPServerTool(id string) error
    GetMCPServerTool(id string) (*ai.MCPServerTool, error)
    GetMCPServerToolsByServerID(serverID string) ([]*ai.MCPServerTool, error)
}
```

### 2.3 MCP 协议用途

- **服务发现**: AI Agent 通过 MCP 协议发现可用服务
- **工具调用**: AI Agent 调用 MCP Server 提供的工具
- **上下文获取**: AI Agent 获取服务元数据和配置

## 3. AI Agent 集成架构

```
┌─────────────────────────────────────────────────────┐
│                    AI Agent                         │
│  ┌──────────────────────────────────────────────┐  │
│  │              MCP Client                       │  │
│  │  ┌────────────┐  ┌────────────┐              │  │
│  │  │ Discover   │  │  Invoke    │              │  │
│  │  │  Servers   │  │  Tools     │              │  │
│  │  └────────────┘  └────────────┘              │  │
│  └──────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────┘
                         │ MCP Protocol
                         ▼
┌─────────────────────────────────────────────────────┐
│               Pole Control Plane                    │
│  ┌──────────────────────────────────────────────┐  │
│  │              MCP Registry                     │  │
│  │  ┌────────────┐  ┌────────────┐              │  │
│  │  │  Registry  │  │  Execute   │              │  │
│  │  │  Servers   │  │  Tools     │              │  │
│  │  └────────────┘  └────────────┘              │  │
│  └──────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────┐  │
│  │           Service Discovery                  │  │
│  │  ┌────────────┐  ┌────────────┐              │  │
│  │  │  Services  │  │ Instances  │              │  │
│  │  └────────────┘  └────────────┘              │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## 4. 使用场景

### 4.1 AI Agent 服务发现

```go
// AI Agent 通过 MCP 协议发现服务
client.DiscoverServices(ctx, &DiscoverRequest{
    Namespace: "production",
    Labels: map[string]string{
        "type": "api",
    },
})
```

### 4.2 AI Agent 调用工具

```go
// AI Agent 调用 MCP Server 提供的工具
client.InvokeTool(ctx, &InvokeToolRequest{
    Server: "data-service",
    Tool: "query_database",
    Input: map[string]interface{}{
        "sql": "SELECT * FROM users",
    },
})
```

## 5. 配置示例

```yaml
# MCP Server 配置
mcp:
  servers:
    - name: "data-analyzer"
      namespace: "ai-tools"
      endpoint: "stdio:///path/to/analyzer"
      transport: "stdio"
      tools:
        - name: "analyze_data"
          description: "Analyze data and return insights"
```
