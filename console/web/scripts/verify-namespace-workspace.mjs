import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');

const page = read('src/pages/Namespace/index.tsx');
const styles = read('src/pages/Namespace/index.module.less');
const namespaceService = read('src/services/namespace.ts');

const pageRule = /\.page\s*\{[\s\S]*?display:\s*flex;[\s\S]*?flex-direction:\s*column;[\s\S]*?height:\s*calc\(100vh - 113px\);[\s\S]*?overflow:\s*hidden;/;
const workspaceRule = /\.namespaceWorkspace\s*\{[\s\S]*?flex:\s*1 1 auto;[\s\S]*?min-height:\s*0;/;
const tableRule = /\.namespaceTableSurface\s*\{[\s\S]*?flex:\s*1 1 auto;[\s\S]*?min-height:\s*0;[\s\S]*?overflow:\s*hidden;/;
const scrollRule = /\.namespaceTableSurface\s*\{[\s\S]*?\.fluent-table-scroll\)\s*\{[\s\S]*?flex:\s*1 1 auto;[\s\S]*?min-height:\s*0;[\s\S]*?overflow:\s*auto;[\s\S]*?overscroll-behavior:\s*contain;/;

if (!pageRule.test(styles)) throw new Error('命名空间页根容器必须固定高度并禁止页面滚动');
if (!workspaceRule.test(styles)) throw new Error('命名空间工作区必须占用页头以下的剩余高度');
if (!tableRule.test(styles)) throw new Error('命名空间表格容器必须作为内部滚动布局的伸缩项');
if (!scrollRule.test(styles)) throw new Error('命名空间表格内容区必须独占滚动，且不能把滚动冒泡给页面');
if (!page.includes('style.namespaceWorkspace')) throw new Error('命名空间页必须使用固定工作区容器');
if (!page.includes('style.namespaceTableSurface')) throw new Error('命名空间列表必须使用内部滚动表格容器');
if (!page.includes("title: '配置文件'")) throw new Error('命名空间列表必须显示配置文件数量列');
if (!page.includes('当前页配置文件')) throw new Error('命名空间摘要必须显示当前页配置文件总数');
if (!page.includes('item.total_config_file_count')) throw new Error('命名空间摘要必须使用后端返回的配置文件统计');
if (!namespaceService.includes('total_config_file_count?: number')) throw new Error('命名空间服务模型必须声明配置文件统计字段');

console.log('命名空间工作台滚动与配置数量静态约束验证通过');
