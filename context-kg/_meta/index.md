---
title: pole-control-plane 知识库
tags: [meta, index]
links: [schema, log]
updated: 2026-07-20
sources: 0
---

# pole-control-plane 知识库

本目录是 `context-kg/` wiki 的入口。所有页面均使用 [[schema]] 中定义的格式规范。操作历史见 [[log]]。

知识库按三大知识域组织：业务知识域、技术知识域、质量保障知识域。

---

## 任务过程记录（tasks/）

- [[todo]] — 当前任务计划、进度、review 和验证记录 | tasks, todo
- [[lessons]] — 用户纠正后沉淀的经验规则 | tasks, lessons

## 业务知识域（business/）

- [[terminology]] — 领域术语表：命名空间、服务、实例、配置、治理规则、MCP 服务 | business, terminology
- [[domain-models]] — 核心业务实体及实体关系 | business, domain-model
- [[business-rules]] — 跨域业务规则、服务发现规则、配置规则和治理规则 | business, rules

### 功能档案（business/feature-registry/）

- [[namespace]] — 多租户命名空间：关键常量、接口定义与 singleflight 防重复创建 | business, feature, namespace
- [[service-discovery]] — 服务发现核心：DiscoverServer 接口、批处理、健康检查、空推保护 | business, feature, service, healthcheck
- [[config-center]] — 配置中心：版本化配置文件、灰度发布、Watch 长轮询机制 | business, feature, config
- [[governance-rules]] — 治理规则：路由、限流、熔断、故障探测、泳道、无损规则 | business, feature, governance
- [[admin]] — 管理后台：AdminOperateServer 接口与跨域运维操作 | business, feature, admin

## 技术知识域（technical/）

### 架构（technical/arch/）

- [[overview]] — 项目介绍、技术选型、顶层目录结构与启动流程 | overview
- [[architecture]] — 四层架构设计、插件系统、Store 接口、CacheManager、拦截器链与请求流程 | architecture, design

### 接口契约（technical/apis/）

- [[api-servers]] — HTTP、gRPC、xDS、Nacos、Apollo、Eureka 六类协议服务端与 MCP 集成 | api, http, grpc, xds, nacos

### 环境部署（technical/Environment/）

- [[configuration]] — YAML 配置文件结构、插件配置与部署目录说明 | config, yaml, deploy

### 模块说明（technical/modules/）

- [[storage]] — 存储接口层次结构、MySQL 实现特性（软删除、增量查询、分布式锁）与 Mock Store | storage, mysql, database
- [[cache-layer]] — 缓存子类型、增量刷新循环、AI 缓存子包与按需开启机制 | cache, performance
- [[auth-system]] — 认证（Authentication）与授权（Authorization）插件接口、拦截器链与 Token 流程 | auth, security
- [[common-infra]] — 日志、EventHub、Batch Controller、OTel 指标、同步原语与通用工具 | infra, logging, eventhub
- [[ai-features]] — MCP 与 A2A Agent Registry 的 AI Native 设计 | ai, mcp

### 技术约定（technical/conventions/）

- [[patterns]] — 代码库中的关键模式与约定（插件注册、单例、拦截器、缓存、治理规则编辑抽屉交互等） | patterns, conventions

### 架构决策（technical/adr/）

- [[adr-governance-rule-unified-storage]] — 治理规则统一存储与缓存更新决策 | adr, governance, storage, cache
- [[adr-governance-request-parameter-capture]] — 治理请求参数采集与运行变量移除边界 | adr, governance, routing, ratelimit, sdk
- [[adr-a2a-agent-registry]] — A2A Agent Registry 功能设计与分期 | adr, ai, a2a, mcp, registry
- [[adr-console-oidc-identity-source]] — Console OIDC 用户来源与目录同步 | adr, auth, console, oidc, identity

## 质量保障知识域（quality/）

### 自动化背景知识（quality/automation/）

- [[testing]] — 测试目录结构、Mock Store/Auth、集成测试套件与测试规范 | quality, automation, testing, mock

### 测试用例集（quality/testcases/）

- [[console-client-auth-e2e-testcases]] — Console API、Client 与权限接口 E2E 覆盖矩阵 | quality, testcases, e2e, api, console, client, auth

---

## 如何快速定位

| 我想了解… | 读哪里 |
|-----------|--------|
| 这个项目是什么，能做什么 | [[overview]] |
| 代码整体结构和层次 | [[architecture]] |
| 领域术语和核心实体 | [[terminology]]、[[domain-models]] |
| 命名空间如何管理多租户 | [[namespace]] |
| 服务发现与健康检查 | [[service-discovery]] |
| 配置文件发布与 Watch | [[config-center]] |
| 路由/限流/熔断等治理规则 | [[governance-rules]] |
| 治理规则统一存储方案 | [[adr-governance-rule-unified-storage]] |
| 管理后台运维操作 | [[admin]] |
| MCP 如何运作 | [[ai-features]] |
| 数据库表/查询是怎样的 | [[storage]] |
| 内存缓存如何刷新 | [[cache-layer]] |
| 认证 Token 如何校验 | [[auth-system]] |
| HTTP/gRPC 端点在哪里 | [[api-servers]] |
| 公共工具包（日志、EventHub） | [[common-infra]] |
| YAML 配置项含义 | [[configuration]] |
| 代码约定和惯用模式 | [[patterns]] |
| 如何写/跑测试 | [[testing]] |
| wiki 本身的维护规范 | [[schema]] |

## 相关页面

- [[schema]]
- [[log]]
