---
title: 管理后台（`pkg/admin/`）
tags: [business, feature, admin]
links: [architecture, auth-system]
updated: 2026-06-15
sources: 1
---

# 管理后台（`pkg/admin/`）

控制平面自身的运维操作。整体架构见 [[architecture]]，认证拦截器见 [[auth-system]]。

## `AdminOperateServer` 接口

- `Maintain` — 维护操作（数据清理、批量操作）
- 服务端诊断与检查
- 拦截器链：`["auth"]`（管理操作需要认证）

## 依赖关系

持有所有其他服务组件的引用，以便执行跨领域的管理操作。

## 相关页面

- [[architecture]]
- [[auth-system]]
