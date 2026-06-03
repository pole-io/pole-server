import React, { } from 'react';
import { Tabs } from 'tdesign-react';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';
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