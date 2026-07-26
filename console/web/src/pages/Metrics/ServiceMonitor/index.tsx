import React from 'react';
import { Button, Empty, Table, TableColumnData, Tag } from 'components/Fluent';
import { ArrowRightIcon, ServiceIcon } from 'components/Fluent/icons';
import QueryComposer from 'components/QueryComposer';
import { useNavigate } from 'react-router-dom';

import {
  GOVERNANCE_LABEL,
  GovernanceKind,
  SERVICE_SIGNALS,
  ServiceSignal,
  average,
  formatNumber,
  serviceKey,
  uniqueOptions,
} from './data';
import style from './index.module.less';

type ServiceOverview = {
  id: string;
  namespace: string;
  service: string;
  language: ServiceSignal['language'];
  apis: string[];
  instances: string[];
  kinds: GovernanceKind[];
  requests: number;
  governedRequests: number;
  blocked: number;
  p95: number;
  p99: number;
  errorRate: number;
  successRate: number;
  cpu: number;
  memory: number;
  status: ServiceSignal['status'];
  [key: string]: unknown;
};

const governanceKindOptions = Object.entries(GOVERNANCE_LABEL).map(([value, label]) => ({ value, label }));

function toneClass(status: ServiceSignal['status']) {
  if (status === 'critical') return style.toneDanger;
  if (status === 'warning') return style.toneWarning;
  return style.toneSuccess;
}

function governanceToneClass(kind: GovernanceKind) {
  if (kind === 'circuitbreaker' || kind === 'ratelimit') return style.toneWarning;
  if (kind === 'auth') return style.toneSuccess;
  return style.toneInfo;
}

function Panel(props: { title: string; subtitle?: string; children: React.ReactNode; className?: string }) {
  const { title, subtitle, children, className } = props;
  return (
    <section className={`${style.panel} ${className || ''}`}>
      <div className={style.panelHeader}>
        <div>
          <h2>{title}</h2>
          {subtitle && <span>{subtitle}</span>}
        </div>
      </div>
      {children}
    </section>
  );
}

function StatPanel(props: { title: string; value: React.ReactNode; hint: string; tone?: string }) {
  const { title, value, hint, tone } = props;
  return (
    <Panel title={title} className={`${style.statPanel} ${tone || ''}`}>
      <strong className={style.statValue}>{value}</strong>
      <span className={style.statHint}>{hint}</span>
    </Panel>
  );
}

function buildServiceOverview(rows: ServiceSignal[]): ServiceOverview[] {
  const groups = rows.reduce<Record<string, ServiceSignal[]>>((acc, row) => {
    const key = serviceKey(row);
    acc[key] = acc[key] || [];
    acc[key].push(row);
    return acc;
  }, {});

  return Object.entries(groups).map(([id, group]) => {
    const first = group[0];
    const warning = group.some((item) => item.status !== 'normal');
    return {
      id,
      namespace: first.namespace,
      service: first.service,
      language: first.language,
      apis: Array.from(new Set(group.map((item) => item.api))),
      instances: Array.from(new Set(group.map((item) => item.instance))),
      kinds: Array.from(new Set(group.map((item) => item.kind))),
      requests: group.reduce((sum, item) => sum + item.requests, 0),
      governedRequests: group.reduce((sum, item) => sum + Math.round(item.requests * item.hitRate / 100), 0),
      blocked: group.reduce((sum, item) => sum + item.blocked, 0),
      p95: Math.round(average(group, 'p95')),
      p99: Math.round(average(group, 'p99')),
      errorRate: Number(average(group, 'errorRate').toFixed(2)),
      successRate: Number(average(group, 'successRate').toFixed(2)),
      cpu: Math.round(average(group, 'cpu')),
      memory: Math.round(average(group, 'memory')),
      status: warning ? 'warning' : 'normal',
    };
  }).sort((a, b) => b.requests - a.requests);
}

