import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Tag, Popconfirm } from 'tdesign-react';
import { DeleteIcon, EditIcon, RefreshIcon, UserVisibleIcon } from 'tdesign-icons-react';
import { useNavigate } from 'react-router-dom';

import Search from 'components/Search';
import ErrorPage from 'components/ErrorPage';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { describeUserToken, User, USER_ROLE_MAP } from 'services/users';
import ShowToken from './ShowToken';
import UserEditor from './UserEditor';
import { editorUser, listUsers, removeUsers, resetUser, selectUser } from 'modules/user/users';
import { Op } from 'services/types';

interface IUsersProps {

}

const ServerError = () => <ErrorPage code={500} />;

const columns = (operateUser: (op: Op, res: string, row?: TableRowData) => void, redirect: (id: string, name: string) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false || row.user_type === 'main' }),
    },
    {
        colKey: 'name',
        title: '用户名',
        cell: ({ row: { name, id } }) => (
            <Link theme='primary' onClick={() => redirect(id, name)}>{name}</Link>
        ),
    },
    {
        colKey: 'user_type',
        title: '用户类型',
        cell: ({ row: { user_type } }: TableRowData) => (
            <Tag theme={USER_ROLE_MAP?.[user_type as keyof typeof USER_ROLE_MAP]?.theme as "success" | "danger" | "default" | "primary" | "warning"} variant="outline">{USER_ROLE_MAP?.[user_type as keyof typeof USER_ROLE_MAP]?.text ?? '-'}</Tag>
        ),

    },
    {
        colKey: 'commnet',
        title: '描述',
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
    },
    {
        colKey: 'token_enable',
        title: 'Token状态',
        cell: ({ row: { token_enable } }: TableRowData) => (<Tag theme={token_enable ? 'success' : 'danger'} variant="outline">{token_enable ? '启用中' : '禁用中'}</Tag>),
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
                    <Tooltip content={row.editable === false ? '无权限操作' : '查看 Token'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => {
                                describeUserToken({ id: row.id }).then((res) => {
                                    if (res) {
                                        operateUser('view', 'user_token', { ...row, auth_token: res.user.auth_token });
                                    }
                                });
                            }}>
                            <UserVisibleIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => operateUser('edit', 'user', { ...row })}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
                    {row.user_type !== 'main' && (
                        <Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
                            <Popconfirm
                                content="确认删除吗"
                                destroyOnClose
                                placement="top"
                                showArrow
                                theme="default"
                                onConfirm={() => {
                                    operateUser('delete', 'user', row);
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

const UsersTable: React.FC<IUsersProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    const userState = useAppSelector(selectUser);
    const { datas, total, page, limit, loading } = userState;

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        resource: string;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, resource: '', mode: 'create', data: undefined });

    // 编辑、新建事件
    const operateUser = (mode: Op, res: string, row?: TableRowData) => {
        switch (mode) {
            case 'delete':
                dispatch(removeUsers({ ids: [row?.id as string] }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification('请求成功', `用户 ${row?.name} 删除成功`);
                        } else {
                            openErrNotification('请求失败', `删除用户 ${row?.name} 失败, ${res.payload as string}`);
                        }
                    });
                return;
            case 'edit':
                dispatch(editorUser({ ...row as User }));
                setEditorState(prev => ({ ...prev, visible: true, mode: mode, resource: res, data: { ...row } }))
                return;
            case 'create':
                setEditorState(prev => ({ ...prev, visible: true, mode: mode, resource: res, data: undefined }));
                return;
            case 'view':
                setEditorState(prev => ({ ...prev, visible: true, mode: mode, resource: res, data: { ...row } }))
        }
    }

    const refreshData = (page = 1, limit = 10, query?: string) => {
        dispatch(listUsers({
            param: {
                limit: limit,
                offset: (page - 1) * limit,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("请求失败", `获取用户列表失败: ${res.payload as string}`);
            }
        })
    }

    React.useEffect(() => {
        refreshData();
    }, []);

    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button onClick={() => {
                                operateUser('create', 'user');
                            }}>新建</Button>
                        </Col>
                    </Row>
                </Col>
                <Col>
                    <Space>
                        <Search
                            onChange={(value: string) => { refreshData(1, limit, value); }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshData(1, limit)} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            <ShowToken
                key={editorState.mode + editorState.data?.name + '_showtoken'}
                row={editorState.data || {} as TableRowData}
                visible={editorState.visible && editorState.resource === 'user_token'}
                close={() => {
                    setEditorState(s => ({ ...s, resource: 'user_token', visible: false }))
                }} />
            {editorState.visible && editorState.resource === 'user' && (
                <UserEditor
                    op={editorState.mode}
                    visible={editorState.visible && editorState.resource === 'user'}
                    closeDrawer={() => {
                        // 关闭后重置编辑器状态
                        dispatch(resetUser());
                        setEditorState(s => ({ ...s, visible: false }));
                        refreshData(1, limit);
                    }}
                />
            )}
            <Table
                data={datas}
                columns={columns(operateUser, (id: string, name: string) => {
                    navigate(`userdetail?id=${id}&name=${name}`);
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
        <>
            {table}
        </>
    )
}

export default React.memo(UsersTable);