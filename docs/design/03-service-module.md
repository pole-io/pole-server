# 服务模块架构设计

## 1. 概述

服务模块 (`pkg/service`) 是 Pole Control Plane 的核心模块之一，提供服务注册发现、实例管理、健康检查等核心能力。该模块实现了服务发现协议的核心业务逻辑。

## 2. 模块结构

```
pkg/service/
├── default.go           # 模块初始化与 Server 代理工厂
├── server.go            # Server 结构定义与基础方法
├── service.go           # 服务 CRUD 操作
├── instance.go          # 实例 CRUD 操作
├── service_alias.go     # 服务别名管理
├── service_contract.go  # 服务契约管理
├── event.go             # 事件处理
├── options.go           # 配置选项
├── log.go               # 日志
├── batch/               # 批量操作控制器
│   ├── config.go
│   └── batch.go
├── healthcheck/         # 健康检查
│   ├── config.go
│   ├── server.go
│   └── time_adjust.go
└── interceptor/         # 拦截器
    ├── register.go
    ├── auth/
    └── paramcheck/
```

## 3. 核心接口设计

### 3.1 DiscoverServer 接口

```go
// 服务发现服务端接口
type DiscoverServer interface {
    // 服务管理
    CreateServices(ctx context.Context, req []*apiservice.Service) *apimodel.BatchWriteResponse
    DeleteServices(ctx context.Context, req []*apiservice.Service) *apimodel.BatchWriteResponse
    UpdateServices(ctx context.Context, req []*apiservice.Service) *apimodel.BatchWriteResponse
    GetServices(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 实例管理
    CreateInstances(ctx context.Context, reqs []*apiservice.Instance) *apimodel.BatchWriteResponse
    DeleteInstances(ctx context.Context, req []*apiservice.Instance) *apimodel.BatchWriteResponse
    UpdateInstances(ctx context.Context, req []*apiservice.Instance) *apimodel.BatchWriteResponse
    GetInstances(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 服务别名管理
    CreateServiceAliases(...)
    DeleteServiceAliases(...)
    GetServiceAliases(...)

    // 服务订阅者查询
    GetServiceSubscribers(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
}
```

### 3.2 Server 结构体

```go
type Server struct {
    config              Config                          // 配置
    storage             storeapi.Store                  // 数据存储
    namespaceSvr        namespace.NamespaceOperateServer // 命名空间服务
    caches              cacheapi.CacheManager           // 缓存管理器
    bc                  *batch.Controller               // 批量写控制器
    healthServer        *healthcheck.Server             // 健康检查服务
    createServiceSingle *singleflight.Group             // 服务创建并发控制
    subCtxs             []*eventhub.SubscribtionContext // 事件订阅上下文
    instanceChains      []InstanceChain                 // 实例变化回调链
    emptyPushProtectSvs *container.SyncMap[string, time.Time] // 推空保护服务
}
```

## 4. 拦截器链模式

服务模块使用拦截器链实现横切关注点的分离：

```go
// 拦截器注册顺序
func GetChainOrder() []string {
    return []string{
        "auth",       // 鉴权拦截器
        "paramcheck", // 参数校验拦截器
    }
}
```

### 4.1 代理工厂模式

```go
type ServerProxyFactory func(pre DiscoverServer, s store.Store) (DiscoverServer, error)

// 注册代理工厂
func RegisterServerProxy(name string, factor ServerProxyFactory) error

// 初始化时构建拦截器链
func InitServer(ctx context.Context, opt *Config, opts ...InitOption) (*Server, DiscoverServer, error) {
    actualSvr := new(Server)
    // ... 初始化

    var proxySvr DiscoverServer
    proxySvr = actualSvr

    // 按顺序包装拦截器
    for i := range opt.Interceptors {
        factory := serverProxyFactories[opt.Interceptors[i]]
        afterSvr, _ := factory(proxySvr, actualSvr.storage)
        proxySvr = afterSvr
    }
    return actualSvr, proxySvr, nil
}
```

