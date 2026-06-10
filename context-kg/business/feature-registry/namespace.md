---
title: 命名空间域
tags: [business, feature, namespace]
links: [architecture, storage, service-discovery]
updated: 2026-06-09
sources: 1
---

# 命名空间（`pkg/namespace/`）

命名空间是最顶层的多租户单元。所有其他资源（服务、配置、技能）都归属于某个命名空间。整体架构见 [[architecture]]，存储接口见 [[storage]]，服务发现如何依赖命名空间见 [[service-discovery]]。

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
- 命名空间可见性管理（跨命名空间资源共享）

## 实现方式

`pkg/namespace/Server` 使用 singleflight 防止并发重复创建命名空间。

## 相关页面

- [[architecture]]
- [[storage]]
- [[service-discovery]]
