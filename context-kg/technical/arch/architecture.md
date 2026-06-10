---
title: 架构 — 分层设计与插件系统
tags: [architecture, design]
links: [storage, cache-layer, auth-system, patterns]
updated: 2026-05-14
sources: 1
---

# 架构 — 分层设计与插件系统

## 架构分层

代码库采用严格的四层架构，每一层仅依赖其下方的层。详细的存储层设计见 [[storage]]，缓存层见 [[cache-layer]]，认证拦截器见 [[auth-system]]，代码模式见 [[patterns]]。

```
┌─────────────────────────────────────────────────────────┐
│  API 服务层 (plugin/apiserver/)                          │
│  HTTP · gRPC · xDS · Nacos · Apollo · Eureka           │
├─────────────────────────────────────────────────────────┤
│  业务逻辑层 (pkg/)                                       │
│  service · config · namespace · goverrule · admin       │
├─────────────────────────────────────────────────────────┤
│  缓存层 (pkg/cache/)                                    │
│  19 种缓存类型，每秒增量更新                              │
├─────────────────────────────────────────────────────────┤
│  存储层 (plugin/store/)                                 │
│  MySQL（默认）· mock（测试用）                           │
└─────────────────────────────────────────────────────────┘
```

横切关注点：认证拦截器、EventHub、可观测性（OTel 指标、历史记录、发现事件）

## 插件架构

所有可扩展组件均使用**注册与初始化**模式：

```go
// apis/plugin.go — 全局注册表
var pluginSet = make(map[PluginType]map[string]Plugin)

func RegisterPlugin(name string, plugin Plugin) { ... }
func GetPlugin(t PluginType, name string) (Plugin, bool) { ... }
```

插件类型（`apis/plugin.go` 中的 `PluginType` 枚举）：
- `PluginTypeStatis` — 统计信息
- `PluginTypeHistory` — 审计历史
- `PluginTypeDiscoverEvent` — 发现事件
- `PluginTypeRateLimit` — 限流
- `PluginTypeWhitelist` — IP 白名单
- `PluginTypeResourceAuth` — 资源鉴权
- `PluginTypeCMDB` — CMDB
- `PluginTypeApiServer` — API 服务端
- `PluginTypeCrypto` — 加密
- `PluginTypeStore` — 存储后端
- `PluginTypeHealthCheck` — 健康检查

**插件注册方式：** 每个插件包都有一个 `init()` 函数，调用 `apis.RegisterPlugin(name, impl)` 完成注册。根目录的 `plugin.go` 通过空导入所有插件包来触发其 `init()` 函数：

```go
// plugin.go（根目录）
import (
  _ "github.com/pole-io/pole-server/plugin/store/mysql"
  _ "github.com/pole-io/pole-server/plugin/apiserver/httpserver"
  // ... 其他所有插件
)
```

## Store 接口分解

`apis/store/store.go` 将 `Store` 定义为多个子接口的组合。完整的接口层次见 [[storage]]。

```go
type Store interface {
    DiscoverStore      // 服务、实例、契约、路由、限流、熔断、故障探测、泳道
    ConfigFileStore    // 配置文件、分组、发布、历史、灰度发布
    AdminStore         // 管理操作、分布式锁
    AuthStore          // 用户、用户组、策略、角色
    AIStore            // MCP 服务器、工具
    Transaction        // Begin()、Commit()、Rollback()
}
```

`AIStore` 子接口（位于 `apis/store/ai.go`）：
- `MCPServerStore` — 增删改查 + 增量查询 + 工具管理

## CacheManager 接口

`apis/cache/types.go` 将 `CacheManager` 定义为 19 种缓存子接口的组合。详细说明见 [[cache-layer]]。

```
ServiceName, InstanceName, RoutingConfigName, RateLimitConfigName,
CircuitBreakerName, FaultDetectRuleName, NamespaceName, ClientName,
UserName, StrategyName, RoleName, ConfigFileName, ConfigGroupName,
GrayConfigName, MCPServerName, LaneRuleName
```

`pkg/cache/cache.go` 中的 `CacheManager` 实现：
- 在后台 goroutine 中每隔 **1 秒** 刷新所有缓存
- 使用增量查询并设置 **-5 秒** 时间窗口（用于捕获最近的数据库变更）
- 支持通过 `OpenResourceCache()` 按需开启特定缓存
- `warmUp()` 触发初始缓存填充

## 拦截器/链式模式

每个业务层都使用装饰器/链式模式进行认证。详细设计见 [[auth-system]]。

```
pkg/service/interceptor/auth/   — 服务操作的认证链
pkg/config/interceptor/auth/    — 配置操作的认证链
pkg/goverrule/interceptor/auth/ — 治理操作的认证链
pkg/namespace/interceptor/auth/ — 命名空间操作的认证链
```

链式结构通过组件配置中的 `interceptors: ["auth"]` 进行配置。该模式采用"内层/外层"方式——外层 `Server` 包装真实实现，在委托之前先执行认证检查。

## EventHub

`pkg/common/eventhub/` 提供发布/订阅总线：
- 组件在资源变更时发布事件（实例注册、配置发布等）
- 订阅者（如健康检查、XDS 推送）异步响应
- 用于缓存失效信号和可观测性钩子

## Batch Controller

`pkg/service/batch/` 提供高吞吐量批处理能力：
- 处理注册、注销和心跳操作
- 可配置的批次大小和刷新间隔
- 在实例高频变动场景下减少数据库写入放大

## 请求流程示例（服务发现）

```
HTTP POST /naming/v1/instances
  → httpserver/discover 路由处理器
  → 认证拦截器（Token 检查、资源策略）
  → pkg/service.Server.RegisterInstance()
  → batch.Controller（积累、刷新）
  → store/mysql.InstanceStore.BatchAddInstances()
  → EventHub 发布（实例变更事件）
  → cache.InstanceCache 失效（下一个 1 秒 tick）
  → history 插件（审计记录）
```

## 相关页面

- [[storage]]
- [[cache-layer]]
- [[auth-system]]
- [[patterns]]
