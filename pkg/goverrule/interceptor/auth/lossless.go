package goverrule_auth

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/security"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

func (s *Server) CreateLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apimodel.BatchWriteResponse {
	authCtx := s.collectLosslessAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateLosslessRules)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	rsp := s.nextSvr.CreateLossLessRules(ctx, reqs)
	for index := range rsp.Responses {
		item := rsp.GetResponses()[index].GetData()
		rule := &apitraffic.LosslessRule{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = s.afterRuleResource(ctx, types.RLosslessRule, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_LosslessRules,
		}, false)
	}
	return rsp
}

func (s *Server) DeleteLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apimodel.BatchWriteResponse {
	authCtx := s.collectLosslessAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteLosslessRules)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	rsp := s.nextSvr.DeleteLossLessRules(ctx, reqs)
	for index := range rsp.Responses {
		item := rsp.GetResponses()[index].GetData()
		rule := &apitraffic.LosslessRule{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = s.afterRuleResource(ctx, types.RLosslessRule, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_LosslessRules,
		}, true)
	}
	return rsp
}

func (s *Server) UpdateLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apimodel.BatchWriteResponse {
	authCtx := s.collectLosslessAuthContext(ctx, reqs, authtypes.Create, authtypes.UpdateLosslessRules)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextSvr.UpdateLossLessRules(ctx, reqs)
}

func (s *Server) GetLossLessRules(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	authCtx := s.collectLosslessAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeLosslessRules)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendLosslessRulePredicate(ctx, func(ctx context.Context, cbr *rules.LosslessRule) bool {
		return s.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_LosslessRules,
			ID:       cbr.ID,
			Metadata: cbr.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := s.nextSvr.GetLossLessRules(ctx, filter)

	for index := range resp.Data {
		item := &apitraffic.LosslessRule{}
		_ = anypb.UnmarshalTo(resp.Data[index], item, proto.UnmarshalOptions{})
		item.Editable = true
		item.Deleteable = true
		authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
			security.ResourceType_LosslessRules: {
				{
					Type:     apisecurity.ResourceType_LosslessRules,
					ID:       item.GetId(),
					Metadata: item.Metadata,
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateLosslessRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = false
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteLosslessRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = false
		}
		_ = anypb.MarshalFrom(resp.Data[index], item, proto.MarshalOptions{})
	}
	return resp
}

func (s *Server) GetOneLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apimodel.Response {
	authCtx := s.collectLosslessAuthContext(ctx, []*apitraffic.LosslessRule{req}, authtypes.Read, authtypes.DescribeLosslessRules)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendLosslessRulePredicate(ctx, func(ctx context.Context, cbr *rules.LosslessRule) bool {
		return s.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_LosslessRules,
			ID:       cbr.ID,
			Metadata: cbr.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := s.nextSvr.GetOneLossLessRule(ctx, req)
	rule := &apitraffic.LosslessRule{}
	_ = anypb.UnmarshalTo(resp.Data, rule, proto.UnmarshalOptions{})
	rule.Editable = true
	rule.Deleteable = true

	authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
		security.ResourceType_LosslessRules: {
			{
				Type:     apisecurity.ResourceType_LosslessRules,
				ID:       rule.GetId(),
				Metadata: rule.Metadata,
			},
		},
	})

	// 检查 write 操作权限
	authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateLaneGroups, authtypes.PublishLaneGroups})
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		rule.Editable = false
	}

	// 检查 delete 操作权限
	authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteLaneGroups})
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		rule.Deleteable = false
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, rule)
}
