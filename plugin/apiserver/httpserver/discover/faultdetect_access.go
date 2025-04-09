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

func (h *HTTPServer) addFaultDetectRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichGetFaultDetectRulesApiDocs(
		ws.GET("/faultdetectors").To(h.GetFaultDetectRules)))
	ws.Route(docs.EnrichCreateFaultDetectRulesApiDocs(
		ws.POST("/faultdetectors").To(h.CreateFaultDetectRules)))
	ws.Route(docs.EnrichUpdateFaultDetectRulesApiDocs(
		ws.PUT("/faultdetectors").To(h.UpdateFaultDetectRules)))
	ws.Route(docs.EnrichDeleteFaultDetectRulesApiDocs(
		ws.POST("/faultdetectors/delete").To(h.DeleteFaultDetectRules)))
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

	ret := h.ruleServer.DeleteFaultDetectRules(ctx, faultDetectRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}
	handler.WriteHeaderAndProto(ret)
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

	ret := h.ruleServer.UpdateFaultDetectRules(ctx, faultDetectRules)
	if code := api.CalcCode(ret); code != http.StatusOK {
		handler.WriteHeaderAndProto(ret)
		return
	}

	handler.WriteHeaderAndProto(ret)
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
