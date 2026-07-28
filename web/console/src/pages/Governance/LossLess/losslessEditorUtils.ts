import { Label } from 'services/types';

export type LosslessSpecFormat = 'yaml' | 'json';
export type LosslessDelayStrategy = 'DELAY_BY_TIME' | 'DELAY_BY_HEALTH_CHECK';
export type LosslessProbeProtocol = 'HTTP' | 'TCP' | 'UDP';
export type LosslessPayloadMatch = 'EXACT' | 'CONTAINS' | 'REGEX';

export interface LosslessProbePayloadDraft {
    request: string;
    response: string;
    match: LosslessPayloadMatch;
}

export interface LosslessRuleDraft {
    id?: string;
    service: string;
    namespace: string;
    description?: string;
    metadata?: Label[];
    lossless_online: {
        delay_register: {
            enable: boolean;
            strategy: LosslessDelayStrategy;
            interval: number;
            health_check_protocol: LosslessProbeProtocol;
            health_check_method: string;
            health_check_path: string;
            health_check_interval: number;
            payload: LosslessProbePayloadDraft;
        };
        warmup: {
            enable: boolean;
            interval: number;
            enable_overload_protection: boolean;
            overload_protection_threshold: number;
            curvature: number | '';
        };
    };
    lossless_offline: {
        enable: boolean;
        interval: number;
    };
}

export interface LosslessValidationError {
    field: string;
    message: string;
}

export interface LosslessPreviewSpec {
    kind: 'LosslessRule';
    metadata: {
        name: string;
        enabled: boolean;
        priority: number;
        tags: string[];
    };
    spec: {
        scope: {
            namespace: string;
            service: string;
        };
        description: string;
        online?: {
            delayedRegistration?: {
                enabled: boolean;
                strategy: 'TIME_DELAY' | 'PROBE_DELAY';
                delaySec?: number;
                healthCheck?: {
                    protocol: LosslessProbeProtocol;
                    method?: string;
                    path?: string;
                    intervalSec: number;
                    payload?: {
                        request: string;
                        expectedResponse: string;
                        match: string;
                    };
                };
            };
            warmup?: {
                enabled: boolean;
                durationSec: number;
                terminationProtect: boolean;
                terminationPercent?: number;
                curveValue?: string;
            };
        };
        offline?: {
            enabled: boolean;
            waitIntervalSec: number;
        };
    };
}

export interface LosslessSubmitPayload {
    id?: string;
    service: string;
    namespace: string;
    lossless_online: {
        delay_register: {
            enable: boolean;
            strategy: LosslessDelayStrategy;
            interval: string;
            health_check_protocol?: LosslessProbeProtocol;
            health_check_method?: string;
            health_check_path?: string;
            health_check_interval?: string;
        };
        warmup: {
            enable: boolean;
            interval: string;
            enable_overload_protection: boolean;
            overload_protection_threshold: number;
            curvature: number;
        };
    };
    lossless_offline: {
        enable: boolean;
        interval: string;
    };
    metadata?: Record<string, string>;
}

const allowedProtocols = new Set<string>(['HTTP', 'TCP', 'UDP']);
const allowedHttpMethods = new Set<string>(['GET', 'POST', 'HEAD']);
const allowedPayloadMatches = new Set<string>(['EXACT', 'CONTAINS', 'REGEX']);

export const delayStrategyOptions = [
    { label: '时长延迟', value: 'DELAY_BY_TIME' },
    { label: '探测延迟', value: 'DELAY_BY_HEALTH_CHECK' },
];

export const protocolOptions = ['HTTP', 'TCP', 'UDP'].map(item => ({ label: item, value: item }));
export const httpMethodOptions = ['GET', 'POST', 'HEAD'].map(item => ({ label: item, value: item }));

export const payloadMatchOptions = [
    { label: '完全匹配', value: 'EXACT' },
    { label: '包含匹配', value: 'CONTAINS' },
    { label: '正则匹配', value: 'REGEX' },
];

