# 缓存层架构设计

## 1. 概述

缓存层 (`pkg/cache`) 是 Pole Control Plane 的核心性能优化层，通过内存缓存减少对存储层的直接访问，提升服务发现和配置读取的性能。缓存层采用增量更新机制，定期从存储层同步数据变更。

## 2. 模块结构

```
pkg/cache/
├── cache.go              # CacheManager 核心实现
├── default.go            # 模块初始化与缓存注册
├── config.go             # 缓存配置
├── test_export.go        # 测试导出
├── base/
│   ├── types.go          # BaseCache 基础缓存实现
│   └── tools.go          # 缓存工具函数
├── service/              # 服务发现缓存
│   ├── default.go        # 服务缓存实现
│   ├── service.go        # 服务缓存接口
│   ├── instance.go       # 实例缓存
│   ├── instance_query.go # 实例查询
│   ├── service_query.go  # 服务查询
│   └── service_bucket.go # 服务分桶
├── config/               # 配置中心缓存
│   ├── default.go        # 配置缓存实现
│   ├── config_file.go    # 配置文件缓存
│   └── config_group.go   # 配置分组缓存
├── rules/                # 治理规则缓存
│   ├── default.go        # 规则缓存实现
│   ├── router_rule.go    # 路由规则缓存
│   ├── ratelimit_config.go # 限流规则缓存
│   ├── circuitbreaker.go # 熔断规则缓存
│   ├── faultdetect.go    # 故障探测规则缓存
│   ├── lane.go           # 泳道规则缓存
│   └── lossless.go       # 无损规则缓存
├── auth/                 # 鉴权缓存
│   ├── default.go        # 鉴权缓存实现
│   ├── user.go           # 用户缓存
│   ├── policy.go         # 策略缓存
│   ├── role.go           # 角色缓存
│   └── container.go      # 容器结构
├── namespace/
│   └── namespace.go      # 命名空间缓存
├── client/
│   └── client.go         # 客户端缓存
├── gray/
│   └── gray.go           # 灰度规则缓存
└── mock/
    └── cache_mock.go     # Mock 实现
```

## 3. 核心接口设计

### 3.1 Cache 接口

```go
// apis/cache/types.go
type Cache interface {
    // Initialize 初始化缓存
    Initialize(c map[string]interface{}) error
    // Update 更新缓存数据
    Update() error
    // Clear 清除缓存数据
    Clear() error
    // Name 缓存名称
    Name() string
    // Close 关闭缓存
    Close() error
}
```

### 3.2 CacheManager 接口

```go
// apis/cache/types.go
type CacheManager interface {
    // GetUpdateCacheInterval 获取缓存更新间隔
    GetUpdateCacheInterval() time.Duration
    // GetReportInterval 获取上报间隔
    GetReportInterval() time.Duration
    // GetTimeDiff 获取拉取 store 的时间偏移
    GetTimeDiff() time.Duration
    // GetCacher 获取缓存实现
    GetCacher(cacheIndex CacheIndex) Cache
    // RegisterCacher 注册缓存实现
    RegisterCacher(cacheIndex CacheIndex, item Cache)
    // OpenResourceCache 开启资源缓存
    OpenResourceCache(entries ...ConfigEntry) error

    // 各类缓存访问方法
    Service() ServiceCache
    Instance() InstanceCache
    RoutingConfig() RouterRuleCache
    RateLimit() RateLimitCache
    CircuitBreaker() CircuitBreakerCache
    FaultDetector() FaultDetectCache
    User() UserCache
    AuthStrategy() StrategyCache
    Namespace() NamespaceCache
    Client() ClientCache
    ConfigFile() ConfigFileCache
    ConfigGroup() ConfigGroupCache
    Gray() GrayCache
    Role() RoleCache
}
```

## 4. 缓存类型索引

```go
// apis/cache/types.go
type CacheIndex int

const (
    CacheService CacheIndex = iota      // 服务缓存
    CacheInstance                        // 实例缓存
    CacheRoutingConfig                   // 路由配置缓存
    CacheRateLimit                       // 限流规则缓存
    CacheLaneRule                        // 泳道规则缓存
    CacheLossLess                        // 无损规则缓存
    CacheCircuitBreaker                  // 熔断规则缓存
    CacheUser                            // 用户缓存
    CacheAuthStrategy                    // 鉴权策略缓存
    CacheNamespace                       // 命名空间缓存
    CacheClient                          // 客户端缓存
    CacheConfigFile                      // 配置文件缓存
    CacheFaultDetector                   // 故障探测缓存
    CacheConfigGroup                     // 配置分组缓存
    CacheServiceContract                 // 服务契约缓存
    CacheGray                            // 灰度规则缓存
    CacheRole                            // 角色缓存
    CacheLast                            // 缓存类型总数
)
```

