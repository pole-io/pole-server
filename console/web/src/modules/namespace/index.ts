import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { modifyNamespace, createNamespace, NamespaceView, Namespace, CreateNamespaceRequest, ModifyNamespaceRequest, DeleteNamespaceRequest, deleteNamespace, describeNamespaces, DescribeNamespaceRequest, describeAllNamespaces } from '../../services/namespace'
import { toRequestErrorPayload } from '../../utils/request';

// State 和 Action 类型定义
export interface NamespaceState {
    datas: NamespaceView[]
    total: number
    page: number
    limit: number
    loading: boolean

    editNs: Namespace | null
}

const initialState: NamespaceState = {
    datas: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    editNs: null
};

export const listAllNamespaces = createAsyncThunk(`namespace/list_all`, async (_, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeAllNamespaces();
        return fulfillWithValue({
            datas: res,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue(toRequestErrorPayload(error));
    }
});

export const listNamespaces = createAsyncThunk(`namespace/list`, async ({ param }: { param: DescribeNamespaceRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeNamespaces(param);
        return fulfillWithValue({
            datas: res.namespaces,
            total: res.amount,
            page: Math.floor(param.offset / param.limit) + 1,
            limit: param.limit || 10
        }); // 返回 token
    } catch (error) {
        return rejectWithValue(toRequestErrorPayload(error));
    }
});

export const saveNamespace = createAsyncThunk(`namespace/create`, async ({ param }: { param: CreateNamespaceRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createNamespace([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue(toRequestErrorPayload(error));
    }
});

export const updateNamespace = createAsyncThunk(`namespace/update`, async ({ param }: { param: ModifyNamespaceRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyNamespace([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue(toRequestErrorPayload(error));
    }
});

export const removeNamespace = createAsyncThunk(`namespace/delete`, async ({ param }: { param: DeleteNamespaceRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteNamespace([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue(toRequestErrorPayload(error));
    }
});

const namespaceReducer = createSlice({
    name: 'namespace/edit',
    initialState,
    reducers: {
        editorNamespace: (state, action: PayloadAction<Namespace>) => {
            state = {
                ...state,
                editNs: {
                    ...action.payload
                }
            };
            return state;
        },
        resetNamespace: (state) => {
            state = {
                ...state,
                editNs: null,
            };
            return state;
        },
        cleanNamespacePage: (state) => {
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
            .addCase(listAllNamespaces.fulfilled, (state, action) => {
                state.datas = action.payload.datas;
                state.loading = false;
            })
            .addCase(listAllNamespaces.pending, (state) => {
                state.loading = true;
            })
            .addCase(listAllNamespaces.rejected, (state) => {
                state.loading = false;
            })
            .addCase(listNamespaces.fulfilled, (state, action) => {
                state.datas = action.payload.datas;
                state.total = action.payload.total;
                state.page = action.payload.page;
                state.limit = action.payload.limit;
                state.loading = false;
            })
            .addCase(listNamespaces.pending, (state) => {
                state.loading = true;
            })
            .addCase(listNamespaces.rejected, (state) => {
                state.loading = false;
            });
    },
});

export const selectNamespace = (state: RootState) => state.namespace;

export const {
    editorNamespace,
    resetNamespace,
    cleanNamespacePage,
} = namespaceReducer.actions;

export default namespaceReducer.reducer;
