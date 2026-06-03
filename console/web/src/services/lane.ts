import request, { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { RoutingRuleDestination, RoutingSourceArgument } from './router';
import { BaseURL, MatchLogic, MatchString, RuleRelease, RuleReleaseRequest } from './types';

export const LaneGatewaySelectorType = 'type.googleapis.com/v1.ServiceGatewaySelector';
export const LaneServiceSelectorType = 'type.googleapis.com/v1.ServiceSelector';

export enum LaneMatchLogic {
    STRICT = 'STRICT', // 严格匹配
    PERMISSIVE = 'PERMISSIVE', // 松散匹配
}

export interface TrafficMatchRule {
    arguments: RoutingSourceArgument[] // 匹配参数
    matchMode: MatchLogic // 匹配逻辑
}

export interface TrafficEntry {
    '@type'?: string // 类型标识
    // 入口类型
    type: '' | 'service' | 'gateway'
    selector: ServiceGatewaySelector | ServiceSelector
}

export interface ServiceGatewaySelector {
    namespace?: string // 命名空间
    service?: string // 服务名
    labels?: Record<string, MatchString> // 标签选择器
}

export interface ServiceSelector {
    namespace?: string // 命名空间
    service?: string // 服务名
    labels?: Record<string, MatchString> // 标签选择器
}

export interface LaneGroup {
    id?: string
    name?: string // 规则名
    entries: TrafficEntry[] // 条目列表
    destinations: RoutingRuleDestination[] // 路由目标列表
    rules?: LaneRuleView[] // 规则列表
    metadata?: Record<string, string> // 元数据
    description?: string
}

export interface LaneGroupView extends LaneGroup {
    ctime?: string
    mtime?: string
    editable?: boolean
    deleteable?: boolean
}

export interface DescribeLaneGroupsRequest {
    offset: number
    limit: number
    id?: string // LaneGroup ID
    name?: string // LaneGroup 名称
    brief: boolean
}

export interface DescribeLaneGroupsResponse {
    amount: number
    data: LaneGroupView[]
}

export async function describeLaneGroups(req: DescribeLaneGroupsRequest) {
    const result = await getApiRequest<DescribeLaneGroupsResponse>({
        action: `${BaseURL.LANE_GROUP}`,
        data: req,
    });
    return result;
}

export type CreateLaneGroupsRequest = LaneGroup

export async function createLaneGroups(req: CreateLaneGroupsRequest[]) {
    const result = await apiRequest({
        action: `${BaseURL.LANE_GROUP}`,
        data: req,
    });
    return result;
}

export type ModifyLaneGroupsRequest = LaneGroup

export async function modifyLaneGroups(req: ModifyLaneGroupsRequest[]) {
    const result = await putApiRequest({
        action: `${BaseURL.LANE_GROUP}`,
        data: req,
    });
    return result;
}

export interface DeleteLaneGroupsRequest {
    id: string // LaneGroup ID
}

export async function deleteLaneGroups(req: DeleteLaneGroupsRequest[]) {
    const result = await apiRequest({
        action: `${BaseURL.LANE_GROUP}/delete`,
        data: req,
    });
    return result;
}

export interface DescribeLaneGroupVersionsRequest {
    offset: number
    limit: number
    id: string
}

export interface DescribeLaneGroupVersionsResponse {
    total: number
    data: RuleRelease[]
}

export async function describeLaneGroupVersions(req: DescribeLaneGroupVersionsRequest) {
    const result = await getApiRequest<DescribeLaneGroupVersionsResponse>({
        action: `${BaseURL.LANE_GROUP}/releases`,
        data: req,
    });
    return {
        list: result.data,
        totalCount: result.total,
    };
}

export async function publishLaneGroups(req: RuleRelease[]) {
    const result = await apiRequest({
        action: `${BaseURL.LANE_GROUP}/releases`,
        data: req,
    });
    return result;
}

export async function rollbackLaneGroup(req: RuleReleaseRequest[]) {
    const result = await putApiRequest({
        action: `${BaseURL.LANE_GROUP}/releases/rollback`,
        data: req,
    });
    return result;
}

export async function deleteLaneGroupReleases(req: RuleReleaseRequest[]) {
    const result = await apiRequest({
        action: `${BaseURL.LANE_GROUP}/releases/delete`,
        data: req,
    });
    return result;
}

export async function stopBetaLaneGroup(req: RuleReleaseRequest[]) {
    const result = await putApiRequest({
        action: `${BaseURL.LANE_GROUP}/releases/stopbeta`,
        data: req,
    });
    return result;
}

export interface LaneRule {
    id?: string
    name?: string // 规则名
    groupName: string // 所属 LaneGroup 名称
    enable?: boolean // 是否启用
    priority?: number
    description?: string
    matchMode: LaneMatchLogic // 匹配逻辑
    trafficMatchRule: TrafficMatchRule
    defaultLabelValue: string // 保存这个泳道的默认实例标签
    labelKey: string // 属于该泳道的实例标签KEY，不填默认为lane
}

export interface LaneRuleView extends LaneRule {
    ctime?: string
    mtime?: string
    etime?: string
    editable?: boolean
    deleteable?: boolean
}

export type CreateLaneRulesRequest = LaneRule

export async function createLaneRules(req: CreateLaneRulesRequest[]) {
    const result = await apiRequest({
        action: `${BaseURL.LANE_GROUP}/rules`,
        data: req,
    });
    return result;
}

export type ModifyLaneRulesRequest = LaneRule

export async function modifyLaneRules(req: ModifyLaneRulesRequest[]) {
    const result = await putApiRequest({
        action: `${BaseURL.LANE_GROUP}/rules`,
        data: req,
    });
    return result;
}

export interface DeleteLaneRulesRequest {
    id: string // 泳道ID
    groupName: string
}

export async function deleteLaneRules(req: DeleteLaneRulesRequest[]) {
    const result = await apiRequest({
        action: `${BaseURL.LANE_GROUP}/rules/delete`,
        data: req,
    });
    return result;
}
