import React, { useState, useEffect } from 'react';
import { Drawer, Form, Input, Space, Button, InputNumber, Radio, Switch, Select, Transfer } from "tdesign-react";
import type { FormProps } from 'tdesign-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { saveUsers, updateUsers } from 'modules/user/users';
import { selectUser } from 'modules/user/users';
import { saveUserGroups, selectUserGroup, updateUserGroups } from 'modules/user/groups';
import { describeAllUsers, User } from 'services/users';
import { describeUserGroupDetail } from 'services/user_group';


const { FormItem } = Form;

interface IGroupEditorProps {
    modify: boolean;
    op: string;
    closeDrawer: () => void;
    refresh: () => void;
    visible: boolean;
}

const GroupEditor: React.FC<IGroupEditorProps> = ({ visible, op, modify, closeDrawer, refresh }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const currentUserGroup = useAppSelector(selectUserGroup);
    const {
        id,
        name,
        token_enable,
        comment,
        metadata,
    } = currentUserGroup;

    const [searchState, setSearchState] = React.useState<{
        users: { value: string; label: string }[];
        userLoading: boolean,
    }>({
        users: [],
        userLoading: false,
    });

    React.useEffect(() => {
        if (visible) {
            fetchUserGroupDetail();
            fetchUserData();
        }
    }, [id, visible]);

    const fetchUserGroupDetail = async () => {
        try {
            // 默认只查询简要信息
            const response = await describeUserGroupDetail({
                id: id as string,
            });
            if (!response.userGroup) {
                openErrNotification("获取用户组详情失败", "用户组不存在");
                return;
            }
            const users = response.userGroup.relation.users ? response.userGroup.relation.users.map((user) => user.id) : []
            const group_labels = response.userGroup.metadata ? Object.entries(metadata).map(([key, value]) => ({ key, value })) : [];

            form.setFieldsValue({
                name: response.userGroup.name,
                comment: response.userGroup.comment,
                token_enable: response.userGroup.token_enable,
                users: users,
                group_labels: group_labels,
            })
        } catch (error: Error | any) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification("获取用户组详情失败", error);
        }
    }

    const fetchUserData = async () => {
        setSearchState(s => ({ ...s, user_loading: true }));
        try {
            // 请求可能存在跨域问题
            const response = await describeAllUsers();
            const users = response.map((user: User) => {
                return {
                    value: user.id,
                    label: user.name,
                    disabled: user.user_type !== 'sub',
                }
            })
            setSearchState(s => ({ ...s, users: users, user_loading: false }));
        } catch (error) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification('获取用户列表失败', (error as Error).message);
        }
    }

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log(e);
        if (e.validateResult !== true) {
            return;
        }

        const inusers = form.getFieldValue('users') as string[]
        const labels = form.getFieldValue('group_labels') as { key: string, value: string }[]

        const newData = {
            id: id,
            name: form.getFieldValue('name') as string,
            token_enable: form.getFieldValue('token_enable') as boolean,
            comment: form.getFieldValue('comment') as string,
            metadata: labels.reduce((acc: { [key: string]: string }, label: { key: string, value: string }) => {
                acc[label.key] = label.value;
                return acc;
            }, {}),
            relation: {
                group_id: id,
                users: inusers.map((v) => {
                    return {
                        id: v,
                    }
                }),
            },
        }

        let result;
        if (modify) {
            result = await dispatch(updateUserGroups({ state: { ...newData } }))
        } else {
            result = await dispatch(saveUserGroups({ state: { ...newData } }))
        }

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', modify ? '修改用户组信息成功' : '创建用户组成功');
            closeDrawer();
            refresh();
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
            <FormItem label={'用户组名'} name={'name'} initialData={name}
                rules={[
                    { required: true, message: '用户组名不能为空' },
                    { max: 64, message: '长度不超过64个字符' },
                ]}
            >
                <Input disabled={op === 'edit'} />
            </FormItem>
            <FormItem label={'备注'} name={'comment'} initialData={comment}
                rules={[
                    { max: 255, message: '长度不超过255个字符' }
                ]}>
                <Input />
            </FormItem>
            <FormItem label="Token 启用状态" name="token_enable" initialData={op === 'create' ? true : token_enable}>
                <Switch disabled={op === 'view'} size="large" label={['启用', '禁用']} />
            </FormItem>
            <FormItem
                label={'用户'}
                name={'users'}
            >
                <Transfer
                    search={true}
                    data={searchState.users}
                />
            </FormItem>
            <LabelInput form={form} label='用户组标签' name='group_labels' disabled={op === 'view'} />
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
                header={op === 'view' ? '详细' : modify ? "编辑" : "创建"}
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

export default React.memo(GroupEditor);