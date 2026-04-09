# 访问控制模块架构设计

## 1. 概述

访问控制模块提供认证和授权能力，包括用户管理、策略管理和角色管理。该模块采用插件化设计，支持细粒度的资源访问控制。

## 2. 模块结构

```
apis/access_control/
├── auth/
│   ├── auth.go          # 认证核心接口
│   ├── api.go           # API 接口定义
│   └── context.go       # 上下文定义
├── ratelimit/
│   └── ratelimit.go     # 限流接口
└── whitelist/
    └── whitelist.go     # 白名单接口

plugin/access_control/
├── auth/
│   ├── user/            # 用户管理插件
│   │   ├── default.go   # 默认实现
│   │   ├── token.go     # Token 管理
│   │   └── inteceptor/  # 拦截器
│   └── policy/          # 策略管理插件
│       ├── default.go
│       ├── policy.go
│       ├── role.go
│       └── inteceptor/
├── ratelimit/
│   └── token/           # Token 限流插件
│       ├── api_limit.go
│       ├── limiter.go
│       └── resource_limiter.go
└── whitelist/
    └── ip/              # IP 白名单插件
        └── ip_whitelist.go
```

## 3. 核心接口设计

### 3.1 UserServer 接口

```go
// apis/access_control/auth/api.go
type UserServer interface {
    // 初始化
    Initialize(opt *Config, storage store.Store, policySvr StrategyServer, cacheMgr cachetypes.CacheManager) error

    // 用户管理
    CreateUser(ctx context.Context, user *apisecurity.User) *apimodel.Response
    UpdateUser(ctx context.Context, user *apisecurity.User) *apimodel.Response
    DeleteUser(ctx context.Context, user *apisecurity.User) *apimodel.Response
    GetUser(ctx context.Context, user *apisecurity.User) *apimodel.Response
    QueryUsers(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 用户组管理
    CreateUserGroup(ctx context.Context, group *apisecurity.UserGroup) *apimodel.Response
    UpdateUserGroup(ctx context.Context, group *apisecurity.UserGroup) *apimodel.Response
    DeleteUserGroup(ctx context.Context, group *apisecurity.UserGroup) *apimodel.Response
    QueryUserGroups(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // Token 管理
    CreateToken(ctx context.Context, req *apisecurity.User) *apimodel.Response
    DeleteToken(ctx context.Context, req *apisecurity.User) *apimodel.Response
    UpdateToken(ctx context.Context, req *apisecurity.User) *apimodel.Response

    // 登录认证
    Login(ctx context.Context, req *apisecurity.LoginRequest) *apimodel.Response
    Logout(ctx context.Context) *apimodel.Response
}
```

### 3.2 StrategyServer 接口

```go
// apis/access_control/auth/api.go
type StrategyServer interface {
    // 初始化
    Initialize(opt *Config, storage store.Store, cacheMgr cachetypes.CacheManager, userSvr UserServer) error

    // 策略管理
    CreateStrategy(ctx context.Context, strategy *apisecurity.AuthStrategy) *apimodel.Response
    UpdateStrategy(ctx context.Context, strategy *apisecurity.AuthStrategy) *apimodel.Response
    DeleteStrategy(ctx context.Context, strategy *apisecurity.AuthStrategy) *apimodel.Response
    GetStrategy(ctx context.Context, strategy *apisecurity.AuthStrategy) *apimodel.Response
    QueryStrategies(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 角色管理
    CreateRole(ctx context.Context, role *apisecurity.Role) *apimodel.Response
    UpdateRole(ctx context.Context, role *apisecurity.Role) *apimodel.Response
    DeleteRole(ctx context.Context, role *apisecurity.Role) *apimodel.Response
    QueryRoles(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

    // 权限检查
    VerifyResourceAction(ctx context.Context, req *apisecurity.VerifyResourceActionRequest) *apimodel.Response
}
```

## 4. 认证模型

### 4.1 User 模型

```go
type User struct {
    Id           string
    Name         string
    Password     string
    Owner        string
    Source       string       // Polaris, LDAP, OIDC
    Type         UserType     // Admin, SubAdmin, Normal
    Comment      string
    Token        string
    TokenEnable  bool
    CreateTime   time.Time
    ModifyTime   time.Time
}
```

### 4.2 UserGroup 模型

```go
type UserGroup struct {
    Id          string
    Name        string
    Comment     string
    Owner       string
    Relation    []*UserGroupRelation
    CreateTime  time.Time
    ModifyTime  time.Time
}

type UserGroupRelation struct {
    UserId    string
    GroupId   string
}
```

## 5. 授权模型

### 5.1 Strategy 模型

```go
type StrategyDetail struct {
    Id          string
    Name        string
    Action      string        // READ, WRITE, DELETE, ALL
    Comment     string
    Principals  []*Principal
    Resources   []*ResourceEntry
    Conditions  []*Condition
    CreateTime  time.Time
    ModifyTime  time.Time
}

type Principal struct {
    Type  PrincipalType  // User, Group
    Id    string
    Name  string
}

type ResourceEntry struct {
    Type      ResourceType  // Namespace, Service, Config, etc.
    Id        string
    Name      string
    Namespace string
}
```

### 5.2 Role 模型

```go
type Role struct {
    Id          string
    Name        string
    Comment     string
    Owner       string
    Permissions []*Permission
    CreateTime  time.Time
    ModifyTime  time.Time
}

type Permission struct {
    Resource  *ResourceEntry
    Actions   []string
}
```

## 6. Token 管理

### 6.1 Token 结构

```go
type Token struct {
    Id         string
    UserId     string
    Token      string
    Type       TokenType    // Main, Sub
    Enable     bool
    ExpireTime time.Time
    CreateTime time.Time
}
```

