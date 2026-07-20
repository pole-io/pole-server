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

治理条件中的 `MatchString.value_type` 用于区分固定匹配与请求值采集。此前还曾提供 `VARIABLE` 读取运行机器环境变量，导致同一规则随实例环境变化，缺少跨数据面一致性和可移植性。

## 决策

统一值来源语义：

| 值来源 | 配置形态 | 运行时语义 |
|---|---|---|
| `TEXT` | `value` 填固定值 | 用固定值匹配当前请求参数 |
| `PARAMETER` | 新配置保持 `value` 为空 | 采集当前条件 `key` 对应的实际请求值，供后续治理动作消费 |

`VARIABLE` 从 specification 中删除，并保留枚举号 `2` 与名称 `VARIABLE` 不可复用。Console 不再展示该选项；SDK 遇到历史二进制枚举值 `2` 时必须拒绝匹配，不能回退成 `TEXT`，也不能读取进程环境变量。

为兼容既有 Rust SDK 配置，`PARAMETER + 非空 value` 暂时保留“读取同类型另一参数键”的旧语义；Console 新建和编辑请求参数时只生成 `PARAMETER + 空 value`。

请求参数采集按参数类型和键形成显式维度，不把敏感原值写入日志、规则或限流器键。动态限流器键使用采集值摘要；动态路由只允许目标标签引用流量条件中已经采集的同名请求参数。

## 能力边界

| 数据面 | 请求参数采集 | 动态本地限流 | 动态目标标签路由 |
|---|---:|---:|---:|
| Proxyless Rust SDK | 支持 | 支持 | 支持 |
| xDS | 当前不消费 | 当前不支持 | 当前不支持 |

API 资源的协议、接口、方法和路径不包含值来源；值来源只属于请求参数匹配条件。Console 必须展示数据面能力边界，不能把配置可保存等同于所有数据面均可执行。

## 验证

- Console 静态回归覆盖值来源选项和历史值归一化。
- Rust SDK 单元测试覆盖未知枚举值 fail closed，确保历史值 `2` 不读取机器环境也不参与匹配。
- Console 构建、Go 全量测试、知识库结构校验和真实页面交互作为交付门禁。

## 相关页面

- [[governance-rules]]
- [[patterns]]
- [[adr-governance-rule-unified-storage]]
