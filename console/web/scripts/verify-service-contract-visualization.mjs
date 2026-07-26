import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const componentFile = path.join(
  root,
  'src/pages/Discovery/Services/Instance/ServiceContractPanel.tsx',
);
const styleFile = path.join(
  root,
  'src/pages/Discovery/Services/Instance/ServiceContractPanel.module.less',
);

assert.ok(fs.existsSync(componentFile), '服务详情必须提供独立的服务契约展示组件');
assert.ok(fs.existsSync(styleFile), '服务契约展示组件必须使用独立样式，避免污染服务详情页');

const component = fs.readFileSync(componentFile, 'utf8');
const style = fs.readFileSync(styleFile, 'utf8');

assert.match(
  component,
  /from ['"]components\/Fluent['"]/,
  '服务契约展示必须复用 Fluent 适配层',
);
assert.match(
  component,
  /describeGovernanceServiceContracts|DescribeGovernanceServiceContracts/,
  '服务契约展示必须使用 service.ts 导出的契约查询函数',
);
assert.match(
  component,
  /describeGovernanceServiceContractVersions|DescribeGovernanceServiceContractVersions/,
  '服务契约展示必须使用 service.ts 导出的版本查询函数',
);

for (const protocol of ['HTTP / OpenAPI', 'Dubbo', 'gRPC', 'Thrift']) {
  assert.match(component, new RegExp(protocol.replace('/', '\\/')), `必须展示 ${protocol} 协议标签`);
}

for (const anchor of [
  '契约定义',
  '契约版本',
  '接口清单',
  '原始契约',
  '加载服务契约中',
  '服务契约加载失败',
  '暂无服务契约',
]) {
  assert.match(component, new RegExp(anchor), `服务契约组件必须覆盖“${anchor}”`);
}

assert.match(
  component,
  /protocolOperationLabel/,
  '接口行必须通过协议化文案函数解释 path/method',
);
assert.match(
  component,
  /selectedContractId/,
  '同一服务存在多个契约时必须允许选择契约',
);
assert.match(
  component,
  /selectedVersion/,
  '同一契约存在多个版本时必须允许选择版本',
);
assert.match(
  component,
  /role="tablist"[\s\S]*role="tab"[\s\S]*aria-selected/,
  '接口清单与原始契约必须提供可访问的视图切换',
);
assert.match(
  component,
  /<pre[\s\S]*selectedContract\?\.content/,
  '原始契约必须以保留格式的代码区域展示',
);

assert.match(
  style,
  /\.panel[\s\S]*display: grid/,
  '服务契约面板必须建立独立布局',
);
assert.match(
  style,
  /\.protocolGrid[\s\S]*grid-template-columns: repeat\(4, minmax\(0, 1fr\)\)/,
  '四类协议能力必须在宽屏保持稳定四列展示',
);
assert.match(
  style,
  /\.rawContent[\s\S]*overflow: auto[\s\S]*font-family:/,
  '原始契约需要独立滚动并使用等宽字体',
);
assert.match(
  style,
  /@media \(max-width: 760px\)[\s\S]*\.protocolGrid[\s\S]*grid-template-columns: repeat\(2, minmax\(0, 1fr\)\)/,
  '窄屏下四协议标签必须收敛为两列',
);

console.log('service contract visualization verification passed');
