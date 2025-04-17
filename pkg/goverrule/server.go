package goverrule

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/cmdb"
	"github.com/pole-io/pole-server/apis/observability/history"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/namespace"
)

// Config 核心逻辑层配置
type Config struct {
	AutoCreate   *bool                  `yaml:"autoCreate"`
	Batch        map[string]interface{} `yaml:"batch"`
	Interceptors []string               `yaml:"-"`
	// Caches 缓存配置
	Caches map[string]map[string]interface{} `yaml:"caches"`
}

type Server struct {
	// config 配置
	config Config
	// store 数据存储
	storage store.Store
	// namespaceSvr 命名空间相关的操作
	namespaceSvr namespace.NamespaceOperateServer
	// caches 缓存
	caches cacheapi.CacheManager
	// cmdb CMDB 插件
	cmdb cmdb.CMDB
	// subCtxs eventhub 的 subscriber 的控制
	subCtxs []*eventhub.SubscribtionContext
	// emptyPushProtectSvs 开启了推空保护的服务数据
	emptyPushProtectSvs *container.SyncMap[string, *time.Timer]
}

// StopbetaCircuitBreakerRules implements GoverRuleServer.
func (s *Server) StopbetaCircuitBreakerRules(ctx context.Context, request []*fault_tolerance.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaFaultDetectRules implements GoverRuleServer.
func (s *Server) StopbetaFaultDetectRules(ctx context.Context, request []*fault_tolerance.FaultDetectRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaLaneGroups implements GoverRuleServer.
func (s *Server) StopbetaLaneGroups(ctx context.Context, req []*traffic_manage.LaneGroup) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaRateLimits implements GoverRuleServer.
func (s *Server) StopbetaRateLimits(ctx context.Context, request []*traffic_manage.Rule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaRouterRules implements GoverRuleServer.
func (s *Server) StopbetaRouterRules(ctx context.Context, req []*traffic_manage.RouteRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishCircuitBreakerRules implements GoverRuleServer.
func (s *Server) PublishCircuitBreakerRules(ctx context.Context, request []*fault_tolerance.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishFaultDetectRules implements GoverRuleServer.
func (s *Server) PublishFaultDetectRules(ctx context.Context, request []*fault_tolerance.FaultDetectRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishLaneGroups implements GoverRuleServer.
func (s *Server) PublishLaneGroups(ctx context.Context, req []*traffic_manage.LaneGroup) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishRateLimits implements GoverRuleServer.
func (s *Server) PublishRateLimits(ctx context.Context, request []*traffic_manage.Rule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishRouterRules implements GoverRuleServer.
func (s *Server) PublishRouterRules(ctx context.Context, req []*traffic_manage.RouteRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackCircuitBreakerRules implements GoverRuleServer.
func (s *Server) RollbackCircuitBreakerRules(ctx context.Context, request []*fault_tolerance.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackFaultDetectRules implements GoverRuleServer.
func (s *Server) RollbackFaultDetectRules(ctx context.Context, request []*fault_tolerance.FaultDetectRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackLaneGroups implements GoverRuleServer.
func (s *Server) RollbackLaneGroups(ctx context.Context, req []*traffic_manage.LaneGroup) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackRateLimits implements GoverRuleServer.
func (s *Server) RollbackRateLimits(ctx context.Context, request []*traffic_manage.Rule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackRouterRules implements GoverRuleServer.
func (s *Server) RollbackRouterRules(ctx context.Context, req []*traffic_manage.RouteRule) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

func (s *Server) Store() store.Store {
	return s.storage
}

// Cache 返回Cache
func (s *Server) Cache() cacheapi.CacheManager {
	return s.caches
}

// Namespace 返回NamespaceOperateServer
func (s *Server) Namespace() namespace.NamespaceOperateServer {
	return s.namespaceSvr
}

// RecordHistory server对外提供history插件的简单封装
func (s *Server) RecordHistory(ctx context.Context, entry *types.RecordEntry) {
	// 如果数据为空，则不需要打印了
	if entry == nil {
		return
	}

	fromClient, _ := ctx.Value(types.ContextIsFromClient).(bool)
	if fromClient {
		return
	}
	// 调用插件记录history
	history.GetHistory().Record(entry)
}

// storeError2AnyResponse store code
func storeError2AnyResponse(err error, msg proto.Message) *apiservice.Response {
	if err == nil {
		return nil
	}
	if nil == msg {
		return api.NewResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}
	resp := api.NewAnyDataResponse(storeapi.StoreCode2APICode(err), msg)
	resp.Info = &wrappers.StringValue{Value: err.Error()}
	return resp
}
