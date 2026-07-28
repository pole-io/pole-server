import { FaultDetectProtocol, FaultDetectRule, FaultDetectSubRule } from 'services/faultdetect';

export type FaultDetectSpecFormat = 'yaml' | 'json';
export type FaultDetectPayloadMatch = 'EXACT' | 'CONTAINS' | 'REGEX';

export interface FaultDetectValidationError {
    field: string;
    message: string;
    ruleIndex?: number;
}

export interface FaultDetectPreviewSpec {
    apiVersion: string;
    kind: 'FaultDetectRule';
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
        rules: Array<{
            name: string;
            enabled: boolean;
            protocol: string;
            intervalSec: number;
            timeoutSec: number;
            portMode: 'INSTANCE_PROTOCOL_PORT' | 'CUSTOM';
            port?: number;
            http?: {
                method: string;
                path: string;
                headers?: Record<string, string>;
            };
            payload?: {
                request: string;
                expectedResponse: string;
                match: string;
            };
        }>;
    };
}

const allowedProtocols = new Set<string>([FaultDetectProtocol.HTTP, FaultDetectProtocol.TCP, FaultDetectProtocol.UDP]);
const allowedPayloadMatches = new Set<string>(['EXACT', 'CONTAINS', 'REGEX']);

export const PayloadMatchOptions = [
    { label: '完全匹配', value: 'EXACT' },
    { label: '包含匹配', value: 'CONTAINS' },
    { label: '正则匹配', value: 'REGEX' },
];

export const PayloadMatchText: Record<string, string> = {
    EXACT: '完全匹配',
    CONTAINS: '包含匹配',
    REGEX: '正则匹配',
};

export function receiveToText(receive?: string[] | string): string {
    if (Array.isArray(receive)) {
        return receive.join('\n');
    }
    return receive || '';
}

export function receiveToArray(receive?: string[] | string): string[] {
    if (Array.isArray(receive)) {
        return receive.filter(item => item !== undefined && item !== null).map(item => String(item));
    }
    const text = receive === undefined || receive === null ? '' : String(receive);
    return text ? [text] : [];
}

export function normalizeProbePort(port?: number | string | null): number {
    if (port === undefined || port === null || port === '') {
        return 0;
    }
    const num = Number(port);
    return Number.isFinite(num) ? num : 0;
}

export function isCustomProbePort(rule?: Pick<FaultDetectSubRule, 'port'>): boolean {
    return normalizeProbePort(rule?.port) > 0;
}

export function describeProbePort(rule?: Pick<FaultDetectSubRule, 'port'>): string {
    const port = normalizeProbePort(rule?.port);
    return port > 0 ? `指定端口 ${port}` : '实例协议端口';
}

export function describeProbeSummary(rule: FaultDetectSubRule): string {
    const protocol = rule.protocol || FaultDetectProtocol.HTTP;
    const timing = `${Number(rule.interval || 30)}s 间隔 / ${Number(rule.timeout || 60)}s 超时`;
    const status = rule.disable ? '禁用' : '启用';
    return `${protocol} · ${describeProbePort(rule)} · ${timing} · ${status}`;
}

function normalizePayloadMatch(match?: string): FaultDetectPayloadMatch {
    if (match === '包含匹配') return 'CONTAINS';
    if (match === '正则匹配') return 'REGEX';
    if (match === '完全匹配') return 'EXACT';
    return (allowedPayloadMatches.has(match || '') ? match : 'EXACT') as FaultDetectPayloadMatch;
}

function normalizeHttpConfig(rule: FaultDetectSubRule) {
    return {
        method: rule.httpConfig?.method || 'GET',
        url: rule.httpConfig?.url || '',
        headers: (rule.httpConfig?.headers || [])
            .filter(Boolean)
            .map(item => ({ key: item.key || '', value: item.value || '' })),
        body: rule.httpConfig?.body || '',
    };
}

function normalizeTcpConfig(rule: FaultDetectSubRule) {
    return {
        send: rule.tcpConfig?.send || '',
        receive: receiveToArray(rule.tcpConfig?.receive),
        match: normalizePayloadMatch(rule.tcpConfig?.match),
    };
}

function normalizeUdpConfig(rule: FaultDetectSubRule) {
    return {
        send: rule.udpConfig?.send || '',
        receive: receiveToArray(rule.udpConfig?.receive),
        match: normalizePayloadMatch(rule.udpConfig?.match),
    };
}

