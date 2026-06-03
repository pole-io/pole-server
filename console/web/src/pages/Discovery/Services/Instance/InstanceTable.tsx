import React, { useState, useEffect } from 'react';
import { Popup, Table, Button, PageInfo, PrimaryTableProps, TableProps, Tooltip, Space, Row, Col, TableRowData, Tag, Breadcrumb, Link, Loading, Popconfirm } from 'tdesign-react';
import { DeleteIcon, EditIcon, RefreshIcon } from 'tdesign-icons-react';
import { useNavigate, BrowserRouterProps } from 'react-router-dom';

import Search from 'components/Search';
import ErrorPage from 'components/ErrorPage';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { describeInstances, HEALTH_STATUS_MAP, Instance, ISOLATE_STATUS_MAP } from 'services/instance';
import { openErrNotification } from 'utils/notifition';
import style from './index.module.less';
import InstanceEditor from './InstanceEditor';
import { cleanInsPage, editorInstance, listInstances, resetInstance, selectInstance } from 'modules/discovery/instance';
import { Op } from 'services/types';

export interface IInstanceListProps {
    namespace: string;
    serviceName: string;
}

const ServerError = () => <ErrorPage code={500} />;

const columns = (operateInstance: (op: Op, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: 'ID',
        type: 'multiple',
        checkProps: ({ row }) => ({ disabled: row.editable === false || row.deleteable === false }),
    },
    {
        colKey: 'host',
        title: '主机',
        cell: ({ row }) => (
            <Link theme='primary' onClick={() => operateInstance('view', row)}>
                <Text>{row.host}</Text>
            </Link>
        ),
    },
    {
        colKey: 'port',
        title: '端口',
        cell: ({ row: { port } }) => <Text>{port || '-'}</Text>,
    },
    {
        colKey: 'protocol',
        title: '协议',
        cell: ({ row: { protocol } }) => <Text>{protocol || '-'}</Text>,
    },
    {
        colKey: 'version',
        title: '版本',
        cell: ({ row: { version } }) => (
            <Text>{version || '-'}</Text>
        ),
    },
    {
        colKey: 'weight',
        title: '权重',
        cell: ({ row: { weight } }) => (
            <Text>{weight || '-'}</Text>
        ),
    },
    {
        colKey: 'healthy',
        title: '健康状态',
        cell: ({ row: { healthy } }) => (
            <Tag theme={HEALTH_STATUS_MAP?.[healthy as keyof typeof HEALTH_STATUS_MAP]?.theme as "success" | "danger" | "default" | "primary" | "warning"} variant="outline">{HEALTH_STATUS_MAP?.[healthy as keyof typeof HEALTH_STATUS_MAP]?.text ?? '-'}</Tag>
        ),
    },
    {
        colKey: 'isolate',
        title: '隔离状态',
        cell: ({ row: { isolate } }) => (
            <Tag theme={ISOLATE_STATUS_MAP?.[isolate as keyof typeof ISOLATE_STATUS_MAP]?.theme as "success" | "danger" | "default" | "primary" | "warning"} variant="outline">{ISOLATE_STATUS_MAP?.[isolate as keyof typeof ISOLATE_STATUS_MAP]?.text ?? '-'}</Tag>
        ),
    },
    {
        colKey: 'location',
        title: '地理位置',
        cell: ({ row: { location } }) => (
            <Text>{location?.region || '-'}/{location?.zone || '-'}/{location?.campus || '-'}</Text>
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
        cell: ({ row }) => {
            return (
                <Space>
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => operateInstance('edit', row)}>
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
                                operateInstance('delete', row);
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

export default React.memo((props: IInstanceListProps & BrowserRouterProps) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate()

    const [namespace, serviceName] = [props.namespace, props.serviceName];
    const [selectedRowKeys, setSelectedRowKeys] = useState<Array<string | number>>([]);

    const insState = useAppSelector(selectInstance);
    const { datas, loading, total, page, limit } = insState;

    // 合并编辑相关状态
    const [editState, setEditState] = useState<{
        selectedRow?: TableRowData;
        visible: boolean;
        mode: 'create' | 'edit' | 'view' | 'delete';
    }>({ selectedRow: undefined, visible: false, mode: 'create' });

    React.useEffect(() => {
        refreshTable();
        return () => {
            dispatch(cleanInsPage());
        }
    }, []);

    const refreshTable = (page = 1, limit = 10, query = '') => {
        dispatch(listInstances({
            param: {
                namespace: namespace,
                service: serviceName,
                offset: (page - 1) * limit,
                limit: limit,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取实例列表失败', res.payload as string);
            }
        });
    }

    // 编辑、新建事件
    const operateInstance = (op: Op, row?: TableRowData) => {
        switch (op) {
            case 'create':
                dispatch(resetInstance());
                setEditState(prev => ({ ...prev, visible: true, mode: 'create', selectedRow: undefined }));
                break;
            case 'edit':
                dispatch(editorInstance({ ...row as Instance }));
                setEditState(prev => ({ ...prev, visible: true, mode: op, selectedRow: row }));
                break;
            case 'delete':
                dispatch(editorInstance({ ...row as Instance }));
                setEditState(prev => ({ ...prev, visible: true, mode: 'delete', selectedRow: row }));
                break;
            case 'view':
                dispatch(editorInstance({ ...row as Instance }));
                setEditState(prev => ({ ...prev, visible: true, mode: 'view', selectedRow: row }));
                break;
            default:
                break;
        }
    }

    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button onClick={() => operateInstance('create')}>新建</Button>
                        </Col>
                        {selectedRowKeys.length > 0 && (
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
                                refreshTable(1, limit, value);
                            }} />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshTable(1, limit)} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editState.visible && (
                <InstanceEditor
                    key={editState.mode + (editState.selectedRow?.host || 'new') + (editState.visible ? '1' : '0')}
                    op={editState.mode}
                    namespace={namespace}
                    service={serviceName}
                    visible={editState.visible}
                    closeDrawer={() => {
                        // 关闭后重置编辑器状态
                        dispatch(resetInstance());
                        setEditState(s => ({ ...s, visible: false }));
                        refreshTable();
                    }} />
            )}
            <Table
                data={datas}
                columns={columns(operateInstance)}
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
                        refreshTable(pageInfo.current, pageInfo.pageSize);
                    },
                }}
                onPageChange={(pageInfo) => {
                    refreshTable(pageInfo.current, pageInfo.pageSize);
                }}
                selectOnRowClick={false}
                selectedRowKeys={selectedRowKeys}
                onSelectChange={(selected: Array<string | number>) => {
                    setSelectedRowKeys(selected);
                }}
            />
        </>
    );

    return (
        <div style={{ margin: 20 }}>
            {table}
        </div>
    );
});