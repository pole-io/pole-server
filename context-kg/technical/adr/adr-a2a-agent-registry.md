---
title: A2A Agent Registry 功能设计
tags: [adr, ai, a2a, mcp, registry]
links: [ai-features, api-servers, storage, cache-layer, auth-system, architecture]
updated: 2026-06-10
sources: 8
---

# A2A Agent Registry 功能设计

## 背景

A2A（Agent2Agent）是面向 Agent 之间协作的开放协议。官方项目说明它用于让不同框架、不同组织、不同服务端上的 Agent 发现彼此能力、协商交互方式、安全地处理长任务，并且不暴露内部状态、记忆或工具实现。

现有 Pole AI Native 能力已经包含 MCP Registry。MCP 解决“Agent 调工具和资源”的问题，A2A 解决“Agent 与 Agent 作为对等系统协作”的问题。两者应并列在 AI Native 功能域下，而不是互相替代。

官方资料：

- [a2aproject/A2A](https://github.com/a2aproject/A2A)
- [A2A Specification](https://a2a-protocol.org/latest/specification/)
- [A2A and MCP](https://a2a-protocol.org/latest/topics/a2a-and-mcp/)
- [Enterprise Implementation of A2A](https://a2a-protocol.org/latest/topics/enterprise-ready/)

## 决策

将 A2A 作为 AI Native 的第二个注册域引入，首期定位为 **A2A Agent Registry**：

- 管理 A2A Agent Card 的注册、发现、查询、更新和软删除。
- 按 `namespace/name` 唯一管理 Agent，与现有 MCP Server 的命名空间模型保持一致。
- 按 skill、tag、protocol binding、capability、backend service、business、department 等维度检索。
- 支持 Agent endpoint 直接地址和 Pole 服务发现后端两种来源。
- 保留完整 raw Agent Card JSON，结构化抽取高频查询字段，避免因 A2A 上游字段演进而频繁改表。
- 只做控制面注册与发现，不实现 A2A task runtime、任务状态机、artifact 持久化、SSE 转发和 push notification 代理；这些属于数据面或网关运行时能力。

## 功能边界

### 首期必须有

1. **Agent Card Registry**
   - 注册 A2A Agent 的 `name`、`description`、`version`、`provider`、`supported_interfaces`、`capabilities`、`default_input_modes`、`default_output_modes`、`skills`、`security` 和 raw card。
   - 校验 Agent Card 的最小必填字段：名称、描述、版本、至少一个 supported interface、capabilities、输入/输出模式、skills。
   - 允许从远端 `/.well-known/agent-card.json` 或用户提交 JSON 导入；`/.well-known/agent.json` 只作为旧路径兼容读取，新注册默认使用 `agent-card.json`。
   - 区分 public card 与 authenticated extended card：首期只缓存 public card；如果远端声明 extended card 能力，仅记录 capability 和受保护入口，不自动拉取敏感能力。

2. **Discovery API**
   - `GET /ai/a2a/v1/agents`：分页查询 Agent。
   - `POST /ai/a2a/v1/agents`：批量注册 Agent。
   - `PUT /ai/a2a/v1/agents`：批量更新 Agent。
   - `POST /ai/a2a/v1/agents/delete`：批量软删除 Agent。
   - `GET /ai/a2a/v1/agent/skills`：查询某个 Agent 的 skill 列表。
   - `GET /ai/a2a/v1/agents/{id}/card`：返回标准 Agent Card JSON。

3. **缓存**
   - 新增 `A2AAgentCache`，按 ID、`namespace/name`、namespace、skill tag 建索引。
   - 复用 `mtime > lastMtime - 5s` 的增量更新和 `flag=1` 软删除模式。

4. **安全元数据**
   - 记录 Agent Card 声明的 security schemes 和 security requirements。
   - 管理 API 复用现有 HTTP 鉴权和资源授权。
   - 不存储下游 Agent 的明文调用凭据；仅保存鉴权方案、scope、header 名称等元数据。

### 后续可选增强

- MCP 管理视图：可在 MCP Registry 侧新增 `list_a2a_agents`、`get_a2a_agent_card` 等只读工具，让 Agent 通过 MCP 查询 A2A Registry；这不是首期控制面闭环的必需项。
- Agent Card 自动导入：从 `/.well-known/agent-card.json` 拉取并审批入库。
- Agent 健康探测：周期性刷新 card 元数据和 endpoint 可达状态。

### 不属于 pole-control-plane 的范围

以下能力不放入 pole-control-plane 的 A2A Agent Registry 方案。它们如果需要实现，应归属数据面、网关或独立 A2A runtime：

- `message/send`、`message/stream`、`tasks/{id}`、`tasks/{id}:cancel`、`tasks/{id}:subscribe` 等 A2A 调用代理。
- SSE streaming 连接保持、flush、断线恢复、超时、取消和背压处理。
- Push notification webhook 托管、签名、重放防护、投递重试和死信记录。
- taskId/contextId 映射、任务状态历史、artifact 存储和运行时审计。

### 可选控制面增强

1. **Agent Health 与 Agent Card 自动刷新**
   - 周期性拉取远端 Agent Card，发现 version、skills、capabilities 变化。
   - 根据 endpoint 可达性、TLS、协议版本和鉴权错误生成健康状态。
   - 只更新注册表状态和发现元数据，不代理 A2A 任务流量。

## 数据模型

推荐先在 `github.com/pole-io/specification` 中新增 `api/v1/ai/a2a.proto`，再由本仓库引用生成类型。字段以协议中立为主：

```text
A2AAgent
- id
- name
- namespace
- visibility                    // public | private | internal
- description
- version
- protocol_version
- provider_organization
- provider_url
- documentation_url
- icon_url
- business
- department
- backend_type                  // service | address
- backend_service_namespace
- backend_service_name
- backend_address
- preferred_interface_url
- preferred_protocol_binding    // JSONRPC | GRPC | HTTP+JSON
- preferred_protocol_version
- streaming
- push_notifications
- extended_agent_card
- raw_card_json
- source_type                   // manual | well-known | curated
- source_url
- last_fetch_status
- last_fetch_time
- metadata
- flag
- ctime
- mtime

A2AAgentInterface
- id
- agent_id
- url
- protocol_binding
- protocol_version
- tenant
- flag
- ctime
- mtime

A2AAgentSkill
- id
- agent_id
- skill_id
- name
- description
- tags_json
- examples_json
- input_modes_json
- output_modes_json
- security_requirements_json
- flag
- ctime
- mtime
```

MySQL 表名建议为 `a2a_agent`、`a2a_agent_interface`、`a2a_agent_skill`。`raw_card_json` 用于兼容上游字段演进；interface 和 skill 拆表用于查询和缓存索引。安全方案可以首期放在 `raw_card_json`，若控制台需要单独筛选 OAuth/API key/mTLS，再拆 `a2a_agent_security_scheme`。

## 架构落点

### Store

在 `apis/store/ai.go` 中将 `AIStore` 扩展为：

```go
type AIStore interface {
    MCPServerStore
    A2AAgentStore
}
```

MySQL 实现参考 `plugin/store/mysql/mcp_server.go`，保持：

- 创建时自动生成 32 位无横杠 ID。
- `(namespace, name)` 唯一。
- 逻辑删除使用 `flag=1`。
- 查询和增量同步按 `mtime`。
- 写入 Agent 时在事务内同步 interfaces 和 skills。

实施文件：

- `apis/store/ai.go`：新增 `A2AAgentStore` 并组合进 `AIStore`。
- `plugin/store/mysql/a2a_agent.go`：新增 MySQL store。
- `plugin/store/mysql/default.go`：`stableStore` 增加 `*a2aAgentStore` 并在 `newStore()` 初始化。
- `plugin/store/mysql/scripts/pole_server.sql`：新增 `a2a_agent`、`a2a_agent_interface`、`a2a_agent_skill` 表。

### Cache

新增 `apis/cache/a2a.go` 和 `pkg/cache/ai/a2a_agent.go`：

- `GetA2AAgentByID(id)`
- `GetA2AAgentByName(name, namespace)`
- `GetA2AAgentsByNamespace(namespace)`
- `GetA2AAgentSkills(agentID)`
- `QueryA2AAgents(query)`

`CacheManager` 新增 `A2AAgentName` 和 `CacheA2AAgent`，初始化方式与 MCP Server 一致。A2A HTTP server 启动时按需 `OpenResourceCache(A2AAgentName)`。

实施文件：

- `apis/cache/ai.go`：新增 `A2AAgentCache` 接口。
- `apis/cache/types.go`：新增缓存名、缓存索引和 `CacheManager.A2AAgent()`。
- `pkg/cache/ai/default.go`：`NewAICaches` 返回 A2A 缓存。
- `pkg/cache/ai/a2a_agent.go`：实现 id/name/namespace/skill 索引和增量更新。
- `pkg/cache/default.go`：注册缓存名到索引，并在 cache manager 初始化时注册 A2A 缓存。
- `pkg/cache/cache.go`：新增 `A2AAgent()` accessor。

### API Server

新增 `plugin/apiserver/httpserver/aia2a/`，不要把 A2A 路由继续塞进 `aimcp/`：

```text
plugin/apiserver/httpserver/aia2a/
├── server.go          # REST 路由注册、cache 打开
├── agent.go           # REST 管理 API
└── log.go
```

配置建议：

```yaml
api-http:
  option:
    include:
      - aimcp
      - aia2a
```

基础路径建议：

- REST 管理：`/ai/a2a/v1`
- 不在 pole-control-plane 中预留 A2A 数据面代理路径；AgentInterface 只作为发现和治理元数据。

实施文件：

- `plugin/apiserver/httpserver/server.go`：导入 `aia2a`、增加 server 字段、初始化实例、在 `createRestfulContainer` 中接入 `aia2a`。
- `test/data/bootstrap/pole-apiserver.yaml` 和 `deploy/conf/pole-apiserver.yaml`：新增 `aia2a` API 配置。
- 根目录 `plugin.go` 不需要改动，前提是 A2A 仍作为现有 `httpserver` 插件的子包接入。

### Console

控制台新增 AI Native 下的 A2A Agents 页面：

- 列表：名称、命名空间、版本、协议、skills 数、是否 streaming、是否 push、backend、健康状态。
- 详情：Agent Card、interfaces、skills、security、raw JSON。
- 操作：注册、导入 Agent Card、编辑、删除、复制 Agent Card URL。

## 与 MCP Registry 的关系

MCP Registry 继续管理工具和资源端点。A2A Registry 管理 Agent 端点。两者的协同方式：

- Agent 可以通过 MCP 工具 `list_a2a_agents` 发现可协作 Agent。
- A2A Agent 的某些 skill 可以派生为 MCP Resource 或 Tool 描述，但这是适配视图，不改变 A2A 的任务协作语义。
- MCP Server 与 A2A Agent 都可以关联 Pole 服务发现后端，共享 backend selector 模型。

## 安全与治理

- 生产环境只接受 HTTPS A2A endpoint；HTTP 仅允许本地/测试环境显式开启。
- 自动拉取 Agent Card 必须做 SSRF 防护：禁止内网、环回、链路本地地址和非 allowlist 域名，限制重定向次数、响应体大小和超时。
- well-known 导入必须记录 `source_url`、拉取时间、状态码、校验错误和审批状态；外部 Agent Card 不能因为可访问就自动变成可信 Agent。
- Agent Card 中的 security schemes 只作为调用要求声明；调用凭据通过运行时上下文、外部 Secret 管理或调用方提供，不写入注册表明文字段。
- 管理 API 走现有 `auth-system`，新增资源类型建议为 `A2AAgent`、`A2AAgentSkill`。
- 对下游 Agent 的实际调用鉴权、用户 token exchange、push webhook 鉴权和 payload 安全检查不属于 Registry；这些策略应由数据面或调用方 runtime 承担，control-plane 只记录声明和治理元数据。

## 测试策略

首期测试覆盖：

- `apis/cache`：A2A cache 接口和索引查询。
- `pkg/cache/ai`：增量更新、软删除、namespace/name 查询、skill tag 查询。
- `plugin/store/mysql`：创建/更新/删除、事务写入 interfaces/skills、唯一约束、backend 校验、`GetMoreA2AAgents`。
- `plugin/apiserver/httpserver/aia2a`：REST 解析、Agent Card JSON 输出。
- `apis/...`：spec 类型编译和 Any/JSON 转换。

建议验证命令：

```bash
go test -count=1 ./plugin/store/mysql ./pkg/cache/ai ./plugin/apiserver/httpserver/aia2a ./apis/...
GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-a2a .
git diff --check -- apis pkg plugin context-kg
```

## 取舍

选择 Registry-first 的原因：

- 与 Pole 现有控制面定位一致：注册、发现、治理和可观测性，而不是接管业务 Agent 的运行时。
- 可复用 MCP Registry 的成熟 Store/Cache/API 模式，影响面可控。
- A2A 官方协议仍在扩展 discovery、client methods、streaming reliability 和 push notification；保留 raw card 能降低 schema 追随成本。

不做完整 Task Proxy 的原因：

- A2A 任务生命周期涉及 `SUBMITTED`、`WORKING`、`INPUT_REQUIRED`、`AUTH_REQUIRED`、`COMPLETED`、`FAILED`、`CANCELED`、`REJECTED` 等状态。
- SSE 和 push notification 需要连接管理、断线恢复、安全回调、投递语义和审计策略。
- 这些属于数据面能力，应该由网关、sidecar、agent runtime 或独立 A2A 数据面组件设计；pole-control-plane 只负责注册表、发现索引、配置元数据和管理 API。

## 相关页面

- [[ai-features]]
- [[api-servers]]
- [[storage]]
- [[cache-layer]]
- [[auth-system]]
- [[architecture]]
