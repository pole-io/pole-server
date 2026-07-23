import React from 'react';

import TrafficMatchConditionEditor, { TrafficMatchConditionRow } from 'pages/Governance/shared/TrafficMatchConditionEditor';
import { ClientLabelType, ClientLabelTypeOption, MatcheLabel, MatchLogic, MatchType, MatchValueType } from 'services/types';

interface GrayRuleEditorProps {
    rows: TrafficMatchConditionRow[];
    editable: boolean;
    onChange: (rows: TrafficMatchConditionRow[]) => void;
}

export const defaultGrayRuleRow = (): TrafficMatchConditionRow => ({
    paramType: ClientLabelType.CLIENT_IP,
    matchType: MatchType.EXACT,
    matchValue: '',
});

export const grayRowsToBetaLabels = (rows: TrafficMatchConditionRow[]): MatcheLabel[] => (
    rows
        .map((row) => ({
            key: row.paramType || ClientLabelType.CLIENT_IP,
            value: {
                type: row.matchType || MatchType.EXACT,
                value_type: MatchValueType.TEXT,
                value: row.matchValue || '',
            },
        }))
        .filter(label => label.key && label.value.value)
);

export const betaLabelsToGrayRows = (labels?: MatcheLabel[]): TrafficMatchConditionRow[] => {
    if (!labels?.length) {
        return [defaultGrayRuleRow()];
    }
    return labels.map((label) => ({
        paramType: label.key || ClientLabelType.CLIENT_IP,
        matchType: label.value?.type || MatchType.EXACT,
        matchValue: label.value?.value || '',
    }));
};

const GrayRuleEditor: React.FC<GrayRuleEditorProps> = ({ rows, editable, onChange }) => {
    const currentRows = rows.length ? rows : [defaultGrayRuleRow()];

    const updateRow = (index: number, row: TrafficMatchConditionRow) => {
        const nextRows = [...currentRows];
        nextRows[index] = row;
        onChange(nextRows);
    };

    return (
        <TrafficMatchConditionEditor
            rows={currentRows}
            editable={editable}
            relation={MatchLogic.AND}
            relationEditable={false}
            title="灰度规则"
            hint="客户端需同时满足全部灰度规则"
            addText="添加灰度规则"
            minRows={1}
            showParamKey={false}
            paramTypeOptions={ClientLabelTypeOption}
            onRowChange={updateRow}
            onAdd={() => onChange([...currentRows, defaultGrayRuleRow()])}
            onRemove={(index) => onChange(currentRows.filter((_, idx) => idx !== index))}
        />
    );
};

export default React.memo(GrayRuleEditor);
