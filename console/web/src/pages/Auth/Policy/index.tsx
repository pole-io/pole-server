import React from 'react';
import { TabPanel, Tabs } from 'components/Fluent';
import { ResourceHeader } from 'components/ResourceLayout';

import PolicyTable from './PolicyTable';
import style from './index.module.less';

export default React.memo(() => {
    return (
        <main className={style.authPage}>
            <ResourceHeader
                eyebrow="Access Control / Policies"
                title="访问策略"
                description="集中查看主体与资源之间的授权规则，区分人工维护的策略和系统生成的默认权限。"
            />

            <section className={style.authWorkspace}>
                <Tabs className={style.authTabs}>
                    <TabPanel label="自定义策略" value="1">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>按业务边界为用户、用户组或角色配置资源范围与允许执行的操作。</p>
                            <div className={style.panelBody}><PolicyTable type="custom" /></div>
                        </section>
                    </TabPanel>
                    <TabPanel label="默认策略" value="2">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>查看系统为主体自动维护的基础权限，确认默认授权的来源与生效范围。</p>
                            <div className={style.panelBody}><PolicyTable type="default" /></div>
                        </section>
                    </TabPanel>
                </Tabs>
            </section>
        </main>
    );
});
