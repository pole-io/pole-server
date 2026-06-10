package goverrule_auth

import (
	"testing"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func TestCircuitBreakerRuleMetadata(t *testing.T) {
	if got := circuitBreakerRuleMetadata(nil); got != nil {
		t.Fatalf("expected nil metadata for nil rule, got %v", got)
	}

	if got := circuitBreakerRuleMetadata(&rules.CircuitBreakerRule{}); got != nil {
		t.Fatalf("expected nil metadata for empty cached rule, got %v", got)
	}

	meta := map[string]string{"demo": "governance"}
	got := circuitBreakerRuleMetadata(&rules.CircuitBreakerRule{
		Proto: &apifault.CircuitBreakerRule{
			Metadata: meta,
		},
	})
	if got["demo"] != "governance" {
		t.Fatalf("expected proto metadata, got %v", got)
	}
}
