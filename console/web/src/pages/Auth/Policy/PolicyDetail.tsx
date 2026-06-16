import React from "react";
import { Avatar, Breadcrumb, Card, Descriptions, Form, List, Loading, PrimaryTableProps, Space, Table, TableRowData, Tabs, Tag, Tree } from "tdesign-react";
import { useNavigate, useParams } from "react-router-dom";

import style from './index.module.less';
import { useAppDispatch, useAppSelector } from "modules/store";
import { describeAuthPolicyDetail, PolicyResource, PolicyRule } from "services/auth_policy";
import LabelInput from "components/LabelInput";
import { openErrNotification } from "utils/notifition";
import ErrorPage from "components/ErrorPage";

const { BreadcrumbItem } = Breadcrumb;
const { TabPanel } = Tabs;
const { DescriptionsItem } = Descriptions;
const { ListItem, ListItemMeta } = List;

const resourceTreeData = [
    {
        label: '命名空间',
        value: 'namespaces',
    },
    {
        label: '服务',
        value: 'services',
    },
    {
        label: '配置组',
        value: 'config_groups',
    },
    {
        label: '路由规则',
        value: 'route_rules',
    },
    {
        label: '泳道规则',
        value: 'lane_rules',
    },
    {
        label: '熔断规则',
        value: 'circuitbreaker_rules',
    },
    {
        label: '主动探测规则',
        value: 'faultdetect_rules',
    },
    {
        label: '限流规则',
        value: 'ratelimit_rules',
    },
    {
        label: '调用鉴权规则',
        value: 'security_rules',
    },
    {
        label: '流量镜像规则',
        value: 'mirror_rules',
    },
    {
        label: '流量 Mock 规则',
        value: 'mock_rules',
    },
    {
        label: '用户',
        value: 'users',
    },
    {
        label: '用户组',
        value: 'user_groups',
    },
    {
        label: '资源鉴权规则',
        value: 'auth_policies',
    },
    {
        label: '角色',
        value: 'roles',
    }
]

const resourceColumns = (onClick: (row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: '资源ID',
        cell: ({ row: { id } }) => <span>{id === '*' ? '全部（包括新增）' : id}</span>,
    },
    {
        colKey: 'name',
        title: '资源名称',
        cell: ({ row: { name } }) => <span>{name === '*' ? '全部（包括新增）' : name}</span>,
    }
];

interface IPolicyDetailProps {

}

