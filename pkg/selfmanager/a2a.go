package selfmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

func (m *Manager) reconcileA2A(ctx context.Context, state DesiredState, result *Result) error {
	agent, err := normalizeA2AAgent(state.A2A)
	if err != nil {
		return err
	}
	existing, err := m.store.GetA2AAgentByName(agent.Name, agent.Namespace)
	if err != nil {
		return fmt.Errorf("get A2A agent %s/%s: %w", agent.Namespace, agent.Name, err)
	}
	if existing == nil {
		existing, err = m.store.GetA2AAgent(agent.Id)
		if err != nil {
			return fmt.Errorf("get A2A agent %s by deterministic ID: %w", agent.Id, err)
		}
	}
	if existing == nil {
		if err := m.store.CreateA2AAgent(agent); err != nil {
			return fmt.Errorf("create A2A agent %s/%s: %w", agent.Namespace, agent.Name, err)
		}
		result.Created++
		revision, err := a2aRevision(agent)
		if err != nil {
			return err
		}
		return m.recordAudit(ctx, AuditEvent{
			Actor:         SystemActor,
			Trigger:       strings.TrimSpace(state.Trigger),
			ConfiguredBy:  strings.TrimSpace(state.ConfiguredBy),
			ResourceKind:  ResourceA2AAgent,
			Namespace:     agent.Namespace,
			Name:          agent.Name,
			ResourceID:    agent.Id,
			Action:        ActionCreate,
			AfterRevision: revision,
		})
	}

	current, err := normalizeA2AAgent(existing)
	if err != nil {
		return fmt.Errorf("normalize existing A2A agent %s/%s: %w", agent.Namespace, agent.Name, err)
	}
	beforeRevision, err := a2aRevision(current)
	if err != nil {
		return err
	}
	agent.Id = existing.Id
	deleted := preserveA2AChildIDs(agent, current)
	if existing.Flag != 1 && sameA2AAgent(current, agent) {
		result.Unchanged++
		return nil
	}
	if err := m.store.UpdateA2AAgent(agent); err != nil {
		return fmt.Errorf("update A2A agent %s/%s: %w", agent.Namespace, agent.Name, err)
	}
	afterRevision, err := a2aRevision(agent)
	if err != nil {
		return err
	}
	result.Updated++
	result.Deleted += deleted
	return m.recordAudit(ctx, AuditEvent{
		Actor:          SystemActor,
		Trigger:        strings.TrimSpace(state.Trigger),
		ConfiguredBy:   strings.TrimSpace(state.ConfiguredBy),
		ResourceKind:   ResourceA2AAgent,
		Namespace:      agent.Namespace,
		Name:           agent.Name,
		ResourceID:     agent.Id,
		Action:         ActionUpdate,
		BeforeRevision: beforeRevision,
		AfterRevision:  afterRevision,
	})
}

