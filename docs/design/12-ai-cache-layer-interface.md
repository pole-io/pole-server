# AI Native 缓存层接口设计

## 1. 概述

本文档定义 AI Native 功能模块的缓存层接口规范，包括 SkillCache、MCPServerCache、SkillVersionCache 和 SkillSubscriptionCache。

## 2. 设计原则

1. **遵循现有模式**: 参考 `pkg/cache/rules/` 和 `pkg/cache/service/` 的实现模式
2. **增量更新**: 使用 mtime 时间戳进行增量数据同步
3. **SingleFlight**: 防止缓存击穿，避免并发重复更新
4. **BaseCache 组合**: 复用 `cachebase.BaseCache` 的通用能力

## 3. 缓存索引定义

### 3.1 缓存名称常量

```go
// apis/cache/types.go 新增

const (
    // 现有缓存名称...

    // AI Native 缓存
    CacheSkill            CacheName = "skill"
    CacheMCPServer        CacheName = "mcpServer"
    CacheSkillVersion     CacheName = "skillVersion"
    CacheSkillSubscription CacheName = "skillSubscription"
)

const (
    // 缓存索引
    DefaultIndexSkill             = 100
    DefaultIndexMCPServer         = 101
    DefaultIndexSkillVersion      = 102
    DefaultIndexSkillSubscription = 103
)
```

## 4. SkillCache 接口设计

### 4.1 接口定义

```go
// apis/cache/ai.go (新文件)

package cache

import (
    "context"
    "github.com/pole-io/pole-control-plane/apis/pkg/types/ai"
)

// SkillCache Skill 缓存接口
type SkillCache interface {
    Cache

    // GetSkillByID 根据 ID 获取 Skill
    GetSkillByID(id string) *ai.Skill

    // GetSkillByName 根据名称和命名空间获取 Skill
    GetSkillByName(name, namespace string) *ai.Skill

    // GetSkillsByNamespace 获取命名空间下的所有 Skill
    GetSkillsByNamespace(namespace string) []*ai.Skill

    // GetSkillsByCategory 获取指定分类的 Skill
    GetSkillsByCategory(category string) []*ai.Skill

    // GetSkillsByTag 根据 Tag 获取 Skill
    GetSkillsByTag(tag string) []*ai.Skill

    // QuerySkills 条件查询 Skill
    QuerySkills(ctx context.Context, filter map[string]string, offset, limit uint32) (
        uint32, []*ai.Skill)
}
```

### 4.2 实现结构

