# Skill Interceptor 实现指南

## 目录结构

```
pkg/skill/interceptor/
├── auth/                # 认证拦截器
│   ├── server.go       # 拦截器主结构
│   ├── skill.go        # Skill 认证逻辑
│   ├── skill_group.go  # SkillGroup 认证逻辑
│   ├── skill_version.go # SkillVersion 认证逻辑
│   ├── skill_subscription.go # SkillSubscription 认证逻辑
│   └── skill_test.go   # 测试文件 (已创建)
├── paramcheck/         # 参数校验拦截器
│   ├── server.go       # 拦截器主结构
│   ├── skill.go        # Skill 参数校验
│   ├── skill_group.go  # SkillGroup 参数校验
│   ├── skill_version.go # SkillVersion 参数校验
│   ├── skill_subscription.go # SkillSubscription 参数校验
│   └── skill_test.go   # 测试文件 (已创建)
└── initialize.go       # 拦截器链初始化 (可选)
```

## 已完成的测试文件

### 1. 认证拦截器测试 (`auth/skill_test.go`)
- Mock 结构：MockUserServer, MockStrategyServer, MockAuthChecker
- 测试场景：11个测试用例，覆盖所有 CRUD 操作的权限验证
- 已实现：权限检查、错误处理、owner 自动填充测试

### 2. 参数校验拦截器测试 (`paramcheck/skill_test.go`)
- Mock 结构：MockSkillServer, MockStore
- 测试场景：12个测试用例，覆盖参数验证
- 已实现：空值检查、格式验证、错误消息验证

## 实现要求

### 认证拦截器 (auth)
```go
// Server 结构体
type Server struct {
    nextSvr   skill.SkillServer
    userSvr   auth.UserServer
    policySvr auth.StrategyServer
    cacheMgr  cacheapi.CacheManager // 可选
}

// 必须实现的方法
func NewServer(nextSvr skill.SkillServer, userSvr auth.UserServer, policySvr auth.StrategyServer) skill.SkillServer
func (svr *Server) CreateSkill(ctx context.Context, skill *aiTypes.Skill) error
func (svr *Server) UpdateSkill(ctx context.Context, skill *aiTypes.Skill) error
func (svr *Server) DeleteSkill(ctx context.Context, id string) error
func (svr *Server) GetSkill(ctx context.Context, id string) (*aiTypes.Skill, error)
func (svr *Server) GetSkillByName(ctx context.Context, name, namespace string) (*aiTypes.Skill, error)
```

### 参数校验拦截器 (paramcheck)
```go
// Server 结构体
type Server struct {
    storage store.Store      // 可选
    nextSvr skill.SkillServer
}

// 必须实现的方法
func NewServer(nextSvr skill.SkillServer) skill.SkillServer
func (svr *Server) CreateSkill(ctx context.Context, skill *aiTypes.Skill) error
func (svr *Server) UpdateSkill(ctx context.Context, skill *aiTypes.Skill) error
func (svr *Server) DeleteSkill(ctx context.Context, id string) error
```

## 错误变量定义 (paramcheck)
```go
var (
    ErrSkillNameEmpty      = errors.New("skill name cannot be empty")
    ErrSkillNamespaceEmpty = errors.New("skill namespace cannot be empty")
    ErrSkillTypeEmpty      = errors.New("skill type cannot be empty")
    ErrSkillIDEmptry       = errors.New("skill id cannot be empty")
)
```

## 测试同步注意事项

1. **接口一致性**：确保 Server 结构和 NewServer 函数与测试文件中的调用匹配
2. **Mock 接口**：Mock 的方法签名必须与 SkillServer 接口完全一致
3. **错误处理**：实际返回的错误应该与测试中预期的错误匹配
4. **依赖注入**：测试中使用的依赖（如 cacheMgr）需要在实现中支持

## 实现优先级

### P0 - 核心功能
- auth/server.go
- auth/skill.go
- paramcheck/server.go
- paramcheck/skill.go

### P1 - 扩展功能
- auth/skill_group.go
- auth/skill_version.go
- auth/skill_subscription.go
- paramcheck/skill_group.go
- paramcheck/skill_version.go
- paramcheck/skill_subscription.go

### 可选功能
- initialize.go (拦截器链组装)
- 更完善的错误处理和日志