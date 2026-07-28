import { lazy } from 'react';
import { SettingIcon } from 'components/Fluent/icons';
import { IRouter } from '../index';

const systemConfiguration: IRouter[] = [
  {
    path: '/system-configuration',
    Component: lazy(() => import('pages/SystemConfiguration')),
    meta: {
      title: 'menu.systemConfiguration',
      Icon: SettingIcon,
      single: true,
      adminOnly: true,
    },
  },
];

export default systemConfiguration;
