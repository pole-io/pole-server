---
title: 访问控制与认证系统
tags: [auth, security]
links: [architecture, domain-components, patterns]
updated: 2026-05-14
sources: 1
---

# 访问控制与认证系统

## 架构

认证系统包含两个维度：
1. **身份认证（Authentication）** — 验证身份（Token、用户名/密码）
2. **授权（Authorization）** — 执行资源级别的权限控制

两者均以插件方式实现，并采用拦截器链模式。整体架构见 [[architecture]]，各业务域的认证拦截器见 [[domain-components]]，拦截器模式详解见 [[patterns]]。

## 插件接口（`apis/access_control/auth/`）

```go
// 认证系统配置
type Config struct {
    User     ConfigEntry   // 例如 "defaultUser"
    Strategy ConfigEntry   // 例如 "defaultStrategy"
    Interceptors []string  // 拦截器链名称
}

// UserServer — 管理用户和用户组
type UserServer interface {
    CreateUser(ctx, req) *apimodel.Response
    UpdateUser(ctx, req) *apimodel.Response
    DeleteUsers(ctx, reqs) *apimodel.BatchWriteResponse
    GetUser(ctx, filter) *apimodel.Response
    GetUsers(ctx, query) *apimodel.BatchQueryResponse
    
    CreateGroup, UpdateGroup, DeleteGroup, GetGroup, GetGroups
    
    Login(ctx, req) *apimodel.Response
    GetUserToken, UpdateUserToken
    CheckCredential(ctx, req) *authtypes.AuthResponse
}

// StrategyServer — 管理访问策略
type StrategyServer interface {
    CreateStrategy, UpdateStrategy, DeleteStrategy
    GetStrategy, GetStrategies
    
    CreateRole, UpdateRole, DeleteRole, GetRoles
    
    // 核心授权检查
    CheckResourcesAuth(ctx, req) *apimodel.BatchWriteResponse
    CheckConsolePermission(ctx, req) *authtypes.AuthResponse
}
```

## 插件实现（`plugin/access_control/auth/`）

### 用户管理（`auth/user/`）
- 默认实现：`defaultUser`
- 将用户/用户组存储在 MySQL 中（`AuthStore`）
- 使用缓存加速查询（`UserCache`、`RoleCache`）
- 使用 `golang.org/x/crypto` 进行密码哈希

### 策略管理（`auth/policy/`）
- 默认实现：`defaultStrategy`
- 资源级访问控制（类似 AWS IAM）
- 策略将用户/用户组与资源及允许的操作绑定
- 使用 `StrategyCache` 加速策略评估

### 限流（`access_control/ratelimit/token/`）
- 令牌桶算法，用于 API 级别的速率限制
- 防止管理/运维 API 被滥用

### IP 白名单（`access_control/whitelist/ip/`）
- 将访问限制在特定 IP 范围内
- 可按部署环境配置

## 拦截器链模式

每个业务领域都有认证拦截器来包装真实的服务端：

```
pkg/service/interceptor/auth/server.go
pkg/config/interceptor/auth/server.go
pkg/goverrule/interceptor/auth/server.go
pkg/namespace/interceptor/auth/server.go
pkg/admin/interceptor/auth/server.go
```

模式示例（以服务认证拦截器为例）：
```go
type Server struct {
    nextSvr     service.DiscoverServer  // 包装真实服务端
    userSvr     auth.UserServer
    strategySvr auth.StrategyServer
}

func (s *Server) CreateService(ctx context.Context, req *apiservice.Service) *apimodel.Response {
    // 1. 从 ctx 中提取认证上下文（来自 HTTP 请求头的 Token/用户信息）
    // 2. 通过 strategySvr.CheckConsolePermission() 检查权限
    // 3. 授权通过后，委托给 s.nextSvr.CreateService()
}
```

## 认证 Token 流程

1. 客户端在请求中携带 `X-Polaris-Token` 或 `Authorization` 请求头
2. HTTP 服务端中间件将 Token 提取到 `context.Context` 中
3. 认证拦截器调用 `UserServer.CheckCredential()` 验证 Token
4. `StrategyServer.CheckConsolePermission()` 评估资源策略
5. 授权通过后，请求继续流向业务逻辑层

## 参数检查拦截器

与认证拦截器相邻，参数检查拦截器负责验证请求参数：
```
pkg/service/interceptor/paramcheck/
pkg/config/interceptor/paramcheck/
```

这些拦截器在请求到达业务逻辑之前，对必填字段、长度限制、格式约束等进行校验。

## 相关页面

- [[architecture]]
- [[domain-components]]
- [[patterns]]
