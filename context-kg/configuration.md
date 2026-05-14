---
title: 配置参考
tags: [config, yaml, deploy]
links: [overview, architecture]
updated: 2026-05-14
sources: 1
---

# 配置参考

本文覆盖 YAML 配置文件的完整结构。项目的启动流程见 [[overview]]，配置如何影响架构分层见 [[architecture]]。

## 配置文件位置

默认路径：`conf/pole-server.yaml`（可通过 `-c` 参数覆盖）

`utils.ConfDir` 被设置为配置文件所在的目录。

## 配置结构（`bootstrap/config/`）

顶层结构：

```yaml
bootstrap:
  logger:
    # 日志配置
  polarisService:
    # 自注册配置
    
apiServers:
  - name: httpserver
    option:
      listenPort: 8080
    api:
      admin:
        enable: true
        include: [...]
      client:
        enable: true
        include: [...]
        
store:
  name: defaultStore
  option:
    master:
      dbType: mysql
      dbName: pole_server
      dbUser: root
      dbPassword: ""
      dbAddr: 127.0.0.1:3306
      maxOpenConns: 100
      maxIdleConns: 10

cache:
  open: true
  resources:
    - name: service
    - name: instance
    - name: routingConfig
    # ... 其他资源缓存

namespace:
  autoCreate: true
  interceptors:
    - auth

service:
  interceptors:
    - paramcheck
    - auth
  batch:
    register:
      open: true
      queueSize: 10000
      waitTime: 32ms
      maxBatchCount: 128
      concurrency: 128
    deregister:
      open: true
      # 类似字段
    heartbeat:
      open: true
      # 类似字段

healthcheck:
  open: true
  service: polaris.checker
  slotNum: 30
  minCheckInterval: 10s
  maxCheckInterval: 60s
  checkers:
    - name: heartbeatChecker

config:
  interceptors:
    - paramcheck
    - auth

goverrule:
  interceptors:
    - paramcheck  
    - auth

auth:
  user:
    name: defaultUser
  strategy:
    name: defaultStrategy

plugin:
  statis:
    name: local
    option:
      interval: 60
  history:
    entries:
      - name: logger
  discoverEvent:
    entries:
      - name: local
  ratelimit:
    name: token-bucket
  whitelist:
    name: whitelist
  crypto:
    entries:
      - name: aes
```

## 插件配置

`apis.Config` 结构体映射到 YAML 中的 `plugin:` 节：

```go
type Config struct {
    CMDB                 ConfigEntry
    RateLimit            ConfigEntry
    History              PluginChanConfig   // 支持多个条目（链式）
    Statis               PluginChanConfig
    DiscoverStatis       ConfigEntry
    ParsePassword        ConfigEntry
    Whitelist            ConfigEntry
    MeshResourceValidate ConfigEntry
    DiscoverEvent        PluginChanConfig
    Crypto               PluginChanConfig
}
```

`PluginChanConfig` 支持 `entries` 数组，用于链式配置多个插件实现。

## API 服务端配置

`apiServers` 列表中每个 API 服务端的配置：
```yaml
- name: httpserver         # 必须与插件 Name() 返回值一致
  option:                  # 传递给 Initialize()
    listenPort: 8080
    enableSwagger: true
  api:
    admin:
      enable: true
      include:
        - defaultAccess
    client:
      enable: true
      include:
        - defaultAccess
```

`include` 列表中的条目是各服务端实现注册的 API 分组名称。

## 环境变量

无文档记录的环境变量——配置完全基于文件。配置文件路径是唯一的运行时参数。

## 部署目录

```
deploy/
├── conf/
│   ├── pole-server.yaml     # 示例配置
│   ├── i18n/                # 翻译文件
│   └── plugin/
│       └── ratelimit/       # 限流规则文件
└── tools/                   # 部署脚本
```

## 相关页面

- [[overview]]
- [[architecture]]
