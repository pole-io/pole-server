import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const files = {
  servicesIndex: 'src/pages/Discovery/Services/index.tsx',
  servicesTable: 'src/pages/Discovery/Services/services.tsx',
  serviceEditor: 'src/pages/Discovery/Services/ServiceEditor.tsx',
  labelInput: 'src/components/LabelInput/index.tsx',
  aliasTable: 'src/pages/Discovery/Services/alias.tsx',
  aliasEditor: 'src/pages/Discovery/Services/AliasEditor.tsx',
  serviceInstance: 'src/pages/Discovery/Services/Instance/index.tsx',
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
const labelInput = sources.labelInput;
const alias = sources.aliasTable;
const aliasEditor = sources.aliasEditor;
const serviceInstance = sources.serviceInstance;
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

assert.match(index, /<section className=\{style\.header\}[\s\S]*<Space>/,
  '注册发现页头必须像命名空间页一样在右侧提供操作区');
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
assert.match(style, /\.aliasDetailActions[\s\S]*grid-template-columns: auto minmax\(180px, 260px\) auto auto/,
  '服务详情内别名清单操作区必须稳定排列新建、搜索、查询和重置');
assert.match(style, /\.aliasDetailTableSurface[\s\S]*:global\(\.t-table\)[\s\S]*min-width: 720px/,
  '服务详情内别名表格必须降低最小宽度，避免内嵌宽表挤压');
assert.match(style, /\.aliasSummaryBar[\s\S]*display: grid[\s\S]*grid-template-columns: minmax\(220px, 1\.4fr\) repeat\(2, minmax\(120px, 1fr\)\)[\s\S]*border-bottom: 1px solid #e5e8ef/,
  '服务详情内别名总栏必须放在表格面板顶部，并保持一行轻量统计');
assert.match(style, /\.aliasSummaryItem[\s\S]*min-height: 64px[\s\S]*border-right: 1px solid #e5e8ef/,
  '别名总栏必须是轻量横向统计，不应升级为外部统计卡阵列');
assert.match(style, /\.aliasSummaryValue[\s\S]*font-family: ui-monospace[\s\S]*text-overflow: ellipsis/,
  '别名总栏目标服务等值必须使用 mono 且避免长文本撑破布局');
assert.match(style, /\.aliasDetailTableSurface[\s\S]*:global\(\.t-table__empty\)[\s\S]*box-sizing: border-box[\s\S]*height: 92px[\s\S]*min-height: 92px[\s\S]*padding: 18px 0/,
  '服务详情内别名表格空态必须收敛高度，不能撑出大面积空白');
assert.match(style, /\.aliasDetailTableSurface[\s\S]*:global\(\.t-table__pagination\)[\s\S]*padding: 12px 16px/,
  '服务详情内别名表格分页必须使用紧凑安全区');
assert.match(style, /\.aliasReadonlyValue[\s\S]*background: #f7f9fc[\s\S]*font-family: ui-monospace/,
  '服务别名抽屉内目标服务必须作为只读信息展示，不应伪装成可切换输入');
assert.match(style, /\.aliasDrawer[\s\S]*:global\(\.t-drawer__body\)[\s\S]*overflow-y: auto/,
  '服务别名抽屉内部应自行滚动，避免把滚动交给页面');
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
assert.match(style, /\.tableSurface[\s\S]*overflow-x: auto/,
  '服务表格外层应支持横向溢出，避免宽表挤压');
assert.match(style, /\.serviceTableSurface[\s\S]*display: flex[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0[\s\S]*overflow: hidden/,
  '服务清单表格外层必须吃满剩余高度并阻止外层页面滚动');
assert.match(style, /\.serviceTableSurface[\s\S]*:global\(\.t-table\)[\s\S]*display: flex[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0/,
  '服务清单 TDesign 表格必须参与内部 flex 高度分配');
assert.match(style, /\.serviceTableSurface[\s\S]*:global\(\.t-table__content\)[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0[\s\S]*overflow: auto[\s\S]*overscroll-behavior: contain/,
  '服务清单必须把纵向滚动挂到 t-table__content，并阻止滚动链传到页面');
assert.match(style, /\.tableSurface[\s\S]*:global\(\.t-table th\)[\s\S]*padding: 14px/,
  '服务表头密度应与命名空间页一致');
assert.match(style, /\.healthTrack\b/,
  '服务健康列应采用与命名空间页同类的健康进度条');
assert.match(style, /\.actionCell\b/,
  '服务操作列应采用与命名空间页同类的紧凑操作容器');
assert.match(style, /\.serviceForm[\s\S]*height: 100%/,
  '创建服务抽屉表单必须占满抽屉正文并交给抽屉纵向滚动');
assert.match(style, /\.serviceForm[\s\S]*:global\(\.t-form__item\)[\s\S]*margin-bottom: 0/,
  '创建服务抽屉字段间距必须挂到字段 wrapper，不能依赖 FormItem 内部 margin');
assert.match(style, /\.serviceDrawerContent[\s\S]*gap: 22px/,
  '创建服务抽屉分段之间必须保留稳定间距');
assert.match(style, /\.serviceFormSection[\s\S]*> div\[id\^="in"\][\s\S]*margin-bottom: 18px[\s\S]*> div\[id\^="in"\]:last-child[\s\S]*margin-bottom: 0/,
  '创建服务抽屉连续输入字段必须通过锚点 wrapper 保持纵向间距');
assert.match(style, /\.serviceFormSectionTitle[\s\S]*&::before[\s\S]*background: #0052d9[\s\S]*&::after[\s\S]*background: #e5e8ef/,
  '创建服务抽屉分段标题必须使用竖色条和发丝线');
assert.match(style, /\.nameMeta[\s\S]*margin-left: 104px[\s\S]*justify-content: space-between/,
  '服务名称计数和不可修改提示必须对齐输入控件列');
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
assert.match(services, /<section className=\{style\.metricRail\}[\s\S]*<section className=\{style\.listSection\}>[\s\S]*<section className=\{style\.filterBar\}>[\s\S]*<section id="tbody" className=\{`\$\{style\.tableSurface\} \$\{style\.serviceTableSurface\}`\}>/,
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
assert.match(services, /tableLayout=\{'fixed'\}/,
  '服务表格布局必须对齐命名空间页');

assert.match(serviceEditor, /import LabelInput from 'components\/LabelInput'/,
  '创建服务抽屉必须复用统一 LabelInput 标签控件');
assert.match(serviceEditor, /const SERVICE_NAME_REG = \/\^\[0-9A-Za-z\._-\]\+\$\/;/,
  '创建服务名称必须显式约束数字、英文字母、.、-、_');
assert.match(serviceEditor, /describeAllServices\(\)[\s\S]*then\(setAllServices\)/,
  '创建服务抽屉必须加载全量服务用于同命名空间重名校验，且不污染服务列表 store');
assert.match(serviceEditor, /validateServiceName[\s\S]*该命名空间下服务名已存在/,
  '创建服务提交前必须校验同命名空间服务名不可重复');
assert.match(serviceEditor, /validateLabels[\s\S]*标签键不能为空[\s\S]*标签键 \$\{key\} 重复/,
  '服务标签提交前必须校验键为空和值重复');
assert.match(serviceEditor, /id="drawer"[\s\S]*基础信息[\s\S]*id="inNs"[\s\S]*id="inName"[\s\S]*id="inDesc"[\s\S]*归属信息[\s\S]*id="inDept"[\s\S]*id="inBiz"[\s\S]*服务标签/,
  '创建服务抽屉必须按基础信息、归属信息、服务标签三段组织并提供字段锚点');
assert.match(serviceEditor, /id="nameCount"[\s\S]*id="errName"/,
  '服务名称字段必须提供实时计数和内联错误锚点');
assert.match(serviceEditor, /<LabelInput[\s\S]*name="service_labels"[\s\S]*hideLabel[\s\S]*editorId="tagRows"[\s\S]*emptyId="tagsEmpty"[\s\S]*countId="tagCount"/,
  '服务标签编辑器必须通过统一 LabelInput 提供行容器、空态和计数锚点');
assert.match(serviceEditor, /openInfoNotification\('请求成功', op === 'create' \? '服务已创建' : '修改服务成功'\)/,
  '创建服务成功提示必须使用服务已创建文案');
assert.match(serviceEditor, /size=\{op === 'view' \? '680px' : 'min\(720px, 94vw\)'\}[\s\S]*showOverlay[\s\S]*closeOnOverlayClick[\s\S]*closeOnEscKeydown[\s\S]*destroyOnClose/,
  '创建服务抽屉必须使用 min(720px, 94vw)，并支持遮罩点击、Esc 关闭和销毁');
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
assert.match(alias, /import \{ AddIcon, CopyIcon, DeleteIcon, EditIcon, SearchIcon \} from 'tdesign-icons-react'/,
  '服务详情内别名行操作必须包含复制别名图标');
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
assert.match(alias, /embedded && namespace && serviceName[\s\S]*`\$\{namespace\}\/\$\{serviceName\} 下当前显示 \$\{datas\.length\} 条`/,
  '服务详情内别名清单必须说明当前服务上下文');
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

assert.match(namespaceIndex, /<section className=\{style\.header\}[\s\S]*<Space>/,
  '命名空间页仍应作为服务页对齐基准');
assert.match(namespaceStyle, /\.metricRail[\s\S]*margin: 18px 0/,
  '命名空间页指标条节奏是本次对齐基准');

console.log('discovery services layout verification passed');
