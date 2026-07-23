import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');
const assert = (condition, message) => {
  if (!condition) throw new Error(message);
};

const service = read('src/services/config_group.ts');
const table = read('src/pages/Configuration/Group/group.tsx');

assert(service.includes('file_count?: number | string'), '配置分组响应必须声明 proto JSON 的 file_count 字段');
assert(service.includes('fileCount: Number(group.fileCount ?? group.file_count ?? 0)'), '配置分组必须将 file_count 归一化为页面 fileCount');
assert(table.includes("const fileCount = Number(row.fileCount || 0);"), '删除操作必须读取归一化后的配置文件数');
assert(table.includes('disabled={row.deleteable === false || fileCount > 0}'), '含配置文件的分组不能再展示可点击的删除确认');
assert(table.includes('请先删除 ${fileCount} 个配置文件'), '删除禁用提示必须说明阻塞资源数量');

console.log('配置分组删除保护静态约束验证通过');
