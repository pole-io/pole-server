import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';

export enum USER_ROLE {
    ADMIN = 'main',
    SUB_USER = 'sub',
}

export const USER_ROLE_MAP = {
    [USER_ROLE.ADMIN]: {
        text: '管理员',
        theme: 'primary',
    },
    [USER_ROLE.SUB_USER]: {
        text: '子用户',
        theme: 'primary',
    },
}

// 用户
export interface User {
    // 用户ID
    id: string
    // 用户名称
    name: string
    // 用户来源
    source: string
    // 用户鉴权Token
    auth_token: string
    // 该token是否被禁用
    token_enable: boolean
    // 该账户的简单描述
    comment: string
    // 该账户的创建时间
    ctime: string
    // 该账户的最近一次修改时间
    mtime: string
    // 账户对应邮箱
    email: string
    // 账户对应手机号
    mobile: string
    // 用户标签
    metadata: Record<string, string>
    // 用户类型
    user_type?: string
}

// 删除用户
export interface DeleteUsersRequest {
    // 用户ID列表
    id: string
}

export interface DeleteUsersResponse {
    // 执行结果
    responses: ApiResponse[]
}

export async function deleteUsers(params: DeleteUsersRequest[]) {
    const result = await apiRequest<DeleteUsersResponse>({ action: '/auth/v1/users/delete', data: params })
    return result.responses.every(item => Number(item.code) === SuccessCode)
}

// 查询用户列表
export interface DescribeUsersRequest {
    // 用户id
    id?: string
    // 用户名称，模糊搜索最后加上 * 字符
    name?: string
    // 主账户ID
    owner?: string
    // 分页偏移量
    offset: number
    // 查询条数
    limit: number
    // 账户来源，QCloud | Polaris
    source?: string
    // 用户组ID
    group_id?: string
}

export interface DescribeUsersResponse {
    // 总数
    amount: number
    // 用户列表
    users: User[]
}

export async function describeUsers(params: DescribeUsersRequest) {
    const result = await getApiRequest<DescribeUsersResponse>({
        action: '/auth/v1/users',
        data: params,
    })
    return {
        totalCount: result.amount,
        content: result.users ? result.users : [],
    }
}

export async function describeAllUsers() {
    const { list: users } = await getAllList(describeUsers, {
        listKey: 'content',
        totalKey: 'totalCount',
    })({})
    return users
}

// 查询用户Token
export interface DescribeUserTokenRequest {
    // 用户ID
    id: string
}

export interface DescribeUserTokenResponse {
    // 用户
    user: User
}

export async function describeUserToken(params: DescribeUserTokenRequest) {
    const result = await getApiRequest<DescribeUserTokenResponse>({ action: '/auth/v1/user/token', data: params })
    return result
}

// 修改用户信息
export interface ModifyUserRequest {
    // 用户
    id: string
    mobile?: string
    email?: string
    comment?: string
}

export interface ModifyUserResponse {
    // 请求结果
    result: boolean
}

export async function modifyUsers(params: ModifyUserRequest[]) {
    const result = await putApiRequest<ModifyUserResponse>({ action: '/auth/v1/users', data: params })
    return Number(result.code) === SuccessCode
}

// 修改用户密码
export interface ModifyUserPasswordRequest {
    // 用户
    id: string
    old_password?: string
    new_password: string
}

export interface ModifyUserPasswordResponse {
    // 请求结果
    result: boolean
}

export async function modifyUserPassword(params: ModifyUserPasswordRequest) {
    const result = await putApiRequest<ModifyUserPasswordResponse>({
        action: '/auth/v1/user/password',
        data: params,
    })
    return Number(result.code) === SuccessCode
}

// 修改用户Token
export interface ModifyUserTokenRequest {
    // 用户Token信息
    id: string
    token_enable: boolean
}

export interface ModifyUserTokenResponse {
    // 执行结果
    result: boolean
}

export async function modifyUserToken(params: ModifyUserTokenRequest) {
    const result = await putApiRequest<ModifyUserTokenResponse>({
        action: '/auth/v1/user/token/enable',
        data: params,
    })
    return Number(result.code) === SuccessCode
}

// 修改用户Token
export interface RefreshUserTokenRequest {
    // 用户Token信息
    id: string
}

export interface RefreshUserTokenResponse {
    // 执行结果
    result: boolean
}

export async function refreshUserToken(params: RefreshUserTokenRequest) {
    const result = await putApiRequest<RefreshUserTokenResponse>({
        action: '/auth/v1/user/token/refresh',
        data: params,
    })
    return Number(result.code) === SuccessCode
}

// 批量创建用户
export interface CreateUsersRequest {
    // 用户列表
    name: string
    password: string
    comment: string
    mobile?: string
    email?: string
    source?: string
}

export interface CreateUsersResponse {
    // 请求结果
    responses: ApiResponse[]
}

export async function createUsers(params: CreateUsersRequest[]) {
    const result = await apiRequest<CreateUsersResponse>({ action: '/auth/v1/users', data: params })
    return result.responses.every(item => Number(item.code) === SuccessCode)
}

// 重置用户的资源访问 token
export interface ResetUserTokenRequest {
    // 用户ID
    id: string
}

export interface ResetUserTokenResponse {
    // 执行结果
    result: boolean
}

export async function resetUserToken(params: ResetUserTokenRequest) {
    const result = await putApiRequest<ResetUserTokenResponse>({
        action: '/auth/v1/user/token/refresh',
        data: params,
    })
    return Number(result.code) === SuccessCode
}
