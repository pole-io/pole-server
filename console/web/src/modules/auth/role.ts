import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { PolicyResources, Principals } from 'services/auth_policy';
import { User } from 'services/users';
import { UserGroup } from 'services/user_group';
import { createRoles, deleteRoles, modifyRoles } from 'services/role';


// State 和 Action 类型定义
export interface RoleState {
    // 角色ID
    id: string;
    // 角色名称
    name: string;
    // 默认角色
    default_role?: boolean;
    // 备注
    comment: string;
    // 角色来源
    source?: string;
    // 角色标签
    metadata: Record<string, string>;
    // 角色对应的用户列表
    users: { id: string, name?: string }[];
    // 角色对应的用户组列表
    user_groups: { id: string, name?: string }[];
}

const initialState: RoleState = {
    id: '',
    name: '',
    comment: '',
    metadata: {},
    default_role: false,
    users: [],
    user_groups: [],
};

export const removeRoles = createAsyncThunk(`role/delete`, async ({ state }: { state: {id: string}[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteRoles(state);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});


export const saveRoles = createAsyncThunk(`role/create`, async ({ state }: { state: RoleState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createRoles([{...state, source: 'pole.io'}]);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateRoles = createAsyncThunk(`role/update`, async ({ state }: { state: RoleState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyRoles([{...state, source: 'pole.io'}]);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});


const roleReducer = createSlice({
    name: 'role/edit',
    initialState,
    reducers: {
        editorRoles: (state, action: PayloadAction<RoleState>) => {
            state = {
                ...state,
                ...action.payload
            };
            return state;
        },
        resetRoles: (state) => {
            state = {
                ...initialState
            };
            return state;
        }
    },
    extraReducers: () => { },
})

export const {
    editorRoles,
    resetRoles,
} = roleReducer.actions;
export const selectRole = (state: RootState) => state.authRoles;

export default roleReducer.reducer;
