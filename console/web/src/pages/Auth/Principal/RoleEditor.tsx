import React, { useState, useEffect } from 'react';
import { Drawer, Form, Input, Space, Button, InputNumber, Radio, Switch, Select, Transfer } from "tdesign-react";
import type { FormProps, PageInfo } from 'tdesign-react';
import { Icon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { saveRoles, selectRole, updateRoles } from 'modules/auth/role';
import { describeAllUsers, describeUsers, User } from 'services/users';
import { describeAllUserGroups, describeUserGroups, UserGroup } from 'services/user_group';
import { describeRoles } from 'services/role';

const { FormItem } = Form;

interface IRoleEditorProps {
    modify: boolean;
    op: string;
    closeDrawer: () => void;
    refresh: () => void;
    visible: boolean;
}

const RoleEditor: React.FC<IRoleEditorProps> = ({ visible, op, modify, closeDrawer, refresh }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const currentRole = useAppSelector(selectRole);

    const [searchState, setSearchState] = React.useState<{
        users: { value: string; label: string }[];
        userLoading: boolean,
        groups: { value: string; label: string }[];
        groupsLoading: boolean;
    }>({
        users: [],
        userLoading: false,
        groups: [],
        groupsLoading: false,
    });

    const {
        id,
        name,
        comment,
        source,
        users,
        user_groups,
        metadata,
    } = currentRole;

    const role_labels = metadata ? Object.entries(metadata).map(([key, value]) => ({ key, value })) : [];

    React.useEffect(() => {
        if (visible) {

            const userlist = users ? users.map((user) => ({ value: user.id, label: user.name || '' })) : []
            const grouplist = user_groups ? user_groups.map((group) => ({ value: group.id, label: group.name || '' })) : []

            form.setFieldsValue({
                users: userlist,
                user_groups: grouplist,
                role_labels: role_labels,
            });

            // 只有在编辑时才需要获取角色详情
            if (op === 'edit' && id) {
                fetchRoleDetail();
            }
            fetchUserData();
            fetchGroupData();
        }
    }, [id, visible])

    const fetchRoleDetail = async () => {
        try {
            // 默认只查询简要信息
            const response = await describeRoles({
                limit: 10, offset: 0, berif: false, id: id,
            });
            if (response.totalCount === 0) {
                openErrNotification("获取角色详情失败", "角色不存在");
                return;
            }
            const role = response.content[0];
            const users = role.users ? role.users.map((user) => user.id) : []
            const groups = role.user_groups ? role.user_groups.map((group) => group.id) : []
            console.log(users, groups);

            form.setFieldsValue({
                users: users,
                user_groups: groups,
            })
        } catch (error: Error | any) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification("获取角色详情失败", error);
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

    const fetchGroupData = async () => {
        setSearchState(s => ({ ...s, groups_loading: true }));
        try {
            // 请求可能存在跨域问题
            const response = await describeAllUserGroups();
            const groups = response.map((group: UserGroup) => {
                return {
                    value: group.id,
                    label: group.name,
                }
            })
            setSearchState(s => ({ ...s, groups: groups, groups_loading: false }));
        } catch (error) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification('获取用户组列表失败', (error as Error).message);
        }
    }

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log(e);
        if (e.validateResult !== true) {
            return;
        }
        const labels = form.getFieldValue('role_labels') as { key: string, value: string }[]
        const users = form.getFieldValue('users') ? form.getFieldValue('users') as string[] : [];
        const groups = form.getFieldValue('user_groups') ? form.getFieldValue('user_groups') as string[] : [];
        const newData = {
            id: id,
            name: form.getFieldValue('name') as string,
            comment: form.getFieldValue('comment') as string,
            metadata: labels.reduce((acc: { [key: string]: string }, label: { key: string, value: string }) => {
                acc[label.key] = label.value;
                return acc;
            }, {}),
            users: users.map((user: string) => ({ id: user })),
            user_groups: groups.map((group: string) => ({ id: group })),
        };

        console.log("role submit", newData);

        let result;
        if (modify) {
            result = await dispatch(updateRoles({ state: { ...newData } }))
        } else {
            result = await dispatch(saveRoles({ state: { ...newData } }))
        }
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', modify ? '修改角色信息成功' : '创建角色成功');
            closeDrawer();
            refresh();
        }
    }

    const userForm = (
        <Form
            form={form}
            layout="vertical"
            labelWidth={120}
            labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <FormItem
                label={'角色名'}
                name={'name'}
                initialData={name}
                rules={[
                    { required: true, message: '角色名不能为空' },
                    { max: 64, message: '长度不超过64个字符' },
                ]}
            >
                <Input disabled={op === 'edit'} />
            </FormItem>
            {op === 'edit' && (
                <FormItem
                    label={'来源'}
                    name={'source'}
                    initialData={source}
                >
                    <Input disabled={true} />
                </FormItem>
            )}
            <FormItem
                label={'备注'}
                name={'comment'}
                initialData={comment}
                rules={[
                    { max: 255, message: '长度不超过255个字符' }
                ]}
            >
                <Input />
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
            <FormItem
                label={'用户组'}
                name={'user_groups'}
            >
                <Transfer
                    search={true}
                    data={searchState.groups}
                />
            </FormItem>
            <LabelInput
                form={form}
                label='角色标签'
                name='role_labels'
                disabled={op === 'view'}
            />
            {op !== 'view' && (
                <FormItem style={{ marginTop: 100 }}>
                    <Space>
                        <Button type="submit" theme="primary">
                            提交
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

export default React.memo(RoleEditor);