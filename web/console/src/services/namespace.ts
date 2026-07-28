import { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { BaseURL } from './types';

export type NamespaceKind = 'BUSINESS' | 'SYSTEM'

export interface Namespace {
    name: string
    comment: string
    metadata: Record<string, string>
    kind?: NamespaceKind
}

export interface NamespaceView extends Namespace {
    ctime: string
    mtime: string
    total_service_count?: number
    total_health_instance_count?: number
    total_instance_count?: number
    total_config_file_count?: number
    editable: boolean
    deleteable: boolean
}

const normalizeNamespaceKind = (item: NamespaceView): NamespaceKind => {
    const rawKind = item.kind as NamespaceKind | 'NAMESPACE_KIND_BUSINESS' | 'NAMESPACE_KIND_SYSTEM' | number | undefined
    if (rawKind === 'SYSTEM' || rawKind === 'NAMESPACE_KIND_SYSTEM' || rawKind === 1 || item.name === 'pole-system') {
        return 'SYSTEM'
    }
    return 'BUSINESS'
}

export const isBusinessNamespace = (item: Pick<NamespaceView, 'name' | 'kind'>) =>
    normalizeNamespaceKind(item as NamespaceView) === 'BUSINESS'

export interface DescribeNamespaceRequest {
    limit: number
    offset: number
    name?: string
    owners?: string
    kind?: 'business' | 'system'
}


export interface DescribeNamespacesResponse {
    amount: number
    size: number
    data?: Array<NamespaceView>
    namespaces?: Array<NamespaceView>
}

export async function describeNamespaces(params: DescribeNamespaceRequest) {
    const res = await getApiRequest<DescribeNamespacesResponse>({
        action: `${BaseURL.NAMESPACE}`,
        data: params,
    })

    const list = res.data ?? res.namespaces ?? []
    const ns = list.map((item) => ({ ...item, kind: normalizeNamespaceKind(item) }))

    return { ...res, namespaces: ns, amount: res.amount ?? ns.length, size: res.size ?? ns.length }
}

export async function describeAllNamespaces() {
    const { list: namespaceList } = await getAllList(describeNamespaces, {
        listKey: 'namespaces',
        totalKey: 'amount',
    })({})

    return namespaceList.map((item) => ({ ...item })) as NamespaceView[]
}

export async function describeBusinessNamespaces() {
    const { list } = await getAllList(describeNamespaces, {
        listKey: 'namespaces',
        totalKey: 'amount',
    })({ kind: 'business' })
    return list.filter(isBusinessNamespace) as NamespaceView[]
}

export async function describeSystemNamespaces() {
    const { list } = await getAllList(describeNamespaces, {
        listKey: 'namespaces',
        totalKey: 'amount',
    })({ kind: 'system' })
    return list.filter(item => !isBusinessNamespace(item)) as NamespaceView[]
}

export interface CreateNamespaceRequest {
    name: string
    comment: string
    metadata: Record<string, string>
}

export interface CreateNamespaceResponse {
    namespace: Namespace
}

export async function createNamespace(params: CreateNamespaceRequest[]) {
    const res = await apiRequest<CreateNamespaceResponse>({
        action: `${BaseURL.NAMESPACE}`,
        data: params,
    })
    return res
}

export interface ModifyNamespaceRequest {
    name: string
    comment?: string
    metadata: Record<string, string>
}

export interface ModifyNamespaceResponse {
    size: number
}

export async function modifyNamespace(params: ModifyNamespaceRequest[]) {
    const res = await putApiRequest<ModifyNamespaceResponse>({
        action: `${BaseURL.NAMESPACE}`,
        data: params,
    })
    return res
}

export interface DeleteNamespaceRequest {
    name: string
    token?: string
}

export interface DeleteNamespaceResponse {
    size: number
}

export async function deleteNamespace(params: DeleteNamespaceRequest[]) {
    const res = await apiRequest<DeleteNamespaceResponse>({
        action: `${BaseURL.NAMESPACE}/delete`,
        data: params,
    })

    return res
}
