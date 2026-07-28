import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const groupPage = read('src/pages/Configuration/Group/index.tsx');
const groupTable = read('src/pages/Configuration/Group/group.tsx');
const groupEditor = read('src/pages/Configuration/Group/ConfigGroupEditor.tsx');
const groupStyle = read('src/pages/Configuration/Group/index.module.less');
const authorizeInput = read('src/components/Authorize/index.tsx');

assert.match(groupPage, /className=\{style\.groupListPage\}/, '配置分组列表页必须使用浅灰工作区页面骨架。');
assert.match(groupTable, /<ResourceHeader[\s\S]*title="配置分组"/, '配置分组列表页必须通过统一 ResourceHeader 展示页面标题。');

for (const label of ['分组数', '配置文件数', '配置加密数', '待发布数']) {
  assert.match(groupTable, new RegExp(label), `配置分组列表总栏必须包含指标：${label}`);
}
assert.doesNotMatch(groupTable, /加密占比|encryptedRatio|isDefaultEncrypted|defaultEncrypted/, '配置分组列表不能用分组 metadata 推导文件加密指标。');
assert.match(groupTable, /describeAllConfigFiles/, '配置加密数必须从配置文件列表真实计算，不能从配置分组 metadata 伪造。');
assert.match(groupTable, /encryptedFileCounts/, '配置分组列表必须维护按分组聚合的配置加密数。');
assert.match(groupTable, /const\s+summaryStats/, '配置分组列表页必须从表格数据计算总栏指标，不能写无来源数字。');
assert.match(groupTable, /className=\{style\.summaryBar\}/, '配置分组列表页必须展示总栏 KPI。');
assert.match(groupTable, /warningStat/, '待发布数必须使用警告态样式。');

for (const label of ['关键字', '命名空间', '发布状态']) {
  assert.match(groupTable, new RegExp(label), `配置分组筛选条必须包含：${label}`);
}
assert.doesNotMatch(groupTable, /加密状态|默认加密|默认未加密/, '配置分组筛选条不能包含文件加密状态筛选。');
assert.match(groupTable, /filterState/, '配置分组列表必须维护组合筛选状态。');
assert.match(groupTable, /<ResourceToolbar[\s\S]*title="配置分组列表"/, '配置分组列表筛选条必须使用统一 ResourceToolbar 并和表格分离。');
assert.match(groupTable, /<Button variant="outline" onClick=\{submitFilter\}>查询<\/Button>/, '配置分组筛选条必须提供统一查询按钮。');
assert.match(groupTable, /<Button variant="text" onClick=\{resetFilter\}>重置<\/Button>/, '配置分组筛选条必须提供统一重置入口。');

for (const label of ['分组名称', '命名空间', '文件数', '加密数', '待发布', '操作']) {
  assert.match(groupTable, new RegExp(label), `配置分组表格必须包含列：${label}`);
}
for (const label of ['活跃版本', '订阅客户端']) {
  assert.doesNotMatch(groupTable, new RegExp(label), `配置分组列表不展示文件级/订阅级字段：${label}`);
}
assert.doesNotMatch(groupTable, /title:\s*'部门'|title:\s*'业务'/, '配置分组列表主体列不应继续以部门/业务替代设计要求的运行态列。');

for (const label of ['资源摘要', '授权对象', '权限范围', '查看', '编辑', '发布', '授权管理']) {
  assert.match(authorizeInput, new RegExp(label), `授权抽屉必须表达配置分组级授权语义：${label}`);
}
assert.match(groupTable, /PolicySourceType\.ConfigGroups/, '授权仍必须落到配置分组资源类型。');
assert.doesNotMatch([groupTable, authorizeInput].join('\n'), /store\/cache|Store\/Cache|灰度规则 key|规则 key/, '列表页不能暴露 store/cache/灰度规则 key 这类实现规格。');

for (const label of ['基础信息', '标签']) {
  assert.match(groupEditor, new RegExp(label), `创建/编辑配置分组抽屉必须包含：${label}`);
}
for (const label of ['默认配置', '文件格式默认值', '是否默认加密']) {
  assert.doesNotMatch(groupEditor, new RegExp(label), `配置分组抽屉不能包含文件级默认配置：${label}`);
}
assert.doesNotMatch(groupEditor, /name=["']defaultFormat["']|name=["']defaultEncrypted["']|defaultFormat:\s|defaultEncrypted:\s/, '配置分组抽屉不能把文件级默认配置作为表单字段或提交字段。');
assert.match(groupEditor, /readonly=\{op === 'edit'\}/, '编辑配置分组时命名空间和分组名不能随意切换。');
assert.doesNotMatch(groupEditor, /import\s+\{[^}]*\bSwitch\b[^}]*\}\s+from\s+["']tdesign-react["']|<Switch\b/, '配置分组抽屉不应使用文件加密开关。');
assert.match(groupEditor, /创建成功后回到列表/, '创建/编辑抽屉应在代码注释中保留刷新列表意图，避免改成进入文件正文编辑。');

for (const className of ['groupListPage', 'summaryBar', 'filterBar', 'groupTableSurface', 'mono', 'warningStat']) {
  assert.match(groupStyle, new RegExp(`\\.${className}\\s*\\{`), `配置分组列表样式必须包含 ${className}。`);
}

console.log('config group list design checks passed');
