---
title: ADR：pole-rust-client 观测职责与上报边界
tags: [adr, observability, otel, rust-sdk]
links: [adr-otel-observability-platform, service-discovery, config-center, governance-rules]
updated: 2026-07-18
sources: 6
---

# ADR：pole-rust-client 观测职责与上报边界

## 状态

Proposed。

## 背景

Pole 的业务侧观测由 sidecar 与 SDK 共同完成。`pole-control-plane` 只负责展示、查询分析和自身内部观测上报，不采集业务流量。`pole-rust-client` 作为 Rust 业务服务接入 Pole 的 SDK，需要把服务发现、配置订阅、治理执行和业务调用上下文转换为 OTel 语义。

本文定义 `pole-rust-client` 在可观测性体系中的职责边界。总体平台方案见 [[adr-otel-observability-platform]]。

## 职责边界

`pole-rust-client` 负责：

- 注入 Pole 服务绑定属性，保证业务请求 metrics、runtime metrics、Kubernetes CPU/Mem、trace 和 event 可以归属到同一服务。
- 上报 Rust SDK 自身的服务发现、配置订阅、治理决策、缓存刷新、错误和重试指标。
- 产生结构化业务 event，例如服务发现更新、配置变更感知、治理规则命中、治理拒绝、SDK 内部异常。
- 在 SDK 参与路由、负载均衡、限流、熔断、故障探测时，把治理上下文写入当前 span attributes 或 span events。
- 提供 HTTP/gRPC 框架集成 hooks/layers，使业务框架的请求 span 能关联 Pole 选址和治理结果。
- 遵循 OTel 环境变量和 Resource 规范，允许数据发送到本地 sidecar、OpenTelemetry Collector 或外部 OTLP endpoint。
- 支持通过 Pole 服务发现动态获取 OTLP 上报地址，减少用户显式配置 Collector endpoint。
- 支持通过 Pole 配置中心 remote 下发 SDK 观测配置，例如采样、开关、batch/export 参数和 event 策略。

`pole-rust-client` 不负责：

- 不采集业务普通日志。
- 不做日志全文检索、日志解析、日志脱敏。
- 不直接采集 Kubernetes CPU/Mem；这部分由 Collector 的 Kubernetes/kubelet receiver 负责。
- 不替代业务框架的标准 HTTP/gRPC instrumentation。
- 不承担 GreptimeDB/OpenObserve 查询、Console 展示或控制面审计职责。

## Resource Attributes

SDK 初始化时必须提供稳定 Resource attributes。优先来源为用户配置、环境变量、Kubernetes Downward API、Pole 服务实例元数据。

| 属性 | 来源 | 说明 |
|---|---|---|
| `service.name` | 业务应用配置或 `pole.io/service` | OTel 标准服务名 |
| `service.version` | 业务应用配置 | 应用版本 |
| `service.instance.id` | `pole.io/instance-id` / Pod UID / SDK 生成实例 ID | 服务实例 ID |
| `telemetry.sdk.language` | SDK 固定值 | 固定为 `rust` |
| `pole.runtime.language` | `pole.io/runtime-language` 或 SDK 默认 | Rust SDK 默认 `rust` |
| `pole.namespace` | `pole.io/namespace` / SDK namespace | Pole 命名空间 |
| `pole.service.name` | `pole.io/service` / SDK service name | Pole 服务名 |
| `pole.service.instance.id` | `pole.io/instance-id` / 注册实例 ID | Pole 实例 ID |
| `k8s.namespace.name` | Downward API / `pole.io/k8s-namespace` | Kubernetes namespace |
| `k8s.pod.uid` | Downward API / `pole.io/pod-uid` | Pod UID |
| `k8s.pod.name` | Downward API / `pole.io/pod-name` | Pod 名称 |
| `container.name` | `pole.io/container-name` | 容器名 |
| `k8s.workload.name` | `pole.io/workload-name` | 工作负载名 |
| `k8s.workload.kind` | `pole.io/workload-kind` | 工作负载类型 |
| `k8s.cluster.name` | `pole.io/cluster` | 集群名 |

