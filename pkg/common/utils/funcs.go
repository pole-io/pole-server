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

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

var emptyVal = struct{}{}

// Int2bool 整数转换为bool值
func Int2bool(entry int) bool {
	return entry != 0
}

func BoolPtr(v bool) *bool {
	return &v
}

// StatusBoolToInt 状态bool转int
func StatusBoolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// ConvertFilter map[string]string to  map[string][]string
func ConvertFilter(filters map[string]string) map[string][]string {
	newFilters := make(map[string][]string)

	for k, v := range filters {
		val := make([]string, 0)
		val = append(val, v)
		newFilters[k] = val
	}

	return newFilters
}

// CollectMapKeys collect filters key to slice
func CollectMapKeys(filters map[string]string) []string {
	fields := make([]string, 0, len(filters))
	for k := range filters {
		fields = append(fields, k)
		if k != "" {
			fields = append(fields, strings.ToUpper(string(k[:1]))+k[1:])
		}
	}

	return fields
}

// IsPrefixWildName 判断名字是否为通配名字，只支持前缀索引(名字最后为*)
func IsPrefixWildName(name string) bool {
	length := len(name)
	return length >= 1 && name[length-1:length] == "*"
}

// IsWildName 判断名字是否为通配名字，前缀或者后缀
func IsWildName(name string) bool {
	return IsPrefixWildName(name) || IsSuffixWildName(name)
}

// ParseWildNameForSql 如果 name 是通配字符串，将通配字符*替换为sql中的%
func ParseWildNameForSql(name string) string {
	if IsPrefixWildName(name) {
		name = name[:len(name)-1] + "%"
	}
	if IsSuffixWildName(name) {
		name = "%" + name[1:]
	}
	return name
}

// IsSuffixWildName 判断名字是否为通配名字，只支持后缀索引(名字第一个字符为*)
func IsSuffixWildName(name string) bool {
	length := len(name)
	return length >= 1 && name[0:1] == "*"
}

// ParseWildName 判断是否为格式化查询条件并且返回真正的查询信息
func ParseWildName(name string) (string, bool) {
	length := len(name)
	ok := length >= 1 && name[length-1:length] == "*"

	if ok {
		return name[:len(name)-1], ok
	}

	return name, false
}

// IsWildMatchIgnoreCase 判断 name 是否匹配 pattern，pattern 可以是前缀或者后缀，忽略大小写
func IsWildMatchIgnoreCase(name, pattern string) bool {
	return IsWildMatch(strings.ToLower(name), strings.ToLower(pattern))
}

// IsWildNotMatch .
func IsWildNotMatch(name, pattern string) bool {
	return !IsWildMatch(name, pattern)
}

// IsWildMatch 判断 name 是否匹配 pattern，pattern 可以是前缀或者后缀
func IsWildMatch(name, pattern string) bool {
	if IsPrefixWildName(pattern) {
		pattern = strings.TrimRight(pattern, "*")
		if strings.HasPrefix(name, pattern) {
			return true
		}
		if IsSuffixWildName(pattern) {
			pattern = strings.TrimLeft(pattern, "*")
			return strings.Contains(name, pattern)
		}
		return false
	} else if IsSuffixWildName(pattern) {
		pattern = strings.TrimLeft(pattern, "*")
		if strings.HasSuffix(name, pattern) {
			return true
		}
		return false
	}
	return pattern == name
}

// NewUUID 返回一个随机的UUID
func NewUUID() string {
	uuidBytes := uuid.New()
	return hex.EncodeToString(uuidBytes[:])
}

// NewUUID 返回一个随机的UUID
func NewRoutingV2UUID() string {
	uuidBytes := uuid.New()
	return hex.EncodeToString(uuidBytes[:])
}

// NewRevision .
func NewRevision() string {
	uuidBytes := uuid.New()
	return hex.EncodeToString(uuidBytes[:])
}

func DefaultString(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

// StringSliceDeDuplication 字符切片去重
func StringSliceDeDuplication(s []string) []string {
	m := make(map[string]struct{}, len(s))
	res := make([]string, 0, len(s))
	for k := range s {
		if _, ok := m[s[k]]; !ok {
			m[s[k]] = emptyVal
			res = append(res, s[k])
		}
	}

	return res
}

func MustJson(v interface{}) string {
	data, err := json.Marshal(v)
	_ = err
	return string(data)
}

// IsNotEqualMap metadata need update
func IsNotEqualMap(req map[string]string, old map[string]string) bool {
	if req == nil {
		return false
	}

	if len(req) != len(old) {
		return true
	}

	needUpdate := false
	for key, value := range req {
		oldValue, ok := old[key]
		if !ok {
			needUpdate = true
			break
		}
		if value != oldValue {
			needUpdate = true
			break
		}
	}
	if needUpdate {
		return needUpdate
	}

	for key, value := range old {
		newValue, ok := req[key]
		if !ok {
			needUpdate = true
			break
		}
		if value != newValue {
			needUpdate = true
			break
		}
	}

	return needUpdate
}

// ConvertGRPCContext 将GRPC上下文转换成内部上下文
func ConvertGRPCContext(ctx context.Context) context.Context {
	var requestID, userAgent, token string

	meta, exist := metadata.FromIncomingContext(ctx)
	if exist {
		ids := meta["request-id"]
		if len(ids) > 0 {
			requestID = ids[0]
		}
		agents := meta["user-agent"]
		if len(agents) > 0 {
			userAgent = agents[0]
		}
		if tokens := meta["x-polaris-token"]; len(tokens) > 0 {
			token = tokens[0]
		}
	} else {
		meta = metadata.MD{}
	}

	var (
		clientIP = ""
		address  = ""
	)
	if pr, ok := peer.FromContext(ctx); ok && pr.Addr != nil {
		address = pr.Addr.String()
		addrSlice := strings.Split(address, ":")
		if len(addrSlice) == 2 {
			clientIP = addrSlice[0]
		}
	}

	ctx = context.Background()
	ctx = context.WithValue(ctx, types.ContextGrpcHeader, meta)
	ctx = context.WithValue(ctx, types.ContextRequestHeaders, meta)
	ctx = context.WithValue(ctx, types.ContextRequestId, requestID)
	ctx = context.WithValue(ctx, types.ContextClientIP, clientIP)
	ctx = context.WithValue(ctx, types.ContextClientAddress, address)
	ctx = context.WithValue(ctx, types.ContextUserAgent, userAgent)
	ctx = context.WithValue(ctx, types.ContextAuthTokenKey, token)

	return ctx
}

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
