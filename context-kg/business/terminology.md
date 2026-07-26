---
title: 领域术语表
tags: [business, terminology]
links: [domain-models, business-rules, auth-system, adr-console-agent-resource-workbench, adr-system-configuration-control-plane, adr-rpc-first-governance-scope, adr-service-contract-reporting-and-visualization]
updated: 2026-07-27
sources: 0
---

# 领域术语表

| 中文术语 | English | 含义 |
|----------|---------|------|
| 命名空间 | Namespace | Pole 的运行环境边界；服务、配置、治理规则等资源在各命名空间中形成相互隔离的环境实例 |
| 服务 | Service | 由全局唯一服务名标识的逻辑服务；同名服务在不同命名空间中的记录表示该服务的不同环境实例 |
| 实例 | Instance | 服务的运行节点，包含 host、port、协议、健康状态、隔离状态和元数据 |
| 服务契约 | Service Contract | 协议原始定义与可检索接口投影的组合；原文是事实来源，投影用于发现、治理和展示 |
| Dubbo 元数据快照 | Dubbo Metadata Snapshot | Dubbo 按 application 与 metadata revision 聚合的接口配置和方法定义 |
| Dubbo 服务映射 | Dubbo Service Mapping | Dubbo interface 到 provider application 的一对多关系，用于应用级服务发现 |
| 配置文件 | Config File | 配置中心管理的版本化配置内容 |
| 服务调用 | Service Invocation | 一次具有调用方、被调方、接口、请求属性和结果状态的 HTTP/RPC 交互 |
| 服务治理能力 | Service Governance Capability | 作用于服务调用的路由、限流、鉴权、镜像、Mock、熔断等能力 |
| 治理规则 | Governance Rule | 声明对哪些服务调用、在何种条件下施加一个服务治理效果的期望策略聚合根 |
| 执行点 | Enforcement Point | 真正执行服务治理规则的 SDK、Sidecar 或 Gateway |
| 能力档案 | Capability Profile | 执行点声明的服务治理规则类型、匹配能力、动作版本和可选特性集合 |
| 应用回执 | Apply Receipt | 执行点对某个服务治理 Bundle 返回的已应用、拒绝或部分应用状态 |
| 发布版本 | Release | 配置或治理规则发布时形成的快照版本 |
| 灰度发布 | Gray Release | 面向部分客户端、标签或条件生效的发布方式 |
| 泳道组 | Lane Group | 一组泳道路由规则的聚合根 |
| MCP 服务 | MCP Server | AI 工具注册中心中的 MCP 服务定义，可关联 Pole 服务或自定义地址 |
| Pole Agent | Pole Agent | Console 提供的一等对话主体，内部使用模型和 MCP 工具操作 Pole，写操作必须等待人工确认 |
| Agent 定义 | Agent Definition | Pole Agent 的版本化行为定义，包含 Prompt 版本、模型 profile、MCP 工具集和运行限制，不包含密钥明文 |
| 启动配置 | Bootstrap Configuration | 让 Console 能够启动并定位外部依赖的最小不可热更新配置，包括配置源、凭证引用和紧急开关 |
| Console 系统设置 | Console System Settings | 面向管理员的类型化运行配置页面；底层使用 Pole 配置中心的草稿、发布、历史和回滚能力 |
| 系统配置定义 | Setting Definition | 描述系统设置的类型、默认值、校验、敏感级别、归属组件、兼容版本和生效方式 |
| 期望配置快照 | Desired Snapshot | 控制面当前希望目标实例采用的已发布系统配置版本 |
| 有效配置快照 | Effective Snapshot | 某个实例实际使用的最终配置及其默认值、静态文件或动态覆盖来源 |
| 配置应用回执 | Apply Receipt | 实例对某个系统配置版本返回的已应用、拒绝或等待重启状态 |
| 系统角色 | Built-in Role | Pole 固定提供的 `admin`、`resource-reader`、`resource-writer` 权限集合；角色定义不可增删改，只允许维护用户和用户组成员关系 |
| 自定义角色 | Custom Role | 管理员按职责创建的角色；可维护角色定义、用户/用户组成员及资源/API 权限，并可删除 |

## 相关页面

- [[domain-models]]
- [[business-rules]]
- [[auth-system]]
- [[adr-console-agent-resource-workbench]]
- [[adr-system-configuration-control-plane]]
- [[adr-rpc-first-governance-scope]]
- [[adr-service-contract-reporting-and-visualization]]
