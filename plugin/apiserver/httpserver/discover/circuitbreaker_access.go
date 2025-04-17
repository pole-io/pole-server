package discover

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addCircuitBreakerRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichGetCircuitBreakerRulesApiDocs(ws.GET("/circuitbreaker/rules").To(h.GetCircuitBreakerRules)))
	ws.Route(docs.EnrichCreateCircuitBreakerRulesApiDocs(ws.POST("/circuitbreaker/rules").To(h.CreateCircuitBreakerRules)))
	ws.Route(docs.EnrichUpdateCircuitBreakerRulesApiDocs(ws.PUT("/circuitbreaker/rules").To(h.UpdateCircuitBreakerRules)))
	ws.Route(docs.EnrichDeleteCircuitBreakerRulesApiDocs(ws.POST("/circuitbreaker/rules/delete").To(h.DeleteCircuitBreakerRules)))
	ws.Route(docs.EnrichEnableCircuitBreakerRulesApiDocs(ws.POST("/circuitbreaker/rules/publish").To(h.PublishCircuitBreakerRules)))
	ws.Route(docs.EnrichEnableCircuitBreakerRulesApiDocs(ws.PUT("/circuitbreaker/rules/rollback").To(h.RollbackCircuitBreakerRules)))
	ws.Route(docs.EnrichEnableCircuitBreakerRulesApiDocs(ws.PUT("/circuitbreaker/rules/stopbeta").To(h.StopbetaCircuitBreakerRules)))
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

// EnableCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) EnableCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.EnableCircuitBreakerRules(ctx, circuitBreakerRules))
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

// PublishCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) PublishCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	ret := h.ruleServer.PublishCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
}

// RollbackCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) RollbackCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.RollbackCircuitBreakerRules(ctx, circuitBreakerRules))
}

// StopbetaCircuitBreakerRules enable the circuitbreaker rues
func (h *HTTPServer) StopbetaCircuitBreakerRules(req *restful.Request, rsp *restful.Response) {
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
	handler.WriteHeaderAndProto(h.ruleServer.StopbetaCircuitBreakerRules(ctx, circuitBreakerRules))
}
