import React, { useState } from 'react';
import { Space, Button, Table, Tooltip, Descriptions, Tree } from "tdesign-react";
import type { PrimaryTableProps, TableProps, TableRowData } from 'tdesign-react';
import { ListIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
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
        cell: ({ row: { id } }) => <Text>{id}</Text>,
    },
    {
        colKey: 'host',
        title: '客户端IP',
        cell: ({ row: { host } }) => <Text>{host}</Text>,
    },
    {
        colKey: 'client_type',
        title: '客户端类型',
        cell: ({ row: { client_type } }) => <Text>{client_type}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
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
    const dispatch = useAppDispatch();

    const [searchState, setSearchState] = useState<{
        versionTree: any[]
        subscribers: TableProps['data'];
        total: number;
        query: string;
        fetchError: boolean;
        isLoading: boolean;
    }>({ versionTree: [], subscribers: [], total: 0, query: '', fetchError: false, isLoading: false });

    const renderVersionTree = (subscribers: VersionClient[]) => {
        const versions: Record<string, boolean> = {}
        subscribers.forEach((client) => {
            versions[client.version.toString()] = true;
        });
        return Object.keys(versions).map((version) => ({
            label: version,
            value: version,
            children: false,
        }));
    }

    const handleViewRelease = (view: boolean, row: TableRowData) => {

    }

    const table = (
        <>
            <Descriptions
                itemLayout="horizontal"
                layout="horizontal"
                size="small"
                title={`${props.title}`}
            ></Descriptions>
            <Space>
                <div className={style.treeContent}>
                    <Tree data={renderVersionTree(props.subscribers)} activable hover transition />
                </div>
                <Table
                    data={props.subscribers}
                    columns={columns(props, handleViewRelease)}
                    loading={searchState.isLoading}
                    rowKey="id"
                    size={"large"}
                    tableLayout={'fixed'}
                    cellEmptyContent={'-'}
                    pagination={{
                        defaultCurrent: 1,
                        defaultPageSize: 10,
                        total: searchState.total,
                        showJumper: true,
                    }}
                    selectOnRowClick={false}
                />
            </Space>
        </>
    )

    return (
        <>
            {table}
        </>
    )
}

export default React.memo(SubscribeTable);