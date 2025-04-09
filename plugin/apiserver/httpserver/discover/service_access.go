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

package discover

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// CreateServices 创建服务
func (h *HTTPServer) CreateServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var services ServiceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Service{}
		services = append(services, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.CreateServices(ctx, services))
}

// DeleteServices 删除服务
func (h *HTTPServer) DeleteServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var services ServiceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Service{}
		services = append(services, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteServices(ctx, services)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// UpdateServices 修改服务
func (h *HTTPServer) UpdateServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var services ServiceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Service{}
		services = append(services, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateServices(ctx, services)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
}

// GetAllServices 查询服务
func (h *HTTPServer) GetAllServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()
	ret := h.namingServer.GetAllServices(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetServices 查询服务
func (h *HTTPServer) GetServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()
	ret := h.namingServer.GetServices(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetServicesCount 查询服务总数
func (h *HTTPServer) GetServicesCount(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	ret := h.namingServer.GetServicesCount(handler.ParseHeaderContext())
	handler.WriteHeaderAndProto(ret)
}

// GetServiceToken 获取服务token
func (h *HTTPServer) GetServiceToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	token := req.HeaderParameter("Polaris-Token")
	ctx := context.WithValue(context.Background(), types.StringContext("polaris-token"), token)

	queryParams := httpcommon.ParseQueryParams(req)
	service := &apiservice.Service{
		Name:      protobuf.NewStringValue(queryParams["name"]),
		Namespace: protobuf.NewStringValue(queryParams["namespace"]),
		Token:     protobuf.NewStringValue(queryParams["token"]),
	}

	ret := h.namingServer.GetServiceToken(ctx, service)
	handler.WriteHeaderAndProto(ret)
}

// UpdateServiceToken 更新服务token
func (h *HTTPServer) UpdateServiceToken(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var service apiservice.Service
	ctx, err := handler.Parse(&service)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.UpdateServiceToken(ctx, &service))
}

// CreateServiceAlias service alias
func (h *HTTPServer) CreateServiceAlias(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var alias apiservice.ServiceAlias
	ctx, err := handler.Parse(&alias)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.CreateServiceAlias(ctx, &alias))
}

// UpdateServiceAlias 修改服务别名
func (h *HTTPServer) UpdateServiceAlias(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var alias apiservice.ServiceAlias
	ctx, err := handler.Parse(&alias)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateServiceAlias(ctx, &alias)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// DeleteServiceAliases 删除服务别名
func (h *HTTPServer) DeleteServiceAliases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var aliases ServiceAliasArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.ServiceAlias{}
		aliases = append(aliases, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	ret := h.namingServer.DeleteServiceAliases(ctx, aliases)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// GetServiceAliases 根据源服务获取服务别名
func (h *HTTPServer) GetServiceAliases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetServiceAliases(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetServiceOwner 根据服务获取服务负责人
func (h *HTTPServer) GetServiceOwner(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var services ServiceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Service{}
		services = append(services, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.GetServiceOwner(ctx, services))
}
