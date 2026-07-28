---
title: 操作日志
tags: [meta, log]
links: [index, schema]
updated: 2026-07-28
sources: 0
---

# 操作日志

追加式记录，只增不改。最新记录在文件末尾。格式规范见 [[schema]]。

---

## [2026-05-14] ingest | pole-control-plane 全量初始化

- 扫描 656 个 Go 文件（模块：`github.com/pole-io/pole-server`）
- 生成 12 个知识页面，覆盖以下领域：
  - 概览（overview）：项目介绍、技术选型、目录结构、启动流程
  - 架构（architecture）：四层架构、插件系统、Store/Cache 接口、拦截器链
  - 业务领域（domain-components）：命名空间、服务发现、配置中心、治理规则、管理后台
  - API 服务端（api-servers）：HTTP、gRPC、xDS v3、Nacos v1/v2、Apollo、Eureka、MCP 集成
  - 存储层（storage）：接口层次、MySQL 实现、软删除、增量查询、分布式锁
  - 缓存层（cache-layer）：缓存子类型、1 秒刷新循环、-5 秒安全窗口
  - 认证系统（auth-system）：UserServer、StrategyServer、拦截器链、Token 流程
  - AI 功能（ai-features）：MCP Registry、领域模型、HTTP API、缓存
  - 公共基础设施（common-infra）：日志、EventHub、Batch Controller、OTel、同步原语
  - 配置参考（configuration）：YAML 结构、插件配置、API 服务端配置、部署目录
  - 模式与约定（patterns）：11 种关键代码模式
  - 测试（testing）：Mock Store/Auth、集成套件、测试规范
- 首次建立 wiki 结构，新建元文件：schema.md、index.md、log.md
- 替换旧 README.md 为 index.md

## [2026-05-14] restructure | wiki 目录结构重组为方案 C

- 新建子目录：_meta/, overview/, domains/, infra/, ai/, guides/
- 拆分 domain-components.md → 5 个独立域文件：
  - domains/namespace.md
  - domains/service-discovery.md
  - domains/config-center.md
  - domains/governance-rules.md
  - domains/admin.md
- 移动 12 个内容页面至对应子目录：
  - overview/: overview.md, architecture.md, api-servers.md, configuration.md
  - infra/: storage.md, cache-layer.md, auth-system.md, common-infra.md
  - ai/: ai-features.md
  - guides/: patterns.md, testing.md
- 更新 _meta/index.md 为分类视图（五大类别：overview/domains/infra/ai/guides）
- 更新 _meta/schema.md 目录结构约定为新分层结构
- 删除根目录下所有旧文件（domain-components.md 及已移动的 12 个页面）

## [2026-06-08] ingest | 治理规则统一存储与缓存更新方案归档

- 扫描 0 个 Go 文件，本次为技术方案归档。
- 更新页面：governance-rules, index
- 变更摘要：
  - 将治理规则统一表 `governance_rule` / `governance_rule_release` 的设计归入治理规则域。
  - 记录 `GovernanceRuleRepository` 统一 DB 操作原则。
  - 记录 `GovernanceRuleUpdateCache` 按 `rule_type` fan-out 给 watcher 的缓存更新模型。
  - 记录泳道方案 A：`LaneGroup` 是聚合根，`LaneRule` 是 group JSON 内最多 20 个的子对象。
  - 清理重组前残留的 `domain-components` 失效链接，改为当前知识库入口 `index`。

## [2026-06-09] restructure | context-kg 重组为 business / technical / quality 三域结构

- 按业务知识域、技术知识域、质量保障知识域重组目录。
- 新增业务基础页面：
  - business/terminology.md
  - business/domain-models.md
  - business/business-rules.md
- 移动功能档案到 business/feature-registry/。
- 移动架构、模块、接口、环境、约定页面到 technical/。
- 移动测试知识到 quality/automation/。
- 将治理规则统一存储与缓存更新方案拆为 `technical/adr/adr-governance-rule-unified-storage.md`，业务功能页只保留摘要和链接。
- 更新 _meta/schema.md 为三域结构规则，并明确长期技术方案归档到 `technical/adr/`。
- 更新 _meta/index.md 为三域目录入口。

## [2026-06-09] restructure | 任务记录迁移与 AGENTS 主文件反转

- 将任务记录从 `docs/tasks/` 迁移到 `context-kg/tasks/`。
- 删除旧 `docs/design/` 长期设计文档，长期知识统一归档到 `context-kg/`。
- 将仓库内指导文件反转为 `AGENTS.md` 保存实际内容，`CLAUDE.md` 使用相对软链接指向 `AGENTS.md`。
- 更新 schema 与 index，将 `todo`、`lessons` 纳入任务过程记录入口。

## [2026-06-10] ingest | A2A Agent Registry 功能设计归档

- 扫描 8 个相关资料与源码入口，本次为技术方案归档。
- 新增页面：technical/adr/adr-a2a-agent-registry.md。
- 更新页面：ai-features, index。
- 变更摘要：
  - 明确 A2A 与 MCP 的功能边界：MCP 管工具/资源，A2A 管 Agent 间任务协作。
  - 决定首期做 A2A Agent Card 注册、发现、缓存和 MCP 工具暴露，不接管完整 A2A task runtime。
  - 记录 A2AAgent、A2AAgentInterface、A2AAgentSkill 的建议数据模型。
  - 记录 Store、Cache、HTTP 插件、Console、安全和测试落点。

## [2026-06-10] refine | A2A Agent Registry 控制面范围收敛

- 按用户纠正收敛 A2A 方案边界。
- 更新页面：adr-a2a-agent-registry, ai-features, todo, lessons。
- 变更摘要：
  - 明确 pole-control-plane 只负责 A2A Agent Registry。
  - 将 A2A task proxy、SSE streaming 转发、push notification broker、task 状态机和 artifact 存储归类为数据面/网关/agent runtime 能力。
  - 将该边界沉淀到 lessons，避免后续把数据面能力写成 control-plane 二期规划。

## [2026-06-10] implement | A2A Agent Registry 实现集成

- 将 A2A Agent Registry 从独立 worktree 集成到当前 develop 工作区。
- 更新页面：ai-features, adr-a2a-agent-registry, todo, lessons。
- 变更摘要：
  - 新增 `A2AAgent`、`A2AAgentInterface`、`A2AAgentSkill` 类型、Store 接口、MySQL 实现和 SQL schema。
  - 新增 `A2AAgentCache`，接入 CacheManager 并支持按 namespace/name、skill、协议绑定和后端信息查询。
  - 新增 `/ai/a2a/v1` REST API 子包，提供 Agent 注册、更新、删除、查询、skill 查询和 Agent Card 查询。
  - 新增 Console A2A Agents 页面，包含列表筛选、抽屉新建/编辑、Agent Card 查看和技能详情查看。
  - 保持控制面边界，不提供 A2A task proxy、SSE 转发或 push broker。

## [2026-06-15] ingest | Console API、Client 与权限接口 E2E 测试用例设计

- 扫描 12 个相关测试、Console router、HTTP access、接口 service 和知识库入口。
- 新增页面：quality/testcases/console-client-auth-e2e-testcases.md。
- 更新页面：testing, index, schema, todo, lessons。
- 变更摘要：
  - 将 E2E 测试收敛为 Console API 和 Client/Auth 两层接口套件，不纳入前端交互测试能力。
  - 设计命名空间、MCP、A2A、服务、别名、实例、9 类治理规则、用户、用户组、角色、权限策略的读写覆盖矩阵。
  - 设计通过 Console 写入后由 8090 Client API 轮询验证服务发现和治理规则发布态的用例。
  - 设计 `consoleOpen` / `clientOpen` 四象限配置开关和 Console/Client 权限策略生效用例。

## [2026-06-15] lint | context-kg skill 与标题契约校验

- 新增本地 Codex skill：`context-kg-maintainer`，用于快速更新、生成和校验 context-kg 知识库。
- 使用 skill 自带 lint 脚本检查当前知识库。
- 修正若干页面 frontmatter `title` 与首个 H1 不一致的问题。
- 更新页面：index, schema, namespace, service-discovery, config-center, governance-rules, admin, todo。

## [2026-06-15] refine | 接口 E2E 测试用例边界收敛

- 按用户要求移除前端页面和 Playwright 测试能力口径。
- 更新页面：console-client-auth-e2e-testcases, testing, todo, lessons。
- 变更摘要：
  - 在测试用例文档中明确 E2E 只验证 HTTP/API 链路。
  - 增加非目标清单，排除 Playwright、浏览器、DOM、截图、页面布局和前端路由。
  - 将少量 UI 口径描述改为接口字段、接口响应和接口可用性断言。
  - 保留 8080 Console API、8090 Client API、权限策略和 `consoleOpen/clientOpen` 开关矩阵作为接口验收边界。

## [2026-06-16] implement | 接口 E2E 测试套件补齐

- 新增测试目录：test/e2e/internal/e2e, test/e2e/console_api, test/e2e/client, test/e2e/auth。
- 更新页面：testing, todo。
- 变更摘要：
  - 新增基于 testcontainers 的 MySQL 环境、30000 起始端口分配和临时 all 模式配置。
  - 新增 Console API 资源读写、Client Discover 传播和权限策略接口 E2E 测试代码。
  - 使用 `//go:build e2e` 隔离 E2E，默认测试命令不启动 Docker。
  - 记录 macOS 下 testcontainers 间接依赖 `go-m1cpu` 的 cgo 崩溃规避方式，推荐 `CGO_ENABLED=0`。

## [2026-06-16] refine | 接口 E2E 测试文档边界强化

- 按用户要求再次收敛测试用例文档，只保留接口维度测试能力。
- 更新页面：console-client-auth-e2e-testcases, testing, todo。
- 变更摘要：
  - 新增“能力边界”说明，明确自动化入口统一为 Go test。
  - 明确 8080 console proxy 属于接口链路验证，不等同于控制台页面测试。
  - 明确不建设 `console/web/e2e`、Playwright、浏览器脚本、DOM 断言、截图比对或页面交互流程。

## [2026-06-16] implement | 接口 E2E 测试覆盖补齐

- 补齐设计文档中仍缺失的接口 E2E 覆盖点。
- 更新页面：todo。
- 变更摘要：
  - Console API E2E 增加资源删除后不可见、服务 update/delete、auth 资源 delete 和治理规则 gray/stopbeta/rollback/release delete 覆盖。
  - Client E2E 增加实例删除和治理规则删除后的 Discover 响应不包含目标资源断言。
  - Auth E2E 增加 RegisterInstance、Heartbeat 在 `clientOpen` 开关和授权策略下的放行/拒绝用例。
  - 继续保持接口维度，不引入前端页面或 Playwright 测试。

## [2026-06-16] refine | 接口 E2E 测试用例文档二次整理

- 按用户要求继续收敛测试用例文档，只保留接口维度测试能力。
- 更新页面：console-client-auth-e2e-testcases, testing, todo。
- 变更摘要：
  - 将主测试用例文档中的前端自动化表述收敛为非目标边界。
  - 保留 8080 console proxy 作为接口链路验证，不纳入控制台页面验收。
  - 测试能力入口继续统一为 Go test。

