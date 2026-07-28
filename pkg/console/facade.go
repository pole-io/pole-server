package console

import (
	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	consoleconfig "github.com/pole-io/pole-server/pkg/console/config"
)

// Config 是 Console 模块接受的完整配置。
type Config = consoleconfig.Config
type LoggerOptions = consoleconfig.LoggerOptions

type PoleServer = consoleconfig.PoleServer
type MonitorServer = consoleconfig.MonitorServer
type WebServer = consoleconfig.WebServer
type JWT = consoleconfig.JWT
type AgentConfig = consoleconfig.AgentConfig
type AgentDefinitionConfig = consoleconfig.AgentDefinitionConfig
type AgentSystemPromptConfig = consoleconfig.AgentSystemPromptConfig
type AgentModelConfig = consoleconfig.AgentModelConfig
type AgentMCPConfig = consoleconfig.AgentMCPConfig
type AgentA2AConfig = consoleconfig.AgentA2AConfig
type SystemSecretsConfig = consoleconfig.SystemSecretsConfig

// A2A 契约由共享 API 类型层定义，Console 根包只暴露模块启动所需别名。
type A2AAgentCard = aitypes.A2AAgentCard
type A2ASkill = aitypes.A2ASkill

func DefaultLoggerOptions() LoggerOptions {
	return consoleconfig.DefaultLoggerOptions()
}
