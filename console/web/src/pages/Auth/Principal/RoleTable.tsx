import React, { useEffect, useState } from 'react';
import { Button, PageInfo, PrimaryTableProps, Space, Table, TableProps, TableRowData, Tooltip } from 'components/Fluent';
import { AddIcon, RefreshIcon } from 'components/Fluent/icons';
import { useNavigate } from 'react-router-dom';

import Search from 'components/Search';
import ErrorPage from 'components/ErrorPage';
import Text from 'components/Text';
import { ConfirmOperationButton, OperationButton } from 'components/OperationButton';
import ResourceNameLink from 'components/ResourceNameLink';
import { ResourceToolbar } from 'components/ResourceLayout';
import { useAppDispatch } from 'modules/store';
import { removeRoles } from 'modules/auth/role';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { describeRoles, Role } from 'services/role';
import RoleEditor from './RoleEditor';

type Op = 'create' | 'view' | 'edit';

const ServerError = () => <ErrorPage code={500} />;
const isBuiltInRole = (role: Partial<Role>) => role.default_role === true;

const columns = (
    openRoleDetail: (row: TableRowData) => void,
    openEditor: (row: Role, mode: Op) => void,
    deleteRole: (role: Role) => void,
): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: isBuiltInRole(row as Role) }),
    },
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }) => <ResourceNameLink name={row.name} onClick={() => openRoleDetail(row)} />,
    },
    {
        colKey: 'source',
        title: '类型',
        cell: ({ row }) => <Text>{isBuiltInRole(row as Role) ? '系统内置' : '自定义'}</Text>,
    },
    {
        title: '描述',
        colKey: 'comment',
        ellipsis: true,
        cell: ({ row }) => <Text>{row.comment || '-'}</Text>,
    },
    {
        colKey: 'time',
        title: '操作时间',
        cell: ({ row }) => <Text>修改: {row.mtime}<br />创建: {row.ctime}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        cell: ({ row }) => {
            const role = row as Role;
            return (
                <Space>
                    <OperationButton action="view" onClick={() => openRoleDetail(row)} />
                    {isBuiltInRole(role) ? (
                        <Button variant="transparent" aria-label="管理成员" onClick={() => openEditor(role, 'edit')}>管理成员</Button>
                    ) : (
                        <>
                            <OperationButton action="edit" onClick={() => openEditor(role, 'edit')} />
                            <ConfirmOperationButton action="delete" confirmContent="确认删除该自定义角色吗" onConfirm={() => deleteRole(role)} />
                        </>
                    )}
                </Space>
            );
        },
    },
];

const RoleTable: React.FC = () => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const [selectedRowKeys, setSelectedRowKeys] = useState<Array<string | number>>([]);
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
    const [editorState, setEditorState] = useState<{ visible: boolean; mode: Op; role: Role | null }>({
        visible: false,
        mode: 'create',
        role: null,
    });

    const openRoleDetail = (row: TableRowData) => {
        navigate(`/auth/principals/roledetail?name=${encodeURIComponent(String(row.name || ''))}&id=${encodeURIComponent(String(row.id || ''))}`);
    };

    async function fetchData(pageInfo: PageInfo, searchParam = '') {
        setSearchState(s => ({ ...s, current: pageInfo.current, limit: pageInfo.pageSize, previous: pageInfo.previous, query: searchParam, fetchError: false, isLoading: true }));
        try {
            const response = await describeRoles({
                limit: pageInfo.pageSize,
                offset: (pageInfo.current - 1) * pageInfo.pageSize,
                berif: true,
                ...(searchParam && { name: searchParam }),
            });
            setSearchState(s => ({ ...s, roles: response.content, total: response.totalCount, isLoading: false }));
        } catch (error: Error | any) {
            setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
            openErrNotification('获取数据失败', error);
        }
    }

    const refreshTables = () => {
        setSelectedRowKeys([]);
        fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, searchState.query);
    };

    const deleteRolesByIds = async (ids: string[]) => {
        const result = await dispatch(removeRoles({ state: ids.map(id => ({ id })) }));
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求失败', String(result.payload || '删除角色失败'));
            return;
        }
        openInfoNotification('请求成功', ids.length > 1 ? '批量删除角色成功' : '删除角色成功');
        refreshTables();
    };

    useEffect(() => {
        fetchData({ current: 1, pageSize: searchState.limit, previous: 0 });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const table = (
        <>
            <ResourceToolbar
                className={style.toolBar}
                title="角色列表"
                count={searchState.isLoading ? '正在同步列表' : `共 ${searchState.total} 条`}
                filters={(
                    <>
                        <Search onChange={(value: string) => fetchData({ current: 1, pageSize: searchState.limit, previous: 0 }, value)} />
                        <Tooltip content="刷新">
                            <Button shape="square" variant="outline" onClick={refreshTables}><RefreshIcon /></Button>
                        </Tooltip>
                        {selectedRowKeys.length > 0 && (
                            <>
                                <span className={style.selectionHint}>已选 {selectedRowKeys.length} 项</span>
                                <Button theme="danger" onClick={() => deleteRolesByIds(selectedRowKeys.map(String))}>批量删除</Button>
                            </>
                        )}
                        <Button theme="primary" icon={<AddIcon />} onClick={() => setEditorState({ visible: true, mode: 'create', role: null })}>新建角色</Button>
                    </>
                )}
            />
            <RoleEditor
                key={`${editorState.mode}-${editorState.role?.id || 'new'}-${editorState.visible}`}
                visible={editorState.visible}
                op={editorState.mode}
                role={editorState.role}
                refresh={refreshTables}
                closeDrawer={() => setEditorState(s => ({ ...s, visible: false }))}
            />
            <Table
                data={searchState.roles}
                columns={columns(openRoleDetail, (role, mode) => setEditorState({ visible: true, mode, role }), role => deleteRolesByIds([role.id]))}
                loading={searchState.isLoading}
                rowKey="id"
                size="large"
                tableLayout="auto"
                cellEmptyContent="-"
                pagination={{
                    current: searchState.current,
                    pageSize: searchState.limit,
                    total: searchState.total,
                    showJumper: true,
                    onChange: pageInfo => fetchData(pageInfo, searchState.query),
                }}
                onPageChange={pageInfo => fetchData(pageInfo, searchState.query)}
                selectOnRowClick={false}
                selectedRowKeys={selectedRowKeys}
                onSelectChange={(selected: Array<string | number>) => setSelectedRowKeys(selected)}
            />
        </>
    );

    return searchState.fetchError ? <ServerError /> : table;
};

export default React.memo(RoleTable);
