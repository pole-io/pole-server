package discover

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addRateLimitRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateRateLimitsApiDocs(ws.POST("/ratelimits").To(h.CreateRateLimits)))
	ws.Route(docs.EnrichDeleteRateLimitsApiDocs(ws.POST("/ratelimits/delete").To(h.DeleteRateLimits)))
	ws.Route(docs.EnrichUpdateRateLimitsApiDocs(ws.PUT("/ratelimits").To(h.UpdateRateLimits)))
	ws.Route(docs.EnrichGetRateLimitsApiDocs(ws.GET("/ratelimits").To(h.GetRateLimits)))
	ws.Route(docs.EnrichGetRateLimitsApiDocs(ws.GET("/ratelimits/detail").To(h.GetOneRateLimitRule)))
	ws.Route(docs.EnrichEnableRateLimitsApiDocs(ws.GET("/ratelimits/releases").To(h.GetPublishRateLimits)))
	ws.Route(docs.EnrichEnableRateLimitsApiDocs(ws.POST("/ratelimits/releases").To(h.DeleteRateLimitReleases)))
	ws.Route(docs.EnrichEnableRateLimitsApiDocs(ws.POST("/ratelimits/releases/delete").To(h.PublishRateLimits)))
	ws.Route(docs.EnrichEnableRateLimitsApiDocs(ws.PUT("/ratelimits/releases/rollback").To(h.RollbackRateLimits)))
	ws.Route(docs.EnrichEnableRateLimitsApiDocs(ws.PUT("/ratelimits/releases/stopbeta").To(h.StopbetaRateLimits)))
}

// CreateRateLimits 创建限流规则
func (h *HTTPServer) CreateRateLimits(req *restful.Request, rsp *restful.Response) {
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

	handler.WriteHeaderAndProto(h.ruleServer.CreateRateLimits(ctx, rateLimits))
}

// DeleteRateLimits 删除限流规则
func (h *HTTPServer) DeleteRateLimits(req *restful.Request, rsp *restful.Response) {
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

	ret := h.ruleServer.DeleteRateLimits(ctx, rateLimits)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
}

// UpdateRateLimits 修改限流规则
func (h *HTTPServer) UpdateRateLimits(req *restful.Request, rsp *restful.Response) {
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

	ret := h.ruleServer.UpdateRateLimits(ctx, rateLimits)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// GetRateLimits 查询限流规则
func (h *HTTPServer) GetRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.ruleServer.GetRateLimits(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetOneRateLimitRule 查询某条详细的限流规则
func (h *HTTPServer) GetOneRateLimitRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	msg := &apitraffic.Rule{
		Id: protobuf.NewStringValue(queryParams["id"]),
	}
	handler.WriteHeaderAndProto(h.ruleServer.GetOneRateLimitRule(handler.ParseHeaderContext(), msg))
}

// GetPublishRateLimits 查询已发布限流规则
func (h *HTTPServer) GetPublishRateLimits(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = apimodel.RuleRelease_RateLimitRules.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

// PublishRateLimits 激活限流规则
func (h *HTTPServer) PublishRateLimits(req *restful.Request, rsp *restful.Response) {
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
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.PublishGovernanceRules(ctx, rules))
}

// RollbackRateLimits 激活限流规则
func (h *HTTPServer) RollbackRateLimits(req *restful.Request, rsp *restful.Response) {
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
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.RollbackGovernanceRules(ctx, rules))
}

// StopbetaRateLimits 激活限流规则
func (h *HTTPServer) StopbetaRateLimits(req *restful.Request, rsp *restful.Response) {
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
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.StopbetaGovernanceRules(ctx, rules))
}

// DeleteRateLimitReleases 激活限流规则
func (h *HTTPServer) DeleteRateLimitReleases(req *restful.Request, rsp *restful.Response) {
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
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteGovernanceRules(ctx, rules))
}
