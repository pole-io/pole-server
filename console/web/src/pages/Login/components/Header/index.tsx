import { Button } from 'components/Fluent';
import { HelpCircleIcon, SettingIcon } from 'components/Fluent/icons';
import { useAppDispatch } from 'modules/store';
import { toggleSetting } from 'modules/global';

import LogoFullIcon from 'assets/svg/assets-logo-full.svg?react';
import Style from './index.module.less';

export default function Header() {
  const dispatch = useAppDispatch();

  const navToHelper = () => {
    window.open('https://github.com/pole-io/pole-control-plane');
  };

  const toggleSettingPanel = () => {
    dispatch(toggleSetting());
  };

  return (
    <div>
      <header className={Style.loginHeader}>
        <LogoFullIcon />
        <div className={Style.operationsContainer}>
          <Button
            className={Style.operationsButton}
            theme='default'
            shape='square'
            variant='text'
            aria-label='打开项目帮助'
            onClick={navToHelper}
          >
            <HelpCircleIcon className={Style.icon} />
          </Button>
          <Button
            className={Style.operationsButton}
            theme='default'
            shape='square'
            variant='text'
            aria-label='打开外观设置'
            onClick={toggleSettingPanel}
          >
            <SettingIcon className={Style.icon} />
          </Button>
        </div>
      </header>
    </div>
  );
}
