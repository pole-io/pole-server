import React, { useState } from 'react';
import { Space, Button, Table, Tag, Link, Tabs } from 'components/Fluent';
import type { PrimaryTableProps, TableRowData } from 'components/Fluent';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { ConfirmOperationButton, OperationButton } from 'components/OperationButton';
import Text from 'components/Text';
import { ConfigFileRelease } from 'services/config_release';
import { describeUsers, User } from 'services/users';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { cleanFileReleasePage, editorFileRelease, listConfigFileReleases, promoteGrayReleaseToDraft, publishConfigFiles, releaseRollback, releasesRemove, selectFileRelease, stopGrayRelease } from 'modules/configuration/release';
import ReleaseDetail from './ReleaseDetail';
import { Op } from 'services/types';
import style from '../Files/index.module.less';

const { TabPanel } = Tabs;

interface IReleaseTableProps {
    namespace: string;
    group: string;
    filename: string;
    editable: boolean;
    deleteable: boolean;
}

type ReleaseTab = 'normal' | 'gray' | 'draft' | 'history';

const releaseStatusText = (row: TableRowData) => {
    if (row.active && row.releaseType === 'gray') {
        return '灰度中';
    }
    if (row.active) {
        return '当前全量';
    }
    if (row.releaseStatus === 'to-be-released' || row.releaseStatus === 'draft') {
        return '正式草稿';
    }
    return '已归档';
};

const releaseStatusTheme = (row: TableRowData) => {
    if (row.active && row.releaseType === 'gray') {
        return 'warning';
    }
    if (row.active) {
        return 'success';
    }
    if (row.releaseStatus === 'to-be-released' || row.releaseStatus === 'draft') {
        return 'primary';
    }
    return 'default';
};

const releaseTypeText = (row: TableRowData) => {
    if (row.releaseStatus === 'to-be-released' || row.releaseStatus === 'draft') {
        return '草稿';
    }
    if (row.releaseType === 'gray') {
        return '灰度发布';
    }
    return row.releaseReason === 'rollback' ? '回滚发布' : '全量发布';
};

const grayRuleText = (row: TableRowData) => {
    if (row.releaseType !== 'gray') {
        return '—';
    }
    const labels = row.betaLabels || row.beta_labels || [];
    if (!labels.length) {
        return '未设置';
    }
    return labels.map((item: any) => `${item.key || item.type || 'label'}=${item.value?.value || item.value || ''}`).join(' / ');
};

const releaseTimeText = (row: TableRowData) => row.createTime || row.ctime || row.modifyTime || row.mtime || '-';

const operatorName = (row: Pick<TableRowData, 'createBy'> & Record<string, any>) => row.createBy || row.create_by || '';

type OperatorUsers = Record<string, User | null>;

const userDetailHref = (name: string, user: User) => {
    const displayName = user.name || name;
    return `/auth/principals/userdetail?name=${encodeURIComponent(displayName)}&id=${encodeURIComponent(user.id)}`;
};

const OperatorLink: React.FC<{ name?: string; users: OperatorUsers }> = ({ name, users }) => {
    if (!name) {
        return <Text>-</Text>;
    }
    const user = users[name];
    if (!user?.id) {
        return <Text>{name}</Text>;
    }
    return (
        <a
            className={style.operatorLink}
            href={userDetailHref(name, user)}
            target="_blank"
            rel="noreferrer"
        >
            {user.name || name}
        </a>
    );
};

