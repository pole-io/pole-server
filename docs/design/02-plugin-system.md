# 插件系统架构设计

## 1. 概述

Pole Control Plane 采用完全插件化的架构设计，核心功能通过插件机制实现可扩展性。插件系统定义在 `apis/` 目录，具体实现在 `plugin/` 目录。

## 2. 插件接口定义

### 2.1 基础插件接口

```go
// apis/plugin.go
type Plugin interface {
    Name() string                              // 插件名称
    Initialize(c *ConfigEntry) error           // 初始化
    Destroy() error                            // 销毁
    Type() PluginType                          // 插件类型
}
```

### 2.2 插件配置结构

```go
// 单个插件配置
type ConfigEntry struct {
    Name   string                 `yaml:"name"`
    Option map[string]interface{} `yaml:"option"`
}

// 插件执行链配置
type PluginChanConfig struct {
    Name    string                 `yaml:"name"`
    Option  map[string]interface{} `yaml:"option"`
    Entries []ConfigEntry          `yaml:"entries"`
}

// 全局插件配置
type Config struct {
    CMDB                 ConfigEntry      `yaml:"cmdb"`
    RateLimit            ConfigEntry      `yaml:"ratelimit"`
    History              PluginChanConfig `yaml:"history"`
    Statis               PluginChanConfig `yaml:"statis"`
    DiscoverStatis       ConfigEntry      `yaml:"discoverStatis"`
    ParsePassword        ConfigEntry      `yaml:"parsePassword"`
    Whitelist            ConfigEntry      `yaml:"whitelist"`
    MeshResourceValidate ConfigEntry      `yaml:"meshResourceValidate"`
    DiscoverEvent        PluginChanConfig `yaml:"discoverEvent"`
    Crypto               PluginChanConfig `yaml:"crypto"`
}
```

## 3. 插件类型

```go
type PluginType int32

const (
    _ PluginType = iota
    // -------- 可观测性插件 --------
    PluginTypeStatis         // 统计插件
    PluginTypeHistory        // 历史记录插件
    PluginTypeDiscoverEvent  // 发现事件插件

    // -------- 访问控制插件 --------
    PluginTypeRateLimit      // 限流插件
    PluginTypeWhitelist      // 白名单插件
    PluginTypeResourceAuth   // 资源鉴权插件

    // -------- 其他插件 --------
    PluginTypeCMDB           // CMDB插件
    PluginTypeApiServer      // API服务插件
    PluginTypeCrypto         // 加密插件
    PluginTypeStore          // 存储插件
    PluginTypeHealthCheck    // 健康检查插件
)
```

## 4. 插件注册机制

### 4.1 注册函数

```go
// apis/plugin.go
var pluginSet = make(map[PluginType]map[string]Plugin)

func RegisterPlugin(name string, plugin Plugin) {
    if _, exist := pluginSet[plugin.Type()]; !exist {
        pluginSet[plugin.Type()] = make(map[string]Plugin)
    }
    pluginSet[plugin.Type()][plugin.Name()] = plugin
}

func GetPlugin(t PluginType, name string) (Plugin, bool) {
    if plugins, exist := pluginSet[t]; exist {
        if plugin, ok := plugins[name]; ok {
            return plugin, true
        }
    }
    return nil, false
}
```

### 4.2 插件初始化流程

```go
// plugin.go - 通过 import 触发插件注册
import (
    _ "github.com/pole-io/pole-server/plugin/access_control/auth/policy"
    _ "github.com/pole-io/pole-server/plugin/access_control/auth/user"
    _ "github.com/pole-io/pole-server/plugin/access_control/ratelimit/token"
    _ "github.com/pole-io/pole-server/plugin/apiserver/httpserver"
    _ "github.com/pole-io/pole-server/plugin/store/mysql"
    // ... 更多插件
)
```

## 5. 插件目录结构

