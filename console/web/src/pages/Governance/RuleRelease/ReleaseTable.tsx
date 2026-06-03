import React from 'react';
import { Space, Button, Table, Tooltip, Tag, Popconfirm, Link } from "tdesign-react";
import type { PageInfo, PaginationProps, PrimaryTableProps, TableRowData } from 'tdesign-react';
import { Delete1Icon, RollbackIcon } from 'tdesign-icons-react';

import Text from 'components/Text';
import { Op } from 'services/types';

interface IReleaseTableProps {
    datas: TableRowData[];
    loading: boolean;
    pagination: PaginationProps;
    onPageChange?: (pageInfo: PageInfo) => void;
    action:  (op: Op, row: TableRowData) => void;
    editable: boolean;
    deleteable: boolean;
}

export const ReleaseColumns = (status: { editable: boolean; deleteable: boolean }, operateRelease: (op: Op, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'release_name',
        title: '名称',
        cell: ({ row }) => <Link theme='primary' onClick={() => operateRelease('view', { ...row })}>{row.release_name}</Link>,
    },
    {
        colKey: 'version',
        title: '系统版本',
        cell: ({ row: { version } }) => <Text>{version}</Text>,
    },
    {
        colKey: 'description',
        title: '描述',
        cell: ({ row: { description } }) => <Text>{description}</Text>,
    },
    {
        colKey: 'active',
        title: '状态',
        cell: ({ row: { active } }) => active ? <Tag theme='primary'>使用中</Tag> : <></>,
    },
    {
        colKey: 'releaseType',
        title: '发布类型',
        cell: ({ row: { releaseType } }) => releaseType === 'gray' ? <Tag theme='warning'>灰度发布</Tag> : <Tag theme='success'>全量发布</Tag>,
    },
    {
        colKey: 'ctime',
        title: '发布时间',
        cell: ({ row: { ctime } }) => <Text>{ctime}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    {row.releaseType !== 'gray' && !row.active && (
                        <>
                            <Tooltip content={status.editable ? '回滚至此版本' : '无权限操作'}>
                                <Popconfirm
                                    content="确认回滚至此版本吗"
                                    destroyOnClose
                                    placement="top"
                                    showArrow
                                    theme="default"
                                    onConfirm={() => {
                                        operateRelease('rollback', row)
                                    }}
                                >
                                    <Button
                                        shape="square"
                                        icon={<RollbackIcon />}
                                        variant="text"
                                        disabled={!status.editable}
                                    />
                                </Popconfirm>
                            </Tooltip>
                        </>
                    )}
                    <Tooltip content={status.editable ? '撤销' : '无权限操作'}>
                        <Popconfirm
                            content={row.active ? "当前版本正在使用中，删除后客户端将获取不到此配置" : "确认删除版本吗"}
                            destroyOnClose
                            placement="top"
                            showArrow
                            theme="default"
                            onConfirm={() => {
                                operateRelease('delete', row)
                            }}
                        >
                            <Button
                                shape="square"
                                icon={<Delete1Icon />}
                                variant="text"
                                disabled={!status.deleteable}
                            />
                        </Popconfirm>
                    </Tooltip>
                </Space>
            )
        },
    },
]

const ReleaseTable: React.FC<IReleaseTableProps> = ({ datas, loading, pagination, onPageChange, action, editable, deleteable }) => {
    const table = (
        <>
            <Table
                data={datas}
                columns={ReleaseColumns({
                    editable: editable,
                    deleteable: deleteable,
                }, action)}
                loading={loading}
                rowKey="id"
                size={"large"}
                tableLayout={'auto'}
                cellEmptyContent={'-'}
                pagination={pagination}
                onPageChange={(pageInfo) => {
                    if (onPageChange) {
                        onPageChange(pageInfo)
                    }
                }}
            />
        </>
    )

    return (
        <>
            {table}
        </>
    )
}

export default React.memo(ReleaseTable);