import React, { } from 'react';
import {
    Form,
    Input,
    Select,
    InputNumber,
    Radio,
    Button,
    Space,
    Textarea,
    StickyTool,
    FormProps,
    InputAdornment,
    Row,
    Col,
    Tag,
    Dialog,
    Popup
} from 'tdesign-react';
import { AddIcon, CloseIcon, Edit1Icon, RocketIcon, RollbackIcon, SaveIcon } from 'tdesign-icons-react';

import { API, HTTPMethod, HTTPMethodOption, InterfaceProtocol, InterfaceProtocolOption, Label, MatchType, MatchTypeMap, MatchTypeOption, Op } from "services/types";
import {
    FaultDetectRule,
    FaultDetectProtocol,
    FaultDetectHttpMethodOptions
} from 'services/faultdetect';
import LabelInput from 'components/LabelInput';
import PublishForm from '../RuleRelease/PublishForm';
import RuleStickyAction from '../RuleRelease/RuleStickyAction';
import { PolicySourceType } from 'services/auth_policy';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { listOneFaultDetect, resetFaultDetect, saveFaultDetects, selectFaultDetect, updateFaultDetects } from 'modules/governance/faultdetect';
import { cleanServicePage, listAllServices, selectService } from 'modules/discovery/service';
import { cleanNamespacePage, listAllNamespaces, selectNamespace } from 'modules/namespace';

const { FormItem } = Form;
const { Group: RadioGroup } = Radio;
const { StickyItem } = StickyTool;


interface FaultDetectDO {
    id?: string
    name: string
    description: string
    targetService: {
        namespace: string
        service: string
        api?: API
    }
    interval: number
    timeout: number
    port: number
    // 协议，支持HTTP, TCP, UPD
    protocol: string
    httpConfig?: {
        method: string
        url: string
        headers: Label[]
        body: string
    }
    tcpConfig?: {
        send: string
        receive: string[]
    }
    udpConfig?: {
        send: string
        receive: string[]
    }
    metadata?: Label[]
}

export const defaultFaultDetectRule: () => FaultDetectDO = () => ({
    name: '',
    description: '',
    targetService: {
        namespace: '',
        service: '',
        method: {
            type: MatchType.EXACT,
            value: ''
        }
    },
    interval: 30,
    timeout: 60,
    port: 0,
    protocol: InterfaceProtocol.HTTP,
    httpConfig: {
        method: HTTPMethod.GET,
        url: '',
        headers: [],
        body: ''
    }
});

interface IFaultDetectEditorProps {
    op: Op;
    refresh: (close: boolean) => void;
}

