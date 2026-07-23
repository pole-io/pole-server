---
title: ADR：服务调用鉴权采用托管身份与子规则请求匹配
tags: [adr, governance, auth, security, identity]
links: [auth-system, governance-rules, service-discovery, patterns]
updated: 2026-07-20
sources: 20
---

# ADR：服务调用鉴权采用托管身份与子规则请求匹配

## 状态

Accepted；SERVICE_TOKEN bootstrap 的 Ed25519 JWT V1 已实现，实例级强绑定留待后续增强。

## 背景

当前 `TrafficSecurityRule` 把调用鉴权表达为 API 范围、`TrafficMatchRule` 和命中后的 `ALLOW/DENY`。Console 默认创建 `HEADER / authorization / EXACT / value` 条件，用户可以直接填写 Header value。

这种能力本质上是“按请求字段匹配”，适合兼容已有 API Key 或自定义 Header，但不能作为 Pole 默认的服务间信任模型：

- Header value 由用户创建、查看和修改，无法稳定代表平台内的服务身份。
- 仅凭一个服务 ID 或服务名不能认证来源；任何拿到该值的调用方都可以伪造。
- 当前 `Service.token` 用于 SDK 与 control-plane 之间的客户端认证；它可以证明 SDK 有权代表某个服务请求 control-plane，但不是服务间调用的数据面身份标识或数据面凭证。
- 当前 Rust client 在调用方出站链路自行匹配规则并决定 `ALLOW/DENY`；恶意或未接入 SDK 的调用方可以绕过，因此它只能用于提前失败，不能作为服务信任的安全边界。
- 调用鉴权是数据面服务到服务的认证与策略执行，不应混入 Console 用户 Token、Role/Policy 等管理面授权链路；管理面鉴权边界见 [[auth-system]]。

## 决策

Console 对新建和编辑规则提供两种显式模式；兼容模式复用原有策略级请求匹配模型：

| 模式 | 定位 | 默认值 |
|---|---|---|
| `MANAGED_IDENTITY` | Pole 托管服务身份，由 control-plane 建立信任，业务将 SDK 提供的注入器和验签器显式接入网络栈 | 默认 |
| `LEGACY_REQUEST_MATCH` | 每个鉴权子规则在受保护接口之外，独立定义 `traffic_match_rule`；可用 `HEADER` 条件兼容 API Key、网关或遗留服务约定 | 可选 |

服务端仍可解析历史 `CUSTOM_HEADER` 规则级凭证，以免既有资源无法读取；Console 不再创建该模式。打开该类历史规则时，Console 以兼容请求匹配方式编辑，并在保存时写为 `LEGACY_REQUEST_MATCH`。

默认模式不再要求用户填写认证 Header 或静态 value。Console 只展示“托管身份已启用、凭证健康状态、最近轮换时间”等状态，不展示内部身份 ID 和凭证明文，也不提供修改身份 ID 的入口。

## 身份与凭证必须分离

托管模式包含两个不同对象：

- `ServiceIdentity`：每个服务唯一的内部主体标识，由 control-plane 在服务生命周期内创建并绑定 `namespace/service`；不可由用户查看、填写或修改。
- `WorkloadCredential`：某个合法 workload 用来证明自己属于该服务的短期凭证；可签发、轮换、吊销和过期，但不能改变 `ServiceIdentity`。

“身份不可见”是管理面和产品面的封装要求，不是安全边界。安全性必须来自不可伪造的凭证校验，而不是依赖 ID 保密。日志和链路内部可以记录脱敏后的主体引用用于审计，但不能记录凭证明文。

现有 `Service.id`、`Service.token` 与新模型职责不同：`Service.id` 是管理面服务资源 ID；`Service.token` 是 SDK → control-plane 的认证凭据；`ServiceIdentity` 是数据面服务主体。control-plane 可以在校验 `Service.token` 后向 SDK 下发其绑定的 `ServiceIdentity`，但不能把 token 本身放入服务调用请求，也不能把 token 解释为数据面主体。

