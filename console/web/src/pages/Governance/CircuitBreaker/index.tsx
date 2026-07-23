import React, { } from 'react';
import { Tabs } from 'components/Fluent';
import { TabPanel } from 'components/Fluent';
import CircuitBreakerTable from './CircuitBreakerTable';
import FaultDetectTable from './FaultDetectTable';

export default React.memo(() => {

    return (
        <>
            <Tabs>
                <TabPanel label="故障熔断" value="1">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <CircuitBreakerTable />
                    </div>
                </TabPanel>
                <TabPanel label="主动探测" value="2">
                    <div
                        style={{
                            margin: 20
                        }}
                    >
                        <FaultDetectTable />
                    </div>
                </TabPanel>
            </Tabs>
        </>
    )
});
