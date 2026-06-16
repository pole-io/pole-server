import React from "react";
import { Col, Form, Input, Row, Space, Button, Select, Switch, Dialog, InputNumber, Table, FormProps, Tag, Popup, TableRowData, PrimaryTableProps, StickyTool } from "tdesign-react";
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
    buildRoutingConfigForApi,
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

const { FormItem } = Form;
const { StickyItem } = StickyTool;

const defaultMatchArgs: () => RoutingSourceArgument = () => ({
    type: RoutingArgumentsType.HEADER,
    key: '',
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

const defaultGroup: (id: number) => RoutingRuleDestination = (id: number) => ({
    service: '',
    namespace: '',
    name: `分组 ${id}`,
    weight: 100,
    isolate: false,
    labels: {},
    priority: 0
});

export const defaultCustomRoute: () => CustomRouteDO = () => ({
    name: '',
    description: '',
    priority: 0,
    routing_policy: 'RulePolicy',
    metadata: [],
    enable: true,
    routing_config: {
        '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
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
    const [draggingRuleIndex, setDraggingRuleIndex] = React.useState<number | null>(null);
    const [draggingGroupKey, setDraggingGroupKey] = React.useState<string>('');

    const resetCurRule = (rule: CustomRouteView | null) => {
        if (!rule) {
            return;
        }
        const cloneRule = cloneDeep(rule);
        const routingConfig = normalizeRoutingConfigForEditor(cloneRule.routing_config);
        const firstRule = routingConfig?.rules?.[0];
        const caller = routingConfig?.caller;
        const callee = routingConfig?.callee;
        setCustomRouteRule({
            ...cloneRule,
            routing_config: routingConfig,
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
        const data: CustomRoute = {
            id: customRouteRule.id,
            name: customRouteRule.name || '',
            description: customRouteRule.description || '',
            enable: customRouteRule.enable,
            priority: customRouteRule.priority || 0,
            routing_config: buildRoutingConfigForApi(
                customRouteRule.routing_config,
                { namespace: customRouteRule.caller_namespace, service: customRouteRule.caller_service },
                { namespace: customRouteRule.callee_namespace, service: customRouteRule.callee_service },
            ),
            routing_policy: customRouteRule.routing_policy || 'RulePolicy',
            metadata: customRouteRule.metadata?.reduce<Record<string, string>>((acc, tag) => {
                acc[tag.key] = tag.value;
                return acc;
            }, {})
        }

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
                '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                rules: updater(cloneDeep(prev.routing_config?.rules || [])),
            },
        }));
    };

    const addMatch = (ruleIdx: number) => {
        updateRoutingRules((rules) => {
            rules[ruleIdx].sources[0].arguments.push(defaultMatchArgs());
            return rules;
        });
    };

    const updateMatchArgs = (del: boolean, ruleIdx: number, idx: number, args?: RoutingSourceArgument) => {
        updateRoutingRules((rules) => {
            if (del) {
                rules[ruleIdx].sources[0].arguments.splice(idx, 1);
            } else {
                if (!rules[ruleIdx].sources[0].arguments) {
                    rules[ruleIdx].sources[0].arguments = [];
                }
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

    const addGroup = (ruleIdx: number) => {
        updateRoutingRules((rules) => {
            const newGroup = defaultGroup(rules[ruleIdx].destinations.length + 1);
            rules[ruleIdx].destinations.push(newGroup);
            return rules;
        });
    };

    const removeGroup = (ruleIdx: number, idx: number) => {
        updateRoutingRules((rules) => {
            rules[ruleIdx].destinations.splice(idx, 1);
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
            rules[ruleIdx].destinations = destinations.map((group, idx) => ({
                ...group,
                priority: idx,
            }));
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
                // 编辑标签
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
                // 删除分组
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

    const routeNamespaceOptions = namespaceDatas.map((ns) => ({ label: ns.name, value: ns.name }));
    const callerServiceOptions = serviceDatas
        .filter((opt) => customRouteRule.caller_namespace === '*' || opt.namespace === customRouteRule.caller_namespace)
        .map((opt) => ({ label: opt.name, value: opt.name }));
    const calleeServiceOptions = serviceDatas
        .filter((opt) => customRouteRule.callee_namespace === '*' || opt.namespace === customRouteRule.callee_namespace)
        .map((opt) => ({ label: opt.name, value: opt.name }));

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
                            '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
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
                            <Input
                                value={tag.value.value}
                                placeholder="请输入标签值"
                                onChange={v => updateTagEdit(idx, { ...tag, value: { ...tag.value, value: v } })} />
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

    // 规则区块
    const renderRule = (rule: RoutingRule, ruleIdx: number) => {
        const collapsed = collapsedRuleIndexes.has(ruleIdx);
        const matchCount = rule.sources?.[0]?.arguments?.length || 0;
        const groupCount = rule.destinations?.length || 0;
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
                                {matchCount} 个匹配条件 / {groupCount} 个实例分组
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
                        <div className={shared.stepTitle}>匹配条件<span className={shared.stepHint}>来源服务的请求满足以下条件才命中本规则</span></div>
                        <div className={shared.stepContent}>{renderMatchTable(ruleIdx)}</div>
                    </div>
                    <div className={shared.step} data-step="2">
                        <div className={shared.stepTitle}>目标分组<span className={shared.stepHint}>命中后按权重转发至以下实例分组</span></div>
                        <div className={shared.stepContent}>{renderGroupTable(ruleIdx)}</div>
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

    return (
        <div className={styles.editorBody}>
            <Form
                form={form}
                onSubmit={onSubmit}
                layout="vertical"
                colon
            >
                {ruleeditor}
                <div className={shared.section}>
                    <div className={shared.sectionHeader}>
                        <span>路由规则</span>
                        <span className={shared.countTag}>{customRouteRule.routing_config?.rules.length || 0} 条</span>
                    </div>
                    <div className={shared.sectionBody}>
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
                                            destinations: [defaultGroup(1)],
                                        },
                                    ]);
                                }}>
                                添加规则
                            </Button>
                        )}
                    </div>
                </div>
                {renderPublishForm}
                {renderStickyTool}
            </Form>
        </div>
    );
};

export default React.memo(CustomRouteEditor);
