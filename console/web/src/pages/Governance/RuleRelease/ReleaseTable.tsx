import React from 'react';
import { Space, Button, Drawer, Table, Tooltip, Tag, Popconfirm, Empty } from 'components/Fluent';
import type { PageInfo, PaginationProps, PrimaryTableProps, TableRowData } from 'components/Fluent';
import { Delete1Icon, RollbackIcon } from 'components/Fluent/icons';

import Text from 'components/Text';
import { Op } from 'services/types';
import style from './ReleaseTable.module.less';
import ResourceNameLink from 'components/ResourceNameLink';

interface IReleaseTableProps {
    datas: TableRowData[];
    loading: boolean;
    pagination: PaginationProps;
    onPageChange?: (pageInfo: PageInfo) => void;
    action:  (op: Op, row: TableRowData) => void;
    editable: boolean;
    deleteable: boolean;
    rollbackable?: boolean;
}

export const ReleaseColumns = (
    status: { editable: boolean; deleteable: boolean; rollbackable: boolean },
    operateRelease: (op: Op, row: TableRowData) => void,
): PrimaryTableProps['columns'] => [
    {
        colKey: 'release_name',
        title: '名称',
        width: 176,
        fixed: 'left',
        cell: ({ row }) => <ResourceNameLink name={row.release_name} onClick={() => operateRelease('view', { ...row })} />,
    },
    {
        colKey: 'version',
        title: '系统版本',
        width: 72,
        cell: ({ row: { version } }) => <Text>{version}</Text>,
    },
    {
        colKey: 'description',
        title: '描述',
        ellipsis: true,
        width: 150,
        cell: ({ row: { description } }) => (
            <div className={style.descriptionCell}>
                <Text>{description || '-'}</Text>
            </div>
        ),
    },
    {
        colKey: 'active',
        title: '状态',
        width: 80,
        cell: ({ row: { active } }) => active ? <Tag theme='primary' variant="light">使用中</Tag> : <Tag variant="light">历史</Tag>,
    },
    {
        colKey: 'releaseType',
        title: '发布类型',
        width: 90,
        cell: ({ row: { releaseType } }) => releaseType === 'gray' ? <Tag theme='warning' variant="light-outline">灰度发布</Tag> : <Tag theme='success' variant="light-outline">全量发布</Tag>,
    },
    {
        colKey: 'ctime',
        title: '发布时间',
        width: 124,
        cell: ({ row: { ctime } }) => (
            <div className={style.timeCell}>
                <Text>{ctime || '-'}</Text>
            </div>
        ),
    },
    {
        colKey: 'action',
        title: '操作',
        width: 76,
        fixed: 'right',
        align: 'center',
        cell: ({ row }) => {
            return (
                <Space>
                    {status.rollbackable && row.releaseType !== 'gray' && !row.active && (
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

const ReleaseTable: React.FC<IReleaseTableProps> = ({ datas, loading, pagination, onPageChange, action, editable, deleteable, rollbackable = true }) => {
    const [viewingRelease, setViewingRelease] = React.useState<TableRowData | null>(null);
    const activeRelease = datas.find((item) => item.active);
    const grayCount = datas.filter((item) => item.releaseType === 'gray').length;
    const normalCount = datas.length - grayCount;

    const table = (
        <div className={style.releasePanel}>
            <div className={style.summaryRail}>
                <div className={style.summaryItem}>
                    <span>发布版本</span>
                    <strong>{pagination.total || datas.length}</strong>
                </div>
                <div className={style.summaryItem}>
                    <span>当前使用</span>
                    <strong>{activeRelease?.version || '-'}</strong>
                </div>
                <div className={style.summaryItem}>
                    <span>全量发布</span>
                    <strong>{normalCount}</strong>
                </div>
                <div className={style.summaryItem}>
                    <span>灰度发布</span>
                    <strong>{grayCount}</strong>
                </div>
            </div>
            <Table
                data={datas}
                columns={ReleaseColumns({
                    editable: editable,
                    deleteable: deleteable,
                    rollbackable: rollbackable,
                }, (op, row) => {
                    if (op === 'view') {
                        setViewingRelease(row);
                        return;
                    }
                    action(op, row);
                })}
                loading={loading}
                rowKey="id"
                size="medium"
                tableLayout="fixed"
                cellEmptyContent={'-'}
                empty={(
                    <Empty
                        title="暂无发布版本"
                        description="当前规则还没有发布记录，发布后会在这里查看版本、状态和发布时间。"
                    />
                )}
                pagination={pagination}
                onPageChange={(pageInfo) => {
                    if (onPageChange) {
                        onPageChange(pageInfo)
                    }
                }}
            />
            <Drawer
                visible={Boolean(viewingRelease)}
                header={viewingRelease?.release_name || '发布版本详情'}
                footer={false}
                size="small"
                onClose={() => setViewingRelease(null)}
            >
                {viewingRelease && (
                    <div className={style.releaseDetail}>
                        <div><span>名称</span><strong>{viewingRelease.release_name || '-'}</strong></div>
                        <div><span>系统版本</span><strong>{viewingRelease.version || '-'}</strong></div>
                        <div><span>状态</span><strong>{viewingRelease.active ? '使用中' : '历史'}</strong></div>
                        <div><span>发布类型</span><strong>{viewingRelease.releaseType === 'gray' ? '灰度发布' : '全量发布'}</strong></div>
                        <div><span>发布时间</span><strong>{viewingRelease.ctime || '-'}</strong></div>
                        <div className={style.releaseDetailWide}><span>描述</span><strong>{viewingRelease.description || '-'}</strong></div>
                    </div>
                )}
            </Drawer>
        </div>
    )

    return (
        <>
            {table}
        </>
    )
}

export default React.memo(ReleaseTable);
