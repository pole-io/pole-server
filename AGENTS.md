# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## 全局协作规则

### 工作流编排

#### 1. 默认进入计划模式

- 对于任何非琐碎任务（3 个及以上步骤，或涉及架构决策），都要进入计划模式。
- 如果事情偏离预期，立刻停止并重新规划，不要硬着头皮继续推进。
- 计划模式不仅用于实现，也要用于验证步骤。
- 提前写出详细规格，降低歧义。

#### 2. 子代理策略

- 积极使用子代理，保持主上下文窗口整洁。
- 将调研、探索和并行分析分派给子代理。
- 对复杂问题，使用子代理投入更多算力。
- 每个子代理只负责一个方向，确保执行聚焦。

#### 3. 自我改进闭环

- 只要用户进行了任何纠正，就把对应模式更新到 `context-kg/tasks/lessons.md`。
- 为自己写下规则，避免重复犯同样的错误。
- 持续严格迭代这些经验，直到错误率明显下降。
- 针对相关项目，在每次会话开始时回顾经验记录。

#### 4. 完成前必须验证

- 在没有证明结果有效之前，不要将任务标记为完成。
- 相关时，对比主分支与当前改动后的行为差异。
- 问自己：“一名资深工程师会批准这个结果吗？”
- 运行测试、检查日志，并展示正确性。

#### 5. 追求优雅方案（保持平衡）

- 对非琐碎改动，先停下来问一句：“有没有更优雅的做法？”
- 如果修复方式显得生硬或临时，就按“基于现在掌握的全部信息，实现更优雅的方案”重做。
- 对简单、明显的修复跳过这一步，不要过度设计。
- 在展示结果前，主动质疑并审视自己的方案。

#### 6. 自主修复缺陷

- 收到缺陷报告时，直接修复，不要让用户手把手带着做。
- 从日志、报错和失败测试入手，然后解决问题。
- 不要求用户额外切换上下文。
- 遇到失败的 CI 测试时，直接去修，不要等人告诉你怎么做。

### 任务管理

- 先写计划：把计划写入 `context-kg/tasks/todo.md`，并使用可勾选条目。
- 验证计划：在开始实现前先完成检查。
- 跟踪进度：随着推进，持续将条目标记为完成。
- 解释变更：在每一步提供高层次变更说明。
- 记录结果：在 `context-kg/tasks/todo.md` 中补充 review 小节。
- 沉淀经验：在收到纠正后更新 `context-kg/tasks/lessons.md`。

### 苏格拉底式提问协议

当需要向用户提问时（例如澄清需求、消除歧义、做设计决策），使用苏格拉底式方法：

- 每一轮提问都使用 AskQuestion 工具。
- 从零开始：不要假设用户已经有清晰答案，要帮助他们从第一性原理出发思考。
- 进行 3 到 5 轮对话：
  - 第 1 轮：询问核心问题或目标，例如“我们到底要解决什么问题？”
  - 第 2 轮：挑战既有假设，基于第 1 轮回答探索约束、边界情况和权衡。
  - 第 3 轮：收敛选项，基于前面的洞察提出具体方案。
  - 第 4 到 5 轮（如有需要）：细化细节并确认理解一致。
- 每轮之后：先总结从用户回答中得到的洞察，再进入下一问。
- 所有轮次结束后：输出一份整合性的结论，综合完整讨论结果。
- 不要一次性抛出所有问题：每轮只问一个聚焦问题，并基于前一轮回答递进。

适用场景：需求含糊、架构决策、复杂权衡、功能范围界定。

不适用场景：简单的是/否确认、明显缺陷、边界清晰的任务。

### 核心原则

- 简单优先：每次改动都尽可能保持简单，尽量减少涉及的代码范围。
- 不偷懒：找到根因，不接受临时性修补，按高级工程师标准执行。
- 最小影响：只改必须改的部分，避免引入新问题。
- 语言要求：所有对话和文档输出必须使用中文。

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

# 一键重新构建 console 静态资源和 control-plane 二进制，并以 all 模式启动
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  ./scripts/rebuild-start-all.sh

# Agent / sandbox 环境中用 detached 模式，避免前台常驻进程阻塞工具调用
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  ./scripts/rebuild-start-all.sh --detach

# 本地启动后端（需要 MySQL，test/data 配置默认 mode=server）
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  go run . start -c ./test/data/bootstrap/pole-server.yaml

# 本地启动完整 all 模式（server + console，console 默认 8080，后端 HTTP 默认 8090）
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  go run . start -c ./deploy/conf/pole-server.yaml --mode all

# 仅启动 control-plane server
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  go run . start -c ./deploy/conf/pole-server.yaml --mode server

# 仅启动 console 网关（需要已有后端，配置里 poleServer.address 指向 127.0.0.1:8090）
MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 \
  go run . start -c ./deploy/conf/pole-server.yaml --mode console
```

### 运行模式

- `start --mode all`：默认产品入口，control-plane server 和 console 一起启动。
- `start --mode server`：只启动 control-plane server，适合后端接口和集成测试。
- `start --mode console`：只启动 console 网关，要求 `bootstrap.console.poleServer.address` 指向一个已运行的后端。
- CLI 的 `--mode` 优先级高于 YAML 中的 `bootstrap.mode`；未显式设置时，代码默认回退到 `all`。

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
- `business/`：业务术语、领域模型、业务规则、功能档案、产品决策、用户洞察
- `technical/`：架构、模块、接口契约、环境部署、技术约定、ADR、技术债
- `quality/`：缺陷复盘、测试用例、自动化测试背景知识
- `tasks/`：任务计划、进度、review 和 lessons

回答架构问题时优先读取 `context-kg/_meta/index.md` 定位相关页面。

### 文档归档硬规则

- 架构设计、技术方案、领域设计、存储/缓存/API 方案等长期知识必须归档到 `context-kg/`，并遵循 `context-kg/_meta/schema.md` 的三域分类、frontmatter、双向链接和日志规则。
- 技术方案、架构决策、存储/缓存/API 设计默认进入 `context-kg/technical/adr/`；业务功能页只保留摘要和链接。
- 业务术语、领域实体、业务规则和功能档案进入 `context-kg/business/`；测试策略、缺陷复盘和自动化测试知识进入 `context-kg/quality/`。
- 不要在 `docs/design/` 或其它 `docs/` 子目录新增长期技术方案文档；已有或新生成的长期知识必须按三域结构归档到 `context-kg/`。
- 新增或调整 `context-kg` 内容时，必须同步维护 `context-kg/_meta/index.md` 与 `context-kg/_meta/log.md`；新增页面还必须保证 `## 相关页面` 与 frontmatter `links` 一致。
- `context-kg/tasks/` 只用于任务计划、进度、review 和 lessons 记录，不承载长期架构知识。

## Import 格式约定

使用 `goimports-reviser` 管理 import 分组顺序，公司前缀为 `github.com/pole-io/specification`，项目名为 `github.com/pole-io/pole-server`。直接运行 `bash ./import-format.sh` 一键处理。
