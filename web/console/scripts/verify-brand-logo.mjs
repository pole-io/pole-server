import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const fullLogo = read('src/assets/svg/assets-logo-full.svg');
const miniLogo = read('src/assets/svg/assets-t-logo.svg');
const menu = read('src/layouts/components/Menu.tsx');
const menuStyle = read('src/layouts/components/Menu.module.less');
const globalStyle = read('src/styles/index.less');

assert.match(fullLogo, /aria-label="Lattice\.Hub"/, '完整 Logo 的可访问名称必须是 Lattice.Hub。');
assert.match(fullLogo, />Lattice\.Hub</, '完整 Logo 文案必须展示 Lattice.Hub。');
assert.match(fullLogo, /href="data:image\/png;base64,iVBORw0KGgoAAAANSUhEUgAAAZAAAAGQ/, '完整 Logo 必须内嵌用户提供的产品 Logo PNG，保证图形一致。');
assert.match(miniLogo, /aria-label="Lattice\.Hub"/, '折叠态 Logo 的可访问名称必须是 Lattice.Hub。');
assert.match(miniLogo, /href="data:image\/png;base64,iVBORw0KGgoAAAANSUhEUgAAAZAAAAGQ/, '折叠态 Logo 必须内嵌用户提供的产品 Logo PNG。');

assert.doesNotMatch(`${fullLogo}\n${miniLogo}\n${menu}`, /Pole\.IO/, '品牌展示区域不能继续出现 Pole.IO。');
assert.match(menu, /`Lattice\.Hub \$\{version\}`/, '侧边栏底部版本文案必须使用 Lattice.Hub。');
assert.match(menuStyle, /\.fluentSidebar\s*\{[\s\S]*?width:\s*232px;[\s\S]*?flex:\s*0 0 232px;/, 'Fluent 侧边栏展开宽度必须保持 232px。');
assert.match(menuStyle, /width:\s*184px;[\s\S]*height:\s*32px;/, '侧边栏完整 Logo 尺寸应恢复原有布局规格。');
assert.doesNotMatch(globalStyle, /--console-brand:/, '恢复原状后不应在全局样式中注入 Console 设计 token。');

console.log('brand logo verification passed');
