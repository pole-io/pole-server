package router

import (
	"github.com/gin-gonic/gin"

	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/handlers"
	"github.com/pole-io/pole-server/console/pkg/systemsettings"
)

func SystemConfigurationRouter(webSvr *gin.Engine, config *bootstrap.Config, manager *systemsettings.Manager) {
	v1 := webSvr.Group("/system-config/v1")
	v1.Use(handlers.RequireSystemConfigurationAdmin(config))
	v1.GET("/settings", handlers.DescribeSystemConfiguration(config, manager))
	v1.GET("/domains/:component/:domain/draft", handlers.GetSystemConfigurationDomain(config, manager))
	v1.PUT("/domains/:component/:domain/draft", handlers.SaveSystemConfigurationDraft(config, manager))
	v1.POST("/domains/:component/:domain/connection-test", handlers.TestSystemConfigurationConnection(config, manager))
	v1.POST("/domains/:component/:domain/publish", handlers.PublishSystemConfiguration(config, manager))
	v1.GET("/domains/:component/:domain/releases", handlers.ListSystemConfigurationReleases(manager))
}
