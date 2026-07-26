package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
	"github.com/pole-io/pole-server/console/pkg/poleagent"
	"github.com/pole-io/pole-server/console/pkg/systemsettings"
)

type AgentHandler struct {
	config    *bootstrap.Config
	workbench *agentworkbench.Workbench
	runtime   *systemsettings.Manager
	a2a       *poleagent.A2AAdapter
	a2aErr    error
}

func NewAgentHandler(config *bootstrap.Config, workbench *agentworkbench.Workbench,
	runtime *systemsettings.Manager) *AgentHandler {
	agentConfig := config.Agent.Normalize()
	endpoint := strings.TrimSpace(agentConfig.A2A.Endpoint)
	if endpoint == "" {
		port := config.WebServer.ListenPort
		if port <= 0 {
			port = 8080
		}
		endpoint = fmt.Sprintf("http://127.0.0.1:%d/ai/agent/a2a/v1", port)
	}
	adapter, err := poleagent.NewA2AAdapter(poleagent.A2AConfig{
		Name:        agentConfig.A2A.Name,
		Description: agentConfig.A2A.Description,
		URL:         endpoint,
		Version:     agentConfig.Definition.SystemPrompt.BuiltinVersion,
		Skills: []poleagent.A2ASkill{{
			ID:          "pole-control-plane-management",
			Name:        "Pole 控制面管理",
			Description: "通过 Pole MCP 发现、查询并安全管理控制面资源",
			Tags:        []string{"pole", "mcp", "control-plane"},
		}},
	}, func() poleagent.TurnRunner {
		snapshot := runtime.Current()
		if snapshot == nil {
			return nil
		}
		return snapshot.Agent
	})
	if err != nil {
		return &AgentHandler{
			config: config, workbench: workbench, runtime: runtime,
			a2aErr: fmt.Errorf("initialize Pole Agent A2A adapter: %w", err),
		}
	}
	return &AgentHandler{config: config, workbench: workbench, runtime: runtime, a2a: adapter}
}

func (h *AgentHandler) A2ACard(c *gin.Context) {
	if h.a2aErr != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": h.a2aErr.Error()})
		return
	}
	card := h.a2a.Card()
	if snapshot := h.runtime.Current(); snapshot != nil {
		if name := strings.TrimSpace(snapshot.Profile.AgentID); name != "" {
			card.Name = name
		}
		if version := strings.TrimSpace(snapshot.Profile.PromptVersion); version != "" {
			card.Version = version
		}
	}
	c.JSON(http.StatusOK, card)
}

func (h *AgentHandler) A2ASend(c *gin.Context) {
	if h.a2aErr != nil {
		c.JSON(http.StatusServiceUnavailable, poleagent.A2AErrorResponse(nil, h.a2aErr))
		return
	}
	actor, ok := h.a2aActor(c)
	if !ok {
		return
	}
	var request poleagent.A2ARequest
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		code := -32600
		message := "invalid A2A JSON-RPC request"
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			code = -32700
			message = "invalid JSON payload"
		}
		c.JSON(http.StatusBadRequest, poleagent.A2AErrorResponse(nil,
			&poleagent.A2AProtocolError{Code: code, Message: message}))
		return
	}
	if err := ensureJSONBodyConsumed(decoder); err != nil {
		c.JSON(http.StatusBadRequest, poleagent.A2AErrorResponse(nil,
			&poleagent.A2AProtocolError{Code: -32700, Message: "invalid JSON payload"}))
		return
	}
	response, err := h.a2a.Send(c.Request.Context(), actor, request)
	if err != nil {
		if !request.HasID {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusOK, poleagent.A2AErrorResponse(request.ID, err))
		return
	}
	if !request.HasID {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, response)
}

func ensureJSONBodyConsumed(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func (h *AgentHandler) Runtime(c *gin.Context) {
	actor, ok := h.actor(c)
	if !ok {
		return
	}
	snapshot := h.runtime.Current()
	runtime := snapshot.Agent.Runtime(c.Request.Context(), actor)
	runtime.EffectiveRevision = snapshot.Revision
	c.JSON(http.StatusOK, gin.H{"code": 200000, "info": "execute success", "data": runtime})
}

func (h *AgentHandler) RunTurn(c *gin.Context) {
	actor, ok := h.actor(c)
	if !ok {
		return
	}
	var req struct {
		SessionID string `json:"sessionId"`
		poleagent.TurnRequest
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
		writePoleAgentError(c, &poleagent.RuntimeError{
			Category: poleagent.CategoryInvalidRequest,
			Code:     400001,
			Info:     "sessionId and message are required",
		})
		return
	}
	snapshot := h.runtime.Current()
	result, err := snapshot.Agent.RunTurn(c.Request.Context(), actor, req.TurnRequest)
	if err != nil {
		writePoleAgentError(c, err)
		return
	}
	result.Runtime.EffectiveRevision = snapshot.Revision
	c.JSON(http.StatusOK, gin.H{"code": 200000, "info": "execute success", "data": result})
}

func (h *AgentHandler) PrepareConfigFile(c *gin.Context) {
	actor, ok := h.actor(c)
	if !ok {
		return
	}
	var req agentworkbench.PrepareConfigFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeAgentError(c, agentworkbench.ErrInvalidRequest)
		return
	}
	proposal, err := h.workbench.PrepareConfigFile(c.Request.Context(), actor, req)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200000, "info": "execute success", "data": proposal})
}

