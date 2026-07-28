---
title: 统一进程模式与 Pole Limiter 集成
tags: [adr, runtime, limiter, deploy, process]
links: [architecture, configuration, adr-console-limiter-source-layout-and-embedded-web]
updated: 2026-07-29
sources: 14
---

# 统一进程模式与 Pole Limiter 集成

## 状态

已接受并完成兼容期实现。当前版本保留 `all=control-plane+console`，以 `full` 显式启动三个模块；下一次 breaking release 再评估收敛 `all` 语义。

## 背景

`pole-control-plane` 当前由 `bootstrap.Start` 直接装配进程，支持：

- `server`：只启动 Control Plane；
- `console`：只启动 Console Gateway；
- `all`：启动 Control Plane 与 Console，且是空配置时的默认值。

`pole-limiter-server` 是 Pole SDK 直连的分布式限流数据面。它默认监听 gRPC `8101` 和 HTTP 运维端口 `8100`，并通过 Control Plane Discover gRPC `8091` 注册、心跳和反注册。

Limiter 当前不能直接嵌入 Control Plane 进程：

- bootstrap 在失败时调用 `os.Exit`，并自行监听 OS signal；
- bootstrap 创建脱离宿主的根 Context；
- listener 在 goroutine 内绑定，启动返回与 ready 状态之间没有可靠契约；
- API Server、插件、Limiter Core 和注册状态大量使用包级 map、单例及 `sync.Once`；
- root `package main` 通过 blank import 完成插件注册，单独 import bootstrap 不会获得完整装配；
- Limiter 仍使用较旧的 Go、Specification、gRPC、Cobra 和 Zap 依赖基线。

因此，“加入一个 mode 分支并调用 Limiter bootstrap”不是有效集成。

## 决策

### 统一源码与制品，不强制统一部署

目标是把 Limiter 源码和运行模块纳入 `pole-control-plane` 的版本与构建边界，产出同一二进制和镜像，由 mode 选择当前进程装配。

生产环境仍推荐把 `control-plane`、`console`、`limiter-server` 作为独立 workload 部署。统一制品解决版本漂移和交付复杂度，独立 workload 保留正确的故障域、扩缩容和发布节奏。

不采用 Control Plane 拉起外部 Limiter 子进程作为终态。子进程方案仍需交付两个二进制，并增加 signal、日志、健康检查、配置和版本协调复杂度。

### Mode 是静态 Profile

目标态使用三个规范 Module ID：

| mode | 进程内 Module | 主要用途 |
|---|---|---|
| `console` | Console | Console 独立部署，连接远端 Control Plane |
| `control-plane` | Control Plane | 控制面独立部署 |
| `limiter-server` | Limiter | 限流数据面独立部署，注册到远端 Control Plane |
| `all` | Control Plane + Console | 当前兼容默认入口 |
| `full` | Control Plane + Limiter + Console | 本地开发、演示、单机或轻量部署 |

保留 `server` 作为 `control-plane` 的废弃别名，并输出迁移告警。

Mode 只负责选择 Profile；不要再增加 `limiter.enabled` 形成第二套装配真相。Limiter 的 listener、registry、runtime 和 statistics 配置放在顶层 `limiter:` 下。监听地址与注册给 SDK 的 advertised endpoint 必须分开配置。

### 兼容发布

当前 `all` 已经是默认值，且语义是 Control Plane + Console。静默把它扩展为三模块会让旧部署突然绑定 `8100/8101`、增加内存和 goroutine、要求唯一 `node-id` 并发起 Limiter 自注册，属于破坏性变更。

兼容期采用以下顺序：

1. 新增 `control-plane`、`limiter-server`，保留 `server` 别名；
2. 当前大版本继续保持旧 `all` 语义，并用临时 `full` 显式启用三模块；
3. 下一次 breaking release 将 `all` 收敛为三模块并移除 `full`；
4. 生产 manifests 必须始终显式指定 mode，不依赖默认值。

如果选择立即只保留四个目标 mode，则必须按 breaking change 发布；旧配置缺少完整 `limiter:` 段时 fail closed，禁止隐式使用 `node-id: 1` 启动。

## 进程编排 Seam

对 CLI 暴露一个深 Module：

```go
type Options struct {
    ConfigPath   string
    ModeOverride string
}

func Run(ctx context.Context, opts Options) error
```

内部生命周期 Interface：

```go
type Module interface {
    Start(ctx context.Context) (Running, error)
}

type Running interface {
    Endpoints() []Endpoint
    Wait() error
    Stop(ctx context.Context) error
}
```

Interface 不变量：

- `Start` 只在必需 listener 已同步绑定、内部循环已启动且必要注册已完成后返回；
- `Start` 失败必须回滚本次已获得的全部资源；
- Module 不监听 OS signal、不调用 `os.Exit`、不创建脱离父 Context 的根 Context；
- `Wait` 只报告运行期意外退出；
- `Stop` 幂等且接受 deadline；
- 任一必选 Module 意外退出时，Supervisor fail-fast 并停止整个 Profile；
- 停止按依赖逆序执行，清理错误不能覆盖主错误；
- 进程级日志、metrics 和 event hub 只初始化一次。

Catalog 应使用显式构造，不依赖 `init()` 注册和包级可变 map。Profile resolver 负责 CLI 覆盖、别名和默认值，Supervisor 负责配置预校验、依赖排序、readiness、失败回滚和优雅停机。

### `full` 启停顺序

启动顺序：

