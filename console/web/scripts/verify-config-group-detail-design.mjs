import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const filesPage = read('src/pages/Configuration/Group/Files/index.tsx');
const fileView = read('src/pages/Configuration/Group/Files/FileView.tsx');
const authorizeInput = read('src/components/Authorize/index.tsx');
const publishForm = read('src/pages/Configuration/Group/Releases/PublishForm.tsx');
const grayRuleEditor = read('src/pages/Configuration/Group/Releases/GrayRuleEditor.tsx');
const releaseTable = read('src/pages/Configuration/Group/Releases/ReleaseTable.tsx');
const subscribeTable = read('src/pages/Configuration/Group/Files/SubscribeTable.tsx');
const filesStyle = read('src/pages/Configuration/Group/Files/index.module.less');
const codeEditorStyle = read('src/components/CodeEditor/index.module.less');
const configReleaseServer = read('../../pkg/config/config_file_release.go');
const configReleaseTypes = read('../../apis/pkg/types/config/config_file.go');

const combined = [filesPage, fileView, publishForm, releaseTable, subscribeTable, filesStyle].join('\n');

for (const forbidden of [
  'CONFIG PREVIEW',
  'renderPreview',
  '发布影响',
  'renderImpact',
  'toggleEditorHeight',
]) {
  assert.doesNotMatch(combined, new RegExp(forbidden), `配置分组详情不能恢复旧模块关键词：${forbidden}`);
}

assert.doesNotMatch(fileView, /StickyTool/, '配置编辑页签不能继续使用右侧悬浮 StickyTool 操作条。');
assert.match(filesPage, /className=\{style\.workbench\}/, '配置分组详情必须使用文件树 + 中央业务画布工作台结构。');
assert.match(filesPage, /className=\{style\.fileExplorer\}/, '配置分组详情必须保留左侧文件树资源浏览器。');
assert.match(fileView, /className=\{style\.resourceTabs\}/, '配置分组详情必须使用文件资源业务页签。');
for (const tab of ['文件内容', '基本信息', '发布记录', '订阅查询']) {
  assert.match(fileView, new RegExp(`label="${tab}"`), `配置文件详情必须保留业务页签：${tab}`);
}

