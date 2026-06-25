import React from 'react';
import { Button, Input, Popup, Select, TagInput } from 'tdesign-react';
import { AddIcon, CloseIcon } from 'tdesign-icons-react';

import { MatchLogic, MatchType, MatchTypeMap, MatchTypeOption } from 'services/types';
import { commaStringToTags, isTagInputMatchType, tagsToCommaString } from '../Router/routeEditorUtils';

import styles from './TrafficMatchConditionEditor.module.less';

export interface TrafficMatchConditionRow {
    paramType?: string;
    paramKey?: string;
    matchType?: string;
    matchValue?: string;
}

interface OptionItem {
    label: string;
    value: string;
}

interface TrafficMatchConditionEditorProps {
    rows: TrafficMatchConditionRow[];
    paramTypeOptions: OptionItem[];
    editable: boolean;
    relation?: string;
    relationEditable?: boolean;
    title?: React.ReactNode;
    hint?: React.ReactNode;
    addText?: string;
    minRows?: number;
    className?: string;
    extraControl?: React.ReactNode;
    onRelationChange?: (relation: string) => void;
    onRowChange: (index: number, row: TrafficMatchConditionRow) => void;
    onAdd: () => void;
    onRemove: (index: number) => void;
}

const relationOptions = [
    { label: 'AND', value: MatchLogic.AND },
    { label: 'OR', value: MatchLogic.OR },
];

const getOptionLabel = (options: OptionItem[], value?: string) => {
    const option = options.find(item => item.value === value);
    return option?.label || value || '-';
};

const renderReadonlyValue = (value?: string) => (
    <span className={styles.readonlyText}>{value || '-'}</span>
);

const TrafficMatchConditionEditor: React.FC<TrafficMatchConditionEditorProps> = ({
    rows,
    paramTypeOptions,
    editable,
    relation = MatchLogic.AND,
    relationEditable = true,
    title = '匹配条件',
    hint,
    addText = '添加匹配条件',
    minRows = 1,
    className,
    extraControl,
    onRelationChange,
    onRowChange,
    onAdd,
    onRemove,
}) => {
    const currentRows = rows.length ? rows : [];
    const relationText = relation === MatchLogic.OR ? '同一规则内条件按 OR 关系计算' : '同一规则内条件按 AND 关系计算';

    const updateRow = (index: number, patch: Partial<TrafficMatchConditionRow>) => {
        const row = currentRows[index] || {};
        onRowChange(index, { ...row, ...patch });
    };

    return (
        <div className={`${styles.conditionEditor} ${className || ''}`}>
            <div className={styles.conditionHead}>
                <div>
                    <div className={styles.conditionTitle}>{title}</div>
                    <div className={styles.conditionHint}>{hint || relationText}</div>
                </div>
                <div className={styles.headActions}>
                    {extraControl}
                    <div className={styles.segmented}>
                        {relationOptions.map(option => {
                            const active = relation === option.value;
                            return (
                                <button
                                    key={option.value}
                                    className={active ? styles.segmentButtonActive : styles.segmentButton}
                                    type="button"
                                    disabled={!editable || !relationEditable}
                                    onClick={() => onRelationChange?.(option.value)}
                                >
                                    {option.label}
                                </button>
                            );
                        })}
                    </div>
                </div>
            </div>

            <div className={styles.conditionTable}>
                <div className={styles.tableHeader}>参数类型</div>
                <div className={styles.tableHeader}>参数键</div>
                <div className={styles.tableHeader}>匹配类型</div>
                <div className={styles.tableHeader}>匹配值</div>
                <div className={styles.tableHeader}>操作</div>
                {currentRows.map((row, index) => {
                    const matchType = row.matchType || MatchType.EXACT;
                    return (
                        <React.Fragment key={`${row.paramType || 'param'}-${row.paramKey || 'empty'}-${index}`}>
                            <div className={styles.tableCell}>
                                {editable ? (
                                    <Select
                                        filterable
                                        options={paramTypeOptions}
                                        value={row.paramType}
                                        onChange={(value) => updateRow(index, { paramType: value as string })}
                                    />
                                ) : renderReadonlyValue(getOptionLabel(paramTypeOptions, row.paramType))}
                            </div>
                            <div className={styles.tableCell}>
                                {editable ? (
                                    <Input
                                        value={row.paramKey || ''}
                                        placeholder="请输入参数键"
                                        onChange={(value) => updateRow(index, { paramKey: value })}
                                    />
                                ) : renderReadonlyValue(row.paramKey)}
                            </div>
                            <div className={styles.tableCell}>
                                {editable ? (
                                    <Select
                                        filterable
                                        options={MatchTypeOption}
                                        value={matchType}
                                        onChange={(value) => updateRow(index, { matchType: value as string })}
                                    />
                                ) : renderReadonlyValue(MatchTypeMap[matchType as MatchType] || matchType)}
                            </div>
                            <div className={styles.tableCell}>
                                {editable ? (
                                    isTagInputMatchType(matchType) ? (
                                        <TagInput
                                            value={commaStringToTags(row.matchValue)}
                                            placeholder="请输入多个值，回车分隔"
                                            onChange={(value) => updateRow(index, { matchValue: tagsToCommaString(value as Array<string | number>) })}
                                        />
                                    ) : (
                                        <Input
                                            value={row.matchValue || ''}
                                            placeholder="请输入匹配值"
                                            onChange={(value) => updateRow(index, { matchValue: value })}
                                        />
                                    )
                                ) : renderReadonlyValue(row.matchValue)}
                            </div>
                            <div className={`${styles.tableCell} ${styles.actionCell}`}>
                                {editable && (
                                    <Popup trigger="hover" content={currentRows.length <= minRows ? `至少保留 ${minRows} 条条件` : '删除匹配条件'}>
                                        <Button
                                            shape="circle"
                                            variant="text"
                                            disabled={currentRows.length <= minRows}
                                            onClick={() => onRemove(index)}
                                        >
                                            <CloseIcon />
                                        </Button>
                                    </Popup>
                                )}
                            </div>
                        </React.Fragment>
                    );
                })}
            </div>

            {editable && (
                <Button className={styles.addButton} variant="text" icon={<AddIcon />} onClick={onAdd}>
                    {addText}
                </Button>
            )}
        </div>
    );
};

export default React.memo(TrafficMatchConditionEditor);
