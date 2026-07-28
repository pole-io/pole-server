import {
    TrafficApiScope,
    TrafficGovernanceRule,
    TrafficMatchRule,
    TrafficSecurityAction,
    TrafficSecurityAuthMode,
    ManagedCallerSelector,
    TrafficSecurityPolicy,
    TrafficSecurityRule,
} from 'services/traffic_governance';
import { InterfaceProtocol, MatchLogic, MatchString, MatchType, MatchValueType } from 'services/types';

export type TrafficSecuritySpecFormat = 'yaml' | 'json';
export type SecurityListType = 'DENY_LIST' | 'ALLOW_LIST';
export type SecuritySubRuleKind = 'deny' | 'allow' | 'service';

export interface SecurityViewRule {
    kind: SecuritySubRuleKind;
    listType: SecurityListType;
    interfaces: TrafficApiScope[];
    strategy: TrafficMatchRule;
    managedCaller: ManagedCallerSelector;
    rejectEffect?: TrafficSecurityPolicy['reject_effect'];
}

export interface SecurityValidationError {
    field: string;
    message: string;
}

export interface SecurityPreviewSpec {
    apiVersion: 'governance.pole.io/v1';
    kind: 'AuthRule';
    metadata: {
        name: string;
        enabled: boolean;
        priority: number;
        labels: string[];
    };
    spec: {
        scope: {
            namespace: string;
            service: string;
        };
        description: string;
        authentication: {
            mode: TrafficSecurityAuthMode;
            managedIdentity?: {
                managedBy: 'control-plane';
            };
            customHeader?: {
                headerName: string;
                value: '<redacted>' | '<required>';
            };
        };
        subRules: Array<{
            name: string;
            listType: SecurityListType;
            protectedInterfaces: Array<{
                protocol: string;
                method: string;
                path: string;
                op: string;
            }>;
            strategy: {
                managedCaller?: {
                    anyAuthenticated: boolean;
                    callers: Array<{ namespace: string; service: string }>;
                };
                customHeader?: {
                    headerName: string;
                };
                match?: {
                    relation: string;
                    conditions: Array<{
                        param: string;
                        key: string;
                        op: string;
                        value: string;
                    }>;
                };
            };
        }>;
    };
}

export const listTypeText: Record<SecurityListType, string> = {
    DENY_LIST: '黑名单',
    ALLOW_LIST: '白名单',
};

export const matchParamDefaultKey: Record<string, string> = {
    HEADER: 'authorization',
    QUERY: 'token',
    PATH: 'path',
    COOKIE: 'session',
    METHOD: 'method',
};

export const securityMatchSourceOptions = ['HEADER', 'QUERY', 'PATH', 'COOKIE', 'METHOD'].map((item) => ({ label: item, value: item }));

export const defaultMatchValue = (): MatchString => ({
    type: MatchType.EXACT,
    value: '',
    value_type: MatchValueType.TEXT,
});

export const defaultProtectedInterface = (): TrafficApiScope => ({
    protocol: InterfaceProtocol.HTTP,
    method: 'GET',
    path: {
        type: MatchType.EXACT,
        value: '/',
    } as MatchString,
});

export interface ApiProtocolPresentation {
    methodLabel: string;
    methodPlaceholder: string;
    pathLabel: string;
    pathPlaceholder: string;
    usesHttpMethodSelect: boolean;
}

// The protobuf stores every API scope as protocol/method/path, but those fields
// carry protocol-specific meanings. Keep that distinction in one place so all
// governance editors render and reset the same way.
export function getApiProtocolPresentation(protocol?: string): ApiProtocolPresentation {
    switch (String(protocol || InterfaceProtocol.HTTP).toUpperCase()) {
        case 'GRPC':
            return {
                methodLabel: '可选方法',
                methodPlaceholder: '可选，例如 SayHello',
                pathLabel: '服务名',
                pathPlaceholder: 'helloworld.Greeter',
                usesHttpMethodSelect: false,
            };
        case 'DUBBO':
            return {
                methodLabel: '可选方法',
                methodPlaceholder: '可选，例如 getUser',
                pathLabel: '接口名',
                pathPlaceholder: 'com.example.UserService',
                usesHttpMethodSelect: false,
            };
        default:
            return {
                methodLabel: 'HTTP 方法',
                methodPlaceholder: 'GET',
                pathLabel: '接口路径',
                pathPlaceholder: '/orders',
                usesHttpMethodSelect: true,
            };
    }
}

export function resetApiScopeForProtocol(api: TrafficApiScope, protocol: string): TrafficApiScope {
    const presentation = getApiProtocolPresentation(protocol);
    return {
        ...api,
        protocol,
        method: presentation.usesHttpMethodSelect ? 'GET' : '',
        path: {
            type: api.path?.type || MatchType.EXACT,
            value: '',
        } as MatchString,
    };
}

