---
title: ADR：RPC-first 治理范围与外部资源集成边界
tags: [adr, governance, rpc, scope, integration]
links: [governance-rules, terminology, domain-models, namespace, architecture, adr-governance-rule-unified-storage, adr-governance-request-parameter-capture, adr-managed-service-identity-authentication, adr-otel-observability-platform, testing]
updated: 2026-07-26
sources: 16
---

# ADR：RPC-first 治理范围与外部资源集成边界

## 状态

Accepted，作为当前阶段的产品和技术范围。未来只有满足本文的重新评审条件时才扩展。

## 决策

Pole 核心治理聚焦 HTTP、gRPC、Dubbo 等同步服务调用，不建设通用的消息、存储或任务运行时治理平台。

核心治理能力包括：

- 服务路由与泳道
- 限流
- 熔断与故障探测
- 调用鉴权
- 流量镜像
- 流量 Mock
- 无损上下线
- 面向 RPC 调用的稳定分桶和 A/B Test

Kafka、RocketMQ、Redis、MySQL 暂不支持请求级或消息级路由、灰度、镜像、Mock、限流和运行时鉴权。Pole 可以对这些外部资源提供非侵入式目录、健康、指标和原生管理集成，但不拦截其生产、消费、查询、事务或连接热路径。

任务调度暂不进入通用治理模型。只有 Pole 拥有统一 Job Worker/Agent、租约和执行协议时，才单独评估任务调度域。

## 为什么聚焦 RPC

RPC/HTTP 具有相对统一、低侵入的数据面 seam：

- 请求与响应生命周期清晰。
- Client/Server Interceptor 可以稳定接入。
- 方法、路径、Header、调用方身份和结果状态具有公共语义。
- 路由、限流、鉴权、镜像和 Mock 的效果可以被多种 RPC 数据面一致表达。
- SDK、Sidecar 和 Gateway 已经是实际存在的多个 Adapter，seam 不是假想抽象。

消息和存储不具备同等标准化基础。

### 消息复杂度

- Kafka 的 Partition、Consumer Group、事务消息与 RocketMQ 的 Queue、Tag、事务半消息并不等价。
- Producer 和 Consumer 是两条不同生命周期，无法用一套 RPC Request/Response 模型覆盖。
- 不同语言客户端的拦截、序列化和事务能力不一致。
- “灰度”可能表示换集群、换 Topic、影子生产、影子消费或 Consumer Group 迁移，没有单一语义。
- SDK Wrapper、Broker 插件或 MQ Proxy 都会显著增加侵入和运维成本。

### 存储复杂度

- MySQL 和 Redis 的连接、Session、事务、主从、Shard 与一致性语义不同。
- 事务开始后必须固定数据源和拓扑版本，不能随规则更新逐语句重选。
- 分库分表、Redis slot、读写分离和故障切主应由 Driver、Proxy 或数据库原生能力拥有。
- 错误路由可能造成数据损坏或丢失，风险显著高于普通 RPC 请求失败。
- Shadow Read、双写和迁移涉及幂等、补偿、复制延迟与审计，不能复用普通流量镜像。

在缺少公共执行语义时提前抽象，会得到一个大而浅的万能数据面内核：接口暴露 HTTP、Message、SQL、事务、Shard 和 Job 细节，调用方必须理解全部复杂度，违背 deep module 原则。

## 当前适用性矩阵

| 能力 | RPC/HTTP | Kafka/RocketMQ | Redis/MySQL | Job |
|---|---:|---:|---:|---:|
| 路由/泳道 | 核心范围 | 不做 | 不做 | 不做 |
| 灰度/A/B | 核心范围 | 不做 | 不做 | 不做 |
| 限流 | 核心范围 | 不做运行时拦截 | 不做运行时拦截 | 不做 |
| 运行时鉴权 | 核心范围 | 不做统一策略 | 不做统一策略 | 不做 |
| 镜像 | 核心范围 | 不做 | 不做双写/Shadow Read | 不做 |
| Mock/故障注入 | 核心范围 | 不做 | 不做 | 不做 |
| 熔断/故障探测 | 核心范围 | 仅展示外部健康 | 仅展示外部健康 | 不做 |
| 资源目录/指标 | 服务原生能力 | 可做非侵入集成 | 可做非侵入集成 | 后续评估 |

