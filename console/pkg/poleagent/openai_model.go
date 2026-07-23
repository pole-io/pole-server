package poleagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxModelResponseBytes = 2 << 20

type OpenAIModelConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type OpenAIModel struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

func NewOpenAIModel(cfg OpenAIModelConfig, client *http.Client) (*OpenAIModel, error) {
	endpoint, err := chatCompletionsEndpoint(cfg.BaseURL)
	if err != nil {
		return nil, runtimeError(CategoryRuntimeUnavailable, 503101,
			"invalid LLM Gateway address", false, err)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, runtimeError(CategoryRuntimeUnavailable, 503102,
			"LLM Gateway API Key is not configured", false, nil)
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, runtimeError(CategoryRuntimeUnavailable, 503103,
			"LLM model is not configured", false, nil)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &OpenAIModel{
		endpoint: endpoint,
		apiKey:   cfg.APIKey,
		model:    strings.TrimSpace(cfg.Model),
		client:   client,
	}, nil
}

func (m *OpenAIModel) Complete(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(req.Model) == "" {
		req.Model = m.model
	}
	payload := openAIChatRequest{
		Model:      req.Model,
		Messages:   make([]openAIMessage, 0, len(req.Messages)),
		Tools:      make([]openAITool, 0, len(req.Tools)),
		ToolChoice: "auto",
	}
	for _, message := range req.Messages {
		wire, err := toOpenAIMessage(message)
		if err != nil {
			return ModelResponse{}, runtimeError(CategoryValidationFailed, 422101,
				"invalid model message", false, err)
		}
		payload.Messages = append(payload.Messages, wire)
	}
	for _, tool := range req.Tools {
		payload.Tools = append(payload.Tools, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}
	if len(payload.Tools) == 0 {
		payload.ToolChoice = ""
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ModelResponse{}, runtimeError(CategoryValidationFailed, 422102,
			"encode model request", false, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoint, bytes.NewReader(body))
	if err != nil {
		return ModelResponse{}, runtimeError(CategoryModelUnavailable, 503111,
			"create LLM Gateway request", false, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(httpReq)
	if err != nil {
		retryable := errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
		return ModelResponse{}, runtimeError(CategoryModelUnavailable, 503112,
			"LLM Gateway request failed", retryable, err)
	}
	defer resp.Body.Close()
	requestID := modelRequestID(resp.Header)
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxModelResponseBytes+1))
	if err != nil {
		modelErr := runtimeError(CategoryModelUnavailable, 503113,
			"read LLM Gateway response", true, err)
		modelErr.RequestID = requestID
		return ModelResponse{}, modelErr
	}
	if len(responseBody) > maxModelResponseBytes {
		modelErr := runtimeError(CategoryModelUnavailable, 503114,
			"LLM Gateway response is too large", false, nil)
		modelErr.RequestID = requestID
		return ModelResponse{}, modelErr
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError
		modelErr := runtimeError(CategoryModelUnavailable, 503115,
			fmt.Sprintf("LLM Gateway returned HTTP %d", resp.StatusCode), retryable, nil)
		modelErr.RequestID = requestID
		return ModelResponse{}, modelErr
	}

	var decoded openAIChatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		modelErr := runtimeError(CategoryModelUnavailable, 503116,
			"decode LLM Gateway response", false, err)
		modelErr.RequestID = requestID
		return ModelResponse{}, modelErr
	}
	if len(decoded.Choices) == 0 {
		modelErr := runtimeError(CategoryModelUnavailable, 503117,
			"LLM Gateway response has no choices", false, nil)
		modelErr.RequestID = requestID
		return ModelResponse{}, modelErr
	}
	message, err := fromOpenAIMessage(decoded.Choices[0].Message)
	if err != nil {
		modelErr := runtimeError(CategoryModelUnavailable, 503118,
			"decode LLM tool call", false, err)
		modelErr.RequestID = requestID
		return ModelResponse{}, modelErr
	}
	return ModelResponse{
		Message:      message,
		Model:        decoded.Model,
		FinishReason: decoded.Choices[0].FinishReason,
		RequestID:    requestID,
	}, nil
}

func chatCompletionsEndpoint(baseURL string) (string, error) {
	value := strings.TrimSpace(baseURL)
	if value == "" {
		return "", errors.New("empty base URL")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("base URL must be an absolute HTTP URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("base URL scheme must be http or https")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	if !strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), "/chat/completions") {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/chat/completions"
	}
	return parsed.String(), nil
}

func modelRequestID(header http.Header) string {
	for _, key := range []string{"X-Request-Id", "OpenAI-Request-Id", "Request-Id"} {
		if value := strings.TrimSpace(header.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

type openAIChatRequest struct {
	Model      string          `json:"model"`
	Messages   []openAIMessage `json:"messages"`
	Tools      []openAITool    `json:"tools,omitempty"`
	ToolChoice string          `json:"tool_choice,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type openAIToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function openAIFunctionCallBody `json:"function"`
}

type openAIFunctionCallBody struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
}

func toOpenAIMessage(message Message) (openAIMessage, error) {
	wire := openAIMessage{
		Role:       string(message.Role),
		Content:    message.Content,
		ToolCallID: message.ToolCallID,
	}
	for _, call := range message.ToolCalls {
		if strings.TrimSpace(call.ID) == "" || strings.TrimSpace(call.Name) == "" || len(call.Arguments) == 0 {
			return openAIMessage{}, errors.New("tool call requires id, name and arguments")
		}
		wire.ToolCalls = append(wire.ToolCalls, openAIToolCall{
			ID:   call.ID,
			Type: "function",
			Function: openAIFunctionCallBody{
				Name:      call.Name,
				Arguments: string(call.Arguments),
			},
		})
	}
	return wire, nil
}

func fromOpenAIMessage(wire openAIMessage) (Message, error) {
	message := Message{Role: MessageRole(wire.Role), Content: wire.Content}
	if message.Role == "" {
		message.Role = RoleAssistant
	}
	for _, call := range wire.ToolCalls {
		if call.Type != "" && call.Type != "function" {
			return Message{}, fmt.Errorf("unsupported tool call type %q", call.Type)
		}
		if strings.TrimSpace(call.ID) == "" || strings.TrimSpace(call.Function.Name) == "" {
			return Message{}, errors.New("tool call requires id and function name")
		}
		arguments := json.RawMessage(call.Function.Arguments)
		if len(arguments) == 0 || !json.Valid(arguments) {
			return Message{}, fmt.Errorf("tool %s returned invalid JSON arguments", call.Function.Name)
		}
		var object map[string]any
		if err := json.Unmarshal(arguments, &object); err != nil {
			return Message{}, fmt.Errorf("tool %s arguments must be an object: %w", call.Function.Name, err)
		}
		message.ToolCalls = append(message.ToolCalls, ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: append(json.RawMessage(nil), arguments...),
		})
	}
	return message, nil
}
