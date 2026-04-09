# 治理规则模块架构设计

## 1. 概述

治理规则模块 (`pkg/goverrule`) 提供服务流量治理能力，包括路由规则、限流规则、熔断规则、故障探测规则、泳道规则和无损规则等。该模块实现了服务治理规则的 CRUD 操作和发布管理。

## 2. 模块结构

```
pkg/goverrule/
├── server.go                  # Server 核心实现
├── default.go                 # 模块初始化
├── options.go                 # 配置选项
├── api.go                     # API 接口定义
├── router_rule.go             # 路由规则
├── ratelimit_rule.go          # 限流规则
├── circuitbreaker_rule.go     # 熔断规则
├── faultdetect_rule.go        # 故障探测规则
├── lane_rule.go               # 泳道规则
├── lossless.go                # 无损规则
├── releases.go                # 规则发布
├── client_v1.go               # 客户端相关
├── interceptor/
│   ├── register.go            # 拦截器注册
│   ├── auth/                  # 鉴权拦截器
│   │   ├── server.go
│   │   ├── router_rule.go
│   │   ├── ratelimit_rule.go
│   │   └── ...
│   └── paramcheck/            # 参数校验拦截器
│       ├── server.go
│       ├── route_rule.go
│       └── ...
└── log.go                     # 日志
```

## 3. 核心接口设计

### 3.1 GovernServer 接口

```go
// pkg/goverrule/api.go
type GovernServer interface {
    // 路由规则管理
    CreateRouterRules(ctx context.Context, reqs []*apitraffic.RouteRule) *apimodel.BatchWriteResponse
    UpdateRouterRules(ctx context.Context, reqs []*apitraffic.RouteRule) *apimodel.BatchWriteResponse
    DeleteRouterRules(ctx context.Context, reqs []*apitraffic.RouteRule) *apimodel.BatchWriteResponse
    GetRouterRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 限流规则管理
    CreateRateLimitRules(ctx context.Context, reqs []*apitraffic.Rule) *apimodel.BatchWriteResponse
    UpdateRateLimitRules(ctx context.Context, reqs []*apitraffic.Rule) *apimodel.BatchWriteResponse
    DeleteRateLimitRules(ctx context.Context, reqs []*apitraffic.Rule) *apimodel.BatchWriteResponse
    GetRateLimitRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 熔断规则管理
    CreateCircuitBreakerRules(ctx context.Context, reqs []*apifault.CircuitBreakerRule) *apimodel.BatchWriteResponse
    UpdateCircuitBreakerRules(ctx context.Context, reqs []*apifault.CircuitBreakerRule) *apimodel.BatchWriteResponse
    DeleteCircuitBreakerRules(ctx context.Context, reqs []*apifault.CircuitBreakerRule) *apimodel.BatchWriteResponse
    GetCircuitBreakerRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 故障探测规则管理
    CreateFaultDetectRules(ctx context.Context, reqs []*apifault.FaultDetectRule) *apimodel.BatchWriteResponse
    UpdateFaultDetectRules(ctx context.Context, reqs []*apifault.FaultDetectRule) *apimodel.BatchWriteResponse
    DeleteFaultDetectRules(ctx context.Context, reqs []*apifault.FaultDetectRule) *apimodel.BatchWriteResponse
    GetFaultDetectRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 泳道规则管理
    CreateLaneRules(ctx context.Context, reqs []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse
    UpdateLaneRules(ctx context.Context, reqs []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse
    DeleteLaneRules(ctx context.Context, reqs []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse
    GetLaneRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 无损规则管理
    CreateLosslessRules(ctx context.Context, reqs []*apiservice.LosslessRule) *apimodel.BatchWriteResponse
    UpdateLosslessRules(ctx context.Context, reqs []*apiservice.LosslessRule) *apimodel.BatchWriteResponse
    DeleteLosslessRules(ctx context.Context, reqs []*apiservice.LosslessRule) *apimodel.BatchWriteResponse
    GetLosslessRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 规则发布
    ReleaseRule(ctx context.Context, req *apimodel.Release) *apimodel.Response
    UnReleaseRule(ctx context.Context, req *apimodel.Release) *apimodel.Response
}
```

### 3.2 Server 结构体

