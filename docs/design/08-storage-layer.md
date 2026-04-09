# 存储层架构设计

## 1. 概述

存储层 (`apis/store`) 定义了 Pole Control Plane 的数据持久化接口。采用插件化设计，支持多种存储后端（目前主要支持 MySQL）。存储层提供事务支持、增量查询等高级特性。

## 2. 模块结构

```
apis/store/
├── store.go           # Store 核心接口定义
├── api.go             # 通用存储接口
├── discover_api.go    # 服务发现存储接口
├── config_file_api.go # 配置中心存储接口
├── governance_api.go  # 治理规则存储接口
├── auth_api.go        # 鉴权存储接口
├── admin_api.go       # 管理存储接口
├── ai.go              # AI 模块存储接口
├── status.go          # 存储状态定义
└── sql.go             # SQL 工具函数

plugin/store/
└── mysql/             # MySQL 存储实现
    ├── base_db.go     # 基础数据库操作
    ├── discover.go    # 服务发现存储实现
    ├── config_file.go # 配置存储实现
    ├── common.go      # 公共函数
    └── ...
```

## 3. 核心接口设计

### 3.1 Store 接口

```go
// apis/store/store.go
type Store interface {
    // 基础方法
    Name() string
    Initialize(c *Config) error
    Destroy() error

    // 事务支持
    CreateTransaction() (Transaction, error)
    StartTx() (Tx, error)
    StartReadTx() (Tx, error)

    // 模块存储接口
    NamespaceStore
    DiscoverStore
    ConfigFileModuleStore
    GovernanceStore
    ClientStore
    AdminStore
    GrayStore
    AuthStore
    AIStore
}
```

### 3.2 Store 配置

```go
type Config struct {
    Name   string
    Option map[string]interface{}
}
```

## 4. 事务接口

### 4.1 Transaction 接口

```go
// apis/store/store.go
type Transaction interface {
    // Commit 提交事务
    Commit() error
    // LockBootstrap 启动锁，限制 Server 并发启动
    LockBootstrap(key string, server string) error
    // LockNamespace 行锁命名空间
    LockNamespace(name string) (*types.Namespace, error)
    // DeleteNamespace 删除命名空间
    DeleteNamespace(name string) error
    // LockService 行锁服务
    LockService(name string, namespace string) (*svctypes.Service, error)
    // RLockService 共享锁服务
    RLockService(name string, namespace string) (*svctypes.Service, error)
}
```

### 4.2 Tx 接口

```go
// apis/store/store.go
type Tx interface {
    // Commit 提交事务
    Commit() error
    // Rollback 回滚事务
    Rollback() error
    // GetDelegateTx 获取原始事务对象
    GetDelegateTx() interface{}
    // CreateReadView 创建快照读视图
    CreateReadView() error
}
```

## 5. 命名空间存储接口

```go
type NamespaceStore interface {
    // AddNamespace 保存命名空间
    AddNamespace(namespace *types.Namespace) error
    // UpdateNamespace 更新命名空间
    UpdateNamespace(namespace *types.Namespace) error
    // GetNamespace 根据名称获取命名空间
    GetNamespace(name string) (*types.Namespace, error)
    // GetNamespaces 查询命名空间
    GetNamespaces(filter map[string][]string, offset, limit int) ([]*types.Namespace, uint32, error)
    // GetMoreNamespaces 增量获取命名空间
    GetMoreNamespaces(mtime time.Time) ([]*types.Namespace, error)
}
```

## 6. 服务发现存储接口

```go
type DiscoverStore interface {
    // 服务管理
    AddService(service *svctypes.Service) error
    UpdateService(service *svctypes.Service) error
    DeleteService(service *svctypes.Service) error
    GetService(name, namespace string) (*svctypes.Service, error)
    GetServices(filter map[string]string, offset, limit int) ([]*svctypes.Service, uint32, error)
    GetMoreServices(mtime time.Time, firstUpdate bool) (map[string]*svctypes.Service, error)

    // 实例管理
    AddInstance(instance *svctypes.Instance) error
    UpdateInstance(instance *svctypes.Instance) error
    DeleteInstance(instance *svctypes.Instance) error
    GetInstance(id string) (*svctypes.Instance, error)
    GetInstancesMainByServiceID(serviceID string) ([]*svctypes.Instance, error)
    GetMoreInstances(mtime time.Time, firstUpdate bool) ([]*svctypes.Instance, error)

    // 服务别名管理
    // 服务契约管理
    // ...
}
```

## 7. 配置中心存储接口

```go
type ConfigFileModuleStore interface {
    // 配置分组
    CreateConfigFileGroup(group *conftypes.ConfigFileGroup) error
    UpdateConfigFileGroup(group *conftypes.ConfigFileGroup) error
    DeleteConfigFileGroup(namespace, name string) error
    GetConfigFileGroup(namespace, name string) (*conftypes.ConfigFileGroup, error)
    GetMoreConfigFileGroups(firstUpdate bool, mtime time.Time) ([]*conftypes.ConfigFileGroup, error)

    // 配置文件
    CreateConfigFile(file *conftypes.ConfigFile) error
    UpdateConfigFile(file *conftypes.ConfigFile) error
    DeleteConfigFile(namespace, group, name string) error
    GetConfigFile(namespace, group, name string) (*conftypes.ConfigFile, error)

    // 配置发布
    CreateConfigFileRelease(release *conftypes.ConfigFileRelease) error
    UpdateConfigFileRelease(release *conftypes.ConfigFileRelease) error
    DeleteConfigFileRelease(namespace, group, fileName, releaseName string) error
    GetConfigFileActiveRelease(namespace, group, fileName string) (*conftypes.ConfigFileRelease, error)

    // 发布历史
    // 配置模板
    // ...
}
```

