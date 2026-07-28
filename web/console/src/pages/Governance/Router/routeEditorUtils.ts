export type RouteSpecFormat = 'yaml' | 'json';

export interface RouteEditorLabel {
    key: string;
    value: string;
}

export interface RouteMatchValue {
    type: string;
    value: string;
    value_type: string;
}

export interface RouteMatchArgument {
    type: string;
    key: string;
    value: RouteMatchValue;
}

export interface RouteDestination {
    namespace?: string;
    service?: string;
    name: string;
    isolate: boolean;
    weight: number;
    priority?: number;
    labels?: Record<string, RouteMatchValue>;
}

export interface RouteRuleDraft {
    name: string;
    sources?: Array<{
        namespace?: string;
        service?: string;
        arguments?: RouteMatchArgument[];
    }>;
    arguments?: {
        matchMode?: string;
        randomPercent?: number;
        arguments?: RouteMatchArgument[];
    };
    destinations?: RouteDestination[];
}

export interface RouteEditorDraft {
    id?: string;
    name?: string;
    enable?: boolean;
    priority?: number;
    description?: string;
    metadata?: RouteEditorLabel[];
    routing_policy?: 'RulePolicy' | 'NearbyPolicy';
    caller_namespace?: string;
    caller_service?: string;
    callee_namespace?: string;
    callee_service?: string;
    routing_config?: {
        '@type'?: string;
        rules?: RouteRuleDraft[];
    };
}

export interface RouteRulePreviewSpec {
    id?: string;
    name: string;
    enable: boolean;
    priority: number;
    description: string;
    routing_policy: 'RulePolicy' | 'NearbyPolicy';
    metadata: Record<string, string>;
    routing_config: {
        '@type': string;
        caller: {
            namespace: string;
            service: string;
        };
        callee: {
            namespace: string;
            service: string;
        };
        rules: Array<{
            name: string;
            arguments: {
                arguments: RouteMatchArgument[];
                randomPercent: number;
                matchMode: string;
            };
            destinations: RouteDestination[];
        }>;
    };
}

export interface RouteRuleValidationError {
    field: string;
    message: string;
    ruleIndex?: number;
}

export const MATCH_TYPE_TEXT: Record<string, string> = {
    EXACT: '完全匹配',
    REGEX: '正则匹配',
    NOT_EQUALS: '不等于',
    IN: '包含',
    NOT_IN: '不包含',
    RANGE: '范围匹配',
    PREFIX: '前缀匹配',
};

export function getDefaultParamKey(paramType: string): string {
    return '';
}

export function withParamTypeDefaultKey(arg: RouteMatchArgument, nextType: string): RouteMatchArgument {
    return {
        ...arg,
        type: nextType,
        key: '',
    };
}

export function isTagInputMatchType(matchType?: string): boolean {
    return matchType === 'IN' || matchType === 'NOT_IN';
}

export function commaStringToTags(value?: string): string[] {
    return (value || '')
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean);
}

export function tagsToCommaString(value?: Array<string | number>): string {
    return (value || [])
        .map((item) => String(item).trim())
        .filter(Boolean)
        .join(',');
}

export function getRouteRuleArguments(rule?: RouteRuleDraft): RouteMatchArgument[] {
    return rule?.sources?.[0]?.arguments || rule?.arguments?.arguments || [];
}

export function getRuleWeightTotal(rule?: RouteRuleDraft): number {
    return (rule?.destinations || []).reduce((sum, item) => sum + Number(item.weight || 0), 0);
}

export function getWeightStatus(rule?: RouteRuleDraft) {
    const total = getRuleWeightTotal(rule);
    const delta = 100 - total;
    return {
        total,
        ok: total === 100,
        delta,
        message: total === 100 ? '已满 100%' : `需等于 100% · 差 ${Math.abs(delta)}%`,
    };
}

export function labelsToText(labels?: Record<string, RouteMatchValue>): string {
    return Object.entries(labels || {})
        .filter(([key, value]) => key && value?.value)
        .map(([key, value]) => `${key}=${value.value}`)
        .join(' ');
}

const defaultRoutingConfigType = 'type.googleapis.com/v1.CustomRoute';

function getAutoGroupName(index: number): string {
    return `Group ${index + 1}`;
}

function normalizeDestinationGroupNames(destinations?: RouteDestination[]): RouteDestination[] {
    return (destinations || []).map((group, index) => ({
        ...group,
        name: getAutoGroupName(index),
        priority: index,
    }));
}

function routeMetadataToRecord(metadata?: RouteEditorLabel[]): Record<string, string> {
    return (metadata || []).reduce<Record<string, string>>((acc, tag) => {
        if (tag.key) {
            acc[tag.key] = tag.value;
        }
        return acc;
    }, {});
}