```go
// pkg/cache/ai/skill.go (新文件)

package ai

import (
    "context"
    "sync"
    "time"

    "github.com/pole-io/pole-control-plane/apis/cache"
    "github.com/pole-io/pole-control-plane/apis/pkg/types/ai"
    "github.com/pole-io/pole-control-plane/apis/store"
    "github.com/pole-io/pole-control-plane/pkg/cache/cachebase"
    "github.com/pole-io/pole-control-plane/pkg/common/container"
    "golang.org/x/sync/singleflight"
)

type skillCache struct {
    *cachebase.BaseCache
    storage store.Store

    // 数据存储
    // id -> *ai.Skill
    ids *container.SyncMap[string, *ai.Skill]

    // namespace/name -> *ai.Skill
    names *container.SyncMap[string, *ai.Skill]

    // namespace -> []*ai.Skill
    namespaceIndex *container.SyncMap[string, []*ai.Skill]

    // category -> []*ai.Skill
    categoryIndex *container.SyncMap[string, []*ai.Skill]

    // tag -> []*ai.Skill
    tagIndex *container.SyncMap[string, []*ai.Skill]

    singleFlight *singleflight.Group
}

func NewSkillCache(storage store.Store) cache.Cache {
    return &skillCache{
        BaseCache:      &cachebase.BaseCache{},
        storage:        storage,
        ids:            container.NewSyncMap[string, *ai.Skill](),
        names:          container.NewSyncMap[string, *ai.Skill](),
        namespaceIndex: container.NewSyncMap[string, []*ai.Skill](),
        categoryIndex:  container.NewSyncMap[string, []*ai.Skill](),
        tagIndex:       container.NewSyncMap[string, []*ai.Skill](),
        singleFlight:   &singleflight.Group{},
    }
}

func (sc *skillCache) Name() string {
    return string(cache.CacheSkill)
}

func (sc *skillCache) Initialize(c map[string]interface{}) error {
    return nil
}

func (sc *skillCache) Update() error {
    _, err, _ := sc.singleFlight.Do(sc.Name(), func() (interface{}, error) {
        return nil, sc.DoCacheUpdate(sc.Name(), sc.realUpdate)
    })
    return err
}

func (sc *skillCache) realUpdate() (map[string]time.Time, int64, error) {
    results := make(map[string]time.Time)

    skills, err := sc.storage.GetMoreSkills(sc.LastFetchTime(), sc.IsFirstUpdate())
    if err != nil {
        return nil, 0, err
    }

    var reload bool
    for _, skill := range skills {
        if skill.ModifyTime.After(sc.LastFetchTime()) {
            sc.LastFetchTime().Add(time.Nanosecond)
        }

        key := skillKey(skill)
        results[key] = skill.ModifyTime

        if skill.IsDeleted() {
            sc.removeSkill(skill)
            reload = true
        }} else {
            sc.storeSkill(skill)
        }
    }

    return results, int64(len(skills)), nil
}

func (sc *skillCache) Clear() error {
    sc.ids.Clear()
    sc.names.Clear()
    sc.namespaceIndex.Clear()
    sc.categoryIndex.Clear()
    sc.tagIndex.Clear()
    return nil
}

func (sc *skillCache) Close() error {
    return nil
}

// GetSkillByID 实现 SkillCache 接口
func (sc *skillCache) GetSkillByID(id string) *ai.Skill {
    return sc.ids.Get(id)
}

// GetSkillByName 实现 SkillCache 接口
func (sc *skillCache) GetSkillByName(name, namespace string) *ai.Skill {
    key := skillNameKey(namespace, name)
    return sc.names.Get(key)
}

// GetSkillsByNamespace 实现 SkillCache 接口
func (sc *skillCache) GetSkillsByNamespace(namespace string) []*ai.Skill {
    return sc.namespaceIndex.Get(namespace)
}

// GetSkillsByCategory 实现 SkillCache 接口
func (sc *skillCache) GetSkillsByCategory(category string) []*ai.Skill {
    return sc.categoryIndex.Get(category)
}

// GetSkillsByTag 实现 SkillCache 接口
func (sc *skillCache) GetSkillsByTag(tag string) []*ai.Skill {
    return sc.tagIndex.Get(tag)
}

// QuerySkills 实现 SkillCache 接口
func (sc *skillCache) QuerySkills(ctx context.Context, filter map[string]string,
    offset, limit uint32) (uint32, []*ai.Skill) {
    // 实现分页查询逻辑
    // ...
}

// 辅助方法
func skillKey(skill *ai.Skill) string {
    return skill.Id
}

func skillNameKey(namespace, name string) string {
    return namespace + "/" + name
}

func (sc *skillCache) storeSkill(skill *ai.Skill) {
    // 存储 ID 索引
    sc.ids.Store(skill.Id, skill)

    // 存储名称索引
    key := skillNameKey(skill.Namespace, skill.Name)
    sc.names.Store(key, skill)

    // 更新命名空间索引
    sc.updateNamespaceIndex(skill)

    // 更新分类索引
    sc.updateCategoryIndex(skill)

    // 更新标签索引
    sc.updateTagIndex(skill)
}

func (sc *skillCache) removeSkill(skill *ai.Skill) {
    // 删除 ID 索引
    sc.ids.Delete(skill.Id)

    // 删除名称索引
    key := skillNameKey(skill.Namespace, skill.Name)
    sc.names.Delete(key)

    // 从命名空间索引中移除
    sc.removeFromNamespaceIndex(skill)

    // 从分类索引中移除
    sc.removeFromCategoryIndex(skill)

    // 从标签索引中移除
    sc.removeFromTagIndex(skill)
}
```

