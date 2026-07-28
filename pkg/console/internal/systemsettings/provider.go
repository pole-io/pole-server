package systemsettings

import (
	"strings"

	consoleconfig "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/pole-io/pole-server/pkg/systemconfig"
)

const (
	DefaultAgentProposalTTL     = consoleconfig.DefaultAgentProposalTTL
	DefaultAgentUpstreamTimeout = consoleconfig.DefaultAgentUpstreamTimeout
)

type Config = consoleconfig.Config
type WebServer = consoleconfig.WebServer
type JWT = consoleconfig.JWT
type AgentConfig = consoleconfig.AgentConfig
type AgentModelConfig = consoleconfig.AgentModelConfig

func NewSystemSettingsProvider(cfg *consoleconfig.Config) (systemconfig.EffectiveProvider, error) {
	sources := cfg.SystemConfigSources.Clone()
	if cfg.ObservabilityQuery.Provider == "" {
		sources.MarkCompiledDefault("observabilityQuery.provider")
	}
	if cfg.ObservabilityQuery.Database == "" {
		sources.MarkCompiledDefault("observabilityQuery.database")
	}
	if cfg.ObservabilityQuery.Timeout == "" {
		sources.MarkCompiledDefault("observabilityQuery.timeout")
	}
	if cfg.Agent.RuntimeMode == "" {
		sources.MarkCompiledDefault("agent.runtimeMode")
	}
	if cfg.Agent.Definition.ID == "" {
		sources.MarkCompiledDefault("agent.definition.id")
	}
	if cfg.Agent.Definition.SystemPrompt.BuiltinVersion == "" {
		sources.MarkCompiledDefault("agent.definition.systemPrompt.builtinVersion")
	}
	if cfg.Agent.Model.Provider == "" {
		sources.MarkCompiledDefault("agent.model.provider")
	}
	if cfg.Agent.Model.Model == "" {
		sources.MarkCompiledDefault("agent.model.model")
	}
	if cfg.Agent.Model.Timeout <= 0 {
		sources.MarkCompiledDefault("agent.model.timeout")
	}
	if cfg.Agent.MCP.Endpoint == "" {
		sources.MarkCompiledDefault("agent.mcp.endpoint")
	}
	if cfg.Agent.ProposalTTL <= 0 {
		sources.MarkCompiledDefault("agent.proposalTTL")
	}
	if cfg.Agent.UpstreamTimeout <= 0 {
		sources.MarkCompiledDefault("agent.upstreamTimeout")
	}
	observability := func() observabilityValues {
		value := cfg.ObservabilityQuery.Normalize()
		return observabilityValues{
			provider: value.Provider, endpoint: value.Endpoint,
			database: value.Database, timeout: value.Timeout,
		}
	}
	agent := cfg.Agent.Normalize()
	registrations := []systemconfig.Registration{
		consoleSetting("console.logger.level", "logging", "日志级别", "logger.level", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return cfg.Logger.Level }),
		consoleSetting("console.logger.rotation_max_size", "logging", "单文件大小上限", "logger.RotationMaxSize", systemconfig.ValueTypeInteger, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return cfg.Logger.RotationMaxSize }),
		consoleSetting("console.logger.rotation_max_age", "logging", "日志保留天数", "logger.RotationMaxAge", systemconfig.ValueTypeInteger, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return cfg.Logger.RotationMaxAge }),
		consoleSetting("console.web.mode", "runtime", "Web 运行模式", "webServer.mode", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return cfg.WebServer.Mode }),
		consoleSetting("console.web.listen_ip", "runtime", "监听地址", "webServer.listenIP", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return cfg.WebServer.ListenIP }),
		consoleSetting("console.web.listen_port", "runtime", "监听端口", "webServer.listenPort", systemconfig.ValueTypeInteger, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return cfg.WebServer.ListenPort }),
		consoleSetting("console.web.jwt.expired", "auth", "登录会话有效期", "webServer.jwt.expired", systemconfig.ValueTypeInteger, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return cfg.WebServer.JWT.Expired }),
		consoleSetting("console.web.jwt.secret_key", "auth", "登录会话签名密钥", "webServer.jwt.secretKey", systemconfig.ValueTypeSecret, systemconfig.BootstrapOnly, systemconfig.SensitivitySecret, func() any { return cfg.WebServer.JWT.SecretKey }),
		consoleSetting("console.web.main_user", "auth", "主账号", "webServer.mainUser", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return cfg.WebServer.MainUser }),
		consoleSetting("console.pole_server.address", "upstream", "Pole Server 地址", "poleServer.address", systemconfig.ValueTypeString, systemconfig.BootstrapOnly, systemconfig.SensitivityInternal, func() any { return cfg.PoleServer.Address }),
		consoleSetting("console.monitor_server.address", "upstream", "监控服务地址", "monitorServer.address", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return cfg.MonitorServer.Address }),
		consoleSetting("console.features", "runtime", "功能开关", "futures", systemconfig.ValueTypeList, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return splitFeatures(cfg.Futures) }),
		consoleSetting("console.store.name", "storage", "观测存储插件", "store.name", systemconfig.ValueTypeString, systemconfig.BootstrapOnly, systemconfig.SensitivityInternal, func() any { return cfg.Store.Name }),
		consoleSetting("console.store.db_user", "storage", "观测数据库账号", "store.option.master.dbUser", systemconfig.ValueTypeString, systemconfig.BootstrapOnly, systemconfig.SensitivityInternal, func() any { return systemconfig.LookupValue(cfg.Store.Option, "master", "dbUser") }),
		consoleSetting("console.store.db_password", "storage", "观测数据库密码", "store.option.master.dbPwd", systemconfig.ValueTypeSecret, systemconfig.BootstrapOnly, systemconfig.SensitivitySecret, func() any { return systemconfig.LookupValue(cfg.Store.Option, "master", "dbPwd") }),
		consoleSetting("console.store.db_address", "storage", "观测数据库地址", "store.option.master.dbAddr", systemconfig.ValueTypeString, systemconfig.BootstrapOnly, systemconfig.SensitivityInternal, func() any { return systemconfig.LookupValue(cfg.Store.Option, "master", "dbAddr") }),
		consoleSetting("console.observability.provider", "observability", "观测查询提供方", "observabilityQuery.provider", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return observability().provider }),
		consoleSetting("console.observability.endpoint", "observability", "观测查询地址", "observabilityQuery.endpoint", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return observability().endpoint }),
		consoleSetting("console.observability.database", "observability", "观测数据库", "observabilityQuery.database", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return observability().database }),
		consoleSetting("console.observability.timeout", "observability", "观测查询超时", "observabilityQuery.timeout", systemconfig.ValueTypeDuration, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return observability().timeout }),
		consoleSetting("console.agent.runtime_mode", "agent", "当前运行模式", "agent.runtimeMode", systemconfig.ValueTypeString, systemconfig.BootstrapOnly, systemconfig.SensitivityPublic, func() any { return agent.RuntimeMode }),
		consoleSetting("console.agent.definition_id", "agent", "Agent ID", "agent.definition.id", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return agent.Definition.ID }),
		consoleSetting("console.agent.prompt_builtin_version", "agent", "内置 Prompt 版本", "agent.definition.systemPrompt.builtinVersion", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return agent.Definition.SystemPrompt.BuiltinVersion }),
		consoleSetting("console.agent.prompt_operator_instructions", "agent", "Operator Instructions", "agent.definition.systemPrompt.operatorInstructions", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return agent.Definition.SystemPrompt.OperatorInstructions }),
		consoleSetting("console.agent.model_provider", "agent", "LLM Provider", "agent.model.provider", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return agent.Model.Provider }),
		consoleSetting("console.agent.model_base_url", "agent", "LLM Gateway 地址", "agent.model.baseURL", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return agent.Model.BaseURL }),
		consoleSetting("console.agent.model_name", "agent", "模型", "agent.model.model", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return agent.Model.Model }),
		consoleSetting("console.agent.model_api_key", "agent", "LLM API Key", "agent.model.apiKey", systemconfig.ValueTypeSecret, systemconfig.RestartRequired, systemconfig.SensitivitySecret, func() any { return agent.Model.APIKey }),
		consoleSetting("console.agent.model_timeout", "agent", "模型请求超时", "agent.model.timeout", systemconfig.ValueTypeDuration, systemconfig.RestartRequired, systemconfig.SensitivityPublic, func() any { return agent.Model.Timeout }),
		consoleSetting("console.agent.mcp_endpoint", "agent", "Pole MCP Endpoint", "agent.mcp.endpoint", systemconfig.ValueTypeString, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return agent.MCP.Endpoint }),
		consoleSetting("console.agent.mcp_tool_allowlist", "agent", "MCP 工具白名单", "agent.mcp.toolAllowlist", systemconfig.ValueTypeList, systemconfig.RestartRequired, systemconfig.SensitivityInternal, func() any { return agent.MCP.ToolAllowlist }),
		consoleSetting("console.agent.proposal_ttl", "agent", "提案有效期", "agent.proposalTTL", systemconfig.ValueTypeDuration, systemconfig.GuardedHotReload, systemconfig.SensitivityPublic, func() any { return agent.ProposalTTL }),
		consoleSetting("console.agent.upstream_timeout", "agent", "资源工具上游超时", "agent.upstreamTimeout", systemconfig.ValueTypeDuration, systemconfig.GuardedHotReload, systemconfig.SensitivityPublic, func() any { return agent.UpstreamTimeout }),
	}
	systemconfig.ApplyEditPolicies(registrations, consoleSystemSettingPolicies())
	return systemconfig.NewRegistry(registrations, sources)
}

