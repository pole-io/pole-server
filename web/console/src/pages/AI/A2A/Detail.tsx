import React from 'react';
import { Breadcrumb, Button, Empty, Link, Loading, Space, Tabs, Tag, Textarea, Tooltip } from 'components/Fluent';
import { RefreshIcon, ServerIcon } from 'components/Fluent/icons';
import { useNavigate, useSearchParams } from 'components/Router';

import AuthorizeInput from 'components/Authorize';
import AIEnvironmentBinding from 'components/AIEnvironmentBinding';
import EnvironmentResourceSwitcher from 'components/EnvironmentResourceSwitcher';
import { useAppDispatch } from 'modules/store';
import { editorA2AAgent } from 'modules/ai/a2a';
import {
  A2AAgent,
  A2AAgentSkill,
  A2AEnvironmentBinding,
  describeA2AAgentCard,
  describeA2AAgentDefinitionBinding,
  describeA2AAgentDefinitionEnvironments,
  describeA2AAgentDefinitions,
  describeA2AAgents,
  describeA2AAgentSkills,
  bindA2AAgentDefinition,
  createA2AAgentDefinition,
} from 'services/a2a';
import { PolicySourceType } from 'services/auth_policy';
import {
  A2AEditor,
  AgentCardView,
  AgentSkillsView,
  backendLabel,
  backendServiceRef,
  backendType,
  backendTypeLabel,
  jsonPreview,
  parseJSONValue,
  protocolLabel,
  protocolTheme,
} from './index';
import style from './index.module.less';

const { BreadcrumbItem } = Breadcrumb;
const { TabPanel } = Tabs;

type DetailTab = 'card' | 'skills' | 'edit';

const normalizeTab = (value?: string | null): DetailTab => {
  if (value === 'skills' || value === 'edit') return value;
  return 'card';
};

