---
title: 命名空间（`pkg/namespace/`）
tags: [business, feature, namespace]
links: [architecture, storage, service-discovery, adr-logical-service-environment-binding, adr-system-namespace-kind, adr-environment-promotion-topology]
updated: 2026-08-10
sources: 6
---

# 命名空间（`pkg/namespace/`）

命名空间是 Pole 最顶层的运行边界。`BUSINESS` Namespace 表示相互隔离的业务环境；
`SYSTEM` Namespace 表示当前 Pole 安装实例的内部管理面空间。服务、配置、治理规则等资源
都归属某个 Namespace，但只有业务环境参与跨环境聚合。完整决策见
[[adr-system-namespace-kind]]。

## 跨环境资源身份

- `namespace + runtimeServiceName` 定位一个环境服务；跨环境逻辑服务使用控制面稳定 ID，
  并通过显式环境关联连接不同名称的环境服务。
- 配置分组名是跨环境逻辑标识；`namespace + groupName` 定位该分组的一个环境实例。
- 配置文件以 `groupName + fileName` 作为跨环境逻辑标识；`namespace + groupName + fileName` 定位具体环境中的文件。
- 治理规则以 `ruleType + ruleName` 作为跨环境逻辑标识；`namespace + ruleType + ruleName` 定位具体环境中的规则。
- 同一逻辑资源的不同环境实例可以具有独立内容、实例、版本和发布状态，任何复制、提升或发布都必须显式作用于目标命名空间。

## 全局环境晋升拓扑

- 环境晋升是 Namespace 域的全局能力，不属于配置中心私有流程；配置、治理、服务、MCP 和 A2A 等可晋升资源共用同一拓扑。
- 用户可根据实际流程定义 `dev -> lane -> tst -> pre -> pro` 或带分支、汇合的其它有向路径；系统不写死环境名称和阶段数量。
- 拓扑必须是有向无环图（DAG），允许一个环境存在多个上游或下游，但拒绝任何形成循环的新边。
- 晋升边只表示允许的流转方向，不意味来源版本在目标环境自动生效；各资源域负责将经验证制品转换为目标环境候选变更，再执行目标环境校验和发布。
- 每条 baseline 晋升边同时定义来源版本要求、目标候选行为、审批规则、验证门禁、允许资源域和冲突策略；不同边可有不同安全等级。lane 回归由独立 Lane Base Binding 约束，不写成 DAG 反向边。
- 默认只允许来源环境的正式不可变发布版本进入晋升变更集；灰度命中和草稿不是完整制品，不得绕过正式发布直接晋升。

### 拓扑版本与进行中晋升

- 全局晋升拓扑自身使用草稿和不可变发布版本；发布前必须校验 DAG 无环、Namespace 可用性和边策略完整性。
- 每个晋升变更集创建时固定引用当时生效的拓扑 Revision 和边策略快照；后续修改或删除边只影响新请求，不篡改已创建单据的语义与审计证据。
- 使用历史拓扑的进行中晋升必须在 UI 显式标记；安全策略紧急变更不隐式终止旧流程，管理员通过显式撤销操作终止并记录原因。
- 拓扑 Revision 审计记录包含操作者、发布原因、前后差异、创建时间与发布时间。

### lane 回归 base

- lane 必须通过 Lane Base Binding 绑定唯一 base，并记录每个可晋升资源从哪个 base 已发布版本分叉；`lane -> base` 是回归合并操作，不是拓扑边；`base -> tst` 等边才是跨阶段晋升。
- 回归单位是用户选择的资源变更集，可逐项选择或全选，不直接覆盖整个 base 环境快照。
- 每项资源使用“分叉基线、lane 当前已发布版本、base 当前已发布版本”执行三方比较。无冲突变更进入 base 候选；冲突项必须显式解决。
- 回归不直接改变 base 已生效版本；只有 base 候选变更完成校验并发布后，其新的不可变版本才能沿下游边晋升。

### 可晋升资源边界

- 全局晋升只处理用户声明、可版本化的期望状态。配置文件正式版本、配置模板环境版本、治理规则/Bundle 和具备版本契约的 MCP/A2A 定义可由对应适配器接入。
- 服务实例、注册地址、健康状态、订阅者、流量、监控数据和自动发现的服务契约是运行时事实，不进入晋升变更集。
- 环境密钥、模板 Value、Endpoint 等环境专属值不能从来源环境直接复制；目标候选必须引用或要求目标环境自己的值。
- 每个资源域通过统一 Promotion Adapter 导出不可变来源版本、计算差异/三方合并、生成目标候选并校验；没有适配器或没有版本契约的资源不显示在选择器中。