## [2026-06-16] implement | 接口 E2E 鉴权与治理覆盖补齐

- 补齐权限策略、Client 鉴权传播和治理规则类型隔离的接口 E2E 表达。
- 更新页面：console-client-auth-e2e-testcases, todo。
- 变更摘要：
  - Auth E2E 使用真实 policy principals/resources/functions 结构，并覆盖用户组继承、角色函数变更、Token 禁用/刷新、策略删除和策略更新传播。
  - 治理规则权限增加老规则路由与新增流量治理规则的授权/未授权分支。
  - 测试用例文档将灰度、路由、泳道的客户端 label 断言收敛到当前 HTTP/API 可验证的发布记录、规则内容和 Discover 发布态。
  - 继续保持接口维度，不引入前端页面、浏览器或 Playwright 测试能力。

## [2026-06-21] refine | 治理规则编辑抽屉双栏滚动交互约定

- 将治理规则编辑抽屉的左侧内部滚动、右侧 Spec 高度和 StickyTool 布局要求沉淀为长期技术约定。
- 更新页面：patterns, index, todo, lessons。
- 变更摘要：
  - 明确编辑器正文固定高度、shell hidden、左侧 pane 内部滚动、右侧 Spec 跟随 shell 高度。
  - 明确 column flex 下直接子 section 必须禁止 shrink，避免共享 section 的 overflow hidden 裁切内容。
  - 明确真实验收必须验证 scrollHeight/clientHeight 和 scrollTop，而不是只看 CSS overflow。

## [2026-06-22] ingest | Console OIDC 用户来源与企业目录同步方案归档

- 新增页面：technical/adr/adr-console-oidc-identity-source.md。
- 更新页面：auth-system, index, todo。
- 变更摘要：
  - 明确 OIDC 登录和企业目录同步都是 Console 扩展点，不进入 pole-server 核心鉴权链。
  - 明确 native / oidc 登录模式启动期互斥，运行期间不支持切换。
  - 明确 OIDC 只改变 Pole User 来源，业务请求仍使用 Pole token 和现有资源授权逻辑。
  - 设计飞书、Lark、钉钉 DirectoryProvider 同步企业账号为 `source=oidc` 的 Pole User，解决未登录用户无法提前授权的问题。

## [2026-07-02] implement | 全仓问题修复记录

- 更新页面：todo。
- 变更摘要：
  - 记录 batchctrl graceful stop 卡死、HDS unary 未实现、服务删除订阅图清理、前端标准脚本和 lint 入口修复。
  - 补充本轮 Go、前端和 context-kg 验证命令结果。

## [2026-07-03] implement | 前端构建 warning 清理

- 更新页面：todo。
- 变更摘要：
  - 更新 Browserslist/caniuse 锁文件数据，消除 Browserslist 过期 warning。
  - 恢复 Vite CSS code split，并按依赖族配置 Rollup manualChunks，消除大 JS/CSS chunk warning。
  - 新增 Vite build 包装脚本，为 Node 22+ 提供有效 localStorage 文件路径，消除 `--localstorage-file` warning。

## [2026-07-05] implement | 配置分组详情设计适配与多灰度 spec 契约归档

- 新增页面：technical/adr/adr-config-gray-release-spec-contract.md。
- 更新页面：index, todo, lessons。
- 变更摘要：
  - 配置分组详情页按设计交接调整为文件树 + 中央业务画布，并补配置编辑、发布记录、订阅查询反回归脚本。
  - 明确方案 B 的多 active gray、灰度转正式草稿和停止单个灰度需要同步更新 specification。
  - 记录构建后必须重启 console/all-mode 并验证入口 hash，避免浏览器或运行进程持有旧静态资源。

## [2026-07-05] ingest | OTel 可观测性平台与 Kubernetes 部署方案归档

- 新增页面：technical/adr/adr-otel-observability-platform.md。
- 更新页面：index, todo。
- 变更摘要：
  - 明确系统内部可观测性与业务服务调用可观测性两个平面。
  - 选择 OpenTelemetry Collector Contrib 作为统一采集管道，OpenObserve 作为默认存储查询后端。
  - 明确业务普通日志不进入平台，只有结构化 event 与 audit 走 OTel Logs。
  - 设计 Kubernetes 快速体验路径，由 Collector、OpenObserve、pole-control-plane 和样例业务服务组成。

## [2026-07-05] refine | OTel metrics 与 event 命名规范归档

- 更新页面：technical/adr/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - 定义 Pole 自定义 metrics 使用 `pole.` 前缀，HTTP/RPC 优先复用 OTel 标准语义指标。
  - 明确 metrics label 只放低基数字段，`rule.id`、`config.file`、`instance.id`、`trace_id` 等高基数字段只进入 event 或 trace。
  - 定义平台内部 event、审计 event 和业务 event 的 EventName 与关键 attributes。
  - 明确现有 `prometheus` entry 保持旧指标名，新增 `otel` entry 才输出新 `pole.*` 名称。

## [2026-07-05] refine | OTel metrics 覆盖与联动设计补充

- 更新页面：technical/adr/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - 补齐平台内部 metrics 场景：启动、ready、插件、EventHub、队列、健康检查、配置 Watch、治理发布、推送、鉴权、DB 连接池和 telemetry export。
  - 明确业务 CPU/Mem 由 Kubernetes/Collector 资源指标采集，并通过 `service.instance.id`、`k8s.pod.uid`、`pole.service.name` 与业务请求指标联动。
  - 明确治理 metrics 只做低基数聚合，治理 event 通过 `pole.governance.decision.id`、`rule.id`、`trace_id/span_id` 定位具体规则和调用链。
  - 补充 Console 同时间窗联动展示请求、资源、治理事件和审计信息的分析规则。

## [2026-07-05] refine | OTel 服务绑定标签与语言体系补充

- 更新页面：technical/adr/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - 增加 `pole.io/*` 服务/实例保留标签，映射到 Pole、Kubernetes、OTel runtime 资源属性。
  - 增加 `pole.io/runtime-language` / `pole.runtime.language`，用于 Console 选择 Java、Go、Rust、Node.js、Python、.NET 等运行时面板。
  - 明确 JVM、Go runtime 指标使用 OTel 标准 metrics name，Pole 只定义绑定属性和展示规则。
  - 明确服务调用、CPU/Mem、runtime metrics、trace 和 event 的绑定优先级。

## [2026-07-05] restructure | 可观测性 ADR 子目录与 Rust SDK 职责归档

- 新增目录：technical/adr/observability/。
- 移动页面：technical/adr/adr-otel-observability-platform.md → technical/adr/observability/adr-otel-observability-platform.md。
- 新增页面：technical/adr/observability/adr-pole-rust-client-observability.md。
- 更新页面：index, todo。
- 变更摘要：
  - 将可观测性相关 ADR 单独收纳到 observability 子目录。
  - 明确 `pole-rust-client` 负责 Resource attributes 注入、SDK metrics/events、治理决策关联和 tracing hooks。
  - 明确 Rust SDK 不采集业务普通日志、不采 Kubernetes CPU/Mem、不替代业务框架 instrumentation。
  - 明确治理 metrics、event、trace 通过 `pole.governance.decision.id` 关联。

## [2026-07-07] refine | 可观测性后端选型调整为 GreptimeDB 优先

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, technical/adr/observability/adr-pole-rust-client-observability.md, index, todo。
- 变更摘要：
  - 将长期默认观测后端从 OpenObserve 调整为 GreptimeDB。
  - 保留 OpenObserve 作为 quickstart / 可选 provider，用于快速体验和内置 UI 辅助排查。
  - 明确 Collector 仍是统一采集/处理/路由入口，GreptimeDB/OpenObserve 都不替代 Collector。
  - 明确 Console 通过 `observability-query` provider 查询 GreptimeDB，不允许前端直连观测后端。

## [2026-07-12] ingest | 实例 TCP/HTTP 主动健康检查设计

- 新增页面：technical/adr/adr-instance-active-healthcheck.md。
- 更新页面：business/feature-registry/service-discovery.md, index, todo, lessons。
- 变更摘要：
  - 明确 Console 手工实例禁用心跳，开放 TCP/HTTP 控制面主动探测。
  - 明确 specification、checker 插件、调度器、MySQL 映射与 Console 必须端到端适配。
  - 明确复用 `health_check.ttl` 保存主动探测间隔，HTTP 路径通过实例保留元数据持久化。

## [2026-07-12] implement | 实例 TCP/HTTP 主动健康检查落地

- 更新页面：technical/adr/adr-instance-active-healthcheck.md, business/feature-registry/service-discovery.md, todo, lessons。
- 变更摘要：
  - 完成 TCP/HTTP specification、服务端探测插件、调度、MySQL、Console 创建与编辑链路。
  - 修复无健康 checker 时主动检查无法调度的问题，并增加非隔离节点兜底。
  - 修复 MySQL 先构造健康检查后解析 metadata 导致 HTTP 路径回退的问题。
  - 通过真实 HTTP 服务验证实例健康到不健康状态切换，并完成测试数据清理。

## [2026-07-12] ingest | Console Fluent UI v9 设计系统迁移

- 新增页面：technical/adr/adr-console-fluent-ui-design-system.md。
- 更新页面：index, todo, lessons。
- 变更摘要：
  - 将 Console 全站组件与视觉体系确定为 Microsoft Fluent UI React v9。
  - 建立仓库级 Fluent 适配边界，保留复杂表单状态契约，统一用户可见控件。
  - 明确蓝、灰、白企业软件视觉规范和 TDesign 直接 import 禁止规则。

## [2026-07-12] implement | Console Fluent UI v9 全站落地

- 更新页面：technical/adr/adr-console-fluent-ui-design-system.md, todo, lessons。
- 变更摘要：
  - 完成 Fluent Provider、品牌主题、导航、表格、表单可见控件、弹层、通知和统一图标迁移。
  - 将旧 Form 状态控制器与少数复杂控件限制在 `components/Fluent` 适配边界，页面禁止直接导入 TDesign。
  - 修复 Layout 静态子组件、定位枚举、受控输入、Portal 点击拦截和 UI 手工拆包循环问题。
  - 通过全部前端验证脚本、生产构建、all-mode 重启以及主要页面桌面/移动端真实浏览器验证。

## [2026-07-12] refine | Fluent 侧边栏视觉纠偏

- 更新页面：todo, lessons。
- 变更摘要：
  - 移除 Fluent Nav 普通菜单项常驻灰色 surface，恢复白色连续企业控制台导航面。
  - 统一一级模块、二级缩进、hover、浅蓝选中态和左侧品牌色指示条。
  - 修正折叠态图标与箭头，并增加 Error Toast 内容归一化和静态回归检查。
  - 补齐 FluentProvider 根 DOM 的满高链路，确保侧边栏和 Footer 在长视口下占满页面。
  - 完成展开、折叠、窄屏、全部菜单跳转和 8080 生产环境验证。

