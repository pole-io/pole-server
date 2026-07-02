import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const files = {
  userTable: 'src/pages/Auth/Principal/UserTable.tsx',
  groupTable: 'src/pages/Auth/Principal/GroupTable.tsx',
  roleTable: 'src/pages/Auth/Principal/RoleTable.tsx',
  policyTable: 'src/pages/Auth/Policy/PolicyTable.tsx',
  userEditor: 'src/pages/Auth/Principal/UserEditor.tsx',
  groupEditor: 'src/pages/Auth/Principal/GroupEditor.tsx',
  roleEditor: 'src/pages/Auth/Principal/RoleEditor.tsx',
  principalPolicyTable: 'src/pages/Auth/Principal/PrincipalPolicyTable.tsx',
  authPolicyService: 'src/services/auth_policy.ts',
  policyEditor: 'src/pages/Auth/Policy/PolicyEditor.tsx',
  policyDetailView: 'src/pages/Auth/Policy/PolicyDetailView.tsx',
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

for (const [key, source] of Object.entries({
  userTable: sources.userTable,
  groupTable: sources.groupTable,
  roleTable: sources.roleTable,
  policyTable: sources.policyTable,
})) {
  assert.match(source, /查看 \/ 编辑/, `${key} 操作列必须提供统一查看 / 编辑入口`);
  assert.doesNotMatch(source, /navigate\(`?(userdetail|groupdetail|roledetail|detail)\?/,
    `${key} 列表主入口不应继续跳转旧详情路由`);
}

assert.match(sources.userTable, /operateUser\('view', 'user', \{ \.\.\.row \}\)/,
  '用户名称和主操作必须打开用户详情抽屉');
assert.match(sources.groupTable, /handleEdit(Group|UserGroup)\(row, 'view', 'group'\)/,
  '用户组名称和主操作必须打开用户组详情抽屉');
assert.match(sources.groupTable, /describeUserGroupToken/,
  '用户组 Token 操作不能继续为空实现');
assert.match(sources.roleTable, /handleEditRole\(row, 'view', 'role'\)/,
  '角色名称和主操作必须打开角色详情抽屉');
assert.match(sources.policyTable, /handleEditPolicy\(row, 'view', 'policy_rule'\)/,
  '策略名称和主操作必须打开策略详情抽屉');

assert.match(sources.userEditor, /用户详情/, '用户抽屉必须有详情态标题');
assert.match(sources.groupEditor, /用户组详情/, '用户组抽屉必须有详情态标题');
assert.match(sources.roleEditor, /角色详情/, '角色抽屉必须有详情态标题');
assert.match(sources.policyEditor, /策略详情/, '策略抽屉必须有详情态标题');
assert.match(sources.userEditor, /<PrincipalPolicyTable principalId=\{editUser\?\.id\} principalType=\{1\} \/>/,
  '用户详情抽屉必须保留权限信息视图');
assert.match(sources.groupEditor, /<PrincipalPolicyTable principalId=\{viewGroup\?\.id\} principalType=\{2\} \/>/,
  '用户组详情抽屉必须保留权限信息视图');
assert.match(sources.roleEditor, /<PrincipalPolicyTable principalId=\{viewRole\?\.id\} principalType=\{3\} \/>/,
  '角色详情抽屉必须保留权限信息视图');
assert.match(sources.principalPolicyTable, /describeAuthPolicies/,
  '主体权限视图必须通过关联策略接口加载权限信息');
assert.match(sources.principalPolicyTable, /PolicyDetailView/,
  '主体权限视图中的关联策略必须能打开策略详情抽屉');
assert.match(sources.principalPolicyTable, /onClick=\{\(\) => setSelectedPolicy\(row as PolicyRule\)\}/,
  '关联策略名称点击必须进入策略详情抽屉');
assert.match(sources.policyEditor, /<PolicyDetailView policyId=\{currentPolicy\.id\} \/>/,
  '策略查看态必须渲染专用详情视图，不能复用禁用编辑器充当详情');
assert.match(sources.policyDetailView, /成员信息/, '策略详情视图必须保留成员信息');
assert.match(sources.policyDetailView, /资源信息/, '策略详情视图必须保留资源信息');
assert.match(sources.policyDetailView, /资源标签/, '策略详情视图必须保留资源标签');
assert.match(sources.policyDetailView, /可访问接口/, '策略详情视图必须保留接口范围');
assert.match(sources.authPolicyService, /authStrategy\?: PolicyRule/,
  '策略详情 service 必须兼容 authStrategy 包裹返回');
assert.match(sources.authPolicyService, /DescribeAuthPolicyDetailResponse\)\.authStrategy \?\? \(result as PolicyRule\)/,
  '策略详情 service 必须兼容接口直接返回 AuthStrategy 的形态');
assert.match(sources.authPolicyService, /lossless_rules\?: PolicyResource\[\]/,
  '策略资源类型必须包含无损规则资源');
assert.match(sources.policyDetailView, /value: 'lossless_rules'/,
  '策略详情资源树必须展示无损规则资源');

for (const [key, source] of Object.entries({
  userEditor: sources.userEditor,
  groupEditor: sources.groupEditor,
  roleEditor: sources.roleEditor,
})) {
  assert.match(source, /setEditable\(true\)/, `${key} 详情态必须能在抽屉内切到编辑态`);
}

assert.match(sources.policyEditor, /footer=\{policyViewMode \? false : drawerFooter\}/,
  '策略详情查看态必须是纯查看抽屉，不再提供编辑/关闭 footer');

console.log('auth drawer action verification passed');
