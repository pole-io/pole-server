---
title: ADR：多源 Skill Marketplace、不可变 Bundle 与 CLI 安装边界
tags: [adr, ai, skill, marketplace, registry, supply-chain, cli]
links: [skill-marketplace, ai-features, domain-models, storage, auth-system, api-servers, adr-console-agent-resource-workbench]
updated: 2026-08-11
sources: 0
---

# ADR：多源 Skill Marketplace、不可变 Bundle 与 CLI 安装边界

## 状态

Accepted，首期纵向切片已实现；统一 RBAC 资源类型、真实外部 Registry 与生产 MySQL 仍待集成验收。

## 背景

Pole 已有 MCP Server/Tool Registry、A2A Agent Registry 和 Console Pole Agent，但没有独立、可版本化、
可在团队与外部生态间分发的 Skill 资源。A2A Agent Card 中的 `skills` 是 Agent 能力投影，
没有 Bundle、SemVer、签名、审核或安装语义；Pole 内置 Go Plugin Registry 也仅管理编译期官方实现。

## 决策

### 1. Skill 是独立聚合根

- `Publisher` 是持久发布身份，拥有稳定 slug、成员授权、信任状态和版本化签名公钥。
- `Skill` 以 `publisher/name` 全局定位，是可发布逻辑定义，不绑定 Pole Namespace 或 Agent。
- `SkillRelease` 以严格 SemVer 和 Bundle SHA-256 定位；发布后不可变，同版本不得换包。
- `ReleaseReview` 保存候选版本的扫描证据、审核人、决策和原因。
- `RegistrySource` 表达 Pole、Git Tag/Release 或通用 HTTP Index 来源，带信任等级、同步水位和错误状态。

### 2. 遵循开放 Bundle 格式

Bundle 原样遵循 Agent Skills 格式，`SKILL.md` 只承载开放格式元数据与 Agent 指令。Pole 的
Publisher、SemVer、审核、摘要、签名和来源证据保存在 Marketplace 聚合中，不将 Pole 私有必填字段
注入 `SKILL.md`。

Bundle 必须在入库前通过统一安全验证：

- 压缩后不超过 16 MiB，解包后不超过 64 MiB，不超过 2,048 项；
- 拒绝绝对路径、路径穿越、符号链接、设备文件、异常压缩比和缺少根 `SKILL.md`；
- 生成规范文件清单、字节数、媒体类型与整体 SHA-256；
- 浏览器可展示安全文本与文件元数据，不执行脚本，不预览二进制。

### 3. BundleStore 是存储 seam

Marketplace 模块只依赖小型 `BundleStore` interface：按内容摘要 `Put/Get`，并按保留时间删除无引用
对象。interface 同时定义内容寻址、幂等写入、完整性校验与最大对象大小。

默认 Adapter 为 MySQL BLOB：先完整写入并校验 Bundle Blob，再在同一业务流中创建 Release 引用。
不允许没有完整 Blob 的 Release 进入审核或发布状态。无 Release 引用的 Blob 保留七天后回收。
未来 S3 或文件系统 Adapter 必须经过相同 interface 契约测试，不得让 HTTP 和 Store 调用方感知后端类型。

### 4. 发布、签名与审核

- 发布签名初期使用 Ed25519 detached signature，签名载荷固定为
  `publisher/name + version + canonical bundle digest + signedAt`；multipart 使用包含算法、时间和
  base64 签名的 JSON envelope 文件。
- Publisher 公钥按 key ID 版本化，保存激活、撤销与有效期；Release 保留验签时的 key ID 和证据。
- 私有 Release 扫描通过即可发布。公开 Release 默认进入持久审核流；受信 Publisher
  才能配置扫描通过后自动发布。
- 已发布版本不物理删除；`yank` 使其退出新解析，`deprecate` 保留解析但显示迁移警告。
- 审核、撤回、授权、公钥变更和 Registry Source 变更接入 Pole 现有审计历史。

### 5. 多源联邦不实时透传

`RegistryAdapter` 是外部依赖 seam，统一输出规范目录项、精确版本、原始摘要和 Bundle 流。
首期包含 Pole 原生 Registry、Git Tag/Release 和通用 HTTP Index Adapter。

外部目录由持久定时同步运行记录拉取，不在用户查询路径实时代理。用户查看或管理员预取时，
Release 冻结到本地 BundleStore：

- 同一 `publisher/name@version` 在同一来源中返回不同摘要时进入 `digest_conflict`；
- 上游下架后保留已冻结版本，只更新来源状态；
- `trusted` Source 可在本地扫描通过后进入可见目录，`untrusted` Source 必须进入隔离/审核；
- 同步失败不删除上一个健康快照，并持久化错误、尝试次数和下次重试时间。

### 6. 授权与公共读

