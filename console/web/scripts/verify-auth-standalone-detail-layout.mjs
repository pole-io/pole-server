import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const policyPage = read('src/pages/Auth/Policy/PolicyDetail.tsx');
const policyView = read('src/pages/Auth/Policy/PolicyDetailView.tsx');
const policyStyle = read('src/pages/Auth/Policy/index.module.less');
const userDetail = read('src/pages/Auth/Principal/UserDetail.tsx');
const principalStyle = read('src/pages/Auth/Principal/index.module.less');

assert.match(policyPage, /import PolicyDetailView from '\.\/PolicyDetailView';/,
  '策略独立详情页必须复用唯一的 PolicyDetailView');
assert.match(policyPage, /className=\{style\.policyStandaloneSurface\}/,
  '策略独立详情页必须提供可滚动工作区容器');
assert.doesNotMatch(policyPage, /Descriptions|<Tree|<Card|describeAuthPolicyDetail|resourceTreeData/,
  '策略独立详情页不能保留第二套旧式详情实现');
assert.match(policyView, /策略概要[\s\S]*策略标签[\s\S]*metadataEntries\.length/,
  '策略概要必须只展示标签数量，避免直接渲染未知 metadata 值');
assert.match(policyStyle, /\.policyStandaloneSurface[\s\S]*:global\(\.fluent-loading\)[\s\S]*height: 100%/,
  '策略独立详情页必须把 Loading 中间层纳入高度链');

for (const required of [
  'userDetailSummary',
  'userCredentialPanel',
  'userTagPanel',
  'userPolicySurface',
  'copyToClipboard',
  'resetUserToken',
  'enableUserToken',
  'encodeURIComponent',
  "title: '策略名称'",
  'width: 260',
]) {
  assert.match(userDetail, new RegExp(required), `用户详情必须保留 ${required}`);
}

for (const forbidden of ['Descriptions', '<Card', '<Form', '<Tabs', '<Popup', '\\[object Object\\]']) {
  assert.doesNotMatch(userDetail, new RegExp(forbidden), `用户详情不能回退到旧布局：${forbidden}`);
}

for (const className of [
  'userDetailPage',
  'detailBreadcrumb',
  'userDetailSummary',
  'userMetaGrid',
  'userCredentialPanel',
  'userTokenRow',
  'userPolicySurface',
]) {
  assert.match(principalStyle, new RegExp(`\\.${className}\\b`), `用户详情样式必须包含 ${className}`);
}

console.log('认证独立详情页布局验证通过');
