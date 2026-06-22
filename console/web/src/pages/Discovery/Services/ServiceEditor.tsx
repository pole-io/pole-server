import React from 'react';
import { Drawer, Form, Input, Select, Space, Button } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { NamespaceView } from 'services/namespace';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { listAllServices, saveServices, selectService, updateServices } from 'modules/discovery/service';
import { Op } from 'services/types';
import { listAllNamespaces, selectNamespace } from 'modules/namespace';


const { FormItem } = Form;

interface IServiceEditorProps {
    op: Op;
    visible: boolean;
    closeDrawer: () => void;
}

const ServiceEditor: React.FC<IServiceEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { editSvc } = serviceState;

    React.useEffect(() => {
        if (visible && editSvc) {
            form.setFieldsValue({
                name: editSvc.name,
                namespace: editSvc.namespace,
                comment: editSvc.comment,
                department: editSvc.department,
                business: editSvc.business,
                service_labels: editSvc?.metadata ? Object.entries(editSvc?.metadata).map(([key, value]) => ({ key, value })) : [],
            });
        }
    }, [visible, editSvc]);

    React.useEffect(() => {
        if (visible) {
            dispatch(listAllNamespaces())
                .then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('获取命名空间列表失败', res?.payload as string);
                    }
                });
            dispatch(listAllServices())
                .then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('获取服务列表失败', res?.payload as string);
                    }
                });
        }
    }, [visible]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log(e);
        if (e.validateResult !== true) {
            return;
        }

        const labels = form.getFieldValue('service_labels') as { key: string, value: string }[]

        const newData = {
            id: editSvc?.id || '',
            name: form.getFieldValue('name') as string,
            namespace: form.getFieldValue('namespace') as string,
            comment: form.getFieldValue('comment') as string,
            department: form.getFieldValue('department') as string,
            business: form.getFieldValue('business') as string,
            metadata: labels.reduce((acc: { [key: string]: string }, { key, value }) => {
                acc[key] = value;
                return acc;
            }, {}),
            ports: '',
            owners: '',
        }

        let result;
        if (op === 'edit') {
            result = await dispatch(updateServices({ param: { ...newData } }))
        } else {
            result = await dispatch(saveServices({ param: { ...newData } }))
        }

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'edit' ? '修改命名空间成功' : '创建命名空间成功');
            closeDrawer();
        }
    };

    const serviceForm = (
        <Form
            form={form}
            layout="vertical"
            // labelWidth={120}
            // labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <FormItem
                label={'命名空间'}
                name={'namespace'}>
                <Select
                    filterable={true}
                    options={namespaceDatas.map((item: NamespaceView) => ({
                        label: item.name,
                        value: item.name,
                    }))}
                />
            </FormItem>
            <FormItem
                label={'名称'}
                name={'name'}
                rules={[
                    { required: true, message: '请输入服务名称' },
                    { pattern: /^[a-zA-Z0-9._-]+$/, message: '只允许数字、英文字母、.、-、_' },
                    { max: 128, message: '长度不超过128个字符' }
                ]}>
                <Input
                    size='large'
                    placeholder={'允许数字、英文字母、.、-、_，限制128个字符'}
                />
            </FormItem>
            <FormItem
                label={'部门'}
                name={'department'}
                rules={[{ max: 255, message: '长度不超过255个字符' }]}>
                <Input />
            </FormItem>
            <FormItem
                label={'业务'}
                name='business'
                rules={[{ max: 255, message: '长度不超过255个字符' }]}>
                <Input />
            </FormItem>
            <FormItem
                label={'描述'}
                name={'comment'}
                rules={[{ max: 1024, message: '长度不超过1024个字符' }]}
            >
                <Input />
            </FormItem>
            <LabelInput form={form} label='服务标签' name='service_labels' editable={true} />
            <FormItem style={{ marginTop: 100 }}>
                <Space>
                    <Button type="submit" theme="primary">
                        提交
                    </Button>
                    <Button type="reset" theme="default">
                        重置
                    </Button>
                </Space>
            </FormItem>
        </Form>
    )

    return (
        <div>
            <Drawer
                size='large'
                header={op === 'edit' ? "编辑服务" : "创建服务"}
                footer={false}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
            >
                {serviceForm}
            </Drawer>
        </div>
    );
}

export default React.memo(ServiceEditor);
