import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import {
    ConfigFileRelease,
    ConfigFileReleaseView,
    DeleteFileReleaseRequest,
    deleteFileReleases,
    describeFileReleaseVersions,
    DescribeFileReleaseVersionsRequest,
    describeOneFileRelease,
    DescribeOneFileReleaseRequest,
    releaseConfigFile,
    ReleaseConfigFilePequest,
    ReleaseVersion,
    promoteGrayFileReleaseToDraft,
    PromoteGrayFileReleaseRequest,
    rollbackFileReleases,
    RollbackFileReleasesResquest,
    stopGrayFileReleases,
    StopGrayFileReleaseRequest
} from 'services/config_release';

// State 和 Action 类型定义
export interface FileReleaseState {
    versions: ReleaseVersion[]
    total: number
    page: number
    limit: number
    loading: boolean

    editFileRelease: ConfigFileRelease | null
    viewFileRelease: ConfigFileReleaseView | null
    activeFileRelease?: ConfigFileReleaseView | null
}

const initialState: FileReleaseState = {
    versions: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,

    editFileRelease: null,
    viewFileRelease: null,
    activeFileRelease: null
};

export const listActiveConfigFileRelease = createAsyncThunk(`config_release/list_active`, async ({ param }: { param: DescribeOneFileReleaseRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeOneFileRelease(param);
        return fulfillWithValue({
            activeFileRelease: res.configFileRelease || null,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listOneConfigFileRelease = createAsyncThunk(`config_release/list_one`, async ({ param }: { param: DescribeOneFileReleaseRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeOneFileRelease(param);
        return fulfillWithValue({
            viewFileRelease: res.configFileRelease || null,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const listConfigFileReleases = createAsyncThunk(`config_release/list`, async ({ param }: { param: DescribeFileReleaseVersionsRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await describeFileReleaseVersions(param);
        return fulfillWithValue({
            versions: res.configFileReleases || [],
            total: res.configFileReleases?.length || 0,
        }); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const publishConfigFiles = createAsyncThunk(`config_release/create`, async ({ param }: { param: ReleaseConfigFilePequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await releaseConfigFile(param);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const releaseRollback = createAsyncThunk(`config_release/rollback`, async ({ param }: { param: RollbackFileReleasesResquest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await rollbackFileReleases([param]);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const releasesRemove = createAsyncThunk(`config_release/delete`, async ({ param }: { param: DeleteFileReleaseRequest[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const res = await deleteFileReleases(param);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const stopGrayRelease = createAsyncThunk(`config_release/stop_gray`, async ({ param }: { param: StopGrayFileReleaseRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        await stopGrayFileReleases([param]);
        return fulfillWithValue("ok");
    } catch (error) {
        return rejectWithValue((error as Error).message);
    }
});

export const promoteGrayReleaseToDraft = createAsyncThunk(`config_release/promote_gray`, async ({ param }: { param: PromoteGrayFileReleaseRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
        await promoteGrayFileReleaseToDraft(param);
        return fulfillWithValue("ok");
    } catch (error) {
        return rejectWithValue((error as Error).message);
    }
});

const fileReleaseReducer = createSlice({
    name: 'service/edit',
    initialState,
    reducers: {
        editorFileRelease: (state, action: PayloadAction<ConfigFileRelease>) => {
            state = {
                ...state,
                editFileRelease: {
                    ...action.payload
                }
            };
            return state;
        },
        viewFileRelease: (state, action) => {
            state = {
                ...state,
                viewFileRelease: {
                    ...action.payload
                }
            };
            return state;
        },
        resetFileRelease: (state) => {
            state = {
                ...state,
                editFileRelease: null,
                viewFileRelease: null,
            };
            return state;
        },
        cleanFileReleasePage: (state) => {
            state = {
                ...state,
                versions: [],
                total: 0,
                page: 1,
                limit: 10,
            };
            return state;
        },
    },
    extraReducers: (builder) => {
        builder
            .addCase(listConfigFileReleases.fulfilled, (state, action) => {
                state.versions = action.payload.versions;
                state.total = action.payload.total;
                state.loading = false;
            })
            .addCase(listConfigFileReleases.pending, (state) => {
                state.loading = true;
            })
            .addCase(listConfigFileReleases.rejected, (state) => {
                state.loading = false;
            })
            .addCase(listOneConfigFileRelease.fulfilled, (state, action) => {
                state.viewFileRelease = action.payload.viewFileRelease;
            })
            .addCase(listActiveConfigFileRelease.fulfilled, (state, action) => {
                state.activeFileRelease = action.payload.activeFileRelease;
            })
    },
})

export const {
    editorFileRelease,
    resetFileRelease,
    viewFileRelease,
    cleanFileReleasePage,
} = fileReleaseReducer.actions;
export const selectFileRelease = (state: RootState) => state.configFileRelease;

export default fileReleaseReducer.reducer;
