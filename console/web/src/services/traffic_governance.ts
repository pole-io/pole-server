import { apiRequest, getApiRequest, putApiRequest } from 'utils/request';
import { BaseURL, InterfaceProtocol, MatchLogic, MatchString, MatchType, MatchValueType, RuleRelease } from './types';

export type TrafficGovernanceKind = 'security' | 'mirror' | 'mock';

export const TrafficGovernanceKindLabel: Record<TrafficGovernanceKind, string> = {
    security: '鉴权',
    mirror: '镜像',
    mock: 'Mock',
};

export const TrafficGovernanceReleaseResource: Record<TrafficGovernanceKind, string> = {
    security: 'TrafficSecurityRules',
    mirror: 'TrafficMirrorRules',
    mock: 'TrafficMockRules',
};

export const TrafficGovernanceAuthResource: Record<TrafficGovernanceKind, string> = {
    security: 'SecurityRules',
    mirror: 'MirrorRules',
    mock: 'MockRules',
};

export enum TrafficSecurityAction {
    ALLOW = 'TRAFFIC_SECURITY_ALLOW',
    DENY = 'TRAFFIC_SECURITY_DENY',
}

export const TrafficSecurityActionMap: Record<string, string> = {
    [TrafficSecurityAction.ALLOW]: '放通',
    [TrafficSecurityAction.DENY]: '拒绝',
    TRAFFIC_SECURITY_ALLOW: '放通',
    TRAFFIC_SECURITY_DENY: '拒绝',
    ALLOW: '放通',
    DENY: '拒绝',
    '0': '放通',
    '1': '拒绝',
};

export interface TrafficMatchRule {
    arguments?: TrafficSourceMatch[]
    randomPercent?: number
    matchMode?: MatchLogic | string
}

export interface TrafficSourceMatch {
    type: string
    key: string
    value: MatchString
}

export interface TrafficApiScope {
    protocol?: InterfaceProtocol | string
    method?: string
    path?: MatchString
}

export interface TrafficRuleBase {
    id?: string
    name: string
    namespace: string
    service: string
    description?: string
    priority?: number
    enable?: boolean
    ctime?: string
    mtime?: string
    metadata?: Record<string, string>
    revision?: string
    editable?: boolean
    deleteable?: boolean
}

export interface TrafficSecurityRule extends TrafficRuleBase {
    policies: TrafficSecurityPolicy[]
    default_action: TrafficSecurityAction | string
}

export interface TrafficSecurityPolicy {
    api?: TrafficApiScope
    traffic_match_rule?: TrafficMatchRule
    action: TrafficSecurityAction | string
    reject_effect?: {
        status_code?: number
        code?: string
        message?: string
    }
}

export interface TrafficMirror extends TrafficRuleBase {
    rules: MirrorRule[]
}

export interface MirrorRule {
    source?: {
        namespace?: string
        service?: string
        api?: TrafficApiScope
        traffic_match_rule?: TrafficMatchRule
    }
    destination?: {
        namespace?: string
        service?: string
        labels?: Record<string, MatchString>
    }
    mirror_percent?: number
    duration?: string | { seconds?: number | string; nanos?: number }
    disable?: boolean
}

export interface TrafficMock extends TrafficRuleBase {
    rules: MockRule[]
}

export interface MockRule {
    source?: {
        namespace?: string
        service?: string
        api?: TrafficApiScope
        traffic_match_rule?: TrafficMatchRule
    }
    response?: {
        status_code?: number
        headers?: Record<string, string>
        body?: string
        code?: string
        message?: string
    }
    mock_percent?: number
    delay?: string | { seconds?: number | string; nanos?: number }
    disable?: boolean
}

export type TrafficGovernanceRule = TrafficSecurityRule | TrafficMirror | TrafficMock;

export interface DescribeTrafficGovernanceRequest {
    id?: string
    name?: string
    namespace?: string
    service?: string
    offset: number
    limit: number
}

export interface DescribeTrafficGovernanceResponse<T extends TrafficGovernanceRule> {
    data?: T[]
    amount?: number
    size?: number
    trafficSecurityRules?: T[]
    trafficMirrorRules?: T[]
    trafficMockRules?: T[]
}