## 8. 治理规则存储接口

```go
type GovernanceStore interface {
    // 路由规则
    CreateRouteRule(rule *rules.RouterConfig) error
    UpdateRouteRule(rule *rules.RouterConfig) error
    DeleteRouteRule(rule *rules.RouterConfig) error
    GetRouteRule(id string) (*rules.RouterConfig, error)
    GetMoreRouteRules(mtime time.Time, firstUpdate bool) ([]*rules.RouterConfig, error)

    // 限流规则
    CreateRateLimitRule(rule *rules.RateLimit) error
    UpdateRateLimitRule(rule *rules.RateLimit) error
    DeleteRateLimitRule(rule *rules.RateLimit) error
    GetMoreRateLimitRules(mtime time.Time, firstUpdate bool) ([]*rules.RateLimit, error)

    // 熔断规则
    // 故障探测规则
    // 泳道规则
    // 无损规则
    // ...
}
```

## 9. 鉴权存储接口

```go
type AuthStore interface {
    // 用户管理
    CreateUser(user *authtypes.User) error
    UpdateUser(user *authtypes.User) error
    DeleteUser(id string) error
    GetUser(id string) (*authtypes.User, error)
    GetUserByName(name string) (*authtypes.User, error)
    GetMoreUsers(mtime time.Time, firstUpdate bool) ([]*authtypes.User, error)

    // 用户组管理
    // 角色管理
    // 策略管理
    // ...
}
```

## 10. AI 模块存储接口

```go
// apis/store/ai.go
type AIStore interface {
    MCPServerStore
    SkillStore
    SkillGroupStore
    SkillVersionStore
    SkillSubscriptionStore
}

type MCPServerStore interface {
    CreateMCPServer(server *ai.MCPServer) error
    UpdateMCPServer(server *ai.MCPServer) error
    DeleteMCPServer(id string) error
    GetMCPServer(id string) (*ai.MCPServer, error)
    GetMCPServerByName(name, namespace string) (*ai.MCPServer, error)
    GetMoreMCPServers(mtime time.Time, firstUpdate bool) ([]*ai.MCPServer, error)
    QueryMCPServers(filter map[string]string, offset, limit uint32) (uint32, []*ai.MCPServer, error)

    // MCP Server Tool 相关
    CreateMCPServerTool(tool *ai.MCPServerTool) error
    UpdateMCPServerTool(tool *ai.MCPServerTool) error
    DeleteMCPServerTool(id string) error
    GetMCPServerToolsByServerID(serverID string) ([]*ai.MCPServerTool, error)
}

type SkillStore interface {
    CreateSkill(skill *ai.Skill) error
    UpdateSkill(skill *ai.Skill) error
    DeleteSkill(id string) error
    GetSkill(id string) (*ai.Skill, error)
    GetSkillByName(name, namespace string) (*ai.Skill, error)
    GetMoreSkills(mtime time.Time, firstUpdate bool) ([]*ai.Skill, error)
}
```

## 11. 存储注册机制

### 11.1 注册函数

```go
// apis/store/store.go
var StoreSlots = make(map[string]Store)

func RegisterStore(s Store) error {
    name := s.Name()
    if _, ok := StoreSlots[name]; ok {
        return errors.New("store name already existed")
    }
    StoreSlots[name] = s
    return nil
}
```

### 11.2 获取存储实例

```go
func GetStore() (Store, error) {
    name := config.Name
    if name == "" {
        return nil, errors.New("store name is empty")
    }

    store, ok := StoreSlots[name]
    if !ok {
        return nil, fmt.Errorf("store `%s` not found", name)
    }

    initialize(store)
    return store, nil
}
```

## 12. MySQL 存储实现

### 12.1 初始化

```go
// plugin/store/mysql/base_db.go
type MySQLStore struct {
    masterDB *sql.DB
    slaveDB  *sql.DB
    config   *Config
}

func (m *MySQLStore) Initialize(c *store.Config) error {
    // 解析配置
    m.config = parseConfig(c.Option)

    // 连接数据库
    m.masterDB = connectDB(m.config.Master)
    if m.config.Slave != "" {
        m.slaveDB = connectDB(m.config.Slave)
    }

    return nil
}
```

### 12.2 增量查询实现

```go
func (m *MySQLStore) GetMoreServices(mtime time.Time, firstUpdate bool) (map[string]*svctypes.Service, error) {
    var query string
    if firstUpdate {
        query = "SELECT * FROM service WHERE flag = 0"
    } else {
        query = "SELECT * FROM service WHERE mtime > ? AND flag = 0"
    }

    rows, err := m.slaveDB.Query(query, mtime)
    // ...
}
```

## 13. 工具接口

```go
// apis/store/store.go
type ToolStore interface {
    // GetUnixSecond 获取存储层当前时间
    GetUnixSecond(maxWait time.Duration) (int64, error)
}
```

用于缓存增量查询时获取数据库服务器时间，确保时间戳一致性。

## 14. 存储配置示例

```yaml
store:
  name: mysql
  option:
    master:
      host: localhost
      port: 3306
      user: root
      password: password
      database: polaris
    slave:
      host: localhost
      port: 3306
      user: root
      password: password
      database: polaris
    connectionPool:
      maxOpenConns: 100
      maxIdleConns: 20
      connMaxLifetime: 300s
```

## 15. 错误码映射

```go
// apis/store/status.go
func StoreCode2APICode(err error) apimodel.Code {
    // 将存储层错误码映射到 API 错误码
}
```
