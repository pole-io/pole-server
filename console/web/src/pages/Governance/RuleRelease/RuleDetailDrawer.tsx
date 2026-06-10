import React from 'react';
import { Drawer } from 'tdesign-react';

import style from './RuleDetailDrawer.module.less';

interface RuleDetailDrawerProps {
    visible: boolean;
    title: React.ReactNode;
    subtitle?: React.ReactNode;
    onClose: () => void;
    children: React.ReactNode;
}

const RuleDetailDrawer: React.FC<RuleDetailDrawerProps> = ({ visible, title, subtitle, onClose, children }) => (
    <Drawer
        attach="body"
        className={style.drawer}
        closeOnEscKeydown
        closeOnOverlayClick
        destroyOnClose
        footer={false}
        header={(
            <div className={style.header}>
                <div className={style.title}>{title}</div>
                {subtitle && <div className={style.subtitle}>{subtitle}</div>}
            </div>
        )}
        mode="overlay"
        placement="right"
        showOverlay
        size="min(860px, 88vw)"
        visible={visible}
        onClose={onClose}
    >
        <div className={style.body}>
            {children}
        </div>
    </Drawer>
);

export default React.memo(RuleDetailDrawer);
