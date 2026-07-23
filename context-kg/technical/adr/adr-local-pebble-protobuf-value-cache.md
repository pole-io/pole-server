---
title: ADR：Pebble 本地 protobuf value cache
tags: [adr, cache, pebble, protobuf, performance]
links: [cache-layer, storage, service-discovery, config-center, governance-rules]
updated: 2026-07-19
sources: 0
---

# ADR：Pebble 本地 protobuf value cache

## 状态

Proposed。

当前已落地第一批：配置文件 active release content、服务契约本地 value、OTel event/audit 本地可靠队列。服务发现大响应与治理下发大快照仍待性能基准后接入。

## 背景

Pole 的服务发现、配置中心和治理规则都存在一类高频读取数据：事实来源在 MySQL、注册内存态或统一治理规则缓存中，但客户端读取时需要组装较大的响应对象，并反复执行 protobuf marshal。

这类数据包括：

- 服务实例列表和服务发现响应。
- 配置文件内容、发布版本和配置发现响应。
- 服务契约内容。
- 路由、限流、熔断、探测、鉴权、Mock、镜像等治理规则发布快照。

如果全部作为 Go 对象长期常驻内存，会放大堆占用和 GC 压力；如果每次请求都从对象重新编码，又会增加 CPU 与分配。用户明确希望本地磁盘缓存使用 Pebble 存放 protobuf marshal 后的 bytes，尤其覆盖实例列表和配置。

## 决策

引入一个可配置的进程级 Pebble 本地 value cache，用于保存可重建的客户端响应 protobuf bytes。MySQL、注册内存态和治理规则缓存仍是事实来源，Pebble 只承担本地低成本读取、低内存占用和减少重复编码的角色。

该能力必须以开关方式提供：

- 默认关闭，继续使用当前内存缓存与编码路径，避免默认性能和稳定性发生变化。
- 开启后采用“内存索引 + Pebble value”的模式，大部分面向客户端的大数据 value 从 Go heap 卸载到 Pebble。
- 开启后稳态性能不能低于当前内存路径；如果某类数据无法通过基准验证达到当前性能，不允许默认纳入卸载范围。

保留以下边界：

- 配置文件与服务契约启用 Pebble 后直接存放在 Pebble，内存只保留 namespace/group/file、service/protocol/version、revision、mtime、权限等轻量索引。
- OTel 操作审计与服务事件的本地可靠队列也使用 Pebble，但它是 append、ack、delete 模型，必须与客户端 value cache 使用不同抽象。
- 实例心跳状态继续保存在内存热路径，不对每次心跳同步写 Pebble。
- Pebble 中的 value 丢失、损坏或被清理后必须能从事实来源重建。

## 配置开关

建议在 bootstrap 配置中增加独立配置段：

```yaml
localValueCache:
  enabled: false
  backend: memory
  pebble:
    dataDir: ./data/cache/local-value.pebble
    maxSize: 4GiB
    minValueBytes: 4096
    warmup: true
    writeBufferSize: 64MiB
    maxOpenFiles: 256
  domains:
    serviceDiscovery: true
    config: true
    serviceContract: true
    governance: true

telemetry:
  localQueue:
    enabled: true
    backend: pebble
    pebble:
      dataDir: ./data/observability/otel-events.pebble
      maxSize: 1GiB
      batchSize: 512
      flushInterval: 1s
```

配置语义：

- `enabled=false` 时不得改变当前业务行为，不能在请求热路径引入 Pebble 调用。
- `backend=memory` 保留当前模式，可作为灰度、回退和基准对照。
- `backend=pebble` 才启用内存索引 + Pebble bytes。
- `minValueBytes` 用于避免小对象落盘后反而拖慢请求。
- `domains` 支持按领域灰度启用，便于逐类验证性能。
- `config` 和 `serviceContract` 开启后应直接以 Pebble 作为大 value 承载层，不再保留完整 value 的内存副本。
- `telemetry.localQueue` 是 OTel event/audit 本地可靠队列配置，不受 `localValueCache.enabled` 控制；它保护观测上报稳定性，而不是客户端读性能优化。

