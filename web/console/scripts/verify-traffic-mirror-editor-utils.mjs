import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/Security/trafficMirrorEditorUtils.ts');

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

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'traffic-mirror-editor-utils-'));
const modulePath = path.join(tempDir, 'trafficMirrorEditorUtils.mjs');
fs.writeFileSync(path.join(tempDir, 'traffic_governance_stub.mjs'), '');
fs.writeFileSync(path.join(tempDir, 'types_stub.mjs'), `
export const InterfaceProtocol = { HTTP: 'HTTP', GRPC: 'GRPC', DUBBO: 'DUBBO' };
export const MatchLogic = { AND: 'AND', OR: 'OR' };
export const MatchType = { EXACT: 'EXACT', IN: 'IN', REGEX: 'REGEX' };
export const MatchValueType = { TEXT: 'TEXT' };
`);
fs.writeFileSync(modulePath, compiled.outputText);

const utils = await import(pathToFileURL(modulePath));

const rule = {
  name: 'spec-check-traffic-mirror',
  caller: { namespace: 'spec-governance', service: 'spec-gateway' },
  callee: { namespace: 'spec-governance', service: 'spec-order' },
  description: '镜像样例',
  priority: 10,
  enable: true,
  metadata: { owner: 'codex' },
  rules: [{
    api: {
      protocol: 'HTTP',
      method: 'GET',
      path: { type: 'EXACT', value: '/orders', value_type: 'TEXT' },
    },
    traffic_match_rule: {
      matchMode: 'AND',
      arguments: [
        { type: 'CALLER_SERVICE', key: 'legacy', value: { type: 'EXACT', value: 'legacy-gateway', value_type: 'TEXT' } },
        { type: 'HEADER', key: 'x-mirror', value: { type: 'EXACT', value: 'true', value_type: 'TEXT' } },
      ],
    },
    destination: {
      namespace: 'spec-governance',
      service: 'spec-shadow',
      labels: {
        version: { type: 'EXACT', value: 'shadow', value_type: 'TEXT' },
      },
    },
    mirror_percent: 30,
    duration: '300s',
    disable: false,
  }],
};

const caller = utils.extractMirrorCaller(rule);
assert.deepEqual(caller, { namespace: 'spec-governance', service: 'spec-gateway' });
assert.equal(utils.callerScopeText(caller), 'spec-governance/spec-gateway');
assert.equal(utils.callerScopeText(utils.defaultMirrorCaller()), '全部命名空间/全部服务');
assert.deepEqual(utils.normalizeMirrorCaller({ namespace: 'spec-governance', service: '*' }), utils.defaultMirrorCaller());
assert.equal(utils.mirrorScopeLabel({ namespace: 'spec-governance', service: '*' }), 'spec-governance/全部服务');
assert.deepEqual(utils.normalizeMirrorCallee(rule), { namespace: 'spec-governance', service: 'spec-order' });

const normalized = utils.normalizeMirrorRules(rule);
assert.equal(normalized.length, 1);
assert.equal(normalized[0].interfaces.length, 1);
assert.equal(normalized[0].interfaces[0].path.value, '/orders');
assert.equal(normalized[0].traffic_match_rule.arguments.length, 1);
assert.equal(normalized[0].traffic_match_rule.arguments[0].type, 'HEADER');
assert.equal(normalized[0].traffic_match_rule.randomPercent, undefined);

const multiInterfaceRule = [{
  ...normalized[0],
  interfaces: [
    normalized[0].interfaces[0],
    { protocol: 'DUBBO', method: '*', path: { type: 'REGEX', value: '^/shadow/.*', value_type: 'TEXT' } },
  ],
}];
const submit = utils.buildMirrorRulesForSubmit(multiInterfaceRule);
assert.equal(submit.length, 1);
assert.equal(submit[0].interfaces, undefined);
assert.equal(submit[0].api, undefined);
assert.equal(submit[0].apis.length, 2);
assert.equal(submit[0].apis[1].protocol, 'DUBBO');
assert.equal(submit[0].traffic_match_rule.arguments.some((item) => item.type === 'CALLER_SERVICE'), false);
assert.equal(submit[0].duration, undefined);

const preview = utils.buildMirrorPreviewSpec(rule, multiInterfaceRule, caller);
assert.equal(preview.apiVersion, 'governance.pole.io/v1');
assert.equal(preview.kind, 'MirrorRule');
assert.deepEqual(preview.spec.serviceRange.caller, caller);
assert.deepEqual(preview.spec.serviceRange.callee, { namespace: 'spec-governance', service: 'spec-order' });
assert.equal(preview.spec.rules[0].interfaces.length, 2);
assert.equal(preview.spec.rules[0].trafficLabels.labels[0].key, 'x-mirror');
assert.equal(preview.spec.rules[0].mirror.target.service, 'spec-shadow');
assert.equal(preview.spec.rules[0].mirror.duration, undefined);

assert.deepEqual(utils.validateMirrorView(rule, normalized, caller), []);

const invalid = [{
  ...normalized[0],
  interfaces: [{ ...normalized[0].interfaces[0], path: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
  destination: { namespace: '', service: '' },
  traffic_match_rule: {
    matchMode: 'AND',
    arguments: [{ type: 'HEADER', key: '', value: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
  },
}];
assert.deepEqual(utils.validateMirrorView(rule, invalid, caller).map((item) => item.message), [
  '镜像子规则[1]存在空接口路径',
  '镜像子规则[1]存在空标签键',
  '镜像子规则[1]存在空标签值',
  '镜像子规则[1]镜像目标命名空间不能为空',
  '镜像子规则[1]镜像目标服务不能为空',
]);

const yaml = utils.stringifyMirrorSpec(preview, 'yaml');
assert.match(yaml, /apiVersion:/);
assert.match(yaml, /kind: MirrorRule/);
assert.match(yaml, /serviceRange:/);
assert.match(yaml, /trafficLabels:/);
assert.match(yaml, /mirror:/);
assert.doesNotMatch(yaml, /mirror_config:/);
assert.doesNotMatch(yaml, /duration:/);

console.log('traffic mirror editor utils verification passed');
