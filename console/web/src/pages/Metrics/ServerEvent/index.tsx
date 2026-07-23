import React from 'react';
import dayjs from 'dayjs';
import { Button, DateRangePicker, DateRangePickerProps, Drawer, Empty, Input, Select, Table, TableColumnData, Tag } from 'components/Fluent';
import { FilterIcon, InfoCircleIcon, RefreshIcon, SearchIcon, ServerIcon } from 'components/Fluent/icons';
import { describeEventLog, EventLog, EventType, getEventTypeInfo } from 'services/observer';
import { describeObservabilityEvents } from 'services/observability';
import ErrorPage from 'components/ErrorPage';
import { useSearchParams } from 'react-router-dom';

import style from './index.module.less';

type DataSource = 'remote' | 'mock';

type EventMetricRow = EventLog & {
  rowKey?: string;
  [key: string]: unknown;
};

interface SearchState {
  searchEvent: string;
  searchNamespace: string;
  searchService: string;
  searchResource: string;
  startTime: string;
  endTime: string;
}

const PAGE_SIZE = 10;

const MOCK_EVENTS: EventMetricRow[] = [
  {
    cursor: 'evt-20260718-001',
    namespace: 'default',
    service: 'checkout',
    resource: '10.24.8.31:8080',
    event_type: 'InstanceTurnUnHealth',
    event_time: '2026-07-18 12:42:18',
    server: 'pole-control-plane-0',
  },
  {
    cursor: 'evt-20260718-002',
    namespace: 'default',
    service: 'payment',
    resource: '10.24.9.17:9000',
    event_type: 'InstanceOffline',
    event_time: '2026-07-18 12:35:04',
    server: 'pole-control-plane-1',
  },
  {
    cursor: 'evt-20260718-003',
    namespace: 'mall-prod',
    service: 'inventory',
    resource: 'inventory.mesh.local',
    event_type: 'ServiceOpenEmptyPushProtect',
    event_time: '2026-07-18 12:21:36',
    server: 'pole-control-plane-0',
  },
  {
    cursor: 'evt-20260718-004',
    namespace: 'mall-prod',
    service: 'checkout',
    resource: '10.24.8.33:8080',
    event_type: 'InstanceTurnHealth',
    event_time: '2026-07-18 12:06:57',
    server: 'pole-control-plane-1',
  },
  {
    cursor: 'evt-20260718-005',
    namespace: 'gray',
    service: 'recommend',
    resource: '10.25.2.45:7001',
    event_type: 'InstanceOpenIsolate',
    event_time: '2026-07-18 11:53:14',
    server: 'pole-control-plane-0',
  },
  {
    cursor: 'evt-20260718-006',
    namespace: 'default',
    service: 'user-center',
    resource: '10.24.1.12:8080',
    event_type: 'InstanceOnline',
    event_time: '2026-07-18 11:39:41',
    server: 'pole-control-plane-1',
  },
];

const EVENT_TREND = [
  { label: '08:00', value: 6 },
  { label: '09:00', value: 11 },
  { label: '10:00', value: 8 },
  { label: '11:00', value: 15 },
  { label: '12:00', value: 22 },
  { label: '13:00', value: 13 },
  { label: '14:00', value: 18 },
  { label: '15:00', value: 9 },
];

const presets: DateRangePickerProps['presets'] = {
  '最近 1 小时': [dayjs().subtract(1, 'hour').toDate(), dayjs().toDate()],
  '最近 6 小时': [dayjs().subtract(6, 'hour').toDate(), dayjs().toDate()],
  '最近 24 小时': [dayjs().subtract(24, 'hour').toDate(), dayjs().toDate()],
};

const defaultSearchState: SearchState = {
  searchEvent: '',
  searchNamespace: '',
  searchService: '',
  searchResource: '',
  startTime: '',
  endTime: '',
};

function resolveEventLabel(eventType: string) {
  return getEventTypeInfo(eventType)?.label || eventType || '-';
}

function resolveEventCategory(eventType: string) {
  return getEventTypeInfo(eventType)?.category || '未知事件';
}

function isRiskEvent(eventType: string) {
  return ['InstanceOffline', 'InstanceTurnUnHealth', 'InstanceOpenIsolate', 'ServiceOpenEmptyPushProtect', 'ServiceExpireEmptyPushProtect'].includes(eventType);
}

