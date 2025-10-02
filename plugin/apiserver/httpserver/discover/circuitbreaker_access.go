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

func (h *HTTPServer) addCircuitBreakerRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichGetCircuitBreakerRulesApiDocs(ws.GET("/circuitbreakers").To(h.GetCircuitBreakerRules)))
	ws.Route(docs.EnrichCreateCircuitBreakerRulesApiDocs(ws.POST("/circuitbreakers").To(h.CreateCircuitBreakerRules)))
	ws.Route(docs.EnrichUpdateCircuitBreakerRulesApiDocs(ws.PUT("/circuitbreakers").To(h.UpdateCircuitBreakerRules)))
	ws.Route(docs.EnrichGetCircuitBreakerRulesApiDocs(ws.GET("/circuitbreakers/detail").To(h.GetOneCircuitBreakerRule)))
	ws.Route(docs.EnrichDeleteCircuitBreakerRulesApiDocs(ws.POST("/circuitbreakers/delete").To(h.DeleteCircuitBreakerRules)))
	ws.Route((ws.GET("/circuitbreakers/releases").To(h.GetPublishCircuitBreakerRules)))
	ws.Route((ws.POST("/circuitbreakers/releases").To(h.PublishCircuitBreakerRules)))
	ws.Route((ws.POST("/circuitbreakers/releases/delete").To(h.DeleteCircuitBreakerReleases)))
	ws.Route((ws.PUT("/circuitbreakers/releases/rollback").To(h.RollbackCircuitBreakerRules)))
	ws.Route((ws.PUT("/circuitbreakers/releases/stopbeta").To(h.StopbetaCircuitBreakerRules)))
}

// CreateCircuitBreakerRules create the circuitbreaker rues
func (h *HTTPServer) CreateCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.CreateCircuitBreakerRules(ctx, circuitBreakerRules))
}

// DeleteCircuitBreakerRules delete the circuitbreaker rues
func (h *HTTPServer) DeleteCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.DeleteCircuitBreakerRules(ctx, circuitBreakerRules))
}

// UpdateCircuitBreakerRules update the circuitbreaker rues
func (h *HTTPServer) UpdateCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.UpdateCircuitBreakerRules(ctx, circuitBreakerRules))
}

// GetOneCircuitBreakerRule 查询某条熔断规则
func (h *HTTPServer) GetOneCircuitBreakerRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	msg := &apifault.CircuitBreakerRule{
		Id: string(queryParams["id"]),
	}

	handler.WriteHeaderAndProto(h.ruleServer.GetOneCircuitBreakerRule(handler.ParseHeaderContext(), msg))
}

// GetCircuitBreakerRules query the circuitbreaker rues
func (h *HTTPServer) GetCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ret := h.ruleServer.GetCircuitBreakerRules(handler.ParseHeaderContext(), queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetPublishCircuitBreakerRules 查询已发布的熔断规则
func (h *HTTPServer) GetPublishCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = apimodel.RuleRelease_CircuitBreakerRules.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

// PublishCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) PublishCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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

// RollbackCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) RollbackCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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

// StopbetaCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) StopbetaCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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

// DeleteCircuitBreakerReleases enable the circuitbreaker rues
func (h *HTTPServer) DeleteCircuitBreakerReleases(req *restful.Request, rsp *restful.Response) {
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
