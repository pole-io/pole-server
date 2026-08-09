---
title: ADR：配置模板标签与敏感 Value 加密
tags: [adr, config, template, labels, encryption, security]
links: [adr-config-template-client-rendering, config-center, auth-system]
updated: 2026-08-08
sources: 9
---

# ADR：配置模板标签与敏感 Value 加密

## 状态

Accepted，控制面存储、HTTP 管理接口与 Console 交互已实现。

## 背景

配置模板目录需要按归属、用途等元数据分类；参数 Schema 已有 `sensitive` 标记，但此前只让 Console
使用密码输入框，`namespace_template_values` 与 `namespace_template_value_release` 仍保存明文。

## 决策

### 模板标签

- 标签是全局模板草稿的目录元数据，使用唯一键和值组成的 Map，独立保存于
  `config_file_template.labels`，不写入模板说明或参数 Schema。
- 标签通过 `/config/v1/templates/labels` 独立读写，沿用配置模板的读、改权限；单模板最多 64 个标签。
- 标签不进入不可变 Template Release，也不参与运行时渲染和 revision。发布快照继续只描述会影响输出的
  内容、格式、引擎与参数 Schema。

### 敏感 Value

- `ConfigTemplateParameterSchema.sensitive=true` 表示该参数的 Value 必须加密存储，Console 文案明确为
  “加密存储”。
- 保存 Namespace Value 草稿和发布 Value Release 时，服务端使用内置 AES 插件为每个敏感值生成独立
  数据密钥，并把密文、算法与数据密钥封装进该值的 JSON；非敏感值保持原有格式。
- 管理查询、服务端参考预览和运行时快照解析在鉴权后解密，因而客户端渲染契约不需要认识密文封装。
- revision 使用规范化明文计算，避免随机密钥导致相同业务值每次保存都产生不同 revision。
- 旧版无 `encrypted` 标记的明文 Value 继续按原格式读取，实现滚动升级兼容。

## 安全边界

该实现复用现有配置中心的 AES 数据密钥模型，解决数据库字段和常规查询中的明文暴露；数据密钥仍随
密文记录持久化，不等同于外部 KMS 或 HSM 的密钥隔离。后续如引入 KEK/KMS，只需替换数据密钥封装，
无需改变模板、Value API 或 SDK 渲染契约。

## 相关页面

- [[adr-config-template-client-rendering]]
- [[config-center]]
- [[auth-system]]
