import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const pageSource = fs.readFileSync(path.join(scriptDir, '../src/pages/Namespace/index.tsx'), 'utf8');

assert.match(pageSource, /DEFAULT_NAMESPACE = 'default'/, '前端必须识别默认命名空间');
assert.match(pageSource, /SYSTEM_NAMESPACE = 'pole-system'/, '前端必须识别 Pole 内部系统命名空间');
assert.match(pageSource, /默认命名空间不可删除/, 'default 删除入口必须给出不可删除原因');
assert.match(pageSource, /Pole 内部系统空间不可删除/, 'pole-system 删除入口必须给出不可删除原因');
assert.match(pageSource, /内部系统空间/, 'pole-system 必须在命名空间列表展示内部系统空间标识');
assert.match(pageSource, /disabled=\{Boolean\(protectedReason\) \|\| row\.deleteable === false\}/, '系统命名空间删除按钮必须在权限判断之外强制禁用');
assert.match(pageSource, /if \(protectedNamespaceReason\(row as NamespaceView\)\)/, '删除 dispatch 前必须按类型再次拦截系统命名空间');

console.log('protected namespace verification passed');
