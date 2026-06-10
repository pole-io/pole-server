import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, TableRowData, Popconfirm } from 'tdesign-react';
import { AddIcon, DeleteIcon, EditIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';
import { useNavigate } from 'react-router-dom';

import Search from 'components/Search';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { cleanAliasPage, listServiceAliass, removeServiceAliass, resetServiceAlias } from 'modules/discovery/alias';
import AliasEditor from './AliasEditor';
import { Op } from 'services/types';

interface IServiceAliasProps {

}

const columns = (operateService: (op: Op, row: TableRowData) => void, redirect: (service: string, namespace: string) => void): PrimaryTableProps['columns'] => [
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
    {
        colKey: 'commnet',
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
                    <Tooltip content={row.editable === false ? '无权限操作' : '授权'}>
                        <Button shape="square" variant="text" disabled={row.editable === false}>
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
                                operateService('delete', row);
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

const ServiceAliasTable: React.FC<IServiceAliasProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    const aliasState = useAppSelector(state => state.discoveryServiceAlais);
    const { datas, loading, total, page, limit } = aliasState;

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
                        alias_namespace: row?.namespace || '',
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

    const refreshTable = (page = 1, limit = 10, nextQuery = searchState) => {
        dispatch(listServiceAliass({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                alias: nextQuery.query || undefined,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("获取数据失败", res.payload as string);
            }
        });
    }

    React.useEffect(() => {
        refreshTable();
        return () => {
            // 清理编辑器状态
            dispatch(cleanAliasPage());
        }
    }, []);

    {/* <!-- :defaultExpandedRowKeys="defaultExpandedRowKeys" --> */ }
    const table = (
        <>
            <section className={style.filterBar}>
                <div className={style.filterHint}>
                    <strong>别名清单</strong>
                    <span>{loading ? '正在同步列表' : `当前显示 ${datas.length} 条`}</span>
                </div>
                <Space>
                    <Search
                        placeholder="搜索别名"
                        onChange={(value: string) => {
                            const nextQuery = { query: value };
                            setSearchState(nextQuery);
                            refreshTable(1, limit, nextQuery);
                        }}
                    />
                    <Tooltip content="刷新">
                        <Button shape="square" variant="outline" onClick={() => refreshTable(page, limit, searchState)}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={() => operateService('create')}>新建</Button>
                </Space>
            </section>
            {editorState.visible && (
                <AliasEditor
                    key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                    op={editorState.mode}
                    visible={editorState.visible}
                    closeDrawer={() => {
                        // 关闭后重置编辑器状态
                        dispatch(resetServiceAlias());
                        setEditorState(s => ({ ...s, visible: false }));
                        refreshTable();
                    }} />
            )}
            <section className={style.tableSurface}>
                <Table
                    data={datas}
                    columns={columns(operateService, (service: string, namespace: string) => {
                        navigate(`instance?namespace=${namespace}&service=${service}`);
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
                            refreshTable(pageInfo.current, pageInfo.pageSize, searchState);
                        },
                    }}
                    onPageChange={(pageInfo) => {
                        refreshTable(pageInfo.current, pageInfo.pageSize, searchState);
                    }}
                />
            </section>
        </>
    );

    return (
        <div className={style.workspace}>
            {table}
        </div>
    )
}

export default React.memo(ServiceAliasTable);
