export type RateLimitSpecFormat = 'yaml' | 'json';

export interface RateLimitMatchValue {
    type: string;
    value: string;
    value_type: string;
}

export interface RateLimitArgument {
    type: string;
    key: string;
    value: RateLimitMatchValue;
}

export interface RateLimitAmountView {
    validDuration: number;
    validDurationUnit: string;
    maxAmount: number;
}

export interface RateLimitTriggerViewLike {
    method?: RateLimitMatchValue;
    arguments?: RateLimitArgument[];
    amounts?: RateLimitAmountView[];
    action?: string;
    resource?: string;
    concurrencyAmount?: {
        maxAmount?: number;
    };
    regex_combine?: boolean;
    failover?: string;
    max_queue_delay?: number;
    customResponse?: {
        code?: string;
        headers?: Record<string, string>;
        body?: string;
    };
}

export interface RateLimitDraftLike {
    id?: string;
    name?: string;
    service?: string;
    namespace?: string;
    priority?: number;
    type?: string;
    disable?: boolean;
    report?: unknown;
    cluster?: unknown;
    metadata?: Record<string, string>;
    rules?: RateLimitTriggerViewLike[];
}

export interface RateLimitSubmitPayload {
    id?: string;
    name: string;
    service: string;
    namespace: string;
    priority: number;
    type: string;
    disable: boolean;
    report?: unknown;
    cluster?: unknown;
    metadata?: Record<string, string>;
    rules: Array<{
        method: RateLimitMatchValue;
        arguments: RateLimitArgument[];
        amounts?: Array<{
            validDuration: string;
            maxAmount: number;
        }>;
        action: string;
        resource: string;
        concurrencyAmount?: {
            maxAmount?: number;
        };
        regex_combine: boolean;
        failover?: string;
        max_queue_delay?: number;
        customResponse?: {
            code?: string;
            headers?: Record<string, string>;
            body?: string;
        };
    }>;
}

export interface RateLimitValidationError {
    field: string;
    message: string;
    ruleIndex?: number;
}

const resourceText: Record<string, string> = {
    QPS: '请求数',
    CONCURRENCY: '并发数',
};

const actionText: Record<string, string> = {
    REJECT: '快速失败',
    UNIRATE: '排队等待',
};

function normalizeMethod(method?: RateLimitMatchValue): RateLimitMatchValue {
    return {
        type: method?.type || 'EXACT',
        value: method?.value || '',
        value_type: method?.value_type || 'TEXT',
    };
}

function normalizeAction(rule: RateLimitTriggerViewLike, limitType?: string): string {
    if (limitType === 'GLOBAL') {
        return 'REJECT';
    }
    return rule.action || 'REJECT';
}

function amountToPayload(amount: RateLimitAmountView) {
    return {
        validDuration: `${Number(amount.validDuration || 1)}${amount.validDurationUnit || 's'}`,
        maxAmount: Number(amount.maxAmount || 0),
    };
}

export function buildRateLimitSubmitPayload(draft: RateLimitDraftLike): RateLimitSubmitPayload {
    const type = draft.type || 'LOCAL';
    const rules = (draft.rules || []).map((rule) => {
        const payloadRule: RateLimitSubmitPayload['rules'][number] = {
            method: normalizeMethod(rule.method),
            arguments: rule.arguments || [],
            action: normalizeAction(rule, type),
            resource: rule.resource || 'QPS',
            regex_combine: rule.regex_combine === true,
            failover: rule.failover || 'FAILOVER_LOCAL',
            max_queue_delay: rule.max_queue_delay,
            customResponse: rule.customResponse,
        };

        if (payloadRule.resource === 'CONCURRENCY') {
            payloadRule.amounts = [];
            payloadRule.concurrencyAmount = rule.concurrencyAmount || { maxAmount: 1 };
        } else {
            payloadRule.amounts = (rule.amounts || []).map(amountToPayload);
        }

        return payloadRule;
    });

    return {
        id: draft.id,
        name: draft.name || '',
        service: draft.service || '',
        namespace: draft.namespace || '',
        priority: Number(draft.priority ?? 0),
        type,
        disable: draft.disable === true,
        report: draft.report,
        cluster: draft.cluster,
        metadata: draft.metadata || {},
        rules,
    };
}

