import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const files = {
  fluent: 'src/components/Fluent/index.tsx',
  policyDetailView: 'src/pages/Auth/Policy/PolicyDetailView.tsx',
  policyEditor: 'src/pages/Auth/Policy/PolicyEditor.tsx',
  principalPolicyTable: 'src/pages/Auth/Principal/PrincipalPolicyTable.tsx',
  policyStyle: 'src/pages/Auth/Policy/index.module.less',
};

const sources = Object.fromEntries(
  Object.entries(files).map(([key, rel]) => {
    const file = path.join(root, rel);
    if (!fs.existsSync(file)) {
      throw new Error(`missing file: ${file}`);
    }
    return [key, fs.readFileSync(file, 'utf8')];
  }),
);

const detail = sources.policyDetailView;
const fluent = sources.fluent;
const editor = sources.policyEditor;
const principalPolicyTable = sources.principalPolicyTable;
const style = sources.policyStyle;

for (const text of [
  '策略概要',
  '成员信息',
  '资源信息',
  '资源标签',
  '可访问接口',
  '生效路径',
  '复制成员',
  '复制清单',
  '全部资源',
  '仅看新增继承',
]) {
  assert.match(detail, new RegExp(text), `策略详情必须包含「${text}」`);
}

for (const anchor of [
  'copyMember',
  'copyResourceId',
  'copyResourceRows',
  'resourceTypeSearch',
  'resourceFilter',
  'selectedResourceRowKeys',
  'policySummary',
  'policyResourceShell',
  'policyResourceNav',
  'policyResourceTable',
]) {
  assert.match(detail, new RegExp(anchor), `策略详情必须保留交互锚点 ${anchor}`);
}

