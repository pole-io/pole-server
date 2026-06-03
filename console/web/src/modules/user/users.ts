import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { createUsers, CreateUsersRequest, deleteUsers, describeUsers, DescribeUsersRequest, ModifyUserRequest, modifyUsers, modifyUserToken, refreshUserToken, User } from 'services/users';

// State 和 Action 类型定义
export interface UserState {
    datas: User[];
    total: number;
    page: number;
    limit: number;
    loading: boolean;

    editUser: User | null;
    viewUser: User | null;
}

const initialState: UserState = {
    datas: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    editUser: null,
    viewUser: null
};

export const listOneUser = createAsyncThunk(`user/list_one`, async ({ id }: { id: string }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeUsers({
            id: id,
            limit: 1,
            offset: 0,
        });
        return fulfillWithValue({
            viewUser: res.totalCount > 0 ? res.content[0] : null,
        });
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listUsers = createAsyncThunk(`user/list`, async ({ param }: { param: DescribeUsersRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeUsers(param);
        return fulfillWithValue({
            datas: res.content,
            total: res.totalCount,
            page: Math.floor(param.offset / param.limit) + 1,
            limit: param.limit || 10
        });
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const saveUsers = createAsyncThunk(`user/create`, async ({ param }: { param: CreateUsersRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createUsers([param]);
        return fulfillWithValue(res);
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateUsers = createAsyncThunk(`user/update`, async ({ param }: { param: ModifyUserRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyUsers([param]);
        return fulfillWithValue(res);
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const enableUserToken = createAsyncThunk(`user/token/enable`, async ({ id, token_enable }: { id: string, token_enable: boolean }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyUserToken({ id: id, token_enable: token_enable });
        if (res) {
            return fulfillWithValue(res);
        }
        return rejectWithValue("fail");
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const resetUserToken = createAsyncThunk(`user/token/refresh`, async ({ id }: { id: string }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await refreshUserToken({ id });
        if (res) {
            return fulfillWithValue("ok");
        }
        return rejectWithValue("fail");
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removeUsers = createAsyncThunk(`user/delete`, async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteUsers(ids.map(id => ({ id })));
        if (res) {
            // 删除成功
            return fulfillWithValue(ids); // 返回删除的用户ID
        }
        // 删除失败
        return rejectWithValue('删除用户失败');
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

const userReducer = createSlice({
    name: 'user/edit',
    initialState,
    reducers: {
        editorUser: (state, action: PayloadAction<User>) => {
            state = {
                ...state,
                editUser: {
                    ...action.payload
                },
            };
            return state;
        },
        resetUser: (state) => {
            state = {
                ...state,
                editUser: null,
                viewUser: null,
            };
            return state;
        },
        viewUser: (state, action: PayloadAction<User>) => {
            state = {
                ...state,
                viewUser: {
                    ...action.payload
                },
            };
            return state;
        },
        cleanUserPage: (state) => {
            state = {
                ...state,
                datas: [],
                total: 0,
                page: 1,
                limit: 10,
                loading: false,
            };
            return state;
        }
    },
    extraReducers: (builder) => {
        builder
            .addCase(listUsers.fulfilled, (state, action) => {
                state.datas = action.payload.datas;
                state.total = action.payload.total;
                state.page = action.payload.page;
                state.limit = action.payload.limit;
                state.loading = false;
            })
            .addCase(listUsers.pending, (state) => {
                state.loading = true;
            })
            .addCase(listUsers.rejected, (state) => {
                state.loading = false;
            })
            .addCase(listOneUser.fulfilled, (state, action) => {
                state.viewUser = action.payload.viewUser;
            });
        },
});

export const {
    editorUser,
    resetUser,
    viewUser,
    cleanUserPage
} = userReducer.actions;
export default userReducer.reducer;
export const selectUser = (state: RootState) => state.userUsers;
