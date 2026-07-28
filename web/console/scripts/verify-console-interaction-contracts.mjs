#!/usr/bin/env node

import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = (file) => fs.readFileSync(file, 'utf8');
const fluent = read('src/components/Fluent/index.tsx');
const users = read('src/services/users.ts');
const groups = read('src/services/user_group.ts');
const lossless = read('src/services/lossless.ts');

assert.match(fluent, /Radio\.Button = RadioButton/, 'Fluent 适配层必须支持策略编辑器使用的 Radio.Button。');
assert.match(users, /\.user \?\? \(result as User\)/, '用户 Token 接口必须兼容解包后的直接 User。');
assert.match(groups, /\.userGroup \?\? \(result as UserGroup\)/, '用户组 Token 接口必须兼容解包后的直接 UserGroup。');
assert.match(lossless, /pole\.io\/lossless\/namespace/, '无损规则必须通过当前 spec 的 metadata 传递命名空间。');
assert.match(lossless, /pole\.io\/lossless\/service/, '无损规则必须通过当前 spec 的 metadata 传递服务。');

console.log('Console interaction contracts verified');
