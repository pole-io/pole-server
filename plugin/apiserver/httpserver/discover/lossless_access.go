package discover

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func (h *HTTPServer) addLossLessRuleAccess(ws *restful.WebService) {
	ws.Route((ws.GET("/lossless").To(h.GetLossLessRules)))
	ws.Route((ws.POST("/lossless").To(h.CreateLossLessRules)))
	ws.Route((ws.PUT("/lossless").To(h.UpdateLossLessRules)))
	ws.Route((ws.GET("/lossless/detail").To(h.GetOneLossLessRule)))
	ws.Route((ws.POST("/lossless/delete").To(h.DeleteLossLessRules)))
	ws.Route((ws.GET("/lossless/releases").To(h.GetPublishLossLessRules)))
	ws.Route((ws.POST("/lossless/releases").To(h.PublishLossLessRules)))
	ws.Route((ws.POST("/lossless/releases/delete").To(h.DeleteLossLessReleases)))
	ws.Route((ws.PUT("/lossless/releases/rollback").To(h.RollbackLossLessRules)))
	ws.Route((ws.PUT("/lossless/releases/stopbeta").To(h.StopbetaLossLessRules)))
}

// CreateLossLessRules create the lossless rules
func (h *HTTPServer) CreateLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var rules LosslessRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LosslessRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.CreateLossLessRules(ctx, rules))
}

// DeleteLossLessRules delete the lossless rules
func (h *HTTPServer) DeleteLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var rules LosslessRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LosslessRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteLossLessRules(ctx, rules))
}

// UpdateLossLessRules update the lossless rules
func (h *HTTPServer) UpdateLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var rules LosslessRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LosslessRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.UpdateLossLessRules(ctx, rules))
}

// GetOneLossLessRule 查询某条熔断规则
func (h *HTTPServer) GetOneLossLessRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	msg := &apitraffic.LosslessRule{
		Id: string(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.ruleServer.GetOneLossLessRule(handler.ParseHeaderContext(), msg))
}

// GetLossLessRules query the lossless rules
func (h *HTTPServer) GetLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.ruleServer.GetLossLessRules(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetPublishLossLessRules 查询已发布的熔断规则
func (h *HTTPServer) GetPublishLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = apimodel.RuleRelease_LosslessRules.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

// PublishLossLessRules enable the lossless rules
func (h *HTTPServer) PublishLossLessRules(req *restful.Request, rsp *restful.Response) {
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

// RollbackLossLessRules enable the lossless rules
func (h *HTTPServer) RollbackLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_LosslessRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.RollbackGovernanceRules(ctx, rules))
}

// StopbetaLossLessRules enable the lossless rules
func (h *HTTPServer) StopbetaLossLessRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_LosslessRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.StopbetaGovernanceRules(ctx, rules))
}

// DeleteLossLessReleases enable the lossless rules
func (h *HTTPServer) DeleteLossLessReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_LosslessRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteGovernanceRules(ctx, rules))
}
