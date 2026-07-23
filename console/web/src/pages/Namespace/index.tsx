import React, { useMemo, useState } from 'react';
import { Table, Button, PrimaryTableProps, TableRowData, Input, Tag, Tooltip } from 'components/Fluent';
import { AddIcon, RefreshIcon, SearchIcon } from 'components/Fluent/icons';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import Text from 'components/Text';
import { ConfirmOperationButton, OperationButton, OperationButtonGroup } from 'components/OperationButton';
import { ResourceHeader, ResourceToolbar } from 'components/ResourceLayout';
import NamespaceEditor from './NamespaceEditor';
import style from './index.module.less';
import { cleanNamespacePage, editorNamespace, listNamespaces, removeNamespace, resetNamespace, selectNamespace } from 'modules/namespace';
import AuthorizeInput from 'components/Authorize';
import ResourceNameLink from 'components/ResourceNameLink';
import { PolicySourceType } from 'services/auth_policy';
import { Namespace } from 'services/namespace';
import { Op } from 'services/types';

const DEFAULT_NAMESPACE = 'default';
const SYSTEM_NAMESPACE = 'pole-system';

const protectedNamespaceReason = (name?: string) => {
    if (name === SYSTEM_NAMESPACE) return 'Pole 内部系统空间不可删除';
    if (name === DEFAULT_NAMESPACE) return '默认命名空间不可删除';
    return '';
};

const columns = (operateNamespace: (op: Op, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        width: 220,
        cell: ({ row }: TableRowData) => {
            const isDefault = row.name === DEFAULT_NAMESPACE;
            const isSystem = row.name === SYSTEM_NAMESPACE;
            return (
                <div className={style.namespaceNameRow}>
                    <ResourceNameLink name={row.name} onClick={() => operateNamespace('view', row)} />
                    {isDefault && <Tag size="small" variant="light">默认空间</Tag>}
                    {isSystem && (
                        <Tooltip content="Pole 内部组件使用的系统命名空间">
                            <Tag size="small" theme="primary" variant="light">内部系统空间</Tag>
                        </Tooltip>
                    )}
                </div>
            );
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
        colKey: 'totalConfigFile',
        title: '配置文件',
        width: 90,
        cell: ({ row }: TableRowData) => (
            <div className={style.numberCell}>
                <strong>{row.total_config_file_count ?? 0}</strong>
                <span>文件</span>
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
        width: 108,
        fixed: 'right',
        cell: ({ row }: TableRowData) => {
            const protectedReason = protectedNamespaceReason(row.name);
            return (
                <OperationButtonGroup className={style.actionCell}>
                    <OperationButton action="viewEdit" disabled={row.editable === false} disabledLabel="无权限操作" onClick={() => operateNamespace('view', row)} />
                    <OperationButton action="authorize" disabled={row.editable === false} disabledLabel="无权限操作" onClick={() => operateNamespace('authorize', row)} />
                    <ConfirmOperationButton
                        action="delete"
                        disabled={Boolean(protectedReason) || row.deleteable === false}
                        disabledLabel={protectedReason || '无权限操作'}
                        confirmContent="确认删除吗"
                        onConfirm={() => operateNamespace('delete', row)}
                    />
                </OperationButtonGroup>
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
                if (protectedNamespaceReason(row?.name as string)) {
                    openErrNotification('无法删除命名空间', protectedNamespaceReason(row?.name as string));
                    break;
                }
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
        const configFileCount = datas.reduce((sum, item) => sum + (item.total_config_file_count || 0), 0);
        const instanceCount = datas.reduce((sum, item) => sum + (item.total_instance_count || 0), 0);
        const healthCount = datas.reduce((sum, item) => sum + (item.total_health_instance_count || 0), 0);
        const healthRate = instanceCount > 0 ? `${Math.round((healthCount / instanceCount) * 100)}%` : '-';
        return { serviceCount, configFileCount, instanceCount, healthCount, healthRate };
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
            <ResourceHeader
                eyebrow="Environment / Namespace"
                title="命名空间管理"
                description="命名空间代表独立运行环境，统一纳管该环境下的服务、配置与治理规则，并隔离权限和发布状态。"
                actions={(
                    <>
                    <Tooltip content="刷新列表">
                        <Button shape="square" variant="outline" onClick={() => refreshTable(page, limit, query)}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={() => operateNamespace('create')}>新建命名空间</Button>
                    </>
                )}
            />

            <section className={style.namespaceWorkspace}>
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
                        <span>当前页配置文件</span>
                        <strong>{metrics.configFileCount}</strong>
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

                <ResourceToolbar
                    title="命名空间列表"
                    count={loading ? '正在同步列表' : `当前显示 ${datas.length} 条`}
                    filters={(
                        <>
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
                        </>
                    )}
                />

            {editorState.visible && (
                <NamespaceEditor
                    key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                    op={editorState.mode}
                    visible={editorState.visible}
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
                    resource_id={editorState.data?.name}
                    resource_name={`${editorState.data?.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
                <section className={`${style.tableSurface} ${style.namespaceTableSurface}`}>
                <Table
                    data={datas}
                    columns={columns(operateNamespace)}
                    loading={loading}
                    rowKey="name"
                    size={"large"}
                    tableLayout="fixed"
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
                />
                </section>
            </section>
        </>
    );

    return (
        <div className={style.page}>
            {table}
        </div>
    )
});
