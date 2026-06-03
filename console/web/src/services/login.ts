import request, {apiRequest, getApiRequest} from 'utils/request';
import { User } from './users';

export interface ILoginRequest {
    name: string;
    password: string;
}

export interface ILoginResponse {
    token: string
    name: string
    role: string
    user_id: string
    owner_id: string
}

/** 登录接口返回标准格式解包后即为 ILoginResponse */
export const doLogin = async (params: ILoginRequest) => {
    const response = await apiRequest<ILoginResponse>({ action: '/auth/v1/user/login', data: params });
    return response;
};


export interface InitAdminUserParams {
    /** 用户组ID */
    name: string
    password: string
  }

  export async function initAdminUser(params: InitAdminUserParams) {
    return await apiRequest<any>({ action: '/admin/v1/mainuser/create', data: params })
  }

  /** 解包后：存在时返回用户信息，否则无 data */
  export type DescribeAdminUserResult = User | undefined

  export async function checkExistAdminUser(opts?: { signal?: AbortSignal }) {
    return await getApiRequest<DescribeAdminUserResult>({ action: '/admin/v1/mainuser/exist', opts })
  }