---
title: API 服务端 — 协议实现
tags: [api, http, grpc, xds, nacos]
links: [architecture, ai-features, auth-system]
updated: 2026-05-14
sources: 1
---

# API 服务端 — 协议实现

## 概览

所有 API 服务端均实现 `apis/apiserver/Apiserver` 接口。整体分层设计见 [[architecture]]，AI MCP 集成详情见 [[ai-features]]，认证中间件见 [[auth-system]]。

```go
type Apiserver interface {
    GetProtocol() string
    GetPort() int
    Initialize(ctx context.Context, option map[string]interface{}, apiConf APIConfig) error
    Run(errCh chan error)
    Stop()
    Restart(option map[string]interface{}, apiConf APIConfig) error
}
```

每个服务端通过 `init()` 自行注册，并在配置中通过名称选择使用哪个实现。

---

## HTTP 服务端（`plugin/apiserver/httpserver/`）

基于 `go-restful/v3` 的主 REST API。负责处理：
- 服务发现 REST API
- 配置中心 REST API
- 治理规则管理
- 认证管理（用户、策略）
- 管理后台操作
- AI MCP 服务器管理
- OpenAPI/Swagger 文档

**关键子包：**
```
httpserver/
├── server.go           # 主服务端，路由注册
├── discover/           # 服务发现端点
├── config/             # 配置中心端点
├── auth/               # 认证端点（用户、用户组、策略）
├── aimcp/              # AI MCP 服务器管理端点
│   ├── server.go
│   └── mcp_server.go
├── utils/              # 共享 HTTP 工具函数
├── i18n/               # 国际化
└── docs/               # Swagger 规范生成
```

**MCP 端点**（`plugin/apiserver/httpserver/aimcp/`）：
- `GET /mcp/v1/servers` — 列出 MCP 服务器（支持过滤）
- `POST /mcp/v1/servers` — 创建 MCP 服务器
- 通过 MCP 协议暴露工具，供 AI 智能体消费

---

## gRPC 服务端（`plugin/apiserver/grpcserver/`）

面向服务 SDK 客户端的高性能二进制协议。

**子包：**
```
grpcserver/
├── discover/
│   └── v1/         # 服务发现 gRPC 处理器
└── utils/          # gRPC 工具函数（元数据、错误映射）
```

实现了来自 `pole-io/specification` protobuf 定义的 `polaris.Discover` 和 `polaris.Heartbeat` gRPC 服务。

---

## xDS 服务端 v3（`plugin/apiserver/xdsserverv3/`）

实现 Envoy xDS API v3，用于 Envoy 代理集成：
- LDS（监听器发现服务）
- CDS（集群发现服务）
- EDS（端点发现服务）
- RDS（路由发现服务）
- ADS（聚合发现服务）

以 `envoyproxy/go-control-plane` 为基础框架。

**子包：**
```
xdsserverv3/
├── cache/     # xDS 快照缓存，将 Pole 资源映射为 xDS proto
└── resource/  # xDS 资源构建器（监听器、集群、端点、路由）
```

转换关系：Pole 服务/实例/路由规则 → Envoy xDS 资源

---

## Nacos 服务端（`plugin/apiserver/nacosserver/`）

为 Nacos 客户端提供完整的 Nacos 协议兼容。

**子包：**
```
nacosserver/
├── core/        # Nacos 核心逻辑
├── model/       # Nacos 数据模型
├── logger/      # Nacos 专属日志
├── v1/          # Nacos v1 API
│   ├── config/  # 配置中心（v1）
│   ├── discover/# 服务发现（v1）
│   └── http/    # HTTP 处理器
└── v2/          # Nacos v2 API（基于 gRPC）
    ├── config/
    ├── discover/
    ├── pb/      # Nacos v2 protobuf
    └── remote/  # gRPC 连接管理
```

将 Nacos 格式的请求转换为 Pole 内部格式。

---

## Apollo 服务端（`plugin/apiserver/apolloserver/`）

为 Apollo 配置中心客户端提供协议兼容。将 Apollo SDK 指向该服务端后，客户端即可从 Pole 的配置中心获取配置。

---

## Eureka 服务端（`plugin/apiserver/eurekaserver/`）

Netflix Eureka 协议兼容。使用 Eureka 服务发现的 Spring Cloud 应用程序，可通过 Pole 进行注册和服务发现。

---

## AI MCP 集成（`plugin/apiserver/httpserver/aimcp/`）

将 Pole 的服务注册中心能力暴露为 MCP 工具，供 AI 智能体（Claude、GPT 等）使用。详细说明见 [[ai-features]]。

**已注册的 MCP 工具：**
- `list_mcp_servers` — 查询已注册的 MCP 服务器
- `create_mcp_servers` — 注册新的 MCP 服务器

使用 `mark3labs/mcp-go` 库处理 MCP 协议。

## 相关页面

- [[architecture]]
- [[ai-features]]
- [[auth-system]]
