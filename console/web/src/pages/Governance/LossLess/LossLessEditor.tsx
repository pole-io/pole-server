
import {
    Form,
    Input,
    Select,
    InputNumber,
    Switch,
    Button,
    Space,
    Divider,
    StickyTool,
    FormProps,
    Row,
    Col,
    Tag,
    Dialog,
    Popup,
    RadioGroup,
} from 'tdesign-react';
import Text from 'components/Text';
import { Edit1Icon, SaveIcon, RollbackIcon, RocketIcon, CloseIcon, AddIcon } from 'tdesign-icons-react';
import React from 'react';
import { HTTPMethodOption, Label, Op } from 'services/types';
import { cleanNamespacePage, listAllNamespaces, selectNamespace } from 'modules/namespace';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { cleanServicePage, listAllServices, selectService } from 'modules/discovery/service';
import { listOneLossLessRule, saveLossLessRule, selectLosslessRule, updateLosslessRule } from 'modules/governance/lossless';
import { LossLessRuleView } from 'services/lossless';
import PublishForm from '../RuleRelease/PublishForm';
import RuleStickyAction from '../RuleRelease/RuleStickyAction';
import { PolicySourceType } from 'services/auth_policy';
import { ServiceView } from 'services/service';
import { NamespaceView } from 'services/namespace';
import { set } from 'lodash';

const { FormItem } = Form;
const { StickyItem } = StickyTool;

export interface LossLessRule {
    id?: string;
    service: string;
    namespace: string;
    lossless_online: {
        delay_register: {
            enable: boolean;
            strategy: string;
            interval: number;
            health_check_protocol?: string;
            health_check_method?: string;
            health_check_path?: string;
            health_check_interval?: number;
        };
        warmup: {
            enable: boolean;
            interval: number;
            enable_overload_protection: boolean;
            overload_protection_threshold: number;
            curvature: number;
        };
    };
    lossless_offline: {
        enable: boolean;
        interval: number;
    };
    metadata?: Label[];
}

const defaultLossLess = (): LossLessRule => ({
    service: '',
    namespace: '',
    lossless_online: {
        delay_register: {
            enable: false,
            strategy: 'DELAY_BY_TIME',
            interval: 30,
            health_check_protocol: 'http',
            health_check_method: 'GET',
            health_check_path: '/health',
            health_check_interval: 10,
        },
        warmup: {
            enable: false,
            interval: 300,
            enable_overload_protection: false,
            overload_protection_threshold: 50,
            curvature: 2,
        },
    },
    lossless_offline: {
        enable: false,
        interval: 30,
    },
    metadata: [],
})

export interface LossLessEditorProps {
    op: Op;
    visible: boolean;
    refresh: (close: boolean) => void;
}

