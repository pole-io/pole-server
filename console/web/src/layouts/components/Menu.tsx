import React, { memo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  Button,
  Nav,
  NavCategory,
  NavCategoryItem,
  NavItem,
  NavSubItem,
  NavSubItemGroup,
} from '@fluentui/react-components';
import { IRouter } from 'router';
import router from 'router';
import { resolve } from 'utils/path';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import MenuLogo from './MenuLogo';
import WorkspaceModeSwitch from './Header/WorkspaceModeSwitch';
import Style from './Menu.module.less';
import { useTranslation } from 'react-i18next';

const getSelectedMenuValue = (pathname: string) => {
  if (pathname === '/governance' || pathname.startsWith('/governance/')) return '/governance/workbench';
  return pathname;
};

const visibleRoutes = (routes: IRouter[], isAdmin: boolean) => routes.filter(
  (item) => item.meta && !item.meta.hidden && (!item.meta.adminOnly || isAdmin),
);

interface IMenuProps {
  showLogo?: boolean;
  showOperation?: boolean;
}

interface NavNodesProps {
  routes: IRouter[];
  parentPath?: string;
  compact?: boolean;
  onNavigate: (path: string) => void;
  t: (key: string) => string;
  isAdmin: boolean;
}

const NavNodes: React.FC<NavNodesProps> = ({ routes, parentPath = '', compact, onNavigate, t, isAdmin }) => (
  <>
    {visibleRoutes(routes, isAdmin).map((item) => {
      const { children, meta, path } = item;
      const routePath = resolve(parentPath, path);
      const label = t(meta?.title || '');
      const ItemIcon = meta?.Icon;

      if (!children?.length) {
        return (
          <NavItem key={routePath} value={routePath} icon={ItemIcon ? <ItemIcon /> : undefined} onClick={() => onNavigate(routePath)}>
            <span className={Style.navLabel}>{label}</span>
          </NavItem>
        );
      }

      if (meta?.single) {
        const firstChild = visibleRoutes(children, isAdmin)[0];
        if (!firstChild) return null;
        const childPath = resolve(routePath, firstChild.path);
        return (
          <NavItem key={childPath} value={childPath} icon={ItemIcon ? <ItemIcon /> : undefined} onClick={() => onNavigate(childPath)}>
            <span className={Style.navLabel}>{label}</span>
          </NavItem>
        );
      }

      return (
        <NavCategory key={routePath} value={routePath}>
          <NavCategoryItem icon={ItemIcon ? <ItemIcon /> : undefined}>
            <span className={Style.navLabel}>{label}</span>
          </NavCategoryItem>
          {!compact && (
            <NavSubItemGroup>
              {visibleRoutes(children, isAdmin).map((child) => {
                const childPath = resolve(routePath, child.path);
                return (
                  <NavSubItem key={childPath} value={childPath} onClick={() => onNavigate(childPath)}>
                    {t(child.meta?.title || '')}
                  </NavSubItem>
                );
              })}
            </NavSubItemGroup>
          )}
        </NavCategory>
      );
    })}
  </>
);

export const HeaderMenu = memo(() => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const role = useAppSelector((state) => state.userLogin.currentUser.role);
  const isAdmin = role === 'main' || role === 'admin';
  return (
    <nav className={Style.headerNav} aria-label="主导航">
      {visibleRoutes(router, isAdmin).map((item) => {
        const path = item.meta?.single && item.children?.length ? resolve(item.path, item.children[0].path) : item.path;
        return (
          <Button
            key={path}
            appearance={location.pathname.startsWith(item.path) ? 'primary' : 'subtle'}
            icon={item.meta?.Icon ? <item.meta.Icon /> : undefined}
            onClick={() => navigate(path)}
          >
            {t(item.meta?.title || '')}
          </Button>
        );
      })}
    </nav>
  );
});

export default memo((props: IMenuProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { version, collapsed } = useAppSelector(selectGlobal);
  const role = useAppSelector((state) => state.userLogin.currentUser.role);
  const isAdmin = role === 'main' || role === 'admin';
  const categoryValues = React.useMemo(
    () => visibleRoutes(router, isAdmin)
      .filter((item) => item.children && item.children.length > 0 && !item.meta?.single)
      .map((item) => resolve('', item.path)),
    [isAdmin],
  );
  const selectedValue = getSelectedMenuValue(location.pathname);
  const isAgent = location.pathname === '/agent' || location.pathname.startsWith('/agent/');

  return (
    <aside className={`${Style.fluentSidebar} ${isAgent ? Style.agentSidebar : ''} ${collapsed ? Style.collapsed : ''}`} aria-label="应用导航">
      {props.showLogo && <MenuLogo collapsed={collapsed} />}
      {!isAgent ? (
        <Nav
          className={Style.fluentNav}
          selectedValue={selectedValue}
          openCategories={collapsed ? [] : categoryValues}
          multiple
          density="medium"
          onNavItemSelect={(_, data) => navigate(String(data.value))}
        >
          <NavNodes routes={router} compact={collapsed} onNavigate={navigate} t={t} isAdmin={isAdmin} />
        </Nav>
      ) : (
        <div
          id="agent-session-sidebar-host"
          className={Style.agentSessionHost}
          data-collapsed={collapsed ? 'true' : 'false'}
          aria-label="Agent 会话管理"
        />
      )}
      <WorkspaceModeSwitch placement="sidebar" collapsed={collapsed} />
      <footer className={Style.navFooter}>{collapsed ? version : `Lattice.Hub ${version}`}</footer>
    </aside>
  );
});
