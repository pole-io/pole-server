import React, { } from "react";
import { Col, Form, Input, Row, Space, Button, Select, Switch, InputNumber, Table, FormProps, Tag, Popup, TableRowData, PrimaryTableProps, InputAdornment, RadioGroup, Radio, Textarea, StickyTool } from "tdesign-react";
import { AddIcon, ChevronRightIcon, CloseIcon, Edit1Icon, RocketIcon, RollbackIcon, SaveIcon } from "tdesign-icons-react";

import Text from "components/Text";
import RuleLabelField from "../shared/RuleLabelField";
import shared from "../shared/governance.module.less";
import { useAppDispatch, useAppSelector } from 'modules/store';
import { ServiceView } from "services/service";
import { NamespaceView } from "services/namespace";
import { MatchType, MatchTypeMap, MatchTypeOption, MatchValueType, Op } from "services/types";
import { openErrNotification, openInfoNotification } from "utils/notifition";
import {
    listOneRateLimitRule,
    saveRateLimitRule,
    selectRateLimitRule,
    updateRateLimitRule,
} from 'modules/governance/ratelimit';
import { ConcurrencyAmount, CustomResponse, defaultLimitTriggerView, LimitAction, LimitActionMap, LimitAmountsValidationUnit, LimitAmountsValidationUnitOptions, LimitArgumentsConfig, LimitArgumentsType, LimitArgumentsTypeMap, LimitArgumentsTypeOptions, LimitConfigView, LimitFailover, LimitFailoverMap, LimitType, RateLimitResource, RateLimitResourceMap, RateLimitView } from "services/ratelimit";
import PublishForm from "../RuleRelease/PublishForm";
import RuleStickyAction from "../RuleRelease/RuleStickyAction";
import { PolicySourceType } from "services/auth_policy";
import { cleanNamespacePage, listAllNamespaces, selectNamespace } from "modules/namespace";
import { cleanServicePage, listAllServices, selectService } from "modules/discovery/service";
import cloneDeep from "lodash/cloneDeep";

import styles from './RateLimitEditor.module.less';

const { FormItem } = Form;
const { StickyItem } = StickyTool;
const limitDurationUnitMap = LimitAmountsValidationUnitOptions.reduce((acc, option) => {
    acc[option.value] = option.label;
    return acc;
}, {} as Record<LimitAmountsValidationUnit, string>);

// 定义默认匹配参数
const defaultMatchArgs: () => LimitArgumentsConfig = () => ({
    type: LimitArgumentsType.CUSTOM,
    key: '',
    value: {
        type: MatchType.EXACT,
        value: '',
        value_type: MatchValueType.TEXT
    }
});

// 大规则默认值（含一条默认子规则）
export const defaultRateLimitView = (): RateLimitView => ({
    id: '',
    name: '',
    namespace: '',
    service: '',
    type: LimitType.LOCAL,
    priority: 0,
    disable: false,
    rules: [defaultLimitTriggerView()],
});

// 使用从模块导入的RateLimitRule类型
interface IRateLimitEditorProps {
    limitType: LimitType;
    op: Op; // 操作类型：创建或编辑
    refresh: (close: boolean) => void;
    visible: boolean;
}

