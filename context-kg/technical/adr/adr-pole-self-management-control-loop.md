---
title: ADR：Pole 自身能力自动注册与自管理闭环
tags: [adr, ai, mcp, a2a, agent, prompt, automation]
links: [ai-features, adr-console-agent-resource-workbench, adr-system-configuration-control-plane, auth-system, architecture]
updated: 2026-07-26
sources: 22
---

# ADR：Pole 自身能力自动注册与自管理闭环

## 状态

Accepted，首个闭环已实现。Control Plane 自动把自身 MCP 与真实工具目录注册到 MCP Registry；`all` 模式下 Pole Agent 同时发布 A2A Agent Card、JSON-RPC 消息端点并注册到 A2A Registry；Console Agent 按 Registry 中的确定目标解析 MCP 地址。管理员保存 Agent Prompt/模型/工具策略后，系统自动探测并应用健康版本。

## 背景

此前 Pole 已有自身 MCP 协议端点、MCP Registry、A2A Registry 和 Console Agent，但它们是四条相邻而未闭合的链路：

- Control Plane 的 MCP 只能通过固定地址消费，没有自动登记自身或同步真实 `tools/list`。
- A2A Registry 只能保存 Agent Card 元数据，Pole Agent 没有可调用的 A2A 协议端点，也不会注册自己。
- Console Agent 能查看 Registry，却不会从 Registry 解析并连接自身 MCP。
- Agent Prompt 能版本化保存、测试和发布，但保存与应用仍需要两次人工操作。

目标不是让 LLM 任意修改自己，而是建立确定性的控制器：管理员拥有 desired state，系统身份负责观察、校验和收敛。

## 决策

### 1. 独立执行身份与权限边界

系统执行身份固定为 `pole-self-manager`，它不是管理员用户、没有可转交给模型的管理员 Token，也不进入普通 IAM 角色体系。仅 System Configuration 的系统管理员入口可以修改 Prompt、模型、工具白名单等 desired state；模型本身不能调用 System Configuration 写接口。

一次配置变更区分两个主体：

- `configured_by`：提交目标状态的管理员。
- `executed_by`：执行探测、发布或 Registry 收敛的 `pole-self-manager`。

Prompt 候选探测仍使用当前管理员的短期请求身份访问受保护 MCP，身份只存在于本次服务端调用中，不写入配置、审计详情或 Agent 上下文。

### 2. 采用确定性 Reconcile 深模块

对外只有一个收敛入口：

```go
type Manager interface {
    Reconcile(ctx context.Context, desired DesiredState) (Result, error)
}
```

模块内部负责自然键解析、稳定 ID、规范化、revision、差异比较、软删除复活、子项替换、幂等与审计。调用方不拼装 Create/Update/Delete 流程。

聚合根创建采用自然键幂等 claim，随后必须从主库强一致回读规范 ID，再绑定 Tool、Interface 和 Skill 子项。通用 Registry Update/Delete 的按 ID 保护同样使用主库点查；记录不可解析时失败关闭，避免副本延迟绕过系统投影保护。

受管资源固定在保留空间：

| 资源 | 自然键 | 来源 |
|---|---|---|
| Control Plane MCP | `pole-system/pole-control-plane` | 进程内 MCP `tools/list` 快照 |
| Pole Agent A2A | `pole-system/{agent_card_name}` | `/.well-known/agent-card.json` |

启动时必须先完成 MCP 首次收敛；之后每 30 秒全量 reconcile。A2A Card 尚未就绪时不阻断 MCP，下一轮自动补齐。漂移、手工软删除和工具/skill 变化均由下一轮恢复到目标状态。

### 3. MCP Registry 保存真实能力快照

HTTP MCP Server 通过进程内 `HandleMessage(tools/list)` 暴露快照，不通过网络回调自己。工具名、描述和输入 JSON Schema 进入 `mcp_server_tool`；稳定 ID 与 upsert 语义保证重复执行、软删除复活和子项变化不会制造重复记录。

Console Agent 不再把固定 MCP 地址视为事实来源。运行时按 `pole-system/pole-control-plane` 查询 Registry，要求恰好匹配一个 `address` backend，并只接受合法的 HTTP(S) URL；随后仍以当前用户身份建立 MCP 会话并执行 allowlist。静态 endpoint 只保留为未配置 Registry 时的自举回退。

### 4. Pole Agent 提供真实 A2A 数据面

Pole Agent 发布公开 Agent Card：

```text
GET /.well-known/agent-card.json
```

消息执行端点仍受 Console 登录态保护：

```text
POST /ai/agent/a2a/v1
```

