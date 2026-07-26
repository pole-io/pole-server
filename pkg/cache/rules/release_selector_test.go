package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func TestSelectGovernanceReleasePrefersNewestMatchedGrayPerRule(t *testing.T) {
	normal := releaseForSelection("normal", ruletypes.ReleaseTypeNormal, 7, time.Unix(70, 0), nil)
	grayOld := releaseForSelection("gray-old", ruletypes.ReleaseTypeGray, 8, time.Unix(80, 0), exactLabel("env", "canary"))
	grayNew := releaseForSelection("gray-new", ruletypes.ReleaseTypeGray, 9, time.Unix(90, 0), exactLabel("env", "canary"))

	selected, revision := selectGovernanceReleases(
		[]*ruletypes.RuleRelease{normal, grayOld, grayNew},
		map[string]string{"env": "canary"},
	)

	require.Len(t, selected, 1)
	require.Equal(t, "gray-new", selected[0].ReleaseName)
	require.NotEmpty(t, revision)
}

func TestSelectGovernanceReleaseFallsBackToNormalAndRevisionIdentifiesSnapshot(t *testing.T) {
	normal := releaseForSelection("normal", ruletypes.ReleaseTypeNormal, 7, time.Unix(70, 0), nil)
	gray := releaseForSelection("gray", ruletypes.ReleaseTypeGray, 9, time.Unix(90, 0), exactLabel("env", "canary"))

	normalSelected, normalRevision := selectGovernanceReleases(
		[]*ruletypes.RuleRelease{normal, gray},
		map[string]string{"env": "stable"},
	)
	graySelected, grayRevision := selectGovernanceReleases(
		[]*ruletypes.RuleRelease{normal, gray},
		map[string]string{"env": "canary"},
	)

	require.Equal(t, "normal", normalSelected[0].ReleaseName)
	require.Equal(t, "gray", graySelected[0].ReleaseName)
	require.NotEqual(t, normalRevision, grayRevision)

	again, againRevision := selectGovernanceReleases(
		[]*ruletypes.RuleRelease{gray, normal},
		map[string]string{"env": "stable"},
	)
	require.Equal(t, "normal", again[0].ReleaseName)
	require.Equal(t, normalRevision, againRevision)
}

func TestSelectGovernanceReleaseFallsBackToNormalAfterGrayStops(t *testing.T) {
	normal := releaseForSelection("normal", ruletypes.ReleaseTypeNormal, 4, time.Unix(40, 0), nil)
	gray := releaseForSelection("gray", ruletypes.ReleaseTypeGray, 8, time.Unix(80, 0), exactLabel("env", "canary"))

	selectedGray, grayRevision := selectGovernanceReleases(
		[]*ruletypes.RuleRelease{normal, gray}, map[string]string{"env": "canary"},
	)
	gray.Active = false
	selectedNormal, normalRevision := selectGovernanceReleases(
		[]*ruletypes.RuleRelease{normal, gray}, map[string]string{"env": "canary"},
	)

	require.Equal(t, "gray", selectedGray[0].ReleaseName)
	require.Equal(t, "normal", selectedNormal[0].ReleaseName)
	require.NotEqual(t, grayRevision, normalRevision)
}

func releaseForSelection(name string, releaseType ruletypes.ReleaseType, version uint64, mtime time.Time,
	labels []*apimodel.ClientLabel,
) *ruletypes.RuleRelease {
	return &ruletypes.RuleRelease{
		Id:           name + "-id",
		RuleId:       "rule-1",
		ReleaseName:  name,
		ReleaseType:  releaseType,
		ClientLabels: labels,
		Active:       true,
		Valid:        true,
		Version:      version,
		Mtime:        mtime,
	}
}

func exactLabel(key, value string) []*apimodel.ClientLabel {
	return []*apimodel.ClientLabel{{
		Key: key,
		Value: &apimodel.MatchString{
			Type:  apimodel.MatchString_EXACT,
			Value: value,
		},
	}}
}
