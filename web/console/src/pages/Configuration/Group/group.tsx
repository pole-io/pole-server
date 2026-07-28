import React from 'react'
import { Button, Table, Tag, Tooltip } from 'components/Fluent'
import { AddIcon, RefreshIcon } from 'components/Fluent/icons'
import { useNavigate, useSearchParams } from 'components/Router'

import QueryComposer, { QuerySnapshot } from 'components/QueryComposer'
import { ResourceHeader, ResourceToolbar } from 'components/ResourceLayout'
import ResourceNameLink from 'components/ResourceNameLink'
import { useAppDispatch } from 'modules/store'
import { editorConfigGroup, resetConfigGroup } from 'modules/configuration/group'
import {
  ConfigFileGroupView,
  ConfigGroupSummary,
  describeAllConfigGroups,
  describeSystemConfigGroups,
  summarizeConfigGroups,
} from 'services/config_group'
import { openErrNotification } from 'utils/notifition'
import ConfigGroupEditor from './ConfigGroupEditor'
import style from './index.module.less'

const pageSize = 10

const hasPendingRelease = (group: ConfigFileGroupView) => {
  const pending = Number(group.metadata?.pendingReleaseCount || 0)
  return pending > 0 || (group as any).status === 'to-be-released'
}

export default React.memo(() => {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const systemNamespace = searchParams.get('scope') === 'system'
    ? searchParams.get('namespace') || 'pole-system'
    : ''
  const [groups, setGroups] = React.useState<ConfigGroupSummary[]>([])
  const [loading, setLoading] = React.useState(false)
  const [page, setPage] = React.useState(1)
  const [creatorVisible, setCreatorVisible] = React.useState(false)
  const [filters, setFilters] = React.useState({ keyword: '', namespace: '', publishStatus: '' })

  const refresh = React.useCallback(async () => {
    setLoading(true)
    try {
      const result = systemNamespace
        ? await describeSystemConfigGroups(systemNamespace)
        : await describeAllConfigGroups()
      setGroups(summarizeConfigGroups(result.list))
    } catch (error) {
      openErrNotification('获取配置分组失败', error instanceof Error ? error.message : String(error))
    } finally {
      setLoading(false)
    }
  }, [systemNamespace])

  React.useEffect(() => {
    void refresh()
  }, [refresh])

  const namespaceOptions = React.useMemo(() => Array.from(new Set(
    groups.flatMap(group => group.environments.map(environment => environment.namespace)),
  )).sort().map(namespace => ({ label: namespace, value: namespace })), [groups])

  const filtered = React.useMemo(() => groups.filter((group) => {
    if (filters.keyword && !group.name.includes(filters.keyword)) return false
    const visibleEnvironments = filters.namespace
      ? group.environments.filter(environment => environment.namespace === filters.namespace)
      : group.environments
    if (!visibleEnvironments.length) return false
    if (filters.publishStatus === 'pending' && !visibleEnvironments.some(hasPendingRelease)) return false
    if (filters.publishStatus === 'clean' && visibleEnvironments.some(hasPendingRelease)) return false
    return true
  }), [filters, groups])

  const current = filtered.slice((page - 1) * pageSize, page * pageSize)
  const summary = React.useMemo(() => ({
    logical: groups.length,
    environments: groups.reduce((total, group) => total + group.namespaceCount, 0),
    files: groups.reduce((total, group) => total + group.fileCount, 0),
    pending: groups.reduce((total, group) => total + group.environments.filter(hasPendingRelease).length, 0),
  }), [groups])

  const openGroup = (group: ConfigGroupSummary) => {
    const preferred = filters.namespace
      ? group.environments.find(environment => environment.namespace === filters.namespace)
      : group.environments[0]
    if (!preferred) return
    dispatch(editorConfigGroup(preferred))
    navigate(`files?namespace=${encodeURIComponent(preferred.namespace)}&group=${encodeURIComponent(group.name)}`)
  }

  const columns = [
    {
      colKey: 'name',
      title: '配置分组',
      width: 260,
      cell: ({ row }: any) => <ResourceNameLink name={row.name} onClick={() => openGroup(row)} />,
    },
    { colKey: 'namespaceCount', title: '环境数', width: 120, cell: ({ row }: any) => `${row.namespaceCount} 个` },
    { colKey: 'fileCount', title: '配置文件', width: 130 },
    {
      colKey: 'pending',
      title: '待发布环境',
      width: 150,
      cell: ({ row }: any) => {
        const count = row.environments.filter(hasPendingRelease).length
        return count ? <Tag theme="warning" variant="light">{count} 个环境</Tag> : <Tag theme="success" variant="light">无</Tag>
      },
    },
    {
      colKey: 'environments',
      title: 'Namespace',
      cell: ({ row }: any) => row.environments.map((item: ConfigFileGroupView) => item.namespace).join('、'),
    },
    {
      colKey: 'action',
      title: '操作',
      width: 120,
      cell: ({ row }: any) => <Button variant="text" onClick={() => openGroup(row)}>查看环境</Button>,
    },
  ]

  const submit = ({ keyword, values }: QuerySnapshot) => {
    setPage(1)
    setFilters({
      keyword,
      namespace: String(values.namespace || ''),
      publishStatus: String(values.publishStatus || ''),
    })
  }

  return (
    <>
      <ResourceHeader
        eyebrow={systemNamespace ? '配置中心 / 系统空间' : '配置中心 / 配置分组'}
        title={systemNamespace ? `${systemNamespace} 配置资源` : '配置分组'}
        description={systemNamespace
          ? '显式维护当前控制面系统空间中的配置分组；这些配置不参与业务跨环境聚合。'
          : '先选择逻辑配置分组，再进入具体 Namespace 维护普通配置或模板配置。'}
        actions={(
          <>
            <Tooltip content="刷新列表">
              <Button shape="square" variant="outline" icon={<RefreshIcon />} onClick={() => void refresh()} />
            </Tooltip>
            {!systemNamespace && <Button theme="primary" icon={<AddIcon />} onClick={() => {
              dispatch(resetConfigGroup())
              setCreatorVisible(true)
            }}>新建配置分组</Button>}
          </>
        )}
      />
      <div className={style.summaryBar}>
        <div className={style.summaryItem}><span>逻辑分组</span><strong>{summary.logical}</strong></div>
        <div className={style.summaryItem}><span>环境实例</span><strong>{summary.environments}</strong></div>
        <div className={style.summaryItem}><span>配置文件</span><strong>{summary.files}</strong></div>
        <div className={`${style.summaryItem} ${style.warningStat}`}><span>待发布环境</span><strong>{summary.pending}</strong></div>
      </div>
      <ResourceToolbar
        title="配置分组列表"
        count={loading ? '正在同步列表' : `共 ${filtered.length} 个逻辑分组`}
        filters={(
          <QueryComposer
            keyword={filters.keyword}
            keywordPlaceholder="搜索配置分组名称"
            suggestions={groups.map(group => group.name)}
            fields={[
              { key: 'namespace', label: 'Namespace', type: 'select', options: namespaceOptions },
              {
                key: 'publishStatus',
                label: '发布状态',
                type: 'select',
                options: [
                  { label: '待发布', value: 'pending' },
                  { label: '无待发布', value: 'clean' },
                ],
              },
            ]}
            values={{ namespace: filters.namespace, publishStatus: filters.publishStatus }}
            onKeywordChange={keyword => setFilters(currentFilters => ({ ...currentFilters, keyword }))}
            onValuesChange={values => setFilters(currentFilters => ({
              ...currentFilters,
              namespace: String(values.namespace || ''),
              publishStatus: String(values.publishStatus || ''),
            }))}
            onSubmit={submit}
            onReset={() => {
              setPage(1)
              setFilters({ keyword: '', namespace: '', publishStatus: '' })
            }}
          />
        )}
      />
      <div className={style.groupTableSurface}>
        <Table
          data={current}
          columns={columns}
          loading={loading}
          rowKey="name"
          pagination={{
            current: page,
            pageSize,
            total: filtered.length,
            onChange: (info: any) => setPage(info.current || 1),
          }}
          empty="尚无配置分组。创建分组后，再进入 Namespace 新建普通配置或模板配置。"
        />
      </div>
      {creatorVisible && (
        <ConfigGroupEditor
          op="create"
          visible={creatorVisible}
          closeDrawer={() => {
            setCreatorVisible(false)
            void refresh()
          }}
        />
      )}
    </>
  )
})
