import React from 'react';
import {
  Button,
  Input,
  PrimaryTableProps,
  Radio,
  RadioGroup,
  Select,
  Space,
  Table,
  Tag,
  Tabs,
  Textarea,
  Tooltip,
} from 'components/Fluent';
import { AddIcon, RefreshIcon, RocketIcon, SaveIcon } from 'components/Fluent/icons';
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
  TemplateEngine,
  TemplateValueReleaseType,
  updateConfigTemplate,
} from 'services/config_templates';
import GrayRuleEditor, {
  defaultGrayRuleRow,
  grayRowsToBetaLabels,
} from '../Group/Releases/GrayRuleEditor';
import { TrafficMatchConditionRow } from 'pages/Governance/shared/TrafficMatchConditionEditor';
import SchemaEditor from './SchemaEditor';
import ValueEditor, { textToValue } from './ValueEditor';
import styles from './index.module.less';
import GroupWorkspaceNav from '../Group/GroupWorkspaceNav';

const { TabPanel } = Tabs;

type WorkspaceTab = 'definition' | 'values' | 'releases' | 'preview';

const emptyTemplate = (): ConfigFileTemplate => ({
  id: 0,
  name: '',
  content: '',
  comment: '',
  format: 'text',
  engine: TemplateEngine,
  parameterSchema: [],
});

const initialValues = (template: ConfigFileTemplate) => Object.fromEntries(
  template.parameterSchema.map(parameter => [
    parameter.name,
    parameter.defaultValue || textToValue(
      parameter.type,
      parameter.type === 'TEMPLATE_PARAMETER_BOOLEAN' ? 'false'
        : parameter.type === 'TEMPLATE_PARAMETER_INTEGER' || parameter.type === 'TEMPLATE_PARAMETER_DECIMAL'
          ? '0'
          : '',
    ),
  ]),
);

