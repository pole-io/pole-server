import React from 'react';
import { Breadcrumb } from 'components/Fluent';

import styles from './GroupWorkspaceNav.module.less';

const { BreadcrumbItem } = Breadcrumb;

interface GroupWorkspaceNavProps {
  group: string;
  onBack: () => void;
}

const GroupWorkspaceNav: React.FC<GroupWorkspaceNavProps> = ({
  group,
  onBack,
}) => {
  return (
    <section className={styles.context}>
      <Breadcrumb maxItemWidth="220px" className={styles.breadcrumb}>
        <BreadcrumbItem onClick={onBack}>配置中心</BreadcrumbItem>
        <BreadcrumbItem>配置分组</BreadcrumbItem>
        <BreadcrumbItem>{group || '-'}</BreadcrumbItem>
      </Breadcrumb>
    </section>
  );
};

export default GroupWorkspaceNav;
