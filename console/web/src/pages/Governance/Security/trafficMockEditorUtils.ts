import {
    MockRule,
    TrafficApiScope,
    TrafficGovernanceRule,
    TrafficMatchRule,
    TrafficMock,
} from 'services/traffic_governance';
import { InterfaceProtocol, MatchLogic, MatchString, MatchType, MatchValueType } from 'services/types';

export type TrafficMockSpecFormat = 'yaml' | 'json';

export interface MockValidationError {
    field: string;
    message: string;
}

export interface MockCallerScope {
    namespace: string;
    service: string;
}

export interface MockPreviewSpec {
    apiVersion: 'governance.pole.io/v1';
    kind: 'MockRule';
    metadata: {
        name: string;
        enabled: boolean;
        priority: number;
        labels: string[];
    };
    spec: {
        scope: {
            caller: MockCallerScope;
            callee: MockCallerScope;
        };
        description: string;
        rules: Array<{
            name: string;
            interface: {
                protocol: string;
                method: string;
                path: string;
                op: string;
            };
            match: {
                relation: string;
                conditions: Array<{
                    param: string;
                    key: string;
                    op: string;
                    value: string;
                }>;
            };
            response: {
                delayMs: number;
                statusCode: string;
                headers: Record<string, string>;
                body: string;
            };
        }>;
    };
}

const callerServiceType = 'CALLER_SERVICE';

export const mockMatchSourceOptions = ['HEADER', 'QUERY', 'PATH', 'COOKIE', 'METHOD'].map((item) => ({ label: item, value: item }));

export function defaultMockCaller(): MockCallerScope {
    return { namespace: '*', service: '*' };
}

export function normalizeMockCaller(caller?: MockCallerScope): MockCallerScope {
    if (!caller) return defaultMockCaller();
    if (caller.namespace === '*' || caller.service === '*') return defaultMockCaller();
    return caller;
}

export function isAllMockCaller(caller?: MockCallerScope): boolean {
    return !caller || (caller.namespace === '*' && caller.service === '*');
}

export function mockCallerScopeText(caller?: MockCallerScope): string {
    const normalized = normalizeMockCaller(caller);
    if (isAllMockCaller(normalized)) return '全部服务';
    return `${normalized.namespace}/${normalized.service}`;
}

export function defaultMockMatchValue(): MatchString {
    return {
        type: MatchType.EXACT,
        value: '',
        value_type: MatchValueType.TEXT,
    };
}

export function defaultMockApi(): TrafficApiScope {
    return {
        protocol: InterfaceProtocol.HTTP,
        method: 'GET',
        path: {
            type: MatchType.EXACT,
            value: '',
            value_type: MatchValueType.TEXT,
        },
    };
}

export function defaultMockArgument() {
    return {
        type: 'HEADER',
        key: 'x-mock',
        value: defaultMockMatchValue(),
    };
}

export function defaultMockMatchRule(): TrafficMatchRule {
    return {
        matchMode: MatchLogic.AND,
        arguments: [defaultMockArgument()],
    };
}

export function defaultMockSubRule(): MockRule {
    return {
        api: defaultMockApi(),
        traffic_match_rule: defaultMockMatchRule(),
        response: {
            code: '200',
            headers: { 'content-type': 'application/json' },
            body: '{}',
        },
        delay: '0s',
        mock_percent: 100,
        disable: false,
    };
}

export function durationToMs(value?: string | { seconds?: number | string; nanos?: number }): number {
    if (!value) return 0;
    if (typeof value === 'string') {
        const match = value.trim().match(/^(-?\d+(?:\.\d+)?)s$/);
        if (!match) return 0;
        return Math.max(0, Math.round(Number(match[1]) * 1000));
    }
    const seconds = Number(value.seconds ?? 0);
    const nanos = Number(value.nanos ?? 0);
    if (!Number.isFinite(seconds) || !Number.isFinite(nanos)) return 0;
    return Math.max(0, Math.round(seconds * 1000 + nanos / 1000000));
}

export function msToDuration(value?: number): string {
    const ms = Math.max(0, Number(value ?? 0));
    const seconds = Math.floor(ms / 1000);
    const restMs = Math.round(ms - seconds * 1000);
    if (!restMs) return `${seconds}s`;
    const nanos = String(restMs * 1000000).padStart(9, '0').replace(/0+$/, '');
    return `${seconds}.${nanos}s`;
}

