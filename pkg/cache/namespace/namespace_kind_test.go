package namespace

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestQueryFiltersNamespaceKindBeforePagination(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	cache := &namespaceCache{
		BaseCache: cachebase.NewBaseCache(storage, nil),
		storage:   storage,
		ids:       container.NewSyncMap[string, *types.Namespace](),
		updater:   new(singleflight.Group),
	}
	now := time.Now()
	storage.EXPECT().GetUnixSecond(time.Duration(0)).Return(now.Unix(), nil)
	storage.EXPECT().GetMoreNamespaces(gomock.Any()).Return([]*types.Namespace{
		{Name: "pole-system", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM, Valid: true, ModifyTime: now},
		{Name: "default", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS, Valid: true, ModifyTime: now},
		{Name: "production", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS, Valid: true, ModifyTime: now},
	}, nil)

	total, namespaces, err := cache.Query(context.Background(), &cacheapi.NamespaceArgs{
		Filter: map[string][]string{"kind": {"business"}},
		Offset: 0,
		Limit:  1,
	})

	require.NoError(t, err)
	require.Equal(t, uint32(2), total)
	require.Len(t, namespaces, 1)
	require.Equal(t, apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS, namespaces[0].Kind)
}
