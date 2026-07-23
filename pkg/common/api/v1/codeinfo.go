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

package v1

import apimodel "github.com/pole-io/specification/source/go/api/v1/model"

// pole-server错误码
// 六位构成，前面三位参照HTTP Status的标准
// 后面三位，依据内部的具体错误自定义
const (
	ExecuteSuccess           = uint32(apimodel.Code_ExecuteSuccess)
	DataNoChange             = uint32(apimodel.Code_DataNoChange)
	NoNeedUpdate             = uint32(apimodel.Code_NoNeedUpdate)
	BadRequest               = uint32(apimodel.Code_BadRequest)
	ParseException           = uint32(apimodel.Code_ParseException)
	EmptyRequest             = uint32(apimodel.Code_EmptyRequest)
	BatchSizeOverLimit       = uint32(apimodel.Code_BatchSizeOverLimit)
	InvalidDiscoverResource  = uint32(apimodel.Code_InvalidDiscoverResource)
	InvalidRequestID         = uint32(apimodel.Code_InvalidRequestID)
	InvalidUserName          = uint32(apimodel.Code_InvalidUserName)
	InvalidUserToken         = uint32(apimodel.Code_InvalidUserToken)
	InvalidParameter         = uint32(apimodel.Code_InvalidParameter)
	EmptyQueryParameter      = uint32(apimodel.Code_EmptyQueryParameter)
	InvalidQueryInsParameter = uint32(apimodel.Code_InvalidQueryInsParameter)
	HealthCheckNotOpen       = uint32(apimodel.Code_HealthCheckNotOpen)
	HeartbeatOnDisabledIns   = uint32(apimodel.Code_HeartbeatOnDisabledIns)
	HeartbeatExceedLimit     = uint32(apimodel.Code_HeartbeatExceedLimit)
	HeartbeatTypeNotFound    = uint32(apimodel.Code_HeartbeatTypeNotFound)
	InvalidMetadata          = uint32(apimodel.Code_InvalidMetadata)

	// 网格相关错误码
	ServicesExistedMesh  = uint32(apimodel.Code_ServicesExistedMesh)
	ResourcesExistedMesh = uint32(apimodel.Code_ResourcesExistedMesh)
	InvalidMeshParameter = uint32(apimodel.Code_InvalidMeshParameter)

	// 平台信息相关错误码

	// flux相关错误码

	ExistedResource                 = uint32(apimodel.Code_ExistedResource)
	NotFoundResource                = uint32(apimodel.Code_NotFoundResource)
	NamespaceExistedServices        = uint32(apimodel.Code_NamespaceExistedServices)
	ServiceExistedInstances         = uint32(apimodel.Code_ServiceExistedInstances)
	ServiceExistedRoutings          = uint32(apimodel.Code_ServiceExistedRoutings)
	ServiceExistedRateLimits        = uint32(apimodel.Code_ServiceExistedRateLimits)
	ExistReleasedConfig             = uint32(apimodel.Code_ExistReleasedConfig)
	SameInstanceRequest             = uint32(apimodel.Code_SameInstanceRequest)
	ServiceExistedCircuitBreakers   = uint32(apimodel.Code_ServiceExistedCircuitBreakers)
	ServiceExistedAlias             = uint32(apimodel.Code_ServiceExistedAlias)
	NamespaceExistedMeshResources   = uint32(apimodel.Code_NamespaceExistedMeshResources)
	NamespaceExistedCircuitBreakers = uint32(apimodel.Code_NamespaceExistedCircuitBreakers)
	ServiceSubscribedByMeshes       = uint32(apimodel.Code_ServiceSubscribedByMeshes)
	ServiceExistedFluxRateLimits    = uint32(apimodel.Code_ServiceExistedFluxRateLimits)
	NamespaceExistedConfigGroups    = uint32(apimodel.Code_NamespaceExistedConfigGroups)
	NamespaceExistedGovernanceRules = uint32(apimodel.Code_NamespaceExistedGovernanceRules)

	ClientAPINotOpen        = uint32(apimodel.Code_ClientAPINotOpen)
	Unauthorized            = uint32(apimodel.Code_Unauthorized)
	NotAllowedAccess        = uint32(apimodel.Code_NotAllowedAccess)
	IPRateLimit             = uint32(apimodel.Code_IPRateLimit)
	APIRateLimit            = uint32(apimodel.Code_APIRateLimit)
	CMDBNotFindHost         = uint32(apimodel.Code_CMDBNotFindHost)
	DataConflict            = uint32(apimodel.Code_DataConflict)
	InstanceTooManyRequests = uint32(apimodel.Code_InstanceTooManyRequests)
	ExecuteException        = uint32(apimodel.Code_ExecuteException)
	StoreLayerException     = uint32(apimodel.Code_StoreLayerException)
	CMDBPluginException     = uint32(apimodel.Code_CMDBPluginException)
	HeartbeatException      = uint32(apimodel.Code_HeartbeatException)
	InstanceRegisTimeout    = uint32(apimodel.Code_InstanceRegisTimeout)

	// 配置中心模块的错误码

	InvalidMatchRule                 = uint32(apimodel.Code_InvalidMatchRule)
	InvalidWorkloadCredentialRequest = uint32(apimodel.Code_InvalidWorkloadCredentialRequest)

	// 鉴权相关错误码
	InvalidUserOwners         = uint32(apimodel.Code_InvalidUserOwners)
	InvalidUserID             = uint32(apimodel.Code_InvalidUserID)
	InvalidUserPassword       = uint32(apimodel.Code_InvalidUserPassword)
	InvalidUserGroupOwners    = uint32(apimodel.Code_InvalidUserGroupOwners)
	InvalidUserGroupID        = uint32(apimodel.Code_InvalidUserGroupID)
	InvalidAuthStrategyOwners = uint32(apimodel.Code_InvalidAuthStrategyOwners)
	InvalidAuthStrategyName   = uint32(apimodel.Code_InvalidAuthStrategyName)
	InvalidAuthStrategyID     = uint32(apimodel.Code_InvalidAuthStrategyID)
	InvalidPrincipalType      = uint32(apimodel.Code_InvalidPrincipalType)

	UserExisted                            = uint32(apimodel.Code_UserExisted)
	UserGroupExisted                       = uint32(apimodel.Code_UserGroupExisted)
	AuthStrategyRuleExisted                = uint32(apimodel.Code_AuthStrategyRuleExisted)
	SubAccountExisted                      = uint32(apimodel.Code_SubAccountExisted)
	NotFoundUser                           = uint32(apimodel.Code_NotFoundUser)
	NotFoundOwnerUser                      = uint32(apimodel.Code_NotFoundOwnerUser)
	NotFoundUserGroup                      = uint32(apimodel.Code_NotFoundUserGroup)
	NotFoundAuthStrategyRule               = uint32(apimodel.Code_NotFoundAuthStrategyRule)
	NotAllowModifyDefaultStrategyPrincipal = uint32(apimodel.Code_NotAllowModifyDefaultStrategyPrincipal)
	NotAllowModifyOwnerDefaultStrategy     = uint32(apimodel.Code_NotAllowModifyOwnerDefaultStrategy)

	EmptyAutToken             = uint32(apimodel.Code_EmptyAutToken)
	TokenDisabled             = uint32(apimodel.Code_TokenDisabled)
	TokenNotExisted           = uint32(apimodel.Code_TokenNotExisted)
	InvalidWorkloadCredential = uint32(apimodel.Code_InvalidWorkloadCredential)
	ExpiredWorkloadCredential = uint32(apimodel.Code_ExpiredWorkloadCredential)

	AuthTokenVerifyException            = uint32(apimodel.Code_AuthTokenForbidden)
	OperationRoleException              = uint32(apimodel.Code_OperationRoleForbidden)
	WorkloadCredentialIssueForbidden    = uint32(apimodel.Code_WorkloadCredentialIssueForbidden)
	StaleServiceIdentityRevision        = uint32(apimodel.Code_StaleServiceIdentityRevision)
	WorkloadCredentialRateLimited       = uint32(apimodel.Code_WorkloadCredentialRateLimited)
	WorkloadCredentialIssuerUnavailable = uint32(apimodel.Code_WorkloadCredentialIssuerUnavailable)
)

