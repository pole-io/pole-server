import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { modifyServiceAlias, createServiceAlias, ServiceAliasView, deleteServiceAlias, DeleteServiceAliasRequest, ModifyServiceAliasRequest, CreateServiceAliasRequest, DescribeServiceAliasRequest, describeServiceAlias, ServiceAlias } from '../../services/alias'


// State 和 Action 类型定义
export interface ServiceAliasState {
    datas: ServiceAliasView[]
    total: number
    page: number
    limit: number
    loading: boolean

    editAlias: ServiceAliasView | null
}

const initialState: ServiceAliasState = {
    datas: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    editAlias: null
};

export const listServiceAliass = createAsyncThunk(`service_alias/list`, async ({ param }: { param: DescribeServiceAliasRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeServiceAlias(param);
        return fulfillWithValue({
            datas: res.content,
            total: res.totalCount,
            page: Math.floor(param.offset / param.limit) + 1,
            limit: param.limit || 10
        }); // 返回数据
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const saveServiceAliass = createAsyncThunk(`service_alias/create`, async ({ param }: { param: CreateServiceAliasRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createServiceAlias(param);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateServiceAliass = createAsyncThunk(`service_alias/update`, async ({ param }: { param: ModifyServiceAliasRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyServiceAlias(param);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removeServiceAliass = createAsyncThunk(`service_alias/remove`, async ({ param }: { param: DeleteServiceAliasRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteServiceAlias(param);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

const serviceAliasReducer = createSlice({
    name: 'service_alias/edit',
    initialState,
    reducers: {
        editorServiceAlias: (state, action: PayloadAction<ServiceAlias>) => {
            state = {
                ...state,
                editAlias: {
                    ...action.payload
                }
            };
            return state;
        },
        resetServiceAlias: (state) => {
            state = {
                ...state,
                editAlias: null,
            };
            return state;
        },
        cleanAliasPage: (state) => {
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
            .addCase(listServiceAliass.pending, (state) => {
                state.loading = true;
            })
            .addCase(listServiceAliass.fulfilled, (state, action) => {
                state.loading = false;
                state.datas = action.payload.datas;
                state.total = action.payload.total;
                state.page = action.payload.page;
                state.limit = action.payload.limit;
            })
    },
})

export const {
    editorServiceAlias,
    resetServiceAlias,
    cleanAliasPage,
} = serviceAliasReducer.actions;
export const selectServiceAlias = (state: RootState) => state.discoveryServiceAlais;

export default serviceAliasReducer.reducer;
