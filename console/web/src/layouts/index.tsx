import React, { memo, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { Drawer, Layout } from 'components/Fluent';
import throttle from 'lodash/throttle';
import { useAppSelector, useAppDispatch } from 'modules/store';
import { selectGlobal, toggleSetting, toggleMenu, ELayout, switchTheme, openSystemTheme } from 'modules/global';
import Setting from './components/Setting';
import AppLayout from './components/AppLayout';
import Style from './index.module.less';
import { hydrateSession } from 'modules/user/login';

export default memo(() => {
  const globalState = useAppSelector(selectGlobal);
  const dispatch = useAppDispatch();
  const location = useLocation();
  const { isLogin, sessionResolved } = useAppSelector((state) => state.userLogin);

  const AppContainer = AppLayout[globalState.isFullPage ? ELayout.fullPage : globalState.layout];
  const workspaceMode = location.pathname === '/agent' || location.pathname.startsWith('/agent/') ? 'agent' : 'console';

  useEffect(() => {
    if (isLogin && !sessionResolved) {
      dispatch(hydrateSession());
    }
  }, [dispatch, isLogin, sessionResolved]);

  useEffect(() => {
    if (globalState.systemTheme) {
      dispatch(openSystemTheme());
    } else {
      dispatch(switchTheme(globalState.theme));
    }
  }, []);

  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme:dark)');
    const handleSystemThemeChange = () => dispatch(openSystemTheme());

    if (globalState.systemTheme) {
      media.addEventListener('change', handleSystemThemeChange);
    }

    return () => media.removeEventListener('change', handleSystemThemeChange);
  }, [dispatch, globalState.systemTheme]);

  useEffect(() => {
    const handleResize = throttle(() => {
      if (window.innerWidth < 900) {
        dispatch(toggleMenu(true));
      } else if (window.innerWidth > 1000) {
        dispatch(toggleMenu(false));
      }
    }, 100);
    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [dispatch]);

  return (
    <Layout className={Style.panel}>
      <AppContainer key={workspaceMode} />
      <Drawer
        destroyOnClose
        visible={globalState.setting}
        size='medium'
        footer={false}
        header='页面配置'
        onClose={() => dispatch(toggleSetting())}
      >
        <Setting />
      </Drawer>
    </Layout>
  );
});
