import React from "react";
import { Col, Form, Input, Row, Space, Button, Select, Switch, Dialog, InputNumber, Table, FormProps, Tag, TagInput, Popup, TableRowData, PrimaryTableProps, StickyTool } from "tdesign-react";
import { SendIcon, AddIcon, CloseIcon, Edit1Icon, SaveIcon, RocketIcon, RollbackIcon, DragMoveIcon, ChevronRightIcon } from "tdesign-icons-react";
import cloneDeep from 'lodash/cloneDeep';

import Text from "components/Text";
import RuleLabelField from "../shared/RuleLabelField";
import shared from "../shared/governance.module.less";
import { useAppDispatch, useAppSelector } from 'modules/store';
import {
    CustomRoute,
    CustomRouteView,
    RouteArgumentTextMap,
    RoutingArgumentsType,
    RoutingArgumentsTypeOptions,
    RoutingConfig,
    RoutingLabel,
    RoutingRule,
    RoutingRuleDestination,
    RoutingRuleSource,
    RoutingSourceArgument,
    RoutingValueType,
    normalizeRoutingConfigForEditor,
} from "services/router";
import { Label, MatchString, MatchType, MatchTypeMap, MatchTypeOption, MatchValueType, Op } from "services/types";
import { listOneCustomRoute, resetCustomRoute, saveCustomRoutes, selectCustomRoute, updateCustomRoutes } from "modules/governance/route";

import styles from './CustomRouteEditor.module.less';
import { openErrNotification, openInfoNotification } from "utils/notifition";
import PublishForm from "../RuleRelease/PublishForm";
import RuleStickyAction from "../RuleRelease/RuleStickyAction";
import { PolicySourceType } from "services/auth_policy";
import { cleanNamespacePage, listAllNamespaces, selectNamespace } from "modules/namespace";
import { cleanServicePage, listAllServices, selectService } from "modules/discovery/service";
import {
    buildRouteRuleSubmitPayload,
    commaStringToTags,
    getDefaultParamKey,
    getRouteRuleArguments,
    getWeightStatus,
    isTagInputMatchType,
    prepareRouteRuleDraftForSubmit,
    stringifyRouteRuleSpec,
    tagsToCommaString,
    validateRouteRuleDraft,
    withParamTypeDefaultKey,
    RouteSpecFormat,
    RouteRuleValidationError,
} from "./routeEditorUtils";

const { FormItem } = Form;
const { StickyItem } = StickyTool;
const customRouteConfigType = 'type.googleapis.com/v1.CustomRoute';

const defaultMatchArgs: () => RoutingSourceArgument = () => ({
    type: RoutingArgumentsType.HEADER,
    key: getDefaultParamKey(RoutingArgumentsType.HEADER),
    value: {
        type: MatchType.EXACT,
        value: '',
        value_type: MatchValueType.TEXT
    }
});

const defaultMatch: () => RoutingRuleSource = () => ({
    namespace: '',
    service: '',
    arguments: [defaultMatchArgs()]
});

const getAutoGroupName = (index: number) => `Group ${index + 1}`;

const normalizeDestinationGroupNames = (destinations?: RoutingRuleDestination[]) => (
    (destinations || []).map((group, index) => ({
        ...group,
        name: getAutoGroupName(index),
        priority: index,
    }))
);

const normalizeRouteRuleGroupNames = (rules?: RoutingRule[]) => (
    (rules || []).map((rule) => ({
        ...rule,
        destinations: normalizeDestinationGroupNames(rule.destinations),
    }))
);

const defaultGroup: (id: number) => RoutingRuleDestination = (id: number) => ({
    service: '',
    namespace: '',
    name: getAutoGroupName(id - 1),
    weight: 100,
    isolate: false,
    labels: {},
    priority: 0
});

export const defaultCustomRoute: () => CustomRouteDO = () => ({
    name: '',
    description: '',
    priority: 5,
    routing_policy: 'RulePolicy',
    metadata: [],
    enable: true,
    routing_config: {
        '@type': customRouteConfigType,
        caller: {
            namespace: '*',
            service: '',
        },
        callee: {
            namespace: '*',
            service: '',
        },
        rules: [{
            name: '',
            sources: [defaultMatch()],
            arguments: {
                arguments: [],
                randomPercent: 0,
                matchMode: 'AND',
            },
            destinations: [defaultGroup(1)],
        }]
    },
    caller_namespace: '*',
    caller_service: '',
    callee_namespace: '*',
    callee_service: '',
});

export interface CustomRouteDO {
    id?: string
    name?: string // 规则名
    enable?: boolean // 是否启用
    priority?: number
    description?: string
    routing_config?: RoutingConfig
    routing_policy?: 'RulePolicy' | 'NearbyPolicy'
    metadata: Label[]

    // 额外定义的服务数据信息
    caller_namespace: string;
    caller_service: string;
    callee_namespace: string;
    callee_service: string;
}

interface ICustomRouteEditorProps {
    op: Op;
    refresh: (close: boolean) => void;
    editable: boolean;
}

