---
title: 任务计划与 Review
tags: [tasks, todo]
links: []
updated: 2026-06-10
sources: 0
---

# A2A Agent Registry 实现

- [x] 创建独立 worktree `codex/a2a-agent-registry`，确认当前分支干净
- [x] 运行基线 `go test ./...` 并记录既有失败
- [x] 补充 A2A Registry 本地类型、Store/Cache 接口和资源类型
- [x] 先写 A2A cache 与 MySQL store 的失败测试
- [x] 实现 `A2AAgentStore`、MySQL schema 与增量查询
- [x] 实现 `A2AAgentCache` 和 `CacheManager` 接入
- [x] 新增 `/ai/a2a/v1` Registry HTTP 管理接口
- [x] 接入 HTTP server 配置 `aia2a`
- [x] 更新 AI Native 文档，明确只实现 Registry，不做数据面代理
- [x] 运行格式化、目标包测试、构建和 diff 检查
- [x] 记录 review、验证结果和剩余风险

范围约束：

- pole-control-plane 只实现 A2A Agent Card 注册、发现、索引、健康/拉取状态、治理元数据和管理 API。
- 不实现 A2A task proxy、SSE streaming 转发、push notification broker、task 状态机和 artifact 存储；这些属于数据面、网关或 agent runtime。
- 当前仓库依赖的 `github.com/pole-io/specification` 尚无 A2A 类型；本轮先在 control-plane 内定义轻量本地类型，后续可迁移到 specification。

基线验证：

- `go test ./...` 未完全通过，失败点为既有环境/用例问题：
  - `pkg/service/healthcheck` 缺 `test/data/service_test.yaml`。
  - `plugin/apiserver/httpserver/i18n` 缺 `release/conf/i18n/*.toml`。
  - `plugin/service/healthchecker/heartbeat` 测试触发 nil pointer。
- 与本轮相关的基线包 `pkg/cache/ai`、`plugin/apiserver/httpserver/aimcp`、`plugin/store/mysql` 通过。

当前进展：

- 新增 `apis/pkg/types/ai` 本地 A2A Registry 类型，覆盖 Agent、Interface、Skill、Query、DeleteRequest 和列表响应。
- `AIStore` 组合新增 `A2AAgentStore`，`CacheManager` 新增 `A2AAgent()`。
- 新增 `pkg/cache/ai/a2a_agent.go`，支持按 ID、`namespace/name`、namespace 和 skill tag 查询。
- 新增 `plugin/store/mysql/a2a_agent.go` 和 `a2a_agent/a2a_agent_interface/a2a_agent_skill` DDL。
- 新增 `plugin/apiserver/httpserver/aia2a`，暴露 `/ai/a2a/v1/agents`、`/agents/delete`、`/agent/skills`、`/agents/{id}/card` 等 Registry 管理接口。
- `plugin/apiserver/httpserver/server.go` 按 `aia2a` API 配置懒初始化 A2A server，未启用时不打开 A2A cache，避免旧部署无新表时启动失败；`deploy/conf` 默认关闭，测试 bootstrap 默认开启。
- 新增 Console A2A 页面与数据流：
  - `console/web/src/services/a2a.ts` 对接 `/ai/a2a/v1` REST API。
  - `console/web/src/modules/ai/a2a.ts` 管理列表、编辑对象、skills 和 card 状态。
  - `console/web/src/pages/AI/A2A/index.tsx` 替换未实现页，提供列表筛选、创建/编辑、删除、查看 Agent Card 和查看 Skills。
  - `/ai/a2a` 菜单从隐藏状态改为可见，文案改为 A2A Agent / A2A Agents。
- `context-kg/ai/ai-features.md`、`context-kg/_meta/index.md`、`context-kg/_meta/log.md` 已归档 A2A Registry 设计与范围边界。
- TDD 红绿记录：
  - `go test ./pkg/cache/ai -run 'TestA2AAgentCache' -count=1` 先因缺类型失败，补实现后通过。
  - `go test ./plugin/store/mysql -run 'TestNewA2AID|TestA2AAgentStore' -count=1` 先因缺 store 失败，补实现后通过。
  - `go test ./plugin/apiserver/httpserver/aia2a -run 'TestParseA2A|TestNewA2A|TestStartA2A' -count=1` 先因缺 handler/cache 启动函数失败，补实现后通过。

Review：

- 官方 A2A 规范将 Agent Card 用于 Agent 发现，描述身份、capabilities、skills 和交互要求；本轮实现只落 Registry 元数据，不落 message/task/push 数据面。
- `import-format.sh` 首次因本机无 `wget` 失败；改用 `GOPROXY=https://goproxy.cn,direct go install github.com/incu6us/goimports-reviser/v3@v3.9.1` 安装同版本工具后，重新运行脚本通过。脚本引入的非本次范围 import 分组改动已收敛回去。
- 目标验证通过：
  - `go test -count=1 ./pkg/cache/ai ./plugin/store/mysql ./plugin/apiserver/httpserver/aia2a ./plugin/apiserver/httpserver ./apis/...`
  - `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-a2a .`
  - `npm ci` 后 `npm run build:test` 通过。
  - Playwright 打开 `http://127.0.0.1:3003/ai/a2a`，注入本地登录态后确认 A2A 菜单、筛选栏、表格列和新建抽屉渲染正常；后端未启动导致列表请求报错，属于本地验证环境限制。
  - `git diff --check`
  - `rg -n "^(<<<<<<<|=======|>>>>>>>)" .` 无冲突标记。
- 前端 ESLint 未形成有效验证：`console/web` 当前没有 `.eslintrc*` 或 `eslint.config.*`，直接运行 `npx eslint ...` 报 “couldn't find a configuration file”。
- 全量 `go test ./...` 仍未通过，失败点与基线一致：
  - `pkg/service/healthcheck` 缺 `test/data/service_test.yaml`。
  - `plugin/apiserver/httpserver/i18n` 缺 `release/conf/i18n/*.toml`。
  - `plugin/service/healthchecker/heartbeat` 触发 nil pointer。
- 剩余风险：
  - 当前 A2A 类型定义仍在 control-plane 本地，后续应在 `github.com/pole-io/specification` 增加正式 A2A proto 后迁移。
  - `aia2a` 显式启用后要求数据库已包含 `a2a_agent`、`a2a_agent_interface`、`a2a_agent_skill` 三张表。

## 相关页面
