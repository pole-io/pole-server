import React, { } from 'react';
import { Tabs } from 'tdesign-react';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';
import PolicyTable from './PolicyTable';

export default React.memo(() => {

    return (
        <>
            <Tabs>
                <TabPanel label="自定义策略" value="1">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                      <PolicyTable type="custom"  />
                    </div>
                </TabPanel>
                <TabPanel label="默认策略" value="2">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                      <PolicyTable type='default' />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    )
});