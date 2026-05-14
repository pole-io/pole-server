---
title: 知识库内容目录
tags: [meta, index]
links: [schema, log]
updated: 2026-05-14
sources: 0
---

# pole-control-plane 知识库

本目录是 `context-kg/` wiki 的入口。所有页面均使用 [[schema]] 中定义的格式规范。操作历史见 [[log]]。

---

## 概览类

- [[overview]] — 项目介绍、技术选型、顶层目录结构与启动流程 | overview
- [[configuration]] — YAML 配置文件结构、插件配置与部署目录说明 | config, yaml, deploy

## 架构类

- [[architecture]] — 四层架构设计、插件系统、Store 接口、CacheManager、拦截器链与请求流程 | architecture, design
- [[patterns]] — 代码库中的 11 种关键模式与约定（插件注册、单例、拦截器、缓存、软删除等） | patterns, conventions

## 领域组件类

- [[domain-components]] — 命名空间、服务发现、配置中心、治理规则、管理后台、Skill Hub 六大业务域 | domain, service, config, skill
- [[ai-features]] — MCP Registry 与 Skill Hub 的领域模型、存储、HTTP API 与缓存设计 | ai, mcp, skill
- [[auth-system]] — 认证（Authentication）与授权（Authorization）插件接口、拦截器链与 Token 流程 | auth, security

## 基础设施类

- [[storage]] — 存储接口层次结构、MySQL 实现特性（软删除、增量查询、分布式锁）与 Mock Store | storage, mysql, database
- [[cache-layer]] — 19 种缓存子类型、增量刷新循环、AI 缓存子包与按需开启机制 | cache, performance
- [[api-servers]] — HTTP、gRPC、xDS、Nacos、Apollo、Eureka 六类协议服务端与 MCP 集成 | api, http, grpc, xds, nacos
- [[common-infra]] — 日志、EventHub、Batch Controller、OTel 指标、同步原语与通用工具 | infra, logging, eventhub

## 开发指南类

- [[testing]] — 测试目录结构、Mock Store/Auth、集成测试套件与测试规范 | testing, mock

---

## 如何快速定位

| 我想了解… | 读哪里 |
|-----------|--------|
| 这个项目是什么，能做什么 | [[overview]] |
| 代码整体结构和层次 | [[architecture]] |
| 某个业务功能在哪里实现 | [[domain-components]] |
| MCP / Skill Hub 如何运作 | [[ai-features]] |
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
