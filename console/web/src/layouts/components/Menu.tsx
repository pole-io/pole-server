import React, { memo, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Menu, MenuValue } from 'tdesign-react';
import router, { IRouter } from 'router';
import { resolve } from 'utils/path';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import MenuLogo from './MenuLogo';
import Style from './Menu.module.less';
import { useTranslation } from 'react-i18next';

const { SubMenu, MenuItem, HeadMenu } = Menu;

const getSelectedMenuValue = (pathname: string) => {
  if (pathname === '/governance' || pathname.startsWith('/governance/')) {
    return '/governance/workbench';
  }
  return pathname;
};

interface IMenuProps {
  showLogo?: boolean;
  showOperation?: boolean;
}

const renderMenuItems = (t: (key: string) => string, menu: IRouter[], parentPath = '') => {
  const navigate = useNavigate();
  return menu.map((item) => {
    const { children, meta, path } = item;

    if (!meta || meta?.hidden === true) {
      // 无meta信息 或 hidden == true，路由不显示为菜单
      return null;
    }

    const { Icon, title, single } = meta;
    const routerPath = resolve(parentPath, path);

    if (!children || children.length === 0) {
      return (
        <MenuItem
          key={routerPath}
          value={routerPath}
          icon={Icon ? <Icon /> : undefined}
          onClick={() => navigate(routerPath)}
        >
          {t(title || '')}
        </MenuItem>
      );
    }

    if (single && children?.length > 0) {
      const firstChild = children[0];
      if (firstChild?.meta && !firstChild?.meta?.hidden) {
        const { Icon, title } = meta;
        const singlePath = resolve(resolve(parentPath, path), firstChild.path);
        return (
          <MenuItem
            key={singlePath}
            value={singlePath}
            icon={Icon ? <Icon /> : undefined}
            onClick={() => navigate(singlePath)}
          >
            {t(title || '')}
          </MenuItem>
        );
      }
    }

    // if (item.group) {
    //   // 如果是分组菜单，直接返回一个分组菜单
    //   return (
    //     <MenuGroup title={item.group}>
    //       {/* <SubMenu key={routerPath} value={routerPath} title={item.group} icon={Icon ? <Icon /> : undefined}>
    //         {renderMenuItems(children, routerPath)}
    //       </SubMenu> */}
    //       {renderMenuItems(children, routerPath)}
    //     </MenuGroup>
    //   );
    // }

    return (
      <SubMenu key={routerPath} value={routerPath} title={t(title || '')} icon={Icon ? <Icon /> : undefined}>
        {renderMenuItems(t, children, routerPath)}
      </SubMenu>
    );
  });
};

/**
 * 顶部菜单
 */
export const HeaderMenu = memo(() => {
  const { t } = useTranslation();
  const globalState = useAppSelector(selectGlobal);
  const location = useLocation();
  const [active, setActive] = useState<MenuValue>(location.pathname); // todo

  return (
    <HeadMenu
      expandType='popup'
      style={{ marginBottom: 20 }}
      value={active}
      theme={globalState.theme}
      onChange={(v) => setActive(v)}
    >
      {renderMenuItems(t, router)}
    </HeadMenu>
  );
});

/**
 * 左侧菜单
 */
export default memo((props: IMenuProps) => {
  const { t } = useTranslation();
  const location = useLocation();
  const globalState = useAppSelector(selectGlobal);

  const { version } = globalState;
  const bottomText = globalState.collapsed ? version : `Pole.IO ${version}`;
  const selectedMenuValue = getSelectedMenuValue(location.pathname);

  return (
    <Menu
      width='232px'
      style={{ flexShrink: 0, height: '100%' }}
      className={Style.menuPanel2}
      value={selectedMenuValue}
      theme={globalState.theme}
      collapsed={globalState.collapsed}
      logo={props.showLogo ? <MenuLogo collapsed={globalState.collapsed} /> : undefined}
      defaultExpanded={['/ai', '/discovery', '/configuration', '/governance', '/metrics', '/auth']}
    >
      {renderMenuItems(t, router)}
    </Menu>
  );
});
