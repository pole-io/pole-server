package config

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
)

type configTemplateValueReleaseCache struct {
	*cachebase.BaseCache
	storage   store.Store
	single    *singleflight.Group
	publisher func(*eventhub.ConfigTemplateSnapshotChangedEvent) error
}

func NewConfigTemplateValueReleaseCache(storage store.Store,
	cacheMgr cacheapi.CacheManager) cacheapi.Cache {
	return &configTemplateValueReleaseCache{
		BaseCache: cachebase.NewBaseCache(storage, cacheMgr),
		storage:   storage,
		publisher: func(event *eventhub.ConfigTemplateSnapshotChangedEvent) error {
			return eventhub.Publish(eventhub.ConfigFilePublishTopic, event)
		},
	}
}

func (c *configTemplateValueReleaseCache) Initialize(map[string]interface{}) error {
	c.single = &singleflight.Group{}
	return nil
}

func (c *configTemplateValueReleaseCache) Update() error {
	_, err, _ := c.single.Do(c.Name(), func() (interface{}, error) {
		return nil, c.DoCacheUpdate(c.Name(), c.realUpdate)
	})
	return err
}

func (c *configTemplateValueReleaseCache) realUpdate() (map[string]time.Time, int64, error) {
	releases, err := c.storage.GetMoreNamespaceTemplateValueReleases(c.IsFirstUpdate(), c.LastFetchTime())
	if err != nil {
		return nil, 0, err
	}
	if len(releases) == 0 {
		return nil, 0, nil
	}
	lastMtime := c.BaseCache.LastMtime(c.Name())
	affected := make(map[templateValueReleaseKey]struct{}, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}
		affected[templateValueReleaseKey{
			namespace: release.Namespace, templateID: release.TemplateID,
		}] = struct{}{}
		if release.ModifyTime.After(lastMtime) {
			lastMtime = release.ModifyTime
		}
	}
	for key := range affected {
		if err := c.publisher(&eventhub.ConfigTemplateSnapshotChangedEvent{
			Namespace: key.namespace, TemplateID: key.templateID,
		}); err != nil {
			log.Error("[Cache][ConfigTemplateValueRelease] publish snapshot change event",
				zap.String("namespace", key.namespace), zap.Uint64("template-id", key.templateID),
				zap.Error(err))
		}
	}
	return map[string]time.Time{c.Name(): lastMtime}, int64(len(releases)), nil
}

func (c *configTemplateValueReleaseCache) Clear() error {
	c.BaseCache.Clear()
	c.single = &singleflight.Group{}
	return nil
}

func (c *configTemplateValueReleaseCache) Name() string {
	return cacheapi.ConfigTemplateValueReleaseCacheName
}

type templateValueReleaseKey struct {
	namespace  string
	templateID uint64
}

var _ cacheapi.Cache = (*configTemplateValueReleaseCache)(nil)
