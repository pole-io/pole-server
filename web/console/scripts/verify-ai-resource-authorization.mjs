#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');

const assert = (condition, message) => {
  if (!condition) {
    throw new Error(message);
  }
};

const authPolicy = read('src/services/auth_policy.ts');
assert(authPolicy.includes('MCPServerResources = "MCPServerResources"'), 'PolicySourceType 缺少 MCPServerResources');
assert(authPolicy.includes('A2AAgentResources = "A2AAgentResources"'), 'PolicySourceType 缺少 A2AAgentResources');
assert(authPolicy.includes('mcp_servers?: PolicyResource[]'), 'PolicyResources 缺少 mcp_servers 字段');
assert(authPolicy.includes('a2a_agents?: PolicyResource[]'), 'PolicyResources 缺少 a2a_agents 字段');

const policyEditor = read('src/pages/Auth/Policy/PolicyEditor.tsx');
[
  'describeAllMCPServers',
  'describeAllA2AAgents',
  "value: 'MCPServer'",
  "value: 'A2AAgent'",
  "value: 'AINative'",
  'resources?.mcp_servers',
  'resources?.a2a_agents',
  "state.selectResources['MCPServer']",
  "state.selectResources['A2AAgent']",
].forEach((token) => assert(policyEditor.includes(token), `PolicyEditor 缺少 ${token}`));

const policyDetail = read('src/pages/Auth/Policy/PolicyDetailView.tsx');
assert(policyDetail.includes("value: 'mcp_servers'"), 'PolicyDetailView 缺少 mcp_servers 资源类型');
assert(policyDetail.includes("value: 'a2a_agents'"), 'PolicyDetailView 缺少 a2a_agents 资源类型');

const authorize = read('src/components/Authorize/index.tsx');
assert(authorize.includes('PolicySourceType.MCPServerResources'), 'AuthorizeInput 缺少 MCP Server 资源摘要适配');
assert(authorize.includes('PolicySourceType.A2AAgentResources'), 'AuthorizeInput 缺少 A2A Agent 资源摘要适配');

const mcpPage = read('src/pages/AI/Mcp/index.tsx');
assert(mcpPage.includes('AuthorizeInput'), 'MCP 页面缺少授权抽屉');
assert(mcpPage.includes('PolicySourceType.MCPServerResources'), 'MCP 页面授权资源类型错误');
assert(mcpPage.includes('OperationButton action="view" onClick={() => operateServer(\'detail\', row)}'), 'MCP 页面主入口必须是查看详情');
assert(mcpPage.includes('OperationButton action="authorize"'), 'MCP 页面缺少统一授权按钮');
assert(!mcpPage.includes('OperationButton action="edit" onClick={() => operateServer(\'edit\', row)}'), 'MCP 页面不能在列表操作列直接暴露编辑');

const a2aPage = read('src/pages/AI/A2A/index.tsx');
assert(a2aPage.includes('AuthorizeInput'), 'A2A 页面缺少授权抽屉');
assert(a2aPage.includes('PolicySourceType.A2AAgentResources'), 'A2A 页面授权资源类型错误');
assert(a2aPage.includes('OperationButton action="authorize"'), 'A2A 页面缺少统一授权按钮');

console.log('AI resource authorization checks passed');
