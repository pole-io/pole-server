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

package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// some options config
const (
	// service表
	MaxDbServiceNamespaceLength = 64
	MaxDbServicePortsLength     = 8192
	MaxDbServiceBusinessLength  = 128
	MaxDbServiceDeptLength      = 1024
	MaxDbServiceCMDBLength      = 1024
	MaxDbServiceCommentLength   = 1024
	MaxDbServiceOwnerLength     = 1024
	MaxDbServiceToken           = 2048

	// instance表
	MaxDbInsHostLength     = 128
	MaxDbInsProtocolLength = 32
	MaxDbInsVersionLength  = 32
	MaxDbInsLogicSetLength = 128

	// circuitbreaker表
	MaxDbCircuitbreakerName       = 32
	MaxDbCircuitbreakerNamespace  = 64
	MaxDbCircuitbreakerBusiness   = 64
	MaxDbCircuitbreakerDepartment = 1024
	MaxDbCircuitbreakerComment    = 1024
	MaxDbCircuitbreakerOwner      = 1024
	MaxDbCircuitbreakerVersion    = 32

	// ratelimit表
	MaxDbRateLimitName = MaxRuleName

	MaxRuleName = 64

	MaxPlatformIDLength     = 32
	MaxPlatformNameLength   = 128
	MaxPlatformDomainLength = 1024
	MaxPlatformQPS          = 65535
)

// CalculateInstanceID 计算实例ID
func CalculateInstanceID(namespace string, service string, vpcID string, host string, port uint32) (string, error) {
	h := sha1.New()
	var str string
	// 兼容带有vpcID的instance
	if vpcID == "" {
		str = fmt.Sprintf("%s##%s##%s##%d", namespace, service, host, port)
	} else {
		str = fmt.Sprintf("%s##%s##%s##%s##%d", namespace, service, vpcID, host, port)
	}

	if _, err := io.WriteString(h, str); err != nil {
		return "", err
	}

	out := hex.EncodeToString(h.Sum(nil))
	return out, nil
}

// CalculateRuleID 计算规则ID
func CalculateRuleID(name, namespace string) string {
	return name + "." + namespace
}

// ParseRequestID 从ctx中获取Request-ID
func ParseRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(types.ContextRequestId).(string)
	return rid
}

// ParseClientAddress 从ctx中获取客户端地址
func ParseClientAddress(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(types.ContextClientAddress).(string)
	return rid
}

// ParseClientIP .
func ParseClientIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(types.ContextClientAddress).(string)
	if strings.Contains(rid, ":") {
		return strings.Split(rid, ":")[0]
	}
	return rid
}

// ParseAuthToken 从ctx中获取token
func ParseAuthToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, _ := ctx.Value(types.ContextAuthTokenKey).(string)
	return token
}

// ParseIsOwner 从ctx中获取token
func ParseIsOwner(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	isOwner, _ := ctx.Value(types.ContextIsOwnerKey).(bool)
	return isOwner
}

// ParseUserID 从ctx中解析用户ID
func ParseUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userID, _ := ctx.Value(types.ContextUserIDKey).(string)
	return userID
}

// ParseUserName 从ctx解析用户名称
func ParseUserName(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userName, _ := ctx.Value(types.ContextUserNameKey).(string)
	if userName == "" {
		return ParseOperator(ctx)
	}
	return userName
}

// ParseOwnerID 从ctx解析Owner ID
func ParseOwnerID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	ownerID, _ := ctx.Value(types.ContextOwnerIDKey).(string)
	return ownerID
}

// ParseToken 从ctx中获取token
func ParseToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, _ := ctx.Value(types.ContextPolarisToken).(string)
	return token
}

// ParseOperator 从ctx中获取operator
func ParseOperator(ctx context.Context) string {
	defaultOperator := "Polaris"
	if ctx == nil {
		return defaultOperator
	}

	if operator, _ := ctx.Value(types.ContextOperator).(string); operator != "" {
		return operator
	}

	return defaultOperator
}

// ZapRequestID 生成Request-ID的日志描述
func ZapRequestID(id string) zap.Field {
	return zap.String("request-id", id)
}

