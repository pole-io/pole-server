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

func (h *HTTPServer) addPolicyRuleAccess(ws *restful.WebService) {
	// 鉴权策略
	ws.Route(docs.EnrichCreateStrategyApiDocs(ws.POST("/policies").To(h.CreatePolicies)))
	ws.Route(docs.EnrichUpdateStrategiesApiDocs(ws.PUT("/policies").To(h.UpdatePolicies)))
	ws.Route(docs.EnrichDeleteStrategiesApiDocs(ws.POST("/policies/delete").To(h.DeletePolicies)))
	ws.Route(docs.EnrichGetStrategiesApiDocs(ws.GET("/policies").To(h.GetPolicies)))
	ws.Route(docs.EnrichGetStrategyApiDocs(ws.GET("/policy/detail").To(h.GetPolicy)))
	ws.Route(docs.EnrichGetPrincipalResourcesApiDocs(ws.GET("/principal/resources").To(h.GetPrincipalResources)))
}

// CreatePolicies 创建鉴权策略
func (h *HTTPServer) CreatePolicies(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var policies []*apisecurity.AuthStrategy

	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.AuthStrategy{}
		policies = append(policies, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewAuthResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.policySvr.CreatePolicies(ctx, policies))
}

// UpdatePolicies 更新鉴权策略
func (h *HTTPServer) UpdatePolicies(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var strategies StrategyArr

	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.AuthStrategy{}
		strategies = append(strategies, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.policySvr.UpdatePolicies(ctx, strategies))
}

// DeletePolicies 批量删除鉴权策略
func (h *HTTPServer) DeletePolicies(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var strategies StrategyArr

	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.AuthStrategy{}
		strategies = append(strategies, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.policySvr.DeletePolicies(ctx, strategies))
}

// GetPolicies 批量获取鉴权策略
func (h *HTTPServer) GetPolicies(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.policySvr.GetPolicies(ctx, queryParams))
}

// GetPolicy 获取鉴权策略详细
func (h *HTTPServer) GetPolicy(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	strategy := &apisecurity.AuthStrategy{
		Id: protobuf.NewStringValue(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.policySvr.GetPolicy(ctx, strategy))
}

// GetPrincipalResources 获取鉴权策略详细
func (h *HTTPServer) GetPrincipalResources(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.policySvr.GetPrincipalResources(ctx, queryParams))
}