### 6.2 Token 验证流程

```
Request with Token
        │
        ▼
┌─────────────────┐
│ Token Extractor │ ← 从请求中提取 Token
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Token Cache    │ ← 检查 Token 缓存
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Token Verify   │ ← 验证 Token 有效性
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  User Context   │ ← 设置用户上下文
└─────────────────┘
```

## 7. 权限检查机制

### 7.1 拦截器模式

```go
// pkg/service/interceptor/auth/service.go
type ServiceAuthInterceptor struct {
    server    service.DiscoverServer
    userSvr   auth.UserServer
    policySvr auth.StrategyServer
}

func (i *ServiceAuthInterceptor) CreateService(ctx context.Context, req *apiservice.Service) *apimodel.Response {
    // 1. 获取用户信息
    user, _ := auth.GetUserFromContext(ctx)

    // 2. 检查权限
    if errResp := i.checkPermission(ctx, user, req); errResp != nil {
        return errResp
    }

    // 3. 执行业务逻辑
    return i.server.CreateService(ctx, req)
}
```

### 7.2 权限检查流程

```
┌─────────────────┐
│  Get Principal  │ ← 获取操作主体（用户/组）
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Get Strategies  │ ← 获取主体的策略
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Match Resources │ ← 匹配资源规则
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Check Action   │ ← 检查操作权限
└─────────────────┘
```

### 7.3 资源匹配

```go
func (c *StrategyCache) Hint(ctx context.Context, p Principal, r *ResourceEntry) apisecurity.AuthAction {
    strategies := c.GetPrincipalPolicies("allow", p)

    for _, strategy := range strategies {
        if matchResource(strategy.Resources, r) {
            return apisecurity.AuthAction_ALLOW
        }
    }

    return apisecurity.AuthAction_DENY
}
```

## 8. 限流控制

### 8.1 限流接口

```go
// apis/access_control/ratelimit/ratelimit.go
type RateLimiter interface {
    // Allow 检查是否允许请求
    Allow(ctx context.Context, key string) (bool, error)
    // WaitFor 等待获取令牌
    WaitFor(ctx context.Context, key string, timeout time.Duration) error
    // Release 释放令牌
    Release(ctx context.Context, key string) error
}
```

### 8.2 Token 限流实现

```go
// plugin/access_control/ratelimit/token/limiter.go
type TokenLimiter struct {
    limiters sync.Map  // key -> *rate.Limiter
    qps      int
    burst    int
}

func (l *TokenLimiter) Allow(ctx context.Context, key string) (bool, error) {
    limiter := l.getLimiter(key)
    return limiter.Allow(), nil
}
```

### 8.3 API 级别限流

```go
// plugin/access_control/ratelimit/token/api_limit.go
type ApiRateLimiter struct {
    defaultQps int
    apiLimits  map[string]int  // API -> QPS
}

func (l *ApiRateLimiter) CheckLimit(api string) bool {
    qps, ok := l.apiLimits[api]
    if !ok {
        qps = l.defaultQps
    }
    // ...
}
```

## 9. 白名单控制

### 9.1 白名单接口

```go
// apis/access_control/whitelist/whitelist.go
type Whitelist interface {
    // Initialize 初始化
    Initialize(option map[string]interface{}) error
    // IsInWhitelist 检查是否在白名单中
    IsInWhitelist(ctx context.Context, ip string) bool
}
```

### 9.2 IP 白名单实现

```go
// plugin/access_control/whitelist/ip/ip_whitelist.go
type IPWhitelist struct {
    ips      *container.SyncSet[string]
    ipRanges []*net.IPNet
}

func (w *IPWhitelist) IsInWhitelist(ctx context.Context, ip string) bool {
    // 检查精确 IP
    if w.ips.Contains(ip) {
        return true
    }

    // 检查 IP 范围
    parsedIP := net.ParseIP(ip)
    for _, ipNet := range w.ipRanges {
        if ipNet.Contains(parsedIP) {
            return true
        }
    }

    return false
}
```

## 10. 配置示例

```yaml
auth:
  user:
    name: defaultUser
    option:
      salt: "polaris"
  strategy:
    name: defaultStrategy
  interceptors:
    - auth
    - paramcheck

ratelimit:
  name: token
  option:
    qps: 1000
    burst: 100
    apiLimits:
      "/v1/services": 500
      "/v1/instances": 800

whitelist:
  name: ip
  option:
    ips:
      - "127.0.0.1"
      - "10.0.0.0/8"
```

## 11. 鉴权缓存

### 11.1 UserCache

```go
type UserCache interface {
    Cache
    GetAdmin() *authtypes.User
    GetUserByID(id string) *authtypes.User
    GetUserByName(name string) *authtypes.User
    GetGroup(id string) *authtypes.UserGroupDetail
    IsUserInGroup(userId, groupId string) bool
    IsOwner(id string) bool
}
```

### 11.2 StrategyCache

```go
type StrategyCache interface {
    Cache
    GetPolicyRule(id string) *authtypes.StrategyDetail
    GetPrincipalPolicies(effect string, p authtypes.Principal) []*authtypes.StrategyDetail
    Hint(ctx context.Context, p authtypes.Principal, r *authtypes.ResourceEntry) apisecurity.AuthAction
}
```

## 12. 安全最佳实践

1. **密码加密**: 使用 bcrypt 加密存储密码
2. **Token 安全**: Token 采用随机生成，支持过期时间
3. **最小权限原则**: 默认拒绝，显式允许
4. **审计日志**: 记录所有敏感操作
5. **限流保护**: 防止暴力破解和 DoS 攻击
