import React from 'react';

interface RuleStickyActionProps {
    label: string;
    icon: React.ReactNode;
    onClick: () => void;
}

const actionStyle: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    height: 32,
    margin: '0 -10px',
    padding: '0 10px',
    cursor: 'pointer',
};

const RuleStickyAction: React.FC<RuleStickyActionProps> = ({ label, icon, onClick }) => (
    <span
        style={actionStyle}
        onClick={(event) => {
            event.stopPropagation();
            onClick();
        }}
    >
        {icon}
        <span>{label}</span>
    </span>
);

export default React.memo(RuleStickyAction);
