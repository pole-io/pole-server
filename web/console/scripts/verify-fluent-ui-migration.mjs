import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');
const packageJson = JSON.parse(read('package.json'));

assert.ok(packageJson.dependencies['@fluentui/react-components'], '必须安装 Fluent UI React v9 组件包。');
assert.ok(packageJson.dependencies['@fluentui/react-icons'], '必须安装 Fluent UI React v9 图标包。');

const main = read('src/main.tsx');
assert.match(main, /FluentAppProvider/, '应用根节点必须接入 FluentAppProvider。');
assert.doesNotMatch(main, /tdesign-react/, 'main.tsx 不得加载 TDesign。');

const fluentProvider = read('src/components/Fluent/FluentAppProvider.tsx');
const fluentStyle = read('src/styles/fluent.less');
const appLayoutStyle = read('src/layouts/components/AppLayout.module.less');
const headerStyle = read('src/layouts/components/Header/index.module.less');
assert.match(fluentProvider, /className="fluent-provider-shell"/, '根 FluentProvider 必须有独立满高壳层，不能依赖 Portal Provider。');
assert.match(fluentStyle, /\.fluent-provider-shell,[\s\S]*?height:\s*100%;/, 'Fluent 应用根链路必须继承完整视口高度。');
assert.match(appLayoutStyle, /\.sidePanel\s*\{[\s\S]*?height:\s*100%;[\s\S]*?overflow:\s*hidden;/, '侧栏布局必须占满父级并使用内部滚动。');
assert.match(headerStyle, /\.panel\s*\{[\s\S]*?position:\s*sticky;[\s\S]*?background:\s*var\(--app-surface\);/, '全局 Header 是 sticky 层，必须使用不透明 surface 背景，不能让页面内容透出。');
assert.match(headerStyle, /\.panel\s*\{[\s\S]*?border-bottom:\s*1px solid var\(--app-border-subtle\);/, '全局 Header 必须使用 Fluent 边界色。');
assert.doesNotMatch(headerStyle, /\.panel\s*\{[\s\S]*?background:\s*transparent/, '全局 Header 不能使用透明背景。');

const fluentAdapter = read('src/components/Fluent/index.tsx');
assert.match(fluentAdapter, /@fluentui\/react-components/, 'Fluent 适配层必须渲染 Fluent UI 原生组件。');
assert.match(fluentAdapter, /export const Button/, 'Fluent 适配层必须提供统一按钮。');
assert.match(fluentAdapter, /export const Select/, 'Fluent 适配层必须提供统一选择器。');
assert.match(fluentAdapter, /export const Drawer/, 'Fluent 适配层必须提供统一抽屉。');
assert.match(fluentAdapter, /export const Tabs/, 'Fluent 适配层必须提供统一标签页。');

const fluentToast = read('src/components/Fluent/toast.tsx');
assert.match(fluentToast, /content instanceof Error/, 'Fluent Toast 必须把 Error 对象转换为可渲染消息。');
assert.match(fluentToast, /normalizeToastContent\(title\)/, 'Fluent Toast 标题必须经过内容归一化。');

const sourceFiles = [];
const walk = (directory) => {
  for (const entry of fs.readdirSync(path.join(root, directory), { withFileTypes: true })) {
    const relative = path.join(directory, entry.name);
    if (entry.isDirectory()) walk(relative);
    else if (/\.(ts|tsx)$/.test(entry.name)) sourceFiles.push(relative);
  }
};
walk('src');

for (const file of sourceFiles) {
  const source = read(file);
  assert.doesNotMatch(source, /from ['"]tdesign-react(?:\/[^'"]*)?['"]/, `${file} 不得直接导入 TDesign 组件。`);
  assert.doesNotMatch(source, /from ['"]tdesign-icons-react(?:\/[^'"]*)?['"]/, `${file} 不得直接导入 TDesign 图标。`);
}

const menu = read('src/layouts/components/Menu.tsx');
assert.match(menu, /NavCategory/, '侧边栏必须使用 Fluent Nav 组件。');
assert.match(menu, /openCategories=\{collapsed \? \[\] : categoryValues\}/, '展开侧边栏必须默认展开全部资源分组。');

const menuStyle = read('src/layouts/components/Menu.module.less');
assert.match(menuStyle, /\.fluentSidebar\s*\{[\s\S]*?height:\s*100%;[\s\S]*?min-height:\s*0;/, '侧边栏必须占满应用壳层高度。');
assert.match(menuStyle, /\.fluentNav :global\(\.fui-NavItem\)[\s\S]*?background:\s*transparent;/, '侧边栏普通菜单项必须使用连续透明表面，不能保留灰色卡片背景。');
assert.match(menuStyle, /\[aria-current='page'\][\s\S]*?background:\s*var\(--app-brand-1\);/, '侧边栏当前项必须使用支持明暗主题的 Fluent 品牌浅色 token。');
assert.match(menuStyle, /\[aria-current='page'\]\)::before[\s\S]*?width:\s*3px;/, '侧边栏当前项必须保留左侧品牌色指示条。');
assert.match(menuStyle, /\.collapsed \.fluentNav :global\(\.fui-NavCategoryItem__expandIcon\)[\s\S]*?display:\s*none;/, '折叠侧边栏必须隐藏分组展开箭头。');

console.log('Fluent UI migration checks passed');
