import React, { useMemo, useState } from 'react';
import { Table, Button, PrimaryTableProps, Tooltip, Space, TableRowData, Popconfirm, Input } from 'tdesign-react';
import { AddIcon, DeleteIcon, RefreshIcon, CreditcardIcon, SearchIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import Text from 'components/Text';
import NamespaceEditor from './NamespaceEditor';
import style from './index.module.less';
import { cleanNamespacePage, editorNamespace, listNamespaces, removeNamespace, resetNamespace, selectNamespace } from 'modules/namespace';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { Namespace } from 'services/namespace';
import { Op } from 'services/types';

const columns = (operateNamespace: (op: Op, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        width: 180,
        cell: ({ row }: TableRowData) => {
            const metadata = row.metadata || {};
            const metadataCount = Object.keys(metadata).length;
            return (
                <div className={style.namespaceCell}>
                    <div className={style.namespaceNameRow}>
                        <Text>{row.name}</Text>
                    </div>
                    <div className={style.namespaceMeta}>
                        <span>{metadataCount > 0 ? `${metadataCount} 个标签` : '无标签'}</span>
                        <span>{row.editable === false ? '只读' : '可维护'}</span>
                    </div>
                </div>
            )
        },
        fixed: 'left',
    },
    {
        colKey: 'commnet',
        title: '描述',
        width: 205,
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => (
            <div className={style.descriptionCell}>
                <Text>{comment || '-'}</Text>
            </div>
        ),
    },
    {
        colKey: 'totalSerivce',
        title: '服务数',
        width: 70,
        cell: ({ row }: TableRowData) => (
            <div className={style.numberCell}>
                <strong>{row.total_service_count ?? 0}</strong>
                <span>服务</span>
            </div>
        ),
    },
    {
        colKey: 'health/total',
        title: '健康/总数',
        width: 120,
        cell: ({ row: { total_instance_count, total_health_instance_count } }: TableRowData) => {
            const totalInstances = total_instance_count ?? 0;
            const healthInstances = total_health_instance_count ?? 0;
            const percent = totalInstances > 0 ? Math.round((healthInstances / totalInstances) * 100) : 0;
            return (
                <div className={style.healthCell}>
                    <div>
                        <strong>{`${healthInstances}/${totalInstances}`}</strong>
                        <span>{totalInstances > 0 ? `${percent}%` : '无实例'}</span>
                    </div>
                    <div className={style.healthTrack}>
                        <i style={{ width: `${percent}%` }} />
                    </div>
                </div>
            )
        },
    },
    {
        colKey: 'time',
        title: '操作时间',
        width: 145,
        cell: ({ row: { ctime, mtime } }: TableRowData) => (
            <div className={style.timeCell}>
                <span>修改：{mtime}</span>
                <span>创建：{ctime}</span>
            </div>
        ),
    },
    {
        colKey: 'action',
        title: '操作',
        width: 76,
        align: 'center',
        fixed: 'right',
        cell: ({ row }: TableRowData) => {
            return (
                <div className={style.actionCell}>
                    <Tooltip content={row.editable === false ? '无权限操作' : '查看 / 编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            aria-label="查看 / 编辑"
                            onClick={() => operateNamespace('view', row)}>
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
                                operateNamespace('delete', row);
                            }}
                        >
                            <Button shape="square" variant="text" disabled={row.deleteable === false}>
                                <DeleteIcon />
                            </Button>
                        </Popconfirm>
                    </Tooltip>
                </div>
            )
        },
    },
]