export function normalizeFaultDetectSubRuleDraft(rule?: Partial<FaultDetectSubRule>): FaultDetectSubRule {
    const protocol = allowedProtocols.has(rule?.protocol || '') ? rule?.protocol || FaultDetectProtocol.HTTP : FaultDetectProtocol.HTTP;
    return {
        interval: Number(rule?.interval || 30),
        timeout: Number(rule?.timeout || 60),
        port: normalizeProbePort(rule?.port),
        protocol,
        httpConfig: normalizeHttpConfig(rule as FaultDetectSubRule),
        tcpConfig: normalizeTcpConfig(rule as FaultDetectSubRule),
        udpConfig: normalizeUdpConfig(rule as FaultDetectSubRule),
        disable: rule?.disable === true,
    };
}

export function normalizeFaultDetectRulesDraft(rules?: Partial<FaultDetectSubRule>[]): FaultDetectSubRule[] {
    const items = Array.isArray(rules) ? rules : [];
    if (items.length === 0) {
        return [normalizeFaultDetectSubRuleDraft({})];
    }
    return items.map(normalizeFaultDetectSubRuleDraft);
}

export function buildFaultDetectSubmitPayload(draft: Partial<FaultDetectRule>): FaultDetectRule {
    const rules = normalizeFaultDetectRulesDraft(draft.rules);
    return {
        ...(draft as FaultDetectRule),
        name: draft.name || '',
        description: draft.description || '',
        priority: Number(draft.priority || 0),
        metadata: draft.metadata || {},
        targetService: {
            namespace: draft.targetService?.namespace || '',
            service: draft.targetService?.service || '',
            api: draft.targetService?.api,
        },
        rules: rules.map(rule => ({
            interval: Number(rule.interval || 30),
            timeout: Number(rule.timeout || 60),
            port: normalizeProbePort(rule.port),
            protocol: rule.protocol || FaultDetectProtocol.HTTP,
            httpConfig: {
                method: rule.httpConfig?.method || 'GET',
                url: rule.httpConfig?.url || '',
                headers: rule.httpConfig?.headers || [],
                body: rule.httpConfig?.body || '',
            },
            tcpConfig: {
                send: rule.tcpConfig?.send || '',
                receive: receiveToArray(rule.tcpConfig?.receive),
            },
            udpConfig: {
                send: rule.udpConfig?.send || '',
                receive: receiveToArray(rule.udpConfig?.receive),
            },
            disable: rule.disable === true,
        })),
    };
}

function metadataToLabels(metadata?: Record<string, string>): string[] {
    return Object.entries(metadata || {})
        .filter(([key]) => key)
        .map(([key, value]) => `${key}:${value}`);
}

function headersToRecord(headers?: Array<{ key: string; value: string }>): Record<string, string> | undefined {
    const out = (headers || []).reduce<Record<string, string>>((acc, item) => {
        if (item.key) {
            acc[item.key] = item.value || '';
        }
        return acc;
    }, {});
    return Object.keys(out).length ? out : undefined;
}

export function buildFaultDetectPreviewSpec(draft: Partial<FaultDetectRule>): FaultDetectPreviewSpec {
    const rules = normalizeFaultDetectRulesDraft(draft.rules);
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'FaultDetectRule',
        metadata: {
            name: draft.name || '',
            enabled: rules.some(rule => !rule.disable),
            priority: Number(draft.priority || 0),
            labels: metadataToLabels(draft.metadata),
        },
        spec: {
            scope: {
                namespace: draft.targetService?.namespace || '',
                service: draft.targetService?.service || '',
            },
            description: draft.description || '',
            rules: rules.map((rule, index) => {
                const base = {
                    name: `probe-${index + 1}`,
                    enabled: rule.disable !== true,
                    protocol: rule.protocol || FaultDetectProtocol.HTTP,
                    intervalSec: Number(rule.interval || 30),
                    timeoutSec: Number(rule.timeout || 60),
                    portMode: isCustomProbePort(rule) ? 'CUSTOM' as const : 'INSTANCE_PROTOCOL_PORT' as const,
                    port: isCustomProbePort(rule) ? normalizeProbePort(rule.port) : undefined,
                };

                if (rule.protocol === FaultDetectProtocol.TCP || rule.protocol === FaultDetectProtocol.UDP) {
                    const config = rule.protocol === FaultDetectProtocol.TCP ? normalizeTcpConfig(rule) : normalizeUdpConfig(rule);
                    return {
                        ...base,
                        payload: {
                            request: config.send,
                            expectedResponse: receiveToText(config.receive),
                            match: PayloadMatchText[config.match] || PayloadMatchText.EXACT,
                        },
                    };
                }

                return {
                    ...base,
                    http: {
                        method: rule.httpConfig?.method || 'GET',
                        path: rule.httpConfig?.url || '',
                        headers: headersToRecord(rule.httpConfig?.headers),
                    },
                };
            }),
        },
    };
}

