package rules

import (
	"testing"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
)

func TestFaultDetectRuleToSpecMigratesLegacyTopLevelTargetAndProbeFields(t *testing.T) {
	rule := &FaultDetectRule{
		ID:          "fd-1",
		Name:        "legacy-fault-detect",
		Description: "legacy payload",
		Rule: `{
			"target_service": {
				"namespace": "default",
				"service": "checkout",
				"api": {
					"protocol": "HTTP",
					"method": "GET",
					"path": { "type": "EXACT", "value": "/health", "value_type": "TEXT" }
				}
			},
			"interval": 10,
			"timeout": 3,
			"port": 8080,
			"protocol": "HTTP",
			"http_config": { "method": "GET", "url": "/health" }
		}`,
	}

	spec, err := rule.ToSpec()
	if err != nil {
		t.Fatalf("ToSpec() error = %v", err)
	}
	if got := len(spec.GetRules()); got != 1 {
		t.Fatalf("len(spec.Rules) = %d, want 1", got)
	}
	if spec.GetTargetService().GetNamespace() != "default" {
		t.Fatalf("namespace = %q, want default", spec.GetTargetService().GetNamespace())
	}
	if spec.GetTargetService().GetService() != "checkout" {
		t.Fatalf("service = %q, want checkout", spec.GetTargetService().GetService())
	}
	subRule := spec.GetRules()[0]
	if subRule.GetInterval() != 10 || subRule.GetTimeout() != 3 || subRule.GetPort() != 8080 {
		t.Fatalf("probe timing/port = %d/%d/%d, want 10/3/8080", subRule.GetInterval(), subRule.GetTimeout(), subRule.GetPort())
	}
	if subRule.GetProtocol() != apifault.FaultDetectRule_HTTP {
		t.Fatalf("protocol = %s, want HTTP", subRule.GetProtocol())
	}
	if subRule.GetHttpConfig().GetUrl() != "/health" {
		t.Fatalf("http url = %q, want /health", subRule.GetHttpConfig().GetUrl())
	}
}

func TestFaultDetectRuleToSpecKeepsRuleLevelTargetAboveSubRules(t *testing.T) {
	rule := &FaultDetectRule{
		ID:   "fd-2",
		Name: "subrules-fault-detect",
		Rule: `{
			"target_service": {
				"namespace": "prod",
				"service": "payment"
			},
			"rules": [{
				"interval": 5,
				"timeout": 2,
				"port": 9090,
				"protocol": 2,
				"tcp_config": { "send": "PING", "receive": ["PONG"] }
			}]
		}`,
	}

	spec, err := rule.ToSpec()
	if err != nil {
		t.Fatalf("ToSpec() error = %v", err)
	}
	if got := len(spec.GetRules()); got != 1 {
		t.Fatalf("len(spec.Rules) = %d, want 1", got)
	}
	if spec.GetTargetService().GetNamespace() != "prod" || spec.GetTargetService().GetService() != "payment" {
		t.Fatalf("target = %s/%s, want prod/payment", spec.GetTargetService().GetNamespace(), spec.GetTargetService().GetService())
	}
	subRule := spec.GetRules()[0]
	if subRule.GetProtocol() != apifault.FaultDetectRule_TCP {
		t.Fatalf("protocol = %s, want TCP", subRule.GetProtocol())
	}
	if subRule.GetTcpConfig().GetSend() != "PING" {
		t.Fatalf("tcp send = %q, want PING", subRule.GetTcpConfig().GetSend())
	}
}

func TestFaultDetectRuleToSpecPromotesLegacySubRuleTarget(t *testing.T) {
	rule := &FaultDetectRule{
		ID:   "fd-3",
		Name: "legacy-subrule-target",
		Rule: `{
			"rules": [{
				"target_service": {
					"namespace": "legacy",
					"service": "order"
				},
				"interval": 7,
				"timeout": 2,
				"port": 8081,
				"protocol": "HTTP"
			}]
		}`,
	}

	spec, err := rule.ToSpec()
	if err != nil {
		t.Fatalf("ToSpec() error = %v", err)
	}
	if spec.GetTargetService().GetNamespace() != "legacy" || spec.GetTargetService().GetService() != "order" {
		t.Fatalf("target = %s/%s, want legacy/order", spec.GetTargetService().GetNamespace(), spec.GetTargetService().GetService())
	}
	if got := len(spec.GetRules()); got != 1 {
		t.Fatalf("len(spec.Rules) = %d, want 1", got)
	}
	if spec.GetRules()[0].GetInterval() != 7 {
		t.Fatalf("interval = %d, want 7", spec.GetRules()[0].GetInterval())
	}
}
