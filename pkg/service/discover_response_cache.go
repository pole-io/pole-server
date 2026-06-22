package service

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/protobuf/proto"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/observability/statis"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
)

const defaultDiscoverResponseCacheSize = 1024

type discoverResponseCache struct {
	cache *lru.Cache
}

func newDiscoverResponseCache(size int) *discoverResponseCache {
	if size <= 0 {
		size = defaultDiscoverResponseCacheSize
	}
	cache, err := lru.New(size)
	if err != nil {
		return nil
	}
	return &discoverResponseCache{cache: cache}
}

func discoverInstanceResponseCacheKey(namespace, service, revision string, onlyHealthy bool) string {
	return fmt.Sprintf("INSTANCE:%s:%s:%s:%t", namespace, service, revision, onlyHealthy)
}

func (c *discoverResponseCache) Get(key string) (*apiservice.DiscoverResponse, bool) {
	if c == nil || c.cache == nil || key == "" {
		return nil, false
	}
	val, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	resp, ok := val.(*apiservice.DiscoverResponse)
	if !ok || resp == nil {
		return nil, false
	}
	return proto.Clone(resp).(*apiservice.DiscoverResponse), true
}

func (c *discoverResponseCache) Put(key string, resp *apiservice.DiscoverResponse) {
	if c == nil || c.cache == nil || key == "" || resp == nil {
		return
	}
	c.cache.Add(key, proto.Clone(resp).(*apiservice.DiscoverResponse))
}

func reportDiscoverCacheCall(protocol string, hit bool) {
	statis.GetStatis().ReportCallMetrics(metrics.CallMetric{
		Type:     metrics.DiscoverCacheCallMetric,
		Protocol: protocol,
		Success:  hit,
		Times:    1,
	})
}
