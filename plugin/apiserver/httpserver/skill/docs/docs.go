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

package docs

import (
	"github.com/emicklei/go-restful/v3"
	restfulspec "github.com/polarismesh/go-restful-openapi/v2"
)

var (
	skillsApiTags            = []string{"Skills"}
	skillGroupsApiTags       = []string{"SkillGroups"}
	skillVersionsApiTags     = []string{"SkillVersions"}
	skillSubscriptionsApiTags = []string{"SkillSubscriptions"}
)

// EnrichCreateSkillsApiDocs enriches the CreateSkills API documentation
func EnrichCreateSkillsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("创建技能").
		Metadata(restfulspec.KeyOpenAPITags, skillsApiTags).
		Reads([]map[string]interface{}{}, "创建技能列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichUpdateSkillsApiDocs enriches the UpdateSkills API documentation
func EnrichUpdateSkillsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("更新技能").
		Metadata(restfulspec.KeyOpenAPITags, skillsApiTags).
		Reads([]map[string]interface{}{}, "更新技能列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichDeleteSkillsApiDocs enriches the DeleteSkills API documentation
func EnrichDeleteSkillsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("删除技能").
		Metadata(restfulspec.KeyOpenAPITags, skillsApiTags).
		Reads([]map[string]interface{}{}, "删除技能列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetSkillsApiDocs enriches the GetSkills API documentation
func EnrichGetSkillsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("查询技能列表").
		Metadata(restfulspec.KeyOpenAPITags, skillsApiTags).
		Param(restful.QueryParameter("namespace", "命名空间").
			DataType("string").Required(true)).
		Param(restful.QueryParameter("name", "技能名称（模糊匹配）").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("skill_type", "技能类型").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("owner", "所有者").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("offset", "偏移量").
			DataType("integer").Required(false).DefaultValue("0")).
		Param(restful.QueryParameter("limit", "限制数量").
			DataType("integer").Required(false).DefaultValue("100")).
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetAllSkillsApiDocs enriches the GetAllSkills API documentation
func EnrichGetAllSkillsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("获取所有技能").
		Metadata(restfulspec.KeyOpenAPITags, skillsApiTags).
		Param(restful.QueryParameter("namespace", "命名空间").
			DataType("string").Required(true)).
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetSkillsCountApiDocs enriches the GetSkillsCount API documentation
func EnrichGetSkillsCountApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("获取技能数量").
		Metadata(restfulspec.KeyOpenAPITags, skillsApiTags).
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichCreateSkillGroupsApiDocs enriches the CreateSkillGroups API documentation
func EnrichCreateSkillGroupsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("创建技能组").
		Metadata(restfulspec.KeyOpenAPITags, skillGroupsApiTags).
		Reads([]map[string]interface{}{}, "创建技能组列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichUpdateSkillGroupsApiDocs enriches the UpdateSkillGroups API documentation
func EnrichUpdateSkillGroupsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("更新技能组").
		Metadata(restfulspec.KeyOpenAPITags, skillGroupsApiTags).
		Reads([]map[string]interface{}{}, "更新技能组列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichDeleteSkillGroupsApiDocs enriches the DeleteSkillGroups API documentation
func EnrichDeleteSkillGroupsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("删除技能组").
		Metadata(restfulspec.KeyOpenAPITags, skillGroupsApiTags).
		Reads([]map[string]interface{}{}, "删除技能组列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetSkillGroupsApiDocs enriches the GetSkillGroups API documentation
func EnrichGetSkillGroupsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("查询技能组列表").
		Metadata(restfulspec.KeyOpenAPITags, skillGroupsApiTags).
		Param(restful.QueryParameter("namespace", "命名空间").
			DataType("string").Required(true)).
		Param(restful.QueryParameter("name", "技能组名称（模糊匹配）").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("owner", "所有者").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("offset", "偏移量").
			DataType("integer").Required(false).DefaultValue("1")).
		Param(restful.QueryParameter("limit", "限制数量").
			DataType("integer").Required(false).DefaultValue("100")).
		Returns(200, "成功", map[string]interface{}{})
}

// ===== SkillVersion API Docs =====

// EnrichCreateSkillVersionsApiDocs enriches the CreateSkillVersions API documentation
func EnrichCreateSkillVersionsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("创建技能版本").
		Metadata(restfulspec.KeyOpenAPITags, skillVersionsApiTags).
		Reads([]map[string]interface{}{}, "创建技能版本列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichDeleteSkillVersionsApiDocs enriches the DeleteSkillVersions API documentation
func EnrichDeleteSkillVersionsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("删除技能版本").
		Metadata(restfulspec.KeyOpenAPITags, skillVersionsApiTags).
		Reads([]map[string]interface{}{}, "删除技能版本列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetSkillVersionsApiDocs enriches the GetSkillVersions API documentation
func EnrichGetSkillVersionsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("查询技能版本列表").
		Metadata(restfulspec.KeyOpenAPITags, skillVersionsApiTags).
		Param(restful.QueryParameter("skill_id", "技能 ID").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("skill_name", "技能名称").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("namespace", "命名空间").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("offset", "偏移量").
			DataType("integer").Required(false).DefaultValue("0")).
		Param(restful.QueryParameter("limit", "限制数量").
			DataType("integer").Required(false).DefaultValue("100")).
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichActivateSkillVersionApiDocs enriches the ActivateSkillVersion API documentation
func EnrichActivateSkillVersionApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("激活技能版本").
		Metadata(restfulspec.KeyOpenAPITags, skillVersionsApiTags).
		Reads(map[string]interface{}{}, "包含 version_id 的请求体").
		Returns(200, "成功", map[string]interface{}{})
}

// ===== SkillSubscription API Docs =====

// EnrichCreateSkillSubscriptionsApiDocs enriches the CreateSkillSubscriptions API documentation
func EnrichCreateSkillSubscriptionsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("创建技能订阅").
		Metadata(restfulspec.KeyOpenAPITags, skillSubscriptionsApiTags).
		Reads([]map[string]interface{}{}, "创建技能订阅列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichDeleteSkillSubscriptionsApiDocs enriches the DeleteSkillSubscriptions API documentation
func EnrichDeleteSkillSubscriptionsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("删除技能订阅").
		Metadata(restfulspec.KeyOpenAPITags, skillSubscriptionsApiTags).
		Reads([]map[string]interface{}{}, "删除技能订阅列表").
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetSkillSubscriptionsApiDocs enriches the GetSkillSubscriptions API documentation
func EnrichGetSkillSubscriptionsApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("查询技能订阅列表").
		Metadata(restfulspec.KeyOpenAPITags, skillSubscriptionsApiTags).
		Param(restful.QueryParameter("skill_name", "技能名称").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("client_id", "客户端 ID").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("namespace", "命名空间").
			DataType("string").Required(false)).
		Param(restful.QueryParameter("offset", "偏移量").
			DataType("integer").Required(false).DefaultValue("0")).
		Param(restful.QueryParameter("limit", "限制数量").
			DataType("integer").Required(false).DefaultValue("100")).
		Returns(200, "成功", map[string]interface{}{})
}

// EnrichGetSkillSubscriptionsByClientApiDocs enriches the GetSkillSubscriptionsByClient API documentation
func EnrichGetSkillSubscriptionsByClientApiDocs(r *restful.RouteBuilder) *restful.RouteBuilder {
	return r.
		Doc("查询客户端的技能订阅").
		Metadata(restfulspec.KeyOpenAPITags, skillSubscriptionsApiTags).
		Param(restful.PathParameter("client_id", "客户端 ID").
			DataType("string").Required(true)).
		Returns(200, "成功", map[string]interface{}{})
}
