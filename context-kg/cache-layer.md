---
title: 缓存层
tags: [cache, performance]
links: [storage, architecture, ai-features]
updated: 2026-05-14
sources: 1
---

# 缓存层

## 设计理念

缓存层是一个内存读透缓存，位于业务逻辑层和数据库之间。它使服务发现、配置分发和规则评估能够以低延迟方式运行，无需在每次请求时都访问 MySQL。存储层接口见 [[storage]]，整体架构见 [[architecture]]，AI 缓存子包见 [[ai-features]]。

**位置：** `pkg/cache/`（实现），`apis/cache/`（接口）

## 缓存管理器

`apis/cache/types.go` 中的 `CacheManager` 由 19 种缓存子接口组合而成：

| 缓存名称 | 常量 | 用途 |
|---------|------|------|
| ServiceCache | `ServiceName` | 服务及别名 |
| InstanceCache | `InstanceName` | 服务实例 |
| RoutingConfigCache | `RoutingConfigName` | 路由规则 |
| RateLimitCache | `RateLimitConfigName` | 限流规则 |
| CircuitBreakerCache | `CircuitBreakerName` | 熔断规则 |
| FaultDetectorCache | `FaultDetectRuleName` | 故障探测规则 |
| NamespaceCache | `NamespaceName` | 命名空间 |
| ClientCache | `ClientName` | 已注册的客户端 |
| UserCache | `UserName` | 认证用户 |
| AuthStrategyCache | `StrategyName` | 认证策略/策略 |
| RoleCache | `RoleName` | 认证角色 |
| ConfigFileCache | `ConfigFileName` | 配置文件 |
| ConfigGroupCache | `ConfigGroupName` | 配置文件分组 |
| GrayConfigCache | `GrayConfigName` | 灰度发布配置 |
| SkillCache | `SkillName` | AI 技能 |
| MCPServerCache | `MCPServerName` | MCP 服务器 |
| SkillVersionCache | `SkillVersionName` | 技能版本 |
| SkillSubscriptionCache | `SkillSubscriptionName` | 技能订阅 |
| LaneRuleCache | `LaneRuleName` | 泳道规则 |

## 实现（`pkg/cache/cache.go`）

```go
type CacheManager struct {
    storage  store.Store
    caches   []Cache          // 所有缓存实例的有序列表
    needLoad *SyncSet         // 需要初始化的缓存集合
}
```

**刷新循环：**
```go
func (c *CacheManager) Run() {
    // 每隔 1 秒执行一次
    ticker := time.NewTicker(time.Second)
    for range ticker.C {
        for _, cache := range c.caches {
            cache.Update()  // 增量数据库查询
        }
    }
}
```

**每个缓存的增量更新模式：**
```go
func (c *skillCache) Update() error {
    // 查询数据库中自（lastMtime - 5 秒）以来被修改的记录
    skills, err := c.storage.GetMoreSkills(c.lastMtime.Add(-5*time.Second), c.firstUpdate)
    // 应用到内存 map
    for _, s := range skills {
        if s.Flag == 1 { // 软删除
            delete(c.skills, s.ID)
        } else {
            c.skills[s.ID] = s
        }
    }
    // 更新 lastMtime
}
```

**-5 秒**时间窗口用于防范应用服务器与数据库之间的时钟偏差。

## AI 缓存子包

```
pkg/cache/ai/
├── skill.go           # Skill 缓存（按 ID 和按 name+namespace 索引）
├── mcp.go             # MCPServer 缓存
├── version.go         # SkillVersion 缓存
└── subscription.go    # SkillSubscription 缓存
```

每种缓存均提供基于 ID 和基于名称的两种查找方式。详细的 AI 功能说明见 [[ai-features]]。

## 其他重要缓存

```
pkg/cache/service/    # ServiceCache + InstanceCache（数据量最大、最复杂）
pkg/cache/config/     # ConfigFileCache、ConfigGroupCache、GrayConfigCache
pkg/cache/auth/       # UserCache、StrategyCache、RoleCache
pkg/cache/rules/      # RoutingConfig、RateLimit、CircuitBreaker、FaultDetect、Lane
pkg/cache/namespace/  # NamespaceCache
pkg/cache/client/     # ClientCache
pkg/cache/gray/       # 灰度发布元数据
```

## OpenResourceCache

业务组件可以按需开启所需的缓存：

```go
// 在服务发现初始化中
cacheMgr.OpenResourceCache(
    cacheapi.ResourceWithCacheName(cacheapi.ServiceName),
    cacheapi.ResourceWithCacheName(cacheapi.InstanceName),
    cacheapi.ResourceWithCacheName(cacheapi.RoutingConfigName),
    // ...
)
```

这样可以避免在未使用某些功能时加载不必要的数据。

## 相关页面

- [[storage]]
- [[architecture]]
- [[ai-features]]
