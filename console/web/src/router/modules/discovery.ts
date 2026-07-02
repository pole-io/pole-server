import { lazy } from 'react';
import { ViewModuleIcon } from 'tdesign-icons-react';
import { IRouter } from '../index';

const discovery: IRouter[] = [
  {
    path: '/discovery',
    group: 'menu.discovery',
    meta: {
      title: 'menu.discovery',
      Icon: ViewModuleIcon,
    },
    children: [
      {
        path: 'service',
        Component: lazy(() => import('pages/Discovery/Services')),
        meta: {
          title: 'menu.discovery.service',
        },
      },
      // Gateway 页面尚未实现，避免暴露 501 占位入口。
      // {
      //   path: 'gateway',
      //   Component: lazy(() => import('pages/Discovery/Gateway')),
      //   meta: {
      //     title: 'menu.discovery.gateway',
      //   },
      // },
      // {
      //   path: 'envoy',
      //   Component: lazy(() => import('pages/Discovery/Envoy')),
      //   meta: {
      //     title: 'menu.discovery.envoy',
      //   },
      // },
      // {
      //   path: 'kubernetes',
      //   Component: lazy(() => import('pages/Discovery/Kubernetes')),
      //   meta: {
      //     title: 'menu.discovery.k8s',
      //   },
      // },
      // 这里的路由是为了在服务列表中点击实例跳转到实例详情页
      {
        path: 'service/instance',
        Component: lazy(() => import('pages/Discovery/Services/Instance')),
        isFullPage: false,
        meta: {
          hidden: true,
        },
      },
    ],
  },
];

export default discovery;
