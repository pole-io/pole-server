---
title: 测试
tags: [quality, automation, testing, mock]
links: [storage, index, patterns, console-client-auth-e2e-testcases, console-ui-quality-gates]
updated: 2026-07-29
sources: 4
---

# 测试

## 测试结构

```
test/
├── suit/         # 使用真实组件的集成测试套件
├── data/         # 测试固件
│   ├── bootstrap/ # 测试配置文件
│   └── xds/      # xDS 测试数据
├── listener/     # 测试事件监听器
├── subscriber/   # 测试订阅者
├── topology/     # 服务拓扑测试
└── cluster/      # 集群模式测试
```

单元测试与源码同目录存放：`pkg/config/utils_test.go`、`pkg/cache/ai/mcp_server_test.go` 等。

## Mock Store

`plugin/store/mock/` 包含通过 `golang/mock` 生成的完整 `store.Store` 接口 Mock。用于单元测试，避免依赖 MySQL。存储层接口的完整定义见 [[storage]]。

```go
// 在测试中的用法
ctrl := gomock.NewController(t)
mockStore := mock.NewMockStore(ctrl)
mockStore.EXPECT().CreateMCPServer(gomock.Any()).Return(nil)
```

## Mock Auth（`plugin/access_control/auth/mock/`）

`UserServer` 和 `StrategyServer` 的 Mock 实现，用于在不引入认证开销的情况下测试业务逻辑。

## 测试套件（`test/suit/`）

集成测试套件，将真实组件串联在一起：
- 使用 Mock Store 作为后端（无需 MySQL）
- 初始化完整的服务端栈（namespace、service、config 等）
- 测试端到端的请求流程

各业务域的测试覆盖范围可从 [[index]] 定位。

## 运行测试

```bash
go test ./...                          # 运行全部测试
go test ./pkg/service/...              # 仅运行 service 包
go test -run TestCreateService ./pkg/service/
```

## 关键测试文件

- `pkg/config/utils_test.go` — 配置工具函数测试
- `pkg/cache/ai/mcp_server_test.go` — MCP Server 缓存测试

## E2E 测试设计

Console API、Client 查询和权限开关的接口端到端测试覆盖矩阵见 [[console-client-auth-e2e-testcases]]。

该方案将测试分为两层：

- Console API E2E：通过 8080 console proxy 验证所有控制台读写请求。
- Client E2E：通过 8090 client API 验证服务发现、治理规则发布态、权限开关和缓存传播。

该 E2E 套件只包含 HTTP/API 维度，自动化入口为 Go test；不在 `test/e2e` 中建设 Playwright、DOM 或截图测试，也不把页面交互断言混入接口套件。

这不表示 Console 发布可以跳过 UI 验收。前端另有三层质量证据：

- `web/console/scripts/verify-*.mjs` 负责共享组件、关键样式和交互绑定的源码契约。
- ESLint、Vite production build 和 Console Go tests 负责构建与后端回归。
- Kubernetes 实际产物和真实浏览器负责路由、布局、主题、交互与请求闭环。

完整范围和验收矩阵见 [[console-ui-quality-gates]]。

接口 E2E 默认通过 build tag 隔离，不会被普通测试命令触发：

```bash
go test ./test/e2e/...                                              # 不启动 Docker，仅验证默认隔离
CGO_ENABLED=0 go test -tags=e2e ./test/e2e/... -run TestDoesNotExist # 仅编译 E2E 套件
CGO_ENABLED=0 go test -tags=e2e ./test/e2e/... -count=1              # 显式执行 E2E
```

在本地 macOS 环境中，testcontainers 依赖链可能经 `github.com/shoenig/go-m1cpu` 触发 cgo 初始化崩溃；接口 E2E 推荐显式设置 `CGO_ENABLED=0`。

## 测试规范

1. 使用 `github.com/smartystreets/goconvey` 进行 BDD 风格断言
2. 使用 `github.com/stretchr/testify` 进行标准断言
3. 使用 `DATA-DOG/go-sqlmock` 在 Store 层测试中模拟 MySQL
4. 使用 `golang/mock` 生成接口 Mock

代码库中常见的测试模式见 [[patterns]]。

## go-sqlmock 使用模式

用于测试 MySQL Store 实现：
```go
db, mock, err := sqlmock.New()
mock.ExpectQuery("SELECT").WillReturnRows(...)
mock.ExpectExec("INSERT").WillReturnResult(...)
```

## 相关页面

- [[storage]]
- [[index]]
- [[patterns]]
- [[console-client-auth-e2e-testcases]]
- [[console-ui-quality-gates]]
