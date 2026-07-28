import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';
import { API, BaseURL, Label, RuleRelease } from './types';
import { RollbackCircuitBreakerReleaseRequest } from './circuitbreaker';


export interface FaultDetectRule {
    id?: string
    namespace?: string // 规则归属环境
    name: string
    description: string
    priority?: number
    targetService: {
        namespace: string
        service: string
        api?: API
    }
    rules?: FaultDetectSubRule[]
    ctime?: string
    mtime?: string
    metadata?: Record<string, string>
    editable?: boolean
    deleteable?: boolean
}

export interface FaultDetectSubRule {
    interval: number
    timeout: number
    port: number
    protocol: string
    httpConfig?: {
        method: string
        url: string
        headers: Label[]
        body: string
    }
    tcpConfig?: {
        send: string
        receive: string[] | string
        match?: string
    }
    udpConfig?: {
        send: string
        receive: string[] | string
        match?: string
    }
    disable?: boolean
}

function normalizeProtocol(protocol?: string | number): string {
    const map: Record<number, FaultDetectProtocol> = {
        1: FaultDetectProtocol.HTTP,
        2: FaultDetectProtocol.TCP,
        3: FaultDetectProtocol.UDP,
    };
    if (typeof protocol === 'number') return map[protocol] || FaultDetectProtocol.HTTP;
    return protocol || FaultDetectProtocol.HTTP;
}

function normalizeMatchType(type?: string | number): string {
    const map: Record<number, string> = {
        0: 'EXACT',
        1: 'REGEX',
        2: 'NOT_EQUALS',
        3: 'IN',
        4: 'NOT_IN',
        5: 'RANGE',
    };
    if (typeof type === 'number') return map[type] || 'EXACT';
    return type || 'EXACT';
}

export function normalizeFaultDetectRule(rule: FaultDetectRule | any): FaultDetectRule {
    if (!rule) return rule;
    const rawRules = rule.rules || [];
    const rules = rawRules.map((item: any) => normalizeFaultDetectSubRule(item));
    const primaryRule = rules[0] || {};
    const legacyFirstRuleTarget = rawRules[0]?.targetService ?? rawRules[0]?.target_service;
    const targetService = rule.targetService ?? rule.target_service ?? legacyFirstRuleTarget ?? {};
    const api = targetService.api ?? {};
    return {
        ...rule,
        rules,
        targetService: {
            namespace: targetService.namespace || '',
            service: targetService.service || '',
            api: {
                protocol: api.protocol || 'HTTP',
                method: api.method || '',
                path: {
                    type: normalizeMatchType(api.path?.type),
                    value: api.path?.value || targetService.method?.value || '',
                    value_type: api.path?.value_type || api.path?.valueType || 'TEXT',
                },
            },
        },
        interval: primaryRule.interval ?? rule.interval,
        timeout: primaryRule.timeout ?? rule.timeout,
        port: primaryRule.port ?? rule.port,
        protocol: normalizeProtocol(primaryRule.protocol ?? rule.protocol),
        httpConfig: primaryRule.httpConfig ?? rule.httpConfig ?? rule.http_config,
        tcpConfig: primaryRule.tcpConfig ?? rule.tcpConfig ?? rule.tcp_config,
        udpConfig: primaryRule.udpConfig ?? rule.udpConfig ?? rule.udp_config,
    };
}

function normalizeFaultDetectSubRule(rule: any): FaultDetectSubRule {
    return {
        interval: rule.interval,
        timeout: rule.timeout,
        port: rule.port,
        protocol: normalizeProtocol(rule.protocol),
        httpConfig: rule.httpConfig ?? rule.http_config,
        tcpConfig: rule.tcpConfig ?? rule.tcp_config,
        udpConfig: rule.udpConfig ?? rule.udp_config,
        disable: rule.disable,
    };
}

export enum FaultDetectProtocol {
    HTTP = 'HTTP',
    TCP = 'TCP',
    UDP = 'UDP',
}

export const FaultDetectProtocolOptions = Object.keys(FaultDetectProtocol).map(item => ({ text: item, value: item }))

