---
title: ADR：Agent 与 MCP 逻辑定义及环境实例显式关联
tags: [adr, ai, agent, mcp, namespace, domain-model]
links: [ai-features, namespace, terminology, domain-models, business-rules, adr-a2a-agent-registry, adr-console-agent-resource-workbench, adr-logical-service-environment-binding, adr-system-namespace-kind]
updated: 2026-07-29
sources: 16
---

# ADR：Agent 与 MCP 逻辑定义及环境实例显式关联

## 状态

Accepted。

## 背景

MCP Server 与 A2A Agent 已按 `namespace + name` 登记，但同一能力在开发、测试和生产环境
中的记录没有稳定的跨环境身份。仅靠同名聚合无法支持运行名变化，也会把 Endpoint、后端服务、
凭证、Agent Card 和 Tool Catalog 等环境运行事实错误提升为全局定义。

Console Pole Agent 则是当前 Pole 安装实例的系统能力，不应在每个业务环境复制；但它当前允许
缺少明确 Namespace scope 的对话和工具调用，无法形成可靠的环境安全边界。

## 决策

引入两层模型：

- **AI 逻辑定义**：由控制面稳定 ID 标识，回答“这些环境实例是否代表同一个 Agent 或 MCP
  能力”。逻辑定义只用于管理聚合，不进入 MCP、A2A、SDK 或数据面寻址。
- **AI 环境实例**：现有 MCP Server 与 A2A Agent Registry 记录，以
  `namespace + runtimeName` 定位，承载 Endpoint、后端服务、Card/Tool 快照、版本、健康和
  Secret 引用。

管理员显式关联逻辑定义与环境实例。同名只能作为候选建议，迁移和运行时均不得自动合并。
同一逻辑定义在一个 BUSINESS Namespace 默认最多关联一个环境实例；运行名可以跨环境不同。

现有 Registry API、环境实例 ID、`namespace/name` 查询和授权资源 ID 保持兼容。逻辑定义与
环境绑定属于控制面管理 API，不要求旧 MCP protobuf 或数据面客户端感知。

## 环境与引用不变量

- `namespace` 是环境实例归属，不新增重复的 `environment` 字段。
- `pole-system` 中的 Pole MCP 与 Pole Agent A2A 投影属于 SYSTEM Namespace，不进入业务
  跨环境聚合，也不允许用户绑定到业务逻辑定义。
- Agent/MCP 绑定 Pole 服务时，稳定 `service_id` 是权威引用，Namespace 与服务名只作为展示
  快照。默认只允许绑定同 Namespace 的有效环境服务。
- 跨 Namespace 调用不是普通表单选项；未来如需开放，必须使用独立策略显式授权、审计并默认
  失败关闭。
- Namespace 或 Service 删除必须检查 AI 环境实例和后端引用，不能留下悬空关系。

## Pole Agent 调用作用域

Pole Agent 本体继续属于 `pole-system`，但每个 Turn 必须携带强类型 Namespace scope：

- 普通会话是单环境作用域，且只能包含一个可访问的 BUSINESS Namespace。
- 显式跨环境模式只允许读取、比较等只读能力；任何变更提案必须收敛到单一 Namespace。
- 资源上下文的 Namespace 必须属于 Turn scope。
- 服务端必须使用当前 Actor 身份解析可访问的 BUSINESS Namespace 目录，再接受 Turn scope；
  客户端传入的 Namespace 列表和 Prompt 均不是授权依据。
- 工具会话在调用 seam 强制注入或校验 Namespace；Prompt 提示不能替代服务端安全校验。
- A2A 调用必须显式传递相同 scope，不能从 Pole Agent 位于 `pole-system` 推断业务目标环境。

## 兼容与迁移

历史环境实例保持可用并标记为未关联，不按名称自动聚合。管理员可逐步创建逻辑定义并显式
关联。旧创建接口产生的环境实例继续有效；控制面兼容层可以为其建立独立逻辑定义，但不得与
其它环境的同名记录自动合并。

逻辑定义、环境绑定和后端稳定引用由启动期幂等迁移建立。任何自动回填只能基于唯一、有效的
稳定资源匹配；无法确定的历史引用保留待审计状态，不能静默猜测。

逻辑定义是低频控制面管理数据，首期管理 API 直接读取 Store，不进入秒级 Registry Cache；
现有 MCP/A2A 环境实例仍沿用原有 Cache。若后续 Definition 成为高频查询或运行时依赖，再以
独立指标证明后引入缓存，避免形成双写与失效语义。

## 后果

- Console 可以统一提供“逻辑定义 → 环境实例 → Tool/Card/运行状态”的跨环境视图。
- 环境实例改名不再改变跨环境身份，环境差异也不会污染全局定义。
- 逻辑定义管理权、环境实例管理权和实际调用权保持分离。
- 首期需要增加控制面 Store/API、数据库迁移、删除保护和 Agent scope 校验，但不需要修改
  旧 MCP/A2A 运行协议。

## 相关页面

- [[ai-features]]
- [[namespace]]
- [[terminology]]
- [[domain-models]]
- [[business-rules]]
- [[adr-a2a-agent-registry]]
- [[adr-console-agent-resource-workbench]]
- [[adr-logical-service-environment-binding]]
- [[adr-system-namespace-kind]]