export const payloadMatchText: Record<string, string> = {
    EXACT: '完全匹配',
    CONTAINS: '包含匹配',
    REGEX: '正则匹配',
};

export function metadataToRecord(metadata?: Label[]): Record<string, string> {
    return (metadata || []).reduce<Record<string, string>>((acc, item) => {
        if (item.key) {
            acc[item.key] = item.value || '';
        }
        return acc;
    }, {});
}

function metadataToTags(metadata?: Label[]): string[] {
    return Object.entries(metadataToRecord(metadata)).map(([key, value]) => `${key}:${value}`);
}

export function secondsToDuration(value?: number | string): string {
    const num = Number(value || 0);
    return Number.isFinite(num) && num > 0 ? `${num}s` : '0s';
}

export function durationToSeconds(value?: string | number): number {
    if (typeof value === 'number') return value;
    return Number.parseInt((value || '0').replace(/s$/, ''), 10) || 0;
}

function normalizeProtocol(value?: string): LosslessProbeProtocol {
    const protocol = String(value || 'HTTP').toUpperCase();
    return (allowedProtocols.has(protocol) ? protocol : 'HTTP') as LosslessProbeProtocol;
}

function normalizeStrategy(value?: string): LosslessDelayStrategy {
    if (value === 'health_check' || value === 'PROBE_DELAY' || value === 'DELAY_BY_HEALTH_CHECK') {
        return 'DELAY_BY_HEALTH_CHECK';
    }
    return 'DELAY_BY_TIME';
}

function normalizePayloadMatch(value?: string): LosslessPayloadMatch {
    if (value === '完全匹配') return 'EXACT';
    if (value === '包含匹配') return 'CONTAINS';
    if (value === '正则匹配') return 'REGEX';
    return (allowedPayloadMatches.has(value || '') ? value : 'CONTAINS') as LosslessPayloadMatch;
}

export function defaultLosslessRuleDraft(): LosslessRuleDraft {
    return {
        service: '',
        namespace: '',
        description: '',
        lossless_online: {
            delay_register: {
                enable: true,
                strategy: 'DELAY_BY_HEALTH_CHECK',
                interval: 15,
                health_check_protocol: 'HTTP',
                health_check_method: 'GET',
                health_check_path: '/healthz',
                health_check_interval: 5,
                payload: {
                    request: 'PING\n',
                    response: 'PONG',
                    match: 'CONTAINS',
                },
            },
            warmup: {
                enable: true,
                interval: 0,
                enable_overload_protection: true,
                overload_protection_threshold: 80,
                curvature: '',
            },
        },
        lossless_offline: {
            enable: true,
            interval: 30,
        },
        metadata: [],
    };
}

export function normalizeLosslessRuleDraft(rule?: any): LosslessRuleDraft {
    const defaults = defaultLosslessRuleDraft();
    const online = rule?.lossless_online || rule?.losslessOnline || {};
    const delay = online.delay_register || online.delayRegister || {};
    const warmup = online.warmup || {};
    const offline = rule?.lossless_offline || rule?.losslessOffline || {};
    const metadata = Array.isArray(rule?.metadata)
        ? rule.metadata
        : Object.entries(rule?.metadata || {}).map(([key, value]) => ({ key, value: String(value ?? '') }));

    return {
        ...defaults,
        id: rule?.id,
        service: rule?.service || '',
        namespace: rule?.namespace || '',
        description: rule?.description || '',
        metadata,
        lossless_online: {
            delay_register: {
                ...defaults.lossless_online.delay_register,
                enable: delay.enable === true,
                strategy: normalizeStrategy(delay.strategy),
                interval: durationToSeconds(delay.interval ?? delay.interval_second ?? delay.intervalSecond),
                health_check_protocol: normalizeProtocol(delay.health_check_protocol ?? delay.healthCheckProtocol),
                health_check_method: String(delay.health_check_method ?? delay.healthCheckMethod ?? 'GET').toUpperCase(),
                health_check_path: delay.health_check_path ?? delay.healthCheckPath ?? '/healthz',
                health_check_interval: durationToSeconds(delay.health_check_interval ?? delay.health_check_interval_second ?? delay.healthCheckIntervalSecond),
                payload: {
                    request: delay.payload?.request ?? defaults.lossless_online.delay_register.payload.request,
                    response: delay.payload?.response ?? delay.payload?.expectedResponse ?? defaults.lossless_online.delay_register.payload.response,
                    match: normalizePayloadMatch(delay.payload?.match),
                },
            },
            warmup: {
                ...defaults.lossless_online.warmup,
                enable: warmup.enable === true,
                interval: durationToSeconds(warmup.interval ?? warmup.interval_second ?? warmup.intervalSecond),
                enable_overload_protection: warmup.enable_overload_protection ?? warmup.enableOverloadProtection ?? false,
                overload_protection_threshold: Number(warmup.overload_protection_threshold ?? warmup.overloadProtectionThreshold ?? 80),
                curvature: warmup.curvature === undefined || warmup.curvature === null || warmup.curvature === '' ? '' : Number(warmup.curvature),
            },
        },
        lossless_offline: {
            enable: offline.enable === true,
            interval: durationToSeconds(offline.interval ?? offline.interval_second ?? offline.intervalSecond),
        },
    };
}

