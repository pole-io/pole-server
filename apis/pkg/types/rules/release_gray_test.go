package rules

import (
	"testing"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestGrayReleaseKeyIncludesReleaseName(t *testing.T) {
	a := &RuleRelease{RuleName: "rule-a", RuleId: "rule-id", ReleaseName: "gray-a", ReleaseType: ReleaseTypeGray}
	b := &RuleRelease{RuleName: "rule-a", RuleId: "rule-id", ReleaseName: "gray-b", ReleaseType: ReleaseTypeGray}
	if a.Key() == b.Key() {
		t.Fatalf("gray release keys must differ: %q", a.Key())
	}
}

func TestReleaseKeyIncludesNamespaceAndRuleID(t *testing.T) {
	base := RuleRelease{
		Namespace: "prod", RuleId: "rule-id-a", RuleName: "same-name",
		ReleaseName: "normal", ReleaseType: ReleaseTypeNormal,
	}
	otherNamespace := base
	otherNamespace.Namespace = "staging"
	otherRule := base
	otherRule.RuleId = "rule-id-b"

	if base.Key() == otherNamespace.Key() {
		t.Fatalf("release keys collide across namespaces: %q", base.Key())
	}
	if base.Key() == otherRule.Key() {
		t.Fatalf("release keys collide across rule IDs: %q", base.Key())
	}
}

func TestGrayResourceIncludesReleaseName(t *testing.T) {
	release := &RuleRelease{
		Resource: apimodel.RuleRelease_RouteRules, RuleId: "rule-id",
		ReleaseName: "gray-a", ReleaseType: ReleaseTypeGray,
	}
	if got, want := release.GetGrayResource(), "RouteRules/rule-id/gray-a"; got != want {
		t.Fatalf("gray resource mismatch: got %q want %q", got, want)
	}
}

func TestRuleReleaseSpecRoundTripKeepsClientLabels(t *testing.T) {
	spec := &apimodel.RuleRelease{
		ReleaseName: "gray-a", ReleaseType: string(ReleaseTypeGray),
		ClientLabels: exactReleaseLabels("env", "canary"),
	}
	release := &RuleRelease{}
	release.FromSpec(spec)
	if got := release.GetClientLabels(); len(got) != 1 || got[0].GetKey() != "env" {
		t.Fatalf("client labels lost when reading spec: %#v", got)
	}
	if got := release.ToSpec().GetClientLabels(); len(got) != 1 || got[0].GetValue().GetValue() != "canary" {
		t.Fatalf("client labels lost when writing spec: %#v", got)
	}
}

func exactReleaseLabels(key, value string) []*apimodel.ClientLabel {
	return []*apimodel.ClientLabel{{
		Key: key,
		Value: &apimodel.MatchString{
			Type: apimodel.MatchString_EXACT, Value: value,
		},
	}}
}
