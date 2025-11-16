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
	userAny, err := anypb.New(user)
	if err != nil {
        // 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化用户信息失败: " + err.Error(),
        }
    }
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: userAny,
	}
}

// NewUserResponse 创建回复带用户信息
func NewUserResponseWithMsg(code apimodel.Code, info string, user *apisecurity.User) *apimodel.Response {
	userAny, err := anypb.New(user)
	if err != nil {
        // 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化用户信息失败: " + err.Error(),
        }
    }
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + info,
		Data: userAny,
	}
}

// NewGroupResponse 创建回复带用户组信息
func NewGroupResponse(code apimodel.Code, user *apisecurity.UserGroup) *apimodel.Response {
	userAny, err := anypb.New(user)
	if err != nil {
		// 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化用户组信息失败: " + err.Error(),
        }
    }
	return &apimodel.Response{
		Code:      uint32(code),
		Info:      Code2Info(uint32(code)),
		Data:      userAny,
	}
}

// NewModifyGroupResponse 创建修改用户组的响应信息
func NewModifyGroupResponse(code apimodel.Code, group *apisecurity.ModifyUserGroup) *apimodel.Response {
	groupAny, err := anypb.New(group)
	if err != nil {
        // 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化修改用户组信息失败: " + err.Error(),
        }
    }
	return &apimodel.Response{

		Code:            uint32(code),
		Info:            Code2Info(uint32(code)),
		Data:            groupAny,
	}
}

// NewGroupRelationResponse 创建用户组关联关系的响应体
func NewGroupRelationResponse(code apimodel.Code, relation *apisecurity.UserGroupRelation) *apimodel.Response {
	return &apimodel.Response{
		Code:     uint32(code),
		Info:     Code2Info(uint32(code)),
	}
}

// NewAuthStrategyResponse 创建鉴权策略响应体
func NewAuthStrategyResponse(code apimodel.Code, req *apisecurity.AuthStrategy) *apimodel.Response {
	return &apimodel.Response{
		Code:         uint32(code),
		Info:         Code2Info(uint32(code)),
	}
}

// NewAuthStrategyResponseWithMsg 创建鉴权策略响应体并自定义Info
func NewAuthStrategyResponseWithMsg(
	code apimodel.Code, msg string, req *apisecurity.AuthStrategy) *apimodel.Response {
	reqAny, err := anypb.New(req)
	if err != nil {
		// 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化鉴权策略失败: " + err.Error(),
        }
    }
	return &apimodel.Response{
		Code:        uint32(code),
		Info:        msg,
		Data:        reqAny,
	}
}

// NewModifyAuthStrategyResponse 创建修改鉴权策略响应体
func NewModifyAuthStrategyResponse(code apimodel.Code, req *apisecurity.ModifyAuthStrategy) *apimodel.Response {
	return &apimodel.Response{
		Code:               uint32(code),
		Info:               Code2Info(uint32(code)),
	}
}

// NewStrategyResourcesResponse 创建修改鉴权策略响应体
func NewStrategyResourcesResponse(code apimodel.Code, ret *apisecurity.StrategyResources) *apimodel.Response {
	retAny, err := anypb.New(ret)
	if err != nil {
        // 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化策略资源失败: " + err.Error(),
        }
    }
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: retAny,
	}
}

// NewLoginResponse 创建登录响应体
func NewLoginResponse(code apimodel.Code, loginResponse *apisecurity.LoginResponse) *apimodel.Response {
	loginResponseAny, err := anypb.New(loginResponse)
	if err != nil {
        // 处理序列化失败的错误（比如返回错误响应）
        return &apimodel.Response{
            Code: uint32(code),
            Info: "序列化登录响应失败: " + err.Error(),
        }
    }
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Data: loginResponseAny,
	}
}
