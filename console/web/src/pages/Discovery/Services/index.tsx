import React, { useRef } from 'react';
import { Button, Space, Tooltip } from 'tdesign-react';
import { AddIcon, RefreshIcon } from 'tdesign-icons-react';

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
            <section className={style.header}>
                <div>
                    <div className={style.eyebrow}>注册发现 / 服务实例</div>
                    <h2>注册发现</h2>
                    <p>查看服务清单，确认命名空间、可见范围以及实例健康状态。</p>
                </div>
                <Space>
                    <Tooltip content="刷新服务列表">
                        <Button shape="square" variant="outline" onClick={refreshServices}>
                            <RefreshIcon />
                        </Button>
                    </Tooltip>
                    <Button theme="primary" icon={<AddIcon />} onClick={createService}>新建服务</Button>
                </Space>
            </section>
            <section className={style.serviceWorkspace}>
                <ServicesTable ref={serviceTableRef} />
            </section>
        </div>
    )
});
