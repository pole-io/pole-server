import request, { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { BaseURL, MatcheLabel, MatchString, RuleRelease } from './types';

export type RouteType = 'RulePolicy' | 'NearbyPolicy';

// 匹配规则类型
export enum RoutingArgumentsType {
    CUSTOM = 'CUSTOM',
    METHOD = 'METHOD',
    HEADER = 'HEADER',
    QUERY = 'QUERY',
    COOKIE = 'COOKIE',
    PATH = 'PATH',
    CALLER_IP = 'CALLER_IP',
}

export const RoutingArgumentsTypeLabelMap = {
    [RoutingArgumentsType.PATH]: '$path.',
    [RoutingArgumentsType.METHOD]: '$method.',
    [RoutingArgumentsType.HEADER]: '$header.',
    [RoutingArgumentsType.QUERY]: '$query.',
    [RoutingArgumentsType.CALLER_IP]: '$caller_ip.',
    [RoutingArgumentsType.COOKIE]: '$cookie.',
    [RoutingArgumentsType.CUSTOM]: '',
}

export enum RoutingValueType {
    TEXT = 'TEXT',
    PARAMETER = 'PARAMETER',
}

export const RoutingValueTextMap = {
    [RoutingValueType.TEXT]: '值',
    [RoutingValueType.PARAMETER]: '变量',
}

export const RoutingValueTypeOptions = [
    {
        text: '值',
        value: RoutingValueType.TEXT,
    },
    {
        text: '变量',
        value: RoutingValueType.PARAMETER,
    },
]

export const RoutingArgumentsTypeOptions = [
    {
        value: RoutingArgumentsType.CUSTOM,
        label: '自定义',
    },
    {
        value: RoutingArgumentsType.HEADER,
        label: '请求头(HEADER)',
    },
    {
        value: RoutingArgumentsType.COOKIE,
        label: '请求Cookie(COOKIE)',
    },
    {
        value: RoutingArgumentsType.QUERY,
        label: '请求参数(QUERY)',
    },
    {
        value: RoutingArgumentsType.METHOD,
        label: '方法(METHOD)',
    },
    {
        value: RoutingArgumentsType.CALLER_IP,
        label: '主调IP',
    },
    {
        value: RoutingArgumentsType.PATH,
        label: '路径',
    },
]

export const RouteArgumentTextMap = RoutingArgumentsTypeOptions.reduce((map, curr) => {
    map[curr.value] = curr.label
    return map
}, {} as Record<string, string>)

export interface CustomRoute {
    id?: string
    name: string // 规则名
    enable?: boolean // 是否启用
    priority?: number
    description?: string
    routing_config?: RoutingConfig
    routing_policy?: 'RulePolicy' | 'NearbyPolicy'
    metadata?: Record<string, string>
}

export interface CustomRouteView extends CustomRoute {
    ctime?: string
    mtime?: string
    etime?: string
    editable?: boolean
    deleteable?: boolean
}

export interface RoutingConfig {
    '@type': string
    caller?: RoutingSources
    callee?: RoutingSources
    rules: RoutingRule[]
}

export interface RoutingSources {
    service: string
    namespace: string
}

export interface RoutingRule {
    name: string
    sources?: RoutingRuleSource[]
    destinations: RoutingRuleDestination[]
    arguments?: RoutingRuleArguments
}

export interface RoutingRuleArguments {
    arguments?: RoutingSourceArgument[]
    randomPercent?: number
    matchMode?: string
}

export interface RoutingRuleSource {
    service: string
    namespace: string
    arguments: RoutingSourceArgument[]
}

export interface RoutingRuleDestination {
    service: string
    namespace: string
    weight: number
    isolate: boolean
    labels: Record<string, MatchString>
    name: string
    priority?: number
}

export interface RoutingSourceArgument {
    type: string
    key: string
    value: {
        type: string
        value: string
        value_type: string
    }
}

const defaultRoutingConfigType = 'type.googleapis.com/v1.RuleRoutingConfig';

const routeServiceOrDefault = (service?: RoutingSources, fallback?: RoutingSources): RoutingSources => ({
    namespace: service?.namespace || fallback?.namespace || '*',
    service: service?.service || fallback?.service || '',
});

const routeArguments = (rule?: RoutingRule): RoutingSourceArgument[] => (
    rule?.sources?.[0]?.arguments || rule?.arguments?.arguments || []
);

export function normalizeRoutingConfigForEditor(config?: RoutingConfig): RoutingConfig | undefined {
    if (!config) return config;

    const firstRule = config.rules?.[0];
    const caller = routeServiceOrDefault(config.caller, firstRule?.sources?.[0]);
    const callee = routeServiceOrDefault(config.callee, firstRule?.destinations?.[0]);

    return {
        ...config,
        '@type': config['@type'] || defaultRoutingConfigType,
        caller,
        callee,
        rules: (config.rules || []).map((rule) => {
            const source = routeServiceOrDefault(rule.sources?.[0], caller);
            const args = routeArguments(rule);
            return {
                ...rule,
                sources: rule.sources?.length ? rule.sources : [{
                    ...source,
                    arguments: args,
                }],
                arguments: {
                    ...(rule.arguments || {}),
                    arguments: args,
                    randomPercent: rule.arguments?.randomPercent || 0,
                    matchMode: rule.arguments?.matchMode || 'AND',
                },
                destinations: rule.destinations || [],
            };
        }),
    };
}

export function buildRoutingConfigForApi(config: RoutingConfig | undefined, caller: RoutingSources, callee: RoutingSources): RoutingConfig {
    const normalized = normalizeRoutingConfigForEditor(config) || {
        '@type': defaultRoutingConfigType,
        caller,
        callee,
        rules: [],
    };

    return {
        '@type': normalized['@type'] || defaultRoutingConfigType,
        caller,
        callee,
        rules: (normalized.rules || []).map((rule) => {
            const { sources: _sources, ...rest } = rule;
            return {
                ...rest,
                arguments: {
                    ...(rule.arguments || {}),
                    arguments: routeArguments(rule),
                    randomPercent: rule.arguments?.randomPercent || 0,
                    matchMode: rule.arguments?.matchMode || 'AND',
                },
                destinations: rule.destinations || [],
            };
        }),
    };
}

export interface RoutingDestination {
    service: string
    namespace: string
}

export interface RoutingLabel {
    key: string
    value: {
        type: string
        value: string
        value_type: string
    }
}

// describeCustomRoute 查询自定义路由规则
export interface DescribeCustomRoutesRequest {
    id?: string
    name?: string
    namespace?: string
    service?: string
    route_type: RouteType
    order_field?: string
    order_type?: string
    source_service?: string
    source_namespace?: string
    destination_service?: string
    destination_namespace?: string
    offset: number
    limit: number
}

export interface DescribeCustomRoutesResponse {
    data: CustomRouteView[]
    amount: number
}

export async function describeCustomRoutes(params: DescribeCustomRoutesRequest) {
    const res = await getApiRequest<DescribeCustomRoutesResponse>({
        action: `${BaseURL.CUSTOM_ROUTE}`,
        data: params,
    })
    return {
        list: res.data,
        totalCount: res.amount,
    }
}


/** 解包后即为路由详情 */
export interface DescribeOneCustomRouteResponse {
    data: CustomRouteView
}

export async function describeOneCustomRoute(id: string) {
    return await getApiRequest<CustomRouteView>({
        action: `${BaseURL.CUSTOM_ROUTE}/detail`,
        data: { id },
    })
}

// describeCustomRoute 查询已发布的路由规则
export interface DescribeCustomRouteVersionsRequest {
    id?: string
    offset: number
    limit: number
}

export interface DescribeCustomRouteVersionsResponse {
    data: RuleRelease[]
    amount: number
}

export async function describeCustomRouteVersions(params: DescribeCustomRouteVersionsRequest) {
    const res = await getApiRequest<DescribeCustomRouteVersionsResponse>({
        action: `${BaseURL.CUSTOM_ROUTE}/releases`,
        data: params,
    })
    return {
        list: res.data,
        totalCount: res.amount,
    }
}

export type CreateCustomRoutesRequest = CustomRoute

// createCustomRoutes 创建自定义路由规则
export async function createCustomRoutes(params: CreateCustomRoutesRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CUSTOM_ROUTE}`,
        data: params,
    })
    return res
}

export type ModifyCustomRoutesRequest = CustomRoute

// modifyCustomRoutes 修改自定义路由规则
export async function modifyCustomRoutes(params: ModifyCustomRoutesRequest[]) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.CUSTOM_ROUTE}`,
        data: params,
    })
    return res
}

// DeleteCustomRouteRequest 删除自定义路由规则
export interface DeleteCustomRouteRequest {
    id: string
}

export async function deleteCustomRoute(params: DeleteCustomRouteRequest[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CUSTOM_ROUTE}/delete`,
        data: params,
    })
    return res
}

// publishCustomRoute 发布自定义路由规则
export async function publishCustomRoute(params: RuleRelease[]) {
    const res = await apiRequest<any>({
        action: `${BaseURL.CUSTOM_ROUTE}/releases`,
        data: params,
    })
    return res
}

// rollbackCustomRoutes 回滚自定义路由规则
export async function rollbackCustomRoutes(params: RuleRelease[]) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.CUSTOM_ROUTE}/releases/rollback`,
        data: params,
    })
    return res
}

// stopBetaCustomRoutes 停止灰度发布的自定义路由规则
export async function stopBetaCustomRoutes(params: RuleRelease[]) {
    const res = await putApiRequest<any>({
        action: `${BaseURL.CUSTOM_ROUTE}/releases/stopbeta`,
        data: params,
    })
    return res
}