function ruleDisplayName(draft: LosslessRuleDraft): string {
    if (draft.id) return draft.id;
    if (draft.namespace || draft.service) return `${draft.namespace || '-'}/${draft.service || '-'}`;
    return 'new-lossless-rule';
}

function buildDelayedRegistrationPreview(draft: LosslessRuleDraft): NonNullable<LosslessPreviewSpec['spec']['online']>['delayedRegistration'] | undefined {
    const delay = draft.lossless_online.delay_register;
    if (!delay.enable) return undefined;
    if (delay.strategy === 'DELAY_BY_TIME') {
        return {
            enabled: true,
            strategy: 'TIME_DELAY',
            delaySec: Number(delay.interval || 0),
        };
    }
    const healthCheck: NonNullable<NonNullable<LosslessPreviewSpec['spec']['online']>['delayedRegistration']>['healthCheck'] = {
        protocol: delay.health_check_protocol,
        intervalSec: Number(delay.health_check_interval || 0),
    };
    if (delay.health_check_protocol === 'HTTP') {
        healthCheck.method = delay.health_check_method || 'GET';
        healthCheck.path = delay.health_check_path || '';
    } else {
        healthCheck.payload = {
            request: delay.payload.request || '',
            expectedResponse: delay.payload.response || '',
            match: payloadMatchText[delay.payload.match] || '包含匹配',
        };
    }
    return {
        enabled: true,
        strategy: 'PROBE_DELAY',
        healthCheck,
    };
}

function buildWarmupPreview(draft: LosslessRuleDraft): NonNullable<LosslessPreviewSpec['spec']['online']>['warmup'] | undefined {
    const warmup = draft.lossless_online.warmup;
    if (!warmup.enable) return undefined;
    return {
        enabled: true,
        durationSec: Number(warmup.interval || 0),
        terminationProtect: warmup.enable_overload_protection === true,
        terminationPercent: warmup.enable_overload_protection ? Number(warmup.overload_protection_threshold || 0) : undefined,
        curveValue: warmup.curvature === '' ? undefined : String(warmup.curvature),
    };
}

export function buildLosslessPreviewSpec(draft: LosslessRuleDraft): LosslessPreviewSpec {
    const delayedRegistration = buildDelayedRegistrationPreview(draft);
    const warmup = buildWarmupPreview(draft);
    const online = delayedRegistration || warmup ? { delayedRegistration, warmup } : undefined;
    const offline = draft.lossless_offline.enable ? {
        enabled: true,
        waitIntervalSec: Number(draft.lossless_offline.interval || 0),
    } : undefined;

    return {
        kind: 'LosslessRule',
        metadata: {
            name: ruleDisplayName(draft),
            enabled: draft.lossless_online.delay_register.enable || draft.lossless_online.warmup.enable || draft.lossless_offline.enable,
            priority: 5,
            tags: metadataToTags(draft.metadata),
        },
        spec: {
            scope: {
                namespace: draft.namespace || '',
                service: draft.service || '',
            },
            description: draft.description || '',
            online,
            offline,
        },
    };
}

