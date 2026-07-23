---
title: 实例主动健康检查设计
tags: [adr, service, healthcheck, tcp, http]
links: [service-discovery, storage, configuration]
updated: 2026-07-12
sources: 10
---

# 实例主动健康检查设计

## 背景

通过 SDK 注册的实例可以持续上报心跳，Console 手工创建的实例没有 SDK 心跳进程，因此不能把 `HEARTBEAT` 作为可选检查方式。原 Console 同时存在“心跳可选、TCP/HTTP 禁用”的占位交互，但实例协议、服务端插件和存储映射实际只支持心跳，直接放开选项会形成假功能。

## 决策

- Console 创建实例时保留心跳选项用于解释能力边界，但设置为禁用；默认选择 TCP 探测，TCP 与 HTTP 均可选择。
- 已有心跳实例在查看和编辑时继续正确回显，避免破坏 SDK 注册实例。
- `HealthCheck` 契约增加 `TCP`、`HTTP` 类型，以及各自的探测间隔配置；HTTP 额外包含请求路径，默认 `/`。
- 主动探测间隔默认 5 秒，网络超时由服务端探测插件统一控制为 3 秒。
- MySQL `health_check.ttl` 对心跳表示 TTL，对 TCP/HTTP 表示探测间隔；HTTP 路径存入实例保留元数据 `internal-healthcheck_path`，不新增数据库列。
- 健康检查调度器按类型读取间隔：心跳仍按三倍 TTL 判定过期，TCP/HTTP 每个间隔执行一次真实探测，并按探测结果更新实例健康状态。
- checker 节点选择优先使用健康节点；当没有健康节点时，允许非隔离节点兜底执行，避免 checker 自身未健康导致所有主动检查永久停摆。
- TCP 探测以成功建立 `host:port` 连接为健康；HTTP 探测向 `http://host:port/path` 发起 GET，请求返回 2xx 或 3xx 时判定健康。
- MySQL 行恢复必须先解析实例 metadata，再构造类型化 `HealthCheck`，否则 HTTP 路径会错误回退为默认 `/`。

## 边界

- 本能力是控制面针对具体实例地址的主动健康检查，不替代治理规则中的客户端主动探测 `FaultDetectRule`。
- 本期不开放 UDP、gRPC、MySQL 探测，也不提供自定义 HTTP Header、期望响应码或连续失败阈值。
- Console 表单不允许为手工创建实例选择心跳；API 和 SDK 仍可创建心跳实例。

## 验证要求

- 契约生成代码必须包含 HEARTBEAT/TCP/HTTP 三类结构。
- TCP/HTTP 插件必须使用真实监听端口和 HTTP 测试服务覆盖健康、异常和状态保持场景。
- MySQL round-trip 必须保持检查类型、间隔和 HTTP 路径。
- 浏览器创建请求必须按选中类型发送对应结构，且心跳选项不可选。

## 相关页面

- [[service-discovery]]
- [[storage]]
- [[configuration]]
