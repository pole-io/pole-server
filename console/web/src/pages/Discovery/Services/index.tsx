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
                eyebrow="注册发现 / 服务实例"
                title="注册发现"
                description="服务名是全局逻辑标识；同名服务在不同命名空间中表示该服务的不同环境实例。"
                actions={(
                    <>
                    <Tooltip content="刷新服务列表">
                        <Button shape="square" variant="outline" onClick={refreshServices}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={createService}>新建服务</Button>
                    </>
                )}
            />
            <section className={style.serviceWorkspace}>
                <ServicesTable ref={serviceTableRef} />
            </section>
        </div>
    )
});
