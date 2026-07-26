---
title: 服务发现（`pkg/service/`）
tags: [business, feature, service, healthcheck]
links: [namespace, architecture, storage, cache-layer, governance-rules, adr-instance-active-healthcheck, adr-service-contract-reporting-and-visualization]
updated: 2026-07-26
sources: 6
---

# 服务发现（`pkg/service/`）

核心服务网格功能。服务名称是全局逻辑标识，同名服务在不同命名空间中表示同一服务的不同环境实例；每个环境实例独立管理实例、客户端状态和运行数据。整体架构见 [[architecture]]，命名空间依赖见 [[namespace]]，存储接口见 [[storage]]，缓存层见 [[cache-layer]]，治理规则见 [[governance-rules]]，实例主动检查见 [[adr-instance-active-healthcheck]]。

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

- SDK 注册实例使用 `heartbeat` 插件追踪 `lastHeartbeatTime`。
- Console 手工实例使用 `tcp` 或 `http` 插件由控制面主动探测，创建态不允许选择心跳。
- 不健康的实例在服务发现中不可见；协议、调度和持久化边界见 [[adr-instance-active-healthcheck]]。

## 空推送保护

针对单个服务：当所有实例消失（网络故障）时，可选择保留最后已知的实例集合。防止网络恢复期间的级联故障。

## 服务契约

服务契约支持 HTTP/OpenAPI、gRPC、Dubbo、Thrift 四类协议，采用 SDK gRPC 或 Agent/CI HTTP 推送，统一落库、缓存发现和 Console 展示。具体上报载荷、来源覆盖规则及兼容策略见 [[adr-service-contract-reporting-and-visualization]]。

## 相关页面

- [[namespace]]
- [[architecture]]
- [[storage]]
- [[cache-layer]]
- [[governance-rules]]
- [[adr-instance-active-healthcheck]]
- [[adr-service-contract-reporting-and-visualization]]