## 5. MCPServerCache 接口设计

### 5.1 接口定义

```go
// apis/cache/ai.go (续)

// MCPServerCache MCP Server 缓存接口
type MCPServerCache interface {
    Cache

    // GetMCPServerByID 根据 ID 获取 MCP Server
    GetMCPServerByID(id string) *ai.MCPServer

    // GetMCPServerByName 根据名称和命名空间获取 MCP Server
    GetMCPServerByName(name, namespace string) *ai.MCPServer

    // GetMCPServersByNamespace 获取命名空间下的所有 MCP Server
    GetMCPServersByNamespace(namespace string) []*ai.MCPServer

    // GetMCPServerTools 获取 MCP Server 的所有工具
    GetMCPServerTools(serverID string) []*ai.MCPServerTool

    // QueryMCPServers 条件查询 MCP Server
    QueryMCPServers(ctx context.Context, filter map[string]string,
        offset, limit uint32) (uint32, []*ai.MCPServer)
}
```

### 5.2 实现结构

```go
// pkg/cache/ai/mcp_server.go (新文件)

package ai

import (
    "context"
    "time"

    "github.com/pole-io/pole-control-plane/apis/cache"
    "github.com/pole-io/pole-control-plane/apis/pkg/types/ai"
    "github.com/pole-io/pole-control-plane/apis/store"
    "github.com/pole-io/pole-control-plane/pkg/cache/cachebase"
    "github.com/pole-io/pole-control-plane/pkg/common/container"
    "golang.org/x/sync/singleflight"
)

type mcpServerCache struct {
    *cachebase.BaseCache
    storage store.Store

    // id -> *ai.MCPServer
    ids *container.SyncMap[string, *ai.MCPServer]

    // namespace/name -> *ai.MCPServer
    names *container.SyncMap[string, *ai.MCPServer]

    // namespace -> []*ai.MCPServer
    namespaceIndex *container.SyncMap[string, []*ai.MCPServer]

    // serverID -> []*ai.MCPServerTool
    tools *container.SyncMap[string, []*ai.MCPServerTool]

    singleFlight *singleflight.Group
}

func NewMCPServerCache(storage store.Store) cache.Cache {
    return &mcpServerCache{
        BaseCache:      &cachebase.BaseCache{},
        storage:        storage,
        ids:            container.NewSyncMap[string, *ai.MCPServer](),
        names:          container.NewSyncMap[string, *ai.MCPServer](),
        namespaceIndex: container.NewSyncMap[string, []*ai.MCPServer](),
        tools:          container.NewSyncMap[string, []*ai.MCPServerTool](),
        singleFlight:   &singleflight.Group{},
    }
}

func (mc *mcpServerCache) Name() string {
    return string(cache.CacheMCPServer)
}

func (mc *mcpServerCache) Update() error {
    _, err, _ := mc.singleFlight.Do(mc.Name(), func() (interface{}, error) {
        return nil, mc.DoCacheUpdate(mc.Name(), mc.realUpdate)
    })
    return err
}

func (mc *mcpServerCache) realUpdate() (map[string]time.Time, int64, error) {
    results := make(map[string]time.Time)

    // 更新 MCP Server 数据
    servers, err := mc.storage.GetMoreMCPServers(mc.LastFetchTime(), mc.IsFirstUpdate())
    if err != nil {
        return nil, 0, err
    }

    for _, server := range servers {
        if server.ModifyTime.After(mc.LastFetchTime()) {
            mc.LastFetchTime().Add(time.Nanosecond)
        }

        results[server.Id] = server.ModifyTime

        if server.IsDeleted() {
            mc.removeMCPServer(server)
        } else {
            mc.storeMCPServer(server)
        }
    }

    return results, int64(len(servers)), nil
}

// 接口实现方法...
```

## 6. SkillVersionCache 接口设计

### 6.1 接口定义

