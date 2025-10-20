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
	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
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
		return api.NewResponse(apimodel.Code_ParseCircuitBreakerException)
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
		data.ID, request.GetName(), request.GetNamespace())
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
		if resp.GetCode().GetValue() == uint32(apimodel.Code_NotFoundCircuitBreaker) {
			resp.Code = &wrappers.UInt32Value{Value: uint32(apimodel.Code_ExecuteSuccess)}
		}
		return resp
	}
	cbRuleId := &apifault.CircuitBreakerRule{Id: request.GetId()}
	err := s.storage.DeleteCircuitBreakerRule(request.GetId())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewAnyDataResponse(apimodel.Code_ParseCircuitBreakerException, cbRuleId)
	}
	msg := fmt.Sprintf("delete circuitbreaker rule: id=%v, name=%v, namespace=%v",
		request.GetId(), request.GetName(), request.GetNamespace())
	log.Info(msg, utils.RequestID(ctx))

	cbRule := &rules.CircuitBreakerRule{
		ID: request.GetId(), Name: request.GetName(), Namespace: request.GetNamespace()}
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
		return api.NewAnyDataResponse(apimodel.Code_ParseCircuitBreakerException, cbRuleId)
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
		request.GetId(), request.GetName(), request.GetNamespace())
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
		return api.NewResponse(apimodel.Code_NotFoundCircuitBreaker)
	}
	return nil
}

// GetCircuitBreakerRules Query CircuitBreaker rules
func (s *Server) GetCircuitBreakerRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(query)
	total, cbRules, err := s.storage.GetCircuitBreakerRules(query, offset, limit)
	if err != nil {
		log.Error("get circuitbreaker rules store", utils.RequestID(ctx), zap.Error(err))
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = protobuf.NewUInt32Value(total)
	out.Size = protobuf.NewUInt32Value(uint32(len(cbRules)))
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
		return api.NewResponse(apimodel.Code_NotFoundCircuitBreaker)
	}
	cbRuleProto, err := saveData.ToSpec()
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewResponse(apimodel.Code_ParseCircuitBreakerException)
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
		Namespace:     req.GetNamespace(),
		OperationType: opt,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}

func marshalCircuitBreakerRuleV2(req *apifault.CircuitBreakerRule) (string, error) {
	r := &apifault.CircuitBreakerRule{
		RuleMatcher:        req.RuleMatcher,
		ErrorConditions:    req.ErrorConditions,
		TriggerCondition:   req.TriggerCondition,
		MaxEjectionPercent: req.MaxEjectionPercent,
		RecoverCondition:   req.RecoverCondition,
		FaultDetectConfig:  req.FaultDetectConfig,
		FallbackConfig:     req.FallbackConfig,
		BlockConfigs:       req.BlockConfigs,
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
		Namespace:    req.GetNamespace(),
		Description:  req.GetDescription(),
		Level:        int(req.GetLevel()),
		SrcService:   req.GetRuleMatcher().GetSource().GetService(),
		SrcNamespace: req.GetRuleMatcher().GetSource().GetNamespace(),
		DstService:   req.GetRuleMatcher().GetDestination().GetService(),
		DstNamespace: req.GetRuleMatcher().GetDestination().GetNamespace(),
		DstMethod:    req.GetRuleMatcher().GetDestination().GetMethod().GetValue().GetValue(),
		Enable:       req.GetEnable(),
		Rule:         rule,
		Revision:     utils.NewUUID(),
	}
	if out.Namespace == "" {
		out.Namespace = namespace.DefaultNamespace
	}
	return out, nil
}

// circuitBreaker2ClientAPI 把内部数据结构转化为客户端API参数
func circuitBreaker2ClientAPI(
	req *rules.ServiceWithCircuitBreakerRules, service string, namespace string) (*apifault.CircuitBreaker, error) {
	if req == nil {
		return nil, nil
	}

	out := &apifault.CircuitBreaker{}
	out.Revision = &wrappers.StringValue{Value: req.Revision}
	out.Rules = make([]*apifault.CircuitBreakerRule, 0, req.CountCircuitBreakerRules())
	var iterateErr error
	req.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRelease) {
		out.Rules = append(out.Rules, rule.Rule.Proto)
	})
	if nil != iterateErr {
		return nil, iterateErr
	}

	out.Service = protobuf.NewStringValue(service)
	out.ServiceNamespace = protobuf.NewStringValue(namespace)

	return out, nil
}
