import { apiRequest, getApiRequest, putApiRequest } from 'utils/request';
import { BaseURL, Label, MatcheLabel } from './types';

export interface ConfigFileRelease {
    id: string
    name: string
    namespace: string
    group: string
    fileName: string
    content: string
    comment: string
    md5: string
    version: string
    tags: Label[]
    releaseDescription?: string
    releaseStatus?: string
    grayPriority?: number
    active: boolean
    releaseType?: string
    format: string
    betaLabels: MatcheLabel[]
}

export interface ConfigFileReleaseView extends ConfigFileRelease {
    createTime: string
    createBy: string
    modifyTime: string
    modifyBy: string
}

export interface ConfigFileReleaseHistory {
    id: string
    name: string
    namespace: string
    group: string
    fileName: string
    content: string
    format: string
    comment: string
    md5: string
    type: string
    status: string
    tags: Label[]
    createTime: string
    createBy: string
    modifyTime: string
    modifyBy: string
    releaseDescription?: string
    releaseStatus?: string
    releaseReason?: string
}

/** 配置发布版本信息 */
export interface ReleaseVersion {
    namespace: string
    group: string
    fileName: string
    releaseType: string
    tags: Label[]
    // 名称
    name?: string
    // 是否生效
    active?: boolean
}

type ApiConfigFileRelease = ConfigFileReleaseView & {
    labels?: Record<string, string> | Label[]
    ctime?: string
    mtime?: string
    create_by?: string
    modify_by?: string
    file_name?: string
    release_description?: string
    release_type?: string
    beta_labels?: MatcheLabel[]
    release_status?: string
    gray_priority?: number
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

function normalizeMatchStringForView(value: any) {
    if (!value) {
        return value;
    }
    return {
        ...value,
        value_type: value.value_type || value.valueType || 'TEXT',
    };
}

function normalizeMatchStringForApi(value: any) {
    if (!value) {
        return value;
    }
    return {
        ...value,
        valueType: value.valueType || value.value_type || 'TEXT',
    };
}

function normalizeClientLabelsForView(labels?: MatcheLabel[]): MatcheLabel[] {
    return (labels || []).map((label: any) => ({
        ...label,
        value: normalizeMatchStringForView(label.value),
    }));
}

function normalizeClientLabelsForApi(labels?: MatcheLabel[]): MatcheLabel[] | undefined {
    if (!labels) {
        return undefined;
    }
    return labels.map((label: any) => ({
        ...label,
        value: normalizeMatchStringForApi(label.value),
    }));
}

function normalizeConfigFileRelease(release: ApiConfigFileRelease): ConfigFileReleaseView {
    return {
        ...release,
        fileName: release.fileName || release.file_name || '',
        tags: release.tags || labelsToTags(release.labels),
        createTime: release.createTime || release.ctime || release.modifyTime || release.mtime || '',
        createBy: release.createBy || release.create_by || '',
        modifyTime: release.modifyTime || release.mtime || '',
        modifyBy: release.modifyBy || release.modify_by || '',
        releaseDescription: release.releaseDescription || release.release_description,
        releaseType: release.releaseType || release.release_type,
        releaseStatus: release.releaseStatus || release.release_status,
        grayPriority: release.grayPriority || release.gray_priority,
        active: Boolean(release.active),
        betaLabels: normalizeClientLabelsForView(release.betaLabels || release.beta_labels),
    }
}

function toApiConfigFileRelease(release: ReleaseConfigFilePequest | RollbackFileReleasesResquest | DeleteFileReleaseRequest | StopGrayFileReleaseRequest | PromoteGrayFileReleaseRequest) {
    const {
        tags,
        betaLabels,
        fileName,
        releaseDescription,
        releaseType,
        grayPriority,
        ...rest
    } = release as ReleaseConfigFilePequest & { tags?: Label[] };
    return {
        ...rest,
        file_name: fileName,
        release_description: releaseDescription,
        release_type: releaseType,
        gray_priority: grayPriority,
        labels: tagsToLabels(tags),
        beta_labels: normalizeClientLabelsForApi(betaLabels),
    };
}

// 发布配置文件
export interface ReleaseConfigFilePequest {
    namespace: string
    group: string
    fileName: string
    name?: string
    releaseDescription?: string
    betaLabels?: MatcheLabel[]
    releaseType?: string
    grayPriority?: number
}

export interface ReleaseConfigFileResponse {
    configFileRelease: ConfigFileRelease
}

export async function releaseConfigFile(params: ReleaseConfigFilePequest) {
    const res = await apiRequest<ReleaseConfigFileResponse>({
        action: `${BaseURL.CONFIG_RELEASE}`,
        data: toApiConfigFileRelease(params),
    })
    return res
}

// 查询配置文件发布详情
export interface DescribeOneFileReleaseRequest {
    // 命名空间名称
    namespace: string
    // 配置分组名称
    group: string
    // 配置文件名称
    file_name: string
    // 配置文件版本
    release_name?: string
}

export interface DescribeOneFileReleaseResponse {
    /** 配置文件发布详情 */
    configFileRelease: ConfigFileReleaseView
}

export async function describeOneFileRelease(params: DescribeOneFileReleaseRequest) {
    const res = await getApiRequest<DescribeOneFileReleaseResponse>({
        action: `${BaseURL.CONFIG_RELEASE}`,
        data: params,
    })
    const configFileRelease = res.configFileRelease || normalizeConfigFileRelease(res as unknown as ApiConfigFileRelease)
    return {
        ...res,
        configFileRelease: normalizeConfigFileRelease(configFileRelease as ApiConfigFileRelease),
    }
}

// 查询配置文件发布历史
export interface DescribeConfigFileReleasesRequest {
    // 命名空间
    namespace?: string
    // 配置分组
    group?: string
    // 文件名称
    file_name?: string
    // 只保护处于使用状态
    only_use?: boolean
    only_active?: boolean
    // 发布名称
    release_name?: string
    // 条数
    limit: number
    // 偏移量
    offset: number
}

export interface DescribeConfigFileReleasesResponse {
    configFileReleases: ConfigFileRelease[]
    total: number
}

export async function describeFileReleases(params: DescribeConfigFileReleasesRequest) {
    const { only_use, ...rest } = params;
    const res = await getApiRequest<DescribeConfigFileReleasesResponse>({
        action: `${BaseURL.CONFIG_RELEASES}`,
        data: {
            ...rest,
            only_active: rest.only_active ?? only_use,
        },
    })
    const releases = ((res as any).data ?? res.configFileReleases ?? []).map(normalizeConfigFileRelease)
    return {
        ...res,
        configFileReleases: releases,
        total: res.total ?? (res as any).amount ?? releases.length,
    }
}

// 查询某个配置文件的发布的版本记录
export interface DescribeFileReleaseVersionsRequest {
    // 命名空间
    namespace?: string
    // 配置分组
    group?: string
    // 文件名称
    file_name?: string
}

export interface DescribeFileReleaseVersionsResponse {
    // 版本信息
    configFileReleases?: ReleaseVersion[]
}

export async function describeFileReleaseVersions(params: DescribeFileReleaseVersionsRequest) {
    const res = await getApiRequest<DescribeFileReleaseVersionsResponse>({
        action: `${BaseURL.CONFIG_RELEASE}/versions`,
        data: params,
    })
    const releases = ((res as any).data ?? res.configFileReleases ?? []).map(normalizeConfigFileRelease)
    return {
        ...res,
        configFileReleases: releases,
    }
}

// 回滚配置发布
export interface RollbackFileReleasesResquest {
    namespace: string
    group: string
    fileName: string
    name: string
}

export interface RollbackFileReleasesResponse {
    code: number
    info: string
}

export async function rollbackFileReleases(params: RollbackFileReleasesResquest[]) {
    const res = await putApiRequest<RollbackFileReleasesResponse>({
        action: `${BaseURL.CONFIG_RELEASES}/rollback`,
        data: params.map(toApiConfigFileRelease),
    })
    return res
}

// 删除已发布的配置
export interface DeleteFileReleaseRequest {
    // 命名空间
    namespace?: string
    // 配置分组
    group?: string
    // 文件名称
    fileName?: string
    // 发布名称
    name?: string
    releaseType?: string
}

export interface DeleteConfigFileReleasesResponse {
    /** 删除配置发布结果 */
    result?: boolean
}

export async function deleteFileReleases(params: DeleteFileReleaseRequest[]) {
    const res = await apiRequest<DeleteConfigFileReleasesResponse>({
        action: `${BaseURL.CONFIG_RELEASES}/delete`,
        data: params.map(toApiConfigFileRelease),
    })
    return res
}

export interface StopGrayFileReleaseRequest {
    namespace: string
    group: string
    fileName: string
    name?: string
    releaseType?: string
}

export async function stopGrayFileReleases(params: StopGrayFileReleaseRequest[]) {
    const res = await apiRequest<DeleteConfigFileReleasesResponse>({
        action: `${BaseURL.CONFIG_RELEASES}/stopbeta`,
        data: params.map((item) => toApiConfigFileRelease({
            ...item,
            releaseType: item.releaseType || 'gray',
        })),
    })
    return res
}

export interface PromoteGrayFileReleaseRequest {
    namespace: string
    group: string
    fileName: string
    name: string
    releaseType?: string
}

export async function promoteGrayFileReleaseToDraft(params: PromoteGrayFileReleaseRequest) {
    const res = await apiRequest<RollbackFileReleasesResponse>({
        action: `${BaseURL.CONFIG_RELEASES}/promote-gray`,
        data: toApiConfigFileRelease({
            ...params,
            releaseType: params.releaseType || 'gray',
        }),
    })
    return res
}

// DescribeFileSubscribersRequest 查询文件订阅者
export interface DescribeFileSubscribersRequest {
    // 命名空间
    namespace: string
    // 配置分组
    group: string
    // 文件名称
    file_name: string
}

export interface VersionClient {
    version: number;
    subscribers: FileSubscriber[];
}

export interface FileSubscriber {
    id: string;
    host: string;
    release_name: string;
    version: string;
    client_type: string;
}

export interface FileSubscriptionInfo {
    name: string;
    namespace: string;
    group: string;
    file_name: string;
    release_type: string;
    version: number;
    release_name: string;
}

export interface DescribeFileSubscribersResponse {
    // 订阅者信息
    clients?: VersionClient[]
}

export async function describeFileSubscribers(params: DescribeFileSubscribersRequest) {
    const res = await getApiRequest<DescribeFileSubscribersResponse>({
        action: `${BaseURL.CONFIG_FILE}/subscribers`,
        data: params,
    })
    return res
}
