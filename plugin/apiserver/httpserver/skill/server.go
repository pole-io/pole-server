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

	skillSvr "github.com/pole-io/pole-server/pkg/skill"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/skill/docs"
)

// HTTPServer Skill API HTTP server
type HTTPServer struct {
	skillServer skillSvr.SkillServer
}

// NewHTTPServer creates a new Skill HTTP server
func NewHTTPServer(skillServer skillSvr.SkillServer) *HTTPServer {
	return &HTTPServer{
		skillServer: skillServer,
	}
}

// GetSkillAccessServer returns the Skill API WebService
func (h *HTTPServer) GetSkillAccessServer(include []string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/skill/v1").Consumes(restful.MIME_JSON, "multipart/form-data").Produces(restful.MIME_JSON)

	h.addSkillAccess(ws)
	h.addSkillGroupAccess(ws)

	return ws
}

// addSkillAccess registers Skill API routes
func (h *HTTPServer) addSkillAccess(ws *restful.WebService) {
	// Skill CRUD
	ws.Route(docs.EnrichCreateSkillsApiDocs(ws.POST("/skills").To(h.CreateSkills)))
	ws.Route(docs.EnrichUpdateSkillsApiDocs(ws.PUT("/skills").To(h.UpdateSkills)))
	ws.Route(docs.EnrichDeleteSkillsApiDocs(ws.POST("/skills/delete").To(h.DeleteSkills)))
	ws.Route(docs.EnrichGetSkillsApiDocs(ws.GET("/skills").To(h.GetSkills)))
	ws.Route(docs.EnrichGetAllSkillsApiDocs(ws.GET("/skills/all").To(h.GetAllSkills)))
	ws.Route(docs.EnrichGetSkillsCountApiDocs(ws.GET("/skills/count").To(h.GetSkillsCount)))
}

// addSkillGroupAccess registers SkillGroup API routes
func (h *HTTPServer) addSkillGroupAccess(ws *restful.WebService) {
	// SkillGroup CRUD
	ws.Route(docs.EnrichCreateSkillGroupsApiDocs(ws.POST("/skill/groups").To(h.CreateSkillGroups)))
	ws.Route(docs.EnrichUpdateSkillGroupsApiDocs(ws.PUT("/skill/groups").To(h.UpdateSkillGroups)))
	ws.Route(docs.EnrichDeleteSkillGroupsApiDocs(ws.POST("/skill/groups/delete").To(h.DeleteSkillGroups)))
	ws.Route(docs.EnrichGetSkillGroupsApiDocs(ws.GET("/skill/groups").To(h.GetSkillGroups)))
}
