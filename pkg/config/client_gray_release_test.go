package config

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	cachemock "github.com/pole-io/pole-server/pkg/cache/mock"
)

func TestGetConfigFileWithCacheReturnsNormalAndMatchedGrayContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const (
		namespace = "default"
		group     = "group-a"
		fileName  = "app.yaml"
	)

	normal := configClientReleaseForTest("normal-release", conftypes.ReleaseTypeNormal, 1, "normal-content")
	grayA := configClientReleaseForTest("gray-a", conftypes.ReleaseTypeGray, 2, "gray-a-content")
	grayB := configClientReleaseForTest("gray-b", conftypes.ReleaseTypeGray, 3, "gray-b-content")

	fileCache := cachemock.NewMockConfigFileCache(ctrl)
	grayCache := cachemock.NewMockGrayCache(ctrl)
	fileCache.EXPECT().GetActiveGrayReleases(namespace, group, fileName).
		Return([]*conftypes.ConfigFileRelease{grayA, grayB}).AnyTimes()
	fileCache.EXPECT().GetActiveRelease(namespace, group, fileName).Return(normal).AnyTimes()
	grayCache.EXPECT().HitGrayRule(GetGrayConfigReaseKey(grayA.SimpleConfigFileRelease), gomock.Any()).
		DoAndReturn(func(_ string, labels map[string]string) bool {
			return labels[types.ClientLabel_ID] == "client-a" || labels[types.ClientLabel_ID] == "client-both"
		}).AnyTimes()
	grayCache.EXPECT().HitGrayRule(GetGrayConfigReaseKey(grayB.SimpleConfigFileRelease), gomock.Any()).
		DoAndReturn(func(_ string, labels map[string]string) bool {
			return labels[types.ClientLabel_ID] == "client-b" || labels[types.ClientLabel_ID] == "client-both"
		}).AnyTimes()

	svr := &Server{
		fileCache: fileCache,
		grayCache: grayCache,
	}

	tests := []struct {
		name            string
		clientID        string
		wantReleaseName string
		wantVersion     uint64
		wantContent     string
		wantType        string
	}{
		{
			name:            "client without gray rule gets normal release",
			clientID:        "client-normal",
			wantReleaseName: "normal-release",
			wantVersion:     1,
			wantContent:     "normal-content",
			wantType:        conftypes.ReleaseTypeNormal,
		},
		{
			name:            "client matching first gray gets gray content",
			clientID:        "client-a",
			wantReleaseName: "gray-a",
			wantVersion:     2,
			wantContent:     "gray-a-content",
			wantType:        conftypes.ReleaseTypeGray,
		},
		{
			name:            "client matching second gray gets gray content",
			clientID:        "client-b",
			wantReleaseName: "gray-b",
			wantVersion:     3,
			wantContent:     "gray-b-content",
			wantType:        conftypes.ReleaseTypeGray,
		},
		{
			name:            "client matching multiple gray releases gets newest version",
			clientID:        "client-both",
			wantReleaseName: "gray-b",
			wantVersion:     3,
			wantContent:     "gray-b-content",
			wantType:        conftypes.ReleaseTypeGray,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := svr.GetConfigFileWithCache(context.Background(), &apiconfig.ConfigFile{
				Namespace: namespace,
				Group:     group,
				Name:      fileName,
				Labels: map[string]string{
					types.ClientLabel_ID: tt.clientID,
				},
			})

			require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
			require.Equal(t, apiconfig.ConfigDiscoverResponse_CONFIG_FILE, resp.GetType())
			require.NotNil(t, resp.GetFile())
			require.Equal(t, namespace, resp.GetFile().GetNamespace())
			require.Equal(t, group, resp.GetFile().GetGroup())
			require.Equal(t, fileName, resp.GetFile().GetFileName())
			require.Equal(t, tt.wantReleaseName, resp.GetFile().GetName())
			require.Equal(t, tt.wantVersion, resp.GetFile().GetVersion())
			require.Equal(t, tt.wantContent, resp.GetFile().GetContent())
			require.Equal(t, tt.wantType, resp.GetFile().GetReleaseType())
			require.Equal(t, configReleaseSnapshotRevision(
				map[string]*conftypes.ConfigFileRelease{
					"normal-release": normal,
					"gray-a":         grayA,
					"gray-b":         grayB,
				}[tt.wantReleaseName],
			), resp.GetRevision())
		})
	}
}

func TestGetConfigFileWithCacheFallsBackFromStoppedGrayEvenWhenNormalVersionIsOlder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	normal := configClientReleaseForTest("normal-release", conftypes.ReleaseTypeNormal, 2, "normal-content")
	gray := configClientReleaseForTest("gray-release", conftypes.ReleaseTypeGray, 5, "gray-content")
	fileCache := cachemock.NewMockConfigFileCache(ctrl)
	grayCache := cachemock.NewMockGrayCache(ctrl)
	fileCache.EXPECT().GetActiveGrayReleases("default", "group-a", "app.yaml").
		Return([]*conftypes.ConfigFileRelease{gray})
	grayCache.EXPECT().HitGrayRule(GetGrayConfigReaseKey(gray.SimpleConfigFileRelease), gomock.Any()).Return(true)
	fileCache.EXPECT().GetActiveGrayReleases("default", "group-a", "app.yaml").
		Return(nil)
	fileCache.EXPECT().GetActiveRelease("default", "group-a", "app.yaml").Return(normal)

	svr := &Server{fileCache: fileCache, grayCache: grayCache}
	first := svr.GetConfigFileWithCache(context.Background(), &apiconfig.ConfigFile{
		Namespace: "default", Group: "group-a", Name: "app.yaml",
	})
	second := svr.GetConfigFileWithCache(context.Background(), &apiconfig.ConfigFile{
		Id: first.GetRevision(), Namespace: "default", Group: "group-a", Name: "app.yaml",
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), second.GetCode())
	require.Equal(t, "normal-release", second.GetFile().GetName())
	require.NotEqual(t, first.GetRevision(), second.GetRevision())
}

func configClientReleaseForTest(name string, releaseType rules.ReleaseType, version uint64,
	content string) *conftypes.ConfigFileRelease {
	return &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Id:          name + "-id",
				Name:        name,
				Namespace:   "default",
				Group:       "group-a",
				FileName:    "app.yaml",
				ReleaseType: releaseType,
			},
			Version:    version,
			Md5:        name + "-md5",
			Active:     true,
			Valid:      true,
			Format:     string(conftypes.FileFormatYaml),
			ModifyTime: time.Unix(int64(version), 0),
			ModifyBy:   "admin",
		},
		Content: content,
	}
}