这个矩阵不是待填满的笛卡尔积。`不做` 是当前明确边界，不是功能遗漏。

## RPC 治理内核

现有九类规则继续共享 [[adr-governance-rule-unified-storage]] 定义的统一存储、发布和缓存机制，不引入多资源域公共 Rule Envelope。

治理内核应继续围绕服务调用加深：

```go
type ServiceGovernanceKernel interface {
    ValidateDraft(ctx context.Context, draft RuleDraft) (*ValidationReport, error)
    Publish(ctx context.Context, cmd PublishCommand) (*PublishReceipt, error)
    ResolveBundle(ctx context.Context, req ServiceBundleRequest) (*ServiceBundle, error)
}
```

该 Interface 隐藏：

- 规则类型化校验
- caller/callee 与服务资源解析
- 发布和回滚事务
- 数据面能力兼容判断
- Bundle 生成和冲突排序
- 灰度客户端选择

不在 Interface 中加入 Message、Storage 或 Job 参数。将来若出现真实新 seam，应单独设计 Adapter，再评估是否值得共享更高层生命周期。

## 服务数据面能力协商

RPC 治理内部仍需要机器可读的 Capability Profile，因为 Rust Proxyless、xDS/Envoy 和未来其它 SDK 的表达能力不同。

数据面上报：

- runtime 和版本
- 支持的服务治理规则类型
- matcher：Header、Query、Cookie、Path、AND/OR、动态参数
- action：目标权重、稳定 Hash、Mirror、Mock、Local/Distributed RateLimit
- last applied revision

发布前生成兼容报告：

| 状态 | 行为 |
|---|---|
| 完整支持 | 允许发布 |
| 缺少必需能力 | 阻止发布 |
| 只覆盖部分数据面 | 必须显式选择目标 cohort |
| 未知 | 默认阻止 |

不得把不支持的 OR 改成 AND，不得把稳定分桶降级为请求级随机，不得把动态目标标签当成固定标签。

数据面使用 `staged / active / last-known-good` 三槽原子应用 Service Bundle。Apply Receipt 区分：

- `APPLIED`
- `REJECTED`
- `PARTIAL`
- `OFFLINE/UNKNOWN`

规则写入数据库或生成 release 不等于数据面已经生效。

## RPC A/B Test

A/B Test 只面向具有稳定调用主体的 RPC/HTTP 场景：

- subject 可以来自受管服务身份、用户 ID 摘要、设备 ID 摘要或明确请求属性。
- 使用 `basis_points + hash_key + policy_uid + salt` 确定性分桶。
- 相同 subject 在规则版本不变时稳定进入同一 Variant。
- 数据面必须声明 `STABLE_HASH`，否则禁止发布。
- Exposure Event 和治理结果通过 [[adr-otel-observability-platform]] 进入观测管道。

治理路由只负责 Variant delivery。Experiment 生命周期、指标归因和统计显著性属于未来独立实验模块，不塞入 RouteRule。

## 外部资源集成

Kafka、RocketMQ、Redis、MySQL 可以作为 Console 中的外部资源集成，但不称为“已纳管的运行时治理”。

允许的非侵入能力：

- 集群和逻辑资源目录
- endpoint、版本、地域和标签
- 健康状态和容量摘要
- Consumer Lag、连接数、复制延迟等原生指标
- 外部管理页面跳转
- 经过单独评审的原生 ACL、Quota 或配置管理 Adapter

明确禁止：

