import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const pagePath = path.join(root, 'src/pages/Namespace/index.tsx');
const editorPath = path.join(root, 'src/pages/Namespace/NamespaceEditor.tsx');
const servicePath = path.join(root, 'src/services/namespace.ts');
const authorizePath = path.join(root, 'src/components/Authorize/index.tsx');

for (const file of [pagePath, editorPath, servicePath, authorizePath]) {
  if (!fs.existsSync(file)) {
    throw new Error(`missing file: ${file}`);
  }
}

const pageSource = fs.readFileSync(pagePath, 'utf8');
const editorSource = fs.readFileSync(editorPath, 'utf8');
const serviceSource = fs.readFileSync(servicePath, 'utf8');
const authorizeSource = fs.readFileSync(authorizePath, 'utf8');

assert.match(pageSource, /removeNamespace/, '命名空间列表必须导入并调用删除 action');
assert.match(pageSource, /operateNamespace\('view', row\)/, '行内主入口必须是查看');
assert.match(pageSource, /operateNamespace\('authorize', row\)/, '命名空间作为资源，列表行操作必须提供授权入口');
assert.match(pageSource, /components\/OperationButton/, '命名空间行操作必须使用统一 OperationButton 组件');
assert.match(pageSource, /action="viewEdit"|action="view"/, '统一入口 tooltip 和 aria-label 必须表达查看');
assert.match(pageSource, /action="authorize"/, '命名空间授权入口必须使用统一授权图标');
assert.match(pageSource, /resource_type=\{PolicySourceType\.Namespaces\}/, '命名空间授权必须使用 Namespaces 资源类型');
assert.match(pageSource, /resource_id=\{editorState\.data\?\.name/, '命名空间授权 resource_id 必须使用命名空间名称');
assert.match(pageSource, /resource_name=\{`\$\{editorState\.data\?\.name\}`\}/, '命名空间授权 resource_name 必须使用命名空间名称');
assert.match(pageSource, /openInfoNotification\('请求成功', '删除命名空间成功'\)/, '删除成功后必须反馈成功通知');
assert.match(pageSource, /refreshTable\(page, limit, query\)/, '删除成功后必须刷新当前列表');

assert.match(editorSource, /op === 'view'/, '命名空间抽屉必须支持查看态');
assert.match(editorSource, /命名空间详情/, '查看态标题必须是命名空间详情');
assert.match(editorSource, /<Button theme="primary" onClick=\{\(\) => setEditable\(true\)\}>/, '查看态必须能进入编辑');
assert.match(editorSource, /<Button theme="default" onClick=\{closeDrawer\}>/, '查看态必须能关闭抽屉');

assert.match(serviceSource, /token\?: string/, '删除命名空间请求的 token 应为可选，匹配 Console API 只传 name 的用法');
assert.match(authorizeSource, /props\.resource_type === PolicySourceType\.Namespaces/, '授权抽屉必须针对命名空间资源单独展示资源摘要');
assert.match(authorizeSource, /!isNamespaceResource/, '命名空间授权摘要不能继续显示空的资源名称分段');

console.log('namespace action verification passed');
