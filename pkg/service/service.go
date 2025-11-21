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
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"go.uber.org/zap"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

const (
	MetadataInternalAutoCreated string = "internal-auto-created"
)

// Service2Api *svctypes.Service转换为*api.service
type Service2Api func(service *svctypes.Service) *apiservice.Service

var (
	serviceFilter           = 1 // 过滤服务的
	instanceFilter          = 2 // 过滤实例的
	serviceMetaFilter       = 3 // 过滤service Metadata的
	instanceMetaFilter      = 4 // 过滤instance Metadata的
	ServiceFilterAttributes = map[string]int{
		"id":          serviceFilter,
		"name":        serviceFilter,
		"namespace":   serviceFilter,
		"business":    serviceFilter,
		"department":  serviceFilter,
		"cmdb_mod1":   serviceFilter,
		"cmdb_mod2":   serviceFilter,
		"cmdb_mod3":   serviceFilter,
		"owner":       serviceFilter,
		"offset":      serviceFilter,
		"limit":       serviceFilter,
		"platform_id": serviceFilter,
		// 只返回存在健康实例的服务列表
		"only_exist_health_instance": serviceFilter,
		"host":                       instanceFilter,
		"port":                       instanceFilter,
		"keys":                       serviceMetaFilter,
		"values":                     serviceMetaFilter,
		"instance_keys":              instanceMetaFilter,
		"instance_values":            instanceMetaFilter,
	}
)

// CreateServices 批量创建服务
func (s *Server) CreateServices(ctx context.Context, req []*apiservice.Service) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, service := range req {
		response := s.CreateService(ctx, service)
		api.Collect(responses, response)
	}

	return api.FormatBatchWriteResponse(responses)
}

// CreateService 创建单个服务
func (s *Server) CreateService(ctx context.Context, req *apiservice.Service) *apimodel.Response {
	if _, errResp := s.createNamespaceIfAbsent(ctx, req); errResp != nil {
		return errResp
	}

	namespaceName := req.GetNamespace()
	serviceName := req.GetName()

	// 检查命名空间是否存在
	namespace, err := s.storage.GetNamespace(namespaceName)
	if err != nil {
		log.Error("[Service] get namespace fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewServiceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if namespace == nil {
		return api.NewServiceResponse(apimodel.Code_NotFoundResource, req)
	}

	// 检查是否存在
	service, err := s.storage.GetService(serviceName, namespaceName)
	if err != nil {
		log.Error("[Service] get service fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewServiceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if service != nil {
		req.Id = string(service.ID)
		return api.NewServiceResponse(apimodel.Code_ExistedResource, req)
	}

	// 存储层操作
	data := s.createServiceModel(req)
	if err := s.storage.AddService(data); err != nil {
		log.Error("[Service] save service fail", utils.RequestID(ctx), zap.Error(err))
		// 如果在存储层发现资源存在错误，则需要再一次从存储层获取响应的信息，填充响应的 svc_id 信息
		if storeapi.StoreCode2APICode(err) == apimodel.Code_ExistedResource {
			// 检查是否存在
			service, err := s.storage.GetService(serviceName, namespaceName)
			if err != nil {
				log.Error("[Service] get service fail", utils.RequestID(ctx), zap.Error(err))
				return api.NewServiceResponse(storeapi.StoreCode2APICode(err), req)
			}
			if service != nil {
				req.Id = string(service.ID)
				return api.NewServiceResponse(apimodel.Code_ExistedResource, req)
			}
		}
		return wrapperServiceStoreResponse(req, err)
	}

	log.Info(fmt.Sprintf("create service: namespace=%v, name=%v, meta=%+v",
		namespaceName, serviceName, req.GetMetadata()), utils.RequestID(ctx))
	s.RecordHistory(ctx, serviceRecordEntry(ctx, req, data, types.OCreate))

	out := &apiservice.Service{
		Id:        string(data.ID),
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
		Token:     string(data.Token),
	}
	return api.NewServiceResponse(apimodel.Code_ExecuteSuccess, out)
}

// DeleteServices 批量删除服务
func (s *Server) DeleteServices(ctx context.Context, req []*apiservice.Service) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, service := range req {
		response := s.DeleteService(ctx, service)
		api.Collect(responses, response)
	}

	return api.FormatBatchWriteResponse(responses)
}

// DeleteService 删除单个服务
//
//	删除操作需要对服务进行加锁操作，
//	防止有与服务关联的实例或者配置有新增的操作
func (s *Server) DeleteService(ctx context.Context, req *apiservice.Service) *apimodel.Response {
	namespaceName := req.GetNamespace()
	serviceName := req.GetName()

	// 检查是否存在
	service, err := s.storage.GetService(serviceName, namespaceName)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if service == nil {
		return api.NewServiceResponse(apimodel.Code_ExecuteSuccess, req)
	}

	// 判断service下的资源是否已经全部被删除
	if resp := s.isServiceExistedResource(ctx, service); resp != nil {
		return resp
	}

	if err := s.storage.DeleteService(service.ID, serviceName, namespaceName); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperServiceStoreResponse(req, err)
	}

	msg := fmt.Sprintf("delete service: namespace=%v, name=%v", namespaceName, serviceName)
	log.Info(msg, utils.RequestID(ctx))
	s.RecordHistory(ctx, serviceRecordEntry(ctx, req, nil, types.ODelete))
	return api.NewServiceResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateServices 批量修改服务
func (s *Server) UpdateServices(ctx context.Context, req []*apiservice.Service) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, service := range req {
		response := s.UpdateService(ctx, service)
		api.Collect(responses, response)
	}

	return api.FormatBatchWriteResponse(responses)
}

