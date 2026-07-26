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

package xdsserverv3

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/plugin/apiserver/xdsserverv3/resource"
)

type VHDSBuilder struct {
	svr service.DiscoverServer
}

// Init
func (vhds *VHDSBuilder) Init(svr service.DiscoverServer) {
	vhds.svr = svr
}

// Generate
func (vhds *VHDSBuilder) Generate(option *resource.BuildOption) (interface{}, error) {
	var (
		hosts []types.Resource
	)
	// step 1: 生成服务的 OUTBOUND 规则
	services := option.Services
	for svcKey, serviceInfo := range services {
		vHost := &route.VirtualHost{
			Name:    resource.MakeVHDSServiceName(resource.OutBoundRouteConfigName+"/", svcKey),
			Domains: resource.GenerateServiceDomains(serviceInfo),
			Routes:  vhds.makeSidecarOutBoundRoutes(corev3.TrafficDirection_OUTBOUND, serviceInfo, option),
		}
		hosts = append(hosts, vHost)
	}
	return hosts, nil
}

// makeSidecarOutBoundRoutes .
func (vhds *VHDSBuilder) makeSidecarOutBoundRoutes(trafficDirection corev3.TrafficDirection,
	serviceInfo *resource.ServiceInfo, opt *resource.BuildOption) []*route.Route {
	var (
		routes        []*route.Route
		matchAllRoute *route.Route
	)
	rules := resource.FilterOutboundRouterRules(serviceInfo, opt.SelfService)
	for _, rule := range rules {
		trafficMatch := rule.GetArguments()
		if !resource.SupportsEnvoyTrafficMatch(trafficMatch) {
			log.Warn("[XDS][Envoy] skip route with unsupported traffic match",
				zap.String("service", serviceInfo.ServiceKey.Domain()),
				zap.String("rule", rule.GetName()))
			continue
		}
		matchAll := len(trafficMatch.GetArguments()) == 0

		routeMatch := &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
		}
		if !matchAll {
			resource.BuildSidecarRouteMatch(routeMatch, trafficMatch)
		}

		currentRoute := resource.MakeSidecarRoute(
			trafficDirection, routeMatch, serviceInfo, rule.GetDestinations(), opt)
		if matchAll {
			matchAllRoute = currentRoute
		} else {
			routes = append(routes, currentRoute)
		}
	}
	if matchAllRoute == nil {
		// 如果没有路由，会进入最后的默认处理
		routes = append(routes, resource.MakeDefaultRoute(trafficDirection, serviceInfo.ServiceKey, opt))
	} else {
		routes = append(routes, matchAllRoute)
	}
	return routes
}
