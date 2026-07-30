import React from 'react';
import { Button, Empty, Table, TableColumnData, Tag } from 'components/Fluent';
import { RefreshIcon } from 'components/Fluent/icons';
import QueryComposer from 'components/QueryComposer';
import { ResourceHeader } from 'components/ResourceLayout';
import {
  describePlatformOverview,
  PlatformComponentMetric,
  PlatformResourceMetric,
  PlatformRuntimeMetric,
} from 'services/observability';

import style from './index.module.less';

type ComponentStatus = 'healthy' | 'warning' | 'critical';
type SystemCategory = 'control-plane' | 'observability' | 'controller' | 'storage';
type DataSource = 'remote' | 'mock';

type SystemMetricRow = {
  id: string;
  category: SystemCategory;
  component: string;
  role: string;
  namespace: string;
  pod: string;
  api: string;
  status: ComponentStatus;
  cpu: number;
  memory: number;
  qps: number;
  p95: number;
  p99: number;
  errorRate: number;
  restartCount: number;
  series: number[];
  [key: string]: unknown;
};

type ResourceMetricRow = {
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
};

type RuntimeMetricRow = {
  name: string;
  title: string;
  category: string;
  value: number;
  unit: string;
  description: string;
  series: number[];
};

type QueryTimeRange = {
  startTime: number;
  endTime: number;
};

const CATEGORY_OPTIONS = [
  { label: '控制面', value: 'control-plane' },
  { label: '观测链路', value: 'observability' },
  { label: '控制器', value: 'controller' },
  { label: '存储', value: 'storage' },
];

const INTERFACE_DETAIL_PAGE_SIZE = 10;

const CATEGORY_LABEL: Record<SystemCategory, string> = {
  'control-plane': '控制面',
  observability: '观测链路',
  controller: '控制器',
  storage: '存储',
};

const SYSTEM_ROWS: SystemMetricRow[] = [
  {
    id: 'cp-services',
    category: 'control-plane',
    component: 'pole-control-plane',
    role: '注册发现 API',
    namespace: 'pole-system',
    pod: 'pole-control-plane-0',
    api: '/naming/v1/services',
    status: 'healthy',
    cpu: 38,
    memory: 54,
    qps: 186,
    p95: 42,
    p99: 88,
    errorRate: 0.08,
    restartCount: 0,
    series: [31, 35, 39, 44, 48, 42, 45, 41, 38, 40, 42, 39],
  },
  {
    id: 'cp-events',
    category: 'control-plane',
    component: 'pole-control-plane',
    role: '事件查询 API',
    namespace: 'pole-system',
    pod: 'pole-control-plane-1',
    api: '/metrics/v1/server/events',
    status: 'healthy',
    cpu: 41,
    memory: 58,
    qps: 74,
    p95: 47,
    p99: 96,
    errorRate: 0.1,
    restartCount: 0,
    series: [28, 32, 36, 34, 40, 43, 41, 39, 37, 44, 42, 45],
  },
  {
    id: 'cp-operations',
    category: 'control-plane',
    component: 'pole-control-plane',
    role: '审计查询 API',
    namespace: 'pole-system',
    pod: 'pole-control-plane-1',
    api: '/metrics/v1/server/operations',
    status: 'healthy',
    cpu: 36,
    memory: 52,
    qps: 68,
    p95: 51,
    p99: 104,
    errorRate: 0.13,
    restartCount: 0,
    series: [30, 31, 32, 35, 39, 41, 37, 36, 40, 43, 45, 42],
  },
  {
    id: 'otel-ingest',
    category: 'observability',
    component: 'otel-collector',
    role: 'OTLP ingest',
    namespace: 'observability',
    pod: 'otel-collector-74b7d9c8dd-f2f9w',
    api: '/otlp/v1/metrics',
    status: 'warning',
    cpu: 67,
    memory: 72,
    qps: 920,
    p95: 88,
    p99: 180,
    errorRate: 0.42,
    restartCount: 1,
    series: [54, 58, 63, 69, 74, 78, 83, 88, 84, 91, 86, 88],
  },
  {
    id: 'openobserve-query',
    category: 'observability',
    component: 'openobserve',
    role: '观测查询',
    namespace: 'observability',
    pod: 'openobserve-0',
    api: '/api/default/_search',
    status: 'healthy',
    cpu: 45,
    memory: 63,
    qps: 312,
    p95: 96,
    p99: 214,
    errorRate: 0.16,
    restartCount: 0,
    series: [62, 67, 70, 78, 82, 90, 96, 92, 88, 94, 99, 96],
  },
  {
    id: 'controller-sync',
    category: 'controller',
    component: 'pole-controller',
    role: 'Kubernetes 同步',
    namespace: 'pole-system',
    pod: 'pole-controller-6d778b8778-t62mc',
    api: '/controller/reconcile',
    status: 'healthy',
    cpu: 29,
    memory: 46,
    qps: 74,
    p95: 31,
    p99: 66,
    errorRate: 0.04,
    restartCount: 0,
    series: [24, 25, 28, 30, 33, 29, 31, 30, 28, 32, 31, 29],
  },
  {
    id: 'mysql-store',
    category: 'storage',
    component: 'mysql',
    role: '元数据存储',
    namespace: 'pole-system',
    pod: 'mysql-0',
    api: 'sql://pole_server',
    status: 'warning',
    cpu: 58,
    memory: 76,
    qps: 540,
    p95: 71,
    p99: 162,
    errorRate: 0.26,
    restartCount: 0,
    series: [45, 48, 52, 57, 61, 68, 71, 69, 65, 70, 74, 71],
  },
];

