export type CircuitBreakerSpecFormat = 'yaml' | 'json';

export interface CircuitBreakerMatchValue {
    type: string;
    value: string;
    value_type?: string;
}

export interface CircuitBreakerAPI {
    protocol?: string;
    method?: string;
    path?: CircuitBreakerMatchValue;
}

export interface CircuitBreakerErrorCondition {
    inputType: string;
    condition: {
        type: string;
        value: string;
    };
}

export interface CircuitBreakerTriggerCondition {
    triggerType: string;
    triggerVal?: number;
    errorCount?: number;
    errorPercent?: number;
    interval?: number;
    minimumRequest?: number;
}

export interface CircuitBreakerRecoverCondition {
    sleepWindow?: number;
    consecutiveSuccess?: number;
}

export interface CircuitBreakerFaultDetectConfig {
    enable: boolean;
}

export interface CircuitBreakerFallbackConfig {
    enable: boolean;
    response: {
        code: number;
        headers: Array<{ key: string; value: string }>;
        body: string;
    };
}

export interface CircuitBreakerStrategyDraft {
    name: string;
    ifaces: CircuitBreakerAPI[];
    error_conditions: CircuitBreakerErrorCondition[];
    trigger_conditions: CircuitBreakerTriggerCondition[];
}

export interface CircuitBreakerSubRuleDraft {
    strategies: CircuitBreakerStrategyDraft[];
    max_ejection_percent?: number;
    recoverCondition: CircuitBreakerRecoverCondition;
    faultDetectConfig: CircuitBreakerFaultDetectConfig;
    fallbackConfig: CircuitBreakerFallbackConfig;
}

export interface CircuitBreakerDraftLike {
    id?: string;
    name?: string;
    level?: string;
    description?: string;
    priority?: number;
    ruleMatcher?: {
        source?: {
            service?: string;
            namespace?: string;
        };
        destination?: {
            service?: string;
            namespace?: string;
            method?: {
                type?: string;
                value?: string;
            };
        };
    };
    metadata?: Record<string, string> | Array<{ key: string; value: string }>;
    subrules?: CircuitBreakerSubRuleDraft[];
    block_configs?: CircuitBreakerPolicyLike[];
    blockConfigs?: CircuitBreakerPolicyLike[];
}

export interface CircuitBreakerPolicyLike {
    block_config?: {
        name?: string;
        api?: CircuitBreakerAPI;
        error_conditions?: CircuitBreakerErrorCondition[];
        errorConditions?: CircuitBreakerErrorCondition[];
        trigger_conditions?: CircuitBreakerTriggerCondition[];
        triggerConditions?: CircuitBreakerTriggerCondition[];
    };
    blockConfig?: CircuitBreakerPolicyLike['block_config'];
    name?: string;
    api?: CircuitBreakerAPI;
    error_conditions?: CircuitBreakerErrorCondition[];
    errorConditions?: CircuitBreakerErrorCondition[];
    trigger_conditions?: CircuitBreakerTriggerCondition[];
    triggerConditions?: CircuitBreakerTriggerCondition[];
    max_ejection_percent?: number;
    maxEjectionPercent?: number;
    recoverCondition?: CircuitBreakerRecoverCondition;
    recover_condition?: CircuitBreakerRecoverCondition;
    faultDetectConfig?: CircuitBreakerFaultDetectConfig;
    fault_detect_config?: CircuitBreakerFaultDetectConfig;
    fallbackConfig?: CircuitBreakerFallbackConfig;
    fallback_config?: CircuitBreakerFallbackConfig;
}

export interface CircuitBreakerSubmitPayload {
    id?: string;
    name: string;
    level: string;
    description: string;
    priority: number;
    ruleMatcher: {
        source: {
            service: string;
            namespace: string;
        };
        destination: {
            service: string;
            namespace: string;
            method: {
                type: string;
                value: string;
            };
        };
    };
    metadata: Record<string, string>;
    block_configs: Array<{
        block_config: {
            name: string;
            api?: CircuitBreakerAPI;
            error_conditions: CircuitBreakerErrorCondition[];
            trigger_conditions: CircuitBreakerTriggerCondition[];
        };
        max_ejection_percent?: number;
        recoverCondition: CircuitBreakerRecoverCondition;
        faultDetectConfig: CircuitBreakerFaultDetectConfig;
        fallbackConfig: CircuitBreakerFallbackConfig;
    }>;
}

export interface CircuitBreakerValidationError {
    field: string;
    message: string;
    subRuleIndex?: number;
    strategyIndex?: number;
}

const defaultAPI = (): CircuitBreakerAPI => ({
    protocol: 'HTTP',
    method: 'GET',
    path: { type: 'EXACT', value: '', value_type: 'TEXT' },
});

export const defaultCircuitBreakerStrategy = (idx: number): CircuitBreakerStrategyDraft => ({
    name: `策略-${idx}`,
    ifaces: [defaultAPI()],
    error_conditions: [
        {
            inputType: 'RET_CODE',
            condition: {
                type: 'RANGE',
                value: '500-599',
            },
        },
    ],
    trigger_conditions: [
        {
            triggerType: 'ERROR_RATE',
            errorCount: 0,
            errorPercent: 50,
            triggerVal: 50,
            interval: 30,
            minimumRequest: 5,
        },
    ],
});