function isRecoveryEvent(eventType: string) {
  return ['InstanceOnline', 'InstanceTurnHealth', 'InstanceCloseIsolate', 'ServiceCloseEmptyPushProtect'].includes(eventType);
}

function eventToneClass(eventType: string) {
  if (isRiskEvent(eventType)) return style.toneDanger;
  if (isRecoveryEvent(eventType)) return style.toneSuccess;
  return style.toneInfo;
}

function formatTime(value: string) {
  if (!value) return '-';
  const date = dayjs(value);
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm:ss') : value;
}

function filterMockEvents(rows: EventMetricRow[], searchState: SearchState) {
  return rows.filter((row) => {
    const matchEvent = !searchState.searchEvent || row.event_type === searchState.searchEvent;
    const matchNamespace = !searchState.searchNamespace || row.namespace.includes(searchState.searchNamespace);
    const matchService = !searchState.searchService || row.service.includes(searchState.searchService);
    const matchResource = !searchState.searchResource || row.resource.includes(searchState.searchResource);
    return matchEvent && matchNamespace && matchService && matchResource;
  });
}

function buildCategoryStats(rows: EventMetricRow[]) {
  const total = Math.max(rows.length, 1);
  const categories = rows.reduce<Record<string, number>>((acc, row) => {
    const category = resolveEventCategory(row.event_type);
    acc[category] = (acc[category] || 0) + 1;
    return acc;
  }, {});
  return Object.entries(categories).map(([label, value]) => ({
    label,
    value,
    percent: Math.round((value / total) * 100),
  }));
}

function buildEventSummary(rows: EventMetricRow[]) {
  const affectedServices = new Set(rows.map((row) => `${row.namespace}/${row.service}`));
  const riskCount = rows.filter((row) => isRiskEvent(row.event_type)).length;
  const latestEvent = rows[0];
  return {
    total: rows.length,
    riskCount,
    affectedServices: affectedServices.size,
    latestTime: latestEvent ? formatTime(latestEvent.event_time) : '-',
  };
}

function SummaryCard(props: { title: string; value: React.ReactNode; hint: string; tone?: string; compactValue?: boolean }) {
  const { title, value, hint, tone, compactValue } = props;
  return (
    <section className={`${style.summaryCard} ${tone || ''}`}>
      <span className={style.summaryTitle}>{title}</span>
      <strong className={`${style.summaryValue} ${compactValue ? style.summaryValueCompact : ''}`}>{value}</strong>
      <span className={style.summaryHint}>{hint}</span>
    </section>
  );
}

function TrendBars(props: { data: Array<{ label: string; value: number }> }) {
  const { data } = props;
  const max = Math.max(...data.map((item) => item.value), 1);
  return (
    <div className={style.trendBars}>
      {data.map((item) => (
        <div className={style.trendItem} key={item.label}>
          <span className={style.trendValue}>{item.value}</span>
          <span className={style.trendTrack}>
            <span className={style.trendFill} style={{ height: `${Math.max((item.value / max) * 100, 8)}%` }} />
          </span>
          <span className={style.trendLabel}>{item.label}</span>
        </div>
      ))}
    </div>
  );
}

function DistributionList(props: { data: Array<{ label: string; value: number; percent: number }> }) {
  const { data } = props;
  if (!data.length) return <Empty description="暂无事件分布" />;
  return (
    <div className={style.distributionList}>
      {data.map((item) => (
        <div className={style.distributionItem} key={item.label}>
          <div className={style.distributionMeta}>
            <span>{item.label}</span>
            <strong>{item.value}</strong>
          </div>
          <span className={style.progressTrack}>
            <span className={style.progressFill} style={{ width: `${item.percent}%` }} />
          </span>
        </div>
      ))}
    </div>
  );
}

function DataSourceBadge(props: { source: DataSource; loading: boolean }) {
  const { source, loading } = props;
  if (loading) return <Tag className={style.sourceTag}>加载中</Tag>;
  return <Tag className={source === 'mock' ? style.sourceTagMock : style.sourceTag}>{source === 'mock' ? 'Mock 预览' : '实时数据'}</Tag>;
}

