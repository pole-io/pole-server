package config

import (
	"sort"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
)

func selectMatchedGrayRelease(releases []*conftypes.ConfigFileRelease,
	match func(*conftypes.SimpleConfigFileRelease) bool) *conftypes.ConfigFileRelease {
	candidates := make([]*conftypes.ConfigFileRelease, 0, len(releases))
	for _, release := range releases {
		if release == nil || release.SimpleConfigFileRelease == nil {
			continue
		}
		if match(release.SimpleConfigFileRelease) {
			candidates = append(candidates, release)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Version == candidates[j].Version {
			return candidates[i].ModifyTime.After(candidates[j].ModifyTime)
		}
		return candidates[i].Version > candidates[j].Version
	})
	return candidates[0]
}
