# 配置中心模块架构设计

## 1. 概述

配置中心模块 (`pkg/config`) 提供 KV 配置管理能力，支持配置分组、配置文件、配置发布、配置模板等功能。该模块实现了配置的版本管理、灰度发布和加密存储等高级特性。

## 2. 模块结构

```
pkg/config/
├── server.go                  # ConfigCenterServer 核心实现
├── default.go                 # 模块初始化
├── options.go                 # 配置选项
├── api.go                     # API 接口定义
├── common.go                  # 公共函数
├── utils.go                   # 工具函数
├── watcher.go                 # 配置监听中心
├── client.go                  # 客户端相关
├── config_file.go             # 配置文件 CRUD
├── config_file_group.go       # 配置分组 CRUD
├── config_file_release.go     # 配置发布
├── config_file_release_history.go # 发布历史
├── config_file_template.go    # 配置模板
├── config_chain.go            # 配置处理链
├── interceptor/
│   ├── register.go            # 拦截器注册
│   ├── auth/                  # 鉴权拦截器
│   │   ├── server.go
│   │   ├── config_file.go
│   │   ├── config_file_group.go
│   │   └── ...
│   └── paramcheck/            # 参数校验拦截器
│       ├── server.go
│       ├── config_file_check.go
│       └── ...
└── test_export.go             # 测试导出
```

## 3. 核心接口设计

### 3.1 ConfigCenterServer 接口

```go
// pkg/config/api.go
type ConfigCenterServer interface {
    // 配置分组管理
    CreateConfigFileGroup(ctx context.Context, req *apiconfig.ConfigFileGroup) *apimodel.Response
    UpdateConfigFileGroup(ctx context.Context, req *apiconfig.ConfigFileGroup) *apimodel.Response
    DeleteConfigFileGroup(ctx context.Context, req *apiconfig.ConfigFileGroup) *apimodel.Response
    QueryConfigFileGroups(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
    GetConfigFileGroup(ctx context.Context, namespace, group string) *apimodel.Response

    // 配置文件管理
    CreateConfigFile(ctx context.Context, req *apiconfig.ConfigFile) *apimodel.Response
    UpdateConfigFile(ctx context.Context, req *apiconfig.ConfigFile) *apimodel.Response
    DeleteConfigFile(ctx context.Context, req *apiconfig.ConfigFile) *apimodel.Response
    QueryConfigFiles(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
    GetConfigFile(ctx context.Context, namespace, group, fileName string) *apimodel.Response

    // 配置发布
    PublishConfigFile(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response
    DeleteConfigFileRelease(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response
    QueryConfigFileReleases(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
    GetConfigFileRelease(ctx context.Context, namespace, group, fileName, releaseName string) *apimodel.Response

    // 客户端配置获取
    GetConfigFileForClient(ctx context.Context, namespace, group, fileName string) *apimodel.Response
    WatchConfigFile(ctx context.Context, req *apiconfig.ClientConfigFileInfo) *apimodel.Response
}
```

### 3.2 Server 结构体

```go
// pkg/config/server.go
type Server struct {
    cfg                 *Config
    storage             store.Store
    fileCache           cacheapi.ConfigFileCache
    groupCache          cacheapi.ConfigGroupCache
    grayCache           cacheapi.GrayCache
    caches              cacheapi.CacheManager
    watchCenter         *watchCenter
    namespaceOperator   namespace.NamespaceOperateServer
    initialized         bool
    history             history.History
    cryptoManager       crypto.CryptoManager
    chains              *ConfigChains
    sequence            int64
}
```

## 4. 拦截器链模式

### 4.1 拦截器注册顺序

```go
// pkg/config/server.go
func GetChainOrder() []string {
    return []string{
        "auth",       // 鉴权拦截器
        "paramcheck", // 参数校验拦截器
    }
}
```

### 4.2 代理工厂模式

```go
type ServerProxyFactory func(cacheMgr cacheapi.CacheManager, s store.Store,
    pre ConfigCenterServer, cfg Config) (ConfigCenterServer, error)

func RegisterServerProxy(name string, factor ServerProxyFactory) error
```

### 4.3 调用链路

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

## 5. 配置分组管理

### 5.1 配置分组模型

