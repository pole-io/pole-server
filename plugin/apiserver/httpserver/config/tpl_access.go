package config

import (
	"strconv"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addTemplateRuleAccess 增加默认接口
func (h *HTTPServer) addTemplateRuleAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateConfigFileTemplateApiDocs(ws.POST("/templates").To(h.CreateConfigFileTemplates)))
	ws.Route(docs.EnrichUpdateConfigFileTemplateApiDocs(ws.PUT("/templates").To(h.UpdateConfigFileTemplate)))
	ws.Route(docs.EnrichGetAllConfigFileTemplatesApiDocs(ws.GET("/templates").To(h.GetAllConfigFileTemplates)))
	ws.Route(ws.POST("/templates/preview").To(h.PreviewConfigTemplate))
	ws.Route(ws.POST("/templates/releases").To(h.PublishConfigTemplateRelease))
	ws.Route(ws.GET("/templates/releases").To(h.ListConfigTemplateReleases))
	ws.Route(ws.PUT("/templates/values").To(h.SaveNamespaceTemplateValues))
	ws.Route(ws.GET("/templates/values").To(h.GetNamespaceTemplateValues))
	ws.Route(ws.POST("/templates/values/releases").To(h.PublishNamespaceTemplateValueRelease))
	ws.Route(ws.GET("/templates/values/releases").To(h.ListNamespaceTemplateValueReleases))
	ws.Route(ws.POST("/templates/bindings").To(h.BindConfigFileTemplate))
	ws.Route(ws.GET("/templates/bindings").To(h.ListConfigTemplateBindings))
}

func (h *HTTPServer) PreviewConfigTemplate(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiconfig.RenderPreviewRequest{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewConfigResponse(apimodel.Code_ParseException))
		return
	}
	preview := h.configServer.PreviewConfigTemplate(ctx, message)
	code := apimodel.Code(preview.GetCode())
	if code == 0 {
		code = apimodel.Code_ExecuteSuccess
	}
	handler.WriteHeaderAndProto(api.NewAnyDataResponse(code, preview))
}

func (h *HTTPServer) PublishConfigTemplateRelease(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiconfig.ConfigTemplateRelease{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewConfigResponse(apimodel.Code_ParseException))
		return
	}
	handler.WriteHeaderAndProto(h.configServer.PublishConfigTemplateRelease(ctx, message))
}

func (h *HTTPServer) SaveNamespaceTemplateValues(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiconfig.NamespaceTemplateValues{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewConfigResponse(apimodel.Code_ParseException))
		return
	}
	handler.WriteHeaderAndProto(h.configServer.SaveNamespaceTemplateValues(ctx, message))
}

func (h *HTTPServer) PublishNamespaceTemplateValueRelease(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiconfig.NamespaceTemplateValueRelease{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewConfigResponse(apimodel.Code_ParseException))
		return
	}
	handler.WriteHeaderAndProto(h.configServer.PublishNamespaceTemplateValueRelease(ctx, message))
}

func (h *HTTPServer) BindConfigFileTemplate(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	message := &apiconfig.ConfigFile{}
	ctx, err := handler.Parse(message)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewConfigResponse(apimodel.Code_ParseException))
		return
	}
	handler.WriteHeaderAndProto(h.configServer.BindConfigFileTemplate(ctx, message))
}

func (h *HTTPServer) ListConfigTemplateReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	templateID, _ := strconv.ParseUint(req.QueryParameter("template_id"), 10, 64)
	handler.WriteHeaderAndProto(
		h.configServer.ListConfigTemplateReleases(handler.ParseHeaderContext(), templateID))
}

func (h *HTTPServer) GetNamespaceTemplateValues(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	templateID, _ := strconv.ParseUint(req.QueryParameter("template_id"), 10, 64)
	handler.WriteHeaderAndProto(h.configServer.GetNamespaceTemplateValues(
		handler.ParseHeaderContext(), req.QueryParameter("namespace"), templateID))
}

func (h *HTTPServer) ListNamespaceTemplateValueReleases(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	templateID, _ := strconv.ParseUint(req.QueryParameter("template_id"), 10, 64)
	handler.WriteHeaderAndProto(h.configServer.ListNamespaceTemplateValueReleases(
		handler.ParseHeaderContext(), req.QueryParameter("namespace"), templateID))
}

func (h *HTTPServer) ListConfigTemplateBindings(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	handler.WriteHeaderAndProto(h.configServer.ListConfigTemplateBindings(
		handler.ParseHeaderContext(),
		req.QueryParameter("namespace"),
		req.QueryParameter("group"),
		req.QueryParameter("file_name")))
}

// GetAllConfigFileTemplates get all config file template
func (h *HTTPServer) GetAllConfigFileTemplates(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	response := h.configServer.GetAllConfigFileTemplates(handler.ParseHeaderContext())

	handler.WriteHeaderAndProto(response)
}

// CreateConfigFileTemplate create config file template
func (h *HTTPServer) CreateConfigFileTemplates(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var tpls []*apiconfig.ConfigFileTemplate
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileTemplate{}
		tpls = append(tpls, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.CreateConfigFileTemplates(ctx, tpls))
}

// UpdateConfigFileTemplate create config file template
func (h *HTTPServer) UpdateConfigFileTemplate(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var tpls []*apiconfig.ConfigFileTemplate
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFileTemplate{}
		tpls = append(tpls, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.UpdateConfigFileTemplates(ctx, tpls))
}
