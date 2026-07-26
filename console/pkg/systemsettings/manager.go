package systemsettings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
	store "github.com/pole-io/pole-server/console/pkg/observer"
	"github.com/pole-io/pole-server/console/pkg/poleagent"
	"github.com/pole-io/pole-server/pkg/systemconfig"
)

type RuntimeSnapshot struct {
	Revision  string
	Profile   AgentProfile
	Agent     *poleagent.Agent
	AppliedAt time.Time
}

const SelfManagerActorID = "pole-self-manager"

type candidateBuilder func(context.Context, agentworkbench.Actor, AgentProfile, string) (*poleagent.Agent, error)

type Manager struct {
	config    *bootstrap.Config
	repo      store.SystemSettingsRepository
	workbench *agentworkbench.Workbench
	envelope  *secretEnvelope
	current   atomic.Pointer[RuntimeSnapshot]
	applyMu   sync.Mutex
	statusMu  sync.RWMutex
	status    string
	message   string
	candidate candidateBuilder
}

func NewManager(config *bootstrap.Config, repo store.SystemSettingsRepository,
	workbench *agentworkbench.Workbench) (*Manager, error) {
	envelope, envelopeErr := newSecretEnvelope(strings.TrimSpace(config.SystemSecrets.MasterKey))
	manager := &Manager{config: config, repo: repo, workbench: workbench, envelope: envelope, status: "static"}
	staticProfile := profileFromBootstrap(config.Agent.Normalize())
	agent, err := manager.buildAgent(staticProfile, config.Agent.Model.APIKey)
	if err != nil {
		agent = poleagent.New(poleagent.Options{
			ID: staticProfile.AgentID, RuntimeMode: staticProfile.RuntimeMode,
			PromptVersion: staticProfile.PromptVersion, OperatorInstructions: staticProfile.OperatorInstructions,
			Model: staticProfile.Model,
		}, nil, nil, workbench)
	}
	manager.current.Store(&RuntimeSnapshot{
		Revision: "static", Profile: staticProfile, Agent: agent, AppliedAt: time.Now().UTC(),
	})
	if loadErr := manager.reload(context.Background()); loadErr != nil {
		manager.setStatus("rejected", "继续使用上一次有效配置")
	}
	interval := config.SystemSecrets.PollInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	go manager.poll(interval)
	if envelopeErr != nil {
		manager.setStatus("static", "Secret Store 尚未配置根密钥；静态 Agent 配置仍可使用")
	}
	return manager, nil
}

func (m *Manager) Current() *RuntimeSnapshot {
	return m.current.Load()
}

func (m *Manager) Domain(ctx context.Context) (*DomainView, error) {
	domain, err := m.repo.GetSystemConfigDomain(ctx, AgentComponent, AgentDomain)
	if err != nil {
		return nil, err
	}
	view := &DomainView{Component: AgentComponent, Domain: AgentDomain, SecretStoreReady: m.envelope != nil}
	if domain.ActiveRevision != nil {
		view.Active, err = m.revisionView(ctx, domain.ActiveRevision)
		if err != nil {
			return nil, err
		}
	}
	if domain.DraftRevision != nil {
		view.Draft, err = m.revisionView(ctx, domain.DraftRevision)
		if err != nil {
			return nil, err
		}
	}
	snapshot := m.Current()
	if snapshot != nil {
		view.EffectiveRevision = snapshot.Revision
	}
	m.statusMu.RLock()
	view.ApplyStatus, view.ApplyMessage = m.status, m.message
	m.statusMu.RUnlock()
	return view, nil
}

