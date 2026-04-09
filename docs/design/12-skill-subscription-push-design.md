# Skill 订阅推送机制设计

## 1. 概述

本文档描述 Skill 订阅推送机制的设计，用于在 Skill 变更时通知订阅者。

## 2. 现有模式分析

### 2.1 EventHub 模式

项目使用 `pkg/common/eventhub` 作为事件发布/订阅框架：

```go
// 发布事件
eventhub.Publish(eventhub.InstanceEventTopic, event)

// 订阅事件
subCtx, _ := eventhub.Subscribe(eventhub.InstanceEventTopic, func(event interface{}) {
    // 处理事件
})
```

### 2.2 现有 Topic 定义 (`pkg/common/eventhub/types.go`)

```go
const (
    InstanceEventTopic     = "instance_event"
    ServiceEventTopic      = "service_event"
    LeaderChangeEventTopic = "leader_change_event"
    ConfigFilePublishTopic = "configfile_publish"
    CacheInstanceEventTopic = "cache_instance_event"
    CacheClientEventTopic   = "cache_client_event"
    CacheNamespaceEventTopic = "cache_namespace_event"
    ClientEventTopic        = "client_event"
)
```

## 3. Skill 事件 Topic 扩展

### 3.1 新增 Topic 定义

在 `pkg/common/eventhub/types.go` 添加：

```go
// Skill 相关事件
SkillEventTopic          = "skill_event"           // Skill 变更事件
SkillVersionEventTopic   = "skill_version_event"   // Skill 版本变更事件
SkillSubscriptionEventTopic = "skill_subscription_event" // 订阅变更事件
```

### 3.2 Skill 事件类型

创建 `pkg/skill/event/types.go`:

```go
package event

import (
    "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// 事件类型
const (
    EventSkillCreate     = "skill_create"
    EventSkillUpdate     = "skill_update"
    EventSkillDelete     = "skill_delete"
    EventVersionCreate   = "version_create"
    EventVersionActivate = "version_activate"
)

// SkillEvent Skill 事件
type SkillEvent struct {
    Type      string
    Skill     *ai.Skill
    Timestamp int64
}

// SkillVersionEvent Skill 版本事件
type SkillVersionEvent struct {
    Type      string
    Version   *ai.SkillVersion
    Timestamp int64
}
```

## 4. 实现方案

### 4.1 在 SkillServer 中发布事件

修改 `pkg/skill/origin_server.go` (或相应实现)：

```go
func (s *Server) CreateSkill(ctx context.Context, skill *ai.Skill) error {
    // 1. 执行创建逻辑
    err := s.storage.CreateSkill(skill)
    if err != nil {
        return err
    }

    // 2. 发布事件
    _ = eventhub.Publish(eventhub.SkillEventTopic, &event.SkillEvent{
        Type:      event.EventSkillCreate,
        Skill:     skill,
        Timestamp: time.Now().Unix(),
    })

    return nil
}
```

### 4.2 缓存层订阅事件

修改 `pkg/cache/ai/skill.go`:

```go
func (sc *skillCache) Initialize(opt map[string]interface{}) error {
    // 订阅 Skill 事件，触发缓存更新
    _, err := eventhub.Subscribe(eventhub.SkillEventTopic, func(e interface{}) {
        evt, ok := e.(*event.SkillEvent)
        if !ok {
            return
        }
        // 触发增量更新
        _ = sc.Update()
    })
    return err
}
```

## 5. 推送消息格式

### 5.1 推送消息结构

```go
// SkillPushMessage 推送给客户端的消息
type SkillPushMessage struct {
    Type      string         `json:"type"`       // 事件类型
    Skill     *ai.Skill      `json:"skill"`      // Skill 信息
    Version   *ai.SkillVersion `json:"version"`  // 版本信息（可选）
    Timestamp int64          `json:"timestamp"`  // 时间戳
}
```

### 5.2 推送失败重试

```go
type PushRetryConfig struct {
    MaxRetries    int           // 最大重试次数
    RetryInterval time.Duration // 重试间隔
}

func (p *Pusher) pushWithRetry(clientID string, msg *SkillPushMessage) error {
    var lastErr error
    for i := 0; i < p.config.MaxRetries; i++ {
        err := p.push(clientID, msg)
        if err == nil {
            return nil
        }
        lastErr = err
        time.Sleep(p.config.RetryInterval)
    }
    return lastErr
}
```

## 6. 实现清单

### 6.1 文件创建/修改

| 文件 | 操作 | 说明 |
|------|------|------|
| `pkg/common/eventhub/types.go` | 修改 | 添加 Skill 相关 Topic |
| `pkg/skill/event/types.go` | 创建 | Skill 事件类型定义 |
| `pkg/skill/origin_server.go` | 修改 | 添加事件发布逻辑 |
| `pkg/cache/ai/skill.go` | 修改 | 订阅事件触发缓存更新 |

### 6.2 验收标准

1. Skill 创建/更新/删除时发布事件
2. 订阅者能收到变更通知
3. 缓存层能响应事件更新
4. 推送失败时有重试机制
