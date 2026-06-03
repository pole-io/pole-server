import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { modifyServices, createService, deleteServices, ServiceView, Service, CreateServicesRequest, ModifyServicesRequest, DescribeServicesRequest, describeServices, describeAllServices } from '../../services/service'


// State 和 Action 类型定义
export interface ServiceState {
    datas: ServiceView[]
    total: number
    page: number
    limit: number
    loading: boolean

    editSvc: Service | null
    viewSvc?: Service | null
}

const initialState: ServiceState = {
    datas: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    editSvc: null
};

export const listOneService = createAsyncThunk(`service/list_one`, async ({id}:{id: string}, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeServices({
            id: id,
            limit: 1,
            offset: 0,
         });
        return fulfillWithValue({
            viewSvc: res.list.length > 0 ? res.list[0] : null,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listAllServices = createAsyncThunk(`service/list_all`, async (_, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeAllServices();
        return fulfillWithValue({
            datas: res,
            total: res.length,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listServices = createAsyncThunk(`service/list`, async ({ param }: { param: DescribeServicesRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeServices(param);
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

export const saveServices = createAsyncThunk(`service/create`, async ({ param }: { param: CreateServicesRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createService([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateServices = createAsyncThunk(`service/update`, async ({ param }: { param: ModifyServicesRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyServices([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removeServices = createAsyncThunk(`service/remove`, async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteServices(ids.map((item) => ({ id: item })));
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

const serviceReducer = createSlice({
    name: 'service/edit',
    initialState,
    reducers: {
        editorService: (state, action: PayloadAction<Service>) => {
            state = {
                ...state,
                editSvc: {
                    ...action.payload
                }
            };
            return state;
        },
        resetService: (state) => {
            state = {
                ...state,
                editSvc: null,
            };
            return state;
        },
        cleanServicePage: (state) => {
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
            .addCase(listAllServices.fulfilled, (state, action) => {
                state.datas = action.payload.datas;
                state.total = action.payload.total;
                state.page = 1;
                state.limit = 10;
                state.loading = false;
            })
            .addCase(listAllServices.pending, (state) => {
                state.loading = true;
            })
            .addCase(listAllServices.rejected, (state) => {
                state.loading = false;
            })
            .addCase(listServices.fulfilled, (state, action) => {
                state.datas = action.payload.datas;
                state.total = action.payload.total;
                state.page = action.payload.page;
                state.limit = action.payload.limit;
                state.loading = false;
            })
            .addCase(listServices.pending, (state) => {
                state.loading = true;
            })
            .addCase(listServices.rejected, (state) => {
                state.loading = false;
            })
            .addCase(listOneService.fulfilled, (state, action) => {
                state.viewSvc = action.payload.viewSvc || null;
            });
    },
})

export const {
    editorService,
    resetService,
    cleanServicePage,
} = serviceReducer.actions;
export const selectService = (state: RootState) => state.discoveryService;

export default serviceReducer.reducer;
