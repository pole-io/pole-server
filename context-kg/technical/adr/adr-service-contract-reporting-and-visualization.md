---
title: 四协议服务契约上报、发现与可视化
tags: [adr, service-contract, openapi, grpc, dubbo, thrift]
links: [service-discovery, api-servers, storage, cache-layer]
updated: 2026-07-26
sources: 14
---

# 四协议服务契约上报、发现与可视化

## 背景与目标

服务契约此前已经存在 protobuf、MySQL 表和部分管理接口，但缺少一致的上报约定、统一发现分发和 Console 展示，且存在缓存 miss 返回空对象、字段兼容丢失、软删除详情回流、SDK 全量上报覆盖人工数据等问题。

本方案闭环支持四类同步 RPC 契约：

| 服务协议 | 契约类型 | 原始内容 | 归一化接口 |
|---|---|---|---|
| HTTP | OpenAPI 3.x | OpenAPI JSON/YAML | `path=HTTP path`，`method=HTTP method` |
| gRPC | protobuf | `.proto` 文本或描述信息 | `path=全限定 Service`，`method=RPC Method` |
| Dubbo | dubbo | 接口描述或生成产物 | `path=全限定接口`，`method=方法名` |
| Thrift | thrift | Thrift IDL | `path=Service`，`method=Function` |

## 决策

### 1. 使用调用方推送，不由控制面主动抓取

- SDK 使用 specification 已定义的 gRPC `ReportServiceContract` 上报。
- Agent、Sidecar 或 CI 使用 `POST /naming/v1/ReportServiceContract` 上报同一 `ServiceContract` 模型。
- Console 人工维护保留兼容的两步接口：先用 `POST /naming/v1/service/contracts` 创建主体，再用 `POST /naming/v1/service/contract/methods` 写入 `Manual` 接口；不会把旧 CRUD 强制改成全量发布。

控制面不主动扫描 OpenAPI URL，也不主动连接 gRPC Reflection、Dubbo Metadata Center 或 Thrift 服务。主动抓取会引入 SSRF、网络可达性、凭证托管、超时重试和生产服务额外负载，不能作为默认控制面职责。部署侧可自行实现采集器，再通过统一 HTTP 上报入口推送。

### 2. 上报载荷兼顾原文和统一投影

统一上报入口要求所有协议都提交非空的 `namespace`、`service`、`protocol`、`version`、`content` 和 `interfaces`。服务端统一将协议和契约类型转为小写，并同时写入兼容字段 `type` 与 `name`。

- HTTP/OpenAPI 允许省略 `interfaces`；服务端始终校验原文并从 OpenAPI 3.x `paths` 中稳定抽取有效 HTTP operation，调用方不能用自带接口清单绕过原文校验。
- gRPC、Dubbo、Thrift 必须提交结构化 `interfaces`。编译插件或 SDK 比控制面更接近真实构建产物，能正确处理 import、代码生成、重载和框架扩展。
- 结构化接口的 `type` 承载方法签名；接口 ID 和 Manual 覆盖键使用 `path + method + type`，确保 Dubbo/Thrift 重载方法不会互相覆盖。
- 契约主体和接口均计算内容摘要；四协议最终共享相同的查询与展示投影。
- 契约 ID 和接口 ID 必须与自然键计算结果一致，拒绝调用方借用其他服务 ID；主体创建使用确定 ID 的幂等 upsert，可承受主从复制延迟窗口内的立即重试。
- 非法协议、非 OpenAPI 3.x 文档或缺少 `path/method` 的结构化接口直接返回参数错误。

### 3. Manual 与 Client 分源替换

接口记录以 `source=Manual|Client` 区分来源。一次全量上报只删除并重建同来源记录：

- SDK 重报不会删除 Console 人工补充的接口。
- 人工重新上传不会删除 SDK 自动发现的接口。
- 查询投影中相同 `path/method` 由 Manual 覆盖 Client，其余 Client 接口继续展示。
- 软删除的接口在详情查询、列表查询和缓存增量查询中均不可回流。
- 历史 `source=0` SDK 行按 Client 兼容读取并在下一次 Client 全量上报时收敛。

契约主体与接口是顺序写入，接口写入失败时调用方应按同一自然键重试；主体 upsert 与同来源接口替换均幂等。公开的人工接口写入在主体查询不到时返回 NotFound，只有刚完成主体 upsert 的统一 publish 内部路径允许从库暂时不可见。未来若引入不可变 artifact/snapshot 表，再将两步写入升级为单事务发布。

### 4. 查询与可视化使用同一份投影

- 客户端通过统一 `Discover` 请求的 `SERVICE_CONTRACTS` 类型发现某服务的全部有效契约；响应修订号由各契约缓存键和 revision 稳定计算，支持 `DataNoChange`。
- Console 使用 `/naming/v1/service/contracts` 和 `/naming/v1/service/contract/versions`。
- 服务详情新增“服务契约”页签，按版本和契约切换，统一展示四协议能力卡、接口清单、来源、状态与原始契约。
- 管理查询在已指定 namespace/service 时按目标服务执行授权检查。

## 兼容与演进

- `name` 是历史字段，`type` 是当前字段；读写和 ID 计算都采用 `type` 优先、`name` 兜底。
- 当前沿用既有 `service_contract` 与 `service_contract_detail` 表，避免为展示能力引入不必要迁移。
- 后续若需要 schema diff、审批发布、回滚和多制品引用，可在 V2 引入不可变 Contract Artifact 与 Snapshot；现有自然键、统一上报入口和发现投影保持兼容。

## 验收标准

- 四协议均能通过统一载荷上报，HTTP 可从 OpenAPI 3.x 自动提取接口。
- gRPC 与 HTTP 客户端入口可达，`SERVICE_CONTRACTS` 能通过统一 Discover 返回并正确处理 revision。
- 缓存 miss 返回 `nil`，多版本列表稳定，软删除详情不会出现。
- Client 与 Manual 全量替换互不覆盖。
- Console 可真实查询并展示四协议的版本、接口和原始内容。
- Go 单元测试、MySQL Store 测试、Console 专项契约、ESLint 和生产构建全部通过。

## 相关页面

- [[service-discovery]]
- [[api-servers]]
- [[storage]]
- [[cache-layer]]