function hasBlankHeader(headers?: Array<{ key: string; value: string }>): boolean {
    return (headers || []).some(item => {
        const hasKey = Boolean(item.key?.trim());
        const hasValue = Boolean(item.value?.trim());
        return hasKey !== hasValue;
    });
}

export function validateFaultDetectDraft(draft: Partial<FaultDetectRule>): FaultDetectValidationError[] {
    const errors: FaultDetectValidationError[] = [];
    if (!/^[a-z][a-z0-9-]*$/.test(draft.name || '')) {
        errors.push({ field: 'name', message: '规则名称必须为 kebab-case' });
    }
    if (!draft.targetService?.namespace) {
        errors.push({ field: 'targetService.namespace', message: '请选择被探测命名空间' });
    }
    if (!draft.targetService?.service) {
        errors.push({ field: 'targetService.service', message: '请选择被探测服务' });
    }

    const rules = draft.rules || [];
    if (rules.length === 0) {
        errors.push({ field: 'rules', message: '至少配置 1 条探测规则' });
    }

    rules.forEach((rule, index) => {
        const num = index + 1;
        if (!allowedProtocols.has(rule.protocol || '')) {
            errors.push({ field: `rules.${index}.protocol`, ruleIndex: index, message: `探测规则[${num}] 探测协议必须是 HTTP、TCP 或 UDP` });
        }
        if (Number(rule.interval || 0) <= 0) {
            errors.push({ field: `rules.${index}.interval`, ruleIndex: index, message: `探测规则[${num}] 间隔必须大于 0` });
        }
        if (Number(rule.timeout || 0) <= 0) {
            errors.push({ field: `rules.${index}.timeout`, ruleIndex: index, message: `探测规则[${num}] 超时必须大于 0` });
        }
        const port = normalizeProbePort(rule.port);
        if (port < 0 || port > 65535) {
            errors.push({ field: `rules.${index}.port`, ruleIndex: index, message: `探测规则[${num}] 端口需在 1-65535 之间，留空表示实例协议端口` });
        }

        if ((rule.protocol || FaultDetectProtocol.HTTP) === FaultDetectProtocol.HTTP) {
            if (!rule.httpConfig?.url?.trim()) {
                errors.push({ field: `rules.${index}.httpConfig.url`, ruleIndex: index, message: `探测规则[${num}] HTTP URL 不能为空` });
            }
            if (hasBlankHeader(rule.httpConfig?.headers)) {
                errors.push({ field: `rules.${index}.httpConfig.headers`, ruleIndex: index, message: `探测规则[${num}] Headers 不允许出现空 key 或空 value` });
            }
        }

        if (rule.protocol === FaultDetectProtocol.TCP || rule.protocol === FaultDetectProtocol.UDP) {
            const config = rule.protocol === FaultDetectProtocol.TCP ? rule.tcpConfig : rule.udpConfig;
            const match = normalizePayloadMatch(config?.match);
            if (!allowedPayloadMatches.has(match)) {
                errors.push({ field: `rules.${index}.${rule.protocol.toLowerCase()}Config.match`, ruleIndex: index, message: `探测规则[${num}] 匹配方式不合法` });
            }
            if (!receiveToText(config?.receive).trim()) {
                errors.push({ field: `rules.${index}.${rule.protocol.toLowerCase()}Config.receive`, ruleIndex: index, message: `探测规则[${num}] 接收内容不能为空` });
            }
        }
    });

    return errors;
}

function quoteYAMLValue(value: unknown): string {
    if (value === undefined) {
        return '';
    }
    if (value === null) {
        return 'null';
    }
    if (typeof value === 'number' || typeof value === 'boolean') {
        return String(value);
    }
    const text = String(value);
    if (text === '') {
        return '""';
    }
    if (/[:#\n\r]|^\s|\s$|^\{|\[/.test(text)) {
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
        return value.map(item => {
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

export function stringifyFaultDetectSpec(spec: FaultDetectPreviewSpec, format: FaultDetectSpecFormat): string {
    if (format === 'json') {
        return JSON.stringify(spec, null, 2);
    }
    return stringifyYAMLValue(spec);
}