// code to string
// code的字符串描述信息
var code2info = map[uint32]string{
	ExecuteSuccess:                  "execute success",
	DataNoChange:                    "discover data is no change",
	NoNeedUpdate:                    "update data is no change, no need to update",
	BadRequest:                      "bad request",
	ParseException:                  "request decode failed",
	EmptyRequest:                    "empty request",
	BatchSizeOverLimit:              "batch size over the limit",
	InvalidDiscoverResource:         "invalid discover resource",
	InvalidRequestID:                "invalid request id",
	InvalidUserName:                 "invalid user name",
	InvalidUserToken:                "invalid user token",
	InvalidParameter:                "invalid parameter",
	EmptyQueryParameter:             "query instance parameter is empty",
	InvalidQueryInsParameter:        "query instance, service or namespace or host is required",
	HealthCheckNotOpen:              "server not open health check",
	HeartbeatOnDisabledIns:          "heartbeat on disabled instance",
	HeartbeatExceedLimit:            "instance can only heartbeat 1 time per second",
	InvalidMetadata:                 "the length of metadata is too long or metadata contains invalid characters",
	ExistedResource:                 "existed resource",
	SameInstanceRequest:             "the same instance request",
	NotFoundResource:                "not found resource",
	ClientAPINotOpen:                "client api is not open",
	NamespaceExistedServices:        "some services existed in namespace",
	ServiceExistedInstances:         "some instances existed in service",
	ServiceExistedRoutings:          "some routings existed in service",
	ServiceExistedRateLimits:        "some rate limits existed in service",
	ServiceExistedCircuitBreakers:   "some circuit breakers existed in service",
	ServiceExistedAlias:             "some aliases existed in service",
	NamespaceExistedMeshResources:   "some mesh resources existed in namespace",
	NamespaceExistedCircuitBreakers: "some circuit breakers existed in namespace",
	NamespaceExistedGovernanceRules: "some governance rules existed in namespace",
	ExistReleasedConfig:             "exist released config",
	Unauthorized:                    "unauthorized",
	NotAllowedAccess:                "access is not approved",
	IPRateLimit:                     "server limit the ip access",
	APIRateLimit:                    "server limit the api access",
	CMDBNotFindHost:                 "not found the host cmdb",
	DataConflict:                    "data is conflict, please try again",
	InstanceTooManyRequests:         "your instance has too many requests",
	ExecuteException:                "execute exception",
	StoreLayerException:             "store layer exception",
	CMDBPluginException:             "cmdb plugin exception",

	HeartbeatException: "heartbeat execute exception",

	ServicesExistedMesh:       "services existed mesh",
	ResourcesExistedMesh:      "resources existed mesh",
	ServiceSubscribedByMeshes: "service subscribed by some mesh",
	InvalidMeshParameter:      "invalid mesh parameter",

	InstanceRegisTimeout: "instance async regist timeout",

	// 配置中心的错误信息
	InvalidMatchRule:                 "invalid gray config beta labels",
	InvalidWorkloadCredentialRequest: "invalid workload credential request",

	// 鉴权错误
	NotFoundUser:             "not found user",
	NotFoundOwnerUser:        "not found owner user",
	NotFoundUserGroup:        "not found usergroup",
	NotFoundAuthStrategyRule: "not found auth strategy rule",

	UserExisted:                         "exist user",
	UserGroupExisted:                    "exist usergroup",
	AuthStrategyRuleExisted:             "exist auth strategy rule",
	InvalidUserGroupOwners:              "invalid usergroup owner attribute",
	InvalidAuthStrategyName:             "invalid auth strategy rule name",
	InvalidAuthStrategyOwners:           "invalid auth strategy rule owner",
	InvalidUserPassword:                 "invalid user password",
	InvalidPrincipalType:                "invalid principal type",
	TokenDisabled:                       "token already disabled",
	AuthTokenVerifyException:            "token verify exception",
	OperationRoleException:              "operation role exception",
	EmptyAutToken:                       "auth token empty",
	SubAccountExisted:                   "some sub-account existed in owner",
	InvalidUserID:                       "invalid user-id",
	TokenNotExisted:                     "token not existed",
	InvalidWorkloadCredential:           "invalid workload credential",
	ExpiredWorkloadCredential:           "workload credential expired",
	WorkloadCredentialIssueForbidden:    "workload credential issue forbidden",
	StaleServiceIdentityRevision:        "stale service identity revision",
	WorkloadCredentialRateLimited:       "workload credential rate limited",
	WorkloadCredentialIssuerUnavailable: "workload credential issuer unavailable",

	NotAllowModifyDefaultStrategyPrincipal: "not allow modify default strategy principal",
	NotAllowModifyOwnerDefaultStrategy:     "not allow modify main account default strategy",

	NamespaceExistedConfigGroups: "some config group existed in namespace",
}

// code to info
func Code2Info(code uint32) string {
	info, ok := code2info[code]
	if ok {
		return info
	}

	return ""
}
