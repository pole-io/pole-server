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

// CreateFaultDetectRules Create a FaultDetect rule
func (s *Server) CreateFaultDetectRules(
	ctx context.Context, reqs []*apifault.FaultDetectRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, cbRule := range reqs {
		response := s.createFaultDetectRule(ctx, cbRule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

// DeleteFaultDetectRules Delete current Fault Detect rules
func (s *Server) DeleteFaultDetectRules(
	ctx context.Context, reqs []*apifault.FaultDetectRule) *apimodel.BatchWriteResponse {

	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, cbRule := range reqs {
		response := s.deleteFaultDetectRule(ctx, cbRule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

// UpdateFaultDetectRules Modify the FaultDetect rule
func (s *Server) UpdateFaultDetectRules(
	ctx context.Context, reqs []*apifault.FaultDetectRule) *apimodel.BatchWriteResponse {

	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, cbRule := range reqs {
		response := s.updateFaultDetectRule(ctx, cbRule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

func faultDetectRuleRecordEntry(ctx context.Context, req *apifault.FaultDetectRule, md *rules.FaultDetectRule,
	opt types.OperationType) *types.RecordEntry {
	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)
	entry := &types.RecordEntry{
		ResourceType:  types.RFaultDetectRule,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		Namespace:     md.Namespace,
		OperationType: opt,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}

// createFaultDetectRule Create a FaultDetect rule
func (s *Server) createFaultDetectRule(ctx context.Context, request *apifault.FaultDetectRule) *apimodel.Response {
	data, err := api2FaultDetectRule(request)
	if err != nil {
		log.Error("[faultdetect] parse fault detect rule error", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ParseException)
	}
	exists, err := s.storage.HasFaultDetectRuleByName(data.Name, data.Namespace)
	if err != nil {
		log.Error("[faultdetect] check fault detect rule exists error", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}
	if exists {
		return api.NewResponse(apimodel.Code_ExistedResource)
	}
	data.ID = utils.NewUUID()

	// 存储层操作
	if err := s.storage.CreateFaultDetectRule(data); err != nil {
		log.Error("[faultdetect] create fault detect rule error", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}

	msg := fmt.Sprintf("[faultdetect] create fault detect rule: id=%v, name=%v, namespace=%v",
		data.ID, request.GetName(), data.Namespace)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, faultDetectRuleRecordEntry(ctx, request, data, types.OCreate))
	request.Id = data.ID
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, request)
}

// updateFaultDetectRule Update a FaultDetect rule
func (s *Server) updateFaultDetectRule(ctx context.Context, request *apifault.FaultDetectRule) *apimodel.Response {
	fdRuleId := &apifault.FaultDetectRule{Id: request.GetId()}
	fdRule, err := api2FaultDetectRule(request)
	if err != nil {
		log.Error("[faultdetect] parse fault detect rule error", utils.RequestID(ctx), zap.Error(err))
		return api.NewAnyDataResponse(apimodel.Code_ParseException, fdRuleId)
	}
	fdRule.ID = request.GetId()
	exists, err := s.storage.HasFaultDetectRuleByNameExcludeId(fdRule.Name, fdRule.Namespace, fdRule.ID)
	if err != nil {
		log.Error("[faultdetect] check fault detect rule exists error", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}
	if exists {
		return api.NewAnyDataResponse(apimodel.Code_ExistedResource, fdRuleId)
	}
	if err := s.storage.UpdateFaultDetectRule(fdRule); err != nil {
		log.Error("[faultdetect] update fault detect rule error", utils.RequestID(ctx), zap.Error(err))
		return storeError2AnyResponse(err, fdRuleId)
	}

	msg := fmt.Sprintf("[faultdetect] update fault detect rule: id=%v, name=%v, namespace=%v",
		request.GetId(), request.GetName(), fdRule.Namespace)
	log.Info(msg, utils.RequestID(ctx))

	s.RecordHistory(ctx, faultDetectRuleRecordEntry(ctx, request, fdRule, types.OUpdate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, fdRuleId)
}

// deleteFaultDetectRule Delete a FaultDetect rule
func (s *Server) deleteFaultDetectRule(ctx context.Context, request *apifault.FaultDetectRule) *apimodel.Response {
	cbRuleId := &apifault.FaultDetectRule{Id: request.GetId()}
	err := s.storage.DeleteFaultDetectRule(request.GetId())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewAnyDataResponse(apimodel.Code_ParseException, cbRuleId)
	}
	msg := fmt.Sprintf("[faultdetect] delete fault detect rule: id=%v, name=%v, namespace=%v",
		request.GetId(), request.GetName(), faultDetectRuleNamespace(request))
	log.Info(msg, utils.RequestID(ctx))

	cbRule := &rules.FaultDetectRule{ID: request.GetId(), Name: request.GetName(), Namespace: faultDetectRuleNamespace(request)}
	s.RecordHistory(ctx, faultDetectRuleRecordEntry(ctx, request, cbRule, types.ODelete))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, cbRuleId)
}

func (s *Server) GetFaultDetectRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(query)
	total, cbRules, err := s.caches.FaultDetector().Query(ctx, &cacheapi.FaultDetectArgs{
		Filter: query,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Errorf("[faultdetect] get fault detect rules store err: %s", err.Error())
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = total
	out.Size = uint32(len(cbRules))
	for _, cbRule := range cbRules {
		cbRuleProto, err := cbRule.ToSpec()
		if nil != err {
			log.Error("[faultdetect] marshal circuitbreaker rule fail", utils.RequestID(ctx), zap.Error(err))
			continue
		}
		if nil == cbRuleProto {
			continue
		}
		if err = api.AddAnyDataIntoBatchQuery(out, cbRuleProto); nil != err {
			log.Error("[faultdetect] add circuitbreaker rule as any data fail", utils.RequestID(ctx), zap.Error(err))
			continue
		}
	}
	return out
}

// GetOneFaultDetectRule 查询单个故障检测规则
func (s *Server) GetOneFaultDetectRule(ctx context.Context, req *apifault.FaultDetectRule) *apimodel.Response {
	saveData, err := s.storage.GetFaultDetectRule(req.GetId())
	if err != nil {
		log.Error("[Server][FaultDetect][Query] get fault_detect_rule from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		log.Info("[Server][FaultDetect][Query] fault_detect_rule not found", utils.RequestID(ctx))
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}

	view, err := saveData.ToSpec()
	if err != nil {
		log.Error("[goverrule][faultdetect] fault_detect_rule convert to spec", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ExecuteException)
	}

	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, view)
}

func marshalFaultDetectRule(req *apifault.FaultDetectRule) (string, error) {
	r := &apifault.FaultDetectRule{
		TargetService: req.TargetService,
		Interval:      req.Interval,
		Timeout:       req.Timeout,
		Port:          req.Port,
		Protocol:      req.Protocol,
		HttpConfig:    req.HttpConfig,
		TcpConfig:     req.TcpConfig,
		UdpConfig:     req.UdpConfig,
	}
	rule, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(rule), nil
}

// api2FaultDetectRule 把API参数转化为内部数据结构
func api2FaultDetectRule(req *apifault.FaultDetectRule) (*rules.FaultDetectRule, error) {
	rule, err := marshalFaultDetectRule(req)
	if err != nil {
		return nil, err
	}

	out := &rules.FaultDetectRule{
		Name:         req.GetName(),
		Namespace:    faultDetectRuleNamespace(req),
		Description:  req.GetDescription(),
		DstService:   req.GetTargetService().GetService(),
		DstNamespace: req.GetTargetService().GetNamespace(),
		DstMethod:    req.GetTargetService().GetMethod().GetValue(),
		Rule:         rule,
		Revision:     utils.NewUUID(),
		Metadata:     req.Metadata,
	}
	return out, nil
}

func faultDetectRuleNamespace(req *apifault.FaultDetectRule) string {
	if req == nil {
		return namespace.DefaultNamespace
	}
	if namespace := req.GetTargetService().GetNamespace(); namespace != "" {
		return namespace
	}
	return namespace.DefaultNamespace
}

// faultDetectRule2ClientAPI 把内部数据结构转化为客户端API参数
func faultDetectRule2ClientAPI(req *rules.ServiceWithFaultDetectRules) (*apifault.FaultDetector, error) {
	if req == nil {
		return nil, nil
	}

	out := &apifault.FaultDetector{}
	out.Revision = req.Revision
	out.Rules = make([]*apifault.FaultDetectRule, 0, req.CountFaultDetectRules())
	var iterateErr error
	req.IterateFaultDetectRules(func(rule *rules.FaultDetectRelease) {
		cbRule, err := rule.Rule.ToSpec()
		if err != nil {
			iterateErr = err
			return
		}
		out.Rules = append(out.Rules, cbRule)
	})
	if nil != iterateErr {
		return nil, iterateErr
	}
	return out, nil
}
