---
title: ADR：全局环境晋升拓扑与跨资源 Bundle
tags: [adr, namespace, environment, promotion, lane, release, console]
links: [namespace, terminology, domain-models, config-center, governance-rules, ai-features, adr-system-namespace-kind, adr-config-template-client-rendering, adr-governance-rule-unified-storage, adr-ai-resource-environment-binding]
updated: 2026-08-10
sources: 9
---

# ADR：全局环境晋升拓扑与跨资源 Bundle

## 状态

Accepted；Phase 1 已部分实施。

当前已落地拓扑草稿、DAG/lane binding 校验、不可变 Revision、MySQL 存储、Console API 与拓扑编辑页；
配置模板定义草稿已按 Namespace 隔离，并按“当前环境版本 → 当前环境最新版本 → 旧全局草稿”顺序兼容初始化。
Promotion Adapter、ChangeSet/Bundle、审批门禁与晋升工作台仍属于后续实现，不应把本 ADR 的完整状态机视为已交付。

## 背景

Pole 已将 `BUSINESS` Namespace 作为配置、服务和治理资源的运行环境边界，但跨环境流转仍是各资源域的局部行为。如果模板定义草稿全局共享，dev 中未发布的内容或 Schema 修改也会改变 pre/pro 的预览和下一次发布基线。

用户的实际环境链不一定是固定的 dev/test/pre/prod，可能包含多个开发泳道、hotfix、多条测试路径与不同的审批门禁。因此环境关系不能由配置中心写死，也不能将运行时事实与期望状态一起复制。

## 决策摘要

1. 在 Namespace 域提供一张全局、用户可定义的环境晋升 DAG，不写死环境名称、阶段数量或组织流程。
2. 拓扑只连接 baseline 环境；lane 作为某个 baseline 的隔离开发分支，通过独立 `LaneBaseBinding` 回归 baseline，不与“base 初始化 lane”关系一起塞入 DAG 形成循环。
3. 晋升单位是带来源版本与分叉基线的资源变更集，可逐项选择或全选；不覆盖整个目标环境。
4. 所有目标候选预检、审批和验证后冻结为不可拆分的 `EnvironmentPromotionBundle`。发布保证控制面期望状态原子，数据面通过回执最终收敛。
5. 只有已发布、不可变的期望状态可晋升。实例、健康、指标、自动发现契约和环境密钥等运行时/专属数据永不复制。
6. 每个资源域必须实现统一 `PromotionAdapter` 后才能进入晋升流程。

## 环境和 lane 模型

Namespace Kind 继续只区分 `BUSINESS` 与 `SYSTEM`，不增加写死的 DEV/TEST/PRE/PROD 枚举。`SYSTEM` 永不进入业务晋升拓扑。

BUSINESS Namespace 在晋升模型中可以是：

- `BASELINE`：可作为晋升 DAG 节点的稳定环境。名称完全由用户定义，例如 dev-base、tst、pre、pro。
- `LANE`：绑定唯一 baseline 的隔离开发空间。lane 不作为下游阶段的晋升来源，其变更必须先回归 baseline。

```text
lane-a ─┐
lane-b ─┼─ MERGE_TO_BASE → dev-base ─ PROMOTE → tst ─ PROMOTE → pre ─ PROMOTE → pro
hotfix ─┘
```

`LaneBaseBinding(lane_namespace, base_namespace)` 是拓扑外的显式关系：

- 一个 lane 在一个拓扑 Revision 中只能绑定一个 base。
- lane 初始化或 rebase 只复制所选资源的正式基线版本，不形成“base → lane”晋升边。
- 每个 lane 资源记录自身的 `fork_baseline_version`，不用一个粗粒度环境时间戳代替。
- lane 改绑 base 是高风险操作；存在未合并变更时阻断，必须先回归、显式丢弃或导出变更集。

## 晋升拓扑与边策略

全局仅有一张活跃拓扑，可含多个互不相连的子图。拓扑使用草稿和不可变 `TopologyRevision`：

```text
TopologyDraft --validate/publish--> TopologyRevision N (active)
                                      └─ supersedes Revision N-1
```

发布拓扑草稿前必须校验：

- 所有节点是存在且可用的 BUSINESS Namespace。
- baseline 晋升边构成 DAG，增加任意循环时拒绝发布。
- lane 不是 baseline DAG 节点，且每个 lane 最多有一个有效 base 绑定。
- 边策略引用的角色、外部门禁和资源域存在。
- 被进行中 Bundle 引用的 Namespace 删除或类型变更必须阻断。

