---
title: 治理规则域
tags: [domain, governance, routing, ratelimit]
links: [namespace, architecture, storage, cache-layer, service-discovery]
updated: 2026-05-14
sources: 1
---

# 治理规则（`pkg/goverrule/`）

所有治理规则遵循相同的模式：增删改查 + 版本控制 + 灰度发布。整体架构见 [[architecture]]，命名空间依赖见 [[namespace]]，存储接口见 [[storage]]，缓存层见 [[cache-layer]]，与服务发现的关联见 [[service-discovery]]。

## 规则类型

| 规则 | 用途 |
|------|------|
| **路由规则** | 基于标签/元数据的流量路由 |
| **限流规则** | 针对服务或实例的速率限制 |
| **熔断规则** | 自动故障隔离 |
| **故障探测规则** | 主动健康探测 |
| **泳道规则** | 多泳道（环境隔离）路由 |
| **无损规则** | 优雅关闭/启动 |

所有规则支持：
- 版本控制（版本号追踪）
- 灰度发布（向部分用户发布）
- 细粒度权限（控制谁可以编辑哪些规则）

## `Server` 结构体关键字段

```go
type Server struct {
    config    Config
    storage   store.Store
    namespaceSvr    namespace.NamespaceOperateServer
    caches    CacheManager
    cmdb      plugin.CMDB              // 校验资源标签
    subCtxs   []*eventhub.Subscription
    emptyPushProtectSvs sync.Map
}
```

## 相关页面

- [[namespace]]
- [[architecture]]
- [[storage]]
- [[cache-layer]]
- [[service-discovery]]
