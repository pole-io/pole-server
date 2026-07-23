import React from 'react';

import { RuleDetailActionHostContext } from './RuleDetailDrawer';
import style from './RuleDetailDrawer.module.less';

interface RuleDetailFrameProps {
    title: React.ReactNode;
    subtitle?: React.ReactNode;
    children: React.ReactNode;
}

const RuleDetailFrame: React.FC<RuleDetailFrameProps> = ({ title, subtitle, children }) => {
    const [actionHost, setActionHost] = React.useState<HTMLDivElement | null>(null);
    const titleText = typeof title === 'string' ? title : undefined;
    const subtitleText = typeof subtitle === 'string' ? subtitle : undefined;

    return (
        <section className={style.pageFrame}>
            <header className={style.pageHeader}>
                <div className={style.header}>
                    <div className={style.heading}>
                        <div className={style.title} title={titleText}>{title}</div>
                        {subtitle && <div className={style.subtitle} title={subtitleText}>{subtitle}</div>}
                    </div>
                    <div ref={setActionHost} className={style.headerActions} role="toolbar" aria-label="规则操作" />
                </div>
            </header>
            <RuleDetailActionHostContext.Provider value={actionHost}>
                <div className={style.pageBody}>
                    {children}
                </div>
            </RuleDetailActionHostContext.Provider>
        </section>
    );
};

export default React.memo(RuleDetailFrame);
