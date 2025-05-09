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

package goverrule

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	commontime "github.com/pole-io/pole-server/pkg/common/time"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
)

var (
	// RateLimitFilters rate limit filters
	RateLimitFilters = map[string]bool{
		"id":        true,
		"name":      true,
		"service":   true,
		"namespace": true,
		"brief":     true,
		"method":    true,
		"labels":    true,
		"disable":   true,
		"offset":    true,
		"limit":     true,
	}
)

// CreateRateLimits 批量创建限流规则
func (s *Server) CreateRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, rateLimit := range request {
		response := s.CreateRateLimit(ctx, rateLimit)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

// CreateRateLimit 创建限流规则
func (s *Server) CreateRateLimit(ctx context.Context, req *apitraffic.Rule) *apiservice.Response {
	// 构造底层数据结构
	data, err := api2RateLimit(req, nil)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewRateLimitResponse(apimodel.Code_ParseRateLimitException, req)
	}

	// 存储层操作
	if err := s.storage.CreateRateLimit(data); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperRateLimitStoreResponse(req, err)
	}

	msg := fmt.Sprintf("create rate limit rule: id=%v, namespace=%v, service=%v, name=%v",
		data.ID, req.GetNamespace().GetValue(), req.GetService().GetValue(), req.GetName().GetValue())
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, rateLimitRecordEntry(ctx, req, data, types.OCreate))
	req.Id = protobuf.NewStringValue(data.ID)
	return api.NewRateLimitResponse(apimodel.Code_ExecuteSuccess, req)
}

// DeleteRateLimits 批量删除限流规则
func (s *Server) DeleteRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range request {
		resp := s.DeleteRateLimit(ctx, entry)
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// DeleteRateLimit 删除单个限流规则
func (s *Server) DeleteRateLimit(ctx context.Context, req *apitraffic.Rule) *apiservice.Response {
	// 检查限流规则是否存在
	rateLimit, resp := s.checkRateLimitExisted(ctx, req.GetId().GetValue(), req)
	if resp != nil {
		if resp.GetCode().GetValue() == uint32(apimodel.Code_NotFoundRateLimit) {
			return api.NewRateLimitResponse(apimodel.Code_ExecuteSuccess, req)
		}
		return resp
	}

	// 生成新的revision
	rateLimit.Revision = utils.NewUUID()

	// 存储层操作
	if err := s.storage.DeleteRateLimit(rateLimit); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperRateLimitStoreResponse(req, err)
	}

	msg := fmt.Sprintf("delete rate limit rule: id=%v, namespace=%v, service=%v, name=%v",
		rateLimit.ID, req.GetNamespace().GetValue(), req.GetService().GetValue(), rateLimit.Labels)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx,
		rateLimitRecordEntry(ctx, req, rateLimit, types.ODelete))
	return api.NewRateLimitResponse(apimodel.Code_ExecuteSuccess, req)
}

func (s *Server) EnableRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range request {
		response := s.EnableRateLimit(ctx, entry)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

// EnableRateLimit 启用限流规则
func (s *Server) EnableRateLimit(ctx context.Context, req *apitraffic.Rule) *apiservice.Response {
	// 检查限流规则是否存在
	data, resp := s.checkRateLimitExisted(ctx, req.GetId().GetValue(), req)
	if resp != nil {
		return resp
	}

	// 构造底层数据结构
	rateLimit := &rules.RateLimit{}
	rateLimit.ID = data.ID
	rateLimit.ServiceID = data.ServiceID
	rateLimit.Disable = req.GetDisable().GetValue()
	rateLimit.Revision = utils.NewUUID()

	if err := s.storage.EnableRateLimit(rateLimit); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperRateLimitStoreResponse(req, err)
	}

	msg := fmt.Sprintf("enable rate limit: id=%v, disable=%v",
		rateLimit.ID, rateLimit.Disable)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, rateLimitRecordEntry(ctx, req, rateLimit, types.OUpdateEnable))
	return api.NewRateLimitResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateRateLimits 批量更新限流规则
func (s *Server) UpdateRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range request {
		response := s.UpdateRateLimit(ctx, entry)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

// UpdateRateLimit 更新限流规则
func (s *Server) UpdateRateLimit(ctx context.Context, req *apitraffic.Rule) *apiservice.Response {
	// 检查限流规则是否存在
	data, resp := s.checkRateLimitExisted(ctx, req.GetId().GetValue(), req)
	if resp != nil {
		return resp
	}

	// 构造底层数据结构
	rateLimit, err := api2RateLimit(req, data)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewRateLimitResponse(apimodel.Code_ParseRateLimitException, req)
	}
	rateLimit.ID = data.ID
	if err := s.storage.UpdateRateLimit(rateLimit); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return wrapperRateLimitStoreResponse(req, err)
	}

	msg := fmt.Sprintf("update rate limit: id=%v, namespace=%v, service=%v, name=%v",
		rateLimit.ID, req.GetNamespace().GetValue(), req.GetService().GetValue(), rateLimit.Name)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, rateLimitRecordEntry(ctx, req, rateLimit, types.OUpdate))
	return api.NewRateLimitResponse(apimodel.Code_ExecuteSuccess, req)
}

