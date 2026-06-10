/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 */

package aia2a

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	restful "github.com/emicklei/go-restful/v3"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) ListA2AAgents(req *restful.Request, rsp *restful.Response) {
	query := parseA2AAgentQuery(req.Request.URL.Query())
	count, agents := h.cacheMgr.A2AAgent().Query(query)
	_ = rsp.WriteHeaderAndJson(http.StatusOK, newA2AAgentListResponse(count, agents), restful.MIME_JSON)
}

func (h *HTTPServer) CreateA2AAgents(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	agents, err := readA2AAgents(req)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	for _, agent := range agents {
		if err := h.storage.CreateA2AAgent(agent); err != nil {
			handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			return
		}
	}
	handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_ExecuteSuccess))
}

func (h *HTTPServer) UpdateA2AAgents(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	agents, err := readA2AAgents(req)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	for _, agent := range agents {
		if err := h.storage.UpdateA2AAgent(agent); err != nil {
			handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			return
		}
	}
	handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_ExecuteSuccess))
}

func (h *HTTPServer) DeleteA2AAgents(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var deleteReq aitypes.A2AAgentDeleteRequest
	if err := req.ReadEntity(&deleteReq); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	for _, id := range deleteReq.AgentIds {
		if err := h.storage.DeleteA2AAgent(id); err != nil {
			handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			return
		}
	}
	handler.WriteHeaderAndProto(api.NewResponse(apimodel.Code_ExecuteSuccess))
}

func (h *HTTPServer) ListA2AAgentSkills(req *restful.Request, rsp *restful.Response) {
	agentID := req.QueryParameter("agent_id")
	if agentID == "" {
		name := req.QueryParameter("agent_name")
		namespace := req.QueryParameter("agent_namespace")
		if name != "" && namespace != "" {
			agent := h.cacheMgr.A2AAgent().GetA2AAgentByName(name, namespace)
			if agent != nil {
				agentID = agent.Id
			}
		}
	}
	skills := h.cacheMgr.A2AAgent().GetA2AAgentSkills(agentID)
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]interface{}{
		"code": uint32(apimodel.Code_ExecuteSuccess),
		"data": skills,
		"size": len(skills),
	}, restful.MIME_JSON)
}

func (h *HTTPServer) GetA2AAgentCard(req *restful.Request, rsp *restful.Response) {
	id := req.PathParameter("id")
	agent := h.cacheMgr.A2AAgent().GetA2AAgentByID(id)
	if agent == nil {
		_ = rsp.WriteHeaderAndJson(http.StatusNotFound, map[string]string{"error": "a2a agent not found"}, restful.MIME_JSON)
		return
	}
	if agent.RawCardJson == "" {
		_ = rsp.WriteHeaderAndJson(http.StatusOK, agent, restful.MIME_JSON)
		return
	}
	var card map[string]interface{}
	if err := json.Unmarshal([]byte(agent.RawCardJson), &card); err != nil {
		_ = rsp.WriteHeaderAndJson(http.StatusInternalServerError, map[string]string{"error": err.Error()}, restful.MIME_JSON)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, card, restful.MIME_JSON)
}

func parseA2AAgentQuery(values url.Values) *aitypes.A2AAgentQuery {
	return &aitypes.A2AAgentQuery{
		Name:                    values.Get("name"),
		Namespace:               values.Get("namespace"),
		Business:                values.Get("business"),
		Department:              values.Get("department"),
		ProtocolBinding:         values.Get("protocol_binding"),
		SkillTag:                values.Get("skill_tag"),
		BackendType:             values.Get("backend_type"),
		BackendServiceNamespace: values.Get("backend_service_namespace"),
		BackendServiceName:      values.Get("backend_service_name"),
		Streaming:               boolQuery(values, "streaming"),
		PushNotifications:       boolQuery(values, "push_notifications"),
		Offset:                  uint32Query(values, "offset", 0),
		Limit:                   uint32Query(values, "limit", 100),
	}
}

func newA2AAgentListResponse(amount uint32, agents []*aitypes.A2AAgent) *aitypes.A2AAgentListResponse {
	return &aitypes.A2AAgentListResponse{
		Code:   uint32(apimodel.Code_ExecuteSuccess),
		Amount: amount,
		Size:   uint32(len(agents)),
		Data:   agents,
	}
}

func readA2AAgents(req *restful.Request) ([]*aitypes.A2AAgent, error) {
	var agents []*aitypes.A2AAgent
	if err := json.NewDecoder(req.Request.Body).Decode(&agents); err != nil {
		return nil, err
	}
	return agents, nil
}

func uint32Query(values url.Values, key string, defaultValue uint32) uint32 {
	raw := values.Get(key)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint32(value)
}

func boolQuery(values url.Values, key string) *bool {
	raw := values.Get(key)
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &value
}
