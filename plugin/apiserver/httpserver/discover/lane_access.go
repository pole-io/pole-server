package discover

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addLaneRuleAccess 泳道规则
func (h *HTTPServer) addLaneRuleAccess(ws *restful.WebService) {
	ws.Route(ws.POST("/lane/groups").To(h.CreateLaneGroups))
	ws.Route(ws.POST("/lane/groups/delete").To(h.DeleteLaneGroups))
	ws.Route(ws.PUT("/lane/groups").To(h.UpdateLaneGroups))
	ws.Route(ws.GET("/lane/groups").To(h.GetLaneGroups))
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
