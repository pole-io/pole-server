import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const sourceRoot = path.join(root, 'src');
const fluentThemePath = path.join(sourceRoot, 'styles', 'fluent.less');

const read = (file) => fs.readFileSync(file, 'utf8');
const walk = (directory) => fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
  const entryPath = path.join(directory, entry.name);
  return entry.isDirectory() ? walk(entryPath) : (entry.name.endsWith('.less') ? [entryPath] : []);
});

const legacyLightText = /(?<![-\w])color:\s*#(?:111827|1f2937|1f2329|243044|242424|334155|364152|374151|4b5563|4b5565|4e5969|5f6b7a|5f6673|667085|697586|6b7280|64748b|7b8494|7f8ca3|8492a6|86909c|8a8f99|8b93a3|8b95a5|94a3b8|9aa3b2|9ca3af|a0a8b5|666)\b/i;
const surfaceDeclaration = /(?:background(?:-color)?|border(?:-[\w-]+)?|--[\w-]*(?:surface|background|border|text|muted|weak)):\s*[^;{}]*#([0-9a-f]{3}|[0-9a-f]{6}|[0-9a-f]{8})\b/gi;

const hasHardcodedLightSurface = (source) => {
  surfaceDeclaration.lastIndex = 0;
  for (const match of source.matchAll(surfaceDeclaration)) {
    const raw = match[1].slice(0, 6);
    const hex = raw.length === 3 ? raw.split('').map((digit) => `${digit}${digit}`).join('') : raw;
    const channels = [0, 2, 4].map((offset) => Number.parseInt(hex.slice(offset, offset + 2), 16));
    const lightness = (Math.max(...channels) + Math.min(...channels)) / 510;
    if (lightness >= 0.72) {
      return true;
    }
  }
  return false;
};

const violations = walk(sourceRoot)
  .filter((file) => file !== fluentThemePath)
  .flatMap((file) => {
    const source = read(file);
    return legacyLightText.test(source) || hasHardcodedLightSurface(source) ? [path.relative(root, file)] : [];
  });

assert.deepEqual(violations, [], `暗色主题正文、表面和边框不能使用浅色硬编码：${violations.join(', ')}`);

const fluentTheme = read(fluentThemePath);
for (const token of ['--app-text-tertiary', '--app-muted', '--app-danger', '--app-brand-subtle', '--app-status-warning-subtle', '--app-status-error-subtle', '--app-status-success-subtle']) {
  assert.equal((fluentTheme.match(new RegExp(`${token}:`, 'g')) || []).length, 2, `${token} 必须同时定义亮色和暗色值`);
}

console.log('暗色主题 token 静态约束验证通过');