### 来源与目标资源对齐

- Promotion Adapter 必须使用资源域稳定的逻辑资源 ID 匹配来源与目标实例，禁止用模糊名称推断对应关系。
- 预检返回 `UPDATE`（目标已有同一逻辑资源）、`CREATE`（目标不存在且可创建）、`CONFLICT`（同名已被其它逻辑资源占用）或 `UNSUPPORTED`（目标或资源类型不支持）。
- `CREATE` 必须校验目标环境创建权限，并在确认页显式标记“新增候选”；只生成目标草稿/候选，不直接发布。环境专属 Value、Endpoint 和 Secret 仍由目标环境补齐。
- `CONFLICT` 和 `UNSUPPORTED` 阻断对应资源进入变更集；用户可返回调整选择，不得静默跳过或覆盖目标资源。

### 跨资源发布 Bundle

- 一次晋升变更集中的所有 Adapter 必须先完成权限、冲突、格式和环境专属值预检；任一资源失败时整个变更集不能进入待审批/待发布。
- 全部候选冻结后形成不可拆分的 Environment Promotion Bundle。用户只能整体确认和发布，不能在审批后单独替换或发布其中一项；变更需回到候选阶段并重新验证。
- 发布使用单一控制面事务固定 Bundle、所有目标资源版本引用和 Outbox 事件，保证原子期望状态；不把多个 SDK/Sidecar/Gateway 的实际应用时间误表述为瞬时运行时原子。
- 执行点按资源返回应用回执；全部成功后 Bundle 为 `SUCCEEDED`，部分未收敛时为 `DEGRADED` 并自动重试。需要回退时创建反向的新 Bundle，不改写原 Bundle 或历史资源版本。

## 关键常量

```go
SystemNamespace     = "pole-system"
DefaultNamespace    = "default"
ProductionNamespace = "Production"
DefaultTTL          = 5
```

## `NamespaceOperateServer` 接口（`apis/pkg/types/...`）

- `CreateNamespace`、`UpdateNamespace`、`DeleteNamespace`
- `GetNamespace`、`GetNamespaces`（分页）
- 命名空间可见性管理（跨环境访问关系）

## 实现方式

`pkg/namespace/Server` 使用 singleflight 防止并发重复创建命名空间。

## 系统命名空间不变量

- `default` 是系统自带的默认业务命名空间，不允许删除。
- `pole-system` 的 Kind 固定为 `SYSTEM`，是 Pole 内部组件和系统资源使用的内部命名空间，
  不表示 dev/test/prod，也不允许删除。
- 公共创建接口只能创建 `BUSINESS`；旧 payload 未携带 Kind 时因 `BUSINESS=0` 保持兼容。
- 系统空间不进入逻辑服务、配置分组或配置文件的业务跨环境聚合，但仍可被显式访问和授权。
- 后端 `DeleteNamespace` 在创建事务前拒绝删除这两个命名空间，批量删除同样逐项生效；命名空间列表对两者返回 `deleteable=false`，不能仅依赖前端隐藏入口。
- Console 列表按名称再次禁用删除操作，分别提示默认命名空间和内部系统空间不可删除；`pole-system` 必须展示“内部系统空间”标识，避免用户将其误认为普通业务空间。

## 列表统计与滚动边界

- `GetNamespaces` 在查询当前页命名空间后，复用 Store 的 `CountConfigFileEachGroup` 一次性聚合配置文件；按命名空间汇总各分组数量后通过 `Namespace.total_config_file_count` 返回。该数量是配置文件数，不是配置分组数，也不会为每一行额外查询数据库。
- Console 将当前页配置文件总数与每行“配置文件”列并列展示，服务数、配置文件数和实例健康信息都是命名空间的独立统计维度。
- 长列表页应固定页头、指标栏、工具栏与分页安全区，只有 `.fluent-table-scroll` 承担纵向滚动，并使用 `overscroll-behavior: contain` 阻止滚动链传到页面。

## 证据

- `pkg/namespace/server.go`
- `pkg/namespace/namespace.go`
- `apis/store/config_file_api.go`
- `plugin/store/mysql/config_file.go`
- `../specification/api/v1/model/namespace.proto`
- `web/console/src/pages/Namespace/index.tsx`

## 相关页面

- [[architecture]]
- [[storage]]
- [[service-discovery]]
- [[adr-logical-service-environment-binding]]
- [[adr-system-namespace-kind]]
- [[adr-environment-promotion-topology]]
