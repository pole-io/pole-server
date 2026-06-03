import React, { useState } from 'react';
import { Drawer, Form, Input, Space, Button, Select } from "tdesign-react";
import type { FormProps } from 'tdesign-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { saveConfigGroups, selectConfigGroup, updateConfigGroups } from 'modules/configuration/group';
import { describeAllNamespaces, NamespaceView } from 'services/namespace';
import { Op } from 'services/types';
import { listAllNamespaces, selectNamespace } from 'modules/namespace';
import namespace from 'router/modules/namespace';

const { FormItem } = Form;

interface IConfigGroupEditorProps {
    op: Op;
    closeDrawer: () => void;
    visible: boolean;
}

const ConfigGroupEditor: React.FC<IConfigGroupEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const groupState = useAppSelector(selectConfigGroup);
    const { editGroup } = groupState;

    React.useEffect(() => {
        if (visible && editGroup) {
            form.setFieldsValue({
                name: editGroup?.name,
                namespace: editGroup?.namespace,
                comment: editGroup?.comment,
                department: editGroup?.department,
                business: editGroup?.business,
                group_labels: editGroup?.metadata ? Object.entries(editGroup.metadata).map(([key, value]) => ({ key, value })) : [],
            });
        }
    }, [visible, editGroup]);

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

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log(e);
        if (e.validateResult !== true) {
            return;
        }

        const labels = form.getFieldValue('group_labels') as { key: string, value: string }[]

        const newData = {
            id: editGroup?.id || 0,
            name: form.getFieldValue('name') as string,
            namespace: form.getFieldValue('namespace') as string,
            comment: form.getFieldValue('comment') as string,
            department: form.getFieldValue('department') as string,
            business: form.getFieldValue('business') as string,
            metadata: labels.reduce((acc: { [key: string]: string }, { key, value }) => {
                acc[key] = value;
                return acc;
            }, {}),
        }

        let result;
        if (op === 'edit') {
            result = await dispatch(updateConfigGroups({ param: { ...newData } }))
        } else {
            result = await dispatch(saveConfigGroups({ param: { ...newData } }))
        }

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'edit' ? '修改配置分组成功' : '创建配置分组成功');
            closeDrawer();
        }
    }

    const groupForm = (
        <Form
            form={form}
            layout="vertical"
            // labelWidth={120}
            // labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <FormItem label={'命名空间'} name={'namespace'}>
                <Select
                    readonly={op === 'edit'}
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
                showErrorMessage={op === 'create'}
                rules={[
                    { required: true, message: '请输入配置分组名称' },
                    { pattern: /^[a-zA-Z0-9._-]+$/, message: '只允许数字、英文字母、.、-、_' },
                    { max: 128, message: '长度不超过128个字符' }
                ]}>
                <Input
                    readonly={op === 'edit'}
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
            <LabelInput form={form} label='分组标签' name='group_labels' editable={true} />
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
    )

    return (
        <div>
            <Drawer
                size='large'
                header={op === 'edit' ? "编辑配置分组" : "创建配置分组"}
                footer={null}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
            >
                {groupForm}
            </Drawer>
        </div>
    );
}

export default React.memo(ConfigGroupEditor);