export function textToLabels(text: string): Record<string, RouteMatchValue> {
    return text
        .split(/\s+/)
        .map((item) => item.trim())
        .filter(Boolean)
        .reduce<Record<string, RouteMatchValue>>((acc, item) => {
            const eqIdx = item.indexOf('=');
            if (eqIdx <= 0) {
                return acc;
            }
            const key = item.slice(0, eqIdx).trim();
            const value = item.slice(eqIdx + 1).trim();
            if (key && value) {
                acc[key] = {
                    type: 'EXACT',
                    value,
                    value_type: 'TEXT',
                };
            }
            return acc;
        }, {});
}

export function buildRouteRulePreviewSpec(draft: RouteEditorDraft): RouteRulePreviewSpec {
    const caller = {
        namespace: draft.caller_namespace || '',
        service: draft.caller_service || '',
    };
    const callee = {
        namespace: draft.callee_namespace || '',
        service: draft.callee_service || '',
    };

    return {
        id: draft.id,
        name: draft.name || '',
        enable: draft.enable !== false,
        priority: Number(draft.priority ?? 5),
        description: draft.description || '',
        routing_policy: draft.routing_policy || 'RulePolicy',
        metadata: routeMetadataToRecord(draft.metadata),
        routing_config: {
            '@type': draft.routing_config?.['@type'] || defaultRoutingConfigType,
            caller,
            callee,
            rules: (draft.routing_config?.rules || []).map((rule) => {
                const { sources: _sources, ...rest } = rule;
                return {
                    ...rest,
                    arguments: {
                        ...(rule.arguments || {}),
                        arguments: getRouteRuleArguments(rule),
                        randomPercent: rule.arguments?.randomPercent || 0,
                        matchMode: rule.arguments?.matchMode || 'AND',
                    },
                    destinations: normalizeDestinationGroupNames(rule.destinations),
                };
            }),
        },
    };
}

export function prepareRouteRuleDraftForSubmit(draft: RouteEditorDraft): RouteEditorDraft {
    return {
        ...draft,
        routing_config: {
            ...(draft.routing_config || {}),
            rules: (draft.routing_config?.rules || []).map((rule) => ({
                ...rule,
                destinations: normalizeDestinationGroupNames(rule.destinations),
            })),
        },
    };
}

export function buildRouteRuleSubmitPayload(draft: RouteEditorDraft): RouteRulePreviewSpec {
    return buildRouteRulePreviewSpec(prepareRouteRuleDraftForSubmit(draft));
}

export function validateRouteRuleDraft(draft: RouteEditorDraft): RouteRuleValidationError[] {
    const errors: RouteRuleValidationError[] = [];
    const name = draft.name || '';
    if (!/^[a-z][a-z0-9-]*$/.test(name)) {
        errors.push({ field: 'name', message: '规则名称必须为 kebab-case' });
    }

    (draft.routing_config?.rules || []).forEach((rule, idx) => {
        const ruleNumber = idx + 1;
        const total = getRuleWeightTotal(rule);
        if (total !== 100) {
            errors.push({ field: `rules.${idx}.weight`, ruleIndex: idx, message: `规则[${ruleNumber}] 权重合计 ${total}%` });
        }
        if ((rule.destinations || []).some((dest) => !dest.name?.trim())) {
            errors.push({ field: `rules.${idx}.destinations.name`, ruleIndex: idx, message: `规则[${ruleNumber}] 存在未命名分组` });
        }
        if (getRouteRuleArguments(rule).some((arg) => arg.value?.value_type !== 'PARAMETER' && !arg.value?.value?.trim())) {
            errors.push({ field: `rules.${idx}.arguments.value`, ruleIndex: idx, message: `规则[${ruleNumber}] 存在空匹配值` });
        }
    });

    return errors;
}

function quoteYAMLValue(value: unknown): string {
    if (typeof value === 'boolean' || typeof value === 'number') {
        return String(value);
    }
    const text = String(value ?? '');
    if (text === '') {
        return '""';
    }
    if (/[:{}[\],&*#?|<>=!%@`"]|\s/.test(text)) {
        return JSON.stringify(text);
    }
    return text;
}

function stringifyYAMLValue(value: unknown, indent = 0): string {
    const pad = ' '.repeat(indent);
    if (Array.isArray(value)) {
        if (value.length === 0) {
            return '[]';
        }
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
        if (entries.length === 0) {
            return '{}';
        }
        return entries.map(([key, item]) => {
            if (Array.isArray(item) || (item && typeof item === 'object')) {
                return `${pad}${key}:\n${stringifyYAMLValue(item, indent + 2)}`;
            }
            return `${pad}${key}: ${quoteYAMLValue(item)}`;
        }).join('\n');
    }
    return quoteYAMLValue(value);
}

export function stringifyRouteRuleSpec(spec: RouteRulePreviewSpec, format: RouteSpecFormat): string {
    if (format === 'json') {
        return JSON.stringify(spec, null, 2);
    }
    return stringifyYAMLValue(spec);
}
