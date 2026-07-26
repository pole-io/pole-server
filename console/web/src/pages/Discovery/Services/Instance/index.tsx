import React from 'react';
import { Breadcrumb, Loading, Tabs } from 'components/Fluent';
import { useNavigate } from 'react-router-dom';

import InstanceTable from './InstanceTable';
import ServiceDetail from './ServiceDetail';
import ServiceContractPanel from './ServiceContractPanel';
import SubscribeTable from './SubscribeTable';
import ServiceAliasTable from '../alias';
import style from './index.module.less';

const { TabPanel } = Tabs;
const { BreadcrumbItem } = Breadcrumb;
const GovernanceWorkbench = React.lazy(() => import('pages/Governance/Workbench'));

export default React.memo(() => {
    const navigate = useNavigate()
    const urlParams = new URLSearchParams(window.location.search);
    const namespace = urlParams.get('namespace');
    const serviceName = urlParams.get('service');
    const editMode = urlParams.get('mode') === 'edit';

    const [activeTab, setActiveTab] = React.useState('0');

    return (
        <>
            <Breadcrumb maxItemWidth="200px">
                <BreadcrumbItem onClick={() => {
                    navigate(-1);
                }}>
                    {namespace}
                </BreadcrumbItem>
                <BreadcrumbItem>
                    {serviceName}
                </BreadcrumbItem>
            </Breadcrumb>
            <Tabs style={{ marginTop: 20 }} value={activeTab} onChange={(v) => setActiveTab(v as string)}>
                <TabPanel value={"0"} label="服务详情">
                    {activeTab === '0' && (
                        <ServiceDetail
                            namespace={namespace || ''}
                            serviceName={serviceName || ''}
                            onTabChange={setActiveTab}
                            initialEdit={editMode}
                        />
                    )}
                </TabPanel>
                <TabPanel value={"1"} label="服务实例">
                    {activeTab === '1' && (
                        <InstanceTable
                            namespace={namespace || ''}
                            serviceName={serviceName || ''}
                        />
                    )}
                </TabPanel>
                <TabPanel value={"2"} label="服务别名">
                    {activeTab === '2' && (
                        <ServiceAliasTable
                            namespace={namespace || ''}
                            serviceName={serviceName || ''}
                            embedded
                        />
                    )}
                </TabPanel>
                <TabPanel value={"3"} label="服务订阅">
                    {activeTab === '3' && (
                        <SubscribeTable
                            namespace={namespace || ''}
                            serviceName={serviceName || ''}
                        />
                    )}
                </TabPanel>
                <TabPanel value={"4"} label="服务契约">
                    {activeTab === '4' && (
                        <ServiceContractPanel
                            namespace={namespace || ''}
                            serviceName={serviceName || ''}
                        />
                    )}
                </TabPanel>
                <TabPanel value={"5"} label="流量治理">
                    {activeTab === '5' && (
                        <React.Suspense fallback={<Loading text="加载流量治理中..." />}>
                            <section
                                className={style.serviceGovernance}
                                title={`${namespace || '-'}/${serviceName || '-'}`}
                            >
                                <GovernanceWorkbench
                                    embedded
                                    serviceContext={{ namespace: namespace || '', service: serviceName || '' }}
                                />
                            </section>
                        </React.Suspense>
                    )}
                </TabPanel>
            </Tabs>
        </>
    )
})
