package skill

import (
	"github.com/emicklei/go-restful/v3"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// CreateSkillSubscriptions creates one or more skill subscriptions
func (h *HTTPServer) CreateSkillSubscriptions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var subs SkillSubscriptionArr
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &SkillSubscription{}
		subs = append(subs, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	ret := h.skillServer.CreateSkillSubscriptions(ctx, subs.ToAIType())
	handler.WriteHeaderAndProto(ret)
}

// DeleteSkillSubscriptions deletes one or more skill subscriptions
func (h *HTTPServer) DeleteSkillSubscriptions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	type idObj struct {
		ID string `json:"id"`
	}
	var idObjs []idObj
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &idObj{}
		idObjs = append(idObjs, *msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	var ids []string
	for _, obj := range idObjs {
		ids = append(ids, obj.ID)
	}

	ret := h.skillServer.DeleteSkillSubscriptions(ctx, ids)
	handler.WriteHeaderAndProto(ret)
}

// GetSkillSubscriptions queries skill subscriptions with filters
func (h *HTTPServer) GetSkillSubscriptions(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	if queryParams["skill_name"] == "" && queryParams["client_id"] == "" {
		handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
			apimodel.Code_BadRequest, "skill_name or client_id is required"))
		return
	}

	ret := h.skillServer.GetSkillSubscriptions(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetSkillSubscriptionsByClient gets skill subscriptions by client ID
func (h *HTTPServer) GetSkillSubscriptionsByClient(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	clientID := req.PathParameter("client_id")
	if clientID == "" {
		clientID = queryParams["client_id"]
	}
	if clientID == "" {
		handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
			apimodel.Code_BadRequest, "client_id is required"))
		return
	}

	subs, err := h.skillServer.GetSkillSubscriptionsByClient(ctx, clientID)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
			apimodel.Code_ExecuteException, err.Error()))
		return
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = uint32(len(subs))
	resp.Size = uint32(len(subs))
	handler.WriteHeaderAndProto(resp)
}
