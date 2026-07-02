import {
    MirrorRule,
    TrafficApiScope,
    TrafficGovernanceRule,
    TrafficMatchRule,
    TrafficMirror,
} from 'services/traffic_governance';
import { InterfaceProtocol, MatchLogic, MatchString, MatchType, MatchValueType } from 'services/types';

export type TrafficMirrorSpecFormat = 'yaml' | 'json';

export interface MirrorServiceScope {
    namespace: string;
    service: string;
}

export type MirrorCallerScope = MirrorServiceScope;

export interface MirrorViewRule extends Omit<MirrorRule, 'api' | 'apis'> {
    interfaces: TrafficApiScope[];
}

export interface MirrorValidationError {
    field: string;
    message: string;
}

export interface MirrorPreviewSpec {
    apiVersion: string;
    kind: 'MirrorRule';
    metadata: {
        name: string;
        enabled: boolean;
        priority: number;
        labels: string[];
    };
    spec: {
        serviceRange: {
            caller: MirrorServiceScope;
            callee: MirrorServiceScope;
        };
        description: string;
        enabled: boolean;
        rules: Array<{
            interfaces: Array<{
                protocol: string;
                method: string;
                path: string;
                op: string;
            }>;
            trafficLabels: {
                relation: string;
                labels: Array<{
                    key: string;
                    op: string;
                    value: string;
                }>;
            };
            mirror: {
                percent: number;
                target: MirrorServiceScope;
            };
        }>;
    };
}

const callerServiceType = 'CALLER_SERVICE';
const allNamespaceText = '全部命名空间';
const allServiceText = '全部服务';

export function defaultMirrorCaller(): MirrorCallerScope {
    return { namespace: '*', service: '*' };
}

export function defaultMirrorCallee(): MirrorServiceScope {
    return { namespace: '', service: '' };
}

export function normalizeMirrorCaller(caller?: Partial<MirrorCallerScope>): MirrorCallerScope {
    if (!caller) return defaultMirrorCaller();
    if (caller.namespace === '*' || caller.service === '*') return defaultMirrorCaller();
    return {
        namespace: caller.namespace || '',
        service: caller.service || '',
    };
}

export function normalizeMirrorCallee(rule?: TrafficGovernanceRule): MirrorServiceScope {
    const mirror = rule as TrafficMirror | undefined;
    const callee = mirror?.callee || mirror?.target_service || { namespace: rule?.namespace || '', service: rule?.service || '' };
    return {
        namespace: callee?.namespace || '',
        service: callee?.service || '',
    };
}

export function mirrorScopeLabel(scope?: Partial<MirrorServiceScope>): string {
    const namespace = scope?.namespace === '*' ? allNamespaceText : scope?.namespace || '-';
    const service = scope?.service === '*' ? allServiceText : scope?.service || '-';
    return `${namespace}/${service}`;
}

export function isAllMirrorCaller(caller?: Partial<MirrorCallerScope>): boolean {
    return !caller || (caller.namespace === '*' && caller.service === '*');
}

function defaultMirrorMatchValue(value = ''): MatchString {
    return {
        type: MatchType.EXACT,
        value,
        value_type: MatchValueType.TEXT,
    };
}

function defaultMirrorApi(path = '/'): TrafficApiScope {
    return {
        protocol: InterfaceProtocol.HTTP,
        method: 'GET',
        path: defaultMirrorMatchValue(path),
    };
}

function normalizeMirrorMatchRule(match?: TrafficMatchRule): TrafficMatchRule {
    const args = match?.arguments && match.arguments.length ? match.arguments : [{
        type: 'HEADER',
        key: '',
        value: defaultMirrorMatchValue(),
    }];
    return {
        matchMode: match?.matchMode || MatchLogic.AND,
        arguments: args,
    };
}

export function extractMirrorCaller(rule?: TrafficMirror): MirrorCallerScope {
    if (rule?.caller) return normalizeMirrorCaller(rule.caller);
    const matchArg = (rule?.rules || [])
        .flatMap((item) => item.traffic_match_rule?.arguments || [])
        .find((arg) => arg.type === callerServiceType);
    if (!matchArg) return defaultMirrorCaller();
    return normalizeMirrorCaller({
        namespace: matchArg.key || '*',
        service: matchArg.value?.value || '*',
    });
}

