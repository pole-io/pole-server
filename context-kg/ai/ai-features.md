---
title: AI 原生功能：MCP 与 A2A Registry
tags: [ai, mcp, a2a]
links: [storage, cache-layer, api-servers]
updated: 2026-06-10
sources: 1
---

# AI 原生功能：MCP 与 A2A Registry

## 概览

AI 原生功能使 Pole 的服务注册中心和治理能力对 AI 智能体（LLM）可见，目前包含：

- **MCP Registry** — 注册和发现 MCP（模型上下文协议）服务器
- **A2A Agent Registry** — 注册和发现 A2A Agent Card 及其能力元数据

存储层接口见 [[storage]]，缓存机制见 [[cache-layer]]，HTTP API 端点见 [[api-servers]]。

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

### 缓存（`pkg/cache/ai/`）

- `MCPServerCache` — 按 ID 和按 `namespace/name` 索引，使用增量更新模式（每隔 1 秒轮询 `mtime`）。

---

## A2A Agent Registry

### 是什么

A2A（Agent2Agent）用于描述智能体之间的互操作协议。Pole control-plane 本轮只承载控制面 Registry 能力：
- 注册和维护 A2A Agent Card、访问接口、技能列表与治理元数据
- 按名称、命名空间、业务、部门、协议绑定、后端服务、streaming/push 能力和 skill tag 查询 Agent
- 通过缓存提供低延迟发现接口

不在 control-plane 内实现 A2A task proxy、message/send、message/stream、SSE 转发、push notification broker、task 状态机或 artifact 存储。这些属于数据面、网关或 Agent runtime。

### 领域模型（`apis/pkg/types/ai/a2a.go`）

```go
type A2AAgent struct {
    Id          string
    Name        string // 命名空间内唯一
    Namespace   string
    Description string
    Version     string

    PreferredEndpoint string
    ProtocolBindings  []string
    Interfaces        []*A2AAgentInterface
    Skills            []*A2AAgentSkill

    BackendType             string // service、url、external
    BackendServiceNamespace string
    BackendServiceName      string
    BackendURL              string

    Streaming         bool
    PushNotifications bool
    RawCardJson       string
}
```

`A2AAgentInterface` 表示 Agent Card 中可访问的协议端点；`A2AAgentSkill` 表示技能名称、描述、tags、input/output modes。当前仓库依赖的 `github.com/pole-io/specification` 尚未提供 A2A proto，因此 control-plane 先保留本地 Go 类型，后续可迁移到 specification。

### 存储与缓存

- `apis/store.A2AAgentStore` — 提供 Agent 增删改查、按 `namespace/name` 唯一查询、增量同步、技能查询和分页查询。
- MySQL 表：`a2a_agent`、`a2a_agent_interface`、`a2a_agent_skill`，采用 `flag=1` 软删除和 `mtime` 增量刷新。
- `pkg/cache/ai.A2AAgentCache` — 按 ID、`namespace/name` 和 namespace 建索引，并支持 skill tag 过滤。

### HTTP API（`plugin/apiserver/httpserver/aia2a/`）

接口前缀为 `/ai/a2a/v1`：
- `GET /agents` — 从缓存查询 Agent，支持分页与过滤
- `POST /agents` — 批量创建 Agent
- `PUT /agents` — 批量更新 Agent
- `POST /agents/delete` — 批量软删除 Agent
- `GET /agent/skills` — 按 `agent_id` 或 `agent_name + agent_namespace` 查询技能
- `GET /agents/{id}/card` — 返回原始 Agent Card JSON；无原始 card 时返回 Registry 记录

### Console 页面（`console/web/src/pages/AI/A2A/`）

A2A 页面与 MCP Registry 页面保持同一交互风格：筛选工具栏、主表格、抽屉编辑和详情抽屉。

页面操作流：
- 列表页按名称前缀、命名空间、协议绑定、skill tag、后端类型、streaming 和 push 能力筛选 Agent。
- 表格展示名称、命名空间、版本、协议、能力标签、技能数、后端绑定、拉取状态和操作时间。
- 新建/编辑抽屉分为“基础信息”“接入与能力”“技能”“来源与 Card”四段，覆盖 Agent Card 元数据、访问接口、技能声明和原始 card JSON。
- 详情操作只提供“查看 Agent Card”和“查看技能”，不提供 message/task/push 执行按钮，避免把数据面能力误放到控制面。

前端数据流：
- `services/a2a.ts` 对接 `/ai/a2a/v1` REST API。
- `modules/ai/a2a.ts` 管理列表、编辑对象、skills 和 card 详情状态。
- `modules/store.ts` 将 `aiA2A` reducer 注册到全局 store。
- `router/modules/ai.ts` 将 `/ai/a2a` 作为 AI 工具菜单的一等页面暴露。

---

## 与外部 AI 智能体的集成

典型使用场景：

```
AI 智能体（Claude、GPT 等）
    ↓ （MCP 协议）
Pole MCP 端点（/mcp）
    ↓
list_mcp_servers 工具 → 返回已注册的 MCP 后端
    ↓
Pole 内部 MCP 注册表
```

这使智能体能够自主发现可用的 MCP 后端，无需硬编码配置。

## 相关页面

- [[storage]]
- [[cache-layer]]
- [[api-servers]]
