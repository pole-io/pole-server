---
title: Wiki 操作手册（Schema）
tags: [meta, schema, wiki]
links: [index, log]
updated: 2026-06-09
sources: 0
---

# Wiki 操作手册

本文档是 `context-kg/` 知识库的操作规范，供 LLM 在执行 ingest、query、lint、restructure 操作时参考。

---

## 目录结构约定

`context-kg` 按三大知识域组织：业务知识、技术知识、质量保障知识。

```text
context-kg/
├── _meta/                         # 元文件目录
│   ├── schema.md                  # 本文件：wiki 操作手册
│   ├── index.md                   # 所有页面的内容目录（入口）
│   └── log.md                     # 追加式操作日志
├── tasks/                         # 任务过程记录
│   ├── todo.md                    # 当前计划、进度和 review
│   └── lessons.md                 # 纠正后沉淀的经验规则
├── business/                      # 业务知识域
│   ├── terminology.md             # 领域术语表（中英双语）
│   ├── domain-models.md           # 核心业务实体与实体关系
│   ├── business-rules.md          # 业务规则手册
│   ├── state-machines/            # 状态机定义，按需创建
│   ├── decisions/                 # PDR 产品决策记录，按需创建
│   ├── feature-registry/          # 功能档案
│   ├── user-insights/             # 用户洞察档案，按需创建
│   └── competitive-intel/         # 竞品情报，按需创建
├── technical/                     # 技术知识域
│   ├── arch/                      # 架构文档
│   ├── modules/                   # 模块说明
│   ├── Environment/               # 环境部署信息
│   ├── apis/                      # 接口契约
│   ├── conventions/               # 技术约定与编码规范
│   ├── adr/                       # 架构决策记录（ADR）与完整技术方案
│   ├── tech-debt.md               # 已知技术债，按需创建
│   └── tech-constraints-for-pm.md # PM 友好版技术约束，按需创建
└── quality/                       # 质量保障背景知识库
    ├── defects/                   # 缺陷知识库，按需创建
    ├── testcases/                 # 历史用例集，按需创建
    └── automation/                # 自动化背景知识
```

目录落点规则：

- `_meta/` 只存 `schema.md`、`index.md`、`log.md`，不记录业务或技术正文。
- `tasks/` 只存任务计划、进度、review 和 lessons，不记录长期业务或技术知识。
- `business/` 存产品语义、领域概念、功能档案、业务规则、产品决策、用户洞察和竞品信息。
- `technical/` 存架构、模块实现、接口契约、部署环境、技术约定、ADR、技术方案和技术债。
- `quality/` 存缺陷复盘、测试用例知识、自动化生成逻辑和测试策略。
- 没有实际内容的目录不要创建空壳；需要归档对应知识时再创建目录。
- 长期技术方案必须放入 `technical/adr/`，业务页面只保留摘要和链接。
- `context-kg/tasks/` 只用于任务计划、进度、review 和 lessons，不承载长期知识。

---

## 页面命名规则

- 文件名使用小写英文、数字和连字符，例如 `service-discovery.md`。
- ADR 文件使用 `adr-<topic>.md`，例如 `adr-governance-rule-unified-storage.md`。
- PDR 文件使用 `pdr-<topic>.md`，放入 `business/decisions/`。
- 缺陷复盘可使用 `<module>-defects.md`，放入 `quality/defects/`。
- 历史用例可使用 `<module>-testcases.md`，放入 `quality/testcases/`。

---

## YAML Frontmatter 格式

每个内容页面顶部必须包含以下 frontmatter：

```yaml
---
title: 页面标题（中文）
tags: [tag1, tag2]
links: [page1, page2]
updated: 2026-06-09
sources: 1
---
```

字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | 字符串 | 页面标题，与 H1 标题保持一致 |
| `tags` | 数组 | 内容分类标签，小写英文，可多个 |
| `links` | 数组 | 本文引用的其他页面文件名，不含 `.md` |
| `updated` | 日期 | 最后一次 ingest 或手动更新日期，格式 YYYY-MM-DD |
| `sources` | 整数 | 本次 ingest 扫描到的相关源码文件数；非源码归档可为 0 |

---

## 双向链接约定

正文引用其他页面时使用 wiki 链接语法，例如：

```markdown
详见 [[architecture]] 中的插件系统设计。
缓存层的实现参见 [[cache-layer]]。
```

规则：