绑定优先级：

1. Pod/实例级：`k8s.pod.uid` + `container.name`。
2. Pole 实例级：`pole.service.instance.id` / `service.instance.id`。
3. Workload 级：`pole.namespace` + `pole.service.name` + `k8s.workload.name`。
4. 服务级：`pole.namespace` + `pole.service.name`。

## Metrics

Rust SDK 不重新发明 runtime metrics 名称。Rust 当前没有像 JVM/Go 那样完整统一的 OTel runtime 指标族，首期展示重点为业务请求、SDK 内部状态、治理执行和 Kubernetes 资源指标。

### SDK 内部 metrics

| 指标名 | 类型 | 单位 | 属性 | 说明 |
|---|---|---|---|---|
| `pole.sdk.discovery.refresh.count` | Counter | `{refresh}` | `pole.namespace`, `pole.service.name`, `pole.result`, `pole.error.code` | 服务发现刷新次数 |
| `pole.sdk.discovery.refresh.duration` | Histogram | `s` | `pole.namespace`, `pole.service.name`, `pole.result` | 服务发现刷新耗时 |
| `pole.sdk.discovery.cache.entry.count` | Gauge | `{entry}` | `pole.namespace`, `pole.service.name`, `pole.resource.type` | SDK 本地缓存条目数 |
| `pole.sdk.config.watch.count` | Counter | `{event}` | `pole.namespace`, `pole.service.name`, `pole.config.group`, `pole.result` | 配置订阅事件数 |
| `pole.sdk.config.watch.delay` | Histogram | `s` | `pole.namespace`, `pole.service.name`, `pole.config.group` | 配置变更从接收到应用生效的延迟 |
| `pole.sdk.governance.decision.count` | Counter | `{decision}` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.rule.type`, `pole.result` | SDK 治理决策次数 |
| `pole.sdk.governance.decision.duration` | Histogram | `s` | `pole.namespace`, `pole.service.name`, `pole.rule.type`, `pole.result` | 治理决策耗时 |
| `pole.sdk.request.retry.count` | Counter | `{retry}` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.retry.reason` | SDK 重试次数 |
| `pole.sdk.error.count` | Counter | `{error}` | `pole.namespace`, `pole.service.name`, `pole.component`, `pole.error.code` | SDK 内部错误次数 |
| `pole.sdk.telemetry.endpoint.refresh.count` | Counter | `{refresh}` | `pole.namespace`, `pole.service.name`, `pole.result`, `pole.error.code` | OTLP 上报端点刷新次数 |
| `pole.sdk.telemetry.endpoint.active.count` | Gauge | `{endpoint}` | `pole.namespace`, `pole.service.name`, `pole.endpoint.source` | 当前可用 OTLP endpoint 数 |
| `pole.sdk.telemetry.config.reload.count` | Counter | `{reload}` | `pole.namespace`, `pole.service.name`, `pole.result`, `pole.error.code` | SDK 观测 remote config 重载次数 |

### 业务调用 metrics

如果业务通过 SDK 提供的 client/hook 发起调用，SDK 可以补充或协助业务框架上报：

