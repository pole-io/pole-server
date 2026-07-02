import React from 'react';
import { Breadcrumb, Tabs } from 'tdesign-react';
import { useNavigate } from 'react-router-dom';

import InstanceTable from './InstanceTable';
import ServiceDetail from './ServiceDetail';
import SubscribeTable from './SubscribeTable';
import ServiceAliasTable from '../alias';

const { TabPanel } = Tabs;
const { BreadcrumbItem } = Breadcrumb;

export default React.memo(() => {
    const navigate = useNavigate()
    const urlParams = new URLSearchParams(window.location.search);
    const namespace = urlParams.get('namespace');
    const serviceName = urlParams.get('service');

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
            </Tabs>
        </>
    )
})
