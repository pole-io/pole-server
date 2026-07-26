import React from 'react';
import { Breadcrumb, Button, Empty, Link, Loading, Table, Tag, TableProps } from 'components/Fluent';
import { CopyIcon, RefreshIcon } from 'components/Fluent/icons';
import { useNavigate } from 'components/Router';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { enableUserToken, listOneUser, resetUserToken, selectUser } from 'modules/user/users';
import { describeAuthPolicies, PolicyRule } from 'services/auth_policy';
import { User } from 'services/users';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { copyToClipboard } from 'utils/sys';
import style from './index.module.less';

const { BreadcrumbItem } = Breadcrumb;

const formatValue = (value?: string) => value || '-';

const UserDetailPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const searchParams = new URLSearchParams(window.location.search);
  const userId = searchParams.get('id') || '';
  const username = searchParams.get('name') || '用户详情';
  const { viewUser } = useAppSelector(selectUser);
  const [viewState, setViewState] = React.useState<{
    loading: boolean;
    user: User | null;
    policies: PolicyRule[];
    policyTotal: number;
    policyLoading: boolean;
    fetchError: boolean;
  }>({
    loading: false,
    user: null,
    policies: [],
    policyTotal: 0,
    policyLoading: false,
    fetchError: false,
  });

  const loadUserPolicies = React.useCallback(async (id: string) => {
    setViewState(prev => ({ ...prev, policyLoading: true }));
    try {
      const ret = await describeAuthPolicies({
        principal_id: id,
        principal_type: 1,
        offset: 0,
        limit: 100,
      });
      setViewState(prev => ({
        ...prev,
        policies: ret.content,
        policyTotal: ret.totalCount,
        policyLoading: false,
      }));
    } catch (error) {
      setViewState(prev => ({ ...prev, policyLoading: false }));
      openErrNotification('请求错误', `获取关联策略失败, ${(error as Error).message}`);
    }
  }, []);

  const loadUser = React.useCallback(async () => {
    if (!userId) return;
    setViewState(prev => ({ ...prev, loading: true, fetchError: false }));
    const result = await dispatch(listOneUser({ id: userId }));
    if (result.meta.requestStatus === 'rejected') {
      setViewState(prev => ({ ...prev, loading: false, fetchError: true }));
      openErrNotification('请求错误', `获取用户详情失败, ${result.payload as string}`);
      return;
    }

    const user = (result.payload as { viewUser?: User }).viewUser;
    if (!user) {
      setViewState(prev => ({ ...prev, loading: false, fetchError: true }));
      return;
    }
    setViewState(prev => ({ ...prev, loading: false, user, fetchError: false }));
    loadUserPolicies(user.id || userId);
  }, [dispatch, loadUserPolicies, userId]);

  React.useEffect(() => {
    loadUser();
  }, [loadUser]);

  const currentUser = viewState.user || viewUser || null;
  const metadataEntries = Object.entries(currentUser?.metadata || {}).map(([key, value]) => ({
    key,
    value: typeof value === 'string' ? value : JSON.stringify(value),
  }));
  const isMainUser = currentUser?.user_type === 'main';
  const userInitial = (currentUser?.name || username || 'U').slice(0, 1).toUpperCase();

  const refreshAfterTokenChange = () => {
    setViewState(prev => ({ ...prev, loading: true }));
    window.setTimeout(loadUser, 1000);
  };

  const handleResetToken = async () => {
    if (!userId) return;
    const result = await dispatch(resetUserToken({ id: userId }));
    if (result.meta.requestStatus === 'rejected') {
      openErrNotification('请求错误', '资源访问凭据重置失败');
      return;
    }
    openInfoNotification('请求成功', '资源访问凭据已重置');
    refreshAfterTokenChange();
  };

  const handleToggleToken = async () => {
    if (!userId || !currentUser) return;
    const enabled = Boolean(currentUser.token_enable);
    const result = await dispatch(enableUserToken({ id: userId, token_enable: !enabled }));
    if (result.meta.requestStatus === 'rejected') {
      openErrNotification('请求错误', `资源访问凭据${enabled ? '禁用' : '启用'}失败`);
      return;
    }
    openInfoNotification('请求成功', `资源访问凭据已${enabled ? '禁用' : '启用'}`);
    refreshAfterTokenChange();
  };

  const policyColumns: TableProps['columns'] = [
    {
      colKey: 'name',
      title: '策略名称',
      width: 260,
      minWidth: 220,
      ellipsis: true,
      cell: ({ row }) => (
        <Link
          theme="primary"
          onClick={() => navigate(`/auth/policies/detail?id=${encodeURIComponent(String(row.id || ''))}&name=${encodeURIComponent(String(row.name || ''))}`)}
        >
          {row.name || '-'}
        </Link>
      ),
    },
    {
      colKey: 'action',
      title: '效果',
      width: 112,
      cell: ({ row }) => (
        <Tag theme={row.action === 'ALLOW' ? 'success' : 'danger'} variant="light">
          {row.action === 'ALLOW' ? '允许' : '拒绝'}
        </Tag>
      ),
    },
    {
      colKey: 'default_strategy',
      title: '策略类型',
      width: 126,
      cell: ({ row }) => (
        <Tag theme={row.default_strategy ? 'primary' : 'default'} variant="outline">
          {row.default_strategy ? '默认策略' : '自定义策略'}
        </Tag>
      ),
    },
    {
      colKey: 'comment',
      title: '描述',
      width: 260,
      minWidth: 200,
      ellipsis: true,
      cell: ({ row }) => row.comment || '-',
    },
    {
      colKey: 'mtime',
      title: '更新时间',
      width: 180,
      cell: ({ row }) => row.mtime || row.ctime || '-',
    },
  ];

  return (
    <div className={style.userDetailPage}>
      <Breadcrumb className={style.detailBreadcrumb} maxItemWidth="240px">
        <BreadcrumbItem onClick={() => navigate('/auth/principals')}>用户</BreadcrumbItem>
        <BreadcrumbItem>{currentUser?.name || username}</BreadcrumbItem>
      </Breadcrumb>

      <Loading indicator loading={viewState.loading} preventScrollThrough showOverlay>
        {!userId || viewState.fetchError || !currentUser ? (
          <section className={style.userDetailEmpty}>
            <Empty description={userId ? '未找到用户详情' : '缺少用户标识，无法加载详情'} />
          </section>
        ) : (
          <main className={style.userDetailStack}>
            <section className={style.userDetailSummary}>
              <div className={style.userDetailAvatar}>{userInitial}</div>
              <div className={style.userDetailMain}>
                <div className={style.userDetailTitleLine}>
                  <h2>{currentUser.name}</h2>
                  <Tag theme="primary" variant="light">{isMainUser ? '管理员' : '子用户'}</Tag>
                </div>
                <p>{currentUser.comment || '暂无备注'}</p>
                <div className={style.userMetaGrid}>
                  <div className={style.userMetaItem}>
                    <span>用户 ID</span>
                    <strong className={style.userMono} title={currentUser.id}>{formatValue(currentUser.id)}</strong>
                  </div>
                  <div className={style.userMetaItem}>
                    <span>来源</span>
                    <strong>{formatValue(currentUser.source)}</strong>
                  </div>
                  <div className={style.userMetaItem}>
                    <span>创建时间</span>
                    <strong>{formatValue(currentUser.ctime)}</strong>
                  </div>
                  <div className={style.userMetaItem}>
                    <span>邮箱</span>
                    <strong>{formatValue(currentUser.email)}</strong>
                  </div>
                  <div className={style.userMetaItem}>
                    <span>手机号</span>
                    <strong>{formatValue(currentUser.mobile)}</strong>
                  </div>
                  <div className={style.userMetaItem}>
                    <span>关联策略</span>
                    <strong>{viewState.policyLoading ? '-' : `${viewState.policyTotal} 条`}</strong>
                  </div>
                </div>
              </div>
            </section>

            <div className={style.userDetailContentGrid}>
              <section className={style.userCredentialPanel}>
                <div className={style.userPanelHeader}>
                  <div>
                    <strong>访问凭据</strong>
                    <span>用于 Console 和 API 的用户身份验证。</span>
                  </div>
                  <Tag theme={currentUser.token_enable ? 'success' : 'danger'} variant="light">
                    {currentUser.token_enable ? '已启用' : '已禁用'}
                  </Tag>
                </div>
                <div className={style.userTokenRow}>
                  <div>
                    <span>访问 Token</span>
                    <code>{currentUser.auth_token ? '••••••••••••••••' : '暂无 Token'}</code>
                  </div>
                  <div className={style.userTokenActions}>
                    <Button
                      variant="outline"
                      icon={<CopyIcon />}
                      disabled={!currentUser.auth_token}
                      onClick={() => copyToClipboard(currentUser.auth_token, '资源访问凭据已复制到剪贴板')}
                    >
                      复制
                    </Button>
                    <Button variant="outline" icon={<RefreshIcon />} onClick={handleResetToken}>重置</Button>
                    <Button theme={currentUser.token_enable ? 'danger' : 'primary'} variant="outline" onClick={handleToggleToken}>
                      {currentUser.token_enable ? '禁用' : '启用'}
                    </Button>
                  </div>
                </div>
              </section>

              <section className={style.userTagPanel}>
                <div className={style.userPanelHeader}>
                  <div>
                    <strong>标签</strong>
                    <span>{metadataEntries.length ? `当前有 ${metadataEntries.length} 个标签` : '当前没有标签'}</span>
                  </div>
                </div>
                {metadataEntries.length ? (
                  <div className={style.userLabelList}>
                    {metadataEntries.map(item => <Tag key={item.key} theme="default" variant="outline">{item.key}: {item.value}</Tag>)}
                  </div>
                ) : (
                  <p className={style.userPanelEmpty}>暂无标签</p>
                )}
              </section>
            </div>

            <section className={style.userPolicySurface}>
              <div className={style.userTableHeader}>
                <div>
                  <strong>权限信息</strong>
                  <span>{viewState.policyLoading ? '正在加载关联策略' : `当前关联 ${viewState.policyTotal} 条策略`}</span>
                </div>
              </div>
              <Table
                rowKey="id"
                size="medium"
                tableLayout="fixed"
                cellEmptyContent="-"
                loading={viewState.policyLoading}
                columns={policyColumns}
                data={viewState.policies}
                pagination={{
                  pageSize: 100,
                  total: viewState.policyTotal,
                  current: 1,
                  showPageSize: false,
                  showJumper: false,
                }}
              />
            </section>
          </main>
        )}
      </Loading>
    </div>
  );
};

export default React.memo(UserDetailPage);
