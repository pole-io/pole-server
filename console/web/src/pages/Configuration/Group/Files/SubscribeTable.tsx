import React from 'react';
import { Button, Descriptions, Drawer, Table, Tooltip, Tag } from 'components/Fluent';
import type { PrimaryTableProps } from 'components/Fluent';
import { BrowseIcon } from 'components/Fluent/icons';

import { useAppDispatch, useAppSelector } from 'modules/store';
import Text from 'components/Text';
import { FileSubscriber, VersionClient } from 'services/config_release';
import { openErrNotification } from 'utils/notifition';

import style from './index.module.less';
import { cleanFilePage, listFileSubscribers, selectConfigFile } from 'modules/configuration/file';
import { cleanFileReleasePage, listConfigFileReleases, selectFileRelease } from 'modules/configuration/release';

interface ISubscribeTableProps {
    namespace: string;
    group: string;
    filename: string;
    editable: boolean;
    deleteable: boolean;
}

type SubscriberRow = FileSubscriber & {
    resolvedStatus?: 'gray' | 'normal' | 'missing';
    resolvedText?: string;
}

const { DescriptionsItem } = Descriptions;

const columns = (onView: (row: SubscriberRow) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: '客户端ID',
        cell: ({ row: { id } }) => <Text>{id}</Text>,
    },
    {
        colKey: 'host',
        title: '客户端IP',
        cell: ({ row: { host } }) => <Text>{host}</Text>,
    },
    {
        colKey: 'client_type',
        title: '类型',
        cell: ({ row: { client_type } }) => <Text>{client_type}</Text>,
    },
    {
        colKey: 'labels',
        title: '标签',
        cell: () => '-',
    },
    {
        colKey: 'listeningVersion',
        title: '监听版本',
        cell: ({ row }) => {
            const status = row.resolvedStatus as SubscriberRow['resolvedStatus'];
            const theme = status === 'gray' ? 'warning' : status === 'normal' ? 'success' : 'danger';
            return (
                <Tag theme={theme} variant="light">
                    {row.resolvedText || `v${row.version || '-'}`}
                </Tag>
            );
        },
    },
    {
        colKey: 'lastPull',
        title: '最近拉取',
        cell: () => '-',
    },
    {
        colKey: 'action',
        title: '操作',
        fixed: 'right',
        width: 72,
        cell: ({ row }) => {
            return (
                <Tooltip content="查看详情">
                    <Button aria-label="查看订阅客户端详情" shape="square" variant="text" onClick={() => onView(row as SubscriberRow)}>
                        <BrowseIcon />
                    </Button>
                </Tooltip>
            )
        },
    },
]

