import React from 'react';
import { Button, Empty, Input, Loading, Select, Tag } from 'components/Fluent';
import EditorDrawer, { EditorDrawerActions } from 'components/EditorDrawer';
import { useLocation, useNavigate } from 'components/Router';
import { ResourceHeader } from 'components/ResourceLayout';
import {
  SkillBundle,
  SkillBundleEntry,
  SkillDetail,
  SkillGrant,
  SkillGrantPrincipalType,
  SkillRelease,
  describeMarketplaceSkillGrants,
  describeMarketplaceSkill,
  describeMarketplaceSkillBundle,
  updateMarketplaceSkillGrants,
} from 'services/skill_marketplace';
import style from './index.module.less';

const parseIdentity = (pathname: string) => {
  const parts = pathname.split('/').filter(Boolean);
  const anchor = parts.indexOf('skills');
  return { publisher: anchor >= 0 ? decodeURIComponent(parts[anchor + 1] || '') : '', name: anchor >= 0 ? decodeURIComponent(parts[anchor + 2] || '') : '' };
};

const statusTheme = (status?: string) => status === 'published' ? 'success' : status === 'pending_review' ? 'warning' : status === 'yanked' ? 'danger' : 'default';
const sourceText = (source?: string) => ({ pole: 'Pole 本地', git: 'Git', http_index: 'HTTP Registry' }[source || ''] || source || '未标注');
const releaseText = (release?: SkillRelease) => release?.version || '未发布';
const selectEntry = (entries: SkillBundleEntry[]) => entries.find((entry) => entry.path === 'SKILL.md') || entries.find((entry) => entry.type === 'file' && typeof entry.text === 'string') || entries[0] || null;
const principalTypeOptions = [{ label: 'User', value: 'user' }, { label: 'UserGroup', value: 'group' }, { label: 'Role', value: 'role' }];

const SkillMarketplaceDetailPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { publisher, name } = parseIdentity(location.pathname);
  const [skill, setSkill] = React.useState<SkillDetail | null>(null);
  const [version, setVersion] = React.useState('');
  const [bundle, setBundle] = React.useState<SkillBundle | null>(null);
  const [entry, setEntry] = React.useState<SkillBundleEntry | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [bundleLoading, setBundleLoading] = React.useState(false);
  const [error, setError] = React.useState('');
  const [copyStatus, setCopyStatus] = React.useState('');
  const [policyVisible, setPolicyVisible] = React.useState(false);
  const [grants, setGrants] = React.useState<SkillGrant[]>([]);
  const [grantsLoading, setGrantsLoading] = React.useState(false);
  const [grantsSaving, setGrantsSaving] = React.useState(false);
  const [grantPrincipalType, setGrantPrincipalType] = React.useState<SkillGrantPrincipalType>('user');
  const [grantPrincipalId, setGrantPrincipalId] = React.useState('');
  const [grantStatus, setGrantStatus] = React.useState('');

  const loadSkill = React.useCallback(async () => {
    if (!publisher || !name) { setError('Skill 路径不完整。'); setLoading(false); return; }
    setLoading(true); setError('');
    try {
      const next = await describeMarketplaceSkill(publisher, name);
      setSkill(next);
      const nextVersion = next.latest_version || next.latest_release?.version || next.releases?.[0]?.version || '';
      setVersion(nextVersion);
    } catch (reason) { setSkill(null); setError(reason instanceof Error ? reason.message : 'Skill 详情查询失败。'); }
    finally { setLoading(false); }
  }, [name, publisher]);

  const loadBundle = React.useCallback(async (selectedVersion: string) => {
    if (!selectedVersion || !publisher || !name) { setBundle(null); setEntry(null); return; }
    setBundleLoading(true);
    try {
      const next = await describeMarketplaceSkillBundle(publisher, name, selectedVersion);
      setBundle(next); setEntry(selectEntry(next.entries));
    } catch { setBundle(null); setEntry(null); }
    finally { setBundleLoading(false); }
  }, [name, publisher]);

  React.useEffect(() => { void loadSkill(); }, [loadSkill]);
  React.useEffect(() => { if (version) void loadBundle(version); }, [loadBundle, version]);

  const selectedRelease = skill?.releases?.find((item) => item.version === version) || skill?.latest_release;
  const command = `pole-ai skill install ${publisher}/${name}@${version || '<version>'}`;
  const digest = selectedRelease?.digest || bundle?.digest || '';
  const copyCommand = async () => {
    try {
      if (!navigator.clipboard?.writeText) throw new Error('当前浏览器不支持 Clipboard API');
      await navigator.clipboard.writeText(command);
      setCopyStatus('安装命令已复制到剪贴板。');
    } catch {
      setCopyStatus('复制失败，请手动选择并复制安装命令。');
    }
  };
  const openAccessPolicy = async () => {
    setPolicyVisible(true);
    setGrantsLoading(true);
    setGrantStatus('');
    try {
      setGrants(await describeMarketplaceSkillGrants(publisher, name));
    } catch (reason) {
      setGrants([]);
      setGrantStatus(reason instanceof Error ? reason.message : '访问策略加载失败。');
    } finally { setGrantsLoading(false); }
  };
  const addGrant = () => {
    const principalId = grantPrincipalId.trim();
    if (!principalId) { setGrantStatus('请填写授权主体 ID。'); return; }
    if (grants.some((item) => item.principalType === grantPrincipalType && item.principalId === principalId)) {
      setGrantStatus('该授权主体已存在。');
      return;
    }
    setGrants((current) => [...current, { principalType: grantPrincipalType, principalId }]);
    setGrantPrincipalId('');
    setGrantStatus('');
  };
  const saveAccessPolicy = async () => {
    setGrantsSaving(true);
    setGrantStatus('');
    try {
      await updateMarketplaceSkillGrants(publisher, name, grants);
      setGrantStatus('访问策略已保存。');
    } catch (reason) {
      setGrantStatus(reason instanceof Error ? reason.message : '访问策略保存失败。');
    } finally { setGrantsSaving(false); }
  };

  return (
    <div className={style.page}>
      <ResourceHeader placement="app-header" eyebrow="AI 工具 / Skill Marketplace / 不可变 Release" title={skill ? `${skill.publisher}/${skill.name}` : `${publisher}/${name}`} description="查看已冻结 Release 的元数据和安全 Bundle 文件树；不会执行、安装或预览二进制内容。" actions={<Button variant="outline" onClick={() => navigate('/ai/skills')}>返回目录</Button>} />
      <Loading loading={loading} text="加载 Skill 详情…">
        {error ? <Empty title="Skill 详情不可用" description={error} action={<Button variant="outline" onClick={() => void loadSkill()}>重试</Button>} /> : skill && <>
          <section className={style.identity}>
            <div><h2>{skill.display_name || skill.name}</h2><p>{skill.description || '该 Skill 尚未提供描述。'}</p><div className={style.tagRow}><Tag theme={skill.visibility === 'public' ? 'success' : 'default'} variant="outline">{skill.visibility === 'public' ? '公共' : '私有'}</Tag><Tag variant="outline">{sourceText(skill.source)}</Tag>{(skill.tags || []).map((tag) => <Tag key={tag} variant="outline">{tag}</Tag>)}</div></div>
            <dl><div><dt>Publisher</dt><dd>{skill.publisher}</dd></div><div><dt>Publisher Key</dt><dd>{skill.publisher_key_id || '未提供'} · {skill.publisher_key_status || '未知'}</dd></div><div><dt>来源</dt><dd>{skill.source_url || skill.source || '-'}</dd></div></dl>
          </section>
          <section className={style.releaseControl} aria-label="Release 选择">
            <div><strong>冻结 Release</strong><span>发布后版本、Bundle 和 SHA-256 不可编辑；撤回不会覆盖已有 lockfile。</span></div>
            <Select aria-label="选择 Skill Release" className={style.versionSelect} options={(skill.releases || []).map((release) => ({ label: `${release.version} · ${release.status}`, value: release.version }))} value={version} onChange={setVersion} />
          </section>
          <section className={style.securityGrid} aria-label="Release 安全状态">
            <div><span>Release 状态</span><Tag theme={statusTheme(selectedRelease?.status)} variant="outline">{selectedRelease?.status || '-'}</Tag></div><div><span>签名</span><strong>{selectedRelease?.signature_status || '未提供'}</strong></div><div><span>安全扫描</span><strong>{selectedRelease?.scan_status || '未提供'}</strong></div><div><span>来源冻结时间</span><strong>{selectedRelease?.published_at || selectedRelease?.created_at || '-'}</strong></div>
          </section>
          <section className={style.accessPolicy} aria-label="Skill 访问策略"><div><strong>访问策略</strong><span>{skill.visibility === 'public' ? '公共 Skill 无需私有授权。' : '私有 Skill 仅允许 Publisher Owner 和显式授权的 User、UserGroup、Role 读取。'}</span></div>{skill.visibility !== 'public' && (skill.can_manage_grants === false ? <small>当前账号没有管理访问策略的权限。</small> : <Button size="small" variant="outline" onClick={() => void openAccessPolicy()}>管理访问策略</Button>)}</section>
          <section className={style.command} aria-label="精确安装命令"><div><strong>pole-ai 精确命令</strong><span>复制到受控开发环境执行；Release 摘要由 lockfile 固定，Console 不提供执行或安装按钮。</span></div><div className={style.commandValue}><code>{command}</code><Button size="small" variant="outline" aria-label="复制 pole-ai 安装命令" onClick={() => void copyCommand()}>复制命令</Button></div><span className={style.copyStatus} role="status" aria-live="polite">{copyStatus}</span></section>
          <section className={style.bundleWorkspace} aria-label="Bundle 安全浏览器">
            <header><div><strong>Bundle 文件树</strong><span>{digest ? `SHA-256 ${digest}` : 'Release 摘要未返回'}</span></div><Tag variant="outline">{releaseText(selectedRelease)}</Tag></header>
            <Loading loading={bundleLoading} text="读取冻结 Bundle…"><div className={style.bundleBody}>
              <nav className={style.fileTree} aria-label="Bundle 文件列表">{bundle?.entries?.length ? bundle.entries.map((item) => <button key={item.path} type="button" className={entry?.path === item.path ? style.fileSelected : ''} onClick={() => setEntry(item)}><span>{item.type === 'directory' ? '▸' : item.executable ? '›' : '·'} {item.path}</span><small>{item.type === 'directory' ? '目录' : `${item.size || 0} B`}</small></button>) : <Empty title="该 Release 未返回可浏览文件" description="Bundle 元数据需要由控制面安全解析后返回。" />}</nav>
              <article className={style.filePreview}>{entry ? <><header><strong>{entry.path}</strong><span>{entry.type} · {entry.mode || '-'}{entry.executable ? ' · executable' : ''}</span></header>{entry.type === 'directory' ? <Empty title="目录" description="选择一个文件以查看已验证的文本内容。" /> : entry.type === 'binary' || entry.encoding === 'binary' ? <Empty title="二进制文件不预览" description="为避免执行或渲染不受信内容，Console 仅显示元数据。" /> : typeof entry.text === 'string' ? <pre>{entry.text}</pre> : <Empty title="文本内容未返回" description="服务端可按安全文本限制返回此冻结文件的内容。" />}</> : <Empty title="选择 Bundle 文件" description="默认优先显示 SKILL.md。" />}</article>
            </div></Loading>
          </section>
          <EditorDrawer visible={policyVisible} width="standard" title="私有 Skill 访问策略" description="只管理 User、UserGroup 与 Role 的读取授权；公共 Skill 不使用此策略。" onClose={() => setPolicyVisible(false)} footer={<EditorDrawerActions onCancel={() => setPolicyVisible(false)} onSubmit={() => void saveAccessPolicy()} submitText="保存策略" submitting={grantsSaving} />}>
            <div className={style.grantDrawer}>
              <div className={style.grantAdd}><Select aria-label="授权主体类型" options={principalTypeOptions} value={grantPrincipalType} onChange={(value: SkillGrantPrincipalType) => setGrantPrincipalType(value)} /><Input aria-label="授权主体 ID" value={grantPrincipalId} placeholder="User、UserGroup 或 Role ID" onChange={setGrantPrincipalId} onEnter={addGrant} /><Button variant="outline" onClick={addGrant}>添加</Button></div>
              <Loading loading={grantsLoading} text="加载访问策略…"><div className={style.grantList}>{grants.length ? grants.map((grant) => <div className={style.grantRow} key={`${grant.principalType}:${grant.principalId}`}><Tag variant="outline">{grant.principalType}</Tag><code>{grant.principalId}</code><Button size="small" variant="outline" aria-label={`移除 ${grant.principalType} ${grant.principalId}`} onClick={() => setGrants((current) => current.filter((item) => item.principalType !== grant.principalType || item.principalId !== grant.principalId))}>移除</Button></div>) : <Empty title="尚无额外授权主体" description="Publisher Owner 始终保有私有 Skill 的管理能力。" />}</div></Loading>
              <span className={style.copyStatus} role="status" aria-live="polite">{grantStatus}</span>
            </div>
          </EditorDrawer>
        </>}
      </Loading>
    </div>
  );
};

export default SkillMarketplaceDetailPage;
