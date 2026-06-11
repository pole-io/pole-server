package goverrule_auth

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

func (s *Server) CreateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Create, authtypes.CreateTrafficSecurityRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficSecurityRules(ctx, req)
}
func (s *Server) UpdateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Modify, authtypes.UpdateTrafficSecurityRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficSecurityRules(ctx, req)
}
func (s *Server) DeleteTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Delete, authtypes.DeleteTrafficSecurityRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.DeleteTrafficSecurityRules(ctx, req)
}
func (s *Server) GetTrafficSecurityRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficSecurityRules(ctx, query)
}
func (s *Server) GetOneTrafficSecurityRule(ctx context.Context, req *apisecurity.TrafficSecurityRule) *apimodel.Response {
	return s.nextSvr.GetOneTrafficSecurityRule(ctx, req)
}

func (s *Server) CreateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Create, authtypes.CreateTrafficMirrorRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficMirrorRules(ctx, req)
}
func (s *Server) UpdateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Modify, authtypes.UpdateTrafficMirrorRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficMirrorRules(ctx, req)
}
func (s *Server) DeleteTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Delete, authtypes.DeleteTrafficMirrorRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.DeleteTrafficMirrorRules(ctx, req)
}
func (s *Server) GetTrafficMirrorRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficMirrorRules(ctx, query)
}
func (s *Server) GetOneTrafficMirrorRule(ctx context.Context, req *apitraffic.TrafficMirror) *apimodel.Response {
	return s.nextSvr.GetOneTrafficMirrorRule(ctx, req)
}

func (s *Server) CreateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Create, authtypes.CreateTrafficMockRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficMockRules(ctx, req)
}
func (s *Server) UpdateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Modify, authtypes.UpdateTrafficMockRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficMockRules(ctx, req)
}
func (s *Server) DeleteTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Delete, authtypes.DeleteTrafficMockRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.DeleteTrafficMockRules(ctx, req)
}
func (s *Server) GetTrafficMockRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficMockRules(ctx, query)
}
func (s *Server) GetOneTrafficMockRule(ctx context.Context, req *apitraffic.TrafficMock) *apimodel.Response {
	return s.nextSvr.GetOneTrafficMockRule(ctx, req)
}

func (s *Server) checkSimpleTrafficPermission(ctx context.Context, op authtypes.ResourceOperation, method authtypes.ServerFunctionName) *apimodel.BatchWriteResponse {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(op),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(method),
	)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return nil
}
