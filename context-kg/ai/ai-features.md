---
title: AI 原生功能：MCP Registry
tags: [ai, mcp]
links: [storage, cache-layer, api-servers]
updated: 2026-05-16
sources: 1
---

# AI 原生功能：MCP Registry

## 概览

AI 原生功能使 Pole 的服务注册中心和治理能力对 AI 智能体（LLM）可见，目前包含：

- **MCP Registry** — 注册和发现 MCP（模型上下文协议）服务器

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
