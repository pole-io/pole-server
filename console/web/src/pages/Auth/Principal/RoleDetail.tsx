import React from 'react';
import { Breadcrumb, Empty, Link, Loading, Tag } from 'components/Fluent';
import { useNavigate } from 'components/Router';

import { describeRoles, Role } from 'services/role';
import { openErrNotification } from 'utils/notifition';
import PrincipalPolicyTable from './PrincipalPolicyTable';
import style from './index.module.less';

const { BreadcrumbItem } = Breadcrumb;
const formatValue = (value?: string) => value || '-';

const RoleDetailPage: React.FC = () => {
    const navigate = useNavigate();
    const searchParams = new URLSearchParams(window.location.search);
    const roleId = searchParams.get('id') || '';
    const roleName = searchParams.get('name') || '角色详情';
    const [state, setState] = React.useState<{
        loading: boolean;
        role: Role | null;
        fetchError: boolean;
    }>({ loading: false, role: null, fetchError: false });

    const loadRole = React.useCallback(async () => {
        if (!roleId) return;
        setState(prev => ({ ...prev, loading: true, fetchError: false }));
        try {
            const result = await describeRoles({ id: roleId, limit: 1, offset: 0, berif: false });
            const role = result.content[0] || null;
            setState({ loading: false, role, fetchError: !role });
        } catch (error) {
            setState({ loading: false, role: null, fetchError: true });
            openErrNotification('请求错误', `获取角色详情失败, ${(error as Error).message}`);
        }
    }, [roleId]);

    React.useEffect(() => {
        loadRole();
    }, [loadRole]);

    const role = state.role;
    const labels = Object.entries(role?.metadata || {});
    const users = role?.users || [];
    const groups = role?.user_groups || [];
    const initial = (role?.name || roleName || 'R').slice(0, 1).toUpperCase();

    return (
        <div className={style.userDetailPage}>
            <Breadcrumb className={style.detailBreadcrumb} maxItemWidth="240px">
                <BreadcrumbItem onClick={() => navigate('/auth/principals?tab=3')}>角色</BreadcrumbItem>
                <BreadcrumbItem>{role?.name || roleName}</BreadcrumbItem>
            </Breadcrumb>

            <Loading indicator loading={state.loading} preventScrollThrough showOverlay>
                {!roleId || state.fetchError || !role ? (
                    <section className={style.userDetailEmpty}>
                        <Empty description={roleId ? '未找到角色详情' : '缺少角色标识，无法加载详情'} />
                    </section>
                ) : (
                    <main className={style.userDetailStack}>
                        <section className={style.userDetailSummary}>
                            <div className={style.userDetailAvatar}>{initial}</div>
                            <div className={style.userDetailMain}>
                                <div className={style.userDetailTitleLine}>
                                    <h2>{role.name}</h2>
                                    <Tag theme={role.default_role ? 'primary' : 'success'} variant="light">
                                        {role.default_role ? '系统内置' : '自定义角色'}
                                    </Tag>
                                </div>
                                <p>{role.comment || '暂无备注'}</p>
                                <div className={style.userMetaGrid}>
                                    <div className={style.userMetaItem}>
                                        <span>角色 ID</span>
                                        <strong className={style.userMono} title={role.id}>{formatValue(role.id)}</strong>
                                    </div>
                                    <div className={style.userMetaItem}>
                                        <span>来源</span>
                                        <strong>{formatValue(role.source)}</strong>
                                    </div>
                                    <div className={style.userMetaItem}>
                                        <span>成员用户</span>
                                        <strong>{users.length} 个</strong>
                                    </div>
                                    <div className={style.userMetaItem}>
                                        <span>成员用户组</span>
                                        <strong>{groups.length} 个</strong>
                                    </div>
                                    <div className={style.userMetaItem}>
                                        <span>创建时间</span>
                                        <strong>{formatValue(role.ctime)}</strong>
                                    </div>
                                    <div className={style.userMetaItem}>
                                        <span>修改时间</span>
                                        <strong>{formatValue(role.mtime)}</strong>
                                    </div>
                                </div>
                            </div>
                        </section>

                        <div className={style.userDetailContentGrid}>
                            <section className={style.userCredentialPanel}>
                                <div className={style.userPanelHeader}>
                                    <div>
                                        <strong>成员用户</strong>
                                        <span>直接绑定到当前角色的用户</span>
                                    </div>
                                </div>
                                {users.length > 0 ? (
                                    <div className={style.userLabelList}>
                                        {users.map(user => (
                                            <Link
                                                key={user.id}
                                                theme="primary"
                                                onClick={() => navigate(`/auth/principals/userdetail?name=${encodeURIComponent(user.name || user.id)}&id=${encodeURIComponent(user.id)}`)}
                                            >
                                                {user.name || user.id}
                                            </Link>
                                        ))}
                                    </div>
                                ) : <p className={style.userPanelEmpty}>暂无成员用户</p>}
                            </section>

                            <section className={style.userTagPanel}>
                                <div className={style.userPanelHeader}>
                                    <div>
                                        <strong>成员用户组</strong>
                                        <span>通过用户组继承当前角色的主体</span>
                                    </div>
                                </div>
                                {groups.length > 0 ? (
                                    <div className={style.userLabelList}>
                                        {groups.map(group => (
                                            <Link
                                                key={group.id}
                                                theme="primary"
                                                onClick={() => navigate(`/auth/principals/groupdetail?name=${encodeURIComponent(group.name || group.id)}&id=${encodeURIComponent(group.id)}`)}
                                            >
                                                {group.name || group.id}
                                            </Link>
                                        ))}
                                    </div>
                                ) : <p className={style.userPanelEmpty}>暂无成员用户组</p>}
                            </section>
                        </div>

                        <section className={style.userTagPanel}>
                            <div className={style.userPanelHeader}>
                                <div>
                                    <strong>角色标签</strong>
                                    <span>用于角色检索和权限关系识别。</span>
                                </div>
                            </div>
                            {labels.length > 0 ? (
                                <div className={style.userLabelList}>
                                    {labels.map(([key, value]) => <Tag key={key} variant="outline">{key}: {value}</Tag>)}
                                </div>
                            ) : <p className={style.userPanelEmpty}>暂无标签</p>}
                        </section>

                        <section className={style.userPolicySurface}>
                            <div className={style.userTableHeader}>
                                <div>
                                    <strong>关联策略</strong>
                                    <span>当前角色关联的资源和 API 访问策略</span>
                                </div>
                            </div>
                            <PrincipalPolicyTable principalId={role.id} principalType={3} />
                        </section>
                    </main>
                )}
            </Loading>
        </div>
    );
};

export default React.memo(RoleDetailPage);
