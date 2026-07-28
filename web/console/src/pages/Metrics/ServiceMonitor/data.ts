export type GovernanceKind = 'ratelimit' | 'route' | 'circuitbreaker' | 'faultdetect' | 'auth' | 'mock' | 'mirror';

export type ServiceSignal = {
  id: string;
  kind: GovernanceKind;
  label: string;
  namespace: string;
  service: string;
  api: string;
  instance: string;
  language: 'java' | 'go' | 'rust';
  requests: number;
  hitRate: number;
  blocked: number;
  p95: number;
  p99: number;
  errorRate: number;
  successRate: number;
  cpu: number;
  memory: number;
  status: 'normal' | 'warning' | 'critical';
  series: number[];
  [key: string]: unknown;
};

export const GOVERNANCE_LABEL: Record<GovernanceKind, string> = {
  ratelimit: '限流',
  route: '路由',
  circuitbreaker: '熔断',
  faultdetect: '探测',
  auth: '鉴权',
  mock: 'Mock',
  mirror: '镜像',
};

export const SERVICE_SIGNALS: ServiceSignal[] = [
  {
    id: 'checkout-ratelimit-1',
    kind: 'ratelimit',
    label: 'checkout-submit-limit',
    namespace: 'default',
    service: 'checkout',
    api: 'POST /api/checkout/submit',
    instance: '10.24.8.31:8080',
    language: 'java',
    requests: 182400,
    hitRate: 12.4,
    blocked: 1620,
    p95: 62,
    p99: 145,
    errorRate: 0.18,
    successRate: 99.82,
    cpu: 58,
    memory: 64,
    status: 'normal',
    series: [118, 132, 144, 156, 174, 182, 168, 176, 188, 181, 172, 182],
  },
  {
    id: 'checkout-route-1',
    kind: 'route',
    label: 'checkout-gray-route',
    namespace: 'default',
    service: 'checkout',
    api: 'GET /api/checkout/detail',
    instance: '10.24.8.33:8080',
    language: 'java',
    requests: 220600,
    hitRate: 35,
    blocked: 0,
    p95: 74,
    p99: 166,
    errorRate: 0.22,
    successRate: 99.78,
    cpu: 63,
    memory: 71,
    status: 'normal',
    series: [146, 158, 171, 184, 196, 220, 211, 204, 218, 226, 214, 220],
  },
  {
    id: 'checkout-auth-1',
    kind: 'auth',
    label: 'checkout-user-auth',
    namespace: 'default',
    service: 'checkout',
    api: 'GET /api/checkout/detail',
    instance: '10.24.8.32:8080',
    language: 'java',
    requests: 164200,
    hitRate: 100,
    blocked: 280,
    p95: 68,
    p99: 151,
    errorRate: 0.16,
    successRate: 99.84,
    cpu: 54,
    memory: 66,
    status: 'normal',
    series: [110, 118, 126, 136, 144, 156, 149, 152, 161, 158, 150, 154],
  },
  {
    id: 'inventory-breaker-1',
    kind: 'circuitbreaker',
    label: 'inventory-read-breaker',
    namespace: 'mall-prod',
    service: 'inventory',
    api: 'GET /api/inventory/items',
    instance: '10.24.7.12:9000',
    language: 'go',
    requests: 96400,
    hitRate: 4.8,
    blocked: 840,
    p95: 128,
    p99: 280,
    errorRate: 1.36,
    successRate: 98.64,
    cpu: 72,
    memory: 58,
    status: 'warning',
    series: [66, 72, 88, 96, 110, 128, 118, 136, 142, 132, 125, 128],
  },
  {
    id: 'inventory-detect-1',
    kind: 'faultdetect',
    label: 'inventory-active-detect',
    namespace: 'mall-prod',
    service: 'inventory',
    api: 'GET /healthz',
    instance: '10.24.7.13:9000',
    language: 'go',
    requests: 7200,
    hitRate: 100,
    blocked: 0,
    p95: 36,
    p99: 80,
    errorRate: 0.9,
    successRate: 99.1,
    cpu: 41,
    memory: 45,
    status: 'warning',
    series: [22, 24, 26, 28, 30, 34, 36, 35, 33, 38, 37, 36],
  },
  {
    id: 'payment-auth-1',
    kind: 'auth',
    label: 'payment-token-auth',
    namespace: 'default',
    service: 'payment',
    api: 'POST /api/payment/pay',
    instance: '10.24.9.17:9000',
    language: 'rust',
    requests: 156300,
    hitRate: 100,
    blocked: 312,
    p95: 58,
    p99: 130,
    errorRate: 0.2,
    successRate: 99.8,
    cpu: 46,
    memory: 39,
    status: 'normal',
    series: [104, 116, 122, 130, 141, 151, 156, 148, 152, 159, 154, 156],
  },
  {
    id: 'recommend-mock-1',
    kind: 'mock',
    label: 'recommend-fallback-mock',
    namespace: 'gray',
    service: 'recommend',
    api: 'GET /api/recommend/list',
    instance: '10.25.2.45:7001',
    language: 'go',
    requests: 42800,
    hitRate: 18.6,
    blocked: 0,
    p95: 44,
    p99: 92,
    errorRate: 0.05,
    successRate: 99.95,
    cpu: 38,
    memory: 42,
    status: 'normal',
    series: [28, 31, 35, 38, 42, 44, 41, 39, 43, 46, 42, 43],
  },
  {
    id: 'order-mirror-1',
    kind: 'mirror',
    label: 'order-shadow-mirror',
    namespace: 'default',
    service: 'order',
    api: 'POST /api/order/create',
    instance: '10.24.6.21:8080',
    language: 'java',
    requests: 88400,
    hitRate: 8.2,
    blocked: 0,
    p95: 91,
    p99: 190,
    errorRate: 0.33,
    successRate: 99.67,
    cpu: 49,
    memory: 61,
    status: 'normal',
    series: [74, 80, 84, 91, 96, 88, 93, 98, 90, 86, 92, 91],
  },
];

export const TIME_LABELS = ['12:00', '12:05', '12:10', '12:15', '12:20', '12:25', '12:30', '12:35', '12:40', '12:45', '12:50', '12:55'];

export function serviceKey(row: Pick<ServiceSignal, 'namespace' | 'service'>) {
  return `${row.namespace}/${row.service}`;
}

export function uniqueOptions<T>(rows: T[], resolve: (row: T) => string) {
  return Array.from(new Set(rows.map(resolve))).map((value) => ({ label: value, value }));
}

export function formatNumber(value: number) {
  return value.toLocaleString('zh-CN');
}

export function average(rows: ServiceSignal[], key: 'p95' | 'p99' | 'errorRate' | 'successRate' | 'cpu' | 'memory') {
  if (!rows.length) return 0;
  return rows.reduce((sum, row) => sum + row[key], 0) / rows.length;
}

export function buildPath(values: number[], width = 300, height = 120) {
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const span = Math.max(max - min, 1);
  return values.map((value, index) => {
    const x = (index / Math.max(values.length - 1, 1)) * width;
    const y = height - ((value - min) / span) * (height - 16) - 8;
    return `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
  }).join(' ');
}
