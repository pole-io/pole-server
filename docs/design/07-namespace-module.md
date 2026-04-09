# 命名空间模块架构设计

## 1. 概述

命名空间模块 (`pkg/namespace`) 提供多租户隔离能力，是 Pole Control Plane 的基础模块。所有资源（服务、配置、规则等）都属于某个命名空间。命名空间模块支持资源可见性控制和自动创建。

## 2. 模块结构

```
pkg/namespace/
├── server.go                  # Server 核心实现
├── default.go                 # 模块初始化
├── namespace.go               # 命名空间 CRUD
├── api.go                     # API 接口定义
├── interceptor/
│   ├── register.go            # 拦截器注册
│   ├── auth/                  # 鉴权拦截器
│   │   ├── server.go
│   │   └── resource_listener.go
│   └── log.go
└── log.go                     # 日志
```

## 3. 核心接口设计

### 3.1 NamespaceOperateServer 接口

```go
// pkg/namespace/api.go
type NamespaceOperateServer interface {
    // 命名空间管理
    CreateNamespace(ctx context.Context, req *apimodel.Namespace) *apimodel.Response
    UpdateNamespace(ctx context.Context, req *apimodel.Namespace) *apimodel.Response
    DeleteNamespace(ctx context.Context, req *apimodel.Namespace) *apimodel.Response
    GetNamespaces(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
    GetNamespace(ctx context.Context, name string) *apimodel.Response

    // 自动创建（内部使用）
    CreateNamespaceIfAbsent(ctx context.Context, namespace string) (*types.Namespace, error)
}
```

### 3.2 Server 结构体

```go
// pkg/namespace/server.go
type Server struct {
    storage               store.Store
    caches                *cache.CacheManager
    createNamespaceSingle *singleflight.Group
    cfg                   Config
}
```

## 4. 预定义命名空间

```go
const (
    // SystemNamespace 系统命名空间
    SystemNamespace = "pole-system"
    // DefaultNamespace 默认命名空间
    DefaultNamespace = "default"
    // ProductionNamespace 生产环境命名空间
    ProductionNamespace = "Production"
    // DefaultTLL 默认 TTL
    DefaultTLL = 5
)
```

## 5. 命名空间模型

### 5.1 Namespace 结构

```go
type Namespace struct {
    Name       string
    Comment    string
    CreateTime time.Time
    CreateBy   string
    ModifyTime time.Time
    ModifyBy   string
    Token      string
    Owner      string
    Visible    bool      // 是否对其他命名空间可见
}
```

### 5.2 资源可见性

命名空间支持资源可见性控制：

- `Visible=true`: 该命名空间下的资源对其他命名空间可见
- `Visible=false`: 该命名空间下的资源仅本命名空间可见

## 6. 核心操作

### 6.1 创建命名空间

```go
func (s *Server) CreateNamespace(ctx context.Context, req *apimodel.Namespace) *apimodel.Response {
    // 1. 参数校验
    if req.GetName() == "" {
        return api.NewNamespaceResponse(apimodel.Code_InvalidParameter, req)
    }

    // 2. 检查命名空间是否已存在
    ns, _ := s.storage.GetNamespace(req.GetName())
    if ns != nil {
        return api.NewNamespaceResponse(apimodel.Code_ExistedResource, req)
    }

    // 3. 创建命名空间
    data := s.createNamespaceModel(req)
    if err := s.storage.AddNamespace(data); err != nil {
        return wrapperNamespaceStoreResponse(req, err)
    }

    // 4. 记录操作历史
    s.RecordHistory(namespaceRecordEntry(ctx, req, data, types.OCreate))

    return api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, data)
}
```

### 6.2 自动创建命名空间

```go
func (s *Server) CreateNamespaceIfAbsent(ctx context.Context, namespace string) (*types.Namespace, error) {
    // 使用 singleflight 防止并发重复创建
    ret, err, _ := s.createNamespaceSingle.Do(namespace, func() (interface{}, error) {
        // 检查是否存在
        ns, _ := s.storage.GetNamespace(namespace)
        if ns != nil {
            return ns, nil
        }

        // 创建默认命名空间
        data := &types.Namespace{
            Name:       namespace,
            Comment:    "auto created",
            CreateTime: time.Now(),
        }
        if err := s.storage.AddNamespace(data); err != nil {
            return nil, err
        }
        return data, nil
    })

    return ret.(*types.Namespace), err
}
```

