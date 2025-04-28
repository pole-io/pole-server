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

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addGroupAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateGroupApiDocs(ws.POST("/usergroups").To(h.CreateGroups)))
	ws.Route(docs.EnrichUpdateGroupsApiDocs(ws.PUT("/usergroups").To(h.UpdateGroups)))
	ws.Route(docs.EnrichGetGroupsApiDocs(ws.GET("/usergroups").To(h.GetGroups)))
	ws.Route(docs.EnrichDeleteGroupsApiDocs(ws.POST("/usergroups/delete").To(h.DeleteGroups)))
	ws.Route(docs.EnrichGetGroupApiDocs(ws.GET("/usergroup/detail").To(h.GetGroup)))
	ws.Route(docs.EnrichGetGroupTokenApiDocs(ws.GET("/usergroup/token").To(h.GetGroupToken)))
	ws.Route(docs.EnrichUpdateGroupTokenApiDocs(ws.PUT("/usergroup/token/enable").To(h.EnableGroupToken)))
	ws.Route(docs.EnrichResetGroupTokenApiDocs(ws.PUT("/usergroup/token/refresh").To(h.ResetGroupToken)))
}

// CreateGroup 创建用户组
func (h *HTTPServer) CreateGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	group := &apisecurity.UserGroup{}
	ctx, err := handler.Parse(group)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.CreateGroup(ctx, group))
}

// UpdateGroups 更新用户组
func (h *HTTPServer) UpdateGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var groups UserGroupArr

	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.UserGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.UpdateGroups(ctx, groups))
}

// DeleteGroups 删除用户组
func (h *HTTPServer) DeleteGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var groups UserGroupArr

	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.UserGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.DeleteGroups(ctx, groups))
}

// GetGroups 获取用户组列表
func (h *HTTPServer) GetGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.userSvr.GetGroups(ctx, queryParams))
}

// GetGroup 获取用户组详细
func (h *HTTPServer) GetGroup(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	group := &apisecurity.UserGroup{
		Id: protobuf.NewStringValue(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.userSvr.GetGroup(ctx, group))
}

// GetGroupToken 获取用户组 token
func (h *HTTPServer) GetGroupToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	group := &apisecurity.UserGroup{
		Id: protobuf.NewStringValue(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.userSvr.GetGroupToken(ctx, group))
}

// EnableGroupToken 更新用户组 token
func (h *HTTPServer) EnableGroupToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	group := &apisecurity.UserGroup{}

	ctx, err := handler.Parse(group)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.EnableGroupToken(ctx, group))
}

// ResetGroupToken 重置用户组 token
func (h *HTTPServer) ResetGroupToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	group := &apisecurity.UserGroup{}

	ctx, err := handler.Parse(group)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.ResetGroupToken(ctx, group))
}
