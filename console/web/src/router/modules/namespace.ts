import { lazy } from 'react';
import { ComponentSpaceIcon } from 'components/Fluent/icons';
import { IRouter } from '../index';

const namespace: IRouter[] = [
  {
    path: '/namespace',
    group: 'menu.namespace',
    meta: {
      title: 'menu.namespace',
      Icon: ComponentSpaceIcon,
    },
    Component: lazy(() => import('pages/Namespace'))
  },
];

export default namespace;
