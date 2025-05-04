package config

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addGroupRuleAccess 增加默认接口
func (h *HTTPServer) addGroupRuleAccess(ws *restful.WebService) {
	// 配置文件发布
	ws.Route(docs.EnrichCreateConfigFileGroupApiDocs(ws.POST("/groups").To(h.CreateConfigFileGroups)))
	ws.Route(docs.EnrichUpdateConfigFileGroupApiDocs(ws.PUT("/groups").To(h.UpdateConfigFileGroups)))
	ws.Route(docs.EnrichDeleteConfigFileGroupApiDocs(ws.POST("/groups/delette").To(h.DeleteConfigFileGroups)))
	ws.Route(docs.EnrichQueryConfigFileGroupsApiDocs(ws.GET("/groups").To(h.QueryConfigFileGroups)))
}

// CreateConfigFileGroups 创建配置文件组
func (h *HTTPServer) CreateConfigFileGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var groups []*apiconfig.ConfigFileGroup
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file group from request error.",
			utils.RequestID(ctx),
			zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigFileGroupResponseWithMessage(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.CreateConfigFileGroups(ctx, groups))
}

// QueryConfigFileGroups 查询配置文件组，group 模糊搜索
func (h *HTTPServer) QueryConfigFileGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filter := httpcommon.ParseQueryParams(req)
	response := h.configServer.QueryConfigFileGroups(handler.ParseHeaderContext(), filter)

	handler.WriteHeaderAndProto(response)
}

// DeleteConfigFileGroup 删除配置文件组
func (h *HTTPServer) DeleteConfigFileGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var groups []*apiconfig.ConfigFileGroup
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file group from request error.",
			utils.RequestID(ctx),
			zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigFileGroupResponseWithMessage(apimodel.Code_ParseException, err.Error()))
		return
	}

	response := h.configServer.DeleteConfigFileGroups(handler.ParseHeaderContext(), groups)
	handler.WriteHeaderAndProto(response)
}

// UpdateConfigFileGroups 更新配置文件组，只能更新 comment
func (h *HTTPServer) UpdateConfigFileGroups(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var groups []*apiconfig.ConfigFileGroup
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileGroup{}
		groups = append(groups, msg)
		return msg
	})
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file group from request error.",
			utils.RequestID(ctx),
			zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigFileGroupResponseWithMessage(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.UpdateConfigFileGroups(ctx, groups))
}
