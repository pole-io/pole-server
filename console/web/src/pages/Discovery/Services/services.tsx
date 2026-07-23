import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Table, Button, PrimaryTableProps, TableRowData, Input, Select } from 'components/Fluent';
import { SearchIcon } from 'components/Fluent/icons';
import { useNavigate } from 'react-router-dom';

import Text from 'components/Text';
import { ConfirmOperationButton, OperationButton, OperationButtonGroup } from 'components/OperationButton';
import { ResourceToolbar } from 'components/ResourceLayout';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { NamespaceView } from 'services/namespace';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import ServiceEditor from './ServiceEditor';
import style from './index.module.less';
import { cleanServicePage, listServices, removeServices, resetService, selectService } from 'modules/discovery/service';
import { listAllNamespaces, selectNamespace } from 'modules/namespace';
import { Op } from 'services/types';
import ResourceNameLink from 'components/ResourceNameLink';

function parseCount(value?: string | number) {
    const parsed = Number(value ?? 0);
    return Number.isFinite(parsed) ? parsed : 0;
}

const hasValue = (value?: string) => value !== undefined && value !== null && value !== '';

const columns = (operateService: (op: Op, row: TableRowData) => void, redirect: (row: TableRowData) => void, t: any): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: t('services.name'),
        width: 260,
        fixed: 'left',
        cell: ({ row }) => <ResourceNameLink name={row.name} onClick={() => redirect(row)} />,
    },
    {
        colKey: 'namespace',
        title: '命名空间',
        width: 180,
        cell: ({ row: { namespace } }) => <Text>{namespace || '-'}</Text>,
    },
    {
        colKey: 'owner',
        title: '归属',
        width: 150,
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
        width: 150,
        cell: ({ row: { healthy_instance_count, total_instance_count } }) => {
            const totalInstances = parseCount(total_instance_count);
            const healthyInstances = parseCount(healthy_instance_count);
            const percent = totalInstances > 0 ? Math.round((healthyInstances / totalInstances) * 100) : 0;
            return (
                <div className={style.healthCell}>
                    <div>
                        <strong>{`${healthyInstances}/${totalInstances}`}</strong>
                        <span>{totalInstances > 0 ? `${percent}%` : '无实例'}</span>
                    </div>
                    <div className={style.healthTrack}>
                        <i style={{ width: `${percent}%` }} />
                    </div>
                </div>
            );
        },
    },
    {
        colKey: 'time',
        title: t('services.time'),
        width: 180,
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
        width: 86,
        fixed: 'right',
        cell: ({ row }) => {
            return (
                <OperationButtonGroup className={style.actionCell}>
                    <OperationButton action="viewEdit" disabled={row.editable === false} disabledLabel={t('services.noPermission')} onClick={() => operateService('view', row)} />
                    <ConfirmOperationButton action="delete" disabled={row.deleteable === false} disabledLabel={t('services.noPermission')} label={t('common.delete')} confirmContent={t('services.confirmDelete')} onConfirm={() => operateService('delete', row)} />
                </OperationButtonGroup>
            )
        },
    },
]

interface IServicesProps {

}

export interface ServicesTableHandle {
    refresh: () => void;
    create: () => void;
}

