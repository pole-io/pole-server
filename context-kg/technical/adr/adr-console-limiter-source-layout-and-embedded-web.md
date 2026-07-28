---
title: Console、Limiter 源码归属与前端嵌入制品
tags: [adr, architecture, console, limiter, frontend, embed, build]
links: [architecture, adr-unified-process-mode-and-limiter-integration]
updated: 2026-07-29
sources: 12
---

# Console、Limiter 源码归属与前端嵌入制品

## 状态

已实施。

## 背景

Console 与 Limiter 已由统一 Supervisor 装配，并随 Control Plane 使用同一个 Go Module、二进制、镜像和版本发布。当前源码仍保留迁入前的顶层形态：

- `console/` 同时包含 Go 网关、Console 私有 Go 包和 React/Vite 前端；
- `limiter/` 同时包含运行 Module、协议 Adapter、限流核心和自身的 `pkg/`、`plugin/` 层次；
- 根 `bootstrap` 直接依赖 `console/bootstrap`、`console/pkg/poleagent` 和 `limiter/apiserver`，Module 的内部 Seam 泄漏到进程组合根；
- Console 前端以外部 `webPath` 目录提供静态资源，运行正确性依赖工作目录、构建产物和 hash 文件保持同步。

独立启动或独立部署是运行拓扑，不应自动决定源码必须与 `pkg/` 并列。源码布局需要表达产品 Module 的归属、外部 Interface 和实现可见性。

## 决策

### Go 产品 Module 统一归入 `pkg/`

Console Go 网关与 Limiter 都迁入 `pkg/`：

```text
pkg/
├── console/
│   ├── config/
│   │   └── config.go
│   ├── module.go
│   ├── assets.go
│   └── internal/
│       ├── router/
│       ├── handlers/
│       ├── observer/
│       ├── observabilityquery/
│       ├── poleagent/
│       └── systemsettings/
└── limiter/
    ├── config.go
    ├── module.go
    ├── registry.go
    └── internal/
        ├── apiserver/
        ├── ratelimitv2/
        ├── statistics/
        └── utils/
```

`pkg/console` 和 `pkg/limiter` 分别提供小而深的外部 Interface。根 `bootstrap` 只负责选择 Profile、注入配置和交给 Supervisor，不得直接依赖 Module 的 `internal` Implementation 或协议 Adapter。

两个 Module 的生命周期 Interface 与 [[adr-unified-process-mode-and-limiter-integration]] 保持一致：

```go
type Config struct {
    // Module 完整启动配置。
}

func Start(ctx context.Context, cfg Config) (runtime.Running, error)
```

Interface 不变量：

- `Start` 只在必要 listener 完成同步绑定并达到 readiness 后返回；
- 运行期错误通过 `Running.Wait` 上报；
- `Running.Stop` 幂等并接受截止时间；
- 配置校验、内部依赖构造、路由、存储和协议细节隐藏在 Module 内；
- 监听端口预检通过根 Config 或只读 bindings 描述完成，bootstrap 不穿透协议 Adapter。

### Console 前端源码独立归入 `web/console`

React/Vite 前端不属于 Go 包，迁入顶层前端源码目录：

```text
web/
└── console/
    ├── src/
    ├── public/
    ├── scripts/
    ├── package.json
    └── package-lock.json
```

`web/console` 只表达源码和工具链归属，不建立独立版本或发布生命周期。前端始终由仓库统一 Go 制品流程构建。

### Release/Test 使用 `go:embed`

Release 与 Test 构建执行：

```text
web/console
  → npm ci
  → npm run build
  → 复制 dist 到 Go 包内的生成目录
  → go:embed
  → Go 二进制、镜像和发布包
```

Go 的 embed pattern 不能引用父目录，因此不得从 `pkg/console` 使用 `../../web/console/dist`。构建阶段将前端产物复制到：

```text
pkg/console/internal/assets/dist/
```

随后由 `pkg/console` 嵌入并通过 `fs.Sub` 交给静态文件 Handler。生成目录不作为人工维护源码；构建脚本必须先清理并重新生成，避免旧 hash 文件混入制品。

Release 模式下：

- `index.html`、`assets/*` 和 SPA fallback 均从嵌入文件系统读取；
- 不再依赖运行时工作目录或外部 `webPath`；
- 缺少前端构建产物时 Go 制品构建必须失败；
- 任何静态资源更新都必须产生新的 Go 二进制或镜像。

### 本地开发保留 Vite 热更新

本地开发不强制每次修改 TSX 后重新构建 Go：

- Vite dev server 提供前端页面和 HMR；
- API 请求代理到本地 Console/Control Plane；
- Go Console 仍可独立启动用于后端调试；
- Release/Test 不允许回退到 Vite 或外部静态目录。

