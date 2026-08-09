package namespace

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

func TestValidateEnvironmentPromotionTopologyAcceptsBaselineDAGAndLaneBinding(t *testing.T) {
	topology := &types.EnvironmentPromotionTopology{
		Edges: []types.EnvironmentPromotionEdge{
			{ID: "dev-tst", Source: "dev", Target: "tst", ConflictPolicy: "BLOCK"},
			{ID: "tst-pre", Source: "tst", Target: "pre", ConflictPolicy: "BLOCK"},
			{ID: "pre-pro", Source: "pre", Target: "pro", ConflictPolicy: "BLOCK"},
		},
		LaneBaseBindings: []types.LaneBaseBinding{{Lane: "feature-a", Base: "dev"}},
	}

	require.Empty(t, ValidateEnvironmentPromotionTopology(topology))
}

func TestValidateEnvironmentPromotionTopologyRejectsCycle(t *testing.T) {
	topology := &types.EnvironmentPromotionTopology{Edges: []types.EnvironmentPromotionEdge{
		{ID: "a-b", Source: "a", Target: "b"},
		{ID: "b-a", Source: "b", Target: "a"},
	}}

	issues := ValidateEnvironmentPromotionTopology(topology)
	require.Contains(t, issueCodes(issues), "BASELINE_DAG_CYCLE")
}

func TestValidateEnvironmentPromotionTopologyRejectsLaneInsideBaselineDAG(t *testing.T) {
	topology := &types.EnvironmentPromotionTopology{
		Edges:            []types.EnvironmentPromotionEdge{{ID: "dev-lane", Source: "dev", Target: "lane-a"}},
		LaneBaseBindings: []types.LaneBaseBinding{{Lane: "lane-a", Base: "dev"}},
	}

	issues := ValidateEnvironmentPromotionTopology(topology)
	require.Contains(t, issueCodes(issues), "LANE_IN_BASELINE_DAG")
}

func TestValidateEnvironmentPromotionTopologyRejectsDuplicateLaneBinding(t *testing.T) {
	topology := &types.EnvironmentPromotionTopology{LaneBaseBindings: []types.LaneBaseBinding{
		{Lane: "lane-a", Base: "dev"},
		{Lane: "lane-a", Base: "tst"},
	}}

	issues := ValidateEnvironmentPromotionTopology(topology)
	require.Contains(t, issueCodes(issues), "LANE_BINDING_DUPLICATED")
}

func issueCodes(issues []types.TopologyValidationIssue) []string {
	codes := make([]string, 0, len(issues))
	for _, issue := range issues {
		codes = append(codes, issue.Code)
	}
	return codes
}
