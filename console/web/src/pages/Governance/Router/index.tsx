import React, { } from 'react';
import { Tabs } from 'tdesign-react';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';
import CustomRoute from './CustomRoute';
import LaneGroupTable from './LaneGroupTable';
import NearbyRoute from './NearbyRoute';

export default React.memo(() => {

    return (
        <>
            <Tabs>
                <TabPanel label="自定义路由" value="1">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <CustomRoute />
                    </div>
                </TabPanel>
                {/* <TabPanel label="就近路由" value="2">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <NearbyRoute />
                    </div>
                </TabPanel> */}
                <TabPanel label="全链路灰度" value="3">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <LaneGroupTable />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    )
});