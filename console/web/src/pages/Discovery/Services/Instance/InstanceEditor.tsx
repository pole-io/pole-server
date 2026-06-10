import React from 'react';
import { Drawer, Form, Input, Space, Button, InputNumber, Switch, Select, Tag } from "tdesign-react";
import { ServerIcon } from 'tdesign-icons-react';
import type { FormProps } from 'tdesign-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { selectInstance, updateInstances, saveInstances } from 'modules/discovery/instance';
import { HEALTH_CHECK_STRUCT, HEALTH_STATUS_MAP, ISOLATE_STATUS_MAP } from 'services/instance';
import { get } from 'lodash';
import style from './index.module.less';

const { FormItem } = Form;

interface IInstanceEditorProps {
    op: string;
    namespace: string;
    service: string;
    closeDrawer: () => void;
    visible: boolean;
}

const HealthCheckTypeOptions = [{ label: '心跳上报', value: 1 }, { label: 'TCP 探测', value: 2, disabled: true }, { label: 'HTTTP 探测', value: 3, disabled: true }]

const emptyText = '-';

const isPresent = (value?: React.ReactNode) => value !== undefined && value !== null && value !== '';

const Value = ({ children }: { children?: React.ReactNode }) => (
    <div className={style.instanceDetailValue}>{isPresent(children) ? children : emptyText}</div>
)

const DetailItem = ({ label, children }: { label: string, children?: React.ReactNode }) => (
    <div className={style.instanceDetailItem}>
        <div className={style.instanceDetailLabel}>{label}</div>
        <Value>{children}</Value>
    </div>
)

const getHealthCheckText = (type?: number) => {
    if (type === 1) return '心跳上报';
    if (type === 2) return 'TCP 探测';
    if (type === 3) return 'HTTP 探测';
    return emptyText;
}

interface InstanceSummaryProps {
    address: string;
    namespace: string;
    service: string;
    protocol?: string;
    healthy?: boolean;
    isolate?: boolean;
    enableHealthCheck?: boolean;
}

const InstanceSummary: React.FC<InstanceSummaryProps> = ({
    address,
    namespace,
    service,
    protocol,
    healthy,
    isolate,
    enableHealthCheck,
}) => {
    const healthStatus = HEALTH_STATUS_MAP[String(healthy) as keyof typeof HEALTH_STATUS_MAP];
    const isolateStatus = ISOLATE_STATUS_MAP[String(isolate) as keyof typeof ISOLATE_STATUS_MAP];

    return (
        <div className={style.instanceDetailHeader}>
            <div className={style.instanceDetailIdentity}>
                <div className={style.instanceDetailIcon}>
                    <ServerIcon />
                </div>
                <div className={style.instanceDetailTitleGroup}>
                    <div className={style.instanceDetailAddress}>{address}</div>
                    <Space size={8}>
                        <Tag variant="light">{namespace}</Tag>
                        <Tag variant="light">{service}</Tag>
                        {protocol && <Tag theme="primary" variant="light">{protocol}</Tag>}
                    </Space>
                </div>
            </div>
            <div className={style.instanceStatusStrip}>
                <div className={style.instanceStatusItem}>
                    <span>健康状态</span>
                    <Tag theme={healthStatus?.theme as "success" | "danger" | "default" | "primary" | "warning"} variant="light">
                        {healthStatus?.text || emptyText}
                    </Tag>
                </div>
                <div className={style.instanceStatusItem}>
                    <span>隔离状态</span>
                    <Tag theme={isolateStatus?.theme as "success" | "danger" | "default" | "primary" | "warning"} variant="light">
                        {isolateStatus?.text || emptyText}
                    </Tag>
                </div>
                <div className={style.instanceStatusItem}>
                    <span>健康检查</span>
                    <Tag theme={enableHealthCheck ? 'success' : 'default'} variant="light">
                        {enableHealthCheck ? '开启' : '关闭'}
                    </Tag>
                </div>
            </div>
        </div>
    )
}

