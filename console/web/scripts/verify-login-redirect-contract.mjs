import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const read = (relativePath) => readFileSync(new URL(`../${relativePath}`, import.meta.url), 'utf8');

const router = read('src/layouts/components/AppRouter.tsx');
const loginPage = read('src/pages/Login/index.tsx');

assert.match(
  router,
  /<Navigate to="\/login" replace state=\{\{ from: location \}\} \/>/,
  '受保护页面跳转登录页时必须保留来源位置，供登录成功后返回',
);
assert.doesNotMatch(
  loginPage,
  /MessagePlugin\.warning\(['"]您当前未登录，请先登录['"]\)/,
  '登录页已经表达登录任务，不应再弹出重复的未登录 Toast',
);

console.log('Login redirect contract verified.');
