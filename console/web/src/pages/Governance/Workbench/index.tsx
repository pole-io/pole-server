import React from 'react';
import { Button, Dialog, Input, Link, Popconfirm, Select, Space, Steps, Table, Tag, Tooltip } from 'tdesign-react';
import type { PrimaryTableProps, TableRowData } from 'tdesign-react';
import { AddIcon, CreditcardIcon, DeleteIcon, FilterClearIcon, RefreshIcon } from 'tdesign-icons-react';

import { useAppDispatch } from 'modules/store';
import RuleDetailDrawer, { WIDE_RULE_DETAIL_DRAWER_SIZE } from '../RuleRelease/RuleDetailDrawer';
import RuleTabs from '../RuleRelease/RuleTabs';
import SubscribeTable from 'components/SubscribeTable';
import CustomRouteEditor from '../Router/CustomRouteEditor';
import RateLimitEditor from '../RateLimit/RateLimitEditor';
import CircuitBreakerEditor from '../CircuitBreaker/CircuitBreakerEditor';
import FaultDetectEditor from '../CircuitBreaker/FaultDetectEditor';
import LossLessEditor from '../LossLess/LossLessEditor';
import LaneGroupEdtor from '../Router/LaneGroupEdtor';
import TrafficGovernanceEditor, { trafficRuleCount, trafficRuleSummary } from '../Security/TrafficGovernanceEditor';
import {
    editorCustomRoute,
    listCustomRouteVersions,
    listCustomRoutes,
    removeCustomRoutes,
    resetCustomRoute,
} from 'modules/governance/route';
import {
    editorRateLimitRule,
    listRateLimitRuleVersions,
    listRateLimitRules,
    removeRateLimitRule,
    removeRateLimitRuleVersion,
    resetRateLimitRule,
    rollbackRateLimitRuleVersion,
} from 'modules/governance/ratelimit';
import {
    editorCircuitBreaker,
    listCircuitBreakerVersions,
    listCircuitBreakers,
    removeCircuitBreakers,
    removeCircuitBreakerRelease,
    resetCircuitBreaker,
    rollbackCircuitBreakerRelease,
} from 'modules/governance/circuitbreaker';
import {
    editorFaultDetect,
    listFaultDetectVersions,
    listFaultDetects,
    removeFaultDetects,
    removeFaultDetectVersion,
    resetFaultDetect,
    rollbackFaultDetectVersion,
} from 'modules/governance/faultdetect';
import {
    editorLosslessRule,
    listLossLessRules,
    listLosslessRuleVersions,
    removeLosslessRule,
    removeLosslessVersion,
    resetLosslessRule,
    rollbackLosslessVersion,
} from 'modules/governance/lossless';
import {
    editorLaneGroup,
    listLaneGroupVersions,
    listLaneGroups,
    removeLaneGroups,
    removeLaneGroupVersion,
    resetLaneGroup,
    rollbackLanGroupVersion,
} from 'modules/governance/lane_group';
import { LimitActionMap, LimitType, LimitTypeMap, RateLimitView } from 'services/ratelimit';
import { CustomRouteView, normalizeRoutingConfigForEditor } from 'services/router';
import { CircuitBreakerRule } from 'services/circuitbreaker';
import { FaultDetectRule } from 'services/faultdetect';
import { LossLessRuleView } from 'services/lossless';
import { LaneGroupView } from 'services/lane';
import { Op, RuleRelease } from 'services/types';
import {
    deleteTrafficGovernanceRelease,
    deleteTrafficGovernanceRules,
    describeTrafficGovernanceRules,
    describeTrafficGovernanceVersions,
    TrafficGovernanceKind,
    TrafficGovernanceKindLabel,
    TrafficGovernanceRule,
} from 'services/traffic_governance';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';

const { StepItem } = Steps;

type RuleKind = 'route' | 'ratelimit-local' | 'ratelimit-global' | 'circuitbreaker' | 'faultdetect' | 'lossless' | 'lane' | 'traffic-security' | 'traffic-mirror' | 'traffic-mock';

interface GovernanceRuleRow extends TableRowData {
    key: string;
    kind: RuleKind;
    typeLabel: string;
    name: string;
    description?: string;
    namespace?: string;
    service?: string;
    target: string;
    condition: string;
    status: string;
    release: string;
    ctime?: string;
    mtime?: string;
    editable?: boolean;
    deleteable?: boolean;
    raw: TableRowData;
    limitType?: LimitType;
    trafficKind?: TrafficGovernanceKind;
}

const typeOptions = [
    { label: '路由', value: 'route', scene: '按主调、接口或标签把流量分配到不同目标服务或版本。' },
    { label: '限流', value: 'ratelimit', scene: '接口流量需要按 QPS、并发或条件限制，避免服务被打满。' },
    { label: '熔断', value: 'circuitbreaker', scene: '下游错误率、慢调用或异常升高时，快速切断不稳定依赖。' },
    { label: '探测', value: 'faultdetect', scene: '需要主动探测接口或服务健康，并据此辅助治理决策。' },
    { label: '无损', value: 'lossless', scene: '发布、重启或下线前需要延迟摘除流量，保护存量请求。' },
    { label: '泳道', value: 'lane', scene: '灰度、联调或多环境隔离时，把命中流量导入指定泳道服务。' },
    { label: '鉴权', value: 'traffic-security', scene: '需要按调用方、接口或请求条件控制访问许可。' },
    { label: '镜像', value: 'traffic-mirror', scene: '需要把线上流量复制到影子服务，用于回放、验证或压测。' },
    { label: 'Mock', value: 'traffic-mock', scene: '依赖未就绪或需要固定响应时，按接口返回模拟结果。' },
];