const RateLimitEditor: React.FC<IRateLimitEditorProps> = ({ limitType, op, refresh, visible }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { datas: serviceDatas } = serviceState;

    const rateLimitState = useAppSelector(selectRateLimitRule);
    const { editRule, viewRule } = rateLimitState;

    // 大规则状态（含子规则列表）
    const [rateLimit, setRateLimit] = React.useState<RateLimitView>(defaultRateLimitView());
    const [collapsedRuleIndexes, setCollapsedRuleIndexes] = React.useState<Set<number>>(() => new Set());

    const [editorState, setEditorState] = React.useState<{
        visible: boolean;
        editable?: boolean; // 是否可编辑
        model: Op;
        publishView: boolean;
    }>({ model: 'view', visible: false, editable: op === 'create', publishView: false });

    // 初始化数据
    React.useEffect(() => {
        dispatch(listAllNamespaces()).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求错误', `获取命名空间列表失败: ${res.payload as string}`);
            }
        })
        dispatch(listAllServices()).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求错误', `获取服务列表失败: ${res.payload as string}`);
            }
        });

        return () => {
            dispatch(cleanNamespacePage());
            dispatch(cleanServicePage());
        }
    }, []);

    React.useEffect(() => {
        if (editRule) {
            if (editRule.id !== '') {
                dispatch(listOneRateLimitRule({ id: editRule.id || '' })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        const ret = res.payload as { viewRule: RateLimitView } | null;
                        resetCurRule(ret?.viewRule ?? null);
                    } else {
                        openErrNotification('请求错误', `获取限流规则详情失败: ${res.payload as string}`);
                    }
                });
            } else {
                resetCurRule(editRule);
            }
        }
    }, [editRule]);

    const resetCurRule = (rule: RateLimitView | null) => {
        if (!rule) return;
        const cloneRule = cloneDeep(rule);
        const rules = (cloneRule.rules && cloneRule.rules.length > 0) ? cloneRule.rules : [defaultLimitTriggerView()];
        setRateLimit({
            ...cloneRule,
            name: cloneRule.name || '',
            service: cloneRule.service || '',
            namespace: cloneRule.namespace || '',
            type: cloneRule.type || LimitType.LOCAL,
            priority: cloneRule.priority ?? 0,
            disable: cloneRule.disable ?? false,
            rules,
        });
    };

    // 更新某条子规则
    const updateRule = (ruleIdx: number, patch: Partial<RateLimitView['rules'][0]>) => {
        setRateLimit(prev => ({
            ...prev,
            rules: prev.rules.map((r, i) => i === ruleIdx ? { ...r, ...patch } : r),
        }));
    };

    // 添加子规则
    const addRule = () => {
        setRateLimit(prev => ({ ...prev, rules: [...prev.rules, defaultLimitTriggerView()] }));
    };

    // 删除子规则（至少保留一条）
    const removeRule = (ruleIdx: number) => {
        if (rateLimit.rules.length <= 1) return;
        setRateLimit(prev => ({ ...prev, rules: prev.rules.filter((_, i) => i !== ruleIdx) }));
    };

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

    // 某条子规则：添加匹配条件
    const addMatch = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        updateRule(ruleIdx, { arguments: [...(trigger.arguments || []), defaultMatchArgs()] });
    };

    const updateMatchArgs = (ruleIdx: number, del: boolean, idx: number, args: LimitArgumentsConfig) => {
        const trigger = rateLimit.rules[ruleIdx];
        const list = trigger.arguments || [];
        if (del) {
            updateRule(ruleIdx, { arguments: list.filter((_, i) => i !== idx) });
        } else if (args) {
            const newArgs = [...list];
            if (idx < newArgs.length) {
                newArgs[idx] = { ...newArgs[idx], ...Object.fromEntries(Object.entries(args).filter(([_, v]) => v !== undefined && v !== null)) };
            } else {
                newArgs.push(args);
            }
            updateRule(ruleIdx, { arguments: newArgs });
        }
    };

    const addLimit = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        updateRule(ruleIdx, {
            amounts: [...trigger.amounts, { validDuration: 1, validDurationUnit: LimitAmountsValidationUnit.s, maxAmount: 1 }],
        });
    };

    const updateLimit = (ruleIdx: number, amountIdx: number, limit: LimitConfigView) => {
        const trigger = rateLimit.rules[ruleIdx];
        const newLimits = [...trigger.amounts];
        newLimits[amountIdx] = { ...newLimits[amountIdx], ...Object.fromEntries(Object.entries(limit).filter(([_, v]) => v !== undefined && v !== null)) };
        updateRule(ruleIdx, { amounts: newLimits });
    };

    const removeLimit = (ruleIdx: number, amountIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        if (trigger.amounts.length <= 1) return;
        updateRule(ruleIdx, { amounts: trigger.amounts.filter((_, i) => i !== amountIdx) });
    };

    // 表单提交：大规则 + 子规则列表
    const onSubmit: FormProps['onSubmit'] = async () => {
        const data: RateLimitView = {
            ...rateLimit,
            id: viewRule?.id || rateLimit.id,
            name: rateLimit.name,
            service: rateLimit.service,
            namespace: rateLimit.namespace,
            type: limitType,
            priority: rateLimit.priority ?? 0,
            disable: rateLimit.disable ?? false,
            rules: rateLimit.rules,
        };
        const res = op === 'view'
            ? await dispatch(updateRateLimitRule({ param: data }))
            : await dispatch(saveRateLimitRule({ param: data }));
        if (res.meta.requestStatus === 'rejected') {
            openErrNotification('请求错误', `${op === 'view' ? '修改' : '创建'}限流规则失败: ${res.payload as string}`);
        } else {
            openInfoNotification('请求成功', op === 'view' ? '修改限流规则成功' : '创建限流规则成功');
            if (op === 'view') {
                setEditorState(prev => ({ ...prev, editable: false }));
            }
            refresh(false);
        }
    };

    const getMatchTableColumns = (ruleIdx: number): PrimaryTableProps['columns'] => [
        {
            colKey: 'type',
            title: '参数类型',
            edit: {
                keepEditMode: editorState.editable,
                showEditIcon: editorState.editable,
                component: Select,
                props: { clearable: true, options: LimitArgumentsTypeOptions },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(ruleIdx, false, context.rowIndex, context.newRowData as LimitArgumentsConfig);
                },
                validateTrigger: 'change',
            },
            cell: ({ row }) => <Text>{LimitArgumentsTypeMap[row.type]}</Text>
        },
        { colKey: 'key', title: '参数键', edit: { keepEditMode: editorState.editable, showEditIcon: editorState.editable, component: Input, props: { clearable: true }, onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => { updateMatchArgs(ruleIdx, false, context.rowIndex, context.newRowData as LimitArgumentsConfig); }, validateTrigger: 'change' } },
        { colKey: 'value.type', title: '匹配类型', edit: { keepEditMode: editorState.editable, showEditIcon: editorState.editable, component: Select, props: { clearable: true, options: MatchTypeOption }, onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => { updateMatchArgs(ruleIdx, false, context.rowIndex, context.newRowData as LimitArgumentsConfig); }, validateTrigger: 'change' }, cell: ({ row }) => <Text>{MatchTypeMap[row.value?.type as MatchType]}</Text> },
        { colKey: 'value.value', title: '匹配值', edit: { keepEditMode: editorState.editable, showEditIcon: editorState.editable, component: Input, props: { clearable: true }, onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => { updateMatchArgs(ruleIdx, false, context.rowIndex, context.newRowData as LimitArgumentsConfig); }, validateTrigger: 'change' } },
        {
            colKey: 'action',
            title: '操作',
            cell: ({ rowIndex }) => (
                <Popup trigger="hover" content="删除参数">
                    <Button shape="circle" variant="text" onClick={() => updateMatchArgs(ruleIdx, true, rowIndex, {} as LimitArgumentsConfig)}><CloseIcon /></Button>
                </Popup>
            ),
        }
    ];

    const renderMatchTable = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        const args = trigger.arguments || [];
        return (
            <div className={styles.compactTable}>
                <Table
                    key={`match-${ruleIdx}-${editorState.editable ? 'edit' : 'view'}`}
                    rowKey={(row) => `arg-${ruleIdx}-${(row as LimitArgumentsConfig).type}-${(row as LimitArgumentsConfig).key}`}
                    tableLayout="fixed"
                    data={args.map(arg => ({ ...arg, type: arg.type || LimitArgumentsType.CUSTOM }))}
                    columns={editorState.editable ? getMatchTableColumns(ruleIdx) : getMatchTableColumns(ruleIdx).filter(col => col.colKey !== 'action')}
                />
                {editorState.editable && <Button className={styles.inlineAdd} variant="text" onClick={() => addMatch(ruleIdx)} icon={<AddIcon />}>添加</Button>}
            </div>
        );
    };

    const renderLimitTable = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        const limitsWithKey = trigger.amounts.map((limit, index) => ({ ...limit, key: `limit-${ruleIdx}-${index}` }));
        const columns: PrimaryTableProps['columns'] = [
            {
                colKey: 'validDuration',
                title: '窗口',
                cell: ({ row }) => {
                    const unit = row.validDurationUnit || LimitAmountsValidationUnit.s;
                    return <Text>{`${row.validDuration ?? '-'} ${limitDurationUnitMap[unit as LimitAmountsValidationUnit] || unit}`}</Text>;
                },
                edit: {
                    keepEditMode: editorState.editable,
                    showEditIcon: editorState.editable,
                    component: InputAdornment,
                    props: (value: { key: string; validDuration: number; validDurationUnit: LimitAmountsValidationUnit; maxAmount: number }) => ({
                        append: (
                            <Select
                                autoWidth
                                options={LimitAmountsValidationUnitOptions}
                                value={value.validDurationUnit || LimitAmountsValidationUnit.s}
                                onChange={(v) => {
                                    const amountIdx = Number.parseInt(value.key.split('-')[2], 10);
                                    updateLimit(ruleIdx, amountIdx, { maxAmount: value.maxAmount, validDuration: value.validDuration, validDurationUnit: v as LimitAmountsValidationUnit });
                                }}
                            />
                        ),
                        children: <Input type="number" />
                    }),
                    onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                        updateLimit(ruleIdx, context.rowIndex, context.newRowData as LimitConfigView);
                    },
                },
            },
            {
                colKey: 'maxAmount',
                title: '最大请求数',
                edit: {
                    keepEditMode: editorState.editable,
                    showEditIcon: editorState.editable,
                    component: InputAdornment,
                    props: { append: '次', children: <Input type="number" /> },
                    onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                        updateLimit(ruleIdx, context.rowIndex, context.newRowData as LimitConfigView);
                    },
                },
            },
            {
                colKey: 'action',
                title: '操作',
                cell: ({ rowIndex }) => (
                    <Popup trigger="hover" content="删除配置">
                        <Button shape="circle" variant="text" disabled={trigger.amounts.length <= 1} onClick={() => removeLimit(ruleIdx, rowIndex)}><CloseIcon /></Button>
                    </Popup>
                ),
            }
        ];
        const qpsTable = (
            <div className={styles.compactTable}>
                <Table
                    key={`limit-${ruleIdx}-${editorState.editable ? 'edit' : 'view'}`}
                    rowKey="key"
                    data={limitsWithKey}
                    columns={editorState.editable ? columns : columns.filter(col => col.colKey !== 'action')}
                    tableLayout="fixed"
                />
                {editorState.editable && <Button className={styles.inlineAdd} variant="text" onClick={() => addLimit(ruleIdx)} icon={<AddIcon />}>添加</Button>}
            </div>
        );
        const concurrencyTable = (
            <div className={styles.compactTable}>
                <Table
                    key={`concurrency-${ruleIdx}-${editorState.editable ? 'edit' : 'view'}`}
                    rowKey="key"
                    data={limitsWithKey}
                    columns={[
                        { colKey: 'concurrency', title: '指标类型', cell: () => <Text>并发数</Text> },
                        {
                            colKey: 'maxAmount',
                            title: '阈值',
                            edit: {
                                keepEditMode: true,
                                component: InputNumber,
                                props: { min: 1, step: 1 },
                                onEdited: (context: { newRowData: TableRowData }) => {
                                    updateRule(ruleIdx, { concurrencyAmount: { maxAmount: context.newRowData.maxAmount as number } });
                                },
                            },
                        }
                    ]}
                    tableLayout="fixed"
                />
            </div>
        );
        return trigger.resource === RateLimitResource.QPS ? qpsTable : concurrencyTable;
    };

    const namespaceSelectOptions = namespaceDatas.map((ns: NamespaceView) => ({ label: ns.name, value: ns.name }));
    const serviceSelectOptions = serviceDatas
        .filter((opt: ServiceView) => rateLimit.namespace === '*' || opt.namespace === rateLimit.namespace)
        .map((s: ServiceView) => ({ label: s.name, value: s.name }));

    // 统一基础信息区
    const ruleBaseInfo = (
        <div className={shared.section}>
            <div className={shared.sectionHeader}>基础信息</div>
            <div className={shared.sectionBody}>
                <div className={shared.infoGrid}>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>规则名称</div>
                        {editorState.editable
                            ? <Input maxlength={64} value={rateLimit.name} onChange={(value) => setRateLimit(prev => ({ ...prev, name: value }))} />
                            : <div className={shared.fieldValue}>{rateLimit.name || '-'}</div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>运行状态</div>
                        {editorState.editable
                            ? <Switch label={['启用', '停用']} value={!rateLimit.disable} onChange={(v) => setRateLimit(prev => ({ ...prev, disable: !(v as boolean) }))} />
                            : <div className={shared.fieldValue}><span className={`${shared.pill} ${rateLimit.disable ? shared.pillOff : shared.pillOk} ${shared.pillDot}`}>{rateLimit.disable ? '停用' : '启用'}</span></div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>优先级</div>
                        {editorState.editable
                            ? <InputNumber min={0} value={rateLimit.priority ?? 0} onChange={(value) => setRateLimit(prev => ({ ...prev, priority: (value as number) ?? 0 }))} />
                            : <div className={shared.fieldValue}>{rateLimit.priority ?? 0}</div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>命名空间</div>
                        {editorState.editable
                            ? <Select filterable creatable options={namespaceSelectOptions} value={rateLimit.namespace} onChange={(value) => setRateLimit(prev => ({ ...prev, namespace: value as string }))} />
                            : <div className={shared.fieldValue}>{rateLimit.namespace || '-'}</div>}
                    </div>
                    <div className={shared.field}>
                        <div className={shared.fieldLabel}>服务名称</div>
                        {editorState.editable
                            ? <Select filterable creatable options={serviceSelectOptions} value={rateLimit.service} onChange={(value) => setRateLimit(prev => ({ ...prev, service: value as string }))} />
                            : <div className={shared.fieldValue}>{rateLimit.service || '-'}</div>}
                    </div>
                    {limitType === LimitType.GLOBAL && (
                        <div className={shared.field}>
                            <div className={shared.fieldLabel}>限流集群</div>
                            <div className={shared.fieldValue}>{`${rateLimit.namespace || '-'} / ${rateLimit.service || '-'}`}</div>
                        </div>
                    )}
                    <div className={`${shared.field} ${shared.full}`}>
                        <div className={shared.fieldLabel}>规则标签</div>
                        <RuleLabelField
                            metadata={rateLimit.metadata}
                            editable={editorState.editable}
                            onChange={(next) => setRateLimit(prev => ({ ...prev, metadata: next }))}
                        />
                    </div>
                </div>
            </div>
        </div>
    );

    const matchInterface = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        return (
            <div className={shared.step} data-step="1">
                <div className={shared.stepTitle}>匹配接口<span className={shared.stepHint}>满足该接口的请求才进入限流统计</span></div>
                <div className={shared.stepContent}>
                    <InputAdornment
                        append={
                            <Select
                                autoWidth
                                options={MatchTypeOption}
                                value={trigger.method?.type}
                                readonly={!editorState.editable}
                                onChange={(value) => updateRule(ruleIdx, { method: { ...trigger.method, type: value as MatchType } })}
                            />
                        }
                    >
                        <Input
                            readonly={!editorState.editable}
                            value={trigger.method?.value}
                            onChange={(val) => updateRule(ruleIdx, { method: { ...trigger.method, value: val as string } })}
                        />
                    </InputAdornment>
                </div>
            </div>
        );
    };

    const matchCondition = (ruleIdx: number) => {
        return (
            <div className={shared.step} data-step="2">
                <div className={shared.stepTitle}>匹配条件<span className={shared.stepHint}>满足以下条件的请求将应用该规则</span></div>
                <div className={shared.stepContent}>{renderMatchTable(ruleIdx)}</div>
            </div>
        );
    };

    const rateLimitBlock = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        return (
            <div className={shared.step} data-step="3">
                <div className={shared.stepTitle}>限流方式<span className={shared.stepHint}>设置限流指标与阈值进行流量控制</span></div>
                <div className={shared.stepContent}>
                    <div className={styles.resourceSwitch}>
                        <RadioGroup
                            theme="button"
                            variant="primary-filled"
                            value={trigger.resource}
                            readonly={!editorState.editable}
                            onChange={(value) => updateRule(ruleIdx, { resource: value as RateLimitResource })}
                        >
                            <Radio.Button value={RateLimitResource.QPS}>{RateLimitResourceMap[RateLimitResource.QPS]}</Radio.Button>
                            <Radio.Button disabled={limitType === LimitType.GLOBAL} value={RateLimitResource.Concurrency}>{RateLimitResourceMap[RateLimitResource.Concurrency]}</Radio.Button>
                        </RadioGroup>
                    </div>
                    {renderLimitTable(ruleIdx)}
                    <div className={styles.inlineConfig}>
                        <FormItem label="阈值计算合并">
                            <Switch value={trigger.regex_combine} disabled={!editorState.editable} onChange={(value) => updateRule(ruleIdx, { regex_combine: value as boolean })} />
                        </FormItem>
                        <div className={styles.inlineHint}>
                            {trigger.regex_combine ? '正则匹配的请求将合并计算限流' : '正则匹配的请求将单独计算限流'}
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    const limitActionBlock = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        const renderUnirate = trigger.action === LimitAction.UNIRATE && (
            <Row style={{ marginTop: 16 }}>
                <Col span={12}>
                    <FormItem label="最大排队时长(秒)">
                        <InputAdornment append="秒">
                            <InputNumber
                                min={1}
                                step={1}
                                value={trigger.max_queue_delay ?? 1}
                                readonly={!editorState.editable}
                                inputProps={{
                                    borderless: !editorState.editable,
                                }}
                                onChange={(value) => updateRule(ruleIdx, { max_queue_delay: (value as number) ?? 1 })}
                            />
                        </InputAdornment>
                    </FormItem>
                </Col>
            </Row>
        );
        const renderFailover = limitType === LimitType.GLOBAL && (
            <Row style={{ marginTop: 16 }}>
                <Space>
                    <FormItem label="失败处理策略">
                        <RadioGroup theme="button" variant="primary-filled" value={trigger.failover} readonly={!editorState.editable} onChange={(value) => updateRule(ruleIdx, { failover: value as LimitFailover })}>
                            <Radio.Button value={LimitFailover.FAILOVER_LOCAL}>{LimitFailoverMap[LimitFailover.FAILOVER_LOCAL]}</Radio.Button>
                            <Radio.Button value={LimitFailover.FAILOVER_PASS}>{LimitFailoverMap[LimitFailover.FAILOVER_PASS]}</Radio.Button>
                        </RadioGroup>
                    </FormItem>
                    <div style={{ color: '#999', fontSize: 12, marginTop: 4 }}>当出现通信失败或 Token Server 不可用时，退化为单机限流</div>
                </Space>
            </Row>
        );
        const isReject = trigger.action !== LimitAction.UNIRATE;
        return (
            <div className={shared.step} data-step="4">
                <div className={shared.stepTitle}>限流效果<span className={shared.stepHint}>超过阈值的请求按该策略处理</span></div>
                <div className={shared.stepContent}>
                    {editorState.editable && (
                        <div className={styles.resourceSwitch}>
                            <RadioGroup theme="button" variant="primary-filled" value={trigger.action} onChange={(value) => updateRule(ruleIdx, { action: value as LimitAction })}>
                                <Radio.Button value={LimitAction.REJECT}>{LimitActionMap[LimitAction.REJECT]}</Radio.Button>
                                {trigger.resource === RateLimitResource.QPS && limitType === LimitType.LOCAL && (
                                    <Radio.Button value={LimitAction.UNIRATE}>{LimitActionMap[LimitAction.UNIRATE]}</Radio.Button>
                                )}
                            </RadioGroup>
                        </div>
                    )}
                    <div className={`${shared.verdict} ${isReject ? shared.verdictDeny : shared.verdictWarn}`}>
                        <div className={shared.verdictIcon}>{isReject ? <CloseIcon /> : <RollbackIcon />}</div>
                        <div>
                            <div className={shared.verdictHead}>{LimitActionMap[trigger.action] || (isReject ? '直接拒绝' : '匀速排队')}</div>
                            <div className={shared.verdictDesc}>{isReject ? '超过阈值的请求立即拒绝，返回下方响应' : '超额请求进入队列匀速放行，超时后拒绝'}</div>
                        </div>
                    </div>
                    {renderUnirate}
                    {renderFailover}
                    <div className={styles.responseField}>
                        <div className={shared.editLabel}>触发限流后的响应</div>
                        {editorState.editable ? (
                            <Textarea value={trigger?.customResponse?.body} onChange={(v) => updateRule(ruleIdx, { customResponse: { ...trigger.customResponse, body: v } })} />
                        ) : (
                            <div className={`${shared.fieldValue} ${shared.mono}`}>{trigger?.customResponse?.body || '无自定义响应'}</div>
                        )}
                    </div>
                </div>
            </div>
        );
    };

    const describeLimitAmount = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        if (trigger.resource === RateLimitResource.Concurrency) {
            return `${trigger.concurrencyAmount?.maxAmount ?? trigger.amounts?.[0]?.maxAmount ?? '-'} 并发`;
        }
        const firstAmount = trigger.amounts?.[0];
        if (!firstAmount) {
            return '未配置阈值';
        }
        return `${firstAmount.maxAmount} 次 / ${firstAmount.validDuration}${firstAmount.validDurationUnit}`;
    };

    const renderRateLimitRule = (ruleIdx: number) => {
        const trigger = rateLimit.rules[ruleIdx];
        const collapsed = collapsedRuleIndexes.has(ruleIdx);
        const matchCount = trigger.arguments?.length || 0;
        const amountCount = trigger.resource === RateLimitResource.QPS ? (trigger.amounts?.length || 0) : 1;
        return (
            <div className={`${shared.policy} ${collapsed ? shared.policyCollapsed : ''}`} key={ruleIdx}>
                <div className={shared.policyHead} onClick={() => toggleRuleCollapsed(ruleIdx)}>
                    <div className={shared.policyHeadMain}>
                        <span className={shared.caret}><ChevronRightIcon /></span>
                        <div>
                            <div className={shared.policyIndex}>规则 [{ruleIdx + 1}]</div>
                            <div className={shared.policySummary}>
                                {matchCount} 个匹配条件 / {RateLimitResourceMap[trigger.resource] || '限流'} / {amountCount} 个阈值 / {LimitActionMap[trigger.action] || trigger.action}
                            </div>
                        </div>
                    </div>
                    <div className={shared.policyHeadActions} onClick={(e) => e.stopPropagation()}>
                        <Tag variant="light">{describeLimitAmount(ruleIdx)}</Tag>
                        {editorState.editable && (
                            <Popup trigger="hover" content="删除规则">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => removeRule(ruleIdx)}
                                    disabled={rateLimit.rules.length <= 1}
                                >
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        )}
                    </div>
                </div>
                <div className={shared.policyBody}>
                    {matchInterface(ruleIdx)}
                    {matchCondition(ruleIdx)}
                    {rateLimitBlock(ruleIdx)}
                    {limitActionBlock(ruleIdx)}
                </div>
            </div>
        );
    };

    const renderPublishForm = (
        <>
            {editorState.publishView && (
                <PublishForm
                    ruleId={viewRule?.id || ''}
                    ruleName={viewRule?.name || ''}
                    resource={PolicySourceType.RateLimitRules}
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
                            }
                            resetCurRule(viewRule ?? null);
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
        </>
    )

    return (
        <div style={{ padding: 24 }}>
            {visible && (
                <Form
                    form={form}
                    onSubmit={onSubmit}
                    layout="vertical"
                    labelAlign="left"
                    labelWidth={120}
                    colon
                >
                    {ruleBaseInfo}
                    <div className={shared.section}>
                        <div className={shared.sectionHeader}>
                            <span>限流规则</span>
                            <span className={shared.countTag}>{rateLimit.rules.length} 条</span>
                        </div>
                        <div className={shared.sectionBody}>
                            <div className={shared.ruleList}>
                                {rateLimit.rules.map((_, ruleIdx) => renderRateLimitRule(ruleIdx))}
                            </div>
                            {editorState.editable && (
                                <Button className={shared.addRuleButton} style={{ marginTop: 12 }} variant="dashed" icon={<AddIcon />} onClick={addRule}>添加规则</Button>
                            )}
                        </div>
                    </div>
                    {renderPublishForm}
                    {renderStickyTool}
                </Form>
            )}
        </div>
    );
};

export default React.memo(RateLimitEditor);
