# Skill 拦截器接口设计

## 1. 概述

本文档定义 Skill 模块的拦截器接口规范，包括认证拦截器 (auth) 和参数校验拦截器 (paramcheck)。拦截器采用装饰器模式，对 SkillServer 进行增强。

## 2. 设计原则

1. **装饰器模式**: 拦截器包装原服务器，在调用前后执行额外逻辑
2. **链式调用**: 支持多个拦截器串联
3. **职责单一**: auth 拦截器负责权限校验，paramcheck 拦截器负责参数验证
4. **参考现有实现**: 遵循 `pkg/service/interceptor/auth/` 的设计模式

## 3. 目录结构

```
pkg/skill/
├── skill.go                          # 原始 SkillServer 接口
├── interceptor/
│   ├── auth/
│   │   ├── server.go                 # 认证拦截器主结构
│   │   ├── skill.go                  # Skill 认证逻辑
│   │   ├── skill_group.go            # SkillGroup 认证逻辑
│   │   ├── skill_version.go          # SkillVersion 认证逻辑
│   │   ├── skill_subscription.go     # SkillSubscription 认证逻辑
│   │   └── log.go                    # 日志
│   └── paramcheck/
│       ├── server.go                 # 参数校验拦截器主结构
│       ├── skill.go                  # Skill 参数校验
│       ├── skill_group.go            # SkillGroup 参数校验
│       ├── skill_version.go          # SkillVersion 参数校验
│       └── skill_subscription.go     # SkillSubscription 参数校验
```

## 4. 认证拦截器设计

### 4.1 Server 结构

```go
// pkg/skill/interceptor/auth/server.go

package skill_auth

import (
    "github.com/pole-io/pole-server/apis/access_control/auth"
    cacheapi "github.com/pole-io/pole-server/apis/cache"
    authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
    "github.com/pole-io/pole-server/pkg/skill"
)

// Server 带有鉴权能力的 SkillServer
type Server struct {
    nextSvr   skill.SkillServer
    userSvr   auth.UserServer
    policySvr auth.StrategyServer
    cacheMgr  cacheapi.CacheManager
}

// NewServer 创建带认证的 SkillServer
func NewServer(nextSvr skill.SkillServer,
    userSvr auth.UserServer, policySvr auth.StrategyServer) skill.SkillServer {
    return &Server{
        nextSvr:   nextSvr,
        userSvr:   userSvr,
        policySvr: policySvr,
    }
}

// Cache 获取缓存管理器
func (svr *Server) Cache() cacheapi.CacheManager {
    return svr.cacheMgr
}
```

### 4.2 认证上下文收集

```go
// pkg/skill/interceptor/auth/server.go

// collectSkillAuthContext 收集 Skill 认证上下文
func (svr *Server) collectSkillAuthContext(ctx context.Context,
    skill *aiTypes.Skill, resourceOp authtypes.ResourceOperation,
    methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
    return authtypes.NewAcquireContext(
        authtypes.WithRequestContext(ctx),
        authtypes.WithOperation(resourceOp),
        authtypes.WithModule(authtypes.SkillModule),
        authtypes.WithMethod(methodName),
        authtypes.WithAccessResources(svr.querySkillResource(skill)),
    )
}

// querySkillResource 查询 Skill 资源信息
func (svr *Server) querySkillResource(skill *aiTypes.Skill) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
    if skill == nil {
        return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
    }

    nsRet := make([]authtypes.ResourceEntry, 0)
    if skill.Namespace != "" {
        ns := svr.Cache().Namespace().GetNamespace(skill.Namespace)
        if ns != nil {
            nsRet = append(nsRet, authtypes.ResourceEntry{
                Type:     apisecurity.ResourceType_Namespaces,
                ID:       ns.Name,
                Owner:    ns.Owner,
                Metadata: ns.Metadata,
            })
        }
    }

    skillRet := make([]authtypes.ResourceEntry, 0)
    if skill.ID != "" {
        skillRet = append(skillRet, authtypes.ResourceEntry{
            Type:     apisecurity.ResourceType_Skills,
            ID:       skill.ID,
            Owner:    skill.Owner,
            Metadata: skill.Metadata,
        })
    }

    return map[apisecurity.ResourceType][]authtypes.ResourceEntry{
        apisecurity.ResourceType_Namespaces: nsRet,
        apisecurity.ResourceType_Skills:     skillRet,
    }
}
```