const InstanceEditor: React.FC<IInstanceEditorProps> = ({ visible, op, closeDrawer, namespace, service }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const insState = useAppSelector(selectInstance);
    const { editIns } = insState;
    const watchedNamespace = Form.useWatch('namespace', form) as string | undefined;
    const watchedService = Form.useWatch('service', form) as string | undefined;
    const watchedHost = Form.useWatch('host', form) as string | undefined;
    const watchedPort = Form.useWatch('port', form) as number | undefined;
    const watchedProtocol = Form.useWatch('protocol', form) as string | undefined;
    const watchedHealthy = Form.useWatch('healthy', form) as boolean | undefined;
    const watchedIsolate = Form.useWatch('isolate', form) as boolean | undefined;
    const watchedEnableHealthCheck = Form.useWatch('enableHealthCheck', form) as boolean | undefined;

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
            metadata: labels.filter((label) => label.key.trim() !== '' && label.value.trim() !== '').reduce((acc: { [key: string]: string }, label: { key: string, value: string }) => {
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

    const metadata = Object.entries(editIns?.metadata || {});
    const location = editIns?.location || {};
    const address = editIns ? `${editIns.host || emptyText}:${isPresent(editIns.port) ? editIns.port : emptyText}` : emptyText;
    const formNamespace = watchedNamespace || editIns?.namespace || namespace;
    const formService = watchedService || editIns?.service || service;
    const formHost = watchedHost || editIns?.host || '';
    const formPort = watchedPort ?? editIns?.port;
    const formAddress = formHost ? `${formHost}:${isPresent(formPort) && formPort !== 0 ? formPort : emptyText}` : '待填写实例';
    const formProtocol = watchedProtocol || editIns?.protocol || '';
    const formHealthy = watchedHealthy ?? editIns?.healthy;
    const formIsolate = watchedIsolate ?? editIns?.isolate;
    const formEnableHealthCheck = watchedEnableHealthCheck ?? editIns?.enableHealthCheck;
    const instanceView = (
        <section className={style.instanceDetailPanel}>
            <InstanceSummary
                address={address}
                namespace={editIns?.namespace || namespace}
                service={editIns?.service || service}
                protocol={editIns?.protocol}
                healthy={editIns?.healthy}
                isolate={editIns?.isolate}
                enableHealthCheck={editIns?.enableHealthCheck}
            />

            <div className={style.instanceDetailSections}>
                <section className={style.instanceDetailSection}>
                    <div className={style.instanceDetailSectionTitle}>运行信息</div>
                    <div className={style.instanceDetailGrid}>
                        <DetailItem label="命名空间">{editIns?.namespace}</DetailItem>
                        <DetailItem label="服务">{editIns?.service}</DetailItem>
                        <DetailItem label="主机">{editIns?.host}</DetailItem>
                        <DetailItem label="端口">{editIns?.port}</DetailItem>
                        <DetailItem label="协议">{editIns?.protocol}</DetailItem>
                        <DetailItem label="版本">{editIns?.version}</DetailItem>
                        <DetailItem label="权重">{editIns?.weight}</DetailItem>
                    </div>
                </section>

                <section className={style.instanceDetailSection}>
                    <div className={style.instanceDetailSectionTitle}>健康检查</div>
                    <div className={style.instanceDetailGrid}>
                        <DetailItem label="状态">{editIns?.enableHealthCheck ? '开启' : '关闭'}</DetailItem>
                        <DetailItem label="检查方式">{editIns?.enableHealthCheck ? getHealthCheckText(editIns?.healthCheck?.type) : emptyText}</DetailItem>
                        <DetailItem label="心跳 TTL">
                            {editIns?.enableHealthCheck && editIns?.healthCheck?.type === 1 ? `${get(editIns?.healthCheck, 'heartbeat.ttl', 5)} 秒` : emptyText}
                        </DetailItem>
                    </div>
                </section>

                <section className={style.instanceDetailSection}>
                    <div className={style.instanceDetailSectionTitle}>位置</div>
                    <Space breakLine>
                        <Tag>区域: {location.region || emptyText}</Tag>
                        <Tag>可用区: {location.zone || emptyText}</Tag>
                        <Tag>机房: {location.campus || emptyText}</Tag>
                    </Space>
                </section>

                <section className={style.instanceDetailSection}>
                    <div className={style.instanceDetailSectionTitle}>实例标签</div>
                    {metadata.length > 0 ? (
                        <Space breakLine>
                            {metadata.map(([key, value]) => (
                                <Tag key={key} theme="primary" variant="light">
                                    {key}: {value}
                                </Tag>
                            ))}
                        </Space>
                    ) : (
                        <Value>{emptyText}</Value>
                    )}
                </section>
            </div>
        </section>
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
                <section className={style.instanceDetailPanel}>
                    <InstanceSummary
                        address={formAddress}
                        namespace={formNamespace}
                        service={formService}
                        protocol={formProtocol}
                        healthy={formHealthy}
                        isolate={formIsolate}
                        enableHealthCheck={formEnableHealthCheck}
                    />
                    <div className={style.instanceDetailSections}>
                    <section className={style.instanceDetailSection}>
                        <div className={style.instanceDetailSectionTitle}>运行信息</div>
                        <div className={style.instanceFormGrid}>
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
                        </div>
                    </section>

                    <section className={style.instanceDetailSection}>
                        <div className={style.instanceDetailSectionTitle}>健康检查</div>
                        <div className={style.instanceSwitchGrid}>
                            <FormItem label="隔离状态" name="isolate">
                                <Switch size="large" label={['开', '关']} />
                            </FormItem>
                            <FormItem label="健康状态" name="healthy">
                                <Switch size="large" label={['健康', '异常']} />
                            </FormItem>
                            <FormItem label="健康检查" name="enableHealthCheck">
                                <Switch size="large" label={['开', '关']} />
                            </FormItem>
                        </div>
                        <div className={style.instanceFormGrid}>
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
                        </div>
                    </section>

                    <section className={style.instanceDetailSection}>
                        <div className={style.instanceDetailSectionTitle}>位置</div>
                        <div className={style.instanceFormGrid}>
                            <FormItem name={['location', 'region']} label={'区域'}>
                                <Input placeholder='区域' />
                            </FormItem>
                            <FormItem name={['location', 'zone']} label={'可用区'}>
                                <Input placeholder='可用区' />
                            </FormItem>
                            <FormItem name={['location', 'campus']} label={'机房'}>
                                <Input placeholder='机房' />
                            </FormItem>
                        </div>
                    </section>

                    <section className={style.instanceDetailSection}>
                        <div className={style.instanceDetailSectionTitle}>实例标签</div>
                        <LabelInput form={form} label='实例标签' name='instance_labels' editable={true} />
                    </section>
                    </div>
                </section>
                <FormItem className={style.instanceFormFooter}>
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
                size={op === 'view' ? '680px' : 'large'}
                header={op === 'view' ? '实例详情' : op === 'edit' ? "编辑实例" : "创建实例"}
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
