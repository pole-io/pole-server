---
title: 命名空间（`pkg/namespace/`）
tags: [business, feature, namespace]
links: [architecture, storage, service-discovery]
updated: 2026-07-23
sources: 6
---

# 命名空间（`pkg/namespace/`）

命名空间是 Pole 最顶层的运行环境。服务、配置、治理规则等资源都在命名空间中形成相互隔离的环境实例；权限、发布状态和运行数据也以命名空间为环境边界。整体架构见 [[architecture]]，存储接口见 [[storage]]，服务发现如何依赖命名空间见 [[service-discovery]]。

## 跨环境资源身份

- 服务名是全局逻辑服务标识；`namespace + serviceName` 定位该服务的一个环境实例。
- 配置分组名是跨环境逻辑标识；`namespace + groupName` 定位该分组的一个环境实例。
- 配置文件以 `groupName + fileName` 作为跨环境逻辑标识；`namespace + groupName + fileName` 定位具体环境中的文件。
- 治理规则以 `ruleType + ruleName` 作为跨环境逻辑标识；`namespace + ruleType + ruleName` 定位具体环境中的规则。
- 同一逻辑资源的不同环境实例可以具有独立内容、实例、版本和发布状态，任何复制、提升或发布都必须显式作用于目标命名空间。

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
- `pole-system` 是 Pole 内部组件和系统资源使用的内部命名空间，不允许删除。
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
- `console/web/src/pages/Namespace/index.tsx`

## 相关页面

- [[architecture]]
- [[storage]]
- [[service-discovery]]
