---
title: 公共基础设施包
tags: [infra, logging, eventhub]
links: [architecture, domain-components]
updated: 2026-05-14
sources: 1
---

# 公共基础设施包

本文覆盖 `pkg/common/` 下的公共工具包。整体架构中公共包的角色见 [[architecture]]，各业务域如何使用这些工具见 [[domain-components]]。

## 日志（`pkg/common/log/`）

对 `go.uber.org/zap` 进行了封装，提供以下能力：
- 每个组件有独立命名的 Logger（例如 `[Skill]`、`[Service]`、`[Config]`）
- 通过 `lumberjack` 实现日志文件滚动
- 从 YAML 中读取配置：日志级别、输出路径、滚动设置
- `log.ConfigureFile(cfg)` 在 bootstrap 阶段早期调用

整个代码库的日志使用模式：
```go
log.Infof("[Skill] create skill, namespace: %s, name: %s", ns, name)
log.Errorf("[Service] failed to register instance: %v", err)
```

## EventHub（`pkg/common/eventhub/`）

进程内发布/订阅机制，用于组件间的异步解耦。

```go
// 发布
eventhub.Publish(eventhub.InstanceEventTopic, &InstanceEvent{...})

// 订阅
ctx := eventhub.Subscribe(eventhub.InstanceEventTopic, handler)
// ctx.Cancel() 取消订阅
```

主要主题（Topic）：
- `InstanceEventTopic` — 实例注册/注销/心跳
- `ConfigFilePublishTopic` — 配置文件发布
- `ServiceEventTopic` — 服务创建/删除
- 各类规则变更主题

使用方：健康检查（响应实例事件）、XDS（推送路由变更到 Envoy）、配置监听器（通知长轮询客户端）。

## Batch Controller（`pkg/common/batchctrl/`）

通用批处理框架，服务发现模块用它处理高吞吐量操作：

```go
type BatchController struct {
    queue     chan *Task
    batchSize int
    workers   int
    handler   BatchHandler
}
```

`pkg/service/batch/` 将其封装为：
- `RegisterController` — 批量实例注册
- `DeregisterController` — 批量实例注销
- `HeartbeatController` — 批量心跳更新

在每秒有数千个实例发送心跳的场景下，可有效减少数据库写入放大。

## OTel 指标（`pkg/common/otel/`）

OpenTelemetry 集成：
- OTLP gRPC 导出器，将数据发送到外部采集器
- 通过 `contrib/instrumentation/runtime` 采集运行时指标（GC、goroutine）
- 服务发现、配置中心和治理规则的自定义指标
- `pkg/common/otel/metrics/` — 指标定义与注册

## 同步原语（`pkg/common/syncs/`）

自定义同步原语：

```
syncs/
├── atomic/     # 原子类型封装
├── container/  # 线程安全容器类型（SyncSet、SyncMap 变体）
├── srand/      # 线程安全随机数
└── timewheel/  # 时间轮，用于高效的超时管理
```

`timewheel` 被健康检查模块使用，用于在数百万实例场景下高效检测心跳超时。

## 连接管理（`pkg/common/conn/`）

```
conn/
├── hook/       # 连接生命周期钩子
├── keepalive/  # TCP keepalive 配置
└── limit/      # 连接数限制（最大并发连接数）
    └── mock_net/  # 用于测试的 Mock net.Conn
```

## API 响应工具（`pkg/common/api/v1/`）

用于从 `pole-io/specification` protobuf 构建标准 `apimodel.Response` 和 `BatchWriteResponse` 对象的辅助函数：

```go
api.NewResponse(apimodel.Code_ExecuteSuccess)
api.NewResponseWithMsg(apimodel.Code_BadRequest, "invalid input")
api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
```

## 版本信息（`pkg/common/version/`）

编译时嵌入的版本信息。通过 `version` CLI 命令和自注册元数据对外暴露。

## 通用工具函数

```
pkg/common/utils/
├── hash/    # 一致性哈希
├── http/    # HTTP 客户端辅助函数
├── match/   # 模式匹配（正则、通配符）
├── time/    # 时间格式化辅助函数
└── valid/   # 输入校验辅助函数
```

`pkg/common/secure/` — 用于配置双向 TLS（mTLS）的 TLS 配置辅助函数。

## 国际化（`plugin/apiserver/httpserver/i18n/`）

使用 `go-i18n/v2` 对错误消息和 API 响应进行国际化处理。支持 zh-CN 和 en-US。消息文件位于 `deploy/conf/i18n/`。

## 相关页面

- [[architecture]]
- [[domain-components]]
