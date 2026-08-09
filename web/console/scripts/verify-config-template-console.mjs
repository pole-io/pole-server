import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8');
const assertMatch = (source, pattern, message) => {
  if (!pattern.test(source)) throw new Error(message);
};
const assertNotMatch = (source, pattern, message) => {
  if (pattern.test(source)) throw new Error(message);
};

const router = read('src/router/modules/configuration.ts');
const service = read('src/services/config_templates.ts');
const configFileService = read('src/services/config_files.ts');
const configReleaseService = read('src/services/config_release.ts');
const templatePage = read('src/pages/Configuration/Template/index.tsx');
const templateStyles = read('src/pages/Configuration/Template/index.module.less');
const schemaEditor = read('src/pages/Configuration/Template/SchemaEditor.tsx');
const valueEditor = read('src/pages/Configuration/Template/ValueEditor.tsx');
const versionManager = read('src/pages/Configuration/Template/VersionManager.tsx');
const renderPreviewPanel = read('src/pages/Configuration/Template/RenderPreviewPanel.tsx');
const groupWorkspaceNav = read('src/pages/Configuration/Group/GroupWorkspaceNav.tsx');
const filePage = read('src/pages/Configuration/Group/Files/index.tsx');
const fileView = read('src/pages/Configuration/Group/Files/FileView.tsx');
const fileCreator = read('src/pages/Configuration/Group/Files/FileCreator.tsx');
const releaseDetail = read('src/pages/Configuration/Group/Releases/ReleaseDetail.tsx');

assertMatch(
  router,
  /path:\s*'templates'[\s\S]*pages\/Configuration\/Template[\s\S]*menu\.configuration\.template[\s\S]*hidden:\s*true/,
  '配置模板工作区必须保留深链路由，但不能占用一级菜单。'
);
assertMatch(
  router,
  /path:\s*'group\/templates'[\s\S]*pages\/Configuration\/Template[\s\S]*hidden:\s*true/,
  '配置分组必须提供独立的模板工作区路由。'
);
assertMatch(
  templatePage,
  /requestedGroup[\s\S]*<GroupWorkspaceNav[\s\S]*group=\{requestedGroup\}/,
  '从配置分组进入模板页时必须保留分组上下文。'
);
assertNotMatch(
  groupWorkspaceNav,
  /配置文件[\s\S]*配置模板|activeLink|templatesHref/,
  '配置分组不得再用双入口区分配置文件与配置模板。'
);

for (const [pattern, message] of [
  [/CONFIG_TEMPLATE\}\/preview/, '必须接入服务端预览接口。'],
  [/CONFIG_TEMPLATE\}\/releases/, '必须接入模板发布与版本查询接口。'],
  [/CONFIG_TEMPLATE\}\/values/, '必须接入 Namespace Value 草稿接口。'],
  [/CONFIG_TEMPLATE\}\/environment-releases/, '必须接入环境配置组合发布接口。'],
  [/CONFIG_TEMPLATE\}\/bindings/, '必须接入 ConfigFile 模板绑定接口。'],
  [/CONFIG_TEMPLATE\}\/labels/, '必须接入模板标签读写接口。'],
  [/res\.labels\s*\?\?\s*res\.value\?\.labels/, '模板标签查询必须兼容 protobuf Struct 的 data.value 包装。'],
  [
    /bindConfigFileTemplate[\s\S]*content:\s*file\.content[\s\S]*template_binding/,
    'Binding 请求必须原子携带文件草稿字段。',
  ],
  [/code:\s*number[\s\S]*diagnostics:/, 'RenderPreview 必须区分统一 API code 与渲染 diagnostics。'],
]) {
  assertMatch(service, pattern, message);
}

