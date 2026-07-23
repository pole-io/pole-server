package systemconfig

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRegistryEffectivePreservesSourceAndRedactsSecret(t *testing.T) {
	sources := SourceIndex{
		"runtime.mode":  {Kind: SourceStaticFile},
		"runtime.token": {Kind: SourceEnvironment, Reference: "env:TEST_SYSTEM_TOKEN"},
	}
	registry, err := NewRegistry([]Registration{
		{
			Definition: SettingDefinition{
				Key: "console.runtime.mode", Component: ComponentConsole, Domain: "runtime",
				Label: "运行模式", ValueType: ValueTypeString, ApplyMode: RestartRequired,
				Sensitivity: SensitivityPublic, Owner: "console",
			},
			SourcePath: "runtime.mode",
			Value:      func() any { return "release" },
		},
		{
			Definition: SettingDefinition{
				Key: "console.runtime.token", Component: ComponentConsole, Domain: "runtime",
				Label: "访问令牌", ValueType: ValueTypeSecret, ApplyMode: BootstrapOnly,
				Sensitivity: SensitivitySecret, Owner: "console",
			},
			SourcePath: "runtime.token",
			Value:      func() any { return "must-not-leak" },
		},
	}, sources)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	snapshot, err := registry.Effective(context.Background(), Scope{Component: ComponentConsole})
	if err != nil {
		t.Fatalf("Effective() error = %v", err)
	}
	if len(snapshot.Settings) != 2 {
		t.Fatalf("len(settings) = %d, want 2", len(snapshot.Settings))
	}
	if got := snapshot.Settings[0].Source.Kind; got != SourceStaticFile {
		t.Fatalf("mode source = %q, want %q", got, SourceStaticFile)
	}
	secret := snapshot.Settings[1]
	if !secret.Configured || !secret.Redacted || secret.Value != nil {
		t.Fatalf("secret setting = %#v, want configured redacted value=nil", secret)
	}
	if secret.Source.Reference != "env:TEST_SYSTEM_TOKEN" {
		t.Fatalf("secret source reference = %q", secret.Source.Reference)
	}
	body, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(body) == "" || contains(string(body), "must-not-leak") {
		t.Fatalf("serialized snapshot leaked secret: %s", body)
	}
}

func TestRegistryRejectsDuplicateKey(t *testing.T) {
	definition := SettingDefinition{
		Key: "server.cache.diff_time", Component: ComponentServer, Domain: "cache",
		Label: "增量回溯窗口", ValueType: ValueTypeDuration, ApplyMode: RestartRequired,
		Sensitivity: SensitivityPublic, Owner: "server",
	}
	_, err := NewRegistry([]Registration{
		{Definition: definition, Value: func() any { return "5s" }},
		{Definition: definition, Value: func() any { return "10s" }},
	}, nil)
	if err == nil {
		t.Fatal("NewRegistry() error = nil, want duplicate-key error")
	}
}

func TestLookupValueSupportsYAMLNestedMaps(t *testing.T) {
	value := map[string]interface{}{
		"master": map[interface{}]interface{}{"dns": "configured"},
	}
	if got := LookupValue(value, "master", "dns"); got != "configured" {
		t.Fatalf("LookupValue() = %#v, want configured", got)
	}
}

func contains(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
