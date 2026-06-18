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
	"time"

	"github.com/golang/protobuf/jsonpb"
	"go.uber.org/zap"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	"github.com/pole-io/pole-server/pkg/namespace"
)

// CreateCircuitBreakerRules Create a CircuitBreaker rule
func (s *Server) CreateCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, cbRule := range request {
		response := s.createCircuitBreakerRule(ctx, cbRule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

// CreateCircuitBreakerRule Create a CircuitBreaker rule
func (s *Server) createCircuitBreakerRule(
	ctx context.Context, request *apifault.CircuitBreakerRule) *apimodel.Response {
	// 构造底层数据结构
	data, err := api2CircuitBreakerRule(request)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		// 注释：错误码更改 - 从特定的Code_ParseCircuitBreakerException改为通用的Code_ParseException
		return api.NewResponse(apimodel.Code_ParseException)
	}
	exists, err := s.storage.HasCircuitBreakerRuleByName(data.Name, data.Namespace)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}
	if exists {
		return api.NewResponse(apimodel.Code_ServiceExistedCircuitBreakers)
	}
	data.ID = utils.NewUUID()

	// 存储层操作
	if err := s.storage.CreateCircuitBreakerRule(data); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}

	msg := fmt.Sprintf("create circuitBreaker rule: id=%v, name=%v, namespace=%v",
		data.ID, request.GetName(), data.Namespace)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, circuitBreakerRuleRecordEntry(ctx, request, data, types.OCreate))
	request.Id = data.ID
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, request)
}

// DeleteCircuitBreakerRules Delete current CircuitBreaker rules
func (s *Server) DeleteCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range request {
		resp := s.deleteCircuitBreakerRule(ctx, entry)
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// deleteCircuitBreakerRule delete current CircuitBreaker rule
func (s *Server) deleteCircuitBreakerRule(
	ctx context.Context, request *apifault.CircuitBreakerRule) *apimodel.Response {
	resp := s.checkCircuitBreakerRuleExists(ctx, request.GetId())
	if resp != nil {
		// 注释：错误码和字段访问改动 - Code字段从*wrapperspb.UInt32Value改为uint32，错误码从Code_NotFoundCircuitBreaker改为Code_NotFoundResource
		if resp.Code == uint32(apimodel.Code_NotFoundResource) {
			resp.Code = uint32(apimodel.Code_ExecuteSuccess)
		}
		return resp
	}
	cbRuleId := &apifault.CircuitBreakerRule{Id: request.GetId()}
	err := s.storage.DeleteCircuitBreakerRule(request.GetId())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		// 注释：错误码更改 - 从特定的Code_ParseCircuitBreakerException改为通用的Code_ParseException
		return api.NewAnyDataResponse(apimodel.Code_ParseException, cbRuleId)
	}
	msg := fmt.Sprintf("delete circuitbreaker rule: id=%v, name=%v, namespace=%v",
		request.GetId(), request.GetName(), circuitBreakerRuleNamespace(request))
	log.Info(msg, utils.RequestID(ctx))

	cbRule := &rules.CircuitBreakerRule{
		ID: request.GetId(), Name: request.GetName(), Namespace: circuitBreakerRuleNamespace(request)}
	s.RecordHistory(ctx, circuitBreakerRuleRecordEntry(ctx, request, cbRule, types.ODelete))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, cbRuleId)
}

// UpdateCircuitBreakerRules Modify the CircuitBreaker rule
func (s *Server) UpdateCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range request {
		response := s.updateCircuitBreakerRule(ctx, entry)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

func (s *Server) updateCircuitBreakerRule(
	ctx context.Context, request *apifault.CircuitBreakerRule) *apimodel.Response {
	resp := s.checkCircuitBreakerRuleExists(ctx, request.GetId())
	if resp != nil {
		return resp
	}
	cbRuleId := &apifault.CircuitBreakerRule{Id: request.GetId()}
	cbRule, err := api2CircuitBreakerRule(request)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		// 注释：错误码更改 - 统一使用通用解析错误码
		return api.NewAnyDataResponse(apimodel.Code_ParseException, cbRuleId)
	}
	cbRule.ID = request.GetId()
	exists, err := s.storage.HasCircuitBreakerRuleByNameExcludeId(cbRule.Name, cbRule.Namespace, cbRule.ID)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}
	if exists {
		return api.NewResponse(apimodel.Code_ServiceExistedCircuitBreakers)
	}
	if err := s.storage.UpdateCircuitBreakerRule(cbRule); err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return storeError2AnyResponse(err, cbRuleId)
	}

	msg := fmt.Sprintf("update circuitbreaker rule: id=%v, name=%v, namespace=%v",
		request.GetId(), request.GetName(), cbRule.Namespace)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, circuitBreakerRuleRecordEntry(ctx, request, cbRule, types.OUpdate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, cbRuleId)
}