## 5. 增量更新机制

### 5.1 BaseCache 基础实现

```go
// pkg/cache/base/types.go
type BaseCache struct {
    lock                  sync.RWMutex
    timeDiff              time.Duration
    single                *singleflight.Group
    firstUpdate           bool
    s                     store.Store
    lastFetchTime         int64
    lastMtimes            map[string]time.Time
    CacheMgr              cachetypes.CacheManager
    reportMetrics         func()
    lastReportMetricsTime time.Time
}
```

### 5.2 增量更新流程

```
┌─────────────────────┐
│  CacheManager.Start │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│    warmUp()         │  ← 首次全量加载
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  定时 Ticker (1s)   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  cache.Update()     │  ← 增量更新
│  lastFetchTime+N    │
└─────────────────────┘
```

### 5.3 增量查询实现

```go
func (bc *BaseCache) DoCacheUpdate(name string, executor func() (map[string]time.Time, int64, error)) error {
    // 获取当前存储时间
    curStoreTime, err := bc.s.GetUnixSecond(0)

    // 执行更新逻辑
    lastMtimes, total, err := executor()

    // 更新 lastFetchTime 用于下次增量查询
    bc.lastFetchTime = curStoreTime
    bc.lastMtimes = lastMtimes

    return nil
}
```

## 6. SingleFlight 防止缓存击穿

```go
// 使用 singleflight 确保同一时刻只有一个协程执行更新
func (bc *BaseCache) GetSingle() *singleflight.Group {
    return bc.single
}

// 在具体缓存实现中使用
func (c *ServiceCache) Update() error {
    _, err, _ := c.baseCache.GetSingle().Do(c.Name(), func() (interface{}, error) {
        return nil, c.realUpdate()
    })
    return err
}
```

## 7. 服务缓存实现

### 7.1 ServiceCache 接口

```go
type ServiceCache interface {
    Cache
    // GetNamespaceCntInfo 获取命名空间服务统计
    GetNamespaceCntInfo(namespace string) svctypes.NamespaceServiceCount
    // GetAllNamespaces 返回所有命名空间
    GetAllNamespaces() []string
    // GetServiceByID 根据 ID 查询服务
    GetServiceByID(id string) *svctypes.Service
    // GetServiceByName 根据名称查询服务
    GetServiceByName(name string, namespace string) *svctypes.Service
    // IteratorServices 迭代所有服务
    IteratorServices(iterProc ServiceIterProc) error
    // GetServicesByFilter 按条件过滤服务
    GetServicesByFilter(ctx context.Context, ...) (uint32, []*svctypes.EnhancedService, error)
    // ListServices 列出命名空间下的服务
    ListServices(ctx context.Context, ns string) (string, []*svctypes.Service)
}
```

### 7.2 InstanceCache 接口

```go
type InstanceCache interface {
    Cache
    // GetInstance 根据实例 ID 获取实例
    GetInstance(instanceID string) *svctypes.Instance
    // GetInstancesByServiceID 根据服务 ID 获取实例列表
    GetInstancesByServiceID(serviceID string) []*svctypes.Instance
    // GetInstances 获取服务实例数据
    GetInstances(serviceID string) *svctypes.ServiceInstances
    // GetInstancesCount 获取实例总数
    GetInstancesCount() int
    // DiscoverServiceInstances 服务发现获取实例
    DiscoverServiceInstances(serviceID string, onlyHealthy bool, consumer func(*svctypes.Instance))
}
```

## 8. 配置缓存实现

### 8.1 ConfigGroupCache

```go
type ConfigGroupCache interface {
    Cache
    // GetGroupByName 根据名称获取配置分组
    GetGroupByName(namespace, name string) *conftypes.ConfigFileGroup
    // GetGroupByID 根据 ID 获取配置分组
    GetGroupByID(id string) *conftypes.ConfigFileGroup
    // ListGroups 列出命名空间下的配置分组
    ListGroups(namespace string) ([]*conftypes.ConfigFileGroup, string)
    // Query 查询配置分组
    Query(args *ConfigGroupArgs) (uint32, []*conftypes.ConfigFileGroup, error)
}
```

### 8.2 ConfigFileCache

