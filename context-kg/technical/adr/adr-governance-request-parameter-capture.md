---
title: ADR：治理请求参数采集与动态消费
tags: [adr, governance, routing, ratelimit, sdk]
links: [governance-rules, patterns, adr-governance-rule-unified-storage]
updated: 2026-07-20
sources: 14
---

# ADR：治理请求参数采集与动态消费

## 状态

Accepted。

## 背景

治理条件中的 `MatchString.value_type` 用于区分固定匹配与请求值采集。不同编辑器和数据面曾把 `PARAMETER` 理解成“从另一个键读取值”，无法表达“采集当前请求中这个键的实际值并交给后续治理动作”。此前还曾提供 `VARIABLE` 读取运行机器环境变量，导致同一规则随实例环境变化，缺少跨数据面一致性和可移植性。

## 决策

统一值来源语义：

| 值来源 | 配置形态 | 运行时语义 |
|---|---|---|
| `TEXT` | `value` 填固定值 | 用固定值匹配当前请求参数 |
| `PARAMETER` | 新配置保持 `value` 为空 | 采集当前条件 `key` 对应的实际请求值，供后续治理动作消费 |

`VARIABLE` 从 specification 中删除，并保留枚举号 `2` 与名称 `VARIABLE` 不可复用。Console 不再展示该选项；SDK 遇到历史二进制枚举值 `2` 时必须拒绝匹配，不能回退成 `TEXT`，也不能读取进程环境变量。

为兼容既有 Rust SDK 配置，`PARAMETER + 非空 value` 暂时保留“读取同类型另一参数键”的旧语义；Console 新建和编辑请求参数时只生成 `PARAMETER + 空 value`。

请求参数采集按参数类型和键形成显式维度，不把敏感原值写入日志、规则或限流器键。动态限流器键使用采集值摘要；动态路由只允许目标标签引用流量条件中已经采集的同名请求参数。

## 治理动作消费

- 本地限流：把采集参数加入计数器维度，同一个规则会按不同实际值形成相互独立的限流器。
- 自定义路由：目标实例标签选择“请求参数”时，以同名采集参数的实际值匹配实例标签。
- 鉴权、镜像、Mock、泳道和其它共享请求条件：完整保存固定值或请求参数来源，只有对应数据面声明支持时才执行动态消费。
- API 资源的协议、接口、方法和路径不包含值来源；值来源只属于请求参数匹配条件。

## 能力边界

| 数据面 | 请求参数采集 | 动态本地限流 | 动态目标标签路由 |
|---|---:|---:|---:|
| Proxyless Rust SDK | 支持 | 支持 | 支持 |
| xDS | 当前不消费 | 当前不支持 | 当前不支持 |

Console 必须在选择“请求参数”时展示这一边界，不能把配置可保存等同于所有数据面均可执行。xDS 若要支持，需要分别设计 Envoy descriptor/action 与 dynamic metadata/subset 的转换，不在本次决策中伪装降级。

## 验证

- Console 静态回归覆盖共享编辑器、所有调用方映射、校验器和动态目标标签编辑。
- Rust SDK 单元测试覆盖当前键采集、同名目标标签路由和按采集值拆分本地限流计数器。
- Rust SDK 单元测试覆盖未知枚举值 fail closed，确保历史值 `2` 不读取机器环境也不参与匹配。
- Console 构建、Go 治理参数校验测试、知识库结构校验和真实页面交互作为交付门禁。

## 相关页面

- [[governance-rules]]
- [[patterns]]
- [[adr-governance-rule-unified-storage]]
