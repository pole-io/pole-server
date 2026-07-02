import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const pagePath = path.join(root, 'src/pages/Namespace/index.tsx');
const editorPath = path.join(root, 'src/pages/Namespace/NamespaceEditor.tsx');
const servicePath = path.join(root, 'src/services/namespace.ts');

for (const file of [pagePath, editorPath, servicePath]) {
  if (!fs.existsSync(file)) {
    throw new Error(`missing file: ${file}`);
  }
}

const pageSource = fs.readFileSync(pagePath, 'utf8');
const editorSource = fs.readFileSync(editorPath, 'utf8');
const serviceSource = fs.readFileSync(servicePath, 'utf8');

assert.match(pageSource, /removeNamespace/, '命名空间列表必须导入并调用删除 action');
assert.match(pageSource, /operateNamespace\('view', row\)/, '行内主入口必须是查看 / 编辑');
assert.match(pageSource, /查看 \/ 编辑/, '统一入口 tooltip 和 aria-label 必须表达查看 / 编辑');
assert.doesNotMatch(pageSource, /operateNamespace\('authorize', row\)/, '授权不应继续作为行内并列图标');
assert.match(pageSource, /openInfoNotification\('请求成功', '删除命名空间成功'\)/, '删除成功后必须反馈成功通知');
assert.match(pageSource, /refreshTable\(page, limit, query\)/, '删除成功后必须刷新当前列表');

assert.match(editorSource, /onAuthorize\?: \(\) => void/, '命名空间抽屉必须提供授权入口回调');
assert.match(editorSource, /op === 'view'/, '命名空间抽屉必须支持查看态');
assert.match(editorSource, /命名空间详情/, '查看态标题必须是命名空间详情');
assert.match(editorSource, /<Button theme="primary" onClick=\{\(\) => setEditable\(true\)\}>/, '查看态必须能进入编辑');
assert.match(editorSource, /<Button theme="default" onClick=\{onAuthorize\}>/, '查看态必须能打开授权');
assert.match(editorSource, /<Button theme="default" onClick=\{closeDrawer\}>/, '查看态必须能关闭抽屉');

assert.match(serviceSource, /token\?: string/, '删除命名空间请求的 token 应为可选，匹配 Console API 只传 name 的用法');

console.log('namespace action verification passed');