const SAMPLE_COUNT = 12;
const DEFAULT_WINDOW_MS = 60 * 60 * 1000;

const MOCK_RUNTIME_ROWS: RuntimeMetricRow[] = [
  {
    name: 'process.runtime.go.goroutines',
    title: 'Goroutines',
    category: 'concurrency',
    value: 312,
    unit: '{goroutine}',
    description: '当前 goroutine 数',
    series: [228, 236, 251, 260, 284, 298, 305, 312, 318, 306, 314, 312],
  },
  {
    name: 'process.runtime.go.mem.heap_alloc',
    title: 'Heap Alloc',
    category: 'memory',
    value: 46,
    unit: 'MiB',
    description: 'Go heap 已分配内存',
    series: [32, 35, 34, 38, 41, 43, 44, 45, 47, 46, 48, 46],
  },
  {
    name: 'process.runtime.go.gc.pause',
    title: 'GC Pause Avg',
    category: 'gc',
    value: 3.6,
    unit: 'ms',
    description: '时间窗内 GC pause 平均值',
    series: [2.1, 2.8, 3.2, 3.5, 3.1, 4.2, 3.8, 3.3, 3.6, 3.9, 3.4, 3.6],
  },
  {
    name: 'process.runtime.go.gc.count',
    title: 'GC Count',
    category: 'gc',
    value: 18,
    unit: '{collection}',
    description: '时间窗内 GC 次数',
    series: [1, 2, 2, 1, 3, 1, 2, 2, 1, 1, 2, 0],
  },
];

function uniqueOptions(rows: SystemMetricRow[], key: 'api' | 'component') {
  return Array.from(new Set(rows.map((row) => row[key]))).map((value) => ({ label: value, value }));
}

function average(rows: SystemMetricRow[], key: 'cpu' | 'memory' | 'p95' | 'p99' | 'errorRate') {
  if (!rows.length) return 0;
  return rows.reduce((sum, row) => sum + row[key], 0) / rows.length;
}

function maxValue(rows: SystemMetricRow[], key: 'qps' | 'restartCount') {
  if (!rows.length) return 0;
  return Math.max(...rows.map((row) => row[key]));
}

