import { apiRequest, getApiRequest, putApiRequest } from 'utils/request';

export type SettingComponent = 'pole-server' | 'pole-console';
export type SettingValueType = 'string' | 'boolean' | 'integer' | 'duration' | 'list' | 'secret';
export type SettingApplyMode = 'BootstrapOnly' | 'RestartRequired' | 'HotReload' | 'GuardedHotReload';
export type SettingSensitivity = 'public' | 'internal' | 'secret';
export type SettingSourceKind =
  | 'compiled_default'
  | 'static_file'
  | 'environment'
  | 'command_line'
  | 'dynamic_release';

export interface SettingSource {
  kind: SettingSourceKind;
  reference?: string;
}

export interface EffectiveSystemSetting {
  key: string;
  component: SettingComponent;
  domain: string;
  label: string;
  description?: string;
  value_type: SettingValueType;
  apply_mode: SettingApplyMode;
  sensitivity: SettingSensitivity;
  owner: string;
  value?: unknown;
  display_value: string;
  configured: boolean;
  redacted: boolean;
  source: SettingSource;
  editable: boolean;
  edit_reason?: string;
  validation: {
    required?: boolean;
    format?: string;
    options?: string[];
    min_integer?: number;
    max_integer?: number;
    min_duration?: string;
    max_duration?: string;
  };
  desired_value?: unknown;
  desired_display_value?: string;
  desired_revision?: number;
  drifted?: boolean;
  apply_status?: string;
}

export interface SystemConfigurationComponent {
  name: SettingComponent;
  count: number;
}

export interface EffectiveSystemSettingsResponse {
  settings: EffectiveSystemSetting[];
  components: SystemConfigurationComponent[];
}

export function describeEffectiveSystemSettings(params?: { component?: SettingComponent; domain?: string }) {
  return getApiRequest<EffectiveSystemSettingsResponse>({
    action: '/system-config/v1/settings',
    data: params,
  });
}

export interface AgentSystemProfile {
  runtimeMode: string;
  agentId: string;
  promptVersion: string;
  operatorInstructions: string;
  provider: string;
  baseURL: string;
  model: string;
  modelTimeout: string;
  mcpEndpoint: string;
  mcpToolAllowlist: string[];
  proposalTTL: string;
  upstreamTimeout: string;
}

export interface SystemSecretMetadata {
  configured: boolean;
  version?: number;
  reference?: string;
  fingerprint?: string;
  rotatedAt?: string;
}

export interface AgentSystemRevision {
  revision: number;
  state: 'draft' | 'published';
  values: AgentSystemProfile;
  secret: SystemSecretMetadata;
  createdBy: string;
  publishedBy?: string;
  createdAt: string;
  publishedAt?: string;
}

export interface AgentSystemDomain {
  component: 'pole-console';
  domain: 'agent';
  active?: AgentSystemRevision;
  draft?: AgentSystemRevision;
  effectiveRevision: string;
  applyStatus: 'static' | 'applied' | 'rejected';
  applyMessage?: string;
  secretStoreReady: boolean;
}

export interface SaveAgentDraftRequest {
  expectedDraftRevision: number;
  values: AgentSystemProfile;
  apiKey?: { operation: 'keep' | 'replace' | 'disable'; value?: string };
}

const AGENT_DOMAIN_PATH = '/system-config/v1/domains/pole-console/agent';

export function getAgentSystemDomain() {
  return getApiRequest<AgentSystemDomain>({ action: `${AGENT_DOMAIN_PATH}/draft` });
}

export function saveAgentSystemDraft(data: SaveAgentDraftRequest) {
  return putApiRequest<AgentSystemDomain>({ action: `${AGENT_DOMAIN_PATH}/draft`, data });
}

export function testAgentSystemConnection(data: SaveAgentDraftRequest) {
  return apiRequest<{ success: boolean; latencyMs: number; model?: string; requestId?: string; message: string }>({
    action: `${AGENT_DOMAIN_PATH}/connection-test`,
    data,
    opts: { timeout: 70000 },
  });
}

export function publishAgentSystemDraft(data: { draftRevision: number; description?: string }) {
  return apiRequest<AgentSystemDomain>({ action: `${AGENT_DOMAIN_PATH}/publish`, data, opts: { timeout: 70000 } });
}

export function listAgentSystemReleases() {
  return getApiRequest<{ data: AgentSystemRevision[] }>({ action: `${AGENT_DOMAIN_PATH}/releases` });
}

export interface ManagedSystemRevision {
  revision: number;
  state: 'draft' | 'published';
  values: Record<string, unknown>;
  createdBy: string;
  publishedBy?: string;
  createdAt: string;
  publishedAt?: string;
}

export interface ManagedSystemDomain {
  component: SettingComponent;
  domain: string;
  active?: ManagedSystemRevision;
  draft?: ManagedSystemRevision;
  applyStatus: 'static' | 'pending_restart' | 'applied' | 'rejected';
  applyMessage?: string;
}

function managedDomainPath(component: SettingComponent, domain: string) {
  return `/system-config/v1/domains/${encodeURIComponent(component)}/${encodeURIComponent(domain)}`;
}

export function getManagedSystemDomain(component: SettingComponent, domain: string) {
  return getApiRequest<ManagedSystemDomain>({ action: `${managedDomainPath(component, domain)}/draft` });
}

export function saveManagedSystemDraft(
  component: SettingComponent,
  domain: string,
  data: { expectedDraftRevision: number; values: Record<string, unknown> },
) {
  return putApiRequest<ManagedSystemDomain>({ action: `${managedDomainPath(component, domain)}/draft`, data });
}

export function publishManagedSystemDraft(
  component: SettingComponent,
  domain: string,
  data: { draftRevision: number; description?: string },
) {
  return apiRequest<ManagedSystemDomain>({ action: `${managedDomainPath(component, domain)}/publish`, data });
}

export function listManagedSystemReleases(component: SettingComponent, domain: string) {
  return getApiRequest<ManagedSystemRevision[]>({ action: `${managedDomainPath(component, domain)}/releases` });
}
