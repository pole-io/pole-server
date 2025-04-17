package config

import (
	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addGroupRuleAccess 增加默认接口
func (h *HTTPServer) addGroupRuleAccess(ws *restful.WebService) {
	// 配置文件发布
	ws.Route(docs.EnrichCreateConfigFileGroupApiDocs(ws.POST("/groups").To(h.CreateConfigFileGroup)))
	ws.Route(docs.EnrichUpdateConfigFileGroupApiDocs(ws.PUT("/groups").To(h.UpdateConfigFileGroup)))
	ws.Route(docs.EnrichDeleteConfigFileGroupApiDocs(ws.DELETE("/groups").To(h.DeleteConfigFileGroup)))
	ws.Route(docs.EnrichQueryConfigFileGroupsApiDocs(ws.GET("/groups").To(h.QueryConfigFileGroups)))
}

// CreateConfigFileGroup 创建配置文件组
func (h *HTTPServer) CreateConfigFileGroup(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	configFileGroup := &apiconfig.ConfigFileGroup{}
	ctx, err := handler.Parse(configFileGroup)
	requestId := ctx.Value(types.StringContext("request-id"))

	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file group from request error.",
			zap.String("requestId", requestId.(string)),
			zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigFileGroupResponseWithMessage(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.CreateConfigFileGroup(ctx, configFileGroup))
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
func (h *HTTPServer) DeleteConfigFileGroup(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	namespace := handler.Request.QueryParameter("namespace")
	group := handler.Request.QueryParameter("group")

	response := h.configServer.DeleteConfigFileGroup(handler.ParseHeaderContext(), namespace, group)
	handler.WriteHeaderAndProto(response)
}

// UpdateConfigFileGroup 更新配置文件组，只能更新 comment
func (h *HTTPServer) UpdateConfigFileGroup(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	configFileGroup := &apiconfig.ConfigFileGroup{}
	ctx, err := handler.Parse(configFileGroup)
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file group from request error.",
			utils.RequestID(ctx), zap.Error(err))
		handler.WriteHeaderAndProto(api.NewConfigResponseWithInfo(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.UpdateConfigFileGroup(ctx, configFileGroup))
}
