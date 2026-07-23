package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
)

func TestSelectMatchedGrayReleasePrefersNewestMatchedVersion(t *testing.T) {
	oldMatched := configGrayReleaseForSelectTest("gray-old", 2, time.Unix(2, 0))
	newMatched := configGrayReleaseForSelectTest("gray-new", 4, time.Unix(4, 0))
	newUnmatched := configGrayReleaseForSelectTest("gray-unmatched", 5, time.Unix(5, 0))

	selected := selectMatchedGrayRelease([]*conftypes.ConfigFileRelease{oldMatched, newUnmatched, newMatched},
		func(release *conftypes.SimpleConfigFileRelease) bool {
			return release.Name != "gray-unmatched"
		})

	require.NotNil(t, selected)
	require.Equal(t, "gray-new", selected.Name)
}

func TestSelectMatchedGrayReleaseTieBreaksByModifyTime(t *testing.T) {
	older := configGrayReleaseForSelectTest("gray-older", 3, time.Unix(2, 0))
	newer := configGrayReleaseForSelectTest("gray-newer", 3, time.Unix(3, 0))

	selected := selectMatchedGrayRelease([]*conftypes.ConfigFileRelease{older, newer},
		func(release *conftypes.SimpleConfigFileRelease) bool { return true })

	require.NotNil(t, selected)
	require.Equal(t, "gray-newer", selected.Name)
}

func configGrayReleaseForSelectTest(name string, version uint64, modifyTime time.Time) *conftypes.ConfigFileRelease {
	return &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Name: name,
			},
			Version:    version,
			ModifyTime: modifyTime,
		},
	}
}
