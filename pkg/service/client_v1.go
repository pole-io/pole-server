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

package service

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/cmdb"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// RegisterInstance create one instance
func (s *Server) RegisterInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	return s.CreateInstance(ctx, req)
}

// DeregisterInstance delete one instance
func (s *Server) DeregisterInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	return s.DeleteInstance(ctx, req)
}

// ReportServiceContract report client service interface info
func (s *Server) ReportServiceContract(ctx context.Context, req *apiservice.ServiceContract) *apiservice.Response {
	cacheData := s.caches.ServiceContract().Get(ctx, &svctypes.ServiceContract{
		Namespace: req.GetNamespace(),
		Service:   req.GetService(),
		Type:      req.GetName(),
		Version:   req.GetVersion(),
		Protocol:  req.GetProtocol(),
	})
	// 通过 Cache 模块减少无意义的 CreateServiceContract 逻辑
	if cacheData == nil || cacheData.Content != req.GetContent() {
		rsp := s.CreateServiceContract(ctx, req)
		if !isSuccessReportContract(rsp) {
			return rsp
		}
	}

	rsp := s.CreateServiceContractInterfaces(ctx, req, apiservice.InterfaceDescriptor_Client)
	return rsp
}

func isSuccessReportContract(rsp *apiservice.Response) bool {
	code := rsp.GetCode().GetValue()
	if code == uint32(apimodel.Code_ExecuteSuccess) {
		return true
	}
	if code == uint32(apimodel.Code_NoNeedUpdate) {
		return true
	}
	return false
}

// ReportClient 客户端上报信息
func (s *Server) ReportClient(ctx context.Context, req *apiservice.Client) *apiservice.Response {
	// 客户端信息不写入到DB中
	host := req.GetHost().GetValue()
	// 从CMDB查询地理位置信息
	location, err := cmdb.GetCMDB().GetLocation(host)
	if err != nil {
		log.Errora(utils.RequestID(ctx), zap.Error(err))
	}
	if location != nil {
		req.Location = location.Proto
	}

	// save the client with unique id into store
	if len(req.GetId().GetValue()) > 0 {
		return s.checkAndStoreClient(ctx, req)
	}
	out := &apiservice.Client{
		Host:     req.GetHost(),
		Location: req.Location,
	}
	return api.NewClientResponse(apimodel.Code_ExecuteSuccess, out)
}

// GetServiceWithCache 查询服务列表
func (s *Server) GetServiceWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	if s.caches == nil {
		return api.NewDiscoverServiceResponse(apimodel.Code_ClientAPINotOpen, req)
	}
	if req == nil {
		return api.NewDiscoverServiceResponse(apimodel.Code_EmptyRequest, req)
	}

	resp := api.NewDiscoverServiceResponse(apimodel.Code_ExecuteSuccess, req)
	var (
		revision string
		services []*svctypes.Service
	)

	if req.GetNamespace().GetValue() != "" {
		revision, services = s.Cache().Service().ListServices(ctx, req.GetNamespace().GetValue())
		// 需要加上服务可见性处理
		visibleSvcs := s.caches.Service().GetVisibleServicesInOtherNamespace(ctx, utils.MatchAll, req.GetNamespace().GetValue())
		revisions := make([]string, 0, len(visibleSvcs)+1)
		revisions = append(revisions, revision)
		for i := range visibleSvcs {
			revisions = append(revisions, visibleSvcs[i].Revision)
		}
		services = append(services, visibleSvcs...)
		// 需要重新计算 revison
		if rever, err := cacheapi.CompositeComputeRevision(revisions); err != nil {
			log.Error("[Server][Discover] list services compute multi revision",
				zap.String("namespace", req.GetNamespace().GetValue()), zap.Error(err))
			return api.NewDiscoverInstanceResponse(apimodel.Code_ExecuteException, req)
		} else {
			revision = rever
		}
	} else {
		// 这里拉的是全部服务实例列表，如果客户端可以发起这个请求，应该是不需要
		revision, services = s.Cache().Service().ListAllServices(ctx)
	}
	if revision == "" {
		return resp
	}

	log.Debug("[Service][Discover] list services", zap.Int("size", len(services)),
		zap.String("revision", revision))
	if revision == req.GetRevision().GetValue() {
		return api.NewDiscoverServiceResponse(apimodel.Code_DataNoChange, req)
	}

	ret := make([]*apiservice.Service, 0, len(services))
	for _, svc := range services {
		ret = append(ret, &apiservice.Service{
			Namespace: protobuf.NewStringValue(svc.Namespace),
			Name:      protobuf.NewStringValue(svc.Name),
			Metadata:  svc.Meta,
		})
	}

	resp.Services = ret
	resp.Service = &apiservice.Service{
		Namespace: protobuf.NewStringValue(req.GetNamespace().GetValue()),
		Name:      protobuf.NewStringValue(req.GetName().GetValue()),
		Revision:  protobuf.NewStringValue(revision),
	}

	return resp
}

