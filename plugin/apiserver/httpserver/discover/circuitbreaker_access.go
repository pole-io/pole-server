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
	ws.Route(docs.EnrichGetCircuitBreakerRulesApiDocs(
		ws.GET("/circuitbreaker/rules").To(h.GetCircuitBreakerRules)))
	ws.Route(docs.EnrichCreateCircuitBreakerRulesApiDocs(
		ws.POST("/circuitbreaker/rules").To(h.CreateCircuitBreakerRules)))
	ws.Route(docs.EnrichUpdateCircuitBreakerRulesApiDocs(
		ws.PUT("/circuitbreaker/rules").To(h.UpdateCircuitBreakerRules)))
	ws.Route(docs.EnrichDeleteCircuitBreakerRulesApiDocs(
		ws.POST("/circuitbreaker/rules/delete").To(h.DeleteCircuitBreakerRules)))
	ws.Route(docs.EnrichEnableCircuitBreakerRulesApiDocs(
		ws.PUT("/circuitbreaker/rules/enable").To(h.EnableCircuitBreakerRules)))
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

	ret := h.ruleServer.DeleteCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
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
	ret := h.ruleServer.EnableCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
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

	ret := h.ruleServer.UpdateCircuitBreakerRules(ctx, circuitBreakerRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
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
