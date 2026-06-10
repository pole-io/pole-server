---
title: 知识库内容目录
tags: [meta, index]
links: [schema, log]
updated: 2026-06-10
sources: 0
---

# pole-control-plane 知识库

本目录是 `context-kg/` wiki 的入口。所有页面均使用 [[schema]] 中定义的格式规范。操作历史见 [[log]]。

知识库按五大类别组织：全局视角、业务域、技术基础设施、AI 原生能力、开发指南。

---

## 全局视角（overview/）

- [[overview]] — 项目介绍、技术选型、顶层目录结构与启动流程 | overview
- [[architecture]] — 四层架构设计、插件系统、Store 接口、CacheManager、拦截器链与请求流程 | architecture, design
- [[api-servers]] — HTTP、gRPC、xDS、Nacos、Apollo、Eureka 六类协议服务端与 MCP 集成 | api, http, grpc, xds, nacos
- [[configuration]] — YAML 配置文件结构、插件配置与部署目录说明 | config, yaml, deploy

## 业务域（domains/）

- [[namespace]] — 多租户命名空间：关键常量、接口定义与 singleflight 防重复创建 | domain, namespace
- [[service-discovery]] — 服务发现核心：DiscoverServer 接口、批处理、健康检查、空推保护 | domain, service, healthcheck
- [[config-center]] — 配置中心：版本化配置文件、灰度发布、Watch 长轮询机制 | domain, config
- [[governance-rules]] — 治理规则：路由、限流、熔断、故障探测、泳道、无损规则 | domain, governance, routing, ratelimit
- [[admin]] — 管理后台：AdminOperateServer 接口与跨域运维操作 | domain, admin

## 技术基础设施（infra/）

- [[storage]] — 存储接口层次结构、MySQL 实现特性（软删除、增量查询、分布式锁）与 Mock Store | storage, mysql, database
- [[cache-layer]] — 缓存子类型、增量刷新循环、AI 缓存子包与按需开启机制 | cache, performance
- [[auth-system]] — 认证（Authentication）与授权（Authorization）插件接口、拦截器链与 Token 流程 | auth, security
- [[common-infra]] — 日志、EventHub、Batch Controller、OTel 指标、同步原语与通用工具 | infra, logging, eventhub

## AI 原生能力（ai/）

- [[ai-features]] — MCP Registry 与 A2A Agent Registry 领域模型、存储、HTTP API 与缓存设计 | ai, mcp, a2a

## 开发指南（guides/）

- [[patterns]] — 代码库中的 11 种关键模式与约定（插件注册、单例、拦截器、缓存、软删除等） | patterns, conventions
- [[testing]] — 测试目录结构、Mock Store/Auth、集成测试套件与测试规范 | testing, mock

---

## 如何快速定位

| 我想了解… | 读哪里 |
|-----------|--------|
| 这个项目是什么，能做什么 | [[overview]] |
| 代码整体结构和层次 | [[architecture]] |
| 命名空间如何管理多租户 | [[namespace]] |
| 服务发现与健康检查 | [[service-discovery]] |
| 配置文件发布与 Watch | [[config-center]] |
| 路由/限流/熔断等治理规则 | [[governance-rules]] |
| 管理后台运维操作 | [[admin]] |
| MCP / A2A Registry 如何运作 | [[ai-features]] |
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
