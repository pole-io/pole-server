package router

import (
	"github.com/gin-gonic/gin"

	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/pole-io/pole-server/pkg/console/internal/handlers"
)

// SkillMarketplaceRouter exposes the control-plane Marketplace API through the
// authenticated Console origin. Public Registry clients use port 8090 directly.
func SkillMarketplaceRouter(r *gin.Engine, config *bootstrap.Config) {
	v1 := r.Group("/api/skill-marketplace")
	v1.Any("/*path", handlers.ReverseProxyForServer(&config.PoleServer, config))
}
