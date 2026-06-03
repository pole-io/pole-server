import { lazy } from 'react';
import { CheckCircleIcon, IndicatorIcon } from 'tdesign-icons-react';
import { IRouter } from '../index';

const metrics: IRouter[] = [
  {
    path: '/metrics',
    meta: {
      title: 'menu.metrics',
      Icon: IndicatorIcon,
    },
    children: [
      {
        path: 'event',
        Component: lazy(() => import('pages/Metrics/ServerEvent')),
        meta: {
          title: 'menu.metrics.event',
        },
      },
      {
        path: 'operation',
        Component: lazy(() => import('pages/Metrics/ServerOperation')),
        meta: {
          title: 'menu.metrics.operation',
        },
      },
      // {
      //   path: 'control',
      //   Component: lazy(() => import('pages/Metrics/NetworkError')),
      //   meta: {
      //     title: '监控指标',
      //   },
      // },
      // {
      //   path: 'microservice',
      //   Component: lazy(() => import('pages/Metrics/NetworkError')),
      //   meta: {
      //     title: '服务监控',
      //   },
      // }
    ],
  },
];

export default metrics;
