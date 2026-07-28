import React from 'react';
import { Breadcrumb, Empty } from 'components/Fluent';
import { useNavigate } from 'components/Router';

import PolicyDetailView from './PolicyDetailView';
import style from './index.module.less';

const { BreadcrumbItem } = Breadcrumb;

const PolicyDetailPage: React.FC = () => {
  const navigate = useNavigate();
  const searchParams = new URLSearchParams(window.location.search);
  const policyId = searchParams.get('id') || '';
  const policyName = searchParams.get('name') || '策略详情';

  return (
    <div className={style.policyStandalonePage}>
      <Breadcrumb className={style.policyStandaloneBreadcrumb} maxItemWidth="240px">
        <BreadcrumbItem onClick={() => navigate('/auth/policies')}>权限策略</BreadcrumbItem>
        <BreadcrumbItem>{policyName}</BreadcrumbItem>
      </Breadcrumb>
      {policyId ? (
        <section className={style.policyStandaloneSurface}>
          <PolicyDetailView policyId={policyId} />
        </section>
      ) : (
        <section className={style.policyStandaloneEmpty}>
          <Empty description="缺少策略标识，无法加载策略详情" />
        </section>
      )}
    </div>
  );
};

export default React.memo(PolicyDetailPage);
