import React, { useRef } from 'react';
import { Button, Tooltip } from 'components/Fluent';
import { AddIcon, RefreshIcon } from 'components/Fluent/icons';
import { useSearchParams } from 'components/Router';

import { ResourceHeader } from 'components/ResourceLayout';
import ServicesTable, { ServicesTableHandle } from './services';
import SystemNamespaceServices from './SystemNamespaceServices';
import style from './index.module.less';

export default React.memo(() => {
    const [searchParams] = useSearchParams();
    const serviceTableRef = useRef<ServicesTableHandle>(null);
    const systemNamespace = searchParams.get('scope') === 'system'
        ? searchParams.get('namespace') || 'pole-system'
        : '';

    const refreshServices = () => {
        serviceTableRef.current?.refresh();
    };

    const createService = () => {
        serviceTableRef.current?.create();
    };

    return (
        <div className={style.page}>
            <ResourceHeader
                eyebrow={systemNamespace ? '注册发现 / 系统空间' : '注册发现 / 逻辑服务'}
                title={systemNamespace ? `${systemNamespace} 系统服务` : '服务'}
                description={systemNamespace
                    ? '显式维护当前控制面系统空间中的运行时服务；这些服务不参与业务逻辑服务聚合。'
                    : '逻辑服务用于跨环境聚合；每个环境继续使用自己的 Namespace 与运行时服务名。'}
                actions={(
                    <>
                    {!systemNamespace && (
                    <>
                    <Tooltip content="刷新服务列表">
                        <Button shape="square" variant="outline" onClick={refreshServices}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={createService}>新建逻辑服务</Button>
                    </>
                    )}
                    </>
                )}
            />
            <section className={style.serviceWorkspace}>
                {systemNamespace
                    ? <SystemNamespaceServices namespace={systemNamespace} />
                    : <ServicesTable ref={serviceTableRef} />}
            </section>
        </div>
    )
});
