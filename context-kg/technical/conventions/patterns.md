---
title: 关键模式与约定
tags: [patterns, conventions]
links: [architecture, storage, cache-layer, auth-system, governance-rules, adr-governance-request-parameter-capture, adr-plugin-extension-registry]
updated: 2026-07-31
sources: 10
---

# 关键模式与约定

本文记录代码库中反复出现的关键模式。理解这些模式有助于快速读懂任意业务模块。整体架构见 [[architecture]]，存储层的具体用法见 [[storage]]，缓存层用法见 [[cache-layer]]，认证拦截器见 [[auth-system]]，治理规则产品语义见 [[governance-rules]]。

## 1. 插件注册模式

每个可扩展组件都通过实例化 Registry 注册 Factory。完整生命周期和兼容策略见
[[adr-plugin-extension-registry]]。

```go
// 第一步：在 apis/ 中定义接口
type SomePlugin interface {
    Name() string
    Initialize(c *ConfigEntry) error
    Destroy() error
    Type() PluginType
}

// 第二步：在 plugin/ 中实现接口
type myConcrete struct{}
func (m *myConcrete) Name() string { return "myPlugin" }

// 第三步：显式注册 Factory
func Register(registry *pluginapi.Registry) error {
    return apis.RegisterPluginFactory(registry, descriptor, func() (apis.Plugin, error) {
        return &myConcrete{}, nil
    })
}

// 第四步：由官方组合包集中装配
builtinplugins.Register(registry)
```

Factory 在同一运行 Registry 中只执行一次；`Clone()` 创建新的运行实例集。注册发生在
`Freeze()` 前，启动后禁止动态修改。Bootstrap 退出时，Registry 按实例解析顺序逆序调用
已解析插件的 `Destroy()`；未被配置选择的插件不会实例化或销毁。旧 `RegisterPlugin`
和 blank import 只用于兼容。

## 2. 单例业务服务模式

每个业务领域（service、config、namespace 等）均遵循以下模式：

```go
var (
    server       SomeServer       // 活跃服务端（可能被拦截器包装）
    originServer = &Server{}      // 具体实现
)

func Initialize(opts ...Option) error {
    // 配置 originServer
    server = originServer
    // 如果配置了拦截器则进行包装
    return nil
}

func GetServer() (SomeServer, error) {
    if !originServer.initialized {
        return nil, errors.New("not initialized")
    }
    return server, nil
}
```

## 3. 拦截器包装模式

认证和参数检查拦截器透明地包装真实服务端。详细说明见 [[auth-system]]。

```go
// 真实服务端
type Server struct { storage store.Store; ... }

// 认证包装器
type AuthServer struct {
    nextSvr  ServiceServer    // 包装真实服务端或其他拦截器
    userSvr  auth.UserServer
    policySvr auth.StrategyServer
}

func (s *AuthServer) CreateService(ctx context.Context, svc *service.Service) error {
    if err := s.checkAuth(ctx, svc.Namespace, "write:service"); err != nil {
        return err
    }
    return s.nextSvr.CreateService(ctx, svc)
}
```

链式结构在初始化时以逆序方式构建。

## 4. 增量缓存模式

所有缓存使用相同的增量更新方式。详细说明见 [[cache-layer]]，存储端查询见 [[storage]]。

```go
type SomeCache struct {
    storage    store.SomeStore
    items      map[string]*SomeType    // 内存索引
    lastMtime  time.Time
    firstUpdate bool
}

func (c *SomeCache) Update() error {
    mtime := c.lastMtime.Add(-5 * time.Second)  // 安全时间窗口
    items, err := c.storage.GetMoreItems(mtime, c.firstUpdate)
    for _, item := range items {
        if item.Flag == 1 {  // 软删除
            delete(c.items, item.ID)
        } else {
            c.items[item.ID] = item
            // 更新次级索引（按名称等）
        }
        if item.MTime.After(c.lastMtime) {
            c.lastMtime = item.MTime
        }
    }
    c.firstUpdate = false
    return nil
}
```

## 5. Singleflight 防并发重复创建

防止并发请求下资源被重复创建：

```go
type Server struct {
    createSingle singleflight.Group
}

func (s *Server) CreateService(ctx context.Context, req *Service) *Response {
    key := req.Namespace + "/" + req.Name
    result, err, _ := s.createSingle.Do(key, func() (interface{}, error) {
        return s.storage.CreateService(req)
    })
    // ...
}
```