// `API.path` reuses MatchString in the protobuf contract, but value_type only
// has business meaning for request-parameter matching. Keep API paths limited
// to their actual resource shape and let protobuf use its TEXT default.
function normalizeProtectedInterface(api?: TrafficApiScope): TrafficApiScope {
    const protocol = api?.protocol || InterfaceProtocol.HTTP;
    const presentation = getApiProtocolPresentation(protocol);
    return {
        protocol,
        method: api?.method ?? (presentation.usesHttpMethodSelect ? 'GET' : ''),
        path: {
            type: api?.path?.type || MatchType.EXACT,
            value: api?.path?.value ?? (presentation.usesHttpMethodSelect ? '/' : ''),
        } as MatchString,
    };
}

export const defaultSecurityArgument = () => ({
    type: 'HEADER',
    key: 'authorization',
    value: defaultMatchValue(),
});

export const defaultSecurityMatchRule = (): TrafficMatchRule => ({
    matchMode: MatchLogic.AND,
    randomPercent: 0,
    arguments: [defaultSecurityArgument()],
});

export const defaultManagedCaller = (): ManagedCallerSelector => ({
    any_authenticated: true,
    callers: [],
});

export function readSecurityAuthMode(rule?: TrafficSecurityRule): TrafficSecurityAuthMode {
    const mode = rule?.authentication?.mode;
    if (mode === TrafficSecurityAuthMode.MANAGED_IDENTITY || mode === 'MANAGED_IDENTITY' || mode === '1') {
        return TrafficSecurityAuthMode.MANAGED_IDENTITY;
    }
    if (mode === TrafficSecurityAuthMode.CUSTOM_HEADER || mode === 'CUSTOM_HEADER' || mode === '2') {
        return TrafficSecurityAuthMode.CUSTOM_HEADER;
    }
    return TrafficSecurityAuthMode.LEGACY_REQUEST_MATCH;
}

export function buildSecurityAuthenticationForMode(
    mode: TrafficSecurityAuthMode,
): NonNullable<TrafficSecurityRule['authentication']> {
    if (mode === TrafficSecurityAuthMode.MANAGED_IDENTITY) {
        return { mode, managed_identity: {} };
    }
    return { mode: TrafficSecurityAuthMode.LEGACY_REQUEST_MATCH };
}

export function normalizeManagedCaller(selector?: ManagedCallerSelector): ManagedCallerSelector {
    const callers = selector?.callers || [];
    return {
        any_authenticated: selector?.any_authenticated !== false && callers.length === 0,
        callers,
    };
}

export function normalizeSecurityMatchRule(match?: TrafficMatchRule): TrafficMatchRule {
    return {
        matchMode: match?.matchMode || MatchLogic.AND,
        randomPercent: match?.randomPercent ?? 0,
        arguments: match?.arguments?.length ? match.arguments : [defaultSecurityArgument()],
    };
}

export function readSecurityAction(value?: string | number): TrafficSecurityAction {
    if (value === 1 || value === '1' || value === 'TRAFFIC_SECURITY_DENY' || value === 'DENY') {
        return TrafficSecurityAction.DENY;
    }
    return TrafficSecurityAction.ALLOW;
}

export function actionToListType(action?: string | number): SecurityListType {
    return readSecurityAction(action) === TrafficSecurityAction.DENY ? 'DENY_LIST' : 'ALLOW_LIST';
}

export function listTypeToAction(listType: SecurityListType): TrafficSecurityAction {
    return listType === 'DENY_LIST' ? TrafficSecurityAction.DENY : TrafficSecurityAction.ALLOW;
}

export function isServiceLevelPolicy(policy: TrafficSecurityPolicy): boolean {
    const apis = policy.apis?.length ? policy.apis : (policy.api ? [policy.api] : []);
    if (!apis.length) return true;
    return apis.every((api) => {
        const path = api.path?.value;
        return path === undefined || path === null || path === '';
    });
}

export function normalizeSecurityViewRules(rule: TrafficSecurityRule): SecurityViewRule[] {
    const policies = rule.policies || [];
    const interfaceRules = policies
        .filter((policy) => !isServiceLevelPolicy(policy))
        .map<SecurityViewRule>((policy) => ({
            kind: actionToListType(policy.action) === 'DENY_LIST' ? 'deny' : 'allow',
            listType: actionToListType(policy.action),
            interfaces: (policy.apis?.length ? policy.apis : [policy.api || defaultProtectedInterface()])
                .map(normalizeProtectedInterface),
            strategy: normalizeSecurityMatchRule(policy.traffic_match_rule),
            managedCaller: normalizeManagedCaller(policy.managed_caller),
            rejectEffect: policy.reject_effect,
        }));
    const serviceRule = policies.find(isServiceLevelPolicy);
    const serviceRules = serviceRule ? [{
        kind: 'service' as const,
        listType: actionToListType(serviceRule.action),
        interfaces: [],
        strategy: normalizeSecurityMatchRule(serviceRule.traffic_match_rule),
        managedCaller: normalizeManagedCaller(serviceRule.managed_caller),
        rejectEffect: serviceRule.reject_effect,
    }] : [];
    return normalizeSecurityViewOrder([...interfaceRules, ...serviceRules]);
}

