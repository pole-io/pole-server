import React, { memo } from 'react';
import { useLocation } from 'components/Router';
import { Layout, Button, Space } from 'components/Fluent';
import { ViewListIcon } from 'components/Fluent/icons';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { selectGlobal, toggleMenu } from 'modules/global';
import HeaderIcon from './HeaderIcon';
import { HeaderMenu } from '../Menu';
import WorkspaceModeSwitch from './WorkspaceModeSwitch';
import Style from './index.module.less';

const { Header } = Layout;

export default memo((props: { showMenu?: boolean }) => {
  const globalState = useAppSelector(selectGlobal);
  const dispatch = useAppDispatch();
  const location = useLocation();
  const agentMode = location.pathname === '/agent' || location.pathname.startsWith('/agent/');

  if (!globalState.showHeader) {
    return null;
  }

  let HeaderLeft;
  if (props.showMenu) {
    HeaderLeft = (
      <div className={Style.headerLeft}>
        <HeaderMenu />
        <WorkspaceModeSwitch />
      </div>
    );
  } else {
    HeaderLeft = (
      <Space align='center' className={Style.headerLeft}>
        {!agentMode && (
          <Button
            aria-label={globalState.collapsed ? '展开应用导航' : '收起应用导航'}
            shape='square'
            size='large'
            variant='text'
            onClick={() => dispatch(toggleMenu(null))}
            icon={<ViewListIcon />}
          />
        )}
      </Space>
    );
  }

  return (
    <Header className={Style.panel}>
      {HeaderLeft}
      <div id="app-header-context" className={Style.headerContext} />
      <HeaderIcon />
    </Header>
  );
});
