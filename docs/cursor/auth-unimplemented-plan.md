# 鉴权未实现逻辑 — 实现计划与任务拆解

> 基于单 Owner 前提（一套集群只有一个 owner）。  
> 文档生成后可在 Cursor 中用 Todo 跟踪：任务 2、3 → 任务 1 → 任务 4（可选）。

---

## 一、总览

| 序号 | 任务 | 优先级 | 预估复杂度 |
|-----|------|--------|------------|
| 1 | 实现 **AuthorizeResources** | 高 | 中 |
| 2 | **角色创建操作日志** | 高 | 低 |
| 3 | **策略更新 paramcheck 返回值** | 高 | 低 |
| 4 | （可选）扩展 **checkResourceExist** | 中 | 中 |

---

## 二、任务拆解

### 任务 1：实现 AuthorizeResources

**目标**：根据 `AuthorizeResources` 请求，把指定资源授权给若干 principal（用户/用户组/角色），即给这些 principal 的**默认策略**增加对应资源关联。

**依赖**：
- 请求结构：`pkg/common/api/v1/auth.go` — `ResourceType`、`ResourceName`、`ResourceID`、`Principals`
- 现有能力：`server.changePrincipalPolicies`、`storage.GetDefaultStrategyDetailByPrincipal`、`storage.LooseAddStrategyResources`

**子任务**：

| 步骤 | 内容 | 说明 |
|------|------|------|
| 1.1 | **请求校验** | 校验 `reqs` 非空；单条请求中 `ResourceType`、`ResourceID`、`Principals` 必填。 |
| 1.2 | **ResourceType 映射** | 将请求中的 `ResourceType` 字符串（如 `namespace`/`service`/`config_group`）映射为 `apisecurity.ResourceType`，可复用或扩展 `policy.go` 中 `resTypeFilter`。 |
| 1.3 | **资源存在性校验** | 按映射后的 `ResourceType` 查对应 cache（如 namespace、service、config_group），校验 `ResourceID` 存在；支持 `*` 表示不校验。 |
| 1.4 | **Principal 存在性校验** | 对 `Principals.Users/Groups/Roles` 分别用 `GetUserHelper().CheckUsersExist`、`CheckGroupsExist`，角色用 `PolicyHelper().GetRole` 校验存在。 |
| 1.5 | **按 principal 写默认策略资源** | 对每个 principal（user/group/role），调用 `GetDefaultStrategyDetailByPrincipal(id, principalType)` 取默认策略；构造 `StrategyResource{StrategyID, ResType, ResID}` 列表；调用 `LooseAddStrategyResources` 写入。 |
| 1.6 | **操作日志（可选）** | 与 `AfterResourceOperation` 一致，对本次授权记录一条 history（类型可为 `RAuthStrategy` 或单独类型，按现有约定）。 |
| 1.7 | **返回** | 全部成功返回 `Code_ExecuteSuccess`；任一步失败返回对应错误码与信息。 |

**实现位置**：`plugin/access_control/auth/policy/policy.go` 中 `AuthorizeResources` 函数。

---

### 任务 2：角色创建操作日志

**目标**：创建角色成功后写入操作历史，与用户/策略等保持一致。

**子任务**：

| 步骤 | 内容 | 说明 |
|------|------|------|
| 2.1 | **在 CreateRole 成功分支打日志** | 在 `AddRole` 成功后、`return` 前，调用 `svr.RecordHistory(recordRoleEntry(ctx, req, saveData, types.OCreate))`。 |
| 2.2 | **确认 recordRoleEntry 可用** | `recordRoleEntry` 已在 `role.go` 中定义，入参为 `(ctx, req, data, op)`，直接复用即可。 |

**实现位置**：`plugin/access_control/auth/policy/role.go` 中 `CreateRole`。

---

### 任务 3：策略更新 paramcheck 返回值修复

**目标**：在 paramcheck 层对批量更新策略做校验时，若有任意一条校验失败，应直接返回校验结果，不再调用下层更新。

**子任务**：

| 步骤 | 内容 | 说明 |
|------|------|------|
| 3.1 | **根据校验结果决定是否调用下层** | 在 `UpdatePolicies` 中，for 循环里已用 `api.Collect(batchResp, rsp)` 收集每条 `checkUpdateStrategy` 的 `rsp`。循环结束后：若 `!api.IsSuccess(batchResp)`，则 `return batchResp`；否则再 `return svr.nextSvr.UpdatePolicies(ctx, reqs)`。 |
| 3.2 | **处理 strategy == nil** | 当前 `strategy == nil` 时是 `continue`，未向 `batchResp` 写入。应改为对当前 req 写入一条“未找到策略”的响应并 `api.Collect`，保证批量结果完整。 |

**实现位置**：`plugin/access_control/auth/policy/inteceptor/paramcheck/server.go` 中 `UpdatePolicies`。

---

### 任务 4：（可选）扩展 checkResourceExist

**目标**：创建/更新策略时，除 Namespaces、Services 外，对请求中携带的其它资源类型也做存在性校验，避免策略指向不存在资源。

**子任务**：

| 步骤 | 内容 | 说明 |
|------|------|------|
| 4.1 | **ConfigGroups 校验** | 在 `checkResourceExist` 中增加对 `resources.GetConfigGroups()` 的遍历，用 `cacheMgr.ConfigGroup().GetGroupByID(id)` 校验（`*` 跳过）。 |
| 4.2 | **按需扩展其它类型** | 若策略支持 RouteRules、RatelimitRules、CircuitbreakerRules、FaultDetectRules、LaneRules、Users、UserGroups、Roles、PolicyRules 等，可仿照 `policy.go` 中 `resourceConvert` 的 cache 用法，逐类增加校验。 |
| 4.3 | **统一约定** | 所有资源类型均支持 `id == "*"` 时跳过该校验。 |

**实现位置**：`plugin/access_control/auth/policy/inteceptor/paramcheck/server.go` 中 `checkResourceExist`。

---

## 三、建议执行顺序

1. **先做任务 2、3**：改动小、无依赖，可快速合入并减少明显缺陷。
2. **再做任务 1**：依赖现有存储与 server 逻辑，实现完整授权能力。
3. **最后做任务 4**：按产品/运维需求决定是否扩展以及扩展哪些资源类型。

---

## 四、验收要点

- **任务 1**：调用 `AuthorizeResources` 后，对应 principal 的默认策略下能看到新增的资源关联；资源或 principal 不存在时返回明确错误。
- **任务 2**：创建角色后，history 中能查到一条创建记录，格式与现有角色/用户记录一致。
- **任务 3**：批量更新策略时，若某条校验失败（如 principal 不存在、资源不存在），接口返回该条失败且**不执行**下层更新。
- **任务 4**：创建/更新策略时若带了 ConfigGroup（及后续扩展类型）且 ID 不存在，返回资源不存在类错误。

---

## 五、Todo 清单（可与 Cursor Todo 同步）

- [ ] **任务 1**：实现 AuthorizeResources（解析请求、校验资源与 principal、写默认策略资源）
- [ ] **任务 2**：角色创建操作日志（CreateRole 中调用 RecordHistory(recordRoleEntry)）
- [ ] **任务 3**：策略更新 paramcheck（校验失败时返回 batchResp 而非继续调用下层）
- [ ] **任务 4**：（可选）扩展 checkResourceExist 覆盖 ConfigGroup 等更多资源类型
