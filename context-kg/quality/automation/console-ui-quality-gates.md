---
title: Console UI 质量门禁与发布验收
tags: [quality, automation, console, frontend, ux]
links: [testing, adr-console-fluent-ui-design-system, configuration]
updated: 2026-08-09
sources: 8
---

# Console UI 质量门禁与发布验收

## 适用范围

本页定义 Console 共享组件、资源页面和发布产物的稳定质量契约。接口 E2E 仍按 [[testing]] 的 Go/HTTP 边界执行；这里补充源码契约、生产构建和真实浏览器验收，不把三类证据混为一种“测试通过”。

## 三层证据

| 层级 | 证明什么 | 不能证明什么 |
|---|---|---|
| Node `verify-*.mjs` | 共享组件和关键页面仍具备约定的源码结构、样式规则与真实 action 绑定 | 浏览器实际布局、网络请求结果、级联样式和运行资源 |
| ESLint + production build + Go tests | 类型、依赖、构建产物和 Console 后端回归通过 | 已部署的是本次产物，页面操作真实可达 |
| Kubernetes + 真实浏览器 | 运行镜像、Gateway、登录会话、页面布局、交互和请求形成完整闭环 | 未覆盖页面或未执行的破坏性业务操作 |

静态门禁不是浏览器 E2E。真实浏览器验收当前作为发布检查执行，不并入 `test/e2e` 的 HTTP/API 套件，也不在仓库内建设依赖截图像素比对的脆弱测试。

## 表格统一契约

所有数据表和用列头表达字段语义的编辑网格遵循同一规则：

- 表格外壳宽度自适应内容区，页面根节点不得因为宽表产生横向滚动。
- 超出内容区时只允许表格内部横向滚动，并用 `overscroll-behavior-inline: contain` 阻止滚动穿透。
- 选择列与首个业务列固定在左侧，末尾操作列固定在右侧；业务显式指定的 fixed side 优先。
- 数字或 `px` 列宽落实为像素约束，`minWidth` 必须生效；旧百分比宽度回退为共享最小宽度，不能把全部列压缩进视口。
- 表头和单元格默认左对齐；长文本单行省略，并通过 `title` 或等价 Tooltip 暴露完整可访问文本。
- 非受控分页由共享 Table 对本地数据真实切片；受控分页只触发一套外部回调，不能双重请求。
- 可点击行必须支持键盘聚焦、Enter 和空格；全选必须排除 disabled 行。
- 分页在窄屏允许表格内部换行，不能被固定高度或外壳裁切。

“首尾固定”必须通过真实滚动验证：把内部滚动容器移动到最大 `scrollLeft`，比较滚动前后首列左边界和尾列右边界保持不变。只检查 `position: sticky` 源码不构成完整证据。

## 表单与弹层契约

- 受控 Input/Textarea 的外部值从字符串恢复为 `undefined` 时，DOM value 必须同步清空；重置验收要输入实际内容、点击重置并读取 DOM，而不是只看 FormStore。
- 可重复编辑行的 React key 不能包含正在编辑的字段值，否则每次输入都会重挂载并丢失焦点。
- 新增动态编辑器时必须同步加入 `verify-fluent-input-controls.mjs`；仅有通用文字规则不足以阻止新组件
  漏检。连续输入验收必须使用逐字符键盘输入，并同时断言完整值和焦点保持，不能用原子 `fill` 替代。
- Schema 驱动的表单必须把类型约束落实为即时字段校验，错误态提供可访问说明，并在保存和发布入口统一
  阻断；类型文字、占位符或服务端返回错误不能替代前端反馈。前后端共享语义时，前端规则必须镜像服务端
  canonical 契约，并以服务端为最终权威。
- 发布型编辑器的预览必须覆盖当前未保存输入，而不能只允许选择历史发布版本；预览依赖输入变化后应清空
  旧结果。当前草稿预览和历史版本组合预览可以复用展示组件，但必须隔离状态和入口语义。
- Drawer/Dialog 的自定义宽度必须受视口约束；底部 action 始终可达，空 footer 不占位。
- 用列头表达语义的重复字段行不能在窄屏直接隐藏列头。应保留局部滚动和固定首尾列，或改为带显式字段标签的响应式卡片。
- 加载多个授权对象时必须有界分页并收敛所有请求；单项失败可以展示部分结果，但不能永久 Loading。
- 面包屑和图标操作必须使用可聚焦 Button/Link，并提供稳定的可访问名称。

## 发布级浏览器矩阵

每次影响共享组件、页面壳层、主题、路由或主要资源交互的改动，至少覆盖：

- Admin 登录后的全部可见业务域，以及详情页、主要 Tab、创建/编辑抽屉和确认/取消路径。
- 640px 窄视口、约 900px 中视口和桌面/超宽视口。
- 亮色与暗色主题，尤其检查 sticky 列、弹层、空态和代码编辑器表面。
- 查询、重置、刷新、分页、Tab 切换、详情返回和表单校验；安全的临时资源操作应在验证后清理。
- `documentElement.scrollWidth === clientWidth`，无错误 alert、无残留 progressbar、无浏览器 error。

大规模跨路由巡检本身可能形成压力负载。若 Pod 出现 OOMKilled 或重启，必须如实记录压力条件、资源上限、恢复状态和最终常态占用，不能为了“验收通过”把 restart 事实写成 0。

## 页面不可访问的快速诊断

页面打不开时先建立四个快速失败点：Console 根页面、入口主资源、一个业务深链和 8090 API。若入口拒绝连接：

1. 检查 Deployment、Pod、Service Endpoint 和 readiness。
2. 查看当前及前一容器日志，优先识别 Store、外部依赖和端口绑定失败。
3. 本地 Kubernetes 通过宿主机 Docker MySQL 时，同时检查容器运行状态和 RestartPolicy。
4. 只有 HTTP 入口、静态资源和 Pod 都正常后，才继续排查 SPA cache、动态 chunk 或浏览器运行时。

`pole-mysql` 应使用 `unless-stopped`。应用 Pod Ready 只能证明当前依赖可用，不能替代对外部 MySQL 自动恢复能力的检查。运行部署边界见 [[configuration]]。

## 证据

- `web/console/src/components/Fluent/index.tsx`
- `web/console/src/styles/fluent.less`
- `web/console/scripts/verify-fluent-table-layout.mjs`
- `web/console/scripts/verify-console-ux-closure.mjs`
- `web/console/scripts/verify-fluent-input-controls.mjs`
- `web/console/src/pages/Configuration/Template/SchemaEditor.tsx`
- `web/console/src/pages/Configuration/Template/ValueEditor.tsx`
- `web/console/src/pages/Configuration/Template/RenderPreviewPanel.tsx`
- `deploy/kubernetes/pole-control-plane.yaml`

## 相关页面

- [[testing]]
- [[adr-console-fluent-ui-design-system]]
- [[configuration]]
