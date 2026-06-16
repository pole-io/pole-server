import React from 'react';
import { Drawer } from 'tdesign-react';

import style from './RuleDetailDrawer.module.less';

interface RuleDetailDrawerProps {
    visible: boolean;
    title: React.ReactNode;
    subtitle?: React.ReactNode;
    size?: string;
    onClose: () => void;
    children: React.ReactNode;
}

const RuleDetailDrawer: React.FC<RuleDetailDrawerProps> = ({ visible, title, subtitle, size = 'min(1080px, 92vw)', onClose, children }) => (
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
        size={size}
        visible={visible}
        onClose={onClose}
    >
        <div className={style.body}>
            {children}
        </div>
    </Drawer>
);

export default React.memo(RuleDetailDrawer);
