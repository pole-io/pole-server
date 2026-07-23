import React from 'react';

type AgentPageModule = typeof import('pages/AI/Agent');

let agentPagePromise: Promise<AgentPageModule> | undefined;
let agentPageModule: AgentPageModule | undefined;

export const preloadAgentPage = () => {
  agentPagePromise ??= import('pages/AI/Agent')
    .then((module) => {
      agentPageModule = module;
      return module;
    })
    .catch((error) => {
      agentPagePromise = undefined;
      throw error;
    });
  return agentPagePromise;
};

export const AgentPage = () => {
  const [Page, setPage] = React.useState<AgentPageModule['default'] | undefined>(() => agentPageModule?.default);
  const [loadFailed, setLoadFailed] = React.useState(false);

  React.useEffect(() => {
    if (Page) return undefined;
    let active = true;
    void preloadAgentPage()
      .then((module) => {
        if (active) setPage(() => module.default);
      })
      .catch(() => {
        if (active) setLoadFailed(true);
      });
    return () => {
      active = false;
    };
  }, [Page]);

  if (Page) return React.createElement(Page);
  return React.createElement(
    'div',
    {
      role: loadFailed ? 'alert' : 'status',
      style: { padding: 32, color: 'var(--app-text-secondary)' },
    },
    loadFailed ? 'Agent 加载失败，请返回普通控制台后重试。' : '正在加载 Agent…',
  );
};
