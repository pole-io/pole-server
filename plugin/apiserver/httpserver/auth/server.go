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

package auth

import (
	"strconv"

	"github.com/emicklei/go-restful/v3"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/pkg/admin"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

const (
	defaultReadAccess string = "default-read"
	defaultAccess     string = "default"
	aimcpAccess       string = "aimcp"

	basePath string = "/ai/mcp/v1"

	sseEp string = "/sse"
	msgEp string = "/message"
)

// HTTPServer
type HTTPServer struct {
	maintainServer admin.AdminOperateServer

	userSvr   authapi.UserServer
	policySvr authapi.StrategyServer
}

// NewServer 创建配置中心的 HttpServer
func NewServer(userSvr authapi.UserServer, policySvr authapi.StrategyServer) *HTTPServer {
	return &HTTPServer{
		userSvr:   userSvr,
		policySvr: policySvr,
	}
}

// GetAuthServer 运维接口
func (h *HTTPServer) GetAuthServer() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/auth/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	// 鉴权系统信息
	h.addSystemAccess(ws)
	// 用户
	h.addUserAccess(ws)
	// 用户组
	h.addGroupAccess(ws)
	// 角色
	h.addRoleAccess(ws)
	// 策略
	h.addPolicyRuleAccess(ws)
	return ws
}

func (h *HTTPServer) addSystemAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichAuthStatusApiDocs(ws.GET("/system").To(h.AuthStatus)))
}

// AuthStatus auth status
func (h *HTTPServer) AuthStatus(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	checker := h.policySvr.GetAuthChecker()

	isOpen := (checker.IsOpenClientAuth() || checker.IsOpenConsoleAuth())
	resp := api.NewAuthResponse(apimodel.Code_ExecuteSuccess)
	resp.OptionSwitch = &apiservice.OptionSwitch{
		Options: map[string]string{
			"auth":        strconv.FormatBool(isOpen),
			"clientOpen":  strconv.FormatBool(checker.IsOpenClientAuth()),
			"consoleOpen": strconv.FormatBool(checker.IsOpenConsoleAuth()),
		},
	}

	handler.WriteHeaderAndProto(resp)
}
