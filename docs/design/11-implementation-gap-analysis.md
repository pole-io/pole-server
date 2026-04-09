# 功能比对分析报告 - 未实现/不完整功能

## 1. 概述

本文档对比 `docs/design/` 下的架构设计文档与现有代码实现，识别出未实现或不完整的功能。

## 2. AI Native 特性模块

### 2.1 缓存层 - 完全缺失

| 功能 | 设计文档 | 实现状态 | 优先级 |
|------|----------|----------|--------|
| SkillCache | 10-ai-native-features.md | **未实现** | P0 |
| MCPServerCache | 10-ai-native-features.md | **未实现** | P0 |
| SkillVersionCache | 10-ai-native-features.md | **未实现** | P1 |
| SkillSubscriptionCache | 10-ai-native-features.md | **未实现** | P1 |

**影响**: AI Native 功能无法享受缓存层带来的性能优化，每次请求都直接访问数据库。

### 2.2 服务层 - 部分实现

| 功能 | 设计文档 | 实现状态 | 优先级 |
|------|----------|----------|--------|
| Skill CRUD | 10-ai-native-features.md | ✅ 已实现 | - |
| SkillGroup CRUD | 10-ai-native-features.md | ✅ 已实现 | - |
| SkillVersion 基础操作 | 10-ai-native-features.md | ✅ 已实现 | - |
| SkillSubscription 基础操作 | 10-ai-native-features.md | ✅ 已实现 | - |
| Skill 拦截器 (auth) | 架构设计 | **未实现** | P0 |
| Skill 拦截器 (paramcheck) | 架构设计 | **未实现** | P0 |

### 2.3 存储层 - 部分实现

| 功能 | 文件 | 实现状态 | 优先级 |
|------|------|----------|--------|
| QuerySkillVersions | plugin/store/mysql/skill.go:792 | **TODO** | P1 |
| QuerySkillSubscriptions | plugin/store/mysql/skill.go:1081 | **TODO** | P1 |

### 2.4 MCP Server 工具执行 - 未实现

| 功能 | 文件 | 实现状态 | 优先级 |
|------|------|----------|--------|
| handleQueryMCPServers | aimcp/mcp_server.go:37-82 | **部分实现** (只返回成功) | P0 |
| handleCreateMCPServers | aimcp/mcp_server.go:85-111 | **TODO** | P0 |
| handleUpdateMCPServers | aimcp/mcp_server.go:114-134 | **TODO** | P0 |
| handleDeleteMCPServers | aimcp/mcp_server.go:137-170 | **TODO** | P0 |
| handleQueryMCPServerTools | aimcp/mcp_server.go:173-236 | **TODO** | P0 |
| mcpServerQuery | aimcp/mcp_server.go:237-239 | **TODO** | P0 |
| mcpServerCreate | aimcp/mcp_server.go:243-245 | **TODO** | P0 |
| mcpServerUpdate | aimcp/mcp_server.go:249-251 | **TODO** | P0 |

**影响**: MCP 工具只声明了接口，没有实际业务逻辑。

### 2.5 Skill HTTP API - 完全缺失

| 功能 | 设计文档 | 实现状态 | 优先级 |
|------|----------|----------|--------|
| Skill HTTP API 端点 | 架构设计 | **未实现** | P0 |
| SkillGroup HTTP API 端点 | 架构设计 | **未实现** | P0 |
| SkillVersion HTTP API 端点 | 架构设计 | **未实现** | P1 |
| SkillSubscription HTTP API 端点 | 架构设计 | **未实现** | P1 |

## 3. 订阅推送机制

### 3.1 配置变更推送

| 功能 | 设计文档 | 实现状态 | 优先级 |
|------|----------|----------|--------|
| Skill 订阅变更推送 | 10-ai-native-features.md | **未实现** | P1 |
| 版本更新主动推送 | 10-ai-native-features.md | **未实现** | P1 |

## 4. 未来规划功能 (文档中标记)

根据 10-ai-native-features.md 第 10 节 "未来规划"，以下功能尚未实现：

| 功能 | 描述 | 状态 |
|------|------|------|
| Skill 市场 | 支持 Skill 的分享和发现 | 未开始 |
| Skill 组合 | 支持多个 Skill 组合成工作流 | 未开始 |
| Skill 监控 | 提供调用链追踪和性能监控 | 未开始 |
| Skill 安全 | 支持 Skill 执行沙箱和权限控制 | 未开始 |

## 5. 建议的补充任务

### P0 - 必须立即实现

1. **Skill 缓存层实现**
   - 创建 `pkg/cache/ai/skill.go`
   - 创建 `pkg/cache/ai/mcp_server.go`
   - 注册到 CacheManager
   - 验收标准: Skill/MCP Server 数据能从缓存读取

2. **MCP Tool 业务逻辑实现**
   - 实现 `handleCreateMCPServers`
   - 实现 `handleUpdateMCPServers`
   - 实现 `handleDeleteMCPServers`
   - 实现 `handleQueryMCPServers`
   - 验收标准: MCP 工具能正确执行 CRUD 操作

3. **Skill HTTP API 实现**
   - 创建 `plugin/apiserver/httpserver/skill/server.go`
   - 实现 Skill CRUD API
   - 实现 SkillGroup CRUD API
   - 验收标准: 能通过 HTTP API 管理 Skill

4. **Skill 拦截器实现**
   - 创建 `pkg/skill/interceptor/auth/`
   - 创建 `pkg/skill/interceptor/paramcheck/`
   - 验收标准: Skill API 有完整的鉴权和参数校验

### P1 - 应该尽快实现

5. **存储层 TODO 完成**
   - 实现 `QuerySkillVersions`
   - 实现 `QuerySkillSubscriptions`
   - 验收标准: 能分页查询 Skill 版本和订阅

6. **订阅推送机制**
   - 实现 Skill 变更推送
   - 实现版本更新通知
   - 验收标准: 订阅客户端能收到变更推送

7. **SkillVersion HTTP API**
   - 实现 SkillVersion 管理 API
   - 验收标准: 能通过 HTTP 管理 Skill 版本

8. **SkillSubscription HTTP API**
   - 实现 SkillSubscription 管理 API
   - 验收标准: 能通过 HTTP 管理订阅关系

### P2 - 后续迭代

9. **Skill 市场**
   - Skill 分享机制
   - Skill 发现功能

10. **Skill 监控**
    - 调用链追踪
    - 性能指标收集

## 6. 实现优先级总结

```
P0 (阻塞发布)
├── Skill 缓存层
├── MCP Tool 业务逻辑
├── Skill HTTP API
└── Skill 拦截器

P1 (尽快实现)
├── 存储层 TODO
├── 订阅推送机制
├── SkillVersion HTTP API
└── SkillSubscription HTTP API

P2 (后续迭代)
├── Skill 市场
├── Skill 监控
├── Skill 组合
└── Skill 安全
```
