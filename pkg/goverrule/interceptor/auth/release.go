package goverrule_auth

import (
	"context"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

// PublishLaneGroups 发布多个治理规则
func (svr *Server) PublishGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	var rsp *apiservice.BatchWriteResponse
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.PublishLaneGroups)
	case apimodel.RuleRelease_CircuitBreakerRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.PublishCircuitBreakerRules)
	case apimodel.RuleRelease_FaultDetectRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.PublishFaultDetectRules)
	case apimodel.RuleRelease_RouteRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.PublishRouteRules)
	case apimodel.RuleRelease_RateLimitRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.PublishRateLimitRules)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}

	if rsp != nil {
		return rsp
	}
	return svr.nextSvr.PublishGovernanceRules(ctx, req)
}

func (svr *Server) GetRuleReleases(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse {
	ruleId := filter["rule_id"]
	resource := apimodel.RuleRelease_RuleType(apimodel.RuleRelease_RuleType_value[filter["resource"]])

	reqs := []*apimodel.RuleRelease{
		{
			RuleId:   ruleId,
			Resource: resource,
		},
	}

	var bRsp *apiservice.BatchWriteResponse
	switch resource {
	case apimodel.RuleRelease_LaneRules:
		ctx, bRsp = svr.checkGovernanceRules(ctx, reqs, authtypes.Read, authtypes.DescribeLaneGroupReleases)
	case apimodel.RuleRelease_CircuitBreakerRules:
		ctx, bRsp = svr.checkGovernanceRules(ctx, reqs, authtypes.Read, authtypes.DescribeCircuitBreakerReleases)
	case apimodel.RuleRelease_FaultDetectRules:
		ctx, bRsp = svr.checkGovernanceRules(ctx, reqs, authtypes.Read, authtypes.DescribeFaultDetectReleases)
	case apimodel.RuleRelease_RouteRules:
		ctx, bRsp = svr.checkGovernanceRules(ctx, reqs, authtypes.Read, authtypes.DescribeRouteReleases)
	case apimodel.RuleRelease_RateLimitRules:
		ctx, bRsp = svr.checkGovernanceRules(ctx, reqs, authtypes.Read, authtypes.DescribeRateLimitReleases)
	default:
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	if bRsp != nil {
		return api.NewBatchQueryResponseWithMsg(apimodel.Code(bRsp.GetCode().GetValue()), bRsp.GetInfo().GetValue())
	}

	return svr.nextSvr.GetRuleReleases(ctx, filter)
}

// DeleteLaneGroups 删除多个治理规则已发布版本
func (svr *Server) DeleteGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	var rsp *apiservice.BatchWriteResponse
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Delete, authtypes.DeleteLaneGroupReleases)
	case apimodel.RuleRelease_CircuitBreakerRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Delete, authtypes.DeleteCircuitBreakerReleases)
	case apimodel.RuleRelease_FaultDetectRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Delete, authtypes.DeleteFaultDetectReleases)
	case apimodel.RuleRelease_RouteRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Delete, authtypes.DeleteRouteReleases)
	case apimodel.RuleRelease_RateLimitRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Delete, authtypes.DeleteRateLimitReleases)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}

	if rsp != nil {
		return rsp
	}
	return svr.nextSvr.DeleteGovernanceRules(ctx, req)
}

// RollbackLaneGroups 回滚多个治理规则到目标版本
func (svr *Server) RollbackGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	var rsp *apiservice.BatchWriteResponse
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.RollbackLaneGroups)
	case apimodel.RuleRelease_CircuitBreakerRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.RollbackCircuitBreakerRules)
	case apimodel.RuleRelease_FaultDetectRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.RollbackFaultDetectRules)
	case apimodel.RuleRelease_RouteRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.RollbackRouteRules)
	case apimodel.RuleRelease_RateLimitRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.RollbackRateLimitRules)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}

	if rsp != nil {
		return rsp
	}
	return svr.nextSvr.RollbackGovernanceRules(ctx, req)
}

// StopbetaLaneGroups 停止多个治理规则灰度发布版本
func (svr *Server) StopbetaGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	var rsp *apiservice.BatchWriteResponse
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.StopbetaLaneGroups)
	case apimodel.RuleRelease_CircuitBreakerRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.StopbetaCircuitBreakerRules)
	case apimodel.RuleRelease_FaultDetectRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.StopbetaFaultDetectRules)
	case apimodel.RuleRelease_RouteRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.StopbetaRouteRules)
	case apimodel.RuleRelease_RateLimitRules:
		ctx, rsp = svr.checkGovernanceRules(ctx, req, authtypes.Modify, authtypes.StopbetaRateLimitRules)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}

	if rsp != nil {
		return rsp
	}
	return svr.nextSvr.StopbetaGovernanceRules(ctx, req)
}

func (svr *Server) checkGovernanceRules(
	ctx context.Context,
	request []*apimodel.RuleRelease,
	op authtypes.ResourceOperation,
	fname authtypes.ServerFunctionName) (context.Context, *apiservice.BatchWriteResponse) {
	authCtx := svr.collectRuleReleases(ctx, request, op, fname)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return nil, api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return ctx, nil
}
