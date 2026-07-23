import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { describeConsoleSession, doLogin } from '../../services/login'
import { LoginUserIdKey, PoleTokenKey } from 'utils/request';

const namespace = 'user';
export const LoginRoleKey = 'login-role'
export const LoginUserOwnerIdKey = 'login-owner-id'
export const LoginUserNameKey = 'login-name'

const hasPersistedLogin = !!localStorage.getItem(PoleTokenKey);

const initialState = {
  isLogin: hasPersistedLogin, // 是否存在待服务端确认的登录会话
  sessionResolved: !hasPersistedLogin,
  currentUser: {
    name: localStorage.getItem(LoginUserNameKey) || '', // 用户名
    role: '', // 角色只能从服务端签名会话恢复
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
    localStorage.removeItem(LoginRoleKey);
    localStorage.setItem(LoginUserIdKey, res.user_id);
    localStorage.setItem(LoginUserOwnerIdKey, res.owner_id);
    return fulfillWithValue(res); // 返回 token
  } catch (error) {
    return rejectWithValue((error as Error).message); // 捕获错误并返回
  }
});

export const hydrateSession = createAsyncThunk(`${namespace}/session`, async (_, { rejectWithValue }) => {
  try {
    return await describeConsoleSession();
  } catch (error) {
    return rejectWithValue((error as Error).message);
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
      state.sessionResolved = true;
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
        state.sessionResolved = true;
        state.currentUser = {
          name: action.payload.name,
          role: action.payload.role,
          user_id: action.payload.user_id,
          owner_id: action.payload.owner_id,
        };
      })
      .addCase(login.rejected, (state, action) => {
        state.isLogin = false;
        state.sessionResolved = true;
      })
      .addCase(hydrateSession.pending, (state) => {
        state.sessionResolved = false;
        state.currentUser.role = '';
      })
      .addCase(hydrateSession.fulfilled, (state, action) => {
        state.isLogin = true;
        state.sessionResolved = true;
        state.currentUser.user_id = action.payload.user_id;
        state.currentUser.role = action.payload.role;
      })
      .addCase(hydrateSession.rejected, (state) => {
        state.isLogin = false;
        state.sessionResolved = true;
        state.currentUser.role = '';
      });
  },
});

export const { logout } = loginReducer.actions;
export default loginReducer.reducer;
