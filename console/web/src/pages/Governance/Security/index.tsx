import React from 'react';
import { Button, Col, Empty, Link, Popconfirm, Row, Space, Table, Tabs, Tag, Tooltip } from 'tdesign-react';
import type { PageInfo, PrimaryTableProps, TableRowData } from 'tdesign-react';
import { CreditcardIcon, DeleteIcon, RefreshIcon } from 'tdesign-icons-react';

import AuthorizeInput from 'components/Authorize';
import Search from 'components/Search';
import SubscribeTable from 'components/SubscribeTable';
import { PolicySourceType } from 'services/auth_policy';
import { Op, RuleRelease } from 'services/types';
import {
    deleteTrafficGovernanceRelease,
    deleteTrafficGovernanceRules,
    describeTrafficGovernanceRules,
    describeTrafficGovernanceVersions,
    TrafficGovernanceAuthResource,
    TrafficGovernanceKind,
    TrafficGovernanceKindLabel,
    TrafficGovernanceRule,
} from 'services/traffic_governance';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import RuleDetailDrawer from '../RuleRelease/RuleDetailDrawer';
import ReleaseTable from '../RuleRelease/ReleaseTable';
import RuleTabs from '../RuleRelease/RuleTabs';
import TrafficGovernanceEditor, { trafficRuleCount, trafficRuleSummary } from './TrafficGovernanceEditor';
import style from './index.module.less';

const { TabPanel } = Tabs;

interface TrafficGovernanceTableProps {
    kind: TrafficGovernanceKind;
}

interface EditorState {
    visible: boolean;
    authorizeVisible: boolean;
    mode: Op;
    data?: TrafficGovernanceRule;
}

const columns = (
    kind: TrafficGovernanceKind,
    openRule: (row: TrafficGovernanceRule) => void,
    operate: (row: TrafficGovernanceRule, op: Op) => void,
): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '规则',
        fixed: 'left',
        width: 240,
        cell: ({ row }) => (
            <div className={style.ruleName}>
                <Link className={style.ruleNameText} theme="primary" onClick={() => openRule(row as TrafficGovernanceRule)}>{row.name}</Link>
                <div className={style.ruleDesc}>{row.description || '-'}</div>
            </div>
        ),
    },
    {
        colKey: 'targetNamespace',
        title: '被调命名空间',
        width: 150,
        cell: ({ row }) => (row as TrafficGovernanceRule).target_service?.namespace || (row as TrafficGovernanceRule).namespace || '-',
    },
    {
        colKey: 'targetService',
        title: '被调服务',
        width: 160,
        cell: ({ row }) => (row as TrafficGovernanceRule).target_service?.service || (row as TrafficGovernanceRule).service || '-',
    },
    {
        colKey: 'ruleCount',
        title: kind === 'security' ? '策略数' : '规则数',
        width: 90,
        cell: ({ row }) => trafficRuleCount(kind, row as TrafficGovernanceRule),
    },
    {
        colKey: 'summary',
        title: '规则摘要',
        ellipsis: true,
        cell: ({ row }) => trafficRuleSummary(kind, row as TrafficGovernanceRule),
    },
    {
        colKey: 'enable',
        title: '状态',
        width: 90,
        cell: ({ row }) => <Tag theme={row.enable ? 'success' : 'default'} variant="light-outline">{row.enable ? '启用' : '禁用'}</Tag>,
    },
    {
        colKey: 'priority',
        title: '优先级',
        width: 90,
    },
    {
        colKey: 'operation',
        title: '操作',
        fixed: 'right',
        width: 104,
        cell: ({ row }) => (
            <Space>
                <Tooltip content={row.editable === false ? '无权限操作' : '授权'}>
                    <Button
                        shape="square"
                        variant="text"
                        disabled={row.editable === false}
                        onClick={() => operate(row as TrafficGovernanceRule, 'authorize')}
                    >
                        <CreditcardIcon />
                    </Button>
                </Tooltip>
                <Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
                    <Popconfirm
                        content="确认删除吗"
                        destroyOnClose
                        placement="top"
                        showArrow
                        theme="default"
                        onConfirm={() => operate(row as TrafficGovernanceRule, 'delete')}
                    >
                        <Button shape="square" variant="text" disabled={row.deleteable === false}>
                            <DeleteIcon />
                        </Button>
                    </Popconfirm>
                </Tooltip>
            </Space>
        ),
    },
];

