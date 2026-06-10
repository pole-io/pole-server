import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import {
  A2AAgent,
  A2AAgentSkill,
  createA2AAgents,
  deleteA2AAgents,
  describeA2AAgentCard,
  describeA2AAgents,
  describeA2AAgentSkills,
  DescribeA2AAgentsRequest,
  DescribeA2AAgentSkillsRequest,
  modifyA2AAgents,
} from '../../services/a2a';

export interface A2AState {
  datas: A2AAgent[];
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  editAgent: A2AAgent | null;
  skills: A2AAgentSkill[];
  skillsTotal: number;
  skillsLoading: boolean;
  card: Record<string, any> | null;
  cardLoading: boolean;
}

const initialState: A2AState = {
  datas: [],
  total: 0,
  page: 1,
  limit: 10,
  loading: false,
  editAgent: null,
  skills: [],
  skillsTotal: 0,
  skillsLoading: false,
  card: null,
  cardLoading: false,
};

export const listA2AAgents = createAsyncThunk(
  'a2a/list',
  async ({ param }: { param: DescribeA2AAgentsRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await describeA2AAgents(param);
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

export const saveA2AAgent = createAsyncThunk(
  'a2a/create',
  async ({ param }: { param: A2AAgent }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await createA2AAgents([param]);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const updateA2AAgent = createAsyncThunk(
  'a2a/update',
  async ({ param }: { param: A2AAgent }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await modifyA2AAgents([param]);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const removeA2AAgents = createAsyncThunk(
  'a2a/remove',
  async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await deleteA2AAgents(ids);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const listA2AAgentSkills = createAsyncThunk(
  'a2a/skills',
  async ({ param }: { param: DescribeA2AAgentSkillsRequest }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await describeA2AAgentSkills(param);
      return fulfillWithValue({
        skills: res.list,
        skillsTotal: res.totalCount,
      });
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

export const getA2AAgentCard = createAsyncThunk(
  'a2a/card',
  async ({ id }: { id: string }, { fulfillWithValue, rejectWithValue }) => {
    try {
      const res = await describeA2AAgentCard(id);
      return fulfillWithValue(res);
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  },
);

const a2aReducer = createSlice({
  name: 'ai/a2a',
  initialState,
  reducers: {
    editorA2AAgent: (state, action: PayloadAction<A2AAgent>) => ({
      ...state,
      editAgent: {
        ...action.payload,
        interfaces: action.payload.interfaces ? [...action.payload.interfaces] : [],
        skills: action.payload.skills ? [...action.payload.skills] : [],
      },
    }),
    resetA2AAgent: (state) => ({
      ...state,
      editAgent: null,
    }),
    cleanA2APage: (state) => ({
      ...state,
      datas: [],
      total: 0,
      page: 1,
      limit: 10,
      loading: false,
      editAgent: null,
      skills: [],
      skillsTotal: 0,
      skillsLoading: false,
      card: null,
      cardLoading: false,
    }),
    cleanA2ADetails: (state) => ({
      ...state,
      skills: [],
      skillsTotal: 0,
      skillsLoading: false,
      card: null,
      cardLoading: false,
    }),
  },
  extraReducers: (builder) => {
    builder
      .addCase(listA2AAgents.pending, (state) => {
        state.loading = true;
      })
      .addCase(listA2AAgents.fulfilled, (state, action) => {
        state.datas = action.payload.datas;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
        state.loading = false;
      })
      .addCase(listA2AAgents.rejected, (state) => {
        state.loading = false;
      })
      .addCase(listA2AAgentSkills.pending, (state) => {
        state.skillsLoading = true;
      })
      .addCase(listA2AAgentSkills.fulfilled, (state, action) => {
        state.skills = action.payload.skills;
        state.skillsTotal = action.payload.skillsTotal;
        state.skillsLoading = false;
      })
      .addCase(listA2AAgentSkills.rejected, (state) => {
        state.skillsLoading = false;
      })
      .addCase(getA2AAgentCard.pending, (state) => {
        state.cardLoading = true;
      })
      .addCase(getA2AAgentCard.fulfilled, (state, action) => {
        state.card = action.payload;
        state.cardLoading = false;
      })
      .addCase(getA2AAgentCard.rejected, (state) => {
        state.cardLoading = false;
      });
  },
});

export const {
  editorA2AAgent,
  resetA2AAgent,
  cleanA2APage,
  cleanA2ADetails,
} = a2aReducer.actions;

export const selectA2A = (state: RootState) => state.aiA2A;

export default a2aReducer.reducer;
