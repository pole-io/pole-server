---
title: Wiki 操作手册（Schema）
tags: [meta, schema, wiki]
links: [index]
updated: 2026-05-14
sources: 0
---

# Wiki 操作手册

本文档是 `context-kg/` 知识库的操作规范，供 LLM 在执行 ingest、query、lint 操作时参考。

---

## 目录结构约定

```
context-kg/
├── _meta/                 # 元文件目录
│   ├── schema.md          # 本文件：wiki 操作手册
│   ├── index.md           # 所有页面的内容目录（入口）
│   └── log.md             # 追加式操作日志
├── overview/              # 全局视角
│   ├── overview.md        # 项目概览
│   ├── architecture.md    # 系统架构
│   ├── api-servers.md     # API 服务端协议实现
│   └── configuration.md   # 配置参考
├── domains/               # 业务域
│   ├── namespace.md       # 命名空间
│   ├── service-discovery.md # 服务发现
│   ├── config-center.md   # 配置中心
│   ├── governance-rules.md # 治理规则
│   └── admin.md           # 管理后台
├── infra/                 # 技术基础设施
│   ├── storage.md         # 存储层
│   ├── cache-layer.md     # 缓存层
│   ├── auth-system.md     # 认证与访问控制
│   └── common-infra.md    # 公共基础设施包
├── ai/                    # AI 原生能力
│   └── ai-features.md     # MCP Registry 详细设计
└── guides/                # 开发指南
    ├── patterns.md        # 关键模式与约定
    └── testing.md         # 测试结构与规范
```

**规则：**
- `_meta/` 下的 `schema.md`、`index.md`、`log.md` 是元文件，不记录业务知识
- 所有内容页面必须有 YAML frontmatter
- 新增页面必须同时更新 `_meta/index.md` 和 `_meta/log.md`

---

## YAML Frontmatter 格式

每个内容页面顶部必须包含以下 frontmatter：

```yaml
---
title: 页面标题（中文）
tags: [tag1, tag2]          # 小写，用于分类和过滤
links: [page1, page2]       # 本文引用的其他页面（不含 .md 后缀）
updated: 2026-05-14         # 最后更新日期，格式 YYYY-MM-DD
sources: 1                  # 本次 ingest 扫描到的源码文件数（0 表示元文件）
---
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | 字符串 | 页面标题，与 H1 标题保持一致 |
| `tags` | 数组 | 内容分类标签，小写英文，可多个 |
| `links` | 数组 | 本文中引用的其他页面文件名（不含 .md） |
| `updated` | 日期 | 最后一次 ingest 或手动更新的日期 |
| `sources` | 整数 | 本次 ingest 时扫描到的相关源码文件数 |

---

## 双向链接约定

在正文中引用其他页面时，使用 `[[页面名]]` 格式：

```markdown
详见 [[architecture]] 中的插件系统设计。
缓存层的实现参见 [[cache-layer]]。
```

**规则：**
- 链接使用文件名，**不含** `.md` 扩展名
- 链接内容与文件名完全一致（例如 `[[ai-features]]`，不是 `[[AI特性]]`）
- 每个页面底部必须有 `## 相关页面` section，列出所有 `[[链接]]`
- `links` frontmatter 字段应与 `## 相关页面` 中的链接保持一致
- Obsidian 按文件名全局解析，**无需**在链接中指定子目录路径

**已定义的页面名：**

overview 目录：
- `overview`、`architecture`、`api-servers`、`configuration`

domains 目录：
- `namespace`、`service-discovery`、`config-center`、`governance-rules`、`admin`

infra 目录：
- `storage`、`cache-layer`、`auth-system`、`common-infra`

ai 目录：
- `ai-features`

guides 目录：
- `patterns`、`testing`

_meta 目录：
- `index`、`schema`、`log`

---

## 三类操作工作流

### ingest — 接入新源码变更

触发时机：代码库发生重要变更（新功能、重构、架构调整）后执行。

**步骤：**
1. 扫描变更的 Go 文件，识别涉及的业务域
2. 找到对应的知识页面（或创建新页面）
3. 更新正文内容，保持结构不变，只补充/修正变更部分
4. 更新 frontmatter 的 `updated` 日期和 `sources` 数量
5. 检查是否需要新增或修改 `[[链接]]`
6. 在 `## 相关页面` 中补充新链接
7. 如新建页面，同步更新 `_meta/index.md`
8. 在 `_meta/log.md` 末尾追加操作记录

**日志格式：**
```
## [YYYY-MM-DD] ingest | 变更描述
- 扫描 N 个 Go 文件
- 更新页面：page1, page2
- 新增页面：page3（如有）
- 变更摘要
```

---

### query — 查询知识库

触发时机：需要回答关于代码库的问题时。

**步骤：**
1. 从 `_meta/index.md` 定位相关页面（按分类浏览）
2. 读取目标页面的 frontmatter `links` 字段，获取关联页面列表
3. 按需读取关联页面，构建完整上下文
4. 优先使用页面中的 `[[链接]]` 进行跨页面导航

**查询提示：**
- 架构问题 → [[architecture]]
- 存储/数据库问题 → [[storage]]
- 性能/缓存问题 → [[cache-layer]]
- AI/MCP 问题 → [[ai-features]]
- 认证/权限问题 → [[auth-system]]
- 代码模式/约定 → [[patterns]]
- 配置/部署问题 → [[configuration]]
- 测试问题 → [[testing]]
- 服务发现问题 → [[service-discovery]]
- 配置中心问题 → [[config-center]]
- 治理规则问题 → [[governance-rules]]

---

### lint — 健康检查

触发时机：定期检查 wiki 质量，或在 ingest 后验证一致性。

**检查项：**
1. **frontmatter 完整性** — 所有内容页面是否都有 title/tags/links/updated/sources
2. **链接有效性** — `links` 数组中的页面是否都存在
3. **index 一致性** — `_meta/index.md` 中是否列出了所有页面
4. **双向链接对称性** — 若 A 的 `links` 包含 B，建议 B 的 `## 相关页面` 也引用 A
5. **更新日期陈旧** — `updated` 日期超过 30 天且源码有变更时需提示
6. **孤立页面** — 无任何页面链接到的内容页面

**输出格式：**
```
lint 结果：
✓ frontmatter 完整：17/17 页面
✗ 失效链接：testing.md → links: [nonexistent]
⚠ 陈旧页面：storage.md（上次更新 2026-04-01，超过 30 天）
✓ index 覆盖：17/17 页面
```

---

## 如何维护 _meta/index.md

`_meta/index.md` 是所有页面的目录，格式为：

```markdown
- [[页面名]] — 一句话摘要 | 标签列表
```

**维护规则：**
- 新增内容页面时，必须在 `_meta/index.md` 对应分类下添加一行
- 删除页面时，同步删除 `_meta/index.md` 中对应行
- 一句话摘要应概括页面的核心内容，不超过 30 字
- 标签与 frontmatter 中的 `tags` 保持一致

---

## 如何维护 _meta/log.md

`_meta/log.md` 是追加式日志，只增不改。

**格式：**
```markdown
## [YYYY-MM-DD] 操作类型 | 操作描述
- 条目1
- 条目2
```

操作类型：`ingest`（接入源码）、`lint`（健康检查）、`refactor`（结构调整）、`restructure`（目录重组）

**规则：**
- 每次 ingest 都必须追加一条记录
- 旧记录不得修改或删除
- 最新记录排在文件末尾

## 相关页面

- [[index]]
- [[log]]