const FaultDetectEditor: React.FC<IFaultDetectEditorProps> = (props) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const detectState = useAppSelector(selectFaultDetect);
    const { editRule, viewRule } = detectState;

    const [editorState, setEditorState] = React.useState<{
        model: Op;
        publishView: boolean;
        loading: boolean;
        editable: boolean;
        content: string;
        visible: boolean;
    }>({
        model: 'view',
        content: '',
        editable: props.op === 'create',
        publishView: false,
        loading: false,
        visible: false
    });

    // 选项数据
    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { datas: serviceDatas } = serviceState;

    React.useEffect(() => {
        // 创建模式，设置默认值
        form.setFieldsValue(defaultFaultDetectRule());
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
            dispatch(resetFaultDetect());
        }
    }, []);

    React.useEffect(() => {
        if (!editRule) {
            return;
        }
        if (editRule.id !== '') {
            dispatch(listOneFaultDetect({ id: editRule.id || '' })).then(res => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('请求失败', `获取主动探测规则详情失败: ${res?.payload as string}`);
                    return;
                }
                const { editRule } = res.payload as { editRule: FaultDetectRule };
                form.setFieldsValue(editRule);
            })
        }
    }, [editRule])

    // 表单提交
    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }

        setEditorState(prev => ({ ...prev, loading: true }));
        try {
            const ruleData: FaultDetectRule = {
                ...e.fields,
                id: viewRule?.id || '',
                editable: true,
                deleteable: true
            };

            let res;
            if (props.op === 'create') {
                res = await dispatch(saveFaultDetects({ param: { ...ruleData } }));
            } else {
                res = await dispatch(updateFaultDetects({ param: { ...ruleData } }));
            }
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求错误', res?.payload as string);
            } else {
                openInfoNotification('请求成功', props.op !== 'create' ? '修改主动探测规则成功' : '创建主动探测规则成功');
                if (props.op === 'create') {
                    props.refresh(false); // 刷新列表
                } else {
                    setEditorState(prev => ({ ...prev, loading: false, editable: false }));
                }
            }
        } catch (error) {
            console.error('提交失败:', error);
            openErrNotification('操作失败', '提交故障检测规则失败，请检查输入或稍后重试');
        }
    };

    const metadata = Form.useWatch('metadata', form) || (editRule?.metadata as Record<string, string>) || {};
    const [labelDialogList, setLabelDialogList] = React.useState<{ key: string; value: string }[]>([]);

    React.useEffect(() => {
        if (editorState.visible) {
            const current = form.getFieldValue('metadata') || {};
            setLabelDialogList(Object.entries(current).map(([key, value]) => ({ key, value })));
        }
    }, [editorState.visible]);

    const renderRuleLabelsDialog = () => (
        <Dialog
            visible={editorState.visible}
            header="编辑规则标签"
            width={700}
            onConfirm={() => {
                const next: Record<string, string> = {};
                labelDialogList.forEach(({ key, value }) => { if (key) next[key] = value; });
                form.setFieldsValue({ metadata: next });
                setEditorState(prev => ({ ...prev, visible: false }));
            }}
            onClose={() => setEditorState(prev => ({ ...prev, visible: false }))}
        >
            <div>
                {labelDialogList.map((tag, idx) => (
                    <Row gutter={8} key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Col span={4}>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const next = [...labelDialogList];
                                    next[idx] = { ...next[idx], key: v };
                                    setLabelDialogList(next);
                                }}
                            />
                        </Col>
                        <Col span={4}>
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const next = [...labelDialogList];
                                    next[idx] = { ...next[idx], value: v };
                                    setLabelDialogList(next);
                                }}
                            />
                        </Col>
                        <Col span={2}>
                            <Popup trigger="hover" content="删除标签">
                                <Button shape="circle" variant="text" onClick={() => setLabelDialogList(labelDialogList.filter((_, i) => i !== idx))}>
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        </Col>
                    </Row>
                ))}
                <Button variant="text" icon={<AddIcon />} onClick={() => setLabelDialogList([...labelDialogList, { key: '', value: '' }])}>添加标签</Button>
            </div>
        </Dialog>
    );

    const baseInfoCardStyle = { marginBottom: 24, padding: '24px 32px', border: '1px solid #e5e6eb', borderRadius: 8, boxShadow: '0 2px 8px 0 rgba(0,0,0,0.03)' };

    const ruleBaseInfo = (
        <div style={baseInfoCardStyle}>
            {/* 第一层：规则名称 */}
            <Row style={{ marginBottom: 16 }}>
                <Col span={12}>
                    <FormItem
                        label="规则名称"
                        name="name"
                        showErrorMessage={editorState.editable}
                        rules={[
                            { required: true, message: '请输入规则名称' },
                            { max: 64, message: '最长64个字符' }
                        ]}
                    >
                        {props.op === 'create' ? (
                            <Input style={{ width: '300px' }} placeholder="最长64个字符" />
                        ) : (
                            <Text>{editRule?.name}</Text>
                        )}
                    </FormItem>
                </Col>
            </Row>
            {/* 第二层：描述 */}
            <Row style={{ marginBottom: 16 }}>
                <Col span={12}>
                    <FormItem label="描述" name="description">
                        {editorState.editable ? (
                            <Input style={{ width: '600px' }} placeholder="" />
                        ) : (
                            <Text>{editRule?.description || '-'}</Text>
                        )}
                    </FormItem>
                </Col>
            </Row>
            {/* 第三层：规则标签 */}
            <Row>
                <Col span={12}>
                    <FormItem label="规则标签" name="metadata">
                        <Space align="center">
                            {Object.keys(metadata).length > 0 ? (
                                <>
                                    {Object.entries(metadata).map(([k, v], idx) => (
                                        <Tag key={idx}>{`${k}: ${v}`}</Tag>
                                    ))}
                                    {editorState.editable && (
                                        <Button shape="circle" variant="text" onClick={() => setEditorState(prev => ({ ...prev, visible: true }))}>
                                            <Edit1Icon />
                                        </Button>
                                    )}
                                </>
                            ) : (
                                <>
                                    <Text>暂无标签</Text>
                                    {editorState.editable && (
                                        <Button shape="circle" variant="text" onClick={() => setEditorState(prev => ({ ...prev, visible: true }))}>
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

    const renderEditRule = (
        <Form
            form={form}
            onSubmit={onSubmit}
            labelWidth="120px"
            layout="vertical"
        >
            {ruleBaseInfo}

            {/* 命名空间 */}
            <FormItem
                label="命名空间"
                name={['targetService', 'namespace']}
                rules={[{ required: true, message: '请选择命名空间' }]}
            >
                {editorState.editable ? (
                    <Select
                        style={{ width: '300px' }}
                        filterable={true}
                        creatable={true}
                        options={namespaceDatas.map(ns => ({
                            label: ns.name,
                            value: ns.name,
                            namespace: ns.name
                        }))}
                    />
                ) : (
                    <Text>{editRule?.targetService?.namespace === '*' ? '全部命名空间' : editRule?.targetService?.namespace}</Text>
                )}
            </FormItem>

            {/* 服务名称 */}
            <FormItem shouldUpdate={(prev, next) => prev.targetService?.namespace !== next.targetService?.namespace}>
                {({ getFieldValue, setFieldsValue }) => {
                    return (
                        <FormItem
                            label="服务名称"
                            name={['targetService', 'service']}
                            rules={[{ required: true, message: '请选择服务' }]}
                        >
                            {editorState.editable ? (
                                <Select
                                    style={{ width: '300px' }}
                                    placeholder="请选择服务"
                                    filterable={true}
                                    creatable={true}
                                    options={serviceDatas.filter(opt => {
                                        const selectNs = getFieldValue(['targetService', 'namespace']);
                                        if (selectNs === '*') {
                                            return true; // 允许所有命名空间的服务
                                        }
                                        return opt.namespace === selectNs; // 仅允许当前命名空间
                                    }).map(service => ({
                                        label: service.name,
                                        value: service.name,
                                        namespace: service.namespace
                                    }))}
                                />
                            ) : (
                                <Text>{editRule?.targetService?.service === '*' ? '全部服务' : editRule?.targetService?.service}</Text>
                            )}
                        </FormItem>
                    )
                }}
            </FormItem>

            {/* 接口名称 */}
            <FormItem label="接口名称">
                {editorState.editable ? (
                    <Space direction="horizontal" style={{ width: '100%', display: 'flex' }}>
                        <FormItem name={['targetService', 'api', 'path', 'value']} style={{ flex: 1, marginBottom: 0 }}>
                            {editorState.editable ? (
                                <Input placeholder="请输入接口名称" />
                            ) : (
                                <Text>{editRule?.targetService?.api?.path?.value || '-'}</Text>
                            )}
                        </FormItem>
                        <FormItem name={['targetService', 'api', 'path', 'type']} style={{ width: 150, marginBottom: 0 }}>
                            {editorState.editable ? (
                                <Select
                                    options={MatchTypeOption}
                                />
                            ) : (
                                <Text>{MatchTypeMap[editRule?.targetService?.api?.path?.type as MatchType || MatchType.EXACT]}</Text>
                            )}
                        </FormItem>
                    </Space>

                ) : (
                    <>
                        <InputAdornment append={
                            `${MatchTypeMap[editRule?.targetService?.api?.path?.type as MatchType || MatchType.EXACT]}`
                        }>
                            <Input readonly={true}
                                value={editRule?.targetService?.api?.path?.value || '-'}
                            />
                        </InputAdornment>
                    </>
                )}
            </FormItem>

            {/* 接口协议 */}
            <FormItem label="接口协议" name={['targetService', 'api', 'protocol']} initialData={FaultDetectProtocol.HTTP}>
                <RadioGroup
                    theme='button'
                    variant='primary-filled'
                    readonly={!editorState.editable}
                >
                    {InterfaceProtocolOption.map(option => (
                        <Radio.Button key={option.value} value={option.value}>
                            {option.label}
                        </Radio.Button>
                    ))}
                </RadioGroup>
            </FormItem>

            {/* 接口方法 */}
            <FormItem label="接口方法" name={['targetService', 'api', 'method']}>
                {editorState.editable ? (
                    <Select
                        style={{ width: '300px' }}
                        filterable={true}
                        creatable={true}
                        options={FaultDetectHttpMethodOptions}
                    />
                ) : (
                    <Text>{editRule?.targetService?.api?.method || '-'}</Text>
                )}
            </FormItem>

            {/* 间隔 */}
            <FormItem label="间隔" name="interval">
                {editorState.editable ? (
                    <Space>
                        <InputNumber min={1} placeholder="30" />
                        <span>秒</span>
                    </Space>
                ) : (
                    <Text>{editRule?.interval || 30} 秒</Text>
                )}
            </FormItem>

            {/* 超时时间 */}
            <FormItem label="超时时间" name="timeout">
                {editorState.editable ? (
                    <Space>
                        <InputNumber min={1} placeholder="60" />
                        <span>秒</span>
                    </Space>
                ) : (
                    <Text>{editRule?.timeout || 60} 秒</Text>
                )}
            </FormItem>

            {/* 端口 */}
            <FormItem label="端口" name="port">
                {editorState.editable ? (
                    <InputNumber min={1} placeholder="端口号" />
                ) : (
                    <Text>{editRule?.port || '-'}</Text>
                )}
            </FormItem>

            {/* 协议选择 */}
            <FormItem label="协议" initialData={FaultDetectProtocol.HTTP} name="protocol" help={'服务实例下需要存在所选择用于探测的协议，否则无法探测无法生效'}>
                <RadioGroup
                    theme='button'
                    variant='primary-filled'
                    readonly={!editorState.editable}
                >
                    <Radio.Button value={FaultDetectProtocol.HTTP}>HTTP</Radio.Button>
                    <Radio.Button value={FaultDetectProtocol.TCP}>TCP</Radio.Button>
                    <Radio.Button value={FaultDetectProtocol.UDP}>UDP</Radio.Button>
                </RadioGroup>
            </FormItem>

            <FormItem shouldUpdate={(prev, next) => prev.protocol !== next.protocol}>
                {({ getFieldValue, setFieldsValue }) => {
                    const protocol = getFieldValue('protocol');
                    switch (protocol) {
                        case FaultDetectProtocol.TCP:
                            return (
                                <>
                                    {editorState.editable ? (
                                        <>
                                            <FormItem label="发送内容" name={['tcpConfig', 'send']} help={'配置所需发送的二进制报文'}>
                                                <Textarea placeholder="请输入发送内容" />
                                            </FormItem>
                                            <FormItem label="接收内容" name={['tcpConfig', 'receive']}>
                                                <Textarea placeholder="请输入期望接收的内容" />
                                            </FormItem>
                                        </>
                                    ) : (
                                        <>
                                            <FormItem label="发送内容" name={['tcpConfig', 'send']} help={'配置所需发送的二进制报文'}>
                                                <Text>{editRule?.tcpConfig?.send || '-'}</Text>
                                            </FormItem>
                                            <FormItem label="接收内容" name={['tcpConfig', 'receive']}>
                                                <Text>{editRule?.tcpConfig?.receive || '-'}</Text>
                                            </FormItem>
                                        </>
                                    )}
                                </>
                            )
                        case FaultDetectProtocol.UDP:
                            return (
                                <>
                                    <FormItem label="发送内容" name={['udpConfig', 'send']} help={'配置所需发送的二进制报文'}>
                                        <Textarea placeholder="请输入发送内容" />
                                    </FormItem>
                                    <FormItem label="接收内容" name={['udpConfig', 'receive']}>
                                        <Textarea placeholder="请输入期望接收的内容" />
                                    </FormItem>
                                </>
                            )
                        case FaultDetectProtocol.HTTP:
                            return (
                                <>
                                    {editorState.editable ? (
                                        <>
                                            <FormItem label="方法" name={['httpConfig', 'method']}>
                                                <Select
                                                    style={{ width: '300px' }}
                                                    filterable={true}
                                                    creatable={true}
                                                    options={HTTPMethodOption}
                                                />
                                            </FormItem>
                                            <FormItem label="Url" name={['httpConfig', 'url']}>
                                                <Input placeholder="请输入Url以/开头" />
                                            </FormItem>
                                            <LabelInput
                                                form={form}
                                                name={['httpConfig', 'headers']}
                                                label="Headers"
                                                editable={editorState.editable}
                                            />
                                        </>
                                    ) : (
                                        <>
                                            <FormItem label="方法" name={['httpConfig', 'method']}>
                                                <Text>{editRule?.httpConfig?.method || '-'}</Text>
                                            </FormItem>
                                            <FormItem label="Url" name={['httpConfig', 'url']}>
                                                <Text>{editRule?.httpConfig?.url || '-'}</Text>
                                            </FormItem>
                                            <FormItem name={['httpConfig', 'headers']}>
                                                <LabelInput
                                                    form={form}
                                                    name={['httpConfig', 'headers']}
                                                    label="Headers"
                                                    editable={editorState.editable}
                                                />
                                            </FormItem>
                                        </>
                                    )}
                                </>
                            )
                        default:
                            return (<></>)
                    }
                }}
            </FormItem>
        </Form>
    )

    const renderStickyTool = (
        <>
            {viewRule?.editable && (
                <div style={{ marginTop: 20 }}>
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
                                    if (props.op === 'create') {
                                        props.refresh(true);
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
                </div>
            )}
        </>
    )

    return (
        <div style={{ padding: 24 }}>
            {renderEditRule}
            {editorState.publishView && (
                <PublishForm
                    ruleId={viewRule?.id || ''}
                    ruleName={viewRule?.name || ''}
                    resource={PolicySourceType.FaultDetectRules}
                    visible={editorState.publishView}
                    close={() => {
                        setEditorState(prev => ({ ...prev, publishView: false }));
                    }}
                />
            )}
            {renderStickyTool}
        </div>
    );
};

export default React.memo(FaultDetectEditor);
