import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const editorFiles = [
  'src/pages/Governance/Router/CustomRouteEditor.tsx',
  'src/pages/Governance/Router/LaneGroupEdtor.tsx',
  'src/pages/Governance/Router/LaneRuleEditor.tsx',
  'src/pages/Governance/RateLimit/RateLimitEditor.tsx',
  'src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx',
  'src/pages/Governance/CircuitBreaker/FaultDetectEditor.tsx',
  'src/pages/Governance/LossLess/LossLessEditor.tsx',
  'src/pages/Governance/Security/TrafficGovernanceEditor.tsx',
];
const styleFiles = [
  'src/pages/Governance/Router/CustomRouteEditor.module.less',
  'src/pages/Governance/Router/LaneGroupEditor.module.less',
  'src/pages/Governance/RateLimit/RateLimitEditor.module.less',
  'src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.module.less',
  'src/pages/Governance/CircuitBreaker/FaultDetectEditor.module.less',
  'src/pages/Governance/LossLess/index.module.less',
  'src/pages/Governance/Security/index.module.less',
];

for (const relativePath of [...editorFiles, ...styleFiles]) {
  const source = fs.readFileSync(path.join(root, relativePath), 'utf8');
  assert.doesNotMatch(source, /实时\s*Spec|实时规则\s*SPEC|specPane|specCard|specToolbar|specToggle|specCode|security-live-spec/i, `${relativePath} 不得保留实时 Spec`);
}

console.log('治理规则编辑器实时 Spec 清理验证通过');