export const defaultCircuitBreakerSubRule = (idx: number): CircuitBreakerSubRuleDraft => ({
    strategies: [defaultCircuitBreakerStrategy(idx)],
    max_ejection_percent: 100,
    recoverCondition: {
        sleepWindow: 60,
        consecutiveSuccess: 0,
    },
    faultDetectConfig: {
        enable: false,
    },
    fallbackConfig: {
        enable: false,
        response: {
            code: 500,
            headers: [],
            body: '',
        },
    },
});

function normalizeMetadata(metadata?: CircuitBreakerDraftLike['metadata']): Record<string, string> {
    if (Array.isArray(metadata)) {
        return metadata.reduce<Record<string, string>>((acc, item) => {
            if (item.key && item.value) {
                acc[item.key] = item.value;
            }
            return acc;
        }, {});
    }
    return metadata || {};
}

function normalizeAPI(api?: CircuitBreakerAPI): CircuitBreakerAPI {
    return {
        protocol: api?.protocol || 'HTTP',
        method: api?.method || 'GET',
        path: {
            type: api?.path?.type || 'EXACT',
            value: api?.path?.value || '',
            value_type: api?.path?.value_type || 'TEXT',
        },
    };
}

function normalizeTrigger(condition: CircuitBreakerTriggerCondition): CircuitBreakerTriggerCondition {
    const triggerType = condition.triggerType || 'ERROR_RATE';
    const errorPercent = Number(condition.errorPercent ?? condition.triggerVal ?? 0);
    const errorCount = Number(condition.errorCount ?? condition.triggerVal ?? 0);
    return {
        triggerType,
        errorCount: triggerType === 'ERROR_RATE' ? 0 : errorCount,
        errorPercent: triggerType === 'ERROR_RATE' ? errorPercent : 0,
        triggerVal: triggerType === 'ERROR_RATE' ? errorPercent : errorCount,
        interval: Number(condition.interval || 0),
        minimumRequest: Number(condition.minimumRequest || 0),
    };
}

function normalizeFallback(config?: CircuitBreakerFallbackConfig): CircuitBreakerFallbackConfig {
    return {
        enable: config?.enable === true,
        response: {
            code: Number(config?.response?.code ?? 500),
            headers: config?.response?.headers || [],
            body: config?.response?.body || '',
        },
    };
}

function policyToSubRule(policy: CircuitBreakerPolicyLike, idx: number): CircuitBreakerSubRuleDraft {
    const block = policy.block_config ?? policy.blockConfig ?? policy;
    return {
        strategies: [{
            name: block?.name || `策略-${idx + 1}`,
            ifaces: [normalizeAPI(block?.api)],
            error_conditions: block?.error_conditions ?? block?.errorConditions ?? [],
            trigger_conditions: (block?.trigger_conditions ?? block?.triggerConditions ?? []).map(normalizeTrigger),
        }],
        max_ejection_percent: policy.max_ejection_percent ?? policy.maxEjectionPercent ?? 100,
        recoverCondition: policy.recoverCondition ?? policy.recover_condition ?? { sleepWindow: 60, consecutiveSuccess: 0 },
        faultDetectConfig: policy.faultDetectConfig ?? policy.fault_detect_config ?? { enable: false },
        fallbackConfig: normalizeFallback(policy.fallbackConfig ?? policy.fallback_config),
    };
}

export function createCircuitBreakerDraftFromRule(rule: CircuitBreakerDraftLike): CircuitBreakerDraftLike & { subrules: CircuitBreakerSubRuleDraft[] } {
    const policies = rule.block_configs ?? rule.blockConfigs ?? [];
    return {
        ...rule,
        name: rule.name || '',
        level: rule.level || 'METHOD',
        description: rule.description || '',
        priority: Number(rule.priority || 0),
        metadata: normalizeMetadata(rule.metadata),
        ruleMatcher: {
            source: {
                service: rule.ruleMatcher?.source?.service || '',
                namespace: rule.ruleMatcher?.source?.namespace || '*',
            },
            destination: {
                service: rule.ruleMatcher?.destination?.service || '',
                namespace: rule.ruleMatcher?.destination?.namespace || '*',
                method: {
                    type: rule.ruleMatcher?.destination?.method?.type || 'EXACT',
                    value: rule.ruleMatcher?.destination?.method?.value || '',
                },
            },
        },
        subrules: rule.subrules?.length ? rule.subrules : policies.map(policyToSubRule),
    };
}