## [2026-07-12] refine | 可观测性 Console 查询接口与 Rust SDK 动态配置设计

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, technical/adr/observability/adr-pole-rust-client-observability.md, index, todo。
- 变更摘要：
  - 新增 `/observability/v1` Console 查询接口设计，覆盖服务概览、拓扑、metrics、runtime、events、traces、治理、平台、审计和跨信号关联。
  - 明确 `observability-query` provider 屏蔽 GreptimeDB/OpenObserve 查询语法，Console 前端只传 Pole 领域条件和时间窗。
  - 设计 `pole-rust-client` 通过 Pole 服务发现获取 `pole-otel-collector` OTLP 上报地址，减少用户显式配置。
  - 设计 `pole-rust-client` 通过配置中心 `pole-sdk/observability.yaml` remote 下发采样、开关、batch/export 和 event 策略。
  - 明确 SDK remote config 最终映射为标准 OTel exporter 配置，不改变 OTLP 协议，也不绕过 Collector。

## [2026-07-17] implement | 系统命名空间删除保护与 Console 标识

- 更新页面：business/feature-registry/namespace.md, index, todo。
- 变更摘要：
  - 明确 `default` 与 `pole-system` 是不可删除的系统自带命名空间。
  - 明确后端删除入口与列表 `deleteable` 字段共同表达系统保护，不依赖 Console 单侧限制。
  - 明确 Console 对 `pole-system` 展示内部系统空间标识，并为两个系统空间提供专属禁删提示。

## [2026-07-17] implement | Console 完成零 TDesign Fluent UI 迁移

- 更新页面：technical/adr/adr-console-fluent-ui-design-system.md, todo。
- 变更摘要：
  - 以本地 Fluent 适配实现替换 Form、Popup、Tree、Transfer、日期时间等最后一批旧控件。
  - 从依赖声明、锁文件和生产依赖树移除 `tdesign-react` 与 `tdesign-icons-react`。
  - 清理旧 CSS 类名和主题变量，并增加覆盖依赖、源码、样式及原生交互标签的零残留门禁。
  - 保持页面配置抽屉和主要资源页面的迁移前布局与视觉，完成真实浏览器交互验证。

## [2026-07-18] refine | 可观测性数据体系首轮对接计划

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, technical/adr/observability/adr-pole-rust-client-observability.md, index, todo。
- 变更摘要：
  - 明确首轮先打通 control-plane、Rust SDK、sidecar 到 Collector、GreptimeDB、`/observability/v1`、Console 的真实数据闭环。
  - 补充 SDK 与 sidecar 共存时的上报分工，避免服务请求和治理 metrics 重复计数。
  - 明确 control-plane 侧优先补 `history/discoverEvent/statis` 的 `otel` entry，不绕开现有 chain。
  - 记录当前实现差距：Rust SDK 已有观测语义模型但缺真实 exporter 和服务发现/remote config 生命周期，sidecar 尚缺 OTel 上报模块。

## [2026-07-18] refine | 明确 observability 查询接口归属 console 模块

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, todo, lessons。
- 变更摘要：
  - 明确 `/observability/v1` 由 `pole-control-plane` 内的 console 模块提供，属于 Console 查询分析后端。
  - 区分 control-plane server 自身 OTel 上报职责与 console router/handler/provider 查询职责。
  - 修正首轮对接计划，避免把 `/observability/v1` 写成 core apiserver 职责。

## [2026-07-18] implement | 可观测性本地栈与 Console 查询骨架

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - 新增 `deploy/observability` Docker Compose，本地拉起 GreptimeDB standalone 与 OpenTelemetry Collector Contrib。
  - Collector 统一接收 OTLP gRPC/HTTP，并将 traces、metrics、结构化 event/audit logs 写入 GreptimeDB `/v1/otlp`。
  - console 模块新增 `/observability/v1/platform/overview` router、handler、provider skeleton 和 `observabilityQuery` 配置入口。
  - 已用 Collector health、GreptimeDB health、OTLP smoke event 写入 `pole_events` 和 Go 回归测试验证首轮骨架。

## [2026-07-18] implement | control-plane statis OTel metrics entry

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - 新增 `plugin/observability/statis/otel`，作为 `statis` chain 的 OTel metrics entry。
  - 将 API、Store、内部组件、缓存、服务发现、配置中心和客户端发现调用映射为 `pole.*` metrics。
  - 默认部署配置启用 `local + otel + prometheus`，并在启动阶段先初始化 statis chain，确保 OTel MeterProvider 先于旧 helper 指标注册。
  - 已通过定向 Go 测试、本地 Collector 导出 smoke 和 GreptimeDB 指标表查询验证。

## [2026-07-18] implement | 系统监控页接入 GreptimeDB 真实查询

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - `observability-query` provider 从静态 DTO 改为 GreptimeDB SQL 查询，读取 `pole_control_plane_request_count_total` 与 duration sum/count 指标表。
  - `/observability/v1/platform/overview` 返回 provider 状态、摘要 stats、时间序列和系统监控页组件行。
  - 系统监控页优先调用 console 后端接口展示“实时数据”，请求失败或无数据时保留 mock 预览。
  - 已通过 console Go 测试、前端静态检查/lint/build、本地 all-mode API 和真实浏览器页面验证。

## [2026-07-19] implement | 系统监控资源看板与 Go runtime 看板拆分

- 更新页面：technical/adr/observability/adr-otel-observability-platform.md, todo。
- 变更摘要：
  - 将系统监控接口明细中的 CPU/MEM 移出，独立为组件资源看板；真实模式下只使用 `overview.resources`，暂无 kubeletstats 表时显示空态。
  - `/observability/v1/platform/overview` 增加 `runtime` 和 `resources` 字段，Go runtime 从 GreptimeDB `process_runtime_go_*` 表查询。
  - `pole-control-plane` Go 看板展示 goroutine、heap alloc/inuse/sys、GC count、GC pause，并预留 Go scheduler 调度延迟查询。
  - 已通过 console Go 测试、前端静态检查/lint/build、本地 all-mode API 和真实浏览器页面验证。

## [2026-07-19] fix | 系统监控 Go runtime 看板布局修正

- 更新页面：todo, lessons。
- 变更摘要：
  - 将系统监控页的 Go runtime 看板调整为整行面板，避免在双列 dashboard grid 中只占左列并留下右侧空白。
  - 将 runtime 指标卡改为更紧凑的桌面网格，并增加前端静态约束防止整行布局回归。
  - 用真实浏览器验证 1280 与 1920 宽度下 runtime 面板宽度等于所在 grid 宽度，截图归档到 `output/playwright/system-monitor-runtime-layout-wide.png`。

## [2026-07-19] fix | 系统监控 runtime 小图 hover 数据补齐

- 更新页面：todo, lessons。
- 变更摘要：
  - 将系统监控页 Go runtime 小图从纯静态 SVG 折线升级为可 hover/focus 的轻量趋势图。
  - hover 数据展示指标名、采样点序号和值，并按 runtime 指标 unit 进行格式化。
  - 用真实浏览器 hover `Heap Alloc` 小图验证 tooltip 显示 `Heap Alloc / 采样点 6 / 23.28MiB`，截图归档到 `output/playwright/system-monitor-runtime-hover-tooltip.png`。

## [2026-07-19] refine | 系统监控 Grafana-like 看板视图

- 更新页面：todo。
- 变更摘要：
  - 将系统监控页调整为 dashboard header、variables、stat panels、query panels 的 Grafana-like 信息结构。
  - 顶部展示 `Last 1 hour`、`Step 1m`、`GreptimeDB` 和实时数据状态；variables 区展示 `$category`、`$api`、`$component`。
  - 每个 panel 标题栏展示 query 元信息，主接口延迟趋势补齐 hover tooltip，可查看采样时间和值。
  - 用真实浏览器验证新版 `/metrics/system` 视图和主趋势图 hover，截图归档到 `output/playwright/system-monitor-grafana-dashboard.png`。

## [2026-07-19] fix | 系统监控图表坐标轴补齐

- 更新页面：todo, lessons。
- 变更摘要：
  - 为接口延迟趋势补充 Y 轴刻度和底部时间线。
  - 为延迟热力图补充顶部时间轴和左侧接口/组件维度。
  - 为每个 Go runtime 小图补充简化 Y 轴刻度和底部时间线。
  - 用真实浏览器验证新版 `/metrics/system` 图表坐标轴，截图归档到 `output/playwright/system-monitor-dashboard-axes.png`。

## [2026-07-19] refine | 系统监控 runtime 对齐 Grafana Stat panel

- 更新页面：todo, lessons。
- 变更摘要：
  - 按 Grafana 官方口径区分 Time series 与 Stat panel：主趋势图保留完整 x/y 轴，runtime 单值指标使用 Stat + sparkline。
  - 移除 runtime 小卡片内显眼坐标刻度和大气泡标签，改成弱化 sparkline、底部时间范围和紧凑 hover tooltip。
  - 用真实浏览器验证 `GC Pause Avg` hover 显示 `GC Pause Avg / 12:25 / 1.88ms`，截图归档到 `output/playwright/system-monitor-grafana-stat-sparkline-hover.png`。

## [2026-07-19] refine | 系统监控 variables 顶部变量行优化

- 更新页面：todo, lessons。
- 变更摘要：
  - 将系统监控页 variables 区从带阴影的大卡片收敛为 dashboard 顶部轻量变量行。
  - 去掉 `$category`、`$api`、`$component` 的厚重前缀盒，改为变量 Label + 变量名小标签。
  - 保留变量顺序、筛选联动和重置动作，并补充静态约束防止视觉结构回退。

## [2026-07-19] fix | 系统监控 runtime Stat sparkline 比例修正

- 更新页面：todo, lessons。
- 变更摘要：
  - 扩大 Go runtime Stat 卡片内 sparkline 的逻辑画布和可视高度，避免趋势线缩成短横线。
  - 增加 runtime 卡片高度，让当前值与趋势图形成更均衡的上下结构。
  - 保留 Stat panel、时间范围和 hover tooltip，不回退为完整坐标轴小图。

## [2026-07-19] refine | 系统监控 runtime Stat 对齐 Grafana 风格

- 更新页面：todo, lessons。
- 变更摘要：
  - 将 Go runtime Stat 卡片收敛为 Grafana panel 风格：薄边框、低圆角、平面背景和紧凑标题。
  - 将 sparkline 调整为下半区 area sparkline，保留当前值作为主视觉。
  - 弱化 runtime 分类标签和说明文案，减少 Fluent 卡片装饰感。

## [2026-07-19] fix | 系统监控时间轴按本地时区展示

- 更新页面：todo, lessons。
- 变更摘要：
  - 移除系统监控页固定 `TIME_LABELS`，按当前查询窗口动态生成时间标签。
  - 图表、热力图和 runtime sparkline 的时间轴与 tooltip 统一使用浏览器本地时区格式化。
  - Dashboard header 增加本地时区标识，例如 `UTC+08:00`，避免误解为 UTC 展示。

