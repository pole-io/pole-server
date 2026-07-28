import React from 'react';
import { Drawer } from 'components/Fluent';
import { Op } from 'services/types';
import ServiceForm from './ServiceForm';

interface IServiceEditorProps {
    op: Op;
    visible: boolean;
    closeDrawer: () => void;
}

const ServiceEditor: React.FC<IServiceEditorProps> = ({ visible, closeDrawer }) => {
    return (
        <Drawer
            size="min(720px, 94vw)"
            header="创建服务"
            visible={visible}
            showOverlay
            closeOnOverlayClick
            closeOnEscKeydown
            destroyOnClose
            onClose={closeDrawer}
        >
            <ServiceForm
                mode="create"
                service={null}
                onSubmitted={closeDrawer}
                onCancel={closeDrawer}
            />
        </Drawer>
    );
};

export default React.memo(ServiceEditor);
