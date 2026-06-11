package discover

import (
	"context"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addTrafficGovernanceRuleAccess(ws *restful.WebService) {
	h.addTrafficSecurityAccess(ws)
	h.addTrafficMirrorAccess(ws)
	h.addTrafficMockAccess(ws)
}

func (h *HTTPServer) addTrafficSecurityAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/traffic/security").To(h.GetTrafficSecurityRules))
	ws.Route(ws.POST("/traffic/security").To(h.CreateTrafficSecurityRules))
	ws.Route(ws.PUT("/traffic/security").To(h.UpdateTrafficSecurityRules))
	ws.Route(ws.GET("/traffic/security/detail").To(h.GetOneTrafficSecurityRule))
	ws.Route(ws.POST("/traffic/security/delete").To(h.DeleteTrafficSecurityRules))
	ws.Route(ws.GET("/traffic/security/releases").To(h.GetPublishTrafficSecurityRules))
	ws.Route(ws.POST("/traffic/security/releases").To(h.PublishTrafficGovernanceRules))
	ws.Route(ws.POST("/traffic/security/releases/delete").To(h.DeleteTrafficSecurityReleases))
	ws.Route(ws.PUT("/traffic/security/releases/stopbeta").To(h.StopbetaTrafficSecurityRules))
}

func (h *HTTPServer) addTrafficMirrorAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/traffic/mirrors").To(h.GetTrafficMirrorRules))
	ws.Route(ws.POST("/traffic/mirrors").To(h.CreateTrafficMirrorRules))
	ws.Route(ws.PUT("/traffic/mirrors").To(h.UpdateTrafficMirrorRules))
	ws.Route(ws.GET("/traffic/mirrors/detail").To(h.GetOneTrafficMirrorRule))
	ws.Route(ws.POST("/traffic/mirrors/delete").To(h.DeleteTrafficMirrorRules))
	ws.Route(ws.GET("/traffic/mirrors/releases").To(h.GetPublishTrafficMirrorRules))
	ws.Route(ws.POST("/traffic/mirrors/releases").To(h.PublishTrafficGovernanceRules))
	ws.Route(ws.POST("/traffic/mirrors/releases/delete").To(h.DeleteTrafficMirrorReleases))
	ws.Route(ws.PUT("/traffic/mirrors/releases/stopbeta").To(h.StopbetaTrafficMirrorRules))
}

func (h *HTTPServer) addTrafficMockAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/traffic/mocks").To(h.GetTrafficMockRules))
	ws.Route(ws.POST("/traffic/mocks").To(h.CreateTrafficMockRules))
	ws.Route(ws.PUT("/traffic/mocks").To(h.UpdateTrafficMockRules))
	ws.Route(ws.GET("/traffic/mocks/detail").To(h.GetOneTrafficMockRule))
	ws.Route(ws.POST("/traffic/mocks/delete").To(h.DeleteTrafficMockRules))
	ws.Route(ws.GET("/traffic/mocks/releases").To(h.GetPublishTrafficMockRules))
	ws.Route(ws.POST("/traffic/mocks/releases").To(h.PublishTrafficGovernanceRules))
	ws.Route(ws.POST("/traffic/mocks/releases/delete").To(h.DeleteTrafficMockReleases))
	ws.Route(ws.PUT("/traffic/mocks/releases/stopbeta").To(h.StopbetaTrafficMockRules))
}

func (h *HTTPServer) CreateTrafficSecurityRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficSecurityRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.TrafficSecurityRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.CreateTrafficSecurityRules(ctx, rules))
}

func (h *HTTPServer) UpdateTrafficSecurityRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficSecurityRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.TrafficSecurityRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.UpdateTrafficSecurityRules(ctx, rules))
}

func (h *HTTPServer) DeleteTrafficSecurityRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficSecurityRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apisecurity.TrafficSecurityRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteTrafficSecurityRules(ctx, rules))
}

func (h *HTTPServer) GetTrafficSecurityRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.ruleServer.GetTrafficSecurityRules(handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req)))
}

func (h *HTTPServer) GetOneTrafficSecurityRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	params := httpcommon.ParseQueryParams(req)
	handler.WriteHeaderAndProto(h.ruleServer.GetOneTrafficSecurityRule(handler.ParseHeaderContext(), &apisecurity.TrafficSecurityRule{Id: params["id"]}))
}

