import { BotIcon } from 'components/Fluent/icons';
import { AgentPage } from '../agentLoader';
import { IRouter } from '../index';

const agent: IRouter[] = [
  {
    path: '/agent',
    meta: {
      title: 'menu.agent',
      Icon: BotIcon,
      single: true,
      hidden: true,
    },
    children: [
      {
        path: '',
        Component: AgentPage,
        meta: {
          title: 'menu.agent',
        },
      },
    ],
  },
];

export default agent;
