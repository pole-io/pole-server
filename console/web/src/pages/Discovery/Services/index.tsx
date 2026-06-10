import React, { } from 'react';
import { Tabs } from 'tdesign-react';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';

import ServiceAliasTable from './alias';
import ServicesTable from './services';
import style from './index.module.less';

export default React.memo(() => {

    return (
        <div className={style.page}>
            <section className={style.header}>
                <div>
                    <div className={style.eyebrow}>Service Registry / Discovery</div>
                    <h2>注册发现</h2>
                    <p>查看服务和别名，确认命名空间、可见范围以及实例健康状态。</p>
                </div>
            </section>
            <Tabs className={style.registryTabs}>
                <TabPanel label="服务" value="1">
                    <div className={style.tabContent}>
                        <ServicesTable />
                    </div>
                </TabPanel>
                <TabPanel label="别名" value="2">
                    <div className={style.tabContent}>
                        <ServiceAliasTable />
                    </div>
                </TabPanel>
            </Tabs>
        </div>
    )
});
