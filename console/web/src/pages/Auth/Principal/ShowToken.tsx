import React, { useEffect, useState } from 'react';
import { TableRowData, Dialog, Input } from 'tdesign-react';


interface IShowTokenProps {
    row: TableRowData;
    visible: boolean;
    close: () => void;
}

const ShowToken: React.FC<IShowTokenProps> = ({ row, visible, close }) => {
    return (
        <Dialog
            header={`查看 ${row.name} 的资源访问 token`}
            visible={visible}
            confirmOnEnter
            onClose={close}
            cancelBtn={null}
            confirmBtn={null}
            width={'600px'}
        >
            <Input type="password" size='large' defaultValue={row.auth_token} readonly={true} />
        </Dialog>
    )
}

export default React.memo(ShowToken);
