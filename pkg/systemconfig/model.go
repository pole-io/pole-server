package systemconfig

import "context"

type Component string

const (
	ComponentServer  Component = "pole-server"
	ComponentConsole Component = "pole-console"
)

type ValueType string

const (
	ValueTypeString   ValueType = "string"
	ValueTypeBoolean  ValueType = "boolean"
	ValueTypeInteger  ValueType = "integer"
	ValueTypeDuration ValueType = "duration"
	ValueTypeList     ValueType = "list"
	ValueTypeSecret   ValueType = "secret"
)

type ApplyMode string

const (
	BootstrapOnly    ApplyMode = "BootstrapOnly"
	RestartRequired  ApplyMode = "RestartRequired"
	HotReload        ApplyMode = "HotReload"
	GuardedHotReload ApplyMode = "GuardedHotReload"
)

type Sensitivity string

const (
	SensitivityPublic   Sensitivity = "public"
	SensitivityInternal Sensitivity = "internal"
	SensitivitySecret   Sensitivity = "secret"
)

type SourceKind string

const (
	SourceCompiledDefault SourceKind = "compiled_default"
	SourceStaticFile      SourceKind = "static_file"
	SourceEnvironment     SourceKind = "environment"
	SourceCommandLine     SourceKind = "command_line"
	SourceDynamicRelease  SourceKind = "dynamic_release"
)

type SourceDescriptor struct {
	Kind      SourceKind `json:"kind"`
	Reference string     `json:"reference,omitempty"`
}

type SettingDefinition struct {
	Key         string      `json:"key"`
	Component   Component   `json:"component"`
	Domain      string      `json:"domain"`
	Label       string      `json:"label"`
	Description string      `json:"description,omitempty"`
	ValueType   ValueType   `json:"value_type"`
	ApplyMode   ApplyMode   `json:"apply_mode"`
	Sensitivity Sensitivity `json:"sensitivity"`
	Owner       string      `json:"owner"`
	Editable    bool        `json:"editable"`
	EditReason  string      `json:"edit_reason,omitempty"`
	Validation  Validation  `json:"validation"`
}

// Validation is intentionally small and transport-friendly. It describes the
// product contract shared by the Console editor and the server-side validator;
// it is not a reflection of arbitrary YAML schema.
type Validation struct {
	Required    bool     `json:"required,omitempty"`
	Format      string   `json:"format,omitempty"`
	Options     []string `json:"options,omitempty"`
	MinInteger  *int64   `json:"min_integer,omitempty"`
	MaxInteger  *int64   `json:"max_integer,omitempty"`
	MinDuration string   `json:"min_duration,omitempty"`
	MaxDuration string   `json:"max_duration,omitempty"`
}

type EffectiveSetting struct {
	SettingDefinition
	Value               any              `json:"value,omitempty"`
	DisplayValue        string           `json:"display_value"`
	Configured          bool             `json:"configured"`
	Redacted            bool             `json:"redacted"`
	Source              SourceDescriptor `json:"source"`
	DesiredValue        any              `json:"desired_value,omitempty"`
	DesiredDisplayValue string           `json:"desired_display_value,omitempty"`
	DesiredRevision     int64            `json:"desired_revision,omitempty"`
	Drifted             bool             `json:"drifted,omitempty"`
	ApplyStatus         string           `json:"apply_status,omitempty"`
}

type EffectiveSnapshot struct {
	Component Component          `json:"component"`
	Settings  []EffectiveSetting `json:"settings"`
}

type Scope struct {
	Component Component
	Domain    string
}

type EffectiveProvider interface {
	Describe(context.Context, Scope) ([]SettingDefinition, error)
	Effective(context.Context, Scope) (*EffectiveSnapshot, error)
}
