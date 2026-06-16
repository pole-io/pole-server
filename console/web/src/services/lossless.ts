import { apiRequest, getApiRequest, putApiRequest } from "utils/request";
import { BaseURL, RuleRelease } from "./types";
import { DescribeRateLimitVersionsRequest } from "./ratelimit";
import { RollbackCircuitBreakerReleaseRequest } from "./circuitbreaker";
import { PolicySourceType } from "./auth_policy";

export interface LossLessRuleView {
    id?: string;
    service: string;
    namespace: string;
    lossless_online: GracefulOnline;
    lossless_offline: GracefulOffline;
    metadata?: Record<string, string>;
    ctime: string;
    mtime: string;
    editable: boolean;
    deleteable: boolean;
    revision?: string;
}

export interface LossLessRule {
    id?: string;
    service: string;
    namespace: string;
    lossless_online: GracefulOnline;
    lossless_offline: GracefulOffline;
    metadata?: Record<string, string>;
}

export interface GracefulOnline {
    delay_register: DelayRegister;
    warmup: WarmUp;
}

export interface DelayRegister {
    enable: boolean;
    strategy: string;
    interval: string;
    interval_second?: number;
    health_check_protocol?: string;
    health_check_method?: string;
    health_check_path?: string;
    health_check_interval?: string;
    health_check_interval_second?: string | number;
}

export interface WarmUp {
    enable: boolean;
    interval: string;
    interval_second?: number;
    enable_overload_protection: boolean;
    overload_protection_threshold: number;
    curvature: number;
}

export interface GracefulOffline {
    enable: boolean;
    interval: string;
    interval_second?: number;
}

function secondToDuration(value?: string | number): string {
    if (typeof value === 'number') return value === 0 ? '0s' : `${value}s`;
    if (typeof value === 'string' && value !== '') {
        return value.endsWith('s') ? value : `${value}s`;
    }
    return '0s';
}

function durationToSecond(value?: string | number): number {
    if (typeof value === 'number') return value;
    return Number.parseInt((value || '0').replace(/s$/, ''), 10) || 0;
}

function durationToString(value?: string | number): string {
    if (value === undefined || value === null || value === '') return '0s';
    if (typeof value === 'number') return `${value}s`;
    return value.endsWith('s') ? value : `${value}s`;
}

export function normalizeLosslessRule(rule: LossLessRuleView | any): LossLessRuleView {
    if (!rule) return rule;
    const online = rule.lossless_online ?? rule.losslessOnline ?? {};
    const delay = online.delay_register ?? online.delayRegister ?? {};
    const warmup = online.warmup ?? {};
    const offline = rule.lossless_offline ?? rule.losslessOffline ?? {};
    return {
        ...rule,
        lossless_online: {
            delay_register: {
                ...delay,
                interval: secondToDuration(delay.interval ?? delay.interval_second ?? delay.intervalSecond),
                health_check_interval: secondToDuration(delay.health_check_interval ?? delay.health_check_interval_second ?? delay.healthCheckIntervalSecond),
            },
            warmup: {
                ...warmup,
                interval: secondToDuration(warmup.interval ?? warmup.interval_second ?? warmup.intervalSecond),
                enable_overload_protection: warmup.enable_overload_protection ?? warmup.enableOverloadProtection ?? false,
                overload_protection_threshold: warmup.overload_protection_threshold ?? warmup.overloadProtectionThreshold ?? 0,
            },
        },
        lossless_offline: {
            ...offline,
            interval: secondToDuration(offline.interval ?? offline.interval_second ?? offline.intervalSecond),
        },
    };
}