Specification 工作区已新增 `SkillResources=32` 和 `StrategyResources.skills`，但 control-plane 当前
依赖的已发布 specification 尚未包含该枚举。首期因此不冒用 MCP/A2A 类型，而以现有 credential、
Publisher Member 和独立 Skill grant 实现 User、UserGroup、Role 私有可见性；不可见 Skill 不进入数量、
列表、详情或版本 API。公开已发布目录与 Bundle 提供窄匿名只读路径，写操作、审核、私有元数据与
Registry Source 始终鉴权。通用 Pole RBAC 策略编辑器直接引用 Skill 要在 specification 发布升级后接入。

### 7. CLI 拥有安装生命周期

新仓库 `pole-ai-cli` 生成 `pole-ai` 二进制，提供
`skill search/show/install/upgrade/list/verify/remove`。CLI 先把经过摘要校验与安全解包的 Bundle
写入本地内容寻址 Store，lockfile 只记录精确版本和摘要，然后通过 Codex、Claude Code 或通用
目录 Adapter 原子投影。

CLI 可解析 SemVer 约束，但默认升级不跨主版本。安装失败回滚到旧投影；卸载只删除 CLI
拥有且摘要仍匹配的文件，不覆盖或删除用户的同名手工资产。

## 深模块与 seam

- `Marketplace`：对 HTTP、同步 Job 和审核 Console 提供统一 interface，内部拥有状态迁移、不可变、签名和可见性规则。
- `BundleStore`：隐藏 MySQL BLOB 与未来对象存储差异。
- `BundleInspector`：统一安全解包、Agent Skills 校验、文件清单与扫描结果。
- `RegistryAdapter`：隐藏 Pole、Git 与 HTTP Index 传输差异。
- `LocalSkillStore`（CLI）：隐藏下载、校验、原子替换、lockfile 和回滚。
- `AgentAdapter`（CLI）：只负责从已验证本地 Store 投影到目标 Agent，不访问 Registry。

## 实施分期

1. **首期纵向切片**：MySQL BundleStore、本地发布/审核/公共下载、手工 Registry 同步、Console 全流程和 CLI 三种 Adapter。
2. **联邦强化**：持久同步调度、Git 凭证 SecretReference、负增量/ETag、冲突处置和来源健康页。
3. **公共信任强化**：Sigstore Adapter、更强静态分析、漏洞告警、撤销传播和管理员策略。

## 当前实现证据与边界

- 后端已实现规范化 ZIP、MySQL `LONGBLOB` BundleStore、签名/扫描/审核、公开匿名读、私有 grant、
  Registry 持久同步状态、七天孤儿回收与分页 index；非管理员不能下载任何非 `published` Bundle。
- Git 首期只支持 GitHub Release ZIP asset；手工 Git 导入只允许 private，public 必须改走 detached
  signature upload。通用 HTTP Index 的 Bundle 必须与 index 同源，且禁止 userinfo 与跨源重定向。
- `pole-static-v1` 只提供内嵌私钥和原生可执行文件等基础持久化扫描证据，不等同于恶意软件引擎、
  SBOM、漏洞数据库或沙箱执行。
- CLI 已实现七个命令、三类投影 Adapter、精确 digest lock、默认不跨 major 升级，以及对内容、类型、
  权限、缺失项和额外资产的完整性验证。
- Console fixture 浏览器证据覆盖深链刷新、命令复制、安全文本/二进制预览、720px 暗色视口；真实
  MySQL、实际 GitHub/HTTP Registry、统一审计和生产部署尚未验证，不能由 fixture 证据外推。

## 不做的事

- 不把 A2A Agent Card Skill 改成可执行市场资源。
- 不把 Skill 安装到 Pole Namespace 或 Control Plane 文件系统。
- 不让 Control Plane 执行、沙箱运行或动态加载 Bundle 脚本。
- 不为市场新建 Organization/Workspace 域；团队协作先复用 UserGroup/Role 与资源策略。
- 不用浮动 `latest` 作为 lockfile 或历史详情的事实来源。

## 验收边界

- 后端测试需覆盖 Bundle 限额/路径攻击、内容去重、不可变、SemVer、签名、审核、匿名可见性和摘要冲突。
- CLI 在临时目录和模拟 HTTP Registry 中验证安装、升级、回滚、篡改校验和保护用户文件。
- Console 必须验证深链、版本切换/刷新、Bundle 文件树、审核、黑暗主题、窄视口和抽屉草稿保留。
- 静态测试、模拟 Registry、浏览器 Fixture 和真实外部 Registry/公网数据必须分别报告。

## 相关页面

- [[skill-marketplace]]
- [[ai-features]]
- [[domain-models]]
- [[storage]]
- [[auth-system]]
- [[api-servers]]
- [[adr-console-agent-resource-workbench]]
