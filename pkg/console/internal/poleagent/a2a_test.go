package poleagent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/pkg/console/internal/agentworkbench"
)

type a2aTurnRunnerStub struct {
	request TurnRequest
	result  *TurnResult
	err     error
}

func (s *a2aTurnRunnerStub) RunTurn(_ context.Context, _ agentworkbench.Actor,
	req TurnRequest) (*TurnResult, error) {
	s.request = req
	return s.result, s.err
}

func TestA2AAdapterPublishesCallableAgentCard(t *testing.T) {
	adapter, err := NewA2AAdapter(A2AConfig{
		Name:        "Pole Agent",
		Description: "管理 Pole 控制面资源",
		URL:         "https://pole.example.com/ai/agent/a2a/v1",
		Version:     "v1",
		Skills: []A2ASkill{{
			ID:          "pole-control-plane-management",
			Name:        "Pole 控制面管理",
			Description: "发现并管理 Pole 控制面资源",
			Tags:        []string{"pole", "mcp", "control-plane"},
		}},
	}, func() TurnRunner { return &a2aTurnRunnerStub{} })
	require.NoError(t, err)

	card := adapter.Card()
	require.Equal(t, "Pole Agent", card.Name)
	require.Equal(t, "https://pole.example.com/ai/agent/a2a/v1", card.URL)
	require.Equal(t, "0.3.0", card.ProtocolVersion)
	require.Equal(t, "JSONRPC", card.PreferredTransport)
	require.Equal(t, []string{"text/plain"}, card.DefaultInputModes)
	require.Equal(t, []string{"text/plain"}, card.DefaultOutputModes)
	require.False(t, card.Capabilities.Streaming)
	require.Len(t, card.Skills, 1)
	require.Equal(t, "pole-control-plane-management", card.Skills[0].ID)
	require.Equal(t, []string{"text/plain"}, card.Skills[0].InputModes)
	require.Equal(t, []string{"text/plain"}, card.Skills[0].OutputModes)
	require.Equal(t, A2ASecurityScheme{Type: "http", Scheme: "bearer"},
		card.SecuritySchemes["bearerAuth"])
	require.Equal(t, []map[string][]string{{"bearerAuth": {}}}, card.Security)
}

func TestA2ARequestRejectsInvalidJSONRPCIDsAndTracksNotifications(t *testing.T) {
	var request A2ARequest
	require.NoError(t, json.Unmarshal([]byte(
		`{"jsonrpc":"2.0","id":7,"method":"message/send","params":{}}`,
	), &request))
	require.True(t, request.HasID)
	require.Equal(t, json.Number("7"), request.ID)

	require.NoError(t, json.Unmarshal([]byte(
		`{"jsonrpc":"2.0","method":"message/send","params":{}}`,
	), &request))
	require.False(t, request.HasID)

	for _, invalid := range []string{
		`{"jsonrpc":"2.0","id":true,"method":"message/send"}`,
		`{"jsonrpc":"2.0","id":{},"method":"message/send"}`,
		`{"jsonrpc":"2.0","id":[],"method":"message/send"}`,
		`{"jsonrpc":"2.0","id":1.5,"method":"message/send"}`,
	} {
		require.Error(t, json.Unmarshal([]byte(invalid), &request), invalid)
	}
}

func TestA2AAdapterMessageSendUsesPoleAgentRuntime(t *testing.T) {
	runner := &a2aTurnRunnerStub{result: &TurnResult{Message: "当前有 3 个命名空间"}}
	adapter, err := NewA2AAdapter(A2AConfig{
		Name:    "Pole Agent",
		URL:     "https://pole.example.com/ai/agent/a2a/v1",
		Version: "v1",
	}, func() TurnRunner { return runner })
	require.NoError(t, err)

	response, err := adapter.Send(context.Background(), agentworkbench.Actor{
		UserID: "admin",
		Token:  "token",
	}, A2ARequest{
		JSONRPC: "2.0",
		ID:      "request-1",
		Method:  "message/send",
		Params: A2AMessageParams{Message: A2AMessage{
			Kind:      "message",
			MessageID: "message-1",
			Role:      "user",
			Parts:     []A2APart{{Kind: "text", Text: "列出命名空间"}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "列出命名空间", runner.request.Message)
	require.Equal(t, "2.0", response.JSONRPC)
	require.Equal(t, "request-1", response.ID)
	require.Equal(t, "agent", response.Result.Role)
	require.Equal(t, "当前有 3 个命名空间", response.Result.Parts[0].Text)
}

func TestA2AAdapterRejectsUnsupportedMethodAndNonTextMessage(t *testing.T) {
	adapter, err := NewA2AAdapter(A2AConfig{
		Name: "Pole Agent", URL: "https://pole.example.com/ai/agent/a2a/v1", Version: "v1",
	}, func() TurnRunner { return &a2aTurnRunnerStub{} })
	require.NoError(t, err)

	_, err = adapter.Send(context.Background(), agentworkbench.Actor{}, A2ARequest{
		JSONRPC: "2.0", ID: 1, Method: "tasks/get",
	})
	require.ErrorContains(t, err, "unsupported A2A method")
	require.Equal(t, -32601, A2AErrorResponse(1, err).Error.Code)
	require.Equal(t, 1, A2AErrorResponse(1, err).ID)

	_, err = adapter.Send(context.Background(), agentworkbench.Actor{}, A2ARequest{
		JSONRPC: "2.0", ID: 2, Method: "message/send",
		Params: A2AMessageParams{Message: A2AMessage{
			Kind: "message", Role: "user", Parts: []A2APart{{Kind: "file"}},
		}},
	})
	require.ErrorContains(t, err, "text part")
	require.Equal(t, -32602, A2AErrorResponse(2, err).Error.Code)
}
