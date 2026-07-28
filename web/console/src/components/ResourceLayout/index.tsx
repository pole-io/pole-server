import React from 'react';
import { createPortal } from 'react-dom';

import style from './index.module.less';

interface ResourceHeaderProps {
  eyebrow?: React.ReactNode;
  title: React.ReactNode;
  description?: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
  density?: 'default' | 'compact';
  placement?: 'content' | 'app-header';
}

interface ResourceToolbarProps {
  title: React.ReactNode;
  count?: React.ReactNode;
  description?: React.ReactNode;
  filters?: React.ReactNode;
  className?: string;
  density?: 'default' | 'compact';
}

export const ResourceHeader: React.FC<ResourceHeaderProps> = ({
  eyebrow,
  title,
  description,
  actions,
  className,
  density = 'default',
  placement = 'content',
}) => {
  const [appHeaderTarget, setAppHeaderTarget] = React.useState<HTMLElement | null>(null);

  React.useEffect(() => {
    if (placement !== 'app-header') {
      setAppHeaderTarget(null);
      return;
    }
    setAppHeaderTarget(document.getElementById('app-header-context'));
  }, [placement]);

  const header = (
    <section
      className={`${style.resourceHeader} ${density === 'compact' ? style.resourceHeaderCompact : ''} ${placement === 'app-header' ? style.resourceHeaderIntegrated : ''} ${className || ''}`}
    >
      <div className={style.headerContent}>
        {eyebrow && <div className={style.eyebrow}>{eyebrow}</div>}
        <h2>{title}</h2>
        {description && <p title={typeof description === 'string' ? description : undefined}>{description}</p>}
      </div>
      {actions && <div className={style.headerActions}>{actions}</div>}
    </section>
  );

  if (placement === 'app-header') {
    return appHeaderTarget ? createPortal(header, appHeaderTarget) : null;
  }
  return header;
};

export const ResourceToolbar: React.FC<ResourceToolbarProps> = ({
  title,
  count,
  description,
  filters,
  className,
  density = 'default',
}) => (
  <section className={`${style.resourceToolbar} ${density === 'compact' ? style.resourceToolbarCompact : ''} ${className || ''}`}>
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
