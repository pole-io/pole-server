---
title: 测试
tags: [testing, mock]
links: [storage, domain-components, patterns]
updated: 2026-05-14
sources: 1
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

各业务域的测试覆盖范围见 [[domain-components]]。

## 运行测试

```bash
go test ./...                          # 运行全部测试
go test ./pkg/service/...              # 仅运行 service 包
go test -run TestCreateService ./pkg/service/
```

## 关键测试文件

- `pkg/config/utils_test.go` — 配置工具函数测试
- `pkg/cache/ai/mcp_server_test.go` — MCP Server 缓存测试

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
- [[domain-components]]
- [[patterns]]