每条 baseline 边携带独立策略：

```text
PromotionEdgePolicy {
  source_release = FORMAL_ONLY
  allowed_resource_domains[]
  approval_roles[] + approval_quorum
  allow_creator_self_approval
  validation_gates[]
  candidate_ttl
  publish_permission
  rollback_policy
}
```

安全默认值：正式发布版本才可晋升；发起人不能自审；目标是上游或无环关系的新边不自动生效；候选超时后必须重新预检。非生产边可由管理员显式放宽审批数和门禁，但不能放宽身份、权限、DAG 和不可变来源版本约束。

晋升变更集创建时固定引用拓扑 Revision 和边策略快照。新拓扑只影响新变更集；进行中单据显示“使用历史拓扑”，需要紧急停止时由有权用户显式撤销并记录原因。

## 资源变更集与 lane 三方合并

变更集以资源为单位，不是整个 Namespace 快照。每个项目至少固定：

```text
PromotionItem {
  resource_domain
  logical_resource_id
  source_namespace
  source_release_id
  target_namespace
  fork_baseline_release_id?  // lane merge only
  target_baseline_release_id?
  preflight_action           // CREATE | UPDATE
  content_digest
}
```

lane 回归 base 时执行三方比较：

```text
fork baseline + lane source + current base = base candidate
```

- 仅 lane 改变：自动合并。
- 仅 base 改变：保留 base。
- 双方相同改变：合并一份。
- 同一语义位置不同改变：标记 `CONFLICT`，不静默选边。
- 不支持结构合并的二进制/不透明资源：发生双边改变时必须人工选择完整版本。

用户可逐项选择或全选 lane 已发布变更。未选项保留在 lane，不会因其它项已合并而丢失分叉基线。项目成功进入 base 新正式版本后，它的 lane fork baseline 才前移；未合并项继续使用原基线。

## Promotion Adapter 契约

编排器只依赖稳定深模块：

```go
type PromotionAdapter interface {
    Domain() string
    Export(ctx context.Context, ref SourceVersionRef) (ImmutableArtifact, error)
    Preflight(ctx context.Context, req PreflightRequest) (PreflightResult, error)
    Merge(ctx context.Context, req MergeRequest) (CandidateArtifact, []Conflict, error)
    Validate(ctx context.Context, candidate CandidateArtifact) ([]Diagnostic, error)
    Prepare(ctx context.Context, candidate CandidateArtifact) (PreparedTarget, error)
    Commit(ctx context.Context, tx store.Tx, prepared PreparedTarget) (TargetReleaseRef, error)
}
```

Adapter 使用稳定 `logical_resource_id` 匹配目标，不模糊匹配名称。预检结果是：

- `UPDATE`：目标已有同一逻辑资源。
- `CREATE`：目标缺失且发起人有创建权，将生成明确的新增候选。
- `CONFLICT`：同名或唯一键被其它逻辑资源占用，必须解决。
- `UNSUPPORTED`：Adapter、目标环境或版本契约不支持。

Adapter 必须是幂等的，`Prepare` 不得改变目标运行时状态，`Commit` 必须在控制面事务内只创建不可变目标版本和 Outbox 事件。动态插件不能绕过资源域权限、审计和事务边界。

## 可晋升与不可晋升资源

可晋升资源必须是用户声明、已正式发布、可重现的期望状态：

- 配置文件发布版本。
- 配置模板的环境 Template Snapshot（与目标环境 Value 重新绑定）。
- 治理规则发布版本和 Service Policy Bundle。
- 已建立不可变版本契约的 MCP/A2A 逻辑定义或部署描述。

明确排除：

- 草稿、灰度命中结果与未完成正式发布的候选。
- 服务实例、注册地址、健康、隔离、订阅者和流量状态。
- 监控、Trace、日志和自动发现的服务契约。
- Secret、凭据、模板 Value、Endpoint 和其它环境专属值。

“全局”表示共用拓扑、变更集、门禁和审计编排，不表示复制 Namespace 中的一切。

## 目标环境专属值

晋升永不复制来源环境 Secret、Value 或 Endpoint。Adapter 在预检时加载目标环境现有绑定：