## 默认托管模式调用链

```text
service create
  -> control-plane 创建内部 ServiceIdentity
  -> 合法 workload 使用 bootstrap identity 领取短期 WorkloadCredential
  -> 调用方 SDK 在出站请求中自动携带凭证
  -> 被调方 SDK 使用 control-plane 下发的 trust bundle 本地验签/验链
  -> 从已验证凭证得到 caller ServiceIdentity
  -> 使用 caller identity + callee + API 执行 TrafficSecurityPolicy
```

职责边界：

- control-plane 负责身份生命周期、凭证签发/轮换/吊销、trust bundle 和治理策略下发，不进入每次业务请求的同步热路径。
- 调用方 SDK 只领取本服务自己的凭证并自动注入请求；不得领取其它服务的私钥、密钥或 bearer credential。
- 被调方入站 SDK/sidecar 只接收验证材料和策略，在本地验证凭证并将可信主体写入内部调用上下文；认证和最终拒绝必须发生在被调方，调用方出站检查只能作为优化。
- `CALLER_SERVICE` 等调用方匹配条件必须读取“已认证主体”，不能读取调用方自行填写的普通 Header 或 metadata。

有 sidecar 或统一入站代理时优先采用短期 mTLS/X.509 workload certificate，实例私钥在本地生成且不离开 workload；纯 SDK 场景可采用 control-plane 签发的短期签名 token，并通过 Pole 保留 Header 携带。无论选择哪种格式，都必须具备 audience（被调服务或 Pole service mesh 域）、有效期、签发方、主体、凭证 ID 和 key version，并支持本地校验与轮换。

## Bootstrap 信任

control-plane 不能在“调用方自称自己是某服务”后直接签发凭证，否则只是把伪造入口前移。workload 首次领取凭证必须基于可验证的 bootstrap identity，例如：

- Kubernetes ServiceAccount projected token 与 namespace/service 绑定；
- 节点或 sidecar 的客户端证书；
- 云 workload identity；
- Pole service token，用于证明该 SDK 有权代表声明的 `namespace/service` 向 control-plane 领取身份数据；它只作用于 control-plane 通道，不进入业务调用。

control-plane 只向与目标 `ServiceIdentity` 绑定成功的 workload 签发该服务凭证。bootstrap 失败时不得下发身份凭证。

## WorkloadCredential V1

V1 使用短期 Ed25519 签名 bearer JWT，定位为纯 SDK 场景下的第一版数据面证明：

- `Issue/Renew` 只从 gRPC metadata 的 `authorization` 读取 service token，并由服务端 principal 推导 `ServiceIdentity`；请求体不能选择 namespace、service 或 subject。
- JWT 固定包含 `iss/sub/aud/iat/nbf/exp/jti`、协议版本、trust domain、namespace/service、identity revision 与 binding type，使用独立 `x-pole-workload-credential` Header/metadata，不能复用业务 `Authorization`。
- 当前 binding type 为 `SERVICE_TOKEN`，只证明 SDK 有权代表该服务，不证明具体 Pod、节点或实例唯一性。需要实例级保证时再接入 Kubernetes ServiceAccount、mTLS 或云 workload evidence。
- control-plane 的私钥只允许通过文件引用加载，不进入 YAML 明文、数据库或 Discover；配置要求恰好一个 `ACTIVE` key，并允许多个 `VERIFY_ONLY` key 支撑滚动轮换。
- `SERVICE_IDENTITY_BUNDLE` Discover 只返回 issuer、单调 sequence、bundle 版本、有效期、公开验证 key 与吊销材料。SDK 拒绝 sequence 回滚、同 sequence 不同版本、过期 bundle、未知 key 和非 Ed25519 key。
- Rust SDK 在 Engine 内存中领取并续期凭证，不持久化或暴露 raw credential；descriptor revision 变化时立即丢弃旧凭证并重新 Issue。
- Rust SDK 的 Discover control-plane 同时支持 `grpc` 与 `grpcs`，由部署方按环境选择；control-plane 的身份 Discover 与 Issue/Renew 不把 TLS 作为协议前置条件。由于 service token 和 bearer credential 经该链路传输，生产环境推荐启用 `grpcs`，选择明文 `grpc` 时由部署方承担链路窃听风险。
- SDK 本身不拥有业务 HTTP/gRPC client/server，因此不能宣称“零接入自动拦截”。业务必须显式挂载 HTTP Header 注入器或 tonic client interceptor，并在被调方挂载 HTTP/tonic 入站验签器。
- 入站验签成功后才生成字段私有的 `AuthenticatedCaller`。`managed_caller` 只读取该类型，不能从普通 caller、Header 或 metadata 自报值构造。

