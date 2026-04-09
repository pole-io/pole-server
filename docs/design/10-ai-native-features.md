# AI Native 特性架构设计

## 1. 概述

Pole Control Plane 是一个 AI Native 的服务治理平台，支持 MCP (Model Context Protocol) 协议和 Skill Hub 功能。这些特性使 AI Agent 能够动态发现服务、获取配置、调用技能，实现 AI 与服务治理的深度融合。

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

## 3. Skill Hub 功能

### 3.1 Skill 模型

```go
// apis/pkg/types/ai/skill.go
type Skill struct {
    Id          string
    Name        string
    Namespace   string
    Description string
    Category    string            // tool, prompt, resource
    Tags        []string
    Metadata    map[string]string
    IsActive    bool
    CreateTime  time.Time
    ModifyTime  time.Time
}

type SkillVersion struct {
    Id          string
    SkillId     string
    SkillName   string
    Namespace   string
    Version     uint64
    Content     string            // Skill 定义内容
    IsActive    bool              // 是否为活跃版本
    CreateTime  time.Time
    ModifyTime  time.Time
}
```

### 3.2 Skill Group 模型

```go
type SkillGroup struct {
    Id          string
    Name        string
    Namespace   string
    Description string
    Skills      []*SkillReference
    Metadata    map[string]string
    CreateTime  time.Time
    ModifyTime  time.Time
}

type SkillReference struct {
    SkillId   string
    SkillName string
    Version   uint64  // 0 表示使用活跃版本
}
```

### 3.3 Skill 订阅模型

```go
type SkillSubscription struct {
    Id          string
    ClientId    string
    SkillName   string
    Namespace   string
    Version     uint64      // 订阅的版本
    IsActive    bool
    CreateTime  time.Time
    ModifyTime  time.Time
}
```

## 4. Skill Hub 存储接口

### 4.1 SkillStore

```go
type SkillStore interface {
    CreateSkill(skill *ai.Skill) error
    UpdateSkill(skill *ai.Skill) error
    DeleteSkill(id string) error
    GetSkill(id string) (*ai.Skill, error)
    GetSkillByName(name, namespace string) (*ai.Skill, error)
    GetMoreSkills(mtime time.Time, firstUpdate bool) ([]*ai.Skill, error)
}
```

### 4.2 SkillGroupStore

```go
type SkillGroupStore interface {
    CreateSkillGroup(group *ai.SkillGroup) (*ai.SkillGroup, error)
    UpdateSkillGroup(group *ai.SkillGroup) error
    GetSkillGroup(namespace, name string) (*ai.SkillGroup, error)
    DeleteSkillGroup(namespace, name string) error
    GetMoreSkillGroups(firstUpdate bool, mtime time.Time) ([]*ai.SkillGroup, error)
}
```

### 4.3 SkillVersionStore

```go
type SkillVersionStore interface {
    CreateSkillVersion(version *ai.SkillVersion) error
    UpdateSkillVersion(version *ai.SkillVersion) error
    DeleteSkillVersion(id string) error
    GetSkillVersion(id string) (*ai.SkillVersion, error)
    GetSkillVersionByVersion(skillName, namespace string, version uint64) (*ai.SkillVersion, error)
    GetActiveSkillVersion(skillName, namespace string) (*ai.SkillVersion, error)
    ActiveSkillVersion(version *ai.SkillVersion) error
    InactiveSkillVersion(version *ai.SkillVersion) error
}
```

### 4.4 SkillSubscriptionStore

```go
type SkillSubscriptionStore interface {
    CreateSkillSubscription(sub *ai.SkillSubscription) error
    UpdateSkillSubscription(sub *ai.SkillSubscription) error
    DeleteSkillSubscription(id string) error
    GetSkillSubscriptionByClient(clientID string) ([]*ai.SkillSubscription, error)
    GetSkillSubscriptionsBySkill(skillName, namespace string) ([]*ai.SkillSubscription, error)
    UpdateSubscriptionVersion(clientID, skillName, namespace string, version uint64) error
    DeactiveSubscription(clientID, skillName, namespace string) error
}
```

