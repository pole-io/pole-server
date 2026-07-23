import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8');

const shared = read('src/pages/Governance/shared/TrafficMatchConditionEditor.tsx');
const sharedStyle = read('src/pages/Governance/shared/TrafficMatchConditionEditor.module.less');
const types = read('src/services/types.ts');
const route = read('src/pages/Governance/Router/CustomRouteEditor.tsx');
const rateLimit = read('src/pages/Governance/RateLimit/RateLimitEditor.tsx');
const traffic = read('src/pages/Governance/Security/TrafficGovernanceEditor.tsx');
const lane = read('src/pages/Governance/Router/LaneGroupEdtor.tsx');
const routeUtils = read('src/pages/Governance/Router/routeEditorUtils.ts');
const rateLimitUtils = read('src/pages/Governance/RateLimit/rateLimitEditorUtils.ts');
const securityUtils = read('src/pages/Governance/Security/trafficSecurityEditorUtils.ts');
const laneUtils = read('src/pages/Governance/Router/laneEditorUtils.ts');
const rateLimitService = read('src/services/ratelimit.ts');
const laneService = read('src/services/lane.ts');

assert.match(types, /'TEXT': '固定值'/);
assert.match(types, /'PARAMETER': '请求参数'/);
assert.doesNotMatch(types, /VARIABLE|运行变量/);
assert.doesNotMatch(shared, /VARIABLE|运行变量|变量名/);
assert.doesNotMatch(rateLimitService, /MatchValueType\.VARIABLE|2:\s*MatchValueType/);
assert.doesNotMatch(laneService, /MatchValueType\.VARIABLE|2:\s*MatchValueType/);
assert.doesNotMatch(route, /MatchValueType\.VARIABLE/);
assert.match(shared, /valueType\?: string/);
assert.match(shared, /值来源/);
assert.match(shared, /采集该键的请求值/);
assert.match(shared, /Proxyless SDK 已支持，xDS 当前不消费该动态语义/);
assert.match(shared, /nextValueType === MatchValueType\.PARAMETER/);
assert.match(sharedStyle, /\.tableHeaderRow,[\s\S]*?\.conditionRow\s*\{[\s\S]*?display:\s*grid;/);
assert.match(sharedStyle, /grid-template-columns:.*minmax\(196px, 1\.3fr\).*48px/s);
assert.match(sharedStyle, /@container \(max-width: 920px\)[\s\S]*?\.tableHeaderRow\s*\{[\s\S]*?display:\s*none;/);
assert.match(sharedStyle, /\.tableField :global\(\.fui-Combobox\)/);
assert.match(sharedStyle, /max-width: 100%;[\s\S]*min-width: 0;[\s\S]*width: 100%/);

for (const [name, source] of Object.entries({ route, rateLimit, traffic, lane })) {
  assert.match(source, /valueType:/, `${name} must map protobuf value_type into the shared row`);
  assert.match(source, /row\.valueType/, `${name} must persist the shared row value type`);
}

assert.match(routeUtils, /arg\.value\?\.value_type !== 'PARAMETER'/);
assert.match(rateLimitUtils, /arg\.value\?\.value_type !== 'PARAMETER'/);
assert.match(securityUtils, /condition\.value\?\.value_type !== MatchValueType\.PARAMETER/);
assert.match(laneUtils, /valueType: condition\?\.valueType/);
assert.match(laneUtils, /item\.valueType !== 'PARAMETER'/);

assert.match(route, /使用同名请求参数/);
assert.match(route, /<Col span=\{3\}>[\s\S]*请输入标签键/);
assert.match(route, /<Col span=\{1\}>[\s\S]*删除标签/);
assert.match(route, /value_type: value\.value_type \|\| MatchValueType\.TEXT/);
assert.match(route, /valueType === MatchValueType\.PARAMETER \? '' : tag\.value\.value/);

console.log('governance request value type verification passed');
