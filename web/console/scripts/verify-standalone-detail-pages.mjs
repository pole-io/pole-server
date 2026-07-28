import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const routerAI = read('src/router/modules/ai.ts');
assert.match(routerAI, /path:\s*'mcps\/detail'[\s\S]*pages\/AI\/Mcp\/Detail/, 'MCP 必须注册独立详情路由');
assert.match(routerAI, /path:\s*'a2a\/detail'[\s\S]*pages\/AI\/A2A\/Detail/, 'A2A 必须注册独立详情路由');

const routerGovernance = read('src/router/modules/governance.ts');
assert.match(routerGovernance, /path:\s*'rules\/detail'[\s\S]*pages\/Governance\/RuleDetailPage/, '治理规则必须注册独立详情路由');

const mcpList = read('src/pages/AI/Mcp/index.tsx');
assert.match(mcpList, /navigate\(`\/ai\/mcps\/detail\?id=/, 'MCP 列表查看入口必须跳转独立详情页');
assert.doesNotMatch(mcpList, /case 'detail':[\s\S]{0,160}setToolsState\(\{\s*visible:\s*true/, 'MCP 查看入口不能再打开详情抽屉');

const a2aList = read('src/pages/AI/A2A/index.tsx');
assert.match(a2aList, /navigate\(`\/ai\/a2a\/detail\?id=/, 'A2A 列表查看入口必须跳转独立详情页');
assert.doesNotMatch(a2aList, /case 'detail':[\s\S]{0,160}openAgentDetail/, 'A2A 查看入口不能再打开详情抽屉');
assert.doesNotMatch(a2aList, /case 'skills':[\s\S]{0,160}openAgentDetail/, 'A2A 技能入口不能再打开详情抽屉');

const governanceWorkbench = read('src/pages/Governance/Workbench/index.tsx');
assert.match(governanceWorkbench, /navigate\(`\/governance\/rules\/detail\?\$\{params\.toString\(\)\}`\)/, '治理工作台查看入口必须跳转独立详情页');
const openRuleStart = governanceWorkbench.indexOf('const openRule =');
const deleteRuleStart = governanceWorkbench.indexOf('const deleteRule =', openRuleStart);
assert.ok(openRuleStart >= 0 && deleteRuleStart > openRuleStart, '必须能定位治理工作台 openRule 函数');
const openRuleBody = governanceWorkbench.slice(openRuleStart, deleteRuleStart);
assert.doesNotMatch(openRuleBody, /setDrawerVisible\(true\)/, '治理规则查看函数不能再打开详情抽屉');

const governanceDetail = read('src/pages/Governance/RuleDetailPage.tsx');
assert.match(governanceDetail, /RuleDetailFrame/, '治理规则独立详情页必须使用页面版详情容器');
assert.match(governanceDetail, /RuleTabs/, '治理规则独立详情页必须保留规则、版本、监听 Tab');

const mcpDetail = read('src/pages/AI/Mcp/Detail.tsx');
assert.match(mcpDetail, /ToolExplorer/, 'MCP 独立详情页必须保留工具浏览能力');
assert.match(mcpDetail, /AuthorizeInput/, 'MCP 独立详情页必须保留资源授权入口');

const a2aDetail = read('src/pages/AI/A2A/Detail.tsx');
assert.match(a2aDetail, /AgentCardView/, 'A2A 独立详情页必须保留 Agent Card 视图');
assert.match(a2aDetail, /AgentSkillsView/, 'A2A 独立详情页必须保留技能浏览');
assert.match(a2aDetail, /A2AEditor[\s\S]*embedded/, 'A2A 独立详情页必须在页面内承载编辑 Tab');

const userTable = read('src/pages/Auth/Principal/UserTable.tsx');
assert.match(userTable, /navigate\(`\/auth\/principals\/userdetail\?name=/, '用户查看入口必须跳转独立详情页');

const groupTable = read('src/pages/Auth/Principal/GroupTable.tsx');
assert.match(groupTable, /navigate\(`\/auth\/principals\/groupdetail\?name=/, '用户组查看入口必须跳转独立详情页');

const roleTable = read('src/pages/Auth/Principal/RoleTable.tsx');
assert.match(roleTable, /navigate\(`\/auth\/principals\/roledetail\?name=/, '角色查看入口必须跳转独立详情页');

const policyTable = read('src/pages/Auth/Policy/PolicyTable.tsx');
assert.match(policyTable, /navigate\(`\/auth\/policies\/detail\?name=/, '权限策略查看入口必须跳转独立详情页');

console.log('复杂资源独立详情页静态约束验证通过');
