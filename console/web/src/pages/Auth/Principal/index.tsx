import React from 'react';
import { TabPanel, Tabs } from 'components/Fluent';
import { ResourceHeader } from 'components/ResourceLayout';
import { useSearchParams } from 'react-router-dom';

import UserTable from './UserTable';
import GroupsTable from './GroupTable';
import RoleTable from './RoleTable';
import style from './index.module.less';

export default React.memo(() => {
    const [searchParams, setSearchParams] = useSearchParams();
    const activeTab = searchParams.get('tab') || '1';

    return (
        <main className={style.authPage}>
            <ResourceHeader
                eyebrow="Access Control / Principals"
                title="身份主体"
                description="集中管理可登录或被授权的用户、用户组与角色，并从统一入口维护凭据和权限关系。"
            />

            <section className={style.authWorkspace}>
                <Tabs
                    className={style.authTabs}
                    value={activeTab}
                    onChange={(value: string | number) => {
                        const next = new URLSearchParams(searchParams);
                        next.set('tab', String(value));
                        if (String(value) !== '3') {
                            next.delete('roleMembers');
                        }
                        setSearchParams(next);
                    }}
                >
                    <TabPanel label="用户" value="1">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>维护登录身份、用户类型与访问 Token，查看用户当前关联的角色和策略。</p>
                            <div className={style.panelBody}><UserTable /></div>
                        </section>
                    </TabPanel>
                    <TabPanel label="用户组" value="2">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>按团队或职责聚合用户，通过用户组统一分配策略和管理访问 Token。</p>
                            <div className={style.panelBody}><GroupsTable /></div>
                        </section>
                    </TabPanel>
                    <TabPanel label="角色" value="3">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>可创建并维护自定义角色；管理员、资源全读和资源全写是内置角色，仅允许调整用户与用户组成员，角色定义和资源/API 权限固定。</p>
                            <div className={style.panelBody}><RoleTable /></div>
                        </section>
                    </TabPanel>
                </Tabs>
            </section>
        </main>
    );
});
