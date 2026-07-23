---
title: ADR：治理规则统一存储与缓存更新
tags: [adr, governance, storage, cache, namespace]
links: [governance-rules, namespace, storage, cache-layer, architecture]
updated: 2026-07-23
sources: 6
---

# ADR：治理规则统一存储与缓存更新

## 状态

Accepted。

## 背景

治理规则当前按类型拆分为多组 MySQL 表和 cache 更新逻辑，例如路由、限流、熔断、故障探测、无损上下线、泳道组等规则都有相似的 CRUD、发布版本、active 切换和增量刷新路径。

这导致：

- 每类规则重复维护物理表和 SQL。
- 每类规则 cache 各自查库，和统一治理模型不匹配。
- 泳道规则使用 `lane_group`、`lane_rule`、`lane_group_release` 专用表，发布模型和其它治理规则不一致。
- 泳道组内规则数量限制不一致。

## 决策

治理规则底层统一为两张物理表：

| 表 | 职责 |
|----|------|
| `governance_rule` | 保存所有治理规则当前态，使用 `rule_type` 区分路由、限流、熔断、故障探测、无损上下线、泳道组等类型 |
| `governance_rule_release` | 保存所有治理规则发布快照、版本号、active 状态和发布类型 |

统一表的关键字段：

- `rule_type`：规则类型边界，所有读写、锁定、版本查询和 active 切换都必须携带。
- `namespace`：规则归属环境。规则逻辑身份为 `rule_type + name`，环境实例坐标为 `namespace + rule_type + name`。
- `service`：规则运行时作用对象的公共索引字段；不能用目标服务 namespace 代替规则归属环境。
- `enable` / `priority`：列表与排序中的高频公共字段，避免所有场景解析 JSON。
- `rule`：完整规则 JSON，是领域对象事实来源。
- `flag`：软删除标识，仍遵循 [[storage]] 中的 `flag=1` 删除语义。
- `mtime`：增量缓存同步依据，仍遵循 [[cache-layer]] 中的刷新模型。

新库初始化不再创建普通治理规则分表，也不再创建泳道专用表：

- `router_rule` / `router_rule_release`
- `ratelimit_rule` / `ratelimit_rule_release`
- `circuitbreaker_rule` / `circuitbreaker_rule_release`
- `fault_detect_rule` / `fault_detect_rule_release`
- `lossless_rule` / `lossless_rule_release`
- `lane_group` / `lane_rule` / `lane_group_release`

当前方案不考虑旧分表数据自动迁移。已有环境若要切换到统一模型，应使用新库初始化或单独人工迁移。

## Store 收敛原则

MySQL 层只保留一个统一 repository 负责治理规则 SQL：

```go
type GovernanceRuleRepository interface {
    CreateRule(...)
    UpdateRule(...)
    DeleteRule(...)
    LockRule(...)
    QueryRules(...)
    GetMoreRules(...)

    PublishRule(...)
    ActiveRelease(...)
    InactiveRelease(...)
    QueryReleases(...)
    GetMoreReleases(...)
}
```

实现原则：

- `plugin/store/mysql/governance_rule_store.go` 是治理规则统一表的唯一 SQL 入口。
- 路由、限流、熔断、故障探测、无损、泳道等领域 store 文件继续保留，职责收敛为领域对象和统一记录之间的转换。
- 领域 store 不再维护独立表名和独立 insert/update/release SQL。
- active 切换必须在事务内按 `rule_type + rule_id + release_type` 限定范围，避免不同规则类型互相影响。
- 上层 `pkg/goverrule` 的领域 API、鉴权、参数校验和 console 调用形态尽量保持不变。

## 规则归属环境

所有治理聚合根在 specification 中携带顶层 `namespace`，该字段只表达规则归属环境。caller、callee、target service、泳道入口和目的服务等嵌套 namespace 仍表达数据面的运行时作用域，两者不得互相推导或覆盖。

统一 repository 的名称查询、行锁、更新、删除、发布版本查询和 active 切换必须至少携带 `rule_type + namespace + name/id` 中可用的完整坐标。列表请求必须把 owner namespace 下推到 Store，不能先跨环境读取后只在 Console 过滤。

权限与删除保护遵循以下边界：

