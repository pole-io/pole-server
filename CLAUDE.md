# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 常用命令

```bash
# 构建二进制
make build

# 运行所有测试
go test ./...

# 运行单个包的测试
go test ./pkg/service/...

# 运行单个测试用例
go test -run TestCreateService ./pkg/service/

# 代码格式化（import 排序 + go fmt）
bash ./import-format.sh

# 本地启动（需要 MySQL，连接信息通过环境变量注入）
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  go run . start -c ./test/data/bootstrap/pole-server.yaml
```

### 测试环境变量

MySQL 集成测试需要设置：
```bash
export STORE_MODE=sqldb
export MYSQL_DB_USER=root
export MYSQL_DB_PWD=12345678
export MYSQL_HOST=127.0.0.1:3306
```

## 代码架构

### 四层结构

```
apis/          → 接口定义层（Plugin、Store、Cache、Auth、Apiserver 接口）
pkg/           → 业务逻辑层（service、config、namespace、goverrule、admin）
plugin/        → 插件实现层（httpserver、grpcserver、xdsserverv3、nacosserver、store/mysql、access_control）
bootstrap/     → 启动编排层（初始化顺序、配置加载、自注册）
```

### 插件注册模式

所有可扩展组件通过 `init()` 自注册，`plugin.go`（根目录）通过 blank import 触发所有插件的 `init()`。新增插件须在 `plugin.go` 中补充 blank import。

```go
// 注册示例
func init() {
    apis.RegisterPlugin("myPlugin", &myConcrete{})
}
```

### 业务服务单例模式

每个业务域（service、config、namespace 等）都遵循同一模式：`Initialize()` 初始化单例，`GetServer()` 获取实例，auth/paramcheck 拦截器通过包装器模式注入到调用链前。

### 缓存同步机制

`pkg/cache/` 每秒轮询一次数据库（增量查询 `mtime > lastMtime - 5s`），`flag=1` 表示软删除。业务层读操作走缓存，写操作直接落库。

### Store 接口分解

`apis/store/store.go` 中 `Store` 接口由 `DiscoverStore`、`ConfigFileStore`、`AdminStore`、`AuthStore`、`AIStore` 组合而成，MySQL 实现在 `plugin/store/mysql/`，Mock 实现在 `plugin/store/mock/`（供单元测试使用）。

### AI 原生功能

- **MCP Registry**：`plugin/apiserver/httpserver/aimcp/` — 将 Pole 的服务管理能力暴露为 MCP 工具

### 知识库

`context-kg/` 目录包含本项目的结构化知识库 wiki，按分层目录组织：
- `_meta/`：schema、index（入口）、log
- `overview/`：架构、API 服务端、配置
- `domains/`：各业务域独立文件
- `infra/`：存储、缓存、认证、公共基础设施
- `ai/`：AI 原生功能
- `guides/`：代码模式、测试

回答架构问题时优先读取 `context-kg/_meta/index.md` 定位相关页面。

## Import 格式约定

使用 `goimports-reviser` 管理 import 分组顺序，公司前缀为 `github.com/pole-io/specification`，项目名为 `github.com/pole-io/pole-server`。直接运行 `bash ./import-format.sh` 一键处理。
