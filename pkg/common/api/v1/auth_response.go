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

import (
	anypb "google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
)

// NewAuthResponse 创建回复消息
func NewAuthResponse(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewAuthResponseWithMsg 创建回复消息
func NewAuthResponseWithMsg(code apimodel.Code, msg string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + msg,
	}
}

// NewAuthBatchWriteResponse 创建批量回复
func NewAuthBatchWriteResponse(code apimodel.Code) *apimodel.BatchWriteResponse {
	return &apimodel.BatchWriteResponse{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Size: 0,
	}
}

// NewAuthBatchQueryResponse 创建批量查询回复
func NewAuthBatchQueryResponse(code apimodel.Code) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: 0,
		Size:   0,
	}
}

// NewAuthBatchQueryResponseWithMsg 创建带详细信息的批量查询回复
func NewAuthBatchQueryResponseWithMsg(code apimodel.Code, msg string) *apimodel.BatchQueryResponse {
	resp := NewAuthBatchQueryResponse(code)
	resp.Info += ": " + msg
	return resp
}

// NewUserResponse 创建回复带用户信息
func NewUserResponse(code apimodel.Code, user *apisecurity.User) *apimodel.Response {
	var data *anypb.Any
	if user != nil {
		data, _ = anypb.New(user)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewUserResponseWithMsg 创建回复带用户信息和自定义消息
func NewUserResponseWithMsg(code apimodel.Code, info string, user *apisecurity.User) *apimodel.Response {
	var data *anypb.Any
	if user != nil {
		data, _ = anypb.New(user)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + info,
		Data: data,
	}
}

// NewGroupResponse 创建回复带用户组信息
func NewGroupResponse(code apimodel.Code, group *apisecurity.UserGroup) *apimodel.Response {
	var data *anypb.Any
	if group != nil {
		data, _ = anypb.New(group)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewModifyGroupResponse 创建修改用户组的响应信息
func NewModifyGroupResponse(code apimodel.Code, group *apisecurity.ModifyUserGroup) *apimodel.Response {
	var data *anypb.Any
	if group != nil {
		data, _ = anypb.New(group)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewGroupRelationResponse 创建用户组关联关系的响应体
func NewGroupRelationResponse(code apimodel.Code, relation *apisecurity.UserGroupRelation) *apimodel.Response {
	var data *anypb.Any
	if relation != nil {
		data, _ = anypb.New(relation)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewAuthStrategyResponse 创建鉴权策略响应体
func NewAuthStrategyResponse(code apimodel.Code, req *apisecurity.AuthStrategy) *apimodel.Response {
	var data *anypb.Any
	if req != nil {
		data, _ = anypb.New(req)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewAuthStrategyResponseWithMsg 创建鉴权策略响应体并自定义Info
func NewAuthStrategyResponseWithMsg(
	code apimodel.Code, msg string, req *apisecurity.AuthStrategy) *apimodel.Response {
	var data *anypb.Any
	if req != nil {
		data, _ = anypb.New(req)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: msg,
		Data: data,
	}
}

// NewModifyAuthStrategyResponse 创建修改鉴权策略响应体
func NewModifyAuthStrategyResponse(code apimodel.Code, req *apisecurity.ModifyAuthStrategy) *apimodel.Response {
	var data *anypb.Any
	if req != nil {
		data, _ = anypb.New(req)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewStrategyResourcesResponse 创建修改鉴权策略响应体
func NewStrategyResourcesResponse(code apimodel.Code, ret *apisecurity.StrategyResources) *apimodel.Response {
	var data *anypb.Any
	if ret != nil {
		data, _ = anypb.New(ret)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}

// NewLoginResponse 创建登录响应体
func NewLoginResponse(code apimodel.Code, loginResponse *apisecurity.LoginResponse) *apimodel.Response {
	var data *anypb.Any
	if loginResponse != nil {
		data, _ = anypb.New(loginResponse)
	}
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: data,
	}
}