端点实现同步 JSON-RPC `message/send`，并复用现有 `Agent.RunTurn`，因此 A2A 与 Console 对话使用同一 Prompt、模型、MCP 工具策略和确定性资源确认内核。Agent Card 按 A2A 0.3 声明 `preferredTransport: JSONRPC` 和 MIME 类型输入输出模式；JSON-RPC 区分 parse error、invalid request 和 notification。Registry 记录由真实 Card 投影，而不是维护第二份手写能力清单。

### 5. Prompt 保存后自动校验与应用

管理员保存 Agent desired revision 后，System Settings 在同一服务端流程中：

1. 保存管理员拥有的不可变草稿和 Secret 引用。
2. 构建候选 Agent。
3. 使用管理员请求身份探测模型与 MCP。
4. 探测成功后由 `pole-self-manager` 发布并原子替换当前运行快照。
5. 探测失败则保留 rejected 草稿和上一健康运行快照；周期协调器以系统身份自动重试模型，强一致核验 Registry 中唯一的受管 MCP 投影、地址和目标 endpoint，并校验真实 MCP 工具目录中的 allowlist，成功后发布。`all` 模式走进程内探针；`console` 分离模式调用 Pole Server 的窄只读 `/ai/mcp/v1/self-capabilities/probe`。远程请求使用独立部署的 `selfManagementProbeKey` 生成短期 HMAC 机器签名；`all` 模式未显式配置时可在启动边界从 System Secret 根密钥域分离派生一次，但只把派生子密钥注入 MCP Server。签名绑定 method、path、body hash 与一分钟时间窗；匿名/过期/篡改请求失败关闭，请求体限制为 64 KiB。该端点只返回就绪状态，不返回 Registry 记录、工具 Schema 或凭证。重试采用指数退避，最长 5 分钟；只有结构或凭证本身错误时才需要管理员修正 desired state。

这属于确定性配置协调，不授予 Pole Agent 修改自身 Prompt、模型、Secret 或工具权限的能力。内建安全 Prompt 仍由代码版本控制，页面只管理 operator instructions。

## 安全不变量

- 只有系统管理员路由能写 System Configuration；前端隐藏不是授权依据。
- `pole-self-manager` 只拥有自身 MCP/A2A 投影和 Agent 运行配置的窄执行能力。
- `pole-system` 下的受管 Registry 记录禁止通过通用 MCP/A2A 写接口变更；管理员修改 desired state，投影只由协调器写入。
- Agent Card 可以公开发现，消息执行、Registry 查询和 MCP 调用继续鉴权。
- Registry 地址必须精确解析一个受管自然键，拒绝空值、多值、非 address backend 和非 HTTP(S) 地址。
- Secret 明文、管理员 Token 和系统身份能力均不得进入模型上下文或 Registry。
- 分离部署的 Pole Server 只持有专用 `selfManagementProbeKey`，不得分发或缓存 Console Secret Store 根密钥。
- 候选版本未通过探测时不切换 current runtime；重复 reconcile 必须无副作用。

## 故障与一致性

- MCP 首次注册失败会阻止启动，避免 Console Agent 在无事实来源时进入半可用状态。
- 周期 reconcile 失败只记录告警，保留最近一次 Registry 状态并在下一轮重试。
- Prompt 首次探测失败后不保存管理员 Token；周期重试只使用系统身份执行模型凭证探测和自身 MCP 能力校验，发布审计仍记录 `pole-self-manager`。`all` 与 `console` 分离模式均不依赖管理员 Token；远程探针要求 `pole-self-manager` HMAC 机器身份，只暴露布尔就绪语义且不可用于通用 Registry 查询。
- A2A Card 启动期短暂不可用时跳过本轮 A2A，不阻断 Control Plane 主服务。
- Prompt 采用发布前探测和单进程原子指针切换；当前阶段尚未提供跨多实例 apply receipt，也不宣称已经解决“发布后才发生的外部依赖故障”自动回滚。

## 验证契约

- Create、drift、no-op、软删除复活与稳定 ID 单元测试。
- MCP `tools/list` 到 Registry Schema 的快照测试。
- A2A Card、匿名消息拒绝和认证消息执行路由测试。
- Registry 解析必须携带当前用户身份，并拒绝歧义或非法地址。
- Prompt 健康候选自动发布，失败候选保留 last-known-good。
- MySQL 子项 upsert 必须能恢复软删除记录。

## 相关页面

- [[ai-features]]
- [[adr-console-agent-resource-workbench]]
- [[adr-system-configuration-control-plane]]
- [[auth-system]]
- [[architecture]]