export default function ServerEvent() {
  const [searchParams] = useSearchParams();
  const initialSearchState = React.useRef<SearchState>({
    ...defaultSearchState,
    searchNamespace: searchParams.get('namespace') || '',
    searchService: searchParams.get('resource_name') || '',
  });
  const [events, setEvents] = React.useState<EventMetricRow[]>([]);
  const [searchState, setSearchState] = React.useState<SearchState>(initialSearchState.current);
  const [isLoading, setIsLoading] = React.useState(false);
  const [hasNext, setHasNext] = React.useState(false);
  const [dataSource, setDataSource] = React.useState<DataSource>('mock');
  const [fetchError, setFetchError] = React.useState(false);
  const [selectedEvent, setSelectedEvent] = React.useState<EventMetricRow | null>(null);

  const displayEvents = React.useMemo(() => {
    if (dataSource === 'remote') return events;
    return filterMockEvents(MOCK_EVENTS, searchState);
  }, [dataSource, events, searchState]);

  const summary = React.useMemo(() => buildEventSummary(displayEvents), [displayEvents]);
  const categoryStats = React.useMemo(() => buildCategoryStats(displayEvents), [displayEvents]);

  const fetchData = React.useCallback(async (nextSearchState: SearchState = searchState) => {
    setIsLoading(true);
    try {
      const requestParams = {
        namespace: nextSearchState.searchNamespace,
        service: nextSearchState.searchService,
        resource: nextSearchState.searchResource,
        event_type: nextSearchState.searchEvent,
        start_time: nextSearchState.startTime,
        end_time: nextSearchState.endTime,
        limit: PAGE_SIZE,
        cursor: 0,
        direction: 'next',
      } as const;
      const observabilityRes = await describeObservabilityEvents(requestParams);
      const fallbackRes = (observabilityRes.data || []).length ? observabilityRes : await describeEventLog(requestParams);
      const res = fallbackRes;
      const remoteRows = (res.data || []) as EventMetricRow[];
      setEvents(remoteRows);
      setHasNext(Boolean(res.has_next));
      setDataSource(remoteRows.length ? 'remote' : 'mock');
      setFetchError(false);
    } catch {
      setEvents([]);
      setHasNext(false);
      setDataSource('mock');
      setFetchError(false);
    } finally {
      setIsLoading(false);
    }
  }, [searchState]);

  React.useEffect(() => {
    fetchData(initialSearchState.current);
  }, []);

  const columns = React.useMemo<TableColumnData<EventMetricRow>[]>(() => [
    {
      colKey: 'event_type',
      title: '事件',
      width: 180,
      cell: ({ row }) => (
        <div className={style.eventCell}>
          <Tag className={eventToneClass(row.event_type)}>{resolveEventLabel(row.event_type)}</Tag>
          <span>{resolveEventCategory(row.event_type)}</span>
        </div>
      ),
    },
    {
      colKey: 'namespace',
      title: '命名空间',
      width: 120,
      ellipsis: true,
    },
    {
      colKey: 'service',
      title: '服务',
      width: 140,
      ellipsis: true,
    },
    {
      colKey: 'resource',
      title: '资源',
      width: 180,
      ellipsis: true,
    },
    {
      colKey: 'server',
      title: '上报节点',
      width: 150,
      ellipsis: true,
    },
    {
      colKey: 'event_time',
      title: '发生时间',
      width: 170,
      cell: ({ row }) => formatTime(row.event_time),
    },
  ], []);

  const updateSearchState = (patch: Partial<SearchState>) => {
    setSearchState((previous) => ({ ...previous, ...patch }));
  };

  const handleReset = () => {
    setSearchState(defaultSearchState);
    fetchData(defaultSearchState);
  };

  if (fetchError) return <ErrorPage />;

  return (
    <div className={style.page}>
      <section className={style.headerPanel}>
        <div className={style.headerMain}>
          <span className={style.headerIcon}><ServerIcon /></span>
          <div>
            <h1>事件指标</h1>
            <p>聚合服务实例、服务保护和控制面事件，快速定位事件影响范围与最近变化。</p>
          </div>
        </div>
        <DataSourceBadge source={dataSource} loading={isLoading} />
      </section>

      <div className={style.summaryGrid}>
        <SummaryCard title="事件总数" value={summary.total} hint={dataSource === 'mock' ? '当前筛选下的预览事件' : '当前筛选下的真实事件'} />
        <SummaryCard title="风险事件" value={summary.riskCount} hint="下线、异常、隔离和推空保护" tone={style.summaryDanger} />
        <SummaryCard title="影响服务" value={summary.affectedServices} hint="按 namespace/service 去重" />
        <SummaryCard title="最近事件" value={summary.latestTime} hint={hasNext ? '还有更多历史事件' : '已展示当前结果集'} compactValue />
      </div>

      <div className={style.analysisGrid}>
        <section className={style.panel}>
          <div className={style.panelHeader}>
            <div>
              <h2>事件趋势</h2>
              <span>按小时聚合最近事件量</span>
            </div>
            <InfoCircleIcon />
          </div>
          <TrendBars data={EVENT_TREND} />
        </section>

        <section className={style.panel}>
          <div className={style.panelHeader}>
            <div>
              <h2>类型分布</h2>
              <span>事件类别占比</span>
            </div>
            <FilterIcon />
          </div>
          <DistributionList data={categoryStats} />
        </section>
      </div>

      <section className={style.panel}>
        <div className={style.panelHeader}>
          <div>
            <h2>事件明细</h2>
            <span>筛选服务、资源或事件类型后查看具体事件上下文</span>
          </div>
          <Button icon={<RefreshIcon />} variant="outline" loading={isLoading} onClick={() => fetchData()}>
            刷新
          </Button>
        </div>

        <div className={style.filterGrid}>
          <Select
            className={style.filterControl}
            clearable
            value={searchState.searchEvent}
            options={EventType}
            placeholder="事件类型"
            onChange={(value) => updateSearchState({ searchEvent: value as string })}
          />
          <Input
            className={style.filterControl}
            clearable
            value={searchState.searchNamespace}
            placeholder="命名空间"
            onChange={(value) => updateSearchState({ searchNamespace: value })}
          />
          <Input
            className={style.filterControl}
            clearable
            value={searchState.searchService}
            placeholder="服务名"
            onChange={(value) => updateSearchState({ searchService: value })}
          />
          <Input
            className={style.filterControl}
            clearable
            value={searchState.searchResource}
            placeholder="资源 / 实例"
            onChange={(value) => updateSearchState({ searchResource: value })}
          />
          <DateRangePicker
            className={style.datePicker}
            value={[searchState.startTime, searchState.endTime]}
            clearable
            allowInput
            format="YYYY-MM-DD HH:mm:ss"
            presets={presets}
            onChange={(value) => updateSearchState({
              startTime: String(value?.[0] || ''),
              endTime: String(value?.[1] || ''),
            })}
          />
          <div className={style.filterActions}>
            <Button theme="primary" icon={<SearchIcon />} loading={isLoading} onClick={() => fetchData()}>
              查询
            </Button>
            <Button variant="text" onClick={handleReset}>
              重置
            </Button>
          </div>
        </div>

        <Table
          data={displayEvents}
          columns={columns}
          rowKey="cursor"
          loading={isLoading}
          tableLayout="fixed"
          cellEmptyContent="-"
          empty={<Empty description="暂无事件数据" />}
          pagination={{
            pageSize: PAGE_SIZE,
            total: displayEvents.length,
            showJumper: true,
          }}
          onRowClick={({ row }: { row: EventMetricRow }) => setSelectedEvent(row)}
        />
      </section>

      <Drawer
        visible={Boolean(selectedEvent)}
        size="medium"
        header={selectedEvent ? resolveEventLabel(selectedEvent.event_type) : '事件详情'}
        onClose={() => setSelectedEvent(null)}
      >
        {selectedEvent && (
          <div className={style.detailBody}>
            <div className={style.detailStatus}>
              <Tag className={eventToneClass(selectedEvent.event_type)}>{resolveEventCategory(selectedEvent.event_type)}</Tag>
              <span>{formatTime(selectedEvent.event_time)}</span>
            </div>
            <dl className={style.detailGrid}>
              <dt>命名空间</dt>
              <dd>{selectedEvent.namespace || '-'}</dd>
              <dt>服务</dt>
              <dd>{selectedEvent.service || '-'}</dd>
              <dt>资源</dt>
              <dd>{selectedEvent.resource || '-'}</dd>
              <dt>上报节点</dt>
              <dd>{selectedEvent.server || '-'}</dd>
              <dt>游标</dt>
              <dd>{selectedEvent.cursor || '-'}</dd>
            </dl>
          </div>
        )}
      </Drawer>
    </div>
  );
}