export default React.memo(() => {
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas, loading, page, limit, total } = namespaceState;
    const [query, setQuery] = useState('');

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        mode: 'create' | 'edit' | 'view';
        data?: TableRowData;
        authorizeVisible: boolean;
    }>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });

    // 编辑、新建事件
    const operateNamespace = (op: Op, row?: TableRowData) => {
        switch (op) {
            case 'edit':
                dispatch(editorNamespace({ ...row as Namespace }));
                setEditorState(prev => ({ ...prev, visible: true, mode: 'edit', data: { ...row } }));
                break;
            case 'view':
                dispatch(editorNamespace({ ...row as Namespace }));
                setEditorState(prev => ({ ...prev, visible: true, mode: 'view', data: { ...row } }));
                break;
            case 'create':
                dispatch(resetNamespace());
                setEditorState(prev => ({ ...prev, visible: true, mode: 'create', data: undefined }));
                break;
            case 'authorize':
                setEditorState(prev => ({ ...prev, authorizeVisible: true, data: { ...row } }));
                break;
            case 'delete':
                dispatch(removeNamespace({ param: { name: row?.name as string } }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification('请求成功', '删除命名空间成功');
                            refreshTable(page, limit, query);
                        } else {
                            openErrNotification('请求失败', res.payload as string);
                        }
                    });
                break;
        }
    }

    const refreshTable = (page = 1, limit = 10, query = '') => {
        dispatch(listNamespaces({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                name: query || undefined,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", res.payload as string);
            }
        });
    }

    React.useEffect(() => {
        refreshTable();
        return () => {
            // 清理编辑器状态
            dispatch(cleanNamespacePage());
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const metrics = useMemo(() => {
        const serviceCount = datas.reduce((sum, item) => sum + (item.total_service_count || 0), 0);
        const instanceCount = datas.reduce((sum, item) => sum + (item.total_instance_count || 0), 0);
        const healthCount = datas.reduce((sum, item) => sum + (item.total_health_instance_count || 0), 0);
        const healthRate = instanceCount > 0 ? `${Math.round((healthCount / instanceCount) * 100)}%` : '-';
        return { serviceCount, instanceCount, healthCount, healthRate };
    }, [datas]);

    const submitFilter = () => {
        refreshTable(1, limit, query);
    };

    const resetFilter = () => {
        setQuery('');
        refreshTable(1, limit, '');
    };

    const table = (
        <>
            <section className={style.header}>
                <div>
                    <div className={style.eyebrow}>Service Registry / Namespace</div>
                    <h2>命名空间管理</h2>
                    <p>维护服务隔离边界和访问授权，快速确认各命名空间下的服务与实例健康状态。</p>
                </div>
                <Space>
                    <Tooltip content="刷新列表">
                        <Button shape="square" variant="outline" onClick={() => refreshTable(page, limit, query)}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={() => operateNamespace('create')}>新建命名空间</Button>
                </Space>
            </section>

            <section className={style.metricRail}>
                <div className={style.metricItem}>
                    <span>命名空间</span>
                    <strong>{total}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>当前页服务</span>
                    <strong>{metrics.serviceCount}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>健康实例</span>
                    <strong>{metrics.healthCount}/{metrics.instanceCount}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>实例健康率</span>
                    <strong>{metrics.healthRate}</strong>
                </div>
            </section>

            <section className={style.filterBar}>
                <div className={style.filterHint}>
                    <strong>命名空间列表</strong>
                    <span>{loading ? '正在同步列表' : `当前显示 ${datas.length} 条`}</span>
                </div>
                <Space>
                    <Input
                        className={style.filterInput}
                        clearable
                        prefixIcon={<SearchIcon />}
                        placeholder="名称前缀"
                        value={query}
                        onChange={(value) => setQuery(value as string)}
                        onEnter={submitFilter}
                    />
                    <Button variant="outline" onClick={submitFilter}>查询</Button>
                    <Button variant="text" onClick={resetFilter}>重置</Button>
                </Space>
            </section>

            {editorState.visible && (
                <NamespaceEditor
                    key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                    op={editorState.mode}
                    visible={editorState.visible}
                    onAuthorize={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: true }));
                    }}
                    closeDrawer={() => {
                        // 清理编辑器状态
                        dispatch(resetNamespace());
                        setEditorState(s => ({ ...s, visible: false }))
                        refreshTable(page, limit, query);
                    }} />
            )}
            {editorState.authorizeVisible && (
                <AuthorizeInput
                    resource_type={PolicySourceType.Namespaces}
                    resource_id={editorState.data?.id}
                    resource_name={`${editorState.data?.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
            <section className={style.tableSurface}>
                <Table
                    data={datas}
                    columns={columns(operateNamespace)}
                    loading={loading}
                    rowKey="name"
                    size={"large"}
                    tableLayout={'fixed'}
                    cellEmptyContent={'-'}
                    pagination={{
                        current: page,
                        pageSize: limit,
                        total: total,
                        showJumper: true,
                        onChange(pageInfo) {
                            refreshTable(pageInfo.current, pageInfo.pageSize, query);
                        },
                    }}
                    onPageChange={(pageInfo) => {
                        refreshTable(pageInfo.current, pageInfo.pageSize, query);
                    }}
                />
            </section>
        </>
    );

    return (
        <div className={style.page}>
            {table}
        </div>
    )
});
