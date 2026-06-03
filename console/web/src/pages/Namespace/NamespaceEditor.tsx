import React, { useState, useEffect } from 'react';
import { Drawer, Form, Input, Select, Radio, RadioGroup, Space, Button } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import { describeAllNamespaces, NamespaceView } from 'services/namespace';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { saveNamespace, updateNamespace, selectNamespace, listAllNamespaces } from 'modules/namespace';
import { VisibilityMode_Single, VisibilityMode_All, VisibilityMode_Specified, CheckVisibilityMode } from 'utils/visible';
import LabelInput from 'components/LabelInput';
import { Label, Op } from 'services/types';

const { FormItem, FormList } = Form;

interface NamespaceEditorProps {
    op: Op;
    closeDrawer: () => void;
    visible: boolean;
}

const NamespaceEditor: React.FC<NamespaceEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();

    const dispatch = useAppDispatch();
    const namespaceState = useAppSelector(selectNamespace);
    const { editNs, datas: allNamespaces } = namespaceState;

    React.useEffect(() => {
        if (visible) {
            dispatch(listAllNamespaces())
                .then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('获取命名空间列表失败', res?.payload as string);
                    }
                });
        }
    }, [visible]);

    React.useEffect(() => {
        if (visible && editNs) {
            form.setFieldsValue({
                name: editNs.name,
                comment: editNs.comment,
                service_export_to: editNs.service_export_to,
                visibility_mode: CheckVisibilityMode(editNs.service_export_to, editNs.name),
                namespace_labels: editNs?.metadata ? Object.entries(editNs.metadata).map(([key, value]) => ({ key, value } as Label)) : [],
            });
        }
    }, [visible, editNs]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log('onSubmit', e);
        if (e.validateResult !== true) {
            return;
        }

        const labels = form.getFieldValue('namespace_labels') as { key: string, value: string }[]
        const data = {
            name: form.getFieldValue('name') as string,
            comment: form.getFieldValue('comment') as string,
            service_export_to: form.getFieldValue('service_export_to') as string[],
            metadata: labels.reduce((acc: { [key: string]: string }, { key, value }) => {
                acc[key] = value;
                return acc;
            }, {}),
        }
        let result;
        if (op === 'edit') {
            result = await dispatch(updateNamespace({ param: data }))
        } else {
            result = await dispatch(saveNamespace({ param: data }))
        }
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'edit' ? '修改命名空间成功' : '创建命名空间成功');
            closeDrawer();
        }
    };

    const namespaceForm = (
        <Form
            form={form}
            layout="vertical"
            // labelWidth={120}
            // labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <FormItem
                label={'名称'}
                name={"name"}
                showErrorMessage={op === 'create'}
                rules={[
                    { required: true, message: '命名空间不能为空' },
                    { pattern: /^[a-zA-Z0-9._-]+$/, message: '只允许数字、英文字母、.、-、_' },
                    { max: 128, message: '长度不超过128个字符' }
                ]}>
                <Input readonly={op === 'edit'} />
            </FormItem>
            <FormItem
                label={'描述'}
                name={"comment"}
                rules={[{ max: 1024, message: '长度不超过1024个字符' }]}>
                <Input />
            </FormItem>
            <LabelInput form={form} label='标签' name='namespace_labels' editable={true} />
            <FormItem
                label={'可见性'}
                name={"visibility_mode"}
                tips={'当前命名空间下的服务被允许可见的命名空间列表'}
            >
                <RadioGroup>
                    <Radio value={VisibilityMode_Single}>{'仅当前命名空间'}</Radio>
                    <Radio value={VisibilityMode_All}>{'全部命名空间（包括新增）'}</Radio>
                    <Radio value={VisibilityMode_Specified}>{'指定命名空间'}</Radio>
                </RadioGroup>
            </FormItem>
            <FormItem shouldUpdate={(prev, next) => prev.visibility_mode !== next.visibility_mode}>
                {({ getFieldValue, setFieldsValue }) => {
                    const ret = getFieldValue('visibility_mode') as string;
                    if (ret !== VisibilityMode_Specified) {
                        return <></>;
                    }
                    return (
                        <FormItem label={'选择命名空间'}
                            name={"service_export_to"}
                            tips={'当前命名空间下的服务被允许可见的命名空间列表'}
                        >
                            <Select
                                multiple={true}
                                options={[{ label: '当前全部命名空间', value: '__all__', checkAll: true }, ...allNamespaces.map((item: NamespaceView) => ({
                                    label: item.name,
                                    value: item.name,
                                }))]}
                            />
                        </ FormItem>
                    )
                }}
            </FormItem>
            <FormItem style={{ marginTop: 100 }}>
                <Space>
                    <Button type="submit" theme="primary">
                        提交
                    </Button>
                    <Button type="reset" theme="default" onClick={() => closeDrawer()}>
                        取消
                    </Button>
                </Space>
            </FormItem>
        </Form>
    )

    return (
        <div>
            <Drawer
                size='large'
                header={op === 'edit' ? "编辑命名空间" : "创建命名空间"}
                footer={false}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
            >
                {namespaceForm}
            </Drawer>
        </div>
    );
}

export default React.memo(NamespaceEditor);