const CustomRouteEditor: React.FC<ICustomRouteEditorProps> = ({ op, refresh, editable }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { datas: serviceDatas } = serviceState;

    const customRouteState = useAppSelector(selectCustomRoute);
    const { editRoute, viewRoute } = customRouteState;

    // 创建新的 rules 对象
    const [customRouteRule, setCustomRouteRule] = React.useState<CustomRouteDO>(defaultCustomRoute());
    const [collapsedRuleIndexes, setCollapsedRuleIndexes] = React.useState<Set<number>>(new Set());
    const [serviceCollapsed, setServiceCollapsed] = React.useState(false);
    const [specFormat, setSpecFormat] = React.useState<RouteSpecFormat>('yaml');
    const [validationErrors, setValidationErrors] = React.useState<RouteRuleValidationError[]>([]);
    const [draggingRuleIndex, setDraggingRuleIndex] = React.useState<number | null>(null);
    const [draggingGroupKey, setDraggingGroupKey] = React.useState<string>('');

    const resetCurRule = (rule: CustomRouteView | null) => {
        if (!rule) {
            return;
        }
        const cloneRule = cloneDeep(rule);
        const routingConfig = normalizeRoutingConfigForEditor(cloneRule.routing_config);
        const normalizedRoutingConfig = routingConfig ? {
            ...routingConfig,
            rules: normalizeRouteRuleGroupNames(routingConfig.rules),
        } : routingConfig;
        const firstRule = normalizedRoutingConfig?.rules?.[0];
        const caller = normalizedRoutingConfig?.caller;
        const callee = normalizedRoutingConfig?.callee;
        setCustomRouteRule({
            ...cloneRule,
            routing_config: normalizedRoutingConfig,
            caller_namespace: caller?.namespace || firstRule?.sources?.[0]?.namespace || '*',
            caller_service: caller?.service || firstRule?.sources?.[0]?.service || '',
            callee_namespace: callee?.namespace || firstRule?.destinations?.[0]?.namespace || '*',
            callee_service: callee?.service || firstRule?.destinations?.[0]?.service || '',
            metadata: Object.entries(cloneRule?.metadata || {}).map(([key, value]) => ({ key, value })),
        });
        setCollapsedRuleIndexes(new Set());
    }

    const [tagEdit, setTagEdit] = React.useState<{
        visible: boolean,
        ruleIdx: number,
        groupIdx: number,
        tags: RoutingLabel[]
    }>({ visible: false, ruleIdx: -1, groupIdx: -1, tags: [] });

    const [editorState, setEditorState] = React.useState<{
        model: Op;
        publishView: boolean;
        visible: boolean
        editable?: boolean
    }>({ model: 'view', publishView: false, visible: false, editable: op === 'create' || false, });

    const previewSpec = React.useMemo(() => buildRouteRuleSubmitPayload(customRouteRule), [customRouteRule]);
    const previewText = React.useMemo(() => stringifyRouteRuleSpec(previewSpec, specFormat), [previewSpec, specFormat]);
    const liveValidationErrors = React.useMemo(() => validateRouteRuleDraft(customRouteRule), [customRouteRule]);


    React.useEffect(() => {
        dispatch(listAllNamespaces()).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求失败', `获取命名空间列表失败, ${res.payload as string}`);
            }
        });
        dispatch(listAllServices()).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求失败', `获取服务列表失败, ${res.payload as string}`);
            }
        });

        return () => {
            // 清理编辑器状态
            dispatch(cleanNamespacePage());
            dispatch(cleanServicePage());
        };
    }, []);

    React.useEffect(() => {
        if (editRoute) {
            if (editRoute.id !== '') {
                dispatch(listOneCustomRoute({ id: editRoute.id || '' })).then((res) => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('请求失败', `获取自定义路由详情失败, ${res.payload as string}`);
                    } else {
                        const ret = res.payload as { viewRoute: CustomRouteView } | null
                        resetCurRule(ret?.viewRoute || null);
                    }
                })
            } else {
                resetCurRule(editRoute)
            }
        }
    }, [editRoute?.id])

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        const draftForSubmit = prepareRouteRuleDraftForSubmit(customRouteRule);
        const errors = validateRouteRuleDraft(draftForSubmit);
        setValidationErrors(errors);
        if (errors.length > 0) {
            openErrNotification('校验失败', errors[0].message);
            return;
        }

        const data = buildRouteRuleSubmitPayload(draftForSubmit) as CustomRoute;

        let res;
        if (op === 'view') {
            res = await dispatch(updateCustomRoutes({ param: data }));
        } else {
            res = await dispatch(saveCustomRoutes({ param: data }));
        }

        if (res.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', `${op === 'view' ? '修改' : '创建'}自定义路由规则失败, ${res?.payload as string}`);
        } else {
            openInfoNotification('请求成功', op === 'view' ? '修改自定义路由规则成功' : '创建自定义路由规则成功');
            if (op === 'view') {
                setEditorState(prev => ({ ...prev, editable: false }));
            }
            refresh(op !== 'view'); // 刷新列表
        }
    }

    const updateRoutingRules = (updater: (rules: RoutingRule[]) => RoutingRule[]) => {
        setCustomRouteRule((prev) => ({
            ...prev,
            routing_config: {
                ...(prev.routing_config || {}),
                '@type': prev.routing_config?.['@type'] || customRouteConfigType,
                rules: updater(cloneDeep(prev.routing_config?.rules || [])),
            },
        }));
    };

    const addMatch = (ruleIdx: number) => {
        updateRoutingRules((rules) => {
            if (!rules[ruleIdx].sources?.[0]) {
                rules[ruleIdx].sources = [defaultMatch()];
            }
            if (!rules[ruleIdx].sources[0].arguments) {
                rules[ruleIdx].sources[0].arguments = [];
            }
            rules[ruleIdx].sources[0].arguments.push(defaultMatchArgs());
            return rules;
        });
    };

    const updateMatchArgs = (del: boolean, ruleIdx: number, idx: number, args?: RoutingSourceArgument) => {
        updateRoutingRules((rules) => {
            if (!rules[ruleIdx].sources?.[0]) {
                rules[ruleIdx].sources = [defaultMatch()];
            }
            if (!rules[ruleIdx].sources[0].arguments) {
                rules[ruleIdx].sources[0].arguments = [];
            }
            if (del) {
                if (rules[ruleIdx].sources[0].arguments.length <= 1) {
                    return rules;
                }
                rules[ruleIdx].sources[0].arguments.splice(idx, 1);
            } else {
                if (args) {
                    if (idx < rules[ruleIdx].sources[0].arguments.length) {
                        // 更新已有的参数
                        rules[ruleIdx].sources[0].arguments[idx] = {
                            ...rules[ruleIdx].sources[0].arguments[idx],
                            ...Object.fromEntries(
                                Object.entries(args).filter(([_, value]) => value !== undefined)
                            ),
                        };
                    } else {
                        rules[ruleIdx].sources[0].arguments.push(args);
                    }
                }
            }
            return rules;
        });
    };

    const updateRuleMatchMode = (ruleIdx: number, matchMode: 'AND' | 'OR') => {
        updateRoutingRules((rules) => {
            rules[ruleIdx].arguments = {
                ...(rules[ruleIdx].arguments || {}),
                randomPercent: rules[ruleIdx].arguments?.randomPercent || 0,
                arguments: rules[ruleIdx].arguments?.arguments || [],
                matchMode,
            };
            return rules;
        });
    };

    const addGroup = (ruleIdx: number) => {
        updateRoutingRules((rules) => {
            if (!rules[ruleIdx].destinations) {
                rules[ruleIdx].destinations = [];
            }
            const newGroup = defaultGroup(rules[ruleIdx].destinations.length + 1);
            rules[ruleIdx].destinations.push(newGroup);
            rules[ruleIdx].destinations = normalizeDestinationGroupNames(rules[ruleIdx].destinations);
            return rules;
        });
    };

    const removeGroup = (ruleIdx: number, idx: number) => {
        updateRoutingRules((rules) => {
            if ((rules[ruleIdx].destinations || []).length <= 1) {
                return rules;
            }
            rules[ruleIdx].destinations.splice(idx, 1);
            rules[ruleIdx].destinations = normalizeDestinationGroupNames(rules[ruleIdx].destinations);
            return rules;
        });
    };

    const updateGroup = (ruleIdx: number, idx: number, group: RoutingRuleDestination) => {
        updateRoutingRules((rules) => {
            if (idx < rules[ruleIdx].destinations.length) {
                const existingGroup = rules[ruleIdx].destinations[idx];
                rules[ruleIdx].destinations[idx] = {
                    ...existingGroup,
                    ...Object.fromEntries(
                        Object.entries(group).filter(([_, value]) => value !== undefined)
                    ),
                };
            } else {
                rules[ruleIdx].destinations.push(group);
            }
            return rules;
        });
    }

    const moveRule = (from: number, to: number) => {
        if (from === to || from < 0 || to < 0) return;
        updateRoutingRules((rules) => {
            if (from >= rules.length || to >= rules.length) return rules;
            const [moved] = rules.splice(from, 1);
            rules.splice(to, 0, moved);
            return rules;
        });
        setCollapsedRuleIndexes((prev) => {
            const next = new Set<number>();
            prev.forEach((idx) => {
                if (idx === from) {
                    next.add(to);
                } else if (from < to && idx > from && idx <= to) {
                    next.add(idx - 1);
                } else if (from > to && idx >= to && idx < from) {
                    next.add(idx + 1);
                } else {
                    next.add(idx);
                }
            });
            return next;
        });
    };

    const moveGroup = (ruleIdx: number, from: number, to: number) => {
        if (from === to || from < 0 || to < 0) return;
        updateRoutingRules((rules) => {
            const destinations = rules[ruleIdx]?.destinations || [];
            if (from >= destinations.length || to >= destinations.length) return rules;
            const [moved] = destinations.splice(from, 1);
            destinations.splice(to, 0, {
                ...moved,
                priority: to,
            });
            rules[ruleIdx].destinations = normalizeDestinationGroupNames(destinations);
            return rules;
        });
    };

    const toggleRuleCollapsed = (ruleIdx: number) => {
        setCollapsedRuleIndexes((prev) => {
            const next = new Set(prev);
            if (next.has(ruleIdx)) {
                next.delete(ruleIdx);
            } else {
                next.add(ruleIdx);
            }
            return next;
        });
    };

    const removeRule = (ruleIdx: number) => {
        updateRoutingRules((rules) => rules.filter((_, i) => i !== ruleIdx));
        setCollapsedRuleIndexes((prev) => {
            const next = new Set<number>();
            prev.forEach((idx) => {
                if (idx < ruleIdx) {
                    next.add(idx);
                } else if (idx > ruleIdx) {
                    next.add(idx - 1);
                }
            });
            return next;
        });
    };

    const handleRuleDragStart = (event: React.DragEvent, ruleIdx: number) => {
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('application/x-route-rule-index', String(ruleIdx));
        setDraggingRuleIndex(ruleIdx);
    };

    const handleRuleDrop = (event: React.DragEvent, ruleIdx: number) => {
        event.preventDefault();
        const from = Number(event.dataTransfer.getData('application/x-route-rule-index'));
        if (!Number.isNaN(from)) {
            moveRule(from, ruleIdx);
        }
        setDraggingRuleIndex(null);
    };

    const handleGroupDragStart = (event: React.DragEvent, ruleIdx: number, groupIdx: number) => {
        const dragKey = `${ruleIdx}:${groupIdx}`;
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('application/x-route-group-index', dragKey);
        setDraggingGroupKey(dragKey);
    };

    const handleGroupDrop = (event: React.DragEvent, ruleIdx: number, groupIdx: number) => {
        event.preventDefault();
        const [fromRule, fromGroup] = event.dataTransfer.getData('application/x-route-group-index').split(':').map(Number);
        if (fromRule === ruleIdx && !Number.isNaN(fromGroup)) {
            moveGroup(ruleIdx, fromGroup, groupIdx);
        }
        setDraggingGroupKey('');
    };

    const destGroupOp = (op: 'labels' | 'remove', ruleIdx: number, idx: number) => {
        switch (op) {
            case 'labels':
                setTagEdit({
                    visible: true,
                    ruleIdx,
                    groupIdx: idx,
                    tags: Object.entries(customRouteRule.routing_config?.rules[ruleIdx].destinations[idx].labels || {}).map(([key, value]) => ({
                        key: key,
                        value: {
                            type: value.type || MatchType.EXACT,
                            value: value.value || '',
                            value_type: MatchValueType.TEXT
                        }
                    }))
                });
                break;
            case 'remove':
                removeGroup(ruleIdx, idx);
                break;
            default:
                break;
        }
    }

    function updateTagEdit(idx: number, tag: RoutingLabel) {
        const tags = [...tagEdit.tags];
        tags[idx] = tag;
        setTagEdit({ ...tagEdit, tags: tags });
    }

    function addTagEdit() {
        const newTag: RoutingLabel = {
            key: '',
            value: {
                type: MatchType.EXACT,
                value: '',
                value_type: MatchValueType.TEXT
            }
        };
        const tags = [...tagEdit.tags, newTag];
        setTagEdit({ ...tagEdit, tags: tags });
    }

    function removeTagEdit(idx: number) {
        if (tagEdit.tags.length <= 1) return;
        const tags = tagEdit.tags.filter((_, i) => i !== idx);
        setTagEdit({ ...tagEdit, tags: tags });
    }

    const routeMetadataRecord = React.useMemo(
        () => (customRouteRule.metadata || []).reduce<Record<string, string>>((acc, cur) => {
            if (cur.key) acc[cur.key] = cur.value;
            return acc;
        }, {}),
        [customRouteRule.metadata],
    );

    const routeNamespaceOptions = [{ label: '*', value: '*' }, ...namespaceDatas.map((ns) => ({ label: ns.name, value: ns.name }))];
    const callerServiceOptions = serviceDatas
        .filter((opt) => customRouteRule.caller_namespace === '*' || opt.namespace === customRouteRule.caller_namespace)
        .map((opt) => ({ label: opt.name, value: opt.name }));
    const calleeServiceOptions = serviceDatas
        .filter((opt) => customRouteRule.callee_namespace === '*' || opt.namespace === customRouteRule.callee_namespace)
        .map((opt) => ({ label: opt.name, value: opt.name }));

    const activeValidationErrors = validationErrors.length > 0 ? validationErrors : liveValidationErrors;

    const serviceSummary = `${customRouteRule.caller_namespace || '*'}/${customRouteRule.caller_service || '-'} → ${customRouteRule.callee_namespace || '*'}/${customRouteRule.callee_service || '-'}`;

    const renderSectionHeader = (order: number, title: string, description: string, extra?: React.ReactNode, onClick?: () => void) => (
        <div className={styles.designSectionHeader} onClick={onClick}>
            <div className={styles.designSectionTitleWrap}>
                <span className={styles.designSectionNumber}>{order}</span>
                <div>
                    <div className={styles.designSectionTitle}>{title}</div>
                    <div className={styles.designSectionDesc}>{description}</div>
                </div>
            </div>
            {extra && <div className={styles.designSectionExtra}>{extra}</div>}
        </div>
    );

    const renderTextWithPopup = (text: string) => (
        <Popup trigger="hover" content={text || '-'}>
            <span className={styles.ellipsisText}>{text || '-'}</span>
        </Popup>
    );

    const renderServiceScopeHeader = (
        <div className={`${styles.designSectionHeader} ${styles.serviceScopeHeader} ${serviceCollapsed ? styles.serviceScopeCollapsedHeader : ''}`} onClick={() => setServiceCollapsed(prev => !prev)}>
            <div className={styles.serviceScopeTitleWrap}>
                <span className={styles.designSectionNumber}>2</span>
                <div className={styles.serviceScopeTitleLine}>
                    <span className={styles.designSectionTitle}>服务范围</span>
                    <Popup trigger="hover" content={serviceCollapsed ? serviceSummary : '主调方 → 被调方'}>
                        <span className={`${styles.serviceScopeHint} ${serviceCollapsed ? styles.serviceScopeSummary : ''}`}>
                            {serviceCollapsed ? serviceSummary : '主调方 → 被调方'}
                        </span>
                    </Popup>
                </div>
            </div>
            <button type="button" className={`${styles.caretButton} ${serviceCollapsed ? '' : styles.caretButtonOpen}`} onClick={(event) => { event.stopPropagation(); setServiceCollapsed(prev => !prev); }}>
                <ChevronRightIcon />
            </button>
        </div>
    );

    const renderBaseInfoSection = (
        <div className={styles.designSection}>
            {renderSectionHeader(1, '基础信息', '规则元数据，不包含主调和被调服务范围')}
            <div className={styles.baseGrid}>
                <Row gutter={[16, 14]}>
                    <Col span={8}>
                        <div className={`${styles.fieldBlock} ${styles.fieldBlockWide}`}>
                            <div className={styles.fieldLabel}>规则名称</div>
                            {op === 'create' || editorState.editable
                                ? <Input maxlength={64} value={customRouteRule.name} placeholder="例如：route-by-tenant" onChange={(value) => setCustomRouteRule(prev => ({ ...prev, name: value }))} />
                                : <div className={styles.fieldValue}>{customRouteRule.name || '-'}</div>}
                            <div className={styles.fieldHint}>仅支持 kebab-case，用于生成 RouteRule 资源名</div>
                        </div>
                    </Col>
                    <Col span={2}>
                        <div className={styles.fieldBlock}>
                            <div className={styles.fieldLabel}>优先级</div>
                            {editorState.editable
                                ? <InputNumber className={styles.priorityInput} min={0} max={100} step={1} value={customRouteRule.priority ?? 5} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, priority: value as number }))} />
                                : <div className={styles.fieldValue}>{customRouteRule.priority ?? 5}</div>}
                        </div>
                    </Col>
                    <Col span={2}>
                        <div className={`${styles.fieldBlock} ${styles.statusField}`}>
                            <div className={styles.fieldLabel}>状态</div>
                            {editorState.editable
                                ? <div className={styles.statusSwitchWrap}><Switch label={['启用', '停用']} value={customRouteRule.enable !== false} onChange={(v) => setCustomRouteRule(prev => ({ ...prev, enable: v as boolean }))} /></div>
                                : <div className={styles.fieldValue}><span className={`${shared.pill} ${customRouteRule.enable === false ? shared.pillOff : shared.pillOk} ${shared.pillDot}`}>{customRouteRule.enable === false ? '停用' : '启用'}</span></div>}
                        </div>
                    </Col>
                    <Col span={12}>
                        <div className={styles.fieldBlock}>
                            <div className={styles.fieldLabel}>描述</div>
                            {editorState.editable
                                ? <Input value={customRouteRule.description} maxlength={255} placeholder="描述路由策略的业务意图" onChange={(value) => setCustomRouteRule(prev => ({ ...prev, description: value }))} />
                                : <div className={styles.fieldValue}>{customRouteRule.description || '-'}</div>}
                        </div>
                    </Col>
                    <Col span={12}>
                        <div className={styles.fieldBlock}>
                            <div className={styles.fieldLabel}>标签</div>
                            <RuleLabelField
                                metadata={routeMetadataRecord}
                                editable={editorState.editable}
                                onChange={(next) => setCustomRouteRule(prev => ({ ...prev, metadata: Object.entries(next).map(([key, value]) => ({ key, value })) }))}
                            />
                        </div>
                    </Col>
                </Row>
            </div>
        </div>
    );

    const renderServiceCard = (type: 'caller' | 'callee') => {
        const isCaller = type === 'caller';
        const namespaceValue = isCaller ? customRouteRule.caller_namespace : customRouteRule.callee_namespace;
        const serviceValue = isCaller ? customRouteRule.caller_service : customRouteRule.callee_service;
        const serviceOptions = isCaller ? callerServiceOptions : calleeServiceOptions;
        const updateNamespace = (value: string) => {
            setCustomRouteRule(prev => ({
                ...prev,
                [isCaller ? 'caller_namespace' : 'callee_namespace']: value,
            }));
        };
        const updateService = (value: string) => {
            setCustomRouteRule(prev => ({
                ...prev,
                [isCaller ? 'caller_service' : 'callee_service']: value,
            }));
        };
        return (
            <div className={`${styles.serviceCard} ${isCaller ? styles.serviceCaller : styles.serviceCallee}`}>
                <div className={styles.serviceCardHead}>
                    <span className={styles.serviceDot} />
                    <div>
                        <div className={styles.serviceTitle}>{isCaller ? '主调' : '被调'}</div>
                        <div className={styles.serviceSubtitle}>{isCaller ? '发起调用方' : '目标服务方'}</div>
                    </div>
                </div>
                {editorState.editable ? (
                    <div className={styles.serviceFields}>
                        <div>
                            <div className={styles.fieldLabel}>命名空间</div>
                            <Select filterable creatable options={routeNamespaceOptions} value={namespaceValue} onChange={(value) => updateNamespace(value as string)} />
                        </div>
                        <div>
                            <div className={styles.fieldLabel}>服务</div>
                            <Select filterable creatable options={serviceOptions} value={serviceValue} onChange={(value) => updateService(value as string)} />
                        </div>
                    </div>
                ) : (
                    <div className={styles.serviceReadonly}>
                        {renderTextWithPopup(namespaceValue || '*')}
                        <span>/</span>
                        {renderTextWithPopup(serviceValue || '-')}
                    </div>
                )}
            </div>
        );
    };

    const renderServiceScopeSection = (
        <div className={styles.designSection}>
            {renderServiceScopeHeader}
            {!serviceCollapsed && (
                <div className={styles.serviceFlowGrid}>
                    {renderServiceCard('caller')}
                    <div className={styles.flowConnector}><SendIcon /></div>
                    {renderServiceCard('callee')}
                </div>
            )}
        </div>
    );

    // 统一基础信息区：名称 / 状态 / 优先级 / 作用对象（主调→被调）/ 描述 / 标签
    const ruleeditor = (
        <div className={shared.section}>
            <div className={shared.sectionHeader}>基础信息</div>
            <div className={shared.sectionBody}>
                <div className={shared.infoGrid}>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>规则名称</div>
                        {op === 'create'
                            ? <Input maxlength={64} value={customRouteRule.name} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, name: value }))} />
                            : <div className={shared.fieldValue}>{customRouteRule.name || '-'}</div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>运行状态</div>
                        {editorState.editable
                            ? <Switch label={['启用', '停用']} value={customRouteRule.enable !== false} onChange={(v) => setCustomRouteRule(prev => ({ ...prev, enable: v as boolean }))} />
                            : <div className={shared.fieldValue}><span className={`${shared.pill} ${customRouteRule.enable === false ? shared.pillOff : shared.pillOk} ${shared.pillDot}`}>{customRouteRule.enable === false ? '停用' : '启用'}</span></div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>优先级</div>
                        {editorState.editable
                            ? <InputNumber min={0} value={customRouteRule.priority} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, priority: value as number }))} />
                            : <div className={shared.fieldValue}>{customRouteRule.priority ?? 0}</div>}
                    </div>
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>作用对象</div>
                        {editorState.editable ? (
                            <div className={shared.kv2}>
                                <div>
                                    <div className={shared.editLabel}>主调命名空间</div>
                                    <Select filterable creatable options={routeNamespaceOptions} value={customRouteRule.caller_namespace} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, caller_namespace: value as string }))} />
                                </div>
                                <div>
                                    <div className={shared.editLabel}>主调服务</div>
                                    <Select filterable creatable options={callerServiceOptions} value={customRouteRule.caller_service} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, caller_service: value as string }))} />
                                </div>
                                <div>
                                    <div className={shared.editLabel}>被调命名空间</div>
                                    <Select filterable creatable options={routeNamespaceOptions} value={customRouteRule.callee_namespace} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, callee_namespace: value as string }))} />
                                </div>
                                <div>
                                    <div className={shared.editLabel}>被调服务</div>
                                    <Select filterable creatable options={calleeServiceOptions} value={customRouteRule.callee_service} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, callee_service: value as string }))} />
                                </div>
                            </div>
                        ) : (
                            <div className={shared.flow}>
                                <div className={shared.flowNode}>
                                    <span className={shared.flowNodeLabel}>主调</span>
                                    <span className={shared.flowNodeValue}>{`${customRouteRule.caller_namespace || '-'} / ${customRouteRule.caller_service || '-'}`}</span>
                                </div>
                                <span className={shared.flowArrow}><SendIcon /></span>
                                <div className={shared.flowNode}>
                                    <span className={shared.flowNodeLabel}>被调</span>
                                    <span className={shared.flowNodeValue}>{`${customRouteRule.callee_namespace || '-'} / ${customRouteRule.callee_service || '-'}`}</span>
                                </div>
                            </div>
                        )}
                    </div>
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>描述</div>
                        {editorState.editable
                            ? <Input value={customRouteRule.description} maxlength={255} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, description: value }))} />
                            : <div className={shared.fieldValue}>{customRouteRule.description || '-'}</div>}
                    </div>
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>规则标签</div>
                        <RuleLabelField
                            metadata={routeMetadataRecord}
                            editable={editorState.editable}
                            onChange={(next) => setCustomRouteRule(prev => ({ ...prev, metadata: Object.entries(next).map(([key, value]) => ({ key, value })) }))}
                        />
                    </div>
                </div>
            </div>
        </div>
    );

    const trafficTableColumns = (ruleIdx: number): PrimaryTableProps['columns'] => [
        {
            colKey: 'type',
            title: '参数类型',
            width: 178,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: {
                    clearable: true,
                    options: RoutingArgumentsTypeOptions,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(false, ruleIdx, context.rowIndex, context.newRowData as RoutingSourceArgument);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => <Text>{RouteArgumentTextMap[row.type]}</Text>
        },
        {
            colKey: 'key',
            title: '参数键',
            width: 170,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Input,
                props: {
                    clearable: true,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(false, ruleIdx, context.rowIndex, context.newRowData as RoutingSourceArgument);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
        },
        {
            colKey: 'value.type',
            title: '匹配类型',
            width: 140,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: {
                    clearable: true,
                    options: MatchTypeOption,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(false, ruleIdx, context.rowIndex, context.newRowData as RoutingSourceArgument);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => <Text>{MatchTypeMap[row.value.type as MatchType]}</Text>
        },
        {
            colKey: 'value.value',
            title: '匹配值',
            minWidth: 180,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Input,
                props: {
                    clearable: true,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(false, ruleIdx, context.rowIndex, context.newRowData as RoutingSourceArgument);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => <span className={styles.wrapText}>{row.value?.value || '-'}</span>
        },
        {
            colKey: 'action',
            title: '操作',
            width: 64,
            cell: ({ row, rowIndex }) => (
                <Popup trigger="hover" content="删除参数">
                    <Button
                        shape="circle"
                        variant="text"
                        onClick={() => {
                            updateMatchArgs(true, ruleIdx, rowIndex, undefined);
                        }}>
                        <CloseIcon />
                    </Button>
                </Popup>

            ),
        }
    ]

    // 匹配条件表格（单行参数填写）
    const renderMatchTable = (ruleIdx: number) => (
        <div className={styles.compactTable}>
            <Table
                key={`route-match-${ruleIdx}-${editorState.editable ? 'edit' : 'view'}`}
                rowKey="key"
                tableLayout="fixed"
                data={customRouteRule.routing_config?.rules[ruleIdx].sources[0]?.arguments || []}
                columns={editorState.editable ? trafficTableColumns(ruleIdx) : trafficTableColumns(ruleIdx)?.filter(col => col.colKey !== 'action')}
            />
            {editorState.editable && (
                <Button className={styles.inlineAdd} variant="text" onClick={() => addMatch(ruleIdx)} icon={<AddIcon />}>添加匹配条件</Button>
            )}
        </div>
    );

    // 标签弹窗渲染
    const renderTagDialog = () => (
        <Dialog
            visible={tagEdit.visible}
            header="编辑实例标签"
            onClose={() => setTagEdit({ ...tagEdit, tags: [], visible: false })}
            onConfirm={() => {
                if (tagEdit.ruleIdx >= 0 && tagEdit.groupIdx >= 0) {
                    const newRules = cloneDeep(customRouteRule.routing_config?.rules || []);
                    newRules[tagEdit.ruleIdx].destinations[tagEdit.groupIdx].labels = tagEdit.tags.reduce((acc, tag) => {
                        if (tag.key && tag.value.value) {
                            acc[tag.key] = {
                                type: tag.value.type || MatchType.EXACT,
                                value: tag.value.value,
                                value_type: tag.value.value_type || RoutingValueType.TEXT
                            };
                        }
                        return acc;
                    }, {} as Record<string, MatchString>);
                    // 更新规则
                    setCustomRouteRule((prev) => ({
                        ...prev,
                        routing_config: {
                            ...(prev.routing_config || {}),
                            '@type': prev.routing_config?.['@type'] || customRouteConfigType,
                            rules: newRules,
                        },
                    }));
                }
                // 关闭弹窗后，需要清理掉标签数据
                setTagEdit({ ...tagEdit, visible: false, tags: [] });
            }}
            confirmBtn="确认"
            cancelBtn="取消"
            width={700}
        >
            <div>
                {tagEdit.tags.map((tag, idx) => (
                    <Row gutter={8} key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Col span={4}>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => updateTagEdit(idx, { ...tag, key: v })} />
                        </Col>
                        <Col span={2}>
                            <Select
                                value={tag.value.type || MatchType.EXACT}
                                options={MatchTypeOption}
                                style={{ width: '100%' }}
                                onChange={v => updateTagEdit(idx, { ...tag, value: { ...tag.value, type: v as string } })} />
                        </Col>
                        <Col span={4}>
                            {isTagInputMatchType(tag.value.type) ? (
                                <TagInput
                                    value={commaStringToTags(tag.value.value)}
                                    placeholder="请输入多个标签值，回车分隔"
                                    style={{ width: '100%' }}
                                    onChange={(value) => updateTagEdit(idx, {
                                        ...tag,
                                        value: { ...tag.value, value: tagsToCommaString(value as Array<string | number>) },
                                    })}
                                />
                            ) : (
                                <Input
                                    value={tag.value.value}
                                    placeholder="请输入标签值"
                                    onChange={v => updateTagEdit(idx, { ...tag, value: { ...tag.value, value: v } })} />
                            )}
                        </Col>
                        <Col span={2}>
                            {Object.keys(tagEdit.tags).length > 1 && (
                                <Popup trigger="hover" content="删除标签">
                                    <Button
                                        shape="circle"
                                        variant="text"
                                        onClick={() => removeTagEdit(idx)}
                                    >
                                        <CloseIcon />
                                    </Button>
                                </Popup>
                            )}
                        </Col>
                    </Row>
                ))}
                <Button variant="text" icon={<AddIcon />} onClick={addTagEdit}>添加标签</Button>
            </div>
        </Dialog>
    );

    const destinationTableColumns = (ruleIdx: number): PrimaryTableProps['columns'] => [
        ...(editorState.editable ? [{
            colKey: '__drag__',
            title: '',
            width: 44,
            cell: ({ rowIndex }) => {
                const groupIdx = rowIndex as number;
                const dragKey = `${ruleIdx}:${groupIdx}`;
                return (
                    <Popup trigger="hover" content="拖动调整分组顺序">
                        <button
                            type="button"
                            className={`${styles.dragHandle} ${draggingGroupKey === dragKey ? styles.dragHandleActive : ''}`}
                            draggable
                            onDragStart={(event) => handleGroupDragStart(event, ruleIdx, groupIdx)}
                            onDragEnd={() => setDraggingGroupKey('')}
                            onDragOver={(event) => event.preventDefault()}
                            onDrop={(event) => handleGroupDrop(event, ruleIdx, groupIdx)}
                        >
                            <DragMoveIcon />
                        </button>
                    </Popup>
                );
            },
        }] : []),
        {
            colKey: 'name',
            title: '分组名称',
            width: 142,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Input,
                props: {
                    clearable: true,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateGroup(ruleIdx, context.rowIndex, context.newRowData as RoutingRuleDestination);
                },
            }
        },
        {
            colKey: 'isolate',
            title: '是否隔离',
            width: 96,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Switch,
                props: {
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateGroup(ruleIdx, context.rowIndex, context.newRowData as RoutingRuleDestination);
                },
            },
            cell: ({ row }) => (
                <Text>{row.isolate ? '是' : '否'}</Text>
            )
        },
        {
            colKey: 'weight',
            title: '权重',
            width: 166,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: InputNumber,
                props: {
                    min: 0,
                    max: 100,
                    step: 5,
                    className: styles.weightInput,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateGroup(ruleIdx, context.rowIndex, context.newRowData as RoutingRuleDestination);
                },
            }
        },
        {
            colKey: 'labels',
            title: '实例标签',
            minWidth: 200,
            cell: ({ row, rowIndex }) => (
                <div className={`${styles.tagList} ${styles.instanceTagList}`}>
                    {Object.entries(row.labels as Record<string, MatchString> || {}).map(([key, value]) => (
                        <Tag key={key}>
                            {`${key} ${MatchTypeMap[value.type as MatchType]} ${value.value}`}
                        </Tag>
                    ))}
                </div>
            )
        },
        {
            colKey: 'action',
            title: '操作',
            width: 72,
            cell: ({ row, rowIndex }) => (
                <Space size={4}>
                    <Popup trigger="hover" content="编辑标签">
                        <Button
                            shape="circle"
                            variant="text"
                            onClick={() => destGroupOp('labels', ruleIdx, rowIndex)}>
                            <Edit1Icon />
                        </Button>
                    </Popup>
                    <Popup trigger="hover" content="删除分组">
                        <Button
                            shape="circle"
                            variant="text"
                            onClick={() => destGroupOp('remove', ruleIdx, rowIndex)}>
                            <CloseIcon />
                        </Button>
                    </Popup>
                </Space>
            ),
        }
    ]

    // 路由策略分组表格（弹窗编辑标签）
    const renderGroupTable = (ruleIdx: number) => (
        <div className={styles.compactTable}>
            <Table
                key={`route-destination-${ruleIdx}-${editorState.editable ? 'edit' : 'view'}`}
                rowKey="key"
                tableLayout="fixed"
                data={customRouteRule.routing_config?.rules[ruleIdx].destinations || []}
                columns={editorState.editable ? destinationTableColumns(ruleIdx) : destinationTableColumns(ruleIdx)?.filter(col => col.colKey !== 'action')}
            />
            {editorState.editable && (
                <Button
                    className={styles.inlineAdd}
                    variant="text"
                    onClick={() => addGroup(ruleIdx)}
                    icon={<AddIcon />}>
                    添加实例分组
                </Button>
            )}
            {renderTagDialog()}
        </div>
    );

    const renderMatchModeSwitch = (rule: RoutingRule, ruleIdx: number) => {
        const mode = (rule.arguments?.matchMode || 'AND') as 'AND' | 'OR';
        return (
            <div className={styles.segmented}>
                {(['AND', 'OR'] as const).map((item) => (
                    <button
                        key={item}
                        type="button"
                        className={`${styles.segmentButton} ${mode === item ? styles.segmentButtonActive : ''}`}
                        disabled={!editorState.editable}
                        onClick={() => updateRuleMatchMode(ruleIdx, item)}
                    >
                        {item}
                    </button>
                ))}
            </div>
        );
    };

    const renderMatchRows = (rule: RoutingRule, ruleIdx: number) => {
        const rows = getRouteRuleArguments(rule);
        return (
            <div className={styles.conditionPanel}>
                <div className={styles.conditionPanelHead}>
                    <div>
                        <div className={styles.subPanelTitle}>匹配条件</div>
                        <div className={styles.subPanelHint}>同一规则内条件按 {rule.arguments?.matchMode || 'AND'} 关系计算</div>
                    </div>
                    {renderMatchModeSwitch(rule, ruleIdx)}
                </div>
                <div className={styles.matchGrid}>
                    <div className={styles.matchHeader}>参数类型</div>
                    <div className={styles.matchHeader}>参数键</div>
                    <div className={styles.matchHeader}>匹配类型</div>
                    <div className={styles.matchHeader}>匹配值</div>
                    <div className={styles.matchHeader}>操作</div>
                    {rows.map((arg, idx) => (
                        <React.Fragment key={`${ruleIdx}-match-${idx}`}>
                            <div className={styles.matchCell}>
                                {editorState.editable ? (
                                    <Select
                                        filterable
                                        options={RoutingArgumentsTypeOptions}
                                        value={arg.type}
                                        onChange={(value) => updateMatchArgs(false, ruleIdx, idx, withParamTypeDefaultKey(arg, value as string) as RoutingSourceArgument)}
                                    />
                                ) : <span>{RouteArgumentTextMap[arg.type as RoutingArgumentsType] || arg.type}</span>}
                            </div>
                            <div className={styles.matchCell}>
                                {editorState.editable ? (
                                    <Input
                                        placeholder="请输入参数键"
                                        value={arg.key}
                                        onChange={(value) => updateMatchArgs(false, ruleIdx, idx, { ...arg, key: value as string })}
                                    />
                                ) : renderTextWithPopup(arg.key)}
                            </div>
                            <div className={styles.matchCell}>
                                {editorState.editable ? (
                                    <Select
                                        filterable
                                        options={MatchTypeOption}
                                        value={arg.value?.type || MatchType.EXACT}
                                        onChange={(value) => updateMatchArgs(false, ruleIdx, idx, { ...arg, value: { ...arg.value, type: value as MatchType } })}
                                    />
                                ) : <span>{MatchTypeMap[arg.value?.type as MatchType] || arg.value?.type || '-'}</span>}
                            </div>
                            <div className={styles.matchCell}>
                                {editorState.editable ? (
                                    isTagInputMatchType(arg.value?.type) ? (
                                        <TagInput
                                            value={commaStringToTags(arg.value?.value)}
                                            onChange={(value) => updateMatchArgs(false, ruleIdx, idx, {
                                                ...arg,
                                                value: { ...arg.value, value: tagsToCommaString(value as Array<string | number>) },
                                            })}
                                            placeholder="请输入多个值，回车分隔"
                                        />
                                    ) : (
                                        <Input
                                            value={arg.value?.value}
                                            onChange={(value) => updateMatchArgs(false, ruleIdx, idx, { ...arg, value: { ...arg.value, value } })}
                                        />
                                    )
                                ) : renderTextWithPopup(arg.value?.value || '-')}
                            </div>
                            <div className={`${styles.matchCell} ${styles.actionCell}`}>
                                {editorState.editable && (
                                    <Popup trigger="hover" content={rows.length <= 1 ? '至少保留一个条件' : '删除条件'}>
                                        <Button shape="circle" variant="text" disabled={rows.length <= 1} onClick={() => updateMatchArgs(true, ruleIdx, idx)}>
                                            <CloseIcon />
                                        </Button>
                                    </Popup>
                                )}
                            </div>
                        </React.Fragment>
                    ))}
                </div>
                {editorState.editable && (
                    <Button className={styles.inlineAdd} variant="text" onClick={() => addMatch(ruleIdx)} icon={<AddIcon />}>添加匹配条件</Button>
                )}
            </div>
        );
    };

    const renderWeightMeter = (rule: RoutingRule) => {
        const status = getWeightStatus(rule);
        const destinations = rule.destinations || [];
        return (
            <div className={styles.weightMeterWrap}>
                <div className={styles.weightMeter}>
                    {destinations.map((group, idx) => (
                        <div
                            key={`${group.name}-${idx}`}
                            className={styles.weightSegment}
                            style={{ width: `${Math.max(0, Math.min(Number(group.weight || 0), 100))}%` }}
                        />
                    ))}
                </div>
                <span className={status.ok ? styles.weightOk : styles.weightError}>{status.message}</span>
            </div>
        );
    };

    const renderGroupRows = (rule: RoutingRule, ruleIdx: number) => {
        const destinations = rule.destinations || [];
        const renderGroupLabelCell = (group: RoutingRuleDestination, idx: number) => {
            const labels = Object.entries(group.labels || {}).filter(([key, value]) => key && value?.value);
            return (
                <div className={`${shared.labelDisplay} ${styles.groupLabelDisplay}`}>
                    {labels.length ? labels.map(([key, value]) => (
                        <span className={shared.chip} key={`${key}-${value.value}`}>
                            <span className={shared.chipKey}>{key}</span>
                            <span className={shared.chipValue}>{value.value}</span>
                        </span>
                    )) : <span className={shared.emptyLine}>暂无实例标签</span>}
                    {editorState.editable && (
                        <button type="button" className={shared.editChipButton} onClick={() => destGroupOp('labels', ruleIdx, idx)}>
                            <Edit1Icon />编辑标签
                        </button>
                    )}
                </div>
            );
        };
        return (
            <div className={styles.groupPanel}>
                <div className={styles.conditionPanelHead}>
                    <div>
                        <div className={styles.subPanelTitle}>目标分组</div>
                        <div className={styles.subPanelHint}>命中后按权重转发，合计必须等于 100%</div>
                    </div>
                    {renderWeightMeter(rule)}
                </div>
                <div className={styles.groupGrid}>
                    <div className={styles.groupHeader}>分组名称</div>
                    <div className={styles.groupHeader}>隔离</div>
                    <div className={styles.groupHeader}>权重</div>
                    <div className={styles.groupHeader}>标签</div>
                    <div className={styles.groupHeader}>操作</div>
                    {destinations.map((group, idx) => {
                        const dragKey = `${ruleIdx}:${idx}`;
                        return (
                            <React.Fragment key={`${ruleIdx}-group-${idx}`}>
                                <div
                                    className={styles.groupCell}
                                    draggable={editorState.editable}
                                    onDragStart={(event) => handleGroupDragStart(event, ruleIdx, idx)}
                                    onDragEnd={() => setDraggingGroupKey('')}
                                    onDragOver={(event) => event.preventDefault()}
                                    onDrop={(event) => handleGroupDrop(event, ruleIdx, idx)}
                                >
                                    {editorState.editable && (
                                        <Popup trigger="hover" content="拖动调整分组顺序">
                                            <button type="button" className={`${styles.dragHandle} ${draggingGroupKey === dragKey ? styles.dragHandleActive : ''}`}>
                                                <DragMoveIcon />
                                            </button>
                                        </Popup>
                                    )}
                                    <span className={styles.groupSwatch} />
                                    <span className={styles.generatedGroupName}>{getAutoGroupName(idx)}</span>
                                </div>
                                <div className={styles.groupCell}>
                                    {editorState.editable
                                        ? <Switch value={group.isolate} onChange={(value) => updateGroup(ruleIdx, idx, { ...group, isolate: value as boolean })} />
                                        : <span>{group.isolate ? '是' : '否'}</span>}
                                </div>
                                <div className={styles.groupCell}>
                                    {editorState.editable
                                        ? <InputNumber className={styles.weightInput} min={0} max={100} step={5} value={group.weight} onChange={(value) => updateGroup(ruleIdx, idx, { ...group, weight: value as number })} />
                                        : <span>{group.weight}%</span>}
                                </div>
                                <div className={styles.groupCell}>
                                    {renderGroupLabelCell(group, idx)}
                                </div>
                                <div className={`${styles.groupCell} ${styles.actionCell}`}>
                                    {editorState.editable && (
                                        <Popup trigger="hover" content={destinations.length <= 1 ? '至少保留一个分组' : '删除分组'}>
                                            <Button shape="circle" variant="text" disabled={destinations.length <= 1} onClick={() => removeGroup(ruleIdx, idx)}>
                                                <CloseIcon />
                                            </Button>
                                        </Popup>
                                    )}
                                </div>
                            </React.Fragment>
                        );
                    })}
                </div>
                {editorState.editable && (
                    <Button
                        className={styles.inlineAdd}
                        variant="text"
                        onClick={() => addGroup(ruleIdx)}
                        icon={<AddIcon />}>
                        添加目标分组
                    </Button>
                )}
            </div>
        );
    };

    // 规则区块
    const renderRule = (rule: RoutingRule, ruleIdx: number) => {
        const collapsed = collapsedRuleIndexes.has(ruleIdx);
        const matchCount = getRouteRuleArguments(rule).length;
        const groupCount = rule.destinations?.length || 0;
        const weightStatus = getWeightStatus(rule);
        return (
            <div
                className={`${shared.policy} ${collapsed ? shared.policyCollapsed : ''} ${draggingRuleIndex === ruleIdx ? styles.ruleBlockDragging : ''}`}
                key={ruleIdx}
                onDragOver={(event) => event.preventDefault()}
                onDrop={(event) => handleRuleDrop(event, ruleIdx)}
            >
                <div className={shared.policyHead} onClick={() => toggleRuleCollapsed(ruleIdx)}>
                    <div className={shared.policyHeadMain}>
                        {editorState.editable && (
                            <Popup trigger="hover" content="拖动调整子规则顺序">
                                <button
                                    type="button"
                                    className={`${styles.dragHandle} ${draggingRuleIndex === ruleIdx ? styles.dragHandleActive : ''}`}
                                    draggable
                                    onClick={(e) => e.stopPropagation()}
                                    onDragStart={(event) => handleRuleDragStart(event, ruleIdx)}
                                    onDragEnd={() => setDraggingRuleIndex(null)}
                                >
                                    <DragMoveIcon />
                                </button>
                            </Popup>
                        )}
                        <span className={shared.caret}><ChevronRightIcon /></span>
                        <div>
                            <div className={shared.policyIndex}>规则 [{ruleIdx + 1}]</div>
                            <div className={shared.policySummary}>
                                {matchCount} 个匹配条件 / {groupCount} 个目标分组 / 权重 {weightStatus.total}%
                            </div>
                        </div>
                    </div>
                    <div className={shared.policyHeadActions} onClick={(e) => e.stopPropagation()}>
                        {editorState.editable && (
                            <Popup trigger="hover" content="删除规则">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => removeRule(ruleIdx)}>
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        )}
                    </div>
                </div>
                <div className={shared.policyBody}>
                    <div className={shared.step} data-step="1">
                        <div className={shared.stepContent}>{renderMatchRows(rule, ruleIdx)}</div>
                    </div>
                    <div className={shared.step} data-step="2">
                        <div className={shared.stepContent}>{renderGroupRows(rule, ruleIdx)}</div>
                    </div>
                </div>
            </div>
        );
    }

    const renderPublishForm = (
        <>
            {editorState.publishView && (
                <PublishForm
                    ruleId={customRouteRule.id || ''}
                    ruleName={customRouteRule.name || ''}
                    resource={PolicySourceType.RouteRules}
                    visible={editorState.publishView}
                    close={() => {
                        setEditorState(prev => ({ ...prev, publishView: false }));
                    }}
                />
            )}
        </>
    )

    const renderStickyTool = (
        <>
            {editable && (
                <FormItem style={{ marginTop: 20 }}>
                    <StickyTool
                        style={{ zIndex: 1000 }}
                        placement='right-bottom'
                        offset={[-10, 200]}
                    >
                        <StickyItem
                            label=""
                            icon={!editorState.editable ?
                                <RuleStickyAction label="编辑" icon={<Edit1Icon />} onClick={() => {
                                    setEditorState(prev => ({ ...prev, editable: true }));
                                }} />
                                :
                                <RuleStickyAction label="保存" icon={<SaveIcon />} onClick={() => {
                                    form.submit();
                                }} />
                            }
                        />
                        {(editorState.editable) && (
                            <StickyItem label="" icon={
                                <RuleStickyAction label="撤销" icon={<RollbackIcon />} onClick={() => {
                                    if (op === 'create') {
                                        refresh(true);
                                    } else {
                                        resetCurRule(viewRoute);
                                    }
                                    setEditorState(prev => ({ ...prev, editable: false }));
                                }} />}
                            />
                        )}
                        {(!editorState.editable) && (
                            <StickyItem label="" icon={
                                <RuleStickyAction label="发布" icon={<RocketIcon />} onClick={() => {
                                    setEditorState(prev => ({ ...prev, publishView: true }));
                                }} />}
                            />
                        )}
                    </StickyTool>
                </FormItem>
            )}
        </>
    )

    const copyPreviewText = async () => {
        try {
            await navigator.clipboard.writeText(previewText);
            openInfoNotification('已复制', 'RouteRule Spec 已复制到剪贴板');
        } catch (err) {
            openErrNotification('复制失败', '当前浏览器不允许访问剪贴板');
        }
    };

    const renderSpecPreview = (
        <aside className={styles.specPane}>
            <div className={styles.specCard}>
                <div className={styles.specToolbar}>
                    <div>
                        <div className={styles.specTitle}>实时规则 SPEC</div>
                        <div className={styles.specDesc}>保存前可核对资源形态</div>
                    </div>
                    <div className={styles.specActions}>
                        <div className={styles.specToggle}>
                            {(['yaml', 'json'] as RouteSpecFormat[]).map((item) => (
                                <button
                                    type="button"
                                    key={item}
                                    className={specFormat === item ? styles.specToggleActive : ''}
                                    onClick={() => setSpecFormat(item)}
                                >
                                    {item.toUpperCase()}
                                </button>
                            ))}
                        </div>
                        <Button size="small" variant="outline" onClick={copyPreviewText}>复制</Button>
                    </div>
                </div>
                <pre className={styles.specCode}>{previewText}</pre>
                <div className={styles.specFooter}>
                    {activeValidationErrors.length > 0
                        ? activeValidationErrors.slice(0, 3).map((item) => <span key={`${item.field}-${item.message}`} className={styles.previewError}>{item.message}</span>)
                        : <span className={styles.previewOk}>校验通过，可保存并下发</span>}
                </div>
            </div>
        </aside>
    );

    return (
        <div className={styles.editorBody}>
            <Form
                form={form}
                onSubmit={onSubmit}
                layout="vertical"
                colon
            >
                <div className={styles.routeEditorShell}>
                    <div className={styles.formPane}>
                        {renderBaseInfoSection}
                        {renderServiceScopeSection}
                        <div className={styles.designSection}>
                            {renderSectionHeader(
                                3,
                                '路由规则',
                                '按顺序匹配条件，并把命中的请求路由到目标分组',
                                <span className={styles.countTag}>{customRouteRule.routing_config?.rules.length || 0} 条</span>,
                            )}
                            <div className={shared.ruleList}>
                                {customRouteRule.routing_config?.rules.map((rule, idx) => renderRule(rule, idx))}
                            </div>
                            {editorState.editable && (
                                <Button
                                    className={shared.addRuleButton}
                                    style={{ marginTop: 12 }}
                                    variant="dashed" icon={<AddIcon />}
                                    onClick={() => {
                                        updateRoutingRules((rules) => [
                                            ...rules,
                                            {
                                                name: `规则 ${rules.length + 1}`,
                                                sources: [defaultMatch()],
                                                arguments: {
                                                    arguments: [],
                                                    randomPercent: 0,
                                                    matchMode: 'AND',
                                                },
                                                destinations: [defaultGroup(1)],
                                            },
                                        ]);
                                    }}>
                                    添加规则
                                </Button>
                            )}
                        </div>
                    </div>
                    {renderSpecPreview}
                </div>
                {renderPublishForm}
                {renderTagDialog()}
                {renderStickyTool}
            </Form>
        </div>
    );
};

export default React.memo(CustomRouteEditor);
