---
title: 核心业务实体
tags: [business, domain-model]
links: [terminology, business-rules, namespace, service-discovery, config-center, governance-rules, ai-features, auth-system, skill-marketplace, adr-skill-marketplace-federation, adr-rpc-first-governance-scope, adr-service-contract-reporting-and-visualization, adr-logical-service-environment-binding, adr-system-namespace-kind, adr-ai-resource-environment-binding, adr-config-template-client-rendering, adr-environment-promotion-topology]
updated: 2026-08-11
sources: 5
---

# 核心业务实体

| 实体 | 归属页面 | 说明 |
|------|----------|------|
| Namespace | [[namespace]] | 顶层运行边界；Kind 区分参与跨环境聚合的 BUSINESS 与 Pole 内部 SYSTEM |
| Environment Promotion Topology | [[namespace]] | 用户全局定义的 BUSINESS Namespace 晋升 DAG，统一约束配置、治理、服务与 AI 资源的跨环境流转 |
| Lane Base Binding | [[namespace]] | lane 到唯一 baseline 的分叉归属，不进入 baseline 晋升 DAG |
| Environment Promotion Change Set | [[namespace]] | 用户选择的跨资源域已发布版本集合，带分叉基线并在目标环境生成候选变更 |
| Environment Promotion Bundle | [[namespace]] | 审批后不可拆分发布的目标资源版本集，提供原子期望状态与最终收敛回执 |
| Promotion Adapter | [[namespace]] | 可版本化期望状态资源域的晋升接入契约；未实现适配器的资源不得进入变更集 |
| Logical Service | [[adr-logical-service-environment-binding]] | 控制面跨环境聚合根，由稳定 ID 标识，SDK 不感知 |
| Service Environment | [[service-discovery]] | `namespace + runtimeServiceName` 定位一个独立运行环境服务 |
| Instance | [[service-discovery]] | 服务运行节点，按健康、隔离、位置和元数据参与发现 |
| Service Contract | [[adr-service-contract-reporting-and-visualization]] | 服务在某版本下的协议原始定义及其统一接口投影 |
| Dubbo Metadata Snapshot | [[adr-service-contract-reporting-and-visualization]] | 以 application + metadata revision 标识的 Dubbo 原生接口配置快照 |
| Dubbo Service Mapping | [[adr-service-contract-reporting-and-visualization]] | Dubbo interface 到一个或多个 provider application 的映射关系 |
| Config File | [[config-center]] | 版本化配置内容，支持发布、回滚和监听 |
| Config Template Identity | [[adr-config-template-client-rendering]] | 跨环境稳定逻辑身份，不承载可执行草稿 |
| Namespace Config Template Draft | [[adr-environment-promotion-topology]] | Namespace 隔离的内容、格式、Schema 和引擎草稿 |
| Environment Config Release | [[adr-config-template-client-rendering]] | 环境范围内原子绑定 Template Snapshot 与 Value Snapshot 的不可变发布聚合 |
| Governance Rule | [[governance-rules]] | 针对服务调用的类型化期望策略，支持草稿、发布和回滚 |
| Service Policy Bundle | [[adr-rpc-first-governance-scope]] | 面向某类服务数据面编译并原子应用的规则发布集合 |
| Enforcement Point | [[adr-rpc-first-governance-scope]] | 执行规则并返回能力档案与应用回执的 SDK、Sidecar 或 Gateway |
| MCP Definition | [[adr-ai-resource-environment-binding]] | 控制面跨环境 MCP 聚合根，由稳定 ID 标识 |
| MCP Deployment | [[ai-features]] | Namespace 中的 MCP Server 环境实例 |
| Agent Definition | [[adr-ai-resource-environment-binding]] | 控制面跨环境 Agent 聚合根，由稳定 ID 标识 |
| Agent Deployment | [[ai-features]] | Namespace 中的 A2A Agent 环境实例 |
| Skill Publisher | [[skill-marketplace]] | Marketplace 的发布身份、成员与签名信任聚合 |
| Skill | [[skill-marketplace]] | 以 `publisher/name` 定位的跨 Registry 逻辑资产 |
| Skill Release | [[adr-skill-marketplace-federation]] | 绑定精确 SemVer、Bundle 摘要、签名和审核证据的不可变版本 |
| Skill Bundle | [[adr-skill-marketplace-federation]] | 经安全验证并按 SHA-256 内容寻址的 Agent Skills 目录快照 |
| Registry Source | [[adr-skill-marketplace-federation]] | 可定期同步的 Pole、Git 或 HTTP 外部目录及其信任/健康状态 |
| System Role | [[auth-system]] | 固定权限集合，只维护与 User、UserGroup 的成员关系 |