```go
// apis/cache/ai.go (续)

// SkillVersionCache Skill 版本缓存接口
type SkillVersionCache interface {
    Cache

    // GetSkillVersionByID 根据 ID 获取 SkillVersion
    GetSkillVersionByID(id string) *ai.SkillVersion

    // GetSkillVersionByVersion 根据技能名、命名空间和版本号获取
    GetSkillVersionByVersion(skillName, namespace string, version uint64) *ai.SkillVersion

    // GetActiveSkillVersion 获取活跃版本
    GetActiveSkillVersion(skillName, namespace string) *ai.SkillVersion

    // GetSkillVersions 获取技能的所有版本
    GetSkillVersions(skillName, namespace string) []*ai.SkillVersion
}
```

### 6.2 实现结构

```go
// pkg/cache/ai/skill_version.go (新文件)

package ai

import (
    "time"

    "github.com/pole-io/pole-control-plane/apis/cache"
    "github.com/pole-io/pole-control-plane/apis/pkg/types/ai"
    "github.com/pole-io/pole-control-plane/apis/store"
    "github.com/pole-io/pole-control-plane/pkg/cache/cachebase"
    "github.com/pole-io/pole-control-plane/pkg/common/container"
    "golang.org/x/sync/singleflight"
)

type skillVersionCache struct {
    *cachebase.BaseCache
    storage store.Store

    // id -> *ai.SkillVersion
    ids *container.SyncMap[string, *ai.SkillVersion]

    // namespace/skillName/version -> *ai.SkillVersion
    versionKeys *container.SyncMap[string, *ai.SkillVersion]

    // namespace/skillName -> *ai.SkillVersion (active version)
    activeVersions *container.SyncMap[string, *ai.SkillVersion]

    // namespace/skillName -> []*ai.SkillVersion
    skillVersions *container.SyncMap[string, []*ai.SkillVersion]

    singleFlight *singleflight.Group
}

func NewSkillVersionCache(storage store.Store) cache.Cache {
    return &skillVersionCache{
        BaseCache:      &cachebase.BaseCache{},
        storage:        storage,
        ids:            container.NewSyncMap[string, *ai.SkillVersion](),
        versionKeys:    container.NewSyncMap[string, *ai.SkillVersion](),
        activeVersions: container.NewSyncMap[string, *ai.SkillVersion](),
        skillVersions:  container.NewSyncMap[string, []*ai.SkillVersion](),
        singleFlight:   &singleflight.Group{},
    }
}

func (svc *skillVersionCache) Name() string {
    return string(cache.CacheSkillVersion)
}
```

## 7. SkillSubscriptionCache 接口设计

### 7.1 接口定义

```go
// apis/cache/ai.go (续)

// SkillSubscriptionCache Skill 订阅缓存接口
type SkillSubscriptionCache interface {
    Cache

    // GetSubscriptionByID 根据 ID 获取订阅
    GetSubscriptionByID(id string) *ai.SkillSubscription

    // GetSubscriptionsByClient 获取客户端的所有订阅
    GetSubscriptionsByClient(clientID string) []*ai.SkillSubscription

    // GetSubscriptionsBySkill 获取技能的所有订阅
    GetSubscriptionsBySkill(skillName, namespace string) []*ai.SkillSubscription
}
```

### 7.2 实现结构

```go
// pkg/cache/ai/skill_subscription.go (新文件)

package ai

import (
    "time"

    "github.com/pole-io/pole-control-plane/apis/cache"
    "github.com/pole-io/pole-control-plane/apis/pkg/types/ai"
    "github.com/pole-io/pole-control-plane/apis/store"
    "github.com/pole-io/pole-control-plane/pkg/cache/cachebase"
    "github.com/pole-io/pole-control-plane/pkg/common/container"
    "golang.org/x/sync/singleflight"
)

type skillSubscriptionCache struct {
    *cachebase.BaseCache
    storage store.Store

    // id -> *ai.SkillSubscription
    ids *container.SyncMap[string, *ai.SkillSubscription]

    // clientID -> []*ai.SkillSubscription
    clientSubscriptions *container.SyncMap[string, []*ai.SkillSubscription]

    // namespace/skillName -> []*ai.SkillSubscription
    skillSubscriptions *container.SyncMap[string, []*ai.SkillSubscription]

    singleFlight *singleflight.Group
}

func NewSkillSubscriptionCache(storage store.Store) cache.Cache {
    return &skillSubscriptionCache{
        BaseCache:           &cachebase.BaseCache{},
        storage:             storage,
        ids:                container.NewSyncMap[string, *ai.SkillSubscription](),
        clientSubscriptions: container.NewSyncMap[string, []*ai.SkillSubscription](),
        skillSubscriptions:  container.NewSyncMap[string, []*ai.SkillSubscription](),
        singleFlight:        &singleflight.Group{},
    }
}

func (ssc *skillSubscriptionCache) Name() string {
    return string(cache.CacheSkillSubscription)
}
```

