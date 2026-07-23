import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const files = {
  servicesIndex: 'src/pages/Discovery/Services/index.tsx',
  servicesTable: 'src/pages/Discovery/Services/services.tsx',
  serviceEditor: 'src/pages/Discovery/Services/ServiceEditor.tsx',
  serviceForm: 'src/pages/Discovery/Services/ServiceForm.tsx',
  instanceService: 'src/services/instance.ts',
  labelInput: 'src/components/LabelInput/index.tsx',
  aliasTable: 'src/pages/Discovery/Services/alias.tsx',
  aliasEditor: 'src/pages/Discovery/Services/AliasEditor.tsx',
  serviceInstance: 'src/pages/Discovery/Services/Instance/index.tsx',
  instanceTable: 'src/pages/Discovery/Services/Instance/InstanceTable.tsx',
  instanceEditor: 'src/pages/Discovery/Services/Instance/InstanceEditor.tsx',
  serviceDetail: 'src/pages/Discovery/Services/Instance/ServiceDetail.tsx',
  serviceDetailStyle: 'src/pages/Discovery/Services/Instance/index.module.less',
  appLayoutStyle: 'src/layouts/components/AppLayout.module.less',
  serviceStyle: 'src/pages/Discovery/Services/index.module.less',
  namespaceIndex: 'src/pages/Namespace/index.tsx',
  namespaceStyle: 'src/pages/Namespace/index.module.less',
};

const sources = Object.fromEntries(
  Object.entries(files).map(([key, rel]) => {
    const file = path.join(root, rel);
    if (!fs.existsSync(file)) {
      throw new Error(`missing file: ${file}`);
    }
    return [key, fs.readFileSync(file, 'utf8')];
  }),
);

const index = sources.servicesIndex;
const services = sources.servicesTable;
const serviceEditor = sources.serviceEditor;
const serviceForm = sources.serviceForm;
const instanceService = sources.instanceService;
const labelInput = sources.labelInput;
const alias = sources.aliasTable;
const aliasEditor = sources.aliasEditor;
const serviceInstance = sources.serviceInstance;
const instanceTable = sources.instanceTable;
const instanceEditor = sources.instanceEditor;
const serviceDetail = sources.serviceDetail;
const serviceDetailStyle = sources.serviceDetailStyle;
const appLayoutStyle = sources.appLayoutStyle;
const style = sources.serviceStyle;
const namespaceIndex = sources.namespaceIndex;
const namespaceStyle = sources.namespaceStyle;

for (const anchor of [
  '注册发现 / 服务实例',
  '注册发现',
  'serviceTableRef',
  '新建服务',
]) {
  assert.match(index, new RegExp(anchor), `注册发现页必须保留并接通 ${anchor}`);
}

