package namespace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSumConfigFileCounts(t *testing.T) {
	totals := sumConfigFileCounts(map[string]map[string]int64{
		"default": {
			"base":   2,
			"canary": 3,
		},
		"empty": {
			"base": 0,
		},
		"invalid": {
			"base": -1,
		},
	})

	require.Equal(t, uint32(5), totals["default"])
	require.Zero(t, totals["empty"])
	require.Zero(t, totals["invalid"])
}