func (m *Manager) SaveDraft(ctx context.Context, actor string, request SaveDraftRequest) (*DomainView, error) {
	profile, err := normalizeAndValidateProfile(request.Values)
	if err != nil {
		return nil, err
	}
	domain, err := m.repo.GetSystemConfigDomain(ctx, AgentComponent, AgentDomain)
	if err != nil {
		return nil, err
	}
	secretVersionID := inheritedSecretVersion(domain)
	createdSecretVersionID := int64(0)
	if request.APIKey != nil {
		switch strings.ToLower(strings.TrimSpace(request.APIKey.Operation)) {
		case "replace":
			if m.envelope == nil {
				return nil, errors.New("Pole System Secret Store 根密钥未配置")
			}
			if strings.TrimSpace(request.APIKey.Value) == "" {
				return nil, errors.New("API Key 不能为空")
			}
			encrypted, encryptErr := m.envelope.encrypt(AgentComponent, AgentDomain, agentAPIKeyPurpose,
				actor, []byte(request.APIKey.Value))
			if encryptErr != nil {
				return nil, encryptErr
			}
			created, createErr := m.repo.CreateSystemSecretVersion(ctx, encrypted)
			if createErr != nil {
				return nil, createErr
			}
			secretVersionID = created.ID
			createdSecretVersionID = created.ID
		case "disable":
			secretVersionID = 0
		case "keep", "":
		default:
			return nil, errors.New("不支持的 API Key 操作")
		}
	}
	payload, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	if _, err = m.repo.SaveSystemConfigDraft(ctx, AgentComponent, AgentDomain,
		request.ExpectedDraftRevision, payload, secretVersionID, actor); err != nil {
		if createdSecretVersionID > 0 {
			_ = m.repo.DeleteSystemSecretVersion(ctx, createdSecretVersionID)
		}
		return nil, err
	}
	return m.Domain(ctx)
}

func (m *Manager) Publish(ctx context.Context, actor agentworkbench.Actor, request PublishRequest) (*DomainView, error) {
	return m.publish(ctx, actor, actor.UserID, request)
}

// SaveAndReconcile records the administrator-owned desired state and lets the
// isolated self-manager validate and promote it without a second manual step.
// The administrator credential is used only for candidate probes; revision
// history records pole-self-manager as the executor.
func (m *Manager) SaveAndReconcile(ctx context.Context, actor agentworkbench.Actor,
	request SaveDraftRequest) (*DomainView, error) {
	view, err := m.SaveDraft(ctx, actor.UserID, request)
	if err != nil {
		return nil, err
	}
	if view.Draft == nil {
		return view, errors.New("Agent desired revision was not created")
	}
	result, err := m.publish(ctx, actor, SelfManagerActorID, PublishRequest{
		DraftRevision: view.Draft.Revision,
		Description:   "automatic reconciliation after administrator update",
	})
	if err != nil {
		current, domainErr := m.Domain(ctx)
		if domainErr == nil {
			return current, err
		}
		return nil, err
	}
	return result, nil
}

func (m *Manager) publish(ctx context.Context, actor agentworkbench.Actor, executor string,
	request PublishRequest) (*DomainView, error) {
	m.applyMu.Lock()
	defer m.applyMu.Unlock()
	domain, err := m.repo.GetSystemConfigDomain(ctx, AgentComponent, AgentDomain)
	if err != nil {
		return nil, err
	}
	if domain.DraftRevision == nil || domain.DraftRevision.ID != request.DraftRevision {
		return nil, errors.New("草稿版本已变化，请刷新后重试")
	}
	profile, secret, err := m.materialize(ctx, domain.DraftRevision)
	if err != nil {
		return nil, err
	}
	agent, err := m.prepareCandidate(ctx, actor, profile, secret)
	if err != nil {
		m.setStatus("rejected", "发布前连接校验失败，继续使用上一次有效配置")
		return nil, fmt.Errorf("发布前连接校验失败: %w", err)
	}
	if strings.TrimSpace(executor) == "" {
		executor = SelfManagerActorID
	}
	revision, err := m.repo.PublishSystemConfigDraft(ctx, AgentComponent, AgentDomain,
		request.DraftRevision, executor)
	if err != nil {
		return nil, err
	}
	m.current.Store(&RuntimeSnapshot{
		Revision: strconv.FormatInt(revision.ID, 10), Profile: profile, Agent: agent, AppliedAt: time.Now().UTC(),
	})
	m.applyWorkbenchSettings(profile)
	m.setStatus("applied", "当前实例已热更新；其他实例将通过周期协调收敛")
	return m.Domain(ctx)
}

