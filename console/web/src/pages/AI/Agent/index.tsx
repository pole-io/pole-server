import React from 'react';
import { createPortal } from 'react-dom';
import { useNavigate, useSearchParams } from 'react-router-dom';

import CodeDiffEditor from 'components/CodeDiffEditor';
import { Button, Input, Switch, Tag, Textarea, Tooltip } from 'components/Fluent';
import {
  AddIcon,
  ArrowRightIcon,
  BotIcon,
  CheckCircleIcon,
  CloseIcon,
  DeleteIcon,
  EditIcon,
  IndicatorIcon,
  LockOnIcon,
  PlugConnectedIcon,
  SearchIcon,
  SendIcon,
  ServiceIcon,
  SettingIcon,
  ToolsCircleIcon,
  ViewListIcon,
} from 'components/Fluent/icons';
import HeaderIcon from 'layouts/components/Header/HeaderIcon';
import {
  AgentRuntimeStatus,
  confirmAgentProposal,
  getAgentRuntime,
  sendAgentTurn,
} from 'services/agent';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { toRequestErrorPayload } from 'utils/request';
import {
  AgentChatMessage,
  AgentLocalSession,
  AgentMemorySettings,
  AgentResourceContext,
  deleteAgentSession,
  getActiveAgentSessionID,
  listAgentSessions,
  saveActiveAgentSessionID,
  saveAgentSession,
} from './sessionStore';
import style from './index.module.less';

const nextID = () => globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;

const createContextStarter = (path: string) => `修改配置 ${path}\n\`\`\`yaml\n\n\`\`\`\n变更说明：`;

const createSession = (resourceContext?: AgentResourceContext): AgentLocalSession => {
  const now = Date.now();
  return {
    id: nextID(),
    title: resourceContext ? resourceContext.name : '新会话',
    createdAt: now,
    updatedAt: now,
    messages: [],
    draft: resourceContext ? createContextStarter(resourceContext.path) : '',
    memory: { enabled: true, maxTurns: 10 },
    resourceContext,
  };
};

const message = (role: AgentChatMessage['role'], content: string, title?: string): AgentChatMessage => ({
  id: nextID(),
  role,
  title,
  content,
  createdAt: Date.now(),
});

