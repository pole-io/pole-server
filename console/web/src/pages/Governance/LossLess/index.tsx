import React, { } from 'react';
import { Tabs } from 'components/Fluent';
import { TabPanel } from 'components/Fluent';
import LossLessTable from './LossLessTable';

export default React.memo(() => {

    return (
        <>
            <Tabs>
                <TabPanel label="无损发布" value="1">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <LossLessTable />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    )
});
