import React, { memo } from 'react';
import Style from './Menu.module.less';
import FullLogo from 'assets/svg/assets-logo-full.svg?component';
import MiniLogo from 'assets/svg/assets-t-logo.svg?component';
import { useNavigate } from 'react-router-dom';

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
