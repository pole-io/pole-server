---
name: wiki
version: 1.0.0
description: |
  维护 context-kg/ 知识库 wiki。支持三种操作：
  - ingest：分析代码变更并更新 wiki 页面
  - query：基于 wiki 回答问题
  - lint：检查 wiki 健康状况（孤立页面、过时内容、缺失链接）
  当用户询问代码架构、模块说明、使用方式等问题时，优先从 wiki 读取，而非重新扫描代码。
  当用户说"更新知识库"、"ingest"、"lint wiki" 时主动调用本 skill。
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Agent
---

## 前置步骤（每次运行必须先执行）

```bash
WIKI_DIR="$(git rev-parse --show-toplevel 2>/dev/null)/context-kg"
TODAY=$(date +%Y-%m-%d)
echo "WIKI_DIR: $WIKI_DIR"
echo "TODAY: $TODAY"
echo "FILES: $(ls "$WIKI_DIR"/*.md 2>/dev/null | wc -l | tr -d ' ') 个页面"
```

根据前置输出确认 wiki 目录存在。若不存在，告知用户并停止。

---

## 操作一：ingest（接入代码变更）

**触发词**：用户说"更新知识库"、"ingest"、"代码有变更"、"同步 wiki" 等。

### 步骤

**第一步：确定变更范围**

```bash
cd "$(git rev-parse --show-toplevel)"
# 查看最近变更的 Go 文件
git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.go$' | head -30 || \
git status --short | grep '\.go' | head -30
```

若用户指定了具体变更内容，优先以用户描述为准，不必完全依赖 git。

**第二步：读取 wiki 索引，定位受影响的页面**

读取 `$WIKI_DIR/index.md`，根据变更涉及的模块（pkg/service、plugin/apiserver、apis/store 等）
找出需要更新的 wiki 页面。

**第三步：读取并更新受影响页面**

对每个需要更新的页面：
1. 读取现有内容
2. 更新正文中过时的描述
3. 更新 frontmatter 中的 `updated` 字段为 `$TODAY`
4. 若有新的关联关系，更新 `links` 数组和"相关页面" section
5. 保持 [[双向链接]] 风格

如果变更引入了全新的模块或功能，创建新页面：
```
$WIKI_DIR/<新模块名>.md
```
新页面 frontmatter 模板：
```yaml
---
title: <中文标题>
tags: [<相关标签>]
links: [<关联页面列表>]
updated: <TODAY>
sources: 1
---
```
新页面底部必须有 `## 相关页面` section。

**第四步：更新 index.md**

若有新增页面，在 `$WIKI_DIR/index.md` 对应分类下追加一行：
```
- [[页面名]] — 一句话摘要 | 标签
```

**第五步：追加 log.md**

在 `$WIKI_DIR/log.md` 末尾追加：
```
## [<TODAY>] ingest | <本次变更描述>
- 涉及文件：<列出主要 Go 文件>
- 更新页面：<列出更新的 wiki 页面>
- 新增页面：<若有>
```

**第六步：汇报结果**

简要告知用户：更新了哪些页面、是否新建了页面。

---

## 操作二：query（查询知识库）

**触发词**：用户问关于代码架构、模块说明、接口定义、设计模式等问题。

### 步骤

**第一步：读取 index.md 定位相关页面**

```bash
cat "$WIKI_DIR/index.md"
```

根据问题关键词，找出最相关的 1-3 个页面。

**第二步：读取相关页面内容**

读取定位到的页面，综合内容回答问题。引用时使用 `[[页面名]]` 格式注明来源。

**第三步：判断是否值得存档**

若答案属于以下类型，主动询问用户是否存为新 wiki 页面：
- 跨多个模块的深度分析
- 对比/比较表格
- 某个功能的完整流程说明

存档时，在 `$WIKI_DIR/` 创建新页面并更新 index.md 和 log.md。

---

## 操作三：lint（健康检查）