## [2026-07-19] feature | 操作审计与服务事件接入 OTel logs

- 更新页面：todo, lessons。
- 变更摘要：
  - 在 `history` 与 `discoverEvent` chain 增加 `otel` entry，将操作审计和服务事件映射为 OTel logs 并写入 Collector logs pipeline。
  - 新增 fail-open logs exporter：业务链路只入本地 bounded queue / bbolt spool，Collector 异常不阻塞接口，也不影响旧 `logger/rds` entry；Collector 恢复后补发未确认记录；默认 history/discover-event 共享一个 bbolt 文件并用 bucket 隔离。
  - Console 新增 `/observability/v1/events` 与 `/observability/v1/operations`，从 GreptimeDB `pole_events` 查询结构化 event/audit，并让页面优先读取新接口、旧 `/metrics/v1` 兜底。
  - 已通过 Go/前端测试、本地 Collector/GreptimeDB smoke 和 all-mode 真实接口验证。

## [2026-07-19] ingest | Pebble 本地 protobuf value cache 设计

- 新增页面：adr-local-pebble-protobuf-value-cache。
- 更新页面：cache-layer, index, todo, lessons。
- 变更摘要：
  - 确认 Pebble 用于存放可重建的 protobuf marshal bytes，优先覆盖实例列表、配置文件、服务契约和治理规则发布快照。
  - 明确 MySQL、注册内存态和治理 cache 仍是事实来源；实例心跳继续留在内存热路径。
  - 明确 key 必须包含 revision、release_id 或过滤/权限可见性 hash，避免不同请求上下文误命中同一 value。

## [2026-07-19] refine | Pebble value cache 开关与性能硬约束

- 更新页面：adr-local-pebble-protobuf-value-cache, cache-layer, todo, lessons。
- 变更摘要：
  - 明确本地 Pebble value cache 必须默认关闭，开关启用后才使用“内存索引 + Pebble value”模式卸载客户端大数据。
  - 将适用范围扩大为所有面向客户端返回、可重建且 value 较大的数据，包括服务发现、配置、服务契约和治理下发。
  - 将“开启后性能不能低于当前路径”写入验收标准，并要求补齐服务发现、配置发现、服务契约和治理下发基准。

## [2026-07-19] implement | 配置契约与 OTel event 队列切换 Pebble

- 更新页面：adr-otel-observability-platform, adr-local-pebble-protobuf-value-cache, cache-layer, todo, lessons。
- 变更摘要：
  - 配置文件 active release content 与服务契约本地 value cache 从 bbolt 切换为 Pebble，内存继续只保留轻量索引。
  - OTel history/discoverEvent 本地可靠队列从 bbolt spool 切换为 Pebble queue，按 key prefix 区分 audit 与 service event。
  - 保持 Collector 异常时 fail-open：业务链路只写本地 queue，导出失败保留待重试，导出成功后删除已确认 key。

## [2026-07-19] cleanup | 移除测试套件 boltdb 残留

- 更新页面：overview, todo, lessons。
- 变更摘要：
  - 将 `test/data/service_test.yaml` 的默认 store 从旧 `boltdbStore` 切换为 `defaultStore`。
  - 删除 `test/suit` 中 `bolt-data.yaml` 注入、`polaris.bolt` 清理和 bbolt helper。
  - 删除已无引用的 `test/data/bolt-data.yaml` fixture，并清理 bbolt 依赖与 checksum。

## [2026-07-19] decision | 服务调用鉴权采用托管服务身份与自定义 Header 双模式

- 新增页面：adr-managed-service-identity-authentication。
- 更新页面：auth-system, index, todo, lessons。
- 变更摘要：
  - 将服务调用鉴权从管理面用户 Token/资源授权中拆开，定义为数据面服务间信任。
  - 默认由 control-plane 管理不可见、不可修改的服务身份，并向合法 workload 签发可轮换短期凭证；被调方入站 SDK/sidecar 本地验证，control-plane 不进入业务请求热路径。
  - 保留用户自定义 Header/value 作为显式兼容模式，并要求密文/摘要存储、脱敏读取和精确匹配。

## [2026-07-19] refine | 明确服务身份领取与 Discover 边界

- 更新页面：adr-managed-service-identity-authentication, todo。
- 变更摘要：
  - 明确 service token 只用于 SDK 与 control-plane 认证，ServiceIdentity 用于服务间数据面身份，两者不能互相替代。
  - 新增 `SERVICE_IDENTITY` Discover 方向：token 校验通过后下发稳定身份描述，允许按服务 revision 缓存。
  - 实例级 workload credential 仍通过独立 Issue/Renew RPC 定向返回，禁止进入普通 Discover/failover cache。
  - 身份 Discover 强制使用 gRPC metadata token、服务端 principal 推导和 fail-closed 校验；不信任请求体自报服务，也不复用含 token 的 Service message。

## [2026-07-20] implement | 托管服务身份第一阶段闭环

- 更新页面：adr-managed-service-identity-authentication, todo。
- 变更摘要：
  - specification 增加 managed/custom 认证模式、caller selector 与 `SERVICE_IDENTITY` Discover descriptor。
  - control-plane 为真实服务创建隐藏身份，历史服务按 metadata service token 惰性补齐，并拒绝 body token 覆盖、重复 token 和服务声明错配。
  - Rust SDK 使用独立 `controlPlaneToken` 建立身份订阅，descriptor 仅保存在 Engine 内部；Console 默认展示托管身份并明确凭证闭环仍未完成。
  - Custom Header 写入值在持久化前转为 SHA-256 摘要并清空明文，管理读取、预览和 Discover 不返回原值。

## [2026-07-20] implement | WorkloadCredential 数据面凭证闭环

- 更新页面：adr-managed-service-identity-authentication, index, todo。
- 变更摘要：
  - specification 增加 Ed25519 JWT Issue/Renew、SERVICE_TOKEN binding 与 Trust Bundle Discover 契约。
  - control-plane 通过 metadata service token 推导主体并签发短期凭证，私钥仅从文件引用加载，支持 ACTIVE/VERIFY_ONLY 滚动轮换。
  - Rust SDK 在内存中领取和续期凭证，提供 HTTP/tonic 显式注入与入站验签，并仅从验签结果生成 `AuthenticatedCaller`。
  - 托管身份启用时 SDK 强制使用 grpcs，control-plane 身份入口拒绝非 TLS transport；同时补齐 descriptor 周期刷新与同版本 bundle 安全续租，保证密钥轮换和长期运行不中断。
  - 明确 V1 bearer 的重放风险、服务级而非实例级绑定，以及 SDK 不拥有业务网络栈时必须由业务显式挂载适配器。

## [2026-07-20] refine | 托管身份 TLS 改为按需启用

- 更新页面：adr-managed-service-identity-authentication, todo, lessons。
- 变更摘要：
  - Rust SDK 不再要求启用托管身份时必须使用 `grpcs`，继续同时支持 `grpc` 与 `grpcs` endpoint。
  - control-plane 不再拒绝身份 Discover、Issue/Renew 的非 TLS transport，请求是否启用 TLS 由部署方决定。
  - 文档保留生产环境启用 TLS 的安全建议，并明确明文 bearer 的链路窃听风险，但不把建议升级为功能前置条件。
## [2026-07-20] implement | 命名空间工作台滚动与配置文件统计

- 更新页面：namespace、todo、lessons。
- 变更摘要：
  - 命名空间列表的配置文件数改为后端批量聚合的真实文件数量，并通过协议字段返回到摘要与行级表格。
  - 固定页头、指标、工具栏和分页，仅让表格内容区滚动；浏览器验收确认滚轮不会再驱动页面滚动。
## [2026-07-20] fix | 配置分组删除保护与失败通知

- 更新页面：todo、lessons。
- 变更摘要：
  - 修复 Console 遗漏 `file_count` 的响应归一化，配置文件占用的分组不再显示为零文件或错误开放删除。
  - 后端继续拒绝删除含有效配置或活跃发布的分组；Toast 将 RequestId 从正文收纳为 Fluent 复制操作。

## [2026-07-20] refine | Toast 自动与手动关闭

- 更新页面：todo、lessons。
- 变更摘要：
  - Toast 统一使用独立 ID 和 Fluent 关闭按钮，错误与成功通知均在 10 秒后自动关闭。
  - 浏览器验证覆盖错误通知手动关闭与超时自动移除，以及成功通知关闭入口。

## [2026-07-20] refine | 紧凑错误 Toast 复制动作

- 更新页面：todo、lessons。
- 变更摘要：
  - 将底部 Request ID / 错误 JSON 双复制按钮合并为标题右上角单一“复制”按钮，仍复制完整 JSON。
  - 复制动作保留图标、短文案与可访问标签，关闭按钮独立保留。

## [2026-07-20] refactor | Console 治理鉴权凭证交互收尾

- 更新页面：todo。
- 变更摘要：
  - 将 Custom Header 明文从规则 state 拆为一次性 write-only draft，并在加载、关闭、撤销、模式切换和保存后清空。
  - Console 加载详情时主动丢弃服务端摘要，提交时按认证模式重新构造最小 authentication 载荷，避免摘要或旧凭证回写。
  - 已保存的兼容凭证进入编辑态时明确提示必须重新输入才能轮换，详情和实时 Spec 始终脱敏。
  - 使用真实 Console 完成托管默认值、模式切换、首次保存、轮换、回显和测试数据清理验收。

## [2026-07-20] fix | 错误 Toast 正文宽度

- 更新页面：todo、lessons。
- 变更摘要：
  - 将 Fluent Toaster 容器设为受视口约束的 420px 桌面宽度，避免 Toast 子项溢出容器。
  - 错误正文跨越 Fluent 默认网格的右侧空列；真实 400203 通知已确认中文信息不再过早换行，复制与关闭操作完整可见。

## [2026-07-20] refactor | 治理规则编辑器移除实时 Spec

- 更新页面：governance-rules、patterns、todo、lessons。
- 变更摘要：
  - 路由、泳道、限流、熔断、探测、无损、调用鉴权、流量镜像与 Mock 统一不展示动态 Spec、YAML/JSON 预览。
  - 移除调用鉴权实际 Spec 面板及全套编辑器遗留预览样式，新增静态门禁防止预览列回归。
  - 治理编辑器长期约定改为单栏表单和内部滚动，校验由字段校验、保存和错误提示承担。

## [2026-07-20] refine | 治理鉴权规则编辑布局与来源服务选择

- 更新页面：governance-rules、patterns、todo、lessons。
- 变更摘要：
  - 受保护接口行按真实内容区收缩，路径列弹性扩展，操作列不再溢出。
  - Fluent Select 的可搜索模式会实际过滤候选项；来源服务多选器限定为紧凑宽度。
  - 编辑器 shell 使用自然高度，长规则由详情 Tab 内容区承担滚动。

## [2026-07-20] refine | 治理规则类型即选即建

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 新建规则类型卡片改为直接创建入口，移除步骤条和“进入创建”二次确认。
  - 每张卡片保留明确的可访问名称，鼠标与键盘均可直接打开对应规则编辑器。

