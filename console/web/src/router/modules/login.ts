import { lazy } from 'react';
import { LogoutIcon } from 'components/Fluent/icons';
import { IRouter } from '../index';

const result: IRouter[] = [
  {
    path: '/login',
    meta: {
      title: 'menu.login',
      Icon: LogoutIcon,
    },
    children: [
      {
        path: 'index',
        Component: lazy(() => import('pages/Login')),
        isFullPage: true,
        meta: {
          title: 'menu.login.center',
        },
      },
    ],
  },
];

export default result;
