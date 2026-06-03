import React, { } from 'react';
import { Tabs } from 'tdesign-react';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';

import ServiceAliasTable from './alias';
import ServicesTable from './services';
import GatewayTable from './gateway';

export default React.memo(() => {

    return (
        <>
            <Tabs>
                <TabPanel label="服务" value="1">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <ServicesTable />
                    </div>
                </TabPanel>
                <TabPanel label="别名" value="2">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <ServiceAliasTable />
                    </div>
                </TabPanel>
                <TabPanel label="网关" value="3">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <GatewayTable />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    )
});