export function normalizeSecurityViewOrder(rules: SecurityViewRule[]): SecurityViewRule[] {
    const deny = rules.filter((item) => item.kind === 'deny').map((item) => ({ ...item, listType: 'DENY_LIST' as const }));
    const allow = rules.filter((item) => item.kind === 'allow').map((item) => ({ ...item, listType: 'ALLOW_LIST' as const }));
    const service = rules.find((item) => item.kind === 'service');
    return [
        ...deny,
        ...allow,
        ...(service ? [{ ...service, interfaces: [], kind: 'service' as const }] : []),
    ];
}

export function buildSecurityPoliciesFromView(rules: SecurityViewRule[], authMode: TrafficSecurityAuthMode): TrafficSecurityPolicy[] {
    return normalizeSecurityViewOrder(rules).map((viewRule) => {
        const action = listTypeToAction(viewRule.listType);
        const base: TrafficSecurityPolicy = {
            action,
            reject_effect: viewRule.rejectEffect || (action === TrafficSecurityAction.DENY ? {
                status_code: 403,
                code: 'FORBIDDEN',
                message: 'request denied by auth rule',
            } : undefined),
        };
        if (authMode === TrafficSecurityAuthMode.MANAGED_IDENTITY) {
            base.managed_caller = normalizeManagedCaller(viewRule.managedCaller);
        } else {
            base.traffic_match_rule = normalizeSecurityMatchRule(viewRule.strategy);
        }
        if (viewRule.kind === 'service') {
            return { ...base };
        }
        return {
            ...base,
            apis: (viewRule.interfaces.length ? viewRule.interfaces : [defaultProtectedInterface()])
                .map(normalizeProtectedInterface),
        };
    });
}

function metadataToLabels(metadata?: Record<string, string>): string[] {
    return Object.entries(metadata || {}).map(([key, value]) => `${key}:${value}`);
}

function protectedInterfaceToPreview(api: TrafficApiScope) {
    return {
        protocol: String(api.protocol || 'HTTP'),
        method: String(api.method || 'GET'),
        path: api.path?.value || '',
        op: api.path?.type || MatchType.EXACT,
    };
}

export function buildSecurityPreviewSpec(rule: TrafficGovernanceRule, viewRules: SecurityViewRule[]): SecurityPreviewSpec {
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    const securityRule = rule as TrafficSecurityRule;
    const storedAuthMode = readSecurityAuthMode(securityRule);
    const authMode = storedAuthMode === TrafficSecurityAuthMode.CUSTOM_HEADER
        ? TrafficSecurityAuthMode.LEGACY_REQUEST_MATCH
        : storedAuthMode;
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'AuthRule',
        metadata: {
            name: rule.name || rule.id || '',
            enabled: rule.enable !== false,
            priority: Number(rule.priority || 0),
            labels: metadataToLabels(rule.metadata),
        },
        spec: {
            scope: {
                namespace: target.namespace || '',
                service: target.service || '',
            },
            description: rule.description || '',
            authentication: {
                mode: authMode,
                ...(authMode === TrafficSecurityAuthMode.MANAGED_IDENTITY
                    ? { managedIdentity: { managedBy: 'control-plane' as const } }
                    : {}),
            },
            subRules: normalizeSecurityViewOrder(viewRules).map((item, index) => {
                const match = normalizeSecurityMatchRule(item.strategy);
                const managedCaller = normalizeManagedCaller(item.managedCaller);
                return {
                    name: `auth-subrule-${index + 1}`,
                    listType: item.listType,
                    protectedInterfaces: item.kind === 'service' ? [] : item.interfaces.map(protectedInterfaceToPreview),
                    strategy: authMode === TrafficSecurityAuthMode.MANAGED_IDENTITY ? {
                        managedCaller: {
                            anyAuthenticated: Boolean(managedCaller.any_authenticated),
                            callers: (managedCaller.callers || []).map((caller) => ({
                                namespace: caller.namespace || '',
                                service: caller.service || '',
                            })),
                        },
                    } : {
                        match: {
                            relation: String(match.matchMode || MatchLogic.AND),
                            conditions: (match.arguments || []).map((arg) => ({
                                param: arg.type || 'HEADER',
                                key: arg.key || '',
                                op: arg.value?.type || MatchType.EXACT,
                                value: arg.value?.value || '',
                            })),
                        },
                    },
                };
            }),
        },
    };
}

