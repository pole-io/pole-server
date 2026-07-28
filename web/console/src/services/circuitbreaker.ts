import request, { apiRequest, ApiResponse, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { SuccessCode } from './const';
import { API, BaseURL, Label, RuleRelease } from './types';

export interface BlockConfig {
    name: string
    api?: API
    apis?: API[]
    error_conditions: ErrorCondition[]
    trigger_conditions: TriggerCondition[]
    max_ejection_percent?: number
    recoverCondition: RecoverCondition
    faultDetectConfig: FaultDetectConfig
    fallbackConfig: FallbackConfig
}

export interface CircuitBreakerPolicy {
    block_config: Pick<BlockConfig, 'name' | 'apis' | 'error_conditions' | 'trigger_conditions'>
    max_ejection_percent?: number
    recoverCondition: RecoverCondition
    faultDetectConfig: FaultDetectConfig
    fallbackConfig: FallbackConfig
}

export type CircuitBreakerBlockConfigInput = BlockConfig | CircuitBreakerPolicy

export interface CircuitBreakerRule {
    id?: string
    namespace?: string // 规则归属环境
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
    block_configs: CircuitBreakerBlockConfigInput[]
    metadata?: Record<string, string>
    ctime?: string
    mtime?: string
    etime?: string
    editable?: boolean
    deleteable?: boolean
}

function matchType(type?: string | number): string {
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

function levelToView(level?: string | number): string {
    const map: Record<number, BreakLevelType> = {
        1: BreakLevelType.Service,
        2: BreakLevelType.Method,
        3: BreakLevelType.Group,
        4: BreakLevelType.Instance,
    };
    if (typeof level === 'number') return map[level] || BreakLevelType.Service;
    return level || BreakLevelType.Service;
}

function errorInputType(type?: string | number): ErrorConditionType {
    const map: Record<number, ErrorConditionType> = {
        1: ErrorConditionType.RET_CODE,
        2: ErrorConditionType.DELAY,
    };
    if (typeof type === 'number') return map[type] || ErrorConditionType.RET_CODE;
    return (type as ErrorConditionType) || ErrorConditionType.RET_CODE;
}

function triggerType(type?: string | number): TriggerType {
    const map: Record<number, TriggerType> = {
        1: TriggerType.ERROR_RATE,
        2: TriggerType.CONSECUTIVE_ERROR,
    };
    if (typeof type === 'number') return map[type] || TriggerType.ERROR_RATE;
    return (type as TriggerType) || TriggerType.ERROR_RATE;
}

function normalizeErrorCondition(condition: any): ErrorCondition {
    return {
        inputType: errorInputType(condition?.inputType ?? condition?.input_type),
        condition: {
            type: matchType(condition?.condition?.type),
            value: condition?.condition?.value || '',
        },
    };
}

function normalizeTriggerCondition(condition: any): TriggerCondition {
    const type = triggerType(condition?.triggerType ?? condition?.trigger_type);
    const errorPercent = condition?.errorPercent ?? condition?.error_percent ?? 0;
    const errorCount = condition?.errorCount ?? condition?.error_count ?? 0;
    return {
        triggerType: type,
        errorCount,
        errorPercent,
        interval: condition?.interval ?? 0,
        minimumRequest: condition?.minimumRequest ?? condition?.minimum_request ?? 0,
        triggerVal: type === TriggerType.ERROR_RATE ? errorPercent : errorCount,
    };
}

const defaultRecoverCondition = (): RecoverCondition => ({
    sleepWindow: 60,
    consecutiveSuccess: 0,
});

const defaultFaultDetectConfig = (): FaultDetectConfig => ({
    enable: false,
});

const defaultFallbackConfig = (): FallbackConfig => ({
    enable: false,
    response: { code: 500, headers: [], body: '' },
});

function normalizeRecoverCondition(condition: any): RecoverCondition {
    return {
        sleepWindow: condition?.sleepWindow ?? condition?.sleep_window ?? 60,
        consecutiveSuccess: condition?.consecutiveSuccess ?? 0,
    };
}

function normalizeFallbackConfig(config: any): FallbackConfig {
    const fallback = config ?? defaultFallbackConfig();
    return {
        enable: !!fallback.enable,
        response: {
            code: fallback?.response?.code ?? 500,
            headers: fallback?.response?.headers ?? [],
            body: fallback?.response?.body ?? '',
        },
    };
}

function normalizeBlockConfig(block: any, rule: any): BlockConfig {
    const legacyRecoverCondition = rule?.recoverCondition ?? rule?.recover_condition;
    const legacyFaultDetectConfig = rule?.faultDetectConfig ?? rule?.fault_detect_config;
    const legacyFallbackConfig = rule?.fallbackConfig ?? rule?.fallback_config;
    const blockConfig = block?.block_config ?? block?.blockConfig ?? block;
    return {
        name: blockConfig?.name || '',
        api: blockConfig?.api,
        apis: blockConfig?.apis?.length ? blockConfig.apis : (blockConfig?.api ? [blockConfig.api] : []),
        error_conditions: (blockConfig?.error_conditions ?? blockConfig?.errorConditions ?? []).map(normalizeErrorCondition),
        trigger_conditions: (blockConfig?.trigger_conditions ?? blockConfig?.triggerConditions ?? []).map(normalizeTriggerCondition),
        max_ejection_percent: block?.max_ejection_percent ?? block?.maxEjectionPercent ?? rule?.max_ejection_percent ?? rule?.maxEjectionPercent ?? 100,
        recoverCondition: normalizeRecoverCondition(block?.recoverCondition ?? block?.recover_condition ?? legacyRecoverCondition),
        faultDetectConfig: block?.faultDetectConfig ?? block?.fault_detect_config ?? legacyFaultDetectConfig ?? defaultFaultDetectConfig(),
        fallbackConfig: normalizeFallbackConfig(block?.fallbackConfig ?? block?.fallback_config ?? legacyFallbackConfig),
    };
}

function isCircuitBreakerPolicy(block: CircuitBreakerBlockConfigInput): block is CircuitBreakerPolicy {
    return !!(block as CircuitBreakerPolicy)?.block_config;
}

function toCircuitBreakerPolicy(block: CircuitBreakerBlockConfigInput): CircuitBreakerPolicy {
    if (isCircuitBreakerPolicy(block)) {
        return block;
    }
    return {
        block_config: {
            name: block.name,
            apis: block.apis?.length ? block.apis : (block.api ? [block.api] : []),
            error_conditions: block.error_conditions,
            trigger_conditions: block.trigger_conditions,
        },
        max_ejection_percent: block.max_ejection_percent,
        recoverCondition: block.recoverCondition,
        faultDetectConfig: block.faultDetectConfig,
        fallbackConfig: block.fallbackConfig,
    };
}

function toCircuitBreakerPayload(rule: CircuitBreakerRule): Omit<CircuitBreakerRule, 'block_configs'> & { block_configs: CircuitBreakerPolicy[] } {
    return {
        ...rule,
        block_configs: (rule.block_configs || []).map(toCircuitBreakerPolicy),
    };
}

export function normalizeCircuitBreakerRule(rule: CircuitBreakerRule | any): CircuitBreakerRule {
    if (!rule) return rule;
    const matcher = rule.ruleMatcher ?? rule.rule_matcher ?? {};
    const sourceService = rule.srcService ?? rule.src_service ?? matcher?.source?.service ?? '';
    const sourceNamespace = rule.srcNamespace ?? rule.src_namespace ?? matcher?.source?.namespace ?? '*';
    let destinationService = rule.dstService ?? rule.dst_service ?? matcher?.destination?.service ?? '';
    let destinationNamespace = rule.dstNamespace ?? rule.dst_namespace ?? matcher?.destination?.namespace ?? '*';
    const ruleNamespace = rule.namespace ?? sourceNamespace;
    if (destinationService === ruleNamespace && destinationNamespace && destinationNamespace !== ruleNamespace) {
        [destinationNamespace, destinationService] = [destinationService, destinationNamespace];
    }
    return {
        ...rule,
        level: levelToView(rule.level),
        ruleMatcher: {
            source: {
                service: sourceService,
                namespace: sourceNamespace,
            },
            destination: {
                service: destinationService,
                namespace: destinationNamespace,
                method: {
                    type: matchType(matcher?.destination?.method?.type),
                    value: (rule.dstMethod ?? rule.dst_method ?? matcher?.destination?.method?.value) || '',
                },
            },
        },
        block_configs: (rule.block_configs ?? rule.blockConfigs ?? []).map((block: any) => normalizeBlockConfig(block, rule)),
    };
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
    namespace?: string
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
        list: (res.data || []).map(normalizeCircuitBreakerRule),
        totalCount: res.amount,
    }
}

/** 解包后即为规则详情 */
export interface DescribeOneCircuitBreakerResponse {
    data: CircuitBreakerRule
}

export async function describeOneCircuitBreaker(id: string) {
    const res = await getApiRequest<CircuitBreakerRule>({
        action: `${BaseURL.CIRCUIT_BREAKER}/detail`,
        data: { id },
    })
    return normalizeCircuitBreakerRule(res);
}

export type CreateCircuitBreakerRequest = CircuitBreakerRule

export async function createCircuitBreaker(params: CreateCircuitBreakerRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}`,
        data: params.map(toCircuitBreakerPayload),
    })
    return res
}

export type ModifyCircuitBreakerRequest = CircuitBreakerRule

export async function modifyCircuitBreaker(params: ModifyCircuitBreakerRequest[]) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.CIRCUIT_BREAKER}`,
        data: params.map(toCircuitBreakerPayload),
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
