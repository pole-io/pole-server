import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Popconfirm, Empty } from 'tdesign-react';
import { DeleteIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import Text from 'components/Text';
import { Op } from 'services/types';
import { LaneGroupView } from 'services/lane';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import Search from 'components/Search';
import LaneGroupEdtor from './LaneGroupEdtor';
import LaneRuleTable from './LaneRuleTable';
import { cleanLaneGroupPage, editorLaneGroup, listLaneGroups, listLaneGroupVersions, removeLaneGroups, removeLaneGroupVersion, rollbackLanGroupVersion, selectLaneGroup } from 'modules/governance/lane_group';
import RuleTabs from '../RuleRelease/RuleTabs';
import SubscribeTable from 'components/SubscribeTable';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import RuleDetailDrawer from '../RuleRelease/RuleDetailDrawer';

interface ILaneGroupTableProps {

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
        colKey: 'lane_count',
        title: '泳道',
        cell: ({ row }) => <div>{row?.rules?.length || 0} 个</div>,
    },
    {
        colKey: 'description',
        title: '描述',
        ellipsis: true,
        cell: ({ row: { description } }: TableRowData) => (<Text>{description || '-'}</Text>),
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

const LaneGroupTable: React.FC<ILaneGroupTableProps> = ({ }) => {
    const dispatch = useAppDispatch();

    const laneGroupState = useAppSelector(selectLaneGroup);
    const { datas, total, page, limit, loading } = laneGroupState;
    const { versions, versionTotal, versionPage, versionLimit, versionLoading } = laneGroupState;
    const { subscribers } = laneGroupState;

    // 合并编辑相关状态
    const [editor, seteditor] = useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });

    const handleOpRule = (row: TableRowData, op: Op) => {
        switch (op) {
            case 'create':
                seteditor(prev => ({ ...prev, visible: true, mode: 'create', data: undefined }));
                break;
            case 'view':
                seteditor(prev => ({ ...prev, visible: true, mode: 'view', data: row }));
                dispatch(editorLaneGroup(row as LaneGroupView));
                // 保存当前选中行数据，供其它 Tab(如 泳道/版本/监听) 使用
                break;
            case 'delete':
                dispatch(removeLaneGroups({ ids: [row.id] })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "删除泳道组成功");
                        if (editor.data?.id === row.id) {
                            seteditor(prev => ({ ...prev, visible: false, data: undefined }));
                        }
                        refreshData(1, limit);
                    } else {
                        openErrNotification("请求失败", `删除泳道组失败: ${res.payload as string || '未知'}`);
                    }
                });
                break;
            case 'authorize':
                seteditor(prev => ({ ...prev, authorizeVisible: true, data: { ...row } }));
                break;
        }
    }

    const refreshData = (page = 1, limit = 10, query = '') => {
        dispatch(listLaneGroups({
            param: {
                limit,
                offset: (page - 1) * limit,
                brief: true,
                name: query,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `获取泳道组列表失败, ${res?.payload as string}`);
            }
        })
    }

    const refreshVersions = (page = 1, limit = 10) => {
        dispatch(listLaneGroupVersions({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                id: editor.data?.id as string,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `查询泳道组版本列表错误: ${res.payload as string}`);
            }
        });
    }

    const operateRelease = async (op: Op, row: TableRowData) => {
        switch (op) {
            case 'delete':
                dispatch(removeLaneGroupVersion({
                    ids: [row.id]
                })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "删除泳道组版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `删除泳道组版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
            case 'rollback':
                dispatch(rollbackLanGroupVersion({ id: row.id })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "回滚泳道组版本成功");
                        refreshVersions(versionPage, versionLimit);
                    } else {
                        openErrNotification("请求失败", `回滚泳道组版本失败: ${res.payload as string || '未知'}`);
                    }
                })
                break;
        }
    }

    React.useEffect(() => {
        refreshData();

        return () => {
            dispatch(cleanLaneGroupPage());
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
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
                    total,
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
                {editor.authorizeVisible && editor.data?.id && (
                    <AuthorizeInput
                        resource_type={PolicySourceType.LaneRules}
                        resource_id={editor.data.id as string}
                        resource_name={`lane_group/${editor.data.name || editor.data.id}`}
                        visible={editor.authorizeVisible}
                        onClose={() => {
                            seteditor(pre => ({ ...pre, authorizeVisible: false }));
                        }}
                    />
                )}
            </section>
            <RuleDetailDrawer
                visible={editor.visible}
                title={editor.mode === 'create' ? '新建泳道组' : editor.data?.name || '泳道组详情'}
                subtitle="全链路灰度"
                onClose={() => seteditor(pre => ({ ...pre, visible: false }))}
            >
                <RuleTabs
                    op={editor.mode}
                    onVersionView={() => {
                        refreshVersions()
                    }}
                    view={
                        editor.mode !== 'create' && !editor.data ? (
                            <Empty title="选择泳道组查看详情" />
                        ) : (
                            <div className={style.laneDetailStack}>
                                <LaneGroupEdtor
                                    op={editor.mode}
                                    refresh={(close: boolean) => {
                                        if (close) {
                                            seteditor(pre => ({ ...pre, visible: false }));
                                        }
                                        refreshData(1, limit)
                                    }}
                                />
                                {editor.mode !== 'create' && (
                                    <section className={style.laneRulesPanel}>
                                        <div className={style.laneRulesHeader}>
                                            <div className={style.laneRulesTitle}>泳道列表</div>
                                            <div className={style.laneRulesHint}>当前泳道组内的匹配规则和泳道标签。</div>
                                        </div>
                                        <LaneRuleTable groupId={editor.data?.id || ''} />
                                    </section>
                                )}
                            </div>
                        )
                    }
                    versions={{
                        datas: versions,
                        action: operateRelease,
                        editable: editor.data?.editable ?? true,
                        deleteable: editor.data?.deleteable ?? true,
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
                                    title={`${editor.data?.name}`}
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
    )
}

export default React.memo(LaneGroupTable);
