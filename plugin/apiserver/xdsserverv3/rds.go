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
	"fmt"
	"sort"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/plugin/apiserver/xdsserverv3/resource"
)

// RDSBuilder .
type RDSBuilder struct {
	VHDSBuilder
	svr service.DiscoverServer
}

func (rds *RDSBuilder) Init(svr service.DiscoverServer) {
	rds.VHDSBuilder = VHDSBuilder{
		svr: svr,
	}
	rds.svr = svr
}

func (rds *RDSBuilder) Generate(option *resource.BuildOption) (interface{}, error) {
	var resources []types.Resource

	switch option.RunType {
	case resource.RunTypeGateway:
		// Envoy Gateway 场景只需要支持入流量场景处理即可
		ret, err := rds.makeGatewayRouteConfiguration(option)
		if err != nil {
			return nil, err
		}
		resources = ret
	case resource.RunTypeSidecar:
		switch option.TrafficDirection {
		case corev3.TrafficDirection_INBOUND:
			resources = append(resources, rds.makeSidecarInBoundRouteConfiguration(option)...)
		case corev3.TrafficDirection_OUTBOUND:
			resources = append(resources, rds.makeSidecarOutBoundRouteConfiguration(option)...)
		}
	}
	return resources, nil
}

func (rds *RDSBuilder) makeSidecarInBoundRouteConfiguration(option *resource.BuildOption) []types.Resource {
	selfService := option.SelfService
	// step 2: 生成 sidecar 所属服务的 INBOUND 规则
	// 服务信息不存在或者不精确，不下发 InBound RDS 规则信息
	if !isExactSvc(selfService) {
		return []types.Resource{}
	}
	routeConf := &route.RouteConfiguration{
		Name:             resource.MakeInBoundRouteConfigName(selfService, option.IsDemand()),
		ValidateClusters: wrapperspb.Bool(false),
	}

	if !option.ForceDelete {
		routeConf.VirtualHosts = []*route.VirtualHost{
			{
				Name:    resource.MakeServiceName(selfService, corev3.TrafficDirection_INBOUND, option),
				Domains: []string{"*"},
				Routes:  rds.makeSidecarInBoundRoutes(selfService, corev3.TrafficDirection_INBOUND, option),
			},
		}
	}
	return []types.Resource{
		routeConf,
	}
}

func (rds *RDSBuilder) makeSidecarOutBoundRouteConfiguration(option *resource.BuildOption) []types.Resource {
	// 每个 polaris serviceInfo 对应一个 virtualHost
	var (
		routeConfs []types.Resource
		hosts      []*route.VirtualHost
	)
	baseRouteName := resource.OutBoundRouteConfigName
	if !option.ForceDelete {
		// step 1: 生成服务的 OUTBOUND 规则
		services := option.Services
		for svcKey, serviceInfo := range services {
			vHost := &route.VirtualHost{
				Name:    resource.MakeServiceName(svcKey, corev3.TrafficDirection_OUTBOUND, option),
				Domains: resource.GenerateServiceDomains(serviceInfo),
				Routes:  rds.makeSidecarOutBoundRoutes(corev3.TrafficDirection_OUTBOUND, serviceInfo, option),
			}
			hosts = append(hosts, vHost)
		}
	}
	hosts = append(hosts, resource.BuildAllowAnyVHost())
	if option.IsDemand() {
		baseRouteName = fmt.Sprintf("%s|%s|demand", resource.OutBoundRouteConfigName, option.Namespace)
		// routeConfiguration.Vhds = &route.Vhds{
		// 	ConfigSource: &corev3.ConfigSource{
		// 		ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
		// 			ApiConfigSource: &corev3.ApiConfigSource{
		// 				ApiType:             corev3.ApiConfigSource_DELTA_GRPC,
		// 				TransportApiVersion: corev3.ApiVersion_V3,
		// 				GrpcServices: []*corev3.GrpcService{
		// 					{
		// 						TargetSpecifier: &corev3.GrpcService_GoogleGrpc_{
		// 							GoogleGrpc: &corev3.GrpcService_GoogleGrpc{
		// 								TargetUri:  option.OnDemandServer,
		// 								StatPrefix: "polaris_vhds",
		// 							},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 		ResourceApiVersion: corev3.ApiVersion_V3,
		// 	},
		// }
	}
	routeConfiguration := &route.RouteConfiguration{
		Name:             baseRouteName,
		ValidateClusters: wrapperspb.Bool(false),
	}
	routeConfiguration.VirtualHosts = hosts
	routeConfs = append(routeConfs, routeConfiguration)
	return routeConfs
}

