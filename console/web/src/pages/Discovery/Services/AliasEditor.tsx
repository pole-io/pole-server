import React, {  } from 'react';
import { Drawer, Form, Input, Select, Space, Button } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { NamespaceView } from 'services/namespace';
import { ServiceView } from 'services/service';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { saveServiceAliass, selectServiceAlias, updateServiceAliass } from 'modules/discovery/alias';
import { listAllNamespaces, selectNamespace } from 'modules/namespace';
import { listAllServices, selectService } from 'modules/discovery/service';
import { Op } from 'services/types';

const { FormItem } = Form;

interface IServiceEditorProps {
    op: Op
    closeDrawer: () => void;
    visible: boolean;
}

const ServiceAliasEditor: React.FC<IServiceEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const aliasState = useAppSelector(selectServiceAlias);
    const { editAlias } = aliasState;

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { datas: serviceDatas } = serviceState;

    React.useEffect(() => {
        if (visible) {
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
        }
    }, [visible]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log('onSubmit', e);
        if (e.validateResult !== true) {
            return;
        }

        const selectedSvc = form.getFieldValue('name') as string;
        const selectedSvcNamespace = selectedSvc.split('@')[0];
        const selectedSvcName = selectedSvc.split('@')[1];

        const newData = {
            service: selectedSvcName,
            namespace: selectedSvcNamespace,
            alias_namespace: form.getFieldValue('alias_namespace') as string,
            alias: form.getFieldValue('alias') as string,
            comment: form.getFieldValue('comment') as string,
        }
        let result;
        if (op === 'edit') {
            result = await dispatch(updateServiceAliass({ param: { ...newData } }))
        } else {
            result = await dispatch(saveServiceAliass({ param: { ...newData } }))
        }
        if (result.meta.requestStatus === 'fulfilled') {
            openInfoNotification('操作成功', op === 'edit' ? '修改服务别名成功' : '创建服务别名成功');
            closeDrawer();
        } else {
            openErrNotification('操作失败', result?.payload as string);
        }
    }

    const aliasForm = (
        <Form
            labelWidth={140}
            labelAlign={'left'}
            form={form}
            onSubmit={onSubmit}
        >
            <FormItem label="目标服务" name="name" rules={[{ required: true, message: '服务名称不能为空' }]}>
                <Select
                    filterable={true}
                    options={serviceDatas.map((item: ServiceView) => ({
                        label: `${item.name} (${item.namespace})`,
                        value: `${item.namespace}@${item.name}`,
                    }))}
                />
            </FormItem>
            <FormItem label="别名所在命名空间" name="alias_namespace" rules={[{ required: true, message: '命名空间不能为空' }]}>
                <Select
                    filterable={true}
                    options={namespaceDatas.map((item: NamespaceView) => ({
                        label: item.name,
                        value: item.name,
                    }))}
                />
            </FormItem>
            <FormItem label="别名" name="alias" rules={[{ required: true, message: '别名不能为空' }]}>
                <Input />
            </FormItem>
            <FormItem label="备注" name="comment">
                <Input />
            </FormItem>
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
    );

    return (
        <>
            <div>
                <Drawer
                    size='large'
                    header={op === 'edit' ? "编辑" : "创建"}
                    footer={false}
                    visible={visible}
                    showOverlay={false}
                    onClose={closeDrawer}
                >
                    {aliasForm}
                </Drawer>
            </div>
        </>
    )
}

export default React.memo(ServiceAliasEditor);