// UpdateService 修改单个服务
func (s *Server) UpdateService(ctx context.Context, req *apiservice.Service) *apimodel.Response {
	// 鉴权
	service, _, resp := s.checkServiceAuthority(ctx, req)
	if resp != nil {
		return resp
	}

	// [2020.02.18]If service is alias, not allowed to modify
	if service.IsAlias() {
		return api.NewServiceResponse(apimodel.Code_NotAllowedAccess, req)
	}

	log.Info(fmt.Sprintf("old service: %+v", service), utils.RequestID(ctx))

	// 修改
	err, needUpdate, needUpdateOwner := s.updateServiceAttribute(req, service)
	if err != nil {
		return err
	}
	// 判断是否需要更新
	if !needUpdate {
		log.Info("update service data no change, no need update",
			utils.RequestID(ctx), zap.String("service", req.String()))
		return api.NewServiceResponse(apimodel.Code_NoNeedUpdate, req)
	}

	// 存储层操作
	if err := s.storage.UpdateService(service, needUpdateOwner); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperServiceStoreResponse(req, err)
	}

	msg := fmt.Sprintf("update service: namespace=%v, name=%v", service.Namespace, service.Name)
	log.Info(msg, utils.RequestID(ctx))
	s.RecordHistory(ctx, serviceRecordEntry(ctx, req, service, types.OUpdate))
	return api.NewServiceResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateServiceToken 更新服务token
func (s *Server) UpdateServiceToken(ctx context.Context, req *apiservice.Service) *apimodel.Response {
	// 鉴权
	service, _, resp := s.checkServiceAuthority(ctx, req)
	if resp != nil {
		return resp
	}
	if service.IsAlias() {
		return api.NewServiceResponse(apimodel.Code_NotAllowedAccess, req)
	}

	// 生成一个新的token和revision
	service.Token = utils.NewUUID()
	service.Revision = utils.NewUUID()
	// 更新数据库
	if err := s.storage.UpdateServiceToken(service.ID, service.Token, service.Revision); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperServiceStoreResponse(req, err)
	}
	log.Info("update service token", zap.String("namespace", service.Namespace),
		zap.String("name", service.Name), zap.String("service-id", service.ID),
		utils.RequestID(ctx))
	s.RecordHistory(ctx, serviceRecordEntry(ctx, req, service, types.OUpdateToken))

	// 填充新的token返回
	out := &apiservice.Service{
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
		Token:     string(service.Token),
	}
	return api.NewServiceResponse(apimodel.Code_ExecuteSuccess, out)
}