```
plugin/
├── access_control/           # 访问控制插件
│   ├── auth/
│   │   ├── user/            # 用户认证插件
│   │   │   ├── default.go   # 默认实现
│   │   │   ├── token.go     # Token 管理
│   │   │   └── inteceptor/  # 拦截器
│   │   └── policy/          # 策略鉴权插件
│   │       ├── default.go
│   │       ├── policy.go
│   │       └── role.go
│   ├── ratelimit/
│   │   └── token/           # Token 限流插件
│   │       ├── api_limit.go
│   │       ├── limiter.go
│   │       └── resource_limiter.go
│   └── whitelist/
│       └── ip/              # IP 白名单插件
│           └── ip_whitelist.go
├── apiserver/               # API 服务器插件
│   ├── apolloserver/        # Apollo 协议
│   ├── consulserver/        # Consul 协议
│   ├── eurekaserver/        # Eureka 协议
│   ├── grpcserver/          # gRPC 协议
│   ├── httpserver/          # HTTP REST API
│   ├── nacosserver/         # Nacos 协议
│   └── xdsserverv3/         # XDS v3 协议
├── crypto/                  # 加密插件
│   ├── aes/                 # AES 加密
│   └── rsa/                 # RSA 加密
├── observability/           # 可观测性插件
│   ├── discoverevent/       # 发现事件
│   │   ├── logger/          # 日志实现
│   │   └── rds/             # 数据库实现
│   ├── history/             # 历史记录
│   │   ├── logger/
│   │   └── rds/
│   └── statis/              # 统计
│       ├── logger/
│       └── prometheus/
├── cmdb/
│   └── memory/              # 内存 CMDB
├── service/
│   └── healthchecker/
│       └── heartbeat/       # 心跳检查
└── store/
    └── mysql/               # MySQL 存储
```

## 6. 插件实现示例

### 6.1 API Server 插件

```go
// apis/apiserver/apiserver.go
type Apiserver interface {
    GetProtocol() string                                              // 协议名
    GetPort() uint32                                                  // 监听端口
    Initialize(ctx context.Context, option map[string]interface{},
               api map[string]APIConfig) error                       // 初始化
    Run(errCh chan error)                                             // 运行
    Stop()                                                            // 停止
    Restart(option map[string]interface{}, api map[string]APIConfig,
            errCh chan error) error                                   // 重启
}

// 注册 API Server
var Slots = make(map[string]Apiserver)

func Register(name string, server Apiserver) error {
    if _, exist := Slots[name]; exist {
        return fmt.Errorf("apiserver name:%s exist", name)
    }
    Slots[name] = server
    return nil
}
```

### 6.2 Store 插件

```go
// apis/store/store.go
var StoreSlots = make(map[string]Store)

func RegisterStore(s Store) error {
    name := s.Name()
    if _, ok := StoreSlots[name]; ok {
        return errors.New("store name already existed")
    }
    StoreSlots[name] = s
    return nil
}

func GetStore() (Store, error) {
    name := config.Name
    store, ok := StoreSlots[name]
    if !ok {
        return nil, fmt.Errorf("store `%s` not found", name)
    }
    initialize(store)
    return store, nil
}
```

## 7. 插件链模式

某些插件支持链式调用，如 Crypto、History、DiscoverEvent：

```go
type PluginChanConfig struct {
    Name    string                 `yaml:"name"`     // 主插件名
    Option  map[string]interface{} `yaml:"option"`   // 主插件配置
    Entries []ConfigEntry          `yaml:"entries"`  // 链式插件列表
}
```

**应用场景：**
- Crypto 插件链：支持多种加密算法组合
- History 插件链：支持同时记录到日志和数据库
- DiscoverEvent 插件链：支持多种事件处理方式

## 8. 插件配置示例

```yaml
plugin:
  history:
    name: historyLogger
    entries:
      - name: logger
      - name: rds
        option:
          dbName: polaris
  statis:
    name: prometheus
    option:
      path: /metrics
      port: 8090
  ratelimit:
    name: token
    option:
      qps: 1000
  crypto:
    name: aes
    option:
      secret: "your-secret-key"
```

## 9. 插件生命周期

```
┌─────────────┐
│   import    │  通过 import 触发 init() 注册
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Register   │  调用 RegisterPlugin 注册到 pluginSet
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Initialize  │  系统启动时调用 Initialize() 初始化
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Run      │  插件正常运行
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Destroy   │  系统关闭时调用 Destroy() 清理资源
└─────────────┘
```

## 10. 扩展指南

### 10.1 开发新插件步骤

1. 定义插件结构体，实现 Plugin 接口
2. 在 init() 函数中调用 RegisterPlugin 注册
3. 在 main.go 或 plugin.go 中 import 插件包
4. 在配置文件中配置插件参数

### 10.2 插件开发规范

- 插件名称必须唯一
- 实现完整的生命周期管理
- 提供合理的默认配置
- 支持配置热更新（如适用）
- 记录必要的日志和指标