const TrafficGovernanceTable: React.FC<TrafficGovernanceTableProps> = ({ kind }) => {
    const [datas, setDatas] = React.useState<TrafficGovernanceRule[]>([]);
    const [total, setTotal] = React.useState(0);
    const [page, setPage] = React.useState(1);
    const [limit, setLimit] = React.useState(10);
    const [loading, setLoading] = React.useState(false);
    const [query, setQuery] = React.useState('');
    const [versions, setVersions] = React.useState<RuleRelease[]>([]);
    const [versionTotal, setVersionTotal] = React.useState(0);
    const [versionPage, setVersionPage] = React.useState(1);
    const [versionLimit, setVersionLimit] = React.useState(10);
    const [versionLoading, setVersionLoading] = React.useState(false);
    const [editor, setEditor] = React.useState<EditorState>({
        visible: false,
        authorizeVisible: false,
        mode: '',
    });

    const refreshData = React.useCallback(async (nextPage = page, nextLimit = limit, nextQuery = query) => {
        setLoading(true);
        try {
            const res = await describeTrafficGovernanceRules(kind, {
                offset: (nextPage - 1) * nextLimit,
                limit: nextLimit,
                name: nextQuery,
            });
            setDatas(res.list);
            setTotal(res.totalCount);
            setPage(nextPage);
            setLimit(nextLimit);
        } catch (error) {
            openErrNotification('请求失败', `查询${TrafficGovernanceKindLabel[kind]}规则失败: ${(error as Error).message}`);
        } finally {
            setLoading(false);
        }
    }, [kind, limit, page, query]);

    React.useEffect(() => {
        refreshData(1, 10, '');
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [kind]);

    const refreshVersions = async (nextPage = 1, nextLimit = 10) => {
        if (!editor.data?.id) return;
        setVersionLoading(true);
        try {
            const res = await describeTrafficGovernanceVersions(kind, {
                id: editor.data.id,
                rule_name: editor.data.name,
                offset: (nextPage - 1) * nextLimit,
                limit: nextLimit,
            });
            setVersions(res.list);
            setVersionTotal(res.totalCount);
            setVersionPage(nextPage);
            setVersionLimit(nextLimit);
        } catch (error) {
            openErrNotification('请求失败', `查询${TrafficGovernanceKindLabel[kind]}版本失败: ${(error as Error).message}`);
        } finally {
            setVersionLoading(false);
        }
    };

    const openRule = (row: TrafficGovernanceRule) => {
        setVersions([]);
        setVersionTotal(0);
        setVersionPage(1);
        setVersionLimit(10);
        setEditor((prev) => ({ ...prev, visible: true, mode: 'view', data: row }));
    };

    const operate = async (row: TrafficGovernanceRule, op: Op) => {
        if (op === 'create') {
            setEditor((prev) => ({ ...prev, visible: true, mode: 'create', data: undefined }));
            return;
        }
        if (op === 'authorize') {
            setEditor((prev) => ({ ...prev, authorizeVisible: true, data: row, mode: 'authorize' }));
            return;
        }
        if (op === 'delete') {
            try {
                await deleteTrafficGovernanceRules(kind, [{ id: row.id }]);
                openInfoNotification('请求成功', `删除${TrafficGovernanceKindLabel[kind]}规则成功`);
                refreshData(1, limit, query);
            } catch (error) {
                openErrNotification('请求失败', `删除${TrafficGovernanceKindLabel[kind]}规则失败: ${(error as Error).message}`);
            }
        }
    };

    const operateRelease = async (op: Op, row: TableRowData) => {
        if (op !== 'delete') return;
        try {
            await deleteTrafficGovernanceRelease(kind, row.id as string);
            openInfoNotification('请求成功', `删除${TrafficGovernanceKindLabel[kind]}版本成功`);
            refreshVersions(versionPage, versionLimit);
        } catch (error) {
            openErrNotification('请求失败', `删除${TrafficGovernanceKindLabel[kind]}版本失败: ${(error as Error).message}`);
        }
    };

    const table = (
        <>
            <Row justify="space-between" className={style.toolBar}>
                <Col>
                    <Row gutter={8} align="middle">
                        <Col>
                            <Button theme="primary" onClick={() => operate({} as TrafficGovernanceRule, 'create')}>新建</Button>
                        </Col>
                    </Row>
                </Col>
                <Col>
                    <Space>
                        <Search
                            onChange={(value: string) => {
                                setQuery(value);
                                refreshData(1, limit, value);
                            }}
                        />
                        <Tooltip content="刷新">
                            <RefreshIcon onClick={() => refreshData(1, limit, query)} />
                        </Tooltip>
                    </Space>
                </Col>
            </Row>
            {editor.authorizeVisible && editor.data && (
                <AuthorizeInput
                    resource_type={TrafficGovernanceAuthResource[kind] as PolicySourceType}
                    resource_id={editor.data.id as string}
                    resource_name={`${kind}_rule/${editor.data.name}`}
                    visible={editor.authorizeVisible}
                    onClose={() => setEditor((prev) => ({ ...prev, authorizeVisible: false }))}
                />
            )}
            <Table
                data={datas}
                columns={columns(kind, openRule, operate)}
                loading={loading}
                rowKey="id"
                tableLayout="fixed"
                cellEmptyContent="-"
                empty={<Empty title={`暂无${TrafficGovernanceKindLabel[kind]}规则`} />}
                pagination={{
                    current: page,
                    pageSize: limit,
                    total,
                    showJumper: true,
                    onChange(pageInfo) {
                        refreshData(pageInfo.current, pageInfo.pageSize, query);
                    },
                }}
                onPageChange={(pageInfo: PageInfo) => refreshData(pageInfo.current, pageInfo.pageSize, query)}
            />
        </>
    );

    return (
        <div className={style.ruleWorkspace}>
            <section className={style.ruleListPane}>
                {table}
            </section>
            <RuleDetailDrawer
                visible={editor.visible}
                title={editor.mode === 'create' ? `新建${TrafficGovernanceKindLabel[kind]}规则` : editor.data?.name || `${TrafficGovernanceKindLabel[kind]}详情`}
                subtitle={TrafficGovernanceKindLabel[kind]}
                onClose={() => setEditor((prev) => ({ ...prev, visible: false }))}
            >
                <RuleTabs
                    op={editor.mode}
                    onVersionView={() => refreshVersions()}
                    view={(
                        <TrafficGovernanceEditor
                            kind={kind}
                            op={editor.mode}
                            data={editor.data}
                            visible={editor.visible}
                            refresh={(close) => {
                                if (close) setEditor((prev) => ({ ...prev, visible: false }));
                                refreshData(1, limit, query);
                            }}
                        />
                    )}
                    versions={{
                        datas: versions,
                        action: operateRelease,
                        editable: editor.data?.editable ?? true,
                        deleteable: editor.data?.deleteable ?? true,
                        rollbackable: false,
                        loading: versionLoading,
                        pagination: {
                            current: versionPage,
                            pageSize: versionLimit,
                            total: versionTotal,
                            showJumper: false,
                            onChange(pageInfo) {
                                refreshVersions(pageInfo.current, pageInfo.pageSize);
                            },
                        },
                        onPageChange: (pageInfo) => refreshVersions(pageInfo.current, pageInfo.pageSize),
                    }}
                    subscribe={(
                        <div style={{ marginLeft: 20, marginTop: 20 }}>
                            <SubscribeTable
                                title={editor.data?.name || ''}
                                editable={editor.data?.editable ?? true}
                                deleteable={editor.data?.deleteable ?? true}
                                subscribers={[]}
                            />
                        </div>
                    )}
                />
            </RuleDetailDrawer>
        </div>
    );
};

export default React.memo(() => (
    <Tabs defaultValue="security">
        <TabPanel label="鉴权" value="security">
            <div style={{ margin: 20 }}>
                <TrafficGovernanceTable kind="security" />
            </div>
        </TabPanel>
        <TabPanel label="镜像" value="mirror">
            <div style={{ margin: 20 }}>
                <TrafficGovernanceTable kind="mirror" />
            </div>
        </TabPanel>
        <TabPanel label="Mock" value="mock">
            <div style={{ margin: 20 }}>
                <TrafficGovernanceTable kind="mock" />
            </div>
        </TabPanel>
    </Tabs>
));