const LossLessEditor: React.FC<LossLessEditorProps> = ({ op, refresh }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { datas: serviceDatas } = serviceState;

    const losslessState = useAppSelector(selectLosslessRule);
    const { editRule, viewRule } = losslessState;

    // 创建规则状态
    const [losslessRule, setLosslessRule] = React.useState<LossLessRule>(defaultLossLess());

    const [editor, setEditor] = React.useState<{
        visible: boolean;
        editable?: boolean; // 是否可编辑
        model: Op;
        publishView: boolean;
    }>({ model: 'view', visible: false, editable: op === 'create', publishView: false });

    const renderReadonlySwitch = (enabled?: boolean) => (
        <Tag theme={enabled ? 'success' : 'default'} variant="light">
            {enabled ? '开启' : '关闭'}
        </Tag>
    );

    const renderReadonlyValue = (value: React.ReactNode) => (
        <Text>{value ?? '-'}</Text>
    );

    const renderSeconds = (value?: number) => renderReadonlyValue(value === undefined || value === null ? '-' : `${value} 秒`);

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
                dispatch(listOneLossLessRule({ id: editRule.id || '' })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        const ret = res.payload as { viewRule: LossLessRuleView } | null
                        resetCurRule(ret?.viewRule || null);
                    } else {
                        openErrNotification('请求错误', `获取无损规则详情失败: ${res.payload as string}`);
                    }
                })
            } else {
                resetCurRule(editRule);
            }
        }
    }, [editRule])

    const resetCurRule = (rule: LossLessRuleView | null) => {
        if (!rule) {
            return;
        }
        setLosslessRule({
            id: rule.id,
            service: rule.service,
            namespace: rule.namespace,
            lossless_online: {
                delay_register: {
                    enable: rule.lossless_online.delay_register.enable,
                    strategy: rule.lossless_online.delay_register.strategy,
                    interval: parseInt((rule.lossless_online.delay_register.interval || '0').replace(/s$/, '')) || 0,
                    health_check_protocol: rule.lossless_online.delay_register.health_check_protocol,
                    health_check_method: rule.lossless_online.delay_register.health_check_method,
                    health_check_path: rule.lossless_online.delay_register.health_check_path,
                    health_check_interval: parseInt((rule.lossless_online.delay_register.health_check_interval || '0').replace(/s$/, '')) || 0,
                },
                warmup: {
                    enable: rule.lossless_online.warmup.enable,
                    interval: parseInt((rule.lossless_online.warmup.interval || '0').replace(/s$/, '')) || 0,
                    enable_overload_protection: rule.lossless_online.warmup.enable_overload_protection,
                    overload_protection_threshold: rule.lossless_online.warmup.overload_protection_threshold,
                    curvature: rule.lossless_online.warmup.curvature,
                },
            },
            lossless_offline: {
                enable: rule.lossless_offline.enable,
                interval: parseInt((rule.lossless_offline.interval || '0').replace(/s$/, '')) || 0,
            },
            metadata: rule.metadata ? Object.entries(rule.metadata).map(([key, value]) => ({ key, value })) : [],
        });
    }

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }
        const saveAction = op === 'create' ? saveLossLessRule : updateLosslessRule;
        dispatch(saveAction({
            param: {
                id: losslessRule.id,
                service: losslessRule.service,
                namespace: losslessRule.namespace,
                lossless_online: {
                    delay_register: {
                        enable: losslessRule.lossless_online.delay_register.enable,
                        strategy: losslessRule.lossless_online.delay_register.strategy,
                        interval: losslessRule.lossless_online.delay_register.interval == 0 ? "0s" : `${losslessRule.lossless_online.delay_register.interval}s`,
                        health_check_protocol: losslessRule.lossless_online.delay_register.health_check_protocol,
                        health_check_method: losslessRule.lossless_online.delay_register.health_check_method,
                        health_check_path: losslessRule.lossless_online.delay_register.health_check_path,
                        health_check_interval: losslessRule.lossless_online.delay_register.health_check_interval == 0 ? "0s" : `${losslessRule.lossless_online.delay_register.health_check_interval}s`,
                    },
                    warmup: {
                        enable: losslessRule.lossless_online.warmup.enable,
                        interval: losslessRule.lossless_online.warmup.interval == 0 ? "0s" : `${losslessRule.lossless_online.warmup.interval}s`,
                        enable_overload_protection: losslessRule.lossless_online.warmup.enable_overload_protection,
                        overload_protection_threshold: losslessRule.lossless_online.warmup.overload_protection_threshold,
                        curvature: losslessRule.lossless_online.warmup.curvature,
                    },
                },
                lossless_offline: {
                    enable: losslessRule.lossless_offline.enable,
                    interval: losslessRule.lossless_offline.interval == 0 ? "0s" : `${losslessRule.lossless_offline.interval}s`,
                },
                metadata: losslessRule.metadata?.reduce((acc, cur) => {
                    acc[cur.key] = cur.value;
                    return acc;
                }, {} as Record<string, string>)
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'fulfilled') {
                openInfoNotification("请求成功", "保存无损规则成功");
                refresh(true);
            } else {
                openErrNotification("请求失败", `保存无损规则失败: ${res.payload as string || '未知'}`);
            }
        });
    };


    // 标签弹窗渲染
    const renderRuleLabelsDialog = () => (
        <Dialog
            visible={editor.visible}
            header="编辑规则标签"
            width={700}
            onConfirm={() => setEditor(prev => ({ ...prev, visible: false }))}
            onClose={() => setEditor(prev => ({ ...prev, visible: false }))}
        >
            <div>
                {losslessRule.metadata?.map((tag, idx) => (
                    <Row gutter={8} key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Col span={4}>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const newMetadata = [...(losslessRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], key: v };
                                    setLosslessRule({ ...losslessRule, metadata: newMetadata });
                                }} />
                        </Col>
                        <Col span={4}>
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const newMetadata = [...(losslessRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], value: v };
                                    setLosslessRule({ ...losslessRule, metadata: newMetadata });
                                }} />
                        </Col>
                        <Col span={2}>
                            <Popup trigger="hover" content="删除标签">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newMetadata = [...(losslessRule.metadata || [])];
                                        newMetadata.splice(idx, 1);
                                        setLosslessRule({ ...losslessRule, metadata: newMetadata });
                                    }}
                                >
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        </Col>
                    </Row>
                ))}
                <Button variant="text" icon={<AddIcon />} onClick={() => {
                    const newMetadata = [...(losslessRule.metadata || []), { key: '', value: '' }];
                    setLosslessRule({ ...losslessRule, metadata: newMetadata });
                }}>添加标签</Button>
            </div>
        </Dialog>
    );

    // 基础信息卡片样式与各规则页统一（第一层标题，第二层命名空间与服务名称，第三层规则标签）
    const baseInfoCardStyle = { marginBottom: 24, padding: '24px 32px', border: '1px solid #e5e6eb', borderRadius: 8, boxShadow: '0 2px 8px 0 rgba(0,0,0,0.03)' };
    const serviceInfo = (
        <div style={baseInfoCardStyle}>
            {/* 第一层：标题 */}
            <div style={{ marginBottom: 16 }}>
                <Text style={{ fontWeight: 'bold' }}>目标服务</Text>
            </div>
            {/* 第二层：命名空间、服务名称 */}
            <Row style={{ marginBottom: 16 }}>
                <Space>
                    <FormItem label="命名空间">
                        <div>
                            <Select
                                filterable={true}
                                creatable={true}
                                value={losslessRule.namespace}
                                borderless={!editor.editable}
                                readonly={!editor.editable}
                                options={namespaceDatas.map((ns: NamespaceView) => ({
                                    label: ns.name,
                                    value: ns.name,
                                    namespace: ns.name,
                                }))}
                                onChange={(value) => {
                                    setLosslessRule(prev => ({ ...prev, namespace: value as string }));
                                }}
                            />
                        </div>
                    </FormItem>
                    <FormItem label="服务名称">
                        <div>
                            <Select
                                filterable={true}
                                creatable={true}
                                value={losslessRule.service}
                                borderless={!editor.editable}
                                readonly={!editor.editable}
                                options={serviceDatas.filter(opt => {
                                    if (losslessRule.namespace === '*') {
                                        return true;
                                    }
                                    return opt.namespace === losslessRule.namespace;
                                }).map((service: ServiceView) => ({
                                    label: service.name,
                                    value: service.name,
                                    namespace: service.namespace,
                                }))}
                                onChange={(value) => {
                                    setLosslessRule(prev => ({ ...prev, service: value as string }));
                                }}
                            />
                        </div>
                    </FormItem>
                </Space>
            </Row>
            {/* 第三层：规则标签 */}
            <Row>
                <FormItem label='规则标签' name='labels'>
                    <Space align="center">
                        {losslessRule.metadata && losslessRule.metadata.length > 0 ? (
                            <>
                                {losslessRule.metadata.map((tag, idx) => (
                                    <Tag key={idx}>{`${tag.key}: ${tag.value}`}</Tag>
                                ))}
                                {editor.editable && (
                                    <Button
                                        shape="circle"
                                        variant="text"
                                        onClick={() => setEditor(prev => ({ ...prev, visible: true }))}
                                    >
                                        <Edit1Icon />
                                    </Button>
                                )}
                            </>
                        ) : (
                            <>
                                <Text>暂无标签</Text>
                                {editor.editable && (
                                    <Button
                                        shape="circle"
                                        variant="text"
                                        onClick={() => setEditor(prev => ({ ...prev, visible: true }))}
                                    >
                                        <Edit1Icon />
                                    </Button>
                                )}
                            </>
                        )}
                    </Space>
                </FormItem>
                {renderRuleLabelsDialog()}
            </Row>
        </div>
    );

    const renderEditForm = (
        <>
            <div style={{ marginBottom: 24, padding: 16, border: '1px solid #e0e0e0', borderRadius: 6 }}>
                <Divider>无损上线</Divider>
                <FormItem label="延迟注册">
                    {editor.editable ? (
                        <div>
                            <Switch
                                defaultValue={losslessRule.lossless_online?.delay_register?.enable}
                                value={losslessRule.lossless_online?.delay_register?.enable}
                                disabled={!editor.editable}
                                onChange={(checked) => {
                                    setLosslessRule(prev => ({
                                        ...prev,
                                        lossless_online: {
                                            ...prev.lossless_online,
                                            delay_register: {
                                                ...prev.lossless_online.delay_register,
                                                enable: checked as boolean,
                                            }
                                        }
                                    }))
                                }} />
                        </div>
                    ) : renderReadonlySwitch(losslessRule.lossless_online?.delay_register?.enable)}
                </FormItem>
                {(losslessRule.lossless_online?.delay_register?.enable) && (
                    <>
                        <FormItem label="延迟注册策略">
                            {editor.editable ? (
                                <div>
                                    <RadioGroup
                                        theme='button'
                                        variant='primary-filled'
                                        options={[{
                                            label: '时长延迟',
                                            value: 'DELAY_BY_TIME'
                                        }, {
                                            label: '探测延迟',
                                            value: 'DELAY_BY_HEALTH_CHECK'
                                        }]}
                                        value={losslessRule.lossless_online?.delay_register?.strategy}
                                        defaultValue={losslessRule.lossless_online?.delay_register?.strategy}
                                        readonly={!editor.editable}
                                        onChange={(value) => {
                                            setLosslessRule(prev => ({
                                                ...prev,
                                                lossless_online: {
                                                    ...prev.lossless_online,
                                                    delay_register: {
                                                        ...prev.lossless_online.delay_register,
                                                        strategy: value as string,
                                                    }
                                                }
                                            }))
                                        }}
                                    />
                                </div>
                            ) : renderReadonlyValue(losslessRule.lossless_online?.delay_register?.strategy === 'DELAY_BY_HEALTH_CHECK' ? '探测延迟' : '时长延迟')}
                        </FormItem>
                        {losslessRule.lossless_online?.delay_register?.strategy === 'DELAY_BY_TIME' && (
                            <FormItem label="延迟注册时间(秒)" name={["lossless_online", "delay_register", "interval"]}>
                                {editor.editable ? (
                                    <div>
                                        <InputNumber
                                            min={1}
                                            max={86400}
                                            theme='normal'
                                            suffix="Second"
                                            value={losslessRule.lossless_online?.delay_register?.interval}
                                            readonly={!editor.editable}
                                            inputProps={
                                                {
                                                    borderless: !editor.editable,
                                                }
                                            }
                                            onChange={(value) => {
                                                setLosslessRule(prev => ({
                                                    ...prev,
                                                    lossless_online: {
                                                        ...prev.lossless_online,
                                                        delay_register: {
                                                            ...prev.lossless_online.delay_register,
                                                            interval: value as number,
                                                        }
                                                    }
                                                }))
                                            }}
                                        />
                                    </div>
                                ) : renderSeconds(losslessRule.lossless_online?.delay_register?.interval)}
                            </FormItem>
                        )}
                        {losslessRule.lossless_online?.delay_register?.strategy === 'DELAY_BY_HEALTH_CHECK' && (
                            <>
                                <FormItem label="健康检查协议">
                                    {editor.editable ? (
                                        <div>
                                            <Select
                                                options={[
                                                    {
                                                        label: 'HTTP',
                                                        value: 'http'
                                                    }]}
                                                value={losslessRule.lossless_online?.delay_register?.health_check_protocol}
                                                readonly={!editor.editable}
                                                inputProps={
                                                    {
                                                        borderless: !editor.editable,
                                                    }
                                                }
                                                onChange={(value) => {
                                                    setLosslessRule(prev => ({
                                                        ...prev,
                                                        lossless_online: {
                                                            ...prev.lossless_online,
                                                            delay_register: {
                                                                ...prev.lossless_online.delay_register,
                                                                health_check_protocol: value as string,
                                                            }
                                                        }
                                                    }))
                                                }}
                                            />
                                        </div>
                                    ) : renderReadonlyValue((losslessRule.lossless_online?.delay_register?.health_check_protocol || '-').toUpperCase())}
                                </FormItem>
                                <FormItem label="健康检查方法">
                                    {editor.editable ? (
                                        <div>
                                            <Select
                                                options={HTTPMethodOption}
                                                value={losslessRule.lossless_online?.delay_register?.health_check_method}
                                                readonly={!editor.editable}
                                                inputProps={
                                                    {
                                                        borderless: !editor.editable,
                                                    }
                                                }
                                                onChange={(value) => {
                                                    setLosslessRule(prev => ({
                                                        ...prev,
                                                        lossless_online: {
                                                            ...prev.lossless_online,
                                                            delay_register: {
                                                                ...prev.lossless_online.delay_register,
                                                                health_check_method: value as string,
                                                            }
                                                        }
                                                    }))
                                                }}
                                            />
                                        </div>
                                    ) : renderReadonlyValue(losslessRule.lossless_online?.delay_register?.health_check_method || '-')}
                                </FormItem>
                                <FormItem label="健康检查路径">
                                    {editor.editable ? (
                                        <div>
                                            <Input
                                                value={losslessRule.lossless_online?.delay_register?.health_check_path}
                                                readonly={!editor.editable}
                                                borderless={!editor.editable}
                                                onChange={(value) => {
                                                    setLosslessRule(prev => ({
                                                        ...prev,
                                                        lossless_online: {
                                                            ...prev.lossless_online,
                                                            delay_register: {
                                                                ...prev.lossless_online.delay_register,
                                                                health_check_path: value as string,
                                                            }
                                                        }
                                                    }))
                                                }}
                                            />
                                        </div>
                                    ) : renderReadonlyValue(losslessRule.lossless_online?.delay_register?.health_check_path || '-')}
                                </FormItem>
                                <FormItem label="健康检查间隔(秒)">
                                    {editor.editable ? (
                                        <div>
                                            <InputNumber
                                                min={1}
                                                max={60}
                                                value={losslessRule.lossless_online?.delay_register?.health_check_interval}
                                                readonly={!editor.editable}
                                                inputProps={
                                                    {
                                                        borderless: !editor.editable,
                                                    }
                                                }
                                                theme='normal'
                                                suffix="Second"
                                                onChange={(value) => {
                                                    setLosslessRule(prev => ({
                                                        ...prev,
                                                        lossless_online: {
                                                            ...prev.lossless_online,
                                                            delay_register: {
                                                                ...prev.lossless_online.delay_register,
                                                                health_check_interval: value as number,
                                                            }
                                                        }
                                                    }))
                                                }} />
                                        </div>
                                    ) : renderSeconds(losslessRule.lossless_online?.delay_register?.health_check_interval)}
                                </FormItem>
                            </>
                        )}
                    </>
                )}
                <Divider>服务预热</Divider>
                <FormItem label="预热启用" name={["lossless_online", "warmup", "enable"]}>
                    {editor.editable ? (
                        <div>
                            <Switch
                                value={losslessRule.lossless_online?.warmup?.enable}
                                defaultValue={losslessRule.lossless_online?.warmup?.enable}
                                disabled={!editor.editable}
                                onChange={(checked) => {
                                    setLosslessRule(prev => ({
                                        ...prev,
                                        lossless_online: {
                                            ...prev.lossless_online,
                                            warmup: {
                                                ...prev.lossless_online.warmup,
                                                enable: checked as boolean,
                                            }
                                        }
                                    }))
                                }}
                            />
                        </div>
                    ) : renderReadonlySwitch(losslessRule.lossless_online?.warmup?.enable)}
                </FormItem>
                {(losslessRule.lossless_online?.warmup?.enable) && (
                    <>
                        <FormItem label="预热时长(秒)">
                            {editor.editable ? (
                                <div>
                                    <InputNumber
                                        min={1}
                                        max={86400}
                                        theme='normal'
                                        suffix="Second"
                                        value={losslessRule.lossless_online?.warmup?.interval}
                                        readonly={!editor.editable}
                                        inputProps={
                                            {
                                                borderless: !editor.editable,
                                            }
                                        }
                                        onChange={(value) => {
                                            setLosslessRule(prev => ({
                                                ...prev,
                                                lossless_online: {
                                                    ...prev.lossless_online,
                                                    warmup: {
                                                        ...prev.lossless_online.warmup,
                                                        interval: value as number,
                                                    }
                                                }
                                            }))
                                        }}
                                    />
                                </div>
                            ) : renderSeconds(losslessRule.lossless_online?.warmup?.interval)}
                        </FormItem>
                        <FormItem label="预热终止保护">
                            {editor.editable ? (
                                <div>
                                    <Switch
                                        value={losslessRule.lossless_online?.warmup?.enable_overload_protection}
                                        defaultValue={losslessRule.lossless_online?.warmup?.enable_overload_protection}
                                        disabled={!editor.editable}
                                        onChange={(checked) => {
                                            setLosslessRule(prev => ({
                                                ...prev,
                                                lossless_online: {
                                                    ...prev.lossless_online,
                                                    warmup: {
                                                        ...prev.lossless_online.warmup,
                                                        enable_overload_protection: checked as boolean,
                                                    }
                                                }
                                            }))
                                        }}
                                    />
                                </div>
                            ) : renderReadonlySwitch(losslessRule.lossless_online?.warmup?.enable_overload_protection)}
                        </FormItem>
                        {losslessRule.lossless_online?.warmup?.enable_overload_protection && (
                            <FormItem label="预热终止百分比">
                                {editor.editable ? (
                                    <div>
                                        <InputNumber
                                            min={0}
                                            max={100}
                                            theme='normal'
                                            value={losslessRule.lossless_online?.warmup?.overload_protection_threshold}
                                            readonly={!editor.editable}
                                            inputProps={
                                                {
                                                    borderless: !editor.editable,
                                                }
                                            }
                                            suffix="%"
                                            onChange={(value) => {
                                                setLosslessRule(prev => ({
                                                    ...prev,
                                                    lossless_online: {
                                                        ...prev.lossless_online,
                                                        warmup: {
                                                            ...prev.lossless_online.warmup,
                                                            overload_protection_threshold: value as number,
                                                        }
                                                    }
                                                }))
                                            }}
                                        />
                                    </div>
                                ) : renderReadonlyValue(`${losslessRule.lossless_online?.warmup?.overload_protection_threshold ?? '-'}%`)}
                            </FormItem>
                        )}
                        <FormItem label="预热曲线值" name={["lossless_online", "warmup", "curvature"]}>
                            {editor.editable ? (
                                <InputNumber
                                    min={1}
                                    max={5}
                                    theme='normal'
                                    onChange={(value) => {
                                        setLosslessRule(prev => ({
                                            ...prev,
                                            lossless_online: {
                                                ...prev.lossless_online,
                                                warmup: {
                                                    ...prev.lossless_online.warmup,
                                                    curvature: value as number,
                                                }
                                            }
                                        }))
                                    }}
                                />
                            ) : (
                                <Text>{losslessRule.lossless_online?.warmup?.curvature || '-'}</Text>
                            )}
                        </FormItem>
                    </>
                )}
                <Divider>无损下线</Divider>
                <FormItem label="无损下线启用">
                    {editor.editable ? (
                        <div>
                            <Switch
                                value={losslessRule.lossless_offline?.enable}
                                defaultValue={losslessRule.lossless_offline?.enable}
                                disabled={!editor.editable}
                                onChange={(checked) => {
                                    setLosslessRule(prev => ({
                                        ...prev,
                                        lossless_offline: {
                                            ...prev.lossless_offline,
                                            enable: checked as boolean,
                                        }
                                    }))
                                }} />
                        </div>
                    ) : renderReadonlySwitch(losslessRule.lossless_offline?.enable)}
                </FormItem>
                {(losslessRule.lossless_offline?.enable) && (
                    <FormItem label="无损下线间隔">
                        {editor.editable ? (
                            <div>
                                <InputNumber
                                    min={0}
                                    max={120}
                                    theme='normal'
                                    suffix="Second"
                                    value={losslessRule.lossless_offline?.interval}
                                    readonly={!editor.editable}
                                    inputProps={
                                        {
                                            borderless: !editor.editable,
                                        }
                                    }
                                    onChange={(value) => {
                                        setLosslessRule(prev => ({
                                            ...prev,
                                            lossless_offline: {
                                                ...prev.lossless_offline,
                                                interval: value as number,
                                            }
                                        }))
                                    }}
                                />
                            </div>
                        ) : renderSeconds(losslessRule.lossless_offline?.interval)}
                    </FormItem>
                )}
            </div>
        </>
    );

    const renderStickyTool = (
        <StickyTool
            style={{ zIndex: 1000 }}
            placement="right-bottom"
            offset={[-10, 200]}
        >
            <StickyItem
                label=""
                icon={!editor.editable ?
                    <RuleStickyAction label="编辑" icon={<Edit1Icon />} onClick={() => {
                        setEditor(prev => ({ ...prev, editable: true }));
                    }} />
                    :
                    <RuleStickyAction label="保存" icon={<SaveIcon />} onClick={() => {
                        form.submit();
                    }} />
                }
            />
            {editor.editable && (
                <StickyItem label="" icon={
                    <RuleStickyAction label="撤销" icon={<RollbackIcon />} onClick={() => {
                        setEditor(prev => ({ ...prev, editable: false }));
                    }} />
                } />
            )}
            {!editor.editable && (
                <StickyItem label="" icon={
                    <RuleStickyAction label="发布" icon={<RocketIcon />} onClick={() => {
                        setEditor(prev => ({ ...prev, publishView: true }));
                    }} />
                } />
            )}
        </StickyTool>
    );

    const renderPublishForm = (
        <>
            {editor.publishView && (
                <PublishForm
                    ruleId={losslessRule.id || ''}
                    ruleName={losslessRule.id || ''}
                    resource={PolicySourceType.LossLessRules}
                    visible={editor.publishView}
                    close={() => {
                        setEditor(prev => ({ ...prev, publishView: false }));
                    }}
                />
            )}
        </>
    )

    return (
        <div style={{ padding: 24 }}>
            <Form
                form={form}
                onSubmit={onSubmit}
                layout="vertical"
                labelAlign="left"
                labelWidth={120}
                colon
            >
                {serviceInfo}
                {renderEditForm}
                {renderPublishForm}
                {renderStickyTool}
            </Form>
        </div>
    );
};

export default React.memo(LossLessEditor);
