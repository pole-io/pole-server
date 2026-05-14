---
title: Skill Hub 域
tags: [domain, skill, ai]
links: [architecture, storage, ai-features, auth-system]
updated: 2026-05-14
sources: 1
---

# Skill Hub（`pkg/skill/`）

AI 原生功能：类配置中心风格的 AI 技能（函数/工具/智能体）注册中心。整体架构见 [[architecture]]，存储接口见 [[storage]]，详细设计见 [[ai-features]]，认证拦截器见 [[auth-system]]。

## `SkillServer` 接口

```go
// 核心增删改查
CreateSkill, UpdateSkill, DeleteSkill, GetSkill, GetSkillByName

// 批量操作
CreateSkills, UpdateSkills, DeleteSkills, GetSkills, GetAllSkills

// 分组（相关技能的集合）
CreateSkillGroup, UpdateSkillGroup, DeleteSkillGroup, GetSkillGroup
CreateSkillGroups, UpdateSkillGroups, DeleteSkillGroups, GetSkillGroups

// 版本（技能的版本控制）
CreateSkillVersion, ActivateSkillVersion
CreateSkillVersions, DeleteSkillVersions, GetSkillVersions

// 订阅（客户端与技能的关联关系）
CreateSkillSubscription, DeleteSkillSubscription
GetSkillSubscriptionsBySkill, GetSkillSubscriptionsByClient
CreateSkillSubscriptions, DeleteSkillSubscriptions, GetSkillSubscriptions
```

## 领域模型（`apis/pkg/types/ai/skill.go`）

```go
type Skill struct {
    ID           string
    Name         string            // 命名空间内唯一
    Namespace    string
    Description  string
    InputSchema  string            // JSON Schema
    OutputSchema string            // JSON Schema
    SkillType    string            // "function" | "tool" | "agent"
    Author       string
    Business     string
    Department   string
    Metadata     map[string]string
    Protocol     string
    Revision     string
    ExportTo     []string          // 共享到其他命名空间
    Flag         int               // 0=可见，1=软删除
    CTime, MTime time.Time
}
```

**SkillGroup** — 技能的逻辑分组（类似配置文件分组）。

**SkillVersion** — 不可变快照；`ActivateSkillVersion` 将某个版本提升为活跃版本。

**SkillSubscription** — 记录哪个客户端订阅了哪个技能。

## 初始化

```go
// pkg/skill/skill.go
func Initialize(s store.AIStore) error  // 由 bootstrap 调用
func GetServer() (SkillServer, error)  // 单例访问器
```

## 相关页面

- [[architecture]]
- [[storage]]
- [[ai-features]]
- [[auth-system]]
