import React, { memo, useEffect, useState } from 'react';
import {
  Button,
  Col,
  Drawer,
  Form,
  Input,
  Link,
  Popconfirm,
  PrimaryTableProps,
  Row,
  Select,
  Space,
  Table,
  TableRowData,
  Tag,
  Tooltip,
} from 'tdesign-react';
import type { FormProps, PageInfo } from 'tdesign-react';
import { AddIcon, DeleteIcon, EditIcon, ListIcon, RefreshIcon } from 'tdesign-icons-react';

import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import {
  cleanMCPPage,
  cleanMCPTools,
  editorMCPServer,
  listMCPServers,
  listMCPServerTools,
  removeMCPServers,
  resetMCPServer,
  saveMCPServer,
  selectMCP,
  updateMCPServer,
} from 'modules/ai/mcp';
import { MCPServer, MCPServerTool } from 'services/mcp';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { Op } from 'services/types';
import style from './index.module.less';

const { FormItem } = Form;

const protocolOptions = [
  { label: 'HTTP', value: 'http' },
  { label: 'SSE', value: 'sse' },
  { label: 'Streamable HTTP', value: 'streamable-http' },
];

const serverColumns = (
  operateServer: (op: Op | 'tools', row?: TableRowData) => void,
): PrimaryTableProps['columns'] => [
  {
    colKey: 'name',
    title: '名称',
    fixed: 'left',
    cell: ({ row }) => (
      <Link theme="primary" onClick={() => operateServer('tools', row)}>
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
    colKey: 'protocol',
    title: '协议',
    cell: ({ row }) => <Tag variant="outline">{row.protocol || '-'}</Tag>,
  },
  {
    colKey: 'ports',
    title: '端口',
    ellipsis: true,
    cell: ({ row }) => <Text>{row.ports || '-'}</Text>,
  },
  {
    colKey: 'business',
    title: '业务',
    cell: ({ row }) => <Text>{row.business || '-'}</Text>,
  },
  {
    colKey: 'department',
    title: '部门',
    cell: ({ row }) => <Text>{row.department || '-'}</Text>,
  },
  {
    colKey: 'description',
    title: '描述',
    ellipsis: true,
    cell: ({ row }) => <Text>{row.description || '-'}</Text>,
  },
  {
    colKey: 'time',
    title: '操作时间',
    cell: ({ row }) => <Text>修改: {row.mtime || '-'}<br />创建: {row.ctime || '-'}</Text>,
  },
  {
    colKey: 'action',
    title: '操作',
    cell: ({ row }) => (
      <Space>
        <Tooltip content="查看工具">
          <Button shape="square" variant="text" onClick={() => operateServer('tools', row)}>
            <ListIcon />
          </Button>
        </Tooltip>
        <Tooltip content="编辑">
          <Button shape="square" variant="text" onClick={() => operateServer('edit', row)}>
            <EditIcon />
          </Button>
        </Tooltip>
        <Tooltip content="删除">
          <Popconfirm
            content="确认删除该 MCP Server 吗"
            destroyOnClose
            placement="top"
            showArrow
            theme="default"
            onConfirm={() => operateServer('delete', row)}
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

const toolColumns: PrimaryTableProps['columns'] = [
  {
    colKey: 'name',
    title: '工具名',
    fixed: 'left',
    cell: ({ row }) => <Text>{row.name}</Text>,
  },
  {
    colKey: 'description',
    title: '描述',
    ellipsis: true,
    cell: ({ row }) => <Text>{row.description || '-'}</Text>,
  },
  {
    colKey: 'input_schema',
    title: '输入 Schema',
    ellipsis: true,
    cell: ({ row }) => <Text>{row.input_schema || '-'}</Text>,
  },
  {
    colKey: 'output_schema',
    title: '输出 Schema',
    ellipsis: true,
    cell: ({ row }) => <Text>{row.output_schema || '-'}</Text>,
  },
  {
    colKey: 'time',
    title: '操作时间',
    cell: ({ row }) => <Text>修改: {row.mtime || '-'}<br />创建: {row.ctime || '-'}</Text>,
  },
];

function splitExportTo(value?: string) {
  return value ? value.split(',').map((item) => item.trim()).filter(Boolean) : [];
}

function joinExportTo(value?: string[]) {
  return value?.filter(Boolean).join(',') || '';
}

const MCPEditor: React.FC<{
  op: Op;
  visible: boolean;
  closeDrawer: () => void;
}> = ({ op, visible, closeDrawer }) => {
  const [form] = Form.useForm();
  const dispatch = useAppDispatch();
  const { editServer } = useAppSelector(selectMCP);

  useEffect(() => {
    if (!visible) return;
    form.setFieldsValue({
      name: editServer?.name || '',
      namespace: editServer?.namespace || '',
      ports: editServer?.ports || '',
      protocol: editServer?.protocol || 'http',
      business: editServer?.business || '',
      department: editServer?.department || '',
      description: editServer?.description || '',
      reference: editServer?.reference || '',
      export_to: splitExportTo(editServer?.export_to),
    });
  }, [visible, editServer]);

  const onSubmit: FormProps['onSubmit'] = async (e) => {
    if (e.validateResult !== true) return;

    const data: MCPServer = {
      id: editServer?.id,
      name: form.getFieldValue('name') as string,
      namespace: form.getFieldValue('namespace') as string,
      ports: form.getFieldValue('ports') as string,
      protocol: form.getFieldValue('protocol') as string,
      business: form.getFieldValue('business') as string,
      department: form.getFieldValue('department') as string,
      description: form.getFieldValue('description') as string,
      reference: form.getFieldValue('reference') as string,
      export_to: joinExportTo(form.getFieldValue('export_to') as string[]),
      revision: editServer?.revision,
    };

    const result = op === 'edit'
      ? await dispatch(updateMCPServer({ param: data }))
      : await dispatch(saveMCPServer({ param: data }));

    if (result.meta.requestStatus !== 'fulfilled') {
      openErrNotification('请求错误', result.payload as string);
      return;
    }
    openInfoNotification('请求成功', op === 'edit' ? '修改 MCP Server 成功' : '创建 MCP Server 成功');
    closeDrawer();
  };

  return (
    <Drawer
      size="large"
      header={op === 'edit' ? '编辑 MCP Server' : '创建 MCP Server'}
      footer={false}
      visible={visible}
      showOverlay={false}
      onClose={closeDrawer}
    >
      <Form form={form} layout="vertical" onSubmit={onSubmit}>
        <FormItem
          label="名称"
          name="name"
          rules={[
            { required: true, message: '请输入 MCP Server 名称' },
            { pattern: /^[a-zA-Z0-9._-]+$/, message: '只允许数字、英文字母、.、-、_' },
          ]}
        >
          <Input disabled={op === 'edit'} placeholder="例如 order-query" />
        </FormItem>
        <FormItem label="命名空间" name="namespace" rules={[{ required: true, message: '请输入命名空间' }]}>
          <Input disabled={op === 'edit'} placeholder="例如 default" />
        </FormItem>
        <FormItem label="协议" name="protocol">
          <Select options={protocolOptions} />
        </FormItem>
        <FormItem label="端口" name="ports">
          <Input placeholder="例如 http:8080 或 8080" />
        </FormItem>
        <FormItem label="业务" name="business">
          <Input />
        </FormItem>
        <FormItem label="部门" name="department">
          <Input />
        </FormItem>
        <FormItem label="描述" name="description">
          <Input />
        </FormItem>
        <FormItem label="引用地址" name="reference">
          <Input placeholder="MCP Server 的访问地址或资源引用" />
        </FormItem>
        <FormItem label="可见命名空间" name="export_to">
          <Select creatable multiple filterable placeholder="为空表示默认可见范围" />
        </FormItem>
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
  const { datas, loading, page, limit, total, editServer, tools, toolsLoading } = useAppSelector(selectMCP);
  const [query, setQuery] = useState({
    name: '',
    namespace: '',
    protocol: '',
  });
  const [editorState, setEditorState] = useState<{ visible: boolean; mode: Op }>({ visible: false, mode: 'create' });
  const [toolsState, setToolsState] = useState<{ visible: boolean; server?: MCPServer }>({ visible: false });

  const refreshTable = (current = 1, pageSize = 10, nextQuery = query) => {
    dispatch(listMCPServers({
      param: {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        name: nextQuery.name || undefined,
        namespace: nextQuery.namespace || undefined,
        protocol: nextQuery.protocol || undefined,
      },
    })).then((res) => {
      if (res.meta.requestStatus === 'rejected') {
        openErrNotification('获取 MCP Server 失败', res.payload as string);
      }
    });
  };

  useEffect(() => {
    refreshTable();
    return () => {
      dispatch(cleanMCPPage());
    };
  }, []);

  const operateServer = (op: Op | 'tools', row?: TableRowData) => {
    switch (op) {
      case 'create':
        dispatch(resetMCPServer());
        setEditorState({ visible: true, mode: 'create' });
        break;
      case 'edit':
        dispatch(editorMCPServer(row as MCPServer));
        setEditorState({ visible: true, mode: 'edit' });
        break;
      case 'delete':
        dispatch(removeMCPServers({ ids: [row?.id as string] })).then((res) => {
          if (res.meta.requestStatus === 'fulfilled') {
            openInfoNotification('请求成功', '删除 MCP Server 成功');
            refreshTable(page, limit);
          } else {
            openErrNotification('请求错误', res.payload as string);
          }
        });
        break;
      case 'tools':
        setToolsState({ visible: true, server: row as MCPServer });
        dispatch(listMCPServerTools({
          param: {
            offset: 0,
            limit: 100,
            server_id: row?.id as string,
          },
        })).then((res) => {
          if (res.meta.requestStatus === 'rejected') {
            openErrNotification('获取 MCP 工具失败', res.payload as string);
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
    const nextQuery = { name: '', namespace: '', protocol: '' };
    setQuery(nextQuery);
    refreshTable(1, limit, nextQuery);
  };

  return (
    <div>
      <Row justify="space-between" className={style.toolBar}>
        <Col>
          <Button icon={<AddIcon />} onClick={() => operateServer('create')}>新建</Button>
        </Col>
        <Col>
          <Space>
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
              className={style.protocolFilter}
              clearable
              placeholder="协议"
              options={protocolOptions}
              value={query.protocol}
              onChange={(value) => setQuery((prev) => ({ ...prev, protocol: value as string }))}
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
        columns={serverColumns(operateServer)}
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
        <MCPEditor
          key={`${editorState.mode}-${editServer?.id || 'new'}`}
          op={editorState.mode}
          visible={editorState.visible}
          closeDrawer={() => {
            dispatch(resetMCPServer());
            setEditorState((prev) => ({ ...prev, visible: false }));
            refreshTable(page, limit);
          }}
        />
      )}

      <Drawer
        size="large"
        header={toolsState.server ? `${toolsState.server.namespace}/${toolsState.server.name} 的工具` : 'MCP Server 工具'}
        footer={false}
        visible={toolsState.visible}
        onClose={() => {
          dispatch(cleanMCPTools());
          setToolsState({ visible: false });
        }}
      >
        <Table<MCPServerTool>
          data={tools}
          columns={toolColumns}
          loading={toolsLoading}
          rowKey="id"
          size="large"
          tableLayout="auto"
          cellEmptyContent="-"
        />
      </Drawer>
    </div>
  );
});
