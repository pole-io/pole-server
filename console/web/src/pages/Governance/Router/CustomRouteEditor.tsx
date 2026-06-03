import React from "react";
import { Col, Form, Input, Row, Space, Button, Select, Switch, Dialog, InputNumber, Table, FormProps, Tag, Popup, TableRowData, PrimaryTableProps, StickyTool } from "tdesign-react";
import { SendIcon, AddIcon, CloseIcon, Edit1Icon, SaveIcon, RocketIcon, RollbackIcon } from "tdesign-icons-react";
import cloneDeep from 'lodash/cloneDeep';

import Text from "components/Text";
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
} from "services/router";
import { Label, MatchString, MatchType, MatchTypeMap, MatchTypeOption, MatchValueType, Op } from "services/types";
import { listOneCustomRoute, resetCustomRoute, saveCustomRoutes, selectCustomRoute, updateCustomRoutes } from "modules/governance/route";

import styles from './CustomRouteEditor.module.less';
import { openErrNotification, openInfoNotification } from "utils/notifition";
import PublishForm from "../RuleRelease/PublishForm";
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
        rules: [{
            name: '',
            sources: [defaultMatch()],
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

    const resetCurRule = (rule: CustomRouteView | null) => {
        if (!rule) {
            return;
        }
        const cloneRule = cloneDeep(rule);
        setCustomRouteRule({
            ...cloneRule,
            caller_namespace: cloneRule?.routing_config?.rules?.[0]?.sources[0]?.namespace || '*',
            caller_service: cloneRule?.routing_config?.rules?.[0]?.sources[0]?.service || '',
            callee_namespace: cloneRule?.routing_config?.rules?.[0]?.destinations[0]?.namespace || '*',
            callee_service: cloneRule?.routing_config?.rules?.[0]?.destinations[0]?.service || '',
            metadata: Object.entries(cloneRule?.metadata || {}).map(([key, value]) => ({ key, value })),
        });
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
            routing_config: customRouteRule.routing_config,
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
            refresh(op === 'view'); // 刷新列表
        }
    }

    const addMatch = (ruleIdx: number) => {
        const curRules = customRouteRule.routing_config?.rules || [];
        const newRules = [...curRules];
        newRules[ruleIdx].sources[0].arguments.push(defaultMatchArgs());
        setCustomRouteRule((prev) => ({
            ...prev,
            routing_config: {
                '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                rules: newRules,
            },
        }));
    };

    const updateMatchArgs = (del: boolean, ruleIdx: number, idx: number, args?: RoutingSourceArgument) => {
        const curRules = customRouteRule.routing_config?.rules || [];
        const newRules = [...curRules];
        if (del) {
            newRules[ruleIdx].sources[0].arguments.splice(idx, 1);
        } else {
            if (!newRules[ruleIdx].sources[0].arguments) {
                newRules[ruleIdx].sources[0].arguments = [];
            }
            if (args) {
                if (idx < newRules[ruleIdx].sources[0].arguments.length) {
                    // 更新已有的参数
                    newRules[ruleIdx].sources[0].arguments[idx] = {
                        ...newRules[ruleIdx].sources[0].arguments[idx],
                        ...Object.fromEntries(
                            Object.entries(args).filter(([_, value]) => value !== undefined)
                        ),
                    };
                } else {
                    newRules[ruleIdx].sources[0].arguments.push(args);
                }
            }
        }
        setCustomRouteRule((prev) => ({
            ...prev,
            routing_config: {
                '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                rules: newRules,
            },
        }));
    };

    const addGroup = (ruleIdx: number) => {
        const curRules = customRouteRule.routing_config?.rules || [];
        const newRules = [...curRules];
        const newGroup = defaultGroup(newRules[ruleIdx].destinations.length + 1);
        newRules[ruleIdx].destinations.push(newGroup);
        setCustomRouteRule((prev) => ({
            ...prev,
            routing_config: {
                '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                rules: newRules,
            },
        }));
    };

    const removeGroup = (ruleIdx: number, idx: number) => {
        const curRules = customRouteRule.routing_config?.rules || [];
        const newRules = [...curRules];
        newRules[ruleIdx].destinations.splice(idx, 1);
        setCustomRouteRule((prev) => ({
            ...prev,
            routing_config: {
                '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                rules: newRules,
            },
        }));
    };

    const updateGroup = (ruleIdx: number, idx: number, group: RoutingRuleDestination) => {
        const curRules = customRouteRule.routing_config?.rules || [];
        const newRules = [...curRules];
        if (idx < newRules[ruleIdx].destinations.length) {
            const existingGroup = newRules[ruleIdx].destinations[idx];
            newRules[ruleIdx].destinations[idx] = {
                ...existingGroup,
                ...Object.fromEntries(
                    Object.entries(group).filter(([_, value]) => value !== undefined)
                ),
            };
        } else {
            newRules[ruleIdx].destinations.push(group);
        }
        setCustomRouteRule((prev) => ({
            ...prev,
            routing_config: {
                '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                rules: newRules,
            },
        }));
    }

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
                {customRouteRule.metadata?.map((tag, idx) => (
                    <Row key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Space>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const newMetadata = [...(customRouteRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], key: v };
                                    setCustomRouteRule({ ...customRouteRule, metadata: newMetadata });
                                }} />
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const newMetadata = [...(customRouteRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], value: v };
                                    setCustomRouteRule({ ...customRouteRule, metadata: newMetadata });
                                }} />
                            <Popup trigger="hover" content="删除标签">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newMetadata = [...(customRouteRule.metadata || [])];
                                        newMetadata.splice(idx, 1);
                                        setCustomRouteRule({ ...customRouteRule, metadata: newMetadata });
                                    }}
                                >
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        </Space>
                    </Row>
                ))}
                <Button variant="text" icon={<AddIcon />} onClick={() => {
                    const newMetadata = [...(customRouteRule.metadata || []), { key: '', value: '' }];
                    setCustomRouteRule({ ...customRouteRule, metadata: newMetadata });
                }}>添加标签</Button>
            </div>
        </Dialog>
    );

    // 基础信息表单：第一层名称，第二层优先级与描述，第三层规则标签
    const ruleeditor = (
        <div className={styles.sectionCard}>
            {/* 第一层：规则名称 */}
            <Row style={{ marginBottom: 16 }}>
                <Col span={12}>
                    <FormItem label="规则名称">
                        {op === 'create' ? (
                            <Input
                                maxlength={64}
                                value={customRouteRule.name}
                                onChange={(value) => setCustomRouteRule(prev => ({ ...prev, name: value }))}
                            />
                        ) : (
                            <Text>{customRouteRule.name}</Text>
                        )}
                    </FormItem>
                </Col>
            </Row>
            {/* 第二层：优先级、描述 */}
            <Row style={{ marginBottom: 16 }}>
                <Space>
                    <FormItem label="优先级">
                        {editorState.editable ? (
                            <InputNumber min={0} value={customRouteRule.priority} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, priority: value as number }))} />
                        ) : (
                            <Text>{customRouteRule.priority}</Text>
                        )}
                    </FormItem>
                    <FormItem
                        label="描述"
                        name="description"
                        rules={[{ max: 255, message: '描述长度不能超过255个字符' }]}
                    >
                        {editorState.editable ? (
                            <Input value={customRouteRule.description} onChange={(value) => setCustomRouteRule(prev => ({ ...prev, description: value }))} />
                        ) : (
                            <Text>{customRouteRule.description}</Text>
                        )}
                    </FormItem>
                </Space>
            </Row>
            {/* 第三层：规则标签 */}
            <Row>
                <Col span={12}>
                    <FormItem label='规则标签' name='labels'>
                        <Space align="center">
                            {Array.isArray(customRouteRule.metadata) && customRouteRule.metadata.length > 0 ? (
                                <>
                                    {customRouteRule.metadata.map((item: { key: string; value: string }, idx: number) => (
                                        <Tag key={idx}>{`${item.key}: ${item.value}`}</Tag>
                                    ))}
                                    {editorState.editable && (
                                        <Button
                                            shape="circle"
                                            variant="text"
                                            onClick={() => setEditorState(prev => ({ ...prev, visible: true }))}
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
                                            onClick={() => setEditorState(prev => ({ ...prev, visible: true }))}
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
        </div>
    );

    // 主调/被调服务卡片
    const callinfo = (
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
                >
                    {editorState.editable ? (
                        <>
                            <Select
                                filterable={true}
                                creatable={true}
                                options={namespaceDatas.map(ns => ({
                                    label: ns.name,
                                    value: ns.name,
                                }))}
                                value={customRouteRule.caller_namespace}
                                onChange={(value) => {
                                    setCustomRouteRule(prev => ({ ...prev, caller_namespace: value as string }));
                                }}
                            />
                        </>
                    ) : (
                        <Text>{customRouteRule.caller_namespace}</Text>
                    )}
                </FormItem>
                <FormItem
                    style={{ marginBottom: 0 }}
                    label="服务名称"
                >
                    {editorState.editable ? (
                        <>
                            <Select
                                filterable={true}
                                creatable={true}
                                value={customRouteRule.caller_service}
                                options={serviceDatas.filter(opt => {
                                    if (customRouteRule.caller_namespace === '*') {
                                        return true; // 允许所有命名空间的服务
                                    }
                                    if (customRouteRule.callee_service === opt.name && opt.namespace === customRouteRule.callee_namespace) {
                                        return false; // 如果服务名已选中，则不再显示
                                    }
                                    return opt.namespace === customRouteRule.caller_namespace; // 仅允许当前命名空间
                                }).map(opt => ({
                                    label: opt.name,
                                    value: opt.name,
                                }))}
                                onChange={(value) => {
                                    setCustomRouteRule(prev => ({ ...prev, caller_service: value as string }));
                                }}
                            />
                        </>
                    ) : (
                        <Text>{customRouteRule.caller_service}</Text>
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
                >
                    {editorState.editable ? (
                        <>
                            <Select
                                filterable={true}
                                creatable={true}
                                options={namespaceDatas.map(ns => ({
                                    label: ns.name,
                                    value: ns.name,
                                }))}
                                value={customRouteRule.callee_namespace}
                                onChange={(value) => {
                                    setCustomRouteRule(prev => ({ ...prev, callee_namespace: value as string }));
                                }}
                            />
                        </>
                    ) : (
                        <Text>{customRouteRule.callee_namespace}</Text>
                    )}
                </FormItem>
                <FormItem
                    style={{ marginBottom: 0 }}
                    label="服务名称"
                >
                    {editorState.editable ? (
                        <>
                            <Select
                                filterable={true}
                                creatable={true}
                                options={serviceDatas.filter(opt => {
                                    if (customRouteRule.callee_namespace === '*') {
                                        return true; // 允许所有命名空间的服务
                                    }
                                    if (customRouteRule.caller_service === opt.name && opt.namespace === customRouteRule.caller_namespace) {
                                        return false; // 如果服务名已选中，则不再显示
                                    }
                                    return opt.namespace === customRouteRule.callee_namespace; // 仅允许当前命名空间
                                }).map(opt => ({
                                    label: opt.name,
                                    value: opt.name,
                                }))}
                                value={customRouteRule.callee_service}
                                onChange={(value) => {
                                    setCustomRouteRule(prev => ({ ...prev, callee_service: value as string }));
                                }}
                            />
                        </>
                    ) : (
                        <Text>{customRouteRule.callee_service}</Text>
                    )}
                </FormItem>
            </div>
        </div>
    );

    const trafficTableColumns = (ruleIdx: number): PrimaryTableProps['columns'] => [
        {
            colKey: 'type',
            title: '参数类型',
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
            colKey: 'action',
            title: '操作',
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
        <div>
            <Row align="middle" style={{ marginBottom: 8 }}>
                <Table
                    rowKey="key"
                    data={customRouteRule.routing_config?.rules[ruleIdx].sources[0]?.arguments || []}
                    columns={editorState.editable ? trafficTableColumns(ruleIdx) : trafficTableColumns(ruleIdx)?.filter(col => col.colKey !== 'action')}
                />
                <Col span={2}>
                    {editorState.editable && (
                        <Button variant="text" onClick={() => addMatch(ruleIdx)} icon={<AddIcon />}>添加</Button>
                    )}
                </Col>
            </Row>
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
        {
            colKey: 'name',
            title: '分组名称',
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
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: InputNumber,
                props: {
                    min: 0,
                    max: 100,
                    step: 5,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateGroup(ruleIdx, context.rowIndex, context.newRowData as RoutingRuleDestination);
                },
            }
        },
        {
            colKey: 'labels',
            title: '实例标签',
            cell: ({ row, rowIndex }) => (
                <Space>
                    {Object.entries(row.labels as Record<string, MatchString> || {}).map(([key, value]) => (
                        <Tag key={key}>
                            {`${key} ${MatchTypeMap[value.type as MatchType]} ${value.value}`}
                        </Tag>
                    ))}
                </Space>
            )
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
        <div>
            <Row align="middle" style={{ marginBottom: 8 }}>
                <Table
                    rowKey="key"
                    data={customRouteRule.routing_config?.rules[ruleIdx].destinations || []}
                    columns={editorState.editable ? destinationTableColumns(ruleIdx) : destinationTableColumns(ruleIdx)?.filter(col => col.colKey !== 'action')}
                />
                <Col span={2}>
                    {editorState.editable && (
                        <Button
                            variant="text"
                            onClick={() => addGroup(ruleIdx)}
                            icon={<AddIcon />}>
                            添加实例分组
                        </Button>
                    )}
                </Col>
            </Row>
            {renderTagDialog()}
        </div>
    );

    // 规则区块
    const renderRule = (rule: RoutingRule, ruleIdx: number) => {
        return (
            <div className={styles.sectionCard} key={ruleIdx} style={{ marginBottom: 24 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                    <span style={{ fontWeight: 600 }}>规则 [{ruleIdx + 1}]</span>
                    {editorState.editable && (
                        <Popup trigger="hover" content="删除规则">
                            <Button
                                shape="circle"
                                variant="text"
                                onClick={() =>
                                    setCustomRouteRule((prev) => ({
                                        ...prev,
                                        routing_config: {
                                            '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                                            rules: prev.routing_config?.rules.filter((_, i) => i !== ruleIdx) || [],
                                        },
                                    }))
                                }>
                                <CloseIcon />
                            </Button>
                        </Popup>
                    )}
                </div>
                <div className={styles['route-editor-header-desc']}>
                    来源服务的请求满足以下匹配条件
                </div>
                {renderMatchTable(ruleIdx)}
                <div className={styles['route-editor-header-desc']}>
                    将转发至目标服务的以下实例分组，按优先级成组，可通过调整优先级。
                </div>
                {renderGroupTable(ruleIdx)}
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
                            label={editorState.editable ? '保存' : '编辑'}
                            icon={!editorState.editable ?
                                <Edit1Icon onClick={() => {
                                    setEditorState(prev => ({ ...prev, editable: true }));
                                }} />
                                :
                                <Button type="submit" variant="text" shape="square">
                                    <SaveIcon />
                                </Button>
                            }
                        />
                        {(editorState.editable) && (
                            <StickyItem label="撤销" icon={
                                <RollbackIcon onClick={() => {
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
                            <StickyItem label="发布" icon={
                                <RocketIcon onClick={() => {
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
                {ruleeditor}
                {callinfo}
                <div style={{ margin: '24px 0' }}>
                    {customRouteRule.routing_config?.rules.map((rule, idx) => renderRule(rule, idx))}
                    {editorState.editable && (
                        <Button
                            variant="outline" icon={<AddIcon />}
                            onClick={() => {
                                setCustomRouteRule((prev) => ({
                                    ...prev,
                                    routing_config: {
                                        '@type': 'type.googleapis.com/v1.RuleRoutingConfig',
                                        rules: [...prev.routing_config?.rules || [], {
                                            name: `规则 ${(prev.routing_config?.rules.length || 0) + 1}`,
                                            sources: [defaultMatch()],
                                            destinations: [defaultGroup(1)],
                                        }]
                                    },
                                }));
                            }}>
                            添加规则
                        </Button>
                    )}
                </div>
                {renderPublishForm}
                {renderStickyTool}
            </Form>
        </div>
    );
};

export default React.memo(CustomRouteEditor);