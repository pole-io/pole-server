package discover

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addRateLimitRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateRateLimitsApiDocs(ws.POST("/ratelimits").To(h.CreateRateLimits)))
	ws.Route(docs.EnrichDeleteRateLimitsApiDocs(ws.POST("/ratelimits/delete").To(h.DeleteRateLimits)))
	ws.Route(docs.EnrichUpdateRateLimitsApiDocs(ws.PUT("/ratelimits").To(h.UpdateRateLimits)))
	ws.Route(docs.EnrichGetRateLimitsApiDocs(ws.GET("/ratelimits").To(h.GetRateLimits)))
	ws.Route(docs.EnrichEnableRateLimitsApiDocs(ws.PUT("/ratelimits/enable").To(h.EnableRateLimits)))
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

// EnableRateLimits 激活限流规则
func (h *HTTPServer) EnableRateLimits(req *restful.Request, rsp *restful.Response) {
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
	ret := h.ruleServer.EnableRateLimits(ctx, rateLimits)
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