func normalizeA2AAgent(source *aitypes.A2AAgent) (*aitypes.A2AAgent, error) {
	if source == nil {
		return nil, errors.New("desired A2A agent is required")
	}
	agent := cloneA2AAgent(source)
	trimA2AAgent(agent)
	if agent.Name == "" || agent.Namespace == "" {
		return nil, errors.New("desired A2A agent name and namespace are required")
	}
	if agent.Id == "" {
		agent.Id = stableID("a2a-agent", agent.Namespace, agent.Name)
	}
	var err error
	if agent.RawCardJson, err = canonicalJSON(agent.RawCardJson); err != nil {
		return nil, fmt.Errorf("normalize A2A agent card: %w", err)
	}
	agent.Metadata = normalizeStringMap(agent.Metadata)

	interfaceKeys := make(map[string]struct{}, len(agent.Interfaces))
	interfaces := make([]*aitypes.A2AAgentInterface, 0, len(agent.Interfaces))
	for _, sourceInterface := range agent.Interfaces {
		if sourceInterface == nil {
			continue
		}
		item := *sourceInterface
		item.Url = strings.TrimSpace(item.Url)
		item.ProtocolBinding = strings.TrimSpace(item.ProtocolBinding)
		item.ProtocolVersion = strings.TrimSpace(item.ProtocolVersion)
		item.Tenant = strings.TrimSpace(item.Tenant)
		if item.Url == "" || item.ProtocolBinding == "" {
			return nil, errors.New("desired A2A interface URL and protocol binding are required")
		}
		key := a2aInterfaceKey(&item)
		if _, duplicate := interfaceKeys[key]; duplicate {
			return nil, fmt.Errorf("duplicate desired A2A interface %s", key)
		}
		interfaceKeys[key] = struct{}{}
		item.AgentId = agent.Id
		if item.Id == "" {
			item.Id = stableID("a2a-interface", agent.Id, key)
		}
		item.Flag, item.Ctime, item.Mtime = 0, "", ""
		interfaces = append(interfaces, &item)
	}
	sort.Slice(interfaces, func(i, j int) bool {
		return a2aInterfaceKey(interfaces[i]) < a2aInterfaceKey(interfaces[j])
	})
	agent.Interfaces = interfaces

	skillKeys := make(map[string]struct{}, len(agent.Skills))
	skills := make([]*aitypes.A2AAgentSkill, 0, len(agent.Skills))
	for _, sourceSkill := range agent.Skills {
		if sourceSkill == nil {
			continue
		}
		item := *sourceSkill
		item.SkillId = strings.TrimSpace(item.SkillId)
		item.Name = strings.TrimSpace(item.Name)
		item.Description = strings.TrimSpace(item.Description)
		if item.SkillId == "" {
			return nil, errors.New("desired A2A skill ID is required")
		}
		if _, duplicate := skillKeys[item.SkillId]; duplicate {
			return nil, fmt.Errorf("duplicate desired A2A skill %s", item.SkillId)
		}
		skillKeys[item.SkillId] = struct{}{}
		item.AgentId = agent.Id
		if item.Id == "" {
			item.Id = stableID("a2a-skill", agent.Id, item.SkillId)
		}
		item.Tags = normalizeStringSet(item.Tags)
		item.Examples = normalizeStrings(item.Examples)
		item.InputModes = normalizeStringSet(item.InputModes)
		item.OutputModes = normalizeStringSet(item.OutputModes)
		if item.SecurityRequirementsJson, err = canonicalJSON(item.SecurityRequirementsJson); err != nil {
			return nil, fmt.Errorf("normalize A2A skill %s security requirements: %w", item.SkillId, err)
		}
		item.Flag, item.Ctime, item.Mtime = 0, "", ""
		skills = append(skills, &item)
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].SkillId < skills[j].SkillId
	})
	agent.Skills = skills
	agent.Flag, agent.Ctime, agent.Mtime = 0, "", ""
	return agent, nil
}

func trimA2AAgent(agent *aitypes.A2AAgent) {
	agent.Name = strings.TrimSpace(agent.Name)
	agent.Namespace = strings.TrimSpace(agent.Namespace)
	agent.Visibility = strings.TrimSpace(agent.Visibility)
	agent.Description = strings.TrimSpace(agent.Description)
	agent.Version = strings.TrimSpace(agent.Version)
	agent.ProtocolVersion = strings.TrimSpace(agent.ProtocolVersion)
	agent.ProviderOrganization = strings.TrimSpace(agent.ProviderOrganization)
	agent.ProviderUrl = strings.TrimSpace(agent.ProviderUrl)
	agent.DocumentationUrl = strings.TrimSpace(agent.DocumentationUrl)
	agent.IconUrl = strings.TrimSpace(agent.IconUrl)
	agent.Business = strings.TrimSpace(agent.Business)
	agent.Department = strings.TrimSpace(agent.Department)
	agent.BackendType = strings.TrimSpace(agent.BackendType)
	agent.BackendServiceNamespace = strings.TrimSpace(agent.BackendServiceNamespace)
	agent.BackendServiceName = strings.TrimSpace(agent.BackendServiceName)
	agent.BackendAddress = strings.TrimSpace(agent.BackendAddress)
	agent.PreferredInterfaceUrl = strings.TrimSpace(agent.PreferredInterfaceUrl)
	agent.PreferredProtocolBinding = strings.TrimSpace(agent.PreferredProtocolBinding)
	agent.PreferredProtocolVersion = strings.TrimSpace(agent.PreferredProtocolVersion)
	agent.SourceType = strings.TrimSpace(agent.SourceType)
	agent.SourceUrl = strings.TrimSpace(agent.SourceUrl)
	agent.LastFetchStatus = strings.TrimSpace(agent.LastFetchStatus)
	agent.LastFetchTime = strings.TrimSpace(agent.LastFetchTime)
}