func (m *Manager) prepareCandidate(ctx context.Context, actor agentworkbench.Actor,
	profile AgentProfile, secret string) (*poleagent.Agent, error) {
	if m.candidate != nil {
		return m.candidate(ctx, actor, profile, secret)
	}
	if _, err := m.testMaterializedConnection(ctx, actor, profile, secret); err != nil {
		return nil, err
	}
	return m.buildAgent(profile, secret)
}

func (m *Manager) TestConnection(ctx context.Context, actor agentworkbench.Actor,
	request SaveDraftRequest) (*ConnectionTestResult, error) {
	profile, err := normalizeAndValidateProfile(request.Values)
	if err != nil {
		return nil, err
	}
	secret, err := m.resolveCandidateSecret(ctx, request)
	if err != nil {
		return nil, err
	}
	return m.testMaterializedConnection(ctx, actor, profile, secret)
}

func (m *Manager) testMaterializedConnection(ctx context.Context, actor agentworkbench.Actor,
	profile AgentProfile, secret string) (*ConnectionTestResult, error) {
	started := time.Now()
	model, err := poleagent.NewOpenAIModel(poleagent.OpenAIModelConfig{
		BaseURL: profile.BaseURL, APIKey: secret, Model: profile.Model,
		Timeout: parseDuration(profile.ModelTimeout, 60*time.Second),
	}, nil)
	if err != nil {
		return nil, err
	}
	modelResult, err := model.Complete(ctx, poleagent.ModelRequest{
		Model: profile.Model,
		Messages: []poleagent.Message{
			{Role: poleagent.RoleSystem, Content: "You are a connectivity probe. Reply with OK."},
			{Role: poleagent.RoleUser, Content: "OK"},
		},
	})
	if err != nil {
		return nil, err
	}
	agent, err := m.buildAgent(profile, secret)
	if err != nil {
		return nil, err
	}
	inspection := agent.Runtime(ctx, actor)
	if !inspection.Ready {
		return nil, errors.New(inspection.Reason)
	}
	return &ConnectionTestResult{
		Success: true, LatencyMS: time.Since(started).Milliseconds(), Model: modelResult.Model,
		RequestID: modelResult.RequestID, Message: "LLM Gateway 与 Pole MCP 连接正常",
	}, nil
}

func (m *Manager) Releases(ctx context.Context) ([]*RevisionView, error) {
	revisions, err := m.repo.ListSystemConfigReleases(ctx, AgentComponent, AgentDomain, 20)
	if err != nil {
		return nil, err
	}
	result := make([]*RevisionView, 0, len(revisions))
	for _, revision := range revisions {
		view, viewErr := m.revisionView(ctx, revision)
		if viewErr != nil {
			return nil, viewErr
		}
		result = append(result, view)
	}
	return result, nil
}

