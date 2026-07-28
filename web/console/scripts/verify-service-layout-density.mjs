import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const read = relative => fs.readFileSync(path.join(root, relative), 'utf8')

const resourceLayout = read('src/components/ResourceLayout/index.tsx')
const resourceLayoutStyle = read('src/components/ResourceLayout/index.module.less')
const appHeader = read('src/layouts/components/Header/index.tsx')
const appHeaderStyle = read('src/layouts/components/Header/index.module.less')
const servicesIndex = read('src/pages/Discovery/Services/index.tsx')
const services = read('src/pages/Discovery/Services/services.tsx')
const logicalDetail = read('src/pages/Discovery/Services/LogicalServiceDetail.tsx')
const serviceStyle = read('src/pages/Discovery/Services/index.module.less')
const environmentWorkspace = read('src/pages/Discovery/Services/Instance/index.tsx')
const environmentStyle = read('src/pages/Discovery/Services/Instance/index.module.less')
const alias = read('src/pages/Discovery/Services/alias.tsx')
const subscribe = read('src/pages/Discovery/Services/Instance/SubscribeTable.tsx')
const contractStyle = read('src/pages/Discovery/Services/Instance/ServiceContractPanel.module.less')

assert.match(resourceLayout, /density\?: 'default' \| 'compact'/,
  '通用资源标题和工具栏必须支持显式紧凑密度')
assert.match(resourceLayoutStyle, /\.resourceHeaderCompact[\s\S]*padding: 0 0 14px/,
  '紧凑资源标题必须减少底部占用')
assert.match(resourceLayoutStyle, /\.resourceToolbarCompact[\s\S]*margin-bottom: 8px/,
  '紧凑工具栏必须减少与表格之间的空白')
assert.match(resourceLayout, /placement\?: 'content' \| 'app-header'/,
  '资源标题必须支持进入全局顶部导航')
assert.match(resourceLayout, /createPortal[\s\S]*app-header-context/,
  '资源标题进入顶部导航时必须使用稳定 portal 插槽')
assert.match(resourceLayoutStyle, /\.resourceHeaderIntegrated[\s\S]*border-bottom: 0/,
  '顶部导航中的资源信息不能保留内容区卡片边界')
assert.match(resourceLayoutStyle, /\.resourceHeaderIntegrated[\s\S]*\.eyebrow[\s\S]*color: var\(--app-text-secondary\)[\s\S]*font-weight: 500/,
  '顶部导航路径必须使用弱化的中等字重')