## 适用范围

原则上所有面向客户端返回、且 value 较大的可重建数据都可以进入 Pebble：

| 领域 | 适用数据 | 内存保留 |
|------|----------|----------|
| 服务发现 | 实例列表、服务发现响应、客户端发现响应 | namespace/service/revision/filter hash、实例健康与心跳热状态 |
| 配置中心 | 配置文件内容、发布版本、配置发现响应、Watch 返回快照 | namespace/group/file/release 索引、mtime、权限与订阅索引；value 直接在 Pebble |
| 服务契约 | 契约内容、版本快照 | service/protocol/version 索引；value 直接在 Pebble |
| 治理规则 | 各类 active 发布快照、按服务聚合后的下发响应 | rule type、service、active revision、优先级和轻量匹配索引 |

不适合进入同步 Pebble 热路径的数据：

- 实例心跳、临时连接、长轮询等待队列、EventHub 内部事件等高频瞬态状态。
- 小而高频、编码成本低于 Pebble 读取成本的数据。
- 需要每次请求按调用方实时修改且无法稳定 hash 的响应。

## 缓存接口

在实现上新增统一抽象，例如 `pkg/localcache`：

```go
type ValueCache interface {
    GetBytes(ctx context.Context, key []byte) ([]byte, bool, error)
    PutBytes(ctx context.Context, key []byte, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key []byte) error
    ScanPrefix(ctx context.Context, prefix []byte, fn func(key, value []byte) error) error
    BatchWrite(ctx context.Context, writes []WriteOp) error
    Close() error
}
```

默认提供 no-op / memory 实现，生产可启用 Pebble 实现。这样 cache 层可以分阶段接入，且未开启本地磁盘缓存时行为保持不变。

客户端 value cache 默认按进程只打开一个 DB，例如：

```text
./data/cache/local-value.pebble
```

不同业务域通过 key prefix 隔离，不为每类资源单独打开 Pebble 实例。

OTel event/audit 本地队列使用独立 Pebble DB，例如：

```text
./data/observability/otel-events.pebble
```

原因是 event queue 是持续 append/delete 的写入负载，和客户端 value cache 的读多写少负载不同。默认物理隔离可以降低 compaction 对客户端读延迟的影响；生命周期、容量和指标仍由同一个本地存储管理器统一管理。

## Key 设计

Key 必须携带能够证明 value 有效性的 revision、版本或过滤条件，避免把不同可见性、不同筛选条件的结果误命中。

建议前缀：

| 数据 | Key 前缀 | Value |
|------|----------|-------|
| 实例列表 | `inst/v1/{namespace}/{service}/{revision}/{filter_hash}` | protobuf marshal 后的实例列表或发现响应 |
| 配置发布 | `cfg/v1/release/{namespace}/{group}/{file}/{release_id}` | protobuf marshal 后的配置发布对象 |
| 配置发现响应 | `cfg/v1/discover/{namespace}/{group}/{file}/{client_labels_hash}/{revision}` | protobuf marshal 后的配置发现响应 |
| 服务契约 | `contract/v1/{namespace}/{service}/{protocol}/{version}` | protobuf marshal 后的契约内容 |
| 治理规则发布 | `gov/v1/{rule_type}/{namespace}/{rule_id}/{revision}` | protobuf marshal 后的治理规则发布快照 |
| 治理规则索引 | `gov-index/v1/{namespace}/{service}/{rule_type}/{rule_id}` | 指向当前 active revision 的轻量引用 |

`filter_hash` 至少要覆盖健康状态过滤、实例标签过滤、客户端可见性、权限可见性和协议差异；不能只用 namespace/service，否则会把不同调用者可见的实例列表混在一起。

OTel event/audit queue 使用单调递增 key：

| 数据 | Key 前缀 | Value |
|------|----------|-------|
| 待发送 event/audit | `otel-queue/v1/{signal}/{sequence}` | protobuf marshal 后的 OTel LogRecord 或内部 event envelope |
| 发送游标 | `otel-cursor/v1/{signal}` | 已确认发送的 sequence |

