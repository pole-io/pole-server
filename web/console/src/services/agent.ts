import { apiRequest, getApiRequest } from 'utils/request';

const AgentBaseURL = '/ai/agent/v1';

export interface AgentConfigFile {
  id?: string;
  namespace: string;
  group: string;
  name: string;
  content: string;
  format?: string;
  comment?: string;
  status?: string;
  labels?: Record<string, string>;
  mtime?: string;
  encrypted: boolean;
  encryptAlgo?: string;
}

export interface AgentResourceRef {
  kind: 'config.file';
  namespace: string;
  group: string;
  name: string;
}

export interface ConfigFileProposal {
  id: string;
  version: number;
  status: 'preview_ready' | 'applying' | 'waiting_for_publish';
  resource: AgentResourceRef;
  before: AgentConfigFile;
  after: AgentConfigFile;
  baselineHash: string;
  previewHash: string;
  expiresAt: string;
  warnings: string[];
}

export interface DraftReceipt {
  proposalId: string;
  status: 'waiting_for_publish';
  resource: AgentResourceRef;
  requestId?: string;
  detailUrl: string;
}

export interface PrepareConfigFileProposalRequest {
  namespace: string;
  group: string;
  name: string;
  desiredContent: string;
  comment?: string;
}

export interface AgentTurnMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface AgentTurnResourceContext {
  kind: 'config.file';
  namespace: string;
  group: string;
  name: string;
}

export interface AgentRuntimeStatus {
  ready: boolean;
  mode: 'llm' | 'unavailable';
  configured: boolean;
  model?: string;
  promptVersion: string;
  mcpConnected: boolean;
  tools: string[];
  reason?: string;
}

export interface AgentTurnToolTrace {
  name: string;
  summary: string;
  detail?: string;
  status: 'success' | 'failed';
}

export interface AgentTurnResponse {
  message: string;
  tools: AgentTurnToolTrace[];
  proposal?: ConfigFileProposal;
  runtime: AgentRuntimeStatus;
}

export interface SendAgentTurnRequest {
  sessionId: string;
  message: string;
  history: AgentTurnMessage[];
  resourceContext?: AgentTurnResourceContext;
}

export async function prepareConfigFileProposal(params: PrepareConfigFileProposalRequest) {
  return apiRequest<ConfigFileProposal>({
    action: `${AgentBaseURL}/proposals/config-file`,
    data: params,
  });
}

export async function confirmAgentProposal(proposalId: string, proposalVersion: number, previewHash: string, idempotencyKey: string) {
  return apiRequest<DraftReceipt>({
    action: `${AgentBaseURL}/proposals/${encodeURIComponent(proposalId)}/confirm`,
    data: { proposalVersion, previewHash, idempotencyKey },
  });
}

export async function getAgentRuntime() {
  return getApiRequest<AgentRuntimeStatus>({
    action: `${AgentBaseURL}/runtime`,
    opts: { timeout: 10_000 },
  });
}

export async function sendAgentTurn(params: SendAgentTurnRequest) {
  return apiRequest<AgentTurnResponse>({
    action: `${AgentBaseURL}/turns`,
    data: params,
    opts: { timeout: 70_000 },
  });
}
