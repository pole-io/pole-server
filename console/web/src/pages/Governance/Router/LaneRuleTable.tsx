import React, { useState } from 'react';
import { Link, Table, Button, PrimaryTableProps, Tooltip, Space, Row, Col, TableRowData, Popconfirm, Tag } from 'tdesign-react';
import { DeleteIcon, EditIcon } from 'tdesign-icons-react';

import { useAppDispatch, useAppSelector } from 'modules/store';
import Text from 'components/Text';
import { Op } from 'services/types';
import { LaneRuleView } from 'services/lane';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import LaneRuleEditor from './LaneRuleEditor';
import { removeLaneRules, resetLaneRule, viewLaneRule } from 'modules/governance/lane_rule';
import { listOneLaneGroup, selectLaneGroup } from 'modules/governance/lane_group';
import { LimitArgumentsTypeMap } from 'services/ratelimit';

const formatLaneMatchArguments = (row: LaneRuleView) => {
    const args = row.trafficMatchRule?.arguments || [];
    if (!args.length) return '-';
    return args.map((arg) => {
        const type = LimitArgumentsTypeMap[arg.type as keyof typeof LimitArgumentsTypeMap] || arg.type || '参数';
        const value = arg.value?.value || '-';
        return `${type} ${arg.key || '-'} = ${value}`;
    }).join(' / ');
};

interface ILaneRuleTableProps {
    groupId?: string; // 可选的规则ID，用于编辑时传入
}

const columns = (handleOpRule: (row: TableRowData, op: Op) => void, redirect: (row: TableRowData) => void): PrimaryTableProps['columns'] => [
    {
        colKey: 'name',
        title: '名称',
        cell: ({ row }) => <Link
            theme="primary"
            onClick={() => { redirect(row) }}
        >{row.name}</Link>,
    },
    {
        colKey: 'enable',
        title: '状态',
        cell: ({ row: { enable } }) => (<Tag theme={enable ? 'success' : 'danger'} variant="outline">{enable ? '启用' : '禁用'}</Tag>),
    },
    {
        colKey: 'description',
        title: '描述',
        cell: ({ row: { description } }) => <Text>{description || '-'}</Text>,
    },
    {
        colKey: 'lane_label',
        title: '泳道标签',
        cell: ({ row }) => <div>{row.labelKey}: {row.defaultLabelValue}</div>,
    },
    {
        colKey: 'match',
        title: '匹配条件',
        ellipsis: true,
        cell: ({ row }) => <Text>{formatLaneMatchArguments(row as LaneRuleView)}</Text>,
    },
    {
        colKey: 'time',
        title: '操作时间',
        cell: ({ row: { ctime, mtime } }: TableRowData) => <Text>修改: {mtime}<br />创建: {ctime}</Text>,
    },
    {
        colKey: 'action',
        title: '操作',
        cell: ({ row }) => {
            return (
                <Space>
                    <Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
                        <Button
                            shape="square"
                            variant="text"
                            disabled={row.editable === false}
                            onClick={() => handleOpRule(row, 'edit')}>
                            <EditIcon />
                        </Button>
                    </Tooltip>
                    <Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
                        <Popconfirm
                            content="确认删除吗"
                            destroyOnClose
                            placement="top"
                            showArrow
                            theme="default"
                            onConfirm={() => {
                                handleOpRule(row, 'delete');
                            }}
                        >
                            <Button shape="square" variant="text" disabled={row.deleteable === false}>
                                <DeleteIcon />
                            </Button>
                        </Popconfirm>
                    </Tooltip>
                </Space>
            )
        },
    },
]

const LaneRuleTable: React.FC<ILaneRuleTableProps> = ({ groupId }) => {
    const dispatch = useAppDispatch();

    const laneGroupState = useAppSelector(selectLaneGroup);
    const { editGroup, loading } = laneGroupState;

    React.useEffect(() => {
        if (groupId && groupId.length > 0) {
            loadLaneRules(groupId);
        }
    }, [groupId])

    const loadLaneRules = (groupId: string) => {
        dispatch(listOneLaneGroup({ id: groupId }))
            .then((res) => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification("请求失败", `获取泳道规则失败: ${res.payload as string || '未知'}`);
                }
            });
    }

    // 合并编辑相关状态
    const [editorState, setEditorState] = useState<{
        visible: boolean;
        authorizeVisible: boolean;
        mode: Op;
    }>({ visible: false, mode: 'create', authorizeVisible: false });

    const handleOpRule = (row: TableRowData, op: Op) => {
        // 处理操作
        switch (op) {
            case 'create':
                setEditorState(prev => ({ ...prev, visible: true, mode: 'create' }));
                break;
            case 'view':
            case 'edit':
                dispatch(viewLaneRule({ ...row as LaneRuleView }))
                setEditorState(prev => ({ ...prev, visible: true, mode: op }));
                break;
            case 'delete':
                dispatch(removeLaneRules({
                    ids: [{
                        id: row.id,
                        groupName: row.groupName || editGroup?.name || '',
                    }]
                })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification("请求成功", "删除泳道规则成功");
                        loadLaneRules(editGroup?.id || '');
                    } else {
                        openErrNotification("请求失败", `删除泳道规则失败: ${res.payload as string || '未知'}`);
                    }
                });
                break;
        }
    }

    const table = (
        <>
            <Row justify='space-between' className={style.toolBar}>
                <Col>
                    <Row gutter={8} align='middle'>
                        <Col>
                            <Button onClick={(v) => {
                                handleOpRule({}, 'create')
                            }}>新建</Button>
                        </Col>
                    </Row>
                </Col>
            </Row>
            {editorState.visible && (
                <LaneRuleEditor
                    op={editorState.mode}
                    visible={editorState.visible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, visible: false }))
                        dispatch(resetLaneRule());
                        loadLaneRules(editGroup?.id || '');
                    }}
                />
            )}
            <Table
                data={editGroup?.rules || []}
                columns={columns(handleOpRule, (row: TableRowData) => {
                    handleOpRule(row, 'view');
                })}
                loading={loading}
                rowKey="id"
                size={"large"}
                tableLayout={'auto'}
                cellEmptyContent={'-'}
            />
        </>
    )

    return (
        <div style={{ padding: 24 }}>
            {table}
        </div>
    )
}

export default React.memo(LaneRuleTable);
