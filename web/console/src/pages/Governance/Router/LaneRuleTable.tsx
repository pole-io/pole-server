import React, { useState } from 'react';
import { Button, Tooltip, Space, TableRowData, Popconfirm, Tag, Link } from 'components/Fluent';
import { AddIcon, DeleteIcon, EditIcon } from 'components/Fluent/icons';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { ResourceToolbar } from 'components/ResourceLayout';
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

const isZeroTime = (value?: string) => !value || value.startsWith('0001-01-01') || value.startsWith('1970-01-01');

const renderTimeMeta = (row: LaneRuleView) => {
    const items = [
        { label: '修改', value: row.mtime },
        { label: '创建', value: row.ctime },
    ].filter((item) => !isZeroTime(item.value));

    if (!items.length) return null;

    return (
        <div className={style.timeCell}>
            {items.map((item) => (
                <span className={style.timeRow} key={`${item.label}-${item.value}`}>
                    <span className={style.timeLabel}>{item.label}</span>
                    <span className={style.timeValue}>{item.value}</span>
                </span>
            ))}
        </div>
    );
};

interface ILaneRuleTableProps {
    groupId?: string; // 可选的规则ID，用于编辑时传入
}

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

    const renderActions = (row: LaneRuleView) => (
        <Space size={4} className={style.laneRuleActions}>
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
    );

    const renderRule = (rule: LaneRuleView) => {
        const matchText = formatLaneMatchArguments(rule);
        const timeMeta = renderTimeMeta(rule);

        return (
            <div className={style.laneRuleItem} key={rule.id || rule.name}>
                <div className={style.laneRuleMain}>
                    <div className={style.laneRuleIdentity}>
                        <Link theme="primary" className={style.laneRuleName} onClick={() => handleOpRule(rule, 'view')}>
                            {rule.name || '未命名泳道'}
                        </Link>
                        <Tag theme={rule.enable ? 'success' : 'danger'} variant="outline">
                            {rule.enable ? '启用' : '禁用'}
                        </Tag>
                    </div>
                    <div className={style.laneRuleDescription}>{rule.description || '暂无描述'}</div>
                    <div className={style.laneRuleMetaGrid}>
                        <div className={style.laneRuleMeta}>
                            <span className={style.laneRuleMetaLabel}>泳道标签</span>
                            <Tag variant="light-outline" theme="primary">{`${rule.labelKey || 'lane'}: ${rule.defaultLabelValue || '-'}`}</Tag>
                        </div>
                        <div className={style.laneRuleMeta}>
                            <span className={style.laneRuleMetaLabel}>匹配条件</span>
                            <Text>{matchText}</Text>
                        </div>
                        {timeMeta && (
                            <div className={style.laneRuleMeta}>
                                <span className={style.laneRuleMetaLabel}>操作时间</span>
                                {timeMeta}
                            </div>
                        )}
                    </div>
                </div>
                {renderActions(rule)}
            </div>
        );
    };

    const rules = editGroup?.rules || [];

    const table = (
        <>
            <ResourceToolbar
                density="compact"
                title="泳道规则清单"
                count={`共 ${rules.length} 条`}
                filters={(
                    <Button theme="primary" icon={<AddIcon />} onClick={() => {
                        handleOpRule({}, 'create')
                    }}>新建泳道规则</Button>
                )}
            />
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
            <div className={style.laneRuleList} aria-busy={loading}>
                {loading && <div className={style.laneRuleEmpty}>加载中...</div>}
                {!loading && rules.length === 0 && <div className={style.laneRuleEmpty}>暂无泳道规则</div>}
                {!loading && rules.map(renderRule)}
            </div>
        </>
    )

    return (
        <div className={style.laneRuleTable}>
            {table}
        </div>
    )
}

export default React.memo(LaneRuleTable);
