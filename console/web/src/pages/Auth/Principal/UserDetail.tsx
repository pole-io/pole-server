import React from "react";
import { Card, Form, Input, Link, Loading, Popup, Space, Table, Tag, TableProps, Breadcrumb, Descriptions, Tabs } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import { DiscountIcon } from "tdesign-icons-react";
import { useNavigate } from 'react-router-dom';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { enableUserToken, listOneUser, resetUserToken, selectUser } from "modules/user/users";
import { describeAuthPolicies, PolicyRule } from "services/auth_policy";
import { User } from "services/users";
import { openErrNotification, openInfoNotification } from "utils/notifition";

const { FormItem } = Form;
const { BreadcrumbItem } = Breadcrumb;
const { DescriptionsItem } = Descriptions;
const { TabPanel } = Tabs;

interface IUserDetailProps {

}

const UserDetailTable: React.FC<IUserDetailProps> = ({ }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('id');
    const username = urlParams.get('name');

    const userState = useAppSelector(selectUser);
    const { editUser, viewUser } = userState;

    const [viewState, setViewState] = React.useState<{
        loading: boolean;
        user: User;
        data: TableProps['data'];
        policies: PolicyRule[];
        policyTotal: number;
        policyLoading: boolean;
        fetchError: boolean;
    }>({ loading: false, user: {} as User, data: [], policies: [], policyTotal: 0, policyLoading: false, fetchError: false });

    React.useEffect(() => {
        if (userId && userId.length > 0) {
            loadUser();
        }
    }, [userId]);

    const loadUser = () => {
        setViewState(prev => ({ ...prev, loading: true }));
        dispatch(listOneUser({ id: userId as string })).then((res) => {
            setViewState(prev => ({ ...prev, loading: false }));
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('请求错误', `获取用户详情失败, ${res?.payload as string}`);
                return;
            }

            const { viewUser } = res.payload as { viewUser: User };

            if (viewUser) {
                form.setFieldsValue({
                    id: viewUser.id,
                    name: viewUser.name,
                    comment: viewUser.comment,
                    source: viewUser.source,
                    email: viewUser.email,
                    mobile: viewUser.mobile,
                    user_labels: viewUser.metadata ? Object.entries(viewUser.metadata).map(([key, value]) => ({ key, value })) : [],
                });
                loadUserPolicies(viewUser.id || (userId as string));
            }
        })
    }

    const loadUserPolicies = async (id: string) => {
        setViewState(prev => ({ ...prev, policyLoading: true }));
        try {
            const ret = await describeAuthPolicies({
                principal_id: id,
                principal_type: 1,
                offset: 0,
                limit: 100,
            });
            setViewState(prev => ({
                ...prev,
                policies: ret.content,
                policyTotal: ret.totalCount,
                policyLoading: false,
            }));
        } catch (error) {
            setViewState(prev => ({ ...prev, policyLoading: false }));
            openErrNotification('请求错误', `获取关联策略失败, ${(error as Error).message}`);
        }
    }

    const handleChangePassword = () => {

    }

    const onSubmit: FormProps['onSubmit'] = async (e) => { }

    const userForm = (
        <Form
            form={form}
            layout="vertical"
            labelWidth={120}
            labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <Space direction="vertical" style={{ width: '100%' }}>
                <Descriptions column={2} size="small">
                    <DescriptionsItem label="用户ID">
                        {viewUser?.id}
                    </DescriptionsItem>
                    <DescriptionsItem label="备注">
                        {viewUser?.comment}
                    </DescriptionsItem>
                    <DescriptionsItem label="用户名">
                        {viewUser?.name}
                    </DescriptionsItem>
                    <DescriptionsItem label="来源">
                        {viewUser?.source}
                    </DescriptionsItem>
                    <DescriptionsItem label="邮箱">
                        {viewUser?.email}
                    </DescriptionsItem>
                    <DescriptionsItem label="手机号">
                        {viewUser?.mobile}
                    </DescriptionsItem>
                </Descriptions>
                <FormItem label="资源访问凭据" name="token_enable" shouldUpdate={true}>
                    <Table
                        rowKey="id"
                        size={"large"}
                        tableLayout={'auto'}
                        cellEmptyContent={'-'}
                        columns={[
                            {
                                colKey: 'token',
                                title: 'Token',
                                cell: ({ row: { token } }) => {
                                    return (
                                        <Popup content="已复制" trigger="click">
                                            <Input size="large" borderless={true} type="password" value={token} readonly onClick={() => {
                                                navigator.clipboard.writeText(token as string)
                                            }} />
                                        </Popup>
                                    )
                                }
                            },
                            {
                                colKey: 'status',
                                title: '状态',
                                cell: ({ row: { status } }) => {
                                    return (
                                        <Tag theme={status ? 'success' : 'danger'}>{status ? '启用' : '禁用'}</Tag>
                                    )
                                }
                            },
                            {
                                colKey: 'op',
                                title: '操作',
                                cell: () => {
                                    const enabled = viewUser?.token_enable
                                    return (
                                        <Space>
                                            <Link theme="primary" onClick={() => {
                                                dispatch(resetUserToken({ id: userId as string }))
                                                    .then((res) => {
                                                        if (res.meta.requestStatus === 'rejected') {
                                                            openErrNotification('请求错误', "资源访问凭据重置失败");
                                                        } else {
                                                            openInfoNotification('请求成功', "资源访问凭据重置成功");
                                                            // 由于底层缓存设计的问题，这里需要延迟1s
                                                            setViewState({ ...viewState, loading: true })
                                                            setTimeout(() => loadUser(), 1000)
                                                        }
                                                    });
                                            }}>重置</Link>
                                            <Link theme={enabled ? 'danger' : 'success'} onClick={() => {
                                                dispatch(enableUserToken({ id: userId as string, token_enable: !enabled }))
                                                    .then((res) => {
                                                        if (res.meta.requestStatus === 'rejected') {
                                                            openErrNotification('请求错误', `资源访问凭据${enabled ? '禁用' : '启用'}失败`);
                                                        } else {
                                                            openInfoNotification('请求成功', `资源访问凭据${enabled ? '禁用' : '启用'}成功`);
                                                            setViewState({ ...viewState, loading: true })
                                                            setTimeout(() => loadUser(), 1000)
                                                        }
                                                    });;
                                            }}>{enabled ? '禁用' : '启用'}</Link>
                                        </Space>
                                    )
                                }
                            },
                        ]}
                        data={viewUser ? [{ token: viewUser.auth_token, status: viewUser.token_enable }] : []}
                    />
                </FormItem>
                <FormItem label="标签">
                    <Space>
                        {viewUser?.metadata && Object.keys(viewUser.metadata).length > 0 && (
                            Object.entries(viewUser.metadata).map(([key, value]) => {
                                return <Tag>{`${key}: ${value}`}</Tag>
                            })
                        )}
                    </Space>
                </FormItem>
            </Space>
        </Form>
    )

    const policyColumns: TableProps['columns'] = [
        {
            colKey: 'name',
            title: '策略名称',
            cell: ({ row }) => (
                <Link
                    theme="primary"
                    onClick={() => navigate(`/auth/policies/detail?id=${row.id}&name=${row.name}`)}
                >
                    {row.name}
                </Link>
            ),
        },
        {
            colKey: 'action',
            title: '行为',
            width: 120,
            cell: ({ row }) => row.action || '-',
        },
        {
            colKey: 'default_strategy',
            title: '默认策略',
            width: 120,
            cell: ({ row }) => (
                <Tag theme={row.default_strategy ? 'success' : 'default'} variant="outline">
                    {row.default_strategy ? '是' : '否'}
                </Tag>
            ),
        },
        {
            colKey: 'comment',
            title: '描述',
            ellipsis: true,
            cell: ({ row }) => row.comment || '-',
        },
        {
            colKey: 'time',
            title: '操作时间',
            width: 210,
            cell: ({ row }) => (
                <span>
                    修改: {row.mtime || '-'}
                    <br />
                    创建: {row.ctime || '-'}
                </span>
            ),
        },
    ];

    const policyTable = (
        <Table
            rowKey="id"
            size="medium"
            tableLayout="auto"
            cellEmptyContent="-"
            loading={viewState.policyLoading}
            columns={policyColumns}
            data={viewState.policies}
            pagination={{
                pageSize: 100,
                total: viewState.policyTotal,
                current: 1,
                showPageSize: false,
                showJumper: false,
            }}
        />
    );

    return (
        <>
            <Space direction="vertical" style={{ width: '100%' }}>
                <Breadcrumb maxItemWidth="200px">
                    <BreadcrumbItem onClick={() => {
                        navigate(-1);
                    }}>用户</BreadcrumbItem>
                    <BreadcrumbItem>{username}</BreadcrumbItem>
                </Breadcrumb>
                <Card
                    title={
                        <>
                            {`用户 ${username} 详情`}
                            <Tag icon={<DiscountIcon />} style={{ marginLeft: 10 }} theme="default">
                                {viewUser?.user_type === 'main' ? '管理员' : '子用户'}
                            </Tag>
                        </>
                    }
                    actions={
                        <Link theme="primary" onClick={handleChangePassword} style={{ cursor: 'pointer' }}>
                            修改密码
                        </Link>
                    }
                    hoverShadow
                >
                    <Loading
                        indicator
                        loading={viewState.loading}
                        preventScrollThrough
                        showOverlay
                    >
                        {userForm}
                    </Loading>
                </Card>

                <Card>
                    <Tabs>
                        {viewUser?.user_type !== 'main' && (
                            <>
                                <TabPanel value="user-group" label="用户组信息">
                                </TabPanel>
                                <TabPanel value="role" label="角色信息">
                                </TabPanel>
                            </>
                        )}
                        <TabPanel value="permission" label="权限信息">
                            {policyTable}
                        </TabPanel>
                    </Tabs>
                </Card>

            </Space>
        </>
    )
}

export default React.memo(UserDetailTable);
