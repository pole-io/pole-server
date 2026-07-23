import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const route = read('src/router/modules/systemConfiguration.ts');
const router = read('src/layouts/components/AppRouter.tsx');
const menu = read('src/layouts/components/Menu.tsx');
const loginState = read('src/modules/user/login.ts');
const loginService = read('src/services/login.ts');

assert.match(route, /adminOnly:\s*true/, '系统配置路由必须声明为 admin-only。');
assert.match(
  router,
  /const AdminRoute[\s\S]*role !== 'main'\s*&&\s*role !== 'admin'[\s\S]*<Navigate to="\/namespace" replace \/>/,
  '非 main/admin 角色直达系统配置页时必须被路由守卫拒绝。',
);
assert.match(
  menu,
  /const isAdmin = role === 'main' \|\| role === 'admin';/,
  '侧边栏必须对主账号和绑定内置 admin 角色的账号展示 admin-only 路由。',
);
assert.doesNotMatch(
  loginState,
  /role:\s*localStorage\.getItem\(LoginRoleKey\)/,
  'admin-only 判定不能信任 localStorage 中可能过期或被篡改的角色。',
);
assert.match(loginState, /sessionResolved/, '受保护路由渲染前必须等待签名会话角色解析完成。');
assert.match(loginState, /hydrateSession/, '页面刷新时必须从服务端签名会话恢复权威角色。');
assert.match(loginService, /\/auth\/v1\/user\/session/, '前端必须通过当前会话接口读取权威角色。');
assert.match(router, /!sessionResolved[\s\S]*<Loading/, '会话角色未解析时必须 fail-closed，不渲染目标页面。');

console.log('system configuration admin gate checks passed');