// EnrichSettings overlays only the currently effective Agent revision. Static
// bootstrap values remain visible until the first internal release is applied.
func (m *Manager) EnrichSettings(settings []systemconfig.EffectiveSetting) {
	snapshot := m.Current()
	if snapshot == nil || snapshot.Revision == "static" {
		return
	}
	values := map[string]any{
		"console.agent.runtime_mode":                 snapshot.Profile.RuntimeMode,
		"console.agent.definition_id":                snapshot.Profile.AgentID,
		"console.agent.prompt_builtin_version":       snapshot.Profile.PromptVersion,
		"console.agent.prompt_operator_instructions": snapshot.Profile.OperatorInstructions,
		"console.agent.model_provider":               snapshot.Profile.Provider,
		"console.agent.model_base_url":               snapshot.Profile.BaseURL,
		"console.agent.model_name":                   snapshot.Profile.Model,
		"console.agent.model_timeout":                snapshot.Profile.ModelTimeout,
		"console.agent.mcp_endpoint":                 snapshot.Profile.MCPEndpoint,
		"console.agent.mcp_tool_allowlist":           snapshot.Profile.MCPToolAllowlist,
		"console.agent.proposal_ttl":                 snapshot.Profile.ProposalTTL,
		"console.agent.upstream_timeout":             snapshot.Profile.UpstreamTimeout,
	}
	for index := range settings {
		if settings[index].Domain != AgentDomain {
			continue
		}
		if settings[index].Key == "console.agent.model_api_key" {
			settings[index].Source = systemconfig.SourceDescriptor{
				Kind: systemconfig.SourceDynamicRelease, Reference: "revision:" + snapshot.Revision,
			}
			settings[index].Configured = true
			settings[index].Redacted = true
			settings[index].Value = nil
			settings[index].DisplayValue = "••••••••"
			settings[index].ApplyMode = systemconfig.GuardedHotReload
			continue
		}
		if value, ok := values[settings[index].Key]; ok {
			settings[index].Source = systemconfig.SourceDescriptor{
				Kind: systemconfig.SourceDynamicRelease, Reference: "revision:" + snapshot.Revision,
			}
			settings[index].Value = value
			settings[index].DisplayValue = fmt.Sprint(value)
			settings[index].Configured = true
			settings[index].ApplyMode = systemconfig.GuardedHotReload
		}
	}
}

func (m *Manager) poll(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		_ = m.reload(context.Background())
	}
}

func (m *Manager) reload(ctx context.Context) error {
	m.applyMu.Lock()
	defer m.applyMu.Unlock()
	domain, err := m.repo.GetSystemConfigDomain(ctx, AgentComponent, AgentDomain)
	if err != nil || domain.ActiveRevision == nil {
		return err
	}
	current := m.Current()
	revision := strconv.FormatInt(domain.ActiveRevision.ID, 10)
	if current != nil && current.Revision == revision {
		return nil
	}
	if current != nil {
		if currentID, parseErr := strconv.ParseInt(current.Revision, 10, 64); parseErr == nil &&
			currentID > domain.ActiveRevision.ID {
			return nil
		}
	}
	profile, secret, err := m.materialize(ctx, domain.ActiveRevision)
	if err != nil {
		return err
	}
	agent, err := m.buildAgent(profile, secret)
	if err != nil {
		return err
	}
	m.current.Store(&RuntimeSnapshot{Revision: revision, Profile: profile, Agent: agent, AppliedAt: time.Now().UTC()})
	m.applyWorkbenchSettings(profile)
	m.setStatus("applied", "已从 Pole 动态配置恢复")
	return nil
}

func (m *Manager) buildAgent(profile AgentProfile, secret string) (*poleagent.Agent, error) {
	model, err := poleagent.NewOpenAIModel(poleagent.OpenAIModelConfig{
		BaseURL: profile.BaseURL, APIKey: secret, Model: profile.Model,
		Timeout: parseDuration(profile.ModelTimeout, 60*time.Second),
	}, nil)
	if err != nil {
		return nil, err
	}
	tools, err := poleagent.NewMCPToolPort(poleagent.MCPToolConfig{
		Endpoint:         profile.MCPEndpoint,
		RegistryEndpoint: poleServerRegistryEndpoint(m.config.PoleServer.Address),
		ServerNamespace:  "pole-system",
		ServerName:       "pole-control-plane",
		ToolAllowlist:    profile.MCPToolAllowlist,
		ReadTimeout:      parseDuration(profile.ModelTimeout, 60*time.Second),
	})
	if err != nil {
		return nil, err
	}
	return poleagent.New(poleagent.Options{
		ID: profile.AgentID, RuntimeMode: profile.RuntimeMode, PromptVersion: profile.PromptVersion,
		OperatorInstructions: profile.OperatorInstructions, Model: profile.Model,
	}, model, tools, m.workbench), nil
}

