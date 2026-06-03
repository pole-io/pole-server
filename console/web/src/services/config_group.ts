import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';
import { BaseURL } from './types';

// 配置分组
export interface ConfigFileGroup {
    id: number
    name: string
    namespace: string
    comment: string
    fileCount: number
    department?: string
    business?: string
    metadata?: Record<string, string>
}

// 配置分组
export interface ConfigFileGroupView extends ConfigFileGroup {
    createTime: string
    createBy: string
    modifyTime: string
    modifyBy: string
    editable: boolean
    deleteable: boolean
}


// 查询配置分组列表
export interface DescribeConfigFileGroupRequest {
    offset: number
    limit: number
    namespace?: string
    group?: string
    file_name?: string
}

export interface DescribegroupsResponse {
    total: number
    configFileGroups: Array<ConfigFileGroupView>
}

export async function describeConfigFileGroups(params: DescribeConfigFileGroupRequest) {
    const res = await getApiRequest<DescribegroupsResponse>({
        action: `${BaseURL.CONFIG_GROUP}`,
        data: params,
    })
    return {
        list: res.configFileGroups,
        totalCount: res.total,
    }
}

export async function describeAllConfigGroups() {
    const res = await getAllList(describeConfigFileGroups, {
        listKey: 'list',
        totalKey: 'totalCount',
    })({})
    return {
        list: res.list ? res.list : [],
        totalCount: res.totalCount
    }
}

// 创建配置分组
export interface CreateConfigFileGroupRequest {
    name: string
    namespace: string
    comment?: string
    department?: string
    business?: string
    metadata?: Record<string, string>
}

export interface CreateConfigFileGroupResponse {
    configFileGroup: ConfigFileGroup
}

export async function createConfigFileGroups(params: CreateConfigFileGroupRequest[]) {
    const res = await apiRequest<CreateConfigFileGroupResponse>({
        action: `${BaseURL.CONFIG_GROUP}`,
        data: params,
    })
    return res
}

// 修改配置分组
export interface ModifyConfigFileGroupRequest {
    id: number
    name: string
    namespace: string
    comment?: string
    department?: string
    business?: string
    metadata?: Record<string, string>
}

export interface ModifyConfigFileGroupResponse {
    configFileGroup: ConfigFileGroup
}

export async function modifyConfigFileGroups(params: ModifyConfigFileGroupRequest[]) {
    const res = await putApiRequest<ModifyConfigFileGroupResponse>({
        action: `${BaseURL.CONFIG_GROUP}`,
        data: params,
    })
    return res
}

export interface DeleteConfigFileGroupRequest {
    id: number,
    namespace?: string
    group?: string
}

export type DeleteConfigFileGroupResponse = {}

export async function deleteConfigFileGroups(params: DeleteConfigFileGroupRequest[]) {
    const res = await apiRequest<DeleteConfigFileGroupResponse>({
        action: `${BaseURL.CONFIG_GROUP}/delete`,
        data: params,
    })
    return res
}