func consoleSystemSettingPolicies() map[string]systemconfig.EditPolicy {
	required := systemconfig.Validation{Required: true}
	positive := func(min, max int64) systemconfig.Validation {
		return systemconfig.Validation{Required: true, MinInteger: &min, MaxInteger: &max}
	}
	duration := func(min, max string) systemconfig.Validation {
		return systemconfig.Validation{Required: true, MinDuration: min, MaxDuration: max}
	}
	locked := func(reason, description string) systemconfig.EditPolicy {
		return systemconfig.EditPolicy{Reason: reason, Description: description}
	}
	restart := func(description string, validation systemconfig.Validation) systemconfig.EditPolicy {
		return systemconfig.EditPolicy{
			Editable: true, Reason: "发布后进入待重启状态", Description: description, Validation: validation,
		}
	}
	guarded := func(description string, validation systemconfig.Validation) systemconfig.EditPolicy {
		return systemconfig.EditPolicy{
			Editable: true, Reason: "发布前校验并受控热更新", Description: description, Validation: validation,
		}
	}
	return map[string]systemconfig.EditPolicy{
		"console.logger.level": restart("Console 日志级别。", systemconfig.Validation{
			Required: true, Options: []string{"debug", "info", "warn", "error"},
		}),
		"console.logger.rotation_max_size": restart("单个日志文件的最大大小，单位 MB。", positive(1, 10240)),
		"console.logger.rotation_max_age":  restart("日志文件保留天数。", positive(1, 3650)),

		"console.web.mode":        locked("Web 运行模式属于进程装配边界", "决定 Console HTTP 服务如何启动。"),
		"console.web.listen_ip":   locked("监听地址由 Kubernetes 网络与探针共同约束", "修改后需要同步 Deployment、Service 与探针。"),
		"console.web.listen_port": locked("监听端口由 Kubernetes Service 与探针共同约束", "修改后需要同步 Deployment、Service 与 Gateway。"),
		"console.features":        restart("Console 功能开关列表。", systemconfig.Validation{}),

		"console.web.jwt.expired":    restart("登录会话有效期，单位秒。", positive(60, 2592000)),
		"console.web.jwt.secret_key": locked("会话签名密钥需要双密钥轮换", "建立 old/new key 验证窗口前只能通过部署 Secret 修改。"),
		"console.web.main_user":      locked("主账号属于身份自举边界", "主账号通过认证管理维护，不能在系统配置中重命名。"),

		"console.pole_server.address":    locked("Pole Server 地址属于 Console 自举连接", "Console 必须先连接 Pole Server 才能提供系统配置页面。"),
		"console.monitor_server.address": restart("Console 访问兼容监控服务的地址。", systemconfig.Validation{Format: "host-port"}),

		"console.store.name":        locked("观测存储插件属于自举岛", "存储插件在系统配置仓库之前初始化。"),
		"console.store.db_user":     locked("观测数据库账号属于自举连接", "数据库凭证由部署配置提供。"),
		"console.store.db_password": locked("观测数据库密码属于自举 Secret", "数据库凭证由部署 Secret 提供。"),
		"console.store.db_address":  locked("观测数据库地址属于自举连接", "数据库连接在系统配置仓库之前初始化。"),

		"console.observability.provider": restart("可观测查询 Provider。", systemconfig.Validation{
			Required: true, Options: []string{"greptimedb", "prometheus"},
		}),
		"console.observability.endpoint": restart("可观测查询服务的 HTTP 地址。", systemconfig.Validation{Required: true, Format: "url"}),
		"console.observability.database": restart("可观测查询使用的数据库或 catalog。", required),
		"console.observability.timeout":  restart("单次可观测查询超时。", duration("1s", "5m")),

		"console.agent.runtime_mode":                 locked("Agent 运行模式属于能力装配边界", "当前仅支持 llm 运行模式。"),
		"console.agent.definition_id":                guarded("Agent 的稳定标识。", required),
		"console.agent.prompt_builtin_version":       guarded("内置系统 Prompt 版本。", required),
		"console.agent.prompt_operator_instructions": guarded("叠加到内置系统 Prompt 的管理员指令。", systemconfig.Validation{}),
		"console.agent.model_provider": guarded("LLM Gateway 的协议提供方。", systemconfig.Validation{
			Required: true, Options: []string{"openai-compatible"},
		}),
		"console.agent.model_base_url":     guarded("LLM Gateway 的 OpenAI 兼容地址。", systemconfig.Validation{Required: true, Format: "url"}),
		"console.agent.model_name":         guarded("Agent 调用的模型名。", required),
		"console.agent.model_api_key":      guarded("由 Pole Secret Store 托管的 LLM Gateway 凭证。", systemconfig.Validation{Required: true}),
		"console.agent.model_timeout":      guarded("单次模型请求超时。", duration("1s", "5m")),
		"console.agent.mcp_endpoint":       guarded("Pole MCP SSE Endpoint。", systemconfig.Validation{Required: true, Format: "url"}),
		"console.agent.mcp_tool_allowlist": guarded("Agent 可以调用的 Pole MCP 工具集合。", systemconfig.Validation{}),
		"console.agent.proposal_ttl":       guarded("Agent 变更提案的有效期。", duration("1m", "24h")),
		"console.agent.upstream_timeout":   guarded("Agent 资源工具访问 Pole Server 的超时。", duration("1s", "5m")),
	}
}

type observabilityValues struct {
	provider string
	endpoint string
	database string
	timeout  string
}

func consoleSetting(key, domain, label, sourcePath string, valueType systemconfig.ValueType,
	applyMode systemconfig.ApplyMode, sensitivity systemconfig.Sensitivity, value func() any) systemconfig.Registration {
	return systemconfig.Registration{
		Definition: systemconfig.SettingDefinition{
			Key: key, Component: systemconfig.ComponentConsole, Domain: domain, Label: label,
			ValueType: valueType, ApplyMode: applyMode, Sensitivity: sensitivity, Owner: "pole-console",
		},
		SourcePath: sourcePath,
		Value:      value,
	}
}

func splitFeatures(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if feature := strings.TrimSpace(part); feature != "" {
			result = append(result, feature)
		}
	}
	return result
}
