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
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addServiceAccess .
func (h *HTTPServer) addServiceAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateServicesApiDocs(ws.POST("/services").To(h.CreateServices)))
	ws.Route(docs.EnrichDeleteServicesApiDocs(ws.POST("/services/delete").To(h.DeleteServices)))
	ws.Route(docs.EnrichUpdateServicesApiDocs(ws.PUT("/services").To(h.UpdateServices)))
	ws.Route(docs.EnrichGetServicesApiDocs(ws.GET("/services").To(h.GetServices)))
	ws.Route(docs.EnrichGetAllServicesApiDocs(ws.GET("/services/all").To(h.GetAllServices)))
	ws.Route(docs.EnrichGetServicesCountApiDocs(ws.GET("/services/count").To(h.GetServicesCount)))
	ws.Route(docs.EnrichCreateServiceAliasApiDocs(ws.POST("/service/alias").To(h.CreateServiceAlias)))
	ws.Route(docs.EnrichUpdateServiceAliasApiDocs(ws.PUT("/service/alias").To(h.UpdateServiceAlias)))
	ws.Route(docs.EnrichGetServiceAliasesApiDocs(ws.GET("/service/aliases").To(h.GetServiceAliases)))
	ws.Route(docs.EnrichDeleteServiceAliasesApiDocs(ws.POST("/service/aliases/delete").To(h.DeleteServiceAliases)))
	ws.Route(docs.EnrichDeleteServiceAliasesApiDocs(ws.GET("/service/subscribers").To(h.GetServiceSubscribers)))

	// 服务契约相关
	ws.Route(docs.EnrichCreateServiceContractsApiDocs(ws.POST("/service/contracts").To(h.CreateServiceContract)))
	ws.Route(docs.EnrichGetServiceContractsApiDocs(ws.GET("/service/contracts").To(h.GetServiceContracts)))
	ws.Route(docs.EnrichDeleteServiceContractsApiDocs(ws.POST("/service/contracts/delete").To(h.DeleteServiceContracts)))
	ws.Route(docs.EnrichGetServiceContractsApiDocs(ws.GET("/service/contract/versions").To(h.GetServiceContractVersions)))
	ws.Route(docs.EnrichAddServiceContractInterfacesApiDocs(ws.POST("/service/contract/methods").To(h.CreateServiceContractInterfaces)))
	ws.Route(docs.EnrichAppendServiceContractInterfacesApiDocs(ws.PUT("/service/contract/methods/append").To(h.AppendServiceContractInterfaces)))
	ws.Route(docs.EnrichDeleteServiceContractsApiDocs(ws.POST("/service/contract/methods/delete").To(h.DeleteServiceContractInterfaces)))
}

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

// GetServiceSubscribers 获取服务订阅者
func (h *HTTPServer) GetServiceSubscribers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ctx := handler.ParseHeaderContext()
	queryParams := httpcommon.ParseQueryParams(req)
	handler.WriteHeaderAndProto(h.namingServer.GetServiceSubscribers(ctx, queryParams))
}
