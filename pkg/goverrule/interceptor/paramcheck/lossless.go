package paramcheck

import (
	"context"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func (s *Server) CreateLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse {
	return s.nextSvr.CreateLossLessRules(ctx, reqs)
}

func (s *Server) DeleteLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse {
	return s.nextSvr.DeleteLossLessRules(ctx, reqs)
}

func (s *Server) UpdateLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse {
	return s.nextSvr.UpdateLossLessRules(ctx, reqs)
}

func (s *Server) GetLossLessRules(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	return s.nextSvr.GetLossLessRules(ctx, query)
}

func (s *Server) GetOneLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apiservice.Response {
	return s.nextSvr.GetOneLossLessRule(ctx, req)
}
