import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const baseTypes = read('src/services/types.ts');
const groups = read('src/services/config_group.ts');
const files = read('src/services/config_files.ts');
const releases = read('src/services/config_release.ts');
const groupPage = read('src/pages/Configuration/Group/group.tsx');
const groupEditor = read('src/pages/Configuration/Group/ConfigGroupEditor.tsx');
const fileListPage = read('src/pages/Configuration/Group/Files/index.tsx');
const fileCreator = read('src/pages/Configuration/Group/Files/FileCreator.tsx');
const codeDiffEditor = read('src/components/CodeDiffEditor/index.tsx');

assert.match(
  baseTypes,
  /CONFIG_RELEASES\s*=\s*['"]\/config\/v1\/files\/releases['"]/,
  '发布列表、回滚和删除必须使用后端复数 releases 路由基址。',
);

assert.match(
  groups,
  /function\s+normalizeConfigFileGroup/,
  '配置分组 service 必须归一后端 ctime/mtime 等字段，页面不直接感知后端响应形态。',
);
assert.match(
  groups,
  /export\s+interface\s+ConfigFileGroup\s*\{[\s\S]*\bid:\s*string\b/,
  '配置分组 id 必须按后端 ConfigFileGroup.id string 契约建模，不能保留 number。',
);
assert.match(
  groups,
  /function\s+toApiConfigGroupQuery/,
  '配置分组查询必须把前端 group 搜索项映射为后端 name 参数。',
);
assert.match(
  groups,
  /\$\{BaseURL\.CONFIG_GROUP\}\/delette/,
  '配置分组删除必须按当前后端契约调用 /groups/delette。',
);
assert.doesNotMatch(
  groups,
  /\$\{BaseURL\.CONFIG_GROUP\}\/delete/,
  '配置分组删除不能继续调用后端不存在的 /groups/delete。',
);
assert.match(
  groupPage,
  /removeConfigGroups\(\{\s*param:\s*\{[\s\S]*namespace:\s*row\?\.namespace[\s\S]*name:\s*row\?\.name/,
  '配置组页面删除时必须传 namespace/name，后端不会按 id 删除配置分组。',
);
assert.match(
  groupEditor,
  /const\s+labels\s*=\s*\([^)]*form\.getFieldValue\('group_labels'\)[^)]*\)\s*\|\|\s*\[\]/,
  '配置分组新建不填写标签时 group_labels 必须按空数组处理，不能对 undefined reduce。',
);
assert.match(
  groupEditor,
  /metadata:\s*labels\.reduce/,
  '配置分组提交仍应把标签数组转换为后端 metadata 对象。',
);
assert.doesNotMatch(
  groupEditor,
  /id:\s*editGroup\?\.id\s*\|\|\s*0/,
  '配置分组新建不能发送 id: 0，后端 ConfigFileGroup.id 是 string，会导致 JSON 解码失败。',
);
assert.match(
  fileCreator,
  /form\.setFieldsValue\(\{\s*namespace:\s*namespace,\s*group:\s*group,/,
  '配置文件创建抽屉必须把选中配置分组的 namespace/group 回填到表单字段。',
);
assert.match(
  fileCreator,
  /if\s*\(visible\)\s*\{[\s\S]*form\.setFieldsValue\(\{[\s\S]*namespace:\s*namespace,[\s\S]*group:\s*group,[\s\S]*\}\);/,
  '配置文件创建抽屉必须在打开或路由参数变化时刷新 namespace/group 显示值。',
);
assert.match(
  fileCreator,
  /const\s+\[metaValues,\s*setMetaValues\]/,
  '配置文件创建抽屉是分步表单，提交时不能从已卸载的第一步字段读取 name/comment/encrypted，必须缓存元信息。',
);
assert.doesNotMatch(
  fileCreator,
  /name:\s*form\.getFieldValue\('name'\)[\s\S]*comment:\s*form\.getFieldValue\('comment'\)/,
  '配置文件创建提交不能直接从已卸载的第一步表单字段读取 name/comment，否则第二步提交会丢字段。',
);
assert.match(
  fileListPage,
  /dispatch\(listConfigGroups\(\{\s*param:\s*\{[\s\S]*namespace,[\s\S]*group,[\s\S]*offset:\s*0,[\s\S]*limit:\s*1/,
  '配置文件页刷新或直达时必须按 URL namespace/group 补拉配置分组，否则新建入口会因为 editGroup 为空被禁用。',
);
assert.match(
  fileListPage,
  /const\s+activeGroup\s*=\s*ownerGroup\?\.namespace\s*===\s*namespace\s*&&\s*ownerGroup\?\.name\s*===\s*group/,
  '配置文件页只能使用与当前 URL 匹配的配置分组权限，不能误用旧 Redux editGroup。',
);

assert.match(
  files,
  /function\s+normalizeConfigFile/,
  '配置文件 service 必须把后端 labels/ctime/mtime/rtime 归一成前端 tags/createTime/modifyTime/releaseTime。',
);
assert.match(
  files,
  /function\s+toApiConfigFile/,
  '配置文件写请求必须把前端 tags 转成后端 labels。',
);
assert.match(
  files,
  /function\s+normalizeEncryptAlgorithms/,
  '配置文件加密算法响应必须在 service 边界归一，页面不能感知后端 Any/Struct 形态。',
);
assert.match(
  files,
  /res\.value\?\.algorithms/,
  '配置文件加密算法接口当前通过 protobuf Struct.value.algorithms 返回，前端必须兼容该形态。',
);
assert.match(
  files,
  /brief:\s*true/,
  '查询全部配置文件时不能继续发送拼写错误的 berif 参数。',
);
assert.doesNotMatch(
  files,
  /\bberif\b/,
  '配置文件查询不能保留 berif 拼写错误。',
);

assert.match(
  releases,
  /function\s+normalizeConfigFileRelease/,
  '配置发布 service 必须把后端 labels/ctime/mtime 归一成前端 tags/createTime/modifyTime。',
);
assert.match(
  releases,
  /function\s+toApiConfigFileRelease/,
  '配置发布写请求必须通过 service 边界适配字段，不把后端细节扩散到页面。',
);
assert.match(
  releases,
  /file_name:\s*fileName/,
  '配置发布写请求必须把前端 fileName 转成后端 file_name。',
);
assert.match(
  releases,
  /release_description:\s*releaseDescription/,
  '配置发布写请求必须把前端 releaseDescription 转成后端 release_description。',
);
assert.match(
  releases,
  /release_type:\s*releaseType/,
  '配置发布写请求必须把前端 releaseType 转成后端 release_type。',
);
assert.match(
  releases,
  /beta_labels:\s*normalizeClientLabelsForApi\(betaLabels\)/,
  '配置发布写请求必须把前端 betaLabels 转成后端 beta_labels。',
);
assert.match(
  releases,
  /fileName:\s*release\.fileName\s*\|\|\s*release\.file_name/,
  '配置发布响应必须把后端 file_name 归一为前端 fileName。',
);
assert.match(
  releases,
  /releaseType:\s*release\.releaseType\s*\|\|\s*release\.release_type/,
  '配置发布响应必须把后端 release_type 归一为前端 releaseType。',
);
assert.match(
  releases,
  /action:\s*`\$\{BaseURL\.CONFIG_RELEASES\}`/,
  '配置发布列表必须调用 GET /config/v1/files/releases。',
);
assert.match(
  releases,
  /action:\s*`\$\{BaseURL\.CONFIG_RELEASES\}\/rollback`/,
  '配置发布回滚必须调用 PUT /config/v1/files/releases/rollback。',
);
assert.match(
  releases,
  /action:\s*`\$\{BaseURL\.CONFIG_RELEASES\}\/delete`/,
  '配置发布删除必须调用 POST /config/v1/files/releases/delete。',
);
assert.doesNotMatch(
  releases,
  /CONFIG_RELEASE\}\/(?:rollback|delete)/,
  '配置发布不能继续在单数 /files/release 下拼接 rollback/delete。',
);

assert.match(
  codeDiffEditor,
  /originalModelPath=\{\`\$\{modelPathPrefix\}\/original\.\$\{toHighlightLanguage\(props\.language\)\}`\}/,
  '配置发布 diff 编辑器必须给 original model 设置稳定路径，避免 Monaco 抽屉卸载时报 TextModel dispose 错误。',
);
assert.match(
  codeDiffEditor,
  /modifiedModelPath=\{\`\$\{modelPathPrefix\}\/modified\.\$\{toHighlightLanguage\(props\.language\)\}`\}/,
  '配置发布 diff 编辑器必须给 modified model 设置不同稳定路径，不能与 original 共用空 URI。',
);
assert.match(
  codeDiffEditor,
  /keepCurrentOriginalModel=\{true\}/,
  '配置发布 diff 编辑器卸载时不能主动 dispose original model，否则 Monaco 会在切换步骤时报 TextModel dispose 错误。',
);
assert.match(
  codeDiffEditor,
  /keepCurrentModifiedModel=\{true\}/,
  '配置发布 diff 编辑器卸载时不能主动 dispose modified model，否则 Monaco 会在切换步骤时报 TextModel dispose 错误。',
);

console.log('configuration api contract checks passed');
