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

package namespace

import (
	"context"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	commontime "github.com/pole-io/pole-server/pkg/common/time"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
)

var _ NamespaceOperateServer = (*Server)(nil)

func (s *Server) allowAutoCreate() bool {
	return s.cfg.AutoCreate
}

func AllowAutoCreate(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, utils.ContextKeyAutoCreateNamespace{}, true)
	return ctx
}

// CreateNamespaces 批量创建命名空间
func (s *Server) CreateNamespaces(ctx context.Context, req []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	if checkError := checkBatchNamespace(req); checkError != nil {
		return checkError
	}

	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, namespace := range req {
		response := s.CreateNamespace(ctx, namespace)
		api.Collect(responses, response)
	}

	return responses
}

// CreateNamespaceIfAbsent 创建命名空间，如果不存在
func (s *Server) CreateNamespaceIfAbsent(ctx context.Context, req *apimodel.Namespace) (string, *apiservice.Response) {
	if resp := checkCreateNamespace(req); resp != nil {
		return "", resp
	}
	name := req.GetName().GetValue()
	val, err := s.loadNamespace(name)
	if err != nil {
		return name, nil
	}
	if val == "" && !s.allowAutoCreate() {
		ctxVal := ctx.Value(utils.ContextKeyAutoCreateNamespace{})
		if ctxVal == nil || ctxVal.(bool) != true {
			return "", api.NewResponse(apimodel.Code_NotFoundNamespace)
		}
	}
	ret, err, _ := s.createNamespaceSingle.Do(name, func() (interface{}, error) {
		return s.CreateNamespace(ctx, req), nil
	})
	if err != nil {
		return "", api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}
	var (
		resp = ret.(*apiservice.Response)
		code = resp.GetCode().GetValue()
	)
	if code == uint32(apimodel.Code_ExecuteSuccess) || code == uint32(apimodel.Code_ExistedResource) {
		return name, api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, req)
	}
	return "", resp
}

// CreateNamespace 创建单个命名空间
func (s *Server) CreateNamespace(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	// 参数检查
	if checkError := checkCreateNamespace(req); checkError != nil {
		return checkError
	}

	namespaceName := req.GetName().GetValue()

	// 检查是否存在
	namespace, err := s.storage.GetNamespace(namespaceName)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if namespace != nil {
		return api.NewNamespaceResponse(apimodel.Code_ExistedResource, req)
	}

	data := s.createNamespaceModel(req)

	// 存储层操作
	if err := s.storage.AddNamespace(data); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}

	log.Info("create namespace", utils.RequestID(ctx), zap.String("name", namespaceName))
	out := &apimodel.Namespace{
		Name: req.GetName(),
	}
	s.RecordHistory(namespaceRecordEntry(ctx, req, types.OCreate))
	return api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, out)
}

/**
 * @brief 创建存储层命名空间模型
 */
func (s *Server) createNamespaceModel(req *apimodel.Namespace) *types.Namespace {
	namespace := &types.Namespace{
		Name:            req.GetName().GetValue(),
		Comment:         req.GetComment().GetValue(),
		Owner:           req.GetOwners().GetValue(),
		Token:           utils.NewUUID(),
		ServiceExportTo: types.ExportToMap(req.GetServiceExportTo()),
		Metadata:        req.GetMetadata(),
	}
	return namespace
}

// DeleteNamespaces 批量删除命名空间
func (s *Server) DeleteNamespaces(ctx context.Context, req []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	if checkError := checkBatchNamespace(req); checkError != nil {
		return checkError
	}

	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, namespace := range req {
		response := s.DeleteNamespace(ctx, namespace)
		api.Collect(responses, response)
	}

	return responses
}