### 4.3 Skill 操作认证

```go
// pkg/skill/interceptor/auth/skill.go

package skill_auth

import (
    "context"

    apimodel "github.com/pole-io/specification/source/go/api/v1/model"

    aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
    authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
    api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateSkill 创建 Skill（带认证）
func (svr *Server) CreateSkill(ctx context.Context, skill *aiTypes.Skill) error {
    authCtx := svr.collectSkillAuthContext(ctx, skill, authtypes.Create, authtypes.CreateSkill)

    _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
    if err != nil {
        return err
    }

    ctx = authCtx.GetRequestContext()
    ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

    // 填充 owner 信息
    ownerID := utils.ParseOwnerID(ctx)
    if len(ownerID) > 0 {
        skill.Owner = ownerID
    }

    return svr.nextSvr.CreateSkill(ctx, skill)
}

// UpdateSkill 更新 Skill（带认证）
func (svr *Server) UpdateSkill(ctx context.Context, skill *aiTypes.Skill) error {
    authCtx := svr.collectSkillAuthContext(ctx, skill, authtypes.Modify, authtypes.UpdateSkill)

    // 移除命名空间资源检查，只检查 Skill 本身
    accessRes := authCtx.GetAccessResources()
    delete(accessRes, apisecurity.ResourceType_Namespaces)
    authCtx.SetAccessResources(accessRes)

    _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
    if err != nil {
        return err
    }

    ctx = authCtx.GetRequestContext()
    ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

    return svr.nextSvr.UpdateSkill(ctx, skill)
}

// DeleteSkill 删除 Skill（带认证）
func (svr *Server) DeleteSkill(ctx context.Context, id string) error {
    // 先获取 Skill 信息用于权限检查
    skill, err := svr.nextSvr.GetSkill(ctx, id)
    if err != nil {
        return err
    }

    authCtx := svr.collectSkillAuthContext(ctx, skill, authtypes.Delete, authtypes.DeleteSkill)

    // 移除命名空间资源检查
    accessRes := authCtx.GetAccessResources()
    delete(accessRes, apisecurity.ResourceType_Namespaces)
    authCtx.SetAccessResources(accessRes)

    _, err = svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
    if err != nil {
        return err
    }

    ctx = authCtx.GetRequestContext()
    ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

    return svr.nextSvr.DeleteSkill(ctx, id)
}

// GetSkill 获取 Skill（带认证）
func (svr *Server) GetSkill(ctx context.Context, id string) (*aiTypes.Skill, error) {
    authCtx := svr.collectSkillAuthContext(ctx, nil, authtypes.Read, authtypes.GetSkill)

    _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
    if err != nil {
        return nil, err
    }

    ctx = authCtx.GetRequestContext()
    ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

    return svr.nextSvr.GetSkill(ctx, id)
}

// GetSkillByName 根据名称获取 Skill（带认证）
func (svr *Server) GetSkillByName(ctx context.Context, name, namespace string) (*aiTypes.Skill, error) {
    authCtx := svr.collectSkillAuthContext(ctx, &aiTypes.Skill{
        Name:      name,
        Namespace: namespace,
    }, authtypes.Read, authtypes.GetSkillByName)

    _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
    if err != nil {
        return nil, err
    }

    ctx = authCtx.GetRequestContext()
    ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

    return svr.nextSvr.GetSkillByName(ctx, name, namespace)
}
```

### 4.4 资源类型定义

需要在 `apisecurity.ResourceType` 中添加 Skill 相关资源类型：

```go
// 在 apis/pkg/types/auth/resource.go 中添加

const (
    // 现有资源类型...

    // Skill 相关资源类型
    ResourceType_Skills            ResourceType = "skills"
    ResourceType_SkillGroups       ResourceType = "skill_groups"
    ResourceType_SkillVersions     ResourceType = "skill_versions"
    ResourceType_SkillSubscriptions ResourceType = "skill_subscriptions"
)
```

