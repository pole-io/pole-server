import {
    MirrorRule,
    TrafficApiScope,
    TrafficGovernanceRule,
    TrafficMatchRule,
    TrafficMirror,
} from 'services/traffic_governance';
import { InterfaceProtocol, MatchLogic, MatchString, MatchType, MatchValueType } from 'services/types';

export type TrafficMirrorSpecFormat = 'yaml' | 'json';

export interface MirrorCallerScope {
    namespace: string;
    service: string;
}

export interface MirrorValidationError {
    field: string;
    message: string;
}

export interface MirrorPreviewSpec {
    name: string;
    enable: boolean;
    priority: number;
    description: string;
    metadata: Record<string, string>;
    mirror_config: {
        caller: MirrorCallerScope;
        callee: {
            namespace: string;
            service: string;
        };
        rules: Array<{
            name: string;
            api: {
                protocol: string;
                method: string;
                path: string;
                op: string;
            };
            traffic_match_rule: {
                matchMode: string;
                arguments: Array<{
                    type: string;
                    key: string;
                    op: string;
                    value: string;
                }>;
            };
            mirror_percent: number;
            duration: string;
            destination: {
                namespace: string;
                service: string;
                labels: Record<string, string>;
            };
        }>;
    };
}

const callerServiceType = 'CALLER_SERVICE';

export function defaultMirrorCaller(): MirrorCallerScope {
    return { namespace: '*', service: '*' };
}

export function normalizeMirrorCaller(caller?: MirrorCallerScope): MirrorCallerScope {
    if (!caller) return defaultMirrorCaller();
    if (caller.namespace === '*' || caller.service === '*') return defaultMirrorCaller();
    return caller;
}

function defaultMirrorMatchValue(value = ''): MatchString {
    return {
        type: MatchType.EXACT,
        value,
        value_type: MatchValueType.TEXT,
    };
}

function normalizeMirrorMatchRule(match?: TrafficMatchRule): TrafficMatchRule {
    return {
        matchMode: match?.matchMode || MatchLogic.AND,
        randomPercent: match?.randomPercent ?? 0,
        arguments: match?.arguments || [],
    };
}

export function isAllMirrorCaller(caller?: MirrorCallerScope): boolean {
    return !caller || (caller.namespace === '*' && caller.service === '*');
}

export function extractMirrorCaller(rule?: TrafficMirror): MirrorCallerScope {
    const matchArg = (rule?.rules || [])
        .flatMap((item) => item.traffic_match_rule?.arguments || [])
        .find((arg) => arg.type === callerServiceType);
    if (!matchArg) return defaultMirrorCaller();
    return {
        namespace: matchArg.key || '*',
        service: matchArg.value?.value || '*',
    };
}

export function callerScopeText(caller?: MirrorCallerScope): string {
    const normalized = normalizeMirrorCaller(caller);
    if (isAllMirrorCaller(normalized)) return '全部服务';
    return `${normalized.namespace}/${normalized.service}`;
}

export function removeCallerServiceArguments(match?: TrafficMatchRule): TrafficMatchRule {
    const current = normalizeMirrorMatchRule(match);
    return {
        matchMode: current.matchMode,
        arguments: (current.arguments || []).filter((arg) => arg.type !== callerServiceType),
    };
}

export function applyMirrorCallerToMatchRule(match: TrafficMatchRule | undefined, caller: MirrorCallerScope): TrafficMatchRule {
    const next = removeCallerServiceArguments(match);
    const normalizedCaller = normalizeMirrorCaller(caller);
    if (isAllMirrorCaller(normalizedCaller)) {
        return next;
    }
    return {
        ...next,
        arguments: [{
            type: callerServiceType,
            key: normalizedCaller.namespace,
            value: defaultMirrorMatchValue(normalizedCaller.service),
        }, ...(next.arguments || [])],
    };
}

export function normalizeMirrorRules(rule: TrafficMirror): MirrorRule[] {
    return (rule.rules && rule.rules.length ? rule.rules : []).map((item) => ({
        ...item,
        api: {
            protocol: item.api?.protocol || InterfaceProtocol.HTTP,
            method: item.api?.method || 'GET',
            path: {
                ...defaultMirrorMatchValue('/'),
                ...(item.api?.path || {}),
            },
        },
        traffic_match_rule: removeCallerServiceArguments(item.traffic_match_rule),
        destination: {
            namespace: item.destination?.namespace || '',
            service: item.destination?.service || '',
            labels: item.destination?.labels || {},
        },
        mirror_percent: item.mirror_percent ?? 0,
        duration: item.duration || '0s',
        disable: item.disable ?? false,
    }));
}

