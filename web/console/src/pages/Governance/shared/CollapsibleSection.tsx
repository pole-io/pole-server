import React from 'react';
import { Button } from 'components/Fluent';
import { ChevronRightIcon } from 'components/Fluent/icons';

import shared from './governance.module.less';
import style from './CollapsibleSection.module.less';

interface CollapsibleSectionProps {
    collapsed: boolean;
    onCollapsedChange: (collapsed: boolean) => void;
    header: React.ReactNode;
    summary?: React.ReactNode;
    children: React.ReactNode;
    className?: string;
    headerClassName?: string;
    bodyClassName?: string;
    collapseLabel?: string;
}

const CollapsibleSection: React.FC<CollapsibleSectionProps> = ({
    collapsed,
    onCollapsedChange,
    header,
    summary,
    children,
    className,
    headerClassName,
    bodyClassName,
    collapseLabel = '基础信息',
}) => {
    const toggleCollapsed = () => {
        onCollapsedChange(!collapsed);
    };

    return (
        <section className={className || shared.section}>
            <Button
                variant="text"
                type="button"
                className={`${headerClassName || shared.sectionHeader} ${style.header} ${collapsed ? style.collapsedHeader : ''}`}
                aria-expanded={!collapsed}
                aria-label={`${collapsed ? '展开' : '收起'}${collapseLabel}`}
                onClick={toggleCollapsed}
            >
                <span className={style.headerContent}>{header}</span>
                <span className={style.headerEnd}>
                    {collapsed && summary && <span className={style.summary}>{summary}</span>}
                    <ChevronRightIcon className={`${style.chevron} ${collapsed ? '' : style.chevronOpen}`} />
                </span>
            </Button>
            <div className={bodyClassName || shared.sectionBody} hidden={collapsed}>
                {children}
            </div>
        </section>
    );
};

export default React.memo(CollapsibleSection);
