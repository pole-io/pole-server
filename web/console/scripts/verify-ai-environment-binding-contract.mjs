import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const read = (path) => readFile(new URL(`../${path}`, import.meta.url), 'utf8');
const [definitionService, mcpService, a2aService, mcpDetail, a2aDetail, bindingComponent] = await Promise.all([
  read('src/services/ai_definition.ts'),
  read('src/services/mcp.ts'),
  read('src/services/a2a.ts'),
  read('src/pages/AI/Mcp/Detail.tsx'),
  read('src/pages/AI/A2A/Detail.tsx'),
  read('src/components/AIEnvironmentBinding/index.tsx'),
]);

assert.match(definitionService, /environment-binding/, '共享客户端应支持解析当前环境绑定');
assert.match(definitionService, /environments/, '共享客户端应支持列出逻辑定义的环境实例');
assert.match(definitionService, /environment-bindings/, '共享客户端应支持显式绑定环境实例');

for (const [label, source] of [['MCP', mcpService], ['A2A', a2aService]]) {
  assert.match(source, /describeAIEnvironmentBinding/, `${label} 应复用共享绑定客户端`);
  assert.match(source, /bindAIResourceDefinition/, `${label} 应复用共享显式绑定客户端`);
}

for (const [label, source] of [['MCP', mcpDetail], ['A2A', a2aDetail]]) {
  assert.match(source, /EnvironmentResourceSwitcher/, `${label} 详情应提供环境切换`);
  assert.match(source, /未关联跨环境逻辑定义/, `${label} 详情应显式展示旧记录的未关联状态`);
  assert.match(source, /namespace !== 'pole-system'/, `${label} 不应把系统投影纳入业务聚合`);
}

assert.match(bindingComponent, /不会按名称自动合并/, '关联交互必须声明禁止按同名自动合并');
assert.match(bindingComponent, /选择已有逻辑定义/, '关联交互应允许显式选择已有定义');

console.log('AI 环境逻辑定义与显式绑定契约检查通过。');