const ServicesTable = React.forwardRef<ServicesTableHandle, IServicesProps>((_, ref) => {
    const { t } = useTranslation();
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    const { datas, loading, total, page, limit } = useAppSelector(selectService);
    const { datas: namespaceDatas } = useAppSelector(selectNamespace);

    const metric = useMemo(() => {
        const namespaces = new Set(datas.map((item) => item.namespace).filter(Boolean));
        const healthy = datas.reduce((sum, item) => sum + parseCount(item.healthy_instance_count), 0);
        const instances = datas.reduce((sum, item) => sum + parseCount(item.total_instance_count), 0);
        return {
            namespaces,
            healthy,
            instances,
        };
    }, [datas]);

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        mode: Op;
        data?: TableRowData;
    }>({ visible: false, mode: 'create', data: undefined });
    const [query, setQuery] = useState({ namespace: '', name: '' });

    const openServiceDetail = (row?: TableRowData, edit = false) => {
        if (!row?.namespace || !row?.name) {
            openErrNotification(t('common.fail'), '服务缺少命名空间或名称，无法打开详情');
            return;
        }
        const params = new URLSearchParams({
            namespace: String(row.namespace),
            service: String(row.name),
        });
        if (edit) {
            params.set('mode', 'edit');
        }
        navigate(`instance?${params.toString()}`);
    };

    // 编辑、新建事件
    const operateService = (op: Op, row?: TableRowData) => {
        switch (op) {
            case 'edit':
                openServiceDetail(row, true);
                break;
            case 'create':
                dispatch(resetService());
                setEditorState(prev => ({ ...prev, visible: true, mode: op, data: undefined }));
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
                openServiceDetail(row, true);
                break;
            default:
                openErrNotification(t('common.fail'), t('services.unknownOp'));
                return;
        }
    }

    React.useEffect(() => {
        refreshTable();
        dispatch(listAllNamespaces())
            .then((res) => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('获取命名空间列表失败', res.payload as string);
                }
            });
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
                namespace: nextQuery.namespace || undefined,
                name: nextQuery.name || undefined,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification("获取数据失败", res.payload as string);
            }
        });
    }

    React.useImperativeHandle(ref, () => ({
        refresh: () => refreshTable(page, limit, query),
        create: () => operateService('create'),
    }));

    const submitFilter = () => {
        refreshTable(1, limit, query);
    };

    const resetFilter = () => {
        const nextQuery = { namespace: '', name: '' };
        setQuery(nextQuery);
        refreshTable(1, limit, nextQuery);
    };

    {/* <!-- :defaultExpandedRowKeys="defaultExpandedRowKeys" --> */ }
    const table = (
        <>
            <section className={style.metricRail}>
                <div className={style.metricItem}>
                    <span>服务数</span>
                    <strong id="stSvc">{total}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>命名空间</span>
                    <strong id="stNs">{metric.namespaces.size}</strong>
                </div>
                <div className={style.metricItem}>
                    <span>健康实例</span>
                    <strong className={style.metricValue}>
                        <b id="stHealthy">{metric.healthy}</b>
                        /
                        <b id="stInst">{metric.instances}</b>
                    </strong>
                </div>
            </section>
            <section className={style.listSection}>
                <ResourceToolbar
                    title="服务清单"
                    count={<span id="listCount">{loading ? '正在同步列表' : `当前显示 ${datas.length} 条`}</span>}
                    filters={(
                        <>
                        <div id="nsFilter" className={style.namespaceFilter}>
                            <Select
                                clearable
                                filterable
                                placeholder="全部命名空间"
                                value={query.namespace}
                                options={namespaceDatas.map((item: NamespaceView) => ({
                                    label: item.name,
                                    value: item.name,
                                }))}
                                onChange={(value) => setQuery(prev => ({ ...prev, namespace: String(value || '') }))}
                            />
                        </div>
                        <div id="keyword" className={style.filterInput}>
                            <Input
                                clearable
                                prefixIcon={<SearchIcon />}
                                placeholder="服务名"
                                value={query.name}
                                onChange={(value) => setQuery(prev => ({ ...prev, name: String(value) }))}
                                onEnter={submitFilter}
                            />
                        </div>
                        <Button variant="outline" onClick={submitFilter}>查询</Button>
                        <Button variant="text" onClick={resetFilter}>重置</Button>
                        </>
                    )}
                />
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
                <section id="tbody" className={`${style.tableSurface} ${style.serviceTableSurface}`}>
                    <Table
                        data={datas}
                        columns={columns(operateService, (row: TableRowData) => {
                            openServiceDetail(row, false);
                        }, t)}
                        loading={loading}
                        rowKey="id"
                        size={"large"}
                        tableLayout="fixed"
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
            </section>
        </>
    );

    return (
        <div className={style.workspace}>
            {table}
        </div>
    )
});

export default React.memo(ServicesTable);
