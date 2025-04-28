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

package healthcheck

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/service/healthcheck"
	"github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	commontime "github.com/pole-io/pole-server/pkg/common/time"
	"github.com/pole-io/pole-server/pkg/service/batch"
)

var (
	_server    = new(Server)
	once       = sync.Once{}
	finishInit = false
)

// Server health checks the main server
type Server struct {
	hcOpt          *Config
	storage        store.Store
	defaultChecker healthcheck.HealthChecker
	checkers       map[int32]healthcheck.HealthChecker
	cacheProvider  *CacheProvider
	timeAdjuster   *TimeAdjuster
	dispatcher     *Dispatcher
	checkScheduler *CheckScheduler
	localHost      string
	bc             *batch.Controller
	serviceCache   cacheapi.ServiceCache
	instanceCache  cacheapi.InstanceCache

	subCtxs []*eventhub.SubscribtionContext
}

func NewHealthServer(ctx context.Context, hcOpt *Config, options ...serverOption) (*Server, error) {
	if len(options) == 0 {
		options = make([]serverOption, 0, 4)
	}
	if hcOpt == nil {
		hcOpt = &Config{}
	}
	hcOpt.SetDefault()
	options = append(options,
		withChecker(),
		withCacheProvider(),
		withCheckScheduler(newCheckScheduler(ctx, hcOpt.SlotNum, hcOpt.MinCheckInterval,
			hcOpt.MaxCheckInterval, hcOpt.ClientCheckInterval, hcOpt.ClientCheckTtl)),
		withDispatcher(ctx),
		// 这个必须保证在最后一个 option
		withSubscriber(ctx),
	)

	svr := &Server{
		hcOpt:     hcOpt,
		localHost: hcOpt.LocalHost,
	}
	for i := range options {
		if err := options[i](svr); err != nil {
			return nil, err
		}
	}
	return svr, nil
}

// Initialize 初始化
func Initialize(ctx context.Context, hcOpt *Config, bc *batch.Controller) error {
	var err error
	once.Do(func() {
		err = initialize(ctx, hcOpt, bc)
	})

	if err != nil {
		return err
	}

	finishInit = true
	return nil
}

func initialize(ctx context.Context, hcOpt *Config, bc *batch.Controller) error {
	hcOpt.SetDefault()
	storage, err := store.GetStore()
	if err != nil {
		return err
	}

	svr, err := NewHealthServer(ctx, hcOpt,
		WithStore(storage),
		WithBatchController(bc),
		WithTimeAdjuster(newTimeAdjuster(ctx, storage)),
	)
	if err != nil {
		return err
	}

	_server = svr

	return svr.run(ctx)
}

func (s *Server) run(ctx context.Context) error {
	if !s.isOpen() {
		return nil
	}

	s.checkScheduler.run(ctx)
	s.timeAdjuster.doTimeAdjust(ctx)
	s.dispatcher.startDispatchingJob(ctx)
	return nil
}

// SelfService .
func (s *Server) SelfService() string {
	return s.cacheProvider.selfService
}

// Report heartbeat request
func (s *Server) Report(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	return s.doReport(ctx, req)
}

// Reports batch report heartbeat request
func (s *Server) Reports(ctx context.Context, req []*apiservice.InstanceHeartbeat) *apiservice.Response {
	return s.doReports(ctx, req)
}

// ReportByClient report heartbeat request by client
func (s *Server) ReportByClient(ctx context.Context, req *apiservice.Client) *apiservice.Response {
	return s.doReportByClient(ctx, req)
}

func (s *Server) Destroy() {
	for i := range s.subCtxs {
		s.subCtxs[i].Cancel()
	}
}

// GetServer 获取已经初始化好的Server
func GetServer() (*Server, error) {
	if !finishInit {
		return nil, errors.New("server has not done InitializeServer")
	}

	return _server, nil
}

// SetServer for test only
func SetServer(srv *Server) {
	_server = srv
}

// SetServiceCache 设置服务缓存
func (s *Server) SetServiceCache(serviceCache cacheapi.ServiceCache) {
	s.serviceCache = serviceCache
}

// SetInstanceCache 设置服务实例缓存
func (s *Server) SetInstanceCache(instanceCache cacheapi.InstanceCache) {
	s.instanceCache = instanceCache
}

// CacheProvider get cache provider
func (s *Server) CacheProvider() (*CacheProvider, error) {
	if !finishInit {
		return nil, errors.New("cache provider has not done InitializeServer")
	}
	return s.cacheProvider, nil
}

// ListCheckerServer get checker server instance list
func (s *Server) ListCheckerServer() []*svctypes.Instance {
	ret := make([]*svctypes.Instance, 0, s.cacheProvider.selfServiceInstances.Count())
	s.cacheProvider.selfServiceInstances.Range(func(instanceId string, value ItemWithChecker) {
		ret = append(ret, value.GetInstance())
	})
	return ret
}

// publishInstanceEvent 发布服务事件
func (s *Server) publishInstanceEvent(serviceID string, event *svctypes.InstanceEvent) {
	event.SvcId = serviceID
	if event.Instance != nil {
		// event.Instance = proto.Clone(event.Instance).(*apiservice.Instance)
	}
	_ = eventhub.Publish(eventhub.InstanceEventTopic, event)
}

// GetLastHeartbeat 获取上一次心跳的时间
func (s *Server) GetLastHeartbeat(req *apiservice.Instance) *apiservice.Response {
	if len(s.checkers) == 0 {
		return api.NewResponse(apimodel.Code_HealthCheckNotOpen)
	}
	id, errRsp := checkHeartbeatInstance(req)
	if errRsp != nil {
		return errRsp
	}
	req.Id = protobuf.NewStringValue(id)
	insCache := s.cacheProvider.GetInstance(id)
	if insCache == nil {
		return api.NewInstanceResponse(apimodel.Code_NotFoundResource, req)
	}
	checker, ok := s.checkers[int32(insCache.HealthCheck().GetType())]
	if !ok {
		return api.NewInstanceResponse(apimodel.Code_HeartbeatTypeNotFound, req)
	}
	queryResp, err := checker.Query(context.Background(), &healthcheck.QueryRequest{
		InstanceId: insCache.ID(),
		Host:       insCache.Host(),
		Port:       insCache.Port(),
	})
	if err != nil {
		return api.NewInstanceRespWithError(apimodel.Code_ExecuteException, err, req)
	}
	req.Service = insCache.Proto.GetService()
	req.Namespace = insCache.Proto.GetNamespace()
	req.Host = insCache.Proto.GetHost()
	req.Port = insCache.Proto.Port
	req.VpcId = insCache.Proto.GetVpcId()
	req.HealthCheck = insCache.Proto.GetHealthCheck()
	req.Metadata = make(map[string]string, 3)
	req.Metadata["last-heartbeat-timestamp"] = strconv.Itoa(int(queryResp.LastHeartbeatSec))
	req.Metadata["last-heartbeat-time"] = commontime.Time2String(time.Unix(queryResp.LastHeartbeatSec, 0))
	req.Metadata["system-time"] = commontime.Time2String(time.Unix(s.currentTimeSec(), 0))
	return api.NewInstanceResponse(apimodel.Code_ExecuteSuccess, req)
}

// Checkers get all health checker, for test only
func (s *Server) Checkers() map[int32]healthcheck.HealthChecker {
	return s.checkers
}

func (s *Server) isOpen() bool {
	return s.hcOpt.IsOpen()
}

func (s *Server) currentTimeSec() int64 {
	return time.Now().Unix() - s.timeAdjuster.GetDiff()
}
