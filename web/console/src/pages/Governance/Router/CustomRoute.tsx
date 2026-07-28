import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Popconfirm, Empty } from 'components/Fluent';
import { DeleteIcon, RefreshIcon, CreditcardIcon } from 'components/Fluent/icons';

import { Op } from 'services/types';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { CustomRouteView, normalizeRoutingConfigForEditor } from 'services/router';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import Search from 'components/Search';
import CustomRouteEditor from './CustomRouteEditor';
import AuthorizeInput from 'components/Authorize';
import { cleanCustomRoutePage, cleanCustomRouteVersions, editorCustomRoute, listCustomRoutes, listCustomRouteVersions, removeCustomRoutes, resetCustomRoute, selectCustomRoute } from 'modules/governance/route';
import { PolicySourceType } from 'services/auth_policy';
import RuleTabs from '../RuleRelease/RuleTabs';
import SubscribeTable from 'components/SubscribeTable';
import RuleDetailDrawer, { WIDE_RULE_DETAIL_DRAWER_SIZE } from '../RuleRelease/RuleDetailDrawer';

interface ICustomRouteProps {

}

const columns = (handleOpRule: (row: TableRowData, op: Op) => void, redirect: (row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        width: '20px',
        cell: ({ row }) => <Link
            theme="primary"
            onClick={() => { redirect(row) }}
        >{row.name}</Link>,
    },
    {
        colKey: 'caller',
        title: '主调',
        width: '20px',
        ellipsis: true,
        cell: ({ row: { routing_config } }: TableRowData) => {
            const fallbackRule = routing_config?.rules?.[0];
            const config = normalizeRoutingConfigForEditor(routing_config);
            const source = config?.caller || config?.rules?.[0]?.sources?.[0] || fallbackRule?.sources?.[0];
            return <div className={style.serviceCell}>命名空间: {source?.namespace || '-'}<br />服务: {source?.service || '-'}</div>;
        },
    },
    {
        colKey: 'callee',
        title: '被调',
        width: '20px',
        ellipsis: true,
        cell: ({ row: { routing_config } }: TableRowData) => {
            const fallbackRule = routing_config?.rules?.[0];
            const config = normalizeRoutingConfigForEditor(routing_config);
            const destination = config?.callee || config?.rules?.[0]?.destinations?.[0] || fallbackRule?.destinations?.[0];
            return <div className={style.serviceCell}>命名空间: {destination?.namespace || '-'}<br />服务: {destination?.service || '-'}</div>;
        },
    },
    {
        colKey: 'action',
        title: '操作',
        width: '20px',
        cell: ({ row }) => {
            return (
                <Space>
                    <Tooltip content={row.editable === false ? '无权限操作' : '授权'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => handleOpRule(row, 'authorize')}>
                            <CreditcardIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
                        <Popconfirm
                            content="确认删除吗"
                            destroyOnClose
                            placement="top"
                            showArrow
                            theme="default"
                            onConfirm={() => {
                                handleOpRule(row, 'delete');
                            }}
                        >
                            <Button shape="square" variant="text" disabled={row.deleteable === false}>
                                <DeleteIcon />
                            </Button>
                        </Popconfirm>
                    </Tooltip>
                </Space>
            )
        },
    },
]

