import React, { memo } from 'react';
import { useAppDispatch, useAppSelector } from 'modules/store';
import {
  selectGlobal,
  switchTheme,
  switchColor,
  openSystemTheme,
} from 'modules/global';
import { ETheme } from 'types/index.d';
import RadioColor from './RadioColor';
import RadioRect from './RadioRect';

import Style from './index.module.less';

const SYSTEM_THEME = 'system';

const themeList = [
  {
    value: ETheme.light,
    name: '明亮',
  },
  {
    value: ETheme.dark,
    name: '黑暗',
  },
  {
    value: SYSTEM_THEME,
    name: '跟随系统',
  },
];

export default memo(() => {
  const dispatch = useAppDispatch();
  const globalState = useAppSelector(selectGlobal);

  const handleThemeSwitch = (value: string) => {
    if (value === SYSTEM_THEME) {
      dispatch(openSystemTheme());
    } else {
      dispatch(switchTheme(value as ETheme));
      dispatch(switchColor(globalState.color));
    }
  };

  return (
    <div className={Style.settingContent}>
      <RadioRect
        value={globalState.systemTheme ? SYSTEM_THEME : globalState.theme}
        onChange={handleThemeSwitch}
        options={themeList}
      />

      <RadioColor value={globalState.color} onChange={(value) => dispatch(switchColor(value))} />
    </div>
  );
});
