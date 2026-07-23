import React from 'react';
import ConfigGroupTable from './group';
import style from './index.module.less';

export default React.memo(() => {
    return (
        <div className={style.groupListPage}>
            <ConfigGroupTable />
        </div>
    )
});
