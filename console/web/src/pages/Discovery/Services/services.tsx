import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, Popup, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Popconfirm } from 'tdesign-react';
import { DeleteIcon, EditIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';
import { useNavigate } from 'react-router-dom';

import Search from 'components/Search';
import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { Service } from 'services/service';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { CheckVisibilityMode, VisibilityModeMap } from 'utils/visible';
import ServiceEditor from './ServiceEditor';
import style from './index.module.less';
import { cleanServicePage, editorService, listServices, removeServices, resetService, selectService } from 'modules/discovery/service';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { Op } from 'services/types';


const columns = (operateService: (op: Op, row: TableRowData) => void, redirect: (row: TableRowData) => void, t: any): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: t('services.name'),
        cell: ({ row }) => <Link
            theme="primary"
            onClick={() => { redirect(row) }}
        >{row.name}</Link>,
    },
    {
        colKey: 'namespace',
        title: t('services.namespace'),
        cell: ({ row: { namespace } }) => <Text>{namespace}</Text>,
    },
    {
        colKey: 'export_to',
        title: t('services.visibility'),
        cell: ({ row: { name, export_to } }: TableRowData) => {
            const visibilityMode = CheckVisibilityMode(export_to, name)
            return (
                <div>
                    {visibilityMode ? (
                        VisibilityModeMap[visibilityMode]
                    ) : (
                        <Popup
                            trigger={'hover'}
                            content={
                                <Text>
                                    <div>{t('services.visibilityList')}</div>
                                    {export_to?.map((item: string) => (
                                        <div key={item}>
                                            {item}
                                        </div>
                                    ))}
                                </Text>
                            }
                        >
                            <Text>{export_to ? export_to?.join(',') : '-'}</Text>
                        </Popup>
                    )}
                </div>
            )
        },
    },
    {
        colKey: 'department',
        title: t('services.department'),
        cell: ({ row: { department } }) => <Text>{department || '-'} </Text>,
    },
    {
        colKey: 'business',
        title: t('services.business'),
        cell: ({ row: { business } }) => <Text>{business || '-'} </Text>,
    },
    {
        colKey: 'health/total',
        title: t('services.healthTotal'),
        cell: ({ row: { healthy_instance_count, total_instance_count } }) => (
            <Text>
                {`${healthy_instance_count ?? '-'} / ${total_instance_count ?? '-'}`}
            </Text>
        ),
    },
    {
        colKey: 'commnet',
        title: t('services.comment'),
        ellipsis: true,
        cell: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
    },
    {
        colKey: 'time',
        title: t('services.time'),
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>{t('services.modify')}: {mtime}<br />{t('services.create')}: {ctime}</Text>,
    },
    {
        colKey: 'action',
        title: t('common.action'),
        cell: ({ row }) => {
            return (
                <Space>
                    <Tooltip content={row.editable === false ? t('services.noPermission') : t('common.edit')}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => operateService('edit', row)}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={row.editable === false ? t('services.noPermission') : t('services.authorize')}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => operateService('authorize', row)}>
                            <CreditcardIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={row.deleteable === false ? t('services.noPermission') : t('common.delete')}>
                        <Popconfirm
                            content={t('services.confirmDelete')}
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

interface IServicesProps {

}

const ServicesTable: React.FC<IServicesProps> = ({ }) => {
    const { t } = useTranslation();
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    const { datas, loading, total, page, limit } = useAppSelector(selectService);

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });

    // 编辑、新建事件
    const operateService = (op: Op, row?: TableRowData) => {
        switch (op) {
            case 'edit':
                dispatch(editorService({ ...row as Service }));
                setEditorState(prev => ({ ...prev, visible: true, mode: op, data: row }));
                break;
            case 'create':
                dispatch(resetService());
                setEditorState(prev => ({ ...prev, visible: true, mode: op, data: undefined }));
                break;
            case 'authorize':
                setEditorState(prev => ({ ...prev, authorizeVisible: true, data: { ...row } }));
                break;
            case 'delete':
                dispatch(removeServices({ ids: [row?.id as string] }))
                    .then((res) => {
                        if (res.meta.requestStatus === 'fulfilled') {
                            openInfoNotification(t('common.success'), t('services.deleteSuccess'));
                            // 刷新表格
                            refreshTable();
                        } else {
                            openErrNotification(t('common.fail'), res.payload as string);
                        }
                    })
                break;
            case 'view':
                // 跳转到实例列表
                dispatch(editorService({ ...row as Service }));
                navigate(`instance?namespace=${row?.namespace}&service=${row?.name}`);
                break;
            default:
                openErrNotification(t('common.fail'), t('services.unknownOp'));
                return;
        }
    }

    React.useEffect(() => {
        refreshTable();
        return () => {
            // 清理编辑器状态
            dispatch(cleanServicePage());
        }
    }, []);

    const refreshTable = (page = 1, limit = 10) => {
        dispatch(listServices({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("获取数据失败", res.payload as string);
            }
        });
    }

    {/* <!-- :defaultExpandedRowKeys="defaultExpandedRowKeys" --> */ }
    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button onClick={() => operateService('create')}>{t('common.add')}</Button>
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
                        <Tooltip content={t('common.refresh')}>
                            <RefreshIcon onClick={() => refreshTable()} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editorState.visible && (
                <ServiceEditor
                    op={editorState.mode}
                    visible={editorState.visible}
                    closeDrawer={() => {
                        // 关闭后重置编辑器状态
                        dispatch(resetService());
                        refreshTable();
                        setEditorState(s => ({ ...s, visible: false }));
                    }} />
            )}
            {editorState.authorizeVisible && (
                <AuthorizeInput
                    resource_type={PolicySourceType.Services}
                    resource_id={editorState.data?.id}
                    resource_name={`${editorState.data?.namespace}/${editorState.data?.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
            <Table
                data={datas}
                columns={columns(operateService, (row: TableRowData) => {
                    operateService('view', row);
                }, t)}
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
            />
        </>
    );

    return (
        <div>
            {table}
        </div>
    )
}

export default React.memo(ServicesTable);