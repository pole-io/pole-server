import React, { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Button,
  Drawer,
  Input,
  Select,
  TabPanel,
  Table,
  Tabs,
  Textarea,
} from 'components/Fluent'
import { useNavigate } from 'components/Router'

import { ResourceToolbar } from 'components/ResourceLayout'
import ResourceNameLink from 'components/ResourceNameLink'
import { ConfirmOperationButton, OperationButtonGroup } from 'components/OperationButton'
import {
  bindServiceEnvironment,
  createLogicalService,
  deleteLogicalService,
  describeAllLogicalServices,
  describeLogicalServices,
  describeUnboundServiceEnvironments,
  LogicalServiceView,
} from 'services/logical_service'
import { deleteServices, ServiceView } from 'services/service'
import { openErrNotification, openInfoNotification } from 'utils/notifition'
import style from './index.module.less'

const pageSize = 10

export interface ServicesTableHandle {
  refresh: () => void
  create: () => void
}

const ServicesTable = React.forwardRef<ServicesTableHandle>((_, ref) => {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState('logical')
  const [logicalServices, setLogicalServices] = useState<LogicalServiceView[]>([])
  const [logicalServiceOptions, setLogicalServiceOptions] = useState<LogicalServiceView[]>([])
  const [logicalTotal, setLogicalTotal] = useState(0)
  const [unboundServices, setUnboundServices] = useState<ServiceView[]>([])
  const [unboundTotal, setUnboundTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [creatorVisible, setCreatorVisible] = useState(false)
  const [bindingTarget, setBindingTarget] = useState<ServiceView | null>(null)
  const [selectedLogicalId, setSelectedLogicalId] = useState('')
  const [draft, setDraft] = useState({ name: '', comment: '', business: '', department: '' })

  const loadLogicalServices = useCallback(async (nextPage = page) => {
    const result = await describeLogicalServices({
      name: keyword || undefined,
      offset: (nextPage - 1) * pageSize,
      limit: pageSize,
    })
    setLogicalServices(result.list)
    setLogicalTotal(result.totalCount)
  }, [keyword, page])

  const loadUnboundServices = useCallback(async (nextPage = page) => {
    const result = await describeUnboundServiceEnvironments({
      name: keyword || undefined,
      offset: (nextPage - 1) * pageSize,
      limit: pageSize,
    })
    setUnboundServices(result.list)
    setUnboundTotal(result.totalCount)
  }, [keyword, page])

  const loadLogicalServiceOptions = useCallback(async () => {
    setLogicalServiceOptions(await describeAllLogicalServices())
  }, [])

  const refresh = useCallback(async (nextPage = page) => {
    setLoading(true)
    try {
      if (activeTab === 'logical') await loadLogicalServices(nextPage)
      else await Promise.all([loadUnboundServices(nextPage), loadLogicalServiceOptions()])
    } catch (error) {
      openErrNotification('获取服务失败', error instanceof Error ? error.message : String(error))
    } finally {
      setLoading(false)
    }
  }, [activeTab, loadLogicalServices, loadLogicalServiceOptions, loadUnboundServices, page])

  useEffect(() => {
    setPage(1)
    void refresh(1)
  }, [activeTab])

  React.useImperativeHandle(ref, () => ({
    refresh: () => void refresh(),
    create: () => setCreatorVisible(true),
  }))

  const metrics = useMemo(() => logicalServices.reduce((result, item) => ({
    environments: result.environments + Number(item.environment_count || 0),
    healthy: result.healthy + Number(item.healthy_instance_count || 0),
    total: result.total + Number(item.total_instance_count || 0),
  }), { environments: 0, healthy: 0, total: 0 }), [logicalServices])

  const create = async () => {
    if (!draft.name.trim()) {
      openErrNotification('无法创建', '请输入逻辑服务名称')
      return
    }
    try {
      await createLogicalService({ ...draft, name: draft.name.trim() })
      openInfoNotification('创建成功', '逻辑服务已创建')
      setCreatorVisible(false)
      setDraft({ name: '', comment: '', business: '', department: '' })
      await refresh(1)
    } catch (error) {
      openErrNotification('创建失败', error instanceof Error ? error.message : String(error))
    }
  }

  const bind = async () => {
    if (!bindingTarget?.id || !selectedLogicalId) return
    try {
      await bindServiceEnvironment(selectedLogicalId, bindingTarget.id)
      openInfoNotification('关联成功', `${bindingTarget.namespace}/${bindingTarget.name} 已加入逻辑服务`)
      setBindingTarget(null)
      setSelectedLogicalId('')
      await refresh()
    } catch (error) {
      openErrNotification('关联失败', error instanceof Error ? error.message : String(error))
    }
  }

  const removeLogicalService = async (row: LogicalServiceView) => {
    try {
      await deleteLogicalService(row.id)
      openInfoNotification('删除成功', `逻辑服务 ${row.name} 已删除，环境服务未受影响`)
      await refresh()
    } catch (error) {
      openErrNotification('删除逻辑服务失败', error instanceof Error ? error.message : String(error))
    }
  }

  const removeEnvironmentService = async (row: ServiceView) => {
    try {
      await deleteServices([{ id: row.id, namespace: row.namespace, name: row.name }])
      openInfoNotification('删除成功', `环境服务 ${row.namespace}/${row.name} 已删除`)
      await refresh()
    } catch (error) {
      openErrNotification('删除环境服务失败', error instanceof Error ? error.message : String(error))
    }
  }

  const logicalColumns = [
    {
      colKey: 'name',
      title: '逻辑服务',
      width: 260,
      cell: ({ row }: any) => (
        <ResourceNameLink name={row.name} onClick={() => navigate(`detail?id=${encodeURIComponent(row.id)}`)} />
      ),
    },
    {
      colKey: 'environment_count',
      title: '可访问环境',
      width: 130,
      cell: ({ row }: any) => `${row.environment_count || 0} 个`,
    },
    {
      colKey: 'health',
      title: '健康实例',
      width: 150,
      cell: ({ row }: any) => `${row.healthy_instance_count || 0}/${row.total_instance_count || 0}`,
    },
    {
      colKey: 'owner',
      title: '归属',
      width: 180,
      cell: ({ row }: any) => row.department || row.business || row.owners || '-',
    },
    {
      colKey: 'mtime',
      title: '更新时间',
      width: 180,
      cell: ({ row }: any) => row.mtime || '-',
    },
    {
      colKey: 'action',
      title: '操作',
      width: 140,
      cell: ({ row }: any) => (
        <OperationButtonGroup>
          <Button variant="text" onClick={() => navigate(`detail?id=${encodeURIComponent(row.id)}`)}>查看 / 编辑</Button>
          <ConfirmOperationButton
            action="delete"
            disabled={row.deleteable === false || Number(row.environment_count || 0) > 0}
            disabledLabel={row.deleteable === false ? '无权限操作' : '请先解除环境关联'}
            confirmContent={`确认删除逻辑服务 ${row.name} 吗？环境服务不会被删除。`}
            onConfirm={() => void removeLogicalService(row)}
          />
        </OperationButtonGroup>
      ),
    },
  ]

  const unboundColumns = [
    {
      colKey: 'name',
      title: '运行时服务',
      width: 260,
      cell: ({ row }: any) => (
        <ResourceNameLink
          name={row.name}
          onClick={() => navigate(`instance?namespace=${encodeURIComponent(row.namespace)}&service=${encodeURIComponent(row.name)}`)}
        />
      ),
    },
    { colKey: 'namespace', title: '命名空间', width: 180 },
    {
      colKey: 'health',
      title: '健康实例',
      width: 150,
      cell: ({ row }: any) => `${row.healthy_instance_count || 0}/${row.total_instance_count || 0}`,
    },
    {
      colKey: 'action',
      title: '操作',
      width: 220,
      cell: ({ row }: any) => (
        <OperationButtonGroup>
          <Button variant="text" onClick={() => navigate(`instance?namespace=${encodeURIComponent(row.namespace)}&service=${encodeURIComponent(row.name)}`)}>进入环境</Button>
          <Button variant="text" onClick={() => setBindingTarget(row)}>关联</Button>
          <ConfirmOperationButton
            action="delete"
            disabled={row.deleteable === false}
            disabledLabel="无权限操作"
            confirmContent={`确认删除环境服务 ${row.namespace}/${row.name} 吗？`}
            onConfirm={() => void removeEnvironmentService(row)}
          />
        </OperationButtonGroup>
      ),
    },
  ]

  const currentTotal = activeTab === 'logical' ? logicalTotal : unboundTotal
  const currentData = activeTab === 'logical' ? logicalServices : unboundServices

  return (
    <div className={style.workspace}>
      <section className={style.metricRail}>
        <div className={style.metricItem}><span>逻辑服务</span><strong>{logicalTotal}</strong></div>
        <div className={style.metricItem}><span>环境服务</span><strong>{metrics.environments}</strong></div>
        <div className={style.metricItem}><span>健康实例</span><strong>{metrics.healthy}/{metrics.total}</strong></div>
      </section>
      <Tabs value={activeTab} onChange={(value) => setActiveTab(String(value))}>
        <TabPanel value="logical" label="逻辑服务" />
        <TabPanel value="unbound" label={`未关联环境服务${unboundTotal ? ` (${unboundTotal})` : ''}`} />
      </Tabs>
      <section className={style.listSection}>
        <ResourceToolbar
          title={activeTab === 'logical' ? '逻辑服务清单' : '未关联环境服务'}
          count={loading ? '正在同步列表' : `共 ${currentTotal} 条`}
          filters={(
            <>
              <Input value={keyword} placeholder={activeTab === 'logical' ? '搜索逻辑服务' : '搜索运行时服务'} onChange={setKeyword} />
              <Button variant="outline" onClick={() => { setPage(1); void refresh(1) }}>查询</Button>
            </>
          )}
        />
        <section className={`${style.tableSurface} ${style.serviceTableSurface}`}>
          <Table
            data={currentData}
            columns={activeTab === 'logical' ? logicalColumns : unboundColumns}
            loading={loading}
            rowKey="id"
            pagination={{
              current: page,
              pageSize,
              total: currentTotal,
              onChange: (info: any) => {
                const nextPage = info.current || 1
                setPage(nextPage)
                void refresh(nextPage)
              },
            }}
            empty={activeTab === 'logical'
              ? '尚无逻辑服务。先创建逻辑服务，再显式关联已有环境服务。'
              : '所有可访问环境服务均已关联。'}
          />
        </section>
      </section>

      <Drawer
        visible={creatorVisible}
        header="新建逻辑服务"
        size="min(640px, 94vw)"
        destroyOnClose
        onClose={() => setCreatorVisible(false)}
        footer={<><Button onClick={() => setCreatorVisible(false)}>取消</Button><Button theme="primary" onClick={() => void create()}>创建</Button></>}
      >
        <div className={style.logicalForm}>
          <label><span>逻辑服务名称</span><Input value={draft.name} onChange={name => setDraft(current => ({ ...current, name }))} /></label>
          <label><span>业务</span><Input value={draft.business} onChange={business => setDraft(current => ({ ...current, business }))} /></label>
          <label><span>部门</span><Input value={draft.department} onChange={department => setDraft(current => ({ ...current, department }))} /></label>
          <label><span>说明</span><Textarea value={draft.comment} onChange={comment => setDraft(current => ({ ...current, comment }))} /></label>
        </div>
      </Drawer>

      <Drawer
        visible={Boolean(bindingTarget)}
        header="关联到逻辑服务"
        size="min(560px, 94vw)"
        destroyOnClose
        onClose={() => setBindingTarget(null)}
        footer={<><Button onClick={() => setBindingTarget(null)}>取消</Button><Button theme="primary" disabled={!selectedLogicalId} onClick={() => void bind()}>确认关联</Button></>}
      >
        <div className={style.logicalForm}>
          <p>环境服务：{bindingTarget?.namespace}/{bindingTarget?.name}</p>
          <label>
            <span>目标逻辑服务</span>
            <Select
              value={selectedLogicalId}
              placeholder="选择逻辑服务"
              options={logicalServiceOptions.map(item => ({ label: item.name, value: item.id }))}
              onChange={setSelectedLogicalId}
            />
          </label>
        </div>
      </Drawer>
    </div>
  )
})

export default React.memo(ServicesTable)
