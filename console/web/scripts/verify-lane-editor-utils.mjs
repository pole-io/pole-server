import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/Router/laneEditorUtils.ts');
const editorPath = path.join(root, 'src/pages/Governance/Router/LaneGroupEdtor.tsx');
const stylePath = path.join(root, 'src/pages/Governance/Router/LaneGroupEditor.module.less');
const laneGroupTablePath = path.join(root, 'src/pages/Governance/Router/LaneGroupTable.tsx');
const workbenchPath = path.join(root, 'src/pages/Governance/Workbench/index.tsx');

if (!fs.existsSync(sourcePath)) {
  throw new Error(`missing helper: ${sourcePath}`);
}
if (!fs.existsSync(editorPath)) {
  throw new Error(`missing editor: ${editorPath}`);
}
if (!fs.existsSync(stylePath)) {
  throw new Error(`missing style: ${stylePath}`);
}
if (!fs.existsSync(laneGroupTablePath)) {
  throw new Error(`missing lane table: ${laneGroupTablePath}`);
}
if (!fs.existsSync(workbenchPath)) {
  throw new Error(`missing workbench: ${workbenchPath}`);
}

const source = fs.readFileSync(sourcePath, 'utf8');
const editorSource = fs.readFileSync(editorPath, 'utf8');
const styleSource = fs.readFileSync(stylePath, 'utf8');
const laneGroupTableSource = fs.readFileSync(laneGroupTablePath, 'utf8');
const workbenchSource = fs.readFileSync(workbenchPath, 'utf8');
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ES2020,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: sourcePath,
});

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'lane-editor-utils-'));
const modulePath = path.join(tempDir, 'laneEditorUtils.mjs');
fs.writeFileSync(modulePath, compiled.outputText);
const utils = await import(pathToFileURL(modulePath));

const group = utils.normalizeLaneGroupDraft({
  name: 'spec-check-lane-group',
  description: 'spec LaneGroup + LaneRule sample',
  entries: [
    { kind: 'gateway', ns: 'spec-governance', svc: 'spec-gateway' },
    { kind: 'app', ns: 'spec-governance', svc: 'spec-order' },
  ],
  selected: ['spec-order (spec-governance)', 'spec-payment (spec-governance)'],
  lanes: [{
    name: 'shadow-lane',
    on: true,
    open: false,
    laneValue: 'shadow',
    relation: 'AND',
    matchRatio: 30,
    conditions: [
      { type: 'HEADER', key: 'x-shadow', match: '完全匹配', value: 'true' },
    ],
    serviceTags: [
      { service: 'spec-order' },
      { service: 'spec-payment' },
    ],
  }],
});

assert.equal(utils.LANE_TRAFFIC_TAG_KEY, 'X-Lattice-Traffic-Lane');
assert.equal(group.lanes[0].matchRatio, 30);
assert.equal(group.lanes[0].serviceTags.length, 2);
assert.deepEqual(utils.validateLaneGroupDraft(group), []);
assert.deepEqual(utils.validateLaneRulesDraft(group), []);

const laneSpec = utils.buildLaneRulePreviewSpec(group);
assert.equal(laneSpec.apiVersion, 'governance.pole.io/v1');
assert.equal(laneSpec.kind, 'LaneRule');
assert.equal(laneSpec.metadata.name, 'spec-check-lane-group');
assert.equal(laneSpec.spec.lanes[0].match.ratioPercent, 30);
assert.deepEqual(laneSpec.spec.lanes[0].services, ['spec-order', 'spec-payment']);
assert.deepEqual(laneSpec.spec.lanes[0].laneTag, {
  key: 'X-Lattice-Traffic-Lane',
  value: 'shadow',
});

const groupSpec = utils.buildLaneGroupPreviewSpec(group);
assert.equal(groupSpec.kind, 'LaneGroup');
assert.equal(groupSpec.spec.laneGroup.entries[0].type, 'micro-gateway');
assert.deepEqual(groupSpec.spec.laneGroup.services, ['spec-order', 'spec-payment']);

const topology = utils.buildLaneTopology(group, group.lanes[0]);
assert.equal(topology.entryLabel, 'spec-gateway');
assert.equal(topology.hitRatio, 30);
assert.equal(topology.fallbackRatio, 70);
assert.deepEqual(topology.visibleServices, ['spec-order', 'spec-payment']);

const fullRatioTopology = utils.buildLaneTopology(group, { ...group.lanes[0], matchRatio: 100 });
assert.equal(fullRatioTopology.fallbackLabel, '未命中');
assert.equal(fullRatioTopology.fallbackRatio, 0);

const yaml = utils.stringifyLaneSpec(laneSpec, 'yaml');
assert.match(yaml, /kind: LaneRule/);
assert.match(yaml, /ratioPercent: 30/);
assert.match(yaml, /key: "X-Lattice-Traffic-Lane"/);
assert.match(yaml, /value: "true"/);

