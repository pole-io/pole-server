import React from 'react';
import { Drawer, Link, Table, Tag, TableProps } from 'tdesign-react';

import { describeAuthPolicies, PolicyRule } from 'services/auth_policy';
import { openErrNotification } from 'utils/notifition';
import PolicyDetailView from '../Policy/PolicyDetailView';

interface IPrincipalPolicyTableProps {
    principalId?: string;
    principalType: 1 | 2 | 3;
}

const PrincipalPolicyTable: React.FC<IPrincipalPolicyTableProps> = ({ principalId, principalType }) => {
    const [selectedPolicy, setSelectedPolicy] = React.useState<PolicyRule | null>(null);
    const [state, setState] = React.useState<{
        loading: boolean;
        policies: PolicyRule[];
        total: number;
    }>({ loading: false, policies: [], total: 0 });

    React.useEffect(() => {
        if (!principalId) {
            setState({ loading: false, policies: [], total: 0 });
            return;
        }

        setState(prev => ({ ...prev, loading: true }));
        describeAuthPolicies({
            principal_id: principalId,
            principal_type: principalType,
            offset: 0,
            limit: 100,
        }).then((ret) => {
            setState({ loading: false, policies: ret.content, total: ret.totalCount });
        }).catch((error) => {
            setState(prev => ({ ...prev, loading: false }));
            openErrNotification('请求错误', `获取关联策略失败, ${(error as Error).message}`);
        });
    }, [principalId, principalType]);

    const columns: TableProps['columns'] = [
        {
            colKey: 'name',
            title: '策略名称',
            cell: ({ row }) => (
                <Link theme="primary" onClick={() => setSelectedPolicy(row as PolicyRule)}>
                    {row.name || '-'}
                </Link>
            ),
        },
        {
            colKey: 'action',
            title: '行为',
            width: 120,
            cell: ({ row }) => row.action || '-',
        },
        {
            colKey: 'default_strategy',
            title: '默认策略',
            width: 120,
            cell: ({ row }) => (
                <Tag theme={row.default_strategy ? 'success' : 'default'} variant="outline">
                    {row.default_strategy ? '是' : '否'}
                </Tag>
            ),
        },
        {
            colKey: 'comment',
            title: '描述',
            ellipsis: true,
            cell: ({ row }) => row.comment || '-',
        },
        {
            colKey: 'time',
            title: '操作时间',
            width: 210,
            cell: ({ row }) => (
                <span>
                    修改: {row.mtime || '-'}
                    <br />
                    创建: {row.ctime || '-'}
                </span>
            ),
        },
    ];

    return (
        <>
            <Table
                rowKey="id"
                size="medium"
                tableLayout="auto"
                cellEmptyContent="-"
                loading={state.loading}
                columns={columns}
                data={state.policies}
                pagination={{
                    pageSize: 100,
                    total: state.total,
                    current: 1,
                    showPageSize: false,
                    showJumper: false,
                }}
            />
            <Drawer
                size="min(980px, 94vw)"
                header="策略详情"
                visible={Boolean(selectedPolicy)}
                showOverlay={false}
                onClose={() => setSelectedPolicy(null)}
                footer={false}
            >
                {selectedPolicy?.id && <PolicyDetailView policyId={selectedPolicy.id} />}
            </Drawer>
        </>
    );
};

export default React.memo(PrincipalPolicyTable);
