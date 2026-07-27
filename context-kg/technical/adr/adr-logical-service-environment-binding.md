---
title: ADR：逻辑服务与环境服务显式关联
tags: [adr, service, namespace, domain-model, console]
links: [service-discovery, namespace, terminology, domain-models, business-rules]
updated: 2026-07-28
sources: 0
---

# ADR：逻辑服务与环境服务显式关联

## 状态

Accepted。

## 背景

现有领域语言把服务名同时用作运行时注册名称和跨环境逻辑身份，因此只有不同 Namespace
使用相同服务名时才能聚合。一旦开发、预发和生产采用不同运行时名称，环境服务记录仍然存在，
但控制面无法判断它们属于同一个业务服务；服务改名也会错误地改变跨环境身份。

## 决策

引入只属于控制面管理模型的 `LogicalService`。其稳定 ID 由控制面生成，用于聚合一个业务服务
在不同 Namespace 下的环境服务。`ServiceEnvironmentBinding` 显式关联：

`service_id` 是关联的权威键；`logical_service_id + namespace` 保证同一逻辑服务在每个环境
最多关联一个运行时服务。Namespace 与运行时服务名只作为管理视图快照，读取时以现存
`Service` 的最新值为准。

现有环境服务继续以 `namespace + runtime_service_name` 注册、发现和治理。SDK、数据面协议和
运行时服务寻址不感知 `logical_service_id`，也不要求客户端上报额外逻辑服务元数据。

跨环境关联由管理员在 Console 显式维护。控制面可以根据名称、业务归属、服务契约等信息提供
候选建议，但不得未经确认自动绑定。环境服务改名只调整关联，不改变逻辑服务身份；删除某个
环境服务也不自动删除逻辑服务。

## 考虑过的方案

- **按同名隐式聚合**：实现简单，但无法支持环境命名差异和服务改名。
- **由 SDK 上报逻辑服务 ID**：自动化程度高，但会污染注册发现契约，并把控制面组织模型泄漏给客户端。
- **控制面稳定 ID + 显式关联**：保留运行时协议兼容性，同时获得可靠的跨环境身份，采用此方案。

## 后果

- Console 主层级可以统一为“逻辑服务 → Namespace/环境服务 → 实例、契约和运行状态”。
- 现有服务数据无需迁移为新的运行时主键；逻辑服务和关联作为独立控制面资源演进。
- 未关联的环境服务仍可正常注册、发现和治理，只是不参与跨环境聚合。
- 跨环境比较、提升和审计必须基于显式关联，不能再用服务名相等作为权威判断。
- 升级时会幂等补齐有效逻辑服务名称唯一索引；若实验版本数据库已经存在重复有效名称，
  启动会 fail-fast，部署前必须先完成数据巡检和去重，服务端不得静默猜测应保留哪一条。

## 相关页面

- [[service-discovery]]
- [[namespace]]
- [[terminology]]
- [[domain-models]]
- [[business-rules]]
