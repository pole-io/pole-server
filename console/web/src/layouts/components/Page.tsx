import React, { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../../modules/store';
import { selectGlobal, switchFullPage } from '../../modules/global';
import { Layout, Breadcrumb } from 'components/Fluent';
import Style from './Page.module.less';

const { Content } = Layout;
const { BreadcrumbItem } = Breadcrumb;

const Page = ({
  children,
  isFullPage,
  breadcrumbs,
}: React.PropsWithChildren<{ isFullPage?: boolean; breadcrumbs?: string[] }>) => {
  const globalState = useAppSelector(selectGlobal);
  const dispatch = useAppDispatch();
  const location = useLocation();
  const agentMode = location.pathname === '/agent' || location.pathname.startsWith('/agent/');
  useEffect(() => {
    dispatch(switchFullPage(isFullPage));
  }, [isFullPage]);

  if (isFullPage) {
    return <>{children}</>;
  }

  if (agentMode) {
    return <Content className={Style.agentPanel}>{children}</Content>;
  }

  return (
    <Content className={Style.panel}>
      {globalState.showBreadcrumbs && (
        <Breadcrumb className={Style.breadcrumb}>
          {breadcrumbs?.map((item, index) => (
            <BreadcrumbItem key={index}>{item}</BreadcrumbItem>
          ))}
        </Breadcrumb>
      )}
      {children}
    </Content>
  );
};

export default React.memo(Page);
