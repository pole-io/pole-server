package goverrule

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
)

type ServerProxyFactory func(pre GoverRuleServer, s store.Store) (GoverRuleServer, error)

var (
	server     GoverRuleServer
	ruleServer *Server = new(Server)
	once               = sync.Once{}
	finishInit         = false
	// serverProxyFactories Service Server API 代理工厂
	serverProxyFactories = map[string]ServerProxyFactory{}
)

func RegisterServerProxy(name string, factor ServerProxyFactory) error {
	if _, ok := serverProxyFactories[name]; ok {
		return fmt.Errorf("duplicate ServerProxyFactory, name(%s)", name)
	}
	serverProxyFactories[name] = factor
	return nil
}

// Initialize 初始化
func Initialize(ctx context.Context, namingOpt *Config, opts ...InitOption) error {
	var err error
	once.Do(func() {
		ruleServer, server, err = InitServer(ctx, namingOpt, opts...)
	})

	if err != nil {
		return err
	}

	finishInit = true
	return nil
}

// GetServer 获取已经初始化好的Server
func GetServer() (GoverRuleServer, error) {
	if !finishInit {
		return nil, errors.New("server has not done InitializeServer")
	}

	return server, nil
}

// GetOriginServer 获取已经初始化好的Server
func GetOriginServer() (*Server, error) {
	if !finishInit {
		return nil, errors.New("server has not done InitializeServer")
	}

	return ruleServer, nil
}

// 内部初始化函数
func InitServer(ctx context.Context, opt *Config, opts ...InitOption) (*Server, GoverRuleServer, error) {
	actualSvr := new(Server)
	// l5service
	actualSvr.config = *opt
	actualSvr.subCtxs = make([]*eventhub.SubscribtionContext, 0, 4)

	for i := range opts {
		opts[i](actualSvr)
	}

	var proxySvr GoverRuleServer
	proxySvr = actualSvr
	// 需要返回包装代理的 DiscoverServer
	order := opt.Interceptors
	for i := range order {
		factory, exist := serverProxyFactories[order[i]]
		if !exist {
			return nil, nil, fmt.Errorf("name(%s) not exist in serverProxyFactories", order[i])
		}

		afterSvr, err := factory(proxySvr, actualSvr.storage)
		if err != nil {
			return nil, nil, err
		}
		proxySvr = afterSvr
	}
	return actualSvr, proxySvr, nil
}

func GetChainOrder() []string {
	return []string{
		"auth",
		"paramcheck",
	}
}