const PolicyDetailTable: React.FC<IPolicyDetailProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const [form] = Form.useForm();
    const navigate = useNavigate();
    const urlParams = new URLSearchParams(window.location.search);
    const policyId = urlParams.get('id');
    const policyname = urlParams.get('name');

    const [viewState, setViewState] = React.useState<{
        loading: boolean;
        rule: PolicyRule;
        fetchError: boolean;
        allResources: TableRowData[];
        resources: TableRowData[];
        activePrincipal: string;
        activeResourceType: string;
    }>({
        loading: false,
        rule: {} as PolicyRule,
        fetchError: false,
        allResources: [],
        resources: [],
        activePrincipal: '',
        activeResourceType: 'namespaces'
    });

    async function fetchData() {
        setViewState(prevState => ({ ...prevState, loading: true }));
        try {
            const ret = await describeAuthPolicyDetail({ id: policyId as string })
            if (ret?.strategy) {
                const expectRule = ret.strategy;
                const resources = [] as TableRowData[];
                if (expectRule.resources) {
                    Object.entries(expectRule.resources).forEach(([key, value]) => {
                        if (!value) return;
                        (value as PolicyResource[]).forEach((item) => {
                            resources.push({
                                id: item.id,
                                name: item.name,
                                type: key,
                            });
                        });
                    });
                }

                const activeTab = expectRule.principals?.users && expectRule.principals.users.length > 0 ? "user" :
                    expectRule.principals?.groups && expectRule.principals.groups.length > 0 ? "user-group" :
                        expectRule.principals?.roles && expectRule.principals.roles.length > 0 ? "role" : '';

                form.setFieldsValue({
                    policy_labels: expectRule.metadata ? Object.entries(expectRule.metadata).map(([key, value]) => ({ key, value })) : [],
                });
                setViewState(prevState => {
                    const filteredResources = prevState.activeResourceType ?
                        resources.filter(item => item.type === prevState.activeResourceType) :
                        resources;
                    return {
                        ...prevState,
                        loading: false,
                        rule: expectRule,
                        allResources: resources,
                        resources: filteredResources,
                        activePrincipal: activeTab,
                        fetchError: false
                    };
                });
            } else {
                setViewState(prevState => {
                    return { ...prevState, loading: false };
                });
            }
        } catch (error) {
            setViewState(prevState => {
                return { ...prevState, loading: false, fetchError: true };
            });
            openErrNotification("获取策略详情失败", error as string);
        }
    };

    React.useEffect(() => {
        if (policyId) {
            fetchData();
        }
    }, [policyId]);

    const handleClickResource = (row: TableRowData) => {
        // 如果需要，可以实现资源点击后的导航或其他功能
        console.log('资源详情:', row);
    }

    // 处理资源树点击事件
    const handleTreeNodeClick = (value: string) => {
        if (!value) return;
        // 保存所有原始资源数据
        // 根据选中的资源类型过滤表格数据
        const filteredResources = viewState.allResources.filter(item => item.type === value);
        // 更新状态（同时更新活动类型和过滤后的资源）
        setViewState(prevState => ({
            ...prevState,
            activeResourceType: value,
            resources: filteredResources,
        }));
    };

    const ruleDetail = (
        <>
            <Space direction="vertical" style={{ width: '100%' }}>
                <Breadcrumb maxItemWidth="200px">
                    <BreadcrumbItem onClick={() => {
                        navigate(-1);
                    }}>策略</BreadcrumbItem>
                    <BreadcrumbItem>{policyname}</BreadcrumbItem>
                </Breadcrumb>
                <Card
                    title={
                        <>
                            {`策略 ${policyname} 详情`}
                        </>
                    }
                    hoverShadow
                >
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <Descriptions column={2} tableLayout="auto" itemLayout="horizontal">
                            <DescriptionsItem label="ID">
                                {viewState.rule.id}
                            </DescriptionsItem>
                            <DescriptionsItem label="描述">
                                {viewState.rule.comment}
                            </DescriptionsItem>
                            <DescriptionsItem label="名称">
                                {viewState.rule.name}
                            </DescriptionsItem>
                            <DescriptionsItem label="来源">
                                {viewState.rule.source}
                            </DescriptionsItem>
                            <DescriptionsItem label="效果">
                                {viewState.rule.action === 'ALLOW' ? <Tag theme="success">允许</Tag> : <Tag theme="danger">拒绝</Tag>}
                            </DescriptionsItem>
                            <DescriptionsItem label="创建时间">
                                {viewState.rule.ctime}
                            </DescriptionsItem>
                            <DescriptionsItem label="更新时间">
                                {viewState.rule.mtime}
                            </DescriptionsItem>
                            <DescriptionsItem label="标签">
                                {viewState.rule?.metadata && Object.keys(viewState.rule.metadata).length > 0 && (
                                    Object.entries(viewState.rule.metadata).map(([key, value]) => {
                                        return <Tag>{`${key}: ${value}`}</Tag>
                                    })
                                )}
                            </DescriptionsItem>
                        </Descriptions>
                    </Space>
                </Card>
                <Card>
                    {viewState.activePrincipal ? (
                        <Tabs defaultValue={viewState.activePrincipal}>
                            {viewState.rule?.principals?.users && viewState.rule.principals.users.length > 0 && (
                                <TabPanel value="user" label="用户信息">
                                    {viewState.rule.principals.users.map((item, index) => (
                                        <Avatar
                                            shape="round"
                                            style={{ margin: 10 }}
                                            size="60px"
                                            key={index}
                                            onClick={() => {
                                                navigate(`/auth/principals/userdetail?name=${item.name}&id=${item.id}`);
                                            }}
                                        >
                                            {item.name}
                                        </Avatar>
                                    ))}
                                </TabPanel>
                            )}
                            {viewState.rule?.principals?.groups && viewState.rule.principals.groups.length > 0 && (
                                <TabPanel value="user-group" label="用户组信息">
                                    {viewState.rule.principals.groups.map((item, index) => (
                                        <Avatar
                                            shape="round"
                                            style={{ margin: 10 }}
                                            size="60px"
                                            key={index}
                                            onClick={() => {
                                                navigate(`/auth/principals/groupdetail?name=${item.name}&id=${item.id}`);
                                            }}
                                        >
                                            {item.name}
                                        </Avatar>
                                    ))}
                                </TabPanel>
                            )}
                            {viewState.rule?.principals?.roles && viewState.rule.principals.roles.length > 0 && (
                                <TabPanel value="role" label="角色信息">
                                    {viewState.rule.principals.roles.map((item, index) => (
                                        <Avatar
                                            shape="round"
                                            style={{ margin: 10 }}
                                            size="60px"
                                            key={index}
                                            onClick={() => {
                                                navigate(`/auth/principals/roledetail?name=${item.name}&id=${item.id}`);
                                            }}
                                        >
                                            {item.name}
                                        </Avatar>
                                    ))}
                                </TabPanel>
                            )}
                        </Tabs>
                    ) : (
                        <div style={{ textAlign: 'center', padding: '20px' }}>
                            暂无关联用户/用户组/角色数据
                        </div>
                    )}
                </Card>
                <Card>
                    <Tabs defaultValue={"resource"}>
                        <TabPanel value="resource" label="资源信息">
                            <Space>
                                <div className={style.treeContent}>
                                    <Tree
                                        data={resourceTreeData}
                                        activable
                                        hover
                                        lazy
                                        transition
                                        valueMode={"onlyLeaf"}
                                        actived={[viewState.activeResourceType]}
                                        onClick={({ node }) => handleTreeNodeClick(node.value as string)}
                                    />
                                </div>
                                <Table
                                    data={viewState.resources}
                                    columns={resourceColumns(handleClickResource)}
                                    rowKey="id"
                                    size={"large"}
                                    tableLayout={'fixed'}
                                    cellEmptyContent={'-'}
                                    stripe={true}
                                    pagination={{
                                        defaultCurrent: 1,
                                        defaultPageSize: 10,
                                        showJumper: true,
                                        total: viewState.resources.length,
                                    }}
                                />
                            </Space>
                        </TabPanel>
                        <TabPanel value='resource-labels' label='资源标签'>
                            <Table
                                data={viewState.rule?.resource_labels || []}
                                columns={[
                                    {
                                        colKey: 'key',
                                        title: '标签 Key',
                                        cell: ({ row }) => <span>{row.key}</span>,
                                    },
                                    {
                                        colKey: 'value',
                                        title: '标签值',
                                        cell: ({ row }) => <span>{row.value}</span>,
                                    },
                                    {
                                        colKey: 'compare_type',
                                        title: '匹配方式',
                                        cell: ({ row }) => <span>{row.compare_type}</span>,
                                    }
                                ]}
                                rowKey="key"
                                size={"large"}
                                tableLayout={'fixed'}
                                cellEmptyContent={'-'}
                                stripe={true}
                                pagination={{
                                    defaultCurrent: 1,
                                    defaultPageSize: 10,
                                    showJumper: true,
                                    total: viewState.rule?.resource_labels?.length,
                                }}
                            />
                        </TabPanel>
                        <TabPanel value="function" label="可访问接口">
                            <Space style={{ width: '100%' }} direction="vertical">
                                <List
                                    style={{ maxHeight: '300px' }}
                                    scroll={{ type: 'virtual', bufferSize: 10, threshold: 10 }}
                                    size="small"
                                >
                                    {viewState.rule?.functions?.map((item, index) => (
                                        <ListItem key={index}>
                                            <ListItemMeta description={item === '*' ? '全部接口（包括新增）' : item} />
                                        </ListItem>
                                    ))}
                                </List>
                            </Space>
                        </TabPanel>
                    </Tabs>
                </Card>
            </Space>
        </>
    )

    return (
        <>
            <Loading
                indicator
                loading={viewState.loading}
                preventScrollThrough
                showOverlay
            >
                {viewState.fetchError ? (
                    <ErrorPage code={500} />
                ) : (
                    ruleDetail
                )}
            </Loading>
        </>
    )
}

export default React.memo(PolicyDetailTable);
