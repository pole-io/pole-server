import React from "react";
import { Avatar, Breadcrumb, Card, Descriptions, Form, Loading, Space, Tabs, Tag } from "tdesign-react";
import { useNavigate, useLocation } from 'react-router-dom';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { describeRoles, Role } from "services/role";
import LabelInput from "components/LabelInput";

const { BreadcrumbItem } = Breadcrumb;
const { DescriptionsItem } = Descriptions;
const { TabPanel } = Tabs;

interface IRoleDetailProps {

}

const RoleDetailTable: React.FC<IRoleDetailProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const urlParams = new URLSearchParams(window.location.search);
    const roleId = urlParams.get('id');
    const rolename = urlParams.get('name');

    const [viewState, setViewState] = React.useState<{
        loading: boolean;
        role: Role;
        fetchError: boolean;
    }>({ loading: false, role: {} as Role, fetchError: false });

    async function fetchData() {
        setViewState({ ...viewState, loading: true });
        const ret = await describeRoles({ id: roleId as string })
        if (ret?.content.length > 0) {
            const expectRole = ret.content[0];
            setViewState({
                loading: false, role: expectRole, fetchError: false
            });
        } else {
            setViewState({ ...viewState, loading: false, fetchError: true });
        }
    }

    React.useEffect(() => {
        if (roleId) {
            fetchData();
        }
    }, [roleId]);

    const roleDetail = (
        <Space direction="vertical" style={{ width: '100%' }}>
            <Descriptions column={2} tableLayout="auto" itemLayout="horizontal">
                <DescriptionsItem label="角色ID">
                    {viewState.role.id}
                </DescriptionsItem>
                <DescriptionsItem label="角色描述">
                    {viewState.role.comment}
                </DescriptionsItem>
                <DescriptionsItem label="角色名称">
                    {viewState.role.name}
                </DescriptionsItem>
                <DescriptionsItem label="角色来源">
                    {viewState.role.source}
                </DescriptionsItem>
                <DescriptionsItem label="创建时间">
                    {viewState.role.ctime}
                </DescriptionsItem>
                <DescriptionsItem label="更新时间">
                    {viewState.role.mtime}
                </DescriptionsItem>
                <DescriptionsItem label="角色标签">
                    <Space>
                        {viewState.role.metadata ? Object.entries(viewState.role.metadata).map(([key, value]) => {
                            return (
                                <Tag variant='outline'>{key}: {value}</Tag>
                            )
                        }
                        ) : <Tag>无</Tag>}
                    </Space>
                </DescriptionsItem>
            </Descriptions>
        </Space>
    )

    return (
        <>
            <Space direction="vertical" style={{ width: '100%' }}>
                <Breadcrumb maxItemWidth="200px">
                    <BreadcrumbItem onClick={() => {
                        navigate(-1);
                    }}>角色</BreadcrumbItem>
                    <BreadcrumbItem>{rolename}</BreadcrumbItem>
                </Breadcrumb>
                <Card
                    title={
                        <>
                            {`角色 ${rolename} 详情`}
                        </>
                    }
                    hoverShadow
                >
                    <Loading
                        indicator
                        loading={viewState.loading}
                        preventScrollThrough
                        showOverlay
                    >
                        {roleDetail}
                    </Loading>
                </Card>

                <Card>
                    <Tabs>
                        <TabPanel value="user" label="用户信息">
                            {viewState.role?.users?.map((item, index) => {
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
                        <TabPanel value="user-group" label="用户组信息">
                            {viewState.role?.user_groups?.map((item, index) => {
                                return (
                                    <Avatar
                                        shape="round"
                                        style={{ margin: 10 }}
                                        size="60px"
                                        onClick={() => {
                                            navigate(`/auth/principals/groupdetail?name=${item.name}&id=${item.id}`);
                                        }}
                                    >{item.name}</Avatar>
                                )
                            })
                            }
                        </TabPanel>
                        <TabPanel value="permission" label="权限信息">
                        </TabPanel>
                    </Tabs>
                </Card>
            </Space>
        </>
    )
}

export default React.memo(RoleDetailTable);