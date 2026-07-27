package service_auth

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

func (svr *Server) CreateLogicalServices(
	ctx context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Create, authtypes.CreateLogicalServices)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	owner := utils.ParseOwnerID(ctx)
	if owner != "" {
		for _, req := range reqs {
			req.Owners = owner
		}
	}
	return svr.nextSvr.CreateLogicalServices(ctx, reqs)
}

func (svr *Server) UpdateLogicalServices(
	ctx context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Modify, authtypes.UpdateLogicalServices)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.UpdateLogicalServices(ctx, reqs)
}

func (svr *Server) DeleteLogicalServices(
	ctx context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Delete, authtypes.DeleteLogicalServices)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.DeleteLogicalServices(ctx, reqs)
}

func (svr *Server) GetLogicalServices(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	ctx, response := svr.authorizeLogicalServiceRead(ctx)
	if response != nil {
		return response
	}
	return svr.nextSvr.GetLogicalServices(ctx, query)
}

func (svr *Server) GetLogicalServiceEnvironments(
	ctx context.Context, logicalServiceID string) *apimodel.BatchQueryResponse {
	ctx, response := svr.authorizeLogicalServiceRead(ctx)
	if response != nil {
		return response
	}
	return svr.nextSvr.GetLogicalServiceEnvironments(ctx, logicalServiceID)
}

func (svr *Server) GetUnboundServiceEnvironments(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	ctx, response := svr.authorizeLogicalServiceRead(ctx)
	if response != nil {
		return response
	}
	return svr.nextSvr.GetUnboundServiceEnvironments(ctx, query)
}

func (svr *Server) GetServiceEnvironmentBinding(
	ctx context.Context, serviceID string) *apimodel.Response {
	ctx, response := svr.authorizeLogicalServiceRead(ctx)
	if response != nil {
		return api.NewResponse(apimodel.Code(response.GetCode()))
	}
	return svr.nextSvr.GetServiceEnvironmentBinding(ctx, serviceID)
}

func (svr *Server) BindServiceEnvironment(
	ctx context.Context, req *apiservice.BindServiceEnvironmentRequest) *apimodel.Response {
	if req == nil {
		return api.NewResponse(apimodel.Code_InvalidParameter)
	}
	return svr.authorizeEnvironmentMutation(ctx, req.GetServiceId(), authtypes.BindServiceEnvironments,
		func(ctx context.Context) *apimodel.Response {
			return svr.nextSvr.BindServiceEnvironment(ctx, req)
		})
}

func (svr *Server) UnbindServiceEnvironment(
	ctx context.Context, req *apiservice.UnbindServiceEnvironmentRequest) *apimodel.Response {
	if req == nil {
		return api.NewResponse(apimodel.Code_InvalidParameter)
	}
	return svr.authorizeLogicalServiceMutation(ctx, authtypes.UnbindServiceEnvironments,
		func(ctx context.Context) *apimodel.Response {
			return svr.nextSvr.UnbindServiceEnvironment(ctx, req)
		})
}

func (svr *Server) authorizeLogicalServiceRead(
	ctx context.Context) (context.Context, *apimodel.BatchQueryResponse) {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeLogicalServices)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return ctx, api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	ctx = cacheapi.AppendServicePredicate(ctx, func(_ context.Context, service *svctypes.Service) bool {
		if svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type: apisecurity.ResourceType_Services, ID: service.ID, Metadata: service.Meta,
		}) {
			return true
		}
		namespace := svr.Cache().Namespace().GetNamespace(service.Namespace)
		return namespace != nil && svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type: apisecurity.ResourceType_Namespaces, ID: namespace.Name, Metadata: namespace.Metadata,
		})
	})
	return ctx, nil
}

func (svr *Server) authorizeEnvironmentMutation(
	ctx context.Context, serviceID string, method authtypes.ServerFunctionName,
	next func(context.Context) *apimodel.Response) *apimodel.Response {
	service := svr.Cache().Service().GetServiceByID(serviceID)
	if service == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{{
		Name: service.Name, Namespace: service.Namespace,
	}}, authtypes.Modify, method)
	resources := authCtx.GetAccessResources()
	delete(resources, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(resources)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return next(ctx)
}

func (svr *Server) authorizeLogicalServiceMutation(
	ctx context.Context, method authtypes.ServerFunctionName,
	next func(context.Context) *apimodel.Response) *apimodel.Response {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Modify, method)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return next(ctx)
}