// DeleteNamespace 删除单个命名空间
func (s *Server) DeleteNamespace(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	// 参数检查
	if checkError := checkReviseNamespace(ctx, req); checkError != nil {
		return checkError
	}

	tx, err := s.storage.CreateTransaction()
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}
	defer func() { _ = tx.Commit() }()

	// 检查是否存在
	namespace, err := tx.LockNamespace(req.GetName().GetValue())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if namespace == nil {
		return api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, req)
	}

	// 判断属于该命名空间的服务是否都已经被删除
	total, err := s.getServicesCountWithNamespace(namespace.Name)
	if err != nil {
		log.Error("get services count with namespace err", utils.RequestID(ctx), zap.Error(err))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if total != 0 {
		log.Error("the removed namespace has remain services", utils.RequestID(ctx))
		return api.NewNamespaceResponse(apimodel.Code_NamespaceExistedServices, req)
	}

	// 判断属于该命名空间的服务是否都已经被删除
	total, err = s.getConfigGroupCountWithNamespace(namespace.Name)
	if err != nil {
		log.Error("get config group count with namespace err", utils.RequestID(ctx), zap.Error(err))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if total != 0 {
		log.Error("the removed namespace has remain config-group", utils.RequestID(ctx))
		return api.NewNamespaceResponse(apimodel.Code_NamespaceExistedConfigGroups, req)
	}

	// 存储层操作
	if err := tx.DeleteNamespace(namespace.Name); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}

	s.caches.Service().CleanNamespace(namespace.Name)

	log.Info("delete namespace", utils.RequestID(ctx), zap.String("name", namespace.Name))
	s.RecordHistory(namespaceRecordEntry(ctx, req, types.ODelete))
	return api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateNamespaces 批量修改命名空间
func (s *Server) UpdateNamespaces(ctx context.Context, req []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	if checkError := checkBatchNamespace(req); checkError != nil {
		return checkError
	}

	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, namespace := range req {
		response := s.UpdateNamespace(ctx, namespace)
		api.Collect(responses, response)
	}

	return responses
}

// UpdateNamespace 修改单个命名空间
func (s *Server) UpdateNamespace(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	// 参数检查
	if resp := checkReviseNamespace(ctx, req); resp != nil {
		return resp
	}

	// 权限校验
	namespace, resp := s.checkNamespaceAuthority(ctx, req)
	if resp != nil {
		return resp
	}
	// 修改
	s.updateNamespaceAttribute(req, namespace)

	// 存储层操作
	if err := s.storage.UpdateNamespace(namespace); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}

	log.Info("update namespace", zap.String("name", namespace.Name), utils.RequestID(ctx))
	s.RecordHistory(namespaceRecordEntry(ctx, req, types.OUpdate))
	return api.NewNamespaceResponse(apimodel.Code_ExecuteSuccess, req)
}

/**
 * @brief 修改命名空间属性
 */
func (s *Server) updateNamespaceAttribute(req *apimodel.Namespace, namespace *types.Namespace) {
	if req.GetComment() != nil {
		namespace.Comment = req.GetComment().GetValue()
	}

	if req.GetOwners() != nil {
		namespace.Owner = req.GetOwners().GetValue()
	}

	exportTo := map[string]struct{}{}
	for i := range req.GetServiceExportTo() {
		exportTo[req.GetServiceExportTo()[i].GetValue()] = struct{}{}
	}

	namespace.Metadata = req.GetMetadata()
	namespace.ServiceExportTo = exportTo
}