// RequestID 从ctx中获取Request-ID
func RequestID(ctx context.Context) zap.Field {
	return zap.String("request-id", ParseRequestID(ctx))
}

// ZapPlatformID 生成Platform-ID的日志描述
func ZapPlatformID(id string) zap.Field {
	return zap.String("platform-id", id)
}

// ZapInstanceID 生成instanceID的日志描述
func ZapInstanceID(id string) zap.Field {
	return zap.String("instance-id", id)
}

// ZapNamespace 生成namespace的日志描述
func ZapNamespace(namespace string) zap.Field {
	return zap.String("namesapce", namespace)
}

// ZapGroup 生成group的日志描述
func ZapGroup(group string) zap.Field {
	return zap.String("group", group)
}

// ZapFileName 生成fileName的日志描述
func ZapFileName(fileName string) zap.Field {
	return zap.String("file-name", fileName)
}

// ZapReleaseName 生成fileName的日志描述
func ZapReleaseName(fileName string) zap.Field {
	return zap.String("release-name", fileName)
}

// ZapVersion 生成 version 的日志描述
func ZapVersion(version uint64) zap.Field {
	return zap.Uint64("version", version)
}

// ConvertStringValuesToSlice 转换StringValues为字符串切片
func ConvertStringValuesToSlice(vals []*wrapperspb.StringValue) []string {
	ret := make([]string, 0, 4)

	for index := range vals {
		id := vals[index]
		if strings.TrimSpace(id.GetValue()) == "" {
			continue
		}
		ret = append(ret, id.GetValue())
	}

	return ret
}

// BuildSha1Digest 构建SHA1摘要
func BuildSha1Digest(value string) (string, error) {
	if len(value) == 0 {
		return "", nil
	}
	h := sha1.New()
	if _, err := io.WriteString(h, value); err != nil {
		return "", err
	}
	out := hex.EncodeToString(h.Sum(nil))
	return out, nil
}

func CheckContractInterfaceTetrad(contractId string, source apiservice.InterfaceDescriptor_Source,
	req *apiservice.InterfaceDescriptor) (string, *apiservice.Response) {
	if contractId == "" {
		return "", api.NewResponseWithMsg(apimodel.Code_BadRequest, "invalid service_contract id")
	}
	if req.GetId() != "" {
		return req.GetId(), nil
	}
	if req.GetPath() == "" {
		return "", api.NewResponseWithMsg(apimodel.Code_BadRequest, "invalid service_contract interface path")
	}
	h := sha1.New()
	str := fmt.Sprintf("%s##%s##%s##%s##%d", contractId, req.GetMethod(), req.GetPath(), req.GetName(), source)

	if _, err := io.WriteString(h, str); err != nil {
		return "", api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}
	out := hex.EncodeToString(h.Sum(nil))
	return out, nil
}

func CalculateContractID(namespace, service, name, protocol, version string) (string, error) {
	h := sha1.New()
	str := fmt.Sprintf("%s##%s##%s##%s##%s", namespace, service, name, protocol, version)

	if _, err := io.WriteString(h, str); err != nil {
		return "", err
	}

	out := hex.EncodeToString(h.Sum(nil))
	return out, nil
}

// ConvertMetadataToStringValue 将Metadata转换为可序列化字符串
func ConvertMetadataToStringValue(metadata map[string]string) (string, error) {
	if metadata == nil {
		return "", nil
	}
	v, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// ConvertStringValueToMetadata 将字符串反序列为metadata
func ConvertStringValueToMetadata(str string) (map[string]string, error) {
	if str == "" {
		return nil, nil
	}
	v := make(map[string]string)
	err := json.Unmarshal([]byte(str), &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// NeedUpdateMetadata 判断是否出现了metadata的变更
func NeedUpdateMetadata(metadata map[string]string, inMetadata map[string]string) bool {
	if inMetadata == nil {
		return false
	}
	if len(metadata) != len(inMetadata) {
		return true
	}
	return !reflect.DeepEqual(metadata, inMetadata)
}
