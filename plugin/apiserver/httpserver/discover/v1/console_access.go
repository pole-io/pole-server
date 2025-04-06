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

package v1

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	apiv1 "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// CreateServices 创建服务
func (h *HTTPServerV1) CreateServices(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) DeleteServices(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) UpdateServices(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) GetAllServices(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) GetServices(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) GetServicesCount(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	ret := h.namingServer.GetServicesCount(handler.ParseHeaderContext())
	handler.WriteHeaderAndProto(ret)
}

// GetServiceToken 获取服务token
func (h *HTTPServerV1) GetServiceToken(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) UpdateServiceToken(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) CreateServiceAlias(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) UpdateServiceAlias(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) DeleteServiceAliases(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServerV1) GetServiceAliases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetServiceAliases(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// CreateInstances 创建服务实例
func (h *HTTPServerV1) CreateInstances(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var instances InstanceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Instance{}
		instances = append(instances, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.CreateInstances(ctx, instances))
}

// DeleteInstances 删除服务实例
func (h *HTTPServerV1) DeleteInstances(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var instances InstanceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Instance{}
		instances = append(instances, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteInstances(ctx, instances)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// DeleteInstancesByHost 根据host删除服务实例
func (h *HTTPServerV1) DeleteInstancesByHost(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var instances InstanceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Instance{}
		instances = append(instances, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteInstancesByHost(ctx, instances)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// UpdateInstances 修改服务实例
func (h *HTTPServerV1) UpdateInstances(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var instances InstanceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Instance{}
		instances = append(instances, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateInstances(ctx, instances)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// UpdateInstancesIsolate 修改服务实例的隔离状态
func (h *HTTPServerV1) UpdateInstancesIsolate(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var instances InstanceArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.Instance{}
		instances = append(instances, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateInstancesIsolate(ctx, instances)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// GetInstances 查询服务实例
func (h *HTTPServerV1) GetInstances(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetInstances(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetInstancesCount 查询服务实例
func (h *HTTPServerV1) GetInstancesCount(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ret := h.namingServer.GetInstancesCount(handler.ParseHeaderContext())
	handler.WriteHeaderAndProto(ret)
}

// GetInstanceLabels 查询某个服务下所有实例的标签信息
func (h *HTTPServerV1) GetInstanceLabels(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ret := h.namingServer.GetInstanceLabels(handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req))
	handler.WriteHeaderAndProto(ret)
}

const (
	deprecatedRoutingV2TypeUrl = "type.googleapis.com/v2."
	newRoutingV2TypeUrl        = "type.googleapis.com/v1."
)

func (h *HTTPServerV1) replaceV2TypeUrl(req *restful.Request) (string, error) {
	requestBytes, err := io.ReadAll(req.Request.Body)
	if err != nil {
		return "", err
	}
	requestText := strings.ReplaceAll(string(requestBytes), deprecatedRoutingV2TypeUrl, newRoutingV2TypeUrl)
	return requestText, nil
}

// CreateRoutings 创建规则路由
func (h *HTTPServerV1) CreateRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	requestText, err := h.replaceV2TypeUrl(req)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	var routings RouterArr
	ctx, err := handler.ParseArrayByText(func() proto.Message {
		msg := &apitraffic.RouteRule{}
		routings = append(routings, msg)
		return msg
	}, requestText)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.CreateRoutingConfigs(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// DeleteRoutings 删除规则路由
func (h *HTTPServerV1) DeleteRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	requestText, err := h.replaceV2TypeUrl(req)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	var routings RouterArr
	ctx, err := handler.ParseArrayByText(func() proto.Message {
		msg := &apitraffic.RouteRule{}
		routings = append(routings, msg)
		return msg
	}, requestText)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteRoutingConfigs(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// UpdateRoutings 修改规则路由
func (h *HTTPServerV1) UpdateRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	requestText, err := h.replaceV2TypeUrl(req)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	var routings RouterArr
	ctx, err := handler.ParseArrayByText(func() proto.Message {
		msg := &apitraffic.RouteRule{}
		routings = append(routings, msg)
		return msg
	}, requestText)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateRoutingConfigs(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// GetRoutings 查询规则路由
func (h *HTTPServerV1) GetRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.QueryRoutingConfigs(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// EnableRoutings 启用规则路由
func (h *HTTPServerV1) EnableRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	requestText, err := h.replaceV2TypeUrl(req)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	var routings RouterArr
	ctx, err := handler.ParseArrayByText(func() proto.Message {
		msg := &apitraffic.RouteRule{}
		routings = append(routings, msg)
		return msg
	}, requestText)
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.EnableRoutings(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// CreateRateLimits 创建限流规则
func (h *HTTPServerV1) CreateRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var rateLimits RateLimitArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.Rule{}
		rateLimits = append(rateLimits, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.CreateRateLimits(ctx, rateLimits))
}

// DeleteRateLimits 删除限流规则
func (h *HTTPServerV1) DeleteRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var rateLimits RateLimitArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.Rule{}
		rateLimits = append(rateLimits, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteRateLimits(ctx, rateLimits)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
}

// EnableRateLimits 激活限流规则
func (h *HTTPServerV1) EnableRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	var rateLimits RateLimitArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.Rule{}
		rateLimits = append(rateLimits, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	ret := h.namingServer.EnableRateLimits(ctx, rateLimits)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// UpdateRateLimits 修改限流规则
func (h *HTTPServerV1) UpdateRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var rateLimits RateLimitArr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.Rule{}
		rateLimits = append(rateLimits, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateRateLimits(ctx, rateLimits)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// GetRateLimits 查询限流规则
func (h *HTTPServerV1) GetRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetRateLimits(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetServiceOwner 根据服务获取服务负责人
func (h *HTTPServerV1) GetServiceOwner(req *restful.Request, rsp *restful.Response) {
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

// CreateCircuitBreakerRules create the circuitbreaker rues
func (h *HTTPServerV1) CreateCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var circuitBreakerRules CircuitBreakerRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.CircuitBreakerRule{}
		circuitBreakerRules = append(circuitBreakerRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.CreateCircuitBreakerRules(ctx, circuitBreakerRules))
}

// DeleteCircuitBreakerRules delete the circuitbreaker rues
func (h *HTTPServerV1) DeleteCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var circuitBreakerRules CircuitBreakerRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.CircuitBreakerRule{}
		circuitBreakerRules = append(circuitBreakerRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
}

// EnableCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServerV1) EnableCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	var circuitBreakerRules CircuitBreakerRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.CircuitBreakerRule{}
		circuitBreakerRules = append(circuitBreakerRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	ret := h.namingServer.EnableCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// UpdateCircuitBreakerRules update the circuitbreaker rues
func (h *HTTPServerV1) UpdateCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var circuitBreakerRules CircuitBreakerRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.CircuitBreakerRule{}
		circuitBreakerRules = append(circuitBreakerRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// GetCircuitBreakerRules query the circuitbreaker rues
func (h *HTTPServerV1) GetCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetCircuitBreakerRules(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// CreateFaultDetectRules create the fault detect rues
func (h *HTTPServerV1) CreateFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var faultDetectRules FaultDetectRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.FaultDetectRule{}
		faultDetectRules = append(faultDetectRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.namingServer.CreateFaultDetectRules(ctx, faultDetectRules))
}

// DeleteFaultDetectRules delete the fault detect rues
func (h *HTTPServerV1) DeleteFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var faultDetectRules FaultDetectRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.FaultDetectRule{}
		faultDetectRules = append(faultDetectRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteFaultDetectRules(ctx, faultDetectRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
}

// UpdateFaultDetectRules update the fault detect rues
func (h *HTTPServerV1) UpdateFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var faultDetectRules FaultDetectRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apifault.FaultDetectRule{}
		faultDetectRules = append(faultDetectRules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateFaultDetectRules(ctx, faultDetectRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// GetFaultDetectRules query the fault detect rues
func (h *HTTPServerV1) GetFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetFaultDetectRules(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetServiceContracts 查询服务契约
func (h *HTTPServerV1) GetServiceContracts(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()
	ret := h.namingServer.GetServiceContracts(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetServiceContractVersions 查询服务契约
func (h *HTTPServerV1) GetServiceContractVersions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()
	ret := h.namingServer.GetServiceContractVersions(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// DeleteServiceContracts 删除服务契约
func (h *HTTPServerV1) DeleteServiceContracts(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	contracts := make([]*apiservice.ServiceContract, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.ServiceContract{}
		contracts = append(contracts, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteServiceContracts(ctx, contracts)
	handler.WriteHeaderAndProto(ret)
}

// CreateServiceContract 创建服务契约
func (h *HTTPServerV1) CreateServiceContract(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	contracts := make([]*apiservice.ServiceContract, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiservice.ServiceContract{}
		contracts = append(contracts, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.CreateServiceContracts(ctx, contracts)
	handler.WriteHeaderAndProto(ret)
}

// CreateServiceContractInterfaces 创建服务契约详情
func (h *HTTPServerV1) CreateServiceContractInterfaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	msg := &apiservice.ServiceContract{}
	ctx, err := handler.Parse(msg)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	ret := h.namingServer.CreateServiceContractInterfaces(ctx, msg, apiservice.InterfaceDescriptor_Manual)
	handler.WriteHeaderAndProto(ret)
}

// AppendServiceContractInterfaces 追加服务契约详情
func (h *HTTPServerV1) AppendServiceContractInterfaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	msg := &apiservice.ServiceContract{}
	ctx, err := handler.Parse(msg)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.AppendServiceContractInterfaces(ctx, msg, apiservice.InterfaceDescriptor_Manual)
	handler.WriteHeaderAndProto(ret)
}

// DeleteServiceContractInterfaces 删除服务契约详情
func (h *HTTPServerV1) DeleteServiceContractInterfaces(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	msg := &apiservice.ServiceContract{}
	ctx, err := handler.Parse(msg)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteServiceContractInterfaces(ctx, msg)
	handler.WriteHeaderAndProto(ret)
}

// CreateLaneGroups 批量创建泳道组
func (h *HTTPServerV1) CreateLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	groups := make([]*apitraffic.LaneGroup, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.CreateLaneGroups(ctx, groups)
	handler.WriteHeaderAndProto(ret)
}

// UpdateLaneGroups 批量更新泳道组
func (h *HTTPServerV1) UpdateLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	groups := make([]*apitraffic.LaneGroup, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.UpdateLaneGroups(ctx, groups)
	handler.WriteHeaderAndProto(ret)
}

// DeleteLaneGroups 批量删除泳道组
func (h *HTTPServerV1) DeleteLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	groups := make([]*apitraffic.LaneGroup, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.namingServer.DeleteLaneGroups(ctx, groups)
	handler.WriteHeaderAndProto(ret)
}

// GetLaneGroups 批量删除泳道组
func (h *HTTPServerV1) GetLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()
	ret := h.namingServer.GetLaneGroups(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}
