import { lazy } from 'react';
import { IndicatorIcon } from 'components/Fluent/icons';
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
        path: 'system',
        Component: lazy(() => import('pages/Metrics/SystemMonitor')),
        meta: {
          title: 'menu.metrics.system',
        },
      },
      {
        path: 'service',
        Component: lazy(() => import('pages/Metrics/ServiceMonitor')),
        meta: {
          title: 'menu.metrics.service',
        },
      },
      {
        path: 'service/detail',
        Component: lazy(() => import('pages/Metrics/ServiceMonitor/Detail')),
        meta: {
          hidden: true,
          title: 'menu.metrics.service',
        },
      },
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
    ],
  },
];

export default metrics;
