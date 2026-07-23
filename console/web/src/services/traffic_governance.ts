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

export enum TrafficSecurityAuthMode {
    LEGACY_REQUEST_MATCH = 'LEGACY_REQUEST_MATCH',
    MANAGED_IDENTITY = 'MANAGED_IDENTITY',
    CUSTOM_HEADER = 'CUSTOM_HEADER',
}

export interface TrafficSecurityAuthentication {
    mode: TrafficSecurityAuthMode | string
    managed_identity?: Record<string, never>
    custom_header?: {
        header_name?: string
        /** Write-only management input; never returned after persistence. */
        value?: string
        /** Server-generated verifier digest; not accepted from Console writes. */
        value_sha256?: string
    }
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

export interface TrafficTargetService {
    namespace?: string
    service?: string
}

export interface TrafficSourceService {
    namespace?: string
    service?: string
}

export interface TrafficRuleBase {
    id?: string
    name: string
    namespace?: string
    service?: string
    target_service?: TrafficTargetService
    caller?: TrafficSourceService
    callee?: TrafficTargetService
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
    authentication?: TrafficSecurityAuthentication
    policies: TrafficSecurityPolicy[]
}

export interface ManagedCallerSelector {
    any_authenticated?: boolean
    callers?: TrafficSourceService[]
}

export interface TrafficSecurityPolicy {
    api?: TrafficApiScope
    apis?: TrafficApiScope[]
    traffic_match_rule?: TrafficMatchRule
    managed_caller?: ManagedCallerSelector
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
    api?: TrafficApiScope
    apis?: TrafficApiScope[]
    interfaces?: TrafficApiScope[]
    traffic_match_rule?: TrafficMatchRule
    destination?: {
        namespace?: string
        service?: string
        labels?: Record<string, MatchString>
    }
    mirror_percent?: number
    disable?: boolean
}

export interface TrafficMock extends TrafficRuleBase {
    rules: MockRule[]
}

export interface MockRule {
    api?: TrafficApiScope
    apis?: TrafficApiScope[]
    traffic_match_rule?: TrafficMatchRule
    response?: {
        headers?: Record<string, string>
        body?: string
        code?: string
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
    target_service: { namespace: '', service: '' },
    description: '',
    priority: 0,
    enable: true,
    authentication: {
        mode: TrafficSecurityAuthMode.MANAGED_IDENTITY,
        managed_identity: {},
    },
    policies: [{
        action: TrafficSecurityAction.ALLOW,
        apis: [{
            protocol: InterfaceProtocol.HTTP,
            method: 'GET',
            path: {
                type: MatchType.EXACT,
                value: '/',
                value_type: MatchValueType.TEXT,
            },
        }],
        managed_caller: {
            any_authenticated: true,
            callers: [],
        },
    }],
});

export const defaultTrafficMirrorRule = (): TrafficMirror => ({
    name: '',
    caller: { namespace: '*', service: '*' },
    callee: { namespace: '', service: '' },
    description: '',
    priority: 0,
    enable: true,
    rules: [{
        apis: [{
            protocol: InterfaceProtocol.HTTP,
            method: 'GET',
            path: {
                type: MatchType.EXACT,
                value: '/',
                value_type: MatchValueType.TEXT,
            },
        }],
        traffic_match_rule: defaultTrafficMatchRule(),
        destination: {
            namespace: '',
            service: '',
            labels: {},
        },
        mirror_percent: 10,
        disable: false,
    }],
});

export const defaultTrafficMockRule = (): TrafficMock => ({
    name: '',
    target_service: { namespace: '', service: '' },
    description: '',
    priority: 0,
    enable: true,
    rules: [{
        apis: [{
            protocol: InterfaceProtocol.HTTP,
            method: 'GET',
            path: {
                type: MatchType.EXACT,
                value: '/',
                value_type: MatchValueType.TEXT,
            },
        }],
        traffic_match_rule: defaultTrafficMatchRule(),
        response: {
            code: '200',
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

const normalizeTrafficGovernanceRules = (kind: TrafficGovernanceKind, params: TrafficGovernanceRule[]) => {
    if (kind === 'security') {
        return params.map((item) => ({
            ...item,
            policies: ((item as TrafficSecurityRule).policies || []).map((policy) => {
                const { api: _api, ...submitPolicy } = policy;
                return submitPolicy;
            }),
        })) as TrafficGovernanceRule[];
    }
    return params.map((item) => {
        const next = {
            ...item,
            rules: ((item as TrafficMirror | TrafficMock).rules || []).map((rule) => {
                if (kind === 'mirror') {
                    const mirrorRule = rule as MirrorRule;
                    const { api: _api, interfaces: _interfaces, ...submitRule } = mirrorRule;
                    return {
                        ...submitRule,
                    };
                }
                const mockRule = rule as MockRule;
                const response = mockRule.response || {};
                const { api: _api, delay: _delay, ...submitRule } = mockRule;
                return {
                    ...submitRule,
                    mock_percent: mockRule.mock_percent ?? 100,
                    response: {
                        code: response.code || String((response as Record<string, unknown>).status_code || '200'),
                        headers: response.headers || {},
                        body: response.body || '',
                    },
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
