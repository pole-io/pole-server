import { BrandVariants, createDarkTheme, createLightTheme, Theme } from '@fluentui/react-components';

const latticeBrand: BrandVariants = {
  10: '#020305',
  20: '#071426',
  30: '#092647',
  40: '#0a3765',
  50: '#0a4883',
  60: '#0959a2',
  70: '#086ac1',
  80: '#0f6cbd',
  90: '#2886d8',
  100: '#479ef5',
  110: '#62abf5',
  120: '#77b7f7',
  130: '#96c6fa',
  140: '#b4d6fa',
  150: '#cfe4fa',
  160: '#ebf3fc',
};

export const latticeLightTheme: Theme = {
  ...createLightTheme(latticeBrand),
  fontFamilyBase: '"Segoe UI Variable", "Segoe UI", -apple-system, BlinkMacSystemFont, sans-serif',
  borderRadiusMedium: '4px',
  borderRadiusLarge: '6px',
  colorNeutralBackground1: '#ffffff',
  colorNeutralBackground2: '#f7f7f8',
  colorNeutralBackground3: '#f1f2f4',
  colorNeutralForeground1: '#242424',
  colorNeutralForeground2: '#616161',
  colorNeutralStroke1: '#d1d1d1',
  colorNeutralStroke2: '#e0e0e0',
};

export const latticeDarkTheme: Theme = {
  ...createDarkTheme(latticeBrand),
  fontFamilyBase: '"Segoe UI Variable", "Segoe UI", -apple-system, BlinkMacSystemFont, sans-serif',
  borderRadiusMedium: '4px',
  borderRadiusLarge: '6px',
};
