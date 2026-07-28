import React from 'react';
import { Drawer } from 'components/Fluent';

import style from './RuleDetailDrawer.module.less';

export const RuleDetailActionHostContext = React.createContext<HTMLElement | null>(null);

interface RuleDetailDrawerProps {
    visible: boolean;
    title: React.ReactNode;
    subtitle?: React.ReactNode;
    size?: string;
    onClose: () => void;
    children: React.ReactNode;
}

export const WIDE_RULE_DETAIL_DRAWER_SIZE = 'clamp(860px, 60vw, 1180px)';
const DEFAULT_RULE_DETAIL_DRAWER_SIZE = 'clamp(720px, 52vw, 960px)';

const RuleDetailDrawer: React.FC<RuleDetailDrawerProps> = ({ visible, title, subtitle, size = DEFAULT_RULE_DETAIL_DRAWER_SIZE, onClose, children }) => {
    const [actionHost, setActionHost] = React.useState<HTMLDivElement | null>(null);

    return (
        <Drawer
            attach="body"
            className={style.drawer}
            headerClassName={style.drawerHeader}
            headerTitleClassName={style.drawerHeaderTitle}
            bodyClassName={style.drawerBody}
            closeButtonClassName={style.closeButton}
            closeOnEscKeydown
            closeOnOverlayClick
            destroyOnClose
            footer={false}
            header={(
                <div className={style.header}>
                    <div className={style.heading}>
                        <div className={style.title}>{title}</div>
                        {subtitle && <div className={style.subtitle}>{subtitle}</div>}
                    </div>
                    <div ref={setActionHost} className={style.headerActions} role="toolbar" aria-label="规则操作" />
                </div>
            )}
            mode="overlay"
            placement="right"
            showOverlay
            size={size}
            visible={visible}
            onClose={onClose}
        >
            <RuleDetailActionHostContext.Provider value={actionHost}>
                <div className={style.body}>
                    {children}
                </div>
            </RuleDetailActionHostContext.Provider>
        </Drawer>
    );
};

export default React.memo(RuleDetailDrawer);
