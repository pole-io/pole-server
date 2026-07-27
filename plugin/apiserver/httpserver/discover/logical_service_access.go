package discover

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addLogicalServiceAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/logical-services").To(h.GetLogicalServices))
	ws.Route(ws.POST("/logical-services").To(h.CreateLogicalServices))
	ws.Route(ws.PUT("/logical-services").To(h.UpdateLogicalServices))
	ws.Route(ws.POST("/logical-services/delete").To(h.DeleteLogicalServices))
	ws.Route(ws.GET("/logical-services/environments").To(h.GetLogicalServiceEnvironments))
	ws.Route(ws.GET("/logical-services/unbound-environments").To(h.GetUnboundServiceEnvironments))
	ws.Route(ws.GET("/logical-services/environment-binding").To(h.GetServiceEnvironmentBinding))
	ws.Route(ws.POST("/logical-services/environment-bindings").To(h.BindServiceEnvironment))
	ws.Route(ws.POST("/logical-services/environment-bindings/delete").To(h.UnbindServiceEnvironment))
}

func (h *HTTPServer) CreateLogicalServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var items []*apiservice.LogicalService
	ctx, err := handler.ParseArray(func() proto.Message {
		item := &apiservice.LogicalService{}
		items = append(items, item)
		return item
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.namingServer.CreateLogicalServices(ctx, items))
}

func (h *HTTPServer) UpdateLogicalServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var items []*apiservice.LogicalService
	ctx, err := handler.ParseArray(func() proto.Message {
		item := &apiservice.LogicalService{}
		items = append(items, item)
		return item
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.namingServer.UpdateLogicalServices(ctx, items))
}

func (h *HTTPServer) DeleteLogicalServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var items []*apiservice.LogicalService
	ctx, err := handler.ParseArray(func() proto.Message {
		item := &apiservice.LogicalService{}
		items = append(items, item)
		return item
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.namingServer.DeleteLogicalServices(ctx, items))
}

func (h *HTTPServer) GetLogicalServices(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.namingServer.GetLogicalServices(
		handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req)))
}

func (h *HTTPServer) GetLogicalServiceEnvironments(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.namingServer.GetLogicalServiceEnvironments(
		handler.ParseHeaderContext(), req.QueryParameter("logical_service_id")))
}

func (h *HTTPServer) GetUnboundServiceEnvironments(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.namingServer.GetUnboundServiceEnvironments(
		handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req)))
}

func (h *HTTPServer) GetServiceEnvironmentBinding(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.namingServer.GetServiceEnvironmentBinding(
		handler.ParseHeaderContext(), req.QueryParameter("service_id")))
}

func (h *HTTPServer) BindServiceEnvironment(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiservice.BindServiceEnvironmentRequest{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.namingServer.BindServiceEnvironment(ctx, message))
}

func (h *HTTPServer) UnbindServiceEnvironment(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiservice.UnbindServiceEnvironmentRequest{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.namingServer.UnbindServiceEnvironment(ctx, message))
}