function losslessRuleToApi(rule: LossLessRule | any) {
    const delay = rule.lossless_online?.delay_register || {};
    const warmup = rule.lossless_online?.warmup || {};
    const offline = rule.lossless_offline || {};
    return {
        id: rule.id,
        service: rule.service,
        namespace: rule.namespace,
        metadata: rule.metadata,
        lossless_online: {
            delay_register: {
                enable: delay.enable,
                strategy: delay.strategy,
                interval_second: delay.interval_second ?? durationToSecond(delay.interval),
                health_check_protocol: delay.health_check_protocol,
                health_check_method: delay.health_check_method,
                health_check_path: delay.health_check_path,
                health_check_interval_second: durationToString(delay.health_check_interval_second ?? delay.health_check_interval),
            },
            warmup: {
                enable: warmup.enable,
                interval_second: warmup.interval_second ?? durationToSecond(warmup.interval),
                enable_overload_protection: warmup.enable_overload_protection,
                overload_protection_threshold: warmup.overload_protection_threshold,
                curvature: warmup.curvature,
            },
        },
        lossless_offline: {
            enable: offline.enable,
            interval_second: offline.interval_second ?? durationToSecond(offline.interval),
        },
    };
}

export type CreateLossLessRuleRequest = LossLessRule;

export async function createLossLessRule(params: CreateLossLessRuleRequest[]) {
    return await apiRequest<any>({
        action: `${BaseURL.LOSSLESS}`,
        data: params.map(losslessRuleToApi),
    });
}

export type ModifyLossLessRuleRequest = LossLessRule;

export async function modifyLossLessRule(params: ModifyLossLessRuleRequest[]) {
    return await putApiRequest<any>({
        action: `${BaseURL.LOSSLESS}`,
        data: params.map(losslessRuleToApi),
    });
}

export interface DescribeOneLossLessRulesRequest {
    id: string;
}

export interface DescribeOneLossLessRulesResponse {
    data: LossLessRuleView;
}

export async function describeOneLossLessRules(id: string) {
    const res = await getApiRequest<LossLessRuleView>({
        action: `${BaseURL.LOSSLESS}/detail`,
        data: { id },
    });
    return normalizeLosslessRule(res);
}

export interface DescribeLossLessRulesRequest {
    id?: string;
    name?: string;
    service?: string;
    namespace?: string;
    offset: number;
    limit: number;
}

export interface DescribeLossLessRulesResponse {
    data: LossLessRuleView[];
    amount: number;
}

export async function describeLossLessRules(params: DescribeLossLessRulesRequest) {
    const res = await getApiRequest<DescribeLossLessRulesResponse>({
        action: `${BaseURL.LOSSLESS}`,
        data: params,
    });
    return {
        list: (res.data || []).map(normalizeLosslessRule),
        totalCount: res.amount,
    };
}

export interface DeleteLossLessRuleRequest {
    id: string;
}

export async function deleteLossLessRule(params: DeleteLossLessRuleRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.LOSSLESS}/delete`,
        data: params,
    });
    return res;
}

// publishLossless 发布一个无损规则
export async function publishLossless(params: RuleRelease[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.LOSSLESS}/releases`,
        data: params,
    })
    return res
}

export interface DescribeLosslessVersionsRequest {
    id: string
    offset: number
    limit: number
}

export interface DescribeLosslessVersionsResponse {
    amount?: number
    total: number
    data: RuleRelease[]
}

export async function describeLosslessVersions(req: DescribeLosslessVersionsRequest) {
    const result = await getApiRequest<DescribeLosslessVersionsResponse>({
        action: `${BaseURL.LOSSLESS}/releases`,
        data: req,
    });
    return {
        list: result.data,
        totalCount: result.amount ?? result.total ?? result.data?.length ?? 0,
    };
}

export interface RollbackLosslessReleaseRequest {
    id: string
}

// rollbackLossless 回滚某个规则版本
export async function rollbackLossless(params: RollbackLosslessReleaseRequest) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.LOSSLESS}/releases/rollback`,
        data: [{
            id: params.id,
            "resource": PolicySourceType.LossLessRules,
        }],
    });
    return res;
}

export interface DeleteLosslessReleaseRequest {
    id: string
}

// deleteLosslessRelease 删除某个规则版本
export async function deleteLosslessRelease(params: DeleteLosslessReleaseRequest) {
    const res = await apiRequest<any>({
        action: `${BaseURL.LOSSLESS}/releases/delete`,
        data: [{
            id: params.id,
            "resource": PolicySourceType.LossLessRules,
        }],
    });
    return res;
}
