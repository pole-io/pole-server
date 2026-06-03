import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { ConfigFileGroup, ConfigFileGroupView, CreateConfigFileGroupRequest, createConfigFileGroups, DeleteConfigFileGroupRequest, deleteConfigFileGroups, describeAllConfigGroups, DescribeConfigFileGroupRequest, describeConfigFileGroups, ModifyConfigFileGroupRequest, modifyConfigFileGroups } from 'services/config_group';


// State 和 Action 类型定义
export interface ConfigGroupState {
    datas: ConfigFileGroupView[]
    total: number
    page: number
    limit: number
    loading: boolean

    editGroup: ConfigFileGroupView | null
}

const initialState: ConfigGroupState = {
    datas: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    editGroup: null
};

export const listAllConfigGroups = createAsyncThunk(`config_group/list_all`, async (_, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeAllConfigGroups();
        return fulfillWithValue({
            datas: res.list,
            total: res.totalCount,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listConfigGroups = createAsyncThunk(`config_group/list`, async ({ param }: { param: DescribeConfigFileGroupRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeConfigFileGroups(param);
        return fulfillWithValue({
            datas: res.list,
            total: res.totalCount,
            page: Math.floor(param.offset / param.limit) + 1,
            limit: param.limit || 10
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const saveConfigGroups = createAsyncThunk(`config_group/create`, async ({ param }: { param: CreateConfigFileGroupRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createConfigFileGroups([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateConfigGroups = createAsyncThunk(`config_group/update`, async ({ param }: { param: ModifyConfigFileGroupRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyConfigFileGroups([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removeConfigGroups = createAsyncThunk(`config_group/delete`, async ({ param }: { param: DeleteConfigFileGroupRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteConfigFileGroups([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});


const configgroupReducer = createSlice({
    name: 'config_group/edit',
    initialState,
    reducers: {
        editorConfigGroup: (state, action: PayloadAction<ConfigFileGroupView>) => {
            state = {
                ...state,
                editGroup: {
                    ...action.payload
                }
            };
            return state;
        },
        resetConfigGroup: (state) => {
            state = initialState;
            return state;
        },
        cleanConfigGroupPage: (state) => {
            state = {
                ...state,
                datas: [],
                total: 0,
                page: 1,
                limit: 10,
            };
            return state;
        }
    },
    extraReducers: (builder) => {
        builder.addCase(listAllConfigGroups.pending, (state) => {
            state.loading = true;
        });
        builder.addCase(listAllConfigGroups.rejected, (state) => {
            state.loading = false;
        });
        builder.addCase(listAllConfigGroups.fulfilled, (state, action) => {
            state.datas = action.payload.datas;
            state.total = action.payload.total;
            state.page = 1;
            state.limit = 10;
            state.loading = false;
        });
        builder.addCase(listConfigGroups.pending, (state) => {
            state.loading = true;
        });
        builder.addCase(listConfigGroups.rejected, (state) => {
            state.loading = false;
        });
        builder.addCase(listConfigGroups.fulfilled, (state, action) => {
            state.datas = action.payload.datas;
            state.total = action.payload.total;
            state.page = action.payload.page;
            state.limit = action.payload.limit;
            state.loading = false;
        });
    },
})

export const {
    editorConfigGroup,
    resetConfigGroup,
    cleanConfigGroupPage,
} = configgroupReducer.actions;
export const selectConfigGroup = (state: RootState) => state.configGroup;


export default configgroupReducer.reducer;