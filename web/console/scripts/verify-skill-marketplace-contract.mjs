import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const read = (path) => readFile(new URL(`../${path}`, import.meta.url), 'utf8');
const [router, routeRuntime, service, list, detail, styles, packageJson, zhLocale, enLocale] = await Promise.all([
  read('src/router/modules/ai.ts'),
  read('src/components/Router/index.tsx'),
  read('src/services/skill_marketplace.ts'),
  read('src/pages/AI/Skills/index.tsx'),
  read('src/pages/AI/Skills/Detail.tsx'),
  read('src/pages/AI/Skills/index.module.less'),
  read('package.json'),
  read('src/locales/zh-CN.json'),
  read('src/locales/en-US.json'),
]);
const menu = await read('src/layouts/components/Menu.tsx');

assert.match(router, /path:\s*'skills'/, '必须注册 /ai/skills 目录路由');
assert.match(router, /path:\s*'skills\/:publisher\/:name'/, '必须注册 publisher/name 深链详情路由');
assert.match(routeRuntime, /const routeMatches[\s\S]*segment\.startsWith\(':'\)/, '路由器必须支持受控 :param 深链匹配');
assert.match(menu, /pathname\.startsWith\('\/ai\/skills\/'\).*'\/ai\/skills'/, 'Skill 深链必须保持目录菜单高亮');
assert.match(service, /SkillMarketplaceAPI = '\/api\/skill-marketplace'/, 'Marketplace API 根必须集中定义');
assert.match(zhLocale, /"menu\.ai\.skills": "Skill 市场"/, '中文界面必须使用 Skill 市场');
assert.match(enLocale, /"menu\.ai\.skills": "Skill Marketplace"/, '英文界面必须保留 Skill Marketplace');
assert.match(list, /marketplaceLabel = t\('menu\.ai\.skills'\)/, '目录页标题必须跟随 locale');
assert.match(detail, /marketplaceLabel = t\('menu\.ai\.skills'\)/, '详情页标题必须跟随 locale');
for (const endpoint of ['/v1/skills', '/releases/${encodeURIComponent(version)}/bundle-manifest', '/v1/reviews', '/v1/registry-sources']) {
  assert.ok(service.includes(endpoint), `服务必须覆盖 ${endpoint}`);
}
assert.match(service, /\/v1\/skills\/\$\{encodeURIComponent\(publisher\)\}\/\$\{encodeURIComponent\(name\)\}\/grants/, '服务必须覆盖私有 Skill grants GET/PUT 路径');
assert.match(service, /principalType, principalId/, '更新 grants 必须使用后端 camelCase 主体字段');
assert.match(service, /PUT \/v1\/reviews\/\{releaseId\}/, '审核接口必须保持 GET /v1/reviews + PUT /v1/reviews/{releaseId}');
assert.match(service, /\/v1\/skills\/releases\/upload[\s\S]*\/v1\/skills\/releases\/import-git/, '上传和 Git import 必须保持后端约定路径');
assert.doesNotMatch(service, /releases\/\$\{encodeURIComponent\(version\)\}\/bundle['`]/, 'raw ZIP /bundle 不能被当作 JSON 读取');
for (const field of ['latestVersion', 'displayName', 'publishedAt', 'scanStatus', 'signatureStatus', 'sourceURL']) {
  assert.ok(service.includes(field), `服务必须兼容后端 camelCase 字段 ${field}`);
}
assert.match(service, /http_index/, 'Registry Source 必须使用 http_index 类型');
assert.doesNotMatch(service, /Content-Type['"]\s*:\s*['"]multipart\/form-data/, 'multipart 不得手设 Content-Type，交由 Axios 添加 boundary');
assert.match(list, /搜索 Skill[\s\S]*来源筛选[\s\S]*可见性筛选[\s\S]*状态筛选/, '目录必须提供搜索、来源、可见性与状态筛选');
assert.match(list, /上传 Bundle[\s\S]*Git 导入[\s\S]*审核工作台[\s\S]*Registry Source/, '目录必须提供发布、导入、审核和来源管理入口');
assert.match(list, /visibility === 'public' && !signature/, '公共 Release 必须在提交前要求 detached signature');
assert.match(list, /accept="application\/json,.json"[\s\S]*algorithm=Ed25519、signedAt（RFC3339）和 signature（base64）/, '公共签名必须明确为 detached JSON envelope，不得暗示原始 .sig 文件');
assert.match(list, /EditorDrawer[\s\S]*width="workspace"/, '审核和来源管理必须复用 EditorDrawer 工作区宽度');
assert.doesNotMatch(list, /<Button[^>]*>\s*(执行 Skill|安装 Skill)\s*<\/Button>/, '目录不得提供 Skill 执行或安装动作');
assert.match(list, /pole-ai skill install publisher\/name@version/, '列表必须展示实际 pole-ai 命令语法');
assert.match(detail, /pole-ai skill install \$\{publisher\}\/\$\{name\}@\$\{version \|\| '<version>'\}/, '详情必须展示实际 pole-ai 精确命令');
assert.doesNotMatch(detail, /pole-ai-cli|--sha256/, '详情命令不得使用不存在的 pole-ai-cli 或 digest 参数');
assert.match(detail, /navigator\.clipboard\?\.writeText[\s\S]*复制失败/, '详情必须提供有失败提示的可访问复制操作');
assert.match(detail, /公共 Skill 无需私有授权/, '公共 Skill 必须明确无需私有授权');
for (const text of ['管理访问策略', 'User、UserGroup、Role']) {
  assert.ok(detail.includes(text), `私有 Skill 必须提供 ${text} 访问策略管理`);
}
assert.match(detail, /EditorDrawer[\s\S]*私有 Skill 访问策略/, '访问策略必须复用可访问 EditorDrawer');
assert.match(detail, /二进制文件不预览/, 'Bundle 浏览必须拒绝预览二进制文件');
assert.match(detail, /SKILL\.md/, 'Bundle 浏览应优先选中 SKILL.md');
assert.match(detail, /不可编辑/, '详情必须声明已发布 Bundle 不可编辑');
assert.match(styles, /@media \(max-width: 920px\)/, '样式必须覆盖窄视口');
assert.match(styles, /var\(--app-surface\)/, '样式必须使用主题 token');
assert.match(packageJson, /test:skill-marketplace/, 'package script 必须暴露专项契约检查');

console.log('Skill Marketplace Console 契约检查通过。');
