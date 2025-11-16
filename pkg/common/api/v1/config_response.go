/*
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
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

// ConfigCollect BatchWriteResponse添加Response
func ConfigCollect(batchWriteResponse *apimodel.BatchWriteResponse, response *apimodel.Response) {
	// 非200的code，都归为异常
	if CalcCode(response) != 200 {
		if response.GetCode() >= batchWriteResponse.GetCode() {
			batchWriteResponse.Code = response.GetCode()
			batchWriteResponse.Info = Code2Info(batchWriteResponse.GetCode())
		}
	}
	batchWriteResponse.Responses = append(batchWriteResponse.Responses, response)
}

// NewConfigClientListResponse 创建配置客户端列表响应
func NewConfigClientListResponse(code apimodel.Code) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: 0,
		Size:   0,
	}
}

// NewConfigClientListResponseWithInfo 创建带详细信息的配置客户端列表响应
func NewConfigClientListResponseWithInfo(code apimodel.Code, msg string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: msg,
	}
}

// NewConfigClientResponse0 创建配置客户端响应
func NewConfigClientResponse0(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigClientResponse 创建带配置文件的配置客户端响应
func NewConfigClientResponse(code apimodel.Code, configFile interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: configFile would need to be serialized to Any if needed
	}
}

// NewConfigClientResponseFromConfigResponse 从配置响应创建配置客户端响应
func NewConfigClientResponseFromConfigResponse(response *apimodel.Response) *apimodel.Response {
	return &apimodel.Response{
		Code: response.Code,
		Info: response.Info,
		Data: response.Data,
	}
}

// NewConfigClientResponseWithInfo 创建带详细信息的配置客户端响应
func NewConfigClientResponseWithInfo(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: message,
	}
}

// NewConfigResponse 创建配置响应
func NewConfigResponse(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigGroupResponse 创建配置组响应
func NewConfigGroupResponse(code apimodel.Code, g interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: g would need to be serialized to Any if needed

	}
}

// NewConfigFileGroupResponseWithMessage 创建带消息的配置文件组响应
func NewConfigFileGroupResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

// NewConfigFileGroupBatchQueryResponse 创建配置文件组批量查询响应
func NewConfigFileGroupBatchQueryResponse(code apimodel.Code, total uint32,
	configFileGroups []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFileGroups)),
	}
}

// NewConfigBatchQueryResponse 创建配置批量查询响应
func NewConfigBatchQueryResponse(code apimodel.Code) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: 0,
		Size:   0,
	}
}

// NewConfigBatchQueryResponseWithInfo 创建带详细信息的配置批量查询响应
func NewConfigBatchQueryResponseWithInfo(code apimodel.Code, info string) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   info,
		Amount: 0,
		Size:   0,
	}
}

// NewConfigBatchWriteResponse 创建配置批量写入响应
func NewConfigBatchWriteResponse(code apimodel.Code) *apimodel.BatchWriteResponse {
	return &apimodel.BatchWriteResponse{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Size: 0,
	}
}

// NewConfigBatchWriteResponseWithInfo 创建带详细信息的配置批量写入响应
func NewConfigBatchWriteResponseWithInfo(code apimodel.Code, info string) *apimodel.BatchWriteResponse {
	return &apimodel.BatchWriteResponse{
		Code: uint32(code),
		Info: info,
		Size: 0,
	}
}

// NewConfigFileReleaseHistoryQueryResponse 创建配置文件发布历史查询响应
func NewConfigFileReleaseHistoryQueryResponse(code apimodel.Code, total uint32,
	configFileReleaseHistories []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFileReleaseHistories)),
	}
}

// NewConfigFileResponse 创建配置文件响应
func NewConfigFileResponse(code apimodel.Code, configFile interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigFileBatchQueryResponse 创建配置文件批量查询响应
func NewConfigFileBatchQueryResponse(
	code apimodel.Code, total uint32, configFiles []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFiles)),
	}
}

// NewConfigFileBatchQueryResponseWithMessage 创建带消息的配置文件批量查询响应
func NewConfigFileBatchQueryResponseWithMessage(
	code apimodel.Code, message string) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigFileTemplateResponse 创建配置文件模板响应
func NewConfigFileTemplateResponse(
	code apimodel.Code, template interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigFileTemplateResponseWithMessage 创建带消息的配置文件模板响应
func NewConfigFileTemplateResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

// NewConfigFileTemplateBatchQueryResponse 创建配置文件模板批量查询响应
func NewConfigFileTemplateBatchQueryResponse(code apimodel.Code, total uint32,
	configFileTemplates []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFileTemplates)),
	}
}

// NewConfigFileReleaseResponse 创建配置文件发布响应
func NewConfigFileReleaseResponse(
	code apimodel.Code, configFileRelease interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigResponseWithInfo 创建带详细信息的配置响应
func NewConfigResponseWithInfo(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

// NewSimpleConfigFileImportResponse 创建简单配置文件导入响应
func NewSimpleConfigFileImportResponse(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigFileImportResponse 创建配置文件导入响应
func NewConfigFileImportResponse(code apimodel.Code,
	createConfigFiles, skipConfigFiles, overwriteConfigFiles []interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

// NewConfigFileImportResponseWithMessage 创建带消息的配置文件导入响应
func NewConfigFileImportResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

// NewConfigFileExportResponse 创建配置文件导出响应
func NewConfigFileExportResponse(code apimodel.Code, data []byte) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: data would need special handling in the new model
	}
}

// NewConfigFileExportResponseWithMessage 创建带消息的配置文件导出响应
func NewConfigFileExportResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

// NewConfigEncryptAlgorithmResponse 创建配置加密算法响应
func NewConfigEncryptAlgorithmResponse(code apimodel.Code,
	algorithms []*string) *apimodel.Response {
	resp := &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: algorithms would need special handling in the new model
	}
	return resp
}
