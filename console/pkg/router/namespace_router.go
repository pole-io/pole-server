package router

import (
	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/handlers"
)

// NamespaceRouter 路由请求
func NamespaceRouter(r *gin.Engine, config *bootstrap.Config) {
	// 后端server路由组
	v1 := r.Group(config.WebServer.CoreURL)
	// Handle any path with any HTTP method
	v1.Any("/*path", handlers.ReverseProxyForServer(&config.PoleServer, config))
}
