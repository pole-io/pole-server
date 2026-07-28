---
title: 业务规则手册
tags: [business, rules]
links: [terminology, domain-models, namespace, service-discovery, config-center, governance-rules, auth-system, adr-governance-rule-unified-storage, adr-logical-service-environment-binding, adr-system-namespace-kind]
updated: 2026-07-28
sources: 3
---

# 业务规则手册

## 跨域规则

- `BUSINESS` Namespace 表示业务运行环境；`SYSTEM` Namespace 表示 Pole 内部管理面空间。
  系统空间不参与业务跨环境聚合，完整边界见 [[adr-system-namespace-kind]]。
- 逻辑服务使用控制面稳定 ID 作为跨环境身份；环境服务继续以 `namespace + runtimeServiceName`
  标识，只有经过管理员显式关联才属于同一个逻辑服务。
- 配置分组以名称、配置文件以“分组名 + 文件名”形成跨环境逻辑标识，各命名空间分别保存独立内容、版本和发布状态。
- 服务、实例、配置、治理规则都采用软删除语义，外部查询默认只返回有效资源。
- 写操作直接落库，读路径优先通过 cache 获取内存视图。
- 发布类能力必须形成快照，避免后续草稿修改影响已发布版本。

## 服务发现规则

- 环境服务是发现与治理的运行时对象，以“命名空间 + 运行时服务名”定位；逻辑服务只负责控制面
  跨环境聚合，不进入 SDK 或数据面寻址。
- 实例健康状态、隔离状态和健康检查状态共同决定实例是否参与发现。
- 空推保护用于避免服务实例全量消失时触发级联故障。

## 配置中心规则

- 配置文件编辑态和发布态分离。
- 灰度发布通过标签或条件限制生效范围。
- Watch 通过客户端当前版本判断是否需要返回变更。

## 治理规则规则

- 治理规则支持版本控制、灰度发布和权限控制。
- 路由、限流、熔断、故障探测、无损上下线、泳道组都属于治理规则。
- 治理规则统一存储决策见 [[adr-governance-rule-unified-storage]]。

## 访问控制规则

- Pole 支持创建、修改和删除自定义角色；自定义角色可维护名称、描述、标签、用户/用户组成员及资源/API 权限。
- Pole 同时固定提供 `admin`、`resource-reader`、`resource-writer` 三个内置系统角色，不允许删除或修改其角色定义与资源/API 权限。
- `admin` 可管理控制面、认证授权和全部业务资源；`resource-reader` 只获得业务资源读取能力；`resource-writer` 获得业务资源读写能力，但不因此获得用户、用户组、角色、策略、系统配置或运维管理能力。
- 对三个内置系统角色，管理员只能调整角色与用户、用户组的绑定；用户组成员通过用户组继承其系统角色。
- 三个角色和配套策略使用稳定 ID 幂等补齐；服务端 API 必须执行不可变约束，前端隐藏按钮不能作为安全边界。实现见 [[auth-system]]。

## 相关页面

- [[terminology]]
- [[domain-models]]
- [[namespace]]
- [[service-discovery]]
- [[config-center]]
- [[governance-rules]]
- [[auth-system]]
- [[adr-governance-rule-unified-storage]]
- [[adr-logical-service-environment-binding]]
- [[adr-system-namespace-kind]]
