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
export const InterfaceProtocol = { HTTP: 'HTTP', GRPC: 'GRPC' };
export const MatchLogic = { AND: 'AND', OR: 'OR' };
export const MatchType = { EXACT: 'EXACT', IN: 'IN', REGEX: 'REGEX' };
export const MatchValueType = { TEXT: 'TEXT' };
`);
fs.writeFileSync(modulePath, compiled.outputText);

const utils = await import(pathToFileURL(modulePath));

const rule = {
  name: 'spec-check-traffic-mirror',
  target_service: { namespace: 'spec-governance', service: 'spec-order' },
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
        { type: 'CALLER_SERVICE', key: 'spec-governance', value: { type: 'EXACT', value: 'spec-gateway', value_type: 'TEXT' } },
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
assert.equal(utils.callerScopeText(utils.defaultMirrorCaller()), '全部服务');
assert.deepEqual(utils.normalizeMirrorCaller({ namespace: 'spec-governance', service: '*' }), utils.defaultMirrorCaller());
assert.equal(utils.callerScopeText({ namespace: 'spec-governance', service: '*' }), '全部服务');

const normalized = utils.normalizeMirrorRules(rule);
assert.equal(normalized.length, 1);
assert.equal(normalized[0].traffic_match_rule.arguments.length, 1);
assert.equal(normalized[0].traffic_match_rule.arguments[0].type, 'HEADER');
assert.equal(normalized[0].traffic_match_rule.randomPercent, undefined);

const allCallerSubmit = utils.buildMirrorRulesForSubmit(normalized, utils.defaultMirrorCaller());
assert.equal(allCallerSubmit[0].traffic_match_rule.arguments.some((item) => item.type === 'CALLER_SERVICE'), false);

const specificCallerSubmit = utils.buildMirrorRulesForSubmit(normalized, caller);
assert.equal(specificCallerSubmit[0].traffic_match_rule.arguments[0].type, 'CALLER_SERVICE');
assert.equal(specificCallerSubmit[0].traffic_match_rule.arguments[0].key, 'spec-governance');
assert.equal(specificCallerSubmit[0].traffic_match_rule.arguments[0].value.value, 'spec-gateway');
assert.equal(specificCallerSubmit[0].traffic_match_rule.randomPercent, undefined);

const preview = utils.buildMirrorPreviewSpec(rule, normalized, caller);
assert.deepEqual(preview.mirror_config.caller, caller);
assert.deepEqual(preview.mirror_config.callee, { namespace: 'spec-governance', service: 'spec-order' });
assert.equal(preview.mirror_config.rules[0].traffic_match_rule.arguments[0].type, 'HEADER');
assert.equal(preview.mirror_config.rules[0].destination.service, 'spec-shadow');
assert.equal(preview.apiVersion, undefined);
assert.equal(preview.kind, undefined);

assert.deepEqual(utils.validateMirrorView(rule, normalized, caller), []);

const invalid = [{
  ...normalized[0],
  api: { ...normalized[0].api, path: { type: 'EXACT', value: '', value_type: 'TEXT' } },
  destination: { namespace: '', service: '' },
  traffic_match_rule: {
    matchMode: 'AND',
    arguments: [{ type: 'HEADER', key: 'x-mirror', value: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
  },
}];
assert.deepEqual(utils.validateMirrorView(rule, invalid, caller).map((item) => item.message), [
  '镜像规则[1]接口路径不能为空',
  '镜像规则[1]镜像目标命名空间不能为空',
  '镜像规则[1]镜像目标服务不能为空',
  '镜像规则[1]存在空匹配值',
]);

const yaml = utils.stringifyMirrorSpec(preview, 'yaml');
assert.match(yaml, /mirror_config:/);
assert.match(yaml, /caller:/);
assert.match(yaml, /callee:/);
assert.doesNotMatch(yaml, /apiVersion:/);

console.log('traffic mirror editor utils verification passed');
