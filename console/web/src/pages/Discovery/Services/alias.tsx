import React, { useMemo, useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, TableRowData, Popconfirm, Input } from 'tdesign-react';
import { AddIcon, CopyIcon, DeleteIcon, EditIcon, SearchIcon } from 'tdesign-icons-react';
import { useNavigate } from 'react-router-dom';

import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { cleanAliasPage, listServiceAliass, removeServiceAliass, resetServiceAlias } from 'modules/discovery/alias';
import AliasEditor from './AliasEditor';
import { Op } from 'services/types';
import { copyToClipboard } from 'utils/sys';

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
): PrimaryTableProps['columns'] => {
    const baseColumns: PrimaryTableProps['columns'] = [
        {
        colKey: 'alias_namespace',
        title: '别名命名空间',
        cell: ({ row: { alias_namespace } }) => <Text>{alias_namespace}</Text>,
        },
        {
        colKey: 'alias',
        title: '服务别名',
        cell: ({ row: { alias } }) => <Text>{alias}</Text>,
        },
    ];

    if (!inServiceDetail) {
        baseColumns.push(
            {
                colKey: 'namespace',
                title: '目标服务命名空间',
                cell: ({ row: { namespace } }: TableRowData) => <Text>{namespace}</Text>,
            },
            {
                colKey: 'service',
                title: '目标服务名',
                cell: ({ row: { service, namespace } }) => {
                    return (
                        <Link theme='primary'
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
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
        },
        {
        colKey: 'time',
        title: '操作时间',
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>修改: {mtime}<br />创建: {ctime}</Text>,
        },
        {
        colKey: 'action',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => operateService('edit', row)}>
                            <EditIcon />
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
                                operateService('delete', row);
                            }}
                        >
                            <Button shape="square" variant="text" disabled={row.deleteable === false}>
                                <DeleteIcon />
                            </Button>
                        </Popconfirm>
                    </Tooltip>
                    <Tooltip content="复制别名">
                        <Button
                            shape="square"
                            variant="text"
                            aria-label="复制别名"
                            onClick={() => copyAlias(row)}
                        >
                            <CopyIcon />
                        </Button>
                    </Tooltip>
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
                <section className={embedded ? style.aliasDetailToolbar : style.filterBar}>
                    <div className={style.filterHint}>
                        <strong>别名清单</strong>
                        <span id={embedded ? 'listCount' : undefined}>
                            {loading
                                ? '正在同步列表'
                                : embedded && namespace && serviceName
                                    ? `${namespace}/${serviceName} 下当前显示 ${datas.length} 条`
                                    : `当前显示 ${datas.length} 条`}
                        </span>
                    </div>
                    <div className={embedded ? style.aliasDetailActions : style.filterActions}>
                        {embedded && (
                            <Button theme="primary" icon={<AddIcon />} onClick={() => operateService('create')}>
                                新建别名
                            </Button>
                        )}
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
                    </div>
                </section>
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
                                <strong className={style.aliasSummaryValue} id="sumTarget">{targetServiceLabel}</strong>
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
                            }, inServiceDetail, copyAlias)}
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
        </>
    );

    return (
        <div className={embedded ? style.embeddedWorkspace : style.workspace}>
            {table}
        </div>
    )
});

export default React.memo(ServiceAliasTable);
