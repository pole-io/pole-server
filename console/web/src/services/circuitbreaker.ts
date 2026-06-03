import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';
import { API, BaseURL, Label, RuleRelease } from './types';

export interface BlockConfig {
    name: string
    api?: API
    error_conditions: ErrorCondition[]
    trigger_conditions: TriggerCondition[]
}

export interface CircuitBreakerRule {
    id?: string
    name: string // 规则名
    level: string
    description: string
    priority: number
    ruleMatcher: {
        source: {
            service: string
            namespace: string
        }
        destination: {
            service: string
            namespace: string
            method: {
                type: string
                value: string
            }
        }
    }
    block_configs: BlockConfig[]
    recoverCondition: RecoverCondition
    faultDetectConfig: FaultDetectConfig
    fallbackConfig: FallbackConfig
    metadata?: Record<string, string>
    ctime?: string
    mtime?: string
    etime?: string
    editable?: boolean
    deleteable?: boolean
}

export interface ErrorCondition {
    inputType: string
    condition: {
        type: string
        value: string
    }
}

export interface TriggerCondition {
    // 触发类型：ERROR_RATE错误率，CONSECUTIVE_ERROR连续错误数
    triggerType: string
    triggerVal?: number // 错误率或连续错误数
    errorCount: number
    errorPercent: number
    interval: number
    minimumRequest: number
}

export interface RecoverCondition {
    sleepWindow?: number
    consecutiveSuccess?: number
}

export interface FaultDetectConfig {
    enable: boolean
}

export interface FallbackConfig {
    enable: boolean
    response: {
        code: number
        headers: Label[]
        body: string
    }
}

export enum ErrorConditionType {
    RET_CODE = 'RET_CODE',
    DELAY = 'DELAY',
}

export const ErrorConditionMap = {
    [ErrorConditionType.DELAY]: '时延',
    [ErrorConditionType.RET_CODE]: '返回码',
}

export const ErrorConditionOptions = Object.entries(ErrorConditionMap).map(([key, value]) => ({
    label: value,
    value: key,
}))

export enum TriggerType {
    ERROR_RATE = 'ERROR_RATE',
    CONSECUTIVE_ERROR = 'CONSECUTIVE_ERROR',
}

export const TriggerTypeMap = {
    [TriggerType.CONSECUTIVE_ERROR]: { text: '连续错误数', unit: '个' },
    [TriggerType.ERROR_RATE]: { text: '错误率', unit: '%' },
}

export const TriggerTypeOptions = Object.entries(TriggerTypeMap).map(([key, value]) => ({
    label: value.text,
    value: key,
}))

export enum BreakLevelType {
    Instance = 'INSTANCE',
    Group = 'GROUP',
    Method = 'METHOD',
    Service = 'SERVICE',
}

export const BreakLevelMap = {
    [BreakLevelType.Instance]: '实例',
    [BreakLevelType.Group]: '实例分组',
    [BreakLevelType.Method]: '接口',
    [BreakLevelType.Service]: '服务',
}

export const BreakLevelSearchParamMap = {
    [BreakLevelType.Instance]: 4,
    [BreakLevelType.Group]: 3,
    [BreakLevelType.Method]: 2,
    [BreakLevelType.Service]: 1,
}

export const ServiceLevelType = [BreakLevelType.Method, BreakLevelType.Service]

export const InterfaceLevelType = [BreakLevelType.Instance]

export const ServiceBreakLevelOptions = Object.entries(BreakLevelMap)
    .filter(([key]) => ServiceLevelType.indexOf(key as any) > -1)
    .map(([key, value]) => ({
        label: value,
        value: key,
    }))

export const InterfaceBreakLevelOptions = Object.entries(BreakLevelMap)
    .filter(([key]) => InterfaceLevelType.indexOf(key as any) > -1)
    .map(([key, value]) => ({
        text: value,
        value: key,
    }))

export enum BreakerType {
    Service = 'Service',
    Interface = 'Interface',
    FaultDetect = 'FaultDetect',
}

export const checkRuleType = (level: any) =>
    ServiceLevelType.indexOf(level as any) > -1
        ? BreakerType.Service
        : InterfaceLevelType.indexOf(level as any) > -1
            ? BreakerType.Interface
            : BreakerType.FaultDetect

export const FaultDetectTabs = [
    { id: BreakerType.Service, label: '服务级熔断' },
    { id: BreakerType.Interface, label: '节点级熔断' },
    { id: BreakerType.FaultDetect, label: '主动探测' },
]



export interface DescribeCircuitBreakersRequest {
    brief: boolean
    offset: number
    limit: number
    id?: string
    name?: string
    enable?: boolean
    level?: number
    service?: string
    serviceNamespace?: string
    srcService?: string
    srcNamespace?: string
    dstService?: string
    dstNamespace?: string
    dstMethod?: string
    description?: string
}

export interface DescribeCircuitBreakersResponse {
    data: CircuitBreakerRule[]
    amount: number
    size: number
}

export async function describeCircuitBreakers(params: DescribeCircuitBreakersRequest) {
    const res = await getApiRequest<DescribeCircuitBreakersResponse>({
        action: `${BaseURL.CIRCUIT_BREAKER}`,
        data: params,
    })
    return {
        list: res.data,
        totalCount: res.amount,
    }
}

/** 解包后即为规则详情 */
export interface DescribeOneCircuitBreakerResponse {
    data: CircuitBreakerRule
}

export async function describeOneCircuitBreaker(id: string) {
    return await getApiRequest<CircuitBreakerRule>({
        action: `${BaseURL.CIRCUIT_BREAKER}/detail`,
        data: { id },
    })
}

export type CreateCircuitBreakerRequest = CircuitBreakerRule

export async function createCircuitBreaker(params: CreateCircuitBreakerRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}`,
        data: params,
    })
    return res
}

export type ModifyCircuitBreakerRequest = CircuitBreakerRule

export async function modifyCircuitBreaker(params: ModifyCircuitBreakerRequest[]) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}`,
        data: params,
    })
    return res
}

export interface DeleteCircuitBreakerRequest {
    id: string
}

export async function deleteCircuitBreaker(params: DeleteCircuitBreakerRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}/delete`,
        data: params,
    })
    return res
}

// publishCircuitBreaker 发布一个无损规则
export async function publishCircuitBreaker(params: RuleRelease[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}/releases`,
        data: params,
    })
    return res
}

export interface DescribeCircuitBreakerVersionsRequest {
    rule_name: string
    offset: number
    limit: number
}

export interface DescribeCircuitBreakerVersionsResponse {
    data: RuleRelease[]
    amount: number
}

export async function describeCircuitBreakerVersions(params: DescribeCircuitBreakerVersionsRequest) {
    const res = await getApiRequest<DescribeCircuitBreakerVersionsResponse>({
        action: `${BaseURL.CIRCUIT_BREAKER}/releases`,
        data: params,
    })
    return {
        list: res.data,
        totalCount: res.amount,
    }
}

export interface RollbackCircuitBreakerReleaseRequest {
    id: string
}

// rollbackCircuitBreaker 回滚某个规则版本
export async function rollbackCircuitBreaker(params: RollbackCircuitBreakerReleaseRequest) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}/releases/rollback`,
        data: [params],
    });
    return res;
}

export interface DeleteCircuitBreakerReleaseRequest {
    id: string
}

// deleteCircuitBreakerRelease 删除某个规则版本
export async function deleteCircuitBreakerRelease(params: DeleteCircuitBreakerReleaseRequest) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}/releases/delete`,
        data: [params],
    });
    return res;
}