const A2ADetailPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const id = searchParams.get('id') || '';
  const namespace = searchParams.get('namespace') || '';
  const name = searchParams.get('name') || '';
  const [agent, setAgent] = React.useState<A2AAgent | null>(null);
  const [skills, setSkills] = React.useState<A2AAgentSkill[]>([]);
  const [card, setCard] = React.useState<Record<string, any> | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [skillsLoading, setSkillsLoading] = React.useState(false);
  const [cardLoading, setCardLoading] = React.useState(false);
  const [agentError, setAgentError] = React.useState('');
  const [agentNotFound, setAgentNotFound] = React.useState(false);
  const [skillsError, setSkillsError] = React.useState('');
  const [cardError, setCardError] = React.useState('');
  const [environmentBindings, setEnvironmentBindings] = React.useState<A2AEnvironmentBinding[]>([]);
  const [activeTab, setActiveTab] = React.useState<DetailTab>(normalizeTab(searchParams.get('tab')));
  const [cardViewMode, setCardViewMode] = React.useState<'visual' | 'raw'>('visual');
  const [authorizeVisible, setAuthorizeVisible] = React.useState(false);

  const loadAgent = React.useCallback(async () => {
    setLoading(true);
    setAgentError('');
    setAgentNotFound(false);
    try {
      const response = await describeA2AAgents({
        offset: 0,
        limit: 10000,
        name: name || undefined,
        namespace: namespace || undefined,
      });
      const next = response.list.find((item) => (id ? item.id === id : item.name === name && item.namespace === namespace)) || null;
      setAgent(next);
      setEnvironmentBindings([]);
      if (next) {
        dispatch(editorA2AAgent(next));
        if (next.id && next.namespace !== 'pole-system') {
          try {
            const binding = await describeA2AAgentDefinitionBinding(next.id);
            if (binding?.definition_id) {
              setEnvironmentBindings(await describeA2AAgentDefinitionEnvironments(binding.definition_id));
            }
          } catch {
            // 旧记录允许暂时保持未关联状态。
          }
        }
      } else {
        setAgentNotFound(true);
      }
    } catch (error) {
      setAgent(null);
      setAgentError((error as Error).message || '请求失败');
    } finally {
      setLoading(false);
    }
  }, [dispatch, id, name, namespace]);

  const loadSkills = React.useCallback(async (target?: A2AAgent | null) => {
    const current = target || agent;
    if (!current?.id) {
      setSkills([]);
      setSkillsError('');
      return;
    }
    setSkillsLoading(true);
    setSkillsError('');
    try {
      const response = await describeA2AAgentSkills({ agent_id: current.id });
      setSkills(response.list);
    } catch (error) {
      setSkillsError((error as Error).message || '请求失败');
    } finally {
      setSkillsLoading(false);
    }
  }, [agent]);

  const loadCard = React.useCallback(async (target?: A2AAgent | null) => {
    const current = target || agent;
    if (!current?.id) {
      setCard(null);
      setCardError('');
      return;
    }
    setCardLoading(true);
    setCardError('');
    try {
      const response = await describeA2AAgentCard(current.id);
      setCard(response);
    } catch (error) {
      setCardError((error as Error).message || '请求失败');
    } finally {
      setCardLoading(false);
    }
  }, [agent]);

  React.useEffect(() => {
    loadAgent();
  }, [loadAgent]);

  React.useEffect(() => {
    loadSkills(agent);
    loadCard(agent);
  }, [agent, loadCard, loadSkills]);

  const refreshActive = () => {
    if (activeTab === 'skills') {
      loadSkills(agent);
      return;
    }
    if (activeTab === 'edit') {
      loadAgent();
      return;
    }
    loadCard(agent);
  };

  const goBackendService = () => {
    const service = backendServiceRef(agent || undefined);
    if (!service) return;
    navigate(`/discovery/service/instance?namespace=${encodeURIComponent(service.namespace)}&service=${encodeURIComponent(service.name)}`);
  };

  const cardText = React.useMemo(() => jsonPreview(card), [card]);
  const cardObject = React.useMemo(() => parseJSONValue(card), [card]);

  return (
    <div className={style.page}>
      <Breadcrumb className={style.detailBreadcrumb} maxItemWidth="220px">
        <BreadcrumbItem onClick={() => navigate('/ai/a2a')}>A2A Agent</BreadcrumbItem>
        <BreadcrumbItem>{agent ? `${agent.namespace}/${agent.name}` : name || 'A2A Agent 详情'}</BreadcrumbItem>
      </Breadcrumb>

      {loading && (
        <section className={style.detailLoading}>
          <Loading text="加载 A2A Agent 详情..." />
        </section>
      )}

      {!loading && !agent && (
        <section className={style.detailLoading}>
          <Empty
            title={agentNotFound ? '未找到 A2A Agent' : 'A2A Agent 加载失败'}
            description={agentNotFound ? '该资源可能已删除，或当前链接参数已经失效。' : agentError}
            action={<Button variant="outline" icon={<RefreshIcon />} onClick={loadAgent}>重新加载</Button>}
          />
        </section>
      )}

      {!loading && agent && (
        <div className={style.agentDrawer}>
          {environmentBindings.length > 0 ? (
            <EnvironmentResourceSwitcher
              currentNamespace={agent.namespace}
              items={environmentBindings.map((binding) => ({
                namespace: binding.namespace,
                summary: binding.resource_name,
              }))}
              resourceLabel="A2A Agent"
              presentation="tabs"
              onSelect={(targetNamespace) => {
                const target = environmentBindings.find((binding) => binding.namespace === targetNamespace);
                if (target) {
                  navigate(`/ai/a2a/detail?id=${encodeURIComponent(target.resource_id)}&namespace=${encodeURIComponent(target.namespace)}&name=${encodeURIComponent(target.resource_name)}`);
                }
              }}
            />
          ) : agent.namespace !== 'pole-system' && agent.id ? (
            <Space>
              <Tag variant="light">未关联跨环境逻辑定义</Tag>
              <AIEnvironmentBinding
                resourceLabel="A2A Agent"
                suggestedName={agent.name}
                loadDefinitions={describeA2AAgentDefinitions}
                createDefinition={createA2AAgentDefinition}
                bindDefinition={(definitionId) => bindA2AAgentDefinition(definitionId, agent.id!)}
                onBound={loadAgent}
              />
            </Space>
          ) : null}
          <section className={style.agentDrawerSummary}>
            <div className={style.agentDrawerIcon}>
              <ServerIcon />
            </div>
            <div className={style.agentDrawerMain}>
              <div className={style.agentDrawerTitle}>
                <h3>{agent.namespace}/{agent.name}</h3>
                <Tag theme={protocolTheme(agent.preferred_protocol_binding) as any} variant="light">
                  {protocolLabel(agent.preferred_protocol_binding)}
                </Tag>
              </div>
              <div className={style.agentDrawerDesc}>
                {agent.description || agent.provider_organization || '该 Agent 暂无描述。'}
              </div>
              <div className={style.agentMetaGrid}>
                <div>
                  <span>技能数</span>
                  <strong>{skillsLoading || skillsError ? '-' : (agent.skills?.length || skills.length || 0)}</strong>
                </div>
                <div>
                  <span>后端</span>
                  <strong>{backendTypeLabel(backendType(agent))}</strong>
                </div>
                <div>
                  <span>接入地址</span>
                  <strong>
                    {backendServiceRef(agent) ? (
                      <Link theme="primary" onClick={goBackendService}>
                        {backendLabel(agent)}
                      </Link>
                    ) : backendLabel(agent)}
                  </strong>
                </div>
                <div>
                  <span>最近修改</span>
                  <strong>{agent.mtime || '-'}</strong>
                </div>
              </div>
            </div>
            <div className={style.agentDrawerActions}>
              <Tooltip content="刷新">
                <Button aria-label="刷新 A2A Agent 当前视图" className={style.drawerActionIconButton} shape="square" variant="outline" onClick={refreshActive}>
                  <RefreshIcon />
                </Button>
              </Tooltip>
              <Button variant="outline" onClick={() => setActiveTab('edit')}>
                编辑 Agent
              </Button>
              <Button variant="outline" onClick={() => setAuthorizeVisible(true)}>
                授权
              </Button>
            </div>
          </section>

          <Tabs className={style.agentDetailTabs} value={activeTab} onChange={(value) => setActiveTab(value as DetailTab)}>
            <TabPanel value="card" label="Agent Card">
              <section className={style.detailSurface}>
                <div className={style.tableHeader}>
                  <div>
                    <strong>Agent Card</strong>
                    <span>{cardViewMode === 'visual' ? '概览' : '原始 Card JSON'}</span>
                  </div>
                  <Space className={style.cardViewSwitch} size={4}>
                    <Button
                      size="small"
                      theme={cardViewMode === 'visual' ? 'primary' : 'default'}
                      variant={cardViewMode === 'visual' ? 'base' : 'outline'}
                      onClick={() => setCardViewMode('visual')}
                    >
                      概览
                    </Button>
                    <Button
                      size="small"
                      theme={cardViewMode === 'raw' ? 'primary' : 'default'}
                      variant={cardViewMode === 'raw' ? 'base' : 'outline'}
                      onClick={() => setCardViewMode('raw')}
                    >
                      原始 JSON
                    </Button>
                  </Space>
                </div>
                {cardError ? (
                  <Empty title="Agent Card 加载失败" description={cardError} action={<Button variant="outline" onClick={() => loadCard(agent)}>重新加载</Button>} />
                ) : cardViewMode === 'raw' ? (
                  <Textarea
                    className={style.codeText}
                    readonly
                    value={cardLoading ? '加载中...' : cardText}
                    autosize={{ minRows: 18, maxRows: 26 }}
                  />
                ) : (
                  <AgentCardView card={cardObject} loading={cardLoading} />
                )}
              </section>
            </TabPanel>
            <TabPanel value="skills" label="技能">
              <section className={style.detailSurface}>
                <div className={style.tableHeader}>
                  <div>
                    <strong>技能目录</strong>
                    <span>{skillsLoading ? '正在同步技能' : `当前显示 ${skills.length} 个技能`}</span>
                  </div>
                </div>
                {skillsError ? (
                  <Empty title="技能目录加载失败" description={skillsError} action={<Button variant="outline" onClick={() => loadSkills(agent)}>重新加载</Button>} />
                ) : (
                  <AgentSkillsView skills={skills} loading={skillsLoading} />
                )}
              </section>
            </TabPanel>
            <TabPanel value="edit" label="编辑">
              <section className={style.detailSurface}>
                <div className={style.tableHeader}>
                  <div>
                    <strong>编辑 Agent</strong>
                    <span>修改基础信息、接入方式、技能和 Agent Card 原文</span>
                  </div>
                </div>
                <div className={style.embeddedEditor}>
                  <A2AEditor
                    op="edit"
                    visible={activeTab === 'edit'}
                    embedded
                    closeDrawer={() => {
                      loadAgent();
                      setActiveTab('card');
                    }}
                  />
                </div>
              </section>
            </TabPanel>
          </Tabs>
        </div>
      )}

      {authorizeVisible && agent?.id && (
        <AuthorizeInput
          resource_type={PolicySourceType.A2AAgentResources}
          resource_id={agent.id}
          resource_name={`${agent.namespace}/${agent.name}`}
          visible={authorizeVisible}
          onClose={() => setAuthorizeVisible(false)}
        />
      )}
    </div>
  );
};

export default React.memo(A2ADetailPage);