## 5. Skill Hub 与配置中心的关系

Skill Hub 复用配置中心的基础设施：

```
┌─────────────────────────────────────────────────────┐
│                    Skill Hub                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │  Skill   │  │  Group   │  │  Sub     │         │
│  │  CRUD    │  │  CRUD    │  │  CRUD    │         │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘         │
│       │             │             │                │
│       └─────────────┼─────────────┘                │
│                     │                              │
├─────────────────────┼──────────────────────────────┤
│                     ▼                              │
│  ┌──────────────────────────────────────────┐     │
│  │           Config Center                   │     │
│  │  ┌────────────┐  ┌────────────┐          │     │
│  │  │ ConfigFile │  │ ConfigGroup│          │     │
│  │  │   Cache    │  │   Cache    │          │     │
│  │  └────────────┘  └────────────┘          │     │
│  └──────────────────────────────────────────┘     │
│                     │                              │
│                     ▼                              │
│  ┌──────────────────────────────────────────┐     │
│  │           Storage Layer                   │     │
│  │  ┌────────────┐  ┌────────────┐          │     │
│  │  │   MySQL    │  │   Cache    │          │     │
│  │  └────────────┘  └────────────┘          │     │
│  └──────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────┘
```

## 6. AI Agent 集成架构

```
┌─────────────────────────────────────────────────────┐
│                    AI Agent                         │
│  ┌──────────────────────────────────────────────┐  │
│  │              MCP Client                       │  │
│  │  ┌────────────┐  ┌────────────┐              │  │
│  │  │ Discover   │  │  Invoke    │              │  │
│  │  │  Skills    │  │  Tools     │              │  │
│  │  └────────────┘  └────────────┘              │  │
│  └──────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────┘
                         │ MCP Protocol
                         ▼
┌─────────────────────────────────────────────────────┐
│               Pole Control Plane                    │
│  ┌──────────────────────────────────────────────┐  │
│  │              MCP Server                       │  │
│  │  ┌────────────┐  ┌────────────┐              │  │
│  │  │  Registry  │  │  Execute   │              │  │
│  │  │  Skills    │  │  Tools     │              │  │
│  │  └────────────┘  └────────────┘              │  │
│  └──────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────┐  │
│  │              Skill Hub                        │  │
│  │  ┌────────────┐  ┌────────────┐              │  │
│  │  │   Skills   │  │   Groups   │              │  │
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

## 7. Skill 版本管理

### 7.1 版本控制

- 每个 Skill 支持多版本并存
- 可指定活跃版本供默认调用
- 支持版本回滚

### 7.2 订阅机制

- 客户端订阅特定版本的 Skill
- 版本更新时主动推送变更通知
- 支持订阅状态管理

## 8. 使用场景

### 8.1 AI Agent 服务发现

```go
// AI Agent 通过 MCP 协议发现服务
client.DiscoverServices(ctx, &DiscoverRequest{
    Namespace: "production",
    Labels: map[string]string{
        "type": "api",
    },
})
```

### 8.2 AI Agent 调用工具

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

### 8.3 AI Agent 获取 Skill

```go
// AI Agent 获取 Skill Hub 中的技能
client.GetSkill(ctx, &GetSkillRequest{
    Name: "text-processor",
    Namespace: "ai-tools",
    Version: 0,  // 0 表示使用活跃版本
})
```

## 9. 配置示例

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

# Skill Hub 配置
skillHub:
  skills:
    - name: "text-processor"
      namespace: "ai-tools"
      category: "tool"
      tags:
        - "nlp"
        - "text"
      versions:
        - version: 1
          content: "..."
          isActive: true
```

## 10. 未来规划

1. **Skill 市场**: 支持 Skill 的分享和发现
2. **Skill 组合**: 支持多个 Skill 组合成工作流
3. **Skill 监控**: 提供调用链追踪和性能监控
4. **Skill 安全**: 支持 Skill 执行沙箱和权限控制
