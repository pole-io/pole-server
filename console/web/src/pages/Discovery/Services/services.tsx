import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, TableRowData, Popconfirm, Tag } from 'tdesign-react';
import { AddIcon, DeleteIcon, EditIcon, RefreshIcon, CreditcardIcon } from 'tdesign-icons-react';
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

function parseCount(value?: string | number) {
    const parsed = Number(value ?? 0);
    return Number.isFinite(parsed) ? parsed : 0;
}

function visibilityTheme(mode?: string) {
    if (mode === 'all') return 'success';
    if (mode === 'specified') return 'warning';
    return 'default';
}

const hasValue = (value?: string) => value !== undefined && value !== null && value !== '';

const columns = (operateService: (op: Op, row: TableRowData) => void, redirect: (row: TableRowData) => void, t: any): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: t('services.name'),
        fixed: 'left',
        cell: ({ row }) => {
            return (
                <div className={style.serviceCell}>
                    <div className={style.serviceNameRow}>
                        <Link
                            theme="primary"
                            onClick={() => { redirect(row) }}
                        >{row.name}</Link>
                    </div>
                    <div className={style.serviceMeta}>
                        <span>{row.namespace || '-'}</span>
                        {row.comment && <span>{row.comment}</span>}
                    </div>
                </div>
            );
        },
    },
    {
        colKey: 'owner',
        title: '归属',
        cell: ({ row: { department, business } }) => {
            if (!hasValue(department) && !hasValue(business)) return <Text>-</Text>;
            return (
                <div className={style.compactCell}>
                    {hasValue(business) && <Text>{business}</Text>}
                    {hasValue(department) && <span>{department}</span>}
                </div>
            );
        },
    },
    {
        colKey: 'health/total',
        title: t('services.healthTotal'),
        cell: ({ row: { healthy_instance_count, total_instance_count } }) => {
            const healthy = parseCount(healthy_instance_count);
            const total = parseCount(total_instance_count);
            const rate = total > 0 ? Math.round((healthy / total) * 100) : 0;
            return (
                <div className={style.compactCell}>
                    <Text>{`${healthy_instance_count ?? '-'} / ${total_instance_count ?? '-'}`}</Text>
                    <span>{total > 0 ? `${rate}% 健康` : '暂无实例'}</span>
                </div>
            );
        },
    },
    {
        colKey: 'export_to',
        title: t('services.visibility'),
        cell: ({ row }: TableRowData) => {
            const visibilityMode = CheckVisibilityMode(row.export_to, row.namespace);
            const exports = row.export_to || [];
            if (visibilityMode !== 'specified') {
                return (
                    <Tag theme={visibilityTheme(visibilityMode) as any} variant="light">
                        {VisibilityModeMap[visibilityMode]}
                    </Tag>
                );
            }
            return (
                <Space size={4}>
                    {exports.slice(0, 2).map((item: string) => <Tag key={item} variant="outline">{item}</Tag>)}
                    {exports.length > 2 && <Tag variant="outline">+{exports.length - 2}</Tag>}
                </Space>
            );
        },
    },
    {
        colKey: 'time',
        title: t('services.time'),
        cell: ({ row: { ctime, mtime } }: TableRowData) => {
            if (!hasValue(ctime) && !hasValue(mtime)) return <Text>-</Text>;
            return (
                <div className={style.compactCell}>
                    <Text>{mtime || '-'}</Text>
                    {hasValue(ctime) && <span>{t('services.create')}: {ctime}</span>}
                </div>
            );
        },
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

    const metric = useMemo(() => {
        const namespaces = new Set(datas.map((item) => item.namespace).filter(Boolean));
        const healthy = datas.reduce((sum, item) => sum + parseCount(item.healthy_instance_count), 0);
        const instances = datas.reduce((sum, item) => sum + parseCount(item.total_instance_count), 0);
        const publicVisible = datas.filter((item) => CheckVisibilityMode(item.export_to, item.namespace) === 'all').length;
        return {
            namespaces,
            healthy,
            instances,
            publicVisible,
        };
    }, [datas]);

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });
    const [query, setQuery] = useState({ name: '' });

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

    const refreshTable = (page = 1, limit = 10, nextQuery = query) => {
        dispatch(listServices({
            param: {
                offset: (page - 1) * limit,
                limit: limit,
                name: nextQuery.name || undefined,
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
            <section className={style.metricRail}>
                <div className={style.metricItem}>
                    <span>Services</span>
                    <strong>{total}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>Namespaces</span>
                    <strong>{metric.namespaces.size}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>Healthy Instances</span>
                    <strong>{metric.healthy}/{metric.instances}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>Public Visible</span>
                    <strong>{metric.publicVisible}</strong>
                </div>
            </section>
            <section className={style.filterBar}>
                <div className={style.filterHint}>
                    <strong>服务清单</strong>
                    <span>{loading ? '正在同步列表' : `当前显示 ${datas.length} 条`}</span>
                </div>
                <Space>
                    <Search
                        placeholder="搜索服务名"
                        onChange={(value: string) => {
                            const nextQuery = { name: value };
                            setQuery(nextQuery);
                            refreshTable(1, limit, nextQuery);
                        }}
                    />
                    <Tooltip content={t('common.refresh')}>
                        <Button shape="square" variant="outline" onClick={() => refreshTable(page, limit, query)}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={() => operateService('create')}>{t('common.add')}</Button>
                </Space>
            </section>
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
            <section className={style.tableSurface}>
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
                            refreshTable(pageInfo.current, pageInfo.pageSize, query);
                        },
                    }}
                    onPageChange={(pageInfo) => {
                        refreshTable(pageInfo.current, pageInfo.pageSize, query);
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

export default React.memo(ServicesTable);