```go
type ConfigFileCache interface {
    Cache
    // GetGroupActiveReleases 获取分组下的活跃发布
    GetGroupActiveReleases(namespace, group string) ([]*conftypes.ConfigFileRelease, string)
    // GetActiveRelease 获取活跃发布
    GetActiveRelease(namespace, group, fileName string) *conftypes.ConfigFileRelease
    // GetActiveGrayRelease 获取活跃灰度发布
    GetActiveGrayRelease(namespace, group, fileName string) *conftypes.ConfigFileRelease
    // GetRelease 获取发布
    GetRelease(key conftypes.ConfigFileReleaseKey) *conftypes.ConfigFileRelease
}
```

## 9. 治理规则缓存

### 9.1 RouterRuleCache

```go
type RouterRuleCache interface {
    Cache
    // GetRouterRule 获取路由规则
    GetRouterRule(id, service, namespace string) ([]*apitraffic.RouteRule, string, error)
    // GetNearbyRouteRule 获取就近路由规则
    GetNearbyRouteRule(service, namespace string) ([]*apitraffic.RouteRule, string, error)
    // ListRouterRule 列出路由规则
    ListRouterRule(service, namespace string) []*rules.ExtendRouterConfig
    // QueryRouterRules 查询路由规则
    QueryRouterRules(context.Context, *RoutingArgs) (uint32, []*rules.ExtendRouterConfig, error)
}
```

### 9.2 RateLimitCache

```go
type RateLimitCache interface {
    Cache
    // IteratorRateLimit 遍历限流规则
    IteratorRateLimit(rateLimitIterProc RateLimitIterProc)
    // GetRateLimitRules 获取限流规则
    GetRateLimitRules(serviceKey svctypes.ServiceKey) ([]*rules.RateLimit, string)
    // QueryRateLimitRules 查询限流规则
    QueryRateLimitRules(context.Context, *RateLimitRuleArgs) (uint32, []*rules.RateLimit, error)
}
```

## 10. 鉴权缓存

### 10.1 UserCache

```go
type UserCache interface {
    Cache
    // GetAdmin 获取管理员信息
    GetAdmin() *authtypes.User
    // GetUserByID 根据 ID 获取用户
    GetUserByID(id string) *authtypes.User
    // GetUserByName 根据名称获取用户
    GetUserByName(name string) *authtypes.User
    // GetGroup 获取用户组
    GetGroup(id string) *authtypes.UserGroupDetail
    // IsUserInGroup 判断用户是否在组中
    IsUserInGroup(userId, groupId string) bool
}
```

### 10.2 StrategyCache

```go
type StrategyCache interface {
    Cache
    // GetPolicyRule 获取策略规则
    GetPolicyRule(id string) *authtypes.StrategyDetail
    // GetPrincipalPolicies 获取主体的策略
    GetPrincipalPolicies(effect string, p authtypes.Principal) []*authtypes.StrategyDetail
    // Hint 检查资源访问权限
    Hint(ctx context.Context, p authtypes.Principal, r *authtypes.ResourceEntry) apisecurity.AuthAction
}
```

## 11. 缓存配置

```yaml
cache:
  open: true
  resources:
    - name: service
      option:
        disableBusiness: false
    - name: instance
      option:
        disableBusiness: false
    - name: routingConfig
    - name: rateLimitConfig
    - name: circuitBreakerConfig
    - name: faultDetectRule
    - name: users
    - name: strategyRule
```

## 12. 性能优化设计

### 12.1 分桶存储

服务缓存采用分桶存储优化大规模服务场景：

```go
// pkg/cache/service/service_bucket.go
type ServiceBucket struct {
    buckets []*serviceBucket
}

type serviceBucket struct {
    lock     sync.RWMutex
    services map[string]*svctypes.Service
}
```

### 12.2 Revision 计算

```go
// 服务版本计算，用于客户端判断是否需要更新
func (w *ServiceRevisionWorker) GetServiceInstanceRevision(serviceID string) string {
    // 基于 CRC32 计算实例列表的 revision
}
```

### 12.3 指标上报

```go
func (bc *BaseCache) DoCacheUpdate(name string, executor func() ...) error {
    // ...
    if total >= 0 {
        metrics.RecordCacheUpdateCost(time.Since(start), name, total)
    }
    if bc.reportMetrics != nil {
        if time.Since(bc.lastReportMetricsTime) >= bc.CacheMgr.GetReportInterval() {
            bc.reportMetrics()
        }
    }
}
```

## 13. 缓存一致性保证

1. **时间戳对齐**: 使用存储层时间戳确保增量查询的一致性
2. **SingleFlight**: 防止并发更新导致的数据不一致
3. **读写锁**: 细粒度锁保护缓存数据
4. **定期刷新**: 默认 1 秒刷新一次，可配置
