import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/RateLimit/rateLimitEditorUtils.ts');

if (!fs.existsSync(sourcePath)) {
  throw new Error(`missing helper: ${sourcePath}`);
}

const source = fs.readFileSync(sourcePath, 'utf8');
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ES2020,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: sourcePath,
});

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ratelimit-editor-utils-'));
const modulePath = path.join(tempDir, 'rateLimitEditorUtils.mjs');
fs.writeFileSync(modulePath, compiled.outputText);
const utils = await import(pathToFileURL(modulePath));

const baseDraft = {
  id: 'rl-1',
  name: 'spec-check-ratelimit',
  namespace: 'spec-governance',
  service: 'spec-gateway',
  type: 'GLOBAL',
  priority: 15,
  disable: false,
  metadata: {
    owner: 'codex',
    scenario: 'sample',
  },
  rules: [
    {
      method: { type: 'EXACT', value: '/api/v1/payments', value_type: 'TEXT' },
      arguments: [
        {
          type: 'HEADER',
          key: 'x-tenant',
          value: { type: 'EXACT', value: 'vip', value_type: 'TEXT' },
        },
      ],
      resource: 'QPS',
      amounts: [
        { validDuration: 1, validDurationUnit: 's', maxAmount: 120 },
        { validDuration: 60, validDurationUnit: 's', maxAmount: 3000 },
      ],
      regex_combine: false,
      action: 'UNIRATE',
      failover: 'FAILOVER_PASS',
      max_queue_delay: 1,
      customResponse: { body: '{"code":429,"msg":"rate limited"}' },
    },
  ],
};

const payload = utils.buildRateLimitSubmitPayload(baseDraft);
assert.equal(payload.name, 'spec-check-ratelimit');
assert.equal(payload.type, 'GLOBAL');
assert.equal(payload.disable, false);
assert.deepEqual(payload.metadata, { owner: 'codex', scenario: 'sample' });
assert.equal(payload.rules[0].action, 'REJECT', 'global rate limit must force reject action');
assert.deepEqual(payload.rules[0].method, { type: 'EXACT', value: '/api/v1/payments', value_type: 'TEXT' });
assert.deepEqual(payload.rules[0].amounts, [
  { validDuration: '1s', maxAmount: 120 },
  { validDuration: '60s', maxAmount: 3000 },
]);
assert.equal(payload.rules[0].customResponse.body, '{"code":429,"msg":"rate limited"}');
assert.equal(payload.apiVersion, undefined);
assert.equal(payload.kind, undefined);
assert.equal(payload.spec, undefined);

const yaml = utils.stringifyRateLimitSpec(payload, 'yaml');
assert.match(yaml, /name: spec-check-ratelimit/);
assert.match(yaml, /type: GLOBAL/);
assert.match(yaml, /rules:/);
assert.match(yaml, /validDuration: 1s/);
assert.doesNotMatch(yaml, /apiVersion:/);

const json = utils.stringifyRateLimitSpec(payload, 'json');
assert.match(json, /"name": "spec-check-ratelimit"/);

assert.equal(utils.describeRuleThreshold(baseDraft.rules[0]), '120 次 / 1s 等 2 窗');
assert.equal(utils.describeRuleSummary(baseDraft.rules[0], baseDraft.type), '1 个接口 · 1 个匹配条件 · 请求数 · 快速失败');

const concurrencyRule = {
  ...baseDraft.rules[0],
  resource: 'CONCURRENCY',
  concurrencyAmount: { maxAmount: 50 },
  amounts: [],
  action: 'REJECT',
};
assert.equal(utils.describeRuleThreshold(concurrencyRule), '50 并发');

assert.deepEqual(utils.validateRateLimitDraft(baseDraft), []);

const invalidDraft = {
  ...baseDraft,
  name: 'Bad Name',
  rules: [
    {
      ...baseDraft.rules[0],
      method: { type: 'EXACT', value: '', value_type: 'TEXT' },
      amounts: [{ validDuration: 1, validDurationUnit: 's', maxAmount: 0 }],
      arguments: [
        {
          type: 'HEADER',
          key: 'x-tenant',
          value: { type: 'EXACT', value: '', value_type: 'TEXT' },
        },
      ],
      customResponse: { body: '{"code":429' },
    },
  ],
};

assert.deepEqual(utils.validateRateLimitDraft(invalidDraft).map((item) => item.message), [
  '规则名称必须为 kebab-case',
  '规则[1] 存在空的接口路径',
  '规则[1] 最大请求数必须大于 0',
  '规则[1] 存在空匹配值',
  '规则[1] 自定义响应必须是合法 JSON',
]);

console.log('ratelimit editor utils verification passed');
