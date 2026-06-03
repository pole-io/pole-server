import React, { memo } from 'react';
import { Button, Loading, PageInfo, PrimaryTableProps, TableProps, Table, Col, Space, Row, Select, Card, Input, TimeRangePicker, DateRangePicker, DateRangePickerProps } from 'tdesign-react';

import style from './index.module.less';
import { describeEventLog, EventType, getEventTypeInfo } from 'services/observer';
import { openErrNotification } from 'utils/notifition';
import ErrorPage from 'components/ErrorPage';
import dayjs from 'dayjs';

const ServerError = () => <ErrorPage code={500} />;

// Namespace string `json:"namespace"`
// Service   string `json:"service"`
// Resource  string `json:"resource"`
// EventType string `json:"event_type"`
// EventTime string `json:"event_time"`
const columns: PrimaryTableProps['columns'] = [
  {
    colKey: 'event_category',
    title: '事件类别',
    cell: ({ row }) => {
      return getEventTypeInfo(row.event_type)?.category || '-';
    }
  },
  {
    colKey: 'namespace',
    title: '命名空间',
    cell: ({ row }) => {
      return row.namespace || '-';
    }
  },
  {
    colKey: 'service',
    title: '服务',
    cell: ({ row }) => {
      return row.service || '-';
    }
  },
  {
    colKey: 'resource',
    title: '资源',
    cell: ({ row }) => {
      return row.resource || '-';
    }
  },
  {
    colKey: 'event_type',
    title: '事件类型',
    cell: ({ row }) => {
      return getEventTypeInfo(row.event_type)?.label || '-';
    }
  },
  {
    colKey: 'event_time',
    title: '事件时间',
    cell: ({ row }) => {
      return row.event_time || '-';
    },
  },
  {
    colKey: 'server',
    title: '来源',
    cell: ({ row }) => {
      return row.server || '-';
    },
  },
]

export interface IProps {
  // Define any props if needed
}

const ServerEventTable: React.FC<IProps> = ({ }) => {

  // 合并编辑相关状态
  const [searchState, setSearchState] = React.useState<{
    events: TableProps['data'];
    limit: number;
    fetchError: boolean;
    isLoading: boolean;
    hasNext: boolean;

    searchEvent: string;
    searchNamespace: string;
    searchService: string;
    searchResource: string;

    startTime?: string;
    endTime?: string;
  }>({
    events: [], limit: 10, fetchError: false, isLoading: false, hasNext: true,
    searchEvent: '',
    searchNamespace: '',
    searchService: '',
    searchResource: '',
  });

  const [presets] = React.useState<DateRangePickerProps['presets']>({
    最近7天: [dayjs().subtract(6, 'day').toDate(), dayjs().toDate()],
    最近3天: [dayjs().subtract(2, 'day').toDate(), dayjs().toDate()],
    今天: [dayjs().toDate(), dayjs().toDate()],
  });

  // 模拟远程请求
  async function fetchData(pageInfo: PageInfo) {
    setSearchState(s => ({ ...s, limit: pageInfo.pageSize, fetchError: false, isLoading: true }));
    try {
      const { current, pageSize } = pageInfo;
      // 请求可能存在跨域问题
      const response = await describeEventLog({
        limit: pageSize,
        cursor: pageInfo.current,
        direction: 'next',
        namespace: searchState.searchNamespace,
        service: searchState.searchService,
        resource: searchState.searchEvent,
        event_type: searchState.searchResource,
        start_time: searchState.startTime || '',
        end_time: searchState.endTime || '',
      });
      setSearchState(s => ({ ...s, events: response.data, hasNext: response.has_next, isLoading: false }));
    } catch (error: Error | any) {
      setSearchState(s => ({ ...s, fetchError: true, isLoading: false }));
      openErrNotification("获取数据失败", error);
    }
  }

  React.useEffect(() => {
    fetchData({ current: 1, pageSize: searchState.limit, previous: 0 });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const table = (
    <Card>
      <Row justify='space-between' className={style.toolBar}>
        <Space>
          <Select label='事件类型' options={EventType}
            onChange={(value) => setSearchState(prev => ({ ...prev, searchEvent: value as string }))}
          />
          <Input label='命名空间'
            onChange={(value) => setSearchState(prev => ({ ...prev, searchNamespace: value }))}
          />
          <Input label='服务'
            onChange={(value) => setSearchState(prev => ({ ...prev, searchService: value }))}
          />
          <Input label='资源'
            onChange={(value) => setSearchState(prev => ({ ...prev, searchResource: value }))}
          />
          <DateRangePicker
            clearable
            format="YYYY-MM-DD HH:mm:ss"
            valueType="YYYY-MM-DD HH:mm:ss"
            defaultValue={undefined}
            allowInput
            presets={presets}
            placeholder={['开始时间', '结束时间']}
            onChange={(value) => {
              if (value && value.length === 2) {
                const [start, end] = value;
                console.log('start', start, 'end', end);
                setSearchState(prev => ({
                  ...prev,
                  startTime: start ? start as string : '',
                  endTime: end ? end as string : '',
                }));
              } else {
                setSearchState(prev => ({ ...prev, startTime: '', endTime: '' }));
              }
            }}
          />
          <Button theme='primary' variant='base' onClick={() => fetchData({ current: 1, pageSize: searchState.limit, previous: 0 })}>
            查询
          </Button>
        </Space>
      </Row>
      <Table
        rowKey='cursor'
        columns={columns}
        loading={searchState.isLoading}
        size={"large"}
        tableLayout={'auto'}
        cellEmptyContent={'-'}
        data={searchState.events}
        pagination={{
          totalContent: false,
          pageSize: searchState.limit,
          showJumper: false,
          onChange(pageInfo) {
            fetchData(pageInfo);
          },
        }}
        selectOnRowClick={false}
      />
    </Card>
  )

  return (
    <div>
      {searchState.fetchError ? (
        <ServerError />
      ) : (
        <Loading loading={searchState.isLoading} className={style.loading}>
          {table}
        </Loading>
      )}
    </div>
  )
}


export default React.memo(ServerEventTable);
