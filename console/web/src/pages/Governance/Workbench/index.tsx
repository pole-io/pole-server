import React from 'react';
import { Button, Input, Link, Select, Space, Table, Tag } from 'tdesign-react';
import type { PrimaryTableProps, TableRowData } from 'tdesign-react';
import { AddIcon, RefreshIcon } from 'tdesign-icons-react';

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
} from 'modules/governance/route';
import {
    editorRateLimitRule,
    listRateLimitRuleVersions,
    listRateLimitRules,
    removeRateLimitRuleVersion,
    rollbackRateLimitRuleVersion,
} from 'modules/governance/ratelimit';
import {
    editorCircuitBreaker,
    listCircuitBreakerVersions,
    listCircuitBreakers,
    removeCircuitBreakerRelease,
    resetCircuitBreaker,
    rollbackCircuitBreakerRelease,
} from 'modules/governance/circuitbreaker';
import {
    editorFaultDetect,
    listFaultDetectVersions,
    listFaultDetects,
    removeFaultDetectVersion,
    rollbackFaultDetectVersion,
} from 'modules/governance/faultdetect';
import {
    editorLosslessRule,
    listLossLessRules,
    listLosslessRuleVersions,
    removeLosslessVersion,
    rollbackLosslessVersion,
} from 'modules/governance/lossless';
import {
    editorLaneGroup,
    listLaneGroupVersions,
    listLaneGroups,
    removeLaneGroupVersion,
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
    describeTrafficGovernanceRules,
    describeTrafficGovernanceVersions,
    TrafficGovernanceKind,
    TrafficGovernanceKindLabel,
    TrafficGovernanceRule,
} from 'services/traffic_governance';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';

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
    editable?: boolean;
    deleteable?: boolean;
    raw: TableRowData;
    limitType?: LimitType;
    trafficKind?: TrafficGovernanceKind;
}

const typeOptions = [
    { label: '全部', value: 'all' },
    { label: '路由', value: 'route' },
    { label: '限流', value: 'ratelimit' },
    { label: '熔断', value: 'circuitbreaker' },
    { label: '探测', value: 'faultdetect' },
    { label: '无损', value: 'lossless' },
    { label: '泳道', value: 'lane' },
    { label: '鉴权', value: 'traffic-security' },
    { label: '镜像', value: 'traffic-mirror' },
    { label: 'Mock', value: 'traffic-mock' },
];

const getActionPayload = <T,>(action: unknown): T | undefined => {
    const result = action as { meta?: { requestStatus?: string }, payload?: T };
    if (result.meta?.requestStatus === 'fulfilled') return result.payload;
    return undefined;
};

