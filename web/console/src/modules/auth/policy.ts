import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../store';
import { createAuthPolicies, deleteAuthPolicies, modifyAuthPolicies, PolicyResources, Principals } from 'services/auth_policy';


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
        const result = await createAuthPolicies([{ ...state, source: 'pole.io' }]);
        if (!result) {
            return rejectWithValue('创建鉴权策略失败');
        }
        return fulfillWithValue(result);
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const updatePolicyRules = createAsyncThunk(`policy/update`, async ({ state }: { state: PolicyRuleState }, { fulfillWithValue, rejectWithValue }) => {
    try {
        if (!state.id) {
            return rejectWithValue('鉴权策略 ID 不能为空');
        }
        const result = await modifyAuthPolicies([{
            ...state,
            id: state.id,
            source: 'pole.io',
        }]);
        if (!result) {
            return rejectWithValue('更新鉴权策略失败');
        }
        return fulfillWithValue(result);
    } catch (error) {
        return rejectWithValue((error as Error).message); // 捕获错误并返回
    }
});

export const removePolicyRules = createAsyncThunk(`policy/remove`, async ({ ids }: { ids: string[] }, { fulfillWithValue, rejectWithValue }) => {
    try {
        const result = await deleteAuthPolicies(ids.map((id) => ({ id })));
        if (!result) {
            return rejectWithValue('删除鉴权策略失败');
        }
        return fulfillWithValue(result);
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