### 4.5 模块和方法定义

需要在 `authtypes` 中添加 Skill 模块相关常量：

```go
// 在 apis/pkg/types/auth/module.go 中添加

const (
    // SkillModule Skill 模块
    SkillModule Module = "skill"
)

const (
    // Skill 操作方法
    CreateSkill            ServerFunctionName = "CreateSkill"
    UpdateSkill            ServerFunctionName = "UpdateSkill"
    DeleteSkill            ServerFunctionName = "DeleteSkill"
    GetSkill               ServerFunctionName = "GetSkill"
    GetSkillByName         ServerFunctionName = "GetSkillByName"

    // SkillGroup 操作方法
    CreateSkillGroup       ServerFunctionName = "CreateSkillGroup"
    UpdateSkillGroup       ServerFunctionName = "UpdateSkillGroup"
    DeleteSkillGroup       ServerFunctionName = "DeleteSkillGroup"
    GetSkillGroup          ServerFunctionName = "GetSkillGroup"

    // SkillVersion 操作方法
    CreateSkillVersion     ServerFunctionName = "CreateSkillVersion"
    ActivateSkillVersion   ServerFunctionName = "ActivateSkillVersion"

    // SkillSubscription 操作方法
    CreateSkillSubscription    ServerFunctionName = "CreateSkillSubscription"
    DeleteSkillSubscription    ServerFunctionName = "DeleteSkillSubscription"
    GetSkillSubscriptionsBySkill ServerFunctionName = "GetSkillSubscriptionsBySkill"
    GetSkillSubscriptionsByClient ServerFunctionName = "GetSkillSubscriptionsByClient"
)
```

## 5. 参数校验拦截器设计

### 5.1 Server 结构

```go
// pkg/skill/interceptor/paramcheck/server.go

package paramcheck

import (
    "github.com/pole-io/pole-server/apis/store"
    "github.com/pole-io/pole-server/pkg/skill"
)

// Server 带参数校验的 SkillServer
type Server struct {
    storage store.Store
    nextSvr skill.SkillServer
}

// NewServer 创建带参数校验的 SkillServer
func NewServer(nextSvr skill.SkillServer) skill.SkillServer {
    return &Server{
        nextSvr: nextSvr,
    }
}
```

### 5.2 Skill 参数校验

```go
// pkg/skill/interceptor/paramcheck/skill.go

package paramcheck

import (
    "context"
    "errors"
    "fmt"

    aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

var (
    ErrSkillNameEmpty      = errors.New("skill name cannot be empty")
    ErrSkillNamespaceEmpty = errors.New("skill namespace cannot be empty")
    ErrSkillTypeEmpty      = errors.New("skill type cannot be empty")
    ErrSkillIDEmptry       = errors.New("skill id cannot be empty")
)

// CreateSkill 创建 Skill（带参数校验）
func (svr *Server) CreateSkill(ctx context.Context, skill *aiTypes.Skill) error {
    if err := svr.checkCreateSkill(skill); err != nil {
        return err
    }
    return svr.nextSvr.CreateSkill(ctx, skill)
}

// checkCreateSkill 校验创建 Skill 的参数
func (svr *Server) checkCreateSkill(skill *aiTypes.Skill) error {
    if skill == nil {
        return errors.New("skill cannot be nil")
    }

    if skill.Name == "" {
        return ErrSkillNameEmpty
    }

    if skill.Namespace == "" {
        return ErrSkillNamespaceEmpty
    }

    if skill.SkillType == "" {
        return ErrSkillTypeEmpty
    }

    // 校验名称格式
    if err := valid.CheckName(skill.Name); err != nil {
        return fmt.Errorf("invalid skill name: %w", err)
    }

    // 校验命名空间是否存在
    if !svr.namespaceExists(skill.Namespace) {
        return fmt.Errorf("namespace %s not found", skill.Namespace)
    }

    return nil
}

// UpdateSkill 更新 Skill（带参数校验）
func (svr *Server) UpdateSkill(ctx context.Context, skill *aiTypes.Skill) error {
    if err := svr.checkUpdateSkill(skill); err != nil {
        return err
    }
    return svr.nextSvr.UpdateSkill(ctx, skill)
}

// checkUpdateSkill 校验更新 Skill 的参数
func (svr *Server) checkUpdateSkill(skill *aiTypes.Skill) error {
    if skill == nil {
        return errors.New("skill cannot be nil")
    }

    if skill.ID == "" {
        return ErrSkillIDEmptry
    }

    return nil
}

// DeleteSkill 删除 Skill（带参数校验）
func (svr *Server) DeleteSkill(ctx context.Context, id string) error {
    if id == "" {
        return ErrSkillIDEmptry
    }
    return svr.nextSvr.DeleteSkill(ctx, id)
}

// namespaceExists 检查命名空间是否存在
func (svr *Server) namespaceExists(namespace string) bool {
    // 通过缓存或存储检查命名空间是否存在
    return true // TODO: 实际实现
}
```

