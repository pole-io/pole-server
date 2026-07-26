package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"

	regexp "github.com/dlclark/regexp2"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
)

func selectGovernanceReleases(
	releases []*ruletypes.RuleRelease, labels map[string]string,
) ([]*ruletypes.RuleRelease, string) {
	byRule := make(map[string][]*ruletypes.RuleRelease)
	for _, release := range releases {
		if release == nil || !release.Active || !release.Valid {
			continue
		}
		byRule[release.RuleId] = append(byRule[release.RuleId], release)
	}

	selected := make([]*ruletypes.RuleRelease, 0, len(byRule))
	for _, candidates := range byRule {
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Version == candidates[j].Version {
				return candidates[i].Mtime.After(candidates[j].Mtime)
			}
			return candidates[i].Version > candidates[j].Version
		})
		var normal *ruletypes.RuleRelease
		for _, candidate := range candidates {
			if candidate.ReleaseType == ruletypes.ReleaseTypeGray && governanceLabelsMatch(candidate.ClientLabels, labels) {
				selected = append(selected, candidate)
				normal = nil
				break
			}
			if normal == nil && candidate.ReleaseType == ruletypes.ReleaseTypeNormal {
				normal = candidate
			}
		}
		if normal != nil {
			selected = append(selected, normal)
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].RuleId == selected[j].RuleId {
			return selected[i].ReleaseName < selected[j].ReleaseName
		}
		return selected[i].RuleId < selected[j].RuleId
	})

	hash := sha256.New()
	for _, release := range selected {
		_, _ = hash.Write([]byte(release.RuleId))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(release.Id))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(release.ReleaseName))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(strconv.FormatUint(release.Version, 10)))
		_, _ = hash.Write([]byte{0})
	}
	if len(selected) == 0 {
		return selected, ""
	}
	return selected, hex.EncodeToString(hash.Sum(nil))
}

func governanceLabelsMatch(expected []*apimodel.ClientLabel, actual map[string]string) bool {
	if len(expected) == 0 {
		return true
	}
	for _, label := range expected {
		value, ok := actual[label.GetKey()]
		if !ok || !matchs.MatchString(value, label.GetValue(), func(pattern string) *regexp.Regexp {
			compiled, err := regexp.Compile(pattern, regexp.RE2)
			if err != nil {
				return nil
			}
			return compiled
		}) {
			return false
		}
	}
	return true
}