func (rds *RDSBuilder) makeSidecarInBoundRoutes(selfService svctypes.ServiceKey,
	trafficDirection corev3.TrafficDirection, opt *resource.BuildOption) []*route.Route {
	currentRoute := &route.Route{
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: resource.MakeServiceName(selfService, trafficDirection, opt),
				},
			},
		},
	}

	var rateLimits []*traffic_manage.RateLimit
	if selfInfo := opt.Services[selfService]; selfInfo != nil {
		rateLimits = selfInfo.RateLimits
	}
	limits, typedPerFilterConfig, err := resource.MakeSidecarLocalRateLimitFromRules(rateLimits, selfService)
	if err == nil {
		currentRoute.TypedPerFilterConfig = typedPerFilterConfig
		if opt.IsDemand() {
			currentRoute.TypedPerFilterConfig[resource.EnvoyHttpFilter_OnDemand] =
				resource.BuildOnDemandRouteTypedPerFilterConfig()
		}
		currentRoute.GetRoute().RateLimits = limits
	}
	return []*route.Route{
		currentRoute,
	}
}

// ---------------------- Envoy Gateway ---------------------- //
func (rds *RDSBuilder) makeGatewayRouteConfiguration(option *resource.BuildOption) ([]types.Resource, error) {
	// 每个 polaris serviceInfo 对应一个 virtualHost
	var (
		routeConfs []types.Resource
		hosts      []*route.VirtualHost
	)

	routes, err := rds.makeGatewayRoutes(option)
	if err != nil {
		return nil, err
	}
	if len(routes) == 0 {
		return []types.Resource{}, nil
	}

	vHost := &route.VirtualHost{
		Name:    "gateway-virtualhost",
		Domains: resource.MakeServiceGatewayDomains(),
		Routes:  routes,
	}
	hosts = append(hosts, vHost)
	routeConfiguration := &route.RouteConfiguration{
		Name:         resource.OutBoundRouteConfigName + "-gateway",
		VirtualHosts: hosts,
	}
	return append(routeConfs, routeConfiguration), nil
}

func isExactSvc(s svctypes.ServiceKey) bool {
	return s.Namespace != "" && s.Namespace != rules.MatchAll && s.Name != "" && s.Name != rules.MatchAll
}

// makeGatewayRoutes builds the route.Route list for the envoy_gateway scenario
// In this scenario, it is mainly for the rule forwarding of path, /serviceA => serviceA
// Currently only routing rules that meet the following conditions support xds converted to envoy_gateway
// require 1: The calling service must match the GatewayService & GatewayNamespace in NodeProxy Metadata
// require 2: The $path parameter must be set in the request tag
// require 3: The information of the called service must be accurate, that is, a clear namespace and service
func (rds *RDSBuilder) makeGatewayRoutes(option *resource.BuildOption) ([]*route.Route, error) {

	routes := make([]*route.Route, 0, 16)
	selfService := option.SelfService
	callerService := selfService.Name
	callerNamespace := selfService.Namespace
	if !isExactSvc(selfService) {
		return nil, nil
	}

	serviceKeys := make([]svctypes.ServiceKey, 0, len(option.Services))
	for key := range option.Services {
		serviceKeys = append(serviceKeys, key)
	}
	sort.Slice(serviceKeys, func(i, j int) bool {
		return serviceKeys[i].Domain() < serviceKeys[j].Domain()
	})
	for _, key := range serviceKeys {
		serviceInfo := option.Services[key]
		for _, subRule := range resource.FilterOutboundRouterRules(serviceInfo, selfService) {
			trafficMatch := subRule.GetArguments()
			if !resource.SupportsEnvoyTrafficMatch(trafficMatch) {
				log.Warn("[XDS][Envoy] skip gateway route with unsupported traffic match",
					zap.String("caller", callerNamespace+"/"+callerService),
					zap.String("callee", serviceInfo.Namespace+"/"+serviceInfo.Name),
					zap.String("rule", subRule.GetName()))
				continue
			}
			routeMatch := &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
			}
			if trafficMatch != nil {
				resource.BuildSidecarRouteMatch(routeMatch, trafficMatch)
			}

			gatewayRoute := resource.MakeGatewayRoute(corev3.TrafficDirection_OUTBOUND, routeMatch,
				subRule.GetDestinations(), option)
			pathInfo := gatewayRoute.GetMatch().GetPath()
			if pathInfo == "" {
				pathInfo = gatewayRoute.GetMatch().GetSafeRegex().GetRegex()
			}

			var rateLimits []*traffic_manage.RateLimit
			if selfInfo := option.Services[selfService]; selfInfo != nil {
				rateLimits = selfInfo.RateLimits
			}
			limits, typedPerFilterConfig, err := resource.MakeGatewayLocalRateLimitFromRules(
				rateLimits, pathInfo, selfService)
			if err == nil {
				gatewayRoute.TypedPerFilterConfig = typedPerFilterConfig
				gatewayRoute.GetRoute().RateLimits = limits
			}
			routes = append(routes, gatewayRoute)
		}
	}

	routes = append(routes, &route.Route{
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: resource.PassthroughClusterName,
				},
			},
		},
	})
	return routes, nil
}
