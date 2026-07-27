import React, { useRef } from 'react';
import { Button, Tooltip } from 'components/Fluent';
import { AddIcon, RefreshIcon } from 'components/Fluent/icons';

import { ResourceHeader } from 'components/ResourceLayout';
import ServicesTable, { ServicesTableHandle } from './services';
import style from './index.module.less';

export default React.memo(() => {
    const serviceTableRef = useRef<ServicesTableHandle>(null);

    const refreshServices = () => {
        serviceTableRef.current?.refresh();
    };

    const createService = () => {
        serviceTableRef.current?.create();
    };

    return (
        <div className={style.page}>
            <ResourceHeader
                eyebrow="注册发现 / 逻辑服务"
                title="服务"
                description="逻辑服务用于跨环境聚合；每个环境继续使用自己的 Namespace 与运行时服务名。"
                actions={(
                    <>
                    <Tooltip content="刷新服务列表">
                        <Button shape="square" variant="outline" onClick={refreshServices}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={createService}>新建逻辑服务</Button>
                    </>
                )}
            />
            <section className={style.serviceWorkspace}>
                <ServicesTable ref={serviceTableRef} />
            </section>
        </div>
    )
});
