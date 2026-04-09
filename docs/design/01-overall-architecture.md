# Pole Control Plane 整体架构设计

## 1. 概述

Pole Control Plane 是一个云原生、AI Native 的服务治理平台，基于 Polaris 进行二开。提供服务注册发现、配置中心、服务治理规则管理等核心能力，并支持 MCP 协议以实现 AI Agent 认知。

## 2. 架构分层

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Layer                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ HTTP API │ │ gRPC API │ │ XDS v3   │ │ Nacos    │ Apollo   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
├─────────────────────────────────────────────────────────────────┤
│                       Service Layer                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ Service  │ │ Config   │ │GoverRule │ │Namespace │           │
│  │ Discover │ │ Center   │ │ Manager  │ │ Manager  │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
├─────────────────────────────────────────────────────────────────┤
│                       Cache Layer                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ Service  │ │ Config   │ │ Auth     │ │ Rules    │           │
│  │ Cache    │ │ Cache    │ │ Cache    │ │ Cache    │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
├─────────────────────────────────────────────────────────────────┤
│                       Store Layer                                │
│  ┌──────────────────────────────────────────────────┐           │
│  │              MySQL / Other Storage               │           │
│  └──────────────────────────────────────────────────┘           │
├─────────────────────────────────────────────────────────────────┤
│                      Plugin System                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ Auth     │ │ RateLimit │ │ Crypto  │ │ Observability│       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

## 3. 核心目录结构

```
pole-control-plane/
├── cmd/                    # 命令行入口
│   ├── root.go            # 根命令定义
│   ├── start.go           # 启动命令
│   └── version.go         # 版本命令
├── bootstrap/              # 启动引导
│   ├── run_*.go           # 平台特定启动逻辑
│   └── config/            # 启动配置
├── apis/                   # API 定义层
│   ├── apiserver/         # API Server 接口定义
│   ├── cache/             # Cache 接口定义
│   ├── store/             # Store 接口定义
│   ├── access_control/    # 访问控制接口
│   ├── observability/     # 可观测性接口
│   └── pkg/               # 公共类型定义
├── pkg/                    # 核心业务实现
│   ├── service/           # 服务注册发现
│   ├── config/            # 配置中心
│   ├── goverrule/         # 治理规则管理
│   ├── namespace/         # 命名空间管理
│   ├── cache/             # 缓存层实现
│   ├── admin/             # 管理功能
│   └── common/            # 公共组件
└── plugin/                 # 插件实现
    ├── apiserver/         # API Server 插件
    ├── access_control/    # 访问控制插件
    ├── crypto/            # 加密插件
    └── observability/     # 可观测性插件
```

## 4. 核心设计模式

### 4.1 插件化架构

系统采用插件化设计，所有可扩展组件通过插件机制实现：

```go
// 插件接口定义 (apis/plugin.go)
type Plugin interface {
    Name() string
    Initialize(c *ConfigEntry) error
    Destroy() error
    Type() PluginType
}
```

**支持的插件类型：**
- `PluginTypeStatis` - 统计插件
- `PluginTypeHistory` - 历史记录插件
- `PluginTypeDiscoverEvent` - 发现事件插件
- `PluginTypeRateLimit` - 限流插件
- `PluginTypeWhitelist` - 白名单插件
- `PluginTypeResourceAuth` - 资源鉴权插件
- `PluginTypeCMDB` - CMDB插件
- `PluginTypeApiServer` - API服务插件
- `PluginTypeCrypto` - 加密插件
- `PluginTypeStore` - 存储插件
- `PluginTypeHealthCheck` - 健康检查插件

### 4.2 拦截器链模式

Service、Config、GoverRule 等模块通过拦截器链实现横切关注点：

```go
// 拦截器注册顺序
func GetChainOrder() []string {
    return []string{
        "auth",       // 鉴权拦截器
        "paramcheck", // 参数校验拦截器
    }
}
```

### 4.3 事件驱动架构

通过 EventHub 实现组件间的松耦合通信：

```go
// 事件订阅
subCtx, err := eventhub.Subscribe(eventhub.InstanceEventTopic, eventHandler)

// 事件发布
eventhub.Publish(eventhub.InstanceEventTopic, event)
```

## 5. 启动流程

```
main.go
  └── cmd.Execute()
        └── bootstrap.Start(configFilePath)
              ├── config.Load(filePath)           // 加载配置
              ├── store.GetStore()                // 初始化存储
              ├── cache.Initialize()              // 初始化缓存
              ├── service.Initialize()            // 初始化服务模块
              ├── config.Initialize()             // 初始化配置模块
              ├── goverrule.Initialize()          // 初始化治理规则
              ├── namespace.Initialize()          // 初始化命名空间
              ├── admin.Initialize()              // 初始化管理模块
              └── StartServers(servers)           // 启动API服务器
                    └── WaitSignal()              // 等待信号
```

## 6. 核心模块依赖关系

```
                    ┌─────────────┐
                    │   APIServer │
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
                    │    Cache    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │    Store    │
                    └─────────────┘
```

## 7. 多协议支持

系统支持多种服务发现协议：

| 协议 | 插件路径 | 说明 |
|------|----------|------|
| gRPC | `plugin/apiserver/grpcserver` | Polaris 原生 gRPC 协议 |
| HTTP | `plugin/apiserver/httpserver` | RESTful API |
| XDS v3 | `plugin/apiserver/xdsserverv3` | Envoy XDS 协议 |
| Nacos | `plugin/apiserver/nacosserver` | Nacos 兼容协议 |
| Apollo | `plugin/apiserver/apolloserver` | Apollo 配置协议 |
| Eureka | `plugin/apiserver/eurekaserver` | Eureka 兼容协议 |
| Consul | `plugin/apiserver/consulserver` | Consul 兼容协议 |

## 8. AI Native 特性

### 8.1 MCP 协议支持

系统支持 MCP (Model Context Protocol) 协议，使 AI Agent 能够：
- 发现和调用服务
- 获取服务元数据
- 管理配置

**MCP Server 管理：**
- 支持注册和管理 MCP Server
- 支持声明 MCP Server 提供的工具
- 支持多种传输协议（stdio、sse、websocket）

### 8.2 Skill Hub

基于配置中心实现 Skill Hub 功能：
- 支持多个 Skill 组成一个 Group
- 底层能力复用配置中心
- 支持 AI Agent 动态发现和调用

**Skill 版本管理：**
- 每个 Skill 支持多版本并存
- 可指定活跃版本供默认调用
- 客户端可订阅特定版本

**Skill 订阅机制：**
- 客户端订阅 Skill 变更通知
- 版本更新时主动推送
- 支持订阅状态管理

详细设计参见 [10-ai-native-features.md](./10-ai-native-features.md)

## 9. 高可用设计

### 9.1 推空保护

支持自适应或用户定义的实例推空保护，避免网络故障恢复期间对主调方的影响。

### 9.2 灰度发布

所有治理规则均支持：
- 版本控制
- 灰度下发
- 标签灰度

### 9.3 细粒度鉴权

所有资源支持细粒度鉴权，对标云厂商的 CAM/RAM 能力。