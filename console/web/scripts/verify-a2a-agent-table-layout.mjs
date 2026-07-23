import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const source = read('src/pages/AI/A2A/index.tsx');
const styles = read('src/pages/AI/A2A/index.module.less');
const fluentTable = read('src/components/Fluent/index.tsx');
const fluentStyles = read('src/styles/fluent.less');

assert.match(source, /className=\{style\.agentTable\}/, 'A2A Agent 表格必须使用专用 agentTable 样式边界。');
assert.match(source, /tableLayout="fixed"/, 'A2A Agent 表格必须使用 fixed layout，避免内容撑乱列宽。');
assert.doesNotMatch(source, /tableLayout="auto"/, 'A2A Agent 表格不能使用 auto layout。');
assert.match(source, /title: 'A2A Agent'[\s\S]*fixed: 'left'[\s\S]*width: 340/, 'A2A Agent 身份列必须固定在首列。');
assert.match(source, /title: 'A2A Agent'[\s\S]*width: 340/, '身份列应保留稳定宽度，避免名称和标签挤压。');
assert.match(source, /title: '能力'[\s\S]*width: 250/, '能力列必须有足够宽度承载 Streaming、Push、Extended 标签。');
assert.match(source, /title: '技能数'[\s\S]*width: 88/, '技能数列应保持紧凑宽度。');
assert.match(source, /title: '操作'[\s\S]*fixed: 'right'[\s\S]*width: 132/, '操作列必须固定在尾列。');
assert.match(source, /className=\{style\.skillCountLink\}/, '技能数必须使用单行链接样式，避免数字和单位拆行。');
assert.match(source, /className=\{style\.actionCell\}/, '操作列必须使用统一 actionCell 控制图标间距和右对齐。');
assert.match(source, /Extended<\/Tag>/, '列表能力标签应使用短标签 Extended，避免 Extended Card 撑高行。');

assert.match(styles, /\.filterControls\s*\{[\s\S]*justify-content: end/, 'A2A 筛选区应右对齐并使用确定性网格。');
assert.match(styles, /\.tableSurface\s*\{[\s\S]*\.fluent-table-scroll table[\s\S]*min-width: 1544px/, '表格必须设置最小宽度，保证常见桌面下列宽稳定。');
assert.match(styles, /\.tableSurface\s*\{[\s\S]*\.fluent-table-scroll th[\s\S]*white-space: nowrap/, '表头必须禁止换行。');
assert.match(styles, /\.tableSurface\s*\{[\s\S]*\.fluent-table-scroll td[\s\S]*vertical-align: top/, '表格单元格必须顶部对齐，避免多行摘要上下漂移。');
assert.match(styles, /\.fluent-table-scroll td:last-child[\s\S]*vertical-align: middle/, '操作列必须单独垂直居中。');
assert.match(styles, /\.agentNameLink\s*\{[\s\S]*text-overflow: ellipsis[\s\S]*white-space: nowrap/, 'Agent 名称必须单行省略。');
assert.match(styles, /\.capabilityTagList\s*\{[\s\S]*flex-wrap: nowrap/, '能力标签必须保持单行紧凑排列。');
assert.match(styles, /\.sourceCell\s*\{[\s\S]*flex-wrap: nowrap/, '来源标签必须保持单行扫描。');
assert.match(styles, /\.skillCountLink\s*\{[\s\S]*white-space: nowrap/, '技能数链接必须禁止换行。');
assert.doesNotMatch(styles, /\.t-table/, 'A2A 页面样式不能再残留旧 TDesign 表格选择器。');

assert.match(fluentTable, /resolveFixedClassName[\s\S]*fluent-table-fixed-left[\s\S]*fluent-table-fixed-right/, 'Fluent Table 必须把 column.fixed 映射为固定列 class。');
assert.match(fluentTable, /fixedSide === 'left'[\s\S]*position: 'sticky'[\s\S]*left: leftOffsets\[columnIndex\]/, 'Fluent Table 必须把 fixed left 映射为支持多固定列偏移的 sticky left。');
assert.match(fluentTable, /fixedSide === 'right'[\s\S]*position: 'sticky'[\s\S]*right: rightOffsets\[columnIndex\]/, 'Fluent Table 必须把 fixed right 映射为支持多固定列偏移的 sticky right。');
assert.match(fluentTable, /TableHeaderCell[\s\S]*className=\{resolveFixedClassName\(column\)\}[\s\S]*style=\{resolveCellStyle\(column\)\}/, '表头单元格必须应用 fixed class 和 sticky style。');
assert.match(fluentTable, /TableCell[\s\S]*className=\{resolveFixedClassName\(column\)\}[\s\S]*style=\{resolveCellStyle\(column\)\}/, '内容单元格必须应用 fixed class 和 sticky style。');
assert.match(fluentStyles, /\.fluent-table-scroll \.fluent-table-fixed\s*\{[\s\S]*background: var\(--app-surface\)[\s\S]*z-index: 2/, '固定内容列必须有背景和层级，避免滚动内容透出。');
assert.match(fluentStyles, /\.fluent-table-scroll th\.fluent-table-fixed\s*\{[\s\S]*background: var\(--app-surface-subtle\)[\s\S]*z-index: 4/, '固定表头列必须高于内容列。');
assert.match(fluentStyles, /\.fluent-table-scroll \.fluent-table-fixed-left\s*\{[\s\S]*left: 0/, '固定首列必须声明 left: 0。');
assert.match(fluentStyles, /\.fluent-table-scroll \.fluent-table-fixed-right\s*\{[\s\S]*right: 0/, '固定尾列必须声明 right: 0。');

console.log('a2a agent table layout checks passed');
