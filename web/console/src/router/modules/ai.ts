import { lazy } from 'react';
import { AlphaIcon } from 'components/Fluent/icons';
import { IRouter } from '../index';

const ai: IRouter[] = [
  {
    path: '/ai',
    meta: {
      title: 'menu.ai',
      Icon: AlphaIcon,
    },
    children: [
      {
        path: 'agent',
        redirect: '/agent',
        meta: {
          hidden: true,
        },
      },
      {
        path: 'a2a',
        Component: lazy(() => import('pages/AI/A2A')),
        meta: {
          title: 'menu.ai.a2a',
        },
      },
      {
        path: 'a2a/detail',
        Component: lazy(() => import('pages/AI/A2A/Detail')),
        meta: {
          title: 'menu.ai.a2a',
          hidden: true,
        },
      },
      {
        path: 'mcps',
        Component: lazy(() => import('pages/AI/Mcp')),
        meta: {
          title: 'menu.ai.mcp',
        },
      },
      {
        path: 'mcps/detail',
        Component: lazy(() => import('pages/AI/Mcp/Detail')),
        meta: {
          title: 'menu.ai.mcp',
          hidden: true,
        },
      },
    ],
  },
];

export default ai;