- 规则读写以顶层 namespace 构造授权资源上下文；运行时 target namespace 不代表规则管理权限。
- Namespace 删除前统计统一表中未删除的治理规则；存在任意类型规则时返回 `NamespaceExistedGovernanceRules`。
- Console 创建规则使用独立 `ruleNamespace`，不能复用服务上下文中的 caller/callee namespace。
- 历史 payload 未携带顶层 namespace 时可以被 protobuf 读取，但新建与更新路径必须写入明确环境，不能继续制造无归属规则。

## 泳道聚合规则

泳道采用 LaneGroup 聚合根模型：

```text
governance_rule.rule = LaneGroup JSON
LaneGroup.rules[] = LaneRule 子对象，最多 20 个
governance_rule_release.rule = 发布时完整 LaneGroup JSON 快照
```

约束：

- 不再持久化独立 `lane_rule` 行。
- `LaneRule` 的增删改查实际是事务内锁定并修改完整 `LaneGroup.rule` JSON。
- 创建、更新和单独新增 LaneRule 时统一校验 `len(group.Rules) <= 20`。
- 发布时保存完整 LaneGroup 快照，客户端发现时从 active release 重建泳道服务索引。

## Cache 统一更新源

治理规则 cache 分为两层：统一更新源和类型化 watcher。

```text
CacheManager ticker
  -> GovernanceRuleUpdateCache.Update()
    -> GetMoreRules(governance_rule)
    -> GetMoreReleases(governance_rule_release)
    -> group by rule_type
    -> RouterRuleCache.OnGovernanceRuleUpdate()
    -> RateLimitCache.OnGovernanceRuleUpdate()
    -> CircuitBreakerCache.OnGovernanceRuleUpdate()
    -> FaultDetectCache.OnGovernanceRuleUpdate()
    -> LosslessCache.OnGovernanceRuleUpdate()
    -> LaneCache.OnGovernanceRuleUpdate()
```

`GovernanceRuleUpdateCache` 是治理规则唯一访问 MySQL 的 cache。各类型 cache 保留当前查询 API 和内存索引结构，但不再各自查库。

watcher 接口形态：

```go
type GovernanceRuleChangeSet struct {
    RuleType apimodel.RuleRelease_RuleType
    Rules []*GovernanceRuleRecord
    Releases []*GovernanceRuleReleaseRecord
    RuleLastMtime time.Time
    ReleaseLastMtime time.Time
    FirstUpdate bool
}

type GovernanceRuleWatcher interface {
    RuleType() apimodel.RuleRelease_RuleType
    OnGovernanceRuleUpdate(change *GovernanceRuleChangeSet) error
    ClearGovernanceRules() error
}
```

更新策略：

- `GovernanceRuleUpdateCache` 每轮只查询统一规则表和统一发布表。
- 查询结果按 `rule_type` 分组成 `GovernanceRuleChangeSet`。
- fan-out 给对应 watcher 后，由 watcher 自行维护类型化索引。
- 任一 watcher 失败时，本轮统一更新失败，并且不推进 `lastFetchTime`，避免某类规则丢增量。
- 为兼容当前 `OpenResourceCache` 机制，类型化 cache 可以继续注册给业务查询；其 `Update()` 可调整为 no-op 或兼容入口，真正更新由统一源分发。

## 实施顺序

1. 新增 `governance_rule`、`governance_rule_release` DDL 和统一 repository 测试。
2. 普通规则 store 接入统一 repository，保持现有业务接口不变。
3. 泳道接入统一模型，LaneRule CRUD 改为读改写 LaneGroup JSON，统一上限 20。
4. 新增 `GovernanceRuleUpdateCache`，各类型 cache 实现 watcher。
5. 清理旧分表 DDL、兼容 schema 中旧列补齐逻辑和 store 中残留专用表 SQL。

## 回归测试重点

- 每种 `rule_type` 可创建、更新、查询、删除，且同名规则跨类型互不影响。
- 发布版本号按 `rule_type + rule_id` 自增。
- active 切换只影响同 `rule_type + rule_id + release_type` 范围。
- `GetMoreRules` 与 `GetMoreReleases` 可一次拉取多种类型并正确分组。
- LaneGroup 完整保存 `rules[]`，单条 LaneRule 增删改不会覆盖其它子规则。
- 第 21 个 LaneRule 返回参数错误。
- `GovernanceRuleUpdateCache` 一轮更新只调用统一 rule/release 查询，并正确分发到 watcher。
- 路由、限流、熔断、故障探测、无损、泳道的客户端发现返回结构保持不变。

## 相关页面

- [[governance-rules]]
- [[namespace]]
- [[storage]]
- [[cache-layer]]
- [[architecture]]