- 在 Pole 通用 SDK 中解析消息正文、SQL 或 Redis value。
- 为了统一模型把 Topic、DataSource、Shard 注册成 Service/Instance。
- 通过 DNS 切换冒充事务安全存储路由。
- 在没有统一事务语义时提供存储双写或 MQ 事务灰度。
- Console 展示数据面没有真实执行者的 Route/Mirror/Mock 控件。

外部资源 Adapter 属于 true external seam。只有至少存在生产 Adapter 和可替代测试 Adapter，并且 Interface 能隐藏供应商差异时才引入；不为假想 Kafka/RocketMQ/Redis/MySQL 统一接口提前建层。

## 任务边界

Schedule、Job Run、Lease、Fencing Token、Executor Group 和重试状态属于任务调度域，不复用 RPC RouteRule。

重新评估任务治理前必须先具备：

- Pole 拥有的统一 Worker/Agent 协议。
- 持久化 Execution 和至少一次投递语义。
- 幂等键、租约、续租和 fencing。
- Worker 能力上报和 Apply Receipt。

在此之前，只允许从 Console 链接或展示外部调度平台信息。

## 现有实现的优先工作

在扩展资源类型之前，先让服务治理的声明与执行一致：

1. 修复并测试 caller/callee 到缓存索引的完整映射。
2. 对齐 Console 默认值、服务端校验、Rust SDK 与 xDS 的 matcher 语义。
3. 统一目标权重、随机百分比与稳定 Hash 的契约。
4. 建立数据面 Capability Profile 和发布前兼容门禁。
5. 补齐 Bundle 原子应用、ACK/NACK 和 Apply Receipt。
6. 用真实行为 E2E 验证路由、限流、鉴权、镜像和 Mock，不以 CRUD 成功代替。

## 对 specification 的影响

当前不新增：

- `ResourceDomain`
- `MESSAGE/STORAGE/JOB` Governance Rule
- 通用 Message/Storage/Job matcher
- 多资源域万能 `GovernancePolicy`

允许新增的协议只服务于 RPC 治理深度：

- Service Capability Profile
- Service Policy Bundle
- Apply Receipt
- 确定性 Fraction Match
- 现有九类规则的兼容修正

所有新 enum 零值必须是 `UNSPECIFIED`；protobuf 只做 additive change，删除字段永久 reserve 编号和名称。

## 重新评审条件

只有同时满足以下条件，才重新讨论某个外部资源的运行时治理：

1. 存在清晰、重复出现且原生平台无法满足的用户问题。
2. 至少两个真实执行 Adapter 证明语义可以标准化。
3. 接入不要求修改大量业务调用代码。
4. 能给出明确的 fail-open/fail-closed 和数据安全保证。
5. 有真实行为 E2E、故障注入和回滚证明。
6. 删除公共 Module 后复杂度会重新散落到多个调用方，证明该 Module 具备足够 depth。

在满足条件前，保持外部资源集成，不演进为运行时治理。

## 验收标准

- Console 只展示当前数据面真实支持的服务治理能力。
- 不支持的 matcher/action 在发布前被阻止。
- Service Bundle 可原子应用，失败后继续 last-known-good。
- 控制面可展示目标、已应用、拒绝和未知数据面覆盖率。
- RPC A/B 同一 subject 分桶稳定，并产生可关联 Exposure Event。
- Kafka、RocketMQ、Redis、MySQL 页面不得出现运行时灰度、路由、镜像或 Mock 的虚假入口。
- 知识库、Website 和产品文案统一使用 RPC-first 边界。

## 相关页面

- [[governance-rules]]
- [[terminology]]
- [[domain-models]]
- [[namespace]]
- [[architecture]]
- [[adr-governance-rule-unified-storage]]
- [[adr-governance-request-parameter-capture]]
- [[adr-managed-service-identity-authentication]]
- [[adr-otel-observability-platform]]
- [[testing]]