export enum FaultDetectHttpMethod {
    GET = 'GET',
    POST = 'POST',
    PUT = 'PUT',
    OPTION = 'OPTION',
    DELETE = 'DELETE',
    PATCH = 'PATCH',
    HEAD = 'HEAD',
    CONNECT = 'CONNECT',
    TRACE = 'TRACE',
}

export const FaultDetectHttpMethodOptions = Object.keys(FaultDetectHttpMethod).map(item => ({
    label: item,
    value: item,
}))

export const BlockHttpBodyMethod = [
    FaultDetectHttpMethod.GET,
    FaultDetectHttpMethod.DELETE,
    FaultDetectHttpMethod.HEAD,
] as string[]


export interface DescribeFaultDetectsRequest {
    brief: boolean
    offset: number
    limit: number
    id?: string
    name?: string
    namespace?: string
    service?: string
    serviceNamespace?: string
    dstService?: string
    dstNamespace?: string
    dstMethod?: string
    description?: string
}

export interface DescribeFaultDetectsResponse {
    data: FaultDetectRule[]
    amount: number
    size: number
}

export async function describeFaultDetects(params: DescribeFaultDetectsRequest) {
    const res = await getApiRequest<DescribeFaultDetectsResponse>({
        action: `${BaseURL.FAULT_DETECT}`,
        data: params,
    })
    return {
        list: (res.data || []).map(normalizeFaultDetectRule),
        totalCount: res.amount,
    }
}

export interface DescribeOneFaultDetectResponse {
    data: FaultDetectRule[]
    amount: number
    size: number
}

export async function describeOneFaultDetect(id: string) {
    const res = await getApiRequest<DescribeOneFaultDetectResponse>({
        action: `${BaseURL.FAULT_DETECT}/detail`,
        data: { id },
    })
    return {
        list: (res.data || []).map(normalizeFaultDetectRule),
        totalCount: res.amount,
    }
}

export type CreateFaultDetectRequest = FaultDetectRule

export async function createFaultDetects(params: CreateFaultDetectRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.FAULT_DETECT}`,
        data: params,
    })
    return res
}

export type ModifyFaultDetectRequest = FaultDetectRule

export async function modifyFaultDetects(params: CreateFaultDetectRequest[]) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.FAULT_DETECT}`,
        data: params,
    })
    return res
}

export interface DeleteFaultDetectParams {
    id: string
}

export async function deleteFaultDetects(params: DeleteFaultDetectParams[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.FAULT_DETECT}/delete`,
        data: params,
    })
    return res
}

export interface DescribeFaultDetectVersionsRequest {
    id: string
    offset: number
    limit: number
}

export interface DescribeFaultDetectVersionsResponse {
    data: RuleRelease[]
    amount: number
}

export async function describeFaultDetectVersions(params: DescribeFaultDetectVersionsRequest) {
    const res = await getApiRequest<DescribeFaultDetectVersionsResponse>({
        action: `${BaseURL.FAULT_DETECT}/releases`,
        data: params,
    })
    return {
        list: res.data,
        totalCount: res.amount,
    }
}

// publishFaultDetect 发布一个无损规则
export async function publishFaultDetect(params: RuleRelease[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.FAULT_DETECT}/releases`,
        data: params,
    })
    return res
}
export interface RollbackFaultDetectReleaseRequest {
    id: string
}

// rollbackFaultDetect 回滚某个规则版本
export async function rollbackFaultDetect(params: RollbackFaultDetectReleaseRequest) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.FAULT_DETECT}/releases/rollback`,
        data: [params],
    });
    return res;
}

export interface DeleteFaultDetectReleaseRequest {
    id: string
}

// deleteFaultDetectRelease 删除某个规则版本
export async function deleteFaultDetectRelease(params: DeleteFaultDetectReleaseRequest) {
    const res = await apiRequest<any>({
        action: `${BaseURL.FAULT_DETECT}/releases/delete`,
        data: [params],
    });
    return res;
}