// ServiceInstancesCache 根据服务名查询服务实例列表
func (s *Server) ServiceInstancesCache(ctx context.Context, filter *apiservice.DiscoverFilter,
	req *apiservice.Service) *apiservice.DiscoverResponse {

	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_INSTANCE)
	svcName := req.GetName().GetValue()
	nsName := req.GetNamespace().GetValue()

	// 数据源都来自Cache，这里拿到的service，已经是源服务
	aliasFor, visibleServices := s.findVisibleServices(ctx, svcName, nsName, req)
	if len(visibleServices) == 0 {
		log.Infof("[Server][Service][Instance] not found name(%s) namespace(%s) service",
			svcName, nsName)
		return api.NewDiscoverInstanceResponse(apimodel.Code_NotFoundResource, req)
	}

	revisions := make([]string, 0, len(visibleServices)+1)
	for _, svc := range visibleServices {
		revision := s.caches.Service().GetRevisionWorker().GetServiceInstanceRevision(svc.ID)
		revisions = append(revisions, revision)
	}
	aggregateRevision, err := cacheapi.CompositeComputeRevision(revisions)
	if err != nil {
		log.Errorf("[Server][Service][Instance] compute multi revision service(%s:%s) err: %s",
			svcName, nsName, err.Error())
		return api.NewDiscoverInstanceResponse(apimodel.Code_ExecuteException, req)
	}
	if aggregateRevision == req.GetRevision().GetValue() {
		return api.NewDiscoverInstanceResponse(apimodel.Code_DataNoChange, req)
	}

	finalInstances := make([]*apiservice.Instance, 0, 128)
	openEmptyProtectCnt := 0
	for _, svc := range visibleServices {
		specSvc := &apiservice.Service{
			Id:        protobuf.NewStringValue(svc.ID),
			Name:      protobuf.NewStringValue(svc.Name),
			Namespace: protobuf.NewStringValue(svc.Namespace),
		}
		if stoper, ok := s.emptyPushProtectSvs.Load(svcName + "@" + nsName); ok {
			// 如果在保护时间范围内
			if stoper.After(time.Now()) {
				openEmptyProtectCnt++
			}
			continue
		}

		matchInsCnt := 0
		s.caches.Instance().DiscoverServiceInstances(specSvc.GetId().GetValue(), filter.GetOnlyHealthyInstance(), func(insData *svctypes.Instance) {
			matchInsCnt++
			// 注意：这里的 value 是 cache 的，不修改 cache 的数据，通过 getInstance，浅拷贝一份数据
			copyIns := s.getInstance(specSvc, insData.Proto)
			finalInstances = append(finalInstances, copyIns)
		})
		// 如果是空实例，则直接跳过，不处理实例列表以及 revision 信息
		if matchInsCnt == 0 {
			// 判断服务是否开启了推空保护，如果开启了，此时添加一个占位
			if dur, ok := svc.ProtectEmptyPush(); ok {
				s.emptyPushProtectSvs.ComputeIfAbsent(svcName+"@"+nsName, func(k string) time.Time {
					eventhub.Publish(eventhub.ServiceEventTopic, &svctypes.ServiceEvent{
						EType:      svctypes.EventServiceOpenEmptyPushProtect,
						Id:         specSvc.GetId().GetValue(),
						Namespace:  specSvc.GetNamespace().GetValue(),
						Service:    specSvc.GetName().GetValue(),
						CreateTime: time.Now(),
					})
					return time.Now().Add(dur)
				})
				openEmptyProtectCnt++
			}
			continue
		} else {
			// 如果有实例，则需要清除掉推空保护
			if _, ok := s.emptyPushProtectSvs.Delete(svcName + "@" + nsName); ok {
				eventhub.Publish(eventhub.ServiceEventTopic, &svctypes.ServiceEvent{
					EType:      svctypes.EventServiceCloseEmptyPushProtect,
					Id:         specSvc.GetId().GetValue(),
					Namespace:  specSvc.GetNamespace().GetValue(),
					Service:    specSvc.GetName().GetValue(),
					CreateTime: time.Now(),
				})
			}
		}
		revision := s.caches.Service().GetRevisionWorker().GetServiceInstanceRevision(svc.ID)
		revisions = append(revisions, revision)
	}

	// 所有服务都触发了推空保护
	if openEmptyProtectCnt == len(visibleServices) {
		// 当前存在推空保护，返回给客户端 DataNoChange 变化
		rsp := api.NewDiscoverInstanceResponse(apimodel.Code_DataNoChange, req)
		rsp.Info = protobuf.NewStringValue("trigger empty push protect")
		return rsp
	}

	if aliasFor == nil {
		// 这里只会出现，查询的目标服务和命名空间不存在，但是可见性的服务存在
		// 所以这里需要用入口的服务名和命名空间填充服务数据结构，以便返回最终的应答服务名和命名空间
		aliasFor = &svctypes.Service{Name: svcName, Namespace: nsName}
	}
	// 填充service数据
	resp.Service = service2Api(aliasFor)
	// 这里需要把服务信息改为用户请求的服务名以及命名空间
	resp.Service.Name = req.GetName()
	resp.Service.Namespace = req.GetNamespace()
	resp.Service.Revision = protobuf.NewStringValue(aggregateRevision)
	// 塞入源服务信息数据
	resp.AliasFor = service2Api(aliasFor)
	// 填充instance数据
	resp.Instances = finalInstances
	return resp
}