assert.match(index, /<ResourceHeader[\s\S]*eyebrow="注册发现 \/ 服务实例"[\s\S]*actions=\{\(/,
  '注册发现页头必须使用统一 ResourceHeader，并在右侧提供操作区');
assert.match(index, /<Button shape="square" variant="outline" onClick=\{refreshServices\}>/,
  '注册发现页头必须提供服务列表刷新按钮');
assert.match(index, /<Button theme="primary" icon=\{<AddIcon \/>\} onClick=\{createService\}>新建服务<\/Button>/,
  '注册发现页头必须只提供新建服务按钮');
assert.match(index, /<section className=\{style\.serviceWorkspace\}>[\s\S]*<ServicesTable ref=\{serviceTableRef\} \/>/,
  '注册发现首页必须直接展示服务列表工作台');
assert.doesNotMatch(index, /Service Registry \/ Discovery/,
  '注册发现页头不能继续使用英文 eyebrow');
assert.doesNotMatch(index, /ServiceAliasTable|aliasTableRef|activeTab|TabPanel|<Tabs|新建别名/,
  '注册发现首页不能再把别名作为顶层页签或页头动作');

assert.doesNotMatch(style, /\.registryTabs|\.tabContent/,
  '服务页样式不应再保留顶层服务/别名 Tabs 容器样式');
assert.match(style, /\.page[\s\S]*height: calc\(100vh - 113px\)[\s\S]*overflow: hidden/,
  '注册发现服务页必须固定在当前视区内，避免整个页面跟随表格行滚动');
assert.match(style, /\.serviceWorkspace[\s\S]*margin-top: 24px/,
  '服务列表应作为注册发现首页唯一工作区直接承接页头');
assert.match(style, /\.serviceForm[\s\S]*:global\(\.fui-Input\),[\s\S]*:global\(\.fui-Combobox\)[\s\S]*width: min\(640px, 100%\)/,
  '服务详情页内编辑态输入框必须占满字段区域，不能使用 Fluent 默认窄宽度');
assert.match(style, /\.serviceWorkspace[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0/,
  '服务列表工作区必须建立可收缩高度链，才能把滚动交给表格内部');
assert.match(style, /\.workspace[\s\S]*gap: 28px/,
  '服务指标条和列表区必须用父级 gap 明确分开');
assert.match(style, /\.workspace[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0/,
  '服务列表根工作区必须继承可用高度并允许内部表格收缩滚动');
assert.match(style, /\.embeddedWorkspace[\s\S]*display: grid[\s\S]*gap: 14px/,
  '服务详情内的别名列表必须使用嵌入式工作区，避免继承服务首页指标区间距');
assert.match(style, /\.embeddedWorkspace[\s\S]*margin: 20px/,
  '服务详情内的别名列表必须和 Tabs 内容区四周保留稳定间距');
assert.match(style, /\.aliasDetailSection \{\n  display: grid;\n  gap: 12px;\n  min-width: 0;\n\}/,
  '服务详情内别名清单只负责紧凑分组，不应额外形成外层卡片');
assert.doesNotMatch(style, /\.aliasDetailSection \{(?:(?!\n\}).)*(padding|border|border-radius|background):/s,
  '服务详情内别名清单不能再用外层卡片样式解决留白');
assert.match(style, /\.aliasDetailToolbar[\s\S]*justify-content: space-between[\s\S]*padding: 0 0 2px/,
  '服务详情内别名清单工具栏必须贴合详情页密度，而不是外层列表页工具栏');
assert.match(style, /\.aliasDetailActions[\s\S]*display: flex[\s\S]*flex-wrap: wrap[\s\S]*justify-content: end/,
  '服务详情内别名清单操作区必须稳定排列新建、搜索、查询和重置，并允许控件整体换行');
assert.match(style, /\.aliasDetailTableSurface[\s\S]*:global\(\.fluent-table-shell\)[\s\S]*min-width: 720px/,
  '服务详情内别名表格必须降低最小宽度，避免内嵌宽表挤压');
assert.match(style, /\.aliasSummaryBar[\s\S]*display: grid[\s\S]*grid-template-columns: minmax\(220px, 1\.4fr\) repeat\(2, minmax\(120px, 1fr\)\)[\s\S]*border-bottom: 1px solid var\(--app-border-subtle\)/,
  '服务详情内别名总栏必须放在表格面板顶部，并保持一行轻量统计');
assert.match(style, /\.aliasSummaryItem[\s\S]*min-height: 64px[\s\S]*border-right: 1px solid var\(--app-border-subtle\)/,
  '别名总栏必须是轻量横向统计，不应升级为外部统计卡阵列');
assert.match(style, /\.aliasSummaryValue[\s\S]*font-family: ui-monospace[\s\S]*text-overflow: ellipsis/,
  '别名总栏目标服务等值必须使用 mono 且避免长文本撑破布局');
assert.match(style, /\.aliasDetailTableSurface[\s\S]*:global\(\.fluent-table-state\)[\s\S]*box-sizing: border-box[\s\S]*height: 92px[\s\S]*min-height: 92px[\s\S]*padding: 18px 0/,
  '服务详情内别名表格空态必须收敛高度，不能撑出大面积空白');
assert.match(style, /\.aliasDetailTableSurface[\s\S]*:global\(\.fluent-pagination\)[\s\S]*padding: 12px 16px/,
  '服务详情内别名表格分页必须使用紧凑安全区');
assert.match(style, /\.aliasReadonlyValue[\s\S]*background: var\(--app-surface-subtle\)[\s\S]*font-family: ui-monospace/,
  '服务别名抽屉内目标服务必须作为只读信息展示，不应伪装成可切换输入');
assert.match(style, /\.aliasDrawer[\s\S]*:global\(\.fui-DrawerBody\)[\s\S]*overflow-y: auto/,
  '服务别名抽屉内部应自行滚动，避免把滚动交给页面');
assert.match(serviceEditor, /header="创建服务"[\s\S]*<ServiceForm[\s\S]*mode="create"/,
  'ServiceEditor 只能作为创建服务弹窗容器，不能再承载已有服务查看或编辑');
assert.doesNotMatch(serviceEditor, /服务详情|编辑服务|setEditable|footer=\{op === 'view'/,
  '已有服务查看/编辑不能继续塞进 ServiceEditor Drawer');
assert.match(services, /const openServiceDetail = \(row\?: TableRowData, edit = false\)[\s\S]*params\.set\('mode', 'edit'\)[\s\S]*navigate\(`instance\?\$\{params\.toString\(\)\}`\)/,
  '服务列表必须统一跳转到详情页，并用 mode=edit 表达详情页内编辑态');
assert.match(services, /case 'view':[\s\S]*openServiceDetail\(row, true\);/,
  '服务列表操作列的查看编辑入口必须进入详情页编辑态');
assert.doesNotMatch(services, /case 'view':[\s\S]*setEditorState[\s\S]*mode: op/s,
  '服务列表查看入口不能再打开 ServiceEditor 抽屉');
assert.match(serviceInstance, /const editMode = urlParams\.get\('mode'\) === 'edit'/,
  '服务详情页路由必须识别 mode=edit 参数');
assert.match(serviceInstance, /initialEdit=\{editMode\}/,
  '服务详情组件必须接收路由编辑态');
assert.match(serviceDetail, /initialEdit = false[\s\S]*const \[editing, setEditing\] = React\.useState\(initialEdit\)/,
  '服务详情页必须在页面内维护查看和编辑状态');
assert.match(serviceDetail, /<ServiceForm[\s\S]*mode=\{editing \? 'edit' : 'view'\}[\s\S]*onSubmitted=\{\(\) => \{[\s\S]*setEditing\(false\);[\s\S]*reloadService\(\);/,
  '服务详情页必须复用 ServiceForm，并在提交后回到查看态刷新详情');
assert.match(serviceForm, /const identityEditable = mode === 'create'/,
  '服务表单中命名空间和名称只能在创建服务时编辑');
assert.match(serviceForm, /mode === 'create'[\s\S]*dispatch\(saveServices[\s\S]*dispatch\(updateServices/,
  '服务表单必须同时服务创建弹窗和详情页内编辑提交');
assert.match(style, /\.metricRail[\s\S]*margin: 0;/,
  '服务指标条应作为独立区块，不再用自身 margin 粘连列表节奏');
assert.match(style, /\.metricRail[\s\S]*flex: 0 0 auto/,
  '服务指标条必须固定在表格滚动区之外');
assert.match(style, /\.metricValue[\s\S]*display: flex[\s\S]*align-items: baseline[\s\S]*gap: 3px/,
  '服务指标条健康实例数值必须保持单行基线对齐，避免被 span block 样式拆成竖排');
assert.match(style, /\.listSection[\s\S]*display: flex[\s\S]*flex: 1 1 auto[\s\S]*flex-direction: column[\s\S]*gap: 14px[\s\S]*min-height: 0/,
  '服务列表标题筛选区和表格必须归入可收缩的单独列表区块');
assert.match(style, /\.filterBar[\s\S]*flex: 0 0 auto/,
  '服务列表筛选栏必须固定在表格滚动区之外');
assert.match(style, /\.filterInput[\s\S]*width: 220px/,
  '服务和别名筛选输入框宽度应与命名空间页一致');
assert.match(style, /\.namespaceFilter[\s\S]*width: 180px/,
  '服务列表命名空间筛选应保持稳定宽度');
assert.match(style, /\.tableSurface[\s\S]*min-width: 0[\s\S]*overflow: hidden/,
  '服务表格外壳不得被最小宽度撑破，横向滚动必须交给内部 fluent-table-scroll');
assert.doesNotMatch(style, /:global\(\.fluent-table-shell\)\s*\{[^}]*min-width:\s*1040px/,
  '服务表格最小宽度不得施加在内部滚动容器之外');
assert.match(style, /@media \(max-width: 980px\)[\s\S]*\.serviceTableSurface[\s\S]*min-height: 420px/,
  '中窄屏服务列表必须切换为页面滚动并保留可操作的表体高度');
assert.match(style, /@media \(max-width: 980px\)[\s\S]*\.serviceWorkspace,[\s\S]*\.workspace,[\s\S]*\.listSection[\s\S]*width: 100%[\s\S]*max-width: 100%/,
  '解除固定高度后服务工作区仍必须锁定容器宽度，不能按宽表最大内容无限扩张');
assert.match(style, /\.serviceTableSurface[\s\S]*display: flex[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0[\s\S]*overflow: hidden/,
  '服务清单表格外层必须吃满剩余高度并阻止外层页面滚动');
assert.match(style, /\.serviceTableSurface[\s\S]*:global\(\.fluent-table-shell\)[\s\S]*display: flex[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0/,
  '服务清单 TDesign 表格必须参与内部 flex 高度分配');
assert.match(style, /\.serviceTableSurface[\s\S]*:global\(\.fluent-table-scroll\)[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0[\s\S]*overflow: auto[\s\S]*overscroll-behavior: contain/,
  '服务清单必须把纵向滚动挂到 fluent-table-scroll，并阻止滚动链传到页面');
assert.match(style, /\.tableSurface[\s\S]*:global\(\.fluent-table-shell th\)[\s\S]*padding: 14px/,
  '服务表头密度应与命名空间页一致');
assert.match(style, /\.healthTrack\b/,
  '服务健康列应采用与命名空间页同类的健康进度条');
assert.match(style, /\.actionCell\b/,
  '服务操作列应采用与命名空间页同类的紧凑操作容器');
assert.match(style, /\.serviceForm[\s\S]*height: 100%/,
  '创建服务抽屉表单必须占满抽屉正文并交给抽屉纵向滚动');
assert.match(style, /\.serviceForm[\s\S]*:global\(\.fluent-form-item\)[\s\S]*margin-bottom: 0/,
  '创建服务抽屉字段间距必须挂到字段 wrapper，不能依赖 FormItem 内部 margin');
assert.match(style, /\.serviceDrawerContent[\s\S]*gap: 22px/,
  '创建服务抽屉分段之间必须保留稳定间距');
assert.match(style, /\.serviceFormSection[\s\S]*> div\[id\^="in"\][\s\S]*margin-bottom: 18px[\s\S]*> div\[id\^="in"\]:last-child[\s\S]*margin-bottom: 0/,
  '创建服务抽屉连续输入字段必须通过锚点 wrapper 保持纵向间距');
assert.match(style, /\.serviceFormSectionTitle[\s\S]*&::before[\s\S]*background: #0052d9[\s\S]*&::after[\s\S]*background: var\(--app-border-subtle\)/,
  '创建服务抽屉分段标题必须使用竖色条和发丝线');
assert.match(style, /\.nameMeta[\s\S]*margin-left: 104px[\s\S]*justify-content: space-between/,
  '服务名称计数和不可修改提示必须对齐输入控件列');
assert.match(style, /\.readonlyField[\s\S]*min-height: 32px[\s\S]*color: var\(--app-text\)[\s\S]*line-height: 32px/,
  '服务查看态字段必须是无边框正常文本，不应显示 disabled 输入框背景和边框');
assert.doesNotMatch(style, /\.serviceView|\.viewSection|\.viewGrid|\.viewField|\.viewLabel|\.viewValue|\.labelList/,
  '服务查看态不能再维护独立展示样式，必须复用服务编辑表单布局');
assert.doesNotMatch(style, /\.serviceLabelEditor|\.labelEditorHeader|\.labelEditorRowError|\.tagsEmpty|\.labelEditorFooter/,
  '创建服务抽屉不能再维护私有标签编辑器样式，必须复用统一 LabelInput');

assert.match(services, /export interface ServicesTableHandle/,
  '服务列表必须向页头暴露刷新和新建动作');
assert.match(services, /React\.forwardRef<ServicesTableHandle/,
  '服务列表必须使用 forwardRef 接通页头操作');
assert.match(services, /React\.useImperativeHandle\(ref,[\s\S]*refreshTable\(page, limit, query\)[\s\S]*operateService\('create'\)/,
  '服务列表页头动作必须复用当前筛选和创建逻辑');
assert.match(services, /dispatch\(listAllNamespaces\(\)\)/,
  '服务列表必须加载命名空间选项用于筛选');
assert.match(services, /namespace: nextQuery\.namespace \|\| undefined/,
  '服务列表查询必须带命名空间筛选参数');
assert.match(services, /<strong id="stSvc">[\s\S]*<strong id="stNs">[\s\S]*id="stHealthy"[\s\S]*id="stInst"/,
  '服务统计必须提供 stSvc/stNs/stHealthy/stInst 锚点');
assert.match(services, /<span>服务数<\/span>[\s\S]*<span>命名空间<\/span>[\s\S]*<span>健康实例<\/span>/,
  '服务指标条标题必须使用中文文案');
assert.doesNotMatch(services, />Services<|>Namespaces<|>Healthy Instances</,
  '服务指标条不能继续混用英文标题');
assert.match(services, /<strong className=\{style\.metricValue\}>[\s\S]*<b id="stHealthy">\{metric\.healthy\}<\/b>[\s\S]*\/[\s\S]*<b id="stInst">\{metric\.instances\}<\/b>/,
  '健康实例统计必须保留 stHealthy/stInst 锚点，并用非 span 单行结构展示');
assert.match(services, /<span id="listCount">/,
  '服务清单显示条数必须提供 listCount 锚点');
assert.match(services, /<div id="nsFilter"[\s\S]*placeholder="全部命名空间"/,
  '服务列表必须提供命名空间筛选锚点和全部命名空间 placeholder');
assert.match(services, /<div id="keyword"[\s\S]*placeholder="服务名"/,
  '服务列表关键词筛选必须提供 keyword 锚点');
assert.match(services, /<section id="tbody" className=\{`\$\{style\.tableSurface\} \$\{style\.serviceTableSurface\}`\}>/,
  '服务表格主体必须提供 tbody 锚点并保持内部滚动 surface');
assert.match(services, /<section className=\{style\.metricRail\}[\s\S]*<section className=\{style\.listSection\}>[\s\S]*<ResourceToolbar[\s\S]*<section id="tbody" className=\{`\$\{style\.tableSurface\} \$\{style\.serviceTableSurface\}`\}>/,
  '服务指标条必须和服务列表表格分成两个相邻区块');
assert.doesNotMatch(services, /<Tooltip content=\{t\('common\.refresh'\)\}>/,
  '服务列表工具栏不应再重复放刷新按钮');
assert.doesNotMatch(services, /onClick=\{\(\) => operateService\('create'\)\}>\{t\('common\.add'\)\}/,
  '服务列表工具栏不应再重复放新建按钮');
assert.match(services, /placeholder="服务名"/,
  '服务列表筛选输入应使用命名空间页同类短 placeholder');
assert.match(services, /<Button variant="outline" onClick=\{submitFilter\}>查询<\/Button>/,
  '服务列表筛选区必须提供查询按钮');
assert.match(services, /<Button variant="text" onClick=\{resetFilter\}>重置<\/Button>/,
  '服务列表筛选区必须提供重置按钮');
assert.match(services, /size=\{"large"\}/,
  '服务表格密度必须对齐命名空间页');
assert.match(services, /tableLayout=(\{"fixed"\}|"fixed")/,
  '服务表格布局必须对齐命名空间页');

assert.match(serviceForm, /import LabelInput from 'components\/LabelInput'/,
  '创建服务抽屉必须复用统一 LabelInput 标签控件');
assert.match(serviceForm, /const SERVICE_NAME_REG = \/\^\[0-9A-Za-z\._-\]\+\$\/;/,
  '创建服务名称必须显式约束数字、英文字母、.、-、_');
assert.match(serviceForm, /describeAllServices\(\)[\s\S]*then\(setAllServices\)/,
  '创建服务抽屉必须加载全量服务用于同命名空间重名校验，且不污染服务列表 store');
assert.match(serviceForm, /validateServiceName[\s\S]*该命名空间下服务名已存在/,
  '创建服务提交前必须校验同命名空间服务名不可重复');
assert.match(serviceForm, /validateLabels[\s\S]*标签键不能为空[\s\S]*标签键 \$\{key\} 重复/,
  '服务标签提交前必须校验键为空和值重复');
assert.match(serviceForm, /id="drawer"[\s\S]*基础信息[\s\S]*id="inNs"[\s\S]*id="inName"[\s\S]*id="inDesc"[\s\S]*归属信息[\s\S]*id="inDept"[\s\S]*id="inBiz"[\s\S]*服务标签/,
  '创建服务抽屉必须按基础信息、归属信息、服务标签三段组织并提供字段锚点');
assert.match(serviceForm, /id="nameCount"[\s\S]*id="errName"/,
  '服务名称字段必须提供实时计数和内联错误锚点');
assert.match(serviceForm, /rules=\{identityEditable \? \[\{ required: true, message: '请选择命名空间' \}\] : \[\]\}/,
  '服务查看态和编辑态命名空间 FormItem 不应保留 required 规则和红色必填星号');
assert.match(serviceForm, /rules=\{identityEditable \? \[[\s\S]*请输入服务名称[\s\S]*\] : \[\]\}/,
  '服务查看态和编辑态名称 FormItem 不应保留 required 规则和红色必填星号');
assert.match(serviceForm, /\{identityEditable && \([\s\S]*id="nameCount"[\s\S]*id="errName"[\s\S]*\)\}/,
  '服务名称的编辑提示、字数统计和内联错误只应在创建态显示');
assert.match(serviceForm, /<LabelInput[\s\S]*name="service_labels"[\s\S]*hideLabel[\s\S]*editorId="tagRows"[\s\S]*emptyId="tagsEmpty"[\s\S]*countId="tagCount"/,
  '服务标签编辑器必须通过统一 LabelInput 提供行容器、空态和计数锚点');
assert.match(serviceForm, /openInfoNotification\('请求成功', mode === 'create' \? '服务已创建' : '修改服务成功'\)/,
  '创建服务成功提示必须使用服务已创建文案');
assert.doesNotMatch(serviceForm, /const serviceView|style\.serviceView|style\.viewSection|editable \? serviceForm : serviceView/,
  '服务查看态不能再使用独立 serviceView 分支，必须渲染同一套 serviceForm');
assert.match(serviceForm, /const ReadonlyField[\s\S]*className=\{style\.readonlyField\}/,
  '服务查看态字段必须使用无边框只读文本组件');
assert.match(serviceForm, /identityEditable \? \([\s\S]*<Select[\s\S]*\) : <ReadonlyField value=\{watchedNamespace\} \/>[\s\S]*identityEditable \? \([\s\S]*<Input[\s\S]*\) : <ReadonlyField value=\{watchedName\} \/>/s,
  '命名空间和名称查看态、编辑态必须渲染只读文本，而不是 disabled 控件');
assert.match(serviceForm, /editable \? <Input placeholder="请输入服务描述" \/> : <ReadonlyField value=\{watchedComment\} \/>[\s\S]*editable \? <Input placeholder="请输入部门" \/> : <ReadonlyField value=\{watchedDepartment\} \/>[\s\S]*editable \? <Input placeholder="请输入业务" \/> : <ReadonlyField value=\{watchedBusiness\} \/>/s,
  '描述、部门和业务查看态必须渲染只读文本，而不是 disabled 控件');
assert.doesNotMatch(serviceForm, /<(Input|Select)[\s\S]{0,180}disabled=\{!editable\}/,
  '服务查看态字段不能通过 disabled Input/Select 表现，只能通过只读文本表现');
assert.match(serviceForm, /<LabelInput[\s\S]*editable=\{editable\}[\s\S]*disabled=\{!editable\}/,
  '服务标签仍由统一 LabelInput 根据 editable 切换查看/编辑态');
assert.match(serviceDetail, /<Button theme="primary" disabled=\{service\.editable === false\} onClick=\{\(\) => setEditing\(true\)\}>[\s\S]*编辑[\s\S]*<\/Button>/,
  '服务查看态必须在详情页头部提供编辑按钮，点击后原地放开同一表单编辑');
assert.match(serviceForm, /\{editable && \([\s\S]*<FormItem className=\{style\.serviceFormFooter\}>[\s\S]*<Button type="submit" theme="primary">[\s\S]*提交[\s\S]*<Button theme="default" onClick=\{onCancel \|\| resetForm\}>[\s\S]*取消/s,
  '服务表单只有进入编辑态后才显示提交和取消按钮');
assert.match(serviceEditor, /size="min\(720px, 94vw\)"[\s\S]*showOverlay[\s\S]*closeOnOverlayClick[\s\S]*closeOnEscKeydown[\s\S]*destroyOnClose/,
  '服务查看和编辑抽屉必须使用同一宽度 min(720px, 94vw)，并支持遮罩点击、Esc 关闭和销毁');
assert.doesNotMatch(serviceEditor, /showOverlay=\{false\}|取消/,
  '创建服务抽屉不能关闭遮罩，也不应放冗余取消按钮');

assert.match(labelInput, /missingKey[\s\S]*标签键不能为空/,
  '统一 LabelInput 必须校验有标签值但标签键为空的行');
assert.match(labelInput, /id=\{emptyId\} className=\{style\.emptyEditor\}[\s\S]*<div>暂无标签<\/div>/,
  '统一 LabelInput 空态必须只展示暂无标签');
const emptyBlockStart = labelInput.indexOf('id={emptyId} className={style.emptyEditor}');
const emptyBlockEnd = labelInput.indexOf(') : (', emptyBlockStart);
assert.notEqual(emptyBlockStart, -1, '统一 LabelInput 必须提供 emptyId 空态锚点');
assert.notEqual(emptyBlockEnd, -1, '统一 LabelInput 空态块必须保持在 labels.length 三元表达式内');
assert.doesNotMatch(labelInput.slice(emptyBlockStart, emptyBlockEnd), /添加标签/,
  '统一 LabelInput 空态操作列不能再放添加标签按钮');
assert.match(labelInput, /<Popup trigger="hover" content="删除标签">[\s\S]*<DeleteIcon \/>/,
  '统一 LabelInput 行操作列只承载删除标签动作');
assert.match(labelInput, /<div className=\{style\.footer\}>[\s\S]*icon=\{<AddIcon \/>\}[\s\S]*添加标签[\s\S]*<span id=\{countId\}>/,
  '统一 LabelInput 添加标签入口必须放在底部，并保留计数锚点');

assert.match(alias, /export interface ServiceAliasTableHandle/,
  '别名列表必须向页头暴露刷新和新建动作');
assert.match(alias, /React\.forwardRef<ServiceAliasTableHandle/,
  '别名列表必须使用 forwardRef 接通页头操作');
assert.match(alias, /namespace\?: string;[\s\S]*serviceName\?: string;[\s\S]*embedded\?: boolean;/,
  '别名列表必须支持服务详情上下文参数');
assert.match(alias, /const inServiceDetail = Boolean\(namespace && serviceName\)/,
  '别名列表必须识别是否处于服务详情上下文');
assert.match(alias, /if \(!inServiceDetail\)[\s\S]*目标服务命名空间[\s\S]*目标服务名/,
  '服务详情内别名列表不应展示目标服务两列');
assert.match(alias, /namespace: inServiceDetail \? namespace : undefined[\s\S]*service: inServiceDetail \? serviceName : undefined/,
  '服务详情内别名列表查询必须按当前服务过滤');
assert.match(alias, /targetService=\{inServiceDetail \? \{ namespace: namespace as string, serviceName: serviceName as string \} : undefined\}/,
  '服务详情内创建和编辑别名必须固定当前目标服务');
assert.match(alias, /components\/OperationButton/,
  '服务详情内别名行操作必须使用统一 OperationButton 组件');
assert.match(alias, /<OperationButton action="copy" label="复制别名"/,
  '服务详情内别名行操作必须包含复制别名图标语义');
assert.match(alias, /import \{ copyToClipboard \} from 'utils\/sys'/,
  '复制别名必须复用控制台已有剪贴板工具');
assert.match(alias, /const aliasNamespaceCount = useMemo\(\(\) => \{[\s\S]*new Set\(datas\.map\(item => item\.alias_namespace\)\.filter\(Boolean\)\)\.size/,
  '别名总栏必须根据当前表格数据实时计算覆盖命名空间');
assert.match(alias, /const copyAlias = \(row: TableRowData\) => \{[\s\S]*copyToClipboard\(`\$\{row\.alias_namespace \|\| '-'\}\/\$\{row\.alias \|\| '-'\}`\)/,
  '行级复制必须复制 命名空间/别名');
assert.match(alias, /embedded && \([\s\S]*新建别名/,
  '服务详情内别名列表必须提供新建别名入口');
assert.match(alias, /<section className=\{embedded \? style\.aliasDetailSection : style\.listSection\}>[\s\S]*<section className=\{embedded \? style\.aliasDetailToolbar : style\.filterBar\}>/,
  '别名列表在服务详情内必须切换为紧凑详情子清单布局');
assert.match(alias, /<div className=\{embedded \? style\.aliasDetailActions : style\.filterActions\}>/,
  '服务详情内别名操作区必须使用专用紧凑布局');
assert.match(alias, /embedded && namespace && serviceName[\s\S]*`当前显示 \$\{datas\.length\} 条`/,
  '服务详情内别名清单必须显示当前结果数量，服务上下文由摘要区承担');
assert.match(alias, /id=\{embedded \? 'listCount' : undefined\}/,
  '服务详情内别名清单必须提供 listCount 锚点');
assert.match(alias, /id=\{embedded \? 'keyword' : undefined\}/,
  '服务详情内别名筛选输入必须提供 keyword 锚点');
assert.match(alias, /<section className=\{embedded \? `\$\{style\.tableSurface\} \$\{style\.aliasDetailTableSurface\}` : style\.tableSurface\}>/,
  '服务详情内别名表格必须使用专用紧凑 table surface');
assert.match(alias, /<section className=\{style\.aliasSummaryBar\} id="summary">[\s\S]*id="sumTarget"[\s\S]*id="sumCount"[\s\S]*id="sumNs"/,
  '服务详情内别名表格面板顶部必须提供 summary/sumTarget/sumCount/sumNs 锚点');
assert.match(alias, /<div id=\{embedded \? 'tbody' : undefined\}>[\s\S]*<Table/,
  '服务详情内别名主表必须提供 tbody 锚点');
assert.match(alias, /existingAliases=\{datas\}/,
  '服务别名编辑器必须拿到当前页已有别名用于前端去重校验');
assert.doesNotMatch(alias, /<Tooltip content="刷新">/,
  '别名列表工具栏不应再重复放刷新按钮');
assert.match(alias, /placeholder="别名"/,
  '别名列表筛选输入应使用短 placeholder');
assert.match(alias, /<Button variant="outline" onClick=\{submitFilter\}>查询<\/Button>/,
  '别名列表筛选区必须提供查询按钮');
assert.match(alias, /<Button variant="text" onClick=\{resetFilter\}>重置<\/Button>/,
  '别名列表筛选区必须提供重置按钮');

assert.match(aliasEditor, /targetService\?: \{[\s\S]*namespace: string;[\s\S]*serviceName: string;/,
  '别名编辑器必须支持固定目标服务');
assert.match(aliasEditor, /existingAliases\?: ServiceAliasView\[\];/,
  '别名编辑器必须支持基于当前列表做前端去重校验');
assert.match(aliasEditor, /if \(!targetService\)[\s\S]*dispatch\(listAllServices\(\)\)/,
  '固定目标服务时不应再请求和选择全量服务列表');
assert.match(aliasEditor, /alias_namespace: data\?\.alias_namespace \|\| targetService\?\.namespace/,
  '创建服务别名时别名命名空间必须默认当前命名空间');
assert.match(aliasEditor, /selectedSvcNamespace = targetService\?\.namespace \|\| selectedSvc\.split\('@'\)\[0\]/,
  '别名提交必须优先使用当前服务命名空间');
assert.match(aliasEditor, /selectedSvcName = targetService\?\.serviceName \|\| selectedSvc\.split\('@'\)\[1\]/,
  '别名提交必须优先使用当前服务名');
assert.match(aliasEditor, /const ALIAS_NAME_REG = \/\^\[a-z\]\[a-z0-9-\]\*\$\//,
  '服务别名提交前必须校验别名命名规范');
assert.match(aliasEditor, /const duplicated = op === 'create' && existingAliases\.some\(item => item\.alias_namespace === aliasNamespace && item\.alias === alias\)/,
  '服务别名创建态必须拦截当前页已知的同命名空间重名');
assert.match(aliasEditor, /targetService \? \([\s\S]*<div className=\{style\.aliasReadonlyValue\} aria-readonly="true">[\s\S]*\{targetService\.namespace\}\/\{targetService\.serviceName\}/,
  '服务详情内别名编辑器必须把目标服务作为只读信息展示');
assert.match(aliasEditor, /disabled=\{op === 'edit'\}/,
  '编辑服务别名时必须锁定别名所在命名空间');
assert.match(aliasEditor, /<Button theme="default" onClick=\{resetForm\}>[\s\S]*重置/,
  '服务别名抽屉重置必须保留目标服务和初始上下文');
assert.match(aliasEditor, /size='min\(560px, 92vw\)'[\s\S]*showOverlay[\s\S]*closeOnOverlayClick[\s\S]*closeOnEscKeydown[\s\S]*destroyOnClose/,
  '服务别名抽屉必须使用轻量宽度，并支持遮罩点击与 Esc 关闭');
assert.doesNotMatch(aliasEditor, /showOverlay=\{false\}/,
  '服务别名抽屉不能关闭遮罩');

assert.match(serviceInstance, /import ServiceAliasTable from '\.\.\/alias'/,
  '服务详情页必须复用服务别名列表组件');
assert.match(serviceInstance, /<TabPanel value=\{"2"\} label="服务别名">[\s\S]*<ServiceAliasTable[\s\S]*namespace=\{namespace \|\| ''\}[\s\S]*serviceName=\{serviceName \|\| ''\}[\s\S]*embedded/,
  '服务详情页必须把别名作为当前服务下的 Tab');
assert.match(serviceInstance, /<ServiceDetail[\s\S]*onTabChange=\{setActiveTab\}/,
  '服务详情头部操作必须接通父级受控 Tab 切换');
assert.match(serviceInstance, /<section[\s\S]*className=\{style\.serviceGovernance\}[\s\S]*title=\{`\$\{namespace \|\| '-'\}\/\$\{serviceName \|\| '-'\}`\}[\s\S]*<GovernanceWorkbench/,
  '服务详情内嵌治理工作台必须用可收缩容器承载，并为长服务身份提供完整提示');
assert.match(serviceDetailStyle, /\.serviceGovernance[\s\S]*min-width: 0[\s\S]*:global\(\.fui-Radio\)[\s\S]*white-space: nowrap/,
  '服务详情治理区域必须阻止角色单选项逐字换行');
assert.match(serviceDetailStyle, /@media \(max-width: 900px\)[\s\S]*\.serviceGovernance[\s\S]*> div > div:first-child[\s\S]*flex-wrap: wrap/,
  '900px 左右的服务治理上下文栏必须允许整体换行，不能压缩内部文案');
assert.doesNotMatch(serviceDetailStyle, /@media \(max-width: 960px\)[\s\S]*\.instanceStatusStrip,\s*\n\s*\.instanceDetailGrid/,
  '接近桌面宽度的实例抽屉状态摘要仍应保持横向三栏');
assert.match(serviceDetailStyle, /@media \(max-width: 640px\)[\s\S]*\.instanceStatusStrip[\s\S]*grid-template-columns: 1fr/,
  '实例状态摘要只应在真正窄屏下切换为单栏');
assert.match(instanceEditor, /className=\{style\.instanceForm\}/,
  '实例编辑表单必须提供独立样式锚点');
assert.doesNotMatch(instanceEditor, /<FormItem label=\{'命名空间'\} name=\{'namespace'\}|<FormItem label=\{'服务'\} name=\{'service'\}/,
  '创建实例时命名空间和服务来自当前服务上下文，不能显示成只读表单框');
assert.match(instanceEditor, /namespace: editIns\?\.namespace \|\| namespace,[\s\S]*service: editIns\?\.service \|\| service,/,
  '创建或编辑实例请求必须直接使用当前服务上下文或已有实例归属');
assert.match(instanceEditor, /<InstanceSummary[\s\S]*namespace=\{formNamespace\}[\s\S]*service=\{formService\}/,
  '移除固定表单框后仍应在实例身份摘要中展示当前服务上下文');
assert.match(instanceEditor, /className=\{`\$\{style\.instanceFormGrid\} \$\{style\.instanceRuntimeGrid\}`\}[\s\S]*label=\{'主机'\}[\s\S]*label=\{'权重'\}/,
  '实例运行信息必须使用独立响应式网格承载主机到权重字段');
assert.match(instanceEditor, /className=\{style\.instanceContextList\}[\s\S]*title=\{namespace\}[\s\S]*title=\{service\}/,
  '实例身份摘要必须对命名空间和服务长文本提供省略与完整提示');
assert.match(instanceEditor, /className=\{style\.instanceDrawer\}[\s\S]*bodyClassName=\{style\.instanceDrawerBody\}[\s\S]*footerClassName=\{style\.instanceDrawerFooter\}/,
  '实例抽屉必须提供独立的头、正文滚动和固定页脚样式锚点');
assert.match(instanceEditor, /size="min\(960px, calc\(100vw - 32px\)\)"[\s\S]*showOverlay[\s\S]*closeOnOverlayClick[\s\S]*closeOnEscKeydown/,
  '实例编辑抽屉必须充分利用宽屏并保留标准模态关闭行为');
assert.match(instanceEditor, /footer=\{op === 'view' \? false : \([\s\S]*取消[\s\S]*form\.submit\(\)[\s\S]*提交/,
  '实例编辑操作必须放在 Drawer 固定页脚，而不是正文滚动区域');
assert.doesNotMatch(instanceEditor, /<FormItem className=\{style\.instanceFormFooter\}>/,
  '实例表单正文不能再保留 sticky FormItem 页脚');
assert.match(instanceEditor, /disabled: op === 'create' && option\.value === HealthCheckType\.Heartbeat/,
  'Console 手工创建实例时必须禁用心跳上报选项');
assert.match(instanceEditor, /TCP 探测[\s\S]*HTTP 探测/,
  'Console 手工创建实例必须开放 TCP 和 HTTP 主动探测');
assert.match(instanceEditor, /type: HealthCheckType\.TCP,[\s\S]*tcp: \{ interval: DefaultProbeInterval \}/,
  'Console 手工创建实例必须默认使用 TCP 主动探测');
assert.match(instanceEditor, /HTTP 探测路径[\s\S]*name="health_check_http_path"[\s\S]*探测间隔[\s\S]*name="health_check_probe_interval"/,
  'HTTP 主动探测必须配置路径和探测间隔');
assert.match(instanceEditor, /http: \{[\s\S]*form\.getFieldValue\('health_check_probe_interval'\)[\s\S]*form\.getFieldValue\('health_check_http_path'\)/,
  'HTTP 动态表单字段必须在提交边界组装为协议结构，避免嵌套字段串值');
assert.match(instanceEditor, /Form\.useWatch\(\['healthCheck', 'type'\],[\s\S]*watchedEnableHealthCheck && watchedHealthCheckType === HealthCheckType\.HTTP/,
  '健康检查动态字段必须通过 useWatch 作为顶层 FormItem 挂载');
assert.doesNotMatch(instanceEditor, /<FormItem shouldUpdate=[\s\S]*<FormItem label="HTTP 探测路径"/,
  'HTTP 路径和间隔不能嵌套在 shouldUpdate FormItem 中，否则 TDesign 会串写字段值');
assert.match(instanceService, /enable_health_check: enableHealthCheck,[\s\S]*health_check: healthCheck/,
  '实例创建和更新请求必须把 camelCase 健康检查字段转换为 proto JSON snake_case');
assert.match(instanceService, /enableHealthCheck: value\.enableHealthCheck \?\? value\.enable_health_check[\s\S]*normalizeHealthCheck\(value\.healthCheck \?\? value\.health_check\)/,
  '实例查询响应必须把 proto JSON 健康检查字段归一化回页面 camelCase');
assert.match(serviceDetailStyle, /\.instanceRuntimeGrid\s*\{[^}]*grid-template-columns: repeat\(2, minmax\(0, 1fr\)\)/,
  '实例运行信息必须在宽抽屉中使用双栏，充分利用横向空间');
assert.match(serviceDetailStyle, /\.instanceDrawerBody[\s\S]*overflow-y: auto[\s\S]*\.instanceDrawerFooter[\s\S]*flex: 0 0 auto/,
  '实例抽屉正文必须独立滚动且页脚固定');
assert.match(serviceDetailStyle, /@media \(max-width: 900px\)[\s\S]*\.instanceRuntimeGrid[\s\S]*grid-template-columns: minmax\(0, 1fr\)/,
  '实例表单在 900px 及以下必须退化为单列');
assert.match(serviceDetailStyle, /\.instanceTableSurface[\s\S]*overflow-x: auto[\s\S]*:global\(\.fluent-table-scroll table\)[\s\S]*min-width: 1180px/,
  '实例宽表必须在页面内部横向滚动，而不是把所有列挤进视口');
assert.match(serviceDetailStyle, /\.instanceTableSurface[\s\S]*:global\(\.fluent-table-state\)[\s\S]*min-height: 120px/,
  '实例空表格必须保持紧凑，不能占满大部分视口');
assert.match(instanceEditor, /title=\{isPresent\(children\) \? String\(children\) : emptyText\}/,
  '实例详情长值必须提供原生完整提示');
assert.match(instanceTable, /className=\{style\.instanceToolbar\}[\s\S]*className=\{style\.instanceToolbarActions\}/,
  '实例页工具栏必须使用本地 flex 布局，不能依赖不完整的 Row/Col 兼容层');
assert.match(instanceTable, /className=\{style\.instanceTableSurface\}[\s\S]*<Table/,
  '实例表格必须放进内部横向滚动容器');
assert.doesNotMatch(instanceTable, /<Row justify='space-between'|<Col>|批量删除/,
  '实例工具栏不能继续依赖 Row/Col，也不能显示没有实际处理逻辑的批量删除');
assert.match(instanceTable, /case 'delete':[\s\S]*dispatch\(removeInstances\(\{ ids: \[String\(row\?\.id\)\] \}\)\)[\s\S]*删除服务实例成功[\s\S]*refreshTable\(page, limit\)/,
  '实例删除必须执行真实删除请求并刷新当前页，不能误打开创建抽屉');
assert.doesNotMatch(instanceTable, /case 'delete':(?:(?!case 'view':)[\s\S])*setEditState/,
  '实例删除不能再进入编辑器状态');
assert.match(instanceTable, /host: query \|\| undefined/,
  '实例主机搜索必须真实传入查询请求');
assert.match(alias, /className=\{style\.aliasSummaryValue\} id="sumTarget" title=\{targetServiceLabel\}/,
  '别名摘要中的长目标服务必须省略并在悬浮时展示完整内容');
assert.match(style, /\.aliasDetailActions[\s\S]*display: flex[\s\S]*flex-wrap: wrap/,
  '别名工具栏操作区必须允许控件整体换行，避免压缩成逐字显示');
assert.doesNotMatch(style, /@media \(min-width: 1260px\)[\s\S]*overflow-x: hidden/,
  'Services 页面不得在宽屏禁用表格内部横向滚动');
assert.match(serviceDetail, /className=\{style\.detailIcon\} aria-hidden="true">svc<\/div>/,
  '服务身份头部必须使用稳定的 svc 资源类型标识');
assert.match(serviceDetail, /复制 ID[\s\S]*查看实例[\s\S]*管理别名/,
  '服务身份头部必须保留复制 ID、查看实例和管理别名三个操作');
assert.match(serviceDetail, /退出编辑[\s\S]*编辑/,
  '服务详情头部必须提供页面内编辑和退出编辑入口');
assert.match(serviceDetail, /describeServiceAlias\(\{[\s\S]*namespace,[\s\S]*service: serviceName,[\s\S]*limit: 1/,
  '服务摘要必须通过现有别名查询接口动态获取别名总数');
assert.match(serviceDetail, /可用实例[\s\S]*注册实例[\s\S]*服务别名[\s\S]*服务标签/,
  '服务摘要总栏必须按设计提供四个稳定指标');
assert.match(serviceDetail, /<ServiceForm[\s\S]*mode=\{editing \? 'edit' : 'view'\}[\s\S]*onSubmitted=\{\(\) => \{[\s\S]*reloadService\(\);/,
  '服务详情必须继续在当前页面内复用 ServiceForm 完成查看和编辑');
assert.doesNotMatch(serviceDetail, /<h3>系统信息<\/h3>|<h3>运行状态<\/h3>|<DetailItem|const formatPorts|<Progress|<Tag/,
  '服务详情底部不能再回归系统信息和运行状态重复区块');
assert.doesNotMatch(serviceDetail, /实例健康[\s\S]*发现结果[\s\S]*别名访问/,
  '实例健康、发现结果和别名访问不能再以独立运行状态区块重复展示');
assert.doesNotMatch(serviceDetail, /<Tag variant="light">\{service\.namespace\}<\/Tag>|调用关系|治理入口/,
  '服务详情头部不能重复展示命名空间标签，也不能塞入调用拓扑或治理入口');
assert.match(serviceDetailStyle, /\.detailStats\s*\{[^}]*display: grid[^}]*grid-template-columns: repeat\(4, minmax\(0, 1fr\)\)/,
  '服务摘要必须使用四列等宽布局');
assert.match(serviceDetailStyle, /\.detailSections\s*\{[^}]*display: grid[^}]*grid-template-columns: minmax\(0, 1fr\)/,
  '详情区只保留内嵌基础信息表单，不能继续按系统信息和运行状态做双栏布局');
assert.match(serviceDetailStyle, /\.detailFormSection\s*\{[^}]*grid-column: 1 \/ -1/,
  '服务详情内嵌查看编辑表单必须横跨详情区，避免页面内编辑被挤压');
assert.doesNotMatch(serviceDetailStyle, /\.detailRows|\.detailItem|\.detailItemWide|\.statusList|\.statusRow|\.statusProgress/,
  '服务详情样式不能保留已移除的系统信息键值布局或运行状态布局');
assert.match(serviceDetailStyle, /@media \(max-width: 960px\)[\s\S]*\.detailSections\s*\{[^}]*grid-template-columns: 1fr/,
  '窄屏详情区必须退化为单列');
assert.match(serviceDetailStyle, /@media \(max-width: 640px\)[\s\S]*\.detailStats\s*\{[^}]*grid-template-columns: 1fr/,
  '小屏摘要必须退化为单列，避免横向滚动');
assert.match(appLayoutStyle, /@media \(max-width: 900px\)[\s\S]*\.sideContainer,[\s\S]*\.topPanel,[\s\S]*\.mixContent\s*\{[^}]*min-width: 0/,
  '全局布局折叠菜单后必须解除内容区固定最小宽度，避免响应式页面整体横向溢出');

assert.match(namespaceIndex, /<ResourceHeader[\s\S]*title="命名空间管理"[\s\S]*actions=\{\(/,
  '命名空间页仍应作为服务页对齐基准，并使用统一 ResourceHeader');
assert.match(namespaceIndex, /<ResourceToolbar[\s\S]*title="命名空间列表"/,
  '命名空间页必须使用统一 ResourceToolbar 作为服务页列表工具栏对齐基准');
assert.match(namespaceStyle, /\.namespaceWorkspace[\s\S]*margin-top: 18px/,
  '命名空间页工作区必须保留页头以下的稳定节奏');
assert.match(namespaceStyle, /\.metricRail[\s\S]*margin: 0/,
  '命名空间页指标条应由工作区 gap 控制间距，避免挤占表格滚动高度');

console.log('discovery services layout verification passed');
