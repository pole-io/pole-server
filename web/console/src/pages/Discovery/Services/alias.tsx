import React, { useMemo, useState } from 'react';
import { Link, Table, Button, Drawer, PrimaryTableProps, Space, TableRowData, Input } from 'components/Fluent';
import { AddIcon, SearchIcon } from 'components/Fluent/icons';
import { useNavigate } from 'components/Router';

import Text from 'components/Text';
import { ConfirmOperationButton, OperationButton } from 'components/OperationButton';
import { ResourceToolbar } from 'components/ResourceLayout';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { cleanAliasPage, listServiceAliass, removeServiceAliass, resetServiceAlias } from 'modules/discovery/alias';
import AliasEditor from './AliasEditor';
import { Op } from 'services/types';
import { copyToClipboard } from 'utils/sys';
import ResourceNameLink from 'components/ResourceNameLink';

interface IServiceAliasProps {
    namespace?: string;
    serviceName?: string;
    embedded?: boolean;
}

export interface ServiceAliasTableHandle {
    refresh: () => void;
    create: () => void;
}

const columns = (
    operateService: (op: Op, row: TableRowData) => void,
    redirect: (service: string, namespace: string) => void,
    inServiceDetail: boolean,
    copyAlias: (row: TableRowData) => void,
    viewAlias: (row: TableRowData) => void,
): PrimaryTableProps['columns'] => {
    const baseColumns: PrimaryTableProps['columns'] = [
        {
        colKey: 'alias_namespace',
        title: '别名命名空间',
        width: 180,
        fixed: 'left',
        cell: ({ row: { alias_namespace } }) => <Text title={String(alias_namespace || '-')}>{alias_namespace}</Text>,
        },
        {
        colKey: 'alias',
        title: '服务别名',
        width: 200,
        cell: ({ row }) => (
            <span title={String(row.alias || '-')}>
                <ResourceNameLink name={row.alias} onClick={() => viewAlias(row)} />
            </span>
        ),
        },
    ];

    if (!inServiceDetail) {
        baseColumns.push(
            {
                colKey: 'namespace',
                title: '目标服务命名空间',
                width: 200,
                cell: ({ row: { namespace } }: TableRowData) => <Text title={String(namespace || '-')}>{namespace}</Text>,
            },
            {
                colKey: 'service',
                title: '目标服务名',
                width: 220,
                cell: ({ row: { service, namespace } }) => {
                    return (
                        <Link theme='primary'
                            title={String(service || '-')}
                            onClick={() => {
                                redirect(service, namespace);
                            }}
                        >
                            {service}
                        </Link>
                    )
                }
            },
        );
    }

    baseColumns.push(
        {
        colKey: 'comment',
        title: '描述',
        width: 240,
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => (<Text title={String(comment || '-')}>{comment || '-'}</Text>),
        },
        {
        colKey: 'time',
        title: '操作时间',
        width: 210,
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>修改: {mtime}<br />创建: {ctime}</Text>,
        },
        {
        colKey: 'action',
        title: '操作',
        width: 132,
        fixed: 'right',
        cell: ({ row }) => {
            return (
                <Space>
                    <OperationButton action="view" onClick={() => viewAlias(row)} />
                    <ConfirmOperationButton action="delete" disabled={row.deleteable === false} disabledLabel="无权限操作" confirmContent="确认删除吗" onConfirm={() => operateService('delete', row)} />
                    <OperationButton action="copy" label="复制别名" onClick={() => copyAlias(row)} />
                </Space>
            )
        },
        },
    );

    return baseColumns;
}

