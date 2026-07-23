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

package handlers

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
)

// PolarisPage polaris页面
func PolarisPage(conf *bootstrap.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// index.html 引用带内容哈希的静态资源，入口本身必须每次重新校验。
		// 否则 Console 重启后，浏览器可能继续运行旧 SPA，导致新路由或工作模式不可用。
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File(filepath.Join(conf.WebServer.WebPath, "index.html"))
	}
}
