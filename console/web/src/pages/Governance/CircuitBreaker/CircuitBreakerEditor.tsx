import React from "react";
import { Col, Form, Input, Row, Space, Button, Select, Switch, Dialog, InputNumber, Table, FormProps, Tag, Popup, TableRowData, PrimaryTableProps, StickyTool, RadioGroup, Radio, InputAdornment, Textarea } from "tdesign-react";
import { SendIcon, AddIcon, ChevronDownIcon, ChevronRightIcon, CloseIcon, Edit1Icon, SaveIcon, RocketIcon, RollbackIcon, RemoveIcon } from "tdesign-icons-react";

import Text from "components/Text";
import { useAppDispatch, useAppSelector } from 'modules/store';
import { API, HTTPMethodOption, InterfaceProtocolOption, Label, MatchType, MatchTypeMap, MatchTypeOption, MatchValueType, Op } from "services/types";
import { ServiceView } from "services/service";

import { NamespaceView } from "services/namespace";
import styles from './CircuitBreakerEditor.module.less';
import { openErrNotification, openInfoNotification } from "utils/notifition";
import PublishForm from "../RuleRelease/PublishForm";
import RuleStickyAction from "../RuleRelease/RuleStickyAction";
import { PolicySourceType } from "services/auth_policy";
import { BlockConfig, BreakLevelMap, BreakLevelType, CircuitBreakerRule, ErrorCondition, ErrorConditionMap, ErrorConditionOptions, ErrorConditionType, FallbackConfig, FaultDetectConfig, RecoverCondition, TriggerCondition, TriggerType, TriggerTypeMap, TriggerTypeOptions } from "services/circuitbreaker";
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
    recoverCondition: RecoverCondition
    faultDetectConfig: FaultDetectConfig
    fallbackConfig: FallbackConfig
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
    recoverCondition: {
        sleepWindow: 60,
        consecutiveSuccess: 0
    },
    faultDetectConfig: {
        enable: false
    },
    fallbackConfig: {
        enable: false,
        response: {
            code: 500,
            headers: [],
            body: ''
        },
    }
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
        if (op === 'edit') {
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
            openInfoNotification('请求成功', op === 'edit' ? '修改熔断规则成功' : '创建熔断规则成功');
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

    // 标签弹窗渲染
    const renderRuleLabelsDialog = () => (
        <Dialog
            visible={editorState.visible}
            header="编辑规则标签"
            width={700}
            onConfirm={() => setEditorState(prev => ({ ...prev, visible: false }))}
            onClose={() => setEditorState(prev => ({ ...prev, visible: false }))}
        >
            <div>
                {breakerRule.metadata?.map((tag, idx) => (
                    <Row key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Space>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const newMetadata = [...(breakerRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], key: v };
                                    setBreakerRule(prev => ({ ...prev, metadata: newMetadata }));
                                }} />
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const newMetadata = [...(breakerRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], value: v };
                                    setBreakerRule(prev => ({ ...prev, metadata: newMetadata }));
                                }} />
                            <Popup trigger="hover" content="删除标签">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newMetadata = [...(breakerRule.metadata || [])];
                                        newMetadata.splice(idx, 1);
                                        setBreakerRule(prev => ({ ...prev, metadata: newMetadata }));
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
                        const newMetadata = [...(breakerRule.metadata || []), { key: '', value: '' }];
                        setBreakerRule(prev => ({ ...prev, metadata: newMetadata }));
                    }}>
                    添加标签
                </Button>
            </div>
        </Dialog>
    );

    const breakerType = (
        <Space>
            <FormItem label='熔断粒度' initialData={BreakLevelType.Service}>
                <RadioGroup
                    theme='button'
                    variant='primary-filled'
                    readonly={!editorState.editable}
                    onChange={(value) => {
                        setBreakerRule(prev => ({ ...prev, level: value as string }));
                    }}
                >
                    <Radio.Button value={BreakLevelType.Service}>
                        {BreakLevelMap[BreakLevelType.Service]}
                    </Radio.Button>
                    <Radio.Button
                        value={BreakLevelType.Instance}>
                        {BreakLevelMap[BreakLevelType.Instance]}
                    </Radio.Button>
                    <Radio.Button
                        value={BreakLevelType.Method}>
                        {BreakLevelMap[BreakLevelType.Method]}
                    </Radio.Button>
                </RadioGroup>
            </FormItem>
        </Space>
    )

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

    // 基础信息表单：第一层名称，第二层优先级与描述，第三层规则标签
    const ruleeditorState = (
        <div className={styles.sectionCard}>
            {/* 第一层：规则名称 */}
            <Row style={{ marginBottom: 16 }}>
                <Col span={12}>
                    <FormItem
                        label="规则名称"
                        name="name"
                        showErrorMessage={editorState.editable}
                        rules={[
                            { required: true, message: '请输入规则名称' },
                            { max: 64, message: '规则名称长度不能超过64个字符' }
                        ]}
                        requiredMark={op === 'create'}
                    >
                        {editorState.editable ? (
                            <Input
                                disabled={op === 'edit'}
                                readonly={op === 'view'}
                                onChange={(value) => setBreakerRule(prev => ({ ...prev, name: value }))}
                            />
                        ) : (
                            <Text>{breakerRule.name}</Text>
                        )}
                    </FormItem>
                </Col>
            </Row>
            {/* 第二层：优先级、描述 */}
            <Row style={{ marginBottom: 16 }}>
                <Space>
                    <FormItem label="优先级" name="priority" initialData={0}>
                        {editorState.editable ? (
                            <InputNumber min={0} readonly={op === 'view'} onChange={(value) => setBreakerRule(prev => ({ ...prev, priority: value as number }))} />
                        ) : (
                            <Text>{breakerRule.priority}</Text>
                        )}
                    </FormItem>
                    <FormItem
                        label="描述"
                        name="description"
                        rules={[
                            { max: 255, message: '描述长度不能超过255个字符' }
                        ]}
                    >
                        {editorState.editable ? (
                            <Input readonly={op === 'view'} onChange={(value) => setBreakerRule(prev => ({ ...prev, description: value }))} />
                        ) : (
                            <Text>{breakerRule.description}</Text>
                        )}
                    </FormItem>
                </Space>
            </Row>
            {/* 第三层：规则标签 */}
            <Row>
                <Col span={12}>
                    <FormItem label='规则标签' name='labels'>
                        <Space align="center">
                            {Array.isArray(breakerRule.metadata) && breakerRule.metadata.length > 0 ? (
                                <>
                                    {breakerRule.metadata.map((item: Label, idx: number) => (
                                        <Tag key={idx}>{`${item.key}: ${item.value}`}</Tag>
                                    ))}
                                    {editorState.editable && (
                                        <Button
                                            shape="circle"
                                            variant="text"
                                            onClick={() => setBreakerRule(prev => ({ ...prev, visible: true }))}
                                        >
                                            <Edit1Icon />
                                        </Button>
                                    )}
                                </>
                            ) : (
                                <>
                                    <Text>暂无标签</Text>
                                    {editorState.editable && (
                                        <Button
                                            shape="circle"
                                            variant="text"
                                            onClick={() => setBreakerRule(prev => ({ ...prev, visible: true }))}
                                        >
                                            <Edit1Icon />
                                        </Button>
                                    )}
                                </>
                            )}
                        </Space>
                    </FormItem>
                    {renderRuleLabelsDialog()}
                </Col>
            </Row>
            <Row>
                <Col span={12}>
                    {breakerType}
                </Col>
            </Row>
        </div>
    );


    // 主调/被调服务卡片
    const callinfo = (
        <>
            <div className={styles['route-editor-header']}>
                <div className={styles['route-editor-header-block']}>
                    <div className={styles['route-editor-header-title']}>
                        主调服务
                    </div>
                    <div className={styles['route-editor-header-desc']}>
                        请求将按照匹配规则进行目标服务路由
                    </div>
                    <FormItem
                        style={{ marginBottom: 0 }}
                        label="命名空间"
                        name="caller_namespace"
                        initialData={"*"}
                        showErrorMessage={editorState.editable}
                        rules={[{ required: true, message: '请选择主调服务命名空间' }]}
                        requiredMark={editorState.editable}
                    >
                        {editorState.editable ? (
                            <Select
                                filterable={true}
                                creatable={true}
                                options={namespaceDatas.map((item: NamespaceView) => ({
                                    label: item.name,
                                    value: item.name,
                                }))}
                                onChange={(value) => {
                                    setBreakerRule(prev => ({
                                        ...prev,
                                        ruleMatcher: {
                                            ...prev.ruleMatcher,
                                            source: {
                                                ...prev.ruleMatcher.source,
                                                namespace: value as string,
                                            },
                                        },
                                    }));
                                }}
                            />
                        ) : (
                            <Text>{breakerRule.ruleMatcher.source.namespace}</Text>
                        )}
                    </FormItem>
                    <FormItem
                        style={{ marginBottom: 0 }}
                        label="服务名称"
                    >
                        {editorState.editable ? (
                            <Select
                                filterable={true}
                                creatable={true}
                                options={serviceDatas.filter(opt => {
                                    if (breakerRule.ruleMatcher.source.namespace === '*') {
                                        return true; // 允许所有命名空间的服务
                                    }
                                    return opt.namespace === breakerRule.ruleMatcher.source.namespace; // 仅允许当前命名空间
                                }).map((item: ServiceView) => ({
                                    label: `${item.name}`,
                                    value: item.name,
                                    namespace: item.namespace,
                                }))}
                                onChange={(value) => {
                                    setBreakerRule(prev => ({
                                        ...prev,
                                        ruleMatcher: {
                                            ...prev.ruleMatcher,
                                            source: {
                                                ...prev.ruleMatcher.source,
                                                service: value as string,
                                            },
                                        },
                                    }));
                                }}
                            />
                        ) : (
                            <Text>{breakerRule.ruleMatcher.source.service}</Text>
                        )}
                    </FormItem>
                </div>
                <div className={styles['route-editor-header-arrow']}>
                    <SendIcon style={{ fontSize: 36, color: '#bfbfbf' }} />
                </div>
                <div className={styles['route-editor-header-block']}>
                    <div className={styles['route-editor-header-title']}>
                        被调服务
                    </div>
                    <div className={styles['route-editor-header-desc']}>
                        请求会按照规则路由到目标服务分组
                    </div>
                    <FormItem
                        style={{ marginBottom: 0 }}
                        label="命名空间"
                        name="callee_namespace"
                        initialData={"*"}
                        showErrorMessage={editorState.editable}
                        rules={[{ required: true, message: '请选择被调服务命名空间' }]}
                        requiredMark={editorState.editable}
                    >
                        {editorState.editable ? (
                            <Select
                                filterable={true}
                                creatable={true}
                                options={namespaceDatas.map((item: NamespaceView) => ({
                                    label: item.name,
                                    value: item.name,
                                }))}
                                onChange={(value) => {
                                    setBreakerRule(prev => ({
                                        ...prev,
                                        ruleMatcher: {
                                            ...prev.ruleMatcher,
                                            destination: {
                                                ...prev.ruleMatcher.destination,
                                                namespace: value as string,
                                            },
                                        },
                                    }));
                                }}
                            />
                        ) : (
                            <Text>{destinationView.namespace}</Text>
                        )}
                    </FormItem>
                    <FormItem
                        style={{ marginBottom: 0 }}
                        label="服务名称"
                        name="callee_service"
                        showErrorMessage={editorState.editable}
                        rules={[{ required: true, message: '请选择被调服务' }]}
                        requiredMark={editorState.editable}
                    >
                        {editorState.editable ? (
                            <Select
                                filterable={true}
                                creatable={true}
                                options={serviceDatas.filter(opt => {
                                    if (breakerRule.ruleMatcher.destination.namespace === '*') {
                                        return true; // 允许所有命名空间的服务
                                    }
                                    return opt.namespace === breakerRule.ruleMatcher.destination.namespace; // 仅允许当前命名空间
                                }).map((item: ServiceView) => ({
                                    label: `${item.name}`,
                                    value: item.name,
                                    namespace: item.namespace,
                                }))}
                                onChange={(value) => {
                                    setBreakerRule(prev => ({
                                        ...prev,
                                        ruleMatcher: {
                                            ...prev.ruleMatcher,
                                            destination: {
                                                ...prev.ruleMatcher.destination,
                                                service: value as string,
                                            },
                                        },
                                    }));
                                }}
                            />
                        ) : (
                            <Text>{destinationView.service}</Text>
                        )}
                    </FormItem>
                </div>
            </div>
        </>
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
                rowKey="key"
                tableLayout={'fixed'}
                data={[rule.api ? rule.api : { key: `api-${idx}`, method: '', path: { type: MatchType.EXACT, value: '', value_type: MatchValueType.TEXT }, }]}
                columns={apiTableColumns(idx)}
            />
        </div>
    );

    // 规则区块
    const renderRule = (rule: BlockConfig, idx: number) => {
        const collapsed = collapsedRuleIndexes.has(idx);
        const errorCount = rule.error_conditions?.length || 0;
        const triggerCount = rule.trigger_conditions?.length || 0;
        return (
            <div className={`${styles.sectionCard} ${styles.ruleBlock}`} key={idx}>
                <div className={styles.ruleBlockHeader}>
                    <div className={styles.ruleHeaderMain}>
                        <button
                            type="button"
                            className={styles.collapseButton}
                            onClick={() => toggleRuleCollapsed(idx)}
                            aria-label={collapsed ? '展开熔断策略' : '折叠熔断策略'}
                        >
                            {collapsed ? <ChevronRightIcon /> : <ChevronDownIcon />}
                        </button>
                        <div className={styles.ruleTitleGroup}>
                            <span>熔断策略 [{idx + 1}]{rule.name ? `：${rule.name}` : ''}</span>
                            <div className={styles.ruleSummary}>
                                {errorCount} 个错误判断条件 / {triggerCount} 个触发条件{rule.api ? ' / 指定接口' : ''}
                            </div>
                        </div>
                    </div>
                    <div className={styles.ruleHeaderActions}>
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
                <div className={`${styles.ruleBlockBody} ${collapsed ? styles.ruleBlockBodyCollapsed : ''}`}>
                    {(breakerRule.level === BreakLevelType.Method || rule.api) && (
                        <div className={styles.ruleSection}>
                            <div className={styles.ruleSectionTitle}>接口</div>
                            <div className={styles.ruleHelp}>
                                满足以下接口条件的请求会进入该熔断策略判断
                            </div>
                            {renderApiTable(idx, rule)}
                        </div>
                    )}
                    <div className={styles.ruleSection}>
                        <div className={styles.ruleSectionTitle}>错误判断条件</div>
                        <div className={styles.ruleHelp}>
                            满足以下任一应答条件的请求会被标识为错误请求
                        </div>
                        {renderMatchTable(idx, rule)}
                    </div>
                    <div className={styles.ruleSection}>
                        <div className={styles.ruleSectionTitle}>熔断触发条件</div>
                        <div className={styles.ruleHelp}>
                            满足以下任一统计条件即可触发熔断
                        </div>
                        {renderGroupTable(idx, rule)}
                    </div>
                </div>
            </div>
        );
    }

    const renderRecover = (
        <>
            <div className={styles.sectionCard} style={{ marginBottom: 24 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                    <span style={{ fontWeight: 600 }}>恢复策略</span>
                </div>
                <div className={styles['route-editor-header-desc']}>
                    进入熔断状态后，通过设置超时探测或者主动探测规则，满足条件后即可结束熔断状态恢复业务请求，否则重新回到熔断状态
                </div>
                <Space>
                </Space>
                <FormItem label={"熔断时长"}>
                    {editorState.editable ? (
                        <div>
                            <InputAdornment append={"秒"}>
                                <InputNumber
                                    min={0}
                                    value={breakerRule?.recoverCondition?.sleepWindow}
                                    onChange={(val) => {
                                        const newRules = { ...breakerRule };
                                        newRules.recoverCondition.sleepWindow = val as number;
                                        setBreakerRule(newRules);
                                    }} />
                            </InputAdornment>
                        </div>
                    ) : renderReadonlyValue(`${breakerRule?.recoverCondition?.sleepWindow ?? '-'} 秒`)}
                </FormItem>
                <FormItem label={"主动探测"} help={"开启主动探测时，客户端将会根据您配置的探测规则对目标被调服务进行探测； 主动探测请求与业务调用合并判断熔断恢复（如未匹配到探测规则，则不会生效）； 未开启主动探测时，会仅根据业务调用判断熔断恢复。"}>
                    {editorState.editable ? (
                        <div>
                            <Switch
                                value={breakerRule?.faultDetectConfig?.enable}
                                onChange={(checked) => {
                                    const newRules = { ...breakerRule };
                                    newRules.faultDetectConfig = {
                                        enable: checked as boolean,
                                    }
                                    setBreakerRule(newRules);
                                }} />
                        </div>
                    ) : renderReadonlySwitch(breakerRule?.faultDetectConfig?.enable)}
                </FormItem>
            </div>
        </>
    );

    // 标签弹窗渲染
    const renderRspHeaderDialog = () => (
        <Dialog
            visible={editorState.headerVisible}
            header="编辑响应头"
            width={700}
            onConfirm={() => setEditorState(prev => ({ ...prev, headerVisible: false }))}
            onClose={() => setEditorState(prev => ({ ...prev, headerVisible: false }))}
        >
            <div>
                {breakerRule?.fallbackConfig?.response?.headers?.map((tag, idx) => (
                    <Row key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Space size={8} style={{ width: '100%' }}>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const newMetadata = [...(breakerRule?.fallbackConfig?.response?.headers || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], key: v };
                                    setBreakerRule(prev => ({
                                        ...prev,
                                        fallbackConfig: {
                                            ...prev.fallbackConfig,
                                            response: {
                                                ...prev.fallbackConfig.response,
                                                headers: newMetadata,
                                            },
                                        },
                                    }));
                                }} />
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const newMetadata = [...(breakerRule?.fallbackConfig?.response?.headers || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], value: v };
                                    setBreakerRule(prev => ({
                                        ...prev,
                                        fallbackConfig: {
                                            ...prev.fallbackConfig,
                                            response: {
                                                ...prev.fallbackConfig.response,
                                                headers: newMetadata,
                                            },
                                        },
                                    }));
                                }} />
                            <Popup trigger="hover" content="删除标签">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newMetadata = [...(breakerRule?.fallbackConfig?.response?.headers || [])];
                                        newMetadata.splice(idx, 1);
                                        setBreakerRule(prev => ({
                                            ...prev,
                                            fallbackConfig: {
                                                ...prev.fallbackConfig,
                                                response: {
                                                    ...prev.fallbackConfig.response,
                                                    headers: newMetadata,
                                                },
                                            },
                                        }));
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
                        const newMetadata = [...(breakerRule?.fallbackConfig?.response?.headers || []), { key: '', value: '' }];
                        setBreakerRule(prev => ({
                            ...prev,
                            fallbackConfig: {
                                ...prev.fallbackConfig,
                                response: {
                                    ...prev.fallbackConfig.response,
                                    headers: newMetadata,
                                },
                            },
                        }));
                    }}>
                    添加标签
                </Button>
            </div>
        </Dialog>
    );


    const fallbackRecover = (
        <>
            <div className={styles.sectionCard} style={{ marginBottom: 24 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                    <span style={{ fontWeight: 600 }}>熔断后降级</span>
                </div>
                <FormItem label={"是否开启"} help={"开启后，当熔断规则触发时，将会返回配置的响应内容"}>
                    {editorState.editable ? (
                        <div>
                            <Switch
                                value={breakerRule?.fallbackConfig?.enable}
                                onChange={(checked) => {
                                    const newRules = { ...breakerRule };
                                    newRules.fallbackConfig = {
                                        ...breakerRule.fallbackConfig,
                                        enable: checked as boolean,
                                    }
                                    setBreakerRule(newRules);
                                }} />
                        </div>
                    ) : renderReadonlySwitch(breakerRule?.fallbackConfig?.enable)}
                </FormItem>
                {breakerRule.fallbackConfig.enable && (
                    <>
                        <FormItem label={"响应码"}>
                            {editorState.editable ? (
                                <div>
                                    <InputNumber
                                        value={breakerRule?.fallbackConfig?.response?.code}
                                        onChange={(val) => {
                                            const newRules = { ...breakerRule };
                                            newRules.fallbackConfig.response.code = val as number;
                                            setBreakerRule(newRules);
                                        }} />
                                </div>
                            ) : renderReadonlyValue(breakerRule?.fallbackConfig?.response?.code ?? '-')}
                        </FormItem>
                        <FormItem label={"响应头"}>
                            <div>
                                {Object.entries(breakerRule?.fallbackConfig?.response?.headers || []).length > 0 ? (
                                    <Space>
                                        {Object.entries(breakerRule?.fallbackConfig?.response?.headers || []).map(([key, value], idx) => (
                                            <Tag key={idx}>
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
                                        onClick={() => setEditorState(prev => ({ ...prev, headerVisible: true }))}
                                    >
                                        <Edit1Icon />
                                    </Button>
                                )}
                                {renderRspHeaderDialog()}
                            </div>
                        </FormItem>
                        <FormItem label={"响应体"}>
                            {editorState.editable ? (
                                <div>
                                    <Textarea
                                        style={{ width: '200%' }}
                                        value={breakerRule?.fallbackConfig?.response?.body || ''}
                                        onChange={(value) => {
                                            const newRules = { ...breakerRule };
                                            newRules.fallbackConfig.response.body = value as string;
                                            setBreakerRule(newRules);
                                        }} />
                                </div>
                            ) : (
                                <pre className={styles.readonlyCodeBlock}>
                                    {breakerRule?.fallbackConfig?.response?.body || '-'}
                                </pre>
                            )}
                        </FormItem>
                    </>
                )}
            </div>
        </>
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
                {callinfo}
                <div className={styles.ruleListSection}>
                    <div className={styles.ruleListHeader}>
                        <span className={styles.listTitle}>熔断策略</span>
                        <div className={styles.ruleHelp}>按错误判断条件和触发条件定义服务进入熔断的策略。</div>
                    </div>
                    <div className={styles.ruleList}>
                    {breakerRule.block_configs.map((rule, idx) => {
                        return renderRule(rule, idx)
                    })}
                    </div>
                    {editorState.editable && (
                        <Button
                            className={styles.addRuleButton}
                            variant="outline" icon={<AddIcon />}
                            onClick={() => {
                                const newRule = { ...breakerRule }
                                newRule.block_configs.push(defaultBlockConfig(newRule.block_configs.length + 1))
                                setBreakerRule(newRule);
                            }}>
                            添加熔断策略
                        </Button>
                    )}
                </div>
                {renderRecover}
                {fallbackRecover}
                {renderPublishForm}
                {renderStickyTool}
            </Form>
        </div>
    );
};

export default React.memo(CircuitBreakerEditor);
