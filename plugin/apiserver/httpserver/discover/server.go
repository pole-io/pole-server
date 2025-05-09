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

package discover

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/pole-io/pole-server/apis/apiserver"
	"github.com/pole-io/pole-server/pkg/goverrule"
	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/pkg/service/healthcheck"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
)

type HTTPServer struct {
	namingServer      service.DiscoverServer
	ruleServer        goverrule.GoverRuleServer
	healthCheckServer *healthcheck.Server
}

func NewServer(
	namingServer service.DiscoverServer,
	ruleServer goverrule.GoverRuleServer,
	healthCheckServer *healthcheck.Server) *HTTPServer {
	return &HTTPServer{
		namingServer:      namingServer,
		ruleServer:        ruleServer,
		healthCheckServer: healthCheckServer,
	}
}

const (
	defaultReadAccess    string = "default-read"
	defaultAccess        string = "default"
	serviceAccess        string = "service"
	circuitBreakerAccess string = "circuitbreaker"
	routingAccess        string = "router"
	rateLimitAccess      string = "ratelimit"
)

// GetConsoleAccessServer 注册管理端接口
func (h *HTTPServer) GetConsoleAccessServer(include []string) *restful.WebService {
	consoleAccess := []string{defaultAccess}

	ws := new(restful.WebService)

	ws.Path("/naming/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// 如果为空，则开启全部接口
	if len(include) == 0 {
		include = consoleAccess
	}
	oldInclude := include

	for _, item := range oldInclude {
		if item == defaultReadAccess {
			include = []string{defaultReadAccess}
			break
		}
	}

	for _, item := range oldInclude {
		if item == defaultAccess {
			include = consoleAccess
			break
		}
	}

	for _, item := range include {
		switch item {
		case defaultReadAccess:
			h.addDefaultReadAccess(ws)
		case defaultAccess:
			h.addDefaultAccess(ws)
		case serviceAccess:
			h.addServiceAccess(ws)
			h.addInstanceAccess(ws)
		case circuitBreakerAccess:
			h.addCircuitBreakerRuleAccess(ws)
		case routingAccess:
			h.addRoutingRuleAccess(ws)
			h.addLaneRuleAccess(ws)
		case rateLimitAccess:
			h.addRateLimitRuleAccess(ws)
		}
	}
	return ws
}

// addDefaultReadAccess 增加默认读接口
func (h *HTTPServer) addDefaultReadAccess(ws *restful.WebService) {
	// 管理端接口：只包含读接口
	ws.Route(docs.EnrichGetServicesApiDocs(ws.GET("/services").To(h.GetServices)))
	ws.Route(docs.EnrichGetServicesCountApiDocs(ws.GET("/services/count").To(h.GetServicesCount)))
	ws.Route(docs.EnrichGetServiceAliasesApiDocs(ws.GET("/service/aliases").To(h.GetServiceAliases)))

	ws.Route(docs.EnrichGetInstancesApiDocs(ws.GET("/instances").To(h.GetInstances)))
	ws.Route(docs.EnrichGetInstancesCountApiDocs(ws.GET("/instances/count").To(h.GetInstancesCount)))
	ws.Route(docs.EnrichGetRateLimitsApiDocs(ws.GET("/ratelimits").To(h.GetRateLimits)))
	ws.Route(docs.EnrichGetCircuitBreakerRulesApiDocs(
		ws.GET("/circuitbreaker/rules").To(h.GetCircuitBreakerRules)))
	ws.Route(docs.EnrichGetFaultDetectRulesApiDocs(ws.GET("/faultdetectors").To(h.GetFaultDetectRules)))
	ws.Route(docs.EnrichGetServiceContractsApiDocs(
		ws.GET("/service/contracts").To(h.GetServiceContracts)))
	ws.Route(docs.EnrichGetServiceContractsApiDocs(
		ws.GET("/service/contract/versions").To(h.GetServiceContractVersions)))
	ws.Route(ws.GET("/routings").To(h.GetRoutings))
}

// addDefaultAccess 增加默认接口
func (h *HTTPServer) addDefaultAccess(ws *restful.WebService) {
	// 管理端接口：增删改查请求全部操作存储层
	h.addServiceAccess(ws)
	h.addInstanceAccess(ws)
	h.addRoutingRuleAccess(ws)
	h.addLaneRuleAccess(ws)
	h.addRateLimitRuleAccess(ws)
	h.addCircuitBreakerRuleAccess(ws)
	h.addFaultDetectRuleAccess(ws)
}

// GetClientAccessServer get client access server
func (h *HTTPServer) GetClientAccessServer(ws *restful.WebService, include []string) {
	clientAccess := []string{apiserver.DiscoverAccess, apiserver.RegisterAccess, apiserver.HealthcheckAccess}

	// 如果为空，则开启全部接口
	if len(include) == 0 {
		include = clientAccess
	}

	// 客户端接口：增删改请求操作存储层，查请求访问缓存
	for _, item := range include {
		switch item {
		case apiserver.DiscoverAccess:
			h.addDiscoverAccess(ws)
		case apiserver.RegisterAccess:
			h.addRegisterAccess(ws)
		case apiserver.HealthcheckAccess:
			h.addHealthCheckAccess(ws)
		}
	}
}
