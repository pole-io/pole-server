package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func TestConfigFileCacheKeepsMultipleActiveGrayReleases(t *testing.T) {
	cache := &fileCache{}
	require.NoError(t, cache.Initialize(map[string]interface{}{"cachePath": t.TempDir()}))
	t.Cleanup(func() { _ = cache.Close() })

	grayA := configFileReleaseForCacheTest("gray-a", conftypes.ReleaseTypeGray, 2)
	grayB := configFileReleaseForCacheTest("gray-b", conftypes.ReleaseTypeGray, 3)

	require.NoError(t, cache.saveActiveRelease(grayA))
	require.NoError(t, cache.saveActiveRelease(grayB))

	releases := cache.GetActiveGrayReleases("default", "group-a", "app.yaml")
	require.Len(t, releases, 2)
	require.Equal(t, "gray-b", releases[0].Name)
	require.Equal(t, "gray-a", releases[1].Name)
}

func configFileReleaseForCacheTest(name string, releaseType rules.ReleaseType, version uint64) *conftypes.ConfigFileRelease {
	return &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Id:          name,
				Name:        name,
				Namespace:   "default",
				Group:       "group-a",
				FileName:    "app.yaml",
				ReleaseType: releaseType,
			},
			Version: version,
			Active:  true,
			Valid:   true,
		},
		Content: name + "-content",
	}
}
