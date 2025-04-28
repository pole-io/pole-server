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

func (h *HTTPServer) addUserAccess(ws *restful.WebService) {
	// 用户
	ws.Route(docs.EnrichLoginApiDocs(ws.POST("/user/login").To(h.Login)))
	ws.Route(docs.EnrichGetUsersApiDocs(ws.GET("/users").To(h.GetUsers)))
	ws.Route(docs.EnrichCreateUsersApiDocs(ws.POST("/users").To(h.CreateUsers)))
	ws.Route(docs.EnrichUpdateUserApiDocs(ws.PUT("/users").To(h.UpdateUsers)))
	ws.Route(docs.EnrichDeleteUsersApiDocs(ws.POST("/users/delete").To(h.DeleteUsers)))
	ws.Route(docs.EnrichUpdateUserPasswordApiDocs(ws.PUT("/user/password").To(h.UpdateUserPassword)))
	ws.Route(docs.EnrichGetUserTokenApiDocs(ws.GET("/user/token").To(h.GetUserToken)))
	ws.Route(docs.EnrichUpdateUserTokenApiDocs(ws.PUT("/user/token/enable").To(h.EnableUserToken)))
	ws.Route(docs.EnrichResetUserTokenApiDocs(ws.PUT("/user/token/refresh").To(h.ResetUserToken)))
}

// Login 登录函数
func (h *HTTPServer) Login(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	loginReq := &apisecurity.LoginRequest{}

	_, err := handler.Parse(loginReq)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewAuthResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.Login(loginReq))
}

// CreateUsers 批量创建用户
func (h *HTTPServer) CreateUsers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var users UserArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.User{}
		users = append(users, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.CreateUsers(ctx, users))
}

// UpdateUsers 更新用户
func (h *HTTPServer) UpdateUsers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var users UserArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.User{}
		users = append(users, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.UpdateUsers(ctx, users))
}

// UpdateUserPassword 更新用户
func (h *HTTPServer) UpdateUserPassword(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	user := &apisecurity.ModifyUserPassword{}

	ctx, err := handler.Parse(user)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.UpdateUserPassword(ctx, user))
}

// DeleteUsers 批量删除用户
func (h *HTTPServer) DeleteUsers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var users UserArr

	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.User{}
		users = append(users, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.DeleteUsers(ctx, users))
}

// GetUsers 查询用户
func (h *HTTPServer) GetUsers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.userSvr.GetUsers(ctx, queryParams))
}

// GetUserToken 获取这个用户所关联的所有用户组列表信息，支持翻页
func (h *HTTPServer) GetUserToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)

	user := &apisecurity.User{
		Id: protobuf.NewStringValue(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.userSvr.GetUserToken(handler.ParseHeaderContext(), user))
}

// EnableUserToken 更改用户的token
func (h *HTTPServer) EnableUserToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	user := &apisecurity.User{}
	ctx, err := handler.Parse(user)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.EnableUserToken(ctx, user))
}

// ResetUserToken 重置用户 token
func (h *HTTPServer) ResetUserToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	user := &apisecurity.User{}

	ctx, err := handler.Parse(user)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.userSvr.ResetUserToken(ctx, user))
}
