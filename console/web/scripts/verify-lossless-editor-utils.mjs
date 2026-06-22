import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/LossLess/losslessEditorUtils.ts');

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

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'lossless-editor-utils-'));
const modulePath = path.join(tempDir, 'losslessEditorUtils.mjs');
fs.writeFileSync(modulePath, compiled.outputText);
const utils = await import(pathToFileURL(modulePath));

const baseDraft = utils.normalizeLosslessRuleDraft({
  id: 'lossless-1',
  namespace: 'spec-governance',
  service: 'spec-order',
  metadata: { owner: 'codex', scenario: 'sample' },
  lossless_online: {
    delay_register: {
      enable: true,
      strategy: 'DELAY_BY_HEALTH_CHECK',
      health_check_protocol: 'TCP',
      health_check_interval_second: '5s',
      payload: {
        request: 'PING\\n',
        expectedResponse: 'PONG',
        match: 'CONTAINS',
      },
    },
    warmup: {
      enable: true,
      interval_second: 60,
      enable_overload_protection: true,
      overload_protection_threshold: 80,
      curvature: 5,
    },
  },
  lossless_offline: {
    enable: true,
    interval_second: 30,
  },
});

assert.equal(baseDraft.lossless_online.delay_register.health_check_protocol, 'TCP');
assert.equal(baseDraft.lossless_online.delay_register.payload.match, 'CONTAINS');
assert.deepEqual(utils.validateLosslessDraft(baseDraft), []);

const preview = utils.buildLosslessPreviewSpec(baseDraft);
assert.equal(preview.kind, 'LosslessRule');
assert.equal(preview.metadata.name, 'lossless-1');
assert.deepEqual(preview.metadata.tags, ['owner:codex', 'scenario:sample']);
assert.equal(preview.spec.scope.namespace, 'spec-governance');
assert.equal(preview.spec.online.delayedRegistration.strategy, 'PROBE_DELAY');
assert.equal(preview.spec.online.delayedRegistration.healthCheck.protocol, 'TCP');
assert.equal(preview.spec.online.delayedRegistration.healthCheck.payload.expectedResponse, 'PONG');
assert.equal(preview.spec.online.warmup.curveValue, '5');
assert.equal(preview.spec.offline.waitIntervalSec, 30);

const submit = utils.buildLosslessSubmitPayload(baseDraft);
assert.equal(submit.lossless_online.delay_register.strategy, 'DELAY_BY_HEALTH_CHECK');
assert.equal(submit.lossless_online.delay_register.health_check_protocol, 'TCP');
assert.equal(submit.lossless_online.delay_register.health_check_method, undefined);
assert.equal(submit.lossless_online.delay_register.payload, undefined, 'submit payload must not include frontend-only TCP/UDP payload');
assert.equal(submit.lossless_online.warmup.curvature, 5);

const curvePoints = utils.buildWarmupCurvePoints(5);
assert.deepEqual(
  curvePoints.filter(item => [0, 25, 50, 75, 100].includes(item.progress)).map(item => item.weight),
  [0, 1, 4, 24, 100],
);

const yaml = utils.stringifyLosslessSpec(preview, 'yaml');
assert.match(yaml, /kind: LosslessRule/);
assert.match(yaml, /strategy: PROBE_DELAY/);
assert.match(yaml, /match: 包含匹配/);

const timeDelayDraft = utils.normalizeLosslessRuleDraft({
  namespace: 'spec-governance',
  service: 'spec-order',
  lossless_online: {
    delay_register: {
      enable: true,
      strategy: 'DELAY_BY_TIME',
      interval_second: 15,
    },
    warmup: { enable: false },
  },
  lossless_offline: { enable: false },
});
assert.equal(utils.buildLosslessPreviewSpec(timeDelayDraft).spec.online.delayedRegistration.strategy, 'TIME_DELAY');
assert.equal(utils.buildLosslessPreviewSpec(timeDelayDraft).spec.online.delayedRegistration.delaySec, 15);

const invalidDraft = {
  ...baseDraft,
  namespace: '',
  lossless_online: {
    ...baseDraft.lossless_online,
    delay_register: {
      ...baseDraft.lossless_online.delay_register,
      health_check_protocol: 'UDP',
      health_check_interval: 0,
      payload: {
        ...baseDraft.lossless_online.delay_register.payload,
        response: '',
      },
    },
    warmup: {
      ...baseDraft.lossless_online.warmup,
      interval: 0,
      curvature: 6,
    },
  },
  lossless_offline: {
    enable: true,
    interval: 0,
  },
};
assert.deepEqual(utils.validateLosslessDraft(invalidDraft).map(item => item.message), [
  '主调命名空间不能为空',
  '检查间隔必须大于 0',
  '响应匹配内容不能为空',
  '预热启用时预热时长必须大于 0',
  '预热曲线值必须是 1～5 的整数',
  '无损下线间隔必须大于 0',
]);

console.log('lossless editor utils verification passed');
