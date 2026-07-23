package router

import (
	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/handlers"
)

func ObservabilityRouter(webSvr *gin.Engine, config *bootstrap.Config) {
	v1 := webSvr.Group("/observability/v1")
	v1.GET("/platform/overview", handlers.DescribeObservabilityPlatformOverview(config))
	v1.GET("/events", handlers.DescribeObservabilityEvents(config))
	v1.GET("/operations", handlers.DescribeObservabilityOperations(config))
}
