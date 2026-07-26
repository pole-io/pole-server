---
title: 任务计划与 Review
tags: [tasks, todo]
links: [lessons, adr-otel-observability-platform, adr-pole-rust-client-observability]
updated: 2026-07-27
sources: 0
---

# 任务计划与 Review

## Envoy xDS v3 适配最新内部治理规则（2026-07-26）

目标：让 Envoy xDS v3 复用统一治理规则的 active release 选择结果，按节点身份和治理标签隔离策略快照，并按显式 caller → callee 契约生成路由，避免 namespace 共享策略、旧 destination 反推和目标组信息丢失。

- [x] 审计现有 ADS/LDS/CDS/EDS/RDS/VHDS 请求链、缓存粒度和治理规则转换边界。
- [x] 对照统一存储、多灰度 release、owner namespace、caller/callee 与 RPC-first ADR 明确验收规格。
- [x] 为 Envoy Node 定义治理标签投影，并构造与 Discover 一致的 `ContextDiscoverFilter`。
- [x] 将 RDS/VHDS/CDS 等治理相关资源改为 node-scoped 快照，EDS 保持 namespace 共享。
- [x] 修复路由 wrapper 的 caller/callee 映射、多规则保留、Gateway 生成链和 DestinationGroup 权重/标签转换。
- [x] 对不支持的 OR、动态参数及非 Envoy 执行能力建立显式跳过/错误边界，禁止静默变义。
- [x] 补充 node label 灰度、caller/callee 隔离、sidecar/gateway 与缓存隔离回归测试。
- [x] 完成格式化、定向测试、全仓编译/测试、diff 检查和双轴代码审查。

### Review

- Envoy Node 仅将 `pole.io/governance-label.*` metadata 投影为治理标签，复用 Discover 的 active release 选择；同 namespace、同标签上下文复用一次规则选择，不同上下文生成相互隔离的 LDS/RDS/VHDS/CDS 稳定快照，EDS 继续按 namespace 共享。
- 自定义路由按顶层 caller → callee 契约过滤，normal/gray 都保留 release 内全部路由规则；sidecar 与 gateway 均保留 DestinationGroup 的多目标、权重和 `envoy.lb` 子集标签，网关路由按服务域名稳定排序。
- 基础 QPS 限流只接受可等价表达的固定 HTTP 条件与 token bucket；OR、动态参数、通配匹配、并发/系统资源、排队、自定义响应、爬坡、均摊及自定义 failover/action 等能力整条跳过并记录日志，避免删除条件后扩大命中。
- 节点快照先完成全部资源构建与确定性版本计算再原子替换；构建/序列化失败保留 last-known-good，最后一条 stream 关闭后同时清理节点索引与缓存。
- 新增 wrapper caller/callee、normal/gray 接线、治理标签投影、sidecar/gateway 路由、DestinationGroup 权重/标签、稳定排序、节点缓存隔离/原子性/生命周期及能力边界回归。`go test ./pkg/cache/rules ./pkg/goverrule ./plugin/apiserver/grpcserver/discover/v1 ./plugin/apiserver/xdsserverv3/... ./apis/pkg/types/rules -count=1`、`go test -p 1 ./... -count=1`、`go test -race ./plugin/apiserver/xdsserverv3/... -count=1` 与 `git diff --check` 均通过。
- 双轴代码复审最终无阻断或重要问题。并行全仓测试曾出现一次 Go 工具链读取标准库临时文件失败；目标包单测随即通过，串行全仓复验通过，确认不是代码失败。

## 控制台复合查询交互统一

目标：将控制台查询区统一为“主搜索框 + 已选条件标签 + 高级筛选浮层/折叠面板”，默认保持轻量；时间范围作为独立控件，不纳入智能条件标签。

- [x] 盘点全部查询页面、字段类型、时间范围和现有查询行为
- [x] 定义共享复合查询组件契约与页面迁移验收矩阵
- [x] 实现主搜索、自动补全、条件标签和高级筛选面板
- [x] 迁移资源列表、治理、认证、AI 与观测页面
- [x] 保持时间范围独立，并保留查询、重置、分页回到第一页等原行为
- [x] 补充全局静态契约、交互回归、暗色主题和响应式验证
- [x] 完成 ESLint、专项测试、构建、真实浏览器验收与双轴审查

Review：

- 新增应用级 `QueryComposer`：主搜索使用 Fluent Combobox 自动补全，文本/单选/多选/三态布尔进入高级 Popover，确认后生成可删除 Tag；远程页面使用显式查询，本地页面在主搜索输入或高级条件确认后即时过滤。
- 已迁移命名空间、服务、配置分组、治理工作台、MCP、A2A、系统配置、系统/服务监控、事件指标和操作审计。认证、治理子表与详情侧栏只有单字段局部搜索，继续保持一个轻量搜索框，不生成空高级筛选。
- 事件指标和操作审计的 DateRangePicker 通过独立 `timeRange` 插槽保留，不进入高级字段或 Tag；显式页面使用统一查询快照隔离草稿与已应用条件，查询和重置都回到第一页。
- 新增 `test:query-composer`，同步更新观测回归契约。`test:query-composer`、`test:metrics-observability`、`test:date-range-picker`、`test:dark-theme`、目标 ESLint、`build:test` 与 `git diff --check` 均通过。
- release 产物在 1440px 完成自动补全、中文输入法组合态 Enter、嵌套 Select、取消/Esc 回滚、完成提交、Tag 删除、显式查询草稿隔离和时间独立验收；720px 下页面无横向溢出，浮层宽 522px 且完整位于 900px 视口内，浏览器控制台 0 错误。最终双轴复审无遗留阻断；Orca 桌面辅助权限未授予，因此未执行桌面级读屏验收。

## 观测筛选日期范围选择器

- [x] 核对事件指标、操作审计的时间筛选状态与共享 DateRangePicker 契约
- [x] 将 TypeScript 升级到当前 ESLint 工具链支持的 5.1.6，并验证正常依赖解析
- [x] 将分离的起止日期时间输入与快捷下拉整合为弹层式日期范围选择器
- [x] 保留秒级时间、手动输入、清空和快捷时间范围能力
- [x] 补充共享组件与观测页面的专项回归检查
- [x] 完成 ESLint、专项测试、前端构建与变更审查

Review：

- 根因是 Fluent 适配层中的 `DateRangePicker` 实际只渲染两个原生日期时间输入和一个独立快捷下拉；现已替换为单一范围触发器、Fluent Popover、连续日期范围、开始/结束时间切换与可滚动时/分/秒三列，事件指标和操作审计无需修改页面数据流即可复用。
- 手动输入、清空、确认和快捷范围均保留；预设中的 `Date` 会先按 `YYYY-MM-DD HH:mm:ss` 归一化再回传，避免 API 收到 JavaScript Date 文本。反向点选会按新端点重新计算自然日边界，浏览器验证 `7/25 → 7/24` 得到 `7/24 00:00:00—7/25 23:59:59`。
- 新增 Fluent 官方 `@fluentui/react-calendar-compat@0.4.4`，并将 TypeScript 固定升级到当前 `@typescript-eslint@5.62` 支持的 `5.1.6`；不带兼容参数的 `npm install` 正常通过，未引入 TDesign。
- 已通过 `test:date-range-picker`、`test:no-tdesign`、`test:metrics-observability`、`test:dark-theme`、目标文件 ESLint、`build:test` 与 `git diff --check`。真实浏览器验证亮暗主题、弹层定位、日期范围端点、秒级手动输入、时分秒联动、快捷范围和反向点选；双轴代码复审最终无遗留问题。

## Agent 对话输入器视觉优化

- [x] 收敛输入器边框、阴影、间距与按钮层级
- [x] 消除 Fluent 全局 Textarea 聚焦样式造成的双蓝线
- [x] 补充输入器视觉契约并在 Kubernetes 容器内完成测试与构建
- [x] 发布新镜像并通过 Kubernetes Gateway 验证亮暗主题和交互状态

Review：

- 输入器改为 17px 组合表面、1px 弱边框和单层品牌色焦点光环；资源范围收敛为次级标签，发送按钮补齐 enabled、disabled、hover 和 pressed 层级。
- 根因不仅是内层 `textarea` 的全局 focus shadow，Fluent `Textarea` 根节点还通过 `::after` 绘制 4px 底线；现已在 `composerInput` 局部同时关闭 outline、shadow 和根节点伪元素。
- 第一版焦点声明引用了未定义的 `--app-primary`，K8s 浏览器发现焦点光环未生效后停止收尾；最终改用全局已定义的 `--app-brand`，亮暗主题均得到有效计算样式。
- `test:agent-workbench`、`test:login-redirect`、目标文件 ESLint 和 release build 均在 K8s Node Pod 内通过。
- 已滚动更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-agent-composer-v2`；Pod `pole-control-plane-85d65f4796-skqkp` Ready、0 restart，imageID 为 `sha256:0f72fb4e9d18...`。
- K8s Playwright 在 1440×900 的亮暗主题中确认：内层 border/outline/shadow 均为 0/none，根节点 `::after` 不显示，外层宽 819px、高 134px、圆角 17px、边框 1px，并存在 3px 品牌色焦点光环；输入后发送按钮可用。
- Pod 与 Gateway 都加载 `assets/index.14eb389b.js`，实际截图为 `output/playwright/agent-composer-light.png` 和 `output/playwright/agent-composer-dark.png`。

## 登录页未登录提示降噪

- [x] 建立受保护路由跳转登录页时出现重复 Toast 的回归信号
- [x] 移除登录页重复“您当前未登录，请先登录”提示，保留原路径跳转状态
- [x] 在 Kubernetes 容器内完成前端专项测试与构建
- [x] 发布新镜像并通过 Kubernetes Gateway 验证登录页与登录失败提示

Review：

- 根因是 `PrivateRoute` 将来源位置写入 `state.from` 后，登录页主动把该状态解释成警告；本次只移除重复 Toast，受保护路由仍保留来源位置。
- 新增 `test:login-redirect` 契约测试；在 K8s Node Pod 内确认修复前因精确文案稳定失败，修复后通过，目标文件 ESLint 与 release build 同样在 K8s Pod 内通过。
- 已滚动更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-login-redirect-v1`；Pod `pole-control-plane-8469fc987b-kw9sp` Ready、0 restart，imageID 为 `sha256:34e24030ef62...`。
- K8s Playwright 使用全新无会话浏览器访问 `/namespace`，确认静默跳转 `/login`、`history.state.usr.from.pathname=/namespace`、页面没有重复 Toast；不存在的探针账号登录返回 400 时仍显示“请求错误”反馈。
- Pod 内 `index.html` 与 Gateway 都加载 `assets/index.a389d17f.js`，活动静态资源图中不再包含“您当前未登录，请先登录”。

## Namespace 环境模型全链路实现

- [x] specification 为全部治理规则聚合根增加归属 namespace，并完成生成代码与兼容测试
- [x] control-plane 统一治理规则存储、查询、锁定、发布、鉴权与删除保护的 namespace 语义
- [x] Console 治理工作台和规则编辑器支持归属环境筛选、展示与创建
- [x] Service、Config Group、Config File 详情支持同一逻辑资源的跨环境切换与摘要对照
- [x] 补齐跨环境查询的逐资源权限过滤，禁止泄露无权环境的资源存在性
- [x] 发布新版 specification，并更新 control-plane 与 Rust SDK 正式依赖
- [x] 完成单测、全量测试、前端构建、知识库校验、代码审查和本地 Kubernetes 发布

Review：

- specification 已发布 `v0.1.0-ALPHA.37`：九类治理聚合根增加顶层 `namespace`，并增加 `NamespaceExistedGovernanceRules=400220`；Go/Rust wire、JSON 与旧 payload 读取测试通过。
- control-plane 按 `namespace + rule_type + name` 查询、加锁和防重；发布、回滚、审计、授权资源上下文和 Namespace 删除保护均使用 owner namespace，流量治理目标服务 namespace 保持独立。
- Console 工作台默认进入 `default` 环境，按 owner namespace 查询和过滤；新建规则先选择归属环境，点击规则类型直接进入独立创建页，URL 使用独立 `ruleNamespace`。
- Service、Config Group、Config File 详情提供同一逻辑资源的可访问环境摘要与切换；Config File 使用 `group + file` 身份，并只读取 brief 摘要。配置查询在服务端逐资源授权后重算数量。
- Rust SDK 与 control-plane 正式依赖均更新到 `v0.1.0-ALPHA.37`；Rust SDK 157 个库测试、control-plane 全量 `go test ./... -count=1`、Console 专项契约和 release build 全部通过。
- 已仅更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-namespace-environment-v3`；Pod `pole-control-plane-78549b6f79-k7vmr` Ready、0 restart，imageID 为 `sha256:8a944100fde3...`。本地与 Gateway 的 `index.html` 和 release 构建 SHA256 一致，主资源为 `assets/index.4996b430.js`。
- 真实浏览器验证规则工作台显示“归属环境”，规则卡片点击直接进入 `?kind=route&ruleNamespace=default`，创建页明确显示“归属环境 default”；服务详情显示 Fluent 跨环境视图与可访问环境数量。
- 升级注意：历史 `governance_rule.namespace=''` 无法可靠推断真实环境，尤其旧流量治理曾混用目标环境；必须由部署方制定显式回填策略，不能自动猜测。

## Namespace 环境语义与 Website 描述统一

- [x] 复核 Namespace、Service、Config Group、Config File 与 Governance Rule 的资源身份语义
- [x] 将 Console 主要资源页说明统一为“namespace 是运行环境”
- [x] 更新领域术语、业务规则和功能档案，移除多租户及同名候选资源表述
- [x] 完成前端静态检查、构建、知识库一致性验证和本地 Kubernetes 发布

Review：

- Namespace、服务、配置分组、配置文件详情和治理工作台的说明已统一采用环境语义；治理工作台文案只描述目标领域边界，没有宣称尚未落地的跨环境切换或对比能力。
- 领域知识明确区分全局逻辑标识与 `namespace + resource key` 环境实例坐标；治理规则直接归属 Namespace，caller、callee、target service 保持运行时作用范围。
- 已通过 namespace workspace/drawer 专项检查、ESLint、release build、`git diff --check` 与 context-kg lint。
- 已构建 `linux/arm64` 镜像 `pole-control-plane:local-20260723-namespace-environment-copy-v2`，仅更新 `pole-system/deployment/pole-control-plane`；新 Pod Ready、无重启，8080、Gateway 深链及 8090 鉴权入口正常，新文案所在静态 chunk 与构建产物一致。

## 治理匹配条件 Fluent UI 交互与布局优化

- [x] 审计共享匹配条件编辑器的 Fluent UI 组件、信息层级、列宽和响应式问题
- [x] 将条件关系、字段行、删除和新增交互调整为 Fluent UI v9 的标准模式
- [x] 修复下拉触发区被相邻输入框遮挡、操作列表头换行和宽屏留白失衡
- [x] 覆盖固定值、请求参数、只读、无参数键及多行条件场景
- [x] 完成专项回归、ESLint、构建、all-mode 运行入口和真实页面验收

Review：

- 条件关系改用 Fluent `RadioGroup`，删除与新增改用原生 `Button`、`Tooltip` 和图标；每条条件成为独立语义行，补齐 table/row/cell 与字段可访问名称。
- “值来源”去掉会清空选中展示的可筛选模式，只保留固定值和请求参数；请求参数态明确展示“采集参数 / 采集该键的请求值”，不再伪装成需要填写匹配值。
- 响应式从页面视口判断改为组件容器查询：宽容器使用六列表格，嵌套子规则约 700px 时自动切换两列字段卡，最右操作按钮不再被裁切，窄容器进一步降为单列。
- 所有颜色、表面、边框和文字均使用应用主题 token；Playwright 在亮色和暗色模式下验证可读，截图为 `output/playwright/governance-fluent-match-condition-detail.png` 与 `output/playwright/governance-fluent-match-condition-dark.png`。
- 已通过 `test:traffic-match-condition-editor`、`test:governance-request-value-types`、`test:traffic-security`、`test:dark-theme`、ESLint 和 `build:test`；本地 8080/8090 运行入口持续可用，实时 `index.html` 与最新构建一致。

## 移除治理运行变量值来源

- [x] 核对 `VARIABLE` 在 specification、Console、Go 后端、Rust SDK 与知识库的完整影响面
- [x] 从 specification 删除 `VARIABLE` 并保留枚举号/名称不可复用约束，重新生成 Go/Rust 代码
- [x] 从 Console 的值来源选项、归一化分支、提示文案和专项测试中移除运行变量
- [x] 从 Rust SDK 路由与限流消费逻辑中移除机器环境变量读取
- [x] 发布新 specification 版本并更新 control-plane、Rust SDK 正式依赖
- [x] 完成测试、显式提交推送、all-mode 重建和真实页面验证

Review：

- specification 已发布 `v0.1.0-ALPHA.35`，枚举号 `2` 与名称 `VARIABLE` 均已保留不可复用；Go/Rust 生成代码及发布流水线通过。
- Rust SDK 已移除路由、限流对机器环境变量的读取，未知值类型按 fail-closed 处理；隔离暂存树完整测试通过，提交 `2b7b53f` 已推送 `develop`。
- Console 已删除运行变量选项及数值 `2` 的归一化映射，正式依赖已切换 ALPHA.35；隔离暂存树的 Go 全量测试、前端 ESLint 与 `build:test` 通过。
- control-plane 提交 `99e96eb9` 已推送 `develop`；all-mode 使用当前完整工作树重建，8080/8090 均返回 200。
- Playwright 使用 `admin/admin123` 打开本地限流规则创建页，新增匹配条件后展开“值来源”，下拉只显示“固定值 / 请求参数”，无“运行变量”；截图为 `output/playwright/governance-value-source-without-runtime-variable.png`。

## specification ALPHA.34 发布与依赖联动

- [x] 审核 specification 协议与 Go/Rust 生成代码差异，确认版本号和发布机制
- [x] 完成 specification 生成、测试、显式暂存、提交、推送和 `v0.1.0-ALPHA.34` 发布
- [x] 将 control-plane 的 Go 依赖更新到 ALPHA.34，移除仅用于本地联调的 replace 并完成回归
- [x] 将 Rust SDK 与 e2e 依赖更新到 ALPHA.34，刷新 lock 并验证治理动态参数及完整编译
- [x] 分别显式暂存、提交和推送 control-plane、Rust SDK 变更，记录最终 Review

Review：

- specification 已提交 `d50d691`，发布标签与 GitHub Release `v0.1.0-ALPHA.34`；Release-Rust 工作流成功，crates.io 已可检索和下载 `pole-specification 0.1.0-ALPHA.34`。
- specification 的 Go/Rust 生成脚本、`go test ./...`、Rust `cargo test` 均通过。
- Rust SDK 与 e2e 已切换到 canonical 仓库和 ALPHA.34，完整 `cargo test` 通过（155 个单元测试、6 个 public API 测试），提交 `da9381a` 已推送 `develop`。
- control-plane 已移除本地 specification replace，依赖切换到 ALPHA.34；在隔离 worktree 中完整 `go test ./...` 通过，提交 `33aa806c` 已推送 `develop`。
- 三个仓库均采用显式暂存，未把各自工作区中已有的其它未提交变更混入本次提交。

## 治理请求参数值类型统一闭环

- [x] 核对 `TEXT/PARAMETER/VARIABLE` 在 specification、Console、路由与限流数据面中的现状及缺口
- [x] 定义统一的“固定匹配、请求值提取、运行变量引用”契约，并明确各治理类型的可消费范围
- [x] 让共享条件编辑器与鉴权、路由、限流、泳道、镜像、Mock 读写完整保留值类型
- [x] 实现已具备数据面承载能力的动态限流/路由消费，拒绝或显式降级未支持的场景
- [x] 补齐前后端回归、文档、重建与真实页面验证

Review：

- specification 已有值类型契约，无需新增协议字段；Console 统一显示“固定值 / 请求参数 / 运行变量”，并由共享条件编辑器及所有治理调用方完整读写 `value_type`。
- Rust Proxyless SDK 将新式 `PARAMETER + 空 value` 解释为采集当前键：本地限流按采集值摘要拆分计数器，路由目标标签可消费同名采集值；旧式 `PARAMETER + 非空 value` 保留兼容读取另一参数键。
- xDS 当前不消费该动态语义，Console 和 ADR 均明确展示能力边界，没有静默伪装支持。
- Console 已通过值类型专项、鉴权/路由/限流/泳道专项、ESLint 和 `build:test`；Go 治理参数校验测试与 context-kg lint 通过。
- Rust SDK 在指向当前本地 specification 的临时副本中通过路由 8 项、限流 12 项测试；原工作区仍因远端旧 specification 标签缺少既有身份类型而无法直接编译，本次未擅自改动其正式依赖版本。
- all-mode 已重建，8080/8090 均可访问。真实浏览器在限流创建页新增匹配条件并切换为“请求参数”，确认固定值输入被采集提示替换、六列没有溢出、完整能力说明位于表格下方，浏览器控制台 0 错误/0 警告。截图：`output/playwright/governance-request-parameter-final.png`。

## 鉴权兼容模式恢复子规则 Header 匹配

- [x] 核对托管身份、全局 Custom Header 与原有子规则请求条件的存储及校验边界
- [x] 将“自定义 Header（兼容模式）”收敛为子规则内的请求 Header 匹配，而非规则级凭证
- [x] 清理错误的规则级 Header UI/提交路径并完成前后端、实页回归

Review：

- 已确认后端既有 `LEGACY_REQUEST_MATCH` 会校验每条策略的 `traffic_match_rule`，无需改变其存储或校验边界；历史 `CUSTOM_HEADER` 继续可解析，避免存量规则读取失败。
- Console 将“自定义 Header（兼容模式）”映射为 `LEGACY_REQUEST_MATCH`，并把每个子规则的请求条件作为提交载荷；规则级 Header 名、值、轮换提示和相关样式均已移除。
- 已通过 `npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test` 与 `go test ./pkg/goverrule/interceptor/paramcheck`；all-mode 已重建，8080/8090 均返回 HTTP 200。
- 真实浏览器以 `admin` 打开 `/governance/rules/create?kind=traffic-security`：第③段选择“自定义 Header（兼容模式）”后，只展示“在每个鉴权子规则中配置 Header 匹配条件”；白名单子规则下显示独立的 `HEADER / authorization / 完全匹配 / 匹配值` 条件，没有规则级 Header 名和值输入。截图：`.playwright-cli/page-2026-07-20T09-17-58-283Z.yml`。

## 鉴权认证方式折叠

- [x] 将认证方式接入统一折叠组件并保留现有编辑、只读逻辑
- [x] 在折叠标题展示当前认证模式摘要
- [x] 完成专项检查、构建、重建与真实页面验收

Review：

- 认证方式复用 `CollapsibleSection`，默认展开；关闭页面或切换规则时重置为展开，避免将旧页面的临时折叠状态带入新规则。
- 收起时分别显示“Pole 托管服务身份”“自定义 Header”或“旧版请求匹配”，正文表单和凭证提示不再占用编辑页面空间。
- 已通过 `npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test`。all-mode 重建完成且 8080/8090 返回 200；浏览器确认收起后正文隐藏并显示认证模式摘要。

## RPC 接口字段语义与布局修正

- [x] 核对 gRPC/Dubbo 的 `path`、`method` 协议契约与当前编辑映射
- [x] 使列标题、占位说明和列宽明确表达“接口在 path、方法在 method”
- [x] 增加专项回归并完成构建、重建和真实页面验收

Review：

- 协议契约保持 `path.value` 承载 gRPC service / Dubbo interface，`method` 承载可选 RPC 方法；原实现的数据读写正确，问题是混合表头与宽度分配让这个语义不清晰。
- 表头改为“HTTP 方法 / RPC 接口”和“HTTP 路径 / RPC 方法（可选）”；RPC 行将接口列扩展到最小 280px，方法列收敛为最小 160px。
- 专项回归已先以旧表头失败、再随修复通过；`npm run lint -- --quiet`、`npm run build:test` 通过。all-mode 重建完成且 8080/8090 返回 200；浏览器切换 GRPC 后确认 `helloworld.Greeter` 是接口输入、`SayHello` 是可选方法输入。

## 鉴权服务信息垂直可折叠布局

- [x] 将被调命名空间、服务名称改为上下顺序的紧凑表单
- [x] 将服务信息卡片接入统一折叠组件并提供已选服务摘要
- [x] 完成专项检查、构建、重建和真实页面验收

Review：

- 服务信息复用统一的 `CollapsibleSection`：默认展开，收起后显示“命名空间 / 服务名称”摘要，未选择时显示“未选择被调服务”。
- 可编辑和只读状态均按命名空间、服务名称的上下顺序渲染，并将表单宽度限制在 560px 内，避免宽画布把关联字段拆散。
- 已通过 `npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test`。all-mode 重建完成，8080/8090 均返回 200；真实浏览器确认字段的纵向顺序及收起、展开交互。

## RPC API 编辑顺序与可选方法

- [x] 将 gRPC、Dubbo 的编辑顺序调整为服务/接口、匹配类型、可选方法
- [x] 保持 HTTP 既有的方法、匹配类型、路径顺序，并让混合协议表头准确表达差异
- [x] 补充回归、构建、重建与真实页面验收

Review：

- RPC 的主匹配对象是服务名或接口名，方法只用于进一步收窄范围。因此 gRPC/Dubbo 行按“协议、服务/接口、匹配类型、可选方法、操作”渲染；HTTP 仍为“协议、方法、匹配类型、路径、操作”。
- `method` 空值保持合法，专项脚本已覆盖 gRPC 服务名存在、方法为空时的鉴权规则校验与提交载荷。
- 已通过 `npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test`。all-mode 重建完成且 8080/8090 均返回 200；真实浏览器验证 gRPC 行的服务名位于第二列，方法位于末列并标明可选。

## 治理 API 协议化编辑交互

- [x] 为 HTTP、gRPC、Dubbo 建立可测试的字段语义与默认值映射
- [x] 在鉴权、镜像与 Mock API 编辑行按协议渲染字段、占位说明与切换行为
- [x] 完成专项检查、构建、重建和真实页面验收

Review：

- `API` 的存储字段仍为 `protocol/method/path`，但编辑语义按协议明确区分：HTTP 使用方法下拉和 URI；gRPC 使用方法名和服务名；Dubbo 使用方法名和接口名。
- 切换协议时会重置 `method/path.value`，HTTP 回到 `GET`，RPC 保持空值，避免把 `GET /` 误保存为 RPC 匹配条件；匹配类型会被保留。
- 已通过 `npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test`。all-mode 重建完成且 8080/8090 均返回 200；真实浏览器验证鉴权创建页切换 Dubbo 显示 `getUser` / `com.example.UserService`，切换 gRPC 显示 `SayHello` / `helloworld.Greeter`。

## 鉴权规则 API 接口匹配语义收敛

- [x] 用专项脚本锁定 API 路径不包含值类型、参数匹配仍保留值类型的契约
- [x] 移除受保护接口的值类型控件与布局列，并在提交前清理遗留字段
- [x] 完成专项检查、构建、服务重建和真实页面验收

Review：

- `API.path` 在 protobuf 中复用了 `MatchString`，但 `value_type` 仅是请求参数匹配语义；鉴权接口读取和提交时均收敛为 `type/value`，protobuf 侧使用 `TEXT` 默认枚举值。
- 受保护接口与同屏接口编辑行均移除“值类型”列，保留协议、方法、匹配类型、接口路径和操作；请求参数匹配的 `value_type` 逻辑未改动。
- 已通过 `npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test`；all-mode 重建完成。真实浏览器以 `admin` 打开鉴权创建页，接口表头与行均为五列，未出现“值类型”。

## 治理独立创建页滚动与浮动控件修复

- [x] 建立鉴权创建页滚动与右侧空白控件的可重复浏览器复现
- [x] 修复独立创建页的高度/滚动归属，并移除遗留浮动操作容器
- [x] 添加回归约束，重建并在真实页面验证内部滚动和无空白控件

Review：

- 独立创建页为根、规则 frame、body 与表单内容区建立有限高度的 flex 链，滚动只保留在内容区；页面头部操作仍固定可用。
- `StickyTool` 的保存/撤销按钮已经通过 portal 渲染到页头，原固定容器以 `display: none !important` 隐藏，消除了右侧两个空白方块。
- 已通过 `npm run test:governance-direct-create`、`npm run lint -- --quiet`、`npm run build:test`。all-mode 已重建，8080/8090 返回 200；真实鉴权创建页内容区为 `1782/822px`（内容/可视高度），鼠标滚轮使内容区从 `720` 滚至 `960`，外层工作区保持 `0`，浮动容器计算样式为 `display: none`。

## 治理规则独立创建页

- [x] 复核规则详情独立页与现有九类创建编辑器的复用边界
- [x] 将工作台类型选择跳转到独立创建路由，移除创建抽屉状态和渲染
- [x] 让创建页支持返回工作台、取消和保存后的回流，并完成静态、构建和浏览器验收

Review：

- 新增隐藏路由 `/governance/rules/create?kind=…`，复用规则详情页的页面框架、操作栏和九类编辑器；工作台只保留规则类型选择弹窗，不再管理创建抽屉。
- 创建页在进入时按规则类型重置草稿；从服务详情嵌入工作台进入时，会把 `namespace/service/role` 带到 URL 并预填服务范围。取消或保存完成后均回到治理工作台。
- 已通过专项静态检查、ESLint、前端测试构建；all-mode 重建后 8080/8090 返回 200。真实浏览器点击“创建路由规则”后进入 `/governance/rules/create?kind=route`，显示完整创建页而非抽屉。

## 治理规则类型即选即建

- [x] 定位类型选择弹窗的二次确认与创建分发逻辑
- [x] 点击类型卡片后直接进入对应规则创建，并保留无障碍键盘操作
- [x] 更新回归检查、lessons 与任务记录，完成构建和浏览器验收

Review：

- 新建规则弹窗不再展示步骤条、类型选中态或“进入创建”二次确认；每个类型卡片现在是原生按钮，点击后关闭类型选择弹窗并直接打开对应规则编辑器。
- 新增 `test:governance-direct-create`，静态约束直接创建分发、无二次确认状态和卡片可访问名称。
- 已通过专项检查、ESLint 与前端测试构建；all-mode 重建后，8080/8090 返回 200。真实浏览器以 `admin` 打开治理工作台，点击“创建路由规则”后直接进入“新建路由规则”编辑器，未出现确认按钮。

## 治理鉴权规则编辑布局与服务选择

- [x] 定位受保护接口行及来源服务选择的布局、组件和高度约束
- [x] 将接口行改为可收缩的响应式字段布局，并将来源服务改为可搜索的多选下拉
- [x] 消除规则分段的无效满高拉伸，保留必要的内部滚动
- [x] 完成专项静态检查、lint、构建和真实浏览器验收

Review：

- `Select` 的 filterable 模式现会按输入关键字实际过滤候选项；来源服务按 `namespace/service` 排序，以紧凑的多选搜索下拉呈现。
- 受保护接口行移除过大的最小列宽，路径列吸收余量，删除按钮固定为 32px；在真实内容区宽度 694px 下，六列控件均完整可见。
- 编辑器 shell 改为自然高度，长内容仍由规则详情 Tab 内容区滚动，不再为内容稀少的规则制造整页空白。
- 已通过 Fluent 输入控件与鉴权编辑器专项脚本、ESLint、前端测试构建；all-mode 重建后 8080/8090 返回 200。真实浏览器验证 `spec-gateway` 搜索仅保留 `spec-governance/spec-gateway` 候选项，并验证接口行完整可见。

## 治理编辑器移除实时 Spec

- [x] 横向枚举路由、泳道、限流、熔断、探测、无损、鉴权、镜像和 Mock 编辑器中的实时 Spec 实现
- [x] 删除预览面板及其专用状态、格式化逻辑、样式和不再需要的回归断言，保留真实保存 payload 与编辑交互
- [x] 更新治理功能档案、技术约定、任务记录与 lessons，清除“实时 Spec 为必需”的旧表述
- [x] 完成专项静态校验、lint、构建、all-mode 重建及真实浏览器横向验收

Review：

- 横向复核确认：路由、泳道、限流、熔断、探测、无损、镜像和 Mock 编辑器此前已无预览节点；调用鉴权是唯一仍实际渲染右侧面板的编辑器，现已移除。九类编辑器的遗留预览样式一并清理。
- 新增 `test:governance-no-live-spec` 静态门禁，逐一检查九个编辑器和七个样式入口，不允许 `实时 Spec`、`specPane` 或 YAML/JSON 预览结构重新进入页面。
- 已通过 `npm run test:governance-no-live-spec`、`npm run test:traffic-security`、`npm run lint -- --quiet`、`npm run build:test`、context-kg lint 与范围化 `git diff --check`。
- all-mode 已使用最新静态资源重建；本地 MySQL 容器恢复后，8080 与 8090 均返回成功。真实浏览器以 `admin` 打开 `seed-20260616-security` 的编辑态，确认只有表单单栏、无右侧实时 Spec 面板；截图：`.playwright-cli/page-2026-07-20T01-15-32-917Z.png`。

## Toast 正文可读宽度

- [x] 根据截图定位底部操作区移除后 Toast 回落为默认窄宽度
- [x] 设置受视口约束的桌面正文宽度，避免错误消息过早换行
- [x] 构建、重建并在真实 400203 通知中验收

Review：

- Fluent Toaster 容器固定为 420px 桌面宽度，并以视口宽度减 32px 为上限，避免 Toast 自身超出承载容器。
- Fluent 默认将 `ToastBody` 限制在中间网格列；正文现跨越右侧空列，错误信息可使用完整可读宽度。
- 已通过静态约束、ESLint、测试构建和 `git diff --check`；重建 all-mode 后 8080、8090 返回 200。真实 400203 通知确认错误信息保持单行，复制和关闭按钮完整可见。

## 紧凑错误 Toast 操作区

- [x] 复核截图中的底部双复制按钮及其完整错误复制语义
- [x] 将复制完整错误 JSON 收敛为标题右上角的单一图标加文案按钮
- [x] 重建服务、验证视觉布局和剪贴板内容，并记录 review

Review：

- 错误 Toast 的底部操作区已移除；标题右侧只保留一个 Fluent 图标加“复制”文案按钮和关闭按钮。
- “复制”仍复制格式化 `code/message/requestId` JSON，避免为 Request ID 单独占用一个视觉动作；按钮有同名可访问标签。
- 已在真实 400203 Toast 截图确认紧凑布局，重建后 8080、8090 返回 200。

## Toast 自动与手动关闭

- [x] 核对错误、成功通知与 Fluent Toast 容器的超时和关闭能力
- [x] 为每条 Toast 增加 Fluent 手动关闭入口，并统一错误、成功通知为 10 秒自动关闭
- [x] 运行静态、构建和真实浏览器验收，并记录 review

Review：

- 每条 Toast 都生成独立 ID，并在 Fluent `ToastTitle` 右侧提供“关闭通知”按钮；点击只关闭当前通知。
- Toast 默认超时和成功通知超时均为 10 秒，错误通知继续显式保持 10 秒。
- 已在真实页面验证：错误通知可立即手动关闭，错误通知 11 秒后自动移除；登录成功通知同样拥有关闭按钮，并在手动操作前按 10 秒超时自动关闭。

## 结构化请求错误与国际化展示

- [x] 将请求失败从字符串拼接改为可序列化的 `code/message/requestId` 错误载荷
- [x] 用 Fluent Toast 分别展示错误码和国际化错误信息，并仅以操作按钮复制 Request ID
- [x] 为命名空间、配置分组删除链路保留结构化错误载荷，支持完整 JSON 复制
- [x] 补充静态与真实浏览器验证，重建 all-mode 服务并记录结果

Review：

- `RequestError.message` 只保留后端业务信息；`code` 和 `requestId` 以独立字段随 Redux reject payload 传递，消除了原先依赖字符串解析的主路径。
- Toast 以 Fluent UI 格式化展示“错误码 / 错误信息”，Request ID 不再出现在正文，仅支持点击复制；同时提供“复制错误 JSON”，内容稳定为 `code`、国际化后的 `message`、`requestId`。
- 已在真实 400203（命名空间仍存在服务）链路验收中文与英文消息，并确认剪贴板分别得到纯 Request ID 与格式化 JSON；all-mode 重建后 8080、8090 均返回 200。

## 配置分组删除拒绝与错误提示收敛

- [x] 建立并最小化“删除有配置文件的分组”返回 400201 的可重复反馈环
- [x] 定位删除链路的资源存在性校验，修正前端 `file_count` 到 `fileCount` 的响应映射与删除可用性不一致
- [x] 将失败通知收敛为紧凑 Fluent 提示，RequestId 仅提供复制操作
- [x] 添加回归测试，重建服务并在真实页面验收删除保护与错误提示

Review：

- `400201 existed resource` 是正确的保护：目标分组仍有有效 `app.yaml` 和活跃发布，后端不会级联删除配置。根因是前端遗漏 proto JSON 的 `file_count`，将真实的 1 显示为 0 并开放了删除按钮。
- 配置分组响应现归一化 `file_count`，含文件的分组直接禁用删除，并提示“请先删除 N 个配置文件”。
- 失败 Toast 统一从正文剥离末尾 RequestId，以 Fluent Footer 的“复制 Request ID”按钮提供排障标识；已用真实浏览器确认 UUID 不在正文且复制结果正确。
- 已通过两条新增静态约束、ESLint、测试构建、diff 检查；all-mode 重建后 8080、8090 返回 200。

## 命名空间工作台滚动与配置数量

- [x] 根据截图复现页面整体滚动，定位命名空间列表当前数据与高度链
- [x] 将页面滚动所有权下沉到命名空间表格内容区，固定页头、摘要、工具栏和分页
- [x] 在命名空间后端列表响应中聚合真实配置数量，并在前端摘要和表格中展示
- [x] 添加静态回归检查，执行后端/前端测试、重建服务与真实浏览器验收

Review：

- `Namespace.total_config_file_count` 使用新的协议字段号 23；后端通过一次 `CountConfigFileEachGroup` 调用汇总配置文件数，避免按命名空间 N+1 查询。
- 命名空间页采用与服务列表一致的固定工作区：页面高度固定为视口工作区，滚动只在 `.fluent-table-scroll` 内发生，分页保持可见。
- 已通过命名空间静态约束、抽屉和暗色 token、零 TDesign、ESLint、测试构建、Go 定向包测试与 `go build ./...`；all-mode 重启后 8080、8090 均返回 200。真实浏览器在 2048×1200 下确认页面高度等于视口、表格内容区 `803 > 687` 可滚动，滚轮后表格滚动 116px 而页面仍为 0px。

## 命名空间详情布局与全站暗色可读性修复

- [x] 建立真实浏览器复现，确认命名空间详情标签和值的布局错位，以及暗色主题文字对比度问题
- [x] 将命名空间详情表单切换为与标签宽度匹配的横向字段布局
- [x] 建立暗色主题兼容层，将常见浅色硬编码颜色映射为现有应用主题 token
- [x] 对命名空间、认证管理、配置管理、治理代表页面进行暗色浏览器验收
- [x] 运行静态验证、lint、构建并记录 review

Review：

- 详情抽屉的根因是 Fluent `Form` 默认纵向布局与 `labelAlign="right"` 组合，已改为 inline 布局；浏览器截图确认“名称 / 描述”标签和值同一行。
- 暗色问题来自页面级 LESS 覆盖 Fluent Provider 的文字、背景与边框语义色。已统一映射常见中性颜色到 `--app-*` token，并补齐亮/暗主题的三级文字、弱文字、危险色和悬浮表面 token。
- 已通过 `npm run test:namespace-drawer`、`npm run test:dark-theme`、`npm run test:no-tdesign`、`npm run lint`、`npm run build:test` 以及 `git diff --check`。
- 已用真实浏览器在 `/namespace`、`/auth/principals`、`/configuration/group`、`/governance/workbench` 验收暗色显示；all-mode 重启后 8080 与 8090 均返回 HTTP 200。

## Pebble/观测与客户端完整验证

- [x] 复现 Rust 客户端 e2e 的真实失败：`v1.DiscoverGRPC`/`v1.ConfigGRPC` 注册、服务发现 `get_one`、配置中心远端资源加载、反注册清理。
- [x] 修复 control-plane gRPC client API 注册，`service-grpc` 同时暴露 `DiscoverGRPC`、`PoleHeartbeatGRPC`、`ConfigGRPC`。
- [x] 修复配置发布默认 `release_type=normal`，并补齐 ConfigGRPC discover response 的 `code/revision/file_names/file_groups`。
- [x] 修复 Rust SDK 内存 cache 的 available 实例列表同步，避免 `get_all_instance` 可见但 `get_one_instance` 权重为 0。
- [x] 修复 client deregister 四元组校验，允许 SDK 使用 `namespace/service/host/port` 反注册。
- [x] 修复 Rust e2e control-plane client 的 token header，兼容后端原始 `Authorization` / `X-Polaris-Token`。
- [x] 完成 Go 全量测试、Console 前端检查、Rust workspace 测试、Rust clippy、all-mode 启动和核心客户端真实 e2e。
- [ ] 治理全量 e2e 仍需单独处理 control-plane plan payload：routing 缺 `@type`，ratelimit 参数不合法，部分治理资源 create/publish 标识不匹配导致 publish 404。

Review：

- 已验证 `go test -count=1 ./...` 全量通过。
- 已验证 `npm run lint -- --quiet && npm run test:metrics-observability && npm run build:test` 通过。
- 已验证 `cargo fmt --all -- --check`、`cargo test --workspace --all-features`、`cargo clippy --workspace --all-targets --all-features` 通过；clippy 仍有既有 warning，但退出码为 0。
- 已重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200，`8091` gRPC 正常监听。
- 已验证 Rust 客户端核心真实 e2e：`connectivity`、`service-discovery`、`config-center` 三项通过，覆盖 SDK context、服务注册/心跳/发现/选实例/反注册、配置 upsert/publish/get。
- 已尝试默认 12 项客户端 e2e 与 `--execute-governance-control-plane` 全量治理 e2e；核心三项通过，治理九项当前失败在测试 plan 与后端治理接口契约不匹配，不应计为本轮观测/Pebble/核心 SDK 链路已完成。

## 可观测性数据体系首轮对接

- [x] 复核现有可观测性 ADR、Console 监控入口、control-plane 内部 chain 和相邻 Rust client / sidecar 当前能力
- [x] 梳理首轮端到端接入切片，明确 Collector、存储、control-plane、SDK、sidecar 和 Console 的接口边界
- [x] 更新 ADR，记录先打通数据生产、采集、存储、查询的最小闭环
- [x] 新增 `deploy/observability` 本地栈，用 Docker Compose 拉起 GreptimeDB + OpenTelemetry Collector Contrib
- [x] 在 console 模块新增 `/observability/v1/platform/overview` 查询 provider skeleton 与配置入口
- [x] 当前实现：control-plane server 先完成 `statis` 的 `otel` chain entry，输出平台 metrics
- [x] 当前实现：control-plane server 增加 `history/discoverEvent` 的 `otel` event/audit entry
- [x] 当前实现：console `observability-query` provider 接入 GreptimeDB 平台 metrics 查询，并让系统监控页优先展示真实数据
- [ ] 后续实现：pole-client-rust 接入真实 OTel exporter、服务发现 endpoint 和 remote config 启动链路
- [ ] 后续实现：pole-sidecar 增加 OTel 上报模块并与治理执行链路打点

本轮 control-plane metrics 计划：

- [x] 新增 `plugin/observability/statis/otel` 插件，并注册为 `statis.entries[].name=otel`。
- [x] 映射 `ReportCallMetrics` 为 `pole.control_plane.request.*`、`pole.control_plane.store.request.*`、`pole.control_plane.cache.*` 等 `pole.*` 指标。
- [x] 映射 `ReportDiscoveryMetrics` 为 `pole.discovery.service.count`、`pole.discovery.instance.count`、`pole.control_plane.client.connection.count`。
- [x] 映射 `ReportConfigMetrics` 为 `pole.config.group.count`、`pole.config.file.count`、`pole.config.file.release.count`。
- [x] 接入默认部署配置，但保留 `local` 和 `prometheus` entry 兼容。
- [x] 补充单元测试、配置加载回归和本地 Collector/GreptimeDB smoke 验证。

本轮页面真实查询计划：

- [x] 后端 `observability-query` 默认 provider 从静态 DTO 改为 GreptimeDB SQL 查询。
- [x] `/observability/v1/platform/overview` 返回系统监控页可直接消费的组件行、摘要指标和时间序列。
- [x] 系统监控页优先调用 `/observability/v1/platform/overview`，失败或无数据时回退 mock 预览。
- [x] 保留前端不直连 GreptimeDB 的边界，所有真实查询走 console 模块。
- [x] 用本地 GreptimeDB 中的 `pole_control_plane_request_*` 表验证页面接口返回真实样本。

本轮系统监控看板调整计划：

- [x] 将 CPU/MEM 从接口明细表和组件混合看板中移出，独立成组件资源看板。
- [x] 为 `pole-control-plane` Go 服务端增加 Go runtime 看板，包括 goroutine、heap、GC、调度延迟等标准指标视图。
- [x] 扩展 `/observability/v1/platform/overview` 返回资源 metrics 与 runtime metrics，前端仍不直连 GreptimeDB。
- [x] 当 GreptimeDB 暂无 kubeletstats / Go runtime 表时保持空态或 0 值，不用 mock 伪装成真实资源指标。
- [x] 补充 Go provider 测试、前端静态检查、lint/build 和真实接口验证。

本轮系统监控布局修正计划：

- [x] 将 Go runtime 看板改为 `panelFull` 整行面板，避免落在双列 grid 左侧后右侧留空。
- [x] 将 runtime 指标卡改为更紧凑的桌面网格，减少运行时指标区块的纵向占用。
- [x] 增加前端静态约束，要求 Go runtime 看板保持整行布局。
- [x] 用真实浏览器验证 1280 与 1920 宽度下 runtime 面板均跨满所在 grid。

本轮 runtime 小图 hover 修正计划：

- [x] 将 `TinyTrend` 从静态 SVG 折线升级为支持最近采样点 hover 的轻量图表。
- [x] 在 hover 数据中展示指标名、采样点序号和按 runtime unit 格式化后的值。
- [x] 增加前端静态约束，防止图线回退成无交互静态线。
- [x] 用真实浏览器 hover Go runtime 小图，确认 tooltip 有数据。

本轮 Grafana-like 看板视图调整计划：

- [x] 将系统监控页头调整为 dashboard header，展示 dashboard 名称、时间范围、数据源和刷新动作。
- [x] 将筛选区调整为 Grafana variables 风格，显式展示 `$category`、`$api`、`$component`。
- [x] 将 panel 统一为带标题栏、query 元信息和内容区的 Grafana-like 面板。
- [x] 为接口延迟趋势补 hover 数据层，和 runtime 小图保持一致的可读交互。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮图表坐标轴补齐计划：

- [x] 为接口延迟趋势补充左侧 Y 轴刻度，保留底部时间线。
- [x] 为延迟热力图补充顶部时间轴和左侧接口/组件维度标签。
- [x] 为每个 Go runtime 小图补充简化 Y 轴刻度和底部时间线。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮 Grafana 官方口径校准计划：

- [x] 对齐 Grafana Time series 与 Stat panel 的职责差异：主趋势保留 x/y 轴，runtime 指标按 Stat + sparkline 呈现。
- [x] 移除 runtime 小卡片内显眼坐标刻度和大气泡标签，改为弱化的背景 sparkline、底部时间范围和 hover marker。
- [x] 优化 tooltip 为 Grafana 风格的紧凑 overlay，避免遮挡卡片主体。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮 Grafana variables 区优化计划：

- [x] 按 Grafana variables 默认呈现方式，将变量区从大卡片收敛为顶部轻量变量行。
- [x] 去掉变量控件的厚重前缀盒和整块阴影边框，变量名改为小标签，控件本体保持独立边界。
- [x] 保持 `$category`、`$api`、`$component` 顺序和重置动作，避免破坏查询行为。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮 runtime Stat sparkline 比例修正计划：

- [x] 调整 Go runtime Stat 卡片的高度和内部排布，让当前值与趋势图形成合理比例。
- [x] 扩大 `TinyTrend` 的逻辑画布与可视高度，避免曲线变成过短的底部横线。
- [x] 保持 Stat panel 形态、时间范围和 hover tooltip，不回退为完整坐标轴小图。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮 Grafana 视觉风格对齐计划：

- [x] 将 runtime Stat 卡片从 Fluent 卡片样式收敛为 Grafana panel 风格：薄边框、低圆角、平面背景、紧凑标题。
- [x] 将 sparkline 改为 Grafana Stat 常见的下半区 area sparkline，保留当前值为主视觉。
- [x] 弱化 runtime 分类标签和说明文案，避免和单值指标争夺主视觉。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮系统监控时间轴时区修正计划：

- [x] 移除系统监控页静态 `TIME_LABELS`，按当前查询窗口生成时间标签。
- [x] 所有图表、热力图和 hover tooltip 使用浏览器本地时区格式化时间。
- [x] 在 dashboard header 显示当前本地时区，避免误解为 UTC 展示。
- [x] 补静态约束、前端构建和真实浏览器截图验证。

本轮操作审计与服务事件 OTel logs 接入计划：

- [x] 在 `history` chain 增加 `otel` entry，将操作审计映射为带 `event.name` 的 OTel LogRecord。
- [x] 在 `discoverEvent` chain 增加 `otel` entry，将服务事件映射为 OTel LogRecord，并写入 Collector logs pipeline。
- [x] OTel logs entry 必须 fail-open：请求链路只入本地 batch queue，Collector 异常不能阻塞业务、不能影响旧 `logger/rds` entry。
- [x] 增加本地持久化 spool 队列：`Put` 写入 bounded Pebble queue，后台 consumer 成功写入 Collector 后删除已确认 key，Collector 短暂不可用时可恢复补发；默认 history/discover-event 共享一个 Pebble DB 并用 key prefix 隔离。
- [x] 扩展 console `/observability/v1` 查询接口，从 GreptimeDB `pole_events` 读取操作审计与服务事件。
- [x] 让操作审计、服务事件页面优先读取 `/observability/v1`，旧 `/metrics/v1` 仅作为兼容兜底。
- [x] 补充 Go/前端测试、Collector/GreptimeDB smoke 和 context-kg review。

当前判断：

- 后台存储选型已确定：长期默认是 GreptimeDB，OpenObserve 只作为 quickstart / 可选 provider；Collector 仍是统一 OTLP 接收、处理、过滤和导出入口。
- control-plane 现有 `history`、`discoverEvent`、`statis` 已经形成内部观测 chain，首轮应补 `otel` entry，而不是新建并列插件体系。
- `pkg/common/otel/` 已有 OTLP gRPC exporter 和部分本进程指标，但指标名仍有旧命名，且启动配置、logs/event exporter、trace instrumentation 尚未完整接入。
- Console 当前只有 `/metrics/v1` 的历史事件和操作审计入口；新服务监控、系统监控、event、audit 查询应由 console 模块提供 `/observability/v1` 后端适配层。
- `pole-client-rust` 已经有 observability 语义模型、Resource attributes、低基数 metrics 过滤和 no-op recorder；但真实 OTel exporter、服务发现 endpoint 热更新、配置中心 remote config 启动/热更新链路还需要接入。
- `pole-sidecar` 当前没有可用 OTel 上报模块，首轮需要补流量请求、治理命中/拒绝、结构化 event 和 trace propagation 的统一打点位置。
- `deploy/observability` 已提供本地快速体验路径：Collector 接收 `4317/4318`，写入 GreptimeDB `4000/v1/otlp`；logs 默认进入 `pole_events`，只承载结构化 event/audit。
- console 模块已暴露 `/observability/v1/platform/overview`，并通过 GreptimeDB 查询 `pole_control_plane_request_*` 指标表；系统监控页优先使用真实返回，异常或无数据时回退 mock 预览。
- 系统监控页已将 CPU/MEM 拆成独立组件资源看板；真实模式下只使用 `/observability/v1/platform/overview.resources`，当前本地 GreptimeDB 没有 kubeletstats 资源表，因此显示资源指标空态。
- `pole-control-plane` Go runtime 看板已使用 `/observability/v1/platform/overview.runtime` 返回的 GreptimeDB 真实指标，覆盖 goroutine、heap alloc/inuse/sys、GC count、GC pause，并预留 schedule latency 标准指标查询。
- control-plane metrics 已按现有 `statis` chain 增加 `otel` entry；启动时先初始化 statis chain，再注册本进程旧 helper 指标，确保全局 OTel MeterProvider 已就绪。
- 系统监控 variables 区已从大卡片和 `$name` 前缀盒调整为轻量顶部变量带，保留 `$category`、`$api`、`$component` 顺序，变量名以小标签形式辅助识别。
- Go runtime Stat 卡片已扩大 sparkline 逻辑画布和可视高度，趋势线占满卡片有效宽度，当前值与图线不再比例失衡。
- Go runtime Stat 卡片已进一步对齐 Grafana panel 风格：低圆角薄边框、紧凑标题、弱化分类标签，并使用下半区 area sparkline 表达趋势。
- 系统监控页已移除固定时间标签，图表、热力图和 runtime tooltip 的时间轴都从当前查询窗口动态生成，并按浏览器本地时区显示；header 展示 `UTC+08:00` 等本地时区标识。

验证：

- 已静态复核 `context-kg/technical/adr/observability/`、`console/pkg/router/metrics_router.go`、`pkg/common/otel/`、`plugin/observability/`、`../pole-client-rust/src/observability/` 和 `../pole-sidecar/src`。
- 已执行 `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py context-kg`，通过 Markdown/frontmatter/link/index 基础检查。
- 已执行 `git diff --check` 检查本轮相关 context-kg 文件，未发现 whitespace error。
- 已执行 `docker compose -f deploy/observability/docker-compose.yaml config`，Compose 配置通过。
- 已执行 `docker compose -f deploy/observability/docker-compose.yaml up -d`，`pole-greptimedb` 与 `pole-otel-collector` 均正常启动。
- 已验证 `http://127.0.0.1:13133/` Collector health 返回可用，`http://127.0.0.1:4000/health` GreptimeDB health 返回成功。
- 已通过 OTLP/HTTP 向 Collector `4318/v1/logs` 上报 smoke event，并在 GreptimeDB `public.pole_events` 查询到 `pole observability smoke event`。
- 已执行 `go test ./bootstrap/config ./console/bootstrap ./console/pkg/observabilityquery ./console/pkg/handlers ./console/pkg/router`，通过 console 查询入口和配置加载回归。
- 已执行 `go test -count=1 ./apis/observability/statis ./plugin/observability/statis/... ./pkg/common/otel/... ./bootstrap/config ./bootstrap`，通过 statis chain、otel entry、配置加载和 bootstrap 编译回归。
- 已执行 `POLE_OTEL_COLLECTOR_ENDPOINT=127.0.0.1:4317 go test -count=1 ./plugin/observability/statis/otel -run TestStatisWorkerExportsToCollector -v`，通过本地 Collector 导出 smoke。
- 已在 GreptimeDB 查询到 `pole_control_plane_request_count_total`、`pole_control_plane_request_duration_seconds_*` 表，并确认样本包含 `pole_api_name=OtelSmoke`、`pole_component=apiserver`、`pole_result=success`。
- 已执行 `go test -count=1 ./console/pkg/observabilityquery ./console/pkg/handlers ./console/pkg/router`，通过 GreptimeDB provider、handler 和 console router 回归。
- 已执行 `npm run test:metrics-observability`、`npm run lint -- --quiet`、`npm run build:test`，通过系统/服务/event/audit 页面静态约束、ESLint 和前端测试构建。
- 已执行 `npm run test:metrics-observability`、`npm run lint -- --quiet`、`npm run build:test`，通过 Grafana variables 轻量结构约束、ESLint 和前端测试构建。
- 已通过 Playwright 在 `http://127.0.0.1:8080/metrics/system` 验证 variables 区不再以厚重外层卡片展示，截图保存在 `output/playwright/system-monitor-grafana-variables.png`。
- 已通过 Playwright 在 `http://127.0.0.1:8080/metrics/system` 验证 Go runtime Stat sparkline 比例，截图保存在 `output/playwright/system-monitor-runtime-sparkline-ratio.png`。
- 已通过 Playwright 在 `http://127.0.0.1:8080/metrics/system` 验证 Grafana-like Stat 视觉风格，截图保存在 `output/playwright/system-monitor-grafana-stat-style.png`。
- 已通过 Playwright 在 `http://127.0.0.1:8080/metrics/system` 验证本地时区时间轴，页面 header 显示 `UTC+08:00`，主图、热力图和 runtime sparkline 显示 `16:50 → 17:50` 本地时间，截图保存在 `output/playwright/system-monitor-local-timezone.png`。
- 已重建并启动 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；`GET /observability/v1/platform/overview?category=control-plane` 返回 `provider=greptimedb`、`configured=true`、真实 `components`。
- 已用浏览器登录 `admin/admin123` 进入 `/metrics/system`，页面显示“实时数据”，接口明细表渲染 `pole-control-plane`、`OtelSmoke`、`POST:/auth/v1/user/login`、`GET:/core/v1/namespaces` 等 GreptimeDB 样本行。
- 已再次执行 `go test -count=1 ./console/pkg/observabilityquery ./console/pkg/handlers ./console/pkg/router`，验证 `overview.runtime` Go 指标查询与 `/observability/v1` 路由兼容。
- 已再次执行 `npm run test:metrics-observability`、`npm run lint -- --quiet`、`npm run build:test`，验证系统监控新增资源看板、Go runtime 看板、真实/Mock 边界和前端构建。
- 已重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；`GET /observability/v1/platform/overview?category=control-plane` 返回 `components=4`、`resources=0`、`runtime=7`，其中 Go runtime 包含 `process.runtime.go.goroutines`、`process.runtime.go.mem.heap_alloc`、`process.runtime.go.mem.heap_inuse`、`process.runtime.go.mem.heap_sys`。
- 已用浏览器登录 `admin/admin123` 进入 `/metrics/system`，页面显示“实时数据”；顶部 stat card 显示 Go 协程数，Go runtime 看板展示 Goroutines、Heap Alloc、Heap Inuse、Heap Sys、GC Count、GC Pause Avg、Heap Objects；组件资源看板显示“暂无 CPU/MEM 资源指标”，接口明细表不再包含 CPU/MEM 列。
- 已再次执行 `npm run test:metrics-observability && npm run lint -- --quiet && npm run build:test`，通过 Go runtime 整行面板静态约束、ESLint 和前端测试构建。
- 已再次重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；真实浏览器在 `/metrics/system` 验证 Go runtime 面板 `panelWidth=1607`、所在 `gridWidth=1607`、`leftGap=0`、`rightGap=0`，截图保存为 `output/playwright/system-monitor-runtime-layout-wide.png`。
- 已再次执行 `npm run test:metrics-observability && npm run lint -- --quiet && npm run build:test`，通过 runtime 小图 hover 静态约束、ESLint 和前端测试构建。
- 已再次重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；真实浏览器 hover `Heap Alloc` 小图后显示 `Heap Alloc / 采样点 6 / 23.28MiB`，截图保存为 `output/playwright/system-monitor-runtime-hover-tooltip.png`。
- 已再次执行 `npm run test:metrics-observability && npm run lint -- --quiet && npm run build:test`，通过 Grafana-like dashboard header、variables、panel query metadata 和主趋势图 hover 静态约束。
- 已再次重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；真实浏览器进入 `/metrics/system` 后显示 `Dashboards / Platform / Control Plane`、`Last 1 hour`、`$category/$api/$component`、panel query 标签，并且 hover 接口延迟趋势显示 `p95 latency / 12:25 / 1ms`，截图保存为 `output/playwright/system-monitor-grafana-dashboard.png`。
- 已再次执行 `npm run test:metrics-observability && npm run lint -- --quiet && npm run build:test`，通过主趋势图坐标轴、热力图时间轴/接口维度、runtime 小图坐标轴静态约束。
- 已再次重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；真实浏览器进入 `/metrics/system` 后确认接口延迟趋势显示 `2ms/1ms/0ms` 与 `12:00/12:30/12:55`，热力图显示 `12:20-12:55` 与接口维度，runtime 小图显示纵向刻度和 `12:00/12:55`，截图保存为 `output/playwright/system-monitor-dashboard-axes.png`。
- 已按 Grafana 官方口径校准：Time series 面板保留 x/y 轴，Stat 面板展示大数值和可选 sparkline，不再把完整坐标轴塞进 runtime 小卡片。
- 已再次执行 `npm run test:metrics-observability && npm run lint -- --quiet && npm run build:test`，通过 runtime stat sparkline 静态约束、ESLint 和前端测试构建。
- 已再次重建 all-mode，`http://127.0.0.1:8080/` 与 `http://127.0.0.1:8090/` 均返回 200；真实浏览器进入 `/metrics/system` 后确认 runtime 指标呈现为大值 + sparkline + 时间范围，hover `GC Pause Avg` 显示 `GC Pause Avg / 12:25 / 1.88ms`，截图保存为 `output/playwright/system-monitor-grafana-stat-sparkline-hover.png`。

Review：

- 已在 [[adr-otel-observability-platform]] 增加首轮对接切片，明确 control-plane server、Rust SDK、sidecar、Collector、GreptimeDB、console 模块 `/observability/v1` 和 Console 页面的真实数据闭环。
- 已在 [[adr-otel-observability-platform]] 补充 SDK 与 sidecar 上报分工，要求用 `pole.node.role=sdk|sidecar` 区分来源，并避免同一请求和同一治理决策重复计数。
- 已在 [[adr-pole-rust-client-observability]] 记录 Rust SDK 首轮模块拆分、接入点、与 sidecar 共存规则和当前实现差距。
- 已新增 `deploy/observability/docker-compose.yaml`、`deploy/observability/otel-collector-config.yaml` 与 README，本地可以直接启动 GreptimeDB + Collector 并验证 OTLP event 写入。
- 已新增 console `observabilityquery` GreptimeDB provider、handler、router 和配置项，`/observability/v1/platform/overview` 由 console 模块负责提供，并保持前端不直连 GreptimeDB。
- 系统监控页已接入 `services/observability.describePlatformOverview`，优先展示真实数据标签与 GreptimeDB 返回行；当 provider 未配置、无数据或请求失败时继续保留 mock 预览，便于本地未启动观测栈时体验页面。
- 系统监控页的资源与运行时已经拆分：CPU/MEM 独立进入组件资源看板，Go runtime 独立进入 `pole-control-plane` 服务端看板；真实模式下组件资源不再从 mock 行推导，避免把未接入的 kubeletstats 指标伪装成真实数据。
- Go runtime 看板已从双列 grid 左侧卡位调整为整行看板，并补充静态约束和真实浏览器尺寸验证，避免桌面布局出现右侧大面积空白。
- Go runtime 小图已支持 hover/focus 数据层，悬浮时展示指标名、采样点和值，避免看板只有趋势形状而无法读取具体数据。
- 系统监控页已改为 Grafana-like dashboard 视图：dashboard header 承载时间窗、步长、数据源和刷新，variables 区承载变量筛选，每个 panel 标题栏展示 query 元信息，主趋势图和 runtime 小图都具备 hover 数据层。
- 系统监控页所有图表已补齐基本坐标语义：主折线图有 Y 轴刻度和时间线，热力图有时间轴和接口/组件维度，runtime 小图有简化 Y 轴和时间线。
- 系统监控页已按 Grafana Time series / Stat panel 职责重新校准：主趋势图承担完整坐标轴，runtime 单值指标使用 Stat + sparkline，减少小面板内的刻度和标签噪音。
- 已新增 control-plane `statis/otel` entry，真实输出 API、Store、内部组件、缓存、服务发现、配置中心和客户端发现调用指标；默认部署配置启用 `local + otel + prometheus`，保持旧 logger/prometheus 兼容。
- 已新增 control-plane `history/otel` 和 `discoverEvent/otel` entry，操作审计和服务事件都按 OTel logs 写入 Collector logs pipeline；新增 logs exporter 采用 bounded queue / 本地 Pebble spool + 后台批量发送，Collector 异常不阻塞业务链路，恢复后可继续补发未确认记录。默认两个 OTel entry 共享 `./data/observability/otel-events/otel-events.pebble`，分别使用 `history`、`discover_event` key prefix 隔离。
- 已新增 console `/observability/v1/events` 和 `/observability/v1/operations`，从 GreptimeDB `pole_events` 查询 `pole.event.kind=service|audit` 的结构化日志，并保持 `/metrics/v1` MySQL 历史接口作为页面兜底。
- 已重建 all-mode 并验证真实接口：`/observability/v1/events?namespace=default&service=checkout&event_type=InstanceOffline` 与 `/observability/v1/operations?resource_type=Routing&operation_type=Update&operator=admin` 均返回 GreptimeDB 样本，时间按本地时区正常显示。

## RequestId d42dffb6 失败定位

- [x] 检索 tmux、运行日志和本地文件中的 RequestId
- [x] 确认失败接口、HTTP/业务错误码及服务端堆栈
- [x] 定位根因并修复；若是运行环境问题则给出可复现证据
- [x] 通过真实请求验证并补充 Review

当前判断：

- `d42dffb6-1680-4e37-abc0-dbf676441e35` 是删除 `demo-governance` 的请求。后端以 `400203` 拒绝，`info` 为 `some services existed in namespace`，并在 namespace 服务记录“待删除命名空间仍存在服务”。这是领域完整性保护，不是服务端异常。
- Console 请求层在内部捕获 HTTP 错误后抛出带 `info` 的 Error，却被外层 `catch` 再次包装成通用“请求失败, RequestId”，因此用户无法看到拒绝原因。现已收敛为单次错误包装，统一显示业务码、后端 `info` 与 RequestId。

Review：

- 根因：该 RequestId 对应 `POST /core/v1/namespaces/delete` 删除 `demo-governance`。该命名空间仍有 4 个服务，后端按领域完整性返回 `400203 / some services existed in namespace`，拒绝删除以避免留下孤立服务。
- 修复：四种 HTTP 请求包装统一保留后端标准响应的 `code`、`info` 和 RequestId，避免内部格式化 Error 再被外层 `catch` 泛化；失败提示现在可直接给出可操作原因。Tooltip 内容层统一禁用指针事件，避免悬浮提示遮挡 Popconfirm 的“确认”按钮。
- 稳定性补充：Console 路由不再通过 Gin `LoadHTMLGlob` 缓存启动时的 `index.html`，而是按请求直接读取当前静态入口；重建后 hash 资源更新不会再导致旧入口引用不存在的 JS 文件并出现空白页。
- 验证：以 `admin/admin123` 真实登录，在命名空间列表确认删除 `demo-governance` 后，页面展示“请求失败（400203）：some services existed in namespace，RequestId: ce300c8c-1a48-4e0b-8ba8-47ad4a54b523”，且命名空间仍存在。`go test ./console/pkg/router ./console/pkg/handlers`、`npm run lint`、`npm run build`、`git diff --check` 通过；all-mode 的 `8080/8090` 均返回 `200`，运行中的 `8080` 页面入口与最新 `dist/index.html` 一致。

## AI 资源覆盖新增授权补齐

- [x] 核对 MCP/A2A 策略资源字段、Console 编辑器与详情页的通配表达
- [x] 核对 specification、默认策略生成、策略存在性校验与授权匹配是否完整支持 `id="*"`
- [x] 修复 MCP/A2A 默认策略或资源授权范围遗漏，并补最小回归测试
- [x] 用管理员真实页面验证 MCP/A2A 显示“覆盖新增资源”，并确认新建资源可访问
- [x] 补充 Review 与 lessons

当前判断：

- `mcp_servers`、`a2a_agents` 已是独立资源字段，Console 编辑器也能将“全部”提交为 `id="*"`；默认策略详情仍显示“仅显式资源”，说明通配资源没有被默认策略生成链路写入，或后端未将其按通配资源解释。
- 已确认历史主账号默认策略只含当时创建的 MCP/A2A 具体 ID。默认策略创建时会枚举当时已有的 `ResourceType`，而后续新增资源类型不会回填到既有策略；策略容器本身已支持 `id="*"` 命中未来资源。

Review：

- 根因：MCP Server、A2A Agent 的 resource type 和 Console 全量选择早已支持 `id="*"`，但管理员默认策略是在 AI 资源类型加入前创建的。历史策略只在资源创建后得到具体 ID，缺少 `*`，所以详情只能显示“仅显式资源”，未来资源也不会继承。
- 修复：鉴权策略初始化时只扫描主账号默认策略与两条系统全量策略；对缺失的 MCP/A2A 通配资源执行幂等 `LooseAddStrategyResources`。普通自定义策略不在匹配范围内，不会被自动扩大授权。
- 补充：策略新建/更新的资源存在性检查现覆盖 MCP/A2A；缓存层新增回归，证明 MCP/A2A 的 `*` 会命中策略创建后出现的资源 ID。
- 验证：`go test ./plugin/access_control/... ./pkg/cache/auth/...`、`git diff --check` 通过；all-mode 重启完成且 `8080/8090` 返回 `200`。`admin/admin123` 真实页面的默认策略资源页中 MCP Server、A2A Agent 均显示“覆盖新增资源”，打开详情后均显示“全部（包括新增）/包括新增”，浏览器 console 无 error/warning；验收截图为 `output/playwright/auth-policy-ai-wildcard.png`。

## 权限策略成员页签宽度回归修复

- [x] 根据截图复现成员信息页签内容收缩，并定位 Fluent 活动 panel 的横向 flex 行为
- [x] 将所有活动页签统一为满宽纵向内容容器，且不影响资源页的受限高度和左栏滚动
- [x] 增加成员信息页签满宽布局静态回归约束
- [x] 重建后验证成员、资源、标签和接口页签布局及资源左栏滚动
- [x] 补充 Review 与 lessons

当前判断：

- `fluent-tab-content` 的活动 TabPanel 不能只设置 `display:flex`；当成员页含 `.policyTabBody` 包装层时，默认 `flex-direction: row` 会按内容宽度收缩。活动 panel 必须为 `flex-direction: column` 且占满可用宽度。

Review：

- 根因：上一轮为了资源页的受限高度，将活动 TabPanel 改为 flex 容器，但没有指定方向；资源页的双栏 grid 作为直接子元素会自然撑满，成员页的 `.policyTabBody` 则在默认横向 flex 中按内容宽度收缩，造成右侧大面积空白。
- 修复：活动 TabPanel 统一增加 `width: 100%`、`min-width: 0` 和 `flex-direction: column`，使成员、资源标签和接口表都沿纵向布局并自动拉伸到内容区宽度；资源页仍通过自身 `flex: 1` 保持受限高度与独立滚动。
- 补充：横向验证时发现可访问接口表格只为状态列提供数字宽度，导致共享百分比列宽适配器将其独占为 `100%`；现为接口分组、访问范围和状态分别声明 `30 / 50 / 20` 的比例宽度。
- 验证：认证详情、策略详情、Fluent 表格布局校验、ESLint 与 `git diff --check` 通过；all-mode 重启后 `8080/8090` 返回 `200`。`admin/admin123` 真实页面中成员信息为满宽布局，资源标签正常显示空态，可访问接口三列完整显示；资源类别左栏 `clientHeight=157`、`scrollHeight=1110`，滚轮后 `scrollTop=920`，console 为 0 error/warning。

## 权限策略资源类别滚动修复

- [x] 建立真实策略资源页签的左栏滚动复现与高度测量
- [x] 修复独立详情页中资源类别列表的滚动高度链和滚轮命中
- [x] 增加滚动容器静态回归约束
- [x] 用真实滚轮把左栏滚到末项并完成构建、重启验证
- [x] 补充 Review 与 lessons

当前判断：

- 当前独立详情页沿用抽屉版的固定高度与 Tabs 内部滚动；左侧 `.resourceTypeList` 虽声明 `overflow: auto`，但必须用真实 DOM 的 `scrollHeight / clientHeight / scrollTop` 和鼠标滚轮确认它确实是唯一滚动所有者。

Review：

- 根因：策略详情迁移到 Fluent 后，页面仍使用 `.t-loading__parent`、`.t-tabs__content` 和 `.t-tab-panel` 建立高度链；真实 DOM 使用 `.fluent-loading` 与 `.fluent-tab-content`，规则没有命中，资源类别列表被内容完整撑到 `1110px`，其 `clientHeight` 与 `scrollHeight` 相等，因此无法滚动。
- 修复：共享 `Loading` 透传 `className`，策略详情为 Loading 设置专用高度链 class；Tabs 内容区和活动 panel 改为依据 Fluent 真实结构的 flex 容器，资源 shell 保持受限高度，左侧 `.resourceTypeList` 成为唯一 `overflow: auto` 的滚动所有者。
- 验证：定向布局脚本、策略详情脚本、ESLint、前端 release 构建和 `git diff --check` 通过；all-mode 重启日志出现 `finish starting server`，`8080/8090` 均返回 `200`；使用 `admin/admin123` 进入默认策略“资源信息”页签，左栏 `clientHeight=157`、`scrollHeight=1110`，真实滚轮后 `scrollTop` 从 `0` 变为 `920`，末项“角色 / MCP Server / A2A Agent”可见。

## 认证详情页设计统一

- [x] 复核策略详情页、用户详情页与已统一的 MCP/服务详情页设计语言
- [x] 用唯一的策略详情视图替换独立页残留的旧 `Descriptions` 实现
- [x] 重构用户详情为身份摘要、访问凭据、标签与关联策略工作区
- [x] 补充静态约束，防止旧布局和对象字符串回归
- [x] 运行 lint、构建和真实浏览器截图验证
- [x] 补充 Review 和验证结果

当前判断：

- 认证管理独立详情页必须遵循资源详情页的页面壳、身份摘要、信息分区和表格工作区，不再使用旧式 `Card + Descriptions + 空 Tabs` 拼接。
- 策略详情已有 `PolicyDetailView` 作为抽屉和页面共用的唯一实现，独立页只负责路由上下文和可用高度，避免两份资源、标签和成员逻辑继续漂移。

Review：

- `PolicyDetail.tsx` 已收敛为面包屑与 `PolicyDetailView` 的页面容器，移除了旧资源树、`Descriptions`、成员卡片和标签渲染逻辑；因此策略 metadata 不会再被模板字符串渲染为 `[object Object]`。
- `UserDetail.tsx` 现在以身份摘要、访问凭据、标签和权限表四个层次展示信息；凭据重置、Token 启停、复制和关联策略跳转均保留。关联策略名称列显式分配宽度，浏览器截图确认链接可见且可点击。
- 静态验证覆盖独立详情页不回退到旧布局、策略详情复用唯一实现、用户 Token 操作和策略跳转锚点；实际使用 `admin/admin123` 登录后验证用户详情与默认策略的成员、资源页签。

## 输入控件连续输入修复

- [x] 复现输入控件无法连续输入的具体页面和字段
- [x] 定位焦点丢失、组件重挂载或状态写回根因
- [x] 修复输入控件连续输入问题并补静态/浏览器回归校验
- [x] 运行 lint、构建、重启并用真实浏览器验证
- [x] 补充 Review 和验证结果
- [x] 补充全量输入控件巡检：共享输入组件、动态行 key、主要页面真实输入矩阵

当前判断：

- 用户反馈的是输入交互断裂问题，必须用逐字符键盘输入验证，不能只用 `fill` 或看表单是否能渲染。
- 优先在最近改动的服务详情内嵌编辑表单复现，因为该页面刚从抽屉编辑改为页面内编辑，更容易出现表单组件重挂载导致焦点丢失。

Review：

- 已复现：服务详情编辑页标签键输入 `labelkey` 时，修复前实际只保留 `l`，且 `document.activeElement` 已离开标签键 input。
- 已复现：服务列表命名空间 filterable Select 输入 `spec` 时，修复前 input value 仍为空，无法连续过滤输入。
- 根因 1：`LabelInput` 标签行 `key` 使用 `${index}-${item.key}`，输入标签键时每个字符都会改变 key，React 卸载并重建整行导致焦点丢失。
- 根因 2：Fluent `Select` 将 Combobox `value` 固定为选中项 `displayValue`，filterable/creatable 场景没有独立输入文本状态，键盘输入会被展示值覆盖。
- 已完成：`LabelInput` 行 key 改为稳定的 `label-row-${index}`；Fluent `Select` 增加 `inputValue`，在键盘输入、选项选择和 blur 时同步。
- 已完成：新增 `verify-fluent-input-controls.mjs`，约束标签行 key 不能再包含正在编辑的标签键，并约束 filterable/creatable Select 必须维护输入态。
- 已验证：`node scripts/verify-fluent-input-controls.mjs`、`npm run lint`、`npm run build:test`、`git diff --check` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/` 和 `http://127.0.0.1:8090/` 均返回 200。
- 已验证：使用 `admin/admin123` 真实 Chromium 逐字符输入，服务列表命名空间下拉可连续输入 `spec` 且保持焦点，服务详情标签键可连续输入 `labelkey` 且保持焦点；浏览器 console warning/error 为 0。

补充巡检：

- 已完成：静态扫描 `Input / Textarea / Select / TagInput / RangeInput / InputNumber / LabelInput / ClientLabelInput` 使用点，重点检查会导致输入行重挂载的动态 `key`。
- 已完成：治理鉴权/镜像编辑器中 Header 行、接口行、Mock/Mirror 规则卡、泳道编辑器中的泳道卡和组内服务行 key 改为稳定 index key，避免编辑协议、方法、路径、服务名时重挂载。
- 已完成：`verify-fluent-input-controls.mjs` 扩展覆盖治理鉴权/镜像编辑器和泳道编辑器的输入行 key 规则。
- 已验证：普通文本 Input 覆盖服务列表搜索、服务详情描述/部门/业务、实例主机/版本/位置、A2A Agent 名称/命名空间/技能 ID/技能名。
- 已验证：filterable Select 覆盖服务列表命名空间，连续输入 `spec` 后值完整且焦点保持。
- 已验证：InputNumber 覆盖实例端口和权重，连续输入 `8080`、`88` 后值完整且焦点保持。
- 已验证：LabelInput 覆盖服务详情标签键值、实例标签键值，连续输入后值完整且焦点保持。
- 已验证：TagInput 覆盖 A2A Skill 标签/示例，输入后回车生成标签且输入框状态正常。
- 已验证：Textarea 覆盖 A2A Skill 描述，连续输入后值完整且焦点保持。
- 说明：RangeInput 当前没有在主要页面默认可见路径中自然出现，本轮通过共享组件源码和动态 key 静态规则检查覆盖，没有发现会因输入值变化重挂载的风险。

## 服务详情冗余系统信息/运行状态移除

- [x] 移除服务详情底部“系统信息 / 运行状态”双栏区域
- [x] 清理不再使用的详情键值样式、状态样式和回归脚本断言
- [x] 更新 lessons，记录服务详情不要重复堆叠已覆盖的信息区
- [x] 运行定向校验、lint、构建、重启并用真实浏览器验证
- [x] 补充 Review 和验证结果

当前判断：

- 服务详情已经有身份头部、四项摘要指标和内嵌服务基础信息表单，底部“系统信息 / 运行状态”会重复展示 ID、命名空间、实例健康、别名等信息。
- 本次应直接移除这块重复信息区，而不是继续微调样式或保留空占位。

Review：

- 已完成：`ServiceDetail` 删除系统信息和运行状态两个详情区块，清理 `DetailItem`、端口格式化、进度条和状态 Tag 等只为该区块存在的代码。
- 已完成：`index.module.less` 删除服务详情旧键值布局和运行状态布局，详情内容区收敛为单列内嵌服务表单。
- 已完成：`verify-discovery-services-layout.mjs` 从要求双栏详情改为禁止系统信息/运行状态重复区块回归。
- 已验证：`node scripts/verify-discovery-services-layout.mjs`、`npm run lint`、`npm run build:test`、`git diff --check` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/` 和 `http://127.0.0.1:8090/` 均返回 200。
- 已验证：使用 `admin/admin123` 真实 Chromium 打开服务详情编辑页，页面文本不包含 `系统信息 / 运行状态 / 实例健康 / 别名访问`，没有 Drawer/Dialog，浏览器 console warning/error 为 0；验收截图为 `output/playwright/service-detail-no-system-runtime.png`。

## 全局 Header 透明背景修复

- [x] 定位顶部透明来源：全局 Header 使用 sticky 但没有实体背景
- [x] 补齐 Header 不透明 surface 背景、文本色和 Fluent 边界色
- [x] 增加静态回归校验，防止 sticky Header 重新变透明
- [x] 运行校验、构建、重启并用真实浏览器验证服务详情页顶部不再透出内容
- [x] 补充 Review 和验证结果

当前判断：

- 服务详情页顶部透明不是详情卡片问题，而是 `layouts/components/Header/index.module.less` 的 sticky header 没有设置 background，内容滚到 header 下方时会透出。
- Header 属于全局壳层，应使用 `--app-surface` 作为不透明背景，并使用 `--app-border-subtle` 做底部分割线。

Review：

- 已完成：全局 Header `.panel` 增加 `background: var(--app-surface)`、`color: var(--app-text)`、`border-bottom: 1px solid var(--app-border-subtle)` 和 `box-sizing: border-box`，sticky 顶栏不再透明。
- 已完成：`verify-fluent-ui-migration.mjs` 增加 Header 不透明背景、Fluent 边界色和禁止透明背景的静态回归约束。
- 已验证：`node scripts/verify-fluent-ui-migration.mjs`、`npm run lint`、`npm run build:test`、`git diff --check` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/` 和 `http://127.0.0.1:8090/` 均返回 200。
- 已验证：使用 `admin/admin123` 真实 Chromium 打开服务详情编辑页，Header computed background 为 `rgb(255, 255, 255)`，`position=sticky`，底部分割线为 `1px`，浏览器 console warning/error 为空；验收截图为 `output/playwright/header-opaque-service-detail.png`。

## 服务详情页内编辑统一

- [x] 梳理服务列表、服务详情页、服务编辑弹窗和 Redux service 状态流
- [x] 明确交互：名称点击进入详情查看，操作列查看/编辑进入详情页编辑态，详情页内部切换查看/编辑
- [x] 拆出可复用服务表单，创建保留弹窗，已有服务编辑改为详情页内嵌表单
- [x] 增加静态回归校验，防止服务查看/编辑重新走 Drawer
- [x] 运行定向校验、lint、构建、重启并用 admin 浏览器验证服务详情页编辑链路
- [x] 补充 Review 和验证结果

当前判断：

- 当前服务列表的“查看/编辑”操作会打开 `ServiceEditor` 抽屉，而服务详情 Tab 使用另一套 `ServiceDetail` 展示布局，导致查看和编辑割裂。
- 目标交互应收敛为：服务名称链接进入详情查看；行操作进入同一服务详情页并打开编辑态；详情页内的基础信息、归属信息、标签使用同一套字段结构在查看和编辑之间切换。
- 创建服务仍是短流程，保留创建弹窗更合适；本次取消的是已有服务的弹窗查看/编辑形态。

Review：

- 已完成：新增 `ServiceForm`，服务创建弹窗和服务详情页内查看/编辑共用同一套基础信息、归属信息和服务标签表单；命名空间和名称只在创建态可编辑，已有服务编辑态保持只读。
- 已完成：`ServiceEditor` 收敛为“创建服务”弹窗容器，不再承载已有服务查看或编辑。
- 已完成：服务列表名称点击进入详情查看；操作列 `viewEdit` 进入 `/discovery/service/instance?...&mode=edit`，在详情页原地打开编辑态，不再打开 Drawer。
- 已完成：服务详情页头部提供“编辑 / 退出编辑”，提交后回到查看态并重新拉取详情；系统信息和运行状态仍保留在同一详情页内。
- 已完成：更新 `verify-discovery-services-layout.mjs`，约束已有服务查看/编辑不能回到 `ServiceEditor` Drawer，并校验页面内编辑态、`mode=edit` 路由和输入框宽度。
- 已验证：`node scripts/verify-discovery-services-layout.mjs`、`npm run lint`、`npm run build:test`、`git diff --check` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/` 和 `http://127.0.0.1:8090/` 均返回 200。
- 已验证：使用 `admin/admin123` 真实 Chromium 打开服务列表，操作列进入详情页编辑态且没有 Drawer/Dialog；名称链接进入详情查看态且没有编辑表单；浏览器 console warning/error 为空。验收截图为 `output/playwright/service-detail-edit-mode.png`。

## A2A Agent 列表固定列修复

- [x] 确认 A2A Agent 列表的名称列、操作列和 Fluent Table 封装当前行为
- [x] 补齐 fixed 列的真实 sticky 实现，让名称固定首列、操作固定尾列
- [x] 更新静态回归校验，覆盖列定义和封装实现
- [x] 运行校验、lint/构建，并用浏览器验证横向滚动下固定列仍可见
- [x] 补充 Review 和验证结果

当前判断：

- A2A Agent 列表已经在列定义上声明 `fixed: 'left'` 和 `fixed: 'right'`，但共享 Fluent Table 当前没有把该语义转成真实 sticky 列，导致页面滚动时首尾列不会固定。
- 修复应优先补齐共享表格封装的 fixed 语义，同时保留 A2A 页面的稳定最小宽度和横向滚动容器。

Review：

- 已完成：`components/Fluent/Table` 将 `column.fixed` 映射为 `fluent-table-fixed-left/right` 和 sticky `left/right: 0`，表头与内容单元格都会应用同一套固定列语义。
- 已完成：统一 Fluent 表格 fixed 列背景、层级和边界线，避免横向滚动时中间内容穿透到固定列下面。
- 已完成：A2A Agent 列表继续声明名称列 `fixed: 'left'`，操作列 `fixed: 'right'`，并由 `verify-a2a-agent-table-layout.mjs` 同时约束页面列定义和共享封装实现。
- 已验证：`node scripts/verify-a2a-agent-table-layout.mjs`、`npm run lint`、`npm run build:test`、`git diff --check` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/` 和 `http://127.0.0.1:8090/` 均返回 200。
- 已验证：使用 `admin/admin123` 在真实 Chromium 打开 `/ai/a2a`，表格横向滚动到最右后，首列左边缘和操作列右边缘仍分别贴住表格容器，浏览器 console warning/error 为空；验收截图为 `output/playwright/a2a-fixed-columns.png`。

## 复杂资源详情页路由化

- [x] 盘点治理工作台、MCP、A2A、主体管理、权限策略详情抽屉的现有入口和数据链路
- [x] 明确详情路由化规则：复杂资源查看走独立页面，轻量创建/授权/确认保留抽屉
- [x] 优先改造 MCP/A2A 详情为独立页面，保留工具/技能浏览能力和编辑入口
- [x] 改造认证主体和权限策略详情为独立页面，保留关联策略、成员、资源信息等详情 Tab
- [x] 设计治理规则详情页面承载方式，替换工作台详情抽屉入口
- [x] 增加静态回归校验，防止复杂详情继续从列表打开 Drawer
- [x] 运行 lint、构建、重启并用 admin 浏览器验证关键详情链路
- [x] 补充 Review、截图和运行态结果

当前判断：

- 详情内容复杂、需要多 Tab、工具/技能/版本/资源树/成员信息的资源，不适合继续塞在抽屉里；独立页面可以刷新、直达、复制链接，也和配置中心、服务实例的心智一致。
- 抽屉仍适合短流程：新建、编辑、授权、删除确认、简单表单；不适合承载治理规则详情、MCP 工具浏览、A2A Agent Card、主体权限信息和策略资源详情。
- 第一批要覆盖用户明确点名的治理工作台、MCP、A2A、认证管理、策略管理；实现时优先复用现有详情内容，避免同时重写业务逻辑。

Review：

- 已完成：治理工作台规则名称、行点击和查看操作不再打开详情抽屉，统一跳转 `/governance/rules/detail?kind=...&id=...`；独立详情页按 URL 重新拉取规则并复用现有 RuleTabs、版本、监听和九类规则 editor。
- 已完成：MCP Server 列表查看入口跳转 `/ai/mcps/detail`；详情页保留 Server 摘要、后端跳转、工具浏览、刷新、编辑和授权入口。
- 已完成：A2A Agent 列表查看、技能、Agent Card 入口跳转 `/ai/a2a/detail`；详情页保留 Agent Card、技能浏览和页面内编辑 Tab。
- 已完成：认证主体用户、用户组、角色和权限策略列表名称/查看操作统一跳转既有独立详情页，不再从列表打开查看态编辑抽屉。
- 已完成：新增 `verify-standalone-detail-pages.mjs` 并接入 `npm run test:standalone-details`；兼容旧 `npm run test:ai-detail-drawers` 指向新门禁，防止复杂详情回退成 Drawer。
- 已验证：`npm run test:standalone-details`、`npm run test:ai-detail-drawers`、`npm run test:resource-name-links`、`npm run lint`、`npm run build:test`、`git diff --check` 通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/` 和 `http://127.0.0.1:8090/` 均返回 200。
- 已验证：使用 `admin/admin123` 真实浏览器从列表点击进入 MCP、A2A、治理规则、主体管理、权限策略详情，URL 均为独立详情页，详情页可见核心内容且没有 Drawer/Dialog；console warning/error 收集为空。
- 验收截图：`output/playwright/standalone-governance-detail.png`。
- 说明：仓库当前没有可执行的 `context_kg_lint.py`，本轮未运行 context-kg lint。

## MCP 详情抽屉设计统一

- [x] 对比 MCP/A2A/服务/命名空间详情抽屉的信息架构和操作区样式
- [x] 重构 MCP 服务详情摘要、操作按钮、工具浏览区域为统一 Fluent 资源详情布局
- [x] 横向检查 A2A 详情是否复用同类问题并同步收敛
- [x] 增加静态回归校验，防止详情抽屉继续出现大卡片式操作按钮和分裂布局
- [x] 运行 lint、构建、重启并用 admin 真实浏览器验证 MCP 详情
- [x] 补充 Review、截图和运行态结果

当前判断：

- MCP 详情抽屉已经承载“查看详情 + 工具能力浏览 + 进入编辑”的主路径，不应再把刷新和编辑做成大卡片按钮；操作应回到标题/摘要区右侧的轻量按钮组。
- MCP 与 A2A 都属于 AI 资源详情，信息架构应和服务、命名空间抽屉统一为：身份摘要、关键元信息、分段内容、紧凑操作，不要每个页面独立发明卡片密度和按钮尺寸。
- 工具浏览更接近 API 文档/能力清单，应保持可扫描的目录式布局，但外层容器、标题、计数、空态和操作区需要和资源列表/详情的 Fluent 风格一致。

Review：

- 已完成：MCP 详情摘要区保持身份图标、名称、协议、描述和四项关键元信息，右侧操作收敛为 32px 刷新图标按钮 + 紧凑 `编辑 Server`，不再出现截图里的大卡片式操作块。
- 已完成：A2A Agent 详情同步收敛摘要操作区，防止 AI 资源详情继续出现两套按钮密度。
- 已完成：新增 `verify-ai-detail-drawer-layout.mjs` 并接入 `npm run test:ai-detail-drawers`，约束 MCP/A2A 详情抽屉宽度、摘要三段结构、紧凑操作组和 32px 图标按钮。
- 已验证：`npm run test:ai-detail-drawers`、`npm run test:resource-name-links`、`npm run lint`、`npm run build:test` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/`、`http://127.0.0.1:8090/` 和当前入口静态资源均返回 200。
- 已验证：使用 `admin/admin123` 真实浏览器打开 MCP `default/demo-mcp-2` 详情，刷新按钮实测 32x32、编辑按钮 127x32、工具浏览正常显示 3 个工具；浏览器 console warning/error 为 0，MCP/A2A 业务请求均返回 200。
- 已验证：A2A `default/order-planner-agent` 详情刷新按钮实测 32x32，同类样式没有回退；验收截图为 `output/playwright/mcp-detail-drawer-unified-final.png`。

## 命名空间抽屉查看编辑布局统一

- [x] 对比服务详情/编辑抽屉的布局、宽度、页脚和只读态模式
- [x] 调整命名空间详情抽屉宽度，移除详情页底部授权按钮
- [x] 让命名空间查看态和编辑态复用同一表单式布局
- [x] 运行静态检查、lint、构建、重启并用 admin 真实浏览器验证
- [x] 补充 Review、截图和运行态结果

当前判断：

- 命名空间授权入口已经在列表操作列中存在，详情抽屉底部不应再放“授权”主操作，避免查看态承担权限变更入口。
- 命名空间详情当前像信息展示页，编辑态像表单页；需要参考服务抽屉，将查看态渲染成同一套字段布局，仅控件进入只读展示。
- 命名空间信息字段少，抽屉宽度不应沿用过宽详情面板，应该收敛到更紧凑的中等抽屉宽度。

Review：

- 已完成：命名空间详情抽屉宽度收敛为 `min(720px, 94vw)`，底部只保留 `编辑 / 关闭`，不再重复展示“授权”；列表操作列的授权入口保持不变。
- 已完成：`NamespaceEditor` 移除详情展示页和编辑表单两套结构，查看、编辑、创建统一走同一套 `namespaceForm` 分段布局；查看态使用无边框正常文本渲染，编辑态原地放开描述和标签编辑。
- 已完成：命名空间名称沿用既有规则，创建态可编辑，编辑态和查看态只读展示，避免修改不可变资源标识。
- 已验证：`npm run test:namespace-drawer`、`npm run test:resource-name-links`、`npm run lint`、`npm run build:test` 均通过。
- 已验证：重新拉起 all-mode 后日志出现 `finish starting server`，`http://127.0.0.1:8080/`、`http://127.0.0.1:8090/` 和当前入口静态资源均返回 200。
- 已验证：使用 `admin/admin123` 真实浏览器打开命名空间详情，抽屉实测宽度 720px，详情按钮只有 `编辑 / 关闭`；点击编辑后仍为同一套 `基础信息 / 命名空间标签` 布局，按钮切换为 `提交 / 重置`，无“授权”按钮。
- 已验证：浏览器 console warning/error 为 0，命名空间页面业务请求均返回 200；验收截图为 `output/playwright/namespace-drawer-unified-final.png`。

## 全站资源列表名称列统一

- [x] 盘点所有资源列表的名称列、附加内容和现有查看入口
- [x] 建立共享资源名称链接组件与可访问性约束
- [x] 将可见资源列表名称列收敛为“仅名称 + 查看链接”
- [x] 修复没有现成名称查看入口的列表跳转或抽屉交互
- [x] 增加静态门禁并运行 lint、构建、重启和真实浏览器逐页验证
- [x] 补充 Review、验收截图和最终运行态结果

当前规则：

- “名称”列只展示资源名称，不在同一单元格附带描述、类型、状态、数量、归属或其它摘要。
- 资源名称统一呈现为可访问的链接式控件；点击复用该资源已有的查看页或查看抽屉，不直接进入编辑态。
- 描述等仍有列表扫描价值的信息保留为独立列；没有独立列且非必要的信息从列表移除，到详情中查看。

Review：

- 已完成：新增共享 `ResourceNameLink`，命名空间、服务、配置分组、治理工作台、治理发布版本、A2A Agent、MCP Server、服务别名、服务订阅和权限策略资源清单统一为“名称列仅资源名称 + 查看链接”。
- 已完成：缺少查看入口的资源补齐详情动作；服务别名进入只读详情后可切换编辑，治理发布版本点击名称打开版本详情，权限策略资源名点击打开资源引用详情。
- 已完成：修复策略资源清单中 `ResourceNameLink` 被表格固定布局压成 0 宽的问题；资源名称列现在有稳定列宽，真实点击 `全部（包括新增）` 可打开资源详情。
- 已验证：`npm run test:resource-name-links`、`npm run lint`、`npm run build:test`、`context_kg_lint.py context-kg`、`git diff --check` 均通过。
- 已验证：all-mode 日志出现 `finish starting server`，`http://127.0.0.1:8080/`、`http://127.0.0.1:8090/` 和当前入口资源均返回 200。
- 已验证：使用 `admin/admin123` 真实浏览器点击命名空间、服务、配置分组、治理工作台、治理版本、A2A、MCP、权限策略资源引用；名称列均只显示名称，点击进入对应查看页或查看抽屉。
- 已验证：浏览器 console warning/error 为 0，业务请求均返回 200；验收截图为 `output/playwright/resource-name-links-policy-detail.png`。

## 配置中心布局与治理规则 Fluent 抽屉统一

- [x] 复现配置文件详情页订阅查询的宽度压缩和空状态错位
- [x] 修复配置文件详情主内容、Tab 内容和订阅表格的自适应布局
- [x] 盘点九类治理规则详情抽屉的共享容器、标题操作、Tab 与滚动边界
- [x] 在 `RuleDetailDrawer` 共享边界统一 Fluent UI 视觉和交互规范
- [x] 增加静态门禁并运行 lint、构建、重启和真实浏览器逐类验证
- [x] 补充 Review、验收截图和最终运行态结果

当前判断：

- 配置订阅查询同时渲染空版本索引和主表，外层 `subscribeShell` 的 flex 布局把空索引压成窄列并挤占表格宽度；订阅列表应以主表为主，版本筛选只在存在可用版本时作为紧凑筛选能力出现。
- 治理规则已经使用 Fluent `OverlayDrawer`，但详情样式仍主要覆盖旧 TDesign `.t-drawer__*` 与 `.t-tabs__*` class，导致标题区、关闭按钮、Tab、正文表面和操作组没有真正进入 Fluent 设计语言。
- 本轮应通过共享抽屉与共享 RuleTabs 横向覆盖路由、限流、熔断、主动探测、无损、泳道、调用鉴权、流量镜像和流量 Mock，不在九类编辑器里分别打补丁。

Review：

- 配置订阅查询移除无数据的版本索引栏，Fluent Tab 内容、订阅容器和表格均以容器宽度自适应；真实页面实测 `clientWidth === scrollWidth`，版本栏数量为 0，空状态不再被挤成竖排。
- `RuleDetailDrawer` 统一 Fluent 标题、类型标签、关闭按钮、正文表面、Tab 和滚动边界；编辑、发布、保存、撤销通过共享 action host 固定在标题区，旧 `StickyTool` 仅作为兼容挂载点且不可见。
- 发布抽屉和泳道内部 Tab 不再覆盖 `.t-drawer__*`、`.t-tabs__*`；路由、泳道、安全类编辑器统一按共享抽屉正文高度计算，限流直接入口与工作台保持同一宽型尺寸。
- 九类规则使用 `admin/admin123` 逐条真实打开：复杂规则宽 864px，轻量规则宽 749px；九类均存在“编辑/发布”，进入编辑后均切换为“保存/撤销”，可见旧悬浮操作数为 0，正文实际滚动、版本/监听或泳道版本/审计 Tab 均通过。
- 巡检额外修复熔断详情把缺失 `editable` 错当成拒绝的问题；现在只有接口明确返回 `editable=false` 才隐藏操作，admin 和默认可维护资源保持可编辑。
- `node scripts/verify-fluent-governance-drawer-layout.mjs`、`npm run lint`、`npm run build:test` 通过；all-mode 日志出现 `finish starting server`，8080、8090 和治理工作台均返回 200。
- 验收证据：`output/playwright/config-subscribe-layout-fluent-final.png`、`governance-drawer-fluent-route.png`、`governance-drawer-fluent-security.png` 和 `governance-fluent-drawer-audit.js`。

## Console 全页面交互与跳转巡检

- [x] 生成全部可见菜单、隐藏详情路由和页面内主要交互清单
- [x] 使用 admin 真实登录逐页点击菜单、查看、创建、编辑、授权、Tab、返回和详情链接
- [x] 同步记录 URL、抽屉/弹窗状态、请求失败及浏览器 warning/error
- [x] 修复巡检发现的路由、状态传递和交互问题并增加回归门禁
- [x] 重建 all-mode，完整复跑并记录 Review

巡检范围：

- 可见入口：命名空间、A2A Agent、MCP 服务、服务实例、配置分组、治理工作台、事件指标、操作审计、主体管理、权限策略。
- 隐藏链路：服务详情/实例、配置分组文件、用户/用户组/角色详情、策略详情、治理各类型规则详情。
- 交互类型：菜单跳转、资源名称链接、查看/编辑、创建、授权、删除确认但不执行删除、Tab、分页、返回、新标签用户链接。

发现与修复：

- 用户 Token 接口经过统一响应解包后直接返回 `User`，页面仍读取 `res.user.auth_token` 导致崩溃；用户和用户组 Token service 现同时兼容直接对象与旧包裹对象。
- 权限策略创建页使用 `Radio.Button`，Fluent 适配层迁移时只保留了 `Radio.Group`；现补齐 `Radio.Button` 静态成员，创建抽屉恢复。
- 事件指标和操作审计在观测插件使用日志模式、数据库表尚未创建时返回 MySQL 1146；Console MySQL observer 现把缺表识别为空数据，其它数据库错误仍正常上抛。
- 当前 specification 的 `LosslessRule` 不包含命名空间、服务和描述字段，内部模型信息在 API 输出时丢失，工作台显示 `undefined/undefined`；控制面和 Console 现通过保留 metadata 键完成上下文往返，同时从用户标签中隐藏这些内部键。

Review：

- 可见菜单：命名空间、A2A Agent、MCP 服务、服务实例、配置分组、治理工作台、事件指标、操作审计、主体管理、权限策略，真实侧边栏点击 10/10 到达预期 URL。
- 资源操作：创建、查看、详情转编辑、授权、技能、Token、绑定服务、配置文件、关联策略与删除确认均已点击；删除仅验证确认浮层，未提交破坏性操作。
- 隐藏链路：服务详情 5 个 Tab、配置文件创建、用户权限与关联策略、默认策略 4 个 Tab、治理工作台 9 种规则详情全部通过。
- 修复后定向复验：两个指标页、无损规则名称、用户 Token、权限策略创建均通过，warning/error、pageerror 和 HTTP 4xx/5xx 为 0。
- 自动化证据：`output/playwright/full-navigation-audit.js`、`focused-interaction-fix-audit.js`、`hidden-interaction-audit.js` 和详情转编辑巡检脚本。
- 代码门禁：新增 `verify-console-interaction-contracts.mjs`、无损上下文往返测试和 observer 缺表识别测试。

## 资源表格列对齐与自适应规范

- [x] 核对 Fluent Table 默认布局、页面显式 fixed/auto 和列对齐来源
- [x] 共享表格默认改为按容器缩放的列宽、左对齐和长内容省略
- [x] 命名空间与服务列表移除操作列居中覆盖并接入自适应权重
- [x] 增加静态门禁，验证表头与内容起始位置、列宽和省略行为
- [x] 运行 lint、构建、重启及真实浏览器复验并记录 Review

当前判断：

- Fluent Table 的数字列宽原本直接作为像素宽度使用，窄容器下各列宽度相加会把表格撑出容器；自适应布局需要把数字宽度解释为列权重并转换为百分比。
- 表头与内容虽然读取同一列配置，但操作列显式 `align: center` 且按钮组使用 `margin: 0 auto`，破坏了全列左侧基线。
- 长文本省略目前只在少量页面私有 class 中实现，共享表格没有稳定的单行溢出边界。

Review：

- 已完成：共享 Fluent Table 对数字列宽计算总权重并转换为百分比，配合 fixed 算法让表格始终填满且不超出容器；显式 `auto` 的特殊内容表格仍保留内容驱动行为。
- 已完成：共享表头和单元格统一左对齐，增加表头与内容溢出容器；普通文本默认单行省略。
- 已完成：命名空间与服务列表移除操作列 `align: center` 和按钮组自动居中，操作按钮保持紧凑并统一靠左。
- 已完成：命名空间操作时间、服务列表时间与归属等自定义多行内容补齐内部省略规则。
- 已完成：新增 `verify-fluent-table-layout.mjs`，固定百分比列权重、左对齐、容器宽度和省略规则。
- 已验证：`verify-fluent-table-layout.mjs`、`verify-operation-button-icons.mjs`、`npm run lint`、`npm run build:test` 通过。
- 已验证：all-mode 输出 `finish starting server`，8080/8090 返回 200；1024px 实测命名空间表格与容器均为 739px、无横向溢出，表头/首行逐列起点一致，所有列左对齐，操作时间实际触发 ellipsis；1440px 服务列表表格与容器均为 1155px且逐列对齐；浏览器 warning/error 为 0。
- 验收截图：`output/playwright/namespace-table-left-adaptive-final.png`。

## 资源表格操作列紧凑化

- [x] 在真实命名空间页面测量操作列、按钮和按钮组的实际尺寸
- [x] 在共享 OperationButton 边界统一图标按钮尺寸、间距和对齐
- [x] 命名空间与服务列表接入共享操作按钮组
- [x] 增加静态验证并运行 lint、构建和真实浏览器复验
- [x] 记录 Review 与本次用户纠正经验

当前判断：

- Fluent 迁移后共享操作按钮仍使用中号默认尺寸，但命名空间页面只保留了旧 `.t-button` 的 28px 覆盖规则，该规则对 `.fui-Button` 不生效。
- 命名空间操作列声明宽度为 108px，浏览器实际分配约 155px；操作区必须自身保持 `max-content` 紧凑宽度，不能依赖表格列宽决定图标间距。

Review：

- 已完成：共享 `OperationButton` 默认使用 small 尺寸，并统一为 28×28px；Tooltip 包装节点同步固定为 28px，避免包装层参与拉伸。
- 已完成：新增 `OperationButtonGroup`，使用 `width: max-content`、4px gap 和不换行布局；命名空间与服务列表的操作列已接入。
- 已完成：删除两个资源页中仅适配旧 TDesign `.t-button` 的失效规则，页面仅保留按钮组居中职责。
- 已完成：`verify-operation-button-icons.mjs` 增加共享尺寸、间距、内容宽度和资源页接入门禁。
- 已验证：`node scripts/verify-operation-button-icons.mjs`、`npm run lint`、`npm run build:test` 通过。
- 已验证：all-mode 日志出现 `finish starting server`，8080/8090 均返回 200；真实命名空间页三个按钮组为 92×28px，按钮横向间隔 32px，真实服务列表两个按钮组为 60×28px，浏览器 warning/error 为 0。
- 验收截图：`output/playwright/namespace-operation-column-compact.png`。

## A2A Agent 列表表格排版细节优化

- [x] 对照截图定位表格头部、筛选区、列宽、标签、技能数、操作列和分页的排版问题
- [x] 将 A2A Agent 表格调整为稳定列宽、紧凑标签、单行技能数、右侧固定操作和可控横向溢出
- [x] 优化 A2A 筛选区在宽屏/中屏下的网格节奏，避免控件过宽、掉行或和表格边界不齐
- [x] 增加静态验证脚本，固定本次表格排版约束
- [x] 将本次用户纠正沉淀到 lessons
- [x] 运行前端 lint、相关 verify、构建或可用的页面验证，并记录 review

Review：

- 已完成：A2A Agent 表格改为 `tableLayout="fixed"`，并通过 `agentTable`/`tableSurface` 固定表格最小宽度、表头 nowrap、单元格 padding、分页边界和最后操作列垂直居中。
- 已完成：身份列、接入列、归属列、能力列、技能数、来源、最近修改和操作列重新分配稳定宽度；Agent 名称、后端地址、归属说明和时间字段均使用单行省略。
- 已完成：能力标签改成单行紧凑列表，`Extended Card` 在列表中缩短为 `Extended`；技能数使用 `skillCountLink` 保证 `2 个` 不拆行；操作列使用 `actionCell` 控制图标间距和右对齐。
- 已完成：筛选区改成确定性 grid 宽度，宽屏右对齐，中屏按 4 列/2 列换行。
- 已完成：新增 `console/web/scripts/verify-a2a-agent-table-layout.mjs`，固定 fixed table、列宽、单行技能数、能力/来源标签不换行、操作列居中和禁止旧 `.t-table` 表格选择器回退。
- 已完成：更新 `context-kg/tasks/lessons.md`，记录资源列表交付必须检查表头换行、标签撑高、计数字段拆行、分页边界和操作图标间距。
- 已验证：`node scripts/verify-a2a-agent-table-layout.mjs` 通过。
- 已验证：`npm run lint` 通过。
- 已验证：`npm run build:test` 通过。
- 已验证：Playwright 使用样例 A2A 数据打开 `http://127.0.0.1:4175/ai/a2a`，表头高度 45px，首行高度 107px，技能数链接高度 20px，能力标签高度 20/32px，操作列中心偏移 0，表格 `clientWidth=scrollWidth=1765`，浏览器 warning/error 为 0；截图见 `output/playwright/a2a-agent-table-layout-final.png`。

## Fluent 侧边栏视觉纠偏

- [x] 对照用户截图检查 Fluent Nav 默认表面和层级问题
- [x] 收敛一级分组、二级菜单、选中态和 Footer 的视觉规范
- [x] 验证展开态、折叠态、窄屏布局及所有菜单点击
- [x] 运行前端校验、重建 all-mode 并复验 8080

当前判断：

- 当前每个 NavItem、NavCategoryItem 和 NavSubItem 都保留 Fluent 默认灰色 surface，连续排列后形成大块圆角卡片，破坏了侧边栏的信息层级。
- 侧边栏应保持白色连续导航面，一级模块强调图标和标题，二级菜单只做缩进；仅当前资源显示浅蓝背景与左侧品牌色标记。
- 本轮不改变 Logo、232px 宽度、模块顺序、全部展开和折叠行为。

修复：

- 普通一级、分组标题和二级菜单统一为透明连续表面，移除每行常驻灰底和卡片感。
- 一级模块使用图标与中等字重，二级菜单使用 44px 缩进和 36px 紧凑行高；hover 仅显示轻灰反馈。
- 当前资源使用 `#ebf3fc` 浅蓝背景、品牌色文字和 3px 左侧指示条，并移除 Fluent 默认重复选中标记。
- 折叠态固定 64px，图标居中、隐藏文字和分组箭头；Footer 保留版本号且不参与滚动。
- Toast 增加 Error/对象内容归一化，避免菜单横向验收遇到接口错误时把 Error 对象作为 React 子节点导致页面崩溃。
- FluentProvider 外增加独立 `fluent-provider-shell`，补齐应用根节点满高链路；`sidePanel` 固定内部滚动，侧边栏和内容区均占满视口，Footer 锚定到底部。

验证：

- `npm run lint`、`verify-fluent-ui-migration.mjs`、`verify-brand-logo.mjs`、`verify-production-react-build.mjs` 和生产构建通过。
- Playwright 验证 MCP、服务实例、配置分组、治理工作台、事件指标和主体管理菜单均可跳转。
- 1440x900 展开态、64px 折叠态和 390x844 窄屏截图检查通过，侧栏没有灰色卡片或残留展开箭头。
- 8080 生产环境 1200x1600 长视口实测侧栏 `top=0`、`bottom=1600`、`height=1600`，Footer 位于 `1563–1600`，根页面 `scrollHeight=1600`。
- all-mode 重建后出现 `finish starting server`；8080 使用 `admin/admin123` 登录复验，console error/warning 为 0；8080、8090 均返回 200。

Review：

- 本次只纠正侧边栏视觉和 Toast 错误边界，没有改变路由结构、模块顺序、默认展开、权限或业务页面。
- 侧边栏视觉约束已加入 Fluent 迁移静态检查，防止后续恢复每行灰色 surface 或折叠态箭头。

## Console 全站迁移 Fluent UI v9

- [x] 盘点 TDesign 组件、图标、样式和复杂表单使用面
- [x] 安装 Fluent UI v9 并建立品牌主题、Provider 与全局 token
- [x] 建立 Fluent 组件适配边界，迁移共享按钮、输入、选择、提示、通知和图标
- [x] 迁移应用壳层、侧边导航和资源页共享布局
- [x] 迁移配置、治理、认证、AI 与监控页面的可见控件
- [x] 增加 TDesign 直接 import 禁止检查并将旧依赖限制在适配层
- [x] 运行 lint/build/verify、重启和真实浏览器全站抽检

当前判断：

- 当前约 102 个源码文件直接导入 TDesign，包含 60 余种组件/类型和 457 处表单相关调用；这不是颜色变量调整，而是组件系统迁移。
- 本轮按 [[adr-console-fluent-ui-design-system]] 使用 Fluent UI React v9。通过仓库级适配层保持现有业务流程与表单字段契约，优先统一所有用户可见控件和页面语言。
- 企业软件风格固定为 Fluent 的蓝、灰、白体系，强调信息密度、清晰层级和键盘焦点，不引入营销页式装饰。

修复：

- 引入 Fluent UI React v9 与 Fluent Icons，增加品牌 light/dark theme、根 Provider、Toast Host 和蓝灰白全局 token。
- `components/Fluent` 提供按钮、输入、选择、表格、分页、抽屉、对话框、Tab、Tooltip、Popconfirm、通知、布局和表单控件适配；页面源码不再直接 import TDesign。
- 侧边栏改用 Fluent Nav，并统一展开态、折叠态、图标、品牌 Logo、页头和资源页控件；桌面与移动端使用同一设计语言。
- 修复迁移中发现的 Layout 静态子组件缺失、旧 placement 与 Fluent positioning 枚举不兼容、表单受控状态切换和 Portal Provider 拦截点击问题。
- 移除按 React/TDesign 内部目录硬拆 chunk 的旧构建规则，避免 React、Griffel、Fluent 和兼容层形成运行时循环；增加静态门禁和构建安全检查。
- 旧 Form 状态控制器及少数高复杂度控件只允许存在于适配层，继续保持嵌套字段、校验和现有业务提交契约；页面层已与具体旧组件库解耦。

验证：

- `npm run lint`、`npm run build` 和 `scripts/verify-*.mjs` 全部通过，生产构建无 warning。
- `verify-fluent-ui-migration.mjs` 确认页面、布局及普通共享组件不存在 TDesign/TDesign Icons 直接 import。
- Playwright 使用 `admin/admin123` 验证登录、命名空间、配置分组、治理工作台、服务列表/详情、MCP 和主体管理；下拉、抽屉与 Portal 交互正常。
- 1440x900 与 390x844 视口均完成截图检查；浏览器 console error/warning 为 `0`。
- all-mode 重建日志出现 `finish starting server`，`http://127.0.0.1:8080` 与 `http://127.0.0.1:8090` 均返回 `200`。

Review：

- 迁移边界集中在 `components/Fluent`，业务页面保留原有字段、权限、查询和提交逻辑；后续新增页面只能从该边界或 Fluent v9 直接使用统一控件。
- TDesign 依赖暂不删除，原因是复杂 Form 状态控制器和树/穿梭框等高风险交互仍需兼容；它不能再从页面直接引用，也不能决定全局视觉语言。

## Console 实例 TCP/HTTP 主动健康检查

- [x] 核对现有 Console、specification、健康检查插件、调度器和 MySQL 映射
- [x] 扩展 `HealthCheck` 契约并生成 Go 代码
- [x] 实现并注册 TCP/HTTP 主动探测插件
- [x] 适配实例创建、调度和 MySQL round-trip
- [x] 调整 Console：创建态禁用心跳，默认 TCP，开放 TCP/HTTP
- [x] 补充契约、后端、存储和前端回归测试
- [x] 重建 all-mode 并完成真实页面、请求和探测验证

当前判断：

- SDK 注册实例可以通过心跳维持健康状态，Console 手工创建实例没有心跳上报方，因此创建页面不能允许选择心跳。
- 当前后端 `service.proto`、实例转换和默认插件都只实现 HEARTBEAT；TCP/HTTP 是禁用占位项，不能仅修改前端选项状态。
- 本轮按 [[adr-instance-active-healthcheck]] 实现控制面 TCP/HTTP 实例主动探测；不改变治理规则 `FaultDetectRule` 的客户端探测模型。

修复：

- specification 的 `HealthCheck` 增加 TCP/HTTP 类型与探测间隔，HTTP 增加路径；同步生成 Go 代码并更新 Rust 协议源。
- 服务端新增 TCP 和 HTTP 真实探测插件，调度器按检查类型与间隔执行；当所有 checker 自身尚未健康时，使用非隔离 checker 兜底，避免主动检查无节点执行。
- MySQL 继续复用 `health_check.ttl` 保存心跳 TTL 或主动探测间隔，HTTP 路径保存到保留 metadata；读取时先解析 metadata 再构造类型化健康检查，确保路径不会退回默认 `/`。
- Console 创建实例时禁用心跳，默认 TCP，开放 TCP/HTTP；编辑已有心跳实例仍可正确回显。请求边界将 camelCase 表单数据转换为后端 snake_case。

验证：

- `go test ./...` 通过；TCP/HTTP 插件、实例协议转换、调度器 checker 兜底和 MySQL metadata 恢复均有回归测试。
- `cd console/web && node scripts/verify-discovery-services-layout.mjs && npm run lint && npm run build:test` 通过。
- specification 执行 `go test ./...` 通过；Rust 使用仓库内 protoc 执行 `cargo check` 通过。
- all-mode 重建后日志出现 `finish starting server`，8080 和 8090 均返回 `200`。
- 真实创建 HTTP 主动检查实例指向临时 HTTP 服务：服务运行时实例转为健康，停止服务后实例转为不健康；两个测试实例均已删除，查询结果为 0 条。
- Playwright 验证创建态心跳 option 为 `t-is-disabled`，TCP/HTTP 可选；HTTP 提交 payload 为 `enable_health_check=true`、`health_check.type=3`、`http.interval=5`、`http.path=/ready`，浏览器 warning/error 为 0。截图见 `output/playwright/instance-healthcheck-create-options-final.png`。

Review：

- 本轮不是只开放前端选项，协议、插件、调度、存储、缓存恢复、Console 请求和真实健康状态切换已形成完整闭环。
- API/SDK 仍可使用心跳实例；限制只作用于 Console 手工创建态，不破坏已有心跳注册模型。

## 创建服务实例隐藏固定资源字段

- [x] 确认实例创建表单的命名空间、服务来源及提交依赖
- [x] 从创建/编辑表单移除固定的命名空间和服务输入框
- [x] 让创建请求直接使用当前服务上下文并补充静态回归检查
- [x] 运行前端校验、构建、重启和真实浏览器验证
- [x] 将运行信息字段改为单列纵向排列并完成真实页面复验

当前判断：

- 服务实例只能在当前服务详情上下文创建，命名空间和服务不是本次操作可选择的参数，不应以只读输入框伪装成表单字段。
- 顶部实例身份摘要继续展示当前命名空间和服务，保证用户知道操作对象；运行信息只保留实例自身可填写或可调整的主机、端口、协议、版本和权重。
- 请求中的 `namespace/service` 应直接来自当前服务上下文或已有实例，不再依赖不可见的 Form 字段。

修复：

- `InstanceEditor` 删除命名空间、服务两个只读 `FormItem`，同步移除对应的 `useWatch` 和 `setFieldsValue` 初始化字段。
- 新建实例直接使用详情页传入的 `namespace/service`，编辑实例优先使用已有实例归属；顶部 `InstanceSummary` 继续显示当前资源上下文。
- `verify-discovery-services-layout.mjs` 增加实例创建边界检查，防止固定资源上下文再次退化为只读表单框，并固定请求归属来源。
- 移除实例提交路径遗留的调试 `console.log`。
- 为运行信息增加独立 `instanceRuntimeGrid` 单列样式，主机、端口、协议、版本、权重纵向排列；通用表单网格仍保持双栏，健康检查和位置分段不受影响。

验证：

- `cd console/web && npm run lint && node scripts/verify-discovery-services-layout.mjs && npm run build:test` 通过。
- all-mode 重建后日志出现 `finish starting server`。
- Playwright 使用 `admin/admin123` 打开 `spec-governance/spec-checkout` 服务实例创建抽屉：表单标签为主机、端口、协议、版本、权重、健康与位置等字段，命名空间和服务表单项均不存在；顶部上下文仍正常显示。
- 浏览器拦截 POST 而未实际创建实例，payload 中 `namespace=spec-governance`、`service=spec-checkout`、`host=127.0.0.254`、`port=18080` 均正确；浏览器错误和警告为 `0`。
- 截图见 `output/playwright/instance-create-without-resource-fields.png`。
- Playwright 复验纵向布局：五个运行信息字段均为 `x=552px`、宽度 `696px`，按 `70px` 的纵向节奏排列；截图见 `output/playwright/instance-create-runtime-vertical.png`。

Review：

- 资源归属由服务详情上下文负责，实例表单只采集实例自身属性；UI 展示与请求数据责任保持一致。
- 本次未修改实例 API 契约、权限、列表、查看和删除逻辑，也没有向真实环境写入测试实例。

## 治理规则基础信息统一折叠

- [x] 盘点路由、限流、熔断、探测、无损、泳道和流量治理编辑器的基础信息结构
- [x] 新增共享基础信息折叠组件和紧凑摘要
- [x] 横向接入全部治理规则编辑器，默认展开并保持表单状态
- [x] 补充静态回归检查和 lessons
- [x] 运行前端校验、构建、重启和真实浏览器验证

当前判断：

- 基础信息折叠是治理规则编辑器的共性交互，不应只在当前截图对应的 RouteRule 局部实现。
- 折叠态应保留规则名称、状态或优先级摘要，并采用紧凑单行高度；展开态继续使用现有字段布局和表单控件。
- 本轮只改变编辑器信息分段的展开/折叠，不改规则模型、保存 payload、校验、发布和权限逻辑。

修复：

- 新增共享 `CollapsibleSection`，统一展开图标、`aria-expanded`、折叠摘要和正文显隐；正文使用 `hidden` 保持组件挂载，避免切换时清空尚未保存的表单状态。
- 路由、限流、熔断、主动探测、无损、泳道以及鉴权/镜像/Mock 编辑器全部接入，默认展开，重新打开编辑器时恢复展开态。
- 折叠摘要保留当前规则名称、启停状态和优先级；路由与泳道沿用已有编号式分段标题，其它编辑器沿用共享治理分段样式。
- 删除折叠外层的 `overflow: hidden` 和主动 `scrollIntoView`。正文已经由 `hidden` 收起，额外 overflow 会把 section 自身识别成滚动容器并将标题滚出卡片，落到粘性 Tab 下方导致无法点击。
- 新增 `verify-governance-basic-collapse.mjs`，固定共享组件的可访问状态、表单保留方式、统一图标和全部编辑器接入范围。

验证：

- `cd console/web && npm run lint && for script in scripts/verify-*.mjs; do node "$script"; done && npm run build:test` 通过。
- all-mode 重建后日志出现 `finish starting server`；8080、8090 和服务详情入口均返回 `200`。
- Playwright 使用 `admin/admin123` 从服务详情进入流量治理并新建 RouteRule：折叠后标题 `top=183px`、高度 `57px`，位于粘性 Tab 下方且可用普通点击重新展开。
- 折叠前输入规则名 `collapse-state-check`，重新展开后值保持不变，`aria-expanded=true`；折叠态截图为 `output/playwright/governance-basic-info-collapsed.png`。

Review：

- 本次没有修改规则数据模型、API 请求和保存行为；共享组件只控制信息分段显隐，已有表单校验及发布链路保持不变。
- 治理规则编辑器现在使用同一折叠交互，后续新增规则类型应继续复用共享组件，避免出现局部标题高度、滚动和图标行为不一致。

## 服务详情页信息排版修正

- [x] 对照 `service-alias.html` 原型确认服务详情页的信息架构与交互边界
- [x] 将真实 Console 服务详情改为身份头部、四项摘要、基础信息和运行状态四层结构
- [x] 接通复制服务 ID、查看实例、管理别名三个操作，并动态查询服务别名数量
- [x] 按宽屏双栏、摘要四列和窄屏单列补齐响应式布局
- [x] 更新静态回归检查和 lessons，防止原型与真实源码再次脱节
- [x] 运行前端校验、构建、重启和真实浏览器验证

当前判断：

- `service-alias.html` 是设计原型且已经符合本轮说明，真实落点是 `/discovery/service/instance` 下的 `ServiceDetail.tsx`。
- 当前真实页面仍是顶部命名空间标签、三项统计和单列字段分段，缺少跨 Tab 操作、别名总数和运行状态，因此不能把原型完成误判为源码完成。
- 本轮只调整服务详情 Tab 的阅读态布局和父级 Tab 切换回调；复用现有服务、别名查询接口，不改实例/别名/订阅/治理 Tab 的数据管理和后端契约。

修复：

- 服务详情已按身份头部、四项摘要、基础信息和运行状态重组；移除头部命名空间标签，基础信息保留 ID、命名空间、名称、Revision、端口、业务、部门、描述和时间。
- “复制 ID”复用剪贴板工具并支持页面指定反馈文案；“查看实例”“管理别名”和运行状态“查看”通过父级受控 Tab 切换，不改原有 Tab 数据组件。
- 服务别名数量通过现有 `DescribeServiceAliases` 接口按当前 namespace/service 动态查询；真实环境当前返回 `0`，页面不使用原型中的固定示例值。
- 无实例时后端端口可能返回字符串 `[]`，详情页统一归一化为“未注册实例，暂无端口”。
- 摘要宽屏四列、详情区宽屏双栏、基础信息内部双列；960px 以下详情区单列，640px 以下摘要和键值区单列。
- 修正 `AppLayout` 小屏折叠菜单后仍保留 `min-width: 760px` 的冲突，900px 以下 side/top/mix 内容区允许收缩，避免页面整体横向滚动。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs && npm run lint && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过；相关文件 `git diff --check` 通过。
- all-mode 重建后日志出现 `finish starting server`；8080、8090 和当前入口资源 `/assets/index.c8ca2fdf.js` 均返回 `200`。
- Playwright 在 1440px 下确认摘要为四个等宽 `272.5px` 列，详情区为 `652.5px / 453.5px` 双栏，无横向溢出。
- Playwright 在 640px 下确认摘要、详情区和键值项均为单列，`scrollWidth = clientWidth = 640`；截图见 `output/playwright/service-detail-mobile-final.png`。
- 复制 ID 显示“已复制 服务 ID”；管理别名和查看实例均切换到对应 Tab 并显示轻量反馈；浏览器 console error/warning 为 `0`。

Review：

- 原型和真实 Console 已使用同一信息架构；运行数据继续以真实服务和别名接口为准，不硬编码示例中的实例数、别名数或标签数。
- 页面没有新增调用拓扑或治理入口模块；现有独立“流量治理”Tab、服务别名 CRUD、实例和订阅逻辑保持原实现。

## 服务查看态复用编辑布局

- [x] 复核服务列表行操作里的 `ServiceEditor` 查看态和编辑态布局差异
- [x] 将服务查看态改为复用编辑表单分段布局，字段渲染为无边框只读文本
- [x] 在查看态保留“编辑”按钮，点击后原地放开表单编辑并显示提交/重置
- [x] 增加静态回归检查和 lessons，避免再次维护独立服务查看布局
- [x] 运行前端校验、构建、context-kg 校验和真实浏览器验证

当前判断：

- 用户当前反馈指向服务行操作里的查看/编辑抽屉：查看态和编辑态视觉结构不应是两套页面。
- 本轮不改变服务名点击进入服务详情/实例页的路由，也不改服务创建/更新接口；只收敛 `ServiceEditor` 内部查看态和编辑态的布局语言。
- 正确模型是同一套 `基础信息 / 归属信息 / 服务标签` 表单布局：查看态右侧字段用无边框只读文本正常展示，点击“编辑”后原地切换为可编辑控件。

修复：

- 删除 `ServiceEditor` 里的独立 `serviceView` 查看分支，查看和编辑统一渲染 `serviceForm`。
- 查看态用 `ReadonlyField` 展示命名空间、名称、描述、部门和业务，不再渲染灰底 disabled `Input/Select`；抽屉底部只展示 `编辑 / 关闭`，点击编辑后同一表单原地切换为可编辑控件，并展示 `提交 / 重置`。
- 服务查看和编辑抽屉统一使用 `min(720px, 94vw)` 宽度，避免查看态和编辑态切换时布局跳变。
- 删除未使用的 `serviceView/viewGrid/labelList` 等独立查看样式。
- 更新 `verify-discovery-services-layout.mjs`，固定服务查看态不得再维护独立展示分支，必须复用编辑表单，且阅读态字段不能通过 disabled `Input/Select` 表现。

验证：

- `cd console/web && node scripts/verify-discovery-services-layout.mjs && npm run lint` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 8080/8090 均返回 `200`；8080 当前 `index.html` 引用的 17 个静态资源全部返回 `200`。
- Playwright 使用 `admin/admin123` 打开 `/discovery/service`，点击第一行 `查看 / 编辑`：查看态抽屉展示 `基础信息 / 归属信息 / 服务标签`，5 个输入控件全部 disabled，无旧 `serviceView` class，底部为 `编辑 / 关闭`，console warning/error 为 `0`。
- 点击 `编辑` 后同一抽屉切换为 `编辑服务`，控件启用，底部显示 `提交 / 重置`；截图输出到 `.playwright-cli/service-view-form-readonly.png` 和 `.playwright-cli/service-view-form-editable.png`。
- 根据阅读态字段视觉反馈，已将查看态 5 个字段从 disabled `Input/Select` 改为 `ReadonlyField` 无边框文本；Playwright 复验阅读态 input/textarea 数量为 `0`、disabled input 数量为 `0`、Select 箭头不存在、字段背景透明、边框 `0px`、文字颜色 `rgb(31, 41, 55)`，截图输出到 `.playwright-cli/service-view-readonly-plain-text.png`。
- 根据阅读态名称字段反馈，已将 `FormItem rules` 限制到编辑态，并让名称字段的命名提示、字数统计和内联错误只在编辑态展示；阅读态不再显示 required 红星、`命名后不可修改` 和 `13/128`。
- Playwright 复验阅读态：`requiredMarkCount=0`，`hasNameHint=false`，`hasNameCountText=false`，`hasNameCountNode=false`，`hasErrNameNode=false`，截图输出到 `.playwright-cli/service-view-readonly-no-edit-hints.png`。

## 控制台资源页操作区统一

- [x] 盘点命名空间、注册发现、配置分组、MCP、A2A、治理工作台的页头、创建按钮和查询筛选布局差异
- [x] 抽取共享资源页页头与列表工具栏组件，固定标题、辅助文案、刷新、新建、列表标题和查询区结构
- [x] 将主要资源列表页切换到共享组件，保持蓝白/灰黑白企业软件风格，不重做侧边栏和全局背景
- [x] 增加静态校验，防止资源页继续绕开共享页头/工具栏各自实现
- [x] 运行前端脚本、lint、构建、context-kg 校验、重启服务和真实页面抽检

当前判断：

- 用户当前反馈集中在“资源创建按钮、查询搜索布局不统一”，不是要求再次整体换肤；本轮范围只统一资源页操作区的信息架构和承载方式。
- 资源页统一标准：页头左侧为业务路径、标题、说明，右侧为刷新和主创建动作；列表工具栏左侧为列表名与当前数量，右侧为查询、筛选、查询按钮和重置入口。
- 治理工作台已有“刷新 / 重置筛选 / 新建规则”的局部规则，本轮需要纳入统一资源页语言，同时避免破坏其两步新建规则流程。

修复：

- 新增 `components/ResourceLayout`，提供 `ResourceHeader` 和 `ResourceToolbar`，统一页头、右侧操作、列表标题、数量提示和查询区排列。
- 命名空间、注册发现服务、配置分组、AI MCP、AI A2A、治理工作台接入共享页头和工具栏；创建入口统一放到页头右侧，查询/重置统一放在列表工具栏右侧。
- 配置分组将“新建/刷新”从筛选条移到页头，并补齐显式“查询/重置”；治理工作台保留两步新建规则流程，只调整入口承载。
- 共享页头路径提示统一为企业蓝 `#0052d9`，不改侧边栏和全局背景。
- 新增 `verify-resource-page-layout.mjs`，并更新注册发现、配置分组已有校验，使脚本识别共享布局而不是私有 `header/filterBar`。
- 收到本次纠正后已更新 `context-kg/tasks/lessons.md`，固化资源页头和查询工具栏统一规则。
- 根据治理工作台搜索布局仍与其它资源不一致的反馈，撤回专用搜索区分支和额外说明行；治理工作台回到共享 `ResourceToolbar` 默认单行结构，左侧只展示 `规则清单 / 当前显示`，右侧展示类型筛选、搜索框、查询和重置。
- 根据配置分组截图对齐治理工作台筛选控件：`规则类型` 和 `关键字` 改为控件内 label，关键词 placeholder 收敛为 `规则名、服务、条件`，不再只依赖 placeholder 表达字段名。
- 根据最新截图纠正最终顺序和字段数量：治理工作台查询区改为 `关键字 / 规则类型 / 状态 / 查询 / 重置`，并让状态筛选参与本地列表过滤。
- 根据本次页面设计层级反馈，将治理工作台 `ResourceToolbar` 移出白色 `listPanel`，让查询工具栏独立处在灰色页面背景上；`listPanel` 只承载表格和分页，和配置分组列表保持同一页面结构。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && node scripts/verify-resource-page-layout.mjs` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 8080/8090 均返回 `200`；8080 当前 `index.html` 引用的 17 个静态资源全部返回 `200`。
- Playwright 使用 `admin/admin123` 登录态打开 `/namespace`、`/discovery/service`、`/configuration/group`、`/ai/mcps`、`/ai/a2a`、`/governance/workbench`，六个页面均存在统一标题、列表工具栏、当前数量、查询、重置和新建按钮；路径提示色为 `rgb(0, 82, 217)`；浏览器 console warning/error 为 `0`。
- 治理工作台搜索区回收为单行共享结构后，Playwright 与命名空间页对比验证：治理工作台工具栏无额外说明行，文本为 `规则清单 / 当前显示 9 / 9 条 / 查询 / 重置`，搜索框与标题行顶边节奏和命名空间页一致。
- 按配置分组同款控件内标签修正后，Playwright 验证治理工作台工具栏文本包含 `规则类型`、`关键字`、`查询`、`重置`，截图输出到 `.playwright-cli/workbench-toolbar-labeled.png`。
- 最终 Playwright 验证治理工作台查询区文本顺序为 `关键字 -> 规则类型 -> 状态 -> 查询 -> 重置`，所有控件顶边一致，截图输出到 `.playwright-cli/workbench-toolbar-final.png`。
- 页面层级修正后，真实浏览器验证治理工作台工具栏父级为页面容器、位于 `listPanel` 之前且不被 `listPanel` 包含；工具栏文本顺序为 `关键字 / 规则类型 / 状态 / 查询 / 重置`，截图输出到 `.playwright-cli/workbench-toolbar-page-structure.png`，浏览器 console warning/error 为 `0`。

## 控制台整体设计恢复原状

- [x] 识别本轮整体设计改动范围：全局样式、壳层、侧边栏、核心页面 module.less 和设计校验脚本
- [x] 恢复全局样式、布局壳层、页面样式到仓库原有版本
- [x] 恢复侧边栏宽度、Logo 尺寸和默认 TDesign 导航样式
- [x] 保留用户此前明确要求的 `Lattice.Hub` 品牌文案与 Logo 资产
- [x] 运行前端静态验证、lint、构建、重启和真实页面抽检

当前判断：

- 用户明确要求“整体恢复原状”，因此撤回本轮 `redesign-existing-projects` 引入的整体设计语言、侧栏企业蓝白重做和页面 token 化。
- 业务功能、配置中心能力、授权能力、统一操作按钮、Lattice.Hub Logo 等之前明确完成的功能类改动不属于本次恢复范围。

修复：

- `styles/index.less` 恢复为原始 body/html/code 基础样式，移除本轮新增的 Console 设计 token、全局按钮/表格/抽屉覆盖和背景纹理。
- `layouts` 相关样式恢复原始壳层、Header、Page 和 Menu Logo 样式。
- 命名空间、注册发现、配置分组、配置文件详情、AI MCP/A2A、治理工作台、登录页等页面样式恢复到仓库原有版本。
- 侧边栏宽度恢复为 `232px`，完整 Logo 尺寸恢复为 `184x32`；`Menu.tsx` 仅保留 `Lattice.Hub ${version}` 文案。
- `verify-brand-logo.mjs` 不再约束整体设计主色和侧栏重做，只保留 Lattice.Hub 品牌和恢复后的原始尺寸检查。
- 对配置中心这类已经由业务 TSX 引用的新 class，仅补回基础 TDesign 风格样式，避免恢复原状后出现 CSS module undefined / 页面裸奔。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- CSS module 引用审计通过，配置中心相关 TSX 无缺失 class。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`，8080/8090 均返回 `200`。
- 8080 当前 `index.html` 引用的 17 个静态资源全部返回成功，无 hash 资源 404。
- Playwright 使用 `admin/admin123` 登录态打开 `/discovery/service` 和 `/configuration/group`，侧栏宽度 `232`、Logo `184x32`，配置中心列表可见，浏览器 console warnings/errors 为 `0`。

## 侧边栏企业软件视觉修正

- [x] 复核用户截图反馈：当前侧边栏 active 状态和品牌色偏轻快，不符合企业软件气质
- [x] 将全局主色从偏绿收敛为企业蓝，并保留黑灰白中性色
- [x] 重做侧边栏密度、logo 尺寸、菜单层级、hover 和 active 状态
- [x] 更新相关静态校验，避免旧尺寸和旧色彩断言误判
- [x] 运行前端静态验证、lint、构建、重启和真实页面抽检

当前判断：

- 配置中心和控制台属于企业管理软件，默认应采用蓝白或黑灰白体系；偏绿的品牌色和大面积浅绿色 active 背景会让侧边栏显得消费化。
- 侧边栏应优先表达导航层级和当前页面位置：白底、黑灰文本、蓝色细指示条、轻量 hover；不要使用大胶囊和过大的留白。
- Logo 可以保留用户指定图形和 `Lattice.Hub` 文案，但在侧边栏里需要降低视觉占比，把更多空间留给导航。

修复：

- 全局 `--console-brand` 从偏绿改为企业蓝 `#0052d9`，同步更新 TDesign brand focus/active/light、按钮阴影、输入 hover、页面背景和抽屉背景中的弱品牌色。
- 侧边栏宽度从 `232px` 收敛为 `220px`，完整 Logo 从 `204x36` 收敛为 `186x33`，折叠 Logo 从 `36x36` 收敛为 `32x32`。
- 侧边栏改为白底、黑灰文本、18px 图标、34-38px 菜单高度；hover 使用浅灰，active 使用浅蓝底、企业蓝文字和左侧 3px 指示线。
- 默认展开恢复为所有一级分组全部展开，保证全量导航入口可见。
- 修正顶层单菜单项缩进，`命名空间` 和其它一级分组标题保持同一图标/文字垂直列。
- `verify-brand-logo.mjs` 增加侧栏宽度、logo 尺寸、企业蓝主色、禁止旧绿色 token、默认全部展开和顶层单菜单项对齐的静态检查。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`，8080/8090 均返回 `200`。
- 8080 当前 `index.html` 引用的 17 个静态资源全部返回成功，无 hash 资源 404。
- Playwright 使用 `admin/admin123` 登录态打开 `/discovery/service`：侧边栏宽度 `220`，Logo 宽度 `186`，`服务实例` active 为浅蓝底 `rgb(232, 241, 255)` 和企业蓝文字 `rgb(0, 82, 217)`，浏览器 console warnings/errors 为 `0`。
- 针对侧边栏对齐复测：所有一级分组均默认展开，`命名空间` 与 `AI 工具 / 注册发现 / 配置管理` 等一级项均为 `x=42/iconX=58/contentX=84`，顶层单项和分组标题已在同一垂直列。

Review：

- 截图反馈的根因不是单个 active 样式，而是主色、logo 占比、默认展开策略和侧栏密度共同偏离企业软件气质；本轮已从设计语言层收敛。
- 侧栏目前保持白底蓝白方案；如果后续要走灰黑白，可以在同一 token 层切换，不需要改业务页面结构。
- 单菜单项和分组菜单项在 TDesign DOM 中缩进不同，视觉验收不能只看 active 颜色，还要量图标列和文字列是否一致。

## 控制台整体设计语言统一

- [x] 扫描现有前端技术栈、布局壳层和主要工作台页面样式
- [x] 诊断整体设计不一致点：字体、色彩、页面背景、卡片边界、表格、按钮、抽屉和导航状态
- [x] 建立全局设计语言变量与基础控件状态，不迁移框架、不破坏 TDesign
- [x] 收敛布局壳层、左侧导航、Header、主内容背景和页面容器
- [x] 将命名空间、服务、配置分组、AI MCP/A2A、治理工作台等主要工作台页面对齐同一视觉语言
- [x] 运行前端静态验证、lint、构建和真实页面抽检

诊断：

- 当前前端是 React + Vite + TDesign React 1.18 + LESS，适合通过全局 token 和现有 module.less 做低风险统一。
- 全局字体仍是浏览器系统栈，数据界面的数字没有统一 tabular 处理；不同页面各自写 `#1f2937/#111827/#1677ff/#007f78` 等色值，品牌色和中性色不稳定。
- 主工作台页面已有相近结构，但命名空间、注册发现、配置分组、AI、治理分别维护自己的背景、边框、指标条和表格 surface，视觉语言不完全一致。
- 左侧导航和主内容区仍偏 TDesign 默认模板感，页面背景、表格边界、抽屉、按钮 hover/active 的质感不统一。
- 本轮应先做统一语言层：字体、色彩、surface、交互状态、导航 active、表格和抽屉；避免大幅改业务布局或新增依赖。

修复：

- 在 `styles/index.less` 建立 Console 级设计 token，并映射到 TDesign 品牌色、页面背景、文本、边框、按钮、输入、表格、抽屉和滚动条等基础状态。
- 统一主壳层视觉：左侧导航、Header、页面容器、面包屑、主内容背景改为同一套轻量工作台语言，保留现有 React/Vite/TDesign 架构。
- 对齐命名空间、注册发现服务、配置分组、配置文件详情、AI MCP/A2A、治理工作台、登录页等高频页面的 padding、标题、surface、表格边界、指标条和阴影。
- 更新注册发现布局静态校验到新的 `100dvh` 视区模型和全局 token，避免旧断言把新设计语言误判为回归。
- 补充配置文件详情页固定高度容器的 `box-sizing`，避免 padding 在固定视区高度下撑出页面。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`，8080/8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录态打开 `/login`、`/namespace`、`/discovery/service`、`/configuration/group`、`/ai/mcps`、`/ai/a2a`、`/governance/workbench`，页面均能正常展示且未回跳登录，浏览器 console warnings/errors 为 `0`。
- 直接校验 8080 当前 `index.html` 引用的 17 个 `/assets/*` 静态资源，全部返回成功，无 hash 资源 404。

Review：

- 这轮没有重写组件体系，也没有替换 TDesign；只通过全局 token、壳层样式和主要工作台页面样式收敛设计语言，风险集中在 LESS 层。
- 少量业务语义色仍保留在治理规则类型、AI 状态等局部状态标签内；这些属于语义区分，不应强行压成单色。

## MCP 服务详情入口与行操作查看语义收敛

- [x] 确认 MCP 服务当前列表同时暴露“查看工具”和“编辑”，工具能力不在编辑入口中呈现
- [x] 将 MCP 服务列表主入口收敛为“查看”，打开服务详情/工具抽屉
- [x] 在 MCP 服务详情抽屉内部保留“编辑 Server”入口，用于进入修改
- [x] 将统一 `viewEdit` 操作展示从“查看 / 编辑”改为“查看”
- [x] 将仍直接显示 `edit` 的配置分组、服务别名列表入口改为查看图标语义
- [x] 运行前端静态验证、lint、构建和真实页面回归

当前判断：

- MCP 服务的“查看”入口应该能看到工具，因为工具是 MCP Server 最核心的可观察能力；单独放“查看工具”和“编辑”两个行按钮会让入口语义分裂。
- 行操作列不应继续强调“编辑”；资源列表的主入口统一叫“查看”，详情弹窗/抽屉内部再提供编辑动作，降低误操作和操作列拥挤。

修复：

- `OperationButton` 的 `viewEdit` 显示从“查看 / 编辑”改为“查看”，图标也从编辑图标改为查看图标。
- MCP 服务列表操作列移除单独“查看工具”和“编辑”按钮，保留“查看 / 授权 / 删除”；点击查看打开“MCP 服务详情”抽屉，并直接展示工具能力。
- MCP 服务详情抽屉内部继续提供可见的“编辑 Server”入口，空态也保留编辑入口，满足从详情进入修改。
- 配置分组、服务别名这两个仍直接使用 `action="edit"` 的列表入口改成 `action="view"` 展示，避免行操作继续突出编辑语义。
- 更新静态回归脚本，要求 MCP 列表主入口必须是查看详情，列表操作列不能再直接暴露编辑或单独查看工具。

验证：

- `cd console/web && node scripts/verify-ai-resource-authorization.mjs && node scripts/verify-operation-button-icons.mjs && node scripts/verify-namespace-actions.mjs && node scripts/verify-auth-drawer-actions.mjs` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg && git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`，8080/8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录态打开 `/ai/mcps`，确认操作列存在 `查看 / 授权 / 删除`，不存在直接 `编辑` 和 `查看工具`；点击 `查看` 后抽屉显示 `MCP 服务详情`、工具浏览或空态，以及可见 `编辑 Server`，浏览器 console errors/warnings 均为 `0`。

## AI 工具 MCP/A2A 资源授权模型评估

- [x] 确认当前 specification `ResourceType` 与 `StrategyResources` 未包含 MCP/A2A 资源
- [x] 确认 MCP Server 与 A2A Agent 已是独立注册实体，并具备稳定资源 ID、Store 和 Cache
- [x] 盘点后端授权适配缺口：资源字段映射、资源存在性检查、详情回显和操作鉴权收集
- [x] 盘点前端适配缺口：策略资源枚举、策略编辑器资源选择、授权抽屉和 AI 列表行内授权入口
- [x] 形成推荐方案：先更新 spec 资源模型，再在 control-plane 和 Console 接入授权链路

当前判断：

- MCP Server 和 A2A Agent 都不应借用 `Services` 或 `PolicyRules` 授权；它们有独立 CRUD、独立表、独立 cache 和独立页面，应成为独立 `ResourceType`。
- specification 需要先更新，因为授权策略详情的 `StrategyResources` 是强类型字段集合，不只依赖 `ResourceType` 字符串枚举。
- 推荐新增 `MCPServers` 与 `A2AAgents` 两类资源；MCP Server Tool 和 A2A Agent Skill 首期不单独作为授权资源，先继承所属 MCP Server / A2A Agent 的管理权限。
- control-plane 适配时要横向补齐四条链路：策略资源字段映射、授权请求资源存在性校验、策略详情资源摘要回显、MCP/A2A CRUD 操作鉴权收集。
- Console 适配时需要补 `PolicySourceType`、`PolicyResources` 字段、策略编辑器资源加载、`AuthorizeInput` 摘要，以及 MCP/A2A 列表页统一授权操作。

Review：

- 当前缺口不是单纯前端“放开按钮”；如果前端直接传不存在的资源类型，后端 `authorizeCheckResourceExist` 会返回不支持，策略详情也无法回显。
- A2A 旧 ADR 已提到“管理 API 走现有 auth-system”，但当时只写了建议资源类型，没有落到 `specification/api/v1/security/auth.proto` 和本仓鉴权链路。
- 下一步应先在 `../specification` 更新 `ResourceType` 与 `StrategyResources` 并生成代码，再更新 control-plane 依赖和前后端授权适配。

## AI 工具 MCP/A2A 资源授权实现

- [x] 补充 MCP/A2A 授权资源映射红测，并确认当前失败
- [x] 更新 `../specification` 的 `ResourceType` 与 `StrategyResources`，生成 Go 代码
- [x] 更新 control-plane auth 类型映射、资源存在性校验、策略详情回显和操作鉴权收集
- [x] 更新 Console 策略资源模型、授权抽屉摘要和 MCP/A2A 列表授权入口
- [x] 运行 Go、前端静态验证、构建和必要接口回归

当前判断：

- MCP Server 和 A2A Agent 是独立资源，不能复用服务、策略规则或配置分组的授权类型；specification 已新增 `MCPServerResources` 和 `A2AAgentResources`。
- Tool / Skill 首期继承所属 MCP Server / A2A Agent 的读权限；本轮不把 Tool / Skill 拆成单独资源，避免资源模型过细导致策略编辑复杂化。
- Console 行内授权入口提交的 `resource_type` 是字符串，后端 `AuthorizeResources` 必须同时支持 enum 名称和 snake_case 别名，否则会在资源存在性校验前失败。

修复：

- `../specification/api/v1/security/auth.proto` 新增 MCP/A2A 资源枚举和 `StrategyResources.mcp_servers/a2a_agents` 字段，并重新生成 Go 代码。
- control-plane 使用本地 `../specification`，补齐 `ResourceFieldNames`、`SearchTypeMapping`、资源存在性检查、策略详情资源回显、兼容资源检查和 AI 工具操作鉴权收集。
- MCP Server / A2A Agent 列表读操作接入资源级过滤；创建、更新、删除、工具/技能/Agent Card 查询接入对应服务端函数鉴权。
- Console 策略编辑器支持选择 MCP Server / A2A Agent 资源，策略详情可回显 `mcp_servers/a2a_agents`，MCP/A2A 列表页增加统一授权按钮，授权抽屉支持 AI 资源摘要。
- `AuthorizeResources` 增加 `MCPServerResources`、`A2AAgentResources` 以及 `mcp_servers/a2a_agents` 等别名映射，修复真实冒烟中 `invalid resource_type` 的问题。

验证：

- 新增 `TestAIResourceFieldMappings`、`TestResourceConvertIncludesAIRegistryResources`、`TestAIRegistryResourceTypeFilter`；其中 `TestAIRegistryResourceTypeFilter` 修复前失败于 `resTypeFilter` 未识别 MCP/A2A 资源类型，修复后通过。
- `go test ./apis/pkg/types/auth ./plugin/access_control/auth/policy ./plugin/apiserver/httpserver/aimcp ./plugin/apiserver/httpserver/aia2a -count=1` 通过。
- `cd console/web && node scripts/verify-ai-resource-authorization.mjs` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 在 control-plane 与 `../specification` 均通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`，8080/8090 均返回 `200`。
- Node 冒烟使用 `admin/admin123` 登录后创建临时 MCP Server 和 A2A Agent，等待缓存可见，分别调用 `/auth/v1/resources/authorize` 授权 `MCPServerResources` 与 `A2AAgentResources`，再查询 MCP Tools 和 A2A Skills，最后清理临时资源；结果返回 `ok=true`。

Review：

- 这轮真实接口冒烟发现前端/后端契约还有一层字符串解析入口，不能只补 `SearchTypeMapping`；以后新增授权资源类型要同步 `resTypeFilter` 或改成统一从 enum 名称解析。
- Admin 主账号虽然默认策略覆盖所有 `ResourceType_value`，但操作接口仍要显式记录服务端函数与资源上下文；否则普通账号和行级授权无法正确生效。

## admin 服务列表无权限修复

- [x] 用 `admin/admin123` 真实复现服务列表行操作显示无权限
- [x] 对比登录响应、服务列表接口响应和后端权限打标链路定位根因
- [x] 新增失败测试，约束服务列表响应必须保留服务 ID 和默认操作权限
- [x] 修复服务列表响应构造，复用服务领域模型的 `ToSpec()` 转换
- [x] 运行 Go 针对性测试、all-mode 重启、接口与真实页面回归
- [x] 记录 review 和 lessons

当前判断：

- admin 登录响应返回 `role: main`，`ResourcePredicate` 对主账号会直接放行，因此 admin 服务列表不应显示无权限。
- 真实接口返回的服务行 `id=""` 且 `editable=false/deleteable=false`；根因是 `pkg/service/service.go` 在 `GetServices` / `GetAllServices` 中手工构造 `apiservice.Service`，漏掉 `id/ctime/mtime/revision/editable/deleteable` 等字段，而不是前端按钮误判。

修复：

- 新增 `TestBuildServiceQueryItemKeepsIdentityAndOperateFlags`，要求服务列表响应转换保留 `id/name/namespace/business/department/comment/metadata/instance count/editable/deleteable`。
- 新增 `buildServiceQueryItem`，统一通过 `svctypes.Service.ToSpec()` 构造服务列表响应，再补充 `total_instance_count/healthy_instance_count`。
- `GetServices` 和 `GetAllServices` 均改用 `buildServiceQueryItem`，避免继续手工构造残缺响应。

验证：

- `go test ./pkg/service -run TestBuildServiceQueryItemKeepsIdentityAndOperateFlags -count=1` 修复前失败于缺少转换函数，修复后通过。
- `go test ./pkg/service/... -count=1` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- curl 使用 `admin/admin123` 登录后请求 `/naming/v1/services?offset=0&limit=10`，第一条服务返回 `id=spec-svc-checkout-20260610`、`editable=true`、`deleteable=true`。
- Playwright 使用 `admin/admin123` 登录后打开服务列表，行操作显示 `查看 / 编辑` 和 `删除` 可点击，不再是“无权限操作”；浏览器 console errors/warnings 均为 `0`。

Review：

- 这不是 admin 默认策略缺失，也不是前端误读；是后端服务列表响应构造绕开 `ToSpec()` 导致权限字段保持 proto bool 零值。
- 后续列表响应不要手工挑字段重建 proto；如果领域模型已有 `ToSpec()`，应先复用再追加列表专属聚合字段。

## 命名空间授权入口补齐

- [x] 确认命名空间当前后端授权资源契约
- [x] 新增静态反回归检查，要求命名空间列表提供授权入口并使用资源名作为资源 ID
- [x] 在命名空间列表行操作补回统一授权按钮
- [x] 修正命名空间授权抽屉的 `resource_id/resource_name`
- [x] 运行前端静态验证、lint、构建、all-mode 重启和真实页面回归
- [x] 记录 review 和 lessons

当前判断：

- specification 已有 `ResourceType.Namespaces`；后端授权校验 `Namespaces` 时按 `resourceID` 查命名空间名称。
- 命名空间作为资源应和配置分组、配置文件一样在资源列表上下文可授权；不能只藏在详情抽屉里。

修复：

- `verify-namespace-actions.mjs` 改为要求命名空间列表行操作提供统一 `OperationButton action="authorize"`，并要求授权资源类型使用 `PolicySourceType.Namespaces`。
- 命名空间列表行操作补回授权按钮，保持 `查看 / 编辑`、`授权`、`删除` 三个统一图标操作。
- 命名空间授权 `resource_id` 改为命名空间名称；后端 `authorizeCheckResourceExist` 对 `Namespaces` 正是按该名称查资源。
- `AuthorizeInput` 增加 `Namespaces` 资源摘要分支，命名空间资源只展示命名空间、资源 ID 和资源名称，不再显示空的第二段 `资源名称: -`。

验证：

- `cd console/web && node scripts/verify-namespace-actions.mjs` 修复前失败于缺少行内授权入口；补摘要分支前失败于缺少命名空间资源摘要分支；修复后通过。
- `cd console/web && node scripts/verify-config-group-detail-design.mjs` 通过，确认配置文件授权摘要未受影响。
- `cd console/web && set -e; for script in scripts/verify-*.mjs; do node "$script"; done && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录 `/namespace`，确认命名空间行操作出现 `授权`；点击 `spec-governance` 行授权后，抽屉资源摘要显示 `资源类型 Namespaces`、`命名空间 spec-governance`、`资源ID spec-governance`、`资源名称 spec-governance`，浏览器 console errors/warnings 均为 `0`。

Review：

- 命名空间已有独立后端授权资源类型，不能因为之前操作列收敛就移除列表授权入口；资源列表上下文应直接暴露授权能力。
- 命名空间没有独立 numeric/string id 字段，当前授权资源 ID 就是命名空间名称；前端不能继续传 `editorState.data?.id`。

## 配置文件授权入口补齐

- [x] 确认配置文件当前后端授权资源契约
- [x] 新增静态反回归检查，要求配置文件详情提供授权入口
- [x] 更新通用授权抽屉资源摘要，支持配置文件三段资源名展示
- [x] 在配置文件详情页接入授权按钮和授权抽屉
- [x] 运行前端静态验证、lint、构建、all-mode 重启和真实页面回归
- [x] 记录 review 和 lessons

当前判断：

- 现有后端/spec 的鉴权资源枚举没有独立 `ConfigFiles`；配置文件创建、更新、删除、发布、发布历史等后端鉴权均通过所属 `ConfigGroups` 资源转换。
- 前端仍应在配置文件详情页展示授权入口；当前授权会落到所属配置分组资源上，并在抽屉摘要中展示具体文件名，避免用户从文件详情页找不到授权位置。

修复：

- `FileView` 标题操作区新增统一 `OperationButton action="authorize"` 授权入口，只在已选中配置文件且有编辑权限时展示。
- 配置文件详情页接入 `AuthorizeInput`，按当前后端契约提交 `PolicySourceType.ConfigGroups` 和所属配置分组 ID，同时把 `namespace/group/fileName` 传给授权抽屉作为资源摘要。
- `AuthorizeInput` 的资源摘要支持配置文件字段；文件名可能包含 `/`，因此使用第三段之后整体 join 回完整文件名，避免 `aaa/1111` 被截成 `aaa`。
- `verify-config-group-detail-design.mjs` 增加配置文件授权入口、授权资源类型和斜杠文件名解析的静态反回归检查。

验证：

- `cd console/web && node scripts/verify-config-group-detail-design.mjs` 修复前失败于缺少配置文件授权入口，修复后通过。
- `cd console/web && set -e; for script in scripts/verify-*.mjs; do node "$script"; done && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录后进入 `/configuration/group/files?namespace=spec-governance&group=123`，选中 `aaa/1111`，文件标题区出现 `授权` 操作；打开授权抽屉后资源摘要显示 `资源类型 ConfigGroups`、`命名空间 spec-governance`、`配置分组 123`、`配置文件 aaa/1111`、资源 ID 和完整资源名称，浏览器 console errors/warnings 均为 `0`。

Review：

- 这次不能因为后端没有独立 `ConfigFiles` 枚举就隐藏配置文件授权入口；配置文件同样是用户心智里的资源，入口应出现在文件详情上下文中。
- 当前实现不改后端鉴权模型，只把文件详情入口映射到所属配置分组授权；如果后续 specification 增加独立配置文件资源类型，再把 `resource_type/resource_id` 切换为文件资源即可。

## 前端操作按钮图标统一

- [x] 盘点主要列表页行操作按钮的现状和重复图标模式
- [x] 新增静态反回归检查，要求主要行操作统一走共享组件
- [x] 抽取统一 `OperationButton` / `ConfirmOperationButton`
- [x] 替换配置中心、命名空间、认证、注册发现、治理工作台和 AI 页面主要操作列
- [x] 运行前端静态验证、lint、构建、all-mode 重启和真实页面冒烟
- [x] 记录 review 和 lessons

当前判断：

- 行操作语义应统一映射到一套图标：查看/编辑、编辑、授权、删除、发布、回滚、复制、查看工具、查看 Token、刷新等不能由各页面自行选择。
- 删除、回滚这类确认操作需要统一稳定触发结构，避免 `Tooltip` 直接包 `Popconfirm` 造成悬浮提示不稳定。

修复：

- 新增 `components/OperationButton`，集中维护操作语义到图标、Tooltip、`aria-label` 和 `title` 的映射。
- 新增 `ConfirmOperationButton`，把确认类操作统一为稳定 `span` target + `Popconfirm` + 图标按钮，避免各页面自行组合。
- 替换配置分组、配置发布记录、命名空间、认证主体/策略、服务列表、服务别名、治理工作台、A2A、MCP Server 主要行操作。
- 新增 `verify-operation-button-icons.mjs`，要求主要页面使用统一组件，并检查页面使用 `<Tooltip>` 时必须从 `tdesign-react` 导入。
- 更新旧的配置中心、认证、命名空间、注册发现静态检查，让它们校验统一组件语义，而不是继续绑定旧私有图标实现。

验证：

- `cd console/web && node scripts/verify-operation-button-icons.mjs` 修复前失败于缺少统一组件，修复后通过。
- `cd console/web && set -e; for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录后验证命名空间、配置分组、治理工作台页面可正常渲染；行操作按钮可访问名称统一为“查看 / 编辑 / 编辑 / 授权 / 删除”等语义；最新 console 日志没有 `error`。

Review：

- 本轮先按统一组件收敛主要资源列表和配置发布记录的操作列；表单内部的标签行、协议行增删按钮暂不纳入资源行操作统一范围，避免误改局部编辑器。
- 初次替换时误删了部分页面刷新按钮仍使用的 `Tooltip` import，真实浏览器发现 `Tooltip is not defined`；已补齐 import，并将该类问题纳入 `verify-operation-button-icons.mjs`。

## 配置分组配置加密数恢复

- [x] 确认上一轮过度删除：应移除分组默认加密配置，但应保留配置加密数
- [x] 更新静态反回归检查，让当前缺少配置加密数的实现失败
- [x] 使用配置文件列表真实计算当前页每个配置分组的加密文件数量
- [x] 在配置分组 KPI 和表格列恢复配置加密数/加密数
- [x] 运行前端静态验证、构建、all-mode 重启和真实页面回归
- [x] 记录 review 和 lessons

当前判断：

- 配置加密数是配置文件维度的聚合统计，可以在配置分组列表展示；但它不能来自配置分组 metadata 的 `defaultEncrypted`，也不能恢复成“分组默认加密”编辑项。
- 当前后端配置分组列表只返回 `fileCount`，没有返回加密文件聚合数；本轮以前端复用配置文件列表接口计算当前页展示数量，避免扩大到 specification 变更。

修复：

- `verify-config-group-list-design.mjs` 要求配置分组列表保留 `配置加密数` KPI 和 `加密数` 表格列，并要求通过 `describeAllConfigFiles` / `encryptedFileCounts` 从配置文件真实计算。
- `group.tsx` 为当前页分组异步加载配置文件列表，按 `encrypted=true` 聚合加密文件数量；KPI 汇总当前页加密文件总数，表格展示每个分组的加密数。
- 继续禁止 `加密占比`、`加密状态筛选`、`defaultEncrypted` metadata 伪口径和编辑抽屉的默认加密开关。

验证：

- `cd console/web && node scripts/verify-config-group-list-design.mjs` 修复前失败于缺少 `配置加密数`，修复后通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录后打开配置分组列表，确认 KPI 有 `配置加密数`，表格列有 `加密数`，没有 `加密占比` 和 `加密状态` 筛选；点击编辑抽屉，确认只保留基础信息和标签，没有默认配置、文件格式默认值和是否默认加密，浏览器控制台无 `error`。

Review：

- 上一轮删除了错误的默认加密配置，但把真实聚合指标也删掉了，属于过度修改。
- 本轮恢复的是只读统计口径：配置加密数来自配置文件真实 `encrypted` 字段聚合；配置分组编辑仍不承载文件级默认格式/默认加密配置。

## 配置分组文件级默认配置移除

- [x] 确认配置分组编辑抽屉里误放了文件格式默认值和默认加密
- [x] 先更新配置分组静态反回归检查，让当前错误实现失败
- [x] 移除配置分组编辑抽屉中的文件级默认配置和错误 metadata 写入
- [x] 移除配置分组列表中基于错误 metadata 推导的加密占比/加密状态
- [x] 运行前端静态验证、构建、all-mode 重启和真实点击回归
- [x] 记录 review 和 lessons

当前判断：

- 配置分组的边界是 `namespace/name/comment/department/business/metadata(labels)` 这类分组基础信息；文件格式、文件加密应在配置文件创建/编辑链路处理，不应作为配置分组默认值。
- 当前实现把 `defaultFormat/defaultEncrypted` 写入配置分组 `metadata`，并在列表里用 `defaultEncrypted` 推导“加密占比”，这是把文件级属性错误下沉到了分组级。

修复：

- `verify-config-group-list-design.mjs` 反向约束配置分组不能出现文件格式默认值、默认加密、加密状态筛选和加密占比列。
- `ConfigGroupEditor.tsx` 移除“默认配置”区，不再提交 `defaultFormat/defaultEncrypted`；编辑时仍过滤历史误写入的这两个 metadata key，避免它们作为普通标签显示。
- `group.tsx` 移除加密状态筛选、加密占比 KPI 和加密占比列，列表只展示分组名称、命名空间、文件数、待发布和操作。

验证：

- `cd console/web && node scripts/verify-config-group-list-design.mjs` 修复前失败于配置分组列表仍用 `defaultEncrypted` 推导加密指标，修复后通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- console/web/src/pages/Configuration/Group/ConfigGroupEditor.tsx console/web/src/pages/Configuration/Group/group.tsx console/web/scripts/verify-config-group-list-design.mjs context-kg/tasks/todo.md context-kg/tasks/lessons.md` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录后打开配置分组列表，确认总栏只显示分组数、配置文件数、待发布数，筛选只有关键字、命名空间、发布状态，表格列为分组名称、命名空间、文件数、待发布、操作；点击编辑后抽屉只显示基础信息和标签，未出现默认配置、文件格式默认值和是否默认加密，浏览器控制台无 `error`。

Review：

- 本轮纠正的是配置分组与配置文件的领域边界：格式和加密是配置文件属性，不属于配置分组默认配置。
- 后续如果需要文件格式/加密批量策略，应先有后端/spec 明确字段和产品语义，不能再用分组 metadata 私自承载。

## 配置分组编辑抽屉展示修复

- [x] 使用 `admin/admin123` 真实复现配置分组列表点击编辑按钮后页面空白的问题
- [x] 从浏览器控制台定位运行时异常，并补充静态反回归检查
- [x] 修复配置分组编辑抽屉缺失的组件导入
- [x] 运行前端静态验证、构建、all-mode 重启和真实点击回归
- [x] 记录 review

当前判断：

- 编辑按钮点击后不是路由或权限问题，而是 `ConfigGroupEditor` 渲染时使用了 `<Switch />`，但没有从 `tdesign-react` 导入，导致运行时 `ReferenceError: Switch is not defined`，React 页面直接崩成空白。
- 该问题需要用真实浏览器点击验证，单纯列表渲染或后端接口 200 无法覆盖。

修复：

- `verify-config-group-list-design.mjs` 增加 `Switch` 导入约束，避免编辑抽屉再次使用未导入组件。
- `ConfigGroupEditor.tsx` 从 `tdesign-react` 补齐 `Switch` 导入。

验证：

- `cd console/web && node scripts/verify-config-group-list-design.mjs` 修复前失败于 `Switch` 未从 `tdesign-react` 导入，修复后通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl` 验证 8080 和 8090 均返回 `200`。
- Playwright 使用 `admin/admin123` 登录后进入 `/configuration/group`，点击首行编辑按钮，抽屉显示 `编辑配置分组`，并回填命名空间、配置分组名称、描述、部门、业务、默认配置与加密开关；浏览器控制台没有 `error`。

Review：

- 本轮根因是编辑抽屉点击后才挂载的 JSX 组件缺少 import；列表首屏可用不代表行操作抽屉可用。
- 后续配置中心列表行操作需要覆盖真实点击路径，至少检查抽屉标题、关键字段回填和浏览器控制台错误。

## 配置中心完整链路设计适配

- [x] 对照完整链路设计文档确认当前列表页和详情页缺口
- [x] 先补静态反回归检查，覆盖列表页总栏、筛选、授权抽屉、创建/编辑抽屉结构，以及详情页文件标题区字段减法
- [x] 实现配置分组列表页 KPI 总栏、筛选条、表格列和授权抽屉入口
- [x] 调整创建/编辑配置分组抽屉的信息分段和默认配置字段
- [x] 调整配置分组详情当前文件标题区，只保留文件名、副标题、格式/加密 tag、修改时间、创建时间、加密算法、文件标签
- [x] 运行前端静态验证、lint、构建、smoke 和 all-mode 重启验收
- [x] 记录 review

当前判断：

- 详情页已经具备文件 tree、业务画布、发布记录内部 Tabs、灰度发布和订阅查询基础结构，本轮不重复大改。
- 列表页仍是旧表格形态，缺少总栏 KPI、命名空间/加密状态/发布状态筛选，以及授权抽屉的资源摘要与权限范围语义，是当前完整链路设计的主要缺口。
- 灰度优先级的真实后端/spec 字段仍需要单独契约支撑；本轮只保持已有前端入口和多灰度链路，不把 store/cache/key 暴露成前端字段。

修复：

- 新增 `verify-config-group-list-design.mjs`，覆盖配置分组列表页页面骨架、总栏 KPI、筛选条、设计要求表格列、授权抽屉语义和创建/编辑抽屉结构。
- 更新 `verify-config-group-detail-design.mjs`，禁止当前文件标题区继续展示抽象“文件上下文”，并约束字段区只保留修改时间、创建时间、加密算法和文件标签。
- 配置分组列表页改为浅灰工作区 + 白色业务面：顶部展示分组数、配置文件数、加密占比、待发布数；筛选条包含关键字、命名空间、加密状态、发布状态；表格列调整为分组名称、命名空间、文件数、加密占比、待发布、操作。
- 创建/编辑配置分组抽屉拆成基础信息和默认配置，默认文件格式、是否默认加密、标签写入 `metadata`，不新增后端字段。
- 授权抽屉补资源摘要、权限范围、授权对象三段，继续使用现有配置分组资源授权接口。
- 配置分组详情当前文件标题区改为文件名 + `命名空间 / 配置分组` 副标题 + 文件格式/加密/状态 tag，字段区移除命名空间、配置分组、订阅客户端、活跃版本、发布状态等重复信息。

验证：

- 修复前 `node scripts/verify-config-group-list-design.mjs` 失败于缺少列表页工作区骨架；`node scripts/verify-config-group-detail-design.mjs` 失败于详情页仍展示“文件上下文”。
- 修复后 `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 8080 和 8090 均返回 `200`。
- `cd console/web && node scripts/smoke-configuration-flow.mjs` 通过。
- Playwright 使用 `admin/admin123` 登录后打开 `/configuration/group`，列表页显示总栏、四项筛选和设计要求表格列；点击分组进入详情后，文件 tree 按 `/` 分层，选择文件后当前文件标题区显示文件名和 `spec-governance / 123`，字段区只显示修改时间、创建时间、加密算法、文件标签。
- Playwright 打开授权抽屉，能看到资源摘要、权限范围、授权对象；浏览器 console errors/warnings 均为 `0`。

Review：

- 本轮只把附件要求中当前缺口最大的列表页链路补齐，并收紧详情页字段展示；发布、多灰度、订阅查询沿用前序已落地结构。
- 活跃版本和订阅客户端属于文件/订阅查询视角，不在配置分组列表展示；加密占比目前基于分组默认加密 metadata 推导，后续若后端提供聚合字段再替换。
- 灰度优先级真实命中规则仍需要 specification / 后端契约支持，不能只靠前端字段完成。

## 配置客户端读取发布/多灰度测试

- [x] 明确客户端读取链路测试范围：全量发布、多个灰度发布、未命中灰度回落全量、同时命中多个灰度取最新版本
- [x] 补服务层单元测试，直接覆盖 `GetConfigFileWithCache` 返回给客户端的 `code/file/content/version`
- [x] 修复测试暴露的客户端发现响应缺字段问题
- [x] 运行定向 Go 测试和必要回归检查
- [x] 记录 review

当前判断：

- 配置中心已有缓存层保留多个 active gray、选择器按 version/mtime 取最新的局部测试，但缺少客户端读取最终响应的链路测试。
- 客户端读取链路必须断言响应 `Code`、`File`、`Content` 和 `Version`，否则 Nacos/HTTP 客户端即使命中正确发布版本，也可能拿不到配置正文。

修复：

- 新增 `TestGetConfigFileWithCacheReturnsNormalAndMatchedGrayContent`，覆盖普通客户端回落全量发布、客户端 A 命中灰度 A、客户端 B 命中灰度 B、客户端同时命中多个灰度时选择最新版本。
- `GetConfigFileWithCache` 成功、未找到、无变化和异常响应统一写入真实 `Code`；成功响应补回 `File`。
- `toClientInfo` 改为返回客户端发现协议实际需要的 `ConfigFileRelease`，保留 content、version、release type、release name、mtime/md5 等发布字段。

验证：

- 修复前 `go test ./pkg/config -run TestGetConfigFileWithCacheReturnsNormalAndMatchedGrayContent -count=1` 失败于成功响应 `Code=0`，且后续会缺少 `File`。
- 修复后 `go test ./pkg/config -run TestGetConfigFileWithCacheReturnsNormalAndMatchedGrayContent -count=1` 通过。
- `go test ./pkg/config -run 'TestGetConfigFileWithCacheReturnsNormalAndMatchedGrayContent|TestSelectMatchedGrayRelease|TestPublishConfigFileDoesNotBlockOnActiveGrayRelease' -count=1` 通过。
- `go test ./pkg/cache/config ./pkg/cache/gray ./apis/pkg/types/config -run 'TestConfigFileCacheKeepsMultipleActiveGrayReleases|TestMatch|TestConfigGray' -count=1` 通过。
- `go test ./pkg/config ./pkg/cache/config ./pkg/cache/gray ./apis/pkg/types/config -count=1` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `go test ./...` 通过。

Review：

- 这次新增测试不是只测灰度选择器，而是直接锁定客户端发现响应，能覆盖正常发布和多个灰度发布共存时每类客户端拿到的最终配置正文。
- 客户端读取接口之前计算出了配置对象但没有写入响应，同时成功 `Code` 也是默认 0；Nacos/HTTP 客户端会从 `queryResp.GetFile().GetContent()` 取内容，因此这里必须作为链路级测试长期保留。

## 产品 Logo 与名称替换

- [x] 确认当前 Logo 资产和使用位置
- [x] 写入静态检查，覆盖 SVG 名称、内嵌产品 Logo、侧边栏底部名称
- [x] 替换展开态与折叠态 SVG，产品名称改为 `Lattice.Hub`
- [x] 调整侧边栏/登录页 Logo 尺寸与底部文案
- [x] 运行前端验证、构建，并用真实页面检查 Logo 展示
- [x] 记录 review

当前判断：

- 当前控制台 Logo 由 `assets-logo-full.svg` 和 `assets-t-logo.svg` 承载，登录页和侧边栏展开态复用 full svg，侧边栏折叠态复用 mini svg。
- 用户要求“使用产品Logo”且“一模一样”，手写 path 很难和提供的 PNG 完全一致；本轮采用 SVG 内嵌产品 PNG base64 的方式保证视觉一致，同时文字部分使用 SVG text 改成 `Lattice.Hub`。

修复：

- `assets-logo-full.svg` 改为产品图形 + `Lattice.Hub` 文案，图形部分直接内嵌用户提供的产品 PNG。
- `assets-t-logo.svg` 改为仅产品图形的 36px 折叠态 SVG。
- 侧边栏底部版本文案从 `Pole.IO ${version}` 改为 `Lattice.Hub ${version}`。
- 调整登录页与侧边栏 Logo 容器宽高，避免 `Lattice.Hub` 被裁切。
- 新增 `verify-brand-logo.mjs` 反回归检查。

验证：

- 旧 Logo 下 `cd console/web && node scripts/verify-brand-logo.mjs` 按预期失败于完整 Logo 仍是 `Pole.IO`。
- 修复后 `cd console/web && node scripts/verify-brand-logo.mjs && for script in scripts/verify-*.mjs; do node "$script"; done && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 重启后 8080 返回的入口 `/assets/index.a409eed8.js` 与 `console/web/dist/index.html` 一致，入口资源返回 `200`。
- Playwright 打开登录页，存在 `svg[aria-label="Lattice.Hub"]`，尺寸 `204x36`，图形为 `data:image/png;base64` 内嵌。
- Playwright 使用 `admin/admin123` 登录，展开侧边栏显示 `Lattice.Hub` 完整 Logo，底部文案包含 `Lattice.Hub`。
- 折叠侧边栏后，mini Logo 尺寸为 `36x36`，图形为同一内嵌产品 PNG。

Review：

- SVG 使用内嵌 PNG 是为了满足“和产品 Logo 一模一样”的视觉要求；如果后续拿到官方矢量源，可以再替换为纯 path 版本。

## 配置发布记录发布时间与操作人链接

- [x] 写入失败检查，覆盖 release API 审计字段、前端时间兜底和操作人链接
- [x] 修复后端发布记录列表/详情响应，补齐 `ctime/mtime/create_by/modify_by`
- [x] 修复前端发布记录时间展示，避免灰度行发布时间为空
- [x] 将发布人渲染为新标签打开的用户详情链接，缺少用户 id 时回退文本
- [x] 运行前后端验证，重启并用真实页面验收
- [x] 清理临时数据并记录 review/lesson

当前判断：

- 发布记录列表里灰度行没有发布时间，根因不是表格列不存在，而是列表接口组装 `ConfigFileRelease` 时漏了 `ctime/mtime`；前端只读 `createTime`，没有用 `modifyTime` 兜底。
- 发布人详情已有隐藏路由 `/auth/principals/userdetail?name=...&id=...`；发布记录只有操作人名称，需要前端按名称查询用户 id 后生成新标签链接。

修复：

- `ToConfiogFileReleaseApi` 补齐 `create_by/modify_by`，列表接口 `handleDescribeConfigFileReleases` 补齐 `ctime/mtime`。
- `config_release.ts` 归一发布时间时增加 `modifyTime/mtime` 兜底。
- `ReleaseTable` 新增 `OperatorLink`，按发布人名称调用用户列表解析用户 id；解析成功时使用新标签打开 `/auth/principals/userdetail?name=...&id=...`，解析失败时回退文本。

验证：

- 先写入失败检查，`go test ./apis/pkg/types/config -run TestConfigFileReleaseAPIIncludesAuditFields -count=1` 失败于 `create_by/modify_by` 为空；`node scripts/verify-config-group-detail-design.mjs` 失败于缺少 `OperatorLink`。
- 修复后 `go test ./apis/pkg/types/config ./pkg/config -run 'TestConfigFileReleaseAPIIncludesAuditFields|TestPublishConfigFileDoesNotBlockOnActiveGrayRelease|TestSelectMatchedGrayRelease' -count=1` 通过。
- `cd console/web && node scripts/verify-configuration-api-contract.mjs && node scripts/verify-config-group-detail-design.mjs && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `go test ./apis/pkg/types/config ./pkg/config -count=1` 通过。
- `go test ./...` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 重启后 8080 返回的入口 `/assets/index.7fe9cc13.js` 与 `console/web/dist/index.html` 一致，入口资源返回 `200`。
- 真实接口创建临时灰度发布后，版本列表返回 `ctime/mtime/create_by/modify_by`：`2026-07-05 03:17:14 / admin`。
- Playwright 使用 `admin/admin123` 打开临时配置文件 `default/codex-operator-0704191714/operator.yaml`，发布记录 / 灰度发布行显示发布时间 `2026-07-05 03:17:14`；发布人 `admin` 链接 href 为 `/auth/principals/userdetail?name=admin&id=49dba3c69bca4b668903901d85c61528`，`target="_blank"`。
- 临时配置文件和分组已清理，按分组名查询 `remaining=0`。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg && git diff --check` 通过。

Review：

- 发布记录这类审计信息不能只在前端列上展示字段名，后端列表/详情响应必须完整带出审计字段；前端再做兜底展示和可点击用户详情。

## 配置发布记录操作悬浮提示修复

- [x] 定位发布记录操作列的 Tooltip / Popconfirm 嵌套方式
- [x] 将图标操作收敛为稳定 DOM target 的 `ActionButton` / `ConfirmActionButton`
- [x] 补充 `aria-label`、`title` 和静态反回归检查，确保悬浮提示与可访问名称不丢失
- [x] 运行前端验证、lint、构建，并重新拉起 all-mode 服务
- [x] 使用真实页面 hover 验证发布记录操作提示可见
- [x] 清理临时配置数据并记录 review

当前判断：

- 用户截图对应配置文件详情的发布记录操作列，图标按钮没有文字说明，必须依赖悬浮提示表达“提交为正式草稿 / 删除灰度 / 删除”等动作。
- 原实现里部分按钮是 `Tooltip` 直接包 `Popconfirm`，触发目标不是稳定 DOM；在表格行和确认气泡组合下，悬浮提示容易不出现。
- 修复应收敛在发布记录操作按钮层，不改发布流程语义。

修复：

- 新增 `ActionButton` 和 `ConfirmActionButton`，所有发布记录行操作统一走这两个组件。
- `Tooltip` 挂到稳定的 `span.actionTooltipTarget`，确认类动作仍由 `Popconfirm` 包住按钮，避免两类浮层争抢触发目标。
- 每个图标按钮补 `aria-label` 与 `title`，即使 Tooltip portal 受影响，也有浏览器原生提示和可访问名称兜底。
- 静态验证脚本增加约束，后续不能把发布记录操作退回直接裸写 `Tooltip/Popconfirm/Button` 的不稳定组合。

验证：

- `cd console/web && node scripts/verify-config-group-detail-design.mjs && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 重启后 8080 返回的入口 `/assets/index.52ec3e64.js` 与 `console/web/dist/index.html` 一致，入口资源返回 `200`。
- Playwright 使用 `admin/admin123` 登录真实页面，打开临时配置文件 `default/codex-tooltip-0704190148/tooltip.yaml`，切到发布记录 / 灰度发布；hover “提交为正式草稿”和“删除灰度”两个图标按钮后，页面分别出现对应 Tooltip 文案。
- 同一页面 DOM 验证两个操作按钮均有 `aria-label` 和 `title`，值分别为 `提交为正式草稿`、`删除灰度`。
- 临时配置文件和分组已清理，按分组名查询 `remaining=0`。

Review：

- 行操作图标必须同时满足视觉提示和可访问名称；后续发布记录新增操作时应继续复用 `ActionButton` / `ConfirmActionButton`，不要在行内重新手写裸按钮组合。

## 配置发布灰度规则复用治理通用控件

- [x] 核对治理侧 `TrafficMatchConditionEditor` 和配置发布 `betaLabels` 的数据结构差异
- [x] 新增配置灰度规则适配组件，复用治理通用匹配条件控件
- [x] 替换配置发布抽屉中的旧 `ClientLabelInput` 灰度规则编辑器
- [x] 更新静态反回归脚本，禁止配置灰度规则继续直接使用旧私有控件
- [x] 运行前端验证、lint、构建，并更新 lessons / review

当前判断：

- 用户截图对应配置发布抽屉里的灰度规则，当前由 `components/ClientLabelInput` 私有实现渲染。
- 治理侧已有通用控件 `pages/Governance/shared/TrafficMatchConditionEditor`，支持匹配关系、参数类型、参数键、匹配类型、匹配值和添加/删除行；配置灰度应复用它，而不是继续维护私有行编辑。
- 配置灰度的后端契约仍是 `betaLabels: MatcheLabel[]`，因此需要在配置侧做 `MatcheLabel <-> TrafficMatchConditionRow` 适配。

修复：

- `TrafficMatchConditionEditor` 增加 `showParamKey` 可选参数，默认保持治理规则现有五列形态；配置灰度传 `false`，隐藏内置客户端标签不需要的参数键列。
- 新增 `GrayRuleEditor`，复用治理侧 `TrafficMatchConditionEditor`，并提供 `grayRowsToBetaLabels` / `betaLabelsToGrayRows` 转换。
- `PublishForm` 的灰度发布态从 `ClientLabelInput` 切换到 `GrayRuleEditor`，提交前校验至少存在一条有效灰度规则。
- `verify-config-group-detail-design.mjs` 增加约束：配置发布抽屉不能直接使用旧 `ClientLabelInput`，灰度规则适配组件必须复用 `TrafficMatchConditionEditor`。

验证：

- `cd console/web && node scripts/verify-config-group-detail-design.mjs && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 重启后 8080 返回的入口 `/assets/index.70e1e4ca.js` 与 `console/web/dist/index.html` 一致，入口资源返回 `200`。
- Playwright 使用 `admin/admin123` 打开临时配置文件 `default/codex-gray-control-0704184856/gray-control.yaml`，发布抽屉切换灰度发布后展示通用控件：`AND/OR`、`参数类型`、`匹配类型`、`匹配值`、`添加灰度规则`。
- 真实点击确认发布成功，POST `/config/v1/files/release` 返回 `200`，请求体包含 `beta_labels:[{key:"CLIENT_IP", value:{type:"EXACT", value_type:"TEXT", value:"10.42."}}]`。
- 临时配置分组和文件已清理，按名称查询 `amount=0`。

Review：

- 用户判断是对的，这里应该复用治理通用匹配条件控件；旧 `ClientLabelInput` 仍可留给尚未迁移的发布场景，但配置灰度发布不再直接使用它。
- 配置灰度当前只暴露客户端 IP/ID/语言这类内置标签，因此隐藏参数键列；如果后续 spec 支持自定义客户端标签，需要再打开参数键列或补自定义标签选项。

## 配置中心页面打不开与 spec 适配确认

- [x] 复现 8080 页面打不开现象，确认是后端接口、console 网关、静态资源还是浏览器缓存问题
- [x] 重建并重新拉起 all-mode，确保当前 `dist/index.html` 引用的入口 JS 在 8080 可访问
- [x] 用真实浏览器打开配置中心深链，检查页面非空且没有入口资源 404
- [x] 明确方案 B 对 `github.com/pole-io/specification` 的适配范围，区分当前仓库临时兼容与长期协议契约
- [x] 补充 review、验证结果和剩余风险

当前判断：

- 8080 首页和配置中心深链都能返回 HTML，但浏览器控制台出现 `/assets/index.2b33ff44.js` 404；当前 `console/web/dist/index.html` 实际引用的是 `/assets/index.eccd0215.js`，说明运行中的页面或浏览器缓存拿到了旧入口 HTML。
- 这类问题不能只用 `curl /` 判断页面可用，必须验证 `index.html` 中引用的入口 JS 在同一 8080 服务下返回 200，并用浏览器打开深链确认应用真正渲染。
- 方案 B 已在本仓库内补了后端 HTTP route 和前端调用，但长期协议契约仍需要同步更新 `github.com/pole-io/specification`；否则跨仓库生成代码、OpenAPI 文档和其它客户端不会知道灰度转正式草稿、多灰度 active 和按 releaseName 停止灰度的语义。

验证：

- 重启前复现 8080 返回旧入口 `/assets/index.2b33ff44.js`，而磁盘 `dist/index.html` 已是新入口，确认页面打不开根因是运行中 console 静态资源与构建产物 hash 不一致。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 重启后 `curl http://127.0.0.1:8080/` 返回的入口 JS 与 `console/web/dist/index.html` 一致，入口资源返回 `200`。
- Playwright fresh session 打开配置中心深链不再空白；未登录时正常跳转登录页。

Review：

- spec 需要适配更新，已归档到 [[adr-config-gray-release-spec-contract]]：必须补灰度转正式草稿操作、多 active gray 语义、停止单个灰度语义，以及真正参与客户端命中的 gray priority 字段。
- 当前仓库内的灰度优先级前端字段属于预留入口；在 specification 未发版前，后端会忽略未知字段，实际命中仍按当前 version/mtime 兜底。

## 配置分组详情设计交接适配

- [x] 对照附件交接文档盘点当前配置分组详情页，确认主结构、交互和发布模型差距
- [x] 重构配置分组详情页为文件树 + 中央白色业务画布，不恢复顶部摘要卡、右侧预览栏或悬浮操作条
- [x] 调整配置编辑页签：文件上下文编号分段、Monaco 正文、全屏编辑、保存草稿和发布配置操作
- [x] 调整发布配置抽屉：版本对比、版本信息、发布范围上下布局，灰度态展示规则和优先级入口
- [x] 调整发布记录页签：正式发布、灰度发布、正式草稿、历史记录内部 Tabs，每类独立分页和状态操作
- [x] 调整订阅查询页签：展示客户端 ID/IP/类型/标签/监听版本/最近拉取/操作，并按当前发布记录解析监听版本展示
- [x] 补静态反回归脚本覆盖附件要求的旧关键词禁用、内部 Tabs、监听版本和多灰度关键文案
- [x] 运行前端静态验证、lint、构建，重启后用真实账号打开配置分组详情验收

当前判断：

- 当前页面是 `Row/Col + Tree + Tabs`，发布记录是单表，配置编辑使用 `StickyTool` 悬浮操作，和交接文档要求的工作台结构差距较大。
- 当前后端已支持多 active gray、灰度转正式草稿和正式发布不阻塞灰度；但“灰度优先级”没有 specification 字段，当前实际命中仍按 version/mtime。前端本轮可以展示优先级入口和设计结构，若要真正影响客户端解析，需要后续先改 specification。

修复：

- `Files/index.tsx` 改为灰色工作区中的白色业务画布，左侧固定文件树资源浏览器，右侧为受控业务 Tabs；文件树按路径分层并显示状态 pill。
- `FileView` 移除右侧悬浮 `StickyTool`，改为“文件上下文 / 配置正文”两个编号分段；查看态显示 Monaco 正文，编辑态提供保存草稿、撤销和发布配置。
- `PublishForm` 抽屉固定为版本对比和发布信息两步，第二步按“版本信息 / 发布范围”上下布局；全量态隐藏灰度规则，灰度态展示灰度规则和灰度优先级入口。
- `ReleaseTable` 改为正式发布、灰度发布、正式草稿、历史记录内部 Tabs，每类独立分页 6 条；灰度发布支持提交为正式草稿和删除灰度，当前全量删除有强风险提示。
- `SubscribeTable` 增加客户端 ID、客户端 IP、类型、标签、监听版本、最近拉取、操作列，并按当前发布记录标识命中灰度、当前全量和无可用版本状态。
- 新增 `scripts/verify-config-group-detail-design.mjs`，固化附件中的反回归关键词、内部 Tabs、多灰度文案、监听版本列和全屏编辑 class 约束。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过，包含新增 `config group detail design checks passed`。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `go test ./...` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 重启后 8080 返回的入口 `/assets/index.752d1cb6.js` 与 `console/web/dist/index.html` 一致，入口资源返回 `200`。
- `cd console/web && node scripts/smoke-configuration-flow.mjs` 通过，输出 `configuration flow smoke passed: default/codex-flow-0704183538/app.yaml`。
- Playwright 使用 `admin/admin123` 登录，打开 `default/codex-detail-0704183233/services/app.yaml` 配置分组详情；配置编辑页签显示文件上下文、配置正文和发布操作，发布记录显示正式发布/灰度发布/正式草稿/历史记录内部 Tabs，灰度发布行显示 `env=gray / P100`，订阅查询显示监听版本列；浏览器 console errors/warnings 均为 `0`。
- 临时配置分组 `codex-detail-0704183233` 和文件已通过真实接口清理，按名称查询 `amount=0`。

Review：

- 页面 Tabs 必须受控设置 active value；未受控时真实页面只显示 Tabs 头不挂载内容，本轮已修复。
- 当前订阅接口没有返回客户端标签和最近拉取时间，页面先保留列和明确占位；后续应随 specification / 后端接口补字段后在 service 层归一。
- 当前 gray priority 只是 Console 入口和展示预留，不等同于后端已按优先级选择灰度版本；真正生效需要先完成 [[adr-config-gray-release-spec-contract]]。

## 配置灰度多版本与转正式草稿

- [x] 写入方案 B 实施计划，明确后端发布模型、灰度匹配、前端操作和验证边界
- [x] 补后端 RED 测试：有活跃灰度时仍允许正式发布
- [x] 补后端 RED 测试：同一配置文件允许多个活跃灰度版本并按优先级/更新时间选择命中版本
- [x] 补后端 RED 测试：灰度版本可提交为正式草稿，不自动正式发布且不停止灰度
- [x] 实现灰度 release 独立 key、灰度规则存储、缓存读取和客户端匹配选择
- [x] 实现停止单个灰度、停止全部灰度与灰度转正式草稿接口
- [x] 适配前端发布记录操作：灰度发布不阻塞正式发布，灰度版本支持转正式草稿和停止
- [x] 扩展配置中心静态契约和接口烟测，覆盖 normal + 多 gray + promote-to-draft
- [x] 运行 Go 测试、前端 lint/build、真实接口烟测、context-kg lint 和 diff 检查
- [x] 使用 `admin/admin123` 在真实页面验证灰度发布、正式发布、转草稿和发布记录展示
- [x] 清理临时数据并记录 review、验证结果和剩余风险

当前判断：

- 采用方案 B：正式发布与灰度发布分离；同一配置文件支持多个 active gray release；灰度验证通过后可以把某个灰度内容写回正式草稿，再由用户走 normal 发布。
- 现状里 `handlePublishConfigFile` 在发布前查询 `GetConfigFileBetaReleaseTx`，只要存在活跃灰度就返回冲突，因此 active gray 会阻塞 normal release 和新的 gray release。
- 现状灰度规则 key 是 `config@namespace@group@fileName`，不含 `release_name`；active cache key 也只到 `file + release_type`，因此模型天然限制为每个文件最多一个活跃灰度。
- 多灰度命中冲突本轮不改 specification proto，不新增显式优先级字段；客户端命中多个灰度时按 version 倒序、mtime 倒序兜底，后续如果要运营可配置优先级，需要先补协议字段。

修复：

- 配置发布模型拆分 normal 与 gray：normal 发布仍保持同一配置文件只有一个 active normal，gray 发布不再互斥已有 active gray，也不再阻塞 normal 发布。
- 灰度 active key 和 gray resource key 增加 releaseName 维度，支持同一配置文件多个 active gray release 与多套灰度规则共存；旧无名灰度 key 保留兼容。
- store/cache/client/watch 链路增加 active gray 列表读取，客户端按命中灰度规则的 release 中 version/mtime 最新版本返回，文件状态和订阅视图也按多灰度判断。
- 新增 `/config/v1/files/releases/promote-gray`，将指定 active gray 内容提交回配置文件正式草稿，不自动发布 normal，也不停止灰度。
- `StopGrayConfigFileRelease` 支持带 releaseName 停止单个灰度；不带 releaseName 时兼容旧语义，停止该配置文件全部 active gray。
- 前端发布记录对 active gray 展示“提交为正式草稿”和“停止灰度”操作；删除/停止/promote 都带 releaseType，避免同名 normal/gray 误操作。
- 扩展 `scripts/smoke-configuration-flow.mjs`，覆盖 normal 发布、两个 gray 并存、gray active 时 normal 发布、promote-to-draft、promote 后 normal 发布、停止单个 gray 且另一个 gray 保持 active。

验证：

- `go test ./apis/pkg/types/config ./plugin/store/mysql ./pkg/cache/config ./pkg/config -run 'TestConfigGrayRelease|TestCreateConfigFileGrayReleaseKeepsExistingGrayActive|TestConfigFileCacheKeepsMultipleActiveGrayReleases|TestSelectMatchedGrayRelease|TestPublishConfigFileDoesNotBlockOnActiveGrayRelease' -count=1` 通过。
- `go test ./apis/pkg/types/config ./apis/store ./plugin/store/mysql ./pkg/cache/config ./pkg/config/... -count=1` 通过。
- `go test ./...` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功，日志到 `finish starting server`。
- `cd console/web && node scripts/smoke-configuration-flow.mjs` 通过，输出 `configuration flow smoke passed: default/codex-flow-0704174837/app.yaml`。
- Playwright 使用 `admin/admin123` 登录真实页面，进入 `default/codex-ui-gray-0704175104/app.yaml` 发布记录；页面显示两个 active gray 行，每行有提交草稿、停止灰度、删除三个操作。
- 页面点击 `gray-a-0704175104` 的“提交为正式草稿”成功，接口确认文件草稿内容为 `a: gray-a\n`，active normal 未自动变化。
- 页面点击 `gray-a-0704175104` 的“停止灰度”成功，接口确认 `gray-a` inactive、`gray-b` 仍 active；临时配置文件和配置分组已清理，按分组查询 `amount=0`。
- 浏览器控制台仅有登录页 autocomplete 提示和 Redux log，无配置中心操作 error/warning。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。

Review：

- 方案 B 的关键点是 active 维度从“文件 + releaseType”扩展为“文件 + gray releaseName”，否则无论 store 还是 cache 都会把多灰度压扁成一个版本。
- promote-to-draft 必须只写配置文件草稿，不能顺手发布 normal 或停止 gray；真实烟测已覆盖这个行为。
- 本轮未引入显式灰度优先级字段，因为当前 specification `ConfigFileRelease` 没有对应字段；如后续需要运营可配置优先级，应先补 proto/前端字段，再把选择函数从 version/mtime 切到 priority/mtime。

## 配置中心全流程可用化修复

- [x] 复现配置文件创建失败，记录真实请求、响应和控制台错误
- [x] 编写接口级烟测脚本覆盖配置分组、配置文件创建、详情、更新、发布、版本、回滚、删除和清理
- [x] 按烟测失败点核对后端 handler、业务层、前端 service 与页面提交数据
- [x] 先补回归验证，再修复配置文件创建及后续流程的契约错位
- [x] 运行 Go 测试、前端配置中心静态脚本、lint、构建、context-kg lint 和 diff 检查
- [x] 使用 `admin/admin123` 真实页面走配置分组到配置文件创建/发布流程
- [x] 清理临时数据并记录 review、验证结果和剩余风险

当前判断：

- 本轮以当前前端已暴露的配置中心能力为边界，覆盖配置分组、配置文件、发布记录、订阅查询和加密算法；不新增未设计的导入导出、操作历史或客户端视角订阅入口。
- 先用真实接口脚本把流程打通，避免只修页面上的单个表单字段后遗漏发布、删除或回滚链路。
- 复现配置文件创建失败时，前端请求体为 `namespace/group/format/content/labels`，缺失 `name/comment`；后端返回 `400103 invalid parameter`，根因是两步表单进入第二步后第一步字段卸载，提交时再读 `form.getFieldValue('name')` 得到空值。
- 发布配置失败的接口根因是前端按 `fileName/releaseDescription/releaseType/betaLabels` 提交，当前后端 proto JSON 契约要求 `file_name/release_description/release_type/beta_labels`。
- 直达或刷新配置文件页时，Redux 中可能没有当前配置分组，必须按 URL `namespace/group` 补拉分组信息；不能让新建入口依赖上一次从分组列表进入时残留的 `editGroup`。

修复：

- `FileCreator` 在第一步离开前缓存 `name/comment/encrypted/encryptAlgo/tags`，第二步提交使用缓存的元信息和当前内容组装创建请求，避免字段卸载导致丢参。
- `config_release` service 在边界层把发布、回滚、删除写请求从前端 camelCase 映射为后端 snake_case，并把响应中的 `file_name/release_type/release_description/beta_labels` 归一回前端字段。
- 配置文件页按当前 URL 自动查询配置分组并写回 `editGroup`，同时只使用与 URL 匹配的分组权限，修复直达/刷新后新建入口不可用。
- `CodeDiffEditor` 为 original/modified 设置不同稳定 model path，并在卸载时保留当前 model，修复发布抽屉从版本对比切到发布信息时的 Monaco `TextModel got disposed before DiffEditorWidget model got reset` 错误。
- 新增 `scripts/smoke-configuration-flow.mjs`，用 `admin/admin123` 默认凭据覆盖配置分组、配置文件、详情、更新、发布 v1/v2、发布列表、版本列表、订阅查询、回滚、删除发布、删除文件、删除分组和清理校验。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过，包含配置中心契约、入口、TDesign 升级、生产 React 构建和既有页面静态校验。
- `cd console/web && node scripts/smoke-configuration-flow.mjs` 通过，输出 `configuration flow smoke passed: default/codex-flow-0704034412/app.yaml`。
- `go test ./pkg/common/api/v1 ./pkg/config/... -count=1` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 多次通过，tmux 日志显示 `finish starting server`。
- Playwright 使用 `admin/admin123` 真实登录，直达 `default/codex-ui-flow-0704033043` 配置文件页；页面自动请求 `/config/v1/groups?namespace=default&offset=0&limit=1&name=codex-ui-flow-0704033043`，新建入口可用。
- 同一页面创建 `ui-flow.yaml` 成功，请求体包含 `namespace/group/name/comment/format/content/encrypted/encryptAlgo/labels`，配置树显示文件。
- 同一页面发布 `v1` 成功，请求体包含 `file_name/release_description/release_type`，接口返回 `200000`；切换发布步骤后浏览器 console error 为 `0`。
- 临时配置分组 `codex-ui-flow-0704033043`、`codex-ui-flow-0704111959`、`codex-flow-0704110556` 已清理，按前缀查询剩余 `codex-` 分组为 `none`。

Review：

- 配置文件创建失败不是后端创建接口不可用，而是前端分步表单把第一步字段卸载后仍从 Form 实例读取，导致 `name` 为空。
- 发布、回滚、删除要以当前后端生成 proto JSON 字段为准，页面继续使用 camelCase，转换收敛在 `config_release` service。
- 配置文件页不能假设一定从配置分组列表进入；刷新、直达、复制链接都必须通过 URL 参数恢复分组上下文和权限。
- 页面发布流程仍可捕获到一次 `TDesign Tree Warn: Duplicated value: ui-flow.yaml`，但后端文件查询为 `amount=1`，配置创建和发布流程未受影响；后续若继续收敛控制台 warning，可单独排查 TDesign Tree 数据/状态提示。

## 配置文件加密算法下拉为空修复

- [x] 复现并抓取创建配置文件时加密算法接口请求与响应
- [x] 核对后端加密算法接口、前端 service 和 Redux slice 的数据形态
- [x] 补配置中心静态回归验证，约束加密算法响应必须适配到 Select options
- [x] 实现最小修复并跑前端验证、lint、构建
- [x] 用真实账号打开创建配置文件抽屉，验证加密算法下拉有可选项
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 截图显示加密算法 Select 已打开但为空；优先沿 `listConfigFileCryptoAlgos -> service -> 后端路由` 查真实响应形态，不先做 UI 假数据。
- 根因是后端 `GetAllConfigEncryptAlgorithms` 的简化实现只返回 `code/info`，`data:null`；前端拿不到 `algorithms`，因此 Select 空态显示“暂无数据”。

修复：

- `NewConfigEncryptAlgorithmResponse` 改为用 protobuf `Struct` 承载 `{ algorithms: [...] }`，并写入 `Response.data` 的 `Any`。
- `GetAllConfigEncryptAlgorithms` 从 `cryptoManager.GetCryptoAlgoNames()` 读取真实注册算法；本地 all-mode 当前配置只启用 `AES`。
- `describeEncryptAlgo` 在 service 边界兼容后端 `Struct.value.algorithms` 和直接 `algorithms` 两种形态，Redux 与组件继续消费稳定的 `string[]`。
- `scripts/verify-configuration-api-contract.mjs` 增加加密算法响应归一约束，避免页面层感知后端 `Any/Struct` 细节。

验证：

- `go test ./pkg/common/api/v1 -run 'TestConfigEncryptAlgorithmResponseIncludesAlgorithms|TestConfigSingleObjectResponsesIncludeData' -count=1` 修复前失败，修复后通过。
- `go test ./pkg/common/api/v1 ./pkg/config/... -count=1` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `go test ./...` 通过。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 真实登录 `admin/admin123` 后请求 `/config/v1/files/encrypt/algorithms` 返回 `200000`，响应 `data.value.algorithms: ["AES"]`。
- Playwright 使用真实账号进入临时配置分组 `default/codex-algo-drawer-0704094254`，打开创建配置文件抽屉，启用 `配置加密` 后打开 `加密算法类型` 下拉；下拉显示 `AES`，浏览器 console error/warning 均为 `0`。
- 通过真实删除接口清理临时配置分组，随后按名称查询返回 `amount:0`。

Review：

- 这次不应在前端硬编码算法选项；可选算法来自后端 crypto 插件配置，当前本地只启用 `AES` 是配置结果，不是 UI 限制。
- specification 当前没有专门的加密算法响应 proto，使用标准 `Struct` 是最小可用承载方式；前端兼容逻辑收敛在 service 层，后续若 specification 补正式 message，只需要调整该边界。

## 配置文件创建抽屉分组信息回显修复

- [x] 复现并定位从配置分组进入配置文件创建时 namespace/group 不显示的问题
- [x] 补配置中心静态回归验证，约束创建抽屉必须回填 namespace/group
- [x] 实现最小修复并跑前端验证、lint、构建
- [x] 用真实账号打开配置分组、进入文件创建抽屉验证字段显示
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 文件创建页 URL query 已包含 `namespace/group`，`FileCreator` 也收到了 props；问题在抽屉 Form 没有把这两个外部值写入字段，disabled Input 只显示空占位。

修复：

- `FileCreator` 在抽屉打开或 `namespace/group` 变化时调用 `form.setFieldsValue({ namespace, group })`，让禁用输入框显示当前选中的配置分组上下文。
- 移除 `FileCreator` 中未使用的 `editFile` 解构和提交时的 `console.log('newData')` 调试输出。
- `scripts/verify-configuration-api-contract.mjs` 增加配置文件创建抽屉回填 `namespace/group` 的静态约束。

验证：

- `cd console/web && node scripts/verify-configuration-api-contract.mjs` 修复前失败，修复后通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- 使用真实账号 `admin/admin123` 登录，创建临时配置分组 `default/codex-file-drawer-07040927`，从配置分组列表点击进入文件页后打开创建抽屉；抽屉中 `命名空间` 显示 `default`，`配置组` 显示 `codex-file-drawer-07040927`，浏览器 console error/warning 均为 `0`。
- 通过真实删除接口清理 `codex-file-drawer-07040927`，随后按名称查询返回 `amount:0`。

Review：

- 修复点应放在表单初始化边界，而不是改路由或提交 payload；提交路径之前已经拿到正确 props，本次缺陷只是 UI Form 字段未同步。

## 配置分组新建失败修复

- [x] 用 `admin/admin123` 真实复现新建配置分组失败，记录请求、响应和控制台
- [x] 沿前端表单、service、后端 handler/store 定位契约错位
- [x] 先补失败验证，覆盖新建配置分组请求必须符合后端当前定义
- [x] 实现最小修复，并跑前端静态验证、lint、构建和后端针对性测试
- [x] 用真实账号完成新建配置分组烟测，并清理测试数据
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 配置分组列表可用不代表创建可用；本轮以真实创建链路为准，不用 mock 判定。
- 新建配置分组失败包含三层根因：空标签时前端 `reduce(undefined)`，创建请求带 `id: 0` 与后端 `string id` 契约冲突，创建落库后 MySQL `DATETIME` 与 Go 本地时区比较导致缓存增量同步漏数据。

修复：

- `ConfigGroupEditor` 将空 `group_labels` 按空数组处理，提交时不再构造 `id: 0`，避免创建请求在前端抛错或被后端 JSON 解码拒绝。
- `config_group.ts` 将配置分组 `id` 改为后端真实的 `string` 契约；配置中心契约验证新增空标签与 `id` 类型约束。
- `plugin/store/mysql/config_file_group.go` 的增量查询从 `WHERE mtime >= ?` 改为 `WHERE UNIX_TIMESTAMP(mtime) >= ?`，参数传 `mtime.Unix()`，避免 MySQL `sysdate()` 时区与 Go `loc=Local` 绑定参数不一致时漏同步。
- 新增 `plugin/store/mysql/config_file_group_test.go`，约束配置分组增量查询必须使用 Unix timestamp 比较。

验证：

- 修复前，Playwright 使用 `admin/admin123` 新建不填标签配置分组，控制台报 `TypeError: Cannot read properties of undefined (reading 'reduce')`，且没有 POST。
- 修复空标签后，真实 POST 发出但返回 `400001 request decode failed: json: cannot unmarshal number into Go value of type string`，请求体包含 `id:0`。
- 修复 `id` 后，真实 POST 返回 `200000` 且 DB 落库，但列表仍为空；DB 记录 `mtime=2026-07-03 19:47:00`，服务本地时间为 `2026-07-04 03:47`，确认缓存增量查询存在时区比较漏读。
- `go test ./plugin/store/mysql -run TestConfigFileGroupIncrementalQueryUsesUnixTimestamp -count=1` 修复前失败，修复后通过。
- `go test ./plugin/store/mysql ./pkg/cache/config ./pkg/config/... -count=1` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- Playwright 使用 `admin/admin123` 真实创建 `codex-create-smoke-0704035605`：POST `/config/v1/groups` 返回 `200`，请求体不含 `id`，最终 GET `/config/v1/groups?offset=0&limit=10&namespace=` 返回 `amount:1` 且包含该分组，浏览器 console error/warning 均为 `0`。
- 通过真实删除接口 `POST /config/v1/groups/delette` 清理 `codex-create-smoke-0704035605`，响应 `200000`；随后按名称查询返回 `amount:0`，DB 中该记录 `flag=1`。

Review：

- 这次问题说明“配置中心可用”必须验完整写链路；只打开页面和读空列表会漏掉表单空值、前端/后端类型契约、缓存增量同步三类问题。
- 后端创建响应中 `data.id` 仍返回空字符串，因为 `CreateConfigFileGroup` 返回体使用了 `saveData.Id` 而不是 `ret.Id`；列表可见性不依赖该字段，本轮未扩大修复范围。

## 配置中心 TDesign key spread warning 修复

- [x] 定位配置中心页面运行时 `key` 被 spread 到 JSX 的真实来源
- [x] 升级 `tdesign-react` 到 `1.18.0`，并核对锁文件变化
- [x] 补充可重复验证，约束构建产物不能继续包含 React 开发 JSX runtime
- [x] 运行前端静态验证、lint、构建和浏览器控制台验证
- [x] 记录 review、验证结果和剩余风险

当前判断：

- warning 来自 React 对第三方组件内部 `{...props}` 携带 `key` 的运行时检查，堆栈落在 TDesign 打包产物，不应在配置中心页面层做无效绕行。
- 本轮以消除配置中心页面可见控制台 warning 为目标；允许升级 `tdesign-react` 到 `1.18.0`，同时保留已验证的 vendor 分包策略。

修复：

- `tdesign-react` 升级到 `^1.18.0`，锁文件解析到 `1.18.0`；随版本依赖同步升级 `sortablejs` 到 `1.15.7`，移除旧版间接依赖 `tinycolor2`。
- `scripts/run-vite-build.mjs` 同时设置 `NODE_ENV=production` 与 `VITE_USER_NODE_ENV=production`；Vite 2 自定义 `release/test/site` mode 下必须依赖 `VITE_USER_NODE_ENV` 才会让 `@vitejs/plugin-react` 走 production JSX transform。
- 新增 `scripts/verify-production-react-build.mjs`，检查构建脚本和 dist 产物不能含 `react-jsx-dev-runtime.development`、`jsxDEV(` 调用或 React key-spread warning 文案。
- 新增 `scripts/verify-tdesign-upgrade.mjs`，约束 `package.json` 与 `package-lock.json` 不能回退 `tdesign-react@1.18.0`。
- 更新 `vite.config.js` 中 `tdesignSharedDepChunks` 到 TDesign 1.18.0 的真实 shared `_chunks` 集合；`verify-vite-manual-chunks.mjs` 改为从当前 `node_modules/tdesign-react/es` 追踪 shared 目录依赖，避免内部 `dep-*.js` hash 变更后再次出现 `tdesign-shared -> tdesign-data-form` 循环。

验证：

- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过，无 React key-spread warning、无 React dev runtime、无大 chunk warning。
- `cd console/web && CHECK_DIST=1 node scripts/verify-production-react-build.mjs` 通过。
- `cd console/web && rg "A props object containing|react-jsx-dev-runtime\\.development|\\.jsxDEV\\(|\\bjsxDEV\\(" dist/assets -g '*.js'` 无命中。
- `cd console/web && rg "tdesign-data-form" dist/assets/tdesign-shared.*.js` 无命中，确认 shared chunk 不再反向依赖 data-form。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl http://127.0.0.1:8080/configuration/group` 返回 `200`，页面引用当前 `tdesign-shared.bcbb542a.js` 与 `tdesign-data-form.49d39d7c.js`。
- Playwright 注入临时登录态并 mock `/config/v1/groups` 空列表后打开 `/configuration/group`，配置分组 Table 与分页 Select 渲染完成；`playwright-cli console warning` 与 `console error` 均为 `0`。
- Playwright 清理 mock 与本地登录态后，使用 `admin/admin123` 真实登录成功；打开 `/configuration/group?check=real-admin`，真实 `/config/v1/groups?offset=0&limit=10&namespace=` 返回 `200`，响应 `{code:200000, amount:0, size:0, data:[]}`；配置分组 Table 与分页 Select 渲染完成，`playwright-cli console warning` 与 `console error` 均为 `0`。

Review：

- 升级 TDesign 后，单靠版本变化不能消除该 warning；`tdesign-react@1.18.0` 内部仍存在 `React.createElement(... { key, ... })` 形态。真正让 8080 页面不再报 warning 的关键是 release/test/site 构建必须使用 React production JSX transform。
- 修复过程中发现并消除了一个升级后的分包循环：旧的 TDesign 1.12.2 `_chunks` 白名单不适配 1.18.0，会导致 `tdesign-shared` 反向 import `tdesign-data-form`，进而触发 `dayjs.extend` 初始化错误。
- 页面提示的默认 `pole/pole123` 不适用于当前本地库；用户提供的 `admin/admin123` 已完成真实登录和真实配置分组列表接口验证，无需再依赖 mock 判断配置分组页是否可渲染。

## 配置中心可用化适配

- [x] 补配置中心接口契约验证，先让现有错路由和字段适配缺口暴露出来
- [x] 修复后端配置中心单对象响应未写入 `data` 的实现缺口
- [x] 在前端配置中心 service 层归一后端响应、请求字段和真实路由
- [x] 跑前端静态验证、lint、构建和后端针对性测试
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 本轮目标是让已放开的配置中心入口真正可用，优先修复配置分组、配置文件、发布、订阅者、加密算法这些已有页面会调用的能力。
- 页面层不直接感知后端 `Any`、`labels`、`ctime` 等细节；这些差异统一收敛在 `console/web/src/services/config_*.ts`。
- `createandpub`、导入导出、操作历史、客户端视角订阅暂不新增页面入口，避免能力放开超过当前页面设计。

修复：

- 后端 `NewConfigGroupResponse`、`NewConfigFileResponse`、`NewConfigFileReleaseResponse`、`NewConfigFileTemplateResponse`、`NewConfigClientResponse` 统一把 proto message 写入 `Response.data` 的 `Any`，使配置文件详情、发布详情等单对象接口能返回真实数据。
- 前端新增 `scripts/verify-configuration-api-contract.mjs`，静态约束配置中心 service 必须使用后端真实路由和归一化边界。
- 配置组 service 适配 `/config/v1/groups/delette`，查询时把前端 `group` 搜索项映射到后端 `name` 参数，列表响应归一 `ctime/mtime -> createTime/modifyTime`。
- 配置文件 service 归一 `labels -> tags`、`ctime/mtime/rtime -> createTime/modifyTime/releaseTime`，写请求将 `tags -> labels`，并修正 `berif -> brief`。
- 发布 service 增加 `BaseURL.CONFIG_RELEASES = /config/v1/files/releases`，列表、回滚、删除分别走 `/files/releases`、`/files/releases/rollback`、`/files/releases/delete`；发布/详情继续走 `/files/release`；灰度标签在 `value_type` 与 proto JSON `valueType` 间双向适配。
- 配置组页面删除时改为传 `namespace/name`，不再只传后端不会使用的 `id`。

验证：

- `go test ./pkg/common/api/v1 -run TestConfigSingleObjectResponsesIncludeData -count=1` 修复前失败，修复后通过。
- `cd console/web && node scripts/verify-configuration-api-contract.mjs` 修复前失败，修复后通过。
- `go test ./pkg/common/api/v1 ./pkg/config/... -count=1` 通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过；`rg "Warning|warning|larger than|localstorage-file|Browserslist|error" /tmp/pole-config-build-test.log` 无命中。
- `go test ./...` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功重建并启动 all-mode，日志到 `finish starting server`。
- `curl http://127.0.0.1:8080/` 与 `curl http://127.0.0.1:8080/configuration/group` 均返回 `200`。
- 未登录直连 `8090/config/v1/groups`、`8090/config/v1/files/release`、`8090/config/v1/files/releases` 返回 `401001 access is not approved`，说明配置接口挂载存在且被 Console 鉴权拦截。
- Playwright 使用 `admin/admin123` 真实登录后打开 `/configuration/group?check=real-admin`，配置分组页真实请求 `/config/v1/groups?offset=0&limit=10&namespace=` 返回 `200`，响应体为 `{code:200000, amount:0, size:0, data:[]}`；页面展示配置分组表格空态与分页控件。

Review：

- 本轮让已暴露的配置中心页面入口、service 层和后端单对象响应契约对齐，避免页面拿不到详情、删除走错路由、发布回滚/删除走不存在路由。
- 当前本地 MySQL 已存在主账号 `admin`，有效登录密码为 `admin123`；本轮已完成带 token 的配置分组列表真实读取烟测。
- `/files/client/subscription` 仍由后端错误挂到文件视角 handler；本轮未放开客户端视角订阅页面入口。

## 配置中心接口契约适配评估

- [x] 核对后端配置中心 REST 路由、请求结构和响应结构
- [x] 核对前端配置分组、配置文件、发布和订阅 service 的当前假设
- [x] 对照页面使用点，列出必须适配、可兼容和不应暴露的能力边界
- [x] 形成前端接口适配方案、验证方式和剩余风险

当前判断：

- 本轮以当前后端接口定义为准，不按前端已有 TypeScript 类型反推接口。
- 前端应先把 `services/config_*.ts` 做成配置中心契约适配层，页面和 Redux slice 继续消费稳定 UI 模型，不把后端路由、`Any` 响应和字段别名扩散到组件。
- 批量查询接口当前通过 `BatchQueryResponse.data` 返回列表，现有 `unwrapResponse` 能解出 `data/amount/size`；单对象详情接口当前构造函数没有把对象写进 `Response.data`，这是后端实现缺口，前端无法单独补齐详情内容。

适配建议：

- 配置组：删除接口按当前后端路由适配到 `/config/v1/groups/delette`；分组模糊查询把前端 `group` 搜索项映射为后端 `name` 查询参数；列表响应保留 `fileCount`，并把 `ctime/mtime` 兼容归一到 `createTime/modifyTime`。
- 配置文件：列表、详情、创建、更新、删除继续走 `/config/v1/files`、`/detail`、`/search`、`/delete`；把后端 `labels` 与前端 `tags` 双向转换；`describeAllConfigFiles` 的 `berif` 拼写应收敛为后端可识别的 `brief` 或移除无效参数；详情只依赖 `namespace/group/name`，不依赖 `id`。
- 发布：单个生效/指定发布详情仍走 `/config/v1/files/release`；发布列表走 `/config/v1/files/releases`；版本列表走 `/config/v1/files/release/versions`；回滚走 `PUT /config/v1/files/releases/rollback`；删除走 `POST /config/v1/files/releases/delete`；如页面放开灰度停止，再接 `/config/v1/files/releases/stopbeta`。
- 订阅者：文件视角订阅者使用 `/config/v1/files/subscribers`；`/files/client/subscription` 当前路由到了同一个文件视角 handler，客户端视角订阅能力不应在前端放开，除非后端改挂到 `GetClientSubscription`。
- 能力放开：可优先补全配置分组、配置文件、发布版本、订阅者、加密算法；导入导出、`createandpub` 一键保存发布、灰度停止、操作历史是后端已有能力，但需要先确认页面入口和交互，不建议无设计直接暴露。

Review：

- 当前前端的主要错位在路由和边界转换，不是页面结构；直接在页面里兼容后端细节会扩大改动面。
- 当前后端的 `NewConfigGroupResponse`、`NewConfigFileResponse`、`NewConfigFileReleaseResponse` 未填充 `data`，会影响配置文件详情和发布详情页面；这应作为后端契约实现缺陷处理。
- 后续落代码时建议先补一个配置中心契约静态检查脚本，覆盖 `/groups/delette`、`/files/releases`、`/files/releases/delete`、`/files/releases/rollback`、`labels/tags` 归一和禁用旧的 `/files/release/delete|rollback`。

## 配置中心前端入口放开

- [x] 确认配置中心前端已有页面、路由、菜单和服务接口现状
- [x] 先补静态回归检查，覆盖配置中心入口可见性和未实现页面隐藏策略
- [x] 放开已具备真实闭环的配置分组入口，继续隐藏 Kubernetes 与模板占位能力
- [x] 运行前端静态脚本、构建、context-kg lint 和 diff 检查
- [x] 记录 review、验证结果和剩余风险

当前判断：

- 配置中心的 Redux slice、配置分组、配置文件、发布记录与订阅查询页面已经存在，服务接口也已接入 `/config/v1/groups`、`/config/v1/files` 和 `/config/v1/files/release`。
- 总路由 `console/web/src/router/index.ts` 当前仍把 `configuration` 模块注释掉，导致配置中心能力在侧边栏与页面路由中不可达。
- 顶层 Kubernetes 页面与配置分组页里的模板 Tab 仍是 `ECode.unimplemented` 占位，本轮不应跟随入口一起暴露。

修复：

- 放开 `configuration` 路由模块，让配置中心进入侧边栏与页面路由。
- 只暴露已有真实闭环的配置分组入口；配置文件、发布、订阅页面继续作为配置分组内的隐藏子路由进入。
- 继续隐藏 Kubernetes 配置页面，并移除配置分组页上的模板 Tab，避免把未实现占位能力暴露给用户。
- 修复 release 构建下的 vendor 分包运行时问题：React 生态依赖归入 `react-vendor`，TDesign shared 依赖的 `_chunks` 留在 `tdesign-shared`，其余 TDesign 组件按 data-form、base-overlay、navigation、misc 分组，避免 shared/data-form 循环初始化。

验证：

- `node scripts/verify-configuration-entry.mjs` 修复前失败，修复后通过。
- `node scripts/verify-vite-manual-chunks.mjs` 修复前失败，修复后通过。
- `cd console/web && for script in scripts/verify-*.mjs; do node "$script"; done` 通过。
- `cd console/web && npm run lint` 通过。
- `cd console/web && npm run build:test` 通过；`rg "Warning|warning|larger than|localstorage-file|Browserslist" /tmp/pole-config-build-test.log` 无命中。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check` 通过。
- `MYSQL_USER=root MYSQL_PWD=123456 MYSQL_HOST=127.0.0.1:3306 ./scripts/rebuild-start-all.sh --detach` 成功启动 all-mode，日志到 `finish starting server`。
- `curl http://127.0.0.1:8080/` 与 `curl http://127.0.0.1:8080/configuration/group` 均返回 `200`。
- Playwright 打开 `/configuration/group` 验证：配置管理、配置分组、分组名、新建可见；Kubernetes、模板不可见；页面不再出现 `attachEvent` 运行时错误或空白页。

Review：

- 本轮只放开配置中心中已经有前后端链路的配置分组能力，没有暴露 Kubernetes 和模板这两个未实现占位入口。
- 配置文件、发布记录、订阅查询仍通过配置分组详情流程进入，未新增平级菜单，避免菜单能力超过当前实现边界。
- 浏览器控制台还存在 TDesign Table 在 React 开发警告中的 `key` spread warning；该 warning 不影响页面渲染或本轮配置入口能力，未在本次范围内改第三方组件实现。

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
- OTel events 不应设计成独立第四类协议；按当前 OTel 语义应建模为带 EventName 的 LogRecord，再在 OpenObserver 查询层提供 events 视图。
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
- [[adr-otel-observability-platform]]
- [[adr-pole-rust-client-observability]]

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

- healthcheck：恢复 `test/data/service_test.yaml`、`service_test_sqldb.yaml` 测试 fixture；同时把 `Test_serialSetInsDbStatus` 收窄为内存 fake store 单测，避免为一个元数据写删函数启动整套 DiscoverTestSuit。后续 Pebble 迁移已删除旧 `bolt-data.yaml` fixture。
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
## OTel 可观测性平台设计归档

- [x] 回顾现有 `history`、`discoverEvent`、`statis` chain 与历史设计结论
- [x] 明确系统内部可观测性与业务服务调用可观测性两个平面
- [x] 明确统一走 OTel/OTLP，但不采集业务普通日志，仅结构化 event/audit 使用 OTel Logs
- [x] 选择 OpenTelemetry Collector Contrib 作为采集管道，OpenObserve 作为默认存储查询后端
- [x] 将 Kubernetes 快速体验部署路径写入 ADR
- [x] 更新 context-kg index/log 并执行结构校验

当前判断：

- `pole-control-plane` 负责上报自身内部 metrics、trace、event 和 audit，并负责 Console 查询分析；不负责业务流量采集。
- 业务侧 event、metrics、trace 由 sidecar 和 Rust SDK 上报。
- Collector 是统一采集/处理/路由入口，OpenObserve 是默认存储查询后端，Console 通过后端 `observability-query` 适配层读取分析。
- 生产标准路径应是 `sidecar/Rust SDK/pole-control-plane -> OpenTelemetry Collector Contrib -> OpenObserve -> Console`。
- Kubernetes 是默认交付体验路径，首期以 `pole-observability` namespace 部署 Collector 与 OpenObserve。

Review：

- 方案已归档到 [[adr-otel-observability-platform]]。
- 本轮只沉淀架构决策和部署方案，没有修改运行代码。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/technical/adr/adr-otel-observability-platform.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。
- 需要后续实现时，再拆分 Collector/OpenObserve Helm 示例、control-plane OTel entry、Console query API、SDK/sidecar 样例上报四条任务线。

## OTel metrics 与 event 命名规范设计

- [x] 对齐 OTel metrics 和 event 语义约定，确认 HTTP/RPC 优先使用标准语义名
- [x] 盘点当前仓库已有 metrics、discover event 和 history 字段
- [x] 设计 Pole 自定义 metrics 命名、单位和属性约束
- [x] 设计 Pole 结构化 event 名称、属性和严重级别约束
- [x] 更新 [[adr-otel-observability-platform]] 并补充验证记录

当前判断：

- 适用 OTel 标准语义约定的 HTTP/RPC 指标不另造 `pole.*` 名称；Pole 自定义领域指标统一使用 `pole.` 前缀。
- 指标属性必须低基数，`rule.id`、`config.file`、`trace_id`、`request_id` 等高基数字段不能进入 metrics label，放到 trace 或 event。
- Event 使用 OTel LogRecord 的 EventName 语义，EventName 必须是低基数的全限定名称，动态值全部放 attribute。

Review：

- 已在 [[adr-otel-observability-platform]] 增加 metrics 与 EventName 规范。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/technical/adr/adr-otel-observability-platform.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。

## OTel metrics 覆盖与联动设计补充

- [x] 横向审计平台内部 metrics 缺口：启动、ready、插件、EventHub、配置 watch、治理发布、推送、鉴权、DB pool、telemetry export
- [x] 补充业务 CPU/Mem 与请求监控数据的关联模型
- [x] 补充治理 metrics 与治理 event 的关联规则
- [x] 更新 [[adr-otel-observability-platform]] 并补充验证记录

当前判断：

- 平台内部 metrics 不能只覆盖请求、Store、Cache 和资源数量，还要覆盖启动健康、异步队列、配置发布/Watch、治理发布、协议推送、鉴权、DB 连接池和观测出口健康。
- 业务 CPU/Mem 不由 Pole SDK 自己采进程资源，而由 Kubernetes 模式下 Collector `kubeletstats` / Kubernetes 相关 receiver 采集容器/Pod 指标，再通过 `k8s.pod.uid`、`service.instance.id`、`pole.service.name` 等资源属性与业务请求指标联动。
- 治理 metrics 负责低基数聚合，治理 event 负责具体规则、实例、trace 的高基数定位，两者通过 `pole.governance.decision.id`、`trace_id/span_id`、`pole.rule.type` 和时间窗口关联。

Review：

- 已在 [[adr-otel-observability-platform]] 补齐平台内部 metrics 覆盖面、业务 CPU/Mem 与请求指标联动、治理 metrics 与治理 event 联动。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/technical/adr/adr-otel-observability-platform.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。

## OTel 服务绑定标签与语言体系补充

- [x] 设计 Pole 服务/实例保留标签，用于绑定请求 metrics、runtime metrics、Kubernetes CPU/Mem、trace 和 event
- [x] 增加 `pole.io/runtime-language` 与 `pole.runtime.language`，用于 Console 选择 Java/Go/Rust/Node/Python/.NET 运行时面板
- [x] 明确 Java/Go 使用 OTel 标准 runtime metrics 名称，不新增 `pole.*` runtime 指标名
- [x] 更新 [[adr-otel-observability-platform]] 并补充验证记录

当前判断：

- 服务调用指标和系统资源指标的关联应依赖显式绑定标签，不应靠服务名或 metric name 猜测。
- 语言体系只决定 Console 查询和展示哪些标准 runtime metrics；runtime 指标名称继续使用 JVM、Go 等 OTel 标准语义。
- 绑定优先级为 `k8s.pod.uid + container.name`、`service.instance.id`、`workload`、`namespace/service`。

Review：

- 已在 [[adr-otel-observability-platform]] 增加服务绑定标签、运行时语言体系和 Java/Go runtime metrics 展示规则。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/technical/adr/adr-otel-observability-platform.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。

## 可观测性 ADR 子目录与 pole-rust-client 职责设计

- [x] 明确 `pole-rust-client` 在观测体系里的职责与非职责
- [x] 创建 `context-kg/technical/adr/observability/` 子目录
- [x] 移动 OTel 平台 ADR 到可观测性子目录
- [x] 新增 Rust SDK 观测职责 ADR
- [x] 更新 index/log/todo 并执行结构校验

当前判断：

- `pole-rust-client` 是业务侧观测入口之一，负责服务绑定属性、SDK 内部 metrics/events、治理决策关联和 tracing hooks。
- `pole-rust-client` 不负责业务普通日志、不负责 Kubernetes CPU/Mem 采集、不替代业务框架标准 HTTP/gRPC instrumentation。
- 治理 metrics 只做低基数聚合；具体规则、实例、trace 由 event/trace 通过 `pole.governance.decision.id` 关联。
- 可观测性 ADR 后续统一放到 `context-kg/technical/adr/observability/`。

Review：

- 已创建 `context-kg/technical/adr/observability/`，并将 [[adr-otel-observability-platform]] 移入该目录。
- 已新增 [[adr-pole-rust-client-observability]]，明确 Rust SDK 的 Resource attributes、SDK metrics/events、治理决策关联、trace hooks 和非职责。
- `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg` 通过。
- `git diff --check -- context-kg/technical/adr/observability/adr-otel-observability-platform.md context-kg/technical/adr/observability/adr-pole-rust-client-observability.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md` 通过。

## 可观测性后端 GreptimeDB 优先选型调整

- [x] 核对 GreptimeDB 与 OpenObserve 官方定位和 OTel 支持范围
- [x] 将长期默认后端调整为 GreptimeDB
- [x] 保留 OpenObserve 作为 quickstart / 可选 provider
- [x] 更新 Collector、Kubernetes、Phase、验证要求和 Rust SDK 边界描述
- [x] 更新 index/log/todo 并执行结构校验

当前判断：

- Pole Console 要自己做服务、实例、治理、配置、审计的领域化分析，因此长期默认后端更适合选择 observability database，而不是完整观测平台。
- GreptimeDB 支持 metrics/logs/traces 统一 OTel 后端、SQL 和 PromQL，更适合作为 `observability-query` 的默认 provider。
- OpenObserve 仍适合 quickstart 和内置 UI 辅助排查，但不作为长期默认后端。
- OpenTelemetry Collector Contrib 的职责不变，仍负责 OTLP 接入、processor、过滤、补标签、批量、重试和路由。

Review：

- 已确认当前 ADR 正文中 GreptimeDB 是长期默认后端，OpenObserve 只作为 quickstart / 可选 provider；历史任务记录中的旧判断保留为历史记录。
- 已更新平台 ADR、Rust SDK ADR、`context-kg/_meta/index.md` 与 `context-kg/_meta/log.md`。
- 已执行 `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg`，通过 Markdown/frontmatter/link/index 基础检查。
- 已执行 `git diff --check -- context-kg/technical/adr/observability/adr-otel-observability-platform.md context-kg/technical/adr/observability/adr-pole-rust-client-observability.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md`，未发现 whitespace error。

## 认证管理页面整体布局优化

- [x] 审计认证管理主体与策略页面的真实视觉、信息层级和共享页面范式
- [x] 重构页面级标题、导航与内容区布局，移除零散内联间距
- [x] 横向检查用户、用户组、角色、自定义策略和默认策略五个视图
- [x] 运行前端静态校验、构建和真实页面视觉验收
- [x] 补充 Review 与验证记录

当前判断：

- 认证管理目前直接以 Tabs 包裹带 `margin: 20` 的资源表格，缺少页面身份、功能分组说明和稳定内容边界；主体管理与策略管理也没有形成同一套页面级视觉结构。
- 优化应复用现有 Fluent 与资源列表组件，不改变表格、抽屉和 API 行为，重点修正页面骨架、留白、导航层级和响应式表现。

Review：

- 主体管理与权限策略统一为 `ResourceHeader → Tabs → ResourceToolbar → Table` 四层结构；页头说明认证域目标，Tab 承载资源分类，内容区只保留一行上下文提示，避免重复标题和装饰标签。
- 用户、用户组、角色、自定义策略和默认策略五个视图统一了列表标题、总数、搜索、刷新和新建操作；新建按钮明确具体资源类型，刷新使用标准方形按钮。
- 页面工作区使用同一套边框、圆角、浅色 Tab 导航面和响应式横向滚动策略，原有表格、抽屉、权限判断和 API 行为未调整。
- `npm run lint -- --quiet` 与 `npm run build:test` 通过；all-mode 重建日志出现 `finish starting server`，`8080` 与 `8090` 均返回 `200`。
- 使用 `admin` 真实登录后逐一切换五个视图，列表与空状态均正常加载，浏览器控制台无 error；视觉截图位于 `output/playwright/auth-principals-layout.png` 和 `output/playwright/auth-policies-layout.png`。
## 服务详情增加流量治理上下文

- [x] 核对服务详情标签结构、治理工作台列表和各规则编辑器的服务模型
- [x] 在服务订阅后新增“流量治理”标签并传入当前命名空间/服务
- [x] 将治理工作台适配为服务上下文模式，只展示与当前服务关联的规则
- [x] 增加“作为主调方 / 作为被调方”角色切换，默认主调方
- [x] 新建规则时固定当前服务到选定角色，单端规则固定目标服务
- [x] 补充静态回归校验、构建、重启和真实浏览器验收

当前判断：

- 服务详情中的治理能力应复用现有治理工作台，不能复制一套规则列表、抽屉和发布逻辑。
- 角色切换只决定新建双端规则时当前服务绑定到 caller 还是 callee；规则清单仍展示当前服务作为任一端或目标服务的全部关联规则。
- 路由、熔断、镜像、Mock、泳道等双端规则默认把当前服务固定为 caller，用户可在标签页顶部切换为 callee；鉴权、限流、主动探测、无损等按目标服务归属的规则固定当前服务为目标。

修复：

- 服务详情在“服务订阅”后新增“流量治理”Tab，懒加载并复用 `GovernanceWorkbench`；嵌入模式保留统一查询栏、规则表、详情抽屉、版本、发布和删除链路。
- 工作台按路由/熔断的 caller-callee、流量治理规则的 caller/callee/target、泳道入口/目标和单端规则目标服务计算关联关系，只展示当前服务相关规则；服务上下文请求上限提高到 100，减少全局首页截断导致的漏项。
- 顶部使用分段控件切换“作为主调方 / 作为被调方”，默认主调方；切换只影响后续新建规则的服务绑定，不改变关联规则清单。
- 共享 `ServiceScopeSection` 新增固定端能力：固定端以无边框只读文本和“当前服务”标签展示，另一端保持可编辑。
- 路由、熔断、镜像、Mock 和泳道按角色绑定当前服务；鉴权、限流、主动探测、无损按规则索引模型固定当前目标服务。未提交任何测试规则。

验证：

- `cd console/web && npm run lint && npm run build:test` 通过；全部 `scripts/verify-*.mjs` 通过，新增 `verify-service-governance-context.mjs` 固定入口、筛选、默认角色和编辑器透传契约。
- all-mode release 重建后日志出现 `finish starting server`；`8080`、`8090` 和服务详情 URL 均返回 `200`。
- 真实浏览器使用 `admin/admin123` 打开 `spec-governance/spec-checkout`：确认“流量治理”位于“服务订阅”后，默认主调；新建路由时主调端固定当前服务，切换被调后被调端固定当前服务，未绑定端仍为可编辑 Select。
- 交互截图输出到 `output/playwright/service-governance-callee-binding.png`；`context-kg` lint 与 `git diff --check` 均通过。

Review：

- 新能力没有复制治理页面或后端接口，服务详情与全局治理工作台共享同一数据和操作链路。
- caller/callee 角色仅在有双端语义的规则中生效；单端规则继续遵守当前后端资源归属模型，不伪造 caller 字段。

## 可观测性 Console 查询接口与 Rust SDK 动态配置设计

- [x] 核对现有 Console 监控路由与可观测性 ADR 边界
- [x] 设计 `observability-query` API 分组、请求模型、响应模型和后端 provider 边界
- [x] 设计 `pole-rust-client` 通过服务发现获取 OTLP 上报地址
- [x] 设计 `pole-rust-client` 通过配置中心 remote 下发 SDK 观测配置
- [x] 更新平台 ADR、Rust SDK ADR、index/log/todo 并执行结构校验

当前判断：

- Console 不应直接暴露 GreptimeDB/OpenObserve 查询语法；前端只调用 Pole 领域化查询接口。
- `pole-rust-client` 应优先通过 Pole 服务发现获取 Collector/sidecar OTLP endpoint，减少用户配置。
- SDK 观测开关、采样率、batch/export、治理事件开关等动态项适合通过配置中心下发，但必须保留本地静态配置和 OTel 标准环境变量作为兜底。

Review：

- 已在 [[adr-otel-observability-platform]] 增加 `/observability/v1` Console 查询接口设计，包括 overview、topology、metrics、runtime、events、traces、governance、platform、audit 和 correlate。
- 已在 [[adr-pole-rust-client-observability]] 增加 OTLP endpoint 服务发现、`pole-otel-collector` 注册约定、配置中心 `pole-sdk/observability.yaml` remote config、配置优先级和热更新规则。
- 已确认当前 Console 已有 `/metrics/v1` 历史监控路由，新 OTel 查询接口独立放在 `/observability/v1`，避免扩大旧接口职责。
- 已执行 `python3 /Users/chuntao.liao/.codex/skills/context-kg-maintainer/scripts/context_kg_lint.py ./context-kg`，通过 Markdown/frontmatter/link/index 基础检查。
- 已执行 `git diff --check -- context-kg/technical/adr/observability/adr-otel-observability-platform.md context-kg/technical/adr/observability/adr-pole-rust-client-observability.md context-kg/_meta/index.md context-kg/_meta/log.md context-kg/tasks/todo.md`，未发现 whitespace error。

## 监控指标事件与操作页面 mock-first 设计

- [x] 核对现有事件指标、操作审计页面结构和监控路由
- [x] 设计并落地事件指标页面 mock-first 展示效果
- [x] 设计并落地操作审计页面 mock-first 展示效果
- [x] 增加静态验证，约束页面关键结构和 mock 降级能力
- [x] 执行前端验证、构建检查、context-kg lint 和 diff 检查
- [x] 新增系统监控页面，展示平台自身组件运行看板
- [x] 新增服务监控页面，展示服务流量与治理效果看板
- [x] 更新监控菜单顺序、静态校验和浏览器验证记录
- [x] 将系统监控升级为 Grafana-like 看板，支持类别与接口筛选
- [x] 将服务监控升级为 Grafana-like 看板，支持服务、接口与实例筛选

当前判断：

- “事件指标”和“操作审计”不应只是日志表格，应按观测页面设计：摘要、趋势、分布、筛选、列表、详情联动。
- 当前后端数据可能为空或接口未准备完整，页面应允许通过 mock 数据先把体验效果勾勒出来，同时保留真实接口接入点。
- 这两页属于管理控制台，不做营销式大图和装饰卡片，优先信息密度、时间线、可定位性和低噪声。

Review：

- 事件指标页已改为 mock-first 观测视图，包含事件摘要、风险事件数、影响服务数、最近事件时间、小时趋势、事件类型分布、筛选表格和事件详情抽屉；真实接口为空或失败时自动静默展示 Mock 预览。
- 操作审计页已改为审计观测视图，包含操作总数、高风险操作、活跃操作者、涉及资源、小时趋势、资源类型分布、筛选表格和操作详情抽屉；真实接口为空或失败时自动静默展示 Mock 预览。
- 同步修正前端查询链路：事件页 `resource` 与 `event_type` 参数不再传反；操作页补齐 `resource_type`、`operation_type`、`operator` 查询参数，并使用接口契约中的 `detail` 字段展示操作详情。
- 已新增 `npm run test:metrics-observability` 静态校验，约束两页保留 mock fallback、远程接口调用、详情抽屉和关键字段映射。
- 已通过 `npm run test:metrics-observability`、`npm run lint`、`npm run build:test`、context-kg lint 和相关文件 `git diff --check`。
- 已通过 Playwright 在 `http://127.0.0.1:5174/metrics/event`、`http://127.0.0.1:5174/metrics/operation` 验证页面非空、Mock 预览可见、筛选区不挤压、摘要时间不截断；截图保存在 `output/playwright/metrics-event.png` 与 `output/playwright/metrics-operation.png`。
- 监控菜单已扩展为 `系统监控 / 服务监控 / 事件指标 / 操作审计`。系统监控展示平台自身组件资源、请求、写入链路、平台信号和组件明细；服务监控展示服务请求、治理命中、拦截/降级、治理分布和限流、路由、熔断、探测、鉴权、Mock、镜像明细。
- 已通过 Playwright 在 `http://127.0.0.1:5174/metrics/system`、`http://127.0.0.1:5174/metrics/service` 验证新增页面非空、菜单顺序正确、Mock 预览可见、筛选区和表格可读；截图保存在 `output/playwright/metrics-system.png` 与 `output/playwright/metrics-service.png`。
- 系统监控已升级为 Grafana-like dashboard：顶部变量筛选支持 `类别 / 接口 / 组件或 Pod`，下方展示 stat panels、接口延迟趋势、延迟热力图、组件资源和接口明细，筛选会驱动所有 mock 面板。
- 服务监控已升级为 Grafana-like dashboard：顶部变量筛选支持 `服务信息 / 接口 / 实例 / 治理能力`，下方展示请求量、治理命中、拦截/降级、P95/P99、服务流量趋势、治理分布、实例延迟和治理信号明细，筛选会驱动所有 mock 面板。

## 服务治理监控下钻预览

- [x] 梳理现有服务监控看板、路由和 mock 数据结构
- [x] 将服务监控首页调整为服务级整体概览
- [x] 新增独立服务详情页，支持从服务总览下钻
- [x] 详情页按参考图设计为左侧接口列表、右侧接口概览与治理 Tab
- [x] 运行静态验证、lint、构建和浏览器截图验证
- [x] 补充 Review 与截图路径

当前判断：

- 服务监控入口应先回答“哪些服务值得关注”，不要一开始就把治理信号按规则铺开；规则、接口和实例应放在服务详情页下钻。
- 详情页适合参考用户截图中的工作台结构：左侧资源/接口列表，右侧按 Tab 展示接口概览、节点/实例、调用流量、路由、限流、熔断、探测、鉴权、Mock、镜像和事件时间线。
- mock 预览必须保留服务、接口、实例、语言体系和治理能力标签，后续真实查询接口可按同一组 OTel resource labels 关联调用指标、系统指标和治理事件。

Review：

- 服务监控首页已从治理信号明细看板调整为服务级整体概览，保留 `服务信息 / 接口 / 实例 / 治理能力` 筛选，并按服务聚合请求量、治理命中、拦截/降级、P95/P99、CPU/MEM、语言体系和治理能力标签。
- 新增隐藏路由 `/metrics/service/detail?namespace=...&service=...`，从服务总览行点击进入独立详情页；详情页采用左侧接口目录、服务端/客户端调用方向切换、右侧 Tab 下钻结构，不再区分 RPC / Web 服务类型。
- 服务详情页已覆盖 `接口概览 / 实例详情 / 调用流量 / 路由 / 限流 / 熔断 / 探测 / 鉴权 / Mock / 镜像 / 事件时间线`；实例详情只展示实例维度调用概览、系统资源与延迟表，不展示 JVM、Go runtime、Rust runtime 等语言运行时指标。
- 已新增共享 mock 数据文件，统一服务、接口、实例、语言和治理能力数据，避免首页与详情页指标不一致。
- 已通过 `npm run test:metrics-observability`、`npm run lint`、`npm run build:test`、`git diff --check` 和 `context-kg` lint。
- 已通过 Playwright 在本地 Vite mock 服务 `http://127.0.0.1:5174/metrics/service` 验证服务级总览，并从 `default/checkout` 行点击下钻到详情页；截图保存在 `output/playwright/metrics-service-overview-drilldown.png` 与 `output/playwright/metrics-service-detail-drilldown.png`。
- 根据最新反馈，服务详情页已移除 `WEB服务 / RPC服务` 协议类型切换，接口元信息不再根据协议分支渲染；静态验证新增禁止 RPC/Web 服务类型拆分的约束。
- 根据最新反馈，`实例详情` 已移除语言运行时指标卡片，改为 `实例调用概览` 趋势与 `实例延迟与资源` 表格；静态验证新增禁止 JVM、GC、goroutine、Tokio 等运行时指标回归的约束。Playwright 已切换到 `实例详情` Tab 刷新截图 `output/playwright/metrics-service-detail-drilldown.png`。
- 根据最新反馈，服务监控首页已移除服务级总览表格前的 `服务流量趋势 / 治理分布` 面板，摘要指标后直接展示服务级表格；静态验证新增禁止首页图表面板回归的约束。
- 根据最新反馈，服务监控首页筛选区已从页头下方移动到数字总览和服务级表格之间；静态验证新增 `statGrid -> variableBar -> 服务级总览` 顺序约束。
- 根据最新反馈，服务监控首页已移除 `接口 / 实例` 筛选，只保留 `服务信息 / 治理能力`；接口和实例筛选归属服务详情下钻页。
- 浏览器控制台仍保留登录后命名空间页触发的既有 Fluent Tooltip ref warning，本轮服务监控预览路径未新增阻塞性运行错误。

## 系统命名空间删除保护与前端标识

- [x] 核对 `default`、`pole-system` 常量与命名空间删除调用链
- [x] 后端拒绝删除两个系统自带命名空间并补充测试
- [x] 前端禁用两个系统命名空间的删除操作
- [x] 在命名空间列表明确标记 `pole-system` 为 Pole 内部系统空间
- [x] 更新命名空间知识页并完成后端、前端和真实页面验证

当前判断：

- `default` 与 `pole-system` 都属于系统自带命名空间，不可删除约束必须由后端兜底，前端禁用只负责提前表达限制。
- `pole-system` 还承载 Pole 内部系统资源，列表需要提供稳定、明确的系统空间标识；`default` 只需表达默认命名空间属性，不能与内部系统空间混为一类。

Review：

- 后端在创建删除事务前统一拒绝 `default` 与 `pole-system`，批量删除复用同一保护逻辑；命名空间列表同时返回 `deleteable=false`。
- 前端为 `default` 标记“默认空间”、为 `pole-system` 标记“内部系统空间”，并同时在按钮状态与删除分发入口阻断两个内置空间。
- 已通过 `go test ./pkg/namespace`、`go test ./...`、前端静态校验、`npm run lint -- --quiet`、`npm run build:test` 和 context-kg lint。
- all 模式重建后，`8080`、`8090` 均返回 HTTP 200；真实接口删除两个系统空间均返回 HTTP 400 / `400103`，且数据仍存在。
- Playwright 实页验证两个标签和禁用态均正确，普通命名空间删除按钮不受影响，浏览器控制台无错误。

## TDesign 依赖移除可行性审计

- [x] 定位“页面配置”抽屉及内部控件的真实实现
- [x] 审计生产源码、样式、依赖声明和校验脚本中的 TDesign 残留
- [x] 判断当前是否满足移除 `tdesign-react` 的条件

Review：

- 抽屉外壳已经使用 Fluent UI `OverlayDrawer`，主题卡片与色点为自绘 DOM；彩虹颜色选择器的 `Popup`、`ColorPickerPanel` 仍直接来自 TDesign。
- `components/Fluent/index.tsx` 仍以兼容层方式使用 TDesign Form、全量样式、十余个组件及大量类型，业务页面尚未完成纯 Fluent 化。
- 生产样式仍存在较多 `.t-*` 选择器与 `--td-*` 主题变量；现阶段移除 `tdesign-react` 会导致构建失败和页面交互、样式损坏，因此本次审计没有卸载依赖。

## Console 全量迁移 Fluent UI 并移除 TDesign

- [x] 建立生产源码、依赖和样式层面的零 TDesign 校验
- [x] 迁移 Form/FormItem 及其现有实例、校验和类型契约
- [x] 迁移 Popup、Tree、Transfer、Steps、日期时间等复杂控件
- [x] 将页面配置面板改为 Fluent UI 控件并保持当前视觉布局
- [x] 清理 `tdesign-react`、`tdesign-icons-react`、全量样式和锁文件
- [x] 完成 lint、构建、真实页面交互和视觉回归验证

视觉基准：保持当前 Console 的信息密度、间距、色彩和布局不变，只替换底层控件实现。

内容基准：导航、表格、表单、抽屉、详情页的信息结构和文案完全沿用现状。

交互基准：保留现有打开/关闭、校验、选择、分页与主题切换行为，不增加新的动画或视觉重设计。

Review：

- `components/Fluent` 已由本地 Form 与复杂控件适配实现接管，生产源码、样式和依赖树均无 TDesign；分页及原先自绘按钮也已改用 Fluent 控件。
- `test:no-tdesign` 固化依赖、锁文件、import、CSS 类名、主题变量与原生交互标签六类门禁，`npm ls tdesign-react tdesign-icons-react --all` 返回空树。
- 页面配置抽屉与迁移前截图保持 538px 表面、主题卡片和色点布局；认证策略列表、命名空间表单、策略多步表单、Tree/Transfer、主题切换均完成真实浏览器操作，控制台无错误。
- `npm run lint -- --quiet`、`npm run build:test` 通过；all 模式 release 重建日志出现 `finish starting server`，8080/8090 运行入口可用。

## 页面配置面板切换为 Fluent UI 默认外观

- [x] 核对旧预览卡、色点与抽屉尺寸的自定义样式来源
- [x] 使用 Fluent `Field`、`RadioGroup`、`Radio` 重做主题模式选择
- [x] 使用 Fluent `SwatchPicker`、`ColorSwatch` 重做主题色选择
- [x] 移除旧卡片/色点强制尺寸，并恢复 Fluent Drawer 默认尺寸
- [x] 完成零 TDesign、lint、构建及明暗主题真实页面验证

视觉基准：采用 Fluent UI v9 默认设置表单外观，不再保持旧 TDesign 风格的三张主题预览卡和描边色点。

交互基准：继续即时切换明亮、黑暗、跟随系统和主题色；自定义颜色入口继续可用。

Review：

- 页面配置使用 Fluent `Field + RadioGroup + Radio`、`Field + SwatchPicker + ColorSwatch` 和原生 `Popover + Button`；旧主题图片卡片、彩虹色点和强制尺寸样式已移除，Drawer 使用官方 `medium` 尺寸。
- “跟随系统”在亮暗系统主题下均可进入选中态，刷新后保持，并监听 `prefers-color-scheme` 变化即时同步；HEX 比较统一小写，取色器返回大写预设值时仍能正确选中对应色板。
- `npm run lint`、`npm run test:no-tdesign`、`npm run build:test` 通过；release all 模式重建出现 `finish starting server`，8080/8090 均返回 HTTP 200。
- Playwright 验证亮色、暗色、跟随系统、预设色和自定义取色；页面可访问树为标准 radio/radiogroup，DOM 类名和资源请求中均无 TDesign。

## Pebble 本地 protobuf value cache 设计

- [x] 将配置文件和服务契约本地 value cache 从 bbolt 切换为 Pebble
- [x] 将 OTel 操作审计和服务事件本地可靠队列从 bbolt 切换为 Pebble
- [x] 明确 Pebble 适用场景：实例列表、配置、服务契约和治理规则发布快照
- [x] 明确默认内存、开关启用后才按“内存索引 + Pebble value”卸载大数据
- [x] 明确启用 Pebble 后客户端数据读取性能不能低于当前路径
- [x] 明确心跳状态继续留在内存热路径
- [x] 新增长期 ADR 并同步缓存层页面、index 和 log
- [x] 实现共享 Pebble 封装并接入配置文件、服务契约、OTel event/audit queue
- [x] 移除测试套件中遗留的 boltdb store 默认配置、`bolt-data.yaml` fixture 和 bbolt 依赖残留
- [ ] 后续实现完整 `ValueCache` 开关抽象，并分阶段接入服务发现和治理下发
- [ ] 后续补齐服务发现、配置发现、服务契约和治理下发性能基准

Review：

- Pebble 的定位是本地可重建 value cache，存 protobuf marshal 后的 bytes，不替代 MySQL、注册内存态或治理 cache 的事实来源。
- 实例列表 key 必须携带 revision 与过滤/权限可见性 hash，避免不同调用者可见实例误命中。
- 开关关闭时必须完全保持当前内存路径；开关开启后才把面向客户端的大 value 卸载到 Pebble，且没有通过性能基准的领域不能纳入卸载范围。
- 配置文件和服务契约已先切到 Pebble：内存继续保留轻量索引，本地 value 从 `config_file.pebble` / `service_contract.pebble` 读取。
- OTel 操作审计和服务事件本地队列已切到 Pebble：写入 `otel-events.pebble`，按 `otel-queue/{history|discover_event}/` 前缀隔离，Collector 成功确认后删除。
- 测试默认 fixture 已不再指向 `boltdbStore`，`test/suit` 不再注入 `bolt-data.yaml` 或清理 `polaris.bolt`。
- 治理规则发布快照和服务发现响应后续再接入；实例心跳继续只保留在内存，最多做低频快照，不进入同步写磁盘链路。

## Pebble 切换完整验证

- [x] 运行 control-plane 后端全量 Go 测试
- [ ] 运行 console 前端 lint、构建和现有校验脚本
- [ ] 运行 control-plane all-mode 真实启动和关键 HTTP 接口 smoke
- [ ] 运行 pole-client-rust 客户端测试
- [ ] 汇总验证结论和剩余风险

Review：

- `go test -count=1 ./...` 已通过；补齐 Pebble/Gin 相关 checksum 后未发现后端断言失败。
- 当前代码/配置和 Go 模块图均无 bbolt/boltdb 运行残留。

## 服务调用鉴权：托管服务身份与自定义 Header 双模式

- [x] 核对现有调用鉴权规则、自定义 Header 匹配和客户端下发链路
- [x] 明确服务身份标识、身份凭证、control-plane 与数据面的职责边界
- [x] 形成默认托管身份与可选自定义 Header 模式的长期 ADR
- [x] 同步 index、log 与 lessons，并运行 context-kg lint / diff check
- [x] 核对客户端如何绑定自身服务身份，以及是否复用 Discover 流
- [x] 明确身份句柄、workload credential 与 trust bundle 的下发协议
- [x] 将身份领取/续期边界补充到 ADR 并完成校验
- [x] specification 增加认证模式、托管 caller selector 与 `SERVICE_IDENTITY` Discover 契约
- [x] control-plane 增加隐藏服务身份生命周期与严格 token 认证的身份发现
- [x] Rust SDK 补齐 Discover token metadata、身份订阅和内部 descriptor 缓存
- [x] Console 新建规则默认托管身份，自定义 Header 降为兼容模式
- [x] 完成跨仓定向测试、构建、context-kg lint 与代码审查
- [x] 实现 SERVICE_TOKEN binding 的 WorkloadCredential Issue/Renew 与被调方入站验证

第二阶段（已完成）：

- [x] 固化短期凭证 claims、签名算法、key rotation 与 trust bundle 契约
- [x] specification 增加 WorkloadCredential Issue/Renew RPC 和验证材料 Discover
- [x] control-plane 实现签发密钥生命周期、service token/workload 绑定校验与短期凭证签发
- [x] Rust SDK 实现凭证领取/续期、HTTP/gRPC 出站注入和入站离线验证
- [x] 将验证结果写入可信 `AuthenticatedCaller`，按 `managed_caller` 执行授权
- [x] 增加过期、篡改、跨服务冒用、轮换窗口和 control-plane 暂时不可用时的本地凭证保持语义
- [x] 更新 ADR、context-kg log，完成跨仓构建和代码审查

当前判断：

- 调用鉴权解决服务间信任，不复用 Console 用户 Token 或资源授权策略作为服务身份。
- 默认模式应由 control-plane 为每个服务管理稳定、不可由用户查看或修改的服务身份，并向合法 workload 下发可证明该身份的短期凭证；仅下发一个可复制的身份字符串不足以完成认证。
- 现有“匹配请求 Header 的自定义 value”保留为显式可选的兼容模式，不作为默认安全模型。

Review：

- 当前 `TrafficSecurityRule` 只有 API、通用流量匹配和 `ALLOW/DENY`，Console 默认条件确实是 `HEADER / authorization / EXACT / value`；这属于兼容型请求字段匹配，不是托管身份认证。
- 当前 `Service.token` 出现在公开 Service 契约中，并支持查询和刷新，不能直接当作不可见、不可修改的服务身份；最多在迁移期作为受限 bootstrap credential。
- 当前 Rust client 在调用方出站链路执行 policy，调用方可以绕过；可信模式必须把凭证验证和最终拒绝放到被调方入站 SDK/sidecar。
- ADR 已将 `ServiceIdentity` 与 `WorkloadCredential` 分离，明确调用方只领取自己的短期凭证，被调方只接收 trust bundle 和策略，control-plane 不进入业务调用热路径。
- 已通过 context-kg lint（35 个 Markdown 页面）与相关文件 `git diff --check`；本轮只沉淀架构决策，没有修改运行代码。

身份领取补充判断：

- `Service.token` 专用于 SDK → control-plane 认证，`ServiceIdentity` 专用于服务 → 服务的数据面主体；token 校验通过后，control-plane 才能确认 SDK 有权领取哪个服务的身份描述。
- `SERVICE_IDENTITY` 可以作为新的 Discover type，下发稳定、非秘密且用户不可见/不可改的 identity descriptor，并按服务 revision 缓存。
- 实例级 workload credential 仍使用独立 Issue/Renew RPC 定向返回，不进入 Discover 普通缓存；Discover 还可下发公共 trust bundle/issuer key。
- 身份 Discover 必须由 gRPC metadata 中的 service token 严格认证，服务端从认证 principal 推导 identity；当前 Rust 长连接尚未稳定携带 token metadata，需要先补齐，并禁止 identity type 走 anonymous/clientStrict 兼容路径。

实施范围：

- 当前先交付 `service token -> SERVICE_IDENTITY Discover -> SDK 内部 descriptor` 以及 Console/spec 模式闭环，证明服务身份能稳定创建、严格领取和被策略引用。
- `WorkloadCredential` 的签发、轮换与被调方入站验签是下一阶段安全闭环；在它完成前，不能宣称托管身份已经能阻止绕过 SDK 的恶意调用方。

实施 Review（2026-07-20）：

- `ServiceIdentity` 使用独立表与 subject/revision，新建真实服务事务内创建，历史真实服务在严格 metadata token 认证后原子补齐；alias 不拥有独立身份，管理 Service 列表继续使用不含 token 的窄投影。
- `SERVICE_IDENTITY` 禁止请求体 token 覆盖，并从 token 对应服务推导 descriptor；namespace/service 仅做一致性校验，重复 token、空 token 和声明不一致均 fail-closed。
- Rust 配置显式命名为 `serviceIdentityDiscovery.controlPlaneToken`，避免 token 与 descriptor 混淆；descriptor 只存在 Engine 内部，成功后停止两秒轮询，不进入普通/failover cache。
- Custom Header value 在管理写入时立即转换为 SHA-256 摘要并清空明文；mode、子配置和 managed caller selector 在服务端做一致性校验。
- Rust 仅使用不含 plaintext 的 Custom Header 摘要做精确比较；托管身份在 `AuthenticatedCaller` 尚未落地时 fail-closed，不会因 matcher 缺失而误放行。
- Console 实时 Spec 已按 mode 生成 managed caller/custom header 结构，value 固定显示 `<redacted>`；第二阶段完成后页面改为提示业务显式挂载 SDK 出站注入与入站验签适配器。

第二阶段实施 Review（2026-07-20）：

- specification 新增 Ed25519 JWT 的 Issue/Renew、binding type、错误码与 `SERVICE_IDENTITY_BUNDLE=31`，请求体不允许选择服务主体，request ID 只用于关联。
- control-plane 仅使用 metadata service token 推导签发主体；私钥通过只读文件引用加载，要求唯一 ACTIVE key，并用 VERIFY_ONLY key 支撑滚动轮换。Trust Bundle 只包含公开验证材料、issuer、有效期和单调 sequence，不进入普通发现缓存。
- Rust Engine 在内存中维护 descriptor、trust bundle 和短期凭证，按 renew-after 续期；identity revision 变化立即重新 Issue，续期失败时未过期凭证仍可用于数据面。
- descriptor 与 trust bundle 保持周期刷新；同 sequence/version 的 bundle 只有在 issuer、trust domain、key 和吊销集合一致时才允许续租有效期，防止有效期刷新被用来替换安全材料。
- Rust SDK 与 control-plane 支持按部署需要选择 `grpc` 或 `grpcs`；生产环境推荐 TLS，但托管身份不以 TLS 为强制前置条件。
- SDK 暴露 HTTP Header 注入、tonic client interceptor、HTTP/tonic 入站验证器，但由于 SDK 不拥有业务网络栈，业务必须显式挂载这些适配器，不能描述为自动拦截全部业务调用。
- 入站只在 Ed25519 签名、issuer、audience、时间窗、kid、bundle sequence 和吊销条件全部通过后创建 `AuthenticatedCaller`；托管 caller 策略不读取普通 caller/header/metadata 自报值。
- V1 是 `SERVICE_TOKEN` 服务级 bearer credential，不能证明具体 Pod/实例，且被窃取后在有效期内有重放风险；TLS、短 TTL、audience 与吊销是当前缓解，PoP/mTLS 留待后续。
- 已通过 specification Rust 测试、control-plane `go test ./...`、Rust `cargo test --workspace` 与格式检查；未在本轮引入凭证明文日志、持久化或公共读取接口。

TLS 按需调整（2026-07-20）：

- [x] 移除 Rust 托管身份对 `grpcs` 的强制校验，保留 `grpc` 与 `grpcs` 两种 endpoint。
- [x] 移除 control-plane 身份 Discover、Issue/Renew 对非 TLS transport 的强制拒绝。
- [x] 更新安全边界说明：TLS 为生产推荐的可选传输保护，不是启用托管身份的前置条件。
- [x] 更新测试、ADR、log、lessons，并完成定向与全量验证。

TLS 按需调整 Review（2026-07-20）：

- 需求复核确认代码与 ADR 均未残留强制 `grpcs`/TLS 条件，`grpc` 与 `grpcs` 分别映射到明文和 TLS endpoint，身份认证逻辑不再按 transport 分叉。
- 当前 `grpcs` 使用 WebPKI 根证书；历史 `ssl` 配置虽存在但旧连接器从未消费且始终使用 `http://`。私有 CA/mTLS 配置应作为独立增强恢复，不能宣称本轮已经支持。
- 当前单元测试覆盖协议映射、metadata 和 handler 行为，但尚缺真实 `grpc`/`grpcs` 握手、证书信任与 metadata 传递的双链路集成测试，后续补充时应使用本地 CA 与真实 gRPC server。

## Console Agent 资源变更工作台需求调研（2026-07-20）

- [x] 核对 Console 前端入口、现有资源编辑器和创建/发布交互
- [x] 核对后端资源写入、草稿/发布语义、鉴权与审计能力
- [x] 比较至少三种 Agent 编排模块接口与 seam 放置方案
- [x] 明确临时视图、确认、创建为草稿、等待发布的状态机与失败语义
- [x] 形成分期方案、风险与验收标准，并归档长期 ADR
- [x] 同步 context-kg index/log，运行知识库 lint 与 diff 检查

Review：

- 现有“保存后待发布”仅对配置文件与治理规则成立；命名空间、服务、实例、鉴权、MCP 和 A2A 都是直接生效 CRUD，首期只开放查询。
- 推荐采用 Console 后端 `AgentWorkbench` 深模块，外部保持会话、确认和放弃的小 interface，内部用 ConfigFile/Governance 强类型 adapter 隐藏资源差异。
- 模型只能查询和生成结构化 proposal，不能持有资源写工具；确认时重新鉴权、重读 baseline、校验 preview hash 和幂等键，只保存编辑态资源。
- Agent 页面临时视图展示 semantic diff、影响范围、风险和待发布语义；确认成功后返回现有详情/发布页深链，页面本身不提供发布按钮。
- 长期方案已归档到 `technical/adr/adr-console-agent-resource-workbench.md`，并同步 `ai-features`、index 和 log。
- `context_kg_lint.py ./context-kg` 通过：36 个 Markdown 页面、页面名唯一、frontmatter、链接和 index 基础检查全部通过；`git diff --check -- context-kg` 通过。

## Console Agent 工作台 Phase 0 纵向闭环（2026-07-21）

- [x] 固化可运行范围：配置文件单资源提案、临时 diff、确认保存草稿、禁止发布
- [x] 实现 Console 后端内存 proposal 状态机和 pole-server 配置文件 adapter
- [x] 实现 `/ai/agent` 页面、提案预览与确认交互
- [x] 增加后端状态机/路由测试、前端交互契约测试与真实 smoke 脚本
- [x] 完成 lint、前端构建、定向 Go/race 测试及真实 server + Console 页面/API 验证
- [x] 范围化代码审查并记录 Review

Review：

- Phase 0 严格限定为已有配置文件 update；结构化输入是有意的确定性切片，不宣称已接入自然语言模型、create 或治理规则。
- 提案预览不会写草稿；确认前重读并校验 baseline hash，确认时校验 preview hash 与幂等键；成功后只调用配置 PUT 并返回 `waiting_for_publish`。
- 真实 smoke 证明预览前后草稿与 active release 都不变，确认后草稿变化而 active release 保持原值，重复确认返回同一 receipt；浏览器完成登录、生成临时视图、diff、确认、待发布提示和配置分组深链。
- Go 状态机、HTTP adapter、handler/router 测试和 race 通过；前端契约检查、定向 ESLint、test/release 构建通过。
- 当前工作区已有的 `go-control-plane v0.14.0` 升级与 xDS cache 源码不兼容，导致根二进制全量重建失败；本轮使用旧 pole-server 进程加当前 Console 模块完成真实验证，未修改或回退该用户变更。
- pole-server 配置 API 尚无原子 CAS，确认前重读只能缩小而不能消除 TOCTOU；已在页面警告和 ADR 中明确，后续需要 revision/CAS 契约。
- Standards/Spec 双轴审查发现的提案无界驻留、下游故障误报、成功后过期/容量淘汰破坏幂等、确认缺少 version、空批量响应、Request ID 未展示和 smoke 假断言均已修复；复验无遗留阻塞项。

## Console 治理鉴权交互完整收尾（2026-07-20）

- [x] 明确已有自定义 Header 凭证为 write-only，进入编辑态后提示必须重新输入才能轮换，避免空输入框被误解为数据丢失。
- [x] 保证托管身份与自定义 Header 切换时不会复用、预览或回显旧 secret，同时保留合理的 Header 名编辑上下文。
- [x] 补齐专项静态回归断言，覆盖默认托管身份、模式切换、轮换提示、保存校验和实时 Spec 脱敏。
- [x] 通过 Console lint、测试构建和真实浏览器完成“创建、模式切换、保存、详情回显、再次编辑”验收。
- [x] 完成范围化代码审查，并在本节记录 Review、证据和剩余边界。

Review（待代码审查结论补齐）：

- Custom Header 明文已从持久化规则 state 中拆出为一次性 draft；加载、关闭、撤销、模式切换和成功保存都会清空，服务端返回的 `value_sha256` 也会在进入 Console state 前丢弃。
- 提交载荷由独立纯函数按模式重建：托管模式只发送 `managed_identity`，兼容模式只发送 Header 名和本次新值，不会把详情响应中的摘要或历史值扩散回写请求。
- all 模式使用最新 release 静态资源启动成功，8080/8090 均返回 HTTP 200；浏览器实测创建托管规则、切换模式、首次保存、自定义值轮换、详情掩码、再次编辑空值校验和撤销清理均符合预期。
- 真实 PUT 请求只包含本次 `value` 且不含 `value_sha256`；保存后的详情只返回空明文与摘要，页面只显示 `••••••••`，再次编辑为空并提示“重新输入新值以轮换”。
- 右侧实时 Spec 已接入 YAML/JSON 切换：只读已存凭证显示 `<redacted>`，编辑态空 draft 显示 `<required>`，输入新值后只切换为 `<redacted>`，不会渲染明文；历史 legacy 规则保持 `LEGACY_REQUEST_MATCH`，普通编辑保存不会被误迁移。
- 临时规则 `codex-auth-closeout-0720` 已通过 Console 删除，筛选结果为 0；浏览器控制台无 error/warning。截图：`output/playwright/console-auth-custom-header-rotation.png`。
- 已通过 `npm run test:traffic-security`、`npm run lint` 与 `npm run build:test`。
- Standards 与 Spec 两轴审查发现的模式契约、实时 Spec 接线、只读凭证语义和深色对比问题均已修复，最终复核无遗留阻塞项。
## Console 表格统一规范审计（2026-07-20）

- [x] 盘点 Console 全部生产表格及共享表格封装入口。
- [x] 检查首列与尾列固定、表格内部横向滚动和容器宽度自适应。
- [x] 检查长内容省略以及悬浮展示完整内容。
- [ ] 使用真实页面与不同视口抽样验证高频表格：本轮 in-app Browser 初始化失败，未改用未经用户许可的 Playwright。
- [x] 汇总不符合项、共享修复入口与验收建议。

Review：

- `console/web/src` 共 39 个 `<Table>` 实例、分布在 36 个文件，全部经 `components/Fluent` 共享封装；仅 6 个显式同时固定首尾列、4 个只固定一侧，其余 29 个均未固定。
- 共享滚动容器已有 `overflow: auto`，但 fixed layout 会把数字宽度按总权重换算为百分比，表格 `min-width` 又固定为 100%，因此大多数窄容器会压缩列而不是产生可靠的内部横向滚动；声明的 `column.minWidth` 当前完全未应用。
- 字符串和数字会进入单行省略容器，但悬浮全文只覆盖 `ellipsis=true` 且最终内容为原始字符串的情况；大量返回 ReactNode 的自定义 cell 会被截断，却没有完整内容提示。
- 当前只有 A2A 页面通过局部 `table min-width: 1544px` 可靠形成横向滚动，说明页面侧补丁无法替代共享默认规范。
- 统一修复应落在 Fluent Table：默认固定首尾可见列，保留数字宽度并应用 `minWidth`，按列计算 table 最小宽度，在真实溢出时使用统一 OverflowTooltip；复杂 cell 提供 `tooltipText` resolver，并允许编辑型小表显式退出。
- 额外发现：PolicyTable 将渲染函数误写入 `ellipsis` 且列键拼成 `commnet`；NearbyRoute 的主调、被调两列都错误渲染 `priority`。这两项属于确定性页面缺陷，应独立修复。

## Console Agent 页面不可访问恢复（2026-07-21）

- [x] 用 8080/8090 curl 与端口监听复现连接拒绝。
- [x] 确认临时工具会话退出且没有 tmux 接管。
- [x] 用独立 tmux 启动 pole-server 与当前 Console 二进制。
- [x] 验证首页、Agent 深链、静态资源和真实 smoke。
- [x] 沉淀持久后台启动 lesson。

Review：

- 根因不是 Agent 路由或页面代码，而是上一轮进程绑定在临时工具会话，回合结束后被回收。
- 当前 `pole-agent-backend` 与 `pole-agent-console` 两个 tmux 会话分别守护 8090 和 8080；日志位于 `/tmp/pole-agent-backend.log` 与 `/tmp/pole-agent-console.log`。
- `/`、`/ai/agent`、8090 健康检查以及入口 JS/CSS 均返回 200；`npm run smoke:agent-workbench` 通过，草稿确认后保持 waiting-for-publish 且活动发布版本未变化。

## Console 对话式 Agent 工作台重做（2026-07-21）

- [x] 核对当前 Agent、MCP、提案 API 与一级导航接缝。
- [x] 将 Agent 从“AI 工具”子菜单迁移为独立一级入口。
- [x] 以聊天时间线替换配置表单，展示资源操作与工具调用轨迹。
- [x] 在对话中接入临时视图、用户确认、保存草稿和待发布回执。
- [x] 补齐前端契约测试、API smoke 和 8080 真实页面验收。
- [x] 修订 Agent ADR、AI 功能模块页、知识库索引和变更日志。

设计基线：

- 视觉主张：独立的运维对话工作区，主画布是聊天时间线，右侧只在需要时展示资源变更与发布状态。
- 内容结构：会话列表 → 对话与 MCP 调用轨迹 → 临时视图/diff → 用户确认 → 待发布回执。
- 交互主张：工具调用在消息流中逐步展开；预览结果进入右侧检查器；确认后状态平滑切到“已创建草稿，等待发布”。

Review：

- 本轮纠正的核心不是视觉改版，而是产品模型修正：Agent 不能是资源表单，也不属于 AI 资源管理菜单。
- `/agent` 已成为一级单入口；侧栏“AI 工具”只保留 A2A Agent 与 MCP 服务。
- 首屏是 Pole Agent 对话时间线；发送配置变更消息后依次展示资源查询工具轨迹、右侧临时 diff、确认门禁、草稿修改工具轨迹与待发布回执。
- 真实浏览器使用 `default/codex-agent-0720181130/app.yaml` 跑通预览与确认，Request ID 正常回显；测试夹具已删除。
- `npm run test:agent-workbench`、`npm run lint`、`npm run build:test`、`npm run build` 与 `npm run smoke:agent-workbench` 均通过；8080 `/agent` 和 8090 返回 200。
- 当前仍是确定性 planner + 强类型 HTTP tool port。页面表达已经对齐目标产品，但真正让模型通过 MCP client transport 调用 Pole 工具仍属于 Phase 1，不能把本轮结果描述成完整通用 Agent runtime。

## Console 普通视图与 Agent 视图双模式（2026-07-21）

- [x] 定义两种工作模式的路由、导航和上下文交接契约。
- [x] 建立“普通控制台 / Agent”双向工作区切换。
- [x] Agent 模式隐藏普通资源导航，并可返回最近访问的普通页面。
- [x] 配置文件详情增加“交给 Agent”，携带资源自然键和返回地址。
- [x] Agent 接收资源上下文并生成可继续输入的对话起点。
- [x] 用真实页面验证普通 → Agent → 普通的完整往返。
- [x] 更新 ADR、日志、lesson 和 Review。

设计基线：

- 视觉主张：模式切换像固定工作区入口而不是资源菜单，保持暖白表面、细分隔线和单一蓝色强调。
- 内容结构：全局模式 → 当前环境/资源上下文 → 普通资源操作或 Agent 对话 → 返回检查与发布。
- 交互主张：默认侧边布局在固定底部位置双向切换；进入 Agent 时普通资源导航退出；资源上下文在 Agent 会话首屏明确接入。

Review：

- 两种视图共享同一登录态、权限、API 和发布链路，不复制资源实现。
- 默认侧边布局在 Menu footer 固定工作区入口，普通和 Agent 模式对称切换；顶部布局没有侧栏，因此保留 Header 兜底入口。
- `WorkspaceModeSwitch` 记录最近普通页面，显式 `returnTo` 只接受安全本地非 Agent 路径，避免开放重定向和模式循环。
- 配置文件交接 URL 只包含 `kind=config.file` 与 namespace/group/name；Agent 根据上下文生成消息和目标代码块骨架，不把配置正文写入 URL。
- 静态契约、前端 lint 与测试构建已通过；真实 Chrome 点击验证 `/namespace → /agent?returnTo=%2Fnamespace → /namespace` 完整往返，进入 Agent 后资源导航隐藏且底部返回入口保留，返回后普通导航恢复。
- 用户报告无法切回时，真实浏览器首先复现为旧 SPA 仍驻留在已打开标签页；重新加载当前入口后切换逻辑正常。Console 的 `index.html` 已增加 `no-cache, no-store, must-revalidate`，深链 fallback 同样生效，并由路由测试锁定。
- 继续在用户实际 Chrome 中排查后确认更精确的根因：旧主包在模式切换时动态加载已被后续 Vite 构建删除的哈希 chunk，浏览器报 `Failed to fetch dynamically imported module` 与 404；仅禁止 `index.html` 缓存无法修复已经驻留的旧文档。
- 工作模式边界现改为 `window.location.assign` 整页导航，普通控制台内部仍保持 SPA 导航；这样每次跨模式都会重新获取当前 `index.html` 及与之匹配的资源哈希。
- 用户实际 Chrome 已再次跑通 `/namespace → /agent?returnTo=%2Fnamespace → /namespace`：Agent 工作台、模式选中态、普通侧栏和命名空间列表均恢复，DevTools Console 为 0 条运行错误。

## Agent 底部工作区切换（2026-07-21）

- [x] 从普通资源导航和默认页头移除重复的 Agent 模式入口。
- [x] 在普通控制台侧栏底部、版本信息上方增加固定“进入 Agent 工作台”。
- [x] 在 Agent 模式提供对称的固定“返回普通控制台”，保留最近页面回跳。
- [x] 补齐展开、收起、键盘焦点和窄视口状态。
- [x] 更新静态契约并完成真实 Chrome 双向切换与视觉验收。
- [x] 记录 Design QA、Review 与最终验证结果。

设计基线：

- 模式切换属于侧栏固定 footer，不参与资源菜单滚动，也不占用资源信息架构。
- 普通模式显示 Agent 工作台入口；Agent 模式使用同一位置提供返回普通控制台，避免隐藏普通菜单后丢失返回路径。
- 顶部只保留当前页面与全局账户操作；除无侧栏的顶部布局外，不重复展示模式切换。

Review：

- `/agent` 路由保留注册但标记为隐藏，普通资源导航和默认页头不再出现重复的 Agent 入口。
- `WorkspaceModeSwitch` 复用原有安全 `returnTo` 与整页导航逻辑，并新增 sidebar 呈现；普通模式显示“进入 Agent 工作台”，Agent 模式显示“返回普通控制台”。
- Agent 模式保留侧栏骨架，只替换资源导航为最小工作区标识，因此底部返回入口和版本信息始终可见；收起侧栏时入口退化为带 `aria-label` 和 `title` 的图标按钮。
- 真实 Chrome 已跑通 `/namespace → /agent?returnTo=%2Fnamespace → /namespace`，同时验证展开与收起侧栏；DevTools Console 为 0 条消息。
- `npm run test:agent-workbench`、`npm run lint`、`npm run build` 均通过；视觉对比和问题分级记录在根目录 `design-qa.md`。

## Console Agent Runtime 与 LLM Gateway 接缝分析（2026-07-21）

- [x] 核对现有 Agent ADR、前端、Console handler 与 agentworkbench 的真实调用链。
- [x] 核对 LLM、MCP、身份透传、预览确认和发布隔离的配置与接口现状。
- [x] 判断 LLM Gateway 地址和凭证的配置归属，以及 Console 内部 Agent 的最小职责。
- [x] 给出推荐模块接缝、部署形态与分阶段落地顺序。

分析约束：

- 浏览器只承担会话 UI 和用户确认，不直接持有 LLM Gateway 密钥或 Pole 管理凭证。
- Console 内部 Agent 应复用现有权限、审计和资源发布链路，不能复制第二套资源实现。
- 结论必须区分当前确定性 Phase 0 与真正模型驱动的 Agent Runtime，不能把 stub 描述成已接入 LLM/MCP。

Review：

- 当前没有 LLM Gateway 配置或 ModelPort；前端用正则解析固定格式，后端只暴露 config-file proposal/confirm 两个接口。
- 当前 `agentworkbench.Workbench` 是安全执行内核：负责读取、diff、preview hash、TTL、幂等、stale 检查和保存草稿，不是完整 Agent Runtime。
- 页面展示的 `pole.config.get_file`、`pole.config.update_file_draft` 工具轨迹实际走 Console 到 pole-server 的 REST adapter；Pole MCP 当前只注册 namespace 与 MCP Server Registry 工具，因此“Pole MCP 已连接”会误导用户。
- 用户纠正后将主体收敛为一等 `PoleAgent` 深模块；页面只与其会话 interface 交流，System Prompt、LLM、MCP 导入、tool loop、会话和 approval resume 全部由 Agent 内部拥有。
- `ModelPort`、`MCPToolRegistry`、`SessionRepository` 和现有 Workbench 不是页面拼装的并列后端；它们是 Pole Agent 的内部实现，其中 Workbench 定位为 `ChangeApprovalKernel`。
- Console 配置新增 `agent.model.baseURL/model/apiKey/timeout` 与运行限制；密钥只从服务端环境或 Secret 注入，Pole token 只进入受控 ToolExecutor，不进入模型上下文。
- Phase 1 顺序应为：定义 Pole Agent 与版本化 Prompt → 会话/流式 interface → LLM Gateway ModelPort → Agent 内部 MCP 导入和 tool loop → 接回现有确认内核 → 修正页面真实连接状态。

## Console Agent 系统配置来源分析（2026-07-21）

- [x] 核对 Console 启动 YAML、console-only 模式和现有 MySQL Store 职责。
- [x] 比较 YAML、Pole 配置中心与 Console 直连数据库三种方案。
- [x] 定义启动配置、AgentDefinition、动态运行设置和 Secret 的归属。
- [x] 设计“系统设置 / Agent”页面的保存、校验、发布与热加载语义。

Review：

- 纯 YAML 只适合自举和安全兜底，不适合日常模型、Prompt、MCP 和运行限制调整；当前 `LoadConfig` 也没有 reload/watch。
- 不建议 Console 直连 pole-server 数据库。console-only 模式不初始化核心 Store；现有 Console MySQL 只承载 `pole_observability` 的 history/event reader，`server_setting` 仅有 DDL 而无完整业务链。
- 推荐把非敏感 AgentDefinition 存在 `pole-system/console-agent/agent-runtime.yaml`，只消费配置中心已发布版本；模型密钥和 Pole 后台凭证只保存 `secretRef`。
- Console 提供类型化“系统设置 / Agent”页面，底层复用配置中心草稿、发布、历史、回滚、权限和审计，不暴露原始 YAML。
- Agent Watch active release，完整校验后原子替换 last-known-good；新会话绑定配置 revision，现有会话不中途切换。
- 内置安全 Prompt 不允许页面覆盖，可配置的 operator instructions 独立版本化；Agent 工具集禁止修改自身系统配置或读取密钥。

## Pole Server 与 Console 统一系统配置（2026-07-21）

- [x] 盘点 Pole Server 配置加载、模块初始化和现有局部动态能力。
- [x] 盘点 Console 配置、自有 Store、console-only 模式和动态候选项。
- [x] 定义统一 SettingDefinition、配置分级、来源优先级和自举闭环。
- [x] 设计发布、热加载、重启生效、last-known-good 和多实例 apply receipt。
- [x] 设计系统设置页面、权限审计、Secret 和分阶段迁移。
- [x] 新增统一 System Configuration ADR 并同步关联知识页。

Review：

- 统一能力不是“在线编辑整份 YAML”，而是独立 System Configuration 领域；静态文件始终保留自举和安全基线。
- 动态层复用配置中心保留空间与草稿/发布/历史/回滚；Server 和 Console 分别实现自己的 ConfigApplier，不互相修改内部对象或数据库。
- 每个字段必须声明 `BootstrapOnly/RestartRequired/HotReload/GuardedHotReload`；没有显式并发安全 applier 的字段禁止热更新。
- 正常来源优先级为 `default < static YAML < published overlay < emergency allowlist`，动态删除字段表示回落静态值。
- 发布只切换 DesiredSnapshot；实例通过 ApplyReceipt 报告 applied/rejected/pending_restart，页面展示 desired/effective revision 和版本漂移。
- 系统设置按 Server、Console、发布中心和实例状态组织；字段展示来源与 apply mode，BootstrapOnly 只读。
- Phase 1 先做全量配置目录与只读页面；Phase 2 从 AgentDefinition、Console observability timeout、feature flag 等低风险项开始热更新。
## Pole Kubernetes 本地一体化部署（2026-07-21）

- [x] 新增 `pole-system` namespace 下 Pole、GreptimeDB、OTel Collector 的 Kubernetes manifests。
- [x] 通过 ExternalName Service 让 Pod 继续访问宿主机 MySQL `3306`，不把 MySQL 迁入集群。
- [x] 提供可重复执行的本地镜像构建、配置注入和部署脚本。
- [x] 部署到 OrbStack Kubernetes，验证三个工作负载、持久卷、服务发现和健康检查。
- [x] 验证 Console/Agent 页面、Pole API、OTLP 写入与 GreptimeDB 查询闭环。
- [x] 更新部署文档、可观测性 ADR 和知识库索引日志。
- [x] 完成针对性检查、知识库 lint、代码审查并提交本次变更。

设计基线：

- “合并部署”表示同一 Kubernetes 部署栈和 namespace 编排，不把三个有独立生命周期的组件合并到同一个 Pod。
- MySQL 保持宿主机现有实例，集群内使用 `host.docker.internal:3306`；账号密码只进入 Kubernetes Secret。
- GreptimeDB 使用 PVC 持久化，Collector 通过 ClusterIP 写入 GreptimeDB，Pole 通过 ClusterIP 上报和查询。

Review：

- OrbStack `pole-system` 中 Pole Deployment、Collector Deployment、GreptimeDB StatefulSet 均 Ready 且重启数为 0；GreptimeDB 5Gi PVC 已绑定。
- `pole-mysql` ExternalName 解析到 `host.docker.internal`，Pod 内已读取到宿主机 `pole_server` 和 `pole_observability` 两个数据库。
- Pole 有效配置使用 Pod IP 自注册、`pole-otel-collector` 和 `pole-greptimedb` Service DNS，本地 MySQL 连接池降为 20/10。
- 本机 `127.0.0.1:8080` 和 LoadBalancer `172.30.31.2:8080` 的首页、命名空间、Agent、登录页均返回 200。
- 唯一 OTLP smoke event 已通过 Collector 写入 Kubernetes GreptimeDB，并由 `pole_events` SQL 精确查回。
- 旧 Docker GreptimeDB/Collector 和宿主 tmux Pole 已停止但未删除，旧 GreptimeDB volume 可恢复；宿主 MySQL 保持运行。
- 双轴审查发现并推动修复镜像/Secret/config 重部署、YAML Secret 转义、非 root/只读运行、ServiceAccount token、核心 liveness 和启动日志凭据泄漏问题。

## Pole Console Gateway 域名接入（2026-07-21）

- [x] 盘点现有 GatewayClass、Gateway listener、HTTPRoute 和本地域名解析。
- [x] 设计不修改共享 Gateway 的跨 namespace 路由授权。
- [x] 新增 ReferenceGrant、HTTPRoute 并接入本地部署脚本。
- [x] 验证 Route Accepted/ResolvedRefs、普通控制台和 Agent 深链。
- [x] 完成文档、知识库 lint、代码审查和提交。

设计基线：

- 复用 `tidemind/tidemind-gateway`，不重复部署 Ingress Controller，也不修改共享 listener 的 `allowedRoutes`。
- `HTTPRoute` 与 Gateway 同处 `tidemind`；`pole-system` 通过最小范围 `ReferenceGrant` 只授权访问 `pole-control-plane` Service。
- Gateway 只路由 Console `8080`；Pole 的 LoadBalancer 多协议端口暂保留给本地 SDK/协议调试，GreptimeDB、Collector 和 MySQL 保持集群内部访问。

Review：

- `tidemind/pole-console` 已被 Envoy Gateway 接受，`Accepted=True`、`ResolvedRefs=True`，共享 Gateway attachedRoutes 从 4 增至 5。
- `http://pole.localhost/`、`/namespace`、`/agent`、`/login` 均返回 200，无需修改 hosts。
- OrbStack 的 `*.k8s.orb.local` 当前对既有 `admin` Route 也会断连，因此未作为交付入口；只保留已实测稳定的 `.localhost` 域名。
- ReferenceGrant 只允许 `tidemind` namespace 的 HTTPRoute 引用 `pole-control-plane` Service，没有扩大到其它 Service 或 namespace。
- 双轴审查无阻塞项；已将 ReferenceGrant 文案修正为“按来源 namespace 和目标 Service 收敛授权”，并明确 Pole LoadBalancer 仅保留为本地多协议调试边界。

## Console 登录页视觉重设计（2026-07-22）

- [x] 审计当前登录页的信息层级、表单状态、响应式和视觉问题。
- [x] 收敛为产品说明区与聚焦登录面板，移除默认账号裸露提示。
- [x] 强化主按钮、输入焦点、密码可见性和登录提交态。
- [x] 完成构建、真实页面截图、窄视口与登录功能回归。
- [x] 完成双轴代码审查并修复阻塞项。
- [ ] 待全站 Fluent UI 迁移形成可提交基线后提交登录页改动，避免生成依赖未提交组件的破损 commit。

设计基线：

- 登录是控制台入口，不是营销落地页；标题和背景必须让位于账号输入与主操作。
- 普通控制台与 Agent 只作为产品能力说明，不在登录页复制模式切换或资源导航。
- 保持现有认证 API、管理员检查和登录后跳转不变；移除没有后端注册 API 的伪“创建账号”入口。
- 后台管理员检查异常只提示重试，不得把网络或 5xx 错误误判为“尚未初始化”并跳转首管理员创建页。

Review：

- 标题上限由 64px 收敛到 42px，移除营销式三栏能力块；桌面认证区权重高于说明区，390px 下改为完整单列。
- 默认账号明文和没有后端 API 的注册入口已移除；账号、密码、可见性按钮具有可访问名称，空提交校验可见。
- 管理员检查异常不再跳转 `/init-admin`；只有接口明确返回不存在管理员时才进入初始化流程。
- 登录与初始化管理员页面均补齐自动填充语义、密码可见性按钮和可访问名称；初始化检查失败会安全返回登录页，不再暴露首管理员创建表单。
- `npm run lint`、`npm run build`、知识库 lint、`git diff --check` 通过；Playwright 对登录与初始化管理员页面完成桌面及 390×844 验证，无溢出且 Console warning/error 为 0。
- 全量 `test:no-tdesign` 仍被本任务前已有的 WorkspaceModeSwitch/Agent 原生按钮残留阻塞；本次登录文件未新增原生交互标签或 TDesign 依赖。
- 当前 HEAD 尚未包含工作树中的全站 Fluent 适配层与依赖迁移；因此未把登录文件单独提交成不可独立构建的 commit。

## Console K8s 登录页更新核验（2026-07-22）

- [x] 核对 Deployment 镜像、运行 Pod imageID、启动时间与健康状态。
- [x] 比对本地构建、Pod 内静态入口和 Gateway 响应的 SHA-256。
- [x] 比对域名、localhost、127.0.0.1 与 LoadBalancer 四个入口的资源哈希。
- [x] 显式重启 Deployment，并在新 Pod 上复验页面资源。

Review：

- 新 Pod `pole-control-plane-6b47675975-5m45f` 使用镜像 `sha256:65238729977b...`，Ready 且重启数为 0。
- 本地、Pod 和 Gateway 的 `index.html` SHA-256 均为 `e2c7ed3614ab...`，四个入口均引用新版 `assets/index.e47e79bc.js`。
- SPA 入口返回 `Cache-Control: no-cache, no-store, must-revalidate`；服务端与 Gateway 均已更新，若旧标签仍显示旧界面，应重新加载文档以退出旧 JS 运行时。

## System Configuration Phase 1（2026-07-22）

- [x] 盘点 Pole Server 与 Console 首批核心有效配置结构、来源和敏感字段。
- [x] 定义只读 `SettingDefinition` registry 与 effective snapshot interface。
- [x] 实现 Console 只读查询接口，返回分组、类型、apply mode、敏感级别、有效值与来源。
- [x] 新增独立“系统配置”侧栏入口和只读页面，支持概览、组件/子域筛选与来源说明。
- [x] 为 registry、接口契约和页面映射补充专项测试。
- [x] 完成 lint、构建、Go 测试、真实页面与窄视口验证。
- [x] 更新 System Configuration ADR 状态、知识库索引日志与 review 记录。
- [x] 完成双轴代码审查并处理阻塞项。

设计基线：

- Phase 1 不改变任何配置加载、发布或热更新行为；静态 YAML/环境变量仍是当前有效来源。
- 页面只消费类型化只读快照，不读取或渲染整份原始 YAML。
- Secret 值必须脱敏；页面只展示存在性或引用来源，不返回密钥正文。
- 配置定义、配置读取和 Console 展示通过一个稳定 interface 衔接，字段映射集中在 registry，不散落到 handler 和页面。

Review：

- 首批 registry 当前共 63 项：Pole Server 29、Console 34；页面明确标记为“首批已注册核心配置”，不反射插件任意 `Option`，避免目录覆盖范围误导。
- Console `/system-config/v1/settings` 只接受 GET，匿名请求返回 401，再通过网络读取 Pole Server；Server `/admin/v1/system/configuration` 自身通过 `DescribeSystemConfiguration` 方法权限再次执行策略授权，不能绕过 Console 直接匿名读取。
- 来源索引能区分编译默认值、静态 YAML、环境变量和 CLI 覆盖；复合 DSN 记录全部环境变量引用，归一化回退字段标记为编译默认值。
- 5 个 Secret 字段在服务端清空 `value` 并仅返回配置状态与引用；K8s 实测响应没有密钥正文。
- 专项 Go 测试、ESLint 和 Vite production build 通过；K8s 新 Pod 使用本次镜像且 Ready/0 restart，Gateway 鉴权接口返回 63 项。
- 双轴审查提出的内部接口鉴权、独立读取权限、来源失真、来源筛选统计和目录覆盖误导均已处理；页面不提供编辑、草稿或发布动作。
- 已使用测试管理员完成真实登录态验收：桌面 1280px 下页面无整页横向溢出，搜索 `server.bootstrap.mode` 后结果为 1/52；390x844 窄视口下文档宽度保持 390px，摘要区按两列重排，配置表以 273px 可视宽度承载 1030px 内容并可在表格内部滚动，Secret 仅显示“已配置”和环境变量引用，不出现密钥正文。

## System Configuration 领域化页面（2026-07-22）

- [x] 明确组件、领域、配置项三级信息架构与窄屏降级方式。
- [x] 实现 Pole Server / Console 组件切换和领域导航。
- [x] 将单一总表改为当前领域配置表，同时保留跨领域搜索与来源筛选。
- [x] 完成 ESLint、production build 和真实桌面/390px 页面验证。
- [x] 构建部署到当前 Kubernetes，并复验 Gateway 资源哈希。

设计基线：

- 总览只负责说明当前组件的配置规模和来源分布，不承载全部配置行。
- 领域是配置浏览和后续草稿/发布的工作单元；主表不再重复展示组件和领域列。
- 搜索覆盖当前组件全部领域，领域导航实时展示命中数并自动定位首个有结果的领域。
- 窄屏将领域导航保持为可横向滚动的紧凑入口，表格只在自身容器内横向滚动。

Review：

- Pole Server 依次展示启动与装配、命名空间、服务发现、配置中心、缓存同步、存储、认证授权和工作负载凭证；Console 依次展示运行时、日志、认证授权、上游连接、存储、可观测查询和 Agent。
- 跨领域搜索 `console.agent.upstream_timeout` 会把 Console 从“运行时”自动定位到“Agent”，领域命中显示为 1/2，表格仅保留命中的一行；桌面领域列表和移动端横向导航都会把 Agent 自动滚入视野。
- 1280px 和 390x844 真实页面均无整页横向溢出；390px 下领域导航宽 275px、内容宽 645px，配置表可视宽 247px、内容宽 880px，两者均只在自身容器滚动。
- ESLint 与 Vite production build 通过；Kubernetes Pod `pole-control-plane-7b54865989-g8xg9` Ready、0 restart，运行镜像 ID 为 `sha256:3763b3518a06...`，本地与 Gateway 均引用 `assets/index.9b7513c5.js`。

## Console AgentDefinition 配置目录（2026-07-22）

- [x] 明确 Agent 自举配置、动态定义和 Secret 的职责边界。
- [x] 增加 Agent ID、Prompt 版本、LLM Gateway、模型、API Key、MCP 与运行限制配置结构。
- [x] 将 Agent 配置显式注册到 System Configuration，并保证 API Key 服务端脱敏。
- [x] 让现有 Agent Workbench 使用配置化 proposal TTL 与资源工具上游超时。
- [x] 补充配置加载、来源和 Secret 泄漏测试。
- [x] 完成构建、Kubernetes 部署和真实 Agent 领域页面验证。

设计基线：

- 页面必须同时显示当前 `deterministic` runtime 和目标 AgentDefinition，不能因为登记了 LLM 地址就宣称 ModelPort 已接通。
- LLM API Key 最终由 Pole `SystemSecretStore` 托管；当前环境变量注入仅是 Phase 1 自举兼容。页面响应只返回“已配置/未配置”、版本和 Secret 引用。
- AgentDefinition 归 Console 后端拥有，pole-server 核心只提供受控 MCP/管理能力，不持有模型密钥。

Review：

- Console Agent 领域从 2 项扩展为 13 项，Console 合计 34 项、两组件合计 63 项；页面同时展示 Agent ID、Prompt、LLM、MCP、运行限制和 `deterministic` 当前模式。
- `console.agent.model_api_key` 的 Gateway 鉴权响应为 `redacted=true`、`configured=false`、无 `value` 字段，只保留 `env:POLE_AGENT_LLM_API_KEY` 引用；LLM Gateway 地址同样明确显示未配置。
- YAML 列表来源聚合已修正，MCP 工具白名单在真实页面显示为“静态配置”，不再误标为编译默认值。
- `go test ./...`、前端 ESLint/production build、Kubernetes manifest dry-run 与 context-kg lint 均通过；最终 Pod `pole-control-plane-8597d4c968-gmlh8` 使用镜像 `sha256:66d4a28e3ee4...`，Ready、0 restart。
- Playwright 通过 `pole.localhost` 真实登录态验收：Agent 13/13 行完整可见，API Key 仅显示未配置和 Secret 引用，页面未显示任何密钥正文。

## Pole 托管 Agent Secret 纠偏（2026-07-22）

- [x] 核对现有配置文件加密链与 Agent API key 当前来源。
- [x] 明确 Pole 托管业务 Secret、基础设施只托管 KEK 的目标边界。
- [x] 修正 System Configuration 与 Agent ADR、任务基线和经验规则。
- [ ] 实现 `SystemSecretStore` 深模块、密文表、envelope encryption 与版本化引用。
- [ ] 实现 Agent Secret write-only 页面、轮换/禁用/连接测试和独立权限审计。
- [ ] 将 Agent Runtime 从环境变量引用迁移为受控 `pole-secret://` resolve interface。

Review：

- 当前配置文件加密链将 DEK 与密文共同保存在配置 metadata，并会对配置读取者解密正文，不满足系统 Secret 的 write-only、KEK 包装、用途约束和轮换要求，不能直接复用。
- 目标 interface 由 Pole 隐藏密文、wrapped DEK、版本和审计实现；Console 与 AgentDefinition 只处理 SecretReference，浏览器永远不读取明文。

## System Configuration 编辑保存交互与 Admin 门禁（2026-07-22）

- [x] 明确有效值、草稿、发布和 Secret 新版本的页面状态与操作顺序。
- [x] 为路由元数据增加 admin-only 能力，非 `main/admin` 用户不展示系统配置菜单且直达时拒绝进入。
- [x] 为 `/system-config/v1/*` 增加服务端主账号或内置 admin 角色校验，覆盖查询和后续写接口。
- [ ] 将配置表升级为可选中配置项的编辑面板，区分普通值、列表和 write-only Secret。
- [x] 设计未保存、已保存草稿、待发布、待重启和已生效反馈，不把“保存”伪装成“生效”。
- [x] 补充前后端权限与交互契约测试，完成真实 admin 页面验证。

设计基线：

- 保存只创建领域草稿；发布是独立动作，`RestartRequired` 发布后仍需明确显示待重启。
- Secret 编辑不回填旧值；留空表示保持当前版本，只有显式输入并确认才创建新 Secret 版本。
- 前端菜单隐藏只是体验优化，后端必须以签名会话中的 `main/admin` 身份做最终拒绝。

Review：

- 主账号登录响应使用规范角色 `main`；内置 `admin` 角色成员使用 `admin`。两者均写入 Console 签名 JWT，不能仅按用户名判断。
- 前端对非 `main/admin` 用户隐藏菜单并阻止直达，`/system-config/v1/*` 在后端再次校验；测试覆盖未登录 401、普通子账号 403，以及 `main/admin` 分类。
- 真实 Kubernetes 页面验证：`pole-control-plane-5f578d99c6-2t5kj` 使用镜像 `sha256:dfcbb20fc65c...`，admin 页面显示 Pole Server 29 项、Console 34 项且无 execute exception。
- 本轮完成编辑生命周期交互设计和 Admin 门禁；真实编辑面板仍等待领域草稿、Secret 新版本与发布 API 后接入，不提供浏览器本地伪保存。
## 三个内置系统角色与成员绑定（2026-07-22）

- [x] 固化 Admin、资源全读、资源全写三个系统角色的稳定 ID、名称与权限边界。
- [x] 在服务启动和角色查询路径中幂等补齐三个系统角色及固定策略，避免已有 MySQL 环境继续显示 0 条。
- [x] 在角色 REST API 拒绝创建额外角色、删除系统角色，以及修改系统角色名称/权限定义；只允许更新用户和用户组绑定。
- [x] 将 Console 角色页改为固定角色目录，移除新建、删除、批量选择和可变基础字段，只保留查看与成员绑定。
- [x] 增加服务端行为测试和前端契约测试，并在真实 Kubernetes 页面验证三条角色及不可变交互。
- [x] 更新领域规则、认证模块知识与 Review，运行全量相关测试、context-kg lint 和 diff 检查。

设计基线：

- `admin` 负责控制面管理能力；`resource-reader` 可读取全部业务资源；`resource-writer` 可读取和写入全部业务资源，但不获得认证管理和系统配置能力。
- 三个角色属于系统目录，不是可自由增删的租户资源；角色 ID、名称、描述和权限集合固定，管理员只维护用户/用户组成员关系。
- “资源全部角色”按上下文收敛为“资源全读角色”，与“资源全写角色”形成最小权限梯度。

Review：

- 服务端以稳定 ID 幂等创建 `admin`、`resource-reader`、`resource-writer` 及固定策略；角色列表只返回这三条，创建、删除、基础字段变更和系统策略变更均由后端拒绝，成员绑定继续复用角色更新接口。
- Console 角色目录移除新建、删除和批量选择，只保留查看与“管理成员”；成员管理通过可寻址查询参数恢复，抽屉只提交 `users` 与 `user_groups`，角色名称、描述和权限定义不可编辑。
- 回归通过：相关 Go 包测试、Console ESLint、`test:system-role-directory`、生产构建、context-kg lint 与定向 diff check；此前完整 `go test ./... -count=1` 也已通过。
- Kubernetes 定向发布到 `pole-control-plane-7d6478bd58-jzxl2`，镜像 `sha256:a42b9db8c33a...`；Gateway 返回 200，真实 admin 页面显示三条系统角色，并验证 `admin` 成员抽屉仅包含用户和用户组绑定。
- 补充修复 MySQL 角色主体读取事务未结束导致的连接泄漏；发布后连续采样 `pole_server` Sleep 连接数为 `10/10/10`，未再随缓存刷新增长。

## 自定义角色 CRUD 与内置角色保护纠正（2026-07-22）

- [x] 恢复自定义角色的创建、查询、修改和删除，并保持既有权限关联能力。
- [x] 将三个内置角色识别为受保护子类型：只允许修改用户与用户组绑定。
- [x] 在 Console 恢复自定义角色的新建、编辑和删除入口，内置角色只展示成员管理入口。
- [x] 锁定内置角色名称、描述、标签及资源/API 权限，补齐前后端契约测试。
- [x] 交叉编译 Linux/arm64 镜像并仅发布到 Kubernetes，发布后完成 API 与真实页面验收。

验收矩阵：

| 角色类型 | 创建 | 名称/描述/标签 | 用户/用户组 | 资源/API 权限 | 删除 |
| --- | --- | --- | --- | --- | --- |
| 自定义角色 | 允许 | 允许 | 允许 | 允许 | 允许 |
| 三个内置角色 | 系统幂等创建 | 禁止 | 允许 | 禁止 | 禁止 |

Review：

- 服务端恢复自定义角色创建、全量查询、定义/成员更新与删除；MySQL 更新同时持久化自定义角色名称。三个稳定 ID 的内置角色仍拒绝改定义和删除，只写入成员绑定。
- 策略创建、策略更新和直接资源授权均拒绝把内置角色作为授权主体；Console 策略编辑器不再把内置角色放进可授权角色选项。
- Console 角色表恢复“新建角色”、自定义角色选择/编辑/删除；内置角色只显示“查看/管理成员”，成员抽屉明确锁定名称、描述、标签和资源/API 权限。
- 发布后回归通过：相关 Go 测试、Console 契约、ESLint 与测试构建；真实 API 完成自定义角色创建、改名/改描述/改标签、删除，内置角色改名/删除/新建授权策略均返回拒绝，测试资源已清理。
- Kubernetes Pod `pole-control-plane-64b5d5dfbc-h5psd` 使用 Linux/arm64 镜像 `sha256:47f21b1e7000...`，Ready 且 0 restart；真实 admin 页面显示新建入口，自定义角色行为为“查看/编辑/删除”，内置角色为“查看/管理成员”，浏览器控制台 0 error。
## 系统配置目录 0/0 回归修复（2026-07-22）

- [x] 建立真实 `/system-config/v1/*` 请求与页面目录非空的确定性回归检查，复现截图中的 `0 / 0`。
- [x] 核对接口响应、Pod 日志、配置定义注册和启动装配，验证候选根因。
- [x] 在正确边界补回归测试并实施最小修复。
- [x] 运行相关 Go/Console 测试与生产构建，定向发布 `pole-control-plane`。
- [x] 用 admin 真实浏览器复验领域、配置项和统计非空，更新 Review 与 lessons。

Review：

- 截图中的 `0 / 0` 不是配置定义丢失；新 admin 页面和后端当前稳定返回 Pole Server 29 项、Console 34 项。根因是前端请求异常分支执行 `setSettings([])`，把 rollout 瞬态失败伪装成了合法空目录。
- 修复后首次加载失败显示“系统配置目录加载失败”和重试；已有快照刷新失败继续显示 last-known-good 数据并告警，不再将统计清零。
- `test:system-configuration-error-state` 已先复现失败后转绿；Console ESLint/生产构建，以及 handlers、router、Console/Server 配置注册表相关 Go 测试均通过。
- 已定向发布 Pod `pole-control-plane-945c9f567-94m4b`、镜像 `sha256:c9d662e39eda...`；Gateway 200，真实 admin 页面复验 29/34、启动与装配 3/3，未出现 0/0。
## 非 Admin 完全隐藏系统配置页面（2026-07-22）

- [x] 增加服务端签名会话角色接口和非 admin 前端 fail-closed 契约。
- [x] 前端启动时不再信任 localStorage 角色，鉴权完成前不渲染受保护页面。
- [x] 非 admin 隐藏系统配置菜单，直达 `/system-configuration` 重定向到普通控制台。
- [x] 运行测试与生产构建，定向发布 Kubernetes。
- [x] 使用 admin 与非 admin 真实会话验证入口和直达行为，补齐 Review。

Review：

- 根因是菜单和 `AdminRoute` 虽已有 admin-only 分支，但角色从 `localStorage.login-role` 恢复，旧值或篡改值会在前端放行页面。
- 新增签名会话角色接口；刷新时 Redux 角色初始为空，受保护路由在 `sessionResolved` 前只显示加载态，菜单默认隐藏。解析为非 admin 后 `/system-configuration` 在组件渲染前跳转 `/namespace`。
- 服务端 session/route 测试、前端 admin gate 契约、ESLint 和生产构建通过；全部运行验收在发布 Kubernetes 后执行。
- K8s Pod `pole-control-plane-6466b7d4cb-88wx2` 使用 Linux/arm64 镜像 `sha256:8c37127d75b5...` 并单副本 Ready。main 会话系统配置接口为 200、临时 sub 为 403；测试用户删除并确认剩余 0。
- 宿主机不存在 `pole-server` 或 `go run` 进程；8080/8090 监听归属 OrbStack Kubernetes 网络入口。
## Console 暗色主题全站一致性修复（2026-07-23）

- [x] 从当前 Kubernetes Gateway 页面建立暗色表面误用浅色背景的红灯检查，并复现服务详情页问题。
- [x] 盘点全局主题变量、共享资源布局与页面级硬编码，定位 3–5 个候选根因并验证。
- [x] 在共享主题/组件边界实施最小修复，补充能锁定浅色表面泄漏的回归约束。
- [x] 在容器环境完成前端检查与生产构建，构建 Linux/arm64 镜像并仅更新 `pole-control-plane` Deployment。
- [x] 基于新 K8s Pod、Gateway 与真实浏览器抽检主要页面，记录镜像、Pod、控制台与视觉结果。

Review：

- K8s 真实服务详情页红灯检查在 `theme-mode=dark` 下捕获详情头、4 张统计卡和基础信息区共 6 个纯白表面；根因是页面级 `--console-*` 固定浅色变量覆盖了全局主题，而既有脚本只检查正文颜色。
- 服务详情改为继承 `--app-*`；全局新增品牌、青色、紫色及成功/警告/错误的 subtle/border 双主题 token，并将治理、AI、认证、监控和登录页残留浅色背景/边框收敛到这些 token。
- `test:dark-theme` 现在解析背景、边框和局部主题变量，拒绝高亮浅色硬编码；K8s Linux/arm64 Node Pod 内完成依赖安装、门禁、ESLint 和 production build。
- 后端无代码变化，复用与旧 Pod 内 SHA-256 完全一致的 Linux/arm64 `pole-server`；新镜像 `sha256:190e72f2e01e...` 仅发布到 `deployment/pole-control-plane`，Pod `pole-control-plane-5b847c7794-sfzlg` Ready、0 restart。
- Pod 内与 Gateway 的入口资源均为 `assets/index.c4d47adb.js`；真实浏览器复验服务详情红灯转绿，并抽检命名空间、服务、配置、治理、监控、认证、A2A、MCP 和 Agent 共 20 个路由，均为暗色主题、无高亮浅色表面，控制台 0 error/warn。
## 配置文件详情页 Fluent UI 布局优化

- [x] 将跨环境实例切换收敛为紧凑上下文条，避免单实例占据大面积卡片
- [x] 重组文件身份、状态、操作、元数据与标签的信息层级
- [x] 将配置正文区域改为紧凑工作台，并让 Monaco 编辑器跟随全局明暗主题
- [x] 增加专项静态回归，完成 ESLint、暗色主题检查与 release build
- [x] 仅重建发布 control-plane，并使用真实页面验证布局、滚动与明暗主题

Review：

- 配置分组与配置文件两级环境切换均改为紧凑上下文条，单环境不再占据大面积空卡片；当前环境只使用品牌浅色边框和表面，不再整块高饱和填充。
- 文件名、格式、加密与发布状态、资源路径及主次操作合并到统一摘要头；修改时间、创建时间、加密算法采用三列定义网格，长标签独占底行以避免截断。
- 配置正文改为稳定高度的编辑工作台，标题直接展示格式与只读/编辑状态；Monaco 根据全局 `ETheme` 在 `vs` 与 `vs-dark` 间切换，暗色实测背景 `rgb(30, 30, 30)`、文字 `rgb(212, 212, 212)`。
- `test:config-file-detail-layout`、跨环境视图专项、目标文件 ESLint、暗色主题检查和 release build 通过；全局 `test:no-tdesign` 仍被本任务外 6 处既有原生 `<button>` 门禁失败拦截，本次未扩大修改范围。
- 已仅更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-config-file-detail-v2`；Pod `pole-control-plane-79d8945686-gsl4c` Ready、0 restart，imageID 为 `sha256:1ae0cfcfd109...`，8080 主资源与 release 构建均为 `assets/index.1999739b.js`。
- 真实浏览器在 1800×1100 下验证亮色与暗色页面均无横向页面溢出和控制台错误；截图为 `output/playwright/config-file-detail-light-v2.png`、`output/playwright/config-file-detail-dark-v2.png`。

## 配置文件详情信息渐进披露

- [x] 将环境版本由卡片切换器改为 Fluent 风格环境 Tab
- [x] 将文件内容、基本信息、发布记录、订阅查询合并为同一层资源 Tab
- [x] 文件身份与操作固定展示，非当前 Tab 内容不同时展开
- [x] 完成专项检查、ESLint、release build 和亮暗主题真实页面验证
- [x] 仅更新 control-plane 运行实例并记录 Review

Review：

- 配置文件环境版本改为具备 `tablist/tab/aria-selected` 语义的横向 Tab；选中文件后隐藏重复的配置分组环境条，页面只保留一组与当前文件直接相关的环境版本。
- 外层“配置编辑”Tab 已移除，`FileView` 统一管理“文件内容 / 基本信息 / 发布记录 / 订阅查询”四个同级 Tab；文件身份、状态和主操作固定，正文、元数据与表格不再同时展开。
- 文件内容默认展示并占据剩余工作区；基本信息以有留白的两列 definition grid 展示描述、时间、加密和标签；发布记录与订阅查询沿用原业务组件和请求生命周期。
- 已通过 `test:config-file-detail-layout`、跨环境专项、暗色主题检查、目标文件 ESLint、release build 与 `git diff --check`；全局 `test:no-tdesign` 的 6 处任务外既有原生按钮问题仍未扩大处理。
- 真实浏览器逐一切换四个资源 Tab，发布记录展示 1 条当前全量，订阅查询展示 0 条空态；亮色、暗色均无控制台错误或警告，1800px 视口下 `pageWidth=viewportWidth=1800`，暗色 Monaco 为 `vs-dark`。
- 已仅更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-config-file-tabs-v2`；Pod `pole-control-plane-5b89cb6d77-zsn85` Ready、0 restart，imageID 为 `sha256:e23d08b14945...`，8080 与 release 构建入口均为 `assets/index.260d970a.js`。
- 验收截图：`output/playwright/config-file-tabs-content-light.png`、`output/playwright/config-file-tabs-basic-light.png`、`output/playwright/config-file-tabs-content-dark.png`。
## Pole Agent 本地会话与 ChatGPT 式工作台重构（2026-07-23）

视觉主张：安静、连续、以对话为中心；会话列表是导航，消息流是主画布，临时视图只在需要确认时出现。

内容计划：本地会话栏 → 当前会话标题与资源上下文 → 消息流 → 浮动输入框 → 按需出现的变更检查器。

交互主张：新会话即时进入；切换会话恢复消息和输入草稿；记忆开关与历史窗口按会话持久化；消息、侧栏和检查器使用克制的进入与布局过渡。

- [x] 建立 IndexedDB 会话存储，覆盖会话、消息、输入草稿、资源上下文和记忆设置。
- [x] 实现会话创建、切换、自动标题、重命名、删除及刷新恢复。
- [x] 将主工作区重构为 ChatGPT 式会话栏、消息流和浮动输入框，消除当前横向溢出。
- [x] 实现每会话记忆开关和历史窗口，并让当前确定性解析真实消费所选历史。
- [x] 保持 MCP 工具轨迹、临时视图、确认保存草稿和用户发布边界。
- [x] 在 K8s Pod 内完成契约、暗色、Lint 和 production build，仅发布 `pole-control-plane`。
- [x] 在 K8s Gateway 真实浏览器验证会话持久化、切换、记忆、响应式及提案流程。

Review：

- Agent 工作台改为左侧本地会话导航、中央连续消息流、底部浮动输入框和按需变更检查器；桌面 1280px 实测 `scrollWidth = clientWidth = 1280`，不再出现页面横向滚动。
- `pole-agent-workbench` IndexedDB 分离 `sessions` 与 `metadata`，持久化消息、未发送草稿、资源引用、记忆开关/窗口和当前会话；浏览器实测创建第二会话、切回首会话并刷新后，草稿“K8s IndexedDB 恢复验证”和最近 4 轮设置均恢复。
- 会话记忆提供关闭与 4/10/20 轮档位；确定性解析先消费当前消息，只在上下文不完整时使用当前会话历史，避免旧意图覆盖新指令。浏览器本地历史不保存或恢复服务端 proposal，不改变 preview/confirm/待用户发布的安全边界。
- K8s Node Pod 内通过 `test:agent-workbench`、`test:dark-theme`、ESLint 和 release build；视觉回归发现 Fluent Textarea 外层未撑满后修复，并再次完成同一组容器检查。
- 仅滚动更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-agent-sessions-v2`；GreptimeDB 与 OTel 未更新。真实 Gateway 登录页面会话切换、记忆面板、刷新恢复均正常，浏览器控制台 0 error/warn。

## Pole Agent 单侧栏交互收敛（2026-07-23）

视觉主张：单侧栏、单画布、单输入焦点；移除产品侧栏与页面会话栏的重复层级。

内容计划：品牌 → 会话管理 → 返回普通控制台；主画布只保留会话标题、消息流、输入框和按需临时视图。

交互主张：会话切换只更新主画布；操作按钮悬浮显现；变更检查器仅在生成提案时展开。

- [x] 让最左侧现有 Agent 导航区域直接承载 IndexedDB 会话管理。
- [x] 删除 Agent 页面内部会话栏、重复新会话入口和空态建议按钮。
- [x] 收敛主画布宽度、标题栏和输入区，保持会话记忆与 MCP 安全链路。
- [x] 更新 Agent 契约与架构记录。
- [x] 在 K8s Pod 内完成前端检查、构建和 context-kg lint。
- [x] 仅发布 `pole-control-plane`，通过 Gateway 真实浏览器验证布局、会话管理和刷新恢复。

Review：

- Agent 模式的全局 `Menu` 提供唯一会话宿主，页面通过 Portal 把 IndexedDB 会话管理渲染到最左侧；原“Agent 工作台”占位和内容区内嵌会话栏均移除。
- 新建会话入口只保留在最左侧一次；空态移除三块建议按钮，会话项移除拥挤的时间副文案，主区只保留标题、MCP 状态、记忆设置、消息流与输入框。
- 输入区去除 Fluent Textarea 的内层描边，仅保留外部输入容器；真实暗色截图呈现为一条侧栏和一个完整对话画布。
- K8s Node Pod 内两轮通过 Agent 契约、暗色 token、ESLint 和 release build；真实 Gateway 页面从最左侧创建第二会话、自动标题、切回旧会话并刷新后仍恢复，测试会话已清理。
- 最终浏览器实测 `sessionHosts = 1`、`scrollWidth = clientWidth = 1280`、控制台 0 error/warn；仅滚动更新 `pole-control-plane`，未更新 GreptimeDB 或 OTel。

## Pole Agent 参考稿布局对齐（2026-07-23）

视觉主张：以 `lattice-agent-workbench.html` 为唯一布局基准，保持 Lattice.Hub 暗色设计系统，同时不削弱现有 Agent 真实能力。

内容计划：280px 会话侧栏 → 52px 全局栏 → 62px 对话标题栏 → 对话画布与底部浮动输入器 → 294px 可收起上下文面板。

交互主张：会话支持搜索、日期分组和快捷新建；资源上下文、权限与记忆设置集中到右侧面板；MCP 提案仍按预览、确认保存草稿、用户发布的链路执行。

- [x] 测量参考 HTML 的布局、层级、响应式和关键交互状态。
- [x] 重构 Agent 会话侧栏、全局栏、对话画布、输入器与上下文面板。
- [x] 保持 IndexedDB 会话/记忆和 MCP 临时视图安全链路，并更新契约测试。
- [x] 更新 Agent 架构记录、设计 QA 报告和经验规则。
- [x] 在 K8s Pod 内完成前端契约、暗色、Lint、构建和 context-kg lint。
- [x] 仅发布 `pole-control-plane`，通过 Gateway 真实浏览器完成同视口视觉对比与交互验证。

Review：

- Agent 模式使用 280px 产品侧栏、52px 全局栏、62px 对话标题栏和 294px 上下文面板；1440×900 实测页面 `scrollWidth = clientWidth = 1440`，没有横向溢出或重复会话宿主。
- 会话侧栏新增搜索、今天/昨天/更早分组、摘要、时间和 `Command/Ctrl + N`；空态按参考稿恢复三类任务起点，但只写入真实输入器，不伪造模型结果。
- 资源范围、运行模式、可用工具、4/10/20 轮记忆和操作权限集中到右侧上下文面板；临时 diff 改为消息流内工具结果，确认仍只保存草稿并等待用户发布。
- K8s Node Pod 内通过 Agent 契约、暗色约束、ESLint 和 production build；K8s Python Pod 内 context-kg lint 通过。
- 仅滚动更新 `pole-system/deployment/pole-control-plane` 到 `pole-control-plane:local-20260723-agent-reference-layout-v1`；新 Pod 1/1 Ready、0 restart，Gateway `/agent`、入口资源与健康接口均为 200。
- 真实浏览器验证任务起点、搜索空态、记忆切换、上下文收放和 IndexedDB 刷新恢复，控制台 0 warning/error；测试会话、输入草稿和记忆窗口已恢复清理。

## Pole Agent 真实 LLM + MCP 可用闭环（2026-07-23）

目标：配置 LLM Gateway 后，用户可以在 `/agent` 与 Console 内部 Pole Agent 真实对话；Agent 使用服务端版本化 System Prompt 和白名单 Pole 工具完成查询与变更提案，写操作仍必须经过不可绕过的临时视图、用户确认、保存草稿并等待用户发布。

- [x] 冻结 PoleAgent、ModelPort、ToolPort、会话与审批恢复的最小接口，并补齐失败态契约测试。
- [x] 实现 OpenAI-compatible LLM Gateway 调用、服务端 System Prompt 和有上限的 model-tool loop。
- [x] 将配置查询与配置变更提案接入 Agent 工具面，保证模型无法确认草稿或发布资源。
- [x] 用真实 Agent 会话 API 替换前端本地正则意图解析，保留 IndexedDB 会话和每会话记忆窗口。
- [x] 未配置模型、模型超时、非法工具、工具失败时返回可诊断状态，页面不得显示虚假的“已连接”。
- [x] 构建不可变镜像，仅在 Kubernetes 中完成单测、构建、接口、浏览器和安全边界验收。
- [x] 更新 ADR、AI 模块、配置说明、知识库索引/日志和本节 Review。

### Review

- Console 新增一等 `poleagent` 深模块：OpenAI-compatible ModelPort、v1 System Prompt、最多 8 轮 model-tool loop、actor-bound MCP client、工具白名单和稳定错误分类。模型工具面不含 confirm、publish、delete；配置 update 只能生成 Workbench proposal。
- Pole MCP 新增 `get_config_file`、`search_config_files` 只读工具；前端 `/agent` 已删除本地正则 planner，每轮调用 `/ai/agent/v1/turns`，真实工具轨迹、proposal 与 receipt 均按 IndexedDB 会话隔离持久化。
- K8s Go Pod 通过 `poleagent`、`agentworkbench`、handlers、router、aimcp 测试；K8s Node Pod 通过 Agent 契约、目标 ESLint 和 release build；K8s Python Pod 通过 context-kg lint。
- 使用 K8s 临时 OpenAI-compatible 模型服务完成真实浏览器验收：有工具 namespace 查询、无工具回复、带历史多轮、配置提案、diff 预览、用户确认、草稿保存与 `waiting_for_publish` 均通过；确认前配置正文未变化，确认后只改变草稿，测试配置随后恢复原正文。
- 验收中修复两个空集合契约缺陷：无工具回复的 `tools` 与未配置 runtime 的 `tools` 现在稳定返回空数组，前端同时做空值防御。独立新标签页复验未配置状态显示“Agent 未配置”、发送禁用且 0 warning/error。
- 已发布不可变镜像 `pole-control-plane:local-20260723-agent-runtime-v5`，新 Pod Ready、0 restart。临时模型 Pod/Service 已删除，API key 与 baseURL 已恢复为空；当前部署因没有真实 LLM Gateway 配置而安全地保持未就绪。发送门禁使用 runtime `ready`，模型已配置但 MCP 探针失败时同样禁止发送，不能把临时模型桩或部分配置当成生产可用。
- 本轮仍不包含流式输出、服务端会话持久化、配置 create、治理规则写入和 Pole `SystemSecretStore` 编辑页面；这些是后续增强，不影响当前“查询 + 已有配置 update 提案 + 人工确认保存草稿”的最小闭环。

## Pole 内部 Agent 配置与 Secret 闭环（2026-07-23）

目标：LLM Gateway 地址、模型、API key、Prompt 和 MCP 策略由 pole-control-plane 自己完成 Admin 页面编辑、数据库加密持久化、草稿/发布、生效回执与运行时热加载；Kubernetes 不再承担日常业务配置，只保留数据库连接和不可避免的根加密材料等最小自举信息。

- [x] 复核系统配置定义注册表、数据库表、Store、Console 页面和运行时加载链路。
- [x] 决定配置中心、sys config DB 与 Secret Store 的复用方式，冻结领域模型、来源优先级和启动恢复规则。
- [x] 实现 Agent 配置草稿、校验、发布、历史和有效值读取接口，并统一 Admin gate。
- [x] 实现 API key write-only 加密存储、版本引用、替换/保留语义和脱敏响应。
- [x] 让 PoleAgent 运行时从内部发布快照构建并热切换，失败时保持 last-known-good。
- [x] 实现 Admin-only 页面编辑、连接测试、diff、发布和实例生效反馈。
- [x] 迁移现有 YAML/env 为首次初始化值或应急覆盖，不再作为日常编辑入口。
- [x] 仅在 Kubernetes 中完成数据库迁移、接口、单测、浏览器、热更新、回滚和密钥不泄露验收。
- [x] 更新 ADR、配置说明、AI 模块、知识库日志和本节 Review。

### Review

- Console 首期以独立 `SystemSettingsRepository` 复用现有 Pole MySQL，三张专用表分别保存领域指针、不可变配置 revision 与加密 Secret；普通配置管理、Agent 和浏览器均不能直接操作数据库。
- API key 使用随机 DEK + 根 KEK 的 AES-GCM 信封加密，配置 payload 只保存不可变 `pole-secret://` 引用；读取、diff、连接测试响应和运行日志均不回填明文。新 Secret 后续草稿保存冲突时会删除未引用版本。
- 发布会重新探测 OpenAI-compatible LLM Gateway 与 actor-bound Pole MCP，失败返回 502 并保留 last-known-good；发布和后台 reconcile 共用互斥，旧 revision 不得覆盖新快照。当前实例原子切换，多实例通过周期 reconcile 收敛。
- 修复普通控制面代理成功响应刷新 JWT 时清空 role 的问题，并统一兼容旧无 role 主账号 Cookie；真实浏览器在一次普通资源请求和刷新后仍显示 System Configuration，非 Admin 仍由前后端双重门禁拒绝。
- K8s builder Pod 内通过 Go handler/router/store/runtime 测试、System Configuration Admin/error-state 契约、Agent 工作台契约和 Vite production build。K8s E2E 完成草稿、Pole Secret、连接测试、发布前探活、热更新、MCP 工具发现与真实 turn；错误 Gateway 发布被 502 拒绝且 runtime 保持上一 revision。
- 测试 revision、Secret、mock LLM、builder 和数据库客户端均已清理；最终数据库 Agent domain/draft/secret 计数为 0，Pod 重启后 runtime 回到 `static` 且未配置模型时 fail closed。生产仅保留 `pole-control-plane:local-20260723-system-settings-v5`，Ready、0 restart。
- Deployment 不再包含 `POLE_AGENT_LLM_BASE_URL` 或 Agent API key；`pole-runtime-secrets` 只保留 MySQL、Console JWT、auth salt 与 `SYSTEM_SECRET_MASTER_KEY`。管理员现在可在 `/system-configuration?component=pole-console&domain=agent` 配置真实 Gateway 后使用。

## System Configuration Agent 交互重设计（2026-07-23）

目标：把当前“目录表格上叠加编辑和发布按钮”的 Agent 页面重构为任务导向的配置工作台，让管理员清楚理解当前生效配置、连接状态、草稿状态以及测试/保存/发布顺序，同时保持 Pole Secret、发布前探活和 last-known-good 边界不变。

- [x] 审计当前页面、抽屉、暗色主题、窄屏和交互状态。
- [x] 冻结 Agent 专用信息架构、状态摘要、配置分组与主操作层级。
- [x] 重构 Agent 领域主区，常规领域继续使用配置目录表格。
- [x] 重构编辑抽屉和审阅发布交互，明确测试连接、保存草稿和发布的阶段关系。
- [x] 补齐加载、未配置、草稿、已发布、失败和窄屏状态。
- [x] 仅在 Kubernetes builder 与运行 Pod 中完成契约、生产构建和真实浏览器验收。
- [x] 更新本节 Review、lessons 和知识库日志。

Review：

- Agent 领域已从 13 行通用配置表提升为专用工作区：首屏直接展示运行配置、模型、Pole Secret、草稿状态，以及模型连接、MCP、指令和三阶段发布流程；普通领域仍保留来源统计、搜索、筛选和配置目录表格。
- 编辑器改为 720px 模态任务抽屉，API Key 仅支持首次设置、保留和替换；移除会制造不可发布草稿的禁用入口。任何字段或 Secret 改动都会使连接测试结果失效，未通过测试不能保存草稿。
- 发布 Dialog 改为 effective → draft 的字段级差异、Secret 版本引用和三项发布保护审阅；测试连接、保存草稿、发布仍是三个独立动作，服务端发布前复验与 last-known-good 边界未改变。
- Kubernetes builder Pod 内通过目标 ESLint、System Configuration Admin gate、错误态、暗色约束和 production build。运行镜像 `pole-control-plane:local-20260723-system-config-ui-v2` Ready、0 restart，Pod 与 Gateway 主资源 hash 均为 `index.b74291c1.js`。
- 真实 Gateway 以 admin 会话验证 Agent 默认页、编辑抽屉、普通配置目录回归和 1280×720、1180×820、900×800 三档布局，均无页面横向溢出，浏览器 console 0 warning/error；未写入测试草稿或 Secret。

## 全领域 System Configuration 编辑闭环（2026-07-23）

目标：逐项审查 Pole Server 与 Console 的全部系统配置定义，不再把 Agent 当作唯一可编辑领域；为每个字段明确只读/可编辑、校验、生效方式、敏感策略和发布影响，并实现统一的领域草稿、差异审阅、发布与状态反馈。

- [x] 盘点两个组件全部领域、配置项和当前来源。
- [x] 建立字段级编辑性、类型控件、校验、生效和敏感策略矩阵。
- [x] 实现非 Agent 领域的通用草稿、校验、并发控制、发布和历史接口。
- [x] 实现各领域的类型化编辑、变更差异和发布交互。
- [x] 对 BootstrapOnly、RestartRequired、HotReload 和 GuardedHotReload 分别给出准确行为。
- [x] 复用 Admin gate，并覆盖 Secret、冲突、失败和 last-known-good 边界。
- [x] 仅在 Kubernetes 中完成后端、前端、数据库、生产构建和真实浏览器逐领域验收。
- [x] 更新 ADR、知识库索引/日志、lessons 和本节 Review。

### Review

- 63 个定义均增加显式字段策略和服务端校验元数据：Pole Server 29 项中 21 项可管理、8 项保持部署锁定；Console 34 项中 22 项可管理、12 项保持部署锁定。新增字段没有策略时默认锁定，不能因 YAML tag 自动开放。
- 通用领域复用 `SystemSettingsRepository` 保存不可变草稿/发布版本；服务端按 component/domain、字段归属、编辑权限、类型、枚举、数值/时长范围和跨字段关系验证。重启级发布只记录 desired revision 并回执 `pending_restart`，页面继续独立展示 effective value，未伪装成已生效。
- 通用页面保留组件/领域导航，但表格新增发布目标和管理方式；抽屉只渲染该领域已评审字段，发布 Dialog 展示 effective → draft 差异和重启影响。自举、Secret、监听端口、存储连接等字段显示具体锁定原因。
- Agent 专用编辑器补齐原目录中遗漏的提案有效期与资源工具上游超时；二者随发布通过原子 TTL 和请求 context timeout 立即更新，13 项配置不再有“注册但不可操作”的缺口。
- Kubernetes 串行 Go Pod 通过 systemconfig、bootstrap、systemsettings、agentworkbench、handlers、router 测试；Node Pod 通过 Vite release build。镜像 `pole-control-plane:local-20260723-system-config-all-v3` 已滚动，Pod Ready、0 restart。
- Gateway 真实浏览器验证 Server naming 8 项编辑并保存草稿 r4、Server storage 全锁定、Console runtime 仅开放 features、observability 4 项可编辑、Agent 13 项完整编辑；1280×720 下 `scrollWidth = clientWidth = 1280`，控制台 0 warning/error。

## Console 全流程布局与交互实页审计（2026-07-23）

目标：使用 Admin 账号在真实本地 Console 中覆盖全部可见导航和安全可执行操作，识别页面布局、信息层级、表格、抽屉/弹窗、表单、空态、响应式和可访问性问题；旧截图与源码只作为线索，最终结论以本轮实页证据为准。

- [x] 盘点全部可见路由、详情 Tab 和主要操作，建立“页面 × 操作 × 状态”矩阵。
- [x] 登录 Console，逐页覆盖列表、查询、刷新、分页、详情、创建/编辑入口与校验反馈。
- [x] 仅对本轮创建的测试数据执行保存、删除与恢复，不修改现有业务数据。
- [x] 采集并检查关键状态截图，覆盖桌面与窄视口下的布局和滚动行为。
- [x] 将重复问题映射到共享组件、页面样式和统一设计规范，区分结构问题与页面级问题。
- [x] 输出逐步审计结果、优先级、可访问性风险、证据边界和建议修复顺序。

### Review

- 使用 Admin 账号完成全部可见业务域、主要详情 Tab 和安全创建/编辑入口的实页走查，共保留 71 张本轮截图；未提交或修改现有业务资源。
- 确认问题具有系统性：共享 Table 未落实固定首尾列和内部横向滚动，Row/Col 兼容层忽略布局属性，大抽屉未做响应式表单重排，固定高度空表格制造大量无效空间。
- P0 阻断：熔断规则创建页运行时白屏，错误为 `FormItem 必须在 Form 内使用`；源码确认悬浮操作区的 `FormItem` 位于主 `Form` 外。
- 实测窄屏失败：命名空间表格容器约 1046px、主内容约 787px且右侧不可达；服务实例 11 列被压缩进约 745px，内部无横向滚动、无固定列。
- 系统监控、系统配置和 Agent 工作台整体较稳定，可作为后续统一页面节奏与状态表达的正向参考。
- 完整报告与证据保存在 `output/audits/console-full-flow-2026-07-23/`；配置文件详情受列表数据口径阻断，Agent 工作台受本地运行时不完整阻断。

## Console 全站交互与视觉持续整改（2026-07-24）

目标：基于真实 Console 审计结果，一边实页验证一边修复，直到可见业务页面在桌面、超宽和窄视口下不再存在阻断操作、布局破坏、无效交互和明显体验问题。

- [x] 修复熔断规则创建页 `FormItem` 脱离 `Form` 导致的白屏，并建立回归契约。
- [x] 重构共享 Table 的列宽、最小宽度、固定首尾列、内部滚动、省略与全文提示契约。
- [x] 修复共享 Row/Col 的对齐属性和默认占宽，统一资源工具栏响应式行为。
- [x] 优化服务实例、别名、订阅和流量治理的工具栏、长标识、空态与大抽屉布局。
- [x] 修复认证策略向导逐字换行、无动作批量按钮和服务监控伪切换等交互问题。
- [x] 运行前端专项测试、ESLint、生产构建和 Go 相关回归。
- [x] 重建 all-mode，并以 Admin 真实浏览器覆盖全部业务域、主要操作、亮暗主题和多档视口。
- [x] 对新发现的问题继续迭代，直至逐页验收矩阵没有未解决的阻断、布局和明显体验缺陷。

### Review

- 共享 Table 已统一实现自适应列宽、表格内部横向滚动、选择列与首个业务列左侧固定、操作列右侧固定、长内容省略及 `title` 全文提示；非受控分页会进行真实本地切片，受控分页只触发一套回调，窄屏分页可换行。
- 服务实例、命名空间、认证主体/策略、配置文件、MCP Schema、治理规则编辑器等表格型界面均复用或遵循上述契约；640px 实测 `body.scrollWidth === body.clientWidth`，中间内容由局部容器滚动承接。
- 补齐角色查看、编辑与权限详情闭环，并统一鉴权 mutation 对 `{result}`、批量 `{responses}` 和旧式 `{code}` 的响应判断，避免接口失败时前端伪报成功。
- 使用 Admin 会话真实创建、查看、编辑并删除临时角色 `ux-e2e-role-20260724-001`，最终列表恢复为 3 个系统角色；服务监控“带过滤条件查看事件指标”可跳转并预填命名空间与服务名。
- 900px 浏览器遍历命名空间、注册发现、配置、九类治理入口、监控、认证、AI、系统配置和 Agent 共 19 个主路由，均无页面级横向溢出、错误提示或残留弹窗；640px 复验角色表格、MCP Schema、限流和泳道编辑器，亮暗主题下固定列背景均正确。
- 专项契约、全量 ESLint、生产构建（3536 modules）、`go test ./console/...` 和 `git diff --check` 通过。
- Kubernetes 部署更新为 `pole-control-plane:local-20260724-console-ux-v10`，运行 Pod Ready、重启数 0，Console 与 API 探测均为 HTTP 200。

## Console 全站体验完成审计（2026-07-24）

目标：不沿用上一轮“已完成”结论，重新以完整路由、全部可见操作、多视口和真实浏览器状态寻找反例；只有逐项证据闭合后才判断全站体验整改完成。

- [x] 重建全部可见路由、详情页、抽屉、弹窗和创建/编辑操作清单。
- [x] 在 640px、900px、2048px 和亮暗主题下扫描页面级溢出、局部裁切、不可达操作和空白高度。
- [x] 逐项执行安全的查询、重置、刷新、分页、Tab、详情、创建/编辑校验和确认/取消流程。
- [x] 修复共享按钮尺寸、表格操作列、无障碍名称和所有新发现的布局/交互问题。
- [x] 运行全部前端验证脚本、ESLint、生产构建和 Console Go 回归。
- [x] 重建 Kubernetes 运行产物并完成最终浏览器回归。

### Review

- 以 Admin 会话覆盖 19 个主路由，并执行列表查询、重置、刷新、治理/认证页签、服务详情 5 页签、实例与别名新建、配置文件 4 页签、配置发布两步抽屉，以及命名空间/用户/用户组/A2A/MCP 表单重置；未提交、删除或修改现有业务资源。
- 共享 `Button shape="square/circle"` 现在按尺寸约束为 24/32/40px，治理鉴权固定操作列不再裁掉第二个按钮；Header、治理矩阵和 AI 编辑器的图标操作补齐稳定无障碍名称。
- 共享 Table 保持首尾列固定、内部横向滚动和长内容省略；熔断接口/错误/触发条件、探测 Headers、路由目标分组、流量治理接口、标签编辑等伪表格也统一为局部滚动并固定首尾列，不再在窄屏隐藏列头。
- 共享 Drawer/Dialog 修正自定义宽度覆盖、空 footer/action 和响应式边界；960px 配置创建/发布抽屉在 900px 视口实测宽 868px，未产生页面级溢出。CodeDiffEditor 同步主题并使用受视口约束的高度。
- 真实点击发现并修复受控 Input/Textarea 从字符串恢复为 `undefined` 时不更新 DOM 的公共根因；命名空间、用户、用户组、A2A、MCP 表单均以输入后点击重置、读取实际 DOM value 的方式复验通过。
- 资源授权抽屉改为直接有界分页查询并用 `Promise.allSettled` 收敛，单项失败可报告且不永久 Loading；v12 实测授权抽屉 0 alert、0 可见 progressbar。
- 640px、900px、2048px 三档最终遍历 19 个主路由均满足 `body.scrollWidth === body.clientWidth`，无不可达操作、无未提示裁切、无错误 alert；黑暗主题固定列背景使用深色 surface token。
- ESLint、59 个 `verify-*.mjs`、release build（3536 modules）和 `go test ./console/...` 通过。不可变镜像 `pole-control-plane:local-20260724-console-ux-v12` 已滚动，imageID 为 `sha256:fd2ae7a9bf8c927bff306341f4a28c6c71f7af8a2b824d2b910c11d74e6a45d2`。一次性跨 19 个路由、3 个视口的压力式浏览器巡检触发过 1 次 `OOMKilled`（容器内存上限 2 GiB）；容器自动恢复后常态占用约 92 MiB，并持续 Ready，Console 与 readiness 均为 HTTP 200。该重启作为压力巡检运行态事实保留，不再误记为 0 restart。

## Console 全站体验收尾复验（2026-07-24）

目标：针对“所有表格首尾固定、内部滚动、宽度自适应、长内容省略并悬浮展示全文”的统一规范和 Console 治理交互，再次从真实运行页面寻找反例，修复后完成发布级回归。

- [x] 在 1280×720 真实 Admin 会话中遍历 18 个主路由并检查页面级溢出。
- [x] 打开命名空间、A2A、MCP、服务实例、配置分组、用户、策略及五类治理编辑器，检查宽度、滚动、底部操作和校验反馈。
- [x] 对 AI、配置、监控和鉴权表格执行真实横向滚动，测量首尾列滚动前后位置。
- [x] 检查长服务标识的省略样式、实际宽度和 `title` 全文提示。
- [x] 验证主体搜索、刷新、清空、详情，以及 A2A、MCP、服务详情和服务详情五个页签。
- [x] 修复详情面包屑的键盘可达性，并完成静态检查、镜像滚动和最终运行态复验。

### Review

- 共享 Breadcrumb 的可点击项已从鼠标专用 `span onClick` 收敛为 Fluent Button；User Detail 最新资源实测无障碍树出现“用户”按钮，点击后返回 `/auth/principals`。
- 1280×720 下最终遍历 18 个主路由，全部 `documentElement.scrollWidth === clientWidth`，无错误 alert、无残留 progressbar；服务详情五个页签、A2A/MCP 详情、主体查询/刷新/清空与治理编辑器校验均可达。
- A2A 宽表内部真实滚动 547px 后，首列左边界保持 258px、尾列右边界保持 1255px；配置、服务监控、权限策略也完成相同滚动测量。长服务名实际使用 `overflow: hidden`、`text-overflow: ellipsis`、单行省略，并保留完整 `title`。
- ESLint、59 个 `verify-*.mjs`、release build（3536 modules）、`go test ./console/...` 和目标 `git diff --check` 通过。
- 不可变镜像 `pole-control-plane:local-20260724-console-ux-v14` 已滚动，imageID 为 `sha256:85426205e655519cde355dc346350fb6c0186ccef688e21920a24b2a86a33951`；新 Pod Ready、0 restart，Console、API 与 Deployment readiness 均为 HTTP 200，主资源为 `index.ab37b8a5.js`。

## Console 页面不可访问故障修复（2026-07-24）

目标：复现用户反馈的“页面打不开”，从 Gateway、Pod、SPA 入口、静态资源与浏览器运行时逐层定位根因，修复后重新部署并完成真实页面验收。

- [x] 建立根页面、主资源和业务深链的确定性复现检查。
- [x] 检查 Pod、Gateway、入口资源哈希与浏览器错误，确认唯一根因。
- [x] 恢复外部 MySQL 依赖，并补充容器自动重启策略。
- [x] 重建 Pod 并用 HTTP 与真实浏览器复验页面。
- [x] 更新 Review、lessons，提交并推送故障记录。

### Review

- 快速复现检查在修复前稳定得到 `127.0.0.1:8080 connection refused`；Kubernetes 显示 `pole-control-plane` 为 `CrashLoopBackOff`、Service Endpoint 为空。
- Pod 前一实例日志明确报错 `initialize store defaultStore fail: dial tcp ...:3306: connect: connection refused`；宿主机 `pole-mysql` 容器已退出约 8 小时，根因是外部 MySQL 不可达，不是 Console 静态资源或本轮前端提交。
- 已启动既有 `pole-mysql`，确认 `mysqladmin ping` 成功，并把 RestartPolicy 从 `no` 更新为 `unless-stopped`；删除故障 Pod 后 Deployment 自动恢复为 1/1 Ready，新 Pod 0 restart。
- 原始复现检查转绿：根页面、主资源 `index.ab37b8a5.js`、`/namespace` 深链与 8090 API 均为 HTTP 200。
- 真实浏览器验证登录页正常渲染，使用 Admin 会话成功进入 `/namespace`，浏览器无 error 日志。该故障没有代码层回归 seam，最终以依赖容器状态、Pod 日志和端到端 HTTP/浏览器检查锁定。

## Pole 复用 tidemind GreptimeDB（2026-07-24）

目标：停止在 `pole-system` 重复运行 standalone GreptimeDB，通过稳定的本地服务适配层复用 `tidemind` 已有集群，并用独立逻辑库隔离 Pole 遥测数据。

- [x] 核对现有 OTel 数据流、GreptimeDB 查询配置和集群服务边界。
- [x] 将 `pole-greptimedb` 改为跨 namespace 的稳定服务适配层。
- [x] 为 Collector 写入与 Console 查询统一配置 `pole_observability` 逻辑库。
- [x] 更新 Kubernetes 部署说明、观测 ADR、知识库索引与变更日志。
- [x] 在真实 Kubernetes 环境验证建库、OTLP 写入、SQL 查询、Console 查询和资源收敛。
- [x] 保留旧 PVC 作为可恢复历史数据，并提交、推送全部改动。

### Review

- `pole-greptimedb` 已由 ClusterIP + standalone StatefulSet 收敛为 ExternalName Service，目标为 `maas-greptimedb-frontend.tidemind.svc.cluster.local`；Collector 和 Console 继续只依赖 Pole namespace 内的稳定服务名。
- 部署脚本成功幂等创建共享集群中的 `pole_observability`，Collector traces、metrics、logs exporter 均显式携带该数据库 header，Console provider 返回的有效配置也确认 `database=pole_observability`。
- 真实 OTLP log smoke 经 `pole-otel-collector:4318` 写入，并从共享 GreptimeDB 的 `pole_observability.pole_events` 精确查回唯一 marker；`/observability/v1/platform/overview` 返回 provider configured 和真实统计数据。
- `pole-control-plane`、`pole-otel-collector` 与 `tidemind/maas-greptimedb-frontend` 均 Ready、0 restart；Console `/namespace` 返回 HTTP 200，Collector 与 Control Plane 近十分钟日志无 Greptime exporter 错误。
- 旧 `pole-greptimedb` StatefulSet 和 Pod 已删除，5Gi `data-pole-greptimedb-0` PVC 保持 Bound。部署脚本再次执行后建库、服务与 rollout 均成功，Control Plane Pod UID 保持不变，不再产生镜像 tag 往返导致的瞬时 ReplicaSet。
- `bash -n`、ShellCheck（本机可用时）、Kubernetes client dry-run、context-kg lint、目标 Console Go 测试和 `git diff --check` 全部通过。

## 最新知识沉淀整理（2026-07-25）

目标：审计近期托管身份、Console 全站体验、System Configuration、运行恢复与共享 GreptimeDB 交付，把仍停留在任务记录中的稳定知识归档到业务、技术或质量页面。

- [x] 对照最近提交、知识库索引与任务 Review，识别长期页面的覆盖缺口和过时表述。
- [x] 新增 Console UI 质量验收规范，固化表格、表单、响应式与真实浏览器证据边界。
- [x] 更新测试页，区分接口 E2E、静态契约门禁和发布级浏览器验收。
- [x] 更新配置部署页，补充本地 Kubernetes 共享 GreptimeDB 与外部 MySQL 恢复边界。
- [x] 同步相关 ADR、index、log 和双向链接。
- [x] 运行 context-kg lint 与差异检查，记录 Review 后提交推送。

### Review

- 最近的托管身份、WorkloadCredential、TLS 按需、System Configuration 和共享 GreptimeDB 决策已存在于长期 ADR，未重复创建近义页面；本轮重点补齐仍只存在于任务 Review 的 Console UI 质量知识。
- 新增 `console-ui-quality-gates`，固化三层证据模型、共享表格统一契约、表单/弹层约束、多视口发布矩阵、压力巡检诚实记录和页面不可访问快速诊断。
- `testing` 继续限定 `test/e2e` 为 Go/HTTP API，另行链接 Node 源码契约、构建回归与 Kubernetes 真实浏览器验收，消除“接口 E2E 不用浏览器”等于“发布无需 UI 验收”的歧义。
- `configuration` 补充 `pole-mysql`、`pole-greptimedb` 两个 ExternalName seam、MySQL 与 GreptimeDB 同名逻辑库的实例区别、旧 PVC 保留和依赖恢复顺序。
- Fluent ADR、知识库索引和 ingest 日志已同步；index 快速定位增加托管身份、System Configuration、共享 GreptimeDB 和 Console UI 验收入口。
- context-kg lint 验证 39 个 Markdown 页面、全局唯一 basename、frontmatter、双向链接和 index 覆盖全部通过；所有目标页面均只有一个末尾 `## 相关页面`，`git diff --check` 通过。

## Handoff 本地目录规范（2026-07-25）

目标：把交接文档从系统临时目录迁移到项目 `.handoff/`，确保它只作为本地协作资产存在，并将规则提升到全局 AGENTS。

- [x] 将现有 Pole handoff 移入仓库根目录 `.handoff/`。
- [x] 使用 `.git/info/exclude` 本地排除 `.handoff/`，不修改项目 `.gitignore`。
- [x] 在全局 `~/.codex/AGENTS.md` 增加 handoff 路径、脱敏和 Git 验证规范。
- [x] 将本次用户纠正写入项目 lessons。
- [x] 验证源临时文件已移除、handoff 被 Git 忽略且工作区只剩规范记录。

### Review

- 原系统临时目录中的 handoff 已移动到 `.handoff/2026-07-25-pole-control-plane-handoff.md`，原路径不存在。
- `.git/info/exclude` 已本地加入 `.handoff/`；`git check-ignore -v` 返回该规则，`git status --short --untracked-files=all` 不显示 handoff。
- 全局 `~/.codex/AGENTS.md` 已明确本地目录、禁止 Git、脱敏、引用现有产物和双重验证规则，并声明覆盖 handoff skill 的临时目录默认值。
- context-kg lint、双向链接/index 覆盖和 `git diff --check` 通过；仓库待提交范围仅为本次 todo、lessons 和 log。

## 历史 Worktree 合并核对与清理（2026-07-25）

目标：核对两个额外 worktree 是否仍有未提交或未合并成果，仅在确认已被 `develop` 包含后安全移除。

- [x] 盘点额外 worktree 的分支、HEAD 与工作区状态。
- [x] 验证各分支提交是否已被本地及远端 `develop` 包含。
- [x] 移除确认安全的 worktree，并清理失效 worktree 元数据。
- [x] 复核 worktree 列表、主工作区状态并记录 Review。

### Review

- `codex-a2a-agent-registry` 工作区干净，HEAD `6574b2b1` 已同时被 `develop` 与 `origin/develop` 包含，相对主线无独有提交。
- `traffic-governance-rules` 工作区干净，HEAD `29c68320` 已同时被 `develop` 与 `origin/develop` 包含，相对主线无独有提交。
- 两个 worktree 均使用普通 `git worktree remove` 成功移除，无需 `--force`；随后执行 `git worktree prune`。
- 最终 `git worktree list` 只保留主工作区；对应本地分支及远端分支未删除，避免把 worktree 清理扩大为分支清理。

## 服务实例详情布局与编辑入口修复（2026-07-25）

目标：修复服务实例详情中字段标签和值严重错位、信息密度失衡以及编辑入口不可见的问题，使查看态与编辑态形成完整且可验证的操作闭环。

- [x] 对照用户截图审计详情组件、字段布局、权限数据和编辑状态链路。
- [x] 用紧凑的身份摘要、关键状态与分组字段重构服务和实例查看态布局。
- [x] 为有权限的服务与实例提供清晰的编辑入口，并复用现有编辑提交链路。
- [x] 增加布局与操作契约回归，运行专项检查、Lint 和构建。
- [x] 在真实 8080 页面验证亮暗主题、进入编辑与取消操作并记录 Review。

### Review

- 用户截图实际对应 `/discovery/service/instance` 的“服务详情”Tab，而不是单实例 Drawer。根因是 `ServiceForm` 依赖 Fluent Form 默认纵向布局，同时设置整行标签右对齐，导致标签贴到最右、值从最左另起一行。
- 服务查看态已改为独立 definition grid，按“身份与描述 / 归属信息 / 服务标签”组织；宽屏双列、窄屏单列，标签和值保持同一起点。编辑态显式使用左对齐纵向表单，不再依赖默认布局。
- 服务信息区标题栏增加随内容可见的“编辑服务”，并在后续重复入口修复中移除服务身份页首的同义“编辑”；真正的单实例详情 Drawer 固定页脚增加“编辑实例”，点击后复用原有实例编辑器。
- 真实点击发现服务查询的安全投影未返回 capability，导致管理员的 `editable/deleteable` 被 proto 默认值置为 false。服务鉴权拦截器现按逐服务资源分别检查 `UpdateServices` 与 `DeleteServices` 并回填能力，不暴露 SDK 到 control-plane token。
- `verify-discovery-services-layout.mjs` 新增只读网格、显式表单布局、窄屏退化、内容区编辑入口和单实例 view-to-edit 契约；专项检查、目标 ESLint、release build、`go test ./pkg/service/... -count=1` 和 `go test ./... -count=1` 均通过。
- Kubernetes 已滚动更新为 `pole-control-plane:local-20260725-service-instance-detail-v2`。真实 8080 浏览器在 `pole-system/pole.checker` 验证：5 个只读字段标签和值几何对齐，亮暗主题均正常；“编辑服务”可进入编辑态并取消；960px 单实例 Drawer 固定页脚显示“编辑实例”，可进入编辑态并取消。
- 验证截图：`output/playwright/service-detail-layout-light.png`、`output/playwright/service-detail-layout-dark.png`、`output/playwright/instance-detail-edit-entry-dark.png`。

## 服务详情重复编辑入口修复（2026-07-25）

目标：删除服务详情页首与服务信息区重复的编辑操作，仅保留与被编辑内容直接对应的入口。

- [x] 对照用户截图确认两个按钮触发同一 `setEditing(true)` 状态转换。
- [x] 删除服务身份页首的“编辑 / 退出编辑”，保留服务信息区“编辑服务”和表单内取消操作。
- [x] 增加编辑入口数量必须恰好为 `1` 的静态回归检查。
- [x] 执行专项校验、前端构建，并在 Kubernetes 实际页面验证。

### Review

- 根因是前一轮为解决长页面中编辑入口不可见，在服务信息区新增“编辑服务”后，没有同步删除身份页首原有“编辑”，导致同一 `setEditing(true)` 在查看态出现两次。
- 页首现只保留“复制 ID / 查看实例 / 管理别名”；服务信息区保留唯一“编辑服务”。进入编辑态后的退出继续复用表单内“取消”，不再额外增加页首“退出编辑”。
- `verify-discovery-services-layout.mjs` 已清理“页首必须有编辑”的过时契约，并断言 `setEditing(true)` 恰好出现一次、源码不含“退出编辑”。
- 专项校验、目标 ESLint、release build 和 `git diff --check` 均通过。
- Kubernetes 已滚动更新为 `pole-control-plane:local-20260725-service-instance-detail-v3`，运行镜像 ID 为 `sha256:90bb6a30f4e6dffb941a5fcc27b741a03f6d1e8715a5abc5f4ddd6d1a556e4cb`，Pod Ready 且重启次数为 `0`。
- 真实 8080 浏览器验证：查看态“编辑服务”数量为 `1`、精确文案“编辑”数量为 `0`；点击进入编辑态后可通过“取消”恢复查看态，浏览器 page error 为 `0`。截图：`output/playwright/service-detail-single-edit-light.png`。

## 治理规则对消息路由、存储路由与 A/B Test 的覆盖评估（2026-07-26）

- [x] 明确现有治理规则的领域模型、匹配维度、动作模型与执行位置
- [x] 评估消息路由的直接覆盖、可复用能力与数据面缺口
- [x] 评估存储路由的直接覆盖、可复用能力与数据面缺口
- [x] 评估 A/B Test 的流量分组、稳定分桶、指标归因与实验生命周期缺口
- [x] 用代码、知识库与测试证据复核结论

### Review

- 用户进一步明确“消息路由、存储路由”是指中间件运行时治理，而不是 Pole 自身 Store；继续评估标准化和侵入成本后，最终决定不纳入当前核心治理，只保留非侵入式资源集成。
- 现有治理体系可复用的是控制面 CRUD、版本/灰度发布、条件匹配和目标服务实例选择；实际执行仍位于 SDK、xDS 或其它数据面，不能仅凭规则可保存就认定业务能力已覆盖。
- 消息路由仅能覆盖同步 HTTP/RPC 请求按 Header、Query、Cookie、Path、调用方等条件选择服务实例子集；真正 MQ 的 Topic、Queue、Consumer Group、Partition、DLQ 与投递语义没有领域模型或数据面适配。
- 存储路由只能近似复用“将存储节点注册为服务实例后按标签、优先级、权重选址”的骨架；Pole 自身 Store 是启动时全局单选，分库分表、读写分离、主从、事务粘滞、一致性与故障切主均未覆盖。
- A/B Test 只能近似完成预先分群与流量权重路由，尚无稳定用户分桶、实验/变体模型、曝光事件、指标归因、统计显著性、互斥分层和实验生命周期；治理规则的灰度发布表达“向治理客户端下发哪个规则版本”，不等于最终用户实验。
- 数据面存在明显语义差异：xDS 会把目标分组权重转换成 Envoy WeightedCluster，但没有 HashPolicy，不能保证用户粘滞；Rust Proxyless 当前按优先级取第一个有实例的目标组，未消费目标组权重，且 `random_percent` 仅区分 0 与非 0，没有实际百分比采样。Console 当前又默认持续提交 `randomPercent: 0`，会使 Rust SDK 将新建路由判为不命中。
- xDS 的匹配转换也只是协议子集：主要消费 Header、Method、Query 和单独的 Path，未完整转换 Cookie、Caller IP/Metadata/Service、OR 与 `random_percent`，因此 specification 或 Console 可保存不代表 xDS 可执行。
- 静态复核发现 `parseSubRouteRule` 未把 `CustomRoute.caller/callee` 写入内部 wrapper，但缓存依赖 wrapper 的 `Caller/Callee` 建索引；现有专项包测试虽通过，却没有覆盖该映射缺口，因此当前通用服务路由基础也需要先修复并补回归。
- 验证：`go test ./apis/pkg/types/rules ./pkg/cache/rules -count=1` 与 `git diff --check` 通过。

## 多资源域统一治理平台技术方案（2026-07-26，已被 RPC-first 决策取代）

目标：在不立即实现的前提下，设计服务、消息、存储和任务调度四类资源域共享路由、限流、鉴权、镜像、Mock 等治理能力的长期技术方案。

- [x] 收敛“资源域、治理能力、治理策略、数据面能力”等统一语言
- [x] 对比万能规则、按域复制和公共信封加类型化策略三种架构
- [x] 设计资源注册、策略校验、发布分发、能力协商和执行回执接口
- [x] 明确消息、存储、任务领域的差异化模型和安全不变量
- [x] 设计兼容现有九类规则的渐进迁移与分阶段交付
- [x] 定义测试、可观测性、安全和可用性验收标准
- [x] 归档 ADR，并同步 terminology、domain-models、governance-rules、index 和 log

### Review

- 后续评审确认消息、存储和任务缺少与 RPC 等价的标准低侵入执行 seam；本方案不再作为目标架构，未实现的 `adr-multi-resource-governance-platform` 已撤下，由 `adr-rpc-first-governance-scope` 取代。
- 采用“公共规则信封 + 单效果类型化 Spec + Domain/Compiler Adapter”，拒绝万能 JSON、把所有资源伪装成 Service，以及按资源域复制发布栈。
- 一条 Governance Rule 恰好包含一个 effect；多策略协同通过只引用不可变 release 的 Policy Bundle 原子应用，跨 effect 顺序由领域编译管线定义。
- Governance Kernel 以 `ValidateDraft / Publish / ResolveBundle` 三个入口隐藏资源解析、能力覆盖、编译、发布事务和 Bundle 生成复杂度。
- 数据面上报 Capability Profile；控制面按 profile 编译不可变 artifact。不支持、未知或语义不等价时阻止发布，不允许静默降级。
- 数据面使用 `staged / active / last-known-good` 三槽原子应用 Bundle，控制面不可用时继续 LKG；AUTH、存储写和 Job lease 等高风险动作默认 fail-closed。
- Schedule 属于任务定义，负责产生带幂等键的 Execution；Route、RateLimit、Auth、Dry-run、Mock 和故障恢复才属于任务治理。
- 消息镜像不承诺跨 Broker 原子性，存储普通 Mirror 只允许 Shadow Read，Dual Write 是独立高风险能力，Job Mirror 默认 Dry-run。
- V2 canonical spec 保存 deterministic protobuf bytes/type URL/hash；JSON 仅用于展示和兼容导入，避免旧版本 `DiscardUnknown` 读写丢字段。
- 演进按 Phase 0 修复服务治理基线、Phase 1 抽取 V2 内核、Phase 2 Kafka/RocketMQ、Phase 3 Redis/MySQL、Phase 4 Job、Phase 5 Experiment/高级编排推进。
- 该轮曾归档 `adr-multi-resource-governance-platform` 并同步知识库，后续 RPC-first 范围评审已将其撤下；本条仅保留历史过程。
- 已验证所有 context-kg 页面 frontmatter 和 index 覆盖、新增 ADR 双向链接与 related pages 一致、log 的 related pages 位于末尾，`git diff --check` 通过。

## Control Plane 自身 MCP 注册与 Console Agent 消费链路核查（2026-07-26）

- [x] 回顾 lessons、知识库索引并确认仓库未启用 CodeGraph。
- [x] 核实 Control Plane 自身 MCP 协议端点及工具清单。
- [x] 核实 MCP Registry 的注册模型是否支持把 Control Plane 自身作为 MCP Server 注册。
- [x] 核实 Console Agent 是否从 MCP Registry 读取并实际调用已选 MCP 工具。
- [x] 汇总当前闭环、配置前提和未自动化环节，补充 Review。

### Review

- Control Plane 已在 `/ai/mcp/v1/sse` 暴露真实 MCP Server，并提供 Namespace、MCP Registry、配置文件查询等工具；Console Agent 当前也确实消费这个 MCP。
- 当前消费链路是 System Settings/YAML 中的固定 `agent.mcp.endpoint` → SSE client → `tools/list` → allowlist 过滤 → LLM tool call → 同一 client `tools/call`，默认地址直接指向本机 Control Plane MCP。
- MCP Registry 是另一层服务目录。它允许手工登记 Control Plane 自身，但启动期没有幂等自注册，登记后也不会自动执行远端 `tools/list` 并同步 `mcp_server_tool`。
- Agent 默认可通过 `list_mcp_servers`、`list_mcp_server_tools` 读取 Registry 元数据，但没有根据 Registry 地址建立第二个 MCP client 或动态路由下游工具；因此“能看见 Registry 数据”不等于“消费 Registry 中登记的 MCP”。
- 目标闭环尚需补齐自注册、工具探测同步、动态连接/路由，以及鉴权、SSRF、协议兼容和健康治理；当前 Registry 记录不影响 Agent 是否能连接 Pole MCP。
- 验证：`go test ./plugin/apiserver/httpserver/aimcp ./console/pkg/poleagent ./console/pkg/router -count=1` 全部通过。

## Pole 自注册、Prompt 管理与 Agent 自管理闭环（2026-07-26）

目标：让 Control Plane 自动注册自身 MCP、Pole Agent 自动注册自身 A2A，并将 Registry、Prompt 版本与运行时能力纳入统一、可审计的自管理闭环。

- [x] 通过递进式讨论明确自管理目标、自治边界与安全不变量。
- [x] 核实现有 MCP Registry、A2A Registry、Pole Agent Runtime 与 System Settings/Prompt 管理 seam。
- [x] 设计统一的自身能力声明、协调、探测、版本发布和回滚模块。
- [x] 将长期架构决策归档到 `context-kg/technical/adr/`，同步 index、log 与相关页面。
- [x] 按确认范围实现 MCP/A2A 自动注册、工具同步、Prompt 管理与 Agent 消费闭环。
- [x] 完成单元、集成、权限、安全、故障恢复和真实运行环境验证。

### Review

- 管理员是唯一 desired state 配置主体；确定性执行身份固定为 `pole-self-manager`。它没有可供模型使用的管理员 Token，只负责自身 MCP/A2A 投影和已获管理员授权的 Agent 配置自动应用。
- Control Plane 启动时通过进程内 `tools/list` 获取真实 MCP 工具定义，将 `pole-system/pole-control-plane` 与工具 Schema 幂等写入 Registry；每 30 秒全量 reconcile 可恢复漂移和软删除。
- Pole Agent 已发布 `/.well-known/agent-card.json` 和受认证的 `/ai/agent/a2a/v1` JSON-RPC `message/send`，`all` 模式从真实 Card 自动生成 A2A Registry 投影，避免两份能力清单漂移。
- Console Agent 使用当前用户身份查询 Registry，自然键必须精确解析唯一 address backend 且 URL 仅允许 HTTP(S)，再建立 MCP 会话并应用工具白名单。
- Agent Prompt、模型与工具策略保存后自动构建候选、探测模型/MCP，并由 `pole-self-manager` 发布；失败草稿保留为 rejected，当前运行时继续使用上一健康版本。内建安全 Prompt 和 Secret 不向模型开放。
- MySQL MCP tool、A2A interface/skill 使用稳定子项 ID 与 upsert；聚合根 Update 显式恢复 `flag=0`，保证手工软删除后的下一轮可复活。
- 验证：全量 `go test ./... -count=1`、`go test -race ./pkg/selfmanager`、`go vet ./pkg/selfmanager`、前端目标 ESLint、管理员门禁契约、`npm run build:test`、context-kg lint 与 `git diff --check` 全部通过。

## 治理多灰度与 Discover 标签快照（2026-07-26）

- [x] 核对配置多灰度、治理发布存储、各治理缓存和 gRPC Discover 的现状。
- [x] 以 RED 测试固化同一规则多个 active gray 发布并存。
- [x] 以 RED 测试固化按 caller labels 选择 normal/gray、最新命中优先和稳定快照 revision。
- [x] 以 RED 测试固化 gRPC LANE 分发、DiscoverFilter 透传与配置顶层 revision。
- [x] 实现最小完整链路并完成针对性 Go 测试与差异审查。

### Review

- 治理发布存储不再在发布/激活 gray 时停用同规则其它 gray；normal 只替换 normal，因此 normal 与多个 gray 可并行。gray 的缓存键和灰度资源键均包含 release name，避免发布与匹配条件互相覆盖。
- 新增统一发布选择器：按 rule_id 独立选择；匹配多个 gray 时按 version、mtime 降序取最新，否则回退最新 normal；revision 哈希选中发布的 rule/release 身份和 version，顺序稳定且不同标签快照可区分。
- Router、RateLimit、CircuitBreaker、FaultDetect、Lane、Lossless、TrafficSecurity、TrafficMirror、TrafficMock 九类治理缓存均从同一选择器生成客户端快照；MySQL 转换恢复持久化的 ClientLabels，gRPC 把 DiscoverFilter 放入请求上下文供治理服务消费。
- `RuleRelease.FromSpec/ToSpec` 已补齐 ClientLabels 往返，避免发布入口在保存灰度资源和发布记录前丢失匹配条件；control-plane specification 正式依赖升级到 `v0.1.0-ALPHA.38`。
- 交叉审查修复两项 P1：发布缓存键改为对 `namespace + rule_id + rule_name + release_type + gray release_name` 做长度前缀编码，跨环境同名规则不会覆盖；熔断与故障探测只有在当前 service/namespace 层实际选中发布时才停止，否则继续回退到 namespace/global。
- 新增 inactive gray 后回退 normal 且 revision 改变的核心选择测试，以及跨 namespace key、熔断 namespace 回退、故障探测 global 回退测试。
- gRPC Discover 新增 LANE 分发。Config Discover 使用顶层 revision，不再把任意非空 file.id 误判为未变；同时透传 filter caller labels，并以发布 ID/name/type/version/md5 生成快照 revision，停止高版本 gray 后可正确回退较旧 normal。
- RED → GREEN 验证：专项 `go test ./apis/pkg/types/rules ./pkg/cache/rules ./pkg/goverrule ./pkg/config ./plugin/store/mysql ./plugin/apiserver/grpcserver/discover/v1 -count=1` 通过；全量 `go test ./... -count=1` 通过；`git diff --check` 通过。

## 服务契约能力完成度核查（2026-07-26）

- [x] 回顾 lessons、工作树状态并确认仓库未启用 CodeGraph。
- [x] 核对服务契约 specification、业务服务、存储、缓存与对外 API 链路。
- [x] 核对 Console 入口、页面、请求适配与交互闭环。
- [x] 核对单元测试、集成测试、客户端消费与知识库承诺。
- [x] 运行针对性验证，区分已实现、未接通与缺失能力。
- [x] 汇总结论并补充 Review。

### Review

- 服务契约不是从零未做：已有 specification 模型、MySQL 主表与接口明细表、业务 CRUD、接口全量替换/追加/删除、HTTP 管理接口、gRPC 客户端上报/发现、鉴权/参数校验和本地 Pebble value cache。
- 但 Console 没有菜单、路由、服务详情 Tab、列表/详情/版本/编辑页面；`services/service.ts` 仅残留三个零调用请求，并且 `/services/contract`、`/services/contract/versions`、`/services/contract/interfaces/delete` 均与后端 `/service/contracts`、`/service/contract/versions`、`/service/contract/methods/delete` 不一致。`services/contract.ts` 是办公采购合同 mock 模板，与服务契约无关。
- 后端存在未收口断点：specification 已声明 `SERVICE_CONTRACTS`，但 HTTP/gRPC 统一 Discover 分发均没有对应 case；代码中的 gRPC `GetServiceContract` 未进入 specification 的 service descriptor，实际不会注册。`GetServiceInterfaces` 也未进入业务接口或 HTTP 路由。
- 契约列表/版本查询只做函数权限校验，没有复用服务列表的逐资源权限过滤；契约明细关联查询未过滤 `service_contract_detail.flag`，软删除接口仍可能进入管理响应和客户端缓存。
- 缓存 miss 返回带空 `ServiceContract` 的非 nil 对象，而客户端上报与发现使用 nil 判断是否不存在；空 content 首次上报可能跳过创建，发现 miss 也可能返回成功的空契约。
- 新旧字段兼容没有闭环：proto 已弃用 `name` 并推荐 `type`，但参数拦截器和客户端缓存查询仍依赖 `name`；缓存响应还会遗漏 metadata/content digest。全量替换接口会先删除所有来源明细，SDK 上报可能清掉原有手工接口，与 Manual 覆盖 Client 的领域合并逻辑冲突。
- 专项测试几乎只有 Pebble 默认路径与 gRPC 方法名清单；没有业务 CRUD、接口增删、HTTP handler、MySQL Store、鉴权过滤、缓存命中/未命中/删除、一致性、客户端 Report/Get、集成/E2E 和性能基准。
- 已运行 `go test ./pkg/service ./pkg/service/interceptor/auth ./pkg/service/interceptor/paramcheck ./pkg/cache/service ./plugin/store/mysql ./plugin/apiserver/httpserver/discover ./plugin/apiserver/grpcserver/discover/v1 -count=1`，现有测试均通过；这只能证明已有包与用例未失败，不能覆盖上述断点。`git diff --check` 通过。
- 完成度结论：后端属于“骨架与主干已存在但可靠性未收口”，Console 属于“基本未产品化”，整体服务契约能力不能视为完成。

## 四协议服务契约完整闭环（2026-07-26）

目标：支持 HTTP/OpenAPI、Dubbo、gRPC、Thrift 四类服务契约统一上报、存储、发现和 Console 可视化，并修复现有契约链路的正确性与权限缺口。

- [x] 比较上报与协议适配模块的候选接口，确定深模块 seam 和兼容策略。
- [x] 以 RED 测试固化缓存 miss、首次上报主从延迟、软删除、Manual/Client 合并和字段兼容语义。
- [x] 实现 gRPC SDK 主上报、HTTP 单对象上报与 OpenAPI 接口提取。
- [x] 修通客户端契约查询、管理端列表/版本、权限过滤和真实 HTTP 路由。
- [x] 在服务详情增加服务契约 Tab，完成四协议列表、版本、接口与原始内容可视化。
- [x] 补齐前后端专项测试、集成验证和构建验证。
- [x] 将上报契约、协议映射和兼容决策归档到 `context-kg`，同步 index/log。
- [x] 完成双轴代码审查、修复遗留问题并显式提交本次改动。

### 已确认测试 seam

- 上报 seam：gRPC `ReportServiceContract` 与 HTTP 单对象上报入口。
- 查询 seam：管理端契约列表/版本和客户端按五元组发现。
- 展示 seam：服务详情“服务契约”Tab 及其统一请求适配器。

### Review

- 上报采用 caller push：SDK 走 gRPC `ReportServiceContract`，Agent/CI 走 HTTP `POST /naming/v1/ReportServiceContract`；控制面不主动扫描生产服务。
- HTTP 始终校验并解析 OpenAPI 3.x；gRPC、Dubbo、Thrift 要求原始契约和结构化接口。接口签名进入 ID 与覆盖键，Dubbo 重载不会丢失且重复上报幂等。
- 契约和接口 ID 均由自然键确定并校验，管理端查询逐契约执行资源权限过滤；Client/Manual 分源替换，兼容历史 `source=0` SDK 数据。
- 服务详情新增“服务契约”页签，支持四协议能力概览、全量分页、版本/契约切换、接口来源和原始内容展示。
- 通过服务契约目标 Go 测试、MySQL sqlmock、HTTP/gRPC 路由测试、Console 专项脚本、目标 ESLint、`npm run build:test`、context-kg lint 和 `git diff --check`。
- 双轴复审最初发现 ID 越权、主从延迟、旧 CRUD 兼容、权限泄露、OpenAPI 绕过、Dubbo 重载和幂等问题；全部修复后 Standards 与 Spec 复审均确认无阻塞/高风险项。
- `go test ./...` 与 `make build` 被任务范围外的未提交 xDS 改动阻断：`plugin/apiserver/xdsserverv3/cache/node_resources_test.go` 使用错误 BoolValue 类型，`plugin/apiserver/xdsserverv3/generate.go` 缺少 `fmt` import；本任务目标包和 Console 构建均通过。
- 已提交服务契约闭环 `e6991f4b`；部署时发现后续自管理链路在 MCP Registry 空结果上直接 `Scan` 的启动回归，补充回归测试并以 `6ac2cd34` 修复。
- 已从干净提交 `6ac2cd34` 构建镜像 `pole-control-plane:local-20260726-service-contract-6ac2cd34` 并滚动部署到 `pole-system/pole-control-plane`；Pod `1/1 Ready`、零重启，Console 与健康检查均返回 200。
- 实际演示数据位于 `demo-governance/demo-order`：HTTP/OpenAPI、gRPC、Dubbo、Thrift 各 1 份契约和 3 个接口，共 4 份契约、12 个接口。

## RPC-first 治理范围收敛（2026-07-26）

- [x] 复核 RPC、消息、存储和任务的数据面标准化与侵入成本
- [x] 决定核心治理只面向 HTTP、gRPC、Dubbo 等同步服务调用
- [x] 明确 Kafka、RocketMQ、Redis、MySQL 只保留非侵入式资源集成
- [x] 明确任务调度需等待统一 Worker/Agent 后独立评估
- [x] 撤下多资源域运行时治理提案并归档 RPC-first 范围 ADR
- [x] 同步术语、领域模型、治理功能档案、index、log 和 lessons
- [x] 完成知识库结构、双向链接和差异校验

### Review

- RPC 拥有请求上下文和 Client/Server Interceptor 形成的真实低侵入 seam，适合继续深化路由、限流、鉴权、镜像、Mock、熔断和 A/B Test。
- Kafka/RocketMQ 的 Producer、Consumer、Partition/Queue 和事务语义，以及 Redis/MySQL 的连接、Session、事务、Shard 和一致性语义无法用统一 RPC 模型无损表达；当前不建设运行时路由、灰度或镜像。
- 外部中间件只允许资源目录、健康、指标、管理链接和经单独评审的原生 ACL/Quota/配置 Adapter，不称为已纳管的运行时治理。
- 不向 specification 增加 `ResourceDomain`、MESSAGE/STORAGE/JOB Rule 或万能 GovernancePolicy；服务数据面的 Capability Profile、原子 Bundle 和 Apply Receipt 仍属于 RPC 治理深化范围。
- 原多资源域 ADR 已撤下，当前范围决策归档为 `adr-rpc-first-governance-scope`。
- 已通过全部 context-kg frontmatter、全局唯一 basename、index 覆盖、wiki 链接解析、相关页面一致性和 `git diff --check` 校验。

## Dependabot 安全告警清零（2026-07-27）

目标：清理 `develop` 分支全部可达依赖漏洞，不保留 critical/high 告警，也不通过 dismiss 掩盖没有修复版本的直接依赖。

- [x] 拉取 GitHub Dependabot 明细，按生态、依赖、严重度和修复版本归并。
- [x] 运行本地 npm 审计并定位 Go 依赖路径。
- [x] 升级 Go 安全相关模块并解决兼容问题。
- [x] 升级 Console 运行时与构建工具链，移除无修复版本的 mock 依赖。
- [x] 运行 `govulncheck`、`npm audit`、全量 Go 测试、Console lint/build 和专项脚本。
- [x] 提交并推送 `develop`，复核远端 Dependabot 告警状态。

### Review

- GitHub 开放 Dependabot 告警共 98 条，集中在 Go 的 Docker SDK、Gin、JWT、`x/crypto`、gRPC，以及 Console 的 Axios、Vite、React Router、Lodash、MockJS 等依赖。
- Go 已升级 Gin、JWT、gRPC、`x/crypto`、`x/net`、`x/text`、QUIC 与 Testcontainers；Testcontainers 改用拆分后的 Moby API/Client 模块，移除存在告警的旧 Docker 单体模块。`govulncheck` 结果为 0 个受影响漏洞、0 个导入包漏洞。
- Console 已升级 Axios、Vite、ECharts、Lodash、UUID 与 SVG 工具链；删除无修复版本的 MockJS/Vite Mock，使用轻量路由兼容层替代存在相互冲突安全公告的 React Router，使用 Oxc 替代引入旧 vulnerable glob 依赖链的 ESLint 工具链。
- `npm audit --audit-level=low` 为 0；全部已登记 Console 专项测试、服务契约专项、lint、测试/发布构建、真实浏览器路由冒烟、Go 全量测试、E2E 标签编译、完整打包、知识库检查和 `git diff --check` 均通过。
- 新增 Go 与 Console npm 的 Dependabot 周更配置，目标分支固定为 `develop`；提交推送后 GitHub 重算期间开放告警由 98 降至 46，最终为 0，未 dismiss 任何未修复告警。

## Dependabot 修复版本更新至 K8s（2026-07-27）

目标：将 `develop@69be9be7` 对应的安全修复版本更新到当前 Pole Kubernetes 环境，并验证运行实例、接口与 Console 页面均使用新构建。

- [x] 核对当前 Kubernetes context、namespace、工作负载、镜像与部署脚本。
- [x] 基于当前 `develop` 构建本地镜像并更新目标工作负载。
- [x] 等待 rollout 完成，检查 Pod、镜像摘要、重启与容器日志。
- [x] 验证 Console、Control Plane HTTP 接口和服务契约页面入口。
- [ ] 记录部署 Review，完成知识库校验、提交并推送 `develop`。

### Review

- OrbStack `pole-system/deployment/pole-control-plane` 已从 `local-20260726-service-contract-6ac2cd34` 更新为 `pole-control-plane:local-20260727-dependabot-69be9be7`，镜像 ID 为 `sha256:ee73f1729910a156453f38fd7bdbe37871e1f548fef32785eba08820e0c621f2`。
- 新 Pod `pole-control-plane-d59ccb559-bb9mp` Ready；首次进程因 self-manager 启动瞬时失败以退出码 0 重启一次，随后持续稳定超过两分钟，重启计数未增加，当前日志无 ERROR/panic/fatal，周期 self-management Agent Card 探测均为 200。
- Pod 内 Console、Control Plane 8090 根接口与 functions 健康接口均返回 200；Gateway 根页返回 200，HTTPRoute `Accepted=True`、`ResolvedRefs=True`。
- Pod 内与 `pole.localhost` 返回的 `index.html` SHA-256 均为 `fef5b55c16da87ce78e2d4cd19752a2db102f1c7acd050f7e34b2d2aae4f3f09`，证明实际 Gateway 已提供本次容器产物；Dependabot 开放告警复核仍为 0。
- 真实浏览器登录 K8s Gateway 后打开 `demo-governance/demo-order` 的“服务契约”Tab，页面显示 4 份真实契约，HTTP/OpenAPI、Dubbo、gRPC、Thrift 各 1 份，Dubbo 接口及重载方法可见，浏览器错误为 0。
- Deployment 已补齐 `change-cause=deploy 69be9be7: Dependabot security fixes`、source revision、image ID 与部署时间注解，运行版本可追溯。
