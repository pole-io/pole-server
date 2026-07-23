package router

import (
	"github.com/gin-gonic/gin"

	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
	"github.com/pole-io/pole-server/console/pkg/handlers"
	store "github.com/pole-io/pole-server/console/pkg/observer"
	"github.com/pole-io/pole-server/console/pkg/systemsettings"
)

func NewAgentRuntime(config *bootstrap.Config) (*agentworkbench.Workbench, *systemsettings.Manager, error) {
	agentConfig := config.Agent.Normalize()
	port := agentworkbench.NewHTTPConfigFilePort(config.PoleServer.Address, nil)
	workbench := agentworkbench.New(port, agentworkbench.Options{TTL: agentConfig.ProposalTTL})
	observerStore, err := store.GetStore()
	var repository store.SystemSettingsRepository
	if err != nil {
		if config.Store.Name != "" {
			return nil, nil, err
		}
		repository = systemsettings.NewMemoryRepository()
	} else {
		repository = observerStore
	}
	manager, err := systemsettings.NewManager(config, repository, workbench)
	return workbench, manager, err
}

func AgentRouter(r *gin.Engine, config *bootstrap.Config, workbench *agentworkbench.Workbench,
	runtime *systemsettings.Manager) {
	handler := handlers.NewAgentHandler(config, workbench, runtime)
	v1 := r.Group("/ai/agent/v1")
	v1.GET("/runtime", handler.Runtime)
	v1.POST("/turns", handler.RunTurn)
	v1.POST("/proposals/config-file", handler.PrepareConfigFile)
	v1.POST("/proposals/:proposal_id/confirm", handler.ConfirmProposal)
}
