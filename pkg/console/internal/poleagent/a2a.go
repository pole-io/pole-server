package poleagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/google/uuid"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/pkg/console/internal/agentworkbench"
)

const (
	A2AProtocolVersion = "0.3.0"
	A2AMessageSend     = "message/send"
)

type TurnRunner interface {
	RunTurn(context.Context, agentworkbench.Actor, TurnRequest) (*TurnResult, error)
}

type TurnRunnerProvider func() TurnRunner

type A2AConfig struct {
	Name        string
	Description string
	URL         string
	Version     string
	Skills      []A2ASkill
}

type A2AAdapter struct {
	card   A2AAgentCard
	runner TurnRunnerProvider
}

type A2AAgentCard = aitypes.A2AAgentCard
type A2ASecurityScheme = aitypes.A2ASecurityScheme
type A2ACapabilities = aitypes.A2ACapabilities
type A2ASkill = aitypes.A2ASkill

type A2ARequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      any              `json:"id"`
	Method  string           `json:"method"`
	Params  A2AMessageParams `json:"params"`
	HasID   bool             `json:"-"`
}

type A2AMessageParams struct {
	Message A2AMessage `json:"message"`
}

type A2AMessage struct {
	Kind      string         `json:"kind"`
	MessageID string         `json:"messageId"`
	Role      string         `json:"role"`
	Parts     []A2APart      `json:"parts"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type A2APart struct {
	Kind string `json:"kind"`
	Text string `json:"text,omitempty"`
}

type A2AResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id"`
	Result  *A2AMessage `json:"result,omitempty"`
	Error   *A2AError   `json:"error,omitempty"`
}

type A2AError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type A2AProtocolError struct {
	Code    int
	Message string
}

func (e *A2AProtocolError) Error() string {
	return e.Message
}

func NewA2AAdapter(config A2AConfig, runner TurnRunnerProvider) (*A2AAdapter, error) {
	config.Name = strings.TrimSpace(config.Name)
	config.URL = strings.TrimSpace(config.URL)
	config.Version = strings.TrimSpace(config.Version)
	if config.Name == "" || config.URL == "" || config.Version == "" {
		return nil, errors.New("A2A name, URL and version are required")
	}
	parsed, err := url.ParseRequestURI(config.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("A2A URL must be an absolute HTTP URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("A2A URL must use HTTP or HTTPS")
	}
	if runner == nil {
		return nil, errors.New("A2A turn runner is required")
	}
	skills := append([]A2ASkill(nil), config.Skills...)
	for i := range skills {
		if len(skills[i].InputModes) == 0 {
			skills[i].InputModes = []string{"text/plain"}
		}
		if len(skills[i].OutputModes) == 0 {
			skills[i].OutputModes = []string{"text/plain"}
		}
	}
	return &A2AAdapter{
		card: A2AAgentCard{
			Name:               config.Name,
			Description:        strings.TrimSpace(config.Description),
			URL:                config.URL,
			Version:            config.Version,
			ProtocolVersion:    A2AProtocolVersion,
			PreferredTransport: "JSONRPC",
			Capabilities:       A2ACapabilities{},
			DefaultInputModes:  []string{"text/plain"},
			DefaultOutputModes: []string{"text/plain"},
			Skills:             skills,
			SecuritySchemes: map[string]A2ASecurityScheme{
				"bearerAuth": {Type: "http", Scheme: "bearer"},
			},
			Security: []map[string][]string{{"bearerAuth": {}}},
		},
		runner: runner,
	}, nil
}

func (a *A2AAdapter) Card() A2AAgentCard {
	card := a.card
	card.DefaultInputModes = append([]string(nil), a.card.DefaultInputModes...)
	card.DefaultOutputModes = append([]string(nil), a.card.DefaultOutputModes...)
	card.Skills = append([]A2ASkill(nil), a.card.Skills...)
	card.SecuritySchemes = make(map[string]A2ASecurityScheme, len(a.card.SecuritySchemes))
	for name, scheme := range a.card.SecuritySchemes {
		card.SecuritySchemes[name] = scheme
	}
	card.Security = append([]map[string][]string(nil), a.card.Security...)
	return card
}

func (a *A2AAdapter) Send(ctx context.Context, actor agentworkbench.Actor,
	request A2ARequest) (*A2AResponse, error) {
	if request.JSONRPC != "2.0" {
		return nil, &A2AProtocolError{Code: -32600, Message: "unsupported JSON-RPC version"}
	}
	if !validA2ARequestID(request.ID) {
		return nil, &A2AProtocolError{Code: -32600, Message: "invalid JSON-RPC request id"}
	}
	if request.Method != A2AMessageSend {
		return nil, &A2AProtocolError{Code: -32601, Message: fmt.Sprintf("unsupported A2A method %q", request.Method)}
	}
	message := textMessage(request.Params.Message)
	if message == "" {
		return nil, &A2AProtocolError{Code: -32602, Message: "A2A message requires a non-empty text part"}
	}
	runner := a.runner()
	if runner == nil {
		return nil, errors.New("Pole Agent runtime is unavailable")
	}
	result, err := runner.RunTurn(ctx, actor, TurnRequest{Message: message})
	if err != nil {
		return nil, err
	}
	responseID := uuid.NewString()
	if request.Params.Message.MessageID != "" {
		responseID = request.Params.Message.MessageID + "-response"
	}
	return &A2AResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: &A2AMessage{
			Kind:      "message",
			MessageID: responseID,
			Role:      "agent",
			Parts:     []A2APart{{Kind: "text", Text: result.Message}},
			Metadata: map[string]any{
				"model":     result.Model,
				"requestId": result.RequestID,
			},
		},
	}, nil
}

