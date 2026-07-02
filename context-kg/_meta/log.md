---
title: 操作日志
tags: [meta, log]
links: [index, schema]
updated: 2026-07-03
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

## 相关页面

- [[index]]
- [[schema]]

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
