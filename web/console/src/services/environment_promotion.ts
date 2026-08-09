import { apiRequest, getApiRequest, putApiRequest } from 'utils/request';

const TOPOLOGY_URL = '/core/v1/environment-promotion/topology';

export interface EnvironmentPromotionEdge {
  id: string;
  source: string;
  target: string;
  require_formal_release: boolean;
  require_approval: boolean;
  validation_gates: string[];
  allowed_resource_domains: string[];
  conflict_policy: 'BLOCK';
}

export interface LaneBaseBinding {
  lane: string;
  base: string;
}

export interface EnvironmentPromotionTopology {
  draft_revision: number;
  published_revision: number;
  edges: EnvironmentPromotionEdge[];
  lane_base_bindings: LaneBaseBinding[];
  modify_by?: string;
  modify_time?: string;
}

export interface TopologyValidationIssue {
  code: string;
  path?: string;
  message: string;
}

const normalizeTopology = (raw?: Partial<EnvironmentPromotionTopology>): EnvironmentPromotionTopology => ({
  draft_revision: Number(raw?.draft_revision || 0),
  published_revision: Number(raw?.published_revision || 0),
  edges: raw?.edges || [],
  lane_base_bindings: raw?.lane_base_bindings || [],
  modify_by: raw?.modify_by,
  modify_time: raw?.modify_time,
});

export async function describeEnvironmentPromotionTopology() {
  return normalizeTopology(await getApiRequest<EnvironmentPromotionTopology>({ action: TOPOLOGY_URL }));
}

export async function validateEnvironmentPromotionTopology(topology: EnvironmentPromotionTopology) {
  return apiRequest<{ valid: boolean; issues: TopologyValidationIssue[] }>({
    action: `${TOPOLOGY_URL}/validate`,
    data: topology,
  });
}

export async function saveEnvironmentPromotionTopology(topology: EnvironmentPromotionTopology) {
  return normalizeTopology(await putApiRequest<EnvironmentPromotionTopology>({
    action: TOPOLOGY_URL,
    data: topology,
  }));
}

export async function publishEnvironmentPromotionTopology(expectedRevision: number, comment: string) {
  return apiRequest({
    action: `${TOPOLOGY_URL}/publish`,
    data: { expected_revision: expectedRevision, comment },
  });
}