## [2026-07-20] refactor | 治理规则独立创建页

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 新建规则由工作台抽屉改为独立创建路由，并复用规则详情页面框架。
  - 类型卡片直接导航；服务详情上下文通过 URL 传递给创建表单，保存或取消后回到治理工作台。

## [2026-07-20] fix | 治理独立创建页滚动与遗留浮动控件

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 创建页建立有限高度 flex 链，仅由表单内容区处理长内容滚动，避免滚动落到外层工作区。
  - 页头已接管保存/撤销操作时，隐藏遗留 `StickyTool` 固定容器，消除右侧空白方块。
  - 真实鉴权创建页验证内容区可用滚动且外层不随滚轮移动；8080/8090 均正常返回。

## [2026-07-20] fix | 鉴权规则 API 路径匹配语义

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 受保护接口移除“值类型”控件，仅保留协议、方法、路径匹配类型、路径和操作。
  - 鉴权 API 读写适配层清理遗留 `value_type`，参数匹配条件继续保留该字段。
  - 更新专项回归脚本，覆盖页面结构及 API 路径提交载荷。

## [2026-07-20] fix | 治理 API 协议化编辑交互

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - HTTP、gRPC、Dubbo 共用存储结构但使用各自的表单语义：HTTP 方法/URI、gRPC 方法名/服务名、Dubbo 方法名/接口名。
  - 切换协议清空旧 `method/path` 值，HTTP 重新使用 `GET` 默认值，避免将 HTTP 默认条件误保存到 RPC 规则。
  - 鉴权、镜像和 Mock 的 API 编辑行均复用相同协议展示与重置逻辑。

## [2026-07-20] fix | RPC 接口优先编辑与可选方法

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - gRPC/Dubbo 将服务名或接口名作为主匹配对象，置于方法之前；方法明确为可选的细化条件。
  - HTTP 保持方法、匹配类型、路径的既有顺序，混合协议表头明确说明两种字段语义。
  - 专项回归覆盖 gRPC 服务名匹配且方法为空时的校验与提交行为。

## [2026-07-20] fix | 鉴权服务信息垂直可折叠布局

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 命名空间和服务名称调整为上下顺序，保留紧凑且受限宽度的选择器布局。
  - 服务信息接入统一折叠分段；收起时显示已选择的命名空间和服务名称摘要。
  - 专项回归、构建、重建与浏览器交互验收均已完成。

## [2026-07-20] fix | RPC 接口字段语义与宽度

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 明确 RPC 的 `path` 为 service/interface、`method` 为可选方法，修正混合表头的歧义。
  - RPC 行为 interface 分配更宽输入列，method 使用较短的可选输入列。
  - 专项回归先验证旧表头失败，再通过构建、重建和 gRPC 实页切换验收。

## [2026-07-20] fix | 鉴权认证方式折叠

- 更新页面：governance-rules、todo、lessons。
- 变更摘要：
  - 认证方式改为统一可折叠分段，默认展开并在关闭或切换规则时复位。
  - 收起状态只显示当前认证模式摘要，不展示 Header 值或其它凭证编辑内容。
- 专项回归、构建、重建和真实浏览器交互验收均已完成。

## [2026-07-20] ingest | Console Agent 资源变更工作台需求调研

- 新增页面：adr-console-agent-resource-workbench。
- 更新页面：ai-features、index、todo。
- 变更摘要：
  - 明确 Agent 工作台与 A2A Agent Registry、pole-server 领域核心的边界。
  - 确认配置文件和治理规则才具备真实的编辑态与发布态分离，其他直接生效资源首期只读。
  - 采用 Console 会话深模块与内部强类型资源 adapter，模型只生成提案，确认只保存草稿，发布仍由用户在现有页面完成。
  - 固化 proposal hash、幂等、stale baseline、二次鉴权、敏感字段和审计要求，并给出分期与验收标准。

## [2026-07-20] fix | 鉴权兼容模式恢复子规则 Header 匹配

- 更新页面：governance-rules、adr-managed-service-identity-authentication、todo、lessons。
- 变更摘要：
  - Console 的“自定义 Header（兼容模式）”统一写入 `LEGACY_REQUEST_MATCH`，不再创建规则级 `CUSTOM_HEADER` 凭证。
  - Header 匹配条件随受保护接口保存在每条 `TrafficSecurityPolicy.traffic_match_rule`，各子规则相互独立。
  - 服务端继续解析历史 `CUSTOM_HEADER`，以维持存量资源读取兼容；Console 编辑保存时回归原有子规则请求匹配模型。

## [2026-07-20] ingest | 治理请求参数采集与动态消费

- 新增页面：adr-governance-request-parameter-capture。
- 更新页面：governance-rules、patterns、index、todo、lessons。
- 变更摘要：
  - 统一 `TEXT/PARAMETER/VARIABLE` 为固定匹配、当前键请求值采集和运行变量引用。
  - Proxyless Rust SDK 的本地限流按采集值隔离计数器，路由目标标签可消费同名采集参数。
  - Console 横向保留值来源并明确 xDS 当前不消费动态语义，避免把可配置误报为全数据面支持。

## [2026-07-20] release | specification ALPHA.34 与依赖联动

- 更新页面：todo。
- 变更摘要：
  - 发布 specification `v0.1.0-ALPHA.34`，并完成 Go/Rust 生成代码与 crates.io 产物验证。
  - control-plane 与 Rust SDK 均切换到 ALPHA.34，以显式暂存隔离各仓库已有的其它未提交改动。
  - 记录 specification、control-plane 与 Rust SDK 的提交、推送和全量测试结果。

## [2026-07-20] refactor | 移除治理运行变量值来源

- 更新页面：governance-rules、adr-governance-request-parameter-capture、patterns、todo、lessons。
- 变更摘要：
  - 治理匹配值来源收敛为固定值和请求参数，不再支持读取运行机器环境变量。
  - specification 保留历史枚举号与名称不可复用；Rust SDK 对未知旧值 fail closed。
  - Console 所有共享治理条件编辑器统一移除运行变量选项和相关提示。

## [2026-07-20] refactor | 治理匹配条件 Fluent UI 交互与容器布局

- 更新页面：todo、lessons。
- 变更摘要：
  - 共享条件编辑器改用 Fluent `RadioGroup`、`Button`、`Tooltip`、`Select` 与主题 token，并补齐语义化表格和可访问名称。
  - 固定值与请求参数采用明确的状态化展示，移除导致选中值聚焦后清空的筛选式枚举下拉。
  - 响应式改用组件容器查询，使嵌套治理子规则按自身可用宽度从六列表格切换为两列或单列字段卡，避免操作列裁切。
  - 已完成亮色、暗色、下拉切换和请求参数采集的真实浏览器验收。

## [2026-07-21] feat | Console Agent 配置草稿纵向闭环

- 更新页面：adr-console-agent-resource-workbench、ai-features、todo。
- 变更摘要：
  - 新增 `/ai/agent` 确定性工作台与 Console proposal API，支持已有配置文件的查询、临时 diff、确认和待发布回执。
  - 确认链路绑定 actor、baseline/preview hash、TTL 与幂等键，只保存配置草稿，完全不调用发布接口。
  - 增加状态机、HTTP adapter、路由、前端契约和真实 smoke 验证；明确底层配置 API 缺少原子 CAS 的剩余并发边界。

## [2026-07-21] fix | Console Agent 页面持久后台恢复

- 更新页面：todo、lessons。
- 变更摘要：
  - 确认连接拒绝源于临时工具会话退出，不是 Agent 路由或页面代码故障。
  - 改用独立 tmux 会话分别守护 8080 Console 与 8090 pole-server。
  - 固化回合结束后的可访问性验收：后台会话、端口、深链和静态资源必须同时有效。

## [2026-07-21] refactor | Pole Agent 独立对话入口

- 更新页面：adr-console-agent-resource-workbench、ai-features、todo、lessons、index。
- 变更摘要：
  - 将 Agent 从“AI 工具”子菜单迁移为独立一级 `/agent` 入口；A2A 与 MCP Registry 继续留在 AI 工具。
  - 用聊天时间线替换资源表单，在消息流中展示工具调用轨迹，并在写操作时按需打开临时 diff 检查器。
  - 保持 previewHash、幂等、stale 检查和发布隔离；确认后只保存草稿并停在待发布。
  - 明确 Phase 0 仍是确定性 planner + HTTP tool port，真实 ModelPort 与 MCP client transport 是下一阶段硬缺口。

## [2026-07-21] feat | 普通控制台与 Agent 双工作模式

- 更新页面：adr-console-agent-resource-workbench、ai-features、todo、lessons。
- 变更摘要：
  - 默认侧边布局在侧栏底部、版本信息上方新增工作区切换；普通模式进入 Agent，Agent 模式在同一位置返回普通控制台。
  - Agent 模式隐藏普通资源导航但保留侧栏骨架和返回入口；无侧栏的顶部布局保留页头切换兜底。
  - 切换 Agent 前记录最近普通页面；返回时使用安全 `returnTo`，并对外部或循环路径 fail closed。
  - 配置文件详情新增“交给 Agent”，只传资源自然键和返回地址；Agent 首屏接入上下文并生成可继续编辑的对话起点。
  - 保持登录态、权限、API、审计和正式发布链路共享，不复制第二套资源管理实现。

## [2026-07-21] ingest | Console Agent Runtime 与 LLM Gateway 接缝

- 更新页面：adr-console-agent-resource-workbench、ai-features、todo。
- 变更摘要：
  - 确认当前 Phase 0 是前端正则 planner + Console 强类型 REST adapter，不包含真实 LLM 或 MCP client。
  - 明确 LLM Gateway 地址、模型和凭证属于 Console 后端配置，浏览器不得直接持有密钥。
  - 将 Console Agent Runtime 分为 ModelPort、ToolExecutor、现有确认内核和 SessionRepository 四个内部 seam。
  - 记录 Pole MCP 当前缺少配置/治理资源工具，Phase 1 需先补工具能力再宣称 MCP 已连接。

## [2026-07-21] refactor | Pole Agent 一等运行主体

- 更新页面：adr-console-agent-resource-workbench、ai-features、todo、lessons。
- 变更摘要：
  - 根据用户纠正，将 Pole Agent 明确为页面唯一对话对象，而不是 ModelPort、MCP 与 Workbench 的松散组合。
  - Pole Agent 内部拥有版本化 System Prompt、LLM、MCP 工具导入、model-tool loop、会话和 approval resume。
  - 现有 Workbench 重新定位为 Agent 内部的 ChangeApprovalKernel，页面不再承担意图解析、Prompt 拼装或工具选择。

## [2026-07-21] ingest | Console Agent 系统配置来源

