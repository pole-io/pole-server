package config

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	cachemock "github.com/pole-io/pole-server/pkg/cache/mock"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

func TestCheckQuickResponseClientUsesVisibleTemplateSnapshotRevision(t *testing.T) {
	watchCenter := newTemplateRevisionWatchCenter(func(
		labels map[string]string, file *apiconfig.ConfigFileRelease,
	) *apiconfig.ConfigDiscoverResponse {
		require.Equal(t, "canary", labels["environment"])
		require.Equal(t, "app.yaml", file.GetName())
		return templateWatchResponse("snapshot-new", 7)
	})
	watchContext := templateWatchContext("client", map[string]string{"environment": "canary"})
	watchContext.AppendInterest(&apiconfig.ConfigFileRelease{
		Id: "snapshot-old", Namespace: "prod", Group: "payments", Name: "app.yaml",
	})

	response := watchCenter.CheckQuickResponseClient(watchContext)

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	require.Equal(t, "snapshot-new", response.GetRevision())
	require.Equal(t, uint64(7), response.GetRenderSnapshot().GetTemplateBinding().GetTemplateId())
}

func TestTemplateSnapshotEventNotifiesOnlyClientsWhoseVisibleRevisionChanged(t *testing.T) {
	watchCenter := newTemplateRevisionWatchCenter(func(
		labels map[string]string, _ *apiconfig.ConfigFileRelease,
	) *apiconfig.ConfigDiscoverResponse {
		if labels["environment"] == "canary" {
			return templateWatchResponse("snapshot-new", 7)
		}
		return templateWatchResponse("snapshot-old", 7)
	})
	canary := templateWatchContext("canary", map[string]string{"environment": "canary"})
	stable := templateWatchContext("stable", map[string]string{"environment": "stable"})
	watchFile := &apiconfig.ConfigFileRelease{
		Id: "snapshot-old", Namespace: "prod", Group: "payments", Name: "app.yaml",
	}
	watchCenter.AddWatcher(canary.ClientID(), []*apiconfig.ConfigFileRelease{watchFile},
		func(string, BetaReleaseMatcher) WatchContext { return canary })
	watchCenter.AddWatcher(stable.ClientID(), []*apiconfig.ConfigFileRelease{watchFile},
		func(string, BetaReleaseMatcher) WatchContext { return stable })

	err := watchCenter.OnEvent(context.Background(), &eventhub.ConfigTemplateSnapshotChangedEvent{
		Namespace: "prod", TemplateID: 7,
	})

	require.NoError(t, err)
	response, err := canary.GetNotifieResultWithTime(time.Second)
	require.NoError(t, err)
	require.Equal(t, "snapshot-new", response.GetRevision())
	_, err = stable.GetNotifieResultWithTime(20 * time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestConfigReleaseEventReevaluatesSnapshotRevision(t *testing.T) {
	watchCenter := newTemplateRevisionWatchCenter(func(
		_ map[string]string, _ *apiconfig.ConfigFileRelease,
	) *apiconfig.ConfigDiscoverResponse {
		return templateWatchResponse("snapshot-new", 7)
	})
	watchContext := templateWatchContext("client", nil)
	watchCenter.AddWatcher(watchContext.ClientID(), []*apiconfig.ConfigFileRelease{{
		Id: "snapshot-old", Namespace: "prod", Group: "payments", Name: "app.yaml",
	}}, func(string, BetaReleaseMatcher) WatchContext { return watchContext })

	err := watchCenter.OnEvent(context.Background(), &eventhub.PublishConfigFileEvent{
		Message: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Namespace: "prod", Group: "payments", FileName: "app.yaml",
			},
			Version: 4,
		},
	})

	require.NoError(t, err)
	response, err := watchContext.GetNotifieResultWithTime(time.Second)
	require.NoError(t, err)
	require.Equal(t, "snapshot-new", response.GetRevision())
}

func TestNumericVersionWatchRemainsCompatible(t *testing.T) {
	ctrl := gomock.NewController(t)
	fileCache := cachemock.NewMockConfigFileCache(ctrl)
	release := configClientReleaseForTest("normal-release", conftypes.ReleaseTypeNormal, 3, "content")
	release.Namespace, release.Group, release.FileName = "prod", "payments", "app.yaml"
	fileCache.EXPECT().GetActiveRelease("prod", "payments", "app.yaml").Return(release)
	watchCenter := &watchCenter{fileCache: fileCache}
	watchContext := templateWatchContext("legacy", nil)
	watchContext.AppendInterest(&apiconfig.ConfigFileRelease{
		Id: "2", Namespace: "prod", Group: "payments", Name: "app.yaml",
	})

	response := watchCenter.CheckQuickResponseClient(watchContext)

	require.Equal(t, uint64(3), response.GetFile().GetVersion())
	require.Empty(t, response.GetRevision())
}

func newTemplateRevisionWatchCenter(resolver ConfigSnapshotResolver) *watchCenter {
	return &watchCenter{
		clients:          container.NewSyncMap[string, WatchContext](),
		watchers:         container.NewSyncMap[string, *container.SyncSet[string]](),
		snapshotResolver: resolver,
	}
}

func templateWatchContext(clientID string, labels map[string]string) *LongPollWatchContext {
	return &LongPollWatchContext{
		clientId:         clientID,
		labels:           labels,
		finishTime:       time.Now().Add(time.Minute),
		finishChan:       make(chan *apiconfig.ConfigDiscoverResponse, 1),
		watchConfigFiles: map[string]*apiconfig.ConfigFileRelease{},
		betaMatcher: func(map[string]string, *conftypes.SimpleConfigFileRelease) bool {
			return false
		},
	}
}

func templateWatchResponse(revision string, templateID uint64) *apiconfig.ConfigDiscoverResponse {
	return &apiconfig.ConfigDiscoverResponse{
		Code:     uint32(apimodel.Code_ExecuteSuccess),
		Info:     apimodel.Code_ExecuteSuccess.String(),
		Type:     apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Revision: revision,
		File: &apiconfig.ConfigFileRelease{
			Namespace: "prod", Group: "payments", FileName: "app.yaml", Version: 3,
		},
		RenderSnapshot: &apiconfig.RenderSnapshot{
			TemplateBinding: &apiconfig.ConfigTemplateBinding{TemplateId: templateID},
		},
	}
}