- 已存在且与新 Schema 兼容的值自动复用。
- 新增必填值必须由目标环境有权用户补齐。
- 类型变化立即校验；不合法值阻断预检。
- 字段重命名不自动猜测，用户显式映射或重新填写。
- 已删除字段不进入新 Value Snapshot。
- 敏感值只显示已配置/未配置，不返回明文，不出现在差异、日志或审计摘要。

目标值可在晋升预检页就地补齐，也可跳转目标环境资源编辑器。任何候选内容或目标绑定变化都使已有审批失效，必须重新冻结、预览和审批。

## Bundle 生命周期与原子边界

```text
DRAFT
  -> PREFLIGHTING
  -> NEEDS_INPUT | CONFLICTED | READY_FOR_APPROVAL
  -> AWAITING_APPROVAL
  -> APPROVED
  -> PUBLISHING
  -> SUCCEEDED | DEGRADED | FAILED

Any non-terminal state -> CANCELED
Changed candidate      -> DRAFT (new change-set revision, approvals invalidated)
```

- `DRAFT`：选择资源，固定来源版本、目标基线、拓扑 Revision 和边策略。
- `PREFLIGHTING`：所有 Adapter 执行权限、目标对齐、差异、冲突、专属值和格式校验。
- `NEEDS_INPUT/CONFLICTED`：仅保存候选和诊断，不发布。
- `READY_FOR_APPROVAL`：全部候选准备成功并冻结内容哈希。
- `AWAITING_APPROVAL/APPROVED`：执行边策略的角色、人数、外部门禁和过期时间。
- `PUBLISHING`：各 Adapter `Prepare` 全部成功后，在一次控制面事务内创建不可变目标版本、Bundle 和 Outbox。
- `SUCCEEDED`：所有要求回执收敛；`DEGRADED`：期望状态已提交但部分执行点未收敛；`FAILED` 只表示提交前失败或事务未提交。

Bundle 对用户是一次不可拆分发布。数据面多执行点无法保证物理同时切换，因此 Pole 只声明“控制面期望状态原子 + 可观测最终收敛”。不得在 UI 或 API 中将 `PUBLISHING/DEGRADED` 表述为全部实例已同时生效。

## 审批、外部门禁和权限

发起晋升需同时满足：

- 来源 Namespace 和资源版本读权限。
- 目标 Namespace 对应资源的创建/更新候选权限。
- 目标边的 `promote` 权限。

审批是独立权限，不因具有资源写权而自动获得。边策略可指定角色、人数、是否允许发起人自审和审批过期时间。安全默认禁止自审。

外部验证门禁使用幂等、带签名与有效期的回执，必须固定 Bundle 候选哈希。候选改变、回执过期、门禁配置不可用或目标基线漂移都使门禁失效。

## 并发、幂等和基线漂移

- Change Set 和 Bundle 使用不可变内容哈希与客户端幂等键；重试不创建重复目标版本。
- 预检固定每个目标资源基线 release/revision。在审批前发生漂移时重新预检；在审批后漂移时阻断发布并使审批失效。
- 同一目标逻辑资源同时只允许一个 Bundle 进入 `PUBLISHING`；其它 Bundle 等待或因基线变化重新预检。
- 发布事务使用 Outbox；不在数据库事务内同步调用外部系统或执行点。

## 灰度、下游晋升与回退

- 灰度是目标资源 Adapter 的发布策略，不改变 Bundle 内固定的目标制品。
- 仍在灰度、`PUBLISHING` 或 `DEGRADED` 的 Bundle 不能作为下一条边的晋升来源。只有完成正式发布且符合边策略的不可变版本可继续向下游晋升。
- 回退是目标 Namespace 内的恢复操作，不需要拓扑中存在反向边。它根据原 Bundle 和目标当前状态创建新的 rollback Bundle，重新走回退策略与必要审批，不重新激活或改写历史版本。
- 已被下游晋升的上游版本可回退，但 UI 必须提示下游已存在派生版本；不自动反向传播。

## 存储模型

建议使用显式表，不将拓扑和流程数据编码到 `namespace.metadata`：

