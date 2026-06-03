import { createSlice, createAsyncThunk, current } from '@reduxjs/toolkit';
import { doLogin } from '../../services/login'
import { LoginUserIdKey, PoleTokenKey } from 'utils/request';

const namespace = 'user';
export const LoginRoleKey = 'login-role'
export const LoginUserOwnerIdKey = 'login-owner-id'
export const LoginUserNameKey = 'login-name'

const initialState = {
  isLogin: !!localStorage.getItem(PoleTokenKey), // 是否登录
  currentUser: {
    name: localStorage.getItem(LoginUserNameKey) || '', // 用户名
    role: localStorage.getItem(LoginRoleKey) || '', // 角色
    user_id: localStorage.getItem(LoginUserIdKey) || '', // 用户ID
    owner_id: localStorage.getItem(LoginUserOwnerIdKey) || '', // 所属ID
  }
};

// login
export const login = createAsyncThunk(`${namespace}/login`, async ({ username, password }: { username: string; password: string }, { fulfillWithValue, rejectWithValue }) => {
  try {
    const res = await doLogin({ name: username, password: password });
    localStorage.setItem(PoleTokenKey, res.token);
    localStorage.setItem(LoginUserNameKey, res.name);
    localStorage.setItem(LoginRoleKey, res.role);
    localStorage.setItem(LoginUserIdKey, res.user_id);
    localStorage.setItem(LoginUserOwnerIdKey, res.owner_id);
    return fulfillWithValue(res); // 返回 token
  } catch (error) {
    return rejectWithValue((error as Error).message); // 捕获错误并返回
  }
});

const loginReducer = createSlice({
  name: namespace,
  initialState,
  reducers: {
    logout: (state) => {
      localStorage.removeItem(PoleTokenKey);
      localStorage.removeItem(LoginUserNameKey);
      localStorage.removeItem(LoginRoleKey);
      localStorage.removeItem(LoginUserIdKey);
      localStorage.removeItem(LoginUserOwnerIdKey);
      // 清空当前用户信息
      state.isLogin = false;
      state.currentUser = {
        name: '',
        role: '',
        user_id: '',
        owner_id: '',
      };
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.fulfilled, (state, action) => {
        state.isLogin = true;
        state.currentUser = {
          name: action.payload.name,
          role: action.payload.role,
          user_id: action.payload.user_id,
          owner_id: action.payload.owner_id,
        };
      })
      .addCase(login.rejected, (state, action) => {
        state.isLogin = false;
      });
  },
});

export const { logout } = loginReducer.actions;
export default loginReducer.reducer;
