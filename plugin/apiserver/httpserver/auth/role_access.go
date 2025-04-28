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

package auth

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addRoleAccess(ws *restful.WebService) {
	// 角色
	ws.Route(docs.EnrichGetRolesApiDocs(ws.GET("/roles").To(h.GetRoles)))
	ws.Route(docs.EnrichCreateRolesApiDocs(ws.POST("/roles").To(h.CreateRoles)))
	ws.Route(docs.EnrichDeleteRolesApiDocs(ws.POST("/roles/delete").To(h.DeleteRoles)))
	ws.Route(docs.EnrichUpdateRolesApiDocs(ws.PUT("/roles").To(h.UpdateRoles)))
}

// CreateRoles .
func (h *HTTPServer) CreateRoles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	roles := make([]*apisecurity.Role, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.Role{}
		roles = append(roles, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.policySvr.CreateRoles(ctx, roles))
}

// UpdateRoles .
func (h *HTTPServer) UpdateRoles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	roles := make([]*apisecurity.Role, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.Role{}
		roles = append(roles, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.policySvr.UpdateRoles(ctx, roles))
}

// DeleteRoles .
func (h *HTTPServer) DeleteRoles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	roles := make([]*apisecurity.Role, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.Role{}
		roles = append(roles, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.policySvr.DeleteRoles(ctx, roles))
}

// GetRoles 查询角色列表
func (h *HTTPServer) GetRoles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.policySvr.GetRoles(ctx, queryParams))
}
