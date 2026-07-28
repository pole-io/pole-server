import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/Router/routeEditorUtils.ts');
const editorPath = path.join(root, 'src/pages/Governance/Router/CustomRouteEditor.tsx');

if (!fs.existsSync(sourcePath)) {
  throw new Error(`missing helper: ${sourcePath}`);
}
if (!fs.existsSync(editorPath)) {
  throw new Error(`missing editor: ${editorPath}`);
}

const source = fs.readFileSync(sourcePath, 'utf8');
const editorSource = fs.readFileSync(editorPath, 'utf8');
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ES2020,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: sourcePath,
});

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'route-editor-utils-'));
const modulePath = path.join(tempDir, 'routeEditorUtils.mjs');
fs.writeFileSync(modulePath, compiled.outputText);
const utils = await import(pathToFileURL(modulePath));

const baseDraft = {
  name: 'spec-check-route',
  enable: true,
  priority: 5,
  description: 'spec RouteRule + CustomRoute sample',
  metadata: [
    { key: 'owner', value: 'codex' },
    { key: 'scenario', value: 'sample' },
  ],
  caller_namespace: 'spec-governance',
  caller_service: 'spec-order',
  callee_namespace: 'spec-governance',
  callee_service: 'spec-payment',
  routing_config: {
    '@type': 'type.googleapis.com/v1.CustomRoute',
    rules: [{
      name: 'vip-route',
      sources: [{
        namespace: 'spec-governance',
        service: 'spec-order',
        arguments: [
          { type: 'HEADER', key: 'x-tenant', value: { type: 'EXACT', value: 'vip', value_type: 'TEXT' } },
          { type: 'QUERY', key: 'region', value: { type: 'IN', value: 'ap-guangzhou,ap-shanghai', value_type: 'TEXT' } },
        ],
      }],
      arguments: { matchMode: 'AND', randomPercent: 0, arguments: [] },
      destinations: [
        { namespace: 'spec-governance', service: 'spec-payment', name: 'payment-blue', isolate: false, weight: 80, labels: { lane: { type: 'EXACT', value: 'blue', value_type: 'TEXT' }, version: { type: 'EXACT', value: 'v2', value_type: 'TEXT' } } },
        { namespace: 'spec-governance', service: 'spec-payment', name: 'payment-stable', isolate: false, weight: 20, labels: { lane: { type: 'EXACT', value: 'stable', value_type: 'TEXT' }, version: { type: 'EXACT', value: 'v1', value_type: 'TEXT' } } },
      ],
    }],
  },
};

assert.equal(utils.getDefaultParamKey('HEADER'), '');
assert.equal(utils.getDefaultParamKey('METHOD'), '');

const changed = utils.withParamTypeDefaultKey(
  { type: 'HEADER', key: 'x-tenant', value: { type: 'EXACT', value: 'vip', value_type: 'TEXT' } },
  'COOKIE',
);
assert.deepEqual(changed, {
  type: 'COOKIE',
  key: '',
  value: { type: 'EXACT', value: 'vip', value_type: 'TEXT' },
});

assert.deepEqual(utils.commaStringToTags('ap-guangzhou, ap-shanghai,,ap-beijing'), [
  'ap-guangzhou',
  'ap-shanghai',
  'ap-beijing',
]);
assert.equal(utils.tagsToCommaString(['ap-guangzhou', ' ap-shenzhen ', '', 'ap-shanghai']), 'ap-guangzhou,ap-shenzhen,ap-shanghai');
assert.equal(utils.isTagInputMatchType('IN'), true);
assert.equal(utils.isTagInputMatchType('NOT_IN'), true);
assert.equal(utils.isTagInputMatchType('EXACT'), false);

const tagDialogStart = editorSource.indexOf('const renderTagDialog');
const tagDialogEnd = editorSource.indexOf('const renderGroupTable');
assert.ok(tagDialogStart > -1 && tagDialogEnd > tagDialogStart, 'missing renderTagDialog block');
const tagDialogSource = editorSource.slice(tagDialogStart, tagDialogEnd);
assert.match(tagDialogSource, /isTagInputMatchType/);
assert.match(tagDialogSource, /<TagInput/);
assert.match(tagDialogSource, /commaStringToTags/);
assert.match(tagDialogSource, /tagsToCommaString/);

assert.equal(utils.getRuleWeightTotal(baseDraft.routing_config.rules[0]), 100);
assert.deepEqual(utils.getWeightStatus(baseDraft.routing_config.rules[0]), {
  total: 100,
  ok: true,
  delta: 0,
  message: '已满 100%',
});

const spec = utils.buildRouteRulePreviewSpec(baseDraft);
assert.equal(spec.name, 'spec-check-route');
assert.equal(spec.enable, true);
assert.equal(spec.priority, 5);
assert.equal(spec.routing_policy, 'RulePolicy');
assert.deepEqual(spec.metadata, {
  owner: 'codex',
  scenario: 'sample',
});
assert.equal(spec.routing_config['@type'], 'type.googleapis.com/v1.CustomRoute');
assert.deepEqual(spec.routing_config.caller, {
  namespace: 'spec-governance',
  service: 'spec-order',
});
assert.deepEqual(spec.routing_config.callee, {
  namespace: 'spec-governance',
  service: 'spec-payment',
});
assert.equal(spec.routing_config.rules[0].arguments.matchMode, 'AND');
assert.equal(spec.routing_config.rules[0].arguments.arguments[0].key, 'x-tenant');
assert.equal(spec.routing_config.rules[0].destinations[0].labels.lane.value, 'blue');
assert.equal(spec.apiVersion, undefined);
assert.equal(spec.kind, undefined);
assert.equal(spec.spec, undefined);

const defaultTypeSpec = utils.buildRouteRulePreviewSpec({
  ...baseDraft,
  routing_config: {
    rules: baseDraft.routing_config.rules,
  },
});
assert.equal(defaultTypeSpec.routing_config['@type'], 'type.googleapis.com/v1.CustomRoute');
assert.deepEqual(utils.textToLabels('lane=blue version=v2'), {
  lane: { type: 'EXACT', value: 'blue', value_type: 'TEXT' },
  version: { type: 'EXACT', value: 'v2', value_type: 'TEXT' },
});

const yaml = utils.stringifyRouteRuleSpec(spec, 'yaml');
assert.match(yaml, /routing_config:/);
assert.match(yaml, /@type: type\.googleapis\.com\/v1\.CustomRoute/);
assert.doesNotMatch(yaml, /apiVersion:/);

assert.deepEqual(utils.validateRouteRuleDraft(baseDraft), []);

const invalidDraft = {
  ...baseDraft,
  name: 'Bad Name',
  routing_config: {
    ...baseDraft.routing_config,
    rules: [{
      ...baseDraft.routing_config.rules[0],
      sources: [{
        ...baseDraft.routing_config.rules[0].sources[0],
        arguments: [{ type: 'HEADER', key: 'x-tenant', value: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
      }],
      destinations: [
        { namespace: 'spec-governance', service: 'spec-payment', name: '', isolate: false, weight: 40, labels: {} },
      ],
    }],
  },
};

assert.deepEqual(utils.validateRouteRuleDraft(invalidDraft).map((item) => item.message), [
  '规则名称必须为 kebab-case',
  '规则[1] 权重合计 40%',
  '规则[1] 存在未命名分组',
  '规则[1] 存在空匹配值',
]);

console.log('route editor utils verification passed');
