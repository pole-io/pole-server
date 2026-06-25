export type LaneSpecFormat = 'yaml' | 'json';

export const LANE_TRAFFIC_TAG_KEY = 'X-Lattice-Traffic-Lane';

export interface LaneDraftEntry {
    kind: 'gateway' | 'app';
    ns: string;
    svc: string;
}

export interface LaneDraftCondition {
    type: string;
    key: string;
    match: string;
    value: string;
}

export interface LaneDraftServiceTag {
    service: string;
}

export interface LaneDraftRule {
    id?: string;
    name: string;
    on: boolean;
    open: boolean;
    laneValue: string;
    relation: 'AND' | 'OR';
    matchRatio: number;
    conditions: LaneDraftCondition[];
    serviceTags: LaneDraftServiceTag[];
    note?: string;
}

export interface LaneGroupDraft {
    id?: string;
    name: string;
    enabled: boolean;
    priority: number;
    desc: string;
    tags: { key: string; value: string }[];
    laneGroup: {
        name: string;
        status: string;
    };
    entries: LaneDraftEntry[];
    selected: string[];
    lanes: LaneDraftRule[];
}

export interface ValidationIssue {
    field: string;
    message: string;
}

export interface LaneTopology {
    entryLabel: string;
    matchLabel: string;
    laneLabel: string;
    hitRatio: number;
    fallbackRatio: number;
    fallbackLabel: string;
    visibleServices: string[];
    omittedServiceCount: number;
}

const DEFAULT_CONDITION: LaneDraftCondition = {
    type: 'HEADER',
    key: '',
    match: '完全匹配',
    value: '',
};

const clampRatio = (value: unknown) => {
    const numberValue = Number(value);
    if (!Number.isFinite(numberValue)) return 100;
    return Math.min(100, Math.max(0, Math.round(numberValue)));
};

const rawRatio = (value: unknown) => {
    const numberValue = Number(value);
    if (!Number.isFinite(numberValue)) return 100;
    return Math.round(numberValue);
};

const normalizeRelation = (value: unknown): 'AND' | 'OR' => (value === 'OR' ? 'OR' : 'AND');

export function parseServiceName(value?: string) {
    const text = value || '';
    const bracketMatch = text.match(/^(.+?)\s+\((.+)\)$/);
    if (bracketMatch) {
        return {
            service: bracketMatch[1],
            namespace: bracketMatch[2],
        };
    }
    const slashParts = text.split('/');
    if (slashParts.length >= 2) {
        return {
            namespace: slashParts[0],
            service: slashParts[1],
        };
    }
    return {
        namespace: '',
        service: text,
    };
}

function normalizeCondition(condition?: Partial<LaneDraftCondition> | any): LaneDraftCondition {
    const value = condition?.value;
    return {
        type: condition?.type || condition?.param || DEFAULT_CONDITION.type,
        key: condition?.key || '',
        match: condition?.match || condition?.op || DEFAULT_CONDITION.match,
        value: typeof value === 'object' ? value?.value || '' : value || '',
    };
}

function normalizeServiceTag(tag?: Partial<LaneDraftServiceTag> | any, fallbackService = ''): LaneDraftServiceTag {
    return {
        service: tag?.service || tag?.name || fallbackService,
    };
}

function normalizeLaneRule(rule?: Partial<LaneDraftRule> | any, groupServices: string[] = []): LaneDraftRule {
    const matchRule = rule?.trafficMatchRule || rule?.traffic_match_rule || {};
    const conditions = rule?.conditions || matchRule.arguments || [];
    const serviceTags = rule?.serviceTags || rule?.services || [];
    const fallbackService = groupServices[0] || '';
    return {
        id: rule?.id,
        name: rule?.name || '',
        on: rule?.on ?? rule?.enable ?? rule?.enabled ?? true,
        open: rule?.open ?? false,
        laneValue: rule?.laneValue ?? rule?.defaultLabelValue ?? rule?.default_label_value ?? '',
        relation: normalizeRelation(rule?.relation ?? matchRule.matchMode ?? matchRule.match_mode),
        matchRatio: rawRatio(rule?.matchRatio ?? rule?.ratioPercent ?? rule?.ratio_percent ?? 100),
        conditions: (conditions.length ? conditions : [DEFAULT_CONDITION]).map(normalizeCondition),
        serviceTags: (serviceTags.length ? serviceTags : [{ service: fallbackService }]).map((item: any) => normalizeServiceTag(item, fallbackService)),
        note: rule?.note || rule?.description || '',
    };
}