1. 预校验全部配置、端口冲突、Limiter `node-id` 和 advertised endpoint；
2. 启动 Control Plane，等待 `8091` 等必需端口 ready；
3. 启动 Limiter，等待 `8101` ready，并通过 loopback `8091` 完成注册；
4. 启动 Console；
5. 进入统一 Wait/signal 循环。

停止按 Console → Limiter → Control Plane 的逆序执行。Limiter 停止时先关闭 listener 并排空长流，再反注册和销毁统计插件。

`full` 中 Limiter 仍经 gRPC Adapter 注册到 Control Plane，不增加只在同进程生效的注册捷径，以保持合并和分离部署语义一致。

## Limiter Module 边界

Limiter Module 内部隐藏：

- counter、client、sliding-window 和 push worker；
- gRPC `8101` 数据面 Adapter；
- 可选 HTTP `8100` 运维 Adapter；
- Pole Discover 注册、心跳和反注册 Adapter；
- statistics Adapter；
- listener、goroutine、readiness 和停机细节。

Limiter 到 Control Plane 的注册是“远程但自有”依赖。定义窄 `RegistryPort`，生产使用 gRPC Adapter，测试使用内存 Adapter。

当前 counter state 是进程内实现。配置中的 remote storage 相关字段尚无完整消费链，不提前制造假想 Storage seam；出现第二个真实 Adapter 后再抽象。

## 部署与安全约束

Limiter 是维护 gRPC 双向长流和内存计数器的有状态数据面，不能按普通控制面 Module 对待：

- Control Plane 按管理请求和数据库压力扩容，Limiter 按长连接、counter 数和限流 QPS 扩容；
- 合并进程会耦合 CPU、GC、内存、OOM、滚动发布和故障恢复；
- Control Plane 发布会使同进程 Limiter 重启，导致所有配额流无谓重连；
- counter key 高 8 位编码 `node-id`，Limiter 实例间必须使用唯一 ID；
- 当前多副本若共享同一 advertised `host:port` 会产生服务实例身份冲突；
- 当前状态以内存为主，重启会丢失计数状态。

因此三模块 `full` 只作为 quickstart 和轻量部署 Profile，不作为默认生产拓扑。

## 实现落点

- `bootstrap/config/start_mode.go` 负责规范 mode、`server` 别名和 Profile 解析；
- `bootstrap/process.go` 负责进程配置加载与 Control Plane、Limiter、Console Module 装配；
- `internal/runtime/supervisor/` 负责顺序启动、运行期 fail-fast、失败回滚和逆序停止；
- `limiter/` 提供可重复构造的 `Start → Running.Wait/Stop` 模块，不拥有 OS signal；
- 选中 Profile 的 listener 在获取运行资源前完成端口冲突检查；
- Control Plane gRPC/HTTP 使用显式 ready 信号，旧协议 Server 使用有界 TCP readiness 探测；
- Limiter gRPC/HTTP listener 同步绑定后才报告启动成功；默认示例只启用内部 gRPC `8101`；
- Limiter statistics、core、registry 和 API Server 状态均绑定到实例生命周期，不再依赖迁移前的生命周期全局单例。

以上为当前实现落点。后续源码归属将把 Console Go Module 与 Limiter 分别迁入 `pkg/console`、`pkg/limiter`，并将 Console 前端迁入 `web/console` 后通过 `go:embed` 纳入同一制品；详见 [[adr-console-limiter-source-layout-and-embedded-web]]。

安全边界：

- `8100` 当前包含无鉴权的 pprof/maintain 能力，目标配置应默认关闭，或只监听 loopback/管理网络；
- `8101` 必须使用独立内部 Service 与 NetworkPolicy，不能随 Console 公网入口一起暴露；
- TLS 与工作负载认证在生产开放前应形成明确 Adapter 和验收门槛。

## 实施阶段

1. 固化现有 Limiter gRPC 双向流、配额、注册、重连和错误码 contract tests；
2. 修复既有 vet 问题，并将 Specification、gRPC 和公共依赖对齐到 Control Plane 基线；
3. 从 Limiter bootstrap 中移除 `os.Exit`、signal ownership、根 Context 和包级生命周期单例；
4. 建立构造器装配的 Limiter Module，旧 `pole-limiter` main 暂时作为薄 Adapter；
5. 将 Control Plane bootstrap 重构为返回错误的 Supervisor，并让各 listener 同步报告 ready；
6. 加入 `control-plane` 和 `limiter-server` Profile，与旧两个二进制做行为对照；
7. 使用同一镜像、不同 mode 验证独立 workload 的 node-id、advertised endpoint、长流排空、滚动升级和故障隔离；
8. 最后按兼容发布策略开放三模块 Profile，稳定后停止 sibling 仓库独立发版。

## 验收标准

- 每个 mode 的 Module 集合、监听端口和外部依赖均有表驱动测试；
- CLI mode 优先于 YAML，未知 mode 和不完整 Limiter 配置 fail closed；
- listener 绑定失败、注册失败、运行期异常和 Stop 超时均有确定性错误；
- `all` 的部分启动失败会完整回滚，进程无残留 listener/goroutine；
- split 与 combined 模式通过相同的 Limiter gRPC/注册 contract tests；
- Limiter 可在同一测试进程中重复构造和销毁，不依赖全局 `sync.Once`；
- 完成 `node-id` 唯一性、长流重连、滚动升级、OOM/CPU 隔离和端口暴露验证；
- Limiter 全仓 Go test、vet、race 与 Control Plane 全仓测试通过。

## 相关页面

- [[architecture]]
- [[configuration]]
- [[adr-console-limiter-source-layout-and-embedded-web]]
