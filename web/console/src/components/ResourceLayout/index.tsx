import React from 'react';

import style from './index.module.less';

interface ResourceHeaderProps {
  eyebrow?: React.ReactNode;
  title: React.ReactNode;
  description?: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
}

interface ResourceToolbarProps {
  title: React.ReactNode;
  count?: React.ReactNode;
  description?: React.ReactNode;
  filters?: React.ReactNode;
  className?: string;
}

export const ResourceHeader: React.FC<ResourceHeaderProps> = ({
  eyebrow,
  title,
  description,
  actions,
  className,
}) => (
  <section className={`${style.resourceHeader} ${className || ''}`}>
    <div className={style.headerContent}>
      {eyebrow && <div className={style.eyebrow}>{eyebrow}</div>}
      <h2>{title}</h2>
      {description && <p>{description}</p>}
    </div>
    {actions && <div className={style.headerActions}>{actions}</div>}
  </section>
);

export const ResourceToolbar: React.FC<ResourceToolbarProps> = ({
  title,
  count,
  description,
  filters,
  className,
}) => (
  <section className={`${style.resourceToolbar} ${className || ''}`}>
    <div className={style.toolbarIntro}>
      <div className={style.toolbarTitle}>
        <strong>{title}</strong>
        {count && <span>{count}</span>}
      </div>
      {description && <p>{description}</p>}
    </div>
    {filters && <div className={style.toolbarFilters}>{filters}</div>}
  </section>
);