- 更新页面：adr-console-agent-resource-workbench、terminology、todo、index。
- 变更摘要：
  - 采用启动 YAML/Secret 自举、Pole 配置中心保存非敏感 AgentDefinition、Console 类型化系统设置页面的三层方案。
  - 配置中心只发布非敏感动态设置，模型密钥和后台凭证只保存 Secret 引用。
  - Agent 对已发布 revision 做完整校验和原子热加载，失败时保留 last-known-good，现有会话固定原 revision。
  - 拒绝 Console 直连核心数据库，避免破坏 console-only 模式并重复配置中心的发布、历史、Watch 和审计能力。

## [2026-07-21] ingest | Server 与 Console 统一系统配置

- 新增页面：adr-system-configuration-control-plane。
- 更新页面：configuration、adr-console-agent-resource-workbench、terminology、todo、index。
- 变更摘要：
  - 将 Agent 配置方案提升为 Pole Server 与 Console 共用的 System Configuration 领域。
  - 静态 YAML 保留自举和安全基线，配置中心承载已发布动态覆盖，Secret Provider 管理密钥。
  - 配置按 BootstrapOnly、RestartRequired、HotReload、GuardedHotReload 分级，没有 applier 的字段禁止热更新。
  - 引入 DesiredSnapshot、EffectiveSnapshot 和 ApplyReceipt，页面可观察实例生效、拒绝、待重启和漂移。
  - 给出只读目录、Console 低风险热更、Server 热更、重启配置与高级治理的分阶段路径。

## [2026-07-21] feat | Pole 本地 Kubernetes 一体化部署

- 更新页面：adr-otel-observability-platform、todo。
- 变更摘要：
  - 新增 `pole-system` namespace 下 Pole Deployment、Collector Deployment、GreptimeDB StatefulSet/PVC 的一体化编排。
  - 使用 ExternalName Service 继续访问宿主机 MySQL `3306`，凭证通过 Secret 注入，不把 MySQL 迁入集群。
  - 修复本地 all-mode 镜像内容，包含 Console 静态资源、运行配置和 Kubernetes 配置渲染入口。
  - OrbStack 实测三个工作负载 Ready，Console/Agent 页面可访问，OTLP event 经 Collector 写入 GreptimeDB 并由 SQL 查回。

## [2026-07-21] feat | Pole Console 接入共享 Gateway

- 更新页面：adr-otel-observability-platform、todo。
- 变更摘要：
  - 复用 `tidemind/tidemind-gateway`，为 Console 增加无需修改 hosts 的 `pole.localhost` 域名。
  - HTTPRoute 留在 Gateway namespace，通过最小 ReferenceGrant 跨 namespace 引用 `pole-control-plane:8080`。
  - GreptimeDB、OTel Collector 与宿主机 MySQL 继续保持内部访问边界。

## [2026-07-22] refactor | Console 登录入口视觉收敛

- 更新页面：adr-console-fluent-ui-design-system、todo、lessons。
- 变更摘要：
  - 将超大宣传标题与悬空表单重构为克制的产品说明区和聚焦登录面板。
  - 移除默认账号明文和没有后端 API 的伪注册入口，强化主按钮、输入焦点、密码可见性和提交状态。
  - 管理员检查异常不再误跳初始化页；保留认证 API 和登录后跳转，并补齐桌面与窄视口真实浏览器验收。

## [2026-07-22] feat | System Configuration Phase 1 只读目录

- 更新页面：adr-system-configuration-control-plane、todo、lessons、index。
- 变更摘要：
  - 新增 `pkg/systemconfig` 定义注册表、有效快照和 default/YAML/env/CLI 来源索引。
  - Pole Server 与 Console 分别注册首批核心设置，Console 通过独立只读接口跨进程聚合 52 项配置。
  - Secret 在服务端清空值；Server 内部快照接口执行独立 `DescribeSystemConfiguration` 策略授权，Console 页面不提供任何编辑或发布动作。
  - 新增独立“系统配置”侧栏入口、来源概览、搜索和组件/子域/来源筛选，并部署到 `pole.localhost`。

## [2026-07-22] refactor | System Configuration 按领域组织

- 更新页面：adr-system-configuration-control-plane、todo、lessons。
- 变更摘要：
  - 将 52 项配置的单一总表改为“组件 → 领域 → 配置项”三级信息架构。
  - Pole Server 与 Console 分别提供稳定的领域顺序，主表只展示当前领域并移除重复的组件/领域列。
  - 跨领域搜索保留在当前组件范围内，领域导航显示命中数并自动将命中领域滚入视野。
  - 完成桌面与 390px 真实页面验收并部署到 Kubernetes；领域导航和配置表在窄屏各自独立滚动。

## [2026-07-22] feat | Console AgentDefinition 配置目录

- 更新页面：adr-console-agent-resource-workbench、adr-system-configuration-control-plane、todo、lessons、index。
- 变更摘要：
  - Console Agent 领域由 2 项扩展为 13 项，补齐 Agent ID、Prompt、LLM Gateway/model/API key 引用、MCP endpoint/工具白名单和运行限制。
  - API key 仅由环境变量或 Kubernetes Secret 注入，系统配置响应删除原值并只返回配置状态与引用。
  - 现有确定性 Workbench 使用配置化 proposal TTL 和上游 timeout，同时明确真实 ModelPort 尚未接通。
  - 修正列表型设置的来源聚合，完成 Go、前端、Kubernetes 与真实登录页面验收。

## [2026-07-22] decision | Pole 托管 Agent Secret

- 更新页面：adr-system-configuration-control-plane、adr-console-agent-resource-workbench、todo、lessons、index。
- 变更摘要：
  - 纠正将 Kubernetes Secret/环境变量作为 LLM API key 最终入口的设计，明确 Pole 默认托管业务 Secret。
  - 定义 `SystemSecretStore` write-only interface、envelope encryption、版本化 `pole-secret://` 引用和运行时受控解析。
  - Kubernetes Secret/环境变量只保存 KEK 等自举材料，Vault 等保留为可选外部 adapter。
  - 现有配置文件加密链不满足系统 Secret 的密钥隔离、轮换和用途约束，不能直接复用。

## [2026-07-22] design | System Configuration 编辑交互与 Admin 门禁

- 更新页面：adr-system-configuration-control-plane、todo、lessons。
- 变更摘要：
  - 编辑采用单字段右侧工作面板，保存只产生领域草稿，审阅和发布保持独立动作。
  - Secret 采用 write-only 输入和版本轮换语义，diff 只显示“将轮换”，不回显新旧值。
  - System Configuration 首期只允许 `main`/admin；前端隐藏入口并守卫直达，后端统一保护全部 `/system-config/v1/*` 接口。
  - Console 签名 JWT 增加角色 claim 并设为 HttpOnly，旧会话仅在用户 ID 匹配 canonical main account 时兼容。
## [2026-07-22] feat | 三个内置系统角色与成员绑定

- 更新页面：terminology、domain-models、business-rules、auth-system、adr-system-configuration-control-plane、todo、lessons、index。
- 变更摘要：
  - 固定 `admin`、`resource-reader`、`resource-writer` 三个系统角色及稳定权限边界。
  - 启动和角色查询时幂等补齐角色与固定策略；服务端禁止角色定义和系统策略的普通增删改。
  - Console 角色页当时移除通用 CRUD，只保留用户与用户组成员绑定；后续由下方纠正记录恢复自定义角色 CRUD。
  - 内置 admin 成员签发 `admin` 会话角色，可与主账号 `main` 一同访问 admin-only 系统配置页。

## [2026-07-22] fix | 恢复自定义角色并收敛内置角色保护边界

- 影响页面：[[terminology]]、[[domain-models]]、[[business-rules]]、[[auth-system]]
- 变更摘要：
  - 角色聚合恢复自定义角色完整 CRUD、成员及权限管理。
  - 三个内置角色保持定义与资源/API 权限不可变，只开放用户和用户组绑定。

## [2026-07-22] fix | 系统配置请求失败不再伪装为空目录

- 更新页面：adr-system-configuration-control-plane、todo、lessons。
- 变更摘要：
  - 修复系统配置请求失败时清空已有设置、将错误渲染成 `0 / 0` 和全零统计的问题。
  - 首次加载失败改为独立错误态和重试入口；刷新失败保留 last-known-good 快照并显示告警。
  - 增加前端错误态契约，完成 Console/配置注册相关测试、Kubernetes 发布和真实 admin 页面复验。

## [2026-07-22] fix | 非 Admin 完全隐藏系统配置页面

- 更新页面：adr-system-configuration-control-plane、todo、lessons。
- 变更摘要：
  - 新增 `/auth/v1/user/session`，由 Console HttpOnly 签名会话返回权威角色和 admin 判定。
  - 前端不再从 localStorage 恢复角色；会话未解析时 fail-closed，非 admin 不显示菜单且直达系统配置页重定向到普通控制台。
  - 本地 Kubernetes 作为唯一运行验收环境；修复一次误将 macOS Mach-O 放入 Linux/arm64 镜像导致的 CrashLoop，并以新 Pod Ready 为发布门槛。
  - K8s 实测 main 访问系统配置为 200、临时 sub 为 403；临时用户验收后已删除。

## [2026-07-23] design | Namespace 统一为运行环境

- 更新页面：terminology、domain-models、business-rules、namespace、service-discovery、config-center、governance-rules、overview、todo、lessons、index。
- 变更摘要：
  - 明确 Namespace 是 Pole 的运行环境边界，不再定义为泛化的多租户容器。
  - 服务名是全局逻辑标识，同名服务在不同 Namespace 中表示同一服务的环境实例。
  - 配置分组、配置文件和治理规则建立一致的跨环境逻辑身份与环境实例坐标。
  - Console 命名空间、服务、配置和治理入口说明同步采用环境语义。
  - 本地 OrbStack 仅更新 `pole-system/pole-control-plane`，部署镜像为 `pole-control-plane:local-20260723-namespace-environment-copy-v2`，运行入口与新静态资源完成验证。

## [2026-07-23] feat | Namespace 环境模型全链路实现

- 更新页面：adr-governance-rule-unified-storage、todo、lessons、index。
- 变更摘要：
  - specification 九类治理聚合根增加顶层 namespace，明确区分规则归属环境与 caller、callee、target 等运行时作用域。
  - 统一治理存储的名称查询、锁、发布与删除保护按 owner namespace 隔离，并增加治理规则占用业务码。
  - Service、Config Group、Config File 详情增加同一逻辑资源的跨环境摘要与切换，服务端逐条过滤无权环境。
  - Console 规则工作台按环境查询和创建；配置文件跨环境查询仅返回摘要，不泄露正文。

## [2026-07-23] fix | Console 暗色主题全站一致性

- 更新页面：adr-console-fluent-ui-design-system、todo、lessons。
- 变更摘要：
  - 修复服务详情局部浅色变量覆盖全局暗色主题，并将治理、AI、认证、监控等页面的浅色背景和边框统一到双主题语义 token。
  - 暗色静态门禁扩展到背景、边框与局部主题变量；前端检查、ESLint 和 production build 均在 Kubernetes Linux/arm64 Pod 内执行。
  - 仅发布 `pole-control-plane` 新镜像，并通过 Pod、Gateway 和 20 个真实页面路由完成暗色表面及控制台回归。

