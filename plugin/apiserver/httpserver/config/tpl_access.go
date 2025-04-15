package config

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

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

	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file template from request error.",
			utils.RequestID(ctx),
			zap.String("error", err.Error()))
		handler.WriteHeaderAndProto(api.NewConfigFileTemplateResponseWithMessage(apimodel.Code_ParseException, err.Error()))
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
