import React, { useMemo } from 'react';
import { Button, Tag } from 'components/Fluent';
import {
  BotIcon,
  CheckCircleIcon,
  EditIcon,
  LinkIcon,
  LockOnIcon,
  PlugConnectedIcon,
  RocketIcon,
  ToolsCircleIcon,
} from 'components/Fluent/icons';
import { AgentSystemDomain, AgentSystemProfile } from 'services/system_configuration';

import style from '../index.module.less';

interface Props {
  domain?: AgentSystemDomain;
  fallback: AgentSystemProfile;
  apiKeyConfigured: boolean;
  loading: boolean;
  onEdit: () => void;
  onPublish: () => void;
}

const displayHost = (value: string) => {
  if (!value) return '未配置';
  try {
    return new URL(value).host || value;
  } catch {
    return value;
  }
};

const formatDate = (value?: string) => {
  if (!value) return '—';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN', { hour12: false });
};

export default function AgentConfigurationWorkspace({
  domain,
  fallback,
  apiKeyConfigured,
  loading,
  onEdit,
  onPublish,
}: Props) {
  const effective = domain?.active?.values || fallback;
  const draft = domain?.draft;
  const draftValues = draft?.values;
  const effectiveSecret = domain?.active?.secret;
  const draftSecret = draft?.secret;
  const secretConfigured = Boolean(effectiveSecret?.configured || apiKeyConfigured);
  const hasRuntimeConfiguration = Boolean(effective.baseURL && effective.model && secretConfigured);
  const status = loading
    ? { label: '正在读取', className: style.agentStateWarning }
    : !domain
      ? { label: '运行状态不可用', className: style.agentStateDanger }
      : domain.applyStatus === 'rejected'
        ? { label: '应用异常', className: style.agentStateDanger }
        : hasRuntimeConfiguration
          ? { label: '运行配置就绪', className: style.agentStateSuccess }
          : { label: '等待配置', className: style.agentStateWarning };
  const toolCount = effective.mcpToolAllowlist.length;

  const draftChanges = useMemo(() => {
    if (!draftValues) return [];
    const rows = [
      ['Gateway', effective.baseURL, draftValues.baseURL],
      ['模型', effective.model, draftValues.model],
      ['请求超时', effective.modelTimeout, draftValues.modelTimeout],
      ['MCP Endpoint', effective.mcpEndpoint, draftValues.mcpEndpoint],
      ['工具白名单', effective.mcpToolAllowlist.join(', '), draftValues.mcpToolAllowlist.join(', ')],
      ['Agent 指令', effective.operatorInstructions, draftValues.operatorInstructions],
    ];
    return rows.filter(([, before, after]) => before !== after);
  }, [draftValues, effective]);

  return (
    <div className={style.agentWorkbench} aria-busy={loading}>
      <header className={style.agentWorkbenchHeader}>
        <div className={style.agentIdentity}>
          <span className={style.agentIdentityIcon}><BotIcon /></span>
          <div>
            <div className={style.agentTitleLine}>
              <h2>Pole Agent</h2>
              <span className={status.className}>{status.label}</span>
            </div>
            <p>配置模型连接、Pole Secret、MCP 工具与运行指令。</p>
          </div>
        </div>
        <div className={style.agentPrimaryActions}>
          {draft ? (
            <>
              <Button variant="outline" icon={<EditIcon />} onClick={onEdit}>继续编辑</Button>
              <Button theme="primary" icon={<RocketIcon />} onClick={onPublish}>
                审阅并发布 r{draft.revision}
              </Button>
            </>
          ) : (
            <Button theme="primary" icon={<EditIcon />} loading={loading} disabled={!domain} onClick={onEdit}>
              {hasRuntimeConfiguration ? '编辑配置' : '完成首次配置'}
            </Button>
          )}
        </div>
      </header>

      <div className={style.agentReadiness} aria-label="Agent 运行状态">
        <div>
          <span>当前生效</span>
          <strong>{domain?.active ? `r${domain.active.revision}` : '静态启动配置'}</strong>
          <small>{domain?.applyStatus === 'applied' ? '已热更新到当前实例' : '当前运行时基线'}</small>
        </div>
        <div>
          <span>模型</span>
          <strong>{effective.model || '未配置'}</strong>
          <small>{displayHost(effective.baseURL)}</small>
        </div>
        <div>
          <span>Pole Secret</span>
          <strong>{secretConfigured ? `已托管${effectiveSecret?.version ? ` · v${effectiveSecret.version}` : ''}` : '未配置'}</strong>
          <small>{domain?.secretStoreReady ? '加密存储可用' : 'Secret Store 未就绪'}</small>
        </div>
        <div>
          <span>待发布草稿</span>
          <strong>{draft ? `r${draft.revision}` : '无'}</strong>
          <small>{draft ? `${draftChanges.length} 项配置变化` : '运行配置与发布状态一致'}</small>
        </div>
      </div>

      {draft && (
        <Button variant="text" className={style.agentDraftBanner} onClick={onPublish}>
          <span className={style.agentDraftIcon}><EditIcon /></span>
          <span>
            <strong>草稿 r{draft.revision} 等待审阅</strong>
            <small>{draft.createdBy || 'admin'} 创建于 {formatDate(draft.createdAt)}，尚未影响运行中的 Agent。</small>
          </span>
          <span className={style.agentDraftAction}>查看变更</span>
        </Button>
      )}

      <div className={style.agentWorkbenchBody}>
        <main className={style.agentCapabilityList}>
          <section className={style.agentCapability}>
            <span className={style.agentCapabilityIcon}><PlugConnectedIcon /></span>
            <div className={style.agentCapabilityContent}>
              <div className={style.agentCapabilityHeading}>
                <div>
                  <h3>模型连接</h3>
                  <p>OpenAI-compatible Gateway 与模型运行参数</p>
                </div>
                <Tag variant="outline">{effective.provider || 'openai-compatible'}</Tag>
              </div>
              <dl className={style.agentDefinitionList}>
                <div><dt>Gateway</dt><dd>{effective.baseURL || '未配置'}</dd></div>
                <div><dt>模型</dt><dd>{effective.model || '未配置'}</dd></div>
                <div><dt>请求超时</dt><dd>{effective.modelTimeout || '60s'}</dd></div>
              </dl>
            </div>
          </section>

          <section className={style.agentCapability}>
            <span className={style.agentCapabilityIcon}><LockOnIcon /></span>
            <div className={style.agentCapabilityContent}>
              <div className={style.agentCapabilityHeading}>
                <div>
                  <h3>Pole Secret</h3>
                  <p>API Key 由 Pole 加密托管，读取接口永不回填正文</p>
                </div>
                <span className={secretConfigured ? style.agentInlineReady : style.agentInlineMissing}>
                  {secretConfigured ? <><CheckCircleIcon /> 已配置</> : '需要配置'}
                </span>
              </div>
              <dl className={style.agentDefinitionList}>
                <div><dt>存储状态</dt><dd>{domain?.secretStoreReady ? 'Envelope encryption 已就绪' : '根加密材料未就绪'}</dd></div>
                <div><dt>版本</dt><dd>{effectiveSecret?.version ? `Secret v${effectiveSecret.version}` : '—'}</dd></div>
                <div><dt>最近轮换</dt><dd>{formatDate(effectiveSecret?.rotatedAt)}</dd></div>
              </dl>
            </div>
          </section>

          <section className={style.agentCapability}>
            <span className={style.agentCapabilityIcon}><ToolsCircleIcon /></span>
            <div className={style.agentCapabilityContent}>
              <div className={style.agentCapabilityHeading}>
                <div>
                  <h3>MCP 能力</h3>
                  <p>Agent 只加载显式允许的 Pole 控制面工具</p>
                </div>
                <span className={style.agentInlineMeta}>{toolCount} 个工具</span>
              </div>
              <div className={style.agentEndpoint}><LinkIcon /><span>{effective.mcpEndpoint || '未配置 MCP Endpoint'}</span></div>
              <div className={style.agentToolList}>
                {effective.mcpToolAllowlist.length
                  ? effective.mcpToolAllowlist.map((tool) => <span key={tool}>{tool}</span>)
                  : <em>尚未配置工具白名单</em>}
              </div>
            </div>
          </section>

          <section className={style.agentCapability}>
            <span className={style.agentCapabilityIcon}><BotIcon /></span>
            <div className={style.agentCapabilityContent}>
              <div className={style.agentCapabilityHeading}>
                <div>
                  <h3>Agent 指令</h3>
                  <p>内置安全指令 + 管理员补充的操作约束</p>
                </div>
                <span className={style.agentInlineMeta}>{effective.promptVersion || 'v1'}</span>
              </div>
              <p className={style.agentInstructionPreview}>
                {effective.operatorInstructions || '未设置额外操作指令，将使用 Pole 内置安全策略。'}
              </p>
            </div>
          </section>
        </main>

        <aside className={style.agentReleaseRail}>
          <div className={style.agentReleaseHeading}>
            <span>发布流程</span>
            <small>三步独立确认</small>
          </div>
          <ol className={style.agentReleaseSteps}>
            <li className={draft ? style.agentReleaseStepDone : style.agentReleaseStepCurrent}>
              <span>1</span>
              <div><strong>配置</strong><small>{draft ? `草稿 r${draft.revision} 已保存` : '编辑后保存为草稿'}</small></div>
            </li>
            <li className={draft ? style.agentReleaseStepCurrent : undefined}>
              <span>2</span>
              <div><strong>验证与审阅</strong><small>检查连接和字段变化</small></div>
            </li>
            <li>
              <span>3</span>
              <div><strong>发布</strong><small>复验后原子热更新</small></div>
            </li>
          </ol>
          <div className={style.agentSafetyNote}>
            <CheckCircleIcon />
            <p><strong>运行时保护</strong>发布失败时继续使用上一次有效配置，不会切换到不完整草稿。</p>
          </div>
          {domain?.applyMessage && (
            <div className={domain.applyStatus === 'rejected' ? style.agentApplyError : style.agentApplyMessage}>
              {domain.applyMessage}
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}