V1 bearer credential 被窃取后在过期前存在重放风险，缓解措施包括按需启用 TLS、短 TTL、audience、凭证/主体吊销和快速 key 轮换。TLS 不是功能启用的强制条件，但生产与跨不可信网络部署应优先开启。它不等价于 PoP 或 mTLS；高保证环境应升级为实例私钥不出 workload 的证明方式。

## 自定义 Header 兼容模式

“自定义 Header（兼容模式）”不是规则级凭证设置，而是每条 `TrafficSecurityPolicy` 的请求条件：

- 用户先在子规则配置受保护接口，再在同一子规则填写 Header 名、匹配类型和值；每条子规则独立保存，不能由一个规则级 Header 值影响全部接口。
- 默认条件沿用 `HEADER / authorization`，也保留原请求条件结构支持的来源和匹配类型；条件与受保护接口共同构成该子规则的命中范围。
- 兼容请求匹配只决定策略是否命中，不生成托管 `AuthenticatedCaller`，也不具备托管身份的自动轮换、workload 绑定和来源不可伪造保证。
- 规则级 `CUSTOM_HEADER` 的摘要写入和脱敏读取仅作为历史服务端契约保留，不能作为 Console 的新建交互或新的数据模型方向。

该模式只解决与既有系统的兼容，不替代托管身份。

## 契约方向

`TrafficSecurityRule.authentication` 只表达托管身份或兼容请求匹配；策略级 `TrafficSecurityPolicy` 继续承载受保护接口、请求条件和 `ALLOW/DENY` 动作：

```proto
message TrafficSecurityAuthentication {
  TrafficSecurityAuthMode mode = 1;
  ManagedIdentityAuthentication managed_identity = 2;
  // 仅为读取历史规则保留，Console 不再写入。
  CustomHeaderAuthentication custom_header = 3;
}

message TrafficSecurityRule {
  // existing fields ...
  TrafficSecurityAuthentication authentication = 14;
  repeated TrafficSecurityPolicy policies = 7;
}
```

`MANAGED_IDENTITY` 不包含用户可编辑 secret，并使用 `managed_caller`；`LEGACY_REQUEST_MATCH` 由每个 `TrafficSecurityPolicy.traffic_match_rule` 表达 Header 等请求条件。历史 `CUSTOM_HEADER` 的读写脱敏边界继续由服务端维护，但 Console 不再把它作为规则级凭证新建或提交。

兼容解析必须区分“历史规则未携带 mode”“历史 `CUSTOM_HEADER`”与“新建规则默认模式”：缺失 mode 与 Console 编辑的历史规则均按旧请求匹配语义解释；新建规则由 Console 显式写入 `MANAGED_IDENTITY` 或 `LEGACY_REQUEST_MATCH`，不能依赖 proto3 枚举零值静默改变语义。

## 客户端如何获得自身身份

`Service.token` 与 `ServiceIdentity` 是串联关系，不是同一个对象：SDK 先用 token 完成 SDK → control-plane 认证，control-plane 再根据 token 绑定的 `namespace/service` 返回该服务的数据面身份描述。

身份描述可以复用现有 Discover 长连接，新增独立资源类型：

