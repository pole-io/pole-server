---
title: Pole Control Plane — 项目概览
tags: [overview]
links: [architecture, domain-components, configuration]
updated: 2026-05-14
sources: 1
---

# Pole Control Plane — 项目概览

## 项目介绍

**pole-control-plane**（模块名：`github.com/pole-io/pole-server`）是一个云原生、AI 原生的服务治理控制平面。它基于 Polaris（腾讯开源的服务网格控制平面）演进而来，在其基础上扩展了 AI 原生能力（MCP 协议、Skill Hub）以及增强的治理功能。详细架构设计见 [[architecture]]，各业务域实现见 [[domain-components]]，部署配置见 [[configuration]]。

服务端管理以下能力：
- **服务发现** — 服务实例的注册、注销与查询
- **配置中心** — 支持灰度发布的版本化配置文件管理
- **治理规则** — 路由、限流、熔断、故障探测、泳道规则
- **命名空间管理** — 多租户资源隔离
- **AI 原生功能** — MCP 服务器注册中心、Skill Hub（函数/工具/智能体管理）
- **多协议兼容** — 支持 Polaris gRPC、REST、Nacos、Apollo、Eureka、xDS/Envoy

## 核心技术选型

| 关注点 | 选型 |
|--------|------|
| 编程语言 | Go 1.25 |
| HTTP 框架 | `go-restful/v3`（支持 OpenAPI） |
| gRPC | `google.golang.org/grpc` |
| 数据库 | MySQL（通过 `go-sql-driver`） |
| xDS/Envoy | `envoyproxy/go-control-plane` v0.13 |
| MCP 协议 | `mark3labs/mcp-go` |
| CLI | `spf13/cobra` |
| 日志 | `go.uber.org/zap` |
| 指标监控 | OpenTelemetry（OTLP gRPC 导出） |
| API 规范 | `github.com/pole-io/specification`（protobuf） |
| 内嵌 KV | `go.etcd.io/bbolt` |

## 顶层目录结构

```
pole-control-plane/
├── main.go                  # 程序入口 → cmd.Execute()
├── plugin.go                # 空导入所有插件以触发 init()
├── cmd/                     # Cobra CLI（root、start、version）
├── bootstrap/               # 服务初始化编排器
│   └── config/              # YAML 配置结构定义
├── apis/                    # 共享接口（plugin、store、cache、auth、apiserver）
│   ├── pkg/types/           # 领域模型结构体（service、config、auth、ai、rules）
│   ├── store/               # 按领域分解的存储接口
│   ├── cache/               # CacheManager 接口
│   ├── access_control/      # 认证插件接口
│   └── apiserver/           # Apiserver 插件接口
├── pkg/                     # 业务逻辑实现
│   ├── service/             # 服务发现 + 健康检查
│   ├── config/              # 配置中心
│   ├── namespace/           # 命名空间管理
│   ├── goverrule/           # 治理规则
│   ├── admin/               # 管理后台操作
│   ├── skill/               # AI Skill Hub 逻辑
│   ├── cache/               # 缓存管理器实现
│   └── common/              # 公共工具（log、eventhub、batchctrl、otel、syncs）
├── plugin/                  # 插件实现
│   ├── apiserver/           # HTTP、gRPC、xDS、Nacos、Apollo、Eureka 服务端
│   ├── store/mysql/         # MySQL 存储后端
│   ├── access_control/      # 认证（用户、策略、限流、白名单）
│   ├── crypto/              # AES/RSA 加密
│   ├── observability/       # 历史记录、统计、发现事件
│   ├── service/healthchecker/ # 健康检查心跳插件
│   └── cmdb/                # CMDB 插件（内存实现）
├── context-kg/              # 知识库（当前目录）
├── deploy/                  # 部署配置和脚本
├── docs/                    # 设计文档
└── test/                    # 集成测试与测试套件
```

## 功能路线图状态（来自 README）

### AI 原生
- [x] MCP 协议支持 — AI 智能体可通过 MCP 发现服务
- [ ] MCP Registry API — 管理 MCP 服务器
- [ ] Skill Hub — 类配置中心风格的技能管理，支持分组

### 服务发现
- [x] Apollo 协议兼容
- [x] 自适应空推送保护
- [ ] 全局集群模式
- [ ] 服务订阅者查询
- [ ] 服务拓扑视图

### 治理
- [x] 所有规则支持版本控制 + 灰度发布
- [x] 细粒度访问控制（类似云端 CAM/RAM）

## 启动流程（高层次）

```
main.go → cmd.Execute() → cmd/start.go → bootstrap.Start(configFile)
  1. 加载 YAML 配置
  2. 初始化日志
  3. 获取本机 IP
  4. 初始化指标（OTel）
  5. 初始化 EventHub
  6. 加载插件配置
  7. 初始化存储层（MySQL）
  8. 获取启动锁（防止集群中并行初始化）
  9. 预热缓存（全部 19 种资源类型）
 10. 初始化认证（用户服务 + 策略服务）
 11. 初始化命名空间
 12. 初始化服务发现 + 健康检查 + BatchController
 13. 初始化治理规则
 14. 初始化配置中心
 15. 初始化管理后台
 16. 启动 API 服务端（goroutine：HTTP、gRPC、xDS、Nacos、Apollo、Eureka）
 17. 自注册为服务实例
 18. 释放启动锁
 19. WaitSignal() — 阻塞，直到收到 SIGTERM/SIGINT
```

## 相关页面

- [[architecture]]
- [[domain-components]]
- [[configuration]]
