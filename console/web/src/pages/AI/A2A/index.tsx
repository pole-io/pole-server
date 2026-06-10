import React, { memo, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Col,
  Descriptions,
  Drawer,
  Form,
  Input,
  Link,
  Popconfirm,
  PrimaryTableProps,
  Row,
  Select,
  Space,
  Switch,
  Table,
  TableRowData,
  Tabs,
  Tag,
  TagInput,
  Textarea,
  Tooltip,
} from 'tdesign-react';
import type { FormProps, PageInfo } from 'tdesign-react';
import {
  AddIcon,
  DeleteIcon,
  EditIcon,
  FileIcon,
  ListIcon,
  RefreshIcon,
} from 'tdesign-icons-react';

import Text from 'components/Text';
import {
  cleanA2ADetails,
  cleanA2APage,
  editorA2AAgent,
  getA2AAgentCard,
  listA2AAgents,
  listA2AAgentSkills,
  removeA2AAgents,
  resetA2AAgent,
  saveA2AAgent,
  selectA2A,
  updateA2AAgent,
} from 'modules/ai/a2a';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { A2AAgent, A2AAgentInterface, A2AAgentSkill } from 'services/a2a';
import { Op } from 'services/types';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';

const { FormItem } = Form;
const { DescriptionsItem } = Descriptions;
const { TabPanel } = Tabs;

const protocolOptions = [
  { label: 'JSON-RPC', value: 'jsonrpc' },
  { label: 'HTTP+JSON', value: 'http-json' },
  { label: 'gRPC', value: 'grpc' },
];

const backendOptions = [
  { label: 'Pole 服务', value: 'service' },
  { label: 'URL', value: 'url' },
  { label: '外部系统', value: 'external' },
];

const fetchStatusOptions = [
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' },
  { label: '未拉取', value: 'pending' },
];

const capabilityFilterOptions = [
  { label: '支持', value: 'true' },
  { label: '不支持', value: 'false' },
];

const defaultInterface = (): A2AAgentInterface => ({
  url: '',
  protocol_binding: 'jsonrpc',
  protocol_version: '0.3.0',
  tenant: '',
});

const defaultSkill = (): A2AAgentSkill => ({
  skill_id: '',
  name: '',
  description: '',
  tags: [],
  examples: [],
  input_modes: ['text/plain'],
  output_modes: ['text/plain'],
  security_requirements_json: '',
});

function tagList(value?: string[]) {
  return Array.isArray(value) ? value.filter(Boolean) : [];
}

function jsonPreview(value: unknown) {
  if (!value) return '';
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}

function parseMetadata(raw?: string): Record<string, string> | undefined {
  if (!raw) return undefined;
  return raw.split('\n').reduce<Record<string, string>>((acc, line) => {
    const index = line.indexOf('=');
    if (index <= 0) return acc;
    const key = line.slice(0, index).trim();
    const value = line.slice(index + 1).trim();
    if (key) acc[key] = value;
    return acc;
  }, {});
}

function stringifyMetadata(metadata?: Record<string, string>) {
  return metadata ? Object.entries(metadata).map(([key, value]) => `${key}=${value}`).join('\n') : '';
}

function normalizeInterfaces(items: A2AAgentInterface[]) {
  return items
    .filter((item) => item.url || item.protocol_binding || item.protocol_version || item.tenant)
    .map((item) => ({
      ...item,
      protocol_binding: item.protocol_binding || 'jsonrpc',
      protocol_version: item.protocol_version || '0.3.0',
    }));
}

function normalizeSkills(items: A2AAgentSkill[]) {
  return items
    .filter((item) => item.name || item.skill_id)
    .map((item) => ({
      ...item,
      skill_id: item.skill_id || item.name,
      tags: tagList(item.tags),
      examples: tagList(item.examples),
      input_modes: tagList(item.input_modes),
      output_modes: tagList(item.output_modes),
    }));
}

const capabilityTags = (agent: A2AAgent) => (
  <Space size="small" breakLine>
    {agent.streaming && <Tag theme="success" variant="light-outline">Streaming</Tag>}
    {agent.push_notifications && <Tag theme="warning" variant="light-outline">Push</Tag>}
    {agent.extended_agent_card && <Tag theme="primary" variant="light-outline">Extended Card</Tag>}
    {!agent.streaming && !agent.push_notifications && !agent.extended_agent_card && <Text>-</Text>}
  </Space>
);

