package systemsettings

import "time"

const (
	AgentComponent     = "pole-console"
	AgentDomain        = "agent"
	agentAPIKeyPurpose = "llm-gateway-api-key"
)

type AgentProfile struct {
	RuntimeMode          string   `json:"runtimeMode"`
	AgentID              string   `json:"agentId"`
	PromptVersion        string   `json:"promptVersion"`
	OperatorInstructions string   `json:"operatorInstructions"`
	Provider             string   `json:"provider"`
	BaseURL              string   `json:"baseURL"`
	Model                string   `json:"model"`
	ModelTimeout         string   `json:"modelTimeout"`
	MCPEndpoint          string   `json:"mcpEndpoint"`
	MCPToolAllowlist     []string `json:"mcpToolAllowlist"`
	ProposalTTL          string   `json:"proposalTTL"`
	UpstreamTimeout      string   `json:"upstreamTimeout"`
}

type SecretMetadata struct {
	Configured  bool       `json:"configured"`
	Version     int64      `json:"version,omitempty"`
	Reference   string     `json:"reference,omitempty"`
	Fingerprint string     `json:"fingerprint,omitempty"`
	RotatedAt   *time.Time `json:"rotatedAt,omitempty"`
}

type RevisionView struct {
	Revision    int64          `json:"revision"`
	State       string         `json:"state"`
	Values      AgentProfile   `json:"values"`
	Secret      SecretMetadata `json:"secret"`
	CreatedBy   string         `json:"createdBy"`
	PublishedBy string         `json:"publishedBy,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	PublishedAt *time.Time     `json:"publishedAt,omitempty"`
}

type DomainView struct {
	Component         string        `json:"component"`
	Domain            string        `json:"domain"`
	Active            *RevisionView `json:"active,omitempty"`
	Draft             *RevisionView `json:"draft,omitempty"`
	EffectiveRevision string        `json:"effectiveRevision"`
	ApplyStatus       string        `json:"applyStatus"`
	ApplyMessage      string        `json:"applyMessage,omitempty"`
	SecretStoreReady  bool          `json:"secretStoreReady"`
}

type SaveDraftRequest struct {
	ExpectedDraftRevision int64         `json:"expectedDraftRevision"`
	Values                AgentProfile  `json:"values"`
	APIKey                *SecretUpdate `json:"apiKey,omitempty"`
}

type SecretUpdate struct {
	Operation string `json:"operation"`
	Value     string `json:"value,omitempty"`
}

type PublishRequest struct {
	DraftRevision int64  `json:"draftRevision"`
	Description   string `json:"description,omitempty"`
}

type ConnectionTestResult struct {
	Success   bool   `json:"success"`
	LatencyMS int64  `json:"latencyMs"`
	Model     string `json:"model,omitempty"`
	RequestID string `json:"requestId,omitempty"`
	Message   string `json:"message"`
}

type ManagedRevisionView struct {
	Revision    int64          `json:"revision"`
	State       string         `json:"state"`
	Values      map[string]any `json:"values"`
	CreatedBy   string         `json:"createdBy"`
	PublishedBy string         `json:"publishedBy,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	PublishedAt *time.Time     `json:"publishedAt,omitempty"`
}

type ManagedDomainView struct {
	Component    string               `json:"component"`
	Domain       string               `json:"domain"`
	Active       *ManagedRevisionView `json:"active,omitempty"`
	Draft        *ManagedRevisionView `json:"draft,omitempty"`
	ApplyStatus  string               `json:"applyStatus"`
	ApplyMessage string               `json:"applyMessage,omitempty"`
}

type SaveManagedDraftRequest struct {
	ExpectedDraftRevision int64          `json:"expectedDraftRevision"`
	Values                map[string]any `json:"values"`
}
