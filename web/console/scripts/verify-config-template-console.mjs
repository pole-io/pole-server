import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8');
const assertMatch = (source, pattern, message) => {
  if (!pattern.test(source)) throw new Error(message);
};

const router = read('src/router/modules/configuration.ts');
const service = read('src/services/config_templates.ts');
const configFileService = read('src/services/config_files.ts');
const configReleaseService = read('src/services/config_release.ts');
const templatePage = read('src/pages/Configuration/Template/index.tsx');
const groupWorkspaceNav = read('src/pages/Configuration/Group/GroupWorkspaceNav.tsx');
const fileView = read('src/pages/Configuration/Group/Files/FileView.tsx');
const fileCreator = read('src/pages/Configuration/Group/Files/FileCreator.tsx');
const releaseDetail = read('src/pages/Configuration/Group/Releases/ReleaseDetail.tsx');

assertMatch(
  router,
  /path:\s*'templates'[\s\S]*pages\/Configuration\/Template[\s\S]*menu\.configuration\.template[\s\S]*hidden:\s*true/,
  '配置模板工作区必须保留深链路由，但不能占用一级菜单。',
);
assertMatch(
  router,
  /path:\s*'group\/templates'[\s\S]*pages\/Configuration\/Template[\s\S]*hidden:\s*true/,
  '配置分组必须提供独立的模板工作区路由。',
);
assertMatch(
  templatePage,
  /requestedGroup[\s\S]*<GroupWorkspaceNav[\s\S]*active="templates"/,
  '从配置分组进入模板页时必须展示分组级工作区切换。',
);
assertMatch(
  groupWorkspaceNav,
  /配置文件[\s\S]*环境 → 文件[\s\S]*配置模板[\s\S]*模板 → 环境 Value/,
  '配置分组工作区必须明确文件与模板的不同导航顺序。',
);

for (const [pattern, message] of [
  [/CONFIG_TEMPLATE\}\/preview/, '必须接入服务端预览接口。'],
  [/CONFIG_TEMPLATE\}\/releases/, '必须接入模板发布与版本查询接口。'],
  [/CONFIG_TEMPLATE\}\/values/, '必须接入 Namespace Value 草稿接口。'],
  [/CONFIG_TEMPLATE\}\/values\/releases/, '必须接入 Namespace Value 发布接口。'],
  [/CONFIG_TEMPLATE\}\/bindings/, '必须接入 ConfigFile 模板绑定接口。'],
  [/bindConfigFileTemplate[\s\S]*content:\s*file\.content[\s\S]*template_binding/, 'Binding 请求必须原子携带文件草稿字段。'],
  [/code:\s*number[\s\S]*diagnostics:/, 'RenderPreview 必须区分统一 API code 与渲染 diagnostics。'],
]) {
  assertMatch(service, pattern, message);
}

for (const [pattern, message] of [
  [/普通配置[\s\S]*模板配置/, '新建配置必须显式选择普通配置或模板配置。'],
  [/describeConfigTemplateReleases/, '模板配置创建必须加载不可变 Template Release。'],
  [/templateBinding/, '模板配置创建必须通过单次 ConfigFile 创建请求携带 binding。'],
  [/在新标签页管理模板与当前 Namespace Value/, '创建流程必须在保留未保存表单的前提下提供模板库入口。'],
  [/configuration\/group\/templates\?group=/, '创建流程必须保留配置分组上下文进入模板工作区。'],
  [/window\.open[\s\S]*'_blank'[\s\S]*'noopener,noreferrer'/, '创建流程必须在隔离标签页打开模板工作区，避免卸载未保存表单。'],
  [/visibilitychange[\s\S]*refreshTemplates/, '返回创建页时必须刷新模板目录。'],
  [/refreshWhenReturning[\s\S]*refreshTemplateReleases\(metaValues\.templateId\)/, '返回创建页时必须同步刷新当前模板的发布版本。'],
]) {
  assertMatch(fileCreator, pattern, message);
}

for (const [pattern, message] of [
  [/pole-mustache-v1/, '模板页面必须明确展示固定引擎版本。'],
  [/参数 Schema/, '模板页面必须提供参数 Schema 编辑能力。'],
  [/label="Namespace Value"/, '模板 Value 页签必须使用 Namespace 领域术语。'],
  [/Namespace Value/, '模板页面必须按 Namespace 维护 Value。'],
  [/GrayRuleEditor/, 'Value 灰度发布必须复用现有灰度规则编辑器。'],
  [/renderedSha256|rendered_sha256/, '模板页面必须展示服务端参考哈希。'],
  [/diagnostics/, '模板页面必须展示渲染诊断。'],
]) {
  assertMatch(templatePage, pattern, message);
}

for (const [pattern, message] of [
  [/configType/, 'ConfigFile 详情必须识别普通文本和模板类型。'],
  [/bindConfigFileTemplate/, 'ConfigFile 详情必须通过显式 binding 接口切换模板版本。'],
  [/templateReleaseId/, 'ConfigFile 详情必须保存明确的 Template Release。'],
  [/普通文本[\s\S]*模板渲染|模板渲染[\s\S]*普通文本/, 'ConfigFile 编辑态必须显式提供两种配置类型。'],
]) {
  assertMatch(fileView, pattern, message);
}

for (const [source, pattern, message] of [
  [configFileService, /config_type[\s\S]*template_binding/, 'ConfigFile service 必须双向映射配置类型与模板绑定。'],
  [configReleaseService, /config_type[\s\S]*template_binding/, '发布快照 service 必须映射当时固化的模板绑定。'],
  [releaseDetail, /配置类型[\s\S]*Template Release[\s\S]*Binding Release/, '发布详情必须展示固化的模板版本。'],
]) {
  assertMatch(source, pattern, message);
}

console.log('配置模板 Console 契约检查通过。');
