import React from 'react';
import { Breadcrumb, Button, Empty, Link, Loading, Space, Tag, Tooltip } from 'components/Fluent';
import { EditIcon, InfoCircleIcon, RefreshIcon, ServerIcon, ToolsCircleIcon } from 'components/Fluent/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';

import AuthorizeInput from 'components/Authorize';
import { useAppDispatch } from 'modules/store';
import { editorMCPServer, resetMCPServer } from 'modules/ai/mcp';
import { describeMCPServers, describeMCPServerTools, MCPServer, MCPServerTool } from 'services/mcp';
import { PolicySourceType } from 'services/auth_policy';
import { Op } from 'services/types';
import {
  backendLabel,
  backendServiceRef,
  backendType,
  backendTypeLabel,
  MCPEditor,
  protocolLabel,
  protocolTheme,
  ToolExplorer,
} from './index';
import style from './index.module.less';

const { BreadcrumbItem } = Breadcrumb;

const MCPDetailPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const id = searchParams.get('id') || '';
  const namespace = searchParams.get('namespace') || '';
  const name = searchParams.get('name') || '';
  const [server, setServer] = React.useState<MCPServer | null>(null);
  const [tools, setTools] = React.useState<MCPServerTool[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [toolsLoading, setToolsLoading] = React.useState(false);
  const [serverError, setServerError] = React.useState('');
  const [serverNotFound, setServerNotFound] = React.useState(false);
  const [toolsError, setToolsError] = React.useState('');
  const [editorState, setEditorState] = React.useState<{ visible: boolean; mode: Op }>({ visible: false, mode: 'edit' });
  const [authorizeVisible, setAuthorizeVisible] = React.useState(false);

  const loadServer = React.useCallback(async () => {
    setLoading(true);
    setServerError('');
    setServerNotFound(false);
    try {
      const response = await describeMCPServers({
        offset: 0,
        limit: 10000,
        name: name || undefined,
        namespace: namespace || undefined,
      });
      const next = response.list.find((item) => (id ? item.id === id : item.name === name && item.namespace === namespace)) || null;
      setServer(next);
      if (!next) {
        setServerNotFound(true);
      }
    } catch (error) {
      setServer(null);
      setServerError((error as Error).message || '请求失败');
    } finally {
      setLoading(false);
    }
  }, [id, name, namespace]);

  const loadTools = React.useCallback(async (target?: MCPServer | null) => {
    const current = target || server;
    if (!current?.id) {
      setTools([]);
      setToolsError('');
      return;
    }
    setToolsLoading(true);
    setToolsError('');
    try {
      const response = await describeMCPServerTools({
        offset: 0,
        limit: 100,
        server_id: current.id,
      });
      setTools(response.list);
    } catch (error) {
      setToolsError((error as Error).message || '请求失败');
    } finally {
      setToolsLoading(false);
    }
  }, [server]);

  React.useEffect(() => {
    loadServer();
  }, [loadServer]);

  React.useEffect(() => {
    loadTools(server);
  }, [server, loadTools]);

  const goBackendService = () => {
    const service = backendServiceRef(server || undefined);
    if (!service) return;
    navigate(`/discovery/service/instance?namespace=${encodeURIComponent(service.namespace)}&service=${encodeURIComponent(service.name)}`);
  };

  const openEditor = () => {
    if (!server) return;
    dispatch(editorMCPServer(server));
    setEditorState({ visible: true, mode: 'edit' });
  };

  const closeEditor = () => {
    dispatch(resetMCPServer());
    setEditorState((prev) => ({ ...prev, visible: false }));
    loadServer();
  };

  return (
    <div className={style.page}>
      <Breadcrumb className={style.detailBreadcrumb} maxItemWidth="220px">
        <BreadcrumbItem onClick={() => navigate('/ai/mcps')}>MCP 服务</BreadcrumbItem>
        <BreadcrumbItem>{server ? `${server.namespace}/${server.name}` : name || 'MCP 服务详情'}</BreadcrumbItem>
      </Breadcrumb>

      {loading && (
        <section className={style.detailLoading}>
          <Loading text="加载 MCP 服务详情..." />
        </section>
      )}

      {!loading && !server && (
        <section className={style.detailLoading}>
          <Empty
            title={serverNotFound ? '未找到 MCP Server' : 'MCP Server 加载失败'}
            description={serverNotFound ? '该资源可能已删除，或当前链接参数已经失效。' : serverError}
            action={<Button variant="outline" icon={<RefreshIcon />} onClick={loadServer}>重新加载</Button>}
          />
        </section>
      )}

      {!loading && server && (
        <div className={style.toolDrawer}>
          <section className={style.toolDrawerSummary}>
            <div className={style.toolDrawerIcon}>
              <ServerIcon />
            </div>
            <div className={style.toolDrawerMain}>
              <div className={style.toolDrawerTitle}>
                <h3>{server.namespace}/{server.name}</h3>
                <Tag theme={protocolTheme(server.protocol) as any} variant="light">
                  {protocolLabel(server.protocol)}
                </Tag>
              </div>
              <div className={style.toolDrawerDesc}>
                {server.description || server.reference || '该 MCP Server 暂无描述。'}
              </div>
              <div className={style.toolMetaGrid}>
                <div>
                  <span>工具数</span>
                  <strong>{toolsLoading || toolsError ? '-' : tools.length}</strong>
                </div>
                <div>
                  <span>接入</span>
                  <strong>{backendTypeLabel(backendType(server))}</strong>
                </div>
                <div>
                  <span>后端</span>
                  <strong>
                    {backendServiceRef(server) ? (
                      <Link theme="primary" onClick={goBackendService}>
                        {backendLabel(server)}
                      </Link>
                    ) : backendLabel(server)}
                  </strong>
                </div>
                <div>
                  <span>最近修改</span>
                  <strong>{server.mtime || '-'}</strong>
                </div>
              </div>
            </div>
            <div className={style.toolDrawerActions}>
              <Tooltip content="刷新工具">
                <Button aria-label="刷新 MCP 工具" className={style.drawerActionIconButton} shape="square" variant="outline" onClick={() => loadTools(server)}>
                  <RefreshIcon />
                </Button>
              </Tooltip>
              <Button variant="outline" icon={<EditIcon />} onClick={openEditor}>
                编辑 Server
              </Button>
              <Button variant="outline" onClick={() => setAuthorizeVisible(true)}>
                授权
              </Button>
            </div>
          </section>

          {toolsError ? (
            <section className={style.toolEmpty}>
              <ToolsCircleIcon />
              <h4>MCP 工具同步失败</h4>
              <p>{toolsError}</p>
              <Button theme="primary" icon={<RefreshIcon />} onClick={() => loadTools(server)}>重新同步</Button>
            </section>
          ) : !toolsLoading && tools.length === 0 ? (
            <section className={style.toolEmpty}>
              <ToolsCircleIcon />
              <h4>暂无工具同步</h4>
              <p>当前 MCP Server 没有返回工具定义。确认服务已连通，并检查协议、引用地址和命名空间配置。</p>
              <div className={style.toolCheckList}>
                <div><InfoCircleIcon />后端：{backendLabel(server)}</div>
                <div><InfoCircleIcon />接入协议：{protocolLabel(server.protocol)}</div>
                <div><InfoCircleIcon />Server ID：{server.id || '-'}</div>
              </div>
              <Space>
                <Button theme="primary" icon={<RefreshIcon />} onClick={() => loadTools(server)}>
                  刷新工具
                </Button>
                <Button variant="outline" icon={<EditIcon />} onClick={openEditor}>
                  编辑 Server
                </Button>
              </Space>
            </section>
          ) : (
            <section className={style.toolTableSurface}>
              <div className={style.tableHeader}>
                <div>
                  <strong>工具浏览</strong>
                  <span>{toolsLoading ? '正在同步工具' : `当前显示 ${tools.length} 个工具`}</span>
                </div>
              </div>
              {toolsLoading ? (
                <div className={style.detailLoading}>
                  <Loading text="同步工具..." />
                </div>
              ) : (
                <ToolExplorer tools={tools} />
              )}
            </section>
          )}
        </div>
      )}

      {editorState.visible && (
        <MCPEditor
          key={`${editorState.mode}-${server?.id || 'new'}`}
          op={editorState.mode}
          visible={editorState.visible}
          closeDrawer={closeEditor}
        />
      )}

      {authorizeVisible && server?.id && (
        <AuthorizeInput
          resource_type={PolicySourceType.MCPServerResources}
          resource_id={server.id}
          resource_name={`${server.namespace}/${server.name}`}
          visible={authorizeVisible}
          onClose={() => setAuthorizeVisible(false)}
        />
      )}
    </div>
  );
};

export default React.memo(MCPDetailPage);
