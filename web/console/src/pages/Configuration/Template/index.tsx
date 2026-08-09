import React from 'react';
import {
  Button,
  Dialog,
  Input,
  PrimaryTableProps,
  Radio,
  RadioGroup,
  Select,
  Space,
  Table,
  Tag,
  TagInput,
  Tabs,
  Textarea,
  Tooltip,
} from 'components/Fluent';
import { AddIcon, Edit1Icon, RefreshIcon, RocketIcon, RollbackIcon, SaveIcon } from 'components/Fluent/icons';
import CodeEditor from 'components/CodeEditor';
import { ResourceHeader, ResourceToolbar } from 'components/ResourceLayout';
import { useNavigate, useSearchParams } from 'components/Router';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { toRequestErrorPayload } from 'utils/request';
import { describeBusinessNamespaces } from 'services/namespace';
import {
  ConfigFileTemplate,
  ConfigTemplateRelease,
  ConfigTemplateValue,
  createConfigTemplate,
  describeConfigTemplateLabels,
  describeConfigTemplateReleases,
  describeConfigTemplates,
  describeNamespaceTemplateValueReleases,
  describeNamespaceTemplateValues,
  NamespaceTemplateValueRelease,
  previewConfigTemplate,
  publishConfigTemplateRelease,
  publishNamespaceTemplateValueRelease,
  RenderPreview,
  saveNamespaceTemplateValues,
  saveConfigTemplateLabels,
  TemplateEngine,
  TemplateValueReleaseType,
  updateConfigTemplate,
} from 'services/config_templates';
import GrayRuleEditor, { defaultGrayRuleRow, grayRowsToBetaLabels } from '../Group/Releases/GrayRuleEditor';
import { TrafficMatchConditionRow } from 'pages/Governance/shared/TrafficMatchConditionEditor';
import SchemaEditor, { emptySchemaParameter } from './SchemaEditor';
import ValueEditor, { textToValue, validateTemplateValues } from './ValueEditor';
import RenderPreviewPanel from './RenderPreviewPanel';
import VersionManager from './VersionManager';
import styles from './index.module.less';
import GroupWorkspaceNav from '../Group/GroupWorkspaceNav';

const { TabPanel } = Tabs;

type WorkspaceTab = 'content' | 'schema' | 'values' | 'versions' | 'basic';

interface TemplateWorkspaceProps {
  embedded?: boolean;
  templateId?: string;
  namespace?: string;
  group?: string;
}

const emptyTemplate = (): ConfigFileTemplate => ({
  id: 0,
  name: '',
  content: '',
  comment: '',
  format: 'text',
  engine: TemplateEngine,
  parameterSchema: [],
  labels: {},
});

const initialValues = (template: ConfigFileTemplate) =>
  Object.fromEntries(
    template.parameterSchema.map((parameter) => [
      parameter.name,
      parameter.defaultValue ||
        textToValue(
          parameter.type,
          parameter.type === 'TEMPLATE_PARAMETER_BOOLEAN'
            ? 'false'
            : parameter.type === 'TEMPLATE_PARAMETER_INTEGER' || parameter.type === 'TEMPLATE_PARAMETER_DECIMAL'
            ? '0'
            : ''
        ),
    ])
  );

