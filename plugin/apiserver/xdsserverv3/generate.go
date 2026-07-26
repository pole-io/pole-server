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
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/observability/statis"
	apitypes "github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/plugin/apiserver/xdsserverv3/cache"
	"github.com/pole-io/pole-server/plugin/apiserver/xdsserverv3/resource"
)

var (
	ErrorNoSupportXDSType = errors.New("unsupport xds build type")
)

type (
	ServiceInfos               map[string]map[svctypes.ServiceKey]*resource.ServiceInfo
	CurrentServiceInfoProvider func() ServiceInfos
	governanceRuleProvider     interface {
		GetRouterRuleWithCache(context.Context, *apiservice.Service) *apiservice.DiscoverResponse
		GetRateLimitWithCache(context.Context, *apiservice.Service) *apiservice.DiscoverResponse
		GetCircuitBreakerWithCache(context.Context, *apiservice.Service) *apiservice.DiscoverResponse
		GetFaultDetectWithCache(context.Context, *apiservice.Service) *apiservice.DiscoverResponse
	}
)

// XdsResourceGenerator is the xDS resource generator
type XdsResourceGenerator struct {
	namingServer    service.DiscoverServer
	ruleServer      governanceRuleProvider
	cache           *cache.ResourceCache
	versionNum      *atomic.Uint64
	xdsNodesMgr     *resource.XDSNodeManager
	svcInfoProvider CurrentServiceInfoProvider
}