function entryFromRaw(entry: any): LaneDraftEntry {
    if (entry?.kind) {
        return {
            kind: entry.kind === 'gateway' ? 'gateway' : 'app',
            ns: entry.ns || entry.namespace || '',
            svc: entry.svc || entry.service || '',
        };
    }
    const selector = entry?.selector || {};
    const type = entry?.type || '';
    const selectorType = selector?.['@type'] || '';
    const isGateway = type === 'gateway' || selectorType.includes('ServiceGatewaySelector');
    return {
        kind: isGateway ? 'gateway' : 'app',
        ns: selector.namespace || '',
        svc: selector.service || '',
    };
}

function serviceFromDestination(destination: any): string {
    if (typeof destination === 'string') return parseServiceName(destination).service;
    return destination?.service || destination?.name || '';
}

export function normalizeLaneGroupDraft(input: Partial<LaneGroupDraft> | any = {}): LaneGroupDraft {
    const selected = (input.selected || input.destinations || []).map(serviceFromDestination).filter(Boolean);
    const entries = (input.entries || []).map(entryFromRaw);
    const tags = Array.isArray(input.tags)
        ? input.tags
        : Object.entries(input.metadata || {}).map(([key, value]) => ({ key, value: String(value) }));
    const name = input.name || input.laneGroup?.name || '';
    const lanes = (input.lanes || input.rules || []).map((rule: any) => normalizeLaneRule(rule, selected));
    return {
        id: input.id,
        name,
        enabled: input.enabled ?? input.enable ?? true,
        priority: Number(input.priority ?? 5),
        desc: input.desc ?? input.description ?? '',
        tags,
        laneGroup: {
            name,
            status: input.laneGroup?.status || input.status || 'CREATED',
        },
        entries,
        selected,
        lanes: lanes.length ? lanes : [normalizeLaneRule(undefined, selected)],
    };
}

export function validateLaneGroupDraft(draft: LaneGroupDraft): ValidationIssue[] {
    const issues: ValidationIssue[] = [];
    if (!draft.name.trim()) {
        issues.push({ field: 'name', message: '泳道组名称不能为空' });
    }
    if (!draft.entries.length) {
        issues.push({ field: 'entries', message: '至少配置 1 个泳道组入口' });
    }
    if (draft.entries.some(entry => !entry.ns || !entry.svc)) {
        issues.push({ field: 'entries', message: '泳道组入口命名空间 / 服务不能为空' });
    }
    if (!draft.selected.length) {
        issues.push({ field: 'selected', message: '至少纳入 1 个泳道组服务' });
    }
    return issues;
}

export function validateLaneRulesDraft(draft: LaneGroupDraft): ValidationIssue[] {
    const issues: ValidationIssue[] = [];
    const groupServices = new Set(draft.selected);
    if (!draft.lanes.length) {
        issues.push({ field: 'lanes', message: '至少配置 1 条泳道' });
        return issues;
    }
    draft.lanes.forEach((lane, index) => {
        const prefix = `泳道[${index + 1}]`;
        if (!lane.name.trim()) {
            issues.push({ field: `lanes.${index}.name`, message: `${prefix} 名称不能为空` });
        }
        if (lane.matchRatio < 0 || lane.matchRatio > 100) {
            issues.push({ field: `lanes.${index}.matchRatio`, message: `${prefix} 命中后放量比例必须在 0–100 之间` });
        }
        if (!lane.conditions.length || lane.conditions.some(item => !item.type || !item.key || !item.match || !item.value)) {
            issues.push({ field: `lanes.${index}.conditions`, message: `${prefix} 存在空匹配条件` });
        }
        if (!lane.laneValue.trim()) {
            issues.push({ field: `lanes.${index}.laneValue`, message: `${prefix} 泳道标签 Value 不能为空` });
        }
        if (!lane.serviceTags.length || lane.serviceTags.some(item => !item.service)) {
            issues.push({ field: `lanes.${index}.serviceTags`, message: `${prefix} 至少选择 1 个组内服务` });
        } else if (lane.serviceTags.some(item => !groupServices.has(item.service))) {
            issues.push({ field: `lanes.${index}.serviceTags`, message: `${prefix} 引用了泳道组之外的服务` });
        }
    });
    return issues;
}

export function laneRuleSummary(lane: LaneDraftRule) {
    const ratio = clampRatio(lane.matchRatio);
    return `${lane.conditions.length} 个匹配规则 · 放量 ${ratio}% · ${lane.serviceTags.length} 个组内服务 · ${LANE_TRAFFIC_TAG_KEY}=${lane.laneValue || '-'}`;
}

