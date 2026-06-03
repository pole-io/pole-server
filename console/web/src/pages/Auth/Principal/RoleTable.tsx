import React, { useEffect, useState } from 'react';
import { Link, Table, Button, PageInfo, PrimaryTableProps, TableProps, Tooltip, Space, Row, Col, TableRowData, Tabs, Tag, Dialog, Input, Loading, Popconfirm } from 'tdesign-react';
import { DeleteIcon, EditIcon, RefreshIcon, System2Icon, User1Icon, UsergroupIcon, UserVisibleIcon } from 'tdesign-icons-react';
import { useNavigate, useLocation } from 'react-router-dom';

import Search from 'components/Search';
import ErrorPage from 'components/ErrorPage';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import PolicyEditor from './RoleEditor';
import { describeRoles } from 'services/role';
import { editorRoles, removeRoles, resetRoles } from 'modules/auth/role';
import RoleEditor from './RoleEditor';

type Op = 'view' | 'create' | 'edit' | 'delete';

interface IPolicyTableProps {
    type: 'default' | 'custom';
}

const ServerError = () => <ErrorPage code={500} />;

const customColumns = (handleEditRole: (row: TableRowData, op: Op, res: string) => void, redirect: (id: string, name: string) => void): TableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false || row.user_type === 'main' }),
    },
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row: { name, id } }) => {
            return (
                <Link theme='primary' onClick={() => redirect(id, name)}>{name}</Link>
            )
        },
    },
    {
        colKey: 'source',
        title: '来源',
        cell: ({ row: { source } }: TableRowData) => (<Text>{source}</Text>),
    },
    {
        title: '描述',
        colKey: 'commnet',
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => {
            console.log(comment);
            return <Text>{comment}</Text>
        }
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
                            onClick={() => handleEditRole(row, 'edit', 'role')}>
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
                                handleEditRole(row, 'delete', 'role');
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
                </Space>
            )
        },
    },
]


const RoleTable: React.FC<IPolicyTableProps> = (props) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const [selectedRowKeys, setSelectedRowKeys] = useState<Array<string | number>>([]);

    // 合并编辑相关状态
    const [searchState, setSearchState] = useState<{
        roles: TableProps['data'];
        current: number;
        limit: number;
        previous: number;
        total: number;
        query: string;
        fetchError: boolean;
        isLoading: boolean;
    }>({ roles: [], current: 1, limit: 10, previous: 0, total: 0, query: '', fetchError: false, isLoading: false });

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        resource: string;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, resource: '', mode: 'create', data: undefined });

    // 编辑、新建事件
    const handleEditRole = (row: TableRowData, mode: Op, res: string) => {
        if (mode === 'delete') {
            handleDelete([row.id]);
            return;
        }

        dispatch(editorRoles({
            id: row.id,
            name: row.name,
            comment: row.comment,
            metadata: row.metadata,
            source: row.source,
            users: row.users,
            user_groups: row.user_groups,
        }))
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
            resource: 'user',
            data: undefined,
        });
    };

    const handleDelete = (ids: string[]) => {
        dispatch(removeRoles({ state: ids.map(id => ({ id: String(id) })) }))
            .then((res) => {
                openInfoNotification("请求成功", "删除角色成功");
            })
            .catch((err) => {
                openErrNotification("请求失败", "删除角色失败：" + err);
            })
            .finally(() => {
                refreshTables()
            });
    }

    const refreshTables = () => {
        setSearchState(s => ({ ...s, fetchError: false, isLoading: true }));
        fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, searchState.query);
    }

    // 模拟远程请求
    async function fetchData(pageInfo: PageInfo, searchParam?: string) {
        setSearchState(s => ({ ...s, current: pageInfo.current, limit: pageInfo.pageSize, previous: pageInfo.previous, fetchError: false, isLoading: true }));
        try {
            const { current, pageSize } = pageInfo;

            const params = {
                name: searchParam,
            }

            // 默认只查询简要信息
            const response = await describeRoles({
                limit: pageSize, offset: (current - 1) * pageSize, berif: true, ...params
            });
            setSearchState(s => ({ ...s, roles: response.content, total: response.totalCount, isLoading: false }));
        } catch (error: Error | any) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification("获取数据失败", error);
        }
    }

    useEffect(() => {
        refreshTables();
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
                                    <Button theme='danger' onClick={() => {
                                        handleDelete(selectedRowKeys as string[]);
                                    }}>批量删除</Button>
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
            <RoleEditor
                key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                modify={editorState.mode === 'edit'}
                visible={editorState.visible && editorState.resource === 'role'}
                refresh={refreshTables}
                closeDrawer={() => {
                    // 关闭后重置编辑器状态
                    dispatch(resetRoles());
                    setEditorState(s => ({ ...s, visible: false }));
                }} op={editorState.mode} />
            <Table
                data={searchState.roles}
                columns={
                    customColumns(handleEditRole, (id: string, name: string) => {
                        navigate(`roledetail?id=${id}&name=${name}`);
                    })
                }
                loading={searchState.isLoading}
                rowKey="id"
                size={"large"}
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

export default React.memo(RoleTable);