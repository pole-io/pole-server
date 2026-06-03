import React, { useEffect, useState } from 'react';
import { Link, Table, Button, PageInfo, PrimaryTableProps, TableProps, Tooltip, Space, Row, Col, TableRowData, Tabs, Tag, Dialog, Input, Popconfirm } from 'tdesign-react';
import { DeleteIcon, EditIcon, RefreshIcon, System2Icon, User1Icon, UsergroupIcon, UserVisibleIcon } from 'tdesign-icons-react';
import { useNavigate, useLocation } from 'react-router-dom';

import Search from 'components/Search';
import ErrorPage from 'components/ErrorPage';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { describeAuthPolicies } from 'services/auth_policy';
import PolicyEditor from './PolicyEditor';
import { editorPolicyRules, removePolicyRules, resetPolicyRules, updatePolicyRules } from 'modules/auth/policy';
import { update } from 'lodash';

interface IPolicyTableProps {
    type: 'default' | 'custom';
}

const ServerError = () => <ErrorPage code={500} />;

const defaultColumns = (handleEditPolicy: (row: TableRowData, op: 'view' | 'create' | 'edit' | 'delete', res: string) => void, redirect: (id: string, name: string) => void): PrimaryTableProps['columns'] => [
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
        cell: ({ row: { name, id } }) => {
            let displayName = name
            if (name.indexOf('(用户组)') === 0) {
                displayName = name.replace('(用户组)', '')
            }
            if (name.indexOf('(用户)') === 0) {
                displayName = name.replace('(用户)', '')
            }
            return (
                <Link theme='primary' onClick={() => redirect(id, name)}>{displayName}</Link>
            )
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
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => handleEditPolicy(row, 'edit', 'policy_rule')}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
                    {!row.default_strategy && (
                        <Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
                            <Button
                                shape="square"
                                variant="text"
                                disabled={row.deleteable === false}
                                onClick={() => {
                                    handleEditPolicy(row, 'delete', 'policy_rule');
                                }}
                            >
                                <DeleteIcon />
                            </Button>
                        </Tooltip>
                    )}
                </Space>
            )
        },
    },
]


const customColumns = (handleEditPolicy: (row: TableRowData, op: 'view' | 'create' | 'edit' | 'delete', res: string) => void, redirect: (id: string, name: string) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false || row.user_type === 'main' }),
    },
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row: { name, id } }) => (
            <Link theme='primary' onClick={() => redirect(id, name)}>{name}</Link>
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
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => handleEditPolicy(row, 'edit', 'user')}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
                    {!row.default_strategy && (
                        <Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
                            <Popconfirm
                                content="确认删除吗"
                                destroyOnClose
                                placement="top"
                                showArrow
                                theme="default"
                                onConfirm={() => {
                                    handleEditPolicy(row, 'delete', 'user');
                                }}
                            >
                                <Button
                                    shape="square"
                                    variant="text"
                                    disabled={row.deleteable === false}
                                >
                                    <DeleteIcon />
                                </Button>
                            </Popconfirm>
                        </Tooltip>
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

    // 编辑、新建事件
    const handleEditPolicy = (row: TableRowData, mode: 'view' | 'create' | 'edit' | 'delete', res: string) => {
        if (mode === 'delete') {
            setSearchState(s => ({ ...s, isLoading: true }));
            dispatch(removePolicyRules({ ids: [row.id as string] }))
                .then((res) => {
                    openInfoNotification("请求成功", "删除鉴权策略成功");
                })
                .catch((err) => {
                    openErrNotification("请求失败", err);
                })
                .finally(() => {
                    setSearchState(s => ({ ...s, isLoading: false }));
                });
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
        setSearchState(s => ({ ...s, current: pageInfo.current, limit: pageInfo.pageSize, previous: pageInfo.previous, fetchError: false, isLoading: true }));
        try {
            const { current, pageSize } = pageInfo;

            const params = {
                name: searchParam,
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
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            {props.type === 'custom' && (
                                <Button onClick={handleCreatePolicy}>新建</Button>
                            )}
                        </Col>
                        {(selectedRowKeys.length > 0 && props.type === 'custom') && (
                            <>
                                <Col>
                                    <Button theme='danger'>批量删除</Button>
                                </Col>
                                <Col>
                                    <div>已选 {selectedRowKeys?.length || 0} 项</div>
                                </Col>
                            </>
                        )}
                    </Row>
                </Col>
                <Col>
                    <Space>
                        <Search
                            onChange={(value: string) => {
                                fetchData({ current: 1, pageSize: searchState.limit, previous: 0, }, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => fetchData({ current: 1, pageSize: searchState.limit, previous: 0 })} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            <PolicyEditor
                key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                visible={editorState.visible && editorState.resource === 'policy_rule'}
                closeDrawer={() => {
                    // 关闭后重置编辑器状态
                    dispatch(resetPolicyRules());
                    setEditorState(s => ({ ...s, visible: false }));
                    fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, '');
                }} op={editorState.mode} />
            <Table
                data={searchState.policies}
                columns={props.type === 'default' ?
                    defaultColumns(handleEditPolicy, (id: string, name: string) => {
                        navigate(`detail?id=${id}&name=${name}`);
                    })
                    :
                    customColumns(handleEditPolicy, (id: string, name: string) => {
                        navigate(`detail?id=${id}&name=${name}`);
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
            {searchState.fetchError ? (
                <ServerError />
            ) : (
                table
            )}
        </>
    )
}

export default React.memo(PolicyTable);