export function validateSecurityView(rule: TrafficSecurityRule, viewRules: SecurityViewRule[]): SecurityValidationError[] {
    const errors: SecurityValidationError[] = [];
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    if (!rule.name) errors.push({ field: 'name', message: '规则名称不能为空' });
    if (rule.name && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(rule.name)) {
        errors.push({ field: 'name', message: '规则名称需为 kebab-case' });
    }
    if (!target.namespace) errors.push({ field: 'target.namespace', message: '被调命名空间不能为空' });
    if (!target.service) errors.push({ field: 'target.service', message: '被调服务不能为空' });
    const normalized = normalizeSecurityViewOrder(viewRules);
    const storedAuthMode = readSecurityAuthMode(rule);
    const authMode = storedAuthMode === TrafficSecurityAuthMode.CUSTOM_HEADER
        ? TrafficSecurityAuthMode.LEGACY_REQUEST_MATCH
        : storedAuthMode;
    if (!normalized.length) errors.push({ field: 'rules', message: '至少配置 1 条鉴权子规则' });
    if (normalized.filter((item) => item.kind === 'service').length > 1) {
        errors.push({ field: 'service', message: '服务级规则最多只能配置 1 条' });
    }
    normalized.forEach((item, ruleIndex) => {
        if (item.kind !== 'service') {
            if (!item.interfaces.length) {
                errors.push({ field: `rules.${ruleIndex}.interfaces`, message: '接口级规则至少配置 1 个受保护接口' });
            }
            item.interfaces.forEach((api, apiIndex) => {
                if (!api.path?.value) {
                    errors.push({ field: `rules.${ruleIndex}.interfaces.${apiIndex}.path`, message: '接口路径不能为空' });
                }
            });
        } else if (item.interfaces.length) {
            errors.push({ field: `rules.${ruleIndex}.interfaces`, message: '服务级规则不能配置受保护接口' });
        }
        if (authMode === TrafficSecurityAuthMode.MANAGED_IDENTITY) {
            const caller = normalizeManagedCaller(item.managedCaller);
            if (!caller.any_authenticated && !(caller.callers || []).length) {
                errors.push({ field: `rules.${ruleIndex}.callers`, message: '请选择至少一个可信来源服务，或允许任意已认证服务' });
            }
        } else {
            const match = normalizeSecurityMatchRule(item.strategy);
            if (!(match.arguments || []).length) {
                errors.push({ field: `rules.${ruleIndex}.conditions`, message: '名单匹配策略至少配置 1 条匹配条件' });
            }
            (match.arguments || []).forEach((condition, conditionIndex) => {
                if (condition.value?.value_type !== MatchValueType.PARAMETER && !condition.value?.value) {
                    errors.push({ field: `rules.${ruleIndex}.conditions.${conditionIndex}.value`, message: '匹配条件的匹配值不能为空' });
                }
            });
        }
    });
    return errors;
}

function quoteYAMLValue(value: unknown): string {
    if (typeof value === 'boolean' || typeof value === 'number') return String(value);
    const text = String(value ?? '');
    if (text === '') return '""';
    if (/[:{}[\],&*#?|<>=!%@`"]|\s/.test(text)) return JSON.stringify(text);
    return text;
}

function stringifyYAMLValue(value: unknown, indent = 0): string {
    const pad = ' '.repeat(indent);
    if (Array.isArray(value)) {
        if (!value.length) return '[]';
        if (value.every(item => typeof item !== 'object' || item === null)) {
            return value.map(item => `${pad}- ${quoteYAMLValue(item)}`).join('\n');
        }
        return value.map(item => `${pad}- ${stringifyYAMLValue(item, indent + 2).trimStart()}`).join('\n');
    }
    if (typeof value === 'object' && value !== null) {
        const entries = Object.entries(value as Record<string, unknown>).filter(([, item]) => item !== undefined);
        if (!entries.length) return '{}';
        return entries.map(([key, item]) => {
            if (typeof item === 'object' && item !== null) {
                return `${pad}${key}: ${Array.isArray(item) && item.length === 0 ? '[]' : `\n${stringifyYAMLValue(item, indent + 2)}`}`;
            }
            return `${pad}${key}: ${quoteYAMLValue(item)}`;
        }).join('\n');
    }
    return `${pad}${quoteYAMLValue(value)}`;
}

export function stringifySecuritySpec(spec: SecurityPreviewSpec, format: TrafficSecuritySpecFormat): string {
    if (format === 'json') return JSON.stringify(spec, null, 2);
    return stringifyYAMLValue(spec);
}
