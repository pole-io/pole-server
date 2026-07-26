package poleagent

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
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

type A2AAgentCard struct {
	Name               string          `json:"name"`
	Description        string          `json:"description,omitempty"`
	URL                string          `json:"url"`
	Version            string          `json:"version"`
	ProtocolVersion    string          `json:"protocolVersion"`
	Capabilities       A2ACapabilities `json:"capabilities"`
	DefaultInputModes  []string        `json:"defaultInputModes"`
	DefaultOutputModes []string        `json:"defaultOutputModes"`
	Skills             []A2ASkill      `json:"skills"`
}

type A2ACapabilities struct {
	Streaming         bool `json:"streaming"`
	PushNotifications bool `json:"pushNotifications"`
}

type A2ASkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Examples    []string `json:"examples,omitempty"`
	InputModes  []string `json:"inputModes,omitempty"`
	OutputModes []string `json:"outputModes,omitempty"`
}

type A2ARequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      any              `json:"id"`
	Method  string           `json:"method"`
	Params  A2AMessageParams `json:"params"`
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
			skills[i].InputModes = []string{"text"}
		}
		if len(skills[i].OutputModes) == 0 {
			skills[i].OutputModes = []string{"text"}
		}
	}
	return &A2AAdapter{
		card: A2AAgentCard{
			Name:               config.Name,
			Description:        strings.TrimSpace(config.Description),
			URL:                config.URL,
			Version:            config.Version,
			ProtocolVersion:    A2AProtocolVersion,
			Capabilities:       A2ACapabilities{},
			DefaultInputModes:  []string{"text"},
			DefaultOutputModes: []string{"text"},
			Skills:             skills,
		},
		runner: runner,
	}, nil
}

func (a *A2AAdapter) Card() A2AAgentCard {
	card := a.card
	card.DefaultInputModes = append([]string(nil), a.card.DefaultInputModes...)
	card.DefaultOutputModes = append([]string(nil), a.card.DefaultOutputModes...)
	card.Skills = append([]A2ASkill(nil), a.card.Skills...)
	return card
}

func (a *A2AAdapter) Send(ctx context.Context, actor agentworkbench.Actor,
	request A2ARequest) (*A2AResponse, error) {
	if request.JSONRPC != "2.0" {
		return nil, errors.New("unsupported JSON-RPC version")
	}
	if request.Method != A2AMessageSend {
		return nil, fmt.Errorf("unsupported A2A method %q", request.Method)
	}
	message := textMessage(request.Params.Message)
	if message == "" {
		return nil, errors.New("A2A message requires a non-empty text part")
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

func textMessage(message A2AMessage) string {
	parts := make([]string, 0, len(message.Parts))
	for _, part := range message.Parts {
		if part.Kind == "text" && strings.TrimSpace(part.Text) != "" {
			parts = append(parts, strings.TrimSpace(part.Text))
		}
	}
	return strings.Join(parts, "\n")
}