开发模式与制品模式是同一功能的两个 Adapter：前者优化反馈速度，后者保证单文件交付和运行确定性。两者必须共享路由基路径和 API 契约，并由自动化测试覆盖 SPA 深链与静态资源访问。

## 共享类型与 Seam 收口

`bootstrap/self_management.go` 只通过 `pkg/console` 根包使用 A2A Agent Card 等 Console 启动契约；模型调用、工具会话和工作台逻辑保留在 `pkg/console/internal`，组合根不再导入内部实现。

Console 配置从 `console/bootstrap.Config` 收口为 `pkg/console.Config`。配置加载和系统配置来源描述可以在根 bootstrap 组装，但 Console 内部默认值、校验和运行时构造由 Console Module 自己负责。

Limiter 对外只保留 `Config`、`Start` 和 `Running`。原 `limiter/pkg/*`、`limiter/plugin/*` 不机械迁成 `pkg/limiter/pkg/*`、`pkg/limiter/plugin/*`，而是按职责收敛到 `pkg/limiter/internal`。

## 不采用的方案

### 整体保留顶层 `console/`、`limiter/`

该方案延续迁入前的源码形态，但会继续把部署拓扑误写成源码归属，并让仓库出现两套业务实现层次。

### 将前端放入 `pkg/console/web`

前端不是 Go Implementation；放入 `pkg` 会混淆 Go 包和 Node 工程的职责，也不利于未来在 `web/` 下组织其他浏览器应用。

### 整体迁入 `internal/console`、`internal/limiter`

Go 可见性最严格，但与本仓库现有 `pkg/service`、`pkg/config` 等产品 Module 约定不一致。局部 `internal` 已足以保护实现 Seam。

### Release 继续依赖外部 `webPath`

外部目录允许不重编译 Go 即替换前端，但会产生二进制与静态资源版本漂移、工作目录依赖和 hash 文件缺失问题，不符合统一单文件制品目标。

## 迁移顺序

1. 为 Console 与 Limiter 的现有启动、readiness、停止、SPA fallback 和静态资源行为建立回归测试；
2. 将共享 A2A/Agent 类型上提到 `apis/pkg/types/ai`，收口 `pkg/console.Config` 与生命周期 Interface；
3. 将 Limiter 迁入 `pkg/limiter`，同时扁平化历史 `pkg/`、`plugin/` 目录；
4. 将 Console Go 代码迁入 `pkg/console`，内部实现改为 `internal`；
5. 将前端源码迁入 `web/console`，同步更新 CI、Dependabot、脚本、E2E 和知识库路径；
6. 建立确定性的前端构建、产物复制和 `go:embed` 流程；
7. 删除 Release 运行时 `webPath`，保留显式 Vite 开发入口；
8. 完成全仓测试、前端专项检查、单二进制无外部静态目录启动和真实浏览器验证。

## 实施结果

- Console 生命周期收口为 `pkg/console.Start → Running.Wait/Stop`，Limiter 生命周期收口为 `pkg/limiter.Start → Running.Wait/Stop`；
- 根 `bootstrap` 只导入两个 Module 根包，协议服务、路由、Observer、Pole Agent、限流核心和统计实现均由局部 `internal` 隐藏；
- 前端源码位于 `web/console`，`scripts/build-console-assets.sh` 负责 release/test 构建、清理旧产物并复制到嵌入目录；
- `pkg/console/assets.go` 使用 `go:embed`，首页、hash 静态资源和 SPA fallback 统一从二进制内文件系统读取；
- YAML 中的 `webPath` 已删除；发布包和 Kubernetes 镜像不再复制外部 Console dist；
- Vite dev server 保留 HMR，并把 Console/API 路径代理到本地 `8080`。

## 验收标准

- `bootstrap` 只导入 `pkg/console` 和 `pkg/limiter` 的根 Interface；
- 仓库内没有 `pkg/console/pkg`、`pkg/limiter/pkg` 或对 Module `internal` 的越层依赖；
- Release/Test 构建缺少前端产物时失败，不静默生成无 Console 的二进制；
- 最终二进制在没有外部静态目录时可访问首页、hash 资源和 SPA 深链；
- Vite 开发模式保留 HMR，并将 API 请求正确代理到本地后端；
- `all`、`full`、`console`、`control-plane`、`limiter-server` 的运行语义不因源码迁移改变；
- Docker、Kubernetes、发布包、E2E 和本地重建脚本不再引用旧 `console/web`、`console/pkg`、`limiter/` 路径。

## 相关页面

- [[architecture]]
- [[adr-unified-process-mode-and-limiter-integration]]
