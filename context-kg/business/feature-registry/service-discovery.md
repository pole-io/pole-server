---
title: 服务发现域
tags: [business, feature, service, healthcheck]
links: [namespace, architecture, storage, cache-layer, governance-rules]
updated: 2026-06-09
sources: 1
---

# 服务发现（`pkg/service/`）

核心服务网格功能。管理服务、实例、客户端和服务契约。整体架构见 [[architecture]]，命名空间依赖见 [[namespace]]，存储接口见 [[storage]]，缓存层见 [[cache-layer]]，治理规则见 [[governance-rules]]。

## `DiscoverServer` 接口

`DiscoverServer` 接口由以下子接口组合而成：
- `ServiceOperateServer` — 服务及别名的增删改查
- `InstanceOperateServer` — 实例的增删改查（注册/注销/更新）
- `ClientServer` — 客户端注册与管理
- `ServiceContractOperateServer` — 服务接口契约

## `Server` 结构体关键字段

```go
type Server struct {
    config          Config
    storage         store.Store
    namespaceSvr    namespace.NamespaceOperateServer
    caches          CacheManager
    bc              *batch.Controller       // 批量注册/心跳
    healthServer    *healthcheck.Server
    createServiceSingle singleflight.Group  // 防止并发重复创建服务
    subCtxs         []*eventhub.Subscription
    instanceChains  []InstanceEventHandler
    emptyPushProtectSvs sync.Map            // 每个服务的空推送保护
}
```

## Batch Controller（`pkg/service/batch/`）

对高频操作进行批处理：注册、注销、心跳。
- 可配置 `batchSize` 和 `concurrency`
- 达到批次上限或定时器触发时刷新

## 健康检查（`pkg/service/healthcheck/`、`plugin/service/healthchecker/`）

- 基于心跳的健康检查机制
- 插件：`heartbeat`（默认）
- 追踪 `lastHeartbeatTime`；不健康的实例在服务发现中不可见

## 空推送保护

针对单个服务：当所有实例消失（网络故障）时，可选择保留最后已知的实例集合。防止网络恢复期间的级联故障。

## 相关页面

- [[namespace]]
- [[architecture]]
- [[storage]]
- [[cache-layer]]
- [[governance-rules]]
