import { lazy } from 'react';
import { LayersIcon } from 'tdesign-icons-react';
import { IRouter } from '../index';

const result: IRouter[] = [
  {
    path: '/governance',
    meta: {
      title: 'menu.governance',
      Icon: LayersIcon,
    },
    children: [
      {
        path: '',
        Component: lazy(() => import('pages/Governance/Workbench')),
      },
      {
        path: 'workbench',
        Component: lazy(() => import('pages/Governance/Workbench')),
        meta: { title: 'menu.governance.workbench' },
      },
      {
        path: 'lossless',
        Component: lazy(() => import('pages/Governance/LossLess')),
        meta: { title: 'menu.governance.lossless', hidden: true },
      },
      {
        path: 'router',
        Component: lazy(() => import('pages/Governance/Router')),
        meta: { title: 'menu.governance.router', hidden: true },
      },
      {
        path: 'ratelimit',
        Component: lazy(() => import('pages/Governance/RateLimit')),
        meta: {
          title: 'menu.governance.ratelimit',
          hidden: true,
        },
      },
      {
        path: 'circuitbreaker',
        Component: lazy(() => import('pages/Governance/CircuitBreaker')),
        meta: { title: 'menu.governance.circuitbreaker', hidden: true },
      },
      {
        path: 'security',
        Component: lazy(() => import('pages/Governance/Security')),
        meta: { title: 'menu.governance.security', hidden: true },
      },
    ],
  },
];

export default result;
