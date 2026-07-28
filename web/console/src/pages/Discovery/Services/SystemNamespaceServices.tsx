import React from 'react'
import { Button, Input, Table } from 'components/Fluent'
import { useNavigate } from 'components/Router'

import { ResourceToolbar } from 'components/ResourceLayout'
import ResourceNameLink from 'components/ResourceNameLink'
import { ConfirmOperationButton, OperationButtonGroup } from 'components/OperationButton'
import { deleteServices, describeServices, ServiceView } from 'services/service'
import { openErrNotification, openInfoNotification } from 'utils/notifition'
import style from './index.module.less'

const pageSize = 10

export default function SystemNamespaceServices({ namespace }: { namespace: string }) {
  const navigate = useNavigate()
  const [services, setServices] = React.useState<ServiceView[]>([])
  const [total, setTotal] = React.useState(0)
  const [page, setPage] = React.useState(1)
  const [keyword, setKeyword] = React.useState('')
  const [loading, setLoading] = React.useState(false)

  const refresh = React.useCallback(async (nextPage = page) => {
    setLoading(true)
    try {
      const result = await describeServices({
        namespace,
        name: keyword || undefined,
        offset: (nextPage - 1) * pageSize,
        limit: pageSize,
      })
      setServices(result.list)
      setTotal(result.totalCount)
    } catch (error) {
      openErrNotification('获取系统服务失败', error instanceof Error ? error.message : String(error))
    } finally {
      setLoading(false)
    }
  }, [keyword, namespace, page])

  React.useEffect(() => {
    void refresh(1)
  }, [namespace])

  const remove = async (row: ServiceView) => {
    try {
      await deleteServices([{ id: row.id, namespace: row.namespace, name: row.name }])
      openInfoNotification('删除成功', `系统服务 ${row.namespace}/${row.name} 已删除`)
      await refresh()
    } catch (error) {
      openErrNotification('删除系统服务失败', error instanceof Error ? error.message : String(error))
    }
  }

  const columns = [
    {
      colKey: 'name',
      title: '系统服务',
      width: 280,
      cell: ({ row }: any) => (
        <ResourceNameLink
          name={row.name}
          onClick={() => navigate(
            `/discovery/service/instance?namespace=${encodeURIComponent(namespace)}`
            + `&service=${encodeURIComponent(row.name)}`,
          )}
        />
      ),
    },
    { colKey: 'namespace', title: '系统空间', width: 180 },
    {
      colKey: 'health',
      title: '健康实例',
      width: 150,
      cell: ({ row }: any) => `${row.healthy_instance_count || 0}/${row.total_instance_count || 0}`,
    },
    { colKey: 'mtime', title: '更新时间', width: 180 },
    {
      colKey: 'action',
      title: '操作',
      width: 160,
      cell: ({ row }: any) => (
        <OperationButtonGroup>
          <Button variant="text" onClick={() => navigate(
            `/discovery/service/instance?namespace=${encodeURIComponent(namespace)}`
            + `&service=${encodeURIComponent(row.name)}`,
          )}>查看实例</Button>
          <ConfirmOperationButton
            action="delete"
            label="删除系统服务"
            disabled={row.deleteable === false}
            disabledLabel="无权限操作"
            confirmContent={`确认删除系统服务 ${row.namespace}/${row.name} 吗？`}
            onConfirm={() => void remove(row)}
          />
        </OperationButtonGroup>
      ),
    },
  ]

  return (
    <section className={style.listSection}>
      <ResourceToolbar
        title={`${namespace} 服务`}
        count={loading ? '正在同步列表' : `共 ${total} 条`}
        filters={(
          <>
            <Input value={keyword} placeholder="搜索系统服务" onChange={setKeyword} />
            <Button variant="outline" onClick={() => { setPage(1); void refresh(1) }}>查询</Button>
          </>
        )}
      />
      <section className={`${style.tableSurface} ${style.serviceTableSurface}`}>
        <Table
          data={services}
          columns={columns}
          loading={loading}
          rowKey="id"
          pagination={{
            current: page,
            pageSize,
            total,
            onChange: (info: any) => {
              const nextPage = info.current || 1
              setPage(nextPage)
              void refresh(nextPage)
            },
          }}
          empty="当前系统空间没有可见服务。"
        />
      </section>
    </section>
  )
}
