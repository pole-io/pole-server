---
title: Skill Marketplace 功能档案
tags: [business, feature, ai, skill, marketplace]
links: [ai-features, domain-models, terminology, auth-system, adr-skill-marketplace-federation]
updated: 2026-08-12
sources: 3
---

# Skill Marketplace 功能档案

Skill Marketplace 是 AI 工具下的一等资源目录，管理可跨团队、跨 Registry 复用的 Agent
Skills。它与 MCP Tool、A2A Agent Card 中的 skill 声明以及 Pole 内置 Go 插件是不同资源。

## 产品边界

- Marketplace 管理 Skill 定义、发布者、不可变 Release、Bundle、审核和外部来源。
- Skill Bundle 遵循 Agent Skills 开放格式：根目录必须有 `SKILL.md`，可包含
  `scripts/`、`references/`、`assets/` 和其他文件。
- Control Plane 只校验、审核、索引和分发 Bundle，不执行其中指令或脚本。
- 客户端安装、升级、本地 lockfile、回滚与 Agent 目录适配归 `pole-ai-cli`。
- Control Plane 不保存客户端安装状态，不将 Skill 作为 Agent Runtime 或容器服务调度。

## 用户角色

- **Publisher Owner**：管理发布者身份、成员和签名公钥。
- **Publisher Member**：创建 Skill，上传 Bundle 或从 Git Tag/Release 导入候选版本。
- **Reviewer**：对公开候选版本批准或驳回，不能覆盖已发布内容。
- **Registry Administrator**：管理外部 Registry Source、信任等级、同步周期和失败重试。
- **Consumer**：在 Console 搜索、查看，或通过 `pole-ai-cli` 下载可见的精确 Release。

## 可见性与生命周期

- 私有 Skill 默认仅 Publisher Owner 可见，可授权给 Pole User、UserGroup 或 Role。
- 公开已发布 Skill 的目录、元数据和 Bundle 允许匿名只读；草稿、审核与写操作必须鉴权。
- Release 状态为 `draft -> pending_review -> published`，候选版本可被 `rejected`；
  已发布版本只能 `yanked` 或 `deprecated`，不能覆盖或原地修改。
- 私有 Release 扫描通过即可发布；公开 Release 默认需要人工审核，受信 Publisher
  可配置为扫描通过后自动发布。

## Console 信息架构

- `/ai/skills`：Marketplace 目录，支持按关键词、发布者、来源、可见性、状态和签名状态筛选。
- `/ai/skills/:publisher/:name`：独立深链详情页，包含精确版本选择、签名/扫描/来源证据和
  安全 Bundle 文件浏览。
- 上传、Git 导入和审核使用短流程编辑层；长文件树和版本历史保留在详情页。
- Git 导入允许单 Skill 仓库，也允许从 `skills/` 等指定目录发现全部一级 Skill。用户必须先预览
  Tag/Release、commit、统一 SemVer、路径和摘要，再确认批量导入；结果按 Skill 区分成功、跳过和失败。
- 页面不提供“执行”或服务端“安装”，只展示可复制的精确 `pole-ai` CLI 命令。

## 相关页面

- [[ai-features]]
- [[domain-models]]
- [[terminology]]
- [[auth-system]]
- [[adr-skill-marketplace-federation]]