func (m *Manager) materialize(ctx context.Context, revision *store.SystemConfigRevision) (AgentProfile, string, error) {
	var profile AgentProfile
	if err := json.Unmarshal(revision.Payload, &profile); err != nil {
		return profile, "", errors.New("Agent 配置文档格式无效")
	}
	profile, err := normalizeAndValidateProfile(profile)
	if err != nil {
		return profile, "", err
	}
	if revision.SecretVersionID == 0 {
		return profile, "", errors.New("Agent LLM API Key 未配置")
	}
	if m.envelope == nil {
		return profile, "", errors.New("Pole System Secret Store 根密钥未配置")
	}
	version, err := m.repo.GetSystemSecretVersion(ctx, revision.SecretVersionID)
	if err != nil {
		return profile, "", err
	}
	plaintext, err := m.envelope.decrypt(version)
	return profile, string(plaintext), err
}

func (m *Manager) resolveCandidateSecret(ctx context.Context, request SaveDraftRequest) (string, error) {
	if request.APIKey != nil && strings.EqualFold(request.APIKey.Operation, "replace") {
		if strings.TrimSpace(request.APIKey.Value) == "" {
			return "", errors.New("API Key 不能为空")
		}
		return request.APIKey.Value, nil
	}
	domain, err := m.repo.GetSystemConfigDomain(ctx, AgentComponent, AgentDomain)
	if err != nil {
		return "", err
	}
	id := inheritedSecretVersion(domain)
	if id == 0 || m.envelope == nil {
		return "", errors.New("请提供 LLM Gateway API Key")
	}
	version, err := m.repo.GetSystemSecretVersion(ctx, id)
	if err != nil {
		return "", err
	}
	plaintext, err := m.envelope.decrypt(version)
	return string(plaintext), err
}

func (m *Manager) revisionView(ctx context.Context, revision *store.SystemConfigRevision) (*RevisionView, error) {
	var profile AgentProfile
	if err := json.Unmarshal(revision.Payload, &profile); err != nil {
		return nil, err
	}
	view := &RevisionView{
		Revision: revision.ID, State: revision.State, Values: profile, CreatedBy: revision.CreatedBy,
		PublishedBy: revision.PublishedBy, CreatedAt: revision.CreatedAt, PublishedAt: revision.PublishedAt,
	}
	if revision.SecretVersionID > 0 {
		version, err := m.repo.GetSystemSecretVersion(ctx, revision.SecretVersionID)
		if err != nil {
			return nil, err
		}
		view.Secret = SecretMetadata{
			Configured: true, Version: version.ID,
			Reference:   fmt.Sprintf("pole-secret://%s/%s/%s/%d", AgentComponent, AgentDomain, agentAPIKeyPurpose, version.ID),
			Fingerprint: version.Fingerprint, RotatedAt: &version.CreatedAt,
		}
	}
	return view, nil
}

func (m *Manager) setStatus(status, message string) {
	m.statusMu.Lock()
	m.status, m.message = status, message
	m.statusMu.Unlock()
}

func inheritedSecretVersion(domain *store.SystemConfigDomain) int64 {
	if domain != nil && domain.DraftRevision != nil && domain.DraftRevision.SecretVersionID > 0 {
		return domain.DraftRevision.SecretVersionID
	}
	if domain != nil && domain.ActiveRevision != nil {
		return domain.ActiveRevision.SecretVersionID
	}
	return 0
}

func profileFromBootstrap(config bootstrap.AgentConfig) AgentProfile {
	return AgentProfile{
		RuntimeMode: config.RuntimeMode, AgentID: config.Definition.ID,
		PromptVersion:        config.Definition.SystemPrompt.BuiltinVersion,
		OperatorInstructions: config.Definition.SystemPrompt.OperatorInstructions,
		Provider:             config.Model.Provider, BaseURL: config.Model.BaseURL, Model: config.Model.Model,
		ModelTimeout: config.Model.Timeout.String(), MCPEndpoint: config.MCP.Endpoint,
		MCPToolAllowlist: append([]string(nil), config.MCP.ToolAllowlist...),
		ProposalTTL:      config.ProposalTTL.String(), UpstreamTimeout: config.UpstreamTimeout.String(),
	}
}