// GetRateLimits 查询限流规则
func (s *Server) GetRateLimits(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	// 处理offset和limit
	args, errResp := parseRateLimitArgs(query)
	if errResp != nil {
		return errResp
	}

	total, extendRateLimits, err := s.Cache().RateLimit().QueryRateLimitRules(ctx, *args)
	if err != nil {
		log.Error("get rate limits store", zap.Error(err), utils.RequestID(ctx))
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = protobuf.NewUInt32Value(total)
	out.Size = protobuf.NewUInt32Value(uint32(len(extendRateLimits)))
	out.RateLimits = make([]*apitraffic.Rule, 0, len(extendRateLimits))
	for _, item := range extendRateLimits {
		limit, err := rateLimit2Console(item)
		if err != nil {
			log.Error("get rate limits convert", zap.Error(err), utils.RequestID(ctx))
			return api.NewBatchQueryResponse(apimodel.Code_ParseRateLimitException)
		}
		out.RateLimits = append(out.RateLimits, limit)
	}

	return out
}

// GetAllRateLimits Query all router_rule rules
func (s *Server) GetAllRateLimits(ctx context.Context) *apiservice.BatchQueryResponse {
	return nil
}

func parseRateLimitArgs(query map[string]string) (*cacheapi.RateLimitRuleArgs, *apiservice.BatchQueryResponse) {
	for key := range query {
		if _, ok := RateLimitFilters[key]; !ok {
			log.Errorf("params %s is not allowed in querying rate limits", key)
			return nil, api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
		}
	}
	// 处理offset和limit
	offset, limit, err := valid.ParseOffsetAndLimit(query)
	if err != nil {
		return nil, api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	args := &cacheapi.RateLimitRuleArgs{
		Filter:     query,
		ID:         query["id"],
		Name:       query["name"],
		Service:    query["service"],
		Namespace:  query["namespace"],
		Offset:     offset,
		Limit:      limit,
		OrderField: query["order_field"],
		OrderType:  query["order_type"],
	}
	if val, ok := query["disable"]; ok {
		disable, _ := strconv.ParseBool(val)
		args.Disable = &disable
	}

	return args, nil
}

// checkRateLimitValid 检查限流规则是否允许修改/删除
func (s *Server) checkRateLimitValid(ctx context.Context, serviceID string, req *apitraffic.Rule) (
	*svctypes.Service, *apiservice.Response) {
	requestID := utils.ParseRequestID(ctx)

	service, err := s.storage.GetServiceByID(serviceID)
	if err != nil {
		log.Error(err.Error(), utils.ZapRequestID(requestID))
		return nil, api.NewRateLimitResponse(storeapi.StoreCode2APICode(err), req)
	}

	return service, nil
}

// checkRateLimitExisted 检查限流规则是否存在
func (s *Server) checkRateLimitExisted(ctx context.Context, id string,
	req *apitraffic.Rule) (*rules.RateLimit, *apiservice.Response) {

	rateLimit, err := s.storage.GetRateLimitWithID(id)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return nil, api.NewRateLimitResponse(storeapi.StoreCode2APICode(err), req)
	}
	if rateLimit == nil {
		return nil, api.NewRateLimitResponse(apimodel.Code_NotFoundRateLimit, req)
	}
	return rateLimit, nil
}

const (
	defaultRuleAction = "REJECT"
)

// api2RateLimit 把API参数转化为内部数据结构
func api2RateLimit(req *apitraffic.Rule, old *rules.RateLimit) (*rules.RateLimit, error) {
	rule, err := marshalRateLimitRules(req)
	if err != nil {
		return nil, err
	}

	labels := req.GetLabels()
	var labelStr []byte
	if len(labels) > 0 {
		labelStr, err = json.Marshal(labels)
	}

	out := &rules.RateLimit{
		ID:       utils.NewUUID(),
		Name:     req.GetName().GetValue(),
		Method:   req.GetMethod().GetValue().GetValue(),
		Disable:  req.GetDisable().GetValue(),
		Priority: req.GetPriority().GetValue(),
		Labels:   string(labelStr),
		Rule:     rule,
		Revision: utils.NewUUID(),
		Metadata: req.Metadata,
	}
	return out, nil
}

// rateLimit2api 把内部数据结构转化为API参数
func rateLimit2Console(rateLimit *rules.RateLimit) (*apitraffic.Rule, error) {
	if rateLimit == nil {
		return nil, nil
	}
	if len(rateLimit.Rule) > 0 {
		rateLimit = rateLimit.CopyNoProto()
		rateLimit.Proto = &apitraffic.Rule{}
		// 控制台查询的请求
		if err := json.Unmarshal([]byte(rateLimit.Rule), rateLimit.Proto); err != nil {
			return nil, err
		}
		// 存量标签适配到参数列表
		if err := rateLimit.AdaptLabels(); err != nil {
			return nil, err
		}
	}
	rule := &apitraffic.Rule{}
	rule.Id = protobuf.NewStringValue(rateLimit.ID)
	rule.Name = protobuf.NewStringValue(rateLimit.Name)
	rule.Priority = protobuf.NewUInt32Value(rateLimit.Priority)
	rule.Ctime = protobuf.NewStringValue(commontime.Time2String(rateLimit.CreateTime))
	rule.Mtime = protobuf.NewStringValue(commontime.Time2String(rateLimit.ModifyTime))
	rule.Disable = protobuf.NewBoolValue(rateLimit.Disable)
	rule.Metadata = rateLimit.Metadata
	if rateLimit.EnableTime.Year() > 2000 {
		rule.Etime = protobuf.NewStringValue(commontime.Time2String(rateLimit.EnableTime))
	} else {
		rule.Etime = protobuf.NewStringValue("")
	}
	rule.Revision = protobuf.NewStringValue(rateLimit.Revision)
	if nil != rateLimit.Proto {
		copyRateLimitProto(rateLimit, rule)
	} else {
		rule.Method = &apimodel.MatchString{Value: protobuf.NewStringValue(rateLimit.Method)}
	}
	return rule, nil
}

func populateDefaultRuleValue(rule *apitraffic.Rule) {
	if rule.GetAction().GetValue() == "" {
		rule.Action = protobuf.NewStringValue(defaultRuleAction)
	}
}

func copyRateLimitProto(rateLimit *rules.RateLimit, rule *apitraffic.Rule) {
	// copy proto values
	rule.Namespace = rateLimit.Proto.Namespace
	rule.Service = rateLimit.Proto.Service
	rule.Method = rateLimit.Proto.Method
	rule.Arguments = rateLimit.Proto.Arguments
	rule.Labels = rateLimit.Proto.Labels
	rule.Resource = rateLimit.Proto.Resource
	rule.Type = rateLimit.Proto.Type
	rule.Amounts = rateLimit.Proto.Amounts
	rule.RegexCombine = rateLimit.Proto.RegexCombine
	rule.Action = rateLimit.Proto.Action
	rule.Failover = rateLimit.Proto.Failover
	rule.AmountMode = rateLimit.Proto.AmountMode
	rule.Adjuster = rateLimit.Proto.Adjuster
	rule.MaxQueueDelay = rateLimit.Proto.MaxQueueDelay
	populateDefaultRuleValue(rule)
}

// rateLimit2api 把内部数据结构转化为API参数
func rateLimit2Client(
	service string, namespace string, rateLimit *rules.RateLimit) (*apitraffic.Rule, error) {
	if rateLimit == nil {
		return nil, nil
	}

	rule := &apitraffic.Rule{}
	rule.Id = protobuf.NewStringValue(rateLimit.ID)
	rule.Name = protobuf.NewStringValue(rateLimit.Name)
	rule.Service = protobuf.NewStringValue(service)
	rule.Namespace = protobuf.NewStringValue(namespace)
	rule.Priority = protobuf.NewUInt32Value(rateLimit.Priority)
	rule.Revision = protobuf.NewStringValue(rateLimit.Revision)
	rule.Disable = protobuf.NewBoolValue(rateLimit.Disable)
	rule.Metadata = rateLimit.Metadata
	copyRateLimitProto(rateLimit, rule)
	return rule, nil
}

// marshalRateLimitRules 序列化限流规则具体内容
func marshalRateLimitRules(req *apitraffic.Rule) (string, error) {
	r := &apitraffic.Rule{
		Name:          req.GetName(),
		Resource:      req.GetResource(),
		Service:       req.GetService(),
		Namespace:     req.GetNamespace(),
		Type:          req.GetType(),
		Amounts:       req.GetAmounts(),
		Action:        req.GetAction(),
		Disable:       req.GetDisable(),
		Report:        req.GetReport(),
		Adjuster:      req.GetAdjuster(),
		RegexCombine:  req.GetRegexCombine(),
		AmountMode:    req.GetAmountMode(),
		Failover:      req.GetFailover(),
		Arguments:     req.GetArguments(),
		Method:        req.GetMethod(),
		MaxQueueDelay: req.GetMaxQueueDelay(),
	}
	rule, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(rule), nil
}

// rateLimitRecordEntry 构建rateLimit的记录entry
func rateLimitRecordEntry(ctx context.Context, req *apitraffic.Rule, md *rules.RateLimit,
	opt types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RRateLimit,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		Namespace:     req.GetNamespace().GetValue(),
		Operator:      utils.ParseOperator(ctx),
		OperationType: opt,
		Detail:        detail,
		HappenTime:    time.Now(),
	}

	return entry
}

// wrapperRateLimitStoreResponse 封装路由存储层错误
func wrapperRateLimitStoreResponse(rule *apitraffic.Rule, err error) *apiservice.Response {
	if err == nil {
		return nil
	}
	resp := api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	resp.RateLimit = rule
	return resp
}