使用场景：命名空间创建、服务创建。

## 6. Batch Controller 模式

适用于高吞吐量操作：

```go
type Controller struct {
    queue   chan *Task
    stopCh  chan struct{}
}

func (c *Controller) Submit(task *Task) {
    c.queue <- task
}

func (c *Controller) mainLoop() {
    batch := make([]*Task, 0, c.maxBatchCount)
    ticker := time.NewTicker(c.waitTime)
    for {
        select {
        case task := <-c.queue:
            batch = append(batch, task)
            if len(batch) >= c.maxBatchCount {
                c.flush(batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                c.flush(batch)
                batch = batch[:0]
            }
        }
    }
}
```

## 7. 软删除约定

所有持久化实体使用 `flag` 字段（详见 [[storage]]）：
- `flag = 0` — 活跃/可见
- `flag = 1` — 软删除

所有数据库查询过滤条件为 `WHERE flag = 0`。增量更新时，缓存在发现 `flag = 1` 的条目后将其从内存中移除。

## 8. 命名空间作用域资源命名

资源通过 `(namespace, name)` 组合唯一标识：
```go
key := namespace + "/" + name
```

作为按名称索引的次级缓存 key 使用。

## 9. 响应类型约定

所有业务方法返回基于 protobuf 的响应类型，来自 `pole-io/specification`：
- 单条操作：`*apimodel.Response`（包含 `code` + `info`）
- 批量写入：`*apimodel.BatchWriteResponse`（包含每条的处理结果）
- 批量查询：`*apimodel.BatchQueryResponse`（包含 `amount` + `size` + 数据列表）

## 10. Context 传递认证身份

认证身份通过 `context.Context` 在调用链中传递：
```go
// HTTP 中间件设置认证上下文
ctx = context.WithValue(ctx, authContextKey, &AuthContext{
    UserID: "user-123",
    Token:  "bearer-xxx",
})

// 业务逻辑读取
authCtx := ctx.Value(authContextKey).(*AuthContext)
```

## 11. 错误日志约定

错误日志使用组件名称前缀：
```go
log.Errorf("[Service] create service failed: namespace=%s name=%s err=%v", ns, name, err)
log.Infof("[Config] publish config file: namespace=%s group=%s name=%s", ns, group, name)
```

便于通过组件名称快速过滤日志。

## 12. 治理规则编辑器单栏交互约定

路由、泳道、限流、熔断、探测、无损、调用鉴权、流量镜像和流量 Mock 的规则编辑界面都以表单为唯一主内容，不展示动态实时 Spec、YAML 或 JSON 预览。新增或重构治理编辑器时必须遵守以下约定：

- 规则详情的 Tab 内容区负责受限滚动；编辑器 shell 自身使用自然高度，不能再叠加 `100vh`、固定高度或第二个滚动容器，以免内容稀少时把规则卡片拉满页面。
- 移除 Spec 后不保留空的双栏网格、固定预览列或相关样式；分段表单随内容收缩，长规则交给外层 Tab 内容区滚动。
- 可重复字段行的列宽必须按实际内容区而不是浏览器视口分配，并为操作按钮预留固定空间；任何宽度下都不能裁切最后一个字段或操作。
- 可搜索的选择控件必须基于输入文本过滤下拉候选项，不能只允许输入；来源服务等多选器需限制为可读宽度，而非占满分段正文。
- 操作按钮、StickyTool、发布弹窗等不能参与正文高度分配；需要放在正文布局流之外，避免挤压表单可用高度。
- 校验由保存前的业务校验、字段错误提示和失败通知承担，不以动态 Spec 预览代替校验。
- 共享请求条件必须保留 `value_type`，且只允许固定值和请求参数：固定值填写匹配值，请求参数采集当前键的实际值。不要读取运行机器环境变量作为治理规则值来源。数据面支持范围按 [[adr-governance-request-parameter-capture]] 显式展示，不能因规则可保存就宣称所有数据面均可执行。
- 验证不能只检查 CSS 声明；必须在真实 8080 页面确认接口行最后一列完整可见、来源服务能按关键字过滤，并验证长规则仍由 Tab 内容区滚动。

## 相关页面

- [[architecture]]
- [[storage]]
- [[cache-layer]]
- [[auth-system]]
- [[governance-rules]]
- [[adr-governance-request-parameter-capture]]
- [[adr-plugin-extension-registry]]
