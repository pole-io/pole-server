import React, { memo, useRef } from 'react';
import ErrorPage, { ECode } from 'components/ErrorPage';


interface ITemplateProps {

}


const TemplateTable: React.FC<ITemplateProps> = ({ }) => {
    return (
        <ErrorPage code={ECode.unimplemented} />
    );
}

export default memo(TemplateTable);
