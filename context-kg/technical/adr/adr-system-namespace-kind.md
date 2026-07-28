---
title: ADR：业务环境与 Pole 系统空间类型化
tags: [adr, namespace, environment, system, console]
links: [namespace, terminology, domain-models, business-rules, adr-logical-service-environment-binding]
updated: 2026-07-28
sources: 8
---

# ADR：业务环境与 Pole 系统空间类型化

## 状态

Accepted。

## 背景

Pole 将 Namespace 定义为服务、配置与治理规则的运行环境边界，但内置的 `pole-system`
承载 Control Plane、MCP、Agent 和系统配置。它描述的是当前 Pole 安装实例的内部管理面，
而不是 dev、test、staging、production 这类业务发布环境。

仅通过名称特判会让 `pole-system` 进入逻辑服务、配置分组和配置文件的跨环境聚合，
也会让 Console 把系统资源计入业务环境指标和选择器。

## 决策

为 Namespace 增加一等 `kind`：

- `BUSINESS = 0`：业务运行环境，也是 protobuf 的兼容默认值。旧 payload 和旧数据库记录
  缺少类型时按业务环境解释。
- `SYSTEM = 1`：Pole 内部系统空间。`pole-system` 在迁移和查询时固定为该类型。

公共 Namespace 创建接口只允许创建 `BUSINESS`，不能通过名称或类型伪造系统空间；
`SYSTEM` 只能由 Pole 安装和迁移流程提供，且不可删除。`default` 仍是不可删除的
`BUSINESS`，因此“系统类型”与“受保护”不是同义词。

系统空间继承当前 Pole 部署自身的阶段、集群和地域信息；这些部署属性不能编码进
Namespace 名称或 `kind`。Kubernetes Namespace 与 Pole Namespace 是不同资源，
即使二者都使用 `pole-system` 这个名称也不能在模型上等同。

## 聚合边界

- Logical Service 只绑定和统计 `BUSINESS` 环境服务；服务端拒绝绑定系统空间服务。
- 配置分组和配置文件的跨环境视图只聚合 `BUSINESS` Namespace。
- 显式访问 `pole-system` 中的内部配置和服务仍然允许，不能通过全局隐藏破坏系统管理。
- 授权继续使用现有 Namespace 可见性和权限谓词；`kind` 是分类，不替代权限。

Console 将业务环境与“Pole 系统空间（当前控制面）”分区展示。业务指标、创建入口、
环境选择器、复制、提升和跨环境比较均排除 `SYSTEM`；系统空间保留受权限控制的查看、
授权和系统维护入口。

## 存储与兼容

MySQL `namespace.kind` 使用 `TINYINT NOT NULL DEFAULT 0`。启动迁移幂等增加字段并将
`pole-system` 回填为 `1`。API 查询支持 `kind=business|system`，缓存必须在分页前完成
类型过滤，保证 `amount` 与页数据一致。

protobuf 使用未占用的 field 10。旧客户端忽略新增字段；新客户端读取旧 payload 时因
`BUSINESS=0` 自动保持历史语义。Console 在滚动升级期间额外以名称识别 `pole-system`，
避免旧服务端未返回 `kind` 时把系统空间混入业务聚合。

## 后果

- Namespace 不再能被笼统等同为业务环境；应先区分 Kind，再解释其运行边界。
- 当前仅由 Pole 提供 `pole-system` 一个系统空间；未来增加其它系统空间时必须沿用
  `SYSTEM` 类型，并把配置查询的服务端类型过滤作为新增空间的前置条件。
- 历史上已经绑定到 Logical Service 的系统服务不再计入环境统计；绑定记录不会被静默删除，
  管理 API 和 Console 会把它单列为“仅清理”记录并保留显式解绑入口。
- 数据库账号必须具备迁移所需 DDL 权限；受控生产环境可以在发布前执行等价迁移。

## 相关页面

- [[namespace]]
- [[terminology]]
- [[domain-models]]
- [[business-rules]]
- [[adr-logical-service-environment-binding]]