const createRuleTargets: Record<string, { kind: RuleKind; label: string; limitType?: LimitType; trafficKind?: TrafficGovernanceKind }> = {
    route: { kind: 'route', label: '路由' },
    ratelimit: { kind: 'ratelimit-local', label: '限流', limitType: LimitType.LOCAL },
    circuitbreaker: { kind: 'circuitbreaker', label: '熔断' },
    faultdetect: { kind: 'faultdetect', label: '探测' },
    lossless: { kind: 'lossless', label: '无损' },
    lane: { kind: 'lane', label: '泳道' },
    'traffic-security': { kind: 'traffic-security', label: '鉴权', trafficKind: 'security' },
    'traffic-mirror': { kind: 'traffic-mirror', label: '镜像', trafficKind: 'mirror' },
    'traffic-mock': { kind: 'traffic-mock', label: 'Mock', trafficKind: 'mock' },
};

const getActionPayload = <T,>(action: unknown): T | undefined => {
    const result = action as { meta?: { requestStatus?: string }, payload?: T };
    if (result.meta?.requestStatus === 'fulfilled') return result.payload;
    return undefined;
};

const text = (value?: string | number | boolean) => {
    if (value === undefined || value === null || value === '') return '-';
    return String(value);
};

const formatTime = (value?: string) => {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    const pad = (num: number) => `${num}`.padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const getTypeClassName = (kind: RuleKind) => {
    if (kind === 'route') return style.typeRoute;
    if (kind === 'ratelimit-local' || kind === 'ratelimit-global') return style.typeRatelimit;
    if (kind === 'circuitbreaker') return style.typeCircuitbreaker;
    if (kind === 'faultdetect') return style.typeFaultdetect;
    if (kind === 'lossless') return style.typeLossless;
    if (kind === 'lane') return style.typeLane;
    if (kind === 'traffic-security') return style.typeSecurity;
    if (kind === 'traffic-mirror') return style.typeMirror;
    if (kind === 'traffic-mock') return style.typeMock;
    return '';
};

const routeTarget = (rule: CustomRouteView) => {
    const routingConfig = normalizeRoutingConfigForEditor(rule.routing_config);
    const firstRule = routingConfig?.rules?.[0];
    const source = firstRule?.sources?.[0];
    const destination = firstRule?.destinations?.[0];
    return {
        namespace: routingConfig?.callee?.namespace || destination?.namespace || source?.namespace || '',
        service: routingConfig?.callee?.service || destination?.service || source?.service || '',
        target: `${text(routingConfig?.caller?.namespace || source?.namespace)}/${text(routingConfig?.caller?.service || source?.service)} -> ${text(routingConfig?.callee?.namespace || destination?.namespace)}/${text(routingConfig?.callee?.service || destination?.service)}`,
        condition: firstRule?.name || firstRule?.arguments?.arguments?.map((arg) => `${arg.type}:${arg.key}`).join(' / ') || '按主调 / 被调服务匹配',
    };
};

const normalizeRules = (items: {
    routes: CustomRouteView[];
    localRateLimits: RateLimitView[];
    globalRateLimits: RateLimitView[];
    circuitBreakers: CircuitBreakerRule[];
    faultDetects: FaultDetectRule[];
    losslessRules: LossLessRuleView[];
    laneGroups: LaneGroupView[];
    trafficSecurityRules: TrafficGovernanceRule[];
    trafficMirrorRules: TrafficGovernanceRule[];
    trafficMockRules: TrafficGovernanceRule[];
}): GovernanceRuleRow[] => {
    const routeRows = items.routes.map((rule) => {
        const target = routeTarget(rule);
        return {
            key: `route-${rule.id || rule.name}`,
            kind: 'route' as RuleKind,
            typeLabel: '路由',
            name: rule.name,
            description: rule.description,
            namespace: target.namespace,
            service: target.service,
            target: target.target,
            condition: target.condition,
            status: rule.enable === false ? '禁用' : '启用',
            release: rule.etime ? '已发布' : '待发布',
            ctime: rule.ctime,
            mtime: rule.mtime,
            editable: rule.editable,
            deleteable: rule.deleteable,
            raw: rule as TableRowData,
        };
    });

    const rateRows = [...items.localRateLimits, ...items.globalRateLimits].map((rule) => {
        const firstRule = rule.rules?.[0];
        return {
            key: `ratelimit-${rule.type}-${rule.id || rule.name}`,
            kind: rule.type === LimitType.GLOBAL ? 'ratelimit-global' as RuleKind : 'ratelimit-local' as RuleKind,
            typeLabel: LimitTypeMap[rule.type] || '限流',
            name: rule.name,
            description: rule.description,
            namespace: rule.namespace,
            service: rule.service,
            target: `${text(rule.namespace)}/${text(rule.service)}`,
            condition: firstRule ? `${text(firstRule.apis?.[0]?.path?.value || firstRule.method?.value || firstRule.resource)} / ${LimitActionMap[firstRule.action] || firstRule.action || '-'}` : '按接口或资源限流',
            status: rule.disable ? '禁用' : '启用',
            release: rule.revision ? '已发布' : '待发布',
            ctime: rule.ctime,
            mtime: rule.mtime,
            editable: rule.editable,
            deleteable: rule.deleteable,
            raw: rule as TableRowData,
            limitType: rule.type,
        };
    });

    const circuitRows = items.circuitBreakers.map((rule) => ({
        key: `circuitbreaker-${rule.id || rule.name}`,
        kind: 'circuitbreaker' as RuleKind,
        typeLabel: '熔断',
        name: rule.name,
        description: rule.description,
        namespace: rule.ruleMatcher?.destination?.namespace,
        service: rule.ruleMatcher?.destination?.service,
        target: `${text(rule.ruleMatcher?.source?.namespace)}/${text(rule.ruleMatcher?.source?.service)} -> ${text(rule.ruleMatcher?.destination?.namespace)}/${text(rule.ruleMatcher?.destination?.service)}`,
        condition: rule.block_configs?.[0]?.name || '按错误条件和触发条件熔断',
        status: '启用',
        release: rule.etime ? '已发布' : '待发布',
        ctime: rule.ctime,
        mtime: rule.mtime,
        editable: rule.editable,
        deleteable: rule.deleteable,
        raw: rule as TableRowData,
    }));

    const faultRows = items.faultDetects.map((rule) => ({
        key: `faultdetect-${rule.id || rule.name}`,
        kind: 'faultdetect' as RuleKind,
        typeLabel: '探测',
        name: rule.name,
        description: rule.description,
        namespace: rule.targetService?.namespace,
        service: rule.targetService?.service,
        target: `${text(rule.targetService?.namespace)}/${text(rule.targetService?.service)}`,
        condition: `${text(rule.protocol)} / ${text(rule.targetService?.api?.path?.value || rule.httpConfig?.url)}`,
        status: '启用',
        release: rule.mtime ? '已发布' : '待发布',
        ctime: rule.ctime,
        mtime: rule.mtime,
        editable: rule.editable,
        deleteable: rule.deleteable,
        raw: rule as TableRowData,
    }));

    const losslessRows = items.losslessRules.map((rule) => ({
        key: `lossless-${rule.id || `${rule.namespace}-${rule.service}`}`,
        kind: 'lossless' as RuleKind,
        typeLabel: '无损',
        name: `${rule.namespace}/${rule.service}`,
        description: '无损上线、服务预热与无损下线',
        namespace: rule.namespace,
        service: rule.service,
        target: `${text(rule.namespace)}/${text(rule.service)}`,
        condition: [
            rule.lossless_online?.delay_register?.enable ? '延迟注册' : '',
            rule.lossless_online?.warmup?.enable ? '服务预热' : '',
            rule.lossless_offline?.enable ? '无损下线' : '',
        ].filter(Boolean).join(' / ') || '无损上下线',
        status: '启用',
        release: rule.revision ? '已发布' : '待发布',
        ctime: rule.ctime,
        mtime: rule.mtime,
        editable: rule.editable,
        deleteable: rule.deleteable,
        raw: rule as TableRowData,
    }));

    const laneRows = items.laneGroups.map((rule) => ({
        key: `lane-${rule.id || rule.name}`,
        kind: 'lane' as RuleKind,
        typeLabel: '泳道',
        name: rule.name || '-',
        description: rule.description,
        namespace: rule.destinations?.[0]?.namespace,
        service: rule.destinations?.[0]?.service,
        target: `${rule.rules?.length || 0} 个泳道 / ${rule.destinations?.length || 0} 个目标服务`,
        condition: rule.entries?.map((entry) => `${text(entry.selector?.namespace)}/${text(entry.selector?.service)}`).join(', ') || '按入口服务匹配',
        status: '启用',
        release: rule.mtime ? '已发布' : '待发布',
        ctime: rule.ctime,
        mtime: rule.mtime,
        editable: rule.editable,
        deleteable: rule.deleteable,
        raw: rule as TableRowData,
    }));

    const trafficRows = ([
        ['traffic-security', 'security', items.trafficSecurityRules],
        ['traffic-mirror', 'mirror', items.trafficMirrorRules],
        ['traffic-mock', 'mock', items.trafficMockRules],
    ] as Array<[RuleKind, TrafficGovernanceKind, TrafficGovernanceRule[]]>).flatMap(([rowKind, trafficKind, rules]) => (
        rules.map((rule) => ({
            key: `${rowKind}-${rule.id || rule.name}`,
            kind: rowKind,
            typeLabel: TrafficGovernanceKindLabel[trafficKind],
            name: rule.name,
            description: rule.description,
            namespace: rule.namespace,
            service: rule.service,
            target: `${text(rule.namespace)}/${text(rule.service)}`,
            condition: `${trafficRuleCount(trafficKind, rule)} 条 / ${trafficRuleSummary(trafficKind, rule)}`,
            status: rule.enable ? '启用' : '禁用',
            release: rule.revision ? '已发布' : '待发布',
            ctime: rule.ctime,
            mtime: rule.mtime,
            editable: rule.editable,
            deleteable: rule.deleteable,
            raw: rule as TableRowData,
            trafficKind,
        }))
    ));

    return [...routeRows, ...rateRows, ...circuitRows, ...faultRows, ...losslessRows, ...laneRows, ...trafficRows];
};

const GovernanceWorkbench: React.FC = () => {
    const dispatch = useAppDispatch();
    const [rules, setRules] = React.useState<GovernanceRuleRow[]>([]);
    const [loading, setLoading] = React.useState(false);
    const [search, setSearch] = React.useState('');
    const [typeFilters, setTypeFilters] = React.useState<string[]>([]);
    const [selected, setSelected] = React.useState<GovernanceRuleRow | null>(null);
    const [drawerVisible, setDrawerVisible] = React.useState(false);
    const [drawerMode, setDrawerMode] = React.useState<Op>('view');
    const [createWizardVisible, setCreateWizardVisible] = React.useState(false);
    const [createWizardStep, setCreateWizardStep] = React.useState(1);
    const [pendingCreateType, setPendingCreateType] = React.useState('');
    const [versions, setVersions] = React.useState<RuleRelease[]>([]);
    const [versionLoading, setVersionLoading] = React.useState(false);
    const [versionTotal, setVersionTotal] = React.useState(0);
    const [versionPage, setVersionPage] = React.useState(1);
    const [versionLimit, setVersionLimit] = React.useState(10);

    const refreshData = React.useCallback(async (query = search) => {
        setLoading(true);
        try {
            const [
                routeAction,
                localRateAction,
                globalRateAction,
                circuitAction,
                faultAction,
                losslessAction,
                laneAction,
                trafficSecurityAction,
                trafficMirrorAction,
                trafficMockAction,
            ] = await Promise.all([
                dispatch(listCustomRoutes({ param: { offset: 0, limit: 20, route_type: 'RulePolicy', name: query } })),
                dispatch(listRateLimitRules({ param: { offset: 0, limit: 20, name: query, limit_type: LimitType.LOCAL } })),
                dispatch(listRateLimitRules({ param: { offset: 0, limit: 20, name: query, limit_type: LimitType.GLOBAL } })),
                dispatch(listCircuitBreakers({ param: { offset: 0, limit: 20, name: query, brief: true } })),
                dispatch(listFaultDetects({ param: { offset: 0, limit: 20, name: query, brief: true } })),
                dispatch(listLossLessRules({ param: { offset: 0, limit: 20, name: query } })),
                dispatch(listLaneGroups({ param: { offset: 0, limit: 20, name: query, brief: true } })),
                describeTrafficGovernanceRules('security', { offset: 0, limit: 20, name: query }),
                describeTrafficGovernanceRules('mirror', { offset: 0, limit: 20, name: query }),
                describeTrafficGovernanceRules('mock', { offset: 0, limit: 20, name: query }),
            ]);

            const routePayload = getActionPayload<{ datas: CustomRouteView[] }>(routeAction);
            const localRatePayload = getActionPayload<{ datas: RateLimitView[] }>(localRateAction);
            const globalRatePayload = getActionPayload<{ datas: RateLimitView[] }>(globalRateAction);
            const circuitPayload = getActionPayload<{ datas: CircuitBreakerRule[] }>(circuitAction);
            const faultPayload = getActionPayload<{ datas: FaultDetectRule[] }>(faultAction);
            const losslessPayload = getActionPayload<{ datas: LossLessRuleView[] }>(losslessAction);
            const lanePayload = getActionPayload<{ datas: LaneGroupView[] }>(laneAction);
            const trafficSecurityPayload = trafficSecurityAction as { list: TrafficGovernanceRule[] };
            const trafficMirrorPayload = trafficMirrorAction as { list: TrafficGovernanceRule[] };
            const trafficMockPayload = trafficMockAction as { list: TrafficGovernanceRule[] };

            setRules(normalizeRules({
                routes: routePayload?.datas || [],
                localRateLimits: localRatePayload?.datas || [],
                globalRateLimits: globalRatePayload?.datas || [],
                circuitBreakers: circuitPayload?.datas || [],
                faultDetects: faultPayload?.datas || [],
                losslessRules: losslessPayload?.datas || [],
                laneGroups: lanePayload?.datas || [],
                trafficSecurityRules: trafficSecurityPayload.list || [],
                trafficMirrorRules: trafficMirrorPayload.list || [],
                trafficMockRules: trafficMockPayload.list || [],
            }));
        } catch (error) {
            openErrNotification('请求失败', `查询治理规则列表错误: ${(error as Error).message}`);
        } finally {
            setLoading(false);
        }
    }, [dispatch, search]);

    React.useEffect(() => {
        refreshData('');
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const openRule = (rule: GovernanceRuleRow, mode: Op = 'view') => {
        setSelected(rule);
        setDrawerMode(mode);
        setVersions([]);
        setVersionTotal(0);
        setVersionPage(1);
        setVersionLimit(10);
        switch (rule.kind) {
            case 'route':
                dispatch(editorCustomRoute(rule.raw as CustomRouteView));
                break;
            case 'ratelimit-local':
            case 'ratelimit-global':
                dispatch(editorRateLimitRule(rule.raw as RateLimitView));
                break;
            case 'circuitbreaker':
                dispatch(editorCircuitBreaker(rule.raw as CircuitBreakerRule));
                break;
            case 'faultdetect':
                dispatch(editorFaultDetect(rule.raw as FaultDetectRule));
                break;
            case 'lossless':
                dispatch(editorLosslessRule(rule.raw as LossLessRuleView));
                break;
            case 'lane':
                dispatch(editorLaneGroup(rule.raw as LaneGroupView));
                break;
            case 'traffic-security':
            case 'traffic-mirror':
            case 'traffic-mock':
                break;
        }
        setDrawerVisible(true);
    };

    const deleteRule = async (rule: GovernanceRuleRow) => {
        const id = rule.raw.id as string | undefined;
        if (!id) {
            openErrNotification('删除失败', '当前规则缺少可删除的 ID');
            return;
        }
        try {
            if (rule.kind === 'route') {
                await dispatch(removeCustomRoutes({ ids: [id] }));
            } else if (rule.kind === 'ratelimit-local' || rule.kind === 'ratelimit-global') {
                await dispatch(removeRateLimitRule({ ids: [id] }));
            } else if (rule.kind === 'circuitbreaker') {
                await dispatch(removeCircuitBreakers({ ids: [id] }));
            } else if (rule.kind === 'faultdetect') {
                await dispatch(removeFaultDetects({ ids: [id] }));
            } else if (rule.kind === 'lossless') {
                await dispatch(removeLosslessRule({ ids: [id] }));
            } else if (rule.kind === 'lane') {
                await dispatch(removeLaneGroups({ ids: [id] }));
            } else if (rule.trafficKind) {
                await deleteTrafficGovernanceRules(rule.trafficKind, [{ id }]);
            }
            openInfoNotification('请求成功', '删除治理规则成功');
            refreshData(search);
        } catch (error) {
            openErrNotification('删除失败', `删除治理规则错误: ${(error as Error).message}`);
        }
    };

    const openCreateRule = (type: string) => {
        const target = createRuleTargets[type];
        if (!target) return;
        setSelected({
            key: `${target.kind}-create`,
            kind: target.kind,
            typeLabel: target.label,
            name: `新建${target.label}规则`,
            description: '',
            namespace: '',
            service: '',
            target: '-',
            condition: '-',
            status: '启用',
            release: '待发布',
            raw: { id: '' } as TableRowData,
            limitType: target.limitType,
            trafficKind: target.trafficKind,
            editable: true,
            deleteable: true,
        });
        setDrawerMode('create');
        setVersions([]);
        setVersionTotal(0);
        setVersionPage(1);
        setVersionLimit(10);
        switch (target.kind) {
            case 'route':
                dispatch(resetCustomRoute());
                break;
            case 'ratelimit-local':
            case 'ratelimit-global':
                dispatch(resetRateLimitRule());
                break;
            case 'circuitbreaker':
                dispatch(resetCircuitBreaker());
                break;
            case 'faultdetect':
                dispatch(resetFaultDetect());
                break;
            case 'lossless':
                dispatch(resetLosslessRule());
                break;
            case 'lane':
                dispatch(resetLaneGroup());
                break;
            default:
                break;
        }
        setDrawerVisible(true);
    };

    const openCreateWizard = () => {
        setPendingCreateType('');
        setCreateWizardStep(1);
        setCreateWizardVisible(true);
    };

    const confirmCreateWizard = () => {
        if (!pendingCreateType) {
            openErrNotification('无法创建', '请先选择规则类型');
            return;
        }
        setCreateWizardStep(2);
        setCreateWizardVisible(false);
        openCreateRule(pendingCreateType);
    };

    const refreshVersions = async (page = 1, limit = 10) => {
        if (!selected) return;
        setVersionLoading(true);
        try {
            let action: unknown;
            if (selected.kind === 'route') {
                action = await dispatch(listCustomRouteVersions({ param: { offset: (page - 1) * limit, limit, id: selected.raw.id as string } }));
            } else if (selected.kind === 'ratelimit-local' || selected.kind === 'ratelimit-global') {
                action = await dispatch(listRateLimitRuleVersions({ param: { offset: (page - 1) * limit, limit, rule_name: selected.name } }));
            } else if (selected.kind === 'circuitbreaker') {
                action = await dispatch(listCircuitBreakerVersions({ param: { offset: (page - 1) * limit, limit, rule_name: selected.name } }));
            } else if (selected.kind === 'faultdetect') {
                action = await dispatch(listFaultDetectVersions({ param: { offset: (page - 1) * limit, limit, id: selected.raw.id as string } }));
            } else if (selected.kind === 'lossless') {
                action = await dispatch(listLosslessRuleVersions({ param: { offset: (page - 1) * limit, limit, id: selected.raw.id as string } }));
            } else if (selected.kind === 'lane') {
                action = await dispatch(listLaneGroupVersions({ param: { offset: (page - 1) * limit, limit, id: selected.raw.id as string } }));
            } else if (selected.trafficKind) {
                const result = await describeTrafficGovernanceVersions(selected.trafficKind, {
                    offset: (page - 1) * limit,
                    limit,
                    id: selected.raw.id as string,
                    rule_name: selected.name,
                });
                setVersions(result.list);
                setVersionTotal(result.totalCount);
                setVersionPage(page);
                setVersionLimit(limit);
                return;
            }

            const payload = getActionPayload<{ datas?: RuleRelease[], versions?: RuleRelease[], total?: number, versionTotal?: number, page?: number, versionPage?: number, limit?: number, versionLimit?: number }>(action);
            setVersions(payload?.datas || payload?.versions || []);
            setVersionTotal(payload?.total || payload?.versionTotal || 0);
            setVersionPage(payload?.page || payload?.versionPage || page);
            setVersionLimit(payload?.limit || payload?.versionLimit || limit);
        } finally {
            setVersionLoading(false);
        }
    };

    const operateRelease = async (op: Op, row: TableRowData) => {
        if (!selected) return;
        if (selected.kind === 'route') return;
        if (op === 'delete') {
            if (selected.kind === 'ratelimit-local' || selected.kind === 'ratelimit-global') {
                await dispatch(removeRateLimitRuleVersion({ id: row.id }));
            } else if (selected.kind === 'circuitbreaker') {
                await dispatch(removeCircuitBreakerRelease({ id: row.id }));
            } else if (selected.kind === 'faultdetect') {
                await dispatch(removeFaultDetectVersion({ id: row.id }));
            } else if (selected.kind === 'lossless') {
                await dispatch(removeLosslessVersion({ id: row.id }));
            } else if (selected.kind === 'lane') {
                await dispatch(removeLaneGroupVersion({ ids: [row.id] }));
            } else if (selected.trafficKind) {
                await deleteTrafficGovernanceRelease(selected.trafficKind, row.id as string);
            }
            openInfoNotification('请求成功', '删除规则版本成功');
            refreshVersions(versionPage, versionLimit);
        }
        if (op === 'rollback') {
            if (selected.kind === 'ratelimit-local' || selected.kind === 'ratelimit-global') {
                await dispatch(rollbackRateLimitRuleVersion({ id: row.id }));
            } else if (selected.kind === 'circuitbreaker') {
                await dispatch(rollbackCircuitBreakerRelease({ id: row.id }));
            } else if (selected.kind === 'faultdetect') {
                await dispatch(rollbackFaultDetectVersion({ id: row.id }));
            } else if (selected.kind === 'lossless') {
                await dispatch(rollbackLosslessVersion({ id: row.id }));
            } else if (selected.kind === 'lane') {
                await dispatch(rollbackLanGroupVersion({ id: row.id }));
            } else if (selected.trafficKind) {
                return;
            }
            openInfoNotification('请求成功', '回滚规则版本成功');
            refreshVersions(versionPage, versionLimit);
        }
    };

    const filteredRules = rules.filter((rule) => {
        const typeMatched = typeFilters.length === 0 || typeFilters.some((type) => (type === 'ratelimit' ? rule.kind.startsWith('ratelimit') : rule.kind === type));
        const searchMatched = !search || [rule.name, rule.description, rule.namespace, rule.service, rule.condition].some((item) => item?.toLowerCase().includes(search.toLowerCase()));
        return typeMatched && searchMatched;
    });

    const namespaceOptions = Array.from(new Set(rules.map((item) => item.namespace).filter(Boolean))).map((item) => ({ label: item as string, value: item as string }));
    const enabledCount = rules.filter((item) => item.status === '启用').length;
    const pendingCount = rules.filter((item) => item.release === '待发布').length;

    const columns: PrimaryTableProps['columns'] = [
        {
            colKey: 'name',
            title: '规则',
            width: 300,
            cell: ({ row }) => (
                <div className={style.ruleName}>
                    <Link className={style.ruleNameText} theme="primary" onClick={() => openRule(row as GovernanceRuleRow)}>{row.name}</Link>
                    <div className={style.ruleDesc}>{row.condition || row.description || '-'}</div>
                </div>
            ),
        },
        {
            colKey: 'typeLabel',
            title: '类型',
            width: 110,
            cell: ({ row }) => <Tag className={`${style.typeTag} ${getTypeClassName(row.kind as RuleKind)}`} variant="light-outline">{row.typeLabel}</Tag>,
        },
        {
            colKey: 'target',
            title: '作用对象',
            minWidth: 280,
            cell: ({ row }) => <div className={style.target}>{row.target}</div>,
        },
        {
            colKey: 'status',
            title: '状态',
            width: 120,
            cell: ({ row }) => (
                <span className={style.statusCell}>
                    <span className={row.status === '启用' ? style.statusEnabled : style.statusDisabled} />
                    {row.status === '启用' ? '已启用' : '已停用'}
                </span>
            ),
        },
        {
            colKey: 'time',
            title: '操作时间',
            width: 190,
            cell: ({ row }) => (
                <div className={style.timeCell}>
                    <span className={style.timeText}>{formatTime(row.mtime)}</span>
                    <span className={style.timeSecondary}>创建: {formatTime(row.ctime)}</span>
                </div>
            ),
        },
        {
            colKey: 'operation',
            title: '操作',
            width: 112,
            fixed: 'right',
            cell: ({ row }) => {
                const rule = row as GovernanceRuleRow;
                return (
                    <Space className={style.rowActions} size={10} onClick={(event) => event.stopPropagation()}>
                        <Tooltip content="查看 / 编辑">
                            <Button
                                shape="square"
                                variant="text"
                                aria-label="查看 / 编辑"
                                onClick={() => openRule(rule)}
                            >
                                <CreditcardIcon />
                            </Button>
                        </Tooltip>
                        <Tooltip content={rule.deleteable === false ? '无权限操作' : '删除'}>
                            <Popconfirm
                                content={`确认删除规则 ${rule.name}？`}
                                destroyOnClose
                                placement="top"
                                showArrow
                                theme="default"
                                onConfirm={() => deleteRule(rule)}
                            >
                                <Button shape="square" variant="text" disabled={rule.deleteable === false}>
                                    <DeleteIcon />
                                </Button>
                            </Popconfirm>
                        </Tooltip>
                    </Space>
                );
            },
        },
    ];

    const renderCreateWizard = () => (
        <Dialog
            header="新建规则"
            visible={createWizardVisible}
            width={720}
            confirmBtn="进入创建"
            cancelBtn="取消"
            onClose={() => setCreateWizardVisible(false)}
            onConfirm={confirmCreateWizard}
        >
            <div className={style.createWizard}>
                <Steps current={createWizardStep}>
                    <StepItem value={1} title="选择规则类型" />
                    <StepItem value={2} title="配置规则" />
                </Steps>
                <div className={style.createTypeGrid}>
                    {typeOptions.map((item) => {
                        const target = createRuleTargets[item.value];
                        const active = pendingCreateType === item.value;
                        return (
                            <button
                                key={item.value}
                                className={active ? style.createTypeCardActive : style.createTypeCard}
                                type="button"
                                onClick={() => setPendingCreateType(item.value)}
                            >
                                <Tag className={`${style.typeTag} ${getTypeClassName(target.kind)}`} variant="light-outline">{item.label}</Tag>
                                <p>{item.scene}</p>
                            </button>
                        );
                    })}
                </div>
            </div>
        </Dialog>
    );

    const renderDetail = () => {
        if (!selected) return null;
        const commonVersions = {
            datas: versions,
            action: operateRelease,
            editable: selected.editable ?? true,
            deleteable: selected.deleteable ?? true,
            rollbackable: selected.kind !== 'lossless' && !selected.trafficKind,
            loading: versionLoading,
            pagination: {
                current: versionPage,
                pageSize: versionLimit,
                total: versionTotal,
                showJumper: false,
                onChange(pageInfo: { current: number, pageSize: number }) {
                    refreshVersions(pageInfo.current, pageInfo.pageSize);
                },
            },
            onPageChange: (page: { current: number, pageSize: number }) => {
                refreshVersions(page.current, page.pageSize);
            },
        };
        const subscribe = (
            <div style={{ marginLeft: 20, marginTop: 20 }}>
                <SubscribeTable
                    title={selected.name}
                    editable={selected.editable ?? true}
                    deleteable={selected.deleteable ?? true}
                    subscribers={[]}
                />
            </div>
        );

        if (selected.kind === 'route') {
            return <RuleTabs op={drawerMode} onVersionView={() => refreshVersions()} view={<CustomRouteEditor op={drawerMode} editable={selected.editable ?? true} refresh={(close) => {
                if (close) {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }
                refreshData();
            }} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.kind === 'ratelimit-local' || selected.kind === 'ratelimit-global') {
            return <RuleTabs op={drawerMode} onVersionView={() => refreshVersions()} view={<RateLimitEditor limitType={selected.limitType || LimitType.LOCAL} visible={drawerVisible} op={drawerMode} refresh={(close) => {
                if (close) {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }
                refreshData();
            }} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.kind === 'circuitbreaker') {
            return <RuleTabs op={drawerMode} onVersionView={() => refreshVersions()} view={<CircuitBreakerEditor op={drawerMode} refresh={(close) => {
                if (close) {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }
                refreshData();
            }} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.kind === 'faultdetect') {
            return <RuleTabs op={drawerMode} onVersionView={() => refreshVersions()} view={<FaultDetectEditor op={drawerMode} refresh={(close) => {
                if (close) {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }
                refreshData();
            }} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.kind === 'lossless') {
            return <RuleTabs op={drawerMode} onVersionView={() => refreshVersions()} view={<LossLessEditor visible={drawerVisible} op={drawerMode} refresh={(close) => {
                if (close) {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }
                refreshData();
            }} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.trafficKind) {
            return (
                <RuleTabs
                    op="view"
                    onVersionView={() => refreshVersions()}
                    view={(
                        <TrafficGovernanceEditor
                            kind={selected.trafficKind}
                            op={drawerMode}
                            data={drawerMode === 'create' ? undefined : selected.raw as TrafficGovernanceRule}
                            visible={drawerVisible}
                            refresh={(close) => {
                                if (close) {
                                    setDrawerVisible(false);
                                    setDrawerMode('view');
                                }
                                refreshData();
                            }}
                        />
                    )}
                    versions={commonVersions}
                    subscribe={subscribe}
                />
            );
        }
        return <LaneGroupEdtor op={drawerMode} refresh={(close) => {
            if (close) {
                setDrawerVisible(false);
                setDrawerMode('view');
            }
            refreshData();
        }} />;
    };

    return (
        <div className={style.page}>
            <div className={style.header}>
                <div>
                    <div className={style.title}>规则治理工作台</div>
                    <div className={style.description}>统一查看治理规则的运行状态、作用范围和发布状态。</div>
                </div>
            </div>
            <div className={style.summary}>
                <div className={style.summaryItem}>
                    <div className={style.summaryLabel}>规则总数</div>
                    <div className={style.summaryValue}>{rules.length}</div>
                    <div className={style.summaryHint}>覆盖 {namespaceOptions.length} 个命名空间</div>
                </div>
                <div className={style.summaryItem}>
                    <div className={style.summaryLabel}>已启用</div>
                    <div className={style.summaryValue}>{enabledCount}</div>
                    <div className={style.summaryHint}>路由、限流、熔断、探测等规则</div>
                </div>
                <div className={style.summaryItem}>
                    <div className={style.summaryLabel}>待发布</div>
                    <div className={style.summaryValue}>{pendingCount}</div>
                    <div className={style.summaryHint}>需进入详情确认版本</div>
                </div>
                <div className={style.summaryItem}>
                    <div className={style.summaryLabel}>规则类型</div>
                    <div className={style.summaryValue}>{new Set(rules.map((item) => item.typeLabel)).size}</div>
                    <div className={style.summaryHint}>按治理场景分类查看</div>
                </div>
            </div>
            <div className={style.listPanel}>
                <div className={style.panelHeader}>
                    <div className={style.panelIntro}>
                        <div className={style.panelTitle}>规则清单</div>
                        <div className={style.panelHint}>点击规则行查看详情，支持按类型和关键词筛选。</div>
                    </div>
                </div>
                <div className={style.filters}>
                    <Select
                        className={style.typeSelect}
                        multiple
                        clearable
                        value={typeFilters}
                        placeholder="全部规则类型"
                        options={typeOptions}
                        onChange={(value) => setTypeFilters(Array.isArray(value) ? value as string[] : [])}
                    />
                    <Input
                        className={style.searchInput}
                        clearable
                        value={search}
                        placeholder="搜索规则名、服务、条件"
                        onChange={(value) => setSearch(value)}
                        onEnter={(value) => refreshData(value)}
                        onClear={() => {
                            setSearch('');
                            refreshData('');
                        }}
                    />
                    <Space className={style.filterActions}>
                        <Tooltip content="刷新">
                            <Button
                                aria-label="刷新"
                                shape="square"
                                theme="primary"
                                icon={<RefreshIcon />}
                                onClick={() => refreshData(search)}
                            />
                        </Tooltip>
                        <Tooltip content="重置筛选">
                            <Button
                                aria-label="重置筛选"
                                shape="square"
                                icon={<FilterClearIcon />}
                                onClick={() => {
                                    setTypeFilters([]);
                                    setSearch('');
                                    refreshData('');
                                }}
                            />
                        </Tooltip>
                        <Button
                            theme="primary"
                            icon={<AddIcon />}
                            onClick={openCreateWizard}
                        >
                            新建规则
                        </Button>
                    </Space>
                </div>
                <Table
                    className={style.table}
                    data={filteredRules}
                    columns={columns}
                    loading={loading}
                    rowKey="key"
                    tableLayout="fixed"
                    cellEmptyContent="-"
                    pagination={{
                        pageSize: 10,
                        total: filteredRules.length,
                        showJumper: true,
                    }}
                    onRowClick={({ row }) => openRule(row as GovernanceRuleRow)}
                />
            </div>
            <RuleDetailDrawer
                visible={drawerVisible}
                title={drawerMode === 'create' ? selected?.name || '新建治理规则' : selected?.name || '治理规则详情'}
                subtitle={selected?.typeLabel}
                size={selected?.kind === 'route' || selected?.kind?.startsWith('ratelimit') || selected?.kind === 'circuitbreaker' || selected?.kind === 'faultdetect' || selected?.kind === 'lane' ? WIDE_RULE_DETAIL_DRAWER_SIZE : undefined}
                onClose={() => {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }}
            >
                {renderDetail()}
            </RuleDetailDrawer>
            {renderCreateWizard()}
        </div>
    );
};

export default React.memo(GovernanceWorkbench);
