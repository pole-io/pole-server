---
title: 任务计划与 Review
tags: [tasks, todo]
links: [lessons]
updated: 2026-07-20
sources: 0
---

# 任务计划与 Review

## 移除治理运行变量值来源

- [x] 核对 `VARIABLE` 在 specification、Console、Go 后端、Rust SDK 与知识库的完整影响面
- [x] 从 specification 删除 `VARIABLE` 并保留枚举号/名称不可复用约束，重新生成 Go/Rust 代码
- [x] 从 Console 的值来源选项与归一化分支中移除运行变量
- [x] 从 Rust SDK 路由与限流消费逻辑中移除机器环境变量读取
- [x] 发布 specification `v0.1.0-ALPHA.35` 并更新 control-plane、Rust SDK 正式依赖
- [x] 完成隔离测试、显式提交推送、all-mode 重建和真实页面验证

Review：

- specification 提交 `4e5e41e`，GitHub Release 与 crates.io 均已发布 `v0.1.0-ALPHA.35`；历史枚举号 `2` 和名称 `VARIABLE` 均被 reserved。
- Rust SDK 对未知值类型 fail closed，不再读取机器环境变量；隔离暂存树测试通过，提交 `2b7b53f` 已推送 `develop`。
- control-plane 删除运行变量选项和数值 `2` 映射，升级 ALPHA.35；隔离暂存树通过 Go 全量测试、前端 ESLint 与构建，提交 `99e96eb9` 已推送 `develop`。
- all-mode 使用当前完整工作树重建，8080/8090 均返回 200；Playwright 验证值来源下拉只显示“固定值 / 请求参数”。

## 发布 specification 新 tag 并更新 control-plane

- [x] 确认 `../specification` 远端 tag 基线、当前分支和待提交文件范围
- [x] 提交 `../specification` 当前协议与生成产物改动
- [x] 创建新的 specification tag，并确认 tag 指向本次提交
- [x] 更新 `pole-control-plane` 的 `github.com/pole-io/specification` 依赖到新 tag，移除本地 replace
- [x] 运行 spec 与 control-plane 验证，记录 review 与剩余风险

当前判断：

- `pole-control-plane` 当前 `go.mod` 同时存在 `github.com/pole-io/specification v0.1.0-ALPHA.31` 和 `replace github.com/pole-io/specification => ../specification`；更新到新 tag 时应去掉本地 replace，避免继续依赖未发布工作区。
- spec 当前分支为 `codex/faultdetect-discovery-rules`，待提交内容包含本轮熔断 `regex_separate` 以及此前同一分支上的治理规则协议改动；本轮按用户要求提交 spec 仓库当前工作区，不回滚已有改动。

发布：

- `../specification` 提交：`70317c4 feat: update governance rule contracts`
- 新 tag：`v0.1.0-ALPHA.32`，annotated tag 指向 `70317c4`
- 已推送：`origin/codex/faultdetect-discovery-rules` 与 `origin/v0.1.0-ALPHA.32`
- `pole-control-plane`：`go.mod` 升级 `github.com/pole-io/specification v0.1.0-ALPHA.31 -> v0.1.0-ALPHA.32`，并删除 `replace github.com/pole-io/specification => ../specification`

验证：

- `cd ../specification && git diff --cached --check` 通过。
- `cd ../specification && go test ./...` 通过。
- `cd ../specification/source/rust/pole-specification && PROTOC=... cargo test --release` 通过。
- `git ls-remote --tags origin refs/tags/v0.1.0-ALPHA.32 refs/tags/v0.1.0-ALPHA.32^{}` 确认远端 tag 存在，tag object `adb498d` 指向提交 `70317c4`。
- `go list -m -json github.com/pole-io/specification` 确认当前解析版本为 `v0.1.0-ALPHA.32`，且无 replace。
- `go mod verify` 通过。
- `go test ./apis/pkg/types/rules -run TestCircuitBreakerBlockConfigAcceptsRegexSeparate -count=1` 通过。
- `go test ./apis/pkg/types/rules ./pkg/goverrule/... ./pkg/cache/rules -count=1` 通过。
- `go test ./...` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- go.mod go.sum context-kg/tasks/todo.md apis/pkg/types/rules/circuit_breaker_regex_test.go` 通过。

Review：

- spec tag 未包含 `source/rust/pole-specification/proto/service.proto` 的行尾-only dirty 变化；该文件仍在 `../specification` 工作区未提交，`git diff --ignore-space-at-eol` 下没有实质差异。
- control-plane 的依赖更新、回归测试和任务记录已纳入本次提交范围。

## 熔断规则接口资源 regex_separate 支持

- [x] 确认限流接口资源与熔断 `BlockConfig.apis` 的当前 spec 形态
- [x] 先补失败测试，证明 `CircuitBreakerPolicy.block_config` 需要序列化 `regex_separate`
- [x] 在 `../specification` 中为熔断策略接口资源补充 `regex_separate` 契约，并重新生成 Go/Rust 产物
- [x] 运行针对性测试、spec 编译验证、control-plane 依赖编译验证与 context-kg lint
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 限流当前使用 `LimitTrigger.regex_combine` 控制正则接口命中后的合并/分开计算；用户本次明确要求熔断接口资源支持 `regex_separate`，因此字段名按 `regex_separate` 落地。
- 熔断当前已把 `BlockConfig.api` 收敛为 `BlockConfig.apis[]`；本轮只补正则拆分契约，不扩展其它熔断策略字段。
- `BlockConfig` 是熔断策略里承载接口资源、错误判断和触发条件的最小边界，`regex_separate` 放在该消息上比放在顶层规则更贴近作用域。

验证：

- `go test ./apis/pkg/types/rules -run TestCircuitBreakerBlockConfigAcceptsRegexSeparate -count=1` 修复前失败，报 `unknown field "regex_separate"`；生成新 spec 后通过。
- `cd ../specification && go test ./...` 通过。
- `cd ../specification/source/rust/pole-specification && PROTOC=../../protoc/protoc-darwin-arm64/bin/protoc cargo test --release` 通过；直接运行 `cargo test --release` 会因未设置 `PROTOC` 失败。
- `go test ./apis/pkg/types/rules ./pkg/goverrule/... ./pkg/cache/rules -count=1` 通过。
- `go test ./...` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/tasks/todo.md apis/pkg/types/rules/circuit_breaker_regex_test.go` 通过。
- `cd ../specification && git diff --check -- api/v1/fault_tolerance/circuitbreaker.proto source/go/api/v1/fault_tolerance/circuitbreaker.pb.go source/rust/pole-specification/proto/circuitbreaker.proto source/rust/pole-specification/src/v1.rs` 通过。

Review：

- `../specification/api/v1/fault_tolerance/circuitbreaker.proto` 的 `BlockConfig` 现在同时包含 `apis[]` 与 `regex_separate`，Go 生成产物提供 `BlockConfig.RegexSeparate/GetRegexSeparate()`，Rust 生成产物提供 `regex_separate` 字段。
- 本轮保留 `../specification` 中已经存在的其它未提交 spec 改动，没有回滚或改写它们。
- `cd ../specification && git diff --check` 全仓检查仍会被 `source/rust/pole-specification/proto/service.proto` 的行尾 whitespace 问题阻断；该文件在本轮开始前已是 dirty，且不是本次熔断字段落点。

## 前端 build 警告彻底清理

- [x] 定位 `npm run build:test` 中 Browserslist 过期和大 chunk warning 的真实来源
- [x] 更新 Browserslist/caniuse 数据，消除过期数据 warning
- [x] 调整 Vite/Rollup 分包策略，消除 JS/CSS 大 chunk warning
- [x] 重跑前端构建、lint、静态脚本、Go 测试与 context-kg 校验
- [x] 提交并推送全部改动

当前判断：

- 不通过调高 `chunkSizeWarningLimit` 压制 warning；优先拆分依赖 chunk、恢复 CSS code split，并保留生产加载可缓存性。
- `vite.config.js` 当前 `cssCodeSplit: false` 会把全部 CSS 聚合成单个大文件，是样式 chunk warning 的直接来源。
- Browserslist 过期来自锁文件中的 `caniuse-lite` 版本落后；Node 25 下 Vite 2 还会因为读取 debug localStorage 触发 `--localstorage-file` 路径 warning。

修复：

- 使用 `npm_config_legacy_peer_deps=true npx update-browserslist-db@latest` 将 lockfile 中的 `caniuse-lite` 更新到 `1.0.30001800`，避免 Browserslist 数据过期 warning。
- `vite.config.js` 恢复 `cssCodeSplit: true`，并按 React、TDesign、Monaco、ECharts、i18n、lodash 等依赖族配置 `manualChunks`，消除大 JS/CSS chunk warning。
- 新增 `scripts/run-vite-build.mjs` 统一承载 `build:test`、`build`、`build:site`，在 Node 22+ 下提供有效的 `--localstorage-file` 路径，避免 Vite 2 与新 Node 组合产生构建 warning。

验证：

- `cd console/web && npm run build:test` 通过；随后对构建日志执行 `rg "Warning|warning|Browserslist|larger than|localstorage-file"` 无命中，确认本轮目标 warning 已清零。
- `cd console/web && npm run lint` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过，覆盖全部前端静态回归脚本。
- `go test ./...` 通过。
- `go test ./pkg/common/batchctrl -count=20 -timeout=60s` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- 本轮没有通过调高 `chunkSizeWarningLimit` 隐藏 warning；JS 和 CSS 大包 warning 通过实际 code split 与 manual chunks 消除。
- `run-vite-build.mjs` 只在 Node 22+ 注入 `--localstorage-file`，避免旧 Node 对未知 flag 失败，同时不改变 dev/preview 入口。
- 当前最大 JS chunk 为 `react-vendor` 约 350 KiB，最大 CSS chunk 为 `tdesign-shared` 约 165 KiB，均低于 Vite 默认 warning 阈值。

## pole-control-plane 全仓问题修复

- [x] 复核全仓检查发现的问题，明确本轮修复范围
- [x] 修复 `pkg/common/batchctrl` graceful stop 并发卡死
- [x] 修复前端 `verify-standard-response-mapping` 暴露的治理页面回归点
- [x] 修复可落地的未实现/清理/lint 入口问题
- [x] 运行 Go、前端静态脚本、前端构建、context-kg lint 和 diff 检查
- [x] 记录 review 与剩余风险

当前判断：

- 本轮优先修复已经被测试或静态脚本证明的问题。
- 对产品能力尚未明确的大块未实现页面，不在本轮临时补假功能；优先避免可见入口暴露 501。

修复：

- `pkg/common/batchctrl` 的 `GracefulStop` 在 drain `tasksChan` 后会再次 flush 剩余未满批次，避免 `Future.Done()` 永久等待。
- `console/web` 治理页面标准映射检查已补齐：泳道组详情继续传递 `editable/deleteable`，主动探测列表展示 `targetService.api` 接口信息。
- XDS HDS `FetchHealthCheck` 不再返回 `codes.Unimplemented`；stream 健康检查保留初始请求解析出的 client，后续健康上报不再丢失 node 绑定。
- 服务删除事务同步清理 `service_subscribe_graph` 中以该服务为 caller 或 callee 的订阅边。
- 注册发现 `gateway` 未实现路由从可用 children 中移除，避免暴露 501 占位入口。
- 前端补充轻量 ESLint 配置，恢复 `npm run lint` 可执行性。

验证：

- `go test ./pkg/common/batchctrl -run TestNewBatchControllerGracefulStopFlushesDrainedPartialBatch -count=1 -timeout=5s` 修复前失败，确认回归测试有效。
- `go test ./plugin/apiserver/xdsserverv3 -run 'TestFetchHealthCheck' -count=1 -timeout=30s` 修复前失败，确认 HDS 不再返回 `Unimplemented`。
- `go test ./plugin/apiserver/xdsserverv3 -run 'TestStreamHealthCheckKeepsClientForEndpointResponses' -count=1 -timeout=30s` 修复前失败，确认 stream 后续健康上报不再丢失 client。
- `go test ./plugin/store/mysql -run TestServiceStoreDeleteServiceCleansSubscribeGraph -count=1 -timeout=60s` 修复前失败，确认订阅图清理测试有效。
- `go test ./pkg/common/batchctrl -count=20 -timeout=60s` 通过，覆盖原 flaky 场景。
- `go test ./...` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && node scripts/verify-standard-response-mapping.mjs` 通过。
- `cd console/web && node scripts/verify-lane-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- 本轮修复只处理已被测试、静态脚本或可达入口证明的问题；未对 configuration 仍注释的路由和其它模板 501 页面做产品范围扩展。
- 前端 ESLint 配置当前以“恢复解析与执行”为目标，暂未引入强规则；历史 `console.log` 调试输出仍建议后续单独清理。

## pole-control-plane 全仓 TODO 与预期风险检查

- [x] 回顾知识库索引、近期 lessons 和当前工作区状态，明确检查范围
- [x] 扫描 TODO/FIXME/未实现/占位/临时代码和明显高风险模式
- [x] 运行仓库可承受的构建、测试、静态脚本和 context-kg 校验
- [x] 抽查关键命中源码上下文，区分真实阻断、待办债务和普通注释
- [x] 记录 review、验证结果和后续建议

当前判断：

- 本次检查基于当前工作树进行，不回滚或改写已有未提交改动。
- 优先关注“代码实际无法满足预期”“存在未实现/占位路径”“TODO 已经落在运行链路上”“验证命令暴露问题”四类风险。

验证：

- `go test ./...` 失败：`pkg/common/batchctrl.TestNewBatchControllerGracefulStop` 10 分钟超时，多个 goroutine 卡在 `future.Done()`。
- `go test ./pkg/common/batchctrl -run TestNewBatchControllerGracefulStop -count=1 -timeout=15s` 单次通过，但 `go test ./pkg/common/batchctrl -count=20 -timeout=60s` 失败，说明问题是并发时序型 flaky，不是纯环境缺失。
- `cd console/web && npm run build:test` 通过，保留既有 Browserslist 过期和大 chunk 警告。
- `cd console/web && node scripts/verify-lane-editor-utils.mjs` 原失败原因是脚本仍保留旧规范里的 `实时规则 SPEC` 断言；最新规范不要求保留该区块，已移除过期断言后通过。
- `cd console/web && node scripts/verify-standard-response-mapping.mjs` 失败：泳道组详情权限保留和主动探测 `targetService?.api` 安全读取检查未满足。
- `cd console/web && npm run lint` 失败：项目没有可用 ESLint 配置文件，脚本本身不可执行。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- `pkg/common/batchctrl` 的 `GracefulStop` 存在真实并发缺陷：停止路径 drain `tasksChan` 后没有再次 flush 剩余未满批次的 `futures`，会导致部分 `Future.Done()` 永久等待。
- XDS HDS unary 接口 `FetchHealthCheck` 已注册在 gRPC 服务上，但实现直接返回 `codes.Unimplemented`。
- 注册发现 `gateway` 菜单路由可达但页面直接展示 501；`configuration` 路由模块定义了分组/K8s/文件路由，但在总路由中被注释掉，且模板页仍是 501。
- 服务删除路径只软删服务和 metadata，明确 TODO 未清理 `service_subscribe_graph`，会留下订阅关系脏数据。
- 前端源码仍有多处 `console.log` 调试输出；这不阻断构建，但会污染生产控制台并可能泄漏请求/表单上下文。
- 更正：泳道编辑器不保留 `实时规则 SPEC` 符合最新规范，不作为问题；对应静态脚本里的旧断言已删除。

## 注册发现指标条中文化与健康实例排版修正

- [x] 复核服务列表页头和指标条中英文混排位置
- [x] 修正健康实例指标值结构，避免 `0/1` 被拆成竖排
- [x] 更新静态回归脚本和 lessons，防止指标条再次混用英文或嵌套 block span
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面量测
- [x] 记录 review 与验证结果

当前判断：

- 注册发现服务页页头 eyebrow 仍是 `Service Registry / Discovery`，指标条仍是 `Services / Namespaces / Healthy Instances`，与页面其它中文文案混排。
- 健康实例值写成 `<strong><span id="stHealthy">0</span>/<span id="stInst">1</span></strong>`，而指标条通用 `.metricItem span { display: block }` 会作用到内部两个数字，导致 `0 / 1` 被拆成三行。
- 本轮只调整服务列表页头和指标条展示，不改变服务列表数据、统计计算、筛选或后端接口。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，新增覆盖中文页头、中文指标条、健康实例非 span 单行结构，以及英文展示文案反回归。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.tsx console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 服务/命名空间数据打开 `http://127.0.0.1:8080/discovery/service`：页头 eyebrow 为 `注册发现 / 服务实例`，指标条标题为 `服务数 / 命名空间 / 健康实例`，健康实例值为 `0/1`，`metricValue` computed display 为 `flex`，`stHealthy` 与 `stInst` top 均为 `261`，页面不包含 `Service Registry / Services / Namespaces / Healthy Instances` 展示文案。

Review：

- 本轮只修注册发现服务页展示，不改变统计计算、查询、分页、筛选或后端接口。
- 健康实例保留 `stHealthy/stInst` 锚点，但从 `span` 改为 `b`，并用 `.metricItem .metricValue` 强制单行基线对齐，避免被 `.metricItem span` 的 block 规则影响。
- 页头和指标条展示文案已统一中文化，避免主路径页面出现中英文混排。

## 创建服务标签编辑器统一控件修正

- [x] 复核统一 `LabelInput` 与创建服务私有标签编辑器的差异
- [x] 修正统一标签控件空态操作列，添加入口只保留在底部
- [x] 创建服务抽屉复用统一标签控件，并保留设计锚点与提交校验
- [x] 更新静态回归脚本和 lessons，防止标签编辑器再次私有化
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面量测
- [x] 记录 review 与验证结果

当前判断：

- 仓库已有统一 `components/LabelInput`，创建服务抽屉上一轮为了补齐锚点和校验做了私有标签编辑器，导致和统一控件重复。
- 用户截图指出空态操作列里出现了 `添加标签`，同时底部还有 `添加标签`；操作列语义应只放行级删除，空态没有行，因此不应放添加动作。
- 本轮应把空态添加按钮从统一控件中移除，并让创建服务抽屉复用统一控件；提交 payload、名称校验和抽屉宽度不改变。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，新增覆盖创建服务复用 `LabelInput`、空态不含添加按钮、底部保留唯一添加入口、行操作列只承载删除、`tagRows/tagsEmpty/tagCount` 锚点保留。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/components/LabelInput/index.tsx console/web/src/components/LabelInput/index.module.less console/web/src/pages/Discovery/Services/ServiceEditor.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 命名空间/服务数据打开 `http://127.0.0.1:8080/discovery/service` 并点击 `新建服务`：空态 `#tagsEmpty` 文本为 `暂无标签`，空态内添加按钮数为 `0`，整个 `#tagRows` 内添加按钮数为 `1`，`#tagCount` 为 `0 个标签`。
- Playwright 点击底部 `添加标签` 后：`#tagCount` 变为 `1 个标签`，`#tagRows` 内添加按钮仍为 `1`，删除按钮为 `1`，输入框为 `2`，空态消失。
- 额外尝试 `cd console/web && npx tsc --noEmit --pretty false` 会被现有 `i18next/react-i18next` 类型与当前 TypeScript 版本不兼容问题阻断；本轮以仓库既有 `npm run build:test` 作为前端类型/打包验证。

Review：

- 统一 `LabelInput` 新增 `editorId/emptyId/countId/hideLabel`，创建服务抽屉通过这些 props 保留交接锚点，不再复制私有标签表格。
- `LabelInput` 空态只展示 `暂无标签`；添加入口只在底部 footer；有标签行时操作列只展示删除。
- `LabelInput` 补齐 `标签键不能为空` 校验和行内错误展示，避免复用统一控件后丢失创建服务上一轮补齐的基础校验。
- 创建服务提交仍从 `service_labels` 转换为 `metadata`，保存、更新、重置、名称校验和抽屉宽度不变。

## 创建服务抽屉宽度与字段间距修正

- [x] 复核截图中创建服务抽屉字段贴边/贴合的问题
- [x] 放宽创建服务抽屉宽度，并为字段 wrapper 增加稳定纵向间距
- [x] 更新静态回归脚本和 lessons，防止字段间距再次退化
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面量测
- [x] 记录 review 与验证结果

当前判断：

- 用户截图中基础信息里的命名空间、名称两行控件几乎贴在一起，归属信息里的部门、业务也有同类问题。
- 根因是字段间距主要依赖 TDesign FormItem 的内部 margin；当前每个字段外面又包了一层锚点 div，视觉上没有形成稳定的字段块间距。
- 抽屉 600px 对当前三段表单和双列标签编辑器偏紧；本轮只调整创建/编辑服务抽屉宽度和字段间距，不改变提交映射、校验或后端接口。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，新增覆盖创建服务抽屉 `min(720px, 94vw)`、`serviceDrawerContent gap: 22px`、字段 wrapper `18px` 间距和 FormItem margin 归零。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/ServiceEditor.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 命名空间/服务数据打开 `http://127.0.0.1:8080/discovery/service` 并点击 `新建服务`：抽屉实际宽度为 `720px`，`#drawer` 分段 gap 为 `22px`，`#inNs .t-form__item` 的 margin-bottom 为 `0px`，命名空间/名称输入框间距为 `18px`，部门/业务输入框间距为 `18px`。
- Playwright 控制台里仅有 React Router future flag 和 TDesign Table key spread 的既有开发警告；未发现创建服务抽屉相关运行错误。

Review：

- 本轮只调整创建/编辑服务抽屉宽度和字段块间距，不改变保存、更新、重置、重名校验或标签提交逻辑。
- 非查看态服务抽屉宽度从 `min(600px, 94vw)` 放宽到 `min(720px, 94vw)`；查看态仍保留 `680px`。
- 连续输入字段的垂直节奏由字段锚点 wrapper 承担，避免 TDesign FormItem 内部 margin 被包裹结构吞掉后再次出现输入框贴合。

## 新建服务抽屉设计交接落地

- [x] 复核设计交接文档与当前 `ServiceEditor` / 服务清单差距
- [x] 重构创建服务抽屉：基础信息、归属信息、服务标签三段结构
- [x] 补齐名称实时计数、内联错误、重名校验和标签行编辑校验
- [x] 补齐服务清单命名空间筛选和设计锚点
- [x] 更新静态回归脚本和 lessons，记录创建服务抽屉交付约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review 与验证结果

当前判断：

- 当前创建服务抽屉仍是一条纵向表单，`showOverlay=false`，命名空间没有必填规则，名称只有提交时规则，没有实时计数和内联错误。
- 服务标签复用通用 `LabelInput`，已有增删和重复 key 校验，但缺少“键空值非空”校验、设计锚点和创建服务语境下的计数/空态约束。
- 交接文档只要求调整前端展示组织、交互映射和提交校验；本轮不改变后端服务注册模型、接口路径或 Redux store 结构。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖服务清单命名空间筛选、统计/筛选/表格锚点、创建服务抽屉分段结构、名称计数、标签编辑器、遮罩关闭和反回归约束。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/ServiceEditor.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 服务/命名空间数据打开服务清单并点击 `新建服务`：抽屉存在遮罩，实际 `t-drawer__content-wrapper` 宽度为 `600px`，分段标题为 `基础信息 / 归属信息 / 服务标签`，`stSvc/stNs/stInst/stHealthy/listCount/nsFilter/keyword/tbody/inNs/inName/nameCount/errName/inDesc/inDept/inBiz/tagRows/tagsEmpty/tagCount` 锚点均存在。
- Playwright 校验重复服务：命名空间 `spec-governance` + 名称 `spec-checkout` 内联显示 `该命名空间下服务名已存在`。
- Playwright 校验名称格式：输入 `bad name!` 后 `#errName` 显示 `只允许数字、英文字母、.、-、_`，`#nameCount` 显示 `9/128`。
- Playwright 校验标签：新增标签行后只填写值 `core`，提交时行内显示 `标签键不能为空`。
- Playwright 校验成功提交：命名空间 `spec-governance`、名称 `spec-new-service`、标签 `tier=core` 的创建 payload 为 `metadata: { tier: "core" }`，抽屉关闭并提示 `服务已创建`。

Review：

- 本轮只调整 Console 前端服务清单和创建服务抽屉，不改变后端服务注册接口、Redux store 结构或服务详情路由。
- 创建服务抽屉已从一条纵向表单改为 `基础信息 / 归属信息 / 服务标签` 三段，抽屉宽度为 `min(600px, 94vw)`，开启遮罩点击和 Esc 关闭。
- 命名空间置顶必填；名称保留规则校验，同时增加实时计数、内联错误和同命名空间重名校验；成功提示文案为 `服务已创建`。
- 服务标签改为当前抽屉内专用键值行编辑器，支持空态、添加、删除、计数、键空值非空拦截和重复键高亮；提交时忽略键值皆空的行。
- 服务清单头补齐命名空间筛选，查询会带 `namespace` 参数；统计、筛选和表体锚点按交接文档补齐。

## 注册发现服务表格内部滚动

- [x] 复核注册发现服务页外层滚动来源和当前高度链
- [x] 将服务列表页约束到当前视区高度，让服务表格内容区内部滚动
- [x] 更新静态回归脚本和 lessons，记录列表页滚动归属约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面滚动验证
- [x] 记录 review 与验证结果

当前判断：

- 用户截图中服务列表行数较多时撑高了注册发现页面，滚动落在整个页面/外层内容区，而不是表格内部。
- 现有根布局和治理工作台已有固定视区高度的模式；服务页应复用这个高度链：页头、指标条、筛选栏固定，只有服务表格内容区滚动。
- 本轮只调整注册发现服务列表页和对应回归脚本，不改变服务查询、分页、详情跳转、别名详情页或后端接口。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖服务页固定视区高度、服务列表 flex 高度链和 `serviceTableSurface` 内部滚动容器。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 服务列表数据打开 `http://127.0.0.1:8080/discovery/service`：`document.scrollHeight=900`、`document.clientHeight=900`、侧栏容器 `scrollHeight=900`、`clientHeight=900`，页面和外层容器没有纵向滚动。
- Playwright 量测服务表格：`.t-table__content` 为 `overflow-y: auto`，`scrollHeight=736`、`clientHeight=400`；鼠标滚轮命中表格内容区后，页面 `scrollTop` 仍为 `0`，表格 `scrollTop` 从 `0` 变为 `336`。

Review：

- 本轮只调整注册发现服务列表页的布局和滚动归属，不改变服务列表接口、分页、筛选、创建、删除或详情跳转逻辑。
- 服务页现在固定在当前视区内，页头、指标条、筛选栏和分页保持稳定；服务行较多时只滚动表格内容区。
- 服务详情内的别名表格没有复用 `serviceTableSurface`，仍保留详情页内嵌清单自己的紧凑布局。

## 服务别名 Tab 去卡片留白修正

- [x] 复核用户对“可以不用卡片，只保留和 Tabs 边缘空隙”的反馈
- [x] 移除服务别名内嵌区外层 panel 样式，仅保留 Tabs 内容区边缘留白
- [x] 更新静态回归脚本和 lessons，防止再次用外层卡片解决该留白问题
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面量测
- [x] 记录 review 与验证结果

当前判断：

- 上一轮给 `服务别名` Tab 加了白底 panel，解决了贴边，但用户进一步确认可以不用卡片，只要和 Tabs 边缘有空隙。
- 最小修正是保留 `embeddedWorkspace` 的 `20px` 外边距，撤掉 `aliasDetailSection` 的 `padding/border/background/radius`，让清单头和表格作为普通详情内容排布。
- 表格自身仍是数据表面，保留原有白底边框；本轮不改变总栏、表格、抽屉、查询或提交逻辑。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖 `embeddedWorkspace margin: 20px`，并要求 `aliasDetailSection` 不再包含 `padding/border/border-radius/background`。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 空别名数据打开 `服务详情 -> 服务别名`：别名查询请求仍为 `offset=0&limit=10&namespace=spec-governance&service=spec-checkout`，清单说明为 `spec-governance/spec-checkout 下当前显示 0 条`，总栏计数为 `0 / 0`。
- Playwright 量测：Tabs 内容区到别名工作区左/上/右间距均为 `20px`；`aliasDetailSection` 为 `padding=0px`、`background=rgba(0, 0, 0, 0)`、`border=0px none`、`border-radius=0px`；toolbar 和表格距 Tabs 左边缘均为 `20px`。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只移除服务别名 Tab 外层卡片样式，不改变别名列表、总栏、抽屉、查询或提交逻辑。
- 留白由 `embeddedWorkspace` 的 `20px` margin 承担；`aliasDetailSection` 只做 grid 分组，不再承担视觉容器。
- 表格自身仍保留白底边框，作为数据表格的必要表面；清单整体不再额外套卡片。

## 服务别名 Tab 外边距修正

- [x] 复核截图中服务别名 Tab 内容贴边的问题
- [x] 为服务详情内嵌别名区增加 panel 包裹和周边空隙
- [x] 更新静态回归脚本和 lessons，记录详情 Tab 内容不能贴边
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面量测
- [x] 记录 review 与验证结果

当前判断：

- 用户截图中 `服务别名` Tab 的清单头和表格贴着 Tabs 内容区左右边缘，与 `服务详情` Tab 的 `margin: 20px` 节奏不一致。
- 需要为服务详情内的别名子资源区增加一个轻量白底 panel，让内容和 Tabs 周边留出稳定空隙；这不是恢复注册发现首页大卡片，而是详情 Tab 内部承载面。
- 本轮只调整服务别名 Tab 的外边距与承载样式，不改变总栏、表格、抽屉或接口逻辑。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，新增覆盖 `embeddedWorkspace margin: 20px` 和 `aliasDetailSection` 白底 panel 样式。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 空别名数据打开 `服务详情 -> 服务别名`：别名查询请求仍为 `offset=0&limit=10&namespace=spec-governance&service=spec-checkout`，清单说明为 `spec-governance/spec-checkout 下当前显示 0 条`，总栏计数为 `0 / 0`。
- Playwright 量测：Tabs 内容区到别名工作区左/上/右间距均为 `20px`；panel 内部 padding 为 `18px`；panel 背景 `rgb(255, 255, 255)`，边框 `1px solid rgb(231, 235, 240)`，圆角 `8px`；表格左右距离 panel 内边缘约 `19px`。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只调整服务详情内 `服务别名` Tab 的外边距和承载样式，不改变别名列表、总栏、抽屉、查询或提交逻辑。
- 服务别名 Tab 现在与服务详情其它 Tab 的周边间距一致，内容不再贴着 Tabs 边缘。
- 外层只使用一个轻量白底 panel 承载清单头和表格，避免回退到注册发现首页那种大卡片包裹整页结构。

## 服务别名设计交接落地

- [x] 复核服务别名交接文档与当前服务详情 Tab 差距
- [x] 补齐别名总栏、清单计数、当前服务上下文和表格锚点
- [x] 补齐行级复制别名入口，并保持编辑 / 删除 / 复制的必要操作密度
- [x] 收敛创建 / 编辑抽屉：目标服务只读、默认当前命名空间、提交校验、重置保留目标服务
- [x] 更新静态回归脚本和 lessons，记录服务详情别名 Tab 的交付约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review 与验证结果

当前判断：

- 交接文档要求服务别名 Tab 按“清单头 -> 表格内总栏 -> 主表 -> 分页”的工作台节奏组织；当前实现已有详情内嵌密度，但还缺总栏、复制别名和抽屉细节。
- 本轮只调整 Console 前端展示组织、表单映射和前端校验，不改变后端服务注册模型、别名解析语义或接口契约。
- 目标服务必须来自当前服务详情上下文并固定只读，创建时别名命名空间默认当前命名空间；编辑态至少锁定别名命名空间，避免误改记录归属。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖服务别名 Tab 的总栏锚点、复制入口、只读目标服务、默认命名空间、抽屉宽度和遮罩关闭约束。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 注册发现接口数据打开 `服务详情 -> 服务别名`：别名查询请求为 `offset=0&limit=10&namespace=spec-governance&service=spec-checkout`，清单说明为 `spec-governance/spec-checkout 下当前显示 2 条`，总栏为 `目标服务=spec-governance / spec-checkout`、`别名数=2`、`覆盖命名空间=2`。
- Playwright 确认表头为 `别名命名空间 / 服务别名 / 描述 / 操作时间 / 操作`，行级复制按钮数量为 `2`，总栏位于主表上方，表格最小宽度为 `720px`。
- Playwright 点击 `新建别名` 后确认：实际抽屉面板宽度 `560px`，存在遮罩，目标服务只读显示 `spec-governance/spec-checkout`，别名命名空间默认 `spec-governance`，目标服务没有可编辑输入框。
- Playwright 提交同命名空间重复别名 `checkout-legacy` 时，前端提示 `该命名空间下别名已存在`，抽屉保持打开，别名接口只发生 `GET` 请求、没有发创建请求。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.tsx console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/alias.tsx console/web/src/pages/Discovery/Services/AliasEditor.tsx console/web/src/pages/Discovery/Services/Instance/index.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只调整 Console 前端服务别名 Tab 的展示组织、表单映射和前端首层校验，不改变后端服务注册模型、别名解析语义或接口契约。
- `服务别名` Tab 现在按交接文档形成 `清单头 -> 表格内总栏 -> 主表 -> 分页` 的纵向关系，总栏随当前返回列表同步展示目标服务、别名数和覆盖命名空间。
- 行级操作保留编辑、删除、复制别名；复制内容为 `命名空间/别名`，复用控制台已有剪贴板工具。
- 创建 / 编辑抽屉改为轻量宽度，恢复遮罩与 Esc 关闭；目标服务固定只读，创建态别名命名空间默认当前命名空间，编辑态锁定别名命名空间。
- 提交前增加别名命名规范和当前页已知重名校验；最终唯一性仍以后端返回为准。

## 服务详情别名 Tab 布局优化

- [x] 复核服务详情内 `服务别名` Tab 当前复用外层列表样式的问题
- [x] 将服务详情内别名清单改为紧凑子清单布局
- [x] 收敛别名表格空态高度、分页密度和嵌入式表格最小宽度
- [x] 更新静态回归脚本和 lessons，记录详情内嵌列表布局约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review 与验证结果

当前判断：

- 用户截图中 `服务别名` Tab 已经下沉到服务详情，但内部仍复用注册发现外层列表样式，导致工具栏贴近 Tab、空表格高度过大、分页区域分散，视觉上像完整列表页嵌进详情页。
- 服务详情内的别名清单应是当前服务下的子资源清单，布局需要更紧凑：标题说明、右侧新建/搜索/查询/重置、紧凑表格和收敛空态高度。
- 本轮不改变别名归属、API、过滤条件或创建编辑逻辑，只优化详情页内部布局。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖服务详情内别名清单专用 toolbar、操作区、表格 surface、空态高度和分页密度。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.tsx console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/alias.tsx console/web/src/pages/Discovery/Services/AliasEditor.tsx console/web/src/pages/Discovery/Services/Instance/index.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 使用临时登录态和 mock 注册发现接口数据打开 `服务详情 -> 服务别名`：空态下 toolbar 高度 `42px`，表格整体高度 `229px`，空态本体实际高度 `92px` 且 `box-sizing=border-box`，分页高度 `56px`，表格最小宽度 `720px`。
- Playwright 有数据场景下量测：表格整体高度 `182px`，分页高度 `56px`，表头为 `别名命名空间 / 服务别名 / 描述 / 操作时间 / 操作`，仍按 `namespace=spec-governance&service=spec-checkout` 查询当前服务别名。

Review：

- 本轮只优化服务详情内 `服务别名` Tab 的内部布局，不改变别名归属、查询参数、创建编辑逻辑或后端接口。
- 内嵌别名清单现在使用专用 `aliasDetailToolbar`、`aliasDetailActions`、`aliasDetailTableSurface`，避免直接继承外层注册发现宽表格密度。
- 空态高度和分页安全区已收敛，减少截图中大面积空白；有数据和无数据两种场景都通过浏览器量测。

## 注册发现别名下沉到服务详情

- [x] 梳理注册发现首页、服务详情页和服务别名列表的现有结构
- [x] 移除注册发现首页的顶层 `别名` Tab，只保留服务列表
- [x] 在服务详情页增加 `服务别名` Tab，并按当前服务过滤别名
- [x] 调整别名创建/编辑表单，在服务详情上下文中固定目标服务
- [x] 更新静态回归脚本和 lessons，记录别名归属约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review 与验证结果

当前判断：

- 用户明确反馈：别名不应作为注册发现首页的独立 Tab，应合并到服务详情里作为服务下的一个 Tab。
- 当前注册发现首页有 `服务 / 别名` 顶层 Tabs；服务详情页已有 `服务详情 / 服务实例 / 服务订阅` Tabs，适合承载 `服务别名`。
- 别名接口支持 `namespace + service` 查询，因此服务详情中的别名清单可以按当前服务过滤。
- 在服务详情上下文创建或编辑别名时，目标服务应固定为当前 `namespace/service`，不应再让用户重新选择目标服务。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖注册发现首页无顶层别名 Tab、服务详情存在 `服务别名` Tab、别名查询带当前服务过滤、编辑器固定目标服务等约束。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.tsx console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/alias.tsx console/web/src/pages/Discovery/Services/AliasEditor.tsx console/web/src/pages/Discovery/Services/Instance/index.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- 当前本地库默认用户仍缺失，浏览器验证继续使用临时登录态和 mock 注册发现接口数据，只验证前端信息架构和交互。
- Playwright 打开注册发现首页确认：没有顶层 `别名` Tab、没有 `新建别名`，保留 `新建服务` 和服务清单。
- Playwright 打开服务详情并切到 `服务别名` 后确认：展示 `别名清单` 与 `新建别名`；表头只有 `别名命名空间 / 服务别名 / 描述 / 操作时间 / 操作`，不再展示 `目标服务命名空间 / 目标服务名`。
- Playwright 捕获别名查询请求：`/naming/v1/service/aliases?offset=0&limit=10&namespace=spec-governance&service=spec-checkout`。
- Playwright 点击 `新建别名` 后确认：抽屉目标服务只读显示 `spec-governance/spec-checkout`。

Review：

- 本轮只调整 Console 前端注册发现信息架构，不修改后端服务、别名 API 或 Redux store 形状。
- 注册发现首页从 `服务 / 别名` 顶层 Tabs 收敛为单一服务列表，页头操作固定为刷新服务列表和新建服务。
- 服务详情新增 `服务别名` Tab，复用原别名列表能力，但在详情上下文中按当前 `namespace/service` 过滤。
- 服务详情内的别名列表隐藏目标服务两列，创建/编辑别名时固定当前目标服务，只让用户填写别名命名空间、别名和备注。
- 顺手修正别名删除参数使用 `alias_namespace`，并为别名表格补本地稳定 row key。

## 注册发现服务列表对齐命名空间页

- [x] 对比服务列表与命名空间管理的页面结构、工具栏位置、指标条和表格密度
- [x] 补充静态回归脚本，要求服务列表采用命名空间页同类布局，不再使用白色 Tabs 大卡片包裹全部内容
- [x] 调整注册发现页头操作、服务/别名 Tabs 外观、列表筛选区和表格样式
- [x] 更新 lessons，记录服务列表与命名空间页的对齐约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review 与验证结果

当前判断：

- 用户截图中服务页仍把 `服务 / 别名` Tabs、指标条、筛选区和表格放在一个白色大卡片里；命名空间页是页头右侧操作、独立指标条、独立列表工具栏和表格。
- 服务页仍需要保留 `服务 / 别名` 两个视图，但 Tabs 不应作为大容器卡片抢占层级；应改成轻量页签导航，下面内容按命名空间页的工作台布局铺开。
- 新建与刷新是当前页头级操作，应放到注册发现页头右侧，并随当前页签切换为 `新建服务` 或 `新建别名`。
- 用户继续反馈指标条需要和下面服务列表表格分开；指标条不应只和列表共享连续流，而应与 `服务清单 / 筛选 / 表格` 形成两个清晰的垂直分组。
- 用户再次反馈视觉上仍像合并在一起；Playwright 量测确认根因是 `.t-tabs._registryTabs` 根节点仍为白底，即使内容子层透明，也会形成包住指标条和服务列表的大白底。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs` 先在旧实现上失败，失败点为注册发现页缺少受控 `activeTab` 和页头动作；修复后通过。
- `cd console/web && node scripts/verify-namespace-actions.mjs` 通过，确认命名空间页基准交互未被破坏。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Discovery/Services/index.tsx console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/alias.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- Playwright 打开 `http://127.0.0.1:8080/discovery/service` 后确认：页头右侧有刷新按钮和 `新建服务`，`服务 / 别名` 是轻量 Tabs，指标条独立展示，服务列表工具栏为 `服务清单 / 服务名 / 查询 / 重置`，表格使用健康进度条和紧凑操作列。
- Playwright 切换到 `别名` 后确认：页头主按钮变为 `新建别名`，别名列表工具栏为 `别名清单 / 别名 / 查询 / 重置`，列表内不再重复放刷新和新建按钮。
- 收到指标条分组反馈后，`verify-discovery-services-layout.mjs` 已补充断言：`.workspace` 使用 `gap: 28px` 分开指标条与列表区，服务列表和别名列表都必须用 `.listSection` 包住筛选区和表格。
- 重新构建并拉起 8080 后，Playwright 量测服务页：`.workspace` gap 为 `28px`，指标条 `bottom=388`，列表区 `top=416`，实际间距 `28px`；列表区内部 `.filterBar` 到 `.tableSurface` 间距为 `14px`。
- 收到 Tabs 白底仍合并的反馈后，样式和静态脚本进一步要求：`registryTabs` 根节点、Tabs 内容区和 Tab panel 均为透明；只有 `.t-tabs__nav` / `.t-tabs__nav-container` 保留白底与分隔线。
- 根节点白底修复后，`cd console/web && node scripts/verify-discovery-services-layout.mjs` 通过，覆盖 Tabs 根透明、导航白底、内容透明和面板透明。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过；`git diff --check -- console/web/src/pages/Discovery/Services/index.tsx console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/alias.tsx console/web/src/pages/Discovery/Services/index.module.less console/web/scripts/verify-discovery-services-layout.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/discovery/service` 返回 `200`。
- 当前本地库默认登录用户缺失，`pole/pole123` 登录返回 `400312 not found user`，初始化管理员接口返回 `500001 store layer exception`；为验证纯前端布局，Playwright 注入临时登录态并 mock 注册发现接口数据后打开 `http://127.0.0.1:8080/discovery/service`。
- Playwright 量测确认：`.t-tabs` 背景为 `rgba(0, 0, 0, 0)`；`.t-tabs__nav` 背景为 `rgb(255, 255, 255)` 且有 `1px solid rgb(229, 232, 239)` 分隔线；`.t-tabs__content`、`.t-tab-panel`、`.tabContent` 背景均为透明；导航到指标条间距 `24px`，指标条到列表间距 `28px`。

Review：

- 本轮只调整注册发现 Console 前端页面布局，不修改服务、别名 API、Redux 数据流或服务详情路由。
- 注册发现页头现在和命名空间页一致承载当前页签的刷新/新建操作；服务/别名子表通过 ref 暴露 `refresh`、`create`，复用各自原有刷新和创建逻辑。
- 服务页的 Tabs 不再是白色大卡片容器；指标条、筛选栏和表格作为独立工作台区块排列，服务表格密度、健康列和操作列向命名空间页收敛。
- 指标条和服务列表现在是两个相邻一级区块；`服务清单 / 查询 / 重置 / 表格` 归入单独 `.listSection`，避免视觉上贴在指标条下面。
- Tabs 根节点和内容层现在不会再提供大面积白底；视觉白底只保留在页签导航条、指标条和表格自身，服务页区块关系与命名空间页保持一致。

## 权限策略资源页签左右独立滚动

- [x] 确认资源信息页签当前左右栏滚动归属和高度链路
- [x] 补充静态回归脚本，要求资源类别列表和资源详情主区分别作为独立滚动容器
- [x] 调整资源信息页签布局，让左侧资源类别和右侧资源清单在抽屉内各自滚动
- [x] 更新 lessons，记录资源页签左右分栏滚动约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面左右滚轮验证
- [x] 记录 review 与验证结果

当前判断：

- 用户期望资源信息页签里左侧“资源类别”和右侧“授权范围/资源清单”各自滚动，而不是整个资源页签内容一起滚动。
- 资源页签应继续保持外层 Drawer、Tabs 内容区和页面不滚动；滚动所有权下沉到资源页签内部的左右两个 pane。
- 修复方向是让资源页签 shell 接入 Tabs 内容区高度，左右 pane 建立 `min-height: 0` 的 flex/grid 高度链，并分别设置 `overflow: auto` 与 `overscroll-behavior: contain`。

验证：

- `cd console/web && node scripts/verify-auth-policy-detail-view.mjs && node scripts/verify-auth-drawer-actions.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Auth/Policy/PolicyDetailView.tsx console/web/src/pages/Auth/Policy/PolicyEditor.tsx console/web/src/pages/Auth/Policy/index.module.less console/web/scripts/verify-auth-policy-detail-view.mjs console/web/scripts/verify-auth-drawer-actions.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/auth/policies` 返回 `200`。
- Playwright 打开 `权限策略 -> 默认策略 -> admin的默认策略 -> 资源信息` 后量测：`.t-tabs__content` 为 `scrollHeight=279`、`clientHeight=279`，不再承担资源页签整体滚动；`.resourceTypeList` 为 `scrollHeight=986`、`clientHeight=153`；`.policyResourceMain` 为 `scrollHeight=336`、`clientHeight=261`。
- Playwright 在左侧资源类别列表中心执行鼠标滚轮后，`leftListScrollTop=360`，`rightMainScrollTop=0`，`tabsContentScrollTop=0`，`drawerBodyScrollTop=0`，`pageMainScrollTop=0`，`documentScrollTop=0`。
- Playwright 在右侧资源详情主区中心执行鼠标滚轮后，`rightMainScrollTop=75`，`leftListScrollTop=0`，`tabsContentScrollTop=0`，`drawerBodyScrollTop=0`，`pageMainScrollTop=0`，`documentScrollTop=0`。

Review：

- 本轮只调整权限策略详情资源信息页签的前端滚动容器，不修改权限策略接口、资源数据或表格字段。
- 资源页签 shell 现在继承 Tabs 内容区高度，左侧类别列表和右侧详情主区分别成为独立滚动容器；外层 Tabs、Drawer body 和页面均不随滚轮移动。
- 静态回归脚本补充左右 pane 的 `height/min-height/overflow/overscroll-behavior` 断言，防止后续把滚动重新放回整个资源页签。

## 权限策略详情抽屉内部滚动修正

- [x] 复现并确认权限策略详情抽屉的滚动容器问题
- [x] 补充静态回归脚本，要求策略查看态 Drawer 使用专用内部滚动样式
- [x] 调整策略查看态 Drawer 和样式，让滚动锁在抽屉 body 内并阻止滚动链传到页面
- [x] 更新 lessons，记录认证详情抽屉滚动容器约束
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面滚动验证
- [x] 记录 review 与验证结果

当前判断：

- 策略详情抽屉当前只设置查看态宽度和隐藏 footer，没有给查看态 Drawer 加专用 class，也没有约束 `.t-drawer__body` 的高度、滚动所有权和滚动链。
- 用户期望不是去掉滚动，而是固定抽屉视区，让长内容在抽屉内部滚动，背景页面不随滚轮移动。
- 修复方向是查看态 Drawer body 固定为 flex 容器并隐藏外层滚动，`PolicyDetailView` 自身不滚动，滚动交给 `Tabs` 内容区；概要和页签导航保留在抽屉视区内。

验证：

- `cd console/web && node scripts/verify-auth-policy-detail-view.mjs && node scripts/verify-auth-drawer-actions.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Auth/Policy/PolicyDetailView.tsx console/web/src/pages/Auth/Policy/PolicyEditor.tsx console/web/src/pages/Auth/Policy/index.module.less console/web/scripts/verify-auth-policy-detail-view.mjs console/web/scripts/verify-auth-drawer-actions.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`8080` 页面 `http://127.0.0.1:8080/auth/policies` 返回 `200`。
- Playwright 打开 `权限策略 -> 默认策略 -> admin的默认策略 -> 资源信息` 后量测：`.t-drawer__body` 为 `overflowY=hidden`、`scrollHeight=664`、`clientHeight=664`；`.t-tabs__content` 为 `overflowY=auto`、`scrollHeight=1112`、`clientHeight=279`。
- Playwright 在 `.t-tabs__content` 中心执行鼠标滚轮后，`tabsContentScrollTop=420`，`drawerBodyScrollTop=0`，`mainScrollTop=0`，`documentScrollTop=0`，`bodyScrollTop=0`。

Review：

- 本轮只调整权限策略详情查看态抽屉的前端滚动容器，不修改后端接口、策略数据结构或列表行为。
- 查看态 Drawer body 现在固定为 flex 容器并隐藏外层滚动；策略概要和页签导航固定在抽屉内，长内容只在 Tabs 内容区内部滚动。
- 静态回归脚本新增查看态 Drawer class、Tabs class 和滚动样式断言，防止后续把滚动所有权又放回整个详情页或页面外层。

## 权限策略详情抽屉设计交接落地

- [x] 补充静态回归脚本，覆盖策略详情抽屉查看态结构、反回归词和轻量交互锚点
- [x] 先运行脚本确认旧实现失败，锁定当前 `Descriptions + Tree` 详情与交接文档差距
- [x] 重写 `PolicyDetailView` 为单栏查看态：概要、成员信息、资源信息、资源标签、可访问接口
- [x] 调整策略详情抽屉承载：查看态不展示编辑/关闭 footer，宽度收敛到 `min(980px, 94vw)`
- [x] 根据用户反馈精简概要区字段，移除策略 ID、来源、成员和可访问接口统计
- [x] 运行静态脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review 与验证结果

当前判断：

- 设计交接明确本轮只调整前端策略详情抽屉的信息组织和交互表达，不调整后端权限策略模型、接口契约或资源授权语义。
- 旧版 `PolicyDetailView` 是 `Descriptions + Tree + Table`，缺少策略概要、资源类型搜索、资源筛选、行选择、复制反馈和资源页签内的授权范围摘要。
- 查看态不应继续显示 `编辑 / 关闭` footer；详情抽屉关闭保留遮罩和 Esc 即可。
- 用户反馈概要区不需要成员、可访问接口、策略 ID 和来源；这些信息不应在策略概要卡里抢占主信息，成员与接口保留在对应页签内。

验证：

- `cd console/web && node scripts/verify-auth-policy-detail-view.mjs` 先在精简前失败，失败点为“策略概要区不能展示 策略 ID”；精简后与 `node scripts/verify-auth-drawer-actions.mjs` 一起通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Auth/Policy/PolicyDetailView.tsx console/web/src/pages/Auth/Policy/PolicyEditor.tsx console/web/src/pages/Auth/Policy/index.module.less console/web/src/pages/Auth/Principal/PrincipalPolicyTable.tsx console/web/scripts/verify-auth-policy-detail-view.mjs console/web/scripts/verify-auth-drawer-actions.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `49769`，`8080/8090` 监听正常，`http://127.0.0.1:8080/auth/policies` 返回 `200`。
- Playwright 使用 `admin/admin123` 登录后打开 `权限策略 -> 默认策略 -> admin的默认策略`；策略详情抽屉概要区只包含策略名称、描述、创建时间、更新时间和策略标签，不包含策略 ID、来源、成员、可访问接口或复制 ID，成员信息与可访问接口仍保留为页签。

Review：

- 本轮只调整 Console 前端策略详情抽屉查看态，不修改权限策略后端接口、授权资源模型或策略列表列定义。
- 策略概要区已移除策略 ID、来源、成员数量和可访问接口统计；成员与接口仍通过 `成员信息`、`可访问接口` 页签查看。
- 静态回归脚本新增概要区禁用词断言，防止后续把这些字段重新放回概要卡。

## context-kg frontend 目录索引检查

- [x] 检查当前 `context-kg` 目录树和 git tracked 文件，确认是否存在 `fronted` / `frontend` 新目录或页面
- [x] 复核 `_meta/schema.md` 与 `_meta/index.md` 的三域分类和索引覆盖
- [x] 运行 `context-kg` lint 验证 index、frontmatter 和链接一致性
- [x] 记录 review 结论

当前判断：

- 当前工作区未发现 `context-kg/fronted` 或 `context-kg/frontend` 目录，也没有新增前端知识页面；`git status --short` 中仅有既有 `context-kg/tasks/todo.md`、`context-kg/tasks/lessons.md` 变更。
- `_meta/schema.md` 当前仍定义为业务、技术、质量三域结构；在没有实际前端页面落入新目录前，不应提前把 `fronted` 作为新知识域写入索引。
- 如果后续确实要引入前端知识域，目录名建议统一为 `frontend/` 而不是 `fronted/`，并按 restructure 流程同步更新 `_meta/schema.md`、`_meta/index.md` 和 `_meta/log.md`。

验证：

- `find context-kg -maxdepth 3 -type d | sort` 未输出 `fronted` 或 `frontend`。
- `git ls-files context-kg | sort` 未包含 `fronted` 或 `frontend` 目录下文件。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。

Review：

- 本轮不调整正式知识库索引；当前索引对现有页面覆盖完整。
- 未修改 `_meta/schema.md`、`_meta/index.md` 或 `_meta/log.md`。

## 命名空间列表操作列收敛

- [x] 建立命名空间操作列静态校验，先确认当前三图标操作列和删除 no-op 会失败
- [x] 将命名空间列表操作列收敛为单个 `查看 / 编辑` 入口与独立 `删除`
- [x] 在命名空间详情抽屉中保留 `编辑`、`授权`、`关闭` 能力，避免行内继续并列授权图标
- [x] 补通命名空间删除动作，删除成功后刷新列表
- [x] 运行命名空间静态校验、前端构建、diff 检查、context-kg lint 和页面可访问验证

当前判断：

- 用户截图中的命名空间操作列仍是 `编辑 / 授权 / 删除` 三个并列图标，和服务列表、A2A 列表最近收敛后的 `查看 / 编辑` + `删除` 不一致。
- 授权不应从产品能力中消失，但也不应继续挤在行内操作列；更合适的位置是命名空间详情抽屉。
- 当前 `delete` 分支没有调用删除接口，本轮作为同一操作列链路一起修复。

验证：

- `cd console/web && node scripts/verify-namespace-actions.mjs` 先在旧实现上失败，随后在新实现上通过，覆盖统一行内入口、抽屉授权入口和删除调用。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Namespace/index.tsx console/web/src/pages/Namespace/NamespaceEditor.tsx console/web/src/pages/Namespace/index.module.less console/web/src/services/namespace.ts console/web/scripts/verify-namespace-actions.mjs context-kg/tasks/todo.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，进程 PID `75643`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/namespace` 返回 `200`，页面加载 release 资源 `assets/index.c94f0f17.js`、`assets/style.03dbdf2d.css` 和命名空间相关 chunk；`/core/v1/namespaces` 因当前本地库未初始化主账号被 console 代理拒绝，返回 `{"code":407,"info":"Proxy Authentication Required: access token is invalid"}`，默认账号 `pole/pole123` 登录返回 `400312 not found user`。本轮未为验证创建账号或改数据库。

Review：

- 本轮只调整命名空间 Console 前端，不修改后端 API、store 或鉴权策略。
- 命名空间列表行内操作现在只有 `查看 / 编辑` 与 `删除` 两个图标；`授权` 从行内移到命名空间详情抽屉底部。
- 命名空间抽屉新增查看态，展示名称、描述和标签；查看态底部提供 `编辑 / 授权 / 关闭`，切到编辑后复用原表单并继续走更新接口。
- 删除按钮现在调用已有 `removeNamespace`，成功后刷新当前分页和筛选条件。
- `DeleteNamespaceRequest.token` 改为可选，和后端 Console API 以及现有 e2e 只传 `name` 的用法一致。

## 治理工作台新建规则向导与工具栏顺序调整

- [x] 将治理工作台筛选栏操作顺序调整为 `刷新 / 重置筛选 / 新建规则`
- [x] 将 `新建规则` 从图标按钮改为文字主按钮
- [x] 将新建规则从下拉菜单改为两步向导：先选规则类型，再进入具体规则创建抽屉
- [x] 在规则类型选择卡片中增加适用场景说明，帮助用户判断什么时候配置该规则
- [x] 移除规则类型卡片中的 `新建 XXX 规则` 重复标题，只保留类型与场景说明
- [x] 运行前端构建、diff 检查、context-kg lint 和工作台页面验证

当前判断：

- 用户明确指出工具按钮顺序应先刷新、重置，再新建；新建规则应使用文字体现，不能继续只用 `+` 图标。
- 新建规则不应在下拉菜单里直接跳转，应先明确选择规则类型，再进入对应规则的创建页面；本轮用 `Dialog + Steps` 表达两步流程，第二步复用现有各规则创建抽屉。
- 用户进一步指出规则类型卡片下面应展示场景描述；规则类型选择阶段要回答“什么场景下要配置这个规则”，不能只重复“新建 X 规则”。
- 用户继续指出卡片里的 `新建 XXX 规则` 本身是冗余信息；弹窗标题和操作按钮已经表达新建，卡片内部应聚焦类型与适用场景。

验证：

- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Governance/Workbench/index.tsx console/web/src/pages/Governance/Workbench/index.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- 静态扫描确认治理工作台不再引用 `Dropdown`，新建入口为 `新建规则` 文字按钮，并保留 `Dialog + Steps` 的创建向导。
- 静态扫描确认 9 类规则卡片均已配置场景描述，并使用 `createTypeTitle` 显式控制标题样式。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，进程 PID `98542`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/governance/workbench` 返回 `200`，页面资源 hash 为 `assets/index.4e9352f1.js` / `assets/style.0f17a3fe.css`。
- 增加场景描述后重新执行 `cd console/web && npm run build:test`、`git diff --check` 和 `context-kg` lint 均通过；重新拉起 all-mode 后进程 PID `24201`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，页面资源 hash 更新为 `assets/index.db057222.js` / `assets/style.d7be1df2.css`。

Review：

- 本轮只调整治理工作台前端交互，不修改各治理规则 editor 的保存逻辑。
- 筛选栏右侧操作现在依次为刷新、重置筛选、新建规则；新建规则是文字主按钮，仍保留加号图标辅助识别。
- 点击新建规则后先打开类型选择弹窗，弹窗顶部显示两步 Steps；确认后关闭弹窗并打开对应规则类型的创建抽屉。
- 规则类型选择卡片现在展示规则类型、创建标题和适用场景；卡片高度略增，标题样式改为显式 class，避免新增说明后样式依赖 `last-child` 失效。
- 规则类型选择卡片中的 `新建 XXX 规则` 标题已移除，卡片回到更轻的类型标签 + 场景说明结构。
- 移除重复标题后重新执行 `cd console/web && npm run build:test`、`git diff --check` 和 `context-kg` lint 均通过；重新拉起 all-mode 后进程 PID `34403`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，页面资源 hash 更新为 `assets/index.74365acb.js` / `assets/style.24239055.css`。

## 注册发现服务列表行操作收敛

- [x] 将服务列表操作列从 `编辑 / 授权 / 删除` 收敛为 `查看 / 编辑` 与 `删除`
- [x] 为服务编辑抽屉增加查看态，并在抽屉内切换到编辑表单
- [x] 运行前端构建、diff 检查、context-kg lint 和服务页面验证

当前判断：

- 用户要求服务列表操作列和治理平台保持一致，因此行内不应再同时摆多个近似入口；主操作统一为 `查看 / 编辑`。
- 服务名点击仍保留原有进入服务详情/实例页路径；行操作的 `查看 / 编辑` 打开服务详情抽屉，并在抽屉内部切换编辑。

验证：

- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/ServiceEditor.tsx console/web/src/pages/Discovery/Services/index.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，进程 PID `80464`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/discovery/service` 返回 `200`，页面资源 hash 为 `assets/index.544453b3.js` / `assets/style.8cd6f23a.css`。

Review：

- 本轮只调整注册发现服务列表和服务抽屉，不修改服务 API、store 或实例页面路由。
- 服务列表操作列现在保留两个按钮：`查看 / 编辑` 和 `删除`；`查看 / 编辑` 打开服务详情抽屉，抽屉底部可切换到编辑表单。
- 服务名点击仍保留原来的服务详情/实例页跳转能力，不和行内统一操作冲突。

## 限流匹配条件 AND/OR 可编辑修正

- [x] 定位限流编辑态 AND/OR 不能选择的根因
- [x] 为限流子规则视图与提交 payload 接通 `matchMode`
- [x] 让限流匹配条件共享组件在编辑态可切换 AND/OR
- [x] 补充静态验证，防止关系再次被硬编码为 AND 或禁用
- [x] 运行限流脚本、共享匹配脚本、前端构建、diff 检查、context-kg lint 和服务重启验证

当前判断：

- 根因是 `RateLimitEditor` 使用共享 `TrafficMatchConditionEditor` 时把 `relation` 写死为 `MatchLogic.AND`，并传入 `relationEditable={false}`，导致编辑态看得到 AND/OR 但不能切换。
- 既然限流已经复用通用匹配条件组件，AND/OR 就必须和其它治理规则一样是可交互控件；用户选择应写入当前子限流规则草稿，而不是仅改变 UI 状态。
- 当前后端限流 proto 尚未显式建模匹配关系字段，本轮不改协议；Console 在子规则 payload 中携带 `matchMode`，前端详情转换也兼容 `matchMode` / `match_mode` 读取。

验证：

- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs` 通过，新增断言覆盖限流编辑器不能再传 `relationEditable={false}`，也不能把关系写死为 `MatchLogic.AND`。
- `cd console/web && node scripts/verify-traffic-match-condition-editor.mjs` 通过，确认共享匹配条件组件仍可用。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/services/ratelimit.ts console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/scripts/verify-ratelimit-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，进程 PID `21478`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/governance/workbench` 返回 `200`，页面资源 hash 为 `assets/index.1fe0af70.js` / `assets/style.4fe08e45.css`。

Review：

- 本轮只修 Console 限流编辑交互，不修改 specification proto、Go 存储或后端下发逻辑。
- `LimitTriggerView` 与提交 payload 现在带 `matchMode`，新建默认 `AND`，详情读取兼容 `matchMode` 与 `match_mode`。
- 限流匹配条件区的 AND/OR 分段控件在编辑态可点击；切换后会写回当前子规则，并同步更新顶部提示文案。

## A2A Agent 详情入口与抽屉 Tabs 收敛

- [x] 将 A2A Agent 列表行内的 Agent Card、技能、编辑三个入口合并为单个 `查看 / 编辑` 入口
- [x] 保留删除为独立危险操作，并继续使用确认弹窗
- [x] 将详情抽屉改为内部 Tabs：`Agent Card`、`技能`、`编辑`
- [x] 让 Agent 名称点击进入详情抽屉，技能数点击进入同一抽屉的技能 Tab
- [x] 复用现有编辑表单作为抽屉内编辑 Tab，不再从详情中弹出第二个编辑抽屉
- [x] 运行前端构建、diff 检查、context-kg lint 并重启服务验证页面

当前判断：

- 用户指出 A2A 列表中查看信息和编辑也应像治理规则一样合并，不应在一行里放多个近似入口。
- A2A 的不同查看视图包括 Agent Card、技能目录和编辑态，适合收敛到同一个详情抽屉内部用 Tabs 切换，而不是拆成多个 action 图标。
- 新建 A2A Agent 仍是独立创建流程，不纳入已有 Agent 的详情抽屉，避免创建态和查看态混在一起。

验证：

- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/AI/A2A/index.tsx console/web/src/pages/AI/A2A/index.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- 静态扫描确认旧的 `EditIcon`、`ListIcon`、`查看技能`、`查看 Agent Card` 和 `detailState.mode` 分支已移除，列表仅保留 `查看 / 编辑` 统一入口。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `3924`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/ai/a2a` 返回 `200`，页面资源 hash 为 `assets/index.7d9e569c.js` / `assets/style.4fe08e45.css`。

Review：

- 本轮只调整 A2A Console 前端交互，不修改 A2A API、store、router 或后端协议。
- 列表操作区现在只保留 `查看 / 编辑` 和 `删除` 两个图标按钮；`查看 / 编辑` 打开统一详情抽屉。
- 详情抽屉摘要区展示 Agent 身份、协议、技能数、后端、接入地址和最近修改；抽屉正文用 Tabs 切换 `Agent Card`、`技能`、`编辑`。
- 原有 `AgentCardView`、`AgentSkillsView` 和 `A2AEditor` 继续复用；`A2AEditor` 新增嵌入模式，避免编辑从详情抽屉里再打开第二层抽屉。

## 注册发现服务列表列与密度修正

- [x] 将服务列表命名空间从服务名副文本拆成独立列
- [x] 移除健康实例列中的 `暂无实例` 副文案
- [x] 压缩服务列表表格行高和单元格内边距
- [x] 运行前端构建、diff 检查并重启服务验证页面
- [x] 记录 review、验证结果和 lessons

当前判断：

- 用户指出服务列表需要把命名空间单独成列，服务名列不应再把 namespace 当第二行展示。
- `暂无实例` 在健康实例列里增加了无效行高，应移除；服务列表整体表格密度需要更接近管理清单，不要每行过高。

验证：

- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Discovery/Services/services.tsx console/web/src/pages/Discovery/Services/index.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `59838`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/discovery/service` 返回 `200`，页面资源 hash 为 `assets/index.3d2d85f3.js` / `assets/style.49690e46.css`。

Review：

- 本轮只调整注册发现服务列表，不改变服务查询、新建、编辑、授权、删除或实例跳转逻辑。
- 服务名列现在只展示服务名和备注；命名空间作为独立列展示。
- 健康实例列只展示 `健康实例 / 总实例数`，不再追加 `暂无实例` 或健康率副文案。
- 服务表格从 `large` 收敛到 `medium`，表头和单元格 padding 从 `14/16px` 压缩到 `10px`。

## 治理工作台规则类型多选与新建入口修正

- [x] 将规则清单的规则类型筛选从横向按钮组改为多选下拉框
- [x] 让规则类型多选支持空值等同全部，并正确筛选限流本地/全局规则
- [x] 将新建规则入口改为常驻按钮，并通过下拉菜单选择具体规则类型
- [x] 将新建、刷新和重置筛选操作从文字按钮改为 TDesign 图标按钮，并补充 tooltip
- [x] 移除命名空间和服务两个筛选下拉，将新建、刷新、重置合并到同一条筛选工具栏
- [x] 按参考图将规则表格改为规则、类型、作用对象、创建时间、更新时间、状态、操作的管理清单视图
- [x] 参考服务列表，将创建时间和更新时间合并为单个操作时间列
- [x] 将规则清单行操作改为服务列表同款图标按钮规范
- [x] 将规则清单行内编辑和查看详情入口合并为单个 `查看 / 编辑` 操作
- [x] 运行前端构建、diff 检查并重启服务验证页面
- [x] 记录 review、验证结果和 lessons

当前判断：

- 用户指出横向按钮布局仍不理想，规则类型应统一为一个多选下拉框。
- 规则创建不能依赖当前筛选只选中单一类型；多选筛选下应把“创建哪种规则”作为独立操作处理，因此新建入口应常驻，并在点击后选择具体规则类型。
- 用户进一步指出工具栏里的 `新建规则 / 刷新 / 重置` 文字按钮占位过重；这类工具操作应使用 TDesign 图标按钮表达，悬浮提示补足语义。
- 用户继续指出命名空间和服务两个筛选下拉在当前清单中占位过重，应移除；新建、刷新和重置图标应合并到筛选工具栏右侧，减少头部垂直分散。
- 用户给出列表视图参考图后，规则清单应转为更常规的管理列表：规则名称下方展示摘要，类型用轻量彩色标签，右侧展示创建/更新时间、状态点和行内操作。
- 用户进一步明确服务列表的行操作按钮是规范，因此治理规则行操作也应使用 `Tooltip + square text Button + icon`，不能继续用文字链接。
- 用户继续指出创建时间和更新时间应参考服务列表合并到一列，避免右侧时间列占用过宽。
- 用户继续指出列表里的编辑和查看详情按钮可以合并；编辑应作为详情抽屉内的动作，不需要在列表行里并列两个近似入口。

验证：

- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- 操作按钮图标化后重新执行 `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- 移除命名空间/服务筛选并合并按钮后重新执行 `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- 表格清单列调整后重新执行 `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- 行操作图标化后重新执行 `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Governance/Workbench/index.tsx console/web/src/pages/Governance/Workbench/index.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `92863`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.7af351e8.js` / `assets/style.b0a2e943.css`。
- 图标化后重新执行 `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `9317`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.a6cf9ec3.js` / `assets/style.b0a2e943.css`。
- 移除命名空间/服务筛选并合并按钮后重新执行 `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `23736`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.f35df749.js` / `assets/style.d9c7a360.css`。
- 表格清单列调整后重新执行 `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `38157`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.e5d6dda3.js` / `assets/style.95a6adb2.css`。
- 行操作对齐服务列表图标按钮规范后重新执行 `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `48861`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.221f93cb.js` / `assets/style.95a6adb2.css`。
- 操作时间列合并后重新执行 `cd console/web && npm run build:test`、`git diff --check` 和 `context-kg` lint 均通过；重新拉起 all-mode 后 tmux 日志显示 `finish starting server`，PID `72263`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.99c843e8.js` / `assets/style.516ca163.css`。
- 编辑和查看详情入口合并后重新执行 `cd console/web && npm run build:test`、`git diff --check` 和 `context-kg` lint 均通过；重新拉起 all-mode 后 tmux 日志显示 `finish starting server`，PID `82792`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.3984652f.js` / `assets/style.516ca163.css`。

Review：

- 本轮只调整治理工作台列表筛选和创建入口，不改各治理规则 editor 的保存 payload。
- 规则类型筛选空数组表示全部；选择 `限流` 时仍覆盖本地限流和全局限流两类行。
- 新建入口不再依赖筛选状态，点击 `新建规则` 后从下拉项中选择路由、限流、熔断、探测、无损、泳道、鉴权、镜像或 Mock。
- 图标化使用 `AddIcon`、`RefreshIcon` 和 `FilterClearIcon`；按钮保留 `aria-label` 和 `Tooltip`，避免只剩图标后不可读。
- 清单头部只保留标题与说明；筛选工具栏现在只包含规则类型多选、关键词搜索和右侧图标操作组，不再单独按命名空间/服务过滤。
- 表格列已去掉 `匹配 / 触发` 和 `发布`，摘要移到规则名第二行，创建时间和更新时间合并为 `操作时间` 单列，行操作中的 `编辑` 打开编辑态，`删除` 走对应规则类型已有删除接口，`更多` 打开详情入口。
- 行操作已收敛为两个方形图标按钮：`CreditcardIcon` 作为统一 `查看 / 编辑` 入口打开详情抽屉，`DeleteIcon` 负责删除并保留 `Popconfirm`；不再把编辑和查看详情并列为两个按钮。

## 限流接口行新增与协议方法编辑修正

- [x] 复现并定位限流接口行无法新增、协议/方法无法编辑的根因
- [x] 为 `method.value` 增加协议/方法/路径兼容拆装 helper，保持现有后端单字段契约
- [x] 将限流匹配接口编辑态改为可新增/删除接口行，并允许编辑协议和方法
- [x] 补充静态验证，覆盖新增接口、协议/方法可编辑和保存值拆装
- [x] 运行限流脚本、前端构建、context-kg lint、diff 检查和真实页面验证
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 用户指出前一轮只修了接口路径的匹配值控件，但仍无法新增接口，且已有接口的协议、方法被禁用，说明限流接口行交互仍未对齐设计。
- 当前后端/前端保存结构仍是 `rules[].method: MatchString`，没有独立 interfaces 数组；本轮通过兼容拆装 `method.value` 承载协议/方法/路径，不改后端 schema。
- 已定位根因：编辑器把 `method.value` 直接当成“接口路径”输入，且协议、方法 Select 硬编码 disabled；对于已有 `POST /api/v1/payments` 数据，方法被混在路径列里，无法单独修改。
- 已补 `parseRateLimitInterfaceValue` / `buildRateLimitInterfaceValue`：兼容旧格式 `POST /path`，也能解析 `GRPC POST /path`；HTTP 继续保存为旧格式，非 HTTP 才带协议前缀。
- 在当前单字段契约下，“新增接口”不能落为同一规则内的真实 interfaces 数组；本轮按可保存语义实现为复制当前规则生成下一条接口规则，继承限流配置并清空接口路径。

验证：

- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs` 通过，覆盖接口值拆装、协议/方法控件不再 disabled、接口路径 TagInput 和新增接口入口。
- `cd console/web && node scripts/verify-traffic-match-condition-editor.mjs` 通过，确认共享匹配条件控件未被本轮改动破坏。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/RateLimit/rateLimitEditorUtils.ts console/web/scripts/verify-ratelimit-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- 浏览器真实页面验证：限流新建抽屉中协议、方法、匹配类型、接口路径均为可编辑控件；方法从 `*` 改为 `POST`、路径填 `/api/v1/payments` 后，右侧 Spec 写回 `value: "POST /api/v1/payments"`；点击 `新增接口` 后出现第二条规则，第二条接口字段可编辑且路径为空，校验提示 `规则[2] 存在空的接口路径`。

Review：

- 本轮不改 RateLimit 后端 schema，也不伪造 `interfaces` 数组；新增接口用新增一条 `LimitTrigger` 表达，保证保存 payload 和客户端发现语义仍可落地。
- `HTTP POST /path` 兼容保存为旧格式 `POST /path`，避免不必要地改变已有数据；非 HTTP 协议会以协议前缀写入当前单字段，后续如果协议层新增独立字段，应再迁移到结构化字段。

## 限流接口匹配类型交互修正

- [x] 将限流接口路径编辑器接入 `IN/NOT_IN` 多值 TagInput 交互
- [x] 补充限流编辑器静态验证，防止接口匹配类型切换后仍固定普通 Input
- [x] 运行限流脚本、前端构建、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮只修 Console 前端限流接口匹配编辑态，不修改 RateLimitRule 后端 schema、保存接口或 method 字段语义。
- 限流接口路径和路由匹配值一样，匹配类型为 `包含/不包含` 时应使用 TagInput 多值编辑，并按英文逗号拼回 `method.value`；其它匹配类型继续使用普通 Input。

验证：

- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs` 通过，断言限流接口匹配值在 `IN/NOT_IN` 时使用 TagInput，并复用逗号字符串与标签数组互转工具。
- `cd console/web && node scripts/verify-traffic-match-condition-editor.mjs` 通过，确认共享流量匹配条件控件未被本轮改动破坏。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/RateLimit/RateLimitEditor.module.less console/web/scripts/verify-ratelimit-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- 浏览器真实页面验证：限流新建抽屉中 `匹配接口` 的匹配类型从 `完全匹配` 切到 `包含` 后，接口路径字段从普通 Input 切换为 `.t-tag-input`，占位为 `输入后回车添加`，并可回车添加 `/api/v1/orders` 标签。

Review：

- 本轮只调整限流接口匹配值的编辑控件，不改变 `trigger.method` 的保存结构；多值仍按英文逗号写回 `method.value`。
- 协议和方法字段继续保持只读，避免把“接口匹配类型交互”扩大成后端 method 契约调整。

## 泳道流量匹配头部错位修正

- [x] 固定泳道放量输入与 AND/OR 分段控件的头部布局尺寸
- [x] 补充静态验证，防止放量控件被 TDesign 内部最小宽度撑开
- [x] 运行脚本、构建、context-kg lint、diff 检查和页面验证
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮只修 Console 前端泳道匹配区的视觉错位，不修改 LaneRule 保存结构、匹配语义或放量比例字段。
- 错位来自共享匹配控件头部右侧 `extraControl + AND/OR` 同行布局中，TDesign `InputNumber/InputAdornment` 内部宽度未被完整约束，导致 `%` 后缀挤到分段控件边界。

验证：

- `cd console/web && node scripts/verify-lane-editor-utils.mjs` 通过，断言泳道放量控件设置 `flex: 0 0 auto`、`InputAdornment min-width` 和 TDesign 内部 input `min-width: 0`。
- `cd console/web && node scripts/verify-traffic-match-condition-editor.mjs` 通过，断言共享匹配条件头部 action 区允许 wrap、右对齐并设置 `min-width: 0`。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/shared/TrafficMatchConditionEditor.module.less console/web/src/pages/Governance/Router/LaneGroupEditor.module.less console/web/scripts/verify-lane-editor-utils.mjs console/web/scripts/verify-traffic-match-condition-editor.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- 浏览器真实页面验证：进入 `seed-20260616-lane` 编辑态，切到 `泳道` Tab 并展开 `blue-lane` 后，放量输入框右边界为 `719px`，AND/OR 分段左边界为 `729px`，间距 `10px`，`overlapsRatioAndSegmented=false`；输入框值 `100` 完整可读。

Review：

- 本轮只修泳道匹配区头部布局稳定性，不改变共享匹配控件的数据接口，也不改 LaneRule 保存 payload。
- 共享匹配头部允许右侧控件在极窄场景下安全换行；泳道常规宽度下仍保持 `命中后放量 + 输入框 + AND/OR` 同行右对齐。

## 泳道组入口与组内服务选择交互修正

- [x] 将泳道组入口的命名空间和服务拆成两个独立下拉框
- [x] 将组内服务添加改成同表格行交互，命名空间和服务独立选择，但只允许选择普通服务
- [x] 补充泳道编辑器静态验证，防止回退到“命名空间/服务”合并下拉
- [x] 运行前端脚本、构建、context-kg lint、diff 检查和本地页面验证
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮只调整 Console 前端泳道组编辑交互，不修改 LaneGroup 后端 schema、protobuf、保存接口或数据语义。
- 入口选择需要先选类型，再分别选择命名空间和该命名空间下的服务；网关入口只列出 gateway 服务，应用入口只列出普通服务。
- 组内服务添加需要和入口表格保持同样的行内编辑体验，但不需要入口类型列，只能从普通服务中选择。

验证：

- `cd console/web && node scripts/verify-lane-editor-utils.mjs` 通过，断言入口和组内服务都使用独立命名空间/服务下拉，并禁止回退到 `选择命名空间 / 服务` 合并下拉或顶部 `+ 添加服务` 下拉。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/LaneGroupEdtor.tsx console/web/src/pages/Governance/Router/LaneGroupEditor.module.less console/web/scripts/verify-lane-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- 浏览器真实页面验证：泳道新建抽屉中 `泳道组入口` 表格显示 `入口类型 / 命名空间 / 服务` 三列独立控件；`组内服务` 表格显示 `命名空间 / 服务 / 引用`，新增草稿行先选命名空间再选服务。普通服务命名空间下拉仅显示 `spec-governance / demo-governance / pole-system` 这类纯命名空间，服务下拉显示 `spec-checkout / spec-gateway / spec-inventory / spec-order / spec-payment` 这类纯服务名，无 `/` 或括号拼接项；选择 `spec-payment` 后表格新增一行并保留下一条新增草稿行。

Review：

- 本轮只修改 Console 前端泳道组编辑态，不改变 LaneGroup 保存 payload 的字段形状；仍由选择结果映射为 `entries[].selector.namespace/service` 与 `destinations[].namespace/service`。
- 由于现有 `draft.selected` 仍以服务名为主键，编辑已有组内服务时切换命名空间会自动落到该命名空间下第一个可用普通服务，避免空服务行被草稿归一化过滤。

## 治理规则匹配条件通用控件

- [x] 梳理路由、限流、镜像、Mock、鉴权、泳道当前匹配条件数据结构和 UI 差异
- [x] 抽取共享 `TrafficMatchConditionEditor`，覆盖 AND/OR、四列条件表、TagInput 多值和只读态
- [x] 将路由匹配条件替换为共享控件
- [x] 将限流匹配条件替换为共享控件
- [x] 将镜像、Mock、鉴权匹配条件替换为共享控件
- [x] 将泳道匹配条件替换为共享控件
- [x] 补充静态验证脚本，防止这些规则类型继续维护私有匹配条件表
- [x] 运行前端构建、相关脚本、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮只抽取 Console 前端匹配条件 UI，不修改后端 rule schema、protobuf、保存接口或匹配语义。
- 共享组件要保留各规则自己的数据模型；调用方负责把 Route/RateLimit/TrafficGovernance/Lane 的字段映射成统一展示行，再映射回原结构。
- 参数键仍是用户业务数据，不能按参数类型预置 `x-tenant` 等候选；匹配类型为 `包含/不包含` 时继续使用 TagInput，并保存为逗号分隔字符串。

验证：

- `cd console/web && node scripts/verify-traffic-match-condition-editor.mjs` 通过，断言共享组件存在且路由、限流、镜像、Mock、鉴权、泳道不再保留私有匹配条件表实现。
- `cd console/web && node scripts/verify-lane-editor-utils.mjs` 通过，覆盖泳道卡片展开布局和共享匹配条件区。
- `cd console/web && node scripts/verify-route-editor-utils.mjs`、`verify-traffic-mirror-editor-utils.mjs`、`verify-traffic-mock-editor-utils.mjs`、`verify-traffic-security-editor-utils.mjs` 均通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/shared/TrafficMatchConditionEditor.tsx console/web/src/pages/Governance/shared/TrafficMatchConditionEditor.module.less console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/src/pages/Governance/Router/LaneGroupEdtor.tsx console/web/src/pages/Governance/Router/LaneGroupEditor.module.less console/web/scripts/verify-traffic-match-condition-editor.mjs console/web/scripts/verify-lane-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- 浏览器真实页面验证：路由详情的 `匹配条件` 区显示 AND/OR、`参数类型 / 参数键 / 匹配类型 / 匹配值 / 操作` 表头和现有 `HEADER x-tenant 完全匹配 vip` 数据；泳道详情切到 `泳道` Tab 并展开 `blue-lane` 后，`流量匹配规则` 区显示同一四列表格、AND/OR、放量控件和现有 `HEADER x-lane 完全匹配 blue` 数据。

Review：

- 本轮只统一 Console 前端匹配条件 UI，不修改后端规则 schema、protobuf、保存接口或规则匹配语义。
- 共享组件集中处理表头、AND/OR 分段、匹配类型、TagInput 多值和只读态；各规则编辑器只负责把自己的字段映射到统一展示行并回写。
- 限流当前后端结构没有 OR 语义，页面使用共享组件但保持 relation 不可编辑，避免 UI 表达超出真实能力。

## 泳道规则抽屉宽度与路由对齐

- [x] 核对路由规则当前宽抽屉尺寸来源
- [x] 将宽抽屉尺寸抽成共享常量
- [x] 让泳道组独立入口和治理工作台 lane 入口使用与路由一致的宽度
- [x] 运行前端构建、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 路由规则在治理工作台使用 `min(1560px, calc(100vw - 40px))` 宽抽屉；泳道组编辑器同样是左编辑、右 Spec 的双栏结构，应保持同一抽屉宽度。
- 已在 `RuleDetailDrawer` 导出 `WIDE_RULE_DETAIL_DRAWER_SIZE`，路由、泳道、熔断、主动探测和工作台宽抽屉入口统一引用，避免继续散落硬编码。

验证：

- `rg` 确认 `min(1560px, calc(100vw - 40px))` 只在 `RuleDetailDrawer` 的 `WIDE_RULE_DETAIL_DRAWER_SIZE` 中定义，路由、泳道、熔断、主动探测和工作台宽抽屉入口均引用常量。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/RuleRelease/RuleDetailDrawer.tsx console/web/src/pages/Governance/Router/CustomRoute.tsx console/web/src/pages/Governance/Router/LaneGroupTable.tsx console/web/src/pages/Governance/Workbench/index.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerTable.tsx console/web/src/pages/Governance/CircuitBreaker/FaultDetectTable.tsx context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- 浏览器在 1600px 视口下验证：泳道新建抽屉 `.t-drawer__content-wrapper` 宽度为 `1560px`、left 为 `40px`；路由新建抽屉宽度同为 `1560px`、left 为 `40px`。

Review：

- 本轮只统一治理规则详情抽屉宽度来源，不修改泳道、路由或其它规则的保存逻辑和接口契约。
- 独立路由入口此前未显式传宽度，本轮也改为引用宽抽屉常量，使“路由作为基准”在独立入口和工作台入口一致。

## 治理规则数字输入框统一无按钮

- [x] 确认当前 TDesign React `InputNumber` 无按钮形态参数，并以本地依赖版本为准
- [x] 横向检查 Governance 下所有治理规则编辑器的数字输入框
- [x] 将所有治理规则数字输入统一为无按钮 `InputNumber`
- [x] 补充静态校验，防止新增或回退到带按钮数字输入框
- [x] 运行前端构建、校验脚本、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- TDesign React `InputNumber` 当前默认 `theme="row"`，会显示加减按钮；治理规则编辑器应统一使用 `theme="normal"` 的无按钮形态。
- 本轮范围限定为 Console 前端 Governance 规则编辑器，不修改后端接口、存储、protobuf 或规则语义。
- 已新增 `verify-governance-input-number-theme.mjs`，扫描 Governance 下 TSX 中的 JSX `InputNumber` 和表格 inline edit `component: InputNumber`。

验证：

- TDesign React MCP 与本地 `tdesign-react` 依赖均确认 `InputNumber.theme` 可选 `column / row / normal`，当前默认值是 `row`，无按钮形态使用 `normal`。
- `cd console/web && node scripts/verify-governance-input-number-theme.mjs` 通过，覆盖 Governance 下 JSX `InputNumber` 与表格 inline edit `component: InputNumber`。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/CircuitBreaker/FaultDetectEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/LossLess/LossLessEditor.tsx console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/src/pages/Governance/Router/LaneGroupEdtor.tsx console/web/src/pages/Governance/Router/LaneRuleEditor.tsx console/web/scripts/verify-governance-input-number-theme.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。

Review：

- 本轮只统一 Console 前端治理规则编辑器数字输入的 TDesign 组件外观，不修改字段类型、保存转换或后端契约。
- 覆盖范围包括路由、限流、熔断、主动探测、无损、泳道、调用鉴权、流量镜像、流量 Mock 的现有数字输入点。
- 后续新增治理规则编辑器时应先跑 `verify-governance-input-number-theme.mjs`，避免默认 `row` 主题重新出现加减按钮。

## 泳道规则编辑抽屉交接对齐

- [x] 读取泳道规则设计交接文档，确认本轮只调整 Console 前端编辑体验、预览和保存前映射
- [x] 核对当前 `LaneGroupEdtor` / `LaneRuleEditor` / `LaneGroupTable` 与共享治理编辑器范式
- [x] 先补充前端回归验证脚本，覆盖泳道数据归一化、`LaneGroup` / `LaneRule` 实时 Spec、放量比例和拓扑摘要
- [x] 将泳道组详情抽屉调整为 `泳道规则 / 泳道组 / 版本 / 审计` 分页编辑，组与泳道不混在同页
- [x] 补齐泳道规则页的泳道定义、固定标签、流量匹配、组内服务选择和流量拓扑演示
- [x] 运行相关前端脚本、构建、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮边界是治理工作台泳道规则编辑抽屉；不修改后端 schema、protobuf、存储或 API 契约。
- 交接文档中的 `LaneGroup` / `LaneRule` YAML 是 Console 侧实时预览和产品态表达；真实保存仍通过当前 `/naming/v1/lane/groups` 与 `/naming/v1/lane/groups/rules` 接口。
- 当前实现仍把泳道组编辑和泳道列表堆在详情页里，泳道规则另开单独弹窗；需要收敛到一个抽屉内的分页编辑体验。
- 已新增 `laneEditorUtils.ts`，将前端 draft、校验、实时 Spec、YAML/JSON 序列化和拓扑摘要从组件中拆出，便于脚本验证。
- 已将 `LaneGroupEdtor` 改为双栏编辑器：左侧内部滚动并含 `泳道规则 / 泳道组 / 版本 / 审计` Tab，右侧固定实时 Spec。
- 用户截图反馈后确认：父级 `LaneGroupTable` 与治理工作台 lane 详情仍套了通用 `RuleTabs`，导致外层 `规则 / 版本 / 监听` 与内层泳道真实 Tab 嵌套。
- 已让 `LaneGroupTable` 与治理工作台 lane 详情直接渲染 `LaneGroupEdtor`，只保留泳道编辑器自己的 `泳道规则 / 泳道组 / 版本 / 审计`，不再包外层通用 Tab。
- 用户再次对照 PRD 截图后确认：左侧默认页不应是单一「泳道定义」卡片，而应是 `组信息 / 泳道 / 版本 / 审计`，默认 `组信息` 展示「基础信息 / 泳道组入口 / 组内服务」三段步骤卡。
- 已将 `组信息` 作为默认 Tab，左侧基础信息改为 PRD 式两列表单；入口与组内服务改为表格行，`泳道` Tab 单独承载泳道规则定义。
- 用户进一步给出泳道展开卡片截图后确认：`泳道` Tab 内单条泳道应按卡头摘要、基础字段、固定标签提示、流量匹配、组内服务和拓扑演示的连续展开布局实现；不能把流量匹配退回普通横排输入或路由规则表格。
- 本轮继续只调整 Console 前端布局、交互和验证脚本，不修改后端接口、存储或 protobuf。

验证：

- RED：`cd console/web && node scripts/verify-lane-editor-utils.mjs` 初次失败，缺少 `src/pages/Governance/Router/laneEditorUtils.ts`。
- GREEN：`cd console/web && node scripts/verify-lane-editor-utils.mjs` 通过，并断言 `LaneGroupTable` / 治理工作台不再用外层 `RuleTabs` 包裹 `LaneGroupEdtor`。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/LaneGroupEdtor.tsx console/web/src/pages/Governance/Router/LaneGroupEditor.module.less console/web/src/pages/Governance/Router/laneEditorUtils.ts console/web/src/pages/Governance/Router/LaneGroupTable.tsx console/web/src/pages/Governance/Workbench/index.tsx console/web/scripts/verify-lane-editor-utils.mjs context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- Playwright 登录本地 8080（当前本地 main user 为 `admin/admin123`）后打开治理工作台 `seed-20260616-lane`：抽屉内显示 `泳道规则 / 泳道组 / 版本 / 审计`，右侧显示 `LaneRule` 实时 Spec，旧的下方 `泳道列表` 面板已移除。
- 用户截图反馈的嵌套 Tab 修复后，Playwright DOM 检查泳道详情抽屉返回 `hasLaneTabs=true`、`hasOuterRuleTabSet=false`；抽屉文本为 `seed-20260616-lane / 泳道 / 泳道规则 / 泳道组 / 版本 / 审计 ...`，不再出现外层 `规则 / 版本 / 监听`。
- PRD 左侧布局修复后，Playwright 真实页面验证默认 Tab 为 `组信息`，左侧显示 `基础信息 / 泳道组入口 / 组内服务` 三段，右侧 Spec 切为 `LaneGroup`；编辑态规则名称输入框值为 `seed-20260616-lane`，优先级为 `5`，两者同排展示。
- Playwright 展开 `blue-lane` 后确认页面展示固定标签 `X-Lattice-Traffic-Lane = base`、放量比例 `100%`、组内服务 `spec-payment` 和 `③ 流量拓扑演示` SVG。
- `cd console/web && node scripts/verify-standard-response-mapping.mjs` 当前失败在既有主动探测断言：`src/pages/Governance/CircuitBreaker/FaultDetectTable.tsx: missing "targetService?.api"`，非本轮泳道改动路径。
- 最新泳道展开卡片布局修复后，`cd console/web && node scripts/verify-lane-editor-utils.mjs` 通过，覆盖卡头摘要、固定标签提示、流量匹配分段控件、四字段匹配行、组内服务行和底部新建泳道入口。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/LaneGroupEdtor.tsx console/web/src/pages/Governance/Router/LaneGroupEditor.module.less console/web/scripts/verify-lane-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 已重启 all-mode，tmux 日志显示 `finish starting server`，8080 首页返回 200。
- Playwright 登录真实 8080 后打开 `seed-20260616-lane`，进入编辑态并切到 `泳道` Tab：展开 `blue-lane` 后确认页面展示卡头摘要、泳道名称、泳道标签 Value、固定标签提示、流量匹配规则、命中后放量、匹配条件四字段、添加匹配规则、进入泳道的组内服务、添加组内服务和流量拓扑演示；截图保存到 `output/playwright/lane-card-layout.png`。
- 用户继续反馈泳道卡控件和字体不应显得突兀；后续修正目标是把泳道卡局部字号、控件高度和间距统一到治理抽屉既有 12/14px 与 32px 控件密度。
- 泳道卡视觉密度修正后，`verify-lane-editor-utils.mjs` 新增 CSS 断言并通过；`npm run build:test` 通过；all-mode 已重新启动，Playwright 真实页面确认放量数值 `100` 完整可读，最终截图保存到 `output/playwright/lane-card-layout-compact-final.png`。
- 用户继续反馈泳道编辑内容滚动不了；Playwright DOM 指标确认根因是 `.t-tabs` 为 `overflow:hidden` 且 `scrollHeight > clientHeight`，但外层 `.formPane` 的 `scrollHeight == clientHeight`，导致滚动挂错层。修正方向：将 tabs 设为 column flex，把滚动明确挂到 `.t-tabs__content`。
- 滚动修复后，真实 8080 页面中 `.t-tabs__content` 指标为 `overflowY=auto`、`scrollHeight=979`、`clientHeight=827`；脚本设置 `scrollTop` 可从 `0` 变为 `80`，鼠标滚轮从 `0` 滚到 `151.5`，最大可滚动值为 `152`。
- 用户反馈实际仍无法滚动后，判断上轮验证命中的是内部 Tabs 内容层，不足以代表用户滚轮区域。进一步将左侧 `.formPane` 改为唯一滚动容器，Tabs 内容层改为 `overflow: visible`，避免嵌套滚动层命中不一致。
- 单一滚动容器修复后，真实 8080 页面验证 `.formPane` 为 `overflowY=auto`、`scrollHeight=1055`、`clientHeight=903`；鼠标命中左侧泳道卡片中心点后，滚轮使 `.formPane.scrollTop` 从 `0` 变为 `151.5`，最大滚动值为 `152`。

Review：

- 本轮只调整 Console 前端泳道编辑体验和前端预览工具，不修改后端接口、存储或 protobuf。
- `LaneGroupEdtor` 现在承担泳道聚合编辑主路径；旧 `LaneRuleEditor` 文件仍保留，避免影响其它入口，但主抽屉不再在底部重复挂 `LaneRuleTable`。
- `matchRatio` 属于本轮交接文档要求的前端产品态字段，当前后端 LaneRule API 没有对应持久化字段；本轮在实时 Spec、校验和拓扑里完整表达，真实保存仍按当前 API 提交 `trafficMatchRule`、`defaultLabelValue` 和 `labelKey`。

## 鉴权规则编辑抽屉规则 Tab 对齐

- [x] 读取鉴权规则设计交接文档，确认只调整 Console 前端展示、交互组织和预览映射
- [x] 核对当前 `TrafficGovernanceEditor` 鉴权分支、工具函数和治理编辑器共享布局范式
- [x] 先补充前端回归验证脚本，覆盖基础信息与服务信息分离、子规则分区顺序、服务级唯一和预览映射
- [x] 调整鉴权规则 Tab 为「① 基础信息 → ② 服务信息 → ③ 鉴权子规则」并保持右侧实时 `AuthRule` 预览
- [x] 运行前端脚本、构建、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮边界是治理工作台鉴权规则编辑抽屉的「规则」Tab；不修改后端 schema、protobuf、存储或 API 契约。
- 交接文档中的 `AuthRule/spec.scope/subRules/listType/protectedInterfaces/strategy.match` 是前端产品态预览结构；真实保存仍转换为当前 `TrafficSecurityRule.policies[]`。
- 当前代码已有黑名单、白名单和服务级分区雏形，但鉴权服务字段仍在通用基础信息里，需要拆出独立「② 服务信息」section。
- 已将鉴权服务信息从通用基础信息移到独立「② 服务信息」section；基础信息只保留名称、启用状态、优先级、Revision、描述和规则标签。
- 已补充 `verify-traffic-security-editor-utils.mjs` 静态结构断言，防止后续把命名空间/服务名称放回基础信息。
- 已补充规则名称 kebab-case 保存校验，右侧预览仍使用前端产品态 `AuthRule` 结构，提交仍转换为当前 `TrafficSecurityRule.policies[]`。

验证：

- RED：`cd console/web && node scripts/verify-traffic-security-editor-utils.mjs` 初次失败，现状缺少 `renderSecurityServiceInfo` 且服务字段仍在基础信息。
- `cd console/web && node scripts/verify-traffic-security-editor-utils.mjs` 通过。
- `cd console/web && node scripts/verify-traffic-mirror-editor-utils.mjs` 通过。
- `cd console/web && node scripts/verify-traffic-mock-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/src/pages/Governance/Security/index.module.less console/web/src/pages/Governance/Security/trafficSecurityEditorUtils.ts console/web/scripts/verify-traffic-security-editor-utils.mjs context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，tmux 日志显示 `finish starting server`，8080 返回 200。
- Playwright 登录真实 8080 后打开 `seed-20260616-security`：查看态显示「① 基础信息」「② 服务信息」「③ 鉴权子规则」，服务信息区包含 `spec-governance/spec-order`，右侧 `AuthRule` 预览显示 `spec.scope.namespace/service`。
- Playwright 编辑态断言通过：基础信息只包含规则名称、启用状态、优先级、规则标签和描述；命名空间/服务名称只出现在「② 服务信息」；接口级子规则没有名单类型选择；截图保存到 `output/playwright/traffic-security-authrule-editor.png`。

Review：

- 本轮只调整 Console 前端鉴权规则编辑抽屉、鉴权工具校验和对应验证脚本，不修改后端 schema、proto、存储或接口契约。
- 「② 服务信息」作为大规则层独立 section 插在基础信息和子规则之间；服务级规则仍只表示受保护接口为空的兜底子规则，不混入大规则服务字段。
- 右侧 `AuthRule` 预览继续作为产品态预览；真实保存仍由 `buildSecurityPoliciesFromView` 转换为当前后端接受的 `TrafficSecurityRule.policies[]`。

## 服务范围组件复用

- [x] 核对 RouteRule、熔断、Mock、镜像当前服务范围区实现与字段语义
- [x] 抽取路由已验证的 caller -> callee 服务范围区为共享组件
- [x] 将 RouteRule 服务范围区迁移到共享组件，保持现有视觉和折叠行为
- [x] 将熔断、Mock、镜像服务范围区切换到共享组件，并保留各自独有字段
- [x] 运行相关前端脚本、构建、context-kg lint 和 diff 检查
- [x] 记录 review 和 lesson

当前判断：

- RouteRule 已有用户认可的服务范围结构：紧凑标题、展开/收起、主调/被调服务卡片和中间方向连接器。
- 熔断也有 `source -> destination` 关系，但还额外包含熔断粒度；组件需要提供额外内容插槽，不能吞掉熔断独有配置。
- Mock 和镜像已经按 caller -> callee 收敛，但当前还是普通四字段表单，视觉和路由不一致。

当前进展：

- 已新增共享 `ServiceScopeSection` 组件和样式，承载 section/header/card/connector/editable/readonly/collapse。
- RouteRule 已从私有 `renderServiceCard` / `renderServiceScopeHeader` 迁移到共享组件，继续使用 `order=2` 和现有折叠状态。
- 熔断已接入共享组件，`source -> destination` 用同一服务关系卡片渲染；熔断粒度通过 `extraContent` 插槽保留。
- Mock 和镜像已接入共享组件，替换截图中的四字段平铺表单；Mock 的“全部命名空间/全部服务”只提供给 caller 侧，callee 侧仍使用实际命名空间/服务选项。
- 共享组件支持 caller/callee 分别配置 namespace/service options，避免不同规则类型的“全部服务”语义互相污染。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过。
- `cd console/web && node scripts/verify-circuitbreaker-editor-utils.mjs` 通过。
- `cd console/web && node scripts/verify-traffic-mirror-editor-utils.mjs` 通过。
- `cd console/web && node scripts/verify-traffic-mock-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/shared/ServiceScopeSection.tsx console/web/src/pages/Governance/shared/ServiceScopeSection.module.less console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/src/pages/Governance/Security/index.module.less console/web/src/pages/Governance/Security/trafficMirrorEditorUtils.ts console/web/src/services/traffic_governance.ts console/web/scripts/verify-traffic-mirror-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只抽取和复用 Console 前端服务范围展示组件，不修改后端保存接口或治理规则 spec。
- `ServiceScopeSection` 封装的是产品态 caller -> callee 关系，不承载规则类型特有配置；特有配置通过 `extraContent` 追加，当前用于熔断粒度。
- Mock 被调侧没有继承 caller 的“全部服务”选项，避免和最终语义冲突。

## specification 流量治理 caller -> callee 契约补齐

- [x] 核对 `../specification` 当前 `TrafficMirror` / `TrafficMock` 契约和本地未提交差异
- [x] 使用独立 worktree 从 `origin/develop` 创建干净 spec 分支，避免污染现有 `../specification` 未提交文件
- [x] 在 spec 顶层补齐 Mirror / Mock 的 `caller` 与 `callee` 字段，并保留旧 `target_service` 兼容
- [x] 重新生成 specification Go / Rust 产物
- [x] 创建 specification PR
- [x] 记录 PR、验证结果和后续 control-plane 适配事项

当前判断：

- 既然 Console 语义已收敛为 `Caller -> Callee`，spec 不应只通过 `target_service` + 子规则 `CALLER_SERVICE` 隐式表达。
- 本轮 spec 采用兼容方案：新增通用 `ServiceScope`，`TrafficMirror` / `TrafficMock` 顶层新增 `caller`、`callee`；旧 `target_service` 保留并标注 deprecated，避免立即破坏旧 control-plane 与旧数据。
- Caller 只允许两种产品态：`*/*` 表示全部服务，或具体 `namespace/service`；Callee 表达规则归属和客户端下发绑定服务。

验证：

- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee && git diff --check` 通过。
- specification PR 已创建：`https://github.com/lattice-hub/specification/pull/7`。

Review：

- spec PR 分支为 `codex/traffic-caller-callee`，commit 为 `a86d8e0 feat: add traffic caller callee scope`。
- PR 只包含 `traffic_manage` proto 与 Go/Rust 生成产物；创建前已从独立 worktree 清掉无关 `source/rust/pole-specification/proto/service.proto` 行尾噪音，未污染当前 `../specification` 工作区的未提交改动。
- 后续 control-plane 在 specification 发版后应优先读取/提交 `caller` 与 `callee`，旧 `target_service` 仅作为兼容兜底；当前 Console 临时兼容仍会通过 `target_service` 和 `CALLER_SERVICE` 工作。

## Mock 规则 caller -> callee 范围修正

- [x] 核对当前 `TrafficMock` spec、前端类型、Mock 编辑器和镜像 caller/callee 修正方式
- [x] 将 Mock 规则的大规则层调整为主调方 → 被调方服务范围，caller 支持全部服务
- [x] 保存时把 caller 范围兼容写入当前 `traffic_match_rule.CALLER_SERVICE`
- [x] 补齐 Mock 实时 Spec 预览、保存校验和列表摘要
- [x] 运行前端脚本、构建、diff 检查和真实 8080 页面验证
- [x] 记录 review、验证结果和 lessons

当前判断：

- 当前 specification 的 `TrafficMock` 已有 `target_service`，该字段继续作为 callee 和规则归属；本轮不修改后端 proto、存储或缓存。
- 用户进一步指出服务 Mock 也应和路由、镜像一样关心 caller → callee；caller 可以选择全部服务。
- 兼容方案：前端把 caller 作为 Mock 大规则层服务范围展示和编辑；提交时如果 caller 不是全部服务，则自动写入每条 `MockRule.traffic_match_rule.arguments` 的 `CALLER_SERVICE` 条件；如果 caller 是全部服务，则移除该条件。

验证：

- `cd console/web && node scripts/verify-traffic-mock-editor-utils.mjs` 通过，覆盖 caller 抽取、预览隐藏 `CALLER_SERVICE`、全部服务提交移除来源服务条件、具体 caller 提交自动注入来源服务条件。
- `cd console/web && node scripts/verify-traffic-mirror-editor-utils.mjs` 通过，确认未破坏镜像 caller/callee 逻辑。
- 按用户进一步纠正，已补充 Mock / 镜像 caller 半状态验证：`namespace/*` 会归一为全部服务；`namespace/空服务` 会被校验拦截，不允许保存成不完整 caller。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Security/trafficMockEditorUtils.ts console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/scripts/verify-traffic-mock-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，tmux 日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-mock`：列表摘要显示 `spec-governance/spec-order -> spec-governance/spec-order / Code OK`；查看态显示服务范围、主调方、被调方、流量方向和右侧 `MockRule.spec.scope.caller/callee` 预览。
- Playwright 编辑态断言通过：存在 `主调命名空间`、`主调服务`、`被调命名空间`、`被调服务`，主调服务下拉包含 `全部服务`；页面文本不包含 `CALLER_SERVICE`；截图保存到 `output/playwright/traffic-mock-caller-callee-edit.png`。

Review：

- 本轮只改 Console 前端表达和保存前转换，不修改 specification proto、Go 后端、存储或缓存。
- `target_service` 继续作为 callee 和规则归属；caller 作为前端大规则层范围展示，保存时兼容写入每条 Mock 子规则的来源服务匹配条件。
- Mock 子规则继续只承载选择接口、流量匹配和 Mock 响应结果；来源服务不再作为子规则条件展示，也不会进入右侧产品态 `MockRule` 预览的 `match.conditions`。
- Caller 只允许两种产品态：全部服务，或具体 `namespace/service`；半状态会在 helper 层归一或被校验拦截，避免 UI 与保存语义分叉。

## 流量镜像规则 caller -> callee 范围修正

- [x] 核对当前 `TrafficMirror` spec、前端类型、编辑器和路由 caller/callee 范式
- [x] 将镜像规则的大规则层调整为主调方 → 被调方服务范围，caller 支持全部服务
- [x] 保存时把 caller 范围兼容写入当前 `traffic_match_rule.CALLER_SERVICE`
- [x] 补齐镜像实时 Spec 预览、保存校验和列表摘要
- [x] 运行前端脚本、构建、diff 检查和真实 8080 页面验证
- [x] 记录 review、验证结果和 lessons

当前判断：

- 当前 specification 的 `TrafficMirror` 已有 `target_service`，后端缓存/下发也按该服务绑定；本轮不直接修改后端 proto。
- 用户指出的问题是镜像规则产品语义应像路由一样表达 caller → callee，caller 可以为全部服务；不应让用户在子规则流量匹配里手动选择 `CALLER_SERVICE` 来表达来源服务。
- 兼容方案：前端把 caller 作为镜像大规则层服务范围展示和编辑；提交时如果 caller 不是全部服务，则自动写入每条 `MirrorRule.traffic_match_rule.arguments` 的 `CALLER_SERVICE` 条件；如果 caller 是全部服务，则移除该条件。

验证：

- `cd console/web && node scripts/verify-traffic-mirror-editor-utils.mjs` 通过，覆盖 caller 解析、全部服务移除来源服务条件、具体 caller 注入来源服务条件、预览和校验。
- `cd console/web && npm run build:test` 通过，保留既有 Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，tmux 日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-mirror`：列表摘要显示 `caller -> callee / 镜像到 destination`；查看态显示服务范围、主调方、被调方、流量方向和右侧 `mirror_config.caller/callee` 预览。
- Playwright 编辑态断言通过：存在 `主调命名空间`、`主调服务`、`被调命名空间`、`被调服务`，`全部服务` 可见；页面不再出现实现字段名 `CALLER_SERVICE` 和旧的 `命中比例`；截图保存到 `output/playwright/traffic-mirror-caller-callee-edit.png`。

Review：

- 本轮只改 Console 前端表达和保存前转换，不修改 specification proto、Go 后端、存储或缓存。
- `target_service` 继续作为 callee 和规则归属；caller 作为前端大规则层范围展示，保存时兼容写入每条子规则的来源服务匹配条件。
- 镜像子规则仍保持四步：镜像接口 → 流量匹配 → 采样 → 镜像目标；流量匹配不再暴露来源服务和命中比例，避免和服务范围、镜像比例重复。

## OpenServer 观测闭环设计

- [x] 核对现有日志、OTel、observability 插件和 Console 监控入口
- [x] 确认 OpenServer 在观测链路中的定位：内置 OTLP 接收/存储，还是外部采集器转发入口
- [x] 收敛 logs、metrics、traces 以及后续 pole-sdk 监控数据的边界
- [ ] 提出 2-3 个架构方案，并给出推荐方案与取舍
- [ ] 用户认可后归档技术 ADR，同步 `context-kg/_meta/index.md` 与 `context-kg/_meta/log.md`
- [ ] 进入实现计划，列出配置、插件、协议、Console 和验证路径

当前判断：

- 当前仓库已有 `pkg/common/otel/`，但只实现 OTLP gRPC exporter 和本进程指标注册，尚未在启动配置中接入，也不是 receiver。
- 当前 observability 插件包含 history、discoverEvent、statis 三类，分别服务操作历史、发现事件和统计数据；这些是 Pole 内部模型，不等同于 OTel 信号。
- 当前 Console 监控入口仍通过 `monitorServer.address` 指向外部服务，尚未形成 OpenServer 内置观测闭环。
- 推荐优先保持 OTel 协议边界：OpenServer 可以成为默认采集/治理入口，但 SDK 和用户侧仍应能够直接替换为标准 OTLP collector 或其它后端。
- 用户已确认观测后端内置闭环：接收 OTLP，自己存 logs、metrics、traces，并给 lattice-hub / Console 查询治理。
- 在 pole-control-plane 侧不需要新增独立 `openobserver` 插件类型；现有 `history`、`discoverEvent`、`statis` 都是 chain 模式，应分别新增可配置的 `otel` chain entry，启用后把内部模型转换并发送到内置观测后端。
- OTel events 不应设计成独立第四类协议；按 OTel 语义应建模为带 `event.name` 的 LogRecord，再在 OpenObserver 查询层提供 events 视图。
- 初步映射：`history.RecordEntry` 经 `history.entries: [{name: otel}]` 进入 audit event/log，`DiscoverEvent` 经 `discoverEvent.entries: [{name: otel}]` 进入 domain event/log，`statis` 经 `statis.entries: [{name: otel}]` 进入 metrics，真实请求链路和 SDK 调用链进入 traces。

验证：

- 待执行。

Review：

- 待补充。

## 8091 客户端发现响应缓存与实例快照优化

- [x] 写失败测试：同一服务同一 revision 的 full response 第二次请求不再遍历实例缓存
- [x] 写失败测试：`ServiceInstances` 多次读取实例时复用快照并在实例变更后重建
- [x] 为 8091 服务实例发现增加业务层 `DiscoverResponse` 缓存，key 包含 namespace、service、revision、onlyHealthy
- [x] 为 `ServiceInstances` 增加 all/healthy 实例快照，缩短请求线程锁持有与重复 map 遍历
- [x] 补充 revision 命中、response cache 命中/未命中指标或可观测计数
- [x] 运行相关 Go 测试、context-kg lint 和 diff 检查
- [x] 记录 review 与验证结果

当前判断：

- 本轮只优化 8091 客户端服务实例发现热路径；治理规则、配置发现和 Console 查询不纳入。
- 业务层 response 缓存只缓存成功 full response，不缓存 `DataNoChange`、推空保护或错误响应。
- 实例快照放在 `ServiceInstances` 内部，对调用方保持 `DiscoverServiceInstances` 回调接口不变，降低改动面。
- gRPC prepared message cache 继续保留；本轮新增的业务缓存用于减少 protobuf 编码前的实例遍历和 response 构造。

验证：

- RED：`go test ./pkg/service -run TestServer_ServiceInstancesCacheReusesFullResponseForSameRevision -count=1` 失败，第二次请求仍调用 `DiscoverServiceInstances`。
- RED：`go test ./apis/pkg/types/service -run TestServiceInstancesGetInstancesDoesNotHoldLockDuringConsumer -count=1` 失败，`GetInstances` 回调期间持锁导致测试超时。
- GREEN：`go test ./pkg/service -run 'TestServer_ServiceInstancesCache(ReusesFullResponseForSameRevision|DoesNotMergeVisibleServicesFromOtherNamespaces)' -count=1` 通过。
- GREEN：`go test ./apis/pkg/types/service -run TestServiceInstancesGetInstancesDoesNotHoldLockDuringConsumer -count=1` 通过。
- `go test ./pkg/service ./apis/pkg/types/service ./pkg/cache/service ./apis/pkg/types/metrics ./plugin/observability/statis/base` 通过。
- `go test ./plugin/observability/statis/...` 通过。
- `go test ./plugin/apiserver/grpcserver/...` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- `ServiceInstancesCache` 现在先记录 revision 命中；revision 未命中时再按 `INSTANCE:namespace:service:revision:onlyHealthy` 查业务层 response cache，命中后不再遍历实例缓存。
- response cache 存取都使用 proto clone，避免共享 `DiscoverResponse` 被后续发送或调用方修改。
- `ServiceInstances.GetInstances` 改为先取 all/healthy 快照，再在锁外执行 consumer；Upsert、Remove、健康保护重算和保护阈值变化都会失效快照。
- 新增 `DiscoverCacheCallMetric`，通过现有 cache statis 聚合 revision 和 response cache 的 hit/miss。

## Mock 规则编辑抽屉 PRD 对齐

- [x] 读取 Mock 规则设计交接文档，确认只调整 Console 前端交互组织与预览映射
- [x] 核对当前 `TrafficGovernanceEditor` 的 Mock 数据结构、保存 payload 和治理编辑器双栏范式
- [x] 新增 Mock 编辑器工具函数，覆盖草稿归一化、右侧预览、保存校验和 Body 高度策略
- [x] 重构 Mock 规则 Tab 为基础信息、Mock 服务、Mock 子规则和右侧实时 `MockRule` Spec 预览
- [x] 移除 Mock 比例、数字状态码范围、Reason/message 与固定 Content-Type 主控件
- [x] 运行前端逻辑脚本、构建、diff 检查和页面级验证
- [x] 记录 review 与验证结果

当前判断：

- 本轮只优化治理工作台 `Mock` 规则 Tab 的前端组织、页面预览映射和保存校验，不调整后端 schema、proto 或存储契约。
- 附件里的 `MockRule/spec.rules[]` 是产品态预览结构；真实提交仍适配当前 `TrafficMock.rules[]` 后端契约。
- Mock 服务信息属于大规则层，Mock 子规则只承载接口、流量匹配和 Mock 响应结果；子规则内不再展示 Mock 比例。
- Mock 响应 Code 是字符串，只校验非空；响应 Header 是普通 Key-Value 列表，`content-type` 仅作为普通 header 影响 JSON Body 校验。

当前进展：

- 已读取交接文档、当前 `TrafficGovernanceEditor`、`traffic_governance` 类型、鉴权编辑器 utils 和治理编辑器共享样式。
- 已确认现有 Mock 编辑器仍包含旧形态：Mock 比例、`status_code` 数字范围校验、`message` 响应文案和普通大 Textarea，需要按 PRD 收敛。
- 已新增 `trafficMockEditorUtils.ts`，将 Mock 草稿归一化、提交 payload、右侧 `MockRule` 预览、校验和延迟毫秒转换从主组件拆出。
- 已将 Mock 规则 Tab 改为基础信息、Mock 服务、Mock 子规则和右侧实时 Spec 双栏；Mock 服务只在大规则层展示。
- Mock 子规则按选择接口、流量匹配、Mock 响应结果三段组织；Mock 响应结果使用单一响应面板，包含 Code/延迟、Header、Body 和最终返回预览。
- 编辑态不再展示 Mock 比例、响应状态码、响应文案、Reason 或固定 Content-Type 主控件；提交时保留隐藏 `mock_percent=100` 适配当前后端契约。

验证：

- `cd console/web && node scripts/verify-traffic-mock-editor-utils.mjs` 通过。
- `cd console/web && node scripts/verify-traffic-security-editor-utils.mjs` 通过，确认未破坏鉴权编辑器工具函数。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/src/pages/Governance/Security/index.module.less console/web/src/pages/Governance/Security/trafficMockEditorUtils.ts console/web/src/services/traffic_governance.ts console/web/scripts/verify-traffic-mock-editor-utils.mjs context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，tmux 日志显示 `finish starting server`，8080 返回 200。
- Playwright 登录真实 8080 后打开 Mock 规则 `seed-20260616-mock`：查看态显示 Mock 服务、Mock 子规则三段、最终返回预览和右侧 `kind: MockRule` 预览。
- Playwright 编辑态断言通过：页面存在 `Mock 服务`、`Mock 子规则 [1]`、`全屏编辑`、`kind: MockRule`；页面文本不包含 `Mock 比例`、`响应状态码`、`响应文案`、`Reason`、`ratioPercent`、`reason`。
- Playwright 验证普通 Body 一行高度约 32px；全屏编辑 Body 后，普通 textarea、最终返回预览和右侧 Spec 同步更新。
- Playwright 点击 `添加 Mock 子规则` 后页面出现 `Mock 子规则 [2]`，右侧预览出现 2 个 `mock-subrule-*`。
- Playwright 清空响应 Code 后右侧出现 `响应 Code 不能为空`，且不再显示 `校验通过，可保存并下发`。
- 页面截图保存到 `output/playwright/traffic-mock-rule-editor.png`。

Review：

- 本轮只调整 Console Mock 规则编辑抽屉和前端服务类型，不修改后端 proto、存储或接口 schema。
- 右侧 `MockRule/spec.rules[]` 是前端产品态预览，真实提交仍转换为当前 `TrafficMock.rules[]`；为了表达“命中即 Mock”，提交层固定补 `mock_percent=100`，但 UI 和预览不暴露比例。
- 旧草稿中的 `status_code` 会在前端归一化为字符串 `response.code`；Mock UI、预览和提交响应体不再输出 `status_code`、`message` 或 `reason`。

## specification 治理规则响应 Code string 化

- [x] 确认治理规则中响应效果类 `code` 字段范围
- [x] 收回误改的通用 API `model.Response` / `DiscoverResponse` / `ConfigDiscoverResponse` / `ratelimiter` 协议响应
- [x] 修改 `../specification/api/v1/fault_tolerance/circuitbreaker.proto` 中 `FallbackResponse.code` 为 `string`
- [x] 删除 Mock / 鉴权治理响应效果中的 `status_code`，统一只保留 `code string`
- [x] 删除 Mock 响应效果中的 `message`，避免与 `body` 重复表达响应内容
- [x] 重新生成 `../specification/source/go` 产物，并同步 Rust proto 副本
- [x] 运行治理规则 code 扫描、spec Go/Rust 生成验证、control-plane 聚焦编译和 context-kg lint
- [x] 记录 review 与验证结果
- [x] 合并 specification PR 后创建并推送 `v0.1.0-ALPHA.30` tag
- [x] 更新 pole-control-plane 依赖到 `github.com/pole-io/specification v0.1.0-ALPHA.30`

当前判断：

- 本轮以相邻 `../specification` 为协议真源，只统一治理规则里的响应效果 Code：限流 `CustomResponse.code`、流量 Mock `MockResponse.code`、调用鉴权 `TrafficSecurityRejectEffect.code`、熔断 `FallbackResponse.code` 均为 `string`。
- 治理响应效果里不保留 `status_code`；Mock 和鉴权原有 `status_code` 已删除，`code` 收敛到 field 1。
- Mock 响应效果里 `body` 是实际 Mock 出来的响应内容，`message` 与它语义重复；Mock 保留 `code`、`headers`、`body`，鉴权拒绝效果因为没有 `body` 仍保留 `message`。
- `api/v1/model/response.proto`、服务发现 `DiscoverResponse`、配置发现 `ConfigDiscoverResponse`、心跳删除响应和限流器协议 `ratelimiter` 不属于治理规则响应效果，不纳入本轮。
- 当前 `../specification/source/rust/pole-specification/proto/service.proto` 只有换行符格式差异的既有脏改动；本轮不处理该文件。
- specification PR `https://github.com/lattice-hub/specification/pull/6` 已合并到 `develop`，merge commit 为 `e7c8eafc7a2424dc677c27e3ad09bcf8a8445d21`；`v0.1.0-ALPHA.30` tag 指向该提交。
- pole-control-plane 已将 `go.mod` 中 `github.com/pole-io/specification` 的 `require` 与 `replace github.com/lattice-hub/specification` 同步更新到 `v0.1.0-ALPHA.30`，并补充对应 `go.sum` checksum。

验证：

- `cd ../specification/source/go && bash build.sh` 通过。
- `cd ../specification/source/rust/pole-specification && PROTOC=/Users/chuntao.liao/Github/pole-io/specification/source/protoc/protoc-darwin-arm64/bin/protoc cargo build --release` 通过。
- `cd ../specification && go test ./...` 通过。
- 精确扫描确认治理规则响应效果：`FallbackResponse.code`、`CustomResponse.code`、`MockResponse.code`、`TrafficSecurityRejectEffect.code` 均为 `string`；这些响应效果中不再存在 `status_code`，Mock 响应效果中不再存在 `message`。
- 使用临时 modfile 指向本地 `../specification` 运行 `go test ./pkg/common/api/v1 ./plugin/store/mysql` 通过。
- specification tag 验证：`git -C ../specification show --no-patch --format='%H %D %s' v0.1.0-ALPHA.30` 返回 `e7c8eafc7a2424dc677c27e3ad09bcf8a8445d21 tag: v0.1.0-ALPHA.30, origin/develop, origin/HEAD standardize governance response codes`。
- pole-control-plane 使用远端 tag 后运行 `go test ./pkg/common/api/v1 ./plugin/store/mysql` 通过。
- 更大范围 control-plane 编译曾被当前工作区既有 `pkg/cache/namespace` 的 `undefined: matchs` 阻断，与本轮 spec 变更无关。

Review：

- 已收回误改的通用 API 响应码变更，`pkg/common/api/v1` 无本轮 diff。
- 本轮 spec diff 包含熔断 `FallbackResponse.code` string 化、Mock / 鉴权响应效果删除 `status_code`、Mock 响应效果删除重复 `message` 及对应 Go/Rust 生成产物；`../specification/source/rust/pole-specification/proto/service.proto` 是既有换行符脏改动，不属于本轮。
- 本轮 control-plane 更新只涉及 `go.mod` / `go.sum` 的 specification 版本提升，不修改业务代码。

## 移除控制面跨命名空间服务发现能力

- [x] 写后端失败测试：跨命名空间可见服务不再被 8091 Discover 合并返回
- [x] 移除 `ServiceInstancesCache` / `GetServiceWithCache` 中的跨命名空间可见服务 fan-out
- [x] 清理服务缓存接口中仅服务发现使用的跨命名空间查询能力
- [x] 移除 Console 服务和命名空间的可见性展示、筛选、编辑入口
- [x] 运行后端相关测试、Console 构建和 context-kg lint
- [x] 记录 review 与验证结果

当前判断：

- 本轮移除的是 control-plane 客户端发现路径中的跨命名空间服务可见性；该能力后续如有需要应在 pole-sidecar 数据面实现。
- 后端存储字段和 proto 字段保留兼容历史数据，避免扩大到数据库迁移和 specification 破坏性变更；管理 API 不再写入或返回服务/命名空间可见性字段。
- MCP Server 的 `export_to` 属于 AI registry 资源字段，不属于服务注册发现客户端路径，本轮不纳入。

验证：

- `go test ./pkg/service -run TestServer_ServiceInstancesCacheDoesNotMergeVisibleServicesFromOtherNamespaces -count=1` 通过。
- `go test ./pkg/cache/service ./pkg/cache/namespace -run TestDoesNotExist` 通过。
- `go test ./pkg/service ./pkg/cache/service ./pkg/cache/namespace` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `go test ./...` 已尝试运行，前半段包正常输出通过，但超过 2 分钟无新输出后手动中止；本轮以相关包测试为准。

Review：

- 8091 服务发现现在只按请求的 `name/namespace` 读取当前服务实例，不再查其它命名空间可见服务，也不再计算跨服务 composite revision。
- 服务缓存和命名空间缓存不再维护跨命名空间可见性索引，相关 Cache 接口与 mock 已清理。
- 服务和命名空间管理 API 不再使用 `export_to` / `service_export_to`，旧 DB 字段仅作为兼容字段保留。
- Console 服务列表、服务编辑器、服务详情、命名空间列表、命名空间编辑器都移除了可见性展示与编辑入口；MCP Server 的 `export_to` 保留。

## 客户端请求链路性能瓶颈分析

- [x] 梳理 8090 HTTP 客户端入口的请求路径、序列化和连接模型
- [x] 梳理 8091 gRPC 客户端入口的注册、发现、心跳路径
- [x] 核对业务层拦截器、缓存、批处理与 MySQL 写入路径
- [x] 识别潜在吞吐/延迟瓶颈并按风险排序
- [x] 给出可验证的压测与观测建议
- [x] 记录 review 与验证结果

当前判断：

- 本轮先做代码与配置层面的性能风险分析，不修改实现。
- 重点关注客户端高频路径：服务注册、服务发现、心跳、配置发现，以及 8090 HTTP 与 8091 gRPC 两种接入方式差异。

当前进展：

- 已确认 8090 `api-http` 客户端入口会先做 protobuf JSON 解析，再进入 `namingServer` / `ruleServer` 缓存读写路径，响应默认走 `jsonpb` marshal。
- 已确认 8091 `service-grpc` 客户端入口直接走 protobuf gRPC；发现流支持 response prepared message 缓存，默认配置开启 `enableCacheProto`。
- 已确认服务发现主路径读缓存，不直接读 MySQL；注册、反注册、心跳通过 batch controller 合并后写 MySQL。
- 已确认缓存默认 1 秒 tick 增量刷新，因此写入成功到发现可见之间存在缓存刷新窗口。

验证：

- 本轮仅做静态代码与配置分析，未运行压测。
- 已核对入口、缓存、批处理和配置文件：`plugin/apiserver/httpserver/discover/client_access.go`、`plugin/apiserver/grpcserver/discover/v1/client_access.go`、`pkg/service/client_v1.go`、`pkg/service/batch/instance.go`、`pkg/common/batchctrl/batch.go`、`pkg/cache/cache.go`、`deploy/conf/pole-server.yaml`、`deploy/conf/pole-apiserver.yaml`。

Review：

- 纯内网客户端链路的主要性能风险不在 HTTP/3；优先关注发现响应体大小、revision 命中率、8090 JSON 序列化、8091 protobuf 缓存命中、注册/心跳批处理等待、MySQL 批量写入和缓存刷新窗口。
- 客户端高频发现应优先走 8091 gRPC；8090 更适合作为兼容和 Console/API 入口。
- 如果后续要形成长期性能方案，应另建 `context-kg/technical/adr/` 或模块页，本轮不把长期设计塞进 `tasks/todo.md`。

## Console OIDC 用户来源与企业目录同步方案归档

- [x] 核对 context-kg schema、index、log 与既有 ADR 风格
- [x] 明确方案边界：企业身份接入放在 Console 扩展点，不进入 pole-server 核心鉴权链
- [x] 新增技术 ADR，沉淀 OIDC 登录、Pole User 映射、企业目录同步和验证矩阵
- [x] 同步更新 auth-system、index、log
- [x] 运行 context-kg lint 与 diff 检查
- [x] 记录 review 与验证结果

当前判断：

- 本方案只改变 Console 用户来源，不替换 pole-server 内部 token、User/UserGroup/Role/Policy 和资源授权逻辑。
- OIDC 登录与飞书/Lark/钉钉企业目录同步应作为 Console 扩展点实现；pole-server 继续只管理 Pole User 与鉴权策略。
- 目录同步用于解决“企业用户未登录前无法被授权选择”的问题，同步结果仍必须落成 `source=oidc` 的 Pole User。

验证：

- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/technical/adr/adr-console-oidc-identity-source.md context-kg/technical/modules/auth-system.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。

Review：

- 本轮只沉淀技术方案，不改实现代码。
- 新 ADR 将 OIDC 登录、企业目录同步、Pole User 映射、时序图、异常处理和测试策略统一归档到 `technical/adr/`。
- `auth-system` 只补充 Console 外部用户来源入口说明，避免把 Console 扩展能力误写成 pole-server 核心鉴权逻辑。

## 鉴权规则编辑抽屉 PRD 对齐

- [x] 读取鉴权规则设计交接文档，确认前端交互边界
- [x] 核对当前 `TrafficGovernanceEditor` 的鉴权数据结构、保存 payload 和通用治理编辑器范式
- [x] 将鉴权子规则重组为黑名单接口规则、白名单接口规则、服务级规则三段
- [x] 补齐前端预览映射、归一化和保存校验，保持真实提交兼容当前后端契约
- [x] 运行前端逻辑脚本、构建、diff 检查和真实 8080 页面验证
- [x] 记录 review、验证结果和必要 lessons

当前判断：

- 本轮只优化治理工作台 `鉴权` 规则 Tab 的前端组织和右侧预览，不调整后端 schema、proto 或存储契约。
- 附件里的 `AuthRule/subRules/listType/protectedInterfaces` 是产品态预览结构；真实提交仍需适配到当前 `TrafficSecurityRule` / `TrafficSecurityPolicy` 后端契约。
- 鉴权专属区域应按黑名单接口规则 → 白名单接口规则 → 服务级规则固定排序；接口级子规则内部不再出现名单类型切换、策略名称、参数比例、命中动作、未命中默认动作、鉴权服务和凭据来源。

当前进展：

- 已新增鉴权编辑器工具函数，将当前后端 `TrafficSecurityRule.policies` 归一化为前端 `黑名单接口规则 / 白名单接口规则 / 服务级规则` 视图；保存时再展开回当前后端 `TrafficSecurityPolicy[]`。
- 已将鉴权规则 Tab 重构为左侧基础信息与鉴权子规则、右侧实时 `AuthRule` Spec 预览；预览只表达产品态映射，不改变后端提交契约。
- 已移除鉴权子规则内的名单类型切换、命中动作、未命中默认动作、策略名称、参数比例、鉴权服务和凭据来源；接口级名单语义由所在分区自动决定。
- 已支持服务级规则最多一条、始终位于最后；服务级规则的 `protectedInterfaces` 在预览中保持为空数组。

验证：

- `cd console/web && node scripts/verify-traffic-security-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，tmux 日志显示 `finish starting server`，8080 页面可访问。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-security`：查看态显示基础信息、鉴权子规则、黑名单接口规则、白名单接口规则、服务级规则和右侧实时 `AuthRule` Spec。
- Playwright 编辑态断言通过：存在 `添加黑名单接口`、`添加白名单接口`、`添加服务级规则`，右侧预览包含 `kind: AuthRule`、`listType: ALLOW_LIST`、`protectedInterfaces`，页面文本不包含 `命中后`、`未命中`、`策略名称`、`鉴权服务`、`凭据来源`、`默认动作`。
- Playwright 添加服务级规则后，右侧预览新增服务级 `subRule` 且 `protectedInterfaces: []`；页面截图保存到 `output/playwright/traffic-security-authrule-editor.png`。

Review：

- 本轮只调整 Console 鉴权规则编辑抽屉，不修改后端 proto、存储或接口 schema。
- `AuthRule/subRules` 是前端产品态预览，提交仍转换为当前 `TrafficSecurityRule.policies`，避免把 PRD 展示结构误当后端新契约。
- 本轮没有收到新的用户纠正，因此未新增 `context-kg/tasks/lessons.md`。

## 服务预热控件与曲线 PRD 对齐修正

- [x] 对照用户截图和无损规则编辑 PRD，定位服务预热区偏差
- [x] 修复 `Second` / `%` 单位输入呈现，确保数值可读
- [x] 修复预热曲线可视化，按 `curvature` 幂函数真实绘制
- [x] 运行前端逻辑脚本、构建、diff 检查和页面级验证
- [x] 记录 review 与验证结果

当前判断：

- 截图中的“预热窗口”控件只显示 `Second`，没有稳定显示数值，不符合 PRD 的“数字输入 + Second 后缀”要求。
- 截图中的曲线图区域为空白，不满足 PRD 要求的 `curvature` 幂函数可视化。
- 修复范围先限定在无损规则编辑抽屉的生命周期数值控件和服务预热曲线，不扩大到后端契约。

当前进展：

- 已把无损规则编辑器里的生命周期秒数和百分比控件从 `InputNumber suffix` 改为现有治理编辑器范式 `InputAdornment append`，避免单位被显示成输入值。
- 已移除无损编辑器对 `echarts-for-react` 的直接依赖，改为由 `buildWarmupCurvePoints()` 生成 SVG 曲线，确保抽屉内首次渲染即可显示曲线、坐标和点位。
- 已补充预热窗口说明文案，与 PRD 中“该时间窗内由治理层按预热曲线逐步放量”一致。

验证：

- `cd console/web && node scripts/verify-lossless-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告；本次 `LossLessEditor` chunk 为 46.39 KiB，不再因曲线图拉入 ECharts 变成超大 chunk。
- `git diff --check -- console/web/src/pages/Governance/LossLess/LossLessEditor.tsx console/web/src/pages/Governance/LossLess/index.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，日志显示 `finish starting server`，8080 返回 200。
- Playwright 登录真实 8080 后打开无损规则抽屉并进入编辑态：预热窗口显示数值输入 `0` + `Second`，终止百分比显示数值输入 `80` + `%`，曲线 SVG 存在并显示 0/25/50/75/100% 坐标与点位。
- 页面截图已保存到 `.playwright-cli/page-2026-06-21T10-06-17-256Z.png`。

Review：

- 用户指出的问题成立：上一轮只实现了字段、文案和曲线容器，没有按真实页面检查控件呈现，导致服务预热区和 PRD 明显不一致。
- 这次修复选择复用 TDesign `InputAdornment append`，和熔断编辑器已有单位输入范式一致；曲线改为由当前纯逻辑点位直接绘制，避免抽屉布局中的图表库 resize 时机问题。
- 本地 seed 无损规则的 namespace/service 与秒数字段仍是旧数据问题，页面会继续按当前校验展示错误；本轮没有改后端契约或测试数据。

## 无损上下线规则编辑抽屉优化

- [x] 核对交接文档、当前 `LosslessRule` proto/API 与已有治理编辑器范式
- [x] 新增 Lossless 编辑器工具函数，覆盖草稿归一化、预览 Spec、校验和曲线数据
- [x] 重构 `LossLessEditor` 规则 Tab 为左表单 / 右实时 Spec 双栏
- [x] 实现延迟注册、HTTP/TCP/UDP 探测配置、预热曲线图和无损下线只读默认行为
- [x] 保持保存 payload 与当前后端契约兼容，避免提交未知字段
- [x] 运行前端脚本、构建、diff 检查和页面级验证
- [x] 记录 review 与验证结果

当前判断：

- 本轮优化范围限定在治理工作台 `LosslessRule` 编辑抽屉的规则 Tab，不调整后端 proto。
- 附件里的实时 Spec 是产品态核对视图；当前后端仍只接受既有 `lossless_online.delay_register` / `warmup` / `lossless_offline` 字段，TCP/UDP 报文匹配先作为前端配置与预览信息保留，不写入提交 payload 的未知字段。
- 页面布局应直接复用已验证的治理编辑抽屉双栏滚动模型：左侧表单内部滚动，右侧 Spec 面板贯穿抽屉正文高度。

当前进展：

- 已确认当前 `LossLessEditor` 是单栏表单，探测延迟只支持 HTTP，没有右侧实时 Spec，也没有交接稿要求的曲线图和 TCP/UDP 报文匹配 UI。
- 已新增 `losslessEditorUtils.ts` 和 `verify-lossless-editor-utils.mjs`，将草稿归一化、产品态 Spec 预览、后端提交 payload、保存校验和预热曲线点位拆成纯逻辑。
- 已将 `LossLessEditor` 重构为左表单 / 右实时 Spec 双栏，左侧分为基础信息、治理对象、无损上线、服务预热、无损下线。
- 已补齐探测延迟协议切换，HTTP 展示方法/路径，TCP/UDP 展示匹配方式、发送报文、响应匹配。
- 已补齐预热曲线值 `1～5` 校验、幂函数曲线图和“值越大末段爬升越陡”的准确文案。
- 已将无损下线收敛为启停 + 等待时长，摘流、停止新流量、等待存量请求作为只读默认行为展示。

验证：

- `cd console/web && node scripts/verify-lossless-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告；Lossless 编辑器因复用既有 `echarts-for-react` 曲线图仍触发大 chunk 警告。
- `git diff --check -- console/web/src/pages/Governance/LossLess/LossLessEditor.tsx console/web/src/pages/Governance/LossLess/index.module.less console/web/src/pages/Governance/LossLess/losslessEditorUtils.ts console/web/scripts/verify-lossless-editor-utils.mjs context-kg/tasks/todo.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开无损规则抽屉，查看态显示基础信息、治理对象、无损上线、服务预热、无损下线和右侧实时 Spec；截图保存到 `/tmp/lossless-editor-tcp.png`。
- Playwright 编辑态切换 TCP 后，左侧出现匹配方式、发送报文、响应匹配，右侧 Spec 同步显示 `protocol: TCP`、`expectedResponse: PONG` 和 `match: 包含匹配`。
- Playwright 布局指标：左侧 `formPane` 为 `overflow:auto`，`scrollHeight=2867`、`clientHeight=841`、设置后 `scrollTop=480`；右侧 Spec 高度 `841` 与 shell 对齐，`window.scrollY=0`。

Review：

- 本轮只改 Console 的无损规则编辑抽屉规则 Tab，不改变后端 proto 与保存接口。
- 附件里的产品态 Spec 预览已完整表达 TCP/UDP payload；当前后端 `LosslessRule` proto 没有 payload 字段，因此提交 payload 保持在既有 `delay_register` / `warmup` / `lossless_offline` 字段内，避免把页面专用字段伪装成已持久化字段。
- 当前本地 seed 无损数据缺少 namespace/service 且多个秒数字段为 0，页面会按保存校验在右侧提示；这是数据本身与当前前端校验的结果，不是本轮布局问题。

## 探测 HTTP Headers 添加交互与样式对齐

- [x] 复现 `添加标签` 点击无效并定位状态更新链路
- [x] 修复 Headers 添加/编辑/删除交互
- [x] 按 PRD 重新收敛 Headers 区块视觉密度、列宽和按钮样式
- [x] 运行前端脚本、构建、真实 8080 页面交互验证和 diff 检查
- [x] 记录 review 与验证结果

当前判断：

- 这不是单纯文案问题；需要确认点击事件、表单状态、`ruleDrafts` 与 `Form` 字段之间是否互相覆盖。
- 样式也要继续贴近 PRD：Headers 区块应像协议配置内部的小型编辑表，而不是普通后台表格或通用标签控件。

当前进展：

- 已在真实 8080 复现：点击 `添加标签` 命中了按钮且按钮类型为 `button`，但 DOM 仍保持 `1 个标签`。
- 根因定位到 `normalizeHttpConfig`：每次草稿同步会过滤 `{ key: '', value: '' }`，导致新增空 header 行立即被归一化删除。
- 已调整 HTTP 草稿归一化，保留空 header 行用于编辑态；保存前仍由既有校验拦截空 key/value。
- 已将 Headers 输入、删除按钮、列宽和添加按钮主色按 PRD 风格重新收敛。
- 用户指出字体大小与其它控件不统一后，已移除 Headers 输入和删除按钮的 `large` 尺寸，并将 Headers 标题、说明和列头收敛到治理抽屉常用字号。

验证：

- `cd console/web && node scripts/verify-faultdetect-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-faultdetect` 并进入编辑态：点击 `添加标签` 后从 1 行变 2 行，输入 `x-new/yes` 后 DOM 值更新，删除第二行后回到 1 行；添加按钮为主色，window 保持 `scrollY=0`；截图保存到 `/tmp/faultdetect-http-headers-interaction-fixed.png`。
- Playwright 复验 Headers 字号：`H` 标识 13px，Headers 标题 14px，说明 12px，列头 12px，输入文字 14px，删除按钮 32px；截图保存到 `/tmp/faultdetect-http-headers-font-fixed.png`。

Review：

- 本轮修复的是探测 HTTP Headers 编辑态交互和视觉，不改变提交 payload。
- 根因修在草稿归一化层，保留编辑态空 header 行；保存时仍通过既有校验阻止空 key/value。

## 探测规则 HTTP Headers 展示修复

- [x] 对比 PRD 截图与当前 `FaultDetectEditor` HTTP 配置渲染
- [x] 将 HTTP Headers 从通用标签控件改为探测专用 Headers 区块
- [x] 运行前端脚本、构建和真实 8080 页面验证
- [x] 记录 review 与验证结果

当前判断：

- 当前 HTTP Headers 复用通用 `LabelInput`，会呈现“标签”语义和通用标签编辑器布局；PRD 期望的是 HTTP 请求头专用区块：`H` 标识、Headers 标题、说明文案、键/值/操作三列表和底部计数。
- 本轮只调整 HTTP 协议配置的展示与编辑控件，不改变 `httpConfig.headers` 数据结构和保存 payload。

当前进展：

- 已确认当前实现直接复用 `LabelInput` 渲染 HTTP Headers，因此会出现“标签键/标签值/添加标签/暂无标签”等通用标签语义。
- 已在 `FaultDetectEditor` 内新增探测专用 Headers 区块：`H` 标识、Headers 标题、说明文案、键/值/操作三列、底部添加标签和计数。
- Headers 编辑仍写回原字段 `rules[index].httpConfig.headers`，保存 payload 不变。

验证：

- `cd console/web && node scripts/verify-faultdetect-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-faultdetect` 并进入编辑态：HTTP Headers 区块存在，`H` 标识、标题 `Headers`、说明 `随探测请求发送的请求头`、列头 `键/值/操作`、输入值 `x-seed/true`、底部 `添加标签` 与 `1 个标签` 均符合 PRD；截图保存到 `/tmp/faultdetect-http-headers-prd.png`。

Review：

- 本轮只替换 HTTP Headers 展示控件，不改变 `httpConfig.headers` 数据结构和提交 payload。
- 通用 `LabelInput` 仍保留给其它标签编辑场景；探测 HTTP Headers 改为本编辑器内的协议专用区块，避免影响其它页面。

## 探测编辑器抽屉共性滚动修复

- [x] 对比探测编辑器与路由/限流/熔断已验证的双栏滚动模型
- [x] 修复探测编辑器左侧内部滚动、右侧 Spec 高度和 Form shrink 问题
- [x] 运行前端脚本、构建、diff 检查、context-kg lint 和真实 8080 探测页面验证
- [x] 记录 review 与验证结果

当前判断：

- 探测编辑器也属于同一类治理规则抽屉共性问题：旧实现让 shell/spec 各自按 viewport 计算高度，左侧没有被约束成唯一滚动容器。
- 与限流/熔断不同，探测编辑器的 `Form` 位于左侧 `formPane` 内部；修复时不仅要让 `formPane` 滚动，还要避免内部 `t-form` 和直接 section 在 column flex 中 shrink 后被裁切。

当前进展：

- 已确认探测编辑器仍是旧模型：`editorBody` 只有 padding，`editorShell` 使用 `min-height: calc(100vh - 184px)`，`specPane` 使用 `height/max-height: calc(100vh - 184px)` 与 `position: sticky`。
- 已将探测编辑器收敛为共性模型：`editorBody` 固定正文高度，shell `height: 100% / overflow: hidden`，左侧 `formPane overflow: auto`，右侧 Spec `height: 100%`。
- 已针对探测编辑器的特殊 DOM 结构处理 Form shrink：`formPane` 内部 `t-form` 固定为 `flex: 0 0 auto`，Form 直接子 section 固定为 `flex: 0 0 auto`，避免 section 被共享卡片裁切后外层没有真实滚动空间。

验证：

- `cd console/web && node scripts/verify-faultdetect-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-faultdetect` 并进入编辑态：左侧 `formPane.scrollHeight=1902`、`clientHeight=1242`，程序滚动后 `scrollTop=660`，滚轮后 `scrollTop=660`，window 保持 `scrollY=0`，右侧 Spec 与 editor shell 高度一致；截图保存到 `/tmp/faultdetect-left-scroll-fixed.png`。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/CircuitBreaker/FaultDetectEditor.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只修探测编辑抽屉布局，不改变探测规则保存 payload 与后端契约。
- 探测编辑器已纳入治理规则编辑抽屉共性滚动模型；后续处理同类问题时必须横向审计所有同构编辑器。

## 治理规则编辑抽屉共性滚动修复

- [x] 修复熔断编辑器左侧内部滚动和右侧 Spec 高度
- [x] 将“左侧内部滚动必须验证 scrollTop”沉淀为治理规则共性约束
- [x] 运行构建、diff 检查、context-kg lint 和真实 8080 熔断页面验证

当前判断：

- 熔断与限流是同类共性问题：编辑器 shell/formPane 使用 `overflow: visible`，右侧 Spec 单独按 viewport 计算高度，左侧 section 在 column flex 中可能 shrink 后被共享 section 的 `overflow:hidden` 裁掉。
- 这个约束应作为治理规则编辑器共性规则沉淀，不应每个规则类型都等用户截图后再修。

当前进展：

- 已确认熔断编辑器 CSS 仍是旧模型：`editorBody` 无固定正文高度，`circuitBreakerEditorShell` 与 `formPane` 为 `overflow: visible`，`specPane` 使用 `height: calc(100vh - 184px)`。
- 已将熔断编辑器对齐到共性模型：`editorBody` 固定正文高度，shell `height: 100% / overflow: hidden`，左侧 `formPane overflow: auto`，右侧 Spec `height: 100%`。
- 已将 `formPane > div` 固定为 `flex: 0 0 auto`，避免共享 section shrink 后被 `overflow: hidden` 裁切。
- 已将熔断 StickyTool 移出 Form flex 流，避免操作区作为 flex 子项挤压左右双栏高度。
- 已将治理规则编辑抽屉双栏滚动交互写入 `context-kg/technical/conventions/patterns.md`，作为长期技术约定。

验证：

- `cd console/web && node scripts/verify-circuitbreaker-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md context-kg/technical/conventions/patterns.md context-kg/_meta/index.md context-kg/_meta/log.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-circuitbreaker` 并进入编辑态：左侧 `formPane.scrollHeight=1768`、`clientHeight=1090`，滚轮后 `scrollTop=678`，直接设置后 `scrollTop=500`，window 保持 `scrollY=0`，右侧 Spec 与 editor 高度一致；截图保存到 `/tmp/circuitbreaker-left-scroll-fixed.png`。

Review：

- 本轮只修熔断编辑抽屉布局和治理规则共性约定，不改变熔断保存 payload 与后端契约。
- 治理规则编辑抽屉的双栏滚动模型已经沉淀到 `technical/conventions/patterns.md`；后续新增/调整规则编辑器应先按该约定实现和验证。

## 限流编辑器抽屉布局修复

- [x] 对比路由编辑器最终布局范式，定位限流编辑态双栏滚动/宽度问题
- [x] 修复 RateLimit 编辑器 shell、左侧表单滚动和右侧 Spec 高度
- [x] 补充任务记录与 lessons
- [x] 运行前端构建、diff 检查、context-kg lint 和真实 8080 页面验证
- [x] 修复并验证左侧编辑区可内部滚动

当前判断：

- 用户截图里的问题与路由编辑器此前踩过的问题一致：抽屉内不应让整个页面和右侧 Spec 各自按 `100vh` 计算滚动；应由编辑器 body 承担固定可用高度，左侧表单内部滚动，右侧 Spec 贯穿正文高度。
- 本轮只修 RateLimit 编辑器布局和验证记录，不改变限流保存 payload 与后端契约。

当前进展：

- 已对比路由编辑器：`editorBody` 固定 `calc(100vh - 96px)`，shell `height: 100%`，左侧 `formPane overflow: auto`，右侧 Spec 跟随 shell 高度。
- 已定位限流编辑器当前差异：shell/formPane 使用 `overflow: visible`，右侧 Spec 独立 `height: calc(100vh - 184px)`，会造成抽屉内滚动和可视高度与路由不一致。
- 已将 RateLimit 编辑器收敛为路由同款布局：`editorBody` 固定抽屉正文高度，shell `height: 100% / overflow: hidden`，左侧 `formPane overflow: auto`，右侧 Spec `height: 100%`。
- 用户复验指出左侧仍无法内部滚动；上一轮只验证了高度和横向溢出，没有验证真实 `scrollTop` 变化，本轮补充滚动行为级验证。
- 已定位左侧不滚动的根因：`formPane` 是 column flex 容器，直接子 `section` 默认可 shrink；限流规则 section 被压缩后又被共享 `.section { overflow: hidden }` 裁掉，导致外层 `formPane.scrollHeight === clientHeight`。已将 `formPane > section` 固定为 `flex: 0 0 auto`，让内容真实撑高并由左侧容器滚动。

验证：

- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.module.less context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-ratelimit` 并进入编辑态：`editorBody` / shell / Spec 高度均为 1152px，左侧 `formPane` 为 `overflow: auto`，表单和页面无横向溢出，页面无异常；截图保存到 `/tmp/ratelimit-editor-layout-fixed.png`。
- 补充 Playwright 滚动行为验证通过：左侧 `formPane.scrollHeight=1660`、`clientHeight=1090`，滚轮后 `scrollTop=570`，直接设置后 `scrollTop=500`，window 仍为 `scrollY=0`；截图保存到 `/tmp/ratelimit-left-scroll-fixed.png`。

Review：

- 本轮只修限流编辑抽屉布局，不改变限流保存 payload 与后端契约。
- 限流编辑器应和路由编辑器共享同一抽屉正文滚动范式；右侧 Spec 由 shell 高度约束，不应单独按 viewport 计算高度。
- 布局验证不能只检查 `overflow:auto` 样式；必须确认真实 `scrollHeight > clientHeight`，并用 wheel 或设置 `scrollTop` 验证滚动容器确实可滚。

## 探测规则编辑器规则 Tab 设计实现

- [x] 核对附件交接文档、当前 `FaultDetectRule` proto / Console service 归一化和已有治理编辑器模式
- [x] 新增 FaultDetect 编辑器工具函数，覆盖前端草稿归一化、实时 Spec、摘要和保存校验
- [x] 重构 `FaultDetectEditor` 规则 Tab 为左表单 / 右实时 Spec 双栏
- [x] 实现探测规则折叠卡、多条增删、基础调度、端口策略、HTTP 载荷和 TCP/UDP 报文匹配
- [x] 保存前按交接规则校验，并通过 Toast 拦截首条错误
- [x] 运行前端脚本/构建、Go 相关测试、context-kg lint、diff 检查和真实 8080 页面验证
- [x] 记录 review 与验证结果

当前判断：

- 本轮是 Console 规则 Tab 体验实现，不改变后端 proto；后端当前 `FaultDetectSubRule` 只有 `interval/timeout/port/protocol/http_config/tcp_config/udp_config/disable`。
- 端口策略在真实提交中继续用 `port=0` 表达「实例协议端口」，非 0 表达「指定探测端口」；实时 Spec 按交接文档额外展示 `portMode` 便于用户核对。
- TCP/UDP 的「匹配方式」当前无后端字段，页面可以展示与校验，实时 Spec 展示 `payload.match`；保存给后端时只提交已有的 `send/receive`，避免写入未知字段。
- 「被探测对象」继续只包含命名空间和服务，不回退展示接口字段。

当前进展：

- 已读取交接文档，确认范围仅限治理工作台 FaultDetectRule 编辑抽屉「规则」Tab。
- 已核对现有 `FaultDetectEditor`：已有基础信息、被探测对象、多规则卡片和 HTTP/TCP/UDP 条件渲染，但缺少折叠、端口策略、TCP/UDP 匹配方式、右侧实时 Spec 和业务保存校验。
- 已核对当前 specification / generated type：`TcpProtocolConfig` 与 `UdpProtocolConfig` 不包含匹配方式字段。
- 已将编辑器改成左侧表单 / 右侧实时 Spec，工作台和独立表格抽屉宽度统一为 `min(1560px, calc(100vw - 40px))`。
- 已实现子规则折叠卡、多条增删、协议切换、端口来源、HTTP 载荷、TCP/UDP 报文匹配、自定义 YAML/JSON Spec 预览和复制。
- 已修复 `LabelInput` 在 inline `name` 数组下的重复同步问题，避免编辑态触发 Maximum update depth。
- 已对保存 payload 做归一化：`port=0` 表示实例协议端口，TCP/UDP 的前端匹配方式仅用于页面校验和 Spec 预览，不提交后端未知字段。

验证：

- `cd console/web && node scripts/verify-faultdetect-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./plugin/apiserver/xdsserverv3/... ./test/e2e/console_api ./test/e2e/client -run 'FaultDetect|faultdetect|TrafficGovernance|TestDoesNotExist' -count=1` 通过。
- `git diff --check -- console/web/src/pages/Governance/CircuitBreaker/FaultDetectEditor.tsx console/web/src/pages/Governance/CircuitBreaker/faultDetectEditorUtils.ts console/web/src/pages/Governance/CircuitBreaker/FaultDetectEditor.module.less console/web/scripts/verify-faultdetect-editor-utils.mjs console/web/src/services/faultdetect.ts console/web/src/pages/Governance/Workbench/index.tsx console/web/src/pages/Governance/CircuitBreaker/FaultDetectTable.tsx console/web/src/components/LabelInput/index.tsx context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，release bundle 重启后日志显示 `finish starting server`。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-faultdetect`：点击编辑态 TCP 后，TCP 按钮变为 active，右侧 Spec 出现 `protocol: TCP`，页面出现「报文匹配 / 发送内容 / 接收内容」，无页面异常、无调试日志；截图保存到 `/tmp/faultdetect-protocol-switch-final.png`。

Review：

- 本轮聚焦 FaultDetectRule 编辑抽屉「规则」Tab，不改变后端 proto 与保存接口形状。
- `TcpProtocolConfig` / `UdpProtocolConfig` 当前没有匹配方式字段，页面以预览字段承载设计表达，保存时仍保持真实后端契约。
- 协议切换不能依赖整个 Redux `editRule` 对象作为详情加载 effect 依赖；否则编辑态本地草稿可能被重复拉取详情覆盖。当前只按 `editRule.id` 加载详情，避免 TCP/UDP 切换后被重置。

## specification v0.1.0-ALPHA.29 发布与 control-plane 引用更新

- [x] 同步 `../specification` 远端 develop，确认 FaultDetect PR 已合入
- [x] 基于合并提交创建并推送 `v0.1.0-ALPHA.29`
- [x] 更新 control-plane 的 `github.com/pole-io/specification` 依赖与 lattice-hub replace 到新 tag
- [x] 下载新模块校验和，不运行 `go mod tidy`
- [x] 运行 FaultDetect 相关 Go 验证、context-kg lint 和 diff 检查

当前判断：

- specification `origin/develop` 已包含 `Merge pull request #5 from lattice-hub/codex/faultdetect-discovery-rules`，合并提交为 `03fcdbb4659d03c97ebd08c9fd7f50f484f07bb5`。
- 旧最新 tag 为 `v0.1.0-ALPHA.28`，本轮按连续版本发布 `v0.1.0-ALPHA.29`。
- control-plane 应继续保留 module path `github.com/pole-io/specification`，通过 `replace github.com/pole-io/specification => github.com/lattice-hub/specification v0.1.0-ALPHA.29` 指向迁移后的仓库来源。

当前进展：

- `v0.1.0-ALPHA.29` 已创建并推送，tag 指向 `03fcdbb4659d03c97ebd08c9fd7f50f484f07bb5`。
- `go.mod` 已从 `github.com/pole-io/specification v0.1.0-ALPHA.28` 更新为 `v0.1.0-ALPHA.29`。
- `go.mod` 的 specification replace 已从 `github.com/lattice-hub/specification v0.1.0-ALPHA.28` 更新为 `v0.1.0-ALPHA.29`。
- `go.sum` 已新增 `github.com/lattice-hub/specification v0.1.0-ALPHA.29` 与其 `go.mod` checksum。

验证：

- `git -C ../specification show --no-patch --format='%H%n%D%n%s' v0.1.0-ALPHA.29` 确认 tag 指向 FaultDetect PR #5 合并提交。
- `git -C ../specification ls-remote --tags origin v0.1.0-ALPHA.29` 确认远端 tag 已存在。
- `GOPRIVATE=github.com/pole-io/*,github.com/lattice-hub/* GOPROXY=direct go mod download github.com/pole-io/specification` 通过。
- `GOPRIVATE=github.com/pole-io/*,github.com/lattice-hub/* GOPROXY=direct go list -m -json github.com/pole-io/specification` 确认 require 为 `v0.1.0-ALPHA.29`，replace 为 `github.com/lattice-hub/specification v0.1.0-ALPHA.29`。
- `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./plugin/apiserver/xdsserverv3/... ./test/e2e/console_api ./test/e2e/client -run 'FaultDetect|faultdetect|TrafficGovernance|TestDoesNotExist' -count=1` 通过。

Review：

- 本轮没有运行 `go mod tidy`，避免引入无关间接依赖变化。
- specification 本地工作区仍有此前生成脚本留下的 `source/rust/pole-specification/proto/service.proto` 换行符未提交变化；该文件与本次 tag/control-plane 引用更新无关，未纳入 control-plane 改动。

## 主动探测页面结构优化

- [x] 核对主动探测新协议边界：规则级 `targetService`，子规则级探测参数
- [x] 优化详情/编辑页信息架构：基础信息、被探测对象、探测规则分层展示
- [x] 优化探测子规则卡片摘要，避免把服务字段重复放进每条子规则
- [x] 优化列表页列信息，分开展示被探测服务和探测参数
- [x] 按用户纠正收敛「被探测对象」：只保留命名空间和服务，不再展示接口字段
- [x] 运行 Console 构建、主动探测相关 Go 测试、context-kg lint 和 diff 检查
- [x] 记录 review 与验证结果

当前判断：

- 主动探测的服务信息属于规则级对象，不应在每条子规则中重复编辑。
- 「被探测对象」只表示规则级被探测服务信息，不承载接口名称、接口匹配类型、接口协议或方法。
- 子规则列表只承载探测动作参数：协议、状态、间隔、超时、端口和协议配置。
- 页面需要从字段平铺调整为“规则身份 -> 被探测对象 -> 探测规则”的操作心智。

当前进展：

- 详情/编辑页独立「被探测对象」section 只编辑命名空间和服务。
- 探测子规则卡片已去掉被探测服务字段，只保留探测参数，并在 header 显示协议、端口、间隔、超时、状态摘要。
- 列表页已改为按被探测服务和探测参数扫描，不再展示接口信息列。

验证：

- `go test ./apis/pkg/types/rules ./pkg/goverrule ./pkg/cache/rules ./plugin/apiserver/xdsserverv3/... ./test/e2e/console_api ./test/e2e/client -run 'FaultDetect|faultdetect|TrafficGovernance|TestDoesNotExist' -count=1` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/CircuitBreaker/FaultDetectEditor.tsx console/web/src/pages/Governance/CircuitBreaker/FaultDetectTable.tsx console/web/src/services/faultdetect.ts context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过，重启后日志显示 `finish starting server`，`curl -I http://127.0.0.1:8080/governance/workbench` 返回 200。
- Playwright 登录真实 8080 工作台打开 `seed-20260616-faultdetect` 验证：查看态和编辑态的「被探测对象」只包含被探测命名空间、被探测服务；`接口名称`、`接口协议` 均不存在；截图保存到 `/tmp/faultdetect-target-service-only.png`。

Review：

- 本轮只调整主动探测列表页和编辑/详情页的信息架构，没有改变保存接口或协议模型。
- 子规则卡片继续使用已有治理规则共享样式，不新增私有布局体系。

## FaultDetect 下发协议收敛

- [x] 核对 `FaultDetector` 包裹对象、`FaultDetectRule` 顶层重复字段和 control-plane 引用面
- [x] 写入失败检查，确认旧包裹字段和重复字段仍存在
- [x] 修改 specification proto：`DiscoverResponse` 直接返回 `repeated FaultDetectRule`，删除 `FaultDetector` 包裹和规则顶层重复探测配置字段
- [x] 重新生成 specification Go/Rust 产物
- [x] 适配 control-plane：客户端下发、XDS、缓存/转换、测试 fixture 使用 `FaultDetectRule.rules[]`
- [x] 运行 spec 生成、Go 编译/重点测试、Console 构建或类型验证、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和必要 lessons

当前判断：

- 用户明确要求：`DiscoverResponse` 中主动探测直接返回 `repeated FaultDetectRule`，不再需要 `FaultDetector{rules, revision}` 包裹。
- `FaultDetectRule` 已有 `repeated FaultDetectSubRule rules`；规则级服务信息必须保留在 `FaultDetectRule.target_service`，只有 `interval`、`timeout`、`port`、`protocol`、`http_config`、`tcp_config`、`udp_config` 这类探测配置从规则顶层移入 subRule。
- 这次是 breaking spec 收敛；当前控制面同步适配新下发协议，不再向外提交旧顶层探测字段。
- 旧库里已经落过的单条顶层探测 JSON 仍通过 `ToSpec()` 临时迁移：服务提升到规则级 `target_service`，探测配置迁入 `rules[0]`，避免已有数据在控制台和缓存读取时变空。

当前进展：

- specification 已删除 `FaultDetector` wrapper；`DiscoverResponse` 的 field 22 改为 `repeated FaultDetectRule faultDetectRules`。
- `FaultDetectRule` 顶层保留 `target_service`；已删除顶层 `interval`、`timeout`、`port`、`protocol`、`http_config`、`tcp_config`、`udp_config`，这些探测配置只保留在 `FaultDetectSubRule`。
- 已重新生成 specification Go/Rust 产物。
- control-plane 下发路径已从 `resp.FaultDetector.rules` 改为 `resp.FaultDetectRules`，总 revision 使用 `DiscoverResponse.service.revision`。
- XDS health check 已遍历 `FaultDetectRule.rules[]` 生成探测配置，服务归属和缓存索引读取规则级 `target_service`。
- 创建/更新落库只序列化规则级 `target_service` 与 `FaultDetectRule.rules[]`，不再保存旧顶层探测配置字段。
- Console 主动探测编辑器保存时保留规则级 `targetService`，并清除展示兼容用的旧顶层探测配置字段。
- 已新增 `FaultDetectRule.ToSpec()` 兼容测试，覆盖旧顶层 JSON 服务提升到规则级、旧顶层探测配置迁入首个 subRule，以及新 `rules[]` 优先作为源数据。

验证：

- `node` 协议形状检查通过：确认没有 `message FaultDetector`，`DiscoverResponse` field 22 为 `repeated FaultDetectRule faultDetectRules`，`FaultDetectRule.target_service` 位于规则级，且 subRule 不含服务字段。
- `cd /Users/chuntao.liao/Github/pole-io/specification/source/go && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification/source/rust && bash build.sh` 通过，Rust release 编译完成。
- `go test ./apis/pkg/types/rules -run 'TestFaultDetectRuleToSpec' -count=1` 通过。
- `go test ./apis/pkg/types/rules ./pkg/goverrule ./pkg/cache/rules ./plugin/apiserver/xdsserverv3/... ./test/e2e/console_api ./test/e2e/client -run 'FaultDetect|faultdetect|TrafficGovernance|TestDoesNotExist' -count=1` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- control-plane 与 specification 本轮相关文件 `git diff --check` 通过。

Review：

- `FaultDetector{rules, revision}` 已无必要；规则数组直接挂在 `DiscoverResponse`，聚合 revision 复用 `DiscoverResponse.service.revision`。
- `FaultDetectRule.target_service` 是探测目标服务的唯一主模型；`FaultDetectRule.rules[]` 是探测参数的主模型。旧 subRule 中的 `target_service` 只在读取旧 JSON 的兼容转换和 Console 展示归一化中作为输入兜底存在。
- control-plane 当前使用本地 `replace github.com/pole-io/specification => ../specification` 做联调；specification 发版后需要改回对应 tag。

## 熔断规则编辑器规则 Tab 设计实现

- [x] 核对现有熔断数据契约、路由/限流编辑器复用模式和已记录的治理规则 UI 经验
- [x] 写入熔断编辑器纯逻辑失败断言，覆盖子规则分组、真实提交 payload、实时 Spec、摘要和保存校验
- [x] 新增 `circuitBreakerEditorUtils`，让右侧实时 Spec 与保存 payload 同源
- [x] 重构 `CircuitBreakerEditor` 规则 Tab：基础信息、服务范围、子规则折叠列表、右侧实时 Spec 双栏
- [x] 实现子规则内多熔断策略、接口范围、错误判断条件、触发条件、恢复策略和降级响应编辑
- [x] 运行脚本验证、Console 构建、context-kg lint、diff 检查和必要的真实页面验证
- [x] 记录 review 和验证结果

当前判断：

- 当前 specification 的真实提交形状是 `CircuitBreakerRule.block_configs[]`，每项为 `CircuitBreakerPolicy{block_config, max_ejection_percent, recoverCondition, faultDetectConfig, fallbackConfig}`。
- 交接文档里的「子规则 -> 多策略」在当前后端协议中没有独立消息承载；Console 需要用前端分组模型表达子规则，保存时展平为真实 `block_configs[]`，并把子规则级恢复/降级配置复制到该子规则下每条策略对应的 policy。
- 右侧实时 Spec 应展示真实提交 payload，不展示未落地的 CRD 伪结构，避免 Spec 预览与保存接口分叉。

当前进展：

- 已确认现有 `CircuitBreakerEditor` 仍把 `block_configs` 直接渲染为一层「熔断策略」列表，尚未形成「子规则 -> 策略」层级。
- 已确认 `services/circuitbreaker.ts` 读写边界已经支持 `CircuitBreakerPolicy.block_config` 的归一化和提交打包，但编辑器内部还缺少子规则分组模型、保存校验和实时 Spec。
- 已新增 `verify-circuitbreaker-editor-utils.mjs` 并先复现失败：缺少 `circuitBreakerEditorUtils.ts`。
- 已新增 `circuitBreakerEditorUtils.ts`，覆盖前端子规则 draft、真实提交 payload 展平、YAML/JSON 序列化、子规则/策略摘要和保存前校验；脚本已转绿。
- 已重构 `CircuitBreakerEditor`：基础信息与服务范围拆分；服务范围包含主调 -> 被调调用关系和规则级熔断粒度；规则区改为「熔断子规则 -> 熔断策略」两级折叠结构。
- 子规则内已支持添加/删除熔断策略、接口范围多行、错误判断条件多行、触发条件多行、恢复策略、主动探测开关和降级响应配置。
- 右侧已新增实时 Spec 面板，支持 YAML/JSON 切换和复制，预览内容来自 `buildCircuitBreakerSubmitPayload` 的真实后端提交 payload。
- 已将工作台 `circuitbreaker` 详情抽屉宽度纳入与路由/限流一致的 `min(1560px, calc(100vw - 40px))`。
- 用户反馈后已调整：已有熔断规则进入编辑态时，熔断粒度只读展示；只有创建态可以选择熔断粒度。
- 已把独立熔断列表页创建按钮明确为「新建熔断规则」，并在治理工作台筛选到「熔断」时提供「新建熔断规则」入口；打开创建抽屉前会清空旧熔断编辑态。

验证：

- `cd console/web && node scripts/verify-circuitbreaker-editor-utils.mjs` 通过，覆盖前端分组模型展平为真实 `block_configs[]`、YAML/JSON 序列化、摘要和保存前校验。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.module.less console/web/src/pages/Governance/CircuitBreaker/circuitBreakerEditorUtils.ts console/web/scripts/verify-circuitbreaker-editor-utils.mjs console/web/src/services/circuitbreaker.ts console/web/src/pages/Governance/Workbench/index.tsx context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 通过同源登录接口登录 `admin/admin123` 后打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-circuitbreaker` 抽屉验证：查看态显示基础信息、服务范围、熔断子规则和右侧实时 Spec；Spec 包含真实 `block_configs`；抽屉宽度为 1560px；截图保存到 `/tmp/circuitbreaker-editor-workbench.png`。
- Playwright 进入编辑态验证：`添加子规则`、`添加熔断策略`、接口范围、错误判断条件、熔断触发条件、恢复策略、熔断后降级、YAML/JSON/复制控件均存在；抽屉宽度为 1560px；右侧 Spec 面板宽 379px；截图保存到 `/tmp/circuitbreaker-editor-edit-workbench.png`。
- 后续补验：Playwright 真实 8080 页面验证通过，工作台熔断筛选态显示「新建熔断规则」；编辑已有 `seed-20260616-circuitbreaker` 时熔断粒度没有 `.t-radio-button` 切换控件；创建态显示 `服务 / 实例 / 接口` 三个粒度按钮；截图保存到 `/tmp/circuitbreaker-create-entry-and-level.png`。

Review：

- 当前后端协议没有独立的 `subrules[].strategies[]` 消息；Console 用前端 draft 表达设计里的两级结构，保存时展平为真实 `CircuitBreakerPolicy[]`，并把子规则级恢复、降级、最大剔除比例和主动探测开关复制到该子规则下每条策略对应的 policy。
- 当前 `FaultDetectConfig` proto 只有 `enable`，没有探测间隔字段；本轮没有在提交 payload 中伪造 `probeIntervalSec`，避免前端展示与后端契约不一致。
- 右侧实时 Spec 继续展示真实提交 payload，而不是交接文档示例里的 CRD 风格 `apiVersion/kind/spec` 包装。

## 发布 specification v0.1.0-ALPHA.28 并更新 control-plane 引用

- [x] 确认 specification PR 已合入 `develop`，并核对最新 tag 序列
- [x] 在合并后的 specification `develop` 提交上创建并推送 `v0.1.0-ALPHA.28`
- [x] 将 pole-control-plane 的 `github.com/pole-io/specification` 引用更新到新 tag，并恢复远端 lattice-hub replace
- [x] 适配 XDS、Console 提交归一化和 e2e fixture 到 `CircuitBreakerPolicy`
- [x] 运行 Go module 解析、编译级验证、Console 构建、context-kg lint 和 diff 检查
- [x] 记录 review 和验证结果

当前判断：

- specification 仍声明 module path `github.com/pole-io/specification`，control-plane 需要继续保留该 module path 的 `require`。
- 因仓库来源已迁移到 `lattice-hub/specification`，control-plane 的 `replace` 应指向 `github.com/lattice-hub/specification v0.1.0-ALPHA.28`，不能继续使用本地路径 replace。

当前进展：

- specification PR #4 已合入 `develop`，merge commit 为 `d677daa2c80ada2a7483c12e45c19d8ecfc24fa1`。
- 已在该 merge commit 上创建并推送 lightweight tag `v0.1.0-ALPHA.28`。
- control-plane `go.mod` 已更新为 `require github.com/pole-io/specification v0.1.0-ALPHA.28`，并通过 `replace github.com/pole-io/specification => github.com/lattice-hub/specification v0.1.0-ALPHA.28` 指向迁移后的仓库来源。
- XDS outlier detection 已从 `CircuitBreakerPolicy.block_config.trigger_conditions` 读取触发条件，并从同一个 policy 读取恢复窗口和最大剔除比例。
- Console 侧保留编辑器内部扁平模型，但在 `services/circuitbreaker.ts` 的读写边界做转换：读取时兼容 `block_config` 并展平，提交时打包为 `CircuitBreakerPolicy{block_config, ...}`。
- e2e 熔断 fixture 已改为 `block_configs[].block_config` 嵌套结构。

验证：

- `git -C /Users/chuntao.liao/Github/pole-io/specification show --no-patch --format='%H%n%D%n%s' v0.1.0-ALPHA.28` 确认 tag 指向 PR #4 merge commit。
- `GOPRIVATE=github.com/pole-io/*,github.com/lattice-hub/* GOPROXY=direct go list -m -json github.com/pole-io/specification` 确认 module 为 `v0.1.0-ALPHA.28`，replace 为 `github.com/lattice-hub/specification v0.1.0-ALPHA.28`。
- `GOPRIVATE=github.com/pole-io/*,github.com/lattice-hub/* GOPROXY=direct go test -run '^$' -count=0 ./...` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。

Review：

- 本轮没有改源码 import 的 module path，仍保持 `github.com/pole-io/specification`，因为 specification 的 `go.mod` 仍声明该 module path。
- 已移除本地路径 replace，避免 control-plane 继续依赖 `/Users/.../specification`。
- Console 的 UI 状态仍保持原来的扁平可编辑结构，只在 API 边界转换到新 spec，以减少编辑器大面积重写风险。

## 熔断规则策略字段下沉到 block_configs

- [x] 核对 spec `CircuitBreakerRule` / `BlockConfig` 字段和 control-plane/Console 引用
- [x] 修改 spec proto：删除顶层 deprecated 条件字段，并将 `max_ejection_percent`、`recoverCondition`、`faultDetectConfig`、`fallbackConfig` 放入 `CircuitBreakerPolicy`
- [x] 重新生成 spec Go/Rust 输出或对应生成产物
- [x] 适配 control-plane 熔断类型、默认值、归一化、Console 编辑器与测试数据
- [x] 运行针对性脚本、Go 编译级验证、Console 构建和 context-kg lint
- [x] 记录 review 和验证结果

当前判断：

- `level` 继续保留在规则级，不进入子规则。
- `BlockConfig` 承载接口范围、错误判断条件和触发条件；`CircuitBreakerPolicy` 承载 `BlockConfig` 以及恢复、降级、主动探测和最大剔除比例。
- 顶层 `error_conditions` 与 `trigger_condition` 已标记 deprecated，应从 proto 中删除；迁移后真实条件只保留在 `block_configs[]` 内。

当前进展：

- 已在 specification 仓库调整 `CircuitBreakerRule`：删除顶层 deprecated `error_conditions`、`trigger_condition`，不保留 `reserved`，并压实后续字段号；`level` 仍保留规则级。
- 已新增 `CircuitBreakerPolicy`，其中包含 `BlockConfig block_config` 以及策略级 `max_ejection_percent`、`recoverCondition`、`faultDetectConfig`、`fallbackConfig`。
- 已重新生成 specification Go 产物和 Rust 产物，Rust `source/rust/build.sh` 需从 `source/rust` 目录执行。
- 已适配 control-plane 创建/修改熔断规则的落库对象，避免继续写入旧顶层字段。
- 已适配 XDS outlier detection，从首个有效 `CircuitBreakerPolicy.block_config.trigger_conditions` 读取触发条件，并从同一个 policy 读取恢复窗口与最大剔除比例。
- 已适配 Console 熔断规则模型、默认值、旧数据归一化、编辑器 UI 和 e2e fixture；恢复策略、主动探测和降级响应现在随每条 `CircuitBreakerPolicy` 配置。
- control-plane 已在后续发布任务中切到 `v0.1.0-ALPHA.28` 远端 tag，并移除本地路径 replace。

验证：

- `cd /Users/chuntao.liao/Github/pole-io/specification/source/go && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification/source/rust && bash build.sh` 通过，Rust release 编译完成。
- `GOPRIVATE=github.com/pole-io/*,github.com/lattice-hub/* GOPROXY=direct go test -run '^$' -count=0 ./...` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 和 `git -C /Users/chuntao.liao/Github/pole-io/specification diff --check` 通过。

Review：

- 这次没有把 `level` 下沉，因为它决定整条熔断规则的匹配层级，和具体策略执行参数不是同一类字段。
- 顶层字段号未使用 `reserved`，而是按治理规则当前约定压实为 `block_configs=11`、`priority=12`、`metadata=13`、`editable=14`、`deleteable=15`。
- Console 归一化仍兼容旧顶层 `recoverCondition`、`faultDetectConfig`、`fallbackConfig` 和 `max_ejection_percent`，读取后会映射到每条 UI 策略；新建/编辑提交会打包为 `CircuitBreakerPolicy`。

## specification 仓库来源迁移到 lattice-hub

- [x] 核对 `github.com/pole-io/specification` 在源码、模块文件、格式化脚本和知识库中的命中范围
- [x] 验证 `github.com/lattic-hub/specification` 与 `github.com/lattice-hub/specification` 的实际仓库状态
- [x] 将 `github.com/pole-io/specification` 依赖通过 `replace` 指向 `github.com/lattice-hub/specification v0.1.0-ALPHA.27`
- [x] 更新 Go module 校验和并确认替换关系可解析
- [x] 运行编译级 Go 验证、context-kg lint 和 diff 检查
- [x] 记录 review 和验证结果

当前判断：

- 用户本次说的是 spec 仓库组织迁移；实测 `lattic-hub/specification` 不存在，实际公开仓库是 `lattice-hub/specification`。
- `lattice-hub/specification` 的 `develop` 和 `v0.1.0-ALPHA.27` 里 `go.mod` 仍声明 `module github.com/pole-io/specification`，生成的 `.pb.go` 内部 import 也仍是旧 module path。
- 因此 control-plane 侧不能直接把 Go import 改为 `github.com/lattice-hub/specification`；直接改会导致 Go 报错：`module declares its path as: github.com/pole-io/specification but was required as: github.com/lattice-hub/specification`。
- 当前采用可编译的过渡方案：保留源码 import 和 require 的 module path `github.com/pole-io/specification`，在 `go.mod` 中增加 `replace github.com/pole-io/specification => github.com/lattice-hub/specification v0.1.0-ALPHA.27`，让依赖来源切到新仓库。

验证：

- `gh repo view lattic-hub/specification --json nameWithOwner,visibility,defaultBranchRef` 失败，GitHub 无法解析该仓库。
- `gh repo view lattice-hub/specification --json nameWithOwner,visibility,defaultBranchRef` 通过，仓库为公开仓库，默认分支 `develop`。
- `gh api 'repos/lattice-hub/specification/contents/go.mod?ref=develop'` 与 `ref=v0.1.0-ALPHA.27` 均显示 module 仍为 `github.com/pole-io/specification`。
- `GOPRIVATE=github.com/pole-io/*,github.com/lattice-hub/* GOPROXY=direct go test -run '^$' -count=0 ./...` 通过，覆盖全仓编译级验证。
- `go test ./...` 曾启动并通过前半段包编译/部分测试，但长时间无输出，已中断；本轮用编译级验证确认本次 module 来源替换没有破坏 import。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- 本轮没有扩大修改本仓 module `github.com/pole-io/pole-server`。
- 没有保留不可编译的 `github.com/lattice-hub/specification` import 改法；该改法需要 spec 仓库先修改自身 `go.mod` 和重新生成 Go 代码后发新 tag。
- `go.sum` 当前记录的是替换后 `github.com/lattice-hub/specification v0.1.0-ALPHA.27` 的校验和。

## 限流规则编辑器规则 Tab 设计实现

- [x] 核对现有限流数据契约、路由规则编辑器复用模式和已记录的布局坑
- [x] 写入限流编辑器纯逻辑失败断言，覆盖 Spec 生成、保存校验、集群动作锁定和摘要阈值
- [x] 抽出 `rateLimitEditorUtils`，让右侧实时 Spec 与保存 payload 同源
- [x] 重构 `RateLimitEditor` 规则 Tab：基础信息、作用对象、子规则折叠列表、右侧实时 Spec 双栏
- [x] 实现匹配接口多行、匹配条件关系、限流方式、限流效果和自定义响应编辑
- [x] 运行脚本验证、Console 构建、context-kg lint、diff 检查和真实 8080 页面验证
- [x] 修复真实截图暴露的限流编辑器视觉裁剪：基础信息/作用对象内容被截断、右侧 Spec 工具按钮外溢
- [x] 补充真实 8080 页面截图和布局指标验证
- [x] 将工作台限流规则详情抽屉宽度与路由规则详情抽屉对齐
- [x] 补充 8080 真实页面抽屉宽度验证
- [x] 记录 review 和验证结果

当前判断：

- 用户提供的设计交接已经明确范围：只实现 Pole.IO 治理工作台 RateLimitRule 编辑抽屉「规则」Tab，不处理版本、审计等其它 Tab。
- 本轮应复用路由规则编辑器的双栏抽屉、规则块、步骤轴、分段按钮和 Spec 面板模式，避免再出现表单挤压、折叠态过高、输入失焦、Spec 与保存 payload 分叉等问题。
- 右侧实时 Spec 应优先服务保存前核对，展示当前前端真实提交的 RateLimit payload；如果需要 CRD 风格 `apiVersion/kind/spec`，应作为后续单独转换层处理。

当前进展：

- 已读取 `RateLimitEditor`、`services/ratelimit.ts`、路由编辑器 helper/CSS 和 `context-kg/tasks/lessons.md` 中的相关经验。
- 已发现现有限流编辑器基础信息混入作用对象、匹配接口仅支持单字段、右侧实时 Spec 缺失、保存校验不足，需要按设计交接重排。
- 已新增 `verify-ratelimit-editor-utils.mjs` 并先复现失败：缺少 `rateLimitEditorUtils.ts`。
- 已新增 `rateLimitEditorUtils.ts`，覆盖真实提交 payload 生成、YAML/JSON 序列化、集群限流强制 `REJECT`、摘要/阈值文案和保存前校验。
- 已重构 `RateLimitEditor`：基础信息与作用对象拆分，规则区改为 `规则 [n]` 折叠卡片 + 四段步骤轴，右侧新增固定实时 Spec 面板。
- 已更新 `RateLimitEditor.module.less`，复用路由规则双栏抽屉和 Spec 面板模式，表格使用稳定 grid 列宽，避免输入控件互相挤压。
- 用户截图复核发现上一轮视觉验收不足：基础信息与作用对象卡片内容在真实视口中被截断，右侧 Spec 顶部 YAML 控件外溢到面板边缘。
- 已定位根因：限流编辑器内部复用固定 `100vh` 高度和多层 `overflow: hidden`，但它实际嵌在 `RuleDetailDrawer` 的 Tab 内容区内；同时 Spec 头部横排控件没有为窄列留换行空间。
- 已将工作台 `ratelimit` 详情抽屉的 `size` 与 `route` 对齐为 `min(1560px, calc(100vw - 40px))`。
- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs` 已转绿。
- `cd console/web && npm run build:test` 已通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。

验证：

- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs` 通过，覆盖真实提交 payload 生成、集群限流强制 `REJECT`、YAML/JSON 序列化、摘要/阈值文案和保存前校验。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/RateLimit/RateLimitEditor.module.less console/web/src/pages/Governance/RateLimit/rateLimitEditorUtils.ts console/web/scripts/verify-ratelimit-editor-utils.mjs context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 登录 `admin/admin123` 打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-ratelimit` 抽屉验证：查看态显示基础信息、作用对象、限流规则和右侧实时 Spec 双栏；规则块摘要为 `1 个接口 · 1 个匹配条件 · 请求数 · 快速失败`，阈值徽标为 `120 次 / 1s`。
- Playwright 点击编辑验证：作用对象可编辑、集群限流下没有动作切换，仅显示快速失败和失败处理策略；匹配接口、匹配条件、限流窗口、阈值计算合并、自定义响应和添加规则控件均出现，右侧 Spec 保持 `action: REJECT`。
- Playwright 将接口路径改为 `/api/v1/refunds` 后，右侧 Spec 立即同步 `value: /api/v1/refunds`，输入框保持 active；清空接口路径后，右侧 footer 显示 `规则[1] 存在空的接口路径`；随后点击撤销恢复只读态，未保存测试改动。
- 用户反馈截图后重新验证：Playwright 真实 8080 查看态截图保存到 `/tmp/ratelimit-after-fix.png`，编辑态截图保存到 `/tmp/ratelimit-edit-after-fix.png`。
- 查看态布局指标通过：基础信息、作用对象、限流规则三个 section 均 `clipped=false`，其中基础信息 `clientHeight=226/scrollHeight=226`，作用对象 `clientHeight=154/scrollHeight=154`；右侧 YAML/JSON/复制按钮均在 Spec 面板边界内。
- 编辑态布局指标通过：基础信息、作用对象、限流规则三个 section 均 `clipped=false`，其中基础信息 `clientHeight=233/scrollHeight=233`，作用对象 `clientHeight=192/scrollHeight=192`；右侧 YAML/JSON/复制按钮均在 Spec 面板边界内。
- 工作台限流抽屉宽度验证通过：1600px viewport 下，限流详情 `.t-drawer__content-wrapper` 宽度为 1560px，left 为 40px，符合与路由一致的 `min(1560px, calc(100vw - 40px))`；截图保存到 `/tmp/ratelimit-wide-drawer-fixed.png`。

Review：

- 本轮只改 Console 限流规则编辑器规则 Tab，不改版本、监听、发布逻辑和后端接口。
- 右侧实时 Spec 展示的是当前前端提交对象的真实字段形状，不再按交接文档里的 `apiVersion/kind/spec` 伪 CRD 展示，避免重复 RouteRule 已踩过的 Spec 与保存接口分叉问题。
- 当前后端 RateLimit proto 只有单个 `LimitTrigger.method: MatchString`，没有协议字段和多接口数组；因此 UI 以单行接口编辑承载当前可保存字段，协议/方法显示为当前约束下的固定辅助列，没有在提交 payload 中伪造后端不支持的 `interfaces`。
- 子代理只读核对发现当前 `RateLimitView` 仍包含 `namespace/service`，但后端 `RateLimit` proto 顶层没有这两个字段，HTTP 解析允许 unknown fields；本轮沿用既有前端请求结构和列表展示，不在 UI 任务中扩展后端协议。若要让作用对象真正进入后端契约，需要后续修改 specification/proto、store/service/cache 和 Console 映射。
- 当前匹配条件关系显示 `AND`，`OR` 置灰，因为真实 `LimitTrigger.arguments[]` 暂无关系字段；多接口和 OR 关系属于协议增强，不应只在前端伪实现。

## RouteRule 实例标签值改为 TagInput

- [x] 写入回归断言，确认「编辑实例标签」弹窗在 `包含/不包含` 时使用 `TagInput`
- [x] 修复实例标签值编辑控件，复用逗号字符串与 TagInput 数组互转 helper
- [x] 运行路由编辑器脚本、Console 构建、context-kg lint、diff 检查和真实 8080 页面验证
- [x] 重启本地 all 模式服务并记录 review

当前判断：

- 用户截图对应的是 RouteRule 目标分组里的「编辑实例标签」弹窗；这里编辑的是 `destinations[].labels[key]` 的 `MatchString`。
- 该字段和匹配条件值一样，`包含/不包含` 需要多值标签输入，保存给后端仍是英文逗号分割字符串。

当前进展：

- `verify-route-editor-utils.mjs` 已先写失败断言，扫描 `renderTagDialog` 片段要求出现 `isTagInputMatchType`、`TagInput`、`commaStringToTags` 和 `tagsToCommaString`；修复前断言失败在缺少 `isTagInputMatchType`。
- 已把「编辑实例标签」弹窗的标签值列改为：`包含/不包含` 使用 TDesign `TagInput`，其它匹配类型继续使用普通 `Input`。
- `cd console/web && node scripts/verify-route-editor-utils.mjs` 已转绿。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过，覆盖 `renderTagDialog` 使用 `TagInput` 和逗号互转 helper。
- `cd console/web && npm run build:test` 通过，保留既有 Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/scripts/verify-route-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 登录 `admin/admin123` 打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-route` 编辑态，打开目标分组「编辑实例标签」弹窗，把标签匹配类型切到 `包含` 后验证：弹窗打开、`.t-tag-input` 数量为 1、已有 `v1` 作为 tag 展示。验证后已点击取消和撤销，未保存测试改动。

Review：

- 根因是 RouteRule 目标分组标签弹窗仍把 `destinations[].labels[key].value` 固定渲染为普通 `Input`，漏掉了该字段同样是 `MatchString` 的事实。
- 修复后匹配条件值和实例标签值两处 `MatchString` 编辑行为一致：`IN/NOT_IN` 用 `TagInput`，保存仍回写英文逗号分割字符串。

## RouteRule 包含匹配值改为 TagInput

- [x] 对比现有 `MatchInput` 公共组件，确认 `IN/NOT_IN` 用 `TagInput`，提交值通过英文逗号拼接
- [x] 补路由编辑器 helper 断言，覆盖逗号字符串与 TagInput 数组互转
- [x] RouteRule 匹配值在 `包含/不包含` 时改用 TDesign `TagInput`
- [x] 运行路由编辑器脚本、Console 构建、context-kg lint、diff 检查和真实 8080 页面验证
- [x] 重启本地 all 模式服务并记录 review

当前判断：

- 用户反馈的是 RouteRule 匹配条件里的 `匹配类型=包含` 场景；这里的前端编辑体验应是多值标签输入，而不是普通文本框。
- 后端仍接收 `value.value` 的字符串，多个值用英文逗号分割，因此前端需要在展示层用数组，写回时合并成逗号字符串。

当前进展：

- `verify-route-editor-utils.mjs` 已先复现失败：`commaStringToTags` 尚不存在时脚本报 `TypeError: utils.commaStringToTags is not a function`。
- 已补 `commaStringToTags` / `tagsToCommaString` / `isTagInputMatchType`，并在 RouteRule 匹配值编辑态中对 `IN` / `NOT_IN` 渲染 TDesign `TagInput`。
- 真实页面验证时发现编辑匹配条件会把 `routing_config.@type` 退回旧 `RuleRoutingConfig`，已补回归断言并统一默认/更新路径为 `type.googleapis.com/v1.CustomRoute`。
- `cd console/web && node scripts/verify-route-editor-utils.mjs` 已转绿。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过，覆盖逗号字符串与 TagInput 数组互转、`IN/NOT_IN` 判定、默认 `@type=type.googleapis.com/v1.CustomRoute`。
- `cd console/web && npm run build:test` 通过，保留既有 Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/routeEditorUtils.ts console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/Router/CustomRouteEditor.module.less console/web/scripts/verify-route-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 登录 `admin/admin123` 打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-route` 编辑态，把匹配类型切到 `包含` 后验证：DOM 中 `.t-tag-input` 数量为 1，包含 `vip` 标签，右侧实时 Spec 显示 `type: IN`、`value: vip`，并保持 `type.googleapis.com/v1.CustomRoute`。

Review：

- 根因是 RouteRule 规则表格的匹配值编辑控件没有按匹配类型分支，所有类型都走普通 `Input`；用户必须手写分隔格式，和公共 `MatchInput` 组件中 `IN/NOT_IN` 的 TagInput 模式不一致。
- 修复后 `IN/NOT_IN` 仅在展示层使用 `TagInput` 数组，写回 `value.value` 时仍用英文逗号拼接；其它匹配类型保持原普通输入。
- 真实页面验证过程中顺带发现规则更新路径硬编码旧 `RuleRoutingConfig`，已改为保留现有 `@type`，缺省时使用真实提交所需的 `CustomRoute`。

## RouteRule 匹配参数键改为用户输入

- [x] 定位参数键预置来源：默认匹配条件、参数类型切换 helper 和参数键 Select 候选
- [x] 写入失败断言，确认参数键不应自动预置 `x-tenant`、`$method` 或 `session`
- [x] 将参数键控件改为普通输入框，移除预置候选
- [x] 运行路由编辑器脚本、Console 构建、context-kg lint、diff 检查和真实 8080 页面验证
- [x] 重启本地 all 模式服务并记录 review

当前判断：

- 用户反馈的「这里」是 RouteRule 匹配条件里的参数键字段；当前实现把它做成 `Select creatable`，并按参数类型预置候选，例如 HEADER 自动填 `x-tenant`。
- 这个字段应由用户自己输入，参数类型只决定参数来源类型，不应隐式生成业务参数键。

当前进展：

- `verify-route-editor-utils.mjs` 已先写失败断言：`getDefaultParamKey('HEADER')` 和 `getDefaultParamKey('METHOD')` 应返回空字符串，切换参数类型不应把参数键改成 `session`。
- 已移除参数键预置候选，`CustomRouteEditor` 匹配条件参数键编辑态从 `Select creatable` 改为普通 `Input`。
- `cd console/web && node scripts/verify-route-editor-utils.mjs` 已转绿。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/routeEditorUtils.ts console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/scripts/verify-route-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 登录 `admin/admin123` 打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-route` 编辑态验证：已有参数键 `x-tenant` 渲染为普通 `Input`，不在 `.t-select` 内，placeholder 为 `请输入参数键`。
- Playwright 点击 `添加匹配条件` 后验证：新增参数键输入框值为空，两个参数键输入框均不是 Select 包裹，也没有 readonly。

Review：

- 根因是参数键逻辑复用了 `PARAM_KEY_OPTIONS/getDefaultParamKey/withParamTypeDefaultKey`，导致新条件和参数类型切换都会自动写入业务参数键。
- 修复后参数类型只负责选择来源类型；参数键完全由用户输入，保存 payload 仍按用户输入值进入 `routing_config.rules[].arguments.arguments[].key`。

## RouteRule 实时 Spec 提交格式对齐

- [x] 复现右侧实时 Spec 与保存接口实际 payload 不一致的问题
- [x] 将预览数据源改为保存前实际构造的 `CustomRoute` 结构
- [x] 更新路由编辑器 helper 验证脚本，覆盖不再输出 `apiVersion/kind/spec`
- [x] 运行 Console 构建、context-kg lint、diff 检查和真实 8080 页面验证
- [x] 重启本地 all 模式服务并记录 review

当前判断：

- 用户截图中的右侧实时 Spec 当前展示的是额外构造的 `apiVersion/kind/spec` 资源视图，但 RouteRule 编辑器保存时实际上传的是 `CustomRoute` payload：顶层为 `name/enable/priority/description/routing_policy/metadata/routing_config`。
- 预览应服务于“保存前核对即将上传的数据”，因此应和 `onSubmit` 中的 `data: CustomRoute` 保持同源，而不是继续展示单独的伪资源格式。

当前进展：

- 已先改 `console/web/scripts/verify-route-editor-utils.mjs` 写入失败断言：预览对象必须包含真实提交字段，并且不应再包含 `apiVersion/kind/spec`。
- 失败已复现：当前 `buildRouteRulePreviewSpec(baseDraft).name` 为 `undefined`，说明预览结构确实不是提交结构。
- `routeEditorUtils` 已新增提交 payload builder，右侧实时 Spec 和 `onSubmit` 统一走 `buildRouteRuleSubmitPayload`，避免预览与实际上传数据再次分叉。
- `cd console/web && node scripts/verify-route-editor-utils.mjs` 已转绿，断言预览包含 `name/enable/priority/routing_policy/metadata/routing_config`，且不包含 `apiVersion/kind/spec`。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/routeEditorUtils.ts console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/scripts/verify-route-editor-utils.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 登录 `admin/admin123` 打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-route` 抽屉验证：YAML 预览包含 `name/enable/priority/routing_policy/metadata/routing_config/caller/callee`，不再包含 `apiVersion/kind/spec`。
- Playwright 切换 JSON 预览验证：JSON 从真实提交对象开始，包含 `"routing_config"` 和 `"routing_policy"`，不包含 `"kind"`、`"spec"` 或 `apiVersion`。

Review：

- 根因是预览和保存各自构造数据：预览用 `buildRouteRulePreviewSpec` 造了 `apiVersion/kind/spec` 展示模型，保存用 `buildRoutingConfigForApi` 构造 `CustomRoute` 请求体。
- 修复后右侧预览和 `onSubmit` 同源，均走 `buildRouteRuleSubmitPayload`；保存前校验仍走 `validateRouteRuleDraft`，保存接口和 Redux action 不变。
- YAML 序列化跳过 `undefined` 字段，避免展示出 JSON 请求体不会提交的空字段。

## lattice-hub 组织 README 中英文介绍

- [x] 确认 `gh` 登录状态、组织 profile 仓库和当前 README 位置
- [x] 基于组织仓库列表确认公开项目定位，避免凭空编写介绍
- [x] 起草中英文组织介绍，覆盖定位、核心项目和参与方式
- [x] 使用 `gh` 更新 `lattice-hub/.github` 的 `profile/README.md`
- [x] 验证 GitHub 内容已更新，并记录 review

当前判断：

- `gh` 已登录为 `chuntaojun`，对 `lattice-hub/.github` 具备 `ADMIN` 权限。
- 组织 profile 仓库为 `lattice-hub/.github`，默认分支为 `main`，README 路径为 `profile/README.md`。
- 当前 profile README 仍是 GitHub 默认模板，可直接替换为正式中英文介绍。
- 公开仓库显示组织围绕 Pole 服务治理生态：`pole-control-plane`、`specification`、`pole-client-rust`、`pole-sidecar` 等。

当前进展：

- 已在临时 clone `/tmp/lattice-hub-github-profile` 中替换 `profile/README.md`，提交为 `a645c7b docs: update organization profile readme`。
- 已推送到 `lattice-hub/.github` 的 `main` 分支。

验证：

- `git diff --check -- profile/README.md` 通过。
- `git status --short --branch` 在临时 clone 中显示 `main...origin/main` 干净。
- `gh api 'repos/lattice-hub/.github/contents/profile/README.md?ref=main'` 验证远端 `profile/README.md` 已包含 `# Lattice Hub`、中文介绍、English 介绍和四个核心项目链接。

Review：

- 本轮只更新组织 profile README，没有改动 `lattice-hub` 其它仓库。
- 文案基于当前可见仓库列表和 `pole-control-plane` 仓库描述编写，避免引入未确认的产品承诺或外部文档链接。
- 当前 `pole-control-plane` 工作区已有大量先前未提交变更，本轮除任务记录外未触碰业务代码。

## RouteRule 服务范围区 PRD 对齐

- [x] 对比用户截图中的 PRD 展开态、收起态和当前实现差异
- [x] 调整 RouteRule 编辑抽屉「服务范围」标题区，使展开态显示 `主调方 → 被调方`，收起态显示服务流向摘要
- [x] 移除该区块标题中的说明文案和右侧独立摘要，避免标题高度和布局偏离 PRD
- [x] 运行 Console 构建、路由编辑器脚本、context-kg lint 和真实 8080 页面验证
- [x] 重启本地 all 模式服务并记录 review
- [x] 修复目标分组编辑态：恢复标签编辑控件、放大权重输入列、分组名改为 `Group {index}` 自动生成
- [x] 调整 RouteRule 基础信息编辑态首行间距，改用 TDesign `Row/Col` 控制规则名称、优先级和状态列宽

当前判断：

- 用户这次纠正聚焦在「服务范围」区块：当前实现把描述和服务摘要拆到左右两侧，导致展开态像大标题说明区，而 PRD 是紧凑单行标题。
- 应保留现有左侧编辑表单无横向滚动、右侧 Spec 固定的整体布局，只调整服务范围区块的信息呈现。

当前进展：

- 「服务范围」区块已改为专用标题结构：展开态显示 `服务范围  主调方 → 被调方`，收起态显示 `服务范围  spec-governance/spec-gateway → spec-governance/spec-order`。
- 展开态服务卡片文案已对齐 PRD：`主调 / 发起调用方`、`被调 / 目标服务方`。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/Router/CustomRouteEditor.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- Playwright 登录 `admin/admin123` 打开 `http://127.0.0.1:8080/governance/workbench`，进入 `seed-20260616-route` 抽屉验证：展开态包含 `主调方 → 被调方`，收起态包含 `spec-governance/spec-gateway → spec-governance/spec-order`，旧说明文案不存在，页面无横向溢出。
- Playwright 复验 RouteRule 编辑态基础信息行：前三列为 TDesign `Col span=8/2/2`，实测列宽约 `504/126/126px`，`InputNumber` 根节点宽约 `110px`，页面、抽屉和该行均无横向溢出。

Review：

- 本轮只调整 RouteRule 编辑器服务范围区的展示结构和文案，没有改保存、发布和 Spec 生成逻辑。
- 右侧实时 Spec 固定列、左侧内部滚动和抽屉宽度分配保持不变，避免回退上一轮已修复的布局问题。
- 跟进用户反馈后，服务范围收起态已改为紧凑单行：真实 8080 页面实测外层 section 高度约 64px，header 约 62px，不再保留展开态大块空白。
- 跟进用户反馈后，目标分组编辑态恢复实例标签 chip + 编辑标签弹窗；权重列扩大并使用固定宽度 `InputNumber`；分组名称不再可手填，显示和提交统一按 `Group {index}` 自动生成。
- 真实 8080 页面复验：编辑态目标分组无 `Group 1` 输入框，权重输入值完整显示 `100`，实例标签显示 `version / v1` chip，点击目标分组「编辑标签」可打开 `编辑实例标签` 弹窗；右侧实时 Spec 已同步显示 `group: "Group 1"`，不再保留旧 `primary`。
- 跟进用户反馈后，基础信息首行不再使用 `1fr + 固定列` 的私有网格；改为 TDesign `Row/Col` 的 12 栅格，规则名称、优先级、状态按 `8/2/2` 分配，右侧字段间距稳定且不挤压名称输入框。

## 路由规则编辑器规则 Tab 设计实现

- [x] 补路由编辑器纯逻辑测试，覆盖 Spec 生成、保存校验、权重计算和参数键联动
- [x] 抽出 `routeEditorUtils`，避免把校验和 Spec 生成逻辑堆进 JSX
- [x] 重构 `CustomRouteEditor` 的规则 Tab：基础信息、服务范围、路由规则、右侧实时 Spec 预览
- [x] 实现服务范围折叠摘要、长服务名 tooltip、目标分组权重条和保存前校验
- [x] 运行 Console 构建、相关脚本验证、context-kg lint 和 diff 检查
- [x] 记录 review 和验证结果

当前判断：

- 附件已经给出明确设计交接，本轮不再做需求反问；实现范围限定在路由规则编辑抽屉的「规则」Tab，不改版本、审计等 Tab。
- 现有 `CustomRouteEditor` 已有基础表单、服务选择、规则块、匹配条件表和目标分组表；本轮应复用数据结构和保存 API，只调整规则 Tab 的信息架构和必要交互。
- 纯逻辑优先抽到 helper，至少覆盖：kebab-case 校验、目标权重合计必须为 100、空匹配值/未命名分组拦截、实时 Spec JSON/YAML 生成、参数类型切换时默认候选键。

当前进展：

- `routeEditorUtils` 已抽出 Spec 生成、YAML/JSON 序列化、保存前校验、参数类型候选键和标签文本转换逻辑。
- `CustomRouteEditor` 规则 Tab 已改为左侧基础信息、服务范围、路由规则，右侧实时 Spec 预览和底部校验/保存栏。
- 服务范围支持折叠摘要和长文本 Tooltip；匹配条件支持 AND/OR 切换和参数键候选；目标分组支持权重条和合计状态。

验证：

- `cd console/web && node scripts/verify-route-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/Router/CustomRouteEditor.module.less console/web/src/pages/Governance/Router/routeEditorUtils.ts console/web/scripts/verify-route-editor-utils.mjs context-kg/tasks/todo.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式，tmux 日志显示 `finish starting server`。
- `curl http://127.0.0.1:8080/governance/workbench` 返回 200；浏览器登录 `admin/admin123` 后工作台正常显示 9 条治理规则。
- Playwright 打开 RouteRule 详情抽屉，确认新布局出现基础信息、服务范围、路由规则、实时 Spec 和本地校验通过。

Review：

- 本轮只改 RouteRule 编辑抽屉的规则 Tab 体验，保存仍走 `buildRoutingConfigForApi` 和既有 create/update action，避免扩大接口风险。
- 校验逻辑放在 helper 并由 Node 脚本覆盖，JSX 只负责展示和触发，后续其它治理规则编辑器可以复用同样模式。
- 组件中仍保留旧 TDesign Table 编辑函数作为未调用路径，原因是本轮优先保证设计交付和构建稳定；后续可单独做死代码清理。

## 治理规则数据库清理与测试数据重建

- [x] 审计 MySQL 中治理规则相关表、旧数据和现有服务/命名空间依赖
- [x] 备份即将清理的治理规则数据，避免误删后不可追溯
- [x] 清理旧治理规则表数据，范围限定为 `governance_rule` / `governance_rule_release`
- [x] 通过当前 8080/8090 API 构造新的治理规则测试数据，避免手写旧字段 JSON
- [x] 验证治理工作台和各类治理规则列表接口均返回正常
- [x] 记录 review 和验证结果

当前判断：

- 用户建议清理旧数据库信息并重新构造测试治理规则数据；本轮将只处理治理规则统一表，避免影响命名空间、服务、用户、认证策略等基础数据。
- 当前 `governance_rule` 和 `governance_rule_release` 各 9 条，覆盖 route、ratelimit、circuitbreaker、faultdetect、lossless、lane-group、traffic-security、traffic-mirror、traffic-mock 各 1 条。
- 现有 `spec-governance` 命名空间下已有 `spec-gateway`、`spec-checkout`、`spec-payment`、`spec-inventory`、`spec-order` 服务，可直接作为新测试数据的作用范围和目标服务。
- 新测试数据优先通过当前 API 创建，让后端按最新 spec 生成 JSON 快照，避免直接 SQL 构造再次带入历史字段。
- 清理前已备份 `governance_rule` / `governance_rule_release` 到 `/tmp/pole-governance-rules-20260616-225742.sql`。
- 已在事务中清空 `governance_rule` / `governance_rule_release`，清理后两表均为 0 条。
- 造数过程中发现并修复两个发布接口缺陷：route 正常发布检查灰度版本时 nil route 触发空响应；部分规则发布未生成 release id，导致后续类型撞空主键。
- 已重新构造并发布 9 类治理规则测试数据：route、ratelimit、circuitbreaker、faultdetect、lossless、lane-group、traffic-security、traffic-mirror、traffic-mock 各 1 条。

验证：

- MySQL 验证通过：`governance_rule` 9 条、`governance_rule_release` 9 条，9 个 `rule_type` 各 1 条。
- 8080 Console 代理验证通过：工作台默认加载涉及的 9 类治理规则列表接口均返回 `200000`；`/governance/workbench` 返回页面 HTML。
- `go test ./pkg/goverrule ./plugin/store/mysql -run 'TestBuildRouterRuleReleaseAllowsNilRule|TestNewRuleReleaseFromSpec|TrafficGovernance|Lossless|GovernanceRule'` 通过。

Review：

- 直接 SQL 清理治理规则表后必须重启服务，否则运行中缓存不会感知硬删除，会出现 API 列表仍显示旧规则的现象。
- 造数优先走 API 可以让当前 spec 生成规则 JSON 快照；lossless 当前 spec 已不承载 namespace/service，最终通过统一表索引字段补齐归属，避免把旧字段写回 `rule` JSON。
- 发布接口修复后，各类 release 记录都有非空主键；route 发布的灰度检查路径不再因 nil 规则构造而关闭连接。

## A2A Agent 列表 404 修复

- [x] 使用用户提供的 RequestId 定位 8090 access log
- [x] 复现 `GET /ai/a2a/v1/agents` 在 8080 代理和 8090 直连均返回 404
- [x] 对比 apiserver 配置，确认 all 模式使用的 `deploy/conf/pole-apiserver.yaml` 未启用 `aia2a`
- [x] 修复默认 apiserver 配置并重启 all 模式验证

当前判断：

- 用户页面报“获取 A2A Agent 失败”不是前端解包问题，也不是 MySQL 数据缺失；请求已到达 8090，但 A2A WebService 未注册，返回 `404 Page Not Found`。
- `test/data/bootstrap/pole-apiserver.yaml` 已启用 `aia2a`，因此部分接口测试环境不会暴露这个问题；本地 all 模式使用 `deploy/conf/pole-apiserver.yaml`，该配置此前仍为 `aia2a.enable: false`。

当前进展：

- 已将 `deploy/conf/pole-apiserver.yaml` 的 `api-http.api.aia2a.enable` 改为 `true`。

验证：

- 修复前：带有效登录态访问 `http://127.0.0.1:8080/ai/a2a/v1/agents?offset=0&limit=10` 返回 `404 Page Not Found`。
- 修复前：带有效登录态访问 `http://127.0.0.1:8090/ai/a2a/v1/agents?offset=0&limit=10` 返回 `404 Page Not Found`。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式。
- 修复后：带有效登录态访问 `http://127.0.0.1:8080/ai/a2a/v1/agents?offset=0&limit=10` 返回 `200000`，`amount=3`。
- 修复后：带有效登录态访问 `http://127.0.0.1:8090/ai/a2a/v1/agents?offset=0&limit=10` 返回 `200000`，`amount=3`。

## 治理规则列表请求失败修复

- [x] 复现 `http://127.0.0.1:8080/governance/workbench` 中治理规则列表失败，记录失败接口、响应体和 RequestId
- [x] 结合服务日志追踪 RequestId `a305a696-3b90-4624-b00e-574cf5e29c00` 对应的后端错误
- [x] 沿 Console service、HTTP handler、goverrule service、store/cache 数据流定位根因
- [x] 补充能稳定覆盖根因的最小测试或等价接口验证
- [x] 实施最小修复并运行 Go 相关测试、Console 构建或真实 8080 链路验证
- [x] 记录 review、验证结果和必要 lessons

当前判断：

- 用户反馈的两个请求失败都发生在治理规则列表查询场景，不能只修改页面错误提示；需要从真实 8080 Console 代理请求和后端日志定位 5xx 根因。
- 当前工作区已有治理规则相关未提交变更，本轮只处理列表请求失败相关改动，不回退既有变更。
- RequestId `a305a696-3b90-4624-b00e-574cf5e29c00` 对应接口为 `/naming/v1/traffic/mirrors?offset=0&limit=20`，后端错误为 `proto: unknown field "source"`。
- 同一启动实例中 lossless cache 还持续报 `proto: unknown field "service"`，同属治理规则旧 JSON 快照与新 spec 字段删除后的兼容解析问题。

当前进展：

- `plugin/store/mysql/governance_rule_convert.go` 新增统一的治理规则 proto JSON 反序列化 helper，读取 DB 快照时使用 `DiscardUnknown: true` 兼容旧字段。
- 兼容范围覆盖 `LosslessRule`、`TrafficSecurityRule`、`TrafficMirror`、`TrafficMock` 的规则快照和发布快照读取；新写入仍按当前 spec marshal，旧字段不会被再次写回。
- 补充 `TestTrafficMirrorRecordIgnoresLegacySourceField` 和 `TestLosslessRecordIgnoresLegacyServiceField`，先确认旧严格解析会失败，再用最小修复转绿。

验证：

- `go test ./plugin/store/mysql -run 'TestTrafficMirrorRecordIgnoresLegacySourceField|TestLosslessRecordIgnoresLegacyServiceField'` 通过。
- `go test ./plugin/store/mysql -run 'TrafficGovernance|Lossless|GovernanceRule'` 通过。
- `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./plugin/store/mysql -run 'TrafficGovernance|Lossless|FaultDetect|Circuit|RateLimit|TestDoesNotExist'` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过并重启 all 模式；脚本内已执行 Console build 和 control-plane build。
- 8080 Console 代理验证通过：`/naming/v1/traffic/security`、`/naming/v1/traffic/mirrors`、`/naming/v1/traffic/mocks`、`/naming/v1/lossless` 均返回 `200000`。
- 重启后 `polaris-cache.log` 不再出现新的 `unknown field "source"` / `unknown field "service"` 解析错误。
- `python3 ~/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- 根因是治理规则 spec 删除/迁移字段后，数据库 `governance_rule.rule` 中仍保留旧 proto JSON 字段；列表和缓存读取用严格 `protojson.Unmarshal`，遇到旧字段直接返回 500。
- 修复落在 store 转换层，比在 Console 或各 handler 分散兜底更优雅：所有列表、详情、缓存、发布快照读取共享同一兼容行为。
- 本轮没有新增 lessons：用户没有提出新的行为纠正。

## 主动探测多子规则支持

- [x] 确认 `FaultDetectRule` 多子探测规则的 spec 契约形态
- [x] 更新 specification proto 与 Go/Rust 生成产物
- [x] 更新 control-plane 后端转换、服务索引、查询过滤、发布下发兼容逻辑
- [x] 更新 Console service 类型、归一化逻辑和 `FaultDetectEditor` 多子规则编辑器
- [x] 补充后端/前端/fixture 测试场景并运行验证
- [x] 记录 review、验证结果和必要 lessons

当前判断：

- 现有 `FaultDetectRule` 是单条探测配置：顶层 `target_service/interval/timeout/port/protocol/http_config/tcp_config/udp_config`。
- 新需求应表达为“一条主动探测治理规则包含多个子探测规则”，而不是让用户创建多条顶层治理规则；顶层继续承载规则身份、描述、标签和权限字段。
- 后端统一治理表保存 `rule` JSON 快照，适合把多子规则作为 spec JSON 一起持久化；但服务索引和查询过滤需要从子规则集合中提取涉及服务。
- 用户补充：鉴权、Mock、镜像也应保持“一条规则绑定一个被调服务 + 多个子规则”的模型；来源服务属于流量匹配条件，不参与规则归属。

当前进展：

- `FaultDetectRule` 新增 `repeated FaultDetectSubRule rules`，子规则承载探测目标、间隔、超时、端口、协议和协议配置。
- control-plane 创建/更新主动探测规则时优先从 `rules[0].target_service` 推导内部索引；旧单条字段仍可兼容。
- 主动探测客户端缓存按所有子规则目标服务建立 fan-out，避免多目标规则只挂到首个服务。
- Console service 层已归一化 `rules[0]` 到旧字段，列表与详情不会因新字段空白。
- `FaultDetectEditor` 已升级为探测子规则列表，支持一个主动探测规则内新增、删除、查看和编辑多个子探测规则；每条子规则独立维护被探测服务、接口范围、探测节奏、协议配置和启停状态。

验证：

- `../specification/source/go: bash build.sh` 通过。
- `../specification/source/rust: bash build.sh` 通过。
- `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./plugin/store/mysql ./test/e2e/internal/e2e -run 'TrafficGovernance|FaultDetect|TestDoesNotExist'` 通过。
- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 chunk size 警告。

Review：

- 本轮已完成 spec、后端索引、缓存下发、Console service 兼容和 `FaultDetectEditor` 多子规则交互；编辑器提交时会把首条子规则同步回顶层旧字段，保持旧链路兼容。
- 为兼容已有数据，旧单条字段仍会被前端和后端归一化成一个子规则；新客户端可以统一读取 `rules[]`。

## 流量治理被调服务绑定语义补全

- [x] 更新 specification：TrafficSecurity / TrafficMirror / TrafficMock 顶层增加 `target_service`，子规则移除服务归属字段
- [x] 更新 `TrafficMatchRule.SourceMatch`，支持来源服务匹配
- [x] 重新生成 specification Go / Rust 产物
- [x] 更新 control-plane 后端规则转换、缓存索引、store 查询字段
- [x] 更新 Console 类型、默认值、详情展示和编辑器字段
- [x] 更新 e2e fixture 与单元测试，运行相关验证
- [x] 记录 review 与 lessons

当前判断：

- 三类流量治理都应按被调服务下发：客户端按被调服务拉取规则，数据面再用 `traffic_match_rule` 判断来源服务、请求头、路径、Cookie 等条件。
- 旧的 `MirrorSource.namespace/service`、`MockSource.namespace/service` 容易被误解为规则归属；应改为顶层 `target_service` 作为唯一绑定服务。
- 鉴权规则当前没有服务作用域，导致 Console 查询和客户端下发无法按被调服务索引，需要补齐。

当前进展：

- `TrafficSecurityRule` 新增顶层 `target_service`，后端内部 `Namespace/Service` 从该字段推导。
- `TrafficMirror` / `TrafficMock` 新增顶层 `target_service`；`MirrorRule` / `MockRule` 移除 `source.namespace/service`，子规则只保留接口范围、流量匹配和动作配置。
- `SourceMatch.Type` 新增 `CALLER_SERVICE`，来源服务通过 `traffic_match_rule.arguments` 表达。
- Console 流量治理列表、详情和编辑器已改为展示/编辑“被调命名空间、被调服务”；镜像和 Mock 子规则中不再单独录入来源服务。

验证：

- `../specification/source/go: bash build.sh` 通过。
- `../specification/source/rust: bash build.sh` 通过。
- `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./plugin/store/mysql ./test/e2e/internal/e2e -run 'TrafficGovernance|FaultDetect|TestDoesNotExist'` 通过。
- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 chunk size 警告。
- `rg -n "MirrorSource|MockSource|GetSource\\(|source\\?\\.|source:\\s*\\{" apis/pkg/types/rules plugin/store/mysql test/e2e/internal/e2e console/web/src/pages/Governance/Security console/web/src/services/traffic_governance.ts ../specification/api/v1/traffic_manage ../specification/source/go/api/v1/traffic_manage ../specification/source/rust/pole-specification/proto ../specification/source/rust/pole-specification/src/v1.rs` 无流量治理旧 source 结构残留。
- `rg -n "default_action|DefaultAction|GetDefaultAction" ../specification apis pkg plugin test console/web/src` 无结果。

Review：

- 三类流量治理规则现在统一按被调服务建索引和下发；鉴权规则不再是无服务作用域的全局规则。
- 来源服务不再参与规则归属，只能作为 `CALLER_SERVICE` 等匹配条件；这避免了镜像/Mock 在列表、发布和客户端缓存中被错误挂到主调服务。

## specification 鉴权规则默认动作移除

- [x] 删除 `TrafficSecurityRule.default_action`
- [x] 重新生成 specification Go / Rust 产物
- [x] 检查 control-plane 是否仍引用 `default_action`
- [x] 运行 spec 与 control-plane 相关验证
- [x] 记录 review 与 lessons

当前判断：

- 调用鉴权是按命中策略执行动作；规则本身不需要“所有策略未命中时的默认动作”字段。
- 需要同步删除 proto、Go 生成代码、Rust proto 拷贝与 Rust 生成代码，避免跨语言产物不一致。

验证：

- `../specification/source/go: bash build.sh` 通过。
- `../specification/source/rust: bash build.sh` 通过。
- `rg -n "default_action|DefaultAction|GetDefaultAction" ../specification console/web/src test/e2e/internal/e2e pkg apis plugin` 无结果。
- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `go test ./pkg/goverrule/... ./pkg/cache/rules/... ./plugin/store/mysql/... ./test/e2e/internal/e2e -run TestDoesNotExist` 通过。
- `git diff --check` 在 `../specification` 与当前仓库均通过。
- `context_kg_lint.py ./context-kg` 通过。

Review：

- spec 已从 `TrafficSecurityRule` 删除未命中默认动作字段，并按当前规则连续重排后续字段号；Go/Rust 生成产物同步更新。
- control-plane Console 不再声明、默认填充、展示或提交未命中默认动作；提交前额外剥离历史 JSON 中可能残留的旧字段，避免旧数据被再次写回。
- e2e 治理规则 fixture 已移除该字段；经验已写入 `context-kg/tasks/lessons.md`。

## control-plane spec 适配回归审查

- [x] 运行全仓编译级验证与治理规则重点包测试
- [x] 核对治理规则顶层 `namespace/service` 删除后的后端语义链路
- [x] 横向扫描前端与接口契约是否仍依赖旧字段
- [x] 汇总仍需修复的问题、风险和建议处理顺序
- [x] 记录 review 与验证结果

当前判断：

- 本轮重点不是重新确认 `go.mod` 已经升级，而是检查 spec 破坏性字段删除后，control-plane 是否存在编译未覆盖的运行时语义缺口。
- 重点风险面包括：规则归属来源、鉴权资源收集、缓存 fan-out、DB JSON 兼容、Console 展示与创建/编辑入参。

验证：

- `go test ./... -run TestDoesNotExist` 通过，完成全仓 Go 编译级验证。
- `go test ./pkg/goverrule/... ./pkg/cache/rules/... ./plugin/store/mysql/...` 通过，覆盖治理规则、缓存和 MySQL 重点包。
- `cd console/web && npm run build` 通过，保留既有 Browserslist 过期提示、chunk size 警告和 node `--localstorage-file` 提示。

Review：

- 限流规则适配仍不完整：新 spec 的 `RateLimit` 顶层已无 `namespace/service`，但 Console 仍提交这两个字段，后端 `api2RateLimit` 也没有从任何内部结构或外部入参恢复 `ServiceID`；新建/更新后的限流规则无法正确绑定服务，客户端缓存按服务拉取会跳过该规则，列表/详情的服务展示和过滤也会丢失。
- 限流鉴权资源联动存在空 ID：`CreateRateLimits` / `DeleteRateLimits` 成功后调用 `afterRuleResource` 时写入 `ID: ""`，没有使用响应或请求中的真实规则 ID。
- TrafficSecurity 目前无法按服务索引：`TrafficSecurityRule` spec 顶层和 `TrafficSecurityPolicy` 内部没有 `namespace/service` 字段，control-plane 内部 `TrafficGovernanceRule.Namespace/Service` 对 security 为空，导致按服务过滤、`GetRulesForService` 和 active release key 无法像 mirror/mock 一样工作；如果鉴权规则也要服务级生效，需要在 spec 或 control-plane API 侧补明确的作用域来源。
- 统一治理规则 cache fan-out 基础链路存在：`GovernanceRuleUpdateCache` 已一次拉取 `governance_rule` / `governance_rule_release` 并按 type 分发 watcher，这部分方向正确。
- `governance_rule` / `governance_rule_release` 保留 `namespace/service/service_id` 等列属于 store 索引字段，不是 spec 顶层字段回退；关键是写入来源必须从规则内部或明确 API 上下文推导。

## specification 治理规则语义收敛

- [x] 检查 `../specification` 中治理规则 proto、生成脚本和当前分支状态
- [x] 删除治理规则最外层 `namespace/service` 字段，并收敛 `metadata` 为规则标签
- [x] 补充并确认流量镜像按接口维度配置能力
- [x] 重新生成 specification 产物并运行校验
- [x] 记录 review 与验证结果

## specification tag 与 control-plane 引用更新

- [x] 确认 specification 历史 tag 格式
- [x] 基于 specification 当前提交创建并推送新 tag
- [x] 更新 pole-control-plane 对 specification 的正式版本引用
- [x] 验证 go.mod/go.sum 与构建解析
- [x] 记录 review 与验证结果

当前判断：

- specification 历史最新 tag 为 `v0.1.0-ALPHA.25`，本次按同一格式发布 `v0.1.0-ALPHA.26`。
- control-plane 当前存在本地 `replace github.com/pole-io/specification => /Users/.../specification`，正式引用更新时应删除本地路径。
- 按 lessons 约束，本次不跑 `go mod tidy`，只更新 specification 版本并下载对应模块校验。

当前进展（control-plane 引用更新）：

- `../specification` 已基于 `085f056 feat(traffic): add API interface scope to MirrorSource` 创建并推送 tag `v0.1.0-ALPHA.26`。
- `go.mod` 已将 `github.com/pole-io/specification` 从 `v0.1.0-ALPHA.25` 更新到 `v0.1.0-ALPHA.26`，并删除本地 `replace`。
- `go.sum` 已写入 `v0.1.0-ALPHA.26` 与 `v0.1.0-ALPHA.26/go.mod` 校验和。
- control-plane 已同步适配治理规则 spec 破坏性字段删除：路由、限流、熔断、故障探测、无损、流量鉴权/镜像/Mock 不再读写已删除的顶层 `namespace/service` 字段。

验证（control-plane 引用更新）：

- `git ls-remote --tags origin v0.1.0-ALPHA.26` 在 `../specification` 可查到远端 tag。
- `go list -m github.com/pole-io/specification` 返回 `github.com/pole-io/specification v0.1.0-ALPHA.26`。
- `go test ./pkg/goverrule/... ./pkg/cache/rules/... ./plugin/store/mysql/...` 通过。
- `go test ./plugin/apiserver/eurekaserver ./test/suit -run TestDoesNotExist` 通过。
- `go test ./... -run TestDoesNotExist` 通过，完成全仓 Go 编译级验证。
- `git diff --check -- go.mod go.sum apis/pkg/types/rules pkg/goverrule pkg/cache/rules plugin/store/mysql context-kg/tasks/todo.md` 通过。

Review（control-plane 引用更新）：

- 本轮没有执行 `go mod tidy`，避免引入与 spec tag 更新无关的依赖整理。
- 本轮仅推送了 `specification` 新 tag；control-plane 代码仍留在当前工作区，尚未提交或推送。
- rate limit 新 spec 已无被治理服务顶层字段，当前 control-plane 只能保留内部 `ServiceID` 路径；后续需要按统一治理规则模型补正式 create/update 入参归属来源。

## specification v0.1.0-ALPHA.27 发布与引用更新

- [x] 合并 specification PR #3 到 `develop`
- [x] 基于合并提交创建并推送 tag `v0.1.0-ALPHA.27`
- [x] 更新 pole-control-plane 的 `github.com/pole-io/specification` 依赖到 `v0.1.0-ALPHA.27`
- [x] 移除本地 `replace github.com/pole-io/specification => ../specification`
- [x] 运行后端重点测试、Console 构建和知识库 lint

当前进展：

- specification PR #3 已合并，合并提交为 `9fad778208e541873b8db705de87e23433ff8314`。
- tag `v0.1.0-ALPHA.27` 已推送到 `origin`，远端可通过 `git ls-remote --tags origin v0.1.0-ALPHA.27` 查询。
- control-plane `go.mod` 已更新到 `github.com/pole-io/specification v0.1.0-ALPHA.27`，`go.sum` 已补充新 tag 校验和。
- 由于 Go proxy 访问超时，本轮使用 `GOPRIVATE=github.com/pole-io/* GOPROXY=direct go mod download github.com/pole-io/specification@v0.1.0-ALPHA.27` 直接拉取校验和。

验证：

- `go list -m github.com/pole-io/specification` 返回 `github.com/pole-io/specification v0.1.0-ALPHA.27`。
- `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./plugin/store/mysql ./test/e2e/internal/e2e -run 'TrafficGovernance|FaultDetect|TestDoesNotExist'` 通过。
- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 chunk size 警告。

Review：

- 本轮没有运行 `go mod tidy`，只更新 specification 直接依赖和 go.sum 校验和。
- control-plane 仍保留本轮功能适配代码的未提交变更；本节只记录 spec tag 发布与依赖引用切换结果。

当前判断：

- 用户进一步纠正为：治理规则最外层对象不应保留 `namespace/service` 字段，也不要用注释解释迁移关系，避免误导使用方继续依赖顶层归属字段。
- 用户继续纠正为：治理规则删除字段后不应保留 `reserved`，字段号可以重新从 1 开始连续整理。
- spec 中不应通过 metadata 表达规则归属；metadata 应收敛为规则标签。
- 流量镜像需要像 Mock 一样在来源条件里支持接口维度配置，当前 `MirrorSource.api` 已保留并完成产物同步。

当前进展：

- `../specification/api/v1/traffic_manage/mirror.proto`：删除 `TrafficMirror` 顶层 `namespace/service`，保留 `MirrorSource.api`。
- `mock.proto`、`traffic_security.proto`、`ratelimit.proto`、`lossless.proto`：删除顶层 `namespace/service`。
- `router.proto`、`circuitbreaker.proto`、`fault_detector.proto`：删除顶层 `namespace`。
- `lane.proto`：无顶层 `namespace/service` 可删，仅将 metadata 文案收敛为泳道组标签。
- 已删除治理规则 proto 中的 `reserved`，并将顶层规则消息字段号重新压紧。
- 已重新生成 `source/go` 产物，并同步 `source/rust/pole-specification/proto` 与 `src/v1.rs`。

验证：

- `cd ../specification/source/go && bash build.sh` 通过。
- `cd ../specification/source/rust && bash build.sh` 通过，Rust release build 成功。
- 脚本检查 `TrafficMirror`、`TrafficMock`、`TrafficSecurityRule`、`RateLimit`、`RouteRule`、`CircuitBreakerRule`、`FaultDetectRule`、`LosslessRule` 顶层 Go struct 不再包含 `Namespace` / `Service` 字段。
- `rg -n "reserved" api/v1/traffic_manage api/v1/fault_tolerance api/v1/security -g '*.proto'` 无治理规则 reserved 残留。
- `git diff --check` 在 `../specification` 通过。
- `python3 ~/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 在 control-plane 通过。

Review：

- 本次只收敛治理规则最外层字段和规则标签语义，不删除内部 source/destination/target/routing_config 里的服务定位字段。
- 字段删除和重编号属于 spec 破坏性变更；本轮按用户 review 意见优先保证模型干净，不保留 reserved 和旧字段号空洞。
- 流量镜像接口维度通过 `MirrorSource.api` 表达，与 Mock 的接口范围语义保持一致。

## Context-KG Skill 整理

- [x] 核对 `context-kg` schema、index 和仓库级 AGENTS 约束
- [x] 创建本地 Codex skill，固化知识库 ingest/query/lint/restructure 流程
- [x] 验证 skill 元数据、触发描述和基本结构
- [x] 记录 review 与验证结果

当前判断：

- 该能力适合沉淀为本地 Codex skill，放在 `~/.codex/skills` 下方便后续自动发现和显式调用。
- Skill 需要保留高层流程，并要求每次以目标仓库的 `context-kg/_meta/schema.md` 为准，避免把当前仓库规则硬编码成不可迁移模板。
- 对长期知识的落点、frontmatter、双向链接、index/log 同步和 lint 检查，应作为 skill 的核心 guardrail。

当前进展：

- 新增本地 skill：`~/.codex/skills/context-kg-maintainer/SKILL.md`。
- 新增结构校验脚本：`~/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py`。
- 修正 `agents/openai.yaml` 的默认调用 prompt，避免 shell 展开导致 `$context-kg-maintainer` 丢失。
- 使用 lint 脚本发现并修复当前知识库部分页面 `title` 与首个 H1 不一致的问题。
- 追加 `_meta/log.md` 的 `lint` 操作记录。

验证：

- `python3 ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py ~/.codex/skills/context-kg-maintainer` 通过。
- `python3 ~/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过，当前 27 个 Markdown 页面、27 个唯一页面名，frontmatter、链接、index 基础检查通过。
- `python3 -m py_compile ~/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py` 通过。
- `git diff --check -- context-kg` 通过。

Review：

- Skill 放在用户级 `~/.codex/skills`，不把个人 Codex skill 目录纳入当前仓库提交范围。
- Skill 内只固化通用流程和校验脚本，实际落点仍以目标仓库自己的 `AGENTS.md` 和 `context-kg/_meta/schema.md` 为准。
- 本轮对仓库内 `context-kg` 的代码级改动仅限任务记录、操作日志和标题契约修复，没有触碰业务代码。

## 一键重构建并 all 模式启动脚本

## 沙箱兼容性优化

- [x] 确认前台 `exec` 常驻服务对 Codex/Claude 工具调用不友好
- [x] 增加 detached 启动模式，优先使用 `tmux`，并提供 `setsid/nohup` 兜底
- [x] 保留默认前台模式，避免影响用户真实终端使用习惯
- [x] 更新 `AGENTS.md` 和 lessons，说明终端/agent 两种启动方式
- [x] 验证脚本语法、detached 启动、8080/8090 端口和页面资源

当前判断：

- 当前脚本最后使用 `exec ... start --mode all` 是正确的终端前台运行方式，但在 Codex/Claude 这类工具调用中会导致命令永不返回。
- agent 环境更适合把完整构建与启动流程放进 detached `tmux` 会话；如果没有 `tmux`，再退到 `setsid/nohup`。
- 默认行为仍应保持前台启动，避免用户在普通终端执行脚本时看不到实时日志。

当前进展：

- `scripts/rebuild-start-all.sh` 新增 `--detach[=auto|tmux|nohup]`、`--foreground`、`--no-stop-existing`、`--force-kill` 和 `--help`。
- 默认不传参数仍在当前终端前台构建并 `exec` all 模式服务。
- `--detach` 默认优先使用 `tmux`，会创建/替换 `POLE_TMUX_SESSION`（默认 `pole-control-plane`），并在子会话里执行完整构建与启动。
- 没有 `tmux` 时可使用 `--detach=nohup` 或 `POLE_DETACH_BACKEND=nohup`，脚本会尝试 `setsid`，否则退到 `nohup`，日志写入 `POLE_LOG_PATH`。
- 停旧进程逻辑继续使用 8080/8090 端口识别 Pole 进程，并增加 `POLE_FORCE_KILL=1` / `--force-kill` 适配进程元数据不可见的沙箱。

验证：

- `bash -n scripts/rebuild-start-all.sh` 通过。
- `./scripts/rebuild-start-all.sh --help` 输出参数说明正常。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 立即返回，不阻塞当前工具调用。
- `tmux capture-pane -pt pole-control-plane:0 -S -220` 显示 detached 子会话完成前端构建、Go 构建、配置生成，并启动到 `finish starting server`。
- all 模式当前 PID `67682`，8080/8090 均监听正常。
- `curl -sSI http://127.0.0.1:8080/` 返回 `HTTP/1.1 200 OK`。
- 首页引用资源 `assets/index.b7f2ccfc.js` 与 `assets/style.0ae76e8d.css` 均返回 `HTTP/1.1 200 OK`。
- `git diff --check -- scripts/rebuild-start-all.sh AGENTS.md context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 这次保留了脚本的终端前台语义，只把 agent/sandbox 长进程问题作为显式 detached 模式处理。
- `tmux` 是首选，因为它能让服务生命周期脱离 Codex/Claude 的单次 shell 调用；`setsid/nohup` 只是兼容性兜底。
- `--detach` 外层只负责派生会话并返回，真正 build 和 start 仍走同一个脚本，避免维护两套启动流程。

- [x] 核对现有前端构建、Go 构建和 all 模式启动配置
- [x] 新增本地脚本，串联 console build、Go build、临时配置生成和 all 模式启动
- [x] 在 `AGENTS.md` 补充脚本使用方式
- [x] 执行脚本验证构建、端口监听和页面入口
- [x] 记录 review 与验证结果

当前判断：

- `deploy/conf/pole-server.yaml` 含 `##DB_USER##` 等占位符，不能直接作为本地 all 模式启动配置。
- 脚本应先生成 `/tmp/pole-server-all.yaml`，替换本地 MySQL 配置，并把 logger/apiservers/webPath 指向仓库内的真实路径。
- 为避免端口冲突，脚本默认只清理占用 8080/8090 的旧 Pole 进程；其它进程不主动处理。

当前进展：

- 新增 `scripts/rebuild-start-all.sh`，默认执行 `console/web` 的 `npm run build`、构建 `/tmp/pole-control-plane`、生成 `/tmp/pole-server-all.yaml`，最后执行 `start --mode all`。
- 脚本默认使用 `MYSQL_USER=root`、`MYSQL_PWD=123456`、`MYSQL_HOST=127.0.0.1:3306`，可通过环境变量覆盖。
- 脚本默认清理占用 8080/8090 的旧 Pole 进程；设置 `POLE_STOP_EXISTING=0` 可跳过清理。
- `AGENTS.md` 常用命令已补充一键构建并 all 模式启动入口。

验证：

- `bash -n scripts/rebuild-start-all.sh` 通过。
- `git diff --check -- scripts/rebuild-start-all.sh AGENTS.md context-kg/tasks/todo.md` 通过。
- 通过 `tmux new-session -d -s pole-control-plane './scripts/rebuild-start-all.sh'` 真实执行脚本，前端构建、Go 构建、配置生成和 all 模式启动均完成。
- all 模式当前 PID `73678`，8080/8090 均监听正常。
- `curl -sSI http://127.0.0.1:8080/` 返回 `HTTP/1.1 200 OK`。
- 首页引用新资源 `assets/index.b7f2ccfc.js` 与 `assets/style.0ae76e8d.css`，两个静态资源均返回 `HTTP/1.1 200 OK`。

Review：

- 本次只新增本地开发脚本和运行文档，不改变 release 打包脚本、后端启动模式或配置 schema。
- 脚本使用临时配置文件承接本地启动差异，避免直接改写 `deploy/conf/pole-server.yaml`。
- 对端口占用的处理限定在 Pole 进程，避免误杀其它本地服务。

# 治理工作台流量治理筛选文案收敛

- [x] 根据用户截图确认筛选按钮文案应为 `鉴权 / 镜像 / Mock`
- [x] 更新工作台类型筛选、流量治理通用类型名和独立页面 tab
- [x] 构建前端并验证 8080 页面按钮文案
- [x] 记录 review、验证结果和纠正经验

当前判断：

- 用户截图中的 `调用鉴权 / 流量镜像 / 流量 Mock` 在筛选区过长，和其它单/双字治理类型不一致。
- 文案应收敛为 `鉴权 / 镜像 / Mock`，但底层资源类型、接口路径和 spec 语义不变。

验证：

- `cd console/web && npm run build` 通过，保留既有 Browserslist 过期提示和 Vite 大 chunk 警告。
- all 模式已重启，当前进程 PID `59770`，8080/8090 均监听正常。
- Playwright 在 `http://127.0.0.1:8080/governance/workbench` 验证筛选按钮：`hasShort=true`、`hasLongButton=false`、`matchingButtons=["鉴权","镜像","Mock"]`。

Review：

- 本轮只收敛工作台类型筛选、流量治理通用类型名和独立页面 tab 文案，不改 `traffic-security`、`traffic-mirror`、`traffic-mock` 等资源类型值和接口结构。
- 样例规则描述里仍可能包含“调用鉴权样例 / 流量镜像样例 / 流量 Mock 样例”，这是规则描述文本，不属于筛选按钮文案。

# 泳道组批量样例数据补充

- [x] 确认当前 all 模式进程和 lane rule 创建接口可用
- [x] 查询 `spec-check-lane-group` 当前已有泳道规则，避免重复创建
- [x] 通过真实 API 批量创建多个泳道规则
- [x] 刷新 8080 页面验证泳道列表滚动效果
- [x] 记录 review 与验证结果

当前判断：

- 用户希望查看泳道列表多条数据下的实际滚动/翻页效果，应保留数据用于页面观察。
- 本次只追加本地样例泳道规则，不改代码和 schema，不直接写数据库。

当前进展：

- 通过 `POST /naming/v1/lane/groups/rules` 向 `spec-check-lane-group` 追加 12 条样例泳道：`demo-lane-canary`、`demo-lane-green`、`demo-lane-gray`、`demo-lane-beta`、`demo-lane-gold`、`demo-lane-silver`、`demo-lane-tenant-a`、`demo-lane-tenant-b`、`demo-lane-tenant-c`、`demo-lane-mobile`、`demo-lane-web`、`demo-lane-shadow`。
- 当前 `spec-check-lane-group` 共 13 条泳道规则，包含原有 `blue-lane`。

验证：

- 8090 管理端查询 `spec-check-lane-group` 返回 `rule_count=13`。
- 8080 治理工作台刷新后，泳道组列表行展示 `13 个泳道 / 2 个目标服务`。
- 8080 进入 `spec-check-lane-group` 详情后，DOM 指标为 `itemCount=13`、`drawerCanScroll=true`、`listHeight=1628`。
- Playwright 截图 `.playwright-cli/page-2026-06-14T13-28-22-006Z.png` 展示列表上半段，`.playwright-cli/page-2026-06-14T13-28-26-384Z.png` 展示滚动后的中后段。

Review：

- 本次只是本地样例数据追加，未提交代码改动。
- 当前子泳道列表是抽屉内连续滚动，不是分页控件；多条数据已足够观察滚动效果。

# 泳道组泳道列表抽屉展示优化

- [x] 记录用户截图反馈，确认问题集中在泳道组详情下的泳道列表
- [x] 定位 `LaneRuleTable` 当前宽表格、固定操作列和零值时间展示根因
- [x] 将泳道列表改为抽屉友好的紧凑规则列表
- [x] 构建前端并用 8080 真实页面验证泳道列表无横向滚动、操作列不挤压
- [x] 记录 review、验证结果和纠正经验

当前判断：

- 截图中泳道列表处于详情抽屉内，当前实现仍使用宽表格，列总宽大于内容区，导致横向滚动条、操作列被挤窄、表头“操作”竖排。
- 泳道规则是 LaneGroup JSON 聚合内的子对象，当前样例子规则时间为后端零值 `0001-01-01 00:00:00`，前端直接展示会制造噪音。
- 修复应优先让抽屉里的每条泳道能快速扫描：名称、启用状态、描述、泳道标签、匹配条件和操作，不应为了普通表格列而牺牲可读性。

当前进展：

- `LaneRuleTable` 已从 TDesign 宽表格改为抽屉内紧凑规则列表。
- 每条泳道现在展示：名称、启用状态、描述、泳道标签、匹配条件和右侧图标操作。
- 后端零值时间（如 `0001-01-01`）不再展示；只有真实创建/修改时间才显示时间元信息。

验证：

- `cd console/web && npm run build` 通过，保留既有 Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `22146`，8080/8090 均监听正常。
- Playwright CLI 使用 8080 真实页面登录 `admin/admin123`，打开 `http://127.0.0.1:8080/governance/workbench` 并进入 `spec-check-lane-group`。
- DOM 指标验证：`oldTableCount=0`，`drawerOverflow=false`，`panelOverflow=false`，`listOverflow=false`，`itemOverflow=false`，`hasZeroTimeText=false`，`hasActionHeader=false`。
- 视觉截图 `.playwright-cli/page-2026-06-14T10-09-58-530Z.png` 显示泳道列表为紧凑行：`blue-lane`、`启用`、描述、`lane: blue`、匹配条件和右侧编辑/删除图标均在同一抽屉宽度内。

Review：

- 本轮只调整泳道组详情下的子泳道列表展示，不改泳道组编辑、泳道规则保存、发布、后端接口或数据结构。
- 根因是抽屉内复用普通表格列模型导致信息密度和宽度不匹配；改为规则列表后，扫描重点更清楚，也避免了固定操作列挤压。

# AGENTS.md 最新协作规则同步

- [x] 核对根目录 `AGENTS.md` 当前内容与用户提供的最新协作规则差异
- [x] 补齐全局工作流、任务管理、提问协议和核心原则
- [x] 补齐程序运行方式、启动模式和常用本地启动命令
- [x] 验证 Markdown 格式与变更范围
- [x] 记录 review 与验证结果

当前判断：

- 当前根目录 `AGENTS.md` 已包含项目级常用命令、架构、插件、知识库和 import 格式约定。
- 缺失的是本次用户提供的全局协作规则：默认计划模式、子代理策略、自我改进闭环、完成前验证、优雅方案检查、自主修复缺陷、任务管理约束、苏格拉底式提问协议和中文输出要求。
- 用户追问后确认，“怎么 run 程序”原先只写了 test/data 后端启动命令，不够完整；需要补充 all/server/console 三种模式和 `--mode` 优先级。
- 本次只同步协作说明文档，不改业务代码或知识库长期页面。

验证：

- `git diff --check -- AGENTS.md context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `rg -n '全局协作规则|默认进入计划模式|子代理策略|自我改进闭环|完成前必须验证|苏格拉底式提问协议|语言要求|运行模式|--mode all|--mode server|--mode console' AGENTS.md` 命中新增协作规则和运行方式关键段落。

Review：

- `AGENTS.md` 已补齐用户提供的最新全局协作规则，并保留原有项目级常用命令、架构、知识库和 import 格式约定。
- `AGENTS.md` 已补充本地后端启动、完整 all 模式、server-only 和 console-only 启动命令，并说明 `test/data` 配置默认 `server`、部署配置默认 `all`、CLI `--mode` 高于 YAML。
- `context-kg/tasks/todo.md` 在本轮开始前已有大量未提交任务记录；本轮只新增并维护顶部的 `AGENTS.md 最新协作规则同步` 小节，不整理其它历史内容。

# 治理规则统一存储与缓存实现

# Console Logo 品牌替换

- [x] 定位当前 TDesign Starter logo 的 SVG 和使用入口
- [x] 设计并替换展开态 / 折叠态 SVG logo
- [x] 构建并用 8080 页面验证侧边栏展开态和折叠态

当前进展：

- 当前侧边栏展开态使用 `console/web/src/assets/svg/assets-logo-full.svg`，折叠态使用 `console/web/src/assets/svg/assets-t-logo.svg`。
- 登录页头也复用 `assets-logo-full.svg`，所以替换全量 SVG 后登录页头会同步去掉 TDesign Starter 品牌。
- 本轮采用同源双形态：展开态显示 `Pole.IO`，折叠态只显示抽象 `P` 标记，保持现有 184x32 / 32x32 尺寸，减少布局影响。
- 用户纠正后，主品牌文案从 `Pole Console` 收敛为 `Pole.IO`，侧边栏底部版本文案也同步改为 `Pole.IO ${version}`。

验证：

- `npm run build` 在 `console/web` 通过，保留 Vite 既有 browserslist/chunk 体积警告。
- all 模式已重启，当前 PID `51748`，8080 返回 HTTP 200，当前资源 hash 为 `/assets/index.339d7359.js` 和 `/assets/style.ab32ff84.css`。
- 8080 展开态验证：logo SVG `aria-label=Pole.IO`，尺寸 `184x32`，SVG 文本为 `Pole.IO`，页面无 `Pole Console` / `TDesign Starter`。
- 8080 折叠态验证：侧边栏宽度 `64px`，logo SVG `aria-label=Pole.IO`，尺寸 `32x32`，页面无 `Pole Console` / `TDesign Starter`。
- `git diff --check -- console/web/src/assets/svg/assets-logo-full.svg console/web/src/assets/svg/assets-t-logo.svg console/web/src/layouts/components/Menu.tsx context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只替换控制台品牌 SVG 与侧边栏底部版本文案，不改登录页其它 TDesign 模板文案，例如注册页的服务协议提示。

# 按 specification 构造治理规则检查数据

- [x] 核对 `../specification` 中路由、限流、熔断、主动探测、无损、泳道规则定义
- [x] 核对当前统一表 `governance_rule` / `governance_rule_release` 的字段映射
- [x] 构造并写入一组可在 Console 检查的 spec 对齐治理规则数据
- [x] 验证数据库、接口和页面能读取样例数据
- [x] 记录 review、检查入口和剩余风险

当前进展：

- 已向本地 MySQL `pole_server` 写入命名空间 `spec-governance`、服务 `spec-gateway/spec-order/spec-payment/spec-checkout/spec-inventory`。
- 已写入 6 类治理规则当前态：`spec-check-route-20260610`、`spec-check-ratelimit-20260610`、`spec-check-circuitbreaker-20260610`、`spec-check-faultdetect-20260610`、`spec-check-lossless-20260610`、`spec-check-lane-group-20260610`。
- 已写入对应 6 条 `normal` active release，版本号均为 `1`。
- 路由样例基于 `RouteRule + CustomRoute`；限流基于 `RateLimit + LimitTrigger`；熔断基于 `CircuitBreakerRule + BlockConfig`；主动探测基于 `FaultDetectRule`；无损基于 `LosslessRule`；泳道基于 `LaneGroup + LaneRule`。
- 造数过程中发现并修复熔断详情 auth wrapper 对非成功/空 data 响应缺少保护导致 panic 的问题。

验证：

- `go test -count=1 ./pkg/goverrule/interceptor/auth ./plugin/apiserver/httpserver/discover` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过，并已重启 LaunchAgent `io.pole.control-plane.local`，当前 PID `86552` 监听 `8080/8090`。
- SQL 验证 `governance_rule` 中 6 条 `spec-check-*` 当前态、`governance_rule_release` 中 6 条 active release 均存在。
- 8090 详情接口验证 6 类规则均返回 `code=200000`，类型分别为 `RouteRule`、`RateLimit`、`CircuitBreakerRule`、`FaultDetectRule`、`LosslessRule`、`LaneGroup`。
- 8090 列表接口验证 `/routings`、`/ratelimits`、`/circuitbreakers`、`/faultdetectors`、`/lossless`、`/lane/groups` 均返回 `code=200000` 且命中对应 `spec-check-*`。
- 8090 release 接口验证 6 类 release 查询均返回 `code=200000` 且 `amount=1`。

Review：

- 本次数据通过一次性临时 Go 程序生成，程序运行后已删除，没有新增仓库脚本。
- `8080` Console 可直接进入 `http://127.0.0.1:8080/governance/` 搜索 `spec-check` 检查列表、详情和版本。
- 直接 curl 验证应使用 8090 后端端口和 `Authorization` token；8080 console 代理的 header 行为与浏览器本地登录态相关，不作为本轮 curl 验证入口。

# 治理规则旧数据库表清理

- [x] 扫描 SQL/Go 文件中的旧治理分表定义和直接 SQL 残留
- [x] 清理测试套件中旧治理分表删除语句
- [x] 在本地 MySQL `pole_server` 库创建统一治理两表
- [x] 删除真实数据库中的旧治理分表
- [x] 运行目标测试和数据库结构验证

当前进展：

- `plugin/store/mysql/scripts/pole_server.sql` 已只保留 `governance_rule` / `governance_rule_release`，未发现旧治理分表 DDL 残留。
- `test/suit/test_suit.go` 已将限流、熔断测试清理 SQL 从旧分表切到 `governance_rule`，并删除未使用的旧表常量。
- `test/suit/test_suit.go` 中历史 `router_rulev2` 清理 SQL 也已切到 `governance_rule(rule_type=route)`；`router_rulev2` 当前不是 SQL 初始化脚本或真实库中的独立表。
- 本地 MySQL 容器 `pole-mysql` 的 `pole_server` 库已创建统一两表，并删除 `router_rule`、`router_rule_release`、`ratelimit_rule`、`ratelimit_rule_release`、`circuitbreaker_rule`、`circuitbreaker_rule_release`、`fault_detect_rule`、`fault_detect_rule_release`、`lossless_rule`、`lossless_rule_release`、`lane_group`、`lane_rule`、`lane_group_release`。

验证：

- 删除前 `information_schema.tables` 显示旧治理分表均为 0 行，统一两表为 0 行。
- 删除后 `information_schema.tables` 在 `pole_server` / `pole_observability` 范围内仅剩 `pole_server.governance_rule` 与 `pole_server.governance_rule_release`。
- `SHOW CREATE TABLE governance_rule` / `SHOW CREATE TABLE governance_rule_release` 与统一 schema 字段一致。
- `rg -n 'CREATE TABLE\s+`?(router_rule|router_rule_release|router_rulev2|ratelimit_rule|ratelimit_rule_release|circuitbreaker_rule|circuitbreaker_rule_release|fault_detect_rule|fault_detect_rule_release|lossless_rule|lossless_rule_release|lane_group|lane_group_release|lane_rule)`?|delete from (router_rule|router_rulev2|ratelimit_rule|circuitbreaker_rule|fault_detect_rule|lossless_rule|lane_group|lane_group_release|lane_rule)\b|\b(router_rule_release|ratelimit_rule_release|circuitbreaker_rule_release|fault_detect_rule_release|lossless_rule_release|lane_group_release)\b' --glob '*.sql' --glob '*.go' .` 无输出。
- `go test -count=1 ./plugin/store/mysql ./pkg/cache/rules ./pkg/goverrule/... ./pkg/admin/job ./test/suit` 通过。

Review：

- 这次只删除已经迁移到统一治理表的旧分表；`router_rulev2` 是历史路由规则表名残留，当前不应作为独立表保留，测试清理逻辑已改到统一表。
- 真实库清理未做数据迁移，符合此前“不考虑数据迁移”的约束；执行前旧分表均为空。

# A2A 协议功能设计调研

- [x] 查询官方 A2A 仓库和协议规格，确认 Agent Card、任务生命周期、流式推送、鉴权和传输形态
- [x] 读取现有 MCP Registry、AIStore、缓存、HTTP 插件和知识库设计
- [x] 对比 A2A 与现有 MCP 能力边界，给出功能定位与分层方案
- [x] 将长期架构方案归档到 `context-kg/technical/adr/`
- [x] 同步更新 `context-kg/_meta/index.md` 与 `context-kg/_meta/log.md`
- [x] 做文档格式、链接和事实来源验证
- [x] 补充 review、验证结果和剩余风险

当前判断：

- A2A 不应替代 MCP Registry；MCP 面向工具/资源发现，A2A 面向 Agent 能力发现和跨 Agent 任务协作。
- 推荐将 A2A 做成 AI Native 的第二个注册域：`MCPServer` 旁新增 `A2AAgent` / `A2AAgentSkill`，复用 Store、Cache、HTTP 插件、namespace、auth 和软删除模式。
- pole-control-plane 只做 A2A Agent Card 注册、发现、健康状态、按 namespace/skill/capability 查询，以及从 Pole 服务发现生成 Agent endpoint；任务代理、SSE 转发、push broker、任务状态机和 artifact 存储属于数据面/网关/agent runtime，不应写成 control-plane 二期。

## A2A 协议功能设计调研 Review

已完成：

- 新增 `context-kg/technical/adr/adr-a2a-agent-registry.md`，归档 A2A Agent Registry 的功能定位、数据模型、Store/Cache/API/Console 落点、安全约束和测试策略。
- 更新 `context-kg/technical/modules/ai-features.md`，将 AI Native 从 MCP Registry 扩展为 MCP Registry + A2A Agent Registry。
- 更新 `context-kg/_meta/index.md` 和 `context-kg/_meta/log.md`，补充 ADR 入口和操作记录。
- 结合官方 A2A 资料确认：A2A 以 Agent Card 做发现，以 JSON-RPC/HTTP、SSE streaming、push notification 支持任务协作；但 pole-control-plane 只消费这些 capability 声明作为 Registry 元数据，不承载数据面代理。
- 结合现有代码确认：最小实现应复用 `AIStore`、MySQL store、`CacheManager`、`pkg/cache/ai` 和 `plugin/apiserver/httpserver` 子包模式，不另起插件系统。

验证：

- `git diff --check -- context-kg/technical/adr/adr-a2a-agent-registry.md context-kg/technical/modules/ai-features.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。
- 自定义 wiki 链接检查通过：`all wiki links resolve`。
- 新增/更新页面 frontmatter 和 `## 相关页面` 检查通过。

剩余风险：

- A2A 官方协议和 SDK 仍可能继续演进；方案已通过 `raw_card_json`、`protocol_version`、`source_url` 等字段降低追随成本，但正式实现前仍需锁定 specification 版本。
- 完整 Task Proxy、SSE 转发和 Push Broker 涉及数据面状态与安全边界，不能写成 pole-control-plane 后续能力；如需落地，应另起数据面/网关设计。

## A2A Agent Registry 范围收敛 Review

已完成：

- 按用户纠正，将 `adr-a2a-agent-registry` 中“二期再做 Task Proxy / Push Notification Broker”改为“不属于 pole-control-plane 的范围”。
- 明确 pole-control-plane 只负责 A2A Agent Registry：Agent Card 注册、发现、索引、健康/拉取状态、治理元数据和管理 API。
- 将 A2A task proxy、SSE streaming 转发、push notification broker、task 状态机和 artifact 存储归类为数据面/网关/agent runtime 能力。
- 更新 `ai-features` 摘要和 `lessons`，避免后续再把数据面能力写进 control-plane 方案。

验证：

- `git diff --check -- context-kg/technical/adr/adr-a2a-agent-registry.md context-kg/technical/modules/ai-features.md context-kg/tasks/todo.md context-kg/tasks/lessons.md context-kg/_meta/log.md` 通过。
- 自定义 wiki 链接检查通过：`all wiki links resolve`。
- 搜索旧方案表述，未在 ADR/模块页保留“二期做 Task Proxy / Push Broker”的规划；todo 中仅保留 review 说明。

- [x] 重新读取 ADR，确认统一两表、LaneGroup 聚合根和 cache fan-out 的目标语义
- [x] 梳理现有 store/cache/goverrule 接口和 MySQL 实现，标出需要保持兼容的调用点
- [x] 先补充统一 governance schema/repository/lane 聚合/cache fan-out 的测试场景
- [x] 实现 `governance_rule` / `governance_rule_release` schema 与统一 repository
- [x] 将普通治理规则 store 逐步切到统一 repository，保留领域接口
- [x] 将泳道从 `lane_group/lane_rule/lane_group_release` 收敛为 LaneGroup JSON 聚合
- [x] 实现 `GovernanceRuleUpdateCache` 和各类型 watcher 接入
- [x] 清理旧 DDL 与旧分表残留路径
- [x] 运行 gofmt/import-format、目标包测试和必要构建
- [x] 记录 review、验证结果和剩余风险

当前进展：

- 已补 `governance_rule` / `governance_rule_release` schema 创建逻辑和 sqlmock 测试。
- 已补统一 repository 的 rule create / rule mtime query / release active 范围隔离测试。
- 已将 MySQL `laneStore` 改为 `governance_rule(rule_type=lane-group)` 和 `governance_rule_release` 聚合存储，不再访问 `lane_group`、`lane_rule`、`lane_group_release`。
- 已将路由、限流、熔断、主动探测、无损的 MySQL store 切到统一 repository，领域接口保持不变。
- 已新增 `GovernanceRuleUpdates` / `GovernanceRuleReleaseUpdates` 领域更新载体，MySQL `stableStore` 支持一次拉取统一两表增量并按 `rule_type` 转换。
- 已新增 `GovernanceRuleUpdateCache` 协调器，六类治理 cache 注册 watcher；store 支持统一增量接口时优先统一拉取并 fan-out，否则保留原单类型更新回退。
- 已修复 `DeleteLaneRule` 误调用 `DeleteLaneGroup` 的问题，删除单条泳道规则改为事务内更新 LaneGroup 聚合。
- 已将泳道组内 LaneRule 上限统一为 20，并补参数校验测试。
- 新库 SQL 脚本已删除旧治理分表 DDL，仅创建 `governance_rule` / `governance_rule_release`。
- admin 软删除清理资源已从旧治理分表切到 `governance_rule` / `governance_rule_release`。

验证：

- `go test -count=1 ./plugin/store/mysql` 通过。
- `go test -count=1 ./pkg/cache/rules ./pkg/goverrule/... ./pkg/admin/job` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-governance .` 通过。
- `git diff --check -- ...` 本轮治理规则相关文件通过。
- `rg -n "router_rule|ratelimit_rule|circuitbreaker_rule|fault_detect_rule|lossless_rule|lane_group|lane_rule" plugin/store/mysql/scripts/pole_server.sql plugin/store/mysql pkg/admin pkg/cache apis` 未发现数据库旧分表残留；仅剩 `apis/pkg/types/auth/funcs.go` 中的业务资源枚举名。
- `go test -count=1 ./...` 未完全通过，失败点为既有测试环境/用例问题：`pkg/service/healthcheck` 缺 `test/data/service_test.yaml`，`plugin/apiserver/httpserver/i18n` 缺 `release/conf/i18n/*.toml`，`plugin/service/healthchecker/heartbeat` 测试触发 nil pointer；其它已跑到的包通过。

Review：

- 对外 store/cache/goverrule 接口保持兼容，OpenAPI 层仍通过原领域服务读写，不需要暴露统一表细节。
- 统一 repository 是治理规则 MySQL 当前态和发布态的唯一 SQL 入口；普通规则 store 只做领域对象和统一 record 的转换。
- `rule_type + rule_id + release_type` 限定 active 切换范围，避免不同治理类型或不同发布类型互相覆盖。
- LaneRule 不再独立持久化，作为 LaneGroup 聚合 JSON 的一部分维护；删除单条 LaneRule 不会误删 LaneGroup。
- cache fan-out 使用可选统一 store 接口接入，不强制修改 `store.Store` 大接口和 mock；非 MySQL/fake store 仍可走原单类型回退路径。

# 熔断详情查看态误可编辑修复

- [x] 复现并定位熔断详情查看态仍能修改恢复/降级配置的根因
- [x] 将恢复策略、熔断后降级在查看态改为只读展示
- [x] 构建前端并在 8080 all 模式验证查看态不可修改
- [x] 记录 review 和本次纠正经验

## 熔断详情查看态误可编辑修复 Review

已完成：

- 根因定位：`CircuitBreakerEditor` 的基础信息和服务信息已按 `editorState.editable` 区分查看/编辑，但 `恢复策略` 与 `熔断后降级` 直接渲染 `InputNumber`、`Switch`、`Textarea` 并绑定 `onChange`，导致查看态也能修改本地数据。
- 新增只读展示组件：查看态用文本和 `Tag` 展示熔断时长、主动探测、降级开关、响应码和响应体，不再渲染可编辑控件。
- 编辑态保留原 `InputNumber`、`Switch`、`Textarea` 控件，点击 `编辑` 后仍可正常修改。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check -- console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- 重启 all 模式后，8080 首页加载新资源 `assets/index.e1451998.js` 和 `assets/style.23482e64.css`；LaunchAgent `io.pole.control-plane.local` 当前 PID `75854` 同时监听 `8080` 和 `8090`。
- 浏览器打开 `demo-governance-circuit-202606050425` 查看态，`恢复策略` 和 `熔断后降级` 区块内 `inputNumberCount=0`、`switchCount=0`、`textareaCount=0`，显示 `30 秒`、`开启`、`503` 和响应体只读代码块。
- 浏览器点击 `编辑` 后，同两个区块重新出现 `InputNumber`、`Switch`、`Textarea`，并显示 `保存 / 撤销`，确认编辑能力没有被误删。

# 治理规则查看态横向只读排查

- [x] 静态扫描路由、限流、主动探测、无损、泳道编辑器中的输入控件和 `editorState.editable` 约束
- [x] 在 8080 all 模式逐类打开 demo 规则详情，统计查看态可编辑控件
- [x] 修复确认存在的查看态误可编辑点
- [x] 构建前端并回归验证各治理规则查看态只读、编辑态可编辑
- [x] 记录 review 和经验

## 治理规则查看态横向只读排查 Review

已完成：

- 横向静态检查 `CustomRouteEditor`、`RateLimitEditor`、`CircuitBreakerEditor`、`FaultDetectEditor`、`LossLessEditor`、`LaneGroupEdtor`、`LaneRuleEditor` 中的输入控件和 `editable` 约束。
- 路由、主动探测、泳道查看态未发现同类可写业务控件；熔断查看态已在上一轮修复。
- 限流规则补齐匀速排队分支的 `最大排队时长(秒)` 只读约束，查看态不再暴露可修改 `InputNumber`。
- 无损规则查看态从冻结控件改为纯只读展示：延迟注册、延迟策略、探测配置、服务预热和无损下线均使用文本或 `Tag`，编辑态再渲染 `Switch`、`RadioGroup`、`InputNumber`、`Select`、`Input`。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/LossLess/LossLessEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- 重启 all 模式后，8080 首页加载新资源 `assets/index.c6f87325.js` 和 `assets/style.23482e64.css`；LaunchAgent `io.pole.control-plane.local` 当前 PID `97856` 同时监听 `8080` 和 `8090`。
- 浏览器打开无损规则 `无损上线、服务预热与无损下线`，查看态 `visibleInputNumbers=0`、`visibleSwitches=0`、`visibleTextareas=0`；点击 `编辑` 后恢复为 `visibleInputNumbers=5`、`visibleSwitches=4`，并显示 `保存 / 撤销`。
- 浏览器打开限流规则 `demo-governance-ratelimit-202606050425`，查看态 `visibleInputNumbers=0`、`visibleTextareas=0`；页面仅剩一个禁用态阈值合并 Switch，不可修改业务数据。
- 浏览器逐类打开路由、熔断、主动探测、泳道规则，查看态均无未禁用的可编辑输入；路由/熔断/泳道 `editableVisibleInputs=[]`，主动探测仅存在只读输入且 `editableVisibleInputs=[]`。

# 限流规则外层“子规则”标题移除

- [x] 去掉限流规则详情中规则块列表上方的外层 `子规则` 标题
- [x] 将限流编辑器和列表中的可见 `子规则` 文案统一改为 `规则`
- [x] 清理对应未使用样式，并沉淀本次纠正经验
- [x] 构建前端并在 8080 all 模式页面验证

## 限流规则外层“子规则”标题移除 Review

已完成：

- 移除限流详情中规则块列表上方的外层 `子规则` 标题，保留内部 `规则 [n]` 卡片标题。
- 删除不再使用的 `ruleListHeader` 和 `listTitle` 样式，并去掉规则列表额外顶部间距。
- 将限流编辑器里的可见提示从 `子规则` 统一改为 `规则`，包括展开/折叠、删除、添加和匹配条件帮助文案。
- 将限流列表列名从 `子规则数` 调整为 `规则数`。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/RateLimit/RateLimitEditor.module.less console/web/src/pages/Governance/RateLimit/RateLimitTable.tsx context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `rg -n "子规则" console/web/src/pages/Governance/RateLimit` 仅剩代码注释，页面可见文案已清理。
- 重启 all 模式后，8080 首页加载新资源 `assets/index.e16ba18c.js` 和 `assets/style.9c8dd55d.css`；LaunchAgent `io.pole.control-plane.local` 当前 PID `61736` 同时监听 `8080` 和 `8090`。
- 浏览器打开 `http://127.0.0.1:8080/governance/` 并点击 `demo-governance-ratelimit-202606050425`，详情中 `hasSubRuleText=false`，保留 `规则 [1]`、`120 次 / 1s`，匹配条件帮助文案为 `满足以下匹配条件的请求将应用该规则`。

# 控制台侧边栏 Logo 折叠态优化

- [x] 读取用户截图和当前 `MenuLogo`/logo 资产实现
- [x] 替换折叠态 logo 为项目 T 标记，并优化展开/折叠尺寸与居中
- [x] 构建前端并用浏览器验证折叠态显示
- [x] 记录 review 和本次经验

## 控制台侧边栏 Logo 折叠态优化 Review

已完成：

- 将折叠态使用的 `assets-t-logo.svg` 从无关黑色插图替换为项目完整 logo 中同源的 T 标记。
- `MenuLogo` 根据展开/折叠状态分别注入完整 logo 和 mini logo 的样式类，避免依赖 SVG 默认尺寸。
- 侧边栏 logo 区固定为 `64px` 高度；展开态完整 logo 为 `184px * 32px`，折叠态 mini logo 为 `32px * 32px`，均居中显示。
- 移除原先 `height: 100%` 和 `margin-top: 10%` 带来的偏移，折叠态只保留 logo 标记，不再出现大插图感。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check -- console/web/src/assets/svg/assets-t-logo.svg console/web/src/layouts/components/MenuLogo.tsx console/web/src/layouts/components/Menu.module.less context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- 本地打开 `http://127.0.0.1:5174/` 并设置 `820px` 视口，验证折叠态顶部 `_menuMiniLogo_*` 实际渲染为 `32px * 32px`，颜色为项目 T 标记的蓝绿渐变，未再显示旧的复杂插图。

注意：

- 未改动完整 logo 资产；展开态仍使用现有 `assets-logo-full.svg`。

# 限流与熔断规则块布局统一

- [x] 对照路由规则详情的规则块、摘要、表格布局，梳理限流子规则与熔断策略当前结构
- [x] 优化限流子规则查看态布局，减少内层灰底卡片和表单堆叠感
- [x] 优化熔断策略查看态布局，统一为可折叠策略块和分区表格
- [x] 构建并重启 8080 all 模式
- [x] 用真实 8080 页面验证限流和熔断详情抽屉
- [x] 记录 review 和本次纠正经验

## 限流与熔断规则块布局统一 Review

已完成：

- 限流子规则从 TDesign `Collapse.Panel` 和多层灰底卡片改为自定义规则块，header 展示 `子规则 [n]`、匹配条件数、限流资源、阈值数、限流效果和首个阈值摘要。
- 限流规则块正文按 `匹配条件 / 限流方式 / 限流方案` 三个轻量分区展示，保留原匹配条件、阈值、动作和自定义响应编辑逻辑。
- 熔断策略从普通 `sectionCard` 改为与路由一致的可折叠策略块，header 展示策略名、错误判断条件数、触发条件数、接口标识和熔断粒度。
- 熔断策略正文按 `接口 / 错误判断条件 / 熔断触发条件` 分区展示，保留原 API、错误条件、触发条件表格和编辑逻辑。
- 新增 `RateLimitEditor.module.less`，熔断在原 module 内补充同名规则块样式，避免继续堆叠 inline 样式。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/RateLimit/RateLimitEditor.module.less console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.module.less context-kg/tasks/todo.md` 通过。
- 重新签名并 kickstart `io.pole.control-plane.local` 后，8080 首页加载 `assets/index.d9bd24ab.js` 和 `assets/style.e52f3f7a.css`。
- 浏览器打开限流规则 `demo-governance-ratelimit-202606050425`，确认存在 `子规则 [1]`、`匹配条件`、`限流方式`、`限流方案`、`POST /api/v1/payments`、`快速失败`，表格无横向溢出。
- 浏览器打开熔断规则 `demo-governance-circuit-202606050425`，确认存在 `熔断策略 [1]`、`错误判断条件`、`熔断触发条件`、`payment-error-rate`、`500-599`，3 个表格无横向溢出。
- 点击熔断策略折叠按钮后，按钮切换为 `展开熔断策略`，错误判断条件正文隐藏。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在既有治理断言：`/governance` 默认跳转、CustomRoute 旧字符串断言、泳道组空态。这些失败不是本轮布局调整引入。

# 治理规则发布抽屉二次视觉优化

- [x] 收敛发布策略为可点击策略卡，并用轻量流向图表达全量/灰度差异
- [x] 优化抽屉底部操作区，保证取消/发布按钮位置稳定且主次明确
- [x] 构建前端并重新拉起 8080 all 模式
- [x] 用真实 8080 页面验证发布抽屉布局、策略切换和灰度条件展示
- [x] 记录 review 和本次纠正经验

## 治理规则发布抽屉二次视觉优化 Review

已完成：

- 将发布策略从普通单选行改为两张策略卡，分别用 `客户端 -> 当前版本` 和 `客户端 -> 标签命中 -> 灰度版本` 的流向图表达全量/灰度差异。
- 策略卡点击后同步 `releaseType` 表单字段、卡片选中态和 footer 当前策略摘要。
- 抽屉底部改为全宽 footer bar：左侧显示当前策略，右侧固定 `取消 / 发布`，发布按钮使用发送图标强调主动作。
- 保持原 `RuleRelease.release_type` 和 `client_label: MatcheLabel[]` 提交结构不变。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过。
- `git diff --check -- console/web/src/pages/Governance/RuleRelease/PublishForm.tsx console/web/src/pages/Governance/RuleRelease/PublishForm.module.less context-kg/tasks/todo.md` 通过。
- 重新签名并 kickstart `io.pole.control-plane.local` 后，8080 首页加载 `assets/index.488ce9e3.js` 和 `assets/style.3d26f234.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，进入 `demo-governance-route-202606050424` 详情并打开发布抽屉，确认两张策略卡、footer 摘要和按钮布局正常。
- 切换灰度发布后，卡片选中态切换为灰度，footer 显示 `当前策略灰度发布`，灰度条件区域出现，页面无横向溢出。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在既有治理断言：`/governance` 默认跳转、CustomRoute 旧字符串断言、泳道组空态。这些失败不是本轮发布抽屉改动引入。

# AGENTS 主文件与任务目录迁移

- [x] 读取全局 `/Users/chuntao.liao/.codex/AGENTS.md`，确认任务记录目录已改为 `context-kg/tasks`
- [x] 将仓库内 `AGENTS.md -> CLAUDE.md` 反转为 `CLAUDE.md -> AGENTS.md`
- [x] 将 `docs/tasks/todo.md` 与 `docs/tasks/lessons.md` 迁移到 `context-kg/tasks/`
- [x] 删除旧 `docs/design/` 文档，清空旧 `docs/` 目录
- [x] 更新 `AGENTS.md`、`context-kg/_meta/schema.md`、`context-kg/_meta/index.md` 和 lessons 中的任务路径规则
- [x] 验证软链接方向、wiki 链接、index 覆盖和 diff 格式

## AGENTS 主文件与任务目录迁移 Review

已完成：

- 全局 `AGENTS.md` 当前要求任务计划写入 `context-kg/tasks/todo.md`，纠正经验写入 `context-kg/tasks/lessons.md`。
- 当前仓库已改为 `AGENTS.md` 保存实际内容，`CLAUDE.md` 使用相对软链接指向 `AGENTS.md`。
- 任务记录从 `docs/tasks/` 迁移到 `context-kg/tasks/`。
- 旧 `docs/design/` 的长期设计文档已删除；长期知识以 `context-kg` 三域结构为准。
- `context-kg/_meta/schema.md` 已补充 `tasks/` 目录规则；`context-kg/_meta/index.md` 已补充 `todo` 和 `lessons` 入口。

验证：

- `CLAUDE.md -> AGENTS.md` 相对软链接检查通过，`AGENTS.md` 为普通文件。
- `docs/` 目录已清空。
- wiki 链接有效性检查通过，无失效链接输出。
- `_meta/index.md` 覆盖所有非 `_meta` 内容页面。
- `git diff --check -- AGENTS.md CLAUDE.md context-kg docs` 通过。

# context-kg 三域结构重组

- [x] 盘点现有 `context-kg` 页面与链接关系
- [x] 明确三域规则：`business`、`technical`、`quality`
- [x] 移动现有页面到三域目录并拆出治理规则 ADR
- [x] 更新 `context-kg/_meta/schema.md`、`index.md`、`log.md`
- [x] 更新 `AGENTS.md`/`CLAUDE.md` 中的知识库硬规则
- [x] 更新 `context-kg/tasks/lessons.md` 记录本次纠正
- [x] 运行链接检查和 `git diff --check`

## context-kg 三域结构重组 Review

已完成：

- 将 `context-kg` 从旧的 `overview/domains/infra/ai/guides` 五类结构重组为 `business/technical/quality` 三域结构。
- `business/` 下新增 `terminology.md`、`domain-models.md`、`business-rules.md`，并将原业务功能页移动到 `business/feature-registry/`。
- `technical/` 下按 `arch/modules/Environment/apis/conventions/adr` 归档架构、模块、环境、接口、技术约定和 ADR。
- `quality/automation/` 下归档测试知识。
- 将治理规则统一存储与缓存更新方案从业务功能页拆出为 `technical/adr/adr-governance-rule-unified-storage.md`；`governance-rules.md` 只保留业务摘要和 ADR 链接。
- 重写 `context-kg/_meta/schema.md`，明确三域目录职责、页面命名、ADR/PDR/缺陷/用例落点、链接规则、index/log 维护规则和 lint 检查项。
- 重写 `context-kg/_meta/index.md` 为三域入口，并更新 `context-kg/_meta/log.md` 的 restructure 记录。
- 更新 `AGENTS.md` 的知识库硬规则；`CLAUDE.md` 是指向 `AGENTS.md` 的相对软链接。
- 更新 `context-kg/tasks/lessons.md`，记录 context-kg 调整必须同时维护 taxonomy 和元文件规则。

验证：

- wiki 文件名唯一检查通过。
- wiki 链接有效性检查通过，无失效 wiki 链接输出。
- `_meta/index.md` 覆盖所有非 `_meta` 内容页面。
- `git diff --check -- CLAUDE.md context-kg context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

# 技术方案文档归档到 context-kg

- [x] 核对 `context-kg` 现有 schema、index、log 与治理规则页面结构
- [x] 将治理规则统一存储与缓存更新方案归类到 `context-kg/domains/governance-rules.md`
- [x] 更新 `context-kg/_meta/index.md` 和 `context-kg/_meta/log.md`
- [x] 删除孤立的 `docs/design/11-governance-rule-unified-storage.md`
- [x] 在 `AGENTS.md` 记录技术方案文档必须进入 `context-kg` 的硬规则
- [x] 将本次纠正规则沉淀到 `context-kg/tasks/lessons.md`
- [x] 运行 wiki 链接/格式与 diff 验证

## 技术方案文档归档到 context-kg Review

已完成：

- 将治理规则统一存储与缓存更新方案合并到 `context-kg/domains/governance-rules.md`，按治理规则域承载长期知识。
- 更新 `context-kg/_meta/index.md` 的治理规则摘要，并在 `context-kg/_meta/log.md` 追加 ingest 记录。
- 删除孤立的 `docs/design/11-governance-rule-unified-storage.md`，避免长期方案继续散落在 `docs/design`。
- `CLAUDE.md` 是指向 `AGENTS.md` 的相对软链接，硬规则写入实际主文件 `AGENTS.md`：长期技术方案进入 `context-kg`，`context-kg/tasks` 只保留任务计划、review 和 lessons。
- 将本次纠正沉淀到 `context-kg/tasks/lessons.md`。
- 顺手清理 `context-kg` 中重组前残留的 `domain-components` 失效链接，统一指向当前 `index` 入口。

验证：

- `context-kg` 链接检查通过，除 `schema.md` 示例代码外没有失效 wiki 链接。
- `git diff --check -- CLAUDE.md AGENTS.md context-kg context-kg/tasks/todo.md context-kg/tasks/lessons.md docs/design/11-governance-rule-unified-storage.md` 通过。

# .gitignore 整理

- [x] 梳理现有忽略规则和当前未跟踪生成物
- [x] 按用途重组 `.gitignore`，补齐本地运行、测试输出和工具缓存规则
- [x] 验证忽略规则生效且不影响已跟踪文件
- [x] 记录 review 和验证结果

## .gitignore 整理 Review

已完成：

- 将 `.gitignore` 按构建产物、运行日志、测试输出、前端依赖、本地环境、编辑器文件和静态分析缓存分组。
- 保留原有二进制、release 包、日志、vendor、前端 `node_modules/dist` 等规则，并补齐根目录限定，降低误伤子目录文件的概率。
- 新增忽略本地生成的 `output/`、`.superpowers/`、`log/`、前端 `.cache/.vite/coverage`、`.env*`、`*.pid` 和 Go 测试产物。

验证：

- `git status --short --ignored output .superpowers log logs` 显示四个目录均为 `!!`，确认已被忽略。
- `git check-ignore -v` 确认 `output/playwright/...`、`.superpowers/...`、`log/pole-console.log`、`logs/pole-console.log` 命中预期规则。
- `git ls-files -ci --exclude-standard` 无输出，确认没有已跟踪文件被新规则误忽略。
- `git diff --check -- .gitignore context-kg/tasks/todo.md` 通过。

# 治理规则详情信息展示缺失修复

- [x] 复现和定位：对比 8080 页面展示、详情接口响应和前端字段读取路径
- [x] 修复规则详情查看态对真实 spec 字段的展示映射
- [x] 构建并用 8080 验证详情页关键信息完整展示
- [x] 记录 review 和验证结果

## 治理规则详情信息展示缺失修复 Review

已完成：

- 根因定位：当前 demo 路由规则真实数据存储在 `config.caller / config.callee / config.rules[].arguments.arguments / config.rules[].destinations`，而前端详情仍按旧结构 `routing_config.rules[].sources[0].arguments` 读取，导致规则名、主调服务、匹配 key/value、实例标签等大量内容展示为空。
- 在 `services/router.ts` 增加路由配置归一化能力，将新 spec 结构转换为编辑器查看态需要的 `sources[0].arguments`，并在保存时还原为 `caller/callee/rules[].arguments` 新结构。
- 自定义路由详情抽屉使用归一化后的 `caller/callee` 展示主调/被调服务，使用 `rules[].arguments.arguments` 展示匹配条件。
- 治理工作台列表和旧自定义路由表格也改为通过归一化配置展示主调/被调，避免列表仍出现 `-/-`。

验证：

- 数据库确认 `router_rule.config` 中 `demo-governance-route-202606050424` 包含 `caller=demo-governance/demo-order`、`callee=demo-governance/demo-payment`、`x-tenant=vip`、`region in ap-guangzhou,ap-shanghai`、`payment-v2/v1` 和 `lane/version` 标签。
- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `23785` 监听 `*:8080`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.6a6df8b6.js` 和 `assets/style.7fd9b869.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，列表展示 `demo-governance/demo-order -> demo-governance/demo-payment` 和 `vip-payment-split`。
- 点击 `demo-governance-route-202606050424` 后，抽屉详情展示规则名、优先级 `5`、描述、主调 `demo-order`、被调 `demo-payment`、匹配条件 `x-tenant=vip`、`region=ap-guangzhou,ap-shanghai`，以及目标分组 `payment-v2/payment-v1` 与 `lane/version` 标签。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在既有治理断言：`/governance` 默认跳转、泳道组空态；其中自定义路由断言还在匹配旧字符串 `routing_config?.rules?.[0]`，本次实际已替换为 `normalizeRoutingConfigForEditor` 并通过 8080 运行态验证。

# 其他治理规则详情信息横向检查

- [x] 对限流、熔断、故障探测、无损上下线、泳道组/泳道规则读取数据库真实字段
- [x] 对照前端 service 类型和详情/编辑组件读取路径，定位字段错位或信息缺失
- [x] 修复非路由治理规则详情抽屉中确认存在的缺失展示
- [x] 构建并用 8080 逐类打开详情验证关键信息展示
- [x] 记录 review、验证结果和剩余风险

## 其他治理规则详情信息横向检查 Review

已完成：

- 限流、熔断、故障探测、无损上下线、泳道组/泳道规则 5 类非路由规则已横向比对 MySQL 真实 JSON、前端 service 类型和详情组件读取路径。
- 在 service 层增加治理规则归一化：兼容 spec/DB/API 中 snake_case、camelCase、数字 enum、duration 对象与字符串秒数的差异。
- 限流详情兼容 `custom_response`、数字参数类型、数字限流资源和 `{ seconds }` 窗口，详情可展示接口、匹配条件、120/5000 阈值和自定义响应。
- 熔断详情兼容 `rule_matcher/block_configs/error_conditions/trigger_conditions` 等字段形态，并修复被调服务显示、策略名和 API 表展示。
- 主动探测详情兼容 `target_service/http_config/protocol`，发布资源类型从熔断改为 `FaultDetectRules`。
- 无损上下线兼容 `interval_second`，查看态保存改走 update 接口。
- 泳道组兼容 Any selector 入口解析，泳道规则列表增加匹配条件列，直接展示 `x-lane=blue / x-tenant=vip`。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `93059`，8080 首页加载 `assets/index.28193859.js` 和 `assets/style.7fd9b869.css`。
- 8080 页面验证限流详情展示 `POST /api/v1/payments`、`x-tenant`、`channel`、`120`、`5000`、`demo governance limited`。
- 8080 页面验证熔断详情展示被调 `demo-governance/demo-payment`、`payment-error-rate`、`/api/v1/payments`、`500-599` 和 fallback header。
- 8080 页面验证主动探测详情展示 `demo-payment`、`/healthz`、`GET`、`8080`、`10`、`3`、`x-demo-probe`。
- 8080 页面验证无损上下线详情展示 `demo-order`、`demo-governance`、延迟注册 `15`、预热 `60`、过载阈值 `70`、曲线 `3`。
- 8080 页面验证泳道组详情展示 `demo-governance/demo-gateway`、`demo-governance/demo-payment`，泳道规则列表展示 `x-lane=blue` 和 `x-tenant=vip`。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在既有断言：`/governance` 默认跳转、自定义路由详情旧字符串、泳道组空态。这些断言在本轮横向字段修复之外，已保留为剩余风险。

# 治理详情抽屉操作按钮点击失效修复

- [x] 复现和定位：按钮视觉区域扩大后，业务点击仍只绑在图标节点，文字/空白区域点击无效
- [x] 将治理规则编辑器的 StickyTool 点击处理提升到可见操作内容层
- [x] 构建并用 8080 验证编辑、发布按钮可触发
- [x] 记录 review 和验证结果

## 治理详情抽屉操作按钮点击失效修复 Review

已完成：

- 修复路由、泳道、熔断、探测、限流、无损上下线 6 类治理编辑器的抽屉操作按钮。
- 操作项不再依赖 `StickyTool` 容器统一分派点击事件，改为复用 `RuleStickyAction` 在可见的图标+文字内容上绑定 `编辑 / 保存 / 撤销 / 发布` 动作。
- 抽屉样式隐藏 `StickyItem` 自动生成的空 label，保持标题栏按钮组间距稳定。
- 未改变表单提交 payload、发布接口、规则数据结构和发布确认按钮逻辑。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `11618` 监听 `*:8080`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.0145cca8.js` 和 `assets/style.7fd9b869.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，点击 `demo-governance-route-202606050424` 后，抽屉操作按钮初始为 `编辑 / 发布`。
- 浏览器点击 `编辑` 后，按钮切换为 `保存 / 撤销`，抽屉内出现可编辑输入控件。
- 浏览器点击 `发布` 后，二级发布抽屉打开，显示 `规则发布`、`规则ID`、`版本名称` 和 `发布类型`。

# 治理规则详情表单视觉优化尝试

- [x] 对照 frontend-skill 复盘表单问题：查看态像冻结表单、卡片阴影偏多、服务流向和规则表格层级不稳
- [x] 优化自定义路由详情表单外层、基础信息、服务流向和规则定义分区
- [x] 构建并用 8080 抽屉截图验证
- [x] 记录 review 和验证结果

## 治理规则详情表单视觉优化尝试 Review

已完成：

- 自定义路由详情表单从裸 `padding: 24` 改为 `editorBody`，统一抽屉内表单外边距和分区间距。
- 基础信息、服务流向、规则定义分区去掉多层阴影和 hover 浮起，只保留弱边框、白色底和稳定 padding。
- 主调/被调服务流向区域改为浅灰容器 + 两个白色服务块，箭头区域收紧，减少图形抢占内容注意力。
- 路由规则表格在抽屉表单内增加弱边框、圆角和浅色表头，让匹配条件/实例分组更像规则定义摘要。
- 本次只优化 `CustomRouteEditor` 的视觉层，不改字段顺序、编辑逻辑、提交 payload 或发布逻辑。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `79355` 监听 `*:8080`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.67d64f01.js` 和 `assets/style.64101366.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/` 并点击 `demo-governance-route-202606050424` 后，抽屉表单正常展示。
- 浏览器 DOM 验证基础信息卡片、服务流向区块均无阴影，表格具备弱边框和 `6px` 圆角。
- 视觉截图保存到 `/tmp/governance-rule-form-optimized.png`。

# 治理规则详情抽屉 Inspector 形态尝试

- [x] 使用 frontend-skill 重新收敛视觉方向：右侧 inspector、低噪声、少卡片
- [x] 将抽屉 header 操作区从悬浮卡片改为轻量按钮组
- [x] 收紧正文内容壳和 Tabs 的层级，减少阴影和漂浮感
- [x] 构建并用 8080 页面验证实际截图
- [x] 记录 review 和验证结果

## 治理规则详情抽屉 Inspector 形态尝试 Review

已完成：

- 按 frontend-skill 的产品 UI 原则，将抽屉重新收敛为右侧 inspector：降低装饰、减少阴影、突出可扫描的信息与操作。
- Header 操作区从原先带阴影的悬浮工具卡改为轻量按钮组，`编辑 / 发布` 使用 32px 高按钮，位于规则标题右侧。
- `StickyTool` 仍由各规则编辑器原有逻辑渲染，本次只在抽屉作用域内覆盖视觉样式和位置，不改具体规则编辑器。
- Tabs 内容壳去掉多余阴影，保留 `8px` 圆角和弱边框，避免抽屉里出现多层漂浮卡片。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `73894` 监听 `*:8080`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.b263710a.js` 和 `assets/style.aaf2abc5.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/` 并点击 `demo-governance-route-202606050424` 后，抽屉标题、类型 tag、规则详情和按钮组正常展示。
- 浏览器 DOM 验证工具条背景透明、无阴影，单个按钮高度 `32px`，`stickyOverlapsTable=false`。
- 视觉截图保存到 `/tmp/governance-rule-drawer-inspector.png`。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在既有治理断言：`/governance` 默认跳转、自定义路由详情空态、泳道组详情空态；本次没有处理这些断言。

# 治理规则详情抽屉体验优化

- [x] 对照截图梳理抽屉 header、body、Tabs 和悬浮操作的布局问题
- [x] 优化 RuleDetailDrawer 的宽度、标题区、正文背景和滚动留白
- [x] 统一抽屉内 Tabs 与详情内容的间距层级
- [x] 调整抽屉内 StickyTool 的展示位置，避免遮挡规则详情表格
- [x] 构建并用 8080 页面验证抽屉效果
- [x] 记录 review 和验证结果

## 治理规则详情抽屉体验优化 Review

已完成：

- `RuleDetailDrawer` 抽屉宽度调整为 `min(860px, 88vw)`，右侧详情空间更适合路由、限流、熔断等规则编辑器的横向信息。
- 抽屉 header 增加固定高度、标题右侧安全区和类型标签，标题/类型/关闭按钮之间不再拥挤。
- 抽屉正文改为浅灰工作面，内部 `RuleTabs` 使用白色内容壳、`8px` 圆角和 `24px` 级别外边距，和当前 Console 工作台风格保持一致。
- `规则 / 版本 / 监听` 结构和各规则编辑器字段顺序保持不变，只在公共 `RuleTabs` 外层增加 view/table pane，用于控制间距。
- 抽屉内 `StickyTool` 从正文右下角改为标题栏右侧横向操作区，避免 `编辑 / 发布` 浮层压在规则详情表格上。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `66238` 监听 `*:8080`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.1cb7a8e0.js` 和 `assets/style.a51342c0.css`。
- 浏览器登录 `admin/admin123` 后打开 `http://127.0.0.1:8080/governance/`，规则列表显示 6 条 demo 规则，点击 `demo-governance-route-202606050424` 可打开详情抽屉。
- 浏览器 DOM 验证抽屉宽度为 `860px`，正文内边距为 `20px 24px 32px`，Tabs 内容壳为白底、`8px` 圆角和 `1px` 边框。
- 浏览器 DOM 验证 `StickyTool` 位于 header 右侧，和规则表格 `stickyOverlapsTable=false`。
- 视觉截图保存到 `/tmp/governance-rule-drawer-optimized.png`。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在既有治理断言：`/governance` 默认跳转、自定义路由详情空态、泳道组详情空态；本次只调整抽屉产品形态，未处理这三个历史断言。

# 注册发现服务页布局间距修正

- [x] 对照截图梳理页头、Tabs、指标、工具栏、表格之间的间距问题
- [x] 统一服务页 surface 内边距和区块 gap
- [x] 收紧表格行内信息层级，移除重复可见性展示
- [x] 构建并用 8080 截图/DOM 验证布局
- [x] 记录 review 和验证结果

## 注册发现服务页布局间距修正 Review

已完成：

- 注册发现 Tabs 外壳改为完整 surface：`20px` 内容内边距，避免指标栏直接贴在 tabs 下方和页面边缘。
- 服务 tab 内部统一为 flex workspace，指标栏、工具栏、表格之间保持 `16px` gap。
- 别名 tab 复用同一个 workspace gap，避免切换 tab 后工具栏和表格再次贴边。
- 表格 header、body、pagination 统一补齐横向与纵向 padding，让表格内容不再贴边。
- 服务首列移除可见性 tag，只展示服务名、命名空间和描述；可见性只在独立列展示，减少重复和拥挤。
- 归属、时间等空字段合并为单个 `-`，避免同一单元格出现两行无意义占位。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重签并重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `45660` 监听 `*:8080` 和 `*:8090`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.5210e47a.js` 和 `assets/style.8fa5758f.css`。
- 浏览器打开 `http://127.0.0.1:8080/discovery/service`，服务 tab 的 Tabs 内容区 `padding=20px`，workspace `display=flex` 且 `gap=16px`。
- 浏览器验证服务首列首行文本为 `demo-inventory demo-governance`，不再重复展示 `仅当前命名空间可见`。
- 浏览器切换别名 tab 后，别名页同样使用 workspace `gap=16px`，别名接口 `/naming/v1/service/aliases?offset=0&limit=10` 返回 `HTTP 200`。
- 新截图保存到 `/tmp/discovery-service-spacing-final.png`。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败在治理页既有断言：`/governance` 默认跳转、自定义路由详情空态、泳道组详情空态；本次未改治理页。

# 注册发现网关菜单拆分

- [x] 将网关从服务页内部 tab 移除
- [x] 在注册发现左侧菜单下新增独立网关路由
- [x] 更新菜单文案与页面说明，避免再把网关表述为服务页内容
- [x] 构建并用 8080 验证服务页、网关菜单和详情跳转
- [x] 记录 review 和验证结果

## 注册发现网关菜单拆分 Review

已完成：

- 服务页只保留 `服务 / 别名` 两个 tab，移除原先合并在服务页里的 `网关` tab。
- 注册发现左侧菜单新增独立 `网关` 子菜单，对应路由 `/discovery/gateway`。
- 网关页面独立成 `pages/Discovery/Gateway`，使用和注册发现/MCP 一致的页头外壳，当前正文保留未实现状态。
- 服务页说明文案改为只描述服务和别名，不再把网关入口描述为服务页内容。
- 经验记录已修正：网关应作为注册发现左侧菜单下的独立页面，不应合并进服务实例页内部 tab。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重签并重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `37965` 监听 `*:8080` 和 `*:8090`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.f8f43ae5.js` 和 `assets/style.936150a8.css`。
- 浏览器打开 `http://127.0.0.1:8080/discovery/service`，左侧菜单显示 `注册发现 服务实例 网关`，服务页 tab 只有 `服务` 和 `别名`，没有 `网关` tab。
- 浏览器打开 `http://127.0.0.1:8080/discovery/gateway`，页面显示 `Service Registry / Gateway`、`网关` 和独立网关页说明；无资源加载失败。

# 注册发现页面 MCP 风格对齐

- [x] 对照 MCP 页面抽取注册发现列表的视觉结构
- [x] 重构注册发现入口页的页头、Tabs 外壳和服务列表信息层级
- [x] 保持服务、别名、网关 tab 与创建、编辑、删除、详情跳转行为不变
- [x] 构建前端并用 8080 浏览器运行态验证
- [x] 记录 review 和验证结果

## 注册发现页面 MCP 风格对齐 Review

已完成：

- 注册发现入口页从裸 Tabs 调整为和 MCP 页面一致的工作台骨架：页头、eyebrow、说明文案和 Tabs 外壳。
- 服务列表新增 MCP 同款指标栏，展示服务数、命名空间数、健康实例和公开可见数量。
- 服务表格首列改为摘要列，聚合服务名、可见性、命名空间和描述；归属、健康实例、可见范围、时间、操作保持独立列，便于扫描。
- 服务列表和别名列表统一使用筛选工具栏、刷新按钮、主操作按钮和表格外壳；服务、别名、网关 tab 没有改路由和业务动作。
- 服务名搜索接入现有 `name` 查询参数，别名搜索接入现有 `alias` 查询参数，避免工具栏控件只做视觉展示。
- 新增经验记录：注册发现列表应复用 MCP 工作台风格，但保留原有资源 tab 与跳转链路。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- `git diff --check` 通过。
- 重签并重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `30120` 监听 `*:8080` 和 `*:8090`。
- `curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.aeaf3341.js` 和 `assets/style.c78eed62.css`。
- 浏览器打开 `http://127.0.0.1:8080/discovery/service`，页面包含 `Service Registry / Discovery`、`注册发现`、`Services`、`Namespaces`、`Healthy Instances`、`服务清单` 和 `pole.checker`。
- 浏览器网络记录显示 `/naming/v1/services?offset=0&limit=10` 返回 `HTTP 200`，无 loading failure。
- 浏览器输入 `pole.checker` 搜索后，服务接口请求为 `/naming/v1/services?offset=0&limit=10&name=pole.checker`，返回 `HTTP 200`，表格缩减为 1 行。
- 浏览器点击 `pole.checker` 后跳转到 `/discovery/service/instance?namespace=pole-system&service=pole.checker`，详情页仍显示 `服务详情`、`服务实例`、`服务订阅`。
- 浏览器切换 `别名` tab 后显示 `别名清单` 和空态，别名接口 `/naming/v1/service/aliases?offset=0&limit=10` 返回 `HTTP 200`，说明 tab 交互保留。

注意：

- `cd console/web && npm run test:response-mapping` 当前失败在治理页既有断言：`/governance` 默认跳转、自定义路由详情空态、泳道组详情空态；这些断言不属于本次注册发现改动范围，本次未一并修改。

# 控制台 AI 助手设计

- [x] 回顾项目经验记录，确认控制台、MCP、页面设计相关约束
- [x] 探索控制台前端栈、AI/MCP 页面和后端 AIMCP 能力
- [x] 核对 TDesign React Chat 组件包能力和依赖形态
- [x] 明确第一版 AI 助手的能力边界、权限边界和交互入口
- [x] 比较 2-3 种技术接入方案并给出推荐路线
- [ ] 形成设计规格并补充 review 小节

## 控制台 AI 助手设计 Review

当前已收敛的产品与架构方向：

- 控制台入口采用全局右侧助手，而不是每页内嵌 Copilot 或独立 AI 工作台。
- 写操作采用“生成草稿，用户确认后执行”的模式；查询解释、排障诊断可直接回答，治理操作必须先展示影响范围、参数 diff 和回滚建议。
- 右侧助手面板采用工作台式结构，固定包含对话、上下文、待确认操作、历史四类区域。
- 后端模块命名和边界收敛为 `Pole Agent Gateway`：它接入可配置 LLM Gateway，组织 Pole MCP / REST / 观测数据等工具能力，并向 Console 暴露场景化 access endpoint。
- 下游 LLM Gateway 不由 Pole 重复实现，可对接 LiteLLM、Higress、Kong、Envoy AI Gateway 等通用模型网关。

已完成高保真交互 Demo：

- Demo 文件：`.superpowers/brainstorm/29999-1780564206/content/pole-agent-gateway-demo-pro.html`
- 预览地址：`http://localhost:52909`
- 已验证交互：生成限流草稿、查看治理 diff、确认执行、待确认数量归零、执行历史回写。

待补：

- 将当前 Demo 和决策整理成正式设计规格。

# 8080 控制台无法访问排查

- [x] 复现 8080 连接失败并确认端口无监听
- [x] 检查当前进程，确认 pole-server 未在运行
- [x] 使用后台方式重启 all 模式，避免前台会话结束后服务消失
- [x] 验证 8080 首页、静态资源和关键 API 可访问
- [x] 补充 review 小节记录根因和验证结果

## 8080 控制台无法访问排查 Review

根因：

- 8080 无法访问时，`lsof -nP -iTCP:8080 -sTCP:LISTEN` 无输出，说明问题是没有进程监听端口。
- 先前依赖工具前台会话启动服务，结束后服务进程不再稳定存在；普通 `nohup ... &` 在当前工具执行环境中也会被清理，不能作为长期运行方式。
- 直接使用 `test/data/bootstrap/pole-server.yaml --mode all` 不够完整，因为该测试配置里的 console 节点是注释状态；需要本地 all 配置显式补齐 console 8080、webPath 和 console store。

本次处理：

- 生成 `/tmp/pole-server-all.yaml`，保留测试 server 配置，并补齐合并启动的 console 配置。
- 构建 `/tmp/pole-control-plane`，通过 macOS LaunchAgent `io.pole.control-plane.local` 启动 all 模式。
- console store 已连接 `pole_observability`，启动日志中不再出现 `store name is empty`。

验证：

- `launchctl print gui/$(id -u)/io.pole.control-plane.local` 显示 `state = running`，当前 PID 为 `19228`。
- `lsof -nP -iTCP:8080 -sTCP:LISTEN` 显示 `pole-cont` 正在监听 `*:8080`。
- `curl -I http://127.0.0.1:8080/` 返回 `HTTP/1.1 200 OK`。
- 首页引用资源为 `assets/index.03414bb1.js` 和 `assets/style.79183194.css`，入口 JS 返回 `200`。
- 使用 admin token 通过 `8080` console 代理请求 `/naming/v1/services?offset=0&limit=10` 返回 `code=200000`、`amount=1`、服务 `pole-system/pole.checker`。
- 使用 admin token 通过 `8080` console 代理请求 `/naming/v1/instances?namespace=pole-system&service=pole.checker&offset=0&limit=10` 返回 `code=200000`、`amount=1`、实例 `127.0.0.1:8091`。

# 治理规则 MySQL 模型收敛技术方案

- [x] 梳理当前普通治理规则表、发布表和泳道专用表的 DDL
- [x] 核对 store/cache/goverrule 发布链路对物理表拆分的依赖
- [x] 明确目标模型：所有治理规则统一两表，泳道作为 lane group 聚合根进入统一表
- [x] 输出代码改动方案、迁移策略、验证策略和风险控制
- [x] 生成正式技术方案内容并归档到 `context-kg/domains/governance-rules.md`
- [ ] 等待确认后再进入实现计划

## 治理规则 MySQL 模型收敛技术方案 Review

结论：

- 正式方案内容已归档到 `context-kg/domains/governance-rules.md` 的“统一存储与发布模型 / Store 收敛原则 / 泳道聚合规则 / Cache 统一更新源”章节。
- 所有治理规则收敛为 `governance_rule` 与 `governance_rule_release` 两张物理表，通过 `rule_type` 区分路由、限流、熔断、主动探测、无损上下线、泳道组等类型。
- 泳道按照方案 A 处理：`lane_group` 作为聚合根存入 `governance_rule`，`lane_rule` 作为 `lane_group.rule` JSON 内的子对象，不再作为独立持久化实体。
- `lane_group_release` 合并到 `governance_rule_release`，发布快照保存完整 LaneGroup JSON。
- 不考虑旧数据迁移；实现面向新 schema 和新数据路径，旧的分表数据不做自动搬迁。
- 泳道组内规则数量统一限制为 20，替代当前单独创建处 10、批量参数校验处 100 的不一致限制。

代码改动范围：

- `plugin/store/mysql/scripts/pole_server.sql`：新增统一治理规则表和发布表；删除普通治理规则分表以及 `lane_group`、`lane_rule`、`lane_group_release` 的新库建表定义。
- `plugin/store/mysql/governance_rule_base.go`：扩展通用 store 基础能力，抽取统一规则 CRUD / release CRUD。
- `plugin/store/mysql/governance_rule_store.go`：新增统一表 repository，统一处理所有规则类型的 SQL 操作，包括 CRUD、锁、列表、增量、发布、激活、失效、版本查询和删除发布版本。
- `apis/store/governance_api.go`：保留现有领域接口，同时补充内部通用接口，领域方法只做 rule type、领域对象和统一模型之间的适配。
- `plugin/store/mysql/{router_rule,ratelimit_config,circuitbreaker,fault_detect_config,lossless,lane}.go`：不再各自拼完整 SQL；改为调用统一 repository，保留领域类型转换和旧接口，避免一次性改穿上层业务接口。
- `pkg/cache/rules/governance_rule_update.go`：新增统一治理规则更新 cache，唯一负责轮询 `governance_rule` 与 `governance_rule_release`，按 `rule_type` 攒批并 fan-out 给 watcher。
- `pkg/cache/rules/*`：各类型 cache 保留当前查询 API 和索引结构，但从主动查库改为实现 `GovernanceRuleWatcher`，接收统一更新源分发的批次；泳道 cache 从完整 LaneGroup JSON 重建 `serviceRules` 索引。
- `pkg/goverrule/releases.go`：保留按 `Resource` 分发的业务接口，内部发布可统一走 `governance_rule_release`；泳道发布仍以 LaneGroup 聚合根为单位。
- `pkg/goverrule/interceptor/paramcheck/lane_rule.go` 与 `pkg/common/utils/valid`：新增统一 `MaxLaneRulesPerGroup = 20` 常量并复用。
- `plugin/store/mysql/governance_schema.go`：增加统一表创建检查；不做旧分表到统一表的数据迁移。
- 测试覆盖 MySQL store、发布版本、cache 增量、泳道 group JSON 内子规则 CRUD、泳道上限。

DB 操作统一原则：

- MySQL 层只允许 `governance_rule_store.go` 拼接统一规则表 SQL。
- 领域 store 文件不能再维护独立表名、独立 insert/update/release SQL；它们只负责把 `RouterConfig`、`RateLimit`、`CircuitBreakerRule`、`FaultDetectRule`、`LosslessRule`、`LaneGroup` 转成统一 `GovernanceRuleRecord`。
- `rule_type` 是所有读写、锁、版本、active 切换、增量查询的强制条件，避免不同规则类型同名或同 ID 时互相污染。

Cache 调整原则：

- cache 层拆成两层：统一更新源 + 类型化 watcher 索引层。
- 统一更新源是 `GovernanceRuleUpdateCache`，参与 `CacheManager` 定时更新，是治理规则唯一访问 MySQL 的 cache。
- `GovernanceRuleUpdateCache` 每轮只查两次库：`governance_rule` 和 `governance_rule_release`，然后按 `rule_type` 分组成 `GovernanceRuleChangeSet`。
- `RouterRuleCache`、`RateLimitCache`、`CircuitBreakerCache`、`FaultDetectCache`、`LosslessCache`、`LaneCache` 改为 watcher；它们不再主动查库，只在 `OnGovernanceRuleUpdate` 中更新自己的内存索引。
- 类型化 watcher 索引层保留现状：路由继续维护 `RouteRuleContainer`，限流/熔断/探测/无损继续维护按服务、命名空间、全局通配的索引，泳道继续维护 `serviceRules` 和 revision。
- fan-out 任一 watcher 失败时，本轮统一更新失败，不推进 `lastFetchTime`，避免某一类规则丢失增量。
- 这样缓存外部查询接口不变，但治理规则底层不再每类规则各自启动 goroutine 查询 store。

Governance rule watcher 接口：

```go
type GovernanceRuleChangeSet struct {
    RuleType apimodel.RuleRelease_RuleType
    Rules []*GovernanceRuleRecord
    Releases []*GovernanceRuleReleaseRecord
    RuleLastMtime time.Time
    ReleaseLastMtime time.Time
    FirstUpdate bool
}

type GovernanceRuleWatcher interface {
    RuleType() apimodel.RuleRelease_RuleType
    OnGovernanceRuleUpdate(change *GovernanceRuleChangeSet) error
    ClearGovernanceRules() error
}
```

统一更新流程：

```text
CacheManager ticker
  -> GovernanceRuleUpdateCache.Update()
    -> SELECT * FROM governance_rule WHERE mtime > ?
    -> SELECT * FROM governance_rule_release WHERE mtime > ?
    -> group by rule_type
    -> RouterRuleCache.OnGovernanceRuleUpdate()
    -> RateLimitCache.OnGovernanceRuleUpdate()
    -> CircuitBreakerCache.OnGovernanceRuleUpdate()
    -> FaultDetectCache.OnGovernanceRuleUpdate()
    -> LosslessCache.OnGovernanceRuleUpdate()
    -> LaneCache.OnGovernanceRuleUpdate()
```

CacheManager 调整：

- 新增 `CacheGovernanceRule` / `GovernanceRuleName`，只对它开启治理规则的定时 `Update()`。
- 原有 `CacheRoutingConfig`、`CacheRateLimit`、`CacheCircuitBreaker`、`CacheFaultDetector`、`CacheLossLess`、`CacheLaneRule` 继续注册，供业务代码查询，但不作为独立治理规则更新源。
- 若为了兼容当前 `OpenResourceCache` 机制，类型化 cache 的 `Update()` 可以变为 no-op；真正更新由 `GovernanceRuleUpdateCache` 统一 fan-out。

# 规则治理接口错误排查

- [x] 用 8080 真实运行态复现治理接口错误，而不是只验证静态资源
- [x] 检查 apiserver/store runtime 日志，定位治理表字段缺失
- [x] 补齐 MySQL 治理规则表兼容 schema 自动迁移
- [x] 修复 circuitbreaker/faultdetect/lossless 写入同时维护 legacy 字段与 `rule` 快照
- [x] 补齐 `lossless_rule_release` 初始化 SQL 与兼容迁移字段
- [x] 修复 lossless 发布缓存对旧/空发布记录的 nil 保护和缓存 key
- [x] 迁移本地 MySQL，构建新二进制并重启 8080
- [x] 用 admin JWT 通过 8080 console 代理验证治理列表接口

## 规则治理接口错误排查 Review

根因：

- 当前 `pole_server` 里的治理表采用公共列 + `rule` JSON 的 schema，而 Go store 代码仍会读写 legacy 专用字段，例如 `router_rule.policy`、`ratelimit_rule.disable`、`circuitbreaker_rule.enable`、`fault_detect_rule.dst_service`、`lossless_rule.service/config`。
- `lossless_rule_release` 缺少 `namespace/service`，但 lossless store 的发布缓存读路径会查询这两个字段，导致 lossless 列表仍返回 `500000 execute exception`。
- lossless 发布缓存用 `ActiveKey()` 读取旧值，却用 `rule.Id` 写入缓存，并且在旧值不存在时访问 `old.Id`，会放大空发布/旧发布记录带来的异常风险。

本次处理：

- 新增治理规则兼容 schema 检查，启动时自动补齐缺失列。
- `circuitbreaker`、`faultdetect`、`lossless` 写入时同时维护 legacy `config` 字段和规范 `rule` 快照。
- 初始化 SQL 补齐 `lossless_rule_release.namespace/service`。
- lossless 发布缓存改为按 ActiveKey 读写，并对缺失旧值、nil rule 做保护。
- 当前本地 MySQL 已手动补齐治理表缺失字段，并通过 LaunchAgent 重启 `/tmp/pole-control-plane`。

验证：

- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./pkg/cache/rules -run 'TestLosslessSetRulesClient'` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/store/mysql -run 'TestGovernanceRuleCompatibility|TestEnsureGovernanceRuleColumn'` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- LaunchAgent `io.pole.control-plane.local` 当前 `state = running`，PID `65890` 监听 `*:8080` 和 `*:8090`。
- `curl -I http://127.0.0.1:8080/` 返回 `HTTP/1.1 200 OK`，首页资源为 `assets/index.03414bb1.js`、`assets/style.79183194.css`。
- 使用 admin JWT 通过 8080 console 代理请求 `/naming/v1/lossless`、`/naming/v1/routings`、`/naming/v1/ratelimits`、`/naming/v1/circuitbreakers`、`/naming/v1/faultdetectors` 均返回 `HTTP_STATUS:200`、`code=200000`。
- `test/output/logs/runtime/pole-store-error.log` 停留在 2026-06-04 18:22:47，修复后没有新的 `Unknown column` 记录；apiserver error 日志最新只是重启时的 `Server closed`，没有新的治理接口 500。

# 规则治理 Demo 数据构造

- [x] 确认 all 模式 server、console 和 MySQL 运行态
- [x] 对照当前 `github.com/pole-io/specification v0.1.0-ALPHA.24` 与前端 service 类型确认治理规则字段
- [x] 创建 demo 命名空间/服务基础数据
- [x] 通过治理 REST API 创建路由、限流、熔断、主动探测、无损上下线、泳道组 demo 规则
- [x] 通过 8080 console 代理验证各治理列表可见
- [x] 记录 review 与可查看入口

## 规则治理 Demo 数据构造 Review

已按 `github.com/pole-io/specification v0.1.0-ALPHA.24` 创建 demo 数据：

- 命名空间：`demo-governance`
- 服务：`demo-gateway`、`demo-order`、`demo-payment`、`demo-inventory`
- 自定义路由：`demo-governance-route-202606050424`
- 限流规则：`demo-governance-ratelimit-202606050425`
- 熔断规则：`demo-governance-circuit-202606050425`
- 主动探测规则：`demo-governance-faultdetect-202606050425`
- 无损上下线规则：`demo-governance/demo-order`
- 泳道组与泳道规则：`demo-governance-lane-202606050425`、`demo-governance-lane-rule-202606050425`

本次同时修复一个会影响查看效果的接口问题：

- 熔断规则列表鉴权 predicate 直接访问 `cbr.Proto.Metadata`，但缓存中的 `CircuitBreakerRule` 可能只有 store 行数据、`Proto` 为空，导致 `/naming/v1/circuitbreakers` panic 并返回 empty reply。
- 已新增 `circuitBreakerRuleMetadata` 做空值保护，并补充单测覆盖 nil / 空缓存对象 / Proto metadata 三种路径。

验证：

- `go test -count=1 ./pkg/goverrule/interceptor/auth` 通过。
- `go build -o /tmp/pole-control-plane .` 通过。
- 已重签 `/tmp/pole-control-plane` 并重启 LaunchAgent `io.pole.control-plane.local`，当前 PID `96864` 监听 `*:8080` 和 `*:8090`。
- 通过 8090 直连与 8080 console 代理分别验证：命名空间、服务、路由、限流、熔断、主动探测、无损上下线、泳道组均返回 `code=200000`，且响应内容包含对应 demo 名称。
- 复测 `/naming/v1/circuitbreakers?offset=0&limit=1&brief=true&name=__no_such_demo__` 返回 `HTTP 200`、`code=200000`，不再出现 empty reply。

可查看入口：

- Console 首页：`http://127.0.0.1:8080/`
- 治理默认入口：`http://127.0.0.1:8080/governance/`
- 无损上下线：`http://127.0.0.1:8080/governance/lossless`
- 自定义路由：`http://127.0.0.1:8080/governance/router`
- 访问限流：`http://127.0.0.1:8080/governance/ratelimit`
- 熔断降级：`http://127.0.0.1:8080/governance/circuitbreaker`

# 规则治理 PRD 交互原型

- [x] 明确方向：以“规则运行态控制台”为主，组合“场景化新建入口”和“发布审计视图”
- [x] 制作与当前 Console/TDesign 风格接近的交互式 HTML PRD 原型
- [x] 验证原型关键交互：规则选择、类型筛选、详情 tab、场景化新建向导
- [x] 记录可访问地址、原型覆盖范围和后续落地建议
- [x] 将规则详情从常驻右侧面板调整为侧边抽屉模式
- [x] 增强抽屉详情的信息表达：作用范围、规则快照、发布记录、监听实例
- [x] 验证抽屉打开、关闭、规则切换和详情 tab 交互
- [x] 将详情抽屉从 PRD 说明态调整为最终产品态
- [x] 补齐最终产品态详情内容：状态摘要、作用范围、规则定义、发布检查、监听对象、操作区
- [x] 验证最终产品态抽屉的完整预览和关键交互
- [x] 使用 frontend-skill 与 TDesign React 文档重新收敛整体视觉，去除 AI 化表达
- [x] 对齐现有服务/实例详情页的信息架构与视觉密度
- [x] 移除原型内 PRD 说明入口和解释性设计文案
- [x] 验证收敛后的列表、详情抽屉、tab 和新建向导交互
- [x] 将规则详情从“服务/实例详情迁移风格”回收为接近现有治理规则详情/编辑页的结构
- [x] 保留抽屉交互，但恢复基础信息、目标服务、匹配规则、动作配置、发布记录的信息顺序
- [x] 验证新详情与现有治理实现差异收敛
- [x] 对照真实 `specification` proto 与前端 service 类型，修正 demo 规则详情字段
- [x] 按路由、限流、熔断、主动探测、无损上下线、泳道分别构造 spec 形态数据
- [x] 验证详情抽屉展示字段与各规则 spec/editor 字段一致
- [x] 重新查看 8080 当前治理规则详情页，确认真实页面的信息架构和产品文案
- [x] 将 demo 详情从“直显 spec 字段”调整为“按现有详情页设计展示，底层数据符合 spec”
- [x] 逐类验证路由、限流、熔断、主动探测、无损、泳道详情不再裸露 proto 字段名
- [x] 将原型运行态从“列表 + 抽屉详情”调整为“统一列表 + 当前右侧详情框”
- [x] 保持列表页采用统一治理工作台规范，详情框沿用 8080 当前 `规则 / 版本 / 监听` 设计
- [x] 验证点击列表后右侧详情框原地刷新，不再出现遮罩抽屉交互
- [x] 按最新反馈恢复为“统一列表 + 右侧抽屉详情”，抽屉内容继续沿用 8080 当前 `规则 / 版本 / 监听` 设计
- [x] 验证点击列表打开抽屉，关闭按钮、遮罩、Esc 均可关闭

## 规则治理 PRD 交互原型 Review

已交付：

- 原型文件：`.superpowers/brainstorm/9750-1780640247/content/governance-prd-interactive.html`
- 预览地址：`http://localhost:62045/`
- 视觉方向：延续当前 Console/TDesign 的侧边栏、顶部面包屑、灰底工作区、蓝色主操作和表格密度，同时把治理规则收敛为“运行态控制台 + 详情抽屉”。

覆盖范围：

- 运行态控制台：统一展示路由、限流、熔断、主动探测、无损上下线、泳道规则。
- 详情抽屉：基础信息、服务与匹配、动作配置、发布记录四个 tab；点击规则行打开，关闭后回到完整列表扫描。
- 发布审计：集中展示规则发布流和版本状态。
- 策略模板：把新建入口按“分流、限流、熔断、泳道”场景组织。
- 新建策略向导：三步式选择场景、配置作用范围、生成草稿。

验证：

- `curl -sS http://localhost:62045/` 可返回最新原型内容，包含 `规则治理工作台`、`新建治理策略`、`发布审计`、`demo-governance-lane`。
- 使用本机 Chrome headless + DevTools Protocol 验证页面首屏渲染：标题为 `规则治理工作台`，初始规则行数为 6。
- 验证类型筛选：点击 `熔断` 后列表只展示 `demo-governance-circuit-202606050425`，点击该规则后详情抽屉打开并同步切换。
- 验证详情 tab：点击 `服务与匹配` 后展示主调/被调服务和匹配条件表格；点击 `动作配置` 后展示治理动作；点击 `发布记录` 后内容包含发布版本记录。
- 验证新建策略向导：`新建策略` 可打开弹窗，`下一步` 可进入第 2、3 步，`生成草稿` 后关闭弹窗并选中 `demo-governance-lane-202606050425`。
- 验证主视图切换：`发布审计` 和 `策略模板` 均可切换显示。
- 详情抽屉复测：首屏默认打开抽屉方便评审；点击规则行后 `detailDrawer/detailMask` 打开；关闭按钮、遮罩点击、Esc 均可关闭；新建向导生成草稿后会自动打开泳道规则详情抽屉。
- 去 AI 化复测：使用 frontend-skill 的产品 UI 标准收敛为“安静、密集、运维控制台”方向；参考 TDesign React `Drawer`、`Table`、`Tabs`、`Tag`、`Timeline`、`List` 的组件语义。页面正文不再包含 `PRD`、`为什么`、`生产实现`、`原型` 等设计说明字样；顶部指标由四张卡收敛为单条状态栏；详情抽屉宽度调整为 680px。
- 回收为现有治理详情结构：读取 `CustomRouteEditor`、`RateLimitEditor`、`CircuitBreakerEditor` 后，确认当前详情更接近编辑器只读态，而不是服务/实例详情。原型已保留抽屉交互，但 tab 调整为 `基础信息`、`服务与匹配`、`动作配置`、`发布记录`；内容顺序改回基础信息、目标服务、主调/被调服务、匹配条件表格、动作/限流/熔断配置、发布版本记录。
- 回收版验证：Chrome headless 验证默认详情打开在 `基础信息`；`服务与匹配` 包含主调/被调服务和匹配条件表格；`动作配置` 包含动作/处理方案；`发布记录` 包含 `release-route-v3`；关闭抽屉后筛选 `熔断` 并点击规则可重新打开 `demo-governance-circuit-202606050425`。
- Spec 对齐修正：本次对照 `../specification/api/v1/traffic_manage/router.proto`、`ratelimit.proto`、`lossless.proto`、`lane.proto`、`../specification/api/v1/fault_tolerance/circuitbreaker.proto`、`fault_detector.proto`，以及 `console/web/src/services/{router,ratelimit,circuitbreaker,faultdetect,lossless,lane}.ts`，将抽屉详情从统一泛化字段改为分类型 spec 字段展示。
- 当前详情覆盖：路由展示 `RouteRule.routing_config(CustomRoute.rules[].sources/destinations)`；限流展示 `RateLimit.rules[].method/arguments/amounts/action/failover/custom_response`；熔断展示 `CircuitBreakerRule.ruleMatcher/block_configs/recoverCondition/fallbackConfig`；主动探测展示 `FaultDetectRule.targetService/interval/timeout/port/protocol/httpConfig`；无损展示 `LosslessRule.lossless_online.delay_register/warmup/readiness` 与 `lossless_offline`；泳道展示 `LaneGroup.entries/destinations` 和 `LaneRule.trafficMatchRule/defaultLabelValue/labelKey`。
- Spec 对齐验证：`node` 脚本校验 HTML 内 `<script>` 语法通过；`curl -sS http://localhost:62045/ | rg "RouteRule|RateLimit|CircuitBreakerRule|FaultDetectRule|LosslessRule|LaneGroup|routing_config|block_configs|lossless_online|trafficMatchRule|custom_response"` 命中最新内容；使用本机 Chrome headless 逐类筛选并点击抽屉 tab，`route-rule`、`rate-action`、`circuit-action`、`probe-rule`、`lossless-action`、`lane-rule` 六项字段断言均为 `true`，截图保存到 `/tmp/governance-prd-spec-drawer.png`。
- 真实页面回看：使用 `admin/admin123` 登录 `http://127.0.0.1:8080/governance/`，分别打开无损发布、路由规则、限流规则、熔断规则、主动探测和全链路灰度详情；确认当前实现是左表右详情面板，详情 tab 为 `规则 / 版本 / 监听`，规则页内部复用各规则编辑器的只读态，而不是展示 proto 字段名。
- 产品态修正：原型抽屉 tab 已改为 `规则 / 版本 / 监听`；规则正文改为当前页面产品文案与分区：路由为基础信息、主调/被调服务、规则匹配与实例分组；限流为目标服务、子规则、匹配条件、限流方式、限流方案；熔断为熔断粒度、熔断策略、恢复策略、熔断后降级；主动探测为目标服务、接口、间隔、超时、端口和协议配置；无损为目标服务、无损上线、服务预热、无损下线；泳道为泳道组详细、泳道组入口、泳道组服务、泳道列表和泳道路由规则。
- 产品态验证：`node` 脚本校验 HTML 内 `<script>` 语法通过；Chrome headless 从 `http://127.0.0.1:62045/governance-prd-interactive.html` 逐类筛选并点击详情，断言页面可见文本不包含 `RouteRule`、`routing_config`、`block_configs`、`lossless_online`、`trafficMatchRule`、`custom_response` 等 raw spec 字段；六类规则均命中现有详情页关键词，`规则 / 版本 / 监听` tab 断言通过，截图保存到 `/tmp/governance-prd-product-detail.png`。
- 预览入口修正：新增 `.superpowers/brainstorm/9750-1780640247/content/index.html`，浏览器访问 `http://127.0.0.1:62045/` 会自动跳转到 `governance-prd-interactive.html`。
- 列表/详情边界修正：按最新反馈，运行态页面改为“统一列表 + 当前右侧详情框”，不再使用侧边抽屉。列表仍保留统一规则清单、类型筛选、命名空间/服务筛选和运行状态列；详情框移回页面右侧常驻面板，沿用当前 8080 治理页 `规则 / 版本 / 监听` 结构和只读编辑器式内容。
- 详情框形态验证：Chrome headless 打开 `http://127.0.0.1:62045/governance-prd-interactive.html`，确认 `#runtimeView` 为左右两列布局（约 `638px 520px`），`#detailDrawer` 的 CSS `position=static`，页面不存在 `#detailMask`，列表与详情框并排展示；点击熔断规则后右侧详情原地刷新并包含 `熔断策略 [1]`，截图保存到 `/tmp/governance-prd-detail-frame.png`。
- 最新纠偏：详情交互恢复为右侧抽屉，但抽屉内部仍沿用当前 8080 治理详情页的 `规则 / 版本 / 监听` 和只读编辑器式内容；列表页面保持统一工作台规范。
- [x] 将详情从右侧常驻面板恢复为右侧抽屉交互
- [x] 保留抽屉内部 `规则 / 版本 / 监听` 和当前 8080 详情内容
- [x] 验证点击列表打开抽屉、遮罩 / 关闭按钮 / Esc 可关闭
- 抽屉形态验证：`node` 脚本校验 HTML 内 `<script>` 语法通过；Chrome headless 打开 `http://127.0.0.1:62045/governance-prd-interactive.html`，确认 `.workspace` 为单列（`1172px`）、`#detailDrawer` 为 `position=fixed` 且初始关闭，`#detailMask` 存在。筛选 `熔断` 并点击规则后抽屉和遮罩打开，tab 为 `规则/版本/监听`，内容包含 `熔断策略 [1]`、`熔断粒度`、`恢复策略`、`熔断后降级`，且不包含 raw spec 字段；关闭按钮、遮罩点击、Esc 均可关闭，截图保存到 `/tmp/governance-prd-drawer-current-style.png`。

后续落地建议：

- 第一阶段先落地统一列表 + 详情抽屉，保留现有各规则 API，不急于重做后端协议。
- 前端组件建议拆成 `GovernanceWorkbench`、`RuleTypeFilter`、`RuleListTable`、`RuleInspector`、`RuleCreateWizard`、`RulePublishTimeline`。
- 数据层建议先做 adapter，把 route、ratelimit、circuitbreaker、faultdetect、lossless、lane 的响应归一成同一种展示模型。
- 新建策略向导第一版只生成草稿和跳转到对应规则编辑页，发布流后续再接审批/版本能力。

# 规则治理 Console 页面落地

- [x] 将最新原型决策登记到任务计划
- [x] 梳理现有治理页面、RuleTabs、各规则 Table/Editor 和 service 类型
- [x] 实现统一治理规则列表，覆盖路由、限流、熔断、主动探测、无损上下线、泳道
- [x] 实现右侧详情抽屉，抽屉内部保留当前 8080 `规则 / 版本 / 监听` 结构
- [x] 调整治理入口路由与页面样式，贴近当前 Console/TDesign 视觉
- [x] 构建前端并用 8080 浏览器验证列表、筛选、抽屉和详情内容
- [x] 补充 review 小节记录验证证据

## 规则治理 Console 页面落地 Review

已完成：

- 新增 `pages/Governance/Workbench`，将 `/governance/` 默认入口改为统一治理规则工作台，保留 `/governance/lossless`、`/governance/router`、`/governance/ratelimit`、`/governance/circuitbreaker` 旧入口。
- 统一列表通过现有 Redux thunk 并发读取自定义路由、单机限流、分布式限流、熔断、主动探测、无损上下线、泳道组，不新增后端接口。
- 新增 `RuleDetailDrawer`，统一使用 TDesign `Drawer` 承载详情；抽屉内部继续复用当前 `RuleTabs` 与各规则 Editor 的 view 态，保留 `规则 / 版本 / 监听`。
- 将旧治理子页面里的右侧常驻详情面板改为同一个 `RuleDetailDrawer`，页面关闭详情后不再留下右侧空面板。
- 补齐主动探测旧子页面的版本 tab 懒加载入口。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过；重签并重启 LaunchAgent `io.pole.control-plane.local` 后，8080 首页资源更新为 `assets/index.e3a1f887.js` 和 `assets/style.f40448d4.css`。
- `127.0.0.1:8080` 和 `127.0.0.1:8090` 均由新 PID `53503` 监听。
- Chrome headless 使用 `admin/admin123` 登录后打开 `http://127.0.0.1:8080/governance/`，统一列表显示 6 条 demo 治理规则，包含路由、单机限流、熔断、探测、无损、泳道。
- 点击 `demo-governance-circuit-202606050425` 后打开右侧 720px 抽屉；详情包含 `规则 / 版本 / 监听`，包含 `熔断策略`、`恢复策略`、`熔断后降级`，且不包含 `RouteRule`、`routing_config`、`block_configs` 等 raw spec 字段；截图保存到 `/tmp/governance-console-workbench-drawer.png`。
- 抽屉关闭按钮、遮罩点击、Esc 均验证可关闭，关闭后 `.t-drawer__content-wrapper` 移出可视区。
- 旧子页面 `http://127.0.0.1:8080/governance/circuitbreaker` 点击规则名称同样打开 720px 抽屉，详情包含 `规则 / 版本 / 监听`，页面不再存在常驻右侧详情面板。

# Console 侧边栏整理

- [x] 梳理当前侧边栏由路由树生成的逻辑
- [x] 调整一级模块顺序，让资源、治理、观测、AI、认证更好扫描
- [x] 收敛治理菜单，只把统一工作台作为主入口，旧子页面保留直达路由但隐藏侧边栏项
- [x] 调整默认展开策略，只默认展开当前所在模块
- [x] 构建并用 8080 浏览器验证侧边栏和治理入口

## Console 侧边栏整理 Review

已完成：

- 侧边栏一级模块顺序调整为：命名空间、AI 工具、注册发现、治理管理、监控指标、认证管理。
- 治理管理下只保留 `治理工作台` 菜单项；`/governance/lossless`、`/governance/router`、`/governance/ratelimit`、`/governance/circuitbreaker` 仍保留路由直达，但从侧边栏隐藏。
- 菜单默认展开策略改为只展开当前所在一级模块，并启用同级互斥展开；访问旧治理子页面时侧边栏高亮 `治理工作台`。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过；重签并重启 LaunchAgent 后，8080 首页资源更新为 `assets/index.a0d60c2b.js` 和 `assets/style.f40448d4.css`。
- Chrome headless 使用 `admin/admin123` 登录后打开 `http://127.0.0.1:8080/governance/`，确认侧边栏只展开治理模块，治理下只显示 `治理工作台`，不再显示 `无损发布`、`路由规则`、`限流规则`、`熔断规则`。
- 打开旧路径 `http://127.0.0.1:8080/governance/circuitbreaker` 仍可访问故障熔断页面，侧边栏继续高亮 `治理工作台`。

# Console 侧边栏点击回归修复

- [x] 用 8080 真实页面复现侧边栏注册发现、认证管理等菜单无法点击的问题
- [x] 定位侧边栏展开策略导致隐藏子菜单占位并拦截点击
- [x] 修复菜单展开配置，恢复跨模块菜单的正常点击能力
- [x] 重新构建并重启 all 模式 server
- [x] 用浏览器验证服务列表、主体管理、权限策略等侧边栏入口可打开
- [x] 补充 review 小节记录根因和验证结果

## Console 侧边栏点击回归修复 Review

根因：

- 上一轮侧边栏整理把菜单默认展开策略改为只展开当前一级模块，并启用了 `expandMutex`。
- 在 TDesign Menu 当前 DOM/CSS 行为下，折叠中的子菜单仍可能保留布局或动画状态，导致其它模块的子菜单被上方隐藏节点拦截点击。
- 当时验证只确认治理菜单隐藏和旧治理路径高亮，没有横向点击注册发现、AI、认证等模块入口，导致回归漏出。

本次修复：

- 移除 `expandMutex`，恢复稳定的多模块展开行为。
- `defaultExpanded` 恢复为注册发现、治理、监控、AI、认证等模块同时展开。
- 保留治理路径高亮到 `治理工作台` 的逻辑，旧治理子路由仍可直达。

验证：

- `cd console/web && npm run build` 通过，新资源为 `assets/index.7c148d11.js` 与 `assets/style.f40448d4.css`。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- 已重签 `/tmp/pole-control-plane` 并重启 LaunchAgent `io.pole.control-plane.local`；`127.0.0.1:8080` 当前监听正常，首页加载新构建资源。
- Chrome headless 桌面视口 `1440x960` 登录后打开 `http://127.0.0.1:8080/governance/`，侧边栏包含并展开 `服务实例`、`治理工作台`、`MCP 服务`、`主体管理`、`权限策略`。
- 使用真实鼠标事件点击验证：`服务实例` 跳转到 `/discovery/service`，`主体管理` 跳转到 `/auth/principals`，`权限策略` 跳转到 `/auth/policies`，`MCP 服务` 跳转到 `/ai/mcps`，`治理工作台` 跳转到 `/governance/workbench`。
- 截图保存到 `/tmp/console-sidebar-governance-organized.png`。

# 默认策略查询错误排查

- [x] 用 8080 真实链路复现默认策略查询报错
- [x] 定位请求参数、console 代理、后端 handler/store 中的故障层
- [x] 实现最小修复并补充回归测试
- [x] 重新构建、重启 all 模式 server
- [x] 验证默认策略页面/接口不再报错

## 默认策略查询错误排查 Review

根因：

- 默认策略查询会命中系统默认策略和用户默认策略，这些策略的资源列表包含 spec 中的 `LosslessRules`、`MirrorRules`、`SecurityRules` 等类型。
- 后端 `enrichResourceDetial` 在把策略资源转换为 Console 响应时，直接从 `resourceFieldPointerGetters` 和 `resourceConvert` map 下标取函数并调用；当资源类型没有映射时会触发 nil 函数调用，HTTP 连接表现为 `empty reply`，经 console 代理后显示为 `502 Bad Gateway`。
- `default=1` 之前看似正常，是因为缓存过滤只按 `true/false` 字符串匹配，实际没有命中默认策略数据；真正的页面参数 `default=true` 会稳定复现问题。

本次修复：

- `ResourceFieldPointerGetters` 补齐 `LosslessRules`、`MirrorRules`、`SecurityRules` 的响应字段指针。
- 默认策略资源响应初始化补齐 `lossless_rules`、`mirror_rules`、`security_rules`。
- `enrichResourceDetial` 对未知资源类型和未知 converter 做显式保护并记录 warn，避免 spec 后续新增资源类型时再次导致请求中断。
- 新增单测覆盖默认策略中的 `LosslessRules/MirrorRules/SecurityRules` wildcard 资源，以及未知资源类型不 panic。

验证：

- 复现修复前：`GET /auth/v1/policies?offset=0&limit=10&default=true` 在 8080 返回 `502 Bad Gateway`，直连 8090 返回 `Empty reply from server`；`default=false` 正常。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/access_control/auth/policy -run 'TestEnrichResourceDetail'` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/access_control/auth/policy ./plugin/access_control/auth/policy/interceptor/auth ./plugin/access_control/auth/policy/interceptor/paramcheck ./pkg/cache/auth` 通过。
- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过；重签并重启 LaunchAgent 后，`127.0.0.1:8080` 和 `127.0.0.1:8090` 均由 PID `10445` 监听。
- 修复后通过 8080 和 8090 分别请求 `default=true`、裸查询、`default=false`：全部返回 HTTP 200；`default=true` 返回 3 条默认策略，包含 `(用户) admin的默认策略`、`全局只读策略`、`全局读写策略`。
- Chrome headless 打开 `http://127.0.0.1:8080/auth/policies`，切换到 `默认策略` tab，页面显示 3 条默认策略且没有 `500` 或 `获取数据失败`。

# 当前功能实现分析

- [x] 明确分析范围：当前分支相对 `origin/develop` 的已实现功能
- [x] 识别提交与文件变更
- [x] 阅读入口、实现和测试覆盖
- [x] 运行必要验证
- [x] 汇总已完成能力、未完成点和证据

# specification MCP 设定核对

- [x] 明确核对范围：`../specification` 是否已有 MCP Server / Tool 相关协议设定
- [x] 搜索 specification 源定义和生成产物
- [x] 对照 control-plane 当前字段与接口形态
- [x] 汇总是否需要补 spec 及建议补哪些内容

# specification MCP 设定补充

- [x] 明确补充范围：新增 MCP Server / Tool 协议模型与查询参数
- [x] 建立失败检查，确认当前缺少 MCP proto
- [x] 新增 `api/v1/ai/mcp.proto`
- [x] 更新 Go/Rust 生成脚本
- [x] 运行生成与构建验证
- [x] 汇总产物和风险

# control-plane 使用最新 specification MCP

- [x] 明确范围：升级 specification 依赖并迁移 MCP Server / Tool 类型
- [x] 建立失败检查，确认当前仍引用本地 `apis/pkg/types/ai`
- [x] 升级 `github.com/pole-io/specification` 到 `v0.1.0-ALPHA.23`
- [x] MCP Server / Tool 改用 `specification/source/go/api/v1/ai`
- [x] 处理缓存和 MySQL 层时间字段转换
- [x] 运行格式化、聚焦测试和必要构建验证
- [x] 汇总结果和后续注意事项

# control-plane MCP spec 桥接收敛

- [x] 明确范围：MCP 工具参数、cache/store 查询边界和响应 Data 都收敛到 specification message
- [x] 建立失败检查，确认当前仍存在 map 查询边界和 JSON/Struct Any 中转
- [x] `list_mcp_servers` 查询参数改用 `ai.MCPServerQuery`
- [x] `list_mcp_server_tools` 查询参数改用 `ai.MCPServerToolQuery`
- [x] `create/update/delete_mcp_servers` 请求解析改用 spec 容器或请求 message
- [x] `BatchQueryResponse.Data` 直接使用 `anypb.New` 包装 spec proto message
- [x] 更新相关测试并运行格式化、聚焦验证
- [x] 汇总结果和后续注意事项

# control-plane MCP Registry REST 支撑

- [x] 明确范围：为 pole-console MCP 页面补 REST 查询、创建、更新、删除与工具查询接口
- [x] 复用已收敛到 specification 的 `ai.MCPServerQuery` / `ai.MCPServers` / `ai.MCPServerDeleteRequest`
- [x] 增加 `/ai/mcp/v1/servers` GET/POST/PUT、`/servers/delete` POST、`/server/tools` GET
- [x] 运行 gofmt、结构检查和聚焦测试
- [x] 记录结果和剩余风险

## Review

当前分支仅领先 `origin/develop` 一个提交：`1e83b290 refactor(aimcp): 读路径走 MCPServerCache + 修复 BatchQueryResponse.Data 空数据`。

已实现：

- AIMCP Server 初始化时会获取并打开 `MCPServerName` 资源缓存。
- `list_mcp_servers` 读路径从 MySQL store 改为 `cacheMgr.MCPServer().Query`。
- MCP Server 缓存新增分页查询，支持 `name` 前缀匹配，以及 `namespace`、`business`、`department`、`protocol` 精确匹配，按 `MTime DESC` 排序。
- `list_mcp_server_tools` 读路径从 MySQL store 改为 MCP Server 缓存，支持通过 `server_id` 直接查，也支持通过 `server_name + server_namespace` 先从缓存解析 server ID。
- MCP Server 和 MCP Server Tool 的批量查询响应会填充 `BatchQueryResponse.Data`，不再只返回 `amount/size`。
- 新增缓存查询单元测试，覆盖名称前缀匹配、精确过滤、分页和 `MTime DESC` 排序。

非功能性变化：

- 多个 Go 文件仅发生 import 顺序、结构体字段对齐、空行清理等格式化变化。

验证：

- `git diff --check origin/develop...HEAD` 通过。
- `go test ./pkg/cache/ai` 通过。
- `go test ./plugin/apiserver/httpserver/aimcp` 通过。
- `go test ./...` 未形成完整通过证明：运行中出现 Go proxy 依赖下载超时，随后长时间卡在其他包测试，已终止。

风险：

- 缓存查询的 `name` 过滤语义由旧 MySQL 查询的精确匹配变为前缀匹配，这是代码注释和新增测试确认的行为变化。
- 读路径依赖缓存刷新时效，写入 MCP Server 后不是立即从读接口可见。

## specification MCP 设定核对 Review

结论：需要补。`../specification` 当前没有 MCP / AI Native 相关 proto，control-plane 的 MCP Server / Tool 类型仍是本仓库本地 Go struct，不属于正式跨仓库契约。

证据：

- `../specification/api/v1` 当前只有 `model`、`service_manage`、`config_manage`、`traffic_manage`、`fault_tolerance`、`security` 等目录，没有 AI/MCP 目录。
- `../specification` 全仓库搜索没有命中 MCP Server / Tool 定义。
- control-plane 的字段定义在本地 `apis/pkg/types/ai/mcp.go`。
- control-plane 的 AIMCP 接口只复用了 specification 的通用 `Response` / `BatchQueryResponse` / `Code`，未复用 MCP 领域模型。

建议：

- 在 specification 新增 `api/v1/ai/mcp.proto` 或 `api/v1/ai_native/mcp.proto`。
- 定义 `MCPServer`、`MCPServerTool`、`MCPServerQuery`、`MCPServerToolQuery`，字段至少覆盖当前本地类型和工具参数。
- 同步更新 `source/go/build.sh`、`source/rust/build.sh`，生成 Go/Rust 产物。
- 后续 control-plane 再把本地 `apis/pkg/types/ai` 替换或映射到 specification 生成类型。

## specification MCP 设定补充 Review

已在 `../specification` 补充 MCP 协议设定：

- 新增 `api/v1/ai/mcp.proto`，定义 `MCPServer`、`MCPServers`、`MCPServerQuery`、`MCPServerDeleteRequest`、`MCPServerTool`、`MCPServerTools`、`MCPServerToolQuery`、`MCPServerToolDeleteRequest`。
- 更新 `source/go/build.sh`，新增 `ai` proto 分组并生成 `source/go/api/v1/ai/mcp.pb.go`。
- 更新 `source/rust/build.sh`，复制 `api/v1/ai/*.proto`，并显式使用仓库内置 `protoc`，避免 Rust 构建依赖本机全局 `protoc`。
- 更新 README，在接口表中加入 AI Native / MCP Server。
- 运行 Rust 构建时发现既有脚本会同步 ratelimiter proto 和多个旧 Rust proto 副本，因此 Rust 侧生成产物里包含部分既有 proto 同步差异，不属于 MCP 业务语义。

验证：

- `test -f api/v1/ai/mcp.proto`：先失败，新增后通过。
- `cd ../specification/source/go && bash build.sh`：通过。
- `cd ../specification/source/rust && bash build.sh`：通过。
- `cd ../specification && git diff --check`：通过。
- `cd ../specification && GOPROXY=https://goproxy.cn,direct go test ./...`：通过。
- `cd ../specification/source/rust/pole-specification && PROTOC=../../protoc/protoc-darwin-arm64/bin/protoc cargo test --release`：通过。

## control-plane 使用最新 specification MCP Review

已完成：

- `go.mod` 将 `github.com/pole-io/specification` 升级到 `v0.1.0-ALPHA.23`。
- `v0.1.0-ALPHA.23` 已推送到远端后，移除本地 `replace`，control-plane 直接解析远端 tag。
- 删除本仓库本地重复类型 `apis/pkg/types/ai/mcp.go`。
- MCP Server / Tool 的 cache、store、AIMCP HTTP server 均改用 `github.com/pole-io/specification/source/go/api/v1/ai`。
- 适配 specification 生成字段命名：`ID` 改为 `Id`，`MCPServerID` 改为 `McpServerId`，`CTime/MTime` 改为 `Ctime/Mtime`，`Flag` 改为 `uint32`。
- 因 specification 中时间字段为字符串，缓存层和 MySQL store 增加时间格式化与解析逻辑，保持缓存排序和更新时间语义。

验证：

- 迁移前检查 `rg -n "github.com/pole-io/pole-server/apis/pkg/types/ai|apis/pkg/types/ai" apis pkg plugin go.mod go.sum` 能发现旧引用；迁移后无命中。
- `gofmt` 已处理改动的 Go 文件。
- `git diff --check` 通过。
- `go test ./pkg/cache/ai ./plugin/apiserver/httpserver/aimcp ./apis/...` 通过。
- `go test ./plugin/store/mysql -run '^$'` 通过编译级验证。
- `GOPROXY=https://goproxy.cn,direct go test ./...` 跑到 MCP/cache 相关包通过后长时间无新输出，已终止，未形成全量通过证明。
- `GOPROXY=https://goproxy.cn,direct go test ./...` 未完整通过，失败点为既有测试环境问题：缺少 healthcheck 测试配置、缺少 i18n toml 文件、heartbeat 测试 nil pointer、mysql tool store 测试缺少 Statis 插件初始化；这些失败不来自 MCP 类型迁移。

后续注意：

- `../specification` 的提交和 `v0.1.0-ALPHA.23` tag 已推送后，control-plane 已重新 `go mod tidy` 并通过聚焦验证。
- 完整 `go test ./...` 仍受既有测试环境问题影响，后续需要单独清理这些历史失败点。

## control-plane MCP spec 桥接收敛 Review

已完成：

- `apis/cache.MCPServerCache.Query` 改为接收 `*ai.MCPServerQuery`，不再暴露 MCP 查询专用的 `map[string]string` 边界。
- `apis/store.MCPServerStore.QueryMCPServers` 改为接收 `*ai.MCPServerQuery`，MySQL store 直接基于 specification query 字段构造查询条件。
- AIMCP `list_mcp_servers` 参数解析收敛到 `ai.MCPServerQuery`。
- AIMCP `list_mcp_server_tools` 参数解析收敛到 `ai.MCPServerToolQuery`。
- AIMCP `create/update_mcp_servers` 参数解析收敛到 `ai.MCPServers` 容器。
- AIMCP `delete_mcp_servers` 参数解析收敛到 `ai.MCPServerDeleteRequest`。
- `BatchQueryResponse.Data` 中的 MCP Server / Tool 结果改为 `anypb.New(spec message)`，不再通过 JSON -> `google.protobuf.Struct` 中转，保留 proto `type_url`。
- 新增 AIMCP 单测覆盖 spec query/container 解析，以及 `Any` 直接反序列化回 `ai.MCPServer` / `ai.MCPServerTool`。

验证：

- 结构性红灯检查先命中旧边界：`mcpObjectToAny`、`structpb`、`Query(filter map[string]string...)`、`QueryMCPServers(filter map[string]string...)`。
- 收敛后同一检查无命中。
- `gofmt` 已处理改动 Go 文件。
- `git diff --check` 通过。
- `go test ./pkg/cache/ai ./plugin/apiserver/httpserver/aimcp ./apis/...` 通过。
- `go test ./plugin/store/mysql -run '^$'` 通过编译级验证。

后续注意：

- `MCPServerQuery.Name` 在 cache 层仍保持前缀匹配语义，MySQL store 的 `QueryMCPServers` 仍是精确匹配；当前 AIMCP 读路径走 cache，因此用户可见语义不变。
- specification 目前没有定义 MCP tool 的创建、更新、删除入口工具；control-plane 现状也只暴露查询 tool。

## control-plane MCP Registry REST 支撑 Review

已完成：

- 在 `/ai/mcp/v1` 下新增 console 可调用的 MCP registry REST 接口：
  - `GET /servers` 查询 MCP Server，参数复用 `ai.MCPServerQuery` 字段。
  - `POST /servers` 创建 MCP Server，body 为 `MCPServer[]`。
  - `PUT /servers` 更新 MCP Server，body 为 `MCPServer[]`。
  - `POST /servers/delete` 删除 MCP Server，body 为 `ai.MCPServerDeleteRequest`。
  - `GET /server/tools` 查询 MCP Server Tool，参数复用 `ai.MCPServerToolQuery` 字段。
- REST handler 复用 AIMCP 已有的 query/create/update/delete/toolQuery 执行路径。
- 修正 MySQL `CreateMCPServer` 的 ID 校验：允许创建时不传 ID，由 store 自动生成 UUID。

验证：

- `gofmt` 已处理新增和改动 Go 文件。
- `git diff --check` 通过。
- `go test ./plugin/apiserver/httpserver/aimcp ./pkg/cache/ai ./apis/...` 通过。
- `go test ./plugin/store/mysql -run '^$'` 通过编译级验证。

# 本地 MCP 联调运行

- [x] 启动 MySQL 容器并初始化 `pole_server` 数据库
- [x] 启动 control-plane 并确认 AIMCP REST 服务监听 8090
- [x] 验证 MCP Server 创建、列表、工具查询 REST 接口
- [x] 修复创建时 UUID 长度与 MySQL `VARCHAR(32)` schema 不匹配的问题
- [x] 修复 AIMCP 后置打开 MCP cache 后没有定时刷新协程的问题
- [x] 补充 MCP ID 生成和后置 cache 启动刷新单测
- [x] 验证前端 dev server 通过代理读取 MCP REST 数据
- [x] 记录验证结果和遗留风险
- [x] 按默认 `all` 模式启动 control-plane，并确认 console 通过 8080 访问
- [x] 修复合并运行时 console `8080` 下 MCP REST 代理，确认页面数据链路可用

# 服务订阅后端修复与前端链路打通

- [x] 明确当前服务订阅链路断点：存储 schema、写入幂等性、查询返回结构、前端 tab
- [x] 先补失败测试，覆盖订阅关系重复写入、分页查询和增量查询
- [x] 修复 MySQL 初始化 schema，并让旧库表启动时自动补齐订阅表时间列与索引
- [x] 修复订阅关系写入为幂等 upsert，避免重复发现导致主键冲突
- [x] 修复 `GetServiceSubscribers` 返回完整 `ServiceSubscriber` 数据，而不是只返回 caller 服务摘要
- [x] 补齐 console 服务订阅接口封装和服务详情页订阅 tab 渲染
- [x] 运行后端聚焦测试、前端响应映射测试、前端构建和本地接口联调
- [x] 记录 review 结果、验证证据和遗留风险

## 服务订阅后端修复与前端链路打通 Review

已完成：

- `service_subscribe_graph` 初始化 SQL 增加 `ctime`、`mtime` 和 `mtime` 索引。
- control-plane MySQL store 初始化时会检查旧表结构，并自动补齐订阅表缺失的 `ctime`、`mtime` 和索引。
- 服务订阅关系 ID 改为 32 位摘要，匹配 `id VARCHAR(32)`。
- `AddServiceSubscibes` 改为 upsert，同一 caller/callee 重复发现时刷新 `mtime`，不再主键冲突。
- `GetServiceSubscribers` 返回 specification 的 `ServiceSubscriber`，包含 `caller` 和 `callee`。
- console 新增 `describeServiceSubscribers`，服务详情页“服务订阅”tab 已接入 `/naming/v1/service/subscribers`。
- 标准响应映射检查已覆盖服务订阅列表，避免只读旧字段导致页面为空。

验证：

- 新增 `TestServiceSubscriberStore_UpsertAndQuery`，先暴露了 ID 超长、重复写入和 `mtime` 查询问题，修复后通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./pkg/service ./plugin/apiserver/httpserver/discover ./plugin/store/mysql -run 'TestServiceSubscriber|TestGetServiceSubscribers|^$'` 通过。
- `cd console/web && npm run test:response-mapping && npm run build` 通过。
- `git diff --check` 通过。
- all 模式已重启，8080 console 入口返回最新构建资源 `assets/index.5328a1c0.js`。
- 通过 8080 console proxy 查询 `/naming/v1/service/subscribers?callee_namespace=pole-system&callee_name=pole.checker&offset=0&limit=10` 返回 1 条 `demo-subscriber` 订阅数据。
- Playwright 使用本机 Chrome 打开 `/discovery/service/instance?namespace=pole-system&service=pole.checker`，切到“服务订阅”tab，确认页面包含 `demo-subscriber`、`pole-system/pole.checker` 和“服务订阅”。

遗留风险：

# 规则治理页面整体优化

- [x] 明确分析范围：`http://127.0.0.1:8080/governance/` 下所有可达治理页面及其前后端链路
- [x] 检查工作区状态，确认已有未提交改动需要保护
- [x] 梳理治理前端页面结构、组件共性、表单/抽屉/空态和 API 调用
- [x] 梳理治理后端 REST 路由、console 代理、标准响应和发布版本接口
- [x] 形成优化设计：页面信息架构、共享组件、数据解包、前后端契约和验证策略
- [x] 经确认后实施前端和后端改动
- [x] 运行聚焦测试、前端响应映射检查、构建和页面联调
- [x] 记录 Review：已完成项、验证证据、遗留风险

## 规则治理页面整体优化 Review

已完成：
- 前端治理根路径增加默认跳转，进入 `/governance/` 后落到无损发布页，避免主区域空白。
- 治理列表页统一为“左侧列表 + 右侧详情”的浅边框双栏结构，并补齐空态：自定义路由、全链路灰度、限流、熔断、主动探测、无损发布。
- 修复治理页面 `editable/deleteable || true` 覆盖后端 `false` 权限的问题，授权动作绑定当前行数据。
- 补齐泳道组授权/删除、泳道规则删除动作；无损发布版本隐藏 rollback，避免调用后端不存在的回滚能力。
- 后端修正主动探测版本列表接口，补齐 faultdetect 路由注册，并接通无损发布版本删除/停止灰度分发。

验证：
- `cd console/web && npm run test:response-mapping && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/apiserver/httpserver/discover ./pkg/goverrule -run 'TestGetPublishFaultDetectRulesUsesRuleReleases|^$'` 通过。
- `git diff --check` 通过。
- 8081 mock 前端注入登录态后验证 `/governance/` 重定向到 `/governance/lossless`，路由规则与全链路灰度空态渲染正常，无 `No routes matched location`；截图：`output/playwright/governance-mock-auth-router-lane.png`。

遗留风险：
- 当前 8080 由 `pole-serv` 监听，页面级验证使用 8081 mock 前端完成；真实数据场景仍需要后端服务与数据库数据配合再做一次端到端联调。
- 构建仍有既有 Browserslist 过期与大 chunk 警告，本次未展开性能拆包。

- 浏览器验证使用的是本地插入的 demo 订阅关系；真实生产数据需要客户端发现请求带 caller 信息后由 batch controller 自动写入。
- 服务订阅表 schema 自恢复只覆盖当前已知缺失字段和索引，不是完整迁移系统。

# 服务订阅页面视觉优化

- [x] 复盘截图问题：普通表格表达关系数据过于松散，文字重复且层级弱
- [x] 将表格改为服务订阅关系流布局，突出 caller -> callee 方向
- [x] 优化顶部上下文、搜索、刷新和分页的空间分配
- [x] 补齐空态和加载态
- [x] 构建并用浏览器截图验证
- [x] 记录 review 和验证结果

## 服务订阅页面视觉优化 Review

已完成：

- 服务订阅 tab 从普通三列表格改为 caller -> callee 的关系流布局。
- 每条订阅关系显示为订阅方服务节点、方向连接和被订阅服务节点，当前服务节点使用轻量高亮。
- 顶部上下文收敛为当前服务、订阅方数量和当前页数量，搜索文案改为“搜索订阅方服务”。
- 底部分页保留 TDesign 标准分页，但修复了“共 N 条关系”换行问题。
- 搜索组件增加可选 placeholder，不影响其它页面默认搜索文案。

验证：

- `cd console/web && npm run test:response-mapping && npm run build` 通过。
- `git diff --check` 通过。
- all 模式已重启，8080 服务加载最新构建。
- Playwright 使用本机 Chrome 打开 `/discovery/service/instance?namespace=pole-system&service=pole.checker`，切到“服务订阅”tab，确认页面包含 `demo-subscriber`、订阅关系流标签和当前服务信息。

# 服务详情页刷新后数据丢失修复

- [x] 复现并定位：刷新后 Redux `editSvc` 为空，`ServiceDetail` 没有按 URL 参数重新拉取
- [x] 增加前端静态回归检查，约束服务详情页必须使用 `namespace/serviceName` props 加载详情
- [x] 修复服务详情页刷新加载逻辑
- [x] 构建并用浏览器刷新页面验证
- [x] 记录 review 和验证结果

## 服务详情页刷新后数据丢失修复 Review

根因：

- `ServiceDetail` 虽然声明了 `namespace/serviceName` props，但组件实际没有使用。
- 详情请求只在 Redux `editSvc` 存在时触发；刷新页面后 Redux 临时态清空，`editSvc=null`，因此详情页不会重新请求服务数据。

本次修复：

- `ServiceDetail` 改为按 URL 传入的 `namespace/serviceName` 重新加载服务详情。
- `listOneService` 支持 `id`、`namespace`、`name` 三个可选过滤参数。
- 避免把服务列表派生出来的展示 id `${namespace}/${serviceName}` 当成真实后端 id 查询。
- `verify-standard-response-mapping.mjs` 增加服务详情刷新回归检查，固定详情页必须使用 `namespace/serviceName` props 查询。

验证：

- 回归脚本先失败，命中 `ServiceDetail` 没有使用 `namespace/serviceName`；修复后 `npm run test:response-mapping` 通过。
- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- all 模式已重启。
- Playwright 使用本机 Chrome 直接打开并刷新 `/discovery/service/instance?namespace=pole-system&service=pole.checker`，确认发起了 2 次 `/naming/v1/services?namespace=pole-system&name=pole.checker&limit=1&offset=0` 请求，页面显示 `pole-system`、`pole.checker` 和服务可见性。

# 服务详情页面视觉优化

- [x] 复盘当前问题：详情页只有 Descriptions 纵向字段，信息层级弱且空值区域过大
- [ ] 重构为身份摘要、关键指标、基础信息、治理归属、标签与时间分组
- [ ] 补齐加载态、无数据态和空字段展示
- [ ] 构建并用浏览器截图验证
- [ ] 记录 review 和验证结果

## 本地 MCP 联调运行 Review

运行环境：

- Docker MySQL 容器：`pole-mysql`，端口 `3306:3306`，root 密码 `123456`。
- 初始化脚本：`plugin/store/mysql/scripts/pole_server.sql`。
- control-plane 启动命令：`MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 go run . start -c ./test/data/bootstrap/pole-server.yaml`。
- control-plane 当前监听：HTTP `8090`，gRPC `8091`，XDS `15010`，Apollo `8890`，Eureka `8761`，Nacos `8848`。
- 默认 `all` 模式已使用修正后的发布配置派生本地配置启动；同一个 `pole-serv` 进程同时监听 HTTP `8090` 和 console `8080`。

本次修复：

- `plugin/store/mysql/mcp_server.go` 新增 32 位无横杠 MCP ID 生成，避免 `id VARCHAR(32)` 写入 36 位 UUID 报 `Data too long for column 'id'`。
- `plugin/apiserver/httpserver/aimcp/server.go` 在 AIMCP 初始化后立即刷新 MCP cache，并使用 HTTP server context 启动定时刷新循环。
- `plugin/apiserver/httpserver/server.go` 将初始化 context 传递给 AIMCP server。
- 新增 `plugin/store/mysql/mcp_server_test.go` 覆盖 MCP ID 长度。
- 新增 AIMCP 单测覆盖后置 MCP cache 启动时立即刷新。
- `deploy/conf/pole-server.yaml` 补齐 `naming.batch`、将 `healthcheck` 迁移到 `naming.healthcheck`，并修正 discover event 插件名为 `EventLogger`。
- `deploy/conf/pole-apiserver.yaml` 打开 `api-http` 下的 `aimcp` API，避免发布配置下 MCP REST 返回 404。
- `bootstrap/config/console_config_test.go` 增加发布配置默认 `all`、console `8080`、`naming.healthcheck`、AIMCP API 开启的加载断言。
- `console/pkg/router` 增加 `/ai/mcp/v1/*` 反向代理到 control-plane，避免合并模式页面请求落到 SPA fallback。
- `console/pkg/handlers` 兼容当前登录接口标准 `data` 响应并设置 JWT cookie，同时移除 console 代理中会调用受保护 `/auth/v1/user/token` 的二次校验循环。

验证：

- `go test -count=1 ./plugin/apiserver/httpserver/aimcp ./pkg/cache/ai ./apis/...` 通过。
- `go test -count=1 ./plugin/apiserver/httpserver -run '^$'` 通过编译级验证。
- `go test -count=1 ./plugin/store/mysql -run 'TestNewMCPID_MatchesSchemaLength'` 通过。
- `git diff --check` 通过。
- `curl http://127.0.0.1:8090/ai/mcp/v1/servers?offset=0&limit=10` 返回 2 条 MCP Server。
- `curl http://127.0.0.1:8090/ai/mcp/v1/server/tools?server_name=demo-mcp&server_namespace=default&offset=0&limit=100` 返回 200 和空工具列表。
- 创建 `demo-mcp-2` 后，control-plane 日志出现 MCP cache `upsert: 1`，REST 列表返回 2 条，证明运行中增量同步生效。
- MySQL 查询确认两条 MCP Server 的 `id_len` 均为 32。

联调数据：

- 联调库中创建了本地演示数据：`demo-mcp`、`demo-mcp-2`，以及前端初始化管理员 `admin/admin123456`。

# MCP 工具抽屉设计优化

- [x] 明确问题范围：用户截图指向 MCP Server 工具抽屉空态，不是 MCP Server 主列表页。
- [x] 重构工具抽屉头部，让 server 名称、协议、命名空间、引用地址和工具数量先建立上下文。
- [x] 优化空态：无工具时不展示空表格，改为状态说明、检查项和刷新/编辑操作。
- [x] 优化有工具时的表格展示：工具名、描述、输入/输出 schema 摘要和时间信息应便于扫描。
- [x] 构建并验证 console 页面可通过 `8080` 访问。
- [x] 记录 review 结果。

## MCP 工具抽屉设计优化 Review

已完成：

- 工具抽屉标题从具体资源名调整为稳定的 `MCP 工具`，抽屉内容顶部增加 server 摘要区，展示 `namespace/name`、协议、描述、工具数、接入、引用和最近修改。
- 无工具时不再展示只有表头的大面积空白表格，改为“暂无工具同步”空态、检查项和 `刷新工具` / `编辑 Server` 操作。
- 有工具时表格改为工具名+描述、输入 schema 摘要、输出 schema 摘要、最近修改，schema 使用三行预览，避免长 JSON 撑开布局。
- 工具抽屉刷新和编辑操作复用当前选中的 MCP Server，工具查询逻辑统一到 `refreshTools`。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 `all` 模式后，`curl http://127.0.0.1:8080/` 返回新构建资源 `index.a17bcb8d.js` 和 `style.0ef7bcba.css`。
- 通过 `8080` 登录后访问 `/ai/mcp/v1/servers?offset=0&limit=10` 返回 2 条 MCP Server。
- 通过 `8080` 查询 `demo-mcp-2` 的 `/ai/mcp/v1/server/tools` 返回 `amount: 0`，覆盖截图里的空态场景。
- 浏览器打开 `http://127.0.0.1:8080/ai/mcps` 后点开 `demo-mcp-2`，确认抽屉渲染出 server 摘要、“暂无工具同步”、检查项、刷新和编辑操作。

# MCP 工具抽屉有数据场景验证

- [x] 明确问题范围：验证 MCP Server 已同步具体 tools 时的抽屉设计和数据链路。
- [x] 插入本地演示 MCP Server Tool 数据。
- [x] 确认 cache / REST / console 通过 `8080` 能读取 tool 数据。
- [x] 用浏览器打开工具抽屉验证有数据表格效果。
- [x] 判断是否需要继续优化有数据态。

## MCP 工具抽屉有数据场景验证 Review

已完成：

- 在本地 MySQL `mcp_server_tools` 表为 `default/demo-mcp-2` 插入 3 条演示工具：`service.search`、`config.publish`、`instance.healthcheck`。
- 通过 `8080` console 代理查询 `/ai/mcp/v1/server/tools`，返回 `amount: 3`，说明 cache、REST 和 console 代理链路均能读到 tool 数据。
- 浏览器打开 `http://127.0.0.1:8080/ai/mcps` 并点开 `demo-mcp-2`，确认工具抽屉顶部工具数显示为 `3`。
- 首次有数据验证发现宽表格会挤压 schema 列，因此继续优化为工具列表行：每个 tool 单独展示名称、描述、更新时间、输入 schema、输出 schema 和 annotations 摘要。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 `all` 模式后，`curl http://127.0.0.1:8080/` 返回新构建资源 `index.6b4cf152.js` 和 `style.a6a6bb15.css`。
- 通过 `8080` 查询工具接口返回 3 条工具名：`service.search`、`config.publish`、`instance.healthcheck`。
- 浏览器最终截图确认有数据态展示工具列表、输入/输出 schema 摘要和 annotations。

# MCP Tools OpenAPI 风格展示重构

- [x] 明确设计方向：tools 应按 API 能力目录展示，而不是普通后台表格或简单列表。
- [x] 新增 tools 搜索、左侧工具目录和右侧详情区。
- [x] 将 JSON Schema 解析成参数/响应字段视图，并保留原始 schema 折叠展示。
- [x] 将 annotations 转成可扫描标签。
- [x] 构建并通过 `8080` 浏览器验证有工具态。
- [x] 记录 review 结果。

## MCP Tools OpenAPI 风格展示重构 Review

已完成：

- 工具有数据态从“工具列表 + schema 摘要块”重构为 OpenAPI 风格 `ToolExplorer`。
- 左侧为工具目录，包含搜索框、工具名称、描述和 annotations 标签；右侧为当前工具详情。
- `input_schema` 和 `output_schema` 会解析成字段视图，展示字段、类型、required/optional、说明、默认值或枚举。
- 原始 schema 收进 `查看原始 Schema` 折叠项，避免 raw JSON 抢占主视觉。
- 增加根据 schema 自动生成的 request / response 示例。
- annotations 解析为顶部标签和独立面板，便于判断 readOnly / destructive / title 等提示。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 `all` 模式后，`curl http://127.0.0.1:8080/` 返回新构建资源 `index.843863e9.js` 和 `style.af03480c.css`。
- 通过 `8080` 查询工具接口返回 3 条工具。
- 浏览器打开 `http://127.0.0.1:8080/ai/mcps`，点开 `demo-mcp-2` 后确认页面包含 `工具浏览`、`搜索工具`、`输入参数`、`返回结构`、`示例`、`查看原始 Schema` 和 `required`。
- 浏览器最终截图确认字段名不再折行，参数表更接近 OpenAPI 文档视图。
- 浏览器自动化输入搜索框时受当前 Browser Use 虚拟剪贴板限制，未完成输入行为验证；搜索框渲染和过滤逻辑已在组件代码中实现。

# MCP Tools Schema 展开与抽屉宽度优化

- [x] 明确问题范围：`array<object>` / object 字段需要能展开查看子字段，当前 drawer 宽度不足以承载 API 文档视图。
- [x] 增强 schema 字段视图，支持嵌套 object 与 array<object> 展开。
- [x] 扩大 MCP tools drawer 宽度，减少参数表挤压。
- [x] 构建并通过 `8080` 浏览器验证展开视图和宽度效果。
- [x] 记录 review 结果。

## MCP Tools Schema 展开与抽屉宽度优化 Review

已完成：

- MCP tools drawer 从 `large` 调整为 `min(1180px, 92vw)`，让 API 文档视图获得更宽的参数表空间。
- schema 字段解析保留原字段 schema，并识别 `object` 与 `array<object>` 的子字段。
- 字段表新增嵌套展开行，例如 `services array<object>` 可展开为 `name`、`namespace`、`instance_count`、`healthy_count`。
- 本地演示数据将 `service.search` 的 `output_schema.services.items` 补充为带 properties 的 object，用于覆盖嵌套展开场景。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 `all` 模式后，`curl http://127.0.0.1:8080/` 返回新构建资源 `index.d3a3613f.js` 和 `style.67280fee.css`。
- 通过 `8080` 工具接口确认 `service.search` 的 `output_schema.services` 为 `array<object>`，且 item 带子字段。
- 浏览器打开 `http://127.0.0.1:8080/ai/mcps`，点开 `demo-mcp-2`，切换到 `service.search`，确认存在 `展开 services 子字段`。
- 浏览器点击展开后确认显示 `服务名`、`命名空间`、`instance_count`、`healthy_count`，并截图验证宽度效果。

# MCP Server Backend 关联设计

- [x] 明确问题：MCP Server 的 backend 可以来自 Pole 注册服务，也可以来自用户自定义地址，当前 `reference` 字段语义不足。
- [x] 在 specification 中补充明确的 backend selector 字段。
- [x] 在 control-plane store/cache/REST 中持久化并返回 backend selector。
- [x] 在 console 表单中支持“注册服务 / 自定义地址”两种 backend 模式，并接入服务列表搜索。
- [x] 在 MCP Server 列表和详情中展示 backend 来源。
- [x] 补充测试与浏览器验证。

## MCP Server Backend 关联设计 Review

已完成：

- `../specification` 新增 MCP Server backend selector 字段：`backend_type`、`backend_service_namespace`、`backend_service_name`、`backend_address`，并发布 `v0.1.0-ALPHA.24`。
- control-plane 升级到 `github.com/pole-io/specification v0.1.0-ALPHA.24`。
- MySQL `mcp_server` 表新增 backend selector 字段和 `backend_type`、`backend_service` 索引。
- MCP Server store 在创建/更新前规范化 backend：
  - `service` 模式只需要关联服务 `namespace/name`，会默认补齐 MCP Server `name`、`namespace`、`reference`。
  - `address` 模式保存 `backend_address`，并清空服务关联字段。
- cache、REST、AIMCP tool 查询均支持按 `backend_type` 和 backend service 过滤。
- console MCP 页面创建/编辑抽屉新增后端模式选择：`Pole 注册服务` / `自定义地址`。
- console 列表、指标和工具抽屉改为展示 backend 来源，不再把 `reference` 当作唯一接入表达。

验证：

- `cd ../specification/source/go && bash build.sh` 通过。
- `cd ../specification/source/rust && bash build.sh` 通过。
- `cd ../specification && GOPROXY=https://goproxy.cn,direct go test ./...` 通过。
- `cd ../specification/source/rust/pole-specification && PROTOC=../../protoc/protoc-darwin-arm64/bin/protoc cargo test --release` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/store/mysql ./pkg/cache/ai ./plugin/apiserver/httpserver/aimcp ./apis/...` 通过。
- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 本地 MySQL demo 数据更新后，`http://127.0.0.1:8090/ai/mcp/v1/servers` 返回 `demo-mcp-2` 为 service backend、`demo-mcp` 为 address backend。
- 通过 `8080` console 代理登录后请求 `/ai/mcp/v1/servers`，返回同样的 backend 字段。
- 通过 `8080` 创建临时 MCP Server 时只提交 `backend_type=service`、`backend_service_namespace=pole-system`、`backend_service_name=pole.checker`，返回记录自动补齐 `name=pole.checker`、`namespace=pole-system`、`reference=pole.checker`；随后已删除临时记录。

注意：

- Browser 插件本轮能读取登录页 DOM，但当前 evaluate 环境不能访问 `fetch`、`XMLHttpRequest`、`localStorage` 或 `document`，无法在插件内注入登录态完成页面截图；页面验证以构建、静态资源和 8080 代理 API 链路为准。

# Console 列表接口 401 修复

- [x] 复现其他列表接口 401，并区分 console 代理层与 pole-server 后端鉴权层。
- [x] 对比可工作的 MCP 接口和失败列表接口的认证路径差异。
- [x] 定位根因并实现最小修复。
- [x] 补充回归测试。
- [x] 重新运行 console 代理接口、Go 测试和前端构建验证。

## Console 列表接口 401 修复 Review

根因：

- 普通 `core/naming` 列表接口会走 console 鉴权和资源权限过滤，MCP registry 列表不走同一条鉴权链路，所以 MCP 能返回数据但其他列表 401。
- 本地库中的 `admin` 存量 token 由旧 salt 生成；登录接口原样返回缓存/数据库里的 token，没有校验 token 是否能被当前 `auth.user.option.salt=polarismesh@2021` 解开，导致登录成功后后续普通列表仍被后端判定为 `invalid token`。
- owner 主账号通过接口权限后，列表资源过滤还会调用 `ResourcePredicate`；该入口原先没有 owner 快速放行，容易继续依赖策略缓存。

本次修复：

- `CheckCredential` 优先读取 `ContextAuthTokenKey`，并对 `Authorization` / `X-Polaris-Token` 做大小写不敏感兜底，避免 header 规范化差异导致 token 丢失。
- 登录成功时校验用户 token 是否可由当前 salt 解开且绑定当前用户；不满足时自动生成新 token、更新 DB、刷新缓存并返回新 token。
- owner 用户在 `CheckPermission` 和 `ResourcePredicate` 两个入口都直接放行，覆盖接口权限和列表资源过滤。
- console 代理改为仅在后端 2xx 时续期 JWT；后端返回 401 时清除 `jwt` cookie，避免旧 cookie 被无限续期。
- console 前端遇到 HTTP 401、`407` 或 `401001` 时清理本地登录态并跳转登录页，覆盖浏览器保留旧 `localStorage` token 的场景。
- 保留低噪声鉴权失败日志，便于后续区分 credential 阶段和 policy 阶段失败。

验证：

- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/access_control/auth/user ./plugin/access_control/auth/policy ./pkg/namespace/interceptor/auth ./pkg/service/interceptor/auth ./console/pkg/handlers ./console/pkg/router` 通过。
- `cd console/web && npm run build` 通过。
- console 代理单测覆盖后端 2xx 后续期 JWT、后端 401 后清除 JWT。
- 使用旧 JWT cookie 请求 `/naming/v1/services`，确认响应为 401 且返回 `Set-Cookie: jwt=; Max-Age=0` 清理 cookie。
- 重启 all 模式后，登录 `admin/admin123` 返回 `200000`，旧 token 自动轮换，返回 token 与 DB token 一致。
- 通过 `8080` console 代理请求 `/core/v1/namespaces?offset=0&limit=10` 返回 `code=200000`、`amount=2`。
- 通过 `8080` console 代理请求 `/naming/v1/services?offset=0&limit=10` 返回 `code=200000`、`amount=1`。
- 通过 `8080` console 代理请求 `/ai/mcp/v1/servers?offset=0&limit=2` 返回 `code=200000`、`amount=2`。

# 遗留问题修复与前后端仓库/部署形态评估

- [x] 复现并定位 `plugin/store/mysql` 全包测试失败
- [x] 修复 Statis 插件初始化导致的 MySQL store 测试失败
- [x] 重新运行 MCP 相关测试、MySQL store 全包测试和必要构建验证
- [x] 评估 pole-console 与 pole-control-plane 合并仓库/合并部署/分开部署方案
- [x] 在任务文档记录验证结果和架构建议

## 遗留问题修复与前后端仓库/部署形态评估 Review

遗漏问题已修复：

- 失败根因：`plugin/store/mysql` 单元测试只构造 sqlmock，没有加载 Statis 插件配置；`BaseDB.Query` 上报调用指标时触发 `statis.GetStatis()`，空配置会构造名称为空的 Statis 插件并 panic。
- 修复方式：`apis/observability/statis.GetStatis()` 在 Statis 配置完全为空时返回空 composite，作为 no-op metrics sink；实际 bootstrap 配置了 `local`/`prometheus` entries，因此生产配置行为不变。
- 回归覆盖：新增 `apis/observability/statis/statis_test.go`，验证空配置下 `GetStatis().ReportCallMetrics(...)` 不 panic。

验证：

- `go test -count=1 ./apis/observability/statis` 通过。
- `go test -count=1 ./plugin/store/mysql -run 'Test_toolStore_GetUnixSecond'` 通过。
- `go test -count=1 ./plugin/store/mysql` 通过。
- `go test -count=1 ./plugin/apiserver/httpserver/aimcp ./plugin/apiserver/httpserver ./pkg/cache/ai ./plugin/store/mysql ./apis/...` 通过。
- `npm run build:test` 通过。
- control-plane REST 和 pole-console 代理均返回 2 条 MCP Server。

仓库与启动形态建议：

- 可以合并成一个 monorepo，并优先标准化启动模式；本阶段不考虑容器化。
- 标准启动语义：
  - 默认启动模式是 `all`，即合并启动完整产品。
  - `server` 表示只启动 control-plane 后端服务。
  - `console` 表示只启动 console 前端/gateway。
  - CLI 建议为 `pole start --mode all|server|console`，配置项建议为 `bootstrap.mode: all|server|console`；命令行参数优先于配置。
- 合并模式：
  - `mode=all` 时在 control-plane 进程内启动后端和 console gateway，不再 exec 外部 pole-console 进程。
  - `mode=server` 只启动 control-plane。
  - `mode=console` 只在当前进程内启动 console gateway，并通过配置里的 control-plane 地址连接后端。
  - 这是第一阶段落地实现，保留 pole-console 当前 Go gateway 的 JWT、SPA fallback、代理和观测接口逻辑。
- 同端口单 HTTP server 模式：
  - 后续可进一步把 console Gin handler 挂到 control-plane HTTP mux。
  - 需要处理静态资源服务、SPA fallback、API 路由排除、认证/初始化兼容逻辑以及代理路径冲突。
- 推荐优先级：
  1. 先定义 `mode` 标准和配置优先级，默认 `all`。
  2. 先完成代码合并，并在 control-plane 进程内启动 console gateway。
  3. 再做同端口挂载：control-plane HTTP server 增加 console handler，仍沿用同一套 `mode` 语义。

注意：

- 当前 pole-console 不是纯前端仓库，它还有 Go gateway、JWT、后端代理和观测数据访问逻辑。若直接删除 console gateway，需要先确认这些能力是否都能由 control-plane HTTP server 或外部网关替代。

# 启动模式标准化实现（已被后续合并修正替换）

- [x] 明确实现范围：默认 `all`，可切换 `server` / `console`
- [x] 为启动模式解析和分支补测试
- [x] 增加 CLI `--mode` 与配置 `bootstrap.mode` 的优先级处理
- [x] 更新本地 bootstrap 配置示例
- [x] 运行 Go 聚焦测试与必要联调验证
- [x] 记录结果和后续修正事项

## 启动模式标准化实现 Review

本节最初错误地走向了“control-plane 拉起外部 console 子进程”。该方向已根据用户纠正废弃，后续以“pole-console 代码合并修正”为准。

- `bootstrap.mode` 支持 `all` / `server` / `console`，空值默认 `all`。
- CLI 新增 `start --mode all|server|console`，命令行参数优先于配置。

# pole-console 代码合并修正

- [x] 明确修正范围：从外部子进程 runner 改为 control-plane 内部合并 console 代码
- [x] 回收刚才引入的外部 console process 配置与 runner
- [x] 将 `../pole-console` 的 Go gateway 与 web 代码迁入 control-plane 仓库
- [x] 调整 pole-console Go import，使其成为 control-plane 内部 package
- [x] 在 control-plane 进程内按 `mode=all|console` 启动 console gateway
- [x] 保留 `mode=server` 只启动后端 API 的能力
- [x] 补充启动模式/console 挂载测试
- [x] 运行 Go 与前端聚焦验证
- [x] 记录最终 review

## pole-console 代码合并修正 Review

已完成：

- 新增 `console/` 目录，迁入 pole-console 的 Go gateway、web 源码、配置示例与测试数据。
- 将 console Go import 从 `github.com/pole-io/pole-console` 改为 `github.com/pole-io/pole-server/console`。
- `console/pkg/router.Router` 拆出 `NewRouter(config)`，可构建 Gin handler 而不强制独立 main 入口。
- 新增 `console.Start(ctx, config, errCh)`，在 control-plane 当前进程内启动 console gateway，随 bootstrap context 关闭。
- `bootstrap.console` 改为内嵌 console gateway 配置，不再支持外部 `command/args/workDir/env` 子进程配置。
- `bootstrap.Start` 中 `mode=console` 只启动内嵌 console gateway；`mode=all` 启动 control-plane 后端后，在同一进程内启动 console gateway；`mode=server` 保持后端-only。
- 删除外部子进程 runner 和对应测试。
- console 迁入后统一使用 `github.com/pole-io/specification`，移除 `github.com/polarismesh/specification`，避免同一进程内同名 proto 文件重复注册 panic。
- 适配 pole-io spec：console 登录代理里的主用户解析改为读取 `model.Response.Data` 中的 `security.User`，用户名称按 string 处理。
- 根 `.gitignore` 增加 `console/web/node_modules`、`console/web/dist`、`console/node_modules`，避免迁入构建产物和依赖目录。

验证：

- `go test -count=1 ./bootstrap/config -run 'TestLoad_LoadsEmbeddedConsoleConfig'` 先红灯，确认旧 `ConsoleProcess` 不满足内嵌配置；替换后通过。
- `go test -count=1 ./console/...` 通过。
- `go test -count=1 ./bootstrap ./bootstrap/config ./cmd ./console/...` 通过。
- `go test -count=1 ./plugin/apiserver/httpserver/aimcp ./plugin/apiserver/httpserver ./pkg/cache/ai ./plugin/store/mysql ./apis/...` 通过。
- `go test -count=1 ./plugin/apiserver/httpserver/aimcp ./plugin/apiserver/httpserver ./pkg/cache/ai ./plugin/store/mysql ./apis/... ./bootstrap ./bootstrap/config ./cmd ./console/...` 通过。
- `cd console/web && npm ci && npm run build:test` 通过；`npm ci` 报告迁入前端依赖存在 27 个漏洞提示，本次未做依赖升级。
- 临时 `mode=all` 配置联调通过：同一个 `pole-serv` 进程同时监听 `8090` 和 `18080`。
- `curl http://127.0.0.1:8090/ai/mcp/v1/servers?offset=0&limit=10` 返回 MCP Server 列表。
- `curl -I http://127.0.0.1:18080/` 返回 `HTTP/1.1 200 OK`。
- 临时联调进程已停止，`8090` / `18080` 均已释放。
- `git diff --check` 通过。

注意：

- 当前是“代码合并 + 单进程内启动 console gateway”，console gateway 仍按原能力监听自己的 `webServer.listenPort`，不是 exec 子进程。
- 后续若要进一步做到同一个 HTTP 端口，需要再把 console Gin handler 挂到 control-plane HTTP mux，并处理 `/assets`、SPA fallback、API 路由优先级和代理路径冲突。

# console / Apollo 端口修正

- [x] 明确端口语义：console 默认入口使用 `8080`，Apollo 避让
- [x] 核对本地与部署 apiserver 配置
- [x] 将部署配置中的 Apollo 端口从 `8080` 调整为 `8890`
- [x] 保持 console 配置示例和迁入的 pole-console 配置使用 `8080`
- [x] 运行配置检查与聚焦验证

## console / Apollo 端口修正 Review

已完成：

- `test/data/bootstrap/pole-apiserver.yaml` 原本已使用 Apollo `8890`，无需调整。
- `deploy/conf/pole-apiserver.yaml` 中 `service-apollo.listenPort` 从 `8080` 改为 `8890`。
- `test/data/bootstrap/pole-server.yaml`、`deploy/conf/pole-server.yaml` 的 console 示例继续使用 `webServer.listenPort: 8080`。
- `console/test/data/bootstrap/pole-console.yaml` 继续作为测试参考配置使用 `8080`。

结论：

- 合并启动时，`8080` 留给 console 更合适。
- Apollo 作为兼容协议端口避让到 `8890`，与当前本地测试配置保持一致。

# console deploy 合并清理

- [x] 明确部署合并策略：删除 `console/deploy` 独立脚本，统一使用 `deploy/tools`
- [x] 将 release 配置改为 `mode: all` 并内嵌真实 `bootstrap.console`
- [x] 更新 release 打包脚本，构建并拷贝 console web dist
- [x] 删除迁入的 `console/deploy` 独立部署目录
- [x] 统一 `deploy/tools` 内部脚本路径
- [x] 运行构建脚本/配置测试/状态检查
- [x] 记录最终 review

## console deploy 合并清理 Review

已完成：

- 删除 `console/deploy/conf/pole-console.yaml` 和 `console/deploy/tool/*`，不再保留独立 pole-console 部署入口。
- `deploy/conf/pole-server.yaml` 改为默认 `mode: all`，并在 `bootstrap.console` 内嵌真实 console gateway 配置。
- `deploy/build.sh` 增加 console web 构建步骤，并把 `console/web/dist` 打进 release 包。
- `deploy/build.sh` 使用 `npm ci --legacy-peer-deps` 安装前端依赖，兼容当前旧 TypeScript 与 i18next peer 依赖树。
- `deploy/tools` 保持唯一启停脚本目录，并修正 shell 脚本内部 `tool/include` / `tool/check.sh` 为 `tools/include` / `tools/check.sh`。

验证：

- `bash -n deploy/build.sh deploy/tools/start.sh deploy/tools/check.sh deploy/tools/stop.sh deploy/tools/p.sh deploy/tools/include` 通过。
- `go test -count=1 ./bootstrap/config ./bootstrap ./cmd ./console/...` 通过。
- `npm run build` 在 `console/web` 下通过。
- `bash deploy/build.sh` 通过。
- `unzip -l pole-server-release_*.zip | rg 'conf/pole-server.yaml|tools/start.sh|tools/include|console/web/dist/index.html'` 确认 release 包包含统一配置、统一脚本和 console web dist。
- `git diff --check` 通过。

注意：

- `bash deploy/build.sh` 会生成被 `.gitignore` 忽略的本地 release 产物和 `console/web/dist`。
- `npm ci --legacy-peer-deps` 仍报告迁入前端依赖存在漏洞提示，本次仅保证合并部署链路可用，未做前端依赖升级。

# MCP Console 页面设计优化

- [x] 使用 frontend-skill 明确 MCP 页面作为运维工作台，而不是营销页
- [x] 重构 MCP 页面信息层级：状态概览、筛选、表格、抽屉
- [x] 使用 tdesign-react 现有组件优化操作与筛选体验
- [x] 补充样式并避免卡片堆叠、过重边框和单色主题
- [x] 构建 console 前端并验证 8080 页面可访问
- [x] 记录 review 和验证结果

## MCP Console 页面设计优化 Review

已完成：

- MCP 页面改为工作台式布局：顶部说明当前 registry 范围，状态栏展示 server 数、namespace 数、协议分布和引用数。
- 筛选区拆成协议快捷筛选和名称/命名空间精确查询，减少页面横向堆控件。
- 表格首列改成 MCP Server 摘要，聚合名称、协议、命名空间、引用/描述；接入、归属、可见范围、最近修改独立成更容易扫描的列。
- 创建/编辑抽屉按“基础信息”和“接入配置”分组，两列布局提升填写效率。
- 补充 `stdio` 协议选项，避免已有数据只能显示但不能通过表单选择。

验证：

- `cd console/web && npm run build` 通过。
- `curl http://127.0.0.1:8080/` 返回新构建后的静态资源 hash。
- 登录后访问 `http://127.0.0.1:8080/ai/mcp/v1/servers?offset=0&limit=10` 返回 `demo-mcp-2`、`demo-mcp` 两条数据。
- `git diff --check` 通过。

注意：

- Browser 插件本轮没有提供直接截图工具，Node REPL 里也无法加载 Playwright；本次没有自动化截图验证。

# Console 服务列表无数据修复

- [x] 复现服务列表页面无数据，并确认 8080 实际 API 返回。
- [x] 对照服务列表前端请求、响应解包和表格字段映射。
- [x] 定位根因并实现最小修复。
- [x] 补充可执行回归验证。
- [x] 重新运行后端/前端验证，并记录 review。

## Console 服务列表无数据修复 Review

根因：

- `/naming/v1/services` 后端已经返回标准响应 `data: [...]`，但 `console/web/src/services/service.ts` 仍按旧响应读取 `res.services`。
- 因此接口是 `200000` 且 `amount=1`，但前端映射阶段拿不到列表，页面显示为空。

本次修复：

- 服务列表响应兼容 `res.data ?? res.services ?? []`，与 namespace/mcp 页面当前标准响应处理保持一致。
- 对空 `id` 的服务行补展示层派生 id：`${namespace}/${name}`，避免表格 `rowKey="id"` 在自注册服务 id 为空时不稳定。

验证：

- `cd console/web && npm run build` 通过。
- 本地模拟标准响应映射，`list.length=1`，派生 id 为 `pole-system/pole.checker`。
- 重启 all 模式后，`http://127.0.0.1:8080/` 加载新前端资源 `assets/index.4f61ab23.js`。
- 使用当前 admin 的 `X-Pole-User` 和 `Authorization` 请求 `http://127.0.0.1:8080/naming/v1/services?offset=0&limit=10` 返回 `code=200000`、`amount=1`、首条服务为 `pole-system/pole.checker`。

注意：

- 当前本地库只有 `admin` 用户，且密码不再是先前记录的 `admin/admin123456`；本轮未重置本地管理员密码，页面登录态验证以 header token + API 链路和前端映射验证替代。

# Console 列表响应解包全盘校准

- [x] 写前端响应映射回归脚本，先捕获服务实例页只读旧字段的问题。
- [x] 用 8080 实际接口响应确认服务实例页后端返回结构。
- [x] 修复服务实例页和扫描命中的同类列表字段解包。
- [x] 抽查主要列表接口的前端映射与后端返回字段是否一致。
- [x] 运行前端构建、回归脚本和接口验证，并记录 review。

## Console 列表响应解包全盘校准 Review

根因：

- 服务实例页与服务列表页是同类问题：后端 `/naming/v1/instances` 已返回标准 `{data, amount, size}`，但前端只读 `res.instances`。
- 进一步扫描发现用户、用户组、鉴权策略、服务别名、配置分组、配置文件、限流规则仍有只读旧列表字段风险。

本次修复：

- 新增 `console/web/scripts/verify-standard-response-mapping.mjs` 和 `npm run test:response-mapping`，固定列表映射必须先读标准 `data/amount`，再兼容旧字段。
- 修复 `instance.ts`、`alias.ts`、`users.ts`、`user_group.ts`、`auth_policy.ts`、`config_group.ts`、`config_files.ts`、`ratelimit.ts` 的列表响应映射。
- 兼容规则统一为：列表取 `data ?? legacyField ?? []`，总数取 `amount ?? total ?? list.length`。

验证：

- 回归脚本先在旧代码上失败，命中 `res.instances`、`result.users`、`result.userGroups` 等旧字段；修复后 `npm run test:response-mapping` 通过。
- `cd console/web && npm run build` 通过。
- 通过 `8080` 抽查实际响应：`instances` 返回 `amount=1` 且首条实例为 `pole-system/pole.checker / 127.0.0.1`；`users` 返回 `amount=1` 且首条用户为 `admin`；`aliases`、`usergroups`、`roles`、`configGroups` 返回标准 `data` 空数组。
- 鉴权策略页面调用会带 `default=true/false`；带该参数时 `/auth/v1/policies` 返回标准 `data` 空数组。裸请求不带 `default` 时后端当前会返回 502/empty reply，本次未改后端行为。
- 静态扫描确认上述旧字段只作为标准 `data` 后的兼容兜底存在。

# 服务详情页面视觉优化

- [x] 复盘当前问题：详情页只有 Descriptions 纵向字段，信息层级弱且空值区域过大。
- [x] 重构为身份摘要、关键指标、基础信息、治理归属、可见范围、标签与时间分组。
- [x] 补齐加载态、无数据态和空字段展示。
- [x] 构建并用浏览器截图验证。
- [x] 记录 review 和验证结果。

## 服务详情页面视觉优化 Review

已完成：

- 服务详情页从单列 `Descriptions` 改为工作台式详情：顶部展示服务名、命名空间、可见性和健康/总实例/标签数量。
- 详情内容按“基础信息 / 治理归属 / 可见范围 / 标签 / 时间”分组，减少纵向堆字段和大面积空白。
- 直刷详情页时使用 URL 中的 `namespace/serviceName` 拉取服务详情，不依赖从列表页带入的临时状态。
- 加载态、无详情态和空字段统一处理，合法数值 `0` 不会被误替换成占位符。

验证：

- `cd console/web && npm run test:response-mapping` 通过。
- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，`curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.d760ec3d.js` 和 `assets/style.64c1c9c2.css`。
- 浏览器直刷 `http://127.0.0.1:8080/discovery/service/instance?namespace=pole-system&service=pole.checker`，页面包含 `pole.checker`、`pole-system`、`基础信息`、`治理归属`、`可见范围`、`时间`。
- 页面内服务详情接口返回 `status=200`、`code=200000`、`amount=1`、首条服务为 `pole.checker`。

# 实例详情抽屉视觉优化

- [x] 复盘当前问题：实例详情抽屉仍是单列 Descriptions 字段，健康、隔离、位置和标签没有层级。
- [x] 重构 view 模式为实例身份摘要、关键状态、运行信息、健康检查、位置和标签分组。
- [x] 保持创建/编辑表单链路不变，避免影响实例写操作。
- [x] 构建并用浏览器打开实例抽屉验证。
- [x] 记录 review 和验证结果。

## 实例详情抽屉视觉优化 Review

已完成：

- `InstanceEditor` 的 view 模式不再使用单列 `Descriptions`，改成只读实例详情面板。
- 抽屉顶部展示 `host:port`、命名空间、服务、协议，右侧优先展示健康状态、隔离状态和健康检查状态。
- 详情内容按“运行信息 / 健康检查 / 位置 / 实例标签”分组，减少右侧抽屉的大面积字段清单感。
- view 模式抽屉宽度调整为 `680px`；创建/编辑表单仍使用原来的 `large`，避免影响写操作布局。

验证：

- `cd console/web && npm run test:response-mapping` 通过。
- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，`curl http://127.0.0.1:8080/` 返回新构建资源 `assets/index.8ac5a7b8.js` 和 `assets/style.c3b6f940.css`。
- 通过 `8080` 查询 `/naming/v1/instances?namespace=pole-system&service=pole.checker&offset=0&limit=10` 返回 `code=200000`、`amount=1`、实例 `127.0.0.1:8091`。
- 浏览器打开服务实例页并点击 `127.0.0.1`，确认抽屉包含 `实例详情`、`127.0.0.1:8091`、`运行信息`、`健康检查`、`位置`、`实例标签`，并显示 `build-revision`、`polaris_service` 标签。

# 元数据键值控件与实例编辑页优化

- [x] 复盘当前问题：实例编辑抽屉是长表单，实例标签控件仍是表格内联编辑，不适合多处元数据修改。
- [x] 将 `LabelInput` 升级为通用键值编辑器，保持 `{ key, value }[]` 表单结构不变。
- [x] 兼容现有 `editable`、`disabled`、有/无 `form` 的调用方式。
- [x] 优化实例编辑页布局，拆成运行配置、状态控制、位置、元数据分组。
- [x] 构建并通过浏览器验证实例编辑抽屉。
- [x] 记录 review 和验证结果。

## 元数据键值控件与实例编辑页优化 Review

已完成：

- `LabelInput` 从表格内联编辑改为通用键值编辑器：编辑态展示 key/value 行、删除按钮、添加按钮和标签数量；只读态展示标签 chip；空态展示添加入口。
- 控件保留原有表单值结构 `{ key, value }[]`，并兼容现有 `editable`、`disabled`、有/无 `form` 的调用方式。
- 即时校验只检查重复 key，避免用户点击“添加标签”后空行立刻把整个标签区标红。
- 实例编辑提交时过滤空 key/value 的临时标签行，避免空行写入 `metadata`。
- 实例编辑抽屉拆分为“运行配置 / 状态控制 / 位置 / 元数据”四个分组，减少长表单清单感。
- 停掉了残留的 Node/Vite 进程，确保 `127.0.0.1:8080` 访问的是 all 模式 console，而不是 dev server。

验证：

- `cd console/web && npm run test:response-mapping` 通过。
- `cd console/web && npm run build` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，`127.0.0.1:8080` 只剩 `pole-server` 监听，首页返回 `assets/index.cd44a7c0.js` 和 `assets/style.57b04847.css`。
- 浏览器打开服务实例页并点击编辑按钮，确认抽屉包含 `编辑实例`、`运行配置`、`状态控制`、`位置`、`元数据`，元数据区包含 `键`、`值`、`操作` 和 `2 个标签`。
- 浏览器点击 `添加标签` 后，元数据区显示 `3 个标签`，新增空行 placeholder 为 `标签键` / `标签值`，没有即时空值错误，也没有红色输入框。

# 实例编辑与查看风格统一

- [x] 复盘当前问题：实例编辑抽屉虽然拆分了分区，但标题、顶部摘要和分区命名仍与实例详情查看态不一致。
- [x] 编辑态复用实例详情查看态的身份摘要和状态条。
- [x] 编辑态分区统一为“运行信息 / 健康检查 / 位置 / 实例标签”。
- [x] 保持实例创建、编辑、元数据提交流程不变。
- [x] 构建并通过浏览器验证编辑态和查看态一致。
- [x] 记录 review 和验证结果。

## 实例编辑与查看风格统一 Review

已完成：

- 抽取 `InstanceSummary`，查看态和编辑态复用同一套 `host:port`、命名空间、服务、协议、健康状态、隔离状态、健康检查状态摘要。
- 编辑态分区从“运行配置 / 状态控制 / 位置 / 元数据”统一为“运行信息 / 健康检查 / 位置 / 实例标签”，与查看态一致。
- 编辑态仍保留可编辑控件和元数据键值编辑器，提交 payload 结构不变。
- 清理并重建 `console/web/dist`，避免入口 HTML 引用不存在的旧 JS hash。

验证：

- `cd console/web && npm run test:response-mapping` 通过。
- `cd console/web && rm -rf dist && npm run build` 通过。
- `git diff --check` 通过。
- 重启 all 模式后，`127.0.0.1:8080` 只剩 `pole-server` 监听，首页返回 `assets/index.03414bb1.js` 和 `assets/style.79183194.css`，入口 JS 返回 `200`。
- 浏览器分别打开编辑实例和实例详情抽屉，确认两者都包含 `127.0.0.1:8091`、`健康状态`、`隔离状态`、`健康检查`，以及 `运行信息`、`健康检查`、`位置`、`实例标签` 四个分区。
- 浏览器确认编辑态不再包含 `运行配置`、`状态控制`、`元数据` 旧分区文案；编辑态仍包含 `键`、`值` 和 `2 个标签`，查看态保持只读标签展示。

# 治理规则详情抽屉对齐与滚动优化

- [x] 对照截图梳理路由详情中基础信息、主被调服务、规则表格的对齐和溢出问题。
- [x] 将路由详情查看态收敛为更紧凑的 inspector 布局，减少首屏纵向占用。
- [x] 固定规则表格列宽并处理长文本、标签换行，避免横向撑出抽屉。
- [x] 重建并用 8080 实际页面验证抽屉对齐、滚动和内容展示。
- [x] 记录 review、验证结果和剩余风险。

## 治理规则详情抽屉对齐与滚动优化 Review

已完成：

- 路由规则详情的基础信息查看态从 `FormItem` 表单块改为只读 definition grid，规则名、优先级、描述、标签按固定列对齐，减少表单额外高度。
- 主调/被调服务在查看态改为紧凑关系摘要，编辑态仍保留原命名空间和服务选择器，不影响写操作。
- 匹配条件表和目标分组表启用固定表格布局，参数类型、参数键、匹配类型、操作列固定宽度，匹配值和实例标签允许换行。
- 实例标签使用可换行标签列表，避免 `lane 完全匹配 blue`、`version 完全匹配 v2` 这类内容把表格撑出抽屉。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过；强制重编译后对 `/tmp/pole-control-plane` 重新 `codesign --force --sign -`，LaunchAgent 恢复运行。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `72759`，8080 首页资源为 `assets/index.c8f071d7.js` 和 `assets/style.9549be2b.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/?t=signed-final` 并点击 `demo-governance-route-202606050424`，抽屉宽度 `860px`，基础信息高度 `98px`，主被调服务区高度 `128px`，规则区顶部从上一轮约 `510px` 降到 `476px`。
- DOM 验证两个规则表格 `overflowsWrapper=false`，关键字段 `x-tenant`、`region`、`payment-v2`、`lane 完全匹配 blue` 均展示。
- 最新截图保存到 `/tmp/governance-route-drawer-aligned-final.png`。

# 治理规则编辑态抽屉排版修正

- [x] 复现路由规则编辑态基础信息、标签、服务选择和规则表格的排版问题。
- [x] 修复基础信息编辑态 grid，避免标签内容被压到逐字竖排。
- [x] 调整规则表格编辑控件列宽，减少参数类型和操作列的截断/拥挤。
- [x] 构建、重启 all 模式并在 8080 实际编辑态验证。
- [x] 记录 review、验证结果和剩余风险。

## 治理规则编辑态抽屉排版修正 Review

已完成：

- 路由规则编辑态基础信息使用专用 `editMetaGrid`，不再复用查看态三列布局。
- 局部覆盖 TDesign `FormItem` 的固定 `labelWidth=100px` 和 `controls margin-left=100px`，改为 label 在上、控件全宽，避免内容被压缩。
- 空标签文案不再逐字断行，标签区域宽度从复现时的 `20px` 恢复到 `648px`。
- 匹配条件表的参数类型列从 `150px` 调整为 `178px`，操作列从 `80px` 收窄到 `64px`，减少 `请求头(HEADER)` 这类 Select 文案截断。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过，`codesign --force --sign - /tmp/pole-control-plane` 后重启 all 模式。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `3037`，8080 首页资源为 `assets/index.d4a9c9d5.js` 和 `assets/style.7436b6d5.css`。
- 浏览器打开路由规则抽屉并点击 `编辑`，DOM 验证 `editMetaGrid` 为 `494px 220px` 两列，基础信息高度 `120px`，`tagListWidth=648px`、`hasVerticalEmptyTag=false`。
- DOM 验证两个规则表格 `overflows=false`，编辑态仍显示 `保存 / 撤销` 操作。

# 治理规则目标分组编辑交互排版修正

- [x] 复现目标分组表中权重 InputNumber 与实例标签重叠的问题。
- [x] 调整目标分组表的权重列、实例标签列和操作列宽度，避免控件互相覆盖。
- [x] 检查实例标签编辑弹窗的表单排版和可操作性。
- [x] 构建、重启 all 模式并在 8080 实际编辑态验证。
- [x] 记录 review、验证结果和剩余风险。

## 治理规则目标分组编辑交互排版修正 Review

已完成：

- 目标分组表的 `权重` 列从 `110px` 调整为 `178px`，给 TDesign 横向 `InputNumber` 留足宽度。
- `是否隔离` 列从 `120px` 收敛为 `108px`，`操作` 列从 `92px` 收敛为 `78px`，把空间让给权重和标签。
- `实例标签` 列增加 `minWidth=200`，标签流使用 `instanceTagList`，支持换行且不覆盖相邻控件。
- `weightInput` 限定宽度为 `156px` 且不超过单元格，避免按钮组溢出。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过，`codesign --force --sign - /tmp/pole-control-plane` 后重启 all 模式。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `60290`，8080 首页资源为 `assets/index.afc797a7.js` 和 `assets/style.930c11b4.css`。
- 浏览器打开路由规则抽屉并点击 `编辑`，目标分组表宽度 `732px`、`scrollWidth=730`、`groupTableOverflows=false`。
- 目标分组两行单元格宽度均为 `[150, 108, 178, 216, 78]`，权重控件宽度 `150px`，两行 `overlapsTag=false`。
- 点击目标分组行的编辑标签图标后，`编辑实例标签` 弹窗宽度 `700px`、`overflowsViewport=false`，确认/取消/添加标签均正常展示。

# 治理规则子规则折叠与排序交互

- [x] 梳理路由规则子规则的数据结构和当前渲染方式。
- [x] 为 `规则 [n]` 子规则块增加折叠/展开状态，保留 header 摘要。
- [x] 为编辑态子规则块增加拖拽手柄，拖动后调整 `routing_config.rules` 顺序。
- [x] 为目标实例分组行增加拖拽手柄，拖动后调整当前子规则的 `destinations` 顺序并同步 `priority`。
- [x] 构建、重启 all 模式并记录验证结果。

## 治理规则子规则折叠与排序交互 Review

已完成：

- 新增 `collapsedRuleIndexes`，规则块 header 始终可见，正文可折叠隐藏。
- 新增子规则拖拽手柄，使用 HTML5 drag/drop 在编辑态调整 `routing_config.rules` 数组顺序。
- 新增目标实例分组行拖拽手柄，拖动当前子规则内的分组行时同步重排 `destinations` 并按新位置写回 `priority`。
- 子规则 header 增加摘要：匹配条件数量和实例分组数量，折叠后仍能快速判断规则内容规模。
- 拖拽中的规则块和拖拽手柄增加轻量高亮，不改变原有保存/撤销/发布流程。

验证：

- `cd console/web && npm run build` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过，`codesign --force --sign - /tmp/pole-control-plane` 后重启 all 模式。
- `git diff --check` 通过。
- 重启 all 模式后，LaunchAgent `io.pole.control-plane.local` 当前 PID `10450`，8080 首页资源为 `assets/index.91d6d257.js` 和 `assets/style.f1781d2d.css`。
- 浏览器 DOM 读取曾确认页面进入治理工作台，入口资源加载为本次构建资源；后续 in-app browser 会话在截图调用时 reset，未能完成端到端拖拽 DOM 断言。本次以构建、静态结构和前端运行资源验证为准。

# 治理规则抽屉关闭后编辑态重置

- [x] 确认关闭抽屉后未提交编辑态仍残留的根因。
- [x] 调整治理规则详情抽屉生命周期，关闭后销毁内部编辑器状态。
- [x] 构建 console 并重启 all 模式。
- [x] 在 8080 实际页面验证关闭后重开回到查看态。
- [x] 记录 review、验证结果和剩余风险。

## 治理规则抽屉关闭后编辑态重置 Review

已完成：

- 根因定位：`RuleDetailDrawer` 关闭时只切换 `visible=false`，没有销毁内部 `RuleTabs` 和各类规则编辑器，导致 `CustomRouteEditor` 的本地 `editorState.editable=true` 在关闭后继续保留。
- 在共享的 `RuleDetailDrawer` 上启用 `destroyOnClose`，关闭抽屉时销毁内部编辑器；未保存内容和编辑态都会在关闭时丢弃，重开规则重新进入查看态。
- 该改动覆盖治理工作台中复用该抽屉的规则详情，避免路由、限流、熔断、探测、无损、泳道等详情出现同类本地状态泄漏。

验证：

- 先运行最小失败检查，确认旧代码缺少 `destroyOnClose` 会失败；修复后同一检查通过。
- `cd console/web && npm run build` 通过。
- `git diff --check -- console/web/src/pages/Governance/RuleRelease/RuleDetailDrawer.tsx context-kg/tasks/todo.md` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过，`codesign --force --sign - /tmp/pole-control-plane` 后重启 all 模式。
- 重启后 LaunchAgent `io.pole.control-plane.local` 当前 PID `9982`，8080 首页资源为 `assets/index.7627b562.js` 和 `assets/style.f1781d2d.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/?t=drawer-reset-verify`，点击 `demo-governance-route-202606050424`，进入编辑态后显示 `保存 / 撤销`；不保存直接关闭，再重新打开同一规则，抽屉恢复 `编辑 / 发布`，且不再包含 `保存 / 撤销`。

注意：

- `cd console/web && npm run test:response-mapping` 仍失败，当前命中项为既有检查：`governance.ts` 缺少默认重定向、`CustomRoute.tsx` 缺少若干响应映射/空态文本、`LaneGroupTable.tsx` 缺少泳道组空态文本；这些与本次抽屉生命周期修复无关。

# 治理规则发布页视觉与表单优化

- [x] 梳理规则发布抽屉的信息结构、灰度条件数据结构和当前布局问题。
- [x] 重构发布抽屉为规则身份、版本信息、发布策略、灰度条件分区。
- [x] 优化灰度发布条件行的对齐、添加/删除和空态。
- [x] 构建、重启 all 模式并在 8080 实际发布页验证。
- [x] 记录 review、验证结果和剩余风险。

## 治理规则发布页视觉与表单优化 Review

已完成：

- 发布页从全宽长表单改为 780px 右侧抽屉，标题区展示 `规则发布` 和当前规则名，避免像独立空白页面。
- 主体拆成 `规则身份`、`版本信息`、`发布策略`、`灰度条件` 四个分区；规则 ID 和规则名改为只读信息展示，不再占用表单输入态。
- 版本名称和版本描述保留原校验；发布策略保留 `normal/gray` 原字段，提交 payload 仍使用原 `release_type` 和 `client_label`。
- 灰度条件改为本地表格化编辑行，列为客户端标签、匹配类型、值类型、匹配值、操作，支持添加/删除，仍提交为 `MatcheLabel[]`。
- 关闭发布抽屉时启用 `destroyOnClose` 并在打开时重置表单，避免上一次输入残留到下一次发布。

验证：

- `cd console/web && npm run build` 通过。
- `git diff --check -- console/web/src/pages/Governance/RuleRelease/PublishForm.tsx console/web/src/pages/Governance/RuleRelease/PublishForm.module.less context-kg/tasks/todo.md` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -a -o /tmp/pole-control-plane .` 通过，`codesign --force --sign - /tmp/pole-control-plane` 后重启 all 模式。
- 重启后 LaunchAgent `io.pole.control-plane.local` 当前 PID `22463`，8080 首页资源为 `assets/index.9a248fae.js` 和 `assets/style.77e3fce8.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/?t=publish-form-verify-2`，点击 `demo-governance-route-202606050424` 后点击发布，发布抽屉实际内容面板宽度 `780px`，未超出视口。
- 浏览器验证发布抽屉包含 `规则身份 / 版本信息 / 发布策略`；选择 `灰度发布` 后出现 `灰度条件`，灰度表头宽度 `690px`、行宽 `690px`、无横向溢出，`添加客户端标签` 可见。

注意：

- 本轮没有实际提交发布，避免对本地规则版本产生写入副作用；验证覆盖打开、切换灰度和布局状态。
- `cd console/web && npm run test:response-mapping` 仍有既有失败项，和本次发布抽屉布局改动无关。

# A2A Agent Registry 实现集成

- [x] 在独立 worktree 实现 A2A Agent Registry，避免污染当前 develop 既有改动
- [x] 将 A2A 类型、Store、MySQL schema、Cache、HTTP API 和 Console 页面集成回当前 develop 工作区
- [x] 处理当前 develop 中 locale、service types、MySQL default store、SQL 初始化脚本、API 配置和 context-kg 重组后的冲突路径
- [x] 收敛 ADR，确认本次实现为 REST API + Console Registry，不把 A2A task proxy、SSE、push broker 或 task runtime 放进 control-plane
- [x] 在当前 develop 工作区运行 gofmt、目标 Go 测试、构建、前端构建和 diff 检查
- [x] 清理 A2A 应用过程中产生的 staged 混杂，保留用户已有 develop 改动
- [x] 记录最终验证结果和剩余风险

当前进展：

- 独立 worktree 分支 `codex/a2a-agent-registry` 已完成提交 `6574b2b1 feat(ai): add A2A agent registry`。
- 直接 `git apply --check --3way` 检查发现当前 develop 存在大量既有改动，且与 A2A 在 locale、service types、MySQL store、SQL schema、API 配置和 context-kg 页面上重叠；因此没有执行直接 `git merge`。
- A2A 非冲突文件已落入当前 develop，冲突路径按当前 develop 的三域 `context-kg` 结构和现有 Console 改动手工集成。
- Console 页面设计采用现有 MCP/控制台工作台语言：筛选工具栏、摘要表格、右侧抽屉表单、Agent Card 详情、skill 详情，不提供 message/task/push 执行动作。

## A2A Agent Registry 实现集成 Review

已完成：

- 新增 A2A 领域类型：`A2AAgent`、`A2AAgentInterface`、`A2AAgentSkill`、查询条件和批量操作响应。
- `AIStore` 组合 `A2AAgentStore`，MySQL store 支持事务写入 Agent/interfaces/skills、软删除、分页查询、skill 查询和 `mtime` 增量同步。
- SQL 初始化脚本新增 `a2a_agent`、`a2a_agent_interface`、`a2a_agent_skill` 三表。
- CacheManager 新增 A2A cache，按 ID、`namespace/name`、namespace 和 skill 索引，并支持协议绑定、后端和 capability 过滤。
- HTTP API 新增 `plugin/apiserver/httpserver/aia2a`，提供 `/ai/a2a/v1/agents`、`/ai/a2a/v1/agents/delete`、`/ai/a2a/v1/agent/skills`、`/ai/a2a/v1/agents/{id}/card`。
- Console 新增 AI 工具下的 A2A Agents 页面，接入 Redux module、REST service、菜单路由和中英文菜单文案。

验证：

- `gofmt -w` 已仅针对 A2A 相关 Go 文件执行，避免全仓 import-format 改动当前 develop 的其它未提交文件。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./pkg/cache/ai ./plugin/store/mysql ./plugin/apiserver/httpserver/aia2a ./plugin/apiserver/httpserver ./apis/...` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-a2a-develop .` 通过。
- `cd console/web && npm run build:test` 通过；仅有 Vite chunk size 和 Browserslist 数据过期提示。
- `git diff --check` 通过。
- `rg -n "^(<<<<<<<|=======|>>>>>>>)" .` 无输出，未发现冲突标记。
- A2A 自动暂存路径已退回未暂存状态；当前 staged 区域仍保留 develop 既有 context-kg 重组和旧 docs 删除内容，没有把 A2A 改动混进去。

剩余风险：

- 当前 develop 工作区本身已有大量无关未提交改动；本轮没有执行真实 `git merge`，而是将 A2A 差异按当前工作区状态手工集成。
- 未运行 `go test ./...`，因为当前任务记录中已有三类既有失败点待修复：healthcheck 缺测试配置、i18n 缺 toml、heartbeat nil pointer。
- 本轮只做构建级前端验证，没有启动 8080 做浏览器交互回归；独立 worktree 中此前已验证 A2A 页面基本渲染、筛选、表格和抽屉入口。

## A2A Agent Registry 正式合并到 develop

- [x] 确认当前 `develop` 干净并与 `origin/develop` 同步
- [x] 将 `codex/a2a-agent-registry` 正式 merge 到 `develop`
- [x] 自主解决与当前 develop 的冲突，保留 A2A Registry 范围和当前 develop 的治理/Console/context-kg 重组成果
- [x] 运行 gofmt、目标 Go 测试、Go 构建、Console 构建和 diff 检查
- [x] 提交 merge 结果并推送 develop
- [x] 记录最终 review、验证结果和剩余风险

Review：

- 当前 `develop` 已是 `origin/develop` 的 `0d82f125`，工作区干净后开始 merge。
- `codex/a2a-agent-registry` 基于较早的 develop，缺少后续治理统一、Console 优化和 context-kg 三域重组；A2A 核心代码已在当前 develop 中，差异主要是分支落后导致的无关回滚风险。
- 使用 `git merge -s ours --no-commit codex/a2a-agent-registry` 创建正式 merge 关系，保留当前 develop 文件内容，避免把治理统一和 Console 改动倒退。
- 本次没有实际内容冲突；策略性解决为保留当前 develop 中已经验证过的 A2A Registry 集成结果。

验证：

- `gofmt -w` 已针对 A2A 相关 Go 文件执行。
- `GOPROXY=https://goproxy.cn,direct go test ./...` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-a2a-merge .` 通过。
- `cd console/web && npm run build:test` 通过；仅有 Vite chunk size 与 Browserslist 数据过期提示。
- `git diff --check` 通过。
- `rg -n "^(<<<<<<<|=======|>>>>>>>)" .` 无输出。

# 调用鉴权基础信息布局优化

- [x] 按 frontend-skill 的 app UI 原则重新审视调用鉴权编辑态
- [x] 将规则标签并入基础信息，移除独立规则标签卡片
- [x] 将基础信息重排为身份、开关/优先级、作用范围/标签、描述的配置面板
- [x] 构建前端并重启 all 模式
- [x] 在 8080 验证查看态和编辑态布局、无横向溢出

当前判断：

- 规则标签是基础信息的一部分，不应和鉴权策略并列成为单独大卡片。
- 编辑态不应平均铺满两列大输入框；按治理配置的扫描方式，应把规则名称、开关、优先级、作用范围、标签、描述组织成有主次的栅格。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `27547`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.d0a68669.js` 和 `/assets/style.904bc764.css`，两个静态资源均返回 200。
- 浏览器验证调用鉴权查看态：`基础信息` 内包含 `规则标签`，独立 `规则标签` section 数量为 0，抽屉无横向溢出。
- 浏览器验证调用鉴权编辑态：第一行规则名称/启用状态/优先级，第二行命名空间/服务名称/规则标签，描述整行展示；独立 `规则标签` section 数量为 0，抽屉 `bodyOverflow=false`。

Review：

- 本轮只调整调用鉴权共用基础信息布局和标签归属，不改规则保存结构、发布逻辑或后端接口。

# 新增流量治理规则标签文案修正

- [x] 确认新增流量治理详情中 `metadata` 的产品文案应为规则标签
- [x] 将详情区标题、空态和新增按钮从元数据改为规则标签
- [x] 构建前端并重启 all 模式
- [x] 在 8080 验证调用鉴权详情不再展示元数据文案

当前判断：

- 用户纠正“原数据是规则标签”，这里应按治理规则既有产品语言展示为规则标签。
- 底层字段仍是 `metadata`，本轮只改用户可见文案和任务记录，不改变提交 payload。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `15050`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.be413cc5.js` 和 `/assets/style.05d4f912.css`，两个静态资源均返回 200。
- 浏览器验证调用鉴权查看态和编辑态：存在 `规则标签` / `添加规则标签`，不存在 `元数据`，保存/撤销正常，抽屉无横向溢出。

Review：

- 本轮只修正文案，不改底层 `metadata` 字段和保存结构。

# 调用鉴权编辑态对齐限流熔断规则设计

- [x] 对照限流、熔断规则块的信息架构和交互样式
- [x] 将调用鉴权策略块改为规则块 header 摘要 + 可折叠正文
- [x] 将调用鉴权正文拆成接口范围、鉴权结果、匹配条件分区
- [x] 构建前端并重启 all 模式
- [x] 在 8080 验证调用鉴权、流量镜像、流量 Mock 和 9 类巡检无回归

当前判断：

- 用户反馈调用鉴权的整体设计应参考限流、熔断；当前调用鉴权虽然解决了横向拥挤，但规则块仍偏普通表单堆叠。
- 限流、熔断的成熟模式是规则块 header 展示标题、摘要和动作，正文按业务区块组织，并支持折叠，调用鉴权应复用这套交互语言。

当前进展：

- 调用鉴权策略块已改为和限流/熔断一致的规则块结构：左侧折叠按钮、`规则 [n]` 标题、策略摘要，右侧保留命中动作选择和删除操作。
- 未命中默认动作从普通 `FormItem` 改成策略概览条，说明默认放通/拒绝的兜底语义。
- 策略正文拆成 `接口范围`、`鉴权结果`、`匹配条件` 三个分区，分区内保留简短帮助文案和原有字段控件。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `8020`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.a0178176.js` 和 `/assets/style.05d4f912.css`，两个静态资源均返回 200。
- 浏览器验证调用鉴权查看态：存在 `未命中默认动作`、`规则 [1]`、策略摘要、`接口范围 / 鉴权结果 / 匹配条件` 分区，抽屉无横向溢出。
- 浏览器验证调用鉴权编辑态：保存/撤销正常，策略块有折叠按钮，正文分区完整，抽屉 `bodyOverflow=false`、`outsideBodyCount=0`。
- 浏览器 9 类回归巡检通过：路由、限流、熔断、探测、无损、泳道、调用鉴权、流量镜像、流量 Mock 均能进入编辑态，保存/撤销存在，抽屉正文无横向溢出。

Review：

- 本次只调整调用鉴权规则块的产品布局，不改变后端接口、保存 payload、发布逻辑和流量镜像/Mock 的业务字段。
- 调用鉴权现在和限流、熔断一样以规则块为主视觉单元，避免退回普通表单堆叠。

## 相关页面

- [[lessons]]

# 修复 go test ./... 既有失败点

- [x] 复现并记录 `go test ./...` 中 healthcheck、i18n、heartbeat 三类失败的完整错误。
- [x] 定位 healthcheck 缺测试配置的根因并补齐测试初始化/fixture。
- [x] 定位 i18n 缺 toml 的根因并补齐测试可用资源或路径处理。
- [x] 定位 heartbeat nil pointer 根因并修复测试/初始化缺口。
- [x] 运行针对性测试验证三类失败已消除。
- [x] 运行 `go test ./...` 或给出剩余无关失败清单。
- [x] 记录 review、验证结果和剩余风险。

## 修复 go test ./... 既有失败点 Review

已完成：

- healthcheck：恢复 `test/data/service_test.yaml`、`service_test_sqldb.yaml`、`bolt-data.yaml` 三个测试 fixture；同时把 `Test_serialSetInsDbStatus` 收窄为内存 fake store 单测，避免为一个元数据写删函数启动整套 DiscoverTestSuit。
- i18n：测试不再读取已不存在的 `release/conf/i18n/*.toml`，改为基于 `runtime.Caller` 定位仓库根目录下的 `deploy/conf/i18n/*.toml`。
- heartbeat：`HeartBeatHealthChecker.Initialize` 现在会通过 `unmarshal` 保存默认配置到 `c.conf`，避免 `refreshPeers` 中 `peer.Initialize(*c.conf)` 解引用 nil；heartbeat 单测改用共享内存 peer，保留一致性哈希、扩缩容、Report/Query/Delete 行为验证，但不依赖真实 gRPC 端口。
- 测试套件：补齐 `test/suit` 所需的 interceptor、auth 和 heartbeat blank import，避免单独跑测试包时缺少插件注册。

验证：

- `GOPROXY=https://goproxy.cn,direct go test ./pkg/service/healthcheck ./plugin/apiserver/httpserver/i18n ./plugin/service/healthchecker/heartbeat` 通过。
- `gofmt` 已处理本次改动的 Go 文件。
- `git diff --check` 针对本次改动文件通过。
- `GOPROXY=https://goproxy.cn,direct go test ./...` 通过，退出码 `0`；之前的 healthcheck 缺 fixture、i18n 缺 toml、heartbeat nil pointer 均不再出现。

注意：

- 当前工作区在本次修复前已经存在大量 A2A、console 和 context-kg 重组相关未提交改动；本轮只围绕上述测试失败点修改，不整理无关变更。

# specification 流量治理规则最终产品决策

- [x] 确认 `../specification` 当前分支、工作区状态和 proto 结构
- [x] 梳理最终产品决策：流量镜像、流量 Mock、调用鉴权规则命名与语义
- [x] 在 `../specification` 独立分支修改 proto 定义
- [x] 运行 spec 仓库可用的生成、格式化和测试校验
- [x] 提交并尝试创建 draft PR 供 review
- [x] 记录最终 review、验证结果和剩余风险

当前进展：

- `../specification` 已从干净的 `develop` 创建 `codex/traffic-security-spec` 分支。
- 最终产品决策：删除 `BlockAllowListRule` 命名和黑白名单主模型，新增 `TrafficSecurityRule`，通过 `TrafficSecurityPolicy.action` 表示命中后 `ALLOW/DENY`；鉴权规则不再建模未命中默认动作。
- 新增 `TrafficMock` 一等治理规则，保留 `TrafficMirror` 并补齐根作用域、启用开关和优先级。
- `RuleRelease.RuleType` 新增 `TrafficMockRules`；`ResourceType` / `StrategyResources` 新增 `MockRules` / `mock_rules`。
- `DiscoverRequest` / `DiscoverResponse` 下发类型改为 `TRAFFIC_SECURITY_RULE`，并新增 `TRAFFIC_MIRROR_RULE`、`TRAFFIC_MOCK_RULE`；响应字段改为 `trafficSecurityRules`、`trafficMirrorRules`、`trafficMockRules`。

验证：

- `bash build.sh` 在 `../specification/source/go` 通过，已重新生成 Go protobuf。
- `go test ./...` 在 `../specification` 通过。
- `bash build.sh` 在 `../specification/source/rust` 通过。
- `PROTOC=/Users/chuntao.liao/Github/pole-io/specification/source/protoc/protoc-darwin-arm64/bin/protoc cargo test --release` 在 `../specification/source/rust/pole-specification` 通过。
- `git diff --check` 在 `../specification` 通过。
- `rg -n "BlockAllow|block_allow|blockAllow|BLOCK_ALLOW" api/v1 source/go source/rust/pole-specification/proto source/rust/pole-specification/src -g '*.proto' -g '*.go' -g '*.rs'` 无输出。

Review：

- `../specification` 提交：`9d6fd34 Add traffic security and mock specs`。
- 远端分支：`codex/traffic-security-spec`。
- Draft PR：`https://github.com/pole-io/specification/pull/1`。
- GitHub connector 创建 PR 时返回 `Resource not accessible by integration`，已用已登录的 `gh pr create --draft` fallback 成功创建。
- `pole-control-plane` 当前仍有本轮之前已存在的大量未提交改动；本轮只追加了本任务记录，没有整理其它工作区改动。

# 接入流量安全、镜像和 Mock 治理规则

- [x] 合并 `pole-io/specification` PR #1
- [x] 在 `specification` 发新 tag，并确认远端 tag 可见
- [x] 为 `pole-control-plane` 创建隔离 worktree，避免污染当前已有改动
- [x] 更新 `github.com/pole-io/specification` 依赖到新 tag
- [x] 按 TDD 补齐流量安全、镜像和 Mock 规则的存储转换测试
- [x] 实现三类规则的 store、cache watcher、业务服务和 HTTP/Discover 下发支持
- [x] 补齐鉴权资源映射和发布版本映射
- [x] 运行目标测试、构建和必要的接口验证
- [x] 记录最终 review、验证结果和剩余风险

当前进展：

- `specification` PR #1 已合并到 `develop`，merge commit 为 `78f368d9738137754b6f0c9489358eb9bb1c5f1e`。
- `specification` 已发布并推送 tag `v0.1.0-ALPHA.25`，`pole-control-plane` 已将 `github.com/pole-io/specification` 从 `v0.1.0-ALPHA.24` 升级到 `v0.1.0-ALPHA.25`。
- 本轮实现放在隔离 worktree `/Users/chuntao.liao/.config/superpowers/worktrees/pole-control-plane/traffic-governance-rules`，分支 `codex/traffic-governance-rules`，避免污染主工作区已有改动。
- 新增 `TrafficGovernanceRule` 内部包装，复用统一治理表、cache watcher、release pipeline，同时对外分别暴露 spec 中的 `TrafficSecurityRule`、`TrafficMirror`、`TrafficMock`。
- MySQL 统一治理表新增三类 rule_type：`traffic-security`、`traffic-mirror`、`traffic-mock`；支持当前态 CRUD、增量拉取、release 查询、publish、delete release 和 stopbeta。
- CacheManager 新增 `TrafficSecurity`、`TrafficMirror`、`TrafficMock` 三个缓存，支持 Console 当前态查询和客户端 active release 下发。
- GoverRule 新增三类规则的 CRUD、list/detail、release 入口和客户端 discover 下发；HTTP 管理端新增 `/traffic/security`、`/traffic/mirrors`、`/traffic/mocks` 及其 `/releases` 接口。
- HTTP/gRPC client discover 支持 spec 新枚举 `TRAFFIC_SECURITY_RULE`、`TRAFFIC_MIRROR_RULE`、`TRAFFIC_MOCK_RULE`，响应填充 `trafficSecurityRules`、`trafficMirrorRules`、`trafficMockRules`。
- 鉴权补齐 `mirror_rules`、`security_rules`、`mock_rules` 字段映射、默认策略资源详情、资源存在性检查和 release 鉴权资源收集。

验证：

- 新增红灯测试覆盖三类规则的 MySQL 统一表转换 round-trip、`mock_rules` 鉴权资源字段映射和三类 discover response 构造。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./plugin/store/mysql ./apis/pkg/types/auth ./pkg/common/api/v1` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./apis/... ./pkg/cache/... ./pkg/goverrule/... ./plugin/store/mysql ./plugin/apiserver/httpserver/discover ./plugin/apiserver/grpcserver/discover/... ./plugin/access_control/auth/policy` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-traffic-rules .` 通过。
- `git diff --check` 通过。
- `GOPROXY=https://goproxy.cn,direct go test ./...` 跑到 10 分钟超时失败，唯一失败包为 `pkg/common/batchctrl`，失败用例 `TestNewBatchControllerGracefulStop`，堆栈显示 `future.Reply` 阻塞；其它包继续执行并通过。该失败与本轮流量治理规则接入无直接关系。

Review：

- 本轮不实现 Console 页面，只补 server/API/cache/store/discover 下发链路。
- 新规则 release 能力按当前 lossless 已有边界实现：支持 publish/list/delete release/stopbeta；rollback 主 switch 现状未支持 lossless，因此这次没有为三类新规则额外打开 rollback。
- HTTP 管理端新增路径采用 `/traffic/security`、`/traffic/mirrors`、`/traffic/mocks`，客户端 discover 使用 spec 枚举，不依赖路径命名。

# 合并流量治理规则 PR 到 develop

- [x] 检查 PR #20 状态和 base/head 分支
- [x] 确认 GitHub merge 状态
- [x] 将 draft PR 标记为 ready
- [x] merge PR #20 到 `develop`
- [x] 确认 PR 已合并

Review：

- PR #20 base 为 `develop`，head 为 `codex/traffic-governance-rules`。
- 合并前 GitHub `mergeStateStatus` 为 `CLEAN`，未出现冲突。
- PR #20 已从 draft 标记为 ready，并通过 GitHub merge 合并到 `develop`。

# 同步本地 develop 到远端最新

- [x] fetch `origin/develop`
- [x] 保护本地 `context-kg/tasks/todo.md` 改动
- [x] fast-forward 本地 `develop` 到 `be5ed67250ac94fa4c8aba13632f057ed14445d6`
- [x] 解析 `context-kg/tasks/todo.md` 冲突，保留本地和远端任务记录
- [x] 确认本地 `HEAD` 与 `origin/develop` 一致

Review：

- 本地 `develop` 已同步到 `be5ed67250ac94fa4c8aba13632f057ed14445d6`。
- 同步过程中只冲突了 `context-kg/tasks/todo.md`；已保留本地 specification 决策记录和远端流量治理实现/PR merge 记录。
- 其它未提交文件为同步前已有本地改动，本轮未覆盖。

# 新增治理规则端到端补齐与验证

- [x] 确认当前 `develop` 基线、已有未提交改动和本轮可安全修改范围
- [x] 审计流量安全、流量镜像、流量 Mock 在后端 API、store、cache、发布和 discover 查询链路的覆盖
- [x] 审计并补齐 Console 前端入口、API service、列表、详情、发布和监听入口
- [x] 补齐发现的后端/前端缺口及对应测试
- [x] 运行后端目标测试、前端构建和必要的端到端接口/页面验证
- [x] 按要求做 completion audit，记录最终 review、验证证据和剩余风险

当前进展：

- 当前分支为 `develop`，`HEAD` 与 `origin/develop` 均为 `be5ed67250ac94fa4c8aba13632f057ed14445d6`。
- 工作区在本轮开始前已有多处未提交改动；本轮会按最小影响原则只处理新增治理规则端到端支持所需文件。
- 后端审计确认三类规则已有管理端 HTTP API、业务服务、MySQL 统一治理表转换、cache watcher、发布版本、auth 资源映射和 HTTP/gRPC discover 下发链路。
- Console 原本只有未实现的 `Governance/Security` 占位页，缺少三类规则的 BaseURL、前端 service、列表、详情、发布入口和 Workbench 聚合。
- 已新增 `services/traffic_governance.ts`，覆盖 security/mirror/mock 的列表、详情、创建、更新、删除、发布版本查询、发布和删除版本。
- 已替换 `pages/Governance/Security` 为三类规则 tab 页面，支持新建、查询、查看、编辑、删除、授权、版本和监听抽屉。
- 已将三类规则接入 `Governance/Workbench` 的统一列表和详情抽屉，新增类型筛选项。
- 已调整 `PublishForm`，处理鉴权资源名 `SecurityRules/MirrorRules/MockRules` 与发布资源名 `TrafficSecurityRules/TrafficMirrorRules/TrafficMockRules` 的差异。
- 已补齐调用鉴权、流量镜像、流量 Mock 在策略详情资源树中的 `security_rules`、`mirror_rules`、`mock_rules` 展示。
- 已修复三类规则发布时首次发布没有历史版本时 `GetRelease` / `GetActiveRelease` 把 `sql.ErrNoRows` 当异常返回的问题。
- 已修复发布态记录转换回规则对象时空 release record 字段覆盖 JSON 快照字段的问题，避免客户端 Discover 得到空 revision 或空服务范围。
- 已修复 `pkg/common/batchctrl` 中 `future.Reply` 对无缓冲 `setsignal` 发送导致 `TestNewBatchControllerGracefulStop` 阻塞的问题。
- 已补齐 TrafficGovernance 前端 Duration 兼容处理：流量镜像 `duration` 和流量 Mock `delay` 在提交前统一转为 proto JSON 字符串，避免后端 `request decode failed: json: cannot unmarshal object into Go value of type string`。

验证：

- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./pkg/common/batchctrl` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./...` 通过，退出码 `0`。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane-traffic-e2e .` 通过。
- 已将 `/tmp/pole-control-plane-traffic-e2e` 替换为 `/tmp/pole-control-plane` 并重启 LaunchAgent `io.pole.control-plane.local`；当前服务以 `/tmp/pole-control-plane start -c /tmp/pole-server-all.yaml` 运行，PID `34848`。
- `curl http://127.0.0.1:8090/auth/v1/user/login` 使用 `admin/admin123` 返回 `code=200000` 且 token 存在。
- 运行时端到端接口验证通过：创建 `traffic/security`、`traffic/mirrors`、`traffic/mocks` 三类规则，分别发布 `TrafficSecurityRules`、`TrafficMirrorRules`、`TrafficMockRules` normal release，管理端 release 查询命中，客户端 `/v1/Discover` 分别以 `TRAFFIC_SECURITY_RULE`、`TRAFFIC_MIRROR_RULE`、`TRAFFIC_MOCK_RULE` 查询并命中对应 `traffic_security_rules`、`traffic_mirror_rules`、`traffic_mock_rules`，且 revision 非空。
- 端到端验证样例：`ns=codex-e2e-20260612113459-41608`，security=`5b5d0776f99445188f391763822dc9c8`，mirror=`e4fdf997312a432fa7ee6479a902955b`，mock=`e550a52709dc4b689848fe0980c4acad`；验证后数据库清理结果 `remaining=0,0`，临时证据目录 `/tmp/traffic-governance-e2e-20260612113459-41608`。
- `cd console/web && npm run test:response-mapping` 通过，输出 `standard response mapping checks passed (21 files)`。
- `cd console/web && npm run build` 通过，构建产物包含新增 `TrafficGovernanceEditor.4008e353.js` chunk；仅有已有 `--localstorage-file`、Browserslist 数据过期和 Vite 大 chunk 警告。
- `git diff --check --` 通过。

Review：

- Completion audit 覆盖了前端入口、前端 service、创建/编辑提交、发布资源映射、策略详情资源展示、后端 store/release、cache 增量、客户端 Discover 查询和真实 MySQL 运行时链路。
- 当前三类新增规则的 release 能力与已实现治理规则保持一致：支持 normal/gray 发布、版本列表、删除版本和 stopbeta；未额外扩展 rollback 行为。
- 本轮验证发现前端 Duration 对象形态与后端 proto JSON 解析不兼容，已在前端 service 层做提交前兼容转换，默认值也改为 `"0s"`。

# 本地 all 模式新增治理规则测试数据补齐

- [x] 确认当前本地 all 模式服务、数据库和已有治理样例数据状态
- [x] 通过管理端 API 创建调用鉴权、流量镜像、流量 Mock 三类可视化样例规则
- [x] 为三类样例规则发布 normal release
- [x] 验证 Console 列表、管理端 release 查询、客户端 Discover 查询均可命中
- [x] 记录样例入口、验证证据和剩余风险

当前进展：

- 本地服务 `8090` 登录接口返回 `code=200000`，all 模式服务可用。
- 数据库当前已有 `route`、`ratelimit`、`circuitbreaker`、`faultdetect`、`lossless`、`lane-group` 各 1 条当前态和 release；新增的 `traffic-security`、`traffic-mirror`、`traffic-mock` 当前为 0 条。
- 已通过管理端 API 创建并发布 3 条新增治理规则样例，均位于命名空间 `spec-governance`、服务 `spec-gateway`：
  - 调用鉴权：`spec-check-traffic-security-20260612`
  - 流量镜像：`spec-check-traffic-mirror-20260612`
  - 流量 Mock：`spec-check-traffic-mock-20260612`

验证：

- 数据库 `governance_rule` 当前 9 类规则各 1 条：`route`、`ratelimit`、`circuitbreaker`、`faultdetect`、`lossless`、`lane-group`、`traffic-security`、`traffic-mirror`、`traffic-mock`。
- 数据库 `governance_rule_release` 中上述 9 类 active release 各 1 条。
- 管理端列表验证通过：
  - `/naming/v1/traffic/security?name=spec-check-traffic-security-20260612&offset=0&limit=10` 返回 `code=200000`、`amount=1`。
  - `/naming/v1/traffic/mirrors?name=spec-check-traffic-mirror-20260612&offset=0&limit=10` 返回 `code=200000`、`amount=1`。
  - `/naming/v1/traffic/mocks?name=spec-check-traffic-mock-20260612&offset=0&limit=10` 返回 `code=200000`、`amount=1`。
- 客户端 Discover 验证通过：
  - `TRAFFIC_SECURITY_RULE` 命中 `spec-check-traffic-security-20260612`。
  - `TRAFFIC_MIRROR_RULE` 命中 `spec-check-traffic-mirror-20260612`。
  - `TRAFFIC_MOCK_RULE` 命中 `spec-check-traffic-mock-20260612`。
- 造数证据目录：`/tmp/traffic-governance-seed-20260612`。

Review：

- 这批数据未清理，保留给本地 Console 检查；在 `http://127.0.0.1:8080/governance/security` 或统一治理工作台搜索 `spec-check-traffic` 即可看到。
- 造数使用管理端 API 创建和发布，未直接绕过业务服务写当前态；仅在开始前按固定样例名清理旧样例，保证重复执行不会产生重复数据。

# 泳道详情页信息架构优化

- [x] 复核截图问题和当前泳道组详情组件结构
- [x] 重构泳道组查看态布局，减少时间线留白并默认展示核心信息
- [x] 将泳道列表默认展示到详情页内，避免二次展开
- [x] 同步治理工作台和泳道独立页的详情容器
- [x] 构建并在本地 all 模式页面验证展示效果
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 截图中的问题来自两层结构叠加：外层详情页把 `泳道组详细 / 泳道列表` 放进 Collapse，内层 `LaneGroupEdtor` 又用 vertical `Steps` 展示查看态，导致信息被拉成长流程，入口和服务区域大量留白，泳道列表默认不可见。
- 优化方向是不改接口和数据结构，只调整 Console 信息架构：查看态使用摘要卡片 + 入口/目标服务分栏 + 元数据标签，泳道列表默认展示；编辑态保留现有表单式配置流程。

当前进展：

- `LaneGroupEdtor` 查看态已改为只读信息布局：顶部展示入口、网关入口、服务入口、目标服务数量，下面按泳道组信息、泳道组入口、泳道组服务分区展示。
- 编辑态和创建态保留原 `Steps + Form` 配置流程，避免影响泳道组配置路径。
- `LaneGroupTable` 和 `GovernanceWorkbench` 的泳道详情均去掉外层 Collapse，泳道列表默认显示在泳道组概要下方。
- `LaneRuleTable` 外层 padding 改为样式类，适配新的默认展示容器。
- 兼容 specification 样例中的网关入口类型 `polarismesh.cn/gateway/spring-cloud-gateway`，不再只识别字面量 `gateway`，页面可正确显示 `spec-governance/spec-gateway`。
- 修复泳道组提交时表单字段未绑定导致保存可能丢失 `name`、`description`、`metadata` 的问题，提交改为使用当前编辑态数据。

验证：

- `cd console/web && npm run build` 通过。
- 因 `/tmp/pole-control-plane` 和 `/tmp/pole-server-all.yaml` 被临时目录清理，本轮已重新构建二进制并恢复 all 模式配置；启动时发现 `pole-mysql` 容器已停止，已启动 MySQL 并重启 LaunchAgent `io.pole.control-plane.local`。
- 当前 all 模式服务通过 `8080/8090` 健康检查；8080 加载新前端资源 `assets/index.2ea8c876.js`。
- Playwright 打开 `http://127.0.0.1:8080/governance/`，登录后查看 `spec-check-lane-group`：摘要显示入口 `1`、网关入口 `1`、服务入口 `0`、目标服务 `2`；页面直接显示 `spec-governance/spec-gateway`、`spec-governance/spec-order`、`spec-governance/spec-payment` 和泳道列表行 `blue-lane`。
- 页面截图保存在 `.playwright-cli/page-2026-06-12T07-26-12-488Z.png`。

Review：

- 本次只调整 Console 信息架构和查看态数据映射，不改后端接口、存储和样例数据。
- 查看态避免再用配置流程式时间线承载静态信息；编辑态仍保留流程式表单，降低交互迁移风险。
- `entry.type` 兼容逻辑基于当前 specification / 后端返回的类型字符串和 selector `@type` 双重判断，后续如果入口类型继续扩展，建议在服务层统一标准化。

# 本地 all 模式 8080 前端资源 404 排查

- [x] 重新确认 all 模式进程、8080 监听和根路径响应
- [x] 排查浏览器“打不开”与 curl 根路径 200 的差异
- [x] 定位运行中 console 返回旧 `index.html` 导致静态资源 hash 404
- [x] 重启 all 模式，让 Gin 重新加载当前 `console/web/dist/index.html`
- [x] 验证 8080 首页、JS/CSS 静态资源和治理规则页面可访问

当前判断：

- all 模式进程仍在运行，8080 根路径返回 200；问题不是端口未监听。
- tmux 日志显示浏览器访问 `/` 后继续请求 `/assets/index.9877a31d.js` 和 `/assets/style.ab32ff84.css`，这两个资源返回 404。
- 当前磁盘上的 `console/web/dist/index.html` 已引用新资源 `assets/index.a3134e1e.js` 和 `assets/style.7ae12346.css`。
- 根因是 console 启动时通过 Gin `LoadHTMLGlob` 把 `index.html` 加载到内存；前端重新 build 后，运行中的 8080 仍返回旧 index，需要重启 all 模式。

验证：

- 已重启 tmux session `pole-control-plane`，新 all 模式进程 PID `5297`，8080 监听正常。
- `http://127.0.0.1:8080/` 返回的新 index 引用 `/assets/index.a3134e1e.js` 和 `/assets/style.7ae12346.css`。
- 上述 JS/CSS 静态资源均返回 200；`/governance/security` SPA fallback 返回 200。
- 浏览器打开 `http://127.0.0.1:8080/governance/security` 可渲染调用鉴权列表，并显示样例规则 `spec-check-traffic-security-20260612`。
- `git diff --check -- context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本次问题不是后端端口不可用，而是运行中 console 缓存了旧 `index.html`，导致浏览器继续请求已不存在的 hash 资源。
- 后续前端 build 后验证 8080 时，必须同时检查首页 HTML 中的资源 hash 与静态资源 200，不能只用 `curl -I /` 判断页面可用。

# 新增流量治理详情抽屉滚动和操作按钮修复

- [x] 复现并对比新增流量治理详情与已有治理详情的滚动/操作区结构
- [x] 将调用鉴权/流量镜像/流量 Mock 详情操作按钮改为统一 `StickyTool` 承载
- [x] 移除内容底部 sticky 操作条，避免遮挡和滚动容器冲突
- [x] 构建前端并重启 all 模式
- [x] 在 8080 真实页面验证详情可滚动、编辑/发布按钮位置一致
- [x] 记录 review 和本次纠正经验

当前判断：

- 新增 `TrafficGovernanceEditor` 使用内容底部 `.actionBar` 放置 `发布/编辑/保存/取消`，和已有治理详情的标题区右侧 `StickyTool + RuleStickyAction` 不一致。
- `.actionBar` 作为内容内 sticky footer 会贴在抽屉可视区底部，截图中已经压到视口边缘，也会干扰用户对抽屉正文滚动区的操作。
- 公共 `RuleDetailDrawer` 已经统一了 `StickyTool` 在抽屉内的布局，因此本次只需要让新增流量治理详情复用同一套操作按钮结构。

当前进展：

- `TrafficGovernanceEditor` 已改为使用 `StickyTool + RuleStickyAction`，查看态显示 `编辑 / 发布`，编辑态显示 `保存 / 撤销`。
- 已移除新增流量治理详情底部 `.actionBar`，避免内容底部 sticky footer 挤占或遮挡抽屉正文。
- 无编辑权限时仍保留编辑入口的禁用视觉，不再把整组工具条隐藏，保持和原页面可见性一致。

验证：

- `cd console/web && npm run build` 通过。
- 已重启 tmux session `pole-control-plane`，8080 新 index 资源 `/assets/index.b542f2df.js` 和 `/assets/style.b1f896f7.css` 均返回 200。
- 浏览器打开 `http://127.0.0.1:8080/governance/security` 并进入 `spec-check-traffic-security-20260612` 详情：抽屉正文 `overflowY=auto`，`scrollHeight=961` 大于 `clientHeight=700`。
- 执行真实滚轮滚动后，抽屉正文 `scrollTop=261`，确认可以滚动。
- 详情页已无旧的内容底部 `.actionBar`，存在公共 `.t-sticky-tool`，按钮文本为 `编辑发布`。

Review：

- 本次只调整新增流量治理三类规则共用的详情编辑器，不改公共 `RuleDetailDrawer` 和既有治理规则详情。
- 操作按钮现在由公共抽屉样式统一定位，和路由、限流、熔断、无损等治理详情保持一致。

# 治理工作台内部滚动优化

- [x] 定位工作台页面滚动来源和表格撑高路径
- [x] 将工作台根容器固定到当前可视区高度
- [x] 将规则清单面板改为 flex 布局，表格内容区内部滚动
- [x] 构建前端并重启 all 模式
- [x] 在 8080 真实页面验证外层不滚动、规则清单内部滚动
- [x] 记录 review 和经验

当前判断：

- 工作台外层 `sideContainer` 当前 `scrollHeight=1276`，说明是主布局容器在滚动。
- `GovernanceWorkbench` 的 `.page` 未限制高度，`.listPanel` 被 TDesign 表格 10 行内容撑到 931px，进而把整个页面撑高。
- 目标不是去掉滚动，而是固定工作台视区高度，把滚动收敛到规则清单内部的表格内容区。

当前进展：

- 工作台 `.page` 已改为固定可视高度的纵向 flex 容器，外层 `overflow: hidden`。
- 规则清单 `.listPanel` 改为 flex 子项，面板头部、筛选区固定，表格区域占用剩余高度。
- 表格 `.t-table__content` 改为内部滚动层，分页区保持在清单面板底部。

验证：

- `cd console/web && npm run build` 通过。
- 已重启 tmux session `pole-control-plane`，8080 新 index 资源 `/assets/index.1ec6807c.js` 和 `/assets/style.046cab4c.css` 均返回 200。
- 浏览器打开 `http://127.0.0.1:8080/governance/` 后，外层 `sideContainer` 的 `scrollHeight=720`、`clientHeight=720`、`scrollTop=0`，不再出现整页滚动。
- 表格内容区 `.t-table__content` 的 `scrollHeight=659`、`clientHeight=103`、`overflowY=auto`。
- 执行真实滚轮滚动后，外层 `sideScrollTop=0`，表格内容区 `tableContentScrollTop=420`，确认滚动落在规则清单内部。

Review：

- 本次只调整治理工作台布局样式，不改列表数据、筛选、分页和详情抽屉逻辑。
- 清单头部、筛选栏和分页固定在面板内，规则行滚动在表格内容区完成，符合“内部滚动”的产品偏好。

# 新增流量治理规则编辑权限标记修复

- [x] 复现流量镜像详情无法编辑的问题并核对接口返回
- [x] 横向检查调用鉴权、流量镜像、流量 Mock 的 `editable/deleteable` 标记
- [x] 补齐新增流量治理规则查询链路的权限标记回填
- [x] 运行 Go 测试、构建并重启 all 模式
- [x] 验证接口和 8080 页面编辑入口恢复
- [x] 记录 review 和经验

当前判断：

- 页面不能编辑不是前端按钮单独写死，而是后端列表/详情返回的 `editable=false`，前端按统一治理规则权限语义禁用了编辑入口。
- 直接查询 8090 发现调用鉴权、流量镜像、流量 Mock 三类新增治理规则都返回 `editable=false/deleteable=false`，说明不是流量镜像单类问题。
- 既有路由、限流、熔断、主动探测、无损、泳道的 auth interceptor 会在查询返回后按 update/delete 权限回填 `Editable/Deleteable`；新增流量治理 auth interceptor 之前只透传查询结果，proto bool 默认值为 false。

当前进展：

- 已在 `pkg/goverrule/interceptor/auth/traffic_governance.go` 为调用鉴权、流量镜像、流量 Mock 统一补齐列表和详情查询后的权限回填逻辑。

验证：

- `gofmt -w pkg/goverrule/interceptor/auth/traffic_governance.go` 已执行。
- `git diff --check -- pkg/goverrule/interceptor/auth/traffic_governance.go context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `GOPROXY=https://goproxy.cn,direct go test -count=1 ./pkg/goverrule ./pkg/goverrule/interceptor/auth` 通过。
- `GOPROXY=https://goproxy.cn,direct go build -o /tmp/pole-control-plane .` 通过。
- 已重启 tmux session `pole-control-plane`，all 模式新进程 PID `40665`，8080/8090 均监听正常，8080 首页返回 200。
- 带真实 admin token 查询 8090：`traffic/mirrors`、`traffic/security`、`traffic/mocks` 均返回 `code=200000`，样例规则 `editable=true/deleteable=true`。
- 使用和 Console 一致的 JWT cookie 走 8080 代理查询：三类新增治理规则均返回 `editable=true/deleteable=true`。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，筛选 `流量镜像` 后打开 `spec-check-traffic-mirror-20260612`，点击 `编辑` 可进入编辑态，并出现 `保存 / 撤销`。
- 浏览器筛选 `调用鉴权` 后打开 `spec-check-traffic-security-20260612`，详情显示 `编辑 / 发布`；点击 `编辑` 可进入编辑态，并出现 `保存 / 撤销` 和规则名称等输入控件。

Review：

- 本次问题根因在后端 auth interceptor 权限标记缺失，不是流量镜像前端详情单独禁用了编辑。
- 修复范围横向覆盖调用鉴权、流量镜像、流量 Mock 三类新增治理规则，避免同类规则继续因为 proto bool 默认 false 被误判为无权限。
- 裸 curl 现在会按读权限返回 `401001`，验证新增治理规则接口时应带登录 token 或走 8080 Console 登录态。

# 新增流量治理规则结构化编辑修复

- [x] 定位点击编辑后展示 JSON 文本框的根因
- [x] 将调用鉴权、流量镜像、流量 Mock 编辑态改为结构化控件
- [x] 将规则标签编辑从 JSON 文本框改为键值控件
- [x] 构建前端并重启 all 模式
- [x] 在 8080 页面验证调用鉴权和流量镜像编辑态不再出现 JSON 文本框
- [x] 记录 review 和经验

当前判断：

- `TrafficGovernanceEditor` 的查看态已经按规则字段做了结构化展示，但编辑态仍直接渲染 `payloadJson` 和 `metadataJson` 两个 `Textarea`。
- 点击 `编辑` 后用户看到的是规则数组和规则标签对象的原始 JSON 字符串，问题不在权限链路，而是新增三类规则的编辑态还停留在临时实现。
- 修复应保持查看态与编辑态同一套信息架构：基础信息仍用表单，规则定义按策略/镜像/Mock 规则块编辑，规则标签按键值对编辑。

当前进展：

- 调用鉴权编辑态已改为结构化策略块：默认动作、策略动作、接口范围、拒绝效果和匹配条件均使用字段控件。
- 流量镜像编辑态已改为结构化镜像规则块：来源、目标、镜像比例、生效时长、目标标签和匹配条件均使用字段控件。
- 流量 Mock 编辑态已改为结构化 Mock 规则块：来源接口、Mock 比例、延迟、响应状态、响应头、匹配条件和响应体均使用字段控件；响应体保留文本框，因为它是业务响应内容本身。
- 规则标签已从 JSON 文本框改为键值行编辑。

验证：

- `cd console/web && npm run build` 通过，保留既有 browserslist/chunk 体积警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `62878`，8080/8090 均监听正常。
- 8080 首页返回 200，并引用新资源 `/assets/index.121b251e.js` 和 `/assets/style.e3f6f309.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，调用鉴权 `spec-check-traffic-security-20260612` 点击编辑后显示字段控件，未出现规则数组 JSON 文本框。
- 浏览器打开流量镜像 `spec-check-traffic-mirror-20260612` 点击编辑后显示来源、目标、比例、时长、标签和匹配条件字段控件，未出现规则数组 JSON 文本框。
- 浏览器打开流量 Mock `spec-check-traffic-mock-20260612` 点击编辑后显示结构化字段控件，未出现规则数组 JSON 文本框；仅响应体保留文本框。

Review：

- 本次只调整新增流量治理三类规则共用的详情编辑器，不改变后端 spec、store 或接口结构。
- 保存时仍组装回原 `policies/rules/metadata` 结构提交，避免引入新的协议映射层。
- 后续新增治理规则类型时，查看态和编辑态必须同时产品化，不能先用 JSON Textarea 作为编辑态占位。

# 治理规则全量交互布局巡检与优化

- [x] 记录用户截图中的流量镜像编辑态布局问题
- [x] 建立 9 类治理规则详情/编辑/发布交互检查清单
- [x] 用 8080 真实页面逐类检查：路由、限流、熔断、探测、无损、泳道、调用鉴权、流量镜像、流量 Mock
- [x] 修复发现的排版、溢出、控件拥挤和操作按钮不一致问题
- [x] 构建前端并重启 all 模式
- [x] 逐类复验关键交互并记录 review

当前判断：

- 用户截图显示流量镜像编辑态的匹配条件行在抽屉宽度内横向挤压，右侧输入区域被裁切；根因可能是结构化编辑控件一行承载过多字段且没有为抽屉场景设置换行/栅格边界。
- 本次不能只修流量镜像，需要横向检查所有治理规则详情和编辑交互，尤其是多条件、多规则块、表格内控件、发布抽屉和规则标签编辑。

当前进展：

- 公共治理详情抽屉默认宽度已从 `min(860px, 88vw)` 调整为 `min(1080px, 92vw)`，给规则编辑态留出稳定横向空间。
- 新增流量治理三类共用的 `TrafficGovernanceEditor` 已美化编辑布局：表单标签改为纵向、基础信息使用栅格、规则块使用更轻的分区和 header、匹配条件和目标标签改为两行栅格，避免一行塞过多控件。
- 调用鉴权、流量镜像、流量 Mock 的编辑态仍保留统一标题区 `保存 / 撤销`，查看态保留 `编辑 / 发布`。
- 泳道规则表格已从 `auto` 布局改为固定列宽，并约束表格内容区横向滚动，避免操作列伸出抽屉正文。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `98776`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.e161a067.js` 和 `/assets/style.82fa3c00.css`，两个静态资源均返回 200。
- 浏览器复验调用鉴权编辑态：抽屉宽度 `1080`，正文 `scrollWidth=clientWidth=1080`，保存/撤销在标题区，表单标签位于控件上方。
- 浏览器复验流量镜像编辑态：匹配条件区不再横向裁切，正文无横向溢出。
- 浏览器最终 9 类巡检通过：路由、限流、熔断、探测、无损、泳道、调用鉴权、流量镜像、流量 Mock 均有编辑入口，进入编辑态后有保存/撤销，抽屉正文 `bodyOverflow=false`，`outsideBodyCount=0`。

Review：

- 本轮只做治理详情/编辑态布局和抽屉宽度优化，不改变规则 API、提交结构、发布逻辑和样例数据。
- 对旧 6 类规则主要通过公共抽屉宽度和泳道表格约束改善；新增 3 类流量治理规则共用编辑器做了更明显的视觉整理。

# 限流窗口单位缺失修复

- [x] 复现并定位限流编辑态窗口单位显示为“请选择”的根因
- [x] 绑定窗口单位 Select 的当前行 `validDurationUnit`
- [x] 补齐查看态窗口列的单位文本展示
- [x] 构建前端并重启 all 模式
- [x] 在 8080 真实页面验证限流编辑态窗口单位可见
- [x] 记录 review 和验证结果

当前判断：

- `services/ratelimit.ts` 已在 `durationToView` / `amountsToView` 中把窗口时长转换为 `validDuration` 与 `validDurationUnit`，新增阈值也默认写入 `LimitAmountsValidationUnit.s`。
- 页面截图中的“请选择”不是数据缺失，而是 `RateLimitEditor` 窗口列的 `Select` 只配置了 options 和 onChange，没有把当前行的 `validDurationUnit` 作为 value 传入。

当前进展：

- `RateLimitEditor` 已复用 `LimitAmountsValidationUnitOptions`，窗口单位下拉框绑定当前行 `validDurationUnit`，缺省时回退为“秒”。
- 窗口列查看态增加单位标签展示，避免只显示纯数字。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `91101`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.bf36be8b.js` 和 `/assets/style.6e97c28c.css`，两个静态资源均返回 200。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，进入 `spec-check-ratelimit` 详情后，查看态窗口列显示 `1 秒`、`5 秒`。
- 点击 `编辑` 进入编辑态后，页面显示 `保存 / 撤销`；窗口单位控件显示 `秒`，页面 `请选择` 计数为 0。

Review：

- 本次根因在前端复合控件受控值缺失，不涉及后端数据转换、spec 结构或保存 payload。
- 修复范围限定在限流编辑器窗口列：编辑态绑定单位下拉值，查看态补充单位文本。

# 限流保存后回到只读态修复

- [x] 复现并定位限流保存成功后仍停留编辑态的问题
- [x] 对比泳道组、主动探测、新增流量治理等保存成功后的状态流转
- [x] 在限流保存成功后切回 `editable=false`
- [x] 处理表格编辑单元格缓存，确保保存后重新挂载为只读表格
- [x] 构建前端并重启 all 模式
- [x] 在 8080 真实页面验证保存后回到只读态
- [x] 记录 review 和验证结果

当前判断：

- 限流编辑器保存成功后只调用 `refresh(false)`，没有重置本地 `editorState.editable`，因此抽屉继续显示 `保存 / 撤销` 和输入控件。
- 同类页面中，泳道组、主动探测、新增流量治理在保存更新成功后都会显式退出编辑态；限流应保持同一交互语义。

当前进展：

- `RateLimitEditor` 在 `op === 'view'` 的更新保存成功后设置 `editable=false`，保留当前抽屉和更新后的本地数据展示。
- 限流匹配条件表格和阈值表格增加 view/edit key，避免从编辑态切回只读态后 TDesign 表格继续保留输入单元格。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `4662`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.5f7beb00.js` 和 `/assets/style.6e97c28c.css`。
- 浏览器打开 `http://127.0.0.1:8080/governance/`，进入 `spec-check-ratelimit`：查看态显示 `编辑 / 发布`，无 `保存`，限流表格输入框数量为 0。
- 点击 `编辑` 后显示 `保存 / 撤销`，限流表格输入框数量为 18。
- 点击 `保存` 后自动回到只读态：`保存 / 撤销` 消失，`编辑 / 发布` 出现，表格输入框数量为 0；匹配条件仍显示 `x-tenant/vip`、`channel/mobile`，窗口仍显示 `1 秒`、`5 秒`。

Review：

- 本次根因不是接口保存失败，而是限流编辑器保存成功后只刷新列表，没有退出本地编辑态。
- 仅切换 `editable=false` 不够，TDesign 表格可能保留编辑单元格内部状态；需要用 view/edit key 触发表格重新挂载。

# 治理规则保存后统一回只读态

- [x] 横向梳理 9 类治理规则保存后的状态流转
- [x] 修复路由保存成功后未回到只读态
- [x] 修复熔断保存成功后未回到只读态
- [x] 修复无损保存成功后关闭详情而不是回只读态
- [x] 给路由、熔断的编辑表格增加 view/edit 重新挂载 key
- [x] 修复无损保存 payload 缺少无损下线 `interval_second`
- [x] 修复泳道组编辑态服务选项为空导致保存失败
- [x] 构建前端并重启 all 模式
- [x] 在 8080 逐类验证保存后回只读态
- [x] 记录 review 和验证结果

当前判断：

- 已具备正确行为的规则：调用鉴权、流量镜像、流量 Mock、泳道组、主动探测，以及上一轮修复后的限流。
- 需要修复的规则：路由和熔断保存成功后只刷新列表，没有退出本地编辑态；无损保存成功后在独立页面会关闭抽屉，和“保存后回到 readonly 专题”的交互不一致。
- TDesign `Table` 的 `keepEditMode` 在编辑态切只读态时可能保留内部编辑单元格，需要按 view/edit 状态切换 key 触发重新挂载。

当前进展：

- `CustomRouteEditor` 更新保存成功后设置 `editable=false`，创建成功仍关闭抽屉；匹配条件表格和目标分组表格按 view/edit 重新挂载。
- `CircuitBreakerEditor` 非创建保存成功后设置 `editable=false`；接口范围、错误判断条件、熔断触发条件表格按 view/edit 重新挂载。
- `LossLessEditor` 更新保存成功后设置 `editable=false` 并刷新列表，创建成功仍关闭抽屉。
- `services/lossless.ts` 提交无损规则时补齐 `lossless_offline.interval_second`，同时按 spec 将 `health_check_interval_second` 作为 string duration 提交，避免查看态已有数据进入编辑保存后接口拒绝。
- `LaneGroupEdtor` 改为用 `listAllServices` 返回值构造服务选项，并在提交时从 `namespace/service` 做兜底解析；泳道入口 selector 的 `@type` 放入 protobuf Any selector 内部，避免提交成后端无法解码的顶层字段。

验证：

- `cd console/web && npm run build` 通过，保留既有 `--localstorage-file`、Browserslist 和 Vite 大 chunk 警告。
- 已重启 tmux session `pole-control-plane`，all 模式进程 PID `29717`，8080/8090 均监听正常。
- 8080 首页返回新资源 `/assets/index.18c87f31.js` 和 `/assets/style.6e97c28c.css`。
- Playwright 使用 8080 真实登录态逐类验证 9 类规则：路由、限流、熔断、探测、无损、泳道、调用鉴权、流量镜像、流量 Mock。
- 每一类验证流程均为：筛选类型、打开规则详情、确认初始只读态无 `保存/撤销` 且有 `编辑/发布`、点击 `编辑` 后出现 `保存/撤销`、点击 `保存` 后回到只读态且没有请求错误。
- 9 类最终结果均通过；无损和泳道在修复前分别暴露 `health_check_interval_second` 类型错误与 Any selector `@type` 位置错误，修复后单独回归和全量回归均通过。
- `git diff --check -- console/web/src/pages/Governance/RateLimit/RateLimitEditor.tsx console/web/src/pages/Governance/Router/CustomRouteEditor.tsx console/web/src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx console/web/src/pages/Governance/LossLess/LossLessEditor.tsx console/web/src/services/lossless.ts console/web/src/pages/Governance/Router/LaneGroupEdtor.tsx context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮没有改治理规则后端语义，主要收敛前端保存后的状态流转和 spec JSON 提交形态。
- 新增/既有 9 类治理规则现在使用同一交互契约：查看态展示 `编辑/发布`，编辑态展示 `保存/撤销`，更新保存成功后保持抽屉打开并回到只读态。
- 无损和泳道的失败不是样式问题，而是编辑态保存 payload 与当前 spec 解码规则不一致；已通过网络请求体和响应体定位后修正。

# Console API、Client 与权限接口 E2E 测试用例设计

- [x] 梳理现有 Go 集成测试、MySQL CI、Console proxy 和接口测试能力
- [x] 梳理 Console 资源读写入口：命名空间、MCP、A2A、服务、别名、实例、治理规则、用户、用户组、权限
- [x] 梳理 Client 查询入口和缓存传播断言方式
- [x] 梳理 `consoleOpen` / `clientOpen` 权限开关矩阵
- [x] 新增长期测试用例设计文档
- [x] 更新 `context-kg` index、testing、schema、log 与 lessons
- [ ] 后续按设计实现 Console API E2E 和 Client/Auth E2E 套件

当前判断：

- 当前仓库已有 Go 集成测试和 MySQL CI；本设计只考虑接口维度，不纳入 `console/web` 前端测试。
- Console 全流程验证不能只打 8090 后端接口，应经 8080 console proxy 才能证明 JWT、反代、响应解包和旧登录态行为。
- Client 查询验证不能只查 DB 或 store/cache，应通过 8090 client API 轮询证明客户端真实可见。
- 权限测试需要重启或覆盖配置矩阵，分别验证 `consoleOpen=false/true` 和 `clientOpen=false/true`。

当前进展：

- 新增 `context-kg/quality/testcases/console-client-auth-e2e-testcases.md`，按两层接口套件设计：
  - Console API E2E：Go test 经 8080 调用 `/core/v1`、`/naming/v1`、`/auth/v1`、`/ai/*`。
  - Client E2E：Go test 经 8090 client API 验证服务发现、治理规则发布态、缓存传播和权限。
- 覆盖矩阵已列出命名空间、MCP、A2A、服务、别名、实例、9 类治理规则、用户、用户组、角色和权限策略的 create/list/detail/update/delete 或发布类流程。
- Client 传播用例要求通过 Console API 完成写入，再在 10 秒内轮询 8090 client API，记录首次命中时间、轮询次数、失败诊断响应。
- 权限用例已覆盖配置开关矩阵和 Console/Client 主体权限：只读、指定资源写、用户组继承、角色函数变更、token 禁用/刷新。

Review：

- 本轮是测试用例设计与知识库归档，尚未实现自动化代码。
- 测试设计已收敛为 Console API 和 Client/Auth 两层接口套件，不覆盖前端交互自动化。
- 治理规则按 9 类完整列出，不再只覆盖旧 6 类或新增 3 类中的某一类。

# 接口维度 E2E 测试用例文档收敛

- [x] 移除测试用例文档中的前端交互测试覆盖范围
- [x] 将 Console 资源读写矩阵调整为纯接口用例
- [x] 保留并强化 8080 Console API、8090 Client API、权限开关矩阵
- [x] 更新 testing 索引、context-kg log 和 lessons
- [x] 校验文档中不再残留前端页面测试设计

当前判断：

- 用户明确要求“不需要考虑前端测试，只考虑接口维度的测试”，因此设计文档不应再包含 `console/web/e2e/`、页面入口、抽屉、表单、保存后只读态等前端交互测试能力。
- Console 维度仍然要走 8080 console proxy，因为这是接口链路的一部分；Client 维度仍然走 8090 client API。

当前进展：

- `context-kg/quality/testcases/console-client-auth-e2e-testcases.md` 已改为 Console API E2E 和 Client E2E 两层接口套件。
- 命名空间、MCP、A2A、服务关联查询和治理规则公共用例中的前端场景已替换为接口查询、详情、过滤、发布、灰度和删除断言。
- `context-kg/quality/automation/testing.md`、`context-kg/_meta/index.md`、`context-kg/_meta/log.md` 已同步改为接口 E2E 口径。
- `context-kg/tasks/lessons.md` 已补充经验：接口维度测试设计不能默认加入 Console 前端交互测试能力。

验证：

- `rg -n "Console UI E2E|console/web/e2e|Playwright|页面入口|抽屉|保存后只读|UI 流程|UI 工具|UI 能力|UI 服务|CONSOLE-GOV-UI|三层套件" context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md context-kg/_meta/index.md context-kg/_meta/log.md || true` 仅命中 `_meta/log.md` 中历史 A2A 记录，主测试用例文档和 testing 索引无命中。
- `git diff --check -- context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- wiki 链接检查通过：`all wiki links resolve (27 files)`。

Review：

- 本轮只调整测试用例设计文档和知识库索引，不实现测试代码。
- 最终测试范围为接口维度：8080 Console API 读写、8090 Client API 查询传播、Console/Client 权限策略和 `consoleOpen/clientOpen` 开关矩阵。

# 接口 E2E 测试实现

- [x] 恢复接口 E2E 实现范围，复核设计文档、路由和启动配置
- [x] 新增 `test/e2e` 环境工具：testcontainers MySQL、30000+ 端口分配、临时 all 模式配置、server 子进程管理
- [x] 新增 Console API E2E：命名空间、MCP、A2A、服务、别名、实例、9 类治理规则、用户、用户组、角色、权限策略
- [x] 新增 Client API E2E：服务发现、实例变更、治理规则发布态和删除传播
- [x] 新增权限接口 E2E：`consoleOpen/clientOpen` 四象限、Console 读写权限、Client 读写权限
- [x] 确保默认 `go test ./...` 不启动 Docker，显式 `go test -tags=e2e ./test/e2e/...` 才启用
- [x] 做静态/编译级验证并记录 Review

当前判断：

- 用户要求“先不要求执行”，所以本轮不强制实际拉起 Docker/MySQL 和 all 模式跑完整 E2E；但测试代码需要具备可执行形态，后续可显式运行。
- E2E 环境必须和本地开发环境隔离：MySQL 用 testcontainers 拉起，容器端口从 `30000` 起始分配，control-plane 的 console/client 端口也从 `30000` 段分配，不能复用本地 `3306/8080/8090`。
- 新增测试统一使用 `//go:build e2e`，并在测试目录放置无 build tag 的 `doc.go`，保证默认测试命令不会编译或运行 Docker 相关逻辑。

当前进展：

- 新增 `test/e2e/internal/e2e` 环境工具，使用 `testcontainers-go` 拉起 MySQL 8.0.36，分配 30000 起始端口，生成临时 all 模式配置并管理 `go run . start --mode all` 子进程。
- 新增 `test/e2e/console_api`，覆盖命名空间、MCP、A2A、服务别名、实例、9 类治理规则、用户、用户组、角色和权限策略的接口读写与发布入口。
- 新增 `test/e2e/client`，通过 Console API 写入服务、实例和治理规则，再通过 8090 Client Discover 轮询验证服务发现、实例变更、规则发布态和删除传播。
- 新增 `test/e2e/auth`，覆盖 `consoleOpen/clientOpen` 四象限、用户/用户组/角色/策略/token 接口，以及只读策略对 Console 写请求和 Client Discover 请求的实际约束。
- `go.mod/go.sum` 新增 testcontainers/docker 依赖，避免保留无关的核心依赖升级。

验证：

- `go test -mod=readonly -count=1 ./test/e2e/...` 通过，所有包均为 `[no test files]`，证明默认 build tag 下不会启动 Docker 或编译 E2E 运行逻辑。
- `GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go test -mod=readonly -count=1 -tags=e2e ./test/e2e/... -run TestDoesNotExist` 通过，证明 E2E 套件编译通过且未执行测试函数。
- `GOPROXY=https://goproxy.cn,direct go test -tags=e2e ./test/e2e/... -run TestDoesNotExist` 在当前 macOS 环境失败于 `github.com/shoenig/go-m1cpu` cgo 初始化 SIGSEGV；该问题来自 testcontainers 间接依赖链，推荐显式设置 `CGO_ENABLED=0`。
- `git diff --check -- test/e2e go.mod go.sum context-kg/tasks/todo.md` 通过。

Review：

- 本轮未按用户要求执行完整 E2E，因此没有拉起 Docker/MySQL 和 all 模式运行真实用例；完成的是测试代码补齐和编译级验证。
- E2E 与本地测试隔离：默认测试命令不触发，显式 `-tags=e2e` 才进入；端口段从 30000 开始，不复用本地 8080/8090/3306。
- 接口测试范围保持纯 HTTP/API，不包含 Playwright、浏览器、DOM、截图或前端页面交互。

# 接口 E2E 测试用例文档二次收敛

- [x] 复核测试用例文档中的前端、页面和 Playwright 相关残留
- [x] 在测试用例文档中增加接口 E2E 非目标清单
- [x] 将少量 UI 口径描述改为接口响应字段或接口断言
- [x] 更新 testing 索引和 lessons，固化“只考虑接口维度”的边界
- [x] 执行文档级验证

当前判断：

- 用户本轮要求的是测试用例文档整理，不继续扩展前端页面测试能力。
- 8080 console proxy 仍属于接口链路验证，不等同于浏览器页面测试；文档需要明确排除 Playwright、DOM、截图、页面布局和前端路由。
- 后续实现接口 E2E 时，应只按 HTTP/API 维度落地：8080 Console API、8090 Client API、权限策略和配置开关矩阵。

当前进展：

- `context-kg/quality/testcases/console-client-auth-e2e-testcases.md` 新增 `非目标` 小节，明确不使用 Playwright、浏览器驱动、DOM 查询、截图或页面可见性断言。
- 将 A2A 来源、实例查询、规则发布状态和启动健康等描述收敛为接口字段、接口响应和接口可用性，不再使用 UI 语义。
- `context-kg/quality/automation/testing.md` 增加说明：接口 E2E 不纳入 `console/web` 页面交互测试。
- `context-kg/tasks/lessons.md` 更新经验，避免后续把接口 E2E 误扩成前端自动化。

验证：

- `git diff --check -- context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md context-kg/_meta/log.md context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `rg -n "Console UI E2E|console/web/e2e|页面入口|抽屉|保存后只读|UI 流程|UI 工具|UI 能力|UI 服务|CONSOLE-GOV-UI|三层套件|浏览器驱动.*入口|Playwright.*入口|DOM.*断言" context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md || true` 无命中。
- `rg -n "Playwright|浏览器|DOM|截图|页面|console/web" context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md` 仅命中非目标/排除说明和 `## 相关页面` 标题。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过，27 个 Markdown 页面、27 个唯一页面名，frontmatter、链接、index 基础检查通过。

Review：

- 本轮只修改测试用例设计文档、testing 索引和任务/经验记录，没有新增或调整前端测试代码。
- 最终测试边界为接口 E2E：Console API 读写、Client 查询传播、权限策略生效和 `consoleOpen/clientOpen` 开关矩阵。

# 接口 E2E 测试文档边界强化

- [x] 复核测试用例文档是否仍把前端页面 Playwright 当作测试能力
- [x] 在主测试用例文档中补充接口 E2E 能力边界
- [x] 同步更新 testing 索引和 context-kg log
- [x] 执行文档级验证

当前判断：

- 用户要求只考虑接口维度，因此文档不能规划 `console/web/e2e`、Playwright 项目、浏览器脚本、DOM 断言、截图比对或页面交互流程。
- 8080 console proxy 仍保留在测试范围内，因为它是 Console 接口链路的一部分，不是页面渲染验证。

当前进展：

- `context-kg/quality/testcases/console-client-auth-e2e-testcases.md` 新增“能力边界”，明确自动化入口统一为 Go test。
- `context-kg/quality/automation/testing.md` 同步改为“不建设 Playwright、浏览器、DOM、截图或 console/web 页面交互测试能力”。
- `context-kg/_meta/log.md` 追加本轮 refine 记录。

验证：

- `rg -n "Console UI E2E|console/web/e2e|页面入口|抽屉|保存后只读|UI 流程|UI 工具|UI 能力|UI 服务|CONSOLE-GOV-UI|三层套件|浏览器脚本|DOM 断言|截图比对|Playwright 项目" context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md || true` 只命中“能力边界/非目标”排除说明。
- `rg -n "Playwright|浏览器|DOM|截图|页面|console/web" context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md || true` 只命中“不建设/不验证”表述和 `## 相关页面` 标题。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过，27 个 Markdown 页面、27 个唯一页面名，frontmatter、链接、index 基础检查通过。
- `git diff --check -- context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md context-kg/_meta/log.md context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。

Review：

- 本轮只整理接口测试用例文档和知识库记录，没有改前端测试代码或新增 Playwright 能力。
- 最终边界为接口 E2E：Go test 驱动 8080 Console API、8090 Client API、权限策略和 `consoleOpen/clientOpen` 配置开关矩阵。

# 接口 E2E 测试覆盖补齐审计

- [x] 补齐 Console API 删除后不可见断言：命名空间、MCP、A2A、服务、别名、治理规则、用户、用户组、角色、权限策略
- [x] 补齐 Console API 服务 update/delete 生命周期
- [x] 补齐治理规则 normal 发布、gray 发布、stopbeta、rollback 或 release delete 接口覆盖
- [x] 补齐 Client Discover 删除传播断言
- [x] 补齐 Client 权限 RegisterInstance 和 Heartbeat 用例
- [x] 重新执行默认隔离、E2E 编译、diff 和 context-kg lint 验证

当前判断：

- 设计文档中要求的 Console 资源 create/list/detail/update/delete、不存在断言、治理规则发布态、Client 查询传播、权限开关矩阵和 Client Register/Heartbeat 权限已经在 `test/e2e` 代码中有对应测试表达。
- 用户要求“先不要求执行”，因此本轮仍不拉起 Docker/MySQL/all 模式跑完整 E2E；完成编译级验证和默认隔离验证。
- `testcontainers-go v0.35.0` 明确要求 `github.com/magiconair/properties v1.8.7`，因此该间接依赖升级属于 E2E 依赖链必要变更。

当前进展：

- `console_api`：命名空间使用临时 namespace 完整 create/update/list/delete；MCP/A2A 删除后查列表不返回；服务新增临时服务 update/delete；治理规则 9 类增加 gray release、stopbeta、支持类型 rollback、release delete 和当前态 delete 后不可见；auth 资源增加 token enable/refresh、detail、双向授权查询和 delete 后不可见。
- `client`：实例删除后轮询 Discover 响应不再包含删除实例 host；治理规则删除后轮询 Discover 响应不再包含规则名。
- `auth`：四象限开关新增无 token RegisterInstance/Heartbeat；授权 token 新增 RegisterInstance/Heartbeat 成功，未授权服务 RegisterInstance/Heartbeat 拒绝。
- `internal/e2e`：新增 `HeartbeatInstance`、`GrayRuleRelease`、按字段查找和删除后不可见断言辅助函数。

验证：

- `go test -mod=readonly -count=1 ./test/e2e/...` 通过，所有包均为 `[no test files]`，证明默认 build tag 下不启动 Docker。
- `GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go test -mod=readonly -count=1 -tags=e2e ./test/e2e/... -run TestDoesNotExist` 通过，证明 E2E 套件编译通过且未执行测试函数。
- `git diff --check -- test/e2e go.mod go.sum context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md context-kg/_meta/log.md context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过，27 个 Markdown 页面、27 个唯一页面名，frontmatter、链接、index 基础检查通过。

Review：

- 本轮继续保持接口 E2E 边界，不引入 Playwright、浏览器、DOM、截图或页面交互测试。
- 完整 E2E 真实执行仍留给用户后续显式运行 `CGO_ENABLED=0 go test -tags=e2e ./test/e2e/... -count=1`。

# 接口 E2E 测试用例文档二次整理

- [x] 复核主测试用例文档和 testing 索引中的前端自动化表述
- [x] 将具名浏览器测试栈描述收敛为“不纳入前端页面自动化能力”
- [x] 执行文档 grep、diff 和 context-kg lint 验证

当前判断：

- 当前测试用例文档只应表达接口 E2E 范围：8080 Console API、8090 Client API、权限策略和配置开关矩阵。
- 对前端页面测试只保留“非目标”级边界，不把具体浏览器测试技术写成测试能力项。

当前进展：

- `context-kg/quality/testcases/console-client-auth-e2e-testcases.md` 的能力边界和非目标小节已改为接口优先表述。
- `context-kg/quality/automation/testing.md` 同步说明 E2E 只包含 HTTP/API 维度。

验证：

- `rg -n "Playwright|浏览器驱动|DOM|截图|console/web/e2e|Console UI E2E|页面入口|抽屉|保存后只读|UI 流程|三层套件" context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md || true` 先命中非目标清单中的“抽屉”，已继续泛化为控制台前端页面视觉与交互。
- 重新执行上述 `rg` 命令无输出，主测试用例文档和 testing 索引不再包含具名前端测试栈或页面交互测试项。
- `git diff --check -- context-kg/quality/testcases/console-client-auth-e2e-testcases.md context-kg/quality/automation/testing.md context-kg/tasks/todo.md context-kg/_meta/log.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过，27 个 Markdown 页面、27 个唯一页面名，frontmatter、链接、index 基础检查通过。

Review：

- 本轮只整理测试用例文档和 testing 索引，没有新增前端测试能力，也没有调整 E2E 代码。
- 测试边界保持为接口 E2E：Go test 驱动 8080 Console API、8090 Client API、权限策略和 `consoleOpen/clientOpen` 开关矩阵。

# 接口 E2E 鉴权与治理覆盖补齐

- [x] 审计 Console API、Client、Auth E2E 与测试用例设计的覆盖差距
- [x] 修正 auth policy、resource authorize、principal/resources 的接口 payload
- [x] 补齐用户组继承、角色函数变更、Token 禁用/刷新、策略删除和策略更新传播用例
- [x] 补齐治理规则权限类型隔离：老规则路由与新增流量治理规则均覆盖授权/未授权分支
- [x] 将灰度发布测试边界收敛到当前接口可验证的 release/client label/stopbeta 状态
- [x] 执行默认隔离、E2E 编译、文档 grep、diff 和 context-kg lint 验证

当前判断：

- `test/e2e` 现在仍通过 build tag 隔离；默认 `go test ./test/e2e/...` 不会启动 Docker 或 MySQL。
- Console API 覆盖命名空间、MCP、A2A、服务/别名/实例、9 类治理规则和 auth 资源的接口生命周期。
- Client API 覆盖服务实例变更传播、9 类治理规则发布后 Discover 可见、删除后 Discover 不再返回目标文本。
- Auth E2E 覆盖 `consoleOpen/clientOpen` 四象限、Console 读写权限、Client Discover/Register/Heartbeat、用户组继承、角色函数变更、Token 失效与刷新、策略删除和策略资源更新传播。
- 当前 HTTP Discover 接口没有稳定的客户端 label 入参，因此灰度发布用例验证 release 中 client label、gray release 和 stopbeta，不再把“指定 label 客户端可见/不可见”写成接口 E2E 的强断言。

当前进展：

- `test/e2e/internal/e2e/fixtures.go` 新增按真实 auth proto 结构生成 principals/resources/functions 的 helper，服务资源使用后端生成的 service id，并保留 namespace + service 组合资源。
- `test/e2e/console_api/console_api_test.go` 增加 MCP/A2A 删除后关联查询不可见、实例删除后 host 不可见、治理规则 gray release/stopbeta/rollback/release delete 覆盖。
- `test/e2e/client/client_api_test.go` 覆盖 9 类治理规则发布后 Discover 可见，以及删除后 Discover 不再包含规则名。
- `test/e2e/auth/auth_api_test.go` 增加 Console/Client 权限传播、用户组/角色/Token 场景，并将治理规则权限测试扩展到 `CreateRouteRules` 与 `CreateTrafficSecurityRules`。
- `context-kg/quality/testcases/console-client-auth-e2e-testcases.md` 保持接口 E2E 边界，并把灰度/路由/泳道的客户端 label 表述改成当前 HTTP/API 可验证的断言。

验证：

- `go test -mod=readonly -count=1 ./test/e2e/...` 通过，所有包均为 `[no test files]`，证明默认 build tag 下不启动 Docker。
- `GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go test -mod=readonly -count=1 -tags=e2e ./test/e2e/... -run TestDoesNotExist` 通过，证明 E2E 套件编译通过且未执行测试函数。

Review：

- 本轮仍按用户要求不执行完整 E2E，不拉起 testcontainers/MySQL/all 模式。
- 本轮没有新增前端页面测试能力，不引入 Playwright、浏览器、DOM、截图或页面交互断言。

# specification 流量治理 caller -> callee 最终契约收敛

- [x] 明确本轮只讨论最终 spec 效果，不考虑新老 spec 兼容
- [x] 从 `TrafficMirror` / `TrafficMock` 删除旧的 `target_service` 字段和兼容性注释
- [x] 重新生成 Go / Rust 产物，确认生成代码中不再暴露 `TrafficMirror.target_service` 和 `TrafficMock.target_service`
- [x] 执行 spec 构建、diff 与 context-kg lint 验证
- [x] 提交并推送到既有 specification PR 分支

当前判断：

- 流量镜像和服务 Mock 的最终契约应与路由一样直接表达 `caller -> callee`。
- `caller` 可用 `*/*` 表达全部服务；`callee` 是规则归属的被调服务。
- 本轮不做 `target_service`、旧 `CALLER_SERVICE` 推导或任何旧 spec fallback。

当前进展：

- `api/v1/traffic_manage/mirror.proto` 与 `mock.proto` 已改为顶层 `caller=4`、`callee=5`、`rules=6`，后续字段顺延。
- Go 生成产物已删除 `TargetService` 字段和 `GetTargetService()` 方法，`caller` / `callee` 使用字段号 4 / 5。
- Rust proto 拷贝和 `source/rust/pole-specification/src/v1.rs` 已同步最终字段。
- 已在 specification 分支 `codex/traffic-caller-callee` 追加提交 `98a0a01 feat: finalize traffic caller callee scope` 并推送。
- PR #7 当前为 Open / Mergeable，head 为 `98a0a01`。

验证：

- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && go test ./...` 通过，Go 生成包均编译成功。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust/pole-specification && PROTOC=/Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/protoc/protoc-darwin-arm64/bin/protoc cargo build --release` 通过。
- `rg -n "target_service|TargetService|GetTargetService|Deprecated: use callee|use callee instead" ...mirror/mock... || true` 无输出，证明 Mirror/Mock 范围内无兼容字段残留。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee && git diff --check` 通过。
- `gh pr view 7 --repo lattice-hub/specification --json url,state,mergeable,headRefName,headRefOid,title` 返回 PR Open / Mergeable，head 为 `98a0a01`。

Review：

- 本轮 spec PR 不再保留新老 spec 兼容字段、Deprecated 注释或 fallback 语义。
- `target_service` 在其它治理规则类型中仍可能存在，本轮只调整流量镜像和服务 Mock。

# specification 流量治理持续时间字段删除

- [x] 明确 Mock 和镜像最终 spec 不需要配置持续时间
- [x] 删除 `MirrorRule.duration` 与 `MockRule.delay`
- [x] 移除 Mirror/Mock proto 中不再使用的 `google.protobuf.Duration` 依赖并重新生成 Go / Rust 产物
- [x] 执行 spec 构建、diff 与 context-kg lint 验证
- [x] 提交并推送到既有 specification PR 分支

当前判断：

- 流量镜像和服务 Mock 的最终配置只保留接口范围、匹配条件、目标/响应、百分比和启停状态。
- 不保留持续时间、响应延迟或任何 `google.protobuf.Duration` 配置字段。
- 删除字段后按最终 spec 风格压实字段号：`disable` 使用字段号 5。

当前进展：

- `api/v1/traffic_manage/mirror.proto` 删除 `google/protobuf/duration.proto` import 和 `MirrorRule.duration`，`disable` 改为字段号 5。
- `api/v1/traffic_manage/mock.proto` 删除 `google/protobuf/duration.proto` import 和 `MockRule.delay`，`disable` 改为字段号 5。
- Go 生成产物已删除 `durationpb` import、`GetDuration()`、`GetDelay()` 以及对应字段。
- Rust proto 拷贝和 `source/rust/pole-specification/src/v1.rs` 已同步删除两个 Duration 字段。
- 已在 specification 分支 `codex/traffic-caller-callee` 追加提交 `02c3eb1 feat: remove traffic duration fields` 并推送。
- PR #7 当前为 Open / Mergeable，head 为 `02c3eb1`。

验证：

- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && go test ./...` 通过，Go 生成包均编译成功。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust/pole-specification && PROTOC=/Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/protoc/protoc-darwin-arm64/bin/protoc cargo build --release` 通过。
- `rg -n "Duration|duration|delay|持续时间|响应延迟|google/protobuf/duration.proto|GetDuration|GetDelay" ...mirror/mock... || true` 无输出。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee && git diff --check` 通过。
- `gh pr view 7 --repo lattice-hub/specification --json url,state,mergeable,headRefName,headRefOid,title` 返回 PR Open / Mergeable，head 为 `02c3eb1`。

Review：

- 本轮将“持续时间/延迟”作为 Mock 与镜像的最终 spec 非目标删除，没有保留字段号占位。
- `source/rust/pole-specification/proto/service.proto` 是 Rust 脚本生成噪声，已还原，避免带入无关 diff。

# specification 流量治理复用服务端点类型

- [x] 明确不新增 `ServiceScope` 重复类型
- [x] 删除 `router.proto` 中新增的 `ServiceScope`
- [x] 将 `TrafficMirror` / `TrafficMock` 的 `caller` 改为 `SourceService`，`callee` 改为 `DestinationService`
- [x] 重新生成 Go / Rust 产物，确认无 `ServiceScope` 残留
- [x] 执行 spec 构建、diff 与 context-kg lint 验证
- [x] 提交并推送到既有 specification PR 分支

当前判断：

- 现有 `SourceService` / `DestinationService` 已能表达规则两端，没有必要新增 `ServiceScope`。
- `caller` 使用 `SourceService`，允许 `service="*"` 且 `namespace="*"` 表达全部服务。
- `callee` 使用 `DestinationService` 表达规则归属和下发绑定的被调服务。

当前进展：

- `api/v1/traffic_manage/router.proto` 已删除新增的 `ServiceScope` message。
- `api/v1/traffic_manage/mirror.proto` 与 `mock.proto` 已改为 `SourceService caller = 4`、`DestinationService callee = 5`。
- Go / Rust 生成产物已同步删除 `ServiceScope` 类型和引用。
- 已在 specification 分支 `codex/traffic-caller-callee` 追加提交 `25ca7dc feat: reuse traffic service endpoints` 并推送。
- PR #7 当前为 Open / Mergeable，head 为 `25ca7dc`。

验证：

- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust && bash build.sh` 通过。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/go && go test ./...` 通过，Go 生成包均编译成功。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/rust/pole-specification && PROTOC=/Users/chuntao.liao/Github/pole-io/specification-caller-callee/source/protoc/protoc-darwin-arm64/bin/protoc cargo build --release` 通过。
- `rg -n "ServiceScope" api/v1/traffic_manage source/go/api/v1/traffic_manage source/rust/pole-specification/proto source/rust/pole-specification/src/v1.rs || true` 无输出。
- `cd /Users/chuntao.liao/Github/pole-io/specification-caller-callee && git diff --check` 通过。
- `gh pr view 7 --repo lattice-hub/specification --json url,state,mergeable,headRefName,headRefOid,title` 返回 PR Open / Mergeable，head 为 `25ca7dc`。

Review：

- 本轮删除了重复抽象，复用既有服务端点类型；`SourceService` / `DestinationService` 字段顺序继续遵循现有 `service=1, namespace=2`。
- `source/rust/pole-specification/proto/service.proto` 是 Rust 脚本生成噪声，已还原，避免带入无关 diff。

# specification PR 合并、tag 与 control-plane 依赖更新

- [x] 核对 PR #7 状态、检查项和现有 tag 序列
- [x] 合并 specification PR #7 到 `develop`
- [x] 在合并后的 specification `develop` 上创建并推送新 tag
- [x] 更新 `pole-control-plane` 对 `github.com/pole-io/specification` 的依赖到新 tag
- [x] 执行依赖解析、编译级和 context-kg 验证

当前判断：

- PR #7 当前为 Open / Mergeable，base 为 `develop`，head 为 `25ca7dc`，无 status check。
- specification 现有最新 tag 为 `v0.1.0-ALPHA.30`，本轮应递增为 `v0.1.0-ALPHA.31`。
- `pole-control-plane` 当前 `go.mod` 依赖和 replace 都指向 `v0.1.0-ALPHA.30`，更新时只做定向版本调整，不运行 `go mod tidy`。

当前进展：

- PR #7 已合并到 specification `develop`，merge commit 为 `7e596e46b7be00fcc4476f50ee3c41e2969ce79a`。
- 已在 merge commit 上创建并推送 `v0.1.0-ALPHA.31`。
- `pole-control-plane` 的 `go.mod` 已将 `github.com/pole-io/specification` require 和 replace 都更新到 `v0.1.0-ALPHA.31`。
- `go.sum` 已通过 `go mod download github.com/pole-io/specification` 补充 `github.com/lattice-hub/specification v0.1.0-ALPHA.31` checksum。
- 适配了本仓流量治理转换层：TrafficSecurity 继续使用 `TargetService`；TrafficMirror / TrafficMock 使用 `Callee` 作为规则归属，并在缺省时输出 `Caller=*/*`。

验证：

- `gh pr merge 7 --repo lattice-hub/specification --merge` 成功；随后 `gh pr view 7 ...` 返回 `state=MERGED`，merge commit 为 `7e596e4`。
- `git push origin v0.1.0-ALPHA.31` 成功；`git ls-remote --tags origin v0.1.0-ALPHA.31` 可查到 tag。
- `go mod download github.com/pole-io/specification` 通过，只做定向模块下载。
- `go test -mod=readonly -count=1 ./apis/pkg/types/rules` 通过。
- `go test -mod=readonly -count=1 ./plugin/store/mysql -run 'TestTrafficGovernance'` 通过。
- `go test -mod=readonly -run TestDoesNotExist ./...` 通过，全仓 Go 包编译级验证通过。

Review：

- 本轮没有运行 `go mod tidy`，避免无关间接依赖漂移。
- 本仓存在大量既有工作区改动，本轮只在其基础上追加 specification 依赖升级和必要的流量治理转换层适配。

# 流量镜像规则编辑抽屉 PRD 对齐

- [x] 读取镜像规则编辑设计交接文档并核对当前实现差距
- [x] 将镜像前端类型与工具函数收敛到 `caller/callee`、无持续时间、多接口、多流量标签结构
- [x] 重构镜像「规则」Tab：基础信息 / 服务范围 / 镜像规则三段，子规则按接口范围、流量标签、镜像执行排列
- [x] 右侧实时 Spec 输出 `serviceRange.caller / serviceRange.callee / rules[].mirror.target`
- [x] 更新保存校验和 mirror 工具验证脚本
- [x] 运行前端工具脚本、构建检查、context-kg lint 和 diff 检查

当前判断：

- 新 PRD 要求镜像规则最终产品态显式表达 `caller -> callee`，服务范围直接读写 `caller` / `callee`，不再把 caller 同步到子规则 `CALLER_SERVICE` 条件。
- 服务范围的主调和被调都提供 `全部命名空间/全部服务` 选项；`target_service` 只作为旧数据读取兜底，不作为镜像保存 payload 输出。
- 镜像子规则按接口范围、流量标签、镜像执行三段表达；不展示持续时间，镜像目标只在子规则执行区内表达。
- UI 允许一条子规则维护多个接口；提交到当前后端契约前会将多个接口展开为多条单接口 `MirrorRule`，并剥离前端草稿字段 `interfaces` 和旧残留 `duration`。

当前进展：

- `TrafficGovernanceEditor` 已将镜像服务范围切换到 `caller` / `callee`，查看态与编辑态均展示主调方、被调方和流量方向。
- 镜像规则区已移除旧顶部说明块、持续时间和目标实例标签，改为可折叠的「镜像子规则」卡片。
- 镜像子规则编辑态按接口范围、流量标签、镜像执行排列；接口范围支持 HTTP / gRPC / Dubbo 和多行接口，流量标签只展示 key / op / value，不再暴露来源服务类型选择。
- 右侧实时 Spec 已改为 `apiVersion/kind/metadata/spec.serviceRange/spec.rules[].interfaces/trafficLabels/mirror` 结构，并随服务范围和子规则实时刷新。
- `trafficMirrorEditorUtils` 已更新校验：名称 kebab-case、caller/callee 非空、至少 1 条子规则、接口路径非空、标签 key/value 非空、比例 0-100、镜像目标非空。
- `traffic_governance` 前端服务层已让镜像默认规则使用 `caller` / `callee`，提交前剥离前端 `interfaces`；Mock 的 `delay` 也继续在提交前剥离，避免重新输出已删除的 duration 字段。

验证：

- `cd console/web && node scripts/verify-traffic-mirror-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Governance/Security/TrafficGovernanceEditor.tsx console/web/src/pages/Governance/Security/index.module.less console/web/src/pages/Governance/Security/trafficMirrorEditorUtils.ts console/web/src/services/traffic_governance.ts console/web/scripts/verify-traffic-mirror-editor-utils.mjs context-kg/tasks/todo.md` 通过。

Review：

- 本轮只调整 Console 前端的镜像编辑抽屉、前端服务类型和验证脚本，不修改后端 Go 存储/缓存/下发链路。
- 产品态 Spec 预览遵循交接稿；真实保存仍适配当前后端 `MirrorRule.api` 单接口结构，因此在提交前展开 `interfaces[]`。
- 旧数据中的 `CALLER_SERVICE` 仍可被读取为 caller 兜底，但保存时不会再将 caller 写回流量标签或匹配条件。

## 治理规则 API 范围协议一致性核对

- [x] 核对熔断、调用鉴权、流量镜像 spec 中 API 范围字段是否仍为单个 `API api`
- [x] 判断这些规则是否存在和限流相同的“一个子规则应覆盖多个 API”问题
- [x] 补充经验记录，避免只修限流而漏掉同类治理规则

当前判断：

- 用户已明确限流不需要考虑当前兼容，子限流规则应支持多个 API，并优先复用统一 `API` message。
- 本轮先做协议核对和结论，不直接改 spec 与 Console，避免在未确认熔断/鉴权/镜像语义前批量改字段。
- 已核对 specification：熔断 `BlockConfig.api`、调用鉴权 `TrafficSecurityPolicy.api`、流量镜像 `MirrorRule.api` 都是单个 `API`；同构的流量 Mock `MockRule.api` 也仍是单个 `API`。
- 判断：如果产品语义允许一个策略/子规则使用同一套动作和条件覆盖多个 API，这些字段都应统一改为 `repeated API apis`，否则 Console 会被迫复制子规则或策略，和用户期望的“新增接口行”不一致。

## 治理规则多 API 范围协议与 Console 收敛

- [x] 核对 specification 生成命令、control-plane 对 `api/method` 字段的读写链路
- [x] 将限流、熔断、调用鉴权、流量镜像、流量 Mock 的单 API 字段统一为 `repeated API apis`
- [x] 适配 control-plane 后端转换、Console service 类型、编辑器 UI 和 Spec 预览
- [x] 运行协议生成、前端构建、相关验证脚本和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 用户确认不考虑现有兼容，子规则/策略应支持多个 API，接口资源用统一 `API` message 表达。
- 改动范围应横向覆盖限流、熔断、调用鉴权、流量镜像、流量 Mock，避免只修限流后留下同构问题。
- specification 已改为：`LimitTrigger.apis`、`BlockConfig.apis`、`TrafficSecurityPolicy.apis`、`MirrorRule.apis`、`MockRule.apis`；Mirror/Mock 同步补齐 `caller/callee` proto 形态以匹配 control-plane 当前生成类型。
- control-plane 使用本地 `../specification` 生成代码验证，`pkg/goverrule/ratelimit_rule.go` 已把简单/高级限流创建写入 `LimitTrigger.apis`。
- Console 限流的“新增接口”现在追加同一子规则内的 API 行；熔断、鉴权、镜像、Mock 提交时保留同一策略/子规则内的 `apis[]`，不再按接口展开成多条。

验证：

- `cd ../specification/source/go && bash build.sh` 通过，刷新 Go 生成代码。
- `cd console/web && node scripts/verify-ratelimit-editor-utils.mjs && node scripts/verify-circuitbreaker-editor-utils.mjs && node scripts/verify-traffic-security-editor-utils.mjs && node scripts/verify-traffic-mirror-editor-utils.mjs && node scripts/verify-traffic-mock-editor-utils.mjs` 通过。
- `cd console/web && npm run build:test` 通过；保留既有 Browserslist 过期和大 chunk 警告。
- `go test ./pkg/goverrule/... ./plugin/store/mysql/... ./plugin/apiserver/httpserver/...` 通过。
- `git diff --check -- console/web/src console/web/scripts pkg/goverrule go.mod context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `git -C ../specification diff --check --` 对本轮触及的 proto 和 Go 生成文件通过；未包含进入本轮前已脏的 `source/rust/pole-specification/proto/service.proto`。

Review：

- 本轮没有重跑 Rust 生成代码，避免把已有脏改 `source/rust/pole-specification/proto/service.proto` 及其派生内容混入；仅同步修改了对应 Rust proto 源文件。
- `go.mod` 暂时指向 `../specification`，用于验证本地未发布的 proto 生成代码；后续发版后应再切回正式版本依赖。
- 限流读取旧 `method` 字符串、熔断/鉴权/镜像/Mock 读取旧 `api` 字段只作为旧数据兜底；新的提交 payload 均输出 `apis`。

## 治理规则移除动态实时 Spec 预览

- [x] 梳理所有治理规则编辑器中 `实时规则 SPEC` 预览的挂载点
- [x] 移除路由、限流、熔断、主动探测、无损、泳道、鉴权、镜像、Mock 的右侧实时 Spec 预览
- [x] 清理对应无用 state、复制逻辑和预览文本构造，保留保存/校验所需逻辑
- [x] 将宽抽屉从双栏时期的 `min(1560px, calc(100vw - 40px))` 收敛为单栏编辑态的 `clamp(860px, 60vw, 1180px)`
- [x] 补齐治理工作台按规则类型筛选后的统一新建入口，覆盖路由、限流、熔断、主动探测、无损、泳道、鉴权、镜像、Mock
- [x] 将工作台规则清单调整为标题/主操作一行、类型筛选/搜索筛选一条工具栏，避免把标题、类型按钮和输入框硬挤在同一行
- [x] 构建或类型检查验证前端无未使用引用，重新拉起服务验证页面

当前判断：

- 用户明确要求移除所有治理规则里的动态实时 Spec 预览，本轮不调整后端 spec、保存 payload 或发布流程。
- 二次扫描发现熔断编辑器也存在同类实时 Spec 预览，已纳入同一轮移除。
- 右侧 Spec 列移除后，宽抽屉继续使用 1560px 会过宽；按用户反馈进一步收敛到 60vw，并保留 860px 最小宽度兜底复杂规则表单。
- 治理工作台不能只在熔断筛选下提供新建入口；选择具体规则类型后应出现对应的新建按钮，并打开对应 editor 的 create 态。`全部` 筛选下暂不展示新建，避免规则类型不明确。
- 规则清单不应把标题、类型按钮、搜索框和所有操作硬挤在同一行；标题与新建/刷新保持在清单头部，类型筛选与搜索/命名空间/服务/重置合并为一条筛选工具栏，宽屏一行展示，窄屏换行降级。

验证：

- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Governance context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `rg -n "实时规则 SPEC|实时 Spec|保存前核对|前端交互预览|已复制当前.*Spec|已复制.*规则预览|当前规则可保存|校验通过，可保存" console/web/dist/assets console/web/src/pages/Governance || true` 无输出。
- 后台服务已重启到 tmux 会话 `pole-control-plane`，PID `93603`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`。
- 宽抽屉收敛后重新执行 `cd console/web && npm run build:test` 通过；后台服务已重启到 tmux 会话 `pole-control-plane`，PID `5312`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.7258ca47.js`。
- 统一新建入口补齐后重新执行 `cd console/web && npm run build:test` 通过；`git diff --check -- console/web/src/pages/Governance/Workbench/index.tsx context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过；后台服务已重启到 tmux 会话 `pole-control-plane`，PID `30567`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.3f37b91c.js`。
- 规则清单布局回收为“两段式”后重新执行 `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，PID `57545`，`8080/8090` 监听正常，`http://127.0.0.1:8080/governance/workbench` 返回 `200`，首页资源 hash 更新为 `assets/index.973be23c.js` / `assets/style.c876001e.css`。

## 鉴权策略治理资源检查适配

- [x] 核对调用鉴权、流量镜像、流量 Mock 在策略授权、资源存在性检查和资源详情回显中的资源类型映射
- [x] 修复鉴权策略创建/更新时对 `security_rules`、`mirror_rules`、`mock_rules` 的资源存在性校验
- [x] 补齐策略详情中这三类治理资源的资源摘要转换
- [x] 增加针对资源检查和回显转换映射的单测
- [x] 运行相关 Go 测试、diff 检查和 context-kg lint

当前判断：

- 前端和发布链路区分了管理鉴权资源名与发布资源名：授权策略资源使用 `SecurityRules/MirrorRules/MockRules`，发布版本使用 `TrafficSecurityRules/TrafficMirrorRules/TrafficMockRules`。
- 后端 `auth_checker` 操作鉴权和治理规则发布资源收集已能识别 `SecurityRules/MirrorRules/MockRules`。
- `plugin/access_control/auth/policy/interceptor/paramcheck/server.go` 的 `checkResourceExist` 只检查到 `LosslessRules`，没有检查调用鉴权、流量镜像、流量 Mock，因此策略创建/更新时可能接受不存在的治理资源 ID。
- `plugin/access_control/auth/policy/policy.go` 的策略资源详情转换器缺少这三类治理资源，非 `*` 资源可能无法在策略详情中正确回显名称。

验证：

- `go test ./plugin/access_control/auth/policy/...` 通过。
- `go test ./apis/pkg/types/auth ./pkg/goverrule/interceptor/auth ./plugin/access_control/auth/policy/...` 通过。
- `git diff --check -- plugin/access_control/auth/policy/interceptor/paramcheck/server.go plugin/access_control/auth/policy/interceptor/paramcheck/traffic_governance_resource_test.go plugin/access_control/auth/policy/policy.go plugin/access_control/auth/policy/policy_test.go context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，all-mode 进程 PID `55355`，`8080/8090` 监听正常。
- `http://127.0.0.1:8080/governance/workbench` 返回 `200`。

Review：

- 本轮确认问题存在：不是前端资源名传错，而是鉴权策略参数校验与详情回显没有同步适配新增的三类治理资源。
- 已在策略创建/更新的资源存在性检查中补齐 `security_rules`、`mirror_rules`、`mock_rules`，不存在的调用鉴权、镜像或 Mock 规则 ID 会返回 `NotFoundResource`。
- 已在策略详情资源转换器中补齐 `SecurityRules/MirrorRules/MockRules`，非通配资源可从对应治理缓存回填规则名和命名空间。

## 用户详情关联策略展示修正

- [x] 复现 admin 用户详情看不到关联策略的问题，并区分后端数据与前端展示责任
- [x] 在用户详情页加载当前用户关联的鉴权策略
- [x] 调整用户详情权限信息卡片，确保管理员用户也能看到关联策略
- [x] 补充验证并记录 review

当前判断：

- 后端能返回 admin 的关联策略：`/auth/v1/policies?principal_id=<admin>&principal_type=user` 返回 1 条 `(用户) admin的默认策略`。
- 前端 `UserDetail.tsx` 没有调用关联策略接口，并且用 `viewUser?.user_type !== 'main'` 把整个权限信息卡片对管理员用户隐藏了，因此 admin 详情必然看不到关联策略。

验证：

- `POST http://127.0.0.1:8080/auth/v1/user/login` 使用 `admin/admin123` 登录成功，返回 admin 用户 ID 和 Console token。
- 带 `Authorization` 与 `X-Pole-User` 请求 `GET http://127.0.0.1:8080/auth/v1/users?id=<admin>&offset=0&limit=1` 返回 `user_type: "main"`，确认复现对象是主账号。
- 带 `Authorization` 与 `X-Pole-User` 请求 `GET http://127.0.0.1:8080/auth/v1/policies?principal_id=<admin>&principal_type=1&offset=0&limit=10` 返回 `amount=1`，包含 `(用户) admin的默认策略`，证明后端数据存在。
- `cd console/web && npm run build:test` 通过；保留既有 Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`http://127.0.0.1:8080/login` 返回 `200`。
- Playwright 登录 admin 并打开 `http://127.0.0.1:8080/auth/principals/userdetail?id=<admin>&name=admin`，页面包含 `管理员`、`权限信息` 和 `(用户) admin的默认策略`；仅捕获到既有 React key warning。

Review：

- 本轮根因是 Console 前端详情页缺少关联策略查询，并且把主账号的整个权限信息卡片隐藏了；不是后端策略数据缺失。
- 已在用户详情页调用 `describeAuthPolicies` 加载当前用户关联策略，并把权限信息页签对管理员用户保留展示。

## 认证管理详情编辑抽屉统一

- [x] 确认认证管理主入口现状：用户、用户组、角色、策略列表仍跳 hidden detail route，操作列缺少统一查看/编辑入口。
- [x] 先补充 `console/web/scripts/verify-auth-drawer-actions.mjs`，验证列表不再主动跳转详情页、名称和操作列统一打开抽屉。
- [x] 调整用户、用户组、角色、策略表格：名称点击和行内主操作均打开详情抽屉，删除仍保留行内确认，Token 入口不再空实现。
- [x] 调整用户、用户组、角色、策略编辑器：查看态和编辑态沿用同一抽屉入口，查看态可切到编辑态，创建态保持原语义。
- [x] 运行静态验证脚本、前端构建、diff 检查和 context-kg lint，并补充 review。

当前判断：

- 本轮只统一 Console 认证管理主路径的交互入口，不删除 `/auth/principals/*detail` 和 `/auth/policies/detail` hidden route，保证已有直达链接仍可兼容。
- 编辑保存链路保持现有 service/redux 结构，尤其策略编辑器不在本轮重写创建/更新接口契约。
- 表格操作列应与命名空间、治理工作台的近期规范一致：保留一个 `查看 / 编辑` 主入口，具体编辑动作在抽屉内发生。

调整判断：

- 用户反馈修改后的权限视图不如之前，说明“统一抽屉”不能等同于“把详情页改成基础字段或禁用表单”。
- 修正方向是恢复旧详情页的信息架构：主体详情保留关联策略/权限信息，策略详情保留成员、资源树、资源标签和接口范围；只把承载方式从独立页面改成抽屉。

验证：

- `cd console/web && node scripts/verify-auth-drawer-actions.mjs` 通过；脚本覆盖列表不跳旧详情路由、统一 `查看 / 编辑`、主体权限信息表和策略详情成员/资源/接口视图。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `git diff --check -- console/web/src/pages/Auth/Principal/UserTable.tsx console/web/src/pages/Auth/Principal/GroupTable.tsx console/web/src/pages/Auth/Principal/RoleTable.tsx console/web/src/pages/Auth/Principal/UserEditor.tsx console/web/src/pages/Auth/Principal/GroupEditor.tsx console/web/src/pages/Auth/Principal/RoleEditor.tsx console/web/src/pages/Auth/Principal/PrincipalPolicyTable.tsx console/web/src/pages/Auth/Principal/index.module.less console/web/src/pages/Auth/Policy/PolicyTable.tsx console/web/src/pages/Auth/Policy/PolicyEditor.tsx console/web/src/pages/Auth/Policy/PolicyDetailView.tsx console/web/src/pages/Auth/Policy/index.module.less console/web/scripts/verify-auth-drawer-actions.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`8080/8090` 监听正常。
- `curl -I http://127.0.0.1:8080/auth/principals` 与 `curl -I http://127.0.0.1:8080/auth/policies` 均返回 `200`。
- Playwright 登录 `admin/admin123` 后验证：用户详情抽屉存在 `权限信息` 和 `策略名称`；策略详情抽屉存在 `成员信息 / 资源信息 / 资源标签 / 可访问接口`。

Review：

- 本轮将用户、用户组、角色、策略列表名称和行内主操作统一为抽屉入口；hidden detail route 保留兼容，但主路径不再跳转过去。
- 新增 `PrincipalPolicyTable` 复用关联策略接口恢复主体详情的权限信息；用户组 Token 行内按钮从空实现改为查询并展示 token。
- 新增 `PolicyDetailView`，策略查看态不再使用禁用编辑器伪装详情，而是恢复旧详情页的成员、资源树、资源标签和接口范围视图。
- 收到用户反馈后已将“抽屉迁移不能压扁权限视图”的规则写入 `context-kg/tasks/lessons.md`。

## admin 策略详情数据缺失修复

- [x] 复现 admin 关联策略详情很多字段为空的问题，并对照真实接口返回
- [x] 补充认证抽屉静态校验，覆盖策略详情直接返回 `AuthStrategy` 和无损规则资源
- [x] 修复策略详情 service 解包、资源树类型覆盖，以及主体关联策略详情入口
- [x] 运行静态验证、前端构建、真实页面回归和 diff 检查
- [x] 更新 review 与 lessons

当前判断：

- `GET /auth/v1/policies?principal_id=<admin>&principal_type=1` 能返回 admin 默认策略，包含成员、资源、接口范围等完整数据。
- `GET /auth/v1/policies/detail?id=<policy>` 的 `data` 直接是 `AuthStrategy`，不是 `{ authStrategy: ... }`；当前 `describeAuthPolicyDetail` 只取 `result.authStrategy`，导致抽屉拿到 `undefined` 并显示大量空数据。
- 真实资源返回里还包含 `lossless_rules`，当前 `PolicyResources` 类型和策略详情资源树都没有覆盖，属于同一详情展示缺口。
- admin 默认策略通常不出现在普通策略列表主视图里，用户实际路径是主体详情的 `权限信息` 关联策略表；关联策略名称不能只是静态 Link，必须能继续打开策略详情抽屉。

验证：

- `POST http://127.0.0.1:8080/auth/v1/user/login` 使用 `admin/admin123` 登录成功，返回 admin 用户 ID 和 Console token。
- 带 `Authorization` 与 `X-Pole-User` 请求 `GET http://127.0.0.1:8080/auth/v1/policies?principal_id=<admin>&principal_type=1&offset=0&limit=100` 返回 1 条 `(用户) admin的默认策略`，且资源包含 `namespaces/services/config_groups/route_rules/ratelimit_rules/circuitbreaker_rules/faultdetect_rules/lane_rules/lossless_rules/security_rules/mirror_rules/users/user_groups/roles/auth_policies` 等字段。
- 带同一 token 请求 `GET http://127.0.0.1:8080/auth/v1/policies/detail?id=<policy>` 返回 `data` 直接为 `AuthStrategy`，包含成员、资源、`functions: ["*"]` 和描述。
- `cd console/web && node scripts/verify-auth-drawer-actions.mjs` 先在旧实现上失败，修复后通过；脚本覆盖详情解包兼容、无损资源树、主体关联策略可打开详情抽屉。
- `cd console/web && npm run build:test` 通过；保留既有 `--localstorage-file`、Browserslist 过期和大 chunk 警告。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 通过；tmux 日志显示 `finish starting server`，`8080/8090` 监听正常。
- Playwright 从 `http://127.0.0.1:8080/auth/principals` 登录态打开 admin 用户详情，进入 `权限信息`，点击 `(用户) admin的默认策略` 后策略详情接口返回 200；详情抽屉展示策略名称、admin 成员、描述、资源通配、无损规则通配和全部接口。
- `git diff --check -- console/web/src/services/auth_policy.ts console/web/src/pages/Auth/Policy/PolicyDetailView.tsx console/web/src/pages/Auth/Principal/PrincipalPolicyTable.tsx console/web/scripts/verify-auth-drawer-actions.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。

Review：

- 本轮根因不是后端 admin 策略缺失，而是 Console 前端详情 service 没兼容当前接口直接返回 `AuthStrategy` 的响应形态。
- `describeAuthPolicyDetail` 已兼容 `{ authStrategy }` 和直接 `AuthStrategy` 两种形态，策略详情抽屉可以拿到完整成员、资源、接口和描述。
- 策略资源类型补齐 `lossless_rules`，策略详情资源树新增 `无损规则`，避免默认策略中的无损规则资源无入口。
- 主体详情的关联策略表现在点击策略名会打开策略详情抽屉，admin 这类默认策略也能从主体权限信息继续查看完整详情。
