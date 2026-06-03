import React, { memo, useRef } from 'react';
import ErrorPage, { ECode } from 'components/ErrorPage';


interface IGatewayProps {

}


const GatewayTable: React.FC<IGatewayProps> = ({ }) => {
    return (
        <ErrorPage code={ECode.unimplemented} />
    );
}

export default memo(GatewayTable);
