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
        path: 'lossless',
        Component: lazy(() => import('pages/Governance/LossLess')),
        meta: { title: 'menu.governance.lossless' },
      },
      {
        path: 'router',
        Component: lazy(() => import('pages/Governance/Router')),
        meta: { title: 'menu.governance.router' },
      },
      {
        path: 'ratelimit',
        Component: lazy(() => import('pages/Governance/RateLimit')),
        meta: {
          title: 'menu.governance.ratelimit',
        },
      },
      {
        path: 'circuitbreaker',
        Component: lazy(() => import('pages/Governance/CircuitBreaker')),
        meta: { title: 'menu.governance.circuitbreaker' },
      },
      // {
      //   path: 'security',
      //   Component: lazy(() => import('pages/Governance/Security')),
      //   meta: { title: 'menu.governance.security' },
      // }
    ],
  },
];

export default result;