export function buildMirrorRulesForSubmit(rules: MirrorRule[], caller: MirrorCallerScope): MirrorRule[] {
    return (rules || []).map((item) => ({
        ...item,
        traffic_match_rule: applyMirrorCallerToMatchRule(item.traffic_match_rule, caller),
    }));
}

function metadataToRecord(metadata?: Record<string, string>): Record<string, string> {
    return { ...(metadata || {}) };
}

function apiToPreview(api?: TrafficApiScope) {
    return {
        protocol: String(api?.protocol || InterfaceProtocol.HTTP),
        method: String(api?.method || 'GET'),
        path: api?.path?.value || '',
        op: api?.path?.type || MatchType.EXACT,
    };
}

function labelsToPreview(labels?: Record<string, MatchString>): Record<string, string> {
    return Object.entries(labels || {}).reduce<Record<string, string>>((acc, [key, value]) => {
        if (key) acc[key] = value?.value || '';
        return acc;
    }, {});
}

function matchToPreview(match?: TrafficMatchRule) {
    const current = normalizeMirrorMatchRule(removeCallerServiceArguments(match));
    return {
        matchMode: String(current.matchMode || MatchLogic.AND),
        arguments: (current.arguments || []).map((arg) => ({
            type: arg.type || 'HEADER',
            key: arg.key || '',
            op: arg.value?.type || MatchType.EXACT,
            value: arg.value?.value || '',
        })),
    };
}

export function buildMirrorPreviewSpec(rule: TrafficGovernanceRule, rules: MirrorRule[], caller: MirrorCallerScope): MirrorPreviewSpec {
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    const normalizedCaller = normalizeMirrorCaller(caller);
    return {
        name: rule.name || rule.id || '',
        enable: rule.enable !== false,
        priority: Number(rule.priority || 0),
        description: rule.description || '',
        metadata: metadataToRecord(rule.metadata),
        mirror_config: {
            caller: normalizedCaller,
            callee: {
                namespace: target.namespace || '',
                service: target.service || '',
            },
            rules: normalizeMirrorRules({ ...(rule as TrafficMirror), rules }).map((item, index) => ({
                name: `mirror-rule-${index + 1}`,
                api: apiToPreview(item.api),
                traffic_match_rule: matchToPreview(item.traffic_match_rule),
                mirror_percent: Number(item.mirror_percent || 0),
                duration: typeof item.duration === 'string' ? item.duration : `${item.duration?.seconds || 0}s`,
                destination: {
                    namespace: item.destination?.namespace || '',
                    service: item.destination?.service || '',
                    labels: labelsToPreview(item.destination?.labels),
                },
            })),
        },
    };
}

export function validateMirrorView(rule: TrafficGovernanceRule, rules: MirrorRule[], caller: MirrorCallerScope): MirrorValidationError[] {
    const errors: MirrorValidationError[] = [];
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    if (!rule.name) errors.push({ field: 'name', message: '规则名称不能为空' });
    if (!target.namespace) errors.push({ field: 'callee.namespace', message: '被调命名空间不能为空' });
    if (!target.service) errors.push({ field: 'callee.service', message: '被调服务不能为空' });
    if (!isAllMirrorCaller(caller) && !caller.namespace) {
        errors.push({ field: 'caller.namespace', message: '主调命名空间不能为空' });
    }
    if (!rules.length) errors.push({ field: 'rules', message: '至少配置 1 条镜像规则' });
    normalizeMirrorRules({ ...(rule as TrafficMirror), rules }).forEach((item, index) => {
        const ruleNumber = index + 1;
        if (!item.api?.path?.value?.trim()) {
            errors.push({ field: `rules.${index}.api.path`, message: `镜像规则[${ruleNumber}]接口路径不能为空` });
        }
        if (Number(item.mirror_percent || 0) < 0 || Number(item.mirror_percent || 0) > 100) {
            errors.push({ field: `rules.${index}.mirror_percent`, message: `镜像规则[${ruleNumber}]镜像比例必须在 0～100 之间` });
        }
        if (!item.destination?.namespace) {
            errors.push({ field: `rules.${index}.destination.namespace`, message: `镜像规则[${ruleNumber}]镜像目标命名空间不能为空` });
        }
        if (!item.destination?.service) {
            errors.push({ field: `rules.${index}.destination.service`, message: `镜像规则[${ruleNumber}]镜像目标服务不能为空` });
        }
        (removeCallerServiceArguments(item.traffic_match_rule).arguments || []).forEach((arg) => {
            if (!arg.value?.value?.trim()) {
                errors.push({ field: `rules.${index}.match.value`, message: `镜像规则[${ruleNumber}]存在空匹配值` });
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
