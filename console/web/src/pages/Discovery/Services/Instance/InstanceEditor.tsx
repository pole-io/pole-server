import React, {  } from 'react';
import { Drawer, Form, Input, Space, Button, InputNumber, Switch, Select, Descriptions, Tag } from "tdesign-react";
import type { FormProps } from 'tdesign-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { selectInstance, updateInstances, saveInstances } from 'modules/discovery/instance';
import { HEALTH_CHECK_STRUCT } from 'services/instance';
import { get } from 'lodash';

const { FormItem } = Form;
const { DescriptionsItem } = Descriptions;

interface IInstanceEditorProps {
    op: string;
    namespace: string;
    service: string;
    closeDrawer: () => void;
    visible: boolean;
}

const HealthCheckTypeOptions = [{ label: '心跳上报', value: 1 }, { label: 'TCP 探测', value: 2, disabled: true }, { label: 'HTTTP 探测', value: 3, disabled: true }]

const InstanceEditor: React.FC<IInstanceEditorProps> = ({ visible, op, closeDrawer, namespace, service }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const insState = useAppSelector(selectInstance);
    const { editIns } = insState;

    React.useEffect(() => {
        form.setFieldsValue({
            namespace: namespace,
            service: service,
            host: editIns?.host || '',
            port: editIns?.port || 0,
            protocol: editIns?.protocol || '',
            version: editIns?.version || '',
            weight: editIns?.weight || 0,
            healthy: editIns?.healthy || false,
            isolate: editIns?.isolate || false,
            enableHealthCheck: editIns?.enableHealthCheck || false,
            healthCheck: editIns?.healthCheck || {},
            // 健康检查类型默认为心跳上报
            instance_labels: editIns?.metadata ? Object.entries(editIns.metadata).map(([key, value]) => ({ key, value })) : [],
            location: editIns?.location || {},
        })
    }, [editIns])

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }

        const labels = form.getFieldValue('instance_labels') as { key: string, value: string }[]

        let newData = {
            id: editIns?.id || '',
            namespace: form.getFieldValue('namespace') as string,
            service: form.getFieldValue('service') as string,
            host: form.getFieldValue('host') as string,
            port: form.getFieldValue('port') as number,
            protocol: form.getFieldValue('protocol') as string,
            version: form.getFieldValue('version') as string,
            weight: form.getFieldValue('weight') as number,
            healthy: form.getFieldValue('healthy') as boolean,
            isolate: form.getFieldValue('isolate') as boolean,
            location: {
                region: form.getFieldValue(['location', 'region']) as string,
                zone: form.getFieldValue(['location', 'zone']) as string,
                campus: form.getFieldValue(['location', 'campus']) as string,
            },
            healthCheck: {} as HEALTH_CHECK_STRUCT,
            enableHealthCheck: form.getFieldValue('enableHealthCheck') as boolean,
            metadata: labels.reduce((acc: { [key: string]: string }, label: { key: string, value: string }) => {
                acc[label.key] = label.value;
                return acc;
            }, {}),
        }
        if (newData.enableHealthCheck) {
            newData.healthCheck = {
                type: form.getFieldValue(['healthCheck', 'type']) as number,
                heartbeat: {
                    ttl: form.getFieldValue(['healthCheck', 'heartbeat', 'ttl']) as number,
                }
            };
        }

        console.log(e, newData);
        let result;
        if (op === 'edit') {
            result = await dispatch(updateInstances({ param: { ...newData } }))
        } else {
            result = await dispatch(saveInstances({ param: { ...newData } }))
        }

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'edit' ? '修改服务实例成功' : '创建服务实例成功');
            closeDrawer();
        }
    };

    const instanceView = (
        <>
            <Descriptions column={1}>
                <DescriptionsItem label="命名空间">{editIns?.namespace}</DescriptionsItem>
                <DescriptionsItem label="服务">{editIns?.service}</DescriptionsItem>
                <DescriptionsItem label="主机">{editIns?.host}</DescriptionsItem>
                <DescriptionsItem label="端口">{editIns?.port}</DescriptionsItem>
                <DescriptionsItem label="协议">{editIns?.protocol}</DescriptionsItem>
                <DescriptionsItem label="版本">{editIns?.version}</DescriptionsItem>
                <DescriptionsItem label="权重">{editIns?.weight}</DescriptionsItem>
                <DescriptionsItem label="健康状态">{editIns?.healthy ? '健康' : '异常'}</DescriptionsItem>
                <DescriptionsItem label="隔离状态">{editIns?.isolate ? '隔离' : '正常'}</DescriptionsItem>
                <DescriptionsItem label="健康检查">{editIns?.enableHealthCheck ? '开启' : '关闭'}</DescriptionsItem>
                {editIns?.enableHealthCheck && <>
                    <DescriptionsItem label="健康检查类型">
                        {editIns?.healthCheck?.type === 1 ? '心跳上报' : editIns?.healthCheck?.type === 2 ? 'TCP 探测' : editIns?.healthCheck?.type === 3 ? 'HTTP 探测' : ''}
                    </DescriptionsItem>
                    {editIns?.healthCheck?.type === 1 && <>
                        <DescriptionsItem label="心跳上报 TTL">
                            {get(editIns?.healthCheck, 'heartbeat.ttl', 5)} 秒
                        </DescriptionsItem>
                    </>}
                </>}
                <DescriptionsItem label="位置">
                    <Space>
                        <Tag>区域: {editIns?.location.region === '' ? '-' : editIns?.location.region}</Tag>
                        <Tag>可用区: {editIns?.location.zone === '' ? '-' : editIns?.location.zone}</Tag>
                        <Tag>机房: {editIns?.location.campus === '' ? '-' : editIns?.location.campus}</Tag>
                    </Space>
                </DescriptionsItem>
                <DescriptionsItem label="实例标签">
                    <Space>
                        {Object.entries(editIns?.metadata || {}).map((item, index) => (
                            <Tag key={index} variant='outline'>
                                {item[0]} : {item[1]}
                            </Tag>
                        ))}
                    </Space>
                </DescriptionsItem>
            </Descriptions>
        </>
    )

    const instanceForm = (
        <>
            <Form
                form={form}
                layout="vertical"
                labelWidth={120}
                labelAlign={'left'}
                onSubmit={onSubmit}
            >
                <FormItem label={'命名空间'} name={'namespace'} initialData={namespace}>
                    <Input readonly={true} />
                </FormItem>
                <FormItem label={'服务'} name={'service'} initialData={service}>
                    <Input readonly={true} />
                </FormItem>
                <FormItem label={'主机'} name={'host'} rules={[
                    { required: true, message: '请输入服务IP' },
                    { pattern: /^(((\d{1,2}|1\d{2}|2[0-4]\d|25[0-5])\.){3}(\d{1,2}|1\d{2}|2[0-4]\d|25[0-5]))$|^([a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?)$/, message: '请输入正确的IP地址或域名' },
                    { max: 15, message: '长度不超过15个字符' }
                ]}>
                    <Input readonly={op === 'edit'} />
                </FormItem>
                <FormItem label={'端口'} name={'port'} rules={[
                    { required: true, message: '请输入服务端口' },
                    { pattern: /^[0-9]{1,5}$/, message: '请输入正确的端口号' },
                ]}>
                    <InputNumber
                        readonly={op === 'edit'}
                        min={1}
                        max={65535}
                        step={1}
                    />
                </FormItem>
                <FormItem label={'协议'} name={'protocol'} rules={[
                    { required: false, message: '请输入服务协议, TCP/UDP/HTTP/gRPC/DUBBO/...' },
                    { max: 128, message: '长度不超过128个字符' }
                ]}>
                    <Select
                        readonly={op === 'view'}
                        creatable={true}
                        filterable={true}
                        options={[
                            { label: 'TCP', value: 'TCP' },
                            { label: 'UDP', value: 'UDP' },
                            { label: 'HTTP', value: 'HTTP' },
                            { label: 'gRPC', value: 'gRPC' },
                            { label: 'DUBBO', value: 'DUBBO' },
                        ]}
                    />
                </FormItem>
                <FormItem label={'版本'} name={'version'} rules={[
                    { required: false, message: '请输入服务版本' },
                    { max: 128, message: '长度不超过128个字符' }
                ]}>
                    <Input readonly={op === 'view'} />
                </FormItem>
                <FormItem label={'权重'} name={'weight'} rules={[
                    { required: false, message: '请输入服务权重' },
                    { pattern: /^[0-9]{1,5}$/, message: '请输入正确的权重' },
                ]}>
                    <InputNumber
                        min={0}
                        max={100}
                        step={1}
                    />
                </FormItem>
                <FormItem label="隔离状态" name="isolate">
                    <Switch size="large" label={['开', '关']} />
                </FormItem>
                <FormItem label="健康状态" name="healthy">
                    <Switch size="large" label={['健康', '异常']} />
                </FormItem>
                <FormItem label="健康检查" name="enableHealthCheck">
                    <Switch size="large" label={['开', '关']} />
                </FormItem>
                <FormItem shouldUpdate={(prev, next) => prev.enableHealthCheck !== next.enableHealthCheck}>
                    {({ getFieldValue }) => {
                        if (getFieldValue('enableHealthCheck') === true) {
                            return (
                                <FormItem label="健康检查类型" key="ice" name={['healthCheck', 'type']}>
                                    <Select
                                        options={HealthCheckTypeOptions}
                                    />
                                </FormItem>
                            );
                        }
                        return <></>;
                    }}
                </FormItem>
                <FormItem shouldUpdate={(prev, next) => {
                    const enableChange = prev.enableHealthCheck !== next.enableHealthCheck;
                    const typeChange = prev?.healthCheck?.type !== next?.healthCheck?.type;
                    return enableChange || typeChange;
                }}>
                    {({ getFieldValue }) => {
                        // 心跳健康检查
                        if (getFieldValue('enableHealthCheck') === true && getFieldValue(['healthCheck', 'type']) === 1) {
                            return (
                                <FormItem label="心跳上报 TTL" key="ttl" name={['healthCheck', 'heartbeat', 'ttl']}>
                                    <InputNumber
                                        min={1}
                                        max={60}
                                        step={1}
                                        suffix="秒"
                                    />
                                </FormItem>
                            );
                        }
                        return <></>
                    }}
                </FormItem>
                <Space>
                    <FormItem name={['location', 'region']} label={'位置'}>
                        <Input label={'区域=>'} placeholder='' />
                    </FormItem>
                    <FormItem name={['location', 'zone']}>
                        <Input label={'可用区=>'} placeholder='' />
                    </FormItem>
                    <FormItem name={['location', 'campus']}>
                        <Input label={'机房=>'} placeholder='' />
                    </FormItem>
                </Space>
                <LabelInput form={form} label='实例标签' name='instance_labels' editable={true} />
                <FormItem style={{ marginTop: 100 }}>
                    <Space>
                        <Button theme="primary" type="submit">
                            提交
                        </Button>
                        <Button theme="default" onClick={closeDrawer}>
                            取消
                        </Button>
                    </Space>
                </FormItem>
            </Form>
        </>
    )

    return (
        <div>
            <Drawer
                size='large'
                header={op === 'view' ? '实例详细' : op === 'edit' ? "编辑实例" : "创建实例"}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
                footer={null}
            >
                {op === 'view' ? instanceView : instanceForm}
            </Drawer>
        </div>
    );
}

export default React.memo(InstanceEditor);