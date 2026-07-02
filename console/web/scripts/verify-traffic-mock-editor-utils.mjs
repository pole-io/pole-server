import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/Security/trafficMockEditorUtils.ts');

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

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'traffic-mock-editor-utils-'));
const modulePath = path.join(tempDir, 'trafficMockEditorUtils.mjs');
fs.writeFileSync(path.join(tempDir, 'traffic_governance_stub.mjs'), '');
fs.writeFileSync(path.join(tempDir, 'types_stub.mjs'), `
export const InterfaceProtocol = { HTTP: 'HTTP', GRPC: 'GRPC' };
export const MatchLogic = { AND: 'AND', OR: 'OR' };
export const MatchType = { EXACT: 'EXACT', IN: 'IN', REGEX: 'REGEX' };
export const MatchValueType = { TEXT: 'TEXT' };
`);
fs.writeFileSync(modulePath, compiled.outputText);

const utils = await import(pathToFileURL(modulePath));

const rule = {
  name: 'spec-check-traffic-mock-20260612',
  target_service: { namespace: 'spec-governance', service: 'spec-gateway' },
  description: '流量 Mock 样例',
  priority: 30,
  enable: true,
  metadata: { owner: 'codex', scenario: 'sample' },
  rules: [{
    api: {
      protocol: 'HTTP',
      method: 'GET',
      path: { type: 'EXACT', value: '/mock/orders', value_type: 'TEXT' },
    },
    traffic_match_rule: {
      matchMode: 'AND',
      randomPercent: 20,
      arguments: [
        {
          type: 'CALLER_SERVICE',
          key: 'source-ns',
          value: { type: 'EXACT', value: 'checkout', value_type: 'TEXT' },
        },
        {
          type: 'HEADER',
          key: 'x-mock',
          value: { type: 'EXACT', value: 'orders', value_type: 'TEXT' },
        },
      ],
    },
    response: {
      status_code: 201,
      code: 'MOCK_OK',
      headers: { 'content-type': 'application/json' },
      body: '{"orderId":"mock-001","status":"PAID"}',
      message: 'legacy message',
    },
    mock_percent: 10,
    delay: '0.125s',
  }],
};

const normalized = utils.normalizeMockRules(rule);
assert.equal(normalized.length, 1);
assert.equal(normalized[0].response.code, 'MOCK_OK');
assert.equal(normalized[0].response.message, undefined);
assert.equal(normalized[0].response.status_code, undefined);
assert.equal(normalized[0].mock_percent, 100);
assert.equal(normalized[0].traffic_match_rule.randomPercent, undefined);
assert.equal(normalized[0].traffic_match_rule.arguments.length, 1);
assert.equal(normalized[0].traffic_match_rule.arguments[0].type, 'HEADER');
assert.equal(utils.durationToMs(normalized[0].delay), 125);
assert.deepEqual(utils.extractMockCaller(rule), { namespace: 'source-ns', service: 'checkout' });
assert.equal(utils.mockCallerScopeText(utils.defaultMockCaller()), '全部服务');
assert.equal(utils.mockCallerScopeText({ namespace: 'source-ns', service: 'checkout' }), 'source-ns/checkout');
assert.deepEqual(utils.normalizeMockCaller({ namespace: 'source-ns', service: '*' }), utils.defaultMockCaller());
assert.equal(utils.mockCallerScopeText({ namespace: 'source-ns', service: '*' }), '全部服务');

const preview = utils.buildMockPreviewSpec(rule, normalized, { namespace: 'source-ns', service: 'checkout' });
assert.equal(preview.kind, 'MockRule');
assert.deepEqual(preview.spec.scope.caller, { namespace: 'source-ns', service: 'checkout' });
assert.deepEqual(preview.spec.scope.callee, { namespace: 'spec-governance', service: 'spec-gateway' });
assert.equal(preview.spec.rules[0].name, 'mock-subrule-1');
assert.equal(preview.spec.rules[0].interfaces[0].path, '/mock/orders');
assert.equal(preview.spec.rules[0].match.conditions.length, 1);
assert.equal(preview.spec.rules[0].match.conditions[0].param, 'HEADER');
assert.equal(preview.spec.rules[0].match.conditions[0].value, 'orders');
assert.equal(preview.spec.rules[0].response.statusCode, 'MOCK_OK');
assert.equal(preview.spec.rules[0].response.delayMs, 125);
assert.equal(JSON.stringify(preview).includes('ratioPercent'), false);
assert.equal(JSON.stringify(preview).includes('reason'), false);
assert.equal(JSON.stringify(preview).includes('CALLER_SERVICE'), false);

const submit = utils.buildMockRulesForSubmit(normalized, { namespace: 'source-ns', service: 'checkout' });
assert.equal(submit[0].mock_percent, 100);
assert.equal(submit[0].api, undefined);
assert.equal(submit[0].apis[0].path.value, '/mock/orders');
assert.equal(submit[0].traffic_match_rule.randomPercent, undefined);
assert.equal(submit[0].traffic_match_rule.arguments[0].type, 'CALLER_SERVICE');
assert.equal(submit[0].traffic_match_rule.arguments[0].key, 'source-ns');
assert.equal(submit[0].traffic_match_rule.arguments[0].value.value, 'checkout');
assert.deepEqual(submit[0].response, {
  code: 'MOCK_OK',
  headers: { 'content-type': 'application/json' },
  body: '{"orderId":"mock-001","status":"PAID"}',
});

const allCallerSubmit = utils.buildMockRulesForSubmit(normalized, utils.defaultMockCaller());
assert.equal(allCallerSubmit[0].traffic_match_rule.arguments.some(item => item.type === 'CALLER_SERVICE'), false);

assert.deepEqual(utils.validateMockView(rule, normalized, { namespace: 'source-ns', service: 'checkout' }), []);
assert.deepEqual(utils.validateMockView(rule, normalized, { namespace: 'source-ns', service: '' }).map(item => item.message), [
  'Mock 主调方必须同时选择命名空间和服务',
]);

const invalid = [{
  ...normalized[0],
  apis: [{ ...normalized[0].apis[0], path: { ...normalized[0].apis[0].path, value: '' } }],
  traffic_match_rule: {
    matchMode: 'AND',
    arguments: [{ type: 'HEADER', key: 'x-mock', value: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
  },
  response: {
    code: '',
    headers: { '': 'application/json', 'content-type': 'application/json' },
    body: '{bad-json',
  },
}];
assert.deepEqual(utils.validateMockView(rule, invalid, utils.defaultMockCaller()).map(item => item.message), [
  'Mock 子规则[1] 接口路径不能为空',
  'Mock 子规则[1] 响应 Code 不能为空',
  'Mock 子规则[1] 流量匹配条件存在空匹配值',
  'Mock 子规则[1] 响应头存在空 Header 名',
  'Mock 子规则[1] Content-Type 为 JSON 时，响应正文必须是合法 JSON',
]);

const yaml = utils.stringifyMockSpec(preview, 'yaml');
assert.match(yaml, /kind: MockRule/);
assert.match(yaml, /caller:/);
assert.match(yaml, /callee:/);
assert.match(yaml, /statusCode: MOCK_OK/);
assert.doesNotMatch(yaml, /ratioPercent/);
assert.doesNotMatch(yaml, /reason/);
assert.doesNotMatch(yaml, /CALLER_SERVICE/);

console.log('traffic mock editor utils verification passed');
