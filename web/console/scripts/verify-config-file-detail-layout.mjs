import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');
const fileView = read('src/pages/Configuration/Group/Files/FileView.tsx');
const filePage = read('src/pages/Configuration/Group/Files/index.tsx');
const fileStyles = read('src/pages/Configuration/Group/Files/index.module.less');
const groupWorkspaceNav = read('src/pages/Configuration/Group/GroupWorkspaceNav.tsx');
const groupWorkspaceStyles = read('src/pages/Configuration/Group/GroupWorkspaceNav.module.less');
const switcher = read('src/components/EnvironmentResourceSwitcher/index.tsx');
const switcherStyles = read('src/components/EnvironmentResourceSwitcher/index.module.less');
const codeEditor = read('src/components/CodeEditor/index.tsx');

const checks = [
  ['配置分组使用统一清单工作区', filePage.includes('<GroupWorkspaceNav') && filePage.includes('配置清单')],
  ['配置文件与全局模板在同一资源树', filePage.includes('renderTree(datas, templates)') && filePage.includes("resourceKind: 'template'")],
  ['模板节点在右侧画布就地展示', filePage.includes('<TemplateWorkspace') && filePage.includes('embedded')],
  ['分组导航只表达当前位置', groupWorkspaceNav.includes('<Breadcrumb') && !groupWorkspaceNav.includes('配置分组工作区') && !groupWorkspaceNav.includes('role="tab"')],
  ['配置清单顶部环境 Tab 始终存在', filePage.includes('resourceLabel="配置分组"') && filePage.includes('presentation="tabs"') && !filePage.includes('{!editState.activeNode && (')],
  ['文件详情不再重复渲染环境 Tab', !fileView.includes('EnvironmentResourceSwitcher')],
  ['分组导航只保留面包屑', groupWorkspaceNav.includes('<Breadcrumb') && !groupWorkspaceNav.includes('workspaceBar') && !groupWorkspaceStyles.includes('.workspaceBar')],
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
