import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';
import { User } from './users';
import { UserGroup } from './user_group';

export interface Role {
    // 角色ID
    id: string;
    // 角色名称
    name: string;
    // 默认角色
    default_role: boolean;
    // 来源
    source: string;
    // 备注
    comment: string;
    // 创建时间
    ctime: string;
    // 最近一次修改时间
    mtime: string;
    // 角色标签
    metadata: Record<string, string>;
    // 角色对应的用户列表
    users: {id: string, name?: string}[];
    // 角色对应的用户组列表
    user_groups: {id: string, name?: string}[];
}


// 查询用户列表
export interface DescribeRolesRequest {
    // 用户id
    id?: string
    // 用户名称，模糊搜索最后加上 * 字符
    name?: string
    // 主账户ID
    owner?: string
    // 分页偏移量
    offset?: number
    // 查询条数
    limit?: number
    // 账户来源，QCloud | Polaris
    source?: string
    // 是否展示角色详情，该参数为 true 时，返回的角色列表中包含用户和用户组列表
    berif?: boolean
}

export interface DescribeRolesResponse {
    // 总数
    amount: number
    // 角色列表
    data: Role[]
}

export async function describeRoles(params: DescribeRolesRequest) {
    const result = await getApiRequest<DescribeRolesResponse>({
        action: '/auth/v1/roles',
        data: params,
    })
    return {
        totalCount: result.amount,
        content: result.data ? result.data : [],
    }
}


export async function describeAllRoles() {
    const { list: userGroups } = await getAllList(describeRoles, {
        listKey: 'content',
        totalKey: 'totalCount',
    })({})
    return userGroups
}


// 修改角色信息
export interface CreateRoleRequest {
    // 用户
    id: string
    // 角色名称
    name: string
    // 备注
    comment?: string
    // 角色对应的用户列表
    users?: { id: string }[]
    // 角色对应的用户组列表
    user_groups?: { id: string }[]
    // 角色标签
    metadata?: Record<string, string>
    // 角色来源
    source?: string
}

export interface CreateRoleResponse {
    // 请求结果
    result: boolean
}

export async function createRoles(params: CreateRoleRequest[]) {
    const result = await apiRequest<CreateRoleResponse>({ action: '/auth/v1/roles', data: params })
    return Number(result.code) === SuccessCode
}


// 修改用户信息
export interface ModifyRoleRequest {
    // 用户
    id: string
    // 备注
    comment?: string
    // 角色对应的用户列表
    users?: { id: string }[]
    // 角色对应的用户组列表
    user_groups?: { id: string }[]
    // 角色标签
    metadata?: Record<string, string>
    // 角色来源
    source?: string
}

export interface ModifyRoleResponse {
    // 请求结果
    result: boolean
}

export async function modifyRoles(params: ModifyRoleRequest[]) {
    const result = await putApiRequest<ModifyRoleResponse>({ action: '/auth/v1/roles', data: params })
    return Number(result.code) === SuccessCode
}

// 修改用户信息
export interface DeleteRoleRequest {
    id: string
    name?: string
}

export interface DeleteRoleResponse {
    // 请求结果
    result: boolean
}

export async function deleteRoles(params: DeleteRoleRequest[]) {
    const result = await apiRequest<DeleteRoleResponse>({ action: '/auth/v1/roles/delete', data: params })
    return Number(result.code) === SuccessCode
}