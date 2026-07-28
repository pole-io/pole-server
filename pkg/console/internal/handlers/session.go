package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/pole-io/pole-server/pkg/console/internal/common/api"
	"github.com/pole-io/pole-server/pkg/console/internal/common/model"
)

type ConsoleSessionView struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Admin  bool   `json:"admin"`
}

// DescribeConsoleSession returns identity attributes from the signed HttpOnly
// session. The browser must not use localStorage as an authorization source.
func DescribeConsoleSession(config *bootstrap.Config) gin.HandlerFunc {
	adminGetter := &AdminUserGetter{conf: config}
	return func(ctx *gin.Context) {
		claims, err := parseJWTClaims(ctx, config)
		if err != nil || claims == nil || claims.UserID == "" {
			if err == nil {
				err = errors.New("console session is missing")
			}
			response := model.NewResponse(int32(api.ExecuteException))
			response.Data = err.Error()
			ctx.JSON(http.StatusUnauthorized, response)
			return
		}
		role, admin := resolveSystemConfigurationSessionRole(claims, adminGetter)
		if role != claims.Role {
			_ = refreshJWTWithRole(ctx, claims.UserID, claims.Token, role, config)
		}

		response := model.NewResponse(int32(api.ExecuteSuccess))
		response.Data = ConsoleSessionView{
			UserID: claims.UserID,
			Role:   role,
			Admin:  admin,
		}
		ctx.JSON(http.StatusOK, response)
	}
}
