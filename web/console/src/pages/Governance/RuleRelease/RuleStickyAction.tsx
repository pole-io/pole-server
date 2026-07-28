import React from 'react';
import { createPortal } from 'react-dom';
import { Button } from 'components/Fluent';

import { RuleDetailActionHostContext } from './RuleDetailDrawer';

interface RuleStickyActionProps {
    label: string;
    icon: React.ReactNode;
    onClick: () => void;
    disabled?: boolean;
    loading?: boolean;
}

const primaryActions = new Set(['保存', '发布']);

const RuleStickyAction: React.FC<RuleStickyActionProps> = ({ label, icon, onClick, disabled, loading }) => {
    const actionHost = React.useContext(RuleDetailActionHostContext);
    const action = (
        <Button
            size="small"
            theme={primaryActions.has(label) ? 'primary' : 'default'}
            variant={primaryActions.has(label) ? 'base' : 'outline'}
            icon={icon}
            disabled={disabled}
            loading={loading}
            onClick={(event) => {
                event.stopPropagation();
                onClick();
            }}
        >
            {label}
        </Button>
    );

    return actionHost ? createPortal(action, actionHost) : action;
};

export default React.memo(RuleStickyAction);