const SubscribeTable: React.FC<ISubscribeTableProps> = (props) => {
    const dispatch = useAppDispatch();

    const fileState = useAppSelector(selectConfigFile);
    const { subscribers } = fileState;
    const releaseState = useAppSelector(selectFileRelease);
    const { versions } = releaseState;
    const [loading, setLoading] = React.useState(false);
    const [selectedSubscriber, setSelectedSubscriber] = React.useState<SubscriberRow | null>(null);

    React.useEffect(() => {
        let active = true;
        dispatch(cleanFilePage());
        dispatch(cleanFileReleasePage());
        setSelectedSubscriber(null);
        if (props.namespace && props.group && props.filename) {
            setLoading(true);
            Promise.all([
                dispatch(listFileSubscribers({
                    param: {
                        namespace: props.namespace,
                        group: props.group,
                        file_name: props.filename,
                    }
                })),
                dispatch(listConfigFileReleases({
                    param: {
                        namespace: props.namespace,
                        group: props.group,
                        file_name: props.filename,
                    }
                })),
            ]).then(([subscriberResult, releaseResult]) => {
                if (!active) return;
                if (subscriberResult.meta.requestStatus === 'rejected') {
                    openErrNotification('请求失败', String(subscriberResult.payload || '获取订阅客户端失败'));
                }
                if (releaseResult.meta.requestStatus === 'rejected') {
                    openErrNotification('获取发布记录失败', String(releaseResult.payload || '请求失败'));
                }
            }).finally(() => {
                if (active) setLoading(false);
            });
        }
        return () => {
            active = false;
            dispatch(cleanFilePage());
            dispatch(cleanFileReleasePage());
        }
    }, [dispatch, props.namespace, props.group, props.filename]);

    const activeNormal = React.useMemo(() => versions.find((item: any) => item.active && item.releaseType !== 'gray'), [versions]);
    const activeGrayByName = React.useMemo(() => versions.reduce((acc: Record<string, any>, item: any) => {
        if (item.active && item.releaseType === 'gray' && item.name) {
            acc[item.name] = item;
        }
        return acc;
    }, {}), [versions]);

    const subscriberRows = React.useMemo<SubscriberRow[]>(() => {
        return (subscribers || []).flatMap((item: VersionClient) => (item.subscribers || []).map((client) => {
            const gray = activeGrayByName[client.release_name];
            if (gray) {
                return {
                    ...client,
                    resolvedStatus: 'gray',
                    resolvedText: `命中灰度 v${gray.version || client.version}`,
                } as SubscriberRow;
            }
            if (activeNormal && String(activeNormal.version) === String(client.version)) {
                return {
                    ...client,
                    resolvedStatus: 'normal',
                    resolvedText: `当前全量 v${client.version}`,
                } as SubscriberRow;
            }
            return {
                ...client,
                resolvedStatus: activeNormal ? 'normal' : 'missing',
                resolvedText: activeNormal ? `当前全量 v${client.version}` : '无可用版本',
            } as SubscriberRow;
        }));
    }, [activeGrayByName, activeNormal, subscribers]);

    const table = (
        <>
            <div className={style.subscribeShell}>
                <Table
                    data={subscriberRows}
                    columns={columns(setSelectedSubscriber)}
                    loading={loading}
                    rowKey="id"
                    size={"medium"}
                    tableLayout={'fixed'}
                    cellEmptyContent={'-'}
                    pagination={{
                        defaultCurrent: 1,
                        defaultPageSize: 6,
                        total: subscriberRows.length,
                        showJumper: true,
                    }}
                    selectOnRowClick={false}
                    ariaLabel="配置文件订阅客户端"
                />
            </div>
        </>
    )

    return (
        <>
            {table}
            <Drawer
                size="min(560px, calc(100vw - 32px))"
                header="订阅客户端详情"
                footer={false}
                visible={Boolean(selectedSubscriber)}
                onClose={() => setSelectedSubscriber(null)}
            >
                {selectedSubscriber && (
                    <Descriptions column={1} tableLayout="fixed">
                        <DescriptionsItem label="客户端 ID">{selectedSubscriber.id || '-'}</DescriptionsItem>
                        <DescriptionsItem label="客户端 IP">{selectedSubscriber.host || '-'}</DescriptionsItem>
                        <DescriptionsItem label="客户端类型">{selectedSubscriber.client_type || '-'}</DescriptionsItem>
                        <DescriptionsItem label="命名空间">{props.namespace || '-'}</DescriptionsItem>
                        <DescriptionsItem label="配置分组">{props.group || '-'}</DescriptionsItem>
                        <DescriptionsItem label="配置文件">{props.filename || '-'}</DescriptionsItem>
                        <DescriptionsItem label="监听版本">{selectedSubscriber.resolvedText || `v${selectedSubscriber.version || '-'}`}</DescriptionsItem>
                        <DescriptionsItem label="灰度发布">{selectedSubscriber.release_name || '未命中灰度发布'}</DescriptionsItem>
                    </Descriptions>
                )}
            </Drawer>
        </>
    )
}

export default React.memo(SubscribeTable);