const text = (value?: string | number | boolean) => {
    if (value === undefined || value === null || value === '') return '-';
    return String(value);
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
            condition: firstRule ? `${text(firstRule.method?.value || firstRule.resource)} / ${LimitActionMap[firstRule.action] || firstRule.action || '-'}` : '按接口或资源限流',
            status: rule.disable ? '禁用' : '启用',
            release: rule.revision ? '已发布' : '待发布',
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
    const [typeFilter, setTypeFilter] = React.useState('all');
    const [namespaceFilter, setNamespaceFilter] = React.useState('');
    const [serviceFilter, setServiceFilter] = React.useState('');
    const [selected, setSelected] = React.useState<GovernanceRuleRow | null>(null);
    const [drawerVisible, setDrawerVisible] = React.useState(false);
    const [drawerMode, setDrawerMode] = React.useState<Op>('view');
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

    const openRule = (rule: GovernanceRuleRow) => {
        setSelected(rule);
        setDrawerMode('view');
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

    const openCreateCircuitBreaker = () => {
        setSelected({
            key: 'circuitbreaker-create',
            kind: 'circuitbreaker',
            typeLabel: '熔断',
            name: '新建熔断规则',
            description: '',
            namespace: '',
            service: '',
            target: '-',
            condition: '-',
            status: '启用',
            release: '待发布',
            raw: { id: '' } as TableRowData,
        });
        setDrawerMode('create');
        setVersions([]);
        setVersionTotal(0);
        setVersionPage(1);
        setVersionLimit(10);
        dispatch(resetCircuitBreaker());
        setDrawerVisible(true);
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
        const typeMatched = typeFilter === 'all' || (typeFilter === 'ratelimit' ? rule.kind.startsWith('ratelimit') : rule.kind === typeFilter);
        const namespaceMatched = !namespaceFilter || rule.namespace === namespaceFilter;
        const serviceMatched = !serviceFilter || rule.service === serviceFilter;
        const searchMatched = !search || [rule.name, rule.description, rule.namespace, rule.service, rule.condition].some((item) => item?.toLowerCase().includes(search.toLowerCase()));
        return typeMatched && namespaceMatched && serviceMatched && searchMatched;
    });

    const namespaceOptions = Array.from(new Set(rules.map((item) => item.namespace).filter(Boolean))).map((item) => ({ label: item as string, value: item as string }));
    const serviceOptions = Array.from(new Set(rules.map((item) => item.service).filter(Boolean))).map((item) => ({ label: item as string, value: item as string }));
    const enabledCount = rules.filter((item) => item.status === '启用').length;
    const pendingCount = rules.filter((item) => item.release === '待发布').length;

    const columns: PrimaryTableProps['columns'] = [
        {
            colKey: 'name',
            title: '规则',
            width: 280,
            cell: ({ row }) => (
                <div className={style.ruleName}>
                    <Link className={style.ruleNameText} theme="primary" onClick={() => openRule(row as GovernanceRuleRow)}>{row.name}</Link>
                    <div className={style.ruleDesc}>{row.description || '-'}</div>
                </div>
            ),
        },
        {
            colKey: 'typeLabel',
            title: '类型',
            width: 120,
            cell: ({ row }) => <Tag theme={row.kind === 'circuitbreaker' ? 'danger' : row.kind === 'lossless' ? 'success' : 'primary'} variant="light-outline">{row.typeLabel}</Tag>,
        },
        {
            colKey: 'target',
            title: '作用对象',
            width: 240,
            cell: ({ row }) => <div className={style.target}>{row.target}</div>,
        },
        {
            colKey: 'condition',
            title: '匹配 / 触发',
            ellipsis: true,
        },
        {
            colKey: 'status',
            title: '运行状态',
            width: 120,
            cell: ({ row }) => <Tag theme={row.status === '启用' ? 'success' : 'default'} variant="light-outline">{row.status}</Tag>,
        },
        {
            colKey: 'release',
            title: '发布',
            width: 120,
            cell: ({ row }) => <Tag theme={row.release === '待发布' ? 'warning' : 'primary'} variant="light-outline">{row.release}</Tag>,
        },
    ];

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
            return <RuleTabs op="view" onVersionView={() => refreshVersions()} view={<CustomRouteEditor op="view" editable={selected.editable ?? true} refresh={() => refreshData()} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.kind === 'ratelimit-local' || selected.kind === 'ratelimit-global') {
            return <RuleTabs op="view" onVersionView={() => refreshVersions()} view={<RateLimitEditor limitType={selected.limitType || LimitType.LOCAL} visible={drawerVisible} op="view" refresh={() => refreshData()} />} versions={commonVersions} subscribe={subscribe} />;
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
            return <RuleTabs op="view" onVersionView={() => refreshVersions()} view={<FaultDetectEditor op="view" refresh={() => refreshData()} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.kind === 'lossless') {
            return <RuleTabs op="view" onVersionView={() => refreshVersions()} view={<LossLessEditor visible={drawerVisible} op="view" refresh={() => refreshData()} />} versions={commonVersions} subscribe={subscribe} />;
        }
        if (selected.trafficKind) {
            return (
                <RuleTabs
                    op="view"
                    onVersionView={() => refreshVersions()}
                    view={(
                        <TrafficGovernanceEditor
                            kind={selected.trafficKind}
                            op="view"
                            data={selected.raw as TrafficGovernanceRule}
                            visible={drawerVisible}
                            refresh={() => refreshData()}
                        />
                    )}
                    versions={commonVersions}
                    subscribe={subscribe}
                />
            );
        }
        return <LaneGroupEdtor op="view" refresh={() => refreshData()} />;
    };

    return (
        <div className={style.page}>
            <div className={style.header}>
                <div>
                    <div className={style.title}>规则治理工作台</div>
                    <div className={style.description}>统一查看治理规则的运行状态、作用范围和发布状态。</div>
                </div>
                <Space>
                    {typeFilter === 'circuitbreaker' && (
                        <Button theme="primary" prefix={<AddIcon />} onClick={openCreateCircuitBreaker}>新建熔断规则</Button>
                    )}
                    <Button onClick={() => refreshData(search)} prefix={<RefreshIcon />}>刷新</Button>
                </Space>
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
                    <div>
                        <div className={style.panelTitle}>规则清单</div>
                        <div className={style.panelHint}>点击规则行查看详情，支持按类型、命名空间和服务筛选。</div>
                    </div>
                    <Space>
                        {typeOptions.map((item) => (
                            <Button
                                key={item.value}
                                theme={typeFilter === item.value ? 'primary' : 'default'}
                                variant={typeFilter === item.value ? 'base' : 'outline'}
                                onClick={() => setTypeFilter(item.value)}
                            >
                                {item.label}
                            </Button>
                        ))}
                    </Space>
                </div>
                <div className={style.filters}>
                    <Input
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
                    <Select
                        clearable
                        value={namespaceFilter}
                        placeholder="全部命名空间"
                        options={namespaceOptions}
                        onChange={(value) => setNamespaceFilter(value as string || '')}
                    />
                    <Select
                        clearable
                        value={serviceFilter}
                        placeholder="全部服务"
                        options={serviceOptions}
                        onChange={(value) => setServiceFilter(value as string || '')}
                    />
                    <Button onClick={() => {
                        setTypeFilter('all');
                        setNamespaceFilter('');
                        setServiceFilter('');
                        setSearch('');
                        refreshData('');
                    }}>重置</Button>
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
                title={drawerMode === 'create' ? '新建熔断规则' : selected?.name || '治理规则详情'}
                subtitle={selected?.typeLabel}
                size={selected?.kind === 'route' || selected?.kind?.startsWith('ratelimit') || selected?.kind === 'circuitbreaker' || selected?.kind === 'faultdetect' || selected?.kind === 'lane' ? WIDE_RULE_DETAIL_DRAWER_SIZE : undefined}
                onClose={() => {
                    setDrawerVisible(false);
                    setDrawerMode('view');
                }}
            >
                {renderDetail()}
            </RuleDetailDrawer>
        </div>
    );
};

export default React.memo(GovernanceWorkbench);
