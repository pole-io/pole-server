#!/usr/bin/env node

import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const adapter = read('src/components/Fluent/index.tsx');
const styles = read('src/styles/fluent.less');

assert.match(adapter, /tableLayout = 'fixed'/, 'Fluent Table 默认必须使用稳定的 fixed 布局。');
assert.doesNotMatch(adapter, /column\.width \/ totalColumnWeight/, '数字列宽不能再换算成百分比后压缩。');
assert.match(adapter, /resolveColumnMinWidth/, '表格必须统一解析业务列的最小宽度。');
assert.match(adapter, /Math\.max\(widthPixels, minWidthPixels\)/, '同时声明 width/minWidth 时必须保留两者中的较大值。');
assert.match(adapter, /if \(minWidthPixels !== undefined\) return minWidthPixels;[\s\S]*if \(widthPixels !== undefined\) return widthPixels;/, '百分比等非像素旧列宽必须回退到共享最小宽度，不能撑爆 fixed 表格。');
assert.match(adapter, /width:\s*getPixelSize\(column\.width\) \?\? resolveColumnMinWidth\(column\)/, '单元格宽度只接受数字或 px，旧百分比宽度必须自适应。');
assert.match(adapter, /minWidth:\s*resolveColumnMinWidth\(column\)/, '业务列的 width/minWidth 必须落实为 CSS 最小宽度。');
assert.match(adapter, /tableMinWidth/, '表格必须根据各列最小宽度计算内部滚动宽度。');
assert.match(adapter, /firstDataColumnIndex/, '表格必须识别首个数据列。');
assert.match(adapter, /lastDataColumnIndex/, '表格必须识别末尾数据列。');
assert.match(adapter, /resolveFixedSide/, '表格必须由共享组件统一解析首尾固定列。');
assert.match(adapter, /textAlign: column\.align \|\| 'left'/, '表头与内容必须遵循列对齐配置，并以左对齐为默认值。');
assert.match(adapter, /TableHeaderCell[\s\S]*style=\{resolveCellStyle\(column\)\}/, '表头必须应用共享列样式。');
assert.match(adapter, /TableCell[\s\S]*style=\{resolveCellStyle\(column\)\}/, '表格内容必须应用共享列样式。');
assert.match(adapter, /fluent-table-header-content/, '表头缺少稳定的溢出容器。');
assert.match(adapter, /fluent-table-cell-content--ellipsis/, '表格内容缺少共享省略容器。');
assert.match(adapter, /getCellAccessibleText/, '表格必须为可省略的长文本提取全文提示。');
assert.match(adapter, /title=\{shouldEllipsize \? cellText \|\| undefined : undefined\}/, '长文本必须提供可悬浮查看的全文提示。');
assert.match(adapter, /visibleData = pagination && !hasExternalPagination[\s\S]*data\.slice/, '没有外部数据源的分页必须在共享 Table 内真实切片。');
assert.match(adapter, /if \(pagination\.onChange\)[\s\S]*else if \(onPageChange\)/, '共享 Table 翻页只能调用一套回调，禁止重复请求。');
assert.match(adapter, /tabIndex=\{onRowClick \? 0 : undefined\}/, '可点击表格行必须可通过键盘聚焦。');
assert.match(adapter, /event\.key === 'Enter' \|\| event\.key === ' '/, '可点击表格行必须支持 Enter 和空格触发。');
assert.match(adapter, /selectableColumn\?\.checkProps/, '共享 Table 必须执行选择列 checkProps。');
assert.match(adapter, /selectableKeys[\s\S]*filter\(\(\{ disabled \}/, '全选必须排除禁用行。');
assert.match(styles, /\.fluent-table-scroll table[\s\S]*width: 100%/, '表格必须填满可用容器宽度。');
assert.match(styles, /\.fluent-table-scroll[\s\S]*overflow-x: auto/, '宽表格必须在表格内部横向滚动。');
assert.match(styles, /\.fluent-table-scroll[\s\S]*overscroll-behavior-inline: contain/, '表格横向滚动不得穿透到页面。');
assert.match(styles, /\.fluent-table-scroll th,[\s\S]*text-align: left/, '表头和内容必须共享左对齐基线。');
assert.match(styles, /\.fluent-table-cell-content--ellipsis[\s\S]*text-overflow: ellipsis[\s\S]*white-space: nowrap/, '长内容必须单行省略。');
assert.match(styles, /\.fluent-pagination[\s\S]*flex-wrap: wrap/, '窄屏分页必须在表格内部换行，不能被外壳裁切。');
assert.match(adapter, /const justifyContentMap/, 'Row 必须支持 justify 兼容映射。');
assert.match(adapter, /const alignItemsMap/, 'Row 必须支持 align 兼容映射。');
assert.match(adapter, /fluent-row/, 'Row 必须提供共享栅格类。');
assert.match(adapter, /`fluent-col--has-\$\{breakpoint\}`/, 'Col 必须支持响应式断点。');
assert.match(styles, /--fluent-grid-columns:\s*12/, 'Col 必须遵循现有页面使用的 12 列栅格。');
assert.match(styles, /@media \(min-width: 1200px\)[\s\S]*--fluent-col-xl/, 'Col 必须在宽屏落实 xl 响应式跨度。');

console.log('Fluent table scrolling, sticky columns, ellipsis and grid contracts verified');