const columns = (
    props: IReleaseTableProps,
    operateRelease: (op: Op | 'compare', row: TableRowData) => void,
    operatorUsers: OperatorUsers,
): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '版本',
        cell: ({ row }) => <Link theme='primary' onClick={() => operateRelease('view', { ...row })}>{row.name}</Link>,
    },
    {
        colKey: 'version',
        title: '版本号',
        cell: ({ row: { version } }) => <Text>{version ? `v${version}` : '-'}</Text>,
    },
    {
        colKey: 'status',
        title: '状态',
        cell: ({ row }) => <Tag theme={releaseStatusTheme(row) as any} variant="light">{releaseStatusText(row)}</Tag>,
    },
    {
        colKey: 'releaseType',
        title: '发布类型',
        cell: ({ row }) => releaseTypeText(row),
    },
    {
        colKey: 'grayRule',
        title: '灰度规则 / 优先级',
        ellipsis: true,
        cell: ({ row }) => (
            <span className={style.mono}>
                {row.releaseType === 'gray' ? `${grayRuleText(row)} / P${row.grayPriority || 100}` : '—'}
            </span>
        ),
    },
    {
        colKey: 'createBy',
        title: '发布人',
        cell: ({ row }) => <OperatorLink name={operatorName(row)} users={operatorUsers} />,
    },
    {
        colKey: 'ctime',
        title: '发布时间',
        cell: ({ row }) => <Text>{releaseTimeText(row)}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    {!row.active && row.releaseType !== 'gray' && row.releaseStatus !== 'to-be-released' && (
                        <>
                            <OperationButton action="view" label="对比" onClick={() => operateRelease('view', row)} />
                            <ConfirmOperationButton
                                action="rollback"
                                label={props.editable ? '回滚至此版本' : '无权限操作'}
                                disabled={!props.editable}
                                disabledLabel="无权限操作"
                                confirmContent="回滚前请先查看版本对比；确认后会生成新的当前全量，不会停止 active gray release。"
                                onConfirm={() => operateRelease('rollback', row)}
                            />
                        </>
                    )}
                    {row.releaseType === 'gray' && row.active && (
                        <>
                            <ConfirmOperationButton
                                action="publish"
                                label={props.editable ? '提交为正式草稿' : '无权限操作'}
                                disabled={!props.editable}
                                disabledLabel="无权限操作"
                                confirmContent="确认将此灰度版本内容提交为正式草稿吗"
                                onConfirm={() => operateRelease('promote', row)}
                            />
                            <ConfirmOperationButton
                                action="delete"
                                label={props.editable ? '删除灰度' : '无权限操作'}
                                disabled={!props.editable}
                                disabledLabel="无权限操作"
                                confirmContent="确认删除此灰度版本吗？命中客户端会回落当前全量。"
                                onConfirm={() => operateRelease('stopGray', row)}
                            />
                        </>
                    )}
                    {(row.releaseStatus === 'to-be-released' || row.releaseStatus === 'draft') && (
                        <OperationButton
                            action="publish"
                            label={props.editable ? '发布正式草稿' : '无权限操作'}
                            disabled={!props.editable}
                            disabledLabel="无权限操作"
                            onClick={() => operateRelease('publish', row)}
                        />
                    )}
                    {!(row.releaseType === 'gray' && row.active) && (
                        <ConfirmOperationButton
                            action="delete"
                            label={props.deleteable ? '删除' : '无权限操作'}
                            disabled={!props.deleteable}
                            disabledLabel="无权限操作"
                            confirmContent={row.active ? "当前全量删除后，未命中灰度的客户端将无法获取配置。确认删除吗？" : "确认删除记录吗？历史记录删除不影响当前内容和生效版本。"}
                            onConfirm={() => operateRelease('delete', row)}
                        />
                    )}
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
    const [activeTab, setActiveTab] = useState<ReleaseTab>('normal');
    const [pages, setPages] = useState<Record<ReleaseTab, number>>({
        normal: 1,
        gray: 1,
        draft: 1,
        history: 1,
    });
    const [operatorUsers, setOperatorUsers] = useState<OperatorUsers>({});

    const operatorNames = React.useMemo(() => {
        return Array.from(new Set(
            versions
                .map((item) => operatorName(item as TableRowData))
                .filter((item): item is string => Boolean(item)),
        ));
    }, [versions]);

    const refreshReleases = React.useCallback(() => {
        if (!props.namespace || !props.group || !props.filename) {
            return Promise.resolve();
        }
        return dispatch(listConfigFileReleases({
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
    }, [dispatch, props.namespace, props.group, props.filename]);

    React.useEffect(() => {
        refreshReleases();
        return () => {
            dispatch(cleanFileReleasePage());
        }
    }, [refreshReleases, dispatch]);

    React.useEffect(() => {
        const missingNames = operatorNames.filter((name) => !(name in operatorUsers));
        if (!missingNames.length) {
            return undefined;
        }

        let canceled = false;
        Promise.all(missingNames.map(async (name) => {
            try {
                const ret = await describeUsers({ name, offset: 0, limit: 10 });
                const user = ret.content.find((item) => item.name === name) || null;
                return [name, user] as const;
            } catch (error) {
                return [name, null] as const;
            }
        })).then((entries) => {
            if (canceled) {
                return;
            }
            setOperatorUsers((prev) => ({
                ...prev,
                ...Object.fromEntries(entries),
            }));
        });

        return () => {
            canceled = true;
        };
    }, [operatorNames, operatorUsers]);

    const operateRelease = async (op: Op | 'compare', row: TableRowData) => {
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
                    refreshReleases();
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
                    refreshReleases();
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
                    name: row.name as string,
                    releaseType: row.releaseType as string,
                }]
            })).then((res) => {
                if (res.meta.requestStatus === 'fulfilled') {
                    openInfoNotification('请求成功', '撤销发布成功');
                    refreshReleases();
                } else {
                    openErrNotification('撤销发布失败', res.payload as string);
                }
            })
        } else if (op === 'stopGray') {
            await dispatch(stopGrayRelease({
                param: {
                    namespace: props.namespace,
                    group: props.group,
                    fileName: props.filename,
                    name: row.name as string,
                    releaseType: 'gray',
                }
            })).then((res) => {
                if (res.meta.requestStatus === 'fulfilled') {
                    openInfoNotification('停止灰度成功', '停止灰度成功');
                    refreshReleases();
                } else {
                    openErrNotification('停止灰度失败', res.payload as string);
                }
            })
        } else if (op === 'promote') {
            await dispatch(promoteGrayReleaseToDraft({
                param: {
                    namespace: props.namespace,
                    group: props.group,
                    fileName: props.filename,
                    name: row.name as string,
                    releaseType: 'gray',
                }
            })).then((res) => {
                if (res.meta.requestStatus === 'fulfilled') {
                    openInfoNotification('提交正式草稿成功', '已提交为正式草稿，可继续执行全量发布');
                    refreshReleases();
                } else {
                    openErrNotification('提交正式草稿失败', res.payload as string);
                }
            })
        } else if (op === 'view') {
            // 查看发布详细
            dispatch(editorFileRelease({ ...row as ConfigFileRelease }))
            setSearchState((prev) => ({ ...prev, releaseVisible: true }));
        }
    }

    const grouped = React.useMemo(() => {
        const normal = versions.filter((item: any) => item.active && item.releaseType !== 'gray');
        const gray = versions.filter((item: any) => item.active && item.releaseType === 'gray');
        const draft = versions.filter((item: any) => item.releaseStatus === 'to-be-released' || item.releaseStatus === 'draft');
        const history = versions.filter((item: any) => !normal.includes(item) && !gray.includes(item) && !draft.includes(item));
        return { normal, gray, draft, history };
    }, [versions]);

    const tabMeta: Array<{ key: ReleaseTab; label: string; desc: string }> = [
        { key: 'normal', label: '正式发布', desc: '当前全量基线，正式发布不会结束 active gray release。' },
        { key: 'gray', label: '灰度发布', desc: '多条 active gray release 独立生效，可提交为正式草稿或删除灰度。' },
        { key: 'draft', label: '正式草稿', desc: '由灰度转正或手动编辑生成，尚未影响线上客户端。' },
        { key: 'history', label: '历史记录', desc: '已归档、回滚生成或删除生效后的记录。' },
    ];

    const renderTable = (tab: ReleaseTab) => {
        const pageSize = 6;
        const data = grouped[tab] as any[];
        const current = pages[tab] || 1;
        const paged = data.slice((current - 1) * pageSize, current * pageSize);
        const meta = tabMeta.find(item => item.key === tab);
        return (
            <>
                <div className={style.tableIntro}>{meta?.desc}</div>
                <Table
                    data={paged}
            columns={columns(props, operateRelease, operatorUsers)}
                    loading={loading}
                    rowKey={(row) => `${row.releaseType || 'normal'}-${row.name || row.id}-${row.version || ''}`}
                    size={"medium"}
                    tableLayout={'auto'}
                    cellEmptyContent={'-'}
                    pagination={{
                        current,
                        pageSize,
                        total: data.length,
                        showJumper: true,
                        onCurrentChange: (next: number) => setPages(prev => ({ ...prev, [tab]: next })),
                    }}
                />
            </>
        );
    };

    const table = (
        <>
            <Tabs
                className={style.releaseTabs}
                value={activeTab}
                onChange={(value) => setActiveTab(value as ReleaseTab)}
            >
                {tabMeta.map((item) => (
                    <TabPanel
                        key={item.key}
                        value={item.key}
                        label={(
                            <span className={style.tabLabel}>
                                {item.label}
                                <span className={style.tabCount}>{grouped[item.key].length}</span>
                            </span>
                        )}
                    >
                        {renderTable(item.key)}
                    </TabPanel>
                ))}
            </Tabs>
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
