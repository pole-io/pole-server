import React, { memo } from 'react';

import ErrorPage, { ECode } from 'components/ErrorPage';
import style from './index.module.less';

const GatewayPage: React.FC = () => (
  <div className={style.page}>
    <section className={style.header}>
      <div>
        <div className={style.eyebrow}>Service Registry / Gateway</div>
        <h2>网关</h2>
        <p>独立查看注册发现下的网关入口，后续网关实例和服务入口能力在这里维护。</p>
      </div>
    </section>
    <section className={style.placeholder}>
      <ErrorPage code={ECode.unimplemented} />
    </section>
  </div>
);

export default memo(GatewayPage);
