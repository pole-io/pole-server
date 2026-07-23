import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');

const sharedSection = read('src/pages/Governance/shared/CollapsibleSection.tsx');
assert.match(sharedSection, /aria-expanded=\{!collapsed\}/, '共享折叠分段必须暴露展开状态');
assert.match(sharedSection, /hidden=\{collapsed\}/, '基础信息折叠必须隐藏但保留正文挂载状态');
assert.match(sharedSection, /collapsed && summary/, '折叠态必须展示规则摘要');
assert.match(sharedSection, /ChevronRightIcon/, '折叠分段必须提供统一方向图标');
assert.doesNotMatch(sharedSection, /scrollIntoView/, '折叠时不能主动滚动容器，否则标题可能被吸入粘性页签下方');

const editors = [
    'src/pages/Governance/Router/CustomRouteEditor.tsx',
    'src/pages/Governance/RateLimit/RateLimitEditor.tsx',
    'src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx',
    'src/pages/Governance/CircuitBreaker/FaultDetectEditor.tsx',
    'src/pages/Governance/LossLess/LossLessEditor.tsx',
    'src/pages/Governance/Router/LaneGroupEdtor.tsx',
    'src/pages/Governance/Security/TrafficGovernanceEditor.tsx',
];

for (const file of editors) {
    const source = read(file);
    assert.match(source, /import CollapsibleSection/, `${file} 必须复用共享基础信息折叠组件`);
    assert.match(source, /basicInfoCollapsed/, `${file} 必须维护基础信息折叠状态`);
    assert.match(source, /<CollapsibleSection[\s\S]*summary=/, `${file} 折叠态必须提供规则摘要`);
}

console.log('governance basic info collapse verification passed');
