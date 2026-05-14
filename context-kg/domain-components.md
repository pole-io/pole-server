---
title: 业务领域组件
tags: [domain, service, config, skill]
links: [architecture, storage, cache-layer, ai-features, auth-system]
updated: 2026-05-14
sources: 1
---

# 业务领域组件

本文覆盖六大业务领域的实现细节。整体架构见 [[architecture]]，存储接口见 [[storage]]，缓存机制见 [[cache-layer]]，AI 原生功能见 [[ai-features]]，认证拦截器见 [[auth-system]]。

## 命名空间（`pkg/namespace/`）

命名空间是最顶层的多租户单元。所有其他资源（服务、配置、技能）都归属于某个命名空间。

**关键常量：**
```go
SystemNamespace     = "pole-system"
DefaultNamespace    = "default"
ProductionNamespace = "Production"
DefaultTTL          = 5
```

**`NamespaceOperateServer` 接口**（`apis/pkg/types/...`）：
- `CreateNamespace`、`UpdateNamespace`、`DeleteNamespace`
- `GetNamespace`、`GetNamespaces`（分页）
- 命名空间可见性管理（跨命名空间资源共享）

**实现方式：** `pkg/namespace/Server` 使用 singleflight 防止并发重复创建命名空间。

---

## 服务发现（`pkg/service/`）

核心服务网格功能。管理服务、实例、客户端和服务契约。

**`DiscoverServer` 接口** 由以下子接口组合而成：
- `ServiceOperateServer` — 服务及别名的增删改查
- `InstanceOperateServer` — 实例的增删改查（注册/注销/更新）
- `ClientServer` — 客户端注册与管理
- `ServiceContractOperateServer` — 服务接口契约

**`Server` 结构体** 关键字段：
```go
type Server struct {
    config          Config
    storage         store.Store
    namespaceSvr    namespace.NamespaceOperateServer
    caches          CacheManager
    bc              *batch.Controller       // 批量注册/心跳
    healthServer    *healthcheck.Server
    createServiceSingle singleflight.Group  // 防止并发重复创建服务
    subCtxs         []*eventhub.Subscription
    instanceChains  []InstanceEventHandler
    emptyPushProtectSvs sync.Map            // 每个服务的空推送保护
}
```

**Batch Controller**（`pkg/service/batch/`）：
- 对高频操作进行批处理：注册、注销、心跳
- 可配置 `batchSize` 和 `concurrency`
- 达到批次上限或定时器触发时刷新

**健康检查**（`pkg/service/healthcheck/`、`plugin/service/healthchecker/`）：
- 基于心跳的健康检查机制
- 插件：`heartbeat`（默认）
- 追踪 `lastHeartbeatTime`；不健康的实例在服务发现中不可见

**空推送保护：**
- 针对单个服务：当所有实例消失（网络故障）时，可选择保留最后已知的实例集合
- 防止网络恢复期间的级联故障

---

## 配置中心（`pkg/config/`）

管理支持灰度发布的版本化配置文件。

**`ConfigCenterServer` 接口** 由以下子接口组合而成：
- `ConfigFileGroupOperate` — 配置分组的增删改查
- `ConfigFileOperate` — 文件的增删改查、导入/导出
- `ConfigFileReleaseOperate` — 发布、释放、回滚、灰度发布
- `ConfigFileClientOperate` — 客户端拉取操作
- `ConfigFileWatchOperate` — 长轮询监听

**限制：**
- 最大分页大小：100
- 最大文件内容：20,000 个字符

**关键流程：**
1. **编辑**：创建/更新配置文件（以草稿形式存储）
2. **发布**：创建发布快照
3. **灰度发布**：通过标签匹配向部分客户端发布
4. **回滚**：恢复到之前的某个版本
5. **监听**：客户端对变更进行长轮询；当版本号变化时，服务端响应