| 指标名 | 类型 | 单位 | 属性 | 说明 |
|---|---|---|---|---|
| `pole.service.request.count` | Counter | `{request}` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.protocol`, `pole.result`, `pole.error.code` | 服务间调用请求数 |
| `pole.service.request.duration` | Histogram | `s` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.protocol`, `pole.result` | 服务间调用耗时 |
| `pole.service.governance.route.match.count` | Counter | `{match}` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.rule.type`, `pole.result` | 治理命中聚合，不带 `rule.id` |
| `pole.service.governance.rate_limit.count` | Counter | `{request}` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.result` | 限流结果聚合 |
| `pole.service.governance.circuit_breaker.count` | Counter | `{request}` | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.result` | 熔断结果聚合 |

metrics label 禁止包含 `pole.rule.id`、`trace_id`、`request_id`、`pole.service.instance.id`、原始 URL path、客户端 IP 等高基数字段。

## Events

Rust SDK 只上报结构化 event，不上报业务普通日志。

| EventName | 触发点 | 关键属性 |
|---|---|---|
| `pole.service.discovery.updated` | 服务发现结果更新 | `pole.namespace`, `pole.service.name`, `pole.callee.service.name`, `pole.revision` |
| `pole.service.config.changed` | 配置变更被 SDK 感知 | `pole.namespace`, `pole.service.name`, `pole.config.group`, `pole.config.file`, `pole.revision` |
| `pole.service.governance.rule.matched` | 路由、泳道、镜像、Mock 等规则命中 | `pole.governance.decision.id`, `pole.rule.type`, `pole.rule.id`, `pole.rule.name`, `trace_id`, `span_id` |
| `pole.service.governance.request.rejected` | 限流、熔断、鉴权等导致拒绝 | `pole.governance.decision.id`, `pole.rule.type`, `pole.rule.id`, `pole.reject.reason`, `trace_id`, `span_id` |
| `pole.sdk.cache.refresh_failed` | SDK 缓存刷新失败 | `pole.component`, `pole.error.code`, `error.message` |
| `pole.sdk.connection.state_changed` | SDK 与 control-plane / sidecar 连接状态变化 | `pole.connection.state`, `pole.endpoint`, `pole.reason` |
| `pole.sdk.telemetry.endpoint.changed` | OTLP 上报端点变化 | `pole.endpoint.source`, `pole.endpoint.old`, `pole.endpoint.new`, `pole.reason` |
| `pole.sdk.telemetry.config.reloaded` | SDK 观测 remote config 生效 | `pole.config.group`, `pole.config.file`, `pole.revision`, `pole.result` |

治理 event 必须和治理 metrics 成对出现：

- metrics 用于低基数聚合，不携带 `rule.id`。
- event 携带 `pole.rule.id`、`pole.governance.decision.id`、`trace_id/span_id`，用于定位具体规则和链路。
- span attributes 也写入同一个 `pole.governance.decision.id`，Console 才能从指标异常跳到事件，再跳到 trace。

## Trace 集成

Rust SDK 不强制接管业务框架 tracing，但必须提供以下能力：

- 读取当前上下文 span，将服务发现、选址、负载均衡和治理结果写入 span attributes。
- 在生成治理 event 时关联当前 trace/span。
- 对外提供 hooks/layers，供 `tonic`、`hyper`、`reqwest` 或业务自定义 RPC 框架集成。
- 遵循 W3C TraceContext 传播，不能发明私有 trace header。

建议 span attributes：

| 属性 | 说明 |
|---|---|
| `pole.namespace` | 调用方命名空间 |
| `pole.service.name` | 调用方服务 |
| `pole.callee.service.name` | 被调服务 |
| `pole.callee.instance.id` | 被选中的实例 ID，仅 span/event 使用 |
| `pole.governance.decision.id` | 单次治理决策 ID |
| `pole.rule.type` | 规则类型 |
| `pole.rule.id` | 具体规则 ID，仅 span/event 使用 |
| `pole.load_balance.policy` | 负载均衡策略 |
| `pole.retry.count` | 重试次数 |

## 配置要求

Rust SDK 观测配置必须允许以下部署形态：

| 形态 | 配置 |
|---|---|
| 通过 sidecar 上报 | OTLP endpoint 指向 localhost sidecar 或本地 Collector |
| 直接上报 Collector | 使用 `OTEL_EXPORTER_OTLP_ENDPOINT` |
| 通过 Pole 服务发现上报 | 发现 `pole-otel-collector` 服务实例并生成 OTLP endpoint |
| 通过配置中心下发 | 配置中心 remote config 指定 endpoint 发现策略、采样、batch 和 event 开关 |
| 关闭 SDK 观测 | 显式 disable，但仍允许业务框架自身 instrumentation 工作 |
| 只注入 Resource attributes | 不导出 SDK metrics/events，只给业务 instrumentation 补 Pole 绑定属性 |

SDK 应优先兼容 OTel 标准环境变量：

- `OTEL_SERVICE_NAME`
- `OTEL_RESOURCE_ATTRIBUTES`
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `OTEL_TRACES_EXPORTER`
- `OTEL_METRICS_EXPORTER`
- `OTEL_LOGS_EXPORTER`

Pole 自定义配置只补充无法由标准 OTel 环境变量表达的内容，例如命名空间、治理决策采样、SDK event 开关。

### 上报地址发现

SDK 必须支持三种 endpoint 来源，按优先级合并：

1. 本地显式配置：代码配置或 `OTEL_EXPORTER_OTLP_*` 环境变量，适合调试、灰度和强制覆盖。
2. 配置中心 remote config：下发 endpoint 发现策略、固定 endpoint 或 endpoint 服务名。
3. Pole 服务发现：按服务名发现 Collector/sidecar 的实例列表。

默认推荐路径：

```text
业务进程启动
  -> 使用最小 Pole bootstrap 地址连接 control-plane
  -> 注册业务服务实例并订阅自身 SDK 配置
  -> 发现 pole-observability / pole-otel-collector
  -> 生成 OTLP endpoint
  -> 初始化或热更新 OTel exporter