for (const [pattern, message] of [
  [/内容来源[\s\S]*直接文本[\s\S]*配置模板/, '新建配置必须首先明确选择直接文本或配置模板。'],
  [/describeNamespaceTemplateValueReleases/, '模板配置创建必须加载当前环境的组合版本。'],
  [/templateBinding/, '模板配置创建必须通过单次 ConfigFile 创建请求携带 binding。'],
  [/在新标签页管理模板与当前 Namespace Value/, '创建流程必须在保留未保存表单的前提下提供模板库入口。'],
  [/configuration\/group\/templates\?group=/, '创建流程必须保留配置分组上下文进入模板工作区。'],
  [
    /window\.open[\s\S]*'_blank'[\s\S]*'noopener,noreferrer'/,
    '创建流程必须在隔离标签页打开模板工作区，避免卸载未保存表单。',
  ],
  [/visibilitychange[\s\S]*refreshTemplates/, '返回创建页时必须刷新模板目录。'],
  [
    /refreshWhenReturning[\s\S]*refreshEnvironmentReleases\(metaValues\.templateId\)/,
    '返回创建页时必须同步刷新当前模板的环境配置版本。',
  ],
  [/当前环境配置版本[\s\S]*模板与 Value 已绑定/, '创建模板配置时不得再让用户手工拼装模板与 Value 版本。'],
]) {
  assertMatch(fileCreator, pattern, message);
}

assertMatch(
  filePage,
  /renderNodeLabel[\s\S]*configType === 'CONFIG_TEMPLATE'[\s\S]*模板[\s\S]*文本/,
  '统一配置文件树必须展示直接文本与配置模板的来源类型。'
);
assertMatch(
  filePage,
  /describeConfigTemplates[\s\S]*renderTree\(datas, templates\)/,
  '统一配置清单必须加载并合并历史配置模板，不能只展示配置文件。'
);
assertMatch(
  filePage,
  /templates\.forEach[\s\S]*root\.push[\s\S]*resourceKind:\s*'template'/,
  '配置模板必须作为同级节点直接合入配置清单。'
);
assertNotMatch(
  filePage,
  /template-directory|__config_templates__|配置模板 · \{templates\.length\}/,
  '配置模板不得额外增加目录层级，只能通过节点尾部类型标签区分。'
);
assertMatch(
  filePage,
  /全局模板[\s\S]*resourceKind:\s*'template'[\s\S]*templateId/,
  '历史模板节点必须标明全局作用域。'
);
assertMatch(
  filePage,
  /resourceKind === 'template'[\s\S]*activeNode:[\s\S]*mode:\s*'view'/,
  '点击历史模板节点必须在当前页面选中模板，不能触发独立页面跳转。'
);
assertMatch(
  filePage,
  /params\.delete\('file'\)[\s\S]*params\.set\('templateId'[\s\S]*window\.location\.pathname[\s\S]*replace:\s*true/,
  '模板选中态必须写回当前页面查询参数，使刷新后仍能恢复右侧详情。'
);
assertMatch(
  filePage,
  /<TemplateWorkspace[\s\S]*embedded[\s\S]*templateId=\{selectedTemplateId\}[\s\S]*namespace=\{namespace \|\| ''\}[\s\S]*group=\{group \|\| ''\}/,
  '统一配置清单右侧画布必须嵌入选中模板的原工作区。'
);
assertNotMatch(
  filePage,
  /resourceKind === 'template'[\s\S]{0,800}navigate\(`\/configuration\/group\/templates/,
  '点击模板节点不得离开当前配置分组路由。'
);
assertMatch(
  templatePage,
  /interface TemplateWorkspaceProps[\s\S]*embedded\?: boolean[\s\S]*templateId\?: string[\s\S]*embeddedWorkspace/,
  '模板工作区必须支持隐藏目录与页头的嵌入模式。'
);

for (const [pattern, message] of [
  [/pole-mustache-v1/, '模板页面必须明确展示固定引擎版本。'],
  [
    /value="versions"[\s\S]*value="basic" label="基本信息"/,
    '基本信息必须位于版本管理之后。',
  ],
  [
    /value="content" label="模板内容"[\s\S]*value="schema"[\s\S]*value="values" label="环境 Value"[\s\S]*value="versions"[\s\S]*value="basic" label="基本信息"/,
    '模板工作区必须按内容、Schema、环境 Value、版本管理、基本信息组织一级任务。',
  ],
  [/value="basic"[\s\S]*模板名称[\s\S]*目标格式[\s\S]*模板说明[\s\S]*模板标签/, '模板元信息必须完整归入基本信息页签。'],
  [/setEditing\(true\)[\s\S]*setActiveTab\('basic'\)[\s\S]*新建配置模板/, '新建模板必须直接进入基本信息页签。'],
  [
    /header="发布环境配置版本"[\s\S]*模板草稿[\s\S]*Value 草稿[\s\S]*全量发布[\s\S]*灰度发布[\s\S]*GrayRuleEditor[\s\S]*发布说明/,
    '环境发布必须在同一确认弹窗审阅模板与 Value，并配置发布类型、说明和灰度规则。',
  ],
  [
    /valueActions[\s\S]*保存 Value 草稿[\s\S]*预览渲染[\s\S]*发布环境配置/,
    '环境 Value 编辑区必须提供唯一的组合发布入口。',
  ],
  [/value="schema"[\s\S]*label=\{`参数 Schema/, '参数 Schema 必须作为独立任务页签，不能堆叠在模板正文下方。'],
  [/Namespace Value/, '模板页面必须按 Namespace 维护 Value。'],
  [/GrayRuleEditor/, 'Value 灰度发布必须复用现有灰度规则编辑器。'],
  [/模板标签[\s\S]*TagInput/, '模板定义必须提供独立的标签管理入口。'],
  [/模板说明[\s\S]*<Input/, '单值模板说明必须与名称、目标格式使用一致高度的输入控件。'],
  [
    /previewEnvironmentReleaseIds[\s\S]*valueReleases\.find[\s\S]*templateReleaseId[\s\S]*previewConfigTemplate/,
    '版本预览必须以真实环境组合版本或当前草稿组合为来源。',
  ],
  [
    /startSchemaEditing[\s\S]*setActiveTab\('schema'\)[\s\S]*setEditing\(true\)/,
    'Schema 任务入口必须自行进入参数页签和编辑态。',
  ],
  [/onAddFirst=\{\(\) => startSchemaEditing\(true\)\}/, 'Schema 空态必须支持一键进入编辑并创建首个参数。'],
  [/onCreateSchema=\{\(\) => startSchemaEditing\(true\)\}/, 'Value 空态必须支持直接返回并创建 Schema 参数。'],
]) {
  assertMatch(templatePage, pattern, message);
}

