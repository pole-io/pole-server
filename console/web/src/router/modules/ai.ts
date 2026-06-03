import { lazy } from 'react';
import { AlphaIcon } from 'tdesign-icons-react';
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
        path: 'a2a',
        Component: lazy(() => import('pages/AI/A2A')),
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
    ],
  },
];

export default ai;
