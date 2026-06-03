import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { createUserGroup, deleteUserGroups, GroupRelation, modifyUserGroup, modifyUserGroupToken, refreshUserGroupToken } from 'services/user_group';

// State 和 Action 类型定义
export interface UserGroupState {
    // 用户组ID
    id?: string
    // 用户组名称
    name: string
    // 该用户组的授权Token
    auth_token?: string
    // 该用户组的授权Token是否可用
    token_enable?: boolean
    // 简单描述`
    comment: string
    // 该用户组下的用户ID列表信息
    relation: GroupRelation
    // 用户组标签
    metadata: Record<string, string>
}

const initialState: UserGroupState = {
    id: '',
    name: '',
    auth_token: '',
    token_enable: false,
    comment: '',
    relation: {
        group_id: '',
        users: []
    },
    metadata: {}
};

export const saveUserGroups = createAsyncThunk(`usergroup/create`, async ({ state }: { state: UserGroupState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createUserGroup([state]);
        return fulfillWithValue(state); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateUserGroups = createAsyncThunk(`usergroup/update`, async ({ state }: { state: UserGroupState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyUserGroup([state]);
        return fulfillWithValue(state); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const enableUserGroupToken = createAsyncThunk(`usergroup/token/enable`, async ({ state }: { state: UserGroupState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyUserGroupToken({id: state.id as string, token_enable: state.token_enable as boolean});
        return fulfillWithValue(state); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const resetUserGroupToken = createAsyncThunk(`usergroup/token/refresh`, async ({ state }: { state: UserGroupState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await refreshUserGroupToken({id: state.id as string});
        return fulfillWithValue(state); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removeUserGroups = createAsyncThunk(`usergroup/delete`, async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteUserGroups(ids.map(id => ({ id })));
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

const usergroupReducer = createSlice({
    name: 'user/edit',
    initialState,
    reducers: {
        editorUserGroup: (state, action: PayloadAction<UserGroupState>) => {
            state = {
                ...state,
                ...action.payload,
            };
            return state;
        },
        resetUserGroup: (state) => {
            return initialState;
        }
    }
});

export const { editorUserGroup, resetUserGroup } = usergroupReducer.actions;
export default usergroupReducer.reducer;
export const selectUserGroup = (state: RootState) => state.userGroups;