assertNotMatch(templatePage, /workflowGuide|模板配置流程/, '一级页签存在时不得再重复展示流程步骤条。');
assertNotMatch(templatePage, />\s*参考预览\s*</, '草稿参考预览不得继续占用一级操作或页签。');
assertNotMatch(templatePage, /<section className=\{styles\.releaseComposer\}/, 'Value 发布表单不得永久占用编辑页。');
assertNotMatch(
  templatePage,
  /<\/header>[\s\S]{0,500}<section className=\{styles\.metadataLayer\}/,
  '模板元信息不得继续固定显示在摘要头与页签之间。'
);

for (const [source, pattern, message] of [
  [schemaEditor, /还没有参数 Schema[\s\S]*添加第一个参数/, 'Schema 空态必须提供显眼的首参数创建入口。'],
  [schemaEditor, /value\.length > 0[\s\S]*编辑参数/, 'Schema 查看态必须提供就地编辑入口。'],
  [schemaEditor, /key=\{`schema-row-\$\{index\}`\}/, 'Schema 参数行必须使用不随可编辑字段变化的稳定 key。'],
  [valueEditor, /先创建参数 Schema[\s\S]*添加 Schema 参数/, 'Value 缺少 Schema 时必须提供可执行的前置任务入口。'],
  [valueEditor, /validateTemplateValue[\s\S]*canonicalIntegerPattern[\s\S]*canonicalDecimalPattern/, 'Value 编辑器必须按服务端规范校验整数和小数。'],
  [valueEditor, /status=\{error \? 'error' : 'default'\}[\s\S]*aria-invalid=\{Boolean\(error\)\}[\s\S]*role="alert"/, '非法 Value 必须提供输入错误状态和行内提示。'],
  [
    versionManager,
    /versionActions[\s\S]*disabled=\{!namespace\}[\s\S]*渲染预览/,
    '版本管理只保留一个预览按钮，当前环境存在时草稿组合无需先发布即可预览。',
  ],
  [versionManager, /ENVIRONMENT_DRAFT_ROW_ID[\s\S]*当前草稿组合/, '环境版本表必须提供明确的当前草稿组合行。'],
  [versionManager, /selectionId[\s\S]*onReleaseChange[\s\S]*onRowClick/, '草稿与历史环境版本必须共用整行选择模型。'],
  [versionManager, /CodeDiffEditor[\s\S]*comparison\.before[\s\S]*comparison\.after/, '选中两个环境版本后必须提供格式化内容对比。'],
  [versionManager, /versionSelectMarkActive/, '选中的历史行必须具有明确的单选标记。'],
  [versionManager, /data-version-selected/, '选中的历史行必须暴露整行高亮状态。'],
  [
    versionManager,
    /环境配置版本[\s\S]*模板快照与 Value 快照原子绑定[\s\S]*环境空间：/,
    '版本管理必须使用一张表展示当前环境真实存在的组合版本。',
  ],
  [renderPreviewPanel, /renderedSha256[\s\S]*diagnostics/, '统一预览面板必须展示渲染哈希和 diagnostics。'],
]) {
  assertMatch(source, pattern, message);
}