export function callerScopeText(caller?: MirrorCallerScope): string {
    return mirrorScopeLabel(normalizeMirrorCaller(caller));
}

export function removeCallerServiceArguments(match?: TrafficMatchRule): TrafficMatchRule {
    const current = normalizeMirrorMatchRule(match);
    return {
        matchMode: current.matchMode,
        arguments: (current.arguments || []).filter((arg) => arg.type !== callerServiceType),
    };
}

export function defaultMirrorSubRule(): MirrorViewRule {
    return {
        interfaces: [defaultMirrorApi('/orders')],
        traffic_match_rule: normalizeMirrorMatchRule(),
        destination: {
            namespace: '',
            service: '',
            labels: {},
        },
        mirror_percent: 30,
        disable: false,
    };
}

function normalizeMirrorInterfaces(item?: MirrorRule): TrafficApiScope[] {
    const interfaces = item?.interfaces?.length ? item.interfaces : (item?.apis?.length ? item.apis : [item?.api || defaultMirrorApi()]);
    return interfaces.map((api) => ({
        protocol: api?.protocol || InterfaceProtocol.HTTP,
        method: api?.method || 'GET',
        path: {
            ...defaultMirrorMatchValue('/'),
            ...(api?.path || {}),
        },
    }));
}

export function normalizeMirrorRules(rule: TrafficMirror): MirrorViewRule[] {
    const rules = rule.rules && rule.rules.length ? rule.rules : [defaultMirrorSubRule()];
    return rules.map((item) => ({
        ...item,
        interfaces: normalizeMirrorInterfaces(item),
        traffic_match_rule: removeCallerServiceArguments(item.traffic_match_rule),
        destination: {
            namespace: item.destination?.namespace || '',
            service: item.destination?.service || '',
            labels: item.destination?.labels || {},
        },
        mirror_percent: item.mirror_percent ?? 0,
        disable: item.disable ?? false,
    }));
}

export function buildMirrorRulesForSubmit(rules: MirrorViewRule[]): MirrorRule[] {
    return (rules || []).map((item) => {
        const { api: _api, interfaces, duration: _duration, ...rest } = item as MirrorViewRule & { api?: TrafficApiScope; duration?: unknown };
        return {
            ...rest,
            apis: interfaces && interfaces.length ? interfaces : [defaultMirrorApi()],
            traffic_match_rule: removeCallerServiceArguments(rest.traffic_match_rule),
        };
    });
}

function metadataLabels(metadata?: Record<string, string>): string[] {
    return Object.entries(metadata || {}).map(([key, value]) => `${key}:${value}`);
}

function apiToPreview(api?: TrafficApiScope) {
    return {
        protocol: String(api?.protocol || InterfaceProtocol.HTTP),
        method: String(api?.method || 'GET'),
        path: api?.path?.value || '',
        op: api?.path?.type || MatchType.EXACT,
    };
}

function matchToPreview(match?: TrafficMatchRule) {
    const current = normalizeMirrorMatchRule(removeCallerServiceArguments(match));
    return {
        relation: String(current.matchMode || MatchLogic.AND),
        labels: (current.arguments || []).map((arg) => ({
            key: arg.key || '',
            op: arg.value?.type || MatchType.EXACT,
            value: arg.value?.value || '',
        })),
    };
}

export function buildMirrorPreviewSpec(rule: TrafficGovernanceRule, rules: MirrorViewRule[], caller: MirrorCallerScope): MirrorPreviewSpec {
    const callee = normalizeMirrorCallee(rule);
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'MirrorRule',
        metadata: {
            name: rule.name || rule.id || '',
            enabled: rule.enable !== false,
            priority: Number(rule.priority || 0),
            labels: metadataLabels(rule.metadata),
        },
        spec: {
            serviceRange: {
                caller: normalizeMirrorCaller(caller),
                callee,
            },
            description: rule.description || '',
            enabled: rule.enable !== false,
            rules: normalizeMirrorRules({ ...(rule as TrafficMirror), rules }).map((item) => ({
                interfaces: item.interfaces.map(apiToPreview),
                trafficLabels: matchToPreview(item.traffic_match_rule),
                mirror: {
                    percent: Number(item.mirror_percent || 0),
                    target: {
                        namespace: item.destination?.namespace || '',
                        service: item.destination?.service || '',
                    },
                },
            })),
        },
    };
}

