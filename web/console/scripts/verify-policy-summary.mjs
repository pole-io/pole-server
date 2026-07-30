import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = (file) => fs.readFileSync(file, 'utf8');

const policyPage = read('src/pages/Auth/Policy/index.tsx');
const policyTable = read('src/pages/Auth/Policy/PolicyTable.tsx');
const policyStyle = read('src/pages/Auth/Policy/index.module.less');

assert.match(
  policyPage,
  /Promise\.allSettled\(\[[\s\S]*default: 'false'[\s\S]*default: 'true'/,
  '访问策略总览必须并行读取自定义与默认策略完整总数。',
);
assert.match(
  policyPage,
  /aria-label="访问策略统计"[\s\S]*自定义策略数[\s\S]*默认策略数/,
  '访问策略页顶部必须展示自定义策略数和默认策略数。',
);
assert.match(
  policyPage,
  /type="custom"[\s\S]*onTotalChange=\{updateCustomTotal\}[\s\S]*onTotalsRefresh=\{loadTotals\}[\s\S]*type="default"[\s\S]*onTotalChange=\{updateDefaultTotal\}[\s\S]*onTotalsRefresh=\{loadTotals\}/,
  '两类策略列表必须同步回写并能触发完整总数刷新。',
);
assert.match(
  policyTable,
  /if \(!query\) props\.onTotalChange\?\.\(response\.totalCount\)/,
  '策略列表只能用未筛选总数更新顶部总览。',
);
assert.match(
  policyTable,
  /props\.onTotalsRefresh\?\.\(\)/,
  '自定义策略变更后必须重新加载两类总数。',
);
assert.match(
  policyStyle,
  /\.policyMetricRail[\s\S]*grid-template-columns: repeat\(2, minmax\(0, 1fr\)\)[\s\S]*\.policyMetricItem/,
  '访问策略统计必须使用两列统一统计卡片。',
);
assert.match(
  policyStyle,
  /@media \(max-width: 760px\)[\s\S]*\.policyMetricRail[\s\S]*grid-template-columns: 1fr/,
  '窄屏下访问策略统计必须退化为单列。',
);
assert.equal(
  (policyStyle.match(/\.policyMetricRail/g) || []).length,
  2,
  '顶部统计样式只能包含基础规则和一个响应式规则，不能复用详情页类名。',
);

console.log('访问策略顶部统计总览契约检查通过。');
