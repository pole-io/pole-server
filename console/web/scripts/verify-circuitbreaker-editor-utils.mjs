import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/CircuitBreaker/circuitBreakerEditorUtils.ts');

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

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'circuitbreaker-editor-utils-'));
const modulePath = path.join(tempDir, 'circuitBreakerEditorUtils.mjs');
fs.writeFileSync(modulePath, compiled.outputText);
const utils = await import(pathToFileURL(modulePath));

const baseDraft = {
  id: 'cb-1',
  name: 'spec-check-circuitbreak',
  description: 'circuit breaker spec check',
  priority: 15,
  level: 'METHOD',
  metadata: {
    owner: 'codex',
    scenario: 'sample',
  },
  ruleMatcher: {
    source: { namespace: 'spec-governance', service: 'spec-order' },
    destination: {
      namespace: 'spec-governance',
      service: 'spec-payment',
      method: { type: 'EXACT', value: '' },
    },
  },
  subrules: [
    {
      strategies: [
        {
          name: 'payment-error-rate',
          ifaces: [
            {
              protocol: 'HTTP',
              method: 'GET',
              path: { type: 'EXACT', value: '/api/v1/payments', value_type: 'TEXT' },
            },
            {
              protocol: 'HTTP',
              method: 'POST',
              path: { type: 'EXACT', value: '/api/v1/refunds', value_type: 'TEXT' },
            },
          ],
          error_conditions: [
            {
              inputType: 'RET_CODE',
              condition: { type: 'RANGE', value: '500-599' },
            },
          ],
          trigger_conditions: [
            {
              triggerType: 'ERROR_RATE',
              errorCount: 0,
              errorPercent: 50,
              triggerVal: 50,
              interval: 30,
              minimumRequest: 10,
            },
          ],
        },
        {
          name: 'payment-consecutive-error',
          ifaces: [
            {
              protocol: 'HTTP',
              method: 'GET',
              path: { type: 'EXACT', value: '/api/v1/payments', value_type: 'TEXT' },
            },
          ],
          error_conditions: [
            {
              inputType: 'RET_CODE',
              condition: { type: 'RANGE', value: '500-599' },
            },
          ],
          trigger_conditions: [
            {
              triggerType: 'CONSECUTIVE_ERROR',
              errorCount: 5,
              errorPercent: 0,
              triggerVal: 5,
              interval: 30,
              minimumRequest: 10,
            },
          ],
        },
      ],
      max_ejection_percent: 100,
      recoverCondition: {
        sleepWindow: 30,
        consecutiveSuccess: 3,
      },
      faultDetectConfig: {
        enable: true,
      },
      fallbackConfig: {
        enable: true,
        response: {
          code: 503,
          headers: [{ key: 'x-fallback', value: 'true' }],
          body: 'fallback',
        },
      },
    },
  ],
};

const payload = utils.buildCircuitBreakerSubmitPayload(baseDraft);
assert.equal(payload.name, 'spec-check-circuitbreak');
assert.equal(payload.level, 'METHOD');
assert.deepEqual(payload.metadata, { owner: 'codex', scenario: 'sample' });
assert.equal(payload.block_configs.length, 3, 'two interfaces in one strategy plus one strategy must flatten to three policies');
assert.equal(payload.block_configs[0].block_config.name, 'payment-error-rate');
assert.equal(payload.block_configs[0].block_config.api.path.value, '/api/v1/payments');
assert.equal(payload.block_configs[1].block_config.api.path.value, '/api/v1/refunds');
assert.equal(payload.block_configs[2].block_config.name, 'payment-consecutive-error');
assert.equal(payload.block_configs[0].recoverCondition.sleepWindow, 30);
assert.equal(payload.block_configs[0].fallbackConfig.response.code, 503);
assert.equal(payload.block_configs[0].max_ejection_percent, 100);
assert.equal(payload.apiVersion, undefined);
assert.equal(payload.kind, undefined);
assert.equal(payload.spec, undefined);

const yaml = utils.stringifyCircuitBreakerSpec(payload, 'yaml');
assert.match(yaml, /name: spec-check-circuitbreak/);
assert.match(yaml, /block_configs:/);
assert.match(yaml, /block_config:/);
assert.match(yaml, /payment-error-rate/);
assert.doesNotMatch(yaml, /apiVersion:/);

const json = utils.stringifyCircuitBreakerSpec(payload, 'json');
assert.match(json, /"name": "spec-check-circuitbreak"/);
assert.match(json, /"block_configs"/);

assert.equal(utils.describeSubRuleSummary(baseDraft.subrules[0]), '2 个熔断策略 · 熔断 30s · 降级开');
assert.equal(utils.describeStrategySummary(baseDraft.subrules[0].strategies[0]), '2 个接口 · 1 个错误条件 · 1 个触发条件');

assert.deepEqual(utils.validateCircuitBreakerDraft(baseDraft), []);

const invalidDraft = {
  ...baseDraft,
  name: 'Bad Name',
  subrules: [
    {
      ...baseDraft.subrules[0],
      recoverCondition: { sleepWindow: 0 },
      strategies: [
        {
          ...baseDraft.subrules[0].strategies[0],
          ifaces: [{ protocol: 'HTTP', method: 'GET', path: { type: 'EXACT', value: '', value_type: 'TEXT' } }],
          error_conditions: [{ inputType: 'RET_CODE', condition: { type: 'RANGE', value: '' } }],
          trigger_conditions: [{ triggerType: 'ERROR_RATE', errorPercent: 120, errorCount: 0, interval: 30, minimumRequest: 10 }],
        },
      ],
    },
  ],
};

assert.deepEqual(utils.validateCircuitBreakerDraft(invalidDraft).map((item) => item.message), [
  '规则名称必须为 kebab-case',
  '子规则[1]·策略[1] 存在空的接口路径',
  '子规则[1]·策略[1] 存在空错误判断值',
  '子规则[1]·策略[1] 比例类触发阈值必须在 0-100 之间',
  '子规则[1] 恢复策略熔断时长必须大于 0',
]);

const normalized = utils.createCircuitBreakerDraftFromRule({
  ...payload,
  block_configs: payload.block_configs.slice(0, 1),
});
assert.equal(normalized.subrules.length, 1);
assert.equal(normalized.subrules[0].strategies.length, 1);
assert.equal(normalized.subrules[0].strategies[0].ifaces[0].path.value, '/api/v1/payments');

console.log('circuitbreaker editor utils verification passed');