```go
type ConfigFileGroup struct {
    Id          string
    Name        string
    Namespace   string
    Comment     string
    CreateTime  time.Time
    CreateBy    string
    ModifyTime  time.Time
    ModifyBy    string
    Business    string
    Department  string
    Metadata    map[string]string
}
```

### 5.2 创建配置分组流程

```go
func (s *Server) CreateConfigFileGroup(ctx context.Context, req *apiconfig.ConfigFileGroup) *apimodel.Response {
    // 1. 参数校验
    if errResp := checkRequestParamConfigFileGroup(req); errResp != nil {
        return errResp
    }

    // 2. 自动创建命名空间（如不存在）
    if errResp := s.createNamespaceIfAbsent(ctx, req); errResp != nil {
        return errResp
    }

    // 3. 检查分组是否已存在
    group, err := s.storage.GetConfigFileGroup(req.Namespace, req.Name)
    if group != nil {
        return api.NewConfigFileGroupResponse(apimodel.Code_ExistedResource, req)
    }

    // 4. 创建分组模型并持久化
    data := s.createConfigFileGroupModel(req)
    if err := s.storage.CreateConfigFileGroup(data); err != nil {
        return wrapperConfigFileGroupStoreResponse(req, err)
    }

    // 5. 记录操作历史
    s.RecordHistory(ctx, configFileGroupRecordEntry(ctx, req, data, types.OCreate))

    return api.NewConfigFileGroupResponse(apimodel.Code_ExecuteSuccess, data)
}
```

## 6. 配置文件管理

### 6.1 配置文件模型

```go
type ConfigFile struct {
    Id          string
    Name        string
    Namespace   string
    Group       string
    Content     string
    Format      string      // json, yaml, properties, xml, txt
    Comment     string
    CreateTime  time.Time
    CreateBy    string
    ModifyTime  time.Time
    ModifyBy    string
    Tags        []*ConfigFileTag
    Metadata    map[string]string
}
```

### 6.2 配置文件内容限制

```go
const (
    // 文件内容限制为 2w 个字符
    fileContentMaxLength = 20000
)

func (s *Server) initialize(...) error {
    if s.cfg.ContentMaxLength <= 0 {
        s.cfg.ContentMaxLength = fileContentMaxLength
    }
}
```

## 7. 配置发布机制

### 7.1 发布模型

```go
type ConfigFileRelease struct {
    Id          string
    Name        string
    Namespace   string
    Group       string
    FileName    string
    Content     string
    Format      string
    Comment     string
    CreateTime  time.Time
    CreateBy    string
    ModifyTime  time.Time
    ModifyBy    string
    Version     uint64
    Active      bool
    Tags        []*ConfigFileTag
    Metadata    map[string]string
}
```

### 7.2 发布流程

```go
func (s *Server) PublishConfigFile(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response {
    // 1. 获取配置文件
    file, errResp := s.getConfigFile(ctx, req.Namespace, req.Group, req.FileName)
    if errResp != nil {
        return errResp
    }

    // 2. 执行发布前处理链
    if errResp := s.chains.BeforeCreateFile(ctx, file); errResp != nil {
        return errResp
    }

    // 3. 创建发布记录
    release := s.createReleaseModel(req, file)
    if err := s.storage.CreateConfigFileRelease(release); err != nil {
        return wrapperConfigFileReleaseStoreResponse(req, err)
    }

    // 4. 通知监听客户端
    s.watchCenter.Notify(release)

    return api.NewConfigFileReleaseResponse(apimodel.Code_ExecuteSuccess, release)
}
```

## 8. 配置监听机制

### 8.1 WatchCenter

```go
// pkg/config/watcher.go
type watchCenter struct {
    cacheMgr cacheapi.CacheManager
    watchers *container.SyncMap[string, *watcherGroup]
}

type watcherGroup struct {
    mutex     sync.RWMutex
    watchers  map[string]*configWatcher // clientID -> watcher
}

type configWatcher struct {
    clientID   string
    configKey  ConfigFileKey
    version    uint64
    notifyCh   chan *ConfigFileChange
}
```

### 8.2 监听流程

```
Client                    Server                     WatchCenter
   │                        │                            │
   │──WatchConfigFile───────▶                            │
   │                        │──registerWatcher──────────▶│
   │                        │                            │
   │                        │    [Config File Updated]   │
   │                        │                            │
   │                        │◀──────Notify───────────────│
   │◀────ConfigChange───────│                            │
```