### 4.2 调用链路

```
API Request
    │
    ▼
┌─────────────────┐
│ Auth Interceptor │ ← 鉴权检查
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ ParamCheck Interceptor │ ← 参数校验
└────────┬────────────┘
         │
         ▼
┌─────────────────┐
│  Server Core    │ ← 核心业务逻辑
└─────────────────┘
```

## 5. 服务管理

### 5.1 服务创建流程

```go
func (s *Server) CreateService(ctx context.Context, req *apiservice.Service) *apimodel.Response {
    // 1. 自动创建命名空间（如不存在）
    if _, errResp := s.createNamespaceIfAbsent(ctx, req); errResp != nil {
        return errResp
    }

    // 2. 检查服务是否已存在
    service, err := s.storage.GetService(serviceName, namespaceName)
    if service != nil {
        return api.NewServiceResponse(apimodel.Code_ExistedResource, req)
    }

    // 3. 创建服务模型并持久化
    data := s.createServiceModel(req)
    if err := s.storage.AddService(data); err != nil {
        return wrapperServiceStoreResponse(req, err)
    }

    // 4. 记录操作历史
    s.RecordHistory(ctx, serviceRecordEntry(ctx, req, data, types.OCreate))

    return api.NewServiceResponse(apimodel.Code_ExecuteSuccess, out)
}
```

### 5.2 服务查询

支持多种过滤条件：

```go
var ServiceFilterAttributes = map[string]int{
    "id":          serviceFilter,
    "name":        serviceFilter,
    "namespace":   serviceFilter,
    "business":    serviceFilter,
    "department":  serviceFilter,
    "owner":       serviceFilter,
    "host":        instanceFilter,
    "port":        instanceFilter,
    "keys":        serviceMetaFilter,
    "values":      serviceMetaFilter,
}
```

## 6. 实例管理

### 6.1 实例创建流程

```go
func (s *Server) CreateInstance(ctx context.Context, req *apiservice.Instance) *apimodel.Response {
    // 1. 自动创建服务（如不存在）
    svcId, errResp := s.createWrapServiceIfAbsent(ctx, req)

    // 2. 填充 CMDB 位置信息
    s.packCmdb(ins)

    // 3. 选择同步或异步创建
    if s.bc == nil || !s.bc.CreateInstanceOpen() {
        return s.serialCreateInstance(ctx, svcId, req, ins)  // 同步创建
    }
    return s.asyncCreateInstance(ctx, svcId, req, ins)       // 异步批量创建
}
```

### 6.2 批量异步创建

为提高吞吐量，支持批量异步创建实例：

```go
func (s *Server) asyncCreateInstance(ctx context.Context, svcId string,
    req *apiservice.Instance, ins *apiservice.Instance) (*svctypes.Instance, *apimodel.Response) {

    allowAsyncRegis, _ := ctx.Value(types.ContextOpenAsyncRegis).(bool)
    future := s.bc.AsyncCreateInstance(svcId, ins, !allowAsyncRegis)

    rsp, err := future.Done()
    // ... 处理结果
}
```

### 6.3 实例事件类型

```go
const (
    EventInstanceOnline        InstanceEventType = iota  // 实例上线
    EventInstanceOffline                                 // 实例下线
    EventInstanceUpdate                                  // 实例更新
    EventInstanceTurnHealth                              // 变为健康
    EventInstanceTurnUnHealth                            // 变为不健康
    EventInstanceOpenIsolate                             // 开启隔离
    EventInstanceCloseIsolate                            // 关闭隔离
)
```

## 7. 事件驱动架构

### 7.1 EventHub 集成

```go
// 订阅实例事件
subCtx, err := eventhub.Subscribe(eventhub.InstanceEventTopic, eventHandler)

// 发布实例事件
func (s *Server) sendDiscoverEvent(event *svctypes.InstanceEvent) {
    _ = eventhub.Publish(eventhub.InstanceEventTopic, event)
}
```