export function buildLaneRulePreviewSpec(draft: LaneGroupDraft) {
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'LaneRule',
        metadata: {
            name: draft.name,
            enabled: draft.enabled,
            priority: draft.priority,
        },
        spec: {
            description: draft.desc,
            laneGroup: {
                name: draft.laneGroup.name || draft.name,
                status: draft.laneGroup.status,
            },
            lanes: draft.lanes.map(lane => ({
                name: lane.name,
                enabled: lane.on,
                note: lane.note,
                match: {
                    relation: lane.relation,
                    ratioPercent: clampRatio(lane.matchRatio),
                    conditions: lane.conditions.map(condition => ({
                        param: condition.type,
                        key: condition.key,
                        op: condition.match,
                        value: condition.value,
                    })),
                },
                services: lane.serviceTags.map(item => item.service).filter(Boolean),
                laneTag: {
                    key: LANE_TRAFFIC_TAG_KEY,
                    value: lane.laneValue,
                },
            })),
        },
    };
}

export function buildLaneGroupPreviewSpec(draft: LaneGroupDraft) {
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'LaneGroup',
        metadata: {
            name: draft.name,
        },
        spec: {
            description: draft.desc,
            laneGroup: {
                name: draft.laneGroup.name || draft.name,
                status: draft.laneGroup.status,
                entries: draft.entries.map(entry => ({
                    type: entry.kind === 'gateway' ? 'micro-gateway' : 'micro-app',
                    namespace: entry.ns,
                    service: entry.svc,
                })),
                services: draft.selected,
            },
        },
    };
}

export function buildLaneVersionPreviewSpec(draft: LaneGroupDraft) {
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'LaneGroupVersion',
        metadata: {
            name: draft.name,
        },
        spec: {
            latestSnapshot: buildLaneGroupPreviewSpec(draft).spec,
        },
    };
}

export function buildLaneAuditPreviewSpec(draft: LaneGroupDraft) {
    return {
        apiVersion: 'governance.pole.io/v1',
        kind: 'LaneGroupAudit',
        metadata: {
            name: draft.name,
        },
        spec: {
            events: [],
        },
    };
}

export function buildLaneTopology(draft: LaneGroupDraft, lane: LaneDraftRule): LaneTopology {
    const ratio = clampRatio(lane.matchRatio);
    const services = lane.serviceTags.map(item => item.service).filter(Boolean);
    let visibleServices = services;
    let omittedServiceCount = 0;
    if (services.length > 3) {
        omittedServiceCount = services.length - 3;
        visibleServices = [services[0], services[1], services[services.length - 1]];
    }
    const firstEntry = draft.entries[0];
    return {
        entryLabel: firstEntry?.svc || '泳道组入口',
        matchLabel: `${lane.relation} · ${lane.conditions.length} 条条件`,
        laneLabel: lane.name || '未命名泳道',
        hitRatio: ratio,
        fallbackRatio: 100 - ratio,
        fallbackLabel: ratio === 100 ? '未命中' : '回落',
        visibleServices,
        omittedServiceCount,
    };
}

function quoteYAMLValue(value: unknown): string {
    if (value === null || value === undefined) return '';
    if (typeof value === 'number' || typeof value === 'boolean') return String(value);
    const text = String(value);
    if (!text || /^(true|false|null|yes|no|on|off)$/i.test(text) || /[:{}\[\],&*#?|\-<>=!%@`"'\n]/.test(text)) {
        return JSON.stringify(text);
    }
    return text;
}

function stringifyYAMLValue(value: unknown, indent = 0): string {
    const pad = ' '.repeat(indent);
    if (Array.isArray(value)) {
        if (!value.length) return `${pad}[]`;
        return value.map((item) => {
            if (typeof item === 'object' && item !== null) {
                return `${pad}- ${stringifyYAMLValue(item, indent + 2).trimStart()}`;
            }
            return `${pad}- ${quoteYAMLValue(item)}`;
        }).join('\n');
    }
    if (typeof value === 'object' && value !== null) {
        return Object.entries(value as Record<string, unknown>).map(([key, item]) => {
            if (Array.isArray(item) && item.length === 0) {
                return `${pad}${key}: []`;
            }
            if (typeof item === 'object' && item !== null) {
                return `${pad}${key}:\n${stringifyYAMLValue(item, indent + 2)}`;
            }
            return `${pad}${key}: ${quoteYAMLValue(item)}`;
        }).join('\n');
    }
    return `${pad}${quoteYAMLValue(value)}`;
}

export function stringifyLaneSpec(spec: unknown, format: LaneSpecFormat): string {
    if (format === 'json') {
        return JSON.stringify(spec, null, 2);
    }
    return stringifyYAMLValue(spec);
}
