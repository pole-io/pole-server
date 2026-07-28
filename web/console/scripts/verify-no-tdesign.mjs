import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');
const packageJson = JSON.parse(read('package.json'));
const forbiddenPackages = ['tdesign-react', 'tdesign-icons-react'];

for (const packageName of forbiddenPackages) {
  assert.equal(
    packageJson.dependencies?.[packageName],
    undefined,
    `package.json dependencies 不得保留 ${packageName}。`,
  );
  assert.equal(
    packageJson.devDependencies?.[packageName],
    undefined,
    `package.json devDependencies 不得保留 ${packageName}。`,
  );
}

if (fs.existsSync(path.join(root, 'package-lock.json'))) {
  const packageLock = JSON.parse(read('package-lock.json'));
  for (const packageName of forbiddenPackages) {
    assert.equal(
      packageLock.packages?.['']?.dependencies?.[packageName],
      undefined,
      `package-lock.json 根依赖不得保留 ${packageName}。`,
    );
    assert.equal(
      packageLock.packages?.[`node_modules/${packageName}`],
      undefined,
      `package-lock.json 不得解析 ${packageName}。`,
    );
  }
  const lockedTDesignPackages = Object.keys(packageLock.packages ?? {}).filter((packagePath) =>
    /(?:^|\/)node_modules\/(?:tdesign-react|tdesign-icons-react)(?:\/|$)/.test(packagePath),
  );
  assert.deepEqual(
    lockedTDesignPackages,
    [],
    `package-lock.json 不得保留嵌套 TDesign React 包：\n${lockedTDesignPackages.join('\n')}`,
  );
}

const sourceFiles = [];
const walk = (directory) => {
  for (const entry of fs.readdirSync(path.join(root, directory), { withFileTypes: true })) {
    const relative = path.join(directory, entry.name);
    if (entry.isDirectory()) walk(relative);
    else sourceFiles.push(relative);
  }
};
walk('src');

const violations = [];
const checks = [
  {
    pattern: /(?:from\s*|import\s*\(|require\s*\()\s*['"](?:tdesign-react|tdesign-icons-react)(?:\/[^'"]*)?['"]/g,
    reason: '仍直接加载 TDesign JavaScript/TypeScript 模块',
  },
  {
    pattern: /(?:@import|import)[^;\n]*tdesign-react[^;\n]*\.css/gi,
    reason: '仍加载 TDesign CSS',
  },
  {
    pattern: /(?:^|[^A-Za-z0-9_-])\.t-[A-Za-z0-9_-]+/gm,
    reason: '仍依赖 TDesign DOM 类名选择器',
  },
  {
    pattern: /--td-[A-Za-z0-9_-]+/g,
    reason: '仍依赖 TDesign CSS 变量',
  },
];

for (const file of sourceFiles) {
  const source = read(file);
  for (const { pattern, reason } of checks) {
    pattern.lastIndex = 0;
    for (const match of source.matchAll(pattern)) {
      const line = source.slice(0, match.index).split('\n').length;
      violations.push(`${file}:${line} ${reason}: ${match[0].trim()}`);
    }
  }
  if (file.endsWith('.tsx')) {
    for (const match of source.matchAll(/<(?:button|input|select|textarea)\b/g)) {
      const line = source.slice(0, match.index).split('\n').length;
      violations.push(`${file}:${line} 仍使用原生交互标签，必须改用 Fluent UI 控件: ${match[0]}`);
    }
  }
}

assert.deepEqual(
  violations,
  [],
  `必须完成零 TDesign 验收，发现 ${violations.length} 处残留：\n${violations.join('\n')}`,
);

console.log('Zero TDesign verification passed');
