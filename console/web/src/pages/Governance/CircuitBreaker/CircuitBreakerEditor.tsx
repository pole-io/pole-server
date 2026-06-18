import React from "react";
import { Col, Form, Input, Row, Space, Button, Select, Switch, Dialog, InputNumber, Table, FormProps, Tag, Popup, TableRowData, PrimaryTableProps, StickyTool, RadioGroup, Radio, InputAdornment, Textarea } from "tdesign-react";
import { AddIcon, ChevronRightIcon, CloseIcon, Edit1Icon, SaveIcon, RocketIcon, RollbackIcon, RemoveIcon } from "tdesign-icons-react";

import Text from "components/Text";
import RuleLabelField from "../shared/RuleLabelField";
import shared from "../shared/governance.module.less";
import { useAppDispatch, useAppSelector } from 'modules/store';
import { API, HTTPMethodOption, InterfaceProtocolOption, Label, MatchType, MatchTypeMap, MatchTypeOption, MatchValueType, Op } from "services/types";
import { ServiceView } from "services/service";

import { NamespaceView } from "services/namespace";
import styles from './CircuitBreakerEditor.module.less';
import { openErrNotification, openInfoNotification } from "utils/notifition";
import PublishForm from "../RuleRelease/PublishForm";
import RuleStickyAction from "../RuleRelease/RuleStickyAction";
import { PolicySourceType } from "services/auth_policy";
import { BlockConfig, BreakLevelMap, BreakLevelType, CircuitBreakerRule, ErrorCondition, ErrorConditionMap, ErrorConditionOptions, ErrorConditionType, TriggerCondition, TriggerType, TriggerTypeMap, TriggerTypeOptions } from "services/circuitbreaker";
import { defaultBlockConfig, listOneCircuitBreaker, resetCircuitBreaker, saveCircuitBreakers, selectCircuitBreaker, updateCircuitBreakers } from "modules/governance/circuitbreaker";
import { cleanNamespacePage, listAllNamespaces, selectNamespace } from "modules/namespace";
import { cleanServicePage, listAllServices, selectService } from "modules/discovery/service";
import { clone, cloneDeep, set } from "lodash";

const { FormItem } = Form;
const { StickyItem } = StickyTool;

interface CircuitBreakerDO {
    id?: string
    name: string // 规则名
    level: string
    description: string
    priority: number
    ruleMatcher: {
        source: {
            service: string
            namespace: string
        }
        destination: {
            service: string
            namespace: string
            method: {
                type: string
                value: string
            }
        }
    }
    block_configs: BlockConfig[]
    metadata: Label[]
}

const defaultCircuitBreakerRule = (): CircuitBreakerDO => ({
    name: '',
    level: BreakLevelType.Service,
    priority: 0,
    description: '',
    metadata: [],
    ruleMatcher: {
        source: {
            service: '',
            namespace: '*'
        },
        destination: {
            service: '',
            namespace: '*',
            method: {
                type: MatchType.EXACT,
                value: ''
            }
        }
    },
    block_configs: [defaultBlockConfig(1)],
})

interface ICircuitBreakerEditorProps {
    op: Op;
    refresh: (close: boolean) => void;
}

