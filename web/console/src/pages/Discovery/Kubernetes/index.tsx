import React, { memo, useRef } from 'react';
import ErrorPage, { ECode } from 'components/ErrorPage';


export default memo(() => {
  return (
    <ErrorPage code={ECode.unimplemented} />
  );
});
