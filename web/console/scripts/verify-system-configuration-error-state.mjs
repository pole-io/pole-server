import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('../src/pages/SystemConfiguration/index.tsx', import.meta.url), 'utf8');
const catchBlock = source.match(/catch \(requestError\) \{([\s\S]*?)\n    \} finally/);

assert.ok(catchBlock, 'system configuration load must expose an explicit request-error branch');
assert.doesNotMatch(
  catchBlock[1],
  /setSettings\(\[\]\)/,
  'refresh failure must preserve the last valid configuration snapshot',
);
assert.match(source, /role=["']alert["']/, 'load failure must be announced as an error, not rendered as 0\/0');
assert.match(source, /settings\.length === 0 && error/, 'initial load failure must replace the empty directory workspace');
assert.match(source, /重新加载/, 'load failure must expose a retry action');
assert.doesNotMatch(source, /scrollIntoView/, '领域切换不能滚动整个页面并把页头推到顶栏下方');
assert.match(source, /closest\(`\.\$\{style\.domainNav\}`\)/, '领域切换滚动必须限制在领域导航自身');

console.log('system configuration error state contract verified');
