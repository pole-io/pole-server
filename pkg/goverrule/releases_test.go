package goverrule

import (
	"testing"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestBuildRouterRuleReleaseAllowsNilRule(t *testing.T) {
	req := &apimodel.RuleRelease{
		RuleId:      "route-id",
		RuleName:    "route-name",
		ReleaseName: "normal",
		ReleaseType: string(rules.ReleaseTypeNormal),
		Resource:    apimodel.RuleRelease_RouteRules,
	}

	release := buildRouterRuleRelease(req, nil)
	if release == nil {
		t.Fatal("expected release")
	}
	if release.Rule != nil {
		t.Fatalf("expected nil route rule, got %#v", release.Rule)
	}
	if release.RuleId != req.RuleId {
		t.Fatalf("expected rule id %q, got %q", req.RuleId, release.RuleId)
	}
	if release.RuleName != req.RuleName {
		t.Fatalf("expected rule name %q, got %q", req.RuleName, release.RuleName)
	}
	if release.ReleaseType != rules.ReleaseTypeNormal {
		t.Fatalf("expected release type %q, got %q", rules.ReleaseTypeNormal, release.ReleaseType)
	}
	if release.Resource != apimodel.RuleRelease_RouteRules {
		t.Fatalf("expected resource %q, got %q", apimodel.RuleRelease_RouteRules, release.Resource)
	}
	if release.Id == "" {
		t.Fatal("expected generated release id")
	}
}

func TestNewRuleReleaseFromSpecGeneratesMissingID(t *testing.T) {
	release := newRuleReleaseFromSpec(&apimodel.RuleRelease{
		RuleId:      "rule-id",
		RuleName:    "rule-name",
		ReleaseName: "normal",
		ReleaseType: string(rules.ReleaseTypeNormal),
		Resource:    apimodel.RuleRelease_CircuitBreakerRules,
	})
	if release.Id == "" {
		t.Fatal("expected generated release id")
	}
}

func TestNewRuleReleaseFromSpecPreservesProvidedID(t *testing.T) {
	release := newRuleReleaseFromSpec(&apimodel.RuleRelease{
		Id:          "release-id",
		RuleId:      "rule-id",
		RuleName:    "rule-name",
		ReleaseName: "normal",
		ReleaseType: string(rules.ReleaseTypeNormal),
		Resource:    apimodel.RuleRelease_CircuitBreakerRules,
	})
	if release.Id != "release-id" {
		t.Fatalf("expected release id %q, got %q", "release-id", release.Id)
	}
}