// GetNamespaces 查询命名空间
func (s *Server) GetNamespaces(ctx context.Context, query map[string][]string) *apiservice.BatchQueryResponse {
	filter, offset, limit, checkError := checkGetNamespace(query)
	if checkError != nil {
		return checkError
	}

	amount, namespaces, err := s.caches.Namespace().Query(ctx, &cacheapi.NamespaceArgs{
		Filter: filter,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = protobuf.NewUInt32Value(amount)
	out.Size = protobuf.NewUInt32Value(uint32(len(namespaces)))
	var totalServiceCount, totalInstanceCount, totalHealthInstanceCount uint32
	for _, namespace := range namespaces {
		nsCntInfo := s.caches.Service().GetNamespaceCntInfo(namespace.Name)
		api.AddNamespace(out, &apimodel.Namespace{
			Id:                       protobuf.NewStringValue(namespace.Name),
			Name:                     protobuf.NewStringValue(namespace.Name),
			Comment:                  protobuf.NewStringValue(namespace.Comment),
			Owners:                   protobuf.NewStringValue(namespace.Owner),
			Ctime:                    protobuf.NewStringValue(commontime.Time2String(namespace.CreateTime)),
			Mtime:                    protobuf.NewStringValue(commontime.Time2String(namespace.ModifyTime)),
			TotalServiceCount:        protobuf.NewUInt32Value(nsCntInfo.ServiceCount),
			TotalInstanceCount:       protobuf.NewUInt32Value(nsCntInfo.InstanceCnt.TotalInstanceCount),
			TotalHealthInstanceCount: protobuf.NewUInt32Value(nsCntInfo.InstanceCnt.HealthyInstanceCount),
			ServiceExportTo:          namespace.ListServiceExportTo(),
			Editable:                 protobuf.NewBoolValue(true),
			Deleteable:               protobuf.NewBoolValue(true),
			Metadata:                 namespace.Metadata,
		})
		totalServiceCount += nsCntInfo.ServiceCount
		totalInstanceCount += nsCntInfo.InstanceCnt.TotalInstanceCount
		totalHealthInstanceCount += nsCntInfo.InstanceCnt.HealthyInstanceCount
	}
	api.AddNamespaceSummary(out, &apimodel.Summary{
		TotalServiceCount:        totalServiceCount,
		TotalInstanceCount:       totalInstanceCount,
		TotalHealthInstanceCount: totalHealthInstanceCount,
	})
	return out
}

// 根据命名空间查询服务总数
func (s *Server) getServicesCountWithNamespace(namespace string) (uint32, error) {
	filter := map[string]string{"namespace": namespace}
	total, _, err := s.storage.GetServices(filter, nil, nil, 0, 1)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// 根据命名空间查询配置分组总数
func (s *Server) getConfigGroupCountWithNamespace(namespace string) (uint32, error) {
	total, err := s.storage.CountConfigGroups(namespace)
	if err != nil {
		return 0, err
	}
	return uint32(total), nil
}

// loadNamespace
func (s *Server) loadNamespace(name string) (string, error) {
	if val := s.caches.Namespace().GetNamespace(name); val != nil {
		return name, nil
	}
	val, err := s.storage.GetNamespace(name)
	if err != nil {
		return "", err
	}
	if val == nil {
		return "", nil
	}
	return val.Name, nil
}

// 检查namespace的权限，并且返回namespace
func (s *Server) checkNamespaceAuthority(
	ctx context.Context, req *apimodel.Namespace) (*types.Namespace, *apiservice.Response) {
	namespaceName := req.GetName().GetValue()
	// namespaceToken := parseNamespaceToken(ctx, req)

	// 检查是否存在
	namespace, err := s.storage.GetNamespace(namespaceName)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return nil, api.NewNamespaceResponse(storeapi.StoreCode2APICode(err), req)
	}
	if namespace == nil {
		return nil, api.NewNamespaceResponse(apimodel.Code_NotFoundResource, req)
	}
	return namespace, nil
}

// 检查批量请求
func checkBatchNamespace(req []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}

	return nil
}

// 检查创建命名空间请求参数
func checkCreateNamespace(req *apimodel.Namespace) *apiservice.Response {
	if req == nil {
		return api.NewNamespaceResponse(apimodel.Code_EmptyRequest, req)
	}

	if err := valid.CheckResourceName(req.GetName()); err != nil {
		return api.NewNamespaceResponse(apimodel.Code_InvalidNamespaceName, req)
	}

	return nil
}

// 检查删除/修改命名空间请求参数
func checkReviseNamespace(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	if req == nil {
		return api.NewNamespaceResponse(apimodel.Code_EmptyRequest, req)
	}

	if err := valid.CheckResourceName(req.GetName()); err != nil {
		return api.NewNamespaceResponse(apimodel.Code_InvalidNamespaceName, req)
	}
	return nil
}

// 检查查询命名空间请求参数
func checkGetNamespace(query map[string][]string) (map[string][]string, int, int, *apiservice.BatchQueryResponse) {
	filter := make(map[string][]string)

	if value := query["name"]; len(value) > 0 {
		filter["name"] = value
	}

	if value := query["owner"]; len(value) > 0 {
		filter["owner"] = value
	}

	offset, err := valid.CheckQueryOffset(query["offset"])
	if err != nil {
		return nil, 0, 0, api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	limit, err := valid.CheckQueryLimit(query["limit"])
	if err != nil {
		return nil, 0, 0, api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	return filter, offset, limit, nil
}

// 生成命名空间的记录entry
func namespaceRecordEntry(ctx context.Context, req *apimodel.Namespace, opt types.OperationType) *types.RecordEntry {
	marshaler := jsonpb.Marshaler{}
	datail, _ := marshaler.MarshalToString(req)
	return &types.RecordEntry{
		ResourceType:  types.RNamespace,
		ResourceName:  req.GetName().GetValue(),
		Namespace:     req.GetName().GetValue(),
		OperationType: opt,
		Operator:      utils.ParseOperator(ctx),
		Detail:        datail,
		HappenTime:    time.Now(),
	}
}
