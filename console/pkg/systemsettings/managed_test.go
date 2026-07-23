package systemsettings

import (
	"testing"

	"github.com/pole-io/pole-server/pkg/systemconfig"
)

func TestValidateManagedValuesRejectsLockedAndInvalidSettings(t *testing.T) {
	minimum := int64(1)
	settings := []systemconfig.EffectiveSetting{
		{SettingDefinition: systemconfig.SettingDefinition{
			Key: "console.web.listen_port", Component: systemconfig.ComponentConsole, Domain: "runtime",
			Label: "监听端口", ValueType: systemconfig.ValueTypeInteger, Editable: false,
		}},
		{SettingDefinition: systemconfig.SettingDefinition{
			Key: "console.features", Component: systemconfig.ComponentConsole, Domain: "runtime",
			Label: "功能开关", ValueType: systemconfig.ValueTypeList, Editable: true,
			Validation: systemconfig.Validation{MinInteger: &minimum},
		}},
	}
	if _, err := validateManagedValues("pole-console", "runtime", settings, map[string]any{
		"console.web.listen_port": float64(8081),
		"console.features":        []any{"agent"},
	}); err == nil {
		t.Fatal("locked setting update unexpectedly succeeded")
	}
	if _, err := validateManagedValues("pole-console", "runtime", settings, map[string]any{
		"console.features": "agent",
	}); err == nil {
		t.Fatal("invalid list update unexpectedly succeeded")
	}
	values, err := validateManagedValues("pole-console", "runtime", settings, map[string]any{
		"console.features": []any{"agent", "metrics"},
	})
	if err != nil {
		t.Fatalf("valid update failed: %v", err)
	}
	if got := len(values["console.features"].([]string)); got != 2 {
		t.Fatalf("features length = %d, want 2", got)
	}
}

func TestValidateManagedValuesChecksCrossFieldHealthIntervals(t *testing.T) {
	settings := []systemconfig.EffectiveSetting{
		{SettingDefinition: systemconfig.SettingDefinition{
			Key: "server.naming.healthcheck.min_check_interval", Component: systemconfig.ComponentServer,
			Domain: "naming", Label: "最小检查周期", ValueType: systemconfig.ValueTypeDuration, Editable: true,
		}},
		{SettingDefinition: systemconfig.SettingDefinition{
			Key: "server.naming.healthcheck.max_check_interval", Component: systemconfig.ComponentServer,
			Domain: "naming", Label: "最大检查周期", ValueType: systemconfig.ValueTypeDuration, Editable: true,
		}},
	}
	_, err := validateManagedValues("pole-server", "naming", settings, map[string]any{
		"server.naming.healthcheck.min_check_interval": "30s",
		"server.naming.healthcheck.max_check_interval": "10s",
	})
	if err == nil {
		t.Fatal("inverted healthcheck interval unexpectedly succeeded")
	}
}