func writePoleAgentError(c *gin.Context, err error) {
	status := http.StatusBadGateway
	response := &poleagent.RuntimeError{
		Category:  poleagent.CategoryDownstreamFailed,
		Code:      502001,
		Info:      "Pole Agent execution failed",
		Retryable: false,
	}
	var runtimeErr *poleagent.RuntimeError
	if errors.As(err, &runtimeErr) {
		response = runtimeErr
		switch runtimeErr.Category {
		case poleagent.CategoryInvalidRequest:
			status = http.StatusBadRequest
		case poleagent.CategoryToolRejected:
			status = http.StatusForbidden
		case poleagent.CategoryValidationFailed, poleagent.CategoryNotDraftable, poleagent.CategoryToolLoopLimit:
			status = http.StatusUnprocessableEntity
		case poleagent.CategoryRuntimeUnavailable, poleagent.CategoryModelUnavailable, poleagent.CategoryToolUnavailable:
			status = http.StatusServiceUnavailable
		default:
			status = http.StatusBadGateway
		}
	}
	c.JSON(status, gin.H{
		"code":      response.Code,
		"info":      response.Info,
		"category":  response.Category,
		"requestId": response.RequestID,
		"retryable": response.Retryable,
	})
}

func (h *AgentHandler) ConfirmProposal(c *gin.Context) {
	actor, ok := h.actor(c)
	if !ok {
		return
	}
	var req agentworkbench.ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeAgentError(c, agentworkbench.ErrInvalidRequest)
		return
	}
	req.ProposalID = c.Param("proposal_id")
	receipt, err := h.workbench.Confirm(c.Request.Context(), actor, req)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200000, "info": "execute success", "data": receipt})
}

func (h *AgentHandler) actor(c *gin.Context) (agentworkbench.Actor, bool) {
	userID, token, ok := verifyAccessPermission(c, h.config)
	if !ok {
		return agentworkbench.Actor{}, false
	}
	// Browser requests use the signed JWT cookie, while API clients may already
	// carry the Pole user and token headers validated above.
	if userID == "" {
		userID = c.GetHeader("X-Pole-User")
		token = c.GetHeader("Authorization")
	}
	return agentworkbench.Actor{UserID: userID, Token: token, RequestID: c.GetHeader("X-Request-Id")}, true
}

func (h *AgentHandler) a2aActor(c *gin.Context) (agentworkbench.Actor, bool) {
	userID, token, ok := verifyAccessPermissionWithStatus(c, h.config, http.StatusUnauthorized)
	if !ok {
		return agentworkbench.Actor{}, false
	}
	if userID == "" {
		userID = c.GetHeader("X-Pole-User")
		token = c.GetHeader("Authorization")
	}
	return agentworkbench.Actor{UserID: userID, Token: token, RequestID: c.GetHeader("X-Request-Id")}, true
}

func writeAgentError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	code := uint32(400001)
	category := "INVALID_REQUEST"
	info := err.Error()
	requestID := c.GetHeader("X-Request-Id")

	var upstream *agentworkbench.UpstreamError
	if errors.As(err, &upstream) {
		status = upstream.Status
		if status < http.StatusBadRequest {
			status = http.StatusBadGateway
		}
		code = upstream.Code
		category = "DOWNSTREAM_FAILED"
		info = upstream.Info
		if upstream.RequestID != "" {
			requestID = upstream.RequestID
		}
	} else {
		switch {
		case errors.Is(err, agentworkbench.ErrStalePreview):
			status, code, category = http.StatusConflict, 409001, "STALE_PREVIEW"
		case errors.Is(err, agentworkbench.ErrProposalExpired):
			status, code, category = http.StatusGone, 410001, "PROPOSAL_EXPIRED"
		case errors.Is(err, agentworkbench.ErrProposalNotFound):
			status, code, category = http.StatusNotFound, 404001, "PROPOSAL_NOT_FOUND"
		case errors.Is(err, agentworkbench.ErrProposalForbidden):
			status, code, category = http.StatusForbidden, 403001, "PERMISSION_DENIED"
		case errors.Is(err, agentworkbench.ErrConfirmationMismatch):
			status, code, category = http.StatusConflict, 409002, "CONFIRMATION_MISMATCH"
		case errors.Is(err, agentworkbench.ErrProposalApplying):
			status, code, category = http.StatusConflict, 409003, "PROPOSAL_APPLYING"
		case errors.Is(err, agentworkbench.ErrSensitiveResource):
			status, code, category = http.StatusUnprocessableEntity, 422001, "NOT_DRAFTABLE"
		case errors.Is(err, agentworkbench.ErrNoChanges):
			status, code, category = http.StatusUnprocessableEntity, 422002, "NO_CHANGES"
		case errors.Is(err, agentworkbench.ErrProposalCapacity):
			status, code, category = http.StatusServiceUnavailable, 503001, "WORKBENCH_CAPACITY"
		}
	}
	c.JSON(status, gin.H{"code": code, "info": info, "category": category, "requestId": requestID})
}
