import React, { useEffect, useState } from 'react';
import { Table, Button, PageInfo, Popconfirm, PrimaryTableProps, TableProps, Space, TableRowData, Tag, Tooltip } from 'components/Fluent';
import { AddIcon, RefreshIcon, System2Icon, User1Icon, UsergroupIcon } from 'components/Fluent/icons';
import { useNavigate } from 'components/Router';

import Search from 'components/Search';
import ErrorPage from 'components/ErrorPage';
import Text from 'components/Text';
import { ConfirmOperationButton, OperationButton } from 'components/OperationButton';
import ResourceNameLink from 'components/ResourceNameLink';
import { ResourceToolbar } from 'components/ResourceLayout';
import { useAppDispatch } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { describeAuthPolicies } from 'services/auth_policy';
import PolicyEditor from './PolicyEditor';
import { editorPolicyRules, removePolicyRules, resetPolicyRules } from 'modules/auth/policy';

interface IPolicyTableProps {
    type: 'default' | 'custom';
}

const ServerError = () => <ErrorPage code={500} />;

const defaultColumns = (
    handleEditPolicy: (row: TableRowData, op: 'view' | 'create' | 'edit' | 'delete', res: string) => void,
    openPolicyDetail: (row: TableRowData) => void,
): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false || row.user_type === 'main' }),
    },
    {
        colKey: 'policy_type',
        title: '类型',
        cell: ({ row: { name } }: TableRowData) => {
            if (name.indexOf('(用户)') !== -1) {
                return (
                    <><User1Icon /><Text> 用户</Text></>
                )
            }
            if (name.indexOf('(用户组)') !== -1) {
                return (
                    <>
                        <UsergroupIcon /><Text> 用户组</Text>
                    </>
                )
            }
            return (
                <><System2Icon /><Text> 系统</Text></>
            )
        },
    },
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }) => {
            const name = row.name as string;
            let displayName = name
            if (name.indexOf('(用户组)') === 0) {
                displayName = name.replace('(用户组)', '')
            }
            if (name.indexOf('(用户)') === 0) {
                displayName = name.replace('(用户)', '')
            }
            return <ResourceNameLink name={displayName} onClick={() => openPolicyDetail(row)} />;
        },
    },
    {
        colKey: 'action',
        title: '行为',
        cell: ({ row: { action } }: TableRowData) => (<Text>{action || '-'}</Text>),
    },
    {
        colKey: 'source',
        title: '来源',
        cell: ({ row: { source } }: TableRowData) => (<Text>{source}</Text>),
    },
    {
        colKey: 'commnet',
        title: '描述',
        ellipsis: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
    },
    {
        colKey: 'default_strategy',
        title: '默认策略',
        cell: ({ row: { default_strategy } }: TableRowData) => (<Tag theme={default_strategy ? 'success' : 'danger'} variant="outline">{default_strategy ? '是' : '否'}</Tag>),
    },
    {
        colKey: 'time',
        title: '操作时间',
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>修改: {mtime}<br />创建: {ctime}</Text>,
    },
    {
        colKey: 'operation',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    <OperationButton action="view" onClick={() => openPolicyDetail(row)} />
                    {!row.default_strategy && (
                        <ConfirmOperationButton action="delete" disabled={row.deleteable === false} disabledLabel="无权限操作" confirmContent="确认删除吗" onConfirm={() => handleEditPolicy(row, 'delete', 'policy_rule')} />
                    )}
                </Space>
            )
        },
    },
]


const customColumns = (
    handleEditPolicy: (row: TableRowData, op: 'view' | 'create' | 'edit' | 'delete', res: string) => void,
    openPolicyDetail: (row: TableRowData) => void,
): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false || row.user_type === 'main' }),
    },
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }) => (
            <ResourceNameLink name={row.name} onClick={() => openPolicyDetail(row)} />
        ),
    },
    {
        colKey: 'action',
        title: '行为',
        cell: ({ row: { action } }: TableRowData) => (<Text>{action || '-'}</Text>),
    },
    {
        colKey: 'source',
        title: '来源',
        cell: ({ row: { source } }: TableRowData) => (<Text>{source}</Text>),
    },
    {
        colKey: 'commnet',
        title: '描述',
        ellipsis: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
    },
    {
        colKey: 'default_strategy',
        title: '默认策略',
        cell: ({ row: { default_strategy } }: TableRowData) => (<Tag theme={default_strategy ? 'success' : 'danger'} variant="outline">{default_strategy ? '是' : '否'}</Tag>),
    },
    {
        colKey: 'time',
        title: '操作时间',
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>修改: {mtime}<br />创建: {ctime}</Text>,
    },
    {
        colKey: 'operation',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    <OperationButton action="view" onClick={() => openPolicyDetail(row)} />
                    {!row.default_strategy && (
                        <ConfirmOperationButton action="delete" disabled={row.deleteable === false} disabledLabel="无权限操作" confirmContent="确认删除吗" onConfirm={() => handleEditPolicy(row, 'delete', 'policy_rule')} />
                    )}
                </Space>
            )
        },
    },
]


