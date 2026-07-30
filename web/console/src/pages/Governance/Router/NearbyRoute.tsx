import React, { useEffect, useState } from 'react';
import { Link, Popup, Table, Button, PageInfo, PrimaryTableProps, TableProps, Tooltip, Space, TableRowData, Tabs, Loading, Popconfirm } from 'components/Fluent';
import { AddIcon, DeleteIcon, EditIcon, RefreshIcon, CreditcardIcon } from 'components/Fluent/icons';
import { useNavigate, useLocation } from 'components/Router';

import { openErrNotification } from 'utils/notifition';
import { ResourceToolbar } from 'components/ResourceLayout';
import { describeCustomRoute } from 'services/router';
import { useAppDispatch, useAppSelector } from 'modules/store';
import Text from 'components/Text';
import { Op } from 'services/types';
import Search from 'components/Search';


interface INearbyRouteProps {

}

const columns = (handleOpRule: (row: TableRowData, op: Op) => void, redirect: (id: string) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false }),
    },
    {
        colKey: 'name',
        title: 'name',
        cell: ({ row: { id, name } }) => <Link
            theme="primary"
            onClick={() => { redirect(id) }}
        >{name}</Link>,
    },
    {
        colKey: 'caller',
        title: '主调',
        ellipsis: true,
        cell: ({ row: { priority } }: TableRowData) => (<Text>{priority || '-'}</Text>),
    },
    {
        colKey: 'callee',
        title: '被调',
        ellipsis: true,
        cell: ({ row: { priority } }: TableRowData) => (<Text>{priority || '-'}</Text>),
    },
    {
        colKey: 'priority',
        title: '优先级',
        ellipsis: true,
        cell: ({ row: { priority } }: TableRowData) => (<Text>{priority || '-'}</Text>),
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
                            onClick={() => handleOpRule(row, 'edit')}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
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

const NearbyRoute: React.FC<INearbyRouteProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const [selectedRowKeys, setSelectedRowKeys] = useState<Array<string | number>>([]);

    // 合并编辑相关状态
    const [searchState, setSearchState] = useState<{
        services: TableProps['data'];
        current: number;
        limit: number;
        previous: number;
        total: number;
        query: string;
        fetchError: boolean;
        isLoading: boolean;
    }>({ services: [], current: 1, limit: 10, previous: 0, total: 0, query: '', fetchError: false, isLoading: false });

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });

    const handleOpRule = (row: TableRowData, op: Op) => {

    }

    // 模拟远程请求
    async function fetchData(pageInfo: PageInfo, searchParam?: string) {
        setSearchState(s => ({ ...s, current: pageInfo.current, limit: pageInfo.pageSize, previous: pageInfo.previous, fetchError: false, isLoading: true }));
        try {
            const { current, pageSize } = pageInfo;
            // 请求可能存在跨域问题
            const response = await describeCustomRoute({
                route_type: 'NearbyPolicy', limit: pageSize, offset: (current - 1) * pageSize, ...(searchParam && { name: searchParam })
            });
            setSearchState(s => ({ ...s, services: response.list, total: response.totalCount, isLoading: false }));
        } catch (error: Error | any) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification("获取数据失败", error);
        }
    }

    useEffect(() => {
        fetchData({ current: 1, pageSize: searchState.limit, previous: 0 });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const refreshTable = () => {
        setSearchState(s => ({ ...s, fetchError: false, isLoading: true }));
        fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, searchState.query);
    }

    const table = (
        <>
            <ResourceToolbar
                density="compact"
                title="就近路由清单"
                count={searchState.isLoading ? '正在同步列表' : `共 ${searchState.total} 条`}
                filters={(
                    <>
                        {selectedRowKeys.length > 0 && (
                            <>
                                <span>已选 {selectedRowKeys.length} 项</span>
                                <Button theme='danger'>批量删除</Button>
                            </>
                        )}
                        <Search
                            onChange={(value: string) => {
                                fetchData({ current: 1, pageSize: searchState.limit, previous: 0, }, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <Button
                                aria-label="刷新就近路由列表"
                                shape="square"
                                variant="outline"
                                icon={<RefreshIcon />}
                                onClick={() => fetchData({ current: 1, pageSize: searchState.limit, previous: 0 })}
                            />
                        </Tooltip>
                        <Button theme="primary" icon={<AddIcon />} onClick={() => {
                            handleOpRule({}, 'create')
                        }}>新建就近路由</Button>
                    </>
                )}
            />
            <Table
                data={searchState.services || []}
                columns={columns(handleOpRule, (id: string) => {
                })}
                loading={searchState.isLoading}
                rowKey="id"
                size={"large"}
                tableLayout={'auto'}
                cellEmptyContent={'-'}
                pagination={{
                    current: searchState.current,
                    pageSize: searchState.limit,
                    total: searchState.total,
                    showJumper: true,
                    onChange(pageInfo) {
                        fetchData(pageInfo, searchState.query);
                    },
                }}
                onPageChange={(pageInfo) => {
                    fetchData(pageInfo, searchState.query);
                }}
                selectOnRowClick={false}
                selectedRowKeys={selectedRowKeys}
                onSelectChange={(selected: Array<string | number>) => {
                    setSelectedRowKeys(selected);
                }}
            />
        </>
    )

    return (
        <>
            {table}
        </>
    )
}

export default React.memo(NearbyRoute);