// Generate 构建 XDS 资源缓存数据信息
func (x *XdsResourceGenerator) Generate(versionLocal string, needUpdate, needRemove ServiceInfos) {
	updateRequest := cache.NewUpdateResourcesRequest()

	deltaOp := func(runType resource.RunType, infos ServiceInfos, isRemove bool) {
		direction := corev3.TrafficDirection_OUTBOUND
		if runType == resource.RunTypeGateway {
			direction = corev3.TrafficDirection_INBOUND
		}
		generate := func(opt *resource.BuildOption) {
			opt.CloseEnvoyDemand()
			opt.TLSMode = resource.TLSModeNone
			// 默认构建没有设置 TLS 的 CDS 资源
			x.buildUpdateRequest(updateRequest, resource.CDS, opt, isRemove)
			// 构建设置了 TLS Mode == Strict 的 CDS 资源
			opt.TLSMode = resource.TLSModeStrict
			x.buildUpdateRequest(updateRequest, resource.CDS, opt, isRemove)
			// 构建设置了 TLS Mode == Permissive 的 CDS 资源
			opt.TLSMode = resource.TLSModePermissive
			x.buildUpdateRequest(updateRequest, resource.CDS, opt, isRemove)
			// 恢复 TLSMode
			opt.TLSMode = resource.TLSModeNone
			x.buildUpdateRequest(updateRequest, resource.EDS, opt, isRemove)
			x.buildUpdateRequest(updateRequest, resource.RDS, opt, isRemove)
			// 开启按需 Demand
			opt.OpenEnvoyDemand()
			x.buildUpdateRequest(updateRequest, resource.RDS, opt, isRemove)
		}

		// CDS/EDS/VHDS 一起构建
		for namespace, services := range infos {
			opt := &resource.BuildOption{
				RunType:          runType,
				Namespace:        namespace,
				Services:         services,
				TrafficDirection: direction,
			}
			// sidecar 和 gateway 大部份资源都是复用的，所以这里只需要构建一次即可，gateway 只有 RDS/LDS 存在特别，单独针对构建即可
			if runType == resource.RunTypeSidecar {
				generate(opt)
				x.buildUpdateRequest(updateRequest, resource.VHDS, opt, isRemove)
			}

			if runType == resource.RunTypeSidecar {
				for svcKey := range services {
					// 换成 INBOUND 构建 CDS、EDS、RDS
					opt.SelfService = svcKey
					opt.TrafficDirection = corev3.TrafficDirection_INBOUND
					generate(opt)
				}
			}
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		// 处理 Sideacr
		deltaOp(resource.RunTypeSidecar, needUpdate, false)
		deltaOp(resource.RunTypeSidecar, needRemove, true)
	}()

	go func() {
		defer wg.Done()
		// 处理 Gateway
		deltaOp(resource.RunTypeGateway, needUpdate, false)
		deltaOp(resource.RunTypeGateway, needRemove, true)
	}()

	wg.Wait()

	if err := x.cache.UpdateResources(context.Background(), updateRequest); err != nil {
		log.Error("[XDS][Envoy] update xds resource fail", zap.Error(err))
	}
	x.RefreshNodeCaches()
}

func (x *XdsResourceGenerator) buildOneEnvoyXDSCache(node *resource.XDSClient) error {
	ldsResources, nodeResources, err := x.buildEnvoyNodeResources(node)
	if err != nil {
		return err
	}
	return x.cache.UpdateResources(context.Background(), &cache.UpdateResourcesRequest{
		Lds: map[string]map[string]types.Resource{
			node.ID: ldsResources,
		},
		NodeResources: map[string]map[resource.XDSType]map[string]types.Resource{
			node.ID: nodeResources,
		},
	})
}

func (x *XdsResourceGenerator) buildEnvoyNodeResources(
	node *resource.XDSClient,
) (map[string]types.Resource, map[resource.XDSType]map[string]types.Resource, error) {
	services, err := x.servicesForNode(node)
	if err != nil {
		return nil, nil, err
	}
	return x.buildEnvoyNodeResourcesWithServices(node, services)
}

func (x *XdsResourceGenerator) buildEnvoyNodeResourcesWithServices(
	node *resource.XDSClient,
	services map[svctypes.ServiceKey]*resource.ServiceInfo,
) (map[string]types.Resource, map[resource.XDSType]map[string]types.Resource, error) {
	opt := &resource.BuildOption{
		RunType:   node.RunType,
		Client:    node,
		TLSMode:   node.TLSMode,
		Namespace: node.GetSelfNamespace(),
		Services:  services,
		SelfService: svctypes.ServiceKey{
			Namespace: node.GetSelfNamespace(),
			Name:      node.GetSelfService(),
		},
	}
	if node.OpenOnDemand {
		opt.OpenEnvoyDemand()
	}

	nodeResources := make(map[resource.XDSType]map[string]types.Resource)
	buildCache := func(xdsType resource.XDSType, opt *resource.BuildOption) error {
		xxds, err := x.generateXDSResource(xdsType, opt)
		if err != nil {
			return fmt.Errorf("generate %v resource: %w", xdsType, err)
		}
		if _, ok := nodeResources[xdsType]; !ok {
			nodeResources[xdsType] = make(map[string]types.Resource)
		}
		for name, item := range cachev3.IndexRawResourcesByName(xxds) {
			nodeResources[xdsType][name] = item
		}
		return nil
	}

	opt.TrafficDirection = corev3.TrafficDirection_OUTBOUND
	for _, xdsType := range []resource.XDSType{resource.LDS, resource.CDS, resource.RDS} {
		if err := buildCache(xdsType, opt); err != nil {
			return nil, nil, err
		}
	}
	if node.RunType == resource.RunTypeSidecar {
		if err := buildCache(resource.VHDS, opt); err != nil {
			return nil, nil, err
		}
		opt.TrafficDirection = corev3.TrafficDirection_INBOUND
		for _, xdsType := range []resource.XDSType{resource.LDS, resource.CDS, resource.RDS} {
			if err := buildCache(xdsType, opt); err != nil {
				return nil, nil, err
			}
		}
	}

	ldsResources := nodeResources[resource.LDS]
	delete(nodeResources, resource.LDS)
	return ldsResources, nodeResources, nil
}

func (x *XdsResourceGenerator) RefreshNodeCaches() {
	updateRequest := cache.NewUpdateResourcesRequest()
	selectedByContext := make(map[string]map[svctypes.ServiceKey]*resource.ServiceInfo)
	for _, node := range x.xdsNodesMgr.ListEnvoyNodes() {
		selectionKey := governanceSelectionKey(node)
		services, ok := selectedByContext[selectionKey]
		var err error
		if !ok {
			services, err = x.servicesForNode(node)
			if err == nil {
				selectedByContext[selectionKey] = services
			}
		}
		if err != nil {
			log.Error("[XDS][Envoy] select node governance snapshot",
				zap.String("node", node.ID), zap.Error(err))
			continue
		}
		ldsResources, nodeResources, err := x.buildEnvoyNodeResourcesWithServices(node, services)
		if err != nil {
			log.Error("[XDS][Envoy] refresh node policy snapshot", zap.String("node", node.ID), zap.Error(err))
			continue
		}
		updateRequest.Lds[node.ID] = ldsResources
		updateRequest.NodeResources[node.ID] = nodeResources
	}
	if len(updateRequest.Lds) == 0 {
		return
	}
	if err := x.cache.UpdateResources(context.Background(), updateRequest); err != nil {
		log.Error("[XDS][Envoy] commit node policy snapshots", zap.Error(err))
	}
}

func governanceSelectionKey(node *resource.XDSClient) string {
	labels := node.GovernanceLabels()
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var keyBuilder strings.Builder
	writeSelectionKeyPart(&keyBuilder, node.GetSelfNamespace())
	for _, key := range keys {
		writeSelectionKeyPart(&keyBuilder, key)
		writeSelectionKeyPart(&keyBuilder, labels[key])
	}
	return keyBuilder.String()
}

func writeSelectionKeyPart(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
}

func (x *XdsResourceGenerator) servicesForNode(
	node *resource.XDSClient,
) (map[svctypes.ServiceKey]*resource.ServiceInfo, error) {
	current := x.svcInfoProvider()
	base := current[node.GetSelfNamespace()]
	selected := make(map[svctypes.ServiceKey]*resource.ServiceInfo, len(base))
	ctx := governanceContextForNode(context.Background(), node)
	for key, info := range base {
		cloned := *info
		serviceReq := &apiservice.Service{Name: info.Name, Namespace: info.Namespace, Revision: "-1"}

		routerResp := x.ruleServer.GetRouterRuleWithCache(ctx, serviceReq)
		if routerResp.GetCode() != api.ExecuteSuccess {
			return nil, fmt.Errorf("select route rules for %s: %s", key.Domain(), routerResp.GetInfo())
		}
		cloned.Routing = routerResp.GetCustomRouteRules()
		cloned.SvcRoutingRevision = routerResp.GetService().GetRevision()

		rateResp := x.ruleServer.GetRateLimitWithCache(ctx, serviceReq)
		if rateResp.GetCode() != api.ExecuteSuccess {
			return nil, fmt.Errorf("select rate limit rules for %s: %s", key.Domain(), rateResp.GetInfo())
		}
		cloned.RateLimits = rateResp.GetRateLimit()
		cloned.SvcRateLimitRevision = rateResp.GetService().GetRevision()

		circuitResp := x.ruleServer.GetCircuitBreakerWithCache(ctx, serviceReq)
		if circuitResp.GetCode() != api.ExecuteSuccess {
			return nil, fmt.Errorf("select circuit breaker rules for %s: %s", key.Domain(), circuitResp.GetInfo())
		}
		cloned.CircuitBreaker = nil
		if rules := circuitResp.GetCircuitBreaker(); len(rules) > 0 {
			cloned.CircuitBreaker = rules[0]
			cloned.CircuitBreakerRevision = circuitResp.GetService().GetRevision()
			if len(rules) > 1 {
				log.Warn("[XDS][Envoy] multiple circuit breaker rules selected; Envoy cluster supports one outlier policy",
					zap.String("service", key.Domain()), zap.Int("count", len(rules)))
			}
		}

		faultResp := x.ruleServer.GetFaultDetectWithCache(ctx, serviceReq)
		if faultResp.GetCode() != api.ExecuteSuccess {
			return nil, fmt.Errorf("select fault detect rules for %s: %s", key.Domain(), faultResp.GetInfo())
		}
		cloned.FaultDetect = faultResp.GetFaultDetectRules()
		cloned.FaultDetectRevision = faultResp.GetService().GetRevision()
		selected[key] = &cloned
	}
	return selected, nil
}

func governanceContextForNode(ctx context.Context, node *resource.XDSClient) context.Context {
	labels := node.GovernanceLabels()
	labelKeys := make([]string, 0, len(labels))
	for key := range labels {
		labelKeys = append(labelKeys, key)
	}
	sort.Strings(labelKeys)
	clientLabels := make([]*apimodel.ClientLabel, 0, len(labelKeys))
	for _, key := range labelKeys {
		clientLabels = append(clientLabels, &apimodel.ClientLabel{
			Key: key,
			Value: &apimodel.MatchString{
				Type:      apimodel.MatchString_EXACT,
				Value:     labels[key],
				ValueType: apimodel.MatchString_TEXT,
			},
		})
	}
	filter := &apiservice.DiscoverFilter{
		Caller: &apimodel.Caller{Labels: clientLabels},
	}
	return context.WithValue(ctx, apitypes.ContextDiscoverFilter, filter)
}

func (x *XdsResourceGenerator) buildUpdateRequest(req *cache.UpdateResourcesRequest, xdsType resource.XDSType,
	opt *resource.BuildOption, isRemove bool) {

	opt.ForceDelete = isRemove
	xxds, err := x.generateXDSResource(xdsType, opt)
	if err != nil {
		log.Error("[XDS][Envoy] generate xds resource fail", zap.Error(err))
		return
	}

	switch opt.TLSMode {
	case resource.TLSModeNone:
		if opt.ForceDelete {
			req.RemoveNormalNamespaces(opt.Namespace, opt.TLSMode, xdsType, xxds)
		} else {
			req.AddNormalNamespaces(opt.Namespace, xdsType, xxds)
		}
	default:
		if opt.ForceDelete {
			req.RemoveTlsNamespaces(opt.Namespace, opt.TLSMode, xdsType, xxds)
		} else {
			req.AddTlsNamespaces(opt.Namespace, opt.TLSMode, xdsType, xxds)
		}
	}
}

func (x *XdsResourceGenerator) generateXDSResource(xdsType resource.XDSType,
	opt *resource.BuildOption) ([]types.Resource, error) {

	// 需要预埋相关 XDS 资源生成时间开销
	start := time.Now()
	defer func() {
		statis.GetStatis().ReportCallMetrics(metrics.CallMetric{
			Type:     metrics.XDSResourceBuildCallMetric,
			API:      xdsType.String(),
			Protocol: "XDS",
			Times:    1,
			Duration: time.Since(start),
			Labels: map[string]string{
				"service_count": strconv.FormatInt(int64(len(opt.Services)), 10),
				"tls_mode":      string(opt.TLSMode),
			},
		})
	}()

	var (
		xdsBuilder resource.XDSBuilder
	)
	switch xdsType {
	case resource.CDS:
		xdsBuilder = &CDSBuilder{}
	case resource.EDS:
		xdsBuilder = &EDSBuilder{}
	case resource.LDS:
		xdsBuilder = &LDSBuilder{}
	case resource.RDS:
		xdsBuilder = &RDSBuilder{}
	case resource.VHDS:
		xdsBuilder = &VHDSBuilder{}
	default:
		return nil, ErrorNoSupportXDSType
	}

	// 构建 XDS 资源缓存数据
	xdsBuilder.Init(x.namingServer)
	resources, err := xdsBuilder.Generate(opt)
	if err != nil {
		return nil, err
	}
	return resources.([]types.Resource), nil
}
