---
title: Console API、Client 与权限接口 E2E 测试用例设计
tags: [quality, testcases, e2e, api, console, client, auth]
links: [testing, service-discovery, governance-rules, auth-system, ai-features, configuration]
updated: 2026-06-16
sources: 12
---

# Console API、Client 与权限接口 E2E 测试用例设计

## 目标

覆盖三类端到端场景：

1. Console 接口全流程读写请求：命名空间、MCP、A2A、服务、别名、实例、治理规则、用户、用户组、权限。
2. Client 查询逻辑：通过 Console 完成规则添加、修改、删除后，客户端接口在限定时间内看到符合预期的数据。
3. 权限控制：通过 Console 调整用户、用户组、角色和策略后，Console 读写、Client 读写均按预期放行或拒绝，并与 `consoleOpen` / `clientOpen` 配置开关一致。

## 能力边界

本文档只定义接口 E2E 能力，自动化入口统一为 Go test。测试代码只围绕 HTTP 请求、响应字段、权限错误码、缓存传播和发布态可见性做断言。

Console 维度使用 8080 console proxy 是为了验证控制台接口链路本身，包括 JWT、反代、标准响应解包和接口权限；这不等同于验证控制台页面。本文档不纳入任何前端页面自动化能力，控制台页面渲染、交互流程和静态资源加载不属于验收对象。

## 总体策略

E2E 只验证接口链路，不验证控制台页面渲染和浏览器交互。建议拆成两层套件，共享同一套测试数据和断言语义。

| 套件 | 建议位置 | 入口 | 主要证明 |
|------|----------|------|----------|
| Console API E2E | `test/e2e/console_api/` | Go test，经 8080 HTTP 代理调用 `/core/v1`、`/naming/v1`、`/auth/v1`、`/ai/*` | 所有 Console 读写请求真实经过 console proxy、鉴权 header、标准响应解包 |
| Client E2E | `test/e2e/client/` | Go test，调用 8090 client API | 缓存传播、服务发现、治理规则发布态、客户端权限控制 |

两层套件共享：

- `test/e2e/fixtures/`：命名空间、服务、实例、AI、治理规则、用户、策略样例。
- `test/e2e/assertions/`：标准响应、轮询等待、权限错误、规则匹配断言。
- `test/e2e/env/`：all 模式启动、MySQL 初始化、配置开关覆盖、admin 登录 token 获取。

## 非目标

以下内容不进入本文档的测试范围：

- 不建设前端页面自动化测试能力。
- 不验证控制台前端页面的视觉呈现、交互行为和路由跳转。
- 不以静态资源、页面 hash、CSS/JS 构建产物或前端本地状态作为成功依据。
- 不通过直接查数据库、store 或内存 cache 替代接口断言；数据库只作为失败诊断辅助信息。

## 验收门槛

- Console API E2E 必须覆盖每个资源的 `create -> list/detail -> update -> list/detail -> delete -> list/detail 不存在`。
- Client E2E 对缓存传播使用轮询：默认 `timeout=10s`、`interval=200ms`；如果规则依赖发布态，必须先发布再断言客户端可见。
- 权限 E2E 必须同时覆盖 `consoleOpen=false/true` 与 `clientOpen=false/true` 四种组合。
- 所有测试数据名称必须带唯一前缀，例如 `e2e-${timestamp}-${case}`，清理时按前缀删除，避免污染开发样例数据。

## Console 读写覆盖矩阵

### 命名空间

| ID | 场景 | 操作 | 断言 |
|----|------|------|------|
| CONSOLE-NS-001 | 创建命名空间 | `POST /core/v1/namespaces` | 返回成功；`GET /core/v1/namespaces?name=` 能查到 |
| CONSOLE-NS-002 | 修改命名空间 | `PUT /core/v1/namespaces` 修改描述/标签 | detail/list 展示新描述和标签 |
| CONSOLE-NS-003 | 删除命名空间 | `POST /core/v1/namespaces/delete` | list 不再返回；重复删除返回可识别错误 |
| CONSOLE-NS-004 | 查询过滤 | `GET /core/v1/namespaces` 使用 name/offset/limit 过滤 | 只返回目标命名空间，分页字段稳定 |

