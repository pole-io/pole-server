import React from 'react';
import { Button, Table, Space, Row, Col, Tooltip, PrimaryTableProps, TableRowData, Popconfirm, Tabs, Link, Empty } from 'tdesign-react';
import { CreditcardIcon, DeleteIcon, RefreshIcon } from 'tdesign-icons-react';

import LossLessEditor, { } from './LossLessEditor';
import { useAppDispatch, useAppSelector } from 'modules/store';
import style from './index.module.less';
import Search from 'components/Search';
import { Op } from 'services/types';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import RuleTabs from '../RuleRelease/RuleTabs';
import SubscribeTable from 'components/SubscribeTable';
import { editorLosslessRule, listLossLessRules, listLosslessRuleVersions, removeLosslessRule, removeLosslessVersion, rollbackLosslessVersion, selectLosslessRule } from 'modules/governance/lossless';
import { LossLessRuleView } from 'services/lossless';
import RuleDetailDrawer from '../RuleRelease/RuleDetailDrawer';

export interface ILossLessTableProps {

}

const columns = (handleOpRule: (row: TableRowData, op: Op) => void, redirect: (row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '规则名称',
        cell: ({ row }) => (
            <Link
                theme="primary"
                onClick={() => {
                    redirect(row);
                }}
            >
                {row.namespace}/{row.service}
            </Link>
        ),
    },
    {
        colKey: 'namespace',
        title: '命名空间',
    },
    {
        colKey: 'service',
        title: '服务名称',
    },
    {
        colKey: 'operation',
        title: '操作',
        fixed: 'right',
        cell: ({ row }) => (
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
                        onConfirm={() => { handleOpRule(row, 'delete') }}
                    >
                        <Button shape="square" variant="text" disabled={row.deleteable === false}>
                            <DeleteIcon />
                        </Button>
                    </Popconfirm>
                </Tooltip>
            </Space>
        )
    }
]

const LossLessTable: React.FC<ILossLessTableProps> = ({ }) => {
    const dispatch = useAppDispatch();

    const losslessState = useAppSelector(selectLosslessRule);
    const { datas, total, page, limit, loading } = losslessState;
    const { versions, versionTotal, versionPage, versionLimit, versionLoading } = losslessState;
    const { subscribers } = losslessState;

    // 编辑器状态
    const [editor, setEditor] = React.useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({
        visible: false,
        authorizeVisible: false,
        mode: '',
    });

    // 初始化加载
    React.useEffect(() => {
        refreshData();
    }, []);

    const refreshData = (page = 1, limit = 10, query = '') => {
        dispatch(listLossLessRules({
            param: {
                limit: limit,
                offset: (page - 1) * limit,
                name: query,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询无损规则列表错误: ${res.payload as string}`);
            }
        });
    }

    const handleOpRule = (row: TableRowData, op: Op) => {
        switch (op) {
            case 'create':
                setEditor((prev) => ({ ...prev, visible: true, mode: 'create' }));
                break;
            case 'view':
                setEditor((prev) => ({ ...prev, visible: true, mode: 'view', data: { ...row } }));
                dispatch(editorLosslessRule({ ...row } as LossLessRuleView));
                break;
            case 'delete':
                dispatch(removeLosslessRule({ ids: [row.id as string] }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification("请求成功", "删除无损规则成功");
                            refreshData(1, limit);
                        } else {
                            openErrNotification("请求失败", `删除无损规则失败: ${res.payload as string || '未知'}`);
                        }
                    });
                break;
            case 'authorize':
                setEditor((prev) => ({ ...prev, authorizeVisible: true, mode: 'authorize', data: { ...row } }));
                break;
            default:
                break;
        }
    }

    const refreshVersions = (page = 1, limit = 10) => {
        dispatch(listLosslessRuleVersions({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                id: editor.data?.id as string,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询无损版本列表错误: ${res.payload as string}`);
            }
        });
    }

    const operateRelease = async (op: Op, row: TableRowData) => {
        switch (op) {
            case 'delete':
                dispatch(removeLosslessVersion({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "删除无损规则版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `删除无损规则版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
            case 'rollback':
                dispatch(rollbackLosslessVersion({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "回滚无损规则版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `回滚无损规则版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
        }
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
                                refreshData(1, limit, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshData(1, limit)} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editor.authorizeVisible && (
                <AuthorizeInput
                    resource_type={PolicySourceType.LossLessRules}
                    resource_id={editor.data?.id as string}
                    resource_name={`lossless_rule/${editor.data?.name}`}
                    visible={editor.authorizeVisible}
                    onClose={() => {
                        setEditor((prev) => ({ ...prev, authorizeVisible: false }));
                    }}
                />
            )}
            <Table
                loading={loading}
                data={datas || []}
                columns={columns(handleOpRule, (row: TableRowData) => {
                    handleOpRule(row, 'view');
                })}
                tableLayout='auto'
                rowKey="id"
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
                visible={editor.visible}
                title={editor.mode === 'create' ? '新建无损规则' : `${editor.data?.namespace || '-'}/${editor.data?.service || '-'}`}
                subtitle="无损上下线"
                onClose={() => setEditor((prev) => ({ ...prev, visible: false }))}
            >
                <RuleTabs
                    op={editor.mode}
                    onVersionView={() => {
                        refreshVersions(1, 10);
                    }}
                    view={
                        <>
                            <LossLessEditor
                                visible={editor.visible}
                                op={editor.mode}
                                refresh={(close: boolean) => {
                                    if (close) {
                                        setEditor((prev) => ({ ...prev, visible: false }));
                                    }
                                    refreshData(1, limit);
                                }}
                            />
                        </>
                    }
                    versions={{
                        datas: versions,
                        action: operateRelease,
                        editable: editor.data?.editable ?? true,
                        deleteable: editor.data?.deleteable ?? true,
                        rollbackable: false,
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
                                    title={`${editor.data?.namespace}/${editor.data?.service}`}
                                    editable={editor.data?.editable ?? true}
                                    deleteable={editor.data?.deleteable ?? true}
                                    subscribers={subscribers || []}
                                />
                            </div>
                        </>
                    }
                />
            </RuleDetailDrawer>
        </div>
    );
}

export default LossLessTable;
