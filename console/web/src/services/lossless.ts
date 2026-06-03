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
    health_check_protocol?: string;
    health_check_method?: string;
    health_check_path?: string;
    health_check_interval?: string;
}

export interface WarmUp {
    enable: boolean;
    interval: string;
    enable_overload_protection: boolean;
    overload_protection_threshold: number;
    curvature: number;
}

export interface GracefulOffline {
    enable: boolean;
    interval: string;
}

export type CreateLossLessRuleRequest = LossLessRule;

export async function createLossLessRule(params: CreateLossLessRuleRequest[]) {
    return await apiRequest<any>({
        action: `${BaseURL.LOSSLESS}`,
        data: params,
    });
}

export type ModifyLossLessRuleRequest = LossLessRule;

export async function modifyLossLessRule(params: ModifyLossLessRuleRequest[]) {
    return await putApiRequest<any>({
        action: `${BaseURL.LOSSLESS}`,
        data: params,
    });
}

export interface DescribeOneLossLessRulesRequest {
    id: string;
}

export interface DescribeOneLossLessRulesResponse {
    data: LossLessRuleView;
}

export async function describeOneLossLessRules(id: string) {
    return await getApiRequest<LossLessRuleView>({
        action: `${BaseURL.LOSSLESS}/detail`,
        data: { id },
    });
}

export interface DescribeLossLessRulesRequest {
    id?: string;
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
        list: res.data,
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
        totalCount: result.total,
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