事件队列只在 Collector 确认成功后删除对应 sequence；导出失败保留并重试，容量达到上限时按配置丢弃最旧或最新记录，并上报 drop metrics。

## 读写路径

写入路径：

1. 事实来源发生变更，例如缓存增量刷新、配置发布 active 切换、治理规则发布切换或实例 revision 变化。
2. 业务 cache 组装最终可复用的 protobuf message。
3. 对超过阈值、且已开启对应 domain 的 value marshal 一次写入 Pebble，并同步更新内存中的轻量索引。
4. 删除同一资源旧 revision 的 key，避免无界增长。

读取路径：

1. 根据请求上下文、服务信息、revision 和过滤条件生成 key。
2. 命中 Pebble bytes 时，优先直接写回协议层，或只在必须修改字段时再 unmarshal。
3. 未命中时走现有内存 cache / DB 组装路径，完成 marshal 后回填 Pebble。
4. Pebble 读取或写入失败时降级为现有路径，并记录 metrics/event，不影响请求成功率。

为保证性能，Pebble 写入应尽量发生在缓存刷新、发布切换、实例 revision 变化等异步或低频路径；客户端请求路径只做 key 计算、内存索引查询和 Pebble `Get`，不能在请求内做全量扫描、compaction 或大量重建。

## 性能验收

开启 Pebble 后必须通过基准门禁，不能只凭主观判断合入。

最低验收标准：

- `enabled=false` 与当前主干路径性能一致，p95、p99、alloc/op 不得有可测退化。
- `backend=pebble` 在热页缓存稳态下，服务发现、配置发现、服务契约和治理下发的 p95 不低于当前内存路径。
- 对大 value 场景，alloc/op 和 Go heap 占用必须明显下降。
- 冷启动或 Pebble miss 可以慢于热路径，但必须有预热机制和 miss metrics。
- Pebble 异常、磁盘满、读延迟超过阈值时必须可按 domain 降级或关闭，不能影响客户端请求成功率。

需要新增基准：

| 场景 | 基准目标 |
|------|----------|
| 大服务实例列表发现 | 对比当前内存对象组装 + marshal 与 Pebble bytes 读取 |
| 配置大文件发现 | 对比当前内容读取 + response marshal 与 Pebble bytes 读取 |
| 服务契约大内容查询 | 对比当前对象缓存与 Pebble bytes |
| 治理规则大快照下发 | 对比当前规则聚合与 Pebble bytes |

## 低拷贝边界

这个设计目标是低对象分配和低重复编解码，不承诺完整内核态零拷贝。Go 的 gRPC/HTTP 栈通常仍会发生必要拷贝。

Pebble `Get` 返回的 value slice 只在 closer 关闭前有效。实现上不能把该 slice 存入长期内存；如果要跨函数持有，必须复制。如果协议层允许在 closer 生命周期内完成写出，可以减少一次复制，但必须保证生命周期清晰。

## 一致性与淘汰

- 事实来源仍由 MySQL、注册内存态和治理 cache 决定。
- value 的有效性由 revision、release_id、mtime 或 active revision 约束。
- 进程冷启动时可以选择懒加载，也可以由现有 cache 首轮刷新预热。
- 磁盘缓存达到容量上限时按 prefix 或 LRU 近似策略清理旧 revision。
- cache hit/miss、bytes size、write error、eviction、Pebble compaction 等指标需要接入系统监控。

## 分阶段实施

1. 已新增共享 Pebble 封装，并接入配置文件发布内容、服务契约和 OTel event/audit queue。
2. 建立完整 `ValueCache` 配置开关与 `enabled=false` 零退化测试。
3. 建立服务发现、配置发现、服务契约和治理下发性能基准。
4. 接入服务实例列表和发现响应，key 必须包含 revision 与调用方可见性相关 hash。
5. 接入治理规则发布快照和服务维度索引。
6. 补齐 Console 系统监控中的 Pebble cache 命中率、容量、错误、降级和 compaction 指标。

## 相关页面

- [[cache-layer]]
- [[storage]]
- [[service-discovery]]
- [[config-center]]
- [[governance-rules]]