function strategyToPolicies(subrule: CircuitBreakerSubRuleDraft, strategy: CircuitBreakerStrategyDraft) {
    const ifaces = strategy.ifaces?.length ? strategy.ifaces : [undefined];
    return ifaces.map((api) => ({
        block_config: {
            name: strategy.name || '',
            api: api ? normalizeAPI(api) : undefined,
            error_conditions: strategy.error_conditions || [],
            trigger_conditions: (strategy.trigger_conditions || []).map(normalizeTrigger),
        },
        max_ejection_percent: subrule.max_ejection_percent,
        recoverCondition: subrule.recoverCondition || { sleepWindow: 60, consecutiveSuccess: 0 },
        faultDetectConfig: { enable: subrule.faultDetectConfig?.enable === true },
        fallbackConfig: normalizeFallback(subrule.fallbackConfig),
    }));
}

export function buildCircuitBreakerSubmitPayload(draft: CircuitBreakerDraftLike): CircuitBreakerSubmitPayload {
    const normalized = createCircuitBreakerDraftFromRule(draft);
    return {
        id: normalized.id,
        name: normalized.name || '',
        level: normalized.level || 'METHOD',
        description: normalized.description || '',
        priority: Number(normalized.priority || 0),
        ruleMatcher: normalized.ruleMatcher,
        metadata: normalizeMetadata(normalized.metadata),
        block_configs: (normalized.subrules || []).flatMap((subrule) =>
            (subrule.strategies || []).flatMap((strategy) => strategyToPolicies(subrule, strategy)),
        ),
    };
}

export function describeSubRuleSummary(subrule: CircuitBreakerSubRuleDraft): string {
    const strategyCount = subrule.strategies?.length || 0;
    const sleepWindow = subrule.recoverCondition?.sleepWindow ?? '-';
    const fallback = subrule.fallbackConfig?.enable ? '开' : '关';
    return `${strategyCount} 个熔断策略 · 熔断 ${sleepWindow}s · 降级${fallback}`;
}

export function describeStrategySummary(strategy: CircuitBreakerStrategyDraft): string {
    return `${strategy.ifaces?.length || 0} 个接口 · ${strategy.error_conditions?.length || 0} 个错误条件 · ${strategy.trigger_conditions?.length || 0} 个触发条件`;
}

export function validateCircuitBreakerDraft(draft: CircuitBreakerDraftLike): CircuitBreakerValidationError[] {
    const normalized = createCircuitBreakerDraftFromRule(draft);
    const errors: CircuitBreakerValidationError[] = [];
    if (!/^[a-z][a-z0-9-]*$/.test(normalized.name || '')) {
        errors.push({ field: 'name', message: '规则名称必须为 kebab-case' });
    }

    const subrules = normalized.subrules || [];
    if (subrules.length === 0) {
        errors.push({ field: 'subrules', message: '至少配置 1 条熔断子规则' });
    }

    subrules.forEach((subrule, subRuleIndex) => {
        const subRuleNo = subRuleIndex + 1;
        if ((subrule.strategies || []).length === 0) {
            errors.push({ field: `subrules.${subRuleIndex}.strategies`, subRuleIndex, message: `子规则[${subRuleNo}] 至少配置 1 个熔断策略` });
        }
        (subrule.strategies || []).forEach((strategy, strategyIndex) => {
            const strategyNo = strategyIndex + 1;
            const prefix = `子规则[${subRuleNo}]·策略[${strategyNo}]`;
            if ((strategy.ifaces || []).some((api) => !api.path?.value?.trim())) {
                errors.push({ field: `subrules.${subRuleIndex}.strategies.${strategyIndex}.ifaces`, subRuleIndex, strategyIndex, message: `${prefix} 存在空的接口路径` });
            }
            if ((strategy.error_conditions || []).some((item) => !item.condition?.value?.trim())) {
                errors.push({ field: `subrules.${subRuleIndex}.strategies.${strategyIndex}.error_conditions`, subRuleIndex, strategyIndex, message: `${prefix} 存在空错误判断值` });
            }
            if ((strategy.trigger_conditions || []).length === 0) {
                errors.push({ field: `subrules.${subRuleIndex}.strategies.${strategyIndex}.trigger_conditions`, subRuleIndex, strategyIndex, message: `${prefix} 至少配置 1 个触发条件` });
            }
            (strategy.trigger_conditions || []).forEach((condition) => {
                const value = Number(condition.triggerType === 'ERROR_RATE' ? condition.errorPercent ?? condition.triggerVal : condition.errorCount ?? condition.triggerVal);
                if (condition.triggerType === 'ERROR_RATE' && (value < 0 || value > 100)) {
                    errors.push({ field: `subrules.${subRuleIndex}.strategies.${strategyIndex}.trigger_conditions.threshold`, subRuleIndex, strategyIndex, message: `${prefix} 比例类触发阈值必须在 0-100 之间` });
                }
            });
        });
        if (Number(subrule.recoverCondition?.sleepWindow || 0) <= 0) {
            errors.push({ field: `subrules.${subRuleIndex}.recoverCondition.sleepWindow`, subRuleIndex, message: `子规则[${subRuleNo}] 恢复策略熔断时长必须大于 0` });
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

export function stringifyCircuitBreakerSpec(spec: CircuitBreakerSubmitPayload, format: CircuitBreakerSpecFormat): string {
    if (format === 'json') {
        return JSON.stringify(spec, null, 2);
    }
    return stringifyYAMLValue(spec);
}