```proto
enum DiscoverRequestType {
  // existing types ...
  SERVICE_IDENTITY = 30;
}

message ServiceIdentityDescriptor {
  string subject = 1;              // SDK 内部使用，不在 Console 展示
  string namespace = 2;
  string service = 3;
  string revision = 4;
  string credential_mode = 5;
  string trust_bundle_version = 6;
  string trust_domain = 7;
  string audience = 8;
  string credential_endpoint = 9;
}
```

领取流程：

```text
SDK 配置 namespace/service + service token
  -> 建立 Discover 流或发送 SERVICE_IDENTITY 请求
  -> control-plane 校验 token，并确认 token 绑定服务与请求服务一致
  -> 返回该服务的 ServiceIdentityDescriptor
  -> SDK 内部保存 subject，不向业务配置、Console 或用户编辑入口暴露
```

这里的“不可见”指用户和管理面不可见、不可修改，不代表协议内部永远不传输。身份 subject 本身不是 secret，安全性仍依赖数据面凭证证明；调用方不能只携带 subject 就被信任。

`SERVICE_IDENTITY` Discover 必须使用严格认证语义：

- SDK 在建立 gRPC Discover stream 时通过 metadata 携带 service token，不能只把 token 放在可变的 `DiscoverRequest.service` body。
- control-plane 从 token 校验得到的 principal 推导 `namespace/service/subject`，请求体中的服务字段只能用于一致性校验，不能决定返回哪个身份。
- 对该 Discover type 必须 fail-closed；即使普通 client discover 配置允许 anonymous，也不能匿名读取 identity descriptor。
- 当前允许 `Service.token` 覆盖 context token 的兼容路径不能用于身份下发。
- 使用独立 response message，不能复用包含 token 字段的公共 `Service` message 作为 identity carrier。
- descriptor 使用独立缓存域，key 至少包含已认证 service principal、trust domain 和 descriptor revision；不能复用只按请求中 namespace/service/revision 建立的普通响应缓存。

当前 Rust Discover 长连接直接创建 stream，尚未稳定附加 service token metadata；实现 `SERVICE_IDENTITY` 前必须先补齐 connector 的 stream 认证和重连时 token 更新。

Discover 与凭证签发继续分层：

- `SERVICE_IDENTITY` Discover：返回每个服务稳定、非秘密的身份描述，可以按 `service + revision` 缓存。
- `SERVICE_IDENTITY_BUNDLE` Discover：返回 trust bundle、issuer/key version、吊销窗口等公共验证材料，可以按 trust domain/version 缓存。
- `WorkloadCredentialService`：返回每个实例或 workload 独有的数据面凭证，不进入普通 Discover cache、failover cache 或广播响应。

```proto
service WorkloadCredentialService {
  rpc Issue(WorkloadCredentialIssueRequest)
      returns (WorkloadCredentialResponse);
  rpc Renew(WorkloadCredentialRenewRequest)
      returns (WorkloadCredentialResponse);
}
```

`Issue` 请求必须先通过 service token 认证，并校验请求服务与 token 绑定一致。若采用 mTLS，workload 本地生成私钥和 CSR；control-plane 只返回证书链。若采用短期签名 token，则响应按单个已认证 SDK/workload 定向返回并禁止共享缓存。

service token 能证明“SDK 有权代表服务”，但若还要求把凭证绑定到具体实例、Pod 或节点，`Issue` 需额外校验 instance registration、Kubernetes ServiceAccount、节点证书或云 workload evidence；不能从共享 service token 推导实例唯一性。

因此，Discover 可以提供“我是谁”的稳定身份描述和公共验证数据；独立 credential RPC 提供“我如何证明自己”的实例级凭证。两类数据不能合并成一个可缓存响应。

## 下发边界

沿用 [[service-discovery]] 与治理发布链路时，必须把四类数据分开：

1. 调用方自己的身份描述：在 service token 认证通过后通过 `SERVICE_IDENTITY` Discover 下发。
2. 调用方自己的短期凭证：通过独立 credential RPC 定向下发给已认证 SDK/workload。
3. 被调方验证材料：trust bundle、issuer/key version、吊销或过渡窗口信息，可随安全配置下发。
4. 调用鉴权规则：继续按被调服务索引和发布，不包含其它服务的秘密凭证。