## [2026-07-23] feat | Pole Agent 本地会话与会话记忆

- 更新页面：adr-console-agent-resource-workbench、todo、lessons。
- 变更摘要：
  - Agent 工作台采用 ChatGPT 式本地会话导航、连续消息流、浮动输入框和按需变更检查器。
  - 浏览器 IndexedDB 持久化会话、消息、未发送草稿、资源上下文、当前会话和每会话记忆设置，不作为服务端提案或审计事实来源。
  - 记忆支持关闭及 4/10/20 轮窗口，当前指令优先，历史只补全缺失上下文；preview、确认保存草稿和用户发布边界保持不变。
  - K8s Pod 内完成契约、暗色、ESLint 和生产构建，仅滚动更新 control-plane，并通过 Gateway 真实浏览器验证切换与刷新恢复。

## [2026-07-23] refactor | Pole Agent 单侧栏会话工作台

- 更新页面：adr-console-agent-resource-workbench、todo、lessons。
- 变更摘要：
  - Agent 模式复用最左侧产品导航直接承载本地会话，不再在内容区嵌套第二条会话栏。
  - 新建、切换、重命名和删除统一留在侧栏；主画布移除重复新会话入口和空态建议按钮。
  - 输入区去除双描边，同时保留 IndexedDB 会话记忆、MCP 工具轨迹、临时视图和等待用户发布边界。
  - K8s 容器检查与 Gateway 真实浏览器验证通过，并仅滚动更新 control-plane。

## [2026-07-23] refactor | Pole Agent 参考稿布局对齐

- 更新页面：adr-console-agent-resource-workbench、ai-features、todo、lessons、index。
- 变更摘要：
  - Agent 工作台按指定 HTML 参考稿收敛为 280px 会话侧栏、52px 全局栏、62px 对话栏、底部浮动输入器与 294px 可收起上下文面板。
  - 本地会话增加搜索、日期分组和快捷新建；资源上下文、连接状态、记忆设置与权限边界集中到右侧面板。
  - 临时 diff 改为消息流内工具结果，继续强制预览、确认保存草稿、用户另行发布的安全边界。
  - K8s 容器检查、Gateway 探测和真实浏览器验收通过，仅滚动发布 `pole-control-plane:local-20260723-agent-reference-layout-v1`。

## [2026-07-23] feat | Pole Agent 真实 LLM 与 MCP 最小闭环

- 更新页面：adr-console-agent-resource-workbench、ai-features、configuration、todo、lessons、index。
- 变更摘要：
  - Console 内部新增 OpenAI-compatible ModelPort、版本化 System Prompt、有上限的 model-tool loop 和 actor-bound Pole MCP client。
  - Pole MCP 增加配置文件只读工具；配置 update 仅生成受控临时视图，模型不能确认、发布或删除资源。
  - Agent 页面改为真实会话 API，保留 IndexedDB 本地会话与记忆；运行时未配置时 fail closed，不显示虚假连接。
  - K8s 容器内完成测试、构建、模型/MCP/提案/确认浏览器闭环，并发布 `pole-control-plane:local-20260723-agent-runtime-v5`；临时模型服务和测试改动已清理。

## [2026-07-23] feat | Pole 内部 Agent 配置与 Secret 闭环

- 更新页面：adr-system-configuration-control-plane、ai-features、configuration、todo、lessons、index。
- 变更摘要：
  - Console Agent 的 Gateway、模型、Prompt、MCP 策略与 API key 迁入 Pole 内部 System Settings repository，不再由 Kubernetes 日常环境变量托管。
  - 新增类型化草稿、不可变发布版本、Pole Secret 信封加密、连接测试、发布前强制探活、单实例原子切换和周期 reconcile。
  - 修复普通控制面请求刷新会话时丢失 admin 角色的问题，并兼容升级旧的无 role 主账号 Cookie。
  - K8s 仅保留数据库连接与 `POLE_SYSTEM_SECRET_MASTER_KEY`，所有测试、构建、部署和真实浏览器验收均在容器运行时完成。

## [2026-07-23] refactor | System Configuration Agent 专用工作区

- 更新页面：adr-system-configuration-control-plane、todo、lessons、index。
- 变更摘要：
  - Agent 领域不再复用通用配置表格，改为运行就绪状态、模型连接、Pole Secret、MCP 能力、指令和发布流程组成的专用工作区。
  - 编辑抽屉按模型与凭证、MCP 工具、Agent 指令分组；任意配置变化会使连接测试结果失效，保存草稿和发布保持独立动作。
  - 发布审阅展示 effective 到 draft 的字段级差异、Secret 版本引用和原子热更新保护；移除会生成不可发布草稿的“禁用 API Key”入口。
  - Kubernetes 容器内通过目标 ESLint、Admin gate、错误态、暗色约束和生产构建；真实 Gateway 在 1280、1180、900 三档视口无横向溢出。

## [2026-07-23] feat | System Configuration 全领域字段级编辑

- 更新页面：adr-system-configuration-control-plane、todo、lessons、index。
- 变更摘要：
  - 对 Pole Server 29 项和 Console 34 项逐字段登记编辑性、锁定原因、类型、范围、枚举、敏感策略和生效方式；缺少策略的新字段默认锁定。
  - 非 Agent 领域复用统一 repository 完成草稿、并发控制、发布、历史与 desired/effective 差异；重启级发布明确回执 `pending_restart`。
  - 通用页面增加类型化编辑抽屉、字段差异和发布影响；Agent 补齐提案 TTL、资源工具超时并接入真实运行时更新。
  - Kubernetes 内通过 Go 测试、Vite release build、Pod/Gateway 与真实浏览器验收，发布镜像 `pole-control-plane:local-20260723-system-config-all-v3`。

## [2026-07-24] refactor | Pole 复用 tidemind GreptimeDB

- 更新页面：adr-otel-observability-platform、todo、lessons、index。
- 变更摘要：
  - 移除 `pole-system` 重复的 GreptimeDB standalone 工作负载，通过 `pole-greptimedb` ExternalName 适配 `tidemind/maas-greptimedb-frontend`。
  - Collector 三类 signal 与 Console 查询统一使用 `pole_observability` 逻辑库，保持 Pole 与 MaaS 数据模型隔离。
  - 部署脚本幂等建库，并在新链路 rollout 成功后删除旧 StatefulSet、保留旧 PVC 作为历史数据导出和回退边界。

## [2026-07-25] ingest | 最新 Console 质量与本地部署知识

- 新增页面：console-ui-quality-gates。
- 更新页面：testing、adr-console-fluent-ui-design-system、configuration、todo、index。
- 变更摘要：
  - 将近期全站体验整改提炼为共享表格、表单重置、响应式弹层、可访问性和多视口发布验收契约。
  - 区分接口 E2E、Node 源码契约、构建回归与 Kubernetes 真实浏览器证据，避免用单层检查替代完整 UI 验收。
  - 固化本地 Kubernetes 对共享 GreptimeDB、外部 MySQL、逻辑库隔离、旧 PVC 保留和依赖恢复的运行边界。

## [2026-07-25] refine | Handoff 本地协作资产规范

- 更新页面：todo、lessons。
- 变更摘要：
  - Handoff 统一保存在项目 `.handoff/`，不再写入系统临时目录。
  - 使用 `.git/info/exclude` 做本地排除，禁止 handoff 文件进入暂存、提交或推送。
  - 交付前同时检查 ignore 来源与完整 Git 状态，避免被默认隐藏的未跟踪文件误入版本控制。

## [2026-07-25] refactor | Console 复合查询交互统一

- 更新页面：adr-console-fluent-ui-design-system、todo、lessons、index。
- 变更摘要：
  - 新增应用级 `QueryComposer`，统一主搜索、自动补全、条件 Tag、高级筛选浮层，以及远程显式查询和本地即时过滤两种模式。
  - 命名空间、服务、配置分组、治理工作台、MCP、A2A、系统配置与观测多条件页完成迁移；单字段和详情局部搜索继续保持轻量。
  - 事件指标与操作审计继续使用独立 DateRangePicker，不把时间范围放入高级条件或 Tag。
  - 新增专项契约并通过 ESLint、观测/日期/暗色回归、test build 和 1440/720 浏览器交互验收。

## [2026-07-26] ingest | 多资源域统一治理平台技术方案

- 新增页面：adr-multi-resource-governance-platform。
- 更新页面：terminology、domain-models、governance-rules、todo、lessons、index。
- 变更摘要：
  - 将服务、消息、存储和任务定义为资源域，将路由、限流、运行时鉴权、镜像、Mock 等定义为治理能力。
  - 采用公共规则信封、单效果类型化 Spec、Policy Bundle、能力协商、编译 Adapter、原子 Bundle 和应用回执。
  - 明确 Kafka/RocketMQ、Redis/MySQL、Job 的资源模型、执行点、安全不变量和 fail-open/fail-closed 边界。
  - 给出兼容现有九类服务规则的 Phase 0—5 演进路线、契约测试和完整纳管验收标准。

## [2026-07-26] feat | Pole 自身能力自动注册与自管理闭环

- 新增页面：adr-pole-self-management-control-loop。
- 更新页面：ai-features、adr-console-agent-resource-workbench、adr-system-configuration-control-plane、auth-system、architecture、todo、index。
- 变更摘要：
  - 新增隔离执行身份 `pole-self-manager`，启动和周期性收敛 Control Plane 自身 MCP、真实工具快照与 Pole Agent A2A Card。
  - Pole Agent 发布公开 Card 和受认证的 JSON-RPC `message/send`；Console Agent 按 Registry 自然键解析并消费自身 MCP。
  - System Settings 保存管理员 desired revision 后自动探测和应用健康 Prompt/模型/工具策略，失败时保留 rejected 草稿和 last-known-good。
  - Registry 写入使用稳定 ID、revision、软删除复活和子项 upsert，重复 reconcile 保持幂等；配置主体与系统执行主体分别审计。

## [2026-07-26] refactor | 治理范围收敛为 RPC-first

- 新增页面：adr-rpc-first-governance-scope。
- 撤下页面：adr-multi-resource-governance-platform。
- 更新页面：terminology、domain-models、governance-rules、todo、lessons、index。
- 变更摘要：
  - 核心治理明确聚焦 HTTP、gRPC、Dubbo 服务调用及其路由、限流、鉴权、镜像、Mock、熔断和 A/B Test。
  - Kafka、RocketMQ、Redis、MySQL 暂不进入运行时灰度和请求级治理，只保留资源目录、健康、指标与原生管理集成。
  - 不向 specification 预埋多资源域万能策略；Capability Profile、原子 Bundle 和 Apply Receipt 只围绕真实服务数据面深化。
  - Job 只有在 Pole 拥有统一 Worker/Agent、租约和执行协议后再独立评估。

## [2026-07-26] feat | 四协议服务契约上报与可视化闭环

