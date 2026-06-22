package rules

import (
	"encoding/json"
	"testing"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	servicetypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/stretchr/testify/require"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
)

func TestGetServicesInvolveByFaultDetectRuleUsesRuleLevelTarget(t *testing.T) {
	spec := &apifault.FaultDetectRule{
		TargetService: &apifault.FaultDetectRule_DestinationService{
			Namespace: "default",
			Service:   "checkout",
		},
		Rules: []*apifault.FaultDetectSubRule{
			{
				Interval: 5,
			},
		},
	}
	rule := &ruletypes.FaultDetectRule{
		DstNamespace: "default",
		DstService:   "checkout",
	}
	data, err := json.Marshal(spec)
	require.NoError(t, err)
	rule.Rule = string(data)

	services := getServicesInvolveByFaultDetectRule(&ruletypes.FaultDetectRelease{Rule: rule})
	require.Contains(t, services, servicetypes.ServiceKey{Namespace: "default", Name: "checkout"})
	require.NotContains(t, services, servicetypes.ServiceKey{Namespace: "default", Name: "payment"})
}