```

Collector 服务注册约定：

| 字段 | 推荐值 | 说明 |
|---|---|---|
| namespace | `pole-observability` | 独立于业务 namespace，便于统一授权和运维 |
| service | `pole-otel-collector` | 默认 Collector 服务名 |
| protocol | `grpc` / `http` | 对应 OTLP/gRPC 4317 或 OTLP/HTTP 4318 |
| metadata `pole.io/otel-endpoint-kind` | `otlp-grpc` / `otlp-http` | 标识实例可作为 OTLP exporter 目标 |
| metadata `pole.io/otel-signal` | `traces,metrics,logs` | 标识支持的信号 |
| metadata `pole.io/otel-path` | `/` / `/v1` | HTTP endpoint path 前缀，避免 SDK 猜测 |
| metadata `pole.io/secure` | `true` / `false` | 是否启用 TLS/mTLS |

服务发现生成 endpoint 的规则：

- `otlp-grpc` 默认端口为实例端口或 metadata 指定端口，生成 `http://host:port` 或 `https://host:port`。
- `otlp-http` 生成 base endpoint，例如 `http://host:4318`，具体 `/v1/traces`、`/v1/metrics`、`/v1/logs` 路径遵循 OTel exporter 规范。
- 多个 Collector 实例按 SDK 负载均衡策略轮询或随机选择；失败时切换实例并记录 `pole.sdk.telemetry.endpoint.refresh.count`。
- endpoint 刷新必须跟随服务发现 revision，不能每次 export 都查服务发现。
- 如果服务发现为空，SDK 继续使用上一次可用 endpoint；如果没有缓存，则回退到环境变量或本地配置。
- endpoint 变化时允许热更新 exporter，但必须 drain 当前 batch 或等待短超时，避免丢失大量遥测数据。

### 配置中心 Remote Config

SDK 观测配置通过配置中心下发，推荐固定配置文件：

| 字段 | 推荐值 |
|---|---|
| namespace | 业务服务所在 namespace |
| group | `pole-sdk` |
| file | `observability.yaml` |
| 订阅方式 | SDK 启动后 watch；断线时使用本地快照 |

推荐配置结构：

```yaml
enabled: true
resource:
  runtime_language: rust
exporter:
  mode: discovery
  discovery:
    namespace: pole-observability
    service: pole-otel-collector
    protocol: otlp-grpc
  endpoint: ""
  timeout: 10s
  compression: gzip
signals:
  traces:
    enabled: true
    sampler: parentbased_traceidratio
    sample_ratio: 0.1
  metrics:
    enabled: true
    export_interval: 60s
  events:
    enabled: true
    governance_decision: true
    config_changed: true
    discovery_updated: true
limits:
  max_event_attributes: 64
  max_queue_size: 2048
```

配置优先级：

