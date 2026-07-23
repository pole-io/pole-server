import { lazy } from 'react';
import { QueueIcon } from 'components/Fluent/icons';
import { IRouter } from '../index';

const configuration: IRouter[] = [
  {
    path: '/configuration',
    meta: {
      title: 'menu.configuration',
      Icon: QueueIcon,
    },
    children: [
      {
        path: 'group',
        Component: lazy(() => import('pages/Configuration/Group')),
        meta: {
          title: 'menu.configuration.group',
        },
      },
      // 这里的路由是为了在服务列表中点击实例跳转到实例详情页
      {
        path: 'group/files',
        Component: lazy(() => import('pages/Configuration/Group/Files')),
        isFullPage: false,
        meta: {
          hidden: true,
        },
      },
    ],
  },
];

export default configuration;
