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

func NewConfigClientListResponse(code apimodel.Code) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: 0,
		Size:   0,
	}
}

func NewConfigClientListResponseWithInfo(code apimodel.Code, msg string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: msg,
	}
}

func NewConfigClientResponse0(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigClientResponse(code apimodel.Code, configFile interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: configFile would need to be serialized to Any if needed
	}
}

func NewConfigClientResponseFromConfigResponse(response *apimodel.Response) *apimodel.Response {
	return &apimodel.Response{
		Code: response.Code,
		Info: response.Info,
		Data: response.Data,
	}
}

func NewConfigClientResponseWithInfo(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: message,
	}
}

func NewConfigResponse(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigGroupResponse(code apimodel.Code, g interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: g would need to be serialized to Any if needed
	}
}

func NewConfigFileGroupResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

func NewConfigFileGroupBatchQueryResponse(code apimodel.Code, total uint32,
	configFileGroups []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFileGroups)),
	}
}

func NewConfigBatchQueryResponse(code apimodel.Code) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: 0,
		Size:   0,
	}
}

func NewConfigBatchQueryResponseWithInfo(code apimodel.Code, info string) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   info,
		Amount: 0,
		Size:   0,
	}
}

func NewConfigBatchWriteResponse(code apimodel.Code) *apimodel.BatchWriteResponse {
	return &apimodel.BatchWriteResponse{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		Size: 0,
	}
}

func NewConfigBatchWriteResponseWithInfo(code apimodel.Code, info string) *apimodel.BatchWriteResponse {
	return &apimodel.BatchWriteResponse{
		Code: uint32(code),
		Info: info,
		Size: 0,
	}
}

func NewConfigFileReleaseHistoryQueryResponse(code apimodel.Code, total uint32,
	configFileReleaseHistories []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFileReleaseHistories)),
	}
}

func NewConfigFileResponse(code apimodel.Code, configFile interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigFileBatchQueryResponse(
	code apimodel.Code, total uint32, configFiles []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFiles)),
	}
}

func NewConfigFileBatchQueryResponseWithMessage(
	code apimodel.Code, message string) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigFileTemplateResponse(
	code apimodel.Code, template interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigFileTemplateResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

func NewConfigFileTemplateBatchQueryResponse(code apimodel.Code, total uint32,
	configFileTemplates []interface{}) *apimodel.BatchQueryResponse {
	return &apimodel.BatchQueryResponse{
		Code:   uint32(code),
		Info:   Code2Info(uint32(code)),
		Amount: total,
		Size:   uint32(len(configFileTemplates)),
	}
}

func NewConfigFileReleaseResponse(
	code apimodel.Code, configFileRelease interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigResponseWithInfo(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

func NewSimpleConfigFileImportResponse(code apimodel.Code) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigFileImportResponse(code apimodel.Code,
	createConfigFiles, skipConfigFiles, overwriteConfigFiles []interface{}) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
	}
}

func NewConfigFileImportResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

func NewConfigFileExportResponse(code apimodel.Code, data []byte) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: data would need special handling in the new model
	}
}

func NewConfigFileExportResponseWithMessage(code apimodel.Code, message string) *apimodel.Response {
	return &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)) + ":" + message,
	}
}

func NewConfigEncryptAlgorithmResponse(code apimodel.Code,
	algorithms []*string) *apimodel.Response {
	resp := &apimodel.Response{
		Code: uint32(code),
		Info: Code2Info(uint32(code)),
		// Note: algorithms would need special handling in the new model
	}
	return resp
}
