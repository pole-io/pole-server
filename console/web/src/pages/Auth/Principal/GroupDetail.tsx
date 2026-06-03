import LabelInput from "components/LabelInput";
import React from "react";
import { Card, Form, Input, Link, Loading, Popup, Space, Table, Tag, TableProps, Collapse, Row, Col, Breadcrumb, Descriptions, Avatar, Tabs } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import { useNavigate, useLocation } from 'react-router-dom';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { enableUserToken, resetUserToken } from "modules/user/users";
import { describeUsers, describeUserToken, User } from "services/users";
import { openErrNotification, openInfoNotification } from "utils/notifition";
import { describeUserGroupDetail, describeUserGroupToken, UserGroup } from "services/user_group";

const { FormItem } = Form;
const { BreadcrumbItem } = Breadcrumb;
const { DescriptionsItem } = Descriptions;
const { TabPanel } = Tabs;

interface IGroupDetailProps {

}

const GroupDetailTable: React.FC<IGroupDetailProps> = ({ }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('id');
    const username = urlParams.get('name');

    const [viewState, setViewState] = React.useState<{
        loading: boolean;
        group: UserGroup;
        data: TableProps['data'];
        fetchError: boolean;
    }>({ loading: false, group: {} as UserGroup, data: [], fetchError: false });

    async function fetchData() {
        setViewState({ ...viewState, loading: true, group: {} as UserGroup, fetchError: false });
        const tokenRet = await describeUserGroupToken({ id: userId as string })
        if (tokenRet?.userGroup) {
            const ret = await describeUserGroupDetail({ id: userId as string })
            if (ret?.userGroup) {
                const expectGroup = ret.userGroup;
                expectGroup.auth_token = tokenRet.userGroup.auth_token;
                setViewState({
                    loading: false, group: expectGroup, data: [
                        { token: expectGroup.auth_token, status: expectGroup.token_enable }
                    ], fetchError: false
                });
                form.setFieldsValue({
                    id: expectGroup.id,
                    name: expectGroup.name,
                    comment: expectGroup.comment,
                    source: expectGroup.source,
                    user_labels: expectGroup.metadata ? Object.entries(expectGroup.metadata).map(([key, value]) => ({ key, value })) : [],
                });
            } else {
                setViewState({ ...viewState, loading: false, fetchError: true });
            }
        } else {
            setViewState({ ...viewState, loading: false, fetchError: true });
        }
    }

    React.useEffect(() => {
        fetchData();
    }, [userId]);

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
                <Descriptions>
                    <DescriptionsItem label={'ID'}>
                        {viewState.group?.id}
                    </DescriptionsItem>
                    <DescriptionsItem label={'备注'}>
                        {viewState.group?.comment}
                    </DescriptionsItem>
                    <DescriptionsItem label={'组名'}>
                        {viewState.group?.name}
                    </DescriptionsItem>
                    <DescriptionsItem label={'来源'}>
                        {viewState.group?.source}
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
                                    const enabled = viewState.group.token_enable
                                    return (
                                        <Space>
                                            <Link theme="primary" onClick={() => {
                                                dispatch(resetUserToken({ id: userId as string }))
                                                    .then((res) => {
                                                        if (res.meta.requestStatus !== 'fulfilled') {
                                                            openErrNotification('请求错误', "资源访问凭据重置失败");
                                                        } else {
                                                            openInfoNotification('请求成功', "资源访问凭据重置成功");
                                                            // 由于底层缓存设计的问题，这里需要延迟1s
                                                            setViewState({ ...viewState, loading: true })
                                                            setTimeout(() => fetchData(), 1000)
                                                        }
                                                    });
                                            }}>重置</Link>
                                            <Link theme={enabled ? 'danger' : 'success'} onClick={() => {
                                                dispatch(enableUserToken({ id: userId as string, token_enable: !enabled }))
                                                    .then((res) => {
                                                        if (res.meta.requestStatus !== 'fulfilled') {
                                                            openErrNotification('请求错误', `资源访问凭据${enabled ? '禁用' : '启用'}失败`);
                                                        } else {
                                                            openInfoNotification('请求成功', `资源访问凭据${enabled ? '禁用' : '启用'}成功`);
                                                            setViewState({ ...viewState, loading: true })
                                                            setTimeout(() => fetchData(), 1000)
                                                        }
                                                    });;
                                            }}>{enabled ? '禁用' : '启用'}</Link>
                                        </Space>
                                    )
                                }
                            },
                        ]}
                        data={viewState.data}
                    />
                </FormItem>
                <FormItem label="标签">
                    {viewState.group?.metadata && Object.keys(viewState.group.metadata).length > 0 && (
                        Object.entries(viewState.group.metadata).map(([key, value]) => {
                            return <Tag>{`${key}: ${value}`}</Tag>
                        })
                    )}
                </FormItem>
            </Space>
        </Form>
    )

    return (
        <>
            <Space direction="vertical" style={{ width: '100%' }}>
                <Breadcrumb maxItemWidth="200px">
                    <BreadcrumbItem onClick={() => {
                        navigate(-1);
                    }}>用户组</BreadcrumbItem>
                    <BreadcrumbItem>{username}</BreadcrumbItem>
                </Breadcrumb>
                <Card
                    title={`用户组 ${username} 详情`}
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
                        <TabPanel value="user" label="用户信息">
                            {viewState.group?.relation?.users?.map((item, index) => {
                                return (
                                    <Avatar
                                        shape="round"
                                        style={{ margin: 10 }}
                                        size="60px"
                                        onClick={() => {
                                            navigate(`/auth/principals/userdetail?name=${item.name}&id=${item.id}`);
                                        }}
                                    >{item.name}</Avatar>
                                )
                            })
                            }
                        </TabPanel>
                        <TabPanel value="role" label="角色信息">
                        </TabPanel>
                        <TabPanel value="permission" label="权限信息">
                        </TabPanel>
                    </Tabs>
                </Card>
            </Space>
        </>
    )
}

export default React.memo(GroupDetailTable);