| 存储 | 主要身份 | 责任 |
|---|---|---|
| `environment_topology_draft` | singleton | 当前全局拓扑草稿与 optimistic revision |
| `environment_topology_revision` | revision_id | 不可变拓扑版本和审计信息 |
| `environment_topology_edge_revision` | revision_id + edge_id | 来源、目标和边策略快照 |
| `environment_lane_binding_revision` | revision_id + lane_namespace | lane 到唯一 base 的绑定 |
| `environment_promotion_change_set` | change_set_id + revision | 变更集状态、拓扑/策略快照与候选哈希 |
| `environment_promotion_item` | change_set_id + item_id | 资源版本、分叉/目标基线、预检、诊断与候选引用 |
| `environment_promotion_approval` | change_set_id + approver | 审批决定、角色、候选哈希和时间 |
| `environment_promotion_gate_receipt` | change_set_id + gate_id | 外部验证回执、签名、过期时间 |
| `environment_promotion_bundle` | bundle_id | 不可拆分发布、期望状态与总体收敛状态 |
| `environment_promotion_bundle_item` | bundle_id + item_id | 目标不可变资源版本 |
| `environment_promotion_apply_receipt` | bundle_id + target | 执行点/资源回执与重试状态 |
| `environment_promotion_outbox` | event_id | 发布事务后可靠分发 |

不对旧 Namespace 表执行破坏性重建。拓扑新表为空时表示“尚未配置晋升关系”，不根据 Namespace 名称自动猜测 dev/test/prod 关系。

## API 边界

建议管理 API：

```text
GET  /environment/v1/promotion-topology/draft
PUT  /environment/v1/promotion-topology/draft
POST /environment/v1/promotion-topology/validate
POST /environment/v1/promotion-topology/publish
GET  /environment/v1/promotion-topology/revisions

POST /environment/v1/promotion-change-sets
PUT  /environment/v1/promotion-change-sets/{id}/items
POST /environment/v1/promotion-change-sets/{id}/preflight
POST /environment/v1/promotion-change-sets/{id}/submit
POST /environment/v1/promotion-change-sets/{id}/approve
POST /environment/v1/promotion-change-sets/{id}/reject
POST /environment/v1/promotion-change-sets/{id}/cancel
POST /environment/v1/promotion-change-sets/{id}/publish

GET  /environment/v1/promotion-bundles/{id}
GET  /environment/v1/promotion-bundles/{id}/receipts
POST /environment/v1/promotion-bundles/{id}/rollback
```

所有写 API 必须接受 idempotency key 和 expected revision/content hash。预检、审批和发布返回类型化诊断，不使用单一字符串混合权限、冲突、门禁和资源格式错误。

## Console 信息架构

环境空间增加两个全局任务：

1. **晋升拓扑**：编辑 baseline DAG、lane/base 绑定和边策略；支持图形与可访问表格两种视图、无环校验、变更影响预览、草稿发布和历史 Revision。
2. **晋升工作台**：发起变更集、资源差异/冲突、目标专属值、审批、门禁、发布与回执。

资源选择先选来源/目标边，再按 Adapter 分组展示可晋升的已发布版本。预检表必须显示 CREATE/UPDATE/CONFLICT/UNSUPPORTED、来源版本、目标基线、差异和环境专属值状态。

线性创建变更集可使用步骤式导航，因为它确实有依赖的顺序和完成态；不应再用并列 Tab 复制同一流程。查看历史 Bundle 使用页签/详情模式，不显示可编辑流程条。

每个 Namespace 详情显示其上游、下游、lane/base 关系、待处理晋升和最近 Bundle，但不在每个资源页复制全局拓扑编辑器。

## 配置模板环境化

配置模板必须拆成：

```text
ConfigTemplateIdentity (global)
  -> id, name, description, labels

NamespaceConfigTemplateDraft (namespace + template_id)
  -> content, format, parameter_schema, engine
  -> target-environment Value draft

EnvironmentConfigRelease (namespace + template_id + version)
  -> TemplateSnapshot + ValueSnapshot
```

全局身份只用于跨环境对齐，不包含可执行模板草稿。内容、格式、Schema 和引擎都是 Namespace 工作副本。从上游晋升模板时，导入来源 TemplateSnapshot 作为目标候选，再与目标 Value 生成新的原子 EnvironmentConfigRelease。

旧全局草稿迁移规则：