const titleFromMessage = (value: string) => {
  const firstLine = value.replace(/```[\s\S]*$/m, '').split('\n')[0].trim() || '新会话';
  return firstLine.length > 28 ? `${firstLine.slice(0, 28)}…` : firstLine;
};

const sessionMeta = (session: AgentLocalSession) => {
  if (session.resourceContext?.path) return session.resourceContext.path;
  const latest = [...session.messages].reverse().find((item) => item.role === 'user');
  return latest?.content.split('\n')[0].slice(0, 36) || '等待输入任务';
};

const sessionTime = (timestamp: number) => {
  const date = new Date(timestamp);
  const now = new Date();
  if (date.toDateString() === now.toDateString()) {
    if (now.getTime() - timestamp < 60_000) return '刚刚';
    return new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false }).format(date);
  }
  return new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric' }).format(date);
};

const sessionGroup = (timestamp: number) => {
  const date = new Date(timestamp);
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  if (date.toDateString() === today.toDateString()) return '今天';
  if (date.toDateString() === yesterday.toDateString()) return '昨天';
  return '更早';
};

export default function AgentWorkbenchPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const resourceContext = React.useMemo<AgentResourceContext | undefined>(() => {
    const kind = searchParams.get('kind');
    const namespace = searchParams.get('namespace');
    const group = searchParams.get('group');
    const name = searchParams.get('name');
    if (kind !== 'config.file' || !namespace || !group || !name) return undefined;
    return { kind, namespace, group, name, path: `${namespace}/${group}/${name}` };
  }, [searchParams]);
  const [sessions, setSessions] = React.useState<AgentLocalSession[]>([]);
  const [activeSessionID, setActiveSessionID] = React.useState('');
  const [hydrated, setHydrated] = React.useState(false);
  const [runtime, setRuntime] = React.useState<AgentRuntimeStatus>({
    ready: false,
    mode: 'unavailable',
    configured: false,
    promptVersion: 'unknown',
    mcpConnected: false,
    tools: [],
    reason: '正在检查 Agent 运行时',
  });
  const [busySessionID, setBusySessionID] = React.useState('');
  const [confirming, setConfirming] = React.useState(false);
  const [contextOpen, setContextOpen] = React.useState(true);
  const [sessionQuery, setSessionQuery] = React.useState('');
  const [editingSessionID, setEditingSessionID] = React.useState('');
  const [editingTitle, setEditingTitle] = React.useState('');
  const [sidebarHost, setSidebarHost] = React.useState<HTMLElement | null>(null);
  const idempotencyKey = React.useRef('');
  const timelineRef = React.useRef<HTMLDivElement>(null);
  const composerRef = React.useRef<HTMLTextAreaElement>(null);
  const storageErrorShown = React.useRef(false);
  const activeSession = sessions.find((session) => session.id === activeSessionID);
  const proposal = activeSession?.proposal;
  const receipt = activeSession?.receipt;
  const busy = busySessionID === activeSessionID;
  const groupedSessions = React.useMemo(() => {
    const query = sessionQuery.trim().toLocaleLowerCase();
    const filtered = sessions.filter((session) => !query
      || session.title.toLocaleLowerCase().includes(query)
      || sessionMeta(session).toLocaleLowerCase().includes(query));
    return ['今天', '昨天', '更早'].map((label) => ({
      label,
      sessions: filtered.filter((session) => sessionGroup(session.updatedAt) === label),
    })).filter((group) => group.sessions.length);
  }, [sessionQuery, sessions]);

  React.useEffect(() => {
    setSidebarHost(document.getElementById('agent-session-sidebar-host'));
  }, []);

  React.useEffect(() => {
    let cancelled = false;
    getAgentRuntime()
      .then((status) => {
        if (!cancelled) setRuntime(status);
      })
      .catch((error) => {
        if (!cancelled) {
          setRuntime({
            ready: false,
            mode: 'unavailable',
            configured: false,
            promptVersion: 'unknown',
            mcpConnected: false,
            tools: [],
            reason: error instanceof Error ? error.message : 'Agent 运行时不可用',
          });
        }
      });
    return () => { cancelled = true; };
  }, []);

  const reportStorageError = React.useCallback((error: unknown) => {
    if (storageErrorShown.current) return;
    storageErrorShown.current = true;
    openErrNotification('本地会话不可用', error instanceof Error ? error.message : '浏览器无法访问 IndexedDB');
  }, []);

  const persistSession = React.useCallback((session: AgentLocalSession) => {
    void saveAgentSession(session).catch(reportStorageError);
  }, [reportStorageError]);

  React.useEffect(() => {
    let cancelled = false;
    Promise.all([listAgentSessions(), getActiveAgentSessionID()])
      .then(([storedSessions, storedActiveID]) => {
        if (cancelled) return;
        let nextSessions = storedSessions;
        let nextActiveID = storedActiveID;
        if (resourceContext) {
          const contextual = storedSessions.find((session) => session.resourceContext?.path === resourceContext.path);
          if (contextual) {
            nextActiveID = contextual.id;
          } else {
            const created = createSession(resourceContext);
            nextSessions = [created, ...storedSessions];
            nextActiveID = created.id;
            persistSession(created);
          }
        }
        if (!nextSessions.length) {
          const created = createSession();
          nextSessions = [created];
          nextActiveID = created.id;
          persistSession(created);
        }
        if (!nextActiveID || !nextSessions.some((session) => session.id === nextActiveID)) {
          nextActiveID = nextSessions[0].id;
        }
        setSessions(nextSessions);
        setActiveSessionID(nextActiveID);
        setHydrated(true);
        void saveActiveAgentSessionID(nextActiveID).catch(reportStorageError);
      })
      .catch((error) => {
        if (cancelled) return;
        const created = createSession(resourceContext);
        setSessions([created]);
        setActiveSessionID(created.id);
        setHydrated(true);
        reportStorageError(error);
      });
    return () => { cancelled = true; };
  }, [persistSession, reportStorageError, resourceContext]);

  React.useEffect(() => {
    timelineRef.current?.scrollTo({ top: timelineRef.current.scrollHeight, behavior: 'smooth' });
  }, [activeSession?.messages.length, proposal, receipt]);

  const updateSession = React.useCallback((sessionID: string, updater: (session: AgentLocalSession) => AgentLocalSession) => {
    setSessions((current) => {
      let updated: AgentLocalSession | undefined;
      const next = current.map((session) => {
        if (session.id !== sessionID) return session;
        updated = updater(session);
        return updated;
      });
      if (updated) persistSession(updated);
      return next.sort((left, right) => right.updatedAt - left.updatedAt);
    });
  }, [persistSession]);

  const updateActiveSession = React.useCallback((updater: (session: AgentLocalSession) => AgentLocalSession) => {
    if (activeSessionID) updateSession(activeSessionID, updater);
  }, [activeSessionID, updateSession]);

  const appendSessionMessages = (sessionID: string, ...nextMessages: AgentChatMessage[]) => {
    updateSession(sessionID, (session) => ({
      ...session,
      messages: [...session.messages, ...nextMessages],
      updatedAt: Date.now(),
    }));
  };

  const replaceSessionMessage = (sessionID: string, id: string, nextMessage: AgentChatMessage) => {
    updateSession(sessionID, (session) => ({
      ...session,
      messages: session.messages.map((item) => (item.id === id ? nextMessage : item)),
      updatedAt: Date.now(),
    }));
  };

  const selectSession = (sessionID: string) => {
    if (sessionID === activeSessionID) return;
    setActiveSessionID(sessionID);
    idempotencyKey.current = '';
    void saveActiveAgentSessionID(sessionID).catch(reportStorageError);
  };

  const newSession = React.useCallback(() => {
    const created = createSession(resourceContext);
    setSessions((current) => [created, ...current]);
    setActiveSessionID(created.id);
    idempotencyKey.current = '';
    persistSession(created);
    void saveActiveAgentSessionID(created.id).catch(reportStorageError);
  }, [persistSession, reportStorageError, resourceContext]);

  React.useEffect(() => {
    const handleShortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase() === 'n') {
        event.preventDefault();
        newSession();
      }
    };
    window.addEventListener('keydown', handleShortcut);
    return () => window.removeEventListener('keydown', handleShortcut);
  }, [newSession]);

  const removeSession = async (sessionID: string) => {
    const remaining = sessions.filter((session) => session.id !== sessionID);
    let nextSessions = remaining;
    if (!nextSessions.length) {
      const created = createSession(resourceContext);
      nextSessions = [created];
      persistSession(created);
    }
    const nextActiveID = sessionID === activeSessionID ? nextSessions[0].id : activeSessionID;
    setSessions(nextSessions);
    setActiveSessionID(nextActiveID);
    await deleteAgentSession(sessionID).catch(reportStorageError);
    await saveActiveAgentSessionID(nextActiveID).catch(reportStorageError);
  };

  const startRename = (session: AgentLocalSession) => {
    setEditingSessionID(session.id);
    setEditingTitle(session.title);
  };

  const finishRename = () => {
    const value = editingTitle.trim();
    if (editingSessionID && value) {
      updateSession(editingSessionID, (session) => ({ ...session, title: value, updatedAt: Date.now() }));
    }
    setEditingSessionID('');
    setEditingTitle('');
  };

  const updateMemory = (patch: Partial<AgentMemorySettings>) => {
    updateActiveSession((session) => ({
      ...session,
      memory: { ...session.memory, ...patch },
      updatedAt: Date.now(),
    }));
  };

  const submit = async (preset?: string) => {
    if (!activeSession) return;
    const value = (preset ?? activeSession.draft).trim();
    if (!value || busy || confirming) return;
    const targetSessionID = activeSession.id;
    const userMessage = message('user', value);
    const history = activeSession.memory.enabled
      ? activeSession.messages.slice(-(activeSession.memory.maxTurns * 2)).map((item) => ({
        role: item.role === 'agent' ? 'assistant' as const : 'user' as const,
        content: item.content,
      }))
      : [];
    updateSession(targetSessionID, (session) => ({
      ...session,
      title: session.title === '新会话' ? titleFromMessage(value) : session.title,
      draft: '',
      messages: [...session.messages, userMessage],
      proposal: undefined,
      receipt: undefined,
      updatedAt: Date.now(),
    }));
    setBusySessionID(targetSessionID);
    idempotencyKey.current = '';

    try {
      const next = await sendAgentTurn({
        sessionId: targetSessionID,
        message: value,
        history,
        resourceContext: activeSession.resourceContext && {
          kind: activeSession.resourceContext.kind,
          namespace: activeSession.resourceContext.namespace,
          group: activeSession.resourceContext.group,
          name: activeSession.resourceContext.name,
        },
      });
      setRuntime(next.runtime);
      const toolMessages = (next.tools ?? []).map((tool) => ({
        ...message('agent', tool.detail || tool.summary),
        tool,
      }));
      updateSession(targetSessionID, (session) => ({
        ...session,
        messages: [...session.messages, ...toolMessages, message('agent', next.message)],
        proposal: next.proposal,
        receipt: undefined,
        updatedAt: Date.now(),
      }));
    } catch (error) {
      appendSessionMessages(targetSessionID, {
        ...message('agent', 'Agent 本轮执行失败，没有写入任何控制面资源。请检查运行时配置后重试。'),
        tool: {
          name: 'pole.agent.turn',
          summary: 'Agent 运行时调用失败',
          detail: error instanceof Error ? error.message : '未知错误',
          status: 'failed',
        },
      });
      openErrNotification('Agent 执行失败', toRequestErrorPayload(error));
    } finally {
      setBusySessionID((current) => current === targetSessionID ? '' : current);
    }
  };

  const confirm = async () => {
    if (!proposal || !activeSession || confirming) return;
    const targetSessionID = activeSession.id;
    if (!idempotencyKey.current) {
      idempotencyKey.current = globalThis.crypto?.randomUUID?.() || `${proposal.id}-${Date.now()}`;
    }
    setConfirming(true);
    const traceID = nextID();
    appendSessionMessages(targetSessionID, {
      ...message('agent', '收到确认，正在保存配置草稿。'),
      id: traceID,
      tool: {
        name: 'pole.config.update_file_draft',
        summary: '正在通过 Pole MCP 修改草稿',
        detail: `proposal ${proposal.id}`,
        status: 'running',
      },
    });
    try {
      const next = await confirmAgentProposal(proposal.id, proposal.version, proposal.previewHash, idempotencyKey.current);
      updateSession(targetSessionID, (session) => ({ ...session, receipt: next, updatedAt: Date.now() }));
      replaceSessionMessage(targetSessionID, traceID, {
        ...message('agent', '我没有创建发布版本，也没有改变当前生效配置。你可以继续留在这里，或进入配置分组完成发布。', '草稿已保存，正在等待你发布'),
        id: traceID,
        tool: {
          name: 'pole.config.update_file_draft',
          summary: '草稿修改成功 · 未发布',
          detail: next.requestId ? `Request ID ${next.requestId}` : next.resource.name,
          status: 'success',
        },
      });
      openInfoNotification('配置草稿已保存', '当前已进入待发布状态，尚未影响已发布配置');
    } catch (error) {
      replaceSessionMessage(targetSessionID, traceID, {
        ...message('agent', '保存草稿失败，临时视图仍然保留，你可以检查后重试。'),
        id: traceID,
        tool: {
          name: 'pole.config.update_file_draft',
          summary: 'MCP 工具调用失败',
          detail: `proposal ${proposal.id}`,
          status: 'failed',
        },
      });
      openErrNotification('保存草稿失败', toRequestErrorPayload(error));
    } finally {
      setConfirming(false);
    }
  };

  if (!hydrated || !activeSession) {
    return <div className={style.loading}>正在恢复本地会话…</div>;
  }

  const applyStarter = (value: string) => {
    updateActiveSession((session) => ({ ...session, draft: value, updatedAt: Date.now() }));
    requestAnimationFrame(() => composerRef.current?.focus());
  };

  return (
    <>
      {sidebarHost && createPortal(
        <section className={style.sessionSidebar} aria-label="本地会话">
          <div className={style.sessionTop}>
            <Button variant="text" className={style.newSession} onClick={newSession}>
              <span><AddIcon />新建会话</span><kbd>⌘ N</kbd>
            </Button>
            <Input
              className={style.sessionSearch}
              value={sessionQuery}
              prefixIcon={<SearchIcon />}
              placeholder="搜索会话"
              aria-label="搜索会话"
              onChange={setSessionQuery}
            />
          </div>
          <nav className={style.sessionList} aria-label="会话列表">
            {groupedSessions.map((group) => (
              <section key={group.label} className={style.sessionGroup}>
                <div className={style.sessionGroupLabel}>{group.label}</div>
                {group.sessions.map((session) => (
                  <div key={session.id} className={`${style.sessionItem} ${session.id === activeSessionID ? style.sessionItemActive : ''}`}>
                    {editingSessionID === session.id ? (
                      <Input
                        autoFocus
                        className={style.renameInput}
                        aria-label="会话名称"
                        value={editingTitle}
                        onChange={setEditingTitle}
                        onBlur={finishRename}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter') finishRename();
                          if (event.key === 'Escape') setEditingSessionID('');
                        }}
                      />
                    ) : (
                      <Button variant="text" className={style.sessionSelect} onClick={() => selectSession(session.id)}>
                        <span className={style.sessionCopy}>
                          <strong>{session.title}</strong>
                          <small>{sessionMeta(session)}</small>
                        </span>
                        <time>{sessionTime(session.updatedAt)}</time>
                      </Button>
                    )}
                    {editingSessionID !== session.id && (
                      <span className={style.sessionActions}>
                        <Tooltip content="重命名">
                          <Button shape="square" size="small" variant="text" icon={<EditIcon />} aria-label={`重命名 ${session.title}`} onClick={() => startRename(session)} />
                        </Tooltip>
                        <Tooltip content="删除会话">
                          <Button shape="square" size="small" variant="text" icon={<DeleteIcon />} aria-label={`删除 ${session.title}`} onClick={() => void removeSession(session.id)} />
                        </Tooltip>
                      </span>
                    )}
                  </div>
                ))}
              </section>
            ))}
            {!groupedSessions.length && <div className={style.sessionEmpty}>没有匹配的会话</div>}
          </nav>
        </section>,
        sidebarHost,
      )}

      <div className={style.page}>
        <header className={style.globalBar}>
          <div className={style.breadcrumb}><span>Pole Agent</span><span>/</span><strong>{activeSession.title}</strong></div>
          <HeaderIcon />
        </header>

        <section className={`${style.workspace} ${contextOpen ? '' : style.contextHidden}`}>
          <main className={style.conversation}>
            <header className={style.chatHeader}>
              <div className={style.chatIdentity}>
                <div className={style.titleRow}>
                  <h1>{activeSession.title}</h1>
                  <span
                    className={`${style.connection} ${runtime.ready ? style.connectionReady : style.connectionUnavailable}`}
                    title={runtime.reason}
                  >
                    <i /><PlugConnectedIcon />
                    {runtime.configured
                      ? runtime.ready ? 'Pole Agent 已就绪' : '模型已配置 · MCP 待连接'
                      : 'Agent 未配置'}
                  </span>
                </div>
                <p>写操作会先生成预览，确认后只保存为草稿</p>
              </div>
              <div className={style.chatActions}>
                <Button variant="text" icon={<SettingIcon />} onClick={() => setContextOpen(true)}>
                  记忆 <code>{activeSession.memory.enabled ? activeSession.memory.maxTurns : 0}</code> 轮
                </Button>
                <Tooltip content={contextOpen ? '收起上下文' : '展开上下文'}>
                  <Button shape="square" variant="text" icon={<ViewListIcon />} aria-label={contextOpen ? '收起上下文' : '展开上下文'} onClick={() => setContextOpen((current) => !current)} />
                </Tooltip>
              </div>
            </header>

            <div className={style.messageArea} ref={timelineRef}>
              {!activeSession.messages.length ? (
                <section className={style.emptyConversation}>
                  <div className={style.emptyInner}>
                    <span className={style.agentMark}><BotIcon /></span>
                    <p className={style.emptyKicker}>Pole Agent · 安全操作模式</p>
                    <h2>{activeSession.resourceContext ? '已接入当前配置' : '今天想管理什么？'}</h2>
                    <p className={style.emptyCopy}>
                      {activeSession.resourceContext
                        ? `已定位 ${activeSession.resourceContext.path}。描述目标状态，我会先生成临时视图。`
                        : '查询资源、解释配置或提出变更。Agent 会先分析当前上下文，任何写操作都会生成可审阅的变更预览。'}
                    </p>
                    <div className={style.taskGrid}>
                      <Button variant="text" onClick={() => applyStarter('读取 production/application/app.yaml，解释当前配置并指出可能的风险。')}>
                        <SettingIcon /><strong>检查配置文件</strong><span>读取真实草稿、解释配置并识别风险</span>
                      </Button>
                      <Button variant="text" onClick={() => applyStarter('列出我有权限查看的命名空间。')}>
                        <ServiceIcon /><strong>查询命名空间</strong><span>通过 Pole MCP 按当前身份读取</span>
                      </Button>
                      <Button variant="text" onClick={() => applyStarter('列出当前注册的 MCP Server 和可用工具。')}>
                        <IndicatorIcon /><strong>检查 MCP 能力</strong><span>查看已注册 Server 与工具目录</span>
                      </Button>
                    </div>
                  </div>
                </section>
              ) : (
                <section className={style.messageList} aria-live="polite">
                  {activeSession.messages.map((item) => (
                    <article key={item.id} className={`${style.message} ${item.role === 'user' ? style.userMessage : style.agentMessage}`}>
                      <span className={style.avatar}>{item.role === 'agent' ? <BotIcon /> : '你'}</span>
                      <div className={style.messageBody}>
                        <div className={style.messageHead}>
                          <strong>{item.role === 'agent' ? 'Pole Agent' : '你'}</strong>
                          <time>{sessionTime(item.createdAt)}</time>
                        </div>
                        {item.title && <strong className={style.messageTitle}>{item.title}</strong>}
                        <div className={style.messageContent}>{item.content}</div>
                        {item.tool && (
                          <details className={`${style.toolCall} ${style[item.tool.status]}`} defaultOpen={item.tool.status !== 'success'}>
                            <summary>
                              <span className={style.toolIcon}><ToolsCircleIcon /></span>
                              <span><code>{item.tool.name}</code><strong>{item.tool.summary}</strong></span>
                              <small>{item.tool.status === 'running' ? '调用中' : item.tool.status === 'success' ? '完成' : '失败'}</small>
                            </summary>
                            {item.tool.detail && <p>{item.tool.detail}</p>}
                          </details>
                        )}
                      </div>
                    </article>
                  ))}
                  {proposal && (
                    <article className={`${style.message} ${style.agentMessage}`}>
                      <span className={style.avatar}><BotIcon /></span>
                      <div className={style.messageBody}>
                        <div className={style.messageHead}><strong>Pole Agent</strong><time>{sessionTime(Date.now())}</time></div>
                        <section className={style.previewCard} aria-label="临时视图检查器">
                          <header><strong>变更预览</strong><Tag theme={receipt ? 'success' : 'warning'} variant="light">{receipt ? '已保存草稿' : '等待确认'}</Tag></header>
                          <div className={style.previewResource}>
                            <span>config.file</span>
                            <strong>{proposal.resource.namespace}/{proposal.resource.group}/{proposal.resource.name}</strong>
                          </div>
                          <div className={style.previewMeta}>
                            <div><span>基线</span><code>{proposal.baselineHash.slice(0, 12)}…</code></div>
                            <div><span>过期时间</span><strong>{new Date(proposal.expiresAt).toLocaleTimeString()}</strong></div>
                          </div>
                          <div className={style.diffSurface}>
                            <CodeDiffEditor
                              namespace={proposal.resource.namespace}
                              group={proposal.resource.group}
                              filename={proposal.resource.name}
                              curValue={proposal.before.content}
                              nextValue={proposal.after.content}
                              language={proposal.after.format}
                              readonly
                            />
                          </div>
                          <div className={style.safetyNote}><CheckCircleIcon /><span>{receipt ? '草稿已更新；当前发布版本保持不变。' : '确认只保存草稿，不会创建发布版本或立即生效。'}</span></div>
                          <footer>
                            {!receipt ? (
                              <><Button variant="outline" disabled={confirming} onClick={() => updateActiveSession((session) => ({ ...session, proposal: undefined }))}>丢弃预览</Button><Button theme="primary" icon={<CheckCircleIcon />} loading={confirming} onClick={confirm}>保存为草稿</Button></>
                            ) : (
                              <Button theme="primary" icon={<ArrowRightIcon />} onClick={() => navigate(receipt.detailUrl)}>前往配置分组发布</Button>
                            )}
                          </footer>
                        </section>
                      </div>
                    </article>
                  )}
                </section>
              )}
            </div>

            <div className={style.composerZone}>
              <div className={style.composer}>
                {activeSession.resourceContext && <div className={style.attachmentChip}><PlugConnectedIcon />{activeSession.resourceContext.path}</div>}
                <Textarea
                  className={style.composerInput}
                  ref={composerRef}
                  value={activeSession.draft}
                  placeholder="向 Pole Agent 提问或描述希望执行的操作"
                  aria-label="向 Pole Agent 提问"
                  autosize={{ minRows: 1, maxRows: 6 }}
                  onChange={(draft) => updateActiveSession((session) => ({ ...session, draft }))}
                  onKeyDown={(event: React.KeyboardEvent<HTMLTextAreaElement>) => {
                    if (event.key === 'Enter' && !event.shiftKey) {
                      event.preventDefault();
                      void submit();
                    }
                  }}
                />
                <div className={style.composerToolbar}>
                  <span className={style.composerScope}>{activeSession.resourceContext ? `当前范围：${activeSession.resourceContext.namespace}` : '当前范围：全部资源'}</span>
                  <Tooltip content="发送">
                    <Button className={style.composerSend} shape="square" theme="primary" icon={<SendIcon />} aria-label="发送消息" disabled={!runtime.ready || !activeSession.draft.trim() || busy || confirming} loading={busy} onClick={() => void submit()} />
                  </Tooltip>
                </div>
              </div>
              <p className={style.composerHelp}>Enter 发送 · Shift + Enter 换行 · 写操作不会直接发布资源</p>
            </div>
          </main>

          <aside className={style.contextPanel} aria-label="当前上下文">
            <div className={style.contextScroll}>
              <header><strong>当前上下文</strong><Button shape="square" variant="text" icon={<CloseIcon />} aria-label="收起上下文" onClick={() => setContextOpen(false)} /></header>
              <section className={style.contextGroup}>
                <div className={style.contextLabel}>资源范围</div>
                <div className={style.contextResource}><strong>{activeSession.resourceContext?.path || '全部资源'}</strong><span>{activeSession.resourceContext ? '配置文件 · 可生成草稿' : '所有命名空间 · 按权限查询'}</span></div>
              </section>
              <section className={style.contextGroup}>
                <div className={style.contextLabel}>连接状态</div>
                <div className={style.contextRow}><span>MCP 服务</span><strong className={runtime.mcpConnected ? '' : style.unavailable}><i />{runtime.mcpConnected ? 'pole-control-plane' : '尚未连接'}</strong></div>
                <div className={style.contextRow}><span>运行模式</span><code>{runtime.mode}</code></div>
                <div className={style.contextRow}><span>模型</span><code>{runtime.model || '未配置'}</code></div>
                <div className={style.contextRow}><span>Prompt</span><code>{runtime.promptVersion}</code></div>
                <div className={style.contextRow}><span>可用工具</span><code title={(runtime.tools ?? []).join(', ')}>{(runtime.tools ?? []).length ? `${runtime.tools.length} 个` : '未发现'}</code></div>
                {runtime.reason && <div className={style.runtimeReason}>{runtime.reason}</div>}
              </section>
              <section className={style.contextGroup}>
                <div className={style.memoryHeading}><div><span className={style.contextLabel}>记忆窗口</span></div><Switch checked={activeSession.memory.enabled} onChange={(enabled: boolean) => updateMemory({ enabled })} /></div>
                <div className={style.memoryWindows} aria-label="记忆窗口">
                  {[4, 10, 20].map((turns) => (
                    <Button variant="text" key={turns} className={activeSession.memory.maxTurns === turns && activeSession.memory.enabled ? style.memoryActive : ''} disabled={!activeSession.memory.enabled} onClick={() => updateMemory({ maxTurns: turns as AgentMemorySettings['maxTurns'] })}>{turns} 轮</Button>
                  ))}
                </div>
              </section>
              <section className={style.contextGroup}>
                <div className={style.contextLabel}>操作权限</div>
                <div className={style.contextRow}><span>资源查询</span><strong><i />允许</strong></div>
                <div className={style.contextRow}><span>生成变更</span><strong className={style.review}><i />需确认</strong></div>
                <div className={style.contextRow}><span>直接发布</span><code>禁止</code></div>
              </section>
              <section className={style.contextGroup}><div className={style.contextSafety}><LockOnIcon /><span>Agent 只能保存草稿。最终发布仍由你在对应工作台完成。</span></div></section>
            </div>
          </aside>
        </section>
      </div>
    </>
  );
}
