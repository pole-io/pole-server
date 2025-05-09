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

package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/metrics"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
)

var (
	clientFilterAttributes = map[string]struct{}{
		"type":    {},
		"host":    {},
		"limit":   {},
		"offset":  {},
		"version": {},
	}
)

// GetPrometheusTargets Used for client acquisition service information
func (s *Server) GetPrometheusTargets(ctx context.Context,
	query map[string]string) *types.PrometheusDiscoveryResponse {
	if s.caches == nil {
		return &types.PrometheusDiscoveryResponse{
			Code:     api.NotFoundInstance,
			Response: make([]types.PrometheusTarget, 0),
		}
	}

	targets := make([]types.PrometheusTarget, 0, 8)
	expectSchema := map[string]struct{}{
		"http":  {},
		"https": {},
	}

	s.Cache().Client().IteratorClients(func(key string, value *types.Client) bool {
		for i := range value.Proto().Stat {
			stat := value.Proto().Stat[i]
			if stat.Target.GetValue() != types.StatReportPrometheus {
				continue
			}
			_, ok := expectSchema[strings.ToLower(stat.Protocol.GetValue())]
			if !ok {
				continue
			}

			target := types.PrometheusTarget{
				Targets: []string{fmt.Sprintf("%s:%d", value.Proto().Host.GetValue(), stat.Port.GetValue())},
				Labels: map[string]string{
					"__metrics_path__":         stat.Path.GetValue(),
					"__scheme__":               stat.Protocol.GetValue(),
					"__meta_polaris_client_id": value.Proto().Id.GetValue(),
				},
			}
			targets = append(targets, target)
		}

		return true
	})

	// 加入pole-server集群自身
	checkers := s.healthServer.ListCheckerServer()
	for i := range checkers {
		checker := checkers[i]
		target := types.PrometheusTarget{
			Targets: []string{fmt.Sprintf("%s:%d", checker.Host(), metrics.GetMetricsPort())},
			Labels: map[string]string{
				"__metrics_path__":         "/metrics",
				"__scheme__":               "http",
				"__meta_polaris_client_id": checker.ID(),
			},
		}
		targets = append(targets, target)
	}

	return &types.PrometheusDiscoveryResponse{
		Code:     api.ExecuteSuccess,
		Response: targets,
	}
}

func (s *Server) checkAndStoreClient(ctx context.Context, req *apiservice.Client) *apiservice.Response {
	clientId := req.GetId().GetValue()
	var needStore bool
	client := s.caches.Client().GetClient(clientId)
	var resp *apiservice.Response
	if nil == client {
		needStore = true
	} else {
		needStore = !ClientEquals(client.Proto(), req)
	}
	if needStore {
		client, resp = s.createClient(ctx, req)
	}

	if resp != nil {
		if resp.GetCode().GetValue() != api.ExistedResource {
			return resp
		}
	}

	resp = s.HealthServer().ReportByClient(context.Background(), req)
	respCode := apimodel.Code(resp.GetCode().GetValue())
	if respCode == apimodel.Code_HealthCheckNotOpen || respCode == apimodel.Code_HeartbeatTypeNotFound {
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}
	return resp
}

func (s *Server) createClient(ctx context.Context, req *apiservice.Client) (*types.Client, *apiservice.Response) {
	if namingServer.bc == nil || !namingServer.bc.ClientRegisterOpen() {
		return nil, nil
	}
	return s.asyncCreateClient(ctx, req) // 批量异步
}

// 异步新建客户端
// 底层函数会合并create请求，增加并发创建的吞吐
// req 原始请求
// ins 包含了req数据与instanceID，serviceToken
func (s *Server) asyncCreateClient(ctx context.Context, req *apiservice.Client) (*types.Client, *apiservice.Response) {
	future := s.bc.AsyncRegisterClient(req)
	if err := future.Wait(); err != nil {
		log.Error("[Server][ReportClient] async create client", zap.Error(err), utils.RequestID(ctx))
		if future.Code() == apimodel.Code_ExistedResource {
			req.Id = protobuf.NewStringValue(req.GetId().GetValue())
		}
		return nil, api.NewClientResponse(apimodel.Code(future.Code()), req)
	}

	return future.Client(), nil
}