assert.doesNotMatch(fileView, /文件上下文/, '当前文件标题区不能展示抽象标题“文件上下文”。');
assert.match(fileView, /当前配置内容[\s\S]*编辑配置内容|编辑配置内容[\s\S]*当前配置内容/, '配置编辑页签必须清楚区分只读内容与编辑内容。');
assert.match(fileView, /fileSummary/, '配置编辑页签必须使用当前文件摘要区展示文件名、路径语境和状态 tag。');
assert.match(fileView, /EnvironmentResourceSwitcher[\s\S]*resourcePath/, '当前文件标题区必须通过环境切换器与资源路径共同表达命名空间和配置分组语境。');
for (const label of ['修改时间', '创建时间', '加密算法', '文件标签']) {
  assert.match(fileView, new RegExp(label), `当前文件字段网格必须保留：${label}`);
}
for (const forbidden of ['fieldLabel}>命名空间', 'fieldLabel}>配置分组', 'fieldLabel}>订阅客户端', 'fieldLabel}>当前版本', 'fieldLabel}>文件状态']) {
  assert.doesNotMatch(fileView, new RegExp(forbidden), `当前文件字段区不能重复展示：${forbidden}`);
}
assert.match(fileView, /保存草稿/, '配置编辑页签编辑态必须提供保存草稿操作。');
assert.match(fileView, /发布配置/, '配置编辑页签查看态必须提供发布配置入口。');
assert.match(fileView, /onAuthorize\?: \(\) => void/, '配置文件详情必须声明授权入口回调。');
assert.match(fileView, /OperationButton[\s\S]*action="authorize"[\s\S]*授权/, '配置文件详情查看态必须提供授权按钮。');
assert.match(codeEditorStyle, /\.monacoSectionFull\s*\{/, 'CodeEditor 全屏 class 必须与组件引用一致，不能让全屏编辑失效。');
assert.match(filesPage, /AuthorizeInput/, '配置文件详情页必须挂载授权抽屉。');
assert.match(filesPage, /resource_type=\{PolicySourceType\.ConfigGroups\}/, '配置文件授权当前必须按后端契约使用所属配置分组资源类型。');
assert.match(filesPage, /resource_id=\{activeGroup\?\.id/, '配置文件授权必须使用所属配置分组资源 ID。');
assert.match(filesPage, /resource_name=\{\`\$\{namespace \|\| '-'\}\/\$\{group \|\| '-'\}\/\$\{selectedFileName/, '配置文件授权资源名必须包含命名空间、配置分组和文件名。');
assert.match(authorizeInput, /resourceFileName/, '通用授权抽屉必须支持配置文件名摘要。');
assert.match(authorizeInput, /resourceNameParts\.slice\(2\)\.join\('\/'\)/, '配置文件名可能包含路径分隔符，授权摘要必须合并第三段之后的完整文件名。');
assert.match(authorizeInput, /FormItem label="配置文件"/, '配置文件授权抽屉必须展示配置文件字段。');

assert.match(publishForm, /版本对比/, '发布抽屉第一步必须是版本对比。');
assert.match(publishForm, /版本信息/, '发布抽屉第二步必须先展示版本信息。');
assert.match(publishForm, /发布范围/, '发布抽屉第二步必须展示发布范围。');
assert.match(publishForm, /灰度优先级/, '灰度发布态必须提供独立优先级入口。');
assert.match(publishForm, /GrayRuleEditor/, '灰度发布态必须复用配置灰度规则适配组件。');
assert.doesNotMatch(publishForm, /ClientLabelInput/, '配置发布抽屉不能继续直接使用旧 ClientLabelInput 私有灰度规则控件。');
assert.match(grayRuleEditor, /TrafficMatchConditionEditor/, '配置灰度规则适配组件必须复用治理侧通用 TrafficMatchConditionEditor。');
assert.match(grayRuleEditor, /grayRowsToBetaLabels/, '配置灰度规则适配组件必须提供通用控件行到 betaLabels 的转换。');
assert.match(grayRuleEditor, /showParamKey=\{false\}/, '配置灰度规则应隐藏治理通用控件的参数键列，避免内置客户端标签出现无意义输入。');

for (const label of ['正式发布', '灰度发布', '正式草稿', '历史记录']) {
  assert.match(releaseTable, new RegExp(label), `发布记录必须包含内部 Tab：${label}`);
}
assert.match(releaseTable, /pageSize\s*=\s*6/, '发布记录每个类型默认每页 6 条。');
assert.match(releaseTable, /提交为正式草稿/, '灰度发布记录必须支持提交为正式草稿。');
assert.match(releaseTable, /命中客户端会回落当前全量/, '删除灰度必须提示命中客户端回落当前全量。');
assert.match(releaseTable, /未命中灰度的客户端将无法获取配置/, '删除当前全量必须强提示客户端无法获取配置。');
assert.match(releaseTable, /components\/OperationButton/, '发布记录行操作必须使用统一 OperationButton 包装，保证图标按钮有悬浮提示。');
assert.match(releaseTable, /OperationButton[\s\S]*action="view"[\s\S]*label="对比"/, '发布记录对比操作必须使用统一查看图标语义。');
assert.match(releaseTable, /ConfirmOperationButton[\s\S]*action="rollback"/, '发布记录确认类回滚操作必须使用统一 ConfirmOperationButton。');
assert.match(releaseTable, /ConfirmOperationButton[\s\S]*action="delete"/, '发布记录确认类删除操作必须使用统一 ConfirmOperationButton。');
assert.match(releaseTable, /const\s+OperatorLink/, '发布记录发布人必须通过统一 OperatorLink 渲染，支持跳转用户详情。');
assert.match(releaseTable, /target="_blank"/, '发布记录发布人链接必须新标签打开用户详情。');
assert.match(releaseTable, /describeUsers/, '发布记录必须按发布人名称解析用户 id，不能拼缺 id 的详情链接。');
assert.match(releaseTable, /releaseTimeText/, '发布记录发布时间必须有 createTime/modifyTime 兜底展示，不能让灰度行为空。');

assert.match(
  configReleaseServer,
  /Ctime:\s*commontime\.Time2String\(item\.CreateTime\)[\s\S]*Mtime:\s*commontime\.Time2String\(item\.ModifyTime\)/,
  '配置发布列表接口必须返回 ctime/mtime，否则发布记录发布时间会为空。',
);
assert.match(
  configReleaseTypes,
  /CreateBy:\s*release\.CreateBy[\s\S]*ModifyBy:\s*release\.ModifyBy/,
  '配置发布详情转换必须返回 create_by/modify_by，否则操作人无法稳定跳转详情。',
);

for (const label of ['客户端ID', '客户端IP', '类型', '标签', '监听版本', '最近拉取']) {
  assert.match(subscribeTable, new RegExp(label), `订阅查询表格必须包含列：${label}`);
}
assert.match(subscribeTable, /命中灰度/, '订阅查询必须展示命中灰度监听版本状态。');
assert.match(subscribeTable, /当前全量/, '订阅查询必须展示当前全量监听版本状态。');
assert.match(subscribeTable, /无可用版本/, '订阅查询必须展示无可用版本状态。');

console.log('config group detail design checks passed');
