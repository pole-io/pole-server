package skill

import (
	"github.com/emicklei/go-restful/v3"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// CreateSkillVersions creates one or more skill versions
func (h *HTTPServer) CreateSkillVersions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var versions SkillVersionArr
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &SkillVersion{}
		versions = append(versions, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.skillServer.CreateSkillVersions(ctx, versions.ToAIType())
	handler.WriteHeaderAndProto(ret)
}

// DeleteSkillVersions deletes one or more skill versions
func (h *HTTPServer) DeleteSkillVersions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	type versionRef struct {
		ID string `json:"id"`
	}
	var refs []versionRef
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &versionRef{}
		refs = append(refs, *msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	var ids []string
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}

	ret := h.skillServer.DeleteSkillVersions(ctx, ids)
	handler.WriteHeaderAndProto(ret)
}

// GetSkillVersions queries skill versions with filters
func (h *HTTPServer) GetSkillVersions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	if queryParams["skill_id"] == "" && queryParams["skill_name"] == "" {
		handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
			apimodel.Code_BadRequest, "skill_id or skill_name is required"))
		return
	}

	ret := h.skillServer.GetSkillVersions(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// ActivateSkillVersion activates a skill version
func (h *HTTPServer) ActivateSkillVersion(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	type activateReq struct {
		VersionID string `json:"version_id"`
	}
	var reqs []activateReq
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &activateReq{}
		reqs = append(reqs, *msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, r := range reqs {
		if err := h.skillServer.ActivateSkillVersion(ctx, r.VersionID); err != nil {
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	handler.WriteHeaderAndProto(api.FormatBatchWriteResponse(resp))
}