### MCP 服务

| ID | 场景 | 操作 | 断言 |
|----|------|------|------|
| CONSOLE-MCP-001 | 创建 MCP Server | `POST /ai/mcp/v1/servers`，backend 使用 Pole 服务 | list/detail 返回 backend namespace/name |
| CONSOLE-MCP-002 | 更新 MCP Server | `PUT /ai/mcp/v1/servers` 修改描述、协议、工具声明 | detail 生效，`/server/tools` 返回工具列表 |
| CONSOLE-MCP-003 | 删除 MCP Server | `POST /ai/mcp/v1/servers/delete` | list 不返回，tool 查询为空或 404 |
| CONSOLE-MCP-004 | 工具目录查询 | `GET /ai/mcp/v1/server/tools` 或等价工具查询接口 | 工具名称、描述、输入 schema 与 MCP Server 保存内容一致 |

### A2A Agent

| ID | 场景 | 操作 | 断言 |
|----|------|------|------|
| CONSOLE-A2A-001 | 创建 Agent | `POST /ai/a2a/v1/agents`，含 Agent Card、skills、backend | list/detail 返回 agent，来源字段包含 `source_type` |
| CONSOLE-A2A-002 | 更新 Agent | `PUT /ai/a2a/v1/agents` 修改 card、skills、backend | `/agents/{id}/card` 和 `/agent/skills` 同步更新 |
| CONSOLE-A2A-003 | 删除 Agent | `POST /ai/a2a/v1/agents/delete` | list/detail/skills 均不可见 |
| CONSOLE-A2A-004 | Agent Card 与技能查询 | 调用 Agent Card、skills 查询接口 | Agent Card、skills、backend 绑定与保存内容一致 |

### 服务、别名、实例

| ID | 场景 | 操作 | 断言 |
|----|------|------|------|
| CONSOLE-SVC-001 | 创建服务 | `POST /naming/v1/services` | list/detail 返回服务，metadata/export_to 正确 |
| CONSOLE-SVC-002 | 修改服务 | `PUT /naming/v1/services` 修改部门、业务、标签、可见性 | detail/list 生效 |
| CONSOLE-SVC-003 | 删除服务 | `POST /naming/v1/services/delete` | list 不返回，相关实例/别名清理符合预期 |
| CONSOLE-ALIAS-001 | 创建别名 | `POST /naming/v1/service/alias` | alias list 返回源服务与目标服务 |
| CONSOLE-ALIAS-002 | 修改别名 | `PUT /naming/v1/service/alias` | alias detail/list 更新 |
| CONSOLE-ALIAS-003 | 删除别名 | `POST /naming/v1/service/aliases/delete` | alias list 不返回 |
| CONSOLE-INS-001 | 创建实例 | `POST /naming/v1/instances` | 实例列表或详情接口返回实例，host/port/metadata 正确 |
| CONSOLE-INS-002 | 修改实例 | `PUT /naming/v1/instances` 修改权重、隔离、metadata | detail/list 生效 |
| CONSOLE-INS-003 | 删除实例 | `POST /naming/v1/instances/delete` | 实例 list 不返回，客户端发现不再返回 |
| CONSOLE-SVC-004 | 服务关联查询 | 分别查询服务详情、实例列表、订阅者列表和别名列表 | 各接口均能用 namespace/name 直达，不依赖创建接口返回的临时上下文 |

### 治理规则

治理规则必须覆盖 9 类：路由、限流、熔断、主动探测、无损、泳道、调用鉴权、流量镜像、流量 Mock。

每类规则都至少包含以下公共用例：

