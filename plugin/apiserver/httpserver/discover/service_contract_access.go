package discover

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// GetServiceContracts 查询服务契约
func (h *HTTPServer) GetServiceContracts(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) GetServiceContractVersions(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) DeleteServiceContracts(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) CreateServiceContract(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) CreateServiceContractInterfaces(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) AppendServiceContractInterfaces(req *restful.Request, rsp *restful.Response) {
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
func (h *HTTPServer) DeleteServiceContractInterfaces(req *restful.Request, rsp *restful.Response) {
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
