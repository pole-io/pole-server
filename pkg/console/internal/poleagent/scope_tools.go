package poleagent

import (
	"fmt"
	"strings"
)

func namespaceScopedRemoteTools(tools []ToolDefinition) []ToolDefinition {
	scoped := make([]ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		if toolHasNamespaceArgument(tool) {
			scoped = append(scoped, tool)
		}
	}
	return scoped
}

func enforceToolNamespace(scope NamespaceScope, arguments map[string]any) error {
	raw, exists := arguments["namespace"]
	if !exists || strings.TrimSpace(fmt.Sprint(raw)) == "" {
		if !scope.IsSingle() {
			return fmt.Errorf("cross-environment tool calls must select one namespace")
		}
		arguments["namespace"] = scope.SingleNamespace()
		return nil
	}
	namespace, ok := raw.(string)
	if !ok || !scope.Contains(namespace) {
		return fmt.Errorf("tool namespace is outside the turn scope")
	}
	arguments["namespace"] = strings.TrimSpace(namespace)
	return nil
}

func toolHasNamespaceArgument(tool ToolDefinition) bool {
	properties, ok := tool.InputSchema["properties"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = properties["namespace"]
	return ok
}
