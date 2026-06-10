---
title: 领域术语表
tags: [business, terminology]
links: [domain-models, business-rules]
updated: 2026-06-09
sources: 0
---

# 领域术语表

| 中文术语 | English | 含义 |
|----------|---------|------|
| 命名空间 | Namespace | 多租户隔离单元，服务、配置、治理规则等资源都归属于某个命名空间 |
| 服务 | Service | 服务发现中的逻辑服务名，是实例、订阅、治理规则的主要归属对象 |
| 实例 | Instance | 服务的运行节点，包含 host、port、协议、健康状态、隔离状态和元数据 |
| 配置文件 | Config File | 配置中心管理的版本化配置内容 |
| 治理规则 | Governance Rule | 路由、限流、熔断、故障探测、无损上下线、泳道等规则的统称 |
| 发布版本 | Release | 配置或治理规则发布时形成的快照版本 |
| 灰度发布 | Gray Release | 面向部分客户端、标签或条件生效的发布方式 |
| 泳道组 | Lane Group | 一组泳道路由规则的聚合根 |
| MCP 服务 | MCP Server | AI 工具注册中心中的 MCP 服务定义，可关联 Pole 服务或自定义地址 |

## 相关页面

- [[domain-models]]
- [[business-rules]]
