/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/pole-io/pole-server/pkg/console/internal/common/model"
	"github.com/pole-io/pole-server/pkg/console/internal/handlers"
)

// AdminRouter 路由请求
func AdminRouter(webSvr *gin.Engine, config *bootstrap.Config) {
	// 后端server路由组
	v1 := webSvr.Group("/admin/v1")
	v1.GET("/server/functions", handlers.ReverseProxyNoAuthForServer(&config.PoleServer, config))
	v1.GET("/mainuser/exist", handlers.ReverseHandleAdminUserExist(&config.PoleServer, config))
	v1.POST("/mainuser/create", handlers.ReverseProxyNoAuthForServer(&config.PoleServer, config))
	v1.GET("/bootstrap/config", handlers.ReverseHandleBootstrap(&config.PoleServer, config))
	v1.GET("/console/ability", func(ctx *gin.Context) {
		futures := strings.Split(config.Futures, ",")
		resp := model.Response{
			Code: 200000,
			Info: "success",
			Data: futures,
		}
		ctx.JSON(http.StatusOK, resp)
	})
}
