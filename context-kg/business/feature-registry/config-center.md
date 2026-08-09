---
title: 配置中心（`pkg/config/`）
tags: [business, feature, config]
links: [namespace, architecture, storage, cache-layer, adr-config-template-client-rendering, pole-product-capability-matrix-research]
updated: 2026-07-28
sources: 1
---

# 配置中心（`pkg/config/`）

管理支持灰度发布的版本化配置文件。相同分组名，以及相同分组下的相同文件名，表示同一配置资源在不同命名空间中的独立环境实例；各环境分别维护草稿、版本和发布状态。整体架构见 [[architecture]]，命名空间依赖见 [[namespace]]，存储接口见 [[storage]]，缓存层见 [[cache-layer]]。

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

## 配置模板与 Namespace Value

配置模板客户端渲染能力已经完成 specification、control-plane 参考实现和 Console 管理入口，
SDK 与组合 Watch 尚待接入，完整方案见 [[adr-config-template-client-rendering]]：

- 模板全局复用并独立版本化，配置文件显式固定一个模板发布版本。
- Value 按 `Namespace + Template` 独立维护，并支持正式与灰度发布。
- 服务端根据客户端标签选择唯一命中的 Value Release，不向 SDK 下发灰度规则。
- SDK 获取模板与 Value 的原子组合快照，按 `pole-mustache-v1` 在本地缓存并确定性渲染。
- 服务端提供渲染预览、格式诊断和参考哈希，但预览内容不是客户端运行时权威配置。
- 模板升级必须由配置文件显式切换，不会自动影响已有绑定文件。
- `ConfigFile` 只有纯文本与模板两种类型；渲染后的 Value 是 SDK 的运行时结果，不是第三种文件类型。
- 配置分组使用统一配置清单；创建时选择直接文本或配置模板，两类文件在同一文件树中以来源类型区分，
  全局模板定义作为同级节点直接可见并用尾部标签区分，不增加模板目录，也不再使用同级页签拆分；点击
  模板后保持当前页面，由右侧画布就地展示模板详情；模板与配置文件详情共享摘要头、字段网格、任务页签和
  内容画布的视觉骨架。模板任务依次为模板内容、参数 Schema、环境 Value、版本管理和基本信息；名称、
  格式、说明与标签只在基本信息页签中展开，新建模板默认进入该页签。
- 配置分组的环境切换器固定显示在统一清单上方，选择文件或模板后不会消失；资源详情复用该环境上下文，
  不再渲染第二套环境切换器。分组顶部只保留路径面包屑，不增加重复的工作区身份卡片。
- Console 提供模板草稿与发布、参数 Schema、Namespace Value 正式/灰度发布、服务端参考预览，
  并在配置文件详情中显式切换固定的 Template Release。模板发布操作只出现在模板定义页签，Value 页只
  发布当前环境 Value；两者分别通过发布确认弹窗审阅对象与发布参数，版本管理保持只读。

## 监听机制（`pkg/config/watcher.go`）

- 客户端使用文件 key + 版本号进行注册
- 服务端通过 EventHub 在发布事件时通知客户端
- 支持 SSE（服务器推送事件）和长轮询

## 相关页面

- [[namespace]]
- [[architecture]]
- [[storage]]
- [[cache-layer]]
- [[adr-config-template-client-rendering]]
- [[pole-product-capability-matrix-research]]