export interface DescribeTrafficGovernanceVersionsRequest {
    id?: string
    rule_name?: string
    offset: number
    limit: number
}

export interface DescribeTrafficGovernanceVersionsResponse {
    data?: RuleRelease[]
    amount?: number
    size?: number
}

const omitEmptyQuery = <T extends Record<string, unknown>>(params: T): T => (
    Object.fromEntries(Object.entries(params).filter(([, value]) => value !== '')) as T
);

const baseUrl = (kind: TrafficGovernanceKind) => {
    switch (kind) {
        case 'security':
            return BaseURL.TRAFFIC_SECURITY;
        case 'mirror':
            return BaseURL.TRAFFIC_MIRROR;
        case 'mock':
            return BaseURL.TRAFFIC_MOCK;
    }
};

const dataKey = (kind: TrafficGovernanceKind) => {
    switch (kind) {
        case 'security':
            return 'trafficSecurityRules';
        case 'mirror':
            return 'trafficMirrorRules';
        case 'mock':
            return 'trafficMockRules';
    }
};

export const defaultTrafficMatchRule = (): TrafficMatchRule => ({
    matchMode: MatchLogic.AND,
    randomPercent: 0,
    arguments: [{
        type: 'HEADER',
        key: '',
        value: {
            type: MatchType.EXACT,
            value: '',
            value_type: MatchValueType.TEXT,
        },
    }],
});

export const defaultTrafficSecurityRule = (): TrafficSecurityRule => ({
    name: '',
    namespace: '',
    service: '',
    description: '',
    priority: 0,
    enable: true,
    default_action: TrafficSecurityAction.DENY,
    policies: [{
        action: TrafficSecurityAction.ALLOW,
        api: {
            protocol: InterfaceProtocol.HTTP,
            method: 'GET',
            path: {
                type: MatchType.EXACT,
                value: '/',
                value_type: MatchValueType.TEXT,
            },
        },
        traffic_match_rule: defaultTrafficMatchRule(),
    }],
});

export const defaultTrafficMirrorRule = (): TrafficMirror => ({
    name: '',
    namespace: '',
    service: '',
    description: '',
    priority: 0,
    enable: true,
    rules: [{
        source: {
            namespace: '',
            service: '',
            api: {
                protocol: InterfaceProtocol.HTTP,
                method: 'GET',
                path: {
                    type: MatchType.EXACT,
                    value: '/',
                    value_type: MatchValueType.TEXT,
                },
            },
            traffic_match_rule: defaultTrafficMatchRule(),
        },
        destination: {
            namespace: '',
            service: '',
            labels: {},
        },
        mirror_percent: 10,
        duration: '0s',
        disable: false,
    }],
});

export const defaultTrafficMockRule = (): TrafficMock => ({
    name: '',
    namespace: '',
    service: '',
    description: '',
    priority: 0,
    enable: true,
    rules: [{
        source: {
            namespace: '',
            service: '',
            api: {
                protocol: InterfaceProtocol.HTTP,
                method: 'GET',
                path: {
                    type: MatchType.EXACT,
                    value: '/',
                    value_type: MatchValueType.TEXT,
                },
            },
            traffic_match_rule: defaultTrafficMatchRule(),
        },
        response: {
            status_code: 200,
            headers: {},
            body: '{}',
        },
        mock_percent: 100,
        delay: '0s',
        disable: false,
    }],
});

export const defaultTrafficGovernanceRule = (kind: TrafficGovernanceKind): TrafficGovernanceRule => {
    if (kind === 'security') return defaultTrafficSecurityRule();
    if (kind === 'mirror') return defaultTrafficMirrorRule();
    return defaultTrafficMockRule();
};

const durationToProtoJson = (value?: string | { seconds?: number | string; nanos?: number }) => {
    if (value === undefined || value === null || value === '') return undefined;
    if (typeof value === 'string') return value;
    const seconds = Number(value.seconds ?? 0);
    const nanos = Number(value.nanos ?? 0);
    if (!Number.isFinite(seconds) || !Number.isFinite(nanos)) return '0s';
    if (nanos === 0) return `${seconds}s`;
    const fraction = String(Math.abs(nanos)).padStart(9, '0').replace(/0+$/, '');
    return `${seconds}.${fraction}s`;
};

