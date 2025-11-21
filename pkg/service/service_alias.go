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

	"go.uber.org/zap"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

var (
	// AliasFilterAttributes filer attrs alias
	AliasFilterAttributes = map[string]bool{
		"alias":           true,
		"alias_namespace": true,
		"namespace":       true,
		"service":         true,
		"owner":           true,
		"offset":          true,
		"limit":           true,
	}
)

// CreateServiceAlias 创建服务别名
func (s *Server) CreateServiceAlias(ctx context.Context, req *apiservice.ServiceAlias) *apimodel.Response {
	tx, err := s.storage.CreateTransaction()
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
	}
	defer func() { _ = tx.Commit() }()

	service, response, done := s.checkPointServiceAlias(ctx, tx, req)
	if done {
		return response
	}

	// 检查是否存在同名的alias
	if req.GetAlias() != "" {
		oldAlias, getErr := s.storage.GetService(req.GetAlias(),
			req.GetAliasNamespace())
		if getErr != nil {
			log.Error(getErr.Error(), utils.RequestID(ctx))
			return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
		}
		if oldAlias != nil {
			return api.NewServiceAliasResponse(apimodel.Code_ExistedResource, req)
		}
	}

	// 构建别名的信息，这里包括了创建SID
	input, resp := s.createServiceAliasModel(req, service.ID)
	if resp != nil {
		return resp
	}
	if err := s.storage.AddService(input); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
	}

	log.Info(fmt.Sprintf("create service alias, service(%s, %s), alias(%s, %s)",
		// 注释：字段访问改动 - Service和Namespace字段直接访问而非通过.Value
		req.Service, req.Namespace, input.Name, input.Namespace), utils.RequestID(ctx))
	out := &apiservice.ServiceAlias{
		Service:        req.Service,
		Namespace:      req.Namespace,
		Alias:          req.Alias,
		AliasNamespace: req.AliasNamespace,
		// 注释：ServiceToken字段类型改动 - 从*wrapperspb.StringValue改为string
		ServiceToken: string(input.Token),
	}
	if out.GetAlias() == "" {
		// 注释：Alias字段赋值改动 - 使用string而非*wrapperspb.StringValue包装
		out.Alias = string(input.Name)
	}
	record := &apiservice.Service{Name: out.Alias, Namespace: out.AliasNamespace}
	s.RecordHistory(ctx, serviceRecordEntry(ctx, record, input, types.OCreate))
	return api.NewServiceAliasResponse(apimodel.Code_ExecuteSuccess, out)
}

func (s *Server) checkPointServiceAlias(ctx context.Context,
	tx store.Transaction, req *apiservice.ServiceAlias) (*svctypes.Service, *apimodel.Response, bool) {
	// 检查指向服务是否存在以及是否为别名
	service, err := tx.LockService(req.GetService(), req.GetNamespace())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return nil, api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req), true
	}
	if service == nil {
		return nil, api.NewServiceAliasResponse(apimodel.Code_NotFoundResource, req), true
	}
	// 检查该服务是否已经是一个别名服务，不允许再为别名创建别名
	if service.IsAlias() {
		return nil, api.NewServiceAliasResponse(apimodel.Code_BadRequest, req), true
	}
	return service, nil, false
}

// DeleteServiceAlias 删除服务别名
//
//	需要带上源服务name，namespace，token
//	另外一种删除别名的方式，是直接调用删除服务的接口，也是可行的
func (s *Server) DeleteServiceAlias(ctx context.Context, req *apiservice.ServiceAlias) *apimodel.Response {
	rid := utils.ParseRequestID(ctx)
	alias, err := s.storage.GetService(req.GetAlias(),
		req.GetAliasNamespace())
	if err != nil {
		log.Error(err.Error(), utils.ZapRequestID(rid))
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
	}
	if alias == nil {
		return api.NewServiceAliasResponse(apimodel.Code_NotFoundResource, req)
	}

	// 直接删除alias
	if err := s.storage.DeleteServiceAlias(req.GetAlias(),
		req.GetAliasNamespace()); err != nil {
		log.Error(err.Error(), utils.ZapRequestID(rid))
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
	}

	s.RecordHistory(ctx, serviceRecordEntry(ctx, &apiservice.Service{
		Name:      req.GetAlias(),
		Namespace: req.GetAliasNamespace(),
	}, alias, types.ODelete))
	return api.NewServiceAliasResponse(apimodel.Code_ExecuteSuccess, req)
}

// DeleteServiceAliases 删除服务别名列表
func (s *Server) DeleteServiceAliases(
	ctx context.Context, req []*apiservice.ServiceAlias) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, alias := range req {
		response := s.DeleteServiceAlias(ctx, alias)
		api.Collect(responses, response)
	}

	return api.FormatBatchWriteResponse(responses)
}