assert.match(resourceLayoutStyle, /\.resourceHeaderIntegrated[\s\S]*\.headerContent h2[\s\S]*overflow: hidden[\s\S]*max-width: clamp\([\s\S]*font-size: 17px[\s\S]*font-weight: 600[\s\S]*text-overflow: ellipsis/,
  '顶部导航资源名必须保持克制的标题层级，并对长名称做受控截断')
assert.match(appHeader, /id="app-header-context"[\s\S]*className=\{Style\.headerContext\}/,
  '全局顶部导航必须提供资源页面上下文插槽')
assert.match(appHeaderStyle, /\.headerContext[\s\S]*flex: 1 1 auto[\s\S]*min-width: 0/,
  '页面上下文插槽必须占用顶部导航剩余空间并允许文本截断')

assert.match(servicesIndex, /<ResourceHeader[\s\S]*density="compact"/,
  '服务首页必须使用紧凑资源标题')
assert.match(servicesIndex, /<ResourceHeader[\s\S]*placement="app-header"/,
  '服务首页标题与说明必须进入顶部导航')
assert.doesNotMatch(servicesIndex, /actions=/,
  '服务首页操作不得进入顶部导航')
assert.match(services, /<ResourceToolbar[\s\S]*刷新服务列表[\s\S]*新建逻辑服务/,
  '服务首页刷新与新建操作必须保留在正文工具栏')
assert.match(services, /<Tabs className=\{style\.workspaceTabs\}/,
  '服务首页页签必须接入紧凑工作区样式')
assert.match(services, /<ResourceToolbar[\s\S]*density="compact"/,
  '服务首页清单工具栏必须使用紧凑密度')
assert.match(logicalDetail, /className=\{`\$\{style\.page\} \$\{style\.logicalDetailPage\}`\}/,
  '逻辑服务详情必须使用紧凑详情布局')
assert.match(logicalDetail, /<ResourceHeader[\s\S]*density="compact"/,
  '逻辑服务详情必须使用紧凑资源标题')
assert.match(logicalDetail, /<ResourceHeader[\s\S]*placement="app-header"/,
  '逻辑服务详情上下文必须进入顶部导航')
assert.doesNotMatch(logicalDetail, /<ResourceHeader[\s\S]*actions=/,
  '逻辑服务详情操作不得进入顶部导航')
assert.match(logicalDetail, /className=\{`\$\{style\.tableSurface\} \$\{style\.serviceTableSurface\}`\}/,
  '逻辑服务详情表格必须接入剩余高度分配链')

assert.match(serviceStyle, /\.serviceWorkspace[\s\S]*margin-top: 14px/,
  '服务工作区应紧接页头，避免首屏顶部留白过大')
assert.match(serviceStyle, /\.workspace[\s\S]*gap: 10px/,
  '服务摘要、页签和清单之间必须使用紧凑节奏')
assert.match(serviceStyle, /\.metricItem[\s\S]*min-height: 58px[\s\S]*padding: 8px 16px/,
  '服务摘要必须使用紧凑高度')
assert.match(serviceStyle, /\.workspaceTabs[\s\S]*\.fluent-tab-content[\s\S]*display: none/,
  '服务切换页签的空内容容器不得继续占用高度')
assert.match(serviceStyle, /\.listSection[\s\S]*gap: 0/,
  '清单容器不得与工具栏 margin 叠加间距')
assert.match(serviceStyle, /\.tableSurface[\s\S]*\.fluent-table-shell th[\s\S]*padding: 10px 14px/,
  '服务表头必须使用数据密集型页面的紧凑行高')
assert.match(serviceStyle, /\.tableSurface[\s\S]*\.fluent-pagination[\s\S]*padding: 10px 16px/,
  '服务分页区不能挤占过多表格高度')

assert.match(environmentWorkspace, /<div className=\{style\.serviceDetailPage\}>[\s\S]*className=\{style\.serviceDetailTabs\}/,
  '环境服务详情页签必须统一接入紧凑样式')
assert.match(environmentWorkspace, /<ResourceHeader[\s\S]*placement="app-header"/,
  '环境服务名称与 Namespace 必须进入顶部导航')
assert.match(environmentWorkspace, /<div className=\{style\.serviceDetailActions\}>[\s\S]*返回/,
  '环境服务返回操作必须保留在正文任务区')
assert.doesNotMatch(environmentWorkspace, /<Breadcrumb|BreadcrumbItem/,
  '环境服务详情不得在内容区重复保留面包屑')
assert.doesNotMatch(environmentWorkspace, /style=\{\{ marginTop: 20 \}\}/,
  '环境服务详情不得继续硬编码宽松页签间距')
assert.match(environmentStyle, /\.serviceDetailTabs[\s\S]*\.fluent-tab-content[\s\S]*padding-top: 8px/,
  '环境服务各子页必须统一压缩页签内容上边距')
assert.match(environmentStyle, /\.serviceDetailTabs[\s\S]*\.fluent-tab-content\) > div[\s\S]*height: 100%[\s\S]*min-height: 0/,
  '页签面板直接子元素必须延续确定高度，避免实例表格 flex 链中断')
assert.match(environmentStyle, /\.serviceDetailPage[\s\S]*height: calc\(100vh - 113px\)[\s\S]*overflow: hidden/,
  '环境服务详情必须建立固定视区高度链')
assert.match(environmentStyle, /\.instanceTablePage[\s\S]*padding: 12px 20px 20px/,
  '实例清单必须减少与页签之间的重复留白')
assert.match(environmentStyle, /\.instanceTableSurface[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0/,
  '实例表格必须填满页签剩余高度并在内部滚动')
assert.match(environmentStyle, /\.detailPanel[\s\S]*margin: 12px 20px 20px/,
  '环境服务详情必须减少与页签之间的重复留白')
assert.match(alias, /embeddedWorkspaceCompact/,
  '内嵌别名页必须使用紧凑工作区')
assert.match(subscribe, /subscribePanelCompact/,
  '服务订阅页必须使用紧凑工作区')
assert.match(contractStyle, /\.panel[\s\S]*padding: 12px 20px 20px/,
  '服务契约页必须减少与页签之间的重复留白')

console.log('服务相关页面纵向密度契约检查通过。')
