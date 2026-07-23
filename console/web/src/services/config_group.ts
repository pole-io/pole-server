import { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { BaseURL } from './types';

// 配置分组
export interface ConfigFileGroup {
    id: string
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

type ApiConfigFileGroup = ConfigFileGroupView & {
    ctime?: string
    mtime?: string
    name?: string
    file_count?: number | string
}

function normalizeConfigFileGroup(group: ApiConfigFileGroup): ConfigFileGroupView {
    return {
        ...group,
        id: group.id,
        name: group.name,
        namespace: group.namespace,
        comment: group.comment,
        fileCount: Number(group.fileCount ?? group.file_count ?? 0),
        createTime: group.createTime || group.ctime || '',
        modifyTime: group.modifyTime || group.mtime || '',
        editable: group.editable ?? true,
        deleteable: group.deleteable ?? true,
    }
}

function toApiConfigGroupQuery(params: DescribeConfigFileGroupRequest) {
    const { group, file_name, ...rest } = params;
    const query: Record<string, unknown> = {
        ...rest,
        name: group || undefined,
    };
    if (group && rest.namespace === group) {
        delete query.namespace;
    }
    return query;
}

function toApiConfigFileGroup(group: DeleteConfigFileGroupRequest | ModifyConfigFileGroupRequest | CreateConfigFileGroupRequest) {
    return {
        ...group,
        name: group.name || ('group' in group ? group.group : undefined),
    };
}

// 查询配置分组列表
export interface DescribeConfigFileGroupRequest {
    offset: number
    limit: number
    namespace?: string
    group?: string
    name?: string
    file_name?: string
}

export interface DescribegroupsResponse {
    amount?: number
    size?: number
    total?: number
    data?: Array<ConfigFileGroupView>
    configFileGroups?: Array<ConfigFileGroupView>
}

export async function describeConfigFileGroups(params: DescribeConfigFileGroupRequest) {
    const res = await getApiRequest<DescribegroupsResponse>({
        action: `${BaseURL.CONFIG_GROUP}`,
        data: toApiConfigGroupQuery(params),
    })
    const groups = res.data ?? res.configFileGroups ?? []
    const normalizedGroups = groups.map(normalizeConfigFileGroup)
    return {
        list: normalizedGroups,
        totalCount: res.amount ?? res.total ?? groups.length,
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

/** 查询同一配置分组在用户有权访问的各个环境中的实例。 */
export async function describeConfigGroupEnvironments(name: string) {
    const res = await describeConfigFileGroups({
        offset: 0,
        limit: 100,
        group: name,
    })
    return res.list.filter((group) => group.name === name)
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
        data: params.map(toApiConfigFileGroup),
    })
    return res
}

// 修改配置分组
export interface ModifyConfigFileGroupRequest {
    id?: string
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
        data: params.map(toApiConfigFileGroup),
    })
    return res
}

export interface DeleteConfigFileGroupRequest {
    id?: string,
    namespace?: string
    group?: string
    name?: string
}

export type DeleteConfigFileGroupResponse = {}

export async function deleteConfigFileGroups(params: DeleteConfigFileGroupRequest[]) {
    const res = await apiRequest<DeleteConfigFileGroupResponse>({
        action: `${BaseURL.CONFIG_GROUP}/delette`,
        data: params.map(toApiConfigFileGroup),
    })
    return res
}
