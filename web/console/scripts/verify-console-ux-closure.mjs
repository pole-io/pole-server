#!/usr/bin/env node

import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = (file) => fs.readFileSync(file, 'utf8');

const policyEditor = read('src/pages/Auth/Policy/PolicyEditor.tsx');
const policyStyle = read('src/pages/Auth/Policy/index.module.less');
const policyTable = read('src/pages/Auth/Policy/PolicyTable.tsx');
const groupTable = read('src/pages/Auth/Principal/GroupTable.tsx');
const policyModule = read('src/modules/auth/policy.ts');
const serviceMonitor = read('src/pages/Metrics/ServiceMonitor/Detail.tsx');
const serviceMonitorStyle = read('src/pages/Metrics/ServiceMonitor/index.module.less');
const governanceWorkbench = read('src/pages/Governance/Workbench/index.tsx');
const governanceStyle = read('src/pages/Governance/Workbench/index.module.less');
const namespaceStyle = read('src/pages/Namespace/index.module.less');
const sharedStyle = read('src/styles/fluent.less');
const laneStyle = read('src/pages/Governance/Router/LaneGroupEditor.module.less');
const rateLimitStyle = read('src/pages/Governance/RateLimit/RateLimitEditor.module.less');
const trafficMatchStyle = read('src/pages/Governance/shared/TrafficMatchConditionEditor.module.less');

assert.match(policyEditor, /className=\{style\.policyModeRow\}/, '策略接口步骤必须使用稳定的模式工具栏，不能依赖 24 栅格压缩标签。');
assert.match(policyEditor, /className=\{style\.policyModeSwitch\}/, '策略模式切换必须有独立不换行承载区。');
assert.match(policyStyle, /\.policyModeRow[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+auto/, '策略模式工具栏应让选项自适应、切换区保持完整。');
assert.match(policyStyle, /\.policyModeSwitch[\s\S]*white-space:\s*nowrap/, '策略模式标签禁止逐字换行。');
assert.match(policyEditor, /className=\{style\.policySelectionLayout\}/, '策略成员、资源与接口选择区必须使用统一响应式布局。');
assert.match(policyStyle, /@media \(max-width: 900px\)[\s\S]*\.treeContent[\s\S]*flex-basis:\s*132px[\s\S]*\.fluent-transfer__pane[\s\S]*min-width:\s*0\s*!important/, '中窄屏策略选择器必须压缩导航并保证双列表完整可达。');

assert.match(policyTable, /handleBatchDeletePolicies/, '策略批量删除按钮必须绑定真实批量删除流程。');
assert.match(groupTable, /handleBatchDeleteGroups/, '用户组批量删除按钮必须绑定真实批量删除流程。');
assert.match(policyModule, /createAuthPolicies\(\[/, '策略创建 thunk 必须调用真实 API。');
assert.match(policyModule, /modifyAuthPolicies\(\[/, '策略更新 thunk 必须调用真实 API。');
assert.match(policyModule, /deleteAuthPolicies\(ids\.map/, '策略删除 thunk 必须调用真实 API。');
assert.doesNotMatch(policyModule, /fulfillWithValue\(\"ok\"\)/, '策略变更不得返回伪成功。');

assert.doesNotMatch(serviceMonitor, />客户端<\/Button>/, '没有客户端方向数据时不得展示无效客户端切换按钮。');
assert.ok(!serviceMonitor.includes('icon={<RefreshIcon />} variant="outline">刷新</Button>'), '没有刷新数据源时不得展示无动作刷新按钮。');
assert.match(serviceMonitor, /服务端调用视角/, '移除无效切换后必须明确当前监控视角。');
assert.match(serviceMonitorStyle, /\.detailTabs[\s\S]*\.fui-TabList[\s\S]*overflow-x:\s*auto/, '服务监控详情的长标签栏必须支持内部横向滚动。');
assert.match(governanceWorkbench, /title=\{`\$\{serviceContext\.namespace\}\/\$\{serviceContext\.service\}`\}/, '服务治理上下文长标识必须提供完整值提示。');
assert.match(governanceStyle, /@media \(max-width: 1180px\)[\s\S]*\.serviceContextBar[\s\S]*flex-wrap:\s*wrap/, '服务治理上下文必须在中等宽度提前换行。');
assert.match(governanceStyle, /\.serviceRoleControl[\s\S]*\.fui-Radio[\s\S]*white-space:\s*nowrap/, '主调方与被调方标签不得逐字换行。');
assert.match(governanceStyle, /@media \(max-width: 1180px\)[\s\S]*\.listPanel[\s\S]*min-height:\s*420px/, '中窄屏治理表格必须保留可操作的表体高度。');
assert.doesNotMatch(namespaceStyle, /:global\(\.fluent-table-shell\)\s*\{[^}]*min-width:\s*(?:956|1046)px/, '命名空间页面不得把表格最小宽度施加到滚动容器外层。');
assert.doesNotMatch(namespaceStyle, /@media \(min-width:\s*1260px\)[\s\S]*overflow-x:\s*hidden/, '宽屏也必须保留表格内部横向滚动能力。');
assert.match(namespaceStyle, /@media \(max-width:\s*1180px\)[\s\S]*\.namespaceTableSurface[\s\S]*min-height:\s*420px/, '中窄屏表格必须保留可操作的表体高度。');
assert.match(sharedStyle, /\.fluent-tabs > \.fui-TabList[\s\S]*overflow-x:\s*auto[\s\S]*overscroll-behavior-inline:\s*contain/, '共享长页签必须在组件内部横向滚动。');
assert.match(sharedStyle, /\.fluent-tabs > \.fui-TabList > \.fui-Tab[\s\S]*white-space:\s*nowrap/, '共享页签文字不得被压缩换行。');
assert.match(laneStyle, /\.prdTable[\s\S]*overflow-x:\s*auto[\s\S]*\.prdTableRow[\s\S]*min-width:\s*696px/, '泳道编辑表必须保留完整列宽并在表内横向滚动。');
assert.match(laneStyle, /\.prdTableRow[\s\S]*> :first-child,[\s\S]*> :last-child[\s\S]*position:\s*sticky/, '泳道编辑表首尾单元必须固定。');
assert.match(rateLimitStyle, /\.interfaceGrid,[\s\S]*\.windowGrid[\s\S]*overflow-x:\s*auto/, '限流接口与窗口网格必须局部横向滚动，不能把滚动交给整个表单。');
assert.match(trafficMatchStyle, /@container \(max-width: 920px\)[\s\S]*\.conditionRow,[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+minmax\(0,\s*1fr\)/, '匹配条件在窄容器中必须切换为响应式卡片布局。');

console.log('Console UX closure contracts verified');
