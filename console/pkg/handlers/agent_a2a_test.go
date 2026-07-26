package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/console/bootstrap"
)

func TestInvalidA2AAdvertisedURLFailsClosedWithoutPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := &bootstrap.Config{
		Agent: bootstrap.AgentConfig{
			Definition: bootstrap.AgentDefinitionConfig{
				ID: "pole-control-plane",
				SystemPrompt: bootstrap.AgentSystemPromptConfig{
					BuiltinVersion: "v1",
				},
			},
			A2A: bootstrap.AgentA2AConfig{
				Name: "Pole Agent", Endpoint: "://invalid",
			},
		},
	}
	handler := NewAgentHandler(config, nil, nil)
	router := gin.New()
	router.GET("/.well-known/agent-card.json", handler.A2ACard)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil))
	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Contains(t, recorder.Body.String(), "initialize Pole Agent A2A adapter")
}
