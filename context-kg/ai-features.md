---
title: AI 原生功能：MCP 与 Skill Hub
tags: [ai, mcp, skill]
links: [domain-components, storage, cache-layer, api-servers]
updated: 2026-05-14
sources: 1
---

# AI 原生功能：MCP 与 Skill Hub

## 概览

AI 原生功能使 Pole 的服务注册中心和治理能力对 AI 智能体（LLM）可见。主要包含两个组件：

1. **MCP Registry** — 注册和发现 MCP（模型上下文协议）服务器
2. **Skill Hub** — 类配置中心风格的 AI 技能（函数/工具/智能体）注册中心

业务域层面的 Skill Hub 实现见 [[domain-components]]，存储层接口见 [[storage]]，缓存机制见 [[cache-layer]]，HTTP API 端点见 [[api-servers]]。

---

## MCP 服务器注册中心

### 是什么

MCP（Model Context Protocol，模型上下文协议）是一个开放标准，使 AI 智能体能够发现并调用外部工具。Pole 的 MCP Registry 支持：
- 注册 MCP 服务器（地址、工具、能力）
- 查询已注册的 MCP 服务器（按名称、命名空间、类型过滤）
- 让 AI 智能体通过 Pole 发现可用的 MCP 服务器

### 领域模型（`apis/pkg/types/ai/mcp.go`）

```go
type MCPServer struct {
    ID          string
    Name        string       // 命名空间内唯一
    Namespace   string
    Description string
    Protocol    string       // "http"、"stdio"、"sse"
    Endpoint    string       // URL 或命令
    Type        string       // 服务器分类
    Metadata    map[string]string
    Flag        int          // 0=活跃，1=已删除
    CTime, MTime time.Time
}

type MCPServerTool struct {
    ID          string
    ServerID    string       // 外键，关联 MCPServer
    Name        string
    Description string
    InputSchema string       // JSON Schema
}
```

### 存储（`apis/store/ai.go → MCPServerStore`）

关键方法：
- `CreateMCPServer`、`UpdateMCPServer`、`DeleteMCPServer`
- `GetMCPServer(id)`、`GetMCPServerByName(name, namespace)`
- `GetMoreMCPServers(mtime, firstUpdate)` — 增量缓存同步
- `QueryMCPServers(filter, offset, limit)` — 分页搜索
- 工具管理：`CreateMCPServerTool`、`GetMCPServerToolsByServerID`

### HTTP API（`plugin/apiserver/httpserver/aimcp/`）

- `GET /mcp/v1/servers` — 列出/过滤 MCP 服务器
- `POST /mcp/v1/servers` — 批量创建 MCP 服务器

### MCP 协议暴露（`plugin/apiserver/httpserver/aimcp/mcp_server.go`）

Pole 将自身的管理功能*以* MCP 工具的形式暴露，供 AI 智能体使用：

```
工具：list_mcp_servers
  输入：{namespace?, name?, type?, offset, limit}
  输出：MCPServer 对象数组

工具：create_mcp_servers
  输入：MCPServer 规格数组
  输出：批量写入响应
```

这些工具通过 `mark3labs/mcp-go` 注册，并通过 HTTP（SSE 或流式传输）提供服务。

---

## Skill Hub

### 是什么

Skill Hub 类似于配置中心，但专门用于存储 AI 技能定义。它存储结构化的技能定义（AI 函数/工具/智能体），具备：
- 输入/输出 Schema（JSON Schema）
- 版本历史
- 客户端订阅
- 分组管理

### 领域模型（`apis/pkg/types/ai/skill.go`）

```go
type Skill struct {
    ID           string
    Name         string
    Namespace    string
    Description  string
    InputSchema  string      // 输入参数的 JSON Schema
    OutputSchema string      // 输出结果的 JSON Schema
    SkillType    string      // "function" | "tool" | "agent"
    Author       string
    Business     string
    Department   string
    Metadata     map[string]string
    Protocol     string      // 例如 "mcp"、"openai-function"
    Revision     string      // 内容哈希，用于变更检测
    ExportTo     []string    // 共享到其他命名空间
    Flag         int         // 0=可见，1=软删除
    CTime, MTime time.Time
}

type SkillGroup struct {
    ID, Name, Namespace  string
    Owner, Business      string
    Department           string
    Metadata             map[string]string
    CTime, MTime         time.Time
}

type SkillVersion struct {
    ID         string
    SkillID    string
    Version    string      // 语义化版本
    Content    string      // 技能定义快照
    Active     bool        // 是否为当前活跃版本？
    CTime      time.Time
}

type SkillSubscription struct {
    ID         string
    SkillName  string
    Namespace  string
    ClientID   string      // 订阅的客户端
    CTime      time.Time
}
```

### 业务逻辑（`pkg/skill/skill.go`）

`SkillServer` 单例在 bootstrap 中初始化：
```go
func Initialize(s store.AIStore) error
func GetServer() (SkillServer, error)
```

`Server` 结构体直接使用 `store.AIStore`（写操作不经过中间缓存——读操作通过 [[cache-layer]] 缓存）。

### HTTP API（`plugin/apiserver/httpserver/skill/`）

**技能：**
- `POST /skill/v1/skills` — 创建技能（支持批量）
- `PUT /skill/v1/skills` — 更新技能（支持批量）
- `DELETE /skill/v1/skills` — 删除技能（支持批量）
- `GET /skill/v1/skills` — 查询技能（分页、可过滤）
- `GET /skill/v1/skills/all` — 获取全部技能
- `GET /skill/v1/skills/count` — 统计技能总数

**分组：**
- `POST /skill/v1/groups`
- `PUT /skill/v1/groups`
- `DELETE /skill/v1/groups`
- `GET /skill/v1/groups`

**版本：**
- `POST /skill/v1/versions`
- `DELETE /skill/v1/versions`
- `GET /skill/v1/versions`
- `PUT /skill/v1/versions/{id}/activate`

**订阅：**
- `POST /skill/v1/subscriptions`
- `DELETE /skill/v1/subscriptions`
- `GET /skill/v1/subscriptions`

### 缓存（`pkg/cache/ai/`）

AI 功能的四种缓存类型（详见 [[cache-layer]]）：
- `SkillCache` — 按 ID 和按 `namespace/name` 索引
- `MCPServerCache` — 按 ID 和按 `namespace/name` 索引
- `SkillVersionCache` — 按技能 ID 索引
- `SkillSubscriptionCache` — 按技能 key 和按客户端 ID 索引

所有缓存均使用增量更新模式（每隔 1 秒轮询 `mtime`）。

### 认证拦截器（`pkg/skill/interceptor/auth/`）

包装 `SkillServer` 以执行访问控制。检查调用方是否具有在目标命名空间中创建/更新/删除技能的权限。

---

## 与外部 AI 智能体的集成

典型使用场景：

```
AI 智能体（Claude、GPT 等）
    ↓ （MCP 协议）
Pole MCP 端点（/mcp）
    ↓
list_mcp_servers 工具 → 返回已注册的 MCP 后端
技能管理工具 → 技能的增删改查
    ↓
Pole 内部技能存储
```

这使智能体能够自主发现可用的工具和技能定义，无需硬编码配置。

## 相关页面

- [[domain-components]]
- [[storage]]
- [[cache-layer]]
- [[api-servers]]