const TemplateWorkspace: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const requestedTemplateId = searchParams.get('templateId') || '';
  const requestedNamespace = searchParams.get('namespace') || '';
  const requestedGroup = searchParams.get('group') || '';
  const requestedTab = searchParams.get('tab') as WorkspaceTab | null;
  const returnTo = searchParams.get('returnTo') || '';
  const [templates, setTemplates] = React.useState<ConfigFileTemplate[]>([]);
  const [selectedId, setSelectedId] = React.useState<string>('');
  const [draft, setDraft] = React.useState<ConfigFileTemplate>(emptyTemplate());
  const [editing, setEditing] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [activeTab, setActiveTab] = React.useState<WorkspaceTab>('definition');
  const [releases, setReleases] = React.useState<ConfigTemplateRelease[]>([]);
  const [templateReleaseId, setTemplateReleaseId] = React.useState('');
  const [namespaces, setNamespaces] = React.useState<string[]>([]);
  const [namespace, setNamespace] = React.useState('');
  const [valuesId, setValuesId] = React.useState('');
  const [valuesRevision, setValuesRevision] = React.useState('');
  const [values, setValues] = React.useState<Record<string, ConfigTemplateValue>>({});
  const [valueReleases, setValueReleases] = React.useState<NamespaceTemplateValueRelease[]>([]);
  const [valueReleaseType, setValueReleaseType] = React.useState<TemplateValueReleaseType>(
    'TEMPLATE_VALUE_RELEASE_NORMAL',
  );
  const [valueComment, setValueComment] = React.useState('');
  const [grayRows, setGrayRows] = React.useState<TrafficMatchConditionRow[]>([defaultGrayRuleRow()]);
  const [preview, setPreview] = React.useState<RenderPreview>();

  const selected = React.useMemo(
    () => templates.find(item => String(item.id) === selectedId),
    [selectedId, templates],
  );

  const loadTemplates = React.useCallback(async (preferredName?: string, preferredId?: string) => {
    setLoading(true);
    try {
      const result = await describeConfigTemplates();
      setTemplates(result.templates);
      const preferred = preferredId
        ? result.templates.find(item => String(item.id) === preferredId)
        : preferredName
        ? result.templates.find(item => item.name === preferredName)
        : result.templates.find(item => String(item.id) === selectedId) || result.templates[0];
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
  }, [selectedId]);

  React.useEffect(() => {
    loadTemplates(undefined, requestedTemplateId);
    describeBusinessNamespaces()
      .then(items => {
        const names = items.map(item => item.name);
        setNamespaces(names);
        setNamespace(current => requestedNamespace || current || names[0] || '');
      })
      .catch(error => openErrNotification('加载命名空间失败', toRequestErrorPayload(error)));
    // 初始化只执行一次，后续刷新由显式操作触发。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  React.useEffect(() => {
    if (!selected) return;
    setDraft({
      ...selected,
      engine: selected.engine || TemplateEngine,
      parameterSchema: selected.parameterSchema.map(item => ({ ...item })),
    });
    setEditing(false);
    setPreview(undefined);
    setActiveTab(requestedTab && ['definition', 'values', 'releases', 'preview'].includes(requestedTab)
      ? requestedTab
      : 'definition');
  }, [requestedTab, selected]);

  React.useEffect(() => {
    if (!draft.id) {
      setReleases([]);
      setTemplateReleaseId('');
      return;
    }
    describeConfigTemplateReleases(draft.id)
      .then(result => {
        setReleases(result.releases);
        setTemplateReleaseId(current => (
          result.releases.some(item => item.id === current) ? current : result.releases[0]?.id || ''
        ));
      })
      .catch(error => openErrNotification('加载模板发布版本失败', toRequestErrorPayload(error)));
  }, [draft.id]);

  React.useEffect(() => {
    if (!draft.id || !namespace) {
      setValues(initialValues(draft));
      setValueReleases([]);
      return;
    }
    Promise.all([
      describeNamespaceTemplateValues(namespace, draft.id),
      describeNamespaceTemplateValueReleases(namespace, draft.id),
    ]).then(([saved, history]) => {
      setValuesId(saved?.id || '');
      setValuesRevision(saved?.revision || '');
      setValues(saved?.values && Object.keys(saved.values).length ? saved.values : initialValues(draft));
      setValueReleases(history.releases);
    }).catch(error => openErrNotification('加载 Namespace Value 失败', toRequestErrorPayload(error)));
  }, [draft.id, namespace]);

  const validateDraft = () => {
    if (!draft.name.trim()) return '模板名称不能为空';
    if (!draft.content) return '模板内容不能为空';
    const names = draft.parameterSchema.map(item => item.name.trim());
    if (names.some(name => !name)) return 'Schema 参数名不能为空';
    if (new Set(names).size !== names.length) return 'Schema 参数名不能重复';
    return '';
  };

  const saveDraft = async () => {
    const message = validateDraft();
    if (message) {
      openErrNotification('无法保存模板', message);
      return;
    }
    try {
      if (draft.id) await updateConfigTemplate(draft);
      else await createConfigTemplate(draft);
      openInfoNotification('保存成功', '模板草稿已保存');
      setEditing(false);
      await loadTemplates(draft.name);
    } catch (error) {
      openErrNotification('保存模板失败', toRequestErrorPayload(error));
    }
  };

  const publishTemplate = async () => {
    if (!draft.id) {
      openErrNotification('无法发布模板', '请先保存模板草稿');
      return;
    }
    const message = validateDraft();
    if (message) {
      openErrNotification('无法发布模板', message);
      return;
    }
    try {
      await publishConfigTemplateRelease(draft);
      const result = await describeConfigTemplateReleases(draft.id);
      setReleases(result.releases);
      setTemplateReleaseId(result.releases[0]?.id || '');
      openInfoNotification('发布成功', '已创建不可变 Template Release');
    } catch (error) {
      openErrNotification('发布模板失败', toRequestErrorPayload(error));
    }
  };

  const runPreview = async () => {
    try {
      const result = await previewConfigTemplate(draft, values);
      setPreview(result);
      setActiveTab('preview');
      if (result.valid) openInfoNotification('预览成功', '服务端参考渲染与格式校验通过');
    } catch (error) {
      // API code/info 错误由统一请求错误处理；渲染 diagnostics 在预览面板内展示。
      openErrNotification('预览请求失败', toRequestErrorPayload(error));
    }
  };

  const saveValues = async () => {
    if (!draft.id || !namespace) return;
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

  const publishValues = async () => {
    if (!draft.id || !namespace || !templateReleaseId) {
      openErrNotification('无法发布 Value', '请先选择已发布的模板版本');
      return;
    }
    const betaLabels = valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY'
      ? grayRowsToBetaLabels(grayRows)
      : [];
    if (valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' && betaLabels.length === 0) {
      openErrNotification('无法发布灰度 Value', '至少需要一条有效灰度规则');
      return;
    }
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
      openInfoNotification('发布成功', `${namespace} 的 Value Release 已生效`);
    } catch (error) {
      openErrNotification('发布 Namespace Value 失败', toRequestErrorPayload(error));
    }
  };

  const templateColumns: PrimaryTableProps['columns'] = [
    {
      colKey: 'name',
      title: '模板名称',
      cell: ({ row }) => (
        <Button
          variant="text"
          className={styles.templateName}
          onClick={() => setSelectedId(String(row.id))}
        >
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
    {
      colKey: 'engine',
      title: '引擎',
      width: 138,
      cell: () => <Tag variant="light">pole-mustache-v1</Tag>,
    },
  ];

  const releaseColumns: PrimaryTableProps['columns'] = [
    { colKey: 'version', title: '版本', width: 90 },
    { colKey: 'id', title: 'Release ID' },
    { colKey: 'format', title: '格式', width: 90 },
    { colKey: 'contentSha256', title: '内容 SHA-256' },
  ];

  const valueReleaseColumns: PrimaryTableProps['columns'] = [
    { colKey: 'version', title: '版本', width: 82 },
    {
      colKey: 'releaseType',
      title: '类型',
      width: 100,
      cell: ({ row }) => row.releaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' ? '灰度' : '全量',
    },
    { colKey: 'templateReleaseId', title: 'Template Release' },
    { colKey: 'id', title: 'Value Release' },
    {
      colKey: 'active',
      title: '状态',
      width: 90,
      cell: ({ row }) => <Tag theme={row.active ? 'success' : 'default'}>{row.active ? '生效' : '停止'}</Tag>,
    },
  ];

  const previewPanel = (
    <div className={styles.previewLayout}>
      <section className={styles.previewOutput}>
        <div className={styles.sectionHeading}>
          <div>
            <strong>服务端参考预览</strong>
            <span>只用于预览和一致性校验；SDK 仍会在本地渲染运行时结果。</span>
          </div>
          {preview && <Tag theme={preview.valid ? 'success' : 'danger'}>{preview.valid ? '校验通过' : '校验失败'}</Tag>}
        </div>
        <div className={styles.previewEditor}>
          <CodeEditor
            readonly
            allowFullScreen
            height="100%"
            language={preview?.format || draft.format}
            value={preview?.renderedContent || ''}
          />
        </div>
        <div className={styles.hashLine}>
          <span>rendered SHA-256</span>
          <code>{preview?.renderedSha256 || '-'}</code>
        </div>
      </section>
      <aside className={styles.diagnosticPanel}>
        <strong>渲染诊断 · {preview?.diagnostics.length || 0}</strong>
        {(preview?.diagnostics || []).map((diagnostic, index) => (
          <div className={styles.diagnosticItem} key={`${diagnostic.code}-${index}`}>
            <Tag theme={diagnostic.severity === 'DIAGNOSTIC_ERROR' ? 'danger' : 'warning'}>
              {diagnostic.code}
            </Tag>
            <p>{diagnostic.message}</p>
            {diagnostic.parameter && <code>{diagnostic.parameter}</code>}
          </div>
        ))}
        {!preview?.diagnostics.length && <span className={styles.muted}>暂无 diagnostics</span>}
      </aside>
    </div>
  );

  return (
    <div className={styles.page}>
      {requestedGroup && (
        <GroupWorkspaceNav
          active="templates"
          group={requestedGroup}
          filesHref={
            returnTo || `/configuration/group/files?namespace=${encodeURIComponent(namespace || requestedNamespace)}`
              + `&group=${encodeURIComponent(requestedGroup)}`
          }
          templatesHref={window.location.pathname + window.location.search}
          onBack={() => navigate('/configuration/group')}
        />
      )}
      <ResourceHeader
        density="compact"
        placement="app-header"
        eyebrow={requestedGroup ? `配置分组 / ${requestedGroup} / 配置模板` : '配置中心 / 配置模板'}
        title={requestedGroup ? '配置模板工作区' : '配置模板'}
        description={requestedGroup
          ? '先选择全局复用的模板，再按 Namespace 维护 Template Value；当前配置分组的文件可显式固定模板版本。'
          : '全局维护模板与参数 Schema，各 Namespace 独立维护 Value；配置文件显式固定不可变模板版本。'}
      />
      <div className={styles.workspace}>
        <aside className={styles.catalog}>
          <ResourceToolbar
            density="compact"
            title="模板目录"
            count={templates.length}
            description="全局逻辑模板"
            filters={(
              <>
                {returnTo && !requestedGroup && (
                  <Button variant="outline" onClick={() => navigate(returnTo)}>返回配置分组</Button>
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
                    setActiveTab('definition');
                  }}
                >
                  新建配置模板
                </Button>
              </>
            )}
          />
          <Table
            data={templates}
            columns={templateColumns}
            rowKey="id"
            loading={loading}
            pagination={false}
            ariaLabel="配置模板目录"
          />
        </aside>
        <main className={styles.detail}>
          <header className={styles.detailHeader}>
            <div>
              <div className={styles.titleLine}>
                <h3>{draft.name || '新建配置模板'}</h3>
                <Tag theme="primary">pole-mustache-v1</Tag>
                <Tag>{draft.format || 'text'}</Tag>
              </div>
              <p>{draft.comment || '声明稳定模板与类型化参数，由 Namespace Value 提供运行时输入。'}</p>
            </div>
            <Space>
              {!editing && draft.id ? (
                <Button variant="outline" onClick={() => setEditing(true)}>编辑草稿</Button>
              ) : (
                <Button icon={<SaveIcon />} theme="primary" onClick={saveDraft}>保存草稿</Button>
              )}
              <Button variant="outline" onClick={runPreview}>参考预览</Button>
              <Button icon={<RocketIcon />} theme="primary" disabled={!draft.id} onClick={publishTemplate}>
                发布模板
              </Button>
            </Space>
          </header>
          <Tabs value={activeTab} onChange={(value: string) => setActiveTab(value as WorkspaceTab)}>
            <TabPanel value="definition" label="模板定义">
              <div className={styles.definitionPane}>
                <section className={styles.metadataGrid}>
                  <label>
                    <span>模板名称</span>
                    <Input
                      value={draft.name}
                      disabled={!editing || Boolean(draft.id)}
                      placeholder="application-config"
                      onChange={name => setDraft(current => ({ ...current, name }))}
                    />
                  </label>
                  <label>
                    <span>目标格式</span>
                    <Select
                      value={draft.format}
                      disabled={!editing}
                      options={['text', 'yaml', 'json', 'toml'].map(value => ({ label: value.toUpperCase(), value }))}
                      onChange={(format: string) => setDraft(current => ({ ...current, format }))}
                    />
                  </label>
                  <label className={styles.commentField}>
                    <span>模板说明</span>
                    <Textarea
                      value={draft.comment || ''}
                      disabled={!editing}
                      rows={2}
                      onChange={comment => setDraft(current => ({ ...current, comment }))}
                    />
                  </label>
                </section>
                <section className={styles.templateEditorSection}>
                  <div className={styles.sectionHeading}>
                    <div>
                      <strong>模板内容</strong>
                      <span>仅允许 triple-mustache dotted scalar，例如 <code>{'{{{database.host}}}'}</code>。</span>
                    </div>
                  </div>
                  <div className={styles.templateEditor}>
                    <CodeEditor
                      allowFullScreen
                      readonly={!editing}
                      height="100%"
                      language={draft.format}
                      value={draft.content}
                      onChange={content => setDraft(current => ({ ...current, content: content || '' }))}
                    />
                  </div>
                </section>
                <SchemaEditor
                  value={draft.parameterSchema}
                  editable={editing}
                  onChange={parameterSchema => setDraft(current => ({ ...current, parameterSchema }))}
                />
              </div>
            </TabPanel>
            <TabPanel value="values" label="Namespace Value">
              <div className={styles.valuesPane}>
                <section className={styles.valueToolbar}>
                  <label>
                    <span>Namespace</span>
                    <Select
                      value={namespace}
                      options={namespaces.map(value => ({ label: value, value }))}
                      onChange={setNamespace}
                      placeholder="选择命名空间"
                    />
                  </label>
                  <label>
                    <span>固定模板版本</span>
                    <Select
                      value={templateReleaseId}
                      options={releases.map(release => ({
                        label: `v${release.version} · ${release.id}`,
                        value: release.id,
                      }))}
                      onChange={setTemplateReleaseId}
                      placeholder="先发布模板版本"
                    />
                  </label>
                  <div className={styles.valueActions}>
                    <Button variant="outline" onClick={saveValues} disabled={!namespace || !draft.id}>保存 Value 草稿</Button>
                    <Button variant="outline" onClick={runPreview}>预览</Button>
                  </div>
                </section>
                <ValueEditor schema={draft.parameterSchema} values={values} onChange={setValues} />
                <section className={styles.releaseComposer}>
                  <div className={styles.sectionHeading}>
                    <div>
                      <strong>发布 Namespace Value</strong>
                      <span>服务端负责灰度匹配，SDK 只接收命中的 Value 快照。</span>
                    </div>
                  </div>
                  <RadioGroup value={valueReleaseType} onChange={setValueReleaseType}>
                    <Radio value="TEMPLATE_VALUE_RELEASE_NORMAL">全量发布</Radio>
                    <Radio value="TEMPLATE_VALUE_RELEASE_GRAY">灰度发布</Radio>
                  </RadioGroup>
                  {valueReleaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' && (
                    <GrayRuleEditor rows={grayRows} editable onChange={setGrayRows} />
                  )}
                  <div className={styles.releaseFooter}>
                    <Input value={valueComment} placeholder="发布说明" onChange={setValueComment} />
                    <Button theme="primary" icon={<RocketIcon />} onClick={publishValues}>发布 Value</Button>
                  </div>
                </section>
                <Table data={valueReleases} columns={valueReleaseColumns} rowKey="id" pagination={false} />
              </div>
            </TabPanel>
            <TabPanel value="releases" label="模板版本">
              <div className={styles.tablePane}>
                <Table data={releases} columns={releaseColumns} rowKey="id" pagination={false} />
              </div>
            </TabPanel>
            <TabPanel value="preview" label="参考预览">{previewPanel}</TabPanel>
          </Tabs>
        </main>
      </div>
    </div>
  );
};

export default React.memo(TemplateWorkspace);
