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

package goverrule_auth

import (
	"context"

	"github.com/polarismesh/specification/source/go/api/v1/security"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

// ResourceEvent 资源事件
type ResourceEvent struct {
	Resource authtypes.ResourceEntry

	AddPrincipals []authtypes.Principal
	DelPrincipals []authtypes.Principal
	IsRemove      bool
}

// Before this function is called before the resource operation
func (svr *Server) Before(ctx context.Context, resourceType types.Resource) {
	// do nothing
}

// After this function is called after the resource operation
func (svr *Server) After(ctx context.Context, resourceType types.Resource, res *ResourceEvent) error {
	// 资源删除，触发所有关联的策略进行一个 update 操作更新
	return svr.onChangeResource(ctx, res)
}

// onChangeResource 服务资源的处理，只处理服务，namespace 只由 namespace 相关的进行处理，
func (svr *Server) onChangeResource(ctx context.Context, res *ResourceEvent) error {
	authCtx := ctx.Value(types.ContextAuthContextKey).(*authtypes.AcquireContext)

	authCtx.SetAttachment(authtypes.ResourceAttachmentKey, map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		res.Resource.Type: {
			res.Resource,
		},
	})

	var users, removeUsers []string
	var groups, removeGroups []string

	for i := range res.AddPrincipals {
		switch res.AddPrincipals[i].PrincipalType {
		case authtypes.PrincipalUser:
			users = append(users, res.AddPrincipals[i].PrincipalID)
		case authtypes.PrincipalGroup:
			groups = append(groups, res.AddPrincipals[i].PrincipalID)
		}
	}
	for i := range res.DelPrincipals {
		switch res.DelPrincipals[i].PrincipalType {
		case authtypes.PrincipalUser:
			removeUsers = append(removeUsers, res.DelPrincipals[i].PrincipalID)
		case authtypes.PrincipalGroup:
			removeGroups = append(removeGroups, res.DelPrincipals[i].PrincipalID)
		}
	}

	authCtx.SetAttachment(authtypes.LinkUsersKey, users)
	authCtx.SetAttachment(authtypes.RemoveLinkUsersKey, removeUsers)

	authCtx.SetAttachment(authtypes.LinkGroupsKey, groups)
	authCtx.SetAttachment(authtypes.RemoveLinkGroupsKey, removeGroups)

	return svr.policySvr.AfterResourceOperation(authCtx)
}

func (s *Server) afterRuleResource(ctx context.Context, r types.Resource, res authtypes.ResourceEntry, remove bool) error {
	event := &ResourceEvent{
		Resource: res,
		IsRemove: remove,
	}

	return s.After(ctx, r, event)
}

func (s *Server) afterServiceResource(ctx context.Context, req *apiservice.Service, remove bool) error {
	event := &ResourceEvent{
		Resource: authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       req.GetId().GetValue(),
			Metadata: req.GetMetadata(),
		},
		AddPrincipals: func() []authtypes.Principal {
			ret := make([]authtypes.Principal, 0, 4)
			for i := range req.UserIds {
				ret = append(ret, authtypes.Principal{
					PrincipalType: authtypes.PrincipalUser,
					PrincipalID:   req.UserIds[i].GetValue(),
				})
			}
			for i := range req.GroupIds {
				ret = append(ret, authtypes.Principal{
					PrincipalType: authtypes.PrincipalGroup,
					PrincipalID:   req.GroupIds[i].GetValue(),
				})
			}
			return ret
		}(),
		DelPrincipals: func() []authtypes.Principal {
			ret := make([]authtypes.Principal, 0, 4)
			for i := range req.RemoveUserIds {
				ret = append(ret, authtypes.Principal{
					PrincipalType: authtypes.PrincipalUser,
					PrincipalID:   req.RemoveUserIds[i].GetValue(),
				})
			}
			for i := range req.RemoveGroupIds {
				ret = append(ret, authtypes.Principal{
					PrincipalType: authtypes.PrincipalGroup,
					PrincipalID:   req.RemoveGroupIds[i].GetValue(),
				})
			}
			return ret
		}(),
		IsRemove: remove,
	}
	return s.After(ctx, types.RService, event)
}