| ID | 场景 | 操作 | 断言 |
|----|------|------|------|
| CONSOLE-GOV-BASE-001 | 创建规则 | 对应 `POST` 接口创建 enabled 规则 | list/detail 返回，`editable/deleteable=true` |
| CONSOLE-GOV-BASE-002 | 修改规则 | 对应 `PUT` 接口修改描述、标签、关键字段 | detail 生效，revision 变化 |
| CONSOLE-GOV-BASE-003 | 发布规则 | `POST <rule>/releases` 发布 normal 版本 | releases 返回 active normal；规则列表或详情响应包含已发布状态 |
| CONSOLE-GOV-BASE-004 | 灰度发布 | `POST <rule>/releases` 发布 gray 版本，带 client label | releases 返回 gray 版本和 client label；`stopbeta` 后灰度发布状态关闭 |
| CONSOLE-GOV-BASE-005 | 回滚/停止灰度 | `PUT <rule>/releases/rollback` 或 `stopbeta` | active 版本恢复或灰度停止 |
| CONSOLE-GOV-BASE-006 | 删除规则 | `POST <rule>/delete` | 当前态不返回，发布态按产品决策清理或不可用 |

分类专项断言：

| 规则类型 | Console 接口 | 关键断言 |
|----------|--------------|----------|
| 路由 | `/naming/v1/routings` | caller/callee、匹配条件、目标分组、权重、实例标签都可读写 |
| 限流 | `/naming/v1/ratelimits` | 请求数/并发数、窗口单位、快速失败响应、条件行可读写 |
| 熔断 | `/naming/v1/circuitbreakers` | 熔断粒度、错误判断、触发条件、恢复策略、降级响应可读写 |
| 主动探测 | `/naming/v1/faultdetectors` | HTTP/gRPC 协议、路径、方法、间隔、超时、header 可读写 |
| 无损 | `/naming/v1/lossless` | 延迟注册、服务预热、无损下线；duration/string 字段按 spec 正确提交 |
| 泳道 | `/naming/v1/lane/groups` | LaneGroup 聚合根、entry selector Any、destinations、LaneRule 子规则可读写 |
| 调用鉴权 | `/naming/v1/traffic/security` | 默认动作、命中动作、接口范围、拒绝响应、匹配条件可读写 |
| 流量镜像 | `/naming/v1/traffic/mirrors` | 来源、目标、比例、生效时长、目标标签、匹配条件可读写 |
| 流量 Mock | `/naming/v1/traffic/mocks` | 接口范围、比例、延迟、状态码、响应头、响应体、匹配条件可读写 |

### 用户、用户组、角色、权限策略

| ID | 场景 | 操作 | 断言 |
|----|------|------|------|
| CONSOLE-AUTH-USER-001 | 创建用户 | `POST /auth/v1/users` | list/detail 返回；可登录或 token 可查询 |
| CONSOLE-AUTH-USER-002 | 修改用户 | `PUT /auth/v1/users` 修改 comment/mobile/email | detail 生效 |
| CONSOLE-AUTH-USER-003 | 用户 token | `GET/PUT /auth/v1/user/token*` | token enable/disable/refresh 生效 |
| CONSOLE-AUTH-GROUP-001 | 创建用户组 | `POST /auth/v1/usergroups` | list/detail 返回 |
| CONSOLE-AUTH-GROUP-002 | 修改用户组成员/标签 | `PUT /auth/v1/usergroups` | detail 生效，成员权限继承 |
| CONSOLE-AUTH-GROUP-003 | 用户组 token | `GET/PUT /auth/v1/usergroup/token*` | token enable/disable/refresh 生效 |
| CONSOLE-AUTH-ROLE-001 | 创建/修改/删除角色 | `/auth/v1/roles` | 角色函数列表与权限校验结果一致 |
| CONSOLE-AUTH-POLICY-001 | 创建权限策略 | `POST /auth/v1/policies` | principals/resources/functions/conditions 持久化 |
| CONSOLE-AUTH-POLICY-002 | 修改权限策略 | `PUT /auth/v1/policies` | 权限立即或在缓存窗口内生效 |
| CONSOLE-AUTH-POLICY-003 | 资源授权 | `POST /auth/v1/resources/authorize` | `principal/resources` 与 `resources/principals` 双向查询一致 |