**监听机制**（`pkg/config/watcher.go`）：
- 客户端使用文件 key + 版本号进行注册
- 服务端通过 EventHub 在发布事件时通知客户端
- 支持 SSE（服务器推送事件）和长轮询

---

## 治理规则（`pkg/goverrule/`）

所有治理规则遵循相同的模式：增删改查 + 版本控制 + 灰度发布。

**规则类型：**
| 规则 | 用途 |
|------|------|
| **路由规则** | 基于标签/元数据的流量路由 |
| **限流规则** | 针对服务或实例的速率限制 |
| **熔断规则** | 自动故障隔离 |
| **故障探测规则** | 主动健康探测 |
| **泳道规则** | 多泳道（环境隔离）路由 |
| **无损规则** | 优雅关闭/启动 |

所有规则支持：
- 版本控制（版本号追踪）
- 灰度发布（向部分用户发布）
- 细粒度权限（控制谁可以编辑哪些规则）

**`Server` 结构体** 关键字段：
```go
type Server struct {
    config    Config
    storage   store.Store
    namespaceSvr    namespace.NamespaceOperateServer
    caches    CacheManager
    cmdb      plugin.CMDB              // 校验资源标签
    subCtxs   []*eventhub.Subscription
    emptyPushProtectSvs sync.Map
}
```

---

## 管理后台（`pkg/admin/`）

控制平面自身的运维操作。

**`AdminOperateServer` 接口：**
- `Maintain` — 维护操作（数据清理、批量操作）
- 服务端诊断与检查
- 拦截器链：`["auth"]`（管理操作需要认证）

**依赖关系：** 持有所有其他服务组件的引用，以便执行跨领域的管理操作。

---

## Skill Hub（`pkg/skill/`）

AI 原生功能：类配置中心风格的 AI 技能（函数/工具/智能体）注册中心。详细设计见 [[ai-features]]。

**`SkillServer` 接口：**

```go
// 核心增删改查
CreateSkill, UpdateSkill, DeleteSkill, GetSkill, GetSkillByName

// 批量操作
CreateSkills, UpdateSkills, DeleteSkills, GetSkills, GetAllSkills

// 分组（相关技能的集合）
CreateSkillGroup, UpdateSkillGroup, DeleteSkillGroup, GetSkillGroup
CreateSkillGroups, UpdateSkillGroups, DeleteSkillGroups, GetSkillGroups

// 版本（技能的版本控制）
CreateSkillVersion, ActivateSkillVersion
CreateSkillVersions, DeleteSkillVersions, GetSkillVersions

// 订阅（客户端与技能的关联关系）
CreateSkillSubscription, DeleteSkillSubscription
GetSkillSubscriptionsBySkill, GetSkillSubscriptionsByClient
CreateSkillSubscriptions, DeleteSkillSubscriptions, GetSkillSubscriptions
```

**领域模型**（`apis/pkg/types/ai/skill.go`）：
```go
type Skill struct {
    ID           string
    Name         string            // 命名空间内唯一
    Namespace    string
    Description  string
    InputSchema  string            // JSON Schema
    OutputSchema string            // JSON Schema
    SkillType    string            // "function" | "tool" | "agent"
    Author       string
    Business     string
    Department   string
    Metadata     map[string]string
    Protocol     string
    Revision     string
    ExportTo     []string          // 共享到其他命名空间
    Flag         int               // 0=可见，1=软删除
    CTime, MTime time.Time
}
```

**SkillGroup** — 技能的逻辑分组（类似配置文件分组）。
**SkillVersion** — 不可变快照；`ActivateSkillVersion` 将某个版本提升为活跃版本。
**SkillSubscription** — 记录哪个客户端订阅了哪个技能。

**初始化：**
```go
// pkg/skill/skill.go
func Initialize(s store.AIStore) error  // 由 bootstrap 调用
func GetServer() (SkillServer, error)  // 单例访问器
```

## 相关页面

- [[architecture]]
- [[storage]]
- [[cache-layer]]
- [[ai-features]]
- [[auth-system]]
