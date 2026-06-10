package discover

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addFaultDetectRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichGetFaultDetectRulesApiDocs(ws.GET("/faultdetectors").To(h.GetFaultDetectRules)))
	ws.Route(docs.EnrichCreateFaultDetectRulesApiDocs(ws.POST("/faultdetectors").To(h.CreateFaultDetectRules)))
	ws.Route(docs.EnrichUpdateFaultDetectRulesApiDocs(ws.PUT("/faultdetectors").To(h.UpdateFaultDetectRules)))
	ws.Route(docs.EnrichGetFaultDetectRulesApiDocs(ws.GET("/faultdetectors/detail").To(h.GetOneFaultDetectRule)))
	ws.Route(docs.EnrichDeleteFaultDetectRulesApiDocs(ws.POST("/faultdetectors/delete").To(h.DeleteFaultDetectRules)))
	ws.Route((ws.GET("/faultdetectors/releases").To(h.GetPublishFaultDetectRules)))
	ws.Route((ws.POST("/faultdetectors/releases").To(h.PublishFaultDetectRules)))
	ws.Route((ws.POST("/faultdetectors/releases/delete").To(h.DeleteFaultDetectReleases)))
	ws.Route((ws.PUT("/faultdetectors/releases/rollback").To(h.RollbackFaultDetectRules)))
	ws.Route((ws.PUT("/faultdetectors/releases/stopbeta").To(h.StopbetaFaultDetectRules)))
}

// CreateFaultDetectRules create the fault detect rues
func (h *HTTPServer) CreateFaultDetectRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.CreateFaultDetectRules(ctx, faultDetectRules))
}

// DeleteFaultDetectRules delete the fault detect rues
func (h *HTTPServer) DeleteFaultDetectRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.DeleteFaultDetectRules(ctx, faultDetectRules))
}

// UpdateFaultDetectRules update the fault detect rues
func (h *HTTPServer) UpdateFaultDetectRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.UpdateFaultDetectRules(ctx, faultDetectRules))
}

func (h *HTTPServer) GetOneFaultDetectRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	msg := &apifault.FaultDetectRule{
		Id: string(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.ruleServer.GetOneFaultDetectRule(handler.ParseHeaderContext(), msg))
}

// GetFaultDetectRules query the fault detect rues
func (h *HTTPServer) GetFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.ruleServer.GetFaultDetectRules(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetPublishFaultDetectRules 查询已发布的主动探测规则
func (h *HTTPServer) GetPublishFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = apimodel.RuleRelease_FaultDetectRules.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

// PublishFaultDetectRules update the fault detect rues
func (h *HTTPServer) PublishFaultDetectRules(req *restful.Request, rsp *restful.Response) {
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

// RollbackFaultDetectRules update the fault detect rues
func (h *HTTPServer) RollbackFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_FaultDetectRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.RollbackGovernanceRules(ctx, rules))
}

// StopbetaFaultDetectRules update the fault detect rues
func (h *HTTPServer) StopbetaFaultDetectRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_FaultDetectRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.StopbetaGovernanceRules(ctx, rules))
}

// DeleteFaultDetectReleases update the fault detect rues
func (h *HTTPServer) DeleteFaultDetectReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_FaultDetectRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteGovernanceRules(ctx, rules))
}
