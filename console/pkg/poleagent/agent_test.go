package poleagent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
)

type modelFunc func(context.Context, ModelRequest) (ModelResponse, error)

func (f modelFunc) Complete(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	return f(ctx, req)
}

type fakeToolPort struct {
	session *fakeToolSession
	openErr error
}

func (f *fakeToolPort) Open(context.Context, agentworkbench.Actor) (ToolSession, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	return f.session, nil
}

type fakeToolSession struct {
	tools     []ToolDefinition
	listErr   error
	result    ToolResult
	callErr   error
	callCount int
	closed    bool
}

func (f *fakeToolSession) ListTools(context.Context) ([]ToolDefinition, error) {
	return f.tools, f.listErr
}

func (f *fakeToolSession) CallTool(context.Context, string, map[string]any) (ToolResult, error) {
	f.callCount++
	return f.result, f.callErr
}

func (f *fakeToolSession) Close() error {
	f.closed = true
	return nil
}

type fakeApprovalKernel struct {
	calls int
	input agentworkbench.PrepareConfigFileRequest
	err   error
}

func (f *fakeApprovalKernel) PrepareConfigFile(
	_ context.Context,
	_ agentworkbench.Actor,
	input agentworkbench.PrepareConfigFileRequest,
) (*agentworkbench.Proposal, error) {
	f.calls++
	f.input = input
	if f.err != nil {
		return nil, f.err
	}
	return &agentworkbench.Proposal{
		ID:        "proposal-1",
		Status:    agentworkbench.ProposalPreviewReady,
		Resource:  agentworkbench.ResourceRef{Kind: "config.file", Namespace: input.Namespace, Group: input.Group, Name: input.Name},
		ExpiresAt: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}, nil
}

