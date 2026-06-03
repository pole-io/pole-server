import React, { useState } from 'react';
import {
    Form,
    Input,
    Select,
    Button,
    Space,
    Checkbox,
    FormProps,
    StickyTool,
    Steps,
    Transfer,
    Row,
    Tag,
    Dialog,
    Col,
    Popup,
    Collapse,
    TransferValue
} from 'tdesign-react';
import { AddIcon, CloseIcon, Edit1Icon, RocketIcon, RollbackIcon, SaveIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { Label, Op } from 'services/types';
import {
    LaneGatewaySelectorType,
    LaneGroup,
    LaneGroupView,
    LaneServiceSelectorType,
    ServiceGatewaySelector,
    ServiceSelector,
    TrafficEntry,
} from 'services/lane';
import { ServiceView } from 'services/service';
import styles from './LaneGroupEditor.module.less';
import PublishForm from '../RuleRelease/PublishForm';
import { PolicySourceType } from 'services/auth_policy';
import { listOneLaneGroup, resetLaneGroup, saveLaneGroups, selectLaneGroup, updateLaneGroups } from 'modules/governance/lane_group';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { RoutingRuleDestination } from 'services/router';
import Text from 'components/Text';
import { cleanServicePage, listAllServices, selectService } from 'modules/discovery/service';
import cloneDeep from 'lodash/cloneDeep';
import CollapsePanel from 'tdesign-react/es/collapse/CollapsePanel';
import { set } from 'lodash';

const { StepItem } = Steps;
const { FormItem } = Form;
const { StickyItem } = StickyTool;

interface LaneGroupRule {
    id: string;
    name: string;
    description: string;
    gateway_entries: string[];
    service_entries: string[];
    destinations: string[];
    metadata: Label[];
}

export const defaultLaneGroupRule: () => LaneGroupRule = () => ({
    id: '',
    name: '',
    description: '',
    gateway_entries: [],
    service_entries: [],
    destinations: [],
    metadata: [],
})

interface SimpleService {
    label: string,
    value: string,
    service: string,
    namespace: string,
}

interface ILaneGroupEditorProps {
    op: Op;
    refresh: (close: boolean) => void;
}

const LaneGroupEditor: React.FC<ILaneGroupEditorProps> = ({ op, refresh }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const { datas: serviceDatas } = useAppSelector(selectService);
    const { editGroup } = useAppSelector(selectLaneGroup)

    const [curRule, setCurRule] = React.useState<LaneGroupRule>(defaultLaneGroupRule());

    const [editorState, setEditorState] = React.useState<{
        model: Op;
        publishView: boolean;
        editable: boolean;
        content: string;
        showMSGateway?: boolean;
        showMSService?: boolean;
        visible: boolean;
    }>({
        model: 'view',
        content: '',
        editable: op === 'create',
        publishView: false,
        visible: false,
    });

    // 记录 svc-id => {namespace, service} 的映射
    // 用于在提交时获取服务的 namespace 和 service 名称
    const [svcMap, setSvcMap] = React.useState<Map<string, SimpleService>>(new Map());

    // 选项数据
    const [serviceOptions, setServiceOptions] = useState<SimpleService[]>([]);
    const [gatewayServices, setGatewayServices] = useState<SimpleService[]>([]);

    // 获取服务列表
    const loadService = (services: ServiceView[]) => {
        // Transform services into SimpleService objects once
        const simpleServices = services.map((service: ServiceView): SimpleService => ({
            label: `${service.name} (${service.namespace})`,
            value: `${service.namespace}/${service.name}`,
            service: service.name,
            namespace: service.namespace
        }));

        // Filter services using the transformed array
        setServiceOptions(simpleServices.filter((s) => {
            const metadata = services.find(sv => sv.id === s.value)?.metadata;
            return !metadata?.['service_gateway'] || metadata['service_gateway'] !== 'true';
        }));

        setGatewayServices(simpleServices.filter((s) => {
            const metadata = services.find(sv => sv.id === s.value)?.metadata;
            return metadata?.['service_gateway'] === 'true';
        }));

        // Build maps in a single pass
        const serviceMap = new Map<string, SimpleService>();

        simpleServices.forEach(service => {
            serviceMap.set(service.value, service);
        });

        setSvcMap(serviceMap);
    };

    React.useEffect(() => {
        // Load services on initial mount
        dispatch(listAllServices()).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求失败', `获取服务列表失败, ${res?.payload as string}`);
            } else {
                loadService(serviceDatas);
            }
        });

        // Cleanup function
        return () => {
            dispatch(cleanServicePage());
            setSvcMap(new Map());
        };
    }, []);

    React.useEffect(() => {
        if (editGroup) {
            // Handle lane group details if available
            if (editGroup?.id) {
                // Fetch lane group details if needed
                dispatch(listOneLaneGroup({ id: editGroup.id })).then((res) => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('请求失败', `获取泳道组详情失败, ${res?.payload as string}`);
                    } else {
                        const ret = res.payload as { viewGroup: LaneGroupView } | null;
                        // Process lane group details after successful fetch
                        fetchLaneGroupDetail(ret?.viewGroup || null);
                    }
                });
            }
            // If in create mode or editGroup is already available, process directly
            else if (editGroup) {
                fetchLaneGroupDetail(editGroup);
            }
        }
    }, [editGroup]);

    // 获取泳道组详情
    const fetchLaneGroupDetail = (laneGroup: LaneGroupView | null) => {
        if (!laneGroup) {
            return;
        }
        const cloneRule = cloneDeep(laneGroup);
        const curData: LaneGroupRule = {
            id: cloneRule.id || '',
            name: cloneRule.name || '',
            description: cloneRule.description || '',
            gateway_entries: cloneRule.entries?.filter((v) => {
                return v.type === 'gateway';
            }).map((v) => {
                return v.selector.namespace + '/' + v.selector.service
            }) as string[] || [],
            service_entries: cloneRule.entries?.filter((v) => {
                return v.type === 'service';
            }).map((v) => {
                return v.selector.namespace + '/' + v.selector.service
            }) as string[] || [],
            destinations: cloneRule.destinations.map((v) => {
                return v.namespace + '/' + v.service
            }) as string[] || [],
            metadata: Object.entries(cloneRule.metadata || {}).reduce((acc, [key, value]) => {
                acc.push({ key, value });
                return acc;
            }, [] as { key: string, value: string }[])
        }
        setEditorState((prev) => ({ ...prev, showMSGateway: curData.gateway_entries.length > 0, showMSService: curData.service_entries.length > 0 }));
        setCurRule(curData)
    };

    // 表单提交
    const onSubmit: FormProps['onSubmit'] = async (e) => {
        const service_entries = curRule?.service_entries as string[] || [];
        const gateway_entries = curRule?.gateway_entries as string[] || [];
        const destinations = curRule?.destinations as string[] || [];
        if (service_entries.length === 0 && gateway_entries.length === 0) {
            openErrNotification('配置错误', '请至少选择一个微服务网关或微服务入口');
            return;
        }

        // 组装完整的泳道组数据
        const laneGroupData: LaneGroup = {
            id: editGroup?.id || '',
            name: e.fields.name || '',
            description: e.fields.description || '',
            entries: [
                ...service_entries.map((id) => {
                    return {
                        '@type': LaneServiceSelectorType,
                        type: 'service',
                        selector: {
                            namespace: svcMap.get(id)?.namespace || '',
                            service: svcMap.get(id)?.service || '',
                        } as ServiceSelector
                    } as TrafficEntry
                }),
                ...gateway_entries.map((id) => {
                    return {
                        '@type': LaneGatewaySelectorType,
                        type: 'gateway',
                        selector: {
                            namespace: svcMap.get(id)?.namespace || '',
                            service: svcMap.get(id)?.service || '',
                        } as ServiceGatewaySelector
                    } as TrafficEntry
                })],
            destinations: destinations.map((id) => {
                return {
                    name: id,
                    namespace: svcMap.get(id)?.namespace || '',
                    service: svcMap.get(id)?.service || '',
                } as RoutingRuleDestination
            }),
            metadata: (e.fields.metadata as Label[] || [])?.reduce((acc, item) => {
                acc[item.key] = item.value;
                return acc;
            }, {} as Record<string, string>)
        };

        let res;
        if (op === 'create') {
            res = await dispatch(saveLaneGroups({ param: laneGroupData }));
        } else {
            res = await dispatch(updateLaneGroups({ param: laneGroupData }));
        }
        if (res.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', `${op === 'view' ? '修改' : '创建'}泳道组失败, ${res?.payload as string}`);
        } else {
            openInfoNotification('请求成功', op === 'view' ? '修改泳道组规则成功' : '创建泳道组规则成功');
            // 重置编辑状态为不可编辑
            setEditorState((prev) => ({ ...prev, editable: false }));
            refresh(op === 'create'); // 刷新列表
        }
    };


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
                {curRule.metadata?.map((tag, idx) => (
                    <Row gutter={8} key={idx} style={{ marginBottom: 12 }} align="middle">
                        <Col span={4}>
                            <Input
                                value={tag.key}
                                placeholder="请输入标签键"
                                onChange={v => {
                                    const newMetadata = [...(curRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], key: v };
                                    setCurRule({ ...curRule, metadata: newMetadata });
                                }} />
                        </Col>
                        <Col span={4}>
                            <Input
                                value={tag.value}
                                placeholder="请输入标签值"
                                onChange={v => {
                                    const newMetadata = [...(curRule.metadata || [])];
                                    newMetadata[idx] = { ...newMetadata[idx], value: v };
                                    setCurRule({ ...curRule, metadata: newMetadata });
                                }} />
                        </Col>
                        <Col span={2}>
                            <Popup trigger="hover" content="删除标签">
                                <Button
                                    shape="circle"
                                    variant="text"
                                    onClick={() => {
                                        const newMetadata = [...(curRule.metadata || [])];
                                        newMetadata.splice(idx, 1);
                                        setCurRule({ ...curRule, metadata: newMetadata });
                                    }}
                                >
                                    <CloseIcon />
                                </Button>
                            </Popup>
                        </Col>
                    </Row>
                ))}
                <Button variant="text" icon={<AddIcon />} onClick={() => {
                    const newMetadata = [...(curRule?.metadata || []), { key: '', value: '' }];
                    setCurRule({ ...curRule, metadata: newMetadata });
                }}>添加标签</Button>
            </div>
        </Dialog>
    );


    const renderEditor = () => {
        return (
            <div className={styles.laneGroupEditor}>
                <Form
                    form={form}
                    onSubmit={onSubmit}
                    layout="vertical"
                >
                    <Steps readonly={true} layout="vertical" theme='default'>
                        <StepItem value={1} title={editorState.editable ? '编辑泳道组信息' : '泳道组信息'} status='process'>
                            {/* 1. 编辑组信息：第一层名称，第二层描述，第三层规则标签（样式与各规则页统一） */}
                            <div style={{ marginTop: 24, padding: '24px 32px', border: '1px solid #e5e6eb', borderRadius: 8, boxShadow: '0 2px 8px 0 rgba(0,0,0,0.03)', marginBottom: 24 }}>
                                <Space direction='vertical' style={{ width: '100%' }}>
                                    {/* 第一层：泳道组名 */}
                                    <Row style={{ marginBottom: 16 }}>
                                        <FormItem
                                            label="泳道组名"
                                            rules={editorState.editable ? [
                                                { required: true, message: '请输入泳道组名' },
                                                { pattern: /^[a-zA-Z0-9._-]{1,63}$/, message: '允许数字、英文字母、.、-、_，限制63个字符' }
                                            ] : []}
                                        >
                                            <div>
                                                {editorState.editable ? (
                                                    <Input
                                                        placeholder="泳道组名（必填）"
                                                        value={curRule.name}
                                                        maxlength={64}
                                                        onChange={(v) => {
                                                            setCurRule((prev) => ({ ...prev, name: v }));
                                                        }} />
                                                ) : (
                                                    <Text>{curRule.name || '未命名泳道组'}</Text>
                                                )}
                                            </div>
                                        </FormItem>
                                    </Row>
                                    {/* 第二层：描述 */}
                                    <Row style={{ marginBottom: 16 }}>
                                        <FormItem label="描述">
                                            <div>
                                                {editorState.editable ? (
                                                    <Input
                                                        placeholder="对泳道组进行描述"
                                                        value={curRule?.description}
                                                        onChange={(v) => {
                                                            setCurRule((prev) => ({ ...prev, description: v }));
                                                        }}
                                                    />
                                                ) : (
                                                    <Text>{curRule.description || '暂无描述'}</Text>
                                                )}
                                            </div>
                                        </FormItem>
                                    </Row>
                                    {/* 第三层：规则标签 */}
                                    <Row>
                                        <FormItem label="规则标签">
                                            <Space align="center">
                                                {curRule.metadata && curRule.metadata.length > 0 ? (
                                                    <>
                                                        {curRule.metadata.map((tag, idx) => (
                                                            <Tag key={idx}>{`${tag.key}: ${tag.value}`}</Tag>
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
                                    </Row>
                                </Space>
                            </div>
                        </StepItem>
                        <StepItem value={2} title={editorState.editable ? '配置泳道组入口' : '泳道组入口'} status='process'>
                            {/* 2. 配置泳道网关入口 */}
                            <div style={{ marginTop: 24 }}>
                                <Space direction="vertical" style={{ width: '100%', marginLeft: 32 }}>
                                    {editorState.editable ? (
                                        <>
                                            <Checkbox
                                                label={'微服务网关SpringCloud Gateway'}
                                                checked={editorState.showMSGateway || false}
                                                value={editorState.showMSGateway || false}
                                                onChange={(checked) => setEditorState(prev => ({ ...prev, showMSGateway: checked }))}
                                            />
                                            {editorState.showMSGateway && (
                                                <FormItem name={'gateway_entries'}>
                                                    <div>
                                                        <Select
                                                            multiple={true}
                                                            options={gatewayServices}
                                                            value={curRule.gateway_entries}
                                                            onChange={(val) => {
                                                                setCurRule((prev) => ({
                                                                    ...prev,
                                                                    gateway_entries: val as string[],
                                                                }));
                                                            }}
                                                        />
                                                    </div>
                                                </FormItem>
                                            )}
                                            <Checkbox
                                                label={'微服务应用'}
                                                checked={editorState.showMSService || false}
                                                value={editorState.showMSService || false}
                                                onChange={(checked) => setEditorState(prev => ({ ...prev, showMSService: checked }))}
                                            />
                                            {editorState.showMSService && (
                                                <FormItem name={'service_entries'}>
                                                    <div>
                                                        <Select
                                                            multiple={true}
                                                            options={serviceOptions}
                                                            value={curRule.service_entries}
                                                            onChange={(val) => {
                                                                setCurRule((prev) => ({
                                                                    ...prev,
                                                                    service_entries: val as string[],
                                                                }));
                                                            }}
                                                        />
                                                    </div>
                                                </FormItem>
                                            )}
                                        </>
                                    ) : (
                                        <>
                                            {editorState.showMSGateway && (
                                                <FormItem label="微服务网关">
                                                    <div>
                                                        {curRule.gateway_entries.map((svc) => {
                                                        return <Tag>{svc}</Tag>;
                                                    })}
                                                    </div>
                                                </FormItem>
                                            )}
                                            {editorState.showMSService && (
                                                <FormItem label="微服务应用">
                                                    <div>
                                                        {curRule.service_entries.map((svc) => {
                                                        return <Tag>{svc}</Tag>
                                                    })}
                                                    </div>
                                                </FormItem>
                                            )}
                                        </>
                                    )}
                                </Space>
                            </div>
                        </StepItem>
                        <StepItem value={3} title={editorState.editable ? '配置泳道组服务' : '泳道组服务'} status='process'>
                            {/* 3. 配置泳道服务 */}
                            <div style={{ marginTop: 24 }}>
                                <FormItem>
                                    <div>
                                        {editorState.editable ? (
                                            <Transfer
                                                search={true}
                                                data={serviceOptions.filter(opt => {
                                                    return !curRule.service_entries.some(entry => entry === opt.value) && !curRule.gateway_entries.some(entry => entry === opt.value);
                                                })}
                                                value={curRule.destinations}
                                                onChange={(val, ctx) => {
                                                    // ctx.movedValue 是本次被移动的项（可能为空），val 是当前目标列表（或源列表，取决于 type）
                                                    const moved = (ctx?.movedValue || []) as TransferValue[];
                                                    if (ctx?.type === 'target') {
                                                        // 添加到 destinations
                                                        setCurRule(prev => ({
                                                            ...prev,
                                                            destinations: [
                                                                ...prev.destinations,
                                                                ...moved.map(v => String(v)).filter(v => !prev.destinations.includes(v)),
                                                            ],
                                                        }));
                                                    } else if (ctx?.type === 'source') {
                                                        // 从 destinations 中移除被移回源列表的项
                                                        const movedStr = moved.map(v => String(v));
                                                        setCurRule(prev => ({
                                                            ...prev,
                                                            destinations: prev.destinations.filter(d => !movedStr.includes(d)),
                                                        }));
                                                    } else {
                                                        // 兜底：直接以传入的 val 作为最新目标列表
                                                        setCurRule(prev => ({
                                                            ...prev,
                                                            destinations: (val as string[]).map(v => String(v)),
                                                        }));
                                                    }
                                                }}
                                            />
                                        ) : (
                                            <>
                                                {curRule.destinations.map((svc) => {
                                                    return <Tag>{svc}</Tag>;
                                                })}
                                            </>
                                        )}
                                    </div>
                                </FormItem>
                            </div>
                        </StepItem>
                    </Steps>
                </Form>
            </div>
        )
    };

    return (
        <div style={{ padding: 24 }}>
            {renderEditor()}
            {editorState.publishView && (
                <PublishForm
                    ruleId={editGroup?.id || ''}
                    ruleName={editGroup?.name || ''}
                    resource={PolicySourceType.LaneRules}
                    visible={editorState.publishView}
                    close={() => {
                        setEditorState(prev => ({ ...prev, publishView: false }));
                    }}
                />
            )}
            <StickyTool
                style={{ zIndex: 1000 }}
                placement='right-bottom'
                offset={[-10, 200]}
            >
                <StickyItem
                    label={editorState.editable ? '保存' : '编辑'}
                    icon={!editorState.editable ?
                        <Edit1Icon onClick={() => {
                            setEditorState({ ...editorState, editable: true });

                        }} />
                        :
                        <Button
                            disabled={!editGroup?.editable}
                            variant="text"
                            shape="square"
                            onClick={() => { form.submit(); }}
                        >
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
                                fetchLaneGroupDetail(editGroup)
                                setEditorState({ ...editorState, editable: false });
                            }
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
        </div>
    )
};

export default React.memo(LaneGroupEditor);