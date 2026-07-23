import { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
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

type ApiConfigFile = ConfigFileView & {
    labels?: Record<string, string> | Label[]
    ctime?: string
    mtime?: string
    rtime?: string
    encrypt_algo?: string
}

function labelsToTags(labels?: Record<string, string> | Label[]): Label[] {
    if (!labels) {
        return [];
    }
    if (Array.isArray(labels)) {
        return labels;
    }
    return Object.entries(labels).map(([key, value]) => ({ key, value }));
}

function tagsToLabels(tags?: Label[] | Record<string, string>) {
    if (!tags) {
        return undefined;
    }
    if (!Array.isArray(tags)) {
        return tags;
    }
    return tags.reduce((acc: Record<string, string>, item) => {
        if (item.key) {
            acc[item.key] = item.value;
        }
        return acc;
    }, {});
}

function normalizeConfigFile(file: ApiConfigFile): ConfigFileView {
    return {
        ...file,
        tags: file.tags || labelsToTags(file.labels),
        createTime: file.createTime || file.ctime,
        modifyTime: file.modifyTime || file.mtime,
        releaseTime: file.releaseTime || file.rtime,
        encryptAlgo: file.encryptAlgo || file.encrypt_algo,
    }
}

function toApiConfigFile(file: CreateConfigFileRequest | ModifyConfigFileRequest) {
    const { tags, createTime, modifyTime, releaseTime, releaseBy, editable, deleteable, ...rest } = file as ConfigFileView;
    return {
        ...rest,
        labels: tagsToLabels(tags),
    };
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
        data: params.map(toApiConfigFile),
    })
    return res
}

// 查询配置文件列表
export interface DescribeConfigFilesRequest {
    offset: number
    limit: number
    namespace?: string
    group: string
    name?: string
    tags?: string
    brief?: boolean
}

export interface DescribeConfigFilesResponse {
    amount?: number
    size?: number
    total?: number
    data?: Array<ConfigFile>
    configFiles?: Array<ConfigFile>
}

export async function describeConfigFiles(params: DescribeConfigFilesRequest) {
    const res = await getApiRequest<DescribeConfigFilesResponse>({
        action: `${BaseURL.CONFIG_FILE}/search`,
        data: params,
    })
    const files = res.data ?? res.configFiles ?? []
    const normalizedFiles = files.map(normalizeConfigFile)
    return {
        list: normalizedFiles,
        totalCount: res.amount ?? res.total ?? files.length,
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
    })({ ...params, brief: true })
    return users
}

/** 查询相同分组、相同文件名在用户有权访问的各个环境中的摘要。 */
export async function describeConfigFileEnvironments(group: string, name: string) {
    const res = await describeConfigFiles({
        offset: 0,
        limit: 100,
        group,
        name,
        brief: true,
    })
    return res.list.filter((file) => file.group === group && file.name === name)
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
    const configFile = res.configFile || normalizeConfigFile(res as unknown as ApiConfigFile)
    return {
        ...res,
        configFile: normalizeConfigFile(configFile as ApiConfigFile),
    }
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
        data: params.map(toApiConfigFile),
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

interface EncryptAlgorithmApiResponse extends EncryptAlgorithmResponse {
    value?: {
        algorithms?: string[]
    }
}

function normalizeEncryptAlgorithms(res: EncryptAlgorithmApiResponse) {
    const algorithms = Array.isArray(res.algorithms) ? res.algorithms : res.value?.algorithms;
    return Array.isArray(algorithms) ? algorithms : [];
}

export async function describeEncryptAlgo() {
    const res = await getApiRequest<EncryptAlgorithmApiResponse>({
        action: `${BaseURL.CONFIG_FILE}/encrypt/algorithms`,
    })
    return {
        ...res,
        algorithms: normalizeEncryptAlgorithms(res),
    }
}