func TestAgentRunsModelThenPreparesPreviewWithoutConfirmOrPublishTool(t *testing.T) {
	var requests []ModelRequest
	model := modelFunc(func(_ context.Context, req ModelRequest) (ModelResponse, error) {
		requests = append(requests, req)
		if len(requests) == 1 {
			return ModelResponse{Message: Message{
				Role: RoleAssistant,
				ToolCalls: []ToolCall{{
					ID:   "call-1",
					Name: PrepareConfigFileUpdateTool,
					Arguments: json.RawMessage(`{
						"namespace":"default",
						"group":"orders",
						"name":"app.yaml",
						"desiredContent":"timeout: 5s\n",
						"comment":"raise timeout"
					}`),
				}},
			}}, nil
		}
		return ModelResponse{
			Message:   Message{Role: RoleAssistant, Content: "临时视图已生成，请确认后再保存草稿。"},
			Model:     "gateway-model",
			RequestID: "model-request",
		}, nil
	})
	session := &fakeToolSession{tools: []ToolDefinition{{
		Name: "list_namespaces",
		InputSchema: map[string]any{
			"type": "object",
		},
	}}}
	kernel := &fakeApprovalKernel{}
	agent := readyAgent(model, &fakeToolPort{session: session}, kernel)

	result, err := agent.RunTurn(context.Background(), testActor(), TurnRequest{
		Message: "把超时改成 5 秒",
		History: []ConversationMessage{
			{Role: RoleUser, Content: "我们处理 orders 配置"},
			{Role: RoleAssistant, Content: "好的"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "临时视图已生成，请确认后再保存草稿。", result.Message)
	require.NotNil(t, result.Proposal)
	require.Equal(t, "proposal-1", result.Proposal.ID)
	require.Equal(t, 1, kernel.calls)
	require.Equal(t, "timeout: 5s\n", kernel.input.DesiredContent)
	require.Len(t, requests, 2)
	require.Len(t, requests[0].Tools, 2)
	require.Empty(t, requests[1].Tools)
	for _, tool := range requests[0].Tools {
		require.NotContains(t, tool.Name, "confirm")
		require.NotContains(t, tool.Name, "publish")
	}
	require.True(t, session.closed)
}

func TestAgentRejectsWriteLikeRemoteMCPTool(t *testing.T) {
	modelCalls := 0
	model := modelFunc(func(context.Context, ModelRequest) (ModelResponse, error) {
		modelCalls++
		return ModelResponse{}, nil
	})
	session := &fakeToolSession{tools: []ToolDefinition{{
		Name: "publish_config_release",
		InputSchema: map[string]any{
			"type": "object",
		},
	}}}
	agent := readyAgent(model, &fakeToolPort{session: session}, &fakeApprovalKernel{})
	_, err := agent.RunTurn(context.Background(), testActor(), TurnRequest{Message: "发布配置"})
	var runtimeErr *RuntimeError
	require.ErrorAs(t, err, &runtimeErr)
	require.Equal(t, CategoryToolRejected, runtimeErr.Category)
	require.Zero(t, modelCalls)
}

func TestAgentBoundsToolLoop(t *testing.T) {
	call := 0
	model := modelFunc(func(context.Context, ModelRequest) (ModelResponse, error) {
		call++
		return ModelResponse{Message: Message{
			Role: RoleAssistant,
			ToolCalls: []ToolCall{{
				ID:        "call",
				Name:      "list_namespaces",
				Arguments: json.RawMessage(`{}`),
			}},
		}}, nil
	})
	session := &fakeToolSession{
		tools: []ToolDefinition{{
			Name: "list_namespaces",
			InputSchema: map[string]any{
				"type": "object",
			},
		}},
		result: ToolResult{Content: `{"namespaces":[]}`},
	}
	agent := New(Options{
		RuntimeMode:   "llm",
		PromptVersion: "v1",
		Model:         "test-model",
		MaxToolRounds: 2,
	}, model, &fakeToolPort{session: session}, &fakeApprovalKernel{})
	_, err := agent.RunTurn(context.Background(), testActor(), TurnRequest{Message: "一直查"})
	var runtimeErr *RuntimeError
	require.ErrorAs(t, err, &runtimeErr)
	require.Equal(t, CategoryToolLoopLimit, runtimeErr.Category)
	require.Equal(t, 2, call)
	require.Equal(t, 2, session.callCount)
}

func TestAgentMarksMCPBusinessErrorTraceFailed(t *testing.T) {
	step := 0
	model := modelFunc(func(context.Context, ModelRequest) (ModelResponse, error) {
		step++
		if step == 1 {
			return ModelResponse{Message: Message{
				Role: RoleAssistant,
				ToolCalls: []ToolCall{{
					ID:        "call-1",
					Name:      "list_namespaces",
					Arguments: json.RawMessage(`{}`),
				}},
			}}, nil
		}
		return ModelResponse{Message: Message{Role: RoleAssistant, Content: "查询失败，请检查权限。"}}, nil
	})
	session := &fakeToolSession{
		tools: []ToolDefinition{{
			Name: "list_namespaces",
			InputSchema: map[string]any{
				"type": "object",
			},
		}},
		result: ToolResult{Content: "permission denied", IsError: true},
	}
	agent := readyAgent(model, &fakeToolPort{session: session}, &fakeApprovalKernel{})
	result, err := agent.RunTurn(context.Background(), testActor(), TurnRequest{Message: "查命名空间"})
	require.NoError(t, err)
	require.Len(t, result.Tools, 1)
	require.Equal(t, ToolTraceFailed, result.Tools[0].Status)
	require.True(t, result.Tools[0].IsError)
}

func TestAgentRejectsUnknownProposalFields(t *testing.T) {
	model := modelFunc(func(context.Context, ModelRequest) (ModelResponse, error) {
		return ModelResponse{Message: Message{
			Role: RoleAssistant,
			ToolCalls: []ToolCall{{
				ID:   "call-1",
				Name: PrepareConfigFileUpdateTool,
				Arguments: json.RawMessage(`{
					"namespace":"default",
					"group":"orders",
					"name":"app.yaml",
					"desiredContent":"a: 2",
					"publish":true
				}`),
			}},
		}}, nil
	})
	kernel := &fakeApprovalKernel{}
	agent := readyAgent(model, &fakeToolPort{session: &fakeToolSession{}}, kernel)
	_, err := agent.RunTurn(context.Background(), testActor(), TurnRequest{Message: "修改并发布"})
	var runtimeErr *RuntimeError
	require.ErrorAs(t, err, &runtimeErr)
	require.Equal(t, CategoryValidationFailed, runtimeErr.Category)
	require.Zero(t, kernel.calls)
}

func TestAgentRuntimeProbesAuthenticatedMCPAndReturnsNoEndpointOrSecret(t *testing.T) {
	session := &fakeToolSession{tools: []ToolDefinition{{
		Name: "list_namespaces",
		InputSchema: map[string]any{
			"type": "object",
		},
	}}}
	agent := readyAgent(modelFunc(func(context.Context, ModelRequest) (ModelResponse, error) {
		return ModelResponse{}, errors.New("not used")
	}), &fakeToolPort{session: session}, &fakeApprovalKernel{})
	runtime := agent.Runtime(context.Background(), testActor())
	require.True(t, runtime.Ready)
	require.True(t, runtime.Configured)
	require.True(t, runtime.MCPConnected)
	require.Equal(t, []string{"list_namespaces"}, runtime.MCPTools)
	require.Equal(t, []string{"list_namespaces", PrepareConfigFileUpdateTool}, runtime.Tools)
	encoded, err := json.Marshal(runtime)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "apiKey")
	require.NotContains(t, string(encoded), "baseURL")
	require.True(t, session.closed)
}

func TestAgentRuntimeReportsSafeProbeFailure(t *testing.T) {
	agent := readyAgent(modelFunc(func(context.Context, ModelRequest) (ModelResponse, error) {
		return ModelResponse{}, nil
	}), &fakeToolPort{openErr: errors.New("http://secret-host?token=secret")}, &fakeApprovalKernel{})
	runtime := agent.Runtime(context.Background(), testActor())
	require.False(t, runtime.Ready)
	require.True(t, runtime.Configured)
	require.False(t, runtime.MCPConnected)
	require.Empty(t, runtime.MCPTools)
	require.Empty(t, runtime.Tools)
	encoded, err := json.Marshal(runtime)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"mcpTools":[]`)
	require.Contains(t, string(encoded), `"tools":[]`)
	require.Equal(t, "Pole MCP connection failed", runtime.Reason)
	require.NotContains(t, runtime.Reason, "secret")
}

func TestAgentStatusExplainsIncompleteRuntime(t *testing.T) {
	agent := New(Options{RuntimeMode: "deterministic"}, nil, nil, nil)
	status := agent.Status()
	require.False(t, status.Ready)
	require.Contains(t, status.Missing, "agent.runtimeMode=llm")
	require.Contains(t, status.Missing, "agent.model")
	require.Contains(t, status.Missing, "agent.mcp")
	require.Contains(t, status.Missing, "agent.approvalKernel")
}

func readyAgent(model ModelPort, tools ToolPort, kernel ChangeApprovalKernel) *Agent {
	return New(Options{
		ID:            "pole-control-plane",
		RuntimeMode:   "llm",
		PromptVersion: "v1",
		Model:         "test-model",
	}, model, tools, kernel)
}

func testActor() agentworkbench.Actor {
	return agentworkbench.Actor{UserID: "admin", Token: "pole-token", RequestID: "console-request"}
}