assert.match(detail, /<Tabs[\s\S]*defaultValue="principal"/, '默认必须打开成员信息页签');
assert.match(detail, /rowKey="rowKey"/, '资源表格需要稳定 rowKey，避免不同资源类型 ID 冲突');
assert.match(detail, /Table\s+data=\{filteredResources\}/, '资源表格必须使用筛选后的资源列表');
assert.match(detail, /selectedRowKeys=\{viewState\.selectedResourceRowKeys\}/, '资源表格必须支持当前页签内行选择');
assert.match(detail, /onSelectChange=\{\(value\) => setViewState/, '资源行选择必须更新当前选择数量');

const summaryStart = detail.indexOf('const renderSummary = () => (');
const principalsStart = detail.indexOf('const renderPrincipals = () => (');
assert.notEqual(summaryStart, -1, '必须保留策略概要渲染函数');
assert.notEqual(principalsStart, -1, '必须保留成员信息渲染函数');
const summaryBlock = detail.slice(summaryStart, principalsStart);

for (const forbiddenSummary of [
  '策略 ID',
  '来源',
  '成员',
  '可访问接口',
  '复制 ID',
  'copyPolicyId',
  'policySummaryCopy',
]) {
  assert.doesNotMatch(summaryBlock, new RegExp(forbiddenSummary), `策略概要区不能展示 ${forbiddenSummary}`);
}

for (const forbidden of [
  'Descriptions',
  '<Tree',
  '授权概览',
  '资源类型 13 类',
  'drawer-aside',
  'drawer-foot',
  'edit-only',
  'view-only',
  'addResource',
  '>编辑<',
  '>关闭<',
]) {
  assert.doesNotMatch(detail, new RegExp(forbidden), `策略详情查看态不能包含 ${forbidden}`);
}

assert.match(editor, /const policyViewMode = op === 'view' && !editable;/,
  '策略抽屉需要显式区分查看态');
assert.match(editor, /size=\{policyViewMode \? 'min\(980px, 94vw\)' : '60%'\}/,
  '策略详情抽屉查看态宽度必须是 min(980px, 94vw)');
assert.match(editor, /footer=\{policyViewMode \? false : drawerFooter\}/,
  '策略详情查看态不能展示底部固定操作栏');
assert.match(editor, /className=\{policyViewMode \? style\.policyDetailDrawer : undefined\}/,
  '策略详情查看态 Drawer 必须使用专用内部滚动样式');
assert.doesNotMatch(editor, /op === 'view' && !editable \? \(\s*<Space>/,
  '策略详情查看态不能再生成编辑/关闭 footer');

assert.match(principalPolicyTable, /size="min\(980px, 94vw\)"/,
  '主体关联策略详情抽屉也应使用相同查看态宽度');
assert.match(principalPolicyTable, /footer=\{false\}/,
  '主体关联策略详情抽屉不能展示关闭 footer');
assert.match(detail, /<Tabs\s+className=\{style\.policyDetailTabs\}\s+defaultValue="principal">/,
  '策略详情 Tabs 内容区必须成为内部滚动容器');
assert.match(detail, /colKey: 'group', title: '接口分组', width: 30[\s\S]*colKey: 'scope', title: '访问范围', width: 50[\s\S]*colKey: 'status',[\s\S]*width: 20/,
  '可访问接口表格必须为所有列声明比例宽度，避免单个状态列占满表格');
assert.match(detail, /<Loading className=\{style\.policyDetailLoading\}/,
  '策略详情 Loading 必须进入自身的高度链');
assert.match(fluent, /className=\{`fluent-loading \$\{className \|\| ''\}`\}/,
  '共享 Loading 必须透传 className，供内容页建立高度链');

for (const className of [
  'policyDetailDrawer',
  'policyDetailLoading',
  'policyDetailView',
  'policyDetailTabs',
  'policySummary',
  'policySummaryTop',
  'policyMetaGrid',
  'policyResourceShell',
  'policyResourceNav',
  'policyResourceMain',
  'policyResourceBrief',
]) {
  assert.match(style, new RegExp(`\\.${className}\\b`), `样式必须包含 ${className}`);
}

assert.doesNotMatch(style, /\.policySummaryCopy\b/, '策略概要区不再需要复制 ID 按钮样式');
assert.match(style, /\.policyDetailDrawer[\s\S]*:global\(\.fui-DrawerBody\)[\s\S]*overflow: hidden/,
  '策略详情抽屉 body 必须隐藏外层滚动，把滚动交给内部内容区');
assert.match(style, /\.policyDetailDrawer[\s\S]*:global\(\.fui-DrawerBody\)[\s\S]*overscroll-behavior: contain/,
  '策略详情抽屉 body 必须阻止滚动链传到页面');
assert.match(style, /\.policyDetailDrawer[\s\S]*:global\(\.fluent-loading\)[\s\S]*flex: 1 1 auto/,
  '策略详情抽屉必须把 Loading 中间层纳入 flex 高度链');
assert.match(style, /\.policyDetailTabs[\s\S]*:global\(\.fluent-tab-content\)[\s\S]*overflow: auto/,
  '策略详情 Tabs 内容区必须独立滚动');
assert.match(style, /\.policyDetailTabs[\s\S]*:global\(\.fluent-tab-content\)[\s\S]*overscroll-behavior: contain/,
  '策略详情 Tabs 内容区必须阻止滚动链传到页面');
assert.match(style, /\.policyDetailTabs[\s\S]*:global\(\.fluent-tab-content\) > div[\s\S]*display: flex/,
  '策略详情活动页签必须成为 flex 容器，资源 pane 才能获得受限高度');
assert.match(style, /\.policyDetailTabs[\s\S]*:global\(\.fluent-tab-content\) > div[\s\S]*width: 100%[\s\S]*flex-direction: column/,
  '策略详情活动页签必须满宽纵向排列，成员信息不能按内容宽度收缩');
assert.match(style, /\.policyResourceShell[\s\S]*height: 100%/,
  '资源信息页签必须接入 Tabs 内容区高度，避免整个页签外层滚动');
assert.match(style, /\.policyResourceShell[\s\S]*min-height: 0/,
  '资源信息页签必须建立 min-height: 0 高度链');
assert.match(style, /\.policyResourceNav[\s\S]*overflow: hidden/,
  '资源类别左栏外层必须隐藏溢出，把滚动交给类别列表');
assert.match(style, /\.resourceTypeList[\s\S]*flex: 1 1 auto[\s\S]*overflow: auto/,
  '资源类别列表必须作为左侧独立滚动容器');
assert.match(style, /\.resourceTypeList[\s\S]*overscroll-behavior: contain/,
  '资源类别列表必须阻止滚动链传到资源页签或页面');
assert.match(style, /\.policyResourceMain[\s\S]*overflow: auto/,
  '资源详情右栏必须作为独立滚动容器');
assert.match(style, /\.policyResourceMain[\s\S]*overscroll-behavior: contain/,
  '资源详情右栏必须阻止滚动链传到资源页签或页面');

console.log('auth policy detail view verification passed');
