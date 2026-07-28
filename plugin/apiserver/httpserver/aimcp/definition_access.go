package aimcp

import (
	"encoding/json"
	"net/http"
	"strconv"

	restful "github.com/emicklei/go-restful/v3"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) ListMCPServerDefinitions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	ctx := handler.ParseHeaderContext()
	if _, err := h.checkMCPServerPermission(
		ctx, authtypes.Read, authtypes.DescribeMCPServers, nil); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error()))
		return
	}
	total, definitions, err := h.storage.ListAIResourceDefinitions(
		aitypes.ResourceKindMCPServer, req.QueryParameter("name"),
		definitionUint32(req.QueryParameter("offset"), 0),
		definitionUint32(req.QueryParameter("limit"), 100))
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, &aitypes.ResourceDefinitionListResponse{
		Code: uint32(apimodel.Code_ExecuteSuccess), Amount: total,
		Size: uint32(len(definitions)), Data: definitions,
	}, restful.MIME_JSON)
}

func (h *HTTPServer) CreateMCPServerDefinitions(req *restful.Request, rsp *restful.Response) {
	h.writeMCPServerDefinitions(req, rsp, true)
}

func (h *HTTPServer) UpdateMCPServerDefinitions(req *restful.Request, rsp *restful.Response) {
	h.writeMCPServerDefinitions(req, rsp, false)
}

func (h *HTTPServer) writeMCPServerDefinitions(
	req *restful.Request, rsp *restful.Response, create bool) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var definitions []*aitypes.ResourceDefinition
	if err := json.NewDecoder(req.Request.Body).Decode(&definitions); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	op, method := authtypes.Modify, authtypes.UpdateMCPServers
	if create {
		op, method = authtypes.Create, authtypes.CreateMCPServers
	}
	if _, err := h.checkMCPServerPermission(handler.ParseHeaderContext(), op, method, nil); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error()))
		return
	}
	for _, definition := range definitions {
		if definition == nil {
			handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_InvalidParameter))
			return
		}
		definition.Kind = aitypes.ResourceKindMCPServer
		if create {
			if definition.ID == "" {
				definition.ID = utils.NewUUID()
			}
			definition.Revision = utils.NewUUID()
			if err := h.storage.CreateAIResourceDefinition(definition); err != nil {
				handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
				return
			}
			continue
		}
		previousRevision := definition.Revision
		definition.Revision = utils.NewUUID()
		if err := h.storage.UpdateAIResourceDefinition(definition, previousRevision); err != nil {
			handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
			return
		}
	}
	handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_ExecuteSuccess))
}

func (h *HTTPServer) DeleteMCPServerDefinitions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var deleteReq aitypes.ResourceDefinitionDeleteRequest
	if err := req.ReadEntity(&deleteReq); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	if _, err := h.checkMCPServerPermission(handler.ParseHeaderContext(),
		authtypes.Delete, authtypes.DeleteMCPServers, nil); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error()))
		return
	}
	for _, id := range deleteReq.DefinitionIDs {
		if err := h.storage.DeleteAIResourceDefinition(aitypes.ResourceKindMCPServer, id); err != nil {
			handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
			return
		}
	}
	handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_ExecuteSuccess))
}

func (h *HTTPServer) ListMCPServerDefinitionEnvironments(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	ctx := handler.ParseHeaderContext()
	authCtx, err := h.checkMCPServerPermission(ctx, authtypes.Read, authtypes.DescribeMCPServers, nil)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error()))
		return
	}
	bindings, err := h.storage.ListAIResourceEnvironmentBindings(
		aitypes.ResourceKindMCPServer, req.QueryParameter("definition_id"))
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
		return
	}
	filtered := make([]*aitypes.EnvironmentBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding != nil && h.canReadMCPServer(authCtx, binding.ResourceID) {
			filtered = append(filtered, binding)
		}
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, &aitypes.EnvironmentBindingListResponse{
		Code: uint32(apimodel.Code_ExecuteSuccess), Amount: uint32(len(filtered)),
		Size: uint32(len(filtered)), Data: filtered,
	}, restful.MIME_JSON)
}

func (h *HTTPServer) GetMCPServerDefinitionBinding(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	resourceID := req.QueryParameter("resource_id")
	if _, err := h.checkMCPServerPermission(handler.ParseHeaderContext(),
		authtypes.Read, authtypes.DescribeMCPServers, []string{resourceID}); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error()))
		return
	}
	binding, err := h.storage.GetAIResourceEnvironmentBinding(aitypes.ResourceKindMCPServer, resourceID)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{
		"code": uint32(apimodel.Code_ExecuteSuccess), "data": binding,
	}, restful.MIME_JSON)
}

func (h *HTTPServer) BindMCPServerDefinitionEnvironment(req *restful.Request, rsp *restful.Response) {
	h.mutateMCPServerDefinitionBinding(req, rsp, true)
}

func (h *HTTPServer) UnbindMCPServerDefinitionEnvironment(req *restful.Request, rsp *restful.Response) {
	h.mutateMCPServerDefinitionBinding(req, rsp, false)
}

func (h *HTTPServer) mutateMCPServerDefinitionBinding(
	req *restful.Request, rsp *restful.Response, bind bool) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	binding := &aitypes.EnvironmentBinding{}
	if err := req.ReadEntity(binding); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	binding.Kind = aitypes.ResourceKindMCPServer
	if _, err := h.checkMCPServerPermission(handler.ParseHeaderContext(),
		authtypes.Modify, authtypes.UpdateMCPServers, []string{binding.ResourceID}); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error()))
		return
	}
	if bind {
		if response := h.validateMCPDefinitionEnvironment(binding.ResourceID); response != nil {
			handler.WriteHeaderAndProto(response)
			return
		}
		if err := h.storage.BindAIResourceEnvironment(binding, utils.NewUUID()); err != nil {
			handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
			return
		}
	} else if err := h.storage.UnbindAIResourceEnvironment(
		binding.Kind, binding.DefinitionID, binding.ResourceID, utils.NewUUID()); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error()))
		return
	}
	handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_ExecuteSuccess))
}

func (h *HTTPServer) validateMCPDefinitionEnvironment(resourceID string) *apimodel.Response {
	server := h.cacheMgr.MCPServer().GetMCPServerByID(resourceID)
	if server == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	namespace := h.cacheMgr.Namespace().GetNamespace(server.GetNamespace())
	if namespace == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	if server.GetNamespace() == "pole-system" ||
		namespace.Kind == apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM {
		return api.NewResponseWithMsg(apimodel.Code_NotAllowedAccess,
			"系统 Namespace 中的 MCP Server 不能绑定业务逻辑定义")
	}
	return nil
}

func definitionUint32(raw string, fallback uint32) uint32 {
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return fallback
	}
	return uint32(value)
}