### 5.3 SkillGroup 参数校验

```go
// pkg/skill/interceptor/paramcheck/skill_group.go

package paramcheck

import (
    "context"
    "errors"

    aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

var (
    ErrSkillGroupNameEmpty      = errors.New("skill group name cannot be empty")
    ErrSkillGroupNamespaceEmpty = errors.New("skill group namespace cannot be empty")
)

// CreateSkillGroup 创建 SkillGroup（带参数校验）
func (svr *Server) CreateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
    if err := svr.checkCreateSkillGroup(group); err != nil {
        return err
    }
    return svr.nextSvr.CreateSkillGroup(ctx, group)
}

func (svr *Server) checkCreateSkillGroup(group *aiTypes.SkillGroup) error {
    if group == nil {
        return errors.New("skill group cannot be nil")
    }

    if group.Name == "" {
        return ErrSkillGroupNameEmpty
    }

    if group.Namespace == "" {
        return ErrSkillGroupNamespaceEmpty
    }

    // 校验引用的 Skill 是否存在
    for _, ref := range group.Skills {
        if ref.SkillID == "" && ref.SkillName == "" {
            return errors.New("skill reference must have either ID or Name")
        }
    }

    return nil
}
```

### 5.4 SkillVersion 参数校验

```go
// pkg/skill/interceptor/paramcheck/skill_version.go

package paramcheck

import (
    "context"
    "errors"

    aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

var (
    ErrSkillVersionSkillIDEmpty   = errors.New("skill version skill_id cannot be empty")
    ErrSkillVersionSkillNameEmpty = errors.New("skill version skill_name cannot be empty")
    ErrSkillVersionInvalid        = errors.New("skill version must be greater than 0")
)

// CreateSkillVersion 创建 SkillVersion（带参数校验）
func (svr *Server) CreateSkillVersion(ctx context.Context, version *aiTypes.SkillVersion) error {
    if err := svr.checkCreateSkillVersion(version); err != nil {
        return err
    }
    return svr.nextSvr.CreateSkillVersion(ctx, version)
}

func (svr *Server) checkCreateSkillVersion(version *aiTypes.SkillVersion) error {
    if version == nil {
        return errors.New("skill version cannot be nil")
    }

    if version.SkillID == "" {
        return ErrSkillVersionSkillIDEmpty
    }

    if version.SkillName == "" {
        return ErrSkillVersionSkillNameEmpty
    }

    if version.Version <= 0 {
        return ErrSkillVersionInvalid
    }

    return nil
}
```

### 5.5 SkillSubscription 参数校验

