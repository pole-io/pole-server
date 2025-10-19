package discover

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addLaneRuleAccess 泳道规则
func (h *HTTPServer) addLaneRuleAccess(ws *restful.WebService) {
	ws.Route(ws.POST("/lane/groups").To(h.CreateLaneGroups))
	ws.Route(ws.POST("/lane/groups/delete").To(h.DeleteLaneGroups))
	ws.Route(ws.PUT("/lane/groups").To(h.UpdateLaneGroups))
	ws.Route(ws.GET("/lane/groups").To(h.GetLaneGroups))
	ws.Route(ws.GET("/lane/groups/detail").To(h.GetOneLaneGroup))
	ws.Route(ws.POST("/lane/groups/rules").To(h.CreateLaneRules))
	ws.Route(ws.PUT("/lane/groups/rules").To(h.UpdateLaneRules))
	ws.Route(ws.POST("/lane/groups/rules/delete").To(h.DeleteLaneRules))

	ws.Route(ws.GET("/lane/groups/releases").To(h.GetPublishLaneGroups))
	ws.Route(ws.POST("/lane/groups/releases").To(h.PublishLaneGroups))
	ws.Route(ws.POST("/lane/groups/releases/delete").To(h.DeleteLaneGroupReleases))
	ws.Route(ws.PUT("/lane/groups/releases/rollback").To(h.RollbackLaneGroups))
	ws.Route(ws.PUT("/lane/groups/releases/stopbeta").To(h.StopbetaLaneGroups))
}

// CreateLaneGroups 批量创建泳道组
func (h *HTTPServer) CreateLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	groups := make([]*apitraffic.LaneGroup, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.ruleServer.CreateLaneGroups(ctx, groups)
	handler.WriteHeaderAndProto(ret)
}

// UpdateLaneGroups 批量更新泳道组
func (h *HTTPServer) UpdateLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	groups := make([]*apitraffic.LaneGroup, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.ruleServer.UpdateLaneGroups(ctx, groups)
	handler.WriteHeaderAndProto(ret)
}

// DeleteLaneGroups 批量删除泳道组
func (h *HTTPServer) DeleteLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	groups := make([]*apitraffic.LaneGroup, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.ruleServer.DeleteLaneGroups(ctx, groups)
	handler.WriteHeaderAndProto(ret)
}

// GetLaneGroups 批量删除泳道组
func (h *HTTPServer) GetLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()
	ret := h.ruleServer.GetLaneGroups(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetOneLaneGroup 查询某条详细的泳道组
func (h *HTTPServer) GetOneLaneGroup(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	queryParams := httpcommon.ParseQueryParams(req)
	msg := &apitraffic.LaneGroup{
		Id: queryParams["id"],
	}

	handler.WriteHeaderAndProto(h.ruleServer.GetOneLaneGroup(handler.ParseHeaderContext(), msg))
}

// GetPublishLaneGroups 查询已发布的泳道规则
func (h *HTTPServer) GetPublishLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	filter["rule_id"] = filter["id"]
	filter["resource"] = apimodel.RuleRelease_LaneRules.String()
	handler.WriteHeaderAndProto(h.ruleServer.GetRuleReleases(handler.ParseHeaderContext(), filter))
}

// PublishLaneGroups 批量更新泳道组
func (h *HTTPServer) PublishLaneGroups(req *restful.Request, rsp *restful.Response) {
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

// DeleteLaneGroupReleases 批量更新泳道组
func (h *HTTPServer) DeleteLaneGroupReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_LaneRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.DeleteGovernanceRules(ctx, rules))
}

// RollbackLaneGroups 批量更新泳道组
func (h *HTTPServer) RollbackLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_LaneRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.RollbackGovernanceRules(ctx, rules))
}

// StopbetaLaneGroups 批量更新泳道组
func (h *HTTPServer) StopbetaLaneGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apimodel.RuleRelease, 0, 4)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apimodel.RuleRelease{}
		msg.Resource = apimodel.RuleRelease_LaneRules
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.ruleServer.StopbetaGovernanceRules(ctx, rules))
}

func (h *HTTPServer) CreateLaneRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apitraffic.LaneRule, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.ruleServer.CreateLaneRules(ctx, rules))
}

func (h *HTTPServer) UpdateLaneRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apitraffic.LaneRule, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.ruleServer.UpdateLaneRules(ctx, rules))
}

func (h *HTTPServer) DeleteLaneRules(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	rules := make([]*apitraffic.LaneRule, 0)
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apitraffic.LaneRule{}
		rules = append(rules, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.ruleServer.DeleteLaneRules(ctx, rules))
}
