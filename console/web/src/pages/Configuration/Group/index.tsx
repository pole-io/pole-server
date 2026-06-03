import React, { } from 'react';
import { Tabs } from 'tdesign-react';
import TabPanel from 'tdesign-react/es/tabs/TabPanel';


import ConfigGroupTable from './group';
import TemplateTable from './template';

export default React.memo(() => {

    return (
        <>
            <Tabs>
                <TabPanel label="分组" value="1">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <ConfigGroupTable />
                    </div>
                </TabPanel>
                <TabPanel label="模板" value="2">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <TemplateTable />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    )
});