---
title: 业务规则手册
tags: [business, rules]
links: [terminology, domain-models, namespace, service-discovery, config-center, governance-rules, adr-governance-rule-unified-storage]
updated: 2026-06-09
sources: 0
---

# 业务规则手册

## 跨域规则

- 所有业务资源必须归属命名空间，系统资源默认位于 `pole-system`。
- 服务、实例、配置、治理规则都采用软删除语义，外部查询默认只返回有效资源。
- 写操作直接落库，读路径优先通过 cache 获取内存视图。
- 发布类能力必须形成快照，避免后续草稿修改影响已发布版本。

## 服务发现规则

- 服务是发现与治理的核心归属对象。
- 实例健康状态、隔离状态和健康检查状态共同决定实例是否参与发现。
- 空推保护用于避免服务实例全量消失时触发级联故障。

## 配置中心规则

- 配置文件编辑态和发布态分离。
- 灰度发布通过标签或条件限制生效范围。
- Watch 通过客户端当前版本判断是否需要返回变更。

## 治理规则规则

- 治理规则支持版本控制、灰度发布和权限控制。
- 路由、限流、熔断、故障探测、无损上下线、泳道组都属于治理规则。
- 治理规则统一存储决策见 [[adr-governance-rule-unified-storage]]。

## 相关页面

- [[terminology]]
- [[domain-models]]
- [[namespace]]
- [[service-discovery]]
- [[config-center]]
- [[governance-rules]]
- [[adr-governance-rule-unified-storage]]
