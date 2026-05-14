---
title: 操作日志
tags: [meta, log]
links: [index, schema]
updated: 2026-05-14
sources: 0
---

# 操作日志

追加式记录，只增不改。最新记录在文件末尾。格式规范见 [[schema]]。

---

## [2026-05-14] ingest | pole-control-plane 全量初始化

- 扫描 656 个 Go 文件（模块：`github.com/pole-io/pole-server`）
- 生成 12 个知识页面，覆盖以下领域：
  - 概览（overview）：项目介绍、技术选型、目录结构、启动流程
  - 架构（architecture）：四层架构、插件系统、Store/Cache 接口、拦截器链
  - 业务领域（domain-components）：命名空间、服务发现、配置中心、治理规则、管理后台、Skill Hub
  - API 服务端（api-servers）：HTTP、gRPC、xDS v3、Nacos v1/v2、Apollo、Eureka、MCP 集成
  - 存储层（storage）：接口层次、MySQL 实现、软删除、增量查询、分布式锁
  - 缓存层（cache-layer）：19 种缓存类型、1 秒刷新循环、-5 秒安全窗口
  - 认证系统（auth-system）：UserServer、StrategyServer、拦截器链、Token 流程
  - AI 功能（ai-features）：MCP Registry、Skill Hub、领域模型、HTTP API、缓存
  - 公共基础设施（common-infra）：日志、EventHub、Batch Controller、OTel、同步原语
  - 配置参考（configuration）：YAML 结构、插件配置、API 服务端配置、部署目录
  - 模式与约定（patterns）：11 种关键代码模式
  - 测试（testing）：Mock Store/Auth、集成套件、测试规范
- 首次建立 wiki 结构，新建元文件：schema.md、index.md、log.md
- 替换旧 README.md 为 index.md

## 相关页面

- [[index]]
- [[schema]]
