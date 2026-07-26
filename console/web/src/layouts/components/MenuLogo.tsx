import React, { memo } from 'react';
import Style from './Menu.module.less';
import FullLogo from 'assets/svg/assets-logo-full.svg?react';
import MiniLogo from 'assets/svg/assets-t-logo.svg?react';
import { useNavigate } from 'components/Router';

interface IProps {
  collapsed?: boolean;
}

export default memo((props: IProps) => {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate('/');
  };

  return (
    <div
      className={`${Style.menuLogo} ${props.collapsed ? Style.menuLogoCollapsed : Style.menuLogoExpanded}`}
      onClick={handleClick}
    >
      {props.collapsed ? <MiniLogo className={Style.menuMiniLogo} /> : <FullLogo className={Style.logoFull} />}
    </div>
  );
});
