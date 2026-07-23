import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const source = fs.readFileSync(
  path.join(root, 'src/pages/Governance/CircuitBreaker/CircuitBreakerEditor.tsx'),
  'utf8',
);

const assert = (condition, message) => {
  if (!condition) throw new Error(message);
};

const stickyStart = source.indexOf('const renderStickyTool = (');
const stickyEnd = source.indexOf('\n    return (', stickyStart);
const stickySource = source.slice(stickyStart, stickyEnd);

assert(stickyStart >= 0 && stickyEnd >= 0, '熔断编辑器必须渲染操作区');
assert(
  !stickySource.includes('<FormItem'),
  '熔断编辑器的操作区不是表单字段，不得依赖 FormItem 上下文',
);
assert(
  (source.match(/<Form form=/g) || []).length === 1 && (source.match(/<\/Form>/g) || []).length === 1,
  '熔断编辑器必须只保留一个主 Form，发布操作不得引入嵌套表单',
);
assert(
  source.includes("op === 'create' || viewRule?.editable !== false"),
  '创建模式始终提供操作区，查看和编辑模式仅在 API 明确拒绝编辑时隐藏操作区',
);
assert(
  source.includes("(!editorState.editable && op !== 'create')"),
  '只读详情模式必须提供发布操作，创建模式不得显示发布操作',
);
assert(stickySource.includes('form.submit();'), '编辑模式的保存操作必须提交主 Form');

console.log('Circuit-breaker editor action rendering contract verified');
