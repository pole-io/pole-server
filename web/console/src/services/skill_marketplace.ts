import { apiRequest, getApiRequest, putApiRequest } from 'utils/request';

/**
 * Skill Marketplace 的独立 API 根。控制面实现可在这一处统一调整前缀，页面不散落 URL。
 */
export const SkillMarketplaceAPI = '/api/skill-marketplace';

export type SkillVisibility = 'private' | 'public';
export type SkillReleaseStatus = 'draft' | 'pending_review' | 'published' | 'rejected' | 'yanked' | 'deprecated';
export type RegistryTrustLevel = 'trusted' | 'untrusted';

export interface SkillRelease {
  version: string;
  status: SkillReleaseStatus | string;
  digest: string;
  published_at?: string;
  created_at?: string;
  signature_status?: string;
  scan_status?: string;
  source?: string;
  source_url?: string;
  deprecation_message?: string;
  yanked?: boolean;
}

export interface MarketplaceSkill {
  publisher: string;
  name: string;
  display_name?: string;
  description?: string;
  visibility: SkillVisibility | string;
  status?: SkillReleaseStatus | string;
  latest_version?: string;
  latest_release?: SkillRelease;
  source?: string;
  source_url?: string;
  updated_at?: string;
  created_at?: string;
  tags?: string[];
}

export interface SkillDetail extends MarketplaceSkill {
  releases?: SkillRelease[];
  publisher_key_id?: string;
  publisher_key_status?: string;
  can_manage_grants?: boolean;
}

export type SkillGrantPrincipalType = 'user' | 'group' | 'role';

export interface SkillGrant {
  principalType: SkillGrantPrincipalType;
  principalId: string;
}

export interface SkillBundleEntry {
  path: string;
  type: 'file' | 'directory' | 'binary' | string;
  size?: number;
  mode?: string;
  executable?: boolean;
  text?: string;
  encoding?: string;
}

export interface SkillBundle {
  publisher: string;
  name: string;
  version: string;
  digest: string;
  entries: SkillBundleEntry[];
}

export interface RegistrySource {
  id: string;
  name: string;
  url: string;
  type: 'pole' | 'git' | 'http_index' | string;
  trust_level: RegistryTrustLevel | string;
  sync_status?: string;
  last_synced_at?: string;
  enabled?: boolean;
  error_message?: string;
}

export interface MarketplaceReview {
  id: string;
  publisher: string;
  name: string;
  version: string;
  status: string;
  requested_at?: string;
  signature_status?: string;
  scan_status?: string;
}

export interface DescribeSkillsRequest {
  offset?: number;
  limit?: number;
  query?: string;
  source?: string;
  visibility?: string;
  status?: string;
}

const asArray = <T>(value: unknown): T[] => Array.isArray(value) ? value as T[] : [];

const normalizeRelease = (value: Record<string, any> = {}): SkillRelease => ({
  version: String(value.version || ''),
  status: value.status || 'draft',
  digest: String(value.digest || ''),
  published_at: value.published_at || value.publishedAt,
  created_at: value.created_at || value.createdAt,
  signature_status: value.signature_status || value.signatureStatus,
  scan_status: value.scan_status || value.scanStatus,
  source: value.source,
  source_url: value.source_url || value.sourceURL,
  deprecation_message: value.deprecation_message || value.deprecationMessage,
  yanked: value.yanked,
});

const normalizeSkill = (value: Record<string, any> = {}): MarketplaceSkill => ({
  publisher: String(value.publisher || ''),
  name: String(value.name || ''),
  display_name: value.display_name || value.displayName,
  description: value.description,
  visibility: value.visibility || 'private',
  status: value.status,
  latest_version: value.latest_version || value.latestVersion,
  latest_release: value.latest_release || value.latestRelease ? normalizeRelease(value.latest_release || value.latestRelease) : undefined,
  source: value.source,
  source_url: value.source_url || value.sourceURL,
  updated_at: value.updated_at || value.updatedAt,
  created_at: value.created_at || value.createdAt,
  tags: asArray<string>(value.tags),
});

const normalizeBundleEntry = (value: Record<string, any> = {}): SkillBundleEntry => ({
  path: String(value.path || ''),
  type: value.type || value.entryType || 'file',
  size: Number(value.size || 0),
  mode: value.mode,
  executable: Boolean(value.executable ?? value.isExecutable),
  text: value.text ?? value.content,
  encoding: value.encoding,
});

const normalizeRegistrySource = (value: Record<string, any> = {}): RegistrySource => ({
  id: String(value.id || ''),
  name: String(value.name || ''),
  url: String(value.url || ''),
  type: value.type || 'http_index',
  trust_level: value.trust_level || value.trustLevel || 'untrusted',
  sync_status: value.sync_status || value.syncStatus,
  last_synced_at: value.last_synced_at || value.lastSyncedAt,
  enabled: value.enabled,
  error_message: value.error_message || value.errorMessage,
});

const normalizeReview = (value: Record<string, any> = {}): MarketplaceReview => ({
  id: String(value.id || ''),
  publisher: String(value.publisher || ''),
  name: String(value.name || ''),
  version: String(value.version || ''),
  status: String(value.status || ''),
  requested_at: value.requested_at || value.requestedAt,
  signature_status: value.signature_status || value.signatureStatus,
  scan_status: value.scan_status || value.scanStatus,
});

/** GET /v1/skills；后端可返回 {items,total}、{data,amount} 或裸数组。 */
export async function describeMarketplaceSkills(params: DescribeSkillsRequest = {}) {
  const response = await getApiRequest<any>({ action: `${SkillMarketplaceAPI}/v1/skills`, data: params });
  const items = asArray<Record<string, any>>(response?.items || response?.data || response).map(normalizeSkill);
  return { items, total: Number(response?.total ?? response?.amount ?? items.length) };
}