export function normalizeMockMatchRule(match?: TrafficMatchRule): TrafficMatchRule {
    return {
        matchMode: match?.matchMode || MatchLogic.AND,
        arguments: match?.arguments?.filter((item) => item.type !== callerServiceType).length
            ? match.arguments?.filter((item) => item.type !== callerServiceType)
            : [defaultMockArgument()],
    };
}

export function removeMockCallerServiceArguments(match?: TrafficMatchRule): TrafficMatchRule {
    const normalized = normalizeMockMatchRule(match);
    return {
        ...normalized,
        arguments: (normalized.arguments || []).filter((item) => item.type !== callerServiceType),
    };
}

export function extractMockCaller(rule?: TrafficMock): MockCallerScope {
    const callerArgument = (rule?.rules || [])
        .flatMap((item) => item.traffic_match_rule?.arguments || [])
        .find((item) => item.type === callerServiceType);
    if (!callerArgument?.value?.value) return defaultMockCaller();
    return {
        namespace: callerArgument.key || '*',
        service: callerArgument.value.value || '*',
    };
}

export function applyMockCallerToMatchRule(match: TrafficMatchRule | undefined, caller: MockCallerScope): TrafficMatchRule {
    const normalized = removeMockCallerServiceArguments(match);
    const normalizedCaller = normalizeMockCaller(caller);
    if (isAllMockCaller(normalizedCaller)) return normalized;
    return {
        ...normalized,
        arguments: [
            {
                type: callerServiceType,
                key: normalizedCaller.namespace,
                value: {
                    type: MatchType.EXACT,
                    value: normalizedCaller.service,
                    value_type: MatchValueType.TEXT,
                },
            },
            ...(normalized.arguments || []),
        ],
    };
}

export function normalizeMockRule(rule?: MockRule): MockRule {
    const source = rule || defaultMockSubRule();
    const response = source.response || {};
    const legacyResponse = response as Record<string, unknown>;
    return {
        ...source,
        api: {
            ...defaultMockApi(),
            ...(source.api || {}),
            path: {
                ...defaultMockMatchValue(),
                ...(source.api?.path || {}),
            },
        },
        traffic_match_rule: normalizeMockMatchRule(source.traffic_match_rule),
        response: {
            code: response.code !== undefined ? String(response.code) : String(legacyResponse.status_code || '200'),
            headers: { ...(response.headers || {}) },
            body: response.body || '',
        },
        delay: msToDuration(durationToMs(source.delay)),
        mock_percent: 100,
        disable: source.disable ?? false,
    };
}

export function normalizeMockRules(rule: TrafficMock): MockRule[] {
    return (rule.rules && rule.rules.length ? rule.rules : [defaultMockSubRule()]).map(normalizeMockRule);
}

function cleanHeaders(headers?: Record<string, string>): Record<string, string> {
    return Object.entries(headers || {}).reduce<Record<string, string>>((acc, [key, value]) => {
        const normalizedKey = key.trim();
        if (normalizedKey) acc[normalizedKey] = value;
        return acc;
    }, {});
}

export function buildMockRulesForSubmit(rules: MockRule[], caller?: MockCallerScope): MockRule[] {
    return (rules.length ? rules : [defaultMockSubRule()]).map((item) => {
        const current = normalizeMockRule(item);
        const match = caller
            ? applyMockCallerToMatchRule(current.traffic_match_rule, caller)
            : removeMockCallerServiceArguments(current.traffic_match_rule);
        return {
            api: current.api,
            traffic_match_rule: {
                matchMode: match.matchMode,
                arguments: match.arguments || [],
            },
            response: {
                code: current.response?.code || '200',
                headers: cleanHeaders(current.response?.headers),
                body: current.response?.body || '',
            },
            delay: msToDuration(durationToMs(current.delay)),
            mock_percent: 100,
            disable: current.disable ?? false,
        };
    });
}

function metadataToLabels(metadata?: Record<string, string>): string[] {
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
    const current = removeMockCallerServiceArguments(match);
    return {
        relation: String(current.matchMode || MatchLogic.AND),
        conditions: (current.arguments || []).map((arg) => ({
            param: arg.type || 'HEADER',
            key: arg.key || '',
            op: arg.value?.type || MatchType.EXACT,
            value: arg.value?.value || '',
        })),
    };
}