const invalid = utils.normalizeLaneGroupDraft({
  name: '',
  entries: [],
  selected: ['spec-order (spec-governance)'],
  lanes: [{
    name: '',
    laneValue: '',
    matchRatio: 120,
    conditions: [{ type: 'HEADER', key: '', match: '完全匹配', value: '' }],
    serviceTags: [{ service: 'not-in-group' }],
  }],
});
assert.deepEqual(utils.validateLaneGroupDraft(invalid).map(item => item.message), [
  '泳道组名称不能为空',
  '至少配置 1 个泳道组入口',
]);
assert.deepEqual(utils.validateLaneRulesDraft(invalid).map(item => item.message), [
  '泳道[1] 名称不能为空',
  '泳道[1] 命中后放量比例必须在 0–100 之间',
  '泳道[1] 存在空匹配条件',
  '泳道[1] 泳道标签 Value 不能为空',
  '泳道[1] 引用了泳道组之外的服务',
]);

assert.match(editorSource, /label="组信息" value="group"/);
assert.match(editorSource, /label="泳道" value="lane"/);
assert.match(editorSource, /label="版本" value="version"/);
assert.match(editorSource, /label="审计" value="audit"/);
assert.match(editorSource, /React\.useState<LanePage>\('group'\)/);
assert.match(editorSource, /renderStepTitle\(1, '基础信息'/);
assert.match(editorSource, /renderStepTitle\(2, '泳道组入口'/);
assert.match(editorSource, /renderStepTitle\(3, '组内服务'/);
assert.match(editorSource, /const getEntryServicePool = \(entry: LaneDraftEntry\)/);
assert.match(editorSource, /const getNamespaceOptions = \(items: SimpleService\[\]\)/);
assert.match(editorSource, /const getServiceOptions = \(items: SimpleService\[\], namespace: string/);
assert.match(editorSource, /placeholder="选择命名空间"/);
assert.match(editorSource, /placeholder="选择服务"/);
assert.match(editorSource, /setServiceDraft\(\{ ns: value as string, svc: '' \}\)/);
assert.match(editorSource, /className=\{`\$\{styles\.prdTableRow\} \$\{styles\.serviceDraftRow\}`\}/);
assert.doesNotMatch(editorSource, /placeholder="选择命名空间 \/ 服务"/);
assert.doesNotMatch(editorSource, /placeholder="\+ 添加服务"/);
assert.match(editorSource, /className=\{styles\.laneExpandIcon\}/);
assert.match(editorSource, /className=\{styles\.laneColorDot\}/);
assert.match(editorSource, /className=\{styles\.fixedTag\}/);
assert.match(editorSource, /TrafficMatchConditionEditor/);
assert.match(editorSource, /extraControl=\{\(/);
assert.match(editorSource, /className=\{styles\.laneRatioControl\}/);
assert.doesNotMatch(editorSource, /renderLaneMatchModeSwitch/);
assert.doesNotMatch(editorSource, /className=\{styles\.laneMatchRows\}/);
assert.match(editorSource, /className=\{styles\.laneServiceSection\}/);
assert.match(editorSource, /className=\{styles\.laneResultHint\}/);
assert.match(editorSource, /className=\{styles\.addLaneCardButton\}/);
assert.match(editorSource, /renderLaneTopology/);
assert.match(editorSource, /LANE_TRAFFIC_TAG_KEY/);
assert.doesNotMatch(editorSource, /renderField\('固定标签 Key'/);
assert.match(styleSource, /\.laneName\s*\{[\s\S]*?font-size:\s*14px;/);
assert.match(styleSource, /\.subPanelTitle\s*\{[\s\S]*?font-size:\s*14px;/);
assert.match(styleSource, /\.subPanelHint\s*\{[\s\S]*?font-size:\s*12px;/);
assert.match(styleSource, /\.laneRatioControl\s*\{/);
assert.match(styleSource, /\.serviceDraftRow\s*\{/);
assert.match(styleSource, /\.laneRatioControl\s*\{[\s\S]*?flex:\s*0 0 auto;/);
assert.match(styleSource, /\.laneRatioControl\s+:global\(\.t-input-adornment\)\s*\{[\s\S]*?min-width:\s*104px;/);
assert.match(styleSource, /\.laneRatioControl\s+:global\(\.t-input\),\s*\n\.laneRatioControl\s+:global\(\.t-input__wrap\)\s*\{[\s\S]*?min-width:\s*0;/);
assert.match(styleSource, /\.formPane\s*\{[\s\S]*?overflow:\s*auto;/);
assert.match(styleSource, /:global\(\.t-tabs\)\s*\{[\s\S]*?min-height:\s*100%;/);
assert.match(styleSource, /:global\(\.t-tabs__content\)\s*\{[\s\S]*?overflow:\s*visible;/);
assert.match(styleSource, /:global\(\.t-tab-panel\)\s*\{[\s\S]*?overflow:\s*visible;/);
assert.match(styleSource, /\.segmentButton\s*\{[\s\S]*?height:\s*32px;/);
assert.match(styleSource, /\.laneCardBody\s+:global\(\.t-input__wrap\),/);
assert.match(styleSource, /\.laneRatioControl\s+:global\(\.t-input-number\)\s*\{[\s\S]*?width:\s*100%;/);
assert.match(styleSource, /min-height:\s*32px;/);
assert.doesNotMatch(laneGroupTableSource, /view=\{\s*<LaneGroupEdtor/);
assert.doesNotMatch(workbenchSource, /view=\{\s*<LaneGroupEdtor/);

console.log('lane editor utils verification passed');
