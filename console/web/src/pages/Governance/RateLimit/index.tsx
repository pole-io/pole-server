import React, { useState, useEffect } from 'react';
import { Tabs } from 'components/Fluent';

import RateLimitTable from './RateLimitTable';
import { LimitType } from 'services/ratelimit';
import RateLimitClusterView from './RateLimitClusterView';

const { TabPanel } = Tabs;

export default React.memo(() => {
    const [curTab, setCurTab] = useState('1');

    return (
        <>
            <Tabs defaultValue="1" onChange={(value) => {
                setCurTab(value as string);
            }}>
                <TabPanel label="单机" value="1">
                    <div style={{ margin: 20 }}>
                        {curTab === '1' && <RateLimitTable limitType={LimitType.LOCAL} />}
                    </div>
                </TabPanel>
                <TabPanel label="分布式" value="2">
                    <div style={{ margin: 20 }}>
                        {curTab === '2' && <RateLimitTable limitType={LimitType.GLOBAL} />}
                    </div>
                </TabPanel>
                <TabPanel label="限流集群" value="3">
                    <div style={{ margin: 20 }}>
                        <RateLimitClusterView />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    );
});