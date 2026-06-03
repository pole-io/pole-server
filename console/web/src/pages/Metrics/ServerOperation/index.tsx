import React, { memo } from 'react';
import { Button, Loading, PageInfo, PrimaryTableProps, TableProps, Table, Col, Space, Row, Select, Card, Input, TimeRangePicker, DateRangePicker, DateRangePickerProps } from 'tdesign-react';

import style from './index.module.less';
import { describeEventLog, describeOperationLog, EventType, getEventTypeInfo, getOperationTypeInfo, getResourceTypeInfo, OperationType, ResourceType } from 'services/observer';
import { openErrNotification } from 'utils/notifition';
import ErrorPage from 'components/ErrorPage';
import dayjs from 'dayjs';
import { FileCopyIcon } from 'tdesign-icons-react';
import { copyToClipboard } from 'utils/sys';

const ServerError = () => <ErrorPage code={500} />;

// resource_type: string;
// resource_name: string;
// namespace: string;
// operator: string;
// operation_type: string;
// detail: string;
// server: string; // 服务器名，用于区分生成记录的服务器
// happen_time: string; // 时间字符串
const columns: TableProps['columns'] = [
  {
    colKey: 'resource_type',
    title: '资源类型',
    cell: ({ row }) => {
      return getResourceTypeInfo(row.resource_type).label || '-';
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
    colKey: 'resource_name',
    title: '资源名称',
    cell: ({ row }) => {
      return row.resource_name || '-';
    }
  },
  {
    colKey: 'operator',
    title: '操作人',
    cell: ({ row }) => {
      return row.operator || '-';
    }
  },
  {
    colKey: 'operation_type',
    title: '操作类型',
    cell: ({ row }) => {
      return getOperationTypeInfo(row.operation_type)?.label || '-';
    }
  },
  {
    colKey: 'operation_detail',
    title: '详情',
    ellipsis: ({ row }) => (
      <div>
        {row.operation_detail}
        <FileCopyIcon
          style={{ cursor: 'pointer', marginLeft: '4px' }}
          onClick={() => copyToClipboard(row.operation_detail)}
        />
      </div>
    ),
  },
  {
    colKey: 'happen_time',
    title: '发生时间',
    cell: ({ row }) => {
      return row.happen_time || '-';
    },
  },
  {
    colKey: 'server',
    title: '来源',
    cell: ({ row }) => {
      return row.server || '-';
    }
  },
]

export interface IProps {
  // Define any props if needed
}

const ServerOperationTable: React.FC<IProps> = ({ }) => {

  // 合并编辑相关状态
  const [searchState, setSearchState] = React.useState<{
    operations: TableProps['data'];
    limit: number;
    cursor?: string;
    fetchError: boolean;
    isLoading: boolean;
    hasNext: boolean;

    searchEvent: string;
    searchOp: string;
    searchNamespace: string;
    searchResource: string;
    searchOperator: string;

    startTime?: string;
    endTime?: string;
  }>({
    operations: [], limit: 10, fetchError: false, isLoading: false, hasNext: true,
    searchEvent: '',
    searchOp: '',
    searchNamespace: '',
    searchResource: '',
    searchOperator: '',
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
      const response = await describeOperationLog({
        limit: pageSize,
        cursor: pageInfo.current,
        direction: 'next',
        namespace: searchState.searchNamespace,
        resource_name: searchState.searchResource,
        start_time: searchState.startTime || '',
        end_time: searchState.endTime || '',
      });
      setSearchState(s => ({ ...s, operations: response.data, hasNext: response.has_next, isLoading: false }));
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
          <Select label='资源类型' options={ResourceType}
            onChange={(value) => setSearchState(prev => ({ ...prev, searchEvent: value as string }))}
          />
          <Select label='操作类型' options={OperationType}
            onChange={(value) => setSearchState(prev => ({ ...prev, searchOp: value as string }))}
          />
          <Input label='命名空间'
            onChange={(value) => setSearchState(prev => ({ ...prev, searchNamespace: value }))}
          />
          <Input label='资源'
            onChange={(value) => setSearchState(prev => ({ ...prev, searchResource: value }))}
          />
          <Input label='操作人'
            onChange={(value) => setSearchState(prev => ({ ...prev, searchOperator: value }))}
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
        // tableLayout={'auto'}
        cellEmptyContent={'-'}
        data={searchState.operations}
        pagination={{
          total: searchState.hasNext ? (searchState.operations?.length || 0) + 1 : (searchState.operations?.length || 0),
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


export default React.memo(ServerOperationTable);