1. 代码显式配置：最高优先级，用于测试或应用强约束。
2. 环境变量：遵循 OTel 标准，例如 `OTEL_SERVICE_NAME`、`OTEL_RESOURCE_ATTRIBUTES`、`OTEL_EXPORTER_OTLP_ENDPOINT`、`OTEL_TRACES_EXPORTER`、`OTEL_METRICS_EXPORTER`、`OTEL_LOGS_EXPORTER`。
3. 配置中心 remote config：用于平台统一下发默认值和动态调整。
4. SDK 默认值：只用于兜底，必须可预测。

动态生效规则：

- `enabled`、采样率、event 开关、metrics export interval、batch queue size 可以热更新。
- Resource identity 相关字段如 `service.name`、`pole.namespace`、`service.instance.id` 默认不热更新；如果确需变更，应要求重启业务进程。
- endpoint mode 从 `discovery` 切到固定 `endpoint` 时可以热更新，但必须发出 `pole.sdk.telemetry.endpoint.changed` event。
- 配置解析失败时保留上一份有效配置，并上报 `pole.sdk.telemetry.config.reload.count{pole.result="failure"}`。
- 配置中心不可用时使用本地快照；没有快照时使用环境变量和 SDK 默认值。

最小用户配置：

```rust
PoleClient::builder()
    .server("http://pole-control-plane:8090")
    .namespace("default")
    .service("order-service")
    .enable_observability()
    .build()
```

在上述模式下，用户无需手写 OTLP endpoint。SDK 通过服务发现找 `pole-otel-collector`，通过配置中心拿 observability 策略；只有 Pole bootstrap 地址、namespace 和 service name 仍是必须信息。

### 标准 OTel 兼容

SDK 不把服务发现 endpoint 写成新的私有协议。内部最终仍映射为 OTel exporter 配置：

| Pole remote 字段 | OTel exporter 配置 |
|---|---|
| `exporter.endpoint` 或服务发现结果 | `OTEL_EXPORTER_OTLP_ENDPOINT` 等价代码配置 |
| `exporter.protocol` | `OTEL_EXPORTER_OTLP_PROTOCOL` 等价代码配置 |
| `exporter.timeout` | `OTEL_EXPORTER_OTLP_TIMEOUT` 等价代码配置 |
| `signals.traces.enabled=false` | `OTEL_TRACES_EXPORTER=none` 等价行为 |
| `signals.metrics.enabled=false` | `OTEL_METRICS_EXPORTER=none` 等价行为 |
| `signals.events.enabled=false` | `OTEL_LOGS_EXPORTER=none` 等价行为 |

因此，remote config 只是配置来源之一，不改变 OTel/OTLP 数据协议，也不绕过 Collector。

## 首轮对接计划

当前 `pole-client-rust` 可以先按“语义模型已具备、真实导出待接入”的方式推进。首轮不要把所有 HTTP/gRPC 框架适配一次做完，先把 SDK 自身和治理链路上报打通。

首轮模块拆分：

| 模块 | 职责 | 首轮交付 |
|---|---|---|
| `observability::req` | 稳定数据契约 | 继续作为 Resource attributes、MetricRecord、ObservabilityEvent、治理决策上下文和低基数过滤的公共模型 |
| `observability::default` | 配置归一 | 合并代码配置、OTel 环境变量、本地 `global.observability`、remote config 和服务发现 endpoint |
| `observability::api` | recorder 接口 | 保持 no-op 默认实现，新增真实 OTel recorder adapter |
| `observability::exporter` | OTel SDK 适配 | 初始化 meter/logger/tracer provider，按 signal 开关导出 metrics、event logs 和 traces |
| `observability::discovery` | endpoint 发现 | 订阅 `pole-observability/pole-otel-collector`，按 revision 更新 exporter endpoint |
| `observability::remote_config` | 配置中心下发 | watch `pole-sdk/observability.yaml`，热更新采样、event 开关、export interval、batch 参数和 endpoint mode |

首轮接入点：