export function buildLosslessSubmitPayload(draft: LosslessRuleDraft): LosslessSubmitPayload {
    const delay = draft.lossless_online.delay_register;
    const protocol = normalizeProtocol(delay.health_check_protocol);
    const isHTTP = protocol === 'HTTP';
    const curvature = draft.lossless_online.warmup.curvature === '' ? 1 : Number(draft.lossless_online.warmup.curvature || 1);

    return {
        id: draft.id,
        service: draft.service,
        namespace: draft.namespace,
        lossless_online: {
            delay_register: {
                enable: delay.enable,
                strategy: normalizeStrategy(delay.strategy),
                interval: secondsToDuration(delay.interval),
                health_check_protocol: delay.enable && delay.strategy === 'DELAY_BY_HEALTH_CHECK' ? protocol : undefined,
                health_check_method: delay.enable && delay.strategy === 'DELAY_BY_HEALTH_CHECK' && isHTTP ? delay.health_check_method : undefined,
                health_check_path: delay.enable && delay.strategy === 'DELAY_BY_HEALTH_CHECK' && isHTTP ? delay.health_check_path : undefined,
                health_check_interval: delay.enable && delay.strategy === 'DELAY_BY_HEALTH_CHECK' ? secondsToDuration(delay.health_check_interval) : undefined,
            },
            warmup: {
                enable: draft.lossless_online.warmup.enable,
                interval: secondsToDuration(draft.lossless_online.warmup.interval),
                enable_overload_protection: draft.lossless_online.warmup.enable_overload_protection,
                overload_protection_threshold: Number(draft.lossless_online.warmup.overload_protection_threshold || 0),
                curvature,
            },
        },
        lossless_offline: {
            enable: draft.lossless_offline.enable,
            interval: secondsToDuration(draft.lossless_offline.interval),
        },
        metadata: metadataToRecord(draft.metadata),
    };
}

export function validateLosslessDraft(draft: LosslessRuleDraft): LosslessValidationError[] {
    const errors: LosslessValidationError[] = [];
    const delay = draft.lossless_online.delay_register;
    const warmup = draft.lossless_online.warmup;
    const offline = draft.lossless_offline;

    if (!draft.namespace) errors.push({ field: 'namespace', message: '主调命名空间不能为空' });
    if (!draft.service) errors.push({ field: 'service', message: '服务名称不能为空' });
    if (!delay.enable && !warmup.enable && !offline.enable) {
        errors.push({ field: 'capability', message: '至少启用一项无损能力' });
    }

    if (delay.enable) {
        if (delay.strategy === 'DELAY_BY_TIME') {
            if (Number(delay.interval) < 0) {
                errors.push({ field: 'delay.interval', message: '延迟注册时长不能小于 0' });
            }
        } else {
            if (!allowedProtocols.has(delay.health_check_protocol)) {
                errors.push({ field: 'delay.protocol', message: '检查协议必须是 HTTP / TCP / UDP' });
            }
            if (Number(delay.health_check_interval) <= 0) {
                errors.push({ field: 'delay.health_check_interval', message: '检查间隔必须大于 0' });
            }
            if (delay.health_check_protocol === 'HTTP') {
                if (!allowedHttpMethods.has(delay.health_check_method)) {
                    errors.push({ field: 'delay.method', message: 'HTTP 方法必须是 GET / POST / HEAD' });
                }
                if (!String(delay.health_check_path || '').trim()) {
                    errors.push({ field: 'delay.path', message: 'HTTP 检查路径不能为空' });
                }
            } else {
                if (!allowedPayloadMatches.has(delay.payload.match)) {
                    errors.push({ field: 'delay.payload.match', message: '匹配方式必须是完全匹配 / 包含匹配 / 正则匹配' });
                }
                if (!String(delay.payload.response || '').trim()) {
                    errors.push({ field: 'delay.payload.response', message: '响应匹配内容不能为空' });
                }
            }
        }
    }

    if (warmup.enable) {
        if (Number(warmup.interval) <= 0) {
            errors.push({ field: 'warmup.interval', message: '预热启用时预热时长必须大于 0' });
        }
        if (warmup.enable_overload_protection) {
            const percent = Number(warmup.overload_protection_threshold);
            if (percent < 0 || percent > 100) {
                errors.push({ field: 'warmup.threshold', message: '预热终止百分比必须在 0–100 之间' });
            }
        }
        if (warmup.curvature !== '') {
            const curvature = Number(warmup.curvature);
            if (!Number.isInteger(curvature) || curvature < 1 || curvature > 5) {
                errors.push({ field: 'warmup.curvature', message: '预热曲线值必须是 1～5 的整数' });
            }
        }
    }

    if (offline.enable && Number(offline.interval) <= 0) {
        errors.push({ field: 'offline.interval', message: '无损下线间隔必须大于 0' });
    }

    return errors;
}