export function buildMockPreviewSpec(rule: TrafficGovernanceRule, viewRules: MockRule[], caller = extractMockCaller(rule as TrafficMock)): MockPreviewSpec {
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    const normalizedCaller = normalizeMockCaller(caller);
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'MockRule',
        metadata: {
            name: rule.name || rule.id || '',
            enabled: rule.enable !== false,
            priority: Number(rule.priority || 0),
            labels: metadataToLabels(rule.metadata),
        },
        spec: {
            scope: {
                caller: {
                    namespace: normalizedCaller.namespace,
                    service: normalizedCaller.service,
                },
                callee: {
                    namespace: target.namespace || '',
                    service: target.service || '',
                },
            },
            description: rule.description || '',
            rules: normalizeMockRules({ ...(rule as TrafficMock), rules: viewRules }).map((item, index) => ({
                name: `mock-subrule-${index + 1}`,
                interface: apiToPreview(item.api),
                match: matchToPreview(item.traffic_match_rule),
                response: {
                    delayMs: durationToMs(item.delay),
                    statusCode: item.response?.code || '',
                    headers: cleanHeaders(item.response?.headers),
                    body: item.response?.body || '',
                },
            })),
        },
    };
}

function findContentType(headers?: Record<string, string>): string {
    const entry = Object.entries(headers || {}).find(([key]) => key.trim().toLowerCase() === 'content-type');
    return entry?.[1] || '';
}

function isValidJson(text: string): boolean {
    try {
        JSON.parse(text);
        return true;
    } catch {
        return false;
    }
}

export function validateMockView(rule: TrafficGovernanceRule, viewRules: MockRule[], caller = defaultMockCaller()): MockValidationError[] {
    const errors: MockValidationError[] = [];
    const target = rule.target_service || { namespace: rule.namespace || '', service: rule.service || '' };
    if (!rule.name) errors.push({ field: 'name', message: '规则名称不能为空' });
    if (!target.namespace) errors.push({ field: 'target.namespace', message: 'Mock 服务命名空间不能为空' });
    if (!target.service) errors.push({ field: 'target.service', message: 'Mock 服务名称不能为空' });
    if (!isAllMockCaller(caller) && (!caller.namespace || !caller.service)) {
        errors.push({ field: 'caller', message: 'Mock 主调方必须同时选择命名空间和服务' });
    }
    const normalized = normalizeMockRules({ ...(rule as TrafficMock), rules: viewRules });
    if (!normalized.length) errors.push({ field: 'rules', message: '至少配置 1 条 Mock 子规则' });
    normalized.forEach((item, index) => {
        const displayIndex = index + 1;
        if (!item.api?.path?.value) {
            errors.push({ field: `rules.${index}.api.path`, message: `Mock 子规则[${displayIndex}] 接口路径不能为空` });
        }
        if (durationToMs(item.delay) < 0) {
            errors.push({ field: `rules.${index}.delay`, message: `Mock 子规则[${displayIndex}] 返回延迟不能小于 0` });
        }
        if (!item.response?.code) {
            errors.push({ field: `rules.${index}.response.code`, message: `Mock 子规则[${displayIndex}] 响应 Code 不能为空` });
        }
        const match = normalizeMockMatchRule(item.traffic_match_rule);
        if (!(match.arguments || []).length) {
            errors.push({ field: `rules.${index}.conditions`, message: `Mock 子规则[${displayIndex}] 请至少配置 1 个流量匹配条件` });
        }
        (match.arguments || []).forEach((condition, conditionIndex) => {
            if (!condition.value?.value) {
                errors.push({ field: `rules.${index}.conditions.${conditionIndex}.value`, message: `Mock 子规则[${displayIndex}] 流量匹配条件存在空匹配值` });
            }
        });
        Object.entries(item.response?.headers || {}).forEach(([key, value]) => {
            if (!key.trim() && value) {
                errors.push({ field: `rules.${index}.headers`, message: `Mock 子规则[${displayIndex}] 响应头存在空 Header 名` });
            }
        });
        const body = item.response?.body || '';
        if (body && findContentType(item.response?.headers).toLowerCase().includes('json') && !isValidJson(body)) {
            errors.push({ field: `rules.${index}.body`, message: `Mock 子规则[${displayIndex}] Content-Type 为 JSON 时，响应正文必须是合法 JSON` });
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

export function stringifyMockSpec(spec: MockPreviewSpec, format: TrafficMockSpecFormat): string {
    if (format === 'json') return JSON.stringify(spec, null, 2);
    return stringifyYAMLValue(spec);
}
