/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package skill

import (
	"github.com/emicklei/go-restful/v3"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// CreateSkills creates one or more skills
func (h *HTTPServer) CreateSkills(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var skills SkillArr
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &Skill{}
		skills = append(skills, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	// Call skill server
	ret := h.skillServer.CreateSkills(ctx, skills.ToAIType())
	handler.WriteHeaderAndProto(ret)
}

// UpdateSkills updates one or more skills
func (h *HTTPServer) UpdateSkills(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var skills SkillArr
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &Skill{}
		skills = append(skills, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	// Call skill server
	ret := h.skillServer.UpdateSkills(ctx, skills.ToAIType())
	handler.WriteHeaderAndProto(ret)
}

// DeleteSkills deletes one or more skills
func (h *HTTPServer) DeleteSkills(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var skills SkillArr
	ctx, err := handler.ParseJSONArray(func() any {
		msg := &Skill{}
		skills = append(skills, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
			apimodel.Code_ParseException, err.Error()))
		return
	}

	// Call skill server
	ret := h.skillServer.DeleteSkills(ctx, skills.ToAIType())
	handler.WriteHeaderAndProto(ret)
}

// GetSkills queries skills with filters
func (h *HTTPServer) GetSkills(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	// Required parameter check
	if queryParams["namespace"] == "" {
		handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
			apimodel.Code_BadRequest, "namespace is required"))
		return
	}

	ret := h.skillServer.GetSkills(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetAllSkills gets all skills
func (h *HTTPServer) GetAllSkills(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	queryParams := httpcommon.ParseQueryParams(req)
	ctx := handler.ParseHeaderContext()

	// Required parameter check
	if queryParams["namespace"] == "" {
		handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
			apimodel.Code_BadRequest, "namespace is required"))
		return
	}

	ret := h.skillServer.GetAllSkills(ctx, queryParams)
	handler.WriteHeaderAndProto(ret)
}

// GetSkillsCount gets the total count of skills
func (h *HTTPServer) GetSkillsCount(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ctx := handler.ParseHeaderContext()

	ret := h.skillServer.GetSkillsCount(ctx)
	handler.WriteHeaderAndProto(ret)
}
