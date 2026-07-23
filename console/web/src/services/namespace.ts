import { apiRequest, getAllList, getApiRequest, putApiRequest } from 'utils/request';
import { BaseURL } from './types';

export interface Namespace {
    name: string
    comment: string
    metadata: Record<string, string>
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

export interface DescribeNamespaceRequest {
    limit: number
    offset: number
    name?: string
    owners?: string
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
    const ns = list.map((item) => ({ ...item }))

    return { ...res, namespaces: ns, amount: res.amount ?? ns.length, size: res.size ?? ns.length }
}

export async function describeAllNamespaces() {
    const { list: namespaceList } = await getAllList(describeNamespaces, {
        listKey: 'namespaces',
        totalKey: 'amount',
    })({})

    return namespaceList.map((item) => ({ ...item })) as NamespaceView[]
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
