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

// AuthRouter 路由请求
func AuthRouter(webSvr *gin.Engine, config *bootstrap.Config) {
	// 后端server路由组
	v1 := webSvr.Group(config.WebServer.AuthURL)
	// 用户登陆
	v1.POST("/user/login", handlers.ReverseProxyForLogin(&config.PoleServer, config))
	// 从 Console 签名会话获取当前用户的权威角色
	v1.GET("/user/session", handlers.DescribeConsoleSession(config))
	// 获取用户列表
	v1.GET("/users", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量创建用户
	v1.POST("/users", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量删除用户
	v1.POST("/users/delete", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 更新用户
	v1.PUT("/users", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 更新用户密码
	v1.PUT("/user/password", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 查询用户Token信息
	v1.GET("/user/token", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 更新用户Token可用状态
	v1.PUT("/user/token/enable", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 重置用户Token
	v1.PUT("/user/token/refresh", handlers.ReverseProxyForServer(&config.PoleServer, config))

	// 创建用户组
	v1.POST("/usergroups", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量更新用户组
	v1.PUT("/usergroups", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 获取用户组列表
	v1.GET("/usergroups", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量删除用户组
	v1.POST("/usergroups/delete", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 查询用户组详细
	v1.GET("/usergroup/detail", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 查询用户组Token
	v1.GET("/usergroup/token", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 更新用户组Token可用状态
	v1.PUT("/usergroup/token/enable", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 重置用户组Token
	v1.PUT("/usergroup/token/refresh", handlers.ReverseProxyForServer(&config.PoleServer, config))

	// 创建鉴权策略
	v1.POST("/policies", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 查询鉴权策略详细
	v1.GET("/policies/detail", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量更新鉴权策略
	v1.PUT("/policies", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量删除鉴权策略
	v1.POST("/policies/delete", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 获取鉴权策略列表
	v1.GET("/policies", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 查看用户/用户组所能操作的所有资源列表数据
	v1.GET("/principal/resources", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 查看某个资源下关联的操作人员
	v1.GET("/resources/principals", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 资源授权
	v1.POST("/resources/authorize", handlers.ReverseProxyForServer(&config.PoleServer, config))

	// 创建鉴权策略
	v1.POST("/roles", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量更新鉴权策略
	v1.PUT("/roles", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 批量删除鉴权策略
	v1.POST("/roles/delete", handlers.ReverseProxyForServer(&config.PoleServer, config))
	// 获取鉴权策略列表
	v1.GET("/roles", handlers.ReverseProxyForServer(&config.PoleServer, config))
}