```go
// pkg/skill/interceptor/paramcheck/skill_subscription.go

package paramcheck

import (
    "context"
    "errors"

    aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

var (
    ErrSubscriptionSkillNameEmpty = errors.New("subscription skill_name cannot be empty")
    ErrSubscriptionNamespaceEmpty = errors.New("subscription namespace cannot be empty")
    ErrSubscriptionClientIDEmpty  = errors.New("subscription client_id cannot be empty")
)

// CreateSkillSubscription 创建 SkillSubscription（带参数校验）
func (svr *Server) CreateSkillSubscription(ctx context.Context, sub *aiTypes.SkillSubscription) error {
    if err := svr.checkCreateSkillSubscription(sub); err != nil {
        return err
    }
    return svr.nextSvr.CreateSkillSubscription(ctx, sub)
}

func (svr *Server) checkCreateSkillSubscription(sub *aiTypes.SkillSubscription) error {
    if sub == nil {
        return errors.New("subscription cannot be nil")
    }

    if sub.SkillName == "" {
        return ErrSubscriptionSkillNameEmpty
    }

    if sub.Namespace == "" {
        return ErrSubscriptionNamespaceEmpty
    }

    if sub.ClientID == "" {
        return ErrSubscriptionClientIDEmpty
    }

    return nil
}
```

## 6. 拦截器链组装

### 6.1 初始化函数

```go
// pkg/skill/interceptor/initialize.go

package interceptor

import (
    "github.com/pole-io/pole-server/apis/access_control/auth"
    "github.com/pole-io/pole-server/pkg/skill"
    "github.com/pole-io/pole-server/pkg/skill/interceptor/auth"
    "github.com/pole-io/pole-server/pkg/skill/interceptor/paramcheck"
)

// Initialize 初始化拦截器链
// 拦截器顺序: paramcheck -> auth -> origin server
func Initialize(originSvr skill.SkillServer, userSvr auth.UserServer, policySvr auth.StrategyServer) skill.SkillServer {
    // 先添加认证拦截器
    authSvr := skill_auth.NewServer(originSvr, userSvr, policySvr)

    // 再添加参数校验拦截器（最外层）
    paramCheckSvr := paramcheck.NewServer(authSvr)

    return paramCheckSvr
}
```

## 7. 测试用例设计要求

### 7.1 认证拦截器测试

```go
// pkg/skill/interceptor/auth/skill_test.go

// 测试场景：
// 1. TestCreateSkill_WithPermission - 有权限创建
// 2. TestCreateSkill_WithoutPermission - 无权限创建
// 3. TestUpdateSkill_WithPermission - 有权限更新
// 4. TestUpdateSkill_WithoutPermission - 无权限更新
// 5. TestDeleteSkill_WithPermission - 有权限删除
// 6. TestDeleteSkill_WithoutPermission - 无权限删除
// 7. TestGetSkill_WithPermission - 有权限读取
// 8. TestGetSkill_WithoutPermission - 无权限读取
```

### 7.2 参数校验拦截器测试

```go
// pkg/skill/interceptor/paramcheck/skill_test.go

// 测试场景：
// 1. TestCreateSkill_ValidParams - 有效参数
// 2. TestCreateSkill_EmptyName - 名称为空
// 3. TestCreateSkill_EmptyNamespace - 命名空间为空
// 4. TestCreateSkill_EmptyType - 类型为空
// 5. TestCreateSkill_InvalidName - 无效名称格式
// 6. TestUpdateSkill_EmptyID - ID 为空
// 7. TestDeleteSkill_EmptyID - ID 为空
```

## 8. 实现优先级

1. **P0 - 认证拦截器**
   - `pkg/skill/interceptor/auth/server.go`
   - `pkg/skill/interceptor/auth/skill.go`

2. **P0 - 参数校验拦截器**
   - `pkg/skill/interceptor/paramcheck/server.go`
   - `pkg/skill/interceptor/paramcheck/skill.go`

3. **P1 - 扩展拦截器**
   - SkillGroup 拦截器
   - SkillVersion 拦截器
   - SkillSubscription 拦截器

## 9. 实现依赖

- `github.com/pole-io/pole-server/apis/access_control/auth` - 认证接口
- `github.com/pole-io/pole-server/apis/cache` - 缓存接口
- `github.com/pole-io/pole-server/apis/pkg/types/auth` - 认证类型
- `github.com/pole-io/pole-server/apis/pkg/types/ai` - AI 类型
- `github.com/pole-io/pole-server/pkg/skill` - Skill 服务器接口
- `github.com/pole-io/pole-server/pkg/common/utils/valid` - 参数验证工具