const agentColumns = (
  operateAgent: (op: Op | 'skills' | 'card', row?: TableRowData) => void,
): PrimaryTableProps['columns'] => [
  {
    colKey: 'name',
    title: '名称',
    fixed: 'left',
    cell: ({ row }) => (
      <Link theme="primary" onClick={() => operateAgent('skills', row)}>
        {row.name}
      </Link>
    ),
  },
  {
    colKey: 'namespace',
    title: '命名空间',
    cell: ({ row }) => <Text>{row.namespace}</Text>,
  },
  {
    colKey: 'version',
    title: '版本',
    cell: ({ row }) => <Text>{row.version || '-'}</Text>,
  },
  {
    colKey: 'protocol',
    title: '协议',
    cell: ({ row }) => <Tag variant="outline">{row.preferred_protocol_binding || '-'}</Tag>,
  },
  {
    colKey: 'capabilities',
    title: '能力',
    cell: ({ row }) => capabilityTags(row as A2AAgent),
  },
  {
    colKey: 'skills',
    title: '技能',
    cell: ({ row }) => <Text>{Array.isArray(row.skills) ? row.skills.length : 0}</Text>,
  },
  {
    colKey: 'backend',
    title: '后端绑定',
    ellipsis: true,
    cell: ({ row }) => (
      <Text>
        {row.backend_type || '-'}
        {row.backend_service_name ? ` / ${row.backend_service_namespace || 'default'}/${row.backend_service_name}` : ''}
        {row.backend_address ? ` / ${row.backend_address}` : ''}
      </Text>
    ),
  },
  {
    colKey: 'last_fetch_status',
    title: '拉取状态',
    cell: ({ row }) => <Tag variant="light-outline">{row.last_fetch_status || 'pending'}</Tag>,
  },
  {
    colKey: 'time',
    title: '操作时间',
    cell: ({ row }) => <Text>修改: {row.mtime || '-'}<br />创建: {row.ctime || '-'}</Text>,
  },
  {
    colKey: 'action',
    title: '操作',
    fixed: 'right',
    cell: ({ row }) => (
      <Space>
        <Tooltip content="查看 Agent Card">
          <Button shape="square" variant="text" onClick={() => operateAgent('card', row)}>
            <FileIcon />
          </Button>
        </Tooltip>
        <Tooltip content="查看技能">
          <Button shape="square" variant="text" onClick={() => operateAgent('skills', row)}>
            <ListIcon />
          </Button>
        </Tooltip>
        <Tooltip content="编辑">
          <Button shape="square" variant="text" onClick={() => operateAgent('edit', row)}>
            <EditIcon />
          </Button>
        </Tooltip>
        <Tooltip content="删除">
          <Popconfirm
            content="确认删除该 A2A Agent 吗"
            destroyOnClose
            placement="top"
            showArrow
            theme="default"
            onConfirm={() => operateAgent('delete', row)}
          >
            <Button shape="square" variant="text">
              <DeleteIcon />
            </Button>
          </Popconfirm>
        </Tooltip>
      </Space>
    ),
  },
];

const skillColumns: PrimaryTableProps['columns'] = [
  {
    colKey: 'name',
    title: '技能名',
    fixed: 'left',
    cell: ({ row }) => <Text>{row.name}</Text>,
  },
  {
    colKey: 'skill_id',
    title: 'Skill ID',
    cell: ({ row }) => <Text>{row.skill_id || '-'}</Text>,
  },
  {
    colKey: 'tags',
    title: '标签',
    cell: ({ row }) => (
      <Space size="small" breakLine>
        {tagList(row.tags).map((tag) => <Tag key={tag} variant="light-outline">{tag}</Tag>)}
      </Space>
    ),
  },
  {
    colKey: 'input_modes',
    title: '输入',
    cell: ({ row }) => <Text>{tagList(row.input_modes).join(', ') || '-'}</Text>,
  },
  {
    colKey: 'output_modes',
    title: '输出',
    cell: ({ row }) => <Text>{tagList(row.output_modes).join(', ') || '-'}</Text>,
  },
  {
    colKey: 'description',
    title: '描述',
    ellipsis: true,
    cell: ({ row }) => <Text>{row.description || '-'}</Text>,
  },
];