func normalizeAndValidateProfile(profile AgentProfile) (AgentProfile, error) {
	profile.RuntimeMode = strings.TrimSpace(profile.RuntimeMode)
	if profile.RuntimeMode == "" {
		profile.RuntimeMode = "llm"
	}
	if profile.RuntimeMode != "llm" {
		return profile, errors.New("Agent 运行模式必须为 llm")
	}
	profile.Provider = strings.TrimSpace(profile.Provider)
	if profile.Provider == "" {
		profile.Provider = "openai-compatible"
	}
	if profile.Provider != "openai-compatible" {
		return profile, errors.New("当前仅支持 openai-compatible Provider")
	}
	if strings.TrimSpace(profile.ProposalTTL) == "" {
		profile.ProposalTTL = bootstrap.DefaultAgentProposalTTL.String()
	}
	if strings.TrimSpace(profile.UpstreamTimeout) == "" {
		profile.UpstreamTimeout = bootstrap.DefaultAgentUpstreamTimeout.String()
	}
	required := map[string]string{
		"Agent ID": profile.AgentID, "Prompt 版本": profile.PromptVersion, "LLM Gateway 地址": profile.BaseURL,
		"模型": profile.Model, "Pole MCP Endpoint": profile.MCPEndpoint,
	}
	for label, value := range required {
		if strings.TrimSpace(value) == "" {
			return profile, fmt.Errorf("%s不能为空", label)
		}
	}
	profile.AgentID = strings.TrimSpace(profile.AgentID)
	profile.PromptVersion = strings.TrimSpace(profile.PromptVersion)
	profile.BaseURL = strings.TrimRight(strings.TrimSpace(profile.BaseURL), "/")
	profile.Model = strings.TrimSpace(profile.Model)
	profile.MCPEndpoint = strings.TrimSpace(profile.MCPEndpoint)
	if _, err := http.NewRequest(http.MethodGet, profile.BaseURL, nil); err != nil {
		return profile, errors.New("LLM Gateway 地址无效")
	}
	if _, err := http.NewRequest(http.MethodGet, profile.MCPEndpoint, nil); err != nil {
		return profile, errors.New("Pole MCP Endpoint 无效")
	}
	if _, err := time.ParseDuration(profile.ModelTimeout); err != nil {
		return profile, errors.New("模型超时必须是 Go duration，例如 60s")
	}
	if proposalTTL, err := time.ParseDuration(profile.ProposalTTL); err != nil || proposalTTL < time.Minute || proposalTTL > 24*time.Hour {
		return profile, errors.New("提案有效期必须是 1m 到 24h 之间的 duration")
	}
	if upstreamTimeout, err := time.ParseDuration(profile.UpstreamTimeout); err != nil || upstreamTimeout < time.Second || upstreamTimeout > 5*time.Minute {
		return profile, errors.New("资源工具上游超时必须是 1s 到 5m 之间的 duration")
	}
	allowlist := make([]string, 0, len(profile.MCPToolAllowlist))
	seen := map[string]struct{}{}
	for _, tool := range profile.MCPToolAllowlist {
		tool = strings.TrimSpace(tool)
		if tool == "" {
			continue
		}
		if _, ok := seen[tool]; ok {
			continue
		}
		seen[tool] = struct{}{}
		allowlist = append(allowlist, tool)
	}
	profile.MCPToolAllowlist = allowlist
	return profile, nil
}

func (m *Manager) applyWorkbenchSettings(profile AgentProfile) {
	m.workbench.SetTTL(parseDuration(profile.ProposalTTL, bootstrap.DefaultAgentProposalTTL))
	m.workbench.SetUpstreamTimeout(parseDuration(profile.UpstreamTimeout, bootstrap.DefaultAgentUpstreamTimeout))
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fallback
	}
	return duration
}

func poleServerRegistryEndpoint(address string) string {
	address = strings.TrimRight(strings.TrimSpace(address), "/")
	if address == "" {
		return ""
	}
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}
	return address + "/ai/mcp/v1/servers"
}