export function validateMirrorView(rule: TrafficGovernanceRule, rules: MirrorViewRule[], caller: MirrorCallerScope): MirrorValidationError[] {
    const errors: MirrorValidationError[] = [];
    const callee = normalizeMirrorCallee(rule);
    if (!/^[a-z][a-z0-9-]*$/.test(rule.name || '')) {
        errors.push({ field: 'name', message: '规则名称必须符合 kebab-case' });
    }
    if (!caller.namespace) errors.push({ field: 'caller.namespace', message: '主调命名空间不能为空' });
    if (!caller.service) errors.push({ field: 'caller.service', message: '主调服务不能为空' });
    if (!callee.namespace) errors.push({ field: 'callee.namespace', message: '被调命名空间不能为空' });
    if (!callee.service) errors.push({ field: 'callee.service', message: '被调服务不能为空' });
    if (!rules.length) errors.push({ field: 'rules', message: '至少配置 1 条镜像子规则' });
    normalizeMirrorRules({ ...(rule as TrafficMirror), rules }).forEach((item, index) => {
        const ruleNumber = index + 1;
        if (!item.interfaces.length) {
            errors.push({ field: `rules.${index}.interfaces`, message: `镜像子规则[${ruleNumber}]至少配置 1 个接口` });
        }
        item.interfaces.forEach((api, apiIndex) => {
            if (!api.path?.value?.trim()) {
                errors.push({ field: `rules.${index}.interfaces.${apiIndex}.path`, message: `镜像子规则[${ruleNumber}]存在空接口路径` });
            }
        });
        const matchArgs = removeCallerServiceArguments(item.traffic_match_rule).arguments || [];
        if (!matchArgs.length) {
            errors.push({ field: `rules.${index}.trafficLabels`, message: `镜像子规则[${ruleNumber}]至少配置 1 条流量标签` });
        }
        matchArgs.forEach((arg) => {
            if (!arg.key?.trim()) {
                errors.push({ field: `rules.${index}.match.key`, message: `镜像子规则[${ruleNumber}]存在空标签键` });
            }
            if (!arg.value?.value?.trim()) {
                errors.push({ field: `rules.${index}.match.value`, message: `镜像子规则[${ruleNumber}]存在空标签值` });
            }
        });
        if (Number(item.mirror_percent || 0) < 0 || Number(item.mirror_percent || 0) > 100) {
            errors.push({ field: `rules.${index}.mirror_percent`, message: `镜像子规则[${ruleNumber}]镜像比例必须在 0～100 之间` });
        }
        if (!item.destination?.namespace) {
            errors.push({ field: `rules.${index}.destination.namespace`, message: `镜像子规则[${ruleNumber}]镜像目标命名空间不能为空` });
        }
        if (!item.destination?.service) {
            errors.push({ field: `rules.${index}.destination.service`, message: `镜像子规则[${ruleNumber}]镜像目标服务不能为空` });
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
        if (value.length === 0) return '[]';
        return value.map((item) => {
            if (item && typeof item === 'object') {
                const nested = stringifyYAMLValue(item, indent + 2);
                return `${pad}- ${nested.startsWith('\n') ? nested.slice(1) : nested.replace(new RegExp(`^ {${indent + 2}}`), '')}`;
            }
            return `${pad}- ${quoteYAMLValue(item)}`;
        }).join('\n');
    }
    if (value && typeof value === 'object') {
        const entries = Object.entries(value as Record<string, unknown>).filter(([, item]) => item !== undefined);
        if (!entries.length) return '{}';
        return entries.map(([key, item]) => {
            if (Array.isArray(item) || (item && typeof item === 'object')) {
                return `${pad}${key}:\n${stringifyYAMLValue(item, indent + 2)}`;
            }
            return `${pad}${key}: ${quoteYAMLValue(item)}`;
        }).join('\n');
    }
    return quoteYAMLValue(value);
}

export function stringifyMirrorSpec(spec: MirrorPreviewSpec, format: TrafficMirrorSpecFormat): string {
    if (format === 'json') return JSON.stringify(spec, null, 2);
    return stringifyYAMLValue(spec);
}
