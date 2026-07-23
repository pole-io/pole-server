package poleagent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAIModelSendsToolsAndParsesToolCalls(t *testing.T) {
	var captured openAIChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer gateway-secret", r.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		w.Header().Set("X-Request-Id", "model-request-1")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"model":"test-model",
			"choices":[{
				"finish_reason":"tool_calls",
				"message":{
					"role":"assistant",
					"tool_calls":[{
						"id":"call-1",
						"type":"function",
						"function":{"name":"list_namespaces","arguments":"{\"limit\":10}"}
					}]
				}
			}]
		}`)
	}))
	defer server.Close()

	model, err := NewOpenAIModel(OpenAIModelConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "gateway-secret",
		Model:   "configured-model",
		Timeout: time.Second,
	}, server.Client())
	require.NoError(t, err)
	response, err := model.Complete(context.Background(), ModelRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "system"},
			{Role: RoleUser, Content: "列出命名空间"},
		},
		Tools: []ToolDefinition{{
			Name: "list_namespaces",
			InputSchema: map[string]any{
				"type": "object",
			},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "configured-model", captured.Model)
	require.Len(t, captured.Tools, 1)
	require.Equal(t, "auto", captured.ToolChoice)
	require.Equal(t, "model-request-1", response.RequestID)
	require.Equal(t, "test-model", response.Model)
	require.Len(t, response.Message.ToolCalls, 1)
	require.JSONEq(t, `{"limit":10}`, string(response.Message.ToolCalls[0].Arguments))
}

func TestOpenAIModelDoesNotLeakGatewayErrorBodyOrAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "model-request-2")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"gateway-secret must never be reflected"}`)
	}))
	defer server.Close()

	model, err := NewOpenAIModel(OpenAIModelConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "gateway-secret",
		Model:   "test-model",
	}, server.Client())
	require.NoError(t, err)
	_, err = model.Complete(context.Background(), ModelRequest{
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	var runtimeErr *RuntimeError
	require.ErrorAs(t, err, &runtimeErr)
	require.Equal(t, CategoryModelUnavailable, runtimeErr.Category)
	require.Equal(t, "model-request-2", runtimeErr.RequestID)
	require.NotContains(t, err.Error(), "gateway-secret")
	require.NotContains(t, err.Error(), "must never be reflected")
}

func TestOpenAIModelRejectsNonObjectToolArguments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{
			"choices":[{
				"message":{
					"role":"assistant",
					"tool_calls":[{
						"id":"call-1",
						"type":"function",
						"function":{"name":"list_namespaces","arguments":"[]"}
					}]
				}
			}]
		}`)
	}))
	defer server.Close()

	model, err := NewOpenAIModel(OpenAIModelConfig{
		BaseURL: server.URL,
		APIKey:  "secret",
		Model:   "test-model",
	}, server.Client())
	require.NoError(t, err)
	_, err = model.Complete(context.Background(), ModelRequest{
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), string(CategoryModelUnavailable))
	require.False(t, strings.Contains(err.Error(), "secret"))
}

func TestNewOpenAIModelRequiresServerManagedConfiguration(t *testing.T) {
	_, err := NewOpenAIModel(OpenAIModelConfig{Model: "test-model", APIKey: "secret"}, nil)
	require.Error(t, err)
	_, err = NewOpenAIModel(OpenAIModelConfig{BaseURL: "http://gateway/v1", Model: "test-model"}, nil)
	require.Error(t, err)
	_, err = NewOpenAIModel(OpenAIModelConfig{BaseURL: "http://gateway/v1", APIKey: "secret"}, nil)
	require.Error(t, err)
}
