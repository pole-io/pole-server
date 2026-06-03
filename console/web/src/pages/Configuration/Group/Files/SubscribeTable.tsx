import React, { useState } from 'react';
import { Space, Button, Table, Tooltip, Descriptions, Tree } from "tdesign-react";
import type { PrimaryTableProps, TableProps, TableRowData } from 'tdesign-react';
import { Delete1Icon, Edit1Icon, RollbackIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import Text from 'components/Text';
import { VersionClient } from 'services/config_release';
import { openErrNotification } from 'utils/notifition';

import style from './index.module.less';
import { cleanFilePage, listFileSubscribers, selectConfigFile } from 'modules/configuration/file';

interface ISubscribeTableProps {
    namespace: string;
    group: string;
    filename: string;
    editable: boolean;
    deleteable: boolean;
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
                    <Tooltip content={props.editable === false ? '无权限操作' : '重发布'}>
                        <Button shape="square" variant="text" disabled={props.editable === false}>
                            <Edit1Icon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={props.editable === false ? '无权限操作' : '回滚至此版本'}>
                        <Button shape="square" variant="text" disabled={row.deleteable === false}>
                            <RollbackIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={props.editable === false ? '无权限操作' : '撤销'}>
                        <Button shape="square" variant="text" disabled={row.deleteable === false}>
                            <Delete1Icon />
                        </Button>
                    </Tooltip>
                </Space>
            )
        },
    },
]

const SubscribeTable: React.FC<ISubscribeTableProps> = (props) => {
    const dispatch = useAppDispatch();

    const fileState = useAppSelector(selectConfigFile);
    const { subscribers } = fileState;

    const [searchState, setSearchState] = useState<{
        versionTree: any[]
        subscribers: TableProps['data'];
        total: number;
        query: string;
        fetchError: boolean;
        isLoading: boolean;
    }>({ versionTree: [], subscribers: [], total: 0, query: '', fetchError: false, isLoading: false });

    React.useEffect(() => {
        if (props.namespace && props.group && props.filename) {
            dispatch(listFileSubscribers({
                param: {
                    namespace: props.namespace,
                    group: props.group,
                    file_name: props.filename,
                }
            })).then((res) => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('请求失败', res?.payload as string);
                }
            });
        }
        return () => {
            cleanFilePage();
        }
    }, [props.namespace, props.group, props.filename]);

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
                title={`${props.filename}`}
            ></Descriptions>
            <Space>
                <div className={style.treeContent}>
                    <Tree data={renderVersionTree(subscribers)} activable hover transition />
                </div>
                <Table
                    data={subscribers}
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