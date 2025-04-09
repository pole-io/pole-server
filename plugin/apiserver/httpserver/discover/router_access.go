package discover

import (
	"io"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	apiv1 "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addRoutingRuleAccess 增加默认接口
func (h *HTTPServer) addRoutingRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateRouterRuleApiDocs(ws.POST("/routings").To(h.CreateRoutings)))
	ws.Route(docs.EnrichDeleteRouterRuleApiDocs(ws.POST("/routings/delete").To(h.DeleteRoutings)))
	ws.Route(docs.EnrichUpdateRouterRuleApiDocs(ws.PUT("/routings").To(h.UpdateRoutings)))
	ws.Route(docs.EnrichGetRouterRuleApiDocs(ws.GET("/routings").To(h.GetRoutings)))
	ws.Route(docs.EnrichEnableRouterRuleApiDocs(ws.PUT("/routings/enable").To(h.EnableRoutings)))
}

const (
	deprecatedRoutingV2TypeUrl = "type.googleapis.com/v2."
	newRoutingV2TypeUrl        = "type.googleapis.com/v1."
)

func (h *HTTPServer) replaceV2TypeUrl(req *restful.Request) (string, error) {
	requestBytes, err := io.ReadAll(req.Request.Body)
	if err != nil {
		return "", err
	}
	requestText := strings.ReplaceAll(string(requestBytes), deprecatedRoutingV2TypeUrl, newRoutingV2TypeUrl)
	return requestText, nil
}

// CreateRoutings 创建规则路由
func (h *HTTPServer) CreateRoutings(req *restful.Request, rsp *restful.Response) {
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

	ret := h.ruleServer.CreateRoutingConfigs(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// DeleteRoutings 删除规则路由
func (h *HTTPServer) DeleteRoutings(req *restful.Request, rsp *restful.Response) {
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

	ret := h.ruleServer.DeleteRoutingConfigs(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// UpdateRoutings 修改规则路由
func (h *HTTPServer) UpdateRoutings(req *restful.Request, rsp *restful.Response) {
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

	ret := h.ruleServer.UpdateRoutingConfigs(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}

// GetRoutings 查询规则路由
func (h *HTTPServer) GetRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.ruleServer.QueryRoutingConfigs(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// EnableRoutings 启用规则路由
func (h *HTTPServer) EnableRoutings(req *restful.Request, rsp *restful.Response) {
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

	ret := h.ruleServer.EnableRoutings(ctx, routings)
	handler.WriteHeaderAndProto(ret)
}