export default function ServiceMonitor() {
  const pageRef = React.useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const [service, setService] = React.useState('');
  const [kind, setKind] = React.useState('');

  const rows = React.useMemo(() => SERVICE_SIGNALS.filter((item) => {
    const matchService = !service || serviceKey(item).toLocaleLowerCase().includes(service.toLocaleLowerCase());
    const matchKind = !kind || item.kind === kind;
    return matchService && matchKind;
  }), [kind, service]);

  const serviceRows = React.useMemo(() => buildServiceOverview(rows), [rows]);
  const totalRequests = rows.reduce((sum, item) => sum + item.requests, 0);
  const governedRequests = rows.reduce((sum, item) => sum + Math.round(item.requests * item.hitRate / 100), 0);
  const blocked = rows.reduce((sum, item) => sum + item.blocked, 0);
  const warningCount = serviceRows.filter((item) => item.status !== 'normal').length;
  const avgP95 = Math.round(average(rows, 'p95'));
  const avgP99 = Math.round(average(rows, 'p99'));

  React.useLayoutEffect(() => {
    pageRef.current?.scrollIntoView({ block: 'start' });
  }, []);

  const openDetail = React.useCallback((row: ServiceOverview) => {
    const params = new URLSearchParams({
      namespace: row.namespace,
      service: row.service,
    });
    navigate(`/metrics/service/detail?${params.toString()}`);
  }, [navigate]);

  const columns = React.useMemo<TableColumnData<ServiceOverview>[]>(() => [
    {
      colKey: 'service',
      title: '服务',
      width: 190,
      fixed: 'left',
      cell: ({ row }) => (
        <Button className={style.serviceCellButton} type="button" variant="text" onClick={(event) => { event.stopPropagation(); openDetail(row); }}>
          <strong>{row.namespace}/{row.service}</strong>
          <span>{row.language.toUpperCase()} · {row.instances.length} 实例 · {row.apis.length} 接口</span>
        </Button>
      ),
    },
    {
      colKey: 'requests',
      title: '请求量',
      width: 120,
      align: 'right',
      cell: ({ row }) => formatNumber(row.requests),
    },
    {
      colKey: 'governedRequests',
      title: '治理命中',
      width: 120,
      align: 'right',
      cell: ({ row }) => formatNumber(row.governedRequests),
    },
    {
      colKey: 'blocked',
      title: '拦截 / 降级',
      width: 120,
      align: 'right',
      cell: ({ row }) => formatNumber(row.blocked),
    },
    {
      colKey: 'latency',
      title: 'P95 / P99',
      width: 130,
      align: 'right',
      cell: ({ row }) => `${row.p95}ms / ${row.p99}ms`,
    },
    {
      colKey: 'resource',
      title: 'CPU / MEM',
      width: 118,
      align: 'right',
      cell: ({ row }) => `${row.cpu}% / ${row.memory}%`,
    },
    {
      colKey: 'kinds',
      title: '治理能力',
      width: 260,
      cell: ({ row }) => (
        <div className={style.tagRow}>
          {row.kinds.map((item) => (
            <Tag key={item} className={governanceToneClass(item)}>{GOVERNANCE_LABEL[item]}</Tag>
          ))}
        </div>
      ),
    },
    {
      colKey: 'status',
      title: '状态',
      width: 90,
      cell: ({ row }) => <Tag className={toneClass(row.status)}>{row.status === 'normal' ? '正常' : '关注'}</Tag>,
    },
    {
      colKey: 'action',
      title: '下钻',
      width: 92,
      fixed: 'right',
      cell: ({ row }) => (
        <Button size="small" variant="outline" icon={<ArrowRightIcon />} onClick={(event) => { event.stopPropagation(); openDetail(row); }}>
          查看
        </Button>
      ),
    },
  ], [openDetail]);

  const resetFilters = () => {
    setService('');
    setKind('');
  };

  return (
    <div className={style.page} ref={pageRef}>
      <section className={style.dashboardHeader}>
        <div className={style.headerTitle}>
          <span className={style.headerIcon}><ServiceIcon /></span>
          <div>
            <h1>服务监控</h1>
            <p>服务级整体概览，点击服务进入接口、实例和治理事件下钻分析。</p>
          </div>
        </div>
        <Tag className={style.sourceTagMock}>Mock 预览</Tag>
      </section>

      <div className={style.statGrid}>
        <StatPanel title="请求量" value={formatNumber(totalRequests)} hint="当前服务集合" />
        <StatPanel title="治理命中" value={formatNumber(governedRequests)} hint="路由、限流、鉴权等" />
        <StatPanel title="拦截 / 降级" value={formatNumber(blocked)} hint="限流、熔断、鉴权拒绝" tone={blocked ? style.panelWarning : undefined} />
        <StatPanel title="P95 / P99" value={`${avgP95}ms / ${avgP99}ms`} hint="服务调用延迟" />
        <StatPanel title="需关注服务" value={warningCount} hint="warning / critical 状态" tone={warningCount ? style.panelWarning : undefined} />
      </div>

      <section className={style.variableBar}>
        <QueryComposer
          keyword={service}
          keywordPlaceholder="搜索命名空间 / 服务"
          suggestions={uniqueOptions(SERVICE_SIGNALS, serviceKey).map((item) => ({
            label: String(item.label),
            value: String(item.value),
          }))}
          fields={[
            {
              key: 'kind',
              label: '治理能力',
              type: 'select',
              options: governanceKindOptions,
            },
          ]}
          values={{ kind }}
          onKeywordChange={setService}
          onValuesChange={(values) => setKind(String(values.kind || ''))}
          onReset={resetFilters}
          mode="instant"
        />
      </section>

      <Panel title="服务级总览" subtitle="点击行进入独立服务详情页，继续按接口、实例和治理能力下钻">
        <Table
          data={serviceRows}
          columns={columns}
          rowKey="id"
          tableLayout="fixed"
          cellEmptyContent="-"
          empty={<Empty description="暂无服务监控数据" />}
          pagination={false}
          onRowClick={({ row }: { row: ServiceOverview }) => openDetail(row)}
        />
      </Panel>
    </div>
  );
}
