import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';
import { BaseURL, Label } from './types';

export enum FileStatus {
    Normal = 'normal',
    Success = 'success',
    Fail = 'failure',
    Edited = 'to-be-released',
    Betaing = 'gray',
}

export const FileStatusMap = {
    [FileStatus.Normal]: {
        text: '发布成功',
        theme: 'success',
    },
    [FileStatus.Success]: {
        text: '发布成功',
        theme: 'success',
    },
    [FileStatus.Fail]: {
        text: '发布失败',
        theme: 'danger',
    },
    [FileStatus.Betaing]: {
        text: '灰度发布中',
        theme: 'warning',
    },
    [FileStatus.Edited]: {
        text: '编辑待发布',
        theme: 'warning',
    },
}

export interface ConfigFile {
    id?: number
    name: string
    namespace: string
    group: string
    content?: string
    format?: string
    comment?: string
    status?: string
    tags?: Array<Label>
    encrypted?: boolean
    encryptAlgo?: string
}

export interface ConfigFileView extends ConfigFile {
    createTime?: string
    createBy?: string
    modifyTime?: string
    modifyBy?: string
    releaseTime?: string
    releaseBy?: string
    editable?: boolean
    deleteable?: boolean
}

// 创建配置文件
export interface CreateConfigFileRequest {
    id?: number,
    name: string
    namespace: string
    group: string
    content: string
    format: string
    comment: string
    tags: Array<Label>
    encrypted: boolean
    encryptAlgo: string
}

export interface CreateConfigFileResponse {
    configFile: ConfigFile
}

export async function createConfigFiles(params: CreateConfigFileRequest[]) {
    const res = await apiRequest<CreateConfigFileResponse>({
        action: `${BaseURL.CONFIG_FILE}`,
        data: params,
    })
    return res
}

// 查询配置文件列表
export interface DescribeConfigFilesRequest {
    offset: number
    limit: number
    namespace: string
    group: string
    name?: string
    tags?: string
}

export interface DescribeConfigFilesResponse {
    total: number
    configFiles: Array<ConfigFile>
}

export async function describeConfigFiles(params: DescribeConfigFilesRequest) {
    const res = await getApiRequest<DescribeConfigFilesResponse>({
        action: `${BaseURL.CONFIG_FILE}/search`,
        data: params,
    })
    return {
        list: res.configFiles,
        totalCount: res.total,
    }
}

// 查询所有配置文件
export interface DescribeAllConfigFilesRequest {
    namespace: string
    group: string
}

export async function describeAllConfigFiles(params: DescribeAllConfigFilesRequest) {
    const { list: users } = await getAllList(describeConfigFiles, {
        listKey: 'list',
        totalKey: 'totalCount',
    })({ ...params, berif: true })
    return users
}

// describeOneConfigFile 查询单个配置文件
export interface DescribeOneConfigFileRequest {
    id?: number
    namespace?: string
    group?: string
    name?: string
}

export interface DescribeOneConfigFileResponse {
    configFile: ConfigFile
}


export async function describeOneConfigFile(params: DescribeOneConfigFileRequest) {
    const res = await getApiRequest<DescribeOneConfigFileResponse>({
        action: `${BaseURL.CONFIG_FILE}/detail`,
        data: params,
    })
    return res
}

// ModifyConfigFileRequest 修改配置文件
export interface ModifyConfigFileRequest {
    id?: number
    name: string
    namespace: string
    group: string
    content?: string
    comment?: string
    tags?: Array<Label>
    format?: string
    encrypted: boolean
    encryptAlgo: string
}
export interface ModifyConfigFileResponse {
    configFile: ConfigFile
}

export async function modifyConfigFiles(params: ModifyConfigFileRequest[]) {
    const res = await putApiRequest<ModifyConfigFileResponse>({
        action: `${BaseURL.CONFIG_FILE}`,
        data: params,
    })
    return res
}

// 删除配置文件
export interface DeleteConfigFileRequest {
    id?: number
    namespace?: string
    group?: string
    name?: string
}

export type DeleteConfigFileResponse = {}

export async function deleteConfigFiles(params: DeleteConfigFileRequest[]) {
    const res = await apiRequest<DeleteConfigFileResponse>({
        action: `${BaseURL.CONFIG_FILE}/delete`,
        data: params,
    })
    return res
}

// 获取加密算法支持列表
export interface EncryptAlgorithmResponse {
    algorithms: string[]
}

export async function describeEncryptAlgo() {
    const res = await getApiRequest<EncryptAlgorithmResponse>({
        action: `${BaseURL.CONFIG_FILE}/encrypt/algorithms`,
    })
    return res
}