import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/Security/trafficSecurityEditorUtils.ts');

if (!fs.existsSync(sourcePath)) {
  throw new Error(`missing helper: ${sourcePath}`);
}

let source = fs.readFileSync(sourcePath, 'utf8');
source = source
  .replace(/from 'services\/traffic_governance'/g, "from './traffic_governance_stub.mjs'")
  .replace(/from 'services\/types'/g, "from './types_stub.mjs'");

const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ES2020,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: sourcePath,
});

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'traffic-security-editor-utils-'));
const modulePath = path.join(tempDir, 'trafficSecurityEditorUtils.mjs');
fs.writeFileSync(path.join(tempDir, 'traffic_governance_stub.mjs'), `
export const TrafficSecurityAction = {
  ALLOW: 'TRAFFIC_SECURITY_ALLOW',
  DENY: 'TRAFFIC_SECURITY_DENY',
};
`);
fs.writeFileSync(path.join(tempDir, 'types_stub.mjs'), `
export const InterfaceProtocol = { HTTP: 'HTTP', GRPC: 'GRPC' };
export const MatchLogic = { AND: 'AND', OR: 'OR' };
export const MatchType = { EXACT: 'EXACT', IN: 'IN', REGEX: 'REGEX' };
export const MatchValueType = { TEXT: 'TEXT' };
`);
fs.writeFileSync(modulePath, compiled.outputText);

const utils = await import(pathToFileURL(modulePath));

const rule = {
  name: 'spec-check-traffic-security-20260612',
  target_service: { namespace: 'spec-governance', service: 'spec-gateway' },
  description: '调用鉴权样例',
  priority: 10,
  enable: true,
  metadata: { owner: 'codex', scenario: 'sample' },
  policies: [{
    action: 'TRAFFIC_SECURITY_ALLOW',
    api: {
      protocol: 'HTTP',
      method: 'GET',
      path: { type: 'IN', value: '/orders', value_type: 'TEXT' },
    },
    traffic_match_rule: {
      matchMode: 'AND',
      arguments: [{
        type: 'HEADER',
        key: 'x-user-type',
        value: { type: 'EXACT', value: 'internal', value_type: 'TEXT' },
      }],
    },
  }, {
    action: 'TRAFFIC_SECURITY_DENY',
    api: {
      protocol: 'HTTP',
      method: 'POST',
      path: { type: 'IN', value: '/admin', value_type: 'TEXT' },
    },
    traffic_match_rule: {
      matchMode: 'AND',
      arguments: [{
        type: 'HEADER',
        key: 'authorization',
        value: { type: 'IN', value: 'deny-', value_type: 'TEXT' },
      }],
    },
  }, {
    action: 'TRAFFIC_SECURITY_ALLOW',
    traffic_match_rule: {
      matchMode: 'OR',
      arguments: [{
        type: 'HEADER',
        key: 'authorization',
        value: { type: 'IN', value: 'Bearer', value_type: 'TEXT' },
      }],
    },
  }],
};

const view = utils.normalizeSecurityViewRules(rule);
assert.deepEqual(view.map(item => item.kind), ['deny', 'allow', 'service']);
assert.deepEqual(view.map(item => item.listType), ['DENY_LIST', 'ALLOW_LIST', 'ALLOW_LIST']);
assert.equal(view[2].interfaces.length, 0);

const preview = utils.buildSecurityPreviewSpec(rule, view);
assert.equal(preview.kind, 'AuthRule');
assert.equal(preview.spec.subRules[0].listType, 'DENY_LIST');
assert.equal(preview.spec.subRules[0].protectedInterfaces[0].path, '/admin');
assert.equal(preview.spec.subRules[1].listType, 'ALLOW_LIST');
assert.equal(preview.spec.subRules[2].protectedInterfaces.length, 0);

const submitPolicies = utils.buildSecurityPoliciesFromView(view);
assert.equal(submitPolicies.length, 3);
assert.equal(submitPolicies[0].action, 'TRAFFIC_SECURITY_DENY');
assert.equal(submitPolicies[1].action, 'TRAFFIC_SECURITY_ALLOW');
assert.equal(submitPolicies[2].api, undefined);

assert.deepEqual(utils.validateSecurityView(rule, view), []);

const invalid = [{
  ...view[0],
  interfaces: [{ protocol: 'HTTP', method: 'GET', path: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
  strategy: { matchMode: 'AND', arguments: [{ type: 'HEADER', key: 'authorization', value: { type: 'EXACT', value: '', value_type: 'TEXT' } }] },
}];
assert.deepEqual(utils.validateSecurityView(rule, invalid).map(item => item.message), [
  '接口路径不能为空',
  '匹配条件的匹配值不能为空',
]);

const yaml = utils.stringifySecuritySpec(preview, 'yaml');
assert.match(yaml, /kind: AuthRule/);
assert.match(yaml, /listType: DENY_LIST/);
assert.match(yaml, /protectedInterfaces: \[\]/);

console.log('traffic security editor utils verification passed');