- 新增页面：adr-service-contract-reporting-and-visualization。
- 更新页面：service-discovery、api-servers、storage、cache-layer、todo、index。
- 变更摘要：
  - 统一 HTTP/OpenAPI、gRPC、Dubbo、Thrift 的 SDK gRPC 与 Agent/CI HTTP 推送模型。
  - OpenAPI 3.x 由服务端抽取接口，其余 RPC 协议由构建侧提交结构化接口；不让控制面主动扫描生产服务。
  - Client 与 Manual 按来源独立全量替换，修复缓存 miss、软删除回流、字段兼容和统一 Discover 分发。
  - 服务详情增加“服务契约”页签，展示四协议版本、接口来源和原始契约。

## [2026-07-26] refactor | Envoy xDS v3 对齐统一治理发布

- 更新页面：api-servers、governance-rules、todo、index。
- 变更摘要：
  - Envoy Node 通过显式治理 metadata 标签复用 normal/gray release 选择，RDS、VHDS、CDS 改为 node-scoped 稳定快照，EDS 继续按 namespace 共享。
  - 路由转换改用顶层 caller → callee 服务范围，保留 DestinationGroup 标签、权重和多目标，不再从目标组反推服务身份。
  - Gateway 接入真实节点策略生成链；LDS 与策略资源先完整构建再原子替换，相同 namespace 与治理标签的节点复用规则选择结果。
  - OR、动态请求参数、非基础 QPS 限流字段及无等价 Envoy 执行模型的能力不再静默降级，并在知识库固化当前支持矩阵。

## [2026-07-27] refactor | Dependabot 安全告警清理

- 更新页面：todo、lessons、index。
- 变更摘要：
  - 清理 Go 与 Console 的 Dependabot 告警依赖，移除无安全修复版本的前端 Mock 与路由依赖。
  - Go 漏洞扫描无受影响符号或导入包，Console npm 审计无开放漏洞。
  - 增加面向 `develop` 的 Go 与 Console npm Dependabot 周更配置，并将远端开放告警归零作为发布门禁。

## [2026-07-27] ingest | Dependabot 修复版本更新至本地 Kubernetes

- 更新页面：todo、index。
- 变更摘要：
  - 为 `develop@69be9be7` 构建 Linux ARM64 镜像 `pole-control-plane:local-20260727-dependabot-69be9be7`，滚动更新 OrbStack `pole-system` Control Plane。
  - 通过 Pod imageID、Pod 内静态资源、Gateway 返回哈希和 HTTPRoute 状态确认实际入口已切换到新容器产物。
  - Pod 内 Console、8090 Control Plane 与 functions 健康接口均返回 200；真实浏览器在 Gateway 打开 `demo-governance/demo-order` 契约 Tab，四协议真实数据可见且无浏览器错误。
  - Deployment 补齐 source revision、image ID、change-cause 和部署时间注解；Dependabot 开放告警保持为 0。

## [2026-07-27] feat | 服务契约详情与 Dubbo Metadata 适配

- 更新页面：terminology、domain-models、service-discovery、adr-service-contract-reporting-and-visualization、todo、lessons、index。
- 变更摘要：
  - 修复接口清单只有静态摘要、无法进入详情的问题，新增可点击和键盘访问的接口详情抽屉。
  - 详情展示协议字段、方法签名、来源、修订、接口定义和全部 metadata；Dubbo 增加 Metadata Center 专属投影。
  - 统一上报入口可直接识别 Dubbo provider/consumer 运维定义和 application revision 快照，保留原始 JSON 并自动抽取接口。
  - Dubbo 原始 metadata 键不丢失，同时生成稳定的 `dubbo.*` 查询键和 Service Key。
  - 运行时 HTTP client 上报路径明确为 `/v1/ReportServiceContract`；`/naming/v1` 只承载 Console 管理 API。
  - 运行时代码提交 `015cbdbd` 构建并更新到本地 Kubernetes；真实 Dubbo 原生样例上报、鼠标/键盘详情交互与 Metadata Center 投影均通过浏览器验证。

## [2026-07-28] ingest | 配置模板与 Namespace Value 客户端渲染方案

- 新增页面：adr-config-template-client-rendering。
- 更新页面：config-center、todo、lessons、index。
- 变更摘要：
  - 模板与 Value 分别版本化，Value 按 `Namespace + Template` 建立聚合并支持正式、灰度发布。
  - 服务端根据客户端标签选择唯一命中的 Value Release，SDK 不接收或解释灰度规则。
  - SDK 获取模板与 Value 的原子组合快照，本地缓存并确定性渲染；失败时保留 last-known-good。
  - 配置文件显式固定和切换模板发布版本，模板升级不会自动影响已有绑定。
  - 首期模板语法只支持类型化变量替换，不允许任意函数、脚本、环境变量或外部 I/O。

## [2026-07-28] refactor | 配置模板跨语言引擎与服务端预览职责

- 更新页面：adr-config-template-client-rendering、config-center、todo、lessons。
- 变更摘要：
  - 服务端渲染结果只用于预览、格式诊断和跨语言一致性参考，不作为 SDK 运行时权威配置。
  - SDK 继续对服务端选中的 Template Release 与 Value Release 执行本地渲染，并校验服务端参考哈希。
  - 不采用完整 Go template，定义基于 Mustache 1.3 core 的 `pole-mustache-v1` 严格 Profile。
  - 首期只允许 triple-mustache dotted scalar 变量，禁用 section、partial、lambda、helper、动态 delimiter 和外部 I/O。
  - 服务端参考实现与所有语言 SDK 必须运行同一份语言无关测试向量。

## [2026-07-28] ingest | 配置模板 specification 与服务端参考实现

- 更新页面：adr-config-template-client-rendering、config-center、todo、lessons。
- 变更摘要：
  - specification 保留现有配置类型 wire 布局，新增模板/Value release、显式 binding、RenderSnapshot、预览 RPC 和客户端能力协商契约。
  - ConfigFile 只有纯文本与模板两种类型；SDK 的渲染结果不是第三种文件类型。
  - control-plane 增加模板与 Namespace Value 持久化、严格参考渲染、格式校验、发布时 binding 事务激活和回滚恢复。
  - Discover 从已经命中的普通或灰度 ConfigFileRelease 读取固化 binding，只向 SDK 返回匹配后的 Value 快照和参考哈希。
  - 预览响应通过统一 `code/info` 表达鉴权、参数和系统错误，渲染 diagnostics 不承担 API 错误。
  - SDK 本地渲染/last-known-good、组合 revision Watch 与 Console 管理入口仍待后续接入。

## [2026-07-28] feat | 配置模板 Console 管理入口

- 更新页面：adr-config-template-client-rendering、config-center、todo。
- 变更摘要：
  - Console 增加模板草稿、参数 Schema、Template Release、Namespace Value 正式/灰度发布与参考预览工作台。
  - ConfigFile 详情显式选择普通文本或模板渲染，并固定明确的 Template Release；发布详情展示固化 binding。
  - 新增模板、Value 与 binding 管理查询接口，前端可在刷新后恢复草稿和发布历史。
  - 模板 binding 与文件草稿字段在同一服务端事务内更新，失败不会留下不完整模板草稿。
  - 模板发布只从已保存的模板草稿生成不可变快照；不支持的预览引擎使用统一参数错误码。

## [2026-07-28] ingest | 统一进程模式与 Pole Limiter 集成方案

- 新增页面：adr-unified-process-mode-and-limiter-integration。
- 更新页面：architecture、configuration、index、todo。
- 变更摘要：
  - 决定以同一源码、二进制和镜像提供 Console、Control Plane、Limiter 与 all Profile，生产仍保持独立 workload。
  - 定义可嵌入 Module、Running 与 Supervisor seam，以及 readiness、失败回滚和逆序优雅停机不变量。
  - 明确 Limiter 是有状态数据面，`all` 只适用于 quickstart、演示和轻量部署。
  - 记录现有 `all=control-plane+console` 的破坏性兼容问题，并给出跨 breaking release 的迁移方案。
  - 记录 Limiter node-id、advertised endpoint、长流、内存状态以及 8100/8101 网络安全约束。

## [2026-07-28] ingest | 逻辑服务与环境服务显式关联

- 新增页面：adr-logical-service-environment-binding。
- 更新页面：terminology、domain-models、business-rules、namespace、service-discovery、index、todo。
- 变更摘要：
  - 不再使用服务名作为跨环境权威身份；环境服务继续以 `namespace + runtimeServiceName` 运行。
  - 引入仅属于控制面的 Logical Service 稳定 ID，SDK、注册发现和数据面治理协议均不感知。
  - 不同 Namespace 下的环境服务由管理员在 Console 显式关联，控制面只提供候选建议，不自动强绑定。

## [2026-07-28] ingest | 业务环境与 Pole 系统空间类型化

- 新增页面：adr-system-namespace-kind。
- 更新页面：namespace、terminology、domain-models、business-rules、index、todo、lessons。
- 变更摘要：
  - Namespace 增加 `BUSINESS | SYSTEM` 一等类型，旧 payload 缺省值保持为 BUSINESS。
  - `pole-system` 定义为当前 Pole 安装实例的内部管理面空间，不再解释为普通业务环境。
  - 公共 API 禁止创建或删除系统空间，MySQL 启动迁移幂等回填系统类型。
  - Logical Service、配置分组与配置文件的跨环境聚合排除系统空间，显式系统管理能力仍保留。
  - Console 将业务环境和“Pole 系统空间（当前控制面）”分区展示。
  - 未关联的环境服务仍可正常注册、发现和治理，但不参与跨环境聚合。

## [2026-07-28] feat | 统一进程模式与 Pole Limiter 集成

- 更新页面：adr-unified-process-mode-and-limiter-integration、architecture、configuration、todo。
- 变更摘要：
  - 将 Limiter 核心、协议服务、注册与统计实现纳入统一源码和二进制。
  - 增加 `control-plane`、`limiter-server`、`full` Profile，保留 `server` 别名和原 `all` 兼容语义。
  - 建立统一 Supervisor，提供同步 listener readiness、部分启动回滚、运行期 fail-fast 和逆序优雅停机。
  - 示例默认只开放内部 gRPC `8101`，HTTP 运维端口保持关闭；生产仍推荐同镜像、分 workload 部署。

## [2026-07-29] ingest | Console、Limiter 源码归属与前端嵌入制品

- 新增页面：adr-console-limiter-source-layout-and-embedded-web。
- 更新页面：architecture、adr-unified-process-mode-and-limiter-integration、index、todo、lessons。
- 变更摘要：
  - Console Go 网关与 Limiter 统一归入 `pkg/console`、`pkg/limiter`，内部实现由局部 `internal` 保护。
  - Console 前端源码迁入 `web/console`，不建立独立版本或发布生命周期。
  - Release/Test 由统一构建流程生成前端产物并复制到 Go 包生成目录，通过 `go:embed` 进入单文件二进制。
  - 本地开发保留 Vite dev server 与 HMR，正式制品不再依赖外部 `webPath`。
  - bootstrap 只依赖两个 Module 的根 Interface；共享 A2A/Agent 类型上提到稳定契约层。

## 相关页面

- [[index]]
- [[schema]]
