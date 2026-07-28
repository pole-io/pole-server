---
title: Pole Control Plane — 项目概览
tags: [overview]
links: [architecture, index, configuration]
updated: 2026-07-29
sources: 1
---

# Pole Control Plane — 项目概览

## 项目介绍

**pole-control-plane**（模块名：`github.com/pole-io/pole-server`）是一个云原生、AI 原生的服务治理控制平面。它基于 Polaris（腾讯开源的服务网格控制平面）演进而来，在其基础上扩展了 AI 原生能力（MCP 协议）以及增强的治理功能。详细架构设计见 [[architecture]]，各业务域入口见 [[index]]，部署配置见 [[configuration]]。

服务端管理以下能力：
- **服务发现** — 服务实例的注册、注销与查询
- **配置中心** — 支持灰度发布的版本化配置文件管理
- **治理规则** — 路由、限流、熔断、故障探测、泳道规则
- **命名空间管理** — 运行环境与资源隔离
- **AI 原生功能** — MCP 服务器注册中心
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
| 内嵌 KV | `github.com/cockroachdb/pebble` |

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
│   ├── console/             # Console Go Module、内嵌静态资源与私有实现
│   ├── limiter/             # Limiter Module、协议适配与限流核心
│   ├── service/             # 服务发现 + 健康检查
│   ├── config/              # 配置中心
│   ├── namespace/           # 命名空间管理
│   ├── goverrule/           # 治理规则
│   ├── admin/               # 管理后台操作
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
├── context-kg/              # 知识库与任务过程记录
├── deploy/                  # 部署配置和脚本
├── test/                    # 集成测试与测试套件
└── web/console/             # React/Vite Console 源码；由 Go 制品流程统一构建
```

## 功能路线图状态（来自 README）

### AI 原生
- [x] MCP 协议支持 — AI 智能体可通过 MCP 发现服务
- [ ] MCP Registry API — 管理 MCP 服务器

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
main.go → cmd.Execute() → cmd/start.go → bootstrap.Run(options)
  1. 加载 YAML，并解析 CLI `--mode` 覆盖后的运行 Profile
  2. 校验选中模块配置和 listener 冲突
  3. 构造 Control Plane、Limiter、Console Module
  4. Supervisor 按 Profile 顺序启动模块并等待 readiness
  5. 任一模块启动或运行失败时回滚；收到 SIGTERM/SIGINT 后逆序停止

Console 静态资源在 release/test 构建中由 `web/console` 生成并复制到
`pkg/console/internal/assets/dist`，随后通过 `go:embed` 进入同一个 Go 二进制。
```

## 相关页面

- [[architecture]]
- [[index]]
- [[configuration]]
