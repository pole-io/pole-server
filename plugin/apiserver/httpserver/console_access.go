/*
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

package httpserver

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

const (
	defaultReadAccess string = "default-read"
	defaultAccess     string = "default"
)

// GetCoreConsoleAccessServer 增加配置中心模块之后，namespace 作为两个模块的公共模块需要独立， restful path 以 /core 开头
func (h *HTTPServer) GetCoreV1ConsoleAccessServer(include []string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/core/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	consoleAccess := []string{defaultAccess}

	if len(include) == 0 {
		include = consoleAccess
	}

	var hasDefault = false
	for _, item := range include {
		if item == defaultAccess {
			hasDefault = true
			break
		}
	}
	for _, item := range include {
		switch item {
		case defaultReadAccess:
			if !hasDefault {
				h.addCoreDefaultReadAccess(ws)
			}
		case defaultAccess:
			h.addCoreDefaultAccess(ws)
		}
	}
	return ws
}

func (h *HTTPServer) addCoreDefaultReadAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichGetNamespacesApiDocsOld(ws.GET("/namespaces").To(h.GetNamespaces)))
}

func (h *HTTPServer) addCoreDefaultAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateNamespacesApiDocsOld(ws.POST("/namespaces").To(h.CreateNamespaces)))
	ws.Route(docs.EnrichDeleteNamespacesApiDocsOld(ws.POST("/namespaces/delete").To(h.DeleteNamespaces)))
	ws.Route(docs.EnrichUpdateNamespacesApiDocsOld(ws.PUT("/namespaces").To(h.UpdateNamespaces)))
	ws.Route(docs.EnrichGetNamespacesApiDocsOld(ws.GET("/namespaces").To(h.GetNamespaces)))

	//
	ws.Route(ws.GET("/clients").To(h.GetReportClients))
}

// CreateNamespaces 创建命名空间
func (h *HTTPServer) CreateNamespaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var namespaces NamespaceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.Namespace{}
		namespaces = append(namespaces, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namespaceServer.CreateNamespaces(ctx, namespaces))
}

// DeleteNamespaces 删除命名空间
func (h *HTTPServer) DeleteNamespaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var namespaces NamespaceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.Namespace{}
		namespaces = append(namespaces, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namespaceServer.DeleteNamespaces(ctx, namespaces))
}

// UpdateNamespaces 修改命名空间
func (h *HTTPServer) UpdateNamespaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var namespaces NamespaceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.Namespace{}
		namespaces = append(namespaces, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namespaceServer.UpdateNamespaces(ctx, namespaces))
}

// GetNamespaces 查询命名空间
func (h *HTTPServer) GetNamespaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ret := h.namespaceServer.GetNamespaces(handler.ParseHeaderContext(), req.Request.URL.Query())
	handler.WriteHeaderAndProto(ret)
}
