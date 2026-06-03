import React, {  } from 'react';
import { Drawer, Form, Input, Space, Button, Switch } from "tdesign-react";
import type { FormProps } from 'tdesign-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { saveUsers, updateUsers } from 'modules/user/users';
import { selectUser } from 'modules/user/users';
import { Label, Op } from 'services/types';

interface UserDO {
    // 用户ID
    id: string
    // 用户名称
    name: string
    // 用户密码
    password: string
    // 用户鉴权Token
    auth_token?: string
    // 该token是否被禁用
    token_enable: boolean
    // 该账户的简单描述
    comment: string
    // 账户对应邮箱
    email: string
    // 账户对应手机号
    mobile: string
    // 用户标签
    metadata: Label[]
}

const { FormItem } = Form;

interface IUserEditorProps {
    op: Op;
    closeDrawer: () => void;
    visible: boolean;
}

const UserEditor: React.FC<IUserEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const userState = useAppSelector(selectUser);
    const {editUser, viewUser} = userState;

    React.useEffect(() => {
        if (visible && editUser) {
            form.setFieldsValue({
                ...editUser,
                user_labels: editUser?.metadata ? Object.entries(editUser.metadata).map(([key, value]) => ({ key, value })) : [],
            });
        }
    }, [editUser, visible]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }

        const labels = form.getFieldValue('user_labels') as { key: string, value: string }[]

        const newData = {
            id: editUser?.id || '',
            name: form.getFieldValue('name') as string,
            password: form.getFieldValue('password') as string || '',
            token_enable: form.getFieldValue('token_enable') as boolean,
            comment: form.getFieldValue('comment') as string,
            email: form.getFieldValue('email') as string,
            mobile: form.getFieldValue('mobile') as string,
            metadata: labels.reduce((acc: { [key: string]: string }, label: { key: string, value: string }) => {
                acc[label.key] = label.value;
                return acc;
            }, {}),
        }

        let result;
        if (op === 'edit') {
            result = await dispatch(updateUsers({ param: { ...newData } }))
        } else {
            result = await dispatch(saveUsers({ param: { ...newData } }))
        }

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'edit' ? '修改用户信息成功' : '创建用户成功');
            closeDrawer();
        }
    };

    const userForm = (
        <Form
            form={form}
            layout="vertical"
            labelWidth={120}
            labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <FormItem label={'用户名'} name={'name'} initialData={name}
                rules={[
                    { required: true, message: '用户名不能为空' },
                    { max: 64, message: '长度不超过64个字符' },
                ]}
            >
                <Input disabled={op === 'edit'} />
            </FormItem>
            {op === 'create' && (
                <FormItem label={'密码'} name={'password'}
                    rules={[
                        { required: true, message: '密码不能为空' },
                        { min: 6, message: '长度不小于6个字符' },
                        { max: 255, message: '长度不超过255个字符' }
                    ]}>
                    <Input type='password' />
                </FormItem>
            )}
            <FormItem label={'备注'} name={'comment'}
                rules={[
                    { max: 255, message: '长度不超过255个字符' }
                ]}>
                <Input />
            </FormItem>
            <FormItem label="Token 启用状态" name="token_enable">
                <Switch disabled={op === 'view'} size="large" label={['启用', '禁用']} />
            </FormItem>
            <LabelInput form={form} label='用户标签' name='user_labels' editable={true} />
            {op !== 'view' && (
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
            )}
        </Form>
    )

    return (
        <div>
            <Drawer
                size='large'
                header={op === 'edit' ? "编辑用户" : "创建用户"}
                footer={false}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
            >
                {userForm}
            </Drawer>
        </div>
    );
}

export default React.memo(UserEditor);