func (r *A2ARequest) UnmarshalJSON(data []byte) error {
	type requestAlias struct {
		JSONRPC string           `json:"jsonrpc"`
		ID      json.RawMessage  `json:"id"`
		Method  string           `json:"method"`
		Params  A2AMessageParams `json:"params"`
	}
	var raw requestAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	r.JSONRPC = raw.JSONRPC
	r.Method = raw.Method
	r.Params = raw.Params
	r.HasID = raw.ID != nil
	r.ID = nil
	if !r.HasID {
		return nil
	}
	idDecoder := json.NewDecoder(bytes.NewReader(raw.ID))
	idDecoder.UseNumber()
	if err := idDecoder.Decode(&r.ID); err != nil {
		return err
	}
	if !validA2ARequestID(r.ID) {
		return errors.New("JSON-RPC id must be a string, integer number, or null")
	}
	return nil
}

func validA2ARequestID(id any) bool {
	switch value := id.(type) {
	case nil, string, json.Number:
		if number, ok := value.(json.Number); ok {
			parsed, err := number.Float64()
			return err == nil && !math.IsInf(parsed, 0) && !math.IsNaN(parsed) && math.Trunc(parsed) == parsed
		}
		return true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return math.Trunc(float64(value)) == float64(value)
	case float64:
		return math.Trunc(value) == value
	default:
		return false
	}
}

func A2AErrorResponse(id any, err error) *A2AResponse {
	response := &A2AResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &A2AError{Code: -32603, Message: "Pole Agent execution failed"},
	}
	var protocolError *A2AProtocolError
	if errors.As(err, &protocolError) {
		response.Error.Code = protocolError.Code
		response.Error.Message = protocolError.Message
	}
	return response
}

func textMessage(message A2AMessage) string {
	parts := make([]string, 0, len(message.Parts))
	for _, part := range message.Parts {
		if part.Kind == "text" && strings.TrimSpace(part.Text) != "" {
			parts = append(parts, strings.TrimSpace(part.Text))
		}
	}
	return strings.Join(parts, "\n")
}
