import React from 'react';
import { useLocation } from 'components/Router';
import { Layout } from 'components/Fluent';
import { ELayout } from 'modules/global';
import Header from './Header';
import Footer from './Footer';
import Menu from './Menu';
import classnames from 'classnames';
import Content from './AppRouter';

import Style from './AppLayout.module.less';

const SideLayout = React.memo(() => {
  const location = useLocation();
  const agentMode = location.pathname === '/agent' || location.pathname.startsWith('/agent/');
  return (
    <Layout className={classnames(Style.sidePanel, 'narrow-scrollbar')}>
      <Menu showLogo showOperation />
      <Layout className={classnames(Style.sideContainer, agentMode && Style.agentContainer)}>
        {!agentMode && <Header />}
        <Content />
        <Footer />
      </Layout>
    </Layout>
  );
});

const TopLayout = React.memo(() => (
  <Layout className={Style.topPanel}>
    <Header showMenu />
    <Content />
    <Footer />
  </Layout>
));

const MixLayout = React.memo(() => (
    <Layout className={Style.mixPanel}>
      <Header />
      <Layout className={Style.mixMain}>
        <Menu />
        <Layout className={Style.mixContent}>
          <Content />
          <Footer />
        </Layout>
      </Layout>
    </Layout>
));

const FullPageLayout = React.memo(() => <Content />);

export default {
  [ELayout.side]: SideLayout,
  [ELayout.top]: TopLayout,
  [ELayout.mix]: MixLayout,
  [ELayout.fullPage]: FullPageLayout,
};