export function describeDelaySummary(draft: LosslessRuleDraft): string {
    const delay = draft.lossless_online.delay_register;
    if (!delay.enable) return '已关闭，实例启动完成后立即注册到注册中心';
    if (delay.strategy === 'DELAY_BY_TIME') {
        return `注册条件：实例启动后等待 ${Number(delay.interval || 0)}s，到时直接进入服务发现列表。`;
    }
    if (delay.health_check_protocol === 'HTTP') {
        return `每 ${Number(delay.health_check_interval || 0)}s 调用 ${delay.health_check_method || 'GET'} ${delay.health_check_path || '-'}，健康检查成功后注册；失败期间实例不进入可发现列表。`;
    }
    return `每 ${Number(delay.health_check_interval || 0)}s 通过 ${delay.health_check_protocol} 发送报文，并按 ${payloadMatchText[delay.payload.match]} 比对响应，健康检查成功后注册；失败期间实例不进入可发现列表。`;
}

export function describeWarmupCurve(curvature: number | ''): string {
    if (curvature === '') return '默认曲线';
    const value = Number(curvature);
    if (value <= 1) return '近似线性增长';
    if (value <= 3) return '前段放量较稳，末段明显爬升';
    return '前段增长较慢，末段陡升';
}

export function describeWarmupSummary(draft: LosslessRuleDraft): string {
    const warmup = draft.lossless_online.warmup;
    if (!warmup.enable) return '已关闭，新实例注册后直接按完整权重接流';
    const protect = warmup.enable_overload_protection ? `开启，阈值 ${Number(warmup.overload_protection_threshold || 0)}%` : '关闭';
    const curve = warmup.curvature === '' ? '默认' : String(warmup.curvature);
    return `当前配置：注册后进入 ${Number(warmup.interval || 0)} Second 预热窗口；终止保护 ${protect}；曲线值 ${curve}，对应 ${describeWarmupCurve(warmup.curvature)}。`;
}

export function describeOfflineSummary(draft: LosslessRuleDraft): string {
    if (!draft.lossless_offline.enable) return '已关闭，实例注销时不等待存量请求完成';
    return `实例注销前等待 ${Number(draft.lossless_offline.interval || 0)} Second，让存量请求自然完成。`;
}

export function buildWarmupCurvePoints(curvature: number | ''): Array<{ progress: number; weight: number }> {
    const power = curvature === '' ? 3 : Number(curvature || 3);
    return Array.from({ length: 21 }).map((_, index) => {
        const progress = index * 5;
        const ratio = progress / 100;
        return {
            progress,
            weight: Math.min(100, Math.ceil(Math.abs(Math.pow(ratio, power) * 100))),
        };
    });
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

export function stringifyLosslessSpec(spec: LosslessPreviewSpec | LosslessSubmitPayload, format: LosslessSpecFormat): string {
    if (format === 'json') {
        return JSON.stringify(spec, null, 2);
    }
    return stringifyYAMLValue(spec);
}
