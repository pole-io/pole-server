import React from 'react';
import { useLocation } from 'react-router-dom';

import { Button } from 'components/Fluent';
import { ArrowRightIcon, BotIcon, System2Icon } from 'components/Fluent/icons';
import Style from './WorkspaceModeSwitch.module.less';

const LAST_CONSOLE_LOCATION = 'pole:last-console-location';

const isSafeConsoleLocation = (value?: string | null) => Boolean(
  value
  && value.startsWith('/')
  && !value.startsWith('//')
  && value !== '/agent'
  && !value.startsWith('/agent?')
  && !value.startsWith('/agent/'),
);

interface WorkspaceModeSwitchProps {
  placement?: 'header' | 'sidebar';
  collapsed?: boolean;
}

export default function WorkspaceModeSwitch({ placement = 'header', collapsed = false }: WorkspaceModeSwitchProps) {
  const location = useLocation();
  const isAgent = location.pathname === '/agent' || location.pathname.startsWith('/agent/');

  React.useEffect(() => {
    if (!isAgent) {
      sessionStorage.setItem(LAST_CONSOLE_LOCATION, `${location.pathname}${location.search}${location.hash}`);
    }
  }, [isAgent, location.hash, location.pathname, location.search]);

  const openConsole = () => {
    if (!isAgent) return;
    const returnTo = new URLSearchParams(location.search).get('returnTo');
    const remembered = sessionStorage.getItem(LAST_CONSOLE_LOCATION);
    const target = (isSafeConsoleLocation(returnTo) && returnTo)
      || (isSafeConsoleLocation(remembered) && remembered)
      || '/namespace';
    window.location.assign(target);
  };

  const openAgent = () => {
    if (isAgent) return;
    const returnTo = `${location.pathname}${location.search}${location.hash}`;
    sessionStorage.setItem(LAST_CONSOLE_LOCATION, returnTo);
    window.location.assign(`/agent?returnTo=${encodeURIComponent(returnTo)}`);
  };

  if (placement === 'sidebar') {
    const label = isAgent ? '返回普通控制台' : '进入 Agent 工作台';
    const description = isAgent ? '回到上次访问页面' : '对话查询与变更资源';
    const ModeIcon = isAgent ? System2Icon : BotIcon;
    return (
      <div className={`${Style.sidebarSwitch} ${collapsed ? Style.sidebarCollapsed : ''}`}>
        <Button variant="text" className={Style.sidebarButton} aria-label={label} title={collapsed ? label : undefined} onClick={isAgent ? openConsole : openAgent}>
          <span className={Style.sidebarIcon}><ModeIcon /></span>
          {!collapsed && (
            <span className={Style.sidebarCopy}>
              <strong>{label}</strong>
              <small>{description}</small>
            </span>
          )}
          {!collapsed && <ArrowRightIcon className={Style.sidebarArrow} />}
        </Button>
      </div>
    );
  }

  return (
    <div className={Style.headerSwitch} role="group" aria-label="工作模式">
      <Button variant="text" className={!isAgent ? Style.modeActive : ''} aria-pressed={!isAgent} onClick={openConsole}>
        <System2Icon />
        <span>普通控制台</span>
      </Button>
      <Button
        variant="text"
        className={isAgent ? Style.modeActive : ''}
        aria-pressed={isAgent}
        onClick={openAgent}
      >
        <BotIcon />
        <span>Agent</span>
      </Button>
    </div>
  );
}
