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
	"github.com/gin-gonic/gin"

	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/handlers"
)

// DiscoveryV1Router 路由请求
func DiscoveryV1Router(r *gin.Engine, config *bootstrap.Config) {
	// 后端server路由组
	v1 := r.Group(config.WebServer.NamingURL)
	// Handle any path with any HTTP method
	v1.Any("/*path", handlers.ReverseProxyForServer(&config.PoleServer, config))
}
