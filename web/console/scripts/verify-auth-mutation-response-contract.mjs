#!/usr/bin/env node

import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = (file) => fs.readFileSync(file, 'utf8');

const serviceFiles = [
  'src/services/auth_policy.ts',
  'src/services/role.ts',
  'src/services/user_group.ts',
  'src/services/users.ts',
];

const services = serviceFiles.map((file) => ({ file, source: read(file) }));
for (const { file, source } of services) {
  assert.doesNotMatch(
    source,
    /Number\(\s*result\.code\s*\)/,
    `${file} 的写接口拿到的是 request 层解包结果，不能继续读取 result.code。`,
  );
}

const policyService = services[0].source;
assert.match(policyService, /export function isAuthMutationSuccessful\(response: unknown\): boolean/, '鉴权 mutation 必须使用统一成功判定。');
assert.match(policyService, /typeof mutation\.result === 'boolean'/, '统一成功判定必须支持解包后的 result 布尔值。');
assert.match(policyService, /Array\.isArray\(mutation\.responses\)/, '统一成功判定必须支持批量写 responses。');
assert.match(policyService, /mutation\.responses\.length > 0[\s\S]*mutation\.responses\.every/, '批量写必须存在响应项且全部成功。');
assert.match(policyService, /Number\(mutation\.code\) === SuccessCode/, '无 data 的旧鉴权响应仍保留顶层 code，统一判定必须兼容。');

for (const { file, source } of services.slice(1)) {
  assert.match(source, /isAuthMutationSuccessful/, `${file} 必须复用统一成功判定。`);
}

const thunkFiles = [
  'src/modules/auth/role.ts',
  'src/modules/user/users.ts',
  'src/modules/user/groups.ts',
];
for (const file of thunkFiles) {
  const source = read(file);
  assert.doesNotMatch(source, /await (?:createRoles|modifyRoles|deleteRoles|createUsers|modifyUsers|createUserGroup|modifyUserGroup|modifyUserGroupToken|refreshUserGroupToken|deleteUserGroups)\([^;]*\);\s*return fulfillWithValue/s, `${file} 不得忽略 mutation 的 false 结果并返回伪成功。`);
}

console.log('auth mutation response contracts verified');