/** GET /v1/skills/{publisher}/{name}；返回 Skill 与不可变 Release 列表。 */
export async function describeMarketplaceSkill(publisher: string, name: string): Promise<SkillDetail> {
  const response = await getApiRequest<any>({
    action: `${SkillMarketplaceAPI}/v1/skills/${encodeURIComponent(publisher)}/${encodeURIComponent(name)}`,
  });
  const skill = normalizeSkill(response?.skill || response);
  return {
    ...skill,
    releases: asArray<Record<string, any>>(response?.releases || response?.skill?.releases).map(normalizeRelease),
    publisher_key_id: response?.publisher_key_id || response?.publisherKeyId || response?.skill?.publisher_key_id || response?.skill?.publisherKeyId,
    publisher_key_status: response?.publisher_key_status || response?.publisherKeyStatus || response?.skill?.publisher_key_status || response?.skill?.publisherKeyStatus,
    can_manage_grants: response?.can_manage_grants ?? response?.canManageGrants ?? response?.skill?.can_manage_grants ?? response?.skill?.canManageGrants,
  };
}

/** GET 精确 Release manifest；/bundle 是 raw ZIP，不能作为 JSON 读取。 */
export async function describeMarketplaceSkillBundle(publisher: string, name: string, version: string): Promise<SkillBundle> {
  const response = await getApiRequest<any>({
    action: `${SkillMarketplaceAPI}/v1/skills/${encodeURIComponent(publisher)}/${encodeURIComponent(name)}/releases/${encodeURIComponent(version)}/bundle-manifest`,
  });
  const bundle = response?.manifest || response || {};
  return {
    publisher: String(bundle.publisher || publisher),
    name: String(bundle.name || name),
    version: String(bundle.version || version),
    digest: String(bundle.digest || ''),
    entries: asArray<Record<string, any>>(bundle.entries || bundle.files).map(normalizeBundleEntry),
  };
}

/** POST multipart /v1/skills/releases/upload，signature 是由服务端解析的 detached JSON envelope。 */
export async function uploadMarketplaceSkillRelease(input: {
  publisher: string;
  name: string;
  version: string;
  visibility: SkillVisibility;
  bundle: File;
  signature?: File;
}) {
  const form = new FormData();
  form.append('publisher', input.publisher);
  form.append('name', input.name);
  form.append('version', input.version);
  form.append('visibility', input.visibility);
  form.append('bundle', input.bundle);
  if (input.signature) form.append('signature', input.signature);
  return apiRequest({ action: `${SkillMarketplaceAPI}/v1/skills/releases/upload`, data: form });
}

/** POST /v1/skills/releases/import-git，由服务端从 Tag 或 Release 拉取并冻结内容。 */
export async function importMarketplaceSkillFromGit(input: {
  repository_url: string;
  reference: string;
  publisher: string;
  name: string;
  visibility: SkillVisibility;
}) {
  return apiRequest({ action: `${SkillMarketplaceAPI}/v1/skills/releases/import-git`, data: input });
}

export async function describeMarketplaceReviews() {
  const response = await getApiRequest<any>({ action: `${SkillMarketplaceAPI}/v1/reviews`, data: { status: 'pending_review' } });
  return asArray<Record<string, any>>(response?.items || response?.data || response).map(normalizeReview);
}

/** PUT /v1/reviews/{releaseId}，决定不修改已发布 Bundle，仅改变审核状态。 */
export async function decideMarketplaceReview(releaseId: string, decision: 'approved' | 'rejected', comment = '') {
  return putApiRequest({
    action: `${SkillMarketplaceAPI}/v1/reviews/${encodeURIComponent(releaseId)}`,
    data: { decision, comment },
  });
}

export async function describeRegistrySources() {
  const response = await getApiRequest<any>({ action: `${SkillMarketplaceAPI}/v1/registry-sources` });
  return asArray<Record<string, any>>(response?.items || response?.data || response).map(normalizeRegistrySource);
}

export async function createRegistrySource(input: Omit<RegistrySource, 'id' | 'sync_status' | 'last_synced_at' | 'error_message'>) {
  return apiRequest({ action: `${SkillMarketplaceAPI}/v1/registry-sources`, data: input });
}

export async function syncRegistrySource(id: string) {
  return apiRequest({ action: `${SkillMarketplaceAPI}/v1/registry-sources/${encodeURIComponent(id)}/sync`, data: {} });
}

const normalizeGrant = (value: Record<string, any> = {}): SkillGrant => ({
  principalType: value.principalType || value.principal_type || 'user',
  principalId: String(value.principalId || value.principal_id || ''),
});

/** 私有 Skill 的访问策略；控制面负责判定当前用户是否具备管理权限。 */
export async function describeMarketplaceSkillGrants(publisher: string, name: string): Promise<SkillGrant[]> {
  const response = await getApiRequest<any>({
    action: `${SkillMarketplaceAPI}/v1/skills/${encodeURIComponent(publisher)}/${encodeURIComponent(name)}/grants`,
  });
  return asArray<Record<string, any>>(response?.grants || response?.items || response?.data || response).map(normalizeGrant);
}

/** PUT 整体替换私有 Skill 的授权主体集合，payload 保持后端 camelCase 契约。 */
export async function updateMarketplaceSkillGrants(publisher: string, name: string, grants: SkillGrant[]) {
  return putApiRequest({
    action: `${SkillMarketplaceAPI}/v1/skills/${encodeURIComponent(publisher)}/${encodeURIComponent(name)}/grants`,
    data: { grants: grants.map(({ principalType, principalId }) => ({ principalType, principalId })) },
  });
}
