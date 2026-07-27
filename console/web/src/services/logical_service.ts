import { apiRequest, getApiRequest, putApiRequest } from 'utils/request'
import { ServiceView } from './service'
import { BaseURL } from './types'

export interface LogicalServiceView {
  id: string
  name: string
  comment?: string
  owners?: string
  business?: string
  department?: string
  revision: string
  ctime?: string
  mtime?: string
  environment_count?: number
  total_instance_count?: number
  healthy_instance_count?: number
  editable?: boolean
  deleteable?: boolean
}

export interface ServiceEnvironmentBindingView {
  logical_service_id: string
  service_id: string
  namespace: string
  service_name: string
  ctime?: string
  mtime?: string
  service?: ServiceView
}

interface ListResponse<T> {
  amount?: number
  size?: number
  data?: T[]
}

export async function describeLogicalServices(params: { id?: string; name?: string; offset: number; limit: number }) {
  const response = await getApiRequest<ListResponse<LogicalServiceView>>({
    action: BaseURL.LOGICAL_SERVICE,
    data: params,
  })
  return { list: response.data ?? [], totalCount: response.amount ?? response.data?.length ?? 0 }
}

export async function describeAllLogicalServices(name?: string) {
  const list: LogicalServiceView[] = []
  const limit = 100
  let totalCount = 0
  do {
    const response = await describeLogicalServices({ name, offset: list.length, limit })
    list.push(...response.list)
    totalCount = response.totalCount
  } while (list.length < totalCount)
  return list
}

export async function createLogicalService(input: Pick<LogicalServiceView, 'name' | 'comment' | 'business' | 'department'>) {
  return apiRequest({ action: BaseURL.LOGICAL_SERVICE, data: [input] })
}

export async function updateLogicalService(input: LogicalServiceView) {
  return putApiRequest({ action: BaseURL.LOGICAL_SERVICE, data: [input] })
}

export async function deleteLogicalService(id: string) {
  return apiRequest({ action: `${BaseURL.LOGICAL_SERVICE}/delete`, data: [{ id }] })
}

export async function describeLogicalServiceEnvironments(logicalServiceId: string) {
  const response = await getApiRequest<ListResponse<ServiceEnvironmentBindingView>>({
    action: `${BaseURL.LOGICAL_SERVICE}/environments`,
    data: { logical_service_id: logicalServiceId },
  })
  return { list: response.data ?? [], totalCount: response.amount ?? response.data?.length ?? 0 }
}

export async function describeUnboundServiceEnvironments(params: {
  namespace?: string
  name?: string
  offset: number
  limit: number
}) {
  const response = await getApiRequest<ListResponse<ServiceView>>({
    action: `${BaseURL.LOGICAL_SERVICE}/unbound-environments`,
    data: params,
  })
  return { list: response.data ?? [], totalCount: response.amount ?? response.data?.length ?? 0 }
}

export async function bindServiceEnvironment(logicalServiceId: string, serviceId: string) {
  return apiRequest({
    action: `${BaseURL.LOGICAL_SERVICE}/environment-bindings`,
    data: { logical_service_id: logicalServiceId, service_id: serviceId },
  })
}

export async function unbindServiceEnvironment(logicalServiceId: string, serviceId: string) {
  return apiRequest({
    action: `${BaseURL.LOGICAL_SERVICE}/environment-bindings/delete`,
    data: { logical_service_id: logicalServiceId, service_id: serviceId },
  })
}

export async function resolveServiceEnvironmentBinding(serviceId: string) {
  return getApiRequest<ServiceEnvironmentBindingView | Record<string, never>>({
    action: `${BaseURL.LOGICAL_SERVICE}/environment-binding`,
    data: { service_id: serviceId },
  })
}