// UpdateServiceAlias 修改服务别名
func (s *Server) UpdateServiceAlias(ctx context.Context, req *apiservice.ServiceAlias) *apimodel.Response {
	// 检查服务别名是否存在
	alias, err := s.storage.GetService(req.GetAlias(), req.GetAliasNamespace())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
	}
	if alias == nil {
		return api.NewServiceAliasResponse(apimodel.Code_NotFoundResource, req)
	}

	// 检查将要指向的服务是否存在
	service, err := s.storage.GetService(req.GetService(), req.GetNamespace())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req)
	}
	if service == nil {
		return api.NewServiceAliasResponse(apimodel.Code_NotFoundResource, req)
	}
	// 检查该服务是否已经是一个别名服务，不允许再为别名创建别名
	if service.IsAlias() {
		return api.NewServiceAliasResponse(apimodel.Code_BadRequest, req)
	}

	// 判断是否需要修改
	resp, needUpdate, needUpdateOwner := s.updateServiceAliasAttribute(req, alias, service.ID)
	if resp != nil {
		return resp
	}

	if !needUpdate {
		log.Info("update service alias data no change, no need update", utils.RequestID(ctx),
			zap.String("service alias", req.String()))
		return api.NewServiceAliasResponse(apimodel.Code_NoNeedUpdate, req)
	}

	// 执行存储层操作
	if err := s.storage.UpdateServiceAlias(alias, needUpdateOwner); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperServiceAliasResponse(req, err)
	}

	log.Info(fmt.Sprintf("update service alias, service(%s, %s), alias(%s)",
		req.GetService(), req.GetNamespace(), req.GetAlias()), utils.RequestID(ctx))

	record := &apiservice.Service{Name: req.Alias, Namespace: req.Namespace}
	s.RecordHistory(ctx, serviceRecordEntry(ctx, record, alias, types.OUpdate))

	return api.NewServiceAliasResponse(apimodel.Code_ExecuteSuccess, req)
}

// GetServiceAliases 查找服务别名
func (s *Server) GetServiceAliases(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	// 先处理offset和limit
	offset, limit, err := valid.ParseOffsetAndLimit(query)
	if err != nil {
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	// 处理剩余的参数
	filter := make(map[string]string)
	for key, value := range query {
		if _, ok := AliasFilterAttributes[key]; !ok {
			log.Errorf("[Server][Alias][Query] attribute(%s) is not allowed", key)
			return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
		}
		filter[key] = value
	}

	total, aliases, err := s.storage.GetServiceAliases(filter, offset, limit)
	if err != nil {
		log.Errorf("[Server][Alias] get aliases err: %s", err.Error())
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = uint32(total)
	resp.Size = uint32(len(aliases))
	// 注释：响应结构重大改动 - 根据 pole-io/specification，BatchQueryResponse 只有 data 字段，需要将别名数据序列化到 data 中
	// 根据 pole-io/specification，BatchQueryResponse 只有 data 字段，需要将别名数据序列化到 data 中
	for _, entry := range aliases {
		item := &apiservice.ServiceAlias{
			Id:             string(entry.ID),
			Service:        string(entry.Service),
			Namespace:      string(entry.Namespace),
			Alias:          string(entry.Alias),
			AliasNamespace: string(entry.AliasNamespace),
			Owners:         string(entry.Owner),
			Comment:        string(entry.Comment),
			Ctime:          commontime.Time2String(entry.CreateTime),
			Mtime:          commontime.Time2String(entry.ModifyTime),
			Editable:       true,
			Deleteable:     bool(true),
		}
		// 将 item 添加到 data 字段中
		if addErr := api.AddAnyDataIntoBatchQuery(resp, item); addErr != nil {
			log.Error("[Server][Alias] add service_alias as any data fail", utils.RequestID(ctx), zap.Error(addErr))
			continue
		}
	}

	return resp
}

// updateServiceAliasAttribute 修改服务别名属性
func (s *Server) updateServiceAliasAttribute(req *apiservice.ServiceAlias, alias *svctypes.Service, serviceID string) (
	*apimodel.Response, bool, bool) {
	var (
		needUpdate      bool
		needUpdateOwner bool
	)

	// 获取当前指向服务
	service, err := s.storage.GetServiceByID(alias.Reference)
	if err != nil {
		return api.NewServiceAliasResponse(storeapi.StoreCode2APICode(err), req), needUpdate, needUpdateOwner
	}

	if service.ID != serviceID {
		alias.Reference = serviceID
		needUpdate = true
	}

	if len(req.GetOwners()) > 0 && req.GetOwners() != alias.Owner {
		alias.Owner = req.GetOwners()
		needUpdate = true
		needUpdateOwner = true
	}

	if req.GetComment() != "" && req.GetComment() != alias.Comment {
		alias.Comment = req.GetComment()
		needUpdate = true
	}

	if needUpdate {
		alias.Revision = utils.NewUUID()
	}

	return nil, needUpdate, needUpdateOwner
}

// createServiceAliasModel 构建存储结构
func (s *Server) createServiceAliasModel(req *apiservice.ServiceAlias, svcId string) (
	*svctypes.Service, *apimodel.Response) {
	out := &svctypes.Service{
		ID:        utils.NewUUID(),
		Name:      req.GetAlias(),
		Namespace: req.GetAliasNamespace(),
		Reference: svcId,
		Token:     utils.NewUUID(),
		Owner:     req.GetOwners(),
		Comment:   req.GetComment(),
		Revision:  utils.NewUUID(),
	}
	return out, nil
}

// wrapperServiceAliasResponse wrapper service alias error
func wrapperServiceAliasResponse(alias *apiservice.ServiceAlias, err error) *apimodel.Response {
	if err == nil {
		return nil
	}
	resp := api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	return resp
}