// GetReportClients create one instance
func (s *Server) GetReportClients(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	searchFilters := make(map[string]string)
	var (
		offset, limit uint32
		err           error
	)

	for key, value := range query {
		if _, ok := clientFilterAttributes[key]; !ok {
			log.Errorf("[Server][Client] attribute(%s) it not allowed", key)
			return api.NewBatchQueryResponseWithMsg(apimodel.Code_InvalidParameter, key+" is not allowed")
		}
		searchFilters[key] = value
	}

	var (
		total   uint32
		clients []*types.Client
	)

	offset, limit, err = valid.ParseOffsetAndLimit(searchFilters)
	if err != nil {
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	total, services, err := s.caches.Client().GetClientsByFilter(searchFilters, offset, limit)
	if err != nil {
		log.Errorf("[Server][Client][Query] req(%+v) store err: %s", query, err.Error())
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = protobuf.NewUInt32Value(total)
	resp.Size = protobuf.NewUInt32Value(uint32(len(services)))
	resp.Clients = enhancedClients2Api(clients, client2Api)
	return resp
}

type Client2Api func(client *types.Client) *apiservice.Client

// client 数组转为[]*api.Client
func enhancedClients2Api(clients []*types.Client, handler Client2Api) []*apiservice.Client {
	out := make([]*apiservice.Client, 0, len(clients))
	for _, entry := range clients {
		outUser := handler(entry)
		out = append(out, outUser)
	}
	return out
}

// model.Client 转为 api.Client
func client2Api(client *types.Client) *apiservice.Client {
	if client == nil {
		return nil
	}
	out := client.Proto()
	return out
}

func ClientEquals(client1 *apiservice.Client, client2 *apiservice.Client) bool {
	if client1.GetId().GetValue() != client2.GetId().GetValue() {
		return false
	}
	if client1.GetHost().GetValue() != client2.GetHost().GetValue() {
		return false
	}
	if client1.GetVersion().GetValue() != client2.GetVersion().GetValue() {
		return false
	}
	if client1.GetType() != client2.GetType() {
		return false
	}
	if client1.GetLocation().GetRegion().GetValue() != client2.GetLocation().GetRegion().GetValue() {
		return false
	}
	if client1.GetLocation().GetZone().GetValue() != client2.GetLocation().GetZone().GetValue() {
		return false
	}
	if client1.GetLocation().GetCampus().GetValue() != client2.GetLocation().GetCampus().GetValue() {
		return false
	}
	if len(client1.Stat) != len(client2.Stat) {
		return false
	}

	sortStat := func(stat []*apiservice.StatInfo) {
		sort.Slice(stat, func(i, j int) bool {
			if client1.Stat[i].GetTarget().GetValue() != client1.Stat[j].GetTarget().GetValue() {
				return client1.Stat[i].GetTarget().GetValue() < client1.Stat[j].GetTarget().GetValue()
			}
			if client1.Stat[i].GetPort().GetValue() != client1.Stat[j].GetPort().GetValue() {
				return client1.Stat[i].GetPort().GetValue() < client1.Stat[j].GetPort().GetValue()
			}
			if client1.Stat[i].GetPath().GetValue() != client1.Stat[j].GetPath().GetValue() {
				return client1.Stat[i].GetPath().GetValue() < client1.Stat[j].GetPath().GetValue()
			}
			return client1.Stat[i].GetProtocol().GetValue() < client1.Stat[j].GetProtocol().GetValue()
		})
	}

	// 针对 client1 和 client2 的 stat 进行排序
	sortStat(client1.Stat)
	sortStat(client2.Stat)

	for i := 0; i < len(client1.Stat); i++ {
		if client1.Stat[i].GetTarget().GetValue() != client2.Stat[i].GetTarget().GetValue() {
			return false
		}
		if client1.Stat[i].GetPort().GetValue() != client2.Stat[i].GetPort().GetValue() {
			return false
		}
		if client1.Stat[i].GetPath().GetValue() != client2.Stat[i].GetPath().GetValue() {
			return false
		}
		if client1.Stat[i].GetProtocol().GetValue() != client2.Stat[i].GetProtocol().GetValue() {
			return false
		}
	}
	return true
}
