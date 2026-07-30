import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = (file) => fs.readFileSync(file, 'utf8');

const principal = read('src/pages/Auth/Principal/index.tsx');
const principalStyle = read('src/pages/Auth/Principal/index.module.less');
const userTable = read('src/pages/Auth/Principal/UserTable.tsx');
const groupTable = read('src/pages/Auth/Principal/GroupTable.tsx');
const roleTable = read('src/pages/Auth/Principal/RoleTable.tsx');

assert.match(
  principal,
  /Promise\.allSettled\(\[[\s\S]*describeUsers\(\{ offset: 0, limit: 1 \}\)[\s\S]*describeUserGroups\(\{ offset: 0, limit: 1 \}\)[\s\S]*describeRoles\(\{ offset: 0, limit: 1 \}\)/,
  '身份主体总览必须并行读取三类资源的完整总数。',
);
assert.match(
  principal,
  /总用户[\s\S]*用户组[\s\S]*角色数[\s\S]*aria-label="身份主体统计"/,
  '身份主体页顶部必须展示总用户、用户组和角色数。',
);
assert.match(
  principal,
  /<UserTable onTotalChange=\{updateUsersTotal\} onTotalsRefresh=\{loadTotals\}[\s\S]*<GroupsTable onTotalChange=\{updateGroupsTotal\} onTotalsRefresh=\{loadTotals\}[\s\S]*<RoleTable onTotalChange=\{updateRolesTotal\} onTotalsRefresh=\{loadTotals\}/,
  '三个列表刷新后必须能把最新总数回写到顶部总览。',
);
assert.match(userTable, /!loading && !query[\s\S]*onTotalChange\?\.\(total\)/, '用户列表只能用未筛选总数更新总览。');
assert.match(groupTable, /if \(!query\) onTotalChange\?\.\(response\.totalCount\)/, '用户组列表只能用未筛选总数更新总览。');
assert.match(roleTable, /if \(!searchParam\) onTotalChange\?\.\(response\.totalCount\)/, '角色列表只能用未筛选总数更新总览。');
for (const [name, source] of [
  ['用户', userTable],
  ['用户组', groupTable],
  ['角色', roleTable],
]) {
  assert.match(source, /onTotalsRefresh\?\.\(\)/, `${name}变更后必须触发三类总数重新加载。`);
}
assert.match(
  principalStyle,
  /\.principalSummary[\s\S]*grid-template-columns: repeat\(3, minmax\(0, 1fr\)\)[\s\S]*\.principalSummaryItem/,
  '身份主体统计必须使用三列统一统计卡片。',
);
assert.match(
  principalStyle,
  /@media \(max-width: 820px\)[\s\S]*\.principalSummary[\s\S]*grid-template-columns: 1fr/,
  '窄屏下身份主体统计必须退化为单列。',
);

console.log('身份主体顶部统计总览契约检查通过。');
