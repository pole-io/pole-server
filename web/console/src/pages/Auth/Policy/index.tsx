import React from 'react';
import { TabPanel, Tabs } from 'components/Fluent';
import { ResourceHeader } from 'components/ResourceLayout';
import { describeAuthPolicies } from 'services/auth_policy';

import PolicyTable from './PolicyTable';
import style from './index.module.less';

export default React.memo(() => {
    const [totals, setTotals] = React.useState<{
        custom: number | null;
        defaults: number | null;
    }>({
        custom: null,
        defaults: null,
    });
    const [totalsLoading, setTotalsLoading] = React.useState(true);

    const loadTotals = React.useCallback(async () => {
        setTotalsLoading(true);
        const [custom, defaults] = await Promise.allSettled([
            describeAuthPolicies({ offset: 0, limit: 1, default: 'false' }),
            describeAuthPolicies({ offset: 0, limit: 1, default: 'true' }),
        ]);
        setTotals((current) => ({
            custom: custom.status === 'fulfilled' ? custom.value.totalCount : current.custom,
            defaults: defaults.status === 'fulfilled' ? defaults.value.totalCount : current.defaults,
        }));
        setTotalsLoading(false);
    }, []);

    React.useEffect(() => {
        void loadTotals();
    }, [loadTotals]);

    const updateCustomTotal = React.useCallback((custom: number) => {
        setTotals((current) => ({ ...current, custom }));
    }, []);
    const updateDefaultTotal = React.useCallback((defaults: number) => {
        setTotals((current) => ({ ...current, defaults }));
    }, []);

    return (
        <main className={style.authPage}>
            <ResourceHeader
                density="compact"
                placement="app-header"
                eyebrow="Access Control / Policies"
                title="访问策略"
                description="集中查看主体与资源之间的授权规则，区分人工维护的策略和系统生成的默认权限。"
            />

            <section className={style.policyMetricRail} aria-label="访问策略统计" aria-busy={totalsLoading}>
                <div className={style.policyMetricItem}>
                    <span>自定义策略数</span>
                    <strong>{totals.custom ?? '—'}</strong>
                    <small>人工维护的授权规则</small>
                </div>
                <div className={style.policyMetricItem}>
                    <span>默认策略数</span>
                    <strong>{totals.defaults ?? '—'}</strong>
                    <small>系统自动维护的基础权限</small>
                </div>
            </section>

            <section className={style.authWorkspace}>
                <Tabs className={style.authTabs}>
                    <TabPanel label="自定义策略" value="1">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>按业务边界为用户、用户组或角色配置资源范围与允许执行的操作。</p>
                            <div className={style.panelBody}>
                                <PolicyTable
                                    type="custom"
                                    onTotalChange={updateCustomTotal}
                                    onTotalsRefresh={loadTotals}
                                />
                            </div>
                        </section>
                    </TabPanel>
                    <TabPanel label="默认策略" value="2">
                        <section className={style.authPanel}>
                            <p className={style.panelHint}>查看系统为主体自动维护的基础权限，确认默认授权的来源与生效范围。</p>
                            <div className={style.panelBody}>
                                <PolicyTable
                                    type="default"
                                    onTotalChange={updateDefaultTotal}
                                    onTotalsRefresh={loadTotals}
                                />
                            </div>
                        </section>
                    </TabPanel>
                </Tabs>
            </section>
        </main>
    );
});
