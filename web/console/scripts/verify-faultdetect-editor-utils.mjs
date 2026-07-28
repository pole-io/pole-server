import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/CircuitBreaker/faultDetectEditorUtils.ts');

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

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'faultdetect-editor-utils-'));
const modulePath = path.join(tempDir, 'faultDetectEditorUtils.mjs');
fs.writeFileSync(
  modulePath,
  compiled.outputText.replace(
    /import \{ FaultDetectProtocol \} from 'services\/faultdetect';\n/,
    "const FaultDetectProtocol = { HTTP: 'HTTP', TCP: 'TCP', UDP: 'UDP' };\n",
  ),
);
const utils = await import(pathToFileURL(modulePath));

const baseDraft = {
  name: 'spec-check-faultdetect',
  description: 'fault detect spec check',
  priority: 15,
  metadata: {
    owner: 'codex',
    scenario: 'sample',
  },
  targetService: {
    namespace: 'spec-governance',
    service: 'spec-payment',
  },
  rules: [
    {
      interval: 5,
      timeout: 2,
      port: 0,
      protocol: 'HTTP',
      httpConfig: {
        method: 'GET',
        url: '/healthz',
        headers: [{ key: 'x-probe', value: 'ready' }],
        body: '',
      },
      tcpConfig: { send: '', receive: [], match: 'EXACT' },
      udpConfig: { send: '', receive: [], match: 'EXACT' },
      disable: false,
    },
    {
      interval: 10,
      timeout: 3,
      port: 18080,
      protocol: 'TCP',
      httpConfig: { method: 'GET', url: '', headers: [], body: '' },
      tcpConfig: { send: 'PING\\n', receive: 'PONG\\nOK', match: 'CONTAINS' },
      udpConfig: { send: '', receive: [], match: 'EXACT' },
      disable: false,
    },
  ],
};

const payload = utils.buildFaultDetectSubmitPayload(baseDraft);
assert.equal(payload.name, 'spec-check-faultdetect');
assert.equal(payload.priority, 15);
assert.equal(payload.targetService.namespace, 'spec-governance');
assert.equal(payload.rules.length, 2);
assert.equal(payload.rules[0].port, 0);
assert.equal(payload.rules[1].port, 18080);
assert.deepEqual(payload.rules[1].tcpConfig.receive, ['PONG\\nOK']);
assert.equal(payload.rules[1].tcpConfig.match, undefined, 'submit payload must not include frontend-only match');

const spec = utils.buildFaultDetectPreviewSpec(baseDraft);
assert.equal(spec.kind, 'FaultDetectRule');
assert.equal(spec.metadata.name, 'spec-check-faultdetect');
assert.equal(spec.spec.scope.service, 'spec-payment');
assert.equal(spec.spec.rules[0].portMode, 'INSTANCE_PROTOCOL_PORT');
assert.equal(spec.spec.rules[1].portMode, 'CUSTOM');
assert.equal(spec.spec.rules[1].port, 18080);
assert.equal(spec.spec.rules[1].payload.match, '包含匹配');
assert.equal(spec.spec.rules[1].payload.expectedResponse, 'PONG\\nOK');

const yaml = utils.stringifyFaultDetectSpec(spec, 'yaml');
assert.match(yaml, /kind: FaultDetectRule/);
assert.match(yaml, /portMode: INSTANCE_PROTOCOL_PORT/);
assert.match(yaml, /portMode: CUSTOM/);
assert.match(yaml, /match: 包含匹配/);

const json = utils.stringifyFaultDetectSpec(spec, 'json');
assert.match(json, /"kind": "FaultDetectRule"/);
assert.match(json, /"portMode": "CUSTOM"/);

assert.equal(utils.describeProbeSummary(baseDraft.rules[0]), 'HTTP · 实例协议端口 · 5s 间隔 / 2s 超时 · 启用');
assert.equal(utils.describeProbePort(baseDraft.rules[1]), '指定端口 18080');
assert.deepEqual(utils.validateFaultDetectDraft(baseDraft), []);

const invalidDraft = {
  ...baseDraft,
  name: 'Bad Name',
  rules: [
    {
      ...baseDraft.rules[0],
      port: 70000,
      httpConfig: {
        ...baseDraft.rules[0].httpConfig,
        url: '',
        headers: [{ key: 'x-half', value: '' }],
      },
    },
    {
      ...baseDraft.rules[1],
      interval: 0,
      tcpConfig: { send: '', receive: '', match: 'EXACT' },
    },
  ],
};

assert.deepEqual(utils.validateFaultDetectDraft(invalidDraft).map(item => item.message), [
  '规则名称必须为 kebab-case',
  '探测规则[1] 端口需在 1-65535 之间，留空表示实例协议端口',
  '探测规则[1] HTTP URL 不能为空',
  '探测规则[1] Headers 不允许出现空 key 或空 value',
  '探测规则[2] 间隔必须大于 0',
  '探测规则[2] 接收内容不能为空',
]);

console.log('faultdetect editor utils verification passed');