### 6.3 删除命名空间

```go
func (s *Server) DeleteNamespace(ctx context.Context, req *apimodel.Namespace) *apimodel.Response {
    // 1. 检查命名空间是否存在
    ns, _ := s.storage.GetNamespace(req.GetName())
    if ns == nil {
        return api.NewNamespaceResponse(apimodel.Code_NotFoundResource, req)
    }

    // 2. 检查是否有资源依赖
    // - 服务数量
    // - 配置分组数量
    // - 治理规则数量
    if hasResources(ns.Name) {
        return api.NewNamespaceResponse(apimodel.Code_NamespaceExistedResources, req)
    }

    // 3. 删除命名空间
    if err := s.storage.DeleteNamespace(req.GetName()); err != nil {
        return wrapperNamespaceStoreResponse(req, err)
    }

    return api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, req)
}
```

## 7. 缓存集成

### 7.1 命名空间缓存

```go
// 从缓存获取命名空间
func (s *Server) getNamespaceFromCache(name string) *types.Namespace {
    return s.caches.Namespace().GetNamespace(name)
}

// 获取可见命名空间列表
func (s *Server) getVisibleNamespaces(namespace string) []*types.Namespace {
    return s.caches.Namespace().GetVisibleNamespaces(namespace)
}
```

### 7.2 缓存接口

```go
type NamespaceCache interface {
    Cache
    // GetNamespace 根据 ID 获取命名空间
    GetNamespace(id string) *types.Namespace
    // GetNamespacesByName 根据名称列表获取命名空间
    GetNamespacesByName(names []string) []*types.Namespace
    // GetNamespaceList 获取所有命名空间
    GetNamespaceList() []*types.Namespace
    // GetVisibleNamespaces 获取指定命名空间可见的其他命名空间
    GetVisibleNamespaces(namespace string) []*types.Namespace
    // Query 查询命名空间
    Query(context.Context, *NamespaceArgs) (uint32, []*types.Namespace, error)
}
```

## 8. 拦截器

### 8.1 鉴权拦截器

```go
// pkg/namespace/interceptor/auth/server.go
type NamespaceAuthInterceptor struct {
    server    *Server
    policySvr auth.StrategyServer
}

func (i *NamespaceAuthInterceptor) CreateNamespace(ctx context.Context, req *apimodel.Namespace) *apimodel.Response {
    // 检查用户是否有创建命名空间的权限
    if errResp := i.checkPermission(ctx, req); errResp != nil {
        return errResp
    }
    return i.server.CreateNamespace(ctx, req)
}
```

## 9. 资源可见性控制

### 9.1 服务可见性

```go
// 获取跨命名空间可见的服务
func (s *Server) getVisibleServicesInOtherNamespace(ctx context.Context, name, namespace string) []*svctypes.Service {
    return s.caches.Service().GetVisibleServicesInOtherNamespace(ctx, name, namespace)
}
```

### 9.2 可见性规则

- 同一命名空间下的资源互相可见
- 标记为 Visible=true 的命名空间下的资源对其他命名空间可见
- 通过策略规则可以精细控制资源的访问权限

## 10. 操作历史记录

```go
func (s *Server) RecordHistory(entry *types.RecordEntry) {
    if history.GetHistory() == nil {
        return
    }
    if entry == nil {
        return
    }
    history.GetHistory().Record(entry)
}
```

## 11. 配置选项

```go
type Config struct {
    AutoCreate   bool                  // 是否自动创建命名空间
    Interceptors []string              // 拦截器列表
}
```

## 12. 与其他模块的关系

```
                    ┌─────────────┐
                    │  APIServer  │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│    Service    │  │    Config     │  │   GoverRule   │
│    Discover   │  │    Center     │  │   Manager     │
└───────┬───────┘  └───────┬───────┘  └───────┬───────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Namespace  │  ← 提供命名空间服务
                    └─────────────┘
```

所有业务模块在创建资源时都会：
1. 自动检查命名空间是否存在
2. 如果不存在且配置允许，自动创建命名空间
3. 通过命名空间进行资源隔离
