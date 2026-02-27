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
	"sort"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

type Client2Api func(client *types.Client) *apiservice.Client

var (
	clientFilterAttributes = map[string]struct{}{
		"type":    {},
		"host":    {},
		"limit":   {},
		"offset":  {},
		"version": {},
	}
)

func (s *Server) checkAndStoreClient(ctx context.Context, req *apiservice.Client) *apimodel.Response {
	clientId := req.GetId()
	var needStore bool
	client := s.caches.Client().GetClient(clientId)
	var resp *apimodel.Response
	if nil == client {
		needStore = true
	} else {
		needStore = !ClientEquals(client.Proto(), req)
	}
	if needStore {
		resp = s.createClient(ctx, req)
	}

	if resp != nil {
		if resp.GetCode() != api.ExistedResource {
			return resp
		}
	}

	resp = s.HealthServer().ReportByClient(context.Background(), req)
	respCode := apimodel.Code(resp.GetCode())
	if respCode == apimodel.Code_HealthCheckNotOpen || respCode == apimodel.Code_HeartbeatTypeNotFound {
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}
	return resp
}

func (s *Server) createClient(ctx context.Context, req *apiservice.Client) *apimodel.Response {
	if namingServer.bc == nil || !namingServer.bc.ClientRegisterOpen() {
		return nil
	}
	return s.asyncCreateClient(ctx, req) // 批量异步
}

// 异步新建客户端
// 底层函数会合并create请求，增加并发创建的吞吐
// req 原始请求
// ins 包含了req数据与instanceID，serviceToken
func (s *Server) asyncCreateClient(ctx context.Context, req *apiservice.Client) *apimodel.Response {
	future := s.bc.AsyncRegisterClient(req)
	rsp, err := future.Done()
	if err != nil {
		rCode := rsp.(apimodel.Code)
		log.Error("[Server][ReportClient] async create client", zap.Error(err), utils.RequestID(ctx))
		if rCode == apimodel.Code_ExistedResource {
			req.Id = req.GetId()
		}
		return api.NewClientResponse(apimodel.Code(rCode), req)
	}

	return nil
}

// GetReportClients create one instance
func (s *Server) GetReportClients(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
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

	var total uint32

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
	resp.Amount = total
	resp.Size = uint32(len(services))
	resp.Data = enhancedClients2Api(services, client2Api)
	return resp
}

// client 数组转为[]*api.Client
func enhancedClients2Api(clients []*types.Client, handler Client2Api) []*anypb.Any {
	out := make([]*anypb.Any, 0, len(clients))
	for _, entry := range clients {
		outUser := handler(entry)
		out = append(out, protobuf.MarshalAny(outUser))
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
	if client1.GetId() != client2.GetId() {
		return false
	}
	if client1.GetHost() != client2.GetHost() {
		return false
	}
	if client1.GetVersion() != client2.GetVersion() {
		return false
	}
	if client1.GetType() != client2.GetType() {
		return false
	}
	if client1.GetLocation().GetRegion() != client2.GetLocation().GetRegion() {
		return false
	}
	if client1.GetLocation().GetZone() != client2.GetLocation().GetZone() {
		return false
	}
	if client1.GetLocation().GetCampus() != client2.GetLocation().GetCampus() {
		return false
	}
	if len(client1.Stat) != len(client2.Stat) {
		return false
	}

	sortStat := func(stat []*apiservice.StatInfo) {
		sort.Slice(stat, func(i, j int) bool {
			if client1.Stat[i].GetTarget() != client1.Stat[j].GetTarget() {
				return client1.Stat[i].GetTarget() < client1.Stat[j].GetTarget()
			}
			if client1.Stat[i].GetPort() != client1.Stat[j].GetPort() {
				return client1.Stat[i].GetPort() < client1.Stat[j].GetPort()
			}
			if client1.Stat[i].GetPath() != client1.Stat[j].GetPath() {
				return client1.Stat[i].GetPath() < client1.Stat[j].GetPath()
			}
			return client1.Stat[i].GetProtocol() < client1.Stat[j].GetProtocol()
		})
	}

	// 针对 client1 和 client2 的 stat 进行排序
	sortStat(client1.Stat)
	sortStat(client2.Stat)

	for i := 0; i < len(client1.Stat); i++ {
		if client1.Stat[i].GetTarget() != client2.Stat[i].GetTarget() {
			return false
		}
		if client1.Stat[i].GetPort() != client2.Stat[i].GetPort() {
			return false
		}
		if client1.Stat[i].GetPath() != client2.Stat[i].GetPath() {
			return false
		}
		if client1.Stat[i].GetProtocol() != client2.Stat[i].GetProtocol() {
			return false
		}
	}
	return true
}
