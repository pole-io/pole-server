import React, { useState, useEffect } from 'react';
import { Drawer, Form, Input, Space, Button, InputNumber, Radio, Switch, Select, Transfer, Tag, Tabs } from "tdesign-react";
import type { FormProps } from 'tdesign-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import LabelInput from 'components/LabelInput';
import { saveUsers, updateUsers } from 'modules/user/users';
import { selectUser } from 'modules/user/users';
import { saveUserGroups, selectUserGroup, updateUserGroups } from 'modules/user/groups';
import { describeAllUsers, User } from 'services/users';
import { describeUserGroupDetail, UserGroup } from 'services/user_group';
import style from './index.module.less';
import PrincipalPolicyTable from './PrincipalPolicyTable';


const { FormItem } = Form;
const { TabPanel } = Tabs;

interface IGroupEditorProps {
    modify: boolean;
    op: string;
    closeDrawer: () => void;
    refresh: () => void;
    visible: boolean;
}

const GroupEditor: React.FC<IGroupEditorProps> = ({ visible, op, modify, closeDrawer, refresh }) => {
    const [form] = Form.useForm();
    const [editable, setEditable] = React.useState(op !== 'view');
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
    const [detailGroup, setDetailGroup] = React.useState<UserGroup | null>(null);

    React.useEffect(() => {
        if (visible) {
            setEditable(op !== 'view');
            if (op !== 'create' && id) {
                fetchUserGroupDetail();
            } else {
                setDetailGroup(null);
                form.reset();
            }
            fetchUserData();
        }
    }, [id, visible]);

    const fetchUserGroupDetail = async () => {
        if (!id) {
            return;
        }
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
            const group_labels = response.userGroup.metadata ? Object.entries(response.userGroup.metadata).map(([key, value]) => ({ key, value })) : [];

            setDetailGroup(response.userGroup);
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
        if (op === 'create') {
            result = await dispatch(saveUserGroups({ state: { ...newData } }))
        } else {
            result = await dispatch(updateUserGroups({ state: { ...newData } }))
        }

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'create' ? '创建用户组成功' : '修改用户组信息成功');
            closeDrawer();
            refresh();
        }
    };

    const viewGroup = detailGroup || currentUserGroup as UserGroup;
    const labels = viewGroup?.metadata ? Object.entries(viewGroup.metadata) : [];
    const users = viewGroup?.relation?.users || [];

    const groupBaseView = (
        <div className={style.authDetail}>
            <div className={style.detailField}>
                <span>用户组名</span>
                <strong>{viewGroup?.name || '-'}</strong>
            </div>
            <div className={style.detailField}>
                <span>用户组 ID</span>
                <p>{viewGroup?.id || '-'}</p>
            </div>
            <div className={style.detailGrid}>
                <div className={style.detailField}>
                    <span>Token 状态</span>
                    <Tag theme={viewGroup?.token_enable ? 'success' : 'danger'} variant="outline">
                        {viewGroup?.token_enable ? '启用中' : '禁用中'}
                    </Tag>
                </div>
                <div className={style.detailField}>
                    <span>用户数量</span>
                    <p>{viewGroup?.user_count ?? users.length ?? 0}</p>
                </div>
                <div className={style.detailField}>
                    <span>来源</span>
                    <p>{viewGroup?.source || '-'}</p>
                </div>
                <div className={style.detailField}>
                    <span>时间</span>
                    <p>修改: {viewGroup?.mtime || '-'}<br />创建: {viewGroup?.ctime || '-'}</p>
                </div>
            </div>
            <div className={style.detailField}>
                <span>备注</span>
                <p>{viewGroup?.comment || '-'}</p>
            </div>
            <div className={style.detailField}>
                <span>成员用户</span>
                <div className={style.detailLabels}>
                    {users.length > 0 ? users.map((user) => (
                        <em key={user.id}>{user.name || user.id}</em>
                    )) : '-'}
                </div>
            </div>
            <div className={style.detailField}>
                <span>用户组标签</span>
                <div className={style.detailLabels}>
                    {labels.length > 0 ? labels.map(([key, value]) => (
                        <em key={key}>{key}: {value}</em>
                    )) : '-'}
                </div>
            </div>
        </div>
    );

    const groupView = (
        <Tabs defaultValue="base">
            <TabPanel value="base" label="基础信息">
                {groupBaseView}
            </TabPanel>
            <TabPanel value="permission" label="权限信息">
                <PrincipalPolicyTable principalId={viewGroup?.id} principalType={2} />
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
            <FormItem label={'用户组名'} name={'name'} initialData={name}
                rules={[
                    { required: true, message: '用户组名不能为空' },
                    { max: 64, message: '长度不超过64个字符' },
                ]}
            >
                <Input disabled={op !== 'create'} />
            </FormItem>
            <FormItem label={'备注'} name={'comment'} initialData={comment}
                rules={[
                    { max: 255, message: '长度不超过255个字符' }
                ]}>
                <Input disabled={!editable} />
            </FormItem>
            <FormItem label="Token 启用状态" name="token_enable" initialData={op === 'create' ? true : token_enable}>
                <Switch disabled={!editable} size="large" label={['启用', '禁用']} />
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
            <LabelInput form={form} label='用户组标签' name='group_labels' editable={editable} />
            {editable && (
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
                header={op === 'create' ? "创建用户组" : editable ? "编辑用户组" : "用户组详情"}
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
                {editable ? userForm : groupView}
            </Drawer>
        </div>
    );
}

export default React.memo(GroupEditor);
