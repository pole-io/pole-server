---
title: 核心业务实体
tags: [business, domain-model]
links: [terminology, business-rules, namespace, service-discovery, config-center, governance-rules, ai-features, auth-system, adr-rpc-first-governance-scope, adr-service-contract-reporting-and-visualization, adr-logical-service-environment-binding, adr-system-namespace-kind]
updated: 2026-07-28
sources: 3
---

# 核心业务实体

| 实体 | 归属页面 | 说明 |
|------|----------|------|
| Namespace | [[namespace]] | 顶层运行边界；Kind 区分参与跨环境聚合的 BUSINESS 与 Pole 内部 SYSTEM |
| Logical Service | [[adr-logical-service-environment-binding]] | 控制面跨环境聚合根，由稳定 ID 标识，SDK 不感知 |
| Service Environment | [[service-discovery]] | `namespace + runtimeServiceName` 定位一个独立运行环境服务 |
| Instance | [[service-discovery]] | 服务运行节点，按健康、隔离、位置和元数据参与发现 |
| Service Contract | [[adr-service-contract-reporting-and-visualization]] | 服务在某版本下的协议原始定义及其统一接口投影 |
| Dubbo Metadata Snapshot | [[adr-service-contract-reporting-and-visualization]] | 以 application + metadata revision 标识的 Dubbo 原生接口配置快照 |
| Dubbo Service Mapping | [[adr-service-contract-reporting-and-visualization]] | Dubbo interface 到一个或多个 provider application 的映射关系 |
| Config File | [[config-center]] | 版本化配置内容，支持发布、回滚和监听 |
| Governance Rule | [[governance-rules]] | 针对服务调用的类型化期望策略，支持草稿、发布和回滚 |
| Service Policy Bundle | [[adr-rpc-first-governance-scope]] | 面向某类服务数据面编译并原子应用的规则发布集合 |
| Enforcement Point | [[adr-rpc-first-governance-scope]] | 执行规则并返回能力档案与应用回执的 SDK、Sidecar 或 Gateway |
| MCP Server | [[ai-features]] | AI 原生能力中的 MCP 服务注册实体 |
| System Role | [[auth-system]] | 固定权限集合，只维护与 User、UserGroup 的成员关系 |

实体之间的核心关系：

```text
Namespace
  -> Namespace Kind (BUSINESS | SYSTEM)
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
  -> Governance Rule
    -> Policy Release

Service Policy Bundle
  -> Policy Release(s)

Enforcement Point
  -> Capability Profile
  -> Apply Receipt

Logical Service
  -> Service Environment Binding
    -> Namespace + Runtime Service Name

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
- [[adr-rpc-first-governance-scope]]
- [[adr-service-contract-reporting-and-visualization]]
- [[adr-logical-service-environment-binding]]
- [[adr-system-namespace-kind]]
