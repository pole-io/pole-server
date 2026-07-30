package config

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestTemplateValueReleaseCachePublishesOneAffectedScopePerIncrement(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := storemock.NewMockStore(ctrl)
	modifyTime := time.Unix(100, 0)
	storage.EXPECT().GetMoreNamespaceTemplateValueReleases(true, time.Unix(1, 0)).Return(
		[]*conftypes.NamespaceTemplateValueRelease{
			{ID: "normal-old", Namespace: "prod", TemplateID: 7, Active: false, ModifyTime: modifyTime},
			{ID: "normal-new", Namespace: "prod", TemplateID: 7, Active: true, ModifyTime: modifyTime},
			{ID: "gray-new", Namespace: "prod", TemplateID: 8, Active: true, ModifyTime: modifyTime},
		}, nil)
	var events []*eventhub.ConfigTemplateSnapshotChangedEvent
	cache := &configTemplateValueReleaseCache{
		BaseCache: cachebase.NewBaseCache(storage, nil),
		storage:   storage,
		publisher: func(event *eventhub.ConfigTemplateSnapshotChangedEvent) error {
			events = append(events, event)
			return nil
		},
	}
	require.NoError(t, cache.Initialize(nil))

	lastMtime, count, err := cache.realUpdate()

	require.NoError(t, err)
	require.Equal(t, int64(3), count)
	require.Equal(t, modifyTime, lastMtime[cache.Name()])
	require.ElementsMatch(t, []*eventhub.ConfigTemplateSnapshotChangedEvent{
		{Namespace: "prod", TemplateID: 7},
		{Namespace: "prod", TemplateID: 8},
	}, events)
}
