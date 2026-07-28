package poleagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const namespaceDirectoryTool = "list_namespaces"

func validateNamespaceScopeAccess(
	ctx context.Context,
	session ToolSession,
	availableTools []ToolDefinition,
	scope NamespaceScope,
) error {
	if _, found := findTool(availableTools, namespaceDirectoryTool); !found {
		return runtimeError(CategoryToolUnavailable, 503131,
			"Pole MCP namespace directory tool is required to validate Agent scope", false, nil)
	}
	result, err := session.CallTool(ctx, namespaceDirectoryTool, map[string]any{"all": true})
	if err != nil {
		return err
	}
	if result.IsError {
		return runtimeError(CategoryToolRejected, 403009,
			"namespace scope could not be authorized", false, nil)
	}

	var response struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(result.Content), &response); err != nil {
		return runtimeError(CategoryDownstreamFailed, 502131,
			"decode authorized namespace directory", false, err)
	}
	allowed := make(map[string]struct{}, len(response.Data))
	for _, namespace := range response.Data {
		name := strings.TrimSpace(fmt.Sprint(namespace["name"]))
		if name != "" && isBusinessNamespaceRecord(name, namespace["kind"]) {
			allowed[name] = struct{}{}
		}
	}
	for _, namespace := range scope.Namespaces {
		if _, ok := allowed[strings.TrimSpace(namespace)]; !ok {
			return runtimeError(CategoryToolRejected, 403009,
				"namespace scope contains an inaccessible or non-business namespace", false, nil)
		}
	}
	return nil
}

func isBusinessNamespaceRecord(name string, rawKind any) bool {
	if name == "pole-system" {
		return false
	}
	kind := strings.ToUpper(strings.TrimSpace(fmt.Sprint(rawKind)))
	switch kind {
	case "", "0", "BUSINESS", "NAMESPACE_KIND_BUSINESS":
		return true
	default:
		return false
	}
}