const CustomRoute: React.FC<ICustomRouteProps> = ({ }) => {
    const dispatch = useAppDispatch();

    const customRouteState = useAppSelector(selectCustomRoute);
    const { datas, loading, total, page, limit } = customRouteState;
    const { versions, versionTotal, versionPage, versionLimit, versionLoading } = customRouteState;
    const { subscribers } = customRouteState;

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, mode: '', data: undefined, authorizeVisible: false });

    const handleOpRule = (row: TableRowData, op: Op) => {
        switch (op) {
            case 'view':
                dispatch(editorCustomRoute({ ...row as CustomRouteView }));
                setEditorState(pre => ({ ...pre, visible: true, mode: op, data: { ...row } }));
                break;
            case 'create':
                dispatch(resetCustomRoute());
                setEditorState(pre => ({ ...pre, visible: true, mode: op }));
                return;
            case 'delete':
                dispatch(removeCustomRoutes({ ids: [row.id] }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification("请求成功", "删除自定义路由规则成功");
                            refreshTable();
                            setEditorState(pre => ({ ...pre, visible: false }));
                        } else {
                            openErrNotification("请求失败", `删除自定义路由规则失败: ${res.payload as string}`);
                        }
                    });
                return;
            case 'authorize':
                setEditorState(pre => ({ ...pre, authorizeVisible: true, data: { ...row } }));
                return;
        }
    }

    React.useEffect(() => {
        refreshTable();
        return () => {
            dispatch(cleanCustomRoutePage());
        }
    }, []);

    React.useEffect(() => {
        if (editorState.data?.id) {
            refreshVersions();
        }

        return () => {
            dispatch(cleanCustomRouteVersions());
        }
    }, [editorState.data])

    const refreshTable = (page = 1, limit = 10, query = '') => {
        dispatch(listCustomRoutes({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                route_type: 'RulePolicy',
                name: query,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询路由规则列表错误: ${res.payload as string}`);
            }
        });
    }

    const refreshVersions = (page = 1, limit = 10) => {
        dispatch(listCustomRouteVersions({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                id: editorState.data?.id as string,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询路由版本列表错误: ${res.payload as string}`);
            }
        });
    }

    const operateRelease = async (op: Op, row: TableRowData) => {
    }

    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button onClick={(v) => {
                                handleOpRule({}, 'create')
                            }}>新建</Button>
                        </Col>
                    </Row>
                </Col>
                <Col>
                    <Space>
                        <Search
                            onChange={(value: string) => {
                                refreshTable(1, limit, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshTable(1, limit)} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editorState.authorizeVisible && (
                <AuthorizeInput
                    resource_type={PolicySourceType.RouteRules}
                    resource_id={editorState.data?.id}
                    resource_name={`custom_route/${editorState.data?.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
            <Table
                data={datas}
                columns={columns(handleOpRule, (row: TableRowData) => {
                    handleOpRule(row, 'view');
                })}
                loading={loading}
                rowKey="id"
                size={"large"}
                tableLayout={'auto'}
                cellEmptyContent={'-'}
                pagination={{
                    current: page,
                    pageSize: limit,
                    total: total,
                    showJumper: true,
                    onChange(pageInfo) {
                        refreshTable(pageInfo.current, pageInfo.pageSize);
                    },
                }}
                onPageChange={(pageInfo) => {
                    refreshTable(pageInfo.current, pageInfo.pageSize);
                }}
            />
        </>
    )

    return (
        <div className={style.ruleWorkspace}>
            <section className={style.ruleListPane}>
                {table}
            </section>
            <RuleDetailDrawer
                visible={editorState.visible}
                title={editorState.mode === 'create' ? '新建自定义路由' : editorState.data?.name || '自定义路由详情'}
                subtitle="自定义路由"
                size={WIDE_RULE_DETAIL_DRAWER_SIZE}
                onClose={() => setEditorState(pre => ({ ...pre, visible: false }))}
            >
                <RuleTabs
                    op={editorState.mode}
                    view={editorState.mode !== 'create' && !editorState.data ? (
                        <Empty title="选择路由规则查看详情" />
                    ) : (
                        <CustomRouteEditor
                            op={editorState.mode}
                            editable={editorState.data?.editable ?? true}
                            refresh={(close: boolean) => {
                                if (close) {
                                    setEditorState(pre => ({ ...pre, visible: false }));
                                } else {
                                    refreshTable(1, limit);
                                }
                            }} />
                    )}
                    versions={{
                        datas: versions,
                        action: operateRelease,
                        editable: editorState.data?.editable ?? true,
                        deleteable: editorState.data?.deleteable ?? true,
                        loading: versionLoading,
                        pagination: {
                            defaultCurrent: versionPage,
                            defaultPageSize: versionLimit,
                            total: versionTotal,
                            showJumper: false,
                            onChange(pageInfo) {
                                refreshVersions(pageInfo.current, pageInfo.pageSize);
                            },
                        },
                        onPageChange: (page) => {
                            refreshVersions(page.current, page.pageSize);
                        }
                    }}
                    subscribe={
                        <>
                            <div style={{ marginLeft: 20, marginTop: 20 }}>
                                <SubscribeTable
                                    title={`${editorState.data?.name}`}
                                    editable={editorState.data?.editable ?? true}
                                    deleteable={editorState.data?.deleteable ?? true}
                                    subscribers={subscribers || []}
                                />
                            </div>
                        </>
                    }
                />
            </RuleDetailDrawer>
        </div>
    )
}

export default React.memo(CustomRoute);