func (s *Server) findVisibleServices(ctx context.Context, svcName, nsName string,
	req *apiservice.Service) (*svctypes.Service, []*svctypes.Service) {
	visibleServices := make([]*svctypes.Service, 0, 4)
	// 数据源都来自Cache，这里拿到的service，已经是源服务
	aliasFor := s.getServiceCache(svcName, nsName)
	if aliasFor != nil {
		// 获取到实际的服务，则将查询的服务名替换成实际的服务名和命名空间
		svcName = aliasFor.Name
		nsName = aliasFor.Namespace
		// 先把自己放进去
		visibleServices = append(visibleServices, aliasFor)
	}
	ret := s.caches.Service().GetVisibleServicesInOtherNamespace(ctx, svcName, nsName)
	if len(ret) > 0 {
		visibleServices = append(visibleServices, ret...)
	}
	return aliasFor, visibleServices
}

// GetServiceContractWithCache User Client Get ServiceContract Rule Information
func (s *Server) GetServiceContractWithCache(ctx context.Context,
	req *apiservice.ServiceContract) *apiservice.Response {
	resp := api.NewResponse(apimodel.Code_ExecuteSuccess)
	// 服务名和request保持一致
	resp.Service = &apiservice.Service{
		Name:      wrapperspb.String(req.GetService()),
		Namespace: wrapperspb.String(req.GetNamespace()),
	}

	// 获取源服务
	aliasFor := s.findServiceAlias(resp.Service)

	out := s.caches.ServiceContract().Get(ctx, &svctypes.ServiceContract{
		Namespace: aliasFor.Namespace,
		Service:   aliasFor.Name,
		Version:   req.Version,
		Type:      req.Name,
		Protocol:  req.Protocol,
	})
	if out == nil {
		resp.Code = wrapperspb.UInt32(uint32(apimodel.Code_NotFoundResource))
		resp.Info = wrapperspb.String(api.Code2Info(uint32(apimodel.Code_NotFoundResource)))
		return resp
	}

	// 获取熔断规则数据，并对比revision
	if len(req.GetRevision()) > 0 && req.GetRevision() == out.Revision {
		resp.Code = wrapperspb.UInt32(uint32(apimodel.Code_DataNoChange))
		resp.Info = wrapperspb.String(api.Code2Info(uint32(apimodel.Code_DataNoChange)))
		return resp
	}

	resp.Service.Revision = wrapperspb.String(out.Revision)
	resp.ServiceContract = out.ToSpec()
	return resp
}

func (s *Server) findServiceAlias(req *apiservice.Service) *svctypes.Service {
	// 获取源服务
	aliasFor := s.getServiceCache(req.GetName().GetValue(), req.GetNamespace().GetValue())
	if aliasFor == nil {
		aliasFor = &svctypes.Service{
			Namespace: req.GetNamespace().GetValue(),
			Name:      req.GetName().GetValue(),
		}
	}
	return aliasFor
}

func CreateCommonDiscoverResponse(req *apiservice.Service,
	dT apiservice.DiscoverResponse_DiscoverResponseType) *apiservice.DiscoverResponse {
	return createCommonDiscoverResponse(req, dT)
}

func createCommonDiscoverResponse(req *apiservice.Service,
	dT apiservice.DiscoverResponse_DiscoverResponseType) *apiservice.DiscoverResponse {
	return &apiservice.DiscoverResponse{
		Code: &wrappers.UInt32Value{Value: uint32(apimodel.Code_ExecuteSuccess)},
		Info: &wrappers.StringValue{Value: api.Code2Info(uint32(apimodel.Code_ExecuteSuccess))},
		Type: dT,
		Service: &apiservice.Service{
			Name:      req.GetName(),
			Namespace: req.GetNamespace(),
		},
	}
}

// getServiceCache 根据服务名获取服务缓存数据, 注意，如果是服务别名查询，这里会返回别名的源服务，不会返回别名
func (s *Server) getServiceCache(name string, namespace string) *svctypes.Service {
	sc := s.caches.Service()
	service := sc.GetServiceByName(name, namespace)
	if service == nil {
		return nil
	}
	// 如果是服务别名，继续查找一下
	if service.IsAlias() {
		service = sc.GetServiceByID(service.Reference)
		if service == nil {
			return nil
		}
	}

	if service.Meta == nil {
		service.Meta = make(map[string]string)
	}
	return service
}