const A2AEditor: React.FC<{
  op: Op;
  visible: boolean;
  closeDrawer: () => void;
}> = ({ op, visible, closeDrawer }) => {
  const [form] = Form.useForm();
  const dispatch = useAppDispatch();
  const { editAgent } = useAppSelector(selectA2A);
  const [interfaces, setInterfaces] = useState<A2AAgentInterface[]>([defaultInterface()]);
  const [skills, setSkills] = useState<A2AAgentSkill[]>([defaultSkill()]);

  useEffect(() => {
    if (!visible) return;
    form.setFieldsValue({
      name: editAgent?.name || '',
      namespace: editAgent?.namespace || '',
      visibility: editAgent?.visibility || 'public',
      description: editAgent?.description || '',
      version: editAgent?.version || '',
      protocol_version: editAgent?.protocol_version || '0.3.0',
      provider_organization: editAgent?.provider_organization || '',
      provider_url: editAgent?.provider_url || '',
      documentation_url: editAgent?.documentation_url || '',
      icon_url: editAgent?.icon_url || '',
      business: editAgent?.business || '',
      department: editAgent?.department || '',
      backend_type: editAgent?.backend_type || 'service',
      backend_service_namespace: editAgent?.backend_service_namespace || '',
      backend_service_name: editAgent?.backend_service_name || '',
      backend_address: editAgent?.backend_address || '',
      preferred_interface_url: editAgent?.preferred_interface_url || '',
      preferred_protocol_binding: editAgent?.preferred_protocol_binding || 'jsonrpc',
      preferred_protocol_version: editAgent?.preferred_protocol_version || '0.3.0',
      streaming: !!editAgent?.streaming,
      push_notifications: !!editAgent?.push_notifications,
      extended_agent_card: !!editAgent?.extended_agent_card,
      source_type: editAgent?.source_type || 'manual',
      source_url: editAgent?.source_url || '',
      last_fetch_status: editAgent?.last_fetch_status || 'pending',
      raw_card_json: jsonPreview(editAgent?.raw_card_json),
      metadata_text: stringifyMetadata(editAgent?.metadata),
    });
    setInterfaces(editAgent?.interfaces?.length ? editAgent.interfaces : [defaultInterface()]);
    setSkills(editAgent?.skills?.length ? editAgent.skills : [defaultSkill()]);
  }, [visible, editAgent]);

  const updateInterface = (index: number, patch: Partial<A2AAgentInterface>) => {
    setInterfaces((prev) => prev.map((item, idx) => (idx === index ? { ...item, ...patch } : item)));
  };

  const updateSkill = (index: number, patch: Partial<A2AAgentSkill>) => {
    setSkills((prev) => prev.map((item, idx) => (idx === index ? { ...item, ...patch } : item)));
  };

  const onSubmit: FormProps['onSubmit'] = async (e) => {
    if (e.validateResult !== true) return;

    const data: A2AAgent = {
      id: editAgent?.id,
      name: form.getFieldValue('name') as string,
      namespace: form.getFieldValue('namespace') as string,
      visibility: form.getFieldValue('visibility') as string,
      description: form.getFieldValue('description') as string,
      version: form.getFieldValue('version') as string,
      protocol_version: form.getFieldValue('protocol_version') as string,
      provider_organization: form.getFieldValue('provider_organization') as string,
      provider_url: form.getFieldValue('provider_url') as string,
      documentation_url: form.getFieldValue('documentation_url') as string,
      icon_url: form.getFieldValue('icon_url') as string,
      business: form.getFieldValue('business') as string,
      department: form.getFieldValue('department') as string,
      backend_type: form.getFieldValue('backend_type') as string,
      backend_service_namespace: form.getFieldValue('backend_service_namespace') as string,
      backend_service_name: form.getFieldValue('backend_service_name') as string,
      backend_address: form.getFieldValue('backend_address') as string,
      preferred_interface_url: form.getFieldValue('preferred_interface_url') as string,
      preferred_protocol_binding: form.getFieldValue('preferred_protocol_binding') as string,
      preferred_protocol_version: form.getFieldValue('preferred_protocol_version') as string,
      streaming: !!form.getFieldValue('streaming'),
      push_notifications: !!form.getFieldValue('push_notifications'),
      extended_agent_card: !!form.getFieldValue('extended_agent_card'),
      source_type: form.getFieldValue('source_type') as string,
      source_url: form.getFieldValue('source_url') as string,
      last_fetch_status: form.getFieldValue('last_fetch_status') as string,
      raw_card_json: form.getFieldValue('raw_card_json') as string,
      metadata: parseMetadata(form.getFieldValue('metadata_text') as string),
      interfaces: normalizeInterfaces(interfaces),
      skills: normalizeSkills(skills),
      flag: editAgent?.flag,
    };

    const result = op === 'edit'
      ? await dispatch(updateA2AAgent({ param: data }))
      : await dispatch(saveA2AAgent({ param: data }));

    if (result.meta.requestStatus !== 'fulfilled') {
      openErrNotification('请求错误', result.payload as string);
      return;
    }
    openInfoNotification('请求成功', op === 'edit' ? '修改 A2A Agent 成功' : '创建 A2A Agent 成功');
    closeDrawer();
  };

  return (
    <Drawer
      size="large"
      header={op === 'edit' ? '编辑 A2A Agent' : '创建 A2A Agent'}
      footer={false}
      visible={visible}
      showOverlay={false}
      onClose={closeDrawer}
    >
      <Form form={form} layout="vertical" onSubmit={onSubmit}>
        <Tabs defaultValue="base">
          <TabPanel value="base" label="基础信息">
            <Row gutter={16}>
              <Col span={6}>
                <FormItem label="名称" name="name" rules={[
                  { required: true, message: '请输入 A2A Agent 名称' },
                  { pattern: /^[a-zA-Z0-9._-]+$/, message: '只允许数字、英文字母、.、-、_' },
                ]}
                >
                  <Input disabled={op === 'edit'} placeholder="例如 order-agent" />
                </FormItem>
              </Col>
              <Col span={6}>
                <FormItem label="命名空间" name="namespace" rules={[{ required: true, message: '请输入命名空间' }]}>
                  <Input disabled={op === 'edit'} placeholder="例如 default" />
                </FormItem>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={6}>
                <FormItem label="版本" name="version">
                  <Input placeholder="例如 1.0.0" />
                </FormItem>
              </Col>
              <Col span={6}>
                <FormItem label="A2A 协议版本" name="protocol_version">
                  <Input placeholder="例如 0.3.0" />
                </FormItem>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={6}>
                <FormItem label="业务" name="business">
                  <Input />
                </FormItem>
              </Col>
              <Col span={6}>
                <FormItem label="部门" name="department">
                  <Input />
                </FormItem>
              </Col>
            </Row>
            <FormItem label="描述" name="description">
              <Textarea autosize={{ minRows: 2, maxRows: 4 }} />
            </FormItem>
            <Row gutter={16}>
              <Col span={6}>
                <FormItem label="Provider 组织" name="provider_organization">
                  <Input />
                </FormItem>
              </Col>
              <Col span={6}>
                <FormItem label="Provider URL" name="provider_url">
                  <Input />
                </FormItem>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={6}>
                <FormItem label="文档地址" name="documentation_url">
                  <Input />
                </FormItem>
              </Col>
              <Col span={6}>
                <FormItem label="图标地址" name="icon_url">
                  <Input />
                </FormItem>
              </Col>
            </Row>
          </TabPanel>

          <TabPanel value="access" label="接入与能力">
            <Row gutter={16}>
              <Col span={4}>
                <FormItem label="首选协议" name="preferred_protocol_binding">
                  <Select options={protocolOptions} />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="首选协议版本" name="preferred_protocol_version">
                  <Input />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="首选访问地址" name="preferred_interface_url">
                  <Input />
                </FormItem>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={4}>
                <FormItem label="后端类型" name="backend_type">
                  <Select options={backendOptions} />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="后端服务命名空间" name="backend_service_namespace">
                  <Input placeholder="service 模式使用" />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="后端服务名" name="backend_service_name">
                  <Input placeholder="service 模式使用" />
                </FormItem>
              </Col>
            </Row>
            <FormItem label="后端地址" name="backend_address">
              <Input placeholder="URL 或外部系统地址" />
            </FormItem>
            <Row gutter={16}>
              <Col span={4}>
                <FormItem label="Streaming 能力" name="streaming">
                  <Switch size="large" label={['支持', '不支持']} />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="Push 能力" name="push_notifications">
                  <Switch size="large" label={['支持', '不支持']} />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="Extended Card" name="extended_agent_card">
                  <Switch size="large" label={['开启', '关闭']} />
                </FormItem>
              </Col>
            </Row>
            <div className={style.sectionHeader}>
              <Text>访问接口</Text>
              <Button size="small" variant="outline" icon={<AddIcon />} onClick={() => setInterfaces((prev) => [...prev, defaultInterface()])}>
                添加接口
              </Button>
            </div>
            {interfaces.map((item, index) => (
              <div className={style.inlineBlock} key={`${item.id || 'new'}-${index}`}>
                <Row gutter={12}>
                  <Col span={5}>
                    <Input value={item.url} placeholder="接口 URL" onChange={(value) => updateInterface(index, { url: value as string })} />
                  </Col>
                  <Col span={3}>
                    <Select value={item.protocol_binding} options={protocolOptions} onChange={(value) => updateInterface(index, { protocol_binding: value as string })} />
                  </Col>
                  <Col span={2}>
                    <Input value={item.protocol_version} placeholder="版本" onChange={(value) => updateInterface(index, { protocol_version: value as string })} />
                  </Col>
                  <Col span={1}>
                    <Input value={item.tenant} placeholder="租户" onChange={(value) => updateInterface(index, { tenant: value as string })} />
                  </Col>
                  <Col span={1}>
                    <Button
                      shape="square"
                      variant="text"
                      disabled={interfaces.length === 1}
                      onClick={() => setInterfaces((prev) => prev.filter((_, idx) => idx !== index))}
                    >
                      <DeleteIcon />
                    </Button>
                  </Col>
                </Row>
              </div>
            ))}
          </TabPanel>

          <TabPanel value="skills" label="技能">
            <div className={style.sectionHeader}>
              <Text>技能列表</Text>
              <Button size="small" variant="outline" icon={<AddIcon />} onClick={() => setSkills((prev) => [...prev, defaultSkill()])}>
                添加技能
              </Button>
            </div>
            {skills.map((skill, index) => (
              <div className={style.skillBlock} key={`${skill.id || 'new'}-${index}`}>
                <Row gutter={12}>
                  <Col span={4}>
                    <Input value={skill.skill_id} placeholder="Skill ID" onChange={(value) => updateSkill(index, { skill_id: value as string })} />
                  </Col>
                  <Col span={4}>
                    <Input value={skill.name} placeholder="技能名" onChange={(value) => updateSkill(index, { name: value as string })} />
                  </Col>
                  <Col span={3}>
                    <TagInput value={tagList(skill.tags)} placeholder="标签" onChange={(value) => updateSkill(index, { tags: value as string[] })} />
                  </Col>
                  <Col span={1}>
                    <Button
                      shape="square"
                      variant="text"
                      disabled={skills.length === 1}
                      onClick={() => setSkills((prev) => prev.filter((_, idx) => idx !== index))}
                    >
                      <DeleteIcon />
                    </Button>
                  </Col>
                </Row>
                <Row gutter={12}>
                  <Col span={4}>
                    <TagInput value={tagList(skill.input_modes)} placeholder="输入模式" onChange={(value) => updateSkill(index, { input_modes: value as string[] })} />
                  </Col>
                  <Col span={4}>
                    <TagInput value={tagList(skill.output_modes)} placeholder="输出模式" onChange={(value) => updateSkill(index, { output_modes: value as string[] })} />
                  </Col>
                  <Col span={4}>
                    <TagInput value={tagList(skill.examples)} placeholder="示例" onChange={(value) => updateSkill(index, { examples: value as string[] })} />
                  </Col>
                </Row>
                <Textarea value={skill.description} autosize={{ minRows: 1, maxRows: 3 }} placeholder="技能描述" onChange={(value) => updateSkill(index, { description: value as string })} />
              </div>
            ))}
          </TabPanel>

          <TabPanel value="source" label="来源与 Card">
            <Row gutter={16}>
              <Col span={4}>
                <FormItem label="来源类型" name="source_type">
                  <Input placeholder="manual / fetch" />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="拉取状态" name="last_fetch_status">
                  <Select options={fetchStatusOptions} />
                </FormItem>
              </Col>
              <Col span={4}>
                <FormItem label="来源地址" name="source_url">
                  <Input />
                </FormItem>
              </Col>
            </Row>
            <FormItem label="元数据" name="metadata_text">
              <Textarea placeholder="每行一个 key=value" autosize={{ minRows: 3, maxRows: 6 }} />
            </FormItem>
            <FormItem label="Agent Card JSON" name="raw_card_json">
              <Textarea className={style.codeText} autosize={{ minRows: 8, maxRows: 14 }} placeholder="可选，保存官方 Agent Card 原文" />
            </FormItem>
          </TabPanel>
        </Tabs>
        <FormItem className={style.formAction}>
          <Space>
            <Button type="submit" theme="primary">提交</Button>
            <Button type="reset" theme="default">重置</Button>
          </Space>
        </FormItem>
      </Form>
    </Drawer>
  );
};

