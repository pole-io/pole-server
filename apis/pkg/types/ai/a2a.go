/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 */

package ai

type A2AAgent struct {
	Id                       string               `json:"id"`
	Name                     string               `json:"name"`
	Namespace                string               `json:"namespace"`
	Visibility               string               `json:"visibility"`
	Description              string               `json:"description"`
	Version                  string               `json:"version"`
	ProtocolVersion          string               `json:"protocol_version"`
	ProviderOrganization     string               `json:"provider_organization"`
	ProviderUrl              string               `json:"provider_url"`
	DocumentationUrl         string               `json:"documentation_url"`
	IconUrl                  string               `json:"icon_url"`
	Business                 string               `json:"business"`
	Department               string               `json:"department"`
	BackendType              string               `json:"backend_type"`
	BackendServiceNamespace  string               `json:"backend_service_namespace"`
	BackendServiceName       string               `json:"backend_service_name"`
	BackendAddress           string               `json:"backend_address"`
	PreferredInterfaceUrl    string               `json:"preferred_interface_url"`
	PreferredProtocolBinding string               `json:"preferred_protocol_binding"`
	PreferredProtocolVersion string               `json:"preferred_protocol_version"`
	Streaming                bool                 `json:"streaming"`
	PushNotifications        bool                 `json:"push_notifications"`
	ExtendedAgentCard        bool                 `json:"extended_agent_card"`
	RawCardJson              string               `json:"raw_card_json"`
	SourceType               string               `json:"source_type"`
	SourceUrl                string               `json:"source_url"`
	LastFetchStatus          string               `json:"last_fetch_status"`
	LastFetchTime            string               `json:"last_fetch_time"`
	Metadata                 map[string]string    `json:"metadata"`
	Interfaces               []*A2AAgentInterface `json:"interfaces"`
	Skills                   []*A2AAgentSkill     `json:"skills"`
	Flag                     uint32               `json:"flag"`
	Ctime                    string               `json:"ctime"`
	Mtime                    string               `json:"mtime"`
}

type A2AAgentInterface struct {
	Id              string `json:"id"`
	AgentId         string `json:"agent_id"`
	Url             string `json:"url"`
	ProtocolBinding string `json:"protocol_binding"`
	ProtocolVersion string `json:"protocol_version"`
	Tenant          string `json:"tenant"`
	Flag            uint32 `json:"flag"`
	Ctime           string `json:"ctime"`
	Mtime           string `json:"mtime"`
}

type A2AAgentSkill struct {
	Id                       string   `json:"id"`
	AgentId                  string   `json:"agent_id"`
	SkillId                  string   `json:"skill_id"`
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	Tags                     []string `json:"tags"`
	Examples                 []string `json:"examples"`
	InputModes               []string `json:"input_modes"`
	OutputModes              []string `json:"output_modes"`
	SecurityRequirementsJson string   `json:"security_requirements_json"`
	Flag                     uint32   `json:"flag"`
	Ctime                    string   `json:"ctime"`
	Mtime                    string   `json:"mtime"`
}

type A2AAgentQuery struct {
	Name                    string `json:"name"`
	Namespace               string `json:"namespace"`
	Business                string `json:"business"`
	Department              string `json:"department"`
	ProtocolBinding         string `json:"protocol_binding"`
	SkillTag                string `json:"skill_tag"`
	BackendType             string `json:"backend_type"`
	BackendServiceNamespace string `json:"backend_service_namespace"`
	BackendServiceName      string `json:"backend_service_name"`
	Streaming               *bool  `json:"streaming"`
	PushNotifications       *bool  `json:"push_notifications"`
	Offset                  uint32 `json:"offset"`
	Limit                   uint32 `json:"limit"`
}

type A2AAgents struct {
	Agents []*A2AAgent `json:"agents"`
}

type A2AAgentDeleteRequest struct {
	AgentIds []string `json:"agent_ids"`
}

type A2AAgentListResponse struct {
	Code   uint32      `json:"code"`
	Info   string      `json:"info"`
	Amount uint32      `json:"amount"`
	Size   uint32      `json:"size"`
	Data   []*A2AAgent `json:"data"`
}