- 链接使用文件名，不含 `.md` 扩展名。
- 链接内容与文件名完全一致。
- 每个页面底部必须有 `## 相关页面` section。
- `links` frontmatter 字段应与 `## 相关页面` 中的链接保持一致。
- Obsidian 按文件名全局解析，因此文件名在整个 `context-kg` 内必须唯一。

当前已定义的页面名：

- `_meta/`：`index`、`schema`、`log`
- `tasks/`：`todo`、`lessons`
- `business/`：`terminology`、`domain-models`、`business-rules`
- `business/feature-registry/`：`namespace`、`service-discovery`、`config-center`、`governance-rules`、`admin`
- `technical/arch/`：`overview`、`architecture`
- `technical/apis/`：`api-servers`
- `technical/Environment/`：`configuration`
- `technical/modules/`：`storage`、`cache-layer`、`auth-system`、`common-infra`、`ai-features`
- `technical/conventions/`：`patterns`
- `technical/adr/`：`adr-governance-rule-unified-storage`
- `quality/automation/`：`testing`

---

## 操作工作流

### ingest：接入新知识

触发时机：代码库发生重要变更、形成产品/技术决策、完成缺陷复盘或新增测试知识后执行。

步骤：

1. 判断知识类型：业务、技术或质量。
2. 找到对应目录和页面；没有页面时创建新页面。
3. 更新正文内容，避免把完整技术方案塞进业务功能页。
4. 更新 frontmatter 的 `updated` 日期和 `sources` 数量。
5. 检查是否需要新增或修改 wiki 链接。
6. 在 `## 相关页面` 中补充新链接。
7. 同步更新 `_meta/index.md`。
8. 在 `_meta/log.md` 末尾追加操作记录。

### query：查询知识库

步骤：

1. 从 `_meta/index.md` 定位相关页面。
2. 读取目标页面的 frontmatter `links` 字段。
3. 按需读取关联页面，构建上下文。
4. 优先使用页面中的 wiki 链接进行跨页面导航。

查询提示：

- 产品术语或业务实体 → [[terminology]]、[[domain-models]]
- 业务规则 → [[business-rules]]
- 功能档案 → `business/feature-registry/`
- 架构问题 → [[architecture]]
- 存储/数据库问题 → [[storage]]
- 性能/缓存问题 → [[cache-layer]]
- API 协议问题 → [[api-servers]]
- 技术方案或架构决策 → `technical/adr/`
- AI/MCP 问题 → [[ai-features]]
- 认证/权限问题 → [[auth-system]]
- 代码模式/约定 → [[patterns]]
- 配置/部署问题 → [[configuration]]
- 测试问题 → [[testing]]

### lint：健康检查

检查项：

1. frontmatter 完整性。
2. `links` 数组中的页面是否存在。
3. `_meta/index.md` 是否列出所有内容页面。
4. `links` 与 `## 相关页面` 是否一致。
5. 页面文件名是否全局唯一。
6. 是否存在长期方案误放在 `docs/` 或业务功能页的问题。
7. 是否存在空目录占位而无内容。

输出格式：

```text
lint 结果：
✓ frontmatter 完整：N/N 页面
✓ 链接有效：N/N 页面
✗ index 缺失：technical/adr/example.md
⚠ 目录落点问题：docs/design/example.md 应归档到 context-kg/technical/adr/
```

---

## 如何维护 _meta/index.md

`_meta/index.md` 是所有页面的目录，格式为：

```markdown
- [[architecture]] — 一句话摘要 | 标签列表
```

维护规则：

- 新增内容页面时，必须在 `_meta/index.md` 对应分类下添加一行。
- 删除页面时，同步删除 `_meta/index.md` 中对应行。
- 一句话摘要应概括页面核心内容，不超过 30 字。
- 标签与 frontmatter 中的 `tags` 保持一致。

---

## 如何维护 _meta/log.md

`_meta/log.md` 是追加式日志，只增不改。

格式：

```markdown
## [YYYY-MM-DD] 操作类型 | 操作描述
- 条目1
- 条目2
```

操作类型：

- `ingest`：接入新知识
- `lint`：健康检查
- `refactor`：内容局部调整
- `restructure`：目录结构重组

规则：

- 每次 ingest 或 restructure 都必须追加一条记录。
- 旧记录不得修改或删除。
- 最新记录排在文件末尾，`## 相关页面` 保持在页面最后。

## 相关页面

- [[index]]
- [[log]]