实体之间的核心关系：

```text
Namespace
  -> Namespace Kind (BUSINESS | SYSTEM)
  -> Environment Promotion Topology
    -> Topology Draft
    -> Immutable Topology Revision
    -> Promotion Edge (source baseline -> target baseline)
      -> Source Release Requirement
      -> Approval Policy
      -> Validation Gate
      -> Allowed Resource Domains
      -> Conflict Policy
    -> Lane Base Binding (lane -> one baseline)
    -> Environment Promotion Change Set
      -> Pinned Topology Revision
      -> Pinned Edge Policy
      -> Resource Version Reference(s)
      -> Fork Baseline Reference(s)
      -> Merge Conflict(s)
      -> Environment Promotion Bundle
        -> Immutable Target Resource Version(s)
        -> Atomic Desired-state Commit
        -> Outbox Event(s)
        -> Apply Receipt(s)
  -> Service Environment
    -> Instance
    -> Service Contract
      -> Interface Descriptor
      -> Dubbo Metadata Snapshot
        -> Dubbo Service Mapping
    -> MCP Server backend selector
  -> Config File Group
    -> Config File
      -> Release
      -> Config Template Identity Binding
  -> Namespace Config Template Draft
  -> Environment Config Release
    -> Template Snapshot
    -> Value Snapshot
  -> Governance Rule
    -> Policy Release

Promotion Adapter
  -> Stable Logical Resource Identity
  -> Export Immutable Source Version
  -> Target Preflight (CREATE | UPDATE | CONFLICT | UNSUPPORTED)
  -> Diff / Three-way Merge
  -> Build Target Candidate
  -> Validate Candidate

Service Policy Bundle
  -> Policy Release(s)

Enforcement Point
  -> Capability Profile
  -> Apply Receipt

Logical Service
  -> Service Environment Binding
    -> Namespace + Runtime Service Name

MCP Definition
  -> MCP Environment Binding
    -> Namespace + MCP Deployment

Agent Definition
  -> Agent Environment Binding
    -> Namespace + Agent Deployment

Skill Publisher
  -> Skill (publisher/name)
    -> Skill Release (SemVer + Bundle digest)
      -> Release Review
      -> Immutable Skill Bundle
  -> Publisher Signing Key

Registry Source
  -> Synchronized Catalog Snapshot
  -> Frozen Skill Release

User -> System Role
User -> UserGroup -> System Role
System Role -> Immutable Policy
```

角色聚合包含可管理的自定义角色与受保护的内置角色。自定义角色可增删改并关联权限；内置角色只有 `admin`、`resource-reader`、`resource-writer` 三个稳定身份。用户可直接加入角色，也可通过用户组继承；内置角色的名称、描述、权限策略和系统标识不是租户可编辑资源。具体规则见 [[business-rules]]。

## 相关页面

- [[terminology]]
- [[business-rules]]
- [[namespace]]
- [[service-discovery]]
- [[config-center]]
- [[governance-rules]]
- [[ai-features]]
- [[auth-system]]
- [[skill-marketplace]]
- [[adr-skill-marketplace-federation]]
- [[adr-rpc-first-governance-scope]]
- [[adr-service-contract-reporting-and-visualization]]
- [[adr-logical-service-environment-binding]]
- [[adr-system-namespace-kind]]
- [[adr-ai-resource-environment-binding]]
- [[adr-config-template-client-rendering]]
- [[adr-environment-promotion-topology]]
