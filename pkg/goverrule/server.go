package goverrule

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"

	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

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
