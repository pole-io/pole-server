---
title: 操作日志
tags: [meta, log]
links: [index, schema]
updated: 2026-06-10
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

## 相关页面

- [[index]]
- [[schema]]
