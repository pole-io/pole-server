import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { createAuthPolicies, PolicyResources, Principals } from 'services/auth_policy';


// State 和 Action 类型定义
export interface PolicyRuleState {
    // 策略唯一ID
    id?: string
    // 策略名称
    name: string
    // 资源操作权限
    action: string
    // 涉及的用户 or 用户组
    principals?: Principals
    // 简单描述
    comment?: string
    // 策略关联的资源
    resources?: PolicyResources
    // 是否默认策略
    default_strategy?: boolean
    // 服务端接口
    functions?: string[]
    // 策略生效的资源标签
    resource_labels?: string[]
    // 策略资源标签
    metadata?: Record<string, string>
}

const initialState: PolicyRuleState = {
    id: '',
    name: '',
    principals: {
        users: [],
        groups: []
    },
    action: '',
    comment: '',
    resources: {},
    default_strategy: false,
    functions: [],
    resource_labels: [],
    metadata: {}
};

export const savePolicyRules = createAsyncThunk(`policy/create`, async ({ state }: { state: PolicyRuleState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        // const ret = await createAuthPolicies([{ ...state, source: 'pole.io' }]);
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updatePolicyRules = createAsyncThunk(`policy/update`, async ({ state }: { state: PolicyRuleState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removePolicyRules = createAsyncThunk(`policy/remove`, async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        return fulfillWithValue("ok"); // 返回 token
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});


const policyRuleReducer = createSlice({
    name: 'policy/edit',
    initialState,
    reducers: {
        editorPolicyRules: (state, action: PayloadAction<PolicyRuleState>) => {
            state = {
                ...state,
                ...action.payload
            };
            return state;
        },
        resetPolicyRules: (state) => {
            state = {
                ...initialState
            };
            return state;
        }
    },
    extraReducers: () => { },
})

export const {
    editorPolicyRules,
    resetPolicyRules,
} = policyRuleReducer.actions;
export const selectPolicyRule = (state: RootState) => state.authPolicyRules;

export default policyRuleReducer.reducer;
