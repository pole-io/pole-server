import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const workbench = fs.readFileSync(path.join(process.cwd(), 'src/pages/Governance/Workbench/index.tsx'), 'utf8');
const router = fs.readFileSync(path.join(process.cwd(), 'src/router/modules/governance.ts'), 'utf8');
const createPage = fs.readFileSync(path.join(process.cwd(), 'src/pages/Governance/RuleCreatePage.tsx'), 'utf8');
const frameStyle = fs.readFileSync(path.join(process.cwd(), 'src/pages/Governance/RuleRelease/RuleDetailDrawer.module.less'), 'utf8');

assert.doesNotMatch(workbench, /pendingCreateType|createWizardStep|confirmCreateWizard|confirmBtn="进入创建"|RuleDetailDrawer|setDrawerVisible|drawerMode/);
assert.match(workbench, /const createRuleFromType = \(type: string\) => \{[\s\S]*setCreateWizardVisible\(false\);[\s\S]*openCreateRule\(type\);/);
assert.match(workbench, /new URLSearchParams\(\{ kind: target\.kind \}\)/);
assert.match(workbench, /navigate\(`\/governance\/rules\/create\?\$\{params\.toString\(\)\}`\)/);
assert.match(workbench, /aria-label=\{`创建\$\{item\.label\}规则`\}[\s\S]*onClick=\{\(\) => createRuleFromType\(item\.value\)\}/);
assert.match(router, /path: 'rules\/create'[\s\S]*RuleCreatePage/);
assert.match(createPage, /RuleDetailFrame title=\{`新建\$\{target\.label\}规则`\}/);
assert.match(createPage, /CustomRouteEditor op="create"[\s\S]*RateLimitEditor[\s\S]*CircuitBreakerEditor[\s\S]*TrafficGovernanceEditor/);
assert.match(createPage, /resetCustomRoute[\s\S]*resetRateLimitRule[\s\S]*resetCircuitBreaker[\s\S]*resetLaneGroup/);
assert.match(createPage, /className=\{`\$\{style\.standalonePage\} \$\{style\.createPage\}`\}/);
assert.match(frameStyle, /\.createPage\s*\{[\s\S]*height: calc\(100dvh - 40px\)[\s\S]*overflow: hidden/);
assert.match(frameStyle, /\.createPane\s*\{[\s\S]*overflow: auto[\s\S]*overscroll-behavior: contain/);
assert.match(frameStyle, /\.pageFrame\s+:global\(\.fluent-sticky-tool\)\s*\{\s*display: none !important/);

console.log('governance standalone create verification passed');
