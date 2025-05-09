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

package policy

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
)

// CreateRoles 批量创建角色
func (svr *Server) CreateRoles(ctx context.Context, reqs []*apisecurity.Role) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range reqs {
		rsp := svr.CreateRole(ctx, reqs[i])
		api.Collect(responses, rsp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// CreateRole 创建角色
func (svr *Server) CreateRole(ctx context.Context, req *apisecurity.Role) *apiservice.Response {
	req.Owner = utils.ParseOwnerID(ctx)

	saveData := &authtypes.Role{}
	saveData.FromSpec(req)
	saveData.ID = utils.NewUUID()

	if err := svr.storage.AddRole(saveData); err != nil {
		log.Error("[Auth][Role] create role into store", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	// TODO : 记录操作日志
	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

// UpdateRoles 批量更新角色
func (svr *Server) UpdateRoles(ctx context.Context, reqs []*apisecurity.Role) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range reqs {
		rsp := svr.UpdateRole(ctx, reqs[i])
		api.Collect(responses, rsp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// UpdateRole 批量更新角色
func (svr *Server) UpdateRole(ctx context.Context, req *apisecurity.Role) *apiservice.Response {
	newData := &authtypes.Role{}
	newData.FromSpec(req)

	saveData, err := svr.storage.GetRole(newData.ID)
	if err != nil {
		log.Error("[Auth][Role] get one role from store", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		log.Error("[Auth][Role] not find expect role", utils.RequestID(ctx),
			zap.String("id", newData.ID))
		return api.NewAuthResponse(apimodel.Code_NotFoundResource)
	}

	if needUpdate := func() bool {
		needUpdate := false
		if saveData.Comment != newData.Comment {
			saveData.Comment = newData.Comment
			needUpdate = true
		}
		if saveData.Source != newData.Source {
			saveData.Source = newData.Source
			needUpdate = true
		}
		if !slices.EqualFunc(saveData.Users, newData.Users, func(e1, e2 authtypes.Principal) bool {
			return e1.PrincipalID == e2.PrincipalID && e1.PrincipalType == e2.PrincipalType
		}) {
			saveData.Users = newData.Users
			needUpdate = true
		}
		if !slices.EqualFunc(saveData.UserGroups, newData.UserGroups, func(e1, e2 authtypes.Principal) bool {
			return e1.PrincipalID == e2.PrincipalID && e1.PrincipalType == e2.PrincipalType
		}) {
			saveData.UserGroups = newData.UserGroups
			needUpdate = true
		}
		if !maps.Equal(saveData.Metadata, newData.Metadata) {
			saveData.Metadata = newData.Metadata
			needUpdate = true
		}
		return needUpdate
	}(); !needUpdate {
		return api.NewResponse(apimodel.Code_NoNeedUpdate)
	}

	if err := svr.storage.UpdateRole(saveData); err != nil {
		log.Error("[Auth][Role] update role into store", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

// DeleteRoles 批量删除角色
func (svr *Server) DeleteRoles(ctx context.Context, reqs []*apisecurity.Role) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range reqs {
		rsp := svr.DeleteRole(ctx, reqs[i])
		api.Collect(responses, rsp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// DeleteRole 批量删除角色
func (svr *Server) DeleteRole(ctx context.Context, req *apisecurity.Role) *apiservice.Response {
	newData := &authtypes.Role{}
	newData.FromSpec(req)

	saveData, err := svr.storage.GetRole(newData.ID)
	if err != nil {
		log.Error("[Auth][Role] get one role from store", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		return api.NewAuthResponse(apimodel.Code_ExecuteSuccess)
	}

	tx, err := svr.storage.StartTx()
	if err != nil {
		log.Error("[Auth][Role] start tx", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := svr.storage.DeleteRole(tx, newData); err != nil {
		log.Error("[Auth][Role] update role into store", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	if err := svr.storage.CleanPrincipalPolicies(tx, authtypes.Principal{
		PrincipalID:   saveData.ID,
		PrincipalType: authtypes.PrincipalRole,
	}); err != nil {
		log.Error("[Auth][Role] clean role link policies", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Auth][Role] delete role commit tx", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

// GetRoles 查询角色列表
func (svr *Server) GetRoles(ctx context.Context, filters map[string]string) *apiservice.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(filters)

	total, ret, err := svr.cacheMgr.Role().Query(ctx, cachetypes.RoleSearchArgs{
		Filters: filters,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		log.Error("[Auth][Role] query roles list", utils.RequestID(ctx), zap.Error(err))
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	rsp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	rsp.Amount = protobuf.NewUInt32Value(total)
	rsp.Size = protobuf.NewUInt32Value(uint32(len(ret)))

	for i := range ret {
		if err := api.AddAnyDataIntoBatchQuery(rsp, ret[i].ToSpec()); err != nil {
			log.Error("[Auth][Role] add role to query list", utils.RequestID(ctx), zap.Error(err))
			return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	return rsp
}

func recordRoleEntry(ctx context.Context, req *apisecurity.Role, data *authtypes.Role, op types.OperationType) *types.RecordEntry {
	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RAuthRole,
		ResourceName:  fmt.Sprintf("%s(%s)", data.Name, data.ID),
		OperationType: op,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}

	return entry
}