### 7.2 实例事件结构

```go
type InstanceEvent struct {
    Id         string
    Namespace  string
    Service    string
    Instance   *apiservice.Instance
    EType      InstanceEventType
    CreateTime time.Time
    Metadata   map[string]string
}
```

## 8. 健康检查模块

### 8.1 健康检查配置

```go
type Config struct {
    Open        bool                              `yaml:"open"`
    ServiceName string                            `yaml:"service"`
    Checkers    map[int32]*CheckerConfig          `yaml:"checkers"`
    TimeAdjust   TimeAdjustConfig                 `yaml:"timeAdjust"`
}

type CheckerConfig struct {
    Name   string
    Config map[string]interface{}
}
```

### 8.2 健康检查器接口

```go
type HealthChecker interface {
    Name() string
    Type() int32
    Initialize(config map[string]interface{}) error
    BatchQuery(ctx context.Context, req *BatchQueryRequest) (*BatchQueryResponse, error)
    Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error)
    Set(ctx context.Context, req *SetRequest) (*SetResponse, error)
}
```

## 9. 推空保护机制

```go
// 记录开启推空保护的服务
emptyPushProtectSvs *container.SyncMap[string, time.Time]

// 检查并记录推空保护状态
func (s *Server) checkEmptyPushProtect(serviceID string) {
    s.emptyPushProtectSvs.Store(serviceID, time.Now())
}
```

## 10. 批量操作控制器

### 10.1 配置选项

```go
type Config struct {
    Open           bool `yaml:"open"`
    RegisterBatch  int  `yaml:"registerBatch"`
    RegisterMaxWait time.Duration `yaml:"registerMaxWait"`
    DeregisterBatch int `yaml:"deregisterBatch"`
}
```

### 10.2 批量操作类型

- `AsyncCreateInstance` - 异步批量创建实例
- `AsyncDeleteInstance` - 异步批量删除实例
- `AsyncUpdateInstance` - 异步批量更新实例

## 11. 服务别名

服务别名支持将一个服务映射到另一个服务：

```go
type ServiceAlias struct {
    Service           string
    Namespace         string
    Alias             string
    AliasNamespace    string
    Type              AliasType
    Owner             string
}

const (
    AliasTypeCluster AliasType = iota  // 集群别名
    AliasTypeNamespace                 // 命名空间别名
)
```

## 12. 性能优化设计

### 12.1 SingleFlight 防止并发重复创建

```go
// 使用 singleflight 确保服务创建的并发控制
key := fmt.Sprintf("%s:%s", simpleService.Namespace, simpleService.Name)
ret, err, _ := s.createServiceSingle.Do(key, func() (interface{}, error) {
    resp := s.CreateService(ctx, simpleService)
    return resp, nil
})
```

### 12.2 缓存优先读取

```go
func (s *Server) loadService(namespace string, svcName string) (*svctypes.Service, *apimodel.Response) {
    // 优先从缓存读取
    svc := s.caches.Service().GetServiceByName(svcName, namespace)
    if svc != nil {
        return svc, nil
    }
    // 缓存未命中，从数据库读取
    svc, err := s.storage.GetService(svcName, namespace)
    return svc, nil
}
```

## 13. 扩展点

### 13.1 InstanceChain 接口

```go
type InstanceChain interface {
    AfterUpdate(ctx context.Context, instances ...*svctypes.Instance)
}

// 注册实例链
func (s *Server) AddInstanceChain(chain ...InstanceChain) {
    s.instanceChains = append(s.instanceChains, chain...)
}
```

### 13.2 自定义拦截器

通过 `RegisterServerProxy` 注册自定义拦截器：

```go
func init() {
    service.RegisterServerProxy("custom", func(pre service.DiscoverServer, s store.Store) (service.DiscoverServer, error) {
        return &CustomInterceptor{server: pre}, nil
    })
}
```