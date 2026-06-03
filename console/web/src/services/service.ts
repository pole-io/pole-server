import request, { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { CheckVisibilityMode } from 'utils/visible';
import { BaseURL } from './types';

export interface Service {
    id: string
    name: string
    namespace: string
    ports: string
    comment: string
    revision: string
    department: string
    business: string
    metadata: Record<string, string>
    export_to?: string[]
}

export interface ServiceView extends Service {
    healthy_instance_count?: string
    total_instance_count?: string
    visibility_mode?: string
    ctime: string
    mtime: string
    editable: boolean
    deleteable: boolean
}

export interface DescribeServicesRequest {
    offset: number
    limit: number
    id?: string
    name?: string
    namespace?: string
    host?: string
    keys?: string
    values?: string
    business?: string
    department?: string
    only_exist_health_instance?: boolean
}

export interface DescribeServicesResponse {
    amount: number
    size: number
    services: Array<ServiceView>
}

export async function describeServices(params: DescribeServicesRequest) {
    const res = await getApiRequest<DescribeServicesResponse>({
        action: `${BaseURL.SERVICE}`,
        data: params,
    })
    return {
        list: res.services.map((item) => {
            const visibilityMode = CheckVisibilityMode(item.export_to, item.namespace)
            if (visibilityMode === 'all') item.export_to = []
            return {
                ...item,
                visibility_mode: visibilityMode,
            } as ServiceView
        }),
        totalCount: res.amount,
    }
}

export async function describeAllServices(params = {}) {
    const res = await getAllList(describeServices, {})(params)
    return res.list ? res.list : []
}

export interface ModifyServicesRequest {
    name: string
    namespace: string
    comment: string
    business: string
    metadata: Record<string, string>
    department: string
    export_to?: string[]
}

export async function modifyServices(params: ModifyServicesRequest[]) {
    const res = await putApiRequest({
        action: `${BaseURL.SERVICE}`,
        data: params,
    })

    return res
}

export interface CreateServicesRequest {
    name: string
    namespace: string
    ports: string
    comment: string
    business: string
    metadata: Record<string, string>
    owners: string
    department: string
    export_to?: string[]
}

export async function createService(params: CreateServicesRequest[]) {
    const res = await apiRequest({
        action: `${BaseURL.SERVICE}`,
        data: params,
    })

    return res
}

export interface DeleteServicesRequest {
    id?: string
    name?: string
    namespace?: string
}

export async function deleteServices(params: DeleteServicesRequest[]) {
    const res = await apiRequest({
        action: `${BaseURL.SERVICE}/delete`,
        data: params,
    })

    return res
}

export interface DescribeGovernanceServiceContractsRequest {
    /** ID */
    id?: string
    /** 命名空间 */
    namespace?: string
    /** 服务名 */
    service?: string
    /** 契约名称 */
    name?: string
    /** 契约版本 */
    version?: string
    /** 契约协议 */
    protocol?: string
    /** 是否只展示基本信息 */
    brief?: boolean
    /** 分页偏移量 */
    offset: number
    /** 分页条数 */
    limit: number
}

export interface DescribeGovernanceServiceContractsResponse {
    /** 总数 */
    amount?: number

    /** 返回条数 */
    size?: number

    /** 契约定义列表 */
    data?: GovernanceServiceContract[]
}

export function DescribeGovernanceServiceContracts(params: DescribeGovernanceServiceContractsRequest) {
    return getApiRequest<DescribeGovernanceServiceContractsResponse>({
        action: `${BaseURL.SERVICE}/contract`,
        data: params,
    })
}

export interface DescribeGovernanceServiceContractVersionsRequest {
    /** 命名空间 */
    namespace: string

    /** 服务名 */
    service: string
}

export interface DescribeGovernanceServiceContractVersionsResponse {
    /** 服务契约版本列表 */
    data?: GovernanceServiceContractVersion[]
    amount: number
    size: number
}

/** 查询服务下契约版本列表 */
export function DescribeGovernanceServiceContractVersions(params: DescribeGovernanceServiceContractVersionsRequest) {
    return getApiRequest<DescribeGovernanceServiceContractVersionsResponse>({
        action: `${BaseURL.SERVICE}/contract/versions`,
        data: params,
    })
}

/** 服务契约版本信息 */
export interface GovernanceServiceContractVersion {
    /** 契约版本 */
    version?: string

    /** 契约名称 */
    name?: string
}

/** 服务契约定义 */
export interface GovernanceServiceContract {
    /** 契约ID */
    id?: string

    /** 契约名称 */
    name?: string

    /** 所属服务命名空间 */
    namespace?: string

    /** 所属服务名称 */
    service?: string

    /** 协议 */
    protocol?: string

    /** 版本 */
    version?: string

    /** 信息摘要 */
    revision?: string

    /** 额外内容描述 */
    content?: string

    /** 创建时间 */
    ctime?: string

    /** 修改时间 */
    mtime?: string

    /** 契约接口列表 */
    interfaces?: GovernanceInterfaceDescription[]

    /**
     * 服务契约在线状态
     *
     * 1、Online：在线
     *
     * 2、Offline：离线
     */
    status?: string
}

/** 服务契约接口定义 */
export interface GovernanceInterfaceDescription {
    /** 契约接口ID */
    id?: string

    /** 方法名称 */
    method?: string

    /** 路径/接口名称 */
    path?: string

    /** 内容 */
    content?: string

    /** 创建来源 */
    source?: string

    /** 信息摘要 */
    revision?: string

    /** 创建时间 */
    ctime?: string

    /** 修改时间 */
    mtime?: string
}

export interface DeleteGovernanceServiceContractInterfacesRequest {
    /** 契约ID */
    id?: string

    /** 契约名称 */
    name?: string

    /** 所属服务命名空间 */
    namespace?: string

    /** 所属服务名称 */
    service?: string

    /** 协议 */
    protocol?: string

    /** 版本 */
    version?: string

    /** 额外内容描述 */
    content?: string

    /** 契约接口列表 */
    interfaces?: GovernanceInterfaceDescription[]
}


/** 批量删除服务契约接口定义 */
export function DeleteGovernanceServiceContractInterfaces(params: DeleteGovernanceServiceContractInterfacesRequest) {
    return apiRequest({
        action: `${BaseURL.SERVICE}/contract/interfaces/delete`,
        data: params,
    })
}