- 服务发现刷新成功/失败：记录 `pole.sdk.discovery.refresh.count`、`pole.sdk.discovery.refresh.duration`，失败时发 `pole.sdk.cache.refresh_failed` 或更具体的 discovery event。
- 配置 watch：记录 `pole.sdk.config.watch.count`、`pole.sdk.config.watch.delay`，感知业务配置变更时发 `pole.service.config.changed`。
- 治理执行：在 route、rate limit、circuit breaker、fault detect、security、mirror、mock 产生决策时，记录低基数 `pole.sdk.governance.decision.count`，并发带 `pole.rule.id`、`pole.governance.decision.id`、`trace_id/span_id` 的治理 event。
- 请求调用：如果业务通过 SDK 提供的 client/hook 发送请求，记录 `pole.service.request.count` 和 `pole.service.request.duration`；如果流量由 sidecar 代理，则 SDK 默认只注入 Resource attributes 和 trace context，避免重复计数。
- endpoint/config 状态：endpoint 变化发 `pole.sdk.telemetry.endpoint.changed`，remote config 生效发 `pole.sdk.telemetry.config.reloaded`，失败只保留上一份有效配置。

与 sidecar 共存规则：

- SDK 与 sidecar 都必须设置 `pole.node.role`，取值分别为 `sdk` 和 `sidecar`。
- 同一次请求只能有一个“实际治理执行方”上报治理 metrics；另一方只补 trace attributes 或转发上下文。
- SDK 与 sidecar 若共同参与同一次治理决策，必须复用同一个 `pole.governance.decision.id`，保证 Console 能从 metrics 时间窗跳到 event，再跳到 trace。
- sidecar 上报代理视角请求指标时，SDK 不应再对同一层调用重复上报 `pole.service.request.count`；SDK 可以只上报内部发现、配置、治理缓存和业务 hook 明确发起的请求。

当前实现差距：

- `observability::api::ObservabilityRecorder` 仍是接口和 no-op 默认实现，需要新增真实 OTel adapter。
- `DefaultObservability` 当前可以从环境变量和 `serverConnectors.observability` 生成 endpoint，但服务发现 endpoint 刷新和配置中心 remote config 启动链路还需要接入到 `SDKContext` / `Engine` 生命周期。
- `TelemetryEndpointSource` 当前只有 `LocalConfig` 和 `Env`，首轮需要补 `RemoteConfig` 与 `ServiceDiscovery`，并记录 endpoint 切换事件。
- 治理 metrics/event 语义模型已能区分低基数 metrics 和高基数 event/span 字段，但各治理执行点还需要实际调用 recorder。

## 验收要求

- Rust SDK 上报的 metrics 不包含 `rule.id`、`trace_id` 等高基数字段。
- 治理命中时，同一次决策能在 metrics、event、trace 中通过 `pole.governance.decision.id` 关联。
- SDK Resource attributes 能和 Kubernetes CPU/Mem 指标通过 `k8s.pod.uid` 或 `pole.service.instance.id` 关联。
- 关闭 SDK 观测后，不影响业务框架原有 OTel instrumentation。
- Console 可以根据 `pole.runtime.language=rust` 展示通用 Rust 服务观测面板：请求、资源、治理、trace、event，而不是 JVM/Go 专属面板。
- 未配置 `OTEL_EXPORTER_OTLP_ENDPOINT` 时，SDK 能通过 Pole 服务发现获取 `pole-otel-collector` 并上报 metrics/events/traces。
- 配置中心下发采样率、event 开关或 exporter 参数后，SDK 能热更新并在失败时保留上一份有效配置。
- 服务发现 Collector 实例变化时，SDK 能切换 endpoint，并发出 endpoint changed event。

## 参考

- [OpenTelemetry SDK environment variables](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/)
- [OpenTelemetry Protocol Exporter](https://opentelemetry.io/docs/specs/otel/protocol/exporter/)
- [OpenTelemetry Logs Data Model](https://opentelemetry.io/docs/specs/otel/logs/data-model/)

## 相关页面

- [[adr-otel-observability-platform]]
- [[service-discovery]]
- [[config-center]]
- [[governance-rules]]