func (h *HTTPServer) GetPublishTrafficSecurityRules(req *restful.Request, rsp *restful.Response) {
	h.getTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficSecurityRules)
}

func (h *HTTPServer) CreateTrafficMirrorRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficMirrorRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.TrafficMirror{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.CreateTrafficMirrorRules(ctx, rules))
}

func (h *HTTPServer) UpdateTrafficMirrorRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficMirrorRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.TrafficMirror{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.UpdateTrafficMirrorRules(ctx, rules))
}

func (h *HTTPServer) DeleteTrafficMirrorRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficMirrorRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.TrafficMirror{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteTrafficMirrorRules(ctx, rules))
}

func (h *HTTPServer) GetTrafficMirrorRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.ruleServer.GetTrafficMirrorRules(handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req)))
}

func (h *HTTPServer) GetOneTrafficMirrorRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	params := httpcommon.ParseQueryParams(req)
	handler.WriteHeaderAndProto(h.ruleServer.GetOneTrafficMirrorRule(handler.ParseHeaderContext(), &apitraffic.TrafficMirror{Id: params["id"]}))
}

func (h *HTTPServer) GetPublishTrafficMirrorRules(req *restful.Request, rsp *restful.Response) {
	h.getTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficMirrorRules)
}

func (h *HTTPServer) CreateTrafficMockRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficMockRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.TrafficMock{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.CreateTrafficMockRules(ctx, rules))
}

func (h *HTTPServer) UpdateTrafficMockRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficMockRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.TrafficMock{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.UpdateTrafficMockRules(ctx, rules))
}

func (h *HTTPServer) DeleteTrafficMockRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var rules TrafficMockRuleAttr
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.TrafficMock{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteTrafficMockRules(ctx, rules))
}

func (h *HTTPServer) GetTrafficMockRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.ruleServer.GetTrafficMockRules(handler.ParseHeaderContext(), httpcommon.ParseQueryParams(req)))
}

func (h *HTTPServer) GetOneTrafficMockRule(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	params := httpcommon.ParseQueryParams(req)
	handler.WriteHeaderAndProto(h.ruleServer.GetOneTrafficMockRule(handler.ParseHeaderContext(), &apitraffic.TrafficMock{Id: params["id"]}))
}

func (h *HTTPServer) GetPublishTrafficMockRules(req *restful.Request, rsp *restful.Response) {
	h.getTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficMockRules)
}

func (h *HTTPServer) getTrafficGovernanceReleases(req *restful.Request, rsp *restful.Response, resource apimodel.RuleRelease_RuleType) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = resource.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

func (h *HTTPServer) PublishTrafficGovernanceRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
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

func (h *HTTPServer) StopbetaTrafficSecurityRules(req *restful.Request, rsp *restful.Response) {
	h.changeTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficSecurityRules, h.ruleServer.StopbetaGovernanceRules)
}

func (h *HTTPServer) StopbetaTrafficMirrorRules(req *restful.Request, rsp *restful.Response) {
	h.changeTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficMirrorRules, h.ruleServer.StopbetaGovernanceRules)
}

func (h *HTTPServer) StopbetaTrafficMockRules(req *restful.Request, rsp *restful.Response) {
	h.changeTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficMockRules, h.ruleServer.StopbetaGovernanceRules)
}

func (h *HTTPServer) DeleteTrafficSecurityReleases(req *restful.Request, rsp *restful.Response) {
	h.changeTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficSecurityRules, h.ruleServer.DeleteGovernanceRules)
}

func (h *HTTPServer) DeleteTrafficMirrorReleases(req *restful.Request, rsp *restful.Response) {
	h.changeTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficMirrorRules, h.ruleServer.DeleteGovernanceRules)
}

func (h *HTTPServer) DeleteTrafficMockReleases(req *restful.Request, rsp *restful.Response) {
	h.changeTrafficGovernanceReleases(req, rsp, apimodel.RuleRelease_TrafficMockRules, h.ruleServer.DeleteGovernanceRules)
}

func (h *HTTPServer) changeTrafficGovernanceReleases(
	req *restful.Request,
	rsp *restful.Response,
	resource apimodel.RuleRelease_RuleType,
	action func(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse,
) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{Resource: resource}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	for i := range rules {
		rules[i].Resource = resource
	}
	handler.WriteHeaderAndProto(action(ctx, rules))
}