const normalizeTrafficGovernanceRules = (kind: TrafficGovernanceKind, params: TrafficGovernanceRule[]) => {
    if (kind === 'security') return params;
    return params.map((item) => {
        const next = {
            ...item,
            rules: ((item as TrafficMirror | TrafficMock).rules || []).map((rule) => {
                if (kind === 'mirror') {
                    const mirrorRule = rule as MirrorRule;
                    return {
                        ...mirrorRule,
                        duration: durationToProtoJson(mirrorRule.duration),
                    };
                }
                const mockRule = rule as MockRule;
                return {
                    ...mockRule,
                    delay: durationToProtoJson(mockRule.delay),
                };
            }),
        };
        return next;
    }) as TrafficGovernanceRule[];
};

export async function describeTrafficGovernanceRules<T extends TrafficGovernanceRule>(kind: TrafficGovernanceKind, params: DescribeTrafficGovernanceRequest) {
    const res = await getApiRequest<DescribeTrafficGovernanceResponse<T>>({
        action: baseUrl(kind),
        data: omitEmptyQuery(params),
    });
    const keyedList = res[dataKey(kind) as keyof DescribeTrafficGovernanceResponse<T>] as T[] | undefined;
    const list = res.data ?? keyedList ?? [];
    return {
        list,
        totalCount: res.amount ?? list.length,
    };
}

export async function describeOneTrafficGovernanceRule<T extends TrafficGovernanceRule>(kind: TrafficGovernanceKind, id: string) {
    return await getApiRequest<T>({
        action: `${baseUrl(kind)}/detail`,
        data: { id },
    });
}

export async function createTrafficGovernanceRules(kind: TrafficGovernanceKind, params: TrafficGovernanceRule[]) {
    return await apiRequest({
        action: baseUrl(kind),
        data: normalizeTrafficGovernanceRules(kind, params),
    });
}

export async function modifyTrafficGovernanceRules(kind: TrafficGovernanceKind, params: TrafficGovernanceRule[]) {
    return await putApiRequest({
        action: baseUrl(kind),
        data: normalizeTrafficGovernanceRules(kind, params),
    });
}

export async function deleteTrafficGovernanceRules(kind: TrafficGovernanceKind, params: Pick<TrafficGovernanceRule, 'id'>[]) {
    return await apiRequest({
        action: `${baseUrl(kind)}/delete`,
        data: params,
    });
}

export async function publishTrafficGovernanceRule(kind: TrafficGovernanceKind, params: RuleRelease[]) {
    return await apiRequest({
        action: `${baseUrl(kind)}/releases`,
        data: params.map((item) => ({ ...item, resource: TrafficGovernanceReleaseResource[kind] })),
    });
}

export async function describeTrafficGovernanceVersions(kind: TrafficGovernanceKind, params: DescribeTrafficGovernanceVersionsRequest) {
    const res = await getApiRequest<DescribeTrafficGovernanceVersionsResponse>({
        action: `${baseUrl(kind)}/releases`,
        data: params,
    });
    const list = res.data ?? [];
    return {
        list,
        totalCount: res.amount ?? list.length,
    };
}

export async function deleteTrafficGovernanceRelease(kind: TrafficGovernanceKind, id: string) {
    return await apiRequest({
        action: `${baseUrl(kind)}/releases/delete`,
        data: [{ id }],
    });
}

export async function stopBetaTrafficGovernanceRelease(kind: TrafficGovernanceKind, id: string) {
    return await putApiRequest({
        action: `${baseUrl(kind)}/releases/stopbeta`,
        data: [{ id }],
    });
}

export function inferTrafficGovernanceKindByResource(resource: string): TrafficGovernanceKind | undefined {
    if (resource === TrafficGovernanceReleaseResource.security || resource === TrafficGovernanceAuthResource.security) return 'security';
    if (resource === TrafficGovernanceReleaseResource.mirror || resource === TrafficGovernanceAuthResource.mirror) return 'mirror';
    if (resource === TrafficGovernanceReleaseResource.mock || resource === TrafficGovernanceAuthResource.mock) return 'mock';
    return undefined;
}
