import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');
const fileView = read('src/pages/Configuration/Group/Files/FileView.tsx');
const filePage = read('src/pages/Configuration/Group/Files/index.tsx');
const fileStyles = read('src/pages/Configuration/Group/Files/index.module.less');
const switcher = read('src/components/EnvironmentResourceSwitcher/index.tsx');
const switcherStyles = read('src/components/EnvironmentResourceSwitcher/index.module.less');
const codeEditor = read('src/components/CodeEditor/index.tsx');

const checks = [
  ['配置文件详情使用环境 Tab', fileView.includes('presentation="tabs"')],
  ['选择文件后隐藏重复的分组环境 Tab', filePage.includes('!selectedFileName') && filePage.includes('resourceLabel="配置分组"') && filePage.includes('presentation="tabs"')],
  ['资源工作台只保留一层详情 Tab', fileView.includes('label="文件内容"') && fileView.includes('label="基本信息"') && fileView.includes('label="发布记录"') && fileView.includes('label="订阅查询"')],
  ['外层页面不再维护配置编辑 Tab', !filePage.includes('label="配置编辑"') && !filePage.includes('activeTab')],
  ['文件身份与操作归入统一摘要头', fileView.includes('fileSummary') && fileView.includes('fileActions')],
  ['文件路径使用结构化上下文展示', fileView.includes('resourcePath')],
  ['基本信息只在对应 Tab 渲染', fileView.includes('renderBasicInfo') && fileView.includes('value="basic"')],
  ['正文只在文件内容 Tab 渲染', fileView.includes('value="content"')],
  ['环境切换器提供 Tab 展示模式', switcher.includes("presentation?: 'cards' | 'tabs'")],
  ['环境 Tab 使用标准 tab 语义', switcher.includes('role="tablist"') && switcher.includes('aria-selected={active}')],
  ['配置详情具备窄屏布局', fileStyles.includes('@container config-file-detail')],
  ['Monaco 编辑器跟随全局主题', codeEditor.includes("theme === ETheme.dark ? 'vs-dark' : 'vs'")],
  ['Monaco 不再固定亮色主题', !codeEditor.includes("theme={'vs'}")],
];

const failed = checks.filter(([, passed]) => !passed);
if (failed.length > 0) {
  for (const [label] of failed) console.error(`FAIL: ${label}`);
  process.exit(1);
}

for (const [label] of checks) console.log(`PASS: ${label}`);
