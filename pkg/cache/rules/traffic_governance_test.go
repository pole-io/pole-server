package rules

import (
	"testing"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/stretchr/testify/require"
)

func TestTrafficGovernanceFilterMatchedIgnoresEmptyFilters(t *testing.T) {
	rule := &ruletypes.TrafficGovernanceRule{
		ID:        "rule-id",
		Name:      "traffic-rule",
		Namespace: "default",
		Service:   "checkout",
	}

	require.True(t, trafficGovernanceFilterMatched(
		rule,
		"", true,
		"", true,
		"", true,
		"", true,
	))
	require.True(t, trafficGovernanceFilterMatched(
		rule,
		"rule-id", true,
		"default", true,
		"checkout", true,
		"traffic-rule", true,
	))
	require.False(t, trafficGovernanceFilterMatched(
		rule,
		"other-rule", true,
		"default", true,
		"checkout", true,
		"traffic-rule", true,
	))
	require.False(t, trafficGovernanceFilterMatched(
		rule,
		"rule-id", true,
		"default", true,
		"checkout", true,
		"other-rule", true,
	))
}
