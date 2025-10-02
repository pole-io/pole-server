package discover

import (
	"io"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

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
	ws.Route(docs.EnrichGetRouterRuleApiDocs(ws.GET("/routings/detail").To(h.GetOneRouterRule)))
	ws.Route(docs.EnrichEnableRouterRuleApiDocs(ws.GET("/routings/releases").To(h.GetPublishRouterRules)))
	ws.Route(docs.EnrichEnableRouterRuleApiDocs(ws.POST("/routings/releases").To(h.PublishRouterRules)))
	ws.Route(docs.EnrichEnableRouterRuleApiDocs(ws.POST("/routings/releases/delete").To(h.DeleteRouterReleases)))
	ws.Route(docs.EnrichEnableRouterRuleApiDocs(ws.PUT("/routings/releases/rollback").To(h.RollbackRouterRules)))
	ws.Route(docs.EnrichEnableRouterRuleApiDocs(ws.PUT("/routings/releases/stopbeta").To(h.StopbetaRouterRules)))
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

	ret := h.ruleServer.CreateRouterRules(ctx, routings)
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

	ret := h.ruleServer.DeleteRouterRules(ctx, routings)
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
	handler.WriteHeaderAndProto(h.ruleServer.UpdateRouterRules(ctx, routings))
}

// GetRoutings 查询规则路由
func (h *HTTPServer) GetRoutings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	handler.WriteHeaderAndProto(h.ruleServer.QueryRouterRules(handler.ParseHeaderContext(), queryParams))
}

// GetOneRouterRule 查询某条详细的路由规则
func (h *HTTPServer) GetOneRouterRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	msg := &apitraffic.RouteRule{
		Id: queryParams["id"],
	}
	handler.WriteHeaderAndProto(h.ruleServer.GetOneRouterRule(handler.ParseHeaderContext(), msg))
}

// GetPublishRouterRules 获取已发布的规则路由
func (h *HTTPServer) GetPublishRouterRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = apimodel.RuleRelease_RouteRules.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

// PublishRouterRules 启用规则路由
func (h *HTTPServer) PublishRouterRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.PublishGovernanceRules(ctx, rules))
}

// RollbackRouterRules 回滚规则路由
func (h *HTTPServer) RollbackRouterRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.RollbackGovernanceRules(ctx, rules))
}

// StopbetaRouterRules 取消灰度
func (h *HTTPServer) StopbetaRouterRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.StopbetaGovernanceRules(ctx, rules))
}

// DeleteRouterReleases 取消灰度
func (h *HTTPServer) DeleteRouterReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(apiv1.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteGovernanceRules(ctx, rules))
}
