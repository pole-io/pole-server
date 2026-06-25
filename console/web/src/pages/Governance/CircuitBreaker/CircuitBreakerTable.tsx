import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Tabs, Popconfirm, Empty } from 'tdesign-react';
import { AddIcon, DeleteIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';
import { useNavigate } from 'react-router-dom';

import { Op } from 'services/types';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import Search from 'components/Search';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { BreakLevelMap, BreakLevelType, CircuitBreakerRule } from 'services/circuitbreaker';
import CircuitBreakerEditor from './CircuitBreakerEditor';
import { cleanCircuitBreakerPage, editorCircuitBreaker, listCircuitBreakers, listCircuitBreakerVersions, removeCircuitBreakerRelease, removeCircuitBreakers, resetCircuitBreaker, rollbackCircuitBreakerRelease, selectCircuitBreaker } from 'modules/governance/circuitbreaker';
import SubscribeTable from 'components/SubscribeTable';
import RuleTabs from '../RuleRelease/RuleTabs';
import RuleDetailDrawer, { WIDE_RULE_DETAIL_DRAWER_SIZE } from '../RuleRelease/RuleDetailDrawer';

interface ICircuitBreakerTableProps {

}

const columns = (handleOpRule: (row: TableRowData, op: Op) => void, redirect: (row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }) => <Link
            theme="primary"
            onClick={() => { redirect(row) }}
        >{row.name}</Link>,
    },
    {
        colKey: 'level',
        title: '粒度',
        cell: ({ row }) => <Text>{BreakLevelMap[row.level as BreakLevelType]}</Text>,
    },
    {
        colKey: 'caller',
        title: '主调',
        ellipsis: true,
        cell: ({ row: { ruleMatcher } }: TableRowData) => (
            <div className={style.serviceCell}>命名空间: {ruleMatcher?.source?.namespace || '-'}<br />服务: {ruleMatcher?.source?.service || '-'}</div>
        ),
    },
    {
        colKey: 'callee',
        title: '被调',
        ellipsis: true,
        cell: ({ row: { ruleMatcher } }: TableRowData) => (
            <div className={style.serviceCell}>命名空间: {ruleMatcher?.destination?.namespace || '-'}<br />服务: {ruleMatcher?.destination?.service || '-'}</div>
        ),
    },
    {
        colKey: 'action',
        title: '操作',
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

const CircuitBreakerTable: React.FC<ICircuitBreakerTableProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    const breakerState = useAppSelector(selectCircuitBreaker);
    const { datas, total, page, limit, loading } = breakerState;
    const { versions, versionTotal, versionPage, versionLimit, versionLoading } = breakerState;
    const { subscribers } = breakerState;

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
                setEditorState(pre => ({ ...pre, visible: true, mode: op, data: { ...row } }));
                dispatch(editorCircuitBreaker({ ...row } as CircuitBreakerRule));
                break;
            case 'create':
                setEditorState(pre => ({ ...pre, visible: true, mode: op }));
                dispatch(resetCircuitBreaker());
                return;
            case 'delete':
                dispatch(removeCircuitBreakers({ ids: [row.id] }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification("请求成功", "删除熔断降级规则成功");
                            refreshData(1, limit);
                        } else {
                            openErrNotification("请求失败", `删除熔断降级规则失败: ${res.payload as string || '未知'}`);
                        }
                    });
                break;
            case 'authorize':
                setEditorState(pre => ({
                    ...pre,
                    authorizeVisible: true,
                    data: { ...row },
                }));
                return;
        }
    }

    const refreshVersions = (page = 1, limit = 10) => {
        dispatch(listCircuitBreakerVersions({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                rule_name: editorState.data?.name as string,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询熔断规则版本列表错误: ${res.payload as string}`);
            }
        });
    }

    const operateRelease = async (op: Op, row: TableRowData) => {
        switch (op) {
            case 'delete':
                dispatch(removeCircuitBreakerRelease({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "删除熔断规则版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `删除熔断规则版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
            case 'rollback':
                dispatch(rollbackCircuitBreakerRelease({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "回滚熔断规则版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `回滚熔断规则版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
        }
    }

    // 模拟远程请求
    const refreshData = (page = 1, limit = 10, query = '') => {
        dispatch(listCircuitBreakers({
            param: {
                limit,
                offset: (page - 1) * limit,
                ...(query && { name: query }),
                brief: true
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `获取熔断降级规则列表失败, ${res?.payload as string}`);
            }
        });
    }

    React.useEffect(() => {
        refreshData();

        return () => {
            dispatch(cleanCircuitBreakerPage());
        }
    }, []);

    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button icon={<AddIcon />} onClick={(v) => {
                                handleOpRule({}, 'create')
                            }}>新建熔断规则</Button>
                        </Col>
                    </Row>
                </Col>
                <Col>
                    <Space>
                        <Search
                            onChange={(value: string) => {
                                refreshData(1, limit, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshData(1, limit)} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editorState.authorizeVisible && (
                <AuthorizeInput
                    resource_type={PolicySourceType.CircuitBreakerRules}
                    resource_id={editorState.data?.id}
                    resource_name={`circuitbreaker/${editorState.data?.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
            <Table
                data={datas || []}
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
                        refreshData(pageInfo.current, pageInfo.pageSize);
                    },
                }}
                onPageChange={(pageInfo) => {
                    refreshData(pageInfo.current, pageInfo.pageSize);
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
                title={editorState.mode === 'create' ? '新建熔断规则' : editorState.data?.name || '熔断规则详情'}
                subtitle="故障熔断"
                size={WIDE_RULE_DETAIL_DRAWER_SIZE}
                onClose={() => setEditorState(pre => ({ ...pre, visible: false }))}
            >
                <RuleTabs
                    op={editorState.mode}
                    onVersionView={() => {
                        refreshVersions(1, 10)
                    }}
                    view={
                        <>
                            <CircuitBreakerEditor
                                op={editorState.mode}
                                refresh={(close: boolean) => {
                                    if (close) {
                                        setEditorState(pre => ({ ...pre, visible: false }));
                                    }
                                    refreshData(1, limit);
                                }}
                            />
                        </>
                    }
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

export default React.memo(CircuitBreakerTable);
