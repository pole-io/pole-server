package paramcheck

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func (s *Server) CreateLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.CreateLossLessRules(ctx, reqs)
}

func (s *Server) DeleteLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteLossLessRules(ctx, reqs)
}

func (s *Server) UpdateLossLessRules(ctx context.Context, reqs []*apitraffic.LosslessRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.UpdateLossLessRules(ctx, reqs)
}

func (s *Server) GetLossLessRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetLossLessRules(ctx, query)
}

func (s *Server) GetOneLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apimodel.Response {
	return s.nextSvr.GetOneLossLessRule(ctx, req)
}
