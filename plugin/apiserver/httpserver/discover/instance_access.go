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
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// CreateInstances 创建服务实例
func (h *HTTPServer) CreateInstances(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) DeleteInstances(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) DeleteInstancesByHost(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) UpdateInstances(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) UpdateInstancesIsolate(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) GetInstances(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.namingServer.GetInstances(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetInstancesCount 查询服务实例
func (h *HTTPServer) GetInstancesCount(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ret := h.namingServer.GetInstancesCount(handler.ParseHeaderContext())
	handler.WriteHeaderAndProto(ret)
}

// GetInstanceLabels 查询某个服务下所有实例的标签信息
func (h *HTTPServer) GetInstanceLabels(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ret := h.namingServer.GetInstanceLabels(handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req))
	handler.WriteHeaderAndProto(ret)
}
