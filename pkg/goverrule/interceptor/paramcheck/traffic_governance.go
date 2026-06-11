package paramcheck

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func (s *Server) CreateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.CreateTrafficSecurityRules(ctx, req)
}
func (s *Server) UpdateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.UpdateTrafficSecurityRules(ctx, req)
}
func (s *Server) DeleteTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteTrafficSecurityRules(ctx, req)
}
func (s *Server) GetTrafficSecurityRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficSecurityRules(ctx, query)
}
func (s *Server) GetOneTrafficSecurityRule(ctx context.Context, req *apisecurity.TrafficSecurityRule) *apimodel.Response {
	return s.nextSvr.GetOneTrafficSecurityRule(ctx, req)
}

func (s *Server) CreateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return s.nextSvr.CreateTrafficMirrorRules(ctx, req)
}
func (s *Server) UpdateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return s.nextSvr.UpdateTrafficMirrorRules(ctx, req)
}
func (s *Server) DeleteTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteTrafficMirrorRules(ctx, req)
}
func (s *Server) GetTrafficMirrorRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficMirrorRules(ctx, query)
}
func (s *Server) GetOneTrafficMirrorRule(ctx context.Context, req *apitraffic.TrafficMirror) *apimodel.Response {
	return s.nextSvr.GetOneTrafficMirrorRule(ctx, req)
}

func (s *Server) CreateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return s.nextSvr.CreateTrafficMockRules(ctx, req)
}
func (s *Server) UpdateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return s.nextSvr.UpdateTrafficMockRules(ctx, req)
}
func (s *Server) DeleteTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteTrafficMockRules(ctx, req)
}
func (s *Server) GetTrafficMockRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficMockRules(ctx, query)
}
func (s *Server) GetOneTrafficMockRule(ctx context.Context, req *apitraffic.TrafficMock) *apimodel.Response {
	return s.nextSvr.GetOneTrafficMockRule(ctx, req)
}