const PolicyTable: React.FC<IPolicyTableProps> = (props) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const [selectedRowKeys, setSelectedRowKeys] = useState<Array<string | number>>([]);

    // 合并编辑相关状态
    const [searchState, setSearchState] = useState<{
        policies: TableProps['data'];
        current: number;
        limit: number;
        previous: number;
        total: number;
        query: string;
        fetchError: boolean;
        isLoading: boolean;
    }>({ policies: [], current: 1, limit: 10, previous: 0, total: 0, query: '', fetchError: false, isLoading: false });

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        resource: string;
        mode: 'create' | 'edit' | 'view' | 'delete';
        data?: TableRowData;
    }>({ visible: false, resource: '', mode: 'create', data: undefined });

    const openPolicyDetail = (row: TableRowData) => {
        navigate(`/auth/policies/detail?name=${encodeURIComponent(String(row.name || ''))}&id=${encodeURIComponent(String(row.id || ''))}`);
    };

    const handleBatchDeletePolicies = async (ids: string[]) => {
        if (!ids.length) {
            return;
        }
        setSearchState(s => ({ ...s, isLoading: true }));
        const result = await dispatch(removePolicyRules({ ids }));
        if (result.meta.requestStatus === 'fulfilled') {
            openInfoNotification('请求成功', `已删除 ${ids.length} 条鉴权策略`);
            setSelectedRowKeys([]);
            await fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, searchState.query);
            return;
        }
        setSearchState(s => ({ ...s, isLoading: false }));
        openErrNotification('删除鉴权策略失败', String(result.payload || '未知错误'));
    };

    // 编辑、新建事件
    const handleEditPolicy = (row: TableRowData, mode: 'view' | 'create' | 'edit' | 'delete', res: string) => {
        if (mode === 'delete') {
            void handleBatchDeletePolicies([row.id as string]);
            return;
        }
        dispatch(editorPolicyRules({
            id: row.id as string,
            name: row.name,
            action: row.action,
            comment: row.comment,
            default_strategy: row.default_strategy,
            metadata: row.metadata,
        }));

        setEditorState({
            visible: true,
            mode: mode,
            resource: res,
            data: { ...row },
        })
    }

    const handleCreatePolicy = () => {
        setEditorState({
            visible: true,
            mode: 'create',
            resource: 'policy_rule',
            data: undefined,
        });
    };

    // 模拟远程请求
    async function fetchData(pageInfo: PageInfo, searchParam?: string) {
        const query = searchParam ?? searchState.query;
        setSearchState(s => ({ ...s, current: pageInfo.current, limit: pageInfo.pageSize, previous: pageInfo.previous, query, fetchError: false, isLoading: true }));
        try {
            const { current, pageSize } = pageInfo;

            const params = {
                name: query,
                default: props.type === 'default' ? "true" : "false",
            }

            // 请求可能存在跨域问题
            const response = await describeAuthPolicies({
                limit: pageSize, offset: (current - 1) * pageSize, ...params
            });
            setSearchState(s => ({ ...s, policies: response.content, total: response.totalCount, isLoading: false }));
        } catch (error: Error | any) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification("获取数据失败", error);
        }
    }

    useEffect(() => {
        fetchData({ current: 1, pageSize: searchState.limit, previous: 0 });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const table = (
        <>
            <ResourceToolbar
                className={style.toolBar}
                title={props.type === 'custom' ? '自定义策略列表' : '默认策略列表'}
                count={searchState.isLoading ? '正在同步列表' : `共 ${searchState.total} 条`}
                filters={(
                    <>
                        <Search
                            value={searchState.query}
                            onChange={(value: string) => {
                                fetchData({ current: 1, pageSize: searchState.limit, previous: 0, }, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <Button aria-label="刷新策略列表" shape="square" variant="outline" onClick={() => fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, searchState.query)}>
                                <RefreshIcon />
                            </Button>
                        </Tooltip>
                        {(selectedRowKeys.length > 0 && props.type === 'custom') && (
                            <>
                                <span className={style.selectionHint}>已选 {selectedRowKeys.length} 项</span>
                                <Popconfirm
                                    content={`确认删除选中的 ${selectedRowKeys.length} 条策略吗？`}
                                    destroyOnClose
                                    placement="top"
                                    showArrow
                                    onConfirm={() => handleBatchDeletePolicies(selectedRowKeys.map(String))}
                                >
                                    <Button theme="danger">批量删除</Button>
                                </Popconfirm>
                            </>
                        )}
                        {props.type === 'custom' && (
                            <Button theme="primary" icon={<AddIcon />} onClick={handleCreatePolicy}>新建策略</Button>
                        )}
                    </>
                )}
            />
            <PolicyEditor
                key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                visible={editorState.visible && editorState.resource === 'policy_rule'}
                closeDrawer={() => {
                    // 关闭后重置编辑器状态
                    dispatch(resetPolicyRules());
                    setEditorState(s => ({ ...s, visible: false }));
                    fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, searchState.query);
                }} op={editorState.mode} />
            <Table
                data={searchState.policies}
                columns={props.type === 'default' ?
                    defaultColumns(handleEditPolicy, openPolicyDetail)
                    :
                    customColumns(handleEditPolicy, openPolicyDetail)}
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
            {searchState.fetchError ? (
                <ServerError />
            ) : (
                table
            )}
        </>
    )
}

export default React.memo(PolicyTable);
