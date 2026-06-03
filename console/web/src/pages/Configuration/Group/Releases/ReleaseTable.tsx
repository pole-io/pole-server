import ErrorPage from 'components/ErrorPage';
import React, { useState } from 'react';
import { Drawer, Form, Input, Space, Button, Select, Table, Tooltip, Descriptions, Tag, Popconfirm, Link } from "tdesign-react";
import type { FormProps, PrimaryTableProps, TableProps, TableRowData } from 'tdesign-react';
import { Delete1Icon, Edit1Icon, RollbackIcon, SendIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import Text from 'components/Text';
import { ConfigFileRelease, describeFileReleaseVersions, releaseConfigFile, ReleaseVersion, rollbackFileReleases } from 'services/config_release';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { cleanFileReleasePage, editorFileRelease, listConfigFileReleases, publishConfigFiles, releaseRollback, releasesRemove, selectFileRelease, viewFileRelease } from 'modules/configuration/release';
import ReleaseDetail from './ReleaseDetail';
import { Op } from 'services/types';

interface IReleaseTableProps {
    namespace: string;
    group: string;
    filename: string;
    editable: boolean;
    deleteable: boolean;
}

const columns = (props: IReleaseTableProps, operateRelease: (op: Op, row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }) => <Link theme='primary' onClick={() => operateRelease('view', { ...row })}>{row.name}</Link>,
    },
    {
        colKey: 'version',
        title: 'Version',
        cell: ({ row: { version } }) => <Text>{version}</Text>,
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
        cell: ({ row: { createTime } }) => <Text>{createTime}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    {row.releaseType !== 'gray' && !row.active && (
                        <>
                            <Tooltip content={props.editable ? '回滚至此版本' : '无权限操作'}>
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
                                        disabled={!props.editable}
                                    />
                                </Popconfirm>
                            </Tooltip>
                        </>
                    )}
                    <Tooltip content={props.editable ? '撤销' : '无权限操作'}>
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
                                disabled={!props.deleteable}
                            />
                        </Popconfirm>

                    </Tooltip>
                </Space>
            )
        },
    },
]

const ReleaseTable: React.FC<IReleaseTableProps> = (props) => {
    const dispatch = useAppDispatch();

    const releaseState = useAppSelector(selectFileRelease);
    const { versions = [], total = 0, loading } = releaseState;

    const [searchState, setSearchState] = useState<{
        query: string;
        releaseVisible: boolean;
    }>({ query: '', releaseVisible: false });

    React.useEffect(() => {
        if (!props.namespace || !props.group || !props.filename) {
            return;
        }
        dispatch(listConfigFileReleases({
            param: {
                namespace: props.namespace,
                group: props.group,
                file_name: props.filename,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取版本列表失败', res.payload as string);
            }
        });
        return () => {
            dispatch(cleanFileReleasePage());
        }
    }, [props.namespace, props.group, props.filename]);

    const operateRelease = async (op: Op, row: TableRowData) => {
        if (op === 'publish') {
            // 重新发布
            dispatch(publishConfigFiles({
                param: {
                    namespace: props.namespace,
                    group: props.group,
                    fileName: props.filename,
                    name: row.name as string,
                    releaseDescription: row.releaseDescription as string,
                    releaseType: row.releaseType as 'normal' | 'gray',
                    betaLabels: []
                }
            })).then((res) => {
                if (res.meta.requestStatus === 'fulfilled') {
                    openInfoNotification('发布配置成功', '发布配置成功');
                } else {
                    openErrNotification('发布配置失败', res.payload as string);
                }
            })
        } else if (op === 'rollback') {
            // 回滚发布
            await dispatch(releaseRollback({
                param: {
                    namespace: props.namespace,
                    group: props.group,
                    fileName: props.filename,
                    name: row.name as string,
                }
            })).then((res) => {
                if (res.meta.requestStatus === 'fulfilled') {
                    openInfoNotification('回滚发布成功', '回滚发布成功');
                } else {
                    openErrNotification('回滚发布失败', res.payload as string);
                }
            })
        } else if (op === 'delete') {
            // 删除发布
            await dispatch(releasesRemove({
                param: [{
                    namespace: props.namespace,
                    group: props.group,
                    fileName: props.filename,
                    name: row.name as string
                }]
            })).then((res) => {
                if (res.meta.requestStatus === 'fulfilled') {
                    openInfoNotification('请求成功', '撤销发布成功');
                } else {
                    openErrNotification('撤销发布失败', res.payload as string);
                }
            })
        } else if (op === 'view') {
            // 查看发布详细
            dispatch(editorFileRelease({ ...row as ConfigFileRelease }))
            setSearchState((prev) => ({ ...prev, releaseVisible: true }));
        }
    }

    const table = (
        <>
            <Descriptions
                itemLayout="horizontal"
                layout="horizontal"
                size="small"
                title={`${props.filename}`}
            ></Descriptions>
            <Table
                data={versions}
                columns={columns(props, operateRelease)}
                loading={loading}
                rowKey="id"
                size={"large"}
                tableLayout={'auto'}
                cellEmptyContent={'-'}
                pagination={{
                    defaultCurrent: 1,
                    defaultPageSize: 10,
                    total: total,
                    showJumper: true,
                }}
            />
            {searchState.releaseVisible && (
                <ReleaseDetail
                    visible={searchState.releaseVisible}
                    namespace={props.namespace}
                    group={props.group}
                    fileName={props.filename}
                    onClose={() => {
                        setSearchState((prev) => ({ ...prev, releaseVisible: false }));
                    }}
                />
            )}
        </>
    )

    return (
        <>
            {table}
        </>
    )
}

export default React.memo(ReleaseTable);