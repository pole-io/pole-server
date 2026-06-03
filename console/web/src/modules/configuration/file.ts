import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { ConfigFileView, CreateConfigFileRequest, createConfigFiles, deleteConfigFiles, describeAllConfigFiles, DescribeAllConfigFilesRequest, describeConfigFiles, DescribeConfigFilesRequest, describeEncryptAlgo, describeOneConfigFile, DescribeOneConfigFileRequest, ModifyConfigFileRequest, modifyConfigFiles } from 'services/config_files';
import { DeleteFileReleaseRequest, describeFileSubscribers, DescribeFileSubscribersRequest, VersionClient } from 'services/config_release';

// State 和 Action 类型定义
export interface FileState {
    datas: ConfigFileView[]
    total: number
    page: number
    limit: number
    loading: boolean

    subscribers: VersionClient[]

    cryptoAlgos: string[]

    editFile: ConfigFileView | null
    viewFile: ConfigFileView | null
}

const initialState: FileState = {
    datas: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    subscribers: [],

    cryptoAlgos: [],

    editFile: null,
    viewFile: null
};

export const listFileSubscribers = createAsyncThunk(`config_file/list_subscribers`, async ({ param }: { param: DescribeFileSubscribersRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeFileSubscribers(param);
        return fulfillWithValue({
            subscribers: res.clients,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listConfigFileCryptoAlgos = createAsyncThunk(`config_file/list_crytptos`, async (_, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeEncryptAlgo();
        return fulfillWithValue({
            cryptoAlgos: res.algorithms,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listOneConfigFile = createAsyncThunk(`config_file/list_one`, async ({ param }: { param: DescribeOneConfigFileRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeOneConfigFile(param);
        return fulfillWithValue({
            viewFile: res.configFile,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listAllConfigFiles = createAsyncThunk(`config_file/list_all`, async ({ param }: { param: DescribeAllConfigFilesRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeAllConfigFiles(param);
        return fulfillWithValue({
            data: res,
            total: res.length,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listConfigFiles = createAsyncThunk(`config_file/list`, async ({ param }: { param: DescribeConfigFilesRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeConfigFiles(param);
        return fulfillWithValue({
            data: res.list,
            total: res.totalCount,
            page: Math.floor(param.offset / param.limit) + 1,
            limit: param.limit || 10
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const saveConfigFiles = createAsyncThunk(`config_file/create`, async ({ param }: { param: CreateConfigFileRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await createConfigFiles([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updateConfigFiles = createAsyncThunk(`config_file/update`, async ({ param }: { param: ModifyConfigFileRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await modifyConfigFiles([param]);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removeConfigFeils = createAsyncThunk(`config_file/delete`, async ({ param }: { param: DeleteFileReleaseRequest[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteConfigFiles(param);
        return fulfillWithValue(res); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

const configFileReducer = createSlice({
    name: 'config_file/edit',
    initialState,
    reducers: {
        editorConfigFile: (state, action: PayloadAction<ConfigFileView>) => {
            state = {
                ...state,
                editFile: {
                    ...action.payload
                }
            };
            return state;
        },
        resetConfigFile: (state) => {
            state = {
                ...state,
                editFile: null,
            };
            return state;
        },
        cleanFilePage: (state) => {
            state = {
                ...state,
                datas: [],
                total: 0,
                page: 1,
                limit: 10,
                loading: false,
                subscribers: [],
                cryptoAlgos: [],
            };
            return state;
        }
    },
    extraReducers: (builder) => {
        builder
            .addCase(listAllConfigFiles.fulfilled, (state, action) => {
                state = {
                    ...state,
                    datas: action.payload.data,
                    total: action.payload.total,
                    loading: false,
                };
                return state;
            })
            .addCase(listAllConfigFiles.pending, (state) => {
                state = {
                    ...state,
                    loading: true,
                };
                return state;
            })
            .addCase(listAllConfigFiles.rejected, (state, action) => {
                state = {
                    ...state,
                    loading: false,
                };
                return state;
            })
            .addCase(listConfigFiles.fulfilled, (state, action) => {
                state = {
                    ...state,
                    datas: action.payload.data,
                    total: action.payload.total,
                    page: action.payload.page,
                    limit: action.payload.limit,
                    loading: false,
                };
                return state;
            })
            .addCase(listConfigFiles.pending, (state) => {
                state = {
                    ...state,
                    loading: true,
                };
                return state;
            })
            .addCase(listConfigFiles.rejected, (state, action) => {
                state = {
                    ...state,
                    loading: false,
                };
                return state;
            })
            .addCase(listConfigFileCryptoAlgos.fulfilled, (state, action) => {
                state = {
                    ...state,
                    cryptoAlgos: action.payload.cryptoAlgos,
                };
                return state;
            })
            .addCase(listOneConfigFile.fulfilled, (state, action) => {
                state = {
                    ...state,
                    viewFile: action.payload.viewFile,
                };
                return state;
            })
            .addCase(listFileSubscribers.fulfilled, (state, action) => {
                state = {
                    ...state,
                    subscribers: action.payload.subscribers || [],
                };
                return state;
            })
    },
})

export const {
    editorConfigFile,
    resetConfigFile,
    cleanFilePage,
} = configFileReducer.actions;
export const selectConfigFile = (state: RootState) => state.configFile;


export default configFileReducer.reducer;