export function describeRuleThreshold(rule: RateLimitTriggerViewLike): string {
    if (rule.resource === 'CONCURRENCY') {
        return `${rule.concurrencyAmount?.maxAmount ?? rule.amounts?.[0]?.maxAmount ?? '-'} 并发`;
    }
    const firstAmount = rule.amounts?.[0];
    if (!firstAmount) {
        return '未配置阈值';
    }
    const suffix = (rule.amounts?.length || 0) > 1 ? ` 等 ${rule.amounts?.length} 窗` : '';
    return `${firstAmount.maxAmount} 次 / ${firstAmount.validDuration}${firstAmount.validDurationUnit}${suffix}`;
}

export function describeRuleSummary(rule: RateLimitTriggerViewLike, limitType?: string): string {
    const ifaceCount = rule.method?.value?.trim() ? 1 : 0;
    const conditionCount = rule.arguments?.length || 0;
    const metric = resourceText[rule.resource || 'QPS'] || '请求数';
    const action = actionText[normalizeAction(rule, limitType)] || normalizeAction(rule, limitType);
    return `${ifaceCount} 个接口 · ${conditionCount} 个匹配条件 · ${metric} · ${action}`;
}

function hasInvalidCustomResponse(rule: RateLimitTriggerViewLike): boolean {
    const body = rule.customResponse?.body?.trim();
    if (!body) {
        return false;
    }
    try {
        JSON.parse(body);
        return false;
    } catch {
        return true;
    }
}

export function validateRateLimitDraft(draft: RateLimitDraftLike): RateLimitValidationError[] {
    const errors: RateLimitValidationError[] = [];
    if (!/^[a-z][a-z0-9-]*$/.test(draft.name || '')) {
        errors.push({ field: 'name', message: '规则名称必须为 kebab-case' });
    }

    const rules = draft.rules || [];
    if (rules.length === 0) {
        errors.push({ field: 'rules', message: '至少配置 1 条限流子规则' });
    }

    rules.forEach((rule, index) => {
        const ruleNumber = index + 1;
        if (!rule.method?.value?.trim()) {
            errors.push({ field: `rules.${index}.method`, ruleIndex: index, message: `规则[${ruleNumber}] 存在空的接口路径` });
        }
        if (rule.resource === 'CONCURRENCY') {
            const maxConcurrent = Number(rule.concurrencyAmount?.maxAmount ?? rule.amounts?.[0]?.maxAmount ?? 0);
            if (maxConcurrent <= 0) {
                errors.push({ field: `rules.${index}.concurrencyAmount`, ruleIndex: index, message: `规则[${ruleNumber}] 最大并发数必须大于 0` });
            }
        } else {
            const amounts = rule.amounts || [];
            if (amounts.length === 0) {
                errors.push({ field: `rules.${index}.amounts`, ruleIndex: index, message: `规则[${ruleNumber}] 至少配置 1 个限流窗口` });
            }
            amounts.forEach((amount) => {
                if (Number(amount.maxAmount || 0) <= 0) {
                    errors.push({ field: `rules.${index}.amounts.maxAmount`, ruleIndex: index, message: `规则[${ruleNumber}] 最大请求数必须大于 0` });
                }
            });
        }
        if ((rule.arguments || []).some((arg) => !arg.value?.value?.trim())) {
            errors.push({ field: `rules.${index}.arguments.value`, ruleIndex: index, message: `规则[${ruleNumber}] 存在空匹配值` });
        }
        if (hasInvalidCustomResponse(rule)) {
            errors.push({ field: `rules.${index}.customResponse`, ruleIndex: index, message: `规则[${ruleNumber}] 自定义响应必须是合法 JSON` });
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

export function stringifyRateLimitSpec(spec: RateLimitSubmitPayload, format: RateLimitSpecFormat): string {
    if (format === 'json') {
        return JSON.stringify(spec, null, 2);
    }
    return stringifyYAMLValue(spec);
}