export default memo(() => {
  const dispatch = useAppDispatch();
  const { datas, loading, page, limit, total, editAgent, skills, skillsLoading, card, cardLoading } = useAppSelector(selectA2A);
  const [query, setQuery] = useState({
    name: '',
    namespace: '',
    protocol_binding: '',
    skill_tag: '',
    backend_type: '',
    streaming: undefined as string | undefined,
    push_notifications: undefined as string | undefined,
  });
  const [editorState, setEditorState] = useState<{ visible: boolean; mode: Op }>({ visible: false, mode: 'create' });
  const [detailState, setDetailState] = useState<{ visible: boolean; mode: 'skills' | 'card'; agent?: A2AAgent }>({
    visible: false,
    mode: 'skills',
  });

  const cardText = useMemo(() => jsonPreview(card), [card]);

  const refreshTable = (current = 1, pageSize = 10, nextQuery = query) => {
    dispatch(listA2AAgents({
      param: {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        name: nextQuery.name || undefined,
        namespace: nextQuery.namespace || undefined,
        protocol_binding: nextQuery.protocol_binding || undefined,
        skill_tag: nextQuery.skill_tag || undefined,
        backend_type: nextQuery.backend_type || undefined,
        streaming: nextQuery.streaming === undefined ? undefined : nextQuery.streaming === 'true',
        push_notifications: nextQuery.push_notifications === undefined ? undefined : nextQuery.push_notifications === 'true',
      },
    })).then((res) => {
      if (res.meta.requestStatus === 'rejected') {
        openErrNotification('获取 A2A Agent 失败', res.payload as string);
      }
    });
  };

  useEffect(() => {
    refreshTable();
    return () => {
      dispatch(cleanA2APage());
    };
  }, []);

  const operateAgent = (op: Op | 'skills' | 'card', row?: TableRowData) => {
    const agent = row as A2AAgent;
    switch (op) {
      case 'create':
        dispatch(resetA2AAgent());
        setEditorState({ visible: true, mode: 'create' });
        break;
      case 'edit':
        dispatch(editorA2AAgent(agent));
        setEditorState({ visible: true, mode: 'edit' });
        break;
      case 'delete':
        dispatch(removeA2AAgents({ ids: [agent?.id as string] })).then((res) => {
          if (res.meta.requestStatus === 'fulfilled') {
            openInfoNotification('请求成功', '删除 A2A Agent 成功');
            refreshTable(page, limit);
          } else {
            openErrNotification('请求错误', res.payload as string);
          }
        });
        break;
      case 'skills':
        setDetailState({ visible: true, mode: 'skills', agent });
        dispatch(listA2AAgentSkills({
          param: {
            agent_id: agent?.id,
          },
        })).then((res) => {
          if (res.meta.requestStatus === 'rejected') {
            openErrNotification('获取 A2A 技能失败', res.payload as string);
          }
        });
        break;
      case 'card':
        setDetailState({ visible: true, mode: 'card', agent });
        dispatch(getA2AAgentCard({ id: agent?.id as string })).then((res) => {
          if (res.meta.requestStatus === 'rejected') {
            openErrNotification('获取 Agent Card 失败', res.payload as string);
          }
        });
        break;
      default:
        break;
    }
  };

  const submitFilter = () => {
    refreshTable(1, limit, query);
  };

  const resetFilter = () => {
    const nextQuery = {
      name: '',
      namespace: '',
      protocol_binding: '',
      skill_tag: '',
      backend_type: '',
      streaming: undefined,
      push_notifications: undefined,
    };
    setQuery(nextQuery);
    refreshTable(1, limit, nextQuery);
  };

  return (
    <div>
      <Row justify="space-between" className={style.toolBar}>
        <Col>
          <Button icon={<AddIcon />} onClick={() => operateAgent('create')}>新建</Button>
        </Col>
        <Col>
          <Space breakLine>
            <Input
              className={style.filterInput}
              clearable
              placeholder="名称前缀"
              value={query.name}
              onChange={(value) => setQuery((prev) => ({ ...prev, name: value as string }))}
            />
            <Input
              className={style.filterInput}
              clearable
              placeholder="命名空间"
              value={query.namespace}
              onChange={(value) => setQuery((prev) => ({ ...prev, namespace: value as string }))}
            />
            <Select
              className={style.filterSelect}
              clearable
              placeholder="协议"
              options={protocolOptions}
              value={query.protocol_binding}
              onChange={(value) => setQuery((prev) => ({ ...prev, protocol_binding: value as string }))}
            />
            <Input
              className={style.filterInput}
              clearable
              placeholder="Skill Tag"
              value={query.skill_tag}
              onChange={(value) => setQuery((prev) => ({ ...prev, skill_tag: value as string }))}
            />
            <Select
              className={style.filterSelect}
              clearable
              placeholder="后端类型"
              options={backendOptions}
              value={query.backend_type}
              onChange={(value) => setQuery((prev) => ({ ...prev, backend_type: value as string }))}
            />
            <Select
              className={style.filterSelect}
              clearable
              placeholder="Streaming"
              options={capabilityFilterOptions}
              value={query.streaming}
              onChange={(value) => setQuery((prev) => ({ ...prev, streaming: value as string }))}
            />
            <Select
              className={style.filterSelect}
              clearable
              placeholder="Push"
              options={capabilityFilterOptions}
              value={query.push_notifications}
              onChange={(value) => setQuery((prev) => ({ ...prev, push_notifications: value as string }))}
            />
            <Button variant="outline" onClick={submitFilter}>查询</Button>
            <Button variant="text" onClick={resetFilter}>重置</Button>
            <Tooltip content="刷新">
              <Button shape="square" variant="text" onClick={() => refreshTable(page, limit)}>
                <RefreshIcon />
              </Button>
            </Tooltip>
          </Space>
        </Col>
      </Row>

      <Table
        data={datas}
        columns={agentColumns(operateAgent)}
        loading={loading}
        rowKey="id"
        size="large"
        tableLayout="auto"
        cellEmptyContent="-"
        pagination={{
          current: page,
          pageSize: limit,
          total,
          showJumper: true,
          onChange(pageInfo: PageInfo) {
            refreshTable(pageInfo.current, pageInfo.pageSize);
          },
        }}
        onPageChange={(pageInfo) => {
          refreshTable(pageInfo.current, pageInfo.pageSize);
        }}
      />

      {editorState.visible && (
        <A2AEditor
          key={`${editorState.mode}-${editAgent?.id || 'new'}`}
          op={editorState.mode}
          visible={editorState.visible}
          closeDrawer={() => {
            dispatch(resetA2AAgent());
            setEditorState((prev) => ({ ...prev, visible: false }));
            refreshTable(page, limit);
          }}
        />
      )}

      <Drawer
        size="large"
        header={detailState.agent ? `${detailState.agent.namespace}/${detailState.agent.name}` : 'A2A Agent'}
        footer={false}
        visible={detailState.visible}
        onClose={() => {
          dispatch(cleanA2ADetails());
          setDetailState({ visible: false, mode: 'skills' });
        }}
      >
        {detailState.mode === 'skills' ? (
          <Table<A2AAgentSkill>
            data={skills}
            columns={skillColumns}
            loading={skillsLoading}
            rowKey="id"
            size="large"
            tableLayout="auto"
            cellEmptyContent="-"
          />
        ) : (
          <div>
            <Descriptions column={2} size="small" className={style.cardMeta}>
              <DescriptionsItem label="名称">{detailState.agent?.name}</DescriptionsItem>
              <DescriptionsItem label="命名空间">{detailState.agent?.namespace}</DescriptionsItem>
              <DescriptionsItem label="首选协议">{detailState.agent?.preferred_protocol_binding || '-'}</DescriptionsItem>
              <DescriptionsItem label="首选地址">{detailState.agent?.preferred_interface_url || '-'}</DescriptionsItem>
            </Descriptions>
            <Textarea
              className={style.codeText}
              readonly
              value={cardLoading ? '加载中...' : cardText}
              autosize={{ minRows: 16, maxRows: 24 }}
            />
          </div>
        )}
      </Drawer>
    </div>
  );
});