## 8. 缓存注册

### 8.1 注册到 CacheManager

```go
// pkg/cache/default.go (修改)

func init() {
    // 现有注册...

    // AI Native 缓存注册
    RegisterCache(cacheapi.CacheSkill, cacheapi.DefaultIndexSkill)
    RegisterCache(cacheapi.CacheMCPServer, cacheapi.DefaultIndexMCPServer)
    RegisterCache(cacheapi.CacheSkillVersion, cacheapi.DefaultIndexSkillVersion)
    RegisterCache(cacheapi.CacheSkillSubscription, cacheapi.DefaultIndexSkillSubscription)
}
```

### 8.2 新建缓存文件

```go
// pkg/cache/ai/default.go (新文件)

package ai

import (
    "github.com/pole-io/pole-control-plane/apis/cache"
    "github.com/pole-io/pole-control-plane/apis/store"
)

// NewAICaches 创建所有 AI Native 缓存
func NewAICaches(storage store.Store) []cache.Cache {
    return []cache.Cache{
        NewSkillCache(storage),
        NewMCPServerCache(storage),
        NewSkillVersionCache(storage),
        NewSkillSubscriptionCache(storage),
    }
}
```

## 9. 测试用例设计要求

### 9.1 单元测试覆盖

测试用例应覆盖以下场景：

1. **基本 CRUD 操作**
   - 创建缓存实例
   - 数据添加和查询
   - 数据更新和删除

2. **增量更新测试**
   - 首次全量更新
   - 增量更新（基于 mtime）
   - 删除标记处理

3. **索引测试**
   - ID 索引查询
   - 名称索引查询
   - 复合索引查询

4. **并发测试**
   - SingleFlight 防重入
   - 并发读写安全

5. **边界条件**
   - 空数据处理
   - 大批量数据处理
   - 无效参数处理

### 9.2 测试文件位置

```
pkg/cache/ai/
├── skill.go
├── skill_test.go
├── mcp_server.go
├── mcp_server_test.go
├── skill_version.go
├── skill_version_test.go
├── skill_subscription.go
└── skill_subscription_test.go
```

## 10. 目录结构

```
pkg/cache/ai/
├── default.go              # 缓存工厂
├── skill.go                # Skill 缓存实现
├── skill_test.go           # Skill 缓存测试
├── mcp_server.go           # MCP Server 缓存实现
├── mcp_server_test.go      # MCP Server 缓存测试
├── skill_version.go        # Skill 版本缓存实现
├── skill_version_test.go   # Skill 版本缓存测试
├── skill_subscription.go   # Skill 订阅缓存实现
└── skill_subscription_test.go  # Skill 订阅缓存测试

apis/cache/
└── ai.go                   # AI 缓存接口定义
```

## 11. 实现依赖

- `github.com/pole-io/pole-control-plane/apis/cache` - 缓存接口
- `github.com/pole-io/pole-control-plane/apis/store` - 存储接口
- `github.com/pole-io/pole-control-plane/apis/pkg/types/ai` - AI 类型定义
- `github.com/pole-io/pole-control-plane/pkg/cache/cachebase` - 缓存基础类
- `github.com/pole-io/pole-control-plane/pkg/common/container` - 并发容器
- `golang.org/x/sync/singleflight` - SingleFlight 组件