1. 目标 Namespace 已有 active EnvironmentConfigRelease：从该环境实际生效的 TemplateSnapshot 初始化 Namespace 草稿，不使用当前全局草稿覆盖。
2. 没有 active release，但存在 Value、binding 或历史版本：使用最近一个本环境 TemplateSnapshot；没有快照时才使用旧全局草稿初始化一份未发布环境草稿。
3. 与该模板没有任何关系的 Namespace 不自动创建副本；首次使用通过晋升、显式初始化或创建候选完成。
4. 旧全局草稿字段保留为迁移证据和旧版读取兼容，新 Console/API 不再写入。已环境化的模板收到旧全局更新请求时返回明确升级错误，不将修改广播到所有环境。

详细渲染和 Template + Value 原子版本契约继续由 [[adr-config-template-client-rendering]] 定义。

## 实施分期

### Phase 1：环境核心与配置域

- 拓扑草稿/Revision、DAG 校验、lane/base 绑定、边策略与 Console 拓扑编辑器。
- Change Set/Bundle 编排、审批、撤销、审计、Outbox 与回执状态框架。
- 配置模板环境草稿迁移和 Promotion Adapter。
- 配置文件 Promotion Adapter。

### Phase 2：治理域

- 治理规则和 Service Policy Bundle Adapter。
- 执行点应用回执与 Bundle 收敛展示。

### Phase 3：AI 声明式资源

- 先为 MCP/A2A 定义和部署描述补齐不可变版本契约，再实现 Adapter。
- Endpoint、Secret 和实时 Tool/Card 发现结果继续留在目标环境，不晋升。

## 兼容与发布策略

- 新表为空时不改变任何现有资源发布路径；用户必须显式配置并发布第一个拓扑 Revision。
- 首期保留各资源域原生发布 API，Promotion Bundle 作为新的编排入口；Adapter 复用原子存储能力，不通过内部 HTTP 自调用。
- 不对 Namespace 名称自动猜测拓扑，不自动创建 lane/base 绑定，不自动推送任何存量版本。
- 旧客户端不感知晋升编排，继续消费目标资源域的普通发布版本。
- 数据库迁移必须幂等，支持先部署新表/读路径，再切换 Console 写入，最后禁用旧全局模板写入。

## 被否决的方案

### 写死 dev/test/pre/prod 阶段

无法表达 lane、hotfix、多测试环境和组织自定义门禁。

### 将拓扑存入 Namespace metadata

缺少引用完整性、DAG 校验、版本、查询、审计和权限边界。

### 将 lane 初始化和回归都建成 DAG 边

`base -> lane` 与 `lane -> base` 必然形成循环。lane/base 是分叉绑定，baseline 之间才是晋升 DAG。

### 复制整个 Namespace 快照

会覆盖其它 lane 已回归变更，并夹带实例、密钥、Endpoint 和监控数据。

### 将来源 Value/Secret 一起晋升

破坏环境隔离，也可能把 dev 凭据和地址带入 pre/pro。

### 候选自动在目标环境生效

绕过目标值、权限、差异、冲突、审批和外部门禁。

### 声明多数据面执行点瞬时原子

运行时无法提供该保证。正确契约是原子提交期望状态并显式跟踪最终收敛。

## 验收标准

- 用户可创建包含分支与汇合的 baseline DAG，循环在草稿发布前被拒绝。
- lane 绑定唯一 base，按资源分叉基线三方合并，不覆盖 base 其它变更。
- 进行中变更集固定拓扑 Revision、边策略、来源版本和目标基线。
- 只有实现 Adapter 的已发布期望状态资源可选，运行时事实和环境专属值不出现在变更集。
- 目标缺失时生成明确 CREATE 候选；同名异 ID、权限不足、基线漂移和未解决冲突阻断发布。
- 全部候选预检成功后才能冻结 Bundle；审批后内容变化使审批失效。
- Bundle 不可拆分发布，控制面期望状态原子写入，执行点回执能区分 PUBLISHING、SUCCEEDED 和 DEGRADED。
- 模板内容/Schema 草稿按 Namespace 隔离，dev 编辑不改变 pre/pro 草稿、预览或发布基线。
- 目标 Value 不从来源复制；新必填、类型变化和敏感值状态在预检中明确处理。
- 回退新建 rollback Bundle，不需要反向拓扑边，不篡改历史版本。

## 相关页面

- [[namespace]]
- [[terminology]]
- [[domain-models]]
- [[config-center]]
- [[governance-rules]]
- [[ai-features]]
- [[adr-system-namespace-kind]]
- [[adr-config-template-client-rendering]]
- [[adr-governance-rule-unified-storage]]
- [[adr-ai-resource-environment-binding]]