// GetAllServices query all service list by namespace
func (s *Server) GetAllServices(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	var (
		svcs []*svctypes.Service
	)

	if ns, ok := query["namespace"]; ok && len(ns) > 0 {
		_, svcs = s.Cache().Service().ListServices(ctx, ns)
	} else {
		_, svcs = s.Cache().Service().ListAllServices(ctx)
	}

	ret := make([]*apiservice.Service, 0, len(svcs))
	for i := range svcs {
		count := s.Cache().Instance().GetInstancesCountByServiceID(svcs[i].ID)
		ret = append(ret, &apiservice.Service{
			Namespace:            string(svcs[i].Namespace),
			Name:                 string(svcs[i].Name),
			TotalInstanceCount:   uint32(count.TotalInstanceCount),
			HealthyInstanceCount: uint32(count.HealthyInstanceCount),
			Metadata:             svcs[i].Meta,
		})
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = uint32(len(ret))
	resp.Size = uint32(len(ret))
	// 注释：响应结构改动 - 根据 pole-io/specification，BatchQueryResponse 不再有 Services 字段
	// 根据 pole-io/specification，BatchQueryResponse 不再有 Services 字段
	// 数据需要通过 data 字段传递
	// TODO: 需要确定正确的序列化方式
	return resp
}

// GetServices 查询服务 注意：不包括别名
func (s *Server) GetServices(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	serviceFilters := make(map[string]string)
	instanceFilters := make(map[string]string)
	var (
		metaKeys, metaValues                   string
		inputInstMetaKeys, inputInstMetaValues string
	)
	for key, value := range query {
		switch typ := ServiceFilterAttributes[key]; typ {
		case serviceFilter:
			serviceFilters[key] = value
		case serviceMetaFilter:
			if key == "keys" {
				metaKeys = value
			} else {
				metaValues = value
			}
		case instanceMetaFilter:
			if key == "instance_keys" {
				inputInstMetaKeys = value
			} else {
				inputInstMetaValues = value
			}
		default:
			instanceFilters[key] = value
		}
	}

	instanceMetas := make(map[string]string)
	if inputInstMetaKeys != "" {
		instMetaKeys := strings.Split(inputInstMetaKeys, ",")
		instMetaValues := strings.Split(inputInstMetaValues, ",")
		for idx, key := range instMetaKeys {
			instanceMetas[key] = instMetaValues[idx]
		}
	}

	instanceArgs, err := ParseInstanceArgs(instanceFilters, instanceMetas)
	if err != nil {
		log.Errorf("[Server][Service][Query] instance args error: %s", err.Error())
		return api.NewBatchQueryResponseWithMsg(apimodel.Code_InvalidParameter, err.Error())
	}

	// 解析metaKeys，metaValues
	serviceMetas := make(map[string]string)
	if metaKeys != "" {
		serviceMetas[metaKeys] = metaValues
	}

	// 判断offset和limit是否为int，并从filters清除offset/limit参数
	offset, limit, _ := valid.ParseOffsetAndLimit(serviceFilters)

	serviceArgs := parseServiceArgs(serviceFilters, serviceMetas, ctx)
	total, services, err := s.caches.Service().GetServicesByFilter(ctx, serviceArgs, instanceArgs, offset, limit)
	if err != nil {
		log.Errorf("[Server][Service][Query] req(%+v) store err: %s", query, err.Error())
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = uint32(total)
	resp.Size = uint32(len(services))
	// 根据 pole-io/specification，BatchQueryResponse 不再有 Services 字段
	// 数据需要通过 data 字段传递
	// TODO: 需要确定正确的序列化方式
	return resp
}

// parseServiceArgs 解析服务的查询条件
func parseServiceArgs(filter map[string]string, metaFilter map[string]string,
	ctx context.Context) *cacheapi.ServiceArgs {

	res := &cacheapi.ServiceArgs{
		Filter:    filter,
		Metadata:  metaFilter,
		Namespace: filter["namespace"],
	}
	var ok bool
	if res.Name, ok = filter["name"]; ok && matchs.IsPrefixWildName(res.Name) {
		log.Infof("[Server][Service][Query] fuzzy search with name %s", res.Name)
		res.WildName = true
	}
	if matchs.IsWildName(res.Namespace) {
		log.Infof("[Server][Service][Query] fuzzy search with namespace %s", res.Namespace)
		res.WildNamespace = true
	}
	if business, ok := filter["business"]; ok {
		log.Infof("[Server][Service][Query] fuzzy search with business %s, operator %s",
			business, utils.ParseOperator(ctx))
		res.WildBusiness = true
	}
	if val, ok := filter["only_exist_health_instance"]; ok {
		res.OnlyExistHealthInstance = val == "true"
	}
	if val, ok := filter["only_exist_instance"]; ok {
		res.OnlyExistInstance = val == "true"
	}
	// 如果元数据条件是空的话，判断是否是空条件匹配
	if len(metaFilter) == 0 {
		// 如果没有匹配条件，那么就是空条件匹配
		if len(filter) == 0 {
			res.EmptyCondition = true
		}
		// 只有一个命名空间条件，也是在这个命名空间下面的空条件匹配
		if len(filter) == 1 && res.Namespace != "" && !res.WildNamespace {
			res.EmptyCondition = true
		}
	}
	return res
}

// GetServicesCount 查询服务总数
func (s *Server) GetServicesCount(ctx context.Context) *apimodel.BatchQueryResponse {
	count, err := s.storage.GetServicesCount()
	if err != nil {
		log.Errorf("[Server][Service][Count] get service count storage err: %s", err.Error())
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = uint32(count)
	// 根据 pole-io/specification，BatchQueryResponse 不再有 Services 字段
	// 数据需要通过 data 字段传递
	// TODO: 需要确定正确的序列化方式
	return out
}

// GetServiceToken 查询Service的token
func (s *Server) GetServiceToken(ctx context.Context, req *apiservice.Service) *apimodel.Response {
	// 鉴权
	_, _, resp := s.checkServiceAuthority(ctx, req)
	if resp != nil {
		return resp
	}

	// s.RecordHistory(serviceRecordEntry(ctx, req, model.OGetToken))
	out := api.NewResponse(apimodel.Code_ExecuteSuccess)
	// 注释：Token响应改动 - 根据 pole-io/specification，Response 只有 data 字段，需要将服务数据序列化到 data 中
	// 根据 pole-io/specification，Response 只有 data 字段，需要将服务数据序列化到 data 中
	// TODO: 需要确定正确的序列化方式
	return out
}

// createNamespaceIfAbsent Automatically create namespaces
func (s *Server) createNamespaceIfAbsent(ctx context.Context, svc *apiservice.Service) (string, *apimodel.Response) {
	val, rsp := s.Namespace().CreateNamespaceIfAbsent(ctx, &apimodel.Namespace{
		Name:   string(svc.GetNamespace()),
		Owners: svc.Owners,
	})
	if !api.IsSuccess(rsp) {
		return "", rsp
	}
	return val, nil
}

// createServiceModel 创建存储层服务模型
func (s *Server) createServiceModel(req *apiservice.Service) *svctypes.Service {
	// 注释：创建服务模型改动 - API规范变更导致字段类型变化，从wrapper类型改为基础类型
	return &svctypes.Service{
		ID:         utils.NewUUID(),
		Name:       req.GetName(),
		Namespace:  req.GetNamespace(),
		Meta:       req.GetMetadata(),
		Ports:      req.GetPorts(),
		Business:   req.GetBusiness(),
		Department: req.GetDepartment(),
		CmdbMod1:   req.GetCmdbMod1(),
		CmdbMod2:   req.GetCmdbMod2(),
		CmdbMod3:   req.GetCmdbMod3(),
		Comment:    req.GetComment(),
		Owner:      req.GetOwners(),
		// 移除 PlatformID 字段，因为 pole-io/specification 中已不存在
		Token:    utils.NewUUID(),
		Revision: utils.NewUUID(),
		ExportTo: types.ExportToMap(req.GetExportTo()),
	}
}

// updateServiceAttribute 修改服务属性
func (s *Server) updateServiceAttribute(
	req *apiservice.Service, service *svctypes.Service) (*apimodel.Response, bool, bool) {
	var (
		needUpdate      = false
		needNewRevision = false
		needUpdateOwner = false
	)

	if req.GetMetadata() != nil {
		if need := serviceMetaNeedUpdate(req, service); need {
			needUpdate = need
			needNewRevision = true
			service.Meta = req.GetMetadata()
		}
	}
	if !needUpdate {
		// 不需要更新metadata
		service.Meta = nil
	}
	// 注释：ExportTo字段处理改动 - 现在是[]string类型而非[]*wrapperspb.StringValue
	// 处理 ExportTo 字段（现在是 []string 类型）
	exportToMap := types.ExportToMap(req.ExportTo)
	if eq, newVal := isEqualServiceExport(exportToMap, service.ExportTo); !eq {
		needUpdate = true
		service.ExportTo = newVal
	}

	if req.GetPorts() != "" && req.GetPorts() != service.Ports {
		service.Ports = req.GetPorts()
		needUpdate = true
	}

	if req.GetBusiness() != "" && req.GetBusiness() != service.Business {
		service.Business = req.GetBusiness()
		needUpdate = true
	}

	if req.GetDepartment() != "" && req.GetDepartment() != service.Department {
		service.Department = req.GetDepartment()
		needUpdate = true
	}

	if req.GetCmdbMod1() != "" && req.GetCmdbMod1() != service.CmdbMod1 {
		service.CmdbMod1 = req.GetCmdbMod1()
		needUpdate = true
	}
	if req.GetCmdbMod2() != "" && req.GetCmdbMod2() != service.CmdbMod2 {
		service.CmdbMod2 = req.GetCmdbMod2()
		needUpdate = true
	}
	if req.GetCmdbMod3() != "" && req.GetCmdbMod3() != service.CmdbMod3 {
		service.CmdbMod3 = req.GetCmdbMod3()
		needUpdate = true
	}

	if req.GetComment() != "" && req.GetComment() != service.Comment {
		service.Comment = req.GetComment()
		needUpdate = true
	}

	if req.GetOwners() != "" && req.GetOwners() != service.Owner {
		service.Owner = req.GetOwners()
		needUpdate = true
		needUpdateOwner = true
	}

	// 注释：PlatformId字段移除 - 移除 PlatformId 相关代码，因为 pole-io/specification 中已不存在此字段
	// 移除 PlatformId 相关代码，因为 pole-io/specification 中已不存在此字段

	if needNewRevision {
		service.Revision = utils.NewUUID()
	}

	return nil, needUpdate, needUpdateOwner
}

func isEqualServiceExport(reqMap map[string]struct{}, save map[string]struct{}) (bool, map[string]struct{}) {
	if len(reqMap) != len(save) {
		return false, reqMap
	}
	for k := range reqMap {
		if _, ok := save[k]; !ok {
			return false, reqMap
		}
	}

	return true, map[string]struct{}{}
}

// getServiceAliasCountWithService 获取服务下别名的总数
func (s *Server) getServiceAliasCountWithService(name string, namespace string) (uint32, error) {
	filter := map[string]string{
		"service":   name,
		"namespace": namespace,
	}
	total, _, err := s.storage.GetServiceAliases(filter, 0, 1)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// getInstancesCountWithService 获取服务下实例的总数
func (s *Server) getInstancesCountWithService(name string, namespace string) (uint32, error) {
	filter := map[string]string{
		"name":      name,
		"namespace": namespace,
	}
	total, _, err := s.storage.GetExpandInstances(filter, nil, 0, 1)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// getRoutingCountWithService 获取服务下路由配置总数
func (s *Server) getRoutingCountWithService(id string) (uint32, error) {
	routing, err := s.storage.GetRoutingConfigWithID(id)
	if err != nil {
		return 0, err
	}

	if routing == nil {
		return 0, nil
	}
	return 1, nil
}

// isServiceExistedResource 检查服务下的资源存在情况，在删除服务的时候需要用到
func (s *Server) isServiceExistedResource(ctx context.Context, service *svctypes.Service) *apimodel.Response {
	// 服务别名，不需要判断
	if service.IsAlias() {
		return nil
	}
	out := &apiservice.Service{
		Name:      string(service.Name),
		Namespace: string(service.Namespace),
	}
	total, err := s.getInstancesCountWithService(service.Name, service.Namespace)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceResponse(storeapi.StoreCode2APICode(err), out)
	}
	if total != 0 {
		return api.NewServiceResponse(apimodel.Code_ServiceExistedInstances, out)
	}

	total, err = s.getServiceAliasCountWithService(service.Name, service.Namespace)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceResponse(storeapi.StoreCode2APICode(err), out)
	}
	if total != 0 {
		return api.NewServiceResponse(apimodel.Code_ServiceExistedAlias, out)
	}

	// TODO will remove until have sync router rule v1 to v2
	total, err = s.getRoutingCountWithService(service.ID)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceResponse(storeapi.StoreCode2APICode(err), out)
	}

	if total != 0 {
		return api.NewServiceResponse(apimodel.Code_ServiceExistedRoutings, out)
	}
	return nil
}

// checkServiceAuthority 对服务进行鉴权，并且返回svctypes.Service
// return service, token, response
func (s *Server) checkServiceAuthority(ctx context.Context, req *apiservice.Service) (*svctypes.Service,
	string, *apimodel.Response) {
	namespaceName := req.GetNamespace()
	serviceName := req.GetName()

	// 检查是否存在
	svc, err := s.storage.GetService(serviceName, namespaceName)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return nil, "", api.NewServiceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if svc == nil {
		return nil, "", api.NewServiceResponse(apimodel.Code_NotFoundResource, req)
	}
	if svc.Reference != "" {
		svc, err = s.storage.GetServiceByID(svc.Reference)
		if err != nil {
			log.Error(err.Error(), utils.RequestID(ctx))
			return nil, "", api.NewServiceResponse(storeapi.StoreCode2APICode(err), req)
		}
		if svc == nil {
			return nil, "", api.NewServiceResponse(apimodel.Code_NotFoundResource, req)
		}
	}

	expectToken := svc.Token

	return svc, expectToken, nil
}

// service2Api svctypes.Service 转为 api.Service
func service2Api(data *svctypes.Service) *apiservice.Service {
	if data == nil {
		return nil
	}
	out := data.ToSpec()
	return out
}

// serviceOwner2Api svctypes.Service转为api.Service
// 只转name+namespace+owner
func serviceOwner2Api(service *svctypes.Service) *apiservice.Service {
	if service == nil {
		return nil
	}
	out := &apiservice.Service{
		Name:      string(service.Name),
		Namespace: string(service.Namespace),
		Owners:    string(service.Owner),
	}
	return out
}

// services2Api service数组转为[]*api.Service
func services2Api(services []*svctypes.Service, handler Service2Api) []*apiservice.Service {
	out := make([]*apiservice.Service, 0, len(services))
	for _, entry := range services {
		out = append(out, handler(entry))
	}

	return out
}

// enhancedServices2Api service数组转为[]*api.Service
func enhancedServices2Api(services []*svctypes.EnhancedService, handler Service2Api) []*apiservice.Service {
	out := make([]*apiservice.Service, 0, len(services))
	for _, entry := range services {
		outSvc := handler(entry.Service)
		outSvc.HealthyInstanceCount = uint32(entry.HealthyInstanceCount)
		outSvc.TotalInstanceCount = uint32(entry.TotalInstanceCount)
		out = append(out, outSvc)
	}

	return out
}

// apis2ServicesName api数组转为[]*svctypes.Service
func apis2ServicesName(reqs []*apiservice.Service) []*svctypes.Service {
	if reqs == nil {
		return nil
	}

	out := make([]*svctypes.Service, 0, len(reqs))
	for _, req := range reqs {
		out = append(out, api2ServiceName(req))
	}
	return out
}

// api2ServiceName api转为*svctypes.Service
func api2ServiceName(req *apiservice.Service) *svctypes.Service {
	if req == nil {
		return nil
	}
	service := &svctypes.Service{
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
	}
	return service
}

// serviceMetaNeedUpdate 检查服务metadata是否需要更新
func serviceMetaNeedUpdate(req *apiservice.Service, service *svctypes.Service) bool {
	// 收到的请求的metadata为空，则代表metadata不需要更新
	if req.GetMetadata() == nil {
		return false
	}

	// metadata个数不一致，肯定需要更新
	if len(req.GetMetadata()) != len(service.Meta) {
		return true
	}

	needUpdate := false
	// 新数据为标准，对比老数据，发现不一致，则需要更新
	for key, value := range req.GetMetadata() {
		oldValue, ok := service.Meta[key]
		if !ok {
			needUpdate = true
			break
		}
		if value != oldValue {
			needUpdate = true
			break
		}
	}
	if needUpdate {
		return true
	}

	// 老数据作为标准，对比新数据，发现不一致，则需要更新
	for key, value := range service.Meta {
		newValue, ok := req.Metadata[key]
		if !ok {
			needUpdate = true
			break
		}
		if value != newValue {
			needUpdate = true
			break
		}
	}

	return needUpdate
}

// wrapperServiceStoreResponse wrapper service error
func wrapperServiceStoreResponse(service *apiservice.Service, err error) *apimodel.Response {
	if err == nil {
		return nil
	}
	resp := api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	// 根据 pole-io/specification，Response 只有 data 字段，需要将服务数据序列化到 data 中
	// TODO: 需要确定正确的序列化方式
	_ = service // 暂时忽略 service 参数
	return resp
}

// parseRequestToken 从request中获取服务token
func parseRequestToken(ctx context.Context, value string) string {
	if value != "" {
		return value
	}

	return utils.ParseToken(ctx)
}

// serviceRecordEntry 生成服务的记录entry
func serviceRecordEntry(ctx context.Context, req *apiservice.Service, md *svctypes.Service,
	operationType types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RService,
		ResourceName:  req.GetName(),
		Namespace:     req.GetNamespace(),
		OperationType: operationType,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}

	return entry
}
