package paramcheck

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

func (svr *Server) CreateLogicalServices(
	ctx context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	if response := validateLogicalServices(reqs, false); response != nil {
		return response
	}
	return svr.nextSvr.CreateLogicalServices(ctx, reqs)
}

func (svr *Server) UpdateLogicalServices(
	ctx context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	if response := validateLogicalServices(reqs, true); response != nil {
		return response
	}
	return svr.nextSvr.UpdateLogicalServices(ctx, reqs)
}

func (svr *Server) DeleteLogicalServices(
	ctx context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	for _, req := range reqs {
		if req.GetId() == "" {
			return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
		}
	}
	return svr.nextSvr.DeleteLogicalServices(ctx, reqs)
}

func validateLogicalServices(
	reqs []*apiservice.LogicalService, requireIdentity bool) *apimodel.BatchWriteResponse {
	for _, req := range reqs {
		if req == nil || valid.CheckResourceName(req.GetName()) != nil ||
			(requireIdentity && (req.GetId() == "" || req.GetRevision() == "")) {
			return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
		}
	}
	return nil
}

func (svr *Server) GetLogicalServices(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return svr.nextSvr.GetLogicalServices(ctx, query)
}

func (svr *Server) GetLogicalServiceEnvironments(
	ctx context.Context, logicalServiceID string) *apimodel.BatchQueryResponse {
	if logicalServiceID == "" {
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	return svr.nextSvr.GetLogicalServiceEnvironments(ctx, logicalServiceID)
}

func (svr *Server) GetUnboundServiceEnvironments(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return svr.nextSvr.GetUnboundServiceEnvironments(ctx, query)
}

func (svr *Server) GetServiceEnvironmentBinding(
	ctx context.Context, serviceID string) *apimodel.Response {
	if serviceID == "" {
		return api.NewResponse(apimodel.Code_InvalidParameter)
	}
	return svr.nextSvr.GetServiceEnvironmentBinding(ctx, serviceID)
}

func (svr *Server) BindServiceEnvironment(
	ctx context.Context, req *apiservice.BindServiceEnvironmentRequest) *apimodel.Response {
	if req == nil || req.GetLogicalServiceId() == "" || req.GetServiceId() == "" {
		return api.NewResponse(apimodel.Code_InvalidParameter)
	}
	return svr.nextSvr.BindServiceEnvironment(ctx, req)
}

func (svr *Server) UnbindServiceEnvironment(
	ctx context.Context, req *apiservice.UnbindServiceEnvironmentRequest) *apimodel.Response {
	if req == nil || req.GetLogicalServiceId() == "" || req.GetServiceId() == "" {
		return api.NewResponse(apimodel.Code_InvalidParameter)
	}
	return svr.nextSvr.UnbindServiceEnvironment(ctx, req)
}
