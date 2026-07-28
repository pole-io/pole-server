import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const viteConfig = fs.readFileSync(path.join(root, 'vite.config.js'), 'utf8');

assert.doesNotMatch(
  viteConfig,
  /manualChunks/,
  'UI 框架迁移期间禁止按第三方库内部目录手工拆包，避免 React、Fluent 和兼容层形成循环 chunk。',
);
assert.doesNotMatch(
  viteConfig,
  /react-vendor|tdesign-(?:shared|data-form|base-overlay|navigation|misc)/,
  '不得恢复已验证会造成运行时初始化循环的 React/TDesign 手工 chunk。',
);
assert.match(
  viteConfig,
  /build:\s*\{[\s\S]*?cssCodeSplit:\s*true/,
  '生产构建仍应保留 CSS 按路由拆分。',
);

console.log('vite chunk safety checks passed');