**触发词**：用户说"lint wiki"、"检查知识库"、"wiki 健康检查" 等。

### 步骤

**第一步：收集所有页面**

```bash
ls "$WIKI_DIR"/*.md
```

**第二步：检查以下问题**

1. **孤立页面**：找出没有被任何其他页面 `[[链接]]` 引用的内容页面（schema/index/log 除外）
2. **断链**：检查 frontmatter `links` 数组中引用了不存在文件名的条目
3. **过时时间戳**：找出 `updated` 距今超过 30 天的页面（可能需要复查）
4. **缺失"相关页面"section**：检查内容页面是否都有 `## 相关页面`
5. **index.md 遗漏**：检查是否有 .md 文件未出现在 index.md 中

检查断链的简单方法：
```bash
grep -h '\[\[' "$WIKI_DIR"/*.md | grep -oP '\[\[\K[^\]]+' | sort -u
```
对比实际文件列表：
```bash
ls "$WIKI_DIR"/*.md | xargs -I{} basename {} .md | sort
```

**第三步：输出报告**

以 Markdown 格式汇报：
- ✅ 健康项
- ⚠️ 警告项（建议修复）
- ❌ 问题项（需要修复）

自动修复明确的问题（如补全缺失的 `## 相关页面`），询问用户后修复模糊的问题。

**第四步：追加 log.md**

```
## [<TODAY>] lint | wiki 健康检查
- 检查页面数：<N>
- 发现问题：<列表>
- 自动修复：<列表>
```

---

## 页面格式规范（供创建/编辑页面时参考）

```markdown
---
title: 页面中文标题
tags: [tag1, tag2]
links: [other-page-name, another-page]
updated: YYYY-MM-DD
sources: 1
---

# 页面中文标题

正文内容。提到其他页面时使用 [[页面名]] 格式。

...

## 相关页面

- [[page-a]]
- [[page-b]]
```

**命名约定**：文件名全小写、连字符分隔，与 frontmatter `links` 引用一致。

**[[链接]] 约定**：使用文件名（不含 .md），例如 `[[architecture]]`、`[[ai-features]]`。

---

## wiki 目录结构速查

```
context-kg/
├── _meta/
│   ├── schema.md           # wiki 操作规范（此文档的永久版本）
│   ├── index.md            # 所有页面的内容目录（入口，先读此文件）
│   └── log.md              # 追加式操作日志
├── overview/               # 全局视角
│   ├── overview.md         # 项目概览、技术选型、启动流程
│   ├── architecture.md     # 四层架构、插件体系、请求流程
│   ├── api-servers.md      # HTTP/gRPC/xDS/Nacos/Apollo/Eureka
│   └── configuration.md    # YAML 配置参考
├── domains/                # 业务域（每个域独立一文件）
│   ├── namespace.md        # 命名空间管理
│   ├── service-discovery.md# 服务发现、健康检查、批处理
│   ├── config-center.md    # 配置中心、灰度发布、Watch
│   ├── governance-rules.md # 路由/限流/熔断/故障探测/泳道
│   └── admin.md            # 管理后台操作
├── infra/                  # 技术基础设施
│   ├── storage.md          # 存储层接口与 MySQL 实现
│   ├── cache-layer.md      # 缓存子类型与增量更新机制
│   ├── auth-system.md      # 认证与授权体系
│   └── common-infra.md     # 日志、EventHub、BatchController
├── ai/                     # AI 原生能力
│   └── ai-features.md      # MCP Registry AI 视角
└── guides/                 # 开发指南
    ├── patterns.md         # 代码模式与约定（11 种）
    └── testing.md          # 测试体系与 Mock 用法
```

**新增页面时的归类原则：**
- 新业务功能 → `domains/`
- 新技术组件（存储/缓存/中间件）→ `infra/`
- AI 相关能力 → `ai/`
- 开发规范/工具 → `guides/`
- 架构级变化 → `overview/`
