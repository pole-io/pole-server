import {
    TrafficApiScope,
    TrafficGovernanceRule,
    TrafficMatchRule,
    TrafficSecurityAction,
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
                match: {
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
        value_type: MatchValueType.TEXT,
    },
});

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
    const api = policy.api;
    if (!api) return true;
    const path = api.path?.value;
    return path === undefined || path === null || path === '';
}

export function normalizeSecurityViewRules(rule: TrafficSecurityRule): SecurityViewRule[] {
    const policies = rule.policies || [];
    const interfaceRules = policies
        .filter((policy) => !isServiceLevelPolicy(policy))
        .map<SecurityViewRule>((policy) => ({
            kind: actionToListType(policy.action) === 'DENY_LIST' ? 'deny' : 'allow',
            listType: actionToListType(policy.action),
            interfaces: [policy.api || defaultProtectedInterface()],
            strategy: normalizeSecurityMatchRule(policy.traffic_match_rule),
            rejectEffect: policy.reject_effect,
        }));
    const serviceRule = policies.find(isServiceLevelPolicy);
    const serviceRules = serviceRule ? [{
        kind: 'service' as const,
        listType: actionToListType(serviceRule.action),
        interfaces: [],
        strategy: normalizeSecurityMatchRule(serviceRule.traffic_match_rule),
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

export function buildSecurityPoliciesFromView(rules: SecurityViewRule[]): TrafficSecurityPolicy[] {
    return normalizeSecurityViewOrder(rules).flatMap((viewRule) => {
        const action = listTypeToAction(viewRule.listType);
        const base = {
            traffic_match_rule: normalizeSecurityMatchRule(viewRule.strategy),
            action,
            reject_effect: viewRule.rejectEffect || (action === TrafficSecurityAction.DENY ? {
                status_code: 403,
                code: 'FORBIDDEN',
                message: 'request denied by auth rule',
            } : undefined),
        };
        if (viewRule.kind === 'service') {
            return [{ ...base }];
        }
        return (viewRule.interfaces.length ? viewRule.interfaces : [defaultProtectedInterface()]).map((api) => ({
            ...base,
            api,
        }));
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
            subRules: normalizeSecurityViewOrder(viewRules).map((item, index) => {
                const match = normalizeSecurityMatchRule(item.strategy);
                return {
                    name: `auth-subrule-${index + 1}`,
                    listType: item.listType,
                    protectedInterfaces: item.kind === 'service' ? [] : item.interfaces.map(protectedInterfaceToPreview),
                    strategy: {
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

export function validateSecurityView(rule: TrafficGovernanceRule, viewRules: SecurityViewRule[]): SecurityValidationError[] {
    const errors: SecurityValidationError[] = [];
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    if (!rule.name) errors.push({ field: 'name', message: '规则名称不能为空' });
    if (!target.namespace) errors.push({ field: 'target.namespace', message: '被调命名空间不能为空' });
    if (!target.service) errors.push({ field: 'target.service', message: '被调服务不能为空' });
    const normalized = normalizeSecurityViewOrder(viewRules);
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
        const match = normalizeSecurityMatchRule(item.strategy);
        if (!(match.arguments || []).length) {
            errors.push({ field: `rules.${ruleIndex}.conditions`, message: '名单匹配策略至少配置 1 条匹配条件' });
        }
        (match.arguments || []).forEach((condition, conditionIndex) => {
            if (!condition.value?.value) {
                errors.push({ field: `rules.${ruleIndex}.conditions.${conditionIndex}.value`, message: '匹配条件的匹配值不能为空' });
            }
        });
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