function mockResourcesFrom(rows: SystemMetricRow[]): ResourceMetricRow[] {
  const resources = new Map<string, ResourceMetricRow>();
  rows.forEach((row) => {
    const key = `${row.component}-${row.pod}`;
    if (!resources.has(key)) {
      resources.set(key, {
        id: key,
        component: row.component,
        namespace: row.namespace,
        pod: row.pod,
        cpu: row.cpu,
        memory: row.memory,
        memoryUnit: '%',
        restartCount: row.restartCount,
        cpuSeries: row.series,
        memorySeries: row.series.map((value) => Math.min(100, Math.round(value * 1.15))),
      });
    }
  });
  return Array.from(resources.values());
}

function statusClass(status: ComponentStatus) {
  if (status === 'critical') return style.toneDanger;
  if (status === 'warning') return style.toneWarning;
  return style.toneSuccess;
}

function categoryClass(category: SystemCategory) {
  if (category === 'observability') return style.toneInfo;
  if (category === 'storage') return style.toneWarning;
  return style.toneNeutral;
}

function normalizeRemoteRow(row: PlatformComponentMetric): SystemMetricRow {
  const category = CATEGORY_LABEL[row.category as SystemCategory] ? row.category as SystemCategory : 'control-plane';
  const status = row.status === 'critical' || row.status === 'warning' ? row.status : 'healthy';
  return {
    id: row.id,
    category,
    component: row.component || 'pole-control-plane',
    role: row.role || '控制面 API',
    namespace: row.namespace || 'pole-system',
    pod: row.pod || row.component || 'pole-control-plane',
    api: row.api || 'unknown',
    status,
    cpu: Math.round(row.cpu || 0),
    memory: Math.round(row.memory || 0),
    qps: Math.round(row.qps || 0),
    p95: Math.round(row.p95 || 0),
    p99: Math.round(row.p99 || 0),
    errorRate: Number((row.errorRate || 0).toFixed(2)),
    restartCount: row.restartCount || 0,
    series: row.series?.length ? row.series.map((value) => Math.round(value)) : [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  };
}

function normalizeResourceRow(row: PlatformResourceMetric): ResourceMetricRow {
  return {
    id: row.id,
    component: row.component || 'pole-control-plane',
    namespace: row.namespace || 'pole-system',
    pod: row.pod || row.component || 'pole-control-plane',
    cpu: Number((row.cpu || 0).toFixed(2)),
    memory: Number((row.memory || 0).toFixed(2)),
    memoryUnit: row.memoryUnit || 'MiB',
    restartCount: row.restartCount || 0,
    cpuSeries: row.cpuSeries || [],
    memorySeries: row.memorySeries || [],
  };
}

function normalizeRuntimeRow(row: PlatformRuntimeMetric): RuntimeMetricRow {
  return {
    name: row.name,
    title: row.title || row.name,
    category: row.category || 'runtime',
    value: Number((row.value || 0).toFixed(2)),
    unit: row.unit || '',
    description: row.description || row.name,
    series: row.series || [],
  };
}

function buildPath(values: number[], width = 300, height = 120) {
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const span = Math.max(max - min, 1);
  return values.map((value, index) => {
    const x = (index / Math.max(values.length - 1, 1)) * width;
    const y = height - ((value - min) / span) * (height - 16) - 8;
    return `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
  }).join(' ');
}

function buildTrendPoints(values: number[], width: number, height: number) {
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const span = Math.max(max - min, 1);
  return values.map((value, index) => {
    const x = (index / Math.max(values.length - 1, 1)) * width;
    const y = height - ((value - min) / span) * (height - 16) - 8;
    return { index, value, x, y };
  });
}

function formatMetricValue(value: number, unit?: string) {
  const formatted = Number.isInteger(value)
    ? value.toLocaleString('zh-CN')
    : value.toLocaleString('zh-CN', { maximumFractionDigits: 2 });
  if (!unit || unit.startsWith('{')) {
    return formatted;
  }
  return `${formatted}${unit}`;
}

function buildQueryTimeRange(): QueryTimeRange {
  const endTime = Date.now();
  return {
    startTime: endTime - DEFAULT_WINDOW_MS,
    endTime,
  };
}

function formatLocalTime(timestamp: number) {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(timestamp));
}

function buildTimeLabels(range: QueryTimeRange, count = SAMPLE_COUNT) {
  if (count <= 1) {
    return [formatLocalTime(range.endTime)];
  }
  const span = Math.max(range.endTime - range.startTime, 0);
  const step = span / (count - 1);
  return Array.from({ length: count }, (_, index) => formatLocalTime(range.startTime + (step * index)));
}

function formatTimezoneOffset(date = new Date()) {
  const offsetMinutes = -date.getTimezoneOffset();
  const sign = offsetMinutes >= 0 ? '+' : '-';
  const absolute = Math.abs(offsetMinutes);
  const hours = String(Math.floor(absolute / 60)).padStart(2, '0');
  const minutes = String(absolute % 60).padStart(2, '0');
  return `UTC${sign}${hours}:${minutes}`;
}

function axisTicks(values: number[], unit?: string) {
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const middle = min + ((max - min) / 2);
  return [max, middle, min].map((value) => formatMetricValue(value, unit));
}

function Panel(props: { title: string; subtitle?: string; query?: string; children: React.ReactNode; className?: string }) {
  const { title, subtitle, query, children, className } = props;
  return (
    <section className={`${style.panel} ${className || ''}`}>
      <div className={style.panelHeader}>
        <div className={style.panelTitleBlock}>
          <h2>{title}</h2>
          {subtitle && <span>{subtitle}</span>}
        </div>
        {query && <code>{query}</code>}
      </div>
      <div className={style.panelBody}>{children}</div>
    </section>
  );
}

function StatPanel(props: { title: string; value: React.ReactNode; hint: string; query?: string; tone?: string }) {
  const { title, value, hint, query, tone } = props;
  return (
    <Panel title={title} query={query} className={`${style.statPanel} ${tone || ''}`}>
      <strong className={style.statValue}>{value}</strong>
      <span className={style.statHint}>{hint}</span>
    </Panel>
  );
}

function TimeSeriesPanel(props: { rows: SystemMetricRow[]; timeLabels: string[] }) {
  const { rows, timeLabels } = props;
  const [hoverIndex, setHoverIndex] = React.useState<number | null>(null);
  const series = rows.length ? rows.map((row) => row.series).reduce((acc, current) => acc.map((value, index) => Math.round((value + current[index]) / 2))) : [];
  const width = 300;
  const height = 120;
  const points = series.length ? buildTrendPoints(series, width, height) : [];
  const hoverPoint = hoverIndex === null ? null : points[hoverIndex];
  const hoverLeft = hoverPoint ? Math.min(90, Math.max(10, (hoverPoint.x / width) * 100)) : 0;
  const yLabels = series.length ? axisTicks(series, 'ms') : [];

  const updateHover = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!series.length) return;
    const rect = event.currentTarget.getBoundingClientRect();
    const offset = Math.min(Math.max(event.clientX - rect.left, 0), rect.width);
    const nextIndex = Math.round((offset / Math.max(rect.width, 1)) * (series.length - 1));
    setHoverIndex(Math.min(Math.max(nextIndex, 0), series.length - 1));
  };

  return (
    <div className={style.timeSeries}>
      {series.length ? (
        <div className={style.chartFrame}>
          <div className={style.yAxis} aria-hidden="true">
            {yLabels.map((label, index) => <span key={`${label}-${index}`}>{label}</span>)}
          </div>
          <div className={style.chartArea}>
            <div
              className={style.chartSurface}
              role="img"
              aria-label="接口延迟趋势，包含时间线和纵坐标，悬浮查看采样值"
              tabIndex={0}
              onMouseMove={updateHover}
              onMouseLeave={() => setHoverIndex(null)}
              onFocus={() => setHoverIndex(series.length - 1)}
              onBlur={() => setHoverIndex(null)}
            >
              {hoverPoint && (
                <div className={style.timeSeriesTooltip} style={{ left: `${hoverLeft}%` }}>
                  <strong>p95 latency</strong>
                  <span>{timeLabels[hoverPoint.index] || `采样点 ${hoverPoint.index + 1}`}</span>
                  <b>{hoverPoint.value}ms</b>
                </div>
              )}
              <svg viewBox={`0 0 ${width} ${height}`} aria-hidden="true">
                <path className={style.gridLine} d="M 0 24 L 300 24 M 0 60 L 300 60 M 0 96 L 300 96" />
                <path className={style.seriesArea} d={`${buildPath(series, width, height)} L ${width} ${height} L 0 ${height} Z`} />
                <path className={style.seriesLine} d={buildPath(series, width, height)} />
                {hoverPoint && (
                  <>
                    <path className={style.chartGuide} d={`M ${hoverPoint.x.toFixed(1)} 8 L ${hoverPoint.x.toFixed(1)} 114`} />
                    <circle className={style.chartPoint} cx={hoverPoint.x} cy={hoverPoint.y} r="3.5" />
                  </>
                )}
              </svg>
            </div>
            <div className={style.axisLabels}>
              <span>{timeLabels[0]}</span>
              <span>{timeLabels[Math.floor(timeLabels.length / 2)]}</span>
              <span>{timeLabels[timeLabels.length - 1]}</span>
            </div>
          </div>
        </div>
      ) : <Empty description="暂无趋势数据" />}
    </div>
  );
}

function HeatmapPanel(props: { rows: SystemMetricRow[]; timeLabels: string[] }) {
  const { rows, timeLabels } = props;
  const labels = timeLabels.slice(-8);
  if (!rows.length) {
    return <Empty description="暂无热力图数据" />;
  }
  return (
    <div className={style.heatmapFrame}>
      <div />
      <div className={style.heatmapXAxis} aria-hidden="true">
        {labels.map((label) => <span key={label}>{label}</span>)}
      </div>
      {rows.map((row) => (
        <React.Fragment key={row.id}>
          <div className={style.heatmapYAxis} title={`${row.component} / ${row.api}`}>{row.api}</div>
          <div className={style.heatmapRow}>
            {row.series.slice(-8).map((value, index) => (
              <span
                key={`${row.id}-${index}`}
                className={value > 80 ? style.heatHigh : value > 60 ? style.heatWarn : style.heatOk}
                title={`${row.component} ${row.api} ${labels[index]}: ${value}ms`}
              />
            ))}
          </div>
        </React.Fragment>
      ))}
    </div>
  );
}

function TinyTrend(props: { values: number[]; label: string; timeLabels: string[]; unit?: string }) {
  const { values, label, timeLabels, unit } = props;
  const [hoverIndex, setHoverIndex] = React.useState<number | null>(null);
  const series = values.length ? values : [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
  const width = 280;
  const height = 64;
  const points = buildTrendPoints(series, width, height);
  const hoverPoint = hoverIndex === null ? null : points[hoverIndex];
  const hoverLeft = hoverPoint ? Math.min(86, Math.max(14, (hoverPoint.x / width) * 100)) : 0;

  const updateHover = (event: React.MouseEvent<HTMLDivElement>) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const offset = Math.min(Math.max(event.clientX - rect.left, 0), rect.width);
    const nextIndex = Math.round((offset / Math.max(rect.width, 1)) * (series.length - 1));
    setHoverIndex(Math.min(Math.max(nextIndex, 0), series.length - 1));
  };

  return (
    <div
      className={style.tinyTrend}
      role="img"
      aria-label={`${label} stat sparkline，悬浮查看采样值`}
      tabIndex={0}
      onMouseMove={updateHover}
      onMouseLeave={() => setHoverIndex(null)}
      onFocus={() => setHoverIndex(series.length - 1)}
      onBlur={() => setHoverIndex(null)}
    >
      {hoverPoint && (
        <div
          className={style.tinyTrendTooltip}
          style={{ left: `${hoverLeft}%` }}
        >
          <strong>{label}</strong>
          <span>{timeLabels[hoverPoint.index] || `采样点 ${hoverPoint.index + 1}`}</span>
          <b>{formatMetricValue(hoverPoint.value, unit)}</b>
        </div>
      )}
      <svg viewBox={`0 0 ${width} ${height}`} aria-hidden="true">
        <path className={style.tinyTrendGrid} d={`M 0 14 L ${width} 14 M 0 42 L ${width} 42`} />
        <path className={style.tinyTrendArea} d={`${buildPath(series, width, height)} L ${width} ${height} L 0 ${height} Z`} />
        <path className={style.tinyTrendLine} d={buildPath(series, width, height)} />
        {hoverPoint && (
          <>
            <path className={style.tinyTrendGuide} d={`M ${hoverPoint.x.toFixed(1)} 8 L ${hoverPoint.x.toFixed(1)} 58`} />
            <circle className={style.tinyTrendPoint} cx={hoverPoint.x} cy={hoverPoint.y} r="3.6" />
          </>
        )}
      </svg>
      <div className={style.tinyTimeRange} aria-hidden="true">
        <span>{timeLabels[0]}</span>
        <span>{timeLabels[timeLabels.length - 1]}</span>
      </div>
    </div>
  );
}

function ResourceBoard(props: { rows: ResourceMetricRow[]; source: DataSource }) {
  const { rows, source } = props;
  if (!rows.length) {
    return <Empty description={source === 'remote' ? '暂无 CPU/MEM 资源指标' : '暂无资源数据'} />;
  }
  return (
    <div className={style.resourceBoard}>
      {rows.map((row) => (
        <div className={style.resourceCard} key={row.id}>
          <div className={style.resourceCardHeader}>
            <div>
              <strong>{row.component}</strong>
              <span>{row.namespace} / {row.pod}</span>
            </div>
            <Tag className={row.restartCount > 0 ? style.toneWarning : style.toneSuccess}>
              {row.restartCount > 0 ? `重启 ${row.restartCount}` : 'Ready'}
            </Tag>
          </div>
          <div className={style.resourceValues}>
            <div>
              <span>CPU</span>
              <strong>{row.cpu}%</strong>
              <i><b style={{ width: `${Math.max(row.cpu, 3)}%` }} /></i>
            </div>
            <div>
              <span>MEM</span>
              <strong>{row.memory}{row.memoryUnit}</strong>
              <i><b style={{ width: `${row.memoryUnit === '%' ? Math.max(row.memory, 3) : Math.min(row.memory, 100)}%` }} /></i>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function RuntimeBoard(props: { rows: RuntimeMetricRow[]; timeLabels: string[] }) {
  const { rows, timeLabels } = props;
  if (!rows.length) {
    return <Empty description="暂无 Go runtime 指标" />;
  }
  return (
    <div className={style.runtimeGrid}>
      {rows.map((row) => (
        <div className={style.runtimeItem} key={row.name}>
          <div className={style.runtimeItemHeader}>
            <span>{row.title}</span>
            <em>{row.category}</em>
          </div>
          <strong>{formatRuntimeValue(row)}</strong>
          <p>{row.description}</p>
          <TinyTrend values={row.series} label={row.title} timeLabels={timeLabels} unit={row.unit} />
        </div>
      ))}
    </div>
  );
}

function formatRuntimeValue(row: RuntimeMetricRow) {
  return formatMetricValue(row.value, row.unit);
}

export default function SystemMonitor() {
  const [category, setCategory] = React.useState('');
  const [api, setApi] = React.useState('');
  const [componentKeyword, setComponentKeyword] = React.useState('');
  const [remoteRows, setRemoteRows] = React.useState<SystemMetricRow[]>([]);
  const [remoteResources, setRemoteResources] = React.useState<ResourceMetricRow[]>([]);
  const [remoteRuntime, setRemoteRuntime] = React.useState<RuntimeMetricRow[]>([]);
  const [dataSource, setDataSource] = React.useState<DataSource>('mock');
  const [timeRange, setTimeRange] = React.useState<QueryTimeRange>(() => buildQueryTimeRange());
  const timezoneLabel = React.useMemo(() => formatTimezoneOffset(), []);

  const loadRemoteData = React.useCallback(async () => {
    const nextTimeRange = buildQueryTimeRange();
    setTimeRange(nextTimeRange);
    try {
      const overview = await describePlatformOverview({
        start_time: String(nextTimeRange.startTime),
        end_time: String(nextTimeRange.endTime),
        step: '1m',
        category,
        api,
      });
      const rows = (overview.components || []).map(normalizeRemoteRow);
      const resources = (overview.resources || []).map(normalizeResourceRow);
      const runtime = (overview.runtime || []).map(normalizeRuntimeRow);
      setRemoteRows(rows);
      setRemoteResources(resources);
      setRemoteRuntime(runtime);
      setDataSource(rows.length || resources.length || runtime.length ? 'remote' : 'mock');
    } catch {
      setRemoteRows([]);
      setRemoteResources([]);
      setRemoteRuntime([]);
      setDataSource('mock');
    }
  }, [api, category]);

  React.useEffect(() => {
    void loadRemoteData();
  }, [loadRemoteData]);

  const sourceRows = dataSource === 'remote' ? remoteRows : SYSTEM_ROWS;
  const sourceResources = dataSource === 'remote' ? remoteResources : mockResourcesFrom(SYSTEM_ROWS);
  const sourceRuntime = dataSource === 'remote' ? remoteRuntime : MOCK_RUNTIME_ROWS;
  const timeLabels = React.useMemo(() => buildTimeLabels(timeRange), [timeRange]);

  const filteredRows = React.useMemo(() => sourceRows.filter((row) => {
    const matchCategory = !category || row.category === category;
    const matchApi = !api || row.api === api;
    const matchComponent = !componentKeyword || `${row.component} ${row.pod} ${row.role}`.includes(componentKeyword);
    return matchCategory && matchApi && matchComponent;
  }), [api, category, componentKeyword, sourceRows]);

  const unhealthyCount = filteredRows.filter((item) => item.status !== 'healthy').length;
  const avgP95 = Math.round(average(filteredRows, 'p95'));
  const avgP99 = Math.round(average(filteredRows, 'p99'));
  const avgErrorRate = average(filteredRows, 'errorRate').toFixed(2);
  const maxQps = maxValue(filteredRows, 'qps');
  const goroutineMetric = sourceRuntime.find((item) => item.name === 'process.runtime.go.goroutines');
  const interfacePaginationKey = [category, api, componentKeyword, dataSource].join('|');

  const columns = React.useMemo<TableColumnData<SystemMetricRow>[]>(() => [
    {
      colKey: 'category',
      title: '类别',
      width: 110,
      cell: ({ row }) => <Tag className={categoryClass(row.category)}>{CATEGORY_LABEL[row.category]}</Tag>,
    },
    {
      colKey: 'component',
      title: '组件',
      width: 190,
      cell: ({ row }) => (
        <div className={style.componentCell}>
          <strong>{row.component}</strong>
          <span>{row.role}</span>
        </div>
      ),
    },
    {
      colKey: 'api',
      title: '接口',
      width: 210,
      ellipsis: true,
    },
    {
      colKey: 'pod',
      title: 'Pod',
      width: 210,
      ellipsis: true,
    },
    {
      colKey: 'status',
      title: '状态',
      width: 90,
      cell: ({ row }) => <Tag className={statusClass(row.status)}>{row.status === 'healthy' ? '健康' : '关注'}</Tag>,
    },
    {
      colKey: 'qps',
      title: 'QPS',
      width: 90,
      align: 'right',
    },
    {
      colKey: 'p95',
      title: 'P95 / P99',
      width: 120,
      align: 'right',
      cell: ({ row }) => `${row.p95}ms / ${row.p99}ms`,
    },
    {
      colKey: 'errorRate',
      title: '错误率',
      width: 90,
      align: 'right',
      cell: ({ row }) => `${row.errorRate}%`,
    },
  ], []);

  const resetFilters = () => {
    setCategory('');
    setApi('');
    setComponentKeyword('');
  };

  return (
    <div className={style.page}>
      <ResourceHeader
        density="compact"
        placement="app-header"
        eyebrow="监控指标 / 系统监控"
        title="系统监控"
        description="Grafana-like 平台组件看板，按类别、接口和组件关键字聚合 OTel 指标。"
        actions={(
          <div className={style.dashboardActions}>
          <span className={style.dashboardPill}>Last 1 hour</span>
          <span className={style.dashboardPill}>Step 1m</span>
          <span className={style.dashboardPill}>{timezoneLabel}</span>
          <span className={style.dashboardPill}>GreptimeDB</span>
          <Tag className={dataSource === 'mock' ? style.sourceTagMock : style.sourceTag}>{
            dataSource === 'mock' ? 'Mock 预览' : '实时数据'
          }</Tag>
          <Button icon={<RefreshIcon />} variant="outline" onClick={() => void loadRemoteData()}>刷新</Button>
          </div>
        )}
      />

      <section className={style.variableBar} aria-label="Dashboard variables">
        <QueryComposer
          keyword={componentKeyword}
          keywordPlaceholder="搜索组件、Pod 或角色"
          suggestions={Array.from(new Set(sourceRows.flatMap((item) => [item.component, item.pod]).filter(Boolean)))}
          fields={[
            {
              key: 'category',
              label: '类别',
              type: 'select',
              options: CATEGORY_OPTIONS,
            },
            {
              key: 'api',
              label: '接口',
              type: 'select',
              filterable: true,
              options: uniqueOptions(sourceRows, 'api'),
            },
          ]}
          values={{ category, api }}
          onKeywordChange={setComponentKeyword}
          onValuesChange={(values) => {
            setCategory(String(values.category || ''));
            setApi(String(values.api || ''));
          }}
          onReset={resetFilters}
          mode="instant"
        />
      </section>

      <div className={style.statGrid}>
        <StatPanel title="组件 / 接口" query="count()" value={`${filteredRows.length} / ${new Set(filteredRows.map((row) => row.api)).size}`} hint="当前变量筛选结果" />
        <StatPanel title="需关注组件" query="status != healthy" value={unhealthyCount} hint="warning / critical 状态" tone={unhealthyCount ? style.panelWarning : undefined} />
        <StatPanel title="P95 / P99" query="histogram_quantile()" value={`${avgP95}ms / ${avgP99}ms`} hint={`错误率 ${avgErrorRate}%`} />
        <StatPanel title="峰值 QPS" query="max(rate())" value={maxQps.toLocaleString('zh-CN')} hint="筛选结果内最高组件 QPS" />
        <StatPanel title="Go 协程" query="process.runtime.go.goroutines" value={goroutineMetric ? formatRuntimeValue(goroutineMetric) : 0} hint="pole-control-plane runtime" />
      </div>

      <div className={style.dashboardGrid}>
        <Panel title="接口延迟趋势" subtitle="按当前变量聚合 p95 序列" query="pole_control_plane_request_duration_seconds" className={style.panelLarge}>
          <TimeSeriesPanel rows={filteredRows} timeLabels={timeLabels} />
        </Panel>
        <Panel title="延迟热力图" subtitle="接口与组件最近采样点" query="p95 by api, component">
          <HeatmapPanel rows={filteredRows} timeLabels={timeLabels} />
        </Panel>
        <Panel title="Go runtime 看板" subtitle="pole-control-plane Go 服务端标准 runtime 指标" query="process_runtime_go_*" className={style.panelFull}>
          <RuntimeBoard rows={sourceRuntime} timeLabels={timeLabels} />
        </Panel>
        <Panel title="组件资源看板" subtitle="CPU/MEM 资源指标" query="k8s.pod.* / container.*" className={style.panelFull}>
          <ResourceBoard rows={sourceResources} source={dataSource} />
        </Panel>
        <Panel title="接口明细" subtitle="component / api / pod 等 OTel resource labels 级别指标" query="group by component, api, pod" className={style.panelFull}>
          <Table
            key={interfacePaginationKey}
            data={filteredRows}
            columns={columns}
            rowKey="id"
            tableLayout="fixed"
            cellEmptyContent="-"
            empty={<Empty description="暂无系统监控数据" />}
            pagination={{
              pageSize: INTERFACE_DETAIL_PAGE_SIZE,
              total: filteredRows.length,
              showJumper: true,
            }}
          />
        </Panel>
      </div>
    </div>
  );
}