const ServiceAliasTable = React.forwardRef<ServiceAliasTableHandle, IServiceAliasProps>((props, ref) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const { namespace, serviceName, embedded = false } = props;
    const inServiceDetail = Boolean(namespace && serviceName);

    const aliasState = useAppSelector(state => state.discoveryServiceAlais);
    const { datas, loading, total, page, limit } = aliasState;
    const targetServiceLabel = namespace && serviceName ? `${namespace} / ${serviceName}` : '-';
    const aliasNamespaceCount = useMemo(() => {
        return new Set(datas.map(item => item.alias_namespace).filter(Boolean)).size;
    }, [datas]);

    // 合并编辑相关状态
    const [searchState, setSearchState] = useState<{
        query: string;
    }>({ query: '' });

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        mode: 'create' | 'edit';
        data?: TableRowData;
    }>({ visible: false, mode: 'create', data: undefined });
    const [viewingAlias, setViewingAlias] = useState<TableRowData | null>(null);

    // 编辑、新建事件
    const operateService = (op: Op, row?: TableRowData) => {
        switch (op) {
            case 'create':
            case 'edit':
                dispatch(resetServiceAlias());
                setEditorState(prev => ({ ...prev, visible: true, mode: op, data: row }));
                break;
            case 'delete':
                dispatch(removeServiceAliass({
                    param: [{
                        alias: row?.alias || '',
                        alias_namespace: row?.alias_namespace || '',
                    }]
                }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'rejected') {
                            openErrNotification("请求失败", res.payload as string);
                        } else {
                            openInfoNotification("请求成功", `删除服务别名 ${row?.alias} 成功`);
                            refreshTable();
                        }
                    });
                break;
            default:
                break;
        }
    }

    const copyAlias = (row: TableRowData) => {
        copyToClipboard(`${row.alias_namespace || '-'}/${row.alias || '-'}`);
    };

    const refreshTable = (page = 1, limit = 10, nextQuery = searchState) => {
        dispatch(listServiceAliass({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                alias: nextQuery.query || undefined,
                namespace: inServiceDetail ? namespace : undefined,
                service: inServiceDetail ? serviceName : undefined,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("获取数据失败", res.payload as string);
            }
        });
    }

    React.useImperativeHandle(ref, () => ({
        refresh: () => refreshTable(page, limit, searchState),
        create: () => operateService('create'),
    }));

    const submitFilter = () => {
        refreshTable(1, limit, searchState);
    };

    const resetFilter = () => {
        const nextQuery = { query: '' };
        setSearchState(nextQuery);
        refreshTable(1, limit, nextQuery);
    };

    React.useEffect(() => {
        refreshTable();
        return () => {
            // 清理编辑器状态
            dispatch(cleanAliasPage());
        }
    }, [namespace, serviceName]);

    {/* <!-- :defaultExpandedRowKeys="defaultExpandedRowKeys" --> */ }
    const table = (
        <>
            <section className={embedded ? style.aliasDetailSection : style.listSection}>
                <ResourceToolbar
                    density="compact"
                    title="别名清单"
                    count={loading ? '正在同步列表' : `当前显示 ${datas.length} 条`}
                    filters={(
                        <>
                        <Input
                            id={embedded ? 'keyword' : undefined}
                            className={style.filterInput}
                            clearable
                            prefixIcon={<SearchIcon />}
                            placeholder="别名"
                            value={searchState.query}
                            onChange={(value) => setSearchState({ query: String(value) })}
                            onEnter={submitFilter}
                        />
                        <Button variant="outline" onClick={submitFilter}>查询</Button>
                        <Button variant="text" onClick={resetFilter}>重置</Button>
                        {embedded && (
                            <Button theme="primary" icon={<AddIcon />} onClick={() => operateService('create')}>
                                新建别名
                            </Button>
                        )}
                        </>
                    )}
                />
                {editorState.visible && (
                    <AliasEditor
                        key={editorState.mode + (editorState.data?.alias || 'new') + (editorState.visible ? '1' : '0')}
                        op={editorState.mode}
                        visible={editorState.visible}
                        data={editorState.data}
                        existingAliases={datas}
                        targetService={inServiceDetail ? { namespace: namespace as string, serviceName: serviceName as string } : undefined}
                        closeDrawer={() => {
                            // 关闭后重置编辑器状态
                            dispatch(resetServiceAlias());
                            setEditorState(s => ({ ...s, visible: false }));
                            refreshTable();
                        }} />
                )}
                <section className={embedded ? `${style.tableSurface} ${style.aliasDetailTableSurface}` : style.tableSurface}>
                    {embedded && (
                        <section className={style.aliasSummaryBar} id="summary">
                            <div className={style.aliasSummaryItem}>
                                <span className={style.aliasSummaryLabel}>目标服务</span>
                                <strong className={style.aliasSummaryValue} id="sumTarget" title={targetServiceLabel}>{targetServiceLabel}</strong>
                            </div>
                            <div className={style.aliasSummaryItem}>
                                <span className={style.aliasSummaryLabel}>别名数</span>
                                <strong className={style.aliasSummaryValue} id="sumCount">{datas.length}</strong>
                            </div>
                            <div className={style.aliasSummaryItem}>
                                <span className={style.aliasSummaryLabel}>覆盖命名空间</span>
                                <strong className={style.aliasSummaryValue} id="sumNs">{aliasNamespaceCount}</strong>
                            </div>
                        </section>
                    )}
                    <div id={embedded ? 'tbody' : undefined}>
                        <Table
                            data={datas.map((item) => ({
                                id: `${item.alias_namespace}/${item.alias}`,
                                ...item,
                            }))}
                            columns={columns(operateService, (service: string, namespace: string) => {
                                navigate(`instance?namespace=${namespace}&service=${service}`);
                            }, inServiceDetail, copyAlias, setViewingAlias)}
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
                                    refreshTable(pageInfo.current, pageInfo.pageSize, searchState);
                                },
                            }}
                            onPageChange={(pageInfo) => {
                                refreshTable(pageInfo.current, pageInfo.pageSize, searchState);
                            }}
                        />
                    </div>
                </section>
            </section>
            <Drawer
                visible={Boolean(viewingAlias)}
                header={viewingAlias?.alias || '服务别名详情'}
                footer={viewingAlias?.editable === false ? false : (
                    <Button
                        theme="primary"
                        onClick={() => {
                            const alias = viewingAlias;
                            setViewingAlias(null);
                            if (alias) operateService('edit', alias);
                        }}
                    >
                        编辑
                    </Button>
                )}
                size="small"
                onClose={() => setViewingAlias(null)}
            >
                {viewingAlias && (
                    <div className={style.aliasViewGrid}>
                        <div><span>服务别名</span><strong>{viewingAlias.alias || '-'}</strong></div>
                        <div><span>别名命名空间</span><strong>{viewingAlias.alias_namespace || '-'}</strong></div>
                        <div><span>目标服务</span><strong>{viewingAlias.service || '-'}</strong></div>
                        <div><span>目标命名空间</span><strong>{viewingAlias.namespace || '-'}</strong></div>
                        <div className={style.aliasViewWide}><span>描述</span><strong>{viewingAlias.comment || '-'}</strong></div>
                    </div>
                )}
            </Drawer>
        </>
    );

    return (
        <div className={embedded ? `${style.embeddedWorkspace} ${style.embeddedWorkspaceCompact}` : style.workspace}>
            {table}
        </div>
    )
});

export default React.memo(ServiceAliasTable);