## 9. 配置处理链

### 9.1 ConfigFileChain 接口

```go
// pkg/config/config_chain.go
type ConfigFileChain interface {
    Init(svr *Server)
    BeforeCreateFile(ctx context.Context, file *conftypes.ConfigFile) *apimodel.Response
    BeforeUpdateFile(ctx context.Context, file *conftypes.ConfigFile) *apimodel.Response
    AfterGetFile(ctx context.Context, file *conftypes.ConfigFile) (*conftypes.ConfigFile, error)
    AfterGetFileRelease(ctx context.Context, release *conftypes.ConfigFileRelease) (*conftypes.ConfigFileRelease, error)
    AfterGetFileHistory(ctx context.Context, history *conftypes.ConfigFileReleaseHistory) (*conftypes.ConfigFileReleaseHistory, error)
}
```

### 9.2 内置处理链

```go
func (s *Server) initialize(...) error {
    s.chains = newConfigChains(s, []ConfigFileChain{
        &CryptoConfigFileChain{},      // 加密处理
        &ReleaseConfigFileChain{},     // 发布处理
    })
}
```

### 9.3 加密处理链

```go
type CryptoConfigFileChain struct {
    svr *Server
}

func (c *CryptoConfigFileChain) BeforeCreateFile(ctx context.Context, file *conftypes.ConfigFile) *apimodel.Response {
    // 检查是否需要加密
    // 对敏感内容进行加密处理
}

func (c *CryptoConfigFileChain) AfterGetFile(ctx context.Context, file *conftypes.ConfigFile) (*conftypes.ConfigFile, error) {
    // 解密内容
    file.OriginContent = file.Content
    decrypted, err := c.svr.cryptoManager.Decrypt(file.Content)
    file.Content = decrypted
    return file, nil
}
```

## 10. 灰度发布

### 10.1 灰度规则

配置文件支持按客户端标签进行灰度发布：

```go
type ConfigFileGrayRelease struct {
    Id          string
    Name        string
    Namespace   string
    Group       string
    FileName    string
    Content     string
    Labels      map[string]string  // 客户端标签匹配规则
    Active      bool
}
```

### 10.2 灰度匹配

```go
func (s *Server) GetConfigFileForClient(ctx context.Context, namespace, group, fileName string) *apimodel.Response {
    // 1. 获取客户端标签
    clientLabels := getClientLabels(ctx)

    // 2. 检查灰度规则
    grayRelease := s.grayCache.GetActiveGrayRelease(namespace, group, fileName)
    if grayRelease != nil && matchLabels(grayRelease.Labels, clientLabels) {
        return api.NewConfigFileResponse(apimodel.Code_ExecuteSuccess, grayRelease)
    }

    // 3. 返回正式发布
    release := s.fileCache.GetActiveRelease(namespace, group, fileName)
    return api.NewConfigFileResponse(apimodel.Code_ExecuteSuccess, release)
}
```

## 11. 配置版本管理

### 11.1 版本号生成

```go
func (s *Server) nextSequence() uint64 {
    return atomic.AddUint64(&s.sequence, 1)
}
```

### 11.2 发布历史

```go
type ConfigFileReleaseHistory struct {
    Id          string
    Name        string
    Namespace   string
    Group       string
    FileName    string
    Content     string
    Format      string
    Type        string      // create, update, delete, rollback
    CreateTime  time.Time
    CreateBy    string
    Version     uint64
}
```

## 12. 配置模板

配置模板支持变量替换：

```go
type ConfigFileTemplate struct {
    Id          string
    Name        string
    Namespace   string
    Content     string      // 包含 {{.Var}} 变量
    Variables   map[string]string
}
```

## 13. 性能优化设计

### 13.1 缓存优先读取

```go
func (s *Server) getActiveRelease(namespace, group, fileName string) *conftypes.ConfigFileRelease {
    // 优先从缓存读取
    release := s.fileCache.GetActiveRelease(namespace, group, fileName)
    if release != nil {
        return release
    }
    // 缓存未命中，从数据库读取
    release, _ = s.storage.GetConfigFileActiveRelease(namespace, group, fileName)
    return release
}
```

### 13.2 Watch 推送机制

- 客户端通过长连接监听配置变更
- 配置发布时主动推送变更通知
- 减少客户端轮询开销
