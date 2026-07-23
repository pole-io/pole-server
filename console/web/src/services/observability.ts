import { getApiRequest } from 'utils/request';
import type { DescribeEventLogRequest, DescribeEventLogResponse, DescribeOperationLogRequest, DescribeOperationLogResponse } from './observer';

export interface PlatformComponentMetric {
  id: string;
  category: string;
  component: string;
  role: string;
  namespace: string;
  pod: string;
  api: string;
  status: string;
  cpu: number;
  memory: number;
  qps: number;
  p95: number;
  p99: number;
  errorRate: number;
  restartCount: number;
  series: number[];
}

export interface PlatformResourceMetric {
  id: string;
  component: string;
  namespace: string;
  pod: string;
  cpu: number;
  memory: number;
  memoryUnit: string;
  restartCount: number;
  cpuSeries: number[];
  memorySeries: number[];
}

export interface PlatformRuntimeMetric {
  name: string;
  title: string;
  category: string;
  value: number;
  unit: string;
  description: string;
  series: number[];
}

export interface PlatformOverview {
  provider: {
    provider: string;
    endpoint?: string;
    database?: string;
    configured: boolean;
  };
  stats: Array<{
    name: string;
    value: number;
    unit?: string;
    labels?: Record<string, string>;
  }>;
  series: Array<{
    name: string;
    unit?: string;
    labels?: Record<string, string>;
    points: Array<{ timestamp: number; value: number }>;
  }>;
  components: PlatformComponentMetric[];
  resources: PlatformResourceMetric[];
  runtime: PlatformRuntimeMetric[];
}

export interface DescribePlatformOverviewRequest {
  start_time?: string;
  end_time?: string;
  step?: string;
  category?: string;
  api?: string;
}

export function describePlatformOverview(params: DescribePlatformOverviewRequest) {
  return getApiRequest<PlatformOverview>({
    action: '/observability/v1/platform/overview',
    data: params,
  });
}

export function describeObservabilityEvents(params: DescribeEventLogRequest) {
  return getApiRequest<DescribeEventLogResponse>({
    action: '/observability/v1/events',
    data: params,
  });
}

export function describeObservabilityOperations(params: DescribeOperationLogRequest) {
  return getApiRequest<DescribeOperationLogResponse>({
    action: '/observability/v1/operations',
    data: params,
  });
}
