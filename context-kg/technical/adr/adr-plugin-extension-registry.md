---
title: 官方内置插件与用户扩展 Registry
tags: [adr, plugin, registry, extension, runtime]
links: [architecture, patterns, adr-unified-process-mode-and-limiter-integration]
updated: 2026-07-31
sources: 12
---

# 官方内置插件与用户扩展 Registry

## 状态

已接受并完成首期实现。进程内用户插件采用独立 Go Module、显式 Factory 注册和自定义
发行版；需要独立升级与故障隔离的动态插件后续使用进程外 RPC，不采用 Go `plugin`
作为产品主路径。

## 背景

历史插件体系由多套包级注册表组成：

- 通用 `Plugin` 使用 `apis.RegisterPlugin`；
- Store 使用 `apis/store.RegisterStore`；
- API Server 使用 `apis/apiserver.Register`；
- Auth User 与 Strategy 使用各自的 slot map；
- 根 `package main` 通过空导入触发各实现包的 `init()`。

这种方式可以完成官方单二进制装配，但外部用户必须 fork 根组合文件或复制全部空导入
列表。注册来源、冲突和生命周期不可集中验证，Store、API Server 与 Auth 也没有共享
同一注册契约。

## 决策

### 代码所有权与包边界

- `pluginapi/` 提供稳定的 Registry、Descriptor、Factory、Kind、Origin 与运行期激活契约；
- `builtinplugins/` 是可导入的官方组合 Module，集中注册 Pole 自带实现；
- `plugin/` 在兼容期继续保存官方实现源码，但实现包改为暴露 `Register(registry)`，不再
  自行通过 `init()` 注册；
- 用户插件放在用户自己的 Go Module，例如
  `example.com/company/pole-plugins/observability/foo`，不在 Pole 主仓库建立
  `customplugins/` 目录；
- 后续可以把官方实现迁入 `internal/builtin/`，外部用户只依赖
  `builtinplugins/` 与 `pluginapi/`，因此迁移不改变外部装配入口。

### Registry 契约

Registry 以 `kind + name` 作为唯一键，保存 Descriptor 与 Factory：

```go
type Descriptor struct {
    Kind       Kind
    Name       string
    Version    string
    APIVersion string
    Origin     Origin
}

type Factory func() (any, error)
```

不变量：

- 同一 `kind + name` 不允许覆盖，用户插件不能静默替换官方插件；
- Factory 在每个运行 Registry 中最多执行一次，同一运行期所有消费者得到同一实例；
- `Clone()` 复制 Catalog 但不复制运行实例，新的 Bootstrap 获得独立实例集；
- Registry 启动前必须 `Freeze()`，冻结后禁止注册；
- 同一进程不允许两个 Bootstrap 重叠激活 Registry；
- Factory 错误和类型不匹配作为显式错误返回，不使用静默覆盖；
- 未被配置选择的插件不实例化，其 Factory 失败不影响当前 Profile 启动。
- Registry 关闭后禁止继续解析；已解析且实现 `Destroy() error` 的实例按解析顺序逆序
  销毁，未解析实例不执行 Factory 或 Destroy；
- 多个销毁错误通过错误链统一返回，Bootstrap 在恢复此前 Active Registry 前完成关闭。

### 官方与用户装配

官方发行版：

```go
registry, err := builtinplugins.NewRegistry()
if err != nil {
    return err
}
cmd.ExecuteWithPluginRegistry(registry)
```

用户自定义发行版：

```go
registry, err := builtinplugins.NewRegistry()
if err != nil {
    return err
}
if err := customerplugin.Register(registry); err != nil {
    return err
}
cmd.ExecuteWithPluginRegistry(registry)
```

Bootstrap 克隆传入 Catalog、冻结运行副本、独占激活，并在退出或启动失败时恢复此前
Registry。为兼容旧注册入口，显式 Catalog 在冻结前还会合并默认 Registry；同一
`kind + name` 冲突会直接终止启动，不允许兼容注册覆盖官方实现。通用插件、Store、
API Server、Auth User、Auth Strategy、RateLimit 与 Whitelist 的消费侧都从同一
Active Registry 解析。

### 兼容策略

- 旧 `RegisterPlugin`、`RegisterStore`、`apiserver.Register`、
  `RegisterUserServer` 与 `RegisterStrategyServer` 保留，转发到默认 Registry；
- 官方集合不再通过 `init()` 注入默认 Registry；自定义发行版必须显式调用
  `builtinplugins.Register` 或 `builtinplugins.NewRegistry`；
- 用户已有实现仍可暂时使用旧 `Register*` 函数注册到默认 Registry，但迁移目标是显式
  `Register(registry)`；
- 旧入口接收的是已构造实例，因此只保证兼容，不承诺跨 Bootstrap 的实例隔离；需要运行
  实例隔离的实现必须迁移到 `RegisterFactory` 或模块自己的 `Register(registry)`；
- 根 `plugin.go` 不再维护官方插件空导入清单，只保留缓存和拦截器等历史内部装配；
- `apiserver.Slots` 暂时保留为已弃用的运行实例投影，内部读取统一使用加锁快照，不再
  把它作为注册真相；
- `store.StoreSlots` 只保留旧入口实例投影，Auth 的旧内部 slot map 已移除，重复注册
  统一由 Registry 判断；
- Store 与 Auth 的初始化状态按 Active Registry 隔离，不再使用跨 Registry 的
  `sync.Once`。

## 动态扩展边界

首期只支持进程内静态编译插件。它适合延迟敏感、需要直接使用 Go Interface 的扩展，
但用户必须重新构建自定义 Pole 发行版，并把第三方依赖纳入 SBOM 与漏洞扫描。

真正需要独立安装、独立升级、跨语言或故障隔离时，使用版本化 RPC/gRPC 插件协议：

- 优先试点 History、Discover Event、Observability Sink、CMDB 和异步 Connector；
- Store、核心 Auth、Crypto 与 HealthChecker 保持进程内，除非先形成更窄且可序列化的
  Interface；
- 用户自定义协议接入优先作为独立 Gateway，不允许第三方代码直接向官方 HTTP Server
  注入任意路由；
- RPC 插件必须定义 capability negotiation、超时、背压、健康检查、mTLS、allowlist
  与失败策略。

不采用 Go `plugin.Open` 作为正式扩展机制。它要求主程序与插件使用高度一致的 Go
工具链、依赖源码和构建参数，不能提供可靠卸载、隔离或跨版本兼容。

## 后续迁移

1. 给 `pluginapi` 建立独立语义版本与兼容测试；
2. 为不同 Kind 增加类型化配置 schema，逐步减少 `map[string]interface{}`；
3. 将官方实现迁入 `internal/builtin/`，保持 `builtinplugins.Register` 不变；
4. 移除旧 slot map 和兼容 `init()`，让 Registry 成为唯一装配真相；
5. 以 Observability Sink 试点进程外插件协议。

## 证据

- `pluginapi/registry.go`
- `builtinplugins/builtin.go`
- `apis/plugin.go`
- `apis/store/store.go`
- `apis/apiserver/apiserver.go`
- `apis/access_control/auth/auth.go`
- `bootstrap/process.go`
- `bootstrap/server.go`
- `cmd/root.go`
- `cmd/start.go`
- `main.go`
- `plugin.go`

## 相关页面

- [[architecture]]
- [[patterns]]
- [[adr-unified-process-mode-and-limiter-integration]]
