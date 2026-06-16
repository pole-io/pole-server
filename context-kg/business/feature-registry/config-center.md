---
title: 配置中心（`pkg/config/`）
tags: [business, feature, config]
links: [namespace, architecture, storage, cache-layer]
updated: 2026-06-15
sources: 1
---

# 配置中心（`pkg/config/`）

管理支持灰度发布的版本化配置文件。整体架构见 [[architecture]]，命名空间依赖见 [[namespace]]，存储接口见 [[storage]]，缓存层见 [[cache-layer]]。

## `ConfigCenterServer` 接口

`ConfigCenterServer` 接口由以下子接口组合而成：
- `ConfigFileGroupOperate` — 配置分组的增删改查
- `ConfigFileOperate` — 文件的增删改查、导入/导出
- `ConfigFileReleaseOperate` — 发布、释放、回滚、灰度发布
- `ConfigFileClientOperate` — 客户端拉取操作
- `ConfigFileWatchOperate` — 长轮询监听

## 限制

- 最大分页大小：100
- 最大文件内容：20,000 个字符

## 关键流程

1. **编辑**：创建/更新配置文件（以草稿形式存储）
2. **发布**：创建发布快照
3. **灰度发布**：通过标签匹配向部分客户端发布
4. **回滚**：恢复到之前的某个版本
5. **监听**：客户端对变更进行长轮询；当版本号变化时，服务端响应

## 监听机制（`pkg/config/watcher.go`）

- 客户端使用文件 key + 版本号进行注册
- 服务端通过 EventHub 在发布事件时通知客户端
- 支持 SSE（服务器推送事件）和长轮询

## 相关页面

- [[namespace]]
- [[architecture]]
- [[storage]]
- [[cache-layer]]
