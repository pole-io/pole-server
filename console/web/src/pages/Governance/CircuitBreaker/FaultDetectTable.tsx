import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Tabs, Popconfirm, Empty } from 'tdesign-react';
import { DeleteIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';
import { useNavigate } from 'react-router-dom';

import { Op } from 'services/types';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import Search from 'components/Search';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';
import { FaultDetectRule } from 'services/faultdetect';
import FaultDetectEditor from './FaultDetectEditor';
import { clearFaultDetect, editorFaultDetect, listFaultDetects, listFaultDetectVersions, removeFaultDetects, removeFaultDetectVersion, rollbackFaultDetectVersion, selectFaultDetect } from 'modules/governance/faultdetect';
import RuleTabs from '../RuleRelease/RuleTabs';
import SubscribeTable from 'components/SubscribeTable';
import RuleDetailDrawer from '../RuleRelease/RuleDetailDrawer';

interface IFaultDetectTableProps {

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
        colKey: 'caller',
        title: '被探测服务',
        ellipsis: true,
        cell: ({ row }) => (
            <div className={style.serviceCell}>命名空间: {row.targetService?.namespace || '-'}<br />服务: {row.targetService?.service || '-'}</div>
        ),
    },
    {
        colKey: 'port',
        title: '探测参数',
        cell: ({ row }) => (
            <div className={style.serviceCell}>
                <div>协议: {row.protocol || '-'}</div>
                <div>端口: {row.port || '-'} / 间隔: {row.interval || '-'}s / 超时: {row.timeout || '-'}s</div>
            </div>
        ),
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

const FaultDetectTable: React.FC<IFaultDetectTableProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    const detectState = useAppSelector(selectFaultDetect);
    const { datas, total, page, limit, loading } = detectState;
    const { versions, versionTotal, versionPage, versionLimit, versionLoading } = detectState;
    const { subscribers } = detectState;

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
                dispatch(editorFaultDetect({ ...row } as FaultDetectRule));
                break;
            case 'create':
                setEditorState(pre => ({ ...pre, visible: true, mode: op }));
                return;
            case 'delete':
                dispatch(removeFaultDetects({ ids: [row.id] }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification("请求成功", "删除主动探测规则成功");
                            refreshData(1, limit);
                            setEditorState(pre => ({ ...pre, visible: false }));
                        } else {
                            openErrNotification("请求失败", `删除主动探测规则失败: ${res.payload as string}`);
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
        dispatch(listFaultDetectVersions({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                id: editorState.data?.id as string,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询主动探测规则版本列表错误: ${res.payload as string}`);
            }
        });
    }

    const operateRelease = async (op: Op, row: TableRowData) => {
        switch (op) {
            case 'delete':
                dispatch(removeFaultDetectVersion({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "删除主动探测规则版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `删除主动探测规则版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
            case 'rollback':
                dispatch(rollbackFaultDetectVersion({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "回滚主动探测规则版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `回滚主动探测规则版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
        }
    }

    // 模拟远程请求
    const refreshData = (page = 1, limit = 10, query = '') => {
        dispatch(listFaultDetects({
            param: {
                limit,
                offset: (page - 1) * limit,
                brief: true,
                name: query,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `获取主动探测规则列表失败, ${res?.payload as string}`);
            }
        })
    }

    React.useEffect(() => {
        refreshData();
        return () => {
            dispatch(clearFaultDetect());
        }
    }, []);

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
                    resource_type={PolicySourceType.FaultDetectRules}
                    resource_id={editorState.data?.id}
                    resource_name={`faultdetect/${editorState.data?.name}`}
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
                title={editorState.mode === 'create' ? '新建主动探测规则' : editorState.data?.name || '主动探测详情'}
                subtitle="主动探测"
                size="min(1560px, calc(100vw - 40px))"
                onClose={() => setEditorState(pre => ({ ...pre, visible: false }))}
            >
                <RuleTabs
                    op={editorState.mode}
                    onVersionView={() => {
                        refreshVersions(1, 10)
                    }}
                    view={
                        <>
                            <FaultDetectEditor
                                op={editorState.mode}
                                refresh={(close: boolean) => {
                                    if (close) {
                                        setEditorState(pre => ({ ...pre, mode: 'view', visible: false }));
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

export default React.memo(FaultDetectTable);
