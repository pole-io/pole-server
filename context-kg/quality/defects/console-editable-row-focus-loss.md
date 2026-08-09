---
title: Console 可编辑列表输入失焦复发
tags: [quality, defect, console, frontend, react, input]
links: [console-ui-quality-gates, adr-config-template-client-rendering]
updated: 2026-08-09
sources: 3
---

# Console 可编辑列表输入失焦复发

## 现象

配置模板参数 Schema 编辑态中，参数名无法连续输入。真实浏览器对输入框逐字符键入
`database.port`，最终只保留首字符 `d`，焦点随即离开原输入框。使用 `fill` 一次性赋值无法暴露该问题，
必须使用真实逐字符键盘输入验证。

这是同类缺陷的第二次出现。此前服务标签编辑器已经发生过“输入 `labelkey` 只保留 `l`”的问题。

## 根因

`SchemaEditor` 曾使用以下 React 行 key：

```text
${item.name}-${index}
```

`item.name` 正是正在编辑的受控字段。首字符写回父状态后 key 发生变化，React 将原行卸载并创建新行，
导致原输入框失去焦点，后续字符没有继续进入该控件。父组件数组更新和共享 Input 均不是根因。

修复后参数行使用编辑期间不变的 `schema-row-${index}`。当前列表只支持尾部新增和按行删除，不支持拖拽
排序；如果以后加入排序，应升级为创建时生成且贯穿草稿生命周期的稳定行 ID，不能回退到可编辑业务字段。

## 为什么已有规则仍然复发

知识库和 `verify-fluent-input-controls.mjs` 已经约束 LabelInput、治理 Header、接口及泳道输入行使用稳定 key，
但该门禁采用组件白名单，新增 `SchemaEditor` 时没有同步注册，因此规则存在而覆盖面缺失。

以后新增或重构任何动态编辑行时，完成条件必须同时包括：

1. React key 不包含 `name`、`key`、`value`、协议、路径等可编辑字段。
2. 将新组件加入连续输入专项门禁，至少包含禁止模式和期望稳定 key 两个断言。
3. 真实浏览器使用逐字符输入，不使用原子 `fill` 代替；断言完整字符串与焦点仍停留在同一输入框。
4. 新增、删除、切换类型后再次执行连续输入，避免只覆盖首行静态场景。

## 防回退证据

- `web/console/src/pages/Configuration/Template/SchemaEditor.tsx`
- `web/console/scripts/verify-fluent-input-controls.mjs`
- `web/console/scripts/verify-config-template-console.mjs`

## 相关页面

- [[console-ui-quality-gates]]
- [[adr-config-template-client-rendering]]