func preserveA2AChildIDs(desired, current *aitypes.A2AAgent) int {
	currentInterfaces := make(map[string]*aitypes.A2AAgentInterface, len(current.Interfaces))
	for _, item := range current.Interfaces {
		currentInterfaces[a2aInterfaceKey(item)] = item
	}
	for _, item := range desired.Interfaces {
		key := a2aInterfaceKey(item)
		if existing := currentInterfaces[key]; existing != nil {
			item.Id = existing.Id
		}
		item.AgentId = desired.Id
		delete(currentInterfaces, key)
	}

	currentSkills := make(map[string]*aitypes.A2AAgentSkill, len(current.Skills))
	for _, item := range current.Skills {
		currentSkills[item.SkillId] = item
	}
	for _, item := range desired.Skills {
		if existing := currentSkills[item.SkillId]; existing != nil {
			item.Id = existing.Id
		}
		item.AgentId = desired.Id
		delete(currentSkills, item.SkillId)
	}
	return len(currentInterfaces) + len(currentSkills)
}

func a2aInterfaceKey(item *aitypes.A2AAgentInterface) string {
	return strings.Join([]string{item.Url, item.ProtocolBinding, item.Tenant}, "\x00")
}

func sameA2AAgent(left, right *aitypes.A2AAgent) bool {
	return reflect.DeepEqual(left, right)
}

func a2aRevision(agent *aitypes.A2AAgent) (string, error) {
	encoded, err := json.Marshal(agent)
	if err != nil {
		return "", fmt.Errorf("build A2A revision: %w", err)
	}
	return digest32(encoded), nil
}

func normalizeStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key != "" {
			result[key] = strings.TrimSpace(value)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func normalizeStringSet(values []string) []string {
	result := normalizeStrings(values)
	sort.Strings(result)
	return result
}

func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func cloneA2AAgent(value *aitypes.A2AAgent) *aitypes.A2AAgent {
	if value == nil {
		return nil
	}
	copyValue := *value
	copyValue.Metadata = make(map[string]string, len(value.Metadata))
	for key, item := range value.Metadata {
		copyValue.Metadata[key] = item
	}
	copyValue.Interfaces = make([]*aitypes.A2AAgentInterface, 0, len(value.Interfaces))
	for _, item := range value.Interfaces {
		if item == nil {
			continue
		}
		itemCopy := *item
		copyValue.Interfaces = append(copyValue.Interfaces, &itemCopy)
	}
	copyValue.Skills = make([]*aitypes.A2AAgentSkill, 0, len(value.Skills))
	for _, item := range value.Skills {
		if item == nil {
			continue
		}
		itemCopy := *item
		itemCopy.Tags = append([]string(nil), item.Tags...)
		itemCopy.Examples = append([]string(nil), item.Examples...)
		itemCopy.InputModes = append([]string(nil), item.InputModes...)
		itemCopy.OutputModes = append([]string(nil), item.OutputModes...)
		copyValue.Skills = append(copyValue.Skills, &itemCopy)
	}
	return &copyValue
}
