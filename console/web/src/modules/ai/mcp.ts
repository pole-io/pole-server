import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import {
  createMCPServers,
  deleteMCPServers,
  describeMCPServers,
  describeMCPServerTools,
  DescribeMCPServersRequest,
  DescribeMCPServerToolsRequest,
  MCPServer,
  MCPServerTool,
  modifyMCPServers,
} from '../../services/mcp';

export interface MCPState {
  datas: MCPServer[];
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  editServer: MCPServer | null;
  tools: MCPServerTool[];
  toolsTotal: number;
  toolsLoading: boolean;
}

const initialState: MCPState = {
  datas: [],
  total: 0,
  page: 1,
  limit: 10,
  loading: false,
  editServer: null,
  tools: [],
  toolsTotal: 0,
  toolsLoading: false,
};

export const listMCPServers = createAsyncThunk(
  'mcp/list',
  async ({ param }: { param: DescribeMCPServersRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await describeMCPServers(param);
      return fulfillWithValue({
        datas: res.list,
        total: res.totalCount,
        page: Math.floor(param.offset / param.limit) + 1,
        limit: param.limit || 10,
      });
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const saveMCPServer = createAsyncThunk(
  'mcp/create',
  async ({ param }: { param: MCPServer }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await createMCPServers([param]);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const updateMCPServer = createAsyncThunk(
  'mcp/update',
  async ({ param }: { param: MCPServer }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await modifyMCPServers([param]);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const removeMCPServers = createAsyncThunk(
  'mcp/remove',
  async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await deleteMCPServers(ids);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const listMCPServerTools = createAsyncThunk(
  'mcp/tools',
  async ({ param }: { param: DescribeMCPServerToolsRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await describeMCPServerTools(param);
      return fulfillWithValue({
        tools: res.list,
        toolsTotal: res.totalCount,
      });
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

const mcpReducer = createSlice({
  name: 'ai/mcp',
  initialState,
  reducers: {
    editorMCPServer: (state, action: PayloadAction<MCPServer>) => ({
      ...state,
      editServer: {
        ...action.payload,
      },
    }),
    resetMCPServer: (state) => ({
      ...state,
      editServer: null,
    }),
    cleanMCPPage: (state) => ({
      ...state,
      datas: [],
      total: 0,
      page: 1,
      limit: 10,
      loading: false,
      editServer: null,
      tools: [],
      toolsTotal: 0,
      toolsLoading: false,
    }),
    cleanMCPTools: (state) => ({
      ...state,
      tools: [],
      toolsTotal: 0,
      toolsLoading: false,
    }),
  },
  extraReducers: (builder) => {
    builder
      .addCase(listMCPServers.pending, (state) => {
        state.loading = true;
      })
      .addCase(listMCPServers.fulfilled, (state, action) => {
        state.datas = action.payload.datas;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
        state.loading = false;
      })
      .addCase(listMCPServers.rejected, (state) => {
        state.loading = false;
      })
      .addCase(listMCPServerTools.pending, (state) => {
        state.toolsLoading = true;
      })
      .addCase(listMCPServerTools.fulfilled, (state, action) => {
        state.tools = action.payload.tools;
        state.toolsTotal = action.payload.toolsTotal;
        state.toolsLoading = false;
      })
      .addCase(listMCPServerTools.rejected, (state) => {
        state.toolsLoading = false;
      });
  },
});

export const {
  editorMCPServer,
  resetMCPServer,
  cleanMCPPage,
  cleanMCPTools,
} = mcpReducer.actions;

export const selectMCP = (state: RootState) => state.aiMCP;

export default mcpReducer.reducer;