## Client 查询传播用例

这些用例全部通过 Console API 完成写操作，再通过 Client API 验证可见性。

| ID | Console 写操作 | Client 查询 | 期望 |
|----|----------------|-------------|------|
| CLIENT-SVC-001 | Console 创建服务和 2 个健康实例 | `POST /naming/v1/Discover` | 10s 内返回服务和 2 个实例 |
| CLIENT-SVC-002 | Console 修改实例权重/隔离 | `POST /naming/v1/Discover` | 10s 内实例权重/隔离状态变化；隔离实例按客户端语义不可用或标记隔离 |
| CLIENT-SVC-003 | Console 删除实例 | `POST /naming/v1/Discover` | 10s 内实例不再返回 |
| CLIENT-ROUTE-001 | Console 创建并发布路由规则 | `POST /naming/v1/Discover` | 返回路由规则，规则内容包含匹配条件、目标分组和权重 |
| CLIENT-RL-001 | Console 创建并发布限流规则 | `POST /naming/v1/Discover` | 返回限流规则，窗口、阈值和条件与 Console 保存一致 |
| CLIENT-CB-001 | Console 创建并发布熔断规则 | `POST /naming/v1/Discover` | 返回熔断规则，错误判断和触发条件一致 |
| CLIENT-FD-001 | Console 创建并发布主动探测规则 | `POST /naming/v1/Discover` | 返回探测规则，协议、路径、间隔一致 |
| CLIENT-LOSSLESS-001 | Console 修改无损规则 | `POST /naming/v1/Discover` | 返回无损上线/下线配置，duration 字段符合 spec |
| CLIENT-LANE-001 | Console 创建泳道组和 LaneRule | `POST /naming/v1/Discover` | 返回泳道组和 LaneRule，规则内容包含 entry selector、lane label 条件和目标服务 |
| CLIENT-SEC-001 | Console 创建调用鉴权规则 | 客户端请求治理规则或 Discover | 默认动作、命中动作、匹配条件符合 Console 配置 |
| CLIENT-MIRROR-001 | Console 创建流量镜像规则 | 客户端请求治理规则或 Discover | 镜像目标、比例、生效时长符合 Console 配置 |
| CLIENT-MOCK-001 | Console 创建流量 Mock 规则 | 客户端请求治理规则或 Discover | Mock 状态码、响应头、响应体符合 Console 配置 |
| CLIENT-GOV-DEL-001 | Console 删除或停止发布规则 | `POST /naming/v1/Discover` | 10s 内客户端不再拿到该规则或 active 版本回退 |

传播断言必须记录：

- 写操作完成时间。
- 第一次客户端命中预期的时间。
- 轮询次数。
- 失败时输出最后一次 client 响应、对应 console detail、数据库当前态和 release 当前态。

## 权限控制用例

### 配置开关矩阵

| ID | `consoleOpen` | `clientOpen` | 未登录 Console | 无 token Client | 授权 token | 未授权 token |
|----|---------------|--------------|----------------|-----------------|------------|--------------|
| AUTH-SWITCH-001 | false | false | Console 读写按关闭鉴权语义放行 | Client 读写按关闭鉴权语义放行 | 放行 | 放行 |
| AUTH-SWITCH-002 | true | false | Console 读写拒绝 | Client 读写放行 | Console 放行 | Console 拒绝 |
| AUTH-SWITCH-003 | false | true | Console 读写放行 | Client 读写拒绝 | Client 放行 | Client 拒绝 |
| AUTH-SWITCH-004 | true | true | Console 读写拒绝 | Client 读写拒绝 | 授权范围内放行 | 授权范围外拒绝 |

