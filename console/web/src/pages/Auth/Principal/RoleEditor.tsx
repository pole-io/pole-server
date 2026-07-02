import React, { useState, useEffect } from 'react';
import { Drawer, Form, Input, Space, Button, InputNumber, Radio, Switch, Select, Transfer, Tag, Tabs } from "tdesign-react";
import type { FormProps, PageInfo } from 'tdesign-react';
import { Icon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { saveRoles, selectRole, updateRoles } from 'modules/auth/role';
import { describeAllUsers, describeUsers, User } from 'services/users';
import { describeAllUserGroups, describeUserGroups, UserGroup } from 'services/user_group';
import { describeRoles, Role } from 'services/role';
import style from './index.module.less';
import PrincipalPolicyTable from './PrincipalPolicyTable';

const { FormItem } = Form;
const { TabPanel } = Tabs;

interface IRoleEditorProps {
    modify: boolean;
    op: string;
    closeDrawer: () => void;
    refresh: () => void;
    visible: boolean;
}

const RoleEditor: React.FC<IRoleEditorProps> = ({ visible, op, modify, closeDrawer, refresh }) => {
    const [form] = Form.useForm();
    const [editable, setEditable] = React.useState(op !== 'view');
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
    const [detailRole, setDetailRole] = React.useState<Role | null>(null);

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
            setEditable(op !== 'view');

            const userlist = users ? users.map((user) => ({ value: user.id, label: user.name || '' })) : []
            const grouplist = user_groups ? user_groups.map((group) => ({ value: group.id, label: group.name || '' })) : []

            form.setFieldsValue({
                users: userlist,
                user_groups: grouplist,
                role_labels: role_labels,
            });

            if (op !== 'create' && id) {
                fetchRoleDetail();
            } else {
                setDetailRole(null);
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
            setDetailRole(role);
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
        if (op === 'create') {
            result = await dispatch(saveRoles({ state: { ...newData } }))
        } else {
            result = await dispatch(updateRoles({ state: { ...newData } }))
        }
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'create' ? '创建角色成功' : '修改角色信息成功');
            closeDrawer();
            refresh();
        }
    }

    const viewRole = detailRole || currentRole as Role;
    const viewLabels = viewRole?.metadata ? Object.entries(viewRole.metadata) : [];
    const viewUsers = viewRole?.users || [];
    const viewGroups = viewRole?.user_groups || [];

    const roleBaseView = (
        <div className={style.authDetail}>
            <div className={style.detailField}>
                <span>角色名</span>
                <strong>{viewRole?.name || '-'}</strong>
            </div>
            <div className={style.detailField}>
                <span>角色 ID</span>
                <p>{viewRole?.id || '-'}</p>
            </div>
            <div className={style.detailGrid}>
                <div className={style.detailField}>
                    <span>来源</span>
                    <p>{viewRole?.source || '-'}</p>
                </div>
                <div className={style.detailField}>
                    <span>默认角色</span>
                    <Tag theme={viewRole?.default_role ? 'success' : 'default'} variant="outline">
                        {viewRole?.default_role ? '是' : '否'}
                    </Tag>
                </div>
                <div className={style.detailField}>
                    <span>关联用户</span>
                    <p>{viewUsers.length}</p>
                </div>
                <div className={style.detailField}>
                    <span>关联用户组</span>
                    <p>{viewGroups.length}</p>
                </div>
                <div className={style.detailField}>
                    <span>时间</span>
                    <p>修改: {viewRole?.mtime || '-'}<br />创建: {viewRole?.ctime || '-'}</p>
                </div>
            </div>
            <div className={style.detailField}>
                <span>备注</span>
                <p>{viewRole?.comment || '-'}</p>
            </div>
            <div className={style.detailField}>
                <span>用户</span>
                <div className={style.detailLabels}>
                    {viewUsers.length > 0 ? viewUsers.map((user) => (
                        <em key={user.id}>{user.name || user.id}</em>
                    )) : '-'}
                </div>
            </div>
            <div className={style.detailField}>
                <span>用户组</span>
                <div className={style.detailLabels}>
                    {viewGroups.length > 0 ? viewGroups.map((group) => (
                        <em key={group.id}>{group.name || group.id}</em>
                    )) : '-'}
                </div>
            </div>
            <div className={style.detailField}>
                <span>角色标签</span>
                <div className={style.detailLabels}>
                    {viewLabels.length > 0 ? viewLabels.map(([key, value]) => (
                        <em key={key}>{key}: {value}</em>
                    )) : '-'}
                </div>
            </div>
        </div>
    );

    const roleView = (
        <Tabs defaultValue="base">
            <TabPanel value="base" label="基础信息">
                {roleBaseView}
            </TabPanel>
            <TabPanel value="permission" label="权限信息">
                <PrincipalPolicyTable principalId={viewRole?.id} principalType={3} />
            </TabPanel>
        </Tabs>
    );

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
                <Input disabled={op !== 'create'} />
            </FormItem>
            {op !== 'create' && (
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
                <Input disabled={!editable} />
            </FormItem>
            <FormItem
                label={'用户'}
                name={'users'}
            >
                <Transfer
                    search={true}
                    data={searchState.users}
                    disabled={!editable}
                />
            </FormItem>
            <FormItem
                label={'用户组'}
                name={'user_groups'}
            >
                <Transfer
                    search={true}
                    data={searchState.groups}
                    disabled={!editable}
                />
            </FormItem>
            <LabelInput
                form={form}
                label='角色标签'
                name='role_labels'
                editable={editable}
            />
            {editable && (
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
                header={op === 'create' ? "创建角色" : editable ? "编辑角色" : "角色详情"}
                footer={op === 'view' && !editable ? (
                    <Space>
                        <Button theme="primary" onClick={() => setEditable(true)}>
                            编辑
                        </Button>
                        <Button theme="default" onClick={closeDrawer}>
                            关闭
                        </Button>
                    </Space>
                ) : false}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
            >
                {editable ? userForm : roleView}
            </Drawer>
        </div>
    );
}

export default React.memo(RoleEditor);