```go
// pkg/goverrule/server.go
type Server struct {
    config               Config
    storage              store.Store
    namespaceSvr         namespace.NamespaceOperateServer
    caches               cacheapi.CacheManager
    cmdb                 cmdb.CMDB
    subCtxs              []*eventhub.SubscribtionContext
    emptyPushProtectSvs  *container.SyncMap[string, *time.Timer]
}
```

## 4. 路由规则

### 4.1 路由规则模型

```go
type RouterConfig struct {
    Id              string
    Name            string
    Namespace       string
    Service         string
    SourceService   *Service
    DestinationService *Service
    Priority        uint32
    Description     string
    Enable          bool
    Rules           []*RouteRule
    Revision        string
    CreateTime      time.Time
    ModifyTime      time.Time
}

type RouteRule struct {
    Source      *Source
    Destination *Destination
    Labels      map[string]string
}
```

### 4.2 路由规则类型

- **按源服务路由**: 根据调用方服务进行路由
- **按目标服务路由**: 根据被调方服务进行路由
- **按标签路由**: 根据请求标签进行路由
- **就近路由**: 根据地理位置进行路由

### 4.3 路由规则查询

```go
func (s *Server) GetRouterRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
    args := buildRoutingArgs(query)
    total, rules, err := s.caches.RoutingConfig().QueryRouterRules(ctx, args)
    // ...
}
```

## 5. 限流规则

### 5.1 限流规则模型

```go
type RateLimit struct {
    Id             string
    Name           string
    Namespace      string
    Service        string
    Method         string
    Labels         map[string]string
    Amounts        []*Amount
    Action         string      // reject, pass
    Disable        bool
    Revision       string
    CreateTime     time.Time
    ModifyTime     time.Time
}

type Amount struct {
    MaxAmount     uint64
    ValidDuration time.Duration
}
```

### 5.2 限流规则匹配

```go
func (s *Server) GetRateLimitRules(serviceKey svctypes.ServiceKey) ([]*rules.RateLimit, string) {
    return s.caches.RateLimit().GetRateLimitRules(serviceKey)
}
```

## 6. 熔断规则

### 6.1 熔断规则模型

```go
type CircuitBreakerRule struct {
    Id              string
    Name            string
    Namespace       string
    Service         string
    Method          string
    SourceService   string
    SourceNamespace string
    Destinations    []*DestinationSet
    Inbounds        []*DetectConfig
    Outbounds       []*DetectConfig
    Revision        string
    Enable          bool
    CreateTime      time.Time
    ModifyTime      time.Time
}

type DetectConfig struct {
    Policy         string
    TriggerCondition *TriggerCondition
    RecoverCondition *RecoverCondition
}
```

### 6.2 熔断策略类型

- **错误率熔断**: 根据错误率触发熔断
- **慢调用熔断**: 根据响应时间触发熔断
- **连续错误熔断**: 根据连续错误次数触发熔断

## 7. 故障探测规则

### 7.1 故障探测规则模型

```go
type FaultDetectRule struct {
    Id          string
    Name        string
    Namespace   string
    Service     string
    Method      string
    Targets     []*FaultDetectTarget
    Interval    time.Duration
    Timeout     time.Duration
    Revision    string
    CreateTime  time.Time
    ModifyTime  time.Time
}

type FaultDetectTarget struct {
    Type     string    // http, tcp, grpc
    Host     string
    Port     uint32
    Path     string
    Method   string
    Expected *ExpectedResponse
}
```

### 7.2 探测协议支持

- **HTTP**: HTTP/HTTPS 健康检查
- **TCP**: TCP 端口检查
- **gRPC**: gRPC 健康检查

## 8. 泳道规则

### 8.1 泳道规则模型

```go
type LaneGroup struct {
    Id          string
    Name        string
    Namespace   string
    Service     string
    Rules       []*LaneRule
    Revision    string
    CreateTime  time.Time
    ModifyTime  time.Time
}

type LaneRule struct {
    Lane       string
    Condition  *LaneCondition
    Priority   uint32
}
```

### 8.2 泳道规则用途

泳道规则用于全链路灰度，将特定流量路由到指定的服务实例组。

## 9. 无损规则

### 9.1 无损规则模型

```go
type LosslessRule struct {
    Id              string
    Name            string
    Namespace       string
    Service         string
    Type            string      // online, offline
    WarmUpDuration  time.Duration
    DelayDuration   time.Duration
    Revision        string
    CreateTime      time.Time
    ModifyTime      time.Time
}
```

### 9.2 无损上下线场景

- **无损上线**: 实例启动后延迟接收流量
- **无损下线**: 实例下线前等待连接关闭

## 10. 规则发布机制

### 10.1 发布流程

```go
func (s *Server) ReleaseRule(ctx context.Context, req *apimodel.Release) *apimodel.Response {
    // 1. 获取规则
    rule, errResp := s.getRule(ctx, req)
    if errResp != nil {
        return errResp
    }

    // 2. 创建发布记录
    release := s.createReleaseModel(req, rule)
    if err := s.storage.CreateRelease(release); err != nil {
        return wrapperReleaseStoreResponse(req, err)
    }

    // 3. 通知缓存更新
    eventhub.Publish(eventhub.RuleReleaseTopic, release)

    return api.NewReleaseResponse(apimodel.Code_ExecuteSuccess, release)
}
```

### 10.2 发布模型

```go
type Release struct {
    Id          string
    Name        string
    Namespace   string
    Service     string
    Type        string      // routing, ratelimit, circuitbreaker, etc.
    RuleId      string
    RuleName    string
    Revision    string
    CreateTime  time.Time
    CreateBy    string
}
```

## 11. 拦截器链

### 11.1 拦截器顺序

```go
// 每个规则类型都有独立的拦截器链
// 例如路由规则的拦截器：
func getRouterRuleChainOrder() []string {
    return []string{
        "auth",       // 鉴权
        "paramcheck", // 参数校验
    }
}
```

### 11.2 鉴权拦截器

```go
// pkg/goverrule/interceptor/auth/router_rule.go
type RouterRuleAuthInterceptor struct {
    server    *Server
    policySvr auth.StrategyServer
}

func (i *RouterRuleAuthInterceptor) CreateRouterRules(ctx context.Context, reqs []*apitraffic.RouteRule) *apimodel.BatchWriteResponse {
    // 检查每个规则的权限
    for _, req := range reqs {
        if errResp := i.checkPermission(ctx, req); errResp != nil {
            return errResp
        }
    }
    return i.server.CreateRouterRules(ctx, reqs)
}
```

## 12. 规则缓存集成

### 12.1 缓存访问

```go
func (s *Server) GetRouterRule(id, service, namespace string) ([]*apitraffic.RouteRule, string, error) {
    return s.caches.RoutingConfig().GetRouterRule(id, service, namespace)
}

func (s *Server) GetRateLimitRules(serviceKey svctypes.ServiceKey) ([]*rules.RateLimit, string) {
    return s.caches.RateLimit().GetRateLimitRules(serviceKey)
}
```

### 12.2 缓存更新通知

```go
// 规则发布后触发缓存更新
eventhub.Publish(eventhub.RuleReleaseTopic, release)

// 缓存层订阅发布事件
subCtx, _ := eventhub.Subscribe(eventhub.RuleReleaseTopic, func(event interface{}) {
    // 触发缓存增量更新
    cache.Update()
})
```

## 13. 规则查询参数

### 13.1 路由规则查询

```go
type RoutingArgs struct {
    Filter              map[string]string
    ID                  string
    Name                string
    Service             string
    Namespace           string
    SourceService       string
    SourceNamespace     string
    DestinationService  string
    DestinationNamespace string
    Enable              *bool
    Offset              uint32
    Limit               uint32
    OrderField          string
    OrderType           string
}
```

### 13.2 限流规则查询

```go
type RateLimitRuleArgs struct {
    Filter      map[string]string
    ID          string
    Name        string
    Service     string
    Namespace   string
    Disable     *bool
    Offset      uint32
    Limit       uint32
    OrderField  string
    OrderType   string
}
```

## 14. 推空保护机制

```go
// 记录开启推空保护的服务
emptyPushProtectSvs *container.SyncMap[string, *time.Timer]

// 检查推空保护
func (s *Server) checkEmptyPushProtect(serviceID string) {
    s.emptyPushProtectSvs.Store(serviceID, time.Now())
}
```

推空保护用于防止规则误删除导致的流量异常。
