package ai

// A2AAgentCard 描述 A2A Agent 对外暴露的稳定能力契约。
type A2AAgentCard struct {
	Name               string                       `json:"name"`
	Description        string                       `json:"description,omitempty"`
	URL                string                       `json:"url"`
	Version            string                       `json:"version"`
	ProtocolVersion    string                       `json:"protocolVersion"`
	PreferredTransport string                       `json:"preferredTransport"`
	Capabilities       A2ACapabilities              `json:"capabilities"`
	DefaultInputModes  []string                     `json:"defaultInputModes"`
	DefaultOutputModes []string                     `json:"defaultOutputModes"`
	Skills             []A2ASkill                   `json:"skills"`
	SecuritySchemes    map[string]A2ASecurityScheme `json:"securitySchemes,omitempty"`
	Security           []map[string][]string        `json:"security,omitempty"`
}

type A2ASecurityScheme struct {
	Type   string `json:"type"`
	Scheme string `json:"scheme"`
}

type A2ACapabilities struct {
	Streaming         bool `json:"streaming"`
	PushNotifications bool `json:"pushNotifications"`
}

type A2ASkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Examples    []string `json:"examples,omitempty"`
	InputModes  []string `json:"inputModes,omitempty"`
	OutputModes []string `json:"outputModes,omitempty"`
}
