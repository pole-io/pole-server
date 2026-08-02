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

package grpc

import (
	"context"
	"io"

	"github.com/modern-go/reflect2"

	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"

	limiterapi "github.com/pole-io/pole-server/pkg/limiter/internal/api/v2"
	"github.com/pole-io/pole-server/pkg/limiter/internal/ratelimitv2"
	"github.com/pole-io/pole-server/pkg/limiter/internal/statistics"
	"github.com/pole-io/pole-server/pkg/limiter/internal/utils"
)

// RateLimitService 限流服务。
type RateLimitService struct {
	coreServer *ratelimitv2.Server
	statics    statistics.Statis
}

func (s *RateLimitService) postService(
	streamCtx *ratelimitv2.StreamContext, ipAddr *utils.IPAddress, wrapper *clientWrapper) {
	client := wrapper.client
	if reflect2.IsNil(client) {
		return
	}
	s.coreServer.CleanupClient(client, streamCtx.ContextId())
	s.statics.AddEventToLog(ratelimitv2.NewStreamUpdateEvent(streamCtx.ContextId(), ipAddr, ratelimitv2.ActionDelete))
}

type clientWrapper struct {
	client ratelimitv2.Client
}

// Service 处理初始化和租约生命周期命令。
func (s *RateLimitService) Service(stream apiv2.RateLimitGRPC_ServiceServer) error {
	ctx := parseContext(stream.Context())
	clientIP := utils.ParseStructClientIP(ctx)
	streamCtx := ratelimitv2.NewStreamContext(stream)
	s.statics.AddEventToLog(ratelimitv2.NewStreamUpdateEvent(streamCtx.ContextId(), clientIP, ratelimitv2.ActionAdd))
	wrapper := &clientWrapper{}
	defer s.postService(streamCtx, clientIP, wrapper)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		startTimeMicro := utils.CurrentMicrosecond()
		var counter ratelimitv2.CounterV2
		response := &apiv2.RateLimitResponse{Cmd: req.GetCmd()}
		switch req.GetCmd() {
		case apiv2.RateLimitCmd_INIT:
			response.RateLimitInitResponse, wrapper.client = s.coreServer.InitializeClient(
				req.GetRateLimitInitRequest(), wrapper.client, clientIP, streamCtx)
			if response.RateLimitInitResponse == nil {
				response.RateLimitInitResponse, counter = s.coreServer.InitializeQuota(
					ctx, wrapper.client, req.GetRateLimitInitRequest())
			}
		case apiv2.RateLimitCmd_RESERVE:
			response.QuotaReserveResponse = s.coreServer.ReserveQuota(wrapper.client, req.GetQuotaReserveRequest())
		case apiv2.RateLimitCmd_BATCH_INIT:
			response.RateLimitBatchInitResponse, wrapper.client = s.coreServer.InitializeClientBatch(
				req.GetRateLimitBatchInitRequest(), wrapper.client, clientIP, streamCtx)
			if response.RateLimitBatchInitResponse == nil {
				response.RateLimitBatchInitResponse, counter = s.coreServer.BatchInitializeQuota(
					ctx, wrapper.client, req.GetRateLimitBatchInitRequest())
			}
		case apiv2.RateLimitCmd_UPDATE:
			response.QuotaUpdateResponse = s.coreServer.UpdateQuota(wrapper.client, req.GetQuotaUpdateRequest())
		case apiv2.RateLimitCmd_SETTLE:
			response.QuotaSettleResponse = s.coreServer.SettleQuota(wrapper.client, req.GetQuotaSettleRequest())
		default:
			response.QuotaSettleResponse = &apiv2.QuotaSettleResponse{
				Code:      uint32(limiterapi.InvalidConsumption),
				Timestamp: utils.CurrentMillisecond(),
			}
		}

		err = stream.Send(response)
		endTimeMicro := utils.CurrentMicrosecond()
		if wrapper.client != nil && counter != nil {
			counter.UpdateClientSendTime(wrapper.client, endTimeMicro)
		}
		apiCallStatValue := statistics.PoolGetAPICallStatValueImpl()
		apiCallStatValue.StatKey.APIKey = limiterapi.GetAPIKey(response)
		apiCallStatValue.StatKey.Code = limiterapi.GetErrorCode(response)
		apiCallStatValue.StatKey.MsgType = statistics.MsgSync
		if !reflect2.IsNil(counter) {
			apiCallStatValue.StatKey.Duration = counter.Identifier().Duration
		}
		apiCallStatValue.Latency = endTimeMicro - startTimeMicro
		apiCallStatValue.ReqCount = 1
		s.statics.AddAPICall(apiCallStatValue)
		statistics.PoolPutAPICallStatValueImpl(apiCallStatValue)
		if err != nil {
			return err
		}
	}
}

// TimeAdjust 获取服务端时间戳。
func (s *RateLimitService) TimeAdjust(
	ctx context.Context, req *apiv2.TimeAdjustRequest) (*apiv2.TimeAdjustResponse, error) {
	return &apiv2.TimeAdjustResponse{ServerTimestamp: utils.CurrentMillisecond()}, nil
}
