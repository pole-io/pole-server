import request, { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { RoutingRuleDestination, RoutingSourceArgument } from './router';
import { BaseURL, MatchLogic, MatchString, MatchType, MatchValueType, RuleRelease, RuleReleaseRequest } from './types';
import { LimitArgumentsType } from './ratelimit';

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

function normalizeMatchType(type?: string | number): string {
    const map: Record<number, string> = {
        0: MatchType.EXACT,
        1: MatchType.REGEX,
        2: MatchType.NOT_EQUALS,
        3: MatchType.IN,
        4: MatchType.NOT_IN,
        5: MatchType.RANGE,
    };
    if (typeof type === 'number') return map[type] || MatchType.EXACT;
    return type || MatchType.EXACT;
}

function normalizeMatchValueType(type?: string | number): string {
    const map: Record<number, string> = {
        0: MatchValueType.TEXT,
        1: MatchValueType.PARAMETER,
        2: MatchValueType.VARIABLE,
    };
    if (typeof type === 'number') return map[type] || MatchValueType.TEXT;
    return type || MatchValueType.TEXT;
}

function normalizeMatchString(value?: MatchString | any): MatchString {
    return {
        type: normalizeMatchType(value?.type),
        value: value?.value || '',
        value_type: normalizeMatchValueType(value?.value_type ?? value?.valueType),
    };
}

function normalizeArgumentType(type?: string | number): string {
    const map: Record<number, LimitArgumentsType> = {
        0: LimitArgumentsType.CUSTOM,
        1: LimitArgumentsType.METHOD,
        2: LimitArgumentsType.HEADER,
        3: LimitArgumentsType.QUERY,
        4: LimitArgumentsType.CALLER_SERVICE,
        5: LimitArgumentsType.CALLER_IP,
    };
    if (typeof type === 'number') return map[type] || LimitArgumentsType.CUSTOM;
    return type || LimitArgumentsType.CUSTOM;
}

function decodeAnyServiceSelector(selector: any): ServiceSelector | ServiceGatewaySelector {
    if (!selector) return {};
    if (selector.namespace || selector.service) return selector;
    const base64Value = selector.value;
    if (!base64Value || typeof atob !== 'function') return selector;

    try {
        const binary = atob(base64Value);
        const bytes = Array.from(binary, ch => ch.charCodeAt(0));
        let index = 0;
        const out: ServiceSelector = {};
        while (index < bytes.length) {
            const tag = bytes[index++];
            let length = bytes[index++];
            if (length === undefined) break;
            if (length & 0x80) {
                let shift = 7;
                length = length & 0x7f;
                while (bytes[index] & 0x80) {
                    length |= (bytes[index++] & 0x7f) << shift;
                    shift += 7;
                }
                length |= (bytes[index++] & 0x7f) << shift;
            }
            const raw = bytes.slice(index, index + length);
            index += length;
            const text = String.fromCharCode(...raw);
            if (tag === 0x0a) out.namespace = text;
            if (tag === 0x12) out.service = text;
        }
        return {
            ...selector,
            namespace: out.namespace,
            service: out.service,
        };
    } catch {
        return selector;
    }
}

function normalizeTrafficEntry(entry: TrafficEntry | any): TrafficEntry {
    const selector = decodeAnyServiceSelector(entry?.selector);
    return {
        ...entry,
        type: entry?.type || '',
        selector,
    };
}

function normalizeDestination(destination: RoutingRuleDestination | any): RoutingRuleDestination {
    const labels = Object.entries(destination?.labels || {}).reduce((acc, [key, value]) => {
        acc[key] = normalizeMatchString(value);
        return acc;
    }, {} as Record<string, MatchString>);
    return {
        ...destination,
        service: destination?.service || '',
        namespace: destination?.namespace || '',
        labels,
    };
}

function normalizeLaneRule(rule: LaneRuleView | any): LaneRuleView {
    const trafficMatchRule = rule?.trafficMatchRule ?? rule?.traffic_match_rule ?? {};
    return {
        ...rule,
        groupName: rule?.groupName ?? rule?.group_name ?? '',
        trafficMatchRule: {
            matchMode: trafficMatchRule.matchMode ?? trafficMatchRule.match_mode ?? MatchLogic.AND,
            arguments: (trafficMatchRule.arguments || []).map((arg: any) => ({
                ...arg,
                type: normalizeArgumentType(arg?.type),
                value: normalizeMatchString(arg?.value),
            })),
        },
        defaultLabelValue: rule?.defaultLabelValue ?? rule?.default_label_value ?? '',
        labelKey: rule?.labelKey ?? rule?.label_key ?? 'lane',
        matchMode: rule?.matchMode ?? rule?.match_mode ?? LaneMatchLogic.PERMISSIVE,
    };
}

function normalizeLaneGroup(group: LaneGroupView | any): LaneGroupView {
    return {
        ...group,
        entries: (group?.entries || []).map(normalizeTrafficEntry),
        destinations: (group?.destinations || []).map(normalizeDestination),
        rules: (group?.rules || []).map(normalizeLaneRule),
    };
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
    return {
        ...result,
        data: (result.data || []).map(normalizeLaneGroup),
    };
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
    amount?: number
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
        totalCount: result.amount ?? result.total ?? result.data?.length ?? 0,
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
        data: req.map(laneRuleToApi),
    });
    return result;
}

export type ModifyLaneRulesRequest = LaneRule

export async function modifyLaneRules(req: ModifyLaneRulesRequest[]) {
    const result = await putApiRequest({
        action: `${BaseURL.LANE_GROUP}/rules`,
        data: req.map(laneRuleToApi),
    });
    return result;
}

function laneRuleToApi(rule: LaneRule | any) {
    return {
        id: rule.id,
        name: rule.name,
        group_name: rule.groupName ?? rule.group_name,
        enable: rule.enable,
        priority: rule.priority,
        description: rule.description,
        match_mode: rule.matchMode ?? rule.match_mode,
        traffic_match_rule: rule.trafficMatchRule ?? rule.traffic_match_rule,
        default_label_value: rule.defaultLabelValue ?? rule.default_label_value,
        label_key: rule.labelKey ?? rule.label_key ?? 'lane',
    };
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
