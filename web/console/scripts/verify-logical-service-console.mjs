import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const read = relative => fs.readFileSync(path.join(root, relative), 'utf8')
const assert = (condition, message) => {
  if (!condition) throw new Error(message)
}

const router = read('src/router/modules/discovery.ts')
const list = read('src/pages/Discovery/Services/services.tsx')
const detail = read('src/pages/Discovery/Services/LogicalServiceDetail.tsx')
const environmentDetail = read('src/pages/Discovery/Services/Instance/ServiceDetail.tsx')
const systemServices = read('src/pages/Discovery/Services/SystemNamespaceServices.tsx')
const api = read('src/services/logical_service.ts')

assert(router.includes("path: 'service/detail'") && router.includes('LogicalServiceDetail'), '缺少逻辑服务详情隐藏路由')
assert(list.includes('逻辑服务') && list.includes('未关联环境服务'), '根列表必须区分逻辑服务与未关联环境服务')
assert(list.includes('describeLogicalServices') && !list.includes('listServices'), '逻辑列表不得再按运行时服务名猜测聚合')
assert(detail.includes('bindServiceEnvironment') && detail.includes('unbindServiceEnvironment'), '详情必须支持显式关联和解除')
assert(list.includes('deleteLogicalService') && list.includes('deleteServices'), '服务列表必须分别支持删除逻辑服务和环境服务')
assert(list.includes('row.environment_count') && list.includes('row.deleteable === false'), '逻辑服务删除必须同时受环境关联和权限约束')
assert(list.includes('id: row.id') && list.includes('namespace: row.namespace') && list.includes('name: row.name'), '环境服务删除必须发送 ID 与 namespace/name 完整坐标')
assert(environmentDetail.includes('deleteServices') && environmentDetail.includes('删除环境服务'), '环境服务详情必须提供受权限约束的删除入口')
assert(systemServices.includes('deleteServices') && systemServices.includes('删除系统服务'), '系统空间服务列表必须保留环境服务删除入口')
assert(detail.includes('历史系统空间关联（仅清理）') && detail.includes('systemBindings'), '历史系统绑定必须可发现并保留解绑入口')
assert(environmentDetail.includes('binding.service_name'), '环境切换必须同时使用 Namespace 与运行时服务名')
assert(!environmentDetail.includes('describeServiceEnvironments'), '环境详情不得继续按同名服务猜测关联')
assert(api.includes('/environment-bindings') && api.includes('/unbound-environments'), '逻辑服务管理 API seam 不完整')

console.log('逻辑服务 Console 契约检查通过。')
