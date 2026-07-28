import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const buildScript = read('scripts/run-vite-build.mjs');

assert.match(
  buildScript,
  /env\.NODE_ENV\s*=\s*['"]production['"]/,
  'Vite build 必须显式设置 NODE_ENV=production；自定义 release/test/site mode 不能让 React 按开发运行时打包。',
);
assert.match(
  buildScript,
  /env\.VITE_USER_NODE_ENV\s*=\s*['"]production['"]/,
  'Vite 2 的自定义 mode 必须显式设置 VITE_USER_NODE_ENV=production，否则 @vitejs/plugin-react 仍会生成 jsxDEV 调试调用。',
);

const checkDist = process.env.CHECK_DIST === '1';
const assetsDir = path.join(root, 'dist/assets');

if (checkDist) {
  assert.ok(fs.existsSync(assetsDir), 'CHECK_DIST=1 时必须先完成 Vite 构建并生成 dist/assets。');

  const jsFiles = fs.readdirSync(assetsDir).filter((file) => file.endsWith('.js'));
  const devRuntimeFiles = jsFiles.filter((file) => {
    const content = fs.readFileSync(path.join(assetsDir, file), 'utf8');
    return /\.jsxDEV\(|\bjsxDEV\(|react-jsx-dev-runtime\.development|A props object containing a "key" prop/.test(content);
  });

  assert.deepEqual(
    devRuntimeFiles,
    [],
    `生产构建产物不能包含 React 开发 JSX runtime: ${devRuntimeFiles.join(', ')}`,
  );
}

console.log('production React build checks passed');
