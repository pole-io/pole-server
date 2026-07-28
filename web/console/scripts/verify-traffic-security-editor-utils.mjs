import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import ts from 'typescript';

const root = process.cwd();
const sourcePath = path.join(root, 'src/pages/Governance/Security/trafficSecurityEditorUtils.ts');
const editorPath = path.join(root, 'src/pages/Governance/Security/TrafficGovernanceEditor.tsx');
const stylePath = path.join(root, 'src/pages/Governance/Security/index.module.less');

if (!fs.existsSync(sourcePath)) {
  throw new Error(`missing helper: ${sourcePath}`);
}
if (!fs.existsSync(editorPath)) {
  throw new Error(`missing editor: ${editorPath}`);
}
if (!fs.existsSync(stylePath)) {
  throw new Error(`missing style: ${stylePath}`);
}

const editorSource = fs.readFileSync(editorPath, 'utf8');
const styleSource = fs.readFileSync(stylePath, 'utf8');
assert.match(editorSource, /const renderSecurityServiceInfo = \(\) =>/);
assert.match(editorSource, /\{kind === 'security' && renderSecurityServiceInfo\(\)\}/);
assert.doesNotMatch(editorSource, /renderCommonFields[\s\S]*?kind === 'security' && readonlyItem\('被调命名空间'/);
assert.doesNotMatch(editorSource, /renderCommonFields[\s\S]*?kind === 'security' && <FormItem[\s\S]*?label="被调命名空间"/);
assert.match(editorSource, /'① 基础信息'/);
assert.match(editorSource, /collapsed=\{securityServiceInfoCollapsed\}/);
assert.match(editorSource, /onCollapsedChange=\{setSecurityServiceInfoCollapsed\}/);
assert.match(editorSource, /header="② 服务信息"/);
assert.match(editorSource, /collapseLabel="服务信息"/);
assert.match(editorSource, /className=\{style\.securityServiceInfoFields\}/);
assert.match(editorSource, /className=\{style\.securityServiceInfoField\}/);
assert.match(styleSource, /\.securityServiceInfoFields\s*\{[\s\S]*max-inline-size: 560px/);
assert.match(editorSource, /collapsed=\{securityAuthenticationCollapsed\}/);
assert.match(editorSource, /onCollapsedChange=\{setSecurityAuthenticationCollapsed\}/);
assert.match(editorSource, /header="③ 认证方式"/);
assert.match(editorSource, /summary=\{modeLabel\}/);
assert.match(editorSource, /collapseLabel="认证方式"/);
assert.match(editorSource, /'④ 鉴权子规则'/);
assert.match(editorSource, /Pole 托管服务身份（推荐）/);
assert.match(editorSource, /自定义 Header（兼容模式）/);
assert.match(editorSource, /在每个鉴权子规则中配置 Header 匹配条件/);
assert.match(editorSource, /自定义 Header 匹配条件/);
assert.doesNotMatch(editorSource, /custom-header-(?:name|value|credential-notice)/);
assert.doesNotMatch(editorSource, /规则级自定义 Header/);
assert.doesNotMatch(editorSource, /实时\s*Spec|security-live-spec|buildSecurityPreviewSpec|stringifySecuritySpec/);
assert.match(editorSource, /setReloadVersion\(\(value\) => value \+ 1\)/);
assert.match(editorSource, /className=\{style\.managedCallerSelect\}[\s\S]*multiple[\s\S]*filterable[\s\S]*label="搜索来源服务"/);
assert.match(editorSource, /\.sort\(\(left, right\) => `\$\{left\.namespace\}\/\$\{left\.name\}`\.localeCompare/);
assert.match(styleSource, /\.managedCallerSelect\s*\{[\s\S]*inline-size: min\(100%, 560px\)/);
assert.match(styleSource, /\.securityEditorShell\s*\{[\s\S]*height: auto;[\s\S]*overflow: visible;/);
assert.match(styleSource, /\.interfaceRow > :global\(\[role='combobox'\]\)[\s\S]*width: 100%/);
assert.doesNotMatch(editorSource, /const renderProtectedInterfacesEditor[\s\S]*?<span>值类型<\/span>/);
assert.doesNotMatch(editorSource, /const renderProtectedInterfacesEditor[\s\S]*?MatchValueTypeOption/);
assert.match(styleSource, /\.interfaceHeader,[\s\S]*grid-template-columns: 80px 80px 104px minmax\(144px, 1fr\) 32px/);

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
export const TrafficSecurityAuthMode = {
  LEGACY_REQUEST_MATCH: 'LEGACY_REQUEST_MATCH',
  MANAGED_IDENTITY: 'MANAGED_IDENTITY',
  CUSTOM_HEADER: 'CUSTOM_HEADER',
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

assert.deepEqual(utils.getApiProtocolPresentation('HTTP'), {
  methodLabel: 'HTTP 方法',
  methodPlaceholder: 'GET',
  pathLabel: '接口路径',
  pathPlaceholder: '/orders',
  usesHttpMethodSelect: true,
});
assert.deepEqual(utils.getApiProtocolPresentation('GRPC'), {
  methodLabel: '可选方法',
  methodPlaceholder: '可选，例如 SayHello',
  pathLabel: '服务名',
  pathPlaceholder: 'helloworld.Greeter',
  usesHttpMethodSelect: false,
});
assert.deepEqual(utils.getApiProtocolPresentation('DUBBO'), {
  methodLabel: '可选方法',
  methodPlaceholder: '可选，例如 getUser',
  pathLabel: '接口名',
  pathPlaceholder: 'com.example.UserService',
  usesHttpMethodSelect: false,
});
assert.deepEqual(utils.resetApiScopeForProtocol({
  protocol: 'HTTP',
  method: 'GET',
  path: { type: 'EXACT', value: '/orders' },
}, 'DUBBO'), {
  protocol: 'DUBBO',
  method: '',
  path: { type: 'EXACT', value: '' },
});
assert.deepEqual(utils.resetApiScopeForProtocol({
  protocol: 'DUBBO',
  method: 'getUser',
  path: { type: 'REGEX', value: 'com.example.UserService' },
}, 'HTTP'), {
  protocol: 'HTTP',
  method: 'GET',
  path: { type: 'REGEX', value: '' },
});
assert.deepEqual(utils.buildSecurityPoliciesFromView([{
  kind: 'allow',
  listType: 'ALLOW_LIST',
  interfaces: [{ protocol: 'DUBBO', method: '', path: { type: 'EXACT', value: '' } }],
  managedCaller: { any_authenticated: true, callers: [] },
  match: { matchMode: 'AND', randomPercent: 0, arguments: [] },
}], 'MANAGED_IDENTITY')[0].apis[0], {
  protocol: 'DUBBO',
  method: '',
  path: { type: 'EXACT', value: '' },
});
assert.match(editorSource, /const presentation = getApiProtocolPresentation\(current\.protocol\)/);
assert.match(editorSource, /placeholder=\{presentation\.methodPlaceholder\}/);
assert.match(editorSource, /placeholder=\{presentation\.pathPlaceholder\}/);
assert.match(editorSource, /<span>HTTP 方法 \/ RPC 接口<\/span>/);
assert.match(editorSource, /<span>HTTP 路径 \/ RPC 方法（可选）<\/span>/);
assert.match(editorSource, /presentation\.usesHttpMethodSelect \? methodField : pathField/);
assert.match(editorSource, /style\.rpcInterfaceRow/);
assert.match(styleSource, /\.rpcInterfaceRow\s*\{[\s\S]*minmax\(280px, 2fr\)[\s\S]*minmax\(160px, 0\.7fr\)/);
assert.deepEqual(utils.validateSecurityView({
  name: 'rpc-optional-method',
  target_service: { namespace: 'spec-governance', service: 'spec-gateway' },
  authentication: { mode: 'MANAGED_IDENTITY' },
}, [{
  kind: 'allow',
  listType: 'ALLOW_LIST',
  interfaces: [{ protocol: 'GRPC', method: '', path: { type: 'EXACT', value: 'helloworld.Greeter' } }],
  managedCaller: { any_authenticated: true, callers: [] },
  strategy: { matchMode: 'AND', randomPercent: 0, arguments: [] },
}]), []);

const rule = {
  name: 'spec-check-traffic-security-20260612',
  target_service: { namespace: 'spec-governance', service: 'spec-gateway' },
  description: '调用鉴权样例',
  priority: 10,
  enable: true,
  metadata: { owner: 'codex', scenario: 'sample' },
  policies: [{
    action: 'TRAFFIC_SECURITY_ALLOW',
    apis: [{
      protocol: 'HTTP',
      method: 'GET',
      path: { type: 'IN', value: '/orders', value_type: 'TEXT' },
    }],
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
    apis: [{
      protocol: 'HTTP',
      method: 'POST',
      path: { type: 'IN', value: '/admin', value_type: 'TEXT' },
    }],
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
assert.deepEqual(view[0].interfaces[0].path, { type: 'IN', value: '/admin' });
assert.deepEqual(view[1].interfaces[0].path, { type: 'IN', value: '/orders' });

const submitPolicies = utils.buildSecurityPoliciesFromView(view, 'LEGACY_REQUEST_MATCH');
assert.equal(submitPolicies.length, 3);
assert.equal(submitPolicies[0].action, 'TRAFFIC_SECURITY_DENY');
assert.equal(submitPolicies[0].apis[0].path.value, '/admin');
assert.deepEqual(submitPolicies[0].apis[0].path, { type: 'IN', value: '/admin' });
assert.equal(submitPolicies[0].api, undefined);
assert.equal(submitPolicies[1].action, 'TRAFFIC_SECURITY_ALLOW');
assert.equal(submitPolicies[1].apis[0].path.value, '/orders');
assert.deepEqual(submitPolicies[1].apis[0].path, { type: 'IN', value: '/orders' });
assert.equal(submitPolicies[2].api, undefined);
assert.equal(submitPolicies[2].apis, undefined);

const managedRule = {
  ...rule,
  authentication: { mode: 'MANAGED_IDENTITY', managed_identity: {} },
  policies: [{
    action: 'TRAFFIC_SECURITY_ALLOW',
    apis: rule.policies[0].apis,
    managed_caller: {
      any_authenticated: false,
      callers: [{ namespace: 'default', service: 'checkout' }],
    },
  }],
};
const managedView = utils.normalizeSecurityViewRules(managedRule);
const managedPolicies = utils.buildSecurityPoliciesFromView(managedView, 'MANAGED_IDENTITY');
assert.equal(utils.readSecurityAuthMode(managedRule), 'MANAGED_IDENTITY');
assert.equal(managedPolicies[0].traffic_match_rule, undefined);
assert.deepEqual(managedPolicies[0].managed_caller.callers, [{ namespace: 'default', service: 'checkout' }]);
assert.deepEqual(utils.validateSecurityView(managedRule, managedView), []);

assert.deepEqual(utils.buildSecurityAuthenticationForMode('MANAGED_IDENTITY', 'x-api-key'), {
  mode: 'MANAGED_IDENTITY',
  managed_identity: {},
});
assert.deepEqual(utils.buildSecurityAuthenticationForMode('LEGACY_REQUEST_MATCH', 'ignored'), {
  mode: 'LEGACY_REQUEST_MATCH',
});
assert.deepEqual(utils.buildSecurityAuthenticationForMode('CUSTOM_HEADER'), {
  mode: 'LEGACY_REQUEST_MATCH',
});
assert.deepEqual(utils.buildSecurityPoliciesFromView(view, 'CUSTOM_HEADER')[0].traffic_match_rule, {
  matchMode: 'AND',
  randomPercent: 0,
  arguments: [{ type: 'HEADER', key: 'authorization', value: { type: 'IN', value: 'deny-', value_type: 'TEXT' } }],
});

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
assert.deepEqual(utils.validateSecurityView({ ...rule, name: 'Invalid Name' }, view).map(item => item.message), [
  '规则名称需为 kebab-case',
]);

console.log('traffic security editor utils verification passed');
