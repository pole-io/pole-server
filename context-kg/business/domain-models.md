---
title: 核心业务实体
tags: [business, domain-model]
links: [terminology, business-rules, namespace, service-discovery, config-center, governance-rules, ai-features, auth-system, adr-rpc-first-governance-scope]
updated: 2026-07-26
sources: 2
---

# 核心业务实体

| 实体 | 归属页面 | 说明 |
|------|----------|------|
| Namespace | [[namespace]] | 顶层运行环境，统一承载并隔离服务、配置和治理规则的环境实例 |
| Service | [[service-discovery]] | 服务名是全局逻辑标识；`namespace + serviceName` 定位一个独立环境实例 |
| Instance | [[service-discovery]] | 服务运行节点，按健康、隔离、位置和元数据参与发现 |
| Config File | [[config-center]] | 版本化配置内容，支持发布、回滚和监听 |
| Governance Rule | [[governance-rules]] | 针对服务调用的类型化期望策略，支持草稿、发布和回滚 |
| Service Policy Bundle | [[adr-rpc-first-governance-scope]] | 面向某类服务数据面编译并原子应用的规则发布集合 |
| Enforcement Point | [[adr-rpc-first-governance-scope]] | 执行规则并返回能力档案与应用回执的 SDK、Sidecar 或 Gateway |
| MCP Server | [[ai-features]] | AI 原生能力中的 MCP 服务注册实体 |
| System Role | [[auth-system]] | 固定权限集合，只维护与 User、UserGroup 的成员关系 |

实体之间的核心关系：

```text
Namespace
  -> Service
    -> Instance
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
