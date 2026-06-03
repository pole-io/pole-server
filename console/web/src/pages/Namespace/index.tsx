import React, { useState } from 'react';
import { Table, Popup, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Popconfirm } from 'tdesign-react';
import { DeleteIcon, EditIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification } from 'utils/notifition';
import Text from 'components/Text';
import { CheckVisibilityMode, VisibilityModeMap } from 'utils/visible';
import NamespaceEditor from './NamespaceEditor';
import Search from 'components/Search';
import style from './index.module.less';
import { cleanNamespacePage, editorNamespace, listNamespaces, resetNamespace, selectNamespace } from 'modules/namespace';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { Namespace } from 'services/namespace';
import { Op } from 'services/types';

const columns = (operateNamespace: (op: Op, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }: TableRowData) => <Text>{row.name}</Text>,
        fixed: 'left',
    },
    {
        colKey: 'service_export_to',
        title: '服务可见性',
        cell: ({ row: { name, service_export_to } }: TableRowData) => {
            const visibilityMode = CheckVisibilityMode(service_export_to, name)
            return (
                <div>
                    {visibilityMode ? (
                        VisibilityModeMap[visibilityMode]
                    ) : (
                        <Popup
                            trigger={'hover'}
                            content={
                                <Text>
                                    <div>{'服务可见的命名空间列表'}</div>
                                    {service_export_to?.map((item: string) => (
                                        <div key={item}>
                                            {item}
                                        </div>
                                    ))}
                                </Text>
                            }
                        >
                            <Text>{service_export_to ? service_export_to?.join(',') : '-'}</Text>
                        </Popup>
                    )}
                </div>
            )
        },
    },

    {
        colKey: 'commnet',
        title: '描述',
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
    },
    {
        colKey: 'totalSerivce',
        title: '服务数',
        cell: ({ row }: TableRowData) => <Text>{row.total_service_count ?? '-'}</Text>,
    },
    {
        colKey: 'health/total',
        title: '健康实例/总实例数',
        cell: ({ row: { total_instance_count, total_health_instance_count } }: TableRowData) => (
            <Text>{`${total_health_instance_count}/${total_instance_count}`}</Text>
        ),
    },
    {
        colKey: 'time',
        title: '操作时间',
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>修改: {mtime}<br />创建: {ctime}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        cell: ({ row }: TableRowData) => {
            return (
                <Space>
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => operateNamespace('edit', row)}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={row.editable === false ? '无权限操作' : '授权'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => { operateNamespace('authorize', row) }}>
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
                                operateNamespace('delete', row);
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

export default React.memo(() => {
    const dispatch = useAppDispatch();

    const namespaceState = useAppSelector(selectNamespace);
    const { datas, loading, page, limit, total } = namespaceState;

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        mode: 'create' | 'edit';
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
            case 'create':
                setEditorState(prev => ({ ...prev, visible: true, mode: 'create', data: undefined }));
                break;
            case 'authorize':
                setEditorState(prev => ({ ...prev, authorizeVisible: true, data: { ...row } }));
                break;
            case 'delete':
                break;
        }
    }

    const refreshTable = (page = 1, limit = 10, query = '') => {
        dispatch(listNamespaces({
            param: {
                offset: (page - 1) * limit,
                limit: limit,

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

    {/* <!-- :defaultExpandedRowKeys="defaultExpandedRowKeys" --> */ }
    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button onClick={() => operateNamespace('create')}>新建</Button>
                        </Col>
                    </Row>
                </Col>
                <Col>
                    <Space>
                        <Search
                            onChange={(value: string) => {
                                refreshTable();
                            }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshTable()} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editorState.visible && (
                <NamespaceEditor
                    key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
                    op={editorState.mode}
                    visible={editorState.visible}
                    closeDrawer={() => {
                        // 清理编辑器状态
                        dispatch(resetNamespace());
                        setEditorState(s => ({ ...s, visible: false }))
                        refreshTable();
                    }} />
            )}
            {editorState.authorizeVisible && (
                <AuthorizeInput
                    resource_type={PolicySourceType.Namespaces}
                    resource_id={editorState.data?.id}
                    resource_name={`${editorState.data?.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
            <Table
                data={datas}
                columns={columns(operateNamespace)}
                loading={loading}
                rowKey="name"
                size={"large"}
                tableLayout={'auto'}
                cellEmptyContent={'-'}
                pagination={{
                    current: page,
                    pageSize: limit,
                    total: total,
                    showJumper: true,
                    onChange(pageInfo) {
                        refreshTable(pageInfo.current, pageInfo.pageSize);
                    },
                }}
                onPageChange={(pageInfo) => {
                    refreshTable(pageInfo.current, pageInfo.pageSize);
                }}
            />
        </>
    );

    return (
        <div>
            {table}
        </div>
    )
});