assertMatch(templatePage, /validateTemplateValues\(draft\.parameterSchema, values\)[\s\S]*disabled=\{!namespace \|\| !draft\.id \|\| hasValueErrors\}[\s\S]*disabled=\{!namespace \|\| !draft\.id \|\| editing \|\| hasValueErrors\}/, '模板或 Value 草稿未就绪时必须阻止组合发布。');
assertMatch(
  templatePage,
  /runValuePreview[\s\S]*previewConfigTemplate\(draft, values\)[\s\S]*disabled=\{!namespace \|\| !draft\.id \|\| hasValueErrors\}[\s\S]*预览渲染/,
  '环境 Value 必须使用当前模板草稿与当前未保存 Value 就地预览。'
);
assertMatch(
  templatePage,
  /valueDraftPreview[\s\S]*versionPreview/,
  '环境 Value 草稿预览与历史版本组合预览必须使用独立状态。'
);
assertMatch(
  templatePage,
  /previewEnvironmentReleaseIds\.includes\(ENVIRONMENT_DRAFT_ROW_ID\)[\s\S]*setVersionPreview\(undefined\)[\s\S]*draft\.content[\s\S]*draft\.parameterSchema[\s\S]*values/,
  '选择当前草稿组合时，任一草稿变化必须立即清除旧预览。'
);

assertNotMatch(templatePage, /发布模板版本|发布 Value 版本|固定模板版本/, 'Console 不得继续暴露两个独立发布动作或要求手工固定模板版本。');

assertNotMatch(schemaEditor, /key=\{`\$\{item\.name\}-\$\{index\}`\}/, 'Schema 参数名不能参与行 key，否则无法连续输入。');

assertNotMatch(versionManager, /<Select\b/, '版本历史表已经承担选择入口，不得再重复展示版本下拉框。');
assertNotMatch(versionManager, /版本组合预览|versionSelectionSummary|等待选择/, '版本表已有选中样式时不得重复展示组合摘要。');

