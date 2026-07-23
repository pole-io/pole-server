import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');

const serviceTabs = read('src/pages/Discovery/Services/Instance/index.tsx');
const workbench = read('src/pages/Governance/Workbench/index.tsx');
const createPage = read('src/pages/Governance/RuleCreatePage.tsx');
const serviceScope = read('src/pages/Governance/shared/ServiceScopeSection.tsx');

assert.match(
    serviceTabs,
    /label="服务订阅"[\s\S]*label="流量治理"/,
    '流量治理必须位于服务订阅之后',
);
assert.match(
    serviceTabs,
    /<GovernanceWorkbench[\s\S]*embedded[\s\S]*serviceContext=\{\{ namespace:/,
    '服务详情必须复用治理工作台并传入当前服务上下文',
);
assert.match(
    workbench,
    /useState<GovernanceServiceRole>\('caller'\)/,
    '服务上下文中新建规则必须默认把当前服务作为 caller',
);
assert.match(
    workbench,
    /isRuleAssociatedWithService[\s\S]*scopedRules[\s\S]*filteredRules/,
    '服务上下文必须先筛选关联规则，再应用工作台查询条件',
);
assert.match(
    workbench,
    /作为主调方[\s\S]*作为被调方/,
    '服务上下文必须允许用户切换 caller/callee 绑定角色',
);
assert.match(
    serviceScope,
    /fixedRole[\s\S]*isFixed[\s\S]*当前服务/,
    '共享服务范围组件必须把固定端渲染为当前服务只读态',
);

for (const component of [
    'CustomRouteEditor',
    'RateLimitEditor',
    'CircuitBreakerEditor',
    'FaultDetectEditor',
    'LossLessEditor',
    'TrafficGovernanceEditor',
    'LaneGroupEdtor',
]) {
    const componentStart = createPage.indexOf(`<${component}`);
    assert.notEqual(componentStart, -1, `${component} 必须存在于治理独立新建页`);
    assert.ok(
        createPage.slice(componentStart, componentStart + 500).includes('serviceContext={serviceContext}'),
        `${component} 必须从独立新建页接收服务上下文创建绑定`,
    );
}

console.log('service governance context verification passed');
