import React, { useMemo } from 'react';
import { Space, Button, Table, Tooltip, Tree, Empty, Tag } from "tdesign-react";
import type { PrimaryTableProps, TableRowData } from 'tdesign-react';
import { ListIcon } from 'tdesign-icons-react';

import Text from 'components/Text';
import { VersionClient } from 'services/config_release';

import style from './index.module.less';
import { t } from 'i18next';

export interface ISubscribeTableProps {
    editable: boolean;
    deleteable: boolean;
    title: string;
    subscribers: VersionClient[];
}

const columns = (props: ISubscribeTableProps, handleViewRelease: (view: boolean, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'id',
        title: '客户端ID',
        width: 176,
        cell: ({ row: { id } }) => <Text>{id}</Text>,
    },
    {
        colKey: 'host',
        title: '客户端IP',
        width: 126,
        cell: ({ row: { host } }) => <Text>{host}</Text>,
    },
    {
        colKey: 'client_type',
        title: '客户端类型',
        width: 112,
        cell: ({ row: { client_type } }) => client_type ? <Tag variant="light">{client_type}</Tag> : <Text>-</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        width: 72,
        align: 'center',
        cell: ({ row }) => {
            return (
                <Space>
                    <Tooltip content={t('button.tip.subscribe.showTrace')}>
                        <Button shape="square" variant="text" disabled={props.editable === false}>
                            <ListIcon />
                        </Button>
                    </Tooltip>
                </Space>
            )
        },
    },
]

const SubscribeTable: React.FC<ISubscribeTableProps> = (props) => {
    const versionTree = useMemo(() => {
        const versions: Record<string, number> = {}
        props.subscribers.forEach((client) => {
            const version = client.version?.toString() || '-';
            versions[version] = (versions[version] || 0) + 1;
        });
        return Object.keys(versions).map((version) => ({
            label: `版本 ${version} (${versions[version]})`,
            value: version,
            children: false,
        }));
    }, [props.subscribers]);

    const clientTypes = useMemo(() => {
        const types = new Set(props.subscribers.map((client) => client.client_type).filter(Boolean));
        return types.size;
    }, [props.subscribers]);

    const handleViewRelease = (view: boolean, row: TableRowData) => {

    }

    const table = (
        <div className={style.subscribePanel}>
            <div className={style.summaryRail}>
                <div className={style.summaryItem}>
                    <span>监听对象</span>
                    <strong>{props.subscribers.length}</strong>
                </div>
                <div className={style.summaryItem}>
                    <span>订阅版本</span>
                    <strong>{versionTree.length}</strong>
                </div>
                <div className={style.summaryItem}>
                    <span>客户端类型</span>
                    <strong>{clientTypes}</strong>
                </div>
            </div>
            <div className={style.content}>
                <aside className={style.treeContent}>
                    <div className={style.sectionTitle}>
                        <strong>{props.title}</strong>
                        <span>按客户端订阅版本分组</span>
                    </div>
                    {versionTree.length > 0 ? (
                        <Tree data={versionTree} activable hover transition />
                    ) : (
                        <Empty
                            title="暂无监听版本"
                            description="还没有客户端订阅该规则版本。"
                        />
                    )}
                </aside>
                <div className={style.tableContent}>
                    <div className={style.tableHeader}>
                        <strong>客户端列表</strong>
                        <span>{props.subscribers.length > 0 ? `当前 ${props.subscribers.length} 个客户端` : '暂无客户端订阅'}</span>
                    </div>
                <Table
                    data={props.subscribers}
                    columns={columns(props, handleViewRelease)}
                    loading={false}
                    rowKey="id"
                    size="medium"
                    tableLayout={'fixed'}
                    cellEmptyContent={'-'}
                    empty={(
                        <Empty
                            title="暂无监听客户端"
                            description="客户端拉取并监听该治理规则后，会在这里显示客户端 ID、IP 和类型。"
                        />
                    )}
                    pagination={{
                        defaultCurrent: 1,
                        defaultPageSize: 10,
                        total: props.subscribers.length,
                        showJumper: true,
                    }}
                    selectOnRowClick={false}
                />
                </div>
            </div>
        </div>
    )

    return (
        <>
            {table}
        </>
    )
}

export default React.memo(SubscribeTable);