for (const [pattern, message] of [
  [/\.page\s*\{[\s\S]*height:\s*calc\(100dvh - 113px\)/, '模板工作区必须占满应用剩余视区。'],
  [/\.workspace\s*\{[\s\S]*margin-top:\s*14px/, '分组导航与模板工作区之间必须保留明确间距。'],
  [
    /\.catalogTable\s*\{[\s\S]*flex:\s*1 1 auto[\s\S]*\.fluent-table-scroll/,
    '模板目录表格必须填满左侧目录的剩余高度。',
  ],
  [
    /\.contentPane\s*\{[\s\S]*height:\s*100%[\s\S]*min-height:\s*0/,
    '模板内容页必须与左侧目录使用同一底部基线。',
  ],
  [
    /\.templateEditorSection\s*\{[\s\S]*flex:\s*1 1 auto[\s\S]*flex-direction:\s*column/,
    '模板编辑卡片必须填满模板内容页的剩余高度。',
  ],
  [
    /\.metadataEditGrid\s*\{[\s\S]*\.fui-Combobox[\s\S]*min-width:\s*0[\s\S]*width:\s*100%/,
    '模板元数据中的目标格式控件必须收缩在自己的网格列内。',
  ],
  [/\.basicPane\s*\{[\s\S]*overflow:\s*auto[\s\S]*padding:/, '基本信息页签必须提供独立可滚动内容区。'],
  [
    /\.versionHistoryGrid\s*\{[\s\S]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    '版本管理必须使用单列环境组合版本表。',
  ],
  [
    /@media \(max-width: 1500px\)[\s\S]*\.versionHistoryGrid,[\s\S]*\.versionPreview[\s\S]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    '中窄宽度下版本历史和预览必须切换为单列。',
  ],
]) {
  assertMatch(templateStyles, pattern, message);
}

for (const [source, pattern, message] of [
  [
    templatePage,
    /templateIdentity[\s\S]*titleLine[\s\S]*resourcePath[\s\S]*templateActions/,
    '模板详情必须复用配置文件的摘要头层级：身份、资源路径与操作区。',
  ],
  [
    templatePage,
    /editing\s*\?[\s\S]*metadataEditGrid[\s\S]*metadataInfoGrid[\s\S]*metadataField/,
    '模板信息查看态必须使用配置文件式字段网格，不能继续展示一排禁用输入框。',
  ],
  [
    templatePage,
    /templateEditorBody[\s\S]*templateEditorStatusBar[\s\S]*UTF-8[\s\S]*pole-mustache-v1/,
    '模板内容画布必须与配置文件编辑器一样提供正文区和底部状态栏。',
  ],
  [
    templateStyles,
    /\.detailHeader\s*\{[\s\S]*grid-template-columns:\s*minmax\(0, 1fr\) auto[\s\S]*padding:\s*18px 20px/,
    '模板摘要头的栅格和间距必须与配置文件详情一致。',
  ],
  [
    templateStyles,
    /\.metadataInfoGrid,[\s\S]*\.metadataEditGrid\s*\{[\s\S]*grid-template-columns:\s*repeat\(2, minmax\(0, 1fr\)\)/,
    '模板信息必须与配置文件基本信息一样使用两列字段网格。',
  ],
  [
    templateStyles,
    /\.templateEditorSection\s*\{[\s\S]*border:\s*0[\s\S]*border-radius:\s*0/,
    '嵌入式模板内容不得继续使用独立圆角卡片视觉。',
  ],
  [
    templateStyles,
    /@container config-template-detail \(max-width: 980px\)[\s\S]*\.detailHeader[\s\S]*grid-template-columns:\s*1fr[\s\S]*\.templateActions[\s\S]*justify-content:\s*flex-start/,
    '模板摘要头必须和配置文件详情一样按右侧容器宽度自动换行。',
  ],
]) {
  assertMatch(source, pattern, message);
}

for (const [pattern, message] of [
  [/configType/, 'ConfigFile 详情必须识别普通文本和模板类型。'],
  [/bindConfigFileTemplate/, 'ConfigFile 详情必须通过显式 binding 接口切换模板版本。'],
  [/templateReleaseId/, 'ConfigFile 详情必须保存明确的 Template Release。'],
  [/直接文本[\s\S]*配置模板|配置模板[\s\S]*直接文本/, 'ConfigFile 编辑态必须显式提供两种内容来源。'],
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
