import React from 'react';
import { Breadcrumb } from 'components/Fluent';
import { useNavigate } from 'components/Router';

import styles from './GroupWorkspaceNav.module.less';

const { BreadcrumbItem } = Breadcrumb;

interface GroupWorkspaceNavProps {
  active: 'files' | 'templates';
  group: string;
  filesHref: string;
  templatesHref: string;
  onBack: () => void;
}

const GroupWorkspaceNav: React.FC<GroupWorkspaceNavProps> = ({
  active,
  group,
  filesHref,
  templatesHref,
  onBack,
}) => {
  const navigate = useNavigate();
  const follow = (event: React.MouseEvent<HTMLAnchorElement>, href: string) => {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    navigate(href);
  };

  return (
    <section className={styles.context}>
      <Breadcrumb maxItemWidth="220px" className={styles.breadcrumb}>
        <BreadcrumbItem onClick={onBack}>配置中心</BreadcrumbItem>
        <BreadcrumbItem>配置分组</BreadcrumbItem>
        <BreadcrumbItem>{group || '-'}</BreadcrumbItem>
      </Breadcrumb>
      <div className={styles.workspaceBar}>
        <div className={styles.identity}>
          <span>配置分组工作区</span>
          <strong>{group || '-'}</strong>
        </div>
        <nav className={styles.links} aria-label="配置分组资源">
          <a
            href={filesHref}
            aria-current={active === 'files' ? 'page' : undefined}
            className={active === 'files' ? styles.activeLink : undefined}
            onClick={event => follow(event, filesHref)}
          >
            <span>配置文件</span>
            <small>环境 → 文件</small>
          </a>
          <a
            href={templatesHref}
            aria-current={active === 'templates' ? 'page' : undefined}
            className={active === 'templates' ? styles.activeLink : undefined}
            onClick={event => follow(event, templatesHref)}
          >
            <span>配置模板</span>
            <small>模板 → 环境 Value</small>
          </a>
        </nav>
      </div>
    </section>
  );
};

export default GroupWorkspaceNav;
