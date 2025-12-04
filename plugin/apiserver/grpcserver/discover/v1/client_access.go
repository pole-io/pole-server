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

package v1

import (
	"context"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/observability/statis"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"

	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

var (
	accesslog = commonlog.GetScopeOrDefaultByName(commonlog.APIServerLoggerName)
)

// ReportClient 客户端上报
func (g *DiscoverGRPCServer) ReportClient(ctx context.Context, in *apiservice.Client) (*apimodel.Response, error) {
	return g.namingServer.ReportClient(utils.ConvertGRPCContext(ctx), in), nil
}

// RegisterInstance 注册服务实例
func (g *DiscoverGRPCServer) RegisterInstance(ctx context.Context, in *apiservice.Instance) (*apimodel.Response, error) {
	// 需要记录操作来源，提高效率，只针对特殊接口添加operator
	rCtx := utils.ConvertGRPCContext(ctx)
	rCtx = context.WithValue(rCtx, types.StringContext("operator"), ParseGrpcOperator(ctx))

	grpcHeader := rCtx.Value(types.ContextGrpcHeader).(metadata.MD)

	if _, ok := grpcHeader["async-regis"]; ok {
		rCtx = context.WithValue(rCtx, types.ContextOpenAsyncRegis, true)
	}

	out := g.namingServer.RegisterInstance(rCtx, in)
	return out, nil
}

// DeregisterInstance 反注册服务实例
func (g *DiscoverGRPCServer) DeregisterInstance(
	ctx context.Context, in *apiservice.Instance) (*apimodel.Response, error) {
	// 需要记录操作来源，提高效率，只针对特殊接口添加operator
	rCtx := utils.ConvertGRPCContext(ctx)
	rCtx = context.WithValue(rCtx, types.StringContext("operator"), ParseGrpcOperator(ctx))

	out := g.namingServer.DeregisterInstance(rCtx, in)
	return out, nil
}

// Discover 统一发现接口
func (g *DiscoverGRPCServer) Discover(server apiservice.DiscoverGRPC_DiscoverServer) error {
	ctx := utils.ConvertGRPCContext(server.Context())
	clientIP, _ := ctx.Value(types.StringContext("client-ip")).(string)
	clientAddress, _ := ctx.Value(types.StringContext("client-address")).(string)
	requestID, _ := ctx.Value(types.ContextRequestId).(string)
	userAgent, _ := ctx.Value(types.StringContext("user-agent")).(string)
	method, _ := grpc.MethodFromServerStream(server)

	for {
		in, err := server.Recv()
		if err != nil {
			if io.EOF == err {
				return nil
			}
			return err
		}

		msg := fmt.Sprintf("receive grpc discover request: %s", in.Service.String())
		accesslog.Info(msg,
			zap.String("type", apiservice.DiscoverRequest_DiscoverRequestType_name[int32(in.Type)]),
			zap.String("client-address", clientAddress),
			zap.String("user-agent", userAgent),
			utils.ZapRequestID(requestID),
		)

		// 是否允许访问
		if ok := g.allowAccess(method); !ok {
			resp := api.NewDiscoverResponse(apimodel.Code_ClientAPINotOpen)
			if sendErr := server.Send(resp); sendErr != nil {
				return sendErr
			}
			continue
		}

		// stream模式，需要对每个包进行检测
		if code := g.enterRateLimit(clientIP, method); code != uint32(apimodel.Code_ExecuteSuccess) {
			resp := api.NewDiscoverResponse(apimodel.Code(code))
			if err = server.Send(resp); err != nil {
				return err
			}
			continue
		}

		if err = server.Send(g.handleDiscoverRequest(ctx, in)); err != nil {
			return err
		}
	}
}

func (g *DiscoverGRPCServer) handleDiscoverRequest(ctx context.Context, in *apiservice.DiscoverRequest) *apiservice.DiscoverResponse {
	var out *apiservice.DiscoverResponse
	var action string
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    action,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: in.GetService().GetNamespace(),
			Resource:  in.GetType().String() + ":" + in.GetService().GetName(),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  out.GetService().GetRevision(),
			Success:   out.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	// 兼容。如果请求中带了token，优先使用该token
	if in.GetService().GetToken() != "" {
		ctx = context.WithValue(ctx, types.ContextAuthTokenKey, in.GetService().GetToken())
	}

	switch in.Type {
	case apiservice.DiscoverRequest_INSTANCE:
		action = metrics.ActionDiscoverInstance
		out = g.namingServer.ServiceInstancesCache(ctx, &apiservice.DiscoverFilter{}, in.Service)
	case apiservice.DiscoverRequest_SERVICE_CONTRACTS:
		action = metrics.ActionDiscoverRouterRule
		out = g.ruleServer.GetOldRouterRuleWithCache(ctx, in.Service)
	case apiservice.DiscoverRequest_CUSTOM_ROUTE_RULE:
		action = metrics.ActionDiscoverRouterRule
		out = g.ruleServer.GetRouterRuleWithCache(ctx, in.Service)
	case apiservice.DiscoverRequest_RATE_LIMIT:
		action = metrics.ActionDiscoverRateLimit
		out = g.ruleServer.GetRateLimitWithCache(ctx, in.Service)
	case apiservice.DiscoverRequest_CIRCUIT_BREAKER:
		action = metrics.ActionDiscoverCircuitBreaker
		out = g.ruleServer.GetCircuitBreakerWithCache(ctx, in.Service)
	case apiservice.DiscoverRequest_SERVICES:
		action = metrics.ActionDiscoverServices
		out = g.namingServer.GetServiceWithCache(ctx, in.Service)
	case apiservice.DiscoverRequest_FAULT_DETECTOR:
		action = metrics.ActionDiscoverFaultDetect
		out = g.ruleServer.GetFaultDetectWithCache(ctx, in.Service)
	case apiservice.DiscoverRequest_LOSSLESS:
		action = metrics.ActionDiscoverLosslessRule
		out = g.ruleServer.GetLosslessRuleWithCache(ctx, in.Service)
	default:
		out = api.NewDiscoverRoutingResponse(apimodel.Code_InvalidDiscoverResource, in.Service)
	}

	return out
}

func (g *DiscoverGRPCServer) ReportServiceContract(ctx context.Context, in *apiservice.ServiceContract) (*apimodel.Response, error) {
	// 需要记录操作来源，提高效率，只针对特殊接口添加operator
	rCtx := utils.ConvertGRPCContext(ctx)
	rCtx = context.WithValue(rCtx, types.StringContext("operator"), ParseGrpcOperator(ctx))

	out := g.namingServer.ReportServiceContract(rCtx, in)
	return out, nil
}

// 查询服务契约
func (g *DiscoverGRPCServer) GetServiceContract(ctx context.Context, req *apiservice.ServiceContract) (*apimodel.Response, error) {
	// 需要记录操作来源，提高效率，只针对特殊接口添加operator
	rCtx := utils.ConvertGRPCContext(ctx)
	rCtx = context.WithValue(rCtx, types.StringContext("operator"), ParseGrpcOperator(ctx))

	out := g.namingServer.GetServiceContractWithCache(rCtx, req)
	return out, nil
}

// ParseGrpcOperator 构造请求源
func ParseGrpcOperator(ctx context.Context) string {
	// 获取请求源
	operator := "GRPC"
	if pr, ok := peer.FromContext(ctx); ok && pr.Addr != nil {
		addrSlice := strings.Split(pr.Addr.String(), ":")
		if len(addrSlice) == 2 {
			operator += ":" + addrSlice[0]
		}
	}
	return operator
}

// GetConfigFile 拉取配置
func (g *ConfigGRPCServer) GetConfigFile(ctx context.Context,
	req *apiconfig.ConfigFile) (*apimodel.Response, error) {
	ctx = utils.ConvertGRPCContext(ctx)

	startTime := commontime.CurrentMillisecond()
	var ret *apiconfig.ConfigDiscoverResponse
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    metrics.ActionGetConfigFile,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: req.GetNamespace(),
			Resource:  metrics.ResourceOfConfigFile(req.GetGroup(), req.GetName()),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  ret.GetRevision(),
			Success:   ret.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()
	ret = g.configServer.GetConfigFileWithCache(ctx, req)
	return &apimodel.Response{
		Code: ret.GetCode(),
		Info: ret.GetInfo(),
	}, nil
}

// CreateConfigFile 创建或更新配置
func (g *ConfigGRPCServer) CreateConfigFile(ctx context.Context,
	configFile *apiconfig.ConfigFile) (*apimodel.Response, error) {
	ctx = utils.ConvertGRPCContext(ctx)
	response := g.configServer.CreateConfigFileFromClient(ctx, configFile)
	return &apimodel.Response{
		Code: response.GetCode(),
		Info: response.GetInfo(),
	}, nil
}

// UpdateConfigFile 创建或更新配置
func (g *ConfigGRPCServer) UpdateConfigFile(ctx context.Context,
	configFile *apiconfig.ConfigFile) (*apimodel.Response, error) {
	ctx = utils.ConvertGRPCContext(ctx)
	response := g.configServer.UpdateConfigFileFromClient(ctx, configFile)
	return &apimodel.Response{
		Code: response.GetCode(),
		Info: response.GetInfo(),
	}, nil
}

// PublishConfigFile 发布配置
func (g *ConfigGRPCServer) PublishConfigFile(ctx context.Context,
	configFile *apiconfig.ConfigFileRelease) (*apimodel.Response, error) {
	ctx = utils.ConvertGRPCContext(ctx)
	response := g.configServer.PublishConfigFileFromClient(ctx, configFile)
	return &apimodel.Response{
		Code: response.GetCode(),
		Info: response.GetInfo(),
	}, nil
}

// UpsertAndPublishConfigFile 创建/更新并发布配置文件
func (g *ConfigGRPCServer) UpsertAndPublishConfigFile(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) (*apimodel.Response, error) {
	ctx = utils.ConvertGRPCContext(ctx)
	response := g.configServer.UpsertAndReleaseConfigFileFromClient(ctx, req)
	return &apimodel.Response{
		Code: response.GetCode(),
		Info: response.GetInfo(),
	}, nil
}

func (g *ConfigGRPCServer) GetConfigFileMetadataList(ctx context.Context,
	req *apiconfig.ConfigFileGroupRequest) (*apimodel.Response, error) {

	startTime := commontime.CurrentMillisecond()
	var ret *apiconfig.ConfigDiscoverResponse
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    metrics.ActionListConfigFiles,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: req.GetConfigFileGroup().GetNamespace(),
			Resource:  metrics.ResourceOfConfigFileList(req.GetConfigFileGroup().GetName()),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  ret.GetRevision(),
			Success:   ret.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	ctx = utils.ConvertGRPCContext(ctx)
	ret = g.configServer.GetConfigFileNamesWithCache(ctx, req)
	return &apimodel.Response{
		Code: ret.GetCode(),
		Info: ret.GetInfo(),
	}, nil
}

func (g *ConfigGRPCServer) Discover(svr apiconfig.ConfigGRPC_DiscoverServer) error {
	ctx := utils.ConvertGRPCContext(svr.Context())
	clientIP, _ := ctx.Value(types.StringContext("client-ip")).(string)
	clientAddress, _ := ctx.Value(types.StringContext("client-address")).(string)
	requestID, _ := ctx.Value(types.ContextRequestId).(string)
	userAgent, _ := ctx.Value(types.StringContext("user-agent")).(string)
	method, _ := grpc.MethodFromServerStream(svr)

	for {
		in, err := svr.Recv()
		if err != nil {
			if io.EOF == err {
				return nil
			}
			return err
		}

		msg := fmt.Sprintf("receive grpc discover request: %s", in.String())
		accesslog.Info(msg,
			zap.String("type", apiconfig.ConfigDiscoverRequest_ConfigDiscoverRequestType_name[int32(in.Type)]),
			zap.String("client-address", clientAddress),
			zap.String("user-agent", userAgent),
			utils.ZapRequestID(requestID),
		)

		// 是否允许访问
		if ok := g.allowAccess(method); !ok {
			resp := api.NewConfigDiscoverResponse(apimodel.Code_ClientAPINotOpen)
			if sendErr := svr.Send(resp); sendErr != nil {
				return sendErr
			}
			continue
		}

		// stream模式，需要对每个包进行检测
		if code := g.enterRateLimit(clientIP, method); code != uint32(apimodel.Code_ExecuteSuccess) {
			resp := api.NewConfigDiscoverResponse(apimodel.Code(code))
			if err = svr.Send(resp); err != nil {
				return err
			}
			continue
		}

		out := g.handleDiscoverRequest(ctx, in)
		if err := svr.Send(out); err != nil {
			return err
		}
	}
}

func (g *ConfigGRPCServer) handleDiscoverRequest(ctx context.Context, in *apiconfig.ConfigDiscoverRequest) *apiconfig.ConfigDiscoverResponse {
	var out *apiconfig.ConfigDiscoverResponse
	var action string
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    action,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: in.GetFile().GetNamespace(),
			Resource:  metrics.ResourceOfConfigFile(in.GetFile().GetGroup(), in.GetFile().GetName()),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  out.GetRevision(),
			Success:   out.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	switch in.Type {
	case apiconfig.ConfigDiscoverRequest_CONFIG_FILE:
		action = metrics.ActionGetConfigFile
		ret := g.configServer.GetConfigFileWithCache(ctx, in.GetFile())
		out = api.NewConfigDiscoverResponse(apimodel.Code(ret.GetCode()))
		out.File = ret.GetFile()
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
		out.Revision = ret.GetRevision()
	case apiconfig.ConfigDiscoverRequest_CONFIG_FILE_NAMES:
		action = metrics.ActionListConfigFiles
		ret := g.configServer.GetConfigFileNamesWithCache(ctx, &apiconfig.ConfigFileGroupRequest{
			Revision: in.GetRevision(),
			ConfigFileGroup: &apiconfig.ConfigFileGroup{
				Namespace: in.GetFile().GetNamespace(),
				Name:      in.GetFile().GetGroup(),
			},
		})
		out = api.NewConfigDiscoverResponse(apimodel.Code(ret.GetCode()))
		out.FileNames = ret.GetFileNames()
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES
		out.Revision = ret.GetRevision()
	case apiconfig.ConfigDiscoverRequest_CONFIG_FILE_GROUPS:
		action = metrics.ActionListConfigGroups
		req := in.GetFile()
		out = g.configServer.GetConfigGroupsWithCache(ctx, req)
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
	default:
		out = api.NewConfigDiscoverResponse(apimodel.Code_InvalidDiscoverResource)
	}

	return out
}
