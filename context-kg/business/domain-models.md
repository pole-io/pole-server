---
title: 核心业务实体
tags: [business, domain-model]
links: [terminology, namespace, service-discovery, config-center, governance-rules, ai-features]
updated: 2026-06-09
sources: 0
---

# 核心业务实体

| 实体 | 归属页面 | 说明 |
|------|----------|------|
| Namespace | [[namespace]] | 顶层多租户隔离单元 |
| Service | [[service-discovery]] | 服务发现核心实体，承载实例、订阅和治理关系 |
| Instance | [[service-discovery]] | 服务运行节点，按健康、隔离、位置和元数据参与发现 |
| Config File | [[config-center]] | 版本化配置内容，支持发布、回滚和监听 |
| Governance Rule | [[governance-rules]] | 治理能力统称，包含路由、限流、熔断、探测、无损、泳道 |
| MCP Server | [[ai-features]] | AI 原生能力中的 MCP 服务注册实体 |

实体之间的核心关系：

```text
Namespace
  -> Service
    -> Instance
    -> Governance Rule
    -> MCP Server backend selector
  -> Config File Group
    -> Config File
      -> Release
```

## 相关页面

- [[terminology]]
- [[namespace]]
- [[service-discovery]]
- [[config-center]]
- [[governance-rules]]
- [[ai-features]]
