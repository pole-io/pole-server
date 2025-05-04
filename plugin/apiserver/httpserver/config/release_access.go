package config

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addReleasesRuleAccess 增加默认接口
func (h *HTTPServer) addReleasesRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/release").To(h.GetConfigFileRelease)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/releases").To(h.GetConfigFileReleases)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.POST("/files/releases/delete").To(h.DeleteConfigFileReleases)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/release/versions").To(h.GetConfigFileReleaseVersions)))
	ws.Route(docs.EnrichPublishConfigFileApiDocs(ws.POST("/files/release").To(h.PublishConfigFile)))
	ws.Route(docs.EnrichUpsertAndReleaseConfigFileApiDocs(ws.POST("/files/createandpub").To(h.UpsertAndReleaseConfigFile)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.PUT("/files/releases/rollback").To(h.RollbackConfigFileReleases)))
	ws.Route(docs.EnrichStopBetaReleaseConfigFileApiDocs(ws.POST("/files/releases/stopbeta").To(h.StopGrayConfigFileReleases)))
}

// PublishConfigFile 发布配置文件
func (h *HTTPServer) PublishConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	configFile := &apiconfig.ConfigFileRelease{}
	ctx, err := handler.Parse(configFile)
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file release from request error.",
			zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigFileReleaseResponseWithMessage(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.configServer.PublishConfigFile(ctx, configFile))
}

// RollbackConfigFileReleases 获取配置文件最后一次发布内容
func (h *HTTPServer) RollbackConfigFileReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var releases []*apiconfig.ConfigFileRelease
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileRelease{}
		releases = append(releases, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.RollbackConfigFileReleases(ctx, releases))
}

// DeleteConfigFileReleases
func (h *HTTPServer) DeleteConfigFileReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var releases []*apiconfig.ConfigFileRelease
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileRelease{}
		releases = append(releases, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.DeleteConfigFileReleases(ctx, releases))
}

// GetConfigFileReleaseVersions 获取配置文件最后一次发布内容
func (h *HTTPServer) GetConfigFileReleaseVersions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.configServer.GetConfigFileReleaseVersions(ctx, queryParams))
}

// GetConfigFileReleases 获取配置文件最后一次发布内容
func (h *HTTPServer) GetConfigFileReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.configServer.GetConfigFileReleases(ctx, queryParams))
}

// GetConfigFileRelease 获取配置文件最后一次发布内容
func (h *HTTPServer) GetConfigFileRelease(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	namespace := handler.Request.QueryParameter("namespace")
	group := handler.Request.QueryParameter("group")
	fileName := handler.Request.QueryParameter("file_name")
	name := handler.Request.QueryParameter("release_name")
	// 兼容旧的查询参数
	if fileName == "" {
		fileName = handler.Request.QueryParameter("name")
	}

	fileReq := &apiconfig.ConfigFileRelease{
		Namespace: protobuf.NewStringValue(namespace),
		Group:     protobuf.NewStringValue(group),
		FileName:  protobuf.NewStringValue(fileName),
		Name:      protobuf.NewStringValue(name),
	}

	handler.WriteHeaderAndProto(h.configServer.GetConfigFileRelease(handler.ParseHeaderContext(), fileReq))
}

// GetConfigFileReleaseHistory 获取配置文件发布历史，按照发布时间倒序排序
func (h *HTTPServer) GetConfigFileReleaseHistory(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filters := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	handler.WriteHeaderAndProto(h.configServer.GetConfigFileReleaseHistories(ctx, filters))
}

// GetAllConfigEncryptAlgorithm get all config encrypt algorithm
func (h *HTTPServer) GetAllConfigEncryptAlgorithms(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}
	handler.WriteHeaderAndProto(h.configServer.GetAllConfigEncryptAlgorithms(handler.ParseHeaderContext()))
}

// UpsertAndReleaseConfigFile
func (h *HTTPServer) UpsertAndReleaseConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	configFile := &apiconfig.ConfigFilePublishInfo{}
	ctx, err := handler.Parse(configFile)
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file from request error.",
			utils.RequestID(ctx), zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigResponseWithInfo(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.UpsertAndReleaseConfigFile(ctx, configFile))
}

// StopGrayConfigFileReleases .
func (h *HTTPServer) StopGrayConfigFileReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var releases []*apiconfig.ConfigFileRelease
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileRelease{}
		releases = append(releases, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.StopGrayConfigFileReleases(ctx, releases))
}