这里的“放行/拒绝”必须以实际接口响应码断言：成功为 `200000`，无权限应匹配 `401001` 或项目定义的权限错误码，不能只看 HTTP status。

### Console 权限策略

| ID | 准备 | 操作 | 断言 |
|----|------|------|------|
| AUTH-CONSOLE-001 | 创建只读用户，授予命名空间/服务 read 函数 | GET 服务列表，POST 创建服务 | GET 成功，POST 拒绝 |
| AUTH-CONSOLE-002 | 创建写用户，授予指定命名空间服务写权限 | 在授权命名空间创建服务，在未授权命名空间创建服务 | 前者成功，后者拒绝 |
| AUTH-CONSOLE-003 | 创建治理规则管理员 | 修改路由/限流/熔断/新增流量治理规则 | 授权规则类型成功，未授权规则类型拒绝 |
| AUTH-CONSOLE-004 | 用户组继承策略 | 用户加入用户组，策略挂在用户组 | 用户无需直接策略也能操作；移出用户组后在缓存窗口内失效 |
| AUTH-CONSOLE-005 | 角色函数变更 | 修改角色函数集合 | 新函数放行，删除的函数拒绝 |
| AUTH-CONSOLE-006 | token 禁用/刷新 | 禁用用户 token 后请求，再刷新 token | 旧 token 拒绝，新 token 按策略放行 |

### Client 权限策略

| ID | 准备 | 操作 | 断言 |
|----|------|------|------|
| AUTH-CLIENT-001 | `clientOpen=true`，创建 client user/group token，授予 Discover 指定服务 | Discover 授权服务和未授权服务 | 授权服务成功，未授权服务拒绝或被过滤 |
| AUTH-CLIENT-002 | 授予 RegisterInstance 指定服务 | 注册授权服务实例和未授权服务实例 | 前者成功，后者拒绝 |
| AUTH-CLIENT-003 | 授予 Heartbeat 指定实例 | 心跳授权实例和未授权实例 | 前者成功，后者拒绝 |
| AUTH-CLIENT-004 | 策略更新 | Console 收回 client Discover 权限 | 10s 内 client Discover 从成功变为拒绝或过滤 |
| AUTH-CLIENT-005 | `clientOpen=false` | 不带 token 调 Discover/Register/Heartbeat | 均按关闭鉴权语义放行 |

## 推荐执行顺序

1. 环境启动：MySQL 初始化，all 模式启动，等待 8080/8090 接口可用。
2. admin 登录：通过 `/auth/v1/user/login` 或测试库 token 获取 Console token。
3. Console API 读写：创建基础命名空间、服务、实例、AI 资源、治理规则、auth 资源。
4. Client 传播：对服务和治理规则执行发布、修改、删除，轮询 8090 client API。
5. 权限矩阵：重启不同配置开关组合，验证 console/client 读写放行和拒绝。
6. 清理：按测试名前缀删除所有资源，验证 list 不返回。

## 实现注意事项

- Console API E2E 必须经 8080，不要直接打 8090；否则无法证明 console proxy、JWT/header/cookie 传递和响应解包。
- Client E2E 必须经 8090 client API，不要只查数据库或缓存对象；否则无法证明客户端实际可见。
- 本文档只设计接口 E2E；控制台交互类测试不混入该用例集。
- 治理规则要同时验证当前态和发布态；只保存未发布规则不能证明客户端能获取。
- 权限策略变更后需要等待 `StrategyCache`、`UserCache`、治理规则 cache 的刷新窗口；统一使用轮询断言，不使用固定 sleep 作为唯一判断。
- 失败诊断输出应包含 request id、请求 payload、响应 body、当前配置开关、登录主体、资源 id。

## 相关页面

- [[testing]]
- [[service-discovery]]
- [[governance-rules]]
- [[auth-system]]
- [[ai-features]]
- [[configuration]]