const CircuitBreakerEditor: React.FC<ICircuitBreakerEditorProps> = ({ op, refresh }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { datas: serviceDatas } = serviceState;

    const breakerState = useAppSelector(selectCircuitBreaker);
    const { editRule, viewRule } = breakerState;

    const [editorState, setEditorState] = React.useState<{
        visible: boolean
        headerVisible: boolean
        headerRuleIndex?: number
        editable?: boolean
        model: Op;
        publishView: boolean;
    }>({ model: 'view', publishView: false, visible: false, headerVisible: false, editable: op === 'create' || false, });

    // 创建新的 rules 对象
    const [breakerRule, setBreakerRule] = React.useState<CircuitBreakerDO>(defaultCircuitBreakerRule());
    const [collapsedRuleIndexes, setCollapsedRuleIndexes] = React.useState<Set<number>>(() => new Set());

    const toggleRuleCollapsed = (ruleIdx: number) => {
        setCollapsedRuleIndexes(prev => {
            const next = new Set(prev);
            if (next.has(ruleIdx)) {
                next.delete(ruleIdx);
            } else {
                next.add(ruleIdx);
            }
            return next;
        });
    };

    const renderReadonlyValue = (value: React.ReactNode) => (
        <div className={styles.readonlyValue}>{value}</div>
    );

    const renderReadonlySwitch = (enabled?: boolean) => (
        <Tag theme={enabled ? 'success' : 'default'} variant="light">
            {enabled ? '开启' : '关闭'}
        </Tag>
    );

    const resetCurRule = (editRule: CircuitBreakerRule) => {
        if (!editRule) {
            return;
        }
        const cloneRule = cloneDeep(editRule);
        const sourceNamespace = cloneRule?.ruleMatcher?.source?.namespace || '*';
        let destinationNamespace = cloneRule?.ruleMatcher?.destination?.namespace || '*';
        let destinationService = cloneRule?.ruleMatcher?.destination?.service || '';
        const knownNamespaces = new Set((namespaceDatas || []).map((item: NamespaceView) => item.name));
        const swappedByOptions = knownNamespaces.has(destinationService)
            && serviceDatas.some((item: ServiceView) => item.namespace === destinationService && item.name === destinationNamespace);
        const swappedByNameShape = destinationService.includes('governance') && !destinationNamespace.includes('governance');
        if ((destinationService === sourceNamespace && destinationNamespace !== sourceNamespace) || swappedByOptions || swappedByNameShape) {
            [destinationNamespace, destinationService] = [destinationService, destinationNamespace];
        }
        setBreakerRule({
            ...cloneRule, // Override with editRule values
            name: cloneRule?.name || '',
            description: cloneRule?.description || '',
            priority: cloneRule?.priority || 0,
            level: cloneRule?.level || BreakLevelType.Service, // Add default value for level
            ruleMatcher: {
                source: {
                    service: cloneRule?.ruleMatcher?.source?.service || '',
                    namespace: cloneRule?.ruleMatcher?.source?.namespace || '*'
                },
                destination: {
                    service: destinationService,
                    namespace: destinationNamespace,
                    method: {
                        type: cloneRule?.ruleMatcher?.destination?.method?.type || MatchType.EXACT,
                        value: cloneRule?.ruleMatcher?.destination?.method?.value || ''
                    }
                }
            },
            metadata: Object.entries(cloneRule?.metadata || {}).map(([key, value]) => ({ key, value })),
        });
    }

    React.useEffect(() => {
        if (editRule) {
            if (editRule.id !== '') {
                // 发生变化，重置当前规则
                setEditorState(pre => ({ ...pre, visible: false, editable: false }));
                dispatch(listOneCircuitBreaker({ id: editRule.id || '' })).then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('请求失败', `获取熔断规则失败: ${res?.payload as string}`);
                    }
                    const { viewRule } = res.payload as { viewRule: CircuitBreakerRule };
                    resetCurRule(viewRule);
                })
            } else {
                // 新建规则，重置当前规则
                resetCurRule(editRule);
            }
        }
    }, [editRule])

    React.useEffect(() => {
        dispatch(listAllNamespaces())
            .then(res => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('获取命名空间列表失败', res?.payload as string);
                }
            })
        dispatch(listAllServices())
            .then(res => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('获取服务列表失败', res?.payload as string);
                }
            })

        return () => {
            dispatch(cleanNamespacePage());
            dispatch(cleanServicePage());
            dispatch(resetCircuitBreaker());
        }
    }, []);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log('提交数据:', breakerRule);
        let res;
        if (op !== 'create') {
            res = await dispatch(updateCircuitBreakers({
                param: {
                    ...breakerRule,
                    metadata: breakerRule.metadata?.reduce((acc, cur) => {
                        if (cur.key && cur.value) {
                            acc[cur.key] = cur.value;
                        }
                        return acc;
                    }, {} as Record<string, string>)
                }
            }));
        } else {
            res = await dispatch(saveCircuitBreakers({
                param: {
                    ...breakerRule,
                    metadata: breakerRule.metadata?.reduce((acc, cur) => {
                        if (cur.key && cur.value) {
                            acc[cur.key] = cur.value;
                        }
                        return acc;
                    }, {} as Record<string, string>)
                }
            }));
        }

        if (res.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', res?.payload as string);
        } else {
            openInfoNotification('请求成功', op !== 'create' ? '修改熔断规则成功' : '创建熔断规则成功');
            if (op !== 'create') {
                setEditorState(prev => ({ ...prev, editable: false }));
            }
            refresh(false); // 刷新列表
        }
    }

    const addErrorConditions = (ruleIdx: number) => {
        const newRules = { ...breakerRule };
        newRules.block_configs[ruleIdx].error_conditions.push({
            inputType: ErrorConditionType.RET_CODE,
            condition: {
                type: MatchType.EXACT,
                value: '',
            }
        });
        setBreakerRule(newRules);
    };

    const updateErrorConditions = (del: boolean, ruleIdx: number, idx: number, args?: ErrorCondition) => {
        const newRules = { ...breakerRule };
        if (del) {
            newRules.block_configs[ruleIdx].error_conditions.splice(idx, 1);
        } else {
            if (idx < newRules.block_configs[ruleIdx].error_conditions.length) {
                newRules.block_configs[ruleIdx].error_conditions[idx] = {
                    inputType: args?.inputType || ErrorConditionType.RET_CODE,
                    condition: {
                        type: args?.condition?.type || MatchType.EXACT,
                        value: args?.condition?.value || '',
                    }
                }
            } else {
                newRules.block_configs[ruleIdx].error_conditions.push(args || {
                    inputType: ErrorConditionType.RET_CODE,
                    condition: {
                        type: MatchType.EXACT,
                        value: '',
                    }
                });
            }
        }
        setBreakerRule(newRules);
    };

    const addTriggerConditions = (ruleIdx: number) => {
        const newRules = { ...breakerRule };
        newRules.block_configs[ruleIdx].trigger_conditions.push({
            triggerType: TriggerType.ERROR_RATE,
            errorCount: 0,
            errorPercent: 0,
            interval: 0,
            minimumRequest: 0
        });
        setBreakerRule(newRules);
    };

    const removeTriggerConditions = (ruleIdx: number, idx: number) => {
        const newRules = { ...breakerRule };
        if (idx < newRules.block_configs[ruleIdx].trigger_conditions.length) {
            newRules.block_configs[ruleIdx].trigger_conditions.splice(idx, 1);
        }
        setBreakerRule(newRules);
    };

    const updateTriggerConditions = (ruleIdx: number, idx: number, item: TriggerCondition) => {
        console.log('更新触发条件:', ruleIdx, idx, item);
        const newRules = { ...breakerRule };
        if (idx < newRules.block_configs[ruleIdx].trigger_conditions.length) {
            newRules.block_configs[ruleIdx].trigger_conditions[idx] = {
                triggerType: item.triggerType || TriggerType.ERROR_RATE,
                errorCount: item.errorCount,
                errorPercent: item.errorPercent,
                interval: item.interval,
                minimumRequest: item.minimumRequest,
                triggerVal: item.triggerVal,
            };
        } else {
            newRules.block_configs[ruleIdx].trigger_conditions.push(item);
        }
        setBreakerRule(newRules);
    }

    const metadataRecord = React.useMemo(
        () => (breakerRule.metadata || []).reduce<Record<string, string>>((acc, cur) => {
            if (cur.key) acc[cur.key] = cur.value;
            return acc;
        }, {}),
        [breakerRule.metadata],
    );

    const destinationView = React.useMemo(() => {
        const destination = breakerRule.ruleMatcher.destination;
        if (destination.service.includes('governance') && !destination.namespace.includes('governance')) {
            return {
                namespace: destination.service,
                service: destination.namespace,
            };
        }
        return destination;
    }, [breakerRule.ruleMatcher.destination]);

    const breakerNamespaceOptions = namespaceDatas.map((item: NamespaceView) => ({ label: item.name, value: item.name }));
    const breakerSourceServiceOptions = serviceDatas
        .filter((opt: ServiceView) => breakerRule.ruleMatcher.source.namespace === '*' || opt.namespace === breakerRule.ruleMatcher.source.namespace)
        .map((item: ServiceView) => ({ label: item.name, value: item.name, namespace: item.namespace }));
    const breakerDestServiceOptions = serviceDatas
        .filter((opt: ServiceView) => breakerRule.ruleMatcher.destination.namespace === '*' || opt.namespace === breakerRule.ruleMatcher.destination.namespace)
        .map((item: ServiceView) => ({ label: item.name, value: item.name, namespace: item.namespace }));

    // 统一基础信息区：名称 / 粒度 / 优先级 / 作用对象（主调→被调流向）/ 描述 / 标签
    const ruleeditorState = (
        <div className={shared.section}>
            <div className={shared.sectionHeader}>基础信息</div>
            <div className={shared.sectionBody}>
                <div className={shared.infoGrid}>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>规则名称</div>
                        {editorState.editable
                            ? <Input value={breakerRule.name} disabled={op === 'edit'} maxlength={64} onChange={(value) => setBreakerRule(prev => ({ ...prev, name: value }))} />
                            : <div className={shared.fieldValue}>{breakerRule.name || '-'}</div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>熔断粒度</div>
                        {editorState.editable ? (
                            <RadioGroup theme="button" variant="primary-filled" value={breakerRule.level} onChange={(value) => setBreakerRule(prev => ({ ...prev, level: value as string }))}>
                                <Radio.Button value={BreakLevelType.Service}>{BreakLevelMap[BreakLevelType.Service]}</Radio.Button>
                                <Radio.Button value={BreakLevelType.Instance}>{BreakLevelMap[BreakLevelType.Instance]}</Radio.Button>
                                <Radio.Button value={BreakLevelType.Method}>{BreakLevelMap[BreakLevelType.Method]}</Radio.Button>
                            </RadioGroup>
                        ) : <div className={shared.fieldValue}>{BreakLevelMap[breakerRule.level as BreakLevelType] || '-'}</div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>优先级</div>
                        {editorState.editable
                            ? <InputNumber min={0} value={breakerRule.priority} onChange={(value) => setBreakerRule(prev => ({ ...prev, priority: value as number }))} />
                            : <div className={shared.fieldValue}>{breakerRule.priority ?? 0}</div>}
                    </div>
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>作用对象</div>
                        {editorState.editable ? (
                            <div className={shared.kv2}>
                                <div>
                                    <div className={shared.editLabel}>主调命名空间</div>
                                    <Select filterable creatable options={breakerNamespaceOptions} value={breakerRule.ruleMatcher.source.namespace} onChange={(value) => setBreakerRule(prev => ({ ...prev, ruleMatcher: { ...prev.ruleMatcher, source: { ...prev.ruleMatcher.source, namespace: value as string } } }))} />
                                </div>
                                <div>
                                    <div className={shared.editLabel}>主调服务</div>
                                    <Select filterable creatable options={breakerSourceServiceOptions} value={breakerRule.ruleMatcher.source.service} onChange={(value) => setBreakerRule(prev => ({ ...prev, ruleMatcher: { ...prev.ruleMatcher, source: { ...prev.ruleMatcher.source, service: value as string } } }))} />
                                </div>
                                <div>
                                    <div className={shared.editLabel}>被调命名空间</div>
                                    <Select filterable creatable options={breakerNamespaceOptions} value={breakerRule.ruleMatcher.destination.namespace} onChange={(value) => setBreakerRule(prev => ({ ...prev, ruleMatcher: { ...prev.ruleMatcher, destination: { ...prev.ruleMatcher.destination, namespace: value as string } } }))} />
                                </div>
                                <div>
                                    <div className={shared.editLabel}>被调服务</div>
                                    <Select filterable creatable options={breakerDestServiceOptions} value={breakerRule.ruleMatcher.destination.service} onChange={(value) => setBreakerRule(prev => ({ ...prev, ruleMatcher: { ...prev.ruleMatcher, destination: { ...prev.ruleMatcher.destination, service: value as string } } }))} />
                                </div>
                            </div>
                        ) : (
                            <div className={shared.flow}>
                                <div className={shared.flowNode}>
                                    <span className={shared.flowNodeLabel}>主调</span>
                                    <span className={shared.flowNodeValue}>{`${breakerRule.ruleMatcher.source.namespace || '-'} / ${breakerRule.ruleMatcher.source.service || '-'}`}</span>
                                </div>
                                <span className={shared.flowArrow}><RocketIcon /></span>
                                <div className={shared.flowNode}>
                                    <span className={shared.flowNodeLabel}>被调</span>
                                    <span className={shared.flowNodeValue}>{`${destinationView.namespace || '-'} / ${destinationView.service || '-'}`}</span>
                                </div>
                            </div>
                        )}
                    </div>
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>描述</div>
                        {editorState.editable
                            ? <Input value={breakerRule.description} maxlength={255} onChange={(value) => setBreakerRule(prev => ({ ...prev, description: value }))} />
                            : <div className={shared.fieldValue}>{breakerRule.description || '-'}</div>}
                    </div>
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>规则标签</div>
                        <RuleLabelField
                            metadata={metadataRecord}
                            editable={editorState.editable}
                            onChange={(next) => setBreakerRule(prev => ({ ...prev, metadata: Object.entries(next).map(([key, value]) => ({ key, value })) }))}
                        />
                    </div>
                </div>
            </div>
        </div>
    );

    const errCondTableColumns = (ruleIdx: number): PrimaryTableProps['columns'] => [
        {
            colKey: 'inputType',
            title: '参数类型',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: {
                    clearable: true,
                    options: ErrorConditionOptions,
                    defaultValue: ErrorConditionType.RET_CODE,

                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateErrorConditions(false, ruleIdx, context.rowIndex, context.newRowData as ErrorCondition);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => {
                return <Text>{ErrorConditionMap[row.inputType as ErrorConditionType] || row.inputType}</Text>
            }
        },
        {
            colKey: 'condition.type',
            title: '匹配类型',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: {
                    clearable: true,
                    options: MatchTypeOption,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateErrorConditions(false, ruleIdx, context.rowIndex, context.newRowData as ErrorCondition);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => <Text>{MatchTypeMap[row.condition.type as MatchType]}</Text>
        },
        {
            colKey: 'condition.value',
            title: '匹配值',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Input,
                props: {
                    clearable: true,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateErrorConditions(false, ruleIdx, context.rowIndex, context.newRowData as ErrorCondition);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
        },
        {
            colKey: 'action',
            title: '操作',
            cell: ({ row, rowIndex }) => (
                <Space>
                    <Popup trigger="hover" content="添加">
                        <Button
                            shape="circle"
                            variant="text"
                            onClick={() => {
                                addErrorConditions(ruleIdx);
                            }}>
                            <AddIcon />
                        </Button>
                    </Popup>
                    <Popup trigger="hover" content="删除">
                        <Button
                            shape="circle"
                            variant="text"
                            disabled={breakerRule.block_configs[ruleIdx].error_conditions.length <= 1}
                            onClick={() => {
                                updateErrorConditions(true, ruleIdx, rowIndex, undefined);
                            }}>
                            <RemoveIcon />
                        </Button>
                    </Popup>
                </Space>
            ),
        }
    ]

    // 匹配条件表格（单行参数填写）
    const renderMatchTable = (idx: number, rule: BlockConfig) => {
        return (
            <div className={styles.compactTable}>
                <Table
                    key={`breaker-error-${idx}-${editorState.editable ? 'edit' : 'view'}`}
                    rowKey="key"
                    tableLayout="fixed"
                    data={rule.error_conditions.map((item, index) => ({ ...item, key: `error-${idx}-${index}` })) || []}
                    columns={editorState.editable ?
                        errCondTableColumns(idx)
                        :
                        errCondTableColumns(idx)?.filter(col => col.colKey !== 'action')}
                />
            </div>
        )
    };

    const triggerTableColumns = (idx: number): PrimaryTableProps['columns'] => [
        {
            colKey: 'triggerType',
            title: '类型',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: {
                    clearable: true,
                    options: TriggerTypeOptions
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateTriggerConditions(idx, context.rowIndex, context.newRowData as TriggerCondition);
                },
            },
            cell: ({ row }) => {
                return <Text>{TriggerTypeMap[row.triggerType as TriggerType]?.text || row.triggerType}</Text>
            }
        },
        {
            colKey: 'op_label',
            title: '类型',
            cell: ({ row }) => {
                return <Text>{">="}</Text>
            }
        },
        {
            colKey: 'triggerVal',
            title: '阈值',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: InputAdornment,
                props: (context: { row: TriggerCondition }) => ({
                    min: 0,
                    // 根据 triggerType 动态设置单位
                    append: context.row.triggerType === TriggerType.ERROR_RATE ? "%" : "个",
                    children: (
                        <InputNumber
                            min={0}
                            // 如果是错误率类型，可以设置最大值为100
                            max={context.row.triggerType === TriggerType.ERROR_RATE ? 100 : undefined}
                            // 如果是错误率类型，可以设置步长为0.1
                            step={context.row.triggerType === TriggerType.ERROR_RATE ? 0.1 : 1}
                        />
                    ),
                }),
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateTriggerConditions(idx, context.rowIndex, {
                        ...context.newRowData as TriggerCondition,
                        errorCount: context.newRowData.triggerType === TriggerType.ERROR_RATE ? 0 : context.newRowData.triggerVal || 0,
                        errorPercent: context.newRowData.triggerType === TriggerType.ERROR_RATE ? context.newRowData.triggerVal || 0 : 0,
                    });
                },
            },
            cell: ({ row }) => (
                editorState.editable ? (
                    <Text>{row.triggerType === TriggerType.ERROR_RATE ? `${row.errorPercent}` : `${row.errorCount}`}</Text>
                ) : (
                    <Text>{row.triggerType === TriggerType.ERROR_RATE ? `${row.errorPercent} %` : `${row.errorCount} 个`}</Text>
                )
            )
        },
        {
            colKey: 'interval',
            title: '统计周期',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: InputAdornment,
                props: {
                    min: 0,
                    append: ("秒"),
                    children: (
                        <InputNumber min={0} />
                    ),
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateTriggerConditions(idx, context.rowIndex, context.newRowData as TriggerCondition);
                },
            },
            cell: ({ row }) => (
                editorState.editable ? (
                    <Text>{row.tinterval}</Text>
                ) : (
                    <Text>{`${row.interval} 秒`}</Text>
                )
            )
        },
        {
            colKey: 'minimumRequest',
            title: '最小请求数',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: InputAdornment,
                props: {
                    min: 0,
                    append: ("个"),
                    children: (
                        <InputNumber min={0} />
                    ),
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateTriggerConditions(idx, context.rowIndex, context.newRowData as TriggerCondition);
                },
            }
        },
        {
            colKey: 'action',
            title: '操作',
            cell: ({ row, rowIndex }) => (
                <Space>
                    <Popup trigger="hover" content="编辑标签">
                        <Button
                            shape="circle"
                            variant="text"
                            onClick={() => addTriggerConditions(idx)}>
                            <AddIcon />
                        </Button>
                    </Popup>
                    <Popup trigger="hover" content="删除分组">
                        <Button
                            shape="circle"
                            variant="text"
                            disabled={breakerRule.block_configs[idx].trigger_conditions.length <= 1}
                            onClick={() => removeTriggerConditions(idx, rowIndex)}>
                            <RemoveIcon />
                        </Button>
                    </Popup>
                </Space>
            ),
        }
    ]

    // 路由策略分组表格（弹窗编辑标签）
    const renderGroupTable = (idx: number, rule: BlockConfig) => (
        <div className={styles.compactTable}>
            <Table
                key={`breaker-trigger-${idx}-${editorState.editable ? 'edit' : 'view'}`}
                rowKey="key"
                tableLayout={'fixed'}
                data={rule.trigger_conditions}
                columns={editorState.editable ?
                    triggerTableColumns(idx)
                    :
                    triggerTableColumns(idx)?.filter(col => col.colKey !== 'action')}
            />
        </div>
    );


    const apiTableColumns = (ruleIdx: number): PrimaryTableProps['columns'] => [
        {
            colKey: 'protocol',
            title: '协议',
            width: 240,
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: {
                    filterable: true,
                    creatable: true,
                    options: InterfaceProtocolOption,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    const newRules = { ...breakerRule };
                    if (!newRules.block_configs[ruleIdx].api) {
                        newRules.block_configs[ruleIdx].api = { protocol: '', method: '', path: { type: MatchType.EXACT, value: '', value_type: MatchValueType.TEXT } };
                    }
                    newRules.block_configs[ruleIdx].api = {
                        ...newRules.block_configs[ruleIdx].api,
                        protocol: context.newRowData.protocol as string,
                    }
                    setBreakerRule(newRules);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => {
                return <Text>{row.protocol}</Text>
            }
        },
        {
            colKey: 'path.value',
            title: '接口路径',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: InputAdornment,
                props: (context: { row: API }) => ({
                    clearable: true,
                    children: (
                        <Input readonly={!editorState.editable} />
                    ),
                    append: (
                        <Select
                            autoWidth={true}
                            defaultValue={MatchType.EXACT}
                            options={MatchTypeOption}
                            value={breakerRule.block_configs[ruleIdx]?.api?.path.type}
                            readonly={!editorState.editable}
                            onChange={(value) => {
                                const newRules = { ...breakerRule };
                                if (!newRules.block_configs[ruleIdx].api) {
                                    newRules.block_configs[ruleIdx].api = { protocol: '', method: '', path: { type: MatchType.EXACT, value: '', value_type: MatchValueType.TEXT } };
                                }
                                newRules.block_configs[ruleIdx].api = {
                                    ...newRules.block_configs[ruleIdx].api,
                                    path: {
                                        type: value as MatchType,
                                        value: context.row.path.value || value as string, // 保持原有值
                                        value_type: MatchValueType.TEXT, // 默认值类型为文本
                                    }
                                }
                                setBreakerRule(newRules);
                            }}
                        />
                    ),
                }),
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    const newRules = { ...breakerRule };
                    if (!newRules.block_configs[ruleIdx].api) {
                        newRules.block_configs[ruleIdx].api = { protocol: '', method: '', path: { type: MatchType.EXACT, value: '', value_type: MatchValueType.TEXT } };
                    }
                    newRules.block_configs[ruleIdx].api.path = {
                        type: context.newRowData.path.type as MatchType,
                        value: context.newRowData.path.value || '',
                        value_type: MatchValueType.TEXT,
                    };
                    setBreakerRule(newRules);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => {
                return <Text>{row.path?.value}</Text>
            }
        },
        {
            colKey: 'method',
            title: '接口方法',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: (context: { row: API }) => ({
                    creatable: true,
                    filterable: true,
                    options: context.row.protocol === 'HTTP' ? HTTPMethodOption : [],
                }),
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    const newRules = { ...breakerRule };
                    if (!newRules.block_configs[ruleIdx].api) {
                        newRules.block_configs[ruleIdx].api = { protocol: '', method: '', path: { type: MatchType.EXACT, value: '', value_type: MatchValueType.TEXT } };
                    }
                    newRules.block_configs[ruleIdx].api = {
                        ...newRules.block_configs[ruleIdx].api,
                        method: context.newRowData.method as string,
                    }
                    setBreakerRule(newRules);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => {
                return <Text>{row.method}</Text>
            }
        },
    ]

    // 路由策略分组表格（弹窗编辑标签）
    const renderApiTable = (idx: number, rule: BlockConfig) => (
        <div className={styles.compactTable}>
            <Table
                key={`breaker-api-${idx}-${editorState.editable ? 'edit' : 'view'}`}
                rowKey="key"
                tableLayout={'fixed'}
                data={[rule.api ? rule.api : { key: `api-${idx}`, method: '', path: { type: MatchType.EXACT, value: '', value_type: MatchValueType.TEXT }, }]}
                columns={apiTableColumns(idx)}
            />
        </div>
    );

    const renderRecover = (idx: number, rule: BlockConfig) => (
        <div className={shared.step} data-step="4">
            <div className={shared.stepTitle}>恢复策略<span className={shared.stepHint}>控制该策略进入熔断后的半开恢复和主动探测</span></div>
            <div className={shared.stepContent}>
                <Row gutter={[16, 16]}>
                    <Col span={4}>
                        <FormItem label={"最大剔除比例"}>
                            {editorState.editable ? (
                                <InputAdornment append={"%"}>
                                    <InputNumber
                                        min={0}
                                        max={100}
                                        value={rule.max_ejection_percent}
                                        onChange={(val) => {
                                            const newRules = cloneDeep(breakerRule);
                                            newRules.block_configs[idx].max_ejection_percent = val as number;
                                            setBreakerRule(newRules);
                                        }} />
                                </InputAdornment>
                            ) : renderReadonlyValue(`${rule.max_ejection_percent ?? '-'}%`)}
                        </FormItem>
                    </Col>
                    <Col span={4}>
                        <FormItem label={"熔断时长"}>
                            {editorState.editable ? (
                                <InputAdornment append={"秒"}>
                                    <InputNumber
                                        min={0}
                                        value={rule.recoverCondition?.sleepWindow}
                                        onChange={(val) => {
                                            const newRules = cloneDeep(breakerRule);
                                            newRules.block_configs[idx].recoverCondition.sleepWindow = val as number;
                                            setBreakerRule(newRules);
                                        }} />
                                </InputAdornment>
                            ) : renderReadonlyValue(`${rule.recoverCondition?.sleepWindow ?? '-'} 秒`)}
                        </FormItem>
                    </Col>
                    <Col span={4}>
                        <FormItem label={"主动探测"}>
                            {editorState.editable ? (
                                <Switch
                                    value={rule.faultDetectConfig?.enable}
                                    onChange={(checked) => {
                                        const newRules = cloneDeep(breakerRule);
                                        newRules.block_configs[idx].faultDetectConfig = {
                                            enable: checked as boolean,
                                        };
                                        setBreakerRule(newRules);
                                    }} />
                            ) : renderReadonlySwitch(rule.faultDetectConfig?.enable)}
                        </FormItem>
                    </Col>
                </Row>
            </div>
        </div>
    );

    const renderFallback = (idx: number, rule: BlockConfig) => (
        <div className={shared.step} data-step="5">
            <div className={shared.stepTitle}>熔断后降级<span className={shared.stepHint}>该策略触发时返回的兜底响应</span></div>
            <div className={shared.stepContent}>
                <FormItem label={"是否开启"}>
                    {editorState.editable ? (
                        <Switch
                            value={rule.fallbackConfig?.enable}
                            onChange={(checked) => {
                                const newRules = cloneDeep(breakerRule);
                                newRules.block_configs[idx].fallbackConfig = {
                                    ...newRules.block_configs[idx].fallbackConfig,
                                    enable: checked as boolean,
                                };
                                setBreakerRule(newRules);
                            }} />
                    ) : renderReadonlySwitch(rule.fallbackConfig?.enable)}
                </FormItem>
                {rule.fallbackConfig?.enable && (
                    <>
                        <FormItem label={"响应码"}>
                            {editorState.editable ? (
                                <InputNumber
                                    value={rule.fallbackConfig?.response?.code}
                                    onChange={(val) => {
                                        const newRules = cloneDeep(breakerRule);
                                        newRules.block_configs[idx].fallbackConfig.response.code = val as number;
                                        setBreakerRule(newRules);
                                    }} />
                            ) : renderReadonlyValue(rule.fallbackConfig?.response?.code ?? '-')}
                        </FormItem>
                        <FormItem label={"响应头"}>
                            <div>
                                {Object.entries(rule.fallbackConfig?.response?.headers || []).length > 0 ? (
                                    <Space>
                                        {Object.entries(rule.fallbackConfig?.response?.headers || []).map(([key, value], headerIdx) => (
                                            <Tag key={headerIdx}>
                                                {`${value.key}: ${value.value}`}
                                            </Tag>
                                        ))}
                                    </Space>
                                ) : (
                                    <Text>暂无响应头</Text>
                                )}
                                {editorState.editable && (
                                    <Button
                                        shape="circle"
                                        variant="text"
                                        onClick={() => setEditorState(prev => ({ ...prev, headerVisible: true, headerRuleIndex: idx }))}
                                    >
                                        <Edit1Icon />
                                    </Button>
                                )}
                                {renderRspHeaderDialog(idx, rule)}
                            </div>
                        </FormItem>
                        <FormItem label={"响应体"}>
                            {editorState.editable ? (
                                <Textarea
                                    value={rule.fallbackConfig?.response?.body || ''}
                                    onChange={(value) => {
                                        const newRules = cloneDeep(breakerRule);
                                        newRules.block_configs[idx].fallbackConfig.response.body = value as string;
                                        setBreakerRule(newRules);
                                    }} />
                            ) : (
                                <pre className={styles.readonlyCodeBlock}>
                                    {rule.fallbackConfig?.response?.body || '-'}
                                </pre>
                            )}
                        </FormItem>
                    </>
                )}
            </div>
        </div>
    );

    // 规则区块
    const renderRule = (rule: BlockConfig, idx: number) => {
        const collapsed = collapsedRuleIndexes.has(idx);
        const errorCount = rule.error_conditions?.length || 0;
        const triggerCount = rule.trigger_conditions?.length || 0;
        const hasApiStep = breakerRule.level === BreakLevelType.Method || !!rule.api;
        return (
            <div className={`${shared.policy} ${collapsed ? shared.policyCollapsed : ''}`} key={idx}>
                <div className={shared.policyHead} onClick={() => toggleRuleCollapsed(idx)}>
                    <div className={shared.policyHeadMain}>
                        <span className={shared.caret}><ChevronRightIcon /></span>
                        <div>
                            <div className={shared.policyIndex}>熔断策略 [{idx + 1}]{rule.name ? `：${rule.name}` : ''}</div>
                            <div className={shared.policySummary}>
                                {errorCount} 个错误判断条件 / {triggerCount} 个触发条件{rule.api ? ' / 指定接口' : ''}
                            </div>
                        </div>
                    </div>
                    <div className={shared.policyHeadActions} onClick={(e) => e.stopPropagation()}>
                        <Tag variant="light">{BreakLevelMap[breakerRule.level as BreakLevelType] || '服务'}粒度</Tag>
                        {editorState.editable && (
                            <Popup trigger="hover" content="删除策略">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newRules = { ...breakerRule };
                                        newRules.block_configs.splice(idx, 1);
                                        setBreakerRule(newRules);
                                    }}>
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        )}
                    </div>
                </div>
                <div className={shared.policyBody}>
                    {hasApiStep && (
                        <div className={shared.step} data-step="1">
                            <div className={shared.stepTitle}>接口范围<span className={shared.stepHint}>满足该接口条件的请求才进入熔断判断</span></div>
                            <div className={shared.stepContent}>{renderApiTable(idx, rule)}</div>
                        </div>
                    )}
                    <div className={shared.step} data-step={hasApiStep ? '2' : '1'}>
                        <div className={shared.stepTitle}>错误判断条件<span className={shared.stepHint}>满足任一应答条件的请求会被标识为错误</span></div>
                        <div className={shared.stepContent}>{renderMatchTable(idx, rule)}</div>
                    </div>
                    <div className={shared.step} data-step={hasApiStep ? '3' : '2'}>
                        <div className={shared.stepTitle}>熔断触发条件<span className={shared.stepHint}>满足任一统计条件即可触发熔断</span></div>
                        <div className={shared.stepContent}>{renderGroupTable(idx, rule)}</div>
                    </div>
                    {renderRecover(idx, rule)}
                    {renderFallback(idx, rule)}
                </div>
            </div>
        );
    }

    // 标签弹窗渲染
    const renderRspHeaderDialog = (ruleIdx: number, rule: BlockConfig) => (
        <Dialog
            visible={editorState.headerVisible && editorState.headerRuleIndex === ruleIdx}
            header="编辑响应头"
            width={700}
            onConfirm={() => setEditorState(prev => ({ ...prev, headerVisible: false, headerRuleIndex: undefined }))}
            onClose={() => setEditorState(prev => ({ ...prev, headerVisible: false, headerRuleIndex: undefined }))}
        >
            <div>
                {rule?.fallbackConfig?.response?.headers?.map((tag, idx) => (
                    <Row key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Space size={8} style={{ width: '100%' }}>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const newRules = cloneDeep(breakerRule);
                                    const headers = [...(newRules.block_configs[ruleIdx]?.fallbackConfig?.response?.headers || [])];
                                    headers[idx] = { ...headers[idx], key: v };
                                    newRules.block_configs[ruleIdx].fallbackConfig.response.headers = headers;
                                    setBreakerRule(newRules);
                                }} />
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const newRules = cloneDeep(breakerRule);
                                    const headers = [...(newRules.block_configs[ruleIdx]?.fallbackConfig?.response?.headers || [])];
                                    headers[idx] = { ...headers[idx], value: v };
                                    newRules.block_configs[ruleIdx].fallbackConfig.response.headers = headers;
                                    setBreakerRule(newRules);
                                }} />
                            <Popup trigger="hover" content="删除标签">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newRules = cloneDeep(breakerRule);
                                        const headers = [...(newRules.block_configs[ruleIdx]?.fallbackConfig?.response?.headers || [])];
                                        headers.splice(idx, 1);
                                        newRules.block_configs[ruleIdx].fallbackConfig.response.headers = headers;
                                        setBreakerRule(newRules);
                                    }}
                                >
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        </Space>
                    </Row>
                ))}
                <Button
                    variant="text"
                    icon={<AddIcon />}
                    onClick={() => {
                        const newRules = cloneDeep(breakerRule);
                        const headers = [...(newRules.block_configs[ruleIdx]?.fallbackConfig?.response?.headers || []), { key: '', value: '' }];
                        newRules.block_configs[ruleIdx].fallbackConfig.response.headers = headers;
                        setBreakerRule(newRules);
                    }}>
                    添加标签
                </Button>
            </div>
        </Dialog>
    );

    const renderPublishForm = (
        <>
            {editorState.publishView && (
                <PublishForm
                    ruleId={viewRule?.id || ''}
                    ruleName={viewRule?.name || ''}
                    resource={PolicySourceType.CircuitBreakerRules}
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
            {viewRule?.editable && (
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
                                        setEditorState(prev => ({ ...prev, editable: false }));
                                    }
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
        <div style={{ padding: 24 }}>
            <Form
                form={form}
                onSubmit={onSubmit}
                layout="vertical"
                colon
            >
                {ruleeditorState}
                <div className={shared.section}>
                    <div className={shared.sectionHeader}>
                        <span>熔断策略</span>
                        <span className={shared.countTag}>{breakerRule.block_configs.length} 条</span>
                    </div>
                    <div className={shared.sectionBody}>
                        <div className={shared.ruleList}>
                            {breakerRule.block_configs.map((rule, idx) => renderRule(rule, idx))}
                        </div>
                        {editorState.editable && (
                            <Button
                                className={shared.addRuleButton}
                                style={{ marginTop: 12 }}
                                variant="dashed" icon={<AddIcon />}
                                onClick={() => {
                                    const newRule = { ...breakerRule }
                                    newRule.block_configs.push(defaultBlockConfig(newRule.block_configs.length + 1))
                                    setBreakerRule(newRule);
                                }}>
                                添加熔断策略
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

export default React.memo(CircuitBreakerEditor);
