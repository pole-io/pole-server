import React, {  } from 'react';
import { useNavigate } from 'react-router-dom';

import ErrorPage, { ECode } from 'components/ErrorPage';
import { useAppDispatch } from 'modules/store';

interface IServiceSubscribeProps {

}

const ServiceSubscribeTable: React.FC<IServiceSubscribeProps> = ({ }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();

    return (
        <ErrorPage code={ECode.unimplemented} />
    )
}

export default React.memo(ServiceSubscribeTable);