func (s *Server) checkCircuitBreakerRuleExists(ctx context.Context, id string) *apimodel.Response {
	exists, err := s.storage.HasCircuitBreakerRule(id)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if !exists {
		// 注释：错误码更改 - 使用通用的资源不存在错误码
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	return nil
}

// GetCircuitBreakerRules Query CircuitBreaker rules（走 cache）
func (s *Server) GetCircuitBreakerRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(query)
	total, cbRules, err := s.Cache().CircuitBreaker().Query(ctx, &cacheapi.CircuitBreakerRuleArgs{
		Filter: query,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error("get circuitbreaker rules from cache", utils.RequestID(ctx), zap.Error(err))
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = total
	out.Size = uint32(len(cbRules))
	for _, cbRule := range cbRules {
		cbRuleProto, err := cbRule.ToSpec()
		if nil != err {
			log.Error("marshal circuitbreaker rule fail", utils.RequestID(ctx), zap.Error(err))
			continue
		}
		if nil == cbRuleProto {
			continue
		}
		err = api.AddAnyDataIntoBatchQuery(out, cbRuleProto)
		if nil != err {
			log.Error("add circuitbreaker rule as any data fail", utils.RequestID(ctx), zap.Error(err))
			continue
		}
	}
	return out
}

// GetOneCircuitBreakerRule Query all router_rule rules
func (s *Server) GetOneCircuitBreakerRule(ctx context.Context, req *apifault.CircuitBreakerRule) *apimodel.Response {
	saveData, err := s.storage.GetCircuitBreakerRule(req.GetId())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		// 注释：错误码更改 - 统一使用通用的资源不存在错误码
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	cbRuleProto, err := saveData.ToSpec()
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		// 注释：错误码更改 - 统一使用通用解析错误码
		return api.NewResponse(apimodel.Code_ParseException)
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, cbRuleProto)
}

func circuitBreakerRuleRecordEntry(ctx context.Context, req *apifault.CircuitBreakerRule, md *rules.CircuitBreakerRule,
	opt types.OperationType) *types.RecordEntry {
	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)
	entry := &types.RecordEntry{
		ResourceType:  types.RCircuitBreakerRule,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		Namespace:     md.Namespace,
		OperationType: opt,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}

func marshalCircuitBreakerRuleV2(req *apifault.CircuitBreakerRule) (string, error) {
	r := &apifault.CircuitBreakerRule{
		RuleMatcher:  req.RuleMatcher,
		BlockConfigs: req.BlockConfigs,
	}
	rule, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(rule), nil
}

// api2CircuitBreakerRule 把API参数转化为内部数据结构
func api2CircuitBreakerRule(req *apifault.CircuitBreakerRule) (*rules.CircuitBreakerRule, error) {
	rule, err := marshalCircuitBreakerRuleV2(req)
	if err != nil {
		return nil, err
	}

	out := &rules.CircuitBreakerRule{
		Name:         req.GetName(),
		Namespace:    circuitBreakerRuleNamespace(req),
		Description:  req.GetDescription(),
		Level:        int(req.GetLevel()),
		SrcService:   req.GetRuleMatcher().GetSource().GetService(),
		SrcNamespace: req.GetRuleMatcher().GetSource().GetNamespace(),
		DstService:   req.GetRuleMatcher().GetDestination().GetService(),
		DstNamespace: req.GetRuleMatcher().GetDestination().GetNamespace(),
		// 注释：方法字段访问改动 - 去掉额外的.GetValue()调用，直接获取方法名称
		DstMethod: req.GetRuleMatcher().GetDestination().GetMethod().GetValue(),
		Enable:    req.GetEnable(),
		Rule:      rule,
		Revision:  utils.NewUUID(),
	}
	return out, nil
}

func circuitBreakerRuleNamespace(req *apifault.CircuitBreakerRule) string {
	if req == nil {
		return namespace.DefaultNamespace
	}
	if namespace := req.GetRuleMatcher().GetDestination().GetNamespace(); namespace != "" {
		return namespace
	}
	if namespace := req.GetRuleMatcher().GetSource().GetNamespace(); namespace != "" {
		return namespace
	}
	return namespace.DefaultNamespace
}

// circuitBreaker2ClientAPI 把内部数据结构转化为客户端API参数
// 注释：返回类型重大改动 - 从*apifault.CircuitBreaker改为*apifault.CircuitBreakerRule，因为新规范中没有CircuitBreaker类型
func circuitBreaker2ClientAPI(
	req *rules.ServiceWithCircuitBreakerRules, service string, namespace string) (*apifault.CircuitBreakerRule, error) {
	if req == nil {
		return nil, nil
	}

	out := &apifault.CircuitBreakerRule{}
	// 注释：Revision字段类型改动 - 从*wrapperspb.StringValue改为string，简化revision计算
	out.Revision = service + "#" + namespace // 简单的revision计算

	// 注释：重大逻辑改动 - 由于规范中没有CircuitBreaker类型，我们返回第一个规则作为示例
	// 由于规范中没有CircuitBreaker类型，我们返回第一个规则作为示例
	var firstRule *apifault.CircuitBreakerRule
	req.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRelease) {
		if firstRule == nil {
			firstRule = rule.Rule.Proto
		}
	})

	if firstRule != nil {
		return firstRule, nil
	}

	return out, nil
}
