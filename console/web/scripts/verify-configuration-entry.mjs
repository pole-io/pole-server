import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const routerIndex = read('src/router/index.ts');
const configurationRoutes = read('src/router/modules/configuration.ts');
const configurationGroupIndex = read('src/pages/Configuration/Group/index.tsx');

assert.match(
  routerIndex,
  /const allRoutes = \[[\s\S]*\.\.\.configuration[\s\S]*\];/,
  '总路由必须展开 configuration 模块，让配置中心入口可达。',
);
assert.doesNotMatch(
  routerIndex,
  /\/\*\*\s*\.\.\.configuration\s*\*\//,
  'configuration 不能继续以注释形式留在总路由中。',
);

assert.match(
  configurationRoutes,
  /path: 'group'[\s\S]*title: 'menu\.configuration\.group'/,
  '配置中心菜单必须暴露已具备闭环的配置分组入口。',
);
assert.doesNotMatch(
  configurationRoutes,
  /path: 'kubernetes'[\s\S]*title: 'menu\.configuration\.k8s'/,
  'Kubernetes 配置页仍是未实现占位，不能作为可见菜单入口暴露。',
);
assert.match(
  configurationRoutes,
  /path: 'group\/files'[\s\S]*hidden: true/,
  '配置文件详情页应作为隐藏子路由从配置分组进入，不应成为侧边栏平级菜单。',
);

assert.match(
  configurationGroupIndex,
  /<ConfigGroupTable \/>/,
  '配置分组页必须直接展示配置分组表格。',
);
assert.doesNotMatch(
  configurationGroupIndex,
  /TemplateTable|label="模板"|<Tabs|TabPanel/,
  '配置分组页不能继续展示未实现的模板 Tab。',
);

console.log('configuration entry checks passed');