不能把 caller credential 塞入当前按 callee 查询的 `TrafficSecurityRule` 发现响应，否则任一被调服务都会获得调用方的可冒用凭证。

## 实施阶段

截至 2026-07-20，前两阶段已经完成：第一阶段交付认证模式、隐藏 `ServiceIdentity`、metadata-only 身份发现、Console 默认托管模式和子规则请求匹配兼容；第二阶段交付 `WorkloadCredentialService`、Trust Bundle Discover、control-plane Ed25519 签发与双 key 轮换、Rust 领取/续期、HTTP/tonic 显式适配器、入站离线验签和可信 `AuthenticatedCaller`。

当前实现的安全闭环只覆盖 `SERVICE_TOKEN` 服务级 bootstrap。未配置身份、缺少凭证、凭证/签名/bundle 无效、unknown mode 或没有可信 caller 时均 fail-closed。实例级 PoP/mTLS、集中吊销管理接口和更多语言 SDK 仍是后续工作。

1. 在 specification 中保持托管认证与策略匹配的边界；历史规则级 secret 仅保留脱敏读取兼容。
2. control-plane 增加服务身份、签发密钥、service-token binding、短期凭证和 trust bundle 生命周期。
3. SDK 增加显式出站注入、入站验证和 `AuthenticatedCaller` 上下文；先覆盖 HTTP/gRPC。
4. Console 将托管身份设为默认；自定义 Header 兼容模式在每条子规则内配置请求匹配条件，不提供规则级共享凭证。
5. 增加轮换、吊销、审计和双 key 过渡窗口，验证 control-plane 暂时不可用时已签发凭证仍可在有效期内工作。

## 验收标准

- 新建服务自动拥有唯一内部身份，管理 API 和 Console 均不能读取或修改该 ID。
- 未通过 bootstrap 绑定的 workload 不能领取任何服务凭证。
- 调用方不能通过伪造服务名、服务 ID、普通 Header 或 metadata 冒充其它服务。
- 不接入或篡改调用方 SDK 也不能绕过鉴权，最终认证和拒绝在被调方入站链路执行。
- control-plane 不在业务调用热路径，被调服务可使用缓存的 trust bundle 离线验证。
- 凭证轮换期间新旧验证 key 有受控重叠窗口，吊销和过期行为有测试覆盖。
- 自定义 Header 兼容模式必须在子规则中独立保存和匹配，且不会与托管身份混用；Console 不得出现规则级共享 Header 凭证输入。
- 认证失败与策略拒绝可观测但不泄露凭证、内部身份 ID 或规则敏感值。

## 证据

- `../specification/api/v1/security/traffic_security.proto`
- `../specification/api/v1/service_manage/service.proto`
- `console/web/src/pages/Governance/Security/trafficSecurityEditorUtils.ts`
- `console/web/src/services/traffic_governance.ts`
- `pkg/goverrule/client_v1.go`
- `pkg/service/service.go`
- `../pole-client-rust/src/traffic/policy/default.rs`
- `../pole-client-rust/src/traffic/router/default.rs`
- `../specification/api/v1/service_manage/grpcapi.proto`
- `../specification/api/v1/security/workload_identity.proto`
- `plugin/apiserver/grpcserver/discover/v1/client_access.go`
- `plugin/apiserver/grpcserver/discover/v1/workload_credential_access.go`
- `plugin/apiserver/grpcserver/discover/server.go`
- `pkg/workloadcredential/server.go`
- `pkg/workloadcredential/keyring.go`
- `../pole-client-rust/src/plugins/connector/grpc/connector.rs`
- `../pole-client-rust/src/core/identity.rs`
- `../pole-client-rust/src/identity/mod.rs`

## 相关页面

- [[auth-system]]
- [[governance-rules]]
- [[service-discovery]]
- [[patterns]]
