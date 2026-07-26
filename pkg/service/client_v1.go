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
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"go.uber.org/zap"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/cmdb"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/workloadcredential"
)

// RegisterInstance create one instance
func (s *Server) RegisterInstance(ctx context.Context, req *apiservice.Instance) *apimodel.Response {
	return s.CreateInstance(ctx, req)
}

// DeregisterInstance delete one instance
func (s *Server) DeregisterInstance(ctx context.Context, req *apiservice.Instance) *apimodel.Response {
	return s.DeleteInstance(ctx, req)
}

// ReportServiceContract report client service interface info
func (s *Server) ReportServiceContract(ctx context.Context, req *apiservice.ServiceContract) *apimodel.Response {
	return s.publishServiceContract(ctx, req, apiservice.InterfaceDescriptor_Client)
}

func isSuccessReportContract(rsp *apimodel.Response) bool {
	code := rsp.GetCode()
	if code == uint32(apimodel.Code_ExecuteSuccess) {
		return true
	}
	if code == uint32(apimodel.Code_NoNeedUpdate) {
		return true
	}
	return false
}

// ReportClient 客户端上报信息
func (s *Server) ReportClient(ctx context.Context, req *apiservice.Client) *apimodel.Response {
	// 客户端信息不写入到DB中
	host := req.GetHost()
	// 从CMDB查询地理位置信息
	location, err := cmdb.GetCMDB().GetLocation(host)
	if err != nil {
		log.Errora(utils.RequestID(ctx), zap.Error(err))
	}
	if location != nil {
		req.Location = location.Proto
	}

	// save the client with unique id into store
	if len(req.GetId()) > 0 {
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

	if req.GetNamespace() == "" {
		req.Namespace = DefaultNamespace
	}
	revision, services = s.Cache().Service().ListServices(ctx, req.GetNamespace())
	if revision == "" {
		return resp
	}

	log.Debug("[Service][Discover] list services", zap.Int("size", len(services)),
		zap.String("revision", revision))
	if revision == req.GetRevision() {
		return api.NewDiscoverServiceResponse(apimodel.Code_DataNoChange, req)
	}

	ret := make([]*apiservice.Service, 0, len(services))
	for _, svc := range services {
		ret = append(ret, &apiservice.Service{
			Namespace: svc.Namespace,
			Name:      svc.Name,
			Metadata:  svc.Meta,
		})
	}

	resp.Services = ret
	resp.Service = &apiservice.Service{
		Namespace: req.GetNamespace(),
		Name:      req.GetName(),
		Revision:  revision,
	}

	return resp
}

// GetServiceIdentity returns the service identity bound to the service token in
// the authenticated gRPC metadata. Request fields are only consistency checks;
// they never select the identity being returned.
func (s *Server) GetServiceIdentity(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := api.NewDiscoverServiceIdentityResponse(apimodel.Code_ExecuteSuccess)
	token := utils.ParseAuthToken(ctx)
	if token == "" {
		return api.NewDiscoverServiceIdentityResponse(apimodel.Code_EmptyAutToken)
	}

	svc, err := s.storage.GetOrCreateServiceIdentityByToken(token)
	if err != nil {
		log.Error("[Service][Identity] resolve service token", utils.RequestID(ctx), zap.Error(err))
		return api.NewDiscoverServiceIdentityResponse(storeapi.StoreCode2APICode(err))
	}
	if svc == nil || svc.Identity == nil {
		return api.NewDiscoverServiceIdentityResponse(apimodel.Code_TokenNotExisted)
	}

	descriptorRevision := svc.Identity.Revision
	issuer := workloadcredential.GetServer()
	if issuer != nil {
		descriptorRevision = issuer.DescriptorRevision(descriptorRevision)
	}
	if req != nil {
		if namespace := req.GetNamespace(); namespace != "" && namespace != svc.Namespace {
			return api.NewDiscoverServiceIdentityResponse(apimodel.Code_NotAllowedAccess)
		}
		if name := req.GetName(); name != "" && name != svc.Name {
			return api.NewDiscoverServiceIdentityResponse(apimodel.Code_NotAllowedAccess)
		}
		if req.GetRevision() == descriptorRevision {
			return api.NewDiscoverServiceIdentityResponse(apimodel.Code_DataNoChange)
		}
	}

	resp.ServiceIdentity = &apiservice.ServiceIdentityDescriptor{
		Subject:   svc.Identity.Subject,
		Namespace: svc.Namespace,
		Service:   svc.Name,
		Revision:  descriptorRevision,
	}
	if issuer != nil {
		mode, bundleVersion, trustDomain, audience, endpoint, protocolVersion, formats := issuer.DescriptorFields()
		resp.ServiceIdentity.CredentialMode = mode
		resp.ServiceIdentity.TrustBundleVersion = bundleVersion
		resp.ServiceIdentity.TrustDomain = trustDomain
		resp.ServiceIdentity.Audience = audience
		resp.ServiceIdentity.CredentialEndpoint = endpoint
		resp.ServiceIdentity.IdentityProtocolVersion = protocolVersion
		resp.ServiceIdentity.CredentialFormats = formats
	}
	return resp
}

// ServiceInstancesCache 根据服务名查询服务实例列表
func (s *Server) ServiceInstancesCache(ctx context.Context, filter *apiservice.DiscoverFilter,
	req *apiservice.Service) *apiservice.DiscoverResponse {

	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_INSTANCE)
	svcName := req.GetName()
	nsName := req.GetNamespace()

	// 数据源都来自Cache，这里拿到的service，已经是源服务
	aliasFor := s.getServiceCache(svcName, nsName)
	if aliasFor == nil {
		log.Infof("[Server][Service][Instance] not found name(%s) namespace(%s) service",
			svcName, nsName)
		return api.NewDiscoverInstanceResponse(apimodel.Code_NotFoundResource, req)
	}
	if filter != nil && filter.Caller != nil {
		if filter.Caller.Service != "" && filter.Caller.Namespace != "" {
			s.recordSvcSubscriberGraph(aliasFor, filter)
		}
	}

	revision := s.caches.Service().GetRevisionWorker().GetServiceInstanceRevision(aliasFor.ID)
	revisionHit := revision != "" && revision == req.GetRevision()
	reportDiscoverCacheCall("revision:INSTANCE", revisionHit)
	if revisionHit {
		return api.NewDiscoverInstanceResponse(apimodel.Code_DataNoChange, req)
	}

	onlyHealthy := false
	if filter != nil {
		onlyHealthy = filter.GetOnlyHealthyInstance()
	}

	specSvc := &apiservice.Service{
		Id:        aliasFor.ID,
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	if stoper, ok := s.emptyPushProtectSvs.Load(svcName + "@" + nsName); ok {
		// 如果在保护时间范围内
		if stoper.After(time.Now()) {
			rsp := api.NewDiscoverInstanceResponse(apimodel.Code_DataNoChange, req)
			rsp.Info = "trigger empty push protect"
			return rsp
		}
	}

	var cacheKey string
	if revision != "" {
		cacheKey = discoverInstanceResponseCacheKey(req.GetNamespace(), req.GetName(), revision, onlyHealthy)
		if cachedResp, ok := s.discoverResponseCache.Get(cacheKey); ok {
			reportDiscoverCacheCall("response:INSTANCE", true)
			return cachedResp
		}
		reportDiscoverCacheCall("response:INSTANCE", false)
	}

	finalInstances := make([]*apiservice.Instance, 0, 128)
	matchInsCnt := 0
	s.caches.Instance().DiscoverServiceInstances(specSvc.GetId(), onlyHealthy, func(insData *svctypes.Instance) {
		matchInsCnt++
		// 注意：这里的 value 是 cache 的，不修改 cache 的数据，通过 getInstance，浅拷贝一份数据
		copyIns := s.getInstance(specSvc, insData.Proto)
		finalInstances = append(finalInstances, copyIns)
	})
	// 如果是空实例，则直接跳过，不处理实例列表以及 revision 信息
	if matchInsCnt == 0 {
		// 判断服务是否开启了推空保护，如果开启了，此时添加一个占位
		if dur, ok := aliasFor.ProtectEmptyPush(); ok {
			s.emptyPushProtectSvs.ComputeIfAbsent(svcName+"@"+nsName, func(k string) time.Time {
				eventhub.Publish(eventhub.ServiceEventTopic, &svctypes.ServiceEvent{
					EType:      svctypes.EventServiceOpenEmptyPushProtect,
					Id:         specSvc.GetId(),
					Namespace:  specSvc.GetNamespace(),
					Service:    specSvc.GetName(),
					CreateTime: time.Now(),
				})
				return time.Now().Add(dur)
			})
			rsp := api.NewDiscoverInstanceResponse(apimodel.Code_DataNoChange, req)
			rsp.Info = "trigger empty push protect"
			return rsp
		}
	} else {
		// 如果有实例，则需要清除掉推空保护
		if _, ok := s.emptyPushProtectSvs.Delete(svcName + "@" + nsName); ok {
			eventhub.Publish(eventhub.ServiceEventTopic, &svctypes.ServiceEvent{
				EType:      svctypes.EventServiceCloseEmptyPushProtect,
				Id:         specSvc.GetId(),
				Namespace:  specSvc.GetNamespace(),
				Service:    specSvc.GetName(),
				CreateTime: time.Now(),
			})
		}
	}
	// 填充service数据
	resp.Service = service2Api(aliasFor)
	// 这里需要把服务信息改为用户请求的服务名以及命名空间
	resp.Service.Name = req.GetName()
	resp.Service.Namespace = req.GetNamespace()
	resp.Service.Revision = revision
	// 塞入源服务信息数据
	resp.AliasFor = service2Api(aliasFor)
	// 填充instance数据
	resp.Instances = finalInstances
	if cacheKey != "" {
		s.discoverResponseCache.Put(cacheKey, resp)
	}
	return resp
}

// GetServiceContractWithCache User Client Get ServiceContract Rule Information
func (s *Server) GetServiceContractWithCache(ctx context.Context,
	req *apiservice.ServiceContract) *apimodel.Response {
	resp := api.NewResponse(apimodel.Code_ExecuteSuccess)
	// 服务名和request保持一致
	rspSvc := &apiservice.Service{
		Name:      req.GetService(),
		Namespace: req.GetNamespace(),
	}

	// 获取源服务
	aliasFor := s.findServiceAlias(rspSvc)

	out := s.caches.ServiceContract().Get(ctx, &svctypes.ServiceContract{
		Namespace: aliasFor.Namespace,
		Service:   aliasFor.Name,
		Version:   req.Version,
		Type:      utils.DefaultString(req.GetType(), req.GetName()),
		Protocol:  req.Protocol,
	})
	if out == nil {
		resp.Data = protobuf.MarshalAny(rspSvc)
		resp.Code = uint32(apimodel.Code_NotFoundResource)
		resp.Info = api.Code2Info(uint32(apimodel.Code_NotFoundResource))
		return resp
	}

	// 获取熔断规则数据，并对比revision
	if len(req.GetRevision()) > 0 && req.GetRevision() == out.Revision {
		resp.Code = uint32(apimodel.Code_DataNoChange)
		resp.Info = api.Code2Info(uint32(apimodel.Code_DataNoChange))
		return resp
	}

	rspSvc.Revision = out.Revision
	resp.Data = protobuf.MarshalAny(out.ToSpec())
	return resp
}

// DiscoverServiceContracts returns the normalized contract snapshots for one service.
func (s *Server) DiscoverServiceContracts(
	ctx context.Context, req *apiservice.Service,
) *apiservice.DiscoverResponse {
	resp := api.NewDiscoverResponse(apimodel.Code_ExecuteSuccess)
	resp.Type = apiservice.DiscoverResponse_SERVICE_CONTRACTS
	resp.Service = &apiservice.Service{Name: req.GetName(), Namespace: req.GetNamespace()}

	contracts := s.caches.ServiceContract().List(ctx, req.GetNamespace(), req.GetName())
	if len(contracts) == 0 {
		resp.Code = uint32(apimodel.Code_NotFoundResource)
		resp.Info = api.Code2Info(resp.Code)
		return resp
	}

	revisionInput := strings.Builder{}
	for _, contract := range contracts {
		revisionInput.WriteString(contract.GetCacheKey())
		revisionInput.WriteByte(0)
		revisionInput.WriteString(contract.Revision)
		revisionInput.WriteByte(0)
		resp.ServiceContracts = append(resp.ServiceContracts, contract.ToSpec())
	}
	digest := sha256.Sum256([]byte(revisionInput.String()))
	resp.Service.Revision = hex.EncodeToString(digest[:])
	if req.GetRevision() == resp.Service.Revision {
		resp.Code = uint32(apimodel.Code_DataNoChange)
		resp.Info = api.Code2Info(resp.Code)
		resp.ServiceContracts = nil
	}
	return resp
}

func (s *Server) findServiceAlias(req *apiservice.Service) *svctypes.Service {
	// 获取源服务
	aliasFor := s.getServiceCache(req.GetName(), req.GetNamespace())
	if aliasFor == nil {
		aliasFor = &svctypes.Service{
			Namespace: req.GetNamespace(),
			Name:      req.GetName(),
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
		Code: uint32(apimodel.Code_ExecuteSuccess),
		Info: api.Code2Info(uint32(apimodel.Code_ExecuteSuccess)),
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

func (s *Server) recordSvcSubscriberGraph(req *svctypes.Service, filter *apiservice.DiscoverFilter) {
	s.bc.AsyncRecordServiceSubscriberGraph(&svctypes.ServiceSubscriber{
		Caller: &svctypes.ServiceKey{
			Name:      filter.Caller.Service,
			Namespace: filter.Caller.Namespace,
		},
		Callee: []*svctypes.ServiceKey{
			{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		},
	})
}