const TemplateWorkspace: React.FC<TemplateWorkspaceProps> = ({
  embedded = false,
  templateId = '',
  namespace: embeddedNamespace = '',
  group: embeddedGroup = '',
}) => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const requestedTemplateId = templateId || searchParams.get('templateId') || '';
  const requestedNamespace = embeddedNamespace || searchParams.get('namespace') || '';
  const requestedGroup = embeddedGroup || searchParams.get('group') || '';
  const requestedTab = searchParams.get('tab');
  const returnTo = searchParams.get('returnTo') || '';
  const [templates, setTemplates] = React.useState<ConfigFileTemplate[]>([]);
  const [selectedId, setSelectedId] = React.useState<string>('');
  const [draft, setDraft] = React.useState<ConfigFileTemplate>(emptyTemplate());
  const [editing, setEditing] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [activeTab, setActiveTab] = React.useState<WorkspaceTab>('content');
  const [releases, setReleases] = React.useState<ConfigTemplateRelease[]>([]);
  const [templateReleaseId, setTemplateReleaseId] = React.useState('');
  const [namespaces, setNamespaces] = React.useState<string[]>([]);
  const [namespace, setNamespace] = React.useState('');
  const [valuesId, setValuesId] = React.useState('');
  const [valuesRevision, setValuesRevision] = React.useState('');
  const [values, setValues] = React.useState<Record<string, ConfigTemplateValue>>({});
  const [valueReleases, setValueReleases] = React.useState<NamespaceTemplateValueRelease[]>([]);
  const [valueReleaseType, setValueReleaseType] = React.useState<TemplateValueReleaseType>(
    'TEMPLATE_VALUE_RELEASE_NORMAL'
  );
  const [valueComment, setValueComment] = React.useState('');
  const [grayRows, setGrayRows] = React.useState<TrafficMatchConditionRow[]>([defaultGrayRuleRow()]);
  const [valueDraftPreview, setValueDraftPreview] = React.useState<RenderPreview>();
  const [versionPreview, setVersionPreview] = React.useState<RenderPreview>();
  const [previewTemplateReleaseId, setPreviewTemplateReleaseId] = React.useState('');
  const [previewValueReleaseId, setPreviewValueReleaseId] = React.useState('');
  const [templatePublishOpen, setTemplatePublishOpen] = React.useState(false);
  const [templateReleaseComment, setTemplateReleaseComment] = React.useState('');
  const [templatePublishing, setTemplatePublishing] = React.useState(false);
  const [valuePublishOpen, setValuePublishOpen] = React.useState(false);
  const [valuePublishing, setValuePublishing] = React.useState(false);

  const selected = React.useMemo(
    () => templates.find((item) => String(item.id) === selectedId),
    [selectedId, templates]
  );
  const valueValidationErrors = React.useMemo(
    () => validateTemplateValues(draft.parameterSchema, values),
    [draft.parameterSchema, values]
  );
  const hasValueErrors = Object.keys(valueValidationErrors).length > 0;
  const templateDefinitionTab = activeTab === 'content' || activeTab === 'schema' || activeTab === 'basic';
  const nextTemplateVersion = React.useMemo(
    () => Math.max(0, ...releases.map((release) => Number(release.version) || 0)) + 1,
    [releases]
  );

  const ensureValidValues = () => {
    const invalidEntry = Object.entries(valueValidationErrors)[0];
    if (!invalidEntry) return true;
    setActiveTab('values');
    openErrNotification('Value 校验失败', `${invalidEntry[0]}：${invalidEntry[1]}`);
    return false;
  };

  const loadTemplates = React.useCallback(
    async (preferredName?: string, preferredId?: string) => {
      setLoading(true);
      try {
        const result = await describeConfigTemplates();
        setTemplates(result.templates);
        const preferred = preferredId
          ? result.templates.find((item) => String(item.id) === preferredId)
          : preferredName
          ? result.templates.find((item) => item.name === preferredName)
          : result.templates.find((item) => String(item.id) === selectedId) || result.templates[0];
        if (preferred) setSelectedId(String(preferred.id));
        else {
          setSelectedId('');
          setDraft(emptyTemplate());
        }
      } catch (error) {
        openErrNotification('加载配置模板失败', toRequestErrorPayload(error));
      } finally {
        setLoading(false);
      }
    },
    [selectedId]
  );

  React.useEffect(() => {
    loadTemplates(undefined, requestedTemplateId);
    describeBusinessNamespaces()
      .then((items) => {
        const names = items.map((item) => item.name);
        setNamespaces(names);
        setNamespace((current) => requestedNamespace || current || names[0] || '');
      })
      .catch((error) => openErrNotification('加载命名空间失败', toRequestErrorPayload(error)));
    // 初始化只执行一次，后续刷新由显式操作触发。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  React.useEffect(() => {
    if (!selected) return;
    setDraft({
      ...selected,
      engine: selected.engine || TemplateEngine,
      parameterSchema: selected.parameterSchema.map((item) => ({ ...item })),
    });
    if (selected.id) {
      describeConfigTemplateLabels(selected.id)
        .then((labels) =>
          setDraft((current) => (String(current.id) === String(selected.id) ? { ...current, labels } : current))
        )
        .catch((error) => openErrNotification('加载模板标签失败', toRequestErrorPayload(error)));
    }
    setEditing(false);
    setTemplatePublishOpen(false);
    setValuePublishOpen(false);
    setTemplateReleaseComment('');
    setValueDraftPreview(undefined);
    setVersionPreview(undefined);
    const requestedTabMap: Record<string, WorkspaceTab> = {
      definition: 'basic',
      content: 'content',
      schema: 'schema',
      values: 'values',
      releases: 'versions',
      preview: 'versions',
      versions: 'versions',
      basic: 'basic',
    };
    setActiveTab((requestedTab && requestedTabMap[requestedTab]) || 'content');
  }, [requestedTab, selected]);

  React.useEffect(() => {
    if (!draft.id) {
      setReleases([]);
      setTemplateReleaseId('');
      setPreviewTemplateReleaseId('');
      setValueDraftPreview(undefined);
      return;
    }
    describeConfigTemplateReleases(draft.id)
      .then((result) => {
        setReleases(result.releases);
        setTemplateReleaseId((current) =>
          result.releases.some((item) => item.id === current) ? current : result.releases[0]?.id || ''
        );
        setPreviewTemplateReleaseId((current) =>
          result.releases.some((item) => item.id === current) ? current : ''
        );
      })
      .catch((error) => openErrNotification('加载模板发布版本失败', toRequestErrorPayload(error)));
  }, [draft.id]);

  React.useEffect(() => {
    if (!draft.id || !namespace) {
      setValues(initialValues(draft));
      setValueReleases([]);
      setPreviewValueReleaseId('');
      setValueDraftPreview(undefined);
      return;
    }
    Promise.all([
      describeNamespaceTemplateValues(namespace, draft.id),
      describeNamespaceTemplateValueReleases(namespace, draft.id),
    ])
      .then(([saved, history]) => {
        setValuesId(saved?.id || '');
        setValuesRevision(saved?.revision || '');
        setValues(saved?.values && Object.keys(saved.values).length ? saved.values : initialValues(draft));
        setValueReleases(history.releases);
        setValueDraftPreview(undefined);
        setPreviewValueReleaseId((current) =>
          history.releases.some((item) => item.id === current) ? current : ''
        );
      })
      .catch((error) => openErrNotification('加载 Namespace Value 失败', toRequestErrorPayload(error)));
  }, [draft.id, namespace]);

  const validateDraft = () => {
    if (!draft.name.trim()) return '模板名称不能为空';
    if (!draft.content) return '模板内容不能为空';
    const names = draft.parameterSchema.map((item) => item.name.trim());
    if (names.some((name) => !name)) return 'Schema 参数名不能为空';
    if (new Set(names).size !== names.length) return 'Schema 参数名不能重复';
    return '';
  };

  const startSchemaEditing = React.useCallback((addParameter = false) => {
    setActiveTab('schema');
    setEditing(true);
    if (addParameter) {
      setDraft((current) => ({
        ...current,
        parameterSchema: current.parameterSchema.length ? current.parameterSchema : [emptySchemaParameter()],
      }));
    }
  }, []);

  const saveDraft = async () => {
    const message = validateDraft();
    if (message) {
      openErrNotification('无法保存模板', message);
      return;
    }
    try {
      if (draft.id) {
        await updateConfigTemplate(draft);
        await saveConfigTemplateLabels(draft.id, draft.labels || {});
      } else {
        await createConfigTemplate(draft);
        const result = await describeConfigTemplates();
        const created = result.templates.find((item) => item.name === draft.name);
        if (created) await saveConfigTemplateLabels(created.id, draft.labels || {});
      }
      openInfoNotification('保存成功', '模板草稿已保存');
      setEditing(false);
      await loadTemplates(draft.name);
    } catch (error) {
      openErrNotification('保存模板失败', toRequestErrorPayload(error));
    }
  };

  const openTemplatePublishDialog = () => {
    if (!draft.id) {
      openErrNotification('无法发布模板', '请先保存模板草稿');
      return;
    }
    if (editing) {
      openErrNotification('无法发布模板', '请先保存当前模板草稿');
      return;
    }
    const message = validateDraft();
    if (message) {
      openErrNotification('无法发布模板', message);
      return;
    }
    setTemplateReleaseComment('');
    setTemplatePublishOpen(true);
  };

  const publishTemplate = async () => {
    if (!draft.id) return;
    setTemplatePublishing(true);
    try {
      await publishConfigTemplateRelease({
        ...draft,
        comment: templateReleaseComment.trim() || draft.comment,
      });
      const result = await describeConfigTemplateReleases(draft.id);
      setReleases(result.releases);
      setTemplateReleaseId(result.releases[0]?.id || '');
      setPreviewTemplateReleaseId('');
      setValueDraftPreview(undefined);
      setVersionPreview(undefined);
      setTemplatePublishOpen(false);
      setTemplateReleaseComment('');
      openInfoNotification('发布成功', '已创建不可变 Template Release');
    } catch (error) {
      openErrNotification('发布模板失败', toRequestErrorPayload(error));
    } finally {
      setTemplatePublishing(false);
    }
  };

  const cancelEditing = async () => {
    setEditing(false);
    if (draft.id) await loadTemplates(undefined, String(draft.id));
    else setDraft(emptyTemplate());
  };

  const runVersionPreview = async () => {
    const selectedTemplateRelease = releases.find((release) => release.id === previewTemplateReleaseId);
    const selectedValueRelease = valueReleases.find((release) => release.id === previewValueReleaseId);
    if (!selectedTemplateRelease || !selectedValueRelease) {
      openErrNotification('无法预览版本组合', '请选择模板版本和环境 Value 版本');
      return;
    }
    try {
      const result = await previewConfigTemplate(selectedTemplateRelease, selectedValueRelease.values);
      setVersionPreview(result);
      if (result.valid) openInfoNotification('预览成功', '所选模板版本与 Value 版本渲染校验通过');
    } catch (error) {
      openErrNotification('版本组合预览失败', toRequestErrorPayload(error));
    }
  };

  const runValuePreview = async () => {
    if (!ensureValidValues()) return;
    const selectedTemplateRelease = releases.find((release) => release.id === templateReleaseId);
    if (!selectedTemplateRelease) {
      openErrNotification('无法预览 Value', '请先选择固定模板版本');
      return;
    }
    try {
      const result = await previewConfigTemplate(selectedTemplateRelease, values);
      setValueDraftPreview(result);
      if (result.valid) openInfoNotification('预览成功', '当前未保存 Value 的渲染与格式校验通过');
    } catch (error) {
      openErrNotification('Value 预览失败', toRequestErrorPayload(error));
    }
  };

  const saveValues = async () => {
    if (!draft.id || !namespace) return;
    if (!ensureValidValues()) return;
    try {
      await saveNamespaceTemplateValues({
        id: valuesId,
        namespace,
        templateId: draft.id,
        values,
        revision: valuesRevision,
      });
      const saved = await describeNamespaceTemplateValues(namespace, draft.id);
      setValuesId(saved?.id || '');
      setValuesRevision(saved?.revision || '');
      openInfoNotification('保存成功', `${namespace} 的 Value 草稿已保存`);
    } catch (error) {
      openErrNotification('保存 Namespace Value 失败', toRequestErrorPayload(error));
    }
  };

  const openValuePublishDialog = () => {
    if (!ensureValidValues()) return;
    if (!draft.id || !namespace || !templateReleaseId) {
      openErrNotification('无法发布 Value', '请先选择已发布的模板版本');
      return;
    }
    setValueReleaseType('TEMPLATE_VALUE_RELEASE_NORMAL');
    setValueComment('');
    setGrayRows([defaultGrayRuleRow()]);
    setValuePublishOpen(true);
  };

  const publishValues = async () => {
    if (!ensureValidValues()) return;
    if (!draft.id || !namespace || !templateReleaseId) return;
    const betaLabels = valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' ? grayRowsToBetaLabels(grayRows) : [];
    if (valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' && betaLabels.length === 0) {
      openErrNotification('无法发布灰度 Value', '至少需要一条有效灰度规则');
      return;
    }
    setValuePublishing(true);
    try {
      await publishNamespaceTemplateValueRelease({
        id: '',
        valuesId,
        namespace,
        templateId: draft.id,
        templateReleaseId,
        values,
        releaseType: valueReleaseType,
        betaLabels,
        priority: valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' ? 100 : 0,
        active: true,
        comment: valueComment,
      });
      const history = await describeNamespaceTemplateValueReleases(namespace, draft.id);
      setValueReleases(history.releases);
      setPreviewValueReleaseId('');
      setVersionPreview(undefined);
      setValuePublishOpen(false);
      setValueComment('');
      openInfoNotification('发布成功', `${namespace} 的 Value Release 已生效`);
    } catch (error) {
      openErrNotification('发布 Namespace Value 失败', toRequestErrorPayload(error));
    } finally {
      setValuePublishing(false);
    }
  };

  const templateColumns: PrimaryTableProps['columns'] = [
    {
      colKey: 'name',
      title: '模板名称',
      cell: ({ row }) => (
        <Button variant="text" className={styles.templateName} onClick={() => setSelectedId(String(row.id))}>
          {row.name}
        </Button>
      ),
    },
    { colKey: 'format', title: '格式', width: 92 },
    {
      colKey: 'parameters',
      title: '参数',
      width: 72,
      cell: ({ row }) => row.parameterSchema?.length || 0,
    },
  ];

  return (
    <div className={`${styles.page} ${embedded ? styles.embeddedPage : ''}`}>
      {!embedded && requestedGroup && (
        <GroupWorkspaceNav
          group={requestedGroup}
          onBack={() => navigate('/configuration/group')}
        />
      )}
      {!embedded && (
        <ResourceHeader
          density="compact"
          placement="app-header"
          eyebrow={requestedGroup ? `配置分组 / ${requestedGroup} / 配置模板` : '配置中心 / 配置模板'}
          title={requestedGroup ? '配置模板工作区' : '配置模板'}
          description={
            requestedGroup
              ? '先选择全局复用的模板，再按 Namespace 维护 Template Value；当前配置分组的文件可显式固定模板版本。'
              : '全局维护模板与参数 Schema，各 Namespace 独立维护 Value；配置文件显式固定不可变模板版本。'
          }
        />
      )}
      <div className={`${styles.workspace} ${embedded ? styles.embeddedWorkspace : ''}`}>
        {!embedded && <aside className={styles.catalog}>
          <ResourceToolbar
            className={styles.catalogToolbar}
            density="compact"
            title="模板目录"
            count={templates.length}
            description="全局逻辑模板"
            filters={
              <>
                {returnTo && (
                  <Button variant="outline" onClick={() => navigate(returnTo)}>
                    返回配置清单
                  </Button>
                )}
                <Tooltip content="刷新配置模板">
                  <Button
                    aria-label="刷新配置模板"
                    shape="square"
                    variant="outline"
                    icon={<RefreshIcon />}
                    onClick={() => loadTemplates()}
                  />
                </Tooltip>
                <Button
                  theme="primary"
                  icon={<AddIcon />}
                  onClick={() => {
                    setSelectedId('');
                    setDraft(emptyTemplate());
                    setEditing(true);
                    setActiveTab('basic');
                  }}
                >
                  新建配置模板
                </Button>
              </>
            }
          />
          <Table
            className={styles.catalogTable}
            data={templates}
            columns={templateColumns}
            rowKey="id"
            loading={loading}
            pagination={false}
            ariaLabel="配置模板目录"
          />
        </aside>}
        <main className={styles.detail}>
          <header className={styles.detailHeader}>
            <div className={styles.templateIdentity}>
              <div className={styles.titleLine}>
                <h3>{draft.name || '新建配置模板'}</h3>
                <Space size={6}>
                  <Tag variant="light" theme="primary">{draft.format || 'text'}</Tag>
                  <Tag variant="light">全局模板</Tag>
                  <Tag variant="light">pole-mustache-v1</Tag>
                  <Tag variant="light" theme={releases.length ? 'success' : 'default'}>
                    {releases.length ? `${releases.length} 个发布版本` : '未发布'}
                  </Tag>
                </Space>
              </div>
              <div className={styles.resourcePath} aria-label="配置模板资源路径">
                <span>{requestedGroup || '全局模板'}</span>
                <i>/</i>
                <strong>{draft.name || '新建配置模板'}</strong>
              </div>
            </div>
            {templateDefinitionTab && (
              <div className={styles.templateActions}>
                {!editing && draft.id ? (
                  <Button variant="outline" icon={<Edit1Icon />} onClick={() => setEditing(true)}>
                    编辑模板
                  </Button>
                ) : (
                  <>
                    <Button variant="outline" icon={<RollbackIcon />} onClick={cancelEditing}>撤销</Button>
                    <Button icon={<SaveIcon />} theme="primary" onClick={saveDraft}>保存模板草稿</Button>
                  </>
                )}
                <Button
                  icon={<RocketIcon />}
                  theme="primary"
                  disabled={!draft.id || editing}
                  onClick={openTemplatePublishDialog}
                >
                  发布模板版本
                </Button>
              </div>
            )}
          </header>
          <Tabs
            className={styles.detailTabs}
            value={activeTab}
            onChange={(value: string) => setActiveTab(value as WorkspaceTab)}
          >
            <TabPanel value="content" label="模板内容">
              <div className={styles.contentPane}>
                <section className={styles.templateEditorSection}>
                  <div className={styles.sectionHeading}>
                    <div>
                      <strong>模板内容</strong>
                      <span>
                        仅允许 triple-mustache dotted scalar，例如 <code>{'{{{database.host}}}'}</code>。
                      </span>
                    </div>
                  </div>
                  <div className={styles.templateEditor}>
                    <div className={styles.templateEditorBody}>
                      <CodeEditor
                        allowFullScreen
                        readonly={!editing}
                        height="100%"
                        language={draft.format}
                        value={draft.content}
                        onChange={(content) =>
                          setDraft((current) => ({
                            ...current,
                            content: content || '',
                          }))
                        }
                      />
                    </div>
                    <div className={styles.templateEditorStatusBar}>
                      <span>UTF-8</span>
                      <span>pole-mustache-v1</span>
                    </div>
                  </div>
                </section>
              </div>
            </TabPanel>
            <TabPanel value="schema" label={`参数 Schema · ${draft.parameterSchema.length}`}>
              <div className={styles.schemaPane}>
                <SchemaEditor
                  value={draft.parameterSchema}
                  editable={editing}
                  onChange={(parameterSchema) => setDraft((current) => ({ ...current, parameterSchema }))}
                  onStartEdit={() => startSchemaEditing(false)}
                  onAddFirst={() => startSchemaEditing(true)}
                />
              </div>
            </TabPanel>
            <TabPanel value="values" label="环境 Value">
              <div className={styles.valuesPane}>
                <section className={styles.valueToolbar}>
                  <label>
                    <span>环境空间</span>
                    <Select
                      value={namespace}
                      options={namespaces.map((value) => ({
                        label: value,
                        value,
                      }))}
                      onChange={(value: string) => {
                        setNamespace(value);
                        setPreviewValueReleaseId('');
                        setValueDraftPreview(undefined);
                        setVersionPreview(undefined);
                      }}
                      placeholder="选择环境空间"
                    />
                  </label>
                  <label>
                    <span>固定模板版本</span>
                    <Select
                      value={templateReleaseId}
                      options={releases.map((release) => ({
                        label: `v${release.version} · ${release.id}`,
                        value: release.id,
                      }))}
                      onChange={(value: string) => {
                        setTemplateReleaseId(value);
                        setValueDraftPreview(undefined);
                      }}
                      placeholder="先发布模板版本"
                    />
                  </label>
                  <div className={styles.valueActions}>
                    <Button
                      variant="outline"
                      onClick={saveValues}
                      disabled={!namespace || !draft.id || hasValueErrors}
                    >
                      保存 Value 草稿
                    </Button>
                    <Button
                      variant="outline"
                      onClick={runValuePreview}
                      disabled={!templateReleaseId || hasValueErrors}
                    >
                      预览渲染
                    </Button>
                    <Button
                      theme="primary"
                      icon={<RocketIcon />}
                      onClick={openValuePublishDialog}
                      disabled={!namespace || !draft.id || !templateReleaseId || hasValueErrors}
                    >
                      发布 Value 版本
                    </Button>
                  </div>
                </section>
                <ValueEditor
                  schema={draft.parameterSchema}
                  values={values}
                  onChange={(nextValues) => {
                    setValues(nextValues);
                    setValueDraftPreview(undefined);
                  }}
                  onCreateSchema={() => startSchemaEditing(true)}
                />
                {valueDraftPreview && (
                  <RenderPreviewPanel
                    preview={valueDraftPreview}
                    title="当前 Value 渲染结果"
                    emptyMessage="填写 Value 后预览最终配置。"
                    compact
                  />
                )}
              </div>
            </TabPanel>
            <TabPanel value="versions" label={`版本管理 · ${releases.length + valueReleases.length}`}>
              <VersionManager
                namespace={namespace}
                releases={releases}
                valueReleases={valueReleases}
                templateReleaseId={previewTemplateReleaseId}
                valueReleaseId={previewValueReleaseId}
                preview={versionPreview}
                onTemplateReleaseChange={(value) => {
                  setPreviewTemplateReleaseId(value);
                  setVersionPreview(undefined);
                }}
                onValueReleaseChange={(value) => {
                  setPreviewValueReleaseId(value);
                  setVersionPreview(undefined);
                }}
                onPreview={runVersionPreview}
              />
            </TabPanel>
            <TabPanel value="basic" label="基本信息">
              <div className={styles.basicPane}>
                <section className={styles.metadataLayer} aria-label="模板基本信息">
                  <div className={styles.metadataLayerHeading}>
                    <div>
                      <strong>基本信息</strong>
                      <span>管理模板名称、目标格式、说明和用于目录查询的标签。</span>
                    </div>
                  </div>
                  {editing ? (
                    <div className={styles.metadataEditGrid}>
                      <label>
                        <span>模板名称</span>
                        <Input
                          value={draft.name}
                          disabled={Boolean(draft.id)}
                          placeholder="application-config"
                          onChange={(name) => setDraft((current) => ({ ...current, name }))}
                        />
                      </label>
                      <label>
                        <span>目标格式</span>
                        <Select
                          value={draft.format}
                          options={['text', 'yaml', 'json', 'toml'].map((value) => ({
                            label: value.toUpperCase(),
                            value,
                          }))}
                          onChange={(format: string) => setDraft((current) => ({ ...current, format }))}
                        />
                      </label>
                      <label className={styles.commentField}>
                        <span>模板说明</span>
                        <Input
                          value={draft.comment || ''}
                          placeholder="简要说明模板用途"
                          onChange={(comment) => setDraft((current) => ({ ...current, comment }))}
                        />
                      </label>
                      <label className={styles.templateLabels}>
                        <span>模板标签</span>
                        <TagInput
                          value={Object.entries(draft.labels || {}).map(([key, value]) => `${key}=${value}`)}
                          placeholder="输入 key=value，回车添加"
                          onChange={(items: string[]) =>
                            setDraft((current) => ({
                              ...current,
                              labels: Object.fromEntries(
                                items
                                  .map((item) => {
                                    const separator = item.indexOf('=');
                                    return separator > 0
                                      ? [item.slice(0, separator).trim(), item.slice(separator + 1).trim()]
                                      : [item.trim(), ''];
                                  })
                                  .filter(([key]) => key)
                              ),
                            }))
                          }
                        />
                      </label>
                    </div>
                  ) : (
                    <div className={styles.metadataInfoGrid}>
                      <div className={styles.metadataField}>
                        <span>模板名称</span>
                        <strong>{draft.name || '-'}</strong>
                      </div>
                      <div className={styles.metadataField}>
                        <span>目标格式</span>
                        <strong>{(draft.format || 'text').toUpperCase()}</strong>
                      </div>
                      <div className={`${styles.metadataField} ${styles.metadataDescription}`}>
                        <span>模板说明</span>
                        <strong>{draft.comment || '暂无说明'}</strong>
                      </div>
                      <div className={`${styles.metadataField} ${styles.metadataTags}`}>
                        <span>模板标签</span>
                        <div className={styles.metadataTagList}>
                          {Object.entries(draft.labels || {}).map(([key, value]) => (
                            <Tag key={key} variant="light">{key}={value}</Tag>
                          ))}
                          {Object.keys(draft.labels || {}).length === 0 && <em>暂无标签</em>}
                        </div>
                      </div>
                    </div>
                  )}
                </section>
              </div>
            </TabPanel>
          </Tabs>
          <Dialog
            visible={templatePublishOpen}
            header="发布模板版本"
            width={640}
            onClose={() => {
              if (!templatePublishing) setTemplatePublishOpen(false);
            }}
            footer={(
              <>
                <Button variant="outline" disabled={templatePublishing} onClick={() => setTemplatePublishOpen(false)}>
                  取消
                </Button>
                <Button theme="primary" icon={<RocketIcon />} loading={templatePublishing} onClick={publishTemplate}>
                  发布 v{nextTemplateVersion}
                </Button>
              </>
            )}
          >
            <div className={styles.publishDialog}>
              <div className={styles.publishObjectSummary}>
                <div>
                  <span>发布对象</span>
                  <strong>模板版本 v{nextTemplateVersion}</strong>
                </div>
                <p>发布后生成不可变模板版本；环境 Value 仍需在对应环境中独立发布。</p>
              </div>
              <div className={styles.publishSummaryGrid}>
                <div><span>模板名称</span><strong>{draft.name || '-'}</strong></div>
                <div><span>目标格式</span><strong>{(draft.format || 'text').toUpperCase()}</strong></div>
                <div><span>Schema 参数</span><strong>{draft.parameterSchema.length}</strong></div>
                <div><span>模板引擎</span><strong>pole-mustache-v1</strong></div>
              </div>
              <label className={styles.publishField}>
                <span>本次发布说明 <small>可选</small></span>
                <Textarea
                  value={templateReleaseComment}
                  rows={3}
                  autosize={{ minRows: 3, maxRows: 6 }}
                  placeholder="说明本次模板内容、Schema 或元信息变更"
                  onChange={setTemplateReleaseComment}
                />
              </label>
            </div>
          </Dialog>
          <Dialog
            visible={valuePublishOpen}
            header="发布环境 Value 版本"
            width={920}
            onClose={() => {
              if (!valuePublishing) setValuePublishOpen(false);
            }}
            footer={(
              <>
                <Button variant="outline" disabled={valuePublishing} onClick={() => setValuePublishOpen(false)}>
                  取消
                </Button>
                <Button theme="primary" icon={<RocketIcon />} loading={valuePublishing} onClick={publishValues}>
                  发布 Value 版本
                </Button>
              </>
            )}
          >
            <div className={styles.publishDialog}>
              <div className={styles.publishObjectSummary}>
                <div>
                  <span>发布对象</span>
                  <strong>{namespace || '-'} 环境 Value</strong>
                </div>
                <p>只发布当前环境的 Value 快照，不会修改模板内容、Schema 或其他环境的 Value。</p>
              </div>
              <div className={styles.publishSummaryGrid}>
                <div><span>环境空间</span><strong>{namespace || '-'}</strong></div>
                <div>
                  <span>固定模板版本</span>
                  <strong>
                    {releases.find((release) => release.id === templateReleaseId)?.version
                      ? `v${releases.find((release) => release.id === templateReleaseId)?.version}`
                      : '-'}
                  </strong>
                </div>
                <div><span>Value 参数</span><strong>{draft.parameterSchema.length}</strong></div>
                <div><span>发布范围</span><strong>{valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' ? '灰度' : '全量'}</strong></div>
              </div>
              <div className={styles.publishTypeField}>
                <span>发布类型</span>
                <RadioGroup value={valueReleaseType} onChange={setValueReleaseType}>
                  <Radio value="TEMPLATE_VALUE_RELEASE_NORMAL">全量发布</Radio>
                  <Radio value="TEMPLATE_VALUE_RELEASE_GRAY">灰度发布</Radio>
                </RadioGroup>
              </div>
              {valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' && (
                <div className={styles.grayReleaseRules}>
                  <GrayRuleEditor rows={grayRows} editable onChange={setGrayRows} />
                </div>
              )}
              <label className={styles.publishField}>
                <span>发布说明 <small>可选</small></span>
                <Textarea
                  value={valueComment}
                  rows={3}
                  autosize={{ minRows: 3, maxRows: 6 }}
                  placeholder="说明本次 Value 变更与发布目的"
                  onChange={setValueComment}
                />
              </label>
            </div>
          </Dialog>
        </main>
      </div>
    </div>
  );
};

export default React.